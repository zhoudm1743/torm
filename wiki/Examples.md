# 示例代码

本文档提供了TORM的完整使用示例，涵盖了从基础操作到高级功能的各种场景。

## 🚀 基础示例

### 连接数据库

```go
package main

import (
    "context"
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
    
    ctx := context.Background()
    err = conn.Ping(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Println("数据库连接成功!")
}
```

### CRUD 操作

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
    Name      string    `db:"name" json:"name"`
    Email     string    `db:"email" json:"email"`
    Age       int       `db:"age" json:"age"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
}

func main() {
    // ... 数据库配置 ...
    
    ctx := context.Background()
    conn, _ := db.DB("default")
    
    // 创建表
    createTable(ctx, conn)
    
    // 插入数据
    user := insertUser(ctx, conn, "张三", "zhangsan@example.com", 25)
    
    // 查询单个用户
    foundUser := getUserByID(ctx, conn, user.ID)
    log.Printf("找到用户: %+v", foundUser)
    
    // 查询多个用户
    users := getUsersByAge(ctx, conn, 20, 30)
    log.Printf("找到 %d 个用户", len(users))
    
    // 更新用户
    updateUser(ctx, conn, user.ID, "李四", 26)
    
    // 删除用户
    deleteUser(ctx, conn, user.ID)
}

func createTable(ctx context.Context, conn db.ConnectionInterface) {
    sql := `
    CREATE TABLE IF NOT EXISTS users (
        id BIGINT AUTO_INCREMENT PRIMARY KEY,
        name VARCHAR(100) NOT NULL,
        email VARCHAR(100) UNIQUE NOT NULL,
        age INT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`
    
    _, err := conn.Exec(ctx, sql)
    if err != nil {
        log.Fatal("创建表失败:", err)
    }
}

func insertUser(ctx context.Context, conn db.ConnectionInterface, name, email string, age int) *User {
    sql := `INSERT INTO users (name, email, age) VALUES (?, ?, ?)`
    
    result, err := conn.Exec(ctx, sql, name, email, age)
    if err != nil {
        log.Fatal("插入用户失败:", err)
    }
    
    id, _ := result.LastInsertId()
    
    return &User{
        ID:        id,
        Name:      name,
        Email:     email,
        Age:       age,
        CreatedAt: time.Now(),
    }
}

func getUserByID(ctx context.Context, conn db.ConnectionInterface, id int64) *User {
    sql := `SELECT id, name, email, age, created_at FROM users WHERE id = ?`
    
    row := conn.QueryRow(ctx, sql, id)
    
    user := &User{}
    err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Age, &user.CreatedAt)
    if err != nil {
        log.Fatal("查询用户失败:", err)
    }
    
    return user
}

func getUsersByAge(ctx context.Context, conn db.ConnectionInterface, minAge, maxAge int) []*User {
    sql := `SELECT id, name, email, age, created_at FROM users WHERE age BETWEEN ? AND ? ORDER BY created_at DESC`
    
    rows, err := conn.Query(ctx, sql, minAge, maxAge)
    if err != nil {
        log.Fatal("查询用户失败:", err)
    }
    defer rows.Close()
    
    var users []*User
    for rows.Next() {
        user := &User{}
        err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Age, &user.CreatedAt)
        if err != nil {
            log.Fatal("扫描用户数据失败:", err)
        }
        users = append(users, user)
    }
    
    return users
}

func updateUser(ctx context.Context, conn db.ConnectionInterface, id int64, name string, age int) {
    sql := `UPDATE users SET name = ?, age = ? WHERE id = ?`
    
    _, err := conn.Exec(ctx, sql, name, age, id)
    if err != nil {
        log.Fatal("更新用户失败:", err)
    }
}

func deleteUser(ctx context.Context, conn db.ConnectionInterface, id int64) {
    sql := `DELETE FROM users WHERE id = ?`
    
    _, err := conn.Exec(ctx, sql, id)
    if err != nil {
        log.Fatal("删除用户失败:", err)
    }
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