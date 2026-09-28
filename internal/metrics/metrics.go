// Package metrics provides a minimal Prometheus text-format registry using
// only the standard library.
package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
)

// Registry holds named counters.
type Registry struct {
	mu       sync.Mutex
	counters map[string]*counter
}

type counter struct {
	help  string
	value atomic.Int64
}

// New returns an empty registry.
func New() *Registry {
	return &Registry{counters: make(map[string]*counter)}
}

// Counter returns a named counter, creating it on first use.
func (r *Registry) Counter(name, help string) *Counter {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.counters[name]
	if !ok {
		c = &counter{help: help}
		r.counters[name] = c
	}
	return &Counter{c}
}

// Counter is a monotonically increasing metric.
type Counter struct{ c *counter }

// Inc adds one to the counter.
func (c *Counter) Inc() { c.c.value.Add(1) }

// Add adds n to the counter.
func (c *Counter) Add(n int64) { c.c.value.Add(n) }

// ServeHTTP writes the registry in Prometheus text exposition format.
func (r *Registry) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	r.mu.Lock()
	names := make([]string, 0, len(r.counters))
	for name := range r.counters {
		names = append(names, name)
	}
	sort.Strings(names)
	r.mu.Unlock()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	for _, name := range names {
		r.mu.Lock()
		c := r.counters[name]
		r.mu.Unlock()
		fmt.Fprintf(w, "# HELP %s %s\n", name, c.help)
		fmt.Fprintf(w, "# TYPE %s counter\n", name)
		fmt.Fprintf(w, "%s %d\n", name, c.value.Load())
	}
}
