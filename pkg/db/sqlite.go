package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/glebarez/sqlite" // SQLite 驱动
)

// SQLiteConnection SQLite数据库连接
type SQLiteConnection struct {
	config *Config
	db     *sql.DB
	logger LoggerInterface
}

// NewSQLiteConnection 创建SQLite连接
func NewSQLiteConnection(config *Config, logger LoggerInterface) (ConnectionInterface, error) {
	conn := &SQLiteConnection{
		config: config,
		logger: logger,
	}
	return conn, nil
}

// Connect 连接到SQLite数据库
func (c *SQLiteConnection) Connect(ctx context.Context) error {
	dsn := c.config.DSN()
	if c.logger != nil {
		c.logger.Debug("Connecting to SQLite", "dsn", dsn)
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Failed to open SQLite connection", "error", err)
		}
		return fmt.Errorf("failed to open SQLite connection: %w", err)
	}

	// 配置连接池
	if c.config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(c.config.MaxOpenConns)
	}
	if c.config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(c.config.MaxIdleConns)
	}
	if c.config.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(c.config.ConnMaxLifetime)
	}
	if c.config.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(c.config.ConnMaxIdleTime)
	}

	// 测试连接
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		if c.logger != nil {
			c.logger.Error("Failed to ping SQLite database", "error", err)
		}
		return fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	c.db = db
	if c.logger != nil {
		c.logger.Info("SQLite connection established successfully")
	}
	return nil
}

// Close 关闭连接
func (c *SQLiteConnection) Close() error {
	if c.db != nil {
		if c.logger != nil {
			c.logger.Debug("Closing SQLite connection")
		}
		err := c.db.Close()
		c.db = nil
		return err
	}
	return nil
}

// Ping 测试连接
func (c *SQLiteConnection) Ping(ctx context.Context) error {
	if c.db == nil {
		return fmt.Errorf("database connection is not established")
	}
	return c.db.PingContext(ctx)
}

// IsConnected 检查连接状态
func (c *SQLiteConnection) IsConnected() bool {
	if c.db == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return c.db.PingContext(ctx) == nil
}

// Query 执行查询
func (c *SQLiteConnection) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	start := time.Now()
	rows, err := c.db.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	if c.logger != nil && c.config.LogQueries {
		if err != nil {
			c.logger.Error("SQLite query failed", "query", query, "args", args, "error", err, "duration", duration)
		} else {
			c.logger.Debug("SQLite query executed", "query", query, "args", args, "duration", duration)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute SQLite query: %w", err)
	}
	return rows, nil
}

// QueryRow 执行单行查询
func (c *SQLiteConnection) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if c.db == nil {
		// 返回一个会出错的Row
		return (&sql.DB{}).QueryRowContext(ctx, query, args...)
	}

	start := time.Now()
	row := c.db.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	if c.logger != nil && c.config.LogQueries {
		c.logger.Debug("SQLite query row executed", "query", query, "args", args, "duration", duration)
	}

	return row
}

// Exec 执行SQL语句
func (c *SQLiteConnection) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	start := time.Now()
	result, err := c.db.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	if c.logger != nil && c.config.LogQueries {
		if err != nil {
			c.logger.Error("SQLite exec failed", "query", query, "args", args, "error", err, "duration", duration)
		} else {
			c.logger.Debug("SQLite exec executed", "query", query, "args", args, "duration", duration)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute SQLite statement: %w", err)
	}
	return result, nil
}

// Begin 开始事务
func (c *SQLiteConnection) Begin(ctx context.Context) (TransactionInterface, error) {
	return c.BeginTx(ctx, nil)
}

// BeginTx 开始事务（带选项）
func (c *SQLiteConnection) BeginTx(ctx context.Context, opts *sql.TxOptions) (TransactionInterface, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	tx, err := c.db.BeginTx(ctx, opts)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Failed to begin SQLite transaction", "error", err)
		}
		return nil, fmt.Errorf("failed to begin SQLite transaction: %w", err)
	}

	if c.logger != nil {
		c.logger.Debug("SQLite transaction started")
	}

	return &SQLiteTransaction{
		tx:     tx,
		logger: c.logger,
		config: c.config,
	}, nil
}

// GetConfig 获取配置
func (c *SQLiteConnection) GetConfig() *Config {
	return c.config
}

// GetDriver 获取驱动名称
func (c *SQLiteConnection) GetDriver() string {
	return "sqlite"
}

// GetStats 获取连接统计信息
func (c *SQLiteConnection) GetStats() sql.DBStats {
	if c.db == nil {
		return sql.DBStats{}
	}
	return c.db.Stats()
}

// SQLiteTransaction SQLite事务
type SQLiteTransaction struct {
	tx     *sql.Tx
	logger LoggerInterface
	config *Config
}

// Query 在事务中执行查询
func (t *SQLiteTransaction) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if t.tx == nil {
		return nil, fmt.Errorf("transaction is not active")
	}

	start := time.Now()
	rows, err := t.tx.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	if t.logger != nil && t.config.LogQueries {
		if err != nil {
			t.logger.Error("SQLite transaction query failed", "query", query, "args", args, "error", err, "duration", duration)
		} else {
			t.logger.Debug("SQLite transaction query executed", "query", query, "args", args, "duration", duration)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute SQLite transaction query: %w", err)
	}
	return rows, nil
}

// QueryRow 在事务中执行单行查询
func (t *SQLiteTransaction) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if t.tx == nil {
		// 返回一个会出错的Row
		return (&sql.DB{}).QueryRowContext(ctx, query, args...)
	}

	start := time.Now()
	row := t.tx.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	if t.logger != nil && t.config.LogQueries {
		t.logger.Debug("SQLite transaction query row executed", "query", query, "args", args, "duration", duration)
	}

	return row
}

// Exec 在事务中执行SQL语句
func (t *SQLiteTransaction) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if t.tx == nil {
		return nil, fmt.Errorf("transaction is not active")
	}

	start := time.Now()
	result, err := t.tx.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	if t.logger != nil && t.config.LogQueries {
		if err != nil {
			t.logger.Error("SQLite transaction exec failed", "query", query, "args", args, "error", err, "duration", duration)
		} else {
			t.logger.Debug("SQLite transaction exec executed", "query", query, "args", args, "duration", duration)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute SQLite transaction statement: %w", err)
	}
	return result, nil
}

// Commit 提交事务
func (t *SQLiteTransaction) Commit() error {
	if t.tx == nil {
		return fmt.Errorf("transaction is not active")
	}

	err := t.tx.Commit()
	if err != nil {
		if t.logger != nil {
			t.logger.Error("Failed to commit SQLite transaction", "error", err)
		}
		return fmt.Errorf("failed to commit SQLite transaction: %w", err)
	}

	if t.logger != nil {
		t.logger.Debug("SQLite transaction committed")
	}

	t.tx = nil
	return nil
}

// Rollback 回滚事务
func (t *SQLiteTransaction) Rollback() error {
	if t.tx == nil {
		return fmt.Errorf("transaction is not active")
	}

	err := t.tx.Rollback()
	if err != nil {
		if t.logger != nil {
			t.logger.Error("Failed to rollback SQLite transaction", "error", err)
		}
		return fmt.Errorf("failed to rollback SQLite transaction: %w", err)
	}

	if t.logger != nil {
		t.logger.Debug("SQLite transaction rolled back")
	}

	t.tx = nil
	return nil
}
