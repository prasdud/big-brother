# Proposal: M6 — Operations, OpenAPI, Checks, and Import

## Intent

Finish v1: publish the API contract, broaden checks to TCP and DNS, ship the
operational packaging, and import an existing Kuma deployment so teams can
adopt without rebuilding their monitor list.

## Scope

- OpenAPI spec for every `/api/v1` endpoint; generated TypeScript client used by
  the web app.
- TCP and DNS check types alongside HTTP.
- Docker image, Slack app manifest, and step-by-step setup docs.
- `/metrics` covering checks, alerts, and deliveries.
- Kuma JSON import: preview then apply, including projects, services, Slack
  channels, and channel overrides.

## Approach

- The OpenAPI document is authored as the contract; the TS client is generated
  and replaces M5's hand-written client. A CI check fails if the generated
  client or spec is stale.
- TCP and DNS implement the M2 `Checker` interface, so the scheduler and state
  machine are unchanged.
- Import is two-phase: parse and produce a report, then apply within a
  transaction. Secrets in the file are encrypted and never logged.

## Out of scope

- Postgres, multi-region probes, cert-expiry and slow alerts, status pages,
  config-as-code UI.

## Success signals

- Every endpoint appears in the served OpenAPI document.
- TCP and DNS services are monitored like HTTP.
- A Kuma export can be previewed and applied; unsupported types are reported.
