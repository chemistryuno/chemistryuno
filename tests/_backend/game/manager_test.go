package game

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"chemistryuno/backend/anticheat"
	"chemistryuno/backend/database"
	"chemistryuno/backend/models"
	"chemistryuno/backend/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRoomExitTest(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&database.User{}, &database.GameHistory{}, &database.PrivateChat{}, &database.Bounty{}, &database.SystemConfig{}); err != nil {
		t.Fatalf("migrate room exit tables: %v", err)
	}

	previousDB := database.DB
	previousUserRepo := repository.UserRepo
	previousGameRepo := repository.GameRepo
	previousBountyRepo := repository.BountyRepo
	previousConfigRepo := configRepo
	database.DB = db
	repository.UserRepo = repository.NewUserRepository(db)
	repository.GameRepo = repository.NewGameRepository()
	repository.BountyRepo = repository.NewBountyRepository()
	configRepo = repository.NewConfigRepository()
	if err := configRepo.InitDefaultConfigs(); err != nil {
		t.Fatalf("init configs: %v", err)
	}

	roomMutex.Lock()
	previousRooms := rooms
	rooms = make(map[string]*GameRoom)
	roomMutex.Unlock()

	t.Cleanup(func() {
		roomMutex.Lock()
		rooms = previousRooms
		roomMutex.Unlock()
		database.DB = previousDB
		repository.UserRepo = previousUserRepo
		repository.GameRepo = previousGameRepo
		repository.BountyRepo = previousBountyRepo
		configRepo = previousConfigRepo
	})
	return db
}

func seedRoomExitUsers(t *testing.T, db *gorm.DB, uids ...uint) {
	t.Helper()
	for _, uid := range uids {
		user := database.User{
			UID:           uid,
			Username:      fmt.Sprintf("user_%d", uid),
			Nickname:      fmt.Sprintf("user_%d", uid),
			Points:        1000,
			MonthlyPoints: 1000,
		}
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("seed user %d: %v", uid, err)
		}
	}
}

func setupSubstanceValidationTest(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&database.User{}, &database.Substance{}, &database.Reaction{}); err != nil {
		t.Fatalf("migrate substance validation tables: %v", err)
	}

	previousDB := database.DB
	previousUserRepo := repository.UserRepo
	previousSubstanceRepo := repository.SubstanceRepo
	previousReactionRepo := repository.ReactionRepo
	database.DB = db
	repository.UserRepo = repository.NewUserRepository(db)
	repository.SubstanceRepo = repository.NewSubstanceRepository()
	repository.ReactionRepo = repository.NewReactionRepository()

	validSubstancesMutex.Lock()
	previousValidSubstances := validSubstances
	validSubstances = nil
	validSubstancesMutex.Unlock()

	roomMutex.Lock()
	previousRooms := rooms
	rooms = make(map[string]*GameRoom)
	roomMutex.Unlock()

	t.Cleanup(func() {
		roomMutex.Lock()
		rooms = previousRooms
		roomMutex.Unlock()

		validSubstancesMutex.Lock()
		validSubstances = previousValidSubstances
		validSubstancesMutex.Unlock()

		database.DB = previousDB
		repository.UserRepo = previousUserRepo
		repository.SubstanceRepo = previousSubstanceRepo
		repository.ReactionRepo = previousReactionRepo
	})
	return db
}

func seedApprovedSubstance(t *testing.T, db *gorm.DB, formula string) {
	t.Helper()
	substance := database.Substance{
		Name:         formula,
		Formula:      formula,
		Status:       "approved",
		CreatedByUID: 100000000,
	}
	if err := db.Create(&substance).Error; err != nil {
		t.Fatalf("seed approved substance %s: %v", formula, err)
	}
}

func putTestRoom(gr *GameRoom) {
	roomMutex.Lock()
	rooms[gr.Room.ID] = gr
	roomMutex.Unlock()
}

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

func TestLeaveRoom_WaitingRoomClosesWhenOnlyAIRemains(t *testing.T) {
	db := setupRoomExitTest(t)
	seedRoomExitUsers(t, db, 100000000)

	roomID := "waiting-ai-only-room"
	putTestRoom(&GameRoom{
		Room: &models.Room{
			ID:         roomID,
			Status:     "waiting",
			MaxPlayers: 2,
			Players:    []int{100000000, -1},
			ReadyUIDs:  []int{100000000, -1},
			Spectators: []int{},
			CreatedAt:  time.Now(),
		},
		OfflineAt: make(map[int]time.Time),
	})

	if err := LeaveRoom(roomID, 100000000); err != nil {
		t.Fatalf("leave room: %v", err)
	}

	if exists, status := GetRoomStatus(roomID); exists {
		t.Fatalf("expected AI-only waiting room to be removed, got status %q", status)
	}
}

