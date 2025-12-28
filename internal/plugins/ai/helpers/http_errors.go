package helpers

import (
	"fmt"
	"io"
	"net/http"
)

// Response size limits to prevent memory exhaustion
const (
	// DefaultErrorBodyLimit is the maximum number of bytes to read from an error response body
	DefaultErrorBodyLimit = 1024
	// DefaultMaxResponseSize is the maximum allowed size for a successful response body
	DefaultMaxResponseSize = 10 * 1024 * 1024 // 10MB
)

// CheckHTTPStatus verifies that an HTTP response has a successful status code.
// If the status is not OK (200-299), it reads up to 'limit' bytes from the response
// body and returns a descriptive error.
func CheckHTTPStatus(resp *http.Response, limit int) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	// Read error details from response body
	bodyPreview, _ := ReadLimitedBody(resp.Body, limit)
	if bodyPreview != "" {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, bodyPreview)
	}
	return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
}

// ReadLimitedBody reads up to maxSize bytes from the reader and returns the content as a string.
// If the read fails, it returns an empty string and the error.
// The reader is NOT closed by this function.
func ReadLimitedBody(reader io.Reader, maxSize int) (string, error) {
	if maxSize <= 0 {
		maxSize = DefaultErrorBodyLimit
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(reader, int64(maxSize)))
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}

// ValidateResponseSize checks if the response body size exceeds the maximum allowed size.
// It reads the body to validate size and returns the body content if valid.
// If the size exceeds maxSize, it returns an error.
func ValidateResponseSize(reader io.Reader, maxSize int) ([]byte, error) {
	if maxSize <= 0 {
		maxSize = DefaultMaxResponseSize
	}

	// Read one byte more than the limit to detect oversized responses
	bodyBytes, err := io.ReadAll(io.LimitReader(reader, int64(maxSize+1)))
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if len(bodyBytes) > maxSize {
		return nil, fmt.Errorf("response body too large: exceeds %d bytes", maxSize)
	}

	return bodyBytes, nil
}
