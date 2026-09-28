# Delta for Platform

## ADDED Requirements

### Requirement: Published API contract
The product SHALL publish an OpenAPI document describing every `/api/v1`
endpoint and SHALL generate the web application's TypeScript client from it.

#### Scenario: Contract is available
- GIVEN a running server
- WHEN `/api/v1/openapi.json` is requested
- THEN the document lists all v1 endpoints

#### Scenario: Client is current
- GIVEN the OpenAPI document has changed
- WHEN continuous integration regenerates the client
- THEN it produces no diff

### Requirement: TCP and DNS checks
The product SHALL support TCP and DNS check types in addition to HTTP, with the
same scheduling, state, and history behavior.

#### Scenario: TCP check
- GIVEN a TCP service with host and port
- WHEN a check runs
- THEN it is up when the connection completes within the timeout

#### Scenario: DNS check
- GIVEN a DNS service with a hostname
- WHEN a check runs
- THEN it is up when the hostname resolves within the timeout

### Requirement: Operational metrics
The product SHALL expose check, state-transition, alert, and delivery-failure
metrics on `/metrics`.

#### Scenario: Counters after activity
- GIVEN a service that went down and recovered with an alert sent
- WHEN `/metrics` is scraped
- THEN check, transition, and delivery counters reflect that activity

### Requirement: Container and setup documentation
The product SHALL ship a Docker image, a Slack app manifest, and step-by-step
setup documentation covering Google OAuth, Slack, configuration, first admin,
and running the container.

#### Scenario: Docker deployment
- GIVEN the published image and required environment variables
- WHEN the container is started
- THEN the application is reachable and usable

#### Scenario: Setup from docs
- GIVEN a fresh operator following the setup guide
- WHEN they complete the documented steps
- THEN they can sign in and receive a Slack alert
