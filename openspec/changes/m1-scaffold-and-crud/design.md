# Design: M1 — Scaffold and Resource CRUD

## Repository layout

```
cmd/server/main.go        entrypoint: config, db, router, shutdown
internal/config           env parsing and defaults
internal/db               sqlite open (WAL), goose runner, sqlc store
internal/api              chi router, middleware, handlers, JSON helpers
internal/workspace        project + service domain logic
internal/slug             slug generation and uniqueness
migrations/               goose .sql migrations (embedded)
internal/store/sqlc       sqlc generated code + queries/*.sql
web/                      Vite React TS app
```

## Dependencies (justify each)

- `chi` — routing; stdlib is close but chi's middleware/param handling removes
  real code.
- `slog` — stdlib structured logging.
- `modernc.org/sqlite` — pure-Go SQLite, keeps the binary cgo-free.
- `github.com/pressly/goose/v3` — migrations; embedded `fs.FS` support.
- `sqlc` (build tool, not a runtime dep) — typed queries.
- Frontend per AGENTS.md stack.

## Configuration (env)

| Var | Default | Purpose |
|---|---|---|
| `BB_ADDR` | `:8080` | Listen address |
| `BB_DB_PATH` | `./data/big-brother.db` | SQLite file |
| `BB_LOG_LEVEL` | `info` | slog level |
| `BB_BASE_URL` | `http://localhost:8080` | Absolute links |
| `BB_RETENTION_DAYS` | `30` | Reserved for M2 |
| `BB_SECRET_KEY` | _(required in prod)_ | Reserved for M3/M4 encryption |

Secrets are read from env only and never logged.

## Schema (M1)

```sql
workspace(id TEXT PK, name TEXT, slug TEXT UNIQUE, created_at)
users(id TEXT PK, workspace_id TEXT, email TEXT UNIQUE, name TEXT,
      role TEXT CHECK(role IN ('admin','member','viewer')),
      google_sub TEXT, disabled INTEGER DEFAULT 0, created_at, last_login_at)
projects(id TEXT PK, workspace_id TEXT, name TEXT, slug TEXT UNIQUE, created_at)
services(id TEXT PK, project_id TEXT, name TEXT, slug TEXT,
         type TEXT CHECK(type IN ('http','tcp','dns')),
         url TEXT, hostname TEXT, port INTEGER,
         interval_seconds INTEGER, timeout_seconds INTEGER,
         failure_threshold INTEGER, enabled INTEGER DEFAULT 1,
         created_at, updated_at,
         UNIQUE(project_id, slug))
```

- IDs are UUIDv7 strings.
- `slug` is generated once from `name` and never changes on rename.
- sqlc queries are written to read/write these tables; `PRAGMA journal_mode=WAL`
  and `PRAGMA busy_timeout` set on open.

## API (v1)

- `GET/POST /api/v1/projects`, `GET/PATCH/DELETE /api/v1/projects/{project}`
- `GET/POST /api/v1/projects/{project}/services`
- `GET/PATCH/DELETE /api/v1/projects/{project}/services/{service}`
- `POST /api/v1/projects/{project}/services/{service}/pause` and `/resume`
- JSON errors: `{"error":{"code","message"}}`.

These paths are treated as stable for future config-as-code.

## Risks

- sqlc + goose schema drift: keep a single `schema.sql` as sqlc input and keep
  goose migrations additive.
- Slug collisions: uniqueness enforced by DB constraint and retried with a
  suffix.
