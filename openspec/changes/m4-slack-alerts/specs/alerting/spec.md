# Delta for Alerting

## ADDED Requirements

### Requirement: Slack workspace connection
The product SHALL connect a Slack workspace through Slack's OAuth app flow and
SHALL store the resulting bot token encrypted at rest.

#### Scenario: Successful install
- GIVEN a valid Slack app and redirect URL
- WHEN an admin completes the OAuth install
- THEN the workspace is marked connected and the bot token is stored encrypted

#### Scenario: Token never exposed
- GIVEN a connected Slack workspace
- WHEN any API response or log is produced
- THEN the raw bot token never appears in the response or the log

### Requirement: Channel selection
The product SHALL send a service's alerts to its service channel if set,
otherwise to the project default channel, and SHALL never fall back to a
workspace default.

#### Scenario: Service override wins
- GIVEN a service with channel A and a project default channel B
- WHEN the service's channel is resolved
- THEN channel A is selected

#### Scenario: Project default used
- GIVEN a service with no channel and a project default channel B
- WHEN the service's channel is resolved
- THEN channel B is selected

#### Scenario: No channel resolves
- GIVEN a service with no channel and no project default
- WHEN an alert would be delivered
- THEN no message is sent and a delivery failure is recorded

### Requirement: Template resolution
The product SHALL resolve a message template per trigger as service override,
else project default, else built-in default.

#### Scenario: Built-in fallback
- GIVEN no service or project template for `down`
- WHEN a `down` alert is rendered
- THEN the built-in `down` template is used

#### Scenario: Service override wins
- GIVEN a service and project template for `down`
- WHEN a `down` alert is rendered
- THEN the service template is used

### Requirement: Alert on down and recovery
The product SHALL send exactly one alert when a service becomes down and
exactly one when it recovers, with no duplicate for repeated checks in the same
state.

#### Scenario: Down alert
- GIVEN an up service with a resolvable channel
- WHEN it transitions to down
- THEN exactly one down alert is delivered

#### Scenario: Recovery alert
- GIVEN a down service
- WHEN it transitions to up
- THEN exactly one recovery alert is delivered

#### Scenario: No duplicate
- GIVEN a service already alerted as down
- WHEN further failing checks occur without a state change
- THEN no additional alert is delivered

### Requirement: Message templates with variables
Templates SHALL support the variables `{{service.name}}`, `{{service.url}}`,
`{{project.name}}`, `{{status}}`, `{{duration}}`, and `{{error}}`.

#### Scenario: Variables render
- GIVEN a template containing `{{service.name}}` and `{{status}}`
- WHEN the alert is rendered
- THEN the output contains the service name and the current status

#### Scenario: Duration reflects time in state
- GIVEN a service down for 10 minutes
- WHEN a recovery alert is rendered
- THEN `{{duration}}` reflects the 10 minutes in the down state

### Requirement: Template preview and test send
The product SHALL render a template preview with sample data and SHALL allow an
admin or member to send a test alert to a selected channel.

#### Scenario: Preview
- GIVEN a template with variables
- WHEN preview is requested
- THEN the rendered output is returned without sending to Slack

#### Scenario: Test send
- GIVEN a connected workspace and a selected channel
- WHEN a test send is requested
- THEN one message is posted to that channel

### Requirement: Delivery failure log
The product SHALL record every alert that cannot be delivered, with the
reason, and SHALL never drop an alert silently.

#### Scenario: Slack API error
- GIVEN a channel that no longer exists
- WHEN an alert is delivered
- THEN a delivery failure is recorded with the Slack error reason
