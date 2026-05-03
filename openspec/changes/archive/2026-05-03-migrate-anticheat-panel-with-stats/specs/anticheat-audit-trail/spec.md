## ADDED Requirements

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
