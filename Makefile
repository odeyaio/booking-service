-include .env
-include .env_example
export

.PHONY: up down test coverage

up:
	docker compose up --build

down:
	docker compose down

test:
	go test ./...

coverage:
	go test ./... -count=1 -covermode=atomic -coverprofile=coverage.out
