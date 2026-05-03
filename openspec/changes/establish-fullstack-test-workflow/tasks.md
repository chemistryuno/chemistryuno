## 1. Test Workflow Documentation

- [x] 1.1 Create a developer-facing test workflow document that defines quick, standard CI, full-stack e2e, and release validation commands
- [x] 1.2 Document which test layer is required for backend logic, handlers, migrations, frontend components, API adapters, and cross-stack user journeys
- [x] 1.3 Document test environment variables, ports, seeded accounts, database lifecycle, and troubleshooting steps

## 2. Root Test Script Alignment

- [x] 2.1 Update root package scripts to expose explicit quick, CI, e2e, and release validation commands
- [x] 2.2 Align `scripts/test-ci.js` with the required standard CI gate: backend tests, frontend type-check, frontend unit/component tests, and frontend build
- [x] 2.3 Ensure test scripts fail consistently when required stages fail and print the failing command name

## 3. Backend Test Environment

- [x] 3.1 Add or update backend test database reset and seed flow for isolated full-stack test runs
- [x] 3.2 Add a deterministic backend test server startup command or script for Playwright usage
- [x] 3.3 Add backend handler or contract tests for representative critical routes not already covered
- [x] 3.4 Add documentation or script coverage for required tagged backend tests used during release validation

## 4. Frontend Unit and Contract Tests

- [x] 4.1 Ensure Vitest setup supports stable tests for components, pages, composables, utilities, and API adapters
- [x] 4.2 Add or update representative frontend API adapter tests for success, validation error, auth error, and network failure responses
- [x] 4.3 Ensure frontend type-check, unit tests, and production build run cleanly from the root workflow

## 5. Full-Stack E2E Workflow

- [x] 5.1 Update Playwright configuration or wrapper scripts so e2e runs use both frontend test mode and a deterministic backend test server
- [x] 5.2 Centralize Playwright helpers for seeded user login, API setup, cleanup, and stable selectors
- [x] 5.3 Ensure e2e failure artifacts include traces, screenshots, and backend/frontend service logs
- [x] 5.4 Add or backfill representative e2e coverage for at least one auth/admin flow and one player gameplay-adjacent flow

## 6. CI and Reporting

- [x] 6.1 Update GitHub Actions CI to run the standard CI gate and cache Go/pnpm dependencies
- [x] 6.2 Add an e2e-capable CI job or documented manual/nightly workflow with artifact upload
- [x] 6.3 Add optional backend and frontend coverage summary generation without enforcing a new threshold

## 7. Verification

- [x] 7.1 Run the quick local workflow and record the command result
- [x] 7.2 Run the standard CI workflow locally and record the command result
- [x] 7.3 Run the full-stack e2e workflow locally and record any environment assumptions or known limitations
- [x] 7.4 Validate the OpenSpec change status is apply-ready
