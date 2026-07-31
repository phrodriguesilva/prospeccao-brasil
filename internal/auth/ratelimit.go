package auth

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter provides per-IP and per-email token bucket rate limiting.
// 5 attempts per 15 seconds per key. Thread-safe via sync.Map.
type RateLimiter struct {
	ipLimiters    sync.Map // map[string]*rate.Limiter
	emailLimiters sync.Map // map[string]*rate.Limiter
	rate          rate.Limit
	burst         int
}

// NewRateLimiter creates a RateLimiter with 5 bursts per 15 seconds.
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		rate:  rate.Every(15 * time.Second / 5), // 5 per 15 seconds
		burst: 5,
	}
}

// getIPLimiter returns or creates the limiter for the given IP.
func (rl *RateLimiter) getIPLimiter(ip string) *rate.Limiter {
	v, ok := rl.ipLimiters.Load(ip)
	if ok {
		return v.(*rate.Limiter)
	}
	l := rate.NewLimiter(rl.rate, rl.burst)
	v, _ = rl.ipLimiters.LoadOrStore(ip, l)
	return v.(*rate.Limiter)
}

// getEmailLimiter returns or creates the limiter for the given email.
func (rl *RateLimiter) getEmailLimiter(email string) *rate.Limiter {
	v, ok := rl.emailLimiters.Load(email)
	if ok {
		return v.(*rate.Limiter)
	}
	l := rate.NewLimiter(rl.rate, rl.burst)
	v, _ = rl.emailLimiters.LoadOrStore(email, l)
	return v.(*rate.Limiter)
}

// Allow checks if the IP is within the rate limit.
func (rl *RateLimiter) Allow(ip string) bool {
	return rl.getIPLimiter(ip).Allow()
}

// AllowEmail checks if the email is within the rate limit.
func (rl *RateLimiter) AllowEmail(email string) bool {
	return rl.getEmailLimiter(email).Allow()
}

// AllowBoth checks both IP and email rate limits. Returns true only if both
// are within limits.
func (rl *RateLimiter) AllowBoth(ip, email string) bool {
	return rl.Allow(ip) && rl.AllowEmail(email)
}

// Middleware returns an HTTP middleware that rate-limits by IP. If the limit
// is exceeded, returns 429 Too Many Requests.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !rl.Allow(ip) {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// clientIP extracts the client IP from the request. Uses RemoteAddr.
func clientIP(r *http.Request) string {
	// Strip port from RemoteAddr
	host := r.RemoteAddr
	for i := len(host) - 1; i >= 0; i-- {
		if host[i] == ':' {
			return host[:i]
		}
	}
	return host
}
