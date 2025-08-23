package db

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// CacheItem 缓存项
type CacheItem struct {
	Value       interface{} `json:"value"`
	ExpiresAt   time.Time   `json:"expires_at"`
	Tags        []string    `json:"tags,omitempty"`
	AccessTime  int64       // 访问时间戳，用于LRU淘汰
	AccessCount int64       // 访问计数，用于LFU淘汰
}

// CacheShard 缓存分片
type CacheShard struct {
	items map[string]*CacheItem
	mutex sync.RWMutex
	tags  map[string][]string // tag -> keys mapping
	// 统计信息
	hits    int64
	misses  int64
	evicted int64
	expired int64
}

// CacheConfig 缓存配置
type CacheConfig struct {
	ShardCount      int            // 分片数量，默认为CPU核心数的2倍
	MaxSize         int            // 最大缓存项数量
	DefaultTTL      time.Duration  // 默认TTL
	CleanupInterval time.Duration  // 清理间隔
	EvictionPolicy  EvictionPolicy // 淘汰策略
}

// EvictionPolicy 淘汰策略
type EvictionPolicy int

const (
	EvictionPolicyLRU EvictionPolicy = iota // 最近最少使用
	EvictionPolicyLFU                       // 最少使用频率
	EvictionPolicyTTL                       // 仅基于TTL
)

// IsExpired 检查是否过期
func (item *CacheItem) IsExpired() bool {
	return !item.ExpiresAt.IsZero() && time.Now().After(item.ExpiresAt)
}

// HighConcurrencyMemoryCache 高并发内存缓存实现
type HighConcurrencyMemoryCache struct {
	shards      []*CacheShard
	shardCount  int
	config      *CacheConfig
	closer      chan struct{}
	cleanupOnce sync.Once

	// 全局统计信息（原子操作）
	totalHits    int64
	totalMisses  int64
	totalEvicted int64
	totalExpired int64
	totalSize    int64
}

// MemoryCache 保持原有接口兼容性
type MemoryCache = HighConcurrencyMemoryCache

// DefaultCacheConfig 默认缓存配置
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		ShardCount:      runtime.NumCPU() * 2, // CPU核心数的2倍
		MaxSize:         100000,               // 10万条记录
		DefaultTTL:      time.Hour,            // 默认1小时TTL
		CleanupInterval: time.Minute,          // 每分钟清理一次
		EvictionPolicy:  EvictionPolicyLRU,    // 默认LRU淘汰
	}
}

// NewMemoryCache 创建新的内存缓存实例（向后兼容）
func NewMemoryCache() *MemoryCache {
	return NewHighConcurrencyMemoryCache(DefaultCacheConfig())
}

// NewMemoryCacheWithConfig 使用自定义配置创建缓存实例
func NewMemoryCacheWithConfig(config *CacheConfig) *MemoryCache {
	return NewHighConcurrencyMemoryCache(config)
}

// NewHighConcurrencyMemoryCache 创建高并发内存缓存实例
func NewHighConcurrencyMemoryCache(config *CacheConfig) *HighConcurrencyMemoryCache {
	if config == nil {
		config = DefaultCacheConfig()
	}

	// 确保分片数量至少为1
	if config.ShardCount <= 0 {
		config.ShardCount = runtime.NumCPU() * 2
	}

	cache := &HighConcurrencyMemoryCache{
		shards:     make([]*CacheShard, config.ShardCount),
		shardCount: config.ShardCount,
		config:     config,
		closer:     make(chan struct{}),
	}

	// 初始化所有分片
	for i := 0; i < config.ShardCount; i++ {
		cache.shards[i] = &CacheShard{
			items: make(map[string]*CacheItem),
			tags:  make(map[string][]string),
		}
	}

	// 启动清理协程
	go cache.cleanup()

	return cache
}

// getShard 根据key获取对应的分片
func (c *HighConcurrencyMemoryCache) getShard(key string) *CacheShard {
	hash := fnv.New32a()
	hash.Write([]byte(key))
	return c.shards[hash.Sum32()%uint32(c.shardCount)]
}

