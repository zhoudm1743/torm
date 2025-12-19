package db

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// Connection 数据库连接接口
// 定义了所有数据库连接必须实现的方法
type Connection interface {
	// 连接管理
	Connect() error
	Close() error
	Ping() error
	IsConnected() bool

	// 查询执行
	Query(sql string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, sql string, args ...interface{}) (*sql.Rows, error)
	QueryRow(sql string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, sql string, args ...interface{}) *sql.Row
	Exec(sql string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, sql string, args ...interface{}) (sql.Result, error)

	// 事务
	Begin() (TxInterface, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (TxInterface, error)
	RunTransaction(fn func(tx TxInterface) error) error

	// 元数据
	GetConfig() *Config
	GetDriver() string
	GetDB() *sql.DB
	GetBuilder() Builder

	// 表信息
	GetTableFields(table string) ([]ConnFieldInfo, error)
	GetTables() ([]string, error)

	// 连接统计
	GetStats() sql.DBStats

	// 日志
	SetLogger(logger LoggerInterface)
	GetLogger() LoggerInterface
}

// TxInterface 事务接口（扩展版）
// 定义了事务操作必须实现的方法，包含嵌套事务支持
type TxInterface interface {
	// 查询执行
	Query(sql string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, sql string, args ...interface{}) (*sql.Rows, error)
	QueryRow(sql string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, sql string, args ...interface{}) *sql.Row
	Exec(sql string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, sql string, args ...interface{}) (sql.Result, error)

	// 事务控制
	Commit() error
	Rollback() error

	// 嵌套事务（保存点）
	Savepoint(name string) error
	RollbackTo(name string) error
	ReleaseSavepoint(name string) error

	// 获取底层事务
	GetTx() *sql.Tx
}

// ConnFieldInfo 字段信息（用于连接层）
type ConnFieldInfo struct {
	Name       string      // 字段名
	Type       string      // 字段类型
	Nullable   bool        // 是否可为空
	Default    interface{} // 默认值
	PrimaryKey bool        // 是否为主键
	AutoInc    bool        // 是否自增
	Comment    string      // 字段注释
	Length     int         // 字段长度
	Precision  int         // 精度（数值类型）
	Scale      int         // 小数位数（数值类型）
}

// Builder SQL构建器接口（前向声明，完整定义在builder.go中）
type Builder interface {
	// SELECT
	BuildSelect(q interface{}) (string, []interface{})

	// INSERT
	BuildInsert(table string, data map[string]interface{}) (string, []interface{})
	BuildInsertBatch(table string, data []map[string]interface{}) (string, []interface{})

	// UPDATE
	BuildUpdate(q interface{}, data map[string]interface{}) (string, []interface{})

	// DELETE
	BuildDelete(q interface{}) (string, []interface{})

	// 聚合
	BuildCount(q interface{}) (string, []interface{})
	BuildAggregate(q interface{}, fn string, field string) (string, []interface{})

	// 辅助
	QuoteIdentifier(name string) string
	QuotePlaceholder(index int) string
	GetPlaceholderStyle() PlaceholderStyle
}

// PlaceholderStyle 占位符风格
type PlaceholderStyle int

const (
	PlaceholderQuestion PlaceholderStyle = iota // ? (MySQL, SQLite)
	PlaceholderDollar                           // $1, $2 (PostgreSQL)
)

// BaseConnection 基础连接实现
// 提供所有数据库连接的通用功能
type BaseConnection struct {
	config    *Config
	db        *sql.DB
	builder   Builder
	logger    LoggerInterface
	connected bool
	mu        sync.RWMutex
}

// NewBaseConnection 创建基础连接
func NewBaseConnection(config *Config) *BaseConnection {
	return &BaseConnection{
		config:    config,
		connected: false,
	}
}

// SetDB 设置数据库连接
func (c *BaseConnection) SetDB(db *sql.DB) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.db = db
	c.connected = db != nil
}

// SetBuilder 设置SQL构建器
func (c *BaseConnection) SetBuilder(builder Builder) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.builder = builder
}

// SetLogger 设置日志记录器
func (c *BaseConnection) SetLogger(logger LoggerInterface) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger = logger
}

// GetLogger 获取日志记录器
func (c *BaseConnection) GetLogger() LoggerInterface {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.logger
}

// GetConfig 获取配置
func (c *BaseConnection) GetConfig() *Config {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config
}

// GetDB 获取底层数据库连接
func (c *BaseConnection) GetDB() *sql.DB {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.db
}

// GetBuilder 获取SQL构建器
func (c *BaseConnection) GetBuilder() Builder {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.builder
}

