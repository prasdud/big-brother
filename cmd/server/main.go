// Command server runs the big-brother uptime monitor: API and embedded web UI
// in a single binary.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prasdud/big-brother/internal/alert"
	"github.com/prasdud/big-brother/internal/api"
	"github.com/prasdud/big-brother/internal/auth"
	"github.com/prasdud/big-brother/internal/check"
	"github.com/prasdud/big-brother/internal/config"
	"github.com/prasdud/big-brother/internal/db"
	"github.com/prasdud/big-brother/internal/metrics"
	"github.com/prasdud/big-brother/internal/monitor"
	"github.com/prasdud/big-brother/internal/scheduler"
	"github.com/prasdud/big-brother/internal/secret"
	"github.com/prasdud/big-brother/internal/slack"
	"github.com/prasdud/big-brother/internal/store"
	"github.com/prasdud/big-brother/internal/web"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	slog.SetDefault(logger)

	sqldb, err := db.Open(cfg.DBPath)
	if err != nil {
		return err
	}
	defer sqldb.Close()

	if err := db.Migrate(sqldb); err != nil {
		return err
	}

	q := store.New(sqldb)
	ctx := context.Background()
	workspaceName := os.Getenv("BB_WORKSPACE_NAME")
	if workspaceName == "" {
		workspaceName = "Default"
	}
	workspace, err := store.EnsureWorkspace(ctx, q, workspaceName)
	if err != nil {
		return err
	}

	var authSvc *auth.Service
	if cfg.GoogleClientID != "" {
		verifier, err := auth.NewOIDCVerifier(ctx, cfg.OIDCIssuer, cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL)
		if err != nil {
			return err
		}
		authSvc = auth.New(q, verifier, auth.Config{
			WorkspaceID:         workspace.ID,
			AllowedDomains:      cfg.AllowedEmailDomains,
			BootstrapAdminEmail: cfg.BootstrapAdminEmail,
			SessionTTL:          cfg.SessionTTL,
			CookieSecure:        cfg.CookieSecure,
		})
	} else {
		logger.Warn("authentication disabled: BB_GOOGLE_CLIENT_ID is not set")
	}

	var slackDeps *api.SlackDeps
	var alerts *alert.Alerter
	if cfg.SlackClientID != "" {
		box, err := secret.NewBox(cfg.SecretKey)
		if err != nil {
			return err
		}
		alerts = alert.New(q, workspace.ID, box, nil, logger)
		slackDeps = &api.SlackDeps{
			Installer: slack.NewOAuth(cfg.SlackClientID, cfg.SlackClientSecret, cfg.SlackRedirectURL),
			Box:       box,
			Alerter:   alerts,
		}
	} else {
		logger.Warn("slack disabled: BB_SLACK_CLIENT_ID is not set")
	}

	reg := metrics.New()
	srv := api.New(q, sqldb, workspace, reg, web.Handler(), authSvc, slackDeps, logger)

	httpSrv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           srv.Router(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var events chan monitor.Event
	if alerts != nil {
		events = make(chan monitor.Event, 64)
	}
	emit := func(e monitor.Event) {
		logger.Info("state change",
			"service_id", e.ServiceID,
			"from", e.From,
			"to", e.To,
		)
		if events != nil {
			select {
			case events <- e:
			default:
				logger.Warn("alert queue full; dropping event", "service_id", e.ServiceID)
			}
		}
	}
	mon := monitor.New(q, emit)
	sch := scheduler.New(q, check.New(), mon, logger, cfg.CheckWorkers, cfg.RetentionDays)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		sch.Run(ctx)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		pruneSessions(ctx, q, logger)
	}()
	if events != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case e := <-events:
					alerts.Handle(ctx, e)
				}
			}
		}()
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server starting",
			"addr", cfg.Addr,
			"db", cfg.DBPath,
			"workspace", workspace.Slug,
			"check_workers", cfg.CheckWorkers,
		)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	var runErr error
	select {
	case runErr = <-errCh:
	case <-ctx.Done():
		logger.Info("shutdown requested")
	}
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil && runErr == nil {
		runErr = err
	}
	wg.Wait()
	if runErr != nil {
		return runErr
	}
	logger.Info("shutdown complete")
	return nil
}

func pruneSessions(ctx context.Context, q *store.Queries, logger *slog.Logger) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n, err := q.DeleteExpiredSessions(ctx, store.NowUTC())
			if err != nil {
				logger.Error("prune sessions", "error", err)
			} else if n > 0 {
				logger.Info("pruned sessions", "rows", n)
			}
		}
	}
}
