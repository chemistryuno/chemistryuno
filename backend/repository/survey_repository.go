package repository

import (
	"chemistryuno/backend/database"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SurveyRepository struct {
	db *gorm.DB
}

func NewSurveyRepository() *SurveyRepository {
	return &SurveyRepository{db: database.DB}
}

// Admin: 获取所有问卷 (包含题目)
func (r *SurveyRepository) GetAll(activeOnly bool) ([]database.Survey, error) {
	var surveys []database.Survey
	query := database.DB.Preload("Questions", func(db *gorm.DB) *gorm.DB {
		return db.Order("`order` ASC")
	}).Order("created_at DESC")

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}
	err := query.Find(&surveys).Error
	return surveys, err
}

// 获取单个问卷详情 (包含题目)
func (r *SurveyRepository) GetByID(id uint) (*database.Survey, error) {
	var survey database.Survey
	err := database.DB.Preload("Questions", func(db *gorm.DB) *gorm.DB {
		return db.Order("`order` ASC")
	}).First(&survey, id).Error
	if err != nil {
		return nil, err
	}
	return &survey, nil
}

// Admin: 创建问卷 (使用事务处理题目)
func (r *SurveyRepository) CreateWithQuestions(survey *database.Survey) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Create(survey).Error
	})
}

// Admin: 更新问卷
func (r *SurveyRepository) Update(survey *database.Survey) error {
	return database.DB.Save(survey).Error
}

