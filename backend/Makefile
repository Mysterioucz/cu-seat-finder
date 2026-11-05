-include .env

DATABASE_URL ?= postgres://seat:seat@localhost:5432/seatfinder
MONGO_URL ?= mongodb://localhost:27018
JWT_SECRET ?= dev-secret
USER_URL ?= http://localhost:8081
LIBRARY_GRPC_ADDR ?= localhost:9082
POLL_INTERVAL ?= 30
RABBITMQ_URL ?= amqp://seat:seat@localhost:5672/

export DATABASE_URL MONGO_URL JWT_SECRET USER_URL LIBRARY_GRPC_ADDR POLL_INTERVAL RABBITMQ_URL

.PHONY: db up down backup run test build proto loadtest stresstest nfr-security nfr-resilience demo-seed demo-ticket demo-noshow

db:
	cd .. && docker compose up -d

# Infra + gateway + all 5 services + web app in containers; the backend finds itself through Consul (ADR-009).
# App: http://localhost:3000. Gateway: http://localhost:8080. Registry UI: http://localhost:8500.
up:
	cd .. && docker compose --profile app up -d --build

down:
	cd .. && docker compose --profile app down

# Dumps Postgres + MongoDB into ../backups/<timestamp>/ (daily-backup NFR).
backup:
	./scripts/backup.sh

# NFR tests against the stack from `make up` (k6 runs from the grafana/k6 image
# if not installed). Results: ../loadtest/results/. UI checks: `pnpm test:nfr` in frontend/.
loadtest:
	../loadtest/run.sh load

stresstest:
	../loadtest/run.sh stress

nfr-security:
	../loadtest/run.sh security

nfr-resilience:
	../loadtest/resilience.sh

# Live-demo data (scripts/demo.sh): make demo-noshow REF=CU-XXXXXX
demo-seed:
	./scripts/demo.sh seed

demo-ticket:
	./scripts/demo.sh ticket

demo-noshow:
	./scripts/demo.sh noshow $(REF)

build:
	go build ./...

test:
	go test ./...

# Regenerates gen/ from ../proto (needs buf, protoc-gen-go, protoc-gen-go-grpc in PATH:
#   go install github.com/bufbuild/buf/cmd/buf@latest google.golang.org/protobuf/cmd/protoc-gen-go@latest google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest)
# gen/ is committed, so this is only needed after editing the .proto.
proto:
	buf generate ../proto --template buf.gen.yaml

# Runs the gateway (:8080) and all 5 services concurrently (library is gRPC on :9082);
# Ctrl-C kills the whole group. No CONSUL_ADDR here: plain localhost addresses, no registry.
run:
	@trap 'kill 0' EXIT; \
	PORT=8080 go run ./cmd/gateway & \
	PORT=8081 go run ./cmd/user & \
	GRPC_PORT=9082 go run ./cmd/library & \
	PORT=8083 go run ./cmd/occupancy & \
	PORT=8084 go run ./cmd/reservation & \
	PORT=8085 go run ./cmd/review & \
	wait
