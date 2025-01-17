package api

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"sample-exchange/backend/auth"
	"sample-exchange/backend/auth/oauth"
	"sample-exchange/backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OAuthHandler struct {
	db        *gorm.DB
	providers map[string]oauth.Provider
}

func NewOAuthHandler(db *gorm.DB, providers map[string]oauth.Provider) *OAuthHandler {
	return &OAuthHandler{
		db:        db,
		providers: providers,
	}
}

// generateState creates a random state string for CSRF protection
func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Login initiates the OAuth flow for a specific provider
func (h *OAuthHandler) Login(c *gin.Context) {
	provider := c.Param("provider")

	// Handle development login
	if provider == "dev" {
		state, err := generateState()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate state"})
			return
		}

		// Redirect to frontend callback with development code
		redirectURI := "http://localhost:3000/auth/dev/login"
		c.Redirect(http.StatusTemporaryRedirect, redirectURI+"?client_id=dev-client&redirect_uri=http://localhost:3000/auth/callback&response_type=code&scope=dev&state="+state)
		return
	}

	p, exists := h.providers[provider]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported provider"})
		return
	}

	state, err := generateState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate state"})
		return
	}

	// Store state in session for validation
	c.SetCookie("oauth_state", state, 3600, "/", "", false, true)

	// Redirect to provider's auth URL
	authURL := p.GetAuthURL(state)
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// Callback handles the OAuth callback from providers
func (h *OAuthHandler) Callback(c *gin.Context) {
	provider := c.Param("provider")

	// Handle development callback
	if provider == "dev" {
		// Create development user
		user := models.User{
			Email:    "dev@example.com",
			Name:     "Development User",
			Provider: "dev",
		}

		// Generate JWT tokens
		accessToken, err := auth.GenerateToken(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": accessToken,
			"user":  user,
		})
		return
	}

	p, exists := h.providers[provider]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported provider"})
		return
	}

	// Verify state
	state, _ := c.Cookie("oauth_state")
	if state == "" || state != c.Query("state") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		return
	}

	// Clear state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	// Exchange code for token
	code := c.Query("code")
	token, err := p.ExchangeCode(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to exchange code"})
		return
	}

	// Get user info from provider
	userInfo, err := p.GetUserInfo(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info"})
		return
	}

	// Find or create user
	var user models.User
	result := h.db.Where("email = ?", userInfo.Email).First(&user)
	if result.Error == gorm.ErrRecordNotFound {
		// Create new user
		user = models.User{
			Email:    userInfo.Email,
			Name:     userInfo.Name,
			Provider: userInfo.Provider,
		}
		if err := h.db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}
	} else if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// Generate JWT token
	accessToken, err := auth.GenerateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": accessToken,
		"user":  user,
	})
}
