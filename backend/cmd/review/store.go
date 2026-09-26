package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ponytail: CREATE SCHEMA/TABLE IF NOT EXISTS on boot instead of a migration tool.
// Upgrade path once schema changes get non-trivial: golang-migrate.
func migrate(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		CREATE SCHEMA IF NOT EXISTS review;
		CREATE TABLE IF NOT EXISTS review.reviews (
			id text PRIMARY KEY,
			library_id text NOT NULL,
			user_id text NOT NULL,
			user_name text NOT NULL,
			rating int NOT NULL CHECK (rating BETWEEN 1 AND 5),
			tags jsonb NOT NULL DEFAULT '[]',
			comment text,
			hidden boolean NOT NULL DEFAULT false,
			created_at timestamptz NOT NULL DEFAULT now()
		);
	`)
	return err
}
