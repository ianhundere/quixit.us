package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	sync.Mutex
	requests map[string][]time.Time
	window   time.Duration
	limit    int
}

func NewRateLimiter(window time.Duration, limit int) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		window:   window,
		limit:    limit,
	}
}

func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		
		rl.Lock()
		defer rl.Unlock()

		now := time.Now()
		
		// Remove old requests
		var recent []time.Time
		for _, t := range rl.requests[ip] {
			if now.Sub(t) <= rl.window {
				recent = append(recent, t)
			}
		}
		rl.requests[ip] = recent

		// Check if limit is exceeded
		if len(recent) >= rl.limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"retry_after": rl.window.Seconds(),
			})
			c.Abort()
			return
		}

		// Add current request
		rl.requests[ip] = append(rl.requests[ip], now)
		
		c.Next()
	}
} 