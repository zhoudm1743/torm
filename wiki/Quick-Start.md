# 快速开始指南

本指南将在5分钟内帮助你开始使用TORM。我们将通过一个简单的博客应用示例来展示TORM的核心功能。

## 📋 前置要求

- Go 1.19 或更高版本
- 支持的数据库之一 (MySQL, PostgreSQL, SQLite, MongoDB)
- 基本的Go语言知识

## 🚀 第1步：安装TORM

```bash
# 创建新项目
mkdir torm-demo
cd torm-demo
go mod init torm-demo

# 安装TORM
go get github.com/zhoudm1743/torm
```

## 🔧 第2步：配置数据库连接

创建 `main.go` 文件：

```go
package main

import (
    "context"
    "log"
    "time"
    
    "torm/pkg/db"
)

func main() {
    // 配置数据库连接（这里使用SQLite作为示例）
    config := &db.Config{
        Driver:          "sqlite",
        Database:        "blog.db",
        MaxOpenConns:    10,
        MaxIdleConns:    5,
        ConnMaxLifetime: time.Hour,
        LogQueries:      true, // 开启查询日志
    }

    // 添加连接到连接池
    err := db.AddConnection("default", config)
    if err != nil {
        log.Fatal("连接数据库失败:", err)
    }

    log.Println("✅ 数据库连接成功!")
}
```

### 🔌 其他数据库配置示例

<details>
<summary>MySQL 配置</summary>

```go
config := &db.Config{
    Driver:   "mysql",
    Host:     "localhost",
    Port:     3306,
    Database: "blog",
    Username: "root",
    Password: "password",
    Charset:  "utf8mb4",
    Options: map[string]string{
        "parseTime": "true",
        "loc":       "Local",
    },
}
```
</details>

<details>
<summary>PostgreSQL 配置</summary>

```go
config := &db.Config{
    Driver:   "postgres",
    Host:     "localhost",
    Port:     5432,
    Database: "blog",
    Username: "postgres",
    Password: "password",
    SSLMode:  "disable",
}
```
</details>

<details>
<summary>MongoDB 配置</summary>

```go
config := &db.Config{
    Driver:   "mongodb",
    Host:     "localhost",
    Port:     27017,
    Database: "blog",
    Username: "", // 可选
    Password: "", // 可选
}
```
</details>

## 🏗️ 第3步：基本查询操作

更新 `main.go`，添加基本的数据库操作：

```go
package main

import (
    "context"
    "log"
    "time"
    
    "torm/pkg/db"
)

func main() {
    // ... 数据库配置代码 ...

    ctx := context.Background()
    
    // 获取数据库连接
    conn, err := db.DB("default")
    if err != nil {
        log.Fatal("获取连接失败:", err)
    }

    // 创建用户表
    err = createUserTable(ctx, conn)
    if err != nil {
        log.Fatal("创建表失败:", err)
    }

    // 插入用户数据
    err = insertUser(ctx, conn, "张三", "zhangsan@example.com", 25)
    if err != nil {
        log.Fatal("插入用户失败:", err)
    }

    // 查询用户
    users, err := queryUsers(ctx, conn)
    if err != nil {
        log.Fatal("查询用户失败:", err)
    }

    log.Printf("✅ 查询到 %d 个用户", len(users))
    for _, user := range users {
        log.Printf("用户: %s, 邮箱: %s, 年龄: %d", user.Name, user.Email, user.Age)
    }
}

// User 用户结构体
type User struct {
    ID        int64     `db:"id" json:"id"`
    Name      string    `db:"name" json:"name"`
    Email     string    `db:"email" json:"email"`
    Age       int       `db:"age" json:"age"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// 创建用户表
