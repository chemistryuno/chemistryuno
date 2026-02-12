package handlers

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/game"
	"chemistryuno/backend/middleware"
	"chemistryuno/backend/repository"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SubstanceUpdateRequest 物质更新请求
type SubstanceUpdateRequest struct {
	Name     string `json:"name" binding:"required"`
	Formula  string `json:"formula" binding:"required"`
	Elements string `json:"elements"`
}

// SubmitSubstanceUpdate 玩家提交物质更新建议
func SubmitSubstanceUpdate(c *gin.Context) {
	userID, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的物质ID"})
		return
	}

	var req SubstanceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	substanceRepo := repository.NewSubstanceRepository()

	// 获取原物质的 group_id 和创建者
	groupID, creatorUID, err := substanceRepo.GetGroupIDAndCreatorByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "物质不存在"})
		return
	}

	// 如果没有 group_id，创建一个新的
	if groupID == nil {
		// 使用当前物质的 ID 作为 group_id
		gid := uint(id)
		groupID = &gid
		// 更新原物质的 group_id
		database.DB.Model(&database.Substance{}).Where("id = ?", id).Update("group_id", gid)
	}

	// 解析元素（如果未提供）
	elements := req.Elements
	if elements == "" {
		elementsMap := game.ParseSubstanceForElements(req.Formula)
		var elementsArr []string
		for e := range elementsMap {
			elementsArr = append(elementsArr, e)
		}
		elements = fmt.Sprintf("%v", elementsArr)
	}

	// 安全获取 userID
	uID := uint(0)
	if u, ok := userID.(int); ok {
		uID = uint(u)
	} else if u, ok := userID.(uint); ok {
		uID = u
	}

	// 创建新的更新建议
	newSubstance := &database.Substance{
		Name:         req.Name,
		Formula:      req.Formula,
		Elements:     elements,
		CreatedByUID: uID,
		Status:       "pending",
		GroupID:      groupID,
	}

	if err := substanceRepo.Create(newSubstance); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交更新建议失败"})
		return
	}

	// 如果原创建者不是当前用户，标记该组需要完善
	if creatorUID != uID {
		substanceRepo.MarkNeedsImprovement(*groupID, true)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "物质更新建议已提交，等待管理员审核",
		"id":      newSubstance.ID,
	})
}

// AdminUpdateSubstance 管理员直接更新物质
func AdminUpdateSubstance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的物质ID"})
		return
	}

	var req SubstanceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	substanceRepo := repository.NewSubstanceRepository()

	// 解析元素（如果未提供）
	elements := req.Elements
	if elements == "" {
		elementsMap := game.ParseSubstanceForElements(req.Formula)
		var elementsArr []string
		for e := range elementsMap {
			elementsArr = append(elementsArr, e)
		}
		elements = fmt.Sprintf("%v", elementsArr)
	}

	// 直接更新物质
	err = substanceRepo.UpdateWithElements(uint(id), req.Formula, req.Name, elements, "approved")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新物质失败"})
		return
	}

	// 获取 group_id 并清除待完善标记
	groupID, _, err := substanceRepo.GetGroupIDAndCreatorByID(uint(id))
	if err == nil && groupID != nil {
		substanceRepo.MarkNeedsImprovement(*groupID, false)
	}

	// 重建物质缓存
	game.RebuildSubstanceCache()

	c.JSON(http.StatusOK, gin.H{"message": "物质已更新"})
}

// ApproveSubstanceUpdate 管理员批准物质更新建议
func ApproveSubstanceUpdate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的物质ID"})
		return
	}

	substanceRepo := repository.NewSubstanceRepository()

	// 获取物质的 group_id
	groupID, _, err := substanceRepo.GetGroupIDAndCreatorByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "物质不存在"})
		return
	}

	if groupID == nil {
		gid := uint(id)
		groupID = &gid
		// 补全 group_id 以免下次失效
		database.DB.Model(&database.Substance{}).Where("id = ?", id).Update("group_id", gid)
	}

	// 批准整个组
	err = substanceRepo.UpdateStatusByGroupID(*groupID, "approved")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批准失败"})
		return
	}

	// 清除待完善标记
	substanceRepo.MarkNeedsImprovement(*groupID, false)

	// 重建物质缓存
	game.RebuildSubstanceCache()

	c.JSON(http.StatusOK, gin.H{"message": "物质更新已批准"})
}

// RejectSubstanceUpdate 管理员拒绝物质更新建议
func RejectSubstanceUpdate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的物质ID"})
		return
	}

	substanceRepo := repository.NewSubstanceRepository()

	// 获取物质的 group_id
	groupID, _, err := substanceRepo.GetGroupIDAndCreatorByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "物质不存在"})
		return
	}

	if groupID == nil {
		gid := uint(id)
		groupID = &gid
	}

	// 删除整个组
	err = substanceRepo.DeleteByGroupID(*groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	// 如果是 NULL 情况，确保也删除了自己（以防 repo 里的 DeleteByGroupID 没查到）
	database.DB.Delete(&database.Substance{}, id)

	// 重建物质缓存
	game.RebuildSubstanceCache()

	c.JSON(http.StatusOK, gin.H{"message": "物质更新建议已拒绝"})
}

