package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	pb "github.com/cu-seat-finder/backend/gen/library/v1"
	"github.com/cu-seat-finder/backend/internal/httpx"
	"github.com/cu-seat-finder/backend/internal/registry"
)

type gateway struct {
	jwtSecret string
	timeout   time.Duration
	library   pb.LibraryServiceClient
	proxies   map[string]http.Handler
}

// routes is the route table from Proposal Step 4.
func (g *gateway) routes(mux *http.ServeMux) {
	for prefix, upstream := range map[string]string{
		"/api/auth":         "user",
		"/api/users":        "user",
		"/api/occupancy":    "occupancy",
		"/api/reservations": "reservation",
		"/api/reviews":      "review",
	} {
		mux.Handle(prefix, g.proxies[upstream])
		mux.Handle(prefix+"/", g.proxies[upstream])
	}
	g.libraryRoutes(mux) // /api/libraries, /api/rooms: REST -> gRPC
}

// access is who may call a route: the coarse check the gateway makes before
// forwarding. Services still check roles themselves (defence in depth).
type access int

const (
	public access = iota
	signedIn
	admin
)

func accessFor(method, path string) access {
	read := method == http.MethodGet || method == http.MethodHead
	under := func(prefix string) bool { return path == prefix || strings.HasPrefix(path, prefix+"/") }

	switch {
	case under("/api/users"):
		// Services read other users' profiles (viewProfile) directly, not through
		// the gateway, so a client only ever needs its own.
		if path == "/api/users/me" {
			return signedIn
		}
		return admin

	case under("/api/reservations"):
		if path == "/api/reservations" && read || path == "/api/reservations/checkin" ||
			strings.HasSuffix(path, "/accept") || strings.HasSuffix(path, "/reject") {
			return admin
		}
		return signedIn

	case under("/api/reviews"):
		if path == "/api/reviews/flagged" || strings.HasSuffix(path, "/moderate") {
			return admin
		}
		if read {
			return public
		}
		return signedIn

	case under("/api/libraries"), under("/api/rooms"):
		if read {
			return public
		}
		return admin
	}
	return public // /api/auth, /api/occupancy, /healthz; unknown paths get the mux's 404
}

// newProxy forwards to the upstream service. For a consul:// upstream every
// request picks a healthy instance from the registry (round-robin); a GET that
// cannot connect is retried once on another instance, which covers the few
// seconds between a container dying and its health check turning critical.
func newProxy(name, addr string, timeout time.Duration) http.Handler {
	transport := &http.Transport{
		DialContext:           (&net.Dialer{Timeout: 2 * time.Second}).DialContext,
		ResponseHeaderTimeout: timeout,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
	}
	rp := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetXForwarded()
			pr.Out.URL.Scheme = "http"
		},
		// The gateway sets CORS itself; drop the service's copy so browsers
		// don't see each header twice ("*, *" is rejected).
		ModifyResponse: func(resp *http.Response) error {
			for k := range resp.Header {
				if strings.HasPrefix(k, "Access-Control-") {
					resp.Header.Del(k)
				}
			}
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("gateway: %s %s -> %s: %v", r.Method, r.URL.Path, name, err)
			if errors.Is(err, registry.ErrNoInstance) {
				httpx.WriteError(w, http.StatusServiceUnavailable, "unavailable", name+" service has no healthy instance")
				return
			}
			if errors.Is(err, context.DeadlineExceeded) {
				httpx.WriteError(w, http.StatusGatewayTimeout, "timeout", name+" service did not answer in time")
				return
			}
			httpx.WriteError(w, http.StatusBadGateway, "bad_gateway", name+" service is unreachable")
		},
	}

	service := registry.Name(addr)
	if service == "" { // plain address: fixed upstream
		u, err := url.Parse(addr)
		if err != nil {
			log.Fatalf("gateway: bad upstream %s=%q: %v", name, addr, err)
		}
		rp.Transport = transport
		rewrite := rp.Rewrite
		rp.Rewrite = func(pr *httputil.ProxyRequest) {
			rewrite(pr)
			pr.Out.URL.Host = u.Host
			pr.Out.Host = u.Host
		}
		return rp
	}
	rp.Transport = &registryTransport{service: service, base: transport}
	return rp
}

// registryTransport fills in the upstream host per request from the registry.
type registryTransport struct {
	service string
	base    http.RoundTripper
}

func (t *registryTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	var tried []string
	for {
		addr, err := registry.Pick(r.Context(), t.service, tried...)
		if err != nil {
			return nil, err
		}
		r.URL.Host, r.Host = addr, addr
		setUpstream(r.Context(), addr)
		resp, err := t.base.RoundTrip(r)
		// A refused connection means the request never reached the service, so
		// one retry elsewhere is safe; a request body can only be sent once.
		noBody := r.Body == nil || r.Body == http.NoBody
		if err != nil && noBody && isConnectErr(err) && len(tried) == 0 {
			tried = append(tried, addr)
			continue
		}
		return resp, err
	}
}

func isConnectErr(err error) bool {
	var op *net.OpError
	return errors.As(err, &op) && op.Op == "dial"
}
