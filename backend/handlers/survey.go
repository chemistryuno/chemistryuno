package handlers

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// CreateSurvey 管理员发布新问卷 (内部系统)
func CreateSurvey(c *gin.Context) {
	var req database.Survey
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}

	uid := c.GetInt("uid")
	req.CreatedBy = uint(uid)
	req.IsActive = true

	if err := repository.SurveyRepo.CreateWithQuestions(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发布问卷失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "问卷已发布", "id": req.ID})
}

// GetSurveys 获取所有问卷 (包含题目)
func GetSurveys(c *gin.Context) {
	surveys, err := repository.SurveyRepo.GetAll(false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取问卷列表失败"})
		return
	}
	c.JSON(http.StatusOK, surveys)
}

// GetSurveyDetail 获取单个问卷详情 (包含题目)
func GetSurveyDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	survey, err := repository.SurveyRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "问卷不存在"})
		return
	}
	c.JSON(http.StatusOK, survey)
}

// UpdateSurveyStatus 启用/禁用问卷
func UpdateSurveyStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}

	survey, err := repository.SurveyRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "问卷不存在"})
		return
	}

	survey.IsActive = req.IsActive
	if err := repository.SurveyRepo.Update(survey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "问卷状态已更新"})
}

// DeleteSurvey 管理员删除问卷
func DeleteSurvey(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	// 级联删除在 Repository 中处理
	if err := repository.SurveyRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除问卷失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "问卷已永久删除"})
}

// GetActiveSurveysForUser 获取当前用户需要填写的问卷 (主页弹窗)
func GetActiveSurveysForUser(c *gin.Context) {
	uid := c.GetInt("uid")
	surveys, err := repository.SurveyRepo.GetPendingForUser(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取问卷失败"})
		return
	}
	c.JSON(http.StatusOK, surveys)
}

// SubmitSurveyResponse 提交问卷答案
func SubmitSurveyResponse(c *gin.Context) {
	uid := c.GetInt("uid")
	idStr := c.Param("id")
	surveyID, _ := strconv.ParseUint(idStr, 10, 32)

	var req struct {
		Answers []struct {
			QuestionID uint   `json:"question_id"`
			Answer     string `json:"answer"` // JSON格式文本
		} `json:"answers"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}

	response := &database.SurveyResponse{
		SurveyID: uint(surveyID),
		UserUID:  uint(uid),
	}

	for _, a := range req.Answers {
		response.Answers = append(response.Answers, database.SurveyAnswer{
			QuestionID: a.QuestionID,
			Answer:     a.Answer,
		})
	}

	if err := repository.SurveyRepo.SubmitResponse(response); err != nil {
		if err == gorm.ErrInvalidData {
			c.JSON(http.StatusConflict, gin.H{"error": "您已经完成过此问卷了，不能重复提交"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交答案失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "答案提交成功，感谢您的反馈！"})
}

// DismissSurvey 玩家选择不再提醒某个问卷
func DismissSurvey(c *gin.Context) {
	uid := c.GetInt("uid")
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	if err := repository.SurveyRepo.Dismiss(uint(uid), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已永久忽略该问卷提醒"})
}

// GetAllActiveSurveys 获取所有激活问卷 (反馈页面展示，排除已完成的，包含已忽略的)
func GetAllActiveSurveys(c *gin.Context) {
	uid := c.GetInt("uid")
	surveys, err := repository.SurveyRepo.GetAllActiveForUser(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取问卷失败"})
		return
	}
	c.JSON(http.StatusOK, surveys)
}

// GetSurveyResponses 管理员查看问卷提交信息 (支持UID和时间排序)
func GetSurveyResponses(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	sortBy := c.DefaultQuery("sort_by", "created_at") // user_uid or created_at
	order := c.DefaultQuery("order", "desc")          // asc or desc

	if sortBy != "user_uid" && sortBy != "created_at" {
		sortBy = "created_at"
	}
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	responses, err := repository.SurveyRepo.GetResponsesBySurveyIDSorted(uint(id), sortBy, order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取答卷数据失败"})
		return
	}

	c.JSON(http.StatusOK, responses)
}

// UpdateSurvey 管理员修改问卷（标题、描述、奖励、题目顺序和必填状态）
func UpdateSurvey(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	type questionReq struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Type        string `json:"type"`
		Options     string `json:"options"` // 前端传来的 JSON 字符串，如 "[\"A\",\"B\"]"
		IsRequired  bool   `json:"is_required"`
		Order       int    `json:"order"`
	}
	var req struct {
		Title        string        `json:"title"`
		Description  string        `json:"description"`
		RewardPoints int           `json:"reward_points"`
		RewardExp    int           `json:"reward_exp"`
		Questions    []questionReq `json:"questions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}

	updateData := map[string]interface{}{
		"description":   req.Description,
		"reward_points": req.RewardPoints,
		"reward_exp":    req.RewardExp,
	}
	if req.Title != "" {
		updateData["title"] = req.Title
	}

	questions := make([]database.SurveyQuestion, 0, len(req.Questions))
	for _, q := range req.Questions {
		sq := database.SurveyQuestion{
			Title:       q.Title,
			Description: q.Description,
			Type:        q.Type,
			IsRequired:  q.IsRequired,
			Order:       q.Order,
		}
		// Options 是前端传来的 JSON 字符串（如 `["A","B"]`），
		// 需要重新编码为 JSON 字符串值（加外层引号），与 CreateSurvey 的存储格式保持一致
		if q.Options != "" {
			encoded, err := json.Marshal(q.Options)
			if err == nil {
				sq.Options = database.JSON(encoded)
			}
		}
		questions = append(questions, sq)
	}

	if err := repository.SurveyRepo.PartialUpdateWithQuestions(uint(id), updateData, questions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新问卷失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "问卷已更新"})
}

