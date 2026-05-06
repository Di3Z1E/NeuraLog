package api

import (
	"encoding/json"
	"net/http"

	"github.com/Di3Z1E/neuralog/internal/collector"
)

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleListPods(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("namespace")

	pods := s.col.ListPods()

	// Filter by namespace if requested
	if ns != "" {
		filtered := pods[:0]
		for _, p := range pods {
			if p.Namespace == ns {
				filtered = append(filtered, p)
			}
		}
		pods = filtered
	}

	// Merge in pods that have logs but are no longer live
	storedPods, _ := s.store.ListPods()
	seen := make(map[string]bool, len(pods))
	for _, p := range pods {
		seen[p.Namespace+"/"+p.Name] = true
	}
	for _, sp := range storedPods {
		if ns != "" && sp.Namespace != ns {
			continue
		}
		if !seen[sp.Namespace+"/"+sp.Name] {
			pods = append(pods, &collector.PodInfo{
				Namespace: sp.Namespace,
				Name:      sp.Name,
				Status:    "stopped",
				HasLogs:   true,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"pods": pods})
}