// IsConnected 检查是否已连接
func (c *BaseConnection) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected && c.db != nil
}

// Ping 测试连接
func (c *BaseConnection) Ping() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.db == nil {
		return NewError(ErrCodeConnectionClosed, "connection not established")
	}

	return c.db.Ping()
}

// Close 关闭连接
func (c *BaseConnection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.db == nil {
		return nil
	}

	err := c.db.Close()
	c.db = nil
	c.connected = false

	return err
}

// GetStats 获取连接统计
func (c *BaseConnection) GetStats() sql.DBStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.db == nil {
		return sql.DBStats{}
	}

	return c.db.Stats()
}

// ConfigurePool 配置连接池
func (c *BaseConnection) ConfigurePool() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.db == nil || c.config == nil {
		return
	}

	if c.config.MaxOpenConns > 0 {
		c.db.SetMaxOpenConns(c.config.MaxOpenConns)
	}
	if c.config.MaxIdleConns > 0 {
		c.db.SetMaxIdleConns(c.config.MaxIdleConns)
	}
	if c.config.ConnMaxLifetime > 0 {
		c.db.SetConnMaxLifetime(c.config.ConnMaxLifetime)
	}
	if c.config.ConnMaxIdleTime > 0 {
		c.db.SetConnMaxIdleTime(c.config.ConnMaxIdleTime)
	}
}

// LogQuery 记录查询日志
func (c *BaseConnection) LogQuery(sql string, args []interface{}, duration time.Duration, err error) {
	if c.logger == nil {
		return
	}

	if err != nil {
		c.logger.Error("SQL执行失败",
			"sql", sql,
			"args", args,
			"duration", duration,
			"error", err)
	} else if c.config != nil && c.config.LogQueries {
		c.logger.Debug("SQL执行",
			"sql", sql,
			"args", args,
			"duration", duration)
	}
}

// Query 执行查询
func (c *BaseConnection) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return c.QueryContext(context.Background(), query, args...)
}

// QueryContext 执行带上下文的查询
func (c *BaseConnection) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	c.mu.RLock()
	db := c.db
	c.mu.RUnlock()

	if db == nil {
		return nil, NewError(ErrCodeConnectionClosed, "connection not established")
	}

	start := time.Now()
	rows, err := db.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	c.LogQuery(query, args, duration, err)

	if err != nil {
		return nil, NewQueryError(err, query, args)
	}

	return rows, nil
}

// QueryRow 执行单行查询
func (c *BaseConnection) QueryRow(query string, args ...interface{}) *sql.Row {
	return c.QueryRowContext(context.Background(), query, args...)
}

// QueryRowContext 执行带上下文的单行查询
func (c *BaseConnection) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	c.mu.RLock()
	db := c.db
	c.mu.RUnlock()

	if db == nil {
		return nil
	}

	start := time.Now()
	row := db.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	c.LogQuery(query, args, duration, nil)

	return row
}

// Exec 执行SQL语句
func (c *BaseConnection) Exec(query string, args ...interface{}) (sql.Result, error) {
	return c.ExecContext(context.Background(), query, args...)
}

// ExecContext 执行带上下文的SQL语句
func (c *BaseConnection) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	c.mu.RLock()
	db := c.db
	c.mu.RUnlock()

	if db == nil {
		return nil, NewError(ErrCodeConnectionClosed, "connection not established")
	}

	start := time.Now()
	result, err := db.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	c.LogQuery(query, args, duration, err)

	if err != nil {
		return nil, NewQueryError(err, query, args)
	}

	return result, nil
}

// RunTransaction 执行事务（需要子类实现Begin方法后调用）
func (c *BaseConnection) RunTransaction(fn func(tx TxInterface) error, beginFn func() (TxInterface, error)) error {
	tx, err := beginFn()
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return WrapErrorf(err, ErrCodeTransactionRollbackFailed, "rollback failed after error: %v", rbErr)
		}
		return err
	}

	return tx.Commit()
}

// BaseTransaction 基础事务实现
type BaseTransaction struct {
	tx             *sql.Tx
	logger         LoggerInterface
	config         *Config
	savepointCount int
	mu             sync.Mutex
}

// NewBaseTransaction 创建基础事务
func NewBaseTransaction(tx *sql.Tx, logger LoggerInterface, config *Config) *BaseTransaction {
	return &BaseTransaction{
		tx:             tx,
		logger:         logger,
		config:         config,
		savepointCount: 0,
	}
}

// GetTx 获取底层事务
func (t *BaseTransaction) GetTx() *sql.Tx {
	return t.tx
}

// Query 在事务中执行查询
func (t *BaseTransaction) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return t.QueryContext(context.Background(), query, args...)
}

