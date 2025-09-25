package http

import (
	"time"

	"github.com/gin-gonic/gin"

	"database-classifier/internal/handler"
)

type Router struct {
	databaseHandler       *handler.DatabaseHandler
	scanHandler           *handler.ScanHandler
	classificationHandler *handler.ClassificationHandler
}

func NewRouter(
	databaseHandler *handler.DatabaseHandler,
	scanHandler *handler.ScanHandler,
	classificationHandler *handler.ClassificationHandler,
) *Router {
	return &Router{
		databaseHandler:       databaseHandler,
		scanHandler:           scanHandler,
		classificationHandler: classificationHandler,
	}
}

func (r *Router) SetupRoutes() *gin.Engine {
	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())


	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"time":   time.Now(),
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Database management routes
		databases := v1.Group("/database")
		{
			databases.POST("", r.databaseHandler.CreateDatabase)
			databases.GET("", r.databaseHandler.GetAllDatabases)
			databases.GET("/:id", r.databaseHandler.GetDatabase)
			databases.PUT("/:id", r.databaseHandler.UpdateDatabase)
			databases.DELETE("/:id", r.databaseHandler.DeleteDatabase)
			databases.POST("/:id/test", r.databaseHandler.TestDatabase)

			// Scanning routes for specific database
			databases.POST("/:id/scan", r.scanHandler.StartScan)
			databases.GET("/:id/scan/history", r.scanHandler.GetScanHistory)
			databases.GET("/:id/classification", r.scanHandler.GetLatestClassification)
		}

		// Scan management routes
		scans := v1.Group("/scan")
		{
			scans.GET("/:scanId", r.scanHandler.GetScanResult)
			scans.POST("/:scanId/cancel", r.scanHandler.CancelScan)
		}

		patterns := v1.Group("/patterns")
		{
			patterns.POST("", r.classificationHandler.CreatePattern)
			patterns.GET("", r.classificationHandler.ListPatterns)
			patterns.GET("/:id", r.classificationHandler.GetPattern)
			patterns.PUT("/:id", r.classificationHandler.UpdatePattern)
			patterns.DELETE("/:id", r.classificationHandler.DeletePattern)
		}
	}

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

