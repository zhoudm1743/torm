# 示例代码

本文档提供了TORM现代化ORM的完整使用示例，涵盖了从基础操作到高级功能的各种场景。

## 🚀 基础示例

### 连接数据库（现代化方式）

```go
package main

import (
    "log"
    "time"
    
    "torm/pkg/db"
)

func main() {
    config := &db.Config{
        Driver:          "mysql",
        Host:            "localhost",
        Port:            3306,
        Database:        "blog",
        Username:        "root",
        Password:        "password",
        Charset:         "utf8mb4",
        MaxOpenConns:    100,
        MaxIdleConns:    10,
        ConnMaxLifetime: time.Hour,
        LogQueries:      true,
    }
    
    err := db.AddConnection("default", config)
    if err != nil {
        log.Fatal(err)
    }
    
    conn, err := db.DB("default")
    if err != nil {
        log.Fatal(err)
    }
    
    // ✅ 现代化API - 无需context参数
    err = conn.Connect()
    if err != nil {
        log.Fatal(err)
    }
    
    // 可选的超时控制
    // err = conn.Ping() // 默认无超时
    
    log.Println("✅ 数据库连接成功！")
}
```

### 查询构建器基础用法

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "torm/pkg/db"
)

func main() {
    // 配置数据库...（省略）
    
    // ✅ 获取查询构建器 - 简洁的API
    query, err := db.Table("users")
    if err != nil {
        log.Fatal(err)
    }
    
    // 1. 插入数据
    userID, err := query.Insert(map[string]interface{}{
        "name":     "张三",
        "email":    "zhangsan@example.com",
        "age":      28,
        "status":   "active",
        "created_at": time.Now(),
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ 插入成功，用户ID: %v\n", userID)
    
    // 2. 查询数据
    users, err := query.
        Where("status", "=", "active").
        Where("age", ">=", 18).
        OrderBy("created_at", "desc").
        Limit(10).
        Get()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ 找到 %d 个用户\n", len(users))
    
    // 3. 更新数据
    affected, err := query.
        Where("id", "=", userID).
        Update(map[string]interface{}{
            "age": 29,
            "updated_at": time.Now(),
        })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ 更新了 %d 条记录\n", affected)
    
    // 4. 删除数据
    deleted, err := query.
        Where("id", "=", userID).
        Delete()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ 删除了 %d 条记录\n", deleted)
}
```

### 高级查询示例

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "torm/pkg/db"
)

func advancedQueryExamples() {
    // 1. 复杂条件查询
    query, _ := db.Table("users")
    
    results, err := query.
        Select("users.name", "profiles.avatar", "COUNT(posts.id) as post_count").
        LeftJoin("profiles", "users.id", "=", "profiles.user_id").
        LeftJoin("posts", "users.id", "=", "posts.user_id").
        Where("users.status", "=", "active").
        WhereIn("users.role", []interface{}{"admin", "editor"}).
        WhereBetween("users.age", 25, 65).
        WhereNotNull("profiles.avatar").
        GroupBy("users.id").
        Having("post_count", ">", 0).
        OrderBy("post_count", "desc").
        Limit(20).
        Get()
    
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ 复杂查询返回 %d 条记录\n", len(results))
    
    // 2. 聚合查询
    stats, err := query.
        Select(
            "department", 
            "COUNT(*) as user_count", 
            "AVG(age) as avg_age",
            "MAX(salary) as max_salary",
            "MIN(created_at) as earliest_join",
        ).
        Where("status", "=", "active").
        GroupBy("department").
        Having("user_count", ">=", 5).
        OrderBy("avg_age", "desc").
        Get()
    
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ 统计查询返回 %d 个部门\n", len(stats))
    
    // 3. 子查询示例
    subQuery, _ := db.Table("posts")
    subQuerySQL, bindings, _ := subQuery.
        Select("user_id").
        Where("status", "=", "published").
        GroupBy("user_id").
        Having("COUNT(*)", ">", 10).
        ToSQL()
    
    activeWriters, err := query.
        Where("status", "=", "active").
        WhereRaw("id IN ("+subQuerySQL+")", bindings...).
        Get()
    
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ 找到 %d 个活跃作者\n", len(activeWriters))
}
```

### 批量操作示例

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "torm/pkg/db"
)

func batchOperationsExample() {
    query, _ := db.Table("users")
    
    // 1. 批量插入
    users := []map[string]interface{}{
        {
            "name":       "用户1",
            "email":      "user1@example.com",
            "age":        25,
            "status":     "active",
            "created_at": time.Now(),
        },
        {
            "name":       "用户2", 
            "email":      "user2@example.com",
            "age":        30,
            "status":     "active",
            "created_at": time.Now(),
        },
        {
            "name":       "用户3",
            "email":      "user3@example.com", 
            "age":        35,
            "status":     "pending",
            "created_at": time.Now(),
        },
    }
    
    affected, err := query.InsertBatch(users)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ 批量插入 %d 条记录\n", affected)
    
    // 2. 批量更新
    affected, err = query.
        Where("status", "=", "pending").
        Update(map[string]interface{}{
            "status":     "active",
            "updated_at": time.Now(),
        })
    
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ 批量更新 %d 条记录\n", affected)
    
    // 3. 条件删除
    affected, err = query.
        Where("status", "=", "inactive").
        Where("last_login", "<", time.Now().AddDate(0, -6, 0)). // 6个月未登录
        Delete()
    
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ 清理了 %d 个非活跃用户\n", affected)
}
```

### 事务处理示例

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "torm/pkg/db"
)

func transactionExample() {
    // ✅ 现代化事务API - 无需context
    err := db.Transaction(func(tx db.TransactionInterface) error {
        // 1. 创建用户
        userResult, err := tx.Exec(`
            INSERT INTO users (name, email, age, status, created_at) 
            VALUES (?, ?, ?, ?, ?)
        `, "事务用户", "transaction@example.com", 28, "active", time.Now())
        
        if err != nil {
            return err // 自动回滚
        }
        
        userID, err := userResult.LastInsertId()
        if err != nil {
            return err
        }
        
        // 2. 创建用户资料
        _, err = tx.Exec(`
            INSERT INTO profiles (user_id, avatar, bio, created_at) 
            VALUES (?, ?, ?, ?)
        `, userID, "default-avatar.png", "新用户", time.Now())
        
        if err != nil {
            return err // 自动回滚
        }
        
        // 3. 记录操作日志
        _, err = tx.Exec(`
            INSERT INTO user_logs (user_id, action, details, created_at) 
            VALUES (?, ?, ?, ?)
        `, userID, "user_created", "用户注册", time.Now())
        
        if err != nil {
            return err // 自动回滚
        }
        
        fmt.Printf("✅ 事务中创建用户，ID: %d\n", userID)
        return nil // 自动提交
    })
    
    if err != nil {
        log.Printf("❌ 事务失败: %v", err)
        return
    }
    
    fmt.Println("✅ 事务执行成功！")
}

// 复杂事务示例：银行转账
func bankTransferExample() {
    err := db.Transaction(func(tx db.TransactionInterface) error {
        // 1. 检查转出账户余额
        var fromBalance float64
        err := tx.QueryRow(`
            SELECT balance FROM accounts WHERE id = ? FOR UPDATE
        `, 1).Scan(&fromBalance)
        
        if err != nil {
            return fmt.Errorf("查询转出账户失败: %v", err)
        }
        
        transferAmount := 1000.0
        if fromBalance < transferAmount {
            return fmt.Errorf("余额不足，当前余额: %.2f", fromBalance)
        }
        
        // 2. 扣除转出账户余额
        _, err = tx.Exec(`
            UPDATE accounts SET balance = balance - ?, updated_at = ? 
            WHERE id = ?
        `, transferAmount, time.Now(), 1)
        
        if err != nil {
            return fmt.Errorf("扣款失败: %v", err)
        }
        
        // 3. 增加转入账户余额
        _, err = tx.Exec(`
            UPDATE accounts SET balance = balance + ?, updated_at = ? 
            WHERE id = ?
        `, transferAmount, time.Now(), 2)
        
        if err != nil {
            return fmt.Errorf("入账失败: %v", err)
        }
        
        // 4. 记录转账日志
        _, err = tx.Exec(`
            INSERT INTO transfer_logs (from_account, to_account, amount, status, created_at) 
            VALUES (?, ?, ?, ?, ?)
        `, 1, 2, transferAmount, "completed", time.Now())
        
        if err != nil {
            return fmt.Errorf("记录日志失败: %v", err)
        }
        
        fmt.Printf("✅ 转账成功: %.2f 元\n", transferAmount)
        return nil
    })
    
    if err != nil {
        log.Printf("❌ 转账失败: %v", err)
        return
    }
    
    fmt.Println("✅ 转账事务完成！")
}
```

### 超时控制示例

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "torm/pkg/db"
)

