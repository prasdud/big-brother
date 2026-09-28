package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// NowUTC returns the current time formatted for storage.
func NowUTC() string { return time.Now().UTC().Format(time.RFC3339) }

// EnsureWorkspace returns the singleton workspace, creating it when the
// database is empty.
func EnsureWorkspace(ctx context.Context, q *Queries, name string) (Workspace, error) {
	n, err := q.CountWorkspaces(ctx)
	if err != nil {
		return Workspace{}, err
	}
	if n > 0 {
		return q.GetWorkspace(ctx)
	}
	err = q.CreateWorkspace(ctx, CreateWorkspaceParams{
		ID:        uuid.NewString(),
		Name:      name,
		Slug:      "default",
		CreatedAt: NowUTC(),
	})
	if err != nil {
		return Workspace{}, err
	}
	return q.GetWorkspace(ctx)
}
