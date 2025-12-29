package helpers

import (
	"net/http"
	"time"
)

// DefaultHTTPTimeout is the default timeout for HTTP requests to AI providers.
// This can be overridden on a per-provider basis if needed.
const DefaultHTTPTimeout = 10 * time.Second

// NewHTTPClient creates a new HTTP client with the specified timeout and optional transport.
// If transport is nil, http.DefaultTransport will be used.
func NewHTTPClient(timeout time.Duration, transport http.RoundTripper) *http.Client {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

// NewHTTPClientWithBearerAuth creates an HTTP client that automatically adds
// a Bearer token to all requests via a custom transport.
func NewHTTPClientWithBearerAuth(timeout time.Duration, bearerToken string) *http.Client {
	transport := &BearerAuthTransport{
		Token: bearerToken,
		Base:  http.DefaultTransport,
	}
	return NewHTTPClient(timeout, transport)
}

// BearerAuthTransport is an http.RoundTripper that adds Bearer token authentication
// to all outgoing requests.
type BearerAuthTransport struct {
	Token string
	Base  http.RoundTripper
}

// RoundTrip implements http.RoundTripper and adds the Bearer token to the Authorization header.
func (t *BearerAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Token != "" {
		// Clone the request to avoid modifying the original
		newReq := req.Clone(req.Context())
		newReq.Header.Set("Authorization", "Bearer "+t.Token)
		return t.Base.RoundTrip(newReq)
	}
	return t.Base.RoundTrip(req)
}
