# TORM

欢迎使用TORM！

## 🚀 亮点

### ⚡ 零配置启动
- **一行代码**完成数据库连接和表创建
- **无需学习**复杂的迁移语法
- **开箱即用**的数据库表管理

### 🏷️ 强大的TORM标签系统
- **30+种标签**覆盖所有数据库特性
- **精确控制**字段类型、长度、约束
- **自动生成**索引、外键、注释

### 🔄 智能自动迁移
- **增量更新**自动检测表结构差异
- **数据安全**保护现有数据完整性
- **跨数据库**MySQL、PostgreSQL、SQLite无缝支持

### 🔗 现代化查询构建器
- **参数化查询**有效防止SQL注入
- **数组参数**自动展开为IN查询
- **跨数据库**占位符自动适配

### 💼 简化事务处理
- **自动管理**提交和回滚
- **异常安全**确保数据一致性
- **简洁API**专注业务逻辑

## 📚 快速导航

### 🎯 核心功能
- [**快速开始**](Quick-Start) - 5分钟上手TORM
- [**数据迁移**](Migrations) - 零配置的表结构管理
- [**查询构建器**](Query-Builder) - 强大的SQL构建工具
- [**模型系统**](Model-System) - 优雅的数据模型设计

### 💡 学习资源
- [**实例代码**](Examples) - 基于实际测试的完整示例
- [**关联关系**](Relationships) - 模型间的关联设计
- [**最佳实践**](Best-Practices) - 生产环境使用指南

## 🎯 30秒快速体验

```go
package main

import (
    "time"
    "github.com/zhoudm1743/torm"
)

// 定义用户模型
type User struct {
    torm.BaseModel
    ID        int       `torm:"primary_key,auto_increment"`
    Username  string    `torm:"type:varchar,size:50,unique,index"`
    Email     string    `torm:"type:varchar,size:100,unique"`
    Age       int       `torm:"type:int,unsigned,default:0"`
    Balance   float64   `torm:"type:decimal,precision:10,scale:2,default:0.00"`
    IsActive  bool      `torm:"type:boolean,default:1"`
    CreatedAt time.Time `torm:"auto_create_time"`
    UpdatedAt time.Time `torm:"auto_update_time"`
}

func main() {
    // 1. 配置数据库
    torm.AddConnection("default", &torm.Config{
        Driver:   "sqlite",
        Database: "demo.db",
    })
    
    // 2. 自动创建表结构
    user := &User{}
    user.AutoMigrate()
    
    // 3. 创建用户
    newUser := &User{
        Username: "zhangsan",
        Email:    "zhangsan@example.com",
        Age:      25,
        Balance:  1000.00,
        IsActive: true,
    }
    newUser.Save()
    
    // 4. 查询用户
    users, _ := torm.Table("users").
        Where("is_active = ? AND age >= ?", true, 18).
        Where("balance > ?", 500.00).
        OrderBy("created_at", "desc").
        Get()
    
    // 完成！无需配置，立即可用
}
```

## 🔧 支持的数据库

### 完全支持
- **MySQL** 5.7+ - 生产环境推荐
- **PostgreSQL** 10+ - 高级功能支持  
- **SQLite** 3.25+ - 开发测试推荐

### 自动适配特性
- **数据类型映射** - 自动转换Go类型到数据库类型
- **SQL方言适配** - 自动生成数据库特定的SQL
- **占位符处理** - MySQL(?), PostgreSQL($N), SQLite(?)
- **功能降级** - 自动处理数据库功能差异

## 🎭 使用场景

### 🚀 快速原型开发
```go
// 30秒搭建博客数据模型
type Post struct {
    torm.BaseModel
    ID       int    `torm:"primary_key,auto_increment"`
    Title    string `torm:"type:varchar,size:200"`
    Content  string `torm:"type:text"`
    AuthorID int    `torm:"type:int,references:users.id"`
    Status   string `torm:"type:varchar,size:20,default:draft"`
}

(&User{}).AutoMigrate()
(&Post{}).AutoMigrate()
```

### 🏢 企业级应用
```go
// 完整的权限系统
type User struct {
    torm.BaseModel
    ID         int     `torm:"primary_key,auto_increment"`
    Username   string  `torm:"type:varchar,size:50,unique,index"`
    Email      string  `torm:"type:varchar,size:100,unique"`
    Password   string  `torm:"type:varchar,size:255"`
    Salary     float64 `torm:"type:decimal,precision:10,scale:2"`
    DeptID     int     `torm:"type:int,references:departments.id,on_delete:set_null"`
    ManagerID  int     `torm:"type:int,references:users.id,on_delete:set_null"`
    IsActive   bool    `torm:"type:boolean,default:1"`
    
    CreatedAt  time.Time `torm:"auto_create_time"`
    UpdatedAt  time.Time `torm:"auto_update_time"`
}
```

