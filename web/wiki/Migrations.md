# 数据迁移系统

TORM提供了强大的数据迁移工具，用于版本化管理数据库结构。该系统参考了现代ORM和Laravel的设计理念，提供了更加灵活和强大的功能。

**🆕 v1.1.6 新增**: AutoMigrate 自动迁移功能，支持从模型结构体自动生成表结构，结合传统迁移提供更完整的数据库管理方案。

## 🎯 核心概念

### 迁移是什么？

数据迁移是一种版本控制系统，用于管理数据库结构的变更。每个迁移文件代表数据库的一个特定变更，可以向前执行（Up）或向后回滚（Down）。

### 主要优势

- **版本控制**: 跟踪数据库结构的所有变更
- **团队协作**: 确保团队成员数据库结构一致
- **环境同步**: 轻松在开发、测试、生产环境间同步
- **回滚支持**: 安全地回滚有问题的变更
- **批次管理**: 按批次组织和管理迁移

### AutoMigrate vs 传统迁移

| 特性 | AutoMigrate | 传统迁移 | 推荐场景 |
|------|-------------|----------|----------|
| **学习成本** | 低 | 中等 | 快速原型开发 |
| **代码维护** | 结构体即文档 | 需要额外迁移文件 | 简单业务逻辑 |
| **版本控制** | 基于模型变更 | 精确的版本控制 | 复杂企业应用 |
| **回滚能力** | 有限 | 完全支持 | 生产环境变更 |
| **复杂变更** | 不支持 | 完全支持 | 数据迁移、索引优化 |
| **团队协作** | 简单 | 需要协调 | 小团队快速迭代 |

### 建议使用策略

```go
// 🚀 快速开发阶段：使用 AutoMigrate
func developmentSetup() {
    models := []interface{}{
        NewUser(),
        NewProduct(), 
        NewOrder(),
    }
    
    for _, model := range models {
        if migrator, ok := model.(interface{ AutoMigrate() error }); ok {
            migrator.AutoMigrate()
        }
    }
}

// 🏭 生产环境：结合使用
func productionSetup() {
    // 1. AutoMigrate 创建基础表结构
    user := NewUser()
    user.AutoMigrate()
    
    // 2. 传统迁移处理复杂变更
    migrator := migration.NewMigrator(conn, logger)
    migrator.RegisterFunc("20240101_001", "优化用户表索引", optimizeUserIndexes, rollbackIndexes)
    migrator.RegisterFunc("20240101_002", "迁移历史数据", migrateHistoricalData, rollbackData)
    migrator.Up()
}
```

## 🚀 快速开始

### AutoMigrate 快速开始

```go
package main

import (
    "log"
    "github.com/zhoudm1743/torm/db"
    "github.com/zhoudm1743/torm/model"
)

// 定义模型
type User struct {
    model.BaseModel
    ID        int64  `json:"id" db:"id" primaryKey:"true" autoIncrement:"true"`
    Email     string `json:"email" db:"email" size:"100" unique:"true"`
    Name      string `json:"name" db:"name" size:"50"`
    Age       int    `json:"age" db:"age" default:"0"`
    CreatedAt int64  `json:"created_at" db:"created_at" autoCreateTime:"true"`
    UpdatedAt int64  `json:"updated_at" db:"updated_at" autoUpdateTime:"true"`
}

func NewUser() *User {
    user := &User{}
    user.BaseModel = *model.NewBaseModelWithAutoDetect(user)
    user.SetTable("users")
    user.SetConnection("default")
    return user
}

func main() {
    // 配置数据库
    config := &db.Config{
        Driver:   "mysql",
        Host:     "localhost",
        Database: "myapp",
        Username: "root",
        Password: "password",
    }
    
    db.AddConnection("default", config)
    
    // 一行代码创建表
    user := NewUser()
    if err := user.AutoMigrate(); err != nil {
        log.Fatalf("AutoMigrate 失败: %v", err)
    }
    
    log.Println("数据库表创建成功！")
}
```

### 传统迁移示例

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

### AutoMigrate 集成

#### 结合传统迁移的最佳实践

