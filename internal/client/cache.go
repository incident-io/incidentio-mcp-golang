package client

import (
	"sync"
	"time"
)

// CacheEntry represents a cached item with expiration
type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

// Cache provides simple TTL-based caching for API responses
type Cache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	ttl     time.Duration
}

// NewCache creates a new cache with the specified TTL
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
	}
}

// Get retrieves a value from the cache if it exists and hasn't expired
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Data, true
}

// Set stores a value in the cache with TTL
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &CacheEntry{
		Data:      value,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
}

// Delete removes a specific entry from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
}

// CleanExpired removes all expired entries from the cache
func (c *Cache) CleanExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			delete(c.entries, key)
		}
	}
}

// Made with Bob
