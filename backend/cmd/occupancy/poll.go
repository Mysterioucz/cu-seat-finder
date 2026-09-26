package main

import (
	"context"
	"log"
	"time"

	pb "github.com/cu-seat-finder/backend/gen/library/v1"
	"github.com/cu-seat-finder/backend/internal/libraryclient"
)

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
	cctx, cancel := libraryclient.Timeout(ctx)
	defer cancel()
	resp, err := s.library.ListLibraries(cctx, &pb.ListLibrariesRequest{})
	if err != nil {
		log.Printf("occupancy: could not fetch libraries (%v), skipping poll", err)
		return
	}
	for _, lib := range resp.Libraries {
		delta, err := s.source.Poll(ctx, lib.Id)
		if err != nil {
			log.Printf("occupancy: turnstile poll failed for %s: %v", lib.Id, err)
			continue
		}
		_, err = s.db.Exec(ctx, `
			INSERT INTO occupancy.readings (library_id, current, capacity, updated_at)
			VALUES ($1, GREATEST(0, LEAST($3::int, $2::int)), $3::int, now())
			ON CONFLICT (library_id) DO UPDATE SET
				current = GREATEST(0, LEAST(occupancy.readings.capacity, occupancy.readings.current + $2::int)),
				capacity = $3::int,
				updated_at = now()
		`, lib.Id, delta, lib.Capacity)
		if err != nil {
			log.Printf("occupancy: could not save reading for %s: %v", lib.Id, err)
		}
	}
}
