# Proposal: M4 — Slack and Alerting

## Intent

Deliver the core value: when a service goes down, a Slack alert is posted, and
when it recovers, exactly one recovery alert is posted. Channels and message
templates resolve through a single resolver with service overrides.

## Scope

- Slack OAuth app install: connect a workspace, store tokens encrypted.
- Channels: pick a Slack channel per project default and optionally per service.
- Single resolver for channel and template, reused everywhere.
- Triggers `down` and `recovered`, driven by M2 state transitions.
- Editable message templates with variables and a preview.
- Test send.
- Delivery-failure log; never drop silently.

## Approach

- Slack OAuth v2 install flow stores the bot token and workspace id.
- Channel resolution: service override, else project default. No workspace
  default. If nothing resolves, log a delivery failure.
- Template resolution per trigger: service override, else project default, else
  built-in default.
- On each M2 transition event, the alerter resolves channel and template,
  renders the message, posts it, and logs the outcome.
- Dedupe is inherited from M2's one-event-per-transition guarantee.

## Out of scope

- Slack interactive buttons and slash commands, per-trigger channel overrides
  beyond down/recovered, digest and reminder alerts.

## Success signals

- A service going down posts exactly one Slack message; recovery posts one.
- Editing the template changes the next alert.
- A misconfigured channel produces a visible delivery-failure entry.
