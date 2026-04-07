package youtube

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	yt "google.golang.org/api/youtube/v3"
)

// OAuthConfig returns the OAuth2 configuration for YouTube API access.
func OAuthConfig() (*oauth2.Config, error) {
	clientID := os.Getenv("YOUTUBE_CLIENT_ID")
	clientSecret := os.Getenv("YOUTUBE_CLIENT_SECRET")
	redirectURL := os.Getenv("YOUTUBE_REDIRECT_URL")

	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("YOUTUBE_CLIENT_ID and YOUTUBE_CLIENT_SECRET environment variables are required")
	}

	if redirectURL == "" {
		redirectURL = "http://localhost:8080/auth/callback"
	}

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  redirectURL,
		Scopes: []string{
			yt.YoutubeScope,
			yt.YoutubeForceSslScope,
		},
	}, nil
}

// GenerateStateToken creates a random state token for OAuth2 CSRF protection.
func GenerateStateToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating state token: %w", err)
	}
	return hex.EncodeToString(b), nil
}
