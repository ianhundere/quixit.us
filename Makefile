.PHONY: all dev frontend backend install setup-dev build build-frontend build-backend docker-build docker-dev db-up db-down db-reset clean reset test help

# Default goal
.DEFAULT_GOAL := dev

# Main targets
all: dev

# Development targets
dev: db-up
	@echo "starting development environment..."
	@source .env.dev && make -j2 frontend backend

frontend:
	cd frontend && npm run dev

backend:
	go run ./backend/main.go

# Docker targets
docker-build: build-frontend
	@echo "building docker image..."
	@source .env.docker && docker-compose build

docker-dev: db-up
	@echo "starting docker development environment..."
	@echo "building docker image (which now includes frontend build)..."
	@source .env.docker && docker-compose build
	@source .env.docker && docker-compose up -d quixit
	@echo "docker development environment is running"
	@echo "application is available at: http://$${HOST_DOMAIN}:3000"

# Installation and setup
install:
	@echo "installing dependencies..."
	cd frontend && npm install
	go mod download

setup-dev: install
	@echo "setting up development environment..."
	mkdir -p storage
	make db-up
	go run ./backend/testdata/cmd/setup_cmd.go

# Build targets
build: build-frontend build-backend

build-frontend:
	@echo "building frontend..."
	cd frontend && npm run build-no-types

build-backend:
	@echo "building backend..."
	go build -o bin/server ./backend/main.go

# Database operations
db-up:
	@echo "starting database..."
	@source .env.docker && docker-compose up -d postgres
	@until docker-compose ps postgres | grep -q "healthy"; do sleep 1; done
	@echo "database is ready"

db-down:
	@echo "stopping database..."
	docker-compose down

db-reset:
	@echo "resetting database..."
	docker-compose down -v
	@source .env.docker && docker-compose up -d postgres
	@until docker-compose ps postgres | grep -q "healthy"; do sleep 1; done
	@echo "database has been reset"

# Cleanup
clean:
	@echo "cleaning up..."
	rm -rf bin/
	rm -rf storage/
	rm -rf frontend/dist/
	rm -rf frontend/node_modules/

# Reset everything and start fresh
reset: clean db-reset setup-dev

# Run tests
test:
	@echo "running tests..."
	go test -v ./...
	cd frontend && npm test

# Help
help:
	@echo "quixit development commands:"
	@echo "  make              - default target, same as 'make dev'"
	@echo "  make dev          - start local development (backend, frontend, db)"
	@echo "  make frontend     - start only the frontend dev server"
	@echo "  make backend      - start only the backend server"
	@echo "  make docker-build - build the frontend and create a docker image"
	@echo "  make docker-dev   - run the application in docker containers"
	@echo "  make install      - install all dependencies"
	@echo "  make setup-dev    - set up the development environment"
	@echo "  make build        - build the application for production"
	@echo "  make test         - run all tests"
	@echo "  make clean        - clean all build artifacts"
	@echo "  make db-up        - start the database"
	@echo "  make db-down      - stop the database"
	@echo "  make db-reset     - reset the database"
	@echo "  make reset        - clean, reset db, and set up dev environment"
