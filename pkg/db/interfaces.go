package db

import (
	"context"
	"database/sql"
	"time"
)

// ConnectionInterface 数据库连接接口
type ConnectionInterface interface {
	// 连接管理
	Connect() error
	Close() error
	Ping() error
	IsConnected() bool

	// 查询操作
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)

	// 事务操作
	Begin() (TransactionInterface, error)
	BeginTx(opts *sql.TxOptions) (TransactionInterface, error)

	// 连接信息
	GetConfig() *Config
	GetDriver() string
	GetStats() sql.DBStats
}

// TransactionInterface 事务接口
type TransactionInterface interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
	Commit() error
	Rollback() error
}

// QueryInterface 查询构造器接口
type QueryInterface interface {
	// 表设置
	From(table string) QueryInterface

	// 查询条件
	Where(field string, operator string, value interface{}) QueryInterface
	WhereIn(field string, values []interface{}) QueryInterface
	WhereNotIn(field string, values []interface{}) QueryInterface
	WhereBetween(field string, start, end interface{}) QueryInterface
	WhereNull(field string) QueryInterface
	WhereNotNull(field string) QueryInterface
	WhereRaw(raw string, bindings ...interface{}) QueryInterface
	OrWhere(field string, operator string, value interface{}) QueryInterface

	// 连接查询
	Join(table string, first string, operator string, second string) QueryInterface
	LeftJoin(table string, first string, operator string, second string) QueryInterface
	RightJoin(table string, first string, operator string, second string) QueryInterface
	InnerJoin(table string, first string, operator string, second string) QueryInterface

	// 字段选择
	Select(fields ...string) QueryInterface
	SelectRaw(raw string, bindings ...interface{}) QueryInterface
	Distinct() QueryInterface

	// 分组和排序
	GroupBy(fields ...string) QueryInterface
	Having(field string, operator string, value interface{}) QueryInterface
	OrderBy(field string, direction string) QueryInterface
	OrderByRaw(raw string, bindings ...interface{}) QueryInterface

	// 限制和偏移
	Limit(limit int) QueryInterface
	Offset(offset int) QueryInterface
	Page(page, pageSize int) QueryInterface

	// 上下文控制
	WithContext(ctx context.Context) QueryInterface
	WithTimeout(timeout time.Duration) QueryInterface

	// 执行查询
	Get() ([]map[string]interface{}, error)
	First() (map[string]interface{}, error)
	Find(id interface{}) (map[string]interface{}, error)
	Count() (int64, error)
	Exists() (bool, error)

	// 分页查询
	Paginate(page, perPage int) (interface{}, error)

	// 数据操作
	Insert(data map[string]interface{}) (int64, error)
	InsertBatch(data []map[string]interface{}) (int64, error)
	Update(data map[string]interface{}) (int64, error)
	Delete() (int64, error)

	// SQL构建
	ToSQL() (string, []interface{}, error)
	Clone() QueryInterface
}

// BuilderInterface SQL构建器接口
type BuilderInterface interface {
	BuildSelect(query QueryInterface) (string, []interface{}, error)
	BuildInsert(table string, data map[string]interface{}) (string, []interface{}, error)
	BuildInsertBatch(table string, data []map[string]interface{}) (string, []interface{}, error)
	BuildUpdate(query QueryInterface, data map[string]interface{}) (string, []interface{}, error)
	BuildDelete(query QueryInterface) (string, []interface{}, error)
	BuildCount(query QueryInterface) (string, []interface{}, error)
	QuoteIdentifier(identifier string) string
	QuoteValue(value interface{}) string
}

// ModelInterface 模型接口
type ModelInterface interface {
	// 表信息
	TableName() string
	PrimaryKey() string
	GetConnection() string

	// 数据操作
	Save() error
	Delete() error
	Reload() error

	// 属性操作
	GetAttribute(key string) interface{}
	SetAttribute(key string, value interface{})
	GetAttributes() map[string]interface{}
	SetAttributes(attributes map[string]interface{})

	// 状态检查
	IsNew() bool
	IsDirty() bool
	GetDirty() map[string]interface{}

	// 事件钩子
	BeforeSave() error
	AfterSave() error
	BeforeCreate() error
	AfterCreate() error
	BeforeUpdate() error
	AfterUpdate() error
	BeforeDelete() error
	AfterDelete() error
}

// CacheInterface 缓存接口
type CacheInterface interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}, ttl time.Duration) error
	Delete(key string) error
	Clear() error
	Has(key string) (bool, error)
}

// LoggerInterface 日志接口
type LoggerInterface interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
}
