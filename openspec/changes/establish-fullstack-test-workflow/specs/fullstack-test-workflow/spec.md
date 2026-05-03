## ADDED Requirements

### Requirement: Project exposes layered test commands
The system SHALL provide documented commands for quick local checks, standard CI checks, full-stack e2e checks, and release validation.

#### Scenario: Developer runs quick local checks
- **WHEN** a developer runs the quick test command
- **THEN** the command executes backend unit tests, frontend type-checking, and frontend unit/component tests without requiring a browser e2e environment

#### Scenario: CI runs standard pull request checks
- **WHEN** CI validates a pull request
- **THEN** CI runs backend tests, frontend type-checking, frontend unit/component tests, and frontend build checks
- **AND** CI fails if any required standard check fails

#### Scenario: Release validation runs the broad suite
- **WHEN** release validation is requested
- **THEN** the workflow runs standard checks, full-stack e2e checks, build verification, and any required script-tag tests
- **AND** the release validation command fails if any required stage fails

### Requirement: Backend tests cover domain, persistence, handlers, and scripts
The system SHALL define backend test coverage expectations for Go domain logic, repository/database behavior, HTTP handlers, middleware-sensitive behavior, and script-tag utilities.

#### Scenario: Backend business logic changes
- **WHEN** a change modifies game, anticheat, auth, repository, database, or utility logic
- **THEN** the change includes or updates Go tests for success paths, failure paths, and relevant edge cases

#### Scenario: Backend API behavior changes
- **WHEN** a change modifies an HTTP route, request body, response body, auth requirement, or status code
- **THEN** backend tests validate the route behavior with representative requests and expected JSON responses

#### Scenario: Database schema changes
- **WHEN** a change adds or modifies a database migration or persisted model
- **THEN** backend tests validate migration compatibility and the repository behavior that depends on the schema

#### Scenario: Script-tag backend checks are required
- **WHEN** a workflow includes backend scripts guarded by Go build tags
- **THEN** the workflow explicitly names the required tagged test command and includes it in release validation

### Requirement: Frontend tests cover components, pages, API adapters, and build health
The system SHALL define frontend test coverage expectations for Vue components, pages, composables, utilities, API adapters, type safety, and production build health.

#### Scenario: Frontend user interface behavior changes
- **WHEN** a change modifies a page, component, composable, or user interaction
- **THEN** the change includes or updates Vitest coverage using Vue Test Utils or utility-level tests

#### Scenario: Frontend API usage changes
- **WHEN** a change modifies frontend API calls, request payloads, response parsing, or error handling
- **THEN** frontend tests cover success and failure responses with mocked network behavior

#### Scenario: Frontend static checks run
- **WHEN** standard CI checks run
- **THEN** the workflow executes frontend type-checking and production build verification
- **AND** failures in either check block the CI result

### Requirement: Full-stack e2e tests use deterministic environment setup
The system SHALL run Playwright e2e tests against a deterministic frontend test server, backend test server, and isolated test database.

#### Scenario: E2E environment starts
- **WHEN** full-stack e2e tests start
- **THEN** the workflow prepares an isolated test database, seeds required admin and player accounts, starts the backend test server, and starts the frontend in test mode
- **AND** tests use documented ports and environment variables

#### Scenario: E2E test data is reset
- **WHEN** a full-stack e2e run begins
- **THEN** prior test data is reset or replaced so test results do not depend on previous local runs

#### Scenario: E2E tests complete
- **WHEN** a full-stack e2e run finishes
- **THEN** the workflow stops any test services it started or documents how reused servers are detected safely
- **AND** the workflow preserves failure artifacts such as Playwright traces, screenshots, and relevant service logs

### Requirement: API contracts are verified across frontend and backend boundaries
The system SHALL verify important frontend/backend contracts with backend handler tests, frontend API tests, and e2e coverage for critical flows.

#### Scenario: Contract-affecting backend change
- **WHEN** a backend endpoint changes request validation, response shape, auth behavior, or error semantics
- **THEN** backend tests assert the new contract
- **AND** frontend tests or e2e tests are updated if the client consumes the changed contract

#### Scenario: Contract-affecting frontend change
- **WHEN** frontend code starts consuming a new endpoint or response field
- **THEN** tests cover the expected payload shape and failure handling
- **AND** the workflow identifies the matching backend route or fixture that supplies the contract

### Requirement: CI reports actionable test results and artifacts
The system SHALL make CI failures diagnosable by preserving command output and relevant test artifacts.

#### Scenario: Unit or integration test fails in CI
- **WHEN** a required backend or frontend check fails in CI
- **THEN** CI output identifies the failing command and exits with a non-zero status

#### Scenario: E2E test fails in CI
- **WHEN** a Playwright test fails in CI
- **THEN** the workflow publishes or preserves Playwright traces, screenshots, and service logs needed for triage

#### Scenario: Coverage reporting is generated
- **WHEN** coverage reporting is enabled for a workflow stage
- **THEN** the workflow records backend and frontend coverage summaries without blocking on a new percentage threshold until a baseline is approved

### Requirement: Test ownership is tied to feature and bug changes
The system SHALL define test ownership rules for new features, bug fixes, refactors, and risk-bearing changes.

#### Scenario: New feature is implemented
- **WHEN** a new user-facing feature is implemented
- **THEN** the change includes backend tests for server behavior, frontend tests for UI behavior, and e2e or contract coverage for cross-stack user journeys when applicable

#### Scenario: Bug fix is implemented
- **WHEN** a bug is fixed
- **THEN** the change includes a regression test that fails before the fix or documents why automation is not practical

#### Scenario: Refactor keeps behavior unchanged
- **WHEN** a refactor claims no user-visible behavior change
- **THEN** the standard test workflow passes without requiring new product scenarios unless the refactor changes shared contracts or risky infrastructure
