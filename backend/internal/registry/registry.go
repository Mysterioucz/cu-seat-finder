// Package registry is the service registry client (ADR-009): every service
// registers itself in Consul with a health check, and callers look services up
// by name instead of by a fixed host:port.
//
// An address is either a plain one ("localhost:8081", "http://localhost:8081")
// or a registry name ("consul://user"). Plain addresses pass through untouched,
// so `make run` works without Consul; inside Compose every address is consul://.
//
// Talks to Consul's HTTP API directly (CONSUL_ADDR, e.g. http://consul:8500),
// so there is no client library to vendor.
package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// Scheme marks an address that must be looked up in the registry.
const Scheme = "consul"

// ErrNoInstance means the registry has no healthy instance of the service.
var ErrNoInstance = errors.New("no healthy instance")

var httpClient = &http.Client{Timeout: 3 * time.Second}

// consulAddr is the agent's HTTP API. Empty disables registration.
func consulAddr() string { return os.Getenv("CONSUL_ADDR") }

// Kind selects the health check Consul runs against the instance.
type Kind int

const (
	HTTP Kind = iota // GET http://<ip>:<port>/healthz must return 2xx
	GRPC             // grpc.health.v1 Check on <ip>:<port>
)

// Register adds this instance to the registry and keeps it there: the
// registration is repeated every 30s, so a restarted Consul (dev mode keeps no
// state) learns about the instance again. On SIGINT/SIGTERM the instance is
// deregistered before the process exits, so callers stop picking it at once
// instead of waiting for its health check to fail.
//
// Without CONSUL_ADDR (local `make run`) it does nothing.
func Register(name, port string, kind Kind) {
	if consulAddr() == "" {
		return
	}
	ip, err := ownIP()
	if err != nil {
		log.Printf("registry: %v; %s not registered", err, name)
		return
	}
	host, _ := os.Hostname()
	id := name + "-" + host
	p, _ := strconv.Atoi(port)
	target := net.JoinHostPort(ip, port)

	check := map[string]any{
		"Interval":                       "5s",
		"Timeout":                        "2s",
		"DeregisterCriticalServiceAfter": "1m",
	}
	if kind == GRPC {
		check["GRPC"] = target
	} else {
		check["HTTP"] = "http://" + target + "/healthz"
	}
	body, _ := json.Marshal(map[string]any{
		"ID": id, "Name": name, "Address": ip, "Port": p, "Check": check,
	})

	go func() {
		registered := false
		for {
			err := put("/v1/agent/service/register", body)
			switch {
			case err != nil:
				log.Printf("registry: register %s: %v (retrying)", id, err)
				registered = false
			case !registered:
				log.Printf("registry: registered %s at %s", id, target)
				registered = true
			}
			if registered {
				time.Sleep(30 * time.Second)
			} else {
				time.Sleep(2 * time.Second)
			}
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		if err := put("/v1/agent/service/deregister/"+id, nil); err != nil {
			log.Printf("registry: deregister %s: %v", id, err)
		} else {
			log.Printf("registry: deregistered %s", id)
		}
		os.Exit(0)
	}()
}

func put(path string, body []byte) error {
	req, err := http.NewRequest(http.MethodPut, consulAddr()+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("consul: %s", resp.Status)
	}
	return nil
}

// ownIP is the container's address on the Compose network, the one other
// containers (and Consul's health checks) can reach.
func ownIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, a := range addrs {
		if n, ok := a.(*net.IPNet); ok && !n.IP.IsLoopback() && n.IP.To4() != nil {
			return n.IP.String(), nil
		}
	}
	return "", errors.New("no non-loopback IPv4 address")
}

// Lookup returns host:port of every instance of name whose health check passes.
// Results are cached for a second, so a hot path does not hit Consul per request.
func Lookup(ctx context.Context, name string) ([]string, error) {
	cache.mu.Lock()
	e, ok := cache.m[name]
	cache.mu.Unlock()
	if ok && time.Since(e.at) < time.Second {
		return e.addrs, nil
	}

	addrs, err := fetch(ctx, name)
	if err != nil {
		if ok { // Consul unreachable: keep using the last known instances.
			return e.addrs, nil
		}
		return nil, err
	}
	cache.mu.Lock()
	cache.m[name] = cacheEntry{addrs: addrs, at: time.Now()}
	cache.mu.Unlock()
	return addrs, nil
}

type cacheEntry struct {
	addrs []string
	at    time.Time
}

var cache = struct {
	mu sync.Mutex
	m  map[string]cacheEntry
}{m: map[string]cacheEntry{}}

func fetch(ctx context.Context, name string) ([]string, error) {
	base := consulAddr()
	if base == "" {
		base = "http://localhost:8500"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/v1/health/service/"+url.PathEscape(name)+"?passing=true", nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("registry lookup %s: %w", name, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry lookup %s: %s", name, resp.Status)
	}
	var entries []struct {
		Service struct {
			Address string
			Port    int
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, err
	}
	addrs := make([]string, 0, len(entries))
	for _, e := range entries {
		addrs = append(addrs, net.JoinHostPort(e.Service.Address, strconv.Itoa(e.Service.Port)))
	}
	return addrs, nil
}

var next atomic.Uint64

// Pick returns one healthy instance of name, round-robin across calls, skipping
// any address in avoid (an instance that just refused a connection).
func Pick(ctx context.Context, name string, avoid ...string) (string, error) {
	addrs, err := Lookup(ctx, name)
	if err != nil {
		return "", err
	}
	n := uint64(len(addrs))
	for i := uint64(0); i < n; i++ {
		a := addrs[(next.Add(1)+i)%n]
		if !contains(avoid, a) {
			return a, nil
		}
	}
	return "", fmt.Errorf("%s: %w", name, ErrNoInstance)
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// Name returns the service name of a consul://name address, or "" for a plain one.
func Name(addr string) string {
	u, err := url.Parse(addr)
	if err != nil || u.Scheme != Scheme {
		return ""
	}
	return u.Host
}

// ResolveURL turns consul://user/api/users/1 into http://<healthy instance>/api/users/1.
// Any other URL is returned unchanged.
func ResolveURL(ctx context.Context, raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != Scheme {
		return raw, nil
	}
	addr, err := Pick(ctx, u.Host)
	if err != nil {
		return "", err
	}
	u.Scheme, u.Host = "http", addr
	return u.String(), nil
}
