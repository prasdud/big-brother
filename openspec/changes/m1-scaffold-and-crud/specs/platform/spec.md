# Delta for Platform

## ADDED Requirements

### Requirement: Single binary distribution
The product SHALL ship and run as a single binary that serves both the API and
the web application.

#### Scenario: Serve API and UI from one process
- GIVEN a built binary and no web server running separately
- WHEN the binary starts
- THEN the API is reachable under `/api/v1`
- AND the web application loads from the same origin

### Requirement: Configuration from environment
The product SHALL read all runtime configuration from environment variables
with documented defaults, and SHALL NOT log secret values.

#### Scenario: Start with defaults
- GIVEN no environment variables set
- WHEN the binary starts
- THEN it listens on `:8080` and uses `./data/big-brother.db`

#### Scenario: Secret values are not logged
- GIVEN `BB_SECRET_KEY` is set
- WHEN the binary starts
- THEN startup logs contain no value of `BB_SECRET_KEY`

### Requirement: Structured logs
The product SHALL emit structured JSON logs and SHALL honor a configurable log
level.

#### Scenario: Log level filters output
- GIVEN `BB_LOG_LEVEL=warn`
- WHEN a debug-level event occurs
- THEN no debug log line is emitted

### Requirement: Health and metrics endpoints
The product SHALL expose `/healthz` reflecting database connectivity and
`/metrics` with basic process and HTTP counters.

#### Scenario: Healthy database
- GIVEN the database is reachable
- WHEN `GET /healthz` is called
- THEN the response is 200

#### Scenario: Unreachable database
- GIVEN the database cannot be pinged
- WHEN `GET /healthz` is called
- THEN the response is a non-2xx status

### Requirement: Graceful shutdown
The product SHALL stop accepting new work and shut down cleanly on SIGINT or
SIGTERM.

#### Scenario: Signal during operation
- GIVEN the server is running
- WHEN SIGTERM is received
- THEN in-flight requests complete and the process exits 0

### Requirement: Durable SQLite persistence
The product SHALL persist all state in SQLite using WAL mode and SHALL apply
embedded migrations automatically on startup.

#### Scenario: Restart preserves state
- GIVEN a project and service were created
- WHEN the binary restarts
- THEN both still exist

#### Scenario: Migrations are idempotent
- GIVEN migrations already applied
- WHEN the binary restarts
- THEN no migration is applied twice and startup succeeds

### Requirement: Stable resource slugs
Every API-addressable resource SHALL have a slug that is unique within its
parent, generated once, and immutable across renames.

#### Scenario: Slug survives rename
- GIVEN a project with slug `payments`
- WHEN the project name changes
- THEN its slug remains `payments`
- AND it is still reachable at its existing API path
