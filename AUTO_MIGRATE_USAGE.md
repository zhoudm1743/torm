# TORM AutoMigrate 功能使用指南

## 概述

TORM 的 `AutoMigrate()` 方法允许您根据模型结构体的标签自动创建和更新数据库表结构。此功能简化了数据库迁移过程，确保模型与数据库表结构保持同步。

## 功能特性

- ✅ **自动表检测**: 检查数据库中是否存在对应的表
- ✅ **结构体标签解析**: 从 Go 结构体标签推断列类型和约束
- ✅ **多数据库支持**: 支持 MySQL、PostgreSQL、SQLite 等
- ✅ **字段类型映射**: 自动映射 Go 类型到数据库列类型
- ✅ **配置检测**: 自动检测主键、时间戳、软删除等配置
- ⚠️ **表创建**: 当前版本提供框架，完整实现需进一步开发

## 支持的标签

### 基础标签

- `db:"column_name"` - 指定数据库列名
- `primaryKey:"true"` - 标记为主键
- `autoIncrement:"true"` - 自动递增（通常用于主键）
- `size:"100"` - 字符串类型的长度限制
- `unique:"true"` - 唯一约束
- `default:"value"` - 默认值
- `comment:"描述"` - 列注释

### 时间戳标签

- `autoCreateTime:"true"` - 创建时间（自动设置为当前时间）
- `autoUpdateTime:"true"` - 更新时间（自动更新为当前时间）

### 约束标签

- `not_null:"true"` - 非空约束
- `db:"-"` - 跳过该字段，不创建对应的数据库列

## 使用示例

### 1. 定义模型

```go
package main

import (
    "github.com/zhoudm1743/torm/db"
    "github.com/zhoudm1743/torm/model"
)

// Admin 管理员模型
type Admin struct {
    model.BaseModel
    ID        string   `json:"id" db:"id" primaryKey:"true" size:"36" comment:"管理员ID"`
    Phone     string   `json:"phone" db:"phone" size:"20" unique:"true" comment:"手机号"`
    Password  string   `json:"password" db:"password" size:"255" comment:"密码"`
    Nickname  string   `json:"nickname" db:"nickname" size:"50" comment:"昵称"`
    Avatar    string   `json:"avatar" db:"avatar" size:"255" comment:"头像"`
    Status    int      `json:"status" db:"status" default:"1" comment:"状态：1启用，0禁用"`
    Role      []string `json:"role" db:"role" comment:"角色列表"`
    CreatedAt int64    `json:"created_at" db:"created_at" autoCreateTime:"true" comment:"创建时间"`
    UpdatedAt int64    `json:"updated_at" db:"updated_at" autoUpdateTime:"true" comment:"更新时间"`
}

// NewAdmin 创建管理员实例
func NewAdmin() *Admin {
    admin := &Admin{BaseModel: *model.NewBaseModel()}
    admin.SetTable("admin")
    admin.SetPrimaryKey("id")
    admin.SetConnection("default")
    admin.DetectConfigFromStruct(admin)  // 重要：检测结构体配置
    return admin
}
```

### 2. 配置数据库连接

```go
func main() {
    // 配置数据库连接
    config := &db.Config{
        Driver:   "sqlite",
        Database: "./app.db",
    }

    // 添加连接配置
    if err := db.AddConnection("default", config); err != nil {
        log.Fatalf("数据库配置失败: %v", err)
    }

    // 创建模型实例
    admin := NewAdmin()

    // 执行自动迁移
    if err := admin.AutoMigrate(); err != nil {
        log.Printf("自动迁移失败: %v", err)
    } else {
        fmt.Println("表迁移成功！")
    }
}
```

## 类型映射

TORM 会根据 Go 类型自动映射到对应的数据库类型：

| Go 类型 | MySQL | PostgreSQL | SQLite |
|---------|-------|------------|--------|
| `string` | `VARCHAR(n)` | `VARCHAR(n)` | `TEXT` |
| `int`, `int32` | `INT` | `INTEGER` | `INTEGER` |
| `int64` | `BIGINT` | `BIGINT` | `INTEGER` |
| `int8` | `TINYINT` | `SMALLINT` | `INTEGER` |
| `int16` | `SMALLINT` | `SMALLINT` | `INTEGER` |
| `float32` | `FLOAT` | `REAL` | `REAL` |
| `float64` | `DOUBLE` | `DOUBLE PRECISION` | `REAL` |
| `bool` | `BOOLEAN` | `BOOLEAN` | `INTEGER` |
| `[]string` | `JSON` | `JSONB` | `TEXT` |
| `time.Time` | `DATETIME` | `TIMESTAMP` | `DATETIME` |

## 最佳实践

### 1. 始终调用 DetectConfigFromStruct

```go
admin.DetectConfigFromStruct(admin)  // 必须调用，用于解析结构体标签
```

### 2. 合理设置字段属性

```go
type User struct {
    model.BaseModel
    ID       int64  `db:"id" primaryKey:"true" autoIncrement:"true"`
    Email    string `db:"email" size:"100" unique:"true" not_null:"true"`
    Name     string `db:"name" size:"50"`
    Age      *int   `db:"age" comment:"年龄（可空）"` // 使用指针表示可空字段
}
```

### 3. 配合传统迁移使用

当前版本的 AutoMigrate 主要用于：
- 快速原型开发
- 配置验证
- 表结构检测

对于生产环境，建议配合传统的迁移文件使用：

```go
// 先尝试自动迁移检测配置
if err := model.AutoMigrate(); err != nil {
    log.Printf("AutoMigrate: %v", err)
}

// 使用传统迁移执行实际的表创建
migrator := migration.NewMigrator(conn, logger)
migrator.Up()
```

## 注意事项

1. **当前限制**: 完整的表创建功能正在开发中，当前版本主要提供配置检测和验证
2. **数据安全**: 在生产环境中使用时要小心，建议先在测试环境验证
3. **性能考虑**: AutoMigrate 会执行数据库查询检查表结构，不要在热路径中频繁调用
4. **兼容性**: 确保数据库驱动已正确安装和配置

## 错误处理

常见错误及解决方案：

```go
if err := admin.AutoMigrate(); err != nil {
    switch {
    case strings.Contains(err.Error(), "connection"):
        log.Println("数据库连接问题，请检查配置")
    case strings.Contains(err.Error(), "DetectConfigFromStruct"):
        log.Println("请先调用 DetectConfigFromStruct 方法")
    case strings.Contains(err.Error(), "table creation"):
        log.Println("表创建功能开发中，请使用传统迁移")
    default:
        log.Printf("未知错误: %v", err)
    }
}
```

## 开发状态

- ✅ **v1.1.6**: 基础 AutoMigrate 框架
- 🚧 **下一版本**: 完整的表创建和更新功能
- 🚧 **未来版本**: 索引管理、外键约束、数据迁移

---

> 📝 **提示**: 这是 TORM v1.1.6 的 AutoMigrate 功能介绍。随着版本更新，功能会逐步完善。最新信息请查看项目文档。
