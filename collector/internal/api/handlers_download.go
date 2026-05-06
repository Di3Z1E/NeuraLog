package api

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ns, pod := vars["namespace"], vars["pod"]

	path := s.store.FilePath(ns, pod)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "no logs found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+ns+"-"+pod+`.log"`)
	http.ServeContent(w, r, pod+".log", stat.ModTime(), f)
}
