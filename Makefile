all: build test

build:
	@echo "Building..."
	
	
	@go build -o main.exe cmd/main.go

up-dev:
	@echo "🔼 Docker Compose set up with dev environment..."
	docker compose -f docker-compose.dev.yml --env-file .env.dev up -d

down-dev:
	@echo "🔽 Docker Compose down for dev environment..."
	docker compose -f docker-compose.dev.yml --env-file .env.dev down

up-prod:
	@echo "🔼 Docker Compose set up with production environment..."
	docker compose -f docker-compose.prod.yml --env-file .env.prod up -d

run:
	@go run cmd/main.go

run-dev:
	@echo Running in development mode with Air...
	@set APP_ENV=development && air

run-prod:
	@echo Running in production mode...
	@set APP_ENV=production && go run cmd/main.go
	
run-test:
	@echo Running in test mode...
	@set APP_ENV=test && go run cmd/main.go

test:
	@echo "Running all tests..."
	@go test ./... -v

test-unit:
	@echo "Running unit tests..."
	@go test ./pkg/... ./internal/domain/userService/service/... ./internal/domain/authService/service/... -v

test-coverage:
	@echo "Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-short:
	@echo "Running short tests (excluding integration tests)..."
	@go test ./... -short -v

test-benchmark:
	@echo "Running benchmark tests..."
	@go test ./... -bench=. -benchmem

test-verbose:
	@echo "Running tests with verbose output..."
	@go test ./... -v -count=1

clean:
	@echo "Cleaning..."
	@rm -f main
	@rm -f coverage.out coverage.html

generate-key:
	@go run scripts/keys.go
	@echo "✓ Keys generated in /keys directory."

.PHONY: all build run test clean watch
