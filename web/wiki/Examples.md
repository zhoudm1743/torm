# 示例代码

本页面提供 TORM 框架的完整使用示例，涵盖基础 CRUD、复杂查询和模型使用。

## 🚀 快速开始

```go
package main

import (
    "fmt"
    "github.com/zhoudm1743/torm"
)

func main() {
    // 1. 配置数据库连接
    torm.AddConnection("default", &torm.Config{
        Driver:   "mysql",
        Host:     "localhost",
        Port:     3306,
        Database: "myapp",
        Username: "root",
        Password: "password",
        Charset:  "utf8mb4",
    })
    
    // 2. 查询数据
    users, _ := torm.Table("users").
        Where("status", "=", "active").
        OrderBy("created_at", "desc").
        Limit(10).
        Get()
    
    fmt.Printf("查询到 %d 个用户\n", len(users))
    
    // 3. 插入数据
    id, _ := torm.Table("users").Insert(map[string]interface{}{
        "name":   "张三",
        "email":  "zhangsan@example.com",
        "status": "active",
    })
    
    fmt.Printf("新用户 ID: %d\n", id)
}
```

## 📝 基础 CRUD

### 创建记录

```go
// 单条插入
id, err := torm.Table("users").Insert(map[string]interface{}{
    "name":   "张三",
    "email":  "zhangsan@example.com",
    "age":    25,
    "status": "active",
})

// 批量插入
users := []map[string]interface{}{
    {"name": "用户1", "email": "user1@example.com", "age": 20},
    {"name": "用户2", "email": "user2@example.com", "age": 22},
    {"name": "用户3", "email": "user3@example.com", "age": 24},
}
count, err := torm.Table("users").InsertBatch(users)
fmt.Printf("插入了 %d 条记录\n", count)
```

### 查询记录

```go
// 查询所有
users, err := torm.Table("users").Get()

// 查询单条
user, err := torm.Table("users").Where("id", "=", 1).First()

// 根据主键查找
user, err := torm.Table("users").Find(1)

// 选择特定字段
users, err := torm.Table("users").
    Select("id", "name", "email").
    Get()

// 检查是否存在
exists, err := torm.Table("users").
    Where("email", "=", "test@example.com").
    Exists()
```

### 更新记录

```go
// 更新数据
affected, err := torm.Table("users").
    Where("id", "=", 1).
    Update(map[string]interface{}{
        "name":   "李四",
        "status": "premium",
    })

// 递增
affected, err := torm.Table("users").
    Where("id", "=", 1).
    Increment("points", 10)

// 递减
affected, err := torm.Table("users").
    Where("id", "=", 1).
    Decrement("balance", 100)
```

### 删除记录

```go
// 删除记录
affected, err := torm.Table("users").
    Where("status", "=", "inactive").
    Delete()

// 清空表
err := torm.Table("logs").Truncate()
```

## 🔍 条件查询

### 基础条件

```go
// 等于
users, _ := torm.Table("users").Where("status", "=", "active").Get()

// 不等于
users, _ := torm.Table("users").Where("status", "!=", "banned").Get()

// 大于/小于
users, _ := torm.Table("users").Where("age", ">=", 18).Get()

// OR 条件
users, _ := torm.Table("users").
    Where("status", "=", "active").
    OrWhere("vip_level", ">", 3).
    Get()
```

### 特殊条件

```go
// IN 查询
users, _ := torm.Table("users").
    WhereIn("status", []interface{}{"active", "pending"}).
    Get()

// NOT IN 查询
users, _ := torm.Table("users").
    WhereNotIn("status", []interface{}{"banned", "deleted"}).
    Get()

// BETWEEN 查询
users, _ := torm.Table("users").
    WhereBetween("age", []interface{}{18, 65}).
    Get()

// NULL 查询
users, _ := torm.Table("users").WhereNull("deleted_at").Get()
users, _ := torm.Table("users").WhereNotNull("email").Get()

// LIKE 查询
users, _ := torm.Table("users").WhereLike("name", "%张%").Get()

// EXISTS 子查询
users, _ := torm.Table("users").
    WhereExists("SELECT 1 FROM orders WHERE orders.user_id = users.id").
    Get()
```

## 📊 聚合查询

```go
// 计数
count, _ := torm.Table("users").Where("status", "=", "active").Count()

// 求和
total, _ := torm.Table("orders").Sum("amount")

// 平均值
avgAge, _ := torm.Table("users").Avg("age")

// 最大值/最小值
maxPrice, _ := torm.Table("products").Max("price")
minPrice, _ := torm.Table("products").Min("price")
```

## 🔗 JOIN 查询

```go
// INNER JOIN
results, _ := torm.Table("users").
    Join("profiles", "users.id", "=", "profiles.user_id").
    Select("users.name", "profiles.avatar", "profiles.bio").
    Get()

// LEFT JOIN
results, _ := torm.Table("users").
    LeftJoin("orders", "users.id", "=", "orders.user_id").
    Select("users.name", "COUNT(orders.id) as order_count").
    GroupBy("users.id").
    Get()

// 多表 JOIN
results, _ := torm.Table("orders").
    Join("users", "orders.user_id", "=", "users.id").
    Join("products", "orders.product_id", "=", "products.id").
    Select("orders.id", "users.name", "products.title", "orders.amount").
    Get()
```

