# 性能优化

本文档提供 TORM 的性能优化指南，包括查询优化、缓存策略和连接池配置。

## 🚀 查询优化

### 避免 N+1 查询

```go
// ❌ 错误做法 - 产生 N+1 查询
users, _ := torm.Table("users").Get()
for _, user := range users {
    posts, _ := torm.Table("posts").Where("user_id", "=", user["id"]).Get()
}

// ✅ 正确做法 - 使用预加载
users, _ := user.With("Posts").Get()

// ✅ 正确做法 - 使用 JOIN
results, _ := torm.Table("users").
    LeftJoin("posts", "users.id", "=", "posts.user_id").
    Select("users.*", "posts.title").
    Get()
```

### 选择性字段查询

```go
// ❌ 避免查询所有字段
users, _ := torm.Table("users").Get()

// ✅ 只查询需要的字段
users, _ := torm.Table("users").
    Select("id", "name", "email").
    Get()
```

### 使用索引字段

```go
// ✅ 利用索引字段进行查询
users, _ := torm.Table("users").
    Where("email", "=", "user@example.com"). // email 有索引
    Where("status", "=", "active").           // status 是复合索引
    Get()
```

## 💾 缓存策略

### 查询结果缓存

```go
// 缓存频繁查询的结果
users, _ := torm.Table("users").
    Where("status", "=", "active").
    Cache(5 * time.Minute).
    Get()

// 使用标签便于缓存管理
users, _ := torm.Table("users").
    Where("status", "=", "active").
    CacheWithTags(5*time.Minute, "users", "active").
    Get()

// 清除标签缓存
torm.ClearCacheByTags("users")
```

## 🔧 连接池优化

### 连接池配置

```go
config := &torm.Config{
    Driver:   "mysql",
    Host:     "localhost",
    Database: "myapp",
    Username: "root",
    Password: "password",
    
    // 连接池配置
    MaxOpenConns:    100,              // 最大打开连接数
    MaxIdleConns:    10,               // 最大空闲连接数
    ConnMaxLifetime: time.Hour,        // 连接最大生存时间
    ConnMaxIdleTime: 30 * time.Minute, // 连接最大空闲时间
}
```

## 📦 批量操作

### 批量插入

```go
// ✅ 使用批量插入
users := []map[string]interface{}{
    {"name": "用户1", "email": "user1@example.com"},
    {"name": "用户2", "email": "user2@example.com"},
}
affected, _ := torm.Table("users").InsertBatch(users)
```

### 分批处理大数据

```go
// 分批处理大量数据
err := torm.Table("users").Chunk(1000, func(users []map[string]interface{}) bool {
    for _, user := range users {
        processUser(user)
    }
    return true // 继续处理下一批
})
```

## 📄 分页优化

### 游标分页（大数据量推荐）

```go
// ✅ 游标分页 - 适合大数据量
users, _ := torm.Table("users").
    Where("id", ">", lastID).
    OrderBy("id", "asc").
    Limit(100).
    Get()

// ❌ 避免大偏移量
// users, _ := torm.Table("users").Offset(10000).Limit(100).Get()
```

## 📊 性能监控

### 启用查询日志

```go
// 启用 SQL 日志
torm.EnableSQLLogging()

// 获取 SQL 语句
query := torm.Table("users").Where("status", "=", "active")
sql, args := query.ToSQL()
fmt.Printf("SQL: %s, Args: %v\n", sql, args)
```

## 💡 最佳实践

1. **选择性查询**: 只查询需要的字段
2. **使用索引**: 在 WHERE 条件字段上创建索引
3. **批量操作**: 使用 InsertBatch 代替循环插入
4. **游标分页**: 大数据量使用游标分页
5. **合理缓存**: 缓存不常变化的数据
6. **连接池**: 根据负载调整连接池参数

## 🔗 相关文档

- [查询构建器](Query-Builder) - 查询优化技巧
- [缓存系统](Caching) - 缓存策略
- [配置文档](Configuration) - 连接池配置
