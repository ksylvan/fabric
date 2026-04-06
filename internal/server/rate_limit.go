package restapi

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	protectedRoutesRateLimit       = 60
	protectedRoutesRateLimitWindow = time.Minute
)

type fixedWindowRateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	now     func() time.Time
	entries map[string]rateLimitEntry
}

type rateLimitEntry struct {
	windowStartedAt time.Time
	lastSeenAt      time.Time
	count           int
}

func newProtectedRoutesRateLimitMiddleware() gin.HandlerFunc {
	limiter := newFixedWindowRateLimiter(
		protectedRoutesRateLimit,
		protectedRoutesRateLimitWindow,
		time.Now,
	)
	return newProtectedRoutesRateLimitMiddlewareWithLimiter(limiter)
}

func newProtectedRoutesRateLimitMiddlewareWithLimiter(limiter *fixedWindowRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		allowed, retryAfter := limiter.allow(rateLimitKey(c.Request))
		if allowed {
			c.Next()
			return
		}

		retryAfterSeconds := int(math.Ceil(retryAfter.Seconds()))
		if retryAfterSeconds < 1 {
			retryAfterSeconds = 1
		}

		c.Header("Retry-After", strconv.Itoa(retryAfterSeconds))
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
	}
}

func newFixedWindowRateLimiter(limit int, window time.Duration, now func() time.Time) *fixedWindowRateLimiter {
	if now == nil {
		now = time.Now
	}

	return &fixedWindowRateLimiter{
		limit:   limit,
		window:  window,
		now:     now,
		entries: make(map[string]rateLimitEntry),
	}
}

func (l *fixedWindowRateLimiter) allow(key string) (bool, time.Duration) {
	now := l.now()

	l.mu.Lock()
	defer l.mu.Unlock()

	l.pruneStaleEntries(now)

	entry, exists := l.entries[key]
	if !exists || now.Sub(entry.windowStartedAt) >= l.window {
		l.entries[key] = rateLimitEntry{
			windowStartedAt: now,
			lastSeenAt:      now,
			count:           1,
		}
		return true, 0
	}

	entry.lastSeenAt = now
	if entry.count >= l.limit {
		l.entries[key] = entry
		return false, l.window - now.Sub(entry.windowStartedAt)
	}

	entry.count++
	l.entries[key] = entry
	return true, 0
}

func (l *fixedWindowRateLimiter) pruneStaleEntries(now time.Time) {
	cutoff := now.Add(-2 * l.window)
	for key, entry := range l.entries {
		if entry.lastSeenAt.Before(cutoff) {
			delete(l.entries, key)
		}
	}
}

func rateLimitKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	if r.RemoteAddr != "" {
		return r.RemoteAddr
	}
	return "unknown"
}
