package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/Di3Z1E/neuralog/internal/config"
)

type configResponse struct {
	config.Config
	StorageUsedGB float64 `json:"storageUsedGB"`
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	cfg := s.cfgMgr.Get()
	usedBytes := s.store.DiskUsageBytes()
	resp := configResponse{
		Config:        cfg,
		StorageUsedGB: float64(usedBytes) / (1024 * 1024 * 1024),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handlePutConfig(w http.ResponseWriter, r *http.Request) {
	var cfg config.Config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusBadRequest)
		return
	}

	// Clamp / validate
	if cfg.RetentionDays < 1 {
		cfg.RetentionDays = 1
	}
	if cfg.RotationKeepFiles < 1 {
		cfg.RotationKeepFiles = 1
	}
	if cfg.StorageQuotaGB < 0 {
		cfg.StorageQuotaGB = 0
	}
	if cfg.RotationMaxMB < 0 {
		cfg.RotationMaxMB = 0
	}
	if cfg.CustomPatterns == nil {
		cfg.CustomPatterns = []config.RedactPattern{}
	}
	if cfg.ExcludeNamespaces == nil {
		cfg.ExcludeNamespaces = []string{}
	}
	for _, p := range cfg.CustomPatterns {
		if _, err := regexp.Compile(p.Pattern); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"invalid pattern %q: %s"}`, p.Pattern, err), http.StatusBadRequest)
			return
		}
	}

	if err := s.cfgMgr.Update(cfg); err != nil {
		http.Error(w, `{"error":"failed to save config"}`, http.StatusInternalServerError)
		return
	}

	// Hot-reload dependent components immediately
	s.redactor.Reload()
	s.col.ApplyExclusions(cfg.ExcludeNamespaces)

	usedBytes := s.store.DiskUsageBytes()
	resp := configResponse{
		Config:        cfg,
		StorageUsedGB: float64(usedBytes) / (1024 * 1024 * 1024),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
