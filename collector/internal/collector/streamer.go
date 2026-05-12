package collector

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	// nil on first connect → k8s returns all logs from container start.
	// Updated to last-seen timestamp on reconnect to resume without gaps or duplicates.
	var sinceTime *metav1.Time

	for {
		if err := ctx.Err(); err != nil {
			return
		}

		opts := &corev1.PodLogOptions{
			Container:  container,
			Follow:     true,
			Timestamps: true,
			SinceTime:  sinceTime,
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
			raw := scanner.Text()
			if t := parseLogTimestamp(raw); !t.IsZero() {
				ts := metav1.NewTime(t.Add(time.Nanosecond))
				sinceTime = &ts
			}
			line := r.Apply(raw)
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
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}

// parseLogTimestamp extracts the RFC3339Nano timestamp prefixed by the k8s log API
// when Timestamps:true is set. Format: "2006-01-02T15:04:05.999999999Z07:00 message..."
func parseLogTimestamp(line string) time.Time {
	i := strings.IndexByte(line, ' ')
	if i < 0 {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339Nano, line[:i])
	if err != nil {
		return time.Time{}
	}
	return t
}
