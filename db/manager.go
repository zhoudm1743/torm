package db

import (
	"fmt"
	"sync"
	"time"

	"github.com/zhoudm1743/torm/logger"
)

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

// Manager 数据库管理器
type Manager struct {
	configs         map[string]*Config
	connections     map[string]ConnectionInterface
	connectionStats map[string]*ConnectionStats
	logger          LoggerInterface
	mutex           sync.RWMutex

	// 健康检查配置
	healthCheckInterval time.Duration
	healthCheckEnabled  bool
	stopHealthCheck     chan bool
}

// defaultManager 默认管理器实例
var defaultManager = NewManager()

// NewManager 创建新的管理器（无日志）
func NewManager() *Manager {
	m := &Manager{
		configs:             make(map[string]*Config),
		connections:         make(map[string]ConnectionInterface),
		connectionStats:     make(map[string]*ConnectionStats),
		logger:              nil, // 无日志记录器
		healthCheckInterval: 30 * time.Second,
		healthCheckEnabled:  false,
		stopHealthCheck:     make(chan bool, 1),
	}
	m.SetLogger(logger.NewSQLLogger(logger.INFO, true))
	return m
}

// NewManagerWithLogger 创建带日志的管理器
func NewManagerWithLogger(logger LoggerInterface) *Manager {
	m := NewManager()
	m.logger = logger
	return m
}

// NewManagerWithDefaultLogger 创建带默认日志的管理器
func NewManagerWithDefaultLogger() *Manager {
	defaultLogger := logger.NewSQLLogger(logger.INFO, true)
	return NewManagerWithLogger(defaultLogger)
}

// SetLogger 设置日志记录器
func (m *Manager) SetLogger(logger LoggerInterface) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.logger = logger
}

// EnableHealthCheck 启用健康检查
func (m *Manager) EnableHealthCheck(interval time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.healthCheckEnabled {
		return // 已经启用
	}

	m.healthCheckInterval = interval
	m.healthCheckEnabled = true

	// 启动健康检查协程
	go m.healthCheckLoop()

	if m.logger != nil {
		m.logger.Info("连接健康检查已启用", "interval", interval)
	}
}

// DisableHealthCheck 禁用健康检查
func (m *Manager) DisableHealthCheck() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.healthCheckEnabled {
		return
	}

	m.healthCheckEnabled = false
	select {
	case m.stopHealthCheck <- true:
	default:
	}

	if m.logger != nil {
		m.logger.Info("连接健康检查已禁用")
	}
}

// healthCheckLoop 健康检查循环
func (m *Manager) healthCheckLoop() {
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
func (m *Manager) performHealthCheck() {
	m.mutex.RLock()
	connections := make(map[string]ConnectionInterface)
	for name, conn := range m.connections {
		connections[name] = conn
	}
	m.mutex.RUnlock()

	for name, conn := range connections {
		go m.checkConnection(name, conn)
	}
}

// checkConnection 检查单个连接的健康状态
func (m *Manager) checkConnection(name string, conn ConnectionInterface) {
	start := time.Now()
	isHealthy := true

	// 添加超时控制，避免长时间阻塞
	done := make(chan error, 1)
	go func() {
		done <- conn.Ping()
	}()

	select {
	case err := <-done:
		if err != nil {
			isHealthy = false
			m.handleUnhealthyConnection(name, conn, err)
		}
	case <-time.After(5 * time.Second): // 5秒超时
		isHealthy = false
		if m.logger != nil {
			m.logger.Warn("连接健康检查超时", "connection", name, "timeout", "5s")
		}
	}

	// 更新统计信息
	m.mutex.Lock()
	if stats, exists := m.connectionStats[name]; exists {
		stats.IsHealthy = isHealthy
		stats.LastCheck = start
		if !isHealthy {
			stats.ErrorCount++
		}
	}
	m.mutex.Unlock()

}

// handleUnhealthyConnection 处理不健康的连接
func (m *Manager) handleUnhealthyConnection(name string, conn ConnectionInterface, err error) {
	if m.logger != nil {
		m.logger.Error("连接健康检查失败", "connection", name, "error", err)
	}

	// 尝试重新连接
	if reconnErr := conn.Connect(); reconnErr != nil {
		if m.logger != nil {
			m.logger.Error("连接重连失败", "connection", name, "error", reconnErr)
		}

		// 如果重连失败，移除这个连接
		m.mutex.Lock()
		delete(m.connections, name)
		delete(m.connectionStats, name)
		m.mutex.Unlock()
	} else {
		if m.logger != nil {
			m.logger.Info("连接重连成功", "connection", name)
		}
	}
}

// AddConfig 添加数据库配置
func (m *Manager) AddConfig(name string, config *Config) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.configs[name] = config
	return nil
}

