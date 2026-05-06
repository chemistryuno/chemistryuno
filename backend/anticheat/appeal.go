package anticheat

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"
)

// AppealManager 申诉管理器
type AppealManager struct {
	repository *repository.CheatRepository
	userRepo   *repository.UserRepository
}

// ApprovalOutcome describes the result of approval and optional compensation.
type ApprovalOutcome struct {
	Appeal              *database.CheatAppeal
	CompensationAmount  int
	CompensationMessage string
	CompensationStatus  string
	CompensationNote    string
	Idempotent          bool
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
	return am.SubmitAppealWithRooms(roomID, playerUID, riskScoreID, sanctionID, reason, evidence, nil)
}

func (am *AppealManager) SubmitAppealWithRooms(roomID string, playerUID uint, riskScoreID uint, sanctionID *uint, reason, evidence string, roomIDs []string) (*database.CheatAppeal, error) {
	hasPending, err := am.repository.HasPendingAppealForContext(playerUID, riskScoreID, sanctionID)
	if err != nil {
		return nil, err
	}
	if hasPending {
		return nil, errors.New("pending appeal already exists for this sanction context")
	}

	appeal := &database.CheatAppeal{
		RoomID:      roomID,
		PlayerUID:   playerUID,
		RiskScoreID: riskScoreID,
		Reason:      reason,
		Evidence:    evidence,
		Status:      "pending",
	}
	if riskScoreID > 0 {
		if score, err := am.repository.GetRiskScoreByID(riskScoreID); err == nil && score != nil {
			appeal.ReplayID = score.ReplayID
			appeal.GameHistoryID = score.GameHistoryID
			appeal.PrimaryEvidence = score.PrimaryEvidence
			appeal.RelatedEvidence = score.RelatedEvidence
		}
	}
	if len(roomIDs) > 0 {
		lockedRooms, _ := json.Marshal(roomIDs)
		appeal.RoomIDs = lockedRooms
	}

	if sanctionID != nil {
		appeal.SanctionID = *sanctionID
	}

	if err := am.repository.SaveAppeal(appeal); err != nil {
		log.Printf("[申诉] 提交申诉失败: %v", err)
		return nil, err
	}

	log.Printf("[申诉] 玩家 %d 对房间 %s 提交申诉 (ID: %d)", playerUID, roomID, appeal.ID)
	auditLog := &database.CheatAuditLog{
		EventType:       "appeal",
		RoomID:          roomID,
		PlayerUID:       playerUID,
		RiskScoreID:     &riskScoreID,
		AppealID:        &appeal.ID,
		ReplayID:        appeal.ReplayID,
		GameHistoryID:   appeal.GameHistoryID,
		PrimaryEvidence: appeal.PrimaryEvidence,
		RelatedEvidence: appeal.RelatedEvidence,
		NewStatus:       "pending",
		Remark:          reason,
	}
	if sanctionID != nil {
		auditLog.SanctionID = sanctionID
	}
	if err := am.repository.SaveAuditLog(auditLog); err != nil {
		log.Printf("[申诉] 记录申诉提交审计失败: %v", err)
	}
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
	_, err := am.ApproveAppealWithCompensation(appealID, reviewerUID, remark, config.UnbanConfig.CompensationAmount, config.UnbanConfig.DefaultMessage, sanctionDecider)
	return err
}

