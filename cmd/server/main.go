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

	"github.com/prasdud/big-brother/internal/api"
	"github.com/prasdud/big-brother/internal/check"
	"github.com/prasdud/big-brother/internal/config"
	"github.com/prasdud/big-brother/internal/db"
	"github.com/prasdud/big-brother/internal/metrics"
	"github.com/prasdud/big-brother/internal/monitor"
	"github.com/prasdud/big-brother/internal/scheduler"
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

	reg := metrics.New()
	srv := api.New(q, sqldb, workspace, reg, web.Handler(), logger)

	httpSrv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           srv.Router(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	emit := func(e monitor.Event) {
		logger.Info("state change",
			"service_id", e.ServiceID,
			"from", e.From,
			"to", e.To,
		)
	}
	mon := monitor.New(q, emit)
	sch := scheduler.New(q, check.New(), mon, logger, cfg.CheckWorkers, cfg.RetentionDays)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		sch.Run(ctx)
	}()

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
