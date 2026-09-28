package api

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prasdud/big-brother/internal/db"
	"github.com/prasdud/big-brother/internal/metrics"
	"github.com/prasdud/big-brother/internal/store"
	"github.com/prasdud/big-brother/internal/web"
)

func newTestHandler(t *testing.T) http.Handler {
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
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(q, sqldb, ws, metrics.New(), web.Handler(), logger).Router()
}

func do(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, r)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHealthAndMetrics(t *testing.T) {
	h := newTestHandler(t)

	if rec := do(t, h, http.MethodGet, "/healthz", ""); rec.Code != http.StatusOK {
		t.Fatalf("healthz status = %d, want 200", rec.Code)
	}
	rec := do(t, h, http.MethodGet, "/metrics", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("metrics status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "bb_http_requests_total") {
		t.Fatalf("metrics body missing counter: %s", rec.Body.String())
	}
}

func TestProjectAndServiceLifecycle(t *testing.T) {
	h := newTestHandler(t)

	rec := do(t, h, http.MethodPost, "/api/v1/projects", `{"name":"Payments API"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create project status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var project projectView
	if err := json.Unmarshal(rec.Body.Bytes(), &project); err != nil {
		t.Fatal(err)
	}
	if project.Slug != "payments-api" {
		t.Fatalf("project slug = %q, want payments-api", project.Slug)
	}

	rec = do(t, h, http.MethodPost, "/api/v1/projects/payments-api/services",
		`{"name":"Checkout","type":"http","url":"https://example.com/health"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create service status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var svc serviceView
	if err := json.Unmarshal(rec.Body.Bytes(), &svc); err != nil {
		t.Fatal(err)
	}
	if svc.Slug != "checkout" || !svc.Enabled {
		t.Fatalf("unexpected service: %+v", svc)
	}
	if svc.IntervalSeconds != 60 || svc.TimeoutSeconds != 10 || svc.FailureThreshold != 3 {
		t.Fatalf("defaults not applied: %+v", svc)
	}

	rec = do(t, h, http.MethodPost, "/api/v1/projects/payments-api/services/checkout/pause", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("pause status = %d", rec.Code)
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &svc); err != nil {
		t.Fatal(err)
	}
	if svc.Enabled {
		t.Fatal("service should be disabled after pause")
	}

	rec = do(t, h, http.MethodGet, "/api/v1/projects/payments-api/services", "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "checkout") {
		t.Fatalf("list services = %d, body = %s", rec.Code, rec.Body.String())
	}

	rec = do(t, h, http.MethodDelete, "/api/v1/projects/payments-api/services/checkout", "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete service status = %d", rec.Code)
	}
	rec = do(t, h, http.MethodGet, "/api/v1/projects/payments-api/services/checkout", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("deleted service status = %d, want 404", rec.Code)
	}
}

func TestServiceValidation(t *testing.T) {
	h := newTestHandler(t)
	do(t, h, http.MethodPost, "/api/v1/projects", `{"name":"P"}`)

	cases := []struct {
		name string
		body string
	}{
		{"http without url", `{"name":"a","type":"http"}`},
		{"tcp without hostname", `{"name":"b","type":"tcp","port":443}`},
		{"tcp bad port", `{"name":"c","type":"tcp","hostname":"h","port":0}`},
		{"unknown type", `{"name":"d","type":"carrier-pigeon"}`},
		{"missing name", `{"type":"http","url":"https://x"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := do(t, h, http.MethodPost, "/api/v1/projects/p/services", tc.body)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400 (body %s)", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestDuplicateProjectSlugGetsSuffix(t *testing.T) {
	h := newTestHandler(t)
	do(t, h, http.MethodPost, "/api/v1/projects", `{"name":"Payments"}`)
	rec := do(t, h, http.MethodPost, "/api/v1/projects", `{"name":"Payments"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d", rec.Code)
	}
	var project projectView
	if err := json.Unmarshal(rec.Body.Bytes(), &project); err != nil {
		t.Fatal(err)
	}
	if project.Slug != "payments-2" {
		t.Fatalf("slug = %q, want payments-2", project.Slug)
	}
}
