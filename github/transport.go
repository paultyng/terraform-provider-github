package github

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/go-github/v65/github"
)

const (
	ctxEtag = ctxEtagType("etag")
	ctxId   = ctxIdType("id")
)

// ctxIdType is used to avoid collisions between packages using context
type ctxIdType string

// ctxEtagType is used to avoid collisions between packages using context
type ctxEtagType string

// etagTransport allows saving API quota by passing previously stored Etag
// available via context to request headers
type etagTransport struct {
	transport http.RoundTripper
}

func (ett *etagTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	etag := ctx.Value(ctxEtag)
	if v, ok := etag.(string); ok && v != "" {
		req.Header.Set("If-None-Match", v)
	}

	return ett.transport.RoundTrip(req)
}

func NewEtagTransport(rt http.RoundTripper) *etagTransport {
	return &etagTransport{transport: rt}
}

// RateLimitTransport implements GitHub's best practices
// for avoiding rate limits
// https://developer.github.com/v3/guides/best-practices-for-integrators/#dealing-with-abuse-rate-limits
type RateLimitTransport struct {
	transport        http.RoundTripper
	nextRequestDelay time.Duration
	writeDelay       time.Duration
	readDelay        time.Duration
	parallelRequests bool

	m sync.Mutex
}

func (rlt *RateLimitTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Make requests for a single user or client ID serially when parallel_requests is false.
	// If parallel_requests is true skips the lock and allow the parallelism defined by terraform itself.
	rlt.smartLock(true)

	// Sleep for the delay that the last request defined. This delay might be different
	// for read and write requests. See isWriteMethod for the distinction between them.
	if rlt.nextRequestDelay > 0 {
		log.Printf("[DEBUG] Sleeping %s between operations", rlt.nextRequestDelay)
		time.Sleep(rlt.nextRequestDelay)
	}

	rlt.nextRequestDelay = rlt.calculateNextDelay(req.Method)

	resp, err := rlt.transport.RoundTrip(req)
	if err != nil {
		rlt.smartLock(false)
		return resp, err
	}

	// Make response body accessible for retries & debugging
	// (work around bug in GitHub SDK)
	// See https://github.com/google/go-github/pull/986
	r1, r2, err := drainBody(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body = r1
	ghErr := github.CheckResponse(resp)
	resp.Body = r2

	// When you have been limited, use the Retry-After response header to slow down.
	if arlErr, ok := ghErr.(*github.AbuseRateLimitError); ok {
		rlt.nextRequestDelay = 0
		retryAfter := arlErr.GetRetryAfter()
		log.Printf("[DEBUG] Abuse detection mechanism triggered, sleeping for %s before retrying",
			retryAfter)
		time.Sleep(retryAfter)
		rlt.smartLock(false)
		return rlt.RoundTrip(req)
	}

	if rlErr, ok := ghErr.(*github.RateLimitError); ok {
		rlt.nextRequestDelay = 0
		retryAfter := time.Until(rlErr.Rate.Reset.Time)
		log.Printf("[DEBUG] Rate limit %d reached, sleeping for %s (until %s) before retrying",
			rlErr.Rate.Limit, retryAfter, time.Now().Add(retryAfter))
		time.Sleep(retryAfter)
		rlt.smartLock(false)
		return rlt.RoundTrip(req)
	}

	rlt.smartLock(false)

	return resp, nil
}

// smartLock wraps the mutex locking system and performs its operation via a boolean input for locking and unlocking.
// It also skips the locking when parallelRequests is set to true since, in this case, the lock is not needed.
func (rlt *RateLimitTransport) smartLock(lock bool) {
	if rlt.parallelRequests {
		return
	}
	if lock {
		rlt.m.Lock()
		return
	}
	rlt.m.Unlock()
}

// calculateNextDelay returns a time.Duration specifying the backoff before the next request
// the actual value depends on the current method being a write or a read request
func (rlt *RateLimitTransport) calculateNextDelay(method string) time.Duration {
	if isWriteMethod(method) {
		return rlt.writeDelay
	}
	return rlt.readDelay
}

type RateLimitTransportOption func(*RateLimitTransport)

// NewRateLimitTransport takes in an http.RoundTripper and a variadic list of
// optional functions that modify the RateLimitTransport struct itself. This
// may be used to alter the write delay in between requests, for example.
func NewRateLimitTransport(rt http.RoundTripper, options ...RateLimitTransportOption) *RateLimitTransport {
	// Default to 1 second of write delay if none is provided
	// Default to no read delay if none is provided
	rlt := &RateLimitTransport{transport: rt, writeDelay: 1 * time.Second, readDelay: 0 * time.Second, parallelRequests: false}

	for _, opt := range options {
		opt(rlt)
	}

	return rlt
}

// WithWriteDelay is used to set the write delay between requests
func WithWriteDelay(d time.Duration) RateLimitTransportOption {
	return func(rlt *RateLimitTransport) {
		rlt.writeDelay = d
	}
}

// WithReadDelay is used to set the delay between read requests
func WithReadDelay(d time.Duration) RateLimitTransportOption {
	return func(rlt *RateLimitTransport) {
		rlt.readDelay = d
	}
}

// WithParallelRequests is used to enforce serial api requests for rate limits
func WithParallelRequests(p bool) RateLimitTransportOption {
	return func(rlt *RateLimitTransport) {
		rlt.parallelRequests = p
	}
}

// drainBody reads all of b to memory and then returns two equivalent
// ReadClosers yielding the same bytes.
func drainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	if b == http.NoBody {
		// No copying needed. Preserve the magic sentinel meaning of NoBody.
		return http.NoBody, http.NoBody, nil
	}
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, b, err
	}
	if err = b.Close(); err != nil {
		return nil, b, err
	}
	return io.NopCloser(&buf), io.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

