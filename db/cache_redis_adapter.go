package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

// RedisInterface Redis操作接口（用户需要实现）
type RedisInterface interface {
	// 基础操作
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (int64, error)

	// 批量操作
	MGet(ctx context.Context, keys ...string) ([]interface{}, error)
	MSet(ctx context.Context, values ...interface{}) error

	// 高级操作
	Incr(ctx context.Context, key string) (int64, error)
	IncrBy(ctx context.Context, key string, value int64) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)

	// 标签操作（基于Set实现）
	SAdd(ctx context.Context, key string, members ...interface{}) error
	SMembers(ctx context.Context, key string) ([]string, error)
	SRem(ctx context.Context, key string, members ...interface{}) error

	// 管道操作
	Pipeline() RedisPipeliner

	// 连接管理
	Close() error
	Ping(ctx context.Context) error
}

// RedisPipeliner Redis管道接口
type RedisPipeliner interface {
	Get(ctx context.Context, key string)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration)
	Del(ctx context.Context, keys ...string)
	Exec(ctx context.Context) ([]interface{}, error)
}

// RedisConfig Redis缓存配置
type RedisConfig struct {
	// Redis连接信息
	Address     string        `json:"address"`
	Password    string        `json:"password"`
	DB          int           `json:"db"`
	MaxRetries  int           `json:"max_retries"`
	PoolSize    int           `json:"pool_size"`
	DialTimeout time.Duration `json:"dial_timeout"`

	// 缓存配置
	KeyPrefix   string        `json:"key_prefix"`
	DefaultTTL  time.Duration `json:"default_ttl"`
	TagsEnabled bool          `json:"tags_enabled"`

	// 序列化配置
	Serializer string `json:"serializer"` // "json", "msgpack", "gob"
}

// RedisCache Redis缓存实现
type RedisCache struct {
	client RedisInterface
	config *RedisConfig
	ctx    context.Context

	// 统计信息
	hits   int64
	misses int64
	errors int64
}

// RedisCacheProvider Redis缓存提供者
type RedisCacheProvider struct {
	ClientFactory func(*RedisConfig) (RedisInterface, error)
}

