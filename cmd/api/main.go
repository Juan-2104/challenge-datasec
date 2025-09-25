package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"database-classifier/internal/config"
	"database-classifier/internal/handler"
	"database-classifier/internal/infrastructure/database"
	httpInfra "database-classifier/internal/infrastructure/http"
	"database-classifier/internal/repository"
	"database-classifier/internal/service"
	"database-classifier/pkg/security"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	gin.SetMode(cfg.Server.GinMode)

	encryptor, err := security.NewEncryptor(cfg.Security.EncryptionKey)
	if err != nil {
		log.Fatalf("Failed to initialize encryptor: %v", err)
	}

    metadataDB, err := database.NewMetadataDB(&cfg.MetadataDB)
    if err != nil {
        log.Fatalf("Failed to connect to metadata database: %v", err)
    }
    defer metadataDB.Close()

    // Initialize repositories
    dbConnRepo := repository.NewDatabaseConnectionRepository(metadataDB)
    scanRepo := repository.NewScanResultRepository(metadataDB)
    patternRepo := repository.NewClassificationPatternRepository(metadataDB)

    // Initialize services
    ctx := context.Background()
    classificationService, err := service.NewClassificationService(ctx, patternRepo, "configs/patterns.json")
    if err != nil {
        log.Fatalf("Failed to initialize classification service: %v", err)
    }

    databaseService := service.NewDatabaseService(dbConnRepo, encryptor)
    scanService := service.NewScanService(scanRepo, dbConnRepo, encryptor, classificationService)

    // Initialize handlers
    databaseHandler := handler.NewDatabaseHandler(databaseService)
    scanHandler := handler.NewScanHandler(scanService)
    classificationHandler := handler.NewClassificationHandler(classificationService)

	// Setup router
    router := httpInfra.NewRouter(databaseHandler, scanHandler, classificationHandler)
	engine := router.SetupRoutes()

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
