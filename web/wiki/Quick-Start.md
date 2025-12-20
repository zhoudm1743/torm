# 快速开始

本指南将帮助你在 5 分钟内上手 TORM。

## 📋 前置要求

- Go 1.19 或更高版本
- 支持的数据库 (MySQL, PostgreSQL, SQLite)

## 🚀 安装

```bash
# 创建项目
mkdir my-project && cd my-project
go mod init my-project

# 安装 TORM
go get github.com/zhoudm1743/torm

# 安装数据库驱动（根据需要选择）
go get github.com/go-sql-driver/mysql      # MySQL
go get github.com/lib/pq                   # PostgreSQL
go get github.com/mattn/go-sqlite3         # SQLite
```

## 🔧 配置数据库连接

### MySQL

```go
import "github.com/zhoudm1743/torm"

err := torm.AddConnection("default", &torm.Config{
    Driver:   "mysql",
    Host:     "localhost",
    Port:     3306,
    Database: "myapp",
    Username: "root",
    Password: "password",
    Charset:  "utf8mb4",
})
```

### PostgreSQL

```go
err := torm.AddConnection("default", &torm.Config{
    Driver:   "postgres",
    Host:     "localhost",
    Port:     5432,
    Database: "myapp",
    Username: "postgres",
    Password: "password",
    SSLMode:  "disable",
})
```

### SQLite

```go
err := torm.AddConnection("default", &torm.Config{
    Driver:   "sqlite",
    Database: "app.db",
})
```

## 📝 基础查询

### 查询数据

```go
// 查询所有记录
users, err := torm.Table("users").Get()

// 条件查询
users, err := torm.Table("users").
    Where("status", "=", "active").
    Where("age", ">=", 18).
    Get()

// 参数化查询
users, err := torm.Table("users").
    Where("status = ? AND age >= ?", "active", 18).
    Get()

// 查询单条记录
user, err := torm.Table("users").
    Where("id", "=", 1).
    First()

// 检查记录是否存在
exists, err := torm.Table("users").
    Where("email", "=", "test@example.com").
    Exists()
```

### 插入数据

```go
// 单条插入
id, err := torm.Table("users").Insert(map[string]interface{}{
    "name":   "张三",
    "email":  "zhangsan@example.com",
    "status": "active",
})

// 批量插入
count, err := torm.Table("users").InsertBatch([]map[string]interface{}{
    {"name": "李四", "email": "lisi@example.com"},
    {"name": "王五", "email": "wangwu@example.com"},
})
```

### 更新数据

```go
// 条件更新
affected, err := torm.Table("users").
    Where("id", "=", 1).
    Update(map[string]interface{}{
        "name":   "新名字",
        "status": "updated",
    })

// 自增/自减
affected, err := torm.Table("users").
    Where("id", "=", 1).
    Increment("login_count", 1)
```

### 删除数据

```go
affected, err := torm.Table("users").
    Where("status", "=", "deleted").
    Delete()
```

## 📦 使用模型

### 定义模型

```go
import (
    "time"
    "github.com/zhoudm1743/torm/model"
)

type User struct {
    model.BaseModel
    ID        int       `json:"id" torm:"primary_key,auto_increment"`
    Name      string    `json:"name" torm:"type:varchar,size:100"`
    Email     string    `json:"email" torm:"type:varchar,size:100,unique"`
    Status    string    `json:"status" torm:"type:varchar,size:20,default:active"`
    CreatedAt time.Time `json:"created_at" torm:"auto_create_time"`
    UpdatedAt time.Time `json:"updated_at" torm:"auto_update_time"`
}
```

### 使用模型

```go
// 创建模型实例
userModel := model.NewModel(&User{})
userModel.SetConnection("default")

// 自动迁移（创建表）
err := userModel.AutoMigrate(&User{})

// 创建记录
user := model.NewModel("users")
user.SetPrimaryKey("id").SetConnection("default")
user.Fill(map[string]interface{}{
    "name":   "张三",
    "email":  "zhangsan@example.com",
    "status": "active",
})
err := user.Save()

// 查询记录
user := model.NewModel("users")
user.SetConnection("default")
err := user.Find(1) // 根据主键查找

// 获取属性
name := user.GetAttribute("name")
email := user.GetAttribute("email")

// 更新记录
user.SetAttribute("name", "新名字")
err := user.Save()

// 删除记录
err := user.Delete()
```

## 💼 事务处理

```go
err := torm.Transaction(func(tx torm.TransactionInterface) error {
    // 在事务中执行操作
    _, err := tx.Exec("INSERT INTO users (name, email) VALUES (?, ?)", 
        "张三", "zhangsan@example.com")
    if err != nil {
        return err // 自动回滚
    }
    
    _, err = tx.Exec("UPDATE accounts SET balance = balance - 100 WHERE user_id = ?", 1)
    if err != nil {
        return err // 自动回滚
    }
    
    return nil // 自动提交
})
```

## 📊 聚合查询

```go
// 计数
count, err := torm.Table("users").
    Where("status", "=", "active").
    Count()

// 求和
total, err := torm.Table("orders").Sum("amount")

// 平均值
avg, err := torm.Table("products").Avg("price")

// 最大/最小值
max, err := torm.Table("products").Max("price")
min, err := torm.Table("products").Min("price")
```

## 📄 分页查询

```go
// 分页查询
result, err := torm.Table("users").
    Where("status", "=", "active").
    OrderBy("created_at", "desc").
    Paginate(1, 20) // 第1页，每页20条

// 访问分页结果
fmt.Printf("总数: %d\n", result.Total)
fmt.Printf("当前页: %d\n", result.CurrentPage)
fmt.Printf("总页数: %d\n", result.LastPage)
for _, user := range result.Data {
    fmt.Printf("用户: %v\n", user["name"])
}
```

## 🔗 关联查询

```go
// 定义关联
func (u *User) Posts() *model.HasMany {
    return u.HasMany(&Post{}, "user_id", "id")
}

// 使用关联
user := model.NewModel(&User{})
user.SetConnection("default")
user.Find(1)

// 获取用户的所有文章
posts, err := user.Posts().Get()

// 带条件的关联查询
publishedPosts, err := user.Posts().
    Where("status", "=", "published").
    OrderBy("created_at", "desc").
    Get()
```

## 🗄️ 自动迁移

```go
// 定义模型
type User struct {
    model.BaseModel
    ID        int       `torm:"primary_key,auto_increment"`
    Name      string    `torm:"type:varchar,size:100"`
    Email     string    `torm:"type:varchar,size:100,unique"`
    CreatedAt time.Time `torm:"auto_create_time"`
}

// 自动创建/更新表结构
userModel := model.NewModel(&User{})
userModel.SetConnection("default")
err := userModel.AutoMigrate(&User{})
```

## 📚 下一步

- [配置文档](Configuration) - 详细配置选项
- [查询构建器](Query-Builder) - 高级查询技巧
- [模型系统](Model-System) - 深入模型设计
- [关联关系](Relationships) - 模型间关联
- [事务处理](Transactions) - 事务管理
