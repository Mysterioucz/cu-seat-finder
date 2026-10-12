// Reservation service: meeting room booking, quota, and check-in reference.
// Owns /api/reservations. Schema: reservation. Collaborates with Library
// (room validation, ADR-006 row locking) and User (quota/penalty check), and
// publishes no-show events to RabbitMQ for User to apply penalties (ADR-007).
// See API-CONTRACT.md.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	pb "github.com/cu-seat-finder/backend/gen/library/v1"
	"github.com/cu-seat-finder/backend/internal/broker"
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
	rabbitURL := getenv("RABBITMQ_URL", "amqp://seat:seat@localhost:5672/")
	noShowInterval, err := time.ParseDuration(getenv("NOSHOW_INTERVAL", "1m"))
	if err != nil || noShowInterval <= 0 {
		log.Fatalf("NOSHOW_INTERVAL must be a positive duration like 1m or 30s")
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

	events := broker.NewPublisher(rabbitURL)
	defer events.Close()

	srv := &server{db: pool, jwtSecret: jwtSecret, library: libraryclient.New(), userURL: userURL,
		events: events, noShowInterval: noShowInterval}
	go srv.noShowLoop(context.Background())

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

	events         *broker.Publisher
	noShowInterval time.Duration
}
