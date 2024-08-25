package main

import (
	"sync"
	"time"
)

type Cache struct {
    data map[string]cacheItem
    mu   sync.RWMutex
}

type cacheItem struct {
    value      interface{}
    expiration time.Time
}

func NewCache() *Cache {
    return &Cache{
        data: make(map[string]cacheItem),
    }
}

func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.data[key] = cacheItem{
        value:      value,
        expiration: time.Now().Add(duration),
    }
}

func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    item, found := c.data[key]
    if !found {
        return nil, false
    }
    if time.Now().After(item.expiration) {
        delete(c.data, key)
        return nil, false
    }
    return item.value, true
}
