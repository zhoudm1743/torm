package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL 驱动
)

// PostgreSQLConnection PostgreSQL数据库连接
type PostgreSQLConnection struct {
	config *Config
	db     *sql.DB
	logger LoggerInterface
}

// NewPostgreSQLConnection 创建PostgreSQL连接
func NewPostgreSQLConnection(config *Config, logger LoggerInterface) (ConnectionInterface, error) {
	conn := &PostgreSQLConnection{
		config: config,
		logger: logger,
	}
	return conn, nil
}

// Connect 连接到PostgreSQL数据库
func (c *PostgreSQLConnection) Connect(ctx context.Context) error {
	dsn := c.config.DSN()
	if c.logger != nil {
		c.logger.Debug("Connecting to PostgreSQL", "dsn", dsn)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Failed to open PostgreSQL connection", "error", err)
		}
		return fmt.Errorf("failed to open PostgreSQL connection: %w", err)
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
			c.logger.Error("Failed to ping PostgreSQL database", "error", err)
		}
		return fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	c.db = db
	if c.logger != nil {
		c.logger.Info("PostgreSQL connection established successfully")
	}
	return nil
}

// Close 关闭连接
func (c *PostgreSQLConnection) Close() error {
	if c.db != nil {
		if c.logger != nil {
			c.logger.Debug("Closing PostgreSQL connection")
		}
		err := c.db.Close()
		c.db = nil
		return err
	}
	return nil
}

// Ping 测试连接
func (c *PostgreSQLConnection) Ping(ctx context.Context) error {
	if c.db == nil {
		return fmt.Errorf("database connection is not established")
	}
	return c.db.PingContext(ctx)
}

// IsConnected 检查连接状态
func (c *PostgreSQLConnection) IsConnected() bool {
	if c.db == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return c.db.PingContext(ctx) == nil
}

// Query 执行查询
func (c *PostgreSQLConnection) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	start := time.Now()
	rows, err := c.db.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	if c.logger != nil && c.config.LogQueries {
		if err != nil {
			c.logger.Error("PostgreSQL query failed", "query", query, "args", args, "error", err, "duration", duration)
		} else {
			c.logger.Debug("PostgreSQL query executed", "query", query, "args", args, "duration", duration)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute PostgreSQL query: %w", err)
	}
	return rows, nil
}

// QueryRow 执行单行查询
func (c *PostgreSQLConnection) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if c.db == nil {
		// 返回一个会出错的Row
		return (&sql.DB{}).QueryRowContext(ctx, query, args...)
	}

	start := time.Now()
	row := c.db.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	if c.logger != nil && c.config.LogQueries {
		c.logger.Debug("PostgreSQL query row executed", "query", query, "args", args, "duration", duration)
	}

	return row
}

// Exec 执行SQL语句
func (c *PostgreSQLConnection) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	start := time.Now()
	result, err := c.db.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	if c.logger != nil && c.config.LogQueries {
		if err != nil {
			c.logger.Error("PostgreSQL exec failed", "query", query, "args", args, "error", err, "duration", duration)
		} else {
			c.logger.Debug("PostgreSQL exec executed", "query", query, "args", args, "duration", duration)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute PostgreSQL statement: %w", err)
	}
	return result, nil
}

// Begin 开始事务
func (c *PostgreSQLConnection) Begin(ctx context.Context) (TransactionInterface, error) {
	return c.BeginTx(ctx, nil)
}

// BeginTx 开始事务（带选项）
func (c *PostgreSQLConnection) BeginTx(ctx context.Context, opts *sql.TxOptions) (TransactionInterface, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	tx, err := c.db.BeginTx(ctx, opts)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Failed to begin PostgreSQL transaction", "error", err)
		}
		return nil, fmt.Errorf("failed to begin PostgreSQL transaction: %w", err)
	}

	if c.logger != nil {
		c.logger.Debug("PostgreSQL transaction started")
	}

	return &PostgreSQLTransaction{
		tx:     tx,
		logger: c.logger,
		config: c.config,
	}, nil
}

// GetConfig 获取配置
func (c *PostgreSQLConnection) GetConfig() *Config {
	return c.config
}

// GetDriver 获取驱动名称
func (c *PostgreSQLConnection) GetDriver() string {
	return "postgres"
}

// GetStats 获取连接统计信息
func (c *PostgreSQLConnection) GetStats() sql.DBStats {
	if c.db == nil {
		return sql.DBStats{}
	}
	return c.db.Stats()
}

// PostgreSQLTransaction PostgreSQL事务
type PostgreSQLTransaction struct {
	tx     *sql.Tx
	logger LoggerInterface
	config *Config
}

// Query 在事务中执行查询
func (t *PostgreSQLTransaction) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if t.tx == nil {
		return nil, fmt.Errorf("transaction is not active")
	}

	start := time.Now()
	rows, err := t.tx.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	if t.logger != nil && t.config.LogQueries {
		if err != nil {
			t.logger.Error("PostgreSQL transaction query failed", "query", query, "args", args, "error", err, "duration", duration)
		} else {
			t.logger.Debug("PostgreSQL transaction query executed", "query", query, "args", args, "duration", duration)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute PostgreSQL transaction query: %w", err)
	}
	return rows, nil
}

// QueryRow 在事务中执行单行查询
func (t *PostgreSQLTransaction) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if t.tx == nil {
		// 返回一个会出错的Row
		return (&sql.DB{}).QueryRowContext(ctx, query, args...)
	}

	start := time.Now()
	row := t.tx.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	if t.logger != nil && t.config.LogQueries {
		t.logger.Debug("PostgreSQL transaction query row executed", "query", query, "args", args, "duration", duration)
	}

	return row
}

// Exec 在事务中执行SQL语句
func (t *PostgreSQLTransaction) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if t.tx == nil {
		return nil, fmt.Errorf("transaction is not active")
	}

	start := time.Now()
	result, err := t.tx.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	if t.logger != nil && t.config.LogQueries {
		if err != nil {
			t.logger.Error("PostgreSQL transaction exec failed", "query", query, "args", args, "error", err, "duration", duration)
		} else {
			t.logger.Debug("PostgreSQL transaction exec executed", "query", query, "args", args, "duration", duration)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute PostgreSQL transaction statement: %w", err)
	}
	return result, nil
}

// Commit 提交事务
func (t *PostgreSQLTransaction) Commit() error {
	if t.tx == nil {
		return fmt.Errorf("transaction is not active")
	}

	err := t.tx.Commit()
	if err != nil {
		if t.logger != nil {
			t.logger.Error("Failed to commit PostgreSQL transaction", "error", err)
		}
		return fmt.Errorf("failed to commit PostgreSQL transaction: %w", err)
	}

	if t.logger != nil {
		t.logger.Debug("PostgreSQL transaction committed")
	}

	t.tx = nil
	return nil
}

// Rollback 回滚事务
func (t *PostgreSQLTransaction) Rollback() error {
	if t.tx == nil {
		return fmt.Errorf("transaction is not active")
	}

	err := t.tx.Rollback()
	if err != nil {
		if t.logger != nil {
			t.logger.Error("Failed to rollback PostgreSQL transaction", "error", err)
		}
		return fmt.Errorf("failed to rollback PostgreSQL transaction: %w", err)
	}

	if t.logger != nil {
		t.logger.Debug("PostgreSQL transaction rolled back")
	}

	t.tx = nil
	return nil
}
