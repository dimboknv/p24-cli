package http

import (
	log "github.com/go-pkgz/lgr"
	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/time/rate"
)

// Option func type
type Option func(c *Client)

// WithRetryableHTTP sets retryablehttp client, retryablehttp.New() by default
func WithRetryableHTTP(client *retryablehttp.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithRateLimiter sets rate limiter, Infinite by default
func WithRateLimiter(limiter *rate.Limiter) Option {
	return func(c *Client) {
		c.rateLimiter = limiter
	}
}

// WithLogger sets logger limiter, Infinite by default
func WithLogger(logger log.L) Option {
	return func(c *Client) {
		c.logger = logger
	}
}