## 📄 分页查询

```go
// 基础分页
result, _ := torm.Table("users").
    Where("status", "=", "active").
    OrderBy("created_at", "desc").
    Paginate(1, 20) // 第1页，每页20条

fmt.Printf("总数: %d\n", result.Total)
fmt.Printf("当前页: %d\n", result.CurrentPage)
fmt.Printf("总页数: %d\n", result.LastPage)
fmt.Printf("数据: %v\n", result.Data)

// 游标分页（大数据量推荐）
users, _ := torm.Table("users").
    Where("id", ">", lastID).
    OrderBy("id", "asc").
    Limit(20).
    Get()
```

## 💼 事务处理

```go
// 闭包事务（推荐）
err := torm.Transaction(func(tx torm.TransactionInterface) error {
    // 扣减余额
    _, err := tx.Exec(
        "UPDATE accounts SET balance = balance - ? WHERE user_id = ?",
        100, 1,
    )
    if err != nil {
        return err // 自动回滚
    }
    
    // 创建订单
    _, err = tx.Exec(
        "INSERT INTO orders (user_id, amount, status) VALUES (?, ?, ?)",
        1, 100, "paid",
    )
    if err != nil {
        return err // 自动回滚
    }
    
    return nil // 自动提交
})

// 手动事务
tx, _ := torm.BeginTransaction()

_, err := tx.Exec("UPDATE accounts SET balance = balance - ? WHERE user_id = ?", 100, 1)
if err != nil {
    tx.Rollback()
    return err
}

_, err = tx.Exec("INSERT INTO orders (user_id, amount) VALUES (?, ?)", 1, 100)
if err != nil {
    tx.Rollback()
    return err
}

tx.Commit()
```

## 💾 缓存查询

```go
// 缓存查询结果 5 分钟
users, _ := torm.Table("users").
    Where("status", "=", "active").
    Cache(5 * time.Minute).
    Get()

// 带标签的缓存
users, _ := torm.Table("users").
    Where("status", "=", "active").
    CacheWithTags(5*time.Minute, "users", "active").
    Get()

// 清除标签缓存
torm.ClearCacheByTags("users")
```

## 🏗️ 模型使用

```go
import "github.com/zhoudm1743/torm/model"

// 定义模型
type User struct {
    model.BaseModel
}

func NewUser() *User {
    user := &User{BaseModel: *model.NewBaseModel()}
    user.SetTable("users")
    user.EnableTimestamps()
    user.EnableSoftDeletes()
    return user
}

// 使用模型
func main() {
    // 创建
    user := NewUser()
    user.Fill(map[string]interface{}{
        "name":  "张三",
        "email": "zhangsan@example.com",
    })
    user.Save()
    
    // 查询
    user2 := NewUser()
    user2.Find(1)
    fmt.Println(user2.GetAttribute("name"))
    
    // 更新
    user2.SetAttribute("name", "李四")
    user2.Save()
    
    // 软删除
    user2.SoftDelete()
    
    // 恢复
    user2.Restore()
    
    // 强制删除
    user2.ForceDelete()
}
```

## 🔄 关联关系

```go
// 定义关联
func (u *User) Posts() *model.HasMany {
    return u.HasMany(&Post{}, "user_id", "id")
}

func (u *User) Profile() *model.HasOne {
    return u.HasOne(&Profile{}, "user_id", "id")
}

// 预加载
users, _ := user.With("Posts", "Profile").Get()

// 条件预加载
users, _ := user.WithConstraint("Posts", func(q *QueryBuilder) *QueryBuilder {
    return q.Where("status", "=", "published").Limit(5)
}).Get()
```

## 🌐 多数据库

```go
// 配置多个连接
torm.AddConnection("mysql", &torm.Config{
    Driver:   "mysql",
    Host:     "localhost",
    Database: "myapp",
    Username: "root",
    Password: "password",
})

torm.AddConnection("postgres", &torm.Config{
    Driver:   "postgres",
    Host:     "localhost",
    Database: "analytics",
    Username: "postgres",
    Password: "password",
})

torm.AddConnection("sqlite", &torm.Config{
    Driver:   "sqlite",
    Database: "cache.db",
})

// 使用不同连接
users, _ := torm.Table("users", "mysql").Get()
reports, _ := torm.Table("reports", "postgres").Get()
cache, _ := torm.Table("cache", "sqlite").Get()
```

## 🔧 调试

```go
// 获取 SQL 语句
query := torm.Table("users").
    Where("status", "=", "active").
    OrderBy("created_at", "desc")

sql, args := query.ToSQL()
fmt.Printf("SQL: %s\n", sql)
fmt.Printf("Args: %v\n", args)

// 启用 SQL 日志
torm.EnableSQLLogging()
```

## 🔗 相关文档

- [查询构建器](Query-Builder) - 完整查询 API
- [模型系统](Model-System) - 模型详细用法
- [事务处理](Transactions) - 事务管理
- [缓存系统](Caching) - 缓存策略
