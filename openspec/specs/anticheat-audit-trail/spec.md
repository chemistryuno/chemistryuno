# anticheat-audit-trail Specification

## Purpose
TBD - created by archiving change migrate-anticheat-panel-with-stats. Update Purpose after archive.
## Requirements
### Requirement: Audit trail records all ban/unban actions with compensation metadata
The system SHALL maintain an immutable historical ledger of anticheat enforcement actions.

#### Scenario: Audit record is created on ban
- **WHEN** anticheat system issues a ban
- **THEN** system creates audit record with fields:
  - player_id
  - action: "ban"
  - reason: detection rule name
  - created_at: timestamp
  - status: "active"
  - compensation_amount: NULL (until appeal approved)
  - compensation_status: NULL
  - compensation_note: NULL
  - compensation_date: NULL

#### Scenario: Audit record is updated on appeal approval
- **WHEN** admin approves an appeal with compensation
- **THEN** system updates audit record with:
  - action: "unban"
  - approval_note: admin's notes
  - compensation_amount: approved amount
  - compensation_status: "pending" (before fuel issued)
  - compensation_message: message sent to player
  - compensation_date: timestamp of approval

#### Scenario: Audit record reflects compensation outcome
- **WHEN** compensation fuel is successfully issued
- **THEN** system updates compensation_status to "ok"
- **AND** stores confirmation timestamp

#### Scenario: Audit record reflects compensation failure
- **WHEN** compensation issuance fails
- **THEN** system updates compensation_status to "failed"
- **AND** stores failure reason in compensation_note
- **AND** record remains queryable for recovery attempts

### Requirement: Audit trail is queryable and filterable
The system SHALL provide admin interface to search audit history.

#### Scenario: Admin filters audit by player
- **WHEN** admin enters player ID in audit log filter
- **THEN** system displays all ban/unban records for that player
- **AND** includes compensation details

#### Scenario: Admin filters audit by date range
- **WHEN** admin selects start and end dates
- **THEN** system displays records created within that range
- **AND** supports sorting by date ascending/descending

#### Scenario: Admin filters audit by action type
- **WHEN** admin selects "Bans Only" or "Unbans Only"
- **THEN** system filters to show only matching action type
- **AND** maintains other active filters

#### Scenario: Admin filters audit by compensation status
- **WHEN** admin selects compensation status: "Pending", "Ok", "Failed"
- **THEN** system displays only unbans matching that status
- **AND** allows multiple selections (OR logic)

### Requirement: Audit export includes full compensation context
The system SHALL support exporting audit logs with all relevant fields.

#### Scenario: Admin exports filtered audit results
- **WHEN** admin clicks "Export" after applying filters
- **THEN** system downloads CSV file with columns:
  - player_id, action, reason, created_at, approval_note,
  - compensation_amount, compensation_status, compensation_message, compensation_date
- **AND** export respects applied filters
- **AND** exports are timestamped for traceability

### Requirement: Audit trail records replay evidence references
The audit trail SHALL preserve replay evidence references for anticheat review, punishment, punishment decision change, appeal, and protected replay cleanup events.

#### Scenario: Audit record is created for anticheat review
- **WHEN** an administrator reviews or processes an anticheat detection
- **THEN** the audit record includes the detection ID, room ID, replay ID, and primary replay evidence anchor
- **AND** the record remains queryable even if the risk score policy later changes

#### Scenario: Audit record is created for punishment decision change
- **WHEN** an administrator changes a processed detection's punishment decision
- **THEN** the audit record stores the previous decision, new decision, reason, admin ID, and replay evidence reference
- **AND** the replay evidence reference points back to the original detection evidence chain

#### Scenario: Audit record is created for protected replay cleanup attempt
- **WHEN** replay cleanup or manual admin action attempts to clear a replay protected by anticheat evidence
- **THEN** the audit record identifies the protected replay and the evidence type preventing deletion
- **AND** records whether the attempt was skipped or rejected

### Requirement: Audit export includes replay evidence context
The audit export SHALL include replay evidence context for anticheat records.

#### Scenario: Admin exports audit records
- **WHEN** admin exports anticheat audit records
- **THEN** the export includes room ID, replay ID, event index or event ID, evidence precision, and action summary when present
- **AND** the export preserves enough fields to reconstruct an admin replay navigation link

