# 快速开始指南

欢迎使用TORM v1.1.0！这个指南将在5分钟内让你体验TORM的现代化ORM功能，包括最新的关联预加载、分页器和JSON查询等高级特性。

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

创建 `main.go` 文件：

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "torm/pkg/db"
    "torm/pkg/migration"
    "torm/pkg/model"
)

// User 用户模型
type User struct {
    *model.BaseModel
    ID        interface{} `json:"id" db:"id"`
    Name      string      `json:"name" db:"name"`
    Email     string      `json:"email" db:"email"`
    Age       int         `json:"age" db:"age"`
    Status    string      `json:"status" db:"status"`
    CreatedAt time.Time   `json:"created_at" db:"created_at"`
    UpdatedAt time.Time   `json:"updated_at" db:"updated_at"`
}

// NewUser 创建新用户实例
func NewUser() *User {
    user := &User{
        BaseModel: model.NewBaseModel(),
        Status:    "active",
    }
    user.SetTable("users")
    user.SetPrimaryKey("id")
    user.SetConnection("default")
    return user
}

func main() {
    // 配置数据库连接（使用SQLite，无需额外设置）
    config := &db.Config{
        Driver:   "sqlite",
        Database: "example.db",
    }

    // 添加连接
    err := db.AddConnection("default", config)
    if err != nil {
        log.Fatal("连接数据库失败:", err)
    }

    // 获取连接
    conn, err := db.DB("default")
    if err != nil {
        log.Fatal("获取连接失败:", err)
    }

    // 连接数据库
    err = conn.Connect()
    if err != nil {
        log.Fatal("连接数据库失败:", err)
    }
    defer conn.Close()

    fmt.Println("🎉 数据库连接成功！")

    // 使用迁移系统创建表
    setupDatabase(conn)

    // 演示现代化查询构建器
    demonstrateQueryBuilder()

    // 演示模型操作
    demonstrateModelOperations()

    // 演示事务
    demonstrateTransactions()

    fmt.Println("✅ 所有示例执行完成！")
}

// setupDatabase 使用迁移系统设置数据库
func setupDatabase(conn db.ConnectionInterface) {
    fmt.Println("📊 设置数据库表结构...")
    
    migrator := migration.NewMigrator(conn, nil)
    
    // 注册用户表迁移
    migrator.RegisterFunc("20240101_000001", "创建用户表", func(conn db.ConnectionInterface) error {
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
    }, func(conn db.ConnectionInterface) error {
        _, err := conn.Exec("DROP TABLE IF EXISTS users")
        return err
    })

    // 执行迁移
    err := migrator.Up()
    if err != nil {
        log.Fatal("迁移失败:", err)
    }
    
    fmt.Println("✅ 数据库表创建完成")
}

// demonstrateQueryBuilder 演示查询构建器功能
func demonstrateQueryBuilder() {
    fmt.Println("🔍 演示查询构建器...")

    // 获取查询构建器
    userQuery, err := db.Table("users")
    if err != nil {
        log.Fatal("创建查询失败:", err)
    }

    // 1. 插入数据
    fmt.Println("📝 插入用户数据...")
    userID, err := userQuery.Insert(map[string]interface{}{
        "name":   "张三",
        "email":  "zhangsan@example.com",
        "age":    28,
        "status": "active",
    })
    if err != nil {
        log.Fatal("插入失败:", err)
    }
    fmt.Printf("✅ 插入成功，用户ID: %v\n", userID)

    // 2. 批量插入
    fmt.Println("📝 批量插入用户数据...")
    users := []map[string]interface{}{
        {"name": "李四", "email": "lisi@example.com", "age": 25, "status": "active"},
        {"name": "王五", "email": "wangwu@example.com", "age": 30, "status": "inactive"},
        {"name": "赵六", "email": "zhaoliu@example.com", "age": 35, "status": "active"},
    }
    _, err = userQuery.InsertBatch(users)
    if err != nil {
        log.Fatal("批量插入失败:", err)
    }
    fmt.Println("✅ 批量插入成功")

    // 3. 查询数据
    fmt.Println("🔍 查询活跃用户...")
    activeUsers, err := userQuery.
        Where("status", "=", "active").
        Where("age", ">=", 25).
        OrderBy("age", "desc").
        Get()
    if err != nil {
        log.Fatal("查询失败:", err)
    }
    fmt.Printf("✅ 找到 %d 个活跃用户\n", len(activeUsers))

    // 4. 计数查询
    totalCount, err := userQuery.Count()
    if err != nil {
        log.Fatal("计数失败:", err)
    }
    fmt.Printf("✅ 总用户数: %d\n", totalCount)

    // 5. 更新数据
    affected, err := userQuery.
        Where("email", "=", "wangwu@example.com").
        Update(map[string]interface{}{
            "status": "active",
        })
    if err != nil {
        log.Fatal("更新失败:", err)
    }
    fmt.Printf("✅ 更新了 %d 条记录\n", affected)
}

