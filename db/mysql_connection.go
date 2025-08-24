package db

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLConnection MySQL数据库连接
type MySQLConnection struct {
	config    *Config
	db        *sql.DB
	logger    LoggerInterface
	connected bool
	mu        sync.RWMutex
}

// NewMySQLConnection 创建MySQL连接
func NewMySQLConnection(config *Config, logger LoggerInterface) (ConnectionInterface, error) {
	return &MySQLConnection{
		config: config,
		logger: logger,
	}, nil
}

// Connect 连接到数据库
func (c *MySQLConnection) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected && c.db != nil {
		return nil
	}

	dsn := c.config.DSN()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open mysql connection: %w", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(c.config.MaxOpenConns)
	db.SetMaxIdleConns(c.config.MaxIdleConns)
	db.SetConnMaxLifetime(c.config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(c.config.ConnMaxIdleTime)

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping mysql database: %w", err)
	}

	c.db = db
	c.connected = true

	if c.logger != nil {
		c.logger.Info("MySQL connection established", "dsn", dsn)
	}

	return nil
}

// Close 关闭连接
func (c *MySQLConnection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.db == nil {
		return nil
	}

	err := c.db.Close()
	c.db = nil
	c.connected = false

	if c.logger != nil {
		if err != nil {
			c.logger.Error("Failed to close MySQL connection", "error", err)
		} else {
			c.logger.Info("MySQL connection closed")
		}
	}

	return err
}

// Ping 测试连接
func (c *MySQLConnection) Ping() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.db == nil {
		return fmt.Errorf("connection not established")
	}

	return c.db.Ping()
}

// IsConnected 检查是否已连接
func (c *MySQLConnection) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected && c.db != nil
}

// Query 执行查询
func (c *MySQLConnection) Query(query string, args ...interface{}) (*sql.Rows, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.db == nil {
		return nil, fmt.Errorf("connection not established")
	}

	start := time.Now()
	rows, err := c.db.Query(query, args...)
	duration := time.Since(start)

	if c.logger != nil {
		if err != nil {
			c.logger.Error("sql", query, "args", args, "duration", duration, "error", err)
		} else if c.config.LogQueries {
			c.logger.Debug("sql", query, "args", args, "duration", duration)
		}
	}

	return rows, err
}

// QueryRow 执行单行查询
func (c *MySQLConnection) QueryRow(query string, args ...interface{}) *sql.Row {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.db == nil {
		return &sql.Row{}
	}

	start := time.Now()
	row := c.db.QueryRow(query, args...)
	duration := time.Since(start)

	if c.logger != nil && c.config.LogQueries {
		c.logger.Debug("sql", query, "args", args, "duration", duration)
	}

	return row
}

// Exec 执行SQL语句
func (c *MySQLConnection) Exec(query string, args ...interface{}) (sql.Result, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.db == nil {
		return nil, fmt.Errorf("connection not established")
	}

	start := time.Now()
	result, err := c.db.Exec(query, args...)
	duration := time.Since(start)

	if c.logger != nil {
		if err != nil {
			c.logger.Error("sql", query, "args", args, "duration", duration, "error", err)
		} else if c.config.LogQueries {
			c.logger.Debug("sql", query, "args", args, "duration", duration)
		}
	}

	return result, err
}

// Begin 开始事务
func (c *MySQLConnection) Begin() (TransactionInterface, error) {
	return c.BeginTx(nil)
}

// BeginTx 开始事务（带选项）
func (c *MySQLConnection) BeginTx(opts *sql.TxOptions) (TransactionInterface, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.db == nil {
		return nil, fmt.Errorf("connection not established")
	}

	tx, err := c.db.Begin()
	if err != nil {
		return nil, err
	}

	return &MySQLTransaction{
		tx:     tx,
		logger: c.logger,
		config: c.config,
	}, nil
}

// GetConfig 获取配置
func (c *MySQLConnection) GetConfig() *Config {
	return c.config
}

// GetDriver 获取驱动名称
func (c *MySQLConnection) GetDriver() string {
	return "mysql"
}

// GetStats 获取连接统计
func (c *MySQLConnection) GetStats() sql.DBStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.db == nil {
		return sql.DBStats{}
	}

	return c.db.Stats()
}

// MySQLTransaction MySQL事务
type MySQLTransaction struct {
	tx     *sql.Tx
	logger LoggerInterface
	config *Config
}

// Query 在事务中执行查询
func (t *MySQLTransaction) Query(query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := t.tx.Query(query, args...)
	duration := time.Since(start)

	if t.logger != nil {
		if err != nil {
			t.logger.Error("sql", query, "args", args, "duration", duration, "error", err)
		} else if t.config.LogQueries {
			t.logger.Debug("sql", query, "args", args, "duration", duration)
		}
	}

	return rows, err
}

// QueryRow 在事务中执行单行查询
func (t *MySQLTransaction) QueryRow(query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := t.tx.QueryRow(query, args...)
	duration := time.Since(start)

	if t.logger != nil && t.config.LogQueries {
		t.logger.Debug("sql", query, "args", args, "duration", duration)
	}

	return row
}

// Exec 在事务中执行SQL语句
func (t *MySQLTransaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := t.tx.Exec(query, args...)
	duration := time.Since(start)

	if t.logger != nil {
		if err != nil {
			t.logger.Error("sql", query, "args", args, "duration", duration, "error", err)
		} else if t.config.LogQueries {
			t.logger.Debug("sql", query, "args", args, "duration", duration)
		}
	}

	return result, err
}

// Commit 提交事务
func (t *MySQLTransaction) Commit() error {
	err := t.tx.Commit()
	if t.logger != nil {
		if err != nil {
			t.logger.Error("error", err)
		} else {
			t.logger.Debug("Transaction committed")
		}
	}
	return err
}

// Rollback 回滚事务
func (t *MySQLTransaction) Rollback() error {
	err := t.tx.Rollback()
	if t.logger != nil {
		if err != nil {
			t.logger.Error("Transaction rollback failed", "error", err)
		} else {
			t.logger.Debug("Transaction rolled back")
		}
	}
	return err
}

// 占位符连接器

// NewSQLServerConnection 创建SQL Server连接（占位符）
func NewSQLServerConnection(config *Config, logger LoggerInterface) (ConnectionInterface, error) {
	return nil, fmt.Errorf("SQL Server connector not implemented yet")
}

// GetDB 获取底层数据库连接
func (c *MySQLConnection) GetDB() *sql.DB {
	return c.db
}
