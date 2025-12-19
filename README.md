# TORM - Go高性能ORM框架

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.19-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-Apache2.0-green.svg)](LICENSE)
[![Documentation](https://img.shields.io/badge/docs-torm.site-brightgreen.svg)](http://torm.site)
[![Version](https://img.shields.io/badge/version-1.2.12-orange.svg)](https://github.com/zhoudm1743/torm/releases)

TORM 是一个参考 ThinkORM 4.0 架构设计的高性能 Go ORM 框架，对 database/sql 进行深度封装，提供流畅的链式查询、完整的模型系统、关联关系、事务支持、查询缓存和数据库迁移功能。

如果您觉得 TORM 有帮助，请帮忙给个 star ⭐

## 🌐 官方网站

**官网地址**: [torm.site](http://torm.site)

- 📚 [完整文档](http://torm.site/docs.html)
- 🚀 [快速开始](http://torm.site/docs.html?doc=quick-start)
- 💡 [示例代码](http://torm.site/docs.html?doc=examples)
- ⚙️ [配置指南](http://torm.site/docs.html?doc=configuration)

## ✨ 主要特性

- **多数据库支持**: MySQL、PostgreSQL、SQLite 完整支持
- **链式查询构建器**: 流畅的 API，支持参数化查询和数组参数
- **完整模型系统**: 时间戳、软删除、访问器/修改器、JSON 字段
- **关联关系**: HasOne、HasMany、BelongsTo、BelongsToMany 及预加载
- **事务支持**: 自动事务、嵌套事务、保存点
- **查询缓存**: 内存缓存、Redis 适配器、缓存标签
- **自动迁移**: 根据模型结构自动创建和更新表
- **读写分离**: 主从配置、负载均衡
- **连接池管理**: 高效的连接复用和自动回收
- **跨数据库兼容**: 自动适配不同数据库的 SQL 语法和占位符

## 🚀 快速开始

### 安装

```bash
go get github.com/zhoudm1743/torm

# 安装数据库驱动（根据需要选择）
go get github.com/go-sql-driver/mysql    # MySQL
go get github.com/lib/pq                 # PostgreSQL
go get github.com/mattn/go-sqlite3       # SQLite
```

### 基础用法

```go
package main

import (
    "fmt"
    "time"
    "github.com/zhoudm1743/torm"
)

// 定义模型
type User struct {
    torm.BaseModel
    ID        int       `json:"id" torm:"primary_key,auto_increment"`
    Username  string    `json:"username" torm:"type:varchar,size:50,unique,index"`
    Email     string    `json:"email" torm:"type:varchar,size:100,unique"`
    Age       int       `json:"age" torm:"type:int,default:0"`
    Status    string    `json:"status" torm:"type:varchar,size:20,default:active"`
    CreatedAt time.Time `json:"created_at" torm:"auto_create_time"`
    UpdatedAt time.Time `json:"updated_at" torm:"auto_update_time"`
}

func main() {
    // 1. 配置数据库连接
    torm.AddConnection("default", &torm.Config{
        Driver:   "mysql",
        Host:     "localhost",
        Port:     3306,
        Database: "myapp",
        Username: "root",
        Password: "password",
    })

    // 2. 自动创建表结构
    user := &User{}
    user.AutoMigrate()

    // 3. 创建记录
    newUser := &User{
        Username: "zhangsan",
        Email:    "zhangsan@example.com",
        Age:      25,
    }
    newUser.Save()

    // 4. 查询记录
    users, _ := torm.Table("users").
        Where("status", "=", "active").
        Where("age", ">=", 18).
        OrderBy("created_at", "desc").
        Get()

    fmt.Printf("查询到 %d 个用户\n", len(users))
}
```

## 📖 核心功能

### 查询构建器

```go
// 基础查询
users, _ := torm.Table("users").
    Select("id", "username", "email").
    Where("status", "=", "active").
    OrderBy("created_at", "desc").
    Limit(10).
    Get()

// 参数化查询
users, _ := torm.Table("users").
    Where("status = ? AND age >= ?", "active", 18).
    Get()

// 数组参数
users, _ := torm.Table("users").
    Where("status IN (?)", []string{"active", "premium"}).
    Get()

// 聚合查询
count, _ := torm.Table("users").Where("status", "=", "active").Count()
sum, _ := torm.Table("orders").Where("user_id", "=", 1).Sum("amount")

// 分页
result, _ := torm.Table("users").Paginate(1, 20)
```

### 事务处理

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

### 模型关联

```go
// 定义关联
func (u *User) Posts() *model.HasMany {
    return u.HasMany(&Post{}, "user_id", "id")
}

func (u *User) Profile() *model.HasOne {
    return u.HasOne(&Profile{}, "user_id", "id")
}

// 预加载
users, _ := torm.Table("users").With("Posts", "Profile").Get()
```

### 缓存查询

```go
// 缓存查询结果 5 分钟
users, _ := torm.Table("users").
    Where("status", "=", "active").
    Cache(5 * time.Minute).
    Get()

// 带标签的缓存
users, _ := torm.Table("users").
    CacheWithTags(5*time.Minute, "users", "active").
    Get()

// 清除缓存
torm.ClearCacheByTags("users")
```

### 读写分离

```go
torm.AddConnection("default", &torm.Config{
    Driver:     "mysql",
    Host:       "master.db.com",
    Database:   "myapp",
    ReadHosts:  []string{"slave1.db.com", "slave2.db.com"},
    WriteHosts: []string{"master.db.com"},
})

// 读操作自动路由到从库
users, _ := torm.Table("users").Get()

// 写操作自动路由到主库
torm.Table("users").Insert(data)
```

## 🏷️ TORM 标签

```go
type Product struct {
    torm.BaseModel
    
    // 主键和自增
    ID int64 `torm:"primary_key,auto_increment,comment:产品ID"`
    
    // 字符串类型
    Name string `torm:"type:varchar,size:200,comment:产品名称"`
    SKU  string `torm:"type:varchar,size:50,unique,comment:产品编码"`
    
    // 数值精度
    Price float64 `torm:"type:decimal,precision:10,scale:2,comment:价格"`
    Stock int     `torm:"type:int,unsigned,default:0,comment:库存"`
    
    // 索引
    Category string `torm:"type:varchar,size:50,index,comment:分类"`
    
    // 外键
    BrandID int `torm:"type:int,references:brands.id,on_delete:cascade"`
    
    // 时间戳
    CreatedAt time.Time `torm:"auto_create_time"`
    UpdatedAt time.Time `torm:"auto_update_time"`
}
```

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./db/...
go test ./model/...

# 查看测试覆盖率
go test -cover ./...
```

## 📚 文档

- [快速开始](http://torm.site/docs.html?doc=quick-start)
- [配置指南](http://torm.site/docs.html?doc=configuration)
- [查询构建器](http://torm.site/docs.html?doc=query-builder)
- [模型系统](http://torm.site/docs.html?doc=model-system)
- [关联关系](http://torm.site/docs.html?doc=relationships)
- [事务处理](http://torm.site/docs.html?doc=transactions)
- [数据迁移](http://torm.site/docs.html?doc=migrations)
- [缓存系统](http://torm.site/docs.html?doc=caching)
- [示例代码](http://torm.site/docs.html?doc=examples)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 📄 许可证

Apache 2.0 License

## 🔗 相关链接

- [GitHub](https://github.com/zhoudm1743/torm)
- [官方文档](http://torm.site)
- [问题反馈](https://github.com/zhoudm1743/torm/issues)
- [讨论区](https://github.com/zhoudm1743/torm/discussions)
