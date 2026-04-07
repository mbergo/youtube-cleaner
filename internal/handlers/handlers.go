package handlers

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/mbergo/youtube-cleaner/internal/youtube"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

// Handler holds shared state for HTTP handlers.
type Handler struct {
	tmpl        *template.Template
	oauthConfig *oauth2.Config

	mu       sync.RWMutex
	sessions map[string]*oauth2.Token // cookie -> token
	states   map[string]bool          // CSRF state tokens
}

// New creates a new Handler with parsed templates.
func New(templateDir string) (*Handler, error) {
	tmpl, err := template.ParseGlob(filepath.Join(templateDir, "*.html"))
	if err != nil {
		return nil, err
	}

	oauthCfg, err := youtube.OAuthConfig()
	if err != nil {
		return nil, err
	}

	return &Handler{
		tmpl:        tmpl,
		oauthConfig: oauthCfg,
		sessions:    make(map[string]*oauth2.Token),
		states:      make(map[string]bool),
	}, nil
}

// RegisterRoutes sets up all HTTP routes.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", h.handleIndex)
	mux.HandleFunc("/auth/login", h.handleLogin)
	mux.HandleFunc("/auth/callback", h.handleCallback)
	mux.HandleFunc("/auth/logout", h.handleLogout)
	mux.HandleFunc("/dashboard", h.requireAuth(h.handleDashboard))
	mux.HandleFunc("/subscriptions", h.requireAuth(h.handleSubscriptions))
	mux.HandleFunc("/playlists", h.requireAuth(h.handlePlaylists))
	mux.HandleFunc("/videos", h.requireAuth(h.handleVideos))
	mux.HandleFunc("/api/unsubscribe", h.requireAuth(h.handleUnsubscribe))
	mux.HandleFunc("/api/delete-playlist", h.requireAuth(h.handleDeletePlaylist))
	mux.HandleFunc("/api/delete-video", h.requireAuth(h.handleDeleteVideo))
	mux.HandleFunc("/api/remove-playlist-item", h.requireAuth(h.handleRemovePlaylistItem))
}

func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	h.render(w, "index.html", nil)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := youtube.GenerateStateToken()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.mu.Lock()
	h.states[state] = true
	h.mu.Unlock()

	url := h.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *Handler) handleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	h.mu.Lock()
	valid := h.states[state]
	delete(h.states, state)
	h.mu.Unlock()

	if !valid {
		http.Error(w, "Invalid state token", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	token, err := h.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("OAuth exchange error: %v", err)
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	sessionID, err := youtube.GenerateStateToken()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.mu.Lock()
	h.sessions[sessionID] = token
	h.mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		h.mu.Lock()
		delete(h.sessions, cookie.Value)
		h.mu.Unlock()
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (h *Handler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	client, err := h.ytClient(r)
	if err != nil {
		http.Error(w, "Failed to create YouTube client", http.StatusInternalServerError)
		return
	}

	stats, err := client.GetStats(r.Context())
	if err != nil {
		log.Printf("Error getting stats: %v", err)
		h.render(w, "dashboard.html", map[string]interface{}{"Error": "Failed to load statistics"})
		return
	}

	h.render(w, "dashboard.html", map[string]interface{}{"Stats": stats})
}

func (h *Handler) handleSubscriptions(w http.ResponseWriter, r *http.Request) {
	client, err := h.ytClient(r)
	if err != nil {
		http.Error(w, "Failed to create YouTube client", http.StatusInternalServerError)
		return
	}

	pageToken := r.URL.Query().Get("pageToken")
	subs, nextPage, err := client.ListSubscriptions(r.Context(), pageToken)
	if err != nil {
		log.Printf("Error listing subscriptions: %v", err)
		h.render(w, "subscriptions.html", map[string]interface{}{"Error": "Failed to load subscriptions"})
		return
	}

	h.render(w, "subscriptions.html", map[string]interface{}{
		"Subscriptions": subs,
		"NextPageToken": nextPage,
	})
}

func (h *Handler) handlePlaylists(w http.ResponseWriter, r *http.Request) {
	client, err := h.ytClient(r)
	if err != nil {
		http.Error(w, "Failed to create YouTube client", http.StatusInternalServerError)
		return
	}

	pageToken := r.URL.Query().Get("pageToken")
	playlists, nextPage, err := client.ListPlaylists(r.Context(), pageToken)
	if err != nil {
		log.Printf("Error listing playlists: %v", err)
		h.render(w, "playlists.html", map[string]interface{}{"Error": "Failed to load playlists"})
		return
	}

	h.render(w, "playlists.html", map[string]interface{}{
		"Playlists":     playlists,
		"NextPageToken": nextPage,
	})
}

func (h *Handler) handleVideos(w http.ResponseWriter, r *http.Request) {
	client, err := h.ytClient(r)
	if err != nil {
		http.Error(w, "Failed to create YouTube client", http.StatusInternalServerError)
		return
	}

	pageToken := r.URL.Query().Get("pageToken")
	videos, nextPage, err := client.ListVideos(r.Context(), pageToken)
	if err != nil {
		log.Printf("Error listing videos: %v", err)
		h.render(w, "videos.html", map[string]interface{}{"Error": "Failed to load videos"})
		return
	}

	h.render(w, "videos.html", map[string]interface{}{
		"Videos":        videos,
		"NextPageToken": nextPage,
	})
}

func (h *Handler) handleUnsubscribe(w http.ResponseWriter, r *http.Request) {
	h.handleAPIAction(w, r, func(client *youtube.Client, id string) error {
		return client.Unsubscribe(r.Context(), id)
	})
}

func (h *Handler) handleDeletePlaylist(w http.ResponseWriter, r *http.Request) {
	h.handleAPIAction(w, r, func(client *youtube.Client, id string) error {
		return client.DeletePlaylist(r.Context(), id)
	})
}

func (h *Handler) handleDeleteVideo(w http.ResponseWriter, r *http.Request) {
	h.handleAPIAction(w, r, func(client *youtube.Client, id string) error {
		return client.DeleteVideo(r.Context(), id)
	})
}

func (h *Handler) handleRemovePlaylistItem(w http.ResponseWriter, r *http.Request) {
	h.handleAPIAction(w, r, func(client *youtube.Client, id string) error {
		return client.RemovePlaylistItem(r.Context(), id)
	})
}

func (h *Handler) handleAPIAction(w http.ResponseWriter, r *http.Request, action func(*youtube.Client, string) error) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	client, err := h.ytClient(r)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create client"})
		return
	}

	if err := action(client, req.ID); err != nil {
		log.Printf("API action error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// requireAuth wraps a handler to enforce authentication.
func (h *Handler) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		h.mu.RLock()
		_, ok := h.sessions[cookie.Value]
		h.mu.RUnlock()

		if !ok {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		next(w, r)
	}
}

func (h *Handler) ytClient(r *http.Request) (*youtube.Client, error) {
	cookie, _ := r.Cookie("session")
	h.mu.RLock()
	token := h.sessions[cookie.Value]
	h.mu.RUnlock()

	tokenSource := h.oauthConfig.TokenSource(r.Context(), token)
	return youtube.NewClient(r.Context(), option.WithTokenSource(tokenSource))
}

func (h *Handler) render(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("Template render error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}
