# ADR Update — MongoDB for the Review Service

Changes to make in *Updated Architecture Decision Records & Microservice Diagram* (ADR_MD_V1).

---

## 1. Replace ADR-006 with the revised version below

### ADR-006: Database Selection (revised)

**Context**

The system stores three kinds of data: reservations, which need transactional writes so two customers cannot be granted the same room and time slot; reviews, which carry a variable list of tags per review; and occupancy logs from turnstile polling (ADR-005), which are append-heavy time-series rows read for recent trends.

Options: MongoDB, MySQL, PostgreSQL. MongoDB has no cross-document constraint that prevents a duplicate booking, so for reservations that guarantee would have to be enforced in application code, which is the wrong place for it. MySQL and PostgreSQL both provide transactional writes; PostgreSQL was preferred for richer indexing on timestamp ranges.

The argument against MongoDB applies only to data with a rule that spans several records. Reviews have no such rule. Each review is a self-contained record (rating, tags, comment, moderation flag); the average rating is computed when reviews are read, not stored elsewhere; and every query is either "all reviews for one library" or "one review by id". A document store models this directly: tags are an embedded array instead of a JSON column, and the review schema can change (new tag types, new fields) without a migration.

For hosting, options were Prisma Postgres, Supabase, and Neon (all managed PostgreSQL). Prisma Postgres was chosen for its free tier and minimal setup. Prisma ORM, the tool usually paired with Prisma Postgres, is Node/TypeScript-only and has no supported Go client, so the Go backend (ADR-003) cannot use it.

**Decision**

Use two databases, chosen per service:

- **PostgreSQL** (hosted on Prisma Postgres, queried with pgx) for the User, Library, Occupancy and Reservation services, one schema per service.
- **MongoDB** for the Review service, queried with the official Go driver (`go.mongodb.org/mongo-driver/v2`). Reviews are stored as documents in the `reviews` collection of the `seatfinder` database.

**Status**

Accepted (amended: Review service moved from PostgreSQL to MongoDB)

**Consequences**

- Row-level locking inside a transaction still enforces the no-double-booking rule for reservations in UC-02; Reservation stays on PostgreSQL for this reason.
- Each review is one document with its tags embedded as an array, so adding a new tag type or field needs no schema migration.
- Review indexes (`library_id + hidden + created_at` and `hidden + created_at`) are created by the Review service on startup. No migration tool is needed on the MongoDB side.
- The team now runs and backs up two database engines instead of one, and must learn MongoDB queries alongside SQL.
- Review data can no longer be joined with SQL against other tables. This was already forbidden by the one-schema-per-service rule, so nothing that worked before is lost.
- The Review REST API (`/api/reviews`) is unchanged, so the frontend and other services are not affected.
- No Prisma ORM types or migration CLI on the Go side; Postgres migrations need a separate tool (golang-migrate, or Prisma's CLI as a dev-only step).
- Dependent on Prisma Postgres's managed hosting availability and free-tier limits. MongoDB hosting for production is not yet decided (local development runs `mongo:8` in Docker Compose).
- Occupancy logs share the relational database; heavy growth would call for extra indexing or partitioning.

---

## 2. Microservice architecture: Step 3 table

In the **Review Service** row, add to the *Collaboration* column:

> Data store: MongoDB `reviews` collection (ADR-006, revised). The service owns this collection; no other service reads it.

In the **Reservation Service** row, the existing text "(ADR-006 row-level locking)" stays as it is.

---

## 3. Architecture diagram (Current Proposed Architecture)

- Add a second database cylinder: **MongoDB**, labelled "reviews collection, ADR-006".
- Connect **Review Service → MongoDB** with an arrow labelled "MongoDB driver".
- Remove `review` from the schema list on the **PostgreSQL / Prisma Postgres** cylinder, leaving `users, library, occupancy, reservation`.
- Remove the arrow from **Review Service** to **PostgreSQL**, if the diagram has one.
