// Review service: ratings, tags, and moderation for libraries.
// Owns /api/reviews. Schema: review. Collaborates with Library (validate
// library exists) and User (attach reviewer name). See API-CONTRACT.md.
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
	port := getenv("PORT", "8085")
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

	log.Printf("review service listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, httpx.WithCORS(mux)))
}

type server struct {
	db        *pgxpool.Pool
	jwtSecret string
	library   pb.LibraryServiceClient
	userURL   string
}
