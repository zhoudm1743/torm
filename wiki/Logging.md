# 日志系统

TORM 提供了完整的日志系统，支持多级别日志、SQL日志记录、结构化日志和灵活的日志配置。

## 📋 目录

- [基础日志](#基础日志)
- [SQL日志](#sql日志)
- [日志配置](#日志配置)
- [日志级别](#日志级别)
- [自定义日志](#自定义日志)

## 🚀 快速开始

```go
import "github.com/zhoudm1743/torm/pkg/logger"

// 创建日志记录器
appLogger := logger.NewLogger(logger.INFO)

// 记录不同级别的日志
appLogger.Debug("调试信息")
appLogger.Info("普通信息")
appLogger.Warn("警告信息")
appLogger.Error("错误信息")
appLogger.Fatal("致命错误")

// 结构化日志
appLogger.Info("用户登录", "user_id", 123, "ip", "192.168.1.1")
```

## 📝 基础日志

### 创建日志记录器

```go
// 基础日志记录器
logger := logger.NewLogger(logger.INFO)

// 带文件输出的日志记录器
fileLogger := logger.NewFileLogger(logger.DEBUG, "app.log")

// 同时输出到控制台和文件
multiLogger := logger.NewMultiLogger(logger.INFO, "app.log", true)
```

### 日志记录

```go
// 简单日志
logger.Info("应用启动")
logger.Error("数据库连接失败")

// 格式化日志
logger.Infof("用户 %s 登录成功", username)
logger.Errorf("查询失败: %v", err)

// 结构化日志
logger.InfoWithFields("订单创建", map[string]interface{}{
    "order_id": 12345,
    "user_id":  67890,
    "amount":   99.99,
})
```

## 🗄️ SQL日志

### 启用SQL日志

```go
// 创建SQL日志记录器
sqlLogger := logger.NewSQLLogger(logger.DEBUG, true)

// 设置为全局SQL日志记录器
db.SetSQLLogger(sqlLogger)
```

### SQL日志记录

```go
// 自动记录所有SQL查询
users, err := db.Table("users").Where("status", "=", "active").Get()
// 输出: [DEBUG] SQL: SELECT * FROM users WHERE status = ? | Bindings: [active] | Duration: 2.3ms

// 手动记录SQL
sqlLogger.LogQuery("SELECT * FROM users WHERE id = ?", []interface{}{1}, 1*time.Millisecond)
```

## ⚙️ 日志配置

### 配置选项

```go
config := &logger.Config{
    Level:      logger.INFO,
    Format:     logger.JSONFormat, // 或 logger.TextFormat
    Output:     "app.log",
    MaxSize:    100, // MB
    MaxBackups: 5,
    MaxAge:     30, // 天
    Compress:   true,
}

logger := logger.NewLoggerWithConfig(config)
```

### 环境配置

```go
// 开发环境
if isDevelopment {
    logger.SetLevel(logger.DEBUG)
    logger.SetFormat(logger.TextFormat)
}

// 生产环境
if isProduction {
    logger.SetLevel(logger.WARN)
    logger.SetFormat(logger.JSONFormat)
    logger.SetOutput("production.log")
}
```

## 📊 日志级别

```go
// 日志级别（从低到高）
logger.DEBUG   // 调试信息
logger.INFO    // 一般信息
logger.WARN    // 警告信息
logger.ERROR   // 错误信息
logger.FATAL   // 致命错误

// 设置日志级别
logger.SetLevel(logger.WARN) // 只记录WARN及以上级别
```

## 🔧 自定义日志

### 自定义格式化器

```go
type CustomFormatter struct{}

func (f *CustomFormatter) Format(entry *logger.Entry) ([]byte, error) {
    return []byte(fmt.Sprintf("[%s] %s: %s\n", 
        entry.Time.Format("2006-01-02 15:04:05"),
        strings.ToUpper(entry.Level.String()),
        entry.Message,
    )), nil
}

logger.SetFormatter(&CustomFormatter{})
```

### 自定义钩子

```go
type EmailHook struct {
    levels []logger.Level
}

func (h *EmailHook) Fire(entry *logger.Entry) error {
    if entry.Level == logger.ERROR || entry.Level == logger.FATAL {
        // 发送错误邮件通知
        return sendErrorEmail(entry.Message)
    }
    return nil
}

func (h *EmailHook) Levels() []logger.Level {
    return h.levels
}

// 添加钩子
logger.AddHook(&EmailHook{
    levels: []logger.Level{logger.ERROR, logger.FATAL},
})
```

## 📚 最佳实践

### 1. 结构化日志

```go
// 好的做法：使用结构化日志
logger.Info("用户操作", 
    "action", "login",
    "user_id", 12345,
    "ip", "192.168.1.1",
    "user_agent", "Mozilla/5.0...",
)

// 避免：字符串拼接
// logger.Info("用户 " + username + " 从 " + ip + " 登录")
```

### 2. 敏感信息处理

```go
// 好的做法：过滤敏感信息
logger.Info("用户认证",
    "user_id", userID,
    "email", maskEmail(email),
    "success", true,
)

func maskEmail(email string) string {
    parts := strings.Split(email, "@")
    if len(parts) != 2 {
        return "***"
    }
    return parts[0][:1] + "***@" + parts[1]
}
```

## 🔗 相关文档

- [配置](Configuration) - 日志配置选项
- [故障排除](Troubleshooting) - 日志相关问题
- [性能优化](Performance) - 日志性能优化 