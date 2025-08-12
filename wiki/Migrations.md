# 数据迁移系统

TORM提供了强大的数据迁移工具，用于版本化管理数据库结构。该系统参考了现代ORM和Laravel的设计理念，提供了更加灵活和强大的功能。

## 🎯 核心概念

### 迁移是什么？

数据迁移是一种版本控制系统，用于管理数据库结构的变更。每个迁移文件代表数据库的一个特定变更，可以向前执行（Up）或向后回滚（Down）。

### 主要优势

- **版本控制**: 跟踪数据库结构的所有变更
- **团队协作**: 确保团队成员数据库结构一致
- **环境同步**: 轻松在开发、测试、生产环境间同步
- **回滚支持**: 安全地回滚有问题的变更
- **批次管理**: 按批次组织和管理迁移

## 🚀 快速开始

### 基本迁移示例

```go
package main

import (
    "context"
    "log"
    
    "github.com/zhoudm1743/torm/db"
    "github.com/zhoudm1743/torm/migration"
)

func main() {
    // 配置数据库连接
    config := &db.Config{
        Driver:   "mysql",
        Host:     "localhost",
        Port:     3306,
        Database: "myapp",
        Username: "root",
        Password: "password",
    }
    
    db.AddConnection("default", config)
    conn, _ := db.DB("default")
    
    // 创建迁移器
    migrator := migration.NewMigrator(conn, nil)
    
    // 注册迁移
    registerMigrations(migrator)
    
    // 执行迁移
    ctx := context.Background()
    if err := migrator.Up(ctx); err != nil {
        log.Fatal("迁移失败:", err)
    }
    
    // 查看迁移状态
    migrator.PrintStatus(ctx)
}

func registerMigrations(migrator *migration.Migrator) {
    // 创建用户表
    migrator.RegisterFunc(
        "20240101_000001",
        "创建用户表",
        createUsersUp,
        createUsersDown,
    )
    
    // 创建文章表
    migrator.RegisterFunc(
        "20240101_000002", 
        "创建文章表",
        createPostsUp,
        createPostsDown,
    )
}

func createUsersUp(ctx context.Context, conn db.ConnectionInterface) error {
    schema := migration.NewSchemaBuilder(conn)
    
    table := &migration.Table{
        Name: "users",
        Columns: []*migration.Column{
            {
                Name:          "id",
                Type:          migration.ColumnTypeBigInt,
                PrimaryKey:    true,
                AutoIncrement: true,
                NotNull:       true,
            },
            {
                Name:    "name",
                Type:    migration.ColumnTypeVarchar,
                Length:  100,
                NotNull: true,
            },
            {
                Name:    "email",
                Type:    migration.ColumnTypeVarchar,
                Length:  100,
                NotNull: true,
            },
            {
                Name:    "created_at",
                Type:    migration.ColumnTypeDateTime,
                Default: "CURRENT_TIMESTAMP",
            },
        },
        Indexes: []*migration.Index{
            {
                Name:    "idx_users_email",
                Columns: []string{"email"},
                Unique:  true,
            },
        },
    }
    
    return schema.CreateTable(ctx, table)
}

func createUsersDown(ctx context.Context, conn db.ConnectionInterface) error {
    schema := migration.NewSchemaBuilder(conn)
    return schema.DropTable(ctx, "users")
}
```

## 🏗️ 结构构建器

### 表定义

