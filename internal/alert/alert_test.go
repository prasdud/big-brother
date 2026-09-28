package alert

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/prasdud/big-brother/internal/db"
	"github.com/prasdud/big-brother/internal/monitor"
	"github.com/prasdud/big-brother/internal/secret"
	"github.com/prasdud/big-brother/internal/slack"
	"github.com/prasdud/big-brother/internal/store"
)

func TestRender(t *testing.T) {
	tmpl := "{{service.name}}|{{service.url}}|{{project.name}}|{{status}}|{{duration}}|{{error}}"
	got := Render(tmpl, Vars{
		ServiceName: "API", ServiceURL: "https://x", ProjectName: "Pay",
		Status: "down", Duration: "3m0s", Error: "timeout",
	})
	want := "API|https://x|Pay|down|3m0s|timeout"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestResolveChannelID(t *testing.T) {
	svc := store.Service{ChannelID: "svc-chan"}
	project := store.Project{DefaultChannelID: "proj-chan"}
	if got := ResolveChannelID(svc, project); got != "svc-chan" {
		t.Fatalf("override = %q", got)
	}
	if got := ResolveChannelID(store.Service{}, project); got != "proj-chan" {
		t.Fatalf("default = %q", got)
	}
	if got := ResolveChannelID(store.Service{}, store.Project{}); got != "" {
		t.Fatalf("no resolution = %q, want empty", got)
	}
}

func TestResolveTemplate(t *testing.T) {
	svc := store.Service{TemplateDown: "svc-down", TemplateRecovered: "svc-up"}
	if got := ResolveTemplate(svc, "proj-down", TriggerDown); got != "svc-down" {
		t.Fatalf("service override = %q", got)
	}
	if got := ResolveTemplate(store.Service{}, "proj-down", TriggerDown); got != "proj-down" {
		t.Fatalf("project = %q", got)
	}
	if got := ResolveTemplate(store.Service{}, "", TriggerDown); got != BuiltinDown {
		t.Fatalf("builtin = %q", got)
	}
	if got := ResolveTemplate(svc, "", TriggerRecovered); got != "svc-up" {
		t.Fatalf("recovered override = %q", got)
	}
}

func TestTriggerFor(t *testing.T) {
	cases := []struct {
		from, to, want string
	}{
		{"up", "down", TriggerDown},
		{"pending", "down", TriggerDown},
		{"up", "pending", ""},
		{"down", "up", TriggerRecovered},
		{"pending", "up", ""},
		{"paused", "pending", ""},
	}
	for _, tc := range cases {
		if got := TriggerFor(tc.from, tc.to); got != tc.want {
			t.Errorf("TriggerFor(%q,%q) = %q, want %q", tc.from, tc.to, got, tc.want)
		}
	}
}

type fakeClient struct {
	mu     sync.Mutex
	posted []string
	fail   bool
}

func (f *fakeClient) PostMessage(_ context.Context, _, text string) error {
	if f.fail {
		return errSlack
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.posted = append(f.posted, text)
	return nil
}

func (f *fakeClient) ListChannels(context.Context) ([]slack.Channel, error) { return nil, nil }

func (f *fakeClient) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.posted)
}

type errString string

func (e errString) Error() string { return string(e) }

const errSlack = errString("channel_not_found")