### 🌐 微服务架构
```go
// 每个服务独立的数据模型
func setupOrderService() {
    torm.AddConnection("orders", orderConfig)
    
    models := []interface{}{
        &Order{}, &OrderItem{}, &Payment{},
    }
    
    for _, model := range models {
        model.(interface{ AutoMigrate() error }).AutoMigrate()
    }
}
```

### 🔄 多环境部署
```go
// 同一套代码，多环境部署
func deployToEnvironment(env string) {
    config := getConfigByEnv(env) // dev/test/prod
    torm.AddConnection("default", config)
    
    // 自动适配不同环境的数据库
    (&User{}).AutoMigrate()
    (&Product{}).AutoMigrate()
}
```

## 📊 性能表现

### 🚀 查询性能
- **零反射开销** - 直接SQL构建，避免运行时反射
- **连接池优化** - 高效的数据库连接管理
- **批量操作** - 原生支持批量插入和数组参数
- **索引自动化** - 根据TORM标签自动创建优化索引

### 💾 内存效率
- **轻量级设计** - 核心库体积小，依赖少
- **对象池** - 复用查询构建器对象
- **延迟加载** - 按需加载关联数据
- **GC友好** - 最小化内存分配

## 🛠️ 开发工具链

### 📝 代码生成
```bash
# 未来版本将支持
torm generate model User
torm generate migration create_users
torm validate schema
```

### 🔍 调试工具
```go
// 查看生成的SQL
sql, params := torm.Table("users").
    Where("status = ?", "active").
    ToSQL()
fmt.Printf("SQL: %s\nParams: %v\n", sql, params)
```

### 📊 性能分析
```go
// 查询性能监控
torm.EnableDebug()  // 显示执行时间
torm.EnableTrace()  // 显示完整调用栈
```

## 🌍 社区与生态

### 📚 学习资源
- **官方文档** - [torm.site](http://torm.site)
- **示例代码** - 基于实际项目的完整示例
- **视频教程** - 从入门到精通的系列教程
- **最佳实践** - 生产环境使用指南

### 🤝 社区支持
- **GitHub Issues** - 问题反馈和功能请求
- **讨论区** - 技术交流和经验分享
- **QQ群** - 实时答疑和讨论
- **微信群** - 官方技术支持

### 🔌 生态扩展
- **缓存集成** - Redis, Memcached支持
- **消息队列** - RabbitMQ, Kafka集成
- **监控集成** - Prometheus, Grafana支持
- **日志集成** - 结构化日志和链路追踪

## 🗺️ 版本路线图

### 🎯 v1.2.x (当前)
- ✅ 零配置自动迁移
- ✅ 30+种TORM标签
- ✅ 跨数据库兼容
- ✅ 参数化查询
- ✅ 数组参数支持

### 🚀 v1.3.0 (规划中)
- 🔄 关联关系预加载
- 🔄 软删除支持
- 🔄 模型事件钩子
- 🔄 JSON查询增强
- 🔄 分库分表支持

### 🌟 v1.4.0 (未来)
- 🔄 代码生成工具
- 🔄 图形化管理界面
- 🔄 性能监控面板
- 🔄 集群支持
- 🔄 云原生集成

## 🎉 立即开始

### 🚀 安装
```bash
go mod init your-project
go get github.com/zhoudm1743/torm
```

### 📚 学习路径
1. [**快速开始**](Quick-Start) - 5分钟体验核心功能
2. [**数据迁移**](Migrations) - 掌握表结构管理
3. [**查询构建器**](Query-Builder) - 学习高级查询技巧
4. [**模型系统**](Model-System) - 深入理解模型设计
5. [**实例代码**](Examples) - 通过实例加深理解

### 🎯 最佳实践
- 从小项目开始，逐步掌握TORM特性
- 充分利用TORM标签的强大功能
- 在开发环境使用AutoMigrate，生产环境谨慎使用
- 利用参数化查询确保安全性
- 根据业务需求选择合适的数据库

---

**🎊 开始你的TORM之旅！** TORM 让Go数据库开发变得简单而强大。

**📞 获取帮助**: [官方文档](http://torm.site) | [GitHub](https://github.com/zhoudm1743/torm) | [Issues](https://github.com/zhoudm1743/torm/issues)