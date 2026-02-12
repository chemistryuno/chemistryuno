package main

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/game"
	"chemistryuno/backend/handlers"
	"chemistryuno/backend/middleware"
	"chemistryuno/backend/repository"
	"chemistryuno/backend/static"
	"chemistryuno/backend/utils"
	"chemistryuno/backend/websocket"
	"context"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
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
			log.Println("✅ 成功从当前目录 .env 加载配置")
			envFileFound = true
		} else {
			log.Printf("⚠️  从 .env 加载配置失败: %v", loadErr)
		}
	} else if _, err := os.Stat("../.env"); err == nil {
		// 上级目录的 .env（从 backend 目录运行）
		loadErr = godotenv.Load("../.env")
		if loadErr == nil {
			log.Println("✅ 成功从上级目录 ../.env 加载配置")
			envFileFound = true
		} else {
			log.Printf("⚠️  从 ../.env 加载配置失败: %v", loadErr)
		}
	}

	// 首次启动检查和提示
	if !envFileFound {
		log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		log.Println("⚠️  警告: 未找到 .env 配置文件")
		log.Println("")
		log.Println("📝 首次部署步骤:")
		log.Println("   1. 复制配置文件模板:")
		log.Println("      Linux/macOS: cp .env.example .env")
		log.Println("      Windows:     copy .env.example .env")
		log.Println("")
		log.Println("   2. 编辑 .env 文件，配置以下必要项:")
		log.Println("      - DB_TYPE (数据库类型: sqlite 或 mysql)")
		log.Println("      - JWT_SECRET (将在首次启动时自动生成)")
		log.Println("      - OAuth 配置 (可选，用于第三方登录)")
		log.Println("")
		log.Println("   3. 重新启动程序")
		log.Println("")
		log.Println("💡 程序将使用默认配置继续启动 (SQLite 数据库)")
		log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		log.Println("")
	}

	if loadErr != nil && !os.IsNotExist(loadErr) {
		log.Printf("⚠️  加载 .env 配置文件出错: %v（将使用系统环境变量）", loadErr)
	}

	// 确保JWT密钥存在（首次启动自动生成）
	if err := utils.EnsureJWTSecret(); err != nil {
		log.Printf("警告: JWT密钥初始化失败: %v", err)
	}

	// 设置生产模式
	gin.SetMode(gin.ReleaseMode)

	// 初始化数据库
	if err := database.InitDB(""); err != nil {
		log.Fatal("数据库初始化失败:", err)
	}
	defer database.Close()

	// 迁移reactions表到R1/R2结构（如果需要）
	if err := database.MigrateReactionsToR1R2(); err != nil {
		log.Printf("警告: reactions表迁移失败: %v (如果是首次启动或已迁移过，可以忽略)", err)
	}

	// 初始化所有Repository（需要在数据库初始化后）
	repository.InitRepositories()

	// 初始化Admin Handlers（需要在数据库初始化后）
	handlers.InitAdminHandlers()

	// 初始化游戏时间配置（需要在数据库初始化后）
	if err := game.InitGameConfig(); err != nil {
		log.Printf("⚠️  游戏配置初始化失败: %v (将使用默认配置)", err)
	}

	// 初始化合法物质缓存
	game.RebuildSubstanceCache()

	// 自动同步物质百科（将反应中的物质录入百科）
	game.SyncSubstancesFromReactions()

	// 标记重复物质为待完善
	game.MarkDuplicateSubstancesForImprovement()

	// 记录启动时间
	startTime := time.Now()

	// 初始化WebSocket Hub
	hub = websocket.NewHub()
	websocket.GlobalHub = hub // 设置全局 Hub 引用
	hub.OnRegister = game.PushOnJoinAnnouncements
	go hub.Run()

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

	// 添加自定义中间件
	r.Use(gin.Logger())                // 日志中间件
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
			redisStatus := "disabled"
			if database.RedisClient != nil {
				if err := database.RedisClient.Ping(context.Background()).Err(); err == nil {
					redisStatus = "ok"
				} else {
					redisStatus = "error"
				}
			}

			c.JSON(200, gin.H{
				"status":    "healthy",
				"database":  dbStatus,
				"redis":     redisStatus,
				"uptime":    time.Since(startTime).String(),
				"timestamp": time.Now().Unix(),
			})
		})

		// 版本信息接口
		api.GET("/version", handlers.GetVersion)

		// 数据管理路由（物质、反应等）
		handlers.RegisterDataRoutes(r)

		// 公开路由 - 认证组
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", handlers.Register)
			authGroup.POST("/login", handlers.Login)
			authGroup.GET("/config", handlers.GetAuthConfig)
			authGroup.POST("/send-code", handlers.SendVerificationCode)
			authGroup.POST("/reset-password", handlers.ResetPasswordByEmail)
			authGroup.POST("/reset_password", handlers.ResetPasswordByEmail) // 别名兼容
			authGroup.POST("/2fa/reset-password", handlers.ResetPasswordBy2FA)
			authGroup.POST("/2fa/verify", handlers.Verify2FALogin)

			// WebAuthn 登录 (公开)
			authGroup.GET("/webauthn/login/begin", handlers.BeginLogin)
			authGroup.POST("/webauthn/login/finish", handlers.FinishLogin)

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
			auth.GET("/user/game-history", handlers.GetMyGameHistory)
			auth.PUT("/user/password", handlers.ChangePassword)
			auth.PUT("/user/avatar", handlers.UpdateAvatar)
			auth.PUT("/user/nickname", handlers.UpdateNickname)
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

			// 好友系统
			auth.POST("/friends/request", handlers.SendFriendRequest)
			auth.GET("/friends/pending", handlers.GetPendingRequests)
			auth.POST("/friends/handle", handlers.HandleFriendRequest)
			auth.GET("/friends", handlers.GetFriendsList)
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

		}

		// 管理员路由
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
		{
			admin.GET("/users", handlers.GetAllUsers)
			admin.POST("/users", handlers.CreateUser)
			admin.DELETE("/users/:uid", handlers.DeleteUser)
			admin.PUT("/users/:uid/password", handlers.AdminChangePassword)
			admin.PUT("/users/:uid/role", handlers.PromoteUser)
			admin.POST("/users/ban", handlers.BanUser)
			admin.POST("/users/kick", handlers.KickPlayer)
			admin.GET("/deck-config", handlers.GetGlobalDeckConfig)
			admin.PUT("/deck-config", handlers.UpdateGlobalDeckConfig)
			admin.GET("/game-history", handlers.GetGameHistory)
			admin.GET("/feedbacks", handlers.GetAllFeedbacks)
			admin.PUT("/feedbacks/:id/status", handlers.UpdateFeedbackStatus)
			admin.GET("/configs", handlers.GetSystemConfigs)
			admin.PUT("/configs", handlers.UpdateSystemConfig)
			admin.GET("/game-time-configs", handlers.GetGameTimeConfigs)
			admin.PUT("/game-time-configs", handlers.UpdateGameTimeConfig)
			// 公告管理
			admin.GET("/announcements", handlers.GetAllAnnouncements)
			admin.POST("/announcements", handlers.CreateAnnouncement)
			admin.PUT("/announcements/:id", handlers.UpdateAnnouncement)
			admin.PUT("/announcements/:id/status", handlers.UpdateAnnouncementStatus)
			admin.DELETE("/announcements/:id", handlers.DeleteAnnouncement)
		}

		// 积分和悬赏
		points := api.Group("/points")
		points.Use(middleware.AuthMiddleware())
		{
			points.GET("/leaderboard", handlers.GetLeaderboard)
			points.POST("/bounty", handlers.CreateBounty)
		}
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

			// 读取 index.html 用于 Vue Router
			data, err := fs.ReadFile(distFS, "index.html")
			if err != nil {
				c.String(500, "Error loading page")
				return
			}
			c.Data(200, "text/html; charset=utf-8", data)
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

func handleWebSocket(c *gin.Context) {
	uid := c.GetInt("uid")
	username := c.GetString("username")

	// 获取用户头像和昵称
	avatar := "🧪"
	nickname := username
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
	hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}
