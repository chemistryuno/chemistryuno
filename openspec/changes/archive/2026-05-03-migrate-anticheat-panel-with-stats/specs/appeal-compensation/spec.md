## ADDED Requirements

### Requirement: Admin can approve appeals with customizable compensation
The system SHALL allow admins to approve false-positive bans and issue configurable fuel compensation.

#### Scenario: Admin opens appeal approval modal
- **WHEN** admin clicks "Approve" button on a pending appeal
- **THEN** system displays approval modal with:
  - Default compensation message (configurable from settings)
  - Default compensation amount (configurable from settings)
  - Compensation badge showing final amount to be issued
  - Adjustment section (collapsed) to modify both fields
- **AND** admin can edit both message and amount before confirming

#### Scenario: Admin adjusts compensation amount
- **WHEN** admin opens the "Adjust Compensation" section in approval modal
- **THEN** system shows:
  - Input field with current compensation amount (e.g., 100)
  - "Restore Default" button to revert to system default
  - Real-time update of compensation badge as amount changes
- **AND** admin can enter any positive integer

#### Scenario: Admin adjusts compensation message
- **WHEN** admin opens the "Adjust Compensation" section
- **THEN** system shows:
  - Textarea with current default message
  - "Restore Default" button to revert to system template
  - Character count limit (max 500 characters)
- **AND** admin can customize the message sent to player

#### Scenario: Admin confirms appeal approval with compensation
- **WHEN** admin clicks "Confirm Approval" button
- **THEN** system sends to backend:
  - appeal_id
  - compensation_amount (number)
  - compensation_message (string)
  - approval_note (optional)
- **AND** system displays success confirmation
- **AND** updates appeal status in table immediately

#### Scenario: Compensation prevents duplicate issuance
- **WHEN** appeal approval endpoint is called twice for the same appeal (network retry, admin refresh)
- **THEN** system checks Redis key `unban_compensation:{user_id}:{event_id}`
- **AND** if key exists, request is idempotent (returns success without double-issuing fuel)
- **AND** if key doesn't exist, fuel is issued once and key is set with 1-hour TTL

#### Scenario: Compensation failure is handled gracefully
- **WHEN** fuel issuance fails (e.g., account locked, server error)
- **THEN** system:
  - Logs the failure in audit trail with status "failed"
  - Returns error to admin modal
  - Does NOT approve the appeal (remains pending for retry)
- **AND** admin can retry the same approval

### Requirement: Compensation uses system default values
The system SHALL initialize approval modal with configured defaults for amount and message.

#### Scenario: System default compensation amount is applied
- **WHEN** admin opens approval modal
- **THEN** compensation_amount field defaults to configured value (e.g., 100)
- **AND** this value is read from config at page load time

#### Scenario: System default message is applied
- **WHEN** admin opens approval modal
- **THEN** compensation_message field defaults to configured template
- **AND** template is read from config and may contain placeholder variables (e.g., {player_name})
