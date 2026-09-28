# Tasks: M1 — Scaffold and Resource CRUD

## 1. Toolchain and module
- [ ] 1.1 Install Go 1.24+, `sqlc`, `goose`
- [ ] 1.2 `go mod init github.com/prasdud/big-brother` and commit `go.mod`
- [ ] 1.3 Verify `go build ./...` on a hello-world `cmd/server`

## 2. Configuration and logging
- [ ] 2.1 `internal/config` parses env with defaults from design.md
- [ ] 2.2 `slog` JSON handler configured from `BB_LOG_LEVEL`
- [ ] 2.3 Verify startup logs one structured line with addr and db path

## 3. Database
- [ ] 3.1 Add `modernc.org/sqlite`; open with WAL + busy_timeout
- [ ] 3.2 Embed `migrations/` and run goose on boot
- [ ] 3.3 Write `schema.sql` consumed by sqlc
- [ ] 3.4 Verify tables exist after first run; restart is idempotent

## 4. Router and platform endpoints
- [ ] 4.1 chi router with `/api/v1` group and JSON error helper
- [ ] 4.2 `GET /healthz` returns 200 when DB ping succeeds
- [ ] 4.3 `GET /metrics` exposes basic counters
- [ ] 4.4 Graceful shutdown on SIGINT/SIGTERM
- [ ] 4.5 Verify with `curl localhost:8080/healthz`

## 5. Workspace, projects, services
- [ ] 5.1 goose migration for `workspace`, `users`, `projects`, `services`
- [ ] 5.2 `internal/slug` with uniqueness retry
- [ ] 5.3 sqlc queries and handlers for project CRUD
- [ ] 5.4 sqlc queries and handlers for service CRUD + pause/resume
- [ ] 5.5 Create the singleton workspace row on first boot
- [ ] 5.6 Verify create project → create service → restart → both persist

## 6. Embedded web app
- [ ] 6.1 Scaffold `web/` with Vite + React + TS + Tailwind + TanStack
- [ ] 6.2 `go:embed` the `web/dist` build with SPA fallback to `index.html`
- [ ] 6.3 Dev proxy from Vite to `BB_ADDR`
- [ ] 6.4 Verify production build serves the SPA and API from one port

## 7. Tooling
- [ ] 7.1 `Makefile`: `dev`, `build`, `test`, `lint`, `sqlc`
- [ ] 7.2 `Dockerfile` multi-stage build producing the single binary
- [ ] 7.3 `.env.example` and update `.gitignore` for `data/` and `web/dist`
- [ ] 7.4 Run `go test ./...`, `go vet ./...`, and `golangci-lint run`
