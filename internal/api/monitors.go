package api

import (
	"math"
	"net/http"
	"time"

	"github.com/prasdud/big-brother/internal/store"
)

type heartbeat struct {
	Status    string `json:"status"`
	LatencyMS int64  `json:"latency_ms"`
	CheckedAt string `json:"checked_at"`
}

type monitorSummary struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Slug       string      `json:"slug"`
	Type       string      `json:"type"`
	URL        string      `json:"url"`
	Hostname   string      `json:"hostname"`
	Port       int64       `json:"port"`
	Enabled    bool        `json:"enabled"`
	Tags       []string    `json:"tags"`
	Interval   int64       `json:"interval_seconds"`
	State      string      `json:"state"`
	Uptime24h  float64     `json:"uptime_24h"`
	Uptime30d  float64     `json:"uptime_30d"`
	Heartbeats []heartbeat `json:"heartbeats"`
}

// listMonitors returns a compact per-service summary for the monitor list and
// detail view: current state, uptime windows, and recent heartbeats.
func (s *Server) listMonitors(w http.ResponseWriter, r *http.Request) {
	p, err := s.projectFromPath(r)
	if err != nil {
		notFound(w)
		return
	}
	services, err := s.q.ListServices(r.Context(), p.ID)
	if err != nil {
		serverError(w, err)
		return
	}

	now := time.Now().UTC()
	from24 := now.Add(-24 * time.Hour).Truncate(time.Hour).Format(time.RFC3339)
	from30 := now.Add(-30 * 24 * time.Hour).Truncate(time.Hour).Format(time.RFC3339)

	out := make([]monitorSummary, 0, len(services))
	for _, svc := range services {
		summary := monitorSummary{
			ID:       svc.ID,
			Name:     svc.Name,
			Slug:     svc.Slug,
			Type:     svc.Type,
			URL:      svc.Url,
			Hostname: svc.Hostname,
			Port:     svc.Port,
			Enabled:  svc.Enabled != 0,
			Tags:     splitTags(svc.Tags),
			Interval: svc.IntervalSeconds,
			State:    "pending",
		}
		if !summary.Enabled {
			summary.State = "paused"
		}
		if st, err := s.q.GetServiceState(r.Context(), svc.ID); err == nil {
			summary.State = st.State
		}
		if u, err := s.q.SumUptime(r.Context(), store.SumUptimeParams{ServiceID: svc.ID, FromHour: from24}); err == nil {
			summary.Uptime24h = uptimePercent(u)
		}
		if u, err := s.q.SumUptime(r.Context(), store.SumUptimeParams{ServiceID: svc.ID, FromHour: from30}); err == nil {
			summary.Uptime30d = uptimePercent(u)
		}

		checks, err := s.q.ListChecks(r.Context(), store.ListChecksParams{
			ServiceID: svc.ID, FromAt: "0000", ToAt: "9999", MaxRows: 30,
		})
		if err != nil {
			serverError(w, err)
			return
		}
		beats := make([]heartbeat, 0, len(checks))
		for i := len(checks) - 1; i >= 0; i-- {
			beats = append(beats, heartbeat{
				Status:    checks[i].Status,
				LatencyMS: checks[i].LatencyMs,
				CheckedAt: checks[i].CheckedAt,
			})
		}
		summary.Heartbeats = beats
		out = append(out, summary)
	}
	writeJSON(w, http.StatusOK, out)
}

func uptimePercent(row store.SumUptimeRow) float64 {
	if row.TotalChecks == 0 {
		return 100
	}
	return math.Round(float64(row.UpChecks)/float64(row.TotalChecks)*10000) / 100
}
