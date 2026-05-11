package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := defaults()
	if cfg.RetentionDays < 1 {
		t.Errorf("RetentionDays default %d < 1", cfg.RetentionDays)
	}
	if cfg.RotationMaxMB < 1 {
		t.Errorf("RotationMaxMB default %d < 1", cfg.RotationMaxMB)
	}
	if cfg.RotationKeepFiles < 1 {
		t.Errorf("RotationKeepFiles default %d < 1", cfg.RotationKeepFiles)
	}
	if cfg.CustomPatterns == nil {
		t.Error("CustomPatterns should not be nil")
	}
	if cfg.ExcludeNamespaces == nil {
		t.Error("ExcludeNamespaces should not be nil")
	}
}

func TestDefaultsFromEnv(t *testing.T) {
	t.Setenv("NEURALOG_RETENTION_DAYS", "14")
	t.Setenv("NEURALOG_REDACT_ENABLED", "false")
	t.Setenv("NEURALOG_EXCLUDE_NAMESPACES", "alpha,beta")

	cfg := defaults()

	if cfg.RetentionDays != 14 {
		t.Errorf("RetentionDays = %d, want 14", cfg.RetentionDays)
	}
	if cfg.RedactEnabled {
		t.Error("RedactEnabled should be false when env is 'false'")
	}
	if len(cfg.ExcludeNamespaces) != 2 || cfg.ExcludeNamespaces[0] != "alpha" {
		t.Errorf("ExcludeNamespaces = %v, want [alpha beta]", cfg.ExcludeNamespaces)
	}
}

func TestUpdateAndLoad(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	want := Config{
		RetentionDays:     30,
		RotationMaxMB:     50,
		RotationKeepFiles: 3,
		StorageQuotaGB:    10,
		RedactEnabled:     false,
		ExcludeNamespaces: []string{"staging"},
		CustomPatterns: []RedactPattern{
			{ID: "p1", Pattern: `secret-\d+`, Replace: "[X]"},
		},
	}

	if err := m.Update(want); err != nil {
		t.Fatalf("Update: %v", err)
	}

	// File must exist
	if _, err := os.Stat(filepath.Join(dir, ".neuralog.json")); err != nil {
		t.Fatalf("config file not written: %v", err)
	}

	got := m.Get()
	if got.RetentionDays != want.RetentionDays {
		t.Errorf("RetentionDays = %d, want %d", got.RetentionDays, want.RetentionDays)
	}
	if got.RedactEnabled != want.RedactEnabled {
		t.Errorf("RedactEnabled = %v, want %v", got.RedactEnabled, want.RedactEnabled)
	}
	if len(got.CustomPatterns) != 1 || got.CustomPatterns[0].ID != "p1" {
		t.Errorf("CustomPatterns = %v", got.CustomPatterns)
	}
}

func TestLoadFromDisk(t *testing.T) {
	dir := t.TempDir()
	m1 := NewManager(dir)
	if err := m1.Update(Config{RetentionDays: 21, RedactEnabled: true, CustomPatterns: []RedactPattern{}, ExcludeNamespaces: []string{}}); err != nil {
		t.Fatal(err)
	}

	// Fresh manager from same dir must load persisted config
	m2 := NewManager(dir)
	if got := m2.Get().RetentionDays; got != 21 {
		t.Errorf("loaded RetentionDays = %d, want 21", got)
	}
}