func createUserTable(ctx context.Context, conn db.ConnectionInterface) error {
    sql := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        email TEXT UNIQUE NOT NULL,
        age INTEGER NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )`
    
    _, err := conn.Exec(ctx, sql)
    if err != nil {
        return err
    }
    
    log.Println("✅ 用户表创建成功")
    return nil
}

// 插入用户
func insertUser(ctx context.Context, conn db.ConnectionInterface, name, email string, age int) error {
    sql := `INSERT INTO users (name, email, age) VALUES (?, ?, ?)`
    
    result, err := conn.Exec(ctx, sql, name, email, age)
    if err != nil {
        return err
    }
    
    id, _ := result.LastInsertId()
    log.Printf("✅ 用户插入成功，ID: %d", id)
    return nil
}

// 查询用户
func queryUsers(ctx context.Context, conn db.ConnectionInterface) ([]*User, error) {
    sql := `SELECT id, name, email, age, created_at FROM users ORDER BY created_at DESC`
    
    rows, err := conn.Query(ctx, sql)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []*User
    for rows.Next() {
        user := &User{}
        err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Age, &user.CreatedAt)
        if err != nil {
            return nil, err
        }
        users = append(users, user)
    }

    return users, nil
}
```

## 📦 第4步：使用数据迁移

创建 `migrations.go` 文件来管理数据库结构：

```go
package main

import (
    "context"
    "log"
    
    "torm/pkg/db"
    "torm/pkg/migration"
)

func runMigrations() {
    // 获取数据库连接
    conn, err := db.DB("default")
    if err != nil {
        log.Fatal("获取连接失败:", err)
    }

    // 创建迁移器
    migrator := migration.NewMigrator(conn, nil)

    // 注册迁移
    registerMigrations(migrator)

    // 执行迁移
    ctx := context.Background()
    err = migrator.Up(ctx)
    if err != nil {
        log.Fatal("执行迁移失败:", err)
    }

    log.Println("✅ 数据库迁移完成")
}

func registerMigrations(migrator *migration.Migrator) {
    // 创建用户表迁移
    migrator.RegisterFunc(
        "20240101_000001",
        "创建用户表",
        func(ctx context.Context, conn db.ConnectionInterface) error {
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
                        Name:    "age",
                        Type:    migration.ColumnTypeInt,
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
        },
        func(ctx context.Context, conn db.ConnectionInterface) error {
            schema := migration.NewSchemaBuilder(conn)
            return schema.DropTable(ctx, "users")
        },
    )

    // 创建文章表迁移
    migrator.RegisterFunc(
        "20240101_000002",
        "创建文章表",
        func(ctx context.Context, conn db.ConnectionInterface) error {
            schema := migration.NewSchemaBuilder(conn)
            
            table := &migration.Table{
                Name: "posts",
                Columns: []*migration.Column{
                    {
                        Name:          "id",
                        Type:          migration.ColumnTypeBigInt,
                        PrimaryKey:    true,
                        AutoIncrement: true,
                        NotNull:       true,
                    },
                    {
                        Name:    "title",
                        Type:    migration.ColumnTypeVarchar,
                        Length:  200,
                        NotNull: true,
                    },
                    {
                        Name: "content",
                        Type: migration.ColumnTypeText,
                    },
                    {
                        Name:    "user_id",
                        Type:    migration.ColumnTypeBigInt,
                        NotNull: true,
                    },
                    {
                        Name:    "created_at",
                        Type:    migration.ColumnTypeDateTime,
                        Default: "CURRENT_TIMESTAMP",
                    },
                },
                ForeignKeys: []*migration.ForeignKey{
                    {
                        Name:              "fk_posts_user_id",
                        Columns:           []string{"user_id"},
                        ReferencedTable:   "users",
                        ReferencedColumns: []string{"id"},
                        OnDelete:          "CASCADE",
                    },
                },
            }
            
            return schema.CreateTable(ctx, table)
        },
        func(ctx context.Context, conn db.ConnectionInterface) error {
            schema := migration.NewSchemaBuilder(conn)
            return schema.DropTable(ctx, "posts")
        },
    )
}
```

更新 `main.go` 来使用迁移：

```go
func main() {
    // ... 数据库配置代码 ...

    // 运行迁移
    runMigrations()

    // ... 其他代码 ...
}
```

## 🔗 第5步：多数据库支持示例

创建 `multi_db.go` 文件展示多数据库支持：

