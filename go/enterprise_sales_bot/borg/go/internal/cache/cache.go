// Package cache provides a generic, thread-safe, TTL+LRU in-memory cache
// ported from packages/core/src/services/CacheService.ts.
//
// It supports:
//   - TTL-based auto-expiration (lazy pruning on access)
//   - LRU eviction when the cache reaches maxSize
//   - Optional periodic cleanup goroutine
//   - Event callbacks for set, hit, miss, delete
//   - Singleton registry keyed by namespace
package cache

import (
	"sync"
	"time"
)

// CacheEvent represents an event emitted by the cache.
type CacheEvent struct {
	Type string      // "set", "hit", "miss", "delete"
	Key  string      `json:"key"`
	TTL  int64       `json:"ttl,omitempty"` // ms, for "set" events
}

// CacheOptions configures a Cache instance.
type CacheOptions struct {
	MaxSize        int   // Maximum entries before LRU eviction (default 1000)
	DefaultTTL     int64 // Default TTL in milliseconds (default 60000)
	CleanupInterval int64 // Periodic cleanup interval in ms; 0 = disabled
}

// cacheEntry holds a cached value with timing metadata.
type cacheEntry struct {
	value      interface{}
	createdAt  int64 // Unix ms
	accessedAt int64 // Unix ms
	ttl        int64 // ms
}

// Cache is a generic in-memory cache with LRU eviction and TTL expiration.
type Cache struct {
	mu        sync.RWMutex
	entries   map[string]*cacheEntry
	maxSize   int
	defaultTTL int64 // ms

	stopClean  chan struct{}
	onEvent    func(CacheEvent)
}

var (
	instancesMu sync.RWMutex
	instances   = map[string]*Cache{}
)

// New creates a new Cache with the given options.
func New(opts CacheOptions) *Cache {
	maxSize := opts.MaxSize
	if maxSize <= 0 {
		maxSize = 1000
	}
	defaultTTL := opts.DefaultTTL
	if defaultTTL <= 0 {
		defaultTTL = 60000
	}

	c := &Cache{
		entries:    make(map[string]*cacheEntry, maxSize),
		maxSize:    maxSize,
		defaultTTL: defaultTTL,
		stopClean:  make(chan struct{}),
	}

	if opts.CleanupInterval > 0 {
		go c.cleanupLoop(time.Duration(opts.CleanupInterval) * time.Millisecond)
	}

	return c
}

// OnEvent registers a callback for cache events.
func (c *Cache) OnEvent(fn func(CacheEvent)) {
	c.onEvent = fn
}

func (c *Cache) emit(event CacheEvent) {
	if c.onEvent != nil {
		c.onEvent(event)
	}
}

// Set stores a value in the cache with the default TTL.
func (c *Cache) Set(key string, value interface{}) {
	c.SetTTL(key, value, 0)
}

// SetTTL stores a value in the cache with a specific TTL in milliseconds.
// If ttl <= 0, the default TTL is used.
func (c *Cache) SetTTL(key string, value interface{}, ttl int64) {
	if ttl <= 0 {
		ttl = c.defaultTTL
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict LRU if at capacity and key is new
	if _, exists := c.entries[key]; !exists && len(c.entries) >= c.maxSize {
		c.evictLRU()
	}

	now := time.Now().UnixMilli()
	c.entries[key] = &cacheEntry{
		value:      value,
		createdAt:  now,
		accessedAt: now,
		ttl:        ttl,
	}

	c.emit(CacheEvent{Type: "set", Key: key, TTL: ttl})
}

// Get retrieves a cached value. Returns (value, true) on hit, (nil, false) on miss or expiry.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.entries[key]
	if !exists {
		c.emit(CacheEvent{Type: "miss", Key: key})
		return nil, false
	}

	if c.isExpired(entry) {
		delete(c.entries, key)
		c.emit(CacheEvent{Type: "miss", Key: key})
		return nil, false
	}

	// Update LRU timestamp
	entry.accessedAt = time.Now().UnixMilli()
	c.emit(CacheEvent{Type: "hit", Key: key})
	return entry.value, true
}

// Has checks if a key exists and is not expired.
func (c *Cache) Has(key string) bool {
	c.mu.RLock()
	entry, exists := c.entries[key]
	c.mu.RUnlock()

	if !exists {
		return false
	}
	if c.isExpired(entry) {
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()
		return false
	}
	return true
}

// Delete removes a specific entry. Returns true if the key existed.
func (c *Cache) Delete(key string) bool {
	c.mu.Lock()
	_, existed := c.entries[key]
	if existed {
		delete(c.entries, key)
	}
	c.mu.Unlock()

	if existed {
		c.emit(CacheEvent{Type: "delete", Key: key})
	}
	return existed
}

// Clear removes all entries.
func (c *Cache) Clear() {
	c.mu.Lock()
	c.entries = make(map[string]*cacheEntry, c.maxSize)
	c.mu.Unlock()
}

// Stats returns current cache statistics.
func (c *Cache) Stats() (size int, maxSize int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries), c.maxSize
}

// Dispose stops the cleanup goroutine and clears entries.
func (c *Cache) Dispose() {
	close(c.stopClean)
	c.Clear()
}

// Cached wraps an async function with transparent caching.
// If the key exists and is not expired, returns the cached value.
// Otherwise, calls fn(), caches the result, and returns it.
func Cached(c *Cache, key string, fn func() (interface{}, error), ttl int64) (interface{}, error) {
	if val, ok := c.Get(key); ok {
		return val, nil
	}
	val, err := fn()
	if err != nil {
		return nil, err
	}
	c.SetTTL(key, val, ttl)
	return val, nil
}

// GetInstance returns a singleton Cache for the given namespace.
func GetInstance(namespace string) *Cache {
	instancesMu.RLock()
	c, ok := instances[namespace]
	instancesMu.RUnlock()

	if ok {
		return c
	}

	instancesMu.Lock()
	defer instancesMu.Unlock()

	// Double-check after acquiring write lock
	if c, ok = instances[namespace]; ok {
		return c
	}

	c = New(CacheOptions{})
	instances[namespace] = c
	return c
}

// --- internal ---

func (c *Cache) isExpired(entry *cacheEntry) bool {
	return time.Now().UnixMilli()-entry.createdAt > entry.ttl
}

func (c *Cache) evictLRU() {
	var oldestKey string
	var oldestAccess int64 = -1

	for k, e := range c.entries {
		if oldestAccess < 0 || e.accessedAt < oldestAccess {
			oldestAccess = e.accessedAt
			oldestKey = k
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
	}
}

func (c *Cache) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopClean:
			return
		}
	}
}

func (c *Cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().UnixMilli()
	for k, e := range c.entries {
		if now-e.createdAt > e.ttl {
			delete(c.entries, k)
		}
	}
}
