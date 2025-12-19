package db

import (
	"context"
	"database/sql"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zhoudm1743/torm/logger"
)

// MySQLConnection MySQL数据库连接
type MySQLConnection struct {
	*BaseConnection
	mu sync.RWMutex
}

// NewMySQLConnection 创建MySQL连接
func NewMySQLConnection(config *Config, log LoggerInterface) (Connection, error) {
	if config == nil {
		return nil, NewError(ErrCodeConfigurationError, "config cannot be nil")
	}

	conn := &MySQLConnection{
		BaseConnection: NewBaseConnection(config),
	}
	conn.SetLogger(log)

	return conn, nil
}

// Connect 连接到数据库
func (c *MySQLConnection) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.IsConnected() {
		return nil
	}

	config := c.GetConfig()
	dsn := config.DSN()

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return NewConnectionError(err, "mysql", config.Host, config.Database)
	}

	// 配置连接池
	c.SetDB(db)
	c.ConfigurePool()

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		c.SetDB(nil)
		return NewConnectionError(err, "mysql", config.Host, config.Database)
	}

	// 设置MySQL构建器
	c.SetBuilder(NewMySQLBuilder())

	if c.GetLogger() != nil {
		c.GetLogger().Info("MySQL连接建立成功",
			"host", config.Host,
			"database", config.Database)
	}

	return nil
}

// Close 关闭连接
func (c *MySQLConnection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	log := c.GetLogger()
	err := c.BaseConnection.Close()

	if log != nil {
		if err != nil {
			log.Error("关闭MySQL连接失败", "error", err)
		} else {
			log.Info("MySQL连接已关闭")
		}
	}

	return err
}

// GetDriver 获取驱动名称
func (c *MySQLConnection) GetDriver() string {
	return "mysql"
}

// Query 执行查询
func (c *MySQLConnection) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return c.QueryContext(context.Background(), query, args...)
}

// QueryContext 执行带上下文的查询
func (c *MySQLConnection) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
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
func (c *MySQLConnection) QueryRow(query string, args ...interface{}) *sql.Row {
	return c.QueryRowContext(context.Background(), query, args...)
}

// QueryRowContext 执行带上下文的单行查询
func (c *MySQLConnection) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
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
func (c *MySQLConnection) Exec(query string, args ...interface{}) (sql.Result, error) {
	return c.ExecContext(context.Background(), query, args...)
}

// ExecContext 执行带上下文的SQL语句
func (c *MySQLConnection) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
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
func (c *MySQLConnection) Begin() (TxInterface, error) {
	return c.BeginTx(context.Background(), nil)
}

// BeginTx 开始事务（带选项）
func (c *MySQLConnection) BeginTx(ctx context.Context, opts *sql.TxOptions) (TxInterface, error) {
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
		c.GetLogger().Debug("MySQL事务已开始")
	}

	return NewMySQLTransaction(tx, c.GetLogger(), c.GetConfig()), nil
}

// RunTransaction 执行事务
func (c *MySQLConnection) RunTransaction(fn func(tx TxInterface) error) error {
	return c.BaseConnection.RunTransaction(fn, c.Begin)
}

// GetTableFields 获取表字段信息
func (c *MySQLConnection) GetTableFields(table string) ([]ConnFieldInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	db := c.GetDB()
	if db == nil {
		return nil, NewError(ErrCodeConnectionClosed, "connection not established")
	}

	query := `
		SELECT 
			COLUMN_NAME,
			DATA_TYPE,
			IS_NULLABLE,
			COLUMN_DEFAULT,
			COLUMN_KEY,
			EXTRA,
			COLUMN_COMMENT,
			CHARACTER_MAXIMUM_LENGTH,
			NUMERIC_PRECISION,
			NUMERIC_SCALE
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	rows, err := db.Query(query, table)
	if err != nil {
		return nil, NewQueryError(err, query, []interface{}{table})
	}
	defer rows.Close()

	var fields []ConnFieldInfo
	for rows.Next() {
		var (
			name, dataType, nullable, columnKey, extra string
			defaultVal, comment                        sql.NullString
			length, precision, scale                   sql.NullInt64
		)

		err := rows.Scan(&name, &dataType, &nullable, &defaultVal, &columnKey, &extra, &comment, &length, &precision, &scale)
		if err != nil {
			return nil, WrapError(err, ErrCodeQueryFailed, "failed to scan column info")
		}

		field := ConnFieldInfo{
			Name:       name,
			Type:       dataType,
			Nullable:   nullable == "YES",
			PrimaryKey: columnKey == "PRI",
			AutoInc:    extra == "auto_increment",
		}

		if defaultVal.Valid {
			field.Default = defaultVal.String
		}
		if comment.Valid {
			field.Comment = comment.String
		}
		if length.Valid {
			field.Length = int(length.Int64)
		}
		if precision.Valid {
			field.Precision = int(precision.Int64)
		}
		if scale.Valid {
			field.Scale = int(scale.Int64)
		}

		fields = append(fields, field)
	}

	return fields, nil
}

// GetTables 获取所有表名
func (c *MySQLConnection) GetTables() ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	db := c.GetDB()
	if db == nil {
		return nil, NewError(ErrCodeConnectionClosed, "connection not established")
	}

	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, NewQueryError(err, "SHOW TABLES", nil)
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
func (c *MySQLConnection) logSQL(query string, args []interface{}, duration time.Duration, err error) {
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

// MySQLTransaction MySQL事务
type MySQLTransaction struct {
	*BaseTransaction
}

// NewMySQLTransaction 创建MySQL事务
func NewMySQLTransaction(tx *sql.Tx, log LoggerInterface, config *Config) *MySQLTransaction {
	return &MySQLTransaction{
		BaseTransaction: NewBaseTransaction(tx, log, config),
	}
}

// NewMySQLBuilder 创建MySQL构建器（使用完整实现）
func NewMySQLBuilder() Builder {
	return NewMySQLBuilderImpl()
}
