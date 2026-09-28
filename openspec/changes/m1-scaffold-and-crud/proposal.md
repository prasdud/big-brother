# Proposal: M1 — Scaffold and Resource CRUD

## Intent

Stand up the single-binary foundation and the core data model so every later
milestone has a running place to live, a stable API surface, and a working
build/run loop. After M1 an admin can create a project and a member can create
a service, but nothing is checked yet and no screen exists.

## Scope

- Project scaffolding: Go module, `chi` router, `slog`, SQLite (WAL), `sqlc`,
  `goose`, and an embedded Vite web app served by the same binary.
- Config from environment with sane defaults; `/healthz`, `/metrics`, graceful
  shutdown.
- Data model and CRUD for `workspace`, `projects`, and `services`.
- Stable slugs and a `project_id` on every project-owned table.
- Role and user fields exist on the model, but enforcement is deferred to M3.

## Approach

- One Go module at `github.com/prasdud/big-brother` with `/cmd/server`,
  `/internal`, `/web`, `/migrations`.
- SQLite via a pure-Go driver so the binary stays cgo-free.
- `goose` migrations embedded and applied on boot; `sqlc` generates typed
  queries from `schema.sql`.
- The built Vite app is embedded with `go:embed` and served by chi with an SPA
  fallback; the API lives under `/api/v1`.
- Resources are addressed by slug: `/api/v1/projects/{project}`, and nested
  services under `/api/v1/projects/{project}/services/{service}`.

## Out of scope

- Running checks, auth enforcement, Slack, UI screens, import.

## Success signals

- `make dev` serves API and UI from one process.
- `POST /api/v1/projects` then `POST /api/v1/projects/{project}/services`
  succeeds and persists across a restart.
- `/healthz` and `/metrics` respond.
