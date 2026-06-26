package interop

import (
	"sync"
	"time"
)

// upstreamBaseCache remembers which tRPC base URL last succeeded
// so subsequent calls can skip dead upstreams.
var upstreamBaseCache struct {
	mu        sync.RWMutex
	base      string
	checkedAt time.Time
	ttl       time.Duration
}

func init() {
	upstreamBaseCache.ttl = 30 * time.Second
}

// SetWorkingBase records a base URL that successfully responded.
func SetWorkingBase(base string) {
	upstreamBaseCache.mu.Lock()
	defer upstreamBaseCache.mu.Unlock()
	upstreamBaseCache.base = base
	upstreamBaseCache.checkedAt = time.Now()
}

// GetWorkingBase returns the last known working base URL,
// or empty string if the cache is stale or empty.
func GetWorkingBase() string {
	upstreamBaseCache.mu.RLock()
	defer upstreamBaseCache.mu.RUnlock()
	if upstreamBaseCache.base == "" {
		return ""
	}
	if time.Since(upstreamBaseCache.checkedAt) > upstreamBaseCache.ttl {
		return ""
	}
	return upstreamBaseCache.base
}
