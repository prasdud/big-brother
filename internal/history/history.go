// Package history retains and prunes check history.
package history

import (
	"context"
	"time"

	"github.com/prasdud/big-brother/internal/store"
)

const (
	pruneBatchSize  = 500
	pruneMaxBatches = 100
)

// PruneBefore deletes raw checks older than cutoff in bounded batches and
// returns the number of rows removed. Uptime rollups are left untouched.
func PruneBefore(ctx context.Context, q *store.Queries, cutoff time.Time) (int64, error) {
	before := cutoff.UTC().Format(time.RFC3339)
	var total int64
	for i := 0; i < pruneMaxBatches; i++ {
		n, err := q.PruneChecksBefore(ctx, store.PruneChecksBeforeParams{
			Before:  before,
			MaxRows: pruneBatchSize,
		})
		if err != nil {
			return total, err
		}
		total += n
		if n < pruneBatchSize {
			break
		}
	}
	return total, nil
}
