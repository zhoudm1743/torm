# 数据迁移系统

TORM 提供了自动迁移系统，通过强大的TORM标签实现零配置的数据库表管理。同时保持对传统迁移的支持，为复杂场景提供完整解决方案。

## 🎯 核心优势

### AutoMigrate vs 传统迁移

| 特性 | AutoMigrate | 传统迁移 | 推荐场景 |
|------|-------------|----------|----------|
| **学习成本** | 零学习成本 | 需要了解SQL | 快速原型开发 |
| **代码维护** | 模型即文档 | 需要迁移文件 | 小团队项目 |
| **表结构同步** | 自动检测差异 | 手动编写变更 | 开发阶段 |
| **跨数据库** | 自动适配 | 需要分别编写 | 多环境部署 |
| **复杂变更** | 基础变更 | 完全支持 | 生产环境 |
| **数据迁移** | 不支持 | 完全支持 | 数据重构 |

## 🚀 快速开始

### AutoMigrate 零配置启动

```go
package main

import (
    "time"
    "github.com/zhoudm1743/torm"
)

// 用户模型 - 使用丰富的TORM标签
type User struct {
    torm.BaseModel
    ID        int       `json:"id" torm:"primary_key,auto_increment"`
    Username  string    `json:"username" torm:"type:varchar,size:50,unique,index"`
    Email     string    `json:"email" torm:"type:varchar,size:100,unique"`
    Password  string    `json:"password" torm:"type:varchar,size:255"`
    Age       int       `json:"age" torm:"type:int,unsigned,default:0"`
    Salary    float64   `json:"salary" torm:"type:decimal,precision:10,scale:2,default:0.00"`
    Status    string    `json:"status" torm:"type:varchar,size:20,default:active,index"`
    Bio       string    `json:"bio" torm:"type:text"`
    IsActive  bool      `json:"is_active" torm:"type:boolean,default:1"`
    DeptID    int       `json:"dept_id" torm:"type:int,references:departments.id,on_delete:set_null"`
    CreatedAt time.Time `json:"created_at" torm:"auto_create_time"`
    UpdatedAt time.Time `json:"updated_at" torm:"auto_update_time"`
}

// 部门模型
type Department struct {
    torm.BaseModel
    ID        int       `json:"id" torm:"primary_key,auto_increment"`
    Name      string    `json:"name" torm:"type:varchar,size:100,unique"`
    Budget    float64   `json:"budget" torm:"type:decimal,precision:12,scale:2,default:0.00"`
    Location  string    `json:"location" torm:"type:varchar,size:255"`
    IsActive  bool      `json:"is_active" torm:"type:boolean,default:1"`
    CreatedAt time.Time `json:"created_at" torm:"auto_create_time"`
    UpdatedAt time.Time `json:"updated_at" torm:"auto_update_time"`
}

func main() {
    // 1. 配置数据库
    torm.AddConnection("default", &torm.Config{
        Driver:   "mysql",
        Host:     "localhost",
        Database: "myapp",
        Username: "root",
        Password: "password",
    })
    
    // 2. 自动创建表结构（包括索引、外键、约束）
    dept := &Department{}
    dept.AutoMigrate()
    
    user := &User{}
    user.AutoMigrate()
    
    // 3. 完成！表已创建，包含所有字段、索引、外键约束
}
```

## 🏷️ TORM标签完整指南

### 基础标签语法

```go
type CompleteModel struct {
    torm.BaseModel
    
    // 主键和自增
    ID int64 `torm:"primary_key,auto_increment,comment:主键ID"`
    
    // 字符串类型控制
    Name     string `torm:"type:varchar,size:100,not_null,comment:姓名"`
    Username string `torm:"type:varchar,size:50,unique,index,comment:用户名"`
    Email    string `torm:"type:varchar,size:150,unique,comment:邮箱"`
    Bio      string `torm:"type:text,comment:个人简介"`
    Profile  string `torm:"type:longtext,comment:详细档案"`
    Code     string `torm:"type:char,size:10,comment:固定编码"`
    
    // 数值类型和精度
    Age      int     `torm:"type:int,unsigned,default:0,comment:年龄"`
    Score    int16   `torm:"type:smallint,default:0,comment:分数"`
    BigNum   int64   `torm:"type:bigint,comment:大数值"`
    Price    float64 `torm:"type:decimal,precision:10,scale:2,default:0.00,comment:价格"`
    Rate     float64 `torm:"type:decimal,precision:5,scale:4,comment:利率"`
    
    // 布尔类型
    IsActive  bool `torm:"type:boolean,default:1,comment:是否启用"`
    IsDeleted bool `torm:"type:boolean,default:0,comment:是否删除"`
    
    // 时间类型
    BirthDate time.Time `torm:"type:date,comment:出生日期"`
    LoginTime time.Time `torm:"type:datetime,comment:登录时间"`
    EventTime time.Time `torm:"type:timestamp,comment:事件时间"`
    
    // 二进制和JSON
    Avatar   []byte                 `torm:"type:blob,comment:头像数据"`
    Settings map[string]interface{} `torm:"type:json,comment:设置信息"`
    Tags     []string               `torm:"type:json,comment:标签列表"`
    
    // 索引类型
    Category   string `torm:"type:varchar,size:50,index,comment:分类"`
    SearchKey  string `torm:"type:varchar,size:200,fulltext,comment:搜索关键词"`
    Location   string `torm:"type:varchar,size:100,spatial,comment:地理位置"`
    CustomIdx  string `torm:"type:varchar,size:100,index:custom_name,comment:自定义索引"`
    
    // 外键和关联
    DepartmentID int `torm:"type:int,references:departments.id,on_delete:cascade,comment:部门ID"`
    ManagerID    int `torm:"type:int,references:users.id,on_delete:set_null,comment:管理员ID"`
    
    // 自动时间戳
    CreatedAt time.Time `torm:"auto_create_time,comment:创建时间"`
    UpdatedAt time.Time `torm:"auto_update_time,comment:更新时间"`
}
```

