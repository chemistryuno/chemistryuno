package handlers

import (
	"chemistryuno/backend/bingo"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func init() {
	bingo.OnAITurnDone = func(room *bingo.BingoRoom) {
		broadcastBingoUpdate(room)
	}
}

// getBingoPlayerHand is a bridge used by GetTeammateHand in team.go.
func getBingoPlayerHand(uid uint) []bingo.HandCard {
	return bingo.GetPlayerHand(uid)
}

// CreateBingoRoom creates a new BINGO room with randomly assigned teams.
// The caller provides the list of participant UIDs; teams are split randomly at creation time.
// Supports AI opponents: specify ai_count and ai_difficulty.
func CreateBingoRoom(c *gin.Context) {
	var req struct {
		PlayerUIDs     []uint `json:"player_uids" binding:"required"`
		TimeoutMinutes int    `json:"timeout_minutes"`
		AICount        int    `json:"ai_count"`
		AIDifficulty   int    `json:"ai_difficulty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.PlayerUIDs)+req.AICount < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "至少需要 2 名玩家（含AI）"})
		return
	}

	room, err := bingo.CreateRoom(req.PlayerUIDs, req.TimeoutMinutes, req.AICount, req.AIDifficulty)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, bingoRoomView(room))
}

// GetBingoRoom returns the current state of a BINGO room.
func GetBingoRoom(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	room := bingo.GetRoom(uint(id))
	if room == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "房间不存在"})
		return
	}
	c.JSON(http.StatusOK, bingoRoomView(room))
}

// VoteBingoRefresh records a vote for board refresh.
func VoteBingoRefresh(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	room := bingo.GetRoom(uint(id))
	if room == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "房间不存在"})
		return
	}
	uid := uint(c.GetInt("uid"))
	teamIdx := room.GetTeamForUID(uid)
	if teamIdx < 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "不是该房间的参与者"})
		return
	}

	var req struct {
		Agree bool `json:"agree"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	refreshed, err := room.VoteRefresh(teamIdx, req.Agree)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "棋盘仅可刷新一次" {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	broadcastBingoUpdate(room)
	c.JSON(http.StatusOK, gin.H{"refreshed": refreshed, "board": room.Board})
}

// StartBingoGame starts the game.
func StartBingoGame(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	room := bingo.GetRoom(uint(id))
	if room == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "房间不存在"})
		return
	}
	uid := uint(c.GetInt("uid"))
	if room.GetTeamForUID(uid) < 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "不是该房间的参与者"})
		return
	}

	room.StartGame(func(roomID uint) {
		r := bingo.GetRoom(roomID)
		if r != nil {
			r.TimeoutSettle()
			broadcastBingoUpdate(r)
		}
	})
	broadcastBingoUpdate(room)
	bingo.TriggerAITurnIfNeeded(room)
	c.JSON(http.StatusOK, bingoRoomView(room))
}

// SwapBingoCells performs a cell swap (turn-based).
func SwapBingoCells(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	room := bingo.GetRoom(uint(id))
	if room == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "房间不存在"})
		return
	}
	uid := uint(c.GetInt("uid"))
	teamIdx := room.GetTeamForUID(uid)
	if teamIdx < 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "不是该房间的参与者"})
		return
	}

	var req struct {
		R1 int `json:"r1"`
		C1 int `json:"c1"`
		R2 int `json:"r2"`
		C2 int `json:"c2"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := room.SwapCells(teamIdx, req.R1, req.C1, req.R2, req.C2); err != nil {
		if err.Error() == "未到你的回合" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	broadcastBingoUpdate(room)
	bingo.TriggerAITurnIfNeeded(room)
	c.JSON(http.StatusOK, bingoRoomView(room))
}

// OccupyBingoCell claims a cell after a correct answer.
func OccupyBingoCell(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	room := bingo.GetRoom(uint(id))
	if room == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "房间不存在"})
		return
	}
	uid := uint(c.GetInt("uid"))
	teamIdx := room.GetTeamForUID(uid)
	if teamIdx < 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "不是该房间的参与者"})
		return
	}

	var req struct {
		Row         int  `json:"row"`
		Col         int  `json:"col"`
		SubstanceID uint `json:"substance_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Row < 0 || req.Row >= room.Board.Size || req.Col < 0 || req.Col >= room.Board.Size {
		c.JSON(http.StatusBadRequest, gin.H{"error": "坐标超出范围"})
		return
	}
	cell := room.Board.Cells[req.Row][req.Col]
	if cell.SubstanceID != req.SubstanceID {
		broadcastBingoUpdate(room)
		c.JSON(http.StatusOK, gin.H{"correct": false})
		return
	}

	win, err := room.OccupyCell(teamIdx, req.Row, req.Col)
	if err != nil {
		if err.Error() == "未到你的回合" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	broadcastBingoUpdate(room)
	if !win {
		bingo.TriggerAITurnIfNeeded(room)
	}
	c.JSON(http.StatusOK, gin.H{"correct": true, "win": win})
}

func broadcastBingoUpdate(room *bingo.BingoRoom) {
	// No-op stub — websocket push can be wired here in the future.
}

func bingoRoomView(room *bingo.BingoRoom) interface{} {
	return gin.H{
		"id":               room.ID,
		"team_a_members":   room.TeamAMembers,
		"team_b_members":   room.TeamBMembers,
		"ai_members":       room.AIMembers,
		"ai_difficulty":    room.AIDifficulty,
		"board":            room.Board,
		"status":           room.Status,
		"current_turn":     room.CurrentTurn,
		"timeout_minutes":  room.TimeoutMinutes,
		"winner_team_idx":  room.WinnerTeamIdx,
	}
}
