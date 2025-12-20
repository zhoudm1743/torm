# 数据库支持

TORM 支持多种主流数据库，提供统一的 API 接口，让你可以轻松在不同数据库间切换。

## 🗄️ 支持的数据库

| 数据库 | 驱动 | 版本要求 | 状态 |
|--------|------|----------|------|
| MySQL | mysql | 5.7+ | ✅ 完整支持 |
| PostgreSQL | postgres | 10+ | ✅ 完整支持 |
| SQLite | sqlite | 3.25+ | ✅ 完整支持 |

## 🐬 MySQL

### 配置示例

```go
config := &torm.Config{
    Driver:   "mysql",
    Host:     "localhost",
    Port:     3306,
    Database: "myapp",
    Username: "root",
    Password: "password",
    Charset:  "utf8mb4",
}

err := torm.AddConnection("default", config)
```

### 特性支持

- ✅ 完整 SQL 支持
- ✅ 事务处理
- ✅ 连接池
- ✅ JSON 字段
- ✅ 全文索引
- ✅ 外键约束
- ✅ 占位符 `?` 自动处理

### 推荐配置

```go
config := &torm.Config{
    Driver:          "mysql",
    Host:            "localhost",
    Port:            3306,
    Database:        "myapp",
    Username:        "root",
    Password:        "password",
    Charset:         "utf8mb4",
    MaxOpenConns:    100,
    MaxIdleConns:    10,
    ConnMaxLifetime: time.Hour,
}
```

## 🐘 PostgreSQL

### 配置示例

```go
config := &torm.Config{
    Driver:   "postgres",
    Host:     "localhost",
    Port:     5432,
    Database: "myapp",
    Username: "postgres",
    Password: "password",
    SSLMode:  "disable",
}

err := torm.AddConnection("default", config)
```

### 特性支持

- ✅ 完整 SQL 支持
- ✅ JSONB 支持
- ✅ 数组类型
- ✅ 高级索引
- ✅ 窗口函数
- ✅ 占位符 `$N` 自动适配

### SSLMode 选项

```go
// 开发环境
config.SSLMode = "disable"

// 生产环境
config.SSLMode = "require"
config.SSLMode = "verify-ca"
config.SSLMode = "verify-full"
```

## 📁 SQLite

### 配置示例

```go
// 文件数据库
config := &torm.Config{
    Driver:   "sqlite",
    Database: "app.db",
}

// 内存数据库（测试用）
config := &torm.Config{
    Driver:   "sqlite",
    Database: ":memory:",
}

err := torm.AddConnection("default", config)
```

### 特性支持

- ✅ 轻量级部署
- ✅ 零配置
- ✅ 嵌入式应用
- ✅ 开发测试推荐
- ⚠️ 并发写入限制

### 适用场景

- 开发和测试环境
- 嵌入式应用
- 小型单机应用
- 原型开发

## 🔄 多数据库连接

TORM 支持同时连接多个数据库：

```go
// MySQL 主库
torm.AddConnection("mysql_main", &torm.Config{
    Driver:   "mysql",
    Host:     "mysql-master",
    Database: "production",
    Username: "app_user",
    Password: "password",
})

// PostgreSQL 分析库
torm.AddConnection("pg_analytics", &torm.Config{
    Driver:   "postgres",
    Host:     "pg-analytics",
    Database: "analytics",
    Username: "analyst",
    Password: "password",
})

// SQLite 本地缓存
torm.AddConnection("sqlite_cache", &torm.Config{
    Driver:   "sqlite",
    Database: "cache.db",
})

// 使用不同连接
users, _ := torm.Table("users", "mysql_main").Get()
reports, _ := torm.Table("reports", "pg_analytics").Get()
cache, _ := torm.Table("cache", "sqlite_cache").Get()
```

## 🔀 跨数据库兼容

TORM 自动处理不同数据库的差异：

### 占位符适配

```go
// 同样的代码，自动适配不同数据库
query := torm.Table("users", connectionName).
    Where("status", "=", "active").
    Where("age", ">", 18)

// MySQL 生成: SELECT * FROM users WHERE status = ? AND age > ?
// PostgreSQL 生成: SELECT * FROM users WHERE status = $1 AND age > $2
```

### 随机排序

```go
// 自动适配数据库的随机函数
query.OrderRand()

// MySQL: ORDER BY RAND()
// PostgreSQL: ORDER BY RANDOM()
// SQLite: ORDER BY RANDOM()
```

### 数组参数展开

```go
// IN 查询自动展开
query.WhereIn("status", []interface{}{"active", "pending", "approved"})

// MySQL: WHERE status IN (?, ?, ?)
// PostgreSQL: WHERE status IN ($1, $2, $3)
```

## 📊 数据库选择建议

| 场景 | 推荐数据库 | 原因 |
|------|-----------|------|
| 生产环境 Web 应用 | MySQL | 成熟稳定，生态丰富 |
| 复杂数据分析 | PostgreSQL | 强大的 SQL 功能 |
| 开发测试 | SQLite | 零配置，快速启动 |
| 嵌入式应用 | SQLite | 无需服务器 |
| 高并发读写 | MySQL | 优秀的并发性能 |
| JSON 数据处理 | PostgreSQL | 原生 JSONB 支持 |

## 🔗 相关文档

- [配置文档](Configuration) - 详细配置选项
- [快速开始](Quick-Start) - 快速上手指南
- [查询构建器](Query-Builder) - SQL 查询构建
