// Reservation service: meeting room booking, quota, and check-in reference.
// Owns /api/reservations. Schema: reservation. Collaborates with Library
// (room validation, ADR-006 row locking) and User (quota/penalty check).
// See API-CONTRACT.md.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	pb "github.com/cu-seat-finder/backend/gen/library/v1"
	"github.com/cu-seat-finder/backend/internal/dbx"
	"github.com/cu-seat-finder/backend/internal/httpx"
	"github.com/cu-seat-finder/backend/internal/libraryclient"
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
	jwtSecret := getenv("JWT_SECRET", "dev-secret")
	port := getenv("PORT", "8084")
	userURL := getenv("USER_URL", "http://localhost:8081")

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

	srv := &server{db: pool, jwtSecret: jwtSecret, library: libraryclient.New(), userURL: userURL}
	mux := http.NewServeMux()
	srv.routes(mux)

	log.Printf("reservation service listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, httpx.WithCORS(mux)))
}

type server struct {
	db        *pgxpool.Pool
	jwtSecret string
	library   pb.LibraryServiceClient
	userURL   string
}
