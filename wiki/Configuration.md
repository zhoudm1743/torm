# 配置文档

TORM提供了灵活的配置选项，支持多种数据库和连接参数的配置。本文档详细介绍了所有可用的配置选项。

## 📋 基本配置

### 数据库连接配置

```go
config := &db.Config{
    Driver:          "mysql",          // 数据库驱动
    Host:            "localhost",      // 主机地址
    Port:            3306,             // 端口号
    Database:        "myapp",          // 数据库名
    Username:        "root",           // 用户名
    Password:        "password",       // 密码
    Charset:         "utf8mb4",        // 字符集
    
    // 连接池配置
    MaxOpenConns:    100,              // 最大打开连接数
    MaxIdleConns:    10,               // 最大空闲连接数
    ConnMaxLifetime: time.Hour,        // 连接最大生存时间
    ConnMaxIdleTime: time.Minute * 30, // 连接最大空闲时间
    
    // 日志配置
    LogQueries:      true,             // 是否记录查询日志
    
    // 额外选项
    Options: map[string]string{
        "parseTime": "true",
        "loc":       "Local",
    },
}
```

## 🗄️ 数据库特定配置

### MySQL 配置

```go
mysqlConfig := &db.Config{
    Driver:   "mysql",
    Host:     "localhost",
    Port:     3306,
    Database: "myapp",
    Username: "root",
    Password: "password",
    Charset:  "utf8mb4",
    
    // MySQL 特定选项
    Options: map[string]string{
        "parseTime":            "true",     // 解析时间
        "loc":                  "Local",    // 时区
        "charset":              "utf8mb4",  // 字符集
        "collation":            "utf8mb4_unicode_ci", // 排序规则
        "timeout":              "30s",      // 连接超时
        "readTimeout":          "30s",      // 读取超时
        "writeTimeout":         "30s",      // 写入超时
        "allowNativePasswords": "true",     // 允许原生密码
        "tls":                  "false",    // TLS连接
    },
}
```

### PostgreSQL 配置

```go
postgresConfig := &db.Config{
    Driver:   "postgres",
    Host:     "localhost",
    Port:     5432,
    Database: "myapp",
    Username: "postgres",
    Password: "password",
    SSLMode:  "disable",
    
    // PostgreSQL 特定选项
    Options: map[string]string{
        "sslmode":            "disable",     // SSL模式
        "connect_timeout":    "30",          // 连接超时
        "statement_timeout":  "30000",       // 语句超时（毫秒）
        "lock_timeout":       "30000",       // 锁超时（毫秒）
        "idle_in_transaction_session_timeout": "30000", // 事务空闲超时
        "search_path":        "public",      // 搜索路径
        "application_name":   "github.com/zhoudm1743/torm",        // 应用名称
    },
}
```

### SQLite 配置

```go
sqliteConfig := &db.Config{
    Driver:   "sqlite",
    Database: "./data/app.db",  // 数据库文件路径
    
    // SQLite 特定选项
    Options: map[string]string{
        "cache":        "shared",      // 缓存模式
        "mode":         "rwc",         // 打开模式
        "journal_mode": "WAL",         // 日志模式
        "synchronous":  "NORMAL",      // 同步模式
        "cache_size":   "1000",        // 缓存大小
        "temp_store":   "memory",      // 临时存储
        "foreign_keys": "on",          // 外键约束
    },
}
```

### MongoDB 配置

```go
mongoConfig := &db.Config{
    Driver:   "mongodb",
    Host:     "localhost",
    Port:     27017,
    Database: "myapp",
    Username: "",               // 可选
    Password: "",               // 可选
    
    // MongoDB 特定选项
    Options: map[string]string{
        "authSource":          "admin",       // 认证数据库
        "replicaSet":          "",            // 副本集名称
        "ssl":                 "false",       // SSL连接
        "connectTimeoutMS":    "30000",       // 连接超时
        "socketTimeoutMS":     "30000",       // Socket超时
        "serverSelectionTimeoutMS": "30000", // 服务器选择超时
        "maxPoolSize":         "100",         // 最大连接池大小
        "minPoolSize":         "5",           // 最小连接池大小
        "maxIdleTimeMS":       "300000",      // 最大空闲时间
        "retryWrites":         "true",        // 重试写入
        "w":                   "majority",    // 写关注
        "readPreference":      "primary",     // 读偏好
    },
}
```

### SQL Server 配置

