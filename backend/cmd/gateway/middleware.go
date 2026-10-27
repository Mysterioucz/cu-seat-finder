package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cu-seat-finder/backend/internal/authn"
	"github.com/cu-seat-finder/backend/internal/httpx"
)

// guard rejects a request the caller may not make (accessFor) before it
// reaches a service: 401 without a valid JWT, 403 with the wrong role.
func guard(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			need := accessFor(r.Method, r.URL.Path)
			if need == public {
				next.ServeHTTP(w, r)
				return
			}
			claims, ok := authn.Optional(secret, r)
			if !ok {
				httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing or invalid token")
				return
			}
			if need == admin && claims.Role != "admin" {
				httpx.WriteError(w, http.StatusForbidden, "forbidden", "requires admin role")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// limiter is a token bucket per client: the JWT subject when signed in (each
// load-test user is its own client), otherwise the IP address.
type limiter struct {
	rate  float64
	burst float64
	mu    sync.Mutex
	b     map[string]*bucket
}

type bucket struct {
	tokens float64
	last   time.Time
}

func newLimiter(rate float64, burst int) *limiter {
	l := &limiter{rate: rate, burst: float64(burst), b: map[string]*bucket{}}
	go func() { // forget clients idle for a minute
		for range time.Tick(time.Minute) {
			l.mu.Lock()
			for k, b := range l.b {
				if time.Since(b.last) > time.Minute {
					delete(l.b, k)
				}
			}
			l.mu.Unlock()
		}
	}()
	return l
}

func (l *limiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	b, ok := l.b[key]
	if !ok {
		b = &bucket{tokens: l.burst, last: now}
		l.b[key] = b
	}
	b.tokens = min(l.burst, b.tokens+now.Sub(b.last).Seconds()*l.rate)
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func (l *limiter) middleware(next http.Handler) http.Handler {
	if l.rate <= 0 {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.allow(clientKey(r)) {
			w.Header().Set("Retry-After", "1")
			httpx.WriteError(w, http.StatusTooManyRequests, "rate_limited", "too many requests")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// clientKey reads the JWT subject without verifying it: a forged token only
// buys its sender a separate bucket, and the guard still rejects it.
func clientKey(r *http.Request) string {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		if parts := strings.Split(strings.TrimPrefix(h, "Bearer "), "."); len(parts) == 3 {
			return "jwt:" + parts[1]
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	return "ip:" + host
}

type ctxKey int

const upstreamKey ctxKey = 0

// setUpstream records which instance served the request, for the access log.
func setUpstream(ctx context.Context, addr string) {
	if p, ok := ctx.Value(upstreamKey).(*string); ok {
		*p = addr
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// logRequests gives every request an X-Request-ID (kept if the client sent
// one, passed on to the service) and logs method, path, status, latency and
// the instance that served it.
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			b := make([]byte, 8)
			rand.Read(b)
			id = hex.EncodeToString(b)
			r.Header.Set("X-Request-ID", id)
		}
		w.Header().Set("X-Request-ID", id)

		upstream := "-"
		r = r.WithContext(context.WithValue(r.Context(), upstreamKey, &upstream))
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		if r.URL.Path != "/healthz" {
			log.Printf("%s %s %s %d %dms upstream=%s", id, r.Method, r.URL.Path, sw.status, time.Since(start).Milliseconds(), upstream)
		}
	})
}
