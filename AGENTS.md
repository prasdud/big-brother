# AGENTS.md

Open-source, pluggable uptime monitoring and Slack alerting for tech teams.

## Principles
- DRY, KISS, minimal. Prefer 10 lines over 100.
- No abstraction until a second use exists.
- Stdlib first. Add a dependency only if it removes substantial code.
- Alert quality over check variety.
- One term per concept. No synonyms in code, schema, UI, or docs.

## Model
- One deployment = one workspace (the organization).
- Workspace contains projects. Project contains services.
- A service is one thing we monitor: an endpoint, microservice, or app. One service = one check config.
- Roles live at workspace level: admin, member, viewer.
- Every project-owned table carries `project_id`.

## Roles
| Action | viewer | member | admin |
|---|---|---|---|
| View everything | yes | yes | yes |
| Create/edit/pause/delete services, channels, alert settings | no | yes | yes |
| Create projects | no | no | yes |
| Add/remove users, set roles | no | no | yes |

## Alert Resolution (single resolver, reused for channel and template)
- Channel: service override, else project default. No workspace default.
- Template per trigger: service override, else project default, else built-in default.
- If nothing resolves, log a delivery failure. Never drop silently.

## Stack
- Backend: Go, chi, slog, sqlc, goose
- DB: SQLite (default, WAL, single writer), Postgres optional
- Frontend: React, Vite, TypeScript, TanStack Query/Router, Tailwind, shadcn/ui
- Embedded via `go:embed`, single binary
- Layout: `/cmd/server`, `/internal`, `/web`

## Functional Requirements
- Auth: Google sign-in (OIDC) restricted to allowed email domains; new sign-ins become viewer
- Admin can add users manually with a role
- First admin bootstrapped on first run
- Project CRUD (admin only)
- Services: HTTP, TCP, DNS (ping optional); interval, timeout, failure threshold
- States: up, down, pending, paused
- Slack via OAuth app: connect workspace, pick channel per project default or per service
- Triggers: down, recovered; editable message templates with variables
- Template variables: `{{service.name}}`, `{{service.url}}`, `{{project.name}}`, `{{status}}`, `{{duration}}`, `{{error}}`
- Test send
- Dedupe: one alert on down, one on recovery
- Check history, current status, uptime per service
- Alert delivery failure log
- Kuma import (see below)

## Kuma Import
- Input: Kuma JSON export (sample version 1.23.17)
- Kuma group monitor -> project
- Kuma monitor -> service (`name`, `type`, `url`, `hostname`, `port`, `interval`, `timeout`, `maxretries`, `parent`)
- Kuma notification (slack) -> channel; `notificationIDList` -> service channel override
- Unsupported types (e.g. telegram, mqtt, grpc) are skipped and listed in an import report
- Import is idempotent-safe: preview first, then apply
- Secrets in the file are stored encrypted and never logged
- History is not imported

## Non-Functional Requirements
- Single binary; SQLite default, Postgres optional
- Checks fire within +/-2s of schedule at 500 services on a small VPS (target, unverified; benchmark)
- Secrets (Slack tokens, imported credentials) encrypted at rest (AES-GCM, key from env)
- CSRF protection, session expiry
- Configurable retention (default 30 days), batched pruning, hourly rollups
- Structured logs, `/metrics`, `/healthz`
- Graceful shutdown, no lost state on restart
- Docker image; ship Slack app manifest; setup documented step by step
- OpenAPI spec; TS client generated from it
- Config as code: not in v1, but resources must be declarative and API-addressable by stable slug so YAML can be added without schema changes

## v1 Scope
- Login, project switcher, service list/create/edit/pause, service detail
- Slack connect, alert settings, template editor with preview
- Dashboard of current statuses
- Kuma import

## Out of v1
- Status pages, Slack buttons/slash commands, on-call, SMS/phone, multi-region probes, cert expiry and slow alerts, config-as-code UI

## Open Decisions
- Imported Kuma Slack webhooks vs Slack OAuth channel mapping
- Manual users: Google-only or password
- Setup time target with Slack and Google OAuth setup
