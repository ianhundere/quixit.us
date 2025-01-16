package config

import (
	"os"
	"time"
)

type Config struct {
	// Server settings
	Port string

	// Upload window settings
	UploadStartDay  time.Weekday // Friday
	UploadDuration  time.Duration // 72 hours
	
	// Submission window settings
	SubmissionStartDay  time.Weekday // Monday
	SubmissionDuration  time.Duration // 96 hours (4 days)

	// Storage settings
	StoragePath string
}

func LoadConfig() *Config {
	return &Config{
		Port:               getEnv("PORT", "8080"),
		UploadStartDay:     time.Friday,
		UploadDuration:     72 * time.Hour,
		SubmissionStartDay: time.Monday,
		SubmissionDuration: 96 * time.Hour,
		StoragePath:        getEnv("STORAGE_PATH", "./storage"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
} 