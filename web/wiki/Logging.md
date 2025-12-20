# 日志系统

TORM 提供完整的日志系统，支持 SQL 查询日志、多级别日志和结构化输出。

## 🚀 快速开始

```go
import "github.com/zhoudm1743/torm"

// 启用 SQL 日志
torm.EnableSQLLogging()

// 执行查询，SQL 会自动记录
users, _ := torm.Table("users").
    Where("status", "=", "active").
    Get()

// 输出示例：
// 2024-01-20 10:30:15 [DEBUG] Query executed [sql SELECT * FROM users WHERE status = ? args [active] duration 2.5ms]
```

## 🗄️ SQL 查询日志

### 启用方式

```go
// 方式1：使用便捷函数
torm.EnableSQLLogging()

// 方式2：通过配置启用
config := &torm.Config{
    Driver:     "mysql",
    Host:       "localhost",
    Database:   "myapp",
    LogQueries: true, // 启用查询日志
}
torm.AddConnection("default", config)
```

### 日志输出示例

```go
// SELECT 查询
users, _ := torm.Table("users").
    Where("age", ">", 18).
    OrderBy("created_at", "DESC").
    Limit(10).
    Get()
// 日志: [DEBUG] Query executed [sql SELECT * FROM users WHERE age > ? ORDER BY created_at DESC LIMIT 10 args [18] duration 3.2ms]

// INSERT 操作
id, _ := torm.Table("users").Insert(map[string]interface{}{
    "name":  "张三",
    "email": "zhangsan@example.com",
})
// 日志: [DEBUG] Exec executed [sql INSERT INTO users (name, email) VALUES (?, ?) args [张三 zhangsan@example.com] duration 5.1ms]

// UPDATE 操作
affected, _ := torm.Table("users").
    Where("id", "=", 1).
    Update(map[string]interface{}{"age": 26})
// 日志: [DEBUG] Exec executed [sql UPDATE users SET age = ? WHERE id = ? args [26 1] duration 2.8ms]

// DELETE 操作
affected, _ := torm.Table("users").Where("id", "=", 1).Delete()
// 日志: [DEBUG] Exec executed [sql DELETE FROM users WHERE id = ? args [1] duration 1.9ms]
```

### 事务日志

```go
err := torm.Transaction(func(tx torm.TransactionInterface) error {
    _, err := tx.Exec("INSERT INTO users (name) VALUES (?)", "张三")
    return err
})

// 日志输出：
// [DEBUG] Transaction started [connection default]
// [DEBUG] Exec executed [sql INSERT INTO users (name) VALUES (?) args [张三] duration 2.1ms]
// [DEBUG] Transaction committed [connection default duration 8.5ms]
```

## 📊 日志级别

```go
import "github.com/zhoudm1743/torm/logger"

// 支持的日志级别（从低到高）
logger.DEBUG  // 调试信息 - 包含所有 SQL 查询
logger.INFO   // 一般信息 - 连接信息
logger.WARN   // 警告信息 - 性能警告
logger.ERROR  // 错误信息 - SQL 执行错误
logger.FATAL  // 致命错误 - 程序退出

// 设置日志级别
torm.SetSQLLogging(logger.DEBUG, true)  // 显示所有 SQL
torm.SetSQLLogging(logger.ERROR, true)  // 只显示错误
```

### 环境配置

```go
// 开发环境：显示所有日志
if isDevelopment {
    torm.EnableSQLLogging()
}

// 生产环境：只记录错误
if isProduction {
    torm.SetSQLLogging(logger.ERROR, true)
}
```

## 📝 独立日志记录器

```go
import "github.com/zhoudm1743/torm/logger"

// 创建控制台日志记录器
appLogger := logger.NewLogger(logger.INFO)
appLogger.Info("应用启动", "version", "1.2.12")
appLogger.Debug("调试信息", "module", "user")
appLogger.Warn("警告信息", "message", "连接池即将满载")
appLogger.Error("错误信息", "error", err.Error())

// 创建文件日志记录器
fileLogger, err := logger.NewFileLogger(logger.DEBUG, "app.log")
if err != nil {
    log.Fatal("文件日志创建失败:", err)
}
fileLogger.Info("应用启动", "version", "1.2.12")
```

## 💡 最佳实践

### 性能考虑

```go
// 生产环境：只记录错误，避免性能影响
if isProduction {
    torm.SetSQLLogging(logger.ERROR, true)
    config.LogQueries = false
}

// 开发环境：记录所有 SQL，便于调试
if isDevelopment {
    torm.EnableSQLLogging()
    config.LogQueries = true
}
```

### 调试技巧

```go
// 临时启用 SQL 日志进行调试
func debugSlowQuery() {
    torm.EnableSQLLogging()
    
    // 执行可能有问题的查询
    results, _ := torm.Table("large_table").
        Join("related_table", "large_table.id", "=", "related_table.large_id").
        Where("status", "=", "active").
        Get()
    
    // 查看 SQL 日志分析性能
    log.Printf("查询结果数量: %d", len(results))
    
    // 恢复生产环境日志级别
    torm.SetSQLLogging(logger.ERROR, true)
}
```

### 敏感信息处理

```go
// SQL 日志会显示参数，注意敏感信息
// 好的做法：在记录前加密敏感信息
user := map[string]interface{}{
    "username": "john_doe",
    "password": hashPassword("secret123"), // 已加密
    "email":    "john@example.com",
}

// 不好的做法：直接传递明文密码
// user["password"] = "secret123" // 会在日志中显示明文！
```

## 🔗 相关文档

- [配置文档](Configuration) - 日志配置选项
- [性能优化](Performance) - 日志性能优化
- [故障排除](Troubleshooting) - 日志相关问题
