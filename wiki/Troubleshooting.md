# 故障排除

本文档帮助您解决使用TORM时可能遇到的常见问题。

## 🔧 连接问题

### MySQL连接失败

**错误**: `Error 1045: Access denied for user 'root'@'localhost'`

**解决方案**:
```go
// 检查用户名和密码
config := &db.Config{
    Driver:   "mysql",
    Username: "root",        // 确认用户名正确
    Password: "your_password", // 确认密码正确
}

// 或使用环境变量
config.Password = os.Getenv("DB_PASSWORD")
```

### PostgreSQL SSL问题

**错误**: `pq: SSL is not enabled on the server`

**解决方案**:
```go
config := &db.Config{
    Driver:  "postgres",
    SSLMode: "disable", // 禁用SSL
}
```

### MongoDB连接超时

**错误**: `context deadline exceeded`

**解决方案**:
```go
config := &db.Config{
    Driver: "mongodb",
    Options: map[string]string{
        "connectTimeoutMS": "30000", // 增加超时时间
        "serverSelectionTimeoutMS": "30000",
    },
}
```

## 📦 迁移问题

### 迁移表创建失败

**错误**: `Table 'migrations' doesn't exist`

**解决方案**:
```go
// 确保自动创建迁移表
migrator := migration.NewMigrator(conn, logger)
migrator.SetAutoCreate(true) // 默认为true
```

### SQLite UNIQUE列问题

**错误**: `Cannot add a UNIQUE column`

**解决方案**:
TORM会自动处理，或手动分步操作：
```go
// 先添加普通列
schema.AddColumn(ctx, "users", &migration.Column{
    Name: "email",
    Type: migration.ColumnTypeVarchar,
    Length: 100,
})

// 再创建UNIQUE索引
schema.CreateIndex(ctx, "users", &migration.Index{
    Name: "idx_users_email",
    Columns: []string{"email"},
    Unique: true,
})
```

## 🚀 性能问题

### 连接池耗尽

**错误**: `too many connections`

**解决方案**:
```go
config := &db.Config{
    MaxOpenConns: 100,  // 增加最大连接数
    MaxIdleConns: 20,   // 适当增加空闲连接
    ConnMaxLifetime: time.Hour, // 设置连接生存时间
}
```

### 查询慢

**解决方案**:
1. 启用查询日志分析
```go
config.LogQueries = true
```

2. 添加适当索引
3. 优化查询语句

## 💾 MongoDB特定问题

### 事务失败

**错误**: `Transaction numbers are only allowed on a replica set member`

**解决方案**:
MongoDB事务需要副本集，单机模式不支持：
```bash
# 启动副本集
mongod --replSet rs0

# 初始化副本集
mongo --eval "rs.initiate()"
```

### 集合不存在

**解决方案**:
MongoDB集合会自动创建：
```go
collection := mongoConn.GetCollection("users")
// 集合在第一次插入时自动创建
```

## 🔍 调试技巧

### 启用详细日志

```go
// 1. 启用查询日志
config.LogQueries = true

// 2. 使用自定义日志器
logger := logrus.New()
logger.SetLevel(logrus.DebugLevel)
migrator := migration.NewMigrator(conn, logger)
```

### 查看连接状态

```go
stats := conn.GetStats()
fmt.Printf("打开连接: %d\n", stats.OpenConnections)
fmt.Printf("使用中连接: %d\n", stats.InUse)
fmt.Printf("空闲连接: %d\n", stats.Idle)
```

## 📞 获取帮助

如果问题仍未解决：

1. 查看 [GitHub Issues](https://github.com/zhoudm1743/torm/issues)
2. 提交新的 Issue 并附上详细信息
3. 发送邮件到 zhoudm1743@163.com
4. 加入我们的讨论群

---

**💡 提示**: 大部分问题都与配置相关，请仔细检查数据库连接配置。 