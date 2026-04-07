.PHONY: build run test clean docker-build docker-run lint

APP_NAME := youtube-cleaner
BUILD_DIR := bin

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

run: build
	$(BUILD_DIR)/$(APP_NAME)

test:
	go test ./... -v -race

lint:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)

docker-build:
	docker build -t $(APP_NAME) .

docker-run: docker-build
	docker run -p 8080:8080 \
		-e YOUTUBE_CLIENT_ID \
		-e YOUTUBE_CLIENT_SECRET \
		$(APP_NAME)

tidy:
	go mod tidy
