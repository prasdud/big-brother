package monitor

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/prasdud/big-brother/internal/check"
	"github.com/prasdud/big-brother/internal/store"
)

// Monitor records check results and derives service state.
type Monitor struct {
	q    *store.Queries
	now  func() time.Time
	emit func(Event)
}

// New builds a Monitor. emit is called once per state change and may be nil.
func New(q *store.Queries, emit func(Event)) *Monitor {
	return &Monitor{q: q, now: time.Now, emit: emit}
}

// Record persists one check result, updates state and rollups, and emits a
// transition event when the state changed.
func (m *Monitor) Record(ctx context.Context, svc store.Service, r check.Result) error {
	now := m.now().UTC()

	prev, err := m.snapshot(ctx, svc.ID)
	if err != nil {
		return err
	}
	next := Next(prev, r.Up, int(svc.FailureThreshold))
	changed := next.State != prev.State

	lastChange := prev.LastChangeAt
	if lastChange.IsZero() || changed {
		lastChange = now
	}

	status := "down"
	if r.Up {
		status = "up"
	}

	if err := m.q.InsertCheck(ctx, store.InsertCheckParams{
		ID:         uuid.NewString(),
		ServiceID:  svc.ID,
		ProjectID:  svc.ProjectID,
		CheckedAt:  now.Format(time.RFC3339),
		Status:     status,
		StatusCode: int64(r.StatusCode),
		LatencyMs:  r.LatencyMS,
		Error:      r.Error,
	}); err != nil {
		return err
	}

	if err := m.q.UpsertServiceState(ctx, store.UpsertServiceStateParams{
		ServiceID:            svc.ID,
		ProjectID:            svc.ProjectID,
		State:                string(next.State),
		ConsecutiveFailures:  int64(next.ConsecutiveFailures),
		ConsecutiveSuccesses: int64(next.ConsecutiveSuccesses),
		LastChangeAt:         lastChange.Format(time.RFC3339),
		LastCheckAt:          sql.NullString{String: now.Format(time.RFC3339), Valid: true},
	}); err != nil {
		return err
	}

	up := int64(0)
	if r.Up {
		up = 1
	}
	if err := m.q.UpsertRollup(ctx, store.UpsertRollupParams{
		ServiceID:   svc.ID,
		ProjectID:   svc.ProjectID,
		Hour:        now.Truncate(time.Hour).Format(time.RFC3339),
		UpChecks:    up,
		TotalChecks: 1,
	}); err != nil {
		return err
	}

	if changed && m.emit != nil {
		var spent time.Duration
		if !prev.LastChangeAt.IsZero() {
			spent = now.Sub(prev.LastChangeAt)
		}
		m.emit(Event{
			ServiceID: svc.ID,
			ProjectID: svc.ProjectID,
			From:      prev.State,
			To:        next.State,
			At:        now,
			Duration:  spent,
			Error:     r.Error,
		})
	}
	return nil
}

func (m *Monitor) snapshot(ctx context.Context, serviceID string) (Snapshot, error) {
	st, err := m.q.GetServiceState(ctx, serviceID)
	if errors.Is(err, sql.ErrNoRows) {
		return Snapshot{State: Pending}, nil
	}
	if err != nil {
		return Snapshot{}, err
	}

	var lastChange time.Time
	if t, perr := time.Parse(time.RFC3339, st.LastChangeAt); perr == nil {
		lastChange = t
	}
	return Snapshot{
		State:                State(st.State),
		ConsecutiveFailures:  int(st.ConsecutiveFailures),
		ConsecutiveSuccesses: int(st.ConsecutiveSuccesses),
		LastChangeAt:         lastChange,
	}, nil
}
