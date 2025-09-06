.PHONY: run up down logs test fmt lint tidy swagger

up:
	docker compose up --build -d

down:
	docker compose down -v

logs:
	docker compose logs -f api

run:
	go run ./cmd/subscriptions

fmt:
	go fmt ./...

tidy:
	go mod tidy

test:
	go test ./... -v
