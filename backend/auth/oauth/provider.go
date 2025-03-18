package oauth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"sample-exchange/backend/config"
)

type Provider interface {
	GetAuthURL(state string) string
	ExchangeCode(code string) (*Token, error)
	GetUserInfo(token *Token) (*UserInfo, error)
	GetName() string
}

type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

type UserInfo struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Provider  string `json:"provider"`
}

type BaseProvider struct {
	name         string
	clientID     string
	clientSecret string
	authURL      string
	tokenURL     string
	userInfoURL  string
	redirectURL  string
	scopes       []string
}

func (p *BaseProvider) GetName() string {
	return p.name
}

func (p *BaseProvider) GetAuthURL(state string) string {
	u, _ := url.Parse(p.authURL)
	q := u.Query()
	q.Set("client_id", p.clientID)
	q.Set("redirect_uri", p.redirectURL)
	q.Set("scope", strings.Join(p.scopes, " "))
	q.Set("state", state)
	q.Set("response_type", "code")
	u.RawQuery = q.Encode()
	return u.String()
}

func (p *BaseProvider) ExchangeCode(code string) (*Token, error) {
	data := url.Values{}
	data.Set("client_id", p.clientID)
	data.Set("client_secret", p.clientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", p.redirectURL)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequest("POST", p.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &token, nil
}

func NewProviders(cfg *config.Config) map[string]Provider {
	providers := make(map[string]Provider)

	// Add development provider if OAuth bypass is enabled
	if cfg.BypassOAuth {
		providers["dev"] = NewDevProvider(config.OAuthConfig{
			RedirectURL: cfg.OAuthRedirectURL,
		})
	}

	// Add GitHub provider
	if cfg.GitHub.ClientID != "" {
		providers["github"] = NewGitHubProvider(cfg.GitHub)
	}

	// Add Google provider
	if cfg.Google.ClientID != "" {
		providers["google"] = NewGoogleProvider(cfg.Google)
	}

	// Add Discord provider
	if cfg.Discord.ClientID != "" {
		providers["discord"] = NewDiscordProvider(cfg.Discord)
	}

	return providers
}
