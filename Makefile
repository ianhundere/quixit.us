.PHONY: run build test clean db-up db-down db-reset reset frontend backend dev install setup-dev all update-deps

# Default target
all: dev

# Development
dev: db-up
	@echo "Starting development environment..."
	@make -j2 frontend backend

frontend:
	@echo "Starting frontend..."
	cd frontend && npm run dev

backend:
	@echo "Starting backend..."
	go run ./backend/main.go

# Installation and setup
install:
	@echo "Installing dependencies..."
	cd frontend && npm install
	go mod download

update-deps:
	@echo "Updating dependencies..."
	cd frontend && rm -f package-lock.json && npm install
	go get -u ./...
	go mod tidy

setup-dev: install
	@echo "Setting up development environment..."
	mkdir -p storage
	make db-up
	@echo "Setting up test data..."
	@go run ./backend/testdata/cmd/setup_cmd.go

# Production build
build: build-frontend build-backend

build-frontend:
	@echo "Building frontend..."
	cd frontend && npm run build

build-backend:
	@echo "Building backend..."
	go build -o bin/server ./backend/main.go

# Database operations
db-up:
	@echo "Starting database..."
	docker-compose up -d
	@echo "Waiting for database to be ready..."
	@until docker-compose ps postgres | grep -q "healthy"; do sleep 1; done

db-down:
	@echo "Stopping database..."
	docker-compose down

db-reset:
	@echo "Resetting database..."
	docker-compose down -v
	docker-compose up -d
	@echo "Waiting for database to be ready..."
	@until docker-compose ps postgres | grep -q "healthy"; do sleep 1; done

# Cleanup
clean:
	@echo "Cleaning up..."
	rm -rf bin/
	rm -rf storage/
	rm -rf frontend/dist/
	rm -rf frontend/node_modules/

# Reset everything and start fresh
reset: clean db-reset setup-dev

# Run tests
test:
	@echo "Running backend tests..."
	go test -v ./...
	@echo "Running frontend tests..."
	cd frontend && npm test

# Help
help:
	@echo "Available commands:"
	@echo "  make dev          - Start development environment (frontend + backend + db)"
	@echo "  make frontend     - Start frontend development server only"
	@echo "  make backend      - Start backend server only"
	@echo "  make install      - Install all dependencies"
	@echo "  make setup-dev    - Set up development environment"
	@echo "  make build        - Build for production"
	@echo "  make clean        - Clean up build artifacts"
	@echo "  make reset        - Reset everything and start fresh"
	@echo "  make test         - Run all tests"
	@echo "  make db-up        - Start database"
	@echo "  make db-down      - Stop database"
	@echo "  make db-reset     - Reset database"
	@echo "  make update-deps  - Update all dependencies to their latest versions"