func TestLeaveRoom_WaitingSpectatorPromotionKeepsRoomAlive(t *testing.T) {
	db := setupRoomExitTest(t)
	seedRoomExitUsers(t, db, 1, 2)

	roomID := "waiting-promotion-room"
	putTestRoom(&GameRoom{
		Room: &models.Room{
			ID:         roomID,
			Status:     "waiting",
			MaxPlayers: 4,
			Players:    []int{1, -1},
			ReadyUIDs:  []int{1, -1},
			Spectators: []int{2},
			CreatedAt:  time.Now(),
		},
		GameState: &models.GameState{
			Spectators: []int{2},
		},
		OfflineAt: make(map[int]time.Time),
	})

	if err := LeaveRoom(roomID, 1); err != nil {
		t.Fatalf("leave room: %v", err)
	}

	roomMutex.RLock()
	gr := rooms[roomID]
	roomMutex.RUnlock()
	if gr == nil {
		t.Fatal("expected room to remain after human spectator promotion")
	}

	gr.mutex.RLock()
	defer gr.mutex.RUnlock()
	if gr.Room.Status != "waiting" {
		t.Fatalf("expected waiting room, got %q", gr.Room.Status)
	}
	if !containsInt(gr.Room.Players, 2) {
		t.Fatalf("expected spectator 2 to be promoted into players, got %v", gr.Room.Players)
	}
	if !gr.hasHumanPlayersLocked() {
		t.Fatalf("expected promoted human to keep room alive, got players %v", gr.Room.Players)
	}
}

func TestStartedRoomTemporaryDisconnectDoesNotMarkExit(t *testing.T) {
	roomID := "temporary-disconnect-room"
	gr := &GameRoom{
		Room: &models.Room{
			ID:         roomID,
			Status:     "playing",
			MaxPlayers: 2,
			Players:    []int{1, 2},
		},
		GameState: &models.GameState{
			RoomID:              roomID,
			Status:              "playing",
			OriginalPlayerCount: 2,
			Players: []*models.PlayerState{
				{UID: 1, Username: "p1"},
				{UID: 2, Username: "p2"},
			},
			ExitedPlayers: []int{},
			PointsChanges: map[int]int{},
			XPChanges:     map[int]int{},
		},
		OfflineAt: map[int]time.Time{
			1: time.Now(),
		},
	}

	gr.mutex.Lock()
	finalize, allZero := gr.shouldEndAfterHumanExitLocked()
	gr.mutex.Unlock()

	if finalize || allZero {
		t.Fatalf("temporary disconnect should not finalize game, finalize=%v allZero=%v", finalize, allZero)
	}
	if len(gr.GameState.ExitedPlayers) != 0 {
		t.Fatalf("temporary disconnect should not mark exits, got %v", gr.GameState.ExitedPlayers)
	}
}

func TestStartedRoomAllHumansLeaveEndsWithZeroIncome(t *testing.T) {
	db := setupRoomExitTest(t)
	seedRoomExitUsers(t, db, 1)

	roomID := "started-all-humans-left-room"
	putTestRoom(&GameRoom{
		Room: &models.Room{
			ID:           roomID,
			Status:       "playing",
			MaxPlayers:   2,
			Players:      []int{1, -1},
			IsPointsMode: true,
		},
		GameState: &models.GameState{
			RoomID:              roomID,
			Status:              "playing",
			OriginalPlayerCount: 1,
			Players: []*models.PlayerState{
				{UID: 1, Username: "p1"},
				{UID: -1, Username: "ai", IsAI: true},
			},
			ExitedPlayers: []int{},
			PointsChanges: map[int]int{},
			XPChanges:     map[int]int{},
			IsPointsMode:  true,
		},
		OfflineAt:    make(map[int]time.Time),
		ReplayEvents: []map[string]interface{}{},
	})

	if err := LeaveRoom(roomID, 1); err != nil {
		t.Fatalf("leave room: %v", err)
	}

	if exists, status := GetRoomStatus(roomID); exists {
		t.Fatalf("expected all-human-exited started room to be removed, got status %q", status)
	}

	var user database.User
	if err := db.Where("uid = ?", 1).First(&user).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if user.Points != 1000 || user.MonthlyPoints != 1000 {
		t.Fatalf("expected zero-income termination to preserve points, got points=%v monthly=%v", user.Points, user.MonthlyPoints)
	}
}

