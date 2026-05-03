## 1. Backend Database & Configuration

- [x] 1.1 Create database migration for audit table compensation columns (compensation_amount, compensation_status, compensation_message, compensation_note, compensation_date, approval_note)
- [x] 1.2 Add compensation configuration keys to config.yaml (unban.compensation_amount, unban.default_message, unban.enabled)
- [x] 1.3 Implement config loader in backend/config/config.go to read compensation defaults
- [x] 1.4 Create Redis cache utilities for idempotency key: `unban_compensation:{user_id}:{event_id}`

## 2. Backend API Endpoints - Statistics

- [x] 2.1 Implement GET /api/admin/anticheat/stats (returns ban counts by period and detection rules)
- [x] 2.2 Implement GET /api/player/anticheat/stats (returns bans_today and system_uptime_days)
- [x] 2.3 Implement caching for player stats (5-minute TTL) using Redis
- [x] 2.4 Add rate limiting to player stats endpoint (1 req/sec per user)
- [x] 2.5 Implement system uptime tracking (store startup timestamp in config or cache on boot)

## 3. Backend API Endpoints - Configuration & Audit

- [x] 3.1 Implement GET /api/admin/anticheat/config (returns current compensation configuration)
- [x] 3.2 Implement POST /api/admin/anticheat/config (updates compensation settings, validate ranges)
- [x] 3.3 Implement GET /api/admin/anticheat/audit (returns audit log with filtering: player_id, date_range, action_type, compensation_status)
- [x] 3.4 Implement CSV export for audit log with all compensation columns

## 4. Backend Appeal Approval & Compensation

- [x] 4.1 Extend POST /api/admin/anticheat/appeals/{id}/approve to accept compensation_amount and compensation_message parameters
- [x] 4.2 Implement idempotency check: query Redis key before issuing fuel, set key with 1-hour TTL on success
- [x] 4.3 Implement AddFuel logic in repository layer (or integrate with existing fuel system)
- [x] 4.4 Update appeal approval handler to persist compensation fields in audit table
- [x] 4.5 Implement compensation status flow: pending → ok/failed based on fuel issuance outcome
- [x] 4.6 Implement failure handling: log failed compensation with reason, do NOT delete approval, allow admin retry
- [x] 4.7 Add system message sending integration (notify player of unban + compensation)

## 5. Backend Database Layer

- [x] 5.1 Update audit repository to include compensation fields in INSERT queries
- [x] 5.2 Implement audit query builder supporting filters: player_id, date_range, action, compensation_status
- [x] 5.3 Implement audit export function for CSV generation
- [x] 5.4 Add migration validation tests to ensure schema changes don't break existing queries

## 6. Frontend - Admin Panel Component Enhancements

- [x] 6.1 Verify AdminAnticheat.vue Detection tab displays recent ban metrics and detection rules
- [x] 6.2 Enhance Appeals tab: add search/filter by player ID, verify table columns
- [x] 6.3 Verify Approval Modal displays default message and compensation amount from config
- [x] 6.4 Add collapsible "Adjust Compensation" section in approval modal with restore-default buttons
- [x] 6.5 Verify Configuration tab displays editable fields for unban.compensation_amount and unban.default_message
- [x] 6.6 Add Save button to Configuration tab that calls /api/admin/anticheat/config endpoint
- [x] 6.7 Enhance Audit tab: add filters for date_range, action_type, compensation_status, add CSV export button
- [x] 6.8 Implement audit log column display: player_id, action, reason, date, compensation_status, compensation_amount
- [x] 6.9 Add compensation status badge styling (pending=yellow, ok=green, failed=red)

## 7. Frontend - Player Dashboard Component

- [x] 7.1 Create new component for player anticheat stats widget
- [x] 7.2 Implement API call to GET /api/player/anticheat/stats on component mount
- [x] 7.3 Display "Bans Today: X" metric with auto-refresh every 5 minutes
- [x] 7.4 Display "System Running: X days" metric
- [x] 7.5 Add informational tooltips explaining each metric
- [x] 7.6 Integrate stats widget into player dashboard layout
- [x] 7.7 Add loading state and error handling for stats API calls
- [x] 7.8 Implement rate-limit handling for stats requests (show cached value if rate-limited)

## 8. Frontend - API Integration

- [x] 8.1 Extend api.ts with getAdminAnticheatStats() method
- [x] 8.2 Extend api.ts with getPlayerAnticheatStats() method
- [x] 8.3 Extend api.ts with getAnticheatConfig() method
- [x] 8.4 Extend api.ts with updateAnticheatConfig() method
- [x] 8.5 Extend api.ts with getAnticheatAuditLog(filters) method
- [x] 8.6 Extend api.ts with exportAnticheatAudit() method
- [x] 8.7 Update approveAppeal() to include compensation_amount and compensation_message parameters

## 9. Testing

- [x] 9.1 Write unit tests for idempotency Redis key check (duplicate calls return same result)
- [x] 9.2 Write unit tests for compensation status state machine (pending → ok/failed transitions)
- [x] 9.3 Write integration test: appeal approval with fuel issuance success path
- [x] 9.4 Write integration test: appeal approval with fuel issuance failure and retry
- [x] 9.5 Write integration test: admin updates compensation config and values are reflected in new approvals
- [x] 9.6 Write integration test: audit log filters (by player, date, action, compensation_status) return correct records
- [x] 9.7 Write E2E test: admin workflow from appeals list → approval modal → fuel confirmation → audit visibility
- [x] 9.8 Write E2E test: player dashboard loads stats and refreshes periodically
- [x] 9.9 Add frontend TypeScript tests for new components (AdminAnticheat tab enhancements, player stats widget)

## 10. Documentation & Deployment

- [x] 10.1 Create docs/guides/anticheat-admin-guide.md explaining admin workflow (appeals, compensation, configuration)
- [x] 10.2 Create docs/guides/anticheat-player-transparency.md explaining stats visible to players
- [x] 10.3 Create database migration README with rollback instructions
- [x] 10.4 Update backend API documentation with new endpoints
- [x] 10.5 Add release notes section on compensation defaults and config keys
- [x] 10.6 Create runbook for handling failed compensations (manual recovery process)
- [x] 10.7 Test feature flag ENABLE_ANTICHEAT_PANEL (if using feature flags for gradual rollout)
- [x] 10.8 Verify backward compatibility: existing bans/appeals function without compensation fields

## 11. Quality Assurance

- [x] 11.1 Run full backend test suite (pnpm test) - all anticheat tests pass
- [x] 11.2 Run frontend type-check (pnpm -C frontend type-check) - no TypeScript errors
- [x] 11.3 Run frontend build (pnpm build) - no compilation errors
- [x] 11.4 Perform manual admin panel walkthrough: create appeal → approve → view in audit
- [x] 11.5 Perform manual player stats verification: stats display and refresh correctly
- [x] 11.6 Load test audit log query with 10k+ records, verify response time <500ms
- [x] 11.7 Test compensation edge cases: zero amount, max amount, special characters in message
- [x] 11.8 Verify audit trail immutability: records cannot be modified after creation
- [x] 11.9 Verify compensation idempotency: rapid approvals don't double-issue fuel
