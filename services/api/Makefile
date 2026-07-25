.PHONY: run test tidy migrate-up migrate-down docker-build

API_ADDR ?= :8080
MIGRATIONS_PATH ?= migrations
DATABASE_URL ?= postgres://donna:donna@localhost:5432/donna?sslmode=disable

run:
	API_ADDR=$(API_ADDR) go run ./cmd/api

test:
	go test ./...

tidy:
	go mod tidy

migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" down 1

docker-build:
	docker build -t donna-api .
