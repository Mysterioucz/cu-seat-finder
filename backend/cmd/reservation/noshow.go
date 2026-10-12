package main

import (
	"context"
	"log"
	"time"

	"github.com/cu-seat-finder/backend/internal/broker"
)

const noShowInterval = time.Minute

// noShowLoop sweeps for accepted bookings whose check-in deadline passed with
// no check-in, and tells the rest of the system via RabbitMQ (ADR-007). The
// User service consumes the event and applies the penalty; Reservation never
// writes to another service's data.
func (s *server) noShowLoop(ctx context.Context) {
	ticker := time.NewTicker(noShowInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.sweepNoShows(ctx); err != nil {
				log.Printf("reservation: no-show sweep: %v", err)
			}
		}
	}
}

// sweepNoShows uses the reservations table itself as the outbox, so no event
// is ever lost even though the DB and RabbitMQ can't share a transaction:
//
//  1. lock due rows (FOR UPDATE SKIP LOCKED: a concurrent check-in or a second
//     replica's sweep never sees the same row half-done),
//  2. publish each one and wait for the broker's confirm,
//  3. only then mark the published rows 'no_show' and commit.
//
// A failed publish or a crash before 3 rolls everything back, leaving the rows
// 'accepted', so the next sweep publishes them again: duplicates are
// possible, losses are not, and the consumer dedupes on reservation_id.
func (s *server) sweepNoShows(ctx context.Context) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT id, user_id, room_id, start_time, checkin_deadline FROM reservation.reservations
		WHERE status = 'accepted' AND checked_in_at IS NULL AND checkin_deadline < now()
		ORDER BY checkin_deadline
		LIMIT 100
		FOR UPDATE SKIP LOCKED`)
	if err != nil {
		return err
	}
	var due []broker.NoShowEvent
	var ids []string
	for rows.Next() {
		var ev broker.NoShowEvent
		if err := rows.Scan(&ev.ReservationID, &ev.UserID, &ev.RoomID, &ev.StartTime, &ev.CheckinDeadline); err != nil {
			rows.Close()
			return err
		}
		due = append(due, ev)
		ids = append(ids, ev.ReservationID)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	if len(due) == 0 {
		return nil
	}

	for _, ev := range due {
		pctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := s.events.Publish(pctx, broker.RoutingNoShow, "no_show:"+ev.ReservationID, ev)
		cancel()
		if err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `
		UPDATE reservation.reservations SET status = 'no_show' WHERE id = ANY($1)`, ids); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	log.Printf("reservation: published %d no-show event(s)", len(due))
	return nil
}
