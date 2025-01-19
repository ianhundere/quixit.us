package oauth

import (
	"fmt"
	"net/url"
	"sample-exchange/backend/config"
	"sample-exchange/backend/models"
)

type DevProvider struct {
	BaseProvider
}

func NewDevProvider(cfg config.OAuthConfig) Provider {
	return &DevProvider{
		BaseProvider: BaseProvider{
			name:         "dev",
			clientID:     "dev-client",
			clientSecret: "dev-secret",
			authURL:      "http://localhost:3000/auth/dev/login",
			tokenURL:     "http://localhost:3000/auth/dev/token",
			userInfoURL:  "http://localhost:3000/auth/dev/userinfo",
			redirectURL:  cfg.RedirectURL,
			scopes:       []string{"dev"},
		},
	}
}

func (p *DevProvider) GetAuthURL(state string) string {
	params := url.Values{}
	params.Add("client_id", p.clientID)
	params.Add("redirect_uri", p.redirectURL)
	params.Add("response_type", "code")
	params.Add("scope", "dev")
	params.Add("state", state)
	return fmt.Sprintf("%s?%s", p.authURL, params.Encode())
}

func (p *DevProvider) ExchangeCode(code string) (*Token, error) {
	// In development mode, we just return a mock token
	return &Token{
		AccessToken: "dev-token",
		TokenType:   "Bearer",
		Scope:       "dev",
	}, nil
}

func (p *DevProvider) GetUserInfo(token *Token) (*UserInfo, error) {
	// Check if this is our dev token
	if token.AccessToken != "dev-token" {
		return nil, fmt.Errorf("invalid dev token")
	}

	// Return development user info
	return &UserInfo{
		Email:    "dev@example.com",
		Name:     "Development User",
		Provider: "dev",
	}, nil
}

func (p *DevProvider) GetOrCreateUser(userInfo *UserInfo) (*models.User, error) {
	// In development mode, always return the same user
	return &models.User{
		Email:    userInfo.Email,
		Name:     userInfo.Name,
		Provider: userInfo.Provider,
	}, nil
}
