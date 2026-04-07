package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleIndex(t *testing.T) {
	// Set env vars so handler creation doesn't fail
	t.Setenv("YOUTUBE_CLIENT_ID", "test-id")
	t.Setenv("YOUTUBE_CLIENT_SECRET", "test-secret")

	h, err := New("../../templates")
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	h.handleIndex(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	if len(body) == 0 {
		t.Error("expected non-empty response body")
	}
}

func TestHandleIndex_NotFound(t *testing.T) {
	t.Setenv("YOUTUBE_CLIENT_ID", "test-id")
	t.Setenv("YOUTUBE_CLIENT_SECRET", "test-secret")

	h, err := New("../../templates")
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rr := httptest.NewRecorder()

	h.handleIndex(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestRequireAuth_Unauthenticated(t *testing.T) {
	t.Setenv("YOUTUBE_CLIENT_ID", "test-id")
	t.Setenv("YOUTUBE_CLIENT_SECRET", "test-secret")

	h, err := New("../../templates")
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	called := false
	wrapped := h.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rr := httptest.NewRecorder()

	wrapped(rr, req)

	if called {
		t.Error("expected handler not to be called without auth")
	}
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected redirect, got %d", rr.Code)
	}
}

func TestHandleAPIAction_MethodNotAllowed(t *testing.T) {
	t.Setenv("YOUTUBE_CLIENT_ID", "test-id")
	t.Setenv("YOUTUBE_CLIENT_SECRET", "test-secret")

	h, err := New("../../templates")
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/unsubscribe", nil)
	rr := httptest.NewRecorder()

	h.handleAPIAction(rr, req, nil)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}
