package cache

import (
	"fmt"
	"sync"
	"time"
)

type Cache struct {
	mutex             sync.RWMutex
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
	items             map[string]item
}

type item struct {
	Value      interface{}
	Created    time.Time
	Expiration int64
}

func New(defaultExpiration, cleanupInterval time.Duration) *Cache {

	items := make(map[string]item)

	cache := Cache{
		items:             items,
		defaultExpiration: defaultExpiration,
		cleanupInterval:   cleanupInterval,
	}

	if cleanupInterval > 0 {
		cache.startGC()
	}

	return &cache
}

func (c *Cache) set(key string, value interface{}, expiration int64) {
	c.mutex.Lock()
	c.items[key] = item{
		Value:      value,
		Expiration: expiration,
		Created:    time.Now(),
	}
	c.mutex.Unlock()
}

func (c *Cache) Set(key string, value interface{}, duration time.Duration) {

	var expiration int64

	if duration == 0 {
		duration = c.defaultExpiration
	}

	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}

	c.set(key, value, expiration)
}

func (c *Cache) Add(key string, value interface{}, duration time.Duration) error {
	var expiration int64

	if duration == 0 {
		duration = c.defaultExpiration
	}

	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}

	c.mutex.RLock()
	if item, found := c.items[key]; found {

		if !isExpiration(item.Expiration) {
			c.mutex.RUnlock()
			return fmt.Errorf("key %s exist", key)
		}
	}
	c.mutex.RUnlock()

	c.set(key, value, expiration)
	return nil
}

func (c *Cache) Replace(key string, value interface{}, duration time.Duration) error {
	c.mutex.Lock()
	_, found := c.Get(key)
	if !found {
		c.mutex.Unlock()
		return fmt.Errorf("item %s doesn't exist", key)
	}

	c.Set(key, value, duration)
	c.mutex.Unlock()
	return nil
}

func (c *Cache) Get(key string) (interface{}, bool) {

	c.mutex.RLock()
	item, found := c.items[key]
	c.mutex.RUnlock()

	if !found {
		return nil, false
	}

	if isExpiration(item.Expiration) {
		return nil, false
	}

	// if item.Expiration > 0 {
	// 	if time.Now().UnixNano() > item.Expiration {
	// 		return nil, false
	// 	}
	// }

	return item.Value, true
}

func (c *Cache) Delete(key string) {

	c.mutex.Lock()

	if _, found := c.items[key]; !found {
		c.mutex.Unlock()
		return
	}
	delete(c.items, key)
	c.mutex.Unlock()
}

func (c *Cache) Count() int {
	c.mutex.RLock()
	var lenght = len(c.items)
	c.mutex.RUnlock()
	return lenght
}

func isExpiration(expiration int64) bool {
	return time.Now().UnixNano() > expiration
}
