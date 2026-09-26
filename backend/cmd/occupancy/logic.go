package main

import (
	"sort"
	"time"
)

type Occupancy struct {
	LibraryID string    `json:"library_id"`
	Current   int       `json:"current"`
	Capacity  int       `json:"capacity"`
	Percent   int       `json:"percent"`
	UpdatedAt time.Time `json:"updated_at"`
	Stale     bool      `json:"stale"`
}

func percent(current, capacity int) int {
	if capacity <= 0 {
		return 0
	}
	p := current * 100 / capacity
	switch {
	case p < 0:
		return 0
	case p > 100:
		return 100
	default:
		return p
	}
}

// isStale matches the contract's "Last updated X mins ago" rule: stale once the
// reading is older than 3 poll intervals.
func isStale(updatedAt, now time.Time, pollInterval time.Duration) bool {
	return now.Sub(updatedAt) > 3*pollInterval
}

// filterAndSuggest implements UC-03: sort least->most crowded, filter by
// min/max percent, and when nothing matches, suggest the overall least crowded
// library instead of an empty list.
func filterAndSuggest(items []Occupancy, min, max *int) (filtered []Occupancy, suggestion *Occupancy) {
	sorted := make([]Occupancy, len(items))
	copy(sorted, items)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Percent < sorted[j].Percent })

	filtered = []Occupancy{}
	for _, o := range sorted {
		if min != nil && o.Percent < *min {
			continue
		}
		if max != nil && o.Percent > *max {
			continue
		}
		filtered = append(filtered, o)
	}

	if len(filtered) == 0 && len(sorted) > 0 {
		least := sorted[0]
		suggestion = &least
	}
	return filtered, suggestion
}