// ApproveAppealWithCompensation approves an appeal and records compensation outcome.
func (am *AppealManager) ApproveAppealWithCompensation(appealID uint, reviewerUID uint, remark string, compensationAmount int, compensationMessage string, sanctionDecider *SanctionDecider) (*ApprovalOutcome, error) {
	appeal, err := am.repository.GetAppealByID(appealID)
	if err != nil {
		log.Printf("[申诉] 查询申诉失败: %v", err)
		return nil, err
	}

	if appeal.Status == "approved" && appeal.CompensationStatus == "ok" {
		if err := am.clearApprovedAppealBanState(appeal); err != nil {
			return nil, err
		}
		return &ApprovalOutcome{
			Appeal:              appeal,
			CompensationAmount:  appeal.CompensationAmount,
			CompensationMessage: compensationMessage,
			CompensationStatus:  "ok",
			CompensationNote:    "补偿已领取",
			Idempotent:          true,
		}, nil
	}

	// 更新申诉状态
	if err := am.repository.UpdateAppealStatus(appealID, "approved", &reviewerUID, remark); err != nil {
		log.Printf("[申诉] 更新申诉状态失败: %v", err)
		return nil, err
	}

	// 如果有关联的处罚，需要撤销
	if appeal.SanctionID > 0 && sanctionDecider != nil {
		if err := sanctionDecider.RevokeSanction(appeal.SanctionID); err != nil {
			log.Printf("[申诉] 撤销处罚失败: %v", err)
			return nil, err
		}
	}

	if am.userRepo != nil {
		if err := am.userRepo.UpdateBanStatusWithReason(appeal.PlayerUID, nil, ""); err != nil {
			log.Printf("[appeal] failed to clear account ban for player %d: %v", appeal.PlayerUID, err)
			return nil, err
		}
	}
	if err := am.revokeApprovedAppealActiveBanSanctions(appeal); err != nil {
		return nil, err
	}

	now := time.Now()
	pendingStatus := "pending"
	approvalNote := remark
	auditLog := &database.CheatAuditLog{
		EventType:           "review",
		RoomID:              appeal.RoomID,
		PlayerUID:           appeal.PlayerUID,
		OperatorUID:         &reviewerUID,
		AppealID:            &appealID,
		ReplayID:            appeal.ReplayID,
		GameHistoryID:       appeal.GameHistoryID,
		PrimaryEvidence:     appeal.PrimaryEvidence,
		RelatedEvidence:     appeal.RelatedEvidence,
		OldStatus:           appeal.Status,
		NewStatus:           "approved",
		Remark:              remark,
		ApprovalNote:        &approvalNote,
		CompensationAmount:  &compensationAmount,
		CompensationStatus:  &pendingStatus,
		CompensationMessage: &compensationMessage,
		CompensationDate:    &now,
	}
	if err := am.repository.SaveAuditLog(auditLog); err != nil {
		log.Printf("[申诉] 记录审批审计日志失败: %v", err)
	}

	outcome := &ApprovalOutcome{
		Appeal:              appeal,
		CompensationAmount:  compensationAmount,
		CompensationMessage: compensationMessage,
		CompensationStatus:  "pending",
		CompensationNote:    "补偿待领取",
	}
	if updateErr := am.repository.UpdateAppealCompensation(appealID, outcome.CompensationStatus, compensationAmount, outcome.CompensationNote); updateErr != nil {
		log.Printf("[申诉] 更新补偿状态失败: %v", updateErr)
	}
	outcome.Appeal.Status = "approved"
	outcome.Appeal.CompensationStatus = outcome.CompensationStatus
	outcome.Appeal.CompensationAmount = outcome.CompensationAmount
	outcome.Appeal.CompensationNote = outcome.CompensationNote
	log.Printf("[申诉] 已批准申诉 ID: %d (玩家: %d，审核人: %d)，补偿待玩家主动领取", appealID, appeal.PlayerUID, reviewerUID)
	return outcome, nil
}

func (am *AppealManager) revokeApprovedAppealActiveBanSanctions(appeal *database.CheatAppeal) error {
	if am == nil || am.repository == nil || appeal == nil {
		return nil
	}
	if err := am.repository.RevokeActiveBanSanctionsByPlayer(appeal.PlayerUID); err != nil {
		log.Printf("[appeal] failed to revoke active ban sanctions for player %d: %v", appeal.PlayerUID, err)
		return err
	}
	return nil
}

