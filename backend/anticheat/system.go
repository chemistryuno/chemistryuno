package anticheat

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"
	"encoding/json"
	"log"
	"os"
	"time"

	"gorm.io/gorm"
)

// System 反作弊系统总管
type System struct {
	Engine        *RiskScoringEngine
	Config        *ConfigManager
	Decider       *SanctionDecider
	AppealManager *AppealManager
	AuditLogger   *AuditLogger
	Repository    *repository.CheatRepository
	Baselines     *BaselineCollector
	StartedAt     time.Time
}

// NewSystem 创建反作弊系统实例
func NewSystem(db *gorm.DB, configPath string) (*System, error) {
	repo := repository.NewCheatRepository(db)
	userRepo := repository.NewUserRepository()

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
	appealMgr := NewAppealManager(repo, userRepo)
	auditLog := NewAuditLogger(repo)

	system := &System{
		Engine:        engine,
		Config:        configMgr,
		Decider:       decider,
		AppealManager: appealMgr,
		AuditLogger:   auditLog,
		Repository:    repo,
		StartedAt:     time.Now(),
	}
	system.Baselines = NewBaselineCollector(repo, func() AdaptiveThresholdConfig {
		return configMgr.GetConfig().Optimization.AdaptiveThreshold
	})

	log.Printf("[反作弊系统] 初始化成功")
	return system, nil
}

// UptimeDays returns elapsed whole days since the anticheat system started.
func (s *System) UptimeDays(now time.Time) int {
	if s == nil || s.StartedAt.IsZero() || now.Before(s.StartedAt) {
		return 0
	}
	return int(now.Sub(s.StartedAt).Hours() / 24)
}

// ProcessGameEnd 处理游戏结束的反作弊检测和处罚
func (s *System) ProcessGameEnd(roomID string, playerUID uint, context *DetectionContext) (*RiskScoringResult, *Decision, error) {
	// 0. 注入优化特性所需的上下文（仅在对应特性启用时查询数据）。
	s.enrichDetectionContext(playerUID, context)

	// 1. 计算风险分数
	result, err := s.Engine.CalculateRiskScore(context)
	if err != nil {
		log.Printf("[反作弊] 风险评分失败: %v", err)
		return nil, nil, err
	}

	// 2. 保存风险分数
	riskScore := &database.CheatRiskScore{
		RoomID:             roomID,
		PlayerUID:          playerUID,
		ReplayID:           result.ReplayID,
		GameHistoryID:      result.GameHistoryID,
		OperationIndex:     result.OperationIndex,
		OperationTimestamp: result.OperationTimestamp,
		PrimaryEvidence:    MarshalReplayEvidenceAnchor(result.PrimaryEvidence),
		RelatedEvidence:    MarshalReplayEvidenceAnchors(result.RelatedEvidence),
		RiskScore:          result.RiskScore,
		ResponseTimeDim:    result.Dimensions["response_time"],
		FrequencyDim:       result.Dimensions["frequency"],
		WinRateDim:         result.Dimensions["win_rate"],
		PatternDim:         result.Dimensions["pattern"],
		AccountAgeDim:      result.Dimensions["account_age"],
		ReportContribution: riskJSONReport(result),
		IndicatorDetails:   riskJSONIndicators(result),
		SuggestedAction:    result.SuggestedAction,
		SuggestionReason:   result.SuggestionReason,
		ReviewStatus:       "pending",
		PunishmentDecision: result.SanctionType,
		ThresholdSource:    result.ThresholdSource,
		BaselineSnapshot:   MarshalBaselineSnapshot(result.AdaptiveDeviations),
		DecayFactor:        result.DecayFactorApplied,
		EffectiveWeights:   marshalEffectiveWeights(result.EffectiveWeights),
		NewPlayerObserve:   result.NewPlayerObserve,
	}

	if err := s.Repository.SaveRiskScore(riskScore); err != nil {
		log.Printf("[反作弊] 保存风险分数失败: %v", err)
		return result, nil, err
	}

	// 3. 记录审计日志
	if err := s.AuditLogger.LogDetectionEvidence(roomID, playerUID, riskScore.ID, result); err != nil {
		log.Printf("[反作弊] 记录审计日志失败: %v", err)
	}

	// 4. 根据风险分数做出处罚决策
	decision := s.Decider.MakeDecision(result.RiskScore, roomID, playerUID, riskScore.ID)

	// 4b. 新玩家保护：观察期内禁止自动封禁，高风险降级为人工复核（观察）。
	if result.NewPlayerObserve && decision.SanctionType == "ban" {
		log.Printf("[反作弊] 玩家 %d 处于新手观察期，高风险封禁改为人工复核", playerUID)
		decision.SanctionType = "observe"
		decision.Reason = "新手观察期高风险，转人工复核（自动封禁已抑制）"
		decision.Duration = nil
		decision.EffectiveUntil = nil
		result.SanctionType = "observe"
		result.SuggestedAction = "observe"
		result.SuggestionReason = "new-player observation: auto-ban suppressed, routed to manual review"
		if err := s.Repository.UpdateRiskScoreReview(riskScore.ID, "pending", "observe"); err != nil {
			log.Printf("[反作弊] 更新观察期处罚建议失败: %v", err)
		}
	}

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

	// 6. 基线采集：仅对非违规检测累积玩家行为基线（采集器内部按开关与违规过滤）。
	s.collectBaseline(playerUID, context, result)

	return result, decision, nil
}

