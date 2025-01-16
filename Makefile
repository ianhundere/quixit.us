.PHONY: run build test clean db-up db-down

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

setup: db-up
	mkdir -p storage 