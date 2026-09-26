package main

import (
	"testing"
	"time"
)

func TestPercent(t *testing.T) {
	cases := []struct {
		current, capacity, want int
	}{
		{50, 100, 50},
		{0, 100, 0},
		{100, 100, 100},
		{150, 100, 100}, // clamp over-capacity
		{-5, 100, 0},    // clamp negative
		{10, 0, 0},      // no div by zero
	}
	for _, c := range cases {
		if got := percent(c.current, c.capacity); got != c.want {
			t.Errorf("percent(%d,%d) = %d, want %d", c.current, c.capacity, got, c.want)
		}
	}
}

func TestIsStale(t *testing.T) {
	now := time.Now()
	interval := 30 * time.Second
	if isStale(now.Add(-1*time.Minute), now, interval) {
		t.Error("1 min old with 30s interval (stale at >90s) should not be stale")
	}
	if !isStale(now.Add(-2*time.Minute), now, interval) {
		t.Error("2 min old with 30s interval (stale at >90s) should be stale")
	}
}

func TestFilterAndSuggestSortsAscending(t *testing.T) {
	items := []Occupancy{
		{LibraryID: "b", Percent: 80},
		{LibraryID: "a", Percent: 20},
		{LibraryID: "c", Percent: 50},
	}
	filtered, suggestion := filterAndSuggest(items, nil, nil)
	if suggestion != nil {
		t.Error("expected no suggestion when items match")
	}
	if filtered[0].LibraryID != "a" || filtered[1].LibraryID != "c" || filtered[2].LibraryID != "b" {
		t.Errorf("not sorted ascending: %+v", filtered)
	}
}

func TestFilterAndSuggestNoMatchSuggestsLeastCrowded(t *testing.T) {
	items := []Occupancy{
		{LibraryID: "a", Percent: 60},
		{LibraryID: "b", Percent: 90},
	}
	max := 30
	filtered, suggestion := filterAndSuggest(items, nil, &max)
	if len(filtered) != 0 {
		t.Fatalf("expected empty filtered list, got %+v", filtered)
	}
	if suggestion == nil || suggestion.LibraryID != "a" {
		t.Fatalf("expected suggestion 'a' (least crowded overall), got %+v", suggestion)
	}
}

func TestFilterAndSuggestRespectsMinAndMax(t *testing.T) {
	items := []Occupancy{
		{LibraryID: "a", Percent: 10},
		{LibraryID: "b", Percent: 50},
		{LibraryID: "c", Percent: 90},
	}
	min, max := 20, 60
	filtered, _ := filterAndSuggest(items, &min, &max)
	if len(filtered) != 1 || filtered[0].LibraryID != "b" {
		t.Errorf("expected only 'b' in [20,60], got %+v", filtered)
	}
}
