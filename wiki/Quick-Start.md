# 快速开始指南

欢迎使用TORM v1.1.0！这个指南将在5分钟内让你体验TORM的强大功能，包括最新的First/Find增强、自定义主键、复合主键等特性。

## 📋 前置要求

- Go 1.19 或更高版本
- 支持的数据库之一 (MySQL, PostgreSQL, SQLite)

## 🚀 第1步：安装TORM

```bash
# 创建新项目
mkdir my-torm-app
cd my-torm-app
go mod init my-torm-app

# 安装TORM
go get github.com/zhoudm1743/torm
```

## 🔧 第2步：创建现代化示例

### 创建模型文件 `models/user.go`

```go
package models

import (
    "time"
    "github.com/zhoudm1743/torm/model"
)

// User 默认主键模型
type User struct {
    model.BaseModel
    ID        interface{} `json:"id" db:"id"`
    Name      string      `json:"name" db:"name"`
    Email     string      `json:"email" db:"email"`
    Age       int         `json:"age" db:"age"`
    Status    string      `json:"status" db:"status"`
    CreatedAt time.Time   `json:"created_at" db:"created_at"`
    UpdatedAt time.Time   `json:"updated_at" db:"updated_at"`
}

// UserWithUUID 使用UUID作为主键的用户模型
type UserWithUUID struct {
    model.BaseModel
    UUID      string    `json:"uuid" db:"uuid" primary:"true"`  // 自定义UUID主键
    Name      string    `json:"name" db:"name"`
    Email     string    `json:"email" db:"email"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserWithCompositePK 使用复合主键的用户模型
