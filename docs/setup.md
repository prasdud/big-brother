# Setup

big-brother ships as a single binary that serves the API and the web UI, and
stores everything in a local SQLite file.

## 1. Run the container

```bash
docker run -d --name big-brother \
  -p 8080:8080 \
  -v big-brother-data:/data \
  -e BB_BASE_URL=https://monitor.example.com \
  -e BB_SECRET_KEY="$(openssl rand -hex 32)" \
  -e BB_GOOGLE_CLIENT_ID=... \
  -e BB_GOOGLE_CLIENT_SECRET=... \
  -e BB_ALLOWED_EMAIL_DOMAINS=example.com \
  -e BB_BOOTSTRAP_ADMIN_EMAIL=you@example.com \
  -e BB_SLACK_CLIENT_ID=... \
  -e BB_SLACK_CLIENT_SECRET=... \
  ghcr.io/prasdud/big-brother:latest
```

The database lives at `/data/big-brother.db`. Keep `BB_SECRET_KEY` stable: it
encrypts Slack tokens at rest, and rotating it makes existing tokens
unreadable.

Prefer building locally? `make build` produces `bin/big-brother`, then run it
with the same environment variables.

## 2. Google sign-in (OIDC)

1. In Google Cloud Console, create an OAuth 2.0 Client ID of type **Web
   application**.
2. Add an authorized redirect URI: `https://monitor.example.com/auth/callback`.
3. Copy the client ID and secret into `BB_GOOGLE_CLIENT_ID` and
   `BB_GOOGLE_CLIENT_SECRET`.
4. Set `BB_ALLOWED_EMAIL_DOMAINS` to the domains allowed to sign in
   (comma-separated). Sign-ins from other domains are rejected.
5. Set `BB_BOOTSTRAP_ADMIN_EMAIL` to the first administrator's email. The first
   time that user signs in they become an admin; everyone else starts as a
   viewer. Admins can change roles and add users under **Users**.

If `BB_GOOGLE_CLIENT_ID` is unset, authentication is disabled. Use that only
for local development: the API and UI run as an implicit admin.

## 3. Slack app

1. Create an app from the manifest in `slack-app-manifest.yaml`
   (https://api.slack.com/apps → **Create New App** → **From an app manifest**).
2. Replace the redirect URL in the manifest with
   `https://monitor.example.com/api/v1/slack/callback`, or edit it under
   **OAuth & Permissions** after creation.
3. Install the app to your workspace and copy the client ID and secret into
   `BB_SLACK_CLIENT_ID` and `BB_SLACK_CLIENT_SECRET`.
4. In big-brother, open **Slack** and click **Connect Slack** as an admin to
   complete the OAuth install. The bot token is stored encrypted with
   `BB_SECRET_KEY` and never returned by the API or written to logs.
5. Add the bot to the channels it should post to, then choose a project default
   channel (**Settings**) and optional per-service overrides.

## 4. Configuration reference

| Variable | Default | Purpose |
|---|---|---|
| `BB_ADDR` | `:8080` | Listen address |
| `BB_DB_PATH` | `./data/big-brother.db` | SQLite file |
| `BB_BASE_URL` | `http://localhost:8080` | Absolute links and OAuth defaults |
| `BB_LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `BB_WORKSPACE_NAME` | `Default` | Used only when the database is first created |
| `BB_RETENTION_DAYS` | `30` | Raw check retention |
| `BB_CHECK_WORKERS` | `16` | Concurrent check workers |
| `BB_SECRET_KEY` | — | Encrypts Slack tokens (required with Slack) |
| `BB_GOOGLE_CLIENT_ID` / `_SECRET` | — | Google OIDC |
| `BB_ALLOWED_EMAIL_DOMAINS` | — | Comma-separated allowlist |
| `BB_BOOTSTRAP_ADMIN_EMAIL` | — | First admin |
| `BB_COOKIE_SECURE` | `true` | Set `false` only for local HTTP |
| `BB_SESSION_TTL` | `168h` | Session lifetime |
| `BB_SLACK_CLIENT_ID` / `_SECRET` | — | Slack OAuth |

Startup fails closed: setting `BB_GOOGLE_CLIENT_ID` requires the secret and
allowed domains, and setting `BB_SLACK_CLIENT_ID` requires its secret and
`BB_SECRET_KEY`.

## 5. First run

1. Start the service with the environment above.
2. Open `https://monitor.example.com/` and sign in with Google.
3. As the bootstrap admin, create a project (**Dashboard → Create project**).
4. Add a service (HTTP, TCP, or DNS) and set its interval, timeout, and failure
   threshold.
5. Connect a Slack channel and send a **Test send** from **Alerts** to confirm
   delivery.

## 6. Health and metrics

- `GET /healthz` returns 200 when the database is reachable.
- `GET /metrics` exposes check, transition, and alert delivery counters.
- `GET /api/v1/openapi.json` serves the API contract.

## 7. Import from Kuma

Export your monitors from Uptime Kuma as JSON. As an admin, POST the file to
preview, then apply:

```bash
curl -X POST https://monitor.example.com/api/v1/import/kuma/preview \
  -H 'Content-Type: application/json' \
  --data-binary @kuma-export.json

curl -X POST https://monitor.example.com/api/v1/import/kuma/apply \
  -H 'Content-Type: application/json' \
  --data-binary @kuma-export.json
```

Group monitors become projects and monitors become services (HTTP, TCP, and DNS;
other types are skipped and listed in the report). Import is idempotent: apply
the same export twice and existing resources are matched by slug. Slack
notifications are not imported — connect Slack and choose channels in the app.