```go
table := &migration.Table{
    Name:    "posts",
    Comment: "文章表",
    Engine:  "InnoDB",  // MySQL引擎
    Charset: "utf8mb4", // 字符集
    
    Columns: []*migration.Column{
        // 主键
        {
            Name:          "id",
            Type:          migration.ColumnTypeBigInt,
            PrimaryKey:    true,
            AutoIncrement: true,
            NotNull:       true,
            Comment:       "主键ID",
        },
        
        // 字符串字段
        {
            Name:    "title",
            Type:    migration.ColumnTypeVarchar,
            Length:  200,
            NotNull: true,
            Comment: "文章标题",
        },
        
        // 文本字段
        {
            Name:    "content",
            Type:    migration.ColumnTypeText,
            Comment: "文章内容",
        },
        
        // 数值字段
        {
            Name:    "view_count",
            Type:    migration.ColumnTypeInt,
            Default: 0,
            Comment: "浏览次数",
        },
        
        // 布尔字段
        {
            Name:    "is_published",
            Type:    migration.ColumnTypeBoolean,
            Default: false,
            Comment: "是否发布",
        },
        
        // 外键字段
        {
            Name:    "user_id",
            Type:    migration.ColumnTypeBigInt,
            NotNull: true,
            Comment: "作者ID",
        },
        
        // 时间字段
        {
            Name:    "created_at",
            Type:    migration.ColumnTypeDateTime,
            Default: "CURRENT_TIMESTAMP",
            Comment: "创建时间",
        },
        {
            Name:    "updated_at",
            Type:    migration.ColumnTypeDateTime,
            Comment: "更新时间",
        },
    },
    
    // 索引定义
    Indexes: []*migration.Index{
        {
            Name:    "idx_posts_user_id",
            Columns: []string{"user_id"},
        },
        {
            Name:    "idx_posts_published",
            Columns: []string{"is_published", "created_at"},
        },
        {
            Name:    "idx_posts_title_unique",
            Columns: []string{"title"},
            Unique:  true,
        },
    },
    
    // 外键定义
    ForeignKeys: []*migration.ForeignKey{
        {
            Name:              "fk_posts_user_id",
            Columns:           []string{"user_id"},
            ReferencedTable:   "users",
            ReferencedColumns: []string{"id"},
            OnDelete:          "CASCADE",
            OnUpdate:          "CASCADE",
        },
    },
}
```

### 支持的列类型

```go
// 数值类型
migration.ColumnTypeInt       // INT
migration.ColumnTypeBigInt    // BIGINT
migration.ColumnTypeSmallInt  // SMALLINT
migration.ColumnTypeTinyInt   // TINYINT
migration.ColumnTypeFloat     // FLOAT
migration.ColumnTypeDouble    // DOUBLE
migration.ColumnTypeDecimal   // DECIMAL(precision, scale)

// 字符串类型
migration.ColumnTypeVarchar   // VARCHAR(length)
migration.ColumnTypeChar      // CHAR(length)
migration.ColumnTypeText      // TEXT
migration.ColumnTypeLongText  // LONGTEXT

// 时间类型
migration.ColumnTypeDateTime  // DATETIME
migration.ColumnTypeTimestamp // TIMESTAMP
migration.ColumnTypeDate      // DATE
migration.ColumnTypeTime      // TIME

// 其他类型
migration.ColumnTypeBoolean   // BOOLEAN
migration.ColumnTypeBlob      // BLOB
migration.ColumnTypeJSON      // JSON
```

### 修改表结构

```go
// 添加列
func addEmailColumn(ctx context.Context, conn db.ConnectionInterface) error {
    schema := migration.NewSchemaBuilder(conn)
    
    column := &migration.Column{
        Name:   "email",
        Type:   migration.ColumnTypeVarchar,
        Length: 100,
    }
    
    return schema.AddColumn(ctx, "users", column)
}

// 删除列
func dropEmailColumn(ctx context.Context, conn db.ConnectionInterface) error {
    schema := migration.NewSchemaBuilder(conn)
    return schema.DropColumn(ctx, "users", "email")
}

// 修改列
func modifyEmailColumn(ctx context.Context, conn db.ConnectionInterface) error {
    schema := migration.NewSchemaBuilder(conn)
    
    column := &migration.Column{
        Name:    "email",
        Type:    migration.ColumnTypeVarchar,
        Length:  150,  // 修改长度
        NotNull: true, // 添加NOT NULL约束
    }
    
    return schema.ModifyColumn(ctx, "users", column)
}

// 创建索引
func createEmailIndex(ctx context.Context, conn db.ConnectionInterface) error {
    schema := migration.NewSchemaBuilder(conn)
    
    index := &migration.Index{
        Name:    "idx_users_email",
        Columns: []string{"email"},
        Unique:  true,
    }
    
    return schema.CreateIndex(ctx, "users", index)
}

// 删除索引
func dropEmailIndex(ctx context.Context, conn db.ConnectionInterface) error {
    schema := migration.NewSchemaBuilder(conn)
    return schema.DropIndex(ctx, "users", "idx_users_email")
}
```

## 🎮 迁移管理

### 迁移器配置

```go
// 创建迁移器
migrator := migration.NewMigrator(conn, logger)

// 自定义迁移表名
migrator.SetTableName("my_migrations")

// 禁用自动创建迁移表
migrator.SetAutoCreate(false)
```

### 迁移操作

