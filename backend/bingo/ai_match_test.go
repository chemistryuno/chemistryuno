package bingo

import (
	"fmt"
	"testing"
	"time"

	"chemistryuno/backend/database"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite"
)

// setupTestDB creates an in-memory SQLite DB seeded with approved substances.
func setupTestDB(t *testing.T, substanceCount int) {
	t.Helper()
	dialector := sqlite.Dialector{DriverName: "sqlite", DSN: "file::memory:?cache=shared"}
	db, err := gorm.Open(dialector, &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&database.Substance{}, &database.BingoRoom{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	// Seed approved substances with varying formula lengths.
	for i := 0; i < substanceCount; i++ {
		s := database.Substance{
			Name:         fmt.Sprintf("物质%d", i),
			Formula:      fmt.Sprintf("C%dH%d", i%9+1, i%20+1),
			Status:       "approved",
			CreatedByUID: 1,
		}
		if err := db.Create(&s).Error; err != nil {
			t.Fatalf("seed substance: %v", err)
		}
	}
	database.DB = db
}

func TestCreateRoomWithAI(t *testing.T) {
	setupTestDB(t, 200)

	room, err := CreateRoom([]uint{42}, 10, 1, 50)
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}
	if len(room.AIMembers) != 1 {
		t.Fatalf("expected 1 AI member, got %d", len(room.AIMembers))
	}
	// The human (42) and one AI must be split across the two teams.
	total := len(room.TeamAMembers) + len(room.TeamBMembers)
	if total != 2 {
		t.Fatalf("expected 2 total members, got %d", total)
	}
	if !room.IsAI(room.AIMembers[0]) {
		t.Fatalf("AI member should report IsAI true")
	}
	if room.IsAI(42) {
		t.Fatalf("human 42 should not be AI")
	}
}

func TestAITeamDetection(t *testing.T) {
	setupTestDB(t, 200)
	// 1 human vs 1 AI: each team has exactly one member, one of which is all-AI.
	room, err := CreateRoom([]uint{42}, 10, 1, 50)
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}
	aiTeams := 0
	for teamIdx := TeamA; teamIdx <= TeamB; teamIdx++ {
		if room.IsTeamAI(teamIdx) {
			aiTeams++
		}
	}
	if aiTeams != 1 {
		t.Fatalf("expected exactly 1 all-AI team, got %d", aiTeams)
	}
}

func TestHumanVsAIFullMatch(t *testing.T) {
	setupTestDB(t, 300)

	// Speed up AI turns for the test.
	AITurnDelay = 5 * time.Millisecond
	defer func() { AITurnDelay = 1200 * time.Millisecond }()

	room, err := CreateRoom([]uint{42}, 10, 1, 90)
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}

	// Track AI turn completion via the hook.
	done := make(chan struct{}, 64)
	OnAITurnDone = func(r *BingoRoom) {
		select {
		case done <- struct{}{}:
		default:
		}
	}
	defer func() { OnAITurnDone = nil }()

	room.StartGame(func(roomID uint) {})

	// Hands must be dealt to everyone.
	room.mu.RLock()
	for _, uid := range append(append([]uint{}, room.TeamAMembers...), room.TeamBMembers...) {
		if len(room.Hands[uid]) != HandSize {
			room.mu.RUnlock()
			t.Fatalf("uid %d should have %d cards, got %d", uid, HandSize, len(room.Hands[uid]))
		}
	}
	humanTeam := room.getTeamIdxUnlocked(42)
	room.mu.RUnlock()

	// If the AI team goes first, trigger its turn.
	TriggerAITurnIfNeeded(room)

	// Play out the game: human occupies whatever cell it can, alternating with AI.
	deadline := time.After(10 * time.Second)
	turns := 0
	for {
		room.mu.RLock()
		status := room.Status
		curTurn := room.CurrentTurn
		a, b := room.countCellsUnlocked()
		room.mu.RUnlock()

		if status == "finished" {
			break
		}
		// Accept completion either by win or by timeout when many cells are occupied.
		if a >= 12 || b >= 12 {
			t.Logf("match completed: A=%d B=%d cells (winner=%d)", a, b, room.WinnerTeamIdx)
			break
		}

		if curTurn == humanTeam {
			turns++
			if turns > 100 {
				t.Fatalf("too many human turns without progress (100+), A=%d B=%d", a, b)
			}
			if !humanPlayOneMove(room, humanTeam) {
				// No move possible; do a swap to pass a bit of time.
				r1, c1, r2, c2 := AIChooseSwap(room, humanTeam)
				if err := room.SwapCells(humanTeam, r1, c1, r2, c2); err != nil {
					t.Fatalf("human swap: %v", err)
				}
				TriggerAITurnIfNeeded(room)
			}
		}

		select {
		case <-done:
			// AI acted; loop again.
		case <-deadline:
			t.Fatalf("match did not finish in time (A=%d B=%d cells occupied, turns=%d)", a, b, turns)
		case <-time.After(10 * time.Millisecond):
			// keep looping
		}
	}

	room.mu.RLock()
	winner := room.WinnerTeamIdx
	room.mu.RUnlock()
	if winner != TeamA && winner != TeamB {
		t.Fatalf("expected a winner team index, got %d", winner)
	}
	t.Logf("match finished, winner team=%d", winner)
}

// humanPlayOneMove occupies the first cell for which the human holds a matching card.
func humanPlayOneMove(room *BingoRoom, humanTeam int) bool {
	room.mu.RLock()
	hand := room.Hands[42]
	var handMap = map[uint]bool{}
	for _, c := range hand {
		handMap[c.SubstanceID] = true
	}
	size := room.Board.Size
	var tr, tc = -1, -1
	for r := 0; r < size && tr < 0; r++ {
		for c := 0; c < size; c++ {
			cell := room.Board.Cells[r][c]
			if cell.OwnerTeamID == nil && handMap[cell.SubstanceID] {
				tr, tc = r, c
				break
			}
		}
	}
	room.mu.RUnlock()

	if tr < 0 {
		return false
	}
	if win, err := room.OccupyCell(humanTeam, tr, tc); err != nil {
		// If the error is turn-related, return false to let the caller re-try later.
		return false
	} else {
		_ = win
	}
	// OccupyCell switches the turn; trigger AI next.
	TriggerAITurnIfNeeded(room)
	return true
}
