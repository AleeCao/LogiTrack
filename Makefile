# Load environment variables from .env
include .env
export

# Variable for the DB Connection
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable

.PHONY: migrate-up migrate-down
migrate-up:
	migrate -path deploy/postgres -database "$(DB_URL)" up

migrate-down:
	migrate -path deploy/postgres -database "$(DB_URL)" down


.PHONY: proto
proto:
	@echo "Generating Go code..."
	protoc --go_out=module=github.com/AleeCao/LogiTrack:. \
       --go-grpc_out=module=github.com/AleeCao/LogiTrack:. \
       --proto_path=proto \
       proto/tracking/v1/tracking.proto