// getShardIndex 根据key获取分片索引
func (c *HighConcurrencyMemoryCache) getShardIndex(key string) int {
	hash := fnv.New32a()
	hash.Write([]byte(key))
	return int(hash.Sum32() % uint32(c.shardCount))
}

// Get 获取缓存项
func (c *HighConcurrencyMemoryCache) Get(key string) (interface{}, error) {
	shard := c.getShard(key)
	shard.mutex.RLock()

	item, exists := shard.items[key]
	if !exists {
		shard.mutex.RUnlock()
		atomic.AddInt64(&shard.misses, 1)
		atomic.AddInt64(&c.totalMisses, 1)
		return nil, fmt.Errorf("cache miss")
	}

	if item.IsExpired() {
		shard.mutex.RUnlock()
		// 异步删除过期项
		go c.deleteExpiredItem(key)
		atomic.AddInt64(&shard.expired, 1)
		atomic.AddInt64(&c.totalExpired, 1)
		return nil, fmt.Errorf("cache expired")
	}

	// 更新访问统计（原子操作，无需写锁）
	now := time.Now().Unix()
	atomic.StoreInt64(&item.AccessTime, now)
	atomic.AddInt64(&item.AccessCount, 1)

	value := item.Value
	shard.mutex.RUnlock()

	atomic.AddInt64(&shard.hits, 1)
	atomic.AddInt64(&c.totalHits, 1)

	return value, nil
}

// deleteExpiredItem 异步删除过期项
func (c *HighConcurrencyMemoryCache) deleteExpiredItem(key string) {
	shard := c.getShard(key)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	if item, exists := shard.items[key]; exists && item.IsExpired() {
		c.removeItemFromShard(shard, key, item)
		atomic.AddInt64(&c.totalSize, -1)
	}
}

// Set 设置缓存项
func (c *HighConcurrencyMemoryCache) Set(key string, value interface{}, ttl time.Duration) error {
	return c.SetWithTags(key, value, ttl, nil)
}

// removeItemFromShard 从分片中移除项目（需要持有写锁）
func (c *HighConcurrencyMemoryCache) removeItemFromShard(shard *CacheShard, key string, item *CacheItem) {
	// 清理标签映射
	for _, tag := range item.Tags {
		if keys, exists := shard.tags[tag]; exists {
			newKeys := make([]string, 0, len(keys)-1)
			for _, k := range keys {
				if k != key {
					newKeys = append(newKeys, k)
				}
			}
			if len(newKeys) == 0 {
				delete(shard.tags, tag)
			} else {
				shard.tags[tag] = newKeys
			}
		}
	}
	delete(shard.items, key)
}

// addItemToShard 向分片添加项目（需要持有写锁）
func (c *HighConcurrencyMemoryCache) addItemToShard(shard *CacheShard, key string, item *CacheItem) {
	// 检查是否需要淘汰
	maxShardSize := c.config.MaxSize / c.shardCount
	if maxShardSize <= 0 {
		maxShardSize = 1000 // 默认值
	}

	// 如果达到容量限制，先淘汰旧项目
	for len(shard.items) >= maxShardSize {
		if !c.evictFromShard(shard, 1) {
			break // 如果无法淘汰，退出循环防止死循环
		}
	}

	shard.items[key] = item

	// 更新标签映射
	for _, tag := range item.Tags {
		if _, exists := shard.tags[tag]; !exists {
			shard.tags[tag] = make([]string, 0)
		}
		shard.tags[tag] = append(shard.tags[tag], key)
	}
}