// demonstrateModelOperations 演示模型操作
func demonstrateModelOperations() {
    fmt.Println("👤 演示模型操作...")

    // 创建新用户
    user := NewUser()
    user.Name = "模型用户"
    user.Email = "model@example.com"
    user.Age = 32

    // 保存用户（会自动决定是插入还是更新）
    err := user.Save()
    if err != nil {
        log.Fatal("保存用户失败:", err)
    }
    fmt.Printf("✅ 用户保存成功，ID: %v\n", user.ID)

    // 根据ID查找用户
    foundUser := NewUser()
    err = foundUser.Find(user.ID)
    if err != nil {
        log.Fatal("查找用户失败:", err)
    }
    fmt.Printf("✅ 找到用户: %s (%s)\n", foundUser.Name, foundUser.Email)

    // 更新用户
    foundUser.Age = 33
    err = foundUser.Save()
    if err != nil {
        log.Fatal("更新用户失败:", err)
    }
    fmt.Println("✅ 用户更新成功")
}

// demonstrateTransactions 演示事务功能
func demonstrateTransactions() {
    fmt.Println("💳 演示事务功能...")

    // 事务成功案例
    err := db.Transaction(func(tx db.TransactionInterface) error {
        // 在事务中插入用户
        result, err := tx.Exec(`
            INSERT INTO users (name, email, age, status) 
            VALUES (?, ?, ?, ?)
        `, "事务用户", "transaction@example.com", 25, "active")
        if err != nil {
            return err
        }

        userID, _ := result.LastInsertId()
        fmt.Printf("✅ 事务中创建用户，ID: %v\n", userID)
        
        return nil
    })
    if err != nil {
        log.Fatal("事务失败:", err)
    }
    fmt.Println("✅ 事务提交成功")
}
```

## 📖 第3步：运行示例

```bash
# 运行示例
go run main.go
```

预期输出：
```
🎉 数据库连接成功！
📊 设置数据库表结构...
✅ 数据库表创建完成
🔍 演示查询构建器...
📝 插入用户数据...
✅ 插入成功，用户ID: 1
📝 批量插入用户数据...
✅ 批量插入成功
🔍 查询活跃用户...
✅ 找到 3 个活跃用户
✅ 总用户数: 4
✅ 更新了 1 条记录
👤 演示模型操作...
✅ 用户保存成功，ID: 5
✅ 找到用户: 模型用户 (model@example.com)
✅ 用户更新成功
💳 演示事务功能...
✅ 事务中创建用户，ID: 6
✅ 事务提交成功
✅ 所有示例执行完成！
```

## 🎯 核心特性亮点

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

### 🎯 重构带来的优势

1. **简洁性**：移除了90%情况下不需要的context参数
2. **一致性**：所有API都遵循相同的调用模式
3. **向后兼容**：通过WithContext()支持需要context的场景
4. **现代化**：符合现代Go ORM的最佳实践
5. **易用性**：降低了学习和使用门槛

### 🚨 注意事项

- 默认情况下，操作使用 `context.Background()`
- MongoDB驱动由于其特性仍然内部使用context，但对外API已简化
- 在高并发场景下，建议使用 `WithTimeout()` 避免无限等待
- 事务会自动处理commit/rollback，无需手动管理

## 🌟 体验v1.1.0新功能

### 关联预加载 (解决N+1查询问题)

```go
// 获取用户数据
users := []interface{}{user1, user2, user3} // 你的用户模型实例

// 预加载关联数据
collection := model.NewModelCollection(users)
collection.With("profile", "posts")
err := collection.Load(context.Background())

// 现在访问关联数据不会产生额外查询
for _, userInterface := range collection.Models() {
    if u, ok := userInterface.(*User); ok {
        profile := u.GetRelation("profile") // 无需查询数据库
        posts := u.GetRelation("posts")     // 无需查询数据库
    }
}
```

### 分页功能

```go
// 简单分页
result, err := userQuery.Paginate(1, 10) // 第1页，每页10条

// 高级分页器
paginator := paginator.NewQueryPaginator(userQuery, ctx)
paginationResult, err := paginator.SetPerPage(15).SetPage(2).Paginate()
```

### JSON字段查询

```go
// 创建高级查询构建器
advQuery := query.NewAdvancedQueryBuilder(baseQuery)

// JSON查询 (支持MySQL、PostgreSQL、SQLite)
users := advQuery.
    WhereJSON("profile", "$.age", ">", 25).
    WhereJSONContains("skills", "$.languages", "Go").
    Get()
```

### 高级查询功能

```go
// 子查询 - 查找有活跃项目的用户
activeUsers := advQuery.WhereExists(func(q db.QueryInterface) db.QueryInterface {
    return q.Where("projects.user_id", "=", "users.id").
        Where("projects.status", "=", "active")
})

// 窗口函数 - 部门内排名
ranking := advQuery.
    WithRowNumber("rank", "department", "salary DESC").
    WithAvgWindow("salary", "dept_avg", "department")
```

---

现在你可以享受更强大、更现代化的TORM v1.1.0体验了！ 🚀

## 🔗 更多资源

- [完整示例](Examples) - 查看详细的功能示例
- [API文档](Configuration) - 深入了解配置选项
- [故障排除](Troubleshooting) - 解决常见问题
- [更新日志](Changelog) - 了解最新变更 