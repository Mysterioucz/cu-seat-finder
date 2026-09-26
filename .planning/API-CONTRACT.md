# API Contract (shared by frontend + backend)

Source of truth for the HTTP interface. Derived from project-proposal.md + microservice-diagram.md.

## Topology (dev)

| Service     | Port | Owns path prefixes                         | DB schema     |
|-------------|------|--------------------------------------------|---------------|
| user        | 8081 | /api/auth, /api/users                      | users         |
| library     | 8082 | /api/libraries, /api/rooms                 | library       |
| occupancy   | 8083 | /api/occupancy                             | occupancy     |
| reservation | 8084 | /api/reservations                          | reservation   |
| review      | 8085 | /api/reviews                               | review        |

- No separate Go gateway: Next.js `rewrites` in `frontend/next.config.ts` route each prefix to its service (browser only ever talks to :3000).
- One local Postgres (docker compose, `postgres://seat:seat@localhost:5432/seatfinder`), one schema per service. Services never read another service's schema; they call each other over HTTP (collaborations below).
- Auth: JWT HS256, secret `JWT_SECRET` (dev default `dev-secret`). Header `Authorization: Bearer <token>`. Claims: `sub` (user id), `role` (`customer`|`admin`), `name`, `exp`.
- JSON everywhere, snake_case fields, timestamps RFC3339. Errors: `{"error": "<code>", "message": "<human text>"}` with proper status.

## User service (8081)
- `POST /api/auth/login` `{student_id, name?}` → mock OIDC login (ADR-004): upserts customer, returns `{token, user}`. Real CU SSO swaps in behind the same `IdentityProvider` interface later.
- `POST /api/auth/admin/login` `{email, password}` → `{token, user}` (bcrypt; seed `admin@cu.ac.th` / `admin123`).
- `POST /api/auth/logout` → 204 (stateless; client drops token).
- `GET /api/users/me`, `PATCH /api/users/me` `{name?, language?}` (auth).
- `GET /api/users/{id}` → `{id, name, role, student_id, active_penalty: bool}` (internal viewProfile).
- User: `{id, student_id?, email?, name, role, language: "th"|"en", created_at}`.

## Library service (8082)
- `GET /api/libraries?q=&amenity=` → `[Library]` (searchLibrary via q/amenity).
- `GET /api/libraries/{id}` → `Library` (404 if missing).
- `GET /api/libraries/{id}/rooms` → `[Room]`.
- `GET /api/rooms/{id}` → `Room`.
- Library: `{id, name_th, name_en, capacity, amenities: [string], floors: [{id, name, zones: [{id, name, capacity}]}], image_url?}`.
- Room: `{id, library_id, name, capacity, amenities: [string]}`.
- Seed ~6 real Chula libraries (Central Library, Engineering, Science, Arts, Architecture, Law...) with 2-4 rooms each.

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
- Reservation: `{id, user_id, room_id, start_time, end_time, status: pending|accepted|rejected|cancelled, reference?, checkin_deadline?, created_at}`. Frontend renders QR from `reference` + countdown to `checkin_deadline`.

## Review service (8085)
- `GET /api/reviews?library_id=` → `{average: number, count: int, items: [Review]}` (hidden reviews excluded unless admin).
- `POST /api/reviews` `{library_id, rating 1-5, tags: [string], comment?}` (auth) → 201. Profanity filter (Thai + English wordlist) → 422 `profanity` so UI asks user to revise. Validates library via Library service; reviewer name via User service.
- `GET /api/reviews/flagged` (admin), `PATCH /api/reviews/{id}/moderate` `{hidden: bool}` (admin).
- Review: `{id, library_id, user_id, user_name, rating, tags, comment, hidden, created_at}`. Tags stored as JSONB.
- Allowed tags: `quiet, noisy, cold, hot, strong_wifi, weak_wifi, clean, many_outlets, crowded`.
