package main

import (
	"chemistryuno/database"
	"chemistryuno/handlers"
	"chemistryuno/middleware"
	"chemistryuno/websocket"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
)

var upgrader = ws.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var hub *websocket.Hub

func main() {
	// 初始化数据库
	if err := database.InitDB("./data.db"); err != nil {
		log.Fatal("数据库初始化失败:", err)
	}
	defer database.Close()

	// 初始化WebSocket Hub
	hub = websocket.NewHub()
	go hub.Run()

	// 创建Gin路由
	r := gin.Default()

	// 中间件
	r.Use(middleware.CORSMiddleware())

	// 公开路由
	r.POST("/auth/register", handlers.Register)
	r.POST("/auth/login", handlers.Login)
	r.POST("/auth/2fa/verify", handlers.Verify2FALogin)

	// 需要认证的路由
	auth := r.Group("/")
	auth.Use(middleware.AuthMiddleware())
	{
		// 用户相关
		auth.GET("/user/info", handlers.GetUserInfo)
		auth.PUT("/user/password", handlers.ChangePassword)
		auth.PUT("/user/avatar", handlers.UpdateAvatar)
		auth.DELETE("/user/account", handlers.DeleteAccount)

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

		// 游戏相关
		auth.GET("/rooms", handlers.GetRooms)
		auth.POST("/rooms", handlers.CreateRoom)
		auth.GET("/rooms/:id", handlers.GetRoomState)
		auth.POST("/rooms/:id/join", handlers.JoinRoom)
		auth.POST("/rooms/:id/leave", handlers.LeaveRoom)
		auth.POST("/rooms/:id/start", handlers.StartGame)
		auth.POST("/rooms/:id/play", handlers.PlayCard)
		auth.POST("/rooms/:id/draw", handlers.DrawCard)
		auth.GET("/rooms/:id/substances", handlers.GetAvailableSubstances)
		auth.POST("/game/check-reaction", handlers.VerifyReaction)

		// WebSocket
		auth.GET("/ws", handleWebSocket)
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
		admin.GET("/deck-config", handlers.GetGlobalDeckConfig)
		admin.PUT("/deck-config", handlers.UpdateGlobalDeckConfig)
		admin.GET("/game-history", handlers.GetGameHistory)
		admin.GET("/feedbacks", handlers.GetAllFeedbacks)
		admin.PUT("/feedbacks/:id/status", handlers.UpdateFeedbackStatus)
	}

	// 反应管理路由（co-worker和admin权限）
	reactions := r.Group("/reactions")
	reactions.Use(middleware.AuthMiddleware(), middleware.CoWorkerMiddleware())
	{
		reactions.GET("", handlers.GetReactions)
		reactions.POST("/batch", handlers.BatchAddReactions)
		reactions.PUT("/:id", handlers.UpdateReaction)
		reactions.PUT("/approve/:group_id", handlers.ApproveReaction)
		reactions.DELETE("/:id", handlers.DeleteReaction)
	}

	log.Println("服务器启动在 :8080")

	// 后台清理任务：删除已到达 remove_at 的反馈（每小时运行）
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			<-ticker.C
			nowStr := time.Now().UTC().Format("2006-01-02 15:04:05")
			res, err := database.DB.Exec("DELETE FROM feedbacks WHERE remove_at IS NOT NULL AND remove_at <= ?", nowStr)
			if err != nil {
				log.Printf("清理过期反馈失败: %v", err)
				continue
			}
			if ra, _ := res.RowsAffected(); ra > 0 {
				log.Printf("已删除 %d 条过期反馈", ra)
			}
		}
	}()

	r.Run(":8080")
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
