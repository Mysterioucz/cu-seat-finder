// Occupancy service: polls turnstile data (ADR-005) and serves live occupancy.
// Owns /api/occupancy. Schema: occupancy. See API-CONTRACT.md.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/cu-seat-finder/backend/internal/dbx"
	"github.com/cu-seat-finder/backend/internal/httpx"
	"github.com/cu-seat-finder/backend/internal/turnstile"
	"github.com/jackc/pgx/v5/pgxpool"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	databaseURL := getenv("DATABASE_URL", "postgres://seat:seat@localhost:5432/seatfinder")
	port := getenv("PORT", "8083")
	libraryURL := getenv("LIBRARY_URL", "http://localhost:8082")
	pollInterval := 30 * time.Second
	if v := os.Getenv("POLL_INTERVAL"); v != "" {
		if secs, err := strconv.Atoi(v); err == nil {
			pollInterval = time.Duration(secs) * time.Second
		}
	}

	pool, err := dbx.Connect(databaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := migrate(ctx, pool); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	srv := &server{db: pool, libraryURL: libraryURL, pollInterval: pollInterval, source: turnstile.NewMock()}
	go srv.pollLoop(context.Background())

	mux := http.NewServeMux()
	srv.routes(mux)

	log.Printf("occupancy service listening on :%s (poll every %s)", port, pollInterval)
	log.Fatal(http.ListenAndServe(":"+port, httpx.WithCORS(mux)))
}

type server struct {
	db           *pgxpool.Pool
	libraryURL   string
	pollInterval time.Duration
	source       turnstile.Source
}
