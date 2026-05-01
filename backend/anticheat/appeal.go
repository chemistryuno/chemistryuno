package anticheat

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"
	"fmt"
	"log"
	"time"
)

// AppealManager 申诉管理器
type AppealManager struct {
	repository *repository.CheatRepository
	userRepo   *repository.UserRepository
}

// NewAppealManager 创建申诉管理器
func NewAppealManager(repo *repository.CheatRepository, userRepo *repository.UserRepository) *AppealManager {
	return &AppealManager{
		repository: repo,
		userRepo:   userRepo,
	}
}

// SubmitAppeal 提交申诉
func (am *AppealManager) SubmitAppeal(roomID string, playerUID uint, riskScoreID uint, sanctionID *uint, reason, evidence string) (*database.CheatAppeal, error) {
	appeal := &database.CheatAppeal{
		RoomID:      roomID,
		PlayerUID:   playerUID,
		RiskScoreID: riskScoreID,
		Reason:      reason,
		Evidence:    evidence,
		Status:      "pending",
	}

	if sanctionID != nil {
		appeal.SanctionID = *sanctionID
	}

	if err := am.repository.SaveAppeal(appeal); err != nil {
		log.Printf("[申诉] 提交申诉失败: %v", err)
		return nil, err
	}

	log.Printf("[申诉] 玩家 %d 对房间 %s 提交申诉 (ID: %d)", playerUID, roomID, appeal.ID)
	return appeal, nil
}

// GetPendingAppeals 获取待审核的申诉列表
func (am *AppealManager) GetPendingAppeals(limit int) ([]database.CheatAppeal, error) {
	appeals, err := am.repository.GetPendingAppeals(limit)
	if err != nil {
		log.Printf("[申诉] 查询待审核申诉失败: %v", err)
		return nil, err
	}
	return appeals, nil
}

// ApproveAppeal 批准申诉并发放补偿燃素
func (am *AppealManager) ApproveAppeal(appealID uint, reviewerUID uint, remark string, sanctionDecider *SanctionDecider, config *RiskScoringConfig) error {
	appeal, err := am.repository.GetAppealByID(appealID)
	if err != nil {
		log.Printf("[申诉] 查询申诉失败: %v", err)
		return err
	}

	// 更新申诉状态
	if err := am.repository.UpdateAppealStatus(appealID, "approved", &reviewerUID, remark); err != nil {
		log.Printf("[申诉] 更新申诉状态失败: %v", err)
		return err
	}

	// 如果有关联的处罚，需要撤销
	if appeal.SanctionID > 0 {
		if err := sanctionDecider.RevokeSanction(appeal.SanctionID); err != nil {
			log.Printf("[申诉] 撤销处罚失败: %v", err)
			return err
		}
	}

	// 发放燃素补偿
	compensationAmount := config.UnbanConfig.CompensationAmount
	compensationID := fmt.Sprintf("appeal_%d", appealID)

	if _, err := am.userRepo.AddFuel(appeal.PlayerUID, compensationAmount, compensationID); err != nil {
		// 记录补偿失败但不中断解封流程
		log.Printf("[申诉] 燃素补偿失败 (玩家: %d, 金额: %d): %v", appeal.PlayerUID, compensationAmount, err)

		// 更新申诉记录的补偿状态为失败
		if updateErr := am.repository.UpdateAppealCompensation(appealID, "failed", compensationAmount, fmt.Sprintf("补偿失败: %v", err)); updateErr != nil {
			log.Printf("[申诉] 更新补偿状态失败: %v", updateErr)
		}
	} else {
		// 补偿成功，更新申诉记录
		if updateErr := am.repository.UpdateAppealCompensation(appealID, "completed", compensationAmount, "补偿成功"); updateErr != nil {
			log.Printf("[申诉] 更新补偿状态失败: %v", updateErr)
		}
		log.Printf("[申诉] 已发放燃素补偿 (玩家: %d, 金额: %d)", appeal.PlayerUID, compensationAmount)
	}

	log.Printf("[申诉] 已批准申诉 ID: %d (玩家: %d，审核人: %d)", appealID, appeal.PlayerUID, reviewerUID)
	return nil
}

// RejectAppeal 拒绝申诉
func (am *AppealManager) RejectAppeal(appealID uint, reviewerUID uint, remark string) error {
	// 更新申诉状态
	if err := am.repository.UpdateAppealStatus(appealID, "rejected", &reviewerUID, remark); err != nil {
		log.Printf("[申诉] 更新申诉状态失败: %v", err)
		return err
	}

	log.Printf("[申诉] 已拒绝申诉 ID: %d (审核人: %d)", appealID, reviewerUID)
	return nil
}

// GetPlayerAppeals 获取玩家的申诉历史
func (am *AppealManager) GetPlayerAppeals(playerUID uint) ([]database.CheatAppeal, error) {
	appeals, err := am.repository.GetAppealsByPlayer(playerUID)
	if err != nil {
		log.Printf("[申诉] 查询玩家申诉历史失败: %v", err)
		return nil, err
	}
	return appeals, nil
}

// HasPendingAppeal 检查玩家是否有待审核的申诉
func (am *AppealManager) HasPendingAppeal(playerUID uint) (bool, error) {
	appeals, err := am.GetPlayerAppeals(playerUID)
	if err != nil {
		return false, err
	}

	for _, appeal := range appeals {
		if appeal.Status == "pending" || appeal.Status == "under_review" {
			return true, nil
		}
	}

	return false, nil
}

// GetLatestAppeal 获取玩家的最新申诉
func (am *AppealManager) GetLatestAppeal(playerUID uint) (*database.CheatAppeal, error) {
	appeals, err := am.GetPlayerAppeals(playerUID)
	if err != nil {
		return nil, err
	}

	if len(appeals) == 0 {
		return nil, nil
	}

	return &appeals[0], nil
}

// UpdateAppealToUnderReview 将申诉状态更新为审核中
func (am *AppealManager) UpdateAppealToUnderReview(appealID uint) error {
	appeal, err := am.repository.GetAppealByID(appealID)
	if err != nil {
		log.Printf("[申诉] 查询申诉失败: %v", err)
		return err
	}

	// 更新状态为审核中
	if err := am.repository.UpdateAppealStatus(appealID, "under_review", appeal.ReviewerUID, appeal.ReviewRemark); err != nil {
		log.Printf("[申诉] 更新申诉状态失败: %v", err)
		return err
	}

	log.Printf("[申诉] 申诉 ID: %d 已进入审核阶段", appealID)
	return nil
}

// GetAppealStats 获取申诉统计信息
func (am *AppealManager) GetAppealStats() map[string]interface{} {
	// 获取所有待审核申诉
	pendingAppeals, _ := am.GetPendingAppeals(100)

	stats := map[string]interface{}{
		"pending_count": len(pendingAppeals),
		"timestamp":     time.Now(),
	}

	return stats
}
