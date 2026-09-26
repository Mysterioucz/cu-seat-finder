-include .env

DATABASE_URL ?= postgres://seat:seat@localhost:5432/seatfinder
JWT_SECRET ?= dev-secret
USER_URL ?= http://localhost:8081
LIBRARY_URL ?= http://localhost:8082
POLL_INTERVAL ?= 30

export DATABASE_URL JWT_SECRET USER_URL LIBRARY_URL POLL_INTERVAL

.PHONY: db run test build

db:
	cd .. && docker compose up -d

build:
	go build ./...

test:
	go test ./...

# Runs all 5 services concurrently; Ctrl-C kills the whole group.
run:
	@trap 'kill 0' EXIT; \
	PORT=8081 go run ./cmd/user & \
	PORT=8082 go run ./cmd/library & \
	PORT=8083 go run ./cmd/occupancy & \
	PORT=8084 go run ./cmd/reservation & \
	PORT=8085 go run ./cmd/review & \
	wait
