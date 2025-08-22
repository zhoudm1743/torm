# 查询构建器

TORM 提供了强大而灵活的查询构建器，支持链式调用、类型安全的SQL构建和复杂查询操作。

## 📋 目录

- [基础查询](#基础查询)
- [条件查询](#条件查询)
- [连接查询](#连接查询)
- [聚合查询](#聚合查询)
- [排序和分页](#排序和分页)
- [高级查询](#高级查询)
- [原生SQL](#原生sql)
- [查询优化](#查询优化)

## 🚀 快速开始

### 创建查询构建器

```go
// 方法1：直接创建
query, err := db.Table("users", "default")
if err != nil {
    log.Fatal(err)
}

// 方法2：指定连接
query, err := db.Table("users", "mysql_connection")
if err != nil {
    log.Fatal(err)
}
```

## 📊 基础查询

### SELECT 查询

```go
// 查询所有字段
users, err := query.Get()

// 查询指定字段
users, err := query.Select("id", "name", "email").Get()

// 查询单条记录
user, err := query.Where("id", "=", 1).First()

// 根据主键查找
user, err := query.Find(1)

// 检查记录是否存在
exists, err := query.Where("email", "=", "user@example.com").Exists()
```

### INSERT 操作

```go
// 插入单条记录
id, err := query.Insert(map[string]interface{}{
    "name":  "张三",
    "email": "user1@example.com",
    "age":   25,
})

// 批量插入
users := []map[string]interface{}{
    {"name": "李四", "email": "user2@example.com", "age": 30},
    {"name": "王五", "email": "user3@example.com", "age": 28},
}
affected, err := query.InsertBatch(users)

// 插入或忽略
id, err := query.InsertIgnore(map[string]interface{}{
    "email": "unique@example.com",
    "name":  "用户",
})
```

### UPDATE 操作

```go
// 更新记录
affected, err := query.
    Where("id", "=", 1).
    Update(map[string]interface{}{
        "name": "新名字",
        "age":  26,
    })

// 条件更新
affected, err := query.
    Where("status", "=", "inactive").
    Where("last_login", "<", "2023-01-01").
    Update(map[string]interface{}{
        "status": "archived",
    })

// 递增/递减
affected, err := query.Where("id", "=", 1).Increment("views", 1)
affected, err := query.Where("id", "=", 1).Decrement("score", 5)
```

### DELETE 操作

```go
// 删除记录
affected, err := query.Where("id", "=", 1).Delete()

// 条件删除
affected, err := query.
    Where("status", "=", "inactive").
    Where("created_at", "<", "2022-01-01").
    Delete()

// 清空表（谨慎使用）
affected, err := query.Truncate()
```

## 🔍 条件查询

### 基础条件

```go
// 等于
query.Where("status", "=", "active")
query.Where("age", ">=", 18)
query.Where("name", "LIKE", "%张%")

// 多个条件（AND）
query.Where("status", "=", "active").
      Where("age", ">=", 18).
      Where("city", "=", "北京")
```

### OR 条件

```go
// OR 条件
query.Where("status", "=", "active").
      OrWhere("vip_level", ">", 3)

// 复杂 OR 条件
query.Where(func(q db.QueryInterface) db.QueryInterface {
    return q.Where("status", "=", "active").
             OrWhere("vip_level", ">", 3)
}).Where("city", "=", "北京")
```

### NULL 值查询

```go
// 查询NULL值
query.WhereNull("deleted_at")
query.WhereNull("phone")

// 查询非NULL值  
query.WhereNotNull("email")
query.WhereNotNull("avatar")

// 组合使用
query.WhereNotNull("email").
      WhereNull("deleted_at").
      Where("status", "=", "active")
```

### BETWEEN 范围查询

```go
// BETWEEN 查询
query.WhereBetween("age", []interface{}{18, 65})
query.WhereBetween("created_at", []interface{}{"2024-01-01", "2024-12-31"})

// NOT BETWEEN 查询
query.WhereNotBetween("score", []interface{}{0, 60})

// 结合其他条件
query.WhereBetween("age", []interface{}{25, 45}).
      WhereNotNull("email").
      Where("status", "=", "active")
```

### EXISTS 子查询

```go
// EXISTS 查询
subQuery := "SELECT 1 FROM orders WHERE orders.user_id = users.id"
query.WhereExists(subQuery)

// NOT EXISTS 查询
query.WhereNotExists("SELECT 1 FROM banned_users WHERE banned_users.user_id = users.id")

// 使用查询构建器作为子查询
subQuery, _ := db.Table("orders").
    Select("1").
    WhereRaw("orders.user_id = users.id")
query.WhereExists(subQuery)
```

### IN 和 NOT IN

```go
// IN 查询
query.WhereIn("id", []interface{}{1, 2, 3, 4, 5})
query.WhereIn("status", []interface{}{"active", "pending"})

// NOT IN 查询
query.WhereNotIn("status", []interface{}{"deleted", "banned"})
```

### BETWEEN

```go
// BETWEEN 查询
query.WhereBetween("age", 18, 65)
query.WhereBetween("created_at", "2023-01-01", "2023-12-31")

// NOT BETWEEN
query.WhereNotBetween("score", 0, 60)
```

### NULL 检查

```go
// IS NULL
query.WhereNull("deleted_at")

// IS NOT NULL
query.WhereNotNull("email_verified_at")
```

### 日期条件

```go
// 日期查询
query.WhereDate("created_at", "2023-06-15")
query.WhereYear("created_at", 2023)
query.WhereMonth("created_at", 6)
query.WhereDay("created_at", 15)

// 时间范围
query.Where("created_at", ">=", "2023-01-01").
      Where("created_at", "<=", "2023-12-31")
```

## 🔗 连接查询

### INNER JOIN

```go
users, err := query.
    Select("users.name", "profiles.avatar", "posts.title").
    Join("profiles", "profiles.user_id", "=", "users.id").
    Join("posts", "posts.user_id", "=", "users.id").
    Where("users.status", "=", "active").
    Get()
```

### LEFT JOIN

```go
users, err := query.
    Select("users.*", "profiles.avatar").
    LeftJoin("profiles", "profiles.user_id", "=", "users.id").
    Get()
```

### RIGHT JOIN

```go
users, err := query.
    Select("users.*", "orders.total").
    RightJoin("orders", "orders.user_id", "=", "users.id").
    Get()
```

### 复杂连接

```go
users, err := query.
    Select("users.name", "COUNT(posts.id) as post_count").
    LeftJoin("posts", func(join db.JoinClause) {
        join.On("posts.user_id", "=", "users.id").
             Where("posts.status", "=", "published")
    }).
    GroupBy("users.id").
    Having("post_count", ">", 5).
    Get()
```

## 📈 聚合查询

### 基础聚合

```go
// 计数
count, err := query.Where("status", "=", "active").Count()

// 求和
totalAge, err := query.Sum("age")

// 平均值
avgAge, err := query.Avg("age")

// 最大值和最小值
maxAge, err := query.Max("age")
minAge, err := query.Min("age")
```

### GROUP BY 和 HAVING

```go
// 分组统计
result, err := query.
    Select("city", "COUNT(*) as user_count", "AVG(age) as avg_age").
    GroupBy("city").
    Having("user_count", ">", 100).
    OrderBy("user_count", "desc").
    Get()

// 多字段分组
result, err := query.
    Select("city", "status", "COUNT(*) as count").
    GroupBy("city", "status").
    Get()
```

## 📋 排序和分页

### 排序

```go
// 单字段排序
query.OrderBy("created_at", "desc")
query.OrderBy("name", "asc")

// 多字段排序
query.OrderBy("status", "asc").
      OrderBy("created_at", "desc")

// 随机排序（跨数据库兼容）
query.OrderRand()

// 按字段值优先级排序
statusOrder := []interface{}{"premium", "active", "trial", "inactive"}
query.OrderField("status", statusOrder, "asc")

// 原生排序表达式
query.OrderByRaw("RAND()")
query.OrderByRaw("FIELD(status, ?, ?, ?)", "active", "pending", "inactive")

// 添加原生字段表达式
query.FieldRaw("COUNT(*) as order_count").
      GroupBy("user_id").
      OrderBy("order_count", "desc")
```

### 分页

```go
// 基础分页
users, err := query.
    Where("status", "=", "active").
    OrderBy("created_at", "desc").
    Limit(10).
    Offset(20).
    Get()

// 使用分页器
result, err := query.
    Where("status", "=", "active").
    Paginate(2, 10) // 第2页，每页10条

// 分页结果包含：
// result.Data      - 数据
// result.Total     - 总记录数
// result.Page      - 当前页
// result.PerPage   - 每页数量
// result.LastPage  - 最后一页
```

## 🚀 高级查询

### 子查询

```go
// EXISTS 子查询
users, err := query.
    Where("status", "=", "active").
    WhereExists(func(q db.QueryInterface) db.QueryInterface {
        return q.Table("orders").
                 Where("orders.user_id", "=", "users.id").
                 Where("orders.status", "=", "completed")
    }).Get()

// IN 子查询
users, err := query.
    WhereIn("id", func(q db.QueryInterface) db.QueryInterface {
        return q.Table("orders").
                 Select("user_id").
                 Where("total", ">", 1000)
    }).Get()
```

### 条件构建器

```go
// 动态条件构建
query := db.Table("users")

if status != "" {
    query = query.Where("status", "=", status)
}

if minAge > 0 {
    query = query.Where("age", ">=", minAge)
}

if city != "" {
    query = query.Where("city", "=", city)
}

users, err := query.Get()
```

### UNION 查询

```go
// UNION 查询
activeUsers := db.Table("users").Where("status", "=", "active")
vipUsers := db.Table("users").Where("vip_level", ">", 3)

users, err := activeUsers.Union(vipUsers).Get()

// UNION ALL
users, err := activeUsers.UnionAll(vipUsers).Get()
```

## 💾 原生SQL

### 原生查询

```go
// 原生 SELECT
users, err := db.Raw("SELECT * FROM users WHERE age > ? AND city = ?", 18, "北京")

// 原生 INSERT
result, err := db.Exec("INSERT INTO users (name, email) VALUES (?, ?)", "张三", "user@example.com")

// 原生查询与构建器结合
users, err := query.
    WhereRaw("YEAR(created_at) = ?", 2023).
    OrderByRaw("FIELD(status, 'active', 'pending', 'inactive')").
    Get()
```

### 复杂原生查询

```go
// 复杂统计查询
sql := `
    SELECT 
        DATE(created_at) as date,
        COUNT(*) as user_count,
        COUNT(CASE WHEN status = 'active' THEN 1 END) as active_count
    FROM users 
    WHERE created_at >= ? AND created_at <= ?
    GROUP BY DATE(created_at)
    ORDER BY date DESC
`
result, err := db.Raw(sql, startDate, endDate)
```

## ⚡ 查询优化

### 查询提示

```go
// 强制使用索引
users, err := query.
    WhereRaw("USE INDEX (idx_email)").
    Where("email", "=", "user@example.com").
    Get()

// 禁用查询缓存
users, err := query.
    WhereRaw("SQL_NO_CACHE").
    Get()
```

### 预编译查询

```go
// 预编译查询语句
stmt, err := db.Prepare("SELECT * FROM users WHERE age > ? AND city = ?")
if err != nil {
    return err
}
defer stmt.Close()

// 执行预编译查询
users, err := stmt.Query(18, "北京")
```

### 批量操作优化

```go
// 分批处理大量数据
query.Chunk(1000, func(users []map[string]interface{}) bool {
    // 处理每批1000条数据
    for _, user := range users {
        // 处理单条用户数据
        processUser(user)
    }
    return true // 返回 true 继续，false 停止
})
```

## 🔧 高级功能

### 查询作用域

```go
// 定义查询作用域
func ActiveUsers(q db.QueryInterface) db.QueryInterface {
    return q.Where("status", "=", "active")
}

func AdultUsers(q db.QueryInterface) db.QueryInterface {
    return q.Where("age", ">=", 18)
}

// 使用作用域
users, err := query.
    Scope(ActiveUsers).
    Scope(AdultUsers).
    Get()
```

### 查询监听器

```go
// 添加查询监听器
db.Listen(func(sql string, bindings []interface{}, duration time.Duration) {
    log.Printf("SQL: %s, Bindings: %v, Duration: %v", sql, bindings, duration)
})
```

### 查询缓存

```go
// 启用查询缓存
users, err := query.
    Where("status", "=", "active").
    Cache(5 * time.Minute). // 缓存5分钟
    Get()

// 缓存标签
users, err := query.
    Where("status", "=", "active").
    CacheWithTags(5*time.Minute, "users", "active").
    Get()

// 清除缓存
db.FlushCache("users")
```

## 🐛 调试和分析

### SQL 调试

```go
// 打印 SQL 而不执行
sql, bindings := query.
    Where("status", "=", "active").
    ToSQL()
fmt.Printf("SQL: %s\nBindings: %v\n", sql, bindings)

// 启用查询日志
db.EnableQueryLog()
users, err := query.Get()
logs := db.GetQueryLog()
for _, log := range logs {
    fmt.Printf("SQL: %s, Time: %v\n", log.SQL, log.Duration)
}
```

### 性能分析

```go
// 查询性能分析
start := time.Now()
users, err := query.
    Where("status", "=", "active").
    Get()
duration := time.Since(start)
log.Printf("Query took: %v", duration)

// EXPLAIN 查询
explain, err := query.
    Where("status", "=", "active").
    Explain()
fmt.Printf("Query plan: %+v\n", explain)
```

## 📚 最佳实践

### 1. 索引优化

```go
// 好的做法：利用索引
users, err := query.
    Where("email", "=", email).    // email 应该有索引
    Where("status", "=", "active"). // status 可以是复合索引的一部分
    Get()

// 避免：在索引字段上使用函数
// 不好：WHERE UPPER(email) = 'USER@EXAMPLE.COM'
// 好的：WHERE email = 'user@example.com'
```

### 2. 分页优化

```go
// 对于大数据量，使用游标分页
users, err := query.
    Where("id", ">", lastID).
    OrderBy("id", "asc").
    Limit(100).
    Get()
```

### 3. 避免 N+1 查询

```go
// 不好的做法
users, err := query.Get()
for _, user := range users {
    // 每个用户都会执行一次查询
    posts, _ := db.Table("posts").Where("user_id", "=", user["id"]).Get()
}

// 好的做法：使用 JOIN 或预加载
users, err := query.
    LeftJoin("posts", "posts.user_id", "=", "users.id").
    Select("users.*", "posts.title", "posts.content").
    Get()
```

### 4. 安全性

```go
// 使用参数绑定防止 SQL 注入
// 好的做法
users, err := query.Where("name", "=", userInput).Get()

// 避免字符串拼接
// 危险的做法
// sql := "SELECT * FROM users WHERE name = '" + userInput + "'"
```

## 🔗 相关文档

- [模型系统](Model-System) - 了解如何在模型中使用查询构建器
- [关联关系](Relationships) - 处理表之间的关联关系
- [性能优化](Performance) - 查询性能优化指南
- [故障排除](Troubleshooting) - 常见查询问题解决方案 