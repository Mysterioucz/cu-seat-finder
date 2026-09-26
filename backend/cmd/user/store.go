package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// ponytail: CREATE SCHEMA/TABLE IF NOT EXISTS on boot instead of a migration tool.
// Upgrade path once schema changes get non-trivial: golang-migrate.
func migrate(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		CREATE SCHEMA IF NOT EXISTS users;
		CREATE TABLE IF NOT EXISTS users.users (
			id text PRIMARY KEY,
			student_id text UNIQUE,
			email text UNIQUE,
			password_hash text,
			name text NOT NULL,
			role text NOT NULL CHECK (role IN ('customer','admin')),
			language text NOT NULL DEFAULT 'th',
			active_penalty boolean NOT NULL DEFAULT false,
			created_at timestamptz NOT NULL DEFAULT now()
		);
	`)
	if err != nil {
		return err
	}
	return seed(ctx, db)
}

func seed(ctx context.Context, db *pgxpool.Pool) error {
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, `
		INSERT INTO users.users (id, email, password_hash, name, role)
		VALUES ('admin-seed-1', 'admin@cu.ac.th', $1, 'CU Admin', 'admin')
		ON CONFLICT (email) DO NOTHING
	`, string(hash))
	return err
}

// User is the public shape returned from login/me: {id, student_id?, email?,
// name, role, language, created_at} per API-CONTRACT.md. active_penalty is a
// separate internal field, exposed only via the /api/users/{id} profile view.
type User struct {
	ID            string  `json:"id"`
	StudentID     *string `json:"student_id,omitempty"`
	Email         *string `json:"email,omitempty"`
	Name          string  `json:"name"`
	Role          string  `json:"role"`
	Language      string  `json:"language"`
	ActivePenalty bool    `json:"-"`
	CreatedAt     string  `json:"created_at"`
}

// Profile is the {id, name, role, student_id, active_penalty} shape other
// services fetch via GET /api/users/{id} (viewProfile collaboration).
type Profile struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Role          string  `json:"role"`
	StudentID     *string `json:"student_id,omitempty"`
	ActivePenalty bool    `json:"active_penalty"`
}

func scanUser(row interface {
	Scan(dest ...any) error
}) (User, error) {
	var u User
	err := row.Scan(&u.ID, &u.StudentID, &u.Email, &u.Name, &u.Role, &u.Language, &u.ActivePenalty, &u.CreatedAt)
	return u, err
}

const userCols = "id, student_id, email, name, role, language, active_penalty, created_at"

func getUserByID(ctx context.Context, db *pgxpool.Pool, id string) (User, error) {
	row := db.QueryRow(ctx, "SELECT "+userCols+" FROM users.users WHERE id=$1", id)
	return scanUser(row)
}

func getUserByStudentID(ctx context.Context, db *pgxpool.Pool, studentID string) (User, error) {
	row := db.QueryRow(ctx, "SELECT "+userCols+" FROM users.users WHERE student_id=$1", studentID)
	return scanUser(row)
}
