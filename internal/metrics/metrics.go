// Package metrics provides a minimal Prometheus text-format registry using
// only the standard library.
package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// Registry holds counters and histograms.
type Registry struct {
	mu       sync.Mutex
	counters map[string]*counterFamily
	hists    map[string]*histFamily
}

// New returns an empty registry.
func New() *Registry {
	return &Registry{
		counters: make(map[string]*counterFamily),
		hists:    make(map[string]*histFamily),
	}
}

// Counter returns a counter family without labels, creating it on first use.
func (r *Registry) Counter(name, help string) *Counter {
	return r.CounterVec(name, help).With()
}

// CounterVec returns a labeled counter family, creating it on first use.
func (r *Registry) CounterVec(name, help string, labels ...string) *CounterVec {
	r.mu.Lock()
	defer r.mu.Unlock()
	fam, ok := r.counters[name]
	if !ok {
		fam = &counterFamily{name: name, help: help, labels: labels, series: map[string]*atomic.Int64{}}
		r.counters[name] = fam
	}
	return &CounterVec{fam: fam}
}

// Histogram returns a labeled histogram family with the given upper bounds.
func (r *Registry) Histogram(name, help string, buckets []float64, labels ...string) *HistogramVec {
	r.mu.Lock()
	defer r.mu.Unlock()
	fam, ok := r.hists[name]
	if !ok {
		fam = &histFamily{
			name: name, help: help, labels: labels, buckets: buckets,
			series: map[string]*histState{},
		}
		r.hists[name] = fam
	}
	return &HistogramVec{fam: fam}
}

type counterFamily struct {
	name, help string
	labels     []string
	mu         sync.Mutex
	series     map[string]*atomic.Int64
}

// CounterVec is a labeled counter family.
type CounterVec struct{ fam *counterFamily }

// With selects a series by label values.
func (v *CounterVec) With(labelValues ...string) *Counter {
	key := seriesKey(v.fam.labels, labelValues)
	v.fam.mu.Lock()
	c, ok := v.fam.series[key]
	if !ok {
		c = &atomic.Int64{}
		v.fam.series[key] = c
	}
	v.fam.mu.Unlock()
	return &Counter{c: c}
}

// Counter is a monotonically increasing series.
type Counter struct{ c *atomic.Int64 }

// Inc adds one.
func (c *Counter) Inc() { c.c.Add(1) }

// Add adds n.
func (c *Counter) Add(n int64) { c.c.Add(n) }

type histFamily struct {
	name, help string
	labels     []string
	buckets    []float64
	mu         sync.Mutex
	series     map[string]*histState
}

type histState struct {
	mu     sync.Mutex
	counts []int64
	sum    float64
	count  int64
}

// HistogramVec is a labeled histogram family.
type HistogramVec struct{ fam *histFamily }

// With selects a series by label values.
func (v *HistogramVec) With(labelValues ...string) *Histogram {
	key := seriesKey(v.fam.labels, labelValues)
	v.fam.mu.Lock()
	h, ok := v.fam.series[key]
	if !ok {
		h = &histState{counts: make([]int64, len(v.fam.buckets))}
		v.fam.series[key] = h
	}
	v.fam.mu.Unlock()
	return &Histogram{fam: v.fam, key: key, labels: labelValues, state: h}
}

// Histogram is a single histogram series.
type Histogram struct {
	fam    *histFamily
	key    string
	labels []string
	state  *histState
}

// Observe records one value.
func (h *Histogram) Observe(value float64) {
	h.state.mu.Lock()
	defer h.state.mu.Unlock()
	for i, bound := range h.fam.buckets {
		if value <= bound {
			h.state.counts[i]++
			break
		}
	}
	h.state.sum += value
	h.state.count++
}

// ServeHTTP writes the registry in Prometheus text exposition format.
func (r *Registry) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	r.mu.Lock()
	counterNames := sortedKeys(r.counters)
	histNames := sortedKeys(r.hists)
	r.mu.Unlock()

	for _, name := range counterNames {
		r.mu.Lock()
		fam := r.counters[name]
		r.mu.Unlock()
		fmt.Fprintf(w, "# HELP %s %s\n# TYPE %s counter\n", fam.name, fam.help, fam.name)

		fam.mu.Lock()
		keys := sortedKeys(fam.series)
		for _, key := range keys {
			value := fam.series[key].Load()
			fmt.Fprintf(w, "%s%s %d\n", fam.name, labelString(fam.labels, key), value)
		}
		fam.mu.Unlock()
	}

	for _, name := range histNames {
		r.mu.Lock()
		fam := r.hists[name]
		r.mu.Unlock()
		fmt.Fprintf(w, "# HELP %s %s\n# TYPE %s histogram\n", fam.name, fam.help, fam.name)

		fam.mu.Lock()
		keys := sortedKeys(fam.series)
		for _, key := range keys {
			state := fam.series[key]
			state.mu.Lock()
			var cumulative int64
			for i, bound := range fam.buckets {
				cumulative += state.counts[i]
				fmt.Fprintf(w, "%s_bucket%s %d\n",
					fam.name, labelStringWith(fam.labels, key, "le", strconv.FormatFloat(bound, 'g', -1, 64)), cumulative)
			}
			fmt.Fprintf(w, "%s_bucket%s %d\n",
				fam.name, labelStringWith(fam.labels, key, "le", "+Inf"), state.count)
			fmt.Fprintf(w, "%s_sum%s %g\n", fam.name, labelString(fam.labels, key), state.sum)
			fmt.Fprintf(w, "%s_count%s %d\n", fam.name, labelString(fam.labels, key), state.count)
			state.mu.Unlock()
		}
		fam.mu.Unlock()
	}
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// seriesKey encodes label values into a stable key.
func seriesKey(labels, values []string) string {
	parts := make([]string, 0, len(values))
	for i, value := range values {
		name := ""
		if i < len(labels) {
			name = labels[i]
		}
		parts = append(parts, name+"="+value)
	}
	return strings.Join(parts, "\x00")
}

func labelString(labels []string, key string) string {
	if key == "" {
		return ""
	}
	pairs := strings.Split(key, "\x00")
	return "{" + strings.Join(pairs, ",") + "}"
}

func labelStringWith(labels []string, key, extraName, extraValue string) string {
	pairs := []string{}
	if key != "" {
		pairs = append(pairs, strings.Split(key, "\x00")...)
	}
	pairs = append(pairs, extraName+"="+strconv.Quote(extraValue))
	return "{" + strings.Join(pairs, ",") + "}"
}