func timeoutControlExample() {
    query, _ := db.Table("users")
    
    // 1. 使用WithTimeout进行超时控制
    users, err := query.
        WithTimeout(5 * time.Second).  // 5秒超时
        Where("status", "=", "active").
        OrderBy("created_at", "desc").
        Limit(100).
        Get()
    
    if err != nil {
        log.Printf("❌ 查询超时: %v", err)
        return
    }
    fmt.Printf("✅ 在5秒内查询到 %d 个用户\n", len(users))
    
    // 2. 使用WithContext进行更精细的控制
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    largeDataset, err := query.
        WithContext(ctx).
        Select("*").
        OrderBy("id", "asc").
        Get()
    
    if err != nil {
        log.Printf("❌ 大数据查询失败: %v", err)
        return
    }
    fmt.Printf("✅ 查询大数据集: %d 条记录\n", len(largeDataset))
    
    // 3. 长时间运行的操作超时控制
    longRunningCtx, longCancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer longCancel()
    
    err = db.Transaction(func(tx db.TransactionInterface) error {
        // 在事务内部也可以检查context状态
        select {
        case <-longRunningCtx.Done():
            return longRunningCtx.Err()
        default:
        }
        
        // 执行长时间操作...
        result, err := tx.Exec(`
            UPDATE users SET status = 'verified' 
            WHERE email_verified = 1 AND status = 'pending'
        `)
        if err != nil {
            return err
        }
        
        affected, _ := result.RowsAffected()
        fmt.Printf("✅ 批量验证了 %d 个用户\n", affected)
        
        return nil
    })
    
    if err != nil {
        log.Printf("❌ 长时间操作失败: %v", err)
        return
    }
    
    fmt.Println("✅ 长时间操作完成！")
}
```

## 📦 数据迁移示例

```go
package main

