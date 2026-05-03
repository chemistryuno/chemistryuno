## Why

The project already has Go tests, Vitest tests, Playwright tests, and CI scripts, but they are not organized into one explicit frontend/backend testing workflow. A documented and automated workflow will make local development, pull request checks, and release validation repeatable across the Go backend, Vue frontend, and full-stack game flows.

## What Changes

- Define a layered test workflow covering backend unit/integration tests, frontend unit/component tests, API contract checks, Playwright end-to-end tests, and release smoke checks.
- Standardize which commands run locally, in CI, and before release, including expected pass/fail gates for each stage.
- Introduce test environment rules for database setup, seeded users, backend service startup, frontend test mode, and deterministic cleanup.
- Define coverage ownership expectations for new features and bug fixes so specs, handlers, components, and full-stack flows receive the right test depth.
- Add reporting expectations for CI output, e2e traces, coverage artifacts, and failure triage.

## Capabilities

### New Capabilities
- `fullstack-test-workflow`: Defines the complete testing workflow, quality gates, environments, and coverage expectations for frontend, backend, and full-stack behavior.

### Modified Capabilities
- None.

## Impact

This change affects repository test scripts, CI configuration, frontend Playwright/Vitest conventions, Go test conventions, test data setup scripts, and developer documentation. It does not require production API behavior changes, but it may add or adjust non-production test utilities, fixtures, and automation scripts.
