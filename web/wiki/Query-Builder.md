# 查询构建器

TORM 提供强大的链式查询构建器，支持参数化查询、跨数据库兼容等特性。

## 🚀 快速开始

```go
import "github.com/zhoudm1743/torm"

// 创建查询
users, err := torm.Table("users").
    Where("status", "=", "active").
    OrderBy("created_at", "desc").
    Limit(10).
    Get()
```

## 📝 SELECT 查询

### 基础查询

```go
// 查询所有记录
users, err := torm.Table("users").Get()

// 选择特定字段
users, err := torm.Table("users").
    Select("id", "name", "email").
    Get()

// 查询单条记录
user, err := torm.Table("users").
    Where("id", "=", 1).
    First()

// 检查记录是否存在
exists, err := torm.Table("users").
    Where("email", "=", "test@example.com").
    Exists()

// DISTINCT 查询
cities, err := torm.Table("users").
    Select("city").
    Distinct().
    Get()
```

### 原生 SELECT 表达式

```go
// SelectRaw 添加原生表达式
results, err := torm.Table("orders").
    Select("user_id").
    SelectRaw("COUNT(*) as order_count").
    SelectRaw("SUM(amount) as total_amount").
    GroupBy("user_id").
    Get()
```

## 🔍 WHERE 条件

### 基础条件

```go
// 三参数格式: 字段, 操作符, 值
users, err := torm.Table("users").
    Where("status", "=", "active").
    Where("age", ">=", 18).
    Get()

// 参数化查询
users, err := torm.Table("users").
    Where("status = ? AND age >= ?", "active", 18).
    Get()

// 简单等于条件
users, err := torm.Table("users").
    Where("status", "active"). // 等同于 Where("status", "=", "active")
    Get()
```

### OR 条件

```go
users, err := torm.Table("users").
    Where("status", "=", "active").
    OrWhere("vip_level", ">", 3).
    Get()

// 参数化 OR 查询
users, err := torm.Table("users").
    Where("status = ?", "active").
    OrWhere("vip_level > ?", 3).
    Get()
```

### IN 条件

```go
// WhereIn
users, err := torm.Table("users").
    WhereIn("status", []interface{}{"active", "premium"}).
    Get()

// 参数化 IN 查询（推荐）
users, err := torm.Table("users").
    Where("status IN (?)", []string{"active", "premium"}).
    Get()

// WhereNotIn
users, err := torm.Table("users").
    WhereNotIn("status", []interface{}{"deleted", "banned"}).
    Get()
```

### BETWEEN 条件

```go
// WhereBetween
users, err := torm.Table("users").
    WhereBetween("age", 18, 65).
    Get()

// 参数化 BETWEEN
users, err := torm.Table("users").
    Where("age BETWEEN ? AND ?", 18, 65).
    Get()

// WhereNotBetween
users, err := torm.Table("users").
    WhereNotBetween("age", 0, 17).
    Get()
```

### NULL 条件

```go
// WhereNull
users, err := torm.Table("users").
    WhereNull("deleted_at").
    Get()

// WhereNotNull
users, err := torm.Table("users").
    WhereNotNull("email_verified_at").
    Get()
```

### 原生 WHERE

```go
// WhereRaw
users, err := torm.Table("users").
    WhereRaw("YEAR(created_at) = ?", 2024).
    Get()
```

## 🔗 JOIN 查询

```go
// INNER JOIN
results, err := torm.Table("users").
    Join("profiles", "profiles.user_id = users.id").
    Select("users.name", "profiles.avatar").
    Get()

// LEFT JOIN
results, err := torm.Table("users").
    LeftJoin("orders", "orders.user_id = users.id").
    Select("users.*", "COUNT(orders.id) as order_count").
    GroupBy("users.id").
    Get()

// RIGHT JOIN
results, err := torm.Table("users").
    RightJoin("orders", "orders.user_id = users.id").
    Get()

// CROSS JOIN
results, err := torm.Table("users").
    CrossJoin("roles").
    Get()
```

