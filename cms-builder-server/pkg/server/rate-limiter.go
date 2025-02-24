package server

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

const RequestsPerMinute = 100
const WaitingSeconds = 15

type RateLimiter struct {
	mu     sync.Mutex
	rates  map[string]*RateLimit // Key: Client identifier (e.g., IP address)
	limit  int                   // Maximum requests per window
	window time.Duration         // Time window (e.g., 1 minute)
}

type RateLimit struct {
	Count     int       `json:"count"`
	LastReset time.Time `json:"last_reset"`
}

// NewRateLimiter creates a new rate limiter with the given limit and window.
// The limit is the maximum number of requests that can be made within the window.
// The window is the time duration for which the limit applies. It is used to
// determine when the rate limit is reset.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		rates:  make(map[string]*RateLimit),
		limit:  limit,
		window: window,
	}
}

// Allow checks if the client identified by clientIdentifier is allowed to make
// a request according to the rate limiter's rules. It returns true if the
// client is allowed, and false otherwise.
//
// The method is thread-safe.
func (rl *RateLimiter) Allow(clientIdentifier string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	r, ok := rl.rates[clientIdentifier]
	if !ok || r.LastReset.Add(rl.window).Before(now) {
		rl.rates[clientIdentifier] = &RateLimit{Count: 1, LastReset: now}
		return true
	}

	if r.Count < rl.limit {
		r.Count++
		return true
	}

	return false // Rate limit exceeded
}

// RateLimitMiddleware is a middleware function that rate-limits incoming requests.
//
// The middleware uses the given rate limiter to determine if a request is allowed
// or not. If the request is allowed, it calls the next handler in the chain.
// If the request is not allowed, it returns an HTTP 429 Too Many Requests response.
//
// The middleware is thread-safe.
func RateLimitMiddleware() func(http.Handler) http.Handler {

	rateLimiter := NewRateLimiter(RequestsPerMinute, WaitingSeconds*time.Second)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			address := r.RemoteAddr
			clientIP := strings.Split(address, ":")[0]

			if !rateLimiter.Allow(clientIP) {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
