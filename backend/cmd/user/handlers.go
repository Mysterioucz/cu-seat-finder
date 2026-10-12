package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/cu-seat-finder/backend/internal/authn"
	"github.com/cu-seat-finder/backend/internal/httpx"
	"github.com/cu-seat-finder/backend/internal/identity"
	"github.com/cu-seat-finder/backend/internal/idgen"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var idp identity.IdentityProvider = identity.MockProvider{}

func (s *server) routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/login", s.login)
	mux.HandleFunc("POST /api/auth/admin/login", s.adminLogin)
	mux.HandleFunc("POST /api/auth/logout", s.logout)
	mux.Handle("GET /api/users/me", authn.Require(s.jwtSecret)(http.HandlerFunc(s.getMe)))
	mux.Handle("PATCH /api/users/me", authn.Require(s.jwtSecret)(http.HandlerFunc(s.patchMe)))
	mux.HandleFunc("GET /api/users/{id}", s.getProfile)
	mux.Handle("GET /api/users/penalized", authn.RequireRole(s.jwtSecret, "admin")(http.HandlerFunc(s.listPenalized)))
	mux.Handle("POST /api/users/{id}/unban", authn.RequireRole(s.jwtSecret, "admin")(http.HandlerFunc(s.unban)))
}

type loginReq struct {
	StudentID string `json:"student_id"`
	Name      string `json:"name"`
}

func (s *server) login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	name, err := idp.Authenticate(req.StudentID, req.Name)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	ctx := r.Context()
	u, err := getUserByStudentID(ctx, s.db, req.StudentID)
	if errors.Is(err, pgx.ErrNoRows) {
		id := idgen.New()
		_, err = s.db.Exec(ctx, `
			INSERT INTO users.users (id, student_id, name, role) VALUES ($1,$2,$3,'customer')`,
			id, req.StudentID, name)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not create user")
			return
		}
		u, err = getUserByID(ctx, s.db, id)
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not load user")
		return
	}

	token, err := authn.Issue(s.jwtSecret, u.ID, u.Role, u.Name)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not issue token")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"token": token, "user": u})
}

type adminLoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *server) adminLogin(w http.ResponseWriter, r *http.Request) {
	var req adminLoginReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	var u User
	var hash string
	row := s.db.QueryRow(r.Context(),
		"SELECT "+userCols+", password_hash FROM users.users WHERE email=$1", req.Email)
	err := row.Scan(&u.ID, &u.StudentID, &u.Email, &u.Name, &u.Role, &u.Language, &u.ActivePenalty, &u.CreatedAt, &hash)
	if errors.Is(err, pgx.ErrNoRows) || bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid email or password")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not load user")
		return
	}
	token, err := authn.Issue(s.jwtSecret, u.ID, u.Role, u.Name)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not issue token")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"token": token, "user": u})
}

func (s *server) logout(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) getMe(w http.ResponseWriter, r *http.Request) {
	claims, _ := authn.FromContext(r.Context())
	u, err := getUserByID(r.Context(), s.db, claims.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "user not found")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, u)
}

type patchMeReq struct {
	Name     *string `json:"name"`
	Language *string `json:"language"`
}

func (s *server) patchMe(w http.ResponseWriter, r *http.Request) {
	claims, _ := authn.FromContext(r.Context())
	var req patchMeReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	_, err := s.db.Exec(r.Context(), `
		UPDATE users.users SET
			name = COALESCE($2, name),
			language = COALESCE($3, language)
		WHERE id = $1`, claims.UserID, req.Name, req.Language)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not update user")
		return
	}
	u, err := getUserByID(r.Context(), s.db, claims.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "user not found")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, u)
}

func (s *server) getProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	u, err := getUserByID(r.Context(), s.db, id)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "user not found")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not load user")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, Profile{
		ID: u.ID, Name: u.Name, Role: u.Role, StudentID: u.StudentID, ActivePenalty: u.ActivePenalty,
	})
}

type penalizedUser struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	StudentID string    `json:"student_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (s *server) listPenalized(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query(r.Context(), `
		SELECT u.id, u.name, COALESCE(u.student_id, ''), max(p.expires_at)
		FROM users.penalties p JOIN users.users u ON u.id = p.user_id
		WHERE p.expires_at > now() GROUP BY u.id ORDER BY 4 DESC`)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not load penalized users")
		return
	}
	defer rows.Close()
	items := []penalizedUser{}
	for rows.Next() {
		var p penalizedUser
		if err := rows.Scan(&p.ID, &p.Name, &p.StudentID, &p.ExpiresAt); err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not read penalized users")
			return
		}
		items = append(items, p)
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

// unban expires the user's active penalties now. Rows are kept (not deleted)
// so a redelivered no_show event still hits ON CONFLICT DO NOTHING and cannot
// re-penalise the user.
func (s *server) unban(w http.ResponseWriter, r *http.Request) {
	tag, err := s.db.Exec(r.Context(),
		"UPDATE users.penalties SET expires_at = now() WHERE user_id = $1 AND expires_at > now()", r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", "could not unban user")
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "user has no active penalty")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
