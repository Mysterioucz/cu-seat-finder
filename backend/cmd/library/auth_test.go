package main

import (
	"context"
	"testing"

	"github.com/cu-seat-finder/backend/internal/authn"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestAdminOnly(t *testing.T) {
	const secret = "s"
	admin, _ := authn.Issue(secret, "1", "admin", "A")
	cust, _ := authn.Issue(secret, "2", "customer", "C")
	other, _ := authn.Issue("wrong", "1", "admin", "A")
	ok := func(context.Context, any) (any, error) { return "ok", nil }
	icpt := adminOnly(secret)

	cases := []struct {
		name, method, auth string
		want               codes.Code
	}{
		{"read no token", "/library.v1.LibraryService/ListLibraries", "", codes.OK},
		{"get no token", "/library.v1.LibraryService/GetRoom", "", codes.OK},
		{"write no token", "/library.v1.LibraryService/CreateLibrary", "", codes.Unauthenticated},
		{"write bad signature", "/library.v1.LibraryService/DeleteRoom", "Bearer " + other, codes.Unauthenticated},
		{"write customer", "/library.v1.LibraryService/UpdateRoom", "Bearer " + cust, codes.PermissionDenied},
		{"write admin", "/library.v1.LibraryService/DeleteLibrary", "Bearer " + admin, codes.OK},
	}
	for _, c := range cases {
		ctx := context.Background()
		if c.auth != "" {
			ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", c.auth))
		}
		_, err := icpt(ctx, nil, &grpc.UnaryServerInfo{FullMethod: c.method}, ok)
		if got := status.Code(err); got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}
