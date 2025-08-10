package db

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// Manager 数据库管理器
type Manager struct {
	configs     map[string]*Config             // 数据库配置
	connections map[string]ConnectionInterface // 连接实例
	mu          sync.RWMutex                   // 读写锁
	logger      LoggerInterface                // 日志接口
	cache       CacheInterface                 // 缓存接口
	queryTimes  int64                          // 查询次数统计
}

// NewManager 创建数据库管理器
func NewManager() *Manager {
	return &Manager{
		configs:     make(map[string]*Config),
		connections: make(map[string]ConnectionInterface),
	}
}

// AddConfig 添加数据库配置
func (m *Manager) AddConfig(name string, config *Config) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid config for connection '%s': %w", name, err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.configs[name] = config
	return nil
}

// GetConfig 获取数据库配置
func (m *Manager) GetConfig(name string) (*Config, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	config, exists := m.configs[name]
	if !exists {
		return nil, fmt.Errorf("connection '%s' not configured", name)
	}

	return config, nil
}

// Connection 获取数据库连接
func (m *Manager) Connection(name string) (ConnectionInterface, error) {
	m.mu.RLock()
	conn, exists := m.connections[name]
	m.mu.RUnlock()

	if exists && conn.IsConnected() {
		return conn, nil
	}

	return m.createConnection(name)
}

// createConnection 创建数据库连接
func (m *Manager) createConnection(name string) (ConnectionInterface, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 双重检查
	if conn, exists := m.connections[name]; exists && conn.IsConnected() {
		return conn, nil
	}

	config, exists := m.configs[name]
	if !exists {
		return nil, fmt.Errorf("connection '%s' not configured", name)
	}

	var conn ConnectionInterface
	var err error

	switch config.Driver {
	case "mysql":
		conn, err = NewMySQLConnection(config, m.logger)
	case "postgres", "postgresql":
		conn, err = NewPostgreSQLConnection(config, m.logger)
	case "sqlite", "sqlite3":
		conn, err = NewSQLiteConnection(config, m.logger)
	case "sqlserver", "mssql":
		conn, err = NewSQLServerConnection(config, m.logger)
	case "mongodb", "mongo":
		conn, err = NewMongoDBConnection(config, m.logger)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", config.Driver)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create connection '%s': %w", name, err)
	}

	// 连接到数据库
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := conn.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to database '%s': %w", name, err)
	}

	m.connections[name] = conn
	return conn, nil
}

// Query 创建查询构造器
func (m *Manager) Query(connectionName ...string) (QueryInterface, error) {
	name := "default"
	if len(connectionName) > 0 {
		name = connectionName[0]
	}

	conn, err := m.Connection(name)
	if err != nil {
		return nil, err
	}

	return NewQuery(conn), nil
}

// Table 指定表名创建查询构造器
func (m *Manager) Table(tableName string, connectionName ...string) (QueryInterface, error) {
	query, err := m.Query(connectionName...)
	if err != nil {
		return nil, err
	}

	return query.From(tableName), nil
}

// Raw 执行原生SQL查询
func (m *Manager) Raw(ctx context.Context, sql string, bindings []interface{}, connectionName ...string) ([]map[string]interface{}, error) {
	name := "default"
	if len(connectionName) > 0 {
		name = connectionName[0]
	}

	conn, err := m.Connection(name)
	if err != nil {
		return nil, err
	}

	rows, err := conn.Query(ctx, sql, bindings...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return m.scanRows(rows)
}

// Exec 执行SQL语句
func (m *Manager) Exec(ctx context.Context, sql string, bindings []interface{}, connectionName ...string) (sql.Result, error) {
	name := "default"
	if len(connectionName) > 0 {
		name = connectionName[0]
	}

	conn, err := m.Connection(name)
	if err != nil {
		return nil, err
	}

	return conn.Exec(ctx, sql, bindings...)
}

// Transaction 执行事务
func (m *Manager) Transaction(ctx context.Context, fn func(tx TransactionInterface) error, connectionName ...string) error {
	name := "default"
	if len(connectionName) > 0 {
		name = connectionName[0]
	}

	conn, err := m.Connection(name)
	if err != nil {
		return err
	}

	tx, err := conn.Begin(ctx)
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
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// SetLogger 设置日志接口
func (m *Manager) SetLogger(logger LoggerInterface) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logger = logger
}

// SetCache 设置缓存接口
func (m *Manager) SetCache(cache CacheInterface) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cache = cache
}

// GetCache 获取缓存接口
func (m *Manager) GetCache() CacheInterface {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cache
}

// GetLogger 获取日志接口
func (m *Manager) GetLogger() LoggerInterface {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.logger
}

// Close 关闭所有连接
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for name, conn := range m.connections {
		if err := conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection '%s': %w", name, err))
		}
	}

	// 清空连接
	m.connections = make(map[string]ConnectionInterface)

	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}

	return nil
}

// GetStats 获取连接统计信息
func (m *Manager) GetStats() map[string]sql.DBStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]sql.DBStats)
	for name, conn := range m.connections {
		stats[name] = conn.GetStats()
	}

	return stats
}

// scanRows 扫描行数据
func (m *Manager) scanRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]

			// 处理NULL值
			if val == nil {
				row[col] = nil
				continue
			}

			// 根据数据库类型转换值
			switch columnTypes[i].DatabaseTypeName() {
			case "VARCHAR", "TEXT", "CHAR":
				if b, ok := val.([]byte); ok {
					row[col] = string(b)
				} else {
					row[col] = val
				}
			case "INT", "INTEGER", "BIGINT":
				row[col] = val
			case "FLOAT", "DOUBLE", "DECIMAL":
				row[col] = val
			case "BOOLEAN", "BOOL":
				row[col] = val
			case "TIMESTAMP", "DATETIME", "DATE", "TIME":
				row[col] = val
			default:
				row[col] = val
			}
		}

		results = append(results, row)
	}

	return results, rows.Err()
}

// 默认管理器实例
var defaultManager = NewManager()

// DefaultManager 获取默认管理器
func DefaultManager() *Manager {
	return defaultManager
}

// 便捷函数，使用默认管理器

// AddConnection 添加连接配置
func AddConnection(name string, config *Config) error {
	return defaultManager.AddConfig(name, config)
}

// DB 获取数据库连接
func DB(name ...string) (ConnectionInterface, error) {
	connectionName := "default"
	if len(name) > 0 {
		connectionName = name[0]
	}
	return defaultManager.Connection(connectionName)
}

// Query 创建查询构造器
func Query(connectionName ...string) (QueryInterface, error) {
	return defaultManager.Query(connectionName...)
}

// Table 指定表名创建查询构造器
func Table(tableName string, connectionName ...string) (QueryInterface, error) {
	return defaultManager.Table(tableName, connectionName...)
}

// Raw 执行原生SQL查询
func Raw(ctx context.Context, sql string, bindings []interface{}, connectionName ...string) ([]map[string]interface{}, error) {
	return defaultManager.Raw(ctx, sql, bindings, connectionName...)
}

// Exec 执行SQL语句
func Exec(ctx context.Context, sql string, bindings []interface{}, connectionName ...string) (sql.Result, error) {
	return defaultManager.Exec(ctx, sql, bindings, connectionName...)
}

// Transaction 执行事务
func Transaction(ctx context.Context, fn func(tx TransactionInterface) error, connectionName ...string) error {
	return defaultManager.Transaction(ctx, fn, connectionName...)
}
