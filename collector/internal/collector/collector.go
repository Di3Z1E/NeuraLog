package collector

import (
	"context"
	"log/slog"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

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
	client    kubernetes.Interface
	store     *store.Store
	hub       *hub.Hub
	redactor  *redactor.Redactor
	excludeNS map[string]bool

	mu     sync.RWMutex
	pods   map[string]*PodInfo           // key: "ns/pod"
	active map[string]context.CancelFunc // key: "ns/pod"
}

func New(
	client kubernetes.Interface,
	st *store.Store,
	h *hub.Hub,
	r *redactor.Redactor,
	excludeNS []string,
) *Collector {
	exc := make(map[string]bool, len(excludeNS))
	for _, ns := range excludeNS {
		exc[ns] = true
	}
	return &Collector{
		client:    client,
		store:     st,
		hub:       h,
		redactor:  r,
		excludeNS: exc,
		pods:      make(map[string]*PodInfo),
		active:    make(map[string]context.CancelFunc),
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
			c.stopStream(key)
			c.mu.Lock()
			delete(c.pods, key)
			c.mu.Unlock()
		},
	})

	factory.Start(ctx.Done())
	factory.WaitForCacheSync(ctx.Done())
	<-ctx.Done()
}

func (c *Collector) handlePod(pod *corev1.Pod) {
	if c.excludeNS[pod.Namespace] {
		return
	}
	key := pod.Namespace + "/" + pod.Name

	c.mu.Lock()
	c.pods[key] = &PodInfo{
		Namespace: pod.Namespace,
		Name:      pod.Name,
		Status:    string(pod.Status.Phase),
		Node:      pod.Spec.NodeName,
		HasLogs:   c.store.HasLogs(pod.Namespace, pod.Name),
	}
	c.mu.Unlock()

	switch pod.Status.Phase {
	case corev1.PodRunning:
		c.mu.RLock()
		_, already := c.active[key]
		c.mu.RUnlock()
		if !already {
			c.startStream(pod)
		}
	case corev1.PodSucceeded, corev1.PodFailed:
		c.stopStream(key)
	}
}

func (c *Collector) startStream(pod *corev1.Pod) {
	ctx, cancel := context.WithCancel(context.Background())
	key := pod.Namespace + "/" + pod.Name

	c.mu.Lock()
	c.active[key] = cancel
	c.mu.Unlock()

	containers := append(pod.Spec.InitContainers, pod.Spec.Containers...)
	for _, ctr := range containers {
		go streamContainer(ctx, c.client, c.store, c.hub, c.redactor, pod.Namespace, pod.Name, ctr.Name)
	}
	slog.Info("attach", "pod", key, "containers", len(containers))
}

func (c *Collector) stopStream(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if cancel, ok := c.active[key]; ok {
		cancel()
		delete(c.active, key)
		slog.Info("detach", "pod", key)
	}
}

func (c *Collector) ListPods() []*PodInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]*PodInfo, 0, len(c.pods))
	for _, p := range c.pods {
		cp := *p
		out = append(out, &cp)
	}
	return out
}
