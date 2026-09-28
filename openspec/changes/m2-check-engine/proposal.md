# Proposal: M2 — Check Engine

## Intent

Turn persisted services into something that is actually monitored. A background
scheduler runs HTTP checks on each service's interval, records results, derives
the up/down/pending/paused state, and exposes current status, history, and
uptime.

## Scope

- HTTP checker: request the configured URL, respect timeout, treat 2xx and 3xx
  as up.
- Scheduler that runs each enabled service on its own interval.
- State machine: `up`, `down`, `pending`, `paused`.
- Failure threshold and recovery threshold to damp flapping.
- Check history rows and hourly uptime rollups.
- Configurable retention with batched pruning.
- API for current status, check history, and uptime per service.

## Approach

- One in-process ticker loop owns scheduling; each due service is dispatched to
  a bounded worker pool so a slow service cannot block others.
- Every check writes one `checks` row and possibly one state transition.
- State transitions are computed in one place so M4 can subscribe to them.
- Rollups are written hourly; raw rows are pruned in batches by retention.

## Out of scope

- TCP/DNS check types (M6), alerting (M4), UI (M5), auth enforcement (M3).

## Success signals

- A service pointing at a failing URL moves `up → pending → down` per threshold.
- Check history and uptime are queryable per service.
- Retention pruning keeps the database bounded.