type UserWithCompositePK struct {
    model.BaseModel
    TenantID  string    `json:"tenant_id" db:"tenant_id" primary:"true"`   // 复合主键1
    UserID    string    `json:"user_id" db:"user_id" primary:"true"`       // 复合主键2
    Name      string    `json:"name" db:"name"`
    Email     string    `json:"email" db:"email"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// NewUser 创建用户模型
func NewUser() *User {
    user := &User{
        BaseModel: *model.NewBaseModel(),
    }
    user.SetTable("users")
    user.SetConnection("default")
    return user
}

// NewUserWithUUID 创建UUID主键的用户模型
func NewUserWithUUID() *UserWithUUID {
    user := &UserWithUUID{
        BaseModel: *model.NewBaseModel(),
    }
    user.SetTable("users_uuid")
    user.SetConnection("default")
    // 自动检测主键标签
    user.DetectPrimaryKeysFromStruct(user)
    return user
}

// NewUserWithCompositePK 创建复合主键的用户模型
func NewUserWithCompositePK() *UserWithCompositePK {
    user := &UserWithCompositePK{
        BaseModel: *model.NewBaseModel(),
    }
    user.SetTable("users_composite")
    user.SetConnection("default")
    // 自动检测主键标签
    user.DetectPrimaryKeysFromStruct(user)
    return user
}
```

### 创建主文件 `main.go`

```go
package main

import (
    "log"
    
    "github.com/zhoudm1743/torm/db"
)

type User struct {
    ID        int    `db:"id" json:"id"`
    Name      string `db:"name" json:"name"`
    Email     string `db:"email" json:"email"`
    Age       int    `db:"age" json:"age"`
    Status    string `db:"status" json:"status"`
    CreatedAt string `db:"created_at" json:"created_at"`
    UpdatedAt string `db:"updated_at" json:"updated_at"`
}

func main() {
    // 配置数据库连接
    config := &db.Config{
        Driver:   "mysql",
        Host:     "localhost",
        Port:     3306,
        Username: "root",
        Password: "password",
        Database: "torm_example",
    }
    err := db.AddConnection("default", config)
    if err != nil {
        log.Fatal(err)
    }
    
    // ===== 查询构建器演示 =====
    log.Println("===== 查询构建器演示 =====")

    // 基础查询
    query, err := db.Table("users", "default")
    if err == nil {
        // 查询所有用户
        users, err := query.Select("id", "name", "email", "age").
            Where("status", "=", "active").
            OrderBy("created_at", "desc").
            Limit(5).
            Get()
        if err == nil {
            log.Printf("查询到 %d 个活跃用户", len(users))
        }

        // 条件查询
        adults, err := query.Where("age", ">=", 18).
            Where("status", "=", "active").
            Count()
        if err == nil {
            log.Printf("成年活跃用户数量: %d", adults)
        }
    }

    // ===== First和Find新功能演示 =====
    log.Println("===== First和Find新功能演示 =====")
    
    // First方法 - 只填充当前模型
    user1 := models.NewUser()
    _, err = user1.Where("id", "=", 1).First()
    if err != nil {
        log.Printf("查询失败: %v", err)
    } else {
        log.Printf("First结果: Name=%s, Age=%d", user1.Name, user1.Age)
    }

    // First方法 - 同时填充传入的指针
    user2 := models.NewUser()
    var anotherUser models.User
    _, err = user2.Where("id", "=", 2).First(&anotherUser)
    if err != nil {
        log.Printf("查询失败: %v", err)
    } else {
        log.Printf("First + 指针填充: 当前=%s, 指针=%s", user2.Name, anotherUser.Name)
    }

    // Find方法 - 同时填充传入的指针
    user3 := models.NewUser()
    var targetUser models.User
    _, err = user3.Find(1, &targetUser)
    if err != nil {
        log.Printf("Find失败: %v", err)
    } else {
        log.Printf("Find + 指针填充: 当前=%s, 指针=%s", user3.Name, targetUser.Name)
    }

    // ===== db包First和Find方法演示 =====
    log.Println("===== db包First和Find方法演示 =====")
    
    // db.Table().First() 
    query1, err := db.Table("users", "default")
    if err == nil {
        dbResult1, err := query1.Where("id", "=", 1).First()
        if err == nil {
            log.Printf("db.First() 结果: %s", dbResult1["name"])
        }
    }

    // db.Table().First(&model)
    query2, err := db.Table("users", "default")
    if err == nil {
        var userStruct models.User
        _, err := query2.Where("id", "=", 1).First(&userStruct)
        if err == nil {
            log.Printf("db.First(&model) 结果: Name=%s", userStruct.Name)
        }
}

    // ===== 自定义主键功能演示 =====
    log.Println("===== 自定义主键功能演示 =====")

    // 默认主键
    user4 := models.NewUser()
    log.Printf("默认主键: %v", user4.PrimaryKeys())

    // UUID主键
    userUUID := models.NewUserWithUUID()
    userUUID.UUID = "550e8400-e29b-41d4-a716-446655440000"
    userUUID.SetAttribute("uuid", userUUID.UUID)
    log.Printf("UUID主键: %v, 值: %v", userUUID.PrimaryKeys(), userUUID.GetKey())

    // 复合主键
    userComposite := models.NewUserWithCompositePK()
    userComposite.SetAttribute("tenant_id", "tenant-001")
    userComposite.SetAttribute("user_id", "user-001")
    log.Printf("复合主键: %v, 值: %v", userComposite.PrimaryKeys(), userComposite.GetKey())

    // 手动设置主键
    user5 := models.NewUser()
    user5.SetPrimaryKeys([]string{"tenant_id", "user_code"})
    log.Printf("手动设置复合主键: %v", user5.PrimaryKeys())

    // ===== 高级查询功能演示 =====
    log.Println("===== 高级查询功能演示 =====")

    // 复杂条件查询
    complexQuery, err := db.Table("users", "default")
    if err == nil {
        result, err := complexQuery.
            Select("id", "name", "email").
            Where("age", "BETWEEN", []interface{}{20, 40}).
            WhereIn("status", []interface{}{"active", "pending"}).
            OrderBy("age", "ASC").
            OrderBy("name", "DESC").
            Limit(10).
            Get()
        if err == nil {
            log.Printf("复杂查询结果数量: %d", len(result))
        }
    }

    // 聚合查询
    aggregateQuery, err := db.Table("users", "default")
    if err == nil {
        count, err := aggregateQuery.Where("status", "=", "active").Count()
        if err == nil {
            log.Printf("活跃用户总数: %d", count)
    }
    }

    log.Println("===== 演示完成 =====")
}
```

## 🎯 第3步：运行代码

```bash
go run main.go
```

你将看到类似输出：

```
===== 查询构建器演示 =====
查询到 4 个活跃用户
成年活跃用户数量: 4

===== First和Find新功能演示 =====
First结果: Name=关联测试用户, Age=30
First + 指针填充: 当前=关联测试用户, 指针=关联测试用户
Find + 指针填充: 当前=关联测试用户, 指针=关联测试用户

===== db包First和Find方法演示 =====
db.First() 结果: 关联测试用户
db.First(&model) 结果: Name=关联测试用户

===== 自定义主键功能演示 =====
默认主键: [id]
UUID主键: [uuid], 值: 550e8400-e29b-41d4-a716-446655440000
复合主键: [tenant_id user_id], 值: map[tenant_id:tenant-001 user_id:user-001]
手动设置复合主键: [tenant_id user_code]

===== 高级查询功能演示 =====
复杂查询结果数量: 4
活跃用户总数: 4

===== 演示完成 =====
```

## ✨ 核心特性展示

### 1. 🚫 无Context依赖
```go
// ❌ 旧方式 - 需要传递context
users, err := query.Get(ctx)

// ✅ 新方式 - 简洁的API
users, err := query.Get()

// 需要超时控制时可选使用
users, err := query.WithTimeout(5*time.Second).Get()
```

### 2. 🔗 链式查询构建器
```go
// 直观的链式调用
results, err := db.Table("users").
    Select("name", "email", "age").
    Where("status", "=", "active").
    Where("age", ">=", 18).
    OrderBy("created_at", "desc").
    Limit(10).
    Get()
```

### 3. 🏗️ 现代化迁移系统
```go
// 结构化的迁移定义
migrator.RegisterFunc("20240101_000001", "创建用户表", 
    func(conn db.ConnectionInterface) error {
        // 迁移up逻辑
    }, 
    func(conn db.ConnectionInterface) error {
        // 迁移down逻辑
    })
```

### 4. 📊 智能模型层
```go
// 自动处理CRUD操作
user := NewUser()
user.Name = "新用户"
user.Save() // 自动决定INSERT或UPDATE
```

## 📚 下一步

现在你已经掌握了TORM的基础用法！接下来可以探索：

- [**详细配置**](Configuration.md) - 数据库连接和高级配置
- [**迁移系统**](Migrations.md) - 数据库版本管理
- [**更多示例**](Examples.md) - 复杂查询和实际应用案例
- [**故障排除**](Troubleshooting.md) - 常见问题解决方案

## 💡 关键改进

TORM 现在提供：
- ✅ **可选的超时控制** (`WithTimeout()`, `WithContext()`)
- ✅ **完整的查询构建器**
- ✅ **自动事务管理**
- ✅ **类型安全的模型操作**
- ✅ **现代化的迁移系统**

享受现代化的Go ORM体验！ 🚀

## 📈 API 重构说明

### 从 Context-Based 到 Context-Free

在最新版本中，TORM 进行了重大重构，移除了强制的 `context.Context` 参数，让API更加简洁易用：

#### ❌ 旧版API（需要传递context）
```go
// 旧方式 - 每个调用都需要context
ctx := context.Background()

// 查询操作
users, err := query.Get(ctx)
user, err := query.First(ctx)

// 数据库操作
result, err := conn.Exec(ctx, "INSERT INTO ...", args...)
rows, err := conn.Query(ctx, "SELECT ...", args...)

// 模型操作
err = user.Save(ctx)
err = user.Find(ctx, id)
err = user.Delete(ctx)

// 事务操作
tx, err := conn.Begin(ctx)
```

#### ✅ 新版API（context-free）
```go
// 新方式 - 简洁的API，无需context
// 查询操作
users, err := query.Get()
user, err := query.First()

// 数据库操作  
result, err := conn.Exec("INSERT INTO ...", args...)
rows, err := conn.Query("SELECT ...", args...)

// 模型操作
err = user.Save()
err = user.Find(id)
err = user.Delete()

// 事务操作
err = db.Transaction(func(tx db.TransactionInterface) error {
    // 事务逻辑
    return nil
})
```

#### 🎛️ 可选的Context控制
当需要超时控制或取消操作时，可以使用新增的方法：

```go
// 超时控制
users, err := query.WithTimeout(5*time.Second).Get()

// 自定义context
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
users, err := query.WithContext(ctx).Get()
```

### 🔄 迁移指南

如果你正在从旧版本升级，请按以下步骤迁移：

1. **移除显式的context参数**：
   ```go
   // 旧代码
   users, err := db.Table("users").Get(ctx)
   
   // 新代码  
   users, err := db.Table("users").Get()
   ```

2. **使用新的事务API**：
   ```go
   // 旧代码
   tx, err := conn.Begin(ctx)
   if err != nil {
       return err
   }
   defer tx.Rollback()
   
   _, err = tx.Exec("INSERT INTO ...", args...)
   if err != nil {
       return err
   }
   
   return tx.Commit()
   
   // 新代码
   err := db.Transaction(func(tx db.TransactionInterface) error {
       _, err := tx.Exec("INSERT INTO ...", args...)
       return err // 自动处理commit/rollback
   })
   ```

3. **在需要时添加超时控制**：
   ```go
   // 长时间运行的查询
   results, err := db.Table("large_table").
       WithTimeout(30*time.Second).
       Get()
   ```