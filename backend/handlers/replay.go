package handlers

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func parseHistoryID(c *gin.Context) (uint, error) {
	rawID := c.Param("id")
	id64, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil || id64 == 0 {
		return 0, errors.New("无效的历史记录ID")
	}
	return uint(id64), nil
}

func parseIntSliceJSON(raw database.JSON) []int {
	if len(raw) == 0 {
		return []int{}
	}
	var result []int
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return []int{}
	}
	return result
}

func containsUID(uids []int, target int) bool {
	for _, uid := range uids {
		if uid == target {
			return true
		}
	}
	return false
}

func buildReplayPlayerProfiles(uids []int) []map[string]interface{} {
	ordered := make([]int, 0, len(uids))
	seen := make(map[int]bool)
	positiveUIDs := make([]uint, 0)

	for _, uid := range uids {
		if seen[uid] {
			continue
		}
		seen[uid] = true
		ordered = append(ordered, uid)
		if uid > 0 {
			positiveUIDs = append(positiveUIDs, uint(uid))
		}
	}

	userMap := make(map[uint]*database.User)
	if len(positiveUIDs) > 0 {
		if loaded, err := repository.UserRepo.FindByUIDs(positiveUIDs); err == nil {
			userMap = loaded
		}
	}

	profiles := make([]map[string]interface{}, 0, len(ordered))
	for _, uid := range ordered {
		if uid < 0 {
			profiles = append(profiles, map[string]interface{}{
				"uid":      uid,
				"username": fmt.Sprintf("AI_%d", -uid),
				"nickname": fmt.Sprintf("AI_%d", -uid),
				"avatar":   "AI",
				"is_ai":    true,
			})
			continue
		}

		user := userMap[uint(uid)]
		username := fmt.Sprintf("UID_%d", uid)
		nickname := username
		avatar := ""
		if user != nil {
			username = user.Username
			nickname = user.Nickname
			if nickname == "" {
				nickname = username
			}
			avatar = user.Avatar
		}

		profiles = append(profiles, map[string]interface{}{
			"uid":      uid,
			"username": username,
			"nickname": nickname,
			"avatar":   avatar,
			"is_ai":    false,
		})
	}

	return profiles
}

func buildReplayResponse(history *database.GameHistory) map[string]interface{} {
	players := parseIntSliceJSON(history.Players)
	cheatUIDs := parseIntSliceJSON(history.CheatUIDs)

	var replayData interface{}
	if history.ReplayLog != "" {
		if err := json.Unmarshal([]byte(history.ReplayLog), &replayData); err != nil {
			replayData = map[string]interface{}{
				"raw": history.ReplayLog,
			}
		}
	}

	return map[string]interface{}{
		"id":                history.ID,
		"room_id":           history.RoomID,
		"is_invalid":        history.IsInvalid,
		"invalid_reason":    history.InvalidReason,
		"has_replay":        history.ReplayLog != "",
		"replay_permanent":  history.ReplayPermanent,
		"replay_expires_at": history.ReplayExpiresAt,
		"replay_cleared_at": history.ReplayClearedAt,
		"cheat_detected":    history.CheatDetected,
		"cheat_uids":        cheatUIDs,
		"players":           players,
		"player_profiles":   buildReplayPlayerProfiles(players),
		"replay":            replayData,
	}
}

// GetMyGameReplay 获取当前用户可访问的单局回放
func GetMyGameReplay(c *gin.Context) {
	uid := c.GetInt("uid")
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	}

	historyID, err := parseHistoryID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	history, err := repository.GameRepo.FindByID(historyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "历史记录不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误: " + err.Error()})
		return
	}

	players := parseIntSliceJSON(history.Players)
	if !containsUID(players, uid) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权访问该回放"})
		return
	}

	if history.ReplayLog == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "回放不存在或已过期"})
		return
	}

	c.JSON(http.StatusOK, buildReplayResponse(history))
}

// GetAdminGameReplay 管理员查看任意回放
func GetAdminGameReplay(c *gin.Context) {
	historyID, err := parseHistoryID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	history, err := repository.GameRepo.FindByID(historyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "历史记录不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误: " + err.Error()})
		return
	}

	if history.ReplayLog == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "回放不存在或已过期"})
		return
	}

	c.JSON(http.StatusOK, buildReplayResponse(history))
}

// ClearAdminGameReplay 管理员消除某条回放内容（含永久回放）
func ClearAdminGameReplay(c *gin.Context) {
	historyID, err := parseHistoryID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := repository.GameRepo.ClearReplayByID(historyID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "历史记录不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清除回放失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "回放已清除"})
}
