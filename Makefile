.PHONY: help deps build run test test-cover lint fmt vet tidy clean \
       docker-build docker-run docker-up docker-up-d docker-down docker-logs docker-clean

APP_NAME   := youtube-cleaner
BUILD_DIR  := bin
COVER_FILE := coverage.out

# ─── Default ──────────────────────────────────────────────────────────────────

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# ─── Go ───────────────────────────────────────────────────────────────────────

deps: ## Download Go module dependencies
	go mod download

build: ## Compile the binary to bin/youtube-cleaner
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

run: build ## Build and run the server locally
	$(BUILD_DIR)/$(APP_NAME)

test: ## Run all tests with race detector
	go test ./... -v -race

test-cover: ## Run tests and generate coverage report
	go test ./... -v -race -coverprofile=$(COVER_FILE)
	go tool cover -func=$(COVER_FILE)

lint: vet fmt ## Run all linters (vet + fmt check)

vet: ## Run go vet static analysis
	go vet ./...

fmt: ## Check code formatting (fails if files need formatting)
	@test -z "$$(gofmt -l .)" || { echo "Files need formatting:"; gofmt -l .; exit 1; }

tidy: ## Tidy go.mod and go.sum
	go mod tidy

clean: ## Remove build artifacts and coverage files
	rm -rf $(BUILD_DIR) $(COVER_FILE)

# ─── Docker ───────────────────────────────────────────────────────────────────

docker-build: ## Build the Docker image
	docker build -t $(APP_NAME) .

docker-run: docker-build ## Build image and run container directly
	docker run --rm -p 8080:8080 \
		-e YOUTUBE_CLIENT_ID \
		-e YOUTUBE_CLIENT_SECRET \
		-e YOUTUBE_REDIRECT_URL \
		$(APP_NAME)

docker-up: ## Start services with Docker Compose (foreground)
	docker compose up --build

docker-up-d: ## Start services with Docker Compose (background)
	docker compose up --build -d

docker-down: ## Stop Docker Compose services
	docker compose down

docker-logs: ## Follow Docker Compose logs
	docker compose logs -f

docker-clean: docker-down ## Stop services and remove images/volumes
	docker compose down --rmi local --volumes
