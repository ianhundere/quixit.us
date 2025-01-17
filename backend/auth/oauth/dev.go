package oauth

import (
	"fmt"
	"sample-exchange/backend/config"
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

	// Return mock user data
	return &UserInfo{
		ID:        "dev-123",
		Email:     "dev@example.com",
		Name:      "Development User",
		AvatarURL: "https://www.gravatar.com/avatar/00000000000000000000000000000000?d=mp",
		Provider:  p.name,
	}, nil
}
