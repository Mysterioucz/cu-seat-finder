// Review service: ratings, tags, and moderation for libraries.
// Owns /api/reviews. Store: MongoDB collection `reviews`. Collaborates with Library (validate
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
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	mongoURL := getenv("MONGO_URL", "mongodb://localhost:27018")
	mongoDB := getenv("MONGO_DB", "seatfinder")
	jwtSecret := getenv("JWT_SECRET", "dev-secret")
	port := getenv("PORT", "8085")
	userURL := getenv("USER_URL", "http://localhost:8081")

	client, err := dbx.ConnectMongo(mongoURL)
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}
	defer client.Disconnect(context.Background())
	reviews := client.Database(mongoDB).Collection("reviews")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := ensureIndexes(ctx, reviews); err != nil {
		log.Fatalf("indexes: %v", err)
	}

	srv := &server{reviews: reviews, jwtSecret: jwtSecret, library: libraryclient.New(), userURL: userURL}
	mux := http.NewServeMux()
	srv.routes(mux)

	log.Printf("review service listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, httpx.WithCORS(mux)))
}

type server struct {
	reviews   *mongo.Collection
	jwtSecret string
	library   pb.LibraryServiceClient
	userURL   string
}
