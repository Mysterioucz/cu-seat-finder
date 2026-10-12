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
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Review struct {
	ID        string    `json:"id" bson:"_id"`
	LibraryID string    `json:"library_id" bson:"library_id"`
	UserID    string    `json:"user_id" bson:"user_id"`
	UserName  string    `json:"user_name" bson:"user_name"`
	Rating    int       `json:"rating" bson:"rating"`
	Tags      []string  `json:"tags" bson:"tags"`
	Comment   string    `json:"comment" bson:"comment"`
	Hidden    bool      `json:"hidden" bson:"hidden"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
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

// findAll runs a query and decodes every matching review.
func (s *server) findAll(ctx context.Context, filter bson.D, opts *options.FindOptionsBuilder) ([]Review, error) {
	cur, err := s.reviews.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	items := []Review{}
	if err := cur.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *server) list(w http.ResponseWriter, r *http.Request) {
	libraryID := r.URL.Query().Get("library_id")
	claims, isAuthed := authn.Optional(s.jwtSecret, r)
	isAdmin := isAuthed && claims.Role == "admin"

	filter := bson.D{{Key: "library_id", Value: libraryID}}
	if !isAdmin {
		filter = append(filter, bson.E{Key: "hidden", Value: false})
	}
	items, err := s.findAll(r.Context(), filter, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not load reviews")
		return
	}

	sum := 0
	for _, rv := range items {
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
	if req.Tags == nil {
		req.Tags = []string{}
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

	rv := Review{
		ID:        idgen.New(),
		LibraryID: req.LibraryID,
		UserID:    claims.UserID,
		UserName:  user.Name,
		Rating:    req.Rating,
		Tags:      req.Tags,
		Comment:   req.Comment,
		// BSON dates are millisecond precision; truncate so the response matches what is stored.
		CreatedAt: time.Now().UTC().Truncate(time.Millisecond),
	}
	if _, err := s.reviews.InsertOne(ctx, rv); err != nil {
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
	var rv Review
	err := s.reviews.FindOne(r.Context(), bson.D{{Key: "_id", Value: r.PathValue("id")}}).Decode(&rv)
	return rv, err
}

// writeFindErr maps a find() failure to 404/500; returns true if it wrote a response.
func writeFindErr(w http.ResponseWriter, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, mongo.ErrNoDocuments):
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
	if req.Tags == nil {
		req.Tags = []string{}
	}
	if !validBody(w, req.Rating, req.Tags, req.Comment) {
		return
	}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "rating", Value: req.Rating},
		{Key: "tags", Value: req.Tags},
		{Key: "comment", Value: req.Comment},
	}}}
	err = s.reviews.FindOneAndUpdate(r.Context(), bson.D{{Key: "_id", Value: rv.ID}}, update,
		options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&rv)
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
	if _, err := s.reviews.DeleteOne(r.Context(), bson.D{{Key: "_id", Value: rv.ID}}); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not delete review")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) flagged(w http.ResponseWriter, r *http.Request) {
	// ponytail: admin queue = all reviews, hidden first, capped at 200; paginate when volume grows
	opts := options.Find().
		SetSort(bson.D{{Key: "hidden", Value: -1}, {Key: "created_at", Value: -1}}).
		SetLimit(200)
	items, err := s.findAll(r.Context(), bson.D{}, opts)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not load flagged reviews")
		return
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
	res, err := s.reviews.UpdateOne(r.Context(), bson.D{{Key: "_id", Value: id}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "hidden", Value: req.Hidden}}}})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not moderate review")
		return
	}
	if res.MatchedCount == 0 {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "review not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
