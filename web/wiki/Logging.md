# 日志系统

TORM 提供了完整的日志系统，支持多级别日志、SQL查询日志记录、结构化日志输出和灵活的日志配置。特别是SQL日志功能，能够帮助开发者调试和监控数据库查询性能。

## 📋 目录

- [快速开始](#快速开始)
- [SQL查询日志](#sql查询日志)
- [基础日志功能](#基础日志功能)
- [日志级别控制](#日志级别控制)
- [连接配置日志](#连接配置日志)
- [最佳实践](#最佳实践)

## 🚀 快速开始

### 启用SQL查询日志

```go
import "github.com/zhoudm1743/torm"

// 最简单的方式：启用SQL日志（DEBUG级别）
torm.EnableSQLLogging()

// 执行查询，SQL会自动记录到控制台
users, err := torm.Table("users", "default").
    Where("status", "=", "active").
    Get()

// 输出示例：
// 2024-01-20 10:30:15 [DEBUG] Query executed [sql SELECT * FROM users WHERE status = ? args [active] duration 2.5ms]
```

## 🗄️ SQL查询日志

### 基础SQL日志配置

SQL查询日志是TORM的核心功能，可以记录所有数据库操作，包括查询语句、参数和执行时间。

```go
// 方式1：使用便捷函数启用SQL日志
torm.EnableSQLLogging()

// 方式2：自定义日志级别
torm.SetSQLLogging(logger.DEBUG, true)  // DEBUG级别，启用
torm.SetSQLLogging(logger.INFO, true)   // INFO级别，启用
torm.SetSQLLogging(logger.WARN, false)  // 禁用SQL日志

// 方式3：通过配置启用
config := &torm.Config{
    Driver:     "mysql",
    Host:       "localhost",
    Database:   "myapp",
    LogQueries: true, // 启用查询日志
}
err := torm.AddConnection("main", config)
```

### SQL日志输出示例

启用SQL日志后，所有数据库操作都会被记录：

```go
// SELECT查询
users, err := torm.Table("users", "default").
    Where("age", ">", 18).
    OrderBy("created_at", "DESC").
    Limit(10).
    Get()

// 日志输出：
// 2024-01-20 10:30:15 [DEBUG] Query executed [sql SELECT * FROM users WHERE age > ? ORDER BY created_at DESC LIMIT 10 args [18] duration 3.2ms]

// INSERT操作
userID, err := torm.Table("users", "default").
    Insert(map[string]interface{}{
        "name":  "张三",
        "email": "zhangsan@example.com",
        "age":   25,
    })

// 日志输出：
// 2024-01-20 10:30:16 [DEBUG] Exec executed [sql INSERT INTO users (name, email, age) VALUES (?, ?, ?) args [张三 zhangsan@example.com 25] duration 5.1ms]

// UPDATE操作
affected, err := torm.Table("users", "default").
    Where("id", "=", userID).
    Update(map[string]interface{}{
        "age": 26,
    })

// 日志输出：
// 2024-01-20 10:30:17 [DEBUG] Exec executed [sql UPDATE users SET age = ? WHERE id = ? args [26 1] duration 2.8ms]

// DELETE操作
affected, err := torm.Table("users", "default").
    Where("id", "=", userID).
    Delete()

// 日志输出：
// 2024-01-20 10:30:18 [DEBUG] Exec executed [sql DELETE FROM users WHERE id = ? args [1] duration 1.9ms]
```

### 事务日志

事务操作也会被完整记录：

```go
err := torm.Transaction(func(tx torm.TransactionInterface) error {
    builder, _ := torm.Table("users", "default")
    builder.InTransaction(tx)
    
    _, err := builder.Insert(map[string]interface{}{
        "name": "事务用户",
        "age":  30,
    })
    return err
}, "default")

// 日志输出：
// 2024-01-20 10:30:19 [DEBUG] Transaction started [connection default]
// 2024-01-20 10:30:19 [DEBUG] Exec executed [sql INSERT INTO users (name, age) VALUES (?, ?) args [事务用户 30] duration 2.1ms]
// 2024-01-20 10:30:19 [DEBUG] Transaction committed [connection default duration 8.5ms]
```

### 错误日志

SQL执行错误会在ERROR级别记录：

```go
// 执行一个有问题的查询
_, err := torm.Table("nonexistent_table", "default").Get()

// 日志输出：
// 2024-01-20 10:30:20 [ERROR] Query failed [sql SELECT * FROM nonexistent_table args [] duration 1.2ms error Table 'myapp.nonexistent_table' doesn't exist]
```

## 📝 基础日志功能

### 独立的日志记录器

除了SQL日志，TORM还提供独立的日志记录器，用于应用程序日志：

```go
import "github.com/zhoudm1743/torm/logger"

// 创建基础日志记录器
appLogger := logger.NewLogger(logger.INFO)

// 记录不同级别的日志
appLogger.Debug("调试信息", "module", "user")
appLogger.Info("用户登录成功", "user_id", 12345)
appLogger.Warn("警告信息", "message", "连接池即将满载")
appLogger.Error("错误信息", "error", err.Error())
appLogger.Fatal("致命错误，程序退出") // 程序会自动退出

// 创建文件日志记录器
fileLogger, err := logger.NewFileLogger(logger.DEBUG, "app.log")
if err != nil {
    log.Fatal("文件日志创建失败:", err)
}

// 记录到文件
fileLogger.Info("应用启动", "version", "1.2.0")
```

### SQL日志记录器

创建专门的SQL日志记录器：

```go
// 创建SQL日志记录器
sqlLogger := logger.NewSQLLogger(logger.DEBUG, true)

// 手动设置为全局SQL日志记录器
torm.SetLogger(sqlLogger)

// 或者直接使用便捷函数
torm.EnableSQLLogging() // 相当于上面的操作
```

## 📊 日志级别控制

### 支持的日志级别

```go
import "github.com/zhoudm1743/torm/logger"

// 日志级别（从低到高）
logger.DEBUG  // 调试信息 - 包含SQL查询
logger.INFO   // 一般信息 - 包含连接信息
logger.WARN   // 警告信息 - 包含性能警告
logger.ERROR  // 错误信息 - 包含SQL执行错误
logger.FATAL  // 致命错误 - 程序退出

// 级别对比示例
torm.SetSQLLogging(logger.DEBUG, true)  // 显示所有SQL查询
torm.SetSQLLogging(logger.INFO, true)   // 显示连接信息和错误
torm.SetSQLLogging(logger.ERROR, true)  // 只显示SQL错误
```

### 动态调整日志级别

```go
// 开发环境：显示所有日志
if isDevelopment {
    torm.SetSQLLogging(logger.DEBUG, true)
}

// 测试环境：显示错误和警告
if isTesting {
    torm.SetSQLLogging(logger.WARN, true)
}

// 生产环境：只记录错误
if isProduction {
    torm.SetSQLLogging(logger.ERROR, true)
}

// 临时调试：动态开启
torm.EnableSQLLogging()
// ... 调试代码 ...
torm.SetSQLLogging(logger.ERROR, true) // 恢复到只记录错误
```

## ⚙️ 连接配置日志

### 在连接配置中启用日志

通过数据库配置启用日志是最常用的方式：

```go
// MySQL配置示例
mysqlConfig := &torm.Config{
    Driver:     "mysql",
    Host:       "localhost",
    Port:       3306,
    Username:   "root",
    Password:   "password",
    Database:   "myapp",
    Charset:    "utf8mb4",
    LogQueries: true, // 关键：启用查询日志
}
err := torm.AddConnection("mysql_main", mysqlConfig)

// PostgreSQL配置示例
pgConfig := &torm.Config{
    Driver:     "postgres",
    Host:       "localhost",
    Port:       5432,
    Username:   "postgres",
    Password:   "password",
    Database:   "myapp",
    SSLMode:    "disable",
    LogQueries: true, // 关键：启用查询日志
}
err = torm.AddConnection("pg_main", pgConfig)

// 设置全局日志记录器（必需）
torm.EnableSQLLogging()
```

### 多连接日志管理

```go
// 不同连接使用不同的日志策略
mainConfig := &torm.Config{
    Driver:     "mysql",
    Host:       "main-db",
    Database:   "production",
    LogQueries: false, // 生产库不记录详细日志
}

debugConfig := &torm.Config{
    Driver:     "mysql", 
    Host:       "debug-db",
    Database:   "debug",
    LogQueries: true, // 调试库记录详细日志
}

torm.AddConnection("main", mainConfig)
torm.AddConnection("debug", debugConfig)

// 根据需要启用SQL日志
torm.SetSQLLogging(logger.DEBUG, true)
```

## 💡 最佳实践

### 1. 性能考虑

```go
// 生产环境：只记录错误，避免性能影响
if isProduction {
    torm.SetSQLLogging(logger.ERROR, true)
    
    // 在配置中禁用详细日志
    config.LogQueries = false
}

// 开发环境：记录所有SQL，便于调试
if isDevelopment {
    torm.EnableSQLLogging()
    config.LogQueries = true
}
```

### 2. 调试技巧

```go
// 临时启用SQL日志进行调试
func debugSlowQuery() {
    // 保存当前日志级别
    torm.EnableSQLLogging()
    
    // 执行可能有问题的查询
    results, err := torm.Table("large_table", "default").
        Join("related_table", "large_table.id", "=", "related_table.large_id").
        Where("status", "=", "active").
        GroupBy("category").
        Having("COUNT(*)", ">", 100).
        OrderBy("created_at", "DESC").
        Get()
    
    // 查看SQL日志来分析性能问题
    log.Printf("查询结果数量: %d", len(results))
    
    // 恢复生产环境日志级别
    torm.SetSQLLogging(logger.ERROR, true)
}
```

### 3. 错误监控

```go
// 创建专门的错误日志文件
errorLogger, err := logger.NewFileLogger(logger.ERROR, "sql_errors.log")
if err != nil {
    log.Fatal("错误日志创建失败:", err)
}

// 设置为SQL错误日志记录器
torm.SetLogger(errorLogger)

// 现在所有SQL错误都会记录到 sql_errors.log 文件
```

### 4. 敏感信息处理

```go
// SQL日志会自动显示参数，注意敏感信息
// 好的做法：在记录前过滤敏感信息
user := map[string]interface{}{
    "username": "john_doe",
    "password": hashPassword("secret123"), // 已加密，安全
    "email":    "john@example.com",
}

// 不好的做法：直接传递明文密码
// user["password"] = "secret123" // 这会在日志中显示明文密码！
```

## 🔗 相关文档

- [查询构建器](Query-Builder) - SQL查询构建
- [数据库支持](Database-Support) - 多数据库连接
- [性能优化](Performance) - 日志性能优化技巧
- [故障排除](Troubleshooting) - 日志相关问题解决 