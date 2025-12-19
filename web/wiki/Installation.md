# 安装指南

本文档提供了TORM的详细安装说明，包括各种环境下的安装方法和依赖配置。

## 📋 系统要求

### Go语言版本
- **最低要求**: Go 1.19
- **推荐版本**: Go 1.21 或更高
- **测试版本**: Go 1.19, 1.20, 1.21, 1.22

### 操作系统支持
- ✅ Linux (Ubuntu, CentOS, RHEL, Debian, etc.)
- ✅ macOS (Intel & Apple Silicon)
- ✅ Windows (Windows 10/11, Windows Server)
- ✅ FreeBSD
- ✅ Docker 容器环境

### 数据库支持
- ✅ MySQL 5.7+ / 8.0+
- ✅ PostgreSQL 11+ / 12+ / 13+ / 14+ / 15+
- ✅ SQLite 3.8+
- ✅ SQL Server 2017+ / Azure SQL
- ✅ MongoDB 4.4+ / 5.0+ / 6.0+ / 7.0+

## 🚀 快速安装

### 方法1: 使用 go get (推荐)

```bash
# 安装最新版本
go get github.com/zhoudm1743/torm

# 安装指定版本
go get github.com/zhoudm1743/torm@v1.3.0
```

### 方法2: 使用 go mod

在你的 `go.mod` 文件中添加：

```go
module your-project

go 1.19

require (
    github.com/zhoudm1743/torm v1.0.0
)
```

然后运行：

```bash
go mod tidy
```

### 方法3: 从源码安装

```bash
# 克隆仓库
git clone https://github.com/zhoudm1743/torm.git
cd torm

# 构建并安装
go build ./...
go install ./...
```

## 🔧 数据库驱动安装

TORM自动包含了常用的数据库驱动，但某些情况下你可能需要手动安装：

### MySQL
```bash
# TORM已包含，通常无需额外安装
go get github.com/go-sql-driver/mysql
```

### PostgreSQL
```bash
# TORM已包含，通常无需额外安装
go get github.com/lib/pq
```

### SQLite
```bash
# TORM已包含，通常无需额外安装
go get github.com/glebarez/sqlite
```

### SQL Server
```bash
# TORM已包含，通常无需额外安装
go get github.com/denisenkom/go-mssqldb
```

### MongoDB
```bash
# TORM已包含，通常无需额外安装
go get go.mongodb.org/mongo-driver/mongo
```

## 🐳 Docker环境安装

### 方法1: 在Dockerfile中安装

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# 安装TORM
RUN go get github.com/zhoudm1743/torm

COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

### 方法2: 使用多阶段构建

```dockerfile
# 构建阶段
FROM golang:1.21 AS builder

WORKDIR /app
COPY . .

# 下载依赖
RUN go mod tidy
RUN go mod download

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 运行阶段
FROM scratch
COPY --from=builder /app/main /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE 8080
CMD ["/main"]
```

### 方法3: 使用 docker-compose

```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - mysql
      - postgres
      - mongodb
    environment:
      - DB_HOST=mysql
      - POSTGRES_HOST=postgres
      - MONGO_HOST=mongodb

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: torm_test
    ports:
      - "3306:3306"

  postgres:
    image: postgres:15
    environment:
      POSTGRES_PASSWORD: password
      POSTGRES_DB: torm_test
    ports:
      - "5432:5432"

  mongodb:
    image: mongo:7.0
    ports:
      - "27017:27017"
```

## 🏗️ 环境配置

### 环境变量

TORM支持通过环境变量进行配置：

```bash
# 数据库配置
export TORM_DB_DRIVER=mysql
export TORM_DB_HOST=localhost
export TORM_DB_PORT=3306
export TORM_DB_NAME=myapp
export TORM_DB_USER=root
export TORM_DB_PASSWORD=password

# 连接池配置
export TORM_MAX_OPEN_CONNS=100
export TORM_MAX_IDLE_CONNS=10
export TORM_CONN_MAX_LIFETIME=3600

# 日志配置
export TORM_LOG_QUERIES=true
export TORM_LOG_LEVEL=info
```

### 配置文件

创建 `config.yaml` 文件：

```yaml
database:
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
    
  logging:
    queries: true
    level: info
    
  cache:
    enabled: true
    ttl: 300
```

## 📦 依赖包版本

TORM的依赖包版本要求：

```go
// go.mod 示例
module your-project

go 1.19

require (
    github.com/zhoudm1743/torm v1.0.0
)

// 间接依赖（自动管理）
require (
    github.com/go-sql-driver/mysql v1.7.1
    github.com/lib/pq v1.10.9
    github.com/glebarez/sqlite v1.9.0
    github.com/denisenkom/go-mssqldb v0.12.3
    go.mongodb.org/mongo-driver v1.17.4
    github.com/patrickmn/go-cache v2.1.0+incompatible
    github.com/sirupsen/logrus v1.9.3
    github.com/stretchr/testify v1.8.4
)
```

