package api

import (
	"database/sql"
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/prasdud/big-brother/internal/store"
)

type statusView struct {
	State                string `json:"state"`
	LastChangeAt         string `json:"last_change_at"`
	LastCheckAt          string `json:"last_check_at"`
	ConsecutiveFailures  int64  `json:"consecutive_failures"`
	ConsecutiveSuccesses int64  `json:"consecutive_successes"`
}

type checkView struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	StatusCode int64  `json:"status_code"`
	LatencyMS  int64  `json:"latency_ms"`
	Error      string `json:"error"`
	CheckedAt  string `json:"checked_at"`
}

type uptimeView struct {
	Window      string  `json:"window"`
	UpChecks    int64   `json:"up_checks"`
	TotalChecks int64   `json:"total_checks"`
	Percent     float64 `json:"percent"`
}

func (s *Server) getStatus(w http.ResponseWriter, r *http.Request) {
	_, svc, err := s.serviceFromPath(r)
	if err != nil {
		notFound(w)
		return
	}

	st, err := s.q.GetServiceState(r.Context(), svc.ID)
	if errors.Is(err, sql.ErrNoRows) {
		state := "pending"
		if svc.Enabled == 0 {
			state = "paused"
		}
		writeJSON(w, http.StatusOK, statusView{State: state})
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}

	view := statusView{
		State:                st.State,
		LastChangeAt:         st.LastChangeAt,
		ConsecutiveFailures:  st.ConsecutiveFailures,
		ConsecutiveSuccesses: st.ConsecutiveSuccesses,
	}
	if st.LastCheckAt.Valid {
		view.LastCheckAt = st.LastCheckAt.String
	}
	writeJSON(w, http.StatusOK, view)
}

func (s *Server) listChecks(w http.ResponseWriter, r *http.Request) {
	_, svc, err := s.serviceFromPath(r)
	if err != nil {
		notFound(w)
		return
	}

	now := time.Now().UTC()
	from, ok := parseTime(w, r.URL.Query().Get("from"), now.Add(-24*time.Hour))
	if !ok {
		return
	}
	to, ok := parseTime(w, r.URL.Query().Get("to"), now)
	if !ok {
		return
	}
	if to.Before(from) {
		badRequest(w, "`to` must not be before `from`")
		return
	}
	limit, ok := parseLimit(w, r.URL.Query().Get("limit"))
	if !ok {
		return
	}

	checks, err := s.q.ListChecks(r.Context(), store.ListChecksParams{
		ServiceID: svc.ID,
		FromAt:    from.Format(time.RFC3339),
		ToAt:      to.Format(time.RFC3339),
		MaxRows:   int64(limit),
	})
	if err != nil {
		serverError(w, err)
		return
	}

	out := make([]checkView, 0, len(checks))
	for _, c := range checks {
		out = append(out, checkView{
			ID:         c.ID,
			Status:     c.Status,
			StatusCode: c.StatusCode,
			LatencyMS:  c.LatencyMs,
			Error:      c.Error,
			CheckedAt:  c.CheckedAt,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

var uptimeWindows = map[string]time.Duration{
	"24h": 24 * time.Hour,
	"7d":  7 * 24 * time.Hour,
	"30d": 30 * 24 * time.Hour,
}

func (s *Server) getUptime(w http.ResponseWriter, r *http.Request) {
	_, svc, err := s.serviceFromPath(r)
	if err != nil {
		notFound(w)
		return
	}

	window := r.URL.Query().Get("window")
	if window == "" {
		window = "24h"
	}
	d, ok := uptimeWindows[window]
	if !ok {
		badRequest(w, "window must be one of 24h, 7d, 30d")
		return
	}

	from := time.Now().UTC().Add(-d).Truncate(time.Hour)
	row, err := s.q.SumUptime(r.Context(), store.SumUptimeParams{
		ServiceID: svc.ID,
		FromHour:  from.Format(time.RFC3339),
	})
	if err != nil {
		serverError(w, err)
		return
	}

	percent := 100.0
	if row.TotalChecks > 0 {
		percent = math.Round(float64(row.UpChecks)/float64(row.TotalChecks)*10000) / 100
	}
	writeJSON(w, http.StatusOK, uptimeView{
		Window:      window,
		UpChecks:    row.UpChecks,
		TotalChecks: row.TotalChecks,
		Percent:     percent,
	})
}

func parseTime(w http.ResponseWriter, value string, fallback time.Time) (time.Time, bool) {
	if value == "" {
		return fallback, true
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		badRequest(w, "timestamps must be RFC3339")
		return time.Time{}, false
	}
	return t.UTC(), true
}

func parseLimit(w http.ResponseWriter, value string) (int, bool) {
	if value == "" {
		return 100, true
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 || n > 1000 {
		badRequest(w, "limit must be between 1 and 1000")
		return 0, false
	}
	return n, true
}
