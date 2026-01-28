package main

import (
	"chemistryuno/database"
	"chemistryuno/handlers"
	"chemistryuno/middleware"
	"chemistryuno/websocket"
	"log"
	"net/http"

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
	}

	// 反应管理路由（co-worker和admin权限）
	reactions := r.Group("/reactions")
	reactions.Use(middleware.AuthMiddleware(), middleware.CoWorkerMiddleware())
	{
		reactions.GET("", handlers.GetReactions)
		reactions.POST("", handlers.AddReaction)
		reactions.DELETE("/:id", handlers.DeleteReaction)
	}

	log.Println("服务器启动在 :8080")
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
