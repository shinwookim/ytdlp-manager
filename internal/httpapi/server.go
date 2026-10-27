package httpapi

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/shinwookim/ytdlp-manager/internal/jobs"
)

type Server struct {
	jobs     *jobs.Service
	mux      *http.ServeMux
	template *template.Template
}

func NewServer(service *jobs.Service) (*Server, error) {
	tmpl, err := template.ParseFiles(filepath.Join("web", "templates", "index.html"))
	if err != nil {
		return nil, err
	}

	s := &Server{
		jobs:     service,
		mux:      http.NewServeMux(),
		template: tmpl,
	}

	s.mux.HandleFunc("GET /{$}", s.index)
	s.mux.Handle("GET /static/", http.StripPrefix(
		"/static/",
		http.FileServer(http.Dir(filepath.Join("web", "static"))),
	))

	s.mux.HandleFunc("GET /api/jobs", s.listJobs)
	s.mux.HandleFunc("POST /api/jobs", s.createJob)
	s.mux.HandleFunc("GET /api/jobs/{id}", s.getJob)
	s.mux.HandleFunc("DELETE /api/jobs/{id}", s.cancelJob)
	s.mux.HandleFunc("DELETE /api/jobs", s.clearJobs)

	return s, nil
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := s.template.Execute(w, nil); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
