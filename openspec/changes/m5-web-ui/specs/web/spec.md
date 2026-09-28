# Delta for Web

## ADDED Requirements

### Requirement: Sign-in screen
The web application SHALL provide a sign-in screen that starts the Google
sign-in flow and SHALL redirect unauthenticated visitors to it.

#### Scenario: Unauthenticated visit
- GIVEN no active session
- WHEN a protected page is opened
- THEN the user is redirected to the sign-in screen

### Requirement: Status dashboard
The web application SHALL show the current status of every service in the
selected project, including up, down, pending, and paused.

#### Scenario: Live status
- GIVEN a project with services in different states
- WHEN the dashboard is viewed
- THEN each service is shown with its current state

### Requirement: Project switcher
The web application SHALL let the user switch between projects, and SHALL scope
all subsequent views to the selected project.

#### Scenario: Switch project
- GIVEN two projects with different services
- WHEN the user switches to the other project
- THEN only that project's services are shown

### Requirement: Service management
The web application SHALL let members create, edit, pause, resume, and delete
services, and SHALL show service detail with recent checks and uptime.

#### Scenario: Create a service from the UI
- GIVEN a signed-in member
- WHEN they submit the create-service form
- THEN the service appears in the list and on the dashboard

#### Scenario: Service detail
- GIVEN a service with check history
- WHEN its detail page is opened
- THEN current status, recent checks, and uptime are shown

### Requirement: Role-aware interface
The web application SHALL hide or disable actions a user's role does not allow,
while the server remains the authority.

#### Scenario: Viewer controls
- GIVEN a signed-in viewer
- WHEN the service list is shown
- THEN create, edit, pause, and delete controls are not available

### Requirement: Alert configuration interface
The web application SHALL let members connect Slack, choose a project default
channel and per-service overrides, edit templates with preview, and send a test
alert.

#### Scenario: Template preview
- GIVEN the template editor is open
- WHEN a template is edited
- THEN a preview with sample data is shown before saving

### Requirement: Delivery failure view
The web application SHALL show recent alert delivery failures with their
reasons.

#### Scenario: View failures
- GIVEN one or more failed deliveries
- WHEN the delivery view is opened
- THEN each failure is listed with its reason and time
