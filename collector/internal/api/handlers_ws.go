package api

import (
	"log/slog"
	"net/http"
	"time"
)

const wsWriteDeadline = 5 * time.Second

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("namespace")
	pod := r.URL.Query().Get("pod")

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Warn("ws upgrade failed", "err", err)
		return
	}
	defer conn.Close()

	filter := ""
	if ns != "" && pod != "" {
		filter = ns + "/" + pod
	}

	client := s.hub.Register(filter)
	defer s.hub.Unregister(client)

	// Send recent history to seed the client's view before live lines start
	if ns != "" && pod != "" {
		if lines, err := s.store.ReadLines(ns, pod, 200, "", ""); err == nil {
			for _, line := range lines {
				conn.SetWriteDeadline(time.Now().Add(wsWriteDeadline))
				if err := conn.WriteMessage(1, []byte(line)); err != nil {
					return
				}
			}
		}
	}

	// Drain reads so the connection stays alive (handles pings and close frames)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case line, ok := <-client.Send():
			if !ok {
				return
			}
			conn.SetWriteDeadline(time.Now().Add(wsWriteDeadline))
			if err := conn.WriteMessage(1, []byte(line)); err != nil {
				return
			}
		}
	}
}
