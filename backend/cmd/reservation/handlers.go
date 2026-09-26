package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	pb "github.com/cu-seat-finder/backend/gen/library/v1"
	"github.com/cu-seat-finder/backend/internal/authn"
	"github.com/cu-seat-finder/backend/internal/httpx"
	"github.com/cu-seat-finder/backend/internal/idgen"
	"github.com/cu-seat-finder/backend/internal/libraryclient"
	"github.com/cu-seat-finder/backend/internal/svcclient"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *server) routes(mux *http.ServeMux) {
	mux.Handle("POST /api/reservations", authn.RequireRole(s.jwtSecret, "customer")(http.HandlerFunc(s.create)))
	mux.Handle("GET /api/reservations/me", authn.Require(s.jwtSecret)(http.HandlerFunc(s.listMine)))
	mux.Handle("GET /api/reservations", authn.RequireRole(s.jwtSecret, "admin")(http.HandlerFunc(s.listAll)))
	mux.Handle("POST /api/reservations/{id}/accept", authn.RequireRole(s.jwtSecret, "admin")(http.HandlerFunc(s.accept)))
	mux.Handle("POST /api/reservations/{id}/reject", authn.RequireRole(s.jwtSecret, "admin")(http.HandlerFunc(s.reject)))
	mux.Handle("POST /api/reservations/{id}/cancel", authn.Require(s.jwtSecret)(http.HandlerFunc(s.cancel)))
}

type profile struct {
	ActivePenalty bool `json:"active_penalty"`
}

type createReq struct {
	RoomID          string `json:"room_id"`
	StartTime       string `json:"start_time"`
	DurationMinutes int    `json:"duration_minutes"`
}

func (s *server) create(w http.ResponseWriter, r *http.Request) {
	claims, _ := authn.FromContext(r.Context())
	var req createReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil || req.DurationMinutes <= 0 || req.RoomID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "bad_request", "room_id, start_time (RFC3339), duration_minutes required")
		return
	}
	endTime := startTime.Add(time.Duration(req.DurationMinutes) * time.Minute)
	ctx := r.Context()

	// Validate room via Library service.
	lctx, cancel := libraryclient.Timeout(ctx)
	defer cancel()
	_, err = s.library.GetRoom(lctx, &pb.GetRoomRequest{Id: req.RoomID})
	if libraryclient.IsNotFound(err) {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "room not found")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusBadGateway, "internal", "could not reach library service")
		return
	}

	// Quota / penalty check via User service (viewProfile collaboration).
	var prof profile
	status, err := svcclient.GetJSON(ctx, s.userURL+"/api/users/"+claims.UserID, &prof)
	if err != nil || status != http.StatusOK {
		httpx.WriteError(w, http.StatusBadGateway, "internal", "could not reach user service")
		return
	}
	existing, err := s.loadActiveBookings(ctx, claims.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not load existing bookings")
		return
	}
	switch quotaCheck(existing, startTime, req.DurationMinutes, prof.ActivePenalty) {
	case errPenalty:
		httpx.WriteError(w, http.StatusForbidden, "penalty", "account has an active penalty")
		return
	case errQuotaExceeded:
		httpx.WriteError(w, http.StatusUnprocessableEntity, "quota_exceeded", "booking quota or daily time limit exceeded")
		return
	}

	res, err := s.insertReservation(ctx, claims.UserID, req.RoomID, startTime, endTime)
	if errors.Is(err, errConflict) {
		httpx.WriteError(w, http.StatusConflict, "conflict", "room is already booked for this time")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not create reservation")
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, res)
}

