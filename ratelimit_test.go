package main

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	now := time.Date(2026, 10, 24, 11, 0, 0, 0, time.UTC)
	l := newRateLimiter(rateLimitConfig{Max: 3, WindowMinutes: 10}, defaultIssueRateLimit)
	l.now = func() time.Time { return now }

	for i := 0; i < 3; i++ {
		if !l.Allow("@flood:beeper.com") {
			t.Fatalf("event %d should be allowed", i+1)
		}
	}
	if l.Allow("@flood:beeper.com") {
		t.Fatal("4th event within the window should be limited")
	}
	if !l.Allow("@other:beeper.com") {
		t.Fatal("other keys should not be affected")
	}
	if !l.Allow("") {
		t.Fatal("empty keys should not be limited")
	}

	now = now.Add(10*time.Minute + time.Second)
	if !l.Allow("@flood:beeper.com") {
		t.Fatal("event after the window should be allowed")
	}
}

func TestRateLimiterDefaultsAndDisable(t *testing.T) {
	l := newRateLimiter(rateLimitConfig{}, defaultStorageRateLimit)
	if l.max != 30 || l.window != 10*time.Minute {
		t.Fatalf("unexpected defaults: %d per %s", l.max, l.window)
	}

	l = newRateLimiter(rateLimitConfig{Max: -1}, defaultIssueRateLimit)
	for i := 0; i < 100; i++ {
		if !l.Allow("@user:beeper.com") {
			t.Fatal("negative max should disable the limit")
		}
	}
}

func TestClientIP(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/submit", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	req.Header.Set("Authorization", "Bearer random")
	req.Header.Set("X-Forwarded-For", "198.51.100.7")
	if got := clientIP(req); got != "10.0.0.1" {
		t.Fatalf("expected connection address, got %q", got)
	}
	req.Header.Set("X-Real-IP", "203.0.113.5")
	if got := clientIP(req); got != "203.0.113.5" {
		t.Fatalf("expected X-Real-IP, got %q", got)
	}
}
