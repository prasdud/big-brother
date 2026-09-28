# Design: M2 — Check Engine

## Components

```
internal/check      Checker interface + HTTP implementation
internal/scheduler  in-process ticker, worker pool, due-service selection
internal/monitor    state machine, thresholds, transition events
internal/history    checks, rollups, retention pruning
```

## Scheduler

- A single ticker fires every second and selects services whose
  `next_run_at <= now`.
- Due services are pushed to a buffered channel consumed by N workers
  (`BB_CHECK_WORKERS`, default 16).
- `next_run_at` is set to `now + interval` on enqueue to avoid pile-up.
- On boot, all enabled services are scheduled immediately.

## HTTP checker

- `GET` the configured URL with `timeout_seconds`.
- Up when status is 200–399 within timeout; otherwise down with an error string.
- Captures status code, latency in milliseconds, and error text.
- Uses a shared `http.Client` with a `Transport` that caps idle connections.

## State machine

| From | Condition | To |
|---|---|---|
| `up` | first failure | `pending` |
| `pending` | consecutive failures ≥ threshold | `down` |
| `pending` | any success | `up` |
| `down` | first success | `up` |
| any | service paused | `paused` |
| `paused` | service resumed | `pending` |

- `pending` means "failing but not yet confirmed down".
- Every transition emits one event; exactly one event per transition prevents
  duplicate alerts in M4.
- Counters `consecutive_failures` and `consecutive_successes` live on the
  service or a companion state table.

## Schema additions

```sql
checks(id TEXT PK, service_id TEXT, project_id TEXT, checked_at,
       status TEXT, status_code INTEGER, latency_ms INTEGER, error TEXT)
service_state(service_id TEXT PK, state TEXT, consecutive_failures INTEGER,
              consecutive_successes INTEGER, last_change_at, last_check_at)
uptime_rollups(service_id TEXT PK, project_id TEXT, hour, up_checks INTEGER,
               total_checks INTEGER, PRIMARY KEY(service_id, hour))
```

Indexes: `checks(service_id, checked_at)` and `checks(project_id, checked_at)`.

## Retention

- `BB_RETENTION_DAYS` (default 30) controls raw `checks` retention.
- Pruning runs hourly, deleting in bounded batches to avoid long write locks.
- Rollups are retained longer than raw checks.

## API

- `GET /api/v1/projects/{project}/services/{service}/status`
- `GET .../checks?from=&to=&limit=`
- `GET .../uptime?window=24h|7d|30d`

## Risks

- Single-writer SQLite: keep writes short and batched; do not hold transactions
  across network calls.
- Clock skew: compute intervals from a monotonic clock where possible.
