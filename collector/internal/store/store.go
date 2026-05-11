package store

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Di3Z1E/neuralog/internal/config"
)

type fileHandle struct {
	f    *os.File
	size int64
}

type Store struct {
	basePath string
	cfgMgr   *config.Manager
	mu       sync.Mutex
	handles  map[string]*fileHandle
}

type PodRef struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

func New(basePath string, cfgMgr *config.Manager) *Store {
	return &Store{
		basePath: basePath,
		cfgMgr:   cfgMgr,
		handles:  make(map[string]*fileHandle),
	}
}

func (s *Store) openHandle(ns, pod string) (*fileHandle, error) {
	dir := filepath.Join(s.basePath, ns)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, pod+".log")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	info, _ := f.Stat()
	return &fileHandle{f: f, size: info.Size()}, nil
}

func (s *Store) Append(ns, pod, line string) error {
	key := ns + "/" + pod
	s.mu.Lock()
	defer s.mu.Unlock()

	fh, ok := s.handles[key]
	if !ok {
		var err error
		fh, err = s.openHandle(ns, pod)
		if err != nil {
			return err
		}
		s.handles[key] = fh
	}

	data := line + "\n"
	if _, err := fh.f.WriteString(data); err != nil {
		return err
	}
	fh.size += int64(len(data))

	cfg := s.cfgMgr.Get()
	if cfg.RotationMaxMB > 0 && fh.size >= int64(cfg.RotationMaxMB)*1024*1024 {
		s.rotateUnlocked(ns, pod, cfg.RotationKeepFiles)
	}

	return nil
}

// rotateUnlocked renames pod.log → pod.log.1 → … and must be called with s.mu held.
func (s *Store) rotateUnlocked(ns, pod string, keep int) {
	if keep < 1 {
		keep = 1
	}
	key := ns + "/" + pod
	base := filepath.Join(s.basePath, ns, pod+".log")

	if fh, ok := s.handles[key]; ok {
		fh.f.Close()
		delete(s.handles, key)
	}

	// Remove oldest beyond the keep limit, then shift
	os.Remove(fmt.Sprintf("%s.%d", base, keep))
	for i := keep - 1; i >= 1; i-- {
		os.Rename(fmt.Sprintf("%s.%d", base, i), fmt.Sprintf("%s.%d", base, i+1))
	}
	if err := os.Rename(base, base+".1"); err == nil {
		slog.Info("rotated", "pod", key)
	}
}

func (s *Store) HasLogs(ns, pod string) bool {
	_, err := os.Stat(filepath.Join(s.basePath, ns, pod+".log"))
	return err == nil
}

func (s *Store) FilePath(ns, pod string) string {
	return filepath.Join(s.basePath, ns, pod+".log")
}

// parseLineTime extracts the RFC3339Nano timestamp from a stored log line.
// Stored format: "[container] 2006-01-02T15:04:05.999999999Z message..."
func parseLineTime(line string) (time.Time, bool) {
	i := strings.Index(line, "] ")
	if i < 0 {
		return time.Time{}, false
	}
	rest := line[i+2:]
	j := strings.IndexByte(rest, ' ')
	if j < 0 {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339Nano, rest[:j])
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func (s *Store) ReadLines(ns, pod string, limit int, search, level string, from, to *time.Time) ([]string, error) {
	f, err := os.Open(s.FilePath(ns, pod))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var all []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if from != nil || to != nil {
			if t, ok := parseLineTime(line); ok {
				if from != nil && t.Before(*from) {
					continue
				}
				if to != nil && t.After(*to) {
					continue
				}
			}
		}
		if search != "" && !strings.Contains(strings.ToLower(line), strings.ToLower(search)) {
			continue
		}
		if level != "" && !strings.Contains(strings.ToUpper(line), strings.ToUpper(level)) {
			continue
		}
		all = append(all, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if limit > 0 && len(all) > limit {
		all = all[len(all)-limit:]
	}
	return all, nil
}

func (s *Store) ListPods() ([]PodRef, error) {
	var pods []PodRef
	nsDirs, err := os.ReadDir(s.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return pods, nil
		}
		return nil, err
	}
	for _, nsDir := range nsDirs {
		if !nsDir.IsDir() {
			continue
		}
		files, err := os.ReadDir(filepath.Join(s.basePath, nsDir.Name()))
		if err != nil {
			continue
		}
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".log") {
				pods = append(pods, PodRef{
					Namespace: nsDir.Name(),
					Name:      strings.TrimSuffix(f.Name(), ".log"),
				})
			}
		}
	}
	return pods, nil
}

// DiskUsageBytes returns the total bytes used under basePath for .log files.
func (s *Store) DiskUsageBytes() int64 {
	var total int64
	filepath.WalkDir(s.basePath, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		name := d.Name()
		if strings.HasSuffix(name, ".log") || strings.Contains(name, ".log.") {
			if info, err := d.Info(); err == nil {
				total += info.Size()
			}
		}
		return nil
	})
	return total
}

func (s *Store) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, fh := range s.handles {
		fh.f.Close()
	}
	s.handles = make(map[string]*fileHandle)
}
