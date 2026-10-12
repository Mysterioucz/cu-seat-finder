# CU SeatFinder backend

Five small Go services. User, Library, Occupancy and Reservation share one
Postgres database (one schema per service); **Review stores its data in MongoDB**
(collection `seatfinder.reviews`). No
gateway — the Next.js frontend proxies REST `/api/*` to the right port via
`next.config.ts` rewrites. The **Library service is gRPC-only**
(`proto/library/v1/library.proto`, full CRUD); Occupancy, Reservation and Review
call it over gRPC, and the Next.js server translates REST to gRPC for the browser.
See `.planning/API-CONTRACT.md` for the full API.

## Run it

```
make db     # starts Postgres + MongoDB in Docker (repo-root docker-compose.yml)
make run    # starts all 5 services, Ctrl-C stops all of them
make test   # go test ./... (pure-logic tests, no DB needed)
make build  # go build ./...
make proto  # regenerate gen/ from ../proto (needs buf + protoc-gen-go + protoc-gen-go-grpc; gen/ is committed)
```

Copy `.env.example` to `.env` to override defaults (DB URL, Mongo URL, JWT secret, ports).

## Ports

| Service     | Port | Prefix                        |
|-------------|------|--------------------------------|
| user        | 8081 | /api/auth, /api/users          |
| library     | 9082 | gRPC `library.v1.LibraryService` (`GRPC_PORT`) |
| occupancy   | 8083 | /api/occupancy                 |
| reservation | 8084 | /api/reservations              |
| review      | 8085 | /api/reviews                   |

Each Postgres service creates its own schema/tables and seeds data on startup
(`CREATE SCHEMA/TABLE IF NOT EXISTS`, idempotent seed inserts) — no separate
migration step for dev. See the `ponytail:` comments in each `store.go` for
the upgrade path (golang-migrate) once schemas stop being this simple.

Library address for clients: `LIBRARY_GRPC_ADDR` (default `localhost:9082`).

## Library gRPC demo (grpcurl)

Server reflection is on, so no `-proto` flag is needed
(`go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest`).
Reads are open; Create/Update/Delete need an admin JWT as `authorization` metadata.

```
TOKEN=$(curl -s localhost:8081/api/auth/admin/login -H 'content-type: application/json' \
  -d '{"email":"admin@cu.ac.th","password":"admin123"}' | jq -r .token)
H="authorization: Bearer $TOKEN"
G="grpcurl -plaintext"

$G localhost:9082 list                                   # services (reflection)
$G -d '{"q":"eng"}' localhost:9082 library.v1.LibraryService/ListLibraries
$G -d '{"id":"lib-central"}' localhost:9082 library.v1.LibraryService/GetLibrary
$G -H "$H" -d '{"library":{"name_th":"ทดสอบ","name_en":"Demo","capacity":50,"amenities":["quiet"],
  "floors":[{"name":"F1","zones":[{"name":"Z1","capacity":50}]}]}}' localhost:9082 library.v1.LibraryService/CreateLibrary
$G -H "$H" -d '{"library":{"id":"<id>","name_th":"ทดสอบ","name_en":"Demo 2","capacity":60}}' \
  localhost:9082 library.v1.LibraryService/UpdateLibrary   # floors/zones replaced wholesale
$G -H "$H" -d '{"id":"<id>"}' localhost:9082 library.v1.LibraryService/DeleteLibrary  # FAILED_PRECONDITION if it has rooms

$G -d '{"library_id":"lib-central"}' localhost:9082 library.v1.LibraryService/ListRooms   # empty = all rooms
$G -d '{"id":"room-central-1"}' localhost:9082 library.v1.LibraryService/GetRoom
$G -H "$H" -d '{"room":{"library_id":"lib-central","name":"Room X","capacity":4}}' localhost:9082 library.v1.LibraryService/CreateRoom
$G -H "$H" -d '{"room":{"id":"<id>","library_id":"lib-central","name":"Room Y","capacity":6}}' localhost:9082 library.v1.LibraryService/UpdateRoom
$G -H "$H" -d '{"id":"<id>"}' localhost:9082 library.v1.LibraryService/DeleteRoom
```

Writes without a token return `Unauthenticated`; with a customer token, `PermissionDenied`.

## Review CRUD (REST, :8085)

`GET /api/reviews/{id}`, `PUT /api/reviews/{id}` (author only), `DELETE /api/reviews/{id}`
(author or admin, 204), alongside the existing list/create/moderate routes. The
average is computed on read, so edits and deletes are reflected immediately.

## Why Review uses MongoDB

ADR-006 rejected MongoDB because it cannot enforce the no-double-booking rule
across documents — that argument applies to **Reservation**, which stays on
Postgres. Reviews have no cross-document invariant: each review is a
self-contained document with a variable `tags` array, the average rating is
computed on read, and all access is by `library_id` or `_id`. So the Review
service now owns a MongoDB collection instead of the `review` Postgres schema.
Indexes (`library_id+hidden+created_at`, `hidden+created_at`) are created on
boot. The REST contract is unchanged.

Inspect data: `docker compose exec mongo mongosh seatfinder --eval 'db.reviews.find()'`

## Seed accounts

- Admin: `admin@cu.ac.th` / `admin123` (`POST /api/auth/admin/login`)
- Customer: any `student_id` works via `POST /api/auth/login` — first login
  creates the account (mock OIDC, ADR-004).

## Libraries seeded

Central, Engineering, Science, Arts, Architecture, Law — each with floors/zones
and 1-3 meeting rooms.
