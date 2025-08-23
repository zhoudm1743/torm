# 缓存系统

TORM 提供了完整的可扩展缓存系统，支持内存缓存、Redis缓存和自定义缓存实现。缓存系统与查询构建器无缝集成，支持自动缓存键生成、TTL管理、标签清理和高并发访问。通过模块化的接口设计，您可以轻松地集成任何第三方缓存解决方案。

## 📋 目录

- [快速开始](#快速开始)
- [缓存类型](#缓存类型)
- [内存缓存](#内存缓存)
- [Redis缓存](#redis缓存)
- [自定义缓存](#自定义缓存)
- [缓存管理器](#缓存管理器)
- [查询缓存](#查询缓存)
- [标签缓存](#标签缓存)
- [性能测试](#性能测试)
- [最佳实践](#最佳实践)

## 🚀 快速开始

### 使用默认内存缓存

```go
import (
    "github.com/zhoudm1743/torm"
    "github.com/zhoudm1743/torm/db"
    "time"
)

// 使用默认的高并发内存缓存
users, err := torm.Table("users", "default").
    Where("status", "=", "active").
    Cache(5 * time.Minute).
    Get()

if err != nil {
    log.Fatal(err)
}

// 第二次相同查询会从缓存获取，显著提升性能
users2, err := torm.Table("users", "default").
    Where("status", "=", "active").
    Cache(5 * time.Minute).
    Get()
```

### 使用Redis缓存

```go
import (
    "github.com/go-redis/redis/v8"
    "github.com/zhoudm1743/torm/db"
)

// 配置Redis缓存
redisConfig := &db.RedisConfig{
    Address:     "localhost:6379",
    Password:    "",
    DB:          0,
    KeyPrefix:   "myapp:cache:",
    DefaultTTL:  time.Hour,
    TagsEnabled: true,
}

// 添加Redis缓存实例
err := db.AddCache("redis", db.CacheTypeRedis, redisConfig)
if err != nil {
    log.Fatal("Failed to add Redis cache:", err)
}

// 使用Redis缓存
redisCache, err := db.GetCache("redis")
if err != nil {
    log.Fatal("Failed to get Redis cache:", err)
}

// 手动使用缓存
redisCache.Set("user:123", userData, time.Minute*30)
userData, err := redisCache.Get("user:123")
```

## 🔄 缓存类型

TORM 支持多种缓存实现，您可以根据需求选择合适的缓存类型：

### 1. 内存缓存 (Memory Cache)
- **特点**: 高性能、低延迟、数据存储在应用内存中
- **适用场景**: 单机部署、临时数据缓存、性能要求极高的场景
- **优点**: 访问速度极快、无网络开销
- **缺点**: 数据不持久化、受内存限制、不支持分布式

### 2. Redis缓存 (Redis Cache)
- **特点**: 分布式、持久化、支持复杂数据结构
- **适用场景**: 分布式部署、需要持久化、大容量缓存
- **优点**: 支持集群、数据持久化、功能丰富
- **缺点**: 需要额外的Redis服务、有网络延迟

### 3. 自定义缓存 (Custom Cache)
- **特点**: 完全可定制、可以集成任何第三方缓存
- **适用场景**: 特殊需求、现有缓存系统集成
- **优点**: 灵活性最高、可以满足任何需求
- **缺点**: 需要自己实现接口

## 💾 内存缓存

TORM 提供了高性能的分片式内存缓存，支持高并发访问和多种淘汰策略。

### 基础配置

```go
import "github.com/zhoudm1743/torm/db"

// 使用默认配置创建内存缓存
cache := db.NewMemoryCache()

// 使用自定义配置
config := &db.CacheConfig{
    ShardCount:      16,                    // 分片数量（默认CPU核心数*2）
    MaxSize:         100000,                // 最大缓存项数量
    DefaultTTL:      time.Hour,             // 默认TTL
    CleanupInterval: time.Minute,           // 清理间隔
    EvictionPolicy:  db.EvictionPolicyLRU,  // 淘汰策略：LRU/LFU/TTL
}
customCache := db.NewMemoryCacheWithConfig(config)
```

### 高级操作

```go
// 基础操作
cache.Set("user:123", userData, time.Minute*30)
userData, err := cache.Get("user:123")
exists, _ := cache.Has("user:123")
cache.Delete("user:123")

// 批量操作
data := map[string]interface{}{
    "user:1": user1Data,
    "user:2": user2Data,
    "user:3": user3Data,
}
cache.SetMulti(data, time.Minute*10)
results, _ := cache.GetMulti([]string{"user:1", "user:2", "user:3"})

// 数值操作
cache.Set("counter", int64(0), time.Hour)
newValue, _ := cache.Increment("counter", 1)     // 递增
newValue, _ := cache.Decrement("counter", 1)     // 递减

// TTL操作
cache.Touch("user:123", time.Minute*60)         // 更新过期时间
ttl, _ := cache.TTL("user:123")                 // 获取剩余TTL

// 获取或设置（防止缓存击穿）
userData, err := cache.GetOrSet("user:123", func() (interface{}, error) {
    // 这个函数只在缓存不存在时调用
    return loadUserFromDB(123)
}, time.Minute*30)
```

### 性能统计

```go
// 获取详细统计信息
stats := cache.Stats()
fmt.Printf("总缓存项: %v\n", stats["total_items"])
fmt.Printf("命中率: %.2f%%\n", stats["hit_rate"].(float64)*100)
fmt.Printf("总命中数: %v\n", stats["total_hits"])
fmt.Printf("总未命中数: %v\n", stats["total_misses"])
fmt.Printf("淘汰项数: %v\n", stats["total_evicted"])
fmt.Printf("分片数量: %v\n", stats["shard_count"])

// 分片统计（用于调试和优化）
shardStats := stats["shard_stats"].([]map[string]interface{})
for i, shardStat := range shardStats {
    fmt.Printf("分片 %d: 项目数=%v, 命中数=%v\n", 
        i, shardStat["items"], shardStat["hits"])
}

// 重置统计信息
cache.ResetStats()
```

## 🔴 Redis缓存

Redis缓存提供分布式缓存能力，支持数据持久化和集群部署。

### Redis适配器实现

首先需要实现Redis客户端工厂函数：

```go
import (
    "github.com/go-redis/redis/v8"
    "github.com/zhoudm1743/torm/db"
    "context"
)

// 实现Redis接口适配器
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

// ... 实现其他必需的方法

// Redis客户端工厂函数
func createRedisClient(config *db.RedisConfig) (db.RedisInterface, error) {
    rdb := redis.NewClient(&redis.Options{
        Addr:         config.Address,
        Password:     config.Password,
        DB:           config.DB,
        PoolSize:     config.PoolSize,
        DialTimeout:  config.DialTimeout,
        MaxRetries:   config.MaxRetries,
    })
    
    return &GoRedisAdapter{client: rdb}, nil
}
```

### 注册和使用Redis缓存

```go
import "github.com/zhoudm1743/torm/db"

func main() {
    // 注册Redis缓存提供者
    manager := db.GetDefaultCacheManager()
    redisProvider := db.NewRedisCacheProvider(createRedisClient)
    manager.RegisterCacheProvider(db.CacheTypeRedis, redisProvider)
    
    // 配置Redis缓存
    redisConfig := &db.RedisConfig{
        Address:     "localhost:6379",
        Password:    "",
        DB:          0,
        KeyPrefix:   "myapp:cache:",
        DefaultTTL:  time.Hour,
        TagsEnabled: true,
        Serializer:  "json",
    }
    
    // 添加Redis缓存实例
    err := manager.AddCache("redis", db.CacheTypeRedis, redisConfig)
    if err != nil {
        log.Fatal("Failed to add Redis cache:", err)
    }
    
    // 获取并使用Redis缓存
    redisCache, err := manager.GetCache("redis")
    if err != nil {
        log.Fatal("Failed to get Redis cache:", err)
    }
    
    // 使用Redis缓存
    redisCache.Set("user:123", userData, time.Minute*30)
    userData, err := redisCache.Get("user:123")
    
    // 带标签的缓存
    redisCache.SetWithTags("user:123", userData, time.Minute*30, []string{"users", "active"})
    redisCache.DeleteByTags([]string{"users"}) // 删除所有用户相关缓存
}
```

## 🛠️ 自定义缓存

您可以实现自定义缓存适配器来集成任何第三方缓存系统。

### 实现缓存接口

```go
import (
    "github.com/zhoudm1743/torm/db"
    "time"
)

// 自定义缓存实现
type MyCustomCache struct {
    // 您的缓存客户端
    client MyThirdPartyCacheClient
    prefix string
}

// 实现基础缓存接口
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

// 实现高级接口（可选）
func (c *MyCustomCache) GetMulti(keys []string) (map[string]interface{}, error) {
    // 实现批量获取
    // ...
}

func (c *MyCustomCache) SetWithTags(key string, value interface{}, ttl time.Duration, tags []string) error {
    // 实现标签缓存
    // ...
}

// ... 实现其他接口方法
```

### 实现缓存提供者

```go
// 自定义缓存配置
type MyCustomCacheConfig struct {
    Endpoint string `json:"endpoint"`
    APIKey   string `json:"api_key"`
    Timeout  time.Duration `json:"timeout"`
}

// 自定义缓存提供者
type MyCustomCacheProvider struct{}

func (p *MyCustomCacheProvider) CreateCache(config interface{}) (db.FullCacheInterface, error) {
    customConfig, ok := config.(*MyCustomCacheConfig)
    if !ok {
        return nil, fmt.Errorf("invalid config type")
    }
    
    client, err := NewMyThirdPartyCacheClient(customConfig)
    if err != nil {
        return nil, err
    }
    
    return &MyCustomCache{
        client: client,
        prefix: "myapp:",
    }, nil
}

func (p *MyCustomCacheProvider) ValidateConfig(config interface{}) error {
    customConfig, ok := config.(*MyCustomCacheConfig)
    if !ok {
        return fmt.Errorf("invalid config type")
    }
    
    if customConfig.Endpoint == "" {
        return fmt.Errorf("endpoint is required")
    }
    
    return nil
}

func (p *MyCustomCacheProvider) GetConfigExample() interface{} {
    return &MyCustomCacheConfig{
        Endpoint: "https://api.mycache.com",
        APIKey:   "your-api-key",
        Timeout:  5 * time.Second,
    }
}
```

### 注册和使用自定义缓存

```go
func main() {
    // 注册自定义缓存提供者
    manager := db.GetDefaultCacheManager()
    customProvider := &MyCustomCacheProvider{}
    manager.RegisterCacheProvider(db.CacheTypeCustom, customProvider)
    
    // 配置自定义缓存
    customConfig := &MyCustomCacheConfig{
        Endpoint: "https://my-cache-service.com",
        APIKey:   "your-api-key",
        Timeout:  5 * time.Second,
    }
    
    // 添加自定义缓存实例
    err := manager.AddCache("custom", db.CacheTypeCustom, customConfig)
    if err != nil {
        log.Fatal("Failed to add custom cache:", err)
    }
    
    // 使用自定义缓存
    customCache, err := manager.GetCache("custom")
    if err != nil {
        log.Fatal("Failed to get custom cache:", err)
    }
    
    customCache.Set("user:123", userData, time.Minute*30)
    userData, err := customCache.Get("user:123")
}
```

## 🎛️ 缓存管理器

缓存管理器允许您在应用中同时使用多种缓存实现。

### 基础使用

```go
import "github.com/zhoudm1743/torm/db"

// 获取默认缓存管理器
manager := db.GetDefaultCacheManager()

// 或创建新的管理器
manager := db.NewCacheManager()

// 添加不同类型的缓存
// 内存缓存
memConfig := db.DefaultCacheConfig()
manager.AddCache("memory", db.CacheTypeMemory, memConfig)

// Redis缓存
redisConfig := &db.RedisConfig{...}
manager.AddCache("redis", db.CacheTypeRedis, redisConfig)

// 自定义缓存
customConfig := &MyCustomCacheConfig{...}
manager.AddCache("custom", db.CacheTypeCustom, customConfig)
```

### 缓存管理操作

```go
// 列出所有缓存实例
cacheNames := manager.ListCaches()
fmt.Println("可用缓存:", cacheNames)

// 获取可用的缓存类型
availableTypes := manager.GetAvailableTypes()
fmt.Println("可用类型:", availableTypes)

// 获取或创建缓存实例
cache, err := manager.GetOrCreateCache("session", db.CacheTypeMemory, memConfig)
if err != nil {
    log.Fatal("Failed to get or create cache:", err)
}

// 获取所有缓存的统计信息
allStats := manager.GetStats()
for name, stats := range allStats {
    fmt.Printf("缓存 %s 统计: %+v\n", name, stats)
}

// 移除缓存实例
manager.RemoveCache("custom")

// 关闭所有缓存
manager.CloseAll()
```

### 工厂模式和自定义工厂

```go
// 创建自定义工厂
factory := db.NewCacheFactory()

// 注册缓存提供者
factory.RegisterCacheProvider("mycache", &MyCustomCacheProvider{})

// 使用自定义工厂创建管理器
manager := db.NewCacheManagerWithFactory(factory)

// 创建缓存实例
cache, err := factory.CreateCache("mycache", customConfig)
if err != nil {
    log.Fatal("Failed to create cache:", err)
}
```

## 🔍 查询缓存

### 基础查询缓存

TORM 的查询缓存会自动根据查询条件生成唯一的缓存键，确保缓存的准确性：

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

当需要更精确控制缓存时，可以设置自定义缓存键：

```go
// 使用自定义缓存键
customKey := "active_users_summary"
users, err := torm.Table("users", "default").
    Where("status", "=", "active").
    Cache(5 * time.Minute).
    CacheKey(customKey).
    Get()

// 方法链可以任意顺序
users2, err := torm.Table("users", "default").
    CacheKey("vip_users").
    Where("vip_level", ">", 5).
    Cache(10 * time.Minute).
    Get()
```

## 🏷️ 标签缓存

标签缓存允许您对相关的缓存项进行分组管理，便于批量清理：

### 带标签的查询缓存

```go
// 为用户相关查询添加标签
users, err := torm.Table("users", "default").
    Where("city", "=", "北京").
    CacheWithTags(10*time.Minute, "users", "city_beijing").
    Get()

// 为不同城市的用户添加不同标签
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

### 标签管理

```go
// 清理特定标签的所有缓存
err := torm.ClearCacheByTags("active_users")
if err != nil {
    log.Printf("清理缓存失败: %v", err)
}

// 清理多个标签
err = torm.ClearCacheByTags("users", "expired_data")

// 用户更新后清理相关缓存
func updateUserStatus(userID int, status string) error {
    // 更新数据库
    _, err := torm.Table("users", "default").
        Where("id", "=", userID).
        Update(map[string]interface{}{
            "status": status,
        })
    
    if err != nil {
        return err
    }
    
    // 清理相关缓存
    torm.ClearCacheByTags("users", "active_users")
    
    return nil
}
```

## 🔧 缓存管理

### 缓存统计

监控缓存使用情况，优化缓存策略：

```go
// 获取缓存统计信息
stats := torm.GetCacheStats()
if stats != nil {
    fmt.Printf("总缓存项数: %v\n", stats["total_items"])
    fmt.Printf("过期项数: %v\n", stats["expired_items"])
    fmt.Printf("标签数量: %v\n", stats["total_tags"])
}
```

### 缓存清理

```go
// 清理所有缓存
err := torm.ClearAllCache()
if err != nil {
    log.Printf("清理所有缓存失败: %v", err)
}

// 在应用启动时清理缓存
func init() {
    torm.ClearAllCache()
    log.Println("应用启动时已清理所有缓存")
}
```

### 缓存过期处理

TORM 的内存缓存会自动处理过期项：

```go
// 设置短期缓存（2秒后过期）
users, err := torm.Table("users", "default").
    Where("status", "=", "active").
    Cache(2 * time.Second).
    Get()

// 立即查询 - 命中缓存
users2, err := torm.Table("users", "default").
    Where("status", "=", "active").
    Cache(2 * time.Second).
    Get()

// 等待 3 秒后查询 - 重新从数据库获取
time.Sleep(3 * time.Second)
users3, err := torm.Table("users", "default").
    Where("status", "=", "active").
    Cache(2 * time.Second).
    Get()
```

## 📊 性能测试

### 缓存效果验证

```go
import "time"

func benchmarkCachePerformance() {
    // 第一次查询 - 从数据库获取
    start := time.Now()
    users1, err := torm.Table("users", "default").
        Where("status", "=", "active").
        Cache(5 * time.Minute).
        Get()
    firstQueryTime := time.Since(start)
    
    // 第二次查询 - 从缓存获取
    start = time.Now()
    users2, err := torm.Table("users", "default").
        Where("status", "=", "active").
        Cache(5 * time.Minute).
        Get()
    secondQueryTime := time.Since(start)
    
    fmt.Printf("第一次查询(数据库): %v\n", firstQueryTime)
    fmt.Printf("第二次查询(缓存): %v\n", secondQueryTime)
    fmt.Printf("性能提升: %.1fx\n", float64(firstQueryTime)/float64(secondQueryTime))
}
```

## 💡 最佳实践

### 1. 合理设置 TTL

```go
// 根据数据更新频率设置 TTL
// 用户基本信息：相对稳定，可以缓存较长时间
userInfo, err := torm.Table("users", "default").
    Where("id", "=", userID).
    Cache(30 * time.Minute).
    First()

// 实时数据：需要较短的缓存时间
onlineUsers, err := torm.Table("users", "default").
    Where("last_active_at", ">", time.Now().Add(-5*time.Minute)).
    Cache(1 * time.Minute).
    Get()

// 统计数据：可以缓存更长时间
dailyStats, err := torm.Table("orders", "default").
    Select("DATE(created_at) as date, COUNT(*) as count").
    GroupBy("DATE(created_at)").
    Cache(1 * time.Hour).
    Get()
```

### 2. 使用标签组织缓存

```go
// 按功能模块组织标签
userCache := torm.Table("users", "default").
    CacheWithTags(10*time.Minute, "users", "user_list")

orderCache := torm.Table("orders", "default").
    CacheWithTags(5*time.Minute, "orders", "order_list")

// 按数据更新频率组织标签
staticData := torm.Table("categories", "default").
    CacheWithTags(1*time.Hour, "static", "categories")

dynamicData := torm.Table("products", "default").
    CacheWithTags(5*time.Minute, "dynamic", "products")
```

### 3. 事务中禁用缓存

```go
// 在事务中查询不使用缓存，确保数据一致性
err := torm.Transaction(func(tx torm.TransactionInterface) error {
    builder, _ := torm.Table("users", "default")
    builder.InTransaction(tx)
    
    // 事务中的查询自动跳过缓存
    users, err := builder.Where("status", "=", "active").
        Cache(5 * time.Minute). // 这个设置在事务中会被忽略
        Get()
    
    return err
}, "default")
```

### 4. 缓存键命名规范

```go
// 推荐的缓存键命名方式
// 使用 CacheKey 方法设置有意义的键名

// 用户列表
users, err := torm.Table("users", "default").
    Where("status", "=", "active").
    CacheKey("active_users_list").
    Cache(10 * time.Minute).
    Get()

// 用户详情
user, err := torm.Table("users", "default").
    Where("id", "=", userID).
    CacheKey(fmt.Sprintf("user_detail_%d", userID)).
    Cache(15 * time.Minute).
    First()

// 统计数据
stats, err := torm.Table("orders", "default").
    Select("COUNT(*) as total").
    Where("status", "=", "completed").
    CacheKey("completed_orders_count").
    Cache(5 * time.Minute).
    Get()
```

### 5. 缓存更新策略

```go
// 数据更新时主动清理缓存
func updateUser(userID int, data map[string]interface{}) error {
    // 更新数据
    _, err := torm.Table("users", "default").
        Where("id", "=", userID).
        Update(data)
    
    if err != nil {
        return err
    }
    
    // 清理相关缓存
    torm.ClearCacheByTags("users")
    
    // 或者清理特定用户的缓存
    userCacheKey := fmt.Sprintf("user_detail_%d", userID)
    // 注意：目前需要通过标签来清理，未来版本可能支持直接按键清理
    
    return nil
}
```

## 🚀 性能优势

TORM 缓存系统提供了多层次的性能优化：

### 内存缓存性能
- **高并发支持**: 分片式设计，支持数万并发操作
- **无锁读取**: 使用原子操作，读取性能提升 10-100 倍
- **智能淘汰**: LRU/LFU 算法，保持缓存命中率
- **内存优化**: 自动清理过期数据，防止内存泄漏

### Redis缓存性能
- **分布式缓存**: 支持集群部署，容量无限扩展
- **持久化支持**: 数据持久化，重启不丢失
- **批量操作**: 管道操作，提升批量处理性能
- **标签管理**: 高效的标签索引，快速批量清理

### 性能数据对比

```go
// 性能基准测试结果（12核CPU，32GB内存）
// 内存缓存（分片模式 vs 单锁模式）
BenchmarkHighConcurrencyCache_Set-20    10210071    1911 ns/op   145 B/op   5 allocs/op
BenchmarkOldCache_Set-20                  3094946   24261 ns/op   146 B/op   5 allocs/op
// 性能提升：12.7x

BenchmarkHighConcurrencyCache_Get-20    40206238    43.36 ns/op   15 B/op   1 allocs/op  
BenchmarkOldCache_Get-20                 17356450    70.00 ns/op   15 B/op   1 allocs/op
// 性能提升：1.6x

BenchmarkHighConcurrencyCache_Mixed-20  18355443    64.28 ns/op   50 B/op   2 allocs/op
// 混合读写操作（70%读，30%写）
```

### 缓存效果示例

```go
func demonstrateCachePerformance() {
    cache := db.NewMemoryCache()
    
    // 模拟复杂查询
    complexQuery := func() (interface{}, error) {
        time.Sleep(100 * time.Millisecond) // 模拟数据库查询
        return "complex result", nil
    }
    
    // 第一次查询 - 100ms+
    start := time.Now()
    result1, _ := cache.GetOrSet("complex", complexQuery, time.Minute)
    firstTime := time.Since(start)
    
    // 第二次查询 - 从缓存获取 < 1ms
    start = time.Now()
    result2, _ := cache.Get("complex")
    secondTime := time.Since(start)
    
    fmt.Printf("第一次查询: %v\n", firstTime)     // ~100ms
    fmt.Printf("第二次查询: %v\n", secondTime)    // ~0.04ms
    fmt.Printf("性能提升: %.0fx\n", float64(firstTime)/float64(secondTime)) // ~2500x
}

## 📚 相关文档

- [查询构建器](Query-Builder) - 查询缓存集成
- [模型系统](Model-System) - 模型级缓存
- [事务处理](Transactions) - 事务与缓存
- [性能优化](Performance) - 缓存性能优化技巧