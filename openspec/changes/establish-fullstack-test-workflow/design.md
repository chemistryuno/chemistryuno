## Context

Chemistry UNO is a Go backend plus Vue/Vite frontend application with existing test assets:

- Backend tests run with `go test ./backend/...`, plus selected script-tag tests such as OAuth checks.
- Frontend checks include `pnpm -C frontend type-check`, `pnpm -C frontend test`, and `pnpm -C frontend build`.
- Playwright is configured under `frontend/tests/e2e` and currently starts the frontend test server on port 5000.
- Root-level scripts include `pnpm test`, `pnpm test:ci`, `pnpm go:test`, and build commands, but their coverage and blocking semantics differ.

The workflow should make these tools feel like one system: fast enough for daily development, broad enough for pull requests, and realistic enough for release confidence.

## Goals / Non-Goals

**Goals:**
- Establish a test pyramid for backend, frontend, API contract, and full-stack e2e coverage.
- Provide deterministic local and CI commands for quick checks, standard PR checks, and release validation.
- Define test environment setup for SQLite database state, seeded users, frontend test mode, backend service lifecycle, and cleanup.
- Make failure artifacts actionable by preserving logs, coverage summaries, and Playwright traces/screenshots where useful.
- Create contribution expectations for new features, bug fixes, migrations, and cross-stack behavior.

**Non-Goals:**
- Replacing Go, Vitest, or Playwright with a different test framework.
- Requiring every developer change to run the slowest full-stack suite locally.
- Introducing production-only dependencies for test automation.
- Mandating a fixed coverage percentage before the project has a stable baseline.

## Decisions

1. Use layered test commands instead of one monolithic test command.

   The workflow will define quick, standard, e2e, and release commands. Quick checks keep local feedback short, while CI and release checks can run broader suites. Alternative considered: making `pnpm test` always run everything. That would be simpler to explain but too slow and brittle for day-to-day iteration.

2. Treat Go backend tests and Vue frontend tests as first-class required PR gates.

   `go test ./backend/...`, `pnpm -C frontend type-check`, `pnpm -C frontend test`, and `pnpm -C frontend build` should all be part of the standard CI gate. Alternative considered: only type-checking and building the frontend. That misses component behavior regressions already covered by Vitest.

3. Promote Playwright to a separate full-stack gate with an explicit backend test server.

   Playwright tests should run against a deterministic test backend and frontend test mode, with seeded users and cleanup. Alternative considered: mocking all backend calls in Playwright. Mocking is useful for component tests, but it cannot validate route wiring, auth, database effects, websocket flows, or cross-stack regressions.

4. Add API contract tests at the backend boundary.

   Backend handler tests should validate status codes, request validation, auth behavior, and JSON response shape for important endpoints. Frontend API utilities should be covered with mocked HTTP responses. Alternative considered: relying only on e2e tests. E2E coverage is slower and makes contract failures harder to pinpoint.

5. Keep test data isolated and repeatable.

   Full-stack tests should use a test database created or reset by scripts, seeded with documented accounts such as admin and test users. Tests must not depend on the developer's live local database. Alternative considered: sharing `chemistryuno.db`. That risks data corruption and produces order-dependent failures.

6. Make coverage expectations feature-based.

   New backend logic needs Go tests, new frontend behavior needs Vitest or component tests, and cross-stack flows need API contract or Playwright coverage. Alternative considered: only enforcing global coverage numbers. A coverage percentage can hide missing critical flows, especially in game, auth, anticheat, and admin paths.

## Risks / Trade-offs

- [Slower CI] -> Split quick PR checks and optional/nightly full e2e checks, and cache Go/pnpm dependencies.
- [Flaky e2e tests] -> Use deterministic seed data, stable selectors, trace-on-retry, explicit waits on user-visible state, and test database reset between runs.
- [Environment drift] -> Document required ports, env vars, test mode, and database lifecycle in one runbook.
- [Duplicated test setup] -> Centralize backend fixtures, frontend test setup, and Playwright login helpers.
- [Coverage churn] -> Start with workflow requirements and baseline reporting before enforcing hard coverage thresholds.

## Migration Plan

1. Document the test matrix and expected commands in developer docs.
2. Align root scripts so quick, CI, e2e, and release commands run the same steps locally and in CI.
3. Add deterministic full-stack environment setup for backend service startup, test database reset, and seed data.
4. Update Playwright configuration and helpers to use the full-stack test environment.
5. Extend CI to publish test logs, coverage summaries, and Playwright artifacts.
6. Add or backfill representative tests for critical backend, frontend, and full-stack flows.
