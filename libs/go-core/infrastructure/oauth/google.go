// Package oauth wraps the Google OAuth2 authorization-code flow: building
// the consent URL and exchanging a callback code for the user's profile.
package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	googleoauth "golang.org/x/oauth2/google"
)

type GoogleProfile struct {
	Sub     string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type GoogleClient struct {
	config *oauth2.Config
}

func NewGoogleClient(clientID, clientSecret, redirectURL string) *GoogleClient {
	return &GoogleClient{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     googleoauth.Endpoint,
		},
	}
}

func (c *GoogleClient) AuthURL(state string) string {
	return c.config.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (c *GoogleClient) Exchange(ctx context.Context, code string) (*GoogleProfile, error) {
	token, err := c.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("oauth: exchange code: %w", err)
	}

	client := c.config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, fmt.Errorf("oauth: fetch userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("oauth: userinfo status %d: %s", resp.StatusCode, body)
	}

	var profile GoogleProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, fmt.Errorf("oauth: decode userinfo: %w", err)
	}
	return &profile, nil
}
