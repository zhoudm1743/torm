# TORM 现代化ORM使用指南

## 🤔 为什么要使用ORM而不是原生SQL？

您提出了很好的问题："为什么测试文件还在使用原生SQL而不是封装好的查询构建函数？"

答案是：**您说得对！应该优先使用TORM的现代化ORM功能。**

## ✅ 推荐的现代化使用方式

### 1. 📊 数据库操作 - 使用查询构建器

```go
// ❌ 不推荐：直接使用原生SQL
conn.Exec("INSERT INTO users (name, email) VALUES (?, ?)", "张三", "test@example.com")
conn.Query("SELECT * FROM users WHERE status = ?", "active")

// ✅ 推荐：使用查询构建器
userQuery, _ := db.Table("users")

// 插入数据
userID, err := userQuery.Insert(map[string]interface{}{
    "name":   "张三",
    "email":  "test@example.com",
    "status": "active",
})

// 查询数据
users, err := userQuery.
    Where("status", "=", "active").
    OrderBy("created_at", "desc").
    Get()

// 更新数据
affected, err := userQuery.
    Where("id", "=", userID).
    Update(map[string]interface{}{
        "status": "verified",
    })

// 删除数据
deleted, err := userQuery.
    Where("status", "=", "inactive").
    Delete()
```

### 2. 🏗️ 数据库结构管理 - 使用迁移系统

```go
// ❌ 不推荐：直接执行DDL语句
conn.Exec("CREATE TABLE users (...)")
conn.Exec("DROP TABLE users")

// ✅ 推荐：使用迁移系统
migrator := migration.NewMigrator(conn, nil)

migrator.RegisterFunc("20240101_000001", "创建用户表", 
    func(conn db.ConnectionInterface) error {
        _, err := conn.Exec(`
            CREATE TABLE users (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT NOT NULL,
                email TEXT UNIQUE NOT NULL,
                status TEXT DEFAULT 'active',
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP
            )
        `)
        return err
    },
    func(conn db.ConnectionInterface) error {
        _, err := conn.Exec("DROP TABLE IF EXISTS users")
        return err
    })

// 执行迁移
err := migrator.Up()
```

### 3. 💳 事务处理 - 使用现代化事务API

```go
// ❌ 不推荐：手动管理事务
tx, err := conn.Begin()
if err != nil {
    return err
}
defer tx.Rollback()

_, err = tx.Exec("INSERT INTO users ...")
if err != nil {
    return err
}

_, err = tx.Exec("INSERT INTO profiles ...")
if err != nil {
    return err
}

return tx.Commit()

// ✅ 推荐：使用自动管理的事务
err := db.Transaction(func(tx db.TransactionInterface) error {
    // 创建用户
    result, err := tx.Exec("INSERT INTO users (name, email) VALUES (?, ?)", 
        "张三", "zhangsan@example.com")
    if err != nil {
        return err
    }
    
    userID, _ := result.LastInsertId()
    
    // 创建用户档案
    _, err = tx.Exec("INSERT INTO profiles (user_id, bio) VALUES (?, ?)", 
        userID, "新用户")
    return err
    // 自动commit，出错则自动rollback
})
```

### 4. 🔄 模型操作 - 使用ORM模型

```go
// ❌ 不推荐：手动处理CRUD
conn.Exec("INSERT INTO users (name, email) VALUES (?, ?)", user.Name, user.Email)
rows, err := conn.Query("SELECT * FROM users WHERE id = ?", id)
// 手动扫描结果...

// ✅ 推荐：使用ORM模型
user := NewUser()
user.Name = "张三"
user.Email = "zhangsan@example.com"

// 保存（自动决定INSERT或UPDATE）
err := user.Save()

// 查找
foundUser := NewUser()
err = foundUser.Find(userID)

// 更新
foundUser.Name = "李四"
err = foundUser.Save()

// 删除
err = foundUser.Delete()
```

## 🎯 完整的现代化示例

以下是一个完整的现代化使用示例：

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/zhoudm1743/torm/pkg/db"
    "github.com/zhoudm1743/torm/pkg/migration"
)

func main() {
    // 1. 配置数据库
    config := &db.Config{
        Driver:   "sqlite",
        Database: "modern_app.db",
    }
    
    err := db.AddConnection("default", config)
    if err != nil {
        log.Fatal(err)
    }
    
    conn, _ := db.DB("default")
    conn.Connect()
    defer conn.Close()
    
    // 2. 使用迁移系统建表
    setupDatabase(conn)
    
    // 3. 使用查询构建器操作数据
    demonstrateQueryBuilder()
    
    // 4. 使用事务处理复杂操作
    demonstrateTransactions()
    
    fmt.Println("✅ 现代化ORM演示完成！")
}

