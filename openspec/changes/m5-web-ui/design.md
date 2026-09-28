# Design: M5 — Web UI

## Structure

```
web/src/routes            TanStack Router route tree
web/src/features          auth, dashboard, projects, services, alerts
web/src/components/ui     shadcn/ui primitives
web/src/lib/api           typed API client, query keys
web/src/lib/session       current user + role context
```

## Routes

| Route | Purpose |
|---|---|
| `/login` | Google sign-in entry |
| `/` | Dashboard of current statuses |
| `/projects/:project` | Service list |
| `/projects/:project/services/new` | Create service |
| `/projects/:project/services/:service` | Service detail |
| `/projects/:project/settings` | Project settings and default channel |
| `/projects/:project/alerts` | Alert settings and template editor |
| `/settings/slack` | Slack connect and channels |
| `/settings/users` | Users and roles (admin) |
| `/settings/deliveries` | Delivery failures |

## Data layer

- TanStack Query with query keys scoped by workspace/project/service.
- Mutations invalidate the relevant keys and surface server errors as toasts.
- Polling for status views (short interval) until a push/SSE channel exists; a
  push channel is out of scope for v1.

## Role-aware UI

- A `useRole` hook exposes the current role.
- Buttons for actions above the user's role are hidden or disabled with a
  tooltip; the server still enforces access.

## Template editor

- Textarea with variable chips that insert `{{...}}`.
- Live preview calls the M4 preview endpoint with sample data.
- Test-send button selects a channel and reports the result.

## Serving

- `vite build` output is embedded in M1; in dev, Vite proxies `/api` to the Go
  server.
