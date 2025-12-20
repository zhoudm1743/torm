# 缓存系统

TORM 提供了完整的缓存系统，支持内存缓存、Redis 缓存和自定义缓存实现。缓存系统与查询构建器无缝集成，支持自动缓存键生成、TTL 管理和标签清理。

## 📋 目录

- [快速开始](#快速开始)
- [缓存类型](#缓存类型)
- [查询缓存](#查询缓存)
- [标签缓存](#标签缓存)
- [缓存管理](#缓存管理)
- [内存缓存](#内存缓存)
- [Redis 缓存](#redis-缓存)
- [自定义缓存](#自定义缓存)
- [最佳实践](#最佳实践)

## 🚀 快速开始

### 基础查询缓存

```go
import (
    "time"
    "github.com/zhoudm1743/torm"
)

// 缓存查询结果 5 分钟
users, err := torm.Table("users", "default").
    Where("status", "=", "active").
    Cache(5 * time.Minute).
    Get()

// 第二次相同查询从缓存获取
users2, err := torm.Table("users", "default").
    Where("status", "=", "active").
    Cache(5 * time.Minute).
    Get()
```

## 🔄 缓存类型

### 内存缓存（默认）

- 高性能、低延迟
- 数据存储在应用内存中
- 适用于单机部署

### Redis 缓存

- 分布式缓存
- 数据持久化
- 适用于集群部署

### 自定义缓存

- 完全可定制
- 可集成任何第三方缓存

## 🔍 查询缓存

### 基础缓存

```go
// 缓存简单查询
users, err := torm.Table("users", "default").
    Where("age", ">", 25).
    Cache(10 * time.Minute).
    Get()

// 缓存复杂查询
orders, err := torm.Table("orders", "default").
    Where("status", "=", "pending").
    Where("created_at", ">", "2024-01-01").
    OrderBy("created_at", "DESC").
    Limit(100).
    Cache(5 * time.Minute).
    Get()

// 缓存聚合查询
stats, err := torm.Table("users", "default").
    Select("COUNT(*) as total, AVG(age) as avg_age").
    Where("status", "=", "active").
    Cache(15 * time.Minute).
    Get()
```

### 自定义缓存键

```go
// 使用自定义缓存键
users, err := torm.Table("users", "default").
    Where("status", "=", "active").
    CacheKey("active_users_list").
    Cache(5 * time.Minute).
    Get()

// 动态缓存键
userID := 123
user, err := torm.Table("users", "default").
    Where("id", "=", userID).
    CacheKey(fmt.Sprintf("user_detail_%d", userID)).
    Cache(15 * time.Minute).
    First()
```

### 缓存过期

```go
// 短期缓存
realtimeData, err := torm.Table("stats", "default").
    Cache(30 * time.Second).
    Get()

// 中期缓存
userList, err := torm.Table("users", "default").
    Cache(10 * time.Minute).
    Get()

// 长期缓存
staticData, err := torm.Table("categories", "default").
    Cache(1 * time.Hour).
    Get()
```

## 🏷️ 标签缓存

标签缓存允许对相关缓存项进行分组管理。

### 带标签的查询

```go
// 为用户查询添加标签
users, err := torm.Table("users", "default").
    Where("city", "=", "北京").
    CacheWithTags(10*time.Minute, "users", "city_beijing").
    Get()

// 为不同城市添加不同标签
shanghaiUsers, err := torm.Table("users", "default").
    Where("city", "=", "上海").
    CacheWithTags(10*time.Minute, "users", "city_shanghai").
    Get()

// 活跃用户查询
activeUsers, err := torm.Table("users", "default").
    Where("status", "=", "active").
    CacheWithTags(15*time.Minute, "users", "active_users").
    Get()
```

### 按标签清理缓存

```go
// 清理特定标签的缓存
err := torm.ClearCacheByTags("active_users")

// 清理多个标签
err = torm.ClearCacheByTags("users", "expired_data")

// 用户更新后清理相关缓存
func updateUser(userID int, data map[string]interface{}) error {
    _, err := torm.Table("users", "default").
        Where("id", "=", userID).
        Update(data)
    
    if err != nil {
        return err
    }
    
    // 清理用户相关缓存
    torm.ClearCacheByTags("users", "active_users")
    
    return nil
}
```

## 🎛️ 缓存管理

### 缓存统计

```go
// 获取缓存统计信息
stats := torm.GetCacheStats()
if stats != nil {
    fmt.Printf("总缓存项: %v\n", stats["total_items"])
    fmt.Printf("命中率: %.2f%%\n", stats["hit_rate"].(float64)*100)
    fmt.Printf("总命中数: %v\n", stats["total_hits"])
    fmt.Printf("总未命中数: %v\n", stats["total_misses"])
}
```

### 清理缓存

```go
// 清理所有缓存
err := torm.ClearAllCache()

// 按标签清理
err = torm.ClearCacheByTags("users")

// 应用启动时清理
func init() {
    torm.ClearAllCache()
}
```

## 💾 内存缓存

TORM 提供高性能的分片式内存缓存。

### 配置

```go
import "github.com/zhoudm1743/torm/db"

// 使用默认配置
cache := db.NewMemoryCache()

// 自定义配置
config := &db.CacheConfig{
    ShardCount:      16,                    // 分片数量
    MaxSize:         100000,                // 最大缓存项
    DefaultTTL:      time.Hour,             // 默认 TTL
    CleanupInterval: time.Minute,           // 清理间隔
    EvictionPolicy:  db.EvictionPolicyLRU,  // 淘汰策略
}
customCache := db.NewMemoryCacheWithConfig(config)
```

### 直接操作

```go
cache := db.NewMemoryCache()

// 基础操作
cache.Set("user:123", userData, time.Minute*30)
userData, err := cache.Get("user:123")
exists, _ := cache.Has("user:123")
cache.Delete("user:123")

// 批量操作
data := map[string]interface{}{
    "user:1": user1,
    "user:2": user2,
}
cache.SetMulti(data, time.Minute*10)
results, _ := cache.GetMulti([]string{"user:1", "user:2"})

// 数值操作
cache.Set("counter", int64(0), time.Hour)
newValue, _ := cache.Increment("counter", 1)
newValue, _ = cache.Decrement("counter", 1)

// 获取或设置（防止缓存击穿）
userData, err := cache.GetOrSet("user:123", func() (interface{}, error) {
    return loadUserFromDB(123)
}, time.Minute*30)
```

### 统计信息

```go
stats := cache.Stats()
fmt.Printf("总缓存项: %v\n", stats["total_items"])
fmt.Printf("命中率: %.2f%%\n", stats["hit_rate"].(float64)*100)
fmt.Printf("淘汰项数: %v\n", stats["total_evicted"])

// 重置统计
cache.ResetStats()
```

## 🔴 Redis 缓存

### 配置 Redis

```go
import "github.com/zhoudm1743/torm/db"

// Redis 配置
redisConfig := &db.RedisConfig{
    Address:     "localhost:6379",
    Password:    "",
    DB:          0,
    KeyPrefix:   "myapp:cache:",
    DefaultTTL:  time.Hour,
    TagsEnabled: true,
}

// 添加 Redis 缓存
err := db.AddCache("redis", db.CacheTypeRedis, redisConfig)
if err != nil {
    log.Fatal(err)
}
```

### 使用 Redis 缓存

```go
// 获取 Redis 缓存实例
redisCache, err := db.GetCache("redis")
if err != nil {
    log.Fatal(err)
}

// 基础操作
redisCache.Set("user:123", userData, time.Minute*30)
userData, err := redisCache.Get("user:123")

// 带标签的缓存
redisCache.SetWithTags("user:123", userData, time.Minute*30, []string{"users", "active"})
redisCache.DeleteByTags([]string{"users"})
```

### Redis 适配器

```go
import "github.com/go-redis/redis/v8"

// 实现 Redis 接口适配器
type GoRedisAdapter struct {
    client *redis.Client
}

func (r *GoRedisAdapter) Get(ctx context.Context, key string) (string, error) {
    return r.client.Get(ctx, key).Result()
}

func (r *GoRedisAdapter) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
    return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *GoRedisAdapter) Del(ctx context.Context, keys ...string) error {
    return r.client.Del(ctx, keys...).Err()
}

// 创建 Redis 客户端
func createRedisClient(config *db.RedisConfig) (db.RedisInterface, error) {
    rdb := redis.NewClient(&redis.Options{
        Addr:     config.Address,
        Password: config.Password,
        DB:       config.DB,
    })
    return &GoRedisAdapter{client: rdb}, nil
}
```

## 🛠️ 自定义缓存

### 实现缓存接口

```go
import "github.com/zhoudm1743/torm/db"

// 自定义缓存实现
type MyCustomCache struct {
    client MyThirdPartyCacheClient
    prefix string
}

func (c *MyCustomCache) Get(key string) (interface{}, error) {
    return c.client.Get(c.prefix + key)
}

func (c *MyCustomCache) Set(key string, value interface{}, ttl time.Duration) error {
    return c.client.Set(c.prefix + key, value, ttl)
}

func (c *MyCustomCache) Delete(key string) error {
    return c.client.Delete(c.prefix + key)
}

func (c *MyCustomCache) Clear() error {
    return c.client.Clear()
}

func (c *MyCustomCache) Has(key string) (bool, error) {
    return c.client.Exists(c.prefix + key)
}

func (c *MyCustomCache) Size() int {
    return c.client.Size()
}

func (c *MyCustomCache) Close() error {
    return c.client.Close()
}
```

### 注册自定义缓存

```go
// 自定义缓存提供者
type MyCustomCacheProvider struct{}

func (p *MyCustomCacheProvider) CreateCache(config interface{}) (db.FullCacheInterface, error) {
    customConfig := config.(*MyCustomCacheConfig)
    client, err := NewMyThirdPartyCacheClient(customConfig)
    if err != nil {
        return nil, err
    }
    return &MyCustomCache{client: client, prefix: "myapp:"}, nil
}

// 注册并使用
manager := db.GetDefaultCacheManager()
manager.RegisterCacheProvider(db.CacheTypeCustom, &MyCustomCacheProvider{})
manager.AddCache("custom", db.CacheTypeCustom, customConfig)

customCache, _ := manager.GetCache("custom")
customCache.Set("key", "value", time.Minute)
```

## 📊 缓存管理器

### 多缓存实例

```go
manager := db.GetDefaultCacheManager()

// 添加不同类型的缓存
manager.AddCache("memory", db.CacheTypeMemory, memConfig)
manager.AddCache("redis", db.CacheTypeRedis, redisConfig)

// 列出所有缓存
cacheNames := manager.ListCaches()

// 获取缓存实例
memCache, _ := manager.GetCache("memory")
redisCache, _ := manager.GetCache("redis")

// 获取所有统计
allStats := manager.GetStats()

// 移除缓存
manager.RemoveCache("custom")

// 关闭所有缓存
manager.CloseAll()
```

## 💡 最佳实践

### 1. 合理设置 TTL

```go
// 根据数据更新频率设置 TTL

// 用户基本信息：相对稳定
userInfo, _ := torm.Table("users", "default").
    Where("id", "=", userID).
    Cache(30 * time.Minute).
    First()

// 实时数据：短期缓存
onlineUsers, _ := torm.Table("users", "default").
    Where("last_active_at", ">", time.Now().Add(-5*time.Minute)).
    Cache(1 * time.Minute).
    Get()

// 统计数据：长期缓存
dailyStats, _ := torm.Table("orders", "default").
    Select("DATE(created_at) as date, COUNT(*) as count").
    GroupBy("DATE(created_at)").
    Cache(1 * time.Hour).
    Get()
```

### 2. 使用标签组织缓存

```go
// 按功能模块组织
userCache := torm.Table("users", "default").
    CacheWithTags(10*time.Minute, "users", "user_list")

orderCache := torm.Table("orders", "default").
    CacheWithTags(5*time.Minute, "orders", "order_list")

// 数据更新时清理相关标签
func updateUser(userID int, data map[string]interface{}) error {
    _, err := torm.Table("users", "default").
        Where("id", "=", userID).
        Update(data)
    
    if err == nil {
        torm.ClearCacheByTags("users")
    }
    return err
}
```

### 3. 事务中禁用缓存

```go
// 事务中的查询自动跳过缓存
err := torm.Transaction(func(tx torm.TransactionInterface) error {
    builder, _ := torm.Table("users", "default")
    builder.InTransaction(tx)
    
    // 事务中不使用缓存，确保数据一致性
    users, err := builder.Where("status", "=", "active").Get()
    
    return err
}, "default")
```

### 4. 缓存键命名规范

```go
// 推荐的命名方式
// 模块:实体:标识

// 用户详情
CacheKey(fmt.Sprintf("user:detail:%d", userID))

// 用户列表
CacheKey("user:list:active")

// 统计数据
CacheKey("stats:orders:daily")
```

### 5. 防止缓存击穿

```go
// 使用 GetOrSet 防止缓存击穿
cache := db.NewMemoryCache()

userData, err := cache.GetOrSet("user:123", func() (interface{}, error) {
    // 只在缓存不存在时执行
    return loadUserFromDB(123)
}, time.Minute*30)
```

### 6. 监控缓存效果

```go
// 定期检查缓存统计
func monitorCache() {
    stats := torm.GetCacheStats()
    if stats != nil {
        hitRate := stats["hit_rate"].(float64)
        if hitRate < 0.5 {
            log.Printf("警告: 缓存命中率过低 %.2f%%", hitRate*100)
        }
    }
}
```

## 🔗 相关文档

- [查询构建器](Query-Builder) - 查询缓存集成
- [配置指南](Configuration) - 缓存配置
- [性能优化](Best-Practices#性能优化) - 缓存优化技巧