### 支持的标签类型

| 分类 | 标签 | 语法 | 说明 |
|------|------|------|------|
| **主键约束** | `primary_key` | `torm:"primary_key"` | 设置为主键 |
| | `auto_increment` | `torm:"auto_increment"` | 自动递增 |
| **数据类型** | `type` | `torm:"type:varchar"` | 指定数据库类型 |
| | `size` | `torm:"size:100"` | 字段长度 |
| | `precision` | `torm:"precision:10"` | 数值精度 |
| | `scale` | `torm:"scale:2"` | 小数位数 |
| **约束条件** | `unique` | `torm:"unique"` | 唯一约束 |
| | `not_null` | `torm:"not_null"` | 非空约束 |
| | `nullable` | `torm:"nullable"` | 允许空值 |
| **默认值** | `default` | `torm:"default:0"` | 设置默认值 |
| **索引类型** | `index` | `torm:"index"` | 普通索引 |
| | `index:名称` | `torm:"index:custom_idx"` | 自定义索引名 |
| | `fulltext` | `torm:"fulltext"` | 全文索引 |
| | `spatial` | `torm:"spatial"` | 空间索引 |
| **外键关系** | `references` | `torm:"references:users.id"` | 外键引用 |
| | `on_delete` | `torm:"on_delete:cascade"` | 删除时行为 |
| | `on_update` | `torm:"on_update:cascade"` | 更新时行为 |
| **时间戳** | `auto_create_time` | `torm:"auto_create_time"` | 自动创建时间 |
| | `auto_update_time` | `torm:"auto_update_time"` | 自动更新时间 |
| **其他** | `comment` | `torm:"comment:描述"` | 字段注释 |
| | `unsigned` | `torm:"unsigned"` | 无符号数值 |

### 跨数据库类型映射

TORM自动处理不同数据库的类型差异：

| Go类型 | TORM标签 | MySQL | PostgreSQL | SQLite |
|--------|----------|-------|------------|--------|
| `string` | `torm:"type:varchar,size:100"` | `VARCHAR(100)` | `VARCHAR(100)` | `TEXT` |
| `string` | `torm:"type:char,size:10"` | `CHAR(10)` | `CHAR(10)` | `TEXT` |
| `string` | `torm:"type:text"` | `TEXT` | `TEXT` | `TEXT` |
| `int` | `torm:"type:int"` | `INT` | `INTEGER` | `INTEGER` |
| `int64` | `torm:"type:bigint"` | `BIGINT` | `BIGINT` | `INTEGER` |
| `int64` | `torm:"auto_increment"` | `AUTO_INCREMENT` | `SERIAL` | `AUTOINCREMENT` |
| `float64` | `torm:"type:decimal,precision:10,scale:2"` | `DECIMAL(10,2)` | `DECIMAL(10,2)` | `REAL` |
| `bool` | `torm:"type:boolean"` | `BOOLEAN` | `BOOLEAN` | `INTEGER` |
| `[]byte` | `torm:"type:blob"` | `BLOB` | `BYTEA` | `BLOB` |
| `map[string]interface{}` | `torm:"type:json"` | `JSON` | `JSONB` | `TEXT` |
| `time.Time` | `torm:"type:datetime"` | `DATETIME` | `TIMESTAMP` | `DATETIME` |

## 🔄 智能差异检测

AutoMigrate会自动检测现有表结构与模型定义的差异：

### 增量更新示例

