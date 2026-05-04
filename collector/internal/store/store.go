package store

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Store struct {
	basePath string
	mu       sync.Mutex
	handles  map[string]*os.File // key: "namespace/pod"
}

type PodRef struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

func New(basePath string) *Store {
	return &Store{
		basePath: basePath,
		handles:  make(map[string]*os.File),
	}
}

func (s *Store) Append(ns, pod, line string) error {
	key := ns + "/" + pod
	s.mu.Lock()
	defer s.mu.Unlock()

	f, ok := s.handles[key]
	if !ok {
		dir := filepath.Join(s.basePath, ns)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		var err error
		f, err = os.OpenFile(filepath.Join(dir, pod+".log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		s.handles[key] = f
	}

	_, err := f.WriteString(line + "\n")
	return err
}

func (s *Store) HasLogs(ns, pod string) bool {
	_, err := os.Stat(filepath.Join(s.basePath, ns, pod+".log"))
	return err == nil
}

func (s *Store) FilePath(ns, pod string) string {
	return filepath.Join(s.basePath, ns, pod+".log")
}

func (s *Store) ReadLines(ns, pod string, limit int, search, level string) ([]string, error) {
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

func (s *Store) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, f := range s.handles {
		f.Close()
	}
	s.handles = make(map[string]*os.File)
}
