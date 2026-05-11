package api

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/Di3Z1E/neuralog/internal/ui"
)

func (s *Server) registerRoutes() {
	s.router.Use(logMiddleware, recoveryMiddleware, securityHeaders)

	s.router.HandleFunc("/healthz", s.handleHealth).Methods("GET")
	s.router.HandleFunc("/ws", s.handleWS).Methods("GET")

	api := s.router.PathPrefix("/api/v1").Subrouter()
	api.Use(corsMiddleware)
	api.HandleFunc("/pods", s.handleListPods).Methods("GET", "OPTIONS")
	api.HandleFunc("/logs/{namespace}/{pod}", s.handleGetLogs).Methods("GET", "OPTIONS")
	api.HandleFunc("/download/{namespace}/{pod}", s.handleDownload).Methods("GET")
	api.HandleFunc("/config", s.handleGetConfig).Methods("GET", "OPTIONS")
	api.HandleFunc("/config", s.handlePutConfig).Methods("PUT", "OPTIONS")

	// Hashed Vite assets get long-lived immutable cache headers.
	uiFS := ui.FS()
	s.router.PathPrefix("/assets/").Handler(
		cacheImmutable(http.StripPrefix("/", http.FileServer(http.FS(uiFS)))),
	)

	// SPA catch-all: serve the requested file if it exists, otherwise index.html.
	s.router.PathPrefix("/").Handler(spaHandler(uiFS))
}

func spaHandler(uiFS fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(uiFS))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := path.Clean("/" + r.URL.Path)
		name = strings.TrimPrefix(name, "/")
		if name == "" {
			name = "index.html"
		}

		if f, err := uiFS.Open(name); err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// Unknown path: serve root so React Router can handle it client-side.
		// We cannot rewrite to /index.html because http.FileServer redirects
		// any path ending in /index.html back to "./" by design.
		r = r.Clone(r.Context())
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}
