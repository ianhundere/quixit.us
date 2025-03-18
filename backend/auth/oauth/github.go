package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"sample-exchange/backend/config"
)

type GitHubProvider struct {
	BaseProvider
}

type GitHubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

func NewGitHubProvider(cfg config.OAuthConfig) Provider {
	return &GitHubProvider{
		BaseProvider: BaseProvider{
			name:         "github",
			clientID:     cfg.ClientID,
			clientSecret: cfg.ClientSecret,
			authURL:      "https://github.com/login/oauth/authorize",
			tokenURL:     "https://github.com/login/oauth/access_token",
			userInfoURL:  "https://api.github.com/user",
			redirectURL:  cfg.RedirectURL,
			scopes:       []string{"read:user", "user:email"},
		},
	}
}

func (p *GitHubProvider) GetUserInfo(token *Token) (*UserInfo, error) {
	req, err := http.NewRequest("GET", p.userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user info request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request failed with status: %d", resp.StatusCode)
	}

	var ghUser GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&ghUser); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	// If email is not public, fetch primary email
	if ghUser.Email == "" {
		email, err := p.getPrimaryEmail(token)
		if err != nil {
			return nil, fmt.Errorf("failed to get primary email: %w", err)
		}
		ghUser.Email = email
	}

	return &UserInfo{
		ID:        fmt.Sprintf("%d", ghUser.ID),
		Email:     ghUser.Email,
		Name:      ghUser.Name,
		AvatarURL: ghUser.AvatarURL,
		Provider:  p.name,
	}, nil
}

func (p *GitHubProvider) getPrimaryEmail(token *Token) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create email request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get emails: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("email request failed with status: %d", resp.StatusCode)
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", fmt.Errorf("failed to decode emails: %w", err)
	}

	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}

	return "", fmt.Errorf("no primary verified email found")
}
