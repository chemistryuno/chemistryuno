# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Chemistry UNO** is a multiplayer online card game that combines classic UNO gameplay with chemistry reaction logic. Players must use chemical elements to synthesize substances and trigger special chemical reaction effects.

- **Backend**: Go 1.20+ with Gin framework, SQLite database
- **Frontend**: Vue 3 (Composition API) with TypeScript, Vite, Tailwind CSS 4
- **Real-time Communication**: WebSocket for multiplayer game state synchronization
- **Security**: WebAuthn (FIDO2), 2FA (TOTP), Argon2 password hashing, JWT authentication

## Development Commands

### Installation & Setup

```bash
# Initial setup (install dependencies for both frontend and backend)
pnpm run init

# Or use the Windows batch script
init.bat
```

### Running the Application

```bash
# Start both frontend and backend (production mode)
pnpm start

# Start frontend only (dev server on port 5000)
pnpm run frontend

# Start backend only (API server on port 8080)
pnpm run backend
```

### Frontend Development

```bash
cd frontend

# Install dependencies
pnpm install

# Run dev server
pnpm dev

# Type checking
pnpm type-check

# Build for production
pnpm build
```

### Backend Development

```bash
cd backend

# Install/update Go dependencies
go mod tidy

# Run backend server
go run main.go

# Build binary
go build -o chemistryuno.exe main.go
```

### Database & Cleanup

```bash
# Clean database and binaries
pnpm run clean
# This removes: backend/chemistryuno.db, backend/main.exe, backend/chemistryuno.exe

# Initialization (Manual)
pnpm run tools:init-db

# WebAuthn Credential Migration
pnpm run tools:migrate-creds
```

## Architecture Overview

### Backend Architecture

**Entry Point**: `backend/main.go`

- Initializes database at `./chemistryuno.db`
- Starts WebSocket hub for real-time communication
- Starts room monitor for inactive player management
- Configures Gin routes with CORS middleware
- Sets up three authentication levels: public, authenticated, admin

**Key Modules**:

1. **Game Engine** (`backend/game/`)
   - `chemistry.go`: Chemical formula parser - handles complex formulas like `Fe(OH)3`, `Ca(HCO3)2`. Function `parseSubstance()` uses a stack-based parser to extract element composition
   - `judge.go`: Chemistry reaction engine - implements acid-base neutralization, double displacement, metal activity series (K-Au sequence), solubility rules
   - `manager.go`: Game room lifecycle management, turn-based game state, deck configuration

2. **WebSocket System** (`backend/websocket/`)
   - `hub.go`: Central message broker for all connected clients
   - `client.go`: Per-user WebSocket connection handler with ReadPump/WritePump goroutines
   - Room-based message broadcasting for multiplayer games

3. **Authentication** (`backend/handlers/`)
   - `auth.go`: Traditional username/password login
   - `webauthn.go`: FIDO2/hardware key support for passwordless authentication
   - `2fa.go`: TOTP-based two-factor authentication
   - Password recovery supports both 2FA codes and WebAuthn keys

4. **Database** (`backend/database/db.go`)
   - SQLite with CGO bindings (modernc.org/sqlite)
   - Schema includes: users, game_rooms, substances, reactions, feedbacks, deck_configs
   - Substances and reactions support "draft" vs "approved" status for content moderation

### Frontend Architecture

**Entry Point**: `frontend/src/App.vue`

- Vue Router for page navigation
- Uses Composition API throughout

**Key Components**:

1. **API Layer** (`frontend/src/utils/api.ts`)
   - Axios-based REST API client with interceptors
   - Automatic JWT token injection via `Authorization: Bearer <token>`
   - 401 handling redirects to login page

2. **WebSocket Client** (`frontend/src/utils/websocket.ts`)
   - Singleton WebSocketService class
   - Auto-reconnection logic (max 5 attempts)
   - Message queuing when disconnected
   - Event-based message routing to listeners

3. **Game Flow**:
   - `Lobby.vue`: Room list and creation
   - `GameRoom.vue`: Active game state, card playing interface
   - Real-time updates via WebSocket messages (type: "room_update", "game_state", etc.)

4. **Security Components**:
   - `HardwareKeyModal.vue`: WebAuthn credential registration
   - `TwoFactorSetupModal.vue`: TOTP QR code generation and verification
   - `ChangePasswordModal.vue`: Password change with verification options

5. **Admin Interface** (`Admin.vue`):
   - User management (create, delete, promote, ban)
   - Substance/reaction approval workflow
   - Game history audit logs ("Reactor Logs")
   - Feedback management system

### Chemistry Logic Flow

1. **Card to Substance**: `GetSubstancesFromElements()` in `chemistry.go` takes player's hand cards and queries approved substances from DB to determine which can be formed
2. **Reaction Validation**: When player attempts to play cards, backend checks:
   - Can the cards form a valid substance? (element composition check)
   - Does the substance react with the current table state? (chemistry rules in `judge.go`)
   - Is the reaction valid per solubility/activity rules?
