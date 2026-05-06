package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

func (s *Server) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ns, pod := vars["namespace"], vars["pod"]

	limit := 500
	if l := r.URL.Query().Get("lines"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 5000 {
			limit = n
		}
	}
	search := r.URL.Query().Get("search")
	level := r.URL.Query().Get("level")
	if level == "ALL" {
		level = ""
	}

	lines, err := s.store.ReadLines(ns, pod, limit, search, level)
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
