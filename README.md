# TORM - Go高性能ORM框架

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.19-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Documentation](https://img.shields.io/badge/docs-torm.site-brightgreen.svg)](http://torm.site)
[![Version](https://img.shields.io/badge/version-1.2.0-orange.svg)](https://github.com/zhoudm1743/torm/releases)

TORM 是一个完全重构的高性能Go ORM框架，提供了极简的API设计、强大的自动迁移、完整的跨数据库支持以及丰富的TORM标签系统。
如果您觉得TORM有帮助到您，请帮忙给个star ⭐

## 🌐 官方网站

**官网地址**: [torm.site](http://torm.site)

- 📚 [完整文档](http://torm.site/docs.html)
- 🚀 [快速开始](http://torm.site/docs.html?doc=quick-start)
- 💡 [示例代码](http://torm.site/docs.html?doc=examples)
- ⚙️ [配置指南](http://torm.site/docs.html?doc=configuration)

## ✨ 主要特性

### 🚀 亮点

#### 🎯 极简API设计
- **零配置启动**: 一行代码完成数据库连接和表创建
- **统一标签语法**: 全新的 `torm` 标签系统，支持30+种配置选项
- **智能类型推断**: 自动映射Go类型到数据库类型，支持跨数据库兼容

#### 🔄 强大的自动迁移
- **AutoMigrate**: 根据模型结构体自动创建和更新表结构
- **智能差异检测**: 自动检测模型变更，只更新必要的字段
- **跨数据库兼容**: MySQL、PostgreSQL、SQLite无缝切换
- **安全更新**: 保护现有数据，智能处理字段变更

#### 🏷️ 丰富的TORM标签
- **数据类型**: `type:varchar,size:100,precision:10,scale:2`
- **约束条件**: `primary_key,auto_increment,unique,not_null`
- **索引优化**: `index,index:custom_name,fulltext`
- **默认值**: `default:0,default:current_timestamp`
- **时间戳**: `auto_create_time,auto_update_time`
- **外键关系**: `references:users.id,on_delete:cascade`

#### 🔗 强大的查询构建器
- **参数化查询**: 完全支持占位符，防止SQL注入
- **跨数据库语法**: 自动适配MySQL的`?`和PostgreSQL的`$N`占位符
- **数组参数**: 原生支持`IN (?, ?, ?)`数组参数展开
- **复杂条件**: WHERE、OR、IN、LIKE、BETWEEN、EXISTS全支持

### ✅ 已实现功能

#### 🔧 核心数据库功能
- **多数据库支持**: MySQL、PostgreSQL（完整支持）、SQLite、SQL Server、MongoDB
- **连接池管理**: 高效的数据库连接池，支持连接复用和自动回收
- **事务支持**: 完整的事务操作，支持嵌套事务和事务回滚
- **查询构造器**: 流畅的链式调用API，支持复杂查询构建

## 🚀 快速开始

### 📦 安装

```bash
go mod init your-project
go get github.com/zhoudm1743/torm
go get github.com/go-sql-driver/mysql    # MySQL 支持
go get github.com/lib/pq                 # PostgreSQL 支持
```

### 🎯 核心功能演示

#### 1. 极简模型定义

```go
package main

import (
    "time"
    "github.com/zhoudm1743/torm"
)

// 用户模型 - 使用丰富的TORM标签
type User struct {
    torm.BaseModel
    ID        int       `json:"id" torm:"primary_key,auto_increment"`
    Username  string    `json:"username" torm:"type:varchar,size:50,unique,index"`
    Email     string    `json:"email" torm:"type:varchar,size:100,unique"`
    Age       int       `json:"age" torm:"type:int,default:0"`
    Salary    float64   `json:"salary" torm:"type:decimal,precision:10,scale:2"`
    Status    string    `json:"status" torm:"type:varchar,size:20,default:active"`
    IsActive  bool      `json:"is_active" torm:"type:boolean,default:1"`
    CreatedAt time.Time `json:"created_at" torm:"auto_create_time"`
    UpdatedAt time.Time `json:"updated_at" torm:"auto_update_time"`
}

func main() {
    // 1. 配置数据库（支持MySQL、PostgreSQL、SQLite）
    torm.AddConnection("default", &torm.Config{
        Driver:   "mysql",
        Host:     "localhost", 
        Port:     3306,
        Database: "testdb",
        Username: "root",
        Password: "password",
    })

    // 2. 自动创建表结构（一行代码完成迁移）
    user := &User{}
    user.AutoMigrate()
    
    // 3. 完成！开始使用
}
```

#### 2. CRUD操作

```go
// 创建记录
user := &User{
    Username: "张三",
    Email:    "zhangsan@example.com", 
    Age:      25,
    Status:   "active",
}
user.Save()

// 查询记录
users, _ := torm.Table("users").
    Where("status", "=", "active").
    Where("age", ">=", 18).
    OrderBy("created_at", "desc").
    Get()

// 参数化查询（支持数组参数）
activeUsers, _ := torm.Table("users").
    Where("status IN (?)", []string{"active", "premium"}).
    Where("age BETWEEN ? AND ?", 18, 65).
    Get()

// 聚合查询
count, _ := torm.Table("users").
    Where("status", "=", "active").
    Count()

// 更新记录 
torm.Table("users").
    Where("id", "=", 1).
    Update(map[string]interface{}{
        "age":    26,
        "status": "premium",
    })

// 删除记录
torm.Table("users").
    Where("status", "=", "inactive").
    Delete()
```

#### 3. 跨数据库支持

```go
// MySQL配置
torm.AddConnection("mysql", &torm.Config{
    Driver:   "mysql",
    Host:     "localhost",
    Port:     3306,
    Database: "testdb",
    Username: "root",
    Password: "password",
})

// PostgreSQL配置  
torm.AddConnection("postgres", &torm.Config{
    Driver:   "postgres",
    Host:     "localhost",
    Port:     5432,
    Database: "testdb", 
    Username: "postgres",
    Password: "password",
    SSLMode:  "disable",
})

// SQLite配置
torm.AddConnection("sqlite", &torm.Config{
    Driver:   "sqlite",
    Database: "test.db",
})

// 同一模型自动适配不同数据库
user := &User{}
user.SetConnection("mysql")    // 使用MySQL
user.AutoMigrate()

user.SetConnection("postgres") // 切换到PostgreSQL  
user.AutoMigrate()            // 相同的模型，不同的数据库实现
```

#### 4. 事务处理

```go
// 自动事务管理
err := torm.Transaction(func(tx torm.TransactionInterface) error {
    // 在事务中执行多个操作
    _, err := tx.Exec("INSERT INTO users (username, email) VALUES (?, ?)", 
                     "张三", "zhangsan@example.com")
    if err != nil {
        return err // 自动回滚
    }

    _, err = tx.Exec("UPDATE departments SET budget = budget + 1000 WHERE id = ?", 1)
    if err != nil {
        return err // 自动回滚
    }

    return nil // 自动提交
})

if err != nil {
    log.Printf("事务失败: %v", err)
}
```

#### 5. 强大的TORM标签系统

```go
type Product struct {
    torm.BaseModel
    
    // 主键和自增
    ID int64 `torm:"primary_key,auto_increment,comment:产品ID"`
    
    // 字符串类型和长度
    Name     string `torm:"type:varchar,size:200,comment:产品名称"`
    SKU      string `torm:"type:varchar,size:50,unique,comment:产品编码"`
    Category string `torm:"type:varchar,size:20,default:normal,index"`
    
    // 数值类型和精度
    Price  float64 `torm:"type:decimal,precision:10,scale:2,comment:价格"`
    Stock  int     `torm:"type:int,unsigned,default:0,comment:库存"`
    Weight float64 `torm:"type:decimal,precision:8,scale:3,comment:重量"`
    
    // 布尔和默认值
    IsActive bool `torm:"type:boolean,default:1,comment:是否启用"`
    IsNew    bool `torm:"type:boolean,default:0,comment:是否新品"`
    
    // 文本类型
    Description string `torm:"type:text,comment:产品描述"`
    Features    string `torm:"type:longtext,comment:产品特性"`
    
    // 外键和关联
    CategoryID int `torm:"type:int,references:categories.id,on_delete:set_null"`
    BrandID    int `torm:"type:int,references:brands.id,on_delete:cascade"`
    
    // 自动时间戳
    CreatedAt time.Time `torm:"auto_create_time,comment:创建时间"`
    UpdatedAt time.Time `torm:"auto_update_time,comment:更新时间"`
}

// 一行代码创建完整的表结构
product := &Product{}
product.AutoMigrate() // 自动创建表、索引、外键约束
```

## 📊 性能优势

- **零反射查询**: 直接SQL构建，避免反射开销
- **智能占位符**: 自动适配数据库占位符语法
- **连接池优化**: 高效的数据库连接复用
- **批量操作**: 原生支持批量插入和更新
- **索引自动化**: 根据模型标签自动创建索引

## 🆚 v1.2.0 对比

| 特性 | v1.1.x | v1.2.0 |
|------|--------|--------|
| **模型定义** | 复杂配置 | 零配置，TORM标签 |
| **表创建** | 手动迁移 | 一行AutoMigrate |
| **跨数据库** | 有限支持 | 完全兼容 |
| **占位符** | 手动处理 | 自动适配 |
| **数组参数** | 不支持 | 原生支持 |
| **表更新** | 跳过检查 | 智能差异检测 |


## 🧪 测试

### 运行测试

运行所有测试：

```bash
go test ./tests/
```

运行特定测试：

```bash
go test -v ./tests/query_builder_test.go
go test -v ./tests/model_test.go
go test -v ./tests/cache_test.go
go test -v ./tests/logger_test.go
```

### 测试覆盖率

查看测试覆盖率：

```bash
go test -cover ./tests/
```

### 更多测试信息
访问 [测试文档](http://torm.site/docs.html?doc=troubleshooting) 了解更多测试相关信息。

## 📊 性能特点

- **高并发**: 支持数千并发连接
- **内存优化**: 高效的内存使用和垃圾回收
- **缓存加速**: 智能缓存减少数据库访问
- **连接池**: 连接复用提高性能
- **预编译语句**: 防止SQL注入，提高执行效率
- **查询优化**: N+1查询解决，关联预加载

## 📚 示例

### 本地示例
查看 `examples/` 目录获取更多使用示例：

- `main.go` - 核心功能演示
- 更多示例持续更新中...

### 在线示例
访问 [torm.site/docs.html?doc=examples](http://torm.site/docs.html?doc=examples) 查看完整的在线示例和教程。

## 🤝 贡献

我们欢迎所有形式的贡献！在参与之前，请阅读我们的 [贡献指南](http://torm.site/docs.html?doc=contributing)。

### 如何贡献
1. **报告问题**: 在 [GitHub Issues](https://github.com/zhoudm1743/torm/issues) 报告bug或提出功能请求
2. **代码贡献**: Fork项目，创建特性分支，提交Pull Request
3. **文档改进**: 帮助改进文档和示例
4. **社区讨论**: 参与 [讨论区](https://github.com/zhoudm1743/torm/discussions) 的技术讨论

### 开发指南
详细的开发指南请访问：[torm.site/docs.html?doc=contributing](http://torm.site/docs.html?doc=contributing)

## 📄 许可证

Apache2.0 License

## 🔗 相关链接

### 📖 文档与学习
- [TORM官方网站](http://torm.site) - 完整的文档和教程
- [Go官方文档](https://golang.org/doc/)
- [database/sql包文档](https:/.go.dev/database/sql)

### 🛠️ 依赖项目
- [MySQL驱动](https://github.com/go-sql-driver/mysql)
- [PostgreSQL驱动](https://github.com/lib/pq)
- [测试框架Testify](https://github.com/stretchr/testify)

### 💬 社区与支持
- [GitHub Issues](https://github.com/zhoudm1743/torm/issues) - 问题报告与功能请求
- [讨论区](https://github.com/zhoudm1743/torm/discussions) - 社区讨论

---

**TORM v1.1.0** - 让Go数据库操作更简单、更高效！ 🚀

访问 [torm.site](http://torm.site) 获取最新文档和教程。 
