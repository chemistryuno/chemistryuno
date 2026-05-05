# anticheat-system Specification

## Purpose
TBD - created by archiving change migrate-anticheat-panel-with-stats. Update Purpose after archive.
## Requirements
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

### Requirement: Risk indicators map to replay evidence anchors
The anticheat system SHALL persist replay evidence anchors for each risk score and each contributing risk indicator.

#### Scenario: Risk score is created with operation evidence
- **WHEN** the anticheat engine creates a risk score for a player in a room
- **THEN** the risk record includes a primary replay evidence anchor
- **AND** each indicator detail includes its own replay evidence anchor or a documented room-level evidence anchor
- **AND** the primary anchor points to the highest-risk operation when multiple operations contributed

#### Scenario: Risk score includes multiple suspicious operations
- **WHEN** several replay operations contribute to one player's risk score
- **THEN** the system stores all related replay evidence anchors in the indicator details
- **AND** the detection detail API returns those anchors in event order or contribution order

### Requirement: Anticheat API exposes replay evidence for review
The anticheat system SHALL expose replay evidence anchors through admin-facing detection, report, sanction, appeal, and audit APIs.

#### Scenario: Admin queries detection detail
- **WHEN** admin calls the detection detail endpoint for a risk record
- **THEN** the response includes the primary replay anchor
- **AND** includes indicator-level replay anchors with score contribution and explanation
- **AND** includes a replay navigation URL or enough route parameters for the frontend to build one

#### Scenario: Appeal entry reads anticheat evidence rooms
- **WHEN** the player appeal entry endpoint builds the locked room list
- **THEN** it uses rooms from the latest anticheat evidence chain
- **AND** preserves the replay evidence references for admin-side appeal review

### Requirement: Reports contribute to risk only with evidence binding
The anticheat system SHALL include player report signals in risk scoring only when the report has a valid evidence binding.

#### Scenario: Valid report contributes to risk
- **WHEN** a player report has a validated replay evidence anchor
- **THEN** the report can contribute to the risk score according to configured report weight and deduplication policy
- **AND** the contribution records the report anchor and report source summary

#### Scenario: Unbound report is stored outside risk scoring
- **WHEN** a report does not have a valid replay evidence anchor
- **THEN** the system may store it as general feedback or moderation context
- **AND** it MUST NOT increase the anticheat risk score

### Requirement: Anticheat system is testable as a discovery-to-disposal flow
The anticheat system SHALL expose stable behavior that can be verified by automated tests from cheat discovery through punishment and appeal entry.

#### Scenario: Automated test creates full anticheat flow
- **WHEN** an automated test initializes the anticheat system with deterministic test configuration
- **THEN** the test can create a high-risk detection
- **AND** process the detection through admin APIs
- **AND** verify sanction, audit, ban status, and appeal entry outcomes without external services

#### Scenario: Test fixture does not depend on real time instability
- **WHEN** the anticheat flow test builds its cheating fixture
- **THEN** the fixture uses fixed timestamps, fixed risk inputs, and deterministic thresholds
- **AND** the expected result does not depend on wall-clock race conditions

