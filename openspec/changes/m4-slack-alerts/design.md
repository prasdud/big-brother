# Design: M4 — Slack and Alerting

## Components

```
internal/slack      oauth.go, client.go, channels.go
internal/alert      resolver.go, template.go, alerter.go, delivery.go
```

## Configuration (env)

| Var | Purpose |
|---|---|
| `BB_SLACK_CLIENT_ID` | Slack app client id |
| `BB_SLACK_CLIENT_SECRET` | Slack app client secret |
| `BB_SLACK_REDIRECT_URL` | OAuth redirect URL |
| `BB_SECRET_KEY` | AES-GCM key for token encryption (required here) |

## Encryption

- Bot tokens are encrypted at rest with AES-256-GCM using a key derived from
  `BB_SECRET_KEY`.
- Tokens are decrypted only in memory at send time and never logged.

## Slack OAuth

1. `GET /api/v1/slack/install` redirects to Slack with `state`.
2. `GET /api/v1/slack/callback` validates `state`, exchanges the code, stores
   the encrypted bot token and team id on the workspace.
3. `GET /api/v1/slack/channels` lists channels for the connected workspace.

## Schema additions

```sql
slack_workspace(workspace_id TEXT PK, team_id TEXT, team_name TEXT,
                bot_token_enc BLOB, installed_at)
channels(id TEXT PK, project_id TEXT, name TEXT, slack_channel_id TEXT,
         created_at)
-- projects gains: default_channel_id TEXT NULL
-- services gains: channel_id TEXT NULL
alert_templates(id TEXT PK, project_id TEXT, trigger TEXT, body TEXT,
                created_at, updated_at, UNIQUE(project_id, trigger))
-- services gains: template_down TEXT NULL, template_recovered TEXT NULL
delivery_failures(id TEXT PK, project_id TEXT, service_id TEXT, trigger TEXT,
                  reason TEXT, created_at)
```

## Single resolver

```
resolveChannel(service)  = service.channel_id  ?? project.default_channel_id ?? none
resolveTemplate(service, trigger) =
    service.template_{trigger} ?? project.alert_templates[trigger] ?? builtin[trigger]
```

- Reused for send, preview, and test send so behavior cannot diverge.
- No workspace default for channels; a miss is a logged delivery failure.
- Resolution for a service is covered by tests, including every fallback.

## Alerter

- Consumes M2 transition events from an in-process channel.
- Dedupe: one alert per transition, enforced by M2; the alerter also records the
  last alerted state per service as a guard.
- Renders the template, posts via `chat.postMessage`, and logs success or
  failure. Failures are written to `delivery_failures` and logged.

## Templates

- Variables: `{{service.name}}`, `{{service.url}}`, `{{project.name}}`,
  `{{status}}`, `{{duration}}`, `{{error}}`.
- `{{duration}}` is the time since the previous state change.
- Built-in defaults exist for `down` and `recovered`.
- `POST /api/v1/projects/{project}/alert-templates/preview` renders with sample
  data; `POST /api/v1/projects/{project}/alert-templates/test-send` posts a test.

## Risks

- Slack rate limits: retry with backoff on 429 and record persistent failures.
- Token rotation: support re-install to overwrite stored credentials.
