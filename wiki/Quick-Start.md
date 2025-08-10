# 快速开始指南

欢迎使用TORM！这个指南将在5分钟内让你体验TORM的核心功能。我们将创建一个简单的用户管理示例。

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

## 🔧 第2步：创建简单示例

创建 `main.go` 文件：

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "torm/pkg/db"
    "torm/pkg/model"
)

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

    fmt.Println("✅ 数据库连接成功!")

    // 创建表
    createTable()

    // 演示基本操作
    ctx := context.Background()
    demoBasicOperations(ctx)
}

// 创建用户表
func createTable() {
    conn, err := db.DB("default")
    if err != nil {
        log.Fatal("获取连接失败:", err)
    }

    sql := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        email TEXT UNIQUE NOT NULL,
        age INTEGER NOT NULL,
        status TEXT DEFAULT 'active',
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )`
    
    _, err = conn.Exec(context.Background(), sql)
    if err != nil {
        log.Fatal("创建表失败:", err)
    }
    
    fmt.Println("✅ 用户表创建成功")
}

// 演示基本的增删改查操作
func demoBasicOperations(ctx context.Context) {
    // 🔥 创建用户（新增）
    fmt.Println("\n📝 创建新用户...")
    user := model.NewUser()
    user.SetName("张三").SetEmail("zhangsan@example.com").SetAge(25)
    
    err := user.Save(ctx)
    if err != nil {
        log.Printf("创建用户失败: %v", err)
        return
    }
    fmt.Printf("✅ 用户创建成功! ID: %v\n", user.GetID())

    // 🔍 查询用户（查找）
    fmt.Println("\n🔍 查找用户...")
    foundUser := model.NewUser()
    err = foundUser.Find(ctx, user.GetID())
    if err != nil {
        log.Printf("查找用户失败: %v", err)
        return
    }
    fmt.Printf("✅ 找到用户: %s (%s), 年龄: %d\n", 
        foundUser.GetName(), foundUser.GetEmail(), foundUser.GetAge())

    // ✏️ 更新用户（修改）
    fmt.Println("\n✏️ 更新用户信息...")
    foundUser.SetAge(26).SetName("张三丰")
    err = foundUser.Save(ctx)
    if err != nil {
        log.Printf("更新用户失败: %v", err)
        return
    }
    fmt.Printf("✅ 用户更新成功! 新姓名: %s, 新年龄: %d\n", 
        foundUser.GetName(), foundUser.GetAge())

    // 📊 查询多个用户
    fmt.Println("\n📊 查询活跃用户...")
    activeUsers, err := model.FindActiveUsers(ctx, 10)
    if err != nil {
        log.Printf("查询活跃用户失败: %v", err)
        return
    }
    fmt.Printf("✅ 找到 %d 个活跃用户\n", len(activeUsers))
    for _, u := range activeUsers {
        fmt.Printf("  - %s (%s)\n", u.GetName(), u.GetEmail())
    }

    // 🗑️ 删除用户（删除）
    fmt.Println("\n🗑️ 删除用户...")
    err = foundUser.Delete(ctx)
    if err != nil {
        log.Printf("删除用户失败: %v", err)
        return
    }
    fmt.Println("✅ 用户删除成功!")
}
```

## 🏃‍♂️ 运行示例

```bash
# 运行程序
go run main.go
```

预期输出：
```
✅ 数据库连接成功!
✅ 用户表创建成功

📝 创建新用户...
正在创建用户: 张三 (zhangsan@example.com)
用户创建成功: ID=1, 姓名=张三
✅ 用户创建成功! ID: 1

🔍 查找用户...
✅ 找到用户: 张三 (zhangsan@example.com), 年龄: 25

✏️ 更新用户信息...
正在更新用户: ID=1
用户更新成功: ID=1, 姓名=张三丰
✅ 用户更新成功! 新姓名: 张三丰, 新年龄: 26

📊 查询活跃用户...
✅ 找到 1 个活跃用户
  - 张三丰 (zhangsan@example.com)

🗑️ 删除用户...
正在删除用户: ID=1, 姓名=张三丰
用户删除成功: ID=1
✅ 用户删除成功!
```

## 🎯 核心概念

### 1. 数据库连接
```go
config := &db.Config{
    Driver:   "sqlite",        // 数据库类型
    Database: "example.db",    // 数据库文件/名称
}
db.AddConnection("default", config)
```

### 2. 模型操作
```go
// 创建新用户
user := model.NewUser()
user.SetName("张三").SetEmail("zhangsan@example.com").SetAge(25)
user.Save(ctx)  // 保存到数据库

// 查找用户
user.Find(ctx, 1)  // 根据ID查找

// 更新用户
user.SetAge(26)
user.Save(ctx)  // 保存更改

// 删除用户
user.Delete(ctx)
```

### 3. 查询方法
```go
// 根据邮箱查找
user, err := model.FindByEmail(ctx, "zhangsan@example.com")

// 查找活跃用户
users, err := model.FindActiveUsers(ctx, 10)

// 统计用户数量
count, err := model.CountByStatus(ctx, "active")
```

## 🔌 其他数据库配置

### MySQL
```go
config := &db.Config{
    Driver:   "mysql",
    Host:     "localhost",
    Port:     3306,
    Database: "myapp",
    Username: "root",
    Password: "password",
}
```

### PostgreSQL
```go
config := &db.Config{
    Driver:   "postgres",
    Host:     "localhost",
    Port:     5432,
    Database: "myapp",
    Username: "postgres",
    Password: "password",
    SSLMode:  "disable",
}
```

## 🎉 恭喜！

你已经成功运行了第一个TORM应用！现在你可以：

### 📚 继续学习
- [配置文档](Configuration) - 了解详细配置选项
- [查询构建器](Query-Builder) - 学习复杂查询
- [数据迁移](Migrations) - 管理数据库结构
- [关联关系](Relationships) - 处理表之间的关系

### 🛠️ 实际应用
- [完整示例](Examples) - 查看真实项目示例
- [最佳实践](Best-Practices) - 学习推荐用法
- [API参考](API-Reference) - 查看所有可用方法

### ❓ 需要帮助？
- [故障排除](Troubleshooting) - 解决常见问题
- [GitHub Issues](https://github.com/zhoudm1743/torm/issues) - 报告问题
- 邮件联系: zhoudm1743@163.com

---

**🚀 开始构建你的应用吧！** 