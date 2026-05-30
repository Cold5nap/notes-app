.PHONY: run build test migrate-up migrate-down

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

test:
	go test ./...

migrate-up:
	migrate -path ./migrations -database "postgres://notes:notes@localhost:5432/notes_app?sslmode=disable" up

migrate-down:
	migrate -path ./migrations -database "postgres://notes:notes@localhost:5432/notes_app?sslmode=disable" down 1
