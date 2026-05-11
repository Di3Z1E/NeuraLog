package janitor

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Di3Z1E/neuralog/internal/config"
)

func newManager(t *testing.T, retentionDays int) *config.Manager {
	t.Helper()
	m := config.NewManager(t.TempDir())
	m.Update(config.Config{
		RetentionDays:     retentionDays,
		RedactEnabled:     false,
		CustomPatterns:    []config.RedactPattern{},
		ExcludeNamespaces: []string{},
		RotationMaxMB:     100,
		RotationKeepFiles: 5,
	})
	return m
}

func writeFileWithMtime(t *testing.T, path string, age time.Duration) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("log line\n"), 0644); err != nil {
		t.Fatal(err)
	}
	mtime := time.Now().Add(-age)
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatal(err)
	}
}

func TestRun_deletesOldFiles(t *testing.T) {
	base := t.TempDir()
	mgr := newManager(t, 7)

	old := filepath.Join(base, "ns1", "old-pod.log")
	writeFileWithMtime(t, old, 10*24*time.Hour) // 10 days old

	if err := Run(base, mgr); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Error("expected old log file to be deleted")
	}
}

func TestRun_keepsNewFiles(t *testing.T) {
	base := t.TempDir()
	mgr := newManager(t, 7)

	fresh := filepath.Join(base, "ns1", "fresh-pod.log")
	writeFileWithMtime(t, fresh, 2*24*time.Hour) // 2 days old

	if err := Run(base, mgr); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Errorf("expected fresh log file to remain: %v", err)
	}
}

func TestRun_deletesRotatedFiles(t *testing.T) {
	base := t.TempDir()
	mgr := newManager(t, 7)

	rotated := filepath.Join(base, "ns1", "pod.log.2")
	writeFileWithMtime(t, rotated, 15*24*time.Hour)

	if err := Run(base, mgr); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(rotated); !os.IsNotExist(err) {
		t.Error("expected old rotated file to be deleted")
	}
}

func TestRun_prunesEmptyDirs(t *testing.T) {
	base := t.TempDir()
	mgr := newManager(t, 7)

	emptyNS := filepath.Join(base, "empty-ns")
	if err := os.MkdirAll(emptyNS, 0755); err != nil {
		t.Fatal(err)
	}

	old := filepath.Join(emptyNS, "pod.log")
	writeFileWithMtime(t, old, 10*24*time.Hour)

	if err := Run(base, mgr); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(emptyNS); !os.IsNotExist(err) {
		t.Error("expected empty namespace dir to be pruned")
	}
}

func TestRun_emptyBasePath(t *testing.T) {
	base := t.TempDir()
	mgr := newManager(t, 7)

	// No files — should succeed silently
	if err := Run(base, mgr); err != nil {
		t.Fatalf("Run on empty dir: %v", err)
	}
}
