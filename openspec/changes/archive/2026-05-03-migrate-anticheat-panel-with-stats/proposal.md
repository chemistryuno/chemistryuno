## Why

The anticheat system lacks a unified admin interface for managing bans/unbans and audit trails. Additionally, players have no visibility into their own anticheat status or system-wide statistics. Creating a centralized anticheat panel with player-facing statistics improves both administrative efficiency and player trust in the system.

## What Changes

- Consolidate all ban/unban/appeal functionality into a dedicated anticheat admin panel
- Provide players with real-time anticheat statistics (bans today, system uptime)
- Create audit trail display for appeal resolution history
- Add compensation tracking for unban cases
- Enable admins to adjust compensation policies (amount, message) without code changes

## Capabilities

### New Capabilities

- `anticheat-admin-panel`: Unified dashboard for managing player bans, appeals, and compensation
- `player-anticheat-stats`: Real-time statistics displayed to players about bans today and system operational days
- `appeal-compensation`: Workflow for approving appeals with configurable compensation (fuel amount and message)
- `anticheat-audit-trail`: Comprehensive audit log showing all ban/unban actions and their outcomes

### Modified Capabilities

- `anticheat-system`: Requires APIs to expose ban statistics and system uptime information to frontend

## Impact

- **Frontend**: New admin panel page (`AdminAnticheat.vue`), player stats component, modal workflows
- **Backend**: New API endpoints for stats, audit logs, compensation handling; database schema additions for compensation tracking
- **Database**: Potentially new audit fields (compensation_amount, compensation_status, compensation_note, compensation_date)
- **Configuration**: New config keys for default compensation (amount, message template)