## 🔍 安装验证

### 基本验证

创建 `verify.go` 文件：

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/zhoudm1743/torm/db"
)

func main() {
    fmt.Println("TORM 安装验证")
    
    // 配置SQLite测试连接
    config := &db.Config{
        Driver:   "sqlite",
        Database: ":memory:",
    }
    
    err := db.AddConnection("test", config)
    if err != nil {
        log.Fatal("连接配置失败:", err)
    }
    
    conn, err := db.DB("test")
    if err != nil {
        log.Fatal("获取连接失败:", err)
    }
    
    ctx := context.Background()
    err = conn.Ping(ctx)
    if err != nil {
        log.Fatal("连接测试失败:", err)
    }
    
    fmt.Println("✅ TORM 安装成功!")
}
```

运行验证：

```bash
go run verify.go
```

预期输出：
```
TORM 安装验证
✅ TORM 安装成功!
```

### 完整功能验证

创建 `full_verify.go` 文件：

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/zhoudm1743/torm/db"
    "github.com/zhoudm1743/torm/migration"
)

func main() {
    fmt.Println("TORM 功能验证测试")
    
    // 1. 测试数据库连接
    testConnection()
    
    // 2. 测试查询构建器
    testQueryBuilder()
    
    // 3. 测试迁移工具
    testMigrations()
    
    // 4. 测试多数据库支持
    testMultiDatabase()
    
    fmt.Println("✅ 所有功能验证通过!")
}

func testConnection() {
    fmt.Println("1. 测试数据库连接...")
    
    config := &db.Config{
        Driver:   "sqlite",
        Database: ":memory:",
    }
    
    err := db.AddConnection("verify", config)
    if err != nil {
        log.Fatal("连接配置失败:", err)
    }
    
    conn, err := db.DB("verify")
    if err != nil {
        log.Fatal("获取连接失败:", err)
    }
    
    ctx := context.Background()
    err = conn.Ping(ctx)
    if err != nil {
        log.Fatal("连接测试失败:", err)
    }
    
    fmt.Println("  ✅ 数据库连接测试通过")
}

func testQueryBuilder() {
    fmt.Println("2. 测试查询功能...")
    
    conn, err := db.DB("verify")
    if err != nil {
        log.Fatal("获取连接失败:", err)
    }
    
    ctx := context.Background()
    
    // 创建测试表
    _, err = conn.Exec(ctx, `
        CREATE TABLE test_table (
            id INTEGER PRIMARY KEY,
            name TEXT,
            value INTEGER
        )
    `)
    if err != nil {
        log.Fatal("创建表失败:", err)
    }
    
    // 插入测试数据
    _, err = conn.Exec(ctx, 
        "INSERT INTO test_table (name, value) VALUES (?, ?)", 
        "test", 42)
    if err != nil {
        log.Fatal("插入数据失败:", err)
    }
    
    // 查询数据
    row := conn.QueryRow(ctx, "SELECT name, value FROM test_table WHERE id = ?", 1)
    var name string
    var value int
    err = row.Scan(&name, &value)
    if err != nil {
        log.Fatal("查询数据失败:", err)
    }
    
    if name != "test" || value != 42 {
        log.Fatal("查询结果不匹配")
    }
    
    fmt.Println("  ✅ 查询功能测试通过")
}

func testMigrations() {
    fmt.Println("3. 测试迁移工具...")
    
    conn, err := db.DB("verify")
    if err != nil {
        log.Fatal("获取连接失败:", err)
    }
    
    migrator := migration.NewMigrator(conn, nil)
    
    // 注册测试迁移
    migrator.RegisterFunc(
        "test_001",
        "测试迁移",
        func(ctx context.Context, conn db.ConnectionInterface) error {
            _, err := conn.Exec(ctx, `
                CREATE TABLE migration_test (
                    id INTEGER PRIMARY KEY,
                    data TEXT
                )
            `)
            return err
        },
        func(ctx context.Context, conn db.ConnectionInterface) error {
            _, err := conn.Exec(ctx, "DROP TABLE migration_test")
            return err
        },
    )
    
    ctx := context.Background()
    
    // 执行迁移
    err = migrator.Up(ctx)
    if err != nil {
        log.Fatal("执行迁移失败:", err)
    }
    
    // 验证迁移状态
    status, err := migrator.Status(ctx)
    if err != nil {
        log.Fatal("获取迁移状态失败:", err)
    }
    
    if len(status) != 1 || !status[0].Applied {
        log.Fatal("迁移状态不正确")
    }
    
    fmt.Println("  ✅ 迁移工具测试通过")
}

func testMultiDatabase() {
    fmt.Println("4. 测试多数据库支持...")
    
    // 添加第二个数据库连接
    config2 := &db.Config{
        Driver:   "sqlite",
        Database: ":memory:",
    }
    
    err := db.AddConnection("verify2", config2)
    if err != nil {
        log.Fatal("第二个连接配置失败:", err)
    }
    
    conn2, err := db.DB("verify2")
    if err != nil {
        log.Fatal("获取第二个连接失败:", err)
    }
    
    ctx := context.Background()
    err = conn2.Ping(ctx)
    if err != nil {
        log.Fatal("第二个连接测试失败:", err)
    }
    
    fmt.Println("  ✅ 多数据库支持测试通过")
}
```