// enrichDetectionContext populates the optimization inputs on the context from
// persisted data, but only for the features that are currently enabled. When all
// optimization features are disabled this performs no queries and the context is
// unchanged, preserving historical behavior.
func (s *System) enrichDetectionContext(playerUID uint, context *DetectionContext) {
	if s == nil || s.Config == nil || context == nil {
		return
	}

	// Check if batch enrichment is enabled
	enableBatch := os.Getenv("ENABLE_ANTICHEAT_BATCH") == "true" || os.Getenv("ENABLE_ANTICHEAT_BATCH") == "1"

	// If batch is enabled but we only have one player, use the batch path anyway
	// (it will optimize itself for single-player case)
	if enableBatch {
		contexts := map[uint]*DetectionContext{playerUID: context}
		s.enrichDetectionContextsBatch([]uint{playerUID}, contexts)
		return
	}

	// Original single-player logic
	opt := s.Config.GetConfig().Optimization

	// Derive a win-rate snapshot for adaptive/z-score tracks.
	if (opt.AdaptiveThreshold.Enabled || opt.ZScore.Enabled) && context.TotalGames > 0 && !context.HasWinRate {
		context.WinRate = float64(context.WinCount) / float64(context.TotalGames)
		context.HasWinRate = true
	}

	// Personal baselines for the adaptive-threshold personal track.
	if opt.AdaptiveThreshold.Enabled && context.PersonalBaselines == nil && s.Repository != nil {
		if baselines, err := s.Repository.GetPlayerBaselines(playerUID); err == nil {
			context.PersonalBaselines = baselines
		} else {
			log.Printf("[反作弊] 读取玩家基线失败: %v", err)
		}
	}

	// New-player observation period + risk-decay profile.
	if (opt.NewPlayer.Enabled || opt.RiskDecay.Enabled) && s.Repository != nil {
		if profile, err := s.Repository.GetPlayerRiskProfile(playerUID); err == nil && profile != nil {
			if opt.NewPlayer.Enabled {
				context.IsNewPlayer = isObservationPeriod(profile, context, opt.NewPlayer)
			}
			if opt.RiskDecay.Enabled {
				context.HistoricalRisk = profile.HistoricalRisk
				context.NormalGamesSinceViolation = profile.NormalGamesSinceViolation
				context.LastViolationAt = profile.LastViolationAt
			}
		} else if err != nil {
			log.Printf("[反作弊] 读取玩家风险档案失败: %v", err)
		}
	}
}

// enrichDetectionContextsBatch populates optimization inputs for multiple players in batch.
// contexts is a map of playerUID -> DetectionContext that will be mutated in place.
func (s *System) enrichDetectionContextsBatch(playerUIDs []uint, contexts map[uint]*DetectionContext) {
	if s == nil || s.Config == nil || len(playerUIDs) == 0 || s.Repository == nil {
		return
	}

	start := time.Now()
	opt := s.Config.GetConfig().Optimization

	// Derive win-rate snapshots (no DB query needed)
	for _, context := range contexts {
		if (opt.AdaptiveThreshold.Enabled || opt.ZScore.Enabled) && context.TotalGames > 0 && !context.HasWinRate {
			context.WinRate = float64(context.WinCount) / float64(context.TotalGames)
			context.HasWinRate = true
		}
	}

	// Batch fetch personal baselines if needed
	if opt.AdaptiveThreshold.Enabled {
		baselinesMap, err := s.Repository.GetPlayerBaselinesMulti(playerUIDs)
		if err != nil {
			log.Printf("[反作弊] 批量读取玩家基线失败: %v", err)
		} else {
			for uid, baselines := range baselinesMap {
				if context, ok := contexts[uid]; ok && context.PersonalBaselines == nil {
					context.PersonalBaselines = baselines
				}
			}
		}
	}

	// Batch fetch risk profiles if needed
	if opt.NewPlayer.Enabled || opt.RiskDecay.Enabled {
		profilesMap, err := s.Repository.GetPlayerRiskProfilesMulti(playerUIDs)
		if err != nil {
			log.Printf("[反作弊] 批量读取玩家风险档案失败: %v", err)
		} else {
			for uid, profile := range profilesMap {
				context, ok := contexts[uid]
				if !ok || profile == nil {
					continue
				}

				if opt.NewPlayer.Enabled {
					context.IsNewPlayer = isObservationPeriod(profile, context, opt.NewPlayer)
				}
				if opt.RiskDecay.Enabled {
					context.HistoricalRisk = profile.HistoricalRisk
					context.NormalGamesSinceViolation = profile.NormalGamesSinceViolation
					context.LastViolationAt = profile.LastViolationAt
				}
			}
		}
	}

	duration := time.Since(start)
	log.Printf("[反作弊] 批量enrichment完成: players=%d, duration=%v", len(playerUIDs), duration)
}

