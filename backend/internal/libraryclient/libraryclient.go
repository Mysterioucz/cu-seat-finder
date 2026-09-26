// Package libraryclient dials the gRPC Library service (LIBRARY_GRPC_ADDR, default localhost:9082).
// Dialing is lazy, so services start fine before Library is up; calls just fail with UNAVAILABLE.
package libraryclient

import (
	"context"
	"log"
	"os"
	"time"

	pb "github.com/cu-seat-finder/backend/gen/library/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func New() pb.LibraryServiceClient {
	addr := os.Getenv("LIBRARY_GRPC_ADDR")
	if addr == "" {
		addr = "localhost:9082"
	}
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("library grpc client: %v", err)
	}
	return pb.NewLibraryServiceClient(conn)
}

// Timeout bounds a single Library call.
func Timeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 5*time.Second)
}

func IsNotFound(err error) bool { return status.Code(err) == codes.NotFound }
