package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/prasdud/big-brother/internal/alert"
	"github.com/prasdud/big-brother/internal/store"
)

type alertTemplateView struct {
	Trigger string `json:"trigger"`
	Body    string `json:"body"`
	Custom  bool   `json:"custom"`
}

type deliveryView struct {
	ID        string `json:"id"`
	ServiceID string `json:"service_id"`
	Trigger   string `json:"trigger"`
	Reason    string `json:"reason"`
	CreatedAt string `json:"created_at"`
}

var alertTriggers = []string{alert.TriggerDown, alert.TriggerRecovered}

func validTrigger(trigger string) bool {
	for _, t := range alertTriggers {
		if t == trigger {
			return true
		}
	}
	return false
}

func (s *Server) listAlertTemplates(w http.ResponseWriter, r *http.Request) {
	p, err := s.projectFromPath(r)
	if err != nil {
		notFound(w)
		return
	}
	views := make([]alertTemplateView, 0, len(alertTriggers))
	for _, trigger := range alertTriggers {
		view := alertTemplateView{Trigger: trigger, Body: alert.Builtin(trigger)}
		if t, err := s.q.GetAlertTemplate(r.Context(), store.GetAlertTemplateParams{
			ProjectID: p.ID,
			Trigger:   trigger,
		}); err == nil {
			view.Body = t.Body
			view.Custom = true
		} else if !errors.Is(err, sql.ErrNoRows) {
			serverError(w, err)
			return
		}
		views = append(views, view)
	}
	writeJSON(w, http.StatusOK, views)
}

type alertTemplateRequest struct {
	Body string `json:"body"`
}

func (s *Server) putAlertTemplate(w http.ResponseWriter, r *http.Request) {
	p, err := s.projectFromPath(r)
	if err != nil {
		notFound(w)
		return
	}
	trigger := chi.URLParam(r, "trigger")
	if !validTrigger(trigger) {
		badRequest(w, "trigger must be one of down, recovered")
		return
	}
	var req alertTemplateRequest
	if !decode(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.Body) == "" {
		badRequest(w, "body is required")
		return
	}
	now := store.NowUTC()
	if err := s.q.UpsertAlertTemplate(r.Context(), store.UpsertAlertTemplateParams{
		ID:        uuid.NewString(),
		ProjectID: p.ID,
		Trigger:   trigger,
		Body:      req.Body,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, alertTemplateView{Trigger: trigger, Body: req.Body, Custom: true})
}

type previewRequest struct {
	Trigger string `json:"trigger"`
	Body    string `json:"body"`
}

type renderedView struct {
	Rendered string `json:"rendered"`
}

func (s *Server) previewAlertTemplate(w http.ResponseWriter, r *http.Request) {
	p, err := s.projectFromPath(r)
	if err != nil {
		notFound(w)
		return
	}
	var req previewRequest
	if !decode(w, r, &req) {
		return
	}
	body := req.Body
	if body == "" {
		body = alert.Builtin(req.Trigger)
	}
	if body == "" {
		badRequest(w, "body is required")
		return
	}
	writeJSON(w, http.StatusOK, renderedView{Rendered: alert.Render(body, sampleVars(p, req.Trigger))})
}

type testSendRequest struct {
	Trigger   string `json:"trigger"`
	Body      string `json:"body"`
	ChannelID string `json:"channel_id"`
}

func (s *Server) testSendAlert(w http.ResponseWriter, r *http.Request) {
	p, err := s.projectFromPath(r)
	if err != nil {
		notFound(w)
		return
	}
	var req testSendRequest
	if !decode(w, r, &req) {
		return
	}
	trigger := req.Trigger
	if trigger == "" {
		trigger = alert.TriggerDown
	}
	if !validTrigger(trigger) {
		badRequest(w, "trigger must be one of down, recovered")
		return
	}

	channelID := req.ChannelID
	if channelID == "" {
		channelID = p.DefaultChannelID
	}
	if channelID == "" {
		badRequest(w, "no channel selected")
		return
	}
	channel, err := s.q.GetChannelByID(r.Context(), channelID)
	if errors.Is(err, sql.ErrNoRows) {
		badRequest(w, "channel not found")
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}

	body := req.Body
	if body == "" {
		body = alert.Builtin(trigger)
		if t, err := s.q.GetAlertTemplate(r.Context(), store.GetAlertTemplateParams{
			ProjectID: p.ID,
			Trigger:   trigger,
		}); err == nil {
			body = t.Body
		}
	}
	rendered := alert.Render(body, sampleVars(p, trigger))

	if err := s.slack.Alerter.TestSend(r.Context(), channel.SlackChannelID, rendered); err != nil {
		_ = s.q.CreateDeliveryFailure(r.Context(), store.CreateDeliveryFailureParams{
			ID:        uuid.NewString(),
			ProjectID: p.ID,
			Trigger:   trigger,
			Reason:    err.Error(),
			CreatedAt: store.NowUTC(),
		})
		writeError(w, http.StatusBadGateway, "delivery_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"sent": true})
}

func (s *Server) listDeliveries(w http.ResponseWriter, r *http.Request) {
	p, err := s.projectFromPath(r)
	if err != nil {
		notFound(w)
		return
	}
	rows, err := s.q.ListDeliveryFailuresByProject(r.Context(), store.ListDeliveryFailuresByProjectParams{
		ProjectID: p.ID,
		MaxRows:   50,
	})
	if err != nil {
		serverError(w, err)
		return
	}
	out := make([]deliveryView, 0, len(rows))
	for _, row := range rows {
		out = append(out, deliveryView{
			ID:        row.ID,
			ServiceID: row.ServiceID,
			Trigger:   row.Trigger,
			Reason:    row.Reason,
			CreatedAt: row.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func sampleVars(p store.Project, trigger string) alert.Vars {
	status := "down"
	if trigger == alert.TriggerRecovered {
		status = "up"
	}
	return alert.Vars{
		ServiceName: "Example Service",
		ServiceURL:  "https://example.com/health",
		ProjectName: p.Name,
		Status:      status,
		Duration:    "10m0s",
		Error:       "connection refused",
	}
}
