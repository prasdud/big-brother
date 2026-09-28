# Delta for Monitoring

## ADDED Requirements

### Requirement: Scheduled HTTP checks
The product SHALL check each enabled HTTP service on its configured interval
and SHALL treat an HTTP response with a 2xx or 3xx status received within the
service timeout as up.

#### Scenario: Successful check
- GIVEN an enabled HTTP service pointing at a healthy URL
- WHEN its interval elapses
- THEN a check is recorded with status up, the response status code, and latency

#### Scenario: Timeout
- GIVEN a service with `timeout_seconds` of 5
- WHEN the target does not respond within 5 seconds
- THEN a check is recorded as down with a timeout error

#### Scenario: Non-success status
- GIVEN a service whose target returns 500
- WHEN the check runs
- THEN the check is recorded as down with the status code

### Requirement: Service states
The product SHALL report exactly one current state per service: `up`, `down`,
`pending`, or `paused`.

#### Scenario: First failure enters pending
- GIVEN an up service with failure threshold 3
- WHEN the first check fails
- THEN its state becomes `pending`

#### Scenario: Threshold reached confirms down
- GIVEN a `pending` service with failure threshold 3
- WHEN consecutive failures reach 3
- THEN its state becomes `down`

#### Scenario: Recovery from down
- GIVEN a `down` service
- WHEN a check succeeds
- THEN its state becomes `up`

#### Scenario: Paused services are not checked
- GIVEN a paused service
- WHEN its interval would have elapsed
- THEN no check runs and its state remains `paused`

### Requirement: One transition event per change
The product SHALL emit exactly one state-transition event each time a service
changes state, so downstream consumers do not double-notify.

#### Scenario: No event without a change
- GIVEN a `down` service
- WHEN another check fails while it remains down
- THEN no transition event is emitted

### Requirement: Check history
The product SHALL retain a queryable history of check results per service,
including timestamp, status, status code, latency, and error.

#### Scenario: Query recent checks
- GIVEN a service that has been checked many times
- WHEN its checks are requested with a time range and limit
- THEN the matching results are returned newest first

### Requirement: Uptime
The product SHALL report uptime per service over 24-hour, 7-day, and 30-day
windows from hourly rollups.

#### Scenario: Uptime percentage
- GIVEN 100 checks in a 24-hour window of which 99 are up
- WHEN uptime is requested for 24h
- THEN the reported value is 99 percent

### Requirement: Configurable retention
The product SHALL prune raw check history older than the configured retention
period and SHALL preserve uptime rollups beyond it.

#### Scenario: Prune old checks
- GIVEN `BB_RETENTION_DAYS=30` and checks older than 30 days
- WHEN the pruning job runs
- THEN raw checks older than 30 days are deleted
- AND uptime rollups for the same period remain
