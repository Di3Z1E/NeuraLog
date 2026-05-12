package collector

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/Di3Z1E/neuralog/internal/config"
	"github.com/Di3Z1E/neuralog/internal/hub"
	"github.com/Di3Z1E/neuralog/internal/redactor"
	"github.com/Di3Z1E/neuralog/internal/store"
)

// PodInfo is the pod state exposed over the API.
type PodInfo struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Node      string `json:"node"`
	HasLogs   bool   `json:"hasLogs"`
}

type Collector struct {
	client   kubernetes.Interface
	store    *store.Store
	hub      *hub.Hub
	redactor *redactor.Redactor
	cfgMgr   *config.Manager

	mu     sync.RWMutex
	pods   map[string]*PodInfo
	active map[string]context.CancelFunc
}

func New(
	client kubernetes.Interface,
	st *store.Store,
	h *hub.Hub,
	r *redactor.Redactor,
	cfgMgr *config.Manager,
) *Collector {
	return &Collector{
		client:   client,
		store:    st,
		hub:      h,
		redactor: r,
		cfgMgr:   cfgMgr,
		pods:     make(map[string]*PodInfo),
		active:   make(map[string]context.CancelFunc),
	}
}

func (c *Collector) Run(ctx context.Context) {
	factory := informers.NewSharedInformerFactory(c.client, 5*time.Minute)
	podInformer := factory.Core().V1().Pods().Informer()

	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if pod, ok := obj.(*corev1.Pod); ok {
				c.handlePod(pod)
			}
		},
		UpdateFunc: func(_, obj interface{}) {
			if pod, ok := obj.(*corev1.Pod); ok {
				c.handlePod(pod)
			}
		},
		DeleteFunc: func(obj interface{}) {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
					pod, ok = tombstone.Obj.(*corev1.Pod)
					if !ok {
						return
					}
				} else {
					return
				}
			}
			key := pod.Namespace + "/" + pod.Name
			c.mu.Lock()
			if cancel, ok := c.active[key]; ok {
				cancel()
				delete(c.active, key)
			}
			delete(c.pods, key)
			c.mu.Unlock()
		},
	})

	factory.Start(ctx.Done())
	factory.WaitForCacheSync(ctx.Done())
	<-ctx.Done()
}

func (c *Collector) handlePod(pod *corev1.Pod) {
	cfg := c.cfgMgr.Get()
	for _, ns := range cfg.ExcludeNamespaces {
		if strings.TrimSpace(ns) == pod.Namespace {
			return
		}
	}
	key := pod.Namespace + "/" + pod.Name

	c.mu.Lock()
	c.pods[key] = &PodInfo{
		Namespace: pod.Namespace,
		Name:      pod.Name,
		Status:    string(pod.Status.Phase),
		Node:      pod.Spec.NodeName,
	}

	switch pod.Status.Phase {
	case corev1.PodRunning:
		if _, already := c.active[key]; !already {
			ctx, cancel := context.WithCancel(context.Background())
			c.active[key] = cancel
			containers := append(pod.Spec.InitContainers, pod.Spec.Containers...)
			c.mu.Unlock()
			for _, ctr := range containers {
				go streamContainer(ctx, c.client, c.store, c.hub, c.redactor, pod.Namespace, pod.Name, ctr.Name)
			}
			slog.Info("attach", "pod", key, "containers", len(containers))
			return
		}
	case corev1.PodSucceeded, corev1.PodFailed:
		if cancel, ok := c.active[key]; ok {
			cancel()
			delete(c.active, key)
			slog.Info("detach", "pod", key)
		}
	}
	c.mu.Unlock()
}

// ApplyExclusions updates the namespace exclusion list and immediately stops
// any active streams for namespaces that are now excluded.
func (c *Collector) ApplyExclusions(excludeNS []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	exc := make(map[string]bool, len(excludeNS))
	for _, ns := range excludeNS {
		exc[strings.TrimSpace(ns)] = true
	}

	var stopped int
	for key, cancel := range c.active {
		ns := strings.SplitN(key, "/", 2)[0]
		if exc[ns] {
			cancel()
			delete(c.active, key)
			delete(c.pods, key)
			stopped++
		}
	}
	if stopped > 0 {
		slog.Info("exclusion update: stopped streams", "count", stopped)
	}
}

func (c *Collector) ListPods() []*PodInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]*PodInfo, 0, len(c.pods))
	for _, p := range c.pods {
		cp := *p
		cp.HasLogs = c.store.HasLogs(cp.Namespace, cp.Name)
		out = append(out, &cp)
	}
	return out
}
