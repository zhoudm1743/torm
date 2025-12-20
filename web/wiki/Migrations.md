# 数据迁移

TORM 提供了两种数据库迁移方式：AutoMigrate 自动迁移和传统迁移系统。AutoMigrate 基于 TORM 标签实现零配置的表结构管理，传统迁移则提供更精细的版本控制。

## 📋 目录

- [快速开始](#快速开始)
- [AutoMigrate 自动迁移](#automigrate-自动迁移)
- [TORM 标签系统](#torm-标签系统)
- [传统迁移系统](#传统迁移系统)
- [跨数据库支持](#跨数据库支持)
- [最佳实践](#最佳实践)

## 🚀 快速开始

### AutoMigrate 快速上手

```go
package main

import (
    "time"
    "github.com/zhoudm1743/torm/db"
    "github.com/zhoudm1743/torm/model"
)

// 用户模型
type User struct {
    model.BaseModel
    ID        int       `torm:"primary_key,auto_increment"`
    Username  string    `torm:"type:varchar,size:50,unique,index"`
    Email     string    `torm:"type:varchar,size:100,unique"`
    Age       int       `torm:"type:int,unsigned,default:0"`
    Status    string    `torm:"type:varchar,size:20,default:active,index"`
    CreatedAt time.Time `torm:"auto_create_time"`
    UpdatedAt time.Time `torm:"auto_update_time"`
}

func main() {
    // 配置数据库
    db.AddConnection("default", &db.Config{
        Driver:   "mysql",
        Host:     "localhost",
        Database: "myapp",
        Username: "root",
        Password: "password",
    })
    
    // 自动创建表结构
    user := model.NewModel(&User{})
    user.SetConnection("default")
    user.AutoMigrate(&User{})
    
    // 完成！表已创建
}
```

## 🔄 AutoMigrate 自动迁移

### 基础用法

```go
// 单模型迁移
user := model.NewModel(&User{})
user.SetConnection("default")
err := user.AutoMigrate(&User{})
if err != nil {
    log.Fatalf("迁移失败: %v", err)
}

// 多模型迁移
err := user.AutoMigrate(&User{}, &Department{}, &Role{})
```

### 智能增量更新

AutoMigrate 会自动检测表结构差异，只执行必要的变更：

```go
// 第一版模型
type User struct {
    model.BaseModel
    ID   int    `torm:"primary_key,auto_increment"`
    Name string `torm:"type:varchar,size:50"`
}
// 执行: CREATE TABLE users (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(50))

// 第二版模型 - 添加字段
type User struct {
    model.BaseModel
    ID    int    `torm:"primary_key,auto_increment"`
    Name  string `torm:"type:varchar,size:50"`
    Email string `torm:"type:varchar,size:100,unique"` // 新增
    Age   int    `torm:"type:int,default:0"`           // 新增
}
// 执行: ALTER TABLE users ADD COLUMN email VARCHAR(100) UNIQUE
// 执行: ALTER TABLE users ADD COLUMN age INT DEFAULT 0

// 第三版模型 - 修改字段
type User struct {
    model.BaseModel
    ID    int    `torm:"primary_key,auto_increment"`
    Name  string `torm:"type:varchar,size:100"` // 长度从50改为100
    Email string `torm:"type:varchar,size:100,unique"`
    Age   int    `torm:"type:int,default:0"`
}
// 执行: ALTER TABLE users MODIFY COLUMN name VARCHAR(100)
```

### 支持的变更操作

- ✅ 添加新字段
- ✅ 修改字段类型
- ✅ 修改字段长度
- ✅ 修改数值精度
- ✅ 添加索引
- ✅ 添加唯一约束
- ✅ 添加外键
- ✅ 修改默认值

### 安全保护

- 🛡️ 不会删除现有字段
- 🛡️ 不会删除现有数据
- 🛡️ 变更失败时保持原始结构

## 🏷️ TORM 标签系统

### 标签语法

```go
type Example struct {
    model.BaseModel
    Field string `torm:"tag1,tag2:value,tag3"`
}
```

### 主键和自增

```go
type Model struct {
    // 自增主键
    ID int `torm:"primary_key,auto_increment"`
    
    // 字符串主键
    UUID string `torm:"primary_key,type:varchar,size:36"`
}
```

### 数据类型

```go
type TypeExamples struct {
    model.BaseModel
    
    // 字符串类型
    Name    string `torm:"type:varchar,size:100"`  // VARCHAR(100)
    Code    string `torm:"type:char,size:10"`      // CHAR(10)
    Bio     string `torm:"type:text"`              // TEXT
    Content string `torm:"type:longtext"`          // LONGTEXT
    
    // 数值类型
    Age      int     `torm:"type:int"`             // INT
    BigNum   int64   `torm:"type:bigint"`          // BIGINT
    SmallNum int16   `torm:"type:smallint"`        // SMALLINT
    TinyNum  int8    `torm:"type:tinyint"`         // TINYINT
    
    // 精度控制
    Price float64 `torm:"type:decimal,precision:10,scale:2"` // DECIMAL(10,2)
    Rate  float64 `torm:"type:decimal,precision:5,scale:4"`  // DECIMAL(5,4)
    
    // 布尔类型
    IsActive bool `torm:"type:boolean"`            // BOOLEAN
    
    // 日期时间
    BirthDate time.Time `torm:"type:date"`         // DATE
    LoginTime time.Time `torm:"type:datetime"`     // DATETIME
    EventTime time.Time `torm:"type:timestamp"`    // TIMESTAMP
    
    // 二进制和JSON
    Avatar   []byte                 `torm:"type:blob"` // BLOB
    Settings map[string]interface{} `torm:"type:json"` // JSON
}
```

### 约束和默认值

```go
type ConstraintExamples struct {
    model.BaseModel
    
    // 唯一约束
    Email    string `torm:"type:varchar,size:100,unique"`
    
    // 非空约束
    Name     string `torm:"type:varchar,size:100,not_null"`
    
    // 允许空值
    Phone    string `torm:"type:varchar,size:20,nullable"`
    
    // 默认值
    Status   string  `torm:"type:varchar,size:20,default:active"`
    Age      int     `torm:"type:int,default:0"`
    Balance  float64 `torm:"type:decimal,precision:10,scale:2,default:0.00"`
    IsActive bool    `torm:"type:boolean,default:1"`
    
    // 无符号数值
    Count    int     `torm:"type:int,unsigned"`
}
```

### 索引

```go
type IndexExamples struct {
    model.BaseModel
    
    // 普通索引
    Category string `torm:"type:varchar,size:50,index"`
    
    // 自定义索引名
    SearchKey string `torm:"type:varchar,size:100,index:search_idx"`
    
    // 唯一索引
    Email string `torm:"type:varchar,size:100,unique"`
    
    // 全文索引
    Title string `torm:"type:varchar,size:200,fulltext"`
    
    // 空间索引
    Location string `torm:"type:varchar,size:100,spatial"`
}
```

### 外键

```go
type Post struct {
    model.BaseModel
    ID     int `torm:"primary_key,auto_increment"`
    Title  string `torm:"type:varchar,size:200"`
    
    // 外键定义
    UserID int `torm:"type:int,references:users.id,on_delete:cascade,on_update:cascade"`
    
    // 可选外键（允许NULL）
    CategoryID int `torm:"type:int,references:categories.id,on_delete:set_null"`
}

// 支持的外键操作：
// on_delete: cascade, set_null, restrict, no_action
// on_update: cascade, set_null, restrict, no_action
```

### 自动时间戳

```go
type TimestampExamples struct {
    model.BaseModel
    
    // 自动创建时间（INSERT时自动设置）
    CreatedAt time.Time `torm:"auto_create_time"`
    
    // 自动更新时间（INSERT和UPDATE时自动设置）
    UpdatedAt time.Time `torm:"auto_update_time"`
}
```

### 字段注释

```go
type CommentExamples struct {
    model.BaseModel
    
    ID       int    `torm:"primary_key,auto_increment,comment:主键ID"`
    Username string `torm:"type:varchar,size:50,unique,comment:用户名"`
    Email    string `torm:"type:varchar,size:100,comment:邮箱地址"`
}
```

### 完整标签参考

| 分类 | 标签 | 语法 | 说明 |
|------|------|------|------|
| 主键 | `primary_key` | `torm:"primary_key"` | 设置为主键 |
| 自增 | `auto_increment` | `torm:"auto_increment"` | 自动递增 |
| 类型 | `type` | `torm:"type:varchar"` | 数据库类型 |
| 长度 | `size` | `torm:"size:100"` | 字段长度 |
| 精度 | `precision` | `torm:"precision:10"` | 数值精度 |
| 小数 | `scale` | `torm:"scale:2"` | 小数位数 |
| 唯一 | `unique` | `torm:"unique"` | 唯一约束 |
| 非空 | `not_null` | `torm:"not_null"` | 非空约束 |
| 可空 | `nullable` | `torm:"nullable"` | 允许空值 |
| 默认 | `default` | `torm:"default:0"` | 默认值 |
| 索引 | `index` | `torm:"index"` | 普通索引 |
| 全文 | `fulltext` | `torm:"fulltext"` | 全文索引 |
| 空间 | `spatial` | `torm:"spatial"` | 空间索引 |
| 外键 | `references` | `torm:"references:users.id"` | 外键引用 |
| 删除 | `on_delete` | `torm:"on_delete:cascade"` | 删除行为 |
| 更新 | `on_update` | `torm:"on_update:cascade"` | 更新行为 |
| 创建时间 | `auto_create_time` | `torm:"auto_create_time"` | 自动创建时间 |
| 更新时间 | `auto_update_time` | `torm:"auto_update_time"` | 自动更新时间 |
| 注释 | `comment` | `torm:"comment:描述"` | 字段注释 |
| 无符号 | `unsigned` | `torm:"unsigned"` | 无符号数值 |

## 📜 传统迁移系统

传统迁移系统提供版本控制和回滚能力。

### 创建迁移器

```go
import (
    "github.com/zhoudm1743/torm/db"
    "github.com/zhoudm1743/torm/migration"
    "github.com/zhoudm1743/torm/logger"
)

// 获取数据库连接
conn, err := db.DefaultManager().Connection("default")
if err != nil {
    log.Fatal(err)
}

// 创建迁移器
migrator := migration.NewMigrator(conn, logger.Default())
```

### 注册迁移

```go
// 方式1: 使用 RegisterFunc
migrator.RegisterFunc(
    "20240101_001",           // 版本号
    "创建用户表",              // 描述
    func(conn db.Connection) error {  // Up 函数
        _, err := conn.Exec(`
            CREATE TABLE users (
                id BIGINT AUTO_INCREMENT PRIMARY KEY,
                username VARCHAR(50) NOT NULL UNIQUE,
                email VARCHAR(100) NOT NULL UNIQUE,
                password VARCHAR(255) NOT NULL,
                status VARCHAR(20) DEFAULT 'active',
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
            )
        `)
        return err
    },
    func(conn db.Connection) error {  // Down 函数
        _, err := conn.Exec("DROP TABLE IF EXISTS users")
        return err
    },
)

// 方式2: 使用 Migration 对象
m := migration.NewMigration(
    "20240101_002",
    "创建部门表",
    func(conn db.Connection) error {
        _, err := conn.Exec(`
            CREATE TABLE departments (
                id BIGINT AUTO_INCREMENT PRIMARY KEY,
                name VARCHAR(100) NOT NULL UNIQUE,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
        `)
        return err
    },
    func(conn db.Connection) error {
        _, err := conn.Exec("DROP TABLE IF EXISTS departments")
        return err
    },
)
migrator.Register(m)
```

### 执行迁移

```go
// 执行所有待执行的迁移
err := migrator.Up()
if err != nil {
    log.Fatalf("迁移失败: %v", err)
}

// 回滚指定数量的迁移
err = migrator.Down(1) // 回滚最近1个迁移

// 重置所有迁移
err = migrator.Reset()

// 清空并重新执行所有迁移
err = migrator.Fresh()
```

### 查看迁移状态

```go
// 获取迁移状态
status, err := migrator.Status()
if err != nil {
    log.Fatal(err)
}

for _, s := range status {
    if s.Applied {
        fmt.Printf("✅ %s - %s (批次 %d)\n", s.Version, s.Description, s.Batch)
    } else {
        fmt.Printf("⏳ %s - %s\n", s.Version, s.Description)
    }
}

// 打印格式化状态
migrator.PrintStatus()
```

### 迁移示例

```go
func setupMigrations(migrator *migration.Migrator) {
    // 创建用户表
    migrator.RegisterFunc("20240101_001", "创建用户表",
        func(conn db.Connection) error {
            _, err := conn.Exec(`
                CREATE TABLE users (
                    id BIGINT AUTO_INCREMENT PRIMARY KEY,
                    username VARCHAR(50) NOT NULL UNIQUE,
                    email VARCHAR(100) NOT NULL UNIQUE,
                    password VARCHAR(255) NOT NULL,
                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                )
            `)
            return err
        },
        func(conn db.Connection) error {
            _, err := conn.Exec("DROP TABLE IF EXISTS users")
            return err
        },
    )
    
    // 添加用户状态字段
    migrator.RegisterFunc("20240101_002", "添加用户状态字段",
        func(conn db.Connection) error {
            _, err := conn.Exec(`
                ALTER TABLE users 
                ADD COLUMN status VARCHAR(20) DEFAULT 'active',
                ADD INDEX idx_status (status)
            `)
            return err
        },
        func(conn db.Connection) error {
            _, err := conn.Exec(`
                ALTER TABLE users 
                DROP INDEX idx_status,
                DROP COLUMN status
            `)
            return err
        },
    )
    
    // 创建部门表
    migrator.RegisterFunc("20240101_003", "创建部门表",
        func(conn db.Connection) error {
            _, err := conn.Exec(`
                CREATE TABLE departments (
                    id BIGINT AUTO_INCREMENT PRIMARY KEY,
                    name VARCHAR(100) NOT NULL UNIQUE,
                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                )
            `)
            return err
        },
        func(conn db.Connection) error {
            _, err := conn.Exec("DROP TABLE IF EXISTS departments")
            return err
        },
    )
    
    // 添加用户部门关联
    migrator.RegisterFunc("20240101_004", "添加用户部门关联",
        func(conn db.Connection) error {
            _, err := conn.Exec(`
                ALTER TABLE users 
                ADD COLUMN dept_id BIGINT,
                ADD CONSTRAINT fk_user_dept 
                    FOREIGN KEY (dept_id) REFERENCES departments(id) 
                    ON DELETE SET NULL
            `)
            return err
        },
        func(conn db.Connection) error {
            _, err := conn.Exec(`
                ALTER TABLE users 
                DROP FOREIGN KEY fk_user_dept,
                DROP COLUMN dept_id
            `)
            return err
        },
    )
}
```

## 🌐 跨数据库支持

### 类型映射

TORM 自动处理不同数据库的类型差异：

| Go类型 | TORM标签 | MySQL | PostgreSQL | SQLite |
|--------|----------|-------|------------|--------|
| `string` | `type:varchar,size:100` | `VARCHAR(100)` | `VARCHAR(100)` | `TEXT` |
| `int` | `type:int` | `INT` | `INTEGER` | `INTEGER` |
| `int64` | `auto_increment` | `AUTO_INCREMENT` | `SERIAL` | `AUTOINCREMENT` |
| `float64` | `type:decimal,precision:10,scale:2` | `DECIMAL(10,2)` | `DECIMAL(10,2)` | `REAL` |
| `bool` | `type:boolean` | `BOOLEAN` | `BOOLEAN` | `INTEGER` |
| `[]byte` | `type:blob` | `BLOB` | `BYTEA` | `BLOB` |
| `map[string]interface{}` | `type:json` | `JSON` | `JSONB` | `TEXT` |
| `time.Time` | `type:datetime` | `DATETIME` | `TIMESTAMP` | `DATETIME` |

### 多环境部署

```go
type User struct {
    model.BaseModel
    ID    int    `torm:"primary_key,auto_increment"`
    Name  string `torm:"type:varchar,size:100"`
    Email string `torm:"type:varchar,size:100,unique"`
}

func deployToEnvironments() {
    environments := map[string]*db.Config{
        "development": {Driver: "sqlite", Database: "dev.db"},
        "testing":     {Driver: "mysql", Host: "test.db.com", Database: "test"},
        "production":  {Driver: "postgres", Host: "prod.db.com", Database: "prod"},
    }
    
    for env, config := range environments {
        db.AddConnection(env, config)
        
        user := model.NewModel(&User{})
        user.SetConnection(env)
        user.AutoMigrate(&User{}) // 同一模型，适配不同数据库
        
        fmt.Printf("✅ %s 环境迁移完成\n", env)
    }
}
```

## 💡 最佳实践

### 1. 迁移顺序

```go
// ✅ 正确：先创建被引用的表
func migrate() {
    models := []interface{}{
        &Department{},  // 先创建（被 User 引用）
        &User{},        // 后创建（引用 Department）
        &Role{},
        &UserRole{},    // 最后创建（引用 User 和 Role）
    }
    
    for _, m := range models {
        model.NewModel(m).SetConnection("default").AutoMigrate(m)
    }
}
```

### 2. 版本号规范

```go
// ✅ 推荐：日期 + 序号
migrator.RegisterFunc("20240101_001", "创建用户表", ...)
migrator.RegisterFunc("20240101_002", "添加索引", ...)
migrator.RegisterFunc("20240102_001", "创建订单表", ...)
```

### 3. 错误处理

```go
// ✅ 完整的错误处理
func runMigrations() error {
    conn, err := db.DefaultManager().Connection("default")
    if err != nil {
        return fmt.Errorf("获取连接失败: %w", err)
    }
    
    migrator := migration.NewMigrator(conn, logger.Default())
    setupMigrations(migrator)
    
    if err := migrator.Up(); err != nil {
        return fmt.Errorf("迁移失败: %w", err)
    }
    
    return nil
}
```

### 4. 生产环境建议

```go
// ✅ 生产环境使用传统迁移
// - 可以回滚
// - 有版本记录
// - 支持数据迁移

// ⚠️ 开发环境可以使用 AutoMigrate
// - 快速迭代
// - 自动检测变更
// - 无需手写 SQL
```

### 5. 数据迁移

```go
// 传统迁移支持数据迁移
migrator.RegisterFunc("20240101_005", "迁移用户状态",
    func(conn db.Connection) error {
        // 数据迁移
        _, err := conn.Exec(`
            UPDATE users 
            SET status = 'active' 
            WHERE status IS NULL
        `)
        return err
    },
    func(conn db.Connection) error {
        // 回滚数据迁移
        _, err := conn.Exec(`
            UPDATE users 
            SET status = NULL 
            WHERE status = 'active'
        `)
        return err
    },
)
```

## 🔗 相关文档

- [模型系统](Model-System) - 模型定义和 TORM 标签
- [配置指南](Configuration) - 数据库连接配置
- [查询构建器](Query-Builder) - 数据库操作
- [事务处理](Transactions) - 迁移中的事务
