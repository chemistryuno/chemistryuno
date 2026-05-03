## 1. Frontend Modal Centering

- [x] 1.1 Locate the shared dialog renderer and all component-local modal overlay roots used by the Vue frontend.
- [x] 1.2 Update shared alert, confirm, and prompt dialog markup/styles so overlays are fixed to the viewport and content centers in the visible screen.
- [x] 1.3 Update component-local modal overlays to use the same viewport-fixed centering contract.
- [x] 1.4 Ensure modal content on small viewports keeps accessible overflow without shifting the dialog to the document center.

## 2. Backend Startup Nickname Repair

- [x] 2.1 Add repository support for finding users whose trimmed nickname is empty or missing.
- [x] 2.2 Add a startup repair function that assigns valid generated nicknames only to users with blank nicknames.
- [x] 2.3 Add uniqueness handling for generated startup nicknames, including collision retries and deterministic fallback.
- [x] 2.4 Invoke the repair function during server startup after database/repository initialization and before player-facing services begin accepting interactions.
- [x] 2.5 Log startup repair results and errors with enough detail to diagnose failed repairs.

## 3. Verification

- [x] 3.1 Add or update backend tests covering blank nickname discovery, idempotent repair, and nickname collision handling.
- [x] 3.2 Add or update frontend checks for viewport-centered dialogs after page scroll and on a small viewport.
- [x] 3.3 Run relevant backend and frontend test/build commands and record any remaining limitations.
