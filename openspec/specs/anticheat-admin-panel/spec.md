# anticheat-admin-panel Specification

## Purpose
TBD - created by archiving change migrate-anticheat-panel-with-stats. Update Purpose after archive.
## Requirements
### Requirement: Admin can access unified anticheat management panel
The system SHALL provide a dedicated admin interface consolidating all anticheat management functions into a single dashboard with tabbed navigation.

#### Scenario: Admin navigates to anticheat panel
- **WHEN** admin with anticheat permissions accesses the admin dashboard
- **THEN** admin can navigate to "Anticheat Management" section
- **AND** the panel displays 4 tabs: Detection, Appeals, Configuration, Audit

#### Scenario: Admin views current detection metrics
- **WHEN** admin opens the Detection tab
- **THEN** system displays active detection policies and recent triggers
- **AND** shows count of bans issued in current period

### Requirement: Admin can manage appeal queue
The system SHALL display pending appeals in a table with filtering and bulk operation support.

#### Scenario: Admin views pending appeals
- **WHEN** admin opens the Appeals tab
- **THEN** system displays list of pending appeals with columns: Player ID, Reason, Date, Status
- **AND** list is sortable by date or status

#### Scenario: Admin searches appeals
- **WHEN** admin enters player ID in search box
- **THEN** system filters appeals to matching player only
- **AND** displays zero results if no matches found

### Requirement: Admin can access configuration management
The system SHALL provide an interface to adjust anticheat policies without code changes.

#### Scenario: Admin views configuration options
- **WHEN** admin opens the Configuration tab
- **THEN** system displays editable fields for default compensation settings
- **AND** shows current values for amount, message template, and policy flags

#### Scenario: Admin updates configuration
- **WHEN** admin modifies a config value and clicks Save
- **THEN** system persists the change
- **AND** displays success confirmation

### Requirement: Admin can view audit trail
The system SHALL provide read-only historical ledger of all ban/unban actions and outcomes.

#### Scenario: Admin views audit log
- **WHEN** admin opens the Audit tab
- **THEN** system displays chronological log with columns: Player ID, Action, Date, Reason, Compensation Status, Amount
- **AND** log is filterable by date range and action type

#### Scenario: Admin exports audit log
- **WHEN** admin clicks Export button
- **THEN** system downloads audit data as CSV file
- **AND** includes all visible columns and rows

### Requirement: Admin can jump from suspicious point to replay operation
The admin anticheat panel SHALL allow administrators to open the exact replay operation for each suspicious point.

#### Scenario: Admin opens suspicious point replay
- **WHEN** admin opens `http://localhost:5000/admin/anticheat` and selects a detection detail
- **THEN** each suspicious point displays its room ID, replay event position, operation type, player, timestamp, score contribution, and explanation
- **AND** each operation-level suspicious point provides an action to open the replay at that operation

#### Scenario: Replay anchor is room-level only
- **WHEN** a suspicious point only has room-level evidence
- **THEN** the panel displays that precise operation positioning is unavailable
- **AND** the replay action opens the room replay without pretending to seek to a specific operation

### Requirement: Admin report review shows replay evidence
The admin anticheat panel SHALL display replay evidence for player reports that contribute to anticheat risk.

#### Scenario: Admin reviews report contribution
- **WHEN** admin views the indicator details for a risk record with report contribution
- **THEN** the panel displays the report reason, report contribution, deduplication status, and replay anchor
- **AND** admin can open the reported replay point when the anchor is operation-level

### Requirement: Admin processing preserves replay evidence
The admin anticheat panel SHALL keep replay evidence visible after review and punishment decision changes.

#### Scenario: Admin processes detection
- **WHEN** admin processes a detection entry
- **THEN** the processed detail view still displays the original replay evidence anchors
- **AND** any later punishment decision change keeps the same evidence chain attached unless a new evidence note is appended

#### Scenario: Admin changes punishment decision
- **WHEN** admin changes the punishment decision for a processed detection
- **THEN** the panel requires a reason
- **AND** the request includes the detection evidence reference so the audit log can preserve the replay evidence chain

### Requirement: Admin anticheat panel has disposal flow test coverage
The admin anticheat panel SHALL have automated tests covering detection detail display and punishment disposal actions.

#### Scenario: Admin panel renders detection evidence
- **WHEN** the admin panel test provides a detection record with risk score, indicators, report contribution, replay evidence, suggested action, and review status
- **THEN** the panel renders the detection details needed for admin review
- **AND** exposes the action controls for processing and permitted punishment decision changes

#### Scenario: Admin panel rejects cancellation interaction
- **WHEN** the admin panel test attempts to submit a cancellation-style punishment decision for a processed detection
- **THEN** the UI does not treat it as a successful allowed disposal
- **AND** displays the backend rejection when the API rejects the request

### Requirement: Admin can preview detection evidence inline

The admin anticheat panel SHALL provide an inline evidence preview for a detection's suspicious points without leaving the panel.

#### Scenario: Admin previews suspicious point evidence
- **WHEN** admin opens a detection detail and selects a suspicious point with an operation-level anchor
- **THEN** the panel SHALL display an inline preview of the operation context (operation type, player perspective, timestamp, score contribution, explanation)
- **AND** the panel SHALL still offer the full replay navigation action

#### Scenario: Room-level evidence preview is honest about precision
- **WHEN** a suspicious point only has a room-level anchor
- **THEN** the inline preview SHALL indicate that operation-level positioning is unavailable
- **AND** SHALL NOT fabricate an operation-level preview

### Requirement: Admin can batch-process detections and appeals

The admin anticheat panel SHALL support selecting multiple detection or appeal entries and applying a permitted action to the selection in one operation.

#### Scenario: Admin batch-processes detections
- **WHEN** admin selects multiple detection entries and applies a permitted disposal action
- **THEN** the system SHALL apply the action to each selected entry
- **AND** SHALL report per-entry success or failure
- **AND** SHALL require a reason where the single-entry flow requires one

#### Scenario: Batch action preserves per-entry audit and evidence
- **WHEN** a batch action is applied to multiple entries
- **THEN** the system SHALL record an individual audit log entry per affected detection
- **AND** SHALL keep each detection's evidence chain attached

### Requirement: Admin can test detection rules offline

The admin anticheat panel SHALL provide a rule-testing function that re-runs a draft or current detection configuration against historical or constructed detection samples in an isolated context without affecting live player state.

#### Scenario: Admin runs a rule test
- **WHEN** admin submits a draft configuration to the rule-testing function with a set of sample detections
- **THEN** the system SHALL compute resulting scores and sanction tiers in an isolated context
- **AND** SHALL display the hit distribution and tier changes compared to the live configuration

#### Scenario: Rule test does not affect live state
- **WHEN** a rule test executes
- **THEN** the system SHALL NOT change any player's risk record, sanction, or ban state
- **AND** SHALL NOT write to the production detection records

