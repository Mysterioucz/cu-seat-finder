// Package authn issues and verifies the JWTs shared across all five services
// (HS256, secret from JWT_SECRET). See API-CONTRACT.md for the claim shape.
package authn

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"sub"`
	Role   string `json:"role"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}

func Issue(secret, userID, role, name string) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		Name:   name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func Parse(secret, tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

type ctxKey int

const claimsKey ctxKey = 0

// Require parses the Authorization header and rejects the request with 401 if
// missing/invalid. Handlers read the claims back with FromContext.
func Require(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := fromHeader(secret, r)
			if err != nil {
				http.Error(w, `{"error":"unauthorized","message":"missing or invalid token"}`, http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole is like Require but additionally rejects the request with 403 if
// the caller's role doesn't match.
func RequireRole(secret, role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := fromHeader(secret, r)
			if err != nil {
				http.Error(w, `{"error":"unauthorized","message":"missing or invalid token"}`, http.StatusUnauthorized)
				return
			}
			if claims.Role != role {
				http.Error(w, `{"error":"forbidden","message":"requires `+role+` role"}`, http.StatusForbidden)
				return
			}
			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func fromHeader(secret string, r *http.Request) (*Claims, error) {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return nil, errors.New("missing bearer token")
	}
	return Parse(secret, strings.TrimPrefix(h, "Bearer "))
}

func FromContext(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(claimsKey).(*Claims)
	return c, ok
}

// Optional parses the Authorization header if present but never rejects the
// request; used by endpoints that behave differently for admins but are
// otherwise public (e.g. GET /api/reviews).
func Optional(secret string, r *http.Request) (*Claims, bool) {
	claims, err := fromHeader(secret, r)
	if err != nil {
		return nil, false
	}
	return claims, true
}
