package store

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Di3Z1E/neuralog/internal/config"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	cfgMgr := config.NewManager(dir)
	return New(filepath.Join(dir, "logs"), cfgMgr)
}

func logLine(container string, ts time.Time, msg string) string {
	return fmt.Sprintf("[%s] %s %s", container, ts.Format(time.RFC3339Nano), msg)
}

func TestAppend(t *testing.T) {
	st := newTestStore(t)
	defer st.Close()

	if err := st.Append("default", "mypod", "hello"); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if !st.HasLogs("default", "mypod") {
		t.Error("HasLogs should return true after Append")
	}
}

func TestReadLines_basic(t *testing.T) {
	st := newTestStore(t)
	defer st.Close()

	for i := 0; i < 5; i++ {
		if err := st.Append("default", "pod1", fmt.Sprintf("line %d", i)); err != nil {
			t.Fatal(err)
		}
	}

	lines, err := st.ReadLines("default", "pod1", 0, "", "", nil, nil)
	if err != nil {
		t.Fatalf("ReadLines: %v", err)
	}
	if len(lines) != 5 {
		t.Errorf("got %d lines, want 5", len(lines))
	}
}

func TestReadLines_limit(t *testing.T) {
	st := newTestStore(t)
	defer st.Close()

	for i := 0; i < 20; i++ {
		st.Append("default", "pod1", fmt.Sprintf("line %d", i))
	}

	lines, err := st.ReadLines("default", "pod1", 5, "", "", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 5 {
		t.Errorf("got %d lines, want 5", len(lines))
	}
	// limit returns tail
	if lines[4] != "line 19" {
		t.Errorf("last line = %q, want 'line 19'", lines[4])
	}
}

func TestReadLines_search(t *testing.T) {
	st := newTestStore(t)
	defer st.Close()

	st.Append("default", "pod1", "error: connection refused")
	st.Append("default", "pod1", "info: server started")
	st.Append("default", "pod1", "error: timeout")

	lines, err := st.ReadLines("default", "pod1", 0, "error", "", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 2 {
		t.Errorf("search=error: got %d lines, want 2", len(lines))
	}
}

func TestReadLines_level(t *testing.T) {
	st := newTestStore(t)
	defer st.Close()

	st.Append("default", "pod1", "2024-01-01 INFO server started")
	st.Append("default", "pod1", "2024-01-01 WARN high memory")
	st.Append("default", "pod1", "2024-01-01 ERROR crash")

	lines, err := st.ReadLines("default", "pod1", 0, "", "ERROR", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 1 {
		t.Errorf("level=ERROR: got %d lines, want 1", len(lines))
	}
}

func TestReadLines_timeRange(t *testing.T) {
	st := newTestStore(t)
	defer st.Close()

	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		ts := base.Add(time.Duration(i) * time.Hour)
		st.Append("default", "pod1", logLine("app", ts, fmt.Sprintf("event %d", i)))
	}

	from := base.Add(1 * time.Hour)
	to := base.Add(3 * time.Hour)
	lines, err := st.ReadLines("default", "pod1", 0, "", "", &from, &to)
	if err != nil {
		t.Fatal(err)
	}
	// events at +1h, +2h, +3h → 3 lines
	if len(lines) != 3 {
		t.Errorf("time range: got %d lines, want 3", len(lines))
	}
}

func TestReadLines_notFound(t *testing.T) {
	st := newTestStore(t)
	defer st.Close()

	_, err := st.ReadLines("missing", "pod", 0, "", "", nil, nil)
	if err == nil {
		t.Error("expected error for missing log file")
	}
}

func TestListPods(t *testing.T) {
	st := newTestStore(t)
	defer st.Close()

	st.Append("ns1", "alpha", "x")
	st.Append("ns1", "beta", "x")
	st.Append("ns2", "gamma", "x")

	pods, err := st.ListPods()
	if err != nil {
		t.Fatal(err)
	}
	if len(pods) != 3 {
		t.Errorf("got %d pods, want 3", len(pods))
	}
}

func TestDiskUsageBytes(t *testing.T) {
	st := newTestStore(t)
	defer st.Close()

	st.Append("default", "pod1", "hello world")
	used := st.DiskUsageBytes()
	if used == 0 {
		t.Error("DiskUsageBytes should be > 0 after Append")
	}
}

func TestRotation(t *testing.T) {
	dir := t.TempDir()
	cfgMgr := config.NewManager(dir)
	// Set rotation at 1 byte to trigger immediately
	cfgMgr.Update(config.Config{
		RotationMaxMB:     0, // 0 disables rotation in normal flow, so write enough bytes
		RotationKeepFiles: 2,
		RetentionDays:     7,
		RedactEnabled:     false,
		CustomPatterns:    []config.RedactPattern{},
		ExcludeNamespaces: []string{},
	})

	logDir := filepath.Join(dir, "logs")
	st := New(logDir, cfgMgr)
	defer st.Close()

	// Write enough to fill >1 MB by directly manipulating via cfgMgr
	cfgMgr.Update(config.Config{
		RotationMaxMB:     1,
		RotationKeepFiles: 2,
		RetentionDays:     7,
		RedactEnabled:     false,
		CustomPatterns:    []config.RedactPattern{},
		ExcludeNamespaces: []string{},
	})

	// Write 1MB+1 of data to trigger rotation
	bigLine := make([]byte, 1024)
	for i := range bigLine {
		bigLine[i] = 'x'
	}
	for i := 0; i < 1025; i++ {
		if err := st.Append("default", "pod1", string(bigLine)); err != nil {
			t.Fatal(err)
		}
	}

	// After rotation, a .log.1 backup should exist
	rotated := filepath.Join(logDir, "default", "pod1.log.1")
	if _, err := os.Stat(rotated); err != nil {
		t.Errorf("expected rotated file %s: %v", rotated, err)
	}
}

func TestParseLineTime(t *testing.T) {
	ts := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	line := logLine("mycontainer", ts, "something happened")

	got, ok := parseLineTime(line)
	if !ok {
		t.Fatalf("parseLineTime(%q) returned ok=false", line)
	}
	if !got.Equal(ts) {
		t.Errorf("parseLineTime = %v, want %v", got, ts)
	}
}

func TestParseLineTime_invalid(t *testing.T) {
	cases := []string{
		"no bracket here at all",
		"[container]no space after bracket",
		"[container] notadate rest",
	}
	for _, c := range cases {
		if _, ok := parseLineTime(c); ok {
			t.Errorf("parseLineTime(%q) should return ok=false", c)
		}
	}
}
