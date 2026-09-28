package api

import (
	"net/http"
	"strings"
	"testing"

	"github.com/prasdud/big-brother/internal/auth"
)

const kumaSample = `{
  "version": "1.23.17",
  "monitors": [
    {"id": 1, "name": "Payments", "type": "group"},
    {"id": 2, "name": "Payments API", "type": "http", "url": "https://payments.example/health", "interval": 30, "timeout": 8, "maxretries": 4, "parent": 1},
    {"id": 3, "name": "gRPC", "type": "grpc", "hostname": "grpc.example", "port": 443}
  ],
  "notifications": []
}`

func TestKumaImportEndpoints(t *testing.T) {
	e := newAuthEnv(t, auth.Config{AllowedDomains: []string{"example.com"}})
	adminTok, adminCSRF := e.userSession(t, "admin@example.com", auth.RoleAdmin)
	viewerTok, viewerCSRF := e.userSession(t, "viewer@example.com", auth.RoleViewer)

	rec := e.do(t, http.MethodPost, "/api/v1/import/kuma/preview", kumaSample, viewerTok, viewerCSRF)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("viewer preview = %d, want 403", rec.Code)
	}

	rec = e.do(t, http.MethodPost, "/api/v1/import/kuma/preview", kumaSample, adminTok, adminCSRF)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "\"create\"") {
		t.Fatalf("preview = %d, body = %s", rec.Code, rec.Body.String())
	}
	// Preview must not write.
	rec = e.do(t, http.MethodGet, "/api/v1/projects", "", adminTok, "")
	if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), "payments") {
		t.Fatalf("preview wrote projects: %s", rec.Body.String())
	}

	rec = e.do(t, http.MethodPost, "/api/v1/import/kuma/apply", kumaSample, adminTok, adminCSRF)
	if rec.Code != http.StatusOK {
		t.Fatalf("apply = %d, body = %s", rec.Code, rec.Body.String())
	}
	rec = e.do(t, http.MethodGet, "/api/v1/projects", "", adminTok, "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Payments") {
		t.Fatalf("apply did not create project: %s", rec.Body.String())
	}

	rec = e.do(t, http.MethodPost, "/api/v1/import/kuma/preview", "{not json", adminTok, adminCSRF)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid body = %d, want 400", rec.Code)
	}
}

func TestOpenAPIDocumentServed(t *testing.T) {
	h := newTestHandler(t)
	rec := do(t, h, http.MethodGet, "/api/v1/openapi.json", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("openapi status = %d", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"\"openapi\":", "/api/v1/projects", "/api/v1/import/kuma/apply"} {
		if !strings.Contains(body, want) {
			t.Fatalf("openapi document missing %q", want)
		}
	}
}