// Connection 获取数据库连接
func (m *Manager) Connection(name string) (ConnectionInterface, error) {
	// 先检查是否已有连接（读锁）
	m.mutex.RLock()
	if conn, exists := m.connections[name]; exists {
		// 更新使用时间
		if stats, exists := m.connectionStats[name]; exists {
			stats.LastUsed = time.Now()
			stats.TotalQueries++
		}
		m.mutex.RUnlock()

		// 检查统计信息中的健康状态，避免频繁Ping
		m.mutex.RLock()
		stats, hasStats := m.connectionStats[name]
		m.mutex.RUnlock()

		// 如果最近检查过且状态不健康，或者从未检查过，则进行快速ping
		shouldPing := !hasStats || !stats.IsHealthy || time.Since(stats.LastCheck) > 30*time.Second

		if shouldPing {
			// 快速健康检查（带超时）
			done := make(chan error, 1)
			go func() {
				done <- conn.Ping()
			}()

			select {
			case err := <-done:
				if err != nil {
					if m.logger != nil {
						m.logger.Warn("连接可能已断开，尝试重新创建", "connection", name, "error", err)
					}
					// 连接不健康，移除并重新创建
					m.mutex.Lock()
					delete(m.connections, name)
					delete(m.connectionStats, name)
					m.mutex.Unlock()
					return m.createNewConnection(name)
				}
			case <-time.After(2 * time.Second): // 2秒超时
				if m.logger != nil {
					m.logger.Warn("连接ping超时，重新创建连接", "connection", name)
				}
				// Ping超时，移除并重新创建
				m.mutex.Lock()
				delete(m.connections, name)
				delete(m.connectionStats, name)
				m.mutex.Unlock()
				return m.createNewConnection(name)
			}
		}

		return conn, nil
	}

	// 获取配置
	_, exists := m.configs[name]
	m.mutex.RUnlock()
	if !exists {
		return nil, fmt.Errorf("连接配置 '%s' 不存在", name)
	}

	return m.createNewConnection(name)
}

// createNewConnection 创建新连接
func (m *Manager) createNewConnection(name string) (ConnectionInterface, error) {
	m.mutex.RLock()
	config, exists := m.configs[name]
	m.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("连接配置 '%s' 不存在", name)
	}

	// 创建连接（无锁状态）
	var conn ConnectionInterface
	var err error

	start := time.Now()

	switch config.Driver {
	case "mysql":
		conn, err = NewMySQLConnection(config, m.logger)
	case "sqlite", "sqlite3":
		conn, err = NewSQLiteConnection(config, m.logger)
	case "postgres", "postgresql":
		conn, err = NewPostgreSQLConnection(config, m.logger)
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %s", config.Driver)
	}

	if err != nil {
		return nil, fmt.Errorf("创建数据库连接失败: %w", err)
	}

	// 连接数据库
	if err := conn.Connect(); err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	duration := time.Since(start)

	// 缓存连接和统计信息（写锁）
	m.mutex.Lock()
	m.connections[name] = conn
	m.connectionStats[name] = &ConnectionStats{
		Name:         name,
		Driver:       config.Driver,
		IsHealthy:    true,
		LastCheck:    time.Now(),
		CreatedAt:    time.Now(),
		LastUsed:     time.Now(),
		ErrorCount:   0,
		TotalQueries: 1,
	}
	m.mutex.Unlock()

	if m.logger != nil {
		m.logger.Info("数据库连接创建成功", "connection", name, "driver", config.Driver, "duration", duration)
	}

	return conn, nil
}

// DefaultManager 获取默认管理器
func DefaultManager() *Manager {
	return defaultManager
}

// AddConnection 添加连接配置（便捷函数）
func AddConnection(name string, config *Config) error {
	return defaultManager.AddConfig(name, config)
}

// DB 获取数据库连接（便捷函数）
func DB(name ...string) (ConnectionInterface, error) {
	connectionName := "default"
	if len(name) > 0 {
		connectionName = name[0]
	}
	return defaultManager.Connection(connectionName)
}

