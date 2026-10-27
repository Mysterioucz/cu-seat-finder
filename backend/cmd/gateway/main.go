// API Gateway (ADR-008): the single entry point for /api/* traffic. Routes each
// path prefix to the service that owns it, checks the JWT before a request
// reaches a service, translates /api/libraries and /api/rooms into Library gRPC
// calls, rate-limits per client, and logs every request with its latency.
// Upstreams are found through the service registry (ADR-009). Owns no data.
package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/cu-seat-finder/backend/internal/httpx"
	"github.com/cu-seat-finder/backend/internal/libraryclient"
	"github.com/cu-seat-finder/backend/internal/registry"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	port := getenv("PORT", "8080")
	jwtSecret := getenv("JWT_SECRET", "dev-secret")
	timeout := 5 * time.Second
	if v, err := time.ParseDuration(os.Getenv("UPSTREAM_TIMEOUT")); err == nil {
		timeout = v
	}
	rate := 50.0 // requests per second per client; 0 disables the limit
	if v, err := strconv.ParseFloat(os.Getenv("RATE_LIMIT"), 64); err == nil {
		rate = v
	}

	// Each upstream is a plain URL for `make run`, or consul://<name> in Compose.
	upstreams := map[string]string{
		"user":        getenv("USER_URL", "http://localhost:8081"),
		"occupancy":   getenv("OCCUPANCY_URL", "http://localhost:8083"),
		"reservation": getenv("RESERVATION_URL", "http://localhost:8084"),
		"review":      getenv("REVIEW_URL", "http://localhost:8085"),
	}

	gw := &gateway{
		jwtSecret: jwtSecret,
		timeout:   timeout,
		library:   libraryclient.New(),
		proxies:   map[string]http.Handler{},
	}
	for name, addr := range upstreams {
		gw.proxies[name] = newProxy(name, addr, timeout)
	}

	mux := http.NewServeMux()
	gw.routes(mux)
	mux.HandleFunc("GET /healthz", httpx.Healthz)

	var h http.Handler = mux
	h = guard(jwtSecret)(h)
	h = newLimiter(rate, int(rate*2)).middleware(h)
	h = httpx.WithCORS(h)
	h = logRequests(h)

	registry.Register("gateway", port, registry.HTTP)
	log.Printf("api gateway listening on :%s (upstream timeout %s, rate limit %.0f/s per client)", port, timeout, rate)
	log.Fatal(http.ListenAndServe(":"+port, h))
}
