# 配置文档

TORM 提供灵活的配置选项，支持多种数据库和连接参数。

## 📋 配置结构

```go
type Config struct {
    // 基础配置
    Driver   string // 数据库驱动: mysql, postgres, sqlite
    Host     string // 主机地址
    Port     int    // 端口
    Database string // 数据库名
    Username string // 用户名
    Password string // 密码
    Charset  string // 字符集
    Timezone string // 时区

    // 连接池配置
    MaxOpenConns    int           // 最大打开连接数
    MaxIdleConns    int           // 最大空闲连接数
    ConnMaxLifetime time.Duration // 连接最大生存时间
    ConnMaxIdleTime time.Duration // 连接最大空闲时间

    // 读写分离配置
    ReadHosts  []string // 读库主机列表
    WriteHosts []string // 写库主机列表

    // SSL配置
    SSLMode     string // SSL模式
    SSLCert     string // SSL证书文件
    SSLKey      string // SSL私钥文件
    SSLRootCert string // SSL根证书文件

    // 其他配置
    Prefix     string            // 表前缀
    Options    map[string]string // 其他选项
    Debug      bool              // 调试模式
    LogQueries bool              // 记录查询日志
}
```

## 🗄️ 数据库配置

### MySQL

```go
config := &torm.Config{
    Driver:   "mysql",
    Host:     "localhost",
    Port:     3306,
    Database: "myapp",
    Username: "root",
    Password: "password",
    Charset:  "utf8mb4",
    
    // 连接池
    MaxOpenConns:    100,
    MaxIdleConns:    10,
    ConnMaxLifetime: time.Hour,
    ConnMaxIdleTime: time.Minute * 30,
    
    // MySQL 特定选项
    Options: map[string]string{
        "parseTime":            "true",
        "loc":                  "Local",
        "timeout":              "30s",
        "readTimeout":          "30s",
        "writeTimeout":         "30s",
        "allowNativePasswords": "true",
    },
}

torm.AddConnection("default", config)
```

### PostgreSQL

```go
config := &torm.Config{
    Driver:   "postgres",
    Host:     "localhost",
    Port:     5432,
    Database: "myapp",
    Username: "postgres",
    Password: "password",
    SSLMode:  "disable",
    
    // 连接池
    MaxOpenConns:    100,
    MaxIdleConns:    10,
    ConnMaxLifetime: time.Hour,
    
    // PostgreSQL 特定选项
    Options: map[string]string{
        "connect_timeout":   "30",
        "statement_timeout": "30000",
        "search_path":       "public",
        "application_name":  "myapp",
    },
}

torm.AddConnection("default", config)
```

### SQLite

```go
config := &torm.Config{
    Driver:   "sqlite",
    Database: "./data/app.db",
    
    // SQLite 建议单连接
    MaxOpenConns: 1,
    MaxIdleConns: 1,
    
    // SQLite 特定选项
    Options: map[string]string{
        "cache":        "shared",
        "mode":         "rwc",
        "journal_mode": "WAL",
        "synchronous":  "NORMAL",
        "foreign_keys": "on",
    },
}

torm.AddConnection("default", config)
```

## 🔧 连接池配置

### 高并发 Web 应用

```go
config := &torm.Config{
    // ... 基础配置
    MaxOpenConns:    200,              // 高并发需要更多连接
    MaxIdleConns:    50,               // 保持足够的空闲连接
    ConnMaxLifetime: time.Hour * 2,    // 较长的生存时间
    ConnMaxIdleTime: time.Minute * 15, // 适中的空闲时间
}
```

### 批处理应用

```go
config := &torm.Config{
    // ... 基础配置
    MaxOpenConns:    50,           // 批处理不需要太多连接
    MaxIdleConns:    10,           // 较少的空闲连接
    ConnMaxLifetime: time.Hour * 4,
    ConnMaxIdleTime: time.Hour,
}
```

### 微服务应用

```go
config := &torm.Config{
    // ... 基础配置
    MaxOpenConns:    20,               // 微服务连接数适中
    MaxIdleConns:    5,                // 少量空闲连接
    ConnMaxLifetime: time.Minute * 30, // 短生存时间
    ConnMaxIdleTime: time.Minute * 10,
}
```

## 🔀 读写分离

```go
config := &torm.Config{
    Driver:   "mysql",
    Host:     "master.example.com", // 主库
    Port:     3306,
    Database: "myapp",
    Username: "root",
    Password: "password",
    
    // 读写分离配置
    WriteHosts: []string{"master.example.com"},
    ReadHosts:  []string{
        "slave1.example.com",
        "slave2.example.com",
    },
}

torm.AddConnection("default", config)

// 读操作自动路由到从库
users, _ := torm.Table("users").Get()

// 写操作自动路由到主库
torm.Table("users").Insert(data)

// 强制使用主库读取
masterConn, _ := torm.UseMasterDB()
// 使用 masterConn 执行查询
```

