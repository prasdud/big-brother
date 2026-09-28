# Tasks: M2 — Check Engine

## 1. Schema
- [ ] 1.1 Migration for `checks`, `service_state`, `uptime_rollups`
- [ ] 1.2 Indexes on `(service_id, checked_at)` and `(project_id, checked_at)`
- [ ] 1.3 sqlc queries for inserting checks and reading history/uptime
- [ ] 1.4 Verify migrations apply on a database created in M1

## 2. HTTP checker
- [ ] 2.1 `Checker` interface and `HTTPChecker` implementation
- [ ] 2.2 Shared `http.Client` with timeout and idle-connection cap
- [ ] 2.3 Unit tests: 2xx up, 3xx up, 4xx down, timeout down, connection refused
- [ ] 2.4 Capture status code, latency, and error text

## 3. Scheduler
- [ ] 3.1 Ticker selects services with `next_run_at <= now`
- [ ] 3.2 Bounded worker pool driven by a due-services channel
- [ ] 3.3 Set `next_run_at = now + interval` on enqueue
- [ ] 3.4 Schedule all enabled services on boot
- [ ] 3.5 Respect `interval_seconds`, `timeout_seconds`, `enabled`
- [ ] 3.6 Verify a 5s service is checked roughly every 5s

## 4. State machine
- [ ] 4.1 Implement transitions from design.md in one component
- [ ] 4.2 Track consecutive failures/successes
- [ ] 4.3 Emit exactly one transition event per state change
- [ ] 4.4 Unit tests for every transition row
- [ ] 4.5 Pause sets `paused`; resume sets `pending`

## 5. History and rollups
- [ ] 5.1 Persist one `checks` row per completed check
- [ ] 5.2 Hourly job aggregates `uptime_rollups`
- [ ] 5.3 Retention pruning in bounded batches
- [ ] 5.4 Verify pruning removes rows older than `BB_RETENTION_DAYS`

## 6. API
- [ ] 6.1 `GET /status` returns current state and last change
- [ ] 6.2 `GET /checks` with time range and limit
- [ ] 6.3 `GET /uptime` for 24h/7d/30d
- [ ] 6.4 Verify against a service with a real failing URL

## 7. Verification
- [ ] 7.1 Point a service at a bad URL; observe `up → pending → down`
- [ ] 7.2 Restore the URL; observe `down → up`
- [ ] 7.3 Run `go test ./...` for the check and monitor packages
