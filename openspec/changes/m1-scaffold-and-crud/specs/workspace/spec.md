# Delta for Workspace

## ADDED Requirements

### Requirement: Singleton workspace
One deployment SHALL represent exactly one workspace, created automatically on
first boot and reused thereafter.

#### Scenario: First boot creates the workspace
- GIVEN an empty database
- WHEN the binary starts
- THEN exactly one workspace row exists

#### Scenario: Restart does not duplicate
- GIVEN a workspace already exists
- WHEN the binary restarts
- THEN no additional workspace row is created

### Requirement: Project management
The product SHALL allow projects to be created, listed, viewed, renamed, and
deleted within the workspace, each identified by a stable slug.

#### Scenario: Create a project
- GIVEN an empty workspace
- WHEN a project named "Payments" is created
- THEN a project with slug `payments` is returned
- AND it appears in the project list

#### Scenario: Delete a project
- GIVEN a project with services
- WHEN the project is deleted
- THEN the project and its services are removed from the project list

### Requirement: Service registry
The product SHALL allow services to be created, listed, viewed, edited,
deleted, paused, and resumed within a project, each carrying exactly one check
configuration and a stable slug unique within the project.

#### Scenario: Create an HTTP service
- GIVEN a project `payments`
- WHEN a service is created with type `http` and a URL
- THEN the service is returned with a `project_id` equal to the project's id
- AND its slug is unique within the project

#### Scenario: Pause and resume a service
- GIVEN an enabled service
- WHEN the service is paused
- THEN its paused state is reported
- WHEN it is resumed
- THEN its enabled state is reported

#### Scenario: Validate check configuration
- GIVEN a service type of `http`
- WHEN a service is created without a URL
- THEN the request fails with a validation error naming the URL field

### Requirement: Project ownership of data
Every project-owned table SHALL carry `project_id`, and every project-owned API
path SHALL be nested under its project.

#### Scenario: Service belongs to one project
- GIVEN services in projects `payments` and `checkout`
- WHEN listing services for `payments`
- THEN only `payments` services are returned
