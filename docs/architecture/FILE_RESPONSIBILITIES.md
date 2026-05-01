# File Responsibilities and Layer Boundaries

This document clarifies where code should live and which layer owns each responsibility.

## Runtime Entrypoints

- `main.go`: only application bootstrap and HTTP router wiring.
- `start.js`: local developer orchestrator for frontend + backend.
- `build.js`: production build orchestration.
- `test.js`: repository-level test orchestration.
- `init.js`: first-time setup workflow.

## Backend Layers (`backend/`)

- `router/`: route composition and middleware binding per route group.
- `handlers/`: HTTP request parsing, response shaping, auth gate usage. No direct SQL.
- `repository/`: data access and query composition. No HTTP concerns.
- `game/`: game engine domain logic and room lifecycle.
- `middleware/`: cross-cutting request filters.
- `database/`: schema, migration, DB initialization.
- `websocket/`: socket lifecycle and publish mechanics.
- `utils/`: shared stateless helpers.
- `scripts/`: one-off and maintenance scripts that operate on backend domain/data.

## Frontend Layers (`frontend/src/`)

- `pages/`: route-level containers.
- `components/`: reusable visual components.
- `composables/`: reusable stateful logic and side effects.
- `services/` and `utils/`: API wrappers, protocol adapters, pure helpers.

## Script Placement Rules

- Put backend-runtime DB/data migration scripts in `backend/scripts/`.
- Put repository-wide automation scripts in `scripts/`.
- Keep `tools/` for standalone Go tooling modules and one-off maintenance binaries.
- Prefer command entrypoints under `tools/cmd/*` and avoid duplicate root-level `tools/*.go` mains.
- Avoid duplicate scripts with the same purpose in different folders.

## Ownership Rules

- If a file primarily defines URL paths/groups/middleware combinations, it belongs in `router`.
- If a file imports `gin` and contains request business handling, it belongs in `handlers`.
- If a file imports `gorm` for queries, it belongs in `repository`.
- If a file changes turn logic, rules, or AI decisions, it belongs in `game`.
- If a file is only used by `go run .../scripts/...`, it belongs in `backend/scripts`.

## Review Checklist for New Files

- Is the file in the layer that owns the behavior?
- Does it avoid crossing into neighboring layer responsibilities?
- Is there already another file owning the same concern?
- Can runtime entry scripts remain orchestration-only after this change?
