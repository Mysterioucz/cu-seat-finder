# API Contract (shared by frontend + backend)

Source of truth for the HTTP interface. Derived from project-proposal.md + microservice-diagram.md.

## Topology (dev)

| Service     | Port | Owns path prefixes                         | DB schema     |
|-------------|------|--------------------------------------------|---------------|
| user        | 8081 | /api/auth, /api/users                      | users         |
| library     | 9082 (gRPC) | — gRPC only; BFF exposes /api/libraries, /api/rooms | library |
| occupancy   | 8083 | /api/occupancy                             | occupancy     |
| reservation | 8084 | /api/reservations                          | reservation   |
| review      | 8085 | /api/reviews                               | review        |

- No separate Go gateway: Next.js is the BFF. `rewrites` in `frontend/next.config.ts` forward REST prefixes to their service; for the gRPC-only Library service, Next.js **route handlers** under `src/app/api/libraries` and `src/app/api/rooms` translate REST↔gRPC (browser only ever talks to :3000).
- Transport (ADR-007): browser→BFF is REST; **Library Service is gRPC** (`proto/library/v1/library.proto`) and every service-to-Library call (Occupancy, Reservation, Review, Next.js server) uses gRPC. Other service-to-service calls (→ User) stay REST.
- One local Postgres (docker compose, `postgres://seat:seat@localhost:5432/seatfinder`), one schema per service — except Review, which owns a MongoDB collection (`mongodb://localhost:27018`, db `seatfinder`, collection `reviews`). Services never read another service's schema; they call each other over HTTP (collaborations below).
- Auth: JWT HS256, secret `JWT_SECRET` (dev default `dev-secret`). Header `Authorization: Bearer <token>`. Claims: `sub` (user id), `role` (`customer`|`admin`), `name`, `exp`.
- JSON everywhere, snake_case fields, timestamps RFC3339. Errors: `{"error": "<code>", "message": "<human text>"}` with proper status.

## User service (8081)
- `POST /api/auth/login` `{student_id, name?}` → mock OIDC login (ADR-004): upserts customer, returns `{token, user}`. Real CU SSO swaps in behind the same `IdentityProvider` interface later.
- `POST /api/auth/admin/login` `{email, password}` → `{token, user}` (bcrypt; seed `admin@cu.ac.th` / `admin123`).
- `POST /api/auth/logout` → 204 (stateless; client drops token).
- `GET /api/users/me`, `PATCH /api/users/me` `{name?, language?}` (auth).
- `GET /api/users/{id}` → `{id, name, role, student_id, active_penalty: bool}` (internal viewProfile). `active_penalty` = the user has an unexpired row in `users.penalties`.
- `GET /api/users/penalized` (admin) → `[{id, name, student_id, expires_at}]` users with an active penalty. `POST /api/users/{id}/unban` (admin) → 204; expires their active penalties now (404 `not_found` if none).
- Consumes `reservation.no_show` (RabbitMQ, see Events) → inserts a 7-day penalty keyed by `reservation_id`, so redelivery is a no-op.
- User: `{id, student_id?, email?, name, role, language: "th"|"en", created_at}`.

## Library service (gRPC :9082) — full CRUD
- Contract: `proto/library/v1/library.proto` (binding). List/Get open; Create/Update/Delete need `authorization: Bearer <JWT>` metadata with role=admin.
- Seed ~6 real Chula libraries with 2-4 rooms each (unchanged data).
- BFF REST facade (Next.js route handlers, same JSON shapes as before so existing pages keep working; gRPC status → HTTP: NOT_FOUND 404, INVALID_ARGUMENT 400, UNAUTHENTICATED 401, PERMISSION_DENIED 403, FAILED_PRECONDITION 409, UNAVAILABLE 503):
  - `GET /api/libraries?q=&amenity=` → `[Library]`; `GET /api/libraries/{id}` → `Library`; `GET /api/libraries/{id}/rooms` → `[Room]`; `GET /api/rooms/{id}` → `Room`.
  - admin: `POST /api/libraries`, `PUT /api/libraries/{id}`, `DELETE /api/libraries/{id}` (204), `POST /api/rooms`, `PUT /api/rooms/{id}`, `DELETE /api/rooms/{id}` (204). Forward the caller's bearer token as gRPC metadata.
- Library JSON: `{id, name_th, name_en, capacity, amenities, floors: [{id, name, zones: [{id, name, capacity}]}], image_url?}`; Room: `{id, library_id, name, capacity, amenities}`.

