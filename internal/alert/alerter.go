package alert

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/prasdud/big-brother/internal/metrics"
	"github.com/prasdud/big-brother/internal/monitor"
	"github.com/prasdud/big-brother/internal/secret"
	"github.com/prasdud/big-brother/internal/slack"
	"github.com/prasdud/big-brother/internal/store"
)

// Alerter turns state transitions into Slack messages and records failures.
type Alerter struct {
	q           *store.Queries
	workspaceID string
	box         *secret.Box
	newClient   func(token string) slack.Client
	logger      *slog.Logger
	now         func() time.Time
	// Metrics is optional; when set, delivery counters are recorded.
	Metrics *metrics.Registry
}

// New builds an Alerter. newClient is injected so tests can avoid the network.
func New(q *store.Queries, workspaceID string, box *secret.Box, newClient func(token string) slack.Client, logger *slog.Logger) *Alerter {
	if newClient == nil {
		newClient = func(token string) slack.Client { return slack.NewAPIClient(token) }
	}
	return &Alerter{
		q:           q,
		workspaceID: workspaceID,
		box:         box,
		newClient:   newClient,
		logger:      logger,
		now:         time.Now,
	}
}

// Handle delivers the alert for a transition, if one applies. It never returns
// an error: failures are recorded and logged.
func (a *Alerter) Handle(ctx context.Context, e monitor.Event) {
	trigger := TriggerFor(string(e.From), string(e.To))
	if trigger == "" {
		return
	}

	svc, err := a.q.GetServiceByID(ctx, e.ServiceID)
	if err != nil {
		a.logger.Error("alert: load service", "service_id", e.ServiceID, "error", err)
		return
	}
	project, err := a.q.GetProjectByID(ctx, e.ProjectID)
	if err != nil {
		a.logger.Error("alert: load project", "project_id", e.ProjectID, "error", err)
		return
	}

	projectTemplate := ""
	if t, err := a.q.GetAlertTemplate(ctx, store.GetAlertTemplateParams{
		ProjectID: project.ID,
		Trigger:   trigger,
	}); err == nil {
		projectTemplate = t.Body
	} else if !errors.Is(err, sql.ErrNoRows) {
		a.logger.Error("alert: load template", "project_id", project.ID, "error", err)
	}

	text := Render(ResolveTemplate(svc, projectTemplate, trigger), Vars{
		ServiceName: svc.Name,
		ServiceURL:  svc.Url,
		ProjectName: project.Name,
		Status:      string(e.To),
		Duration:    e.Duration.String(),
		Error:       e.Error,
	})

	channelID := ResolveChannelID(svc, project)
	if channelID == "" {
		a.recordFailure(ctx, svc, trigger, "no channel resolved")
		return
	}
	channel, err := a.q.GetChannelByID(ctx, channelID)
	if err != nil {
		a.recordFailure(ctx, svc, trigger, "channel not found")
		return
	}

	if err := a.send(ctx, channel.SlackChannelID, text); err != nil {
		a.recordFailure(ctx, svc, trigger, err.Error())
		return
	}
	if a.Metrics != nil {
		a.Metrics.CounterVec("bb_alert_deliveries_total", "Alert deliveries by trigger and result.", "trigger", "result").
			With(trigger, "delivered").Inc()
	}
	a.logger.Info("alert delivered",
		"service", svc.Slug,
		"trigger", trigger,
		"channel", channel.Name,
	)
}

// ListChannels lists channels from the connected Slack workspace.
func (a *Alerter) ListChannels(ctx context.Context) ([]slack.Channel, error) {
	client, err := a.client(ctx)
	if err != nil {
		return nil, err
	}
	return client.ListChannels(ctx)
}

// TestSend posts a message to a Slack channel.
func (a *Alerter) TestSend(ctx context.Context, slackChannelID, text string) error {
	return a.send(ctx, slackChannelID, text)
}

// Status reports whether a Slack workspace is connected.
func (a *Alerter) Status(ctx context.Context) (teamName string, connected bool, err error) {
	ws, err := a.q.GetSlackWorkspace(ctx, a.workspaceID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return ws.TeamName, true, nil
}

func (a *Alerter) send(ctx context.Context, slackChannelID, text string) error {
	client, err := a.client(ctx)
	if err != nil {
		return err
	}
	return client.PostMessage(ctx, slackChannelID, text)
}

func (a *Alerter) client(ctx context.Context) (slack.Client, error) {
	ws, err := a.q.GetSlackWorkspace(ctx, a.workspaceID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("slack workspace is not connected")
	}
	if err != nil {
		return nil, err
	}
	token, err := a.box.Decrypt(ws.BotTokenEnc)
	if err != nil {
		return nil, fmt.Errorf("decrypt bot token: %w", err)
	}
	return a.newClient(string(token)), nil
}

func (a *Alerter) recordFailure(ctx context.Context, svc store.Service, trigger, reason string) {
	if a.Metrics != nil {
		a.Metrics.CounterVec("bb_alert_deliveries_total", "Alert deliveries by trigger and result.", "trigger", "result").
			With(trigger, "failed").Inc()
		a.Metrics.CounterVec("bb_delivery_failures_total", "Alert delivery failures by trigger.", "trigger").
			With(trigger).Inc()
	}
	a.logger.Error("alert delivery failed",
		"service", svc.Slug,
		"trigger", trigger,
		"reason", reason,
	)
	err := a.q.CreateDeliveryFailure(ctx, store.CreateDeliveryFailureParams{
		ID:        uuid.NewString(),
		ProjectID: svc.ProjectID,
		ServiceID: svc.ID,
		Trigger:   trigger,
		Reason:    reason,
		CreatedAt: a.now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		a.logger.Error("alert: record delivery failure", "error", err)
	}
}
