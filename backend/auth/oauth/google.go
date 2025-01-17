package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"sample-exchange/backend/config"
)

type GoogleProvider struct {
	BaseProvider
}

type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func NewGoogleProvider(cfg config.OAuthConfig) Provider {
	return &GoogleProvider{
		BaseProvider: BaseProvider{
			name:         "google",
			clientID:     cfg.ClientID,
			clientSecret: cfg.ClientSecret,
			authURL:      "https://accounts.google.com/o/oauth2/v2/auth",
			tokenURL:     "https://oauth2.googleapis.com/token",
			userInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
			redirectURL:  cfg.RedirectURL,
			scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		},
	}
}

func (p *GoogleProvider) GetUserInfo(token *Token) (*UserInfo, error) {
	req, err := http.NewRequest("GET", p.userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user info request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request failed with status: %d", resp.StatusCode)
	}

	var googleUser GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	if !googleUser.VerifiedEmail {
		return nil, fmt.Errorf("email not verified")
	}

	return &UserInfo{
		ID:        googleUser.ID,
		Email:     googleUser.Email,
		Name:      googleUser.Name,
		AvatarURL: googleUser.Picture,
		Provider:  p.name,
	}, nil
}