// Admin: 同步更新问卷题目 (全量替换)
func (r *SurveyRepository) PartialUpdateWithQuestions(surveyID uint, updateData map[string]interface{}, questions []database.SurveyQuestion) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 先获取旧题目列表，用于检查是否有正在被使用的引用（如果有强一致性需求，这里可以扩展）
		// 但由于业务逻辑通常是允许管理员重置题目的，我们直接执行删除

		// 更新基本字段
		if len(updateData) > 0 {
			if err := tx.Model(&database.Survey{}).Where("id = ?", surveyID).Updates(updateData).Error; err != nil {
				return err
			}
		}

		// 处理题目逻辑
		if questions != nil {
			// BUG 修复: 在某些数据库（如 PostgreSQL 或某些版本的 GORM 配置）下，
			// 如果 Survey 和 SurveyQuestion 之间配置了级联删除或物理外键，
			// 直接 Delete 可能由于正在运行的事务或索引问题导致失败。
			// 我们使用 Unscoped() 确保能清理可能存在的软删除，并显式指定删除条件。
			if err := tx.Unscoped().Where("survey_id = ?", surveyID).Delete(&database.SurveyQuestion{}).Error; err != nil {
				return err
			}

			// 重新插入新题目
			for i := range questions {
				questions[i].SurveyID = surveyID
				questions[i].ID = 0 // 确保生成新的自增 ID

				// 修复: 显式处理题目数据，避免 GORM 因 ID 冲突跳过插入
				if err := tx.Omit("id").Create(&questions[i]).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// Admin: 删除问卷 (级联删除题目和答案)
func (r *SurveyRepository) Delete(id uint) error {
	return database.DB.Delete(&database.Survey{}, id).Error
}

// User: 提交答卷
func (r *SurveyRepository) SubmitResponse(response *database.SurveyResponse) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 检查是否已提交过
		var count int64
		tx.Model(&database.SurveyResponse{}).Where("survey_id = ? AND user_uid = ?", response.SurveyID, response.UserUID).Count(&count)
		if count > 0 {
			return gorm.ErrInvalidData // 简单返回一个错误表示已存在
		}

		// 获取问卷奖励信息
		var survey database.Survey
		if err := tx.First(&survey, response.SurveyID).Error; err != nil {
			return err
		}

		// 保存回答内容
		if err := tx.Create(response).Error; err != nil {
			return err
		}

		// 发放积分奖励 (如果设置了的话)
		if survey.RewardPoints > 0 {
			if err := tx.Model(&database.User{}).Where("uid = ?", response.UserUID).
				UpdateColumn("points", gorm.Expr("points + ?", survey.RewardPoints)).Error; err != nil {
				return err
			}
		}

		// 发放经验值奖励 (如果设置了的话)
		if survey.RewardExp > 0 {
			if err := tx.Model(&database.User{}).Where("uid = ?", response.UserUID).
				UpdateColumns(map[string]interface{}{
					"xp":       gorm.Expr("xp + ?", survey.RewardExp),
					"total_xp": gorm.Expr("total_xp + ?", survey.RewardExp),
				}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetPendingForUser 获取待进行的活动问卷 (排除已完成和已忽略)
func (r *SurveyRepository) GetPendingForUser(userUID uint) ([]database.Survey, error) {
	var surveys []database.Survey

	dismissedIDs := database.DB.Model(&database.SurveyDismissal{}).Select("survey_id").Where("user_uid = ?", userUID)
	respondedIDs := database.DB.Model(&database.SurveyResponse{}).Select("survey_id").Where("user_uid = ?", userUID)

	err := database.DB.Where("is_active = ?", true).
		Where("id NOT IN (?)", dismissedIDs).
		Where("id NOT IN (?)", respondedIDs).
		Order("created_at DESC").
		Find(&surveys).Error
	return surveys, err
}

// GetAllActiveForUser 获取用户所有激活问卷 (排除已完成的，但包含已忽略的)
func (r *SurveyRepository) GetAllActiveForUser(userUID uint) ([]database.Survey, error) {
	var surveys []database.Survey

	respondedIDs := database.DB.Model(&database.SurveyResponse{}).Select("survey_id").Where("user_uid = ?", userUID)

	err := database.DB.Where("is_active = ?", true).
		Where("id NOT IN (?)", respondedIDs).
		Order("created_at DESC").
		Find(&surveys).Error
	return surveys, err
}

// User: 永久忽略问卷提醒
func (r *SurveyRepository) Dismiss(userUID, surveyID uint) error {
	dismissal := &database.SurveyDismissal{
		UserUID:  userUID,
		SurveyID: surveyID,
	}
	return database.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(dismissal).Error
}

// Admin: 获取问卷的所有答卷
func (r *SurveyRepository) GetFullResponsesBySurveyID(surveyID uint) ([]database.SurveyResponse, error) {
	var responses []database.SurveyResponse
	err := database.DB.Preload("Answers").Where("survey_id = ?", surveyID).Order("created_at DESC").Find(&responses).Error
	return responses, err
}

// Admin: 修复答案中 question_id=0 的记录（按插入顺序与题目顺序位置对应）
// 返回修复的答案数量
func (r *SurveyRepository) RepairAnswerQuestionIDs(surveyID uint) (int, error) {
	// 获取该问卷所有题目（按order升序）
	var questions []database.SurveyQuestion
	if err := database.DB.Where("survey_id = ?", surveyID).Order("`order` ASC").Find(&questions).Error; err != nil {
		return 0, err
	}
	if len(questions) == 0 {
		return 0, nil
	}

	// 获取该问卷所有答卷，答案按 id 升序（即插入顺序）
	var responses []database.SurveyResponse
	// 注意：这里必须保证 Preload 加载的 Answers 顺序是创建时的顺序（ID ASC）
	if err := database.DB.Preload("Answers", func(db *gorm.DB) *gorm.DB {
		return db.Order("id ASC")
	}).Where("survey_id = ?", surveyID).Find(&responses).Error; err != nil {
		return 0, err
	}

	fixedCount := 0
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		for _, res := range responses {
			// 如果该份答卷的答案总数与题目总数不符，无法安全推断位置关系
			if len(res.Answers) != len(questions) {
				continue
			}

			for i, a := range res.Answers {
				// 核心修复：即使 question_id 不等于 0，但如果它不属于该问卷题目列表（老数据残留），也进行重定向
				belongs := false
				for _, q := range questions {
					if a.QuestionID == q.ID {
						belongs = true
						break
					}
				}

				if a.QuestionID == 0 || !belongs {
					if err := tx.Model(&database.SurveyAnswer{}).Where("id = ?", a.ID).
						Update("question_id", questions[i].ID).Error; err != nil {
						return err
					}
					fixedCount++
				}
			}
		}
		return nil
	})
	return fixedCount, err
}
func (r *SurveyRepository) GetResponsesBySurveyIDSorted(surveyID uint, sortBy, order string) ([]database.SurveyResponse, error) {
	var responses []database.SurveyResponse
	orderClause := sortBy + " " + order
	err := database.DB.Preload("Answers").Where("survey_id = ?", surveyID).Order(orderClause).Find(&responses).Error
	return responses, err
}
