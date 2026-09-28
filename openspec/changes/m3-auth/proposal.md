# Proposal: M3 — Auth and Roles

## Intent

Put a real identity and authorization layer in front of the API. Users sign in
with Google restricted to allowed email domains, a first admin is bootstrapped
on first run, and every mutating endpoint enforces the workspace role model.

## Scope

- Google OIDC sign-in restricted to allowed email domains.
- Sessions with expiry and CSRF protection.
- Workspace roles: `admin`, `member`, `viewer`, enforced on all routes.
- First admin bootstrapped on first run.
- Admin-created users with an assigned role (invitation flow).
- New sign-ins default to `viewer`.

## Approach

- OIDC via stdlib `net/http` plus a small OIDC client; no heavy framework.
- Server-side session records in SQLite with an opaque session cookie
  (`HttpOnly`, `Secure`, `SameSite=Lax`).
- Double-submit CSRF token for state-changing requests.
- A chi middleware resolves the session to a user and a role, then a
  route-level guard checks the required role.

## Out of scope

- Password login, SSO beyond Google, per-project roles, SCIM.

## Success signals

- Unauthenticated requests to protected routes are rejected.
- A disallowed email domain cannot sign in.
- A viewer cannot create a service; a member can; only an admin can create a
  project or manage users.
