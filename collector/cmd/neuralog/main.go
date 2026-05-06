package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/Di3Z1E/neuralog/internal/api"
	"github.com/Di3Z1E/neuralog/internal/collector"
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
	excludeNS := strings.Split(envOr("NEURALOG_EXCLUDE_NAMESPACES", "log-system,kube-system"), ",")
	redactEnabled := envOr("NEURALOG_REDACT_ENABLED", "true") != "false"

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

	st := store.New(logBase)
	h := hub.New()
	r := redactor.New(redactEnabled)
	col := collector.New(k8sClient, st, h, r, excludeNS)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go h.Run()
	go col.Run(ctx)

	srv := &http.Server{
		Addr:        listenAddr,
		Handler:     api.NewServer(col, st, h),
		ReadTimeout: 30 * time.Second,
		// WriteTimeout intentionally unset — streaming responses must not time out
	}

	go func() {
		slog.Info("server started", "addr", listenAddr, "redaction", redactEnabled)
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
	days := 7
	if d := os.Getenv("NEURALOG_RETENTION_DAYS"); d != "" {
		if n, err := strconv.Atoi(d); err == nil && n > 0 {
			days = n
		}
	}
	if err := janitor.Run(logBase, days); err != nil {
		slog.Error("janitor failed", "err", err)
		os.Exit(1)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
