package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/Di3Z1E/neuralog/internal/collector"
	"github.com/Di3Z1E/neuralog/internal/hub"
	"github.com/Di3Z1E/neuralog/internal/store"
)

type Server struct {
	router   *mux.Router
	col      *collector.Collector
	store    *store.Store
	hub      *hub.Hub
	upgrader websocket.Upgrader
}

func NewServer(col *collector.Collector, st *store.Store, h *hub.Hub) http.Handler {
	s := &Server{
		router: mux.NewRouter(),
		col:    col,
		store:  st,
		hub:    h,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 4096,
			CheckOrigin:     func(*http.Request) bool { return true },
		},
	}
	s.registerRoutes()
	return s.router
}
