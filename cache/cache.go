// Package cache implements a k/v store of items.
package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/KarelKubat/goto-meet/item"
	"github.com/KarelKubat/goto-meet/l"
)

// Cache is the receiver that wraps necessary data.
type Cache struct {
	m  map[string]*item.Item
	mu sync.Mutex
}

// New returns an initialized cache.
func New() *Cache {
	return &Cache{
		m: map[string]*item.Item{},
	}
}

// Lookup returns true when an item was previously stored. When not, the item is stored and false is returned.
func (c *Cache) Lookup(it *item.Item) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	k := itemKey(it)
	if _, ok := c.m[k]; ok {
		return true
	}
	l.Infof("notification added to cache: %v", it)
	c.m[k] = it
	return false
}

// Weed removes items with timestamps in the past. These don't have to be kept in memory.
func (c *Cache) Weed() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for k, it := range c.m {
		if it.Start.Before(now) {
			l.Infof("notification removed from cache: %v (it's in the past)", it)
			delete(c.m, k)
		}
	}
}

// Clear removes all entries from the cache.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	l.Infof("clearing cache")
	c.m = map[string]*item.Item{}
}

// itemKey is a helper to derive a distinctive key for an item.
func itemKey(it *item.Item) string {
	return fmt.Sprintf("%v::%v::%v::%v", it.Title, it.JoinLink, it.CalendarLink, it.Start)
}
