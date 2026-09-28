package monitor

import "testing"

func TestNext(t *testing.T) {
	const threshold = 3

	tests := []struct {
		name    string
		prev    Snapshot
		up      bool
		want    State
		wantF   int
		wantS   int
		changed bool
	}{
		{"up success stays up", Snapshot{State: Up}, true, Up, 0, 1, false},
		{"up first failure enters pending", Snapshot{State: Up}, false, Pending, 1, 0, true},
		{"pending below threshold stays pending", Snapshot{State: Pending, ConsecutiveFailures: 1}, false, Pending, 2, 0, false},
		{"pending at threshold goes down", Snapshot{State: Pending, ConsecutiveFailures: 2}, false, Down, 3, 0, true},
		{"pending any success goes up", Snapshot{State: Pending, ConsecutiveFailures: 2}, true, Up, 0, 1, true},
		{"down success goes up", Snapshot{State: Down, ConsecutiveFailures: 5}, true, Up, 0, 1, true},
		{"down failure stays down", Snapshot{State: Down, ConsecutiveFailures: 5}, false, Down, 6, 0, false},
		{"empty first failure is pending", Snapshot{}, false, Pending, 1, 0, false},
		{"empty first success is up", Snapshot{}, true, Up, 0, 1, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Next(tc.prev, tc.up, threshold)
			if got.State != tc.want {
				t.Fatalf("state = %s, want %s", got.State, tc.want)
			}
			if got.ConsecutiveFailures != tc.wantF {
				t.Fatalf("failures = %d, want %d", got.ConsecutiveFailures, tc.wantF)
			}
			if got.ConsecutiveSuccesses != tc.wantS {
				t.Fatalf("successes = %d, want %d", got.ConsecutiveSuccesses, tc.wantS)
			}
			if changed := got.State != normalize(tc.prev); changed != tc.changed {
				t.Fatalf("changed = %v, want %v", changed, tc.changed)
			}
		})
	}
}

func normalize(s Snapshot) State {
	if s.State == "" {
		return Pending
	}
	return s.State
}
