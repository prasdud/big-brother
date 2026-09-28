package api

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/prasdud/big-brother/internal/auth"
	"github.com/prasdud/big-brother/internal/store"
)

const slackStateCookie = "bb_slack_state"

func randomToken() (string, error) { return auth.RandomToken() }

func (s *Server) cookieSecure() bool {
	if s.auth != nil {
		return s.auth.Secure()
	}
	return false
}

type slackStatusView struct {
	Connected bool   `json:"connected"`
	TeamName  string `json:"team_name"`
}

type channelOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type channelView struct {
	ChannelID      string `json:"channel_id"`
	SlackChannelID string `json:"slack_channel_id"`
	Name           string `json:"name"`
}

type channelRequest struct {
	SlackChannelID string `json:"slack_channel_id"`
	Name           string `json:"name"`
}

func (s *Server) slackInstall(w http.ResponseWriter, r *http.Request) {
	state, err := randomToken()
	if err != nil {
		serverError(w, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     slackStateCookie,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cookieSecure(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   oauthStateTTL,
	})
	http.Redirect(w, r, s.slack.Installer.AuthorizeURL(state), http.StatusFound)
}

func (s *Server) slackCallback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie(slackStateCookie)
	if err != nil {
		badRequest(w, "missing Slack state cookie")
		return
	}
	if r.URL.Query().Get("state") != stateCookie.Value {
		badRequest(w, "Slack state mismatch")
		return
	}
	if e := r.URL.Query().Get("error"); e != "" {
		writeError(w, http.StatusUnauthorized, "slack_error", e)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		badRequest(w, "missing authorization code")
		return
	}
	clearCookie(w, slackStateCookie, s.cookieSecure())

	install, err := s.slack.Installer.Exchange(r.Context(), code)
	if err != nil {
		s.logger.Warn("slack oauth exchange failed", "error", err)
		writeError(w, http.StatusBadGateway, "slack_error", "could not complete Slack install")
		return
	}
	encrypted, err := s.slack.Box.Encrypt([]byte(install.BotToken))
	if err != nil {
		serverError(w, err)
		return
	}
	if err := s.q.UpsertSlackWorkspace(r.Context(), store.UpsertSlackWorkspaceParams{
		WorkspaceID: s.workspace.ID,
		TeamID:      install.TeamID,
		TeamName:    install.TeamName,
		BotTokenEnc: encrypted,
		InstalledAt: store.NowUTC(),
	}); err != nil {
		serverError(w, err)
		return
	}
	s.logger.Info("slack workspace connected", "team", install.TeamName)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) slackStatus(w http.ResponseWriter, r *http.Request) {
	team, connected, err := s.slack.Alerter.Status(r.Context())
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, slackStatusView{Connected: connected, TeamName: team})
}

func (s *Server) slackChannels(w http.ResponseWriter, r *http.Request) {
	channels, err := s.slack.Alerter.ListChannels(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, "slack_error", err.Error())
		return
	}
	out := make([]channelOption, 0, len(channels))
	for _, c := range channels {
		out = append(out, channelOption{ID: c.ID, Name: c.Name})
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) listProjectChannels(w http.ResponseWriter, r *http.Request) {
	p, err := s.projectFromPath(r)
	if err != nil {
		notFound(w)
		return
	}
	rows, err := s.q.ListChannelsByProject(r.Context(), p.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	out := make([]channelView, 0, len(rows))
	for _, row := range rows {
		out = append(out, channelView{ChannelID: row.ID, SlackChannelID: row.SlackChannelID, Name: row.Name})
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) setProjectChannel(w http.ResponseWriter, r *http.Request) {
	p, err := s.projectFromPath(r)
	if err != nil {
		notFound(w)
		return
	}
	var req channelRequest
	if !decode(w, r, &req) {
		return
	}
	if req.SlackChannelID == "" {
		badRequest(w, "slack_channel_id is required")
		return
	}
	channel, err := s.q.UpsertChannel(r.Context(), store.UpsertChannelParams{
		ID:             uuid.NewString(),
		ProjectID:      p.ID,
		Name:           req.Name,
		SlackChannelID: req.SlackChannelID,
		CreatedAt:      store.NowUTC(),
	})
	if err != nil {
		serverError(w, err)
		return
	}
	if err := s.q.SetProjectDefaultChannel(r.Context(), store.SetProjectDefaultChannelParams{
		DefaultChannelID: channel.ID,
		ID:               p.ID,
	}); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, channelView{ChannelID: channel.ID, SlackChannelID: channel.SlackChannelID, Name: channel.Name})
}

func (s *Server) clearProjectChannel(w http.ResponseWriter, r *http.Request) {
	p, err := s.projectFromPath(r)
	if err != nil {
		notFound(w)
		return
	}
	if err := s.q.SetProjectDefaultChannel(r.Context(), store.SetProjectDefaultChannelParams{
		DefaultChannelID: "",
		ID:               p.ID,
	}); err != nil {
		serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) setServiceChannel(w http.ResponseWriter, r *http.Request) {
	_, svc, err := s.serviceFromPath(r)
	if err != nil {
		notFound(w)
		return
	}
	var req channelRequest
	if !decode(w, r, &req) {
		return
	}
	if req.SlackChannelID == "" {
		badRequest(w, "slack_channel_id is required")
		return
	}
	channel, err := s.q.UpsertChannel(r.Context(), store.UpsertChannelParams{
		ID:             uuid.NewString(),
		ProjectID:      svc.ProjectID,
		Name:           req.Name,
		SlackChannelID: req.SlackChannelID,
		CreatedAt:      store.NowUTC(),
	})
	if err != nil {
		serverError(w, err)
		return
	}
	if err := s.q.SetServiceChannel(r.Context(), store.SetServiceChannelParams{
		ChannelID: channel.ID,
		UpdatedAt: store.NowUTC(),
		ID:        svc.ID,
	}); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, channelView{ChannelID: channel.ID, SlackChannelID: channel.SlackChannelID, Name: channel.Name})
}

func (s *Server) clearServiceChannel(w http.ResponseWriter, r *http.Request) {
	_, svc, err := s.serviceFromPath(r)
	if err != nil {
		notFound(w)
		return
	}
	if err := s.q.SetServiceChannel(r.Context(), store.SetServiceChannelParams{
		ChannelID: "",
		UpdatedAt: store.NowUTC(),
		ID:        svc.ID,
	}); err != nil {
		serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
