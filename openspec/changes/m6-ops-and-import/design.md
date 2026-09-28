# Design: M6 — Operations, OpenAPI, Checks, and Import

## OpenAPI and client

- `internal/api/openapi.yaml` is the source of truth; served at `/api/v1/openapi.json`.
- `openapi-generator` (or `oapi-codegen` plus a TS generator) produces
  `web/src/lib/api/generated`.
- CI validates that regenerating produces no diff.
- Server handlers are checked against the spec by contract tests.

## Additional check types

- `TCPChecker`: dial host:port within timeout; success is a completed
  connection.
- `DNSChecker`: resolve hostname for the configured record; success is a
  non-empty answer within timeout.
- Both return the same result struct as HTTP, so `scheduler` and `monitor` are
  untouched. New services select the type via the M1 `type` column.

## Kuma import

Input: a Kuma JSON export (sample version 1.23.17).

Mapping:

| Kuma | Product |
|---|---|
| group monitor | project |
| monitor | service (`name`, `type`, `url`, `hostname`, `port`, `interval`, `timeout`, `maxretries`) |
| slack notification | channel |
| `notificationIDList` | service channel override |

- Unsupported types (telegram, mqtt, grpc, and any other) are skipped and listed
  in the import report with a reason.
- `parent` is ignored; the product has no nested-monitor concept.
- History is not imported.
- Secrets in the file are encrypted at rest and never logged.

Two-phase API:

- `POST /api/v1/import/kuma/preview` returns the full planned report: counts,
  created resources, skipped monitors, and warnings. No writes.
- `POST /api/v1/import/kuma/apply` applies the same payload atomically and
  returns the report.
- Import is idempotent-safe: resources are matched by stable slug, so applying
  the same export twice does not duplicate projects, services, or channels.

## Metrics

- `bb_checks_total{type,result}`, `bb_check_duration_seconds`,
  `bb_state_transitions_total{from,to}`,
  `bb_alert_deliveries_total{trigger,result}`, `bb_delivery_failures_total`.

## Packaging and docs

- Multi-stage `Dockerfile`; final image is the single binary plus CA certs.
- `slack-app-manifest.yaml` shipped at the repo root.
- `docs/setup.md`: Google OAuth, Slack app, env vars, first admin, Docker run.

## Risks

- Generator drift: pin generator versions and enforce the CI diff check.
- Import partial failure: apply runs in one transaction; on error nothing is
  committed.
