package db

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/zhoudm1743/torm/logger"
)

// DbManager 数据库管理器
// 负责管理多个数据库连接的配置、创建和生命周期
type DbManager struct {
	configs     map[string]*Config    // 配置映射
	connections map[string]Connection // 连接实例映射
	cache       CacheInterface        // 缓存实例
	logger      LoggerInterface       // 日志实例
	mu          sync.RWMutex          // 读写锁

	// 健康检查配置
	healthCheckInterval time.Duration
	healthCheckEnabled  bool
	stopHealthCheck     chan struct{}

	// 连接池配置
	maxConnections    int           // 最大连接数
	connectionTimeout time.Duration // 连接超时时间
	idleTimeout       time.Duration // 空闲超时时间
	cleanupInterval   time.Duration // 清理间隔
	stopCleanup       chan struct{} // 停止清理信号
}

// ConnectionStats 连接统计信息
type ConnectionStats struct {
	Name         string    `json:"name"`
	Driver       string    `json:"driver"`
	IsHealthy    bool      `json:"is_healthy"`
	LastCheck    time.Time `json:"last_check"`
	CreatedAt    time.Time `json:"created_at"`
	LastUsed     time.Time `json:"last_used"`
	ErrorCount   int       `json:"error_count"`
	TotalQueries int64     `json:"total_queries"`
}

// connectionMeta 连接元数据
type connectionMeta struct {
	stats     *ConnectionStats
	createdAt time.Time
	lastUsed  time.Time
}

// 连接元数据映射
var connectionMetas = make(map[string]*connectionMeta)
var metaMu sync.RWMutex

// defaultManager 默认管理器实例
var defaultManager *DbManager
var defaultManagerOnce sync.Once

// NewDbManager 创建新的数据库管理器
func NewDbManager() *DbManager {
	m := &DbManager{
		configs:             make(map[string]*Config),
		connections:         make(map[string]Connection),
		logger:              logger.NewInfoLogger(),
		healthCheckInterval: 30 * time.Second,
		healthCheckEnabled:  false,
		stopHealthCheck:     make(chan struct{}),
		maxConnections:      50,
		connectionTimeout:   30 * time.Second,
		idleTimeout:         10 * time.Minute,
		cleanupInterval:     5 * time.Minute,
		stopCleanup:         make(chan struct{}),
	}

	// 启动连接池清理协程
	go m.startConnectionCleanup()

	return m
}

// NewDbManagerWithLogger 创建带日志的管理器
func NewDbManagerWithLogger(log LoggerInterface) *DbManager {
	m := NewDbManager()
	m.logger = log
	return m
}

// DefaultManager 获取默认管理器
func DefaultManager() *DbManager {
	defaultManagerOnce.Do(func() {
		defaultManager = NewDbManager()
	})
	return defaultManager
}

