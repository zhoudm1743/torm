package db

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CacheItem 缓存项
type CacheItem struct {
	Value     interface{} `json:"value"`
	ExpiresAt time.Time   `json:"expires_at"`
	Tags      []string    `json:"tags,omitempty"`
}

// IsExpired 检查是否过期
func (item *CacheItem) IsExpired() bool {
	return !item.ExpiresAt.IsZero() && time.Now().After(item.ExpiresAt)
}

// MemoryCache 内存缓存实现
type MemoryCache struct {
	items  map[string]*CacheItem
	mutex  sync.RWMutex
	tags   map[string][]string // tag -> keys mapping
	closer chan struct{}
}

// NewMemoryCache 创建新的内存缓存实例
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		items:  make(map[string]*CacheItem),
		tags:   make(map[string][]string),
		closer: make(chan struct{}),
	}

	// 启动清理协程
	go cache.cleanup()

	return cache
}

// Get 获取缓存项
func (c *MemoryCache) Get(key string) (interface{}, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, fmt.Errorf("cache miss")
	}

	if item.IsExpired() {
		// 异步删除过期项
		go func() {
			c.mutex.Lock()
			delete(c.items, key)
			c.mutex.Unlock()
		}()
		return nil, fmt.Errorf("cache expired")
	}

	return item.Value, nil
}

// Set 设置缓存项
func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

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

// SetWithTags 设置带标签的缓存项
func (c *MemoryCache) SetWithTags(key string, value interface{}, ttl time.Duration, tags []string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	c.items[key] = &CacheItem{
		Value:     value,
		ExpiresAt: expiresAt,
		Tags:      tags,
	}

	// 更新标签映射
	for _, tag := range tags {
		if _, exists := c.tags[tag]; !exists {
			c.tags[tag] = make([]string, 0)
		}
		c.tags[tag] = append(c.tags[tag], key)
	}

	return nil
}

// Delete 删除缓存项
func (c *MemoryCache) Delete(key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 删除项
	if item, exists := c.items[key]; exists {
		// 清理标签映射
		for _, tag := range item.Tags {
			if keys, exists := c.tags[tag]; exists {
				newKeys := make([]string, 0, len(keys)-1)
				for _, k := range keys {
					if k != key {
						newKeys = append(newKeys, k)
					}
				}
				if len(newKeys) == 0 {
					delete(c.tags, tag)
				} else {
					c.tags[tag] = newKeys
				}
			}
		}
		delete(c.items, key)
	}

	return nil
}

// DeleteByTags 根据标签删除缓存项
func (c *MemoryCache) DeleteByTags(tags []string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	keysToDelete := make(map[string]bool)

	for _, tag := range tags {
		if keys, exists := c.tags[tag]; exists {
			for _, key := range keys {
				keysToDelete[key] = true
			}
			delete(c.tags, tag)
		}
	}

	for key := range keysToDelete {
		if item, exists := c.items[key]; exists {
			// 清理其他标签映射中的这个key
			for _, itemTag := range item.Tags {
				if keys, exists := c.tags[itemTag]; exists {
					newKeys := make([]string, 0, len(keys)-1)
					for _, k := range keys {
						if k != key {
							newKeys = append(newKeys, k)
						}
					}
					if len(newKeys) == 0 {
						delete(c.tags, itemTag)
					} else {
						c.tags[itemTag] = newKeys
					}
				}
			}
			delete(c.items, key)
		}
	}

	return nil
}

// Clear 清空所有缓存
func (c *MemoryCache) Clear() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]*CacheItem)
	c.tags = make(map[string][]string)

	return nil
}

// Has 检查缓存是否存在
func (c *MemoryCache) Has(key string) (bool, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return false, nil
	}

	if item.IsExpired() {
		return false, nil
	}

	return true, nil
}

// Size 获取缓存大小
func (c *MemoryCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.items)
}

// Stats 获取缓存统计信息
func (c *MemoryCache) Stats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	expired := 0
	for _, item := range c.items {
		if item.IsExpired() {
			expired++
		}
	}

	return map[string]interface{}{
		"total_items":   len(c.items),
		"expired_items": expired,
		"total_tags":    len(c.tags),
	}
}

// Close 关闭缓存
func (c *MemoryCache) Close() error {
	close(c.closer)
	return c.Clear()
}

// cleanup 清理过期项
func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute) // 每分钟清理一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanupExpired()
		case <-c.closer:
			return
		}
	}
}

// cleanupExpired 清理过期项
func (c *MemoryCache) cleanupExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	expiredKeys := make([]string, 0)

	for key, item := range c.items {
		if item.IsExpired() {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		if item, exists := c.items[key]; exists {
			// 清理标签映射
			for _, tag := range item.Tags {
				if keys, exists := c.tags[tag]; exists {
					newKeys := make([]string, 0, len(keys)-1)
					for _, k := range keys {
						if k != key {
							newKeys = append(newKeys, k)
						}
					}
					if len(newKeys) == 0 {
						delete(c.tags, tag)
					} else {
						c.tags[tag] = newKeys
					}
				}
			}
			delete(c.items, key)
		}
	}
}

// GenerateCacheKey 生成缓存键
func GenerateCacheKey(prefix string, data interface{}) string {
	jsonData, _ := json.Marshal(data)
	hash := md5.Sum(jsonData)
	return fmt.Sprintf("%s:%x", prefix, hash)
}

// 默认缓存实例
var defaultCache *MemoryCache

func init() {
	defaultCache = NewMemoryCache()
}

// GetDefaultCache 获取默认缓存实例
func GetDefaultCache() CacheInterface {
	return defaultCache
}