func setupDatabase(conn db.ConnectionInterface) {
    migrator := migration.NewMigrator(conn, nil)
    
    // 用户表迁移
    migrator.RegisterFunc("20240101_000001", "创建用户表", 
        func(conn db.ConnectionInterface) error {
            _, err := conn.Exec(`
                CREATE TABLE IF NOT EXISTS users (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    name TEXT NOT NULL,
                    email TEXT UNIQUE NOT NULL,
                    age INTEGER DEFAULT 0,
                    status TEXT DEFAULT 'active',
                    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
                )
            `)
            return err
        },
        func(conn db.ConnectionInterface) error {
            _, err := conn.Exec("DROP TABLE IF EXISTS users")
            return err
        })
    
    migrator.Up()
}

func demonstrateQueryBuilder() {
    fmt.Println("🔧 演示查询构建器...")
    
    query, _ := db.Table("users")
    
    // 批量插入
    users := []map[string]interface{}{
        {"name": "张三", "email": "zhangsan@example.com", "age": 28},
        {"name": "李四", "email": "lisi@example.com", "age": 32},
        {"name": "王五", "email": "wangwu@example.com", "age": 25},
    }
    
    affected, _ := query.InsertBatch(users)
    fmt.Printf("✅ 批量插入 %d 个用户\n", affected)
    
    // 复杂查询
    activeUsers, _ := query.
        Where("status", "=", "active").
        Where("age", ">=", 25).
        OrderBy("age", "desc").
        Limit(10).
        Get()
    
    fmt.Printf("✅ 查询到 %d 个活跃用户\n", len(activeUsers))
    
    // 聚合查询
    count, _ := query.
        Where("status", "=", "active").
        Count()
    
    fmt.Printf("✅ 活跃用户总数: %d\n", count)
}

func demonstrateTransactions() {
    fmt.Println("💳 演示事务处理...")
    
    err := db.Transaction(func(tx db.TransactionInterface) error {
        // 在事务中执行多个操作
        result, err := tx.Exec(`
            INSERT INTO users (name, email, age, status) 
            VALUES (?, ?, ?, ?)
        `, "事务用户", "transaction@example.com", 30, "active")
        
        if err != nil {
            return err
        }
        
        userID, _ := result.LastInsertId()
        fmt.Printf("✅ 事务中创建用户，ID: %d\n", userID)
        
        return nil
    })
    
    if err != nil {
        fmt.Printf("❌ 事务失败: %v\n", err)
    } else {
        fmt.Println("✅ 事务执行成功！")
    }
}
```

## 🚀 关键优势

### 1. 类型安全
```go
// 查询构建器提供类型安全的操作
query.Where("age", ">=", 18)  // 自动处理类型转换
query.WhereIn("status", []interface{}{"active", "verified"})
```

### 2. SQL注入防护
```go
// 自动参数化查询，防止SQL注入
query.Where("email", "=", userInput)  // 安全的
```

### 3. 数据库无关性
```go
// 同样的代码可以在不同数据库上运行
query.Insert(data)  // 在MySQL、PostgreSQL、SQLite上都能工作
```

### 4. 链式调用
```go
// 直观的链式API
results := query.
    Select("name", "email").
    Where("status", "=", "active").
    OrderBy("created_at", "desc").
    Limit(10).
    Get()
```

### 5. 自动事务管理
```go
// 无需手动管理commit/rollback
db.Transaction(func(tx db.TransactionInterface) error {
    // 业务逻辑
    return nil  // 自动commit
    // return err  // 自动rollback
})
```

## 📝 最佳实践建议

1. **优先使用查询构建器**：替代原生SQL查询
2. **使用迁移系统**：管理数据库结构变更
3. **利用事务API**：确保数据一致性
4. **采用ORM模型**：简化CRUD操作
5. **添加适当的超时**：使用`WithTimeout()`防止长时间等待

## 🎉 总结

TORM现在提供了完整的现代化ORM体验：

- ✅ **无Context的简洁API**
- ✅ **强大的查询构建器**
- ✅ **自动事务管理**
- ✅ **完整的迁移系统**
- ✅ **类型安全的操作**

您完全可以告别原生SQL，拥抱现代化的ORM开发方式！ 