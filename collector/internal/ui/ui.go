package ui

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var embedded embed.FS

// FS returns a filesystem rooted at the dist directory.
// In production the dist is populated by the Docker build.
// Locally (dev) only .gitkeep exists; API works, UI returns 404 — use Vite on :3000.
func FS() fs.FS {
	sub, _ := fs.Sub(embedded, "dist")
	return sub
}
