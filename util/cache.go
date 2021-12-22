package util

import (
	"sync"
)

// Cache is a data store
type Cache struct {
	sync.Mutex
	val map[string]Param
}

// NewCache creates cache
func NewCache() *Cache {
	return &Cache{
		val: make(map[string]Param),
	}
}

// Run adds input channel's values to cache
func (c *Cache) Run(in <-chan Param) {
	log := NewLogger("cache")

	for p := range in {
		log.DEBUG.Printf("%s: %v", p.Key, p.Val)
		c.Add(p.UniqueID(), p)
	}
}

// All provides a copy of the cached values
func (c *Cache) All() []Param {
	c.Lock()
	defer c.Unlock()

	copy := make([]Param, 0, len(c.val))
	for _, val := range c.val {
		copy = append(copy, val)
	}

	return copy
}

// Add entry to cache
func (c *Cache) Add(key string, param Param) {
	c.Lock()
	defer c.Unlock()

	c.val[key] = param
}

// Get entry from cache
func (c *Cache) Get(key string) Param {
	c.Lock()
	defer c.Unlock()

	if val, ok := c.val[key]; ok {
		return val
	}

	return Param{}
}
