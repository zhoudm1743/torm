# API 参考

TORM API 完整参考文档，包含所有核心接口、方法和参数说明。

## 📋 目录

- [数据库连接](#数据库连接)
- [查询构建器](#查询构建器)
- [模型接口](#模型接口)
- [事务接口](#事务接口)
- [缓存接口](#缓存接口)
- [迁移接口](#迁移接口)

## 🔌 数据库连接

### db.Config

数据库配置结构体：

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
    
    // SSL配置
    SSLMode         string        // SSL模式: "disable", "require", "verify-ca"
    SSLCert         string        // SSL证书路径
    SSLKey          string        // SSL密钥路径
    SSLRootCert     string        // SSL根证书路径
    
    // 日志配置
    LogQueries      bool          // 是否记录查询日志
}
```

### 连接管理函数

```go
// 添加数据库连接
func AddConnection(name string, config *Config) error

// 获取数据库管理器
func DefaultManager() *DbManager

// 获取数据库连接
func (m *DbManager) Connection(name string) (Connection, error)

// 关闭连接
func (m *DbManager) Close(name string) error

// 关闭所有连接
func (m *DbManager) CloseAll() error
```

### DSN 构建

```go
// MySQL DSN
func (c *Config) MySQLDSN() string

// PostgreSQL DSN
func (c *Config) PostgresDSN() string

// SQLite DSN
func (c *Config) SQLiteDSN() string
```

## 🔍 查询构建器

### QueryBuilder

查询构建器核心方法：

```go
// 创建查询构建器
func NewQueryBuilder(connectionName string) (*QueryBuilder, error)

// 表操作
func (q *QueryBuilder) Table(table string) *QueryBuilder
func (q *QueryBuilder) From(table string) *QueryBuilder

// 字段选择
func (q *QueryBuilder) Select(columns ...interface{}) *QueryBuilder
func (q *QueryBuilder) SelectRaw(expression string, bindings ...interface{}) *QueryBuilder
func (q *QueryBuilder) Distinct() *QueryBuilder

// 条件查询
func (q *QueryBuilder) Where(args ...interface{}) *QueryBuilder
func (q *QueryBuilder) OrWhere(args ...interface{}) *QueryBuilder
func (q *QueryBuilder) WhereRaw(expression string, bindings ...interface{}) *QueryBuilder
func (q *QueryBuilder) WhereIn(column string, values []interface{}) *QueryBuilder
func (q *QueryBuilder) WhereNotIn(column string, values []interface{}) *QueryBuilder
func (q *QueryBuilder) WhereBetween(column string, values []interface{}) *QueryBuilder
func (q *QueryBuilder) WhereNotBetween(column string, values []interface{}) *QueryBuilder
func (q *QueryBuilder) WhereNull(column string) *QueryBuilder
func (q *QueryBuilder) WhereNotNull(column string) *QueryBuilder
func (q *QueryBuilder) WhereLike(column string, pattern string) *QueryBuilder
func (q *QueryBuilder) WhereNotLike(column string, pattern string) *QueryBuilder

// 连接查询
func (q *QueryBuilder) Join(table, first, operator, second string) *QueryBuilder
func (q *QueryBuilder) LeftJoin(table, first, operator, second string) *QueryBuilder
func (q *QueryBuilder) RightJoin(table, first, operator, second string) *QueryBuilder
func (q *QueryBuilder) CrossJoin(table string) *QueryBuilder

// 排序
func (q *QueryBuilder) OrderBy(column, direction string) *QueryBuilder
func (q *QueryBuilder) OrderByRaw(expression string, bindings ...interface{}) *QueryBuilder
func (q *QueryBuilder) OrderRand() *QueryBuilder

// 分组
func (q *QueryBuilder) GroupBy(columns ...string) *QueryBuilder
func (q *QueryBuilder) Having(args ...interface{}) *QueryBuilder
func (q *QueryBuilder) HavingRaw(expression string, bindings ...interface{}) *QueryBuilder

// 限制
func (q *QueryBuilder) Limit(limit int) *QueryBuilder
func (q *QueryBuilder) Offset(offset int) *QueryBuilder
func (q *QueryBuilder) Page(page, size int) *QueryBuilder

// 执行查询
func (q *QueryBuilder) Get() ([]map[string]interface{}, error)
func (q *QueryBuilder) GetRaw() ([]map[string]interface{}, error)
func (q *QueryBuilder) First() (map[string]interface{}, error)
func (q *QueryBuilder) FirstRaw() (map[string]interface{}, error)
func (q *QueryBuilder) Pluck(column string) ([]interface{}, error)
func (q *QueryBuilder) Value(column string) (interface{}, error)
func (q *QueryBuilder) Exists() (bool, error)

// 聚合查询
func (q *QueryBuilder) Count() (int64, error)
func (q *QueryBuilder) Sum(column string) (interface{}, error)
func (q *QueryBuilder) Avg(column string) (interface{}, error)
func (q *QueryBuilder) Max(column string) (interface{}, error)
func (q *QueryBuilder) Min(column string) (interface{}, error)

// 插入
func (q *QueryBuilder) Insert(data map[string]interface{}) (int64, error)
func (q *QueryBuilder) InsertBatch(data []map[string]interface{}) (int64, error)
func (q *QueryBuilder) InsertGetId(data map[string]interface{}) (int64, error)

// 更新
func (q *QueryBuilder) Update(data map[string]interface{}) (int64, error)
func (q *QueryBuilder) Increment(column string, amount interface{}) (int64, error)
func (q *QueryBuilder) Decrement(column string, amount interface{}) (int64, error)

// 删除
func (q *QueryBuilder) Delete() (int64, error)
func (q *QueryBuilder) Truncate() error

// 分页
func (q *QueryBuilder) Paginate(page, perPage int) (*Pagination, error)

// 分块处理
func (q *QueryBuilder) Chunk(size int, callback func([]map[string]interface{}) bool) error

// 缓存
func (q *QueryBuilder) Cache(duration time.Duration) *QueryBuilder
func (q *QueryBuilder) CacheWithTags(duration time.Duration, tags ...string) *QueryBuilder
func (q *QueryBuilder) CacheKey(key string) *QueryBuilder

// 调试
func (q *QueryBuilder) ToSQL() (string, []interface{})

// 克隆
func (q *QueryBuilder) Clone() *QueryBuilder
```

### 分页结果

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

### BaseModel

基础模型方法：

```go
// 创建模型
func NewModel(args ...interface{}) *BaseModel

// 配置方法
func (m *BaseModel) SetTable(table string) *BaseModel
func (m *BaseModel) GetTableName() string
func (m *BaseModel) SetPrimaryKey(key string) *BaseModel
func (m *BaseModel) GetPrimaryKey() string
func (m *BaseModel) SetConnection(connection string) *BaseModel
func (m *BaseModel) GetConnection() string

// 属性操作
func (m *BaseModel) SetAttribute(key string, value interface{}) Model
func (m *BaseModel) GetAttribute(key string) interface{}
func (m *BaseModel) SetAttributes(attrs map[string]interface{}) Model
func (m *BaseModel) GetAttributes() map[string]interface{}
func (m *BaseModel) Fill(data map[string]interface{}) *BaseModel
func (m *BaseModel) ClearAttributes() *BaseModel

// 状态方法
func (m *BaseModel) IsNew() bool
func (m *BaseModel) IsExists() bool
func (m *BaseModel) IsDirty() bool
func (m *BaseModel) GetDirty() map[string]interface{}
func (m *BaseModel) MarkAsNew() *BaseModel
func (m *BaseModel) MarkAsExists() *BaseModel

// 主键操作
func (m *BaseModel) GetKey() interface{}
func (m *BaseModel) SetKey(key interface{}) *BaseModel

// 持久化操作
func (m *BaseModel) Save() error
func (m *BaseModel) Delete() error
func (m *BaseModel) Find(pk interface{}) error
func (m *BaseModel) FindByPK(pk interface{}) error
func (m *BaseModel) Reload() error

// 软删除
func (m *BaseModel) SoftDelete() error
func (m *BaseModel) Restore() error
func (m *BaseModel) ForceDelete() error
func (m *BaseModel) IsSoftDeleted() bool
func (m *BaseModel) EnableSoftDeletes() *BaseModel
func (m *BaseModel) DisableSoftDeletes() *BaseModel

// 时间戳
func (m *BaseModel) EnableTimestamps() Model
func (m *BaseModel) DisableTimestamps() Model
func (m *BaseModel) SetCreatedAtField(field string) *BaseModel
func (m *BaseModel) SetUpdatedAtField(field string) *BaseModel

// 查询方法
func (m *BaseModel) Query() (*QueryBuilder, error)
func (m *BaseModel) RawQuery() (*QueryBuilder, error)
func (m *BaseModel) Where(args ...interface{}) (*QueryBuilder, error)
func (m *BaseModel) Get() ([]map[string]interface{}, error)
func (m *BaseModel) First() (map[string]interface{}, error)
func (m *BaseModel) Count() (int64, error)

// 关联方法
func (m *BaseModel) HasOne(model interface{}, foreignKey, localKey string) *HasOne
func (m *BaseModel) HasMany(model interface{}, foreignKey, localKey string) *HasMany
func (m *BaseModel) BelongsTo(model interface{}, foreignKey, localKey string) *BelongsTo
func (m *BaseModel) BelongsToMany(model interface{}, pivotTable, foreignKey, localKey string) *BelongsToMany

// 预加载
func (m *BaseModel) With(relations ...string) *ModelQueryBuilder
func (m *BaseModel) WithConstraint(relation string, constraint func(*QueryBuilder) *QueryBuilder) *ModelQueryBuilder
func (m *BaseModel) Load(relations ...string) error
func (m *BaseModel) GetRelation(name string) interface{}
func (m *BaseModel) SetRelation(name string, data interface{}) *BaseModel

// 序列化
func (m *BaseModel) ToMap() map[string]interface{}
func (m *BaseModel) ToJSON() (string, error)
func (m *BaseModel) FromJSON(jsonStr string) error

// 迁移
func (m *BaseModel) AutoMigrate(models ...interface{}) error
```

### 事件钩子

```go
// 保存钩子
func (m *BaseModel) BeforeSave() error
func (m *BaseModel) AfterSave() error

// 创建钩子
func (m *BaseModel) BeforeCreate() error
func (m *BaseModel) AfterCreate() error

// 更新钩子
func (m *BaseModel) BeforeUpdate() error
func (m *BaseModel) AfterUpdate() error

// 删除钩子
func (m *BaseModel) BeforeDelete() error
func (m *BaseModel) AfterDelete() error
```

## 🔄 事务接口

### TransactionInterface

```go
type TransactionInterface interface {
    // SQL执行
    Query(query string, args ...interface{}) (*sql.Rows, error)
    QueryRow(query string, args ...interface{}) *sql.Row
    Exec(query string, args ...interface{}) (sql.Result, error)
    
    // 事务控制
    Commit() error
    Rollback() error
}
```

### 事务函数

```go
// 闭包事务
func Transaction(fn func(tx TransactionInterface) error, connectionName ...string) error

// 开始事务
func BeginTransaction(connectionName ...string) (*DBTransaction, error)

// 运行事务
func RunInTransaction(fn func(tx TxInterface) error, connectionName ...string) error
```

## 💾 缓存接口

### CacheInterface

```go
type CacheInterface interface {
    // 基础操作
    Get(key string) (interface{}, error)
    Set(key string, value interface{}, duration time.Duration) error
    Delete(key string) error
    Has(key string) (bool, error)
    Clear() error
    Size() int
    Close() error
}

type FullCacheInterface interface {
    CacheInterface
    
    // 批量操作
    GetMulti(keys []string) (map[string]interface{}, error)
    SetMulti(items map[string]interface{}, duration time.Duration) error
    
    // 标签操作
    SetWithTags(key string, value interface{}, duration time.Duration, tags []string) error
    DeleteByTags(tags []string) error
    
    // 数值操作
    Increment(key string, delta int64) (int64, error)
    Decrement(key string, delta int64) (int64, error)
    
    // 获取或设置
    GetOrSet(key string, fn func() (interface{}, error), duration time.Duration) (interface{}, error)
    
    // 统计
    Stats() map[string]interface{}
    ResetStats()
}
```

### 缓存管理

```go
// 获取默认缓存管理器
func GetDefaultCacheManager() *CacheManager

// 添加缓存
func AddCache(name string, cacheType CacheType, config interface{}) error

// 获取缓存
func GetCache(name string) (FullCacheInterface, error)

// 列出缓存
func (m *CacheManager) ListCaches() []string

// 移除缓存
func (m *CacheManager) RemoveCache(name string)

// 关闭所有缓存
func (m *CacheManager) CloseAll()
```

## 📜 迁移接口

### MigrationInterface

```go
type MigrationInterface interface {
    Up(conn Connection) error
    Down(conn Connection) error
    Version() string
    Description() string
}
```

### Migrator

```go
// 创建迁移器
func NewMigrator(conn Connection, logger LoggerInterface) *Migrator

// 配置
func (m *Migrator) SetTableName(tableName string) *Migrator
func (m *Migrator) SetAutoCreate(autoCreate bool) *Migrator

// 注册迁移
func (m *Migrator) Register(migration MigrationInterface) *Migrator
func (m *Migrator) RegisterFunc(version, description string, upFunc, downFunc func(conn Connection) error) *Migrator

// 执行迁移
func (m *Migrator) Up() error
func (m *Migrator) Down(steps int) error
func (m *Migrator) Reset() error
func (m *Migrator) Fresh() error

// 状态
func (m *Migrator) Status() ([]*MigrationStatus, error)
func (m *Migrator) PrintStatus() error
```

### MigrationStatus

```go
type MigrationStatus struct {
    Version     string     `json:"version"`
    Description string     `json:"description"`
    Applied     bool       `json:"applied"`
    AppliedAt   *time.Time `json:"applied_at,omitempty"`
    Batch       int        `json:"batch,omitempty"`
}
```

## 🔗 相关文档

- [查询构建器](Query-Builder) - 查询构建器详细用法
- [模型系统](Model-System) - 模型系统详细说明
- [事务处理](Transactions) - 事务使用指南
- [缓存系统](Caching) - 缓存系统用法
- [数据迁移](Migrations) - 迁移系统用法
