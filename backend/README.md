# CU SeatFinder backend

Five small Go services, one Postgres database (one schema per service). No
gateway — the Next.js frontend proxies `/api/*` to the right port via
`next.config.ts` rewrites. See `.planning/API-CONTRACT.md` for the full API.

## Run it

```
make db     # starts Postgres in Docker (repo-root docker-compose.yml)
make run    # starts all 5 services, Ctrl-C stops all of them
make test   # go test ./... (pure-logic tests, no DB needed)
make build  # go build ./...
```

Copy `.env.example` to `.env` to override defaults (DB URL, JWT secret, ports).

## Ports

| Service     | Port | Prefix                        |
|-------------|------|--------------------------------|
| user        | 8081 | /api/auth, /api/users          |
| library     | 8082 | /api/libraries, /api/rooms     |
| occupancy   | 8083 | /api/occupancy                 |
| reservation | 8084 | /api/reservations              |
| review      | 8085 | /api/reviews                   |

Each service creates its own schema/tables and seeds data on startup
(`CREATE SCHEMA/TABLE IF NOT EXISTS`, idempotent seed inserts) — no separate
migration step for dev. See the `ponytail:` comments in each `store.go` for
the upgrade path (golang-migrate) once schemas stop being this simple.

## Seed accounts

- Admin: `admin@cu.ac.th` / `admin123` (`POST /api/auth/admin/login`)
- Customer: any `student_id` works via `POST /api/auth/login` — first login
  creates the account (mock OIDC, ADR-004).

## Libraries seeded

Central, Engineering, Science, Arts, Architecture, Law — each with floors/zones
and 1-3 meeting rooms.