3. **Dynamic Reaction Database**: Unlike hardcoded reaction lists, the system uses chemical principles (acid + base → salt + water, metal activity displacement, etc.)

### Authentication Flow

**Standard Login**:

1. POST `/auth/login` with username/password
2. If 2FA enabled: returns `{requires2FA: true}`, client shows 2FA input
3. POST `/auth/2fa/verify` with code → receives JWT token

**WebAuthn Login**:

1. GET `/auth/webauthn/login/begin` → server sends challenge
2. Browser triggers authenticator (fingerprint/Yubikey)
3. POST `/auth/webauthn/login/finish` with signed response → receives JWT token

**Password Recovery**:

- Via 2FA: POST `/auth/2fa/reset-password` with username + 2FA code
- Via WebAuthn: Two-step flow similar to login, then set new password

### Middleware Chain

Routes use layered middleware in `backend/main.go`:

1. **CORS**: `middleware.CORSMiddleware()` - allows cross-origin requests from frontend
2. **Auth**: `middleware.AuthMiddleware()` - validates JWT, extracts UID/username into Gin context
3. **Role-based**:
   - `middleware.AdminMiddleware()` - requires user.role = "admin"
   - `middleware.CoWorkerMiddleware()` - requires role = "admin" OR "co-worker"

### Database Schema Notes

- **users**: Stores password hashes (Argon2), role (admin/co-worker/user), 2FA secrets, banned_until timestamp
- **substances**: formula, name, category, status (draft/approved), submitter info
- **reactions**: formula, conditions, status, group_id for batch approval
- **webauthn_credentials**: FIDO2 credential storage per user
- **feedbacks**: User feedback with status tracking and optional remove_at timestamp (auto-deleted by hourly cleanup job in main.go:179)

## Special Considerations

### CGO & GCC on Windows

The backend uses SQLite with CGO, requiring MinGW-w64 GCC >= 8.0. If you encounter `--high-entropy-va` errors during compilation, the `start.js` script sets `CGO_LDFLAGS="-g -O2"` to disable ASLR flags for compatibility with older GCC versions.

### Game Room Timeout System

`game.StartRoomMonitor()` runs a background goroutine checking for idle players. If a player doesn't act within 30 seconds during their turn, they're auto-kicked to keep games flowing. Repeated offenses result in temporary bans (stored in users.banned_until).

### WebSocket Message Protocol

Messages follow this structure:

```typescript
{
  type: string,           // e.g., "room_update", "game_state", "duel_request"
  room_id?: string,       // optional room context
  data?: any,             // payload varies by type
  error?: string          // error message if applicable
}
```

### Deck Configuration

Each game room can use:

- Default deck (global config in DB, auto-used for points mode)
- Custom user deck (stored in deck_configs table)
Deck defines element card distribution (e.g., "H": 12, "O": 12, "Au": 4, "He": 1).

### Points Mode vs Casual Mode

- **Points Mode**: Uses default deck only, affects player leaderboard, competitive
- **Casual Mode**: Allows custom decks, no ranking impact

## Common Patterns

### Adding a New API Endpoint

1. Define handler in `backend/handlers/` (e.g., `handlers.NewFeature`)
2. Register route in `backend/main.go` under appropriate group (public/auth/admin)
3. Add corresponding API function in `frontend/src/utils/api.ts`
4. Call from Vue component using the API client

### Adding a New Chemical Substance

1. User submits via frontend form → POST `/substances`
2. Substance created with status="draft"
3. Admin/Co-worker reviews in Admin panel
4. PUT `/substances/approve/:id` changes status to "approved"
5. Now available in `GetSubstancesFromElements()` queries

### WebSocket Room Broadcasting

```go
websocket.BroadcastToRoom(roomID, Message{
    Type: "game_state",
    Data: gameState,
})
```

This sends to all clients in the room (tracked in hub.rooms map).

### Frontend Route Protection

Use `beforeEnter` guards in Vue Router or check `localStorage.getItem('token')` in component setup. The API interceptor automatically redirects on 401 responses.

## Testing Notes

- No automated test suite is currently configured
- Manual testing workflow: Create room → Join with multiple browser tabs → Play cards → Verify WebSocket sync
- Test admin features by promoting a test user: PUT `/admin/users/:id/role` with role="admin"

## Default Credentials

After initial database setup:

- Username: `admin@chemistryuno.com`
- Password: `admin123`
- Role: `admin`

Change this immediately in production environments.

## Port Configuration

- Frontend dev server: **5000** (configured in `frontend/package.json` scripts)
- Backend API: **8080** (hardcoded in `backend/main.go:196`)
- Frontend proxy: Vite proxies `/api` and `/ws` to backend automatically (check `vite.config.ts` if debugging connection issues)
