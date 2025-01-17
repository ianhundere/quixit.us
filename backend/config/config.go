package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Security settings
	PasswordMinLength int
	PasswordMaxLength int
	MaxLoginAttempts  int
	LockoutDuration   time.Duration

	// Server settings
	Port               string
	JWTSecret          string
	JWTAccessDuration  time.Duration
	JWTRefreshDuration time.Duration

	// Email settings
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string

	// Upload window settings
	UploadStartDay time.Weekday  // Friday
	UploadDuration time.Duration // 72 hours

	// Submission window settings
	SubmissionStartDay time.Weekday  // Monday
	SubmissionDuration time.Duration // 96 hours (4 days)

	// Storage settings
	StoragePath string

	// Development settings
	DevMode           bool
	BypassTimeWindows bool
}

func LoadConfig() *Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Check required environment variables
	requiredEnvVars := []string{
		"JWT_SECRET",
		"SMTP_USERNAME",
		"SMTP_PASSWORD",
		"SMTP_HOST",
		"SMTP_FROM",
		// ... other required variables
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("Required environment variable not set: %s", envVar)
		}
	}

	return &Config{
		Port:               getEnv("PORT", "8080"),
		JWTSecret:          getEnvRequired("JWT_SECRET"),
		JWTAccessDuration:  getDuration("JWT_ACCESS_DURATION", 15*time.Minute),
		JWTRefreshDuration: getDuration("JWT_REFRESH_DURATION", 7*24*time.Hour),

		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnvInt("SMTP_PORT", 587),
		SMTPUsername: getEnvRequired("SMTP_USERNAME"),
		SMTPPassword: getEnvRequired("SMTP_PASSWORD"),
		SMTPFrom:     getEnvRequired("SMTP_FROM"),

		UploadStartDay:     time.Friday,
		UploadDuration:     72 * time.Hour,
		SubmissionStartDay: time.Monday,
		SubmissionDuration: 96 * time.Hour,
		StoragePath:        getEnv("STORAGE_PATH", "./storage"),

		// Security settings
		PasswordMinLength: 8,
		PasswordMaxLength: 72, // bcrypt max
		MaxLoginAttempts:  5,
		LockoutDuration:   15 * time.Minute,

		// Development settings
		DevMode:           os.Getenv("DEV_MODE") == "true",
		BypassTimeWindows: os.Getenv("DEV_BYPASS_TIME_WINDOWS") == "true",
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvRequired(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Fatalf("Required environment variable not set: %s", key)
	return ""
}

func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return fallback
}
