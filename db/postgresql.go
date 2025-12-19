package db

import (
	"context"
	"database/sql"
	"sync"
	"time"

	_ "github.com/lib/pq" // PostgreSQL 驱动
	"github.com/zhoudm1743/torm/logger"
)

// PostgreSQLConnection PostgreSQL数据库连接
type PostgreSQLConnection struct {
	*BaseConnection
	mu sync.RWMutex
}

// NewPostgreSQLConnection 创建PostgreSQL连接
func NewPostgreSQLConnection(config *Config, log LoggerInterface) (Connection, error) {
	if config == nil {
		return nil, NewError(ErrCodeConfigurationError, "config cannot be nil")
	}

	conn := &PostgreSQLConnection{
		BaseConnection: NewBaseConnection(config),
	}
	conn.SetLogger(log)

	return conn, nil
}

// Connect 连接到PostgreSQL数据库
func (c *PostgreSQLConnection) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.IsConnected() {
		return nil
	}

	config := c.GetConfig()
	dsn := config.DSN()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return NewConnectionError(err, "postgres", config.Host, config.Database)
	}

	// 配置连接池
	c.SetDB(db)
	c.ConfigurePool()

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		c.SetDB(nil)
		return NewConnectionError(err, "postgres", config.Host, config.Database)
	}

	// 设置PostgreSQL构建器
	c.SetBuilder(NewPostgreSQLBuilder())

	if c.GetLogger() != nil {
		c.GetLogger().Info("PostgreSQL连接建立成功",
			"host", config.Host,
			"database", config.Database)
	}

	return nil
}

// Close 关闭连接
func (c *PostgreSQLConnection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	log := c.GetLogger()
	err := c.BaseConnection.Close()

	if log != nil {
		if err != nil {
			log.Error("关闭PostgreSQL连接失败", "error", err)
		} else {
			log.Info("PostgreSQL连接已关闭")
		}
	}

	return err
}

// GetDriver 获取驱动名称
func (c *PostgreSQLConnection) GetDriver() string {
	return "postgres"
}

// Query 执行查询
func (c *PostgreSQLConnection) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return c.QueryContext(context.Background(), query, args...)
}

// QueryContext 执行带上下文的查询
func (c *PostgreSQLConnection) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
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
func (c *PostgreSQLConnection) QueryRow(query string, args ...interface{}) *sql.Row {
	return c.QueryRowContext(context.Background(), query, args...)
}

// QueryRowContext 执行带上下文的单行查询
func (c *PostgreSQLConnection) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
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
func (c *PostgreSQLConnection) Exec(query string, args ...interface{}) (sql.Result, error) {
	return c.ExecContext(context.Background(), query, args...)
}

// ExecContext 执行带上下文的SQL语句
func (c *PostgreSQLConnection) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
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
func (c *PostgreSQLConnection) Begin() (TxInterface, error) {
	return c.BeginTx(context.Background(), nil)
}

// BeginTx 开始事务（带选项）
func (c *PostgreSQLConnection) BeginTx(ctx context.Context, opts *sql.TxOptions) (TxInterface, error) {
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
		c.GetLogger().Debug("PostgreSQL事务已开始")
	}

	return NewPostgreSQLTransaction(tx, c.GetLogger(), c.GetConfig()), nil
}

// RunTransaction 执行事务
func (c *PostgreSQLConnection) RunTransaction(fn func(tx TxInterface) error) error {
	return c.BaseConnection.RunTransaction(fn, c.Begin)
}

// GetTableFields 获取表字段信息
func (c *PostgreSQLConnection) GetTableFields(table string) ([]ConnFieldInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	db := c.GetDB()
	if db == nil {
		return nil, NewError(ErrCodeConnectionClosed, "connection not established")
	}

	query := `
		SELECT 
			c.column_name,
			c.data_type,
			c.is_nullable,
			c.column_default,
			CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END as is_primary,
			CASE WHEN c.column_default LIKE 'nextval%' THEN true ELSE false END as is_auto_inc,
			COALESCE(pgd.description, '') as column_comment,
			c.character_maximum_length,
			c.numeric_precision,
			c.numeric_scale
		FROM information_schema.columns c
		LEFT JOIN (
			SELECT ku.column_name
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage ku
				ON tc.constraint_name = ku.constraint_name
			WHERE tc.constraint_type = 'PRIMARY KEY'
				AND tc.table_name = $1
		) pk ON c.column_name = pk.column_name
		LEFT JOIN pg_catalog.pg_statio_all_tables st
			ON c.table_schema = st.schemaname AND c.table_name = st.relname
		LEFT JOIN pg_catalog.pg_description pgd
			ON pgd.objoid = st.relid AND pgd.objsubid = c.ordinal_position
		WHERE c.table_name = $1
		ORDER BY c.ordinal_position
	`

	rows, err := db.Query(query, table)
	if err != nil {
		return nil, NewQueryError(err, query, []interface{}{table})
	}
	defer rows.Close()

	var fields []ConnFieldInfo
	for rows.Next() {
		var (
			name, dataType, nullable string
			defaultVal, comment      sql.NullString
			isPrimary, isAutoInc     bool
			length, precision, scale sql.NullInt64
		)

		err := rows.Scan(&name, &dataType, &nullable, &defaultVal, &isPrimary, &isAutoInc, &comment, &length, &precision, &scale)
		if err != nil {
			return nil, WrapError(err, ErrCodeQueryFailed, "failed to scan column info")
		}

		field := ConnFieldInfo{
			Name:       name,
			Type:       dataType,
			Nullable:   nullable == "YES",
			PrimaryKey: isPrimary,
			AutoInc:    isAutoInc,
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
func (c *PostgreSQLConnection) GetTables() ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	db := c.GetDB()
	if db == nil {
		return nil, NewError(ErrCodeConnectionClosed, "connection not established")
	}

	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`

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
func (c *PostgreSQLConnection) logSQL(query string, args []interface{}, duration time.Duration, err error) {
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

// PostgreSQLTransaction PostgreSQL事务
type PostgreSQLTransaction struct {
	*BaseTransaction
}

// NewPostgreSQLTransaction 创建PostgreSQL事务
func NewPostgreSQLTransaction(tx *sql.Tx, log LoggerInterface, config *Config) *PostgreSQLTransaction {
	return &PostgreSQLTransaction{
		BaseTransaction: NewBaseTransaction(tx, log, config),
	}
}

// NewPostgreSQLBuilder 创建PostgreSQL构建器（使用完整实现）
func NewPostgreSQLBuilder() Builder {
	return NewPostgreSQLBuilderImpl()
}
