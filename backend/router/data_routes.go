package router

import (
	"chemistryuno/backend/handlers"
	"chemistryuno/backend/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterDataRoutes registers data-management routes and keeps routing out of handlers.
func RegisterDataRoutes(r *gin.Engine) {
	// Public route: substance name map.
	r.GET("/api/substances/names", handlers.GetSubstanceNames)

	data := r.Group("/api/data")
	data.Use(middleware.AuthMiddleware())
	{
		data.GET("/substances", handlers.GetAllSubstancesGrouped)
		data.GET("/substances/my", handlers.GetMySubstances)
		data.GET("/substances/:id/group", handlers.GetSubstancesByGroup)
		data.POST("/substances/:id/update", handlers.SubmitSubstanceUpdate)
		data.POST("/substances/new", handlers.SubmitNewSubstance)

		coworker := data.Group("")
		coworker.Use(middleware.CoWorkerMiddleware())
		{
			coworker.PUT("/substances/:id", handlers.AdminUpdateSubstance)
			coworker.POST("/substances/:id/approve", handlers.ApproveSubstanceUpdate)
			coworker.DELETE("/substances/:id/reject", handlers.RejectSubstanceUpdate)
		}
	}
}
