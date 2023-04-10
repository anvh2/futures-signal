package basic

import (
	"sync"
)

type Cache struct {
	mutex    *sync.RWMutex
	internal map[string]interface{}
}

func NewCache() *Cache {
	return &Cache{
		mutex:    &sync.RWMutex{},
		internal: make(map[string]interface{}),
	}
}

func (c *Cache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.internal[key] = value
}

func (c *Cache) Get(key string) interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.internal[key]
}

// SetEX sets and check exists
func (c *Cache) SetEX(key string, value interface{}) (interface{}, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if old, ok := c.internal[key]; ok {
		c.internal[key] = value
		return old, true
	}

	c.internal[key] = value
	return value, false
}
