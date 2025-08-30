# 快速开始指南

欢迎使用TORM,这个指南将在5分钟内让你体验TORM的强大功能，包括零配置的自动迁移、丰富的TORM标签系统、跨数据库支持等革命性特性。

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

# 安装数据库驱动（根据需要选择）
go get github.com/go-sql-driver/mysql      # MySQL
go get github.com/lib/pq                   # PostgreSQL  
go get github.com/mattn/go-sqlite3         # SQLite
```

## 🔧 第2步：创建现代化示例

### 创建主文件 `main.go`

```go
package main

import (
    "fmt"
    "time"
    "github.com/zhoudm1743/torm"
)

// 用户模型 - 展示丰富的TORM标签
type User struct {
    torm.BaseModel
    
    // 主键和自增
    ID int `json:"id" torm:"primary_key,auto_increment"`
    
    // 字符串类型和约束
    Username string `json:"username" torm:"type:varchar,size:50,unique,index"`
    Email    string `json:"email" torm:"type:varchar,size:100,unique"`
    Password string `json:"password" torm:"type:varchar,size:255"`
    
    // 数值类型和默认值
    Age    int     `json:"age" torm:"type:int,unsigned,default:0"`
    Salary float64 `json:"salary" torm:"type:decimal,precision:10,scale:2,default:0.00"`
    
    // 状态和布尔字段
    Status   string `json:"status" torm:"type:varchar,size:20,default:active,index"`
    IsActive bool   `json:"is_active" torm:"type:boolean,default:1"`
    
    // 文本字段
    Bio string `json:"bio" torm:"type:text"`
    
    // 外键关联
    DeptID int `json:"dept_id" torm:"type:int,references:departments.id,on_delete:set_null"`
    
    // 自动时间戳
    CreatedAt time.Time `json:"created_at" torm:"auto_create_time"`
    UpdatedAt time.Time `json:"updated_at" torm:"auto_update_time"`
}

// 部门模型
type Department struct {
    torm.BaseModel
    ID       int     `json:"id" torm:"primary_key,auto_increment"`
    Name     string  `json:"name" torm:"type:varchar,size:100,unique"`
    Budget   float64 `json:"budget" torm:"type:decimal,precision:12,scale:2,default:0.00"`
    Location string  `json:"location" torm:"type:varchar,size:255"`
    IsActive bool    `json:"is_active" torm:"type:boolean,default:1"`
    
    CreatedAt time.Time `json:"created_at" torm:"auto_create_time"`
    UpdatedAt time.Time `json:"updated_at" torm:"auto_update_time"`
}

