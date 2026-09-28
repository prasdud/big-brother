# Delta for Import

## ADDED Requirements

### Requirement: Import preview
The product SHALL parse a Kuma JSON export and return a report of what would be
created, skipped, and warned about, without modifying any state.

#### Scenario: Preview does not write
- GIVEN a valid Kuma export
- WHEN a preview is requested
- THEN a report is returned
- AND no project, service, or channel is created

#### Scenario: Report lists skipped monitors
- GIVEN an export containing a `grpc` monitor
- WHEN a preview is requested
- THEN the report lists that monitor as skipped with a reason

### Requirement: Import apply
The product SHALL apply a previewed import atomically so that a failure leaves
no partial state.

#### Scenario: Successful apply
- GIVEN a valid Kuma export
- WHEN apply is requested
- THEN projects, services, channels, and overrides are created as reported

#### Scenario: Failure rolls back
- GIVEN an export that fails midway through apply
- WHEN apply is requested
- THEN no partial resources are committed

### Requirement: Kuma mapping
The product SHALL map Kuma group monitors to projects, Kuma monitors to
services, Slack notifications to channels, and `notificationIDList` entries to
service channel overrides.

#### Scenario: Group becomes a project
- GIVEN a Kuma group monitor "Payments"
- WHEN the export is imported
- THEN a project `payments` exists

#### Scenario: Monitor becomes a service
- GIVEN a Kuma monitor with name, url, interval, timeout, and maxretries
- WHEN the export is imported
- THEN a matching service exists with those values

#### Scenario: Override is applied
- GIVEN a Kuma monitor with a `notificationIDList` entry for a Slack notification
- WHEN the export is imported
- THEN the matching service has that channel as its override

### Requirement: Unsupported types skip cleanly
The product SHALL skip monitor types it does not support and SHALL list them in
the import report rather than failing the import.

#### Scenario: Mixed export
- GIVEN an export with supported and unsupported monitor types
- WHEN it is imported
- THEN supported monitors are created and unsupported ones are only reported

### Requirement: Imported secrets stay secret
The product SHALL encrypt credentials found in an import and SHALL never write
them to logs or API responses.

#### Scenario: Notification credential
- GIVEN an export containing a Slack token or webhook
- WHEN it is imported
- THEN the value is stored encrypted and appears in no log or response

### Requirement: Idempotent-safe import
Re-applying the same Kuma export SHALL NOT create duplicate projects, services,
or channels.

#### Scenario: Apply twice
- GIVEN an export already applied
- WHEN the same export is applied again
- THEN existing resources are matched by slug and no duplicates are created
