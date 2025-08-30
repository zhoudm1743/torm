package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/zhoudm1743/torm/logger"
	_ "modernc.org/sqlite" // SQLite 驱动
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
func (c *SQLiteConnection) Connect() error {
	start := time.Now()
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
	if err := db.Ping(); err != nil {
		db.Close()
		if c.logger != nil {
			c.logger.Error("Failed to ping SQLite database", "error", err)
		}
		return fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	c.db = db
	duration := time.Since(start)
	if c.logger != nil {
		// 使用新的LogConnection方法记录连接信息
		if sqlLogger, ok := c.logger.(*logger.SQLLogger); ok {
			sqlLogger.LogConnection("connected", "sqlite", c.config.Database, duration)
		} else {
			c.logger.Info("SQLite连接建立成功", "database", c.config.Database, "duration", duration)
		}
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
func (c *SQLiteConnection) Ping() error {
	if c.db == nil {
		return fmt.Errorf("database connection is not established")
	}
	return c.db.Ping()
}

// IsConnected 检查连接状态
func (c *SQLiteConnection) IsConnected() bool {
	if c.db == nil {
		return false
	}
	return c.db.Ping() == nil
}

// Query 执行查询
func (c *SQLiteConnection) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	start := time.Now()
	rows, err := c.db.Query(query, args...)
	duration := time.Since(start)

	// 统一SQL日志记录
	if c.logger != nil {
		if sqlLogger, ok := c.logger.(*logger.SQLLogger); ok {
			// Query操作我们不知道确切的行数，使用简单版本
			sqlLogger.LogSQL(query, args, duration, err)
		} else {
			// 兼容旧的日志接口
			if err != nil {
				c.logger.Error("SQL查询错误", "sql", query, "args", args, "error", err, "duration", duration)
			} else if c.config.LogQueries {
				c.logger.Debug("SQL查询", "sql", query, "args", args, "duration", duration)
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute SQLite query: %w", err)
	}
	return rows, nil
}

// QueryRow 执行单行查询
func (c *SQLiteConnection) QueryRow(query string, args ...interface{}) *sql.Row {
	if c.db == nil {
		// 返回一个会出错的Row
		return (&sql.DB{}).QueryRow(query, args...)
	}

	start := time.Now()
	row := c.db.QueryRow(query, args...)
	duration := time.Since(start)

	// 统一SQL日志记录（QueryRow总是成功，没有error）
	if c.logger != nil {
		if sqlLogger, ok := c.logger.(*logger.SQLLogger); ok {
			sqlLogger.LogSQL(query, args, duration, nil)
		} else if c.config.LogQueries {
			c.logger.Debug("SQL查询", "sql", query, "args", args, "duration", duration)
		}
	}

	return row
}

// Exec 执行SQL语句
func (c *SQLiteConnection) Exec(query string, args ...interface{}) (sql.Result, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	start := time.Now()
	result, err := c.db.Exec(query, args...)
	duration := time.Since(start)

	// 统一SQL日志记录
	if c.logger != nil {
		if sqlLogger, ok := c.logger.(*logger.SQLLogger); ok {
			if err != nil {
				sqlLogger.LogSQL(query, args, duration, err)
			} else {
				// 获取影响的行数
				rowsAffected, _ := result.RowsAffected()
				sqlLogger.LogSQLWithRows(query, args, duration, rowsAffected, nil)
			}
		} else {
			// 兼容旧的日志接口
			if err != nil {
				c.logger.Error("SQL执行错误", "sql", query, "args", args, "error", err, "duration", duration)
			} else if c.config.LogQueries {
				c.logger.Debug("SQL执行", "sql", query, "args", args, "duration", duration)
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute SQLite statement: %w", err)
	}
	return result, nil
}

// Begin 开始事务
func (c *SQLiteConnection) Begin() (TransactionInterface, error) {
	return c.BeginTx(nil)
}

// BeginTx 开始事务（带选项）
func (c *SQLiteConnection) BeginTx(opts *sql.TxOptions) (TransactionInterface, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	tx, err := c.db.Begin()
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
func (t *SQLiteTransaction) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if t.tx == nil {
		return nil, fmt.Errorf("transaction is not active")
	}

	start := time.Now()
	rows, err := t.tx.Query(query, args...)
	duration := time.Since(start)

	if t.logger != nil && t.config.LogQueries {
		if err != nil {
			t.logger.Error("sql", query, "args", args, "error", err, "duration", duration)
		} else {
			t.logger.Debug("sql", query, "args", args, "duration", duration)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute SQLite transaction query: %w", err)
	}
	return rows, nil
}

// QueryRow 在事务中执行单行查询
func (t *SQLiteTransaction) QueryRow(query string, args ...interface{}) *sql.Row {
	if t.tx == nil {
		// 返回一个会出错的Row
		return (&sql.DB{}).QueryRow(query, args...)
	}

	start := time.Now()
	row := t.tx.QueryRow(query, args...)
	duration := time.Since(start)

	if t.logger != nil && t.config.LogQueries {
		t.logger.Debug("sql", query, "args", args, "duration", duration)
	}

	return row
}

// Exec 在事务中执行SQL语句
func (t *SQLiteTransaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	if t.tx == nil {
		return nil, fmt.Errorf("transaction is not active")
	}

	start := time.Now()
	result, err := t.tx.Exec(query, args...)
	duration := time.Since(start)

	if t.logger != nil && t.config.LogQueries {
		if err != nil {
			t.logger.Error("sql", query, "args", args, "error", err, "duration", duration)
		} else {
			t.logger.Debug("sql", query, "args", args, "duration", duration)
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

// GetDB 获取底层数据库连接
func (c *SQLiteConnection) GetDB() *sql.DB {
	return c.db
}
