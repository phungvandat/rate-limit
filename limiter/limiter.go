package limiter

import (
	"sync"
	"time"
)

type limiter struct {
	sync.Mutex
	// current times request
	current int
	// duration base seconds
	durationSeconds int
	// max request in duration
	max int
	// time for first request
	firstTime time.Time
	// time for last request
	lastTime   time.Time
	lockedTime *time.Time
}

func newLimiter(metaKey string, rules []Rule) *limiter {
	var (
		durationSeconds int = 60
		max             int = 60
	)

	for _, rule := range rules {
		if rule.Key == metaKey {
			durationSeconds = rule.DurationSeconds
			max = rule.Max
			break
		}
	}

	return &limiter{
		durationSeconds: durationSeconds,
		max:             max,
		firstTime:       time.Now(),
	}
}

func (l *limiter) Allow() bool {
	l.Lock()
	defer l.Unlock()
	now := time.Now()
	l.current++

	duration := now.Sub(l.firstTime).Seconds()

	if l.lockedTime != nil || (l.current >= l.max && int(duration) <= l.durationSeconds) {
		if l.lockedTime == nil {
			l.lockedTime = &now
		}
		return false
	}
	if l.current >= l.max || int(duration) > l.durationSeconds {
		l.firstTime = now
		l.current = 1
	}
	l.lastTime = now
	return true
}
