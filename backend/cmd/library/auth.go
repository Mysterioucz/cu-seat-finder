package main

import (
	"context"
	"path"
	"strings"

	"github.com/cu-seat-finder/backend/internal/authn"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// adminOnly guards every Create*/Update*/Delete* RPC: the "authorization: Bearer <JWT>"
// metadata must carry role=admin. Reads (List*/Get*) stay open.
func adminOnly(secret string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
		m := path.Base(info.FullMethod)
		if strings.HasPrefix(m, "Create") || strings.HasPrefix(m, "Update") || strings.HasPrefix(m, "Delete") {
			md, _ := metadata.FromIncomingContext(ctx)
			var claims *authn.Claims
			var err error
			if v := md.Get("authorization"); len(v) > 0 && strings.HasPrefix(v[0], "Bearer ") {
				claims, err = authn.Parse(secret, strings.TrimPrefix(v[0], "Bearer "))
			}
			if claims == nil || err != nil {
				return nil, status.Error(codes.Unauthenticated, "missing or invalid token")
			}
			if claims.Role != "admin" {
				return nil, status.Error(codes.PermissionDenied, "requires admin role")
			}
		}
		return next(ctx, req)
	}
}
