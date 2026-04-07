package youtube

import (
	"context"
	"fmt"
	"time"

	"github.com/mbergo/youtube-cleaner/internal/models"
	"google.golang.org/api/option"
	yt "google.golang.org/api/youtube/v3"
)

// Client wraps the YouTube Data API v3 service.
type Client struct {
	service *yt.Service
}

// NewClient creates a new YouTube API client using the provided token source.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	service, err := yt.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("creating youtube service: %w", err)
	}
	return &Client{service: service}, nil
}

// GetStats returns account statistics.
func (c *Client) GetStats(ctx context.Context) (*models.Stats, error) {
	stats := &models.Stats{}

	// Get channel info for uploaded videos count
	channelResp, err := c.service.Channels.List([]string{"statistics", "contentDetails"}).
		Mine(true).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("fetching channel stats: %w", err)
	}
	if len(channelResp.Items) > 0 {
		s := channelResp.Items[0].Statistics
		stats.VideoCount = int(s.VideoCount)
	}

	// Get subscription count
	subResp, err := c.service.Subscriptions.List([]string{"id"}).
		Mine(true).MaxResults(0).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("fetching subscription count: %w", err)
	}
	stats.SubscriptionCount = int(subResp.PageInfo.TotalResults)

	// Get playlist count
	plResp, err := c.service.Playlists.List([]string{"id"}).
		Mine(true).MaxResults(0).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("fetching playlist count: %w", err)
	}
	stats.PlaylistCount = int(plResp.PageInfo.TotalResults)

	return stats, nil
}

// ListSubscriptions returns the user's channel subscriptions.
func (c *Client) ListSubscriptions(ctx context.Context, pageToken string) ([]models.Subscription, string, error) {
	call := c.service.Subscriptions.List([]string{"snippet"}).
		Mine(true).MaxResults(50).Context(ctx)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	resp, err := call.Do()
	if err != nil {
		return nil, "", fmt.Errorf("listing subscriptions: %w", err)
	}

	subs := make([]models.Subscription, 0, len(resp.Items))
	for _, item := range resp.Items {
		pub, _ := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		subs = append(subs, models.Subscription{
			ID:           item.Id,
			ChannelID:    item.Snippet.ResourceId.ChannelId,
			Title:        item.Snippet.Title,
			Description:  item.Snippet.Description,
			ThumbnailURL: thumbnailURL(item.Snippet.Thumbnails),
			PublishedAt:  pub,
		})
	}

	return subs, resp.NextPageToken, nil
}

// Unsubscribe removes a subscription by its ID.
func (c *Client) Unsubscribe(ctx context.Context, subscriptionID string) error {
	err := c.service.Subscriptions.Delete(subscriptionID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("unsubscribing %s: %w", subscriptionID, err)
	}
	return nil
}

// ListPlaylists returns the user's playlists.
func (c *Client) ListPlaylists(ctx context.Context, pageToken string) ([]models.Playlist, string, error) {
	call := c.service.Playlists.List([]string{"snippet", "contentDetails"}).
		Mine(true).MaxResults(50).Context(ctx)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	resp, err := call.Do()
	if err != nil {
		return nil, "", fmt.Errorf("listing playlists: %w", err)
	}

	playlists := make([]models.Playlist, 0, len(resp.Items))
	for _, item := range resp.Items {
		playlists = append(playlists, models.Playlist{
			ID:           item.Id,
			Title:        item.Snippet.Title,
			Description:  item.Snippet.Description,
			ItemCount:    item.ContentDetails.ItemCount,
			ThumbnailURL: thumbnailURL(item.Snippet.Thumbnails),
		})
	}

	return playlists, resp.NextPageToken, nil
}

// DeletePlaylist removes a playlist by its ID.
func (c *Client) DeletePlaylist(ctx context.Context, playlistID string) error {
	err := c.service.Playlists.Delete(playlistID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting playlist %s: %w", playlistID, err)
	}
	return nil
}

// ListPlaylistItems returns items in a specific playlist.
func (c *Client) ListPlaylistItems(ctx context.Context, playlistID, pageToken string) ([]models.PlaylistItem, string, error) {
	call := c.service.PlaylistItems.List([]string{"snippet"}).
		PlaylistId(playlistID).MaxResults(50).Context(ctx)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	resp, err := call.Do()
	if err != nil {
		return nil, "", fmt.Errorf("listing playlist items: %w", err)
	}

	items := make([]models.PlaylistItem, 0, len(resp.Items))
	for _, item := range resp.Items {
		added, _ := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		items = append(items, models.PlaylistItem{
			ID:           item.Id,
			VideoID:      item.Snippet.ResourceId.VideoId,
			Title:        item.Snippet.Title,
			ChannelTitle: item.Snippet.VideoOwnerChannelTitle,
			ThumbnailURL: thumbnailURL(item.Snippet.Thumbnails),
			Position:     item.Snippet.Position,
			AddedAt:      added,
		})
	}

	return items, resp.NextPageToken, nil
}

// RemovePlaylistItem removes an item from a playlist.
func (c *Client) RemovePlaylistItem(ctx context.Context, playlistItemID string) error {
	err := c.service.PlaylistItems.Delete(playlistItemID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing playlist item %s: %w", playlistItemID, err)
	}
	return nil
}

// ListVideos returns the user's uploaded videos.
func (c *Client) ListVideos(ctx context.Context, pageToken string) ([]models.Video, string, error) {
	// First get the uploads playlist ID
	channelResp, err := c.service.Channels.List([]string{"contentDetails"}).
		Mine(true).Context(ctx).Do()
	if err != nil {
		return nil, "", fmt.Errorf("fetching channel: %w", err)
	}
	if len(channelResp.Items) == 0 {
		return nil, "", nil
	}

	uploadsPlaylistID := channelResp.Items[0].ContentDetails.RelatedPlaylists.Uploads

	call := c.service.PlaylistItems.List([]string{"snippet"}).
		PlaylistId(uploadsPlaylistID).MaxResults(50).Context(ctx)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	resp, err := call.Do()
	if err != nil {
		return nil, "", fmt.Errorf("listing uploaded videos: %w", err)
	}

	videos := make([]models.Video, 0, len(resp.Items))
	for _, item := range resp.Items {
		pub, _ := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		videos = append(videos, models.Video{
			ID:           item.Snippet.ResourceId.VideoId,
			Title:        item.Snippet.Title,
			Description:  item.Snippet.Description,
			ChannelTitle: item.Snippet.ChannelTitle,
			PublishedAt:  pub,
			ThumbnailURL: thumbnailURL(item.Snippet.Thumbnails),
		})
	}

	return videos, resp.NextPageToken, nil
}

// DeleteVideo deletes a video by its ID.
func (c *Client) DeleteVideo(ctx context.Context, videoID string) error {
	err := c.service.Videos.Delete(videoID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting video %s: %w", videoID, err)
	}
	return nil
}

// LikeVideo rates a video as "like".
func (c *Client) LikeVideo(ctx context.Context, videoID string) error {
	return c.service.Videos.Rate(videoID, "like").Context(ctx).Do()
}

// RemoveLike removes a like/dislike rating from a video (sets to "none").
func (c *Client) RemoveLike(ctx context.Context, videoID string) error {
	return c.service.Videos.Rate(videoID, "none").Context(ctx).Do()
}

func thumbnailURL(t *yt.ThumbnailDetails) string {
	if t == nil {
		return ""
	}
	if t.Medium != nil {
		return t.Medium.Url
	}
	if t.Default != nil {
		return t.Default.Url
	}
	if t.High != nil {
		return t.High.Url
	}
	return ""
}
