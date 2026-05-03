## ADDED Requirements

### Requirement: Backend exposes anticheat statistics API
The system SHALL provide endpoints for querying anticheat metrics.

#### Scenario: Admin queries ban statistics
- **WHEN** admin opens Detection tab
- **THEN** system calls `/api/admin/anticheat/stats` endpoint
- **AND** receives response with ban counts by period and detection rules

#### Scenario: Player queries public anticheat stats
- **WHEN** player loads dashboard
- **THEN** client calls `GET /api/player/anticheat/stats`
- **AND** receives JSON: `{ "bans_today": N, "system_uptime_days": X }`
- **AND** endpoint returns cached result (max 5 minutes old)

#### Scenario: System tracks anticheat uptime
- **WHEN** anticheat system starts
- **THEN** system records startup timestamp in configuration or cache key
- **AND** can calculate days elapsed since last startup for uptime calculation
- **AND** admin can manually reset uptime counter if system rebooted

### Requirement: Backend exposes approval endpoint accepting compensation parameters
The system SHALL handle appeal approvals with embedded compensation data.

#### Scenario: Admin submits appeal approval with compensation
- **WHEN** admin calls `POST /api/admin/anticheat/appeals/{appealId}/approve`
- **THEN** endpoint accepts JSON body:
  ```json
  {
    "note": "optional approval notes",
    "compensation_amount": 100,
    "compensation_message": "message to send player"
  }
  ```
- **AND** processes atomically: approval + compensation + audit logging
- **AND** returns 200 on success with updated appeal status

#### Scenario: Compensation fails with partial failure handling
- **WHEN** approval succeeds but fuel issuance fails
- **THEN** system:
  - Returns 200 to frontend (appeal is marked approved)
  - Records compensation_status: "failed" in audit table
  - Stores failure reason for admin retry
  - Does NOT delete the approval (allows manual recovery)

### Requirement: Configuration system supports compensation defaults
The system SHALL provide config keys for customizing compensation behavior.

#### Scenario: Admin updates compensation settings
- **WHEN** admin modifies config in Configuration tab
- **THEN** backend stores changes to:
  - `unban.compensation_amount` (integer, default 100)
  - `unban.default_message` (string, max 500 chars)
  - `unban.enabled` (boolean, whether compensation is active)
- **AND** changes take effect immediately

#### Scenario: Frontend reads compensation config on load
- **WHEN** AdminAnticheat.vue component mounts
- **THEN** system calls `/api/admin/anticheat/config` to fetch current values
- **AND** pre-populates modal fields with live config
- **AND** caches config for 60 seconds to avoid excessive API calls

### Requirement: Audit records support compensation tracking
The system SHALL persist compensation metadata in audit trail.

#### Scenario: Audit table includes compensation columns
- **WHEN** database schema is migrated
- **THEN** audit table includes columns:
  - `compensation_amount` (INT, nullable)
  - `compensation_status` (ENUM: pending/ok/failed, nullable)
  - `compensation_message` (TEXT, nullable)
  - `compensation_note` (TEXT, nullable)
  - `compensation_date` (DATETIME, nullable)
  - `approval_note` (TEXT, nullable)

#### Scenario: Audit query includes compensation data
- **WHEN** admin queries audit log via `/api/admin/anticheat/audit`
- **THEN** response includes compensation fields
- **AND** supports filtering by compensation_status
- **AND** CSV export includes all compensation columns