func TestPointsCalculationExitedPlayerZeroAndReducedRanking(t *testing.T) {
	db := setupRoomExitTest(t)
	seedRoomExitUsers(t, db, 1, 2, 3, 4)

	gr := &GameRoom{
		Room: &models.Room{
			ID:           "reduced-ranking-room",
			Status:       "playing",
			MaxPlayers:   4,
			Players:      []int{1, 2, 3},
			IsPointsMode: true,
		},
		GameState: &models.GameState{
			RoomID:              "reduced-ranking-room",
			Status:              "playing",
			OriginalPlayerCount: 4,
			FinishedPlayers:     []int{1, 2, 3},
			ExitedPlayers:       []int{4},
			Players: []*models.PlayerState{
				{UID: 1, Username: "p1"},
				{UID: 2, Username: "p2"},
				{UID: 3, Username: "p3"},
			},
			PointsChanges: map[int]int{},
			XPChanges:     map[int]int{1: 0, 2: 0, 3: 0, 4: 0},
			IsPointsMode:  true,
		},
	}

	gr.mutex.Lock()
	handlePointsCalculation(gr)
	gr.mutex.Unlock()

	expected := map[int]int{1: 100, 2: 50, 3: 5, 4: 0}
	for uid, want := range expected {
		if got := gr.GameState.PointsChanges[uid]; got != want {
			t.Fatalf("uid %d expected points change %d, got %d (all changes: %v)", uid, want, got, gr.GameState.PointsChanges)
		}
	}

	var exited database.User
	if err := db.Where("uid = ?", 4).First(&exited).Error; err != nil {
		t.Fatalf("load exited user: %v", err)
	}
	if exited.Points != 1000 || exited.MonthlyPoints != 1000 {
		t.Fatalf("expected exited player to receive no DB points, got points=%v monthly=%v", exited.Points, exited.MonthlyPoints)
	}
}

func TestCollectAnticheatDataIncludesReplayBoundReports(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&database.User{}, &database.Feedback{}); err != nil {
		t.Fatalf("migrate feedback tables: %v", err)
	}
	previousDB := database.DB
	previousFeedbackRepo := repository.FeedbackRepo
	database.DB = db
	repository.FeedbackRepo = repository.NewFeedbackRepository()
	t.Cleanup(func() {
		database.DB = previousDB
		repository.FeedbackRepo = previousFeedbackRepo
	})

	anchor := database.ReplayEvidenceAnchor{
		RoomID:           "report-room",
		GameHistoryID:    77,
		ReplayID:         "77",
		EventIndex:       3,
		EventID:          "evt-3",
		EventType:        "play_card",
		PlayerUID:        1001,
		EventTimestampMs: 1710000000000,
		ActionSummary:    "played H2O",
	}
	feedbacks := []database.Feedback{
		{
			UserUID:         2001,
			Type:            "report",
			Content:         "suspicious replay point",
			RoomID:          "report-room",
			ReportedUID:     1001,
			ReplayID:        "77",
			GameHistoryID:   77,
			PrimaryEvidence: anticheat.MarshalReplayEvidenceAnchor(anchor),
			Status:          "pending",
		},
		{
			UserUID:         2001,
			Type:            "report",
			Content:         "duplicate same reporter",
			RoomID:          "report-room",
			ReportedUID:     1001,
			PrimaryEvidence: anticheat.MarshalReplayEvidenceAnchor(anchor),
			Status:          "pending",
		},
		{
			UserUID:         2002,
			Type:            "report",
			Content:         "second reporter",
			RoomID:          "report-room",
			ReportedUID:     1001,
			PrimaryEvidence: anticheat.MarshalReplayEvidenceAnchor(anchor),
			Status:          "pending",
		},
		{
			UserUID:     2003,
			Type:        "report",
			Content:     "unbound report should not score",
			RoomID:      "report-room",
			ReportedUID: 1001,
			Status:      "pending",
		},
	}
	for i := range feedbacks {
		if err := db.Create(&feedbacks[i]).Error; err != nil {
			t.Fatalf("create feedback: %v", err)
		}
	}

	gameRoom := &GameRoom{
		Room: &models.Room{
			ID:      "report-room",
			Players: []int{1001, 1002},
		},
		GameState: &models.GameState{
			Players: []*models.PlayerState{
				{UID: 1001, Username: "reported"},
				{UID: 1002, Username: "other"},
			},
		},
		ReplayEvents: []map[string]interface{}{
			{
				"event":          "play_card",
				"event_index":    3,
				"event_id":       "evt-3",
				"uid":            1001,
				"unix_ms":        float64(1710000000000),
				"action_summary": "played H2O",
			},
		},
	}

	contexts := gameRoom.collectAnticheatDataLocked()
	reported := contexts[1001]
	if reported == nil {
		t.Fatal("expected context for reported player")
	}
	if reported.ReportCount != 2 || len(reported.ReportEvidence) != 2 {
		t.Fatalf("expected two deduplicated replay-bound reports, got count=%d evidence=%d", reported.ReportCount, len(reported.ReportEvidence))
	}
	if reported.ReportEvidence[0].EventID != "evt-3" || reported.ReportSummary == "" {
		t.Fatalf("expected replay report evidence details, got summary=%q anchors=%+v", reported.ReportSummary, reported.ReportEvidence)
	}
	if other := contexts[1002]; other == nil || other.ReportCount != 0 || len(other.ReportEvidence) != 0 {
		t.Fatalf("expected unrelated player to have no report contribution, got %+v", other)
	}
}

