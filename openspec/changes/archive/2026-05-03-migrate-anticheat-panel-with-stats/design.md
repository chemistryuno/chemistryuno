## Context

The Chemistry UNO anticheat system exists in the backend (Go) with an emerging admin UI (Vue 3 frontend). Currently:
- Bans and appeals are managed through disparate handlers
- Audit trails lack structured organization
- Players have zero visibility into anticheat metrics
- Compensation for false bans is managed inconsistently
- No unified admin experience for reviewing and managing cases

This design consolidates these functions into a single coherent system with dual surfaces: admin panel for management and player dashboard for transparency.

## Goals / Non-Goals

**Goals:**
- Provide a unified anticheat admin panel with detection, appeals, configuration, and audit tabs
- Display player-facing stats: bans today, system uptime (running days)
- Implement configurable unban compensation (fuel amount + message template)
- Create audit trail with compensation status tracking
- Enable idempotent appeal approvals (prevent duplicate compensation issuance)
- Support compensation failure recovery

**Non-Goals:**
- Rewriting core anticheat detection logic
- Migrating existing ban records to new schema
- Implementing real-time notifications for bans
- Adding appeal rejection workflow (only approval covered in this change)

## Decisions

### 1. **Single Vue Component for Admin Panel**
**Decision:** Consolidate all anticheat management into `frontend/src/pages/AdminAnticheat.vue` with tabbed navigation

**Rationale:** Reduces cognitive load; all related functions are co-located. Tab structure mirrors admin workflow: detect → appeal → configure → audit.

**Alternatives Considered:**
- Separate pages per function: More granular but scattered context
- Wizard modal: Forces linear flow; appeals may be high-volume random access

### 2. **Compensation as Appeal Approval Parameter**
**Decision:** Compensation (amount + message) is submitted with appeal approval, not as separate operation

**Rationale:** Ensures compensation and case resolution are atomic; prevents approval without compensation. Configurable defaults reduce admin friction.

**Alternatives Considered:**
- Two-step process: Approve → then send compensation: Adds complexity; risk of forgetting second step
- Pre-set policy only: No flexibility for admin to adjust case-by-case

### 3. **Idempotent Approval with Redis Key**
**Decision:** Use Redis key `unban_compensation:{user_id}:{event_id}` with TTL to detect duplicate approvals

**Rationale:** Prevents double-issuing fuel if approval endpoint is called multiple times (network retry, admin refresh). TTL allows retry after window expires.

**Alternatives Considered:**
- Approval status column only: Doesn't prevent duplicate calls from racing
- Database unique constraint: Requires schema migration and slower to check

### 4. **Audit Trail as Read-Only Ledger**
**Decision:** Compensation status and metadata persisted in audit table, not normalized into users/bans table

**Rationale:** Maintains separation of concerns; audit is immutable historical record. Can archive old records independently.

**Alternatives Considered:**
- Update ban record: Couples ban lifecycle to compensation; harder to handle compensation failures
- Separate compensation table: More normalized but requires JOIN for audit queries

### 5. **Player Stats via Real-Time API Endpoints**
**Decision:** Player dashboard calls `/api/player/anticheat/stats` which returns `{ bans_today: N, system_uptime_days: X }`

**Rationale:** Decouples player stats from admin queries; can cache/rate-limit independently. Trivial to query (single endpoint).

**Alternatives Considered:**
- Include in profile endpoint: Pollutes user data; changes to stats require profile migration
- Broadcast system stats separately: Adds complexity; stats may become stale

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| **Compensation Race Condition** - Two admins approve same appeal simultaneously, user gets double fuel | Use Redis key with short TTL + database transaction on fuel issuance. Check compensation already issued before adding fuel. |
| **Player Stats Staleness** - Bans logged but stats delayed | Accept eventual consistency; stats cache TTL ~5min is acceptable for transparency use case |
| **Schema Migration Burden** - Adding compensation columns to audit table requires DB migration | Migration is backward-compatible (nullable columns); can deploy without downtime. Include in release notes. |
| **Overcompensation Liability** - If admin enters large amount by mistake | Implement admin confirmation modal (large amounts require approval). Default to conservative value (100). Log all approvals. |

## Migration Plan

### Phase 1: Backend Infrastructure (2 days)
- Add `compensation_amount`, `compensation_status`, `compensation_note` columns to audit table
- Implement idempotency check in Redis cache
- Extend appeal approval handler to accept compensation parameters
- Add admin API endpoints for stats and config

### Phase 2: Frontend Dashboard (2 days)
- Build AdminAnticheat.vue with tabs
- Implement approval modal with compensation fields
- Wire up API calls for appeal management
- Build player stats display component

### Phase 3: Integration & Testing (1 day)
- E2E tests for approval workflow
- Load testing for audit queries
- Manual admin workflow validation

### Rollback Strategy
- Feature flag: `ENABLE_ANTICHEAT_PANEL` (default false) to gate new admin panel
- If compensation mechanism breaks, disable via config and revert approvals manually (compensation records remain in audit for forensics)

## Open Questions

1. **Should compensation failure auto-retry?** Current design assumes sync failure → error response. Should we implement async retry with exponential backoff?
2. **Player notification format:** Should players receive in-game system message, email, or both?
3. **Appeal rejection:** Only approval covered here—should we implement rejection workflow too?
4. **Audit retention policy:** How long should compensation records be kept (30/60/90 days)?