## Occupancy service (8083)
- Polls turnstile source every `POLL_INTERVAL` (default 30s) with `time.Ticker` (ADR-005). Dev source = in-process mock turnstile generating tap-in/out events; behind a `TurnstileSource` interface.
- Validates library exists via Library `GET /api/libraries/{id}` before publishing.
- `GET /api/occupancy?min=&max=` → `{items: [Occupancy], suggestion: Occupancy|null}`; items sorted least→most crowded; when filter matches nothing, `suggestion` = least crowded overall (UC-03).
- `GET /api/occupancy/{library_id}` → `Occupancy`.
- Occupancy: `{library_id, current, capacity, percent (0-100 int), updated_at, stale: bool}`; `stale` = updated_at older than 3×poll interval (UI shows "Last updated X mins ago").

## Reservation service (8084)
- `POST /api/reservations` `{room_id, start_time, duration_minutes}` (auth customer) → 201 `Reservation` status `pending`.
  - Quota: max 2 active (pending/accepted) bookings, max 180 booked minutes per day, blocked if `active_penalty` (User service). Errors: 422 `quota_exceeded`, 403 `penalty`.
  - Conflict: overlapping pending/accepted booking on same room → 409 `conflict`. Enforced in a transaction with row lock (`SELECT ... FOR UPDATE` on a room lock row) + exclusion constraint (ADR-006).
  - Validates room via Library `GET /api/rooms/{id}`.
- `GET /api/reservations/me` (auth) → `[Reservation]`.
- `GET /api/reservations?status=pending` (admin) → `[Reservation]`.
- `POST /api/reservations/{id}/accept` (admin) → sets `accepted`, generates `reference` (e.g. `CU-7F3K9Q`) and `checkin_deadline` = start_time + 15min.
- `POST /api/reservations/{id}/reject` (admin), `POST /api/reservations/{id}/cancel` (owner).
- `POST /api/reservations/checkin` `{reference}` (admin; staff scan the student's QR) → sets `checked_in_at`. Window: start_time − 15min … `checkin_deadline`. Any reason it can't check in (unknown reference, not accepted, already checked in, outside the window) → 409 `not_checkable`.
- No-show sweeper: once a minute, accepted bookings past `checkin_deadline` with no check-in are published as `reservation.no_show`, then set to `no_show` (publish first, then update; see Events).
- Reservation: `{id, user_id, room_id, start_time, end_time, status: pending|accepted|rejected|cancelled|no_show, reference?, checkin_deadline?, checked_in_at?, created_at}`. Frontend renders QR from `reference` + countdown to `checkin_deadline`.

## Review service (8085)
- `GET /api/reviews?library_id=` → `{average: number, count: int, items: [Review]}` (hidden reviews excluded unless admin).
- `POST /api/reviews` `{library_id, rating 1-5, tags: [string], comment?}` (auth) → 201. Profanity filter (Thai + English wordlist) → 422 `profanity` so UI asks user to revise. Validates library via Library service; reviewer name via User service.
- `PUT /api/reviews/{id}` `{rating, tags, comment?}` (owner only; same validation + profanity 422) → `Review`.
- `DELETE /api/reviews/{id}` (owner or admin) → 204; 403 `forbidden` otherwise, 404 if missing.
- `GET /api/reviews/{id}` → `Review` (404 if missing or hidden and caller not admin).
- `GET /api/reviews/flagged` (admin, all reviews, hidden first — moderation queue), `PATCH /api/reviews/{id}/moderate` `{hidden: bool}` (admin).
- Review: `{id, library_id, user_id, user_name, rating, tags, comment, hidden, created_at}`. Tags stored as JSONB.
- Allowed tags: `quiet, noisy, cold, hot, strong_wifi, weak_wifi, clean, many_outlets, crowded`.

## Events (RabbitMQ, ADR-007)
- Broker: `RABBITMQ_URL` (default `amqp://seat:seat@localhost:5672/`), started by `docker compose`. Management UI on :15672.
- Topology (declared by both publisher and consumer in `backend/internal/broker`): topic exchange `reservation.events`; queue `user.reservation-no-show` bound to `reservation.no_show`; rejected messages dead-letter via the default exchange to queue `reservation.events.dlq`. Consumer prefetch is 1.
- `reservation.no_show` (Reservation → User): `{reservation_id, user_id, room_id, start_time, checkin_deadline}`, persistent, `message_id = "no_show:<reservation_id>"`.
- Delivery is at-least-once: the publisher waits for the broker confirm before marking the row `no_show`, and the consumer dedupes on `reservation_id`. Malformed events or unknown users go to the DLQ; transient failures (DB down) are requeued.
