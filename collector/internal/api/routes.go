package api

func (s *Server) registerRoutes() {
	s.router.Use(logMiddleware, recoveryMiddleware)

	s.router.HandleFunc("/healthz", s.handleHealth).Methods("GET")
	s.router.HandleFunc("/ws", s.handleWS).Methods("GET")

	api := s.router.PathPrefix("/api/v1").Subrouter()
	api.Use(corsMiddleware)
	api.HandleFunc("/pods", s.handleListPods).Methods("GET", "OPTIONS")
	api.HandleFunc("/logs/{namespace}/{pod}", s.handleGetLogs).Methods("GET", "OPTIONS")
	api.HandleFunc("/download/{namespace}/{pod}", s.handleDownload).Methods("GET")
	api.HandleFunc("/config", s.handleGetConfig).Methods("GET", "OPTIONS")
	api.HandleFunc("/config", s.handlePutConfig).Methods("PUT", "OPTIONS")
}
