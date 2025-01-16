.PHONY: run build test clean db-up db-down db-reset reset

run:
	go run ./backend/main.go

build:
	go build -o bin/server ./backend/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/
	rm -rf storage/

db-up:
	docker-compose up -d

db-down:
	docker-compose down

db-reset:
	docker-compose down -v
	docker-compose up -d
	sleep 3  # Wait for database to be ready

reset: db-reset setup

setup: db-up
	mkdir -p storage