```go
// migration_manager.go
type MigrationManager struct {
    migrator *migration.Migrator
    models   []interface{}
}

func NewMigrationManager(conn db.ConnectionInterface, logger db.LoggerInterface) *MigrationManager {
    return &MigrationManager{
        migrator: migration.NewMigrator(conn, logger),
        models:   make([]interface{}, 0),
    }
}

// 注册模型（用于 AutoMigrate）
func (m *MigrationManager) RegisterModel(model interface{}) {
    m.models = append(m.models, model)
}

// 注册传统迁移
func (m *MigrationManager) RegisterMigration(migration migration.MigrationInterface) {
    m.migrator.Register(migration)
}

// 执行完整迁移流程
func (m *MigrationManager) ExecuteAll() error {
    // 1. 首先执行 AutoMigrate 创建基础表结构
    for _, model := range m.models {
        if autoMigrator, ok := model.(interface{ AutoMigrate() error }); ok {
            if err := autoMigrator.AutoMigrate(); err != nil {
                return fmt.Errorf("AutoMigrate failed for %T: %w", model, err)
            }
        }
    }
    
    // 2. 然后执行传统迁移处理复杂变更
    return m.migrator.Up()
}

// 使用示例
func setupDatabase() {
    conn, _ := db.DB("default")
    logger := &CustomLogger{}
    
    manager := NewMigrationManager(conn, logger)
    
    // 注册模型
    manager.RegisterModel(NewUser())
    manager.RegisterModel(NewProduct())
    manager.RegisterModel(NewOrder())
    
    // 注册传统迁移
    manager.RegisterMigration(NewAddUserIndexesMigration())
    manager.RegisterMigration(NewOptimizeProductTableMigration())
    
    // 执行所有迁移
    if err := manager.ExecuteAll(); err != nil {
        log.Fatalf("Migration failed: %v", err)
    }
}
```

#### AutoMigrate 增强迁移

将 AutoMigrate 封装为传统迁移：

```go
// auto_migrate_migration.go
type AutoMigrateMigration struct {
    version     string
    description string
    models      []interface{}
}

func NewAutoMigrateMigration(version, description string, models ...interface{}) *AutoMigrateMigration {
    return &AutoMigrateMigration{
        version:     version,
        description: description,
        models:      models,
    }
}

func (m *AutoMigrateMigration) Version() string {
    return m.version
}

func (m *AutoMigrateMigration) Description() string {
    return m.description
}

func (m *AutoMigrateMigration) Up(ctx context.Context, conn db.ConnectionInterface) error {
    for _, model := range m.models {
        if autoMigrator, ok := model.(interface{ AutoMigrate() error }); ok {
            if err := autoMigrator.AutoMigrate(); err != nil {
                return fmt.Errorf("AutoMigrate failed for %T: %w", model, err)
            }
        }
    }
    return nil
}

func (m *AutoMigrateMigration) Down(ctx context.Context, conn db.ConnectionInterface) error {
    // AutoMigrate 的回滚通常是删除表
    schema := migration.NewSchemaBuilder(conn)
    
    for _, model := range m.models {
        if tableNamer, ok := model.(interface{ TableName() string }); ok {
            tableName := tableNamer.TableName()
            if tableName != "" {
                schema.DropTable(ctx, tableName)
            }
        }
    }
    return nil
}

// 注册 AutoMigrate 作为传统迁移
func registerAutoMigrations(migrator *migration.Migrator) {
    migrator.Register(NewAutoMigrateMigration(
        "20240101_000001",
        "AutoMigrate: 创建基础表结构",
        NewUser(),
        NewProduct(),
        NewOrder(),
    ))
}
```

#### 增量 AutoMigrate

```go
// 检测模型变更并只迁移变更部分
type IncrementalAutoMigrator struct {
    conn db.ConnectionInterface
}

func (i *IncrementalAutoMigrator) MigrateIfChanged(model interface{}) error {
    tableName := i.getTableName(model)
    
    // 检查表是否存在
    exists, err := i.tableExists(tableName)
    if err != nil {
        return err
    }
    
    if !exists {
        // 表不存在，执行完整 AutoMigrate
        if autoMigrator, ok := model.(interface{ AutoMigrate() error }); ok {
            return autoMigrator.AutoMigrate()
        }
    } else {
        // 表存在，检查结构差异
        return i.updateTableStructure(model, tableName)
    }
    
    return nil
}

func (i *IncrementalAutoMigrator) updateTableStructure(model interface{}, tableName string) error {
    // 获取当前表结构
    currentColumns, err := i.getTableColumns(tableName)
    if err != nil {
        return err
    }
    
    // 获取模型期望的结构
    expectedColumns, err := i.getModelColumns(model)
    if err != nil {
        return err
    }
    
    // 比较差异并执行必要的 ALTER TABLE 操作
    return i.applyColumnDifferences(tableName, currentColumns, expectedColumns)
}
```

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

## 🆕 v1.1.6 新增功能总结

### AutoMigrate 核心特性

- ✅ **智能类型映射**: 支持所有 Go 基础类型自动映射到对应数据库类型
- ✅ **跨数据库兼容**: 完美适配 MySQL、PostgreSQL、SQLite 的类型差异
- ✅ **丰富标签支持**: 支持 30+ 种字段标签，精确控制表结构
- ✅ **自动索引创建**: 智能创建唯一索引、普通索引和外键索引
- ✅ **表存在性检查**: 自动检测表是否存在，避免重复创建
- ✅ **默认配置检测**: `NewBaseModelWithAutoDetect` 简化模型创建流程

