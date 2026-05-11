package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/Di3Z1E/neuralog/internal/collector"
	"github.com/Di3Z1E/neuralog/internal/config"
	"github.com/Di3Z1E/neuralog/internal/hub"
	"github.com/Di3Z1E/neuralog/internal/redactor"
	"github.com/Di3Z1E/neuralog/internal/store"
)

type Server struct {
	router   *mux.Router
	col      *collector.Collector
	store    *store.Store
	hub      *hub.Hub
	cfgMgr   *config.Manager
	redactor *redactor.Redactor
	upgrader websocket.Upgrader
}

func NewServer(
	col *collector.Collector,
	st *store.Store,
	h *hub.Hub,
	cfgMgr *config.Manager,
	r *redactor.Redactor,
) http.Handler {
	s := &Server{
		router:   mux.NewRouter(),
		col:      col,
		store:    st,
		hub:      h,
		cfgMgr:   cfgMgr,
		redactor: r,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 4096,
			CheckOrigin:     func(*http.Request) bool { return true },
		},
	}
	s.registerRoutes()
	return s.router
}