func TestReconnectGracePeriodUsesDedicatedConfigKey(t *testing.T) {
	setupRoomExitTest(t)

	configRepo = repository.NewConfigRepository()
	if err := configRepo.SetValue("player_kick_timeout", "120"); err != nil {
		t.Fatalf("set player_kick_timeout: %v", err)
	}
	if err := configRepo.SetValue("reconnect_grace_period", "35"); err != nil {
		t.Fatalf("set reconnect_grace_period: %v", err)
	}

	gotReconnect := getReconnectGracePeriod()
	if gotReconnect != 35*time.Second {
		t.Fatalf("expected reconnect grace period 35s, got %v", gotReconnect)
	}

	gotKick := getPlayerKickTimeout()
	if gotKick != 120*time.Second {
		t.Fatalf("expected player kick timeout 120s, got %v", gotKick)
	}
}

func TestPlayCardRequiresApprovedSubstanceForNormalCards(t *testing.T) {
	db := setupSubstanceValidationTest(t)
	seedApprovedSubstance(t, db, "O")
	RebuildSubstanceCache()

	roomID := "normal-substance-validation-room"
	putTestRoom(&GameRoom{
		Room: &models.Room{
			ID:      roomID,
			Status:  "playing",
			Players: []int{1, 2},
		},
		GameState: &models.GameState{
			RoomID:        roomID,
			Status:        "playing",
			CurrentPlayer: 0,
			Direction:     1,
			Players: []*models.PlayerState{
				{UID: 1, Username: "p1", Nickname: "p1", HandCards: []models.Card{{Type: "H", Count: 1}, {Type: "O", Count: 1}}, CardCount: 2},
				{UID: 2, Username: "p2", Nickname: "p2", HandCards: []models.Card{{Type: "O", Count: 1}}, CardCount: 1},
			},
			AllowedAnyPlayer: -1,
			PointsChanges:    map[int]int{},
			XPChanges:        map[int]int{},
		},
		OfflineAt:    make(map[int]time.Time),
		ReplayEvents: []map[string]interface{}{},
	})

	err := PlayCard(roomID, 1, models.Card{Type: "H", Count: 1}, "H", 0)
	if err == nil || !strings.Contains(err.Error(), "该物质非法") {
		t.Fatalf("expected unapproved normal substance to be rejected, got %v", err)
	}

	gr := rooms[roomID]
	if got := gr.GameState.Players[0].CardCount; got != 2 {
		t.Fatalf("rejected play should not consume card, got count %d", got)
	}
}

