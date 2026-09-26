// Package turnstile holds the TurnstileSource boundary from ADR-005: occupancy
// polling reads from a mock in-process generator today and a real turnstile
// export/endpoint later, without the Occupancy service caring which one.
package turnstile

import (
	"context"
	"math/rand"
)

// Source yields a net tap-in/tap-out delta for a library since the last poll.
// A real source would query the library's turnstile export and can fail (network,
// auth); the mock never does.
type Source interface {
	Poll(ctx context.Context, libraryID string) (delta int, err error)
}

// MockSource generates plausible foot-traffic deltas so occupancy has something
// to show before any library's IT team grants real access (ADR-005 is still
// "Proposed" for that reason).
type MockSource struct{}

func NewMock() MockSource { return MockSource{} }

func (MockSource) Poll(_ context.Context, _ string) (int, error) {
	// ponytail: uniform random walk, not a real diurnal curve — good enough to
	// demo the "stale/percent/sort" UI logic. Replace once real taps arrive.
	return rand.Intn(11) - 4, nil // skews slightly positive: -4..+6
}
