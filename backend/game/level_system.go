package game

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/models"
	"chemistryuno/backend/repository"
	"chemistryuno/backend/websocket"
	"fmt"
	"log"
)

// CalculateXPReward 计算游戏经验奖励
// 根据排名、局内表现、对手等级等因素综合计算经验值
func CalculateXPReward(gr *GameRoom, uid int, rank int) int {
	if uid < 0 {
		return 0 // AI 不获得经验
	}

	// 基础经验
	baseXP := 20

	// 排名奖励
	rankBonus := 0
	switch rank {
	case 1:
		rankBonus = 50
	case 2:
		rankBonus = 30
	case 3:
		rankBonus = 20
	case 4:
		rankBonus = 10
	default:
		if rank <= len(gr.GameState.Players) {
			rankBonus = 5
		}
	}

	// 特殊成就奖励
	achievementBonus := 0
	player := gr.GameState.Players[gr.GameState.CurrentPlayer]
	for _, p := range gr.GameState.Players {
		if p.UID == uid {
			player = p
			break
		}
	}

	// 双联反应奖励（估算：假设记录在 ActionProgress 中）
	if player.DoubleActionAvailable {
		achievementBonus += 5
	}

	// AI对战难度修正
	difficultyMultiplier := 1.0
	if gr.Room.IsPvE {
		if gr.Room.PvEDifficulty < 50 {
			difficultyMultiplier = 0.5 // 低难度减半
		} else {
			// 难度 50-100: 0.5x - 1.0x
			difficultyMultiplier = float64(gr.Room.PvEDifficulty) / 100.0
		}
	} else {
		// 对战真人玩家，检查等级差距
		myLevel := 1
		if user, err := repository.UserRepo.FindByUID(uint(uid)); err == nil {
			myLevel = user.Level
		}

		// 计算对手平均等级
		opponentAvgLevel := 0
		opponentCount := 0
		for _, p := range gr.GameState.Players {
			if p.UID > 0 && p.UID != uid {
				if opponent, err := repository.UserRepo.FindByUID(uint(p.UID)); err == nil {
					opponentAvgLevel += opponent.Level
					opponentCount++
				}
			}
		}

		if opponentCount > 0 {
			opponentAvgLevel /= opponentCount
			levelDiff := opponentAvgLevel - myLevel

			// 对战高等级玩家额外奖励，对战低等级玩家大幅减少奖励
			if levelDiff > 0 {
				// 挑战高手奖励：每级 +5% XP，最多 +100% (翻倍)
				// 例如：挑战高10级玩家 = +50% XP，挑战高20级玩家 = +100% XP
				difficultyMultiplier += float64(levelDiff) * 0.05
				if difficultyMultiplier > 2.0 {
					difficultyMultiplier = 2.0 // 最多翻倍
				}
				log.Printf("[XP] 挑战高手加成: 对手平均等级 %d vs 我的等级 %d，倍率 %.2fx", opponentAvgLevel, myLevel, difficultyMultiplier)
			} else if levelDiff < 0 {
				// 虐菜惩罚：等级差距越大，惩罚越重
				absDiff := -levelDiff
				if absDiff >= 20 {
					// 低20级以上：-80% XP
					difficultyMultiplier = 0.2
					log.Printf("[XP] 虐菜严重惩罚: 对手平均等级 %d vs 我的等级 %d，倍率 %.2fx", opponentAvgLevel, myLevel, difficultyMultiplier)
				} else if absDiff >= 10 {
					// 低10-19级：-60% XP
					difficultyMultiplier = 0.4
					log.Printf("[XP] 虐菜中度惩罚: 对手平均等级 %d vs 我的等级 %d，倍率 %.2fx", opponentAvgLevel, myLevel, difficultyMultiplier)
				} else if absDiff >= 5 {
					// 低5-9级：-40% XP
					difficultyMultiplier = 0.6
					log.Printf("[XP] 虐菜轻度惩罚: 对手平均等级 %d vs 我的等级 %d，倍率 %.2fx", opponentAvgLevel, myLevel, difficultyMultiplier)
				}
				// 低于5级差距不惩罚（允许一定的等级波动）
			}
		}
	}

	// 计算总经验
	totalXP := int(float64(baseXP+rankBonus+achievementBonus) * difficultyMultiplier)

	// 确保至少获得 1 XP（参与奖）
	if totalXP < 1 {
		totalXP = 1
	}

	log.Printf("[XP] 玩家 %d 获得 %d XP (基础:%d 排名:%d 成就:%d 倍率:%.2f)",
		uid, totalXP, baseXP, rankBonus, achievementBonus, difficultyMultiplier)

	return totalXP
}

