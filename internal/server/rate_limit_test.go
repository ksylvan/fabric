package restapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestProtectedRoutesRateLimitMiddlewareBlocksExcessRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Unix(1_700_000_000, 0)
	limiter := newFixedWindowRateLimiter(2, time.Minute, func() time.Time { return now })

	router := gin.New()
	router.Use(newProtectedRoutesRateLimitMiddlewareWithLimiter(limiter))
	router.GET("/patterns/names", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/patterns/names", nil)
		req.RemoteAddr = "203.0.113.10:1234"

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Fatalf("request %d: expected status %d, got %d", i+1, http.StatusOK, recorder.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/patterns/names", nil)
	req.RemoteAddr = "203.0.113.10:1234"

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, recorder.Code)
	}
	if got := recorder.Header().Get("Retry-After"); got != "60" {
		t.Fatalf("expected Retry-After header %q, got %q", "60", got)
	}
}

func TestProtectedRoutesRateLimitMiddlewareTracksClientsSeparately(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Unix(1_700_000_000, 0)
	limiter := newFixedWindowRateLimiter(1, time.Minute, func() time.Time { return now })

	router := gin.New()
	router.Use(newProtectedRoutesRateLimitMiddlewareWithLimiter(limiter))
	router.GET("/patterns/names", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	reqA := httptest.NewRequest(http.MethodGet, "/patterns/names", nil)
	reqA.RemoteAddr = "203.0.113.10:1234"
	recorderA := httptest.NewRecorder()
	router.ServeHTTP(recorderA, reqA)
	if recorderA.Code != http.StatusOK {
		t.Fatalf("expected first client status %d, got %d", http.StatusOK, recorderA.Code)
	}

	reqB := httptest.NewRequest(http.MethodGet, "/patterns/names", nil)
	reqB.RemoteAddr = "198.51.100.8:4321"
	recorderB := httptest.NewRecorder()
	router.ServeHTTP(recorderB, reqB)
	if recorderB.Code != http.StatusOK {
		t.Fatalf("expected second client status %d, got %d", http.StatusOK, recorderB.Code)
	}
}

func TestProtectedRoutesRateLimitMiddlewareAllowsWindowResetAndSkipsOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Unix(1_700_000_000, 0)
	limiter := newFixedWindowRateLimiter(1, time.Minute, func() time.Time { return now })

	router := gin.New()
	router.Use(newProtectedRoutesRateLimitMiddlewareWithLimiter(limiter))
	router.OPTIONS("/chat", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	router.POST("/chat", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	preflight := httptest.NewRequest(http.MethodOptions, "/chat", nil)
	preflight.RemoteAddr = "203.0.113.10:1234"
	preflightRecorder := httptest.NewRecorder()
	router.ServeHTTP(preflightRecorder, preflight)
	if preflightRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected preflight status %d, got %d", http.StatusNoContent, preflightRecorder.Code)
	}

	post := httptest.NewRequest(http.MethodPost, "/chat", nil)
	post.RemoteAddr = "203.0.113.10:1234"
	postRecorder := httptest.NewRecorder()
	router.ServeHTTP(postRecorder, post)
	if postRecorder.Code != http.StatusOK {
		t.Fatalf("expected first POST status %d, got %d", http.StatusOK, postRecorder.Code)
	}

	blocked := httptest.NewRequest(http.MethodPost, "/chat", nil)
	blocked.RemoteAddr = "203.0.113.10:1234"
	blockedRecorder := httptest.NewRecorder()
	router.ServeHTTP(blockedRecorder, blocked)
	if blockedRecorder.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second POST status %d, got %d", http.StatusTooManyRequests, blockedRecorder.Code)
	}

	now = now.Add(time.Minute)
	reset := httptest.NewRequest(http.MethodPost, "/chat", nil)
	reset.RemoteAddr = "203.0.113.10:1234"
	resetRecorder := httptest.NewRecorder()
	router.ServeHTTP(resetRecorder, reset)
	if resetRecorder.Code != http.StatusOK {
		t.Fatalf("expected reset POST status %d, got %d", http.StatusOK, resetRecorder.Code)
	}
}
