package main

import (
	"sync"
	"time"
)

const (
	defaultIssueRateLimitMax    = 5
	defaultIssueRateLimitWindow = 10 * time.Minute
)

type issueRateLimitConfig struct {
	// Maximum number of reports per user that create a Linear issue and fire
	// the webhook within the window. Reports over the limit are still stored.
	// Zero uses the default, a negative value disables the limit.
	MaxIssues     int `yaml:"max_issues"`
	WindowMinutes int `yaml:"window_minutes"`
}

// userRateLimiter is a per-user sliding window counter. It is in-memory, which
// is fine while rageshake runs as a single replica.
type userRateLimiter struct {
	max    int
	window time.Duration
	now    func() time.Time

	lock  sync.Mutex
	users map[string][]time.Time
	calls int
}

func newUserRateLimiter(cfg issueRateLimitConfig) *userRateLimiter {
	max := cfg.MaxIssues
	if max == 0 {
		max = defaultIssueRateLimitMax
	}
	window := time.Duration(cfg.WindowMinutes) * time.Minute
	if window <= 0 {
		window = defaultIssueRateLimitWindow
	}
	return &userRateLimiter{
		max:    max,
		window: window,
		now:    time.Now,
		users:  make(map[string][]time.Time),
	}
}

// Allow records a report for the user and reports whether it is within the limit.
// Reports without a verified user are always allowed.
func (l *userRateLimiter) Allow(userID string) bool {
	if l == nil || l.max < 0 || userID == "" {
		return true
	}
	l.lock.Lock()
	defer l.lock.Unlock()

	now := l.now()
	cutoff := now.Add(-l.window)

	l.calls++
	if l.calls%1000 == 0 {
		for user, times := range l.users {
			if len(times) == 0 || !times[len(times)-1].After(cutoff) {
				delete(l.users, user)
			}
		}
	}

	times := l.users[userID]
	kept := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= l.max {
		l.users[userID] = kept
		return false
	}
	l.users[userID] = append(kept, now)
	return true
}
