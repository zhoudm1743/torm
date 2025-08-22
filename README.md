# TORM - Go高性能ORM框架

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.18-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Documentation](https://img.shields.io/badge/docs-torm.site-brightgreen.svg)](http://torm.site)
[![Version](https://img.shields.io/badge/version-1.1.6-orange.svg)](https://github.com/zhoudm1743/torm/releases)

TORM是一个基于Go语言开发的高性能ORM（对象关系映射）框架，灵感来源于PHP ThinkORM。它提供了简洁易用的API、强大的查询构造器、完整的模型系统以及丰富的功能。
如果您觉得TORM有帮助到您，请帮忙给个star

## 🌐 官方网站

**官网地址**: [torm.site](http://torm.site)

- 📚 [完整文档](http://torm.site/docs.html)
- 🚀 [快速开始](http://torm.site/docs.html?doc=quick-start)
- 💡 [示例代码](http://torm.site/docs.html?doc=examples)
- ⚙️ [配置指南](http://torm.site/docs.html?doc=configuration)

## ✨ 主要特性

### 🆕 v1.1.0 新功能

#### 🔍 First/Find 增强功能
- **双重收益**: 支持指针填充 + 返回原始数据，一次调用双重收益
- **灵活使用**: 既可以只填充当前模型，也可以同时填充传入指针
- **统一接口**: BaseModel 和 db 包方法统一支持

#### 🔑 自定义主键系统
- **UUID主键**: 支持 UUID 作为主键
- **复合主键**: 支持多字段复合主键（多租户场景）
- **标签识别**: 使用 `primary:"true"` 标签自动识别主键
- **灵活配置**: 支持任意类型和数量的主键字段

#### 🔗 关联预加载 (Eager Loading)
- **N+1解决**: 彻底解决N+1查询问题，性能提升10倍+
- **批量加载**: 智能批量加载关联数据
- **深度关联**: 支持多层级关联预加载
- **缓存优化**: 关联数据智能缓存

#### 📄 分页器系统
- **传统分页**: 基于 LIMIT/OFFSET 的传统分页
- **游标分页**: 适用于大数据量的高性能游标分页
- **灵活配置**: 可配置页面大小、排序等参数

#### 🔍 JSON查询支持
- **跨数据库**: MySQL、PostgreSQL、SQLite 统一语法
- **JSONPath**: 支持复杂的 JSONPath 查询语法
- **类型安全**: 查询结果自动类型转换

#### 🏗️ 高级查询功能
- **子查询**: EXISTS、NOT EXISTS、IN、NOT IN 子查询
- **窗口函数**: ROW_NUMBER、RANK、聚合窗口函数
- **复杂条件**: 支持复杂的条件组合和嵌套

### ✅ 已实现功能

#### 🔧 核心数据库功能
- **多数据库支持**: MySQL、PostgreSQL（完整支持）、SQLite、SQL Server、MongoDB
- **连接池管理**: 高效的数据库连接池，支持连接复用和自动回收
- **事务支持**: 完整的事务操作，支持嵌套事务和事务回滚
- **查询构造器**: 流畅的链式调用API，支持复杂查询构建

#### 🏗️ 查询构造器
- **基础查询**: SELECT、INSERT、UPDATE、DELETE操作
- **条件查询**: WHERE、WHERE IN、WHERE BETWEEN、WHERE NULL等
- **连接查询**: INNER JOIN、LEFT JOIN、RIGHT JOIN支持
- **聚合查询**: GROUP BY、HAVING、COUNT、SUM等
- **排序分页**: ORDER BY、LIMIT、OFFSET、分页查询
- **原生SQL**: 支持原生SQL片段和参数绑定
- **查询克隆**: 支持查询对象克隆和复用

#### 📊 模型系统 (Active Record)
- **Active Record模式**: 面向对象的数据库操作，模型内置查询方法
- **属性管理**: 动态属性设置和获取，支持脏数据检测
- **数据验证**: 内置验证规则和自定义验证
- **事件钩子**: BeforeSave、AfterSave、BeforeCreate等事件
- **时间戳**: 自动管理created_at、updated_at字段
- **软删除**: 支持软删除和硬删除操作
- **模型重载**: 从数据库重新加载模型数据

#### 🚀 缓存系统
- **内存缓存**: 高性能的内存缓存实现
- **TTL支持**: 支持缓存过期时间设置
- **并发安全**: 读写锁保证并发安全
- **自动清理**: 定期清理过期缓存项
- **缓存统计**: 缓存命中率和使用情况统计

#### 📝 日志系统
- **多级别日志**: DEBUG、INFO、WARN、ERROR、FATAL
- **文件日志**: 支持日志写入文件
- **SQL日志**: 专门的SQL查询日志记录
- **结构化日志**: 支持字段化日志记录
- **日志过滤**: 基于级别的日志过滤

#### 🧪 测试覆盖
- **单元测试**: 完整的单元测试覆盖
- **集成测试**: 真实数据库环境测试
- **并发测试**: 高并发场景测试
- **性能测试**: 查询性能基准测试

### 🚧 计划中功能

- **断点重连**: 数据库连接断线重连
- **分布式事务**: 跨数据库事务支持
- **数据迁移**: 数据库结构迁移工具 ✅ (已部分实现)
- **代码生成**: 模型和迁移代码自动生成
- **数据工厂**: 测试数据生成器
- **读写分离**: 主从数据库读写分离

## 🚀 快速开始

### 📦 安装

更详细的安装指南请访问：[torm.site/docs.html?doc=installation](http://torm.site/docs.html?doc=installation)

```bash
go mod init your-project
go get github.com/zhoudm1743/torm
go get github.com/go-sql-driver/mysql    # MySQL 支持
go get github.com/lib/pq                 # PostgreSQL 支持
```

> 💡 **提示**: 完整的安装和配置教程，请访问官网的 [安装指南](http://torm.site/docs.html?doc=installation)。

### 🎯 核心功能演示

#### 1. 数据库连接

```go
package main

import (
    "log"
    "github.com/zhoudm1743/torm/db"
)

func main() {
    // 配置数据库连接
    config := &db.Config{
        Driver:   "mysql",
        Host:     "localhost",
        Port:     3306,
        Database: "your_database",
        Username: "your_username",
        Password: "your_password",
        Charset:  "utf8mb4",
        MaxOpenConns: 100,
        MaxIdleConns: 10,
    }

    // 添加连接配置
    err := db.AddConnection("default", config)
    if err != nil {
        log.Fatal(err)
    }
}
```

#### 2. 模型定义与自定义主键

```go
import (
    "time"
    "github.com/zhoudm1743/torm/model"
)

// 默认主键模型
type User struct {
    model.BaseModel
    ID        interface{} `json:"id" db:"id"`
    Name      string      `json:"name" db:"name"`
    Email     string      `json:"email" db:"email"`
    Age       int         `json:"age" db:"age"`
    CreatedAt time.Time   `json:"created_at" db:"created_at"`
}

// UUID主键模型
type Product struct {
    model.BaseModel
    UUID      string    `json:"uuid" db:"uuid" primary:"true"`  // UUID主键
    Name      string    `json:"name" db:"name"`
    Price     float64   `json:"price" db:"price"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// 复合主键模型（多租户场景）
type UserRole struct {
    model.BaseModel
    TenantID  string    `json:"tenant_id" db:"tenant_id" primary:"true"`  // 复合主键1
    UserID    string    `json:"user_id" db:"user_id" primary:"true"`      // 复合主键2
    Role      string    `json:"role" db:"role"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func NewUser() *User {
    user := &User{BaseModel: *model.NewBaseModel()}
    user.SetTable("users")
    return user
}

func NewProduct() *Product {
    product := &Product{BaseModel: *model.NewBaseModel()}
    product.SetTable("products")
    // 自动检测主键标签
    product.DetectPrimaryKeysFromStruct(product)
    return product
}
```

#### 3. First/Find 增强功能

```go
func demonstrateFirstFind() {
    // First方法 - 只填充当前模型
    user1 := NewUser()
    result1, err := user1.Where("id", "=", 1).First()
    if err != nil {
        log.Printf("查询失败: %v", err)
    } else {
        log.Printf("当前模型: Name=%s, Age=%d", user1.Name, user1.Age)
        log.Printf("返回数据: %+v", result1)
    }
    
    // First方法 - 同时填充传入的指针
    user2 := NewUser()
    var anotherUser User
    result2, err := user2.Where("id", "=", 2).First(&anotherUser)
    if err != nil {
        log.Printf("查询失败: %v", err)
    } else {
        log.Printf("当前模型: %s", user2.Name)
        log.Printf("传入指针: %s", anotherUser.Name)
        log.Printf("原始数据: %+v", result2)
    }

    // Find方法 - 根据主键查找并填充指针
    user3 := NewUser()
    var targetUser User
    result3, err := user3.Find(1, &targetUser)
    if err != nil {
        log.Printf("Find失败: %v", err)
    } else {
        log.Printf("当前模型: %s", user3.Name)
        log.Printf("传入指针: %s", targetUser.Name)
    }
}
```

#### 4. db包增强功能

```go
func demonstrateDBPackage() {
    // db.Table().First() - 只返回map
    query1, err := db.Table("users", "default")
    if err == nil {
        result1, err := query1.Where("id", "=", 1).First()
        if err == nil {
            log.Printf("db.First() 结果: %s", result1["name"])
        }
    }

    // db.Table().First(&model) - 填充指针 + 返回map
    query2, err := db.Table("users", "default")
    if err == nil {
        var user User
        result2, err := query2.Where("id", "=", 1).First(&user)
        if err == nil {
            log.Printf("填充的模型: Name=%s", user.Name)
            log.Printf("返回的map: %+v", result2)
        }
    }
}
```

#### 5. 查询构造器

```go
// 创建查询
query, err := db.Table("users")
if err != nil {
    panic(err)
}

// 基础链式查询
users, err := query.
    Select("id", "name", "email", "age").
    Where("status", "=", "active").
    Where("age", ">", 18).
    OrderBy("created_at", "desc").
    Limit(10).
    Get()

// 复杂条件查询
result, err := query.
    Select("id", "name", "email").
    Where("age", "BETWEEN", []interface{}{20, 40}).
    WhereIn("status", []interface{}{"active", "pending"}).
    OrderBy("age", "ASC").
    OrderBy("name", "DESC").
    Limit(10).
    Get()

// 聚合查询
count, err := query.Where("status", "=", "active").Count()
totalAge, err := query.Where("status", "=", "active").Sum("age")

// 插入数据
id, err := query.Insert(map[string]interface{}{
    "name":  "张三",
    "email": "zhangsan@example.com",
    "age":   25,
})

// 更新数据
affected, err := query.
    Where("id", "=", id).
    Update(map[string]interface{}{
        "age": 26,
    })

// 删除数据
affected, err = query.
    Where("id", "=", id).
    Delete()
```

#### 6. Active Record 模式

```go
// 创建用户模型
user := NewUser()
user.Name = "李四"
user.Email = "lisi@example.com"
user.Age = 30

// 保存到数据库
err := user.Save()
if err != nil {
    panic(err)
}

// 使用内置查询方法
users, err := user.Where("age", ">", 25).
    OrderBy("created_at", "desc").
    Limit(10).
    All()

// 根据主键查找
user2 := NewUser()
_, err = user2.Find(user.ID)
if err != nil {
    panic(err)
}

// 更新用户
user2.Age = 31
err = user2.Save()

// 删除用户
err = user2.Delete()
```

#### 7. 关联预加载

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

#### 8. 分页查询

```go
// 简单分页
result, err := userQuery.Paginate(1, 10) // 第1页，每页10条

// 高级分页器
paginator := paginator.NewQueryPaginator(userQuery)
paginationResult, err := paginator.SetPerPage(15).SetPage(2).Paginate()
```

#### 9. 事务操作

```go
err := db.Transaction(func(tx db.TransactionInterface) error {
    // 在事务中执行操作
    _, err := tx.Exec("INSERT INTO users (name, email) VALUES (?, ?)", "事务用户", "tx@example.com")
    if err != nil {
        return err // 自动回滚
    }

    _, err = tx.Exec("UPDATE users SET status = ? WHERE id = ?", "active", 1)
    if err != nil {
        return err // 自动回滚
    }

    return nil // 自动提交
})
```

## 🏗️ 项目架构

```
torm/
├── pkg/                    # 核心包
│   ├── db/                # 数据库核心
│   │   ├── interfaces.go  # 核心接口定义
│   │   ├── config.go      # 数据库配置
│   │   ├── manager.go     # 连接管理器
│   │   ├── mysql.go       # MySQL连接器
│   │   ├── postgresql.go  # PostgreSQL连接器
│   │   ├── sqlite.go      # SQLite连接器
│   │   └── query.go       # 查询构造器
│   ├── model/             # 模型系统
│   │   ├── base_model.go  # 基础模型
│   │   ├── base_model_relations.go # 关联系统
│   │   └── relation.go    # 关联关系
│   ├── cache/             # 缓存系统
│   │   └── memory_cache.go # 内存缓存
│   ├── logger/            # 日志系统
│   │   └── logger.go      # 日志记录器
│   ├── migration/         # 数据迁移
│   │   ├── migration.go   # 迁移管理器
│   │   └── schema.go      # 结构构建器
│   ├── paginator/         # 分页器
│   └── query/             # 高级查询
│       └── advanced_query.go
├── examples/              # 示例代码
├── tests/                 # 单元测试
├── wiki/                  # 文档
├── web/                   # 官方网站
└── README.md             # 项目说明
```

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

MIT License

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
