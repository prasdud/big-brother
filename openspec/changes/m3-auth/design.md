# Design: M3 — Auth and Roles

## Components

```
internal/auth       oidc.go, session.go, csrf.go, middleware.go, roles.go
internal/user       user creation, invitations, role changes
```

## Configuration (env)

| Var | Purpose |
|---|---|
| `BB_GOOGLE_CLIENT_ID` | OIDC client id |
| `BB_GOOGLE_CLIENT_SECRET` | OIDC client secret |
| `BB_ALLOWED_EMAIL_DOMAINS` | Comma-separated allowlist |
| `BB_BOOTSTRAP_ADMIN_EMAIL` | First admin, used only when no admin exists |
| `BB_SESSION_TTL` | Session lifetime, default `168h` |

## OIDC flow

1. `GET /auth/login` sets an anti-forgery `state` cookie and redirects to Google.
2. `GET /auth/callback` validates `state`, exchanges the code, and verifies the
   ID token (`iss`, `aud`, `exp`, `nonce`).
3. The email domain must be in `BB_ALLOWED_EMAIL_DOMAINS`, else sign-in is
   rejected and logged without the token.
4. First successful sign-in creates a `viewer` unless the email matches
   `BB_BOOTSTRAP_ADMIN_EMAIL` and no admin exists yet.
5. `POST /auth/logout` deletes the session and clears the cookie.

## Sessions and CSRF

- `sessions(id, user_id, token_hash, csrf_token, expires_at, created_at)`.
- The cookie holds an opaque random token; only its hash is stored.
- Expired sessions are rejected and pruned periodically.
- State-changing requests require the `X-CSRF-Token` header to match the
  session's CSRF token.

## Roles and guards

- `roleFor(user)` returns the workspace role.
- Route guard matrix, matching AGENTS.md:

| Action | viewer | member | admin |
|---|---|---|---|
| View everything | yes | yes | yes |
| Create/edit/pause/delete services, channels, alert settings | no | yes | yes |
| Create projects | no | no | yes |
| Add/remove users, set roles | no | no | yes |

- Unknown roles deny by default.

## Bootstrap admin

- On startup, if `users` has no admin, the first sign-in matching
  `BB_BOOTSTRAP_ADMIN_EMAIL` becomes admin. If that email is unset, an operator
  can be promoted via a documented one-shot command.

## Schema additions

```sql
sessions(id TEXT PK, user_id TEXT, token_hash TEXT UNIQUE, csrf_token TEXT,
         expires_at, created_at)
-- users gains: google_sub, last_login_at, disabled (already in M1)
```

## Risks

- OIDC misconfiguration: validate issuer and audience strictly; fail closed.
- Session fixation: issue a new session id and token after sign-in.
- CSRF gaps on GET: never mutate state on GET.