// SetConfig 设置数据库配置
func (m *DbManager) SetConfig(name string, config *Config) error {
	if config == nil {
		return NewError(ErrCodeConfigurationError, "config cannot be nil")
	}

	if err := config.Validate(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.configs[name] = config
	return nil
}

// GetConfig 获取数据库配置
func (m *DbManager) GetConfig(name string) (*Config, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	config, exists := m.configs[name]
	if !exists {
		return nil, NewErrorf(ErrCodeConfigurationError, "config '%s' not found", name)
	}

	return config, nil
}

// Connection 获取数据库连接
func (m *DbManager) Connection(name ...string) (Connection, error) {
	connName := "default"
	if len(name) > 0 && name[0] != "" {
		connName = name[0]
	}

	// 检查连接数量限制
	m.mu.RLock()
	connectionCount := len(m.connections)
	m.mu.RUnlock()

	if connectionCount >= m.maxConnections {
		m.cleanupIdleConnections()

		m.mu.RLock()
		connectionCount = len(m.connections)
		m.mu.RUnlock()

		if connectionCount >= m.maxConnections {
			return nil, NewErrorf(ErrCodeConnectionPoolExhausted,
				"connection pool exhausted, current: %d, max: %d", connectionCount, m.maxConnections)
		}
	}

	// 检查是否已有连接
	m.mu.RLock()
	conn, exists := m.connections[connName]
	m.mu.RUnlock()

	if exists {
		// 更新使用时间
		m.updateConnectionMeta(connName)

		// 检查连接健康状态
		if err := conn.Ping(); err != nil {
			if m.logger != nil {
				m.logger.Warn("连接不健康，重新创建", "connection", connName, "error", err)
			}
			m.removeConnection(connName)
			return m.createConnection(connName)
		}

		return conn, nil
	}

	return m.createConnection(connName)
}

// createConnection 创建新连接
func (m *DbManager) createConnection(name string) (Connection, error) {
	m.mu.RLock()
	config, exists := m.configs[name]
	m.mu.RUnlock()

	if !exists {
		return nil, NewErrorf(ErrCodeConfigurationError, "config '%s' not found", name)
	}

	start := time.Now()
	var conn Connection
	var err error

	// 检查是否配置了读写分离
	// Requirements: 1.7
	if len(config.ReadHosts) > 0 || len(config.WriteHosts) > 0 {
		conn, err = NewReadWriteSplitConnection(config, m.logger)
	} else {
		switch config.GetDriverName() {
		case "mysql":
			conn, err = NewMySQLConnection(config, m.logger)
		case "postgres":
			conn, err = NewPostgreSQLConnection(config, m.logger)
		case "sqlite":
			conn, err = NewSQLiteConnection(config, m.logger)
		default:
			return nil, NewErrorf(ErrCodeDriverNotSupported, "unsupported driver: %s", config.Driver)
		}
	}

	if err != nil {
		return nil, err
	}

	// 连接数据库
	if err := conn.Connect(); err != nil {
		return nil, err
	}

	duration := time.Since(start)

	// 缓存连接
	m.mu.Lock()
	m.connections[name] = conn
	m.mu.Unlock()

	// 记录元数据
	m.setConnectionMeta(name, config.GetDriverName())

	if m.logger != nil {
		m.logger.Info("数据库连接创建成功",
			"connection", name,
			"driver", config.GetDriverName(),
			"read_write_split", len(config.ReadHosts) > 0 || len(config.WriteHosts) > 0,
			"duration", duration)
	}

	return conn, nil
}

// CloseAll 关闭所有连接
func (m *DbManager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 停止健康检查
	if m.healthCheckEnabled {
		m.healthCheckEnabled = false
		close(m.stopHealthCheck)
		m.stopHealthCheck = make(chan struct{})
	}

	var errors []string
	for name, conn := range m.connections {
		if err := conn.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("close '%s' failed: %v", name, err))
		}
	}

	// 清空连接
	m.connections = make(map[string]Connection)

	// 清空元数据
	metaMu.Lock()
	connectionMetas = make(map[string]*connectionMeta)
	metaMu.Unlock()

	if len(errors) > 0 {
		return NewErrorf(ErrCodeConnectionFailed, "errors closing connections: %v", errors)
	}

	if m.logger != nil {
		m.logger.Info("所有数据库连接已关闭")
	}

	return nil
}

// SetCache 设置缓存
func (m *DbManager) SetCache(cache CacheInterface) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cache = cache
}

// GetCache 获取缓存
func (m *DbManager) GetCache() CacheInterface {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cache
}

// SetLogger 设置日志记录器
func (m *DbManager) SetLogger(log LoggerInterface) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logger = log
}

// GetLogger 获取日志记录器
func (m *DbManager) GetLogger() LoggerInterface {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.logger
}

// Table 创建表查询构建器
func (m *DbManager) Table(name string, connection ...string) (*QueryBuilder, error) {
	connName := "default"
	if len(connection) > 0 {
		connName = connection[0]
	}

	builder, err := NewQueryBuilder(connName)
	if err != nil {
		return nil, err
	}

	builder.tableName = name
	return builder, nil
}

// EnableHealthCheck 启用健康检查
func (m *DbManager) EnableHealthCheck(interval time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.healthCheckEnabled {
		return
	}

	m.healthCheckInterval = interval
	m.healthCheckEnabled = true

	go m.healthCheckLoop()

	if m.logger != nil {
		m.logger.Info("连接健康检查已启用", "interval", interval)
	}
}

// DisableHealthCheck 禁用健康检查
func (m *DbManager) DisableHealthCheck() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.healthCheckEnabled {
		return
	}

	m.healthCheckEnabled = false
	close(m.stopHealthCheck)
	m.stopHealthCheck = make(chan struct{})

	if m.logger != nil {
		m.logger.Info("连接健康检查已禁用")
	}
}

