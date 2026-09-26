package main

import (
	"testing"
	"time"
)

func day(hour int) time.Time {
	return time.Date(2026, 1, 15, hour, 0, 0, 0, time.UTC)
}

func TestQuotaCheckBlocksOnPenalty(t *testing.T) {
	if err := quotaCheck(nil, day(9), 60, true); err != errPenalty {
		t.Fatalf("expected errPenalty, got %v", err)
	}
}

func TestQuotaCheckBlocksThirdActiveBooking(t *testing.T) {
	existing := []activeBooking{
		{StartTime: day(9), DurationMinutes: 30},
		{StartTime: day(10), DurationMinutes: 30},
	}
	if err := quotaCheck(existing, day(14), 30, false); err != errQuotaExceeded {
		t.Fatalf("expected errQuotaExceeded for 3rd active booking, got %v", err)
	}
}

func TestQuotaCheckBlocksOverDailyMinutes(t *testing.T) {
	existing := []activeBooking{
		{StartTime: day(9), DurationMinutes: 150},
	}
	// same day, 150 + 60 = 210 > 180
	if err := quotaCheck(existing, day(13), 60, false); err != errQuotaExceeded {
		t.Fatalf("expected errQuotaExceeded for >180 min/day, got %v", err)
	}
}

func TestQuotaCheckAllowsDifferentDay(t *testing.T) {
	existing := []activeBooking{
		{StartTime: day(9), DurationMinutes: 150},
	}
	nextDay := day(9).Add(24 * time.Hour)
	if err := quotaCheck(existing, nextDay, 60, false); err != nil {
		t.Fatalf("expected no error for a different day, got %v", err)
	}
}

func TestQuotaCheckAllowsWithinLimits(t *testing.T) {
	existing := []activeBooking{{StartTime: day(9), DurationMinutes: 60}}
	if err := quotaCheck(existing, day(11), 60, false); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestOverlaps(t *testing.T) {
	cases := []struct {
		name                       string
		aStart, aEnd, bStart, bEnd time.Time
		want                       bool
	}{
		{"identical", day(9), day(10), day(9), day(10), true},
		{"partial overlap", day(9), day(10), day(9).Add(30 * time.Minute), day(11), true},
		{"adjacent, touching end", day(9), day(10), day(10), day(11), false},
		{"disjoint", day(9), day(10), day(11), day(12), false},
		{"b within a", day(9), day(12), day(10), day(11), true},
	}
	for _, c := range cases {
		if got := overlaps(c.aStart, c.aEnd, c.bStart, c.bEnd); got != c.want {
			t.Errorf("%s: overlaps() = %v, want %v", c.name, got, c.want)
		}
	}
}
