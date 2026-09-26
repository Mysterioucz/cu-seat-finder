package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/cu-seat-finder/backend/internal/httpx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Zone struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
}
type Floor struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Zones []Zone `json:"zones"`
}
type Library struct {
	ID       string   `json:"id"`
	NameTh   string   `json:"name_th"`
	NameEn   string   `json:"name_en"`
	Capacity int      `json:"capacity"`
	Amenities []string `json:"amenities"`
	Floors   []Floor  `json:"floors"`
	ImageURL *string  `json:"image_url,omitempty"`
}
type Room struct {
	ID        string   `json:"id"`
	LibraryID string   `json:"library_id"`
	Name      string   `json:"name"`
	Capacity  int      `json:"capacity"`
	Amenities []string `json:"amenities"`
}

func (s *server) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/libraries", s.listLibraries)
	mux.HandleFunc("GET /api/libraries/{id}", s.getLibrary)
	mux.HandleFunc("GET /api/libraries/{id}/rooms", s.listRooms)
	mux.HandleFunc("GET /api/rooms/{id}", s.getRoom)
}

func (s *server) listLibraries(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(r.URL.Query().Get("q"))
	amenity := r.URL.Query().Get("amenity")

	rows, err := s.db.Query(r.Context(), "SELECT id FROM library.libraries ORDER BY id")
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not list libraries")
		return
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not scan library")
			return
		}
		ids = append(ids, id)
	}
	rows.Close()

	libraries := make([]Library, 0, len(ids))
	for _, id := range ids {
		lib, err := loadLibrary(r.Context(), s.db, id)
		if err != nil {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(lib.NameEn), q) && !strings.Contains(strings.ToLower(lib.NameTh), q) {
			continue
		}
		if amenity != "" && !contains(lib.Amenities, amenity) {
			continue
		}
		libraries = append(libraries, *lib)
	}
	httpx.WriteJSON(w, http.StatusOK, libraries)
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func (s *server) getLibrary(w http.ResponseWriter, r *http.Request) {
	lib, err := loadLibrary(r.Context(), s.db, r.PathValue("id"))
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "library not found")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not load library")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, lib)
}

func (s *server) listRooms(w http.ResponseWriter, r *http.Request) {
	libID := r.PathValue("id")
	rooms, err := loadRooms(r.Context(), s.db, "WHERE library_id=$1", libID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not load rooms")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, rooms)
}

func (s *server) getRoom(w http.ResponseWriter, r *http.Request) {
	rooms, err := loadRooms(r.Context(), s.db, "WHERE id=$1", r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not load room")
		return
	}
	if len(rooms) == 0 {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "room not found")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, rooms[0])
}

func loadLibrary(ctx context.Context, db *pgxpool.Pool, id string) (*Library, error) {
	var lib Library
	var amenities []byte
	row := db.QueryRow(ctx, "SELECT id, name_th, name_en, capacity, amenities, image_url FROM library.libraries WHERE id=$1", id)
	if err := row.Scan(&lib.ID, &lib.NameTh, &lib.NameEn, &lib.Capacity, &amenities, &lib.ImageURL); err != nil {
		return nil, err
	}
	_ = json.Unmarshal(amenities, &lib.Amenities)
	if lib.Amenities == nil {
		lib.Amenities = []string{}
	}

	frows, err := db.Query(ctx, "SELECT id, name FROM library.floors WHERE library_id=$1 ORDER BY id", id)
	if err != nil {
		return nil, err
	}
	defer frows.Close()
	for frows.Next() {
		var f Floor
		if err := frows.Scan(&f.ID, &f.Name); err != nil {
			return nil, err
		}
		f.Zones = []Zone{}
		lib.Floors = append(lib.Floors, f)
	}
	if lib.Floors == nil {
		lib.Floors = []Floor{}
	}

	for i := range lib.Floors {
		zrows, err := db.Query(ctx, "SELECT id, name, capacity FROM library.zones WHERE floor_id=$1 ORDER BY id", lib.Floors[i].ID)
		if err != nil {
			return nil, err
		}
		for zrows.Next() {
			var z Zone
			if err := zrows.Scan(&z.ID, &z.Name, &z.Capacity); err != nil {
				zrows.Close()
				return nil, err
			}
			lib.Floors[i].Zones = append(lib.Floors[i].Zones, z)
		}
		zrows.Close()
	}
	return &lib, nil
}

func loadRooms(ctx context.Context, db *pgxpool.Pool, where string, arg string) ([]Room, error) {
	rows, err := db.Query(ctx, "SELECT id, library_id, name, capacity, amenities FROM library.rooms "+where+" ORDER BY id", arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rooms := []Room{}
	for rows.Next() {
		var room Room
		var amenities []byte
		if err := rows.Scan(&room.ID, &room.LibraryID, &room.Name, &room.Capacity, &amenities); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(amenities, &room.Amenities)
		if room.Amenities == nil {
			room.Amenities = []string{}
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}
