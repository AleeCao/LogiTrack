# Load environment variables from .env
include .env
export

# Variable for the DB Connection String
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable

.PHONY: migrate-up migrate-down

migrate-up:
	migrate -path pkg/db/migrations/postgres -database "$(DB_URL)" up

migrate-down:
	migrate -path pkg/db/migrations/postgres -database "$(DB_URL)" down
