// Package cache implements a k/v store of items.
package cache

import (
	"fmt"
	"log"
	"time"

	"github.com/KarelKubat/goto-meet/item"
)

// Cache is the receiver that wraps necessary data.
type Cache struct {
	m map[string]*item.Item
}

// New returns an initialized cache.
func New() *Cache {
	return &Cache{
		m: map[string]*item.Item{},
	}
}

// Lookup returns true when an item was previously stored. When not, the item is stored and false is returned.
func (c *Cache) Lookup(it *item.Item) bool {
	k := itemKey(it)
	if _, ok := c.m[k]; ok {
		return true
	}
	log.Printf("notification added to cache: %v", it)
	c.m[k] = it
	return false
}

// Weed removes items with timestamps in the past. These don't have to be kept in memory.
func (c *Cache) Weed() {
	now := time.Now()
	for k, it := range c.m {
		if it.Start.Before(now) {
			log.Printf("notification removed from cache: %v (it's in the past)", it)
			delete(c.m, k)
		}
	}
}

// itemKey is a helper to derive a distinctive key for an item.
func itemKey(it *item.Item) string {
	return fmt.Sprintf("%v::%v::%v::%v", it.Title, it.JoinLink, it.CalendarLink, it.Start)
}
