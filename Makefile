.PHONY: build test clean docker docker-compose-up docker-compose-down install dev

# Build targets
build:
	go build -o bin/subconverter ./cmd/subconverter
	go build -o bin/subctl ./cmd/subctl
	go build -o bin/subworker ./cmd/subworker

# Development build
dev:
	go build -race -o bin/subconverter ./cmd/subconverter
	go build -race -o bin/subctl ./cmd/subctl
	go build -race -o bin/subworker ./cmd/subworker

# Test targets
test:
	go test -v ./...

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Lint targets
lint:
	golangci-lint run

# Clean targets
clean:
	rm -rf bin/
	rm -rf coverage.out coverage.html

# Docker targets
docker:
	docker build -t subconverter-go .

docker-compose-up:
	docker-compose up -d

docker-compose-down:
	docker-compose down

# Install targets
install:
	go install ./cmd/subconverter
	go install ./cmd/subctl
	go install ./cmd/subworker

# Development targets
dev-server:
	go run ./cmd/subconverter

dev-cli:
	go run ./cmd/subctl

dev-worker:
	go run ./cmd/subworker

# Format code
fmt:
	go fmt ./...

# Tidy modules
tidy:
	go mod tidy
	go mod verify

# Security scan
security:
	gosec ./...

# Benchmark
benchmark:
	go test -bench=. -benchmem ./...

# Generate mocks
generate:
	go generate ./...

# All-in-one development setup
dev-setup: tidy fmt lint test
	@echo "Development setup complete"

# Release build
release:
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/subconverter-linux ./cmd/subconverter
	CGO_ENABLED=0 GOOS=darwin go build -ldflags="-w -s" -o bin/subconverter-darwin ./cmd/subconverter
	CGO_ENABLED=0 GOOS=windows go build -ldflags="-w -s" -o bin/subconverter-windows.exe ./cmd/subconverter

	CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/subctl-linux ./cmd/subctl
	CGO_ENABLED=0 GOOS=darwin go build -ldflags="-w -s" -o bin/subctl-darwin ./cmd/subctl
	CGO_ENABLED=0 GOOS=windows go build -ldflags="-w -s" -o bin/subctl-windows.exe ./cmd/subctl

	CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/subworker-linux ./cmd/subworker
	CGO_ENABLED=0 GOOS=darwin go build -ldflags="-w -s" -o bin/subworker-darwin ./cmd/subworker
	CGO_ENABLED=0 GOOS=windows go build -ldflags="-w -s" -o bin/subworker-windows.exe ./cmd/subworker