## 🔒 SSL/TLS 配置

### MySQL SSL

```go
config := &torm.Config{
    Driver:   "mysql",
    Host:     "secure-mysql.example.com",
    // ... 其他配置
    
    Options: map[string]string{
        "tls":        "custom",
        "tls-ca":     "/path/to/ca.pem",
        "tls-cert":   "/path/to/cert.pem",
        "tls-key":    "/path/to/key.pem",
        "tls-verify": "true",
    },
}
```

### PostgreSQL SSL

```go
config := &torm.Config{
    Driver:      "postgres",
    Host:        "secure-postgres.example.com",
    SSLMode:     "require", // disable, require, verify-ca, verify-full
    SSLCert:     "/path/to/cert.pem",
    SSLKey:      "/path/to/key.pem",
    SSLRootCert: "/path/to/ca.pem",
}
```

## 🌍 多连接管理

```go
// 添加多个连接
torm.AddConnection("mysql", mysqlConfig)
torm.AddConnection("postgres", postgresConfig)
torm.AddConnection("sqlite", sqliteConfig)

// 使用指定连接
users, _ := torm.Table("users", "mysql").Get()
posts, _ := torm.Table("posts", "postgres").Get()

// 获取连接实例
conn, _ := torm.DB("mysql")
```

## 📊 连接统计

```go
// 获取连接统计信息
stats := torm.GetConnectionStats()
for name, stat := range stats {
    fmt.Printf("连接: %s\n", name)
    fmt.Printf("  驱动: %s\n", stat.Driver)
    fmt.Printf("  健康: %v\n", stat.IsHealthy)
    fmt.Printf("  查询数: %d\n", stat.TotalQueries)
}

// 获取健康连接数
healthy, total := torm.GetHealthyConnections()
fmt.Printf("健康连接: %d/%d\n", healthy, total)
```

## 🔧 连接池管理

```go
// 设置连接池配置
torm.SetConnectionPoolConfig(
    100,                  // 最大连接数
    30*time.Second,       // 连接超时
    10*time.Minute,       // 空闲超时
    5*time.Minute,        // 清理间隔
)

// 预热连接
err := torm.WarmUpConnections()

// 关闭所有连接
err := torm.CloseAllConnections()
```

## 📝 日志配置

```go
// 启用 SQL 日志
torm.EnableSQLLogging()

// 设置日志级别
torm.SetSQLLogging(torm.DEBUG, true)

// 自定义日志器
customLogger := torm.NewLogger(torm.INFO)
torm.SetLogger(customLogger)
```

## 🌐 环境变量配置

```go
import "os"

config := &torm.Config{
    Driver:   os.Getenv("DB_DRIVER"),
    Host:     os.Getenv("DB_HOST"),
    Port:     getEnvInt("DB_PORT", 3306),
    Database: os.Getenv("DB_NAME"),
    Username: os.Getenv("DB_USER"),
    Password: os.Getenv("DB_PASSWORD"),
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}
```

## ✅ 配置验证

```go
config := &torm.Config{
    Driver:   "mysql",
    Host:     "localhost",
    Database: "myapp",
    Username: "root",
}

// 验证配置
if err := config.Validate(); err != nil {
    log.Fatalf("配置无效: %v", err)
}

// 带上下文的验证
if err := config.ValidateWithContext(); err != nil {
    log.Fatalf("配置无效: %v", err)
}
```

## 🎯 最佳实践

### 1. 分环境配置

```go
func GetConfig(env string) *torm.Config {
    switch env {
    case "development":
        return &torm.Config{
            Driver:     "sqlite",
            Database:   "dev.db",
            Debug:      true,
            LogQueries: true,
        }
    case "production":
        return &torm.Config{
            Driver:       "mysql",
            Host:         os.Getenv("DB_HOST"),
            Database:     os.Getenv("DB_NAME"),
            Username:     os.Getenv("DB_USER"),
            Password:     os.Getenv("DB_PASSWORD"),
            MaxOpenConns: 200,
            MaxIdleConns: 50,
            Debug:        false,
            LogQueries:   false,
        }
    }
    return nil
}
```

### 2. 敏感信息保护

```go
// 从环境变量读取敏感信息
config := &torm.Config{
    Driver:   "mysql",
    Host:     os.Getenv("DB_HOST"),
    Password: os.Getenv("DB_PASSWORD"), // 不要硬编码密码
}

// 日志中脱敏显示
fmt.Println(config.MaskPassword()) // 密码显示为 ****
```

### 3. 连接池优化

```go
// 根据应用类型调整连接池
config := &torm.Config{
    // Web 应用: 高并发，短连接
    MaxOpenConns:    100,
    MaxIdleConns:    25,
    ConnMaxLifetime: time.Hour,
    
    // 批处理: 低并发，长连接
    // MaxOpenConns:    20,
    // MaxIdleConns:    5,
    // ConnMaxLifetime: time.Hour * 4,
}
```
