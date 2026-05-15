ifneq (,$(wildcard .env))
	include .env
	export
endif

.PHONY: run dev build test lint migrate-up migrate-down sqlc docker-up docker-down

run:
	go run ./cmd/api

dev:
	air

build:
	go build -o bin/api ./cmd/api

test:
	go test ./...

lint:
	golangci-lint run

migrate-up:
	goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DATABASE_URL)" down

sqlc:
	sqlc generate

docker-up:
	docker compose up -d

docker-down:
	docker compose down
