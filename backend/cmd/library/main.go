// Library service: library, floor/zone, and meeting room master data.
// gRPC only (proto/library/v1/library.proto). Schema: library. See API-CONTRACT.md.
package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	pb "github.com/cu-seat-finder/backend/gen/library/v1"
	"github.com/cu-seat-finder/backend/internal/dbx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	databaseURL := getenv("DATABASE_URL", "postgres://seat:seat@localhost:5432/seatfinder")
	port := getenv("GRPC_PORT", "9082")
	jwtSecret := getenv("JWT_SECRET", "dev-secret")

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

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	g := grpc.NewServer(grpc.UnaryInterceptor(adminOnly(jwtSecret)))
	pb.RegisterLibraryServiceServer(g, &server{db: pool})
	reflection.Register(g)

	log.Printf("library gRPC service listening on :%s", port)
	log.Fatal(g.Serve(lis))
}