// healthCheckLoop 健康检查循环
func (m *DbManager) healthCheckLoop() {
	ticker := time.NewTicker(m.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.performHealthCheck()
		case <-m.stopHealthCheck:
			return
		}
	}
}

// performHealthCheck 执行健康检查
func (m *DbManager) performHealthCheck() {
	m.mu.RLock()
	connections := make(map[string]Connection)
	for name, conn := range m.connections {
		connections[name] = conn
	}
	m.mu.RUnlock()

	for name, conn := range connections {
		go m.checkConnection(name, conn)
	}
}

// checkConnection 检查单个连接
func (m *DbManager) checkConnection(name string, conn Connection) {
	done := make(chan error, 1)
	go func() {
		done <- conn.Ping()
	}()

	select {
	case err := <-done:
		metaMu.Lock()
		if meta, exists := connectionMetas[name]; exists {
			meta.stats.LastCheck = time.Now()
			if err != nil {
				meta.stats.IsHealthy = false
				meta.stats.ErrorCount++
				if m.logger != nil {
					m.logger.Warn("连接健康检查失败", "connection", name, "error", err)
				}
			} else {
				meta.stats.IsHealthy = true
			}
		}
		metaMu.Unlock()
	case <-time.After(5 * time.Second):
		if m.logger != nil {
			m.logger.Warn("连接健康检查超时", "connection", name)
		}
	}
}

// removeConnection 移除连接
func (m *DbManager) removeConnection(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if conn, exists := m.connections[name]; exists {
		conn.Close()
		delete(m.connections, name)
	}

	metaMu.Lock()
	delete(connectionMetas, name)
	metaMu.Unlock()
}

// cleanupIdleConnections 清理空闲连接
func (m *DbManager) cleanupIdleConnections() {
	now := time.Now()
	var idleConnections []string

	metaMu.RLock()
	for name, meta := range connectionMetas {
		if now.Sub(meta.lastUsed) > m.idleTimeout {
			idleConnections = append(idleConnections, name)
		}
	}
	metaMu.RUnlock()

	for _, name := range idleConnections {
		m.removeConnection(name)
		if m.logger != nil {
			m.logger.Debug("清理空闲连接", "connection", name)
		}
	}
}

// startConnectionCleanup 启动连接池清理协程
func (m *DbManager) startConnectionCleanup() {
	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanupIdleConnections()
		case <-m.stopCleanup:
			return
		}
	}
}

// setConnectionMeta 设置连接元数据
func (m *DbManager) setConnectionMeta(name, driver string) {
	metaMu.Lock()
	defer metaMu.Unlock()

	now := time.Now()
	connectionMetas[name] = &connectionMeta{
		stats: &ConnectionStats{
			Name:         name,
			Driver:       driver,
			IsHealthy:    true,
			LastCheck:    now,
			CreatedAt:    now,
			LastUsed:     now,
			ErrorCount:   0,
			TotalQueries: 0,
		},
		createdAt: now,
		lastUsed:  now,
	}
}

// updateConnectionMeta 更新连接元数据
func (m *DbManager) updateConnectionMeta(name string) {
	metaMu.Lock()
	defer metaMu.Unlock()

	if meta, exists := connectionMetas[name]; exists {
		meta.lastUsed = time.Now()
		meta.stats.LastUsed = meta.lastUsed
		meta.stats.TotalQueries++
	}
}

// GetConnectionStats 获取所有连接统计信息
func (m *DbManager) GetConnectionStats() map[string]*ConnectionStats {
	metaMu.RLock()
	defer metaMu.RUnlock()

	result := make(map[string]*ConnectionStats)
	for name, meta := range connectionMetas {
		result[name] = &ConnectionStats{
			Name:         meta.stats.Name,
			Driver:       meta.stats.Driver,
			IsHealthy:    meta.stats.IsHealthy,
			LastCheck:    meta.stats.LastCheck,
			CreatedAt:    meta.stats.CreatedAt,
			LastUsed:     meta.stats.LastUsed,
			ErrorCount:   meta.stats.ErrorCount,
			TotalQueries: meta.stats.TotalQueries,
		}
	}
	return result
}

// GetHealthyConnections 获取健康的连接数量
func (m *DbManager) GetHealthyConnections() (healthy, total int) {
	metaMu.RLock()
	defer metaMu.RUnlock()

	total = len(connectionMetas)
	for _, meta := range connectionMetas {
		if meta.stats.IsHealthy {
			healthy++
		}
	}
	return
}

