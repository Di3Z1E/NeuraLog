package collector

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/Di3Z1E/neuralog/internal/hub"
	"github.com/Di3Z1E/neuralog/internal/redactor"
	"github.com/Di3Z1E/neuralog/internal/store"
)

func streamContainer(
	ctx context.Context,
	client kubernetes.Interface,
	st *store.Store,
	h *hub.Hub,
	r *redactor.Redactor,
	ns, pod, container string,
) {
	since := int64(60)
	opts := &corev1.PodLogOptions{
		Container:    container,
		Follow:       true,
		Timestamps:   true,
		SinceSeconds: &since,
	}

	for {
		if err := ctx.Err(); err != nil {
			return
		}

		stream, err := client.CoreV1().Pods(ns).GetLogs(pod, opts).Stream(ctx)
		if err != nil {
			slog.Warn("stream open failed", "pod", ns+"/"+pod, "container", container, "err", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}

		scanner := bufio.NewScanner(stream)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				stream.Close()
				return
			default:
			}
			line := r.Apply(scanner.Text())
			entry := fmt.Sprintf("[%s] %s", container, line)
			if err := st.Append(ns, pod, entry); err != nil {
				slog.Warn("store append failed", "pod", ns+"/"+pod, "err", err)
			}
			h.Broadcast(ns+"/"+pod, entry)
		}
		stream.Close()

		if ctx.Err() != nil {
			return
		}
		// Stream ended cleanly (pod restarted etc.) — retry after a short pause
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}
