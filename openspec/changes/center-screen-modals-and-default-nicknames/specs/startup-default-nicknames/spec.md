## ADDED Requirements

### Requirement: Server startup repairs missing player nicknames
The system SHALL check for users with missing or blank nicknames every time the server starts and assign a random nickname to each affected user.

#### Scenario: Server starts with users missing nicknames
- **WHEN** the server starts and one or more user records have an empty or whitespace-only nickname
- **THEN** the system assigns each affected user a generated random nickname
- **AND** the repaired nicknames are persisted before normal player-facing services accept interactions

#### Scenario: Server starts with all users already named
- **WHEN** the server starts and every user has a non-blank nickname
- **THEN** the system does not modify any existing nickname
- **AND** startup continues normally

#### Scenario: Server restarts after previous repair
- **WHEN** the server starts again after missing nicknames were repaired
- **THEN** previously assigned nicknames are preserved
- **AND** the repair pass remains idempotent

### Requirement: Startup-generated nicknames are valid and unique
The system SHALL generate startup fallback nicknames that satisfy existing nickname validation rules and do not duplicate another user's nickname.

#### Scenario: Generated nickname collides with an existing nickname
- **WHEN** a generated nickname already belongs to another user
- **THEN** the system generates another candidate or uses a deterministic unique fallback
- **AND** the user is assigned a nickname that does not duplicate the existing nickname

#### Scenario: Startup repair completes
- **WHEN** the startup nickname repair pass finishes
- **THEN** the system logs how many users were checked or repaired
- **AND** startup continues unless the underlying database operation fails in a way that prevents reliable repair