// evictFromShard 从分片中淘汰指定数量的项目，返回是否成功淘汰
func (c *HighConcurrencyMemoryCache) evictFromShard(shard *CacheShard, count int) bool {
	if len(shard.items) == 0 {
		return false
	}

	evicted := 0
	now := time.Now().Unix()

	switch c.config.EvictionPolicy {
	case EvictionPolicyLRU:
		// LRU淘汰：找到最久未访问的项目
		for evicted < count && len(shard.items) > 0 {
			var oldestKey string
			var oldestTime int64 = now + 1 // 设为比当前时间更大的值

			for key, item := range shard.items {
				accessTime := atomic.LoadInt64(&item.AccessTime)
				if accessTime < oldestTime {
					oldestTime = accessTime
					oldestKey = key
				}
			}

			if oldestKey != "" {
				if item := shard.items[oldestKey]; item != nil {
					c.removeItemFromShard(shard, oldestKey, item)
					atomic.AddInt64(&shard.evicted, 1)
					atomic.AddInt64(&c.totalEvicted, 1)
					atomic.AddInt64(&c.totalSize, -1)
					evicted++
				}
			} else {
				break
			}
		}

	case EvictionPolicyLFU:
		// LFU淘汰：找到使用频率最低的项目
		for evicted < count && len(shard.items) > 0 {
			var leastUsedKey string
			var leastCount int64 = int64(^uint64(0) >> 1) // 最大int64值

			for key, item := range shard.items {
				accessCount := atomic.LoadInt64(&item.AccessCount)
				if accessCount < leastCount {
					leastCount = accessCount
					leastUsedKey = key
				}
			}

			if leastUsedKey != "" {
				if item := shard.items[leastUsedKey]; item != nil {
					c.removeItemFromShard(shard, leastUsedKey, item)
					atomic.AddInt64(&shard.evicted, 1)
					atomic.AddInt64(&c.totalEvicted, 1)
					atomic.AddInt64(&c.totalSize, -1)
					evicted++
				}
			} else {
				break
			}
		}

	case EvictionPolicyTTL:
		// TTL淘汰：随机选择一个项目淘汰
		for evicted < count && len(shard.items) > 0 {
			for key, item := range shard.items {
				c.removeItemFromShard(shard, key, item)
				atomic.AddInt64(&shard.evicted, 1)
				atomic.AddInt64(&c.totalEvicted, 1)
				atomic.AddInt64(&c.totalSize, -1)
				evicted++
				break // 只删除一个
			}
		}
	}

	return evicted > 0
}

// SetWithTags 设置带标签的缓存项
func (c *HighConcurrencyMemoryCache) SetWithTags(key string, value interface{}, ttl time.Duration, tags []string) error {
	shard := c.getShard(key)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	} else if c.config.DefaultTTL > 0 {
		expiresAt = time.Now().Add(c.config.DefaultTTL)
	}

	now := time.Now().Unix()
	item := &CacheItem{
		Value:       value,
		ExpiresAt:   expiresAt,
		Tags:        tags,
		AccessTime:  now,
		AccessCount: 1,
	}

	// 如果key已存在，先删除旧的
	if oldItem, exists := shard.items[key]; exists {
		c.removeItemFromShard(shard, key, oldItem)
		atomic.AddInt64(&c.totalSize, -1)
	}

	// 添加新项目
	c.addItemToShard(shard, key, item)
	atomic.AddInt64(&c.totalSize, 1)

	return nil
}

// Delete 删除缓存项
func (c *HighConcurrencyMemoryCache) Delete(key string) error {
	shard := c.getShard(key)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	if item, exists := shard.items[key]; exists {
		c.removeItemFromShard(shard, key, item)
		atomic.AddInt64(&c.totalSize, -1)
	}

	return nil
}

// DeleteByTags 根据标签删除缓存项
func (c *HighConcurrencyMemoryCache) DeleteByTags(tags []string) error {
	// 收集所有需要删除的key，按分片分组
	shardKeys := make(map[int][]string)

	// 遍历所有分片
	for i, shard := range c.shards {
		shard.mutex.RLock()
		for _, tag := range tags {
			if keys, exists := shard.tags[tag]; exists {
				if shardKeys[i] == nil {
					shardKeys[i] = make([]string, 0)
				}
				shardKeys[i] = append(shardKeys[i], keys...)
			}
		}
		shard.mutex.RUnlock()
	}

	// 按分片删除
	for shardIndex, keys := range shardKeys {
		shard := c.shards[shardIndex]
		shard.mutex.Lock()

		for _, key := range keys {
			if item, exists := shard.items[key]; exists {
				c.removeItemFromShard(shard, key, item)
				atomic.AddInt64(&c.totalSize, -1)
			}
		}

		shard.mutex.Unlock()
	}

	return nil
}

