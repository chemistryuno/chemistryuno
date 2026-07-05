package main

import (
	"chemistryuno/backend/anticheat"
	"chemistryuno/backend/cache"
	"chemistryuno/backend/database"
	"chemistryuno/backend/game"
	"chemistryuno/backend/handlers"
	"chemistryuno/backend/middleware"
	"chemistryuno/backend/plugins"
	"chemistryuno/backend/repository"
	"chemistryuno/backend/router"
	"chemistryuno/backend/static"
	"chemistryuno/backend/utils"
	"chemistryuno/backend/websocket"
	"context"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var upgrader = ws.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var hub *websocket.Hub

func main() {
	// 加载 .env 配置文件（统一从根目录读取）
	// 优先级：1. 当前目录 .env（从根目录运行）2. 上级目录 .env（从 backend 目录运行）
	var loadErr error
	envFileFound := false

	if _, err := os.Stat(".env"); err == nil {
		// 当前目录的 .env（从根目录运行或 dist 目录运行）
		loadErr = godotenv.Load(".env")
		if loadErr == nil {
			log.Println("\n✅ 配置文件加载成功 (当前目录 .env)")
			envFileFound = true
		} else {
			log.Printf("\n⚠️  配置文件加载失败 (当前目录): %v", loadErr)
		}
	} else if _, err := os.Stat("../.env"); err == nil {
		// 上级目录的 .env（从 backend 目录运行）
		loadErr = godotenv.Load("../.env")
		if loadErr == nil {
			log.Println("\n✅ 配置文件加载成功 (上级目录 ../.env)")
			envFileFound = true
		} else {
			log.Printf("\n⚠️  配置文件加载失败 (上级目录): %v", loadErr)
		}
	}

	// 首次启动检查和提示
	if !envFileFound {
		log.Println("\n┌─────────────────────────────────────────────────────┐")
		log.Println("│ ⚠️  警告: 未找到配置文件                          │")
		log.Println("├─────────────────────────────────────────────────────┤")
		log.Println("│ 📝 首次部署步骤:                                  │")
		log.Println("│                                                     │")
		log.Println("│ 1️⃣  复制配置文件模板:                            │")
		log.Println("│    Linux/macOS:  cp .env.example .env            │")
		log.Println("│    Windows:      copy .env.example .env          │")
		log.Println("│                                                     │")
		log.Println("│ 2️⃣  编辑 .env 文件，配置必要项:                 │")
		log.Println("│    • DB_TYPE (数据库类型: sqlite 或 mysql)      │")
		log.Println("│    • JWT_SECRET (首次启动时自动生成)            │")
		log.Println("│    • OAuth 配置 (可选，用于第三方登录)         │")
		log.Println("│                                                     │")
		log.Println("│ 3️⃣  重新启动程序                                │")
		log.Println("│                                                     │")
		log.Println("│ 💡 将使用默认配置 (SQLite) 继续启动              │")
		log.Println("└─────────────────────────────────────────────────────┘")
	}

	if loadErr != nil && !os.IsNotExist(loadErr) {
		log.Printf("⚠️  配置文件读取出错: %v | 将使用系统环境变量", loadErr)
	}

	if delayRaw := os.Getenv("CHEMISTRYUNO_START_DELAY_MS"); delayRaw != "" {
		if delayMs, err := strconv.Atoi(delayRaw); err == nil && delayMs > 0 {
			log.Printf("⏳ 启动延迟 %dms，初始化中...", delayMs)
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
	}

	// 确保JWT密钥存在（首次启动自动生成）
	if err := utils.EnsureJWTSecret(); err != nil {
		log.Printf("⚠️  JWT密钥初始化失败: %v", err)
	}

	// 设置生产模式
	gin.SetMode(gin.ReleaseMode)

	// 初始化数据库
	if err := database.InitDB(""); err != nil {
		log.Fatal("❌ 数据库初始化失败: ", err)
	}
	defer database.Close()

	// 初始化日志系统
	if err := utils.InitLogger(5000); err != nil {
		log.Printf("⚠️  日志系统初始化失败: %v（将继续使用标准日志）", err)
	}
	defer utils.CloseLogger()

	// 初始化 Redis 缓存（可选）
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // 默认地址
	}

	redisUsername := os.Getenv("REDIS_USERNAME")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := 0
	if redisDBRaw := os.Getenv("REDIS_DB"); redisDBRaw != "" {
		if parsedDB, err := strconv.Atoi(redisDBRaw); err == nil {
			redisDB = parsedDB
		}
	}

	if err := cache.InitRedis(redisAddr, redisUsername, redisPassword, redisDB); err != nil {
		log.Printf("⚠️  Redis 初始化失败，将在缓存 miss 时自动降级到数据库查询: %v", err)
		// 不中断启动，允许程序在没有 Redis 的情况下运行
	} else {
		defer cache.Close()
		log.Println("✅ Redis 缓存已启用 - Session 查询性能提升")
	}

	// 迁移reactions表到R1/R2结构（如果需要）
	if err := database.MigrateReactionsToR1R2(); err != nil {
		log.Printf("ℹ️  化学反应表迁移略过: %v (首次启动或已迁移)", err)
	}

	// 初始化所有Repository（需要在数据库初始化后）
	repository.InitRepositories()

	// 启动时修复历史数据中缺失昵称的玩家，确保所有展示面都有可用昵称。
	if _, err := utils.RepairMissingNicknames(repository.UserRepo); err != nil {
		log.Fatalf("❌ 昵称启动检查失败: %v", err)
	}

	// 初始化Admin Handlers（需要在数据库初始化后）
	handlers.InitAdminHandlers()

	// 初始化游戏时间配置（需要在数据库初始化后）
	if err := game.InitGameConfig(); err != nil {
		log.Printf("ℹ️  游戏配置初始化略过: %v | 使用默认配置", err)
	}

	// 初始化反作弊系统
	if err := game.InitializeAnticheatSystem(database.DB); err != nil {
		log.Printf("⚠️  反作弊系统初始化失败: %v", err)
	} else {
		// 设置handlers包中的全局反作弊系统
		if acSystem := game.GetGlobalAnticheatSystem(); acSystem != nil {
			handlers.SetGlobalAnticheatSystem(acSystem)
			// 设置handlers包中的全局反作弊处理程序
			handlers.InitializeAnticheatHandler(acSystem)
			// 启动 Streams 反作弊消费者（仅在 ENABLE_ANTICHEAT_STREAMS=true 时激活）
			if os.Getenv("ENABLE_ANTICHEAT_STREAMS") == "true" || os.Getenv("ENABLE_ANTICHEAT_STREAMS") == "1" {
				anticheat.StartAnticheatConsumer(context.Background(), acSystem)
				log.Println("✅ Anticheat Streams consumer enabled")
			}
		}
	}

	// 加载插件卡牌 registry
	game.LoadPluginCards()

	// 自动同步物质百科（将反应中的物质录入百科）
	game.SyncSubstancesFromReactions()

	// 初始化合法物质缓存。同步完成后再构建，确保出牌校验能读取启动时补齐的 substances 表记录。
	game.RebuildSubstanceCache()

	// 标记重复物质为待完善
	game.MarkDuplicateSubstancesForImprovement()

	// 标记名称等于化学式的物质为待完善
	game.MarkIncompleteNameSubstances()

	// 自动清理非法反应和物质数据
	game.CleanInvalidData()

	// 记录启动时间
	startTime := time.Now()

	// 初始化WebSocket Hub
	hub = websocket.NewHub()
	websocket.GlobalHub = hub // 设置全局 Hub 引用
	hub.OnRegister = game.PushOnJoinAnnouncements
	go hub.Run()

	// 启动 Redis Pub/Sub 跨节点广播监听器
	hub.StartPubSubListener(context.Background())

	// 启动速率限制器清理
	go middleware.GetRateLimitStore().StartCleanup()

	// 在 WebSocket Hub 就绪后加载服务端脚本，确保 onLoad 可推送消息。
	plugins.LoadServerScripts()
	plugins.Emit(plugins.EventServerStart, map[string]interface{}{
		"started_at_unix_mil": startTime.UnixMilli(),
	})

	// 启动房间监控（处理消极游戏踢人逻辑）
	game.StartRoomMonitor()

	// 启动定时任务触发器
	game.StartCron()

	// 初始化 WebAuthn
	handlers.InitWebAuthn()

	// 初始化 OAuth
	handlers.InitOauth()

	// 创建Gin路由
	// 创建Gin引擎（不使用默认中间件）
	r := gin.New()
	r.HandleMethodNotAllowed = true

	// 设置最大请求体限制 (15MB，以容纳 10MB 的 base64 头像 + JSON 额外开销)
	r.MaxMultipartMemory = 15 << 20 // 15 MiB
	r.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 15<<20)
		c.Next()
	})

	// 添加自定义中间件
	r.Use(middleware.RequestLogger())  // 请求日志中间件
	r.Use(gin.Recovery())              // Panic恢复中间件
	r.Use(middleware.CORSMiddleware()) // CORS中间件

	// 信任本地代理，确保 c.ClientIP() 能获取到真实 IP
	r.SetTrustedProxies([]string{"127.0.0.1"})

	// API 路由组
	api := r.Group("/api")
	{
		// 健康检查接口
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})

		api.GET("/health", func(c *gin.Context) {
			// 检查数据库连接
			dbStatus := "ok"
			if database.DB != nil {
				if sqlDB, err := database.DB.DB(); err != nil {
					dbStatus = "error"
				} else if err := sqlDB.Ping(); err != nil {
					dbStatus = "error"
				}
			} else {
				dbStatus = "error"
			}

			// 检查Redis连接
			redisStatus, _ := database.GetRedisStatus(context.Background())
			healthStatus := "healthy"
			if dbStatus != "ok" {
				healthStatus = "error"
			} else if redisStatus == "error" {
				healthStatus = "degraded"
			}

			c.JSON(200, gin.H{
				"status":    healthStatus,
				"database":  dbStatus,
				"redis":     redisStatus,
				"uptime":    time.Since(startTime).String(),
				"timestamp": time.Now().Unix(),
			})
		})

		// 版本信息接口
		api.GET("/version", handlers.GetVersion)

		// 数据管理路由（物质、反应等）
		router.RegisterDataRoutes(r)

		// 公开路由 - 认证组
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", middleware.RegisterRateLimiter(), handlers.Register)
			authGroup.POST("/login", middleware.LoginRateLimiter(), handlers.Login)
			authGroup.POST("/refresh", handlers.RefreshToken)
			authGroup.GET("/config", handlers.GetAuthConfig)
			authGroup.POST("/send-code", middleware.SendCodeRateLimiter(), handlers.SendVerificationCode)
			authGroup.POST("/reset-password", handlers.ResetPasswordByEmail)
			authGroup.POST("/reset_password", handlers.ResetPasswordByEmail) // 别名兼容
			authGroup.POST("/2fa/reset-password", handlers.ResetPasswordBy2FA)
			authGroup.POST("/2fa/verify", handlers.Verify2FALogin)
			// 密保问题相关（公开，用于忘记密码）
			authGroup.GET("/security-question", handlers.GetSecurityQuestion)
			authGroup.POST("/security-question/reset-password", handlers.ResetPasswordBySecurityQuestion)

			// WebAuthn 登录 (公开)
			authGroup.GET("/webauthn/login/begin", handlers.BeginLogin)
			authGroup.POST("/webauthn/login/finish", handlers.FinishLogin)

			// WebAuthn 密码重置 (公开)
			authGroup.POST("/webauthn/reset-password/begin", handlers.BeginResetPasswordWebAuthn)
			authGroup.POST("/webauthn/reset-password/finish", handlers.FinishResetPasswordWebAuthn)

			// OAuth 登录
			authGroup.GET("/github/login", handlers.GitHubLogin)
			authGroup.GET("/github/callback", handlers.GitHubCallback)
			authGroup.GET("/ms/login", handlers.MicrosoftLogin)
			authGroup.GET("/ms/callback", handlers.MicrosoftCallback)
			authGroup.GET("/google/login", handlers.GoogleLogin)
			authGroup.GET("/google/callback", handlers.GoogleCallback)
			authGroup.GET("/apple/login", handlers.AppleLogin)
			authGroup.POST("/apple/callback", handlers.AppleCallback)
			authGroup.GET("/apple/callback", handlers.AppleCallback)
		}

		api.GET("/announcements", handlers.GetActiveAnnouncements)
		api.GET("/hints", handlers.GetRandomHints)

		// 需要认证的路由
		auth := api.Group("/")
		auth.Use(middleware.AuthMiddleware())
		{
			// OAuth 绑定
			auth.GET("/auth/github/bind", handlers.GitHubLogin)
			auth.GET("/auth/ms/bind", handlers.MicrosoftLogin)
			auth.GET("/auth/google/bind", handlers.GoogleLogin)
			auth.GET("/auth/apple/bind", handlers.AppleLogin)
			auth.POST("/auth/oauth/unbind", handlers.UnbindOAuth)

			// 用户相关
			auth.GET("/user/info", handlers.GetUserInfo)
			auth.POST("/user/change-email", handlers.ChangeEmail)
			auth.POST("/user/set-email", handlers.SetEmail)
			auth.GET("/user/security-question", handlers.GetMySecurityQuestion)
			auth.PUT("/user/security-question", handlers.UpdateSecurityQuestion)
			auth.GET("/user/game-history", handlers.GetMyGameHistory)
			auth.GET("/user/game-history/:id/replay", handlers.GetMyGameReplay)
			auth.PUT("/user/password", handlers.ChangePassword)
			auth.PUT("/user/avatar", handlers.UpdateAvatar)
			auth.PUT("/user/profile", handlers.UpdateProfile)
			auth.GET("/user/profile/:uid", handlers.GetUserProfile)
			auth.DELETE("/user/account", handlers.DeleteAccount)
			auth.GET("/users/search", handlers.SearchUsers)

			// 聊天相关
			auth.GET("/chat/global/history", handlers.GetGlobalChatHistory)
			auth.GET("/chat/private/history/:friend_uid", handlers.GetPrivateChatHistory)

			// 会话与设备管理
			auth.GET("/user/sessions", handlers.GetSessions)
			auth.POST("/user/sessions/logout", handlers.RevokeSession)
			auth.POST("/user/account/freeze", handlers.FreezeAccount)

			// 反馈
			auth.POST("/feedback", handlers.CreateFeedback)
			auth.GET("/feedbacks/my", handlers.GetMyFeedbacks)
			auth.POST("/feedbacks/:id/urge", handlers.UrgeFeedback)
			auth.POST("/feedbacks/:id/dismiss", handlers.DismissFeedback)
			auth.POST("/feedback/withdraw", handlers.WithdrawFeedback)

			// 玩家自定义卡组
			auth.GET("/my-decks", handlers.GetMyDecks)
			auth.POST("/my-decks", handlers.CreateMyDeck)
			auth.PUT("/my-decks/:id", handlers.UpdateMyDeck)
			auth.DELETE("/my-decks/:id", handlers.DeleteMyDeck)

			// 方程式相关的普通用户路由
			auth.GET("/reactions/my", handlers.GetMyReactions)
			auth.GET("/reactions/all", handlers.GetAllReactions)

			// 2FA相关
			auth.POST("/user/2fa/setup", handlers.Setup2FA)
			auth.POST("/user/2fa/enable", handlers.Enable2FA)
			auth.POST("/user/2fa/disable", handlers.Disable2FA)

			// WebAuthn 注册与管理
			auth.GET("/user/webauthn/register/begin", handlers.BeginRegistration)
			auth.POST("/user/webauthn/register/finish", handlers.FinishRegistration)
			auth.GET("/user/webauthn/credentials", handlers.ListCredentials)
			auth.DELETE("/user/webauthn/credentials/:id", handlers.RemoveCredential)

			// WebAuthn 密码修改
			auth.POST("/user/webauthn/change-password/begin", handlers.BeginChangePasswordWebAuthn)
			auth.POST("/user/webauthn/change-password/finish", handlers.FinishChangePasswordWebAuthn)

			// 好友系统
			auth.POST("/friends/request", handlers.SendFriendRequest)
			auth.GET("/friends/pending", handlers.GetPendingRequests)
			auth.POST("/friends/handle", handlers.HandleFriendRequest)
			auth.GET("/friends", handlers.GetFriendsListWithRemarks)
			auth.POST("/friends/remark", handlers.SetFriendRemark)
			auth.DELETE("/friends/:uid", handlers.DeleteFriend)

			// 游戏相关
			auth.GET("/rooms", handlers.GetRooms)
			auth.POST("/rooms", handlers.CreateRoom)
			auth.POST("/game/duel", handlers.InitiateDuel)
			auth.POST("/game/duel/respond", handlers.RespondToDuel)
			auth.GET("/rooms/:id", handlers.GetRoomState)
			auth.GET("/rooms/:id/status", handlers.CheckRoomStatus)
			auth.POST("/rooms/:id/join", handlers.JoinRoom)
			auth.POST("/rooms/:id/leave", handlers.LeaveRoom)
			auth.POST("/rooms/:id/ready", handlers.ToggleReady)
			auth.POST("/rooms/:id/start", handlers.StartGame)
			auth.POST("/rooms/:id/play", handlers.PlayCard)
			auth.POST("/rooms/:id/play-double", handlers.DoublePlay)
			auth.POST("/rooms/:id/draw", handlers.DrawCard)
			auth.GET("/rooms/:id/substances", handlers.GetAvailableSubstances)
			auth.GET("/rooms/:id/reaction-hints", handlers.GetReactionHints)
			auth.POST("/game/check-reaction", handlers.VerifyReaction)

			// 插件卡牌（deck builder 用）
			auth.GET("/plugin-cards", handlers.GetPluginCards)
			// 插件浏览（用户只读）
			auth.GET("/plugins", handlers.GetPluginsWithCards)
			auth.GET("/plugins/:id/script", handlers.GetPluginScript)
			auth.GET("/plugins/:id/settings", handlers.GetPluginSettings)

			// WebSocket
			auth.GET("/ws", handleWebSocket)

			// 反应管理路由
			reactions := auth.Group("/reactions")
			{
				reactions.GET("", handlers.GetReactions)
				reactions.POST("", handlers.AddReaction)
				reactions.POST("/batch", middleware.CoWorkerMiddleware(), handlers.BatchAddReactions)
				reactions.PUT("/:id", middleware.CoWorkerMiddleware(), handlers.UpdateReaction)
				reactions.PUT("/approve/:group_id", middleware.CoWorkerMiddleware(), handlers.ApproveReaction)
				reactions.DELETE("/:id", middleware.AdminMiddleware(), handlers.DeleteReaction)
			}

			// 问卷调查路由
			surveys := auth.Group("/surveys")
			{
				surveys.GET("/active", handlers.GetActiveSurveysForUser)
				surveys.GET("/all", handlers.GetAllActiveSurveys)
				surveys.GET("/:id", handlers.GetSurveyDetail)
				surveys.POST("/:id/submit", handlers.SubmitSurveyResponse)
				surveys.POST("/:id/dismiss", handlers.DismissSurvey)
			}

			// 反作弊系统 - 玩家端统计
			auth.GET("/player/anticheat/stats", handlers.GetPlayerAnticheatStats)
			auth.GET("/anticheat/stats", handlers.GetPlayerAnticheatStats)
			auth.GET("/player/appeals", handlers.GetPlayerAppeals)
			auth.GET("/player/appeals/entry", handlers.GetAppealEntryStatus)
			auth.POST("/player/appeals/:id/claim", handlers.ClaimAppealCompensation)
			auth.GET("/player/sanctions", handlers.GetPlayerSanctions)
			auth.POST("/game/:roomId/appeal", handlers.SubmitAppeal)

		}

		// 管理员路由
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware(), middleware.CoWorkerMiddleware())
		{
			admin.GET("/stats", handlers.GetAdminStats)
			admin.GET("/users", handlers.GetAllUsers)
			admin.POST("/users", middleware.AdminMiddleware(), handlers.CreateUser)
			admin.DELETE("/users/:uid", middleware.AdminMiddleware(), handlers.DeleteUser)
			admin.PUT("/users/:uid/password", middleware.AdminMiddleware(), handlers.AdminChangePassword)
			admin.PUT("/users/:uid/role", middleware.AdminMiddleware(), handlers.PromoteUser)
			admin.POST("/users/ban", handlers.BanUser)
			admin.POST("/users/unban", handlers.UnbanUser)
			admin.POST("/users/kick", handlers.KickPlayer)
			admin.GET("/deck-config", middleware.AdminMiddleware(), handlers.GetGlobalDeckConfig)
			admin.PUT("/deck-config", middleware.AdminMiddleware(), handlers.UpdateGlobalDeckConfig)
			admin.POST("/deck-config/reset", middleware.AdminMiddleware(), handlers.ResetGlobalDeckConfig)
			admin.GET("/game-history", middleware.AdminMiddleware(), handlers.GetGameHistory)
			admin.GET("/game-history/:id/replay", middleware.AdminMiddleware(), handlers.GetAdminGameReplay)
			admin.DELETE("/game-history/:id/replay", middleware.AdminMiddleware(), handlers.ClearAdminGameReplay)
			admin.GET("/logs", middleware.AdminMiddleware(), handlers.GetLogs)
			admin.GET("/logs/stream", middleware.AdminMiddleware(), handlers.GetLogsStream)
			admin.POST("/logs/clear", middleware.AdminMiddleware(), handlers.ClearLogs)
			admin.GET("/feedbacks", middleware.AdminMiddleware(), handlers.GetAllFeedbacks)
			admin.PUT("/feedbacks/:id/status", middleware.AdminMiddleware(), handlers.UpdateFeedbackStatus)
			admin.GET("/configs", middleware.AdminMiddleware(), handlers.GetSystemConfigs)
			admin.PUT("/configs", middleware.AdminMiddleware(), handlers.UpdateSystemConfig)
			admin.GET("/game-time-configs", middleware.AdminMiddleware(), handlers.GetGameTimeConfigs)
			admin.PUT("/game-time-configs", middleware.AdminMiddleware(), handlers.UpdateGameTimeConfig)
			// 公告管理
			admin.GET("/announcements", middleware.AdminMiddleware(), handlers.GetAllAnnouncements)
			admin.POST("/announcements", middleware.AdminMiddleware(), handlers.CreateAnnouncement)
			admin.PUT("/announcements/:id", middleware.AdminMiddleware(), handlers.UpdateAnnouncement)
			admin.PUT("/announcements/:id/status", middleware.AdminMiddleware(), handlers.UpdateAnnouncementStatus)
			admin.DELETE("/announcements/:id", middleware.AdminMiddleware(), handlers.DeleteAnnouncement)

			// Excel导出路由（管理员专用）
			admin.GET("/export/substances", middleware.AdminMiddleware(), handlers.ExportSubstancesToExcel)
			admin.GET("/export/reactions", middleware.AdminMiddleware(), handlers.ExportReactionsToExcel)
			admin.GET("/export/all", middleware.AdminMiddleware(), handlers.ExportAllDataToExcel)

			// 批量审批路由（管理员专用）
			admin.POST("/substances/batch-approve", middleware.AdminMiddleware(), handlers.BatchApproveSubstances)
			admin.POST("/substances/batch-reject", middleware.AdminMiddleware(), handlers.BatchRejectSubstances)
			admin.POST("/reactions/batch-approve", middleware.AdminMiddleware(), handlers.BatchApproveReactions)
			admin.POST("/reactions/batch-reject", handlers.BatchRejectReactions)

			// 插件系统管理
			admin.GET("/plugins", handlers.ListPlugins)
			admin.POST("/plugins", handlers.CreatePlugin)
			admin.POST("/plugins/install", handlers.InstallPlugin)
			admin.PUT("/plugins/:id", handlers.UpdatePlugin)
			admin.DELETE("/plugins/:id", handlers.DeletePlugin)
			admin.PUT("/plugins/:id/settings", handlers.UpdatePluginSettings)
			admin.GET("/plugins/:id/settings/history", handlers.GetPluginSettingsHistory)
			admin.POST("/plugins/:id/settings/rollback", handlers.RollbackPluginSettings)
			admin.GET("/plugins/:id/cards", handlers.ListPluginCards)
			admin.POST("/plugins/:id/cards", handlers.CreatePluginCard)
			admin.PUT("/plugin-cards/:id", handlers.UpdatePluginCard)
			admin.DELETE("/plugin-cards/:id", handlers.DeletePluginCard)
			admin.POST("/plugins/reload", handlers.ReloadPlugins)

			// 服务器重启管理
			admin.POST("/server/restart", handlers.ScheduleServerRestart)
			admin.POST("/server/restart/cancel", handlers.CancelServerRestart)

			// 广播系统（全局 / 房间 / 用户）
			admin.POST("/broadcast", handlers.AdminBroadcast)
			admin.GET("/rooms/active", handlers.GetActiveRooms)
			admin.POST("/rooms/:id/terminate", middleware.AdminMiddleware(), handlers.TerminateRoom)

			admin.GET("/surveys", handlers.GetSurveys)
			admin.POST("/surveys", handlers.CreateSurvey)
			admin.PUT("/surveys/:id", handlers.UpdateSurvey)
			admin.PUT("/surveys/:id/status", handlers.UpdateSurveyStatus)
			admin.DELETE("/surveys/:id", handlers.DeleteSurvey)
			admin.GET("/surveys/:id/responses", handlers.GetSurveyResponses)
			admin.POST("/surveys/:id/repair", handlers.RepairSurveyAnswers)
			admin.GET("/surveys/:id/export", middleware.AdminMiddleware(), handlers.ExportSurveyResponses)
			admin.GET("/surveys/:id/config", handlers.GetSurveyConfig) // 导出配置
			admin.POST("/surveys/import", handlers.ImportSurveyConfig) // 导入配置

			// 反作弊系统 - 管理员端
			admin.GET("/anticheat/stats", middleware.AdminMiddleware(), handlers.GetAnticheatStats)
			admin.GET("/anticheat/config", middleware.AdminMiddleware(), handlers.GetConfig)
			admin.POST("/anticheat/config", middleware.AdminMiddleware(), handlers.UpdateConfig)
			admin.GET("/anticheat/detection-list", middleware.AdminMiddleware(), handlers.GetDetectionList)
			admin.GET("/anticheat/detection/:id", middleware.AdminMiddleware(), handlers.GetDetectionDetail)
			admin.POST("/anticheat/detection/:id/review", middleware.AdminMiddleware(), handlers.ReviewDetection)
			admin.POST("/anticheat/detection/:id/punishment", middleware.AdminMiddleware(), handlers.ChangeDetectionPunishment)
			admin.POST("/anticheat/ban", middleware.AdminMiddleware(), handlers.BanFromAnticheatPanel)
			admin.POST("/anticheat/unban", middleware.AdminMiddleware(), handlers.UnbanFromAnticheatPanel)
			admin.GET("/anticheat/appeals", middleware.AdminMiddleware(), handlers.GetAppealsList)
			admin.POST("/anticheat/appeals/:id/approve", middleware.AdminMiddleware(), handlers.ApproveAppeal)
			admin.POST("/anticheat/appeals/:id/reject", middleware.AdminMiddleware(), handlers.RejectAppeal)
			admin.GET("/anticheat/audit-log", middleware.AdminMiddleware(), handlers.GetAuditLog)
			admin.GET("/anticheat/audit-log/export", middleware.AdminMiddleware(), handlers.ExportAuditLog)

			// Additional routes can be added here
		}

		// 积分和悬赏
		points := api.Group("/points")
		points.Use(middleware.AuthMiddleware())
		{
			points.GET("/leaderboard", handlers.GetLeaderboard)
			points.POST("/bounty", handlers.CreateBounty)
		}

		// 等级系统路由
		router.RegisterLevelRoutes(r)
	}

	// 服务前端静态文件（使用 embed 嵌入）
	distFS, err := static.GetDistFS()
	if err != nil {
		log.Printf("⚠ 未找到前端构建文件: %v（仅启用 API 服务）", err)
	} else {
		// 静态资源文件（带缓存头）
		r.GET("/assets/*filepath", func(c *gin.Context) {
			c.FileFromFS(c.Request.URL.Path, http.FS(distFS))
		})

		// 处理其他静态文件（如 favicon.ico）
		r.GET("/favicon.ico", func(c *gin.Context) {
			c.FileFromFS("favicon.ico", http.FS(distFS))
		})

		// SPA 路由回退 - 非 API 请求返回 index.html
		r.NoRoute(func(c *gin.Context) {
			// API 请求返回 404
			if strings.HasPrefix(c.Request.URL.Path, "/api") || strings.HasPrefix(c.Request.URL.Path, "/ws") {
				c.JSON(404, gin.H{"error": "API route not found"})
				return
			}

			// 尝试读取静态文件
			if _, err := fs.Stat(distFS, strings.TrimPrefix(c.Request.URL.Path, "/")); err == nil {
				c.FileFromFS(c.Request.URL.Path, http.FS(distFS))
				return
			}

			if !isFrontendRoute(c.Request.URL.Path) {
				c.Redirect(http.StatusFound, "/")
				return
			}

			// 读取 index.html 用于 Vue Router
			data, err := fs.ReadFile(distFS, "index.html")
			if err != nil {
				c.String(500, "Error loading page")
				return
			}
			c.Data(200, "text/html; charset=utf-8", data)
		})
		r.NoMethod(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api") || strings.HasPrefix(c.Request.URL.Path, "/ws") {
				c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
				return
			}
			c.Redirect(http.StatusFound, "/")
		})

		log.Println("✓ 前端静态文件服务已启用（embed 模式）")
	}

	log.Println("✅ 服务器准备启动在 :8080")

	// 后台清理任务：删除已到达 remove_at 的反馈（每小时运行）
	go func() {
		feedbackRepo := repository.NewFeedbackRepository()
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			<-ticker.C
			ra, err := feedbackRepo.DeleteExpired()
			if err != nil {
				log.Printf("清理过期反馈失败: %v", err)
				continue
			}
			if ra > 0 {
				log.Printf("已删除 %d 条过期反馈", ra)
			}
		}
	}()

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// 在goroutine中启动服务器
	go func() {
		log.Println("✅ 服务器启动在 :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 正在关闭服务器...")

	// 设置5秒超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭HTTP服务器
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("服务器强制关闭: %v", err)
	}

	// 关闭WebSocket hub
	if hub != nil {
		hub.Stop()
	}

	log.Println("✅ 服务器已安全关闭")
}