// QueryContext 在事务中执行带上下文的查询
func (t *BaseTransaction) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if t.tx == nil {
		return nil, NewError(ErrCodeTransactionNotStarted, "transaction not active")
	}

	start := time.Now()
	rows, err := t.tx.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	t.logQuery(query, args, duration, err)

	if err != nil {
		return nil, NewQueryError(err, query, args)
	}

	return rows, nil
}

// QueryRow 在事务中执行单行查询
func (t *BaseTransaction) QueryRow(query string, args ...interface{}) *sql.Row {
	return t.QueryRowContext(context.Background(), query, args...)
}

// QueryRowContext 在事务中执行带上下文的单行查询
func (t *BaseTransaction) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if t.tx == nil {
		return nil
	}

	start := time.Now()
	row := t.tx.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	t.logQuery(query, args, duration, nil)

	return row
}

// Exec 在事务中执行SQL语句
func (t *BaseTransaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	return t.ExecContext(context.Background(), query, args...)
}

// ExecContext 在事务中执行带上下文的SQL语句
func (t *BaseTransaction) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if t.tx == nil {
		return nil, NewError(ErrCodeTransactionNotStarted, "transaction not active")
	}

	start := time.Now()
	result, err := t.tx.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	t.logQuery(query, args, duration, err)

	if err != nil {
		return nil, NewQueryError(err, query, args)
	}

	return result, nil
}

// Commit 提交事务
func (t *BaseTransaction) Commit() error {
	if t.tx == nil {
		return NewError(ErrCodeTransactionNotStarted, "transaction not active")
	}

	err := t.tx.Commit()
	if err != nil {
		return WrapError(err, ErrCodeTransactionCommitFailed, "failed to commit transaction")
	}

	if t.logger != nil {
		t.logger.Debug("Transaction committed")
	}

	t.tx = nil
	return nil
}

// Rollback 回滚事务
func (t *BaseTransaction) Rollback() error {
	if t.tx == nil {
		return NewError(ErrCodeTransactionNotStarted, "transaction not active")
	}

	err := t.tx.Rollback()
	if err != nil {
		return WrapError(err, ErrCodeTransactionRollbackFailed, "failed to rollback transaction")
	}

	if t.logger != nil {
		t.logger.Debug("Transaction rolled back")
	}

	t.tx = nil
	return nil
}

// Savepoint 创建保存点（嵌套事务）
func (t *BaseTransaction) Savepoint(name string) error {
	if t.tx == nil {
		return NewError(ErrCodeTransactionNotStarted, "transaction not active")
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	_, err := t.tx.Exec(fmt.Sprintf("SAVEPOINT %s", name))
	if err != nil {
		return WrapErrorf(err, ErrCodeTransactionFailed, "failed to create savepoint: %s", name)
	}

	t.savepointCount++

	if t.logger != nil {
		t.logger.Debug("Savepoint created", "name", name)
	}

	return nil
}

// RollbackTo 回滚到保存点
func (t *BaseTransaction) RollbackTo(name string) error {
	if t.tx == nil {
		return NewError(ErrCodeTransactionNotStarted, "transaction not active")
	}

	_, err := t.tx.Exec(fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", name))
	if err != nil {
		return WrapErrorf(err, ErrCodeTransactionRollbackFailed, "failed to rollback to savepoint: %s", name)
	}

	if t.logger != nil {
		t.logger.Debug("Rolled back to savepoint", "name", name)
	}

	return nil
}

// ReleaseSavepoint 释放保存点
func (t *BaseTransaction) ReleaseSavepoint(name string) error {
	if t.tx == nil {
		return NewError(ErrCodeTransactionNotStarted, "transaction not active")
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	_, err := t.tx.Exec(fmt.Sprintf("RELEASE SAVEPOINT %s", name))
	if err != nil {
		return WrapErrorf(err, ErrCodeTransactionFailed, "failed to release savepoint: %s", name)
	}

	if t.savepointCount > 0 {
		t.savepointCount--
	}

	if t.logger != nil {
		t.logger.Debug("Savepoint released", "name", name)
	}

	return nil
}

// logQuery 记录查询日志
func (t *BaseTransaction) logQuery(sql string, args []interface{}, duration time.Duration, err error) {
	if t.logger == nil {
		return
	}

	if err != nil {
		t.logger.Error("SQL执行失败(事务)",
			"sql", sql,
			"args", args,
			"duration", duration,
			"error", err)
	} else if t.config != nil && t.config.LogQueries {
		t.logger.Debug("SQL执行(事务)",
			"sql", sql,
			"args", args,
			"duration", duration)
	}
}