```go
// 第一次运行：创建完整表结构
type User struct {
    torm.BaseModel
    ID   int    `torm:"primary_key,auto_increment"`
    Name string `torm:"type:varchar,size:50"`
}
user := &User{}
user.AutoMigrate() // 创建表：users(id, name)

// 第二次运行：添加新字段
type User struct {
    torm.BaseModel
    ID    int    `torm:"primary_key,auto_increment"`
    Name  string `torm:"type:varchar,size:50"`
    Email string `torm:"type:varchar,size:100,unique"` // 新增字段
    Age   int    `torm:"type:int,default:0"`           // 新增字段
}
user.AutoMigrate() // 只添加新字段：ALTER TABLE users ADD COLUMN email, ADD COLUMN age

// 第三次运行：修改字段
type User struct {
    torm.BaseModel
    ID    int    `torm:"primary_key,auto_increment"`
    Name  string `torm:"type:varchar,size:100"`        // 长度从50改为100
    Email string `torm:"type:varchar,size:100,unique"`
    Age   int    `torm:"type:int,default:0"`
}
user.AutoMigrate() // 智能修改字段：ALTER TABLE users MODIFY COLUMN name VARCHAR(100)
```

### 支持的变更操作

- ✅ **添加新字段**: 自动ADD COLUMN
- ✅ **修改字段类型**: 自动MODIFY COLUMN
- ✅ **修改字段长度**: 自动调整VARCHAR长度
- ✅ **修改数值精度**: 自动调整DECIMAL精度和小数位
- ✅ **添加索引**: 自动CREATE INDEX
- ✅ **添加唯一约束**: 自动ADD UNIQUE INDEX
- ✅ **添加外键**: 自动ADD FOREIGN KEY
- ✅ **修改默认值**: 自动ALTER COLUMN DEFAULT

### 安全保护机制

- 🛡️ **数据保护**: 不会删除现有字段和数据
- 🛡️ **约束保护**: 修改约束时保护现有数据完整性
- 🛡️ **回滚支持**: 配合传统迁移实现复杂回滚
- 🛡️ **错误处理**: 变更失败时保持原始表结构

## 📊 实战应用场景

### 场景1：快速原型开发

```go
// 快速创建MVP产品的数据模型
type Product struct {
    torm.BaseModel
    ID          int     `torm:"primary_key,auto_increment"`
    Name        string  `torm:"type:varchar,size:200"`
    Price       float64 `torm:"type:decimal,precision:10,scale:2"`
    CategoryID  int     `torm:"type:int,references:categories.id"`
    CreatedAt   time.Time `torm:"auto_create_time"`
}

type Category struct {
    torm.BaseModel
    ID   int    `torm:"primary_key,auto_increment"`
    Name string `torm:"type:varchar,size:100,unique"`
}

// 一键部署数据库结构
func setupDatabase() {
    torm.AddConnection("default", config)
    
    // 顺序很重要：先创建被引用的表
    (&Category{}).AutoMigrate()
    (&Product{}).AutoMigrate()
}
```

### 场景2：多环境数据库

```go
type User struct {
    torm.BaseModel
    ID    int    `torm:"primary_key,auto_increment"`
    Name  string `torm:"type:varchar,size:100"`
    Email string `torm:"type:varchar,size:100,unique"`
}

func deployToEnvironments() {
    environments := map[string]*torm.Config{
        "development": {Driver: "sqlite", Database: "dev.db"},
        "testing":     {Driver: "mysql", Host: "test.db.com", Database: "test"},
        "production":  {Driver: "postgres", Host: "prod.db.com", Database: "prod"},
    }
    
    for env, config := range environments {
        torm.AddConnection(env, config)
        
        user := &User{}
        user.SetConnection(env)
        user.AutoMigrate() // 同一模型，适配不同数据库
    }
}
```

### 场景3：渐进式迁移策略

```go
// 第一阶段：使用AutoMigrate快速建立基础结构
func phase1_AutoMigrate() {
    models := []interface{}{
        &User{}, &Product{}, &Order{},
    }
    
    for _, model := range models {
        model.(interface{ AutoMigrate() error }).AutoMigrate()
    }
}

// 第二阶段：使用传统迁移处理复杂变更
func phase2_ComplexMigrations() {
    migrator := migration.NewMigrator(conn, logger)
    
    // 数据迁移
    migrator.RegisterFunc("20240101_001", "迁移历史数据", 
        func(conn db.ConnectionInterface) error {
            // 复杂的数据转换逻辑
            return migrateHistoricalData(conn)
        },
        func(conn db.ConnectionInterface) error {
            return rollbackHistoricalData(conn)
        })
    
    // 性能优化
    migrator.RegisterFunc("20240101_002", "添加复合索引",
        func(conn db.ConnectionInterface) error {
            _, err := conn.Exec("CREATE INDEX idx_user_status_created ON users(status, created_at)")
        return err
        },
        func(conn db.ConnectionInterface) error {
            _, err := conn.Exec("DROP INDEX idx_user_status_created")
        return err
        })
    
    migrator.Up()
}
```

