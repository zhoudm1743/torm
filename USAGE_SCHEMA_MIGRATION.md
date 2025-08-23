# TORM 表结构自动对比和更新 - 使用指南

## 🎉 功能完成

TORM v1.1.6 现在支持真正的表结构自动对比和更新功能！当您修改模型的 `torm` 标签后，`AutoMigrate()` 方法会：

1. 🔍 **自动检测变更** - 对比模型定义与数据库当前结构
2. 🔧 **生成 ALTER 语句** - 智能生成跨数据库兼容的SQL
3. 💾 **安全执行变更** - 支持事务回滚和表备份
4. 📊 **详细变更报告** - 显示每个字段的具体变更内容

## 🚀 使用方法

### 方法一：使用 NewAutoMigrateModel (推荐)

```go
package models

import (
    "reflect"
    "github.com/zhoudm1743/torm/model"
)

type User struct {
    model.BaseModel
    ID    int64  `torm:"primary_key,auto_increment,comment:用户ID"`
    Name  string `torm:"type:varchar,size:100,comment:用户名"`
    Email string `torm:"type:varchar,size:150,unique,comment:邮箱"`
    Age   int    `torm:"type:int,default:0,comment:年龄"`
}

func NewUser() *User {
    user := &User{}
    // 🎯 使用新的 NewAutoMigrateModel，完全支持 AutoMigrate
    user.BaseModel = *model.NewAutoMigrateModel(user)
    user.SetTable("users")
    user.SetConnection("default")
    return user
}

// 使用
func main() {
    user := NewUser()
    
    // ✨ 一行代码完成表结构创建/更新
    err := user.AutoMigrate()
    if err != nil {
        log.Fatal(err)
    }
}
```

### 方法二：手动设置模型结构

```go
func NewUser() *User {
    user := &User{}
    user.BaseModel = *model.NewBaseModel()
    user.SetTable("users")
    user.SetConnection("default")
    
    // 🔧 手动设置模型结构类型
    user.SetModelStruct(reflect.TypeOf(*user))
    
    return user
}
```

### 方法三：传统方式（兼容性保持）

```go
func NewUser() *User {
    user := &User{}
    user.BaseModel = *model.NewBaseModelWithAutoDetect(user)
    user.SetTable("users")
    user.SetConnection("default")
    return user
}
```

## 📊 实际效果演示

### 场景：修改字段长度和添加新字段

```go
// 原始模型
type Product struct {
    model.BaseModel
    Name  string `torm:"type:varchar,size:50,comment:产品名称"`
    Price float64 `torm:"type:decimal,precision:8,scale:2,comment:价格"`
}

// 修改后的模型
type Product struct {
    model.BaseModel
    Name        string  `torm:"type:varchar,size:100,comment:产品名称"`    // 长度：50→100
    Price       float64 `torm:"type:decimal,precision:10,scale:2,comment:价格"` // 精度：8→10
    Description string  `torm:"type:text,comment:产品描述"`                 // 新增字段
    Category    string  `torm:"type:varchar,size:50,comment:分类"`          // 新增字段
}
```

**执行 `AutoMigrate()` 后的输出：**

```bash
Executing: ALTER TABLE products MODIFY COLUMN name VARCHAR(100) NOT NULL COMMENT '产品名称', 
           MODIFY COLUMN price DECIMAL(10,2) NOT NULL COMMENT '价格', 
           ADD COLUMN description TEXT COMMENT '产品描述', 
           ADD COLUMN category VARCHAR(50) COMMENT '分类'
✅ Table structure updated successfully. Applied 4 changes.

📋 Schema Changes Applied:
| Column      | Action    | Details                                    |
|-------------|-----------|-------------------------------------------|
| name        | 🔧 MODIFY | length changed from 50 to 100            |
| price       | 🔧 MODIFY | precision changed from 8 to 10           |
| description | ➕ ADD    | Added description column with type TEXT   |
| category    | ➕ ADD    | Added category column with type VARCHAR   |
```

## 🛡️ 安全特性

### 自动备份和回滚

```go
import "github.com/zhoudm1743/torm/migration"

func safeSchemaUpdate() {
    conn, _ := db.DB("default")
    
    // 🛡️ 使用安全迁移器
    safeMigrator := migration.NewSafeMigrator(conn).
        SetDryRun(false).         // 设置为true可预演
        SetBackupTables(true).    // 自动备份表
        SetLogger(log.Default())  // 自定义日志器

    // 创建差异（通常由AutoMigrate内部调用）
    differences := []migration.ColumnDifference{
        {
            Column: "new_field",
            Type:   "add",
            NewValue: migration.ModelColumn{
                Name:   "new_field",
                Type:   migration.ColumnTypeVarchar,
                Length: 100,
            },
        },
    }

    // 🔧 安全执行变更
    result, err := safeMigrator.SafeAlterTable("products", differences)
    if err != nil {
        log.Printf("❌ Migration failed: %v", err)
        if result.RecoveryInstructions != "" {
            log.Printf("🔧 Recovery: %s", result.RecoveryInstructions)
        }
        return
    }

    // 📊 打印详细报告
    result.PrintSummary()
}
```

