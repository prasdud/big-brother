package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/prasdud/big-brother/internal/auth"
	"github.com/prasdud/big-brother/internal/db"
	"github.com/prasdud/big-brother/internal/metrics"
	"github.com/prasdud/big-brother/internal/store"
	"github.com/prasdud/big-brother/internal/web"
)

type fakeVerifier struct {
	identity auth.Identity
	err      error
	nonce    string
}

func (f *fakeVerifier) AuthCodeURL(state, nonce string) string {
	f.nonce = nonce
	return "https://fake.example/auth?state=" + state
}

func (f *fakeVerifier) Exchange(_ context.Context, _, nonce string) (auth.Identity, error) {
	if f.err != nil {
		return auth.Identity{}, f.err
	}
	if nonce != f.nonce {
		return auth.Identity{}, errors.New("nonce mismatch")
	}
	return f.identity, nil
}

type authEnv struct {
	h        http.Handler
	q        *store.Queries
	svc      *auth.Service
	verifier *fakeVerifier
	wsID     string
}

func newAuthEnv(t *testing.T, cfg auth.Config) *authEnv {
	t.Helper()
	sqldb, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { sqldb.Close() })
	if err := db.Migrate(sqldb); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	q := store.New(sqldb)
	ws, err := store.EnsureWorkspace(context.Background(), q, "Test")
	if err != nil {
		t.Fatalf("workspace: %v", err)
	}
	cfg.WorkspaceID = ws.ID
	if cfg.SessionTTL == 0 {
		cfg.SessionTTL = time.Hour
	}
	verifier := &fakeVerifier{}
	svc := auth.New(q, verifier, cfg)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := New(q, sqldb, ws, metrics.New(), web.Handler(), svc, logger).Router()
	return &authEnv{h: h, q: q, svc: svc, verifier: verifier, wsID: ws.ID}
}

func (e *authEnv) userSession(t *testing.T, email, role string) (token, csrf string) {
	t.Helper()
	u := store.User{
		ID: uuid.NewString(), WorkspaceID: e.wsID, Email: email,
		Role: role, CreatedAt: store.NowUTC(),
	}
	if err := e.q.CreateUser(context.Background(), store.CreateUserParams{
		ID: u.ID, WorkspaceID: u.WorkspaceID, Email: u.Email, Name: "Test",
		Role: u.Role, Disabled: 0, CreatedAt: u.CreatedAt,
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}
	token, csrf, err := e.svc.CreateSession(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	return token, csrf
}

func (e *authEnv) do(t *testing.T, method, path, body, token, csrf string) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, r)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.AddCookie(&http.Cookie{Name: auth.CookieSession, Value: token})
	}
	if csrf != "" {
		req.Header.Set(auth.HeaderCSRF, csrf)
	}
	rec := httptest.NewRecorder()
	e.h.ServeHTTP(rec, req)
	return rec
}