func setupAlerter(t *testing.T) (*Alerter, *store.Queries, store.Service, store.Project, *fakeClient) {
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
	box, _ := secret.NewBox("test-key")

	projectID := uuid.NewString()
	if err := q.CreateProject(ctx, store.CreateProjectParams{
		ID: projectID, WorkspaceID: ws.ID, Name: "Payments", Slug: "payments", CreatedAt: store.NowUTC(),
	}); err != nil {
		t.Fatal(err)
	}
	now := store.NowUTC()
	svc := store.Service{
		ID: uuid.NewString(), ProjectID: projectID, Name: "API", Slug: "api", Type: "http",
		Url: "https://example.test", IntervalSeconds: 60, TimeoutSeconds: 10,
		FailureThreshold: 3, Enabled: 1, CreatedAt: now, UpdatedAt: now,
	}
	if err := q.CreateService(ctx, store.CreateServiceParams{
		ID: svc.ID, ProjectID: svc.ProjectID, Name: svc.Name, Slug: svc.Slug, Type: svc.Type,
		Url: svc.Url, IntervalSeconds: svc.IntervalSeconds, TimeoutSeconds: svc.TimeoutSeconds,
		FailureThreshold: svc.FailureThreshold, Enabled: svc.Enabled,
		CreatedAt: svc.CreatedAt, UpdatedAt: svc.UpdatedAt,
	}); err != nil {
		t.Fatal(err)
	}

	channel, err := q.UpsertChannel(ctx, store.UpsertChannelParams{
		ID: uuid.NewString(), ProjectID: projectID, Name: "alerts",
		SlackChannelID: "C123", CreatedAt: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := q.SetProjectDefaultChannel(ctx, store.SetProjectDefaultChannelParams{
		DefaultChannelID: channel.ID, ID: projectID,
	}); err != nil {
		t.Fatal(err)
	}

	token, _ := box.Encrypt([]byte("xoxb-token"))
	if err := q.UpsertSlackWorkspace(ctx, store.UpsertSlackWorkspaceParams{
		WorkspaceID: ws.ID, TeamID: "T1", TeamName: "Acme", BotTokenEnc: token, InstalledAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	client := &fakeClient{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := New(q, ws.ID, box, func(string) slack.Client { return client }, logger)
	project, _ := q.GetProjectByID(ctx, projectID)
	return a, q, svc, project, client
}

func TestHandleDeliversDownAndRecovery(t *testing.T) {
	a, q, svc, _, client := setupAlerter(t)
	ctx := context.Background()

	a.Handle(ctx, monitor.Event{ServiceID: svc.ID, ProjectID: svc.ProjectID, From: monitor.Up, To: monitor.Down, Duration: 0, Error: "timeout"})
	a.Handle(ctx, monitor.Event{ServiceID: svc.ID, ProjectID: svc.ProjectID, From: monitor.Down, To: monitor.Up, Duration: 5 * time.Minute})

	if client.count() != 2 {
		t.Fatalf("posts = %d, want 2", client.count())
	}
	if !strings.Contains(client.posted[0], "DOWN") || !strings.Contains(client.posted[0], "API") {
		t.Fatalf("down message = %q", client.posted[0])
	}
	if !strings.Contains(client.posted[1], "RECOVERED") {
		t.Fatalf("recovery message = %q", client.posted[1])
	}

	failures, err := q.ListDeliveryFailuresByProject(ctx, store.ListDeliveryFailuresByProjectParams{
		ProjectID: svc.ProjectID, MaxRows: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(failures) != 0 {
		t.Fatalf("unexpected failures: %+v", failures)
	}
}

func TestHandleRecordsFailureWithoutChannel(t *testing.T) {
	a, q, svc, project, client := setupAlerter(t)
	ctx := context.Background()

	if err := q.SetProjectDefaultChannel(ctx, store.SetProjectDefaultChannelParams{DefaultChannelID: "", ID: project.ID}); err != nil {
		t.Fatal(err)
	}
	a.Handle(ctx, monitor.Event{ServiceID: svc.ID, ProjectID: svc.ProjectID, From: monitor.Up, To: monitor.Down})

	if client.count() != 0 {
		t.Fatalf("posts = %d, want 0", client.count())
	}
	failures, err := q.ListDeliveryFailuresByProject(ctx, store.ListDeliveryFailuresByProjectParams{
		ProjectID: svc.ProjectID, MaxRows: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(failures) != 1 || failures[0].Reason != "no channel resolved" {
		t.Fatalf("failures = %+v", failures)
	}
}

func TestHandleRecordsSlackError(t *testing.T) {
	a, q, svc, _, client := setupAlerter(t)
	ctx := context.Background()
	client.fail = true

	a.Handle(ctx, monitor.Event{ServiceID: svc.ID, ProjectID: svc.ProjectID, From: monitor.Up, To: monitor.Down})

	failures, err := q.ListDeliveryFailuresByProject(ctx, store.ListDeliveryFailuresByProjectParams{
		ProjectID: svc.ProjectID, MaxRows: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(failures) != 1 || !strings.Contains(failures[0].Reason, "channel_not_found") {
		t.Fatalf("failures = %+v", failures)
	}
}