import (
    "context"
    "log"
    
    "torm/pkg/db"
    "torm/pkg/migration"
)

func main() {
    // ... 数据库配置 ...
    
    conn, _ := db.DB("default")
    migrator := migration.NewMigrator(conn, nil)
    
    // 注册迁移
    registerMigrations(migrator)
    
    ctx := context.Background()
    
    // 执行迁移
    if err := migrator.Up(ctx); err != nil {
        log.Fatal("迁移失败:", err)
    }
    
    // 显示状态
    migrator.PrintStatus(ctx)
}

func registerMigrations(migrator *migration.Migrator) {
    // 创建用户表
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
    
    // 创建文章表
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

## 🔄 事务处理示例

```go
package main

import (
    "context"
    "log"
    
    "torm/pkg/db"
)

func main() {
    // ... 数据库配置 ...
    
    ctx := context.Background()
    conn, _ := db.DB("default")
    
    // 简单事务
    simpleTransaction(ctx, conn)
    
    // 复杂事务
    complexTransaction(ctx, conn)
}

func simpleTransaction(ctx context.Context, conn db.ConnectionInterface) {
    tx, err := conn.Begin(ctx)
    if err != nil {
        log.Fatal("开始事务失败:", err)
    }
    defer tx.Rollback() // 确保回滚
    
    // 执行操作
    _, err = tx.Exec(ctx, "INSERT INTO users (name, email, age) VALUES (?, ?, ?)", "事务用户", "tx@example.com", 30)
    if err != nil {
        log.Fatal("插入失败:", err)
    }
    
    // 提交事务
    if err = tx.Commit(); err != nil {
        log.Fatal("提交事务失败:", err)
    }
    
    log.Println("简单事务完成")
}

func complexTransaction(ctx context.Context, conn db.ConnectionInterface) {
    tx, err := conn.Begin(ctx)
    if err != nil {
        log.Fatal("开始事务失败:", err)
    }
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            log.Printf("事务回滚: %v", r)
        }
    }()
    
    // 插入用户
    result, err := tx.Exec(ctx, "INSERT INTO users (name, email, age) VALUES (?, ?, ?)", "作者", "author@example.com", 35)
    if err != nil {
        tx.Rollback()
        log.Fatal("插入用户失败:", err)
    }
    
    userID, _ := result.LastInsertId()
    
    // 插入文章
    _, err = tx.Exec(ctx, "INSERT INTO posts (title, content, user_id) VALUES (?, ?, ?)", "事务文章", "这是在事务中创建的文章", userID)
    if err != nil {
        tx.Rollback()
        log.Fatal("插入文章失败:", err)
    }
    
    // 提交事务
    if err = tx.Commit(); err != nil {
        log.Fatal("提交事务失败:", err)
    }
    
    log.Println("复杂事务完成")
}
```

## 🗄️ 多数据库示例

```go
package main

import (
    "context"
    "log"
    "time"
    
    "go.mongodb.org/mongo-driver/bson"
    "torm/pkg/db"
)

func main() {
    // 配置多个数据库
    setupDatabases()
    
    ctx := context.Background()
    
    // MySQL 操作
    mysqlOperations(ctx)
    
    // MongoDB 操作
    mongodbOperations(ctx)
    
    // SQLite 操作
    sqliteOperations(ctx)
}

func setupDatabases() {
    // MySQL 连接
    mysqlConfig := &db.Config{
        Driver:   "mysql",
        Host:     "localhost",
        Port:     3306,
        Database: "mysql_db",
        Username: "root",
        Password: "password",
    }
    db.AddConnection("mysql", mysqlConfig)
    
    // MongoDB 连接
    mongoConfig := &db.Config{
        Driver:   "mongodb",
        Host:     "localhost",
        Port:     27017,
        Database: "mongo_db",
    }
    db.AddConnection("mongodb", mongoConfig)
    
    // SQLite 连接
    sqliteConfig := &db.Config{
        Driver:   "sqlite",
        Database: "sqlite_db.db",
    }
    db.AddConnection("sqlite", sqliteConfig)
}

func mysqlOperations(ctx context.Context) {
    conn, err := db.DB("mysql")
    if err != nil {
        log.Printf("MySQL连接失败: %v", err)
        return
    }
    
    // 创建表
    _, err = conn.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS mysql_users (
            id BIGINT AUTO_INCREMENT PRIMARY KEY,
            name VARCHAR(100),
            email VARCHAR(100),
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)
    if err != nil {
        log.Printf("MySQL创建表失败: %v", err)
        return
    }
    
    // 插入数据
    _, err = conn.Exec(ctx, "INSERT INTO mysql_users (name, email) VALUES (?, ?)", "MySQL用户", "mysql@example.com")
    if err != nil {
        log.Printf("MySQL插入失败: %v", err)
        return
    }
    
    log.Println("MySQL操作完成")
}

func mongodbOperations(ctx context.Context) {
    conn, err := db.DB("mongodb")
    if err != nil {
        log.Printf("MongoDB连接失败: %v", err)
        return
    }
    
    mongoConn := db.GetMongoConnection(conn)
    if mongoConn == nil {
        log.Println("MongoDB连接转换失败")
        return
    }
    
    collection := mongoConn.GetCollection("users")
    query := db.NewMongoQuery(collection, nil)
    
    // 插入文档
    user := bson.M{
        "name":       "MongoDB用户",
        "email":      "mongo@example.com",
        "created_at": time.Now(),
    }
    
    _, err = query.InsertOne(ctx, user)
    if err != nil {
        log.Printf("MongoDB插入失败: %v", err)
        return
    }
    
    // 查询文档
    count, err := query.Count(ctx)
    if err != nil {
        log.Printf("MongoDB查询失败: %v", err)
        return
    }
    
    log.Printf("MongoDB操作完成，文档数量: %d", count)
}

func sqliteOperations(ctx context.Context) {
    conn, err := db.DB("sqlite")
    if err != nil {
        log.Printf("SQLite连接失败: %v", err)
        return
    }
    
    // 创建表
    _, err = conn.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS sqlite_users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT,
            email TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `)
    if err != nil {
        log.Printf("SQLite创建表失败: %v", err)
        return
    }
    
    // 插入数据
    _, err = conn.Exec(ctx, "INSERT INTO sqlite_users (name, email) VALUES (?, ?)", "SQLite用户", "sqlite@example.com")
    if err != nil {
        log.Printf("SQLite插入失败: %v", err)
        return
    }
    
    log.Println("SQLite操作完成")
}
```

## 🎯 实际应用示例

### 博客系统

```go
package main