func main() {
    fmt.Println("🚀 TORM 快速开始演示")
    
    // 第1步：配置数据库连接
    fmt.Println("\n📡 配置数据库连接...")
    
    // SQLite（推荐用于快速开始）
    err := torm.AddConnection("default", &torm.Config{
        Driver:   "sqlite",
        Database: "quickstart.db",
    })
    
    // MySQL示例（可选）
    /*
    err := torm.AddConnection("default", &torm.Config{
        Driver:   "mysql",
        Host:     "localhost",
        Port:     3306,
        Username: "root",
        Password: "password",
        Database: "torm_demo",
        Charset:  "utf8mb4",
    })
    */
    
    // PostgreSQL示例（可选）
    /*
    err := torm.AddConnection("default", &torm.Config{
        Driver:   "postgres",
        Host:     "localhost",
        Port:     5432,
        Username: "postgres",
        Password: "password",
        Database: "torm_demo",
        SSLMode:  "disable",
    })
    */
    
    if err != nil {
        panic(fmt.Sprintf("数据库连接失败: %v", err))
    }
    fmt.Println("✅ 数据库连接成功")
    
    // 第2步：自动创建表结构
    fmt.Println("\n🏗️  自动创建表结构...")
    
    // 先创建部门表（被引用的表）
    dept := &Department{}
    if err := dept.AutoMigrate(); err != nil {
        panic(fmt.Sprintf("部门表创建失败: %v", err))
    }
    fmt.Println("✅ 部门表创建成功")
    
    // 再创建用户表（包含外键）
    user := &User{}
    if err := user.AutoMigrate(); err != nil {
        panic(fmt.Sprintf("用户表创建失败: %v", err))
    }
    fmt.Println("✅ 用户表创建成功")
    
    // 第3步：演示CRUD操作
    fmt.Println("\n📊 演示CRUD操作...")
    
    // 创建部门
    techDept := &Department{
        Name:     "技术部",
        Budget:   100000.00,
        Location: "北京",
        IsActive: true,
    }
    if err := techDept.Save(); err != nil {
        fmt.Printf("部门创建失败: %v\n", err)
    } else {
        fmt.Printf("✅ 部门创建成功，ID: %d\n", techDept.ID)
    }
    
    // 创建用户
    newUser := &User{
        Username: "zhangsan",
        Email:    "zhangsan@example.com",
        Password: "password123",
        Age:      28,
        Salary:   8000.50,
        Status:   "active",
        IsActive: true,
        Bio:      "这是一个演示用户",
        DeptID:   techDept.ID,
    }
    
    if err := newUser.Save(); err != nil {
        fmt.Printf("用户创建失败: %v\n", err)
    } else {
        fmt.Printf("✅ 用户创建成功，ID: %d\n", newUser.ID)
    }
    
    // 第4步：演示查询操作
    fmt.Println("\n🔍 演示查询操作...")
    
    // 原始数据查询（高性能）
    users, err := torm.Table("users").
        Where("status", "=", "active").
        Where("age", ">=", 18).
        OrderBy("created_at", "desc").
        GetRaw()
    
    if err != nil {
        fmt.Printf("查询失败: %v\n", err)
    } else {
        fmt.Printf("✅ 查询到 %d 个活跃用户\n", len(users))
    }
    
    // 参数化查询
    activeUsers, err := torm.Table("users").
        Where("status = ? AND age >= ?", "active", 18).
        GetRaw()
    
    if err != nil {
        fmt.Printf("参数化查询失败: %v\n", err)
    } else {
        fmt.Printf("✅ 参数化查询到 %d 个用户\n", len(activeUsers))
    }
    
    // 聚合查询
    count, err := torm.Table("users").
        Where("status", "=", "active").
        Count()
    
    if err != nil {
        fmt.Printf("计数查询失败: %v\n", err)
    } else {
        fmt.Printf("✅ 活跃用户总数: %d\n", count)
    }
    
    // 第5步：演示更新操作
    fmt.Println("\n🔄 演示更新操作...")
    
    affected, err := torm.Table("users").
        Where("username", "=", "zhangsan").
        Update(map[string]interface{}{
            "salary": 9000.00,
            "status": "promoted",
        })
    
    if err != nil {
        fmt.Printf("更新失败: %v\n", err)
    } else {
        fmt.Printf("✅ 更新成功，影响行数: %d\n", affected)
    }
    
    // 第6步：演示事务操作
    fmt.Println("\n💼 演示事务操作...")
    
    err = torm.Transaction(func(tx torm.TransactionInterface) error {
        // 在事务中执行多个操作
        _, err := tx.Exec("INSERT INTO departments (name, budget, location, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
            "事务部门", 50000.00, "上海", true, time.Now(), time.Now())
        if err != nil {
            return err
        }
        
        _, err = tx.Exec("INSERT INTO users (username, email, password, age, status, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
            "transaction_user", "tx@example.com", "password", 25, "active", true, time.Now(), time.Now())
        if err != nil {
            return err
        }
        
        return nil // 自动提交
    })
    
    if err != nil {
        fmt.Printf("事务失败: %v\n", err)
    } else {
        fmt.Println("✅ 事务执行成功")
    }
    
    // 第7步：演示访问器功能（新特性）
    fmt.Println("\n🎨 演示访问器功能...")
    
    // 如果有用户数据，演示访问器
    if count > 0 {
        fmt.Println("访问器演示：")
        
        // 使用模型查询（支持访问器）
        userResult, err := torm.Table("users").
            Model(&User{}).                    // 启用访问器
            Where("username", "=", "zhangsan").
            First()
        
        if err == nil && userResult != nil {
            fmt.Printf("✅ 访问器查询成功\n")
            fmt.Printf("   用户名: %v\n", userResult["username"])
            fmt.Printf("   状态: %v\n", userResult["status"])
            
            // 演示JSON输出
            if jsonBytes, err := json.Marshal(userResult); err == nil {
                fmt.Printf("   JSON: %s\n", string(jsonBytes)[:100] + "...")  // 显示前100字符
            }
        }
        
        // 演示集合查询
        userResults, err := torm.Table("users").
            Model(&User{}).
            Where("status", "=", "promoted").
            Get()
        
        if err == nil && len(userResults) > 0 {
            fmt.Printf("✅ 找到 %d 个晋升用户\n", len(userResults))
            
            // 演示遍历
            for i, user := range userResults {
                if i < 3 {  // 只显示前3个
                    fmt.Printf("   [%d] 用户: %v\n", i+1, user["username"])
                }
            }
        }
    }
    
    fmt.Println("\n🎉 TORM 快速开始演示完成！")
    fmt.Println("\n✨ 新特性亮点：")
    fmt.Println("   🎨 访问器系统 - 支持Get/Set访问器")
    fmt.Println("   🔗 Model().Get() - 简洁的链式调用API")
    fmt.Println("   ⚡ GetRaw() - 高性能原始数据查询")
    fmt.Println("   📊 直接数据操作 - 原生map[string]interface{}")
    fmt.Println("\n📚 接下来你可以：")
    fmt.Println("   - 查看完整文档：http://torm.site/docs.html")
    fmt.Println("   - 学习TORM标签：http://torm.site/docs.html?doc=migrations")
    fmt.Println("   - 探索查询构建器：http://torm.site/docs.html?doc=query-builder")
    fmt.Println("   - 了解模型系统：http://torm.site/docs.html?doc=model-system")
    fmt.Println("   - 体验访问器系统：http://torm.site/docs.html?doc=model-system#accessors")
}
```

## 🎯 第3步：运行代码

```bash
go run main.go
```

你将看到类似输出：

```
🚀 TORM 快速开始演示

