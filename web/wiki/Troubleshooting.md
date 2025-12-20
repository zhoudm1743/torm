# 故障排除

本文档帮助你解决使用 TORM 时可能遇到的常见问题。

## 🔧 连接问题

### MySQL 连接失败

**错误**: `Error 1045: Access denied for user 'root'@'localhost'`

**解决方案**:
```go
// 检查用户名和密码
config := &torm.Config{
    Driver:   "mysql",
    Host:     "localhost",
    Port:     3306,
    Username: "root",        // 确认用户名正确
    Password: "your_password", // 确认密码正确
    Database: "myapp",
}
```

**错误**: `dial tcp: connect: connection refused`

**解决方案**:
```go
// 检查主机和端口
config := &torm.Config{
    Host: "localhost", // 或 "127.0.0.1"
    Port: 3306,        // 确认 MySQL 端口
}

// 检查 MySQL 服务是否运行
// Linux: systemctl status mysql
// macOS: brew services list
// Windows: 检查服务管理器
```

### PostgreSQL SSL 问题

**错误**: `pq: SSL is not enabled on the server`

**解决方案**:
```go
config := &torm.Config{
    Driver:  "postgres",
    Host:    "localhost",
    Port:    5432,
    SSLMode: "disable", // 开发环境禁用 SSL
}
```

### SQLite 文件权限问题

**错误**: `unable to open database file`

**解决方案**:
```go
// 确保目录存在且有写权限
config := &torm.Config{
    Driver:   "sqlite",
    Database: "/path/to/app.db", // 使用绝对路径
}

// 或使用内存数据库测试
config := &torm.Config{
    Driver:   "sqlite",
    Database: ":memory:",
}
```

## 📦 查询问题

### 查询返回空结果

**可能原因**:
1. 条件不匹配
2. 表名错误
3. 连接名错误

**调试方法**:
```go
// 启用 SQL 日志查看实际执行的 SQL
torm.EnableSQLLogging()

// 获取 SQL 语句
query := torm.Table("users").Where("status", "=", "active")
sql, args := query.ToSQL()
fmt.Printf("SQL: %s\n", sql)
fmt.Printf("Args: %v\n", args)

// 检查表是否存在
exists, _ := torm.Table("users").Exists()
fmt.Printf("表存在: %v\n", exists)
```

### 参数绑定错误

**错误**: `sql: expected X arguments, got Y`

**解决方案**:
```go
// ❌ 错误：参数数量不匹配
query.WhereRaw("status = ? AND age > ?", "active") // 缺少一个参数

// ✅ 正确：参数数量匹配
query.WhereRaw("status = ? AND age > ?", "active", 18)

// ✅ 推荐：使用链式方法
query.Where("status", "=", "active").Where("age", ">", 18)
```

### IN 查询问题

**错误**: 数组参数未正确展开

**解决方案**:
```go
// ✅ 使用 WhereIn 方法
query.WhereIn("status", []interface{}{"active", "pending"})

// ✅ 或使用数组参数
query.Where("status IN (?)", []string{"active", "pending"})
```

## 💼 事务问题

### 事务未提交

**可能原因**: 事务函数返回了错误

**解决方案**:
```go
err := torm.Transaction(func(tx torm.TransactionInterface) error {
    _, err := tx.Exec("INSERT INTO users (name) VALUES (?)", "张三")
    if err != nil {
        return err // 返回错误会自动回滚
    }
    return nil // 返回 nil 会自动提交
})

if err != nil {
    log.Printf("事务失败: %v", err)
}
```

### 嵌套事务问题

**说明**: TORM 支持嵌套事务（使用 Savepoint）

```go
err := torm.Transaction(func(tx torm.TransactionInterface) error {
    // 外层事务
    _, err := tx.Exec("INSERT INTO users (name) VALUES (?)", "用户1")
    if err != nil {
        return err
    }
    
    // 嵌套事务会创建 Savepoint
    err = torm.Transaction(func(innerTx torm.TransactionInterface) error {
        _, err := innerTx.Exec("INSERT INTO logs (message) VALUES (?)", "日志")
        return err
    })
    
    return err
})
```

## 🔄 迁移问题

### 迁移表创建失败

**错误**: `Table 'migrations' doesn't exist`

**解决方案**:
```go
// 迁移器会自动创建迁移表
migrator := migration.NewMigrator(conn, nil)
migrator.SetAutoCreate(true) // 默认为 true
```

### 迁移版本冲突

**解决方案**:
```go
// 查看迁移状态
status, _ := migrator.Status()
for _, s := range status {
    fmt.Printf("版本: %s, 已应用: %v\n", s.Version, s.Applied)
}

// 回滚到特定版本
migrator.Down(1) // 回滚 1 步
```

## 🚀 性能问题

### 连接池耗尽

**错误**: `too many connections`

**解决方案**:
```go
config := &torm.Config{
    MaxOpenConns:    100,  // 增加最大连接数
    MaxIdleConns:    20,   // 适当增加空闲连接
    ConnMaxLifetime: time.Hour,
}
```

### 查询慢

**调试方法**:
```go
// 1. 启用查询日志
torm.EnableSQLLogging()

// 2. 检查是否使用了索引
// 查看日志中的 SQL，确认 WHERE 条件字段有索引

// 3. 避免 N+1 查询
// 使用预加载或 JOIN

// 4. 使用 EXPLAIN 分析
// 在数据库中执行 EXPLAIN SELECT ...
```

## 🔍 调试技巧

### 启用详细日志

```go
// 启用 SQL 日志
torm.EnableSQLLogging()

// 设置日志级别
import "github.com/zhoudm1743/torm/logger"
torm.SetSQLLogging(logger.DEBUG, true)
```

### 检查连接状态

```go
// 测试连接
conn, err := torm.DB("default")
if err != nil {
    log.Fatal("获取连接失败:", err)
}

ctx := context.Background()
err = conn.Ping(ctx)
if err != nil {
    log.Fatal("连接测试失败:", err)
}
```

### 获取版本信息

```go
fmt.Printf("TORM 版本: %s\n", torm.GetVersion())
```

## 📞 获取帮助

如果问题仍未解决：

1. 查看 [GitHub Issues](https://github.com/zhoudm1743/torm/issues)
2. 提交新的 Issue 并附上：
   - 错误信息
   - 代码示例
   - Go 版本
   - 数据库版本
3. 发送邮件到 zhoudm1743@163.com

## 🔗 相关文档

- [配置文档](Configuration) - 配置选项
- [日志系统](Logging) - 日志配置
- [性能优化](Performance) - 性能问题排查
