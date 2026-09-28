package monitor

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/prasdud/big-brother/internal/check"
	"github.com/prasdud/big-brother/internal/db"
	"github.com/prasdud/big-brother/internal/store"
)

func setup(t *testing.T) (*store.Queries, store.Service) {
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
	ctx := context.Background()

	ws, err := store.EnsureWorkspace(ctx, q, "Test")
	if err != nil {
		t.Fatalf("workspace: %v", err)
	}
	projectID := uuid.NewString()
	if err := q.CreateProject(ctx, store.CreateProjectParams{
		ID: projectID, WorkspaceID: ws.ID, Name: "P", Slug: "p", CreatedAt: store.NowUTC(),
	}); err != nil {
		t.Fatalf("project: %v", err)
	}

	now := store.NowUTC()
	svc := store.Service{
		ID: uuid.NewString(), ProjectID: projectID, Name: "S", Slug: "s",
		Type: "http", Url: "http://example.test", IntervalSeconds: 60,
		TimeoutSeconds: 10, FailureThreshold: 3, Enabled: 1,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := q.CreateService(ctx, store.CreateServiceParams{
		ID: svc.ID, ProjectID: svc.ProjectID, Name: svc.Name, Slug: svc.Slug,
		Type: svc.Type, Url: svc.Url, Hostname: svc.Hostname, Port: svc.Port,
		IntervalSeconds: svc.IntervalSeconds, TimeoutSeconds: svc.TimeoutSeconds,
		FailureThreshold: svc.FailureThreshold, Enabled: svc.Enabled,
		CreatedAt: svc.CreatedAt, UpdatedAt: svc.UpdatedAt,
	}); err != nil {
		t.Fatalf("service: %v", err)
	}
	return q, svc
}

func TestRecordEmitsOneEventPerTransition(t *testing.T) {
	q, svc := setup(t)
	ctx := context.Background()

	var events []Event
	m := New(q, func(e Event) { events = append(events, e) })

	down := check.Result{Up: false, Error: "boom"}
	up := check.Result{Up: true, StatusCode: 200}

	for i := 0; i < 3; i++ { // up->pending->pending->down
		if err := m.Record(ctx, svc, down); err != nil {
			t.Fatalf("record down %d: %v", i, err)
		}
	}
	if err := m.Record(ctx, svc, down); err != nil { // already down, no event
		t.Fatal(err)
	}
	if err := m.Record(ctx, svc, up); err != nil {
		t.Fatal(err)
	}

	if len(events) != 2 {
		t.Fatalf("events = %d, want 2: %+v", len(events), events)
	}
	if events[0].From != Pending || events[0].To != Down {
		t.Fatalf("first event = %s->%s, want pending->down", events[0].From, events[0].To)
	}
	if events[1].From != Down || events[1].To != Up {
		t.Fatalf("second event = %s->%s, want down->up", events[1].From, events[1].To)
	}

	st, err := q.GetServiceState(ctx, svc.ID)
	if err != nil {
		t.Fatal(err)
	}
	if st.State != string(Up) || st.ConsecutiveSuccesses != 1 {
		t.Fatalf("state = %+v", st)
	}

	checks, err := q.ListChecks(ctx, store.ListChecksParams{
		ServiceID: svc.ID, FromAt: "0000", ToAt: "9999", MaxRows: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(checks) != 5 {
		t.Fatalf("checks = %d, want 5", len(checks))
	}

	uptime, err := q.SumUptime(ctx, store.SumUptimeParams{ServiceID: svc.ID, FromHour: "0000"})
	if err != nil {
		t.Fatal(err)
	}
	if uptime.UpChecks != 1 || uptime.TotalChecks != 5 {
		t.Fatalf("uptime = %+v, want 1/5", uptime)
	}
}
