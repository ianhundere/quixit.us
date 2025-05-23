package config

import (
	"log"
	"os"
	"strings"
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
		Mode:              getEnv("GIN_MODE", "release"),
		DevMode:           getEnv("DEV_MODE", "false") == "true",
		BypassTimeWindows: getEnv("BYPASS_TIME_WINDOWS", "false") == "true",
		BypassOAuth:       getEnv("BYPASS_OAUTH", "false") == "true",
		JWTSecret:         getEnv("JWT_SECRET", "your-super-secret-jwt-key-here"),
		AccessDuration:    getEnvDuration("JWT_ACCESS_DURATION", 15*time.Minute),
		RefreshDuration:   getEnvDuration("JWT_REFRESH_DURATION", 168*time.Hour),
		StoragePath:       getEnv("STORAGE_PATH", "./storage"),
		OAuthRedirectURL:  getEnv("OAUTH_REDIRECT_URL", "http://localhost:3000/auth/callback"),

		// OAuth Providers
		GitHub: OAuthConfig{
			ClientID:     getEnv("OAUTH_GITHUB_CLIENT_ID", ""),
			ClientSecret: getEnv("OAUTH_GITHUB_CLIENT_SECRET", ""),
			RedirectURL:  strings.Replace(getEnv("OAUTH_REDIRECT_URL", "http://localhost:3000/auth/callback"), "/callback", "/github/callback", 1),
		},
		Google: OAuthConfig{
			ClientID:     getEnv("OAUTH_GOOGLE_CLIENT_ID", ""),
			ClientSecret: getEnv("OAUTH_GOOGLE_CLIENT_SECRET", ""),
			RedirectURL:  strings.Replace(getEnv("OAUTH_REDIRECT_URL", "http://localhost:3000/auth/callback"), "/callback", "/google/callback", 1),
		},
		Discord: OAuthConfig{
			ClientID:     getEnv("OAUTH_DISCORD_CLIENT_ID", ""),
			ClientSecret: getEnv("OAUTH_DISCORD_CLIENT_SECRET", ""),
			RedirectURL:  strings.Replace(getEnv("OAUTH_REDIRECT_URL", "http://localhost:3000/auth/callback"), "/callback", "/discord/callback", 1),
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
