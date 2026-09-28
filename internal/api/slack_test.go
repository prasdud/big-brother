package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prasdud/big-brother/internal/auth"
	"github.com/prasdud/big-brother/internal/slack"
)

func (e *authEnv) connectSlack(t *testing.T, adminToken string) {
	t.Helper()
	login := e.do(t, http.MethodGet, "/api/v1/slack/install", "", adminToken, "")
	if login.Code != http.StatusFound {
		t.Fatalf("slack install = %d, body = %s", login.Code, login.Body.String())
	}
	stateCookie := cookieByName(login.Result(), slackStateCookie)
	if stateCookie == nil {
		t.Fatal("slack install did not set a state cookie")
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/slack/callback?code=abc&state="+stateCookie.Value, nil)
	req.AddCookie(&http.Cookie{Name: auth.CookieSession, Value: adminToken})
	req.AddCookie(stateCookie)
	rec := httptest.NewRecorder()
	e.h.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("slack callback = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestSlackConnectAndTokenEncrypted(t *testing.T) {
	e := newAuthEnv(t, auth.Config{AllowedDomains: []string{"example.com"}})
	e.installer.install = slack.Install{TeamID: "T1", TeamName: "Acme", BotToken: "xoxb-plain-secret"}
	adminTok, _ := e.userSession(t, "admin@example.com", auth.RoleAdmin)

	e.connectSlack(t, adminTok)

	rec := e.do(t, http.MethodGet, "/api/v1/slack", "", adminTok, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Acme") || strings.Contains(rec.Body.String(), "xoxb-plain-secret") {
		t.Fatalf("status body leaked or missing team: %s", rec.Body.String())
	}

	ws, err := e.q.GetSlackWorkspace(context.Background(), e.wsID)
	if err != nil {
		t.Fatal(err)
	}
	if string(ws.BotTokenEnc) == "xoxb-plain-secret" || strings.Contains(string(ws.BotTokenEnc), "xoxb-plain-secret") {
		t.Fatal("bot token stored in plaintext")
	}
}

func TestSlackChannelsAndChannelsRBAC(t *testing.T) {
	e := newAuthEnv(t, auth.Config{AllowedDomains: []string{"example.com"}})
	e.installer.install = slack.Install{TeamID: "T1", TeamName: "Acme", BotToken: "xoxb-token"}
	e.slackClient.channels = []slack.Channel{{ID: "C1", Name: "alerts"}}
	adminTok, adminCSRF := e.userSession(t, "admin@example.com", auth.RoleAdmin)
	e.connectSlack(t, adminTok)

	rec := e.do(t, http.MethodGet, "/api/v1/slack/channels", "", adminTok, "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "alerts") {
		t.Fatalf("channels = %d, body = %s", rec.Code, rec.Body.String())
	}

	e.do(t, http.MethodPost, "/api/v1/projects", `{"name":"Pay"}`, adminTok, adminCSRF)
	rec = e.do(t, http.MethodPut, "/api/v1/projects/pay/channel",
		`{"slack_channel_id":"C1","name":"alerts"}`, adminTok, adminCSRF)
	if rec.Code != http.StatusOK {
		t.Fatalf("set channel = %d, body = %s", rec.Code, rec.Body.String())
	}
	var cv channelView
	if err := json.Unmarshal(rec.Body.Bytes(), &cv); err != nil {
		t.Fatal(err)
	}
	if cv.SlackChannelID != "C1" {
		t.Fatalf("channel = %+v", cv)
	}

	viewerTok, viewerCSRF := e.userSession(t, "viewer@example.com", auth.RoleViewer)
	rec = e.do(t, http.MethodPut, "/api/v1/projects/pay/channel",
		`{"slack_channel_id":"C1"}`, viewerTok, viewerCSRF)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("viewer set channel = %d, want 403", rec.Code)
	}
}

func TestAlertTemplatesAndTestSend(t *testing.T) {
	e := newAuthEnv(t, auth.Config{AllowedDomains: []string{"example.com"}})
	e.installer.install = slack.Install{TeamID: "T1", TeamName: "Acme", BotToken: "xoxb-token"}
	adminTok, adminCSRF := e.userSession(t, "admin@example.com", auth.RoleAdmin)
	e.connectSlack(t, adminTok)

	e.do(t, http.MethodPost, "/api/v1/projects", `{"name":"Pay"}`, adminTok, adminCSRF)
	e.do(t, http.MethodPut, "/api/v1/projects/pay/channel",
		`{"slack_channel_id":"C1","name":"alerts"}`, adminTok, adminCSRF)

	rec := e.do(t, http.MethodGet, "/api/v1/projects/pay/alert-templates", "", adminTok, "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "{{service.name}}") {
		t.Fatalf("templates = %d, body = %s", rec.Code, rec.Body.String())
	}

	rec = e.do(t, http.MethodPut, "/api/v1/projects/pay/alert-templates/down",
		`{"body":"CUSTOM {{service.name}} is {{status}}"}`, adminTok, adminCSRF)
	if rec.Code != http.StatusOK {
		t.Fatalf("put template = %d", rec.Code)
	}

	rec = e.do(t, http.MethodPost, "/api/v1/projects/pay/alert-templates/preview",
		`{"trigger":"down","body":"{{service.name}} is {{status}}"}`, adminTok, adminCSRF)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Example Service is down") {
		t.Fatalf("preview = %d, body = %s", rec.Code, rec.Body.String())
	}

	rec = e.do(t, http.MethodPost, "/api/v1/projects/pay/alert-templates/test-send",
		`{"trigger":"down"}`, adminTok, adminCSRF)
	if rec.Code != http.StatusOK {
		t.Fatalf("test send = %d, body = %s", rec.Code, rec.Body.String())
	}
	posts := e.slackClient.posts()
	if len(posts) != 1 || !strings.Contains(posts[0].text, "Example Service") {
		t.Fatalf("posts = %+v", posts)
	}
}

func TestDeliveryFailureRecorded(t *testing.T) {
	e := newAuthEnv(t, auth.Config{AllowedDomains: []string{"example.com"}})
	e.installer.install = slack.Install{TeamID: "T1", TeamName: "Acme", BotToken: "xoxb-token"}
	e.slackClient.postErr = errors.New("channel_not_found")
	adminTok, adminCSRF := e.userSession(t, "admin@example.com", auth.RoleAdmin)
	e.connectSlack(t, adminTok)

	e.do(t, http.MethodPost, "/api/v1/projects", `{"name":"Pay"}`, adminTok, adminCSRF)
	e.do(t, http.MethodPut, "/api/v1/projects/pay/channel",
		`{"slack_channel_id":"C1","name":"alerts"}`, adminTok, adminCSRF)

	rec := e.do(t, http.MethodPost, "/api/v1/projects/pay/alert-templates/test-send",
		`{"trigger":"down"}`, adminTok, adminCSRF)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("test send = %d, want 502", rec.Code)
	}

	rec = e.do(t, http.MethodGet, "/api/v1/projects/pay/deliveries", "", adminTok, "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "channel_not_found") {
		t.Fatalf("deliveries = %d, body = %s", rec.Code, rec.Body.String())
	}
}
