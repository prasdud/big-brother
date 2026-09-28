// Package scheduler runs checks on each service's interval.
package scheduler

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"
	"time"

	"github.com/prasdud/big-brother/internal/check"
	"github.com/prasdud/big-brother/internal/history"
	"github.com/prasdud/big-brother/internal/metrics"
	"github.com/prasdud/big-brother/internal/monitor"
	"github.com/prasdud/big-brother/internal/store"
)

const (
	tickInterval = time.Second
	maxDueBatch  = 256
)

// Scheduler dispatches due services to a worker pool.
type Scheduler struct {
	q             *store.Queries
	checker       check.Checker
	monitor       *monitor.Monitor
	logger        *slog.Logger
	workers       int
	retentionDays int
	checks        *metrics.CounterVec
	duration      *metrics.HistogramVec
}

// New builds a Scheduler.
func New(q *store.Queries, checker check.Checker, mon *monitor.Monitor, logger *slog.Logger, workers, retentionDays int, reg *metrics.Registry) *Scheduler {
	if workers < 1 {
		workers = 1
	}
	return &Scheduler{
		q:             q,
		checker:       checker,
		monitor:       mon,
		logger:        logger,
		workers:       workers,
		retentionDays: retentionDays,
		checks:        reg.CounterVec("bb_checks_total", "Checks executed by type and result.", "type", "result"),
		duration:      reg.Histogram("bb_check_duration_seconds", "Check latency in seconds.", []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}, "type"),
	}
}

// Run blocks until ctx is canceled, then drains in-flight checks.
func (s *Scheduler) Run(ctx context.Context) {
	jobs := make(chan store.Service, maxDueBatch)
	var wg sync.WaitGroup
	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.worker(ctx, jobs)
		}()
	}

	tick := time.NewTicker(tickInterval)
	defer tick.Stop()
	prune := time.NewTicker(time.Hour)
	defer prune.Stop()

	s.prune(ctx)
	for {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return
		case <-tick.C:
			s.dispatch(ctx, jobs)
		case <-prune.C:
			s.prune(ctx)
		}
	}
}

func (s *Scheduler) dispatch(ctx context.Context, jobs chan<- store.Service) {
	now := time.Now().UTC()
	due, err := s.q.ListDueServices(ctx, store.ListDueServicesParams{
		Now:     sql.NullString{String: now.Format(time.RFC3339), Valid: true},
		MaxRows: maxDueBatch,
	})
	if err != nil {
		s.logger.Error("list due services", "error", err)
		return
	}

	for _, svc := range due {
		next := now.Add(time.Duration(svc.IntervalSeconds) * time.Second)
		if err := s.q.SetServiceNextRun(ctx, store.SetServiceNextRunParams{
			NextRunAt: sql.NullString{String: next.Format(time.RFC3339), Valid: true},
			ID:        svc.ID,
		}); err != nil {
			s.logger.Error("schedule next run", "service", svc.Slug, "error", err)
			continue
		}
		select {
		case jobs <- svc:
		default:
			s.logger.Warn("check queue full, skipping service", "service", svc.Slug)
		}
	}
}

func (s *Scheduler) worker(ctx context.Context, jobs <-chan store.Service) {
	bg := context.WithoutCancel(ctx)
	for svc := range jobs {
		result := s.checker.Check(bg, svc)
		if err := s.monitor.Record(bg, svc, result); err != nil {
			s.logger.Error("record check", "service", svc.Slug, "error", err)
			continue
		}
		outcome := "down"
		if result.Up {
			outcome = "up"
		}
		s.checks.With(svc.Type, outcome).Inc()
		s.duration.With(svc.Type).Observe(float64(result.LatencyMS) / 1000)
		s.logger.Debug("check",
			"service", svc.Slug,
			"up", result.Up,
			"status", result.StatusCode,
			"latency_ms", result.LatencyMS,
		)
	}
}

func (s *Scheduler) prune(ctx context.Context) {
	if s.retentionDays <= 0 {
		return
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -s.retentionDays)
	n, err := history.PruneBefore(ctx, s.q, cutoff)
	if err != nil {
		s.logger.Error("prune checks", "error", err)
		return
	}
	if n > 0 {
		s.logger.Info("pruned checks", "rows", n, "before", cutoff.Format(time.RFC3339))
	}
}
