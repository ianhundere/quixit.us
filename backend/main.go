package main

import (
	"log"

	"sample-exchange/backend/api"
	"sample-exchange/backend/config"
	"sample-exchange/backend/db"
	"sample-exchange/backend/middleware"
	"sample-exchange/backend/storage"

	"github.com/gin-gonic/gin"
)

func main() {
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
	
	// Enable CORS
	r.Use(func(c *gin.Context) {
		// More restrictive CORS policy
		if origin := c.Request.Header.Get("Origin"); origin != "" {
			if isAllowedOrigin(origin) {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			}
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

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