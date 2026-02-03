package handlers

import (
	"chemistryuno/database"
	"chemistryuno/models"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetMyDecks(c *gin.Context) {
	uid := c.GetInt("uid")
	rows, err := database.LegacyDB.Query("SELECT id, name, is_global, cards, created_at FROM deck_configs WHERE created_by = ? OR is_global = 1", uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取卡组失败"})
		return
	}
	defer rows.Close()

	var decks []models.DeckConfig
	for rows.Next() {
		var deck models.DeckConfig
		var cardsJSON string
		if err := rows.Scan(&deck.ID, &deck.Name, &deck.IsGlobal, &cardsJSON, &deck.CreatedAt); err != nil {
			continue
		}
		json.Unmarshal([]byte(cardsJSON), &deck.Cards)
		decks = append(decks, deck)
	}

	c.JSON(http.StatusOK, decks)
}

func CreateMyDeck(c *gin.Context) {
	var req struct {
		Name  string         `json:"name" binding:"required"`
		Cards map[string]int `json:"cards" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid := c.GetInt("uid")
	cardsJSON, _ := json.Marshal(req.Cards)

	_, err := database.LegacyDB.Exec("INSERT INTO deck_configs (name, cards, created_by, is_global) VALUES (?, ?, ?, 0)",
		req.Name, string(cardsJSON), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建卡组失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "自定义卡组已创建"})
}

func UpdateMyDeck(c *gin.Context) {
	id := c.Param("id")
	uid := c.GetInt("uid")

	var req struct {
		Name  string         `json:"name" binding:"required"`
		Cards map[string]int `json:"cards" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cardsJSON, _ := json.Marshal(req.Cards)

	result, err := database.LegacyDB.Exec("UPDATE deck_configs SET name = ?, cards = ? WHERE id = ? AND created_by = ?",
		req.Name, string(cardsJSON), id, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新卡组失败"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权修改此卡组或卡组不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "卡组已更新"})
}

func DeleteMyDeck(c *gin.Context) {
	id := c.Param("id")
	uid := c.GetInt("uid")

	result, err := database.LegacyDB.Exec("DELETE FROM deck_configs WHERE id = ? AND created_by = ? AND is_global = 0", id, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除卡组失败"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权删除此卡组或卡组不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "卡组已删除"})
}
