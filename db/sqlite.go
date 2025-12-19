package db

import (
	"context"
	"database/sql"
	"path/filepath"
	"sync"
	"time"

	"github.com/zhoudm1743/torm/logger"
	_ "modernc.org/sqlite" // SQLite 驱动
)

// SQLiteConnection SQLite数据库连接
type SQLiteConnection struct {
	*BaseConnection
	mu sync.RWMutex
}

// NewSQLiteConnection 创建SQLite连接
func NewSQLiteConnection(config *Config, log LoggerInterface) (Connection, error) {
	if config == nil {
		return nil, NewError(ErrCodeConfigurationError, "config cannot be nil")
	}

	conn := &SQLiteConnection{
		BaseConnection: NewBaseConnection(config),
	}
	conn.SetLogger(log)

	return conn, nil
}

// Connect 连接到SQLite数据库
func (c *SQLiteConnection) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.IsConnected() {
		return nil
	}

	config := c.GetConfig()
	dsn := c.buildDSN(config)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return NewConnectionError(err, "sqlite", "", config.Database)
	}

	// 配置连接池（SQLite建议单连接）
	c.SetDB(db)
	if config.MaxOpenConns == 0 {
		db.SetMaxOpenConns(1)
	} else {
		c.ConfigurePool()
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		c.SetDB(nil)
		return NewConnectionError(err, "sqlite", "", config.Database)
	}

	// 启用外键约束
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		c.GetLogger().Warn("启用外键约束失败", "error", err)
	}

	// 设置SQLite构建器
	c.SetBuilder(NewSQLiteBuilder())

	if c.GetLogger() != nil {
		c.GetLogger().Info("SQLite连接建立成功",
			"database", config.Database)
	}

	return nil
}

// buildDSN 构建SQLite DSN
func (c *SQLiteConnection) buildDSN(config *Config) string {
	dsn := config.Database

	// 处理相对路径
	if dsn != ":memory:" && !filepath.IsAbs(dsn) {
		// 保持相对路径，让调用者决定工作目录
	}

	// 添加参数
	if len(config.Options) > 0 {
		params := ""
		for k, v := range config.Options {
			if params == "" {
				params = "?"
			} else {
				params += "&"
			}
			params += k + "=" + v
		}
		dsn += params
	}

	return dsn
}

// Close 关闭连接
func (c *SQLiteConnection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	log := c.GetLogger()
	err := c.BaseConnection.Close()

	if log != nil {
		if err != nil {
			log.Error("关闭SQLite连接失败", "error", err)
		} else {
			log.Info("SQLite连接已关闭")
		}
	}

	return err
}

// GetDriver 获取驱动名称
func (c *SQLiteConnection) GetDriver() string {
	return "sqlite"
}

// Query 执行查询
func (c *SQLiteConnection) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return c.QueryContext(context.Background(), query, args...)
}

// QueryContext 执行带上下文的查询
func (c *SQLiteConnection) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	db := c.GetDB()
	if db == nil {
		return nil, NewError(ErrCodeConnectionClosed, "connection not established")
	}

	start := time.Now()
	rows, err := db.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	c.logSQL(query, args, duration, err)

	if err != nil {
		return nil, NewQueryError(err, query, args)
	}

	return rows, nil
}

// QueryRow 执行单行查询
func (c *SQLiteConnection) QueryRow(query string, args ...interface{}) *sql.Row {
	return c.QueryRowContext(context.Background(), query, args...)
}

// QueryRowContext 执行带上下文的单行查询
func (c *SQLiteConnection) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	c.mu.RLock()
	defer c.mu.RUnlock()

	db := c.GetDB()
	if db == nil {
		return nil
	}

	start := time.Now()
	row := db.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	c.logSQL(query, args, duration, nil)

	return row
}

// Exec 执行SQL语句
func (c *SQLiteConnection) Exec(query string, args ...interface{}) (sql.Result, error) {
	return c.ExecContext(context.Background(), query, args...)
}

// ExecContext 执行带上下文的SQL语句
func (c *SQLiteConnection) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	db := c.GetDB()
	if db == nil {
		return nil, NewError(ErrCodeConnectionClosed, "connection not established")
	}

	start := time.Now()
	result, err := db.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	c.logSQL(query, args, duration, err)

	if err != nil {
		return nil, NewQueryError(err, query, args)
	}

	return result, nil
}

