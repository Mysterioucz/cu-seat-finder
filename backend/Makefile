-include .env

DATABASE_URL ?= postgres://seat:seat@localhost:5432/seatfinder
JWT_SECRET ?= dev-secret
USER_URL ?= http://localhost:8081
LIBRARY_GRPC_ADDR ?= localhost:9082
POLL_INTERVAL ?= 30
RABBITMQ_URL ?= amqp://seat:seat@localhost:5672/

export DATABASE_URL JWT_SECRET USER_URL LIBRARY_GRPC_ADDR POLL_INTERVAL RABBITMQ_URL

.PHONY: db run test build proto

db:
	cd .. && docker compose up -d

build:
	go build ./...

test:
	go test ./...

# Regenerates gen/ from ../proto (needs buf, protoc-gen-go, protoc-gen-go-grpc in PATH:
#   go install github.com/bufbuild/buf/cmd/buf@latest google.golang.org/protobuf/cmd/protoc-gen-go@latest google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest)
# gen/ is committed, so this is only needed after editing the .proto.
proto:
	buf generate ../proto --template buf.gen.yaml

# Runs all 5 services concurrently (library is gRPC on :9082); Ctrl-C kills the whole group.
run:
	@trap 'kill 0' EXIT; \
	PORT=8081 go run ./cmd/user & \
	GRPC_PORT=9082 go run ./cmd/library & \
	PORT=8083 go run ./cmd/occupancy & \
	PORT=8084 go run ./cmd/reservation & \
	PORT=8085 go run ./cmd/review & \
	wait