// AwardXP 授予玩家经验并检查升级
func AwardXP(uid int, xp int) error {
	if uid < 0 || xp <= 0 {
		return nil
	}

	user, err := repository.UserRepo.FindByUID(uint(uid))
	if err != nil {
		return err
	}

	// 更新经验
	newXP := user.XP + xp
	newTotalXP := user.TotalXP + xp

	// 检查是否升级
	leveledUp := false
	newLevels := []int{}
	currentLevel := user.Level
	currentXP := newXP

	for {
		if currentLevel >= 100 {
			break // 已满级
		}

		// 查询当前等级所需经验
		var levelConfig database.LevelConfig
		if err := database.DB.Where("level = ?", currentLevel+1).First(&levelConfig).Error; err != nil {
			break
		}

		// 检查是否可以升级
		if currentXP >= levelConfig.RequiredXP {
			currentXP -= levelConfig.RequiredXP
			currentLevel++
			leveledUp = true
			newLevels = append(newLevels, currentLevel)
		} else {
			break
		}
	}

	// 保存到数据库
	if err := repository.UserRepo.UpdateXP(uint(uid), currentXP, newTotalXP, currentLevel); err != nil {
		return err
	}

	// 如果升级了，发送通知
	if leveledUp && websocket.GlobalHub != nil {
		for _, level := range newLevels {
			// 获取等级配置
			var levelConfig database.LevelConfig
			if err := database.DB.Where("level = ?", level).First(&levelConfig).Error; err == nil {
				websocket.GlobalHub.SendToUID(uid, websocket.Message{
					Type: "level_up",
					Data: map[string]interface{}{
						"level":     level,
						"tier":      levelConfig.Tier,
						"tier_name": levelConfig.TierName,
						"xp":        currentXP,
						"total_xp":  newTotalXP,
					},
				})

				// 发送浮窗提示
				websocket.GlobalHub.SendToUID(uid, websocket.Message{
					Type: "action_toast",
					Data: fmt.Sprintf("🎉 恭喜升级！你现在是 %s %d 级研究员！", levelConfig.TierName, level),
				})
			}
		}
	}

	return nil
}

// GetLevelInfo 获取玩家等级信息
func GetLevelInfo(uid uint) (*models.LevelInfo, error) {
	user, err := repository.UserRepo.FindByUID(uid)
	if err != nil {
		return nil, err
	}

	// 查询当前等级配置
	var currentConfig database.LevelConfig
	if err := database.DB.Where("level = ?", user.Level).First(&currentConfig).Error; err != nil {
		return nil, err
	}

	// 查询下一等级配置
	var nextConfig database.LevelConfig
	nextLevelExists := true
	if user.Level < 100 {
		if err := database.DB.Where("level = ?", user.Level+1).First(&nextConfig).Error; err != nil {
			nextLevelExists = false
		}
	} else {
		nextLevelExists = false
	}

	info := &models.LevelInfo{
		Level:          user.Level,
		XP:             user.XP,
		TotalXP:        user.TotalXP,
		Tier:           currentConfig.Tier,
		TierName:       currentConfig.TierName,
		NextLevelXP:    0,
		ProgressPercent: 100,
	}

	if nextLevelExists {
		info.NextLevelXP = nextConfig.RequiredXP
		info.ProgressPercent = int(float64(user.XP) / float64(nextConfig.RequiredXP) * 100)
	}

	return info, nil
}