// RepairSurveyAnswers 修复问卷中 question_id=0 的历史答案（按位置推断）
func RepairSurveyAnswers(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	fixed, err := repository.SurveyRepo.RepairAnswerQuestionIDs(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "修复失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     fmt.Sprintf("修复完成，共修复 %d 条答案", fixed),
		"fixed_count": fixed,
	})
}

// GetSurveyConfig 获取问卷的完整配置 (用于导出 JSON)
func GetSurveyConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	survey, err := repository.SurveyRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "问卷不存在"})
		return
	}

	c.JSON(http.StatusOK, survey)
}

// ImportSurveyConfig 导入问卷配置 (通过 JSON)
func ImportSurveyConfig(c *gin.Context) {
	var req database.Survey
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON 格式无效"})
		return
	}

	uid := c.GetInt("uid")
	req.CreatedBy = uint(uid)
	req.ID = 0 // 确保是创建新问卷

	// 清理题目 ID 以便重新创建
	for i := range req.Questions {
		req.Questions[i].ID = 0
		req.Questions[i].SurveyID = 0
	}

	if err := repository.SurveyRepo.CreateWithQuestions(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导入失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "问卷导入成功", "id": req.ID})
}

// ExportSurveyResponses 导出问卷答卷到 Excel
func ExportSurveyResponses(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	survey, err := repository.SurveyRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "问卷不存在"})
		return
	}

	responses, err := repository.SurveyRepo.GetFullResponsesBySurveyID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取答卷数据失败"})
		return
	}

	f := excelize.NewFile()
	sheetName := "问卷详情"
	f.SetSheetName("Sheet1", sheetName)

	// 设置表头: 用户ID, 提交时间, 题目1, 题目2...
	headers := []string{"用户ID", "提交时间 (UTC+0)"}
	for _, q := range survey.Questions {
		headers = append(headers, fmt.Sprintf("%d.%s (%s)", q.Order, q.Title, q.Type))
	}

	for i, h := range headers {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheetName, colName+"1", h)
	}

	// 填充数据
	for rowIndex, res := range responses {
		rowNum := rowIndex + 2
		f.SetCellValue(sheetName, "A"+strconv.Itoa(rowNum), res.UserUID)
		f.SetCellValue(sheetName, "B"+strconv.Itoa(rowNum), res.CreatedAt.Format("2006-01-02 15:04:05"))

		// 匹配答案
		ansMap := make(map[uint]string)
		for _, ans := range res.Answers {
			ansMap[ans.QuestionID] = ans.Answer
		}

		for qIndex, q := range survey.Questions {
			colName, _ := excelize.ColumnNumberToName(qIndex + 3)
			ansText := ansMap[q.ID]
			// 如果是多选，美化一下 (假设存储的是JSON数组字符串)
			if q.Type == "checkbox" && strings.HasPrefix(ansText, "[") {
				ansText = strings.ReplaceAll(ansText, "\"", "")
				ansText = strings.Trim(ansText, "[]")
			}
			f.SetCellValue(sheetName, colName+strconv.Itoa(rowNum), ansText)
		}
	}

	// 导出并作为附件返回
	fileName := fmt.Sprintf("survey_%d_%s.xlsx", survey.ID, survey.Title)
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+fileName)

	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导出文件失败"})
	}
}
