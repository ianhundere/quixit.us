package main

import (
	"log"
	"net/http"
	"time"

	"sample-exchange/backend/api"
	"sample-exchange/backend/auth/oauth"
	"sample-exchange/backend/config"
	"sample-exchange/backend/db"
	"sample-exchange/backend/middleware"
	"sample-exchange/backend/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Setup database
	if err := db.SetupDB(); err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}

	// Initialize storage
	store := storage.NewStorage(cfg)

	// Initialize router
	r := gin.Default()

	// Setup CORS
	r.Use(middleware.CORS())

	// Serve static files
	r.Static("/assets", "./frontend/dist/assets")
	r.StaticFile("/favicon.ico", "./frontend/dist/favicon.ico")

	// Initialize OAuth providers
	providers := map[string]oauth.Provider{
		"dev": oauth.NewDevProvider(config.OAuthConfig{
			RedirectURL: cfg.OAuthRedirectURL,
		}),
		"discord": oauth.NewDiscordProvider(config.OAuthConfig{
			ClientID:     cfg.Discord.ClientID,
			ClientSecret: cfg.Discord.ClientSecret,
			RedirectURL:  cfg.Discord.RedirectURL,
		}),
		"github": oauth.NewGitHubProvider(config.OAuthConfig{
			ClientID:     cfg.GitHub.ClientID,
			ClientSecret: cfg.GitHub.ClientSecret,
			RedirectURL:  cfg.GitHub.RedirectURL,
		}),
		"google": oauth.NewGoogleProvider(config.OAuthConfig{
			ClientID:     cfg.Google.ClientID,
			ClientSecret: cfg.Google.ClientSecret,
			RedirectURL:  cfg.Google.RedirectURL,
		}),
	}

	// Initialize handlers
	oauthHandler := api.NewOAuthHandler(db.GetDB(), providers, cfg.OAuthRedirectURL)

	// API routes
	apiGroup := r.Group("/api")
	{
		// OAuth routes
		auth := apiGroup.Group("/auth")
		{
			oauth := auth.Group("/oauth")
			{
				oauth.GET("/:provider", oauthHandler.Login)
				oauth.GET("/:provider/callback", oauthHandler.Callback)
			}
		}
	}

	// Initialize other API routes
	api.Init(r, store, cfg)

	// Health check endpoint that matches the one in the K8s config
	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// SPA fallback route
	r.NoRoute(func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
	})

	// Start server
	log.Printf("Starting server on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