// SetPoolConfig 设置连接池配置
func (m *DbManager) SetPoolConfig(maxConnections int, connectionTimeout, idleTimeout, cleanupInterval time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.maxConnections = maxConnections
	m.connectionTimeout = connectionTimeout
	m.idleTimeout = idleTimeout
	m.cleanupInterval = cleanupInterval

	if m.logger != nil {
		m.logger.Info("连接池配置已更新",
			"maxConnections", maxConnections,
			"connectionTimeout", connectionTimeout,
			"idleTimeout", idleTimeout,
			"cleanupInterval", cleanupInterval)
	}
}

// WarmUpConnections 预热连接
func (m *DbManager) WarmUpConnections() error {
	m.mu.RLock()
	configs := make(map[string]*Config)
	for name, config := range m.configs {
		configs[name] = config
	}
	m.mu.RUnlock()

	for name := range configs {
		if _, err := m.Connection(name); err != nil {
			if m.logger != nil {
				m.logger.Error("连接预热失败", "connection", name, "error", err)
			}
			return err
		}
	}

	if m.logger != nil {
		m.logger.Info("连接预热完成", "count", len(configs))
	}

	return nil
}

// ========== 原生SQL查询方法 ==========

// Query 执行原生SELECT查询
// 支持参数绑定，返回查询结果
// Requirements: 11.1
func (m *DbManager) Query(sql string, args ...interface{}) (*RawQueryResult, error) {
	return m.QueryOnConnection("default", sql, args...)
}

// QueryOnConnection 在指定连接上执行原生SELECT查询
// Requirements: 11.1
func (m *DbManager) QueryOnConnection(connName string, sql string, args ...interface{}) (*RawQueryResult, error) {
	conn, err := m.Connection(connName)
	if err != nil {
		return nil, err
	}

	rows, err := conn.Query(sql, args...)
	if err != nil {
		return nil, WrapError(err, ErrCodeQueryFailed, "raw query failed").
			WithContext("sql", sql).
			WithContext("args", args)
	}

	return NewRawQueryResult(rows), nil
}

// Exec 执行原生INSERT/UPDATE/DELETE查询
// 支持参数绑定，返回执行结果
// Requirements: 11.2
func (m *DbManager) Exec(sql string, args ...interface{}) (*RawExecResult, error) {
	return m.ExecOnConnection("default", sql, args...)
}

// ExecOnConnection 在指定连接上执行原生INSERT/UPDATE/DELETE查询
// Requirements: 11.2
func (m *DbManager) ExecOnConnection(connName string, sql string, args ...interface{}) (*RawExecResult, error) {
	conn, err := m.Connection(connName)
	if err != nil {
		return nil, err
	}

	result, err := conn.Exec(sql, args...)
	if err != nil {
		return nil, WrapError(err, ErrCodeQueryFailed, "raw exec failed").
			WithContext("sql", sql).
			WithContext("args", args)
	}

	return NewRawExecResult(result), nil
}

// RawQueryResult 原生查询结果包装器
type RawQueryResult struct {
	rows *sql.Rows
}

// NewRawQueryResult 创建原生查询结果
func NewRawQueryResult(rows *sql.Rows) *RawQueryResult {
	return &RawQueryResult{rows: rows}
}

// Rows 获取底层sql.Rows
func (r *RawQueryResult) Rows() *sql.Rows {
	return r.rows
}

// Close 关闭结果集
func (r *RawQueryResult) Close() error {
	if r.rows != nil {
		return r.rows.Close()
	}
	return nil
}

// Scan 扫描到目标变量
func (r *RawQueryResult) Scan(dest ...interface{}) error {
	if r.rows == nil {
		return NewError(ErrCodeQueryFailed, "no rows to scan")
	}
	if !r.rows.Next() {
		return ErrRecordNotFound
	}
	return r.rows.Scan(dest...)
}

