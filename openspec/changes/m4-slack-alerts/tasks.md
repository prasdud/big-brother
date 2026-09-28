# Tasks: M4 — Slack and Alerting

## 1. Secrets and schema
- [ ] 1.1 AES-256-GCM encrypt/decrypt helpers keyed from `BB_SECRET_KEY`
- [ ] 1.2 Migration for `slack_workspace`, `channels`, `alert_templates`,
      `delivery_failures`, and the `channel_id`/`default_channel_id` columns
- [ ] 1.3 Verify tokens round-trip and are never written to logs

## 2. Slack OAuth
- [ ] 2.1 `GET /slack/install` with state and redirect
- [ ] 2.2 `GET /slack/callback` validates state, exchanges code, stores token
- [ ] 2.3 `GET /slack/channels` lists channels
- [ ] 2.4 Verify connect with a real Slack app; token stored encrypted

## 3. Channels
- [ ] 3.1 Project default channel set/clear endpoints
- [ ] 3.2 Service channel override set/clear endpoints
- [ ] 3.3 Verify only member+ can change channel settings

## 4. Resolver
- [ ] 4.1 `resolveChannel` with service override then project default
- [ ] 4.2 `resolveTemplate` with service, project, then built-in
- [ ] 4.3 Unit tests covering every fallback and the no-resolution case
- [ ] 4.4 Miss logs a delivery failure instead of dropping

## 5. Templates
- [ ] 5.1 Built-in defaults for `down` and `recovered`
- [ ] 5.2 Render engine for the documented variables
- [ ] 5.3 Template preview endpoint with sample data
- [ ] 5.4 Test-send endpoint
- [ ] 5.5 Verify editing a template changes the next rendered alert

## 6. Alerter
- [ ] 6.1 Subscribe to M2 transition events
- [ ] 6.2 Send on `down` and on `recovered`
- [ ] 6.3 Post via `chat.postMessage`; retry on 429 with backoff
- [ ] 6.4 Record delivery failures and never drop silently
- [ ] 6.5 Verify exactly one message per down and per recovery

## 7. Verification
- [ ] 7.1 Take a real service down; observe one Slack alert and one recovery
- [ ] 7.2 Point a service at a channel that was deleted; observe a delivery
      failure entry
- [ ] 7.3 Run `go test ./internal/alert/...`
