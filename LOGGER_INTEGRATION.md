# TORM 日志系统集成指南

## 概述

TORM 现在完全集成了内置的 `@logger/` 包，替换了之前使用的 `fmt.Printf` 输出。所有的日志输出现在都使用统一的日志系统，支持不同的日志级别和结构化日志。

## 主要特性

### 🎯 统一的日志接口
- 所有组件使用同一个日志系统
- 支持结构化日志记录
- 默认日志级别为 `INFO`
- 支持动态调整日志级别

### 📊 日志级别
```go
logger.DEBUG   // 调试信息 - 开发环境使用
logger.INFO    // 一般信息 - 默认级别  
logger.WARN    // 警告信息
logger.ERROR   // 错误信息
logger.FATAL   // 致命错误
```

### 🔧 集成组件
- **BaseModel**: 每个模型实例都有独立的日志记录器
- **查询构造器**: SQL 查询和执行信息
- **数据库迁移**: 表结构变更日志
- **数据库管理器**: 连接和事务管理日志

## 使用方法

### 1. 基本使用

```go
// 创建模型时会自动集成日志系统
user := model.NewAutoMigrateModel(&User{})

// 获取模型的日志记录器
logger := user.GetLogger()

// 记录不同级别的日志
logger.Debug("调试信息", "user_id", 1)
logger.Info("操作信息", "action", "create")
logger.Warn("警告信息", "message", "性能较慢") 
logger.Error("错误信息", "error", err)
```

### 2. 设置日志级别

```go
// 为单个模型设置日志级别
user.SetLogLevel(logger.DEBUG)  // 开发环境
user.SetLogLevel(logger.INFO)   // 生产环境（默认）

// 全局设置日志级别
logger.SetDefaultLevel(logger.WARN)
```

### 3. 自定义日志记录器

```go
// 创建自定义日志记录器
customLogger := logger.NewLogger(logger.DEBUG)

// 应用到模型
user.SetLogger(customLogger)

// 文件日志记录器
fileLogger, err := logger.NewFileLogger(logger.INFO, "app.log")
if err == nil {
    user.SetLogger(fileLogger)
}
```

### 4. 结构化日志

```go
// 使用键值对记录结构化日志
user.GetLogger().Info("用户操作",
    "user_id", 123,
    "action", "login", 
    "ip", "192.168.1.1",
    "timestamp", time.Now())

// 输出格式: 
// 2024-01-15 10:30:45 [INFO] 用户操作 {user_id:123 action:login ip:192.168.1.1 timestamp:2024-01-15 10:30:45}
```

## 自动日志记录

### SQL 查询日志
```go
// 执行查询时自动记录 SQL 和执行时间
users, err := user.Where("age", ">", 18).Get()
// 自动输出: [INFO] 执行SQL语句 {sql:SELECT * FROM users WHERE age > ? args:[18]}
```

### 数据库迁移日志
```go
// AutoMigrate 时自动记录迁移信息
err := user.AutoMigrate()
// 自动输出: 
// [INFO] 表结构更新成功 {changes:3}
// [INFO] 表结构变更 {column:name action:修改 details:length changed from 50 to 100}
```

### 模型操作日志
```go
// 保存模型时的日志
err := user.Save()
// 如果开启 DEBUG 级别，会记录详细的操作信息
```

## 配置示例

### 开发环境配置
```go
func setupDevelopmentLogging() {
    // 启用详细的调试日志
    logger.SetDefaultLevel(logger.DEBUG)
    
    // 或者使用快速配置
    db.QuickEnableDebugLogging() // 包含 SQL 查询日志
}
```

### 生产环境配置
```go
func setupProductionLogging() {
    // 只记录重要信息
    logger.SetDefaultLevel(logger.INFO)
    
    // 使用文件日志
    err := db.SetupFileLogging("app.log", "info", true)
    if err != nil {
        log.Fatal("设置文件日志失败:", err)
    }
}
```

### 测试环境配置
```go
func setupTestLogging() {
    // 测试时只记录错误
    logger.SetDefaultLevel(logger.ERROR)
    
    // 或者完全禁用日志
    db.QuickDisableLogging()
}
```

## 性能考虑

### 日志级别控制
```go
// 只有当前日志级别允许时才会处理日志
// DEBUG 级别的日志在 INFO 级别下不会被处理，避免性能开销

if logger.GetLevel() <= logger.DEBUG {
    // 只在需要时才进行复杂的日志数据准备
    complexData := prepareComplexLogData()
    logger.Debug("详细调试信息", "data", complexData)
}
```

### 避免频繁的字符串格式化
```go
// 推荐: 使用结构化日志
logger.Info("用户登录", "user_id", userID, "ip", clientIP)

// 避免: 手动格式化字符串  
logger.Info(fmt.Sprintf("用户 %d 从 %s 登录", userID, clientIP))
```

## 迁移指南

如果你之前使用了 `fmt.Printf` 进行日志输出，现在可以这样迁移：

### 之前
```go
fmt.Printf("执行SQL: %s, 参数: %v\n", sql, args)
fmt.Printf("⚠️ 警告: %s\n", message)  
fmt.Printf("✅ 操作成功: %s\n", result)
```

### 现在
```go
logger.Info("执行SQL", "sql", sql, "args", args)
logger.Warn("警告", "message", message)
logger.Info("操作成功", "result", result)
```

## 最佳实践

1. **使用合适的日志级别**
   - `DEBUG`: 详细的调试信息，仅开发环境
   - `INFO`: 重要的业务操作，生产环境默认级别
   - `WARN`: 需要注意但不影响功能的问题
   - `ERROR`: 错误和异常情况

2. **使用结构化日志**
   - 使用键值对而不是格式化字符串
   - 保持键名一致性，便于日志分析

3. **避免敏感信息**
   - 不要记录密码、密钥等敏感信息
   - 对用户隐私数据进行脱敏处理

4. **合理的日志量**
   - 开发环境可以详细记录
   - 生产环境控制日志量，避免影响性能

## 总结

现在 TORM 拥有了完整的日志系统集成：

- ✅ **统一接口**: 所有组件使用同一日志系统
- ✅ **默认 INFO 级别**: 适合生产环境的默认配置  
- ✅ **结构化日志**: 便于分析和监控
- ✅ **性能优化**: 智能的日志级别控制
- ✅ **易于配置**: 简单的 API 和快速配置方法
- ✅ **中文支持**: 所有日志消息都是中文

通过这个集成，你可以更好地监控和调试你的 TORM 应用程序。