## 📊 排序和分组

### 排序

```go
// 单字段排序
users, err := torm.Table("users").
    OrderBy("created_at", "desc").
    Get()

// 多字段排序
users, err := torm.Table("users").
    OrderBy("status", "asc").
    OrderBy("created_at", "desc").
    Get()

// 原生排序
users, err := torm.Table("users").
    OrderByRaw("FIELD(status, ?, ?, ?)", "premium", "active", "trial").
    Get()
```

### 分组

```go
// GroupBy
results, err := torm.Table("orders").
    Select("user_id", "COUNT(*) as count", "SUM(amount) as total").
    GroupBy("user_id").
    Get()

// Having
results, err := torm.Table("orders").
    Select("user_id", "COUNT(*) as count").
    GroupBy("user_id").
    Having("count", ">", 10).
    Get()

// HavingRaw
results, err := torm.Table("orders").
    Select("user_id", "SUM(amount) as total").
    GroupBy("user_id").
    HavingRaw("SUM(amount) > ?", 1000).
    Get()
```

## 📄 分页和限制

### Limit 和 Offset

```go
// 限制返回数量
users, err := torm.Table("users").
    Limit(10).
    Get()

// 限制和偏移
users, err := torm.Table("users").
    Limit(10).
    Offset(20).
    Get()

// Page 方法
users, err := torm.Table("users").
    Page(2, 10). // 第2页，每页10条
    Get()
```

### 分页查询

```go
// Paginate 返回分页结果
result, err := torm.Table("users").
    Where("status", "=", "active").
    OrderBy("created_at", "desc").
    Paginate(1, 20) // 第1页，每页20条

// 分页结果
fmt.Printf("总数: %d\n", result.Total)
fmt.Printf("当前页: %d\n", result.CurrentPage)
fmt.Printf("总页数: %d\n", result.LastPage)
fmt.Printf("每页数量: %d\n", result.PerPage)

for _, user := range result.Data {
    fmt.Printf("用户: %v\n", user["name"])
}

// 简单分页（不查询总数，性能更好）
result, err := torm.Table("users").
    SimplePaginate(1, 20)
```

## 📈 聚合查询

```go
// Count
count, err := torm.Table("users").
    Where("status", "=", "active").
    Count()

// Sum
total, err := torm.Table("orders").Sum("amount")

// Avg
avg, err := torm.Table("products").Avg("price")

// Max
max, err := torm.Table("products").Max("price")

// Min
min, err := torm.Table("products").Min("price")
```

## ✏️ INSERT 操作

```go
// 单条插入
id, err := torm.Table("users").Insert(map[string]interface{}{
    "name":   "张三",
    "email":  "zhangsan@example.com",
    "status": "active",
})

// 插入并获取 ID
id, err := torm.Table("users").InsertGetId(map[string]interface{}{
    "name":  "李四",
    "email": "lisi@example.com",
})

// 批量插入
count, err := torm.Table("users").InsertBatch([]map[string]interface{}{
    {"name": "用户1", "email": "user1@example.com"},
    {"name": "用户2", "email": "user2@example.com"},
    {"name": "用户3", "email": "user3@example.com"},
})
```

## 🔄 UPDATE 操作

```go
// 条件更新
affected, err := torm.Table("users").
    Where("id", "=", 1).
    Update(map[string]interface{}{
        "name":   "新名字",
        "status": "updated",
    })

// 自增
affected, err := torm.Table("users").
    Where("id", "=", 1).
    Increment("login_count", 1)

// 自减
affected, err := torm.Table("products").
    Where("id", "=", 1).
    Decrement("stock", 5)
```

## 🗑️ DELETE 操作

```go
// 条件删除
affected, err := torm.Table("users").
    Where("status", "=", "deleted").
    Delete()

// 批量删除
affected, err := torm.Table("users").
    WhereIn("id", []interface{}{1, 2, 3}).
    Delete()
```

