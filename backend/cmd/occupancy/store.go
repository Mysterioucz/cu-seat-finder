package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ponytail: CREATE SCHEMA/TABLE IF NOT EXISTS on boot instead of a migration tool.
// Upgrade path once schema changes get non-trivial: golang-migrate.
func migrate(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		CREATE SCHEMA IF NOT EXISTS occupancy;
		CREATE TABLE IF NOT EXISTS occupancy.readings (
			library_id text PRIMARY KEY,
			current int NOT NULL DEFAULT 0,
			capacity int NOT NULL,
			updated_at timestamptz NOT NULL DEFAULT now()
		);
	`)
	return err
}
