// Package svcclient is the plain HTTP client services use to call each other
// (e.g. Reservation -> User). A consul://name URL is resolved to a healthy
// instance through the registry on every call (ADR-009). No retries: dev-scale only.
package svcclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cu-seat-finder/backend/internal/registry"
)

var httpClient = &http.Client{Timeout: 5 * time.Second}

// GetJSON GETs url and decodes a 2xx JSON body into out (out may be nil).
// It returns the HTTP status code so callers can branch on 404 etc.
func GetJSON(ctx context.Context, url string, out any) (int, error) {
	url, err := registry.ResolveURL(ctx, url)
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("calling %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return resp.StatusCode, err
		}
	}
	return resp.StatusCode, nil
}