func TestPlayCardAllowsApprovedNormalSubstance(t *testing.T) {
	db := setupSubstanceValidationTest(t)
	seedApprovedSubstance(t, db, "H")
	RebuildSubstanceCache()

	roomID := "approved-normal-substance-room"
	putTestRoom(&GameRoom{
		Room: &models.Room{
			ID:      roomID,
			Status:  "playing",
			Players: []int{1, 2},
		},
		GameState: &models.GameState{
			RoomID:        roomID,
			Status:        "playing",
			CurrentPlayer: 0,
			Direction:     1,
			Players: []*models.PlayerState{
				{UID: 1, Username: "p1", Nickname: "p1", HandCards: []models.Card{{Type: "H", Count: 1}, {Type: "O", Count: 1}}, CardCount: 2},
				{UID: 2, Username: "p2", Nickname: "p2", HandCards: []models.Card{{Type: "O", Count: 1}}, CardCount: 1},
			},
			AllowedAnyPlayer: -1,
			PointsChanges:    map[int]int{},
			XPChanges:        map[int]int{},
		},
		OfflineAt:    make(map[int]time.Time),
		ReplayEvents: []map[string]interface{}{},
	})

	if err := PlayCard(roomID, 1, models.Card{Type: "H", Count: 1}, "H", 0); err != nil {
		t.Fatalf("approved normal substance should be playable: %v", err)
	}

	gr := rooms[roomID]
	if gr.GameState.LastCard == nil || gr.GameState.LastCard.Substance != "H" {
		t.Fatalf("expected H to become last card, got %#v", gr.GameState.LastCard)
	}
}

func TestPlayCardAllowsFunctionalCardsWithoutSubstanceEntry(t *testing.T) {
	setupSubstanceValidationTest(t)
	RebuildSubstanceCache()

	roomID := "functional-card-validation-room"
	putTestRoom(&GameRoom{
		Room: &models.Room{
			ID:      roomID,
			Status:  "playing",
			Players: []int{1, 2},
		},
		GameState: &models.GameState{
			RoomID:        roomID,
			Status:        "playing",
			CurrentPlayer: 0,
			Direction:     1,
			Players: []*models.PlayerState{
				{UID: 1, Username: "p1", Nickname: "p1", HandCards: []models.Card{{Type: "+2", Count: 1, Effect: "+2"}, {Type: "O", Count: 1}}, CardCount: 2},
				{UID: 2, Username: "p2", Nickname: "p2", HandCards: []models.Card{{Type: "O", Count: 1}}, CardCount: 1},
			},
			AllowedAnyPlayer: -1,
			PointsChanges:    map[int]int{},
			XPChanges:        map[int]int{},
		},
		OfflineAt:    make(map[int]time.Time),
		ReplayEvents: []map[string]interface{}{},
	})

	if err := PlayCard(roomID, 1, models.Card{Type: "+2", Count: 1, Effect: "+2"}, "+2", 0); err != nil {
		t.Fatalf("functional card should bypass substance table validation: %v", err)
	}

	gr := rooms[roomID]
	if got := gr.GameState.PendingDrawCount; got != 2 {
		t.Fatalf("expected +2 effect to remain active, got pending draw count %d", got)
	}
}

func TestAIPlayCardUsesSameSubstanceValidation(t *testing.T) {
	setupSubstanceValidationTest(t)
	RebuildSubstanceCache()

	roomID := "ai-substance-validation-room"
	gr := &GameRoom{
		Room: &models.Room{
			ID:      roomID,
			Status:  "playing",
			Players: []int{-1, 1},
		},
		GameState: &models.GameState{
			RoomID:        roomID,
			Status:        "playing",
			CurrentPlayer: 0,
			Direction:     1,
			Players: []*models.PlayerState{
				{UID: -1, Username: "ai", Nickname: "ai", IsAI: true, HandCards: []models.Card{{Type: "H", Count: 1}}, CardCount: 1},
				{UID: 1, Username: "p1", Nickname: "p1", HandCards: []models.Card{{Type: "O", Count: 1}}, CardCount: 1},
			},
			DrawPile:         []models.Card{{Type: "O", Count: 1}, {Type: "O", Count: 1}},
			AllowedAnyPlayer: -1,
			PointsChanges:    map[int]int{},
			XPChanges:        map[int]int{},
		},
		OfflineAt:    make(map[int]time.Time),
		ReplayEvents: []map[string]interface{}{},
	}
	putTestRoom(gr)

	gr.mutex.Lock()
	gr.aiExecutePlay(-1, models.Card{Type: "H", Count: 1}, "H")

	if got := gr.GameState.Players[0].CardCount; got != 3 {
		t.Fatalf("expected invalid AI play to fall back to drawing two cards, got count %d", got)
	}
	if gr.GameState.LastCard != nil {
		t.Fatalf("invalid AI play should not update last card, got %#v", gr.GameState.LastCard)
	}
	if got := gr.GameState.CurrentPlayer; got != 1 {
		t.Fatalf("AI draw fallback should pass turn to player index 1, got %d", got)
	}
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
