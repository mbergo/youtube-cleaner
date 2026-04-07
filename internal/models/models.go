package models

import "time"

// Video represents a YouTube video with relevant metadata.
type Video struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	ChannelTitle string    `json:"channelTitle"`
	PublishedAt  time.Time `json:"publishedAt"`
	ThumbnailURL string    `json:"thumbnailUrl"`
	ViewCount    string    `json:"viewCount"`
	Duration     string    `json:"duration"`
}

// Subscription represents a YouTube channel subscription.
type Subscription struct {
	ID           string    `json:"id"`
	ChannelID    string    `json:"channelId"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	ThumbnailURL string    `json:"thumbnailUrl"`
	PublishedAt  time.Time `json:"publishedAt"`
}

// Playlist represents a YouTube playlist.
type Playlist struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	ItemCount    int64  `json:"itemCount"`
	ThumbnailURL string `json:"thumbnailUrl"`
}

// PlaylistItem represents an item within a YouTube playlist.
type PlaylistItem struct {
	ID           string    `json:"id"`
	VideoID      string    `json:"videoId"`
	Title        string    `json:"title"`
	ChannelTitle string    `json:"channelTitle"`
	ThumbnailURL string    `json:"thumbnailUrl"`
	Position     int64     `json:"position"`
	AddedAt      time.Time `json:"addedAt"`
}

// CleanupResult holds the outcome of a cleanup operation.
type CleanupResult struct {
	Action    string `json:"action"`
	ItemID    string `json:"itemId"`
	ItemTitle string `json:"itemTitle"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

// Stats holds account statistics displayed on the dashboard.
type Stats struct {
	VideoCount        int `json:"videoCount"`
	SubscriptionCount int `json:"subscriptionCount"`
	PlaylistCount     int `json:"playlistCount"`
	LikedVideoCount   int `json:"likedVideoCount"`
}