```go
ctx := context.Background()

// 执行所有待执行的迁移
err := migrator.Up(ctx)

// 回滚最后N个迁移
err := migrator.Down(ctx, 2)

// 重置所有迁移（回滚全部）
err := migrator.Reset(ctx)

// 清空数据库并重新执行所有迁移
err := migrator.Fresh(ctx)

// 查看迁移状态
status, err := migrator.Status(ctx)

// 打印迁移状态
migrator.PrintStatus(ctx)
```

### 迁移状态查看

```go
status, err := migrator.Status(ctx)
if err != nil {
    log.Fatal(err)
}

for _, s := range status {
    fmt.Printf("迁移: %s\n", s.Version)
    fmt.Printf("描述: %s\n", s.Description)
    fmt.Printf("状态: %v\n", s.Applied)
    if s.Applied {
        fmt.Printf("批次: %d\n", s.Batch)
        fmt.Printf("应用时间: %v\n", s.AppliedAt)
    }
    fmt.Println("---")
}
```

## 🔄 批次管理

TORM的迁移系统支持批次管理，这是一个强大的功能：

### 批次概念

- 每次运行 `Up()` 时，所有待执行的迁移会被分配到同一个批次
- 回滚时可以按批次回滚，而不是按单个迁移
- 批次号自动递增

### 批次示例

```go
// 第一次执行 Up() - 批次1
migrator.RegisterFunc("20240101_001", "创建用户表", createUsers, dropUsers)
migrator.RegisterFunc("20240101_002", "创建文章表", createPosts, dropPosts)
migrator.Up(ctx) // 两个迁移都在批次1

// 添加新迁移后再次执行 Up() - 批次2
migrator.RegisterFunc("20240101_003", "添加评论表", createComments, dropComments)
migrator.Up(ctx) // 评论表迁移在批次2

// 回滚最后一个批次（批次2）
migrator.Down(ctx, 1) // 只回滚评论表迁移

// 回滚前两个批次
migrator.Down(ctx, 2) // 回滚所有迁移
```

## 🗄️ 多数据库支持

### MySQL 特定功能

```go
func createMySQLTable(ctx context.Context, conn db.ConnectionInterface) error {
    schema := migration.NewSchemaBuilder(conn)
    
    table := &migration.Table{
        Name:    "mysql_table",
        Engine:  "InnoDB",        // MySQL引擎
        Charset: "utf8mb4",       // 字符集
        Comment: "MySQL专用表",
        
        Columns: []*migration.Column{
            {
                Name:          "id",
                Type:          migration.ColumnTypeBigInt,
                PrimaryKey:    true,
                AutoIncrement: true,
                NotNull:       true,
            },
            {
                Name:    "data",
                Type:    migration.ColumnTypeJSON, // MySQL JSON类型
            },
        },
    }
    
    return schema.CreateTable(ctx, table)
}
```

### PostgreSQL 特定功能

```go
func createPostgreSQLTable(ctx context.Context, conn db.ConnectionInterface) error {
    schema := migration.NewSchemaBuilder(conn)
    
    table := &migration.Table{
        Name: "postgres_table",
        
        Columns: []*migration.Column{
            {
                Name:          "id",
                Type:          migration.ColumnTypeBigInt,
                PrimaryKey:    true,
                AutoIncrement: true, // 自动生成BIGSERIAL
                NotNull:       true,
            },
            {
                Name: "data",
                Type: migration.ColumnTypeJSON, // PostgreSQL JSONB
            },
        },
    }
    
    return schema.CreateTable(ctx, table)
}
```

### SQLite 注意事项

SQLite有一些限制，TORM会自动处理：

```go
// SQLite不支持某些操作，TORM会优雅处理
func sqliteLimitations(ctx context.Context, conn db.ConnectionInterface) error {
    schema := migration.NewSchemaBuilder(conn)
    
    // 添加UNIQUE列（SQLite不支持）
    // TORM会先添加普通列，然后创建UNIQUE索引
    column := &migration.Column{
        Name:   "email",
        Type:   migration.ColumnTypeVarchar,
        Length: 100,
        Unique: true, // TORM会自动处理
    }
    
    return schema.AddColumn(ctx, "users", column)
}
```

## 📝 最佳实践

### 1. 迁移命名规范

```go
// 推荐的命名格式：YYYYMMDD_HHMMSS_description
"20240101_120000_create_users_table"
"20240101_120001_add_email_to_users"
"20240101_120002_create_posts_table"
"20240102_090000_add_index_to_posts"
```

