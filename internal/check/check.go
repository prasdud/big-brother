// Package check runs a single health check against a service.
package check

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/prasdud/big-brother/internal/store"
)

// Result is the outcome of one check.
type Result struct {
	Up         bool
	StatusCode int
	LatencyMS  int64
	Error      string
}

// Checker performs one check for a service.
type Checker interface {
	Check(ctx context.Context, svc store.Service) Result
}

// New returns the default checker. HTTP is the only type in M1; TCP and DNS
// arrive in M6.
func New() Checker {
	return &httpChecker{client: &http.Client{
		// Report the target's own response instead of chasing redirects.
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 4,
			IdleConnTimeout:     90 * time.Second,
		},
	}}
}

type httpChecker struct {
	client *http.Client
}

func (h *httpChecker) Check(ctx context.Context, svc store.Service) Result {
	if svc.Type != "http" {
		return Result{Error: fmt.Sprintf("unsupported check type: %s", svc.Type)}
	}

	timeout := time.Duration(svc.TimeoutSeconds) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, svc.Url, nil)
	if err != nil {
		return Result{Error: err.Error(), LatencyMS: elapsed(start)}
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return Result{Error: err.Error(), LatencyMS: elapsed(start)}
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<16))

	res := Result{StatusCode: resp.StatusCode, LatencyMS: elapsed(start)}
	if resp.StatusCode >= 200 && resp.StatusCode <= 399 {
		res.Up = true
	} else {
		res.Error = fmt.Sprintf("unexpected status %d", resp.StatusCode)
	}
	return res
}

func elapsed(start time.Time) int64 {
	return time.Since(start).Milliseconds()
}
