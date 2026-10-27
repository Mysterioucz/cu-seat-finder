# CU SeatFinder

Review service (UC-04) stores its data in **MongoDB**; every other service uses
PostgreSQL. Setup, ports and API details: [backend/README.md](backend/README.md).

## Why only reviews

ADR-006 rejected MongoDB because it can't prevent two customers booking the same
room and slot. That applies to **Reservation**, which stays on PostgreSQL. Reviews
have no rule spanning several records: each is a self-contained document with a
variable `tags` array, and the average rating is computed on read. ADR text:
[docs/ADR-006-mongodb-update.md](docs/ADR-006-mongodb-update.md).

The `/api/reviews` API is unchanged, so the frontend needed no changes.

## Run

```bash
cd backend && make up                 # everything in Docker: gateway :8080, Consul UI :8500
# or: cd backend && make db && make run   (services on the host, no Consul)
cd frontend && pnpm install && pnpm dev
```

The frontend talks only to the API Gateway (`GATEWAY_URL`, default
`http://localhost:8080`), which routes to the services it finds in Consul
(ADR-008, ADR-009 in [docs/Proposal.md](docs/Proposal.md)).

MongoDB is published on **27018** so it doesn't clash with a local `mongod`. If a
local PostgreSQL already owns 5432, set `DATABASE_URL` in `backend/.env` and remap
the port in a gitignored `docker-compose.override.yml`.

## Check the data

```bash
docker compose exec mongo mongosh seatfinder --eval 'db.reviews.find().pretty()'
```

GUI: MongoDB Compass → `mongodb://localhost:27018` → `seatfinder.reviews`.
