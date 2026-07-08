package bingo

import (
	"chemistryuno/backend/database"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const DefaultBoardSize = 5
const DefaultTimeoutMinutes = 10

// team index constants
const (
	TeamA = 0
	TeamB = 1
)

// BingoCell is a single cell on the board.
type BingoCell struct {
	Row         int    `json:"row"`
	Col         int    `json:"col"`
	SubstanceID uint   `json:"substance_id"`
	Formula     string `json:"formula"`
	Name        string `json:"name"`
	// OwnerTeamID: nil = unoccupied, 1 = team A, 2 = team B
	OwnerTeamID *uint  `json:"owner_team_id"`
}

// BingoBoard is the full board state.
type BingoBoard struct {
	Size  int           `json:"size"`
	Cells [][]BingoCell `json:"cells"`
}

// BingoRoom holds the in-memory state of a running BINGO game.
type BingoRoom struct {
	ID             uint
	TeamAMembers   []uint // UIDs on team A (randomly assigned at creation)
	TeamBMembers   []uint // UIDs on team B
	Board          *BingoBoard
	Status         string // waiting, playing, finished
	TimeoutMinutes int
	CreatedAt      time.Time
	StartedAt      *time.Time
	Timer          *time.Timer
	WinnerTeamIdx  int // 0 = A wins, 1 = B wins, -1 = no winner yet

	// Voting state for board refresh.
	VoteA     *bool // nil = not voted
	VoteB     *bool
	Refreshed bool

	// CurrentTurn: 0 = team A's turn, 1 = team B's turn
	CurrentTurn int

	// Player hands: uid -> []HandCard
	Hands map[uint][]HandCard

	mu sync.RWMutex
}

// HandCard represents a card in a player's hand.
type HandCard struct {
	SubstanceID uint   `json:"substance_id"`
	Formula     string `json:"formula"`
	Name        string `json:"name"`
}

type substanceDifficulty struct {
	substance database.Substance
	score     int
}

var (
	rooms   = make(map[uint]*BingoRoom)
	roomsMu sync.RWMutex
)

// GetRoom returns a bingo room by ID.
func GetRoom(id uint) *BingoRoom {
	roomsMu.RLock()
	defer roomsMu.RUnlock()
	return rooms[id]
}

// GetBingoRoomID returns the WebSocket channel name for a bingo room.
func GetBingoRoomID(roomID uint) string {
	return fmt.Sprintf("bingo_%d", roomID)
}

// GetPlayerHand returns a copy of a player's hand, or nil if not found.
func GetPlayerHand(uid uint) []HandCard {
	roomsMu.RLock()
	defer roomsMu.RUnlock()
	for _, r := range rooms {
		r.mu.RLock()
		hand, ok := r.Hands[uid]
		r.mu.RUnlock()
		if ok {
			cp := make([]HandCard, len(hand))
			copy(cp, hand)
			return cp
		}
	}
	return nil
}

// AreTeammates returns true if uid1 and uid2 are on the same team in any active bingo room.
func AreTeammates(uid1, uid2 uint) bool {
	roomsMu.RLock()
	defer roomsMu.RUnlock()
	for _, r := range rooms {
		if r.Status != "playing" {
			continue
		}
		r.mu.RLock()
		idx1 := r.getTeamIdxUnlocked(uid1)
		idx2 := r.getTeamIdxUnlocked(uid2)
		r.mu.RUnlock()
		if idx1 >= 0 && idx1 == idx2 {
			return true
		}
	}
	return false
}

// GetPlayerTeammates returns the UIDs of teammates (excluding self) in any active bingo room.
func GetPlayerTeammates(uid uint) []uint {
	roomsMu.RLock()
	defer roomsMu.RUnlock()
	for _, r := range rooms {
		if r.Status != "playing" {
			continue
		}
		r.mu.RLock()
		idx := r.getTeamIdxUnlocked(uid)
		var mates []uint
		if idx == TeamA {
			for _, m := range r.TeamAMembers {
				if m != uid {
					mates = append(mates, m)
				}
			}
		} else if idx == TeamB {
			for _, m := range r.TeamBMembers {
				if m != uid {
					mates = append(mates, m)
				}
			}
		}
		r.mu.RUnlock()
		if idx >= 0 {
			return mates
		}
	}
	return nil
}

// CreateRoom creates a new BINGO room with randomly assigned teams.
func CreateRoom(playerUIDs []uint, timeoutMinutes int) (*BingoRoom, error) {
	if len(playerUIDs) < 2 {
		return nil, errors.New("至少需要 2 名玩家")
	}
	if timeoutMinutes <= 0 {
		timeoutMinutes = DefaultTimeoutMinutes
	}

	// Shuffle players and split into two teams.
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	shuffled := make([]uint, len(playerUIDs))
	copy(shuffled, playerUIDs)
	rng.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

	half := len(shuffled) / 2
	teamA := shuffled[:half]
	teamB := shuffled[half:]

	// Generate the board.
	cells, err := generateBoard(DefaultBoardSize)
	if err != nil {
		return nil, err
	}
	board := &BingoBoard{Size: DefaultBoardSize, Cells: cells}

	// Persist to DB.
	boardJSON, _ := json.Marshal(board)
	teamAJSON, _ := json.Marshal(teamA)
	teamBJSON, _ := json.Marshal(teamB)

	dbRoom := database.BingoRoom{
		TeamAMembers:   database.JSON(teamAJSON),
		TeamBMembers:   database.JSON(teamBJSON),
		Board:          database.JSON(boardJSON),
		Status:         "waiting",
		TimeoutMinutes: timeoutMinutes,
	}
	if err := database.DB.Create(&dbRoom).Error; err != nil {
		return nil, err
	}

	room := &BingoRoom{
		ID:             dbRoom.ID,
		TeamAMembers:   teamA,
		TeamBMembers:   teamB,
		Board:          board,
		Status:         "waiting",
		TimeoutMinutes: timeoutMinutes,
		CreatedAt:      dbRoom.CreatedAt,
		CurrentTurn:    TeamA,
		WinnerTeamIdx:  -1,
		Hands:          make(map[uint][]HandCard),
	}

	roomsMu.Lock()
	rooms[room.ID] = room
	roomsMu.Unlock()

	return room, nil
}

// generateBoard creates a board of the given size, placing hard substances on outer cells.
func generateBoard(size int) ([][]BingoCell, error) {
	needed := size * size
	var substances []database.Substance
	database.DB.Where("status = ?", "approved").
		Order("RANDOM()").
		Limit(needed * 3).
		Find(&substances)

	if len(substances) < needed {
		return nil, errors.New("物质数据不足，无法生成棋盘")
	}

	scored := make([]substanceDifficulty, 0, len(substances))
	for _, s := range substances {
		scored = append(scored, substanceDifficulty{substance: s, score: len(s.Formula)})
	}
	// Sort descending by score.
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	type pos struct{ r, c int }
	var outer, inner []pos
	for r := 0; r < size; r++ {
		for c := 0; c < size; c++ {
			if r == 0 || r == size-1 || c == 0 || c == size-1 {
				outer = append(outer, pos{r, c})
			} else {
				inner = append(inner, pos{r, c})
			}
		}
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(outer), func(i, j int) { outer[i], outer[j] = outer[j], outer[i] })
	rng.Shuffle(len(inner), func(i, j int) { inner[i], inner[j] = inner[j], inner[i] })

	cells := make([][]BingoCell, size)
	for i := range cells {
		cells[i] = make([]BingoCell, size)
	}
	idx := 0
	for _, p := range outer {
		s := scored[idx].substance
		cells[p.r][p.c] = BingoCell{Row: p.r, Col: p.c, SubstanceID: s.ID, Formula: s.Formula, Name: s.Name}
		idx++
	}
	for _, p := range inner {
		s := scored[idx].substance
		cells[p.r][p.c] = BingoCell{Row: p.r, Col: p.c, SubstanceID: s.ID, Formula: s.Formula, Name: s.Name}
		idx++
	}
	return cells, nil
}

// GetTeamForUID returns the team index (0 = A, 1 = B, -1 = not a participant).
func (r *BingoRoom) GetTeamForUID(uid uint) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.getTeamIdxUnlocked(uid)
}

