# Implementation Order

Each milestone is an OpenSpec change under `openspec/changes/`. A change is
implemented, archived, and its specs merged before the next one starts. Do not
start a milestone whose dependencies are unarchived.

| Change | Depends on | Delivers |
|---|---|---|
| `m1-scaffold-and-crud` | — | Single-binary scaffold, SQLite/migrations, workspace + project + service CRUD |
| `m2-check-engine` | M1 | HTTP checker, scheduler, up/down/pending/paused states, history, uptime, retention |
| `m3-auth` | M1 | Google OIDC, sessions, CSRF, workspace roles, bootstrap admin, manual users |
| `m4-slack-alerts` | M2, M3 | Slack OAuth, channels, alert resolution, templates, test send, delivery log |
| `m5-web-ui` | M1–M4 | Login, dashboard, project switcher, service CRUD/detail, alert settings, template editor |
| `m6-ops-and-import` | M1–M5 | OpenAPI + TS client, metrics/Docker/docs, TCP/DNS checks, Kuma import |

## Cross-cutting rules

- One term per concept. The canonical terms are `workspace`, `project`,
  `service`, `check`, `channel`, `alert`, `template`, `trigger`.
- Every project-owned table carries `project_id`.
- Every API-addressable resource has an immutable, unique `slug` per parent.
- SQLite is the only database. Do not add a database dialect switch.

## Open questions carried forward

- Kuma import: map imported Kuma notification names onto an existing Slack
  OAuth channel, or store raw webhook URLs. Resolve in M6 before implementation.
- Manual users: password login or Google-only. M3 assumes Google-only plus
  admin-created invitations.
