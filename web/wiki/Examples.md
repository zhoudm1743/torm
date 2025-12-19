# 示例代码

本页面提供 TORM 框架的完整使用示例，涵盖基础 CRUD、复杂查询和模型使用。

## 🚀 快速开始示例

```go
package main

import (
    "fmt"
    "time"
    "github.com/zhoudm1743/torm"
)

// 定义用户模型
type User struct {
    torm.BaseModel
    ID        int       `json:"id" torm:"primary_key,auto_increment"`
    Username  string    `json:"username" torm:"type:varchar,size:50,unique,index"`
    Email     string    `json:"email" torm:"type:varchar,size:100,unique"`
    Age       int       `json:"age" torm:"type:int,unsigned,default:0"`
    Status    string    `json:"status" torm:"type:varchar,size:20,default:active,index"`
    CreatedAt time.Time `json:"created_at" torm:"auto_create_time"`
    UpdatedAt time.Time `json:"updated_at" torm:"auto_update_time"`
}

func main() {
    // 1. 配置数据库连接
    err := torm.AddConnection("default", &torm.Config{
        Driver:   "sqlite",
        Database: "quickstart.db",
    })
    if err != nil {
        panic(err)
    }

    // 2. 自动创建表结构
    user := &User{}
    if err := user.AutoMigrate(); err != nil {
        panic(err)
    }

    // 3. 创建用户
    newUser := &User{
        Username: "zhangsan",
        Email:    "zhangsan@example.com",
        Age:      25,
        Status:   "active",
    }
    if err := newUser.Save(); err != nil {
        panic(err)
    }
    fmt.Printf("用户创建成功，ID: %d\n", newUser.ID)

    // 4. 查询用户
    users, err := torm.Table("users").
        Where("status", "=", "active").
        Where("age", ">=", 18).
        Get()
    if err != nil {
        panic(err)
    }
    fmt.Printf("查询到 %d 个活跃用户\n", len(users))
}
```

## 📝 基础 CRUD 示例

### 创建记录

```go
// 方式1：使用模型 Save
user := &User{
    Username: "zhangsan",
    Email:    "zhangsan@example.com",
    Age:      25,
}
err := user.Save()

// 方式2：使用查询构建器 Insert
id, err := torm.Table("users").Insert(map[string]interface{}{
    "username": "lisi",
    "email":    "lisi@example.com",
    "age":      30,
})

// 方式3：批量插入
users := []map[string]interface{}{
    {"username": "user1", "email": "user1@example.com", "age": 20},
    {"username": "user2", "email": "user2@example.com", "age": 22},
}
count, err := torm.Table("users").InsertBatch(users)
```

### 查询记录

```go
// 查询所有记录
users, err := torm.Table("users").Get()

// 查询单条记录
user, err := torm.Table("users").Where("id", "=", 1).First()

// 根据主键查找
user, err := torm.Table("users").Find(1)

// 查询指定字段
users, err := torm.Table("users").
    Select("id", "username", "email").
    Get()

// 检查记录是否存在
exists, err := torm.Table("users").
    Where("email", "=", "test@example.com").
    Exists()
```

### 更新记录

```go
// 使用查询构建器 Update
affected, err := torm.Table("users").
    Where("id", "=", 1).
    Update(map[string]interface{}{
        "age":    26,
        "status": "premium",
    })

// 递增/递减
affected, err := torm.Table("users").
    Where("id", "=", 1).
    Increment("age", 1)
```

### 删除记录

```go
// 使用查询构建器 Delete
affected, err := torm.Table("users").
    Where("status", "=", "inactive").
    Delete()
```

## 🔍 复杂查询示例

### 条件查询

```go
// 基础条件
users, _ := torm.Table("users").
    Where("status", "=", "active").
    Where("age", ">=", 18).
    Get()

// 参数化查询
users, _ := torm.Table("users").
    Where("status = ? AND age >= ?", "active", 18).
    Get()

// OR 条件
users, _ := torm.Table("users").
    Where("status", "=", "active").
    OrWhere("vip_level", ">", 3).
    Get()

// IN 查询
users, _ := torm.Table("users").
    Where("status IN (?)", []string{"active", "premium"}).
    Get()

// BETWEEN 查询
users, _ := torm.Table("users").
    Where("age BETWEEN ? AND ?", 18, 65).
    Get()
```

### 排序和分页

```go
// 排序
users, _ := torm.Table("users").
    OrderBy("created_at", "desc").
    Get()

// 分页
result, _ := torm.Table("users").
    Where("status", "=", "active").
    Paginate(1, 20)
```

### 聚合查询

```go
// 计数
count, _ := torm.Table("users").
    Where("status", "=", "active").
    Count()

// 求和
total, _ := torm.Table("orders").Sum("amount")

// 平均值
avg, _ := torm.Table("users").Avg("age")
```

### JOIN 查询

```go
// LEFT JOIN
results, _ := torm.Table("users").
    LeftJoin("profiles", "profiles.user_id", "=", "users.id").
    Select("users.username", "profiles.avatar").
    Get()
```

## 💼 事务示例

```go
err := torm.Transaction(func(tx torm.TransactionInterface) error {
    _, err := tx.Exec("INSERT INTO users (username, email) VALUES (?, ?)",
        "zhangsan", "zhangsan@example.com")
    if err != nil {
        return err // 自动回滚
    }

    _, err = tx.Exec("UPDATE accounts SET balance = balance - ? WHERE user_id = ?",
        100, 1)
    if err != nil {
        return err // 自动回滚
    }

    return nil // 自动提交
})
```

## 💾 缓存示例

```go
// 缓存查询结果
users, _ := torm.Table("users").
    Where("status", "=", "active").
    Cache(5 * time.Minute).
    Get()

// 清除缓存
torm.ClearCacheByTags("users")
```

## 🌐 跨数据库示例

```go
// MySQL
torm.AddConnection("mysql", &torm.Config{
    Driver:   "mysql",
    Host:     "localhost",
    Database: "myapp",
    Username: "root",
    Password: "password",
})

// PostgreSQL
torm.AddConnection("postgres", &torm.Config{
    Driver:   "postgres",
    Host:     "localhost",
    Database: "myapp",
    Username: "postgres",
    Password: "password",
    SSLMode:  "disable",
})

// SQLite
torm.AddConnection("sqlite", &torm.Config{
    Driver:   "sqlite",
    Database: "app.db",
})
```
