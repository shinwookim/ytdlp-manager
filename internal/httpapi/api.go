package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"database/sql"
)

type createJobRequest struct {
	URL  string `json:"url"`
	Mode string `json:"mode"`
}

func (s *Server) listJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := s.jobs.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list jobs")
		return
	}

	writeJSON(w, http.StatusOK, jobs)
}

func (s *Server) createJob(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	var req createJobRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	job, err := s.jobs.Create(r.Context(), req.URL, req.Mode)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, job)
}

func (s *Server) getJob(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "Invalid job ID")
		return
	}

	job, err := s.jobs.Get(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "Job not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to retrieve job")
		return
	}

	writeJSON(w, http.StatusOK, job)
}

func (s *Server) cancelJob(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "Invalid job ID")
		return
	}

	if err := s.jobs.Cancel(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to cancel job")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) clearJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := s.jobs.ClearFinishedJobs(r.Context()); err != nil {
		http.Error(w, "failed to clear jobs", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
