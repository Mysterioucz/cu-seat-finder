package main

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/cu-seat-finder/backend/internal/httpx"
	"github.com/jackc/pgx/v5"
)

func (s *server) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/occupancy", s.listOccupancy)
	mux.HandleFunc("GET /api/occupancy/{library_id}", s.getOccupancy)
}

func (s *server) loadAll(r *http.Request) ([]Occupancy, error) {
	rows, err := s.db.Query(r.Context(), "SELECT library_id, current, capacity, updated_at FROM occupancy.readings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	now := time.Now()
	var items []Occupancy
	for rows.Next() {
		var o Occupancy
		if err := rows.Scan(&o.LibraryID, &o.Current, &o.Capacity, &o.UpdatedAt); err != nil {
			return nil, err
		}
		o.Percent = percent(o.Current, o.Capacity)
		o.Stale = isStale(o.UpdatedAt, now, s.pollInterval)
		items = append(items, o)
	}
	return items, nil
}

func parseIntPtr(v string) *int {
	if v == "" {
		return nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return nil
	}
	return &n
}

func (s *server) listOccupancy(w http.ResponseWriter, r *http.Request) {
	items, err := s.loadAll(r)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not load occupancy")
		return
	}
	min := parseIntPtr(r.URL.Query().Get("min"))
	max := parseIntPtr(r.URL.Query().Get("max"))
	filtered, suggestion := filterAndSuggest(items, min, max)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": filtered, "suggestion": suggestion})
}

func (s *server) getOccupancy(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("library_id")
	var o Occupancy
	row := s.db.QueryRow(r.Context(),
		"SELECT library_id, current, capacity, updated_at FROM occupancy.readings WHERE library_id=$1", id)
	err := row.Scan(&o.LibraryID, &o.Current, &o.Capacity, &o.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "no occupancy data for this library")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not load occupancy")
		return
	}
	o.Percent = percent(o.Current, o.Capacity)
	o.Stale = isStale(o.UpdatedAt, time.Now(), s.pollInterval)
	httpx.WriteJSON(w, http.StatusOK, o)
}
