// Package monitor persists check results and derives service state.
package monitor

import "time"

// State is the current status of a service.
type State string

const (
	Up      State = "up"
	Down    State = "down"
	Pending State = "pending"
	Paused  State = "paused"
)

// Snapshot is the state of a service before the next check.
type Snapshot struct {
	State                State
	ConsecutiveFailures  int
	ConsecutiveSuccesses int
	LastChangeAt         time.Time
}

// Event is emitted once per state change.
type Event struct {
	ServiceID string
	ProjectID string
	From      State
	To        State
	At        time.Time
	// Duration is the time spent in the previous state.
	Duration time.Duration
	// Error is the check error for a transition into down.
	Error string
}

// Next applies one check outcome to a snapshot and returns the new snapshot.
// A missing (empty) prior state is treated as pending.
func Next(prev Snapshot, up bool, threshold int) Snapshot {
	if prev.State == "" {
		prev.State = Pending
	}
	if threshold < 1 {
		threshold = 1
	}

	next := prev
	if up {
		next.ConsecutiveSuccesses = prev.ConsecutiveSuccesses + 1
		next.ConsecutiveFailures = 0
		next.State = Up
		return next
	}

	next.ConsecutiveFailures = prev.ConsecutiveFailures + 1
	next.ConsecutiveSuccesses = 0
	switch prev.State {
	case Up:
		next.State = Pending
	case Pending:
		if next.ConsecutiveFailures >= threshold {
			next.State = Down
		} else {
			next.State = Pending
		}
	default:
		next.State = Down
	}
	return next
}