func (r *BingoRoom) getTeamIdxUnlocked(uid uint) int {
	for _, m := range r.TeamAMembers {
		if m == uid {
			return TeamA
		}
	}
	for _, m := range r.TeamBMembers {
		if m == uid {
			return TeamB
		}
	}
	return -1
}

// VoteRefresh records a vote for board refresh.
func (r *BingoRoom) VoteRefresh(teamIdx int, agree bool) (refreshed bool, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Refreshed {
		return false, errors.New("棋盘仅可刷新一次")
	}
	if r.Status != "waiting" {
		return false, errors.New("投票仅在游戏开始前有效")
	}

	switch teamIdx {
	case TeamA:
		r.VoteA = &agree
	case TeamB:
		r.VoteB = &agree
	default:
		return false, errors.New("无效的队伍")
	}

	if r.VoteA == nil || r.VoteB == nil {
		return false, nil
	}
	if *r.VoteA && *r.VoteB {
		cells, err := generateBoard(r.Board.Size)
		if err != nil {
			return false, err
		}
		r.Board.Cells = cells
		r.Refreshed = true
		boardJSON, _ := json.Marshal(r.Board)
		database.DB.Model(&database.BingoRoom{}).Where("id = ?", r.ID).
			Update("board", string(boardJSON))
		return true, nil
	}
	return false, nil
}