func (s *server) loadActiveBookings(ctx context.Context, userID string) ([]activeBooking, error) {
	rows, err := s.db.Query(ctx, `
		SELECT start_time, end_time FROM reservation.reservations
		WHERE user_id=$1 AND status IN ('pending','accepted')`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []activeBooking
	for rows.Next() {
		var start, end time.Time
		if err := rows.Scan(&start, &end); err != nil {
			return nil, err
		}
		out = append(out, activeBooking{StartTime: start, DurationMinutes: int(end.Sub(start).Minutes())})
	}
	return out, nil
}

var errConflict = errors.New("conflict")

// insertReservation implements ADR-006: a row lock on the per-room lock row
// inside a transaction, plus the DB's own EXCLUDE constraint as a second line
// of defense against a race the row lock somehow missed.
func (s *server) insertReservation(ctx context.Context, userID, roomID string, start, end time.Time) (Reservation, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Reservation{}, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `INSERT INTO reservation.room_locks (room_id) VALUES ($1) ON CONFLICT DO NOTHING`, roomID); err != nil {
		return Reservation{}, err
	}
	if _, err := tx.Exec(ctx, `SELECT room_id FROM reservation.room_locks WHERE room_id=$1 FOR UPDATE`, roomID); err != nil {
		return Reservation{}, err
	}

	var overlapCount int
	err = tx.QueryRow(ctx, `
		SELECT count(*) FROM reservation.reservations
		WHERE room_id=$1 AND status IN ('pending','accepted')
		AND tstzrange(start_time, end_time, '[)') && tstzrange($2, $3, '[)')`,
		roomID, start, end).Scan(&overlapCount)
	if err != nil {
		return Reservation{}, err
	}
	if overlapCount > 0 {
		return Reservation{}, errConflict
	}

	id := idgen.New()
	row := tx.QueryRow(ctx, `
		INSERT INTO reservation.reservations (id, user_id, room_id, start_time, end_time, status)
		VALUES ($1,$2,$3,$4,$5,'pending')
		RETURNING `+reservationCols, id, userID, roomID, start, end)
	res, err := scanReservation(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23P01" { // exclusion_violation
			return Reservation{}, errConflict
		}
		return Reservation{}, err
	}
	return res, tx.Commit(ctx)
}

func (s *server) listMine(w http.ResponseWriter, r *http.Request) {
	claims, _ := authn.FromContext(r.Context())
	rows, err := s.db.Query(r.Context(),
		"SELECT "+reservationCols+" FROM reservation.reservations WHERE user_id=$1 ORDER BY created_at DESC", claims.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not load reservations")
		return
	}
	defer rows.Close()
	out := []Reservation{}
	for rows.Next() {
		res, err := scanReservation(rows)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not scan reservation")
			return
		}
		out = append(out, res)
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

func (s *server) listAll(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	query := "SELECT " + reservationCols + " FROM reservation.reservations"
	args := []any{}
	if status != "" {
		query += " WHERE status=$1"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC"
	rows, err := s.db.Query(r.Context(), query, args...)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not load reservations")
		return
	}
	defer rows.Close()
	out := []Reservation{}
	for rows.Next() {
		res, err := scanReservation(rows)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not scan reservation")
			return
		}
		out = append(out, res)
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

func (s *server) accept(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	reference := idgen.Reference()
	row := s.db.QueryRow(r.Context(), `
		UPDATE reservation.reservations
		SET status='accepted', reference=$2, checkin_deadline = start_time + interval '15 minutes'
		WHERE id=$1 AND status='pending'
		RETURNING `+reservationCols, id, reference)
	res, err := scanReservation(row)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "no pending reservation with that id")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not accept reservation")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}

func (s *server) reject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	row := s.db.QueryRow(r.Context(), `
		UPDATE reservation.reservations SET status='rejected'
		WHERE id=$1 AND status='pending'
		RETURNING `+reservationCols, id)
	res, err := scanReservation(row)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "no pending reservation with that id")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not reject reservation")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}

func (s *server) cancel(w http.ResponseWriter, r *http.Request) {
	claims, _ := authn.FromContext(r.Context())
	id := r.PathValue("id")
	row := s.db.QueryRow(r.Context(), `
		UPDATE reservation.reservations SET status='cancelled'
		WHERE id=$1 AND user_id=$2 AND status IN ('pending','accepted')
		RETURNING `+reservationCols, id, claims.UserID)
	res, err := scanReservation(row)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "no cancellable reservation with that id for this user")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not cancel reservation")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}