// Clear 清空所有缓存
func (c *HighConcurrencyMemoryCache) Clear() error {
	for _, shard := range c.shards {
		shard.mutex.Lock()
		shard.items = make(map[string]*CacheItem)
		shard.tags = make(map[string][]string)
		atomic.StoreInt64(&shard.hits, 0)
		atomic.StoreInt64(&shard.misses, 0)
		atomic.StoreInt64(&shard.evicted, 0)
		atomic.StoreInt64(&shard.expired, 0)
		shard.mutex.Unlock()
	}

	// 重置全局统计
	atomic.StoreInt64(&c.totalHits, 0)
	atomic.StoreInt64(&c.totalMisses, 0)
	atomic.StoreInt64(&c.totalEvicted, 0)
	atomic.StoreInt64(&c.totalExpired, 0)
	atomic.StoreInt64(&c.totalSize, 0)

	return nil
}

// Has 检查缓存是否存在
func (c *HighConcurrencyMemoryCache) Has(key string) (bool, error) {
	shard := c.getShard(key)
	shard.mutex.RLock()
	defer shard.mutex.RUnlock()

	item, exists := shard.items[key]
	if !exists {
		return false, nil
	}

	if item.IsExpired() {
		return false, nil
	}

	return true, nil
}

// Size 获取缓存大小
func (c *HighConcurrencyMemoryCache) Size() int {
	return int(atomic.LoadInt64(&c.totalSize))
}

// Stats 获取缓存统计信息
func (c *HighConcurrencyMemoryCache) Stats() map[string]interface{} {
	stats := map[string]interface{}{
		"total_items":     atomic.LoadInt64(&c.totalSize),
		"total_hits":      atomic.LoadInt64(&c.totalHits),
		"total_misses":    atomic.LoadInt64(&c.totalMisses),
		"total_evicted":   atomic.LoadInt64(&c.totalEvicted),
		"total_expired":   atomic.LoadInt64(&c.totalExpired),
		"shard_count":     c.shardCount,
		"max_size":        c.config.MaxSize,
		"eviction_policy": c.config.EvictionPolicy,
	}

	// 计算命中率
	hits := atomic.LoadInt64(&c.totalHits)
	misses := atomic.LoadInt64(&c.totalMisses)
	if hits+misses > 0 {
		stats["hit_rate"] = float64(hits) / float64(hits+misses)
	} else {
		stats["hit_rate"] = 0.0
	}

	// 分片统计
	shardStats := make([]map[string]interface{}, c.shardCount)
	totalExpired := int64(0)
	totalTags := int64(0)

	for i, shard := range c.shards {
		shard.mutex.RLock()
		expired := int64(0)
		for _, item := range shard.items {
			if item.IsExpired() {
				expired++
			}
		}
		totalExpired += expired
		totalTags += int64(len(shard.tags))

		shardStats[i] = map[string]interface{}{
			"items":   len(shard.items),
			"expired": expired,
			"tags":    len(shard.tags),
			"hits":    atomic.LoadInt64(&shard.hits),
			"misses":  atomic.LoadInt64(&shard.misses),
			"evicted": atomic.LoadInt64(&shard.evicted),
		}
		shard.mutex.RUnlock()
	}

	stats["expired_items"] = totalExpired
	stats["total_tags"] = totalTags
	stats["shard_stats"] = shardStats

	return stats
}

// Close 关闭缓存
func (c *HighConcurrencyMemoryCache) Close() error {
	c.cleanupOnce.Do(func() {
		close(c.closer)
	})
	return c.Clear()
}