func isWriteMethod(method string) bool {
	switch method {
	case "POST", "PATCH", "PUT", "DELETE":
		return true
	}
	return false
}

type RetryTransport struct {
	transport       http.RoundTripper
	retryDelay      time.Duration
	maxRetries      int
	retryableErrors map[int]bool
}

type RetryTransportOption func(*RetryTransport)

// NewRetryTransport takes in an http.RoundTripper and a variadic list of
// optional functions that modify the RetryTransport struct itself. This
// may be used to retry after response errors 5xx, for example.
func NewRetryTransport(rt http.RoundTripper, options ...RetryTransportOption) *RetryTransport {
	// Default to no retry if none is provided
	defaultErrors := getDefaultRetriableErrors()
	rlt := &RetryTransport{transport: rt, retryDelay: time.Second, maxRetries: 0, retryableErrors: defaultErrors}

	for _, opt := range options {
		opt(rlt)
	}

	return rlt
}

func (t *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var err error
	var resp *http.Response
	var dataBuffer *bytes.Reader

	for retry := 0; retry <= t.maxRetries; retry++ {
		// Reset the body
		// Code from httpretry (https://github.com/ybbus/httpretry/blob/master/roundtripper.go#L60)
		// if request provides GetBody() we use it as Body,
		// because GetBody can be retrieved arbitrary times for retry
		if req.GetBody != nil {
			bodyReadCloser, _ := req.GetBody()
			req.Body = bodyReadCloser
		} else if req.Body != nil {

			// we need to store the complete body, since we need to reset it if a retry happens
			// but: not very efficient because:
			// a) huge stream data size will all be buffered completely in the memory
			//    imagine: 1GB stream data would work efficiently with io.Copy, but has to be buffered completely in memory
			// b) unnecessary if first attempt succeeds
			// a solution would be to at least support more types for GetBody()

			// store it for the first time
			if dataBuffer == nil {
				data, err := io.ReadAll(req.Body)
				req.Body.Close()
				if err != nil {
					return nil, err
				}
				dataBuffer = bytes.NewReader(data)
				req.ContentLength = int64(dataBuffer.Len())
				req.Body = io.NopCloser(dataBuffer)
			}

			// reset the request body
			if _, err = dataBuffer.Seek(0, io.SeekStart); err != nil {
				return nil, err
			}
		}

		resp, err = t.transport.RoundTrip(req)
		if resp != nil && !t.retryableErrors[resp.StatusCode] {
			return resp, err
		}

		time.Sleep(t.retryDelay)
	}

	return resp, err
}

// WithMaxRetries is used to set the max number of retries when encountering an error
func WithMaxRetries(d int) RetryTransportOption {
	return func(rt *RetryTransport) {
		rt.maxRetries = d
	}
}

// WithRetryableErrors is used to set status codes to retry
func WithRetryableErrors(d map[int]bool) RetryTransportOption {
	return func(rt *RetryTransport) {
		rt.retryableErrors = d
	}
}

// WithRetryDelay is used to set the delay between requests for retrying
func WithRetryDelay(d time.Duration) RetryTransportOption {
	return func(rt *RetryTransport) {
		rt.retryDelay = d
	}
}