func isFrontendRoute(path string) bool {
	if path == "/" {
		return true
	}

	for _, route := range []string{
		"/login",
		"/register",
		"/chat",
		"/profile",
		"/appeals",
		"/feedbacks",
		"/admin",
		"/plugins",
		"/data",
		"/ranking",
		"/oauth-callback",
	} {
		if path == route || strings.HasPrefix(path, route+"/") {
			return true
		}
	}

	for _, prefix := range []string{
		"/room/",
		"/replay/",
		"/surveys/",
		"/user/",
	} {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

func handleWebSocket(c *gin.Context) {
	uid := c.GetInt("uid")
	username := c.GetString("username")

	// 获取用户头像和昵称
	avatar := "🧪"
	nickname := ""
	if user, err := repository.UserRepo.FindByUID(uint(uid)); err == nil {
		avatar = user.Avatar
		nickname = user.Nickname
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	client := websocket.NewClient(hub, conn, uid, username, nickname, avatar)
	client.SetSource(utils.SourceFromRequest(c.Request, c.ClientIP()))
	utils.LogStructured(utils.LogEntry{
		Level:     "INFO",
		Category:  "websocket",
		Message:   "websocket connect uid=" + strconv.Itoa(uid),
		UID:       utils.IntPtr(uid),
		AuthState: "authenticated",
		Source:    utils.SourceFromRequest(c.Request, c.ClientIP()),
		WebSocket: &utils.LogWebSocket{
			Event: "connect",
		},
	})
	hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}