### 2. 迁移文件组织

```go
// migrations/migrations.go
package migrations

import (
    "github.com/zhoudm1743/torm/migration"
)

func RegisterAll(migrator *migration.Migrator) {
    // 按时间顺序注册
    migrator.Register(NewCreateUsersTable())
    migrator.Register(NewCreatePostsTable())
    migrator.Register(NewAddEmailToUsers())
    // ...
}

// migrations/001_create_users_table.go
package migrations

type CreateUsersTable struct{}

func NewCreateUsersTable() *CreateUsersTable {
    return &CreateUsersTable{}
}

func (m *CreateUsersTable) Version() string {
    return "20240101_000001"
}

func (m *CreateUsersTable) Description() string {
    return "创建用户表"
}

func (m *CreateUsersTable) Up(ctx context.Context, conn db.ConnectionInterface) error {
    // 实现Up逻辑
}

func (m *CreateUsersTable) Down(ctx context.Context, conn db.ConnectionInterface) error {
    // 实现Down逻辑
}
```

### 3. 数据迁移

```go
// 结构迁移和数据迁移分离
func migrateUserData(ctx context.Context, conn db.ConnectionInterface) error {
    // 1. 先添加新列
    schema := migration.NewSchemaBuilder(conn)
    newColumn := &migration.Column{
        Name: "full_name",
        Type: migration.ColumnTypeVarchar,
        Length: 200,
    }
    
    if err := schema.AddColumn(ctx, "users", newColumn); err != nil {
        return err
    }
    
    // 2. 迁移数据
    updateSQL := `
        UPDATE users 
        SET full_name = CONCAT(first_name, ' ', last_name) 
        WHERE full_name IS NULL
    `
    
    _, err := conn.Exec(ctx, updateSQL)
    if err != nil {
        return err
    }
    
    // 3. 删除旧列（可选）
    if err := schema.DropColumn(ctx, "users", "first_name"); err != nil {
        return err
    }
    if err := schema.DropColumn(ctx, "users", "last_name"); err != nil {
        return err
    }
    
    return nil
}
```

### 4. 回滚策略

```go
// 确保每个Up操作都有对应的Down操作
func createComplexTableUp(ctx context.Context, conn db.ConnectionInterface) error {
    schema := migration.NewSchemaBuilder(conn)
    
    // 创建表
    if err := schema.CreateTable(ctx, table); err != nil {
        return err
    }
    
    // 创建索引
    for _, index := range indexes {
        if err := schema.CreateIndex(ctx, table.Name, index); err != nil {
            // 如果索引创建失败，清理已创建的表
            schema.DropTable(ctx, table.Name)
            return err
        }
    }
    
    return nil
}

func createComplexTableDown(ctx context.Context, conn db.ConnectionInterface) error {
    schema := migration.NewSchemaBuilder(conn)
    
    // Down操作要与Up操作完全相反
    // 先删除索引，再删除表
    for _, index := range indexes {
        schema.DropIndex(ctx, table.Name, index.Name) // 忽略错误
    }
    
    return schema.DropTable(ctx, table.Name)
}
```

## 🔧 高级功能

### 事务支持

```go
func transactionalMigration(ctx context.Context, conn db.ConnectionInterface) error {
    // 开始事务
    tx, err := conn.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback() // 确保回滚
    
    // 在事务中执行多个操作
    _, err = tx.Exec(ctx, "CREATE TABLE temp_table (id INT)")
    if err != nil {
        return err
    }
    
    _, err = tx.Exec(ctx, "INSERT INTO temp_table VALUES (1)")
    if err != nil {
        return err
    }
    
    // 提交事务
    return tx.Commit()
}
```

### 条件迁移

```go
func conditionalMigration(ctx context.Context, conn db.ConnectionInterface) error {
    // 检查表是否存在
    checkSQL := `
        SELECT COUNT(*) 
        FROM information_schema.tables 
        WHERE table_name = 'existing_table'
    `
    
    row := conn.QueryRow(ctx, checkSQL)
    var count int
    if err := row.Scan(&count); err != nil {
        return err
    }
    
    // 只有表不存在时才创建
    if count == 0 {
        schema := migration.NewSchemaBuilder(conn)
        return schema.CreateTable(ctx, table)
    }
    
    return nil
}
```

---

**📚 更多信息请参考 [API参考文档](API-Reference) 和 [最佳实践](Best-Practices)。** 