// ScanAll 扫描所有行到map切片
func (r *RawQueryResult) ScanAll() ([]map[string]interface{}, error) {
	if r.rows == nil {
		return nil, NewError(ErrCodeQueryFailed, "no rows to scan")
	}
	defer r.rows.Close()

	columns, err := r.rows.Columns()
	if err != nil {
		return nil, WrapError(err, ErrCodeQueryFailed, "failed to get columns")
	}

	var results []map[string]interface{}

	for r.rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := r.rows.Scan(valuePtrs...); err != nil {
			return nil, WrapError(err, ErrCodeQueryFailed, "failed to scan row")
		}

		row := make(map[string]interface{})
		for i, column := range columns {
			row[column] = convertValue(values[i])
		}

		results = append(results, row)
	}

	if err := r.rows.Err(); err != nil {
		return nil, WrapError(err, ErrCodeQueryFailed, "error iterating rows")
	}

	return results, nil
}

// ScanFirst 扫描第一行到map
func (r *RawQueryResult) ScanFirst() (map[string]interface{}, error) {
	if r.rows == nil {
		return nil, NewError(ErrCodeQueryFailed, "no rows to scan")
	}
	defer r.rows.Close()

	columns, err := r.rows.Columns()
	if err != nil {
		return nil, WrapError(err, ErrCodeQueryFailed, "failed to get columns")
	}

	if !r.rows.Next() {
		return nil, ErrRecordNotFound
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))

	for i := range values {
		valuePtrs[i] = &values[i]
	}

	if err := r.rows.Scan(valuePtrs...); err != nil {
		return nil, WrapError(err, ErrCodeQueryFailed, "failed to scan row")
	}

	row := make(map[string]interface{})
	for i, column := range columns {
		row[column] = convertValue(values[i])
	}

	return row, nil
}

// ScanColumn 扫描单列到切片
func (r *RawQueryResult) ScanColumn() ([]interface{}, error) {
	if r.rows == nil {
		return nil, NewError(ErrCodeQueryFailed, "no rows to scan")
	}
	defer r.rows.Close()

	var results []interface{}

	for r.rows.Next() {
		var value interface{}
		if err := r.rows.Scan(&value); err != nil {
			return nil, WrapError(err, ErrCodeQueryFailed, "failed to scan value")
		}
		results = append(results, convertValue(value))
	}

	if err := r.rows.Err(); err != nil {
		return nil, WrapError(err, ErrCodeQueryFailed, "error iterating rows")
	}

	return results, nil
}

// RawExecResult 原生执行结果包装器
type RawExecResult struct {
	result sql.Result
}

// NewRawExecResult 创建原生执行结果
func NewRawExecResult(result sql.Result) *RawExecResult {
	return &RawExecResult{result: result}
}

// Result 获取底层sql.Result
func (r *RawExecResult) Result() sql.Result {
	return r.result
}

// LastInsertId 获取最后插入的ID
func (r *RawExecResult) LastInsertId() (int64, error) {
	if r.result == nil {
		return 0, NewError(ErrCodeQueryFailed, "no result")
	}
	return r.result.LastInsertId()
}

// RowsAffected 获取受影响的行数
func (r *RawExecResult) RowsAffected() (int64, error) {
	if r.result == nil {
		return 0, NewError(ErrCodeQueryFailed, "no result")
	}
	return r.result.RowsAffected()
}

// convertValue 转换数据库值
func convertValue(value interface{}) interface{} {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return string(v)
	default:
		return v
	}
}

// ========== 便捷函数 ==========

// AddConnection 添加连接配置（便捷函数）
func AddConnection(name string, config *Config) error {
	return DefaultManager().SetConfig(name, config)
}

// DB 获取数据库连接（便捷函数）
func DB(name ...string) (Connection, error) {
	return DefaultManager().Connection(name...)
}

// Table 创建表查询构建器（便捷函数）
func Table(tableName string, connectionName ...string) (*QueryBuilder, error) {
	return DefaultManager().Table(tableName, connectionName...)
}

// CloseAllConnections 关闭所有连接（便捷函数）
func CloseAllConnections() error {
	return DefaultManager().CloseAll()
}

// GetConnectionStats 获取连接统计信息（便捷函数）
func GetConnectionStats() map[string]*ConnectionStats {
	return DefaultManager().GetConnectionStats()
}

// GetHealthyConnections 获取健康的连接数量（便捷函数）
func GetHealthyConnections() (healthy, total int) {
	return DefaultManager().GetHealthyConnections()
}

// WarmUpConnections 预热连接（便捷函数）
func WarmUpConnections() error {
	return DefaultManager().WarmUpConnections()
}

