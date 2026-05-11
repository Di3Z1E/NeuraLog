package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/Di3Z1E/neuralog/internal/api"
	"github.com/Di3Z1E/neuralog/internal/collector"
	"github.com/Di3Z1E/neuralog/internal/config"
	"github.com/Di3Z1E/neuralog/internal/hub"
	"github.com/Di3Z1E/neuralog/internal/janitor"
	"github.com/Di3Z1E/neuralog/internal/redactor"
	"github.com/Di3Z1E/neuralog/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cmd := "serve"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "serve":
		runServe()
	case "janitor":
		runJanitor()
	default:
		slog.Error("unknown subcommand", "cmd", cmd, "usage", "neuralog [serve|janitor]")
		os.Exit(1)
	}
}

func runServe() {
	logBase := envOr("NEURALOG_LOG_BASE_PATH", "/mnt/logs")
	listenAddr := envOr("NEURALOG_LISTEN_ADDR", ":8080")
	kubeconfig := os.Getenv("KUBECONFIG")

	cfgMgr := config.NewManager(logBase)

	var (
		k8sCfg *rest.Config
		err    error
	)
	if kubeconfig == "" {
		k8sCfg, err = rest.InClusterConfig()
	} else {
		k8sCfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if err != nil {
		slog.Error("k8s config failed", "err", err)
		os.Exit(1)
	}

	k8sClient, err := kubernetes.NewForConfig(k8sCfg)
	if err != nil {
		slog.Error("k8s client failed", "err", err)
		os.Exit(1)
	}

	st := store.New(logBase, cfgMgr)
	h := hub.New()
	r := redactor.New(cfgMgr)
	col := collector.New(k8sClient, st, h, r, cfgMgr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go h.Run()
	go col.Run(ctx)
	go runQuotaWatcher(ctx, cfgMgr, st, logBase)

	cfg := cfgMgr.Get()
	srv := &http.Server{
		Addr:        listenAddr,
		Handler:     api.NewServer(col, st, h, cfgMgr, r),
		ReadTimeout: 30 * time.Second,
		// WriteTimeout intentionally unset — streaming responses must not time out
	}

	go func() {
		slog.Info("server started", "addr", listenAddr, "redaction", cfg.RedactEnabled)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down gracefully")
	cancel()
	shutCtx, done := context.WithTimeout(context.Background(), 15*time.Second)
	defer done()
	_ = srv.Shutdown(shutCtx)
	st.Close()
}

func runJanitor() {
	logBase := envOr("NEURALOG_LOG_BASE_PATH", "/mnt/logs")
	cfgMgr := config.NewManager(logBase)
	if err := janitor.Run(logBase, cfgMgr); err != nil {
		slog.Error("janitor failed", "err", err)
		os.Exit(1)
	}
}

// runQuotaWatcher enforces the storage quota by deleting oldest log files
// when total disk usage exceeds the configured limit.
func runQuotaWatcher(ctx context.Context, cfgMgr *config.Manager, st *store.Store, logBase string) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cfg := cfgMgr.Get()
			if cfg.StorageQuotaGB <= 0 {
				continue
			}
			maxBytes := int64(cfg.StorageQuotaGB * 1024 * 1024 * 1024)
			if used := st.DiskUsageBytes(); used > maxBytes {
				enforceQuota(logBase, maxBytes)
			}
		}
	}
}

func enforceQuota(logBase string, maxBytes int64) {
	type logFile struct {
		path  string
		mtime time.Time
		size  int64
	}
	var files []logFile
	var total int64

	filepath.WalkDir(logBase, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		name := d.Name()
		if !strings.HasSuffix(name, ".log") && !strings.Contains(name, ".log.") {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		files = append(files, logFile{path, info.ModTime(), info.Size()})
		total += info.Size()
		return nil
	})

	if total <= maxBytes {
		return
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].mtime.Before(files[j].mtime)
	})

	for _, f := range files {
		if total <= maxBytes {
			break
		}
		if err := os.Remove(f.path); err == nil {
			slog.Info("quota: evicted", "file", f.path, "freed_mb", f.size/1024/1024)
			total -= f.size
		}
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