func cookieByName(res *http.Response, name string) *http.Cookie {
	for _, c := range res.Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func TestOIDCLoginFlow(t *testing.T) {
	e := newAuthEnv(t, auth.Config{AllowedDomains: []string{"example.com"}})
	e.verifier.identity = auth.Identity{Email: "alice@example.com", Name: "Alice", Subject: "s1"}

	login := e.do(t, http.MethodGet, "/auth/login", "", "", "")
	if login.Code != http.StatusFound {
		t.Fatalf("login status = %d, want 302", login.Code)
	}
	stateCookie := cookieByName(login.Result(), auth.CookieState)
	if stateCookie == nil {
		t.Fatal("login did not set state cookie")
	}
	state := strings.SplitN(stateCookie.Value, ".", 2)[0]

	req := httptest.NewRequest(http.MethodGet, "/auth/callback?code=abc&state="+state, nil)
	req.AddCookie(stateCookie)
	cb := httptest.NewRecorder()
	e.h.ServeHTTP(cb, req)
	if cb.Code != http.StatusFound {
		t.Fatalf("callback status = %d, body = %s", cb.Code, cb.Body.String())
	}
	sessionCookie := cookieByName(cb.Result(), auth.CookieSession)
	if sessionCookie == nil || sessionCookie.Value == "" {
		t.Fatal("callback did not set session cookie")
	}

	meReq := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	meReq.AddCookie(sessionCookie)
	me := httptest.NewRecorder()
	e.h.ServeHTTP(me, meReq)
	if me.Code != http.StatusOK {
		t.Fatalf("me status = %d", me.Code)
	}
	var v userView
	if err := json.Unmarshal(me.Body.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	if v.Email != "alice@example.com" || v.Role != auth.RoleViewer {
		t.Fatalf("me = %+v", v)
	}
}

func TestOIDCDomainRejected(t *testing.T) {
	e := newAuthEnv(t, auth.Config{AllowedDomains: []string{"example.com"}})
	e.verifier.identity = auth.Identity{Email: "bob@other.com", Subject: "s2"}

	login := e.do(t, http.MethodGet, "/auth/login", "", "", "")
	stateCookie := cookieByName(login.Result(), auth.CookieState)
	state := strings.SplitN(stateCookie.Value, ".", 2)[0]

	req := httptest.NewRequest(http.MethodGet, "/auth/callback?code=abc&state="+state, nil)
	req.AddCookie(stateCookie)
	cb := httptest.NewRecorder()
	e.h.ServeHTTP(cb, req)
	if cb.Code != http.StatusForbidden {
		t.Fatalf("callback status = %d, want 403", cb.Code)
	}
	if cookieByName(cb.Result(), auth.CookieSession) != nil {
		t.Fatal("session cookie must not be set for rejected domain")
	}
}

func TestRoleEnforcement(t *testing.T) {
	e := newAuthEnv(t, auth.Config{AllowedDomains: []string{"example.com"}})
	adminTok, adminCSRF := e.userSession(t, "admin@example.com", auth.RoleAdmin)
	viewerTok, viewerCSRF := e.userSession(t, "viewer@example.com", auth.RoleViewer)
	memberTok, memberCSRF := e.userSession(t, "member@example.com", auth.RoleMember)

	if rec := e.do(t, http.MethodPost, "/api/v1/projects", `{"name":"P"}`, "", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d, want 401", rec.Code)
	}
	if rec := e.do(t, http.MethodGet, "/api/v1/projects", "", viewerTok, ""); rec.Code != http.StatusOK {
		t.Fatalf("viewer read status = %d, want 200", rec.Code)
	}
	if rec := e.do(t, http.MethodPost, "/api/v1/projects", `{"name":"P"}`, adminTok, adminCSRF); rec.Code != http.StatusCreated {
		t.Fatalf("admin create project = %d", rec.Code)
	}

	createSvc := `{"name":"s","type":"http","url":"http://example.test"}`
	if rec := e.do(t, http.MethodPost, "/api/v1/projects/p/services", createSvc, viewerTok, viewerCSRF); rec.Code != http.StatusForbidden {
		t.Fatalf("viewer create service = %d, want 403", rec.Code)
	}
	if rec := e.do(t, http.MethodPost, "/api/v1/projects/p/services", createSvc, memberTok, memberCSRF); rec.Code != http.StatusCreated {
		t.Fatalf("member create service = %d, want 201 (%s)", rec.Code, rec.Body.String())
	}
	if rec := e.do(t, http.MethodPost, "/api/v1/projects", `{"name":"Q"}`, memberTok, memberCSRF); rec.Code != http.StatusForbidden {
		t.Fatalf("member create project = %d, want 403", rec.Code)
	}
}

func TestCSRFProtection(t *testing.T) {
	e := newAuthEnv(t, auth.Config{AllowedDomains: []string{"example.com"}})
	adminTok, adminCSRF := e.userSession(t, "admin@example.com", auth.RoleAdmin)

	if rec := e.do(t, http.MethodPost, "/api/v1/projects", `{"name":"P"}`, adminTok, ""); rec.Code != http.StatusForbidden {
		t.Fatalf("missing CSRF status = %d, want 403", rec.Code)
	}
	if rec := e.do(t, http.MethodPost, "/api/v1/projects", `{"name":"P"}`, adminTok, adminCSRF); rec.Code != http.StatusCreated {
		t.Fatalf("valid CSRF status = %d, want 201", rec.Code)
	}
}

func TestLogoutInvalidatesSession(t *testing.T) {
	e := newAuthEnv(t, auth.Config{AllowedDomains: []string{"example.com"}})
	tok, csrf := e.userSession(t, "admin@example.com", auth.RoleAdmin)

	if rec := e.do(t, http.MethodPost, "/auth/logout", "", tok, csrf); rec.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d", rec.Code)
	}
	if rec := e.do(t, http.MethodGet, "/auth/me", "", tok, ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("me after logout = %d, want 401", rec.Code)
	}
}

func TestUsersCRUD(t *testing.T) {
	e := newAuthEnv(t, auth.Config{AllowedDomains: []string{"example.com"}})
	adminTok, adminCSRF := e.userSession(t, "admin@example.com", auth.RoleAdmin)
	viewerTok, _ := e.userSession(t, "viewer@example.com", auth.RoleViewer)

	if rec := e.do(t, http.MethodGet, "/api/v1/users", "", viewerTok, ""); rec.Code != http.StatusForbidden {
		t.Fatalf("viewer list users = %d, want 403", rec.Code)
	}

	rec := e.do(t, http.MethodPost, "/api/v1/users",
		`{"email":"ops@example.com","name":"Ops","role":"member"}`, adminTok, adminCSRF)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create user = %d, body = %s", rec.Code, rec.Body.String())
	}
	var created userView
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Role != auth.RoleMember || created.Email != "ops@example.com" {
		t.Fatalf("created = %+v", created)
	}

	rec = e.do(t, http.MethodGet, "/api/v1/users", "", adminTok, "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "ops@example.com") {
		t.Fatalf("list users = %d, body = %s", rec.Code, rec.Body.String())
	}

	rec = e.do(t, http.MethodPatch, "/api/v1/users/"+created.ID, `{"role":"admin"}`, adminTok, adminCSRF)
	if rec.Code != http.StatusOK {
		t.Fatalf("update role = %d, body = %s", rec.Code, rec.Body.String())
	}
	if rec := e.do(t, http.MethodDelete, "/api/v1/users/"+created.ID, "", adminTok, adminCSRF); rec.Code != http.StatusNoContent {
		t.Fatalf("delete user = %d", rec.Code)
	}
}
