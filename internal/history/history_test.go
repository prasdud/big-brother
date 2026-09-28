package history

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/prasdud/big-brother/internal/db"
	"github.com/prasdud/big-brother/internal/store"
)

func TestPruneBeforeKeepsRecentChecksAndRollups(t *testing.T) {
	sqldb, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { sqldb.Close() })
	if err := db.Migrate(sqldb); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	q := store.New(sqldb)
	ctx := context.Background()

	ws, _ := store.EnsureWorkspace(ctx, q, "Test")
	projectID := uuid.NewString()
	if err := q.CreateProject(ctx, store.CreateProjectParams{
		ID: projectID, WorkspaceID: ws.ID, Name: "P", Slug: "p", CreatedAt: store.NowUTC(),
	}); err != nil {
		t.Fatal(err)
	}
	now := store.NowUTC()
	svc := store.Service{
		ID: uuid.NewString(), ProjectID: projectID, Name: "S", Slug: "s", Type: "http",
		Url: "http://x", IntervalSeconds: 60, TimeoutSeconds: 10, FailureThreshold: 3,
		Enabled: 1, CreatedAt: now, UpdatedAt: now,
	}
	if err := q.CreateService(ctx, store.CreateServiceParams{
		ID: svc.ID, ProjectID: svc.ProjectID, Name: svc.Name, Slug: svc.Slug, Type: svc.Type,
		Url: svc.Url, Hostname: svc.Hostname, Port: svc.Port, IntervalSeconds: svc.IntervalSeconds,
		TimeoutSeconds: svc.TimeoutSeconds, FailureThreshold: svc.FailureThreshold,
		Enabled: svc.Enabled, CreatedAt: svc.CreatedAt, UpdatedAt: svc.UpdatedAt,
	}); err != nil {
		t.Fatal(err)
	}

	old := time.Now().UTC().Add(-48 * time.Hour).Format(time.RFC3339)
	recent := time.Now().UTC().Format(time.RFC3339)
	insert := func(at string) {
		if err := q.InsertCheck(ctx, store.InsertCheckParams{
			ID: uuid.NewString(), ServiceID: svc.ID, ProjectID: svc.ProjectID,
			CheckedAt: at, Status: "up", StatusCode: 200,
		}); err != nil {
			t.Fatal(err)
		}
	}
	insert(old)
	insert(old)
	insert(recent)

	if err := q.UpsertRollup(ctx, store.UpsertRollupParams{
		ServiceID: svc.ID, ProjectID: svc.ProjectID,
		Hour:     time.Now().UTC().Add(-48 * time.Hour).Truncate(time.Hour).Format(time.RFC3339),
		UpChecks: 2, TotalChecks: 2,
	}); err != nil {
		t.Fatal(err)
	}

	deleted, err := PruneBefore(ctx, q, time.Now().UTC().Add(-24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 2 {
		t.Fatalf("deleted = %d, want 2", deleted)
	}

	remaining, err := q.ListChecks(ctx, store.ListChecksParams{
		ServiceID: svc.ID, FromAt: "0000", ToAt: "9999", MaxRows: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 1 {
		t.Fatalf("remaining checks = %d, want 1", len(remaining))
	}

	uptime, err := q.SumUptime(ctx, store.SumUptimeParams{ServiceID: svc.ID, FromHour: "0000"})
	if err != nil {
		t.Fatal(err)
	}
	if uptime.TotalChecks != 2 {
		t.Fatalf("rollups pruned: %+v", uptime)
	}
}
