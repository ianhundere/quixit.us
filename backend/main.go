package main

import (
	"log"
	"os"

	"sample-exchange/backend/api"
	"sample-exchange/backend/auth/oauth"
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

	// Set up database
	database := db.SetupDB()

	// Initialize OAuth providers
	providers := oauth.NewProviders(cfg)

	// Initialize storage
	store := storage.NewStorage(cfg)

	// Create router
	r := gin.Default()

	// Security middlewares
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORS())

	// Create handlers
	oauthHandler := api.NewOAuthHandler(database, providers)

	// Public routes
	auth := r.Group("/auth")
	{
		// OAuth routes
		auth.GET("/oauth/:provider", oauthHandler.Login)
		auth.POST("/oauth/:provider/callback", oauthHandler.Callback)
	}

	// Protected routes
	protected := r.Group("/api")
	protected.Use(middleware.Auth())
	{
		// Add protected routes here
	}

	// Setup other routes
	api.Init(r, store, cfg)

	// Start server
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
