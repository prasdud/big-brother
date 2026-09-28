package metrics

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegistryRendersCountersAndHistograms(t *testing.T) {
	reg := New()
	reg.CounterVec("bb_checks_total", "checks", "type", "result").With("http", "up").Inc()
	reg.CounterVec("bb_checks_total", "checks", "type", "result").With("http", "up").Inc()
	reg.Histogram("bb_check_duration_seconds", "duration", []float64{0.1, 1}, "type").With("http").Observe(0.05)
	reg.Histogram("bb_check_duration_seconds", "duration", []float64{0.1, 1}, "type").With("http").Observe(0.5)

	rec := httptest.NewRecorder()
	reg.ServeHTTP(rec, httptest.NewRequest("GET", "/metrics", nil))
	body := rec.Body.String()

	for _, want := range []string{
		"# TYPE bb_checks_total counter",
		"bb_checks_total{type=http,result=up} 2",
		"# TYPE bb_check_duration_seconds histogram",
		`bb_check_duration_seconds_bucket{type=http,le="0.1"} 1`,
		`bb_check_duration_seconds_bucket{type=http,le="1"} 2`,
		`bb_check_duration_seconds_bucket{type=http,le="+Inf"} 2`,
		`bb_check_duration_seconds_count{type=http} 2`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("metrics output missing %q\n%s", want, body)
		}
	}
}