// GetSubstancesByGroup 获取组内所有物质版本
func GetSubstancesByGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的物质ID"})
		return
	}

	substanceRepo := repository.NewSubstanceRepository()

	// 获取物质的 group_id
	groupID, _, err := substanceRepo.GetGroupIDAndCreatorByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "物质不存在"})
		return
	}

	if groupID == nil {
		gid := uint(id)
		groupID = &gid
	}

	// 获取组内所有物质
	substances, err := substanceRepo.FindByGroupID(*groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 如果本来就没有组，至少返回它自己
	if len(substances) == 0 {
		var sub database.Substance
		if err := database.DB.First(&sub, id).Error; err == nil {
			substances = append(substances, sub)
		}
	}

	c.JSON(http.StatusOK, substances)
}

// GetAllSubstancesGrouped 获取所有物质（分组显示）
func GetAllSubstancesGrouped(c *gin.Context) {
	substanceRepo := repository.NewSubstanceRepository()

	substances, err := substanceRepo.FindAllGroupedWithCreator()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	log.Printf("[Substances] Found %d substances for grouped view", len(substances))

	c.JSON(http.StatusOK, substances)
}

// GetMySubstances 获取我提交的物质
func GetMySubstances(c *gin.Context) {
	userID, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 安全执行转换
	uID := uint(0)
	if u, ok := userID.(int); ok {
		uID = uint(u)
	} else if u, ok := userID.(uint); ok {
		uID = u
	}

	substances, err := substanceRepo.FindMySubstances(uID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, substances)
}

// GetSubstanceNames 获取物质名称映射（用于游戏中显示）
func GetSubstanceNames(c *gin.Context) {
	substanceRepo := repository.NewSubstanceRepository()

	// 获取所有已批准的物质
	substances, err := substanceRepo.FindApproved()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 构建映射表 formula -> name
	nameMap := make(map[string]string)
	for _, sub := range substances {
		if sub.Formula != "" && sub.Name != "" {
			nameMap[sub.Formula] = sub.Name
		}
	}

	c.JSON(http.StatusOK, nameMap)
}

// SubmitNewSubstance 用户提交新物质建议
func SubmitNewSubstance(c *gin.Context) {
	userID, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req SubstanceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	substanceRepo := repository.NewSubstanceRepository()

	// 检查是否已存在相同化学式的物质
	var existingSubstances []database.Substance
	database.DB.Where("formula = ? AND status = ?", req.Formula, "approved").Find(&existingSubstances)

	var groupID *uint
	if len(existingSubstances) > 0 {
		// 如果已存在，使用现有的 group_id
		if existingSubstances[0].GroupID != nil {
			groupID = existingSubstances[0].GroupID
		} else {
			// 如果现有物质没有 group_id，使用其 ID
			gid := existingSubstances[0].ID
			groupID = &gid
			// 更新现有物质的 group_id
			database.DB.Model(&database.Substance{}).Where("id = ?", existingSubstances[0].ID).Update("group_id", gid)
		}
	}

	// 解析元素（如果未提供）
	elements := req.Elements
	if elements == "" {
		elementsMap := game.ParseSubstanceForElements(req.Formula)
		var elementsArr []string
		for e := range elementsMap {
			elementsArr = append(elementsArr, e)
		}
		elements = fmt.Sprintf("%v", elementsArr)
	}

	// 安全执行转换
	uID := uint(0)
	if u, ok := userID.(int); ok {
		uID = uint(u)
	} else if u, ok := userID.(uint); ok {
		uID = u
	}

	// 创建新的物质建议
	newSubstance := &database.Substance{
		Name:         req.Name,
		Formula:      req.Formula,
		Elements:     elements,
		CreatedByUID: uID,
		Status:       "pending",
		GroupID:      groupID,
	}

	if err := substanceRepo.Create(newSubstance); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交建议失败"})
		return
	}

	// 如果是新建议且有现有物质，标记需要完善
	if len(existingSubstances) > 0 && groupID != nil {
		substanceRepo.MarkNeedsImprovement(*groupID, true)
	} else if groupID == nil {
		// 如果是全新物质，使用新创建的 ID 作为 group_id
		gid := newSubstance.ID
		database.DB.Model(&database.Substance{}).Where("id = ?", newSubstance.ID).Update("group_id", gid)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "物质建议已提交，等待管理员审核",
		"id":      newSubstance.ID,
	})
}

// RegisterDataRoutes 注册数据管理路由
func RegisterDataRoutes(r *gin.Engine) {
	// 公开路由 - 物质名称映射
	r.GET("/api/substances/names", GetSubstanceNames)

	data := r.Group("/api/data")
	data.Use(middleware.AuthMiddleware())
	{
		// 物质管理
		data.GET("/substances", GetAllSubstancesGrouped)
		data.GET("/substances/my", GetMySubstances)
		data.GET("/substances/:id/group", GetSubstancesByGroup)
		data.POST("/substances/:id/update", SubmitSubstanceUpdate)
		data.POST("/substances/new", SubmitNewSubstance) // 新增：提交新物质建议

		// 协作者/管理员专用
		coworker := data.Group("")
		coworker.Use(middleware.CoWorkerMiddleware())
		{
			coworker.PUT("/substances/:id", AdminUpdateSubstance)
			coworker.POST("/substances/:id/approve", ApproveSubstanceUpdate)
			coworker.DELETE("/substances/:id/reject", RejectSubstanceUpdate)
		}
	}
}