// CreateCache 创建Redis缓存实例
func (p *RedisCacheProvider) CreateCache(config interface{}) (FullCacheInterface, error) {
	redisConfig, ok := config.(*RedisConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type for Redis cache, expected *RedisConfig")
	}

	if p.ClientFactory == nil {
		return nil, fmt.Errorf("Redis client factory not provided")
	}

	client, err := p.ClientFactory(redisConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	cache := &RedisCache{
		client: client,
		config: redisConfig,
		ctx:    context.Background(),
	}

	// 测试连接
	if err := client.Ping(cache.ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping Redis server: %w", err)
	}

	return cache, nil
}

// ValidateConfig 验证配置
func (p *RedisCacheProvider) ValidateConfig(config interface{}) error {
	redisConfig, ok := config.(*RedisConfig)
	if !ok {
		return fmt.Errorf("invalid config type for Redis cache, expected *RedisConfig")
	}

	if redisConfig.Address == "" {
		return fmt.Errorf("Redis address is required")
	}

	if redisConfig.DialTimeout <= 0 {
		redisConfig.DialTimeout = 5 * time.Second
	}

	if redisConfig.DefaultTTL <= 0 {
		redisConfig.DefaultTTL = time.Hour
	}

	if redisConfig.Serializer == "" {
		redisConfig.Serializer = "json"
	}

	return nil
}

// GetConfigExample 获取配置示例
func (p *RedisCacheProvider) GetConfigExample() interface{} {
	return &RedisConfig{
		Address:     "localhost:6379",
		Password:    "",
		DB:          0,
		MaxRetries:  3,
		PoolSize:    10,
		DialTimeout: 5 * time.Second,
		KeyPrefix:   "torm:cache:",
		DefaultTTL:  time.Hour,
		TagsEnabled: true,
		Serializer:  "json",
	}
}

// Redis缓存方法实现

// buildKey 构建Redis键
func (c *RedisCache) buildKey(key string) string {
	return c.config.KeyPrefix + key
}

// buildTagKey 构建标签键
func (c *RedisCache) buildTagKey(tag string) string {
	return c.config.KeyPrefix + "tags:" + tag
}

// serialize 序列化值
func (c *RedisCache) serialize(value interface{}) (string, error) {
	switch c.config.Serializer {
	case "json":
		data, err := json.Marshal(value)
		return string(data), err
	default:
		// 简单的字符串转换
		return fmt.Sprintf("%v", value), nil
	}
}

// deserialize 反序列化值
func (c *RedisCache) deserialize(data string) (interface{}, error) {
	switch c.config.Serializer {
	case "json":
		var value interface{}
		err := json.Unmarshal([]byte(data), &value)
		return value, err
	default:
		return data, nil
	}
}

// Get 获取缓存项
func (c *RedisCache) Get(key string) (interface{}, error) {
	redisKey := c.buildKey(key)

	result, err := c.client.Get(c.ctx, redisKey)
	if err != nil {
		atomic.AddInt64(&c.errors, 1)
		if strings.Contains(err.Error(), "nil") || strings.Contains(err.Error(), "not found") {
			atomic.AddInt64(&c.misses, 1)
			return nil, fmt.Errorf("cache miss")
		}
		return nil, err
	}

	if result == "" {
		atomic.AddInt64(&c.misses, 1)
		return nil, fmt.Errorf("cache miss")
	}

	atomic.AddInt64(&c.hits, 1)
	return c.deserialize(result)
}

// Set 设置缓存项
func (c *RedisCache) Set(key string, value interface{}, ttl time.Duration) error {
	redisKey := c.buildKey(key)

	if ttl <= 0 {
		ttl = c.config.DefaultTTL
	}

	serialized, err := c.serialize(value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %w", err)
	}

	err = c.client.Set(c.ctx, redisKey, serialized, ttl)
	if err != nil {
		atomic.AddInt64(&c.errors, 1)
		return err
	}

	return nil
}

// SetWithTags 设置带标签的缓存项
func (c *RedisCache) SetWithTags(key string, value interface{}, ttl time.Duration, tags []string) error {
	// 首先设置缓存项
	if err := c.Set(key, value, ttl); err != nil {
		return err
	}

	// 如果启用了标签且有标签，则更新标签映射
	if c.config.TagsEnabled && len(tags) > 0 {
		pipe := c.client.Pipeline()

		for _, tag := range tags {
			tagKey := c.buildTagKey(tag)
			pipe.Set(c.ctx, tagKey, key, ttl) // 标签键也设置相同的TTL
		}

		_, err := pipe.Exec(c.ctx)
		if err != nil {
			atomic.AddInt64(&c.errors, 1)
			return fmt.Errorf("failed to update tags: %w", err)
		}
	}

	return nil
}

// Delete 删除缓存项
func (c *RedisCache) Delete(key string) error {
	redisKey := c.buildKey(key)

	err := c.client.Del(c.ctx, redisKey)
	if err != nil {
		atomic.AddInt64(&c.errors, 1)
		return err
	}

	return nil
}

// DeleteByTags 根据标签删除缓存项
func (c *RedisCache) DeleteByTags(tags []string) error {
	if !c.config.TagsEnabled || len(tags) == 0 {
		return nil
	}

	// 获取所有标签对应的键
	var keysToDelete []string

	for _, tag := range tags {
		tagKey := c.buildTagKey(tag)
		members, err := c.client.SMembers(c.ctx, tagKey)
		if err != nil {
			atomic.AddInt64(&c.errors, 1)
			continue
		}

		for _, member := range members {
			keysToDelete = append(keysToDelete, c.buildKey(member))
		}

		// 删除标签键
		keysToDelete = append(keysToDelete, tagKey)
	}

	if len(keysToDelete) > 0 {
		err := c.client.Del(c.ctx, keysToDelete...)
		if err != nil {
			atomic.AddInt64(&c.errors, 1)
			return err
		}
	}

	return nil
}

// Clear 清空所有缓存（注意：这会删除所有以KeyPrefix开头的键）
func (c *RedisCache) Clear() error {
	// Redis没有直接的清空前缀键的方法，这里只是一个示例实现
	// 实际使用中可能需要使用SCAN命令或者其他方式
	return fmt.Errorf("Clear operation not implemented for Redis cache")
}

// Has 检查缓存是否存在
func (c *RedisCache) Has(key string) (bool, error) {
	redisKey := c.buildKey(key)

	count, err := c.client.Exists(c.ctx, redisKey)
	if err != nil {
		atomic.AddInt64(&c.errors, 1)
		return false, err
	}

	return count > 0, nil
}

// Size 获取缓存大小（Redis实现较复杂，这里返回-1表示不支持）
func (c *RedisCache) Size() int {
	return -1 // Redis实现获取大小较复杂
}

// GetMulti 批量获取缓存项
func (c *RedisCache) GetMulti(keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	// 构建Redis键
	redisKeys := make([]string, len(keys))
	for i, key := range keys {
		redisKeys[i] = c.buildKey(key)
	}

	results, err := c.client.MGet(c.ctx, redisKeys...)
	if err != nil {
		atomic.AddInt64(&c.errors, 1)
		return nil, err
	}

	result := make(map[string]interface{})
	for i, data := range results {
		if data != nil && data != "" {
			if value, err := c.deserialize(fmt.Sprintf("%v", data)); err == nil {
				result[keys[i]] = value
				atomic.AddInt64(&c.hits, 1)
			}
		} else {
			atomic.AddInt64(&c.misses, 1)
		}
	}

	return result, nil
}

// SetMulti 批量设置缓存项
func (c *RedisCache) SetMulti(data map[string]interface{}, ttl time.Duration) error {
	if len(data) == 0 {
		return nil
	}

	if ttl <= 0 {
		ttl = c.config.DefaultTTL
	}

	pipe := c.client.Pipeline()

	for key, value := range data {
		redisKey := c.buildKey(key)
		serialized, err := c.serialize(value)
		if err != nil {
			return fmt.Errorf("failed to serialize value for key %s: %w", key, err)
		}
		pipe.Set(c.ctx, redisKey, serialized, ttl)
	}

	_, err := pipe.Exec(c.ctx)
	if err != nil {
		atomic.AddInt64(&c.errors, 1)
		return err
	}

	return nil
}

// DeleteMulti 批量删除缓存项
func (c *RedisCache) DeleteMulti(keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	redisKeys := make([]string, len(keys))
	for i, key := range keys {
		redisKeys[i] = c.buildKey(key)
	}

	err := c.client.Del(c.ctx, redisKeys...)
	if err != nil {
		atomic.AddInt64(&c.errors, 1)
		return err
	}

	return nil
}

// GetOrSet 获取缓存，如果不存在则设置
func (c *RedisCache) GetOrSet(key string, valueFunc func() (interface{}, error), ttl time.Duration) (interface{}, error) {
	// 先尝试获取
	if value, err := c.Get(key); err == nil {
		return value, nil
	}

	// 不存在则获取新值并设置
	value, err := valueFunc()
	if err != nil {
		return nil, err
	}

	if err := c.Set(key, value, ttl); err != nil {
		return value, err // 返回值但记录设置错误
	}

	return value, nil
}

// Increment 递增数值
func (c *RedisCache) Increment(key string, delta int64) (int64, error) {
	redisKey := c.buildKey(key)

	result, err := c.client.IncrBy(c.ctx, redisKey, delta)
	if err != nil {
		atomic.AddInt64(&c.errors, 1)
		return 0, err
	}

	return result, nil
}

// Decrement 递减数值
func (c *RedisCache) Decrement(key string, delta int64) (int64, error) {
	return c.Increment(key, -delta)
}

// Touch 更新缓存项的过期时间
func (c *RedisCache) Touch(key string, ttl time.Duration) error {
	redisKey := c.buildKey(key)

	err := c.client.Expire(c.ctx, redisKey, ttl)
	if err != nil {
		atomic.AddInt64(&c.errors, 1)
		return err
	}

	return nil
}

// Expire 设置缓存项的过期时间
func (c *RedisCache) Expire(key string, ttl time.Duration) error {
	return c.Touch(key, ttl)
}

// TTL 获取缓存项的剩余生存时间
func (c *RedisCache) TTL(key string) (time.Duration, error) {
	redisKey := c.buildKey(key)

	ttl, err := c.client.TTL(c.ctx, redisKey)
	if err != nil {
		atomic.AddInt64(&c.errors, 1)
		return 0, err
	}

	return ttl, nil
}

// Stats 获取缓存统计信息
func (c *RedisCache) Stats() map[string]interface{} {
	return map[string]interface{}{
		"type":        "redis",
		"hits":        atomic.LoadInt64(&c.hits),
		"misses":      atomic.LoadInt64(&c.misses),
		"errors":      atomic.LoadInt64(&c.errors),
		"hit_rate":    c.calculateHitRate(),
		"key_prefix":  c.config.KeyPrefix,
		"default_ttl": c.config.DefaultTTL.String(),
		"address":     c.config.Address,
		"db":          c.config.DB,
	}
}

// calculateHitRate 计算命中率
func (c *RedisCache) calculateHitRate() float64 {
	hits := atomic.LoadInt64(&c.hits)
	misses := atomic.LoadInt64(&c.misses)

	if hits+misses == 0 {
		return 0.0
	}

	return float64(hits) / float64(hits+misses)
}

// ResetStats 重置统计信息
func (c *RedisCache) ResetStats() error {
	atomic.StoreInt64(&c.hits, 0)
	atomic.StoreInt64(&c.misses, 0)
	atomic.StoreInt64(&c.errors, 0)
	return nil
}

// Close 关闭缓存
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// NewRedisCacheProvider 创建Redis缓存提供者
func NewRedisCacheProvider(clientFactory func(*RedisConfig) (RedisInterface, error)) CacheProvider {
	return &RedisCacheProvider{
		ClientFactory: clientFactory,
	}
}
