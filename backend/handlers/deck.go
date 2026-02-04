package handlers

import (
	"chemistryuno/database"
	"chemistryuno/repository"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetMyDecks(c *gin.Context) {
	uid := c.GetInt("uid")
	// 获取用户的自定义牌组和全局牌组
	userDecks, _ := repository.DeckRepo.FindByUserID(uint(uid))
	globalDeck, _ := repository.DeckRepo.FindGlobalDeck()

	var decks []interface{}
	if globalDeck != nil {
		decks = append(decks, globalDeck)
	}
	for _, deck := range userDecks {
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

	// 将cards转换为JSON字符串
	cardsJSON, err := json.Marshal(req.Cards)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "序列化牌组失败"})
		return
	}

	deck := &database.DeckConfig{
		Name:      req.Name,
		Cards:     string(cardsJSON),
		CreatedBy: uint(uid),
		IsGlobal:  false,
	}

	err = repository.DeckRepo.Create(deck)
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

	// 将cards转换为JSON字符串
	cardsJSON, err := json.Marshal(req.Cards)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "序列化牌组失败"})
		return
	}

	idUint, _ := strconv.ParseUint(id, 10, 32)
	deck, err := repository.DeckRepo.FindByID(uint(idUint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "卡组不存在"})
		return
	}

	// 检查权限
	if deck.CreatedBy != uint(uid) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权修改此卡组"})
		return
	}

	deck.Name = req.Name
	deck.Cards = string(cardsJSON)
	err = repository.DeckRepo.Update(deck)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新卡组失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "卡组已更新"})
}

func DeleteMyDeck(c *gin.Context) {
	id := c.Param("id")
	uid := c.GetInt("uid")

	idUint, _ := strconv.ParseUint(id, 10, 32)
	deck, err := repository.DeckRepo.FindByID(uint(idUint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "卡组不存在"})
		return
	}

	// 检查权限
	if deck.CreatedBy != uint(uid) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权删除此卡组"})
		return
	}

	err = repository.DeckRepo.Delete(uint(idUint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除卡组失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "卡组已删除"})
}