// SetConnectionPoolConfig 设置连接池配置（便捷函数）
func SetConnectionPoolConfig(maxConnections int, connectionTimeout, idleTimeout, cleanupInterval time.Duration) {
	DefaultManager().SetPoolConfig(maxConnections, connectionTimeout, idleTimeout, cleanupInterval)
}

// RawQuery 执行原生SELECT查询（便捷函数）
// Requirements: 11.1
func RawQuery(sql string, args ...interface{}) (*RawQueryResult, error) {
	return DefaultManager().Query(sql, args...)
}

// RawQueryOn 在指定连接上执行原生SELECT查询（便捷函数）
// Requirements: 11.1
func RawQueryOn(connName string, sql string, args ...interface{}) (*RawQueryResult, error) {
	return DefaultManager().QueryOnConnection(connName, sql, args...)
}

// RawExec 执行原生INSERT/UPDATE/DELETE查询（便捷函数）
// Requirements: 11.2
func RawExec(sql string, args ...interface{}) (*RawExecResult, error) {
	return DefaultManager().Exec(sql, args...)
}

// RawExecOn 在指定连接上执行原生INSERT/UPDATE/DELETE查询（便捷函数）
// Requirements: 11.2
func RawExecOn(connName string, sql string, args ...interface{}) (*RawExecResult, error) {
	return DefaultManager().ExecOnConnection(connName, sql, args...)
}

// ========== 读写分离相关方法 ==========

// GetReadWriteSplitConnection 获取读写分离连接
// 如果连接不是读写分离类型，返回nil
// Requirements: 1.7
func (m *DbManager) GetReadWriteSplitConnection(name ...string) (*ReadWriteSplitConnection, error) {
	conn, err := m.Connection(name...)
	if err != nil {
		return nil, err
	}

	rwConn, ok := conn.(*ReadWriteSplitConnection)
	if !ok {
		return nil, NewError(ErrCodeConfigurationError, "connection is not a read-write split connection")
	}

	return rwConn, nil
}

// UseMaster 强制使用主库执行查询
// 用于需要读取最新数据的场景
// Requirements: 1.9
func (m *DbManager) UseMaster(name ...string) (*MasterOnlyConnection, error) {
	rwConn, err := m.GetReadWriteSplitConnection(name...)
	if err != nil {
		return nil, err
	}

	return rwConn.UseMaster(), nil
}

// GetSlaveCount 获取从库数量
// Requirements: 1.8
func (m *DbManager) GetSlaveCount(name ...string) (int, error) {
	rwConn, err := m.GetReadWriteSplitConnection(name...)
	if err != nil {
		return 0, err
	}

	return rwConn.GetSlaveCount(), nil
}

// SetLoadBalancer 设置负载均衡策略
// Requirements: 1.8
func (m *DbManager) SetLoadBalancer(balancer LoadBalancer, name ...string) error {
	rwConn, err := m.GetReadWriteSplitConnection(name...)
	if err != nil {
		return err
	}

	rwConn.SetLoadBalancer(balancer)
	return nil
}

// SetLoadBalancerType 设置负载均衡类型
// Requirements: 1.8
func (m *DbManager) SetLoadBalancerType(lbType LoadBalancerType, name ...string) error {
	rwConn, err := m.GetReadWriteSplitConnection(name...)
	if err != nil {
		return err
	}

	rwConn.SetLoadBalancerType(lbType)
	return nil
}

// ========== 读写分离便捷函数 ==========

// UseMasterDB 强制使用主库执行查询（便捷函数）
// Requirements: 1.9
func UseMasterDB(name ...string) (*MasterOnlyConnection, error) {
	return DefaultManager().UseMaster(name...)
}

// GetSlaveCountDB 获取从库数量（便捷函数）
// Requirements: 1.8
func GetSlaveCountDB(name ...string) (int, error) {
	return DefaultManager().GetSlaveCount(name...)
}

// SetLoadBalancerDB 设置负载均衡策略（便捷函数）
// Requirements: 1.8
func SetLoadBalancerDB(balancer LoadBalancer, name ...string) error {
	return DefaultManager().SetLoadBalancer(balancer, name...)
}

// SetLoadBalancerTypeDB 设置负载均衡类型（便捷函数）
// Requirements: 1.8
func SetLoadBalancerTypeDB(lbType LoadBalancerType, name ...string) error {
	return DefaultManager().SetLoadBalancerType(lbType, name...)
}
