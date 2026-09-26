package main

import (
	"encoding/json"
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
)

type Review struct {
	ID        string    `json:"id"`
	LibraryID string    `json:"library_id"`
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name"`
	Rating    int       `json:"rating"`
	Tags      []string  `json:"tags"`
	Comment   string    `json:"comment"`
	Hidden    bool      `json:"hidden"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *server) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/reviews", s.list)
	mux.Handle("POST /api/reviews", authn.Require(s.jwtSecret)(http.HandlerFunc(s.create)))
	mux.HandleFunc("GET /api/reviews/{id}", s.get)
	mux.Handle("PUT /api/reviews/{id}", authn.Require(s.jwtSecret)(http.HandlerFunc(s.update)))
	mux.Handle("DELETE /api/reviews/{id}", authn.Require(s.jwtSecret)(http.HandlerFunc(s.remove)))
	mux.Handle("GET /api/reviews/flagged", authn.RequireRole(s.jwtSecret, "admin")(http.HandlerFunc(s.flagged)))
	mux.Handle("PATCH /api/reviews/{id}/moderate", authn.RequireRole(s.jwtSecret, "admin")(http.HandlerFunc(s.moderate)))
}

func scanReview(rows interface{ Scan(dest ...any) error }) (Review, error) {
	var rv Review
	var tags []byte
	err := rows.Scan(&rv.ID, &rv.LibraryID, &rv.UserID, &rv.UserName, &rv.Rating, &tags, &rv.Comment, &rv.Hidden, &rv.CreatedAt)
	if err != nil {
		return rv, err
	}
	_ = json.Unmarshal(tags, &rv.Tags)
	if rv.Tags == nil {
		rv.Tags = []string{}
	}
	return rv, nil
}

const reviewCols = "id, library_id, user_id, user_name, rating, tags, comment, hidden, created_at"

func (s *server) list(w http.ResponseWriter, r *http.Request) {
	libraryID := r.URL.Query().Get("library_id")
	claims, isAuthed := authn.Optional(s.jwtSecret, r)
	isAdmin := isAuthed && claims.Role == "admin"

	query := "SELECT " + reviewCols + " FROM review.reviews WHERE library_id=$1"
	if !isAdmin {
		query += " AND hidden=false"
	}
	query += " ORDER BY created_at DESC"

	rows, err := s.db.Query(r.Context(), query, libraryID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not load reviews")
		return
	}
	defer rows.Close()

	items := []Review{}
	sum := 0
	for rows.Next() {
		rv, err := scanReview(rows)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not scan review")
			return
		}
		items = append(items, rv)
		sum += rv.Rating
	}
	average := 0.0
	if len(items) > 0 {
		average = float64(sum) / float64(len(items))
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"average": average, "count": len(items), "items": items})
}

type createReq struct {
	LibraryID string   `json:"library_id"`
	Rating    int      `json:"rating"`
	Tags      []string `json:"tags"`
	Comment   string   `json:"comment"`
}

type userProfile struct {
	Name string `json:"name"`
}

func (s *server) create(w http.ResponseWriter, r *http.Request) {
	claims, _ := authn.FromContext(r.Context())
	var req createReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if !validBody(w, req.Rating, req.Tags, req.Comment) {
		return
	}

	ctx := r.Context()
	lctx, cancel := libraryclient.Timeout(ctx)
	defer cancel()
	_, err := s.library.GetLibrary(lctx, &pb.GetLibraryRequest{Id: req.LibraryID})
	if libraryclient.IsNotFound(err) {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "library not found")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusBadGateway, "internal", "could not reach library service")
		return
	}

	var user userProfile
	status, err := svcclient.GetJSON(ctx, s.userURL+"/api/users/"+claims.UserID, &user)
	if err != nil || status != http.StatusOK {
		httpx.WriteError(w, http.StatusBadGateway, "internal", "could not reach user service")
		return
	}

	tags, _ := json.Marshal(req.Tags)
	id := idgen.New()
	row := s.db.QueryRow(ctx, `
		INSERT INTO review.reviews (id, library_id, user_id, user_name, rating, tags, comment)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING `+reviewCols, id, req.LibraryID, claims.UserID, user.Name, req.Rating, tags, req.Comment)
	rv, err := scanReview(row)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not create review")
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, rv)
}

// validBody writes the 400/422 response and returns false if the payload is invalid.
func validBody(w http.ResponseWriter, rating int, tags []string, comment string) bool {
	switch {
	case rating < 1 || rating > 5:
		httpx.WriteError(w, http.StatusBadRequest, "bad_request", "rating must be 1-5")
	case !validTags(tags):
		httpx.WriteError(w, http.StatusBadRequest, "bad_request", "unknown tag")
	case containsProfanity(comment):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "profanity", "please revise your comment")
	default:
		return true
	}
	return false
}

// canDelete: the author or any admin.
func canDelete(role, callerID, ownerID string) bool { return role == "admin" || callerID == ownerID }

func (s *server) find(r *http.Request) (Review, error) {
	return scanReview(s.db.QueryRow(r.Context(), "SELECT "+reviewCols+" FROM review.reviews WHERE id=$1", r.PathValue("id")))
}

// writeFindErr maps a find() failure to 404/500; returns true if it wrote a response.
func writeFindErr(w http.ResponseWriter, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, pgx.ErrNoRows):
		httpx.WriteError(w, http.StatusNotFound, "not_found", "review not found")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not load review")
	}
	return true
}

func (s *server) get(w http.ResponseWriter, r *http.Request) {
	rv, err := s.find(r)
	if writeFindErr(w, err) {
		return
	}
	if claims, ok := authn.Optional(s.jwtSecret, r); rv.Hidden && !(ok && claims.Role == "admin") {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "review not found")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, rv)
}

type updateReq struct {
	Rating  int      `json:"rating"`
	Tags    []string `json:"tags"`
	Comment string   `json:"comment"`
}

func (s *server) update(w http.ResponseWriter, r *http.Request) {
	claims, _ := authn.FromContext(r.Context())
	rv, err := s.find(r)
	if writeFindErr(w, err) {
		return
	}
	if rv.UserID != claims.UserID {
		httpx.WriteError(w, http.StatusForbidden, "forbidden", "only the author can edit a review")
		return
	}
	var req updateReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if !validBody(w, req.Rating, req.Tags, req.Comment) {
		return
	}
	tags, _ := json.Marshal(req.Tags)
	row := s.db.QueryRow(r.Context(), `UPDATE review.reviews SET rating=$2, tags=$3, comment=$4 WHERE id=$1 RETURNING `+reviewCols,
		rv.ID, req.Rating, tags, req.Comment)
	rv, err = scanReview(row)
	if writeFindErr(w, err) {
		return
	}
	httpx.WriteJSON(w, http.StatusOK, rv)
}

func (s *server) remove(w http.ResponseWriter, r *http.Request) {
	claims, _ := authn.FromContext(r.Context())
	rv, err := s.find(r)
	if writeFindErr(w, err) {
		return
	}
	if !canDelete(claims.Role, claims.UserID, rv.UserID) {
		httpx.WriteError(w, http.StatusForbidden, "forbidden", "only the author or an admin can delete a review")
		return
	}
	if _, err := s.db.Exec(r.Context(), "DELETE FROM review.reviews WHERE id=$1", rv.ID); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not delete review")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) flagged(w http.ResponseWriter, r *http.Request) {
	// ponytail: admin queue = all reviews, hidden first, capped at 200; paginate when volume grows
	rows, err := s.db.Query(r.Context(), "SELECT "+reviewCols+" FROM review.reviews ORDER BY hidden DESC, created_at DESC LIMIT 200")
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not load flagged reviews")
		return
	}
	defer rows.Close()
	items := []Review{}
	for rows.Next() {
		rv, err := scanReview(rows)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not scan review")
			return
		}
		items = append(items, rv)
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

type moderateReq struct {
	Hidden bool `json:"hidden"`
}

func (s *server) moderate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req moderateReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	tag, err := s.db.Exec(r.Context(), "UPDATE review.reviews SET hidden=$2 WHERE id=$1", id, req.Hidden)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not moderate review")
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "review not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
