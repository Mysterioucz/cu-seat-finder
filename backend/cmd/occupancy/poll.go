package main

import (
	"context"
	"log"
	"time"

	"github.com/cu-seat-finder/backend/internal/svcclient"
)

type libraryRef struct {
	ID       string `json:"id"`
	Capacity int    `json:"capacity"`
}

// pollLoop refreshes occupancy for every known library on a ticker (ADR-005).
// It re-lists libraries each cycle so newly added libraries pick up automatically,
// and validates each one exists in the Library service before publishing a
// reading for it (a library removed there simply stops being polled here).
func (s *server) pollLoop(ctx context.Context) {
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()
	s.pollOnce(ctx) // seed immediately so /api/occupancy isn't empty on boot
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.pollOnce(ctx)
		}
	}
}

func (s *server) pollOnce(ctx context.Context) {
	var libs []libraryRef
	status, err := svcclient.GetJSON(ctx, s.libraryURL+"/api/libraries", &libs)
	if err != nil || status != 200 {
		log.Printf("occupancy: could not fetch libraries (status=%d err=%v), skipping poll", status, err)
		return
	}
	for _, lib := range libs {
		delta, err := s.source.Poll(ctx, lib.ID)
		if err != nil {
			log.Printf("occupancy: turnstile poll failed for %s: %v", lib.ID, err)
			continue
		}
		_, err = s.db.Exec(ctx, `
			INSERT INTO occupancy.readings (library_id, current, capacity, updated_at)
			VALUES ($1, GREATEST(0, LEAST($3, $2)), $3, now())
			ON CONFLICT (library_id) DO UPDATE SET
				current = GREATEST(0, LEAST(occupancy.readings.capacity, occupancy.readings.current + $2)),
				capacity = $3,
				updated_at = now()
		`, lib.ID, delta, lib.Capacity)
		if err != nil {
			log.Printf("occupancy: could not save reading for %s: %v", lib.ID, err)
		}
	}
}
