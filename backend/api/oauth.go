package api

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"

	"sample-exchange/backend/auth"
	"sample-exchange/backend/auth/oauth"
	"sample-exchange/backend/services/user"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OAuthHandler struct {
	db          *gorm.DB
	providers   map[string]oauth.Provider
	userSvc     *user.Service
	frontendURL string
}

func NewOAuthHandler(db *gorm.DB, providers map[string]oauth.Provider, redirectURL string) *OAuthHandler {
	return &OAuthHandler{
		db:          db,
		providers:   providers,
		userSvc:     user.NewService(),
		frontendURL: redirectURL,
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
		// Create or get development user
		user, err := h.userSvc.GetOrCreateOAuthUser("dev@example.com", "Development User", "dev", "")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}

		// Generate JWT token
		token, err := auth.GenerateToken(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}

		// Build redirect URL with query parameters
		redirectURL, err := url.Parse(h.frontendURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid redirect URL"})
			return
		}
		q := redirectURL.Query()
		q.Set("token", token)
		q.Set("provider", "dev")
		redirectURL.RawQuery = q.Encode()

		c.Redirect(http.StatusTemporaryRedirect, redirectURL.String())
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
	c.SetCookie("oauth_state", state, 3600, "/api", "", true, true)

	// Redirect to provider's auth URL
	authURL := p.GetAuthURL(state)
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// Callback handles the OAuth callback from the provider
func (h *OAuthHandler) Callback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")

	// Verify state if available
	storedState, err := c.Cookie("oauth_state")
	if err != nil {
		// If cookie is not available, proceed without state validation
		// This is not ideal for security but necessary for cross-domain OAuth
		fmt.Printf("Warning: OAuth state validation skipped - %v\n", err)
	} else if state != storedState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state parameter"})
		return
	}

	p, exists := h.providers[provider]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported provider"})
		return
	}

	// Exchange code for token
	token, err := p.ExchangeCode(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to exchange code: %v", err)})
		return
	}

	// Get user info from provider
	userInfo, err := p.GetUserInfo(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get user info: %v", err)})
		return
	}

	// Create or update user in database
	user, err := h.userSvc.GetOrCreateOAuthUser(userInfo.Email, userInfo.Name, provider, userInfo.AvatarURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create/update user"})
		return
	}

	// Generate JWT token
	jwtToken, err := auth.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Build redirect URL with query parameters
	redirectURL, err := url.Parse(h.frontendURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid redirect URL"})
		return
	}
	q := redirectURL.Query()
	q.Set("token", jwtToken)
	q.Set("provider", provider)
	redirectURL.RawQuery = q.Encode()

	c.Redirect(http.StatusTemporaryRedirect, redirectURL.String())
}
