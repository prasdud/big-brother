# Tasks: M5 — Web UI

## 1. Foundation
- [ ] 1.1 TanStack Router route tree and layout shell
- [ ] 1.2 Tailwind + shadcn/ui primitives installed
- [ ] 1.3 Typed API client and query-key conventions
- [ ] 1.4 `useRole` and session context
- [ ] 1.5 Verify an unauthenticated visit redirects to `/login`

## 2. Auth and navigation
- [ ] 2.1 Login screen launching the Google flow
- [ ] 2.2 Logout affordance
- [ ] 2.3 Project switcher in the header
- [ ] 2.4 Verify switching projects reloads scoped data

## 3. Dashboard
- [ ] 3.1 Current status table for the selected project
- [ ] 3.2 State badges: up, down, pending, paused
- [ ] 3.3 Verify it reflects a live state change

## 4. Services
- [ ] 4.1 Service list with pause/resume and delete
- [ ] 4.2 Create/edit service form with validation
- [ ] 4.3 Service detail: status, recent checks, uptime
- [ ] 4.4 Verify creating a service from the UI appears in the list
- [ ] 4.5 Verify viewers see no mutating controls

## 5. Projects
- [ ] 5.1 Project create/edit/delete (admin only)
- [ ] 5.2 Project settings: default channel selection
- [ ] 5.3 Verify a member sees no project-create control

## 6. Alerts
- [ ] 6.1 Slack connect flow screen
- [ ] 6.2 Channel picker for project default and per-service override
- [ ] 6.3 Template editor with variable insertion and live preview
- [ ] 6.4 Test-send control with result feedback
- [ ] 6.5 Delivery-failure view
- [ ] 6.6 Verify editing a template previews before saving

## 7. Users
- [ ] 7.1 Users and roles screen with add/disable/role change (admin)
- [ ] 7.2 Verify a non-admin cannot reach the users screen

## 8. Build and verification
- [ ] 8.1 `vite build` succeeds and is embedded
- [ ] 8.2 End-to-end pass: login → project → service → down → Slack alert
- [ ] 8.3 Run `tsc --noEmit` and the frontend test suite
