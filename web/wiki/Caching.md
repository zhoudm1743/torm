# 缓存系统

TORM 提供了完整的缓存系统，包括内存缓存、查询缓存和标签缓存，帮助显著提升应用性能。缓存系统与查询构建器无缝集成，支持自动缓存键生成、TTL管理和标签清理。

## 📋 目录

- [快速开始](#快速开始)
- [查询缓存](#查询缓存)
- [标签缓存](#标签缓存)
- [缓存管理](#缓存管理)
- [性能测试](#性能测试)
- [最佳实践](#最佳实践)

## 🚀 快速开始

### 基础查询缓存

```go
import "github.com/zhoudm1743/torm"

// 缓存查询结果 5 分钟
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

使用 TORM 缓存系统可以获得显著的性能提升：

- **查询速度提升**: 缓存命中时查询速度提升 50-1000 倍
- **数据库负载减少**: 减少重复查询对数据库的压力
- **内存管理**: 自动过期清理，防止内存泄漏
- **灵活控制**: 支持 TTL、标签、自定义键等多种控制方式

## 📚 相关文档

- [查询构建器](Query-Builder) - 查询缓存集成
- [模型系统](Model-System) - 模型级缓存
- [事务处理](Transactions) - 事务与缓存
- [性能优化](Performance) - 缓存性能优化技巧