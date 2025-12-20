# API 参考

TORM API 完整参考文档，包含所有核心接口、方法和参数说明。

## 📋 目录

- [全局函数](#全局函数)
- [数据库连接](#数据库连接)
- [查询构建器](#查询构建器)
- [模型接口](#模型接口)
- [事务接口](#事务接口)
- [缓存接口](#缓存接口)
- [迁移接口](#迁移接口)

## 🌐 全局函数

### torm 包入口函数

```go
import "github.com/zhoudm1743/torm"

// 添加数据库连接
func AddConnection(name string, config *Config) error

// 获取查询构建器
func Table(table string, connectionName ...string) (*QueryBuilder, error)

// 获取数据库连接
func DB(name ...string) (Connection, error)

// 获取版本号
func GetVersion() string

// 事务处理
func Transaction(fn func(tx TransactionInterface) error, connectionName ...string) error
func BeginTransaction(connectionName ...string) (*DBTransaction, error)

// 日志控制
func EnableSQLLogging()
func SetSQLLogging(level int, enabled bool)
func SetLogger(logger LoggerInterface)

// 关闭连接
func Close(name ...string) error
func CloseAll() error
```

## 🔌 数据库连接

### Config 配置结构体

```go
type Config struct {
    Driver          string        // 数据库驱动: "mysql", "postgres", "sqlite"
    Host            string        // 主机地址
    Port            int           // 端口号
    Database        string        // 数据库名
    Username        string        // 用户名
    Password        string        // 密码
    Charset         string        // 字符集，默认 "utf8mb4"
    
    // 连接池配置
    MaxOpenConns    int           // 最大打开连接数，默认 100
    MaxIdleConns    int           // 最大空闲连接数，默认 10
    ConnMaxLifetime time.Duration // 连接最大生存时间，默认 1小时
    ConnMaxIdleTime time.Duration // 连接最大空闲时间，默认 30分钟
    
    // SSL配置 (PostgreSQL)
    SSLMode         string        // SSL模式: "disable", "require", "verify-ca"
    
    // 日志配置
    LogQueries      bool          // 是否记录查询日志
}
```

### 连接管理

```go
// 添加数据库连接
err := torm.AddConnection("default", &torm.Config{
    Driver:   "mysql",
    Host:     "localhost",
    Port:     3306,
    Database: "myapp",
    Username: "root",
    Password: "password",
})

// 获取数据库连接
conn, err := torm.DB("default")

// 测试连接
err = conn.Ping(ctx)

// 关闭连接
err = torm.Close("default")

// 关闭所有连接
err = torm.CloseAll()
```

## 🔍 查询构建器

### 创建查询

```go
// 方式1：使用全局函数
query, err := torm.Table("users")
query, err := torm.Table("users", "mysql_conn")

// 方式2：使用连接创建
conn, _ := torm.DB("default")
query := conn.Table("users")
```

### 字段选择

```go
func (q *QueryBuilder) Select(columns ...interface{}) *QueryBuilder
func (q *QueryBuilder) SelectRaw(expression string, bindings ...interface{}) *QueryBuilder
func (q *QueryBuilder) Distinct() *QueryBuilder
```

```go
// 示例
query.Select("id", "name", "email")
query.Select("users.id", "users.name", "profiles.avatar")
query.SelectRaw("COUNT(*) as total")
query.Distinct()
```

### WHERE 条件

```go
// 基础条件
func (q *QueryBuilder) Where(args ...interface{}) *QueryBuilder
func (q *QueryBuilder) OrWhere(args ...interface{}) *QueryBuilder
func (q *QueryBuilder) WhereRaw(expression string, bindings ...interface{}) *QueryBuilder

// 特殊条件
func (q *QueryBuilder) WhereIn(column string, values []interface{}) *QueryBuilder
func (q *QueryBuilder) WhereNotIn(column string, values []interface{}) *QueryBuilder
func (q *QueryBuilder) WhereBetween(column string, values []interface{}) *QueryBuilder
func (q *QueryBuilder) WhereNotBetween(column string, values []interface{}) *QueryBuilder
func (q *QueryBuilder) WhereNull(column string) *QueryBuilder
func (q *QueryBuilder) WhereNotNull(column string) *QueryBuilder
func (q *QueryBuilder) WhereLike(column string, pattern string) *QueryBuilder
func (q *QueryBuilder) WhereNotLike(column string, pattern string) *QueryBuilder
func (q *QueryBuilder) WhereExists(subQuery string, bindings ...interface{}) *QueryBuilder
func (q *QueryBuilder) WhereNotExists(subQuery string, bindings ...interface{}) *QueryBuilder
```

```go
// 示例
query.Where("status", "=", "active")
query.Where("age", ">=", 18)
query.OrWhere("vip_level", ">", 3)
query.WhereIn("status", []interface{}{"active", "pending"})
query.WhereBetween("age", []interface{}{18, 65})
query.WhereNull("deleted_at")
query.WhereNotNull("email")
query.WhereLike("name", "%张%")
```

### JOIN 查询

```go
func (q *QueryBuilder) Join(table, first, operator, second string) *QueryBuilder
func (q *QueryBuilder) LeftJoin(table, first, operator, second string) *QueryBuilder
func (q *QueryBuilder) RightJoin(table, first, operator, second string) *QueryBuilder
func (q *QueryBuilder) CrossJoin(table string) *QueryBuilder
```

```go
// 示例
query.Join("profiles", "users.id", "=", "profiles.user_id")
query.LeftJoin("orders", "users.id", "=", "orders.user_id")
```

### 排序和分组

```go
func (q *QueryBuilder) OrderBy(column, direction string) *QueryBuilder
func (q *QueryBuilder) OrderByRaw(expression string, bindings ...interface{}) *QueryBuilder
func (q *QueryBuilder) OrderRand() *QueryBuilder
func (q *QueryBuilder) OrderField(column string, values []interface{}, direction string) *QueryBuilder
func (q *QueryBuilder) GroupBy(columns ...string) *QueryBuilder
func (q *QueryBuilder) Having(args ...interface{}) *QueryBuilder
func (q *QueryBuilder) HavingRaw(expression string, bindings ...interface{}) *QueryBuilder
```

```go
// 示例
query.OrderBy("created_at", "desc")
query.OrderRand() // 随机排序，自动适配数据库
query.OrderField("status", []interface{}{"premium", "active", "trial"}, "asc")
query.GroupBy("category")
query.Having("COUNT(*)", ">", 5)
```

### 限制和分页

```go
func (q *QueryBuilder) Limit(limit int) *QueryBuilder
func (q *QueryBuilder) Offset(offset int) *QueryBuilder
func (q *QueryBuilder) Page(page, size int) *QueryBuilder
func (q *QueryBuilder) Paginate(page, perPage int) (*Pagination, error)
```

```go
// 示例
query.Limit(10)
query.Offset(20)
query.Page(2, 10) // 第2页，每页10条
result, err := query.Paginate(1, 20) // 分页查询
```

### 执行查询

```go
// 获取多条记录
func (q *QueryBuilder) Get() ([]map[string]interface{}, error)

// 获取单条记录
func (q *QueryBuilder) First() (map[string]interface{}, error)

// 根据主键查找
func (q *QueryBuilder) Find(id interface{}) (map[string]interface{}, error)

// 获取单列值
func (q *QueryBuilder) Pluck(column string) ([]interface{}, error)
func (q *QueryBuilder) Value(column string) (interface{}, error)

// 检查是否存在
func (q *QueryBuilder) Exists() (bool, error)
```

### 聚合查询

```go
func (q *QueryBuilder) Count() (int64, error)
func (q *QueryBuilder) Sum(column string) (interface{}, error)
func (q *QueryBuilder) Avg(column string) (interface{}, error)
func (q *QueryBuilder) Max(column string) (interface{}, error)
func (q *QueryBuilder) Min(column string) (interface{}, error)
```

### 插入数据

```go
func (q *QueryBuilder) Insert(data map[string]interface{}) (int64, error)
func (q *QueryBuilder) InsertBatch(data []map[string]interface{}) (int64, error)
func (q *QueryBuilder) InsertGetId(data map[string]interface{}) (int64, error)
```

```go
// 示例
id, err := query.Insert(map[string]interface{}{
    "name":  "张三",
    "email": "zhangsan@example.com",
})

// 批量插入
count, err := query.InsertBatch([]map[string]interface{}{
    {"name": "用户1", "email": "user1@example.com"},
    {"name": "用户2", "email": "user2@example.com"},
})
```

### 更新数据

```go
func (q *QueryBuilder) Update(data map[string]interface{}) (int64, error)
func (q *QueryBuilder) Increment(column string, amount interface{}) (int64, error)
func (q *QueryBuilder) Decrement(column string, amount interface{}) (int64, error)
```

```go
// 示例
affected, err := query.Where("id", "=", 1).Update(map[string]interface{}{
    "status": "inactive",
})

affected, err := query.Where("id", "=", 1).Increment("views", 1)
```

### 删除数据

```go
func (q *QueryBuilder) Delete() (int64, error)
func (q *QueryBuilder) Truncate() error
```

### 缓存

```go
func (q *QueryBuilder) Cache(duration time.Duration) *QueryBuilder
func (q *QueryBuilder) CacheWithTags(duration time.Duration, tags ...string) *QueryBuilder
func (q *QueryBuilder) CacheKey(key string) *QueryBuilder
```

```go
// 示例
users, err := query.Where("status", "=", "active").
    Cache(5 * time.Minute).
    Get()
```

### 调试

```go
func (q *QueryBuilder) ToSQL() (string, []interface{})
func (q *QueryBuilder) Clone() *QueryBuilder
```

### 分页结果结构

```go
type Pagination struct {
    Data        []map[string]interface{} `json:"data"`
    Total       int64                    `json:"total"`
    PerPage     int                      `json:"per_page"`
    CurrentPage int                      `json:"current_page"`
    LastPage    int                      `json:"last_page"`
    From        int                      `json:"from"`
    To          int                      `json:"to"`
}
```

## 📊 模型接口

### BaseModel 方法

```go
// 创建模型
func NewModel(args ...interface{}) *BaseModel

// 表配置
func (m *BaseModel) SetTable(table string) *BaseModel
func (m *BaseModel) GetTableName() string
func (m *BaseModel) SetPrimaryKey(key string) *BaseModel
func (m *BaseModel) GetPrimaryKey() string
func (m *BaseModel) SetConnection(connection string) *BaseModel

// 属性操作
func (m *BaseModel) SetAttribute(key string, value interface{}) Model
func (m *BaseModel) GetAttribute(key string) interface{}
func (m *BaseModel) Fill(data map[string]interface{}) *BaseModel
func (m *BaseModel) GetAttributes() map[string]interface{}

// 状态方法
func (m *BaseModel) IsNew() bool
func (m *BaseModel) IsDirty() bool
func (m *BaseModel) GetDirty() map[string]interface{}

// 持久化操作
func (m *BaseModel) Save() error
func (m *BaseModel) Delete() error
func (m *BaseModel) Find(pk interface{}) error

// 软删除
func (m *BaseModel) SoftDelete() error
func (m *BaseModel) Restore() error
func (m *BaseModel) ForceDelete() error
func (m *BaseModel) EnableSoftDeletes() *BaseModel

// 时间戳
func (m *BaseModel) EnableTimestamps() Model
func (m *BaseModel) DisableTimestamps() Model

// 查询方法
func (m *BaseModel) Query() (*QueryBuilder, error)
func (m *BaseModel) Where(args ...interface{}) (*QueryBuilder, error)

// 关联方法
func (m *BaseModel) HasOne(model interface{}, foreignKey, localKey string) *HasOne
func (m *BaseModel) HasMany(model interface{}, foreignKey, localKey string) *HasMany
func (m *BaseModel) BelongsTo(model interface{}, foreignKey, localKey string) *BelongsTo
func (m *BaseModel) BelongsToMany(model interface{}, pivotTable, foreignKey, localKey string) *BelongsToMany

// 预加载
func (m *BaseModel) With(relations ...string) *ModelQueryBuilder
func (m *BaseModel) Load(relations ...string) error

// 序列化
func (m *BaseModel) ToMap() map[string]interface{}
func (m *BaseModel) ToJSON() (string, error)

// 迁移
func (m *BaseModel) AutoMigrate(models ...interface{}) error
```

## 🔄 事务接口

### TransactionInterface

```go
type TransactionInterface interface {
    Query(query string, args ...interface{}) (*sql.Rows, error)
    QueryRow(query string, args ...interface{}) *sql.Row
    Exec(query string, args ...interface{}) (sql.Result, error)
    Commit() error
    Rollback() error
}
```

### 事务使用

```go
// 闭包事务（推荐）
err := torm.Transaction(func(tx torm.TransactionInterface) error {
    _, err := tx.Exec("INSERT INTO users (name) VALUES (?)", "张三")
    if err != nil {
        return err // 自动回滚
    }
    return nil // 自动提交
})

// 手动事务
tx, err := torm.BeginTransaction()
if err != nil {
    return err
}

_, err = tx.Exec("INSERT INTO users (name) VALUES (?)", "张三")
if err != nil {
    tx.Rollback()
    return err
}

err = tx.Commit()
```

## 💾 缓存接口

### CacheInterface

```go
type CacheInterface interface {
    Get(key string) (interface{}, error)
    Set(key string, value interface{}, duration time.Duration) error
    Delete(key string) error
    Has(key string) (bool, error)
    Clear() error
}

type FullCacheInterface interface {
    CacheInterface
    GetMulti(keys []string) (map[string]interface{}, error)
    SetMulti(items map[string]interface{}, duration time.Duration) error
    SetWithTags(key string, value interface{}, duration time.Duration, tags []string) error
    DeleteByTags(tags []string) error
    Increment(key string, delta int64) (int64, error)
    Decrement(key string, delta int64) (int64, error)
    GetOrSet(key string, fn func() (interface{}, error), duration time.Duration) (interface{}, error)
}
```

## 📜 迁移接口

### Migrator

```go
// 创建迁移器
func NewMigrator(conn Connection, logger LoggerInterface) *Migrator

// 注册迁移
func (m *Migrator) Register(migration MigrationInterface) *Migrator
func (m *Migrator) RegisterFunc(version, description string, upFunc, downFunc func(conn Connection) error) *Migrator

// 执行迁移
func (m *Migrator) Up() error
func (m *Migrator) Down(steps int) error
func (m *Migrator) Reset() error

// 状态
func (m *Migrator) Status() ([]*MigrationStatus, error)
```

### AutoMigrate

```go
// 自动迁移
err := model.AutoMigrate(&User{}, &Post{}, &Comment{})
```

## 🔗 相关文档

- [查询构建器](Query-Builder) - 查询构建器详细用法
- [模型系统](Model-System) - 模型系统详细说明
- [事务处理](Transactions) - 事务使用指南
- [缓存系统](Caching) - 缓存系统用法
- [数据迁移](Migrations) - 迁移系统用法
