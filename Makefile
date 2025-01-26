.PHONY: build build-frontend build-backend test clean db-up db-down db-reset reset frontend backend dev install setup-dev

# This is the default target that runs when you just type 'make'
# It explicitly sets 'dev' as the default goal
.DEFAULT_GOAL := dev

# Development
dev: db-up
	@make -j2 frontend backend

frontend:
	cd frontend && npm run dev

backend:
	go run ./backend/main.go

# Installation and setup
install:
	cd frontend && npm install
	go mod download

setup-dev: install
	mkdir -p storage
	make db-up
	go run ./backend/testdata/cmd/setup_cmd.go

# Production build
build: build-frontend build-backend

build-frontend:
	cd frontend && npm run build

build-backend:
	go build -o bin/server ./backend/main.go

# Database operations
db-up:
	docker-compose up -d
	@until docker-compose ps postgres | grep -q "healthy"; do sleep 1; done

db-down:
	docker-compose down

db-reset:
	docker-compose down -v
	docker-compose up -d
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
	go test -v ./...
	cd frontend && npm test

# Help
help:
	@echo "Usage:"
	@echo "  dev         - Start all components (frontend, backend, db)"
	@echo "  frontend    - Start frontend only"
	@echo "  backend     - Start backend only"
	@echo "  install     - Install dependencies"
	@echo "  setup-dev   - Setup dev environment"
	@echo "  build       - Production build"
	@echo "  test        - Run tests"
	@echo "  clean       - Cleanup"
	@echo "  db-up       - Start database"
	@echo "  db-down     - Stop database"
	@echo "  db-reset    - Reset database"
