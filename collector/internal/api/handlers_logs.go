package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// parseTimeParam parses an RFC3339 query param; returns nil if absent or invalid.
func parseTimeParam(r *http.Request, key string) *time.Time {
	v := r.URL.Query().Get(key)
	if v == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return nil
	}
	return &t
}

func (s *Server) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ns, pod := vars["namespace"], vars["pod"]

	from := parseTimeParam(r, "from")
	to := parseTimeParam(r, "to")

	// When a date range is active, return all matching lines (no default cap).
	limit := 500
	if from != nil || to != nil {
		limit = 0
	}
	if l := r.URL.Query().Get("lines"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 10000 {
			limit = n
		}
	}

	search := r.URL.Query().Get("search")
	level := r.URL.Query().Get("level")
	if level == "ALL" {
		level = ""
	}

	lines, err := s.store.ReadLines(ns, pod, limit, search, level, from, to)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, `{"error":"no logs found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	if lines == nil {
		lines = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"namespace": ns,
		"pod":       pod,
		"lines":     lines,
		"count":     len(lines),
	})
}
