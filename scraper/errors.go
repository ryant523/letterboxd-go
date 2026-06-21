package letterboxd

import (
	"fmt"
	"time"
)

// ErrRateLimited is returned when Letterboxd throttles requests (HTTP 429)
type ErrRateLimited struct {
	URL        string
	RetryAfter time.Duration
}

func (e *ErrRateLimited) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("letterboxd: rate limited (HTTP 429) for URL %s, retry after %s", e.URL, e.RetryAfter.String())
	}
	return fmt.Sprintf("letterboxd: rate limited (HTTP 429) for URL %s", e.URL)
}

// ErrUnexpectedStatus is returned when Letterboxd responds with a non-200
// and non-429 status code (e.g., 404, 500, 503).
type ErrUnexpectedStatus struct {
	URL        string
	StatusCode int
}

func (e *ErrUnexpectedStatus) Error() string {
	return fmt.Sprintf("letterboxd: unexpected status code %d for URL %s", e.StatusCode, e.URL)
}

// ErrWAFBlock is returned when Cloudflare intercepts the request
// with a 403 Managed Challenge or Block.
type ErrWAFBlock struct {
	URL   string
	RayID string // Optional: Useful for tracking or debugging
}

func (e *ErrWAFBlock) Error() string {
	if e.RayID != "" {
		return fmt.Sprintf("letterboxd: blocked by Cloudflare WAF (HTTP 403) for URL %s (CF-Ray: %s)", e.URL, e.RayID)
	}
	return fmt.Sprintf("letterboxd: blocked by Cloudflare WAF (HTTP 403) for URL %s", e.URL)
}
