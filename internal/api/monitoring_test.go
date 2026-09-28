package api

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/prasdud/big-brother/internal/check"
	"github.com/prasdud/big-brother/internal/db"
	"github.com/prasdud/big-brother/internal/metrics"
	"github.com/prasdud/big-brother/internal/monitor"
	"github.com/prasdud/big-brother/internal/store"
	"github.com/prasdud/big-brother/internal/web"
)

func TestMonitoringEndpoints(t *testing.T) {
	sqldb, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { sqldb.Close() })
	if err := db.Migrate(sqldb); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	q := store.New(sqldb)
	ws, _ := store.EnsureWorkspace(context.Background(), q, "Test")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := New(q, sqldb, ws, metrics.New(), web.Handler(), nil, logger).Router()

	do(t, h, http.MethodPost, "/api/v1/projects", `{"name":"P"}`)
	rec := do(t, h, http.MethodPost, "/api/v1/projects/p/services",
		`{"name":"S","type":"http","url":"http://example.test"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create service = %d", rec.Code)
	}
	svc, err := q.GetServiceBySlug(context.Background(), store.GetServiceBySlugParams{
		ProjectID: mustProjectID(t, q), Slug: "s",
	})
	if err != nil {
		t.Fatal(err)
	}

	mon := monitor.New(q, nil)
	if err := mon.Record(context.Background(), svc, check.Result{Up: true, StatusCode: 200, LatencyMS: 7}); err != nil {
		t.Fatal(err)
	}

	t.Run("status", func(t *testing.T) {
		rec := do(t, h, http.MethodGet, "/api/v1/projects/p/services/s/status", "")
		var v statusView
		if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
			t.Fatal(err)
		}
		if v.State != "up" {
			t.Fatalf("state = %q, want up (%s)", v.State, rec.Body.String())
		}
	})

	t.Run("checks", func(t *testing.T) {
		rec := do(t, h, http.MethodGet, "/api/v1/projects/p/services/s/checks", "")
		var v []checkView
		if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
			t.Fatal(err)
		}
		if len(v) != 1 || v[0].StatusCode != 200 || v[0].LatencyMS != 7 {
			t.Fatalf("checks = %+v", v)
		}
	})

	t.Run("uptime", func(t *testing.T) {
		rec := do(t, h, http.MethodGet, "/api/v1/projects/p/services/s/uptime?window=24h", "")
		var v uptimeView
		if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
			t.Fatal(err)
		}
		if v.Percent != 100 || v.TotalChecks != 1 {
			t.Fatalf("uptime = %+v", v)
		}
	})

	t.Run("bad window", func(t *testing.T) {
		rec := do(t, h, http.MethodGet, "/api/v1/projects/p/services/s/uptime?window=1y", "")
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("pause reports paused", func(t *testing.T) {
		if rec := do(t, h, http.MethodPost, "/api/v1/projects/p/services/s/pause", ""); rec.Code != http.StatusOK {
			t.Fatalf("pause = %d", rec.Code)
		}
		rec := do(t, h, http.MethodGet, "/api/v1/projects/p/services/s/status", "")
		var v statusView
		if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
			t.Fatal(err)
		}
		if v.State != "paused" {
			t.Fatalf("state = %q, want paused", v.State)
		}
	})
}

func mustProjectID(t *testing.T, q *store.Queries) string {
	t.Helper()
	p, err := q.GetProjectBySlug(context.Background(), "p")
	if err != nil {
		t.Fatal(err)
	}
	return p.ID
}
