package main

import (
	"log"
	"os"

	"sample-exchange/backend/api"
	"sample-exchange/backend/config"
	"sample-exchange/backend/db"
	"sample-exchange/backend/middleware"
	"sample-exchange/backend/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	// Use debug mode in development
	if os.Getenv("GIN_MODE") != "release" {
		gin.SetMode(gin.DebugMode)
	}

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	db.Init()

	// Initialize storage
	store := storage.NewStorage(cfg)

	r := gin.Default()
	
	// Security middlewares
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.SanitizeInputs())
	
	// Rate limiting
	r.Use(middleware.RateLimitByIP(60)) // 60 requests per minute globally

	// Setup routes
	api.Init(r, store, cfg)
	
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func isAllowedOrigin(origin string) bool {
	allowedOrigins := []string{
		"http://localhost:3000",
		"https://yourdomain.com",
	}
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}
	return false
} 