package main

import (
	"errors"
	"time"
)

var (
	errQuotaExceeded = errors.New("quota_exceeded")
	errPenalty       = errors.New("penalty")
)

// activeBooking is the subset of an existing reservation quotaCheck needs.
type activeBooking struct {
	StartTime       time.Time
	DurationMinutes int
}

// quotaCheck applies the UC-02 quota subflow: max 2 active (pending/accepted)
// bookings, max 180 booked minutes on the same calendar day as the new booking,
// blocked entirely while the customer has an active penalty.
func quotaCheck(existing []activeBooking, newStart time.Time, newDurationMinutes int, activePenalty bool) error {
	if activePenalty {
		return errPenalty
	}
	if len(existing)+1 > 2 {
		return errQuotaExceeded
	}
	sameDayMinutes := newDurationMinutes
	for _, b := range existing {
		if sameDay(b.StartTime, newStart) {
			sameDayMinutes += b.DurationMinutes
		}
	}
	if sameDayMinutes > 180 {
		return errQuotaExceeded
	}
	return nil
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// overlaps reports whether two half-open [start,end) time ranges intersect —
// the same rule the DB's EXCLUDE constraint enforces (ADR-006), used here so
// the pure logic is unit-testable without a database.
func overlaps(aStart, aEnd, bStart, bEnd time.Time) bool {
	return aStart.Before(bEnd) && bStart.Before(aEnd)
}