📡 配置数据库连接...
✅ 数据库连接成功

🏗️ 自动创建表结构...
✅ 部门表创建成功
✅ 用户表创建成功

📊 演示CRUD操作...
✅ 部门创建成功，ID: 1
✅ 用户创建成功，ID: 1

🔍 演示查询操作...
✅ 查询到 1 个活跃用户
✅ 参数化查询到 1 个用户
✅ 活跃用户总数: 1

🔄 演示更新操作...
✅ 更新成功，影响行数: 1

💼 演示事务操作...
✅ 事务执行成功

🎉 TORM 快速开始演示完成！
```

## ✨ 核心特性展示

### 1. 🆕 零配置自动迁移

```go
// 定义模型即完成数据库设计
type Product struct {
    torm.BaseModel
    ID          int     `torm:"primary_key,auto_increment"`
    Name        string  `torm:"type:varchar,size:200,comment:产品名称"`
    SKU         string  `torm:"type:varchar,size:50,unique,comment:产品编码"`
    Price       float64 `torm:"type:decimal,precision:10,scale:2,default:0.00"`
    CategoryID  int     `torm:"type:int,references:categories.id,on_delete:cascade"`
    IsActive    bool    `torm:"type:boolean,default:1"`
    CreatedAt   time.Time `torm:"auto_create_time"`
}

// 一行代码创建完整表结构（包括索引、外键、约束）
product := &Product{}
product.AutoMigrate()
```

### 2. 🏷️ 丰富的TORM标签系统

```go
type User struct {
    torm.BaseModel
    
    // 主键和自增
    ID int64 `torm:"primary_key,auto_increment,comment:用户ID"`
    
    // 精确类型控制
    Username string  `torm:"type:varchar,size:50,unique,index"`
    Email    string  `torm:"type:varchar,size:100,unique"`
    Bio      string  `torm:"type:text"`
    
    // 数值精度控制
    Age      int     `torm:"type:int,unsigned,default:0"`
    Salary   float64 `torm:"type:decimal,precision:10,scale:2"`
    
    // 索引优化
    Status   string  `torm:"type:varchar,size:20,default:active,index"`
    City     string  `torm:"type:varchar,size:50,index:city_idx"`
    
    // 外键关系
    DeptID   int     `torm:"type:int,references:departments.id,on_delete:set_null"`
    
    // 自动时间戳
    CreatedAt time.Time `torm:"auto_create_time"`
    UpdatedAt time.Time `torm:"auto_update_time"`
}
```

### 3. 🔗 强大的查询构建器

```go
// 访问器查询（支持访问器）
users, _ := torm.Table("users").
    Model(&User{}).                     // 启用访问器
    Where("status", "=", "active").
    Where("age", ">=", 18).
    OrderBy("created_at", "desc").
    Limit(10).
    Get()                               // 返回 []map[string]interface{}

// 参数化查询（支持数组参数）
activeUsers, _ := torm.Table("users").
    Model(&User{}).
    Where("status IN (?)", []string{"active", "premium"}).
    Where("age BETWEEN ? AND ?", 18, 65).
    Get()