### 预演模式（Dry Run）

```go
func previewChanges() {
    conn, _ := db.DB("default")
    
    // 🔍 预演模式 - 只查看SQL，不执行
    safeMigrator := migration.NewSafeMigrator(conn).SetDryRun(true)
    
    result, err := safeMigrator.SafeAlterTable("products", differences)
    
    // 输出：
    // 🔍 DRY RUN MODE - SQL statements to be executed:
    //   1. ALTER TABLE products ADD COLUMN new_field VARCHAR(100)
}
```

## 📈 性能优化

### 大表变更建议

```go
// 对于大表，建议分步骤执行
func migrateLargeTable() {
    product := NewProduct()
    
    // 1. 先预演检查
    conn, _ := db.DB("default")
    safeMigrator := migration.NewSafeMigrator(conn).SetDryRun(true)
    
    // 2. 手动控制备份（可选择性关闭）
    safeMigrator.SetBackupTables(false) // 大表可能不需要自动备份
    
    // 3. 分批次执行（如果需要）
    err := product.AutoMigrate()
    if err != nil {
        log.Printf("考虑使用传统迁移文件处理复杂变更: %v", err)
    }
}
```

## 🗄️ 跨数据库支持

### MySQL 示例

```sql
-- 自动生成的MySQL语句
ALTER TABLE users 
  MODIFY COLUMN name VARCHAR(100) NOT NULL COMMENT '用户名',
  ADD COLUMN phone VARCHAR(20) COMMENT '手机号'
```

### PostgreSQL 示例

```sql
-- 自动生成的PostgreSQL语句
ALTER TABLE users ALTER COLUMN name TYPE VARCHAR(100);
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
```

### SQLite 示例

```sql
-- SQLite的复杂变更会生成警告
-- WARNING: SQLite table users requires recreation for complex changes
-- Please use migration system for complex SQLite schema changes
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
```

## 🚀 最佳实践

### 1. 开发流程

```go
// 开发环境：快速迭代
user := NewUser()
err := user.AutoMigrate() // 直接使用，自动处理

// 生产环境：谨慎操作
func productionDeploy() {
    // 预演检查
    safeMigrator := migration.NewSafeMigrator(conn).SetDryRun(true)
    result, _ := safeMigrator.SafeAlterTable("users", differences)
    result.PrintSummary()
    
    // 人工确认后执行
    confirm := getUserConfirmation()
    if confirm {
        safeMigrator.SetDryRun(false)
        result, err := safeMigrator.SafeAlterTable("users", differences)
        // ...处理结果
    }
}
```

### 2. 错误处理

```go
err := user.AutoMigrate()
if err != nil {
    if strings.Contains(err.Error(), "cannot auto-detect model structure") {
        // 使用推荐的方式创建模型
        log.Println("请使用 NewAutoMigrateModel 创建模型实例")
    } else if strings.Contains(err.Error(), "SQLite table") {
        // SQLite复杂变更需要传统迁移
        log.Println("SQLite复杂变更建议使用传统迁移文件")
    } else {
        log.Printf("迁移失败: %v", err)
    }
}
```

### 3. 版本控制

```go
// 在模型文件中记录变更历史
type User struct {
    model.BaseModel
    // v1.1.6: 增加手机号字段
    Phone string `torm:"type:varchar,size:20,comment:手机号"`
    // v1.1.5: 邮箱长度从100改为150
    Email string `torm:"type:varchar,size:150,unique,comment:邮箱"`
}
```

## 🎯 总结

TORM v1.1.6 的表结构自动对比和更新功能让您可以：

- ✅ **零配置使用** - `NewAutoMigrateModel` + `AutoMigrate()`
- ✅ **安全可靠** - 自动备份、事务回滚、预演模式
- ✅ **跨数据库** - MySQL、PostgreSQL、SQLite 智能适配
- ✅ **详细报告** - 每个变更的具体信息和原因
- ✅ **向后兼容** - 现有代码无需修改

现在您可以专注于业务逻辑，让 TORM 自动处理数据库表结构的演进！

---

> 💡 **提示**: 对于生产环境，建议结合传统迁移文件使用，实现更精细的版本控制和回滚策略。
