package router

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/handlers"
	"chemistryuno/backend/middleware"
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterAPIRoutes centralizes API route wiring and keeps main.go focused on bootstrap lifecycle.
func RegisterAPIRoutes(r *gin.Engine, startTime time.Time, wsHandler gin.HandlerFunc) {
	api := r.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})

		api.GET("/health", func(c *gin.Context) {
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

		api.GET("/version", handlers.GetVersion)

		RegisterDataRoutes(r)

		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", middleware.RegisterRateLimiter(), handlers.Register)
			authGroup.POST("/login", middleware.LoginRateLimiter(), handlers.Login)
			authGroup.POST("/refresh", handlers.RefreshToken)
			authGroup.GET("/config", handlers.GetAuthConfig)
			authGroup.POST("/send-code", middleware.SendCodeRateLimiter(), handlers.SendVerificationCode)
			authGroup.POST("/reset-password", handlers.ResetPasswordByEmail)
			authGroup.POST("/reset_password", handlers.ResetPasswordByEmail)
			authGroup.POST("/2fa/reset-password", handlers.ResetPasswordBy2FA)
			authGroup.POST("/2fa/verify", handlers.Verify2FALogin)
			authGroup.GET("/security-question", handlers.GetSecurityQuestion)
			authGroup.POST("/security-question/reset-password", handlers.ResetPasswordBySecurityQuestion)
			authGroup.GET("/webauthn/login/begin", handlers.BeginLogin)
			authGroup.POST("/webauthn/login/finish", handlers.FinishLogin)
			authGroup.POST("/webauthn/reset-password/begin", handlers.BeginResetPasswordWebAuthn)
			authGroup.POST("/webauthn/reset-password/finish", handlers.FinishResetPasswordWebAuthn)
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

		auth := api.Group("/")
		auth.Use(middleware.AuthMiddleware())
		{
			auth.GET("/auth/github/bind", handlers.GitHubLogin)
			auth.GET("/auth/ms/bind", handlers.MicrosoftLogin)
			auth.GET("/auth/google/bind", handlers.GoogleLogin)
			auth.GET("/auth/apple/bind", handlers.AppleLogin)
			auth.POST("/auth/oauth/unbind", handlers.UnbindOAuth)
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
			auth.GET("/chat/global/history", handlers.GetGlobalChatHistory)
			auth.GET("/chat/private/history/:friend_uid", handlers.GetPrivateChatHistory)
			auth.GET("/user/sessions", handlers.GetSessions)
			auth.POST("/user/sessions/logout", handlers.RevokeSession)
			auth.POST("/user/account/freeze", handlers.FreezeAccount)
			auth.POST("/feedback", handlers.CreateFeedback)
			auth.GET("/feedbacks/my", handlers.GetMyFeedbacks)
			auth.POST("/feedbacks/:id/urge", handlers.UrgeFeedback)
			auth.POST("/feedbacks/:id/dismiss", handlers.DismissFeedback)
			auth.POST("/feedback/withdraw", handlers.WithdrawFeedback)
			auth.GET("/my-decks", handlers.GetMyDecks)
			auth.POST("/my-decks", handlers.CreateMyDeck)
			auth.PUT("/my-decks/:id", handlers.UpdateMyDeck)
			auth.DELETE("/my-decks/:id", handlers.DeleteMyDeck)
			auth.GET("/reactions/my", handlers.GetMyReactions)
			auth.GET("/reactions/all", handlers.GetAllReactions)
			auth.POST("/user/2fa/setup", handlers.Setup2FA)
			auth.POST("/user/2fa/enable", handlers.Enable2FA)
			auth.POST("/user/2fa/disable", handlers.Disable2FA)
			auth.GET("/user/webauthn/register/begin", handlers.BeginRegistration)
			auth.POST("/user/webauthn/register/finish", handlers.FinishRegistration)
			auth.GET("/user/webauthn/credentials", handlers.ListCredentials)
			auth.DELETE("/user/webauthn/credentials/:id", handlers.RemoveCredential)
			auth.POST("/user/webauthn/change-password/begin", handlers.BeginChangePasswordWebAuthn)
			auth.POST("/user/webauthn/change-password/finish", handlers.FinishChangePasswordWebAuthn)
			auth.POST("/friends/request", handlers.SendFriendRequest)
			auth.GET("/friends/pending", handlers.GetPendingRequests)
			auth.POST("/friends/handle", handlers.HandleFriendRequest)
			auth.GET("/friends", handlers.GetFriendsListWithRemarks)
			auth.POST("/friends/remark", handlers.SetFriendRemark)
			auth.DELETE("/friends/:uid", handlers.DeleteFriend)
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
			auth.GET("/plugin-cards", handlers.GetPluginCards)
			auth.GET("/plugins", handlers.GetPluginsWithCards)
			auth.GET("/plugins/:id/script", handlers.GetPluginScript)
			auth.GET("/plugins/:id/settings", handlers.GetPluginSettings)
			auth.GET("/ws", wsHandler)

			reactions := auth.Group("/reactions")
			{
				reactions.GET("", handlers.GetReactions)
				reactions.POST("", handlers.AddReaction)
				reactions.POST("/batch", middleware.CoWorkerMiddleware(), handlers.BatchAddReactions)
				reactions.PUT("/:id", middleware.CoWorkerMiddleware(), handlers.UpdateReaction)
				reactions.PUT("/approve/:group_id", middleware.CoWorkerMiddleware(), handlers.ApproveReaction)
				reactions.DELETE("/:id", middleware.AdminMiddleware(), handlers.DeleteReaction)
			}

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
			auth.POST("/player/appeals/:id/claim", handlers.ClaimAppealCompensation)
			auth.GET("/player/sanctions", handlers.GetPlayerSanctions)
			auth.POST("/game/:roomId/appeal", handlers.SubmitAppeal)
		}

		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware(), middleware.CoWorkerMiddleware())
		{
			admin.GET("/stats", handlers.GetAdminStats)
			admin.GET("/users", handlers.GetAllUsers)
			admin.POST("/users", middleware.AdminMiddleware(), handlers.CreateUser)
			admin.DELETE("/users/:uid", middleware.AdminMiddleware(), handlers.DeleteUser)
			admin.PUT("/users/:uid/password", middleware.AdminMiddleware(), handlers.AdminChangePassword)
			admin.PUT("/users/:uid/role", handlers.PromoteUser)
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
			admin.GET("/announcements", middleware.AdminMiddleware(), handlers.GetAllAnnouncements)
			admin.POST("/announcements", middleware.AdminMiddleware(), handlers.CreateAnnouncement)
			admin.PUT("/announcements/:id", middleware.AdminMiddleware(), handlers.UpdateAnnouncement)
			admin.PUT("/announcements/:id/status", middleware.AdminMiddleware(), handlers.UpdateAnnouncementStatus)
			admin.DELETE("/announcements/:id", middleware.AdminMiddleware(), handlers.DeleteAnnouncement)
			admin.GET("/export/substances", middleware.AdminMiddleware(), handlers.ExportSubstancesToExcel)
			admin.GET("/export/reactions", middleware.AdminMiddleware(), handlers.ExportReactionsToExcel)
			admin.GET("/export/all", middleware.AdminMiddleware(), handlers.ExportAllDataToExcel)
			admin.POST("/substances/batch-approve", middleware.AdminMiddleware(), handlers.BatchApproveSubstances)
			admin.POST("/substances/batch-reject", middleware.AdminMiddleware(), handlers.BatchRejectSubstances)
			admin.POST("/reactions/batch-approve", middleware.AdminMiddleware(), handlers.BatchApproveReactions)
			admin.POST("/reactions/batch-reject", handlers.BatchRejectReactions)
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
			admin.POST("/server/restart", handlers.ScheduleServerRestart)
			admin.POST("/server/restart/cancel", handlers.CancelServerRestart)
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
			admin.GET("/surveys/:id/config", handlers.GetSurveyConfig)
			admin.POST("/surveys/import", handlers.ImportSurveyConfig)

			admin.GET("/anticheat/stats", middleware.AdminMiddleware(), handlers.GetAnticheatStats)
			admin.GET("/anticheat/config", middleware.AdminMiddleware(), handlers.GetConfig)
			admin.POST("/anticheat/config", middleware.AdminMiddleware(), handlers.UpdateConfig)
			admin.GET("/anticheat/detection-list", middleware.AdminMiddleware(), handlers.GetDetectionList)
			admin.GET("/anticheat/detection/:id", middleware.AdminMiddleware(), handlers.GetDetectionDetail)
			admin.POST("/anticheat/detection/:id/review", middleware.AdminMiddleware(), handlers.ReviewDetection)
			admin.POST("/anticheat/ban", middleware.AdminMiddleware(), handlers.BanFromAnticheatPanel)
			admin.POST("/anticheat/unban", middleware.AdminMiddleware(), handlers.UnbanFromAnticheatPanel)
			admin.GET("/anticheat/appeals", middleware.AdminMiddleware(), handlers.GetAppealsList)
			admin.POST("/anticheat/appeals/:id/approve", middleware.AdminMiddleware(), handlers.ApproveAppeal)
			admin.POST("/anticheat/appeals/:id/reject", middleware.AdminMiddleware(), handlers.RejectAppeal)
			admin.GET("/anticheat/audit-log", middleware.AdminMiddleware(), handlers.GetAuditLog)
			admin.GET("/anticheat/audit-log/export", middleware.AdminMiddleware(), handlers.ExportAuditLog)
		}

		points := api.Group("/points")
		points.Use(middleware.AuthMiddleware())
		{
			points.GET("/leaderboard", handlers.GetLeaderboard)
			points.POST("/bounty", handlers.CreateBounty)
		}

		RegisterLevelRoutes(r)
	}
}
