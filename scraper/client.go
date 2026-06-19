// Package letterboxd will scrape movie metadata, lists, diaries, and user statistics from Letterboxd.
// This also can search lists and search for fans of up to 4 movies.
package letterboxd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sardanioss/httpcloak/client"
)

// Client handles all HTTP communication and HTML scraping operations for Letterboxd.
// This must be initialized using [NewClient] for proper configuration. This uses
// a http client with browser TLS/HTTP fingerprint spoofing.
type Client struct {
	timeout    int
	retry      int
	httpClient *client.Client
	logger     *slog.Logger
}

// ClientOption defines a functional configuration option for configuring a [Client]
type ClientOption func(*Client)

// WithTimeout sets the request timeout duration in seconds.
func WithTimeout(timeout int) ClientOption {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithRetry sets the maximum number of network retries for failed HTTP requests.
func WithRetry(retry int) ClientOption {
	return func(c *Client) {
		c.retry = retry
	}
}

// WithLogger allows the user to pass an active structured logger
func WithLogger(logger *slog.Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

// NewClient instatiates a new Letterboxd Client with modern browser cloaking.
// By default, the client is configured with a 10-second timeout and 3 retries.
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		timeout: 10,
		retry:   3,
	}
	for _, opt := range opts {
		opt(c)
	}
	c.httpClient = client.NewClient("chrome-latest",
		client.WithTimeout(time.Duration(c.timeout)*time.Second),
		client.WithRetry(c.retry),
	)
	return c
}

// getHtml executes a GET request against the target URL and processes the response stream
// into a goquery Document.
func (c *Client) getHtml(ctx context.Context, url string) (*goquery.Document, error) {
	c.logger.Info("fetching letterboxd page", "url", url)
	resp, err := c.httpClient.Get(ctx, url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		io.Copy(io.Discard, resp.Body)
		retryAfter := parseRetryAfter(resp.Headers)
		if retryAfter == 0 {
			retryAfter = 30 * time.Second
		}
		return nil, &ErrRateLimited{
			URL:        url,
			RetryAfter: retryAfter,
		}
	}

	if resp.StatusCode != http.StatusOK {
		// Cleanly drain the remaining body data before closing to reuse the TCP connection.
		io.Copy(io.Discard, resp.Body)
		return nil, &ErrUnexpectedStatus{
			URL:        url,
			StatusCode: resp.StatusCode,
		}
	}

	return getDocument(resp.Body)
}

// getDocument builds a a goquery HTML document directly from an io.Reader stream.
// This is shared by production network requests and offline unit test data.
func getDocument(body io.Reader) (*goquery.Document, error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}
	return doc, nil
}

// parseRetryAfter attempts to extract a retry duration from httpcloak's response headers.
func parseRetryAfter(respHeaders map[string][]string) time.Duration {
	if respHeaders == nil {
		return 0
	}

	var retryHeader string
	if values, exists := respHeaders["Retry-After"]; exists && len(values) > 0 {
		retryHeader = values[0]
	} else if values, exists := respHeaders["retry-after"]; exists && len(values) > 0 {
		retryHeader = values[0]
	}

	if retryHeader != "" {
		if seconds, err := strconv.Atoi(retryHeader); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}

	return 0
}
