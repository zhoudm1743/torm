# 数据库支持

TORM支持多种主流数据库，提供统一的API接口，让您可以轻松在不同数据库间切换。

## 🗄️ 支持的数据库

| 数据库 | 驱动 | 状态 | 版本要求 |
|--------|------|------|----------|
| MySQL | `mysql` | ✅ 完全支持 | 5.7+ / 8.0+ |
| PostgreSQL | `postgres` | ✅ 完全支持 | 11+ |
| SQLite | `sqlite` | ✅ 完全支持 | 3.8+ |
| MongoDB | `mongodb` | ✅ 完全支持 | 4.4+ |
| SQL Server | `sqlserver` | 🚧 基础支持 | 2017+ |

## 🔧 MySQL

### 配置示例

```go
config := &db.Config{
    Driver:   "mysql",
    Host:     "localhost",
    Port:     3306,
    Database: "myapp",
    Username: "root",
    Password: "password",
    Charset:  "utf8mb4",
    
    Options: map[string]string{
        "parseTime": "true",
        "loc":       "Local",
    },
}
```

### 特性支持

- ✅ 完整的SQL支持
- ✅ 事务处理
- ✅ 连接池
- ✅ JSON字段
- ✅ 全文索引
- ✅ 外键约束

## 🐘 PostgreSQL

### 配置示例

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

### 特性支持

- ✅ 完整的SQL支持
- ✅ JSONB支持
- ✅ 数组类型
- ✅ 高级索引
- ✅ 窗口函数

## 📁 SQLite

### 配置示例

```go
config := &db.Config{
    Driver:   "sqlite",
    Database: "app.db",
    
    Options: map[string]string{
        "foreign_keys": "on",
    },
}
```

### 特性支持

- ✅ 轻量级部署
- ✅ 零配置
- ✅ 嵌入式应用
- ⚠️ 并发限制

## 🍃 MongoDB

### 配置示例

```go
config := &db.Config{
    Driver:   "mongodb",
    Host:     "localhost",
    Port:     27017,
    Database: "myapp",
}
```

### 特性支持

- ✅ 文档存储
- ✅ 聚合管道
- ✅ 索引支持
- ✅ 事务支持（副本集）

## 🔄 数据库切换

TORM的设计允许您轻松在不同数据库间切换：

```go
// 开发环境 - SQLite
devConfig := &db.Config{
    Driver:   "sqlite",
    Database: "dev.db",
}

// 生产环境 - MySQL
prodConfig := &db.Config{
    Driver:   "mysql",
    Host:     "prod-db.example.com",
    Database: "myapp",
    Username: "app_user",
    Password: "secure_password",
}
```

---

**📚 更多信息请参考 [配置文档](Configuration) 和 [快速开始](Quick-Start)。** 