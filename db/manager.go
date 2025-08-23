package db

import (
	"fmt"
	"sync"
)

// Manager 简化的数据库管理器
type Manager struct {
	configs     map[string]*Config
	connections map[string]ConnectionInterface
	logger      LoggerInterface
	mutex       sync.RWMutex
}

// defaultManager 默认管理器实例
var defaultManager *Manager

func init() {
	defaultManager = NewManager()
}

// NewManager 创建新的管理器
func NewManager() *Manager {
	return &Manager{
		configs:     make(map[string]*Config),
		connections: make(map[string]ConnectionInterface),
		logger:      nil, // 默认无日志
	}
}

// NewManagerWithLogger 创建带日志的管理器
func NewManagerWithLogger(logger LoggerInterface) *Manager {
	return &Manager{
		configs:     make(map[string]*Config),
		connections: make(map[string]ConnectionInterface),
		logger:      logger,
	}
}

// SetLogger 设置日志记录器
func (m *Manager) SetLogger(logger LoggerInterface) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.logger = logger
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
		m.mutex.RUnlock()
		return conn, nil
	}

	// 获取配置
	config, exists := m.configs[name]
	m.mutex.RUnlock()
	if !exists {
		return nil, fmt.Errorf("连接配置 '%s' 不存在", name)
	}

	// 创建连接（无锁状态）
	var conn ConnectionInterface
	var err error

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

	// 缓存连接（写锁）
	m.mutex.Lock()
	m.connections[name] = conn
	m.mutex.Unlock()

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
