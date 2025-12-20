# TORM

欢迎使用 TORM - 一个高性能的 Go ORM 框架，对 database/sql 进行深度封装。

## 🚀 核心特性

### ⚡ 多数据库支持
- **MySQL** - 完整支持，生产环境推荐
- **PostgreSQL** - 完整支持，包括 $N 占位符自动适配
- **SQLite** - 完整支持，开发测试推荐

### 🔗 链式查询构建器
- 流畅的 API 设计
- 参数化查询防止 SQL 注入
- 数组参数自动展开
- 跨数据库占位符自动适配

### 📦 完整模型系统
- 自动表名推断
- 时间戳自动管理
- 软删除支持
- 访问器/修改器
- 脏数据追踪

### 🔄 关联关系
- HasOne (一对一)
- HasMany (一对多)
- BelongsTo (反向关联)
- BelongsToMany (多对多)
- 预加载支持

### 💼 事务支持
- 自动事务管理
- 嵌套事务 (Savepoint)
- 手动事务控制

### 🗄️ 数据库迁移
- AutoMigrate 自动迁移
- 增量更新
- 跨数据库 DDL 生成

### ⚡ 高级特性
- 查询缓存
- 读写分离
- 连接池管理
- 上下文支持 (超时/取消)

## 📚 快速导航

### 入门指南
- [快速开始](Quick-Start) - 5分钟上手 TORM
- [安装指南](Installation) - 安装和配置
- [配置文档](Configuration) - 详细配置选项

### 核心功能
- [查询构建器](Query-Builder) - SQL 查询构建
- [模型系统](Model-System) - 数据模型设计
- [关联关系](Relationships) - 模型间关联
- [事务处理](Transactions) - 事务管理

### 高级功能
- [数据库迁移](Migrations) - 表结构管理
- [缓存系统](Caching) - 查询缓存
- [日志系统](Logging) - 查询日志

### 参考文档
- [API 参考](API-Reference) - 完整 API 文档
- [示例代码](Examples) - 实用示例
- [最佳实践](Best-Practices) - 使用建议

## 🎯 30秒快速体验

```go
package main

import (
    "fmt"
    "github.com/zhoudm1743/torm"
)

func main() {
    // 1. 配置数据库连接
    torm.AddConnection("default", &torm.Config{
        Driver:   "mysql",
        Host:     "localhost",
        Port:     3306,
        Database: "myapp",
        Username: "root",
        Password: "password",
        Charset:  "utf8mb4",
    })
    
    // 2. 查询数据
    users, _ := torm.Table("users").
        Where("status", "=", "active").
        OrderBy("created_at", "desc").
        Limit(10).
        Get()
    
    fmt.Printf("查询到 %d 个用户\n", len(users))
    
    // 3. 插入数据
    id, _ := torm.Table("users").Insert(map[string]interface{}{
        "name":   "张三",
        "email":  "zhangsan@example.com",
        "status": "active",
    })
    
    fmt.Printf("新用户 ID: %d\n", id)
}
```

## 🔧 支持的数据库

| 数据库 | 版本要求 | 状态 |
|--------|----------|------|
| MySQL | 5.7+ | ✅ 完整支持 |
| PostgreSQL | 10+ | ✅ 完整支持 |
| SQLite | 3.25+ | ✅ 完整支持 |

## 📊 版本信息

当前版本: **v1.2.12**

```go
import "github.com/zhoudm1743/torm"

fmt.Println(torm.GetVersion()) // 输出: 1.2.12
```

## 🔗 相关链接

- **GitHub**: [github.com/zhoudm1743/torm](https://github.com/zhoudm1743/torm)
- **官方文档**: [torm.site](http://torm.site)
- **问题反馈**: [GitHub Issues](https://github.com/zhoudm1743/torm/issues)