## ⏱️ 上下文支持

```go
import "context"

// 设置上下文
ctx := context.Background()
users, err := torm.Table("users").
    WithContext(ctx).
    Get()

// 设置超时
users, err := torm.Table("users").
    WithTimeout(5 * time.Second).
    Get()

// 设置截止时间
deadline := time.Now().Add(10 * time.Second)
users, err := torm.Table("users").
    WithDeadline(deadline).
    Get()
```

## 💾 查询缓存

```go
// 启用缓存
users, err := torm.Table("users").
    Where("status", "=", "active").
    Cache(5 * time.Minute).
    Get()

// 带标签的缓存
users, err := torm.Table("users").
    CacheWithTags(5*time.Minute, "users", "active_users").
    Get()

// 清除缓存
torm.ClearCacheByTags("users")
torm.ClearAllCache()
```

## 🔧 SQL 调试

```go
// 获取生成的 SQL
sql, args := torm.Table("users").
    Where("status = ? AND age >= ?", "active", 18).
    OrderBy("created_at", "desc").
    Limit(10).
    ToSQL()

fmt.Printf("SQL: %s\n", sql)
fmt.Printf("参数: %v\n", args)
// 输出: SQL: SELECT * FROM users WHERE status = ? AND age >= ? ORDER BY created_at DESC LIMIT 10
// 输出: 参数: [active 18]
```

## 🌐 跨数据库兼容

```go
// 相同的代码，自动适配不同数据库
users, err := torm.Table("users").
    Where("status = ? AND age >= ?", "active", 18).
    Get()

// MySQL:      SELECT * FROM users WHERE status = ? AND age >= ?
// PostgreSQL: SELECT * FROM users WHERE status = $1 AND age >= $2
// SQLite:     SELECT * FROM users WHERE status = ? AND age >= ?
```

## 📝 原生 SQL

```go
// 原生查询
result, err := torm.RawQuery("SELECT * FROM users WHERE status = ?", "active")
rows, err := result.ScanAll()

// 原生执行
result, err := torm.RawExec("UPDATE users SET status = ? WHERE id = ?", "inactive", 1)
affected, err := result.RowsAffected()

// 在指定连接上执行
result, err := torm.RawQueryOn("mysql", "SELECT * FROM users")
result, err := torm.RawExecOn("postgres", "UPDATE users SET status = $1", "active")
```

## 🔄 克隆查询

```go
// 克隆查询构建器
baseQuery, _ := torm.Table("users")
baseQuery = baseQuery.Where("status", "=", "active")

// 克隆后添加不同条件
query1 := baseQuery.Clone().Where("age", ">=", 18)
query2 := baseQuery.Clone().Where("vip_level", ">", 0)

users1, _ := query1.Get()
users2, _ := query2.Get()
```

## 🎯 最佳实践

### 1. 使用参数化查询

```go
// ✅ 推荐：参数化查询（安全）
users, err := torm.Table("users").
    Where("name = ?", userInput).
    Get()

// ❌ 危险：字符串拼接
// sql := "SELECT * FROM users WHERE name = '" + userInput + "'"
```

### 2. 只查询需要的字段

```go
// ✅ 推荐：只查询需要的字段
users, err := torm.Table("users").
    Select("id", "name", "email").
    Get()

// ❌ 不推荐：查询所有字段
users, err := torm.Table("users").Get()
```

### 3. 使用索引优化

```go
// ✅ 推荐：利用索引字段
users, err := torm.Table("users").
    Where("email", "=", email). // email 应该有索引
    First()
```

### 4. 分页处理大数据

```go
// ✅ 推荐：分页处理
for page := 1; ; page++ {
    result, err := torm.Table("users").
        Paginate(page, 100)
    
    if err != nil || len(result.Data) == 0 {
        break
    }
    
    // 处理当前页数据
    processUsers(result.Data)
}
```
