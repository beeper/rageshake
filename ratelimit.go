package main

import (
	"sync"
	"time"
)

type rateLimitConfig struct {
	// Maximum number of events per key within the window. Zero uses the
	// default, a negative value disables the limit.
	Max           int `yaml:"max"`
	WindowMinutes int `yaml:"window_minutes"`
}

var (
	// Reports per verified user that create a Linear issue and call the webhook.
	defaultIssueRateLimit = rateLimitConfig{Max: 5, WindowMinutes: 10}
	// Reports per client IP that are accepted and stored.
	defaultStorageRateLimit = rateLimitConfig{Max: 30, WindowMinutes: 10}
)

// rateLimiter is a sliding window counter per key. It is in-memory, which is
// fine while rageshake runs as a single replica.
type rateLimiter struct {
	max    int
	window time.Duration
	now    func() time.Time

	lock  sync.Mutex
	keys  map[string][]time.Time
	calls int
}

func newRateLimiter(cfg, defaults rateLimitConfig) *rateLimiter {
	max := cfg.Max
	if max == 0 {
		max = defaults.Max
	}
	window := time.Duration(cfg.WindowMinutes) * time.Minute
	if window <= 0 {
		window = time.Duration(defaults.WindowMinutes) * time.Minute
	}
	return &rateLimiter{
		max:    max,
		window: window,
		now:    time.Now,
		keys:   make(map[string][]time.Time),
	}
}

// Allow records an event for the key and reports whether it is within the
// limit. Rejected events are not recorded. An empty key is always allowed.
func (l *rateLimiter) Allow(key string) bool {
	if l == nil || l.max < 0 || key == "" {
		return true
	}
	l.lock.Lock()
	defer l.lock.Unlock()

	now := l.now()
	cutoff := now.Add(-l.window)

	l.calls++
	if l.calls%1000 == 0 {
		for k, times := range l.keys {
			if len(times) == 0 || !times[len(times)-1].After(cutoff) {
				delete(l.keys, k)
			}
		}
	}

	times := l.keys[key]
	kept := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= l.max {
		l.keys[key] = kept
		return false
	}
	l.keys[key] = append(kept, now)
	return true
}
