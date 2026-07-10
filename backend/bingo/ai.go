package bingo

import (
	"math/rand"
	"time"
)

// AISwapMove represents a potential swap move evaluated by the AI.
type AISwapMove struct {
	R1, C1, R2, C2 int
	Score          int
}

// OnAITurnDone is called after the AI finishes its turn. Set by the handler package
// to broadcast room state updates. Must be safe to call from goroutines.
var OnAITurnDone func(room *BingoRoom)

// AITurnDelay is the pause before the AI acts, so players can see the turn change.
// Overridable in tests.
var AITurnDelay = 1200 * time.Millisecond

// TriggerAITurnIfNeeded checks if the current team is AI-controlled and spawns a
// goroutine to execute the AI turn after a short delay.
func TriggerAITurnIfNeeded(room *BingoRoom) {
	room.mu.RLock()
	if room.Status != "playing" {
		room.mu.RUnlock()
		return
	}
	teamIdx := room.CurrentTurn
	isAI := room.IsTeamAI(teamIdx)
	room.mu.RUnlock()

	if !isAI {
		return
	}

	go func() {
		time.Sleep(AITurnDelay)
		executeAITurn(room, teamIdx)
	}()
}

func executeAITurn(room *BingoRoom, teamIdx int) {
	room.mu.RLock()
	if room.Status != "playing" || room.CurrentTurn != teamIdx {
		room.mu.RUnlock()
		return
	}
	room.mu.RUnlock()

	// Decide: occupy if AI has a matching hand card for a valuable cell, else swap.
	occupied := tryAIOccupy(room, teamIdx)
	if !occupied {
		r1, c1, r2, c2 := AIChooseSwap(room, teamIdx)
		_ = room.SwapCells(teamIdx, r1, c1, r2, c2)
	}

	if OnAITurnDone != nil {
		OnAITurnDone(room)
	}

	// After AI acts, the turn switches. Check if next turn is also AI.
	TriggerAITurnIfNeeded(room)
}

// tryAIOccupy attempts to occupy a cell using the AI's hand cards.
// Returns true if a cell was occupied.
func tryAIOccupy(room *BingoRoom, teamIdx int) bool {
	room.mu.RLock()
	// Find an AI member on this team to use their hand.
	var aiUID uint
	members := room.TeamAMembers
	if teamIdx == TeamB {
		members = room.TeamBMembers
	}
	for _, m := range members {
		if room.IsAI(m) {
			aiUID = m
			break
		}
	}
	if aiUID == 0 {
		room.mu.RUnlock()
		return false
	}
	hand := room.Hands[aiUID]
	if len(hand) == 0 {
		room.mu.RUnlock()
		return false
	}

	// Build a set of substance IDs in hand for quick lookup.
	handSet := make(map[uint]bool)
	for _, card := range hand {
		handSet[card.SubstanceID] = true
	}

	// Score unoccupied cells by line-progress contribution and find one that matches hand.
	size := room.Board.Size
	type candidate struct {
		row, col int
		score    int
	}
	var candidates []candidate
	ownerVal := uint(teamIdx + 1)

	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			cell := room.Board.Cells[row][col]
			if cell.OwnerTeamID != nil {
				continue
			}
			if !handSet[cell.SubstanceID] {
				continue
			}
			score := cellLineScore(room.Board.Cells, size, row, col, ownerVal)
			candidates = append(candidates, candidate{row, col, score})
		}
	}
	room.mu.RUnlock()

	if len(candidates) == 0 {
		return false
	}

	// Pick best candidate (highest score); break ties randomly.
	best := candidates[0]
	for _, c := range candidates[1:] {
		if c.score > best.score {
			best = c
		}
	}

	// Occupy the cell. OccupyCell handles removing the matching card and replenishing.
	if _, err := room.OccupyCell(teamIdx, best.row, best.col); err != nil {
		return false
	}
	return true
}

// cellLineScore scores how much occupying (row, col) helps teamIdx complete lines.
func cellLineScore(cells [][]BingoCell, size, row, col int, ownerVal uint) int {
	score := 0
	owns := func(r, c int) bool {
		cell := cells[r][c]
		return cell.OwnerTeamID != nil && *cell.OwnerTeamID == ownerVal
	}

	// Row contribution
	cnt := 1
	for c := 0; c < size; c++ {
		if c == col {
			continue
		}
		if owns(row, c) {
			cnt++
		}
	}
	score += cnt * cnt

	// Column contribution
	cnt = 1
	for r := 0; r < size; r++ {
		if r == row {
			continue
		}
		if owns(r, col) {
			cnt++
		}
	}
	score += cnt * cnt

	// Main diagonal
	if row == col {
		cnt = 1
		for i := 0; i < size; i++ {
			if i == row {
				continue
			}
			if owns(i, i) {
				cnt++
			}
		}
		score += cnt * cnt
	}

	// Anti-diagonal
	if row+col == size-1 {
		cnt = 1
		for i := 0; i < size; i++ {
			if i == row {
				continue
			}
			if owns(i, size-1-i) {
				cnt++
			}
		}
		score += cnt * cnt
	}

	// Add randomness to prevent deterministic play.
	score += rand.Intn(3)
	return score
}

// AIChooseSwap evaluates all possible swap moves and picks the one that maximizes
// the AI team's line-completion progress.
func AIChooseSwap(room *BingoRoom, aiTeamIdx int) (r1, c1, r2, c2 int) {
	room.mu.RLock()
	defer room.mu.RUnlock()

	size := room.Board.Size
	best := AISwapMove{R1: 0, C1: 0, R2: 0, C2: 1, Score: -1}

	for ar := 0; ar < size; ar++ {
		for ac := 0; ac < size; ac++ {
			for br := 0; br < size; br++ {
				for bc := 0; bc < size; bc++ {
					if ar == br && ac == bc {
						continue
					}
					score := evalSwap(room, aiTeamIdx, ar, ac, br, bc)
					if score > best.Score {
						best = AISwapMove{R1: ar, C1: ac, R2: br, C2: bc, Score: score}
					}
				}
			}
		}
	}
	return best.R1, best.C1, best.R2, best.C2
}

// evalSwap scores a swap move for aiTeamIdx using BFS line-progress heuristic.
func evalSwap(room *BingoRoom, teamIdx int, r1, c1, r2, c2 int) int {
	ownerVal := uint(teamIdx + 1)
	size := room.Board.Size
	cells := copyCells(room.Board.Cells, size)
	cells[r1][c1], cells[r2][c2] = cells[r2][c2], cells[r1][c1]

	score := 0
	owns := func(row, col int) bool {
		c := cells[row][col]
		return c.OwnerTeamID != nil && *c.OwnerTeamID == ownerVal
	}

	for row := 0; row < size; row++ {
		cnt := 0
		for col := 0; col < size; col++ {
			if owns(row, col) {
				cnt++
			}
		}
		score += cnt * cnt
	}
	for col := 0; col < size; col++ {
		cnt := 0
		for row := 0; row < size; row++ {
			if owns(row, col) {
				cnt++
			}
		}
		score += cnt * cnt
	}
	cnt := 0
	for i := 0; i < size; i++ {
		if owns(i, i) {
			cnt++
		}
	}
	score += cnt * cnt
	cnt = 0
	for i := 0; i < size; i++ {
		if owns(i, size-1-i) {
			cnt++
		}
	}
	score += cnt * cnt

	return score
}

func copyCells(src [][]BingoCell, size int) [][]BingoCell {
	dst := make([][]BingoCell, size)
	for i := range dst {
		dst[i] = make([]BingoCell, size)
		copy(dst[i], src[i])
	}
	return dst
}
