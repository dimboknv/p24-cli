package http

import (
	"net/http"
	"time"

	log "github.com/go-pkgz/lgr"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"
)

// Client is used to make HTTP requests. It adds additional functionality
// like automatic retries to tolerate minor outages and rate limiting
type Client struct {
	httpClient  *retryablehttp.Client
	rateLimiter *rate.Limiter
	logger      log.L
}

// NewClient creates a new Client with specified options.
// NoOp logger, retryablehttp.NewClient(), infinite rate limiter by default
func NewClient(opts ...Option) *Client {
	cli := &Client{
		httpClient:  retryablehttp.NewClient(),
		rateLimiter: rate.NewLimiter(rate.Inf, 1),
	}
	for _, f := range opts {
		f(cli)
	}
	cli.httpClient.Logger = loggerFunc(cli.logger.Logf)
	cli.httpClient.Backoff = func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
		backoff := retryablehttp.DefaultBackoff(min, max, attemptNum, resp)
		return cli.rateLimiter.ReserveN(time.Now().Add(backoff), 1).Delay()
	}

	return cli
}

// Do wraps calling an HTTP method with retries and rate limiting
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	// wait global rate limiter
	if err := c.rateLimiter.Wait(req.Context()); err != nil {
		return nil, errors.Wrapf(err, "wait rate limiter with")
	}

	// send req
	retryableReq, err := retryablehttp.FromRequest(req)
	if err != nil {
		return nil, err
	}
	return c.httpClient.Do(retryableReq)
}

// loggerFunc type is an adapter to allow the use of ordinary functions as Logger.
type loggerFunc func(format string, args ...interface{})

// Printf calls f(format, args...)
func (f loggerFunc) Printf(format string, args ...interface{}) { f(format, args...) }
