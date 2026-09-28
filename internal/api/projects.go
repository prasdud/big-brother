package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/prasdud/big-brother/internal/slug"
	"github.com/prasdud/big-brother/internal/store"
)

func (s *Server) listProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := s.q.ListProjects(r.Context())
	if err != nil {
		serverError(w, err)
		return
	}
	out := make([]projectView, 0, len(projects))
	for _, p := range projects {
		out = append(out, newProjectView(p))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) createProject(w http.ResponseWriter, r *http.Request) {
	var req nameRequest
	if !decode(w, r, &req) {
		return
	}
	if req.Name == "" {
		badRequest(w, "name is required")
		return
	}

	p := store.Project{
		ID:          uuid.NewString(),
		WorkspaceID: s.workspace.ID,
		Name:        req.Name,
		CreatedAt:   store.NowUTC(),
	}
	p.Slug, _ = s.uniqueProjectSlug(r, slug.Make(req.Name))
	if err := s.q.CreateProject(r.Context(), store.CreateProjectParams{
		ID:          p.ID,
		WorkspaceID: p.WorkspaceID,
		Name:        p.Name,
		Slug:        p.Slug,
		CreatedAt:   p.CreatedAt,
	}); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, newProjectView(p))
}

func (s *Server) getProject(w http.ResponseWriter, r *http.Request) {
	p, err := s.projectFromPath(r)
	if err != nil {
		notFound(w)
		return
	}
	writeJSON(w, http.StatusOK, newProjectView(p))
}

func (s *Server) renameProject(w http.ResponseWriter, r *http.Request) {
	p, err := s.projectFromPath(r)
	if err != nil {
		notFound(w)
		return
	}
	var req nameRequest
	if !decode(w, r, &req) {
		return
	}
	if req.Name == "" {
		badRequest(w, "name is required")
		return
	}
	if err := s.q.RenameProject(r.Context(), store.RenameProjectParams{ID: p.ID, Name: req.Name}); err != nil {
		serverError(w, err)
		return
	}
	p.Name = req.Name
	writeJSON(w, http.StatusOK, newProjectView(p))
}

func (s *Server) deleteProject(w http.ResponseWriter, r *http.Request) {
	p, err := s.projectFromPath(r)
	if err != nil {
		notFound(w)
		return
	}
	if err := s.q.DeleteProject(r.Context(), p.ID); err != nil {
		serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// projectFromPath loads the project named by the {project} slug.
func (s *Server) projectFromPath(r *http.Request) (store.Project, error) {
	return s.q.GetProjectBySlug(r.Context(), chi.URLParam(r, "project"))
}

// uniqueProjectSlug appends a numeric suffix until the slug is free.
func (s *Server) uniqueProjectSlug(r *http.Request, base string) (string, error) {
	if base == "" {
		base = "project"
	}
	for n := 0; n < 100; n++ {
		candidate := base
		if n > 0 {
			candidate = base + "-" + strconv.Itoa(n+1)
		}
		_, err := s.q.GetProjectBySlug(r.Context(), candidate)
		if errors.Is(err, sql.ErrNoRows) {
			return candidate, nil
		}
		if err != nil {
			return "", err
		}
	}
	return "", errors.New("could not allocate a unique slug")
}

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		badRequest(w, "invalid JSON body: "+err.Error())
		return false
	}
	return true
}
