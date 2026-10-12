package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/cu-seat-finder/backend/internal/broker"
	"github.com/jackc/pgx/v5/pgconn"
)

// penaltyDuration is how long a no-show blocks new bookings. The penalty
// policy lives here, in the service that owns penalties; Reservation only
// reports the fact that a no-show happened.
const penaltyDuration = 7 * 24 * time.Hour

// handleNoShow applies a penalty for a reservation.no_show event. It is
// idempotent: the event can be delivered more than once (ADR-007), and
// reservation_id is the penalties primary key, so a repeat inserts nothing.
// expires_at is anchored to the check-in deadline, not to when the message
// arrived, so a delayed or redelivered event gives the same penalty.
func (s *server) handleNoShow(ctx context.Context, body []byte) error {
	var ev broker.NoShowEvent
	if err := json.Unmarshal(body, &ev); err != nil || ev.ReservationID == "" || ev.UserID == "" {
		return fmt.Errorf("%w: malformed no_show event", broker.ErrPermanent)
	}
	_, err := s.db.Exec(ctx, `
		INSERT INTO users.penalties (reservation_id, user_id, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (reservation_id) DO NOTHING`,
		ev.ReservationID, ev.UserID, ev.CheckinDeadline.Add(penaltyDuration))
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" { // foreign_key_violation: unknown user
		return fmt.Errorf("%w: %v", broker.ErrPermanent, err)
	}
	if err != nil {
		return err // transient (DB down etc.): requeued and retried
	}
	log.Printf("user: no-show penalty for %s on reservation %s recorded (repeats are no-ops)", ev.UserID, ev.ReservationID)
	return nil
}
