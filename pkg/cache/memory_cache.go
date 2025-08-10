package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/zhoudm1743/torm/pkg/db"
)

// CacheItem 缓存项
type CacheItem struct {
	Value     interface{}
	ExpiresAt time.Time
}

// IsExpired 检查是否过期
func (item *CacheItem) IsExpired() bool {
	return !item.ExpiresAt.IsZero() && time.Now().After(item.ExpiresAt)
}

// MemoryCache 内存缓存实现
type MemoryCache struct {
	items map[string]*CacheItem
	mu    sync.RWMutex
}

// NewMemoryCache 创建内存缓存
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		items: make(map[string]*CacheItem),
	}

	// 启动清理协程
	go cache.cleanup()

	return cache
}

// Get 获取缓存值
func (c *MemoryCache) Get(key string) (interface{}, error) {
	c.mu.RLock()
	item, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("cache key '%s' not found", key)
	}

	if item.IsExpired() {
		c.Delete(key)
		return nil, fmt.Errorf("cache key '%s' has expired", key)
	}

	return item.Value, nil
}

// Set 设置缓存值
func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	c.items[key] = &CacheItem{
		Value:     value,
		ExpiresAt: expiresAt,
	}

	return nil
}

// Delete 删除缓存值
func (c *MemoryCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	return nil
}

// Clear 清空所有缓存
func (c *MemoryCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*CacheItem)
	return nil
}

// Has 检查缓存键是否存在
func (c *MemoryCache) Has(key string) (bool, error) {
	c.mu.RLock()
	item, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		return false, nil
	}

	if item.IsExpired() {
		c.Delete(key)
		return false, nil
	}

	return true, nil
}

// cleanup 定期清理过期的缓存项
func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		for key, item := range c.items {
			if item.IsExpired() {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

// Size 获取缓存项数量
func (c *MemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Keys 获取所有缓存键
func (c *MemoryCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.items))
	for key, item := range c.items {
		if !item.IsExpired() {
			keys = append(keys, key)
		}
	}
	return keys
}

// 确保 MemoryCache 实现了 CacheInterface 接口
var _ db.CacheInterface = (*MemoryCache)(nil)