// Table 创建表查询构建器（便捷函数）
func Table(tableName string, connectionName ...string) (*QueryBuilder, error) {
	connName := "default"
	if len(connectionName) > 0 {
		connName = connectionName[0]
	}

	builder, err := NewQueryBuilder(connName)
	if err != nil {
		return nil, err
	}

	builder.tableName = tableName
	return builder, nil
}

// Model 从模型创建查询构建器（便捷函数）
func Model(model interface{}, connectionName ...string) (*QueryBuilder, error) {
	connName := "default"
	if len(connectionName) > 0 {
		connName = connectionName[0]
	}

	builder, err := NewQueryBuilder(connName)
	if err != nil {
		return nil, err
	}

	// 设置模型
	builder.model = model

	// 从模型获取表名
	tableName := getTableNameFromModel(model)
	if tableName == "" {
		return nil, fmt.Errorf("无法从模型获取表名")
	}

	builder.tableName = tableName
	return builder, nil
}

// ClearCacheByTags 根据标签清理缓存
func ClearCacheByTags(tags ...string) error {
	if memCache, ok := GetDefaultCache().(*MemoryCache); ok {
		return memCache.DeleteByTags(tags)
	}
	return nil
}

// ClearAllCache 清理所有缓存
func ClearAllCache() error {
	return GetDefaultCache().Clear()
}

// GetCacheStats 获取缓存统计信息
func GetCacheStats() map[string]interface{} {
	if memCache, ok := GetDefaultCache().(*MemoryCache); ok {
		return memCache.Stats()
	}
	return nil
}

// GetConnectionStats 获取所有连接统计信息
func (m *Manager) GetConnectionStats() map[string]*ConnectionStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string]*ConnectionStats)
	for name, stats := range m.connectionStats {
		// 创建副本避免并发问题
		result[name] = &ConnectionStats{
			Name:         stats.Name,
			Driver:       stats.Driver,
			IsHealthy:    stats.IsHealthy,
			LastCheck:    stats.LastCheck,
			CreatedAt:    stats.CreatedAt,
			LastUsed:     stats.LastUsed,
			ErrorCount:   stats.ErrorCount,
			TotalQueries: stats.TotalQueries,
		}
	}
	return result
}

// GetConnectionStats 获取连接统计信息（便捷函数）
func GetConnectionStats() map[string]*ConnectionStats {
	return defaultManager.GetConnectionStats()
}

// WarmUpConnections 预热连接
func (m *Manager) WarmUpConnections() error {
	m.mutex.RLock()
	configs := make(map[string]*Config)
	for name, config := range m.configs {
		configs[name] = config
	}
	m.mutex.RUnlock()

	for name := range configs {
		if _, err := m.Connection(name); err != nil {
			if m.logger != nil {
				m.logger.Error("连接预热失败", "connection", name, "error", err)
			}
			return fmt.Errorf("预热连接 '%s' 失败: %w", name, err)
		}
	}

	if m.logger != nil {
		m.logger.Info("连接预热完成", "count", len(configs))
	}

	return nil
}

// WarmUpConnections 预热连接（便捷函数）
func WarmUpConnections() error {
	return defaultManager.WarmUpConnections()
}

// CloseAllConnections 关闭所有连接
func (m *Manager) CloseAllConnections() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 停止健康检查
	if m.healthCheckEnabled {
		m.healthCheckEnabled = false
		select {
		case m.stopHealthCheck <- true:
		default:
		}
	}

	var errors []string
	for name, conn := range m.connections {
		if err := conn.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("关闭连接 '%s' 失败: %v", name, err))
		}
	}

	// 清空连接和统计
	m.connections = make(map[string]ConnectionInterface)
	m.connectionStats = make(map[string]*ConnectionStats)

	if len(errors) > 0 {
		return fmt.Errorf("关闭连接时发生错误: %v", errors)
	}

	if m.logger != nil {
		m.logger.Info("所有数据库连接已关闭")
	}

	return nil
}

// CloseAllConnections 关闭所有连接（便捷函数）
func CloseAllConnections() error {
	return defaultManager.CloseAllConnections()
}

// GetHealthyConnections 获取健康的连接数量
func (m *Manager) GetHealthyConnections() (healthy, total int) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	total = len(m.connectionStats)
	for _, stats := range m.connectionStats {
		if stats.IsHealthy {
			healthy++
		}
	}
	return
}

// GetHealthyConnections 获取健康的连接数量（便捷函数）
func GetHealthyConnections() (healthy, total int) {
	return defaultManager.GetHealthyConnections()
}
