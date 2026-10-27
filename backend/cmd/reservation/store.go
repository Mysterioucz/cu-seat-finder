package main

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ponytail: CREATE SCHEMA/TABLE IF NOT EXISTS on boot instead of a migration tool.
// Upgrade path once schema changes get non-trivial: golang-migrate.
//
// Double-booking (ADR-006) is enforced two ways: an EXCLUDE constraint (belt)
// rejects any overlapping pending/accepted booking for the same room at the DB
// level, and the handler additionally takes a row lock on reservation.room_locks
// inside the transaction before checking/inserting (suspenders) so the 409 path
// is reached by an ordinary row-lock wait rather than only a constraint error.
func migrate(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		CREATE EXTENSION IF NOT EXISTS btree_gist;

		CREATE SCHEMA IF NOT EXISTS reservation;

		CREATE TABLE IF NOT EXISTS reservation.room_locks (
			room_id text PRIMARY KEY
		);

		CREATE TABLE IF NOT EXISTS reservation.reservations (
			id text PRIMARY KEY,
			user_id text NOT NULL,
			room_id text NOT NULL,
			start_time timestamptz NOT NULL,
			end_time timestamptz NOT NULL,
			status text NOT NULL DEFAULT 'pending'
				CHECK (status IN ('pending','accepted','rejected','cancelled','no_show')),
			reference text,
			checkin_deadline timestamptz,
			checked_in_at timestamptz,
			created_at timestamptz NOT NULL DEFAULT now(),
			during tstzrange GENERATED ALWAYS AS (tstzrange(start_time, end_time, '[)')) STORED,
			EXCLUDE USING gist (room_id WITH =, during WITH &&) WHERE (status IN ('pending','accepted'))
		);
		CREATE UNIQUE INDEX IF NOT EXISTS reservations_reference_key ON reservation.reservations (reference);

		-- Tables created before the no-show feature lack these; bring them up to date.
		ALTER TABLE reservation.reservations ADD COLUMN IF NOT EXISTS checked_in_at timestamptz;
		ALTER TABLE reservation.reservations DROP CONSTRAINT IF EXISTS reservations_status_check;
		ALTER TABLE reservation.reservations ADD CONSTRAINT reservations_status_check
			CHECK (status IN ('pending','accepted','rejected','cancelled','no_show'));
	`)
	return err
}

type Reservation struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	RoomID          string     `json:"room_id"`
	StartTime       time.Time  `json:"start_time"`
	EndTime         time.Time  `json:"end_time"`
	Status          string     `json:"status"`
	Reference       *string    `json:"reference,omitempty"`
	CheckinDeadline *time.Time `json:"checkin_deadline,omitempty"`
	CheckedInAt     *time.Time `json:"checked_in_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

const reservationCols = "id, user_id, room_id, start_time, end_time, status, reference, checkin_deadline, checked_in_at, created_at"

func scanReservation(row interface{ Scan(dest ...any) error }) (Reservation, error) {
	var res Reservation
	err := row.Scan(&res.ID, &res.UserID, &res.RoomID, &res.StartTime, &res.EndTime,
		&res.Status, &res.Reference, &res.CheckinDeadline, &res.CheckedInAt, &res.CreatedAt)
	return res, err
}
