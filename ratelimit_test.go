package main

import (
	"testing"
	"time"
)

func TestUserRateLimiter(t *testing.T) {
	now := time.Date(2026, 10, 24, 11, 0, 0, 0, time.UTC)
	l := newUserRateLimiter(issueRateLimitConfig{MaxIssues: 3, WindowMinutes: 10})
	l.now = func() time.Time { return now }

	for i := 0; i < 3; i++ {
		if !l.Allow("@flood:beeper.com") {
			t.Fatalf("report %d should be allowed", i+1)
		}
	}
	if l.Allow("@flood:beeper.com") {
		t.Fatal("4th report within the window should be limited")
	}
	if !l.Allow("@other:beeper.com") {
		t.Fatal("other users should not be affected")
	}
	if !l.Allow("") {
		t.Fatal("unverified reports should not be limited")
	}

	now = now.Add(10*time.Minute + time.Second)
	if !l.Allow("@flood:beeper.com") {
		t.Fatal("report after the window should be allowed")
	}
}

func TestUserRateLimiterDefaultsAndDisable(t *testing.T) {
	l := newUserRateLimiter(issueRateLimitConfig{})
	if l.max != defaultIssueRateLimitMax || l.window != defaultIssueRateLimitWindow {
		t.Fatalf("unexpected defaults: %d per %s", l.max, l.window)
	}

	l = newUserRateLimiter(issueRateLimitConfig{MaxIssues: -1})
	for i := 0; i < 100; i++ {
		if !l.Allow("@user:beeper.com") {
			t.Fatal("negative max_issues should disable the limit")
		}
	}
}
