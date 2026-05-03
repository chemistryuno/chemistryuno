## Why

Some modal dialogs are centered relative to the full page instead of the user's current viewport, so they can appear off-screen after scrolling or on constrained displays. Separately, existing player records may have empty nicknames; these should be repaired automatically at server startup so game, chat, replay, and profile surfaces always have a usable display name.

## What Changes

- Standardize modal placement so every in-app dialog and overlay appears in the center of the user's visible screen/viewport.
- Add a server-startup nickname repair pass that finds players with missing or blank nicknames and assigns each one a generated random nickname.
- Ensure generated nicknames are valid for existing nickname rules and avoid collisions with existing nicknames.
- Log startup repair results without blocking normal startup unless the database operation itself fails critically.

## Capabilities

### New Capabilities
- `screen-centered-modals`: Defines the user-visible placement requirement for all modal dialogs and overlays.
- `startup-default-nicknames`: Defines server startup behavior for repairing player records that do not have a nickname.

### Modified Capabilities

## Impact

- Frontend modal/dialog utilities and Vue modal components that currently rely on page-level centering or scroll-position-sensitive layouts.
- Backend startup flow, user repository methods, and nickname generation/update logic.
- Tests for modal layout behavior and startup nickname repair, including uniqueness and blank-string handling.