// cleanup 清理过期项
func (c *HighConcurrencyMemoryCache) cleanup() {
	ticker := time.NewTicker(c.config.CleanupInterval)
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
func (c *HighConcurrencyMemoryCache) cleanupExpired() {
	// 并发清理各个分片
	var wg sync.WaitGroup

	for i, shard := range c.shards {
		wg.Add(1)
		go func(shardIndex int, s *CacheShard) {
			defer wg.Done()
			c.cleanupShardExpired(s)
		}(i, shard)
	}

	wg.Wait()
}

// cleanupShardExpired 清理分片中的过期项
func (c *HighConcurrencyMemoryCache) cleanupShardExpired(shard *CacheShard) {
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	expiredKeys := make([]string, 0)

	for key, item := range shard.items {
		if item.IsExpired() {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		if item, exists := shard.items[key]; exists {
			c.removeItemFromShard(shard, key, item)
			atomic.AddInt64(&shard.expired, 1)
			atomic.AddInt64(&c.totalExpired, 1)
			atomic.AddInt64(&c.totalSize, -1)
		}
	}
}

// 高级功能方法

// GetMulti 批量获取缓存项
func (c *HighConcurrencyMemoryCache) GetMulti(keys []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	if len(keys) == 0 {
		return result, nil
	}

	// 按分片分组key
	shardKeys := make(map[int][]string)
	for _, key := range keys {
		shardIndex := c.getShardIndex(key)
		if shardKeys[shardIndex] == nil {
			shardKeys[shardIndex] = make([]string, 0)
		}
		shardKeys[shardIndex] = append(shardKeys[shardIndex], key)
	}

	// 并发获取各分片数据
	var mu sync.Mutex
	var wg sync.WaitGroup

	for shardIndex, keys := range shardKeys {
		wg.Add(1)
		go func(sIndex int, sKeys []string) {
			defer wg.Done()
			shard := c.shards[sIndex]
			shard.mutex.RLock()
			defer shard.mutex.RUnlock()

			for _, key := range sKeys {
				if item, exists := shard.items[key]; exists && !item.IsExpired() {
					// 更新访问统计
					now := time.Now().Unix()
					atomic.StoreInt64(&item.AccessTime, now)
					atomic.AddInt64(&item.AccessCount, 1)

					mu.Lock()
					result[key] = item.Value
					mu.Unlock()

					atomic.AddInt64(&shard.hits, 1)
					atomic.AddInt64(&c.totalHits, 1)
				} else {
					atomic.AddInt64(&shard.misses, 1)
					atomic.AddInt64(&c.totalMisses, 1)
				}
			}
		}(shardIndex, keys)
	}

	wg.Wait()
	return result, nil
}

// SetMulti 批量设置缓存项
func (c *HighConcurrencyMemoryCache) SetMulti(data map[string]interface{}, ttl time.Duration) error {
	if len(data) == 0 {
		return nil
	}

	// 按分片分组数据
	shardData := make(map[int]map[string]interface{})
	for key, value := range data {
		shardIndex := c.getShardIndex(key)
		if shardData[shardIndex] == nil {
			shardData[shardIndex] = make(map[string]interface{})
		}
		shardData[shardIndex][key] = value
	}

	// 并发设置各分片数据
	var wg sync.WaitGroup

	for shardIndex, data := range shardData {
		wg.Add(1)
		go func(sIndex int, sData map[string]interface{}) {
			defer wg.Done()

			for key, value := range sData {
				c.Set(key, value, ttl)
			}
		}(shardIndex, data)
	}

	wg.Wait()
	return nil
}

// GetOrSet 获取缓存，如果不存在则设置
func (c *HighConcurrencyMemoryCache) GetOrSet(key string, valueFunc func() (interface{}, error), ttl time.Duration) (interface{}, error) {
	// 先尝试获取
	if value, err := c.Get(key); err == nil {
		return value, nil
	}

	// 不存在则获取新值并设置
	value, err := valueFunc()
	if err != nil {
		return nil, err
	}

	c.Set(key, value, ttl)
	return value, nil
}

// Increment 递增数值（原子操作）
func (c *HighConcurrencyMemoryCache) Increment(key string, delta int64) (int64, error) {
	shard := c.getShard(key)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	item, exists := shard.items[key]
	if !exists {
		return 0, fmt.Errorf("key not found")
	}

	if item.IsExpired() {
		c.removeItemFromShard(shard, key, item)
		atomic.AddInt64(&c.totalSize, -1)
		return 0, fmt.Errorf("key expired")
	}

	// 尝试转换为int64
	var newValue int64
	switch v := item.Value.(type) {
	case int64:
		newValue = v + delta
	case int:
		newValue = int64(v) + delta
	case int32:
		newValue = int64(v) + delta
	default:
		return 0, fmt.Errorf("value is not numeric")
	}

	item.Value = newValue
	atomic.StoreInt64(&item.AccessTime, time.Now().Unix())
	atomic.AddInt64(&item.AccessCount, 1)

	return newValue, nil
}

// Touch 更新缓存项的过期时间
func (c *HighConcurrencyMemoryCache) Touch(key string, ttl time.Duration) error {
	shard := c.getShard(key)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	item, exists := shard.items[key]
	if !exists {
		return fmt.Errorf("key not found")
	}

	if ttl > 0 {
		item.ExpiresAt = time.Now().Add(ttl)
	} else {
		item.ExpiresAt = time.Time{}
	}

	atomic.StoreInt64(&item.AccessTime, time.Now().Unix())
	return nil
}

// Decrement 递减数值
func (c *HighConcurrencyMemoryCache) Decrement(key string, delta int64) (int64, error) {
	return c.Increment(key, -delta)
}

// Expire 设置缓存项的过期时间
func (c *HighConcurrencyMemoryCache) Expire(key string, ttl time.Duration) error {
	return c.Touch(key, ttl)
}

// TTL 获取缓存项的剩余生存时间
func (c *HighConcurrencyMemoryCache) TTL(key string) (time.Duration, error) {
	shard := c.getShard(key)
	shard.mutex.RLock()
	defer shard.mutex.RUnlock()

	item, exists := shard.items[key]
	if !exists {
		return 0, fmt.Errorf("key not found")
	}

	if item.ExpiresAt.IsZero() {
		return -1, nil // 永不过期
	}

	if item.IsExpired() {
		return 0, fmt.Errorf("key expired")
	}

	return time.Until(item.ExpiresAt), nil
}

// DeleteMulti 批量删除缓存项
func (c *HighConcurrencyMemoryCache) DeleteMulti(keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	// 按分片分组key
	shardKeys := make(map[int][]string)
	for _, key := range keys {
		shardIndex := c.getShardIndex(key)
		if shardKeys[shardIndex] == nil {
			shardKeys[shardIndex] = make([]string, 0)
		}
		shardKeys[shardIndex] = append(shardKeys[shardIndex], key)
	}

	// 并发删除各分片数据
	var wg sync.WaitGroup

	for shardIndex, keys := range shardKeys {
		wg.Add(1)
		go func(sIndex int, sKeys []string) {
			defer wg.Done()
			shard := c.shards[sIndex]
			shard.mutex.Lock()
			defer shard.mutex.Unlock()

			for _, key := range sKeys {
				if item, exists := shard.items[key]; exists {
					c.removeItemFromShard(shard, key, item)
					atomic.AddInt64(&c.totalSize, -1)
				}
			}
		}(shardIndex, keys)
	}

	wg.Wait()
	return nil
}

// ResetStats 重置统计信息
func (c *HighConcurrencyMemoryCache) ResetStats() error {
	for _, shard := range c.shards {
		atomic.StoreInt64(&shard.hits, 0)
		atomic.StoreInt64(&shard.misses, 0)
		atomic.StoreInt64(&shard.evicted, 0)
		atomic.StoreInt64(&shard.expired, 0)
	}

	atomic.StoreInt64(&c.totalHits, 0)
	atomic.StoreInt64(&c.totalMisses, 0)
	atomic.StoreInt64(&c.totalEvicted, 0)
	atomic.StoreInt64(&c.totalExpired, 0)

	return nil
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
