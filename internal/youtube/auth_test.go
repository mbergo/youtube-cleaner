package youtube

import (
	"testing"
)

func TestGenerateStateToken(t *testing.T) {
	token1, err := GenerateStateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(token1) != 32 { // 16 bytes = 32 hex characters
		t.Errorf("expected token length 32, got %d", len(token1))
	}

	token2, err := GenerateStateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token1 == token2 {
		t.Error("expected unique tokens, got duplicates")
	}
}

func TestOAuthConfig_MissingEnvVars(t *testing.T) {
	// With no env vars set, OAuthConfig should return an error
	t.Setenv("YOUTUBE_CLIENT_ID", "")
	t.Setenv("YOUTUBE_CLIENT_SECRET", "")

	_, err := OAuthConfig()
	if err == nil {
		t.Error("expected error when env vars are missing")
	}
}

func TestOAuthConfig_WithEnvVars(t *testing.T) {
	t.Setenv("YOUTUBE_CLIENT_ID", "test-client-id")
	t.Setenv("YOUTUBE_CLIENT_SECRET", "test-client-secret")
	t.Setenv("YOUTUBE_REDIRECT_URL", "http://localhost:9090/callback")

	cfg, err := OAuthConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ClientID != "test-client-id" {
		t.Errorf("expected ClientID 'test-client-id', got '%s'", cfg.ClientID)
	}
	if cfg.ClientSecret != "test-client-secret" {
		t.Errorf("expected ClientSecret 'test-client-secret', got '%s'", cfg.ClientSecret)
	}
	if cfg.RedirectURL != "http://localhost:9090/callback" {
		t.Errorf("expected RedirectURL 'http://localhost:9090/callback', got '%s'", cfg.RedirectURL)
	}
}

func TestOAuthConfig_DefaultRedirectURL(t *testing.T) {
	t.Setenv("YOUTUBE_CLIENT_ID", "test-id")
	t.Setenv("YOUTUBE_CLIENT_SECRET", "test-secret")
	t.Setenv("YOUTUBE_REDIRECT_URL", "")

	cfg, err := OAuthConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.RedirectURL != "http://localhost:8080/auth/callback" {
		t.Errorf("expected default redirect URL, got '%s'", cfg.RedirectURL)
	}
}