// StartGame transitions the room to playing state and starts the timeout timer.
func (r *BingoRoom) StartGame(onTimeout func(roomID uint)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	r.StartedAt = &now
	r.Status = "playing"
	database.DB.Model(&database.BingoRoom{}).Where("id = ?", r.ID).Update("status", "playing")
	dur := time.Duration(r.TimeoutMinutes) * time.Minute
	r.Timer = time.AfterFunc(dur, func() { onTimeout(r.ID) })
}

// SwapCells swaps the substances at two board positions (turn-based).
func (r *BingoRoom) SwapCells(teamIdx int, r1, c1, r2, c2 int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Status != "playing" {
		return errors.New("游戏未在进行中")
	}
	if r.CurrentTurn != teamIdx {
		return errors.New("未到你的回合")
	}
	size := r.Board.Size
	if r1 < 0 || r1 >= size || c1 < 0 || c1 >= size || r2 < 0 || r2 >= size || c2 < 0 || c2 >= size {
		return errors.New("坐标超出棋盘范围")
	}
	a := &r.Board.Cells[r1][c1]
	b := &r.Board.Cells[r2][c2]
	a.SubstanceID, b.SubstanceID = b.SubstanceID, a.SubstanceID
	a.Formula, b.Formula = b.Formula, a.Formula
	a.Name, b.Name = b.Name, a.Name

	r.switchTurnUnlocked()
	return nil
}

// OccupyCell marks a cell as owned by teamIdx after a correct answer.
// Returns true if the team wins.
func (r *BingoRoom) OccupyCell(teamIdx int, row, col int) (win bool, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Status != "playing" {
		return false, errors.New("游戏未在进行中")
	}
	size := r.Board.Size
	if row < 0 || row >= size || col < 0 || col >= size {
		return false, errors.New("坐标超出范围")
	}
	cell := &r.Board.Cells[row][col]
	if cell.OwnerTeamID != nil {
		return false, errors.New("该格子已被占领")
	}
	// Use teamIdx+1 as the stable owner value (1=A, 2=B).
	ownerVal := uint(teamIdx + 1)
	cell.OwnerTeamID = &ownerVal

	win = r.checkWinUnlocked(teamIdx)
	if win || r.isBoardFullUnlocked() {
		r.finishUnlocked(teamIdx)
	}
	return win, nil
}

// TimeoutSettle settles by cell count.
func (r *BingoRoom) TimeoutSettle() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Status != "playing" {
		return
	}
	aCount, bCount := r.countCellsUnlocked()
	winner := TeamA
	if bCount > aCount {
		winner = TeamB
	}
	r.finishUnlocked(winner)
}

func (r *BingoRoom) countCellsUnlocked() (a, b int) {
	for _, row := range r.Board.Cells {
		for _, cell := range row {
			if cell.OwnerTeamID == nil {
				continue
			}
			if *cell.OwnerTeamID == 1 {
				a++
			} else {
				b++
			}
		}
	}
	return
}

func (r *BingoRoom) isBoardFullUnlocked() bool {
	for _, row := range r.Board.Cells {
		for _, cell := range row {
			if cell.OwnerTeamID == nil {
				return false
			}
		}
	}
	return true
}

func (r *BingoRoom) finishUnlocked(winnerTeamIdx int) {
	r.Status = "finished"
	r.WinnerTeamIdx = winnerTeamIdx
	if r.Timer != nil {
		r.Timer.Stop()
	}
	database.DB.Model(&database.BingoRoom{}).Where("id = ?", r.ID).Update("status", "finished")
}

// checkWinUnlocked returns true if teamIdx has a complete line.
func (r *BingoRoom) checkWinUnlocked(teamIdx int) bool {
	size := r.Board.Size
	cells := r.Board.Cells
	ownerVal := uint(teamIdx + 1)
	owns := func(row, col int) bool {
		c := cells[row][col]
		return c.OwnerTeamID != nil && *c.OwnerTeamID == ownerVal
	}

	for row := 0; row < size; row++ {
		win := true
		for col := 0; col < size; col++ {
			if !owns(row, col) {
				win = false
				break
			}
		}
		if win {
			return true
		}
	}
	for col := 0; col < size; col++ {
		win := true
		for row := 0; row < size; row++ {
			if !owns(row, col) {
				win = false
				break
			}
		}
		if win {
			return true
		}
	}
	diagMain := true
	for i := 0; i < size; i++ {
		if !owns(i, i) {
			diagMain = false
			break
		}
	}
	if diagMain {
		return true
	}
	diagAnti := true
	for i := 0; i < size; i++ {
		if !owns(i, size-1-i) {
			diagAnti = false
			break
		}
	}
	return diagAnti
}

func (r *BingoRoom) switchTurnUnlocked() {
	r.CurrentTurn = 1 - r.CurrentTurn
}
