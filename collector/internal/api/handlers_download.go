package api

import (
	"bufio"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ns, pod := vars["namespace"], vars["pod"]

	from := parseTimeParam(r, "from")
	to := parseTimeParam(r, "to")

	filename := ns + "-" + pod + ".log"
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)

	if from == nil && to == nil {
		// No date range — stream the raw file unchanged.
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
		http.ServeContent(w, r, pod+".log", stat.ModTime(), f)
		return
	}

	// Date range active — filter and stream matching lines.
	lines, err := s.store.ReadLines(ns, pod, 0, "", "", from, to)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "no logs found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	bw := bufio.NewWriter(w)
	for _, line := range lines {
		bw.WriteString(line)
		bw.WriteByte('\n')
	}
	bw.Flush()
}
