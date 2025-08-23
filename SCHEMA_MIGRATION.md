# TORM 表结构自动对比和更新功能

## 🎯 功能概述

TORM v1.1.6 新增了强大的表结构自动对比和更新功能，可以：

- 🔍 **自动对比**：对比模型定义与数据库表结构的差异
- 🔧 **智能更新**：生成并执行 ALTER TABLE 语句更新表结构
- 💾 **安全备份**：更新前自动备份表数据
- 🔄 **支持回滚**：提供完整的错误恢复机制
- 📊 **多数据库支持**：支持 MySQL、PostgreSQL、SQLite
- 🧪 **预演模式**：可以预览将要执行的 SQL 而不实际执行

## 🚀 快速开始

### 基本使用

现在当您修改模型标签后，再次调用 `AutoMigrate()` 会自动对比并更新表结构：

```go
// 原模型
type User struct {
    model.BaseModel
    Name  string `torm:"type:varchar,size:50"`
    Email string `torm:"type:varchar,size:100"`
}

// 修改后的模型
type User struct {
    model.BaseModel
    Name  string `torm:"type:varchar,size:100"`      // 长度从50改为100
    Email string `torm:"type:varchar,size:200"`      // 长度从100改为200
    Phone string `torm:"type:varchar,size:20"`       // 新增字段
}

func main() {
    user := NewUser()
    
    // 第一次运行：创建表
    err := user.AutoMigrate()
    
    // 修改模型后再次运行：自动对比并更新表结构
    err = user.AutoMigrate()  // 🎉 自动更新！
}
```

### 输出示例

```bash
Executing: ALTER TABLE users MODIFY COLUMN name VARCHAR(100) NOT NULL, MODIFY COLUMN email VARCHAR(200) NOT NULL, ADD COLUMN phone VARCHAR(20)
✅ Table structure updated successfully. Applied 3 changes.

📋 Schema Changes Applied:
| Column | Action | Details |
|--------|--------|---------|
| name | 🔧 MODIFY | length changed from 50 to 100 |
| email | 🔧 MODIFY | length changed from 100 to 200 |
| phone | ➕ ADD | Added phone column with type VARCHAR |
```

## 🔧 高级功能

### 1. 安全迁移器

对于生产环境，建议使用 `SafeMigrator` 获得更多控制：

```go
import "github.com/zhoudm1743/torm/migration"

func safeSchemaUpdate() {
    conn, _ := db.DB("default")
    safeMigrator := migration.NewSafeMigrator(conn).
        SetDryRun(false).        // 设置为true可预演
        SetBackupTables(true)    // 自动备份表

    // 手动执行对比和更新
    user := NewUser()
    modelStruct := reflect.TypeOf(*user)
    
    // 分析模型
    analyzer := migration.NewModelAnalyzer()
    modelColumns, _ := analyzer.AnalyzeModel(modelStruct)
    
    // 对比差异
    comparator := migration.NewSchemaComparator(conn)
    dbColumns, _ := comparator.GetDatabaseColumns("users")
    differences := comparator.CompareColumns(dbColumns, modelColumns)
    
    // 安全执行变更
    result, err := safeMigrator.SafeAlterTable("users", differences)
    if err != nil {
        log.Printf("Migration failed: %v", err)
        if result.RecoveryInstructions != "" {
            log.Printf("Recovery: %s", result.RecoveryInstructions)
        }
    }
    
    result.PrintSummary()
}
```

### 2. 预演模式（Dry Run）

在生产环境执行前，先预览要执行的 SQL：

```go
safeMigrator := migration.NewSafeMigrator(conn).SetDryRun(true)

result, err := safeMigrator.SafeAlterTable("users", differences)
// 只会打印SQL，不会实际执行
```

### 3. 表备份和恢复

```go
// 自动备份（默认启用）
safeMigrator := migration.NewSafeMigrator(conn).SetBackupTables(true)

// 手动恢复
err := safeMigrator.RestoreFromBackup("users", "users_backup_20240123_150405")

// 清理旧备份（保留7天内的备份）
err := safeMigrator.CleanupBackups("users", 7)
```

## 📋 支持的变更类型

### 字段修改

- ✅ **类型变更**：`VARCHAR` → `TEXT`
- ✅ **长度变更**：`VARCHAR(50)` → `VARCHAR(100)`
- ✅ **精度变更**：`DECIMAL(10,2)` → `DECIMAL(15,4)`
- ✅ **约束变更**：`NULL` → `NOT NULL`
- ✅ **默认值变更**：`DEFAULT 0` → `DEFAULT 1`
- ✅ **注释变更**：添加或修改注释

### 字段操作

- ✅ **添加字段**：模型中新增的字段
- ✅ **删除字段**：模型中移除的字段 ⚠️
- ❌ **字段重命名**：需要手动处理

