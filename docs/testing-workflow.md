# Chemistry UNO Testing Workflow

This document defines the frontend/backend testing flow for local development, pull requests, full-stack end-to-end checks, and release validation.

## Command Matrix

| Stage | Command | Runs | Blocks |
| --- | --- | --- | --- |
| Quick local | `pnpm test:quick` | Go backend tests, frontend type-check, Vitest unit/component tests | Local confidence |
| Standard CI | `pnpm test:ci` | Quick checks plus frontend production build | Pull requests and main branch CI |
| Full-stack e2e | `pnpm test:e2e` | Test database reset, backend test server, Vite test server, Playwright | Manual/nightly e2e gate |
| Release validation | `pnpm test:release` | Standard CI, tagged backend script tests, e2e, full build | Release candidate gate |
| Coverage baseline | `pnpm test:coverage` | Backend and frontend coverage summaries | Reporting only |

`pnpm test` remains a developer-friendly alias for the quick workflow. Use `pnpm test:ci` before opening a PR when possible.

## Test Layout

All test sources are archived under `tests/`:

- Backend Go tests live under `tests/_backend/` using the same package-relative layout as `backend/`.
- Frontend unit and component tests live under `tests/frontend/unit/`.
- Playwright tests live under `tests/frontend/e2e/`.
- Repository-level test entrypoints live at `tests/test.js`, `tests/test_main.py`, and `tests/feature-coverage-audit.test.js`.

Run backend tests through `node scripts/run-backend-tests.js ...`; it materializes archived Go test files into `backend/` only for the duration of the command and cleans them up afterward.

## Test Ownership

Backend changes:

- Domain logic in `backend/game`, `backend/anticheat`, `backend/utils`, and similar packages needs Go tests for success, failure, and important edge cases.
- Repository, database, and migration changes need tests that prove schema compatibility and persisted behavior.
- Handler or middleware changes need HTTP-level tests for status codes, auth behavior, validation errors, and JSON response shape.
- Build-tagged scripts must name their required command, such as `node scripts/run-backend-tests.js -tags scripts backend/scripts/oauth_third_party_test.go -v`.

Frontend changes:

- Vue pages, components, composables, and utilities need Vitest coverage using Vue Test Utils or utility-level tests.
- API adapter changes in `frontend/src/utils/api.ts` need mocked success, validation error, auth error, and network failure coverage.
- Type-check and production build failures block CI.

Cross-stack changes:

- Endpoint contract changes need backend handler tests and matching frontend API tests or Playwright coverage.
- User journeys that span login, auth, game rooms, admin flows, anticheat, or persistence need either API contract coverage plus a representative e2e scenario, or a documented reason why automation is not practical.
- Bug fixes should include a regression test that would have failed before the fix.

Refactors:

- Refactors that claim no behavior change must pass the standard workflow.
- Add focused tests when a refactor touches shared contracts, auth, persistence, game state, or other high-risk infrastructure.

## Full-Stack E2E Environment

Default ports:

- Frontend Vite test server: `http://127.0.0.1:5000`
- Backend test server: `http://127.0.0.1:8080`

Default test database:

- `tmp/e2e/chemistryuno-e2e.db`
- The e2e runner resets this database before a run and removes SQLite WAL/SHM sidecar files.
- Do not point e2e at `chemistryuno.db`; that file is for normal local use.

Seeded accounts:

| Username | Password | Role |
| --- | --- | --- |
| `admin` | `admin123` | admin |
| `test` | `test123` | user |
| `test1` - `test4` | `123456` | user |

Important environment variables:

| Variable | Default in e2e | Purpose |
| --- | --- | --- |
| `DB_TYPE` | `sqlite` | Forces SQLite for deterministic local e2e |
| `SQLITE_PATH` | `tmp/e2e/chemistryuno-e2e.db` | Isolated test database |
| `JWT_SECRET` | test-only fixed secret | Avoids mutating `.env` during test startup |
| `REDIS_ENABLED` | `false` | Keeps e2e independent from local Redis |
| `GIN_MODE` | `release` | Matches server runtime expectations |
| `CHEM_SERVER_ORIGIN` | `http://127.0.0.1:8080` | Vite proxy target for `/api` during e2e |

The e2e runner starts services it owns and writes logs under `tmp/e2e/logs/`. Playwright failure traces, screenshots, and videos are written under `frontend/test-results/` and `frontend/playwright-report/`.

## Failure Triage

1. Check the failing stage name printed by the runner.
2. For backend failures, rerun the named `node scripts/run-backend-tests.js` command with `-v`.
3. For frontend unit failures, rerun `pnpm -C frontend test`.
4. For e2e failures, inspect `tmp/e2e/logs/`, `frontend/test-results/`, and the Playwright report.
5. If an e2e failure is timing-related, prefer stable user-visible waits, seeded data, and helper functions over arbitrary sleeps.

## Coverage Baseline

`pnpm test:coverage` generates backend and frontend summaries for review. It intentionally does not enforce a percentage threshold until the project adopts an approved baseline.
