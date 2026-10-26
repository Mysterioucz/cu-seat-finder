// User service: authentication (mock OIDC + admin bcrypt) and profile management.
// Owns /api/auth, /api/users. Schema: users. See API-CONTRACT.md.
// Consumes reservation.no_show from RabbitMQ to apply penalties (ADR-007).
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cu-seat-finder/backend/internal/broker"
	"github.com/cu-seat-finder/backend/internal/dbx"
	"github.com/cu-seat-finder/backend/internal/httpx"
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
	port := getenv("PORT", "8081")
	rabbitURL := getenv("RABBITMQ_URL", "amqp://seat:seat@localhost:5672/")

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

	srv := &server{db: pool, jwtSecret: jwtSecret}
	go broker.Consume(context.Background(), rabbitURL, broker.QueueUserNoShow, srv.handleNoShow)

	mux := http.NewServeMux()
	srv.routes(mux)
	mux.HandleFunc("GET /healthz", httpx.Healthz)

	log.Printf("user service listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, httpx.WithCORS(mux)))
}

type server struct {
	db        *pgxpool.Pool
	jwtSecret string
}