```go
sqlserverConfig := &db.Config{
    Driver:   "sqlserver",
    Host:     "localhost",
    Port:     1433,
    Database: "myapp",
    Username: "sa",
    Password: "password",
    
    // SQL Server 特定选项
    Options: map[string]string{
        "connection timeout":  "30",          // 连接超时
        "dial timeout":        "15",          // 拨号超时
        "keepAlive":          "30",           // 保持连接
        "encrypt":            "disable",      // 加密
        "TrustServerCertificate": "true",     // 信任服务器证书
        "app name":           "github.com/zhoudm1743/torm",         // 应用名称
        "log":                "1",            // 日志级别
    },
}
```

## 🔧 连接池配置

### 详细连接池参数

```go
config := &db.Config{
    // 基本连接参数...
    
    // 连接池配置
    MaxOpenConns:    100,              // 最大打开连接数
    MaxIdleConns:    10,               // 最大空闲连接数
    ConnMaxLifetime: time.Hour,        // 连接最大生存时间
    ConnMaxIdleTime: time.Minute * 30, // 连接最大空闲时间
}
```

### 不同场景的连接池建议

#### 高并发Web应用
```go
config := &db.Config{
    MaxOpenConns:    200,              // 高并发需要更多连接
    MaxIdleConns:    50,               // 保持足够的空闲连接
    ConnMaxLifetime: time.Hour * 2,    // 较长的生存时间
    ConnMaxIdleTime: time.Minute * 15, // 适中的空闲时间
}
```

#### 批处理应用
```go
config := &db.Config{
    MaxOpenConns:    50,               // 批处理不需要太多连接
    MaxIdleConns:    10,               // 较少的空闲连接
    ConnMaxLifetime: time.Hour * 4,    // 长生存时间
    ConnMaxIdleTime: time.Hour,        // 长空闲时间
}
```

#### 微服务应用
```go
config := &db.Config{
    MaxOpenConns:    20,               // 微服务连接数适中
    MaxIdleConns:    5,                // 少量空闲连接
    ConnMaxLifetime: time.Minute * 30, // 短生存时间
    ConnMaxIdleTime: time.Minute * 10, // 短空闲时间
}
```

## 📝 日志配置

### 基本日志配置

```go
config := &db.Config{
    LogQueries: true,                  // 启用查询日志
}

// 创建连接时传入日志器
logger := logrus.New()
migrator := migration.NewMigrator(conn, logger)
```

### 自定义日志器

```go
// 实现 LoggerInterface
type CustomLogger struct {
    logger *logrus.Logger
}

func (l *CustomLogger) Debug(msg string, args ...interface{}) {
    l.logger.WithFields(toFields(args)).Debug(msg)
}

func (l *CustomLogger) Info(msg string, args ...interface{}) {
    l.logger.WithFields(toFields(args)).Info(msg)
}

func (l *CustomLogger) Error(msg string, args ...interface{}) {
    l.logger.WithFields(toFields(args)).Error(msg)
}

func toFields(args []interface{}) logrus.Fields {
    fields := logrus.Fields{}
    for i := 0; i < len(args)-1; i += 2 {
        if key, ok := args[i].(string); ok {
            fields[key] = args[i+1]
        }
    }
    return fields
}
```

## 🌍 环境变量配置

### 支持的环境变量

```bash
# 基本连接配置
export TORM_DB_DRIVER=mysql
export TORM_DB_HOST=localhost
export TORM_DB_PORT=3306
export TORM_DB_NAME=myapp
export TORM_DB_USER=root
export TORM_DB_PASSWORD=password
export TORM_DB_CHARSET=utf8mb4

# 连接池配置
export TORM_MAX_OPEN_CONNS=100
export TORM_MAX_IDLE_CONNS=10
export TORM_CONN_MAX_LIFETIME=3600
export TORM_CONN_MAX_IDLE_TIME=1800

# 日志配置
export TORM_LOG_QUERIES=true
export TORM_LOG_LEVEL=info

# SSL配置
export TORM_SSL_MODE=disable
export TORM_SSL_CERT=/path/to/cert.pem
export TORM_SSL_KEY=/path/to/key.pem
export TORM_SSL_CA=/path/to/ca.pem
```

### 从环境变量加载配置