运行完整验证：

```bash
go run full_verify.go
```

## 🔧 故障排除

### 常见安装问题

#### 1. Go版本过低
```
错误: TORM requires Go 1.19 or later
解决: 升级Go版本到1.19+
```

#### 2. 网络问题
```bash
# 配置代理
export GOPROXY=https://goproxy.cn,direct
export GOSUMDB=sum.golang.google.cn

# 或使用七牛云代理
export GOPROXY=https://goproxy.cn
```

#### 3. 权限问题
```bash
# Linux/macOS
sudo chown -R $USER:$USER $GOPATH

# 或使用用户目录
go env -w GOPATH=$HOME/go
```

#### 4. 依赖冲突
```bash
# 清理mod cache
go clean -modcache

# 重新下载依赖
go mod tidy
go mod download
```

### 数据库驱动问题

#### MySQL连接问题
```go
// 添加时区配置
config := &db.Config{
    Driver:   "mysql",
    Host:     "localhost",
    Port:     3306,
    Database: "test",
    Username: "root",
    Password: "password",
    Options: map[string]string{
        "parseTime": "true",
        "loc":       "Local",
        "charset":   "utf8mb4",
    },
}
```

#### PostgreSQL SSL问题
```go
config := &db.Config{
    Driver:   "postgres",
    Host:     "localhost",
    Port:     5432,
    Database: "test",
    Username: "postgres",
    Password: "password",
    SSLMode:  "disable", // 开发环境可以禁用SSL
}
```

#### MongoDB认证问题
```go
config := &db.Config{
    Driver:   "mongodb",
    Host:     "localhost",
    Port:     27017,
    Database: "test",
    Username: "admin",      // 如果需要认证
    Password: "password",   // 如果需要认证
    Options: map[string]string{
        "authSource": "admin", // 认证数据库
    },
}
```

### 编译问题

#### CGO相关错误
```bash
# 禁用CGO（某些驱动可能需要）
CGO_ENABLED=0 go build

# 或安装必要的C编译器
# Ubuntu/Debian
sudo apt-get install build-essential

# CentOS/RHEL
sudo yum groupinstall "Development Tools"

# macOS
xcode-select --install
```

#### 静态链接问题
```bash
# 静态编译
go build -ldflags '-extldflags "-static"' -a

# 或使用特定的构建标签
go build -tags netgo -ldflags '-w -extldflags "-static"'
```

## 📊 性能配置

### 生产环境推荐配置

```go
config := &db.Config{
    Driver:          "mysql",
    Host:            "localhost",
    Port:            3306,
    Database:        "production",
    Username:        "app_user",
    Password:        "secure_password",
    Charset:         "utf8mb4",
    MaxOpenConns:    100,              // 根据负载调整
    MaxIdleConns:    20,               // 通常是MaxOpenConns的20%
    ConnMaxLifetime: time.Hour,        // 1小时
    ConnMaxIdleTime: time.Minute * 30, // 30分钟
    LogQueries:      false,            // 生产环境关闭
}
```

### 开发环境推荐配置

```go
config := &db.Config{
    Driver:          "sqlite",
    Database:        "development.db",
    MaxOpenConns:    10,
    MaxIdleConns:    2,
    ConnMaxLifetime: time.Hour,
    LogQueries:      true,  // 开发环境开启
}
```

## 📝 安装日志

建议在安装过程中启用详细日志：

```bash
# 启用详细输出
go get -v github.com/zhoudm1743/torm

# 或查看模块信息
go list -m github.com/zhoudm1743/torm

# 查看依赖树
go mod graph | grep torm
```

## 🔄 升级指南

### 从旧版本升级

```bash
# 查看当前版本
go list -m github.com/zhoudm1743/torm

# 升级到最新版本
go get -u github.com/zhoudm1743/torm

# 或升级到指定版本
go get github.com/zhoudm1743/torm@v1.3.0

# 清理未使用的依赖
go mod tidy
```

### 版本兼容性

- **v1.0.x**: 稳定版本，向后兼容
- **v1.1.x**: 新功能版本，向后兼容
- **v2.x.x**: 主要版本，可能包含破坏性变更

## 📞 获取帮助

如果安装过程中遇到问题：

1. 查看 [故障排除文档](Troubleshooting)
2. 搜索 [GitHub Issues](https://github.com/zhoudm1743/torm/issues)
3. 提交新的 [Issue](https://github.com/zhoudm1743/torm/issues/new)
4. 发送邮件到 zhoudm1743@163.com

---

**🎉 安装完成后，请查看 [快速开始指南](Quick-Start) 来开始使用TORM！** 