⚠️ **注意**：删除字段会导致数据丢失，建议谨慎使用

### 数据库兼容性

| 功能 | MySQL | PostgreSQL | SQLite |
|------|-------|------------|--------|
| 添加字段 | ✅ | ✅ | ✅ |
| 修改字段类型 | ✅ | ✅ | ⚠️ 需重建表 |
| 修改字段长度 | ✅ | ✅ | ⚠️ 需重建表 |
| 删除字段 | ✅ | ✅ | ⚠️ 需重建表 |
| 修改约束 | ✅ | ✅ | ⚠️ 需重建表 |

## 🏷️ 标签语法支持

支持所有 TORM 统一标签语法：

```go
type Product struct {
    model.BaseModel
    ID          int64   `torm:"primary_key,auto_increment,comment:产品ID"`
    Name        string  `torm:"type:varchar,size:200,comment:产品名称"`
    Description string  `torm:"type:text,comment:产品描述"`
    Price       float64 `torm:"type:decimal,precision:10,scale:2,comment:价格"`
    SKU         string  `torm:"type:varchar,size:50,unique,comment:产品编码"`
    CategoryID  int64   `torm:"type:bigint,index,comment:分类ID"`
    IsActive    bool    `torm:"type:boolean,default:true,comment:是否启用"`
    Stock       int     `torm:"type:int,default:0,not_null,comment:库存"`
    CreatedAt   int64   `torm:"auto_create_time,comment:创建时间"`
    UpdatedAt   int64   `torm:"auto_update_time,comment:更新时间"`
}
```

## ⚠️ 重要注意事项

### 生产环境使用

1. **始终备份**：在生产环境执行前备份数据库
2. **使用预演**：先用 `SetDryRun(true)` 预览变更
3. **分步执行**：大型变更可以分多次小批量执行
4. **监控性能**：大表的结构变更可能耗时较长

### 数据安全

```go
// 推荐的生产环境流程
func productionMigration() {
    safeMigrator := migration.NewSafeMigrator(conn).
        SetDryRun(true).         // 1. 先预演
        SetBackupTables(true)
    
    // 2. 预演检查
    result, err := safeMigrator.SafeAlterTable("users", differences)
    if err != nil {
        log.Fatal("Pre-flight check failed:", err)
    }
    
    // 3. 人工确认
    fmt.Println("Ready to execute:")
    result.PrintSummary()
    fmt.Print("Continue? (y/N): ")
    // ... 等待确认
    
    // 4. 实际执行
    safeMigrator.SetDryRun(false)
    result, err = safeMigrator.SafeAlterTable("users", differences)
    if err != nil {
        log.Fatal("Migration failed:", err)
    }
}
```

### SQLite 限制

SQLite 对 ALTER TABLE 支持有限，复杂变更需要重建表：

```go
// SQLite复杂变更的处理
if driver == "sqlite" && hasComplexChanges {
    fmt.Println("⚠️ SQLite detected - complex changes require table recreation")
    fmt.Println("💡 Consider using traditional migration system for SQLite")
}
```

## 📚 API 参考

### SchemaComparator

```go
comparator := migration.NewSchemaComparator(conn)

// 获取数据库表结构
dbColumns, err := comparator.GetDatabaseColumns("table_name")

// 对比差异
differences := comparator.CompareColumns(dbColumns, modelColumns)
```

### AlterGenerator

```go
generator := migration.NewAlterGenerator(conn)

// 生成ALTER TABLE语句
statements, err := generator.GenerateAlterSQL("table_name", differences)
```

### SafeMigrator

```go
migrator := migration.NewSafeMigrator(conn).
    SetDryRun(false).
    SetBackupTables(true).
    SetLogger(customLogger)

result, err := migrator.SafeAlterTable("table_name", differences)
```

### ModelAnalyzer

```go
analyzer := migration.NewModelAnalyzer()

// 分析模型结构
modelColumns, err := analyzer.AnalyzeModel(reflect.TypeOf(YourModel{}))
```

## 🎯 最佳实践

1. **渐进式更新**：小步快跑，避免一次性大规模变更
2. **版本控制**：将模型变更纳入版本控制
3. **测试环境验证**：先在测试环境验证变更
4. **监控和日志**：保留详细的变更日志
5. **回滚计划**：准备变更失败的回滚方案

## 🔗 相关文档

- [AutoMigrate 使用指南](Model-System.md#自动迁移)
- [TORM 统一标签语法](TORM_TAG_MIGRATION.md)
- [传统迁移系统](Migrations.md)
- [API 参考](API-Reference.md)

---

🎉 现在您可以放心地修改模型标签，TORM 会自动为您处理复杂的表结构更新！