func (am *AppealManager) clearApprovedAppealBanState(appeal *database.CheatAppeal) error {
	if appeal == nil {
		return nil
	}
	if am.userRepo != nil {
		if err := am.userRepo.UpdateBanStatusWithReason(appeal.PlayerUID, nil, ""); err != nil {
			log.Printf("[appeal] failed to clear account ban for player %d: %v", appeal.PlayerUID, err)
			return err
		}
	}
	return am.revokeApprovedAppealActiveBanSanctions(appeal)
}

func (am *AppealManager) ClaimCompensation(appealID uint, playerUID uint) (*ApprovalOutcome, error) {
	appeal, err := am.repository.GetAppealByID(appealID)
	if err != nil {
		return nil, err
	}
	if appeal.PlayerUID != playerUID {
		return nil, errors.New("appeal not found")
	}
	if appeal.Status != "approved" {
		return nil, errors.New("appeal is not approved")
	}
	if appeal.CompensationAmount <= 0 {
		return nil, errors.New("no compensation is available")
	}

	outcome := &ApprovalOutcome{
		Appeal:             appeal,
		CompensationAmount: appeal.CompensationAmount,
		CompensationStatus: appeal.CompensationStatus,
	}
	if appeal.CompensationStatus == "ok" {
		outcome.CompensationNote = "补偿已领取"
		outcome.Idempotent = true
		return outcome, nil
	}

	compensationID := fmt.Sprintf("appeal_%d", appealID)
	if _, err := am.userRepo.AddFuel(appeal.PlayerUID, appeal.CompensationAmount, compensationID); err != nil {
		if errors.Is(err, repository.ErrFuelCompensationAlreadyIssued) {
			outcome.CompensationStatus = "ok"
			outcome.CompensationNote = "补偿已发放过，本次请求按幂等成功处理"
			outcome.Idempotent = true
		} else {
			outcome.CompensationStatus = "failed"
			outcome.CompensationNote = fmt.Sprintf("补偿失败: %v", err)
		}
		log.Printf("[申诉] 燃素补偿领取失败 (玩家: %d, 金额: %d): %v", appeal.PlayerUID, appeal.CompensationAmount, err)
	} else {
		outcome.CompensationStatus = "ok"
		outcome.CompensationNote = "补偿领取成功"
		log.Printf("[申诉] 玩家已领取燃素补偿 (玩家: %d, 金额: %d)", appeal.PlayerUID, appeal.CompensationAmount)
	}

	if updateErr := am.repository.UpdateAppealCompensation(appealID, outcome.CompensationStatus, appeal.CompensationAmount, outcome.CompensationNote); updateErr != nil {
		log.Printf("[申诉] 更新补偿领取状态失败: %v", updateErr)
	}
	outcome.Appeal.CompensationStatus = outcome.CompensationStatus
	outcome.Appeal.CompensationNote = outcome.CompensationNote

	approvalNote := "player claimed compensation"
	now := time.Now()
	auditLog := &database.CheatAuditLog{
		EventType:          "compensation_claim",
		RoomID:             appeal.RoomID,
		PlayerUID:          appeal.PlayerUID,
		AppealID:           &appealID,
		ReplayID:           appeal.ReplayID,
		GameHistoryID:      appeal.GameHistoryID,
		PrimaryEvidence:    appeal.PrimaryEvidence,
		RelatedEvidence:    appeal.RelatedEvidence,
		NewStatus:          outcome.CompensationStatus,
		Remark:             outcome.CompensationNote,
		ApprovalNote:       &approvalNote,
		CompensationAmount: &appeal.CompensationAmount,
		CompensationStatus: &outcome.CompensationStatus,
		CompensationNote:   &outcome.CompensationNote,
		CompensationDate:   &now,
	}
	if err := am.repository.SaveAuditLog(auditLog); err != nil {
		log.Printf("[申诉] 记录补偿领取审计失败: %v", err)
	}

	return outcome, nil
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
