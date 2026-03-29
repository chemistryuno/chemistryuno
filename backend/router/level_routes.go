package router

import (
	"chemistryuno/backend/handlers"
	"chemistryuno/backend/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterLevelRoutes registers level-related routes and keeps route wiring centralized.
func RegisterLevelRoutes(r *gin.Engine) {
	r.GET("/api/level/configs", handlers.GetAllLevelConfigs)
	r.GET("/api/level/leaderboard", handlers.GetLevelLeaderboard)
	r.GET("/api/level/user/:uid", handlers.GetUserLevelInfo)

	authGroup := r.Group("/api/level")
	authGroup.Use(middleware.AuthMiddleware())
	{
		authGroup.GET("/info", handlers.GetLevelInfo)
	}
}