```go
func LoadConfigFromEnv() *db.Config {
    config := &db.Config{
        Driver:   getEnv("TORM_DB_DRIVER", "mysql"),
        Host:     getEnv("TORM_DB_HOST", "localhost"),
        Port:     getEnvInt("TORM_DB_PORT", 3306),
        Database: getEnv("TORM_DB_NAME", ""),
        Username: getEnv("TORM_DB_USER", ""),
        Password: getEnv("TORM_DB_PASSWORD", ""),
        Charset:  getEnv("TORM_DB_CHARSET", "utf8mb4"),
        
        MaxOpenConns:    getEnvInt("TORM_MAX_OPEN_CONNS", 100),
        MaxIdleConns:    getEnvInt("TORM_MAX_IDLE_CONNS", 10),
        ConnMaxLifetime: time.Duration(getEnvInt("TORM_CONN_MAX_LIFETIME", 3600)) * time.Second,
        ConnMaxIdleTime: time.Duration(getEnvInt("TORM_CONN_MAX_IDLE_TIME", 1800)) * time.Second,
        
        LogQueries: getEnvBool("TORM_LOG_QUERIES", false),
        SSLMode:    getEnv("TORM_SSL_MODE", "disable"),
    }
    
    return config
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
    if value := os.Getenv(key); value != "" {
        if boolValue, err := strconv.ParseBool(value); err == nil {
            return boolValue
        }
    }
    return defaultValue
}
```

## 📁 配置文件支持

### YAML 配置文件

```yaml
# config.yaml
database:
  connections:
    default:
      driver: mysql
      host: localhost
      port: 3306
      database: myapp
      username: root
      password: password
      charset: utf8mb4
      
      pool:
        max_open_conns: 100
        max_idle_conns: 10
        conn_max_lifetime: 3600
        conn_max_idle_time: 1800
        
      options:
        parseTime: "true"
        loc: "Local"
        
    read_replica:
      driver: mysql
      host: read-replica.example.com
      port: 3306
      database: myapp
      username: readonly
      password: password
      charset: utf8mb4
      
      pool:
        max_open_conns: 50
        max_idle_conns: 5
        
  logging:
    queries: true
    level: info
    
  cache:
    enabled: true
    default_ttl: 300
    cleanup_interval: 600
```

### JSON 配置文件

```json
{
  "database": {
    "connections": {
      "default": {
        "driver": "mysql",
        "host": "localhost",
        "port": 3306,
        "database": "myapp",
        "username": "root",
        "password": "password",
        "charset": "utf8mb4",
        "pool": {
          "max_open_conns": 100,
          "max_idle_conns": 10,
          "conn_max_lifetime": 3600,
          "conn_max_idle_time": 1800
        },
        "options": {
          "parseTime": "true",
          "loc": "Local"
        }
      }
    },
    "logging": {
      "queries": true,
      "level": "info"
    }
  }
}
```

### 配置文件加载器

```go
func LoadConfigFromYAML(filename string) (*db.Config, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    
    var configData struct {
        Database struct {
            Connections map[string]struct {
                Driver   string `yaml:"driver"`
                Host     string `yaml:"host"`
                Port     int    `yaml:"port"`
                Database string `yaml:"database"`
                Username string `yaml:"username"`
                Password string `yaml:"password"`
                Charset  string `yaml:"charset"`
                
                Pool struct {
                    MaxOpenConns    int `yaml:"max_open_conns"`
                    MaxIdleConns    int `yaml:"max_idle_conns"`
                    ConnMaxLifetime int `yaml:"conn_max_lifetime"`
                    ConnMaxIdleTime int `yaml:"conn_max_idle_time"`
                } `yaml:"pool"`
                
                Options map[string]string `yaml:"options"`
            } `yaml:"connections"`
        } `yaml:"database"`
    }
    
    err = yaml.Unmarshal(data, &configData)
    if err != nil {
        return nil, err
    }
    
    defaultConn := configData.Database.Connections["default"]
    
    config := &db.Config{
        Driver:          defaultConn.Driver,
        Host:            defaultConn.Host,
        Port:            defaultConn.Port,
        Database:        defaultConn.Database,
        Username:        defaultConn.Username,
        Password:        defaultConn.Password,
        Charset:         defaultConn.Charset,
        MaxOpenConns:    defaultConn.Pool.MaxOpenConns,
        MaxIdleConns:    defaultConn.Pool.MaxIdleConns,
        ConnMaxLifetime: time.Duration(defaultConn.Pool.ConnMaxLifetime) * time.Second,
        ConnMaxIdleTime: time.Duration(defaultConn.Pool.ConnMaxIdleTime) * time.Second,
        Options:         defaultConn.Options,
    }
    
    return config, nil
}
```

## 🔒 安全配置

### SSL/TLS 配置