// isObservationPeriod reports whether the account is still within the new-player
// observation window (below either the game-count or registration-age threshold).
func isObservationPeriod(profile *repository.PlayerRiskProfile, context *DetectionContext, cfg NewPlayerConfig) bool {
	games := profile.TotalGames
	if context.TotalGames > games {
		games = context.TotalGames
	}
	ageDays := profile.AccountAgeDays
	if context.AccountAgeDays > 0 && (ageDays == 0 || context.AccountAgeDays < ageDays) {
		ageDays = context.AccountAgeDays
	}
	belowGames := cfg.MinGames > 0 && games < cfg.MinGames
	belowAge := cfg.MinAgeDays > 0 && ageDays < cfg.MinAgeDays
	return belowGames || belowAge
}

// collectBaseline feeds a non-violation sample to the baseline collector. Samples
// that escalated to a mute/ban sanction are marked as violations and excluded.
func (s *System) collectBaseline(playerUID uint, context *DetectionContext, result *RiskScoringResult) {
	if s == nil || s.Baselines == nil || context == nil || result == nil {
		return
	}
	isViolation := result.SanctionType == "mute" || result.SanctionType == "ban"
	sample := BaselineSample{
		PlayerUID:     playerUID,
		ResponseTimes: context.ResponseTimes,
		IsViolation:   isViolation,
		SampledAt:     result.Timestamp,
	}
	if context.TotalGames > 0 {
		sample.WinRate = float64(context.WinCount) / float64(context.TotalGames)
		sample.HasWinRate = true
	}
	if err := s.Baselines.Collect(sample); err != nil {
		log.Printf("[反作弊] 基线采集失败: %v", err)
	}
}

func riskJSONIndicators(result *RiskScoringResult) database.JSON {
	if result == nil {
		return nil
	}
	return MarshalIndicatorDetails(result.IndicatorDetails)
}

func riskJSONReport(result *RiskScoringResult) database.JSON {
	if result == nil {
		return nil
	}
	return MarshalReportContribution(result.ReportContribution)
}

func marshalEffectiveWeights(weights map[string]float64) database.JSON {
	if len(weights) == 0 {
		return nil
	}
	data, err := json.Marshal(weights)
	if err != nil {
		return nil
	}
	return data
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
		"pending":      0,
		"approved":     0,
		"rejected":     0,
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

// RunRuleTest executes an offline rule-test against the draft config. When the
// provided sample slice is empty it samples recent detection records. It persists
// only an AnticheatRuleTest summary row and never mutates live player, risk,
// sanction, or ban state. operatorUID is recorded for attribution (may be nil).
func (s *System) RunRuleTest(draft *RiskScoringConfig, samples []RuleTestSample, sampleLimit int, operatorUID *uint, note string) (*RuleTestResult, error) {
	if len(samples) == 0 && s.Repository != nil {
		records, err := s.Repository.GetRecentRiskScoresForSampling(sampleLimit)
		if err != nil {
			return nil, err
		}
		samples = make([]RuleTestSample, 0, len(records))
		for _, rec := range records {
			samples = append(samples, SampleFromRiskScore(rec))
		}
	}

	result, err := RunRuleTest(draft, samples)
	if err != nil {
		return nil, err
	}

	// Persist a summary record only (no live-state mutation).
	if s.Repository != nil {
		configSnapshot, _ := json.Marshal(draft)
		record := &database.AnticheatRuleTest{
			ConfigSnapshot: database.JSON(configSnapshot),
			ResultSummary:  MarshalRuleTestResult(result),
			SampleCount:    result.SampleCount,
			CreatedBy:      operatorUID,
			Note:           note,
		}
		if err := s.Repository.SaveRuleTest(record); err != nil {
			log.Printf("[反作弊] 保存规则测试记录失败: %v", err)
		}
	}

	return result, nil
}