import (
    "context"
    "log"
    "time"
    
    "torm/pkg/db"
)

type User struct {
    ID        int64     `db:"id" json:"id"`
    Username  string    `db:"username" json:"username"`
    Email     string    `db:"email" json:"email"`
    Password  string    `db:"password" json:"-"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Post struct {
    ID        int64     `db:"id" json:"id"`
    Title     string    `db:"title" json:"title"`
    Content   string    `db:"content" json:"content"`
    UserID    int64     `db:"user_id" json:"user_id"`
    Status    string    `db:"status" json:"status"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
    UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type BlogService struct {
    conn db.ConnectionInterface
}

func NewBlogService(conn db.ConnectionInterface) *BlogService {
    return &BlogService{conn: conn}
}

func (s *BlogService) CreateUser(ctx context.Context, username, email, password string) (*User, error) {
    sql := `INSERT INTO users (username, email, password) VALUES (?, ?, ?)`
    
    result, err := s.conn.Exec(ctx, sql, username, email, password)
    if err != nil {
        return nil, err
    }
    
    id, _ := result.LastInsertId()
    
    return &User{
        ID:        id,
        Username:  username,
        Email:     email,
        CreatedAt: time.Now(),
    }, nil
}

func (s *BlogService) GetUserByEmail(ctx context.Context, email string) (*User, error) {
    sql := `SELECT id, username, email, created_at FROM users WHERE email = ?`
    
    row := s.conn.QueryRow(ctx, sql, email)
    
    user := &User{}
    err := row.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
    if err != nil {
        return nil, err
    }
    
    return user, nil
}

func (s *BlogService) CreatePost(ctx context.Context, title, content string, userID int64) (*Post, error) {
    sql := `INSERT INTO posts (title, content, user_id, status, updated_at) VALUES (?, ?, ?, 'draft', CURRENT_TIMESTAMP)`
    
    result, err := s.conn.Exec(ctx, sql, title, content, userID)
    if err != nil {
        return nil, err
    }
    
    id, _ := result.LastInsertId()
    
    return &Post{
        ID:        id,
        Title:     title,
        Content:   content,
        UserID:    userID,
        Status:    "draft",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }, nil
}

func (s *BlogService) PublishPost(ctx context.Context, postID int64) error {
    sql := `UPDATE posts SET status = 'published', updated_at = CURRENT_TIMESTAMP WHERE id = ?`
    
    _, err := s.conn.Exec(ctx, sql, postID)
    return err
}

func (s *BlogService) GetPostsByUser(ctx context.Context, userID int64) ([]*Post, error) {
    sql := `SELECT id, title, content, user_id, status, created_at, updated_at FROM posts WHERE user_id = ? ORDER BY created_at DESC`
    
    rows, err := s.conn.Query(ctx, sql, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var posts []*Post
    for rows.Next() {
        post := &Post{}
        err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.UserID, &post.Status, &post.CreatedAt, &post.UpdatedAt)
        if err != nil {
            return nil, err
        }
        posts = append(posts, post)
    }
    
    return posts, nil
}

func main() {
    // ... 数据库配置 ...
    
    conn, _ := db.DB("default")
    blogService := NewBlogService(conn)
    
    ctx := context.Background()
    
    // 创建用户
    user, err := blogService.CreateUser(ctx, "johndoe", "john@example.com", "password123")
    if err != nil {
        log.Fatal("创建用户失败:", err)
    }
    log.Printf("创建用户: %+v", user)
    
    // 创建文章
    post, err := blogService.CreatePost(ctx, "我的第一篇文章", "这是文章内容...", user.ID)
    if err != nil {
        log.Fatal("创建文章失败:", err)
    }
    log.Printf("创建文章: %+v", post)
    
    // 发布文章
    err = blogService.PublishPost(ctx, post.ID)
    if err != nil {
        log.Fatal("发布文章失败:", err)
    }
    
    // 获取用户的文章
    posts, err := blogService.GetPostsByUser(ctx, user.ID)
    if err != nil {
        log.Fatal("获取文章失败:", err)
    }
    log.Printf("用户文章数量: %d", len(posts))
}
```

---

**📚 更多示例请参考 [GitHub仓库](https://github.com/zhoudm1743/torm) 中的 examples 目录。** 