## ADDED Requirements

### Requirement: Players can view real-time anticheat statistics
The system SHALL display current anticheat metrics on the player dashboard.

#### Scenario: Player views bans today
- **WHEN** player opens their dashboard
- **THEN** system displays a widget showing "Bans Today: X" (count of total player bans issued in last 24 hours)
- **AND** updates in real-time as new bans are issued

#### Scenario: Player views system uptime
- **WHEN** player opens their dashboard
- **THEN** system displays a widget showing "System Running: X days" (anticheat system uptime since last reset)
- **AND** includes tooltip explaining this is a transparency metric

#### Scenario: Stats are accessible without admin role
- **WHEN** a non-admin player loads the dashboard
- **THEN** system calls `/api/player/anticheat/stats` endpoint
- **AND** returns stats regardless of player role
- **AND** does not expose admin-only details (policy rules, detection methods)

#### Scenario: Stats cache is reasonably fresh
- **WHEN** system generates stats responses
- **THEN** bans_today count reflects last 5 minutes of activity
- **AND** system_uptime_days is accurate to within 1 day
- **AND** response time is <200ms

### Requirement: Stats endpoint returns standardized JSON
The system SHALL expose anticheat stats via a public API endpoint.

#### Scenario: Stats endpoint is queried
- **WHEN** client calls `GET /api/player/anticheat/stats`
- **THEN** system returns HTTP 200 with JSON: `{ "bans_today": N, "system_uptime_days": X }`
- **AND** endpoint requires player authentication
- **AND** endpoint is rate-limited to 1 request per second per player
