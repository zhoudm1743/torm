# API 参考

TORM API 完整参考文档，包含所有接口、方法和参数说明。

## 📋 目录

- [数据库连接](#数据库连接)
- [查询构建器](#查询构建器)
- [模型接口](#模型接口)
- [事务接口](#事务接口)
- [缓存接口](#缓存接口)
- [日志接口](#日志接口)

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
}
```

### 连接管理

```go
// 添加数据库连接
func AddConnection(name string, config *Config) error

// 获取数据库连接
func DB(name ...string) (DatabaseInterface, error)

// 获取默认连接
func GetDefaultConnection() DatabaseInterface

// 设置默认连接
func SetDefaultConnection(name string) error

// 关闭连接
func CloseConnection(name string) error

// 关闭所有连接
func CloseAllConnections() error
```

## 🔍 查询构建器

### QueryInterface

查询构建器核心接口：

```go
type QueryInterface interface {
    // 表操作
    Table(table string) QueryInterface
    From(table string) QueryInterface
    
    // 字段选择
    Select(columns ...string) QueryInterface
    SelectRaw(expression string, bindings ...interface{}) QueryInterface
    Distinct() QueryInterface
    
    // 条件查询
    Where(args ...interface{}) QueryInterface                    // 支持多种调用方式
    WhereRaw(expression string, bindings ...interface{}) QueryInterface
    WhereIn(column string, values []interface{}) QueryInterface
    WhereNotIn(column string, values []interface{}) QueryInterface
    WhereBetween(column string, values []interface{}) QueryInterface  // v1.3.0 更新
    WhereNotBetween(column string, values []interface{}) QueryInterface  // v1.3.0 新增
    WhereNull(column string) QueryInterface
    WhereNotNull(column string) QueryInterface
    WhereExists(subQuery interface{}) QueryInterface              // v1.3.0 增强
    WhereNotExists(subQuery interface{}) QueryInterface          // v1.3.0 增强
    
    // OR 条件
    OrWhere(column string, operator string, value interface{}) QueryInterface
    OrWhereRaw(expression string, bindings ...interface{}) QueryInterface
    OrWhereIn(column string, values []interface{}) QueryInterface
    OrWhereNotIn(column string, values []interface{}) QueryInterface
    OrWhereBetween(column string, min, max interface{}) QueryInterface
    OrWhereNotBetween(column string, min, max interface{}) QueryInterface
    OrWhereNull(column string) QueryInterface
    OrWhereNotNull(column string) QueryInterface
    
    // 连接查询
    Join(table string, first string, operator string, second string) QueryInterface
    LeftJoin(table string, first string, operator string, second string) QueryInterface
    RightJoin(table string, first string, operator string, second string) QueryInterface
    CrossJoin(table string) QueryInterface
    
    // 排序
    OrderBy(column string, direction string) QueryInterface
    OrderByRaw(expression string, bindings ...interface{}) QueryInterface
    OrderRand() QueryInterface                               // v1.3.0 新增：随机排序
    OrderField(field string, values []interface{}, direction string) QueryInterface  // v1.3.0 新增：按值排序
    
    // 字段表达式
    FieldRaw(expression string, bindings ...interface{}) QueryInterface  // v1.3.0 新增：原生字段
    
    // 分组
    GroupBy(columns ...string) QueryInterface
    Having(column string, operator string, value interface{}) QueryInterface
    HavingRaw(expression string, bindings ...interface{}) QueryInterface
    
    // 限制
    Limit(limit int) QueryInterface
    Offset(offset int) QueryInterface
    Take(limit int) QueryInterface
    Skip(offset int) QueryInterface
    
    // 执行查询
    Model(model interface{}) QueryInterface                     // 设置模型实例，启用访问器
    Get() ([]map[string]interface{}, error)                     // 返回数据（可应用访问器）
    First(dest ...interface{}) (map[string]interface{}, error)  // 返回第一条记录
    Find(id interface{}, dest ...interface{}) (map[string]interface{}, error)  // 根据ID查找
    
    // 原始数据查询 (向下兼容)
    GetRaw() ([]map[string]interface{}, error)                  // 返回原始map数据
    FirstRaw() (map[string]interface{}, error)                  // 返回原始map数据
    Pluck(column string) ([]interface{}, error)
    Value(column string) (interface{}, error)
    Exists() (bool, error)
    
    // 聚合查询
    Count() (int64, error)
    Sum(column string) (interface{}, error)
    Avg(column string) (interface{}, error)
    Max(column string) (interface{}, error)
    Min(column string) (interface{}, error)
    
    // 插入
    Insert(data map[string]interface{}) (int64, error)
    InsertBatch(data []map[string]interface{}) (int64, error)
    InsertGetId(data map[string]interface{}) (int64, error)
    InsertIgnore(data map[string]interface{}) (int64, error)
    
    // 更新
    Update(data map[string]interface{}) (int64, error)
    UpdateOrInsert(attributes map[string]interface{}, values map[string]interface{}) error
    Increment(column string, amount interface{}) (int64, error)
    Decrement(column string, amount interface{}) (int64, error)
    
    // 删除
    Delete() (int64, error)
    Truncate() error
    
    // 分页
    Paginate(page int, perPage int) (*SimplePagination, error)
    
    // 分块处理
    Chunk(size int, callback func([]map[string]interface{}) bool) error
    
    // 缓存
    Cache(duration time.Duration) QueryInterface
    CacheWithTags(duration time.Duration, tags ...string) QueryInterface
    
    // 原生SQL
    Raw(sql string, bindings ...interface{}) QueryInterface
    
    // 调试
    ToSQL() (string, []interface{})
    Explain() ([]map[string]interface{}, error)
    
    // 克隆
    Clone() QueryInterface
}
```

### 分页结果

```go
type SimplePagination struct {
    Data        []map[string]interface{} `json:"data"`     // 数据（可包含访问器处理）
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

基础模型结构：

```go
type BaseModel struct {
    // 内部字段（不直接访问）
    attributes   map[string]interface{}
    original     map[string]interface{}
    relations    map[string]interface{}
    exists       bool
    wasRecentlyCreated bool
    
    // 配置字段
    tableName    string
    primaryKeys  []string
    connection   string
    queryBuilder QueryInterface
}
```

### 模型方法

```go
// 表操作
func (m *BaseModel) SetTable(table string) *BaseModel
func (m *BaseModel) TableName() string
func (m *BaseModel) GetTable() string

// 连接操作
func (m *BaseModel) SetConnection(connection string) *BaseModel
func (m *BaseModel) GetConnection() string

// 主键操作
func (m *BaseModel) PrimaryKey() string
func (m *BaseModel) PrimaryKeys() []string
func (m *BaseModel) SetPrimaryKey(key string) *BaseModel
func (m *BaseModel) SetPrimaryKeys(keys []string) *BaseModel
func (m *BaseModel) HasCompositePrimaryKey() bool
func (m *BaseModel) GetKey() interface{}
func (m *BaseModel) SetKey(key interface{}) *BaseModel
func (m *BaseModel) DetectPrimaryKeysFromStruct(structValue interface{}) *BaseModel

// 属性操作
func (m *BaseModel) GetAttribute(key string) interface{}
func (m *BaseModel) SetAttribute(key string, value interface{}) *BaseModel
func (m *BaseModel) GetAttributes() map[string]interface{}
func (m *BaseModel) SetAttributes(attributes map[string]interface{}) *BaseModel
func (m *BaseModel) Fill(attributes map[string]interface{}) *BaseModel

// 脏数据检测
func (m *BaseModel) IsDirty(keys ...string) bool
func (m *BaseModel) GetDirty() map[string]interface{}
func (m *BaseModel) GetOriginal(key ...string) interface{}
func (m *BaseModel) SyncOriginal() *BaseModel

// 状态检查
func (m *BaseModel) Exists() bool
func (m *BaseModel) WasRecentlyCreated() bool
func (m *BaseModel) IsNew() bool

// 查询方法
func (m *BaseModel) All() ([]map[string]interface{}, error)
func (m *BaseModel) Find(id interface{}, dest ...interface{}) (map[string]interface{}, error)
func (m *BaseModel) FindOrFail(id interface{}) error
func (m *BaseModel) First(dest ...interface{}) (map[string]interface{}, error)
func (m *BaseModel) FirstOrFail() error
func (m *BaseModel) FirstOrCreate(attributes map[string]interface{}) error
func (m *BaseModel) FirstOrNew(attributes map[string]interface{}) *BaseModel
func (m *BaseModel) UpdateOrCreate(attributes map[string]interface{}, values map[string]interface{}) error

// CRUD操作
func (m *BaseModel) Save() error
func (m *BaseModel) Create(attributes map[string]interface{}) error
func (m *BaseModel) Update(attributes map[string]interface{}) error
func (m *BaseModel) Delete() error
func (m *BaseModel) ForceDelete() error

// 查询构建器方法（所有QueryInterface方法都可用）
func (m *BaseModel) Where(column string, operator string, value interface{}) *BaseModel
func (m *BaseModel) WhereIn(column string, values []interface{}) *BaseModel
func (m *BaseModel) OrderBy(column string, direction string) *BaseModel
func (m *BaseModel) Limit(limit int) *BaseModel
// ... 其他所有查询方法

// 关联关系
func (m *BaseModel) HasOne(related interface{}, foreignKey string, localKey string) *HasOne
func (m *BaseModel) HasMany(related interface{}, foreignKey string, localKey string) *HasMany
func (m *BaseModel) BelongsTo(related interface{}, foreignKey string, ownerKey string) *BelongsTo
func (m *BaseModel) ManyToMany(related interface{}, table string, foreignPivotKey string, relatedPivotKey string) *ManyToMany

// 预加载
func (m *BaseModel) With(relations ...string) *BaseModel
func (m *BaseModel) WithCount(relations ...string) *BaseModel
func (m *BaseModel) Load(relations ...string) error
func (m *BaseModel) LoadCount(relations ...string) error

// 关联数据访问
func (m *BaseModel) GetRelation(key string) interface{}
func (m *BaseModel) SetRelation(key string, value interface{}) *BaseModel
func (m *BaseModel) HasRelation(key string) bool

// 序列化
func (m *BaseModel) ToMap() map[string]interface{}
func (m *BaseModel) ToJSON() ([]byte, error)
```

## 🔄 事务接口

### TransactionInterface

```go
type TransactionInterface interface {
    QueryInterface  // 继承所有查询方法
    
    // 事务操作
    Commit() error
    Rollback() error
    
    // 保存点
    Savepoint(name string) error
    RollbackToSavepoint(name string) error
    ReleaseSavepoint(name string) error
    
    // 嵌套事务
    Transaction(callback func(TransactionInterface) error) error
}
```

### 事务管理

```go
// 开始事务
func Begin() (TransactionInterface, error)

// 自动事务
func Transaction(callback func(TransactionInterface) error) error

// 带上下文的事务
func TransactionWithContext(ctx context.Context, callback func(TransactionInterface) error) error
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
    Pull(key string) (interface{}, error)
    
    // 批量操作
    GetMultiple(keys []string) (map[string]interface{}, error)
    SetMultiple(items map[string]interface{}, duration time.Duration) error
    DeleteMultiple(keys []string) error
    
    // 标签操作
    SetWithTags(key string, value interface{}, duration time.Duration, tags ...string) error
    FlushByTags(tags ...string) error
    
    // 管理操作
    Flush() error
    Forever(key string, value interface{}) error
    Forget(key string) error
    
    // 统计信息
    Stats() *CacheStats
}

type CacheStats struct {
    Hits         int64
    Misses       int64
    Size         int64
    Evictions    int64
    LastAccess   time.Time
}
```

## 📝 日志接口

### LoggerInterface

```go
type LoggerInterface interface {
    // 基础日志方法
    Debug(message string, fields ...interface{})
    Info(message string, fields ...interface{})
    Warn(message string, fields ...interface{})
    Error(message string, fields ...interface{})
    Fatal(message string, fields ...interface{})
    
    // 格式化日志方法
    Debugf(format string, args ...interface{})
    Infof(format string, args ...interface{})
    Warnf(format string, args ...interface{})
    Errorf(format string, args ...interface{})
    Fatalf(format string, args ...interface{})
    
    // 结构化日志方法
    DebugWithFields(message string, fields map[string]interface{})
    InfoWithFields(message string, fields map[string]interface{})
    WarnWithFields(message string, fields map[string]interface{})
    ErrorWithFields(message string, fields map[string]interface{})
    FatalWithFields(message string, fields map[string]interface{})
    
    // 配置方法
    SetLevel(level LogLevel)
    SetOutput(output io.Writer)
    SetFormatter(formatter FormatterInterface)
    AddHook(hook HookInterface)
}

// 日志级别
type LogLevel int

const (
    DEBUG LogLevel = iota
    INFO
    WARN
    ERROR
    FATAL
)
```

### SQL日志接口

```go
type SQLLoggerInterface interface {
    LoggerInterface
    
    // SQL专用方法
    LogQuery(sql string, bindings []interface{}, duration time.Duration)
    LogSlowQuery(sql string, bindings []interface{}, duration time.Duration, threshold time.Duration)
    
    // 配置方法
    SetSlowQueryThreshold(threshold time.Duration)
    EnableQueryLog(enabled bool)
    GetQueryLog() []*QueryLog
    ClearQueryLog()
}

type QueryLog struct {
    SQL       string
    Bindings  []interface{}
    Duration  time.Duration
    Timestamp time.Time
}
```

## 🔗 相关文档

- [查询构建器](Query-Builder) - 查询构建器详细用法
- [模型系统](Model-System) - 模型系统详细说明
- [事务处理](Transactions) - 事务使用指南
- [缓存系统](Caching) - 缓存系统用法
- [日志系统](Logging) - 日志系统配置 