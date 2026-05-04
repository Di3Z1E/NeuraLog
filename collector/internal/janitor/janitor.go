package janitor

import (
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

func Run(basePath string, retentionDays int) error {
	threshold := time.Now().AddDate(0, 0, -retentionDays)
	deleted := 0

	slog.Info("janitor starting", "basePath", basePath, "retentionDays", retentionDays, "threshold", threshold.Format(time.RFC3339))

	err := filepath.WalkDir(basePath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if d.IsDir() || filepath.Ext(path) != ".log" {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.ModTime().Before(threshold) {
			if err := os.Remove(path); err == nil {
				deleted++
				slog.Info("deleted", "file", path, "age", time.Since(info.ModTime()).Round(time.Hour))
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	pruneEmptyDirs(basePath)
	slog.Info("janitor done", "deleted", deleted)
	return nil
}

func pruneEmptyDirs(basePath string) {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(basePath, e.Name())
		children, err := os.ReadDir(dir)
		if err != nil || len(children) > 0 {
			continue
		}
		if err := os.Remove(dir); err == nil {
			slog.Info("pruned empty dir", "dir", dir)
		}
	}
}
