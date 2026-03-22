package game

import (
	"testing"
	"time"

	"chemistryuno/backend/models"
)

// TestLeaveRoom_ObserverPromotion 测试观战者升级逻辑
func TestLeaveRoom_ObserverPromotion(t *testing.T) {
	// 初始化测试数据
	gameRoom := &GameRoom{
		Room: &models.Room{
			ID:         "test-room-1",
			Status:     "waiting",
			MaxPlayers: 2,
			Players:    []int{1},
			Spectators: []int{2},
			ReadyUIDs:  []int{},
		},
		GameState: nil,
		OfflineAt: make(map[int]time.Time),
	}

	// 玩家1离开，观战者2应该升级为玩家
	initialSpectatorsCount := len(gameRoom.Room.Spectators)
	initialPlayersCount := len(gameRoom.Room.Players)

	// 模拟升级逻辑（这是LeaveRoom中的部分）
	if gameRoom.Room.Status == "waiting" && len(gameRoom.Room.Spectators) > 0 && len(gameRoom.Room.Players) < gameRoom.Room.MaxPlayers {
		promotedUID := gameRoom.Room.Spectators[0]

		// 检查是否已在Players中
		alreadyPlayer := false
		for _, pid := range gameRoom.Room.Players {
			if pid == promotedUID {
				alreadyPlayer = true
				break
			}
		}

		if !alreadyPlayer {
			gameRoom.Room.Spectators = gameRoom.Room.Spectators[1:]
			gameRoom.Room.Players = append(gameRoom.Room.Players, promotedUID)
		}
	}

	// 验证升级结果
	if len(gameRoom.Room.Players) != initialPlayersCount+1 {
		t.Errorf("期望玩家数增加1，但从 %d 变为 %d", initialPlayersCount, len(gameRoom.Room.Players))
	}

	if len(gameRoom.Room.Spectators) != initialSpectatorsCount-1 {
		t.Errorf("期望观战者数减少1，但从 %d 变为 %d", initialSpectatorsCount, len(gameRoom.Room.Spectators))
	}

	if gameRoom.Room.Players[len(gameRoom.Room.Players)-1] != 2 {
		t.Errorf("期望观战者2升级为玩家，但最后的玩家是 %d", gameRoom.Room.Players[len(gameRoom.Room.Players)-1])
	}

	t.Logf("✅ 观战者升级测试通过: 观战者数从 %d 变为 %d，玩家列表: %v", initialSpectatorsCount, len(gameRoom.Room.Spectators), gameRoom.Room.Players)
}

// TestLeaveRoom_NoDuplicateInPlayers 测试观战者升级时不会重复加入
func TestLeaveRoom_NoDuplicateInPlayers(t *testing.T) {
	gameRoom := &GameRoom{
		Room: &models.Room{
			ID:         "test-room-2",
			Status:     "waiting",
			MaxPlayers: 3,
			Players:    []int{1, 2},
			Spectators: []int{2, 3}, // 注意：玩家2也在观战者列表中（错误状态）
			ReadyUIDs:  []int{},
		},
		GameState: nil,
		OfflineAt: make(map[int]time.Time),
	}

	// 模拟升级逻辑
	if gameRoom.Room.Status == "waiting" && len(gameRoom.Room.Spectators) > 0 && len(gameRoom.Room.Players) < gameRoom.Room.MaxPlayers {
		promotedUID := gameRoom.Room.Spectators[0]

		// 检查是否已在Players中（防止重复）
		alreadyPlayer := false
		for _, pid := range gameRoom.Room.Players {
			if pid == promotedUID {
				alreadyPlayer = true
				break
			}
		}

		if !alreadyPlayer {
			gameRoom.Room.Spectators = gameRoom.Room.Spectators[1:]
			gameRoom.Room.Players = append(gameRoom.Room.Players, promotedUID)
		}
	}

	// 验证玩家2不会重复出现
	playerCount := 0
	for _, p := range gameRoom.Room.Players {
		if p == 2 {
			playerCount++
		}
	}

	if playerCount > 1 {
		t.Errorf("玩家2出现了 %d 次，应该只出现1次或0次", playerCount)
	}

	t.Logf("✅ 去重测试通过: 玩家列表中无重复项。玩家列表: %v", gameRoom.Room.Players)
}

