<p align="center">
  <img src="static/icon.svg" alt="YouTube Cleaner" width="128" height="128">
</p>

<h1 align="center">YouTube Cleaner</h1>

<p align="center">
  A web application to clean up your YouTube account — manage subscriptions, playlists, and uploaded videos in one place.
</p>

---

## Features

- **🧹 Bulk Unsubscribe** — Review and unsubscribe from channels you no longer watch
- **📋 Playlist Management** — View, organize, and delete playlists
- **🎬 Video Management** — Browse and delete your uploaded videos
- **📊 Dashboard** — See an overview of your YouTube account statistics
- **🔒 Secure** — OAuth2 authentication with Google; no passwords stored

## Quick Start

### Prerequisites

- [Go 1.23+](https://go.dev/dl/) (or Docker)
- A [Google Cloud project](https://console.cloud.google.com/) with the YouTube Data API v3 enabled
- OAuth2 credentials (Client ID and Client Secret)

### 1. Set Up Google OAuth2

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project (or select an existing one)
3. Enable the **YouTube Data API v3**
4. Go to **APIs & Services → Credentials**
5. Create an **OAuth 2.0 Client ID** (Web application)
6. Add `http://localhost:8080/auth/callback` as an authorized redirect URI

### 2. Configure Environment

```bash
export YOUTUBE_CLIENT_ID="your-client-id"
export YOUTUBE_CLIENT_SECRET="your-client-secret"
export YOUTUBE_REDIRECT_URL="http://localhost:8080/auth/callback"  # optional, this is the default
```

### 3. Run

**With Go:**

```bash
make run
```

**With Docker:**

```bash
docker compose up
```

Then open [http://localhost:8080](http://localhost:8080) in your browser.

## Project Structure

```
youtube-cleaner/
├── cmd/server/          # Application entry point
│   └── main.go
├── internal/
│   ├── handlers/        # HTTP request handlers
│   ├── models/          # Data models
│   └── youtube/         # YouTube API client & OAuth2
├── static/              # CSS, JS, icons
│   ├── icon.svg         # Application icon
│   ├── favicon.svg      # Browser tab icon
│   ├── style.css        # Styles
│   └── app.js           # Client-side JavaScript
├── templates/           # HTML templates
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── go.mod
```

## Development

```bash
# Build
make build

# Run tests
make test

# Lint
make lint

# Clean build artifacts
make clean
```

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `YOUTUBE_CLIENT_ID` | Yes | — | Google OAuth2 Client ID |
| `YOUTUBE_CLIENT_SECRET` | Yes | — | Google OAuth2 Client Secret |
| `YOUTUBE_REDIRECT_URL` | No | `http://localhost:8080/auth/callback` | OAuth2 redirect URI |
| `PORT` | No | `8080` | Server port |
| `TEMPLATE_DIR` | No | `templates` | Path to HTML templates |
| `STATIC_DIR` | No | `static` | Path to static assets |

## License

MIT