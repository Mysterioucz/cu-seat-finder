package main

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/cu-seat-finder/backend/internal/authn"
)

func TestAccessFor(t *testing.T) {
	cases := []struct {
		method, path string
		want         access
	}{
		{"POST", "/api/auth/login", public},
		{"GET", "/api/occupancy", public},
		{"GET", "/api/occupancy/lib-central", public},
		{"GET", "/api/users/me", signedIn},
		{"GET", "/api/users/penalized", admin},
		{"POST", "/api/users/u1/unban", admin},
		{"POST", "/api/reservations", signedIn},
		{"GET", "/api/reservations/me", signedIn},
		{"GET", "/api/reservations", admin},
		{"POST", "/api/reservations/r1/cancel", signedIn},
		{"POST", "/api/reservations/r1/accept", admin},
		{"POST", "/api/reservations/r1/reject", admin},
		{"POST", "/api/reservations/checkin", admin},
		{"GET", "/api/reviews", public},
		{"GET", "/api/reviews/r1", public},
		{"POST", "/api/reviews", signedIn},
		{"DELETE", "/api/reviews/r1", signedIn},
		{"GET", "/api/reviews/flagged", admin},
		{"PATCH", "/api/reviews/r1/moderate", admin},
		{"GET", "/api/libraries", public},
		{"GET", "/api/libraries/lib-central/rooms", public},
		{"POST", "/api/libraries", admin},
		{"DELETE", "/api/rooms/room-1", admin},
		{"GET", "/api/librariesX", public}, // not under /api/libraries; 404s in the mux
	}
	for _, c := range cases {
		if got := accessFor(c.method, c.path); got != c.want {
			t.Errorf("%s %s = %d, want %d", c.method, c.path, got, c.want)
		}
	}
}

func TestGuard(t *testing.T) {
	h := guard("s")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) }))
	customer, _ := authn.Issue("s", "u1", "customer", "A")
	adminTok, _ := authn.Issue("s", "u2", "admin", "B")
	cases := []struct {
		method, path, token string
		want                int
	}{
		{"GET", "/api/occupancy", "", http.StatusTeapot},
		{"POST", "/api/reservations", "", http.StatusUnauthorized},
		{"POST", "/api/reservations", "garbage", http.StatusUnauthorized},
		{"POST", "/api/reservations", customer, http.StatusTeapot},
		{"POST", "/api/reservations/r1/accept", customer, http.StatusForbidden},
		{"POST", "/api/reservations/r1/accept", adminTok, http.StatusTeapot},
	}
	for _, c := range cases {
		r := httptest.NewRequest(c.method, c.path, nil)
		if c.token != "" {
			r.Header.Set("Authorization", "Bearer "+c.token)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != c.want {
			t.Errorf("%s %s: got %d, want %d", c.method, c.path, w.Code, c.want)
		}
	}
}

func TestLimiter(t *testing.T) {
	l := &limiter{rate: 1, burst: 2, b: map[string]*bucket{}}
	if !l.allow("a") || !l.allow("a") {
		t.Fatal("burst of 2 should pass")
	}
	if l.allow("a") {
		t.Fatal("third request should be limited")
	}
	if !l.allow("b") {
		t.Fatal("another client has its own bucket")
	}
}

// fakeConsul answers /v1/health/service/<name> with the given instances.
func fakeConsul(t *testing.T, name string, addrs ...string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/health/service/"+name {
			http.NotFound(w, r)
			return
		}
		var out []map[string]any
		for _, a := range addrs {
			host, port, _ := net.SplitHostPort(a)
			p, _ := strconv.Atoi(port)
			out = append(out, map[string]any{"Service": map[string]any{"Address": host, "Port": p}})
		}
		json.NewEncoder(w).Encode(out)
	}))
	t.Cleanup(srv.Close)
	t.Setenv("CONSUL_ADDR", srv.URL)
}

func instance(t *testing.T, name string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { io.WriteString(w, name) }))
	t.Cleanup(srv.Close)
	return srv.Listener.Addr().String()
}

// Requests through a consul:// upstream are spread over every healthy
// instance, and a GET to an instance that refuses connections is retried on
// another one, so the caller never sees the dead instance.
func TestProxyUsesRegistry(t *testing.T) {
	a, b := instance(t, "a"), instance(t, "b")
	dead := func() string { // a port nothing listens on
		l, _ := net.Listen("tcp", "127.0.0.1:0")
		addr := l.Addr().String()
		l.Close()
		return addr
	}()
	fakeConsul(t, "occupancy-test", a, b, dead)

	p := newProxy("occupancy", "consul://occupancy-test", time.Second)
	seen := map[string]int{}
	for i := 0; i < 30; i++ {
		w := httptest.NewRecorder()
		p.ServeHTTP(w, httptest.NewRequest("GET", "/api/occupancy", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: status %d, body %s", i, w.Code, w.Body)
		}
		seen[w.Body.String()]++
	}
	if seen["a"] == 0 || seen["b"] == 0 {
		t.Fatalf("expected both instances to serve traffic, got %v", seen)
	}
}

func TestProxyNoInstance(t *testing.T) {
	fakeConsul(t, "nothing-test")
	p := newProxy("review", "consul://nothing-test", time.Second)
	w := httptest.NewRecorder()
	p.ServeHTTP(w, httptest.NewRequest("GET", "/api/reviews", nil))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("got %d, want 503: %s", w.Code, w.Body)
	}
}