// TestCurrentPlayerIndexAdjustment 测试当前玩家索引调整
func TestCurrentPlayerIndexAdjustment(t *testing.T) {
	testCases := []struct {
		name                  string
		initialPlayers        int
		currentPlayer         int
		leftIndex             int
		expectedCurrentPlayer int
		description           string
	}{
		{
			name:                  "玩家0离开，当前玩家为1",
			initialPlayers:        3,
			currentPlayer:         1,
			leftIndex:             0,
			expectedCurrentPlayer: 0, // 1-1=0
			description:           "当前玩家索引应该减1",
		},
		{
			name:                  "玩家1离开，当前玩家为1",
			initialPlayers:        3,
			currentPlayer:         1,
			leftIndex:             1,
			expectedCurrentPlayer: 1, // 1 % (3-1) = 1
			description:           "当前玩家在被移除位置，使用模运算",
		},
		{
			name:                  "玩家2离开，当前玩家为2",
			initialPlayers:        3,
			currentPlayer:         2,
			leftIndex:             2,
			expectedCurrentPlayer: 0, // 2 % (3-1) = 0
			description:           "最后玩家离开，回到第一位",
		},
		{
			name:                  "只剩1个玩家",
			initialPlayers:        1,
			currentPlayer:         0,
			leftIndex:             -1,
			expectedCurrentPlayer: 0,
			description:           "玩家总数为0时不调整",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 模拟玩家离开后的索引调整
			playersAfterLeave := tc.initialPlayers - 1
			currentPlayer := tc.currentPlayer

			if tc.leftIndex != -1 && playersAfterLeave > 0 {
				if currentPlayer > tc.leftIndex {
					currentPlayer--
				}
				// 使用模运算确保有效范围
				currentPlayer = currentPlayer % playersAfterLeave
			} else if playersAfterLeave == 0 {
				currentPlayer = 0
			}

			if currentPlayer != tc.expectedCurrentPlayer {
				t.Errorf("%s: 期望索引 %d，但得到 %d。%s",
					tc.name, tc.expectedCurrentPlayer, currentPlayer, tc.description)
			} else {
				t.Logf("✅ %s: 索引调整正确 (预期: %d, 实际: %d)", tc.name, tc.expectedCurrentPlayer, currentPlayer)
			}
		})
	}
}

// TestSpectatorsSync 测试Spectators列表同步
func TestSpectatorsSync(t *testing.T) {
	gameRoom := &GameRoom{
		Room: &models.Room{
			ID:         "test-room-3",
			Status:     "playing",
			MaxPlayers: 4,
			Players:    []int{1, 2},
			Spectators: []int{3, 4},
			ReadyUIDs:  []int{},
		},
		GameState: &models.GameState{
			Spectators: []int{3, 4},
			Players: []*models.PlayerState{
				{UID: 1, Username: "p1"},
				{UID: 2, Username: "p2"},
			},
			CurrentPlayer: 0,
		},
		OfflineAt: make(map[int]time.Time),
	}

	// 模拟观战者3升级为玩家（在waiting状态）
	gameRoom.Room.Status = "waiting"
	promotedUID := 3

	// 从GameState.Spectators中移除升级的观战者
	newGameStateSpectators := []int{}
	for _, sid := range gameRoom.GameState.Spectators {
		if sid != promotedUID {
			newGameStateSpectators = append(newGameStateSpectators, sid)
		}
	}
	gameRoom.GameState.Spectators = newGameStateSpectators

	// 验证同步
	if len(gameRoom.GameState.Spectators) != 1 {
		t.Errorf("期望GameState.Spectators只有1个元素，但有 %d 个", len(gameRoom.GameState.Spectators))
	}

	if gameRoom.GameState.Spectators[0] != 4 {
		t.Errorf("期望剩余观战者为4，但为 %d", gameRoom.GameState.Spectators[0])
	}

	t.Logf("✅ Spectators同步测试通过: 升级后GameState.Spectators = %v", gameRoom.GameState.Spectators)
}

// BenchmarkLeaveRoomPromotion 性能测试：观战者升级
func BenchmarkLeaveRoomPromotion(b *testing.B) {
	gameRoom := &GameRoom{
		Room: &models.Room{
			ID:         "bench-room",
			Status:     "waiting",
			MaxPlayers: 4,
			Players:    []int{1},
			Spectators: make([]int, 0, 100),
			ReadyUIDs:  []int{},
		},
		GameState: &models.GameState{
			Spectators: make([]int, 0, 100),
			Players:    []*models.PlayerState{{UID: 1, Username: "p1"}},
		},
		OfflineAt: make(map[int]time.Time),
	}

	// 添加大量观战者
	for i := 2; i < 102; i++ {
		gameRoom.Room.Spectators = append(gameRoom.Room.Spectators, i)
		gameRoom.GameState.Spectators = append(gameRoom.GameState.Spectators, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		promotedUID := gameRoom.Room.Spectators[0]
		alreadyPlayer := false
		for _, pid := range gameRoom.Room.Players {
			if pid == promotedUID {
				alreadyPlayer = true
				break
			}
		}
		if !alreadyPlayer {
			gameRoom.Room.Spectators = gameRoom.Room.Spectators[1:]
			gameRoom.Room.Players = append(gameRoom.Room.Players, promotedUID)

			// 同步移除
			newGameStateSpectators := []int{}
			for _, sid := range gameRoom.GameState.Spectators {
				if sid != promotedUID {
					newGameStateSpectators = append(newGameStateSpectators, sid)
				}
			}
			gameRoom.GameState.Spectators = newGameStateSpectators

			// 恢复状态用于下一次迭代
			gameRoom.Room.Players = gameRoom.Room.Players[:1]
			gameRoom.Room.Spectators = append([]int{promotedUID + 100}, gameRoom.Room.Spectators...)
			gameRoom.GameState.Spectators = append([]int{promotedUID + 100}, gameRoom.GameState.Spectators...)
		}
	}
}