## 📈 性能优化建议

### 1. 索引策略

```go
type OptimizedUser struct {
    torm.BaseModel
    ID       int    `torm:"primary_key,auto_increment"`
    
    // 频繁查询的字段添加索引
    Email    string `torm:"type:varchar,size:100,unique,index"`
    Status   string `torm:"type:varchar,size:20,index"`
    DeptID   int    `torm:"type:int,index"`
    
    // 全文搜索字段
    Bio      string `torm:"type:text,fulltext"`
    
    // 复合索引通过传统迁移添加
    CreatedAt time.Time `torm:"auto_create_time"`
    UpdatedAt time.Time `torm:"auto_update_time"`
}
```

### 2. 数据类型优化

```go
type EfficientModel struct {
    torm.BaseModel
    
    // 选择合适的数值类型
    TinyFlag  int8    `torm:"type:tinyint"`     // 1字节，范围-128到127
    SmallNum  int16   `torm:"type:smallint"`    // 2字节，范围-32768到32767  
    NormalNum int32   `torm:"type:int"`         // 4字节，范围约±21亿
    BigNum    int64   `torm:"type:bigint"`      // 8字节，大数值
    
    // 精确控制字符串长度
    Code      string  `torm:"type:char,size:10"`      // 固定长度，性能更好
    ShortText string  `torm:"type:varchar,size:50"`   // 短文本
    LongText  string  `torm:"type:text"`              // 长文本
    
    // 精确控制小数精度
    Price     float64 `torm:"type:decimal,precision:10,scale:2"` // 总位数10，小数位2
    Rate      float64 `torm:"type:decimal,precision:5,scale:4"`  // 总位数5，小数位4
}
```

### 3. 批量操作

```go
func batchAutoMigrate() {
    // 批量迁移多个相关模型
    models := []interface{}{
        &Department{},  // 先创建被引用的表
        &User{},       // 后创建引用外键的表
        &Product{},
        &Order{},
        &OrderItem{},
    }
    
    for _, model := range models {
        if err := model.(interface{ AutoMigrate() error }).AutoMigrate(); err != nil {
            log.Printf("AutoMigrate failed for %T: %v", model, err)
        }
    }
}
```

## 🔗 最佳实践

### 1. 模型设计原则

```go
// ✅ 好的设计
type User struct {
    torm.BaseModel
    
    // 主键明确
    ID int64 `torm:"primary_key,auto_increment,comment:用户ID"`
    
    // 业务字段有意义的约束
    Username string `torm:"type:varchar,size:50,unique,index,comment:用户名"`
    Email    string `torm:"type:varchar,size:100,unique,comment:邮箱地址"`
    
    // 合适的数据类型
    Age      int    `torm:"type:int,unsigned,default:0,comment:年龄"`
    Balance  float64 `torm:"type:decimal,precision:10,scale:2,default:0.00,comment:余额"`
    
    // 状态字段有默认值
    Status   string `torm:"type:varchar,size:20,default:active,index,comment:状态"`
    IsActive bool   `torm:"type:boolean,default:1,comment:是否启用"`
    
    // 自动时间戳
    CreatedAt time.Time `torm:"auto_create_time,comment:创建时间"`
    UpdatedAt time.Time `torm:"auto_update_time,comment:更新时间"`
}
```

### 2. 环境配置策略

```go
func setupEnvironment() {
    env := os.Getenv("APP_ENV")
    
    var config *torm.Config
    switch env {
    case "development":
        config = &torm.Config{
            Driver: "sqlite",
            Database: "dev.db",
        }
    case "testing":
        config = &torm.Config{
            Driver: "mysql",
            Host: "localhost",
            Database: "test_db",
        }
    case "production":
        config = &torm.Config{
            Driver: "postgres",
            Host: os.Getenv("DB_HOST"),
            Database: os.Getenv("DB_NAME"),
        }
    }
    
    torm.AddConnection("default", config)
}
```

### 3. 错误处理

```go
func safeAutoMigrate(models ...interface{}) error {
    for _, model := range models {
        if migrator, ok := model.(interface{ AutoMigrate() error }); ok {
            if err := migrator.AutoMigrate(); err != nil {
                return fmt.Errorf("AutoMigrate failed for %T: %w", model, err)
            }
            log.Printf("✅ AutoMigrate success for %T", model)
        } else {
            log.Printf("⚠️  Model %T does not support AutoMigrate", model)
        }
    }
    return nil
}
```