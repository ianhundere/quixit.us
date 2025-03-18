package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"sample-exchange/backend/config"
)

type DiscordProvider struct {
	BaseProvider
}

type DiscordUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Email         string `json:"email"`
	Verified      bool   `json:"verified"`
	Avatar        string `json:"avatar"`
}

func NewDiscordProvider(cfg config.OAuthConfig) Provider {
	return &DiscordProvider{
		BaseProvider: BaseProvider{
			name:         "discord",
			clientID:     cfg.ClientID,
			clientSecret: cfg.ClientSecret,
			authURL:      "https://discord.com/api/oauth2/authorize",
			tokenURL:     "https://discord.com/api/oauth2/token",
			userInfoURL:  "https://discord.com/api/users/@me",
			redirectURL:  cfg.RedirectURL,
			scopes:       []string{"identify", "email"},
		},
	}
}

func (p *DiscordProvider) GetUserInfo(token *Token) (*UserInfo, error) {
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

	var discordUser DiscordUser
	if err := json.NewDecoder(resp.Body).Decode(&discordUser); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	if !discordUser.Verified {
		return nil, fmt.Errorf("email not verified")
	}

	// Construct avatar URL if avatar hash is present
	avatarURL := ""
	if discordUser.Avatar != "" {
		avatarURL = fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", discordUser.ID, discordUser.Avatar)
	}

	// Construct full username with discriminator if present
	name := discordUser.Username
	if discordUser.Discriminator != "0" && discordUser.Discriminator != "" {
		name = fmt.Sprintf("%s#%s", name, discordUser.Discriminator)
	}

	return &UserInfo{
		ID:        discordUser.ID,
		Email:     discordUser.Email,
		Name:      name,
		AvatarURL: avatarURL,
		Provider:  p.name,
	}, nil
}