```go
package main

import (
    "context"
    "log"
    "time"
    
    "go.mongodb.org/mongo-driver/bson"
    "torm/pkg/db"
)

func multiDatabaseExample() {
    ctx := context.Background()

    // 配置多个数据库连接
    setupMultipleConnections()

    // SQLite 操作
    sqliteExample(ctx)

    // MongoDB 操作
    mongoExample(ctx)
}

func setupMultipleConnections() {
    // SQLite 连接
    sqliteConfig := &db.Config{
        Driver:   "sqlite",
        Database: "sqlite_blog.db",
    }
    db.AddConnection("sqlite", sqliteConfig)

    // MongoDB 连接
    mongoConfig := &db.Config{
        Driver:   "mongodb",
        Host:     "localhost",
        Port:     27017,
        Database: "mongo_blog",
    }
    db.AddConnection("mongodb", mongoConfig)

    log.Println("✅ 多数据库连接配置完成")
}

func sqliteExample(ctx context.Context) {
    conn, err := db.DB("sqlite")
    if err != nil {
        log.Printf("SQLite连接失败: %v", err)
        return
    }

    // 创建表
    sql := `CREATE TABLE IF NOT EXISTS articles (
        id INTEGER PRIMARY KEY,
        title TEXT,
        content TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )`
    
    _, err = conn.Exec(ctx, sql)
    if err != nil {
        log.Printf("SQLite创建表失败: %v", err)
        return
    }

    // 插入数据
    _, err = conn.Exec(ctx, 
        "INSERT INTO articles (title, content) VALUES (?, ?)",
        "SQLite文章", "这是SQLite数据库中的文章")
    if err != nil {
        log.Printf("SQLite插入失败: %v", err)
        return
    }

    log.Println("✅ SQLite操作完成")
}

func mongoExample(ctx context.Context) {
    conn, err := db.DB("mongodb")
    if err != nil {
        log.Printf("MongoDB连接失败: %v", err)
        return
    }

    // 获取MongoDB连接
    mongoConn := db.GetMongoConnection(conn)
    if mongoConn == nil {
        log.Println("MongoDB连接转换失败")
        return
    }

    // 获取集合
    collection := mongoConn.GetCollection("articles")
    query := db.NewMongoQuery(collection, nil)

    // 插入文档
    article := bson.M{
        "title":      "MongoDB文章",
        "content":    "这是MongoDB数据库中的文章",
        "created_at": time.Now(),
    }

    _, err = query.InsertOne(ctx, article)
    if err != nil {
        log.Printf("MongoDB插入失败: %v", err)
        return
    }

    log.Println("✅ MongoDB操作完成")
}
```

## 🏃‍♂️ 运行示例

```bash
# 安装依赖
go mod tidy

# 运行程序
go run *.go
```

预期输出：
```
✅ 数据库连接成功!
✅ 数据库迁移完成
✅ 用户插入成功，ID: 1
✅ 查询到 1 个用户
用户: 张三, 邮箱: zhangsan@example.com, 年龄: 25
✅ 多数据库连接配置完成
✅ SQLite操作完成
✅ MongoDB操作完成
```

## 🎯 下一步

恭喜！你已经成功完成了TORM的快速入门。现在你可以：

### 📚 深入学习
- [配置文档](Configuration) - 了解更多配置选项
- [查询构建器](Query-Builder) - 学习强大的查询构建功能
- [模型系统](Model-System) - 使用ActiveRecord模式
- [关联关系](Relationships) - 处理表之间的关系

### 🛠️ 实际应用
- [最佳实践](Best-Practices) - 学习推荐的使用模式
- [性能优化](Performance) - 优化应用性能
- [示例代码](Examples) - 查看更多完整示例

### 📖 参考资料
- [API参考](API-Reference) - 完整的API文档
- [故障排除](Troubleshooting) - 解决常见问题

## 💡 小贴士

1. **开启查询日志**: 在开发环境中设置 `LogQueries: true` 来查看执行的SQL
2. **连接池配置**: 根据应用负载调整 `MaxOpenConns` 和 `MaxIdleConns`
3. **错误处理**: 始终检查并处理错误，特别是数据库操作
4. **上下文使用**: 合理使用 `context.Context` 来控制超时和取消

## ❓ 遇到问题？

- 查看 [故障排除](Troubleshooting) 文档
- 提交 [GitHub Issue](https://github.com/zhoudm1743/torm/issues)
- 发送邮件到 zhoudm1743@163.com

---

**🎉 现在你已经掌握了TORM的基础用法，开始构建你的应用吧！** 