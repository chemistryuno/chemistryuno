## Context

The frontend uses a shared dialog utility plus several component-local modal overlays. Some overlays can inherit layout behavior from page scroll or container geometry, which makes "centered" mean the document/page center rather than the user's visible screen. The backend already has a random nickname generator and user repository update methods, but startup does not currently repair legacy or corrupted users whose nickname is empty or whitespace-only.

## Goals / Non-Goals

**Goals:**
- Make all modal dialogs visually center in the browser viewport the user is currently looking at.
- Keep modal overlays fixed to the viewport across scrolling, responsive layouts, and nested page containers.
- Repair blank player nicknames once during server startup after database and repositories are initialized.
- Generate valid, unique fallback nicknames using the existing nickname generation style.
- Keep the repair pass idempotent so repeated restarts do not change players who already have nicknames.

**Non-Goals:**
- Redesigning modal visual styling, content, animation, or interaction patterns.
- Changing nickname requirements for registration or profile editing.
- Assigning nicknames to non-player system records unless they are stored in the same user table and missing the same required display value.
- Adding a user-facing flow for choosing a replacement nickname during startup repair.

## Decisions

1. Use viewport-fixed modal positioning as the frontend contract.

   Modal roots should use fixed positioning against the viewport (`position: fixed; inset: 0`) and center content with viewport-relative alignment. This matches the user request directly and avoids scroll-offset math. Alternatives considered were measuring `window.scrollY` and manually offsetting absolute dialogs, but that is more fragile with nested scrollers, mobile browser chrome, and responsive shells.

2. Centralize shared dialog behavior and audit component-local modals.

   The shared dialog utility should remain the primary path for alert/confirm/prompt behavior, with its rendered component enforcing fixed viewport centering. Component-local modals should be updated where needed to follow the same overlay contract. This avoids introducing a new modal framework while still covering existing bespoke overlays.

3. Run nickname repair during backend startup after repository initialization.

   The repair pass needs database access and should run before long-lived services begin accepting player interactions. Placing it after `repository.InitRepositories()` keeps it near other startup data normalization tasks. Running it before database initialization is impossible; running it lazily on player login would leave blank names visible in admin, replay, chat, or background flows.

4. Generate-and-check unique nicknames per affected user.

   For each blank nickname, generate a candidate via the existing random nickname helper, check for collisions through the user repository, and retry with a bounded attempt count before falling back to a UID-derived suffix. This keeps normal output friendly while ensuring startup cannot loop forever.

## Risks / Trade-offs

- Existing modal components may have inconsistent local markup -> audit likely modal files with search and update each overlay root rather than only the shared dialog.
- Mobile browser viewport behavior can vary when the URL bar collapses -> use CSS fixed positioning and safe max-height/overflow rules, then verify at desktop and mobile viewport sizes.
- Random nickname collisions are possible -> repository-level existence checks and bounded retries reduce the risk, with deterministic fallback for the rare exhausted case.
- Startup repair mutates existing data -> log the number of repaired users and only update records whose trimmed nickname is empty.

## Migration Plan

1. Deploy the backend change with the startup repair pass.
2. On first startup after deployment, users with blank nicknames receive generated nicknames and subsequent restarts leave them unchanged.
3. Deploy frontend modal positioning updates with normal frontend assets.
4. Rollback is safe for frontend behavior. Backend rollback does not undo assigned nicknames; repaired values remain normal user data.

## Open Questions

- None. The requested behavior is specific enough to implement without additional product choices.
