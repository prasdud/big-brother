package kuma

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/prasdud/big-brother/internal/slug"
	"github.com/prasdud/big-brother/internal/store"
)

const fallbackProjectName = "Imported"

// Importer turns a Kuma export into projects and services.
type Importer struct {
	q           *store.Queries
	db          *sql.DB
	workspaceID string
}

// New builds an Importer.
func New(q *store.Queries, db *sql.DB, workspaceID string) *Importer {
	return &Importer{q: q, db: db, workspaceID: workspaceID}
}

// Preview parses the export and returns a report without writing anything.
func (im *Importer) Preview(ctx context.Context, data []byte) (Report, error) {
	export, err := Parse(data)
	if err != nil {
		return Report{}, err
	}
	return im.run(ctx, im.q, export, true)
}

// Apply writes the export atomically and returns a report.
func (im *Importer) Apply(ctx context.Context, data []byte) (Report, error) {
	export, err := Parse(data)
	if err != nil {
		return Report{}, err
	}
	tx, err := im.db.BeginTx(ctx, nil)
	if err != nil {
		return Report{}, err
	}
	defer func() { _ = tx.Rollback() }()

	report, err := im.run(ctx, im.q.WithTx(tx), export, false)
	if err != nil {
		return Report{}, err
	}
	if err := tx.Commit(); err != nil {
		return Report{}, err
	}
	return report, nil
}

func (im *Importer) run(ctx context.Context, q *store.Queries, export Export, dryRun bool) (Report, error) {
	report := Report{
		Version:  export.Version,
		Projects: []string{},
		Services: []ServiceReport{},
		Skipped:  []Skip{},
		Warnings: []string{},
	}
	seenProjects := map[string]bool{}
	seenServices := map[string]bool{}

	groupNames := map[int]string{}
	for _, monitor := range export.Monitors {
		if monitor.Type == groupType {
			groupNames[monitor.ID] = monitor.Name
		}
	}

	for _, notification := range export.Notifications {
		reason := "notification type is not supported"
		if notification.Type == "slack" {
			reason = "slack notification skipped; connect Slack and choose channels in the app"
		}
		report.Skipped = append(report.Skipped, Skip{Name: notification.Name, Type: notification.Type, Reason: reason})
	}

	projectFor := func(monitor Monitor) string {
		if monitor.Parent != nil {
			if name, ok := groupNames[*monitor.Parent]; ok {
				return name
			}
		}
		return fallbackProjectName
	}

	for _, monitor := range export.Monitors {
		if monitor.Type == groupType {
			if _, err := im.ensureProject(ctx, q, monitor.Name, dryRun, seenProjects, &report); err != nil {
				return report, err
			}
			continue
		}
		if !supportedTypes[monitor.Type] {
			report.Skipped = append(report.Skipped, Skip{
				Name: monitor.Name, Type: monitor.Type, Reason: "unsupported monitor type",
			})
			continue
		}

		projectName := projectFor(monitor)
		project, err := im.ensureProject(ctx, q, projectName, dryRun, seenProjects, &report)
		if err != nil {
			return report, err
		}
		action, err := im.ensureService(ctx, q, project, monitor, dryRun, seenServices)
		if err != nil {
			return report, err
		}
		report.Services = append(report.Services, ServiceReport{
			Name: monitor.Name, Project: projectName, Type: monitor.Type, Action: action,
		})
		if len(monitor.NotificationIDList) > 0 {
			report.Warnings = append(report.Warnings,
				fmt.Sprintf("service %q channel override was not imported; map Slack channels in the app", monitor.Name))
		}
	}
	return report, nil
}

func (im *Importer) ensureProject(ctx context.Context, q *store.Queries, name string, dryRun bool, seen map[string]bool, report *Report) (store.Project, error) {
	if name == "" {
		name = fallbackProjectName
	}
	slugValue := slug.Make(name)
	if slugValue == "" {
		slugValue = "imported"
	}

	if project, err := q.GetProjectBySlug(ctx, slugValue); err == nil {
		return project, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return store.Project{}, err
	}

	if !seen[slugValue] {
		report.Projects = append(report.Projects, slugValue)
		seen[slugValue] = true
	}
	if dryRun {
		return store.Project{Name: name, Slug: slugValue}, nil
	}

	project := store.Project{
		ID:          uuid.NewString(),
		WorkspaceID: im.workspaceID,
		Name:        name,
		Slug:        slugValue,
		CreatedAt:   store.NowUTC(),
	}
	if err := q.CreateProject(ctx, store.CreateProjectParams{
		ID: project.ID, WorkspaceID: project.WorkspaceID,
		Name: project.Name, Slug: project.Slug, CreatedAt: project.CreatedAt,
	}); err != nil {
		return store.Project{}, err
	}
	return project, nil
}

func (im *Importer) ensureService(ctx context.Context, q *store.Queries, project store.Project, monitor Monitor, dryRun bool, seen map[string]bool) (string, error) {
	slugValue := slug.Make(monitor.Name)
	if slugValue == "" {
		slugValue = "service"
	}
	key := project.Slug + "/" + slugValue
	if seen[key] {
		return "exists", nil
	}
	seen[key] = true

	if _, err := q.GetServiceBySlug(ctx, store.GetServiceBySlugParams{
		ProjectID: project.ID, Slug: slugValue,
	}); err == nil {
		return "exists", nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	if dryRun {
		return "create", nil
	}

	enabled := int64(1)
	if monitor.Active != nil && !*monitor.Active {
		enabled = 0
	}
	port := int64(0)
	if monitor.Port != nil {
		port = int64(*monitor.Port)
	}
	interval := int64(monitor.Interval)
	if interval < 1 {
		interval = 60
	}
	timeout := int64(monitor.Timeout)
	if timeout < 1 {
		timeout = 10
	}
	retries := int64(monitor.MaxRetries)
	if retries < 1 {
		retries = 3
	}

	now := store.NowUTC()
	err := q.CreateService(ctx, store.CreateServiceParams{
		ID:               uuid.NewString(),
		ProjectID:        project.ID,
		Name:             monitor.Name,
		Slug:             slugValue,
		Type:             monitor.Type,
		Url:              monitor.URL,
		Hostname:         monitor.Hostname,
		Port:             port,
		IntervalSeconds:  interval,
		TimeoutSeconds:   timeout,
		FailureThreshold: retries,
		Enabled:          enabled,
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	if err != nil {
		return "", err
	}
	return "create", nil
}
