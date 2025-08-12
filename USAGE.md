# TORM 使用指南

## 安装

```bash
go get -u github.com/zhoudm1743/torm@v1.1.4
```

## 使用方式

### 方式一：使用统一入口（推荐）

从 v1.1.4 开始，您可以通过统一入口使用所有功能：

```go
package main

import (
    "log"
    "github.com/zhoudm1743/torm"
)

type User struct {
    torm.BaseModel
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    // 配置数据库
    config := &torm.Config{
        Driver:   "mysql",
        Host:     "localhost",
        Port:     3306,
        Username: "root",
        Password: "password",
        Database: "test",
    }

    // 添加连接
    err := torm.AddConnection("default", config)
    if err != nil {
        log.Fatal(err)
    }

    // 创建查询
    query, err := torm.Query()
    if err != nil {
        log.Fatal(err)
    }

    // 使用表查询
    users, err := torm.Table("users").Get()
    if err != nil {
        log.Fatal(err)
    }
    
    // 执行事务
    err = torm.Transaction(func(tx torm.TransactionInterface) error {
        // 在事务中执行操作
        return nil
    })
}
```

### 方式二：分别导入子包

您也可以按需导入特定的子包：

```go
package main

import (
    "github.com/zhoudm1743/torm/db"
    "github.com/zhoudm1743/torm/model"
    "github.com/zhoudm1743/torm/cache"
)

type User struct {
    model.BaseModel
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    config := &db.Config{
        Driver:   "sqlite",
        Database: "test.db",
    }
    
    err := db.AddConnection("default", config)
    if err != nil {
        panic(err)
    }
    
    // 使用缓存
    cache := cache.NewMemoryCache()
    cache.Set("key", "value", 0)
}
```

## 支持的数据库

- MySQL
- PostgreSQL  
- SQLite
- MongoDB

## 主要功能

### 统一入口提供的功能

- `torm.AddConnection()` - 添加数据库连接
- `torm.DB()` - 获取数据库连接
- `torm.Query()` - 创建查询构建器
- `torm.Table()` - 创建表查询
- `torm.Transaction()` - 执行事务
- `torm.Raw()` - 执行原生 SQL
- `torm.Exec()` - 执行 SQL 语句

### 导出的类型

- `torm.Config` - 数据库配置
- `torm.BaseModel` - 基础模型
- `torm.ConnectionInterface` - 连接接口
- `torm.QueryInterface` - 查询接口
- `torm.TransactionInterface` - 事务接口
- 更多类型请查看 `torm.go` 文件

## 版本信息

```go
fmt.Println(torm.Version()) // 输出：1.1.4
```

## 迁移指南

如果您之前使用的是 v1.1.3 或更早版本，建议升级到 v1.1.4 并使用统一入口：

1. 更新版本：`go get -u github.com/zhoudm1743/torm@v1.1.4`
2. 将 `import "github.com/zhoudm1743/torm/db"` 改为 `import "github.com/zhoudm1743/torm"`
3. 使用 `torm.` 前缀调用函数

这样可以避免依赖包缺失的问题，并获得更好的开发体验。 