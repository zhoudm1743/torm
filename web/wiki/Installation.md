# 安装指南

TORM 的安装非常简单，只需几个命令即可完成。

## 📋 系统要求

- **Go 版本**: 1.19 或更高
- **操作系统**: Linux、macOS、Windows
- **数据库**: MySQL 5.7+、PostgreSQL 10+、SQLite 3.25+

## 🚀 快速安装

### 使用 go get

```bash
# 安装最新版本
go get github.com/zhoudm1743/torm

# 安装指定版本
go get github.com/zhoudm1743/torm@v1.2.12
```

### 使用 go mod

在 `go.mod` 文件中添加：

```go
require github.com/zhoudm1743/torm v1.2.12
```

然后运行：

```bash
go mod tidy
```

## 🔧 数据库驱动

TORM 已内置常用数据库驱动，通常无需额外安装：

| 数据库 | 驱动包 | 状态 |
|--------|--------|------|
| MySQL | github.com/go-sql-driver/mysql | ✅ 已内置 |
| PostgreSQL | github.com/lib/pq | ✅ 已内置 |
| SQLite | github.com/glebarez/sqlite | ✅ 已内置 |

## ✅ 验证安装

创建测试文件 `main.go`：

```go
package main

import (
    "fmt"
    "github.com/zhoudm1743/torm"
)

func main() {
    // 配置 SQLite 内存数据库测试
    err := torm.AddConnection("default", &torm.Config{
        Driver:   "sqlite",
        Database: ":memory:",
    })
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("TORM v%s 安装成功！\n", torm.GetVersion())
}
```

运行测试：

```bash
go run main.go
# 输出: TORM v1.2.12 安装成功！
```

## 🐳 Docker 环境

### Dockerfile 示例

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

FROM alpine:latest
COPY --from=builder /app/main /
CMD ["/main"]
```

### docker-compose 示例

```yaml
version: '3.8'

services:
  app:
    build: .
    depends_on:
      - mysql
    environment:
      - DB_HOST=mysql
      - DB_PORT=3306

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: myapp
    ports:
      - "3306:3306"
```

## 🔧 常见问题

### 网络问题

如果下载缓慢，配置 Go 代理：

```bash
# 使用国内代理
go env -w GOPROXY=https://goproxy.cn,direct
```

### 依赖冲突

清理并重新下载依赖：

```bash
go clean -modcache
go mod tidy
```

### CGO 问题

某些环境可能需要禁用 CGO：

```bash
CGO_ENABLED=0 go build
```

## 📚 下一步

安装完成后，查看以下文档开始使用：

- [快速开始](Quick-Start) - 5分钟上手 TORM
- [配置文档](Configuration) - 详细配置选项
- [查询构建器](Query-Builder) - SQL 查询构建
