package anticheat

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"
	"log"
	"time"

	"gorm.io/gorm"
)

// System 反作弊系统总管
type System struct {
	Engine          *RiskScoringEngine
	Config          *ConfigManager
	Decider         *SanctionDecider
	AppealManager   *AppealManager
	AuditLogger     *AuditLogger
	Repository      *repository.CheatRepository
}

// NewSystem 创建反作弊系统实例
func NewSystem(db *gorm.DB, configPath string) (*System, error) {
	repo := repository.NewCheatRepository(db)

	// 初始化配置管理器
	configMgr, err := NewConfigManager(configPath)
	if err != nil {
		log.Printf("[反作弊系统] 配置管理器初始化失败: %v", err)
		return nil, err
	}

	// 初始化风险评分引擎
	config := configMgr.GetConfig()
	engine := NewRiskScoringEngine(config)

	// 注册所有内置策略
	strategies := NewBuiltInStrategies()
	if err := strategies.RegisterAll(engine); err != nil {
		log.Printf("[反作弊系统] 策略注册失败: %v", err)
		return nil, err
	}

	// 监听配置变化并更新引擎
	configMgr.Watch(func(cfg *RiskScoringConfig) {
		engine.UpdateConfig(cfg)
		log.Printf("[反作弊系统] 配置已更新")
	})

	// 初始化其他组件
	decider := NewSanctionDecider(config, repo)
	appealMgr := NewAppealManager(repo)
	auditLog := NewAuditLogger(repo)

	system := &System{
		Engine:        engine,
		Config:        configMgr,
		Decider:       decider,
		AppealManager: appealMgr,
		AuditLogger:   auditLog,
		Repository:    repo,
	}

	log.Printf("[反作弊系统] 初始化成功")
	return system, nil
}

// ProcessGameEnd 处理游戏结束的反作弊检测和处罚
func (s *System) ProcessGameEnd(roomID string, playerUID uint, context *DetectionContext) (*RiskScoringResult, *Decision, error) {
	// 1. 计算风险分数
	result, err := s.Engine.CalculateRiskScore(context)
	if err != nil {
		log.Printf("[反作弊] 风险评分失败: %v", err)
		return nil, nil, err
	}

	// 2. 保存风险分数
	riskScore := &database.CheatRiskScore{
		RoomID:         roomID,
		PlayerUID:      playerUID,
		RiskScore:      result.RiskScore,
		ResponseTimeDim: result.Dimensions["response_time"],
		FrequencyDim:   result.Dimensions["frequency"],
		WinRateDim:     result.Dimensions["win_rate"],
		PatternDim:     result.Dimensions["pattern"],
		AccountAgeDim:  result.Dimensions["account_age"],
	}

	if err := s.Repository.SaveRiskScore(riskScore); err != nil {
		log.Printf("[反作弊] 保存风险分数失败: %v", err)
		return result, nil, err
	}

	// 3. 记录审计日志
	if err := s.AuditLogger.LogDetection(roomID, playerUID, riskScore.ID, result.RiskScore, result.Dimensions); err != nil {
		log.Printf("[反作弊] 记录审计日志失败: %v", err)
	}

	// 4. 根据风险分数做出处罚决策
	decision := s.Decider.MakeDecision(result.RiskScore, roomID, playerUID, riskScore.ID)

	// 5. 应用处罚
	if decision.SanctionType != "none" {
		sanction, err := s.Decider.ApplySanction(decision, roomID, playerUID, riskScore.ID)
		if err != nil {
			log.Printf("[反作弊] 应用处罚失败: %v", err)
		} else if sanction != nil {
			// 记录处罚审计日志
			if err := s.AuditLogger.LogSanction(roomID, playerUID, riskScore.ID, sanction.ID,
				decision.SanctionType, decision.Reason); err != nil {
				log.Printf("[反作弊] 记录处罚日志失败: %v", err)
			}
		}

		// 发送通知
		if decision.SanctionType == "warning" {
			if err := s.Decider.SendWarningNotification(playerUID, decision.Reason); err != nil {
				log.Printf("[反作弊] 发送警告通知失败: %v", err)
			}
		}
	}

	return result, decision, nil
}

// GetPlayerStats 获取玩家的反作弊统计信息
func (s *System) GetPlayerStats(playerUID uint) map[string]interface{} {
	sanctions, _ := s.Decider.GetActiveSanctionsForPlayer(playerUID)
	appeals, _ := s.AppealManager.GetPlayerAppeals(playerUID)

	// 统计各类型处罚
	sanctionCount := make(map[string]int)
	for _, sanction := range sanctions {
		if sanction.Status == "active" {
			sanctionCount[sanction.SanctionType]++
		}
	}

	// 统计申诉结果
	appealStats := map[string]int{
		"pending":     0,
		"approved":    0,
		"rejected":    0,
		"under_review": 0,
	}
	for _, appeal := range appeals {
		if status, exists := appealStats[appeal.Status]; exists {
			appealStats[appeal.Status] = status + 1
		}
	}

	return map[string]interface{}{
		"player_uid":       playerUID,
		"active_sanctions": sanctionCount,
		"appeals":          appealStats,
		"total_sanctions":  len(sanctions),
		"total_appeals":    len(appeals),
	}
}

// GetSystemStats 获取系统统计信息
func (s *System) GetSystemStats(startTime, endTime time.Time) (map[string]interface{}, error) {
	stats, err := s.AuditLogger.GetAuditStatistics(startTime, endTime)
	if err != nil {
		return nil, err
	}

	// 添加系统配置信息
	config := s.Config.GetConfig()
	stats["enabled_strategies"] = config.EnabledStrategies
	stats["sanction_thresholds"] = config.SanctionThresholds

	return stats, nil
}

// Shutdown 关闭系统并清理资源
func (s *System) Shutdown() error {
	log.Printf("[反作弊系统] 正在关闭...")
	// 可以在这里添加清理代码
	return nil
}
