package anticheat

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"
	"log"
	"time"
)

// SanctionDecider 处罚决策器
type SanctionDecider struct {
	config     *RiskScoringConfig
	repository *repository.CheatRepository
	userRepo   *repository.UserRepository
}

// NewSanctionDecider 创建处罚决策器
func NewSanctionDecider(config *RiskScoringConfig, repo *repository.CheatRepository) *SanctionDecider {
	return &SanctionDecider{
		config:     config,
		repository: repo,
		userRepo:   repository.NewUserRepository(),
	}
}

// Decision 处罚决策
type Decision struct {
	SanctionType string    // "none", "observe", "warning", "mute", "ban"
	RiskScore    float64
	Reason       string
	Duration     *int       // 处罚时长(分钟)
	EffectiveUntil *time.Time // 处罚有效期
}

// MakeDecision 根据风险评分和上下文做出处罚决策
func (sd *SanctionDecider) MakeDecision(riskScore float64, roomID string, playerUID uint, riskScoreID uint) *Decision {
	decision := &Decision{
		RiskScore: riskScore,
	}

	thresholds := sd.config.SanctionThresholds

	switch {
	case riskScore >= thresholds.BanMin:
		decision.SanctionType = "ban"
		decision.Reason = "检测到严重的游戏作弊行为（高风险）"
		// 设置封号时长（示例：7天）
		duration := 10080 // 7天（分钟）
		decision.Duration = &duration
		effectiveUntil := time.Now().Add(time.Duration(duration) * time.Minute)
		decision.EffectiveUntil = &effectiveUntil

	case riskScore >= thresholds.MuteMin:
		decision.SanctionType = "mute"
		decision.Reason = "检测到游戏作弊行为（禁言）"
		// 设置禁言时长（示例：1小时）
		duration := 60
		decision.Duration = &duration
		effectiveUntil := time.Now().Add(time.Duration(duration) * time.Minute)
		decision.EffectiveUntil = &effectiveUntil

	case riskScore >= thresholds.WarningMin:
		decision.SanctionType = "warning"
		decision.Reason = "检测到可疑的游戏行为（警告）"

	case riskScore >= thresholds.ObserveMin:
		decision.SanctionType = "observe"
		decision.Reason = "记录游戏行为异常（观察中）"

	default:
		decision.SanctionType = "none"
		decision.Reason = "无异常"
	}

	log.Printf("[处罚决策] 房间 %s 玩家 %d: 风险分数 %.1f → 处罚: %s",
		roomID, playerUID, riskScore, decision.SanctionType)

	return decision
}

// ApplySanction 应用处罚决策
func (sd *SanctionDecider) ApplySanction(decision *Decision, roomID string, playerUID uint, riskScoreID uint) (*database.CheatSanction, error) {
	if decision.SanctionType == "none" || decision.SanctionType == "observe" {
		// 不需要实际应用处罚，仅记录日志
		return nil, nil
	}

	sanction := &database.CheatSanction{
		RoomID:      roomID,
		PlayerUID:   playerUID,
		RiskScoreID: riskScoreID,
		SanctionType: decision.SanctionType,
		RiskScore:   decision.RiskScore,
		Reason:      decision.Reason,
		Duration:    decision.Duration,
		EffectiveUntil: decision.EffectiveUntil,
		Status:      "active",
	}

	// 保存到数据库
	if err := sd.repository.SaveSanction(sanction); err != nil {
		log.Printf("[处罚] 保存处罚记录失败: %v", err)
		return nil, err
	}

	log.Printf("[处罚] 已应用处罚 %s 给玩家 %d（房间 %s）",
		decision.SanctionType, playerUID, roomID)

	if sanction.SanctionType == "ban" && sanction.EffectiveUntil != nil && sd.userRepo != nil {
		if err := sd.userRepo.UpdateBanStatusWithReason(playerUID, sanction.EffectiveUntil, decision.Reason); err != nil {
			log.Printf("[sanction] failed to apply account ban for player %d: %v", playerUID, err)
			return nil, err
		}
	}

	return sanction, nil
}

// RevokeSanction 撤销处罚
func (sd *SanctionDecider) RevokeSanction(sanctionID uint) error {
	sanction, err := sd.repository.GetSanctionByID(sanctionID)
	if err != nil {
		log.Printf("[sanction] failed to load sanction before revoke: %v", err)
		return err
	}
	if err := sd.repository.UpdateSanctionStatus(sanctionID, "revoked"); err != nil {
		log.Printf("[处罚] 撤销处罚失败: %v", err)
		return err
	}

	log.Printf("[处罚] 已撤销处罚 ID: %d", sanctionID)
	if sanction.SanctionType == "ban" && sd.userRepo != nil {
		if err := sd.userRepo.UpdateBanStatusWithReason(sanction.PlayerUID, nil, ""); err != nil {
			log.Printf("[sanction] failed to clear account ban for player %d: %v", sanction.PlayerUID, err)
			return err
		}
	}
	return nil
}

// CheckMutedStatus 检查玩家是否被禁言
func (sd *SanctionDecider) CheckMutedStatus(playerUID uint) (bool, *database.CheatSanction) {
	sanctions, err := sd.repository.GetActiveSanctionsByPlayer(playerUID)
	if err != nil {
		log.Printf("[处罚检查] 查询处罚失败: %v", err)
		return false, nil
	}

	now := time.Now()
	for i := range sanctions {
		if sanctions[i].SanctionType == "mute" && sanctions[i].Status == "active" {
			// 检查是否已过期
			if sanctions[i].EffectiveUntil != nil && sanctions[i].EffectiveUntil.Before(now) {
				// 处罚已过期，更新状态
				if err := sd.repository.UpdateSanctionStatus(sanctions[i].ID, "expired"); err != nil {
					log.Printf("[处罚检查] 更新处罚状态失败: %v", err)
				}
				continue
			}
			return true, &sanctions[i]
		}
	}

	return false, nil
}

// SendWarningNotification 发送警告通知（由实际的消息系统实现）
func (sd *SanctionDecider) SendWarningNotification(playerUID uint, reason string) error {
	// 这是一个占位符，实际的通知系统应该在其他地方实现
	log.Printf("[通知] 发送警告给玩家 %d: %s", playerUID, reason)
	return nil
}

// GetActiveSanctionsForPlayer 获取玩家当前的活跃处罚
func (sd *SanctionDecider) GetActiveSanctionsForPlayer(playerUID uint) ([]database.CheatSanction, error) {
	sanctions, err := sd.repository.GetActiveSanctionsByPlayer(playerUID)
	if err != nil {
		return nil, err
	}

	// 清理过期的处罚
	now := time.Now()
	activeSanctions := make([]database.CheatSanction, 0)
	for i := range sanctions {
		if sanctions[i].EffectiveUntil != nil && sanctions[i].EffectiveUntil.Before(now) {
			if err := sd.repository.UpdateSanctionStatus(sanctions[i].ID, "expired"); err != nil {
				log.Printf("[处罚] 更新过期处罚状态失败: %v", err)
			}
		} else {
			activeSanctions = append(activeSanctions, sanctions[i])
		}
	}

	return activeSanctions, nil
}

// CleanupExpiredSanctions 清理过期的处罚记录
func (sd *SanctionDecider) CleanupExpiredSanctions() error {
	// 这是一个定期维护任务，应该由后台服务调用
	// 具体实现需要添加到数据库查询中
	log.Printf("[处罚] 清理过期处罚记录（待实现）")
	return nil
}
