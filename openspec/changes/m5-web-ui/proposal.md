# Proposal: M5 — Web UI

## Intent

Give the product a usable interface. The React app, already embedded in M1,
grows the screens needed to operate the product end to end: sign in, switch
projects, watch status, manage services, and configure Slack alerts.

## Scope

- Sign-in screen and session handling.
- Dashboard of current statuses across the selected project.
- Project switcher; admin project create/edit/delete.
- Service list, create/edit, pause/resume, delete.
- Service detail: current status, recent checks, uptime.
- Slack connect, channel assignment, alert settings.
- Template editor with live preview.
- Delivery-failure view.

## Approach

- TanStack Router for routes and TanStack Query for data, with a generated TS
  client (from M6's OpenAPI; until then, a typed hand-written client).
- shadcn/ui components with Tailwind for a consistent, minimal interface.
- The app is served by the Go binary; API calls go to `/api/v1` on the same
  origin, so no CORS configuration is needed.
- Role-aware rendering: hide or disable actions the user cannot perform, while
  the server remains the authority.

## Out of scope

- Status pages, on-call views, mobile apps, theming beyond defaults.

## Success signals

- A member can create a service and watch it go up or down on the dashboard.
- An admin can connect Slack and set a project default channel without the API.
- Every M1–M4 capability is reachable from the UI.
