# 性能优化

本文档提供 TORM 的性能优化指南，包括查询优化、缓存策略、连接池配置等最佳实践。

## 🚀 v1.1.6 性能优化

### 依赖精简
- **移除GORM依赖**: 使用纯Go `modernc.org/sqlite` 驱动，减少依赖复杂度  
- **SQL构建优化**: 增强的SQL构建器，支持更高效的查询条件组合
- **接口统一**: 减少类型转换开销，提升查询性能

### 查询优化  
- **智能随机排序**: `OrderRand()` 自动检测数据库类型，使用最优随机函数
- **优化范围查询**: `WhereBetween/WhereNotBetween` 使用参数绑定，防止SQL注入
- **子查询优化**: `WhereExists/WhereNotExists` 支持查询构建器，减少字符串拼接开销

## 📋 目录

- [查询优化](#查询优化)
- [索引优化](#索引优化)
- [缓存策略](#缓存策略)
- [连接池优化](#连接池优化)
- [批量操作](#批量操作)
- [性能监控](#性能监控)

## 🚀 查询优化

### 避免 N+1 查询

```go
// ❌ 错误做法 - 产生 N+1 查询
users, _ := db.Table("users").Get()
for _, user := range users {
    posts, _ := db.Table("posts").Where("user_id", "=", user["id"]).Get()
}

// ✅ 正确做法 - 使用预加载
users, _ := user.With("Posts").Get()
```

### 选择性字段查询

```go
// ❌ 避免查询所有字段
users, _ := db.Table("users").Get()

// ✅ 只查询需要的字段
users, _ := db.Table("users").Select("id", "name", "email").Get()
```

### 使用索引

```go
// ✅ 利用索引字段进行查询
users, _ := db.Table("users").
    Where("email", "=", "user@example.com"). // email 有索引
    Where("status", "=", "active").           // 复合索引
    Get()
```

## 📈 索引优化

### 创建适当的索引

```sql
-- 单列索引
CREATE INDEX idx_users_email ON users(email);

-- 复合索引
CREATE INDEX idx_users_status_created ON users(status, created_at);

-- 唯一索引
CREATE UNIQUE INDEX idx_users_email_unique ON users(email);
```

### 索引使用建议

```go
// ✅ 好的做法：WHERE 条件使用索引字段
query.Where("email", "=", email).           // email 有索引
      Where("created_at", ">", startDate)   // 范围查询使用索引

// ❌ 避免：在索引字段上使用函数
// query.WhereRaw("UPPER(email) = ?", strings.ToUpper(email))

// ✅ 正确做法：
query.Where("email", "=", strings.ToLower(email))
```

## 💾 缓存策略

### 查询结果缓存

```go
// 缓存频繁查询的结果
users, err := db.Table("users").
    Where("status", "=", "active").
    Cache(5 * time.Minute).
    Get()

// 使用标签便于缓存管理
users, err := db.Table("users").
    Where("status", "=", "active").
    CacheWithTags(5*time.Minute, "users", "active").
    Get()
```

### 模型缓存

```go
// 缓存模型查询
user := models.NewUser()
userData, err := user.Where("id", "=", 1).
    Cache(10 * time.Minute).
    First()
```

## 🔧 连接池优化

### 连接池配置

```go
config := &db.Config{
    Driver:   "mysql",
    Host:     "localhost",
    Port:     3306,
    Database: "myapp",
    Username: "user",
    Password: "password",
    
    // 连接池配置
    MaxOpenConns:    100,              // 最大打开连接数
    MaxIdleConns:    10,               // 最大空闲连接数
    ConnMaxLifetime: time.Hour,        // 连接最大生存时间
    ConnMaxIdleTime: 30 * time.Minute, // 连接最大空闲时间
}
```

### 监控连接池

```go
// 获取连接池统计信息
stats := db.Stats()
log.Printf("打开连接数: %d", stats.OpenConnections)
log.Printf("使用中连接数: %d", stats.InUse)
log.Printf("空闲连接数: %d", stats.Idle)
```

## 📦 批量操作

### 批量插入

```go
// ✅ 使用批量插入
users := []map[string]interface{}{
    {"name": "用户1", "email": "user1@example.com"},
    {"name": "用户2", "email": "user2@example.com"},
    {"name": "用户3", "email": "user3@example.com"},
}
affected, err := db.Table("users").InsertBatch(users)

// ❌ 避免逐条插入
// for _, user := range users {
//     db.Table("users").Insert(user)
// }
```

### 分批处理大数据

```go
// 分批处理大量数据
err := db.Table("users").Chunk(1000, func(users []map[string]interface{}) bool {
    // 处理每批1000条数据
    for _, user := range users {
        processUser(user)
    }
    return true // 继续处理下一批
})
```

## 📊 性能监控

### SQL查询监控

```go
// 启用查询日志
db.EnableQueryLog()

// 设置慢查询阈值
db.SetSlowQueryThreshold(100 * time.Millisecond)

// 获取查询日志
logs := db.GetQueryLog()
for _, log := range logs {
    if log.Duration > 100*time.Millisecond {
        fmt.Printf("慢查询: %s, 耗时: %v\n", log.SQL, log.Duration)
    }
}
```

### 性能分析

```go
// 查询执行计划
explain, err := db.Table("users").
    Where("status", "=", "active").
    Explain()
```

## 📚 最佳实践

### 1. 分页优化

```go
// ✅ 对于大数据量，使用游标分页
users, err := db.Table("users").
    Where("id", ">", lastID).
    OrderBy("id", "asc").
    Limit(100).
    Get()

// ❌ 避免大偏移量
// users, err := db.Table("users").Offset(10000).Limit(100).Get()
```

### 2. 事务粒度

```go
// ✅ 适中的事务粒度
err := db.Transaction(func(tx db.TransactionInterface) error {
    // 只包含相关的操作
    userID, err := tx.Table("users").Insert(userData)
    if err != nil {
        return err
    }
    
    _, err = tx.Table("profiles").Insert(map[string]interface{}{
        "user_id": userID,
    })
    return err
})
```

### 3. 预编译查询

```go
// 对于重复执行的查询，使用预编译
stmt, err := db.Prepare("SELECT * FROM users WHERE status = ? AND age > ?")
defer stmt.Close()

for _, condition := range conditions {
    rows, err := stmt.Query(condition.Status, condition.MinAge)
    // 处理结果
}
```

## 🔗 相关文档

- [查询构建器](Query-Builder) - 查询优化技巧
- [缓存系统](Caching) - 缓存策略
- [配置](Configuration) - 数据库配置优化
- [故障排除](Troubleshooting) - 性能问题排查 