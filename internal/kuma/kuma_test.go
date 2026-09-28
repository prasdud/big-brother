package kuma

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prasdud/big-brother/internal/db"
	"github.com/prasdud/big-brother/internal/store"
)

const sampleExport = `{
  "version": "1.23.17",
  "monitors": [
    {"id": 1, "name": "Payments", "type": "group", "active": true},
    {"id": 2, "name": "Payments API", "type": "http", "url": "https://payments.example/health",
     "interval": 30, "timeout": 8, "maxretries": 4, "parent": 1, "active": true, "notificationIDList": {"7": true}},
    {"id": 3, "name": "Postgres", "type": "tcp", "hostname": "db.example", "port": 5432,
     "interval": 60, "timeout": 5, "maxretries": 2, "active": true},
    {"id": 4, "name": "gRPC", "type": "grpc", "hostname": "grpc.example", "port": 443, "active": true}
  ],
  "notifications": [
    {"id": 7, "name": "slack-alerts", "type": "slack", "isDefault": true,
     "config": "{\"webhookURL\":\"https://hooks.slack.com/services/SECRETPART\"}"}
  ]
}`

func setup(t *testing.T) (*Importer, *store.Queries) {
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
	return New(q, sqldb, ws.ID), q
}

func TestPreviewDoesNotWrite(t *testing.T) {
	importer, q := setup(t)
	ctx := context.Background()

	report, err := importer.Preview(ctx, []byte(sampleExport))
	if err != nil {
		t.Fatal(err)
	}
	if report.Version != "1.23.17" {
		t.Fatalf("version = %q", report.Version)
	}
	if len(report.Projects) != 2 { // "payments" and "imported"
		t.Fatalf("projects = %v", report.Projects)
	}
	if len(report.Services) != 2 {
		t.Fatalf("services = %+v", report.Services)
	}
	var skippedTypes []string
	for _, skip := range report.Skipped {
		skippedTypes = append(skippedTypes, skip.Type)
	}
	if !contains(skippedTypes, "grpc") || !contains(skippedTypes, "slack") {
		t.Fatalf("skipped = %+v", report.Skipped)
	}

	projects, err := q.ListProjects(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 0 {
		t.Fatalf("preview wrote %d projects", len(projects))
	}
}

func TestApplyIsIdempotent(t *testing.T) {
	importer, q := setup(t)
	ctx := context.Background()

	first, err := importer.Apply(ctx, []byte(sampleExport))
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Projects) != 2 || len(first.Services) != 2 {
		t.Fatalf("first apply = %+v", first)
	}
	for _, service := range first.Services {
		if service.Action != "create" {
			t.Fatalf("expected create, got %+v", service)
		}
	}

	project, err := q.GetProjectBySlug(ctx, "payments")
	if err != nil {
		t.Fatal(err)
	}
	service, err := q.GetServiceBySlug(ctx, store.GetServiceBySlugParams{ProjectID: project.ID, Slug: "payments-api"})
	if err != nil {
		t.Fatal(err)
	}
	if service.Url != "https://payments.example/health" || service.IntervalSeconds != 30 || service.TimeoutSeconds != 8 || service.FailureThreshold != 4 {
		t.Fatalf("service = %+v", service)
	}

	second, err := importer.Apply(ctx, []byte(sampleExport))
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Projects) != 0 {
		t.Fatalf("second apply created projects: %v", second.Projects)
	}
	for _, service := range second.Services {
		if service.Action != "exists" {
			t.Fatalf("second apply action = %+v", service)
		}
	}
	services, err := q.ListServices(ctx, project.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(services) != 1 {
		t.Fatalf("payments services = %d, want 1", len(services))
	}
}

func TestReportNeverContainsSecrets(t *testing.T) {
	importer, _ := setup(t)
	report, err := importer.Preview(context.Background(), []byte(sampleExport))
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "SECRETPART") || strings.Contains(string(encoded), "webhookURL") {
		t.Fatalf("report leaked a secret: %s", encoded)
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