// Begin 开始事务
func (c *SQLiteConnection) Begin() (TxInterface, error) {
	return c.BeginTx(context.Background(), nil)
}

// BeginTx 开始事务（带选项）
func (c *SQLiteConnection) BeginTx(ctx context.Context, opts *sql.TxOptions) (TxInterface, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	db := c.GetDB()
	if db == nil {
		return nil, NewError(ErrCodeConnectionClosed, "connection not established")
	}

	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return nil, NewTransactionError(err, "begin")
	}

	if c.GetLogger() != nil {
		c.GetLogger().Debug("SQLite事务已开始")
	}

	return NewSQLiteTransaction(tx, c.GetLogger(), c.GetConfig()), nil
}

// RunTransaction 执行事务
func (c *SQLiteConnection) RunTransaction(fn func(tx TxInterface) error) error {
	return c.BaseConnection.RunTransaction(fn, c.Begin)
}

// GetTableFields 获取表字段信息
func (c *SQLiteConnection) GetTableFields(table string) ([]ConnFieldInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	db := c.GetDB()
	if db == nil {
		return nil, NewError(ErrCodeConnectionClosed, "connection not established")
	}

	query := "PRAGMA table_info(" + table + ")"
	rows, err := db.Query(query)
	if err != nil {
		return nil, NewQueryError(err, query, nil)
	}
	defer rows.Close()

	var fields []ConnFieldInfo
	for rows.Next() {
		var (
			cid        int
			name       string
			dataType   string
			notNull    int
			defaultVal sql.NullString
			pk         int
		)

		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultVal, &pk)
		if err != nil {
			return nil, WrapError(err, ErrCodeQueryFailed, "failed to scan column info")
		}

		field := ConnFieldInfo{
			Name:       name,
			Type:       dataType,
			Nullable:   notNull == 0,
			PrimaryKey: pk > 0,
			AutoInc:    pk > 0 && dataType == "INTEGER", // SQLite INTEGER PRIMARY KEY 自动自增
		}

		if defaultVal.Valid {
			field.Default = defaultVal.String
		}

		fields = append(fields, field)
	}

	return fields, nil
}

// GetTables 获取所有表名
func (c *SQLiteConnection) GetTables() ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	db := c.GetDB()
	if db == nil {
		return nil, NewError(ErrCodeConnectionClosed, "connection not established")
	}

	query := "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name"
	rows, err := db.Query(query)
	if err != nil {
		return nil, NewQueryError(err, query, nil)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, WrapError(err, ErrCodeQueryFailed, "failed to scan table name")
		}
		tables = append(tables, table)
	}

	return tables, nil
}

// logSQL 记录SQL日志
func (c *SQLiteConnection) logSQL(query string, args []interface{}, duration time.Duration, err error) {
	log := c.GetLogger()
	if log == nil {
		return
	}

	if sqlLogger, ok := log.(*logger.SQLLogger); ok {
		sqlLogger.LogSQL(query, args, duration, err)
	} else {
		config := c.GetConfig()
		if err != nil {
			log.Error("SQL执行失败",
				"sql", query,
				"args", args,
				"duration", duration,
				"error", err)
		} else if config != nil && config.LogQueries {
			log.Debug("SQL执行",
				"sql", query,
				"args", args,
				"duration", duration)
		}
	}
}

// SQLiteTransaction SQLite事务
type SQLiteTransaction struct {
	*BaseTransaction
}

// NewSQLiteTransaction 创建SQLite事务
func NewSQLiteTransaction(tx *sql.Tx, log LoggerInterface, config *Config) *SQLiteTransaction {
	return &SQLiteTransaction{
		BaseTransaction: NewBaseTransaction(tx, log, config),
	}
}

// Savepoint 创建保存点（SQLite支持）
func (t *SQLiteTransaction) Savepoint(name string) error {
	return t.BaseTransaction.Savepoint(name)
}

// RollbackTo 回滚到保存点
func (t *SQLiteTransaction) RollbackTo(name string) error {
	return t.BaseTransaction.RollbackTo(name)
}

// ReleaseSavepoint 释放保存点
func (t *SQLiteTransaction) ReleaseSavepoint(name string) error {
	return t.BaseTransaction.ReleaseSavepoint(name)
}

// NewSQLiteBuilder 创建SQLite构建器（使用完整实现）
func NewSQLiteBuilder() Builder {
	return NewSQLiteBuilderImpl()
}
