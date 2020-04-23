package limiter

import (
	"fmt"
	"sync"
	"time"
)

// Rule for key
type Rule struct {
	Key             string
	DurationSeconds int
	Max             int
}

// Executor struct
type Executor struct {
	sync.Mutex
	// the visitor map, using ip address and route as the key
	visitors map[string]*limiter
	rules    []Rule
}

// NewExecutor new executor
func NewExecutor(rules []Rule) *Executor {
	return &Executor{
		visitors: make(map[string]*limiter),
		rules:    rules,
	}
}

// GetVisitor with ip and key
func (e *Executor) GetVisitor(ip, metaKey string) *limiter {
	e.Lock()
	defer e.Unlock()

	key := fmt.Sprintf("%v_%v", ip, metaKey)

	visitor, ok := e.visitors[key]
	if !ok {
		visitor = newLimiter(metaKey, e.rules)
		e.visitors[key] = visitor
	}

	return visitor
}

// CleanupVisitors clean map
func (e *Executor) CleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		e.Lock()
		for key, val := range e.visitors {
			now := time.Now()
			if (val.lockedTime != nil && now.Sub(*val.lockedTime).Seconds() >= 3600) ||
				(val.lockedTime == nil && now.Sub(val.lastTime).Seconds() >= 180) {
				delete(e.visitors, key)
			}
		}
		e.Unlock()
	}
}