### TORM 统一标签语法

v1.1.6 引入统一的 `torm` 标签，大大简化模型定义：

| 分类 | 标签语法 | 作用 | 示例 |
|------|----------|------|------|
| **主键和约束** | `primary_key` / `pk` | 主键 | `torm:"primary_key"` |
| | `auto_increment` | 自增 | `torm:"primary_key,auto_increment"` |
| | `unique` | 唯一约束 | `torm:"unique"` |
| | `nullable` | 允许NULL | `torm:"nullable"` |
| | `not_null` | 不允许NULL | `torm:"not_null"` |
| **数据类型** | `type:类型名` | 明确数据库类型 | `torm:"type:varchar,size:100"` |
| | `size:数字` | 字段长度 | `torm:"size:100"` |
| | `precision:数字` | 数值精度 | `torm:"type:decimal,precision:10"` |
| | `scale:数字` | 小数位数 | `torm:"precision:10,scale:2"` |
| **索引优化** | `index` | 普通索引 | `torm:"index"` |
| | `index:名称` | 自定义索引名 | `torm:"index:phone_idx"` |
| **默认值** | `default:值` | 默认值 | `torm:"default:1"` |
| | `default:true/false` | 布尔默认值 | `torm:"default:true"` |
| | `default:current_timestamp` | 时间默认值 | `torm:"default:current_timestamp"` |
| **时间戳** | `auto_create_time` | 创建时间 | `torm:"auto_create_time"` |
| | `auto_update_time` | 更新时间 | `torm:"auto_update_time"` |
| **其他** | `comment:描述` | 列注释 | `torm:"comment:用户名"` |

#### 组合使用示例

```go
type Product struct {
    model.BaseModel
    // 主键：自增+主键+注释
    ID     int64  `db:"id" torm:"primary_key,auto_increment,comment:产品ID"`
    
    // 字符串：类型+长度+唯一+注释
    SKU    string `db:"sku" torm:"type:varchar,size:50,unique,comment:产品编码"`
    
    // 数值：类型+精度+默认值+注释
    Price  float64 `db:"price" torm:"type:decimal,precision:10,scale:2,default:0.00,comment:价格"`
    
    // 索引：自定义索引名+注释
    UserID int64  `db:"user_id" torm:"index:product_user_idx,comment:用户ID"`
    
    // 时间：自动创建时间
    CreatedAt int64 `db:"created_at" torm:"auto_create_time,comment:创建时间"`
}
```

### 类型映射支持矩阵

| Go 类型 | MySQL | PostgreSQL | SQLite | 支持标签 |
|---------|-------|------------|--------|----------|
| `string` | VARCHAR(n) | VARCHAR(n) | TEXT | size, type, fixed |
| `int8` | TINYINT | SMALLINT | INTEGER | type |
| `int16` | SMALLINT | SMALLINT | INTEGER | type |
| `int32` | INT | INTEGER | INTEGER | type |
| `int64` | BIGINT | BIGINT | INTEGER | type |
| `float32` | FLOAT | REAL | REAL | type, decimal |
| `float64` | DOUBLE | DOUBLE PRECISION | REAL | type, decimal, precision, scale |
| `bool` | BOOLEAN | BOOLEAN | INTEGER | type |
| `[]byte` | BLOB | BYTEA | BLOB | type |
| `[]string` | JSON | JSONB | TEXT | type |
| `map[string]interface{}` | JSON | JSONB | TEXT | type |
| `time.Time` | DATETIME | TIMESTAMP | DATETIME | type, timestamp |
| `*T` | NULL-able | NULL-able | NULL-able | nullable |

### 迁移策略对比

| 场景 | AutoMigrate | 传统迁移 | 推荐组合 |
|------|-------------|----------|----------|
| **新项目快速启动** | ⭐⭐⭐⭐⭐ | ⭐⭐ | 纯 AutoMigrate |
| **简单 CRUD 应用** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | AutoMigrate 为主 |
| **复杂业务逻辑** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | AutoMigrate + 传统迁移 |
| **生产环境部署** | ⭐⭐ | ⭐⭐⭐⭐⭐ | 传统迁移为主 |
| **团队协作开发** | ⭐⭐⭐ | ⭐⭐⭐⭐ | 混合使用 |
| **数据迁移需求** | ❌ | ⭐⭐⭐⭐⭐ | 传统迁移 |

---

**📚 更多信息请参考 [API参考文档](API-Reference)、[模型系统](Model-System) 和 [最佳实践](Best-Practices)。** 