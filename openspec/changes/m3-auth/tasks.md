# Tasks: M3 — Auth and Roles

## 1. Configuration and schema
- [ ] 1.1 Add auth env vars with validation and fail-closed defaults
- [ ] 1.2 Migration for `sessions`; confirm `users` role/disabled columns
- [ ] 1.3 Verify startup fails clearly when required OIDC vars are missing

## 2. OIDC
- [ ] 2.1 `GET /auth/login` with state cookie and redirect
- [ ] 2.2 `GET /auth/callback` code exchange and ID token validation
- [ ] 2.3 Enforce `BB_ALLOWED_EMAIL_DOMAINS`
- [ ] 2.4 Create or update the user on first sign-in
- [ ] 2.5 Verify a disallowed domain is rejected

## 3. Sessions and CSRF
- [ ] 3.1 Create session with hashed token and rotate on sign-in
- [ ] 3.2 Session cookie flags: HttpOnly, Secure, SameSite=Lax
- [ ] 3.3 Reject and prune expired sessions
- [ ] 3.4 Double-submit CSRF check on mutating requests
- [ ] 3.5 `POST /auth/logout` invalidates the session
- [ ] 3.6 Verify replaying an expired session is rejected

## 4. Roles
- [ ] 4.1 `roleFor` and request-context user
- [ ] 4.2 Middleware requiring authentication on `/api/v1`
- [ ] 4.3 Guards per the role matrix in design.md
- [ ] 4.4 Unit tests for each role/action cell
- [ ] 4.5 Default-deny on unknown role

## 5. User management
- [ ] 5.1 Bootstrap admin on first run via `BB_BOOTSTRAP_ADMIN_EMAIL`
- [ ] 5.2 Admin endpoints: list, add, disable, set role
- [ ] 5.3 Manual user add assigns the chosen role
- [ ] 5.4 Verify a viewer gets 403 creating a service and a member succeeds
- [ ] 5.5 Verify only an admin can create a project or change a role

## 6. Verification
- [ ] 6.1 Run `go test ./internal/auth/...`
- [ ] 6.2 Manual pass through login → create service → logout
