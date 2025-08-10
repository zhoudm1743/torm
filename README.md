# TORM - Go高性能ORM框架

TORM是一个基于Go语言开发的高性能ORM（对象关系映射）框架，灵感来源于PHP ThinkORM。它提供了简洁易用的API、强大的查询构造器、完整的模型系统以及丰富的企业级功能。

## ✨ 主要特性

### ✅ 已实现功能

#### 🔧 核心数据库功能
- **多数据库支持**: MySQL、PostgreSQL、SQLite、SQL Server
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

#### 📊 模型系统
- **Active Record模式**: 面向对象的数据库操作
- **属性管理**: 动态属性设置和获取
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

- **关联关系**: HasOne、HasMany、BelongsTo、ManyToMany
- **JSON查询**: JSON字段查询支持
- **断点重连**: 数据库连接断线重连
- **MongoDB支持**: NoSQL数据库支持
- **分布式事务**: 跨数据库事务支持
- **数据迁移**: 数据库结构迁移工具
- **分页器**: 高级分页功能

## 🚀 快速开始

### 安装依赖

```bash
go mod init your-project
go get github.com/go-sql-driver/mysql
go get github.com/stretchr/testify
```

### 基础使用

#### 1. 数据库连接

```go
package main

import (
    "context"
    "time"
    "torm/pkg/db"
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
        ConnMaxLifetime: time.Hour,
    }

    // 添加连接配置
    err := db.AddConnection("default", config)
    if err != nil {
        panic(err)
    }
}
```

#### 2. 查询构造器

```go
// 创建查询
query, err := db.Table("users")
if err != nil {
    panic(err)
}

// 链式查询
users, err := query.
    Select("id", "name", "email").
    Where("age", ">", 18).
    Where("status", "=", "active").
    OrderBy("created_at", "desc").
    Limit(10).
    Get(context.Background())

// 插入数据
id, err := query.Insert(context.Background(), map[string]interface{}{
    "name":  "张三",
    "email": "zhangsan@example.com",
    "age":   25,
})

// 更新数据
affected, err := query.
    Where("id", "=", id).
    Update(context.Background(), map[string]interface{}{
        "age": 26,
    })

// 删除数据
affected, err = query.
    Where("id", "=", id).
    Delete(context.Background())
```

#### 3. 模型系统

```go
import "torm/pkg/model"

// 创建用户模型
user := model.NewUser()
user.SetName("李四").SetEmail("lisi@example.com").SetAge(30)

// 保存到数据库
err := user.Save(context.Background())
if err != nil {
    panic(err)
}

// 查找用户
user2 := model.NewUser()
err = user2.Find(context.Background(), user.GetID())
if err != nil {
    panic(err)
}

// 更新用户
user2.SetAge(31)
err = user2.Save(context.Background())

// 删除用户
err = user2.Delete(context.Background())
```

#### 4. 缓存系统

```go
import "torm/pkg/cache"

// 创建内存缓存
memCache := cache.NewMemoryCache()

// 设置缓存
err := memCache.Set(context.Background(), "user:1", userData, 5*time.Minute)

// 获取缓存
data, err := memCache.Get(context.Background(), "user:1")

// 检查缓存是否存在
exists, err := memCache.Has(context.Background(), "user:1")

// 删除缓存
err = memCache.Delete(context.Background(), "user:1")
```

#### 5. 日志系统

```go
import "torm/pkg/logger"

// 创建日志记录器
appLogger := logger.NewLogger(logger.INFO)

// 记录日志
appLogger.Info("用户登录", "user_id", 123, "ip", "192.168.1.1")
appLogger.Warn("内存使用率高", "usage", "85%")
appLogger.Error("数据库连接失败", "error", err.Error())

// SQL日志记录器
sqlLogger := logger.NewSQLLogger(logger.DEBUG, true)
sqlLogger.LogQuery("SELECT * FROM users WHERE id = ?", []interface{}{1}, time.Millisecond*10)
```

#### 6. 事务操作

```go
err := db.Transaction(context.Background(), func(tx db.TransactionInterface) error {
    // 在事务中执行操作
    _, err := tx.Exec(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "事务用户", "tx@example.com")
    if err != nil {
        return err // 自动回滚
    }

    _, err = tx.Exec(ctx, "UPDATE users SET status = ? WHERE id = ?", "active", 1)
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
│   │   ├── interface.go   # 核心接口定义
│   │   ├── config.go      # 数据库配置
│   │   ├── manager.go     # 连接管理器
│   │   ├── mysql.go       # MySQL连接器
│   │   ├── sqlite.go      # SQLite连接器
│   │   └── query_builder.go # 查询构造器
│   ├── model/             # 模型系统
│   │   ├── base_model.go  # 基础模型
│   │   └── user.go        # 用户模型示例
│   ├── cache/             # 缓存系统
│   │   └── memory_cache.go # 内存缓存
│   └── logger/            # 日志系统
│       └── logger.go      # 日志记录器
├── examples/              # 示例代码
├── tests/                 # 单元测试
├── docs/                  # 文档
└── README.md             # 项目说明
```

## 🧪 测试

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

## 📊 性能特点

- **高并发**: 支持数千并发连接
- **内存优化**: 高效的内存使用和垃圾回收
- **缓存加速**: 智能缓存减少数据库访问
- **连接池**: 连接复用提高性能
- **预编译语句**: 防止SQL注入，提高执行效率

## 🔧 配置选项

### 数据库配置

```go
config := &db.Config{
    Driver:   "mysql",           // 数据库驱动
    Host:     "localhost",       // 主机地址
    Port:     3306,             // 端口号
    Database: "test",           // 数据库名
    Username: "root",           // 用户名
    Password: "password",       // 密码
    Charset:  "utf8mb4",        // 字符集
    Timezone: "UTC",            // 时区
    
    // 连接池配置
    MaxOpenConns:    100,               // 最大打开连接数
    MaxIdleConns:    10,                // 最大空闲连接数
    ConnMaxLifetime: time.Hour,         // 连接最大生存时间
    ConnMaxIdleTime: time.Minute * 30,  // 连接最大空闲时间
    
    // 其他配置
    LogQueries: true,           // 是否记录查询日志
    Debug:      false,          // 是否开启调试模式
}
```

## 📚 示例

查看 `examples/` 目录获取更多使用示例：

- `cache_logger_demo.go` - 缓存和日志系统演示
- 更多示例持续更新中...

## 🤝 贡献

欢迎提交Issue和Pull Request来帮助改进TORM！

## 📄 许可证

MIT License

## 🔗 相关链接

- [Go官方文档](https://golang.org/doc/)
- [database/sql包文档](https://pkg.go.dev/database/sql)
- [MySQL驱动](https://github.com/go-sql-driver/mysql)
- [测试框架Testify](https://github.com/stretchr/testify)

---

**TORM** - 让Go数据库操作更简单、更高效！ 🚀 