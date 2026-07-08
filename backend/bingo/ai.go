package bingo

// AISwapMove represents a potential swap move evaluated by the AI.
type AISwapMove struct {
	R1, C1, R2, C2 int
	Score          int
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
