// Package check runs a single health check against a service.
package check

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
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

// New returns the default checker, dispatching on the service type.
func New() Checker {
	return &dispatcher{http: &httpChecker{client: &http.Client{
		// Report the target's own response instead of chasing redirects.
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 4,
			IdleConnTimeout:     90 * time.Second,
		},
	}}}
}

type dispatcher struct {
	http *httpChecker
}

func (d *dispatcher) Check(ctx context.Context, svc store.Service) Result {
	switch svc.Type {
	case "http":
		return d.http.Check(ctx, svc)
	case "tcp":
		return tcpCheck(ctx, svc)
	case "dns":
		return dnsCheck(ctx, svc)
	default:
		return Result{Error: fmt.Sprintf("unsupported check type: %s", svc.Type)}
	}
}

type httpChecker struct {
	client *http.Client
}

func (h *httpChecker) Check(ctx context.Context, svc store.Service) Result {
	ctx, cancel := context.WithTimeout(ctx, timeout(svc))
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

func tcpCheck(ctx context.Context, svc store.Service) Result {
	ctx, cancel := context.WithTimeout(ctx, timeout(svc))
	defer cancel()

	start := time.Now()
	address := net.JoinHostPort(svc.Hostname, strconv.FormatInt(svc.Port, 10))
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", address)
	if err != nil {
		return Result{Error: err.Error(), LatencyMS: elapsed(start)}
	}
	_ = conn.Close()
	return Result{Up: true, LatencyMS: elapsed(start)}
}

func dnsCheck(ctx context.Context, svc store.Service) Result {
	ctx, cancel := context.WithTimeout(ctx, timeout(svc))
	defer cancel()

	start := time.Now()
	addresses, err := net.DefaultResolver.LookupHost(ctx, svc.Hostname)
	if err != nil {
		return Result{Error: err.Error(), LatencyMS: elapsed(start)}
	}
	if len(addresses) == 0 {
		return Result{Error: "no addresses returned", LatencyMS: elapsed(start)}
	}
	return Result{Up: true, LatencyMS: elapsed(start)}
}

func timeout(svc store.Service) time.Duration {
	return time.Duration(svc.TimeoutSeconds) * time.Second
}

func elapsed(start time.Time) int64 {
	return time.Since(start).Milliseconds()
}
