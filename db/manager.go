package db

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"
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

	if err := conn.Connect(); err != nil {
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

	rows, err := conn.Query(sql, bindings...)
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

	return conn.Exec(sql, bindings...)
}

// Transaction 执行事务
func (m *Manager) Transaction(fn func(tx TransactionInterface) error, connectionName ...string) error {
	name := "default"
	if len(connectionName) > 0 {
		name = connectionName[0]
	}

	conn, err := m.Connection(name)
	if err != nil {
		return err
	}

	tx, err := conn.Begin()
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
func Raw(sql string, bindings []interface{}, connectionName ...string) ([]map[string]interface{}, error) {
	return defaultManager.Raw(context.Background(), sql, bindings, connectionName...)
}

// Exec 执行SQL语句
func Exec(sql string, bindings []interface{}, connectionName ...string) (sql.Result, error) {
	return defaultManager.Exec(context.Background(), sql, bindings, connectionName...)
}

// Transaction 执行事务
func Transaction(fn func(tx TransactionInterface) error, connectionName ...string) error {
	return defaultManager.Transaction(fn, connectionName...)
}

// 便利的全局查询方法，提供TORM风格的API

// WhereGlobal 创建带有WHERE条件的新查询 - 便利函数
// 用法: db.WhereGlobal("name = ?", "张三").First(&user)
func WhereGlobal(args ...interface{}) (QueryInterface, error) {
	conn, err := DB()
	if err != nil {
		return nil, err
	}
	return NewQuery(conn).Where(args...), nil
}

// Model 基于模型创建查询 - 自动获取表名和模型特性
// 用法: db.Model(&User{}).Where("age > ?", 18).Find(&users)
func Model(model interface{}) (QueryInterface, error) {
	if model == nil {
		return nil, fmt.Errorf("model cannot be nil")
	}

	// 获取表名
	tableName := getTableNameFromModel(model)
	if tableName == "" {
		return nil, fmt.Errorf("cannot determine table name from model")
	}

	// 尝试从模型获取连接名
	connectionName := getConnectionFromModel(model)
	if connectionName == "" {
		connectionName = "default" // 默认连接
	}

	// 创建查询构建器并绑定模型
	query, err := Table(tableName, connectionName)
	if err != nil {
		return nil, err
	}

	return query.WithModel(model), nil
}

// FirstGlobal 直接查询第一条记录 - 便利函数
// 用法: db.FirstGlobal(&user) 或 db.FirstGlobal(&user, 10)
func FirstGlobal(dest interface{}, conds ...interface{}) error {
	conn, err := DB()
	if err != nil {
		return err
	}

	query := NewQuery(conn)

	// 如果有条件参数，添加Where条件
	if len(conds) > 0 {
		// 如果第一个参数是数字，视为按ID查询
		if len(conds) == 1 {
			query = query.Where("id", "=", conds[0])
		} else {
			// 否则作为条件查询
			query = query.Where(conds...)
		}
	}

	_, err = query.First(dest)
	return err
}

// FindGlobal 直接查询记录 - 便利函数
// 用法: db.FindGlobal(&users) 或 db.FindGlobal(&user, 10)
func FindGlobal(dest interface{}, conds ...interface{}) error {
	conn, err := DB()
	if err != nil {
		return err
	}

	query := NewQuery(conn)

	// 如果有条件参数，添加Where条件
	if len(conds) > 0 {
		// 如果第一个参数是数字，视为按ID查询
		if len(conds) == 1 {
			query = query.Where("id", "=", conds[0])
		} else {
			// 否则作为条件查询
			query = query.Where(conds...)
		}
	}

	_, err = query.Find(dest)
	return err
}

// getTableNameFromModel 从模型中获取表名
func getTableNameFromModel(model interface{}) string {
	if model == nil {
		return ""
	}

	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 1. 尝试调用TableName方法
	modelValue := reflect.ValueOf(model)
	if modelValue.Kind() == reflect.Ptr {
		modelValue = modelValue.Elem()
	}

	// 检查是否有BaseModel字段
	if baseModelField := modelValue.FieldByName("BaseModel"); baseModelField.IsValid() {
		// 尝试调用TableName方法
		tableNameMethod := baseModelField.Addr().MethodByName("TableName")
		if tableNameMethod.IsValid() {
			result := tableNameMethod.Call(nil)
			if len(result) > 0 {
				if tableName, ok := result[0].Interface().(string); ok && tableName != "" {
					return tableName
				}
			}
		}
	}

	// 2. 根据结构体名称推断表名
	modelName := modelType.Name()
	if modelName != "" {
		// 简单的复数形式转换
		tableName := strings.ToLower(modelName)
		if !strings.HasSuffix(tableName, "s") {
			tableName += "s"
		}
		return tableName
	}

	return ""
}

// getConnectionFromModel 从模型中获取连接名
func getConnectionFromModel(model interface{}) string {
	if model == nil {
		return ""
	}

	modelValue := reflect.ValueOf(model)
	if modelValue.Kind() == reflect.Ptr {
		modelValue = modelValue.Elem()
	}

	// 检查是否有BaseModel字段
	if baseModelField := modelValue.FieldByName("BaseModel"); baseModelField.IsValid() {
		// 尝试调用GetConnection方法
		getConnMethod := baseModelField.Addr().MethodByName("GetConnection")
		if getConnMethod.IsValid() {
			result := getConnMethod.Call(nil)
			if len(result) > 0 {
				if connName, ok := result[0].Interface().(string); ok && connName != "" {
					return connName
				}
			}
		}
	}

	return "" // 返回空字符串，使用默认连接
}
