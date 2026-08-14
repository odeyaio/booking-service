-include .env
-include .env_example
export

MIGRATE_STEPS ?= 1

.PHONY: up down migrate-up migrate-down test coverage

up:
	docker compose up --build

down:
	docker compose down

migrate-up:
	docker compose run --rm migrate

migrate-down:
	docker compose run --rm migrate down $(MIGRATE_STEPS)

test:
	go test ./...

coverage:
	go test ./... -count=1 -covermode=atomic -coverprofile=coverage.out
