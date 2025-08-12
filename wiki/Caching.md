# 缓存系统

TORM 提供了完整的缓存系统，包括内存缓存、查询缓存和关联缓存，帮助提升应用性能。

## 📋 目录

- [基础缓存](#基础缓存)
- [查询缓存](#查询缓存)
- [关联缓存](#关联缓存)
- [缓存管理](#缓存管理)
- [缓存策略](#缓存策略)
- [性能优化](#性能优化)

## 🚀 快速开始

### 基础缓存使用

```go
import "github.com/zhoudm1743/torm/cache"

// 创建内存缓存实例
memCache := cache.NewMemoryCache()

// 设置缓存
err := memCache.Set("user:1", userData, 5*time.Minute)

// 获取缓存
data, err := memCache.Get("user:1")

// 删除缓存
err = memCache.Delete("user:1")
```

## 💾 基础缓存

### 内存缓存

```go
// 创建缓存实例
cache := cache.NewMemoryCache()

// 设置缓存（带TTL）
err := cache.Set("key", "value", 10*time.Minute)

// 获取缓存
value, err := cache.Get("key")
if err != nil {
    if errors.Is(err, cache.ErrCacheNotFound) {
        // 缓存不存在
    }
}

// 检查缓存是否存在
exists, err := cache.Has("key")

// 获取缓存并自动删除
value, err := cache.Pull("key")

// 删除缓存
err = cache.Delete("key")

// 清空所有缓存
err = cache.Flush()
```

### 缓存标签

```go
// 设置带标签的缓存
err := cache.SetWithTags("user:1", userData, 5*time.Minute, "users", "user_1")

// 根据标签清除缓存
err = cache.FlushByTags("users")
```

## 🔍 查询缓存

### 基础查询缓存

```go
// 缓存查询结果
users, err := db.Table("users").
    Where("status", "=", "active").
    Cache(5 * time.Minute).
    Get()

// 带标签的查询缓存
users, err := db.Table("users").
    Where("status", "=", "active").
    CacheWithTags(5*time.Minute, "users", "active_users").
    Get()
```

### 模型缓存

```go
user := models.NewUser()

// 缓存模型查询
users, err := user.Where("status", "=", "active").
    Cache(10 * time.Minute).
    Get()

// 缓存单个模型
user, err := user.Where("id", "=", 1).
    Cache(5 * time.Minute).
    First()
```

## 🔗 关联缓存

### 关联查询缓存

```go
user := models.NewUser()

// 缓存关联数据
posts, err := user.Posts().
    Cache(15 * time.Minute).
    Get()

// 预加载缓存
users, err := user.With("Posts").
    Cache(10 * time.Minute).
    Get()
```

## 🔧 缓存管理

### 缓存配置

```go
// 配置缓存
config := &cache.Config{
    DefaultTTL:    5 * time.Minute,
    MaxSize:       1000,
    CleanupInterval: 1 * time.Minute,
}

cache := cache.NewMemoryCache(config)
```

### 缓存统计

```go
// 获取缓存统计信息
stats := cache.Stats()
log.Printf("缓存命中率: %.2f%%", stats.HitRate())
log.Printf("缓存大小: %d", stats.Size())
log.Printf("命中次数: %d", stats.Hits())
log.Printf("未命中次数: %d", stats.Misses())
```

## 📚 最佳实践

### 1. 缓存键设计

```go
// 好的做法：使用有意义的键名
userCacheKey := fmt.Sprintf("user:%d", userID)
postsCacheKey := fmt.Sprintf("user:%d:posts", userID)

// 使用标签分组
cache.SetWithTags(userCacheKey, userData, 5*time.Minute, "users", fmt.Sprintf("user_%d", userID))
```

### 2. 缓存失效

```go
// 数据更新时清除相关缓存
func (u *User) AfterUpdate() error {
    cacheKey := fmt.Sprintf("user_%d", u.ID)
    cache.FlushByTags(cacheKey)
    return nil
}
```

## 🔗 相关文档

- [查询构建器](Query-Builder) - 查询缓存
- [模型系统](Model-System) - 模型缓存
- [关联关系](Relationships) - 关联缓存
- [性能优化](Performance) - 缓存性能优化 