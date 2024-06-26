package cache

import "time"

func (c *Cache) startGC() {
	go c.gC()
}

func (c *Cache) gC() {

	for {
		<-time.After(c.cleanupInterval)

		if c.items == nil {
			return
		}

		if keys := c.expiredKeys(); len(keys) != 0 {
			c.clearItems(keys)
		}
	}
}

func (c *Cache) expiredKeys() (keys []string) {

	c.mutex.RLock()
	for k, i := range c.items {
		if time.Now().UnixNano() > i.Expiration && i.Expiration > 0 {
			keys = append(keys, k)
		}
	}
	c.mutex.RUnlock()
	return
}

func (c *Cache) clearItems(keys []string) {

	c.mutex.Lock()
	for _, k := range keys {
		delete(c.items, k)
	}
	c.mutex.Unlock()
}
