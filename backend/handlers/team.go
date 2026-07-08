package handlers

import (
	"chemistryuno/backend/bingo"
	"chemistryuno/backend/database"
	"chemistryuno/backend/team"
	"chemistryuno/backend/websocket"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateTeam creates a new team with the current user as leader.
func CreateTeam(c *gin.Context) {
	uid := uint(c.GetInt("uid"))
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t, err := team.GlobalManager.CreateTeam(uid, req.Name)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "请先退出当前队伍" {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, t)
}

// JoinTeam joins a team via invite code.
func JoinTeam(c *gin.Context) {
	uid := uint(c.GetInt("uid"))
	var req struct {
		InviteCode string `json:"invite_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t, err := team.GlobalManager.JoinTeam(uid, req.InviteCode)
	if err != nil {
		switch err.Error() {
		case "邀请码无效":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "队伍已满":
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	// Notify team members.
	if websocket.GlobalHub != nil {
		websocket.GlobalHub.BroadcastToRoom(team.GetTeamChatRoomID(t.ID), websocket.Message{
			Type: "team_member_joined",
			UID:  int(uid),
		})
	}
	c.JSON(http.StatusOK, t)
}

// LeaveTeam removes the current user from their team.
func LeaveTeam(c *gin.Context) {
	uid := uint(c.GetInt("uid"))
	disbanded, err := team.GlobalManager.LeaveTeam(uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if websocket.GlobalHub != nil {
		state := team.GlobalManager.GetTeamByUID(uid)
		if state != nil {
			eventType := "team_member_left"
			if disbanded {
				eventType = "team_disbanded"
			}
			websocket.GlobalHub.BroadcastToRoom(team.GetTeamChatRoomID(state.Team.ID), websocket.Message{
				Type: eventType,
				UID:  int(uid),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"disbanded": disbanded})
}

// DisbandTeam disbands the team (leader only).
func DisbandTeam(c *gin.Context) {
	uid := uint(c.GetInt("uid"))
	state := team.GlobalManager.GetTeamByUID(uid)
	var teamID uint
	if state != nil {
		teamID = state.Team.ID
	}
	if err := team.GlobalManager.DisbandTeam(uid); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	if teamID != 0 && websocket.GlobalHub != nil {
		websocket.GlobalHub.BroadcastToRoom(team.GetTeamChatRoomID(teamID), websocket.Message{
			Type: "team_disbanded",
			UID:  int(uid),
		})
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GetMyTeam returns the current user's team info.
func GetMyTeam(c *gin.Context) {
	uid := uint(c.GetInt("uid"))
	state := team.GlobalManager.GetTeamByUID(uid)
	if state == nil {
		c.JSON(http.StatusOK, gin.H{"team": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"team":    state.Team,
		"members": state.Members,
	})
}

// GetTeamHistory returns the last 50 chat messages for the team.
func GetTeamHistory(c *gin.Context) {
	uid := uint(c.GetInt("uid"))
	state := team.GlobalManager.GetTeamByUID(uid)
	if state == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "不在任何队伍中"})
		return
	}
	var msgs []database.TeamChatMessage
	database.DB.Where("team_id = ?", state.Team.ID).
		Order("created_at desc").
		Limit(50).
		Find(&msgs)
	// Reverse to chronological order.
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	c.JSON(http.StatusOK, msgs)
}

// GetTeammateHand returns a teammate's hand cards (BINGO mode only, same team).
func GetTeammateHand(c *gin.Context) {
	uid := uint(c.GetInt("uid"))
	targetUIDStr := c.Param("uid")
	targetUID64, err := strconv.ParseUint(targetUIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uid"})
		return
	}
	targetUID := uint(targetUID64)

	if !bingo.AreTeammates(uid, targetUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "只能查看队友手牌"})
		return
	}
	hand := getBingoPlayerHand(targetUID)
	c.JSON(http.StatusOK, gin.H{"hand": hand})
}
