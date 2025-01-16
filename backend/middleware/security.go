package middleware

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	maxFileSize = 50 * 1024 * 1024 // 50MB
)

var (
	allowedAudioTypes = map[string]bool{
		"audio/wav":    true,
		"audio/x-wav":  true,
		"audio/mp3":    true,
		"audio/mpeg":   true,
		"audio/aiff":   true,
		"audio/x-aiff": true,
		"audio/flac":   true,
		"application/octet-stream": true,
	}

	allowedFileExtensions = map[string]bool{
		".wav":  true,
		".mp3":  true,
		".aiff": true,
		".flac": true,
	}
)

// SecurityHeaders adds security-related headers to all responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline';")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "microphone=(), geolocation=()")
		c.Next()
	}
}

// ValidateFileUpload validates file uploads for size and type
func ValidateFileUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.Next() // Skip validation if no file
			return
		}

		// Check file size
		if file.Size > maxFileSize {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "File size exceeds maximum limit of 50MB",
			})
			return
		}

		// Check file extension
		ext := strings.ToLower(filepath.Ext(file.Filename))
		log.Printf("File extension: %s", ext)
		if !allowedFileExtensions[ext] {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Invalid file type. Allowed types: WAV, MP3, AIFF, FLAC",
			})
			return
		}

		c.Next()
	}
}

// RateLimitByIP implements IP-based rate limiting with different tiers
func RateLimitByIP(requestsPerMinute int) gin.HandlerFunc {
	limiter := NewRateLimiter(time.Minute, requestsPerMinute)
	return limiter.RateLimit()
}

// SanitizeInputs sanitizes user inputs to prevent XSS and injection attacks
func SanitizeInputs() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize URL parameters
		for i, param := range c.Params {
			c.Params[i].Value = sanitize(param.Value)
		}

		// Sanitize query parameters
		for key, values := range c.Request.URL.Query() {
			sanitizedValues := make([]string, len(values))
			for i, value := range values {
				sanitizedValues[i] = sanitize(value)
			}
			c.Request.URL.Query()[key] = sanitizedValues
		}

		c.Next()
	}
}

func sanitize(input string) string {
	// Remove potentially dangerous characters and HTML tags
	input = strings.ReplaceAll(input, "<", "&lt;")
	input = strings.ReplaceAll(input, ">", "&gt;")
	input = strings.ReplaceAll(input, "'", "&#39;")
	input = strings.ReplaceAll(input, "\"", "&quot;")
	input = strings.ReplaceAll(input, ";", "&#59;")
	return input
} 