```go
// MySQL SSL配置
mysqlSSLConfig := &db.Config{
    Driver:   "mysql",
    Host:     "secure-mysql.example.com",
    Port:     3306,
    Database: "myapp",
    Username: "secure_user",
    Password: "secure_password",
    
    Options: map[string]string{
        "tls":           "custom",
        "tls-ca":        "/path/to/ca.pem",
        "tls-cert":      "/path/to/cert.pem",
        "tls-key":       "/path/to/key.pem",
        "tls-verify":    "true",
    },
}

// PostgreSQL SSL配置
postgresSSLConfig := &db.Config{
    Driver:   "postgres",
    Host:     "secure-postgres.example.com",
    Port:     5432,
    Database: "myapp",
    Username: "secure_user",
    Password: "secure_password",
    SSLMode:  "require",
    
    Options: map[string]string{
        "sslcert":     "/path/to/cert.pem",
        "sslkey":      "/path/to/key.pem",
        "sslrootcert": "/path/to/ca.pem",
    },
}
```

### 密码安全

```go
// 从环境变量读取敏感信息
config := &db.Config{
    Driver:   "mysql",
    Host:     os.Getenv("DB_HOST"),
    Port:     3306,
    Database: os.Getenv("DB_NAME"),
    Username: os.Getenv("DB_USER"),
    Password: os.Getenv("DB_PASSWORD"), // 敏感信息从环境变量读取
}

// 或从安全的配置管理服务读取
func LoadSecureConfig() *db.Config {
    // 从 HashiCorp Vault, AWS Secrets Manager 等读取
    password := getSecretFromVault("database/password")
    
    return &db.Config{
        // ... 其他配置
        Password: password,
    }
}
```

## 🎯 最佳实践

### 1. 分环境配置

```go
func GetConfigForEnvironment(env string) *db.Config {
    switch env {
    case "development":
        return &db.Config{
            Driver:       "sqlite",
            Database:     "dev.db",
            LogQueries:   true,
            MaxOpenConns: 10,
        }
    
    case "testing":
        return &db.Config{
            Driver:       "sqlite",
            Database:     ":memory:",
            LogQueries:   false,
            MaxOpenConns: 5,
        }
    
    case "production":
        return &db.Config{
            Driver:          "mysql",
            Host:            os.Getenv("DB_HOST"),
            Port:            3306,
            Database:        os.Getenv("DB_NAME"),
            Username:        os.Getenv("DB_USER"),
            Password:        os.Getenv("DB_PASSWORD"),
            MaxOpenConns:    200,
            MaxIdleConns:    50,
            ConnMaxLifetime: time.Hour * 2,
            LogQueries:      false,
        }
    
    default:
        panic("unknown environment: " + env)
    }
}
```

### 2. 配置验证

```go
func ValidateConfig(config *db.Config) error {
    if config.Driver == "" {
        return fmt.Errorf("driver is required")
    }
    
    if config.Driver != "sqlite" && config.Host == "" {
        return fmt.Errorf("host is required for driver %s", config.Driver)
    }
    
    if config.Database == "" {
        return fmt.Errorf("database is required")
    }
    
    if config.MaxOpenConns <= 0 {
        return fmt.Errorf("max_open_conns must be positive")
    }
    
    if config.MaxIdleConns > config.MaxOpenConns {
        return fmt.Errorf("max_idle_conns cannot be greater than max_open_conns")
    }
    
    return nil
}
```

### 3. 配置热加载

```go
type ConfigManager struct {
    config     *db.Config
    configFile string
    mutex      sync.RWMutex
}

func NewConfigManager(configFile string) *ConfigManager {
    return &ConfigManager{
        configFile: configFile,
    }
}

func (cm *ConfigManager) LoadConfig() error {
    config, err := LoadConfigFromYAML(cm.configFile)
    if err != nil {
        return err
    }
    
    cm.mutex.Lock()
    cm.config = config
    cm.mutex.Unlock()
    
    return nil
}

func (cm *ConfigManager) GetConfig() *db.Config {
    cm.mutex.RLock()
    defer cm.mutex.RUnlock()
    return cm.config
}

func (cm *ConfigManager) WatchConfig() {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    defer watcher.Close()
    
    err = watcher.Add(cm.configFile)
    if err != nil {
        log.Fatal(err)
    }
    
    for {
        select {
        case event := <-watcher.Events:
            if event.Op&fsnotify.Write == fsnotify.Write {
                log.Println("配置文件已修改，重新加载...")
                if err := cm.LoadConfig(); err != nil {
                    log.Printf("重新加载配置失败: %v", err)
                } else {
                    log.Println("配置重新加载成功")
                }
            }
        case err := <-watcher.Errors:
            log.Printf("配置文件监控错误: %v", err)
        }
    }
}
```

---

**📚 更多配置信息请参考 [API参考文档](API-Reference) 和 [最佳实践](Best-Practices)。** 