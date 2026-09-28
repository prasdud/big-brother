# Delta for Auth

## ADDED Requirements

### Requirement: Google sign-in restricted by domain
The product SHALL authenticate users through Google OIDC and SHALL reject any
email whose domain is not on the configured allowlist.

#### Scenario: Allowed domain signs in
- GIVEN `example.com` is an allowed domain
- WHEN a user with `user@example.com` completes sign-in
- THEN a session is created and the user is signed in

#### Scenario: Disallowed domain is rejected
- GIVEN `other.com` is not allowed
- WHEN a user with `user@other.com` completes sign-in
- THEN sign-in is rejected and no session is created

### Requirement: New users default to viewer
A user signing in for the first time SHALL be created with the `viewer` role,
unless they are the bootstrapped admin.

#### Scenario: First sign-in
- GIVEN an allowed user with no account
- WHEN they sign in
- THEN an account is created with role `viewer`

### Requirement: First admin bootstrap
On first run the product SHALL bootstrap exactly one admin from the configured
bootstrap email once that user signs in.

#### Scenario: Bootstrap email becomes admin
- GIVEN no admin exists and `BB_BOOTSTRAP_ADMIN_EMAIL` is set
- WHEN that user signs in
- THEN their role is `admin`

#### Scenario: Only one bootstrap
- GIVEN an admin already exists
- WHEN the bootstrap user signs in again
- THEN no additional bootstrap occurs

### Requirement: Sessions expire
The product SHALL issue sessions with a configurable expiry and SHALL reject
expired sessions.

#### Scenario: Expired session
- GIVEN a session past its expiry
- WHEN a protected request is made with it
- THEN the request is rejected as unauthenticated

### Requirement: CSRF protection
The product SHALL require a valid CSRF token on state-changing requests.

#### Scenario: Missing CSRF token
- GIVEN an authenticated session
- WHEN a POST request omits the CSRF token
- THEN the request is rejected

### Requirement: Role-based access
The product SHALL enforce workspace roles on every endpoint: viewers may only
read, members may manage services, channels, and alert settings, and only
admins may create projects or manage users and roles.

#### Scenario: Viewer cannot mutate
- GIVEN a signed-in viewer
- WHEN they attempt to create a service
- THEN the request is rejected with a forbidden error

#### Scenario: Member cannot create projects
- GIVEN a signed-in member
- WHEN they attempt to create a project
- THEN the request is rejected with a forbidden error

#### Scenario: Admin manages users
- GIVEN a signed-in admin
- WHEN they add a user with role `member`
- THEN the user exists with role `member`

### Requirement: Manual user creation
An admin SHALL be able to add a user manually and assign any role.

#### Scenario: Add a manual user
- GIVEN a signed-in admin
- WHEN they add `ops@example.com` as `admin`
- THEN `ops@example.com` is listed with role `admin`
