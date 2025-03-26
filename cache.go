package imagecapture

import (
	"sync"
	"time"
)

type Cache struct {
	data     map[string]cacheItem
	mutex    sync.RWMutex
	bloom    *BloomFilter
	maxSize  int
	lifetime time.Duration
}

type cacheItem struct {
	value     interface{}
	timestamp time.Time
}

func NewCache(maxSize int, lifetime time.Duration) *Cache {
	return &Cache{
		data:     make(map[string]cacheItem),
		bloom:    NewBloomFilter(uint(maxSize), 0.01),
		maxSize:  maxSize,
		lifetime: lifetime,
	}
}

func (c *Cache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 如果缓存已满，清理过期项
	if len(c.data) >= c.maxSize {
		c.cleanup()
	}

	c.data[key] = cacheItem{
		value:     value,
		timestamp: time.Now(),
	}
	c.bloom.Add(key)
}

func (c *Cache) Get(key string) (interface{}, bool) {
	// 快速检查key是否可能存在
	if !c.bloom.Contains(key) {
		return nil, false
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Since(item.timestamp) > c.lifetime {
		return nil, false
	}

	return item.value, true
}

func (c *Cache) cleanup() {
	now := time.Now()
	for key, item := range c.data {
		if now.Sub(item.timestamp) > c.lifetime {
			delete(c.data, key)
		}
	}
}

func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data = make(map[string]cacheItem)
	c.bloom = NewBloomFilter(uint(c.maxSize), 0.01)
}
