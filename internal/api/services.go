package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/prasdud/big-brother/internal/slug"
	"github.com/prasdud/big-brother/internal/store"
)

func (s *Server) listServices(w http.ResponseWriter, r *http.Request) {
	p, err := s.projectFromPath(r)
	if err != nil {
		notFound(w)
		return
	}
	services, err := s.q.ListServices(r.Context(), p.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	out := make([]serviceView, 0, len(services))
	for _, svc := range services {
		out = append(out, newServiceView(svc))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) createService(w http.ResponseWriter, r *http.Request) {
	p, err := s.projectFromPath(r)
	if err != nil {
		notFound(w)
		return
	}
	var req serviceRequest
	if !decode(w, r, &req) {
		return
	}
	req.applyDefaults()
	if err := req.validate(); err != nil {
		badRequest(w, err.Error())
		return
	}

	now := store.NowUTC()
	svc := store.Service{
		ID:               uuid.NewString(),
		ProjectID:        p.ID,
		Name:             req.Name,
		Type:             req.Type,
		Url:              req.URL,
		Hostname:         req.Hostname,
		Port:             req.Port,
		IntervalSeconds:  req.IntervalSeconds,
		TimeoutSeconds:   req.TimeoutSeconds,
		FailureThreshold: req.FailureThreshold,
		Enabled:          1,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	svc.Slug, err = s.uniqueServiceSlug(r, p.ID, slug.Make(req.Name))
	if err != nil {
		serverError(w, err)
		return
	}
	if err := s.q.CreateService(r.Context(), store.CreateServiceParams{
		ID:               svc.ID,
		ProjectID:        svc.ProjectID,
		Name:             svc.Name,
		Slug:             svc.Slug,
		Type:             svc.Type,
		Url:              svc.Url,
		Hostname:         svc.Hostname,
		Port:             svc.Port,
		IntervalSeconds:  svc.IntervalSeconds,
		TimeoutSeconds:   svc.TimeoutSeconds,
		FailureThreshold: svc.FailureThreshold,
		Enabled:          svc.Enabled,
		CreatedAt:        svc.CreatedAt,
		UpdatedAt:        svc.UpdatedAt,
	}); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, newServiceView(svc))
}

func (s *Server) getService(w http.ResponseWriter, r *http.Request) {
	_, svc, err := s.serviceFromPath(r)
	if err != nil {
		notFound(w)
		return
	}
	writeJSON(w, http.StatusOK, newServiceView(svc))
}

func (s *Server) updateService(w http.ResponseWriter, r *http.Request) {
	_, svc, err := s.serviceFromPath(r)
	if err != nil {
		notFound(w)
		return
	}
	var req serviceRequest
	if !decode(w, r, &req) {
		return
	}
	req.applyDefaults()
	if err := req.validate(); err != nil {
		badRequest(w, err.Error())
		return
	}
	svc.Name = req.Name
	svc.Url = req.URL
	svc.Hostname = req.Hostname
	svc.Port = req.Port
	svc.IntervalSeconds = req.IntervalSeconds
	svc.TimeoutSeconds = req.TimeoutSeconds
	svc.FailureThreshold = req.FailureThreshold
	svc.UpdatedAt = store.NowUTC()

	if err := s.q.UpdateService(r.Context(), store.UpdateServiceParams{
		Name:             svc.Name,
		Url:              svc.Url,
		Hostname:         svc.Hostname,
		Port:             svc.Port,
		IntervalSeconds:  svc.IntervalSeconds,
		TimeoutSeconds:   svc.TimeoutSeconds,
		FailureThreshold: svc.FailureThreshold,
		UpdatedAt:        svc.UpdatedAt,
		ID:               svc.ID,
	}); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, newServiceView(svc))
}

func (s *Server) deleteService(w http.ResponseWriter, r *http.Request) {
	_, svc, err := s.serviceFromPath(r)
	if err != nil {
		notFound(w)
		return
	}
	if err := s.q.DeleteService(r.Context(), svc.ID); err != nil {
		serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) setServiceEnabled(enabled bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, svc, err := s.serviceFromPath(r)
		if err != nil {
			notFound(w)
			return
		}
		var flag int64
		if enabled {
			flag = 1
		}
		now := store.NowUTC()
		if err := s.q.SetServiceEnabled(r.Context(), store.SetServiceEnabledParams{
			Enabled:   flag,
			UpdatedAt: now,
			ID:        svc.ID,
		}); err != nil {
			serverError(w, err)
			return
		}
		// A paused service is not checked; a resumed service is due again.
		if err := s.q.SetServiceNextRun(r.Context(), store.SetServiceNextRunParams{
			NextRunAt: sql.NullString{},
			ID:        svc.ID,
		}); err != nil {
			serverError(w, err)
			return
		}
		state := "paused"
		if enabled {
			state = "pending"
		}
		if err := s.setState(r.Context(), svc, state); err != nil {
			serverError(w, err)
			return
		}
		svc.Enabled = flag
		svc.UpdatedAt = now
		writeJSON(w, http.StatusOK, newServiceView(svc))
	}
}

// setState writes a service state row, preserving the last check time and only
// advancing the last-change time when the state actually changes.
func (s *Server) setState(ctx context.Context, svc store.Service, state string) error {
	params := store.UpsertServiceStateParams{
		ServiceID:    svc.ID,
		ProjectID:    svc.ProjectID,
		State:        state,
		LastChangeAt: store.NowUTC(),
	}
	st, err := s.q.GetServiceState(ctx, svc.ID)
	if err == nil {
		params.LastCheckAt = st.LastCheckAt
		if st.State == state {
			params.LastChangeAt = st.LastChangeAt
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	return s.q.UpsertServiceState(ctx, params)
}

// serviceFromPath loads the project and service named by the path slugs.
func (s *Server) serviceFromPath(r *http.Request) (store.Project, store.Service, error) {
	p, err := s.projectFromPath(r)
	if err != nil {
		return store.Project{}, store.Service{}, err
	}
	svc, err := s.q.GetServiceBySlug(r.Context(), store.GetServiceBySlugParams{
		ProjectID: p.ID,
		Slug:      chi.URLParam(r, "service"),
	})
	if err != nil {
		return store.Project{}, store.Service{}, err
	}
	return p, svc, nil
}

// uniqueServiceSlug appends a numeric suffix until the slug is free within the
// project.
func (s *Server) uniqueServiceSlug(r *http.Request, projectID, base string) (string, error) {
	if base == "" {
		base = "service"
	}
	for n := 0; n < 100; n++ {
		candidate := base
		if n > 0 {
			candidate = base + "-" + strconv.Itoa(n+1)
		}
		_, err := s.q.GetServiceBySlug(r.Context(), store.GetServiceBySlugParams{
			ProjectID: projectID,
			Slug:      candidate,
		})
		if errors.Is(err, sql.ErrNoRows) {
			return candidate, nil
		}
		if err != nil {
			return "", err
		}
	}
	return "", fmt.Errorf("could not allocate a unique slug")
}
