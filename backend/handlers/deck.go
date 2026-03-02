package handlers

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/game"
	"chemistryuno/backend/models"
	"chemistryuno/backend/repository"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetMyDecks(c *gin.Context) {
	uid := c.GetInt("uid")

	// 获取全局牌组
	globalDecks, err := repository.DeckRepo.FindGlobalDecks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取全局牌组失败"})
		return
	}

	// 获取用户的自定义牌组
	userDecks, err := repository.DeckRepo.FindByUserUID(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户牌组失败"})
		return
	}

	decks := make([]models.DeckConfig, 0, len(globalDecks)+len(userDecks))
	// 添加所有全局牌组
	for i := range globalDecks {
		cards := normalizeDeckCardsFromJSON(globalDecks[i].Cards)
		if len(cards) == 0 {
			cards = game.BuiltinDeckDefaults()
		}
		decks = append(decks, models.DeckConfig{
			ID:           int(globalDecks[i].ID),
			Name:         globalDecks[i].Name,
			IsGlobal:     globalDecks[i].IsGlobal,
			Cards:        cards,
			InitialCards: normalizeInitialCards(globalDecks[i].InitialCards),
			CreatedBy:    int(globalDecks[i].CreatedByUID),
			CreatedAt:    globalDecks[i].CreatedAt,
		})
	}
	// 添加用户的所有牌组
	for i := range userDecks {
		cards := normalizeDeckCardsFromJSON(userDecks[i].Cards)
		decks = append(decks, models.DeckConfig{
			ID:           int(userDecks[i].ID),
			Name:         userDecks[i].Name,
			IsGlobal:     userDecks[i].IsGlobal,
			Cards:        cards,
			InitialCards: normalizeInitialCards(userDecks[i].InitialCards),
			CreatedBy:    int(userDecks[i].CreatedByUID),
			CreatedAt:    userDecks[i].CreatedAt,
		})
	}

	c.JSON(http.StatusOK, decks)
}

func CreateMyDeck(c *gin.Context) {
	var req struct {
		Name         string         `json:"name" binding:"required"`
		Cards        map[string]int `json:"cards" binding:"required"`
		InitialCards int            `json:"initial_cards"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid := c.GetInt("uid")

	if req.InitialCards <= 0 {
		req.InitialCards = 10
	}

	normalizedCards, unknownCards := game.NormalizeBuiltinDeckCards(req.Cards)
	if len(unknownCards) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":          "自定义卡组仅支持原有普通牌和特殊牌，插件牌请在插件中管理",
			"unknown_cards":  unknownCards,
			"allowed_scope":  "builtin_only",
			"plugin_managed": true,
		})
		return
	}
	if len(normalizedCards) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "卡组不能为空"})
		return
	}

	// 将cards转换为JSON字符串
	cardsJSON, err := json.Marshal(normalizedCards)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "序列化牌组失败"})
		return
	}

	deck := &database.DeckConfig{
		Name:         req.Name,
		Cards:        cardsJSON,
		InitialCards: req.InitialCards,
		CreatedByUID: uint(uid),
		IsGlobal:     false,
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
		Name         string         `json:"name" binding:"required"`
		Cards        map[string]int `json:"cards" binding:"required"`
		InitialCards int            `json:"initial_cards"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.InitialCards <= 0 {
		req.InitialCards = 10
	}

	normalizedCards, unknownCards := game.NormalizeBuiltinDeckCards(req.Cards)
	if len(unknownCards) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":          "自定义卡组仅支持原有普通牌和特殊牌，插件牌请在插件中管理",
			"unknown_cards":  unknownCards,
			"allowed_scope":  "builtin_only",
			"plugin_managed": true,
		})
		return
	}
	if len(normalizedCards) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "卡组不能为空"})
		return
	}

	// 将cards转换为JSON字符串
	cardsJSON, err := json.Marshal(normalizedCards)
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
	if deck.CreatedByUID != uint(uid) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权修改此卡组"})
		return
	}

	deck.Name = req.Name
	deck.Cards = cardsJSON
	deck.InitialCards = req.InitialCards
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
	if deck.CreatedByUID != uint(uid) {
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

func normalizeDeckCardsFromJSON(raw database.JSON) map[string]int {
	var cards map[string]int
	if err := json.Unmarshal([]byte(raw), &cards); err != nil {
		return map[string]int{}
	}
	normalizedCards, _ := game.NormalizeBuiltinDeckCards(cards)
	return normalizedCards
}

func normalizeInitialCards(v int) int {
	if v <= 0 {
		return 10
	}
	return v
}
