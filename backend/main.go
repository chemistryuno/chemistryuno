package main

import (
	"chemistryuno/database"
	"chemistryuno/game"
	"chemistryuno/handlers"
	"chemistryuno/middleware"
	"chemistryuno/repository"
	"chemistryuno/utils"
	"chemistryuno/websocket"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
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
	// 加载.env文件
	if err := godotenv.Load(); err != nil {
		log.Println("警告: 未找到.env文件或加载失败，将使用环境变量或默认值")
	} else {
		log.Println("✓ 成功加载 .env 配置文件")
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

	// 初始化所有Repository（需要在数据库初始化后）
	repository.InitRepositories()

	// 初始化Admin Handlers（需要在数据库初始化后）
	handlers.InitAdminHandlers()

	// 记录启动时间
	startTime := time.Now()

	// 初始化WebSocket Hub
	hub = websocket.NewHub()
	hub.OnRegister = game.PushOnJoinAnnouncements
	go hub.Run()

	// 启动房间监控（处理消极游戏踢人逻辑）
	game.StartRoomMonitor()

	// 启动定时任务触发器
	game.StartCron()

	// 初始化 WebAuthn
	handlers.InitWebAuthn()

	// 创建Gin路由
	// 创建Gin引擎（不使用默认中间件）
	r := gin.New()

	// 添加自定义中间件
	r.Use(gin.Logger())                // 日志中间件
	r.Use(gin.Recovery())              // Panic恢复中间件
	r.Use(middleware.CORSMiddleware()) // CORS中间件

	// 信任本地代理，确保 c.ClientIP() 能获取到真实 IP
	r.SetTrustedProxies([]string{"127.0.0.1"})

	// 健康检查接口
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	r.GET("/health", func(c *gin.Context) {
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

	// 公开路由
	r.POST("/auth/register", handlers.Register)
	r.POST("/auth/login", handlers.Login)
	r.GET("/announcements", handlers.GetActiveAnnouncements)
	r.POST("/auth/2fa/verify", handlers.Verify2FALogin)
	r.POST("/auth/2fa/reset-password", handlers.ResetPasswordBy2FA)

	// WebAuthn 登录 (公开)
	r.GET("/auth/webauthn/login/begin", handlers.BeginLogin)
	r.POST("/auth/webauthn/login/finish", handlers.FinishLogin)

	// 需要认证的路由
	auth := r.Group("/")
	auth.Use(middleware.AuthMiddleware())
	{
		// 用户相关
		auth.GET("/user/info", handlers.GetUserInfo)
		auth.GET("/user/game-history", handlers.GetMyGameHistory)
		auth.PUT("/user/password", handlers.ChangePassword)
		auth.PUT("/user/avatar", handlers.UpdateAvatar)
		auth.DELETE("/user/account", handlers.DeleteAccount)
		auth.GET("/users/search", handlers.SearchUsers)

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
		auth.POST("/reactions", handlers.AddReaction)

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
		auth.DELETE("/friends/:id", handlers.DeleteFriend)

		// 游戏相关
		auth.GET("/rooms", handlers.GetRooms)
		auth.POST("/rooms", handlers.CreateRoom)
		auth.POST("/game/duel", handlers.InitiateDuel)
		auth.POST("/game/duel/respond", handlers.RespondToDuel)
		auth.GET("/rooms/:id", handlers.GetRoomState)
		auth.POST("/rooms/:id/join", handlers.JoinRoom)
		auth.POST("/rooms/:id/leave", handlers.LeaveRoom)
		auth.POST("/rooms/:id/start", handlers.StartGame)
		auth.POST("/rooms/:id/play", handlers.PlayCard)
		auth.POST("/rooms/:id/play-double", handlers.DoublePlay)
		auth.POST("/rooms/:id/draw", handlers.DrawCard)
		auth.GET("/rooms/:id/substances", handlers.GetAvailableSubstances)
		auth.POST("/game/check-reaction", handlers.VerifyReaction)

		// WebSocket
		auth.GET("/ws", handleWebSocket)

		// 反应管理路由
		reactions := auth.Group("/reactions")
		{
			reactions.GET("", handlers.GetReactions)
			reactions.POST("/batch", middleware.CoWorkerMiddleware(), handlers.BatchAddReactions)
			reactions.PUT("/:id", middleware.CoWorkerMiddleware(), handlers.UpdateReaction)
			reactions.PUT("/approve/:group_id", middleware.CoWorkerMiddleware(), handlers.ApproveReaction)
			reactions.DELETE("/:id", middleware.AdminMiddleware(), handlers.DeleteReaction)
		}

		// 物质管理路由
		substances := auth.Group("/substances")
		{
			substances.GET("", handlers.GetSubstances)
			substances.POST("", handlers.AddSubstance)
			substances.PUT("/:id", middleware.CoWorkerMiddleware(), handlers.UpdateSubstance)
			substances.PUT("/approve/:id", middleware.CoWorkerMiddleware(), handlers.ApproveSubstance)
			substances.DELETE("/:id", middleware.AdminMiddleware(), handlers.DeleteSubstance)
		}
	}

	// 管理员路由
	admin := r.Group("/admin")
	admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		admin.GET("/users", handlers.GetAllUsers)
		admin.POST("/users", handlers.CreateUser)
		admin.DELETE("/users/:id", handlers.DeleteUser)
		admin.PUT("/users/:id/password", handlers.AdminChangePassword)
		admin.PUT("/users/:id/role", handlers.PromoteUser)
		admin.POST("/users/ban", handlers.BanUser)
		admin.POST("/rooms/kick", handlers.KickPlayer)
		admin.GET("/deck-config", handlers.GetGlobalDeckConfig)
		admin.PUT("/deck-config", handlers.UpdateGlobalDeckConfig)
		admin.GET("/game-history", handlers.GetGameHistory)
		admin.GET("/feedbacks", handlers.GetAllFeedbacks)
		admin.PUT("/feedbacks/:id/status", handlers.UpdateFeedbackStatus)
		admin.GET("/configs", handlers.GetSystemConfigs)
		admin.PUT("/configs", handlers.UpdateSystemConfig)
		// 公告管理
		admin.GET("/announcements", handlers.GetAllAnnouncements)
		admin.POST("/announcements", handlers.CreateAnnouncement)
		admin.PUT("/announcements/:id/status", handlers.UpdateAnnouncementStatus)
		admin.DELETE("/announcements/:id", handlers.DeleteAnnouncement)
	}

	// 积分和悬赏
	points := r.Group("/points")
	points.Use(middleware.AuthMiddleware())
	{
		points.GET("/leaderboard", handlers.GetLeaderboard)
		points.POST("/bounty", handlers.CreateBounty)
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

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	client := websocket.NewClient(hub, conn, uid, username)
	hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}
