package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type Config struct {
	// Server settings
	Port string
	Mode string

	// Development settings
	DevMode           bool
	BypassTimeWindows bool
	BypassOAuth       bool

	// JWT settings
	JWTSecret       string
	AccessDuration  time.Duration
	RefreshDuration time.Duration

	// Storage settings
	StoragePath string

	// OAuth settings
	OAuthRedirectURL string
	GitHub           OAuthConfig
	Google           OAuthConfig
	Discord          OAuthConfig
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	cfg := &Config{
		Port:              getEnv("PORT", "8080"),
		Mode:              getEnv("GIN_MODE", "debug"),
		DevMode:           getEnvBool("DEV_MODE", true),
		BypassTimeWindows: getEnvBool("DEV_BYPASS_TIME_WINDOWS", true),
		BypassOAuth:       getEnvBool("DEV_MODE", true),
		JWTSecret:         getEnv("JWT_SECRET", "your-super-secret-jwt-key-here"),
		AccessDuration:    getEnvDuration("JWT_ACCESS_DURATION", 15*time.Minute),
		RefreshDuration:   getEnvDuration("JWT_REFRESH_DURATION", 168*time.Hour),
		StoragePath:       getEnv("STORAGE_PATH", "./storage"),
		OAuthRedirectURL:  getEnv("OAUTH_REDIRECT_URL", "http://localhost:3000/auth/callback"),

		// OAuth Providers
		GitHub: OAuthConfig{
			ClientID:     getEnv("OAUTH_GITHUB_CLIENT_ID", ""),
			ClientSecret: getEnv("OAUTH_GITHUB_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("OAUTH_GITHUB_REDIRECT_URL", "http://localhost:3000/auth/github/callback"),
		},
		Google: OAuthConfig{
			ClientID:     getEnv("OAUTH_GOOGLE_CLIENT_ID", ""),
			ClientSecret: getEnv("OAUTH_GOOGLE_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("OAUTH_GOOGLE_REDIRECT_URL", "http://localhost:3000/auth/google/callback"),
		},
		Discord: OAuthConfig{
			ClientID:     getEnv("OAUTH_DISCORD_CLIENT_ID", ""),
			ClientSecret: getEnv("OAUTH_DISCORD_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("OAUTH_DISCORD_REDIRECT_URL", "http://localhost:3000/auth/discord/callback"),
		},
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		return value == "true" || value == "1"
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		log.Printf("Warning: invalid duration for %s, using fallback", key)
	}
	return fallback
}
