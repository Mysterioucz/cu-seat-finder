# CU SeatFinder — Review Service on MongoDB

This branch moves one business use case, **UC-04 Review Library**, from PostgreSQL
to MongoDB. All other services stay on PostgreSQL.

For the full backend (services, ports, gRPC), see [backend/README.md](backend/README.md).
For the API, see [.planning/API-CONTRACT.md](.planning/API-CONTRACT.md).

## Why reviews, and why not reservations

ADR-006 rejected MongoDB because it has no cross-document constraint to stop two
customers booking the same room and time slot. That argument applies to
**Reservation (UC-02)**, which needs row-level locking inside a transaction, so it
stays on PostgreSQL.

Reviews have no rule that spans several records:

- each review is self-contained (rating, tags, comment, moderation flag)
- `tags` is a variable-length list, stored as a native array instead of a JSON column
- the average rating is computed on read, so nothing else has to stay in sync
- every query is "reviews for one library" or "one review by id"

A document store fits this directly, and new tag types or fields need no schema
migration. The ADR text to update is in
[docs/ADR-006-mongodb-update.md](docs/ADR-006-mongodb-update.md).

## Data stores after the change

| Service     | Store      | Location                                   |
|-------------|------------|--------------------------------------------|
| user        | PostgreSQL | schema `users`                             |
| library     | PostgreSQL | schema `library`                           |
| occupancy   | PostgreSQL | schema `occupancy`                         |
| reservation | PostgreSQL | schema `reservation`                       |
| **review**  | **MongoDB**| database `seatfinder`, collection `reviews`|

Each service still owns its data, and other services reach reviews only through
the Review REST API. **The `/api/reviews` API is unchanged**, so the frontend and
the other services needed no changes.

### Review document

```js
{
  _id: "263168b894da24160eacfb54",   // same id format as before (idgen)
  library_id: "lib-central",
  user_id: "a3368c8656e0c1749ad53650",
  user_name: "Student 6631330021",
  rating: 4,                          // 1–5
  tags: ["quiet", "cold"],
  comment: "Good for focus",
  hidden: false,                      // set by admin moderation
  created_at: ISODate("2026-10-12T14:08:32.799Z")
}
```

Allowed tags: `quiet`, `noisy`, `cold`, `hot`, `strong_wifi`, `weak_wifi`,
`clean`, `many_outlets`, `crowded`.

### Indexes

The Review service creates these on startup (idempotent, no migration tool):

| Index                                   | Used by                                  |
|-----------------------------------------|------------------------------------------|
| `library_id: 1, hidden: 1, created_at: -1` | `GET /api/reviews?library_id=…` (public list) |
| `hidden: -1, created_at: -1`            | `GET /api/reviews/flagged` (admin queue) |

## What changed in the code

| File | Change |
|------|--------|
| [backend/cmd/review/handlers.go](backend/cmd/review/handlers.go) | SQL replaced with MongoDB driver calls (`Find`, `InsertOne`, `FindOneAndUpdate`, `UpdateOne`, `DeleteOne`) |
| [backend/cmd/review/store.go](backend/cmd/review/store.go) | `CREATE TABLE` replaced with index creation |
| [backend/cmd/review/main.go](backend/cmd/review/main.go) | Connects to MongoDB using `MONGO_URL` / `MONGO_DB` |
| [backend/internal/dbx/mongo.go](backend/internal/dbx/mongo.go) | New `ConnectMongo` helper |
| [docker-compose.yml](docker-compose.yml) | New `mongo:8` container |
| [backend/Makefile](backend/Makefile), [backend/.env.example](backend/.env.example) | New `MONGO_URL` and `MONGO_DB` settings |

Driver: `go.mongodb.org/mongo-driver/v2`.

## Run it

Requires Docker, Go and pnpm.

```bash
# Terminal 1: databases + backend
cd backend
make db      # starts PostgreSQL and MongoDB in Docker
make run     # starts all 5 services

# Terminal 2: frontend
cd frontend
pnpm install
pnpm dev     # http://localhost:3000
```

### Ports on your machine

| Container  | Host port | Connection string |
|------------|-----------|-------------------|
| PostgreSQL | **5433**  | `postgres://seat:seat@localhost:5433/seatfinder` |
| MongoDB    | **27018** | `mongodb://localhost:27018` (db `seatfinder`) |

These are not the default ports (5432 / 27017), so the containers don't clash
with a PostgreSQL or MongoDB installed locally (for example with Homebrew).

**If you already had the project set up:** pull, run `make db` again to create the
Mongo container, and update `DATABASE_URL` and `MONGO_URL` in your own
`backend/.env` to the ports above.

### Sign in

Open http://localhost:3000/login.

- **Student:** "CU Account" tab, any student ID (no password, mock CU SSO)
- **Admin:** "Admin" tab, `admin@cu.ac.th` / `admin123`

## Check the data in MongoDB

Run from the project root:

```bash
# list all reviews
docker compose exec mongo mongosh seatfinder --eval 'db.reviews.find().pretty()'

# interactive shell
docker compose exec mongo mongosh seatfinder
```

Useful queries in the shell:

```js
db.reviews.countDocuments()
db.reviews.find({ library_id: "lib-central" })
db.reviews.find({ hidden: true })          // hidden by admin
db.reviews.getIndexes()
```

GUI: connect MongoDB Compass to `mongodb://localhost:27018`, then open
`seatfinder` → `reviews`.

To confirm reviews are no longer in PostgreSQL:

```bash
docker compose exec postgres psql -U seat -d seatfinder -c '\dn'
# lists library, occupancy, reservation, users — no review schema
```

## Trade-offs

- The team now runs two database engines, and must know MongoDB queries as well as SQL.
- Review data can't be joined with SQL. This was already ruled out by the
  one-schema-per-service rule, so nothing that worked before is lost.
- Existing reviews in PostgreSQL are not migrated. A fresh setup starts with no reviews.
- Production MongoDB hosting is not decided yet. Local development uses Docker.
