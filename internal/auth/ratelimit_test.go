package auth

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestRateLimiterAllow(t *testing.T) {
	rl := NewRateLimiter()
	ip := "192.168.1.1"
	for i := 0; i < 5; i++ {
		if !rl.Allow(ip) {
			t.Errorf("attempt %d should be allowed", i+1)
		}
	}
	if rl.Allow(ip) {
		t.Error("6th attempt should be denied")
	}
}

func TestRateLimiterAllowEmail(t *testing.T) {
	rl := NewRateLimiter()
	email := "test@prospeccao.com.br"
	for i := 0; i < 5; i++ {
		if !rl.AllowEmail(email) {
			t.Errorf("email attempt %d should be allowed", i+1)
		}
	}
	if rl.AllowEmail(email) {
		t.Error("6th email attempt should be denied")
	}
}

func TestRateLimiterDifferentIPs(t *testing.T) {
	rl := NewRateLimiter()
	for i := 0; i < 5; i++ {
		if !rl.Allow("ip-1") {
			t.Errorf("ip-1 attempt %d should be allowed", i+1)
		}
	}
	// Different IP should have its own bucket
	if !rl.Allow("ip-2") {
		t.Error("ip-2 first attempt should be allowed")
	}
}

func TestRateLimiterConcurrent(t *testing.T) {
	rl := NewRateLimiter()
	var wg sync.WaitGroup
	allowed := make([]bool, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			allowed[idx] = rl.Allow("concurrent-ip")
		}(i)
	}
	wg.Wait()
	count := 0
	for _, a := range allowed {
		if a {
			count++
		}
	}
	if count > 5 {
		t.Errorf("concurrent: %d allowed, want <= 5", count)
	}
}

func TestRateLimiterMiddleware(t *testing.T) {
	rl := NewRateLimiter()
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("POST", "/login", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("attempt %d: got %d, want %d", i+1, rec.Code, http.StatusOK)
		}
	}
	// 6th should be 429
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("6th attempt: got %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
}

func TestAllowBoth(t *testing.T) {
	rl := NewRateLimiter()
	ip := "10.0.0.1"
	email := "both@prospeccao.com.br"
	for i := 0; i < 5; i++ {
		if !rl.AllowBoth(ip, email) {
			t.Errorf("attempt %d should be allowed", i+1)
		}
	}
	if rl.AllowBoth(ip, email) {
		t.Error("6th attempt should be denied by both")
	}
}

func TestAllowBothEmailExhausted(t *testing.T) {
	rl := NewRateLimiter()
	// Exhaust email only (different IPs)
	for i := 0; i < 5; i++ {
		if !rl.AllowBoth("ip-"+string(rune(i)), "same@prospeccao.com.br") {
			t.Errorf("attempt %d should be allowed (email limit)", i+1)
		}
	}
	// Same email, new IP -- should be denied by email
	if rl.AllowBoth("new-ip", "same@prospeccao.com.br") {
		t.Error("should be denied by email limit")
	}
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		addr string
		want string
	}{
		{"192.168.1.1:12345", "192.168.1.1"},
		{"[::1]:12345", "[::1]"},
		{"noport", "noport"},
	}
	for _, tt := range tests {
		req := &http.Request{RemoteAddr: tt.addr}
		got := clientIP(req)
		if got != tt.want {
			t.Errorf("clientIP(%q) = %q, want %q", tt.addr, got, tt.want)
		}
	}
}