// 复杂条件组合
results, _ := torm.Table("users").
    Model(&User{}).
    Where("(status = ? OR vip_level > ?) AND age >= ?", "premium", 3, 25).
    Get()

// 原始数据查询（高性能）
rawUsers, _ := torm.Table("users").
    Where("status", "=", "active").
    GetRaw()                            // 返回 []map[string]interface{}

// 聚合查询
count, _ := torm.Table("users").Where("status", "=", "active").Count()
```

### 4. 🌐 跨数据库支持

```go
// 同一套代码，支持多种数据库
configs := map[string]*torm.Config{
    "sqlite": {
        Driver:   "sqlite",
        Database: "app.db",
    },
    "mysql": {
        Driver:   "mysql",
        Host:     "localhost",
        Database: "myapp",
        Username: "root",
        Password: "password",
    },
    "postgres": {
        Driver:   "postgres", 
        Host:     "localhost",
        Database: "myapp",
        Username: "postgres",
        Password: "password",
        SSLMode:  "disable",
    },
}

// 同一模型自动适配不同数据库
for name, config := range configs {
    torm.AddConnection(name, config)
    
    user := &User{}
    user.SetConnection(name)
    user.AutoMigrate() // 自动适配数据库差异
}
```

### 5. 💼 自动事务管理

```go
// 简洁的事务API
err := torm.Transaction(func(tx torm.TransactionInterface) error {
    // 所有操作在事务中执行
    _, err := tx.Exec("INSERT INTO users (...) VALUES (...)")
    if err != nil {
        return err // 自动回滚
    }
    
    _, err = tx.Exec("UPDATE departments SET budget = budget - 1000")
    if err != nil {
        return err // 自动回滚
    }
    
    return nil // 自动提交
})
```

## 📊 性能特点

- **🚀 零反射查询**: 直接SQL构建，避免运行时反射开销
- **🔄 智能占位符**: 自动适配MySQL(`?`)和PostgreSQL(`$N`)占位符
- **📦 批量操作**: 原生支持批量插入和数组参数
- **🏗️ 连接池优化**: 高效的数据库连接管理
- **📈 索引自动化**: 根据TORM标签自动创建优化索引
- **🎨 双模式查询**: 访问器查询(功能丰富) + Raw查询(高性能)

### API 性能对比

```go
// 🎯 功能丰富：Model().Get() (获得访问器支持)
users, _ := torm.Table("users").Model(&User{}).Get()
for _, user := range users {
    // 自动格式化的数据，适合前端展示
    status := user["status"]  // 访问器处理后的数据
}

// ⚡ 高性能：GetRaw() (最佳性能，原始数据)
rawUsers, _ := torm.Table("users").GetRaw()
for _, user := range rawUsers {
    // 直接操作原始map，性能最优
    status := user["status"]  // 原始数据库值
}

// 🔄 混合使用：根据场景选择
users, _ := torm.Table("users").Model(&User{}).Get()
totalSalary := 0
for _, user := range users {
    // 显示数据用访问器
    displayInfo := user["salary"] // 访问器格式化后的数据
    
    // 计算逻辑需要原始值时使用GetRaw()
    if rawSalary, ok := user["salary"].(float64); ok {
        totalSalary += int(rawSalary)
    }
}
```

## 🛠️ 常见使用场景

### 1. 快速原型开发

```go
// 30秒搭建博客数据模型
type Post struct {
    torm.BaseModel
    ID       int    `torm:"primary_key,auto_increment"`
    Title    string `torm:"type:varchar,size:200"`
    Content  string `torm:"type:text"`
    AuthorID int    `torm:"type:int,references:users.id"`
    Status   string `torm:"type:varchar,size:20,default:draft,index"`
    CreatedAt time.Time `torm:"auto_create_time"`
}

(&User{}).AutoMigrate()
(&Post{}).AutoMigrate()
```

### 2. 微服务架构

```go
// 每个服务独立的数据模型
type OrderService struct{}

func (s *OrderService) InitDatabase() {
    torm.AddConnection("orders", config)
    
    models := []interface{}{
        &Order{}, &OrderItem{}, &Payment{},
    }
    
    for _, model := range models {
        model.(interface{ AutoMigrate() error }).AutoMigrate()
    }
}
```

### 3. 多环境部署

```go
func deployEnvironment(env string) {
    config := getConfigByEnv(env) // dev/test/prod配置
    torm.AddConnection("default", config)
    
    // 同一套模型，自动适配不同环境的数据库
    (&User{}).AutoMigrate()
    (&Product{}).AutoMigrate()
}
```