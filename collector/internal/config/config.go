package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type RedactPattern struct {
	ID      string `json:"id"`
	Pattern string `json:"pattern"`
	Replace string `json:"replace"`
}

type Config struct {
	StorageQuotaGB    float64         `json:"storageQuotaGB"`
	RotationMaxMB     int             `json:"rotationMaxMB"`
	RotationKeepFiles int             `json:"rotationKeepFiles"`
	RetentionDays     int             `json:"retentionDays"`
	ExcludeNamespaces []string        `json:"excludeNamespaces"`
	RedactEnabled     bool            `json:"redactEnabled"`
	CustomPatterns    []RedactPattern `json:"customPatterns"`
}

// defaults seeds from env vars so existing deployments keep working on first boot.
func defaults() Config {
	excludeNS := []string{"log-system", "kube-system"}
	if v := os.Getenv("NEURALOG_EXCLUDE_NAMESPACES"); v != "" {
		excludeNS = strings.Split(v, ",")
	}
	redactEnabled := os.Getenv("NEURALOG_REDACT_ENABLED") != "false"
	retentionDays := 7
	if d := os.Getenv("NEURALOG_RETENTION_DAYS"); d != "" {
		if n, err := strconv.Atoi(d); err == nil && n > 0 {
			retentionDays = n
		}
	}
	return Config{
		StorageQuotaGB:    0,
		RotationMaxMB:     100,
		RotationKeepFiles: 5,
		RetentionDays:     retentionDays,
		ExcludeNamespaces: excludeNS,
		RedactEnabled:     redactEnabled,
		CustomPatterns:    []RedactPattern{},
	}
}

type Manager struct {
	path string
	mu   sync.RWMutex
	cfg  Config
}

func NewManager(basePath string) *Manager {
	m := &Manager{
		path: filepath.Join(basePath, ".neuralog.json"),
		cfg:  defaults(),
	}
	_ = m.load()
	return m
}

func (m *Manager) Get() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cfg
}

func (m *Manager) Update(cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(m.path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(m.path, data, 0644); err != nil {
		return err
	}
	m.mu.Lock()
	m.cfg = cfg
	m.mu.Unlock()
	return nil
}

func (m *Manager) load() error {
	data, err := os.ReadFile(m.path)
	if err != nil {
		if os.IsNotExist(err) {
			return m.Update(defaults())
		}
		return err
	}
	cfg := defaults()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}
	m.mu.Lock()
	m.cfg = cfg
	m.mu.Unlock()
	return nil
}
