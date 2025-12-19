package db

import (
	"context"
	"database/sql"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// ReadWriteSplitConnection 读写分离连接
// 实现主从数据库的读写分离功能
// Requirements: 1.7, 1.8, 1.9
type ReadWriteSplitConnection struct {
	config       *Config
	masterConn   Connection   // 主库连接（写操作）
	slaveConns   []Connection // 从库连接列表（读操作）
	logger       LoggerInterface
	mu           sync.RWMutex
	connected    bool
	slaveIndex   uint64       // 用于轮询从库
	loadBalancer LoadBalancer // 负载均衡策略
}

// LoadBalancer 负载均衡接口
type LoadBalancer interface {
	// Select 选择一个从库连接
	Select(slaves []Connection) Connection
}

// LoadBalancerType 负载均衡类型
type LoadBalancerType int

const (
	// LoadBalancerRoundRobin 轮询
	LoadBalancerRoundRobin LoadBalancerType = iota
	// LoadBalancerRandom 随机
	LoadBalancerRandom
	// LoadBalancerWeighted 加权（暂未实现）
	LoadBalancerWeighted
)

// RoundRobinBalancer 轮询负载均衡器
type RoundRobinBalancer struct {
	counter uint64
}

// Select 轮询选择从库
func (b *RoundRobinBalancer) Select(slaves []Connection) Connection {
	if len(slaves) == 0 {
		return nil
	}
	index := atomic.AddUint64(&b.counter, 1) % uint64(len(slaves))
	return slaves[index]
}

// RandomBalancer 随机负载均衡器
type RandomBalancer struct {
	rng *rand.Rand
	mu  sync.Mutex
}

// NewRandomBalancer 创建随机负载均衡器
func NewRandomBalancer() *RandomBalancer {
	return &RandomBalancer{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Select 随机选择从库
func (b *RandomBalancer) Select(slaves []Connection) Connection {
	if len(slaves) == 0 {
		return nil
	}
	b.mu.Lock()
	index := b.rng.Intn(len(slaves))
	b.mu.Unlock()
	return slaves[index]
}

// NewReadWriteSplitConnection 创建读写分离连接
// Requirements: 1.7
func NewReadWriteSplitConnection(config *Config, log LoggerInterface) (*ReadWriteSplitConnection, error) {
	if config == nil {
		return nil, NewError(ErrCodeConfigurationError, "config cannot be nil")
	}

	// 验证读写分离配置
	if len(config.WriteHosts) == 0 && config.Host == "" {
		return nil, NewError(ErrCodeConfigurationError, "write host is required for read-write split")
	}

	conn := &ReadWriteSplitConnection{
		config:       config,
		logger:       log,
		slaveConns:   make([]Connection, 0),
		loadBalancer: &RoundRobinBalancer{}, // 默认使用轮询
	}

	return conn, nil
}

// SetLoadBalancer 设置负载均衡策略
func (c *ReadWriteSplitConnection) SetLoadBalancer(balancer LoadBalancer) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.loadBalancer = balancer
}

// SetLoadBalancerType 设置负载均衡类型
func (c *ReadWriteSplitConnection) SetLoadBalancerType(lbType LoadBalancerType) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch lbType {
	case LoadBalancerRoundRobin:
		c.loadBalancer = &RoundRobinBalancer{}
	case LoadBalancerRandom:
		c.loadBalancer = NewRandomBalancer()
	default:
		c.loadBalancer = &RoundRobinBalancer{}
	}
}

// Connect 连接到数据库（主库和从库）
// Requirements: 1.7
func (c *ReadWriteSplitConnection) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	// 创建主库连接
	masterConfig := c.config.Clone()
	if len(c.config.WriteHosts) > 0 {
		masterConfig.Host = c.config.WriteHosts[0]
	}

	masterConn, err := c.createConnection(masterConfig)
	if err != nil {
		return WrapError(err, ErrCodeConnectionFailed, "failed to connect to master").
			WithContext("host", masterConfig.Host)
	}

	if err := masterConn.Connect(); err != nil {
		return WrapError(err, ErrCodeConnectionFailed, "failed to connect to master").
			WithContext("host", masterConfig.Host)
	}

	c.masterConn = masterConn

	// 创建从库连接
	readHosts := c.config.ReadHosts
	if len(readHosts) == 0 {
		// 如果没有配置从库，使用主库作为从库
		c.slaveConns = []Connection{masterConn}
	} else {
		for _, host := range readHosts {
			slaveConfig := c.config.Clone()
			slaveConfig.Host = host

			slaveConn, err := c.createConnection(slaveConfig)
			if err != nil {
				if c.logger != nil {
					c.logger.Warn("创建从库连接失败，跳过",
						"host", host,
						"error", err)
				}
				continue
			}

			if err := slaveConn.Connect(); err != nil {
				if c.logger != nil {
					c.logger.Warn("连接从库失败，跳过",
						"host", host,
						"error", err)
				}
				continue
			}

			c.slaveConns = append(c.slaveConns, slaveConn)
		}

		// 如果所有从库都连接失败，使用主库
		if len(c.slaveConns) == 0 {
			if c.logger != nil {
				c.logger.Warn("所有从库连接失败，使用主库进行读操作")
			}
			c.slaveConns = []Connection{masterConn}
		}
	}

	c.connected = true

	if c.logger != nil {
		c.logger.Info("读写分离连接建立成功",
			"master", masterConfig.Host,
			"slaves", len(c.slaveConns))
	}

	return nil
}

// createConnection 根据驱动类型创建连接
func (c *ReadWriteSplitConnection) createConnection(config *Config) (Connection, error) {
	switch config.GetDriverName() {
	case "mysql":
		return NewMySQLConnection(config, c.logger)
	case "postgres":
		return NewPostgreSQLConnection(config, c.logger)
	case "sqlite":
		return NewSQLiteConnection(config, c.logger)
	default:
		return nil, NewErrorf(ErrCodeDriverNotSupported, "unsupported driver: %s", config.Driver)
	}
}

// Close 关闭所有连接
func (c *ReadWriteSplitConnection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errors []string

	// 关闭从库连接
	for i, slave := range c.slaveConns {
		// 跳过与主库相同的连接
		if slave == c.masterConn {
			continue
		}
		if err := slave.Close(); err != nil {
			errors = append(errors, err.Error())
		}
		c.slaveConns[i] = nil
	}
	c.slaveConns = nil

	// 关闭主库连接
	if c.masterConn != nil {
		if err := c.masterConn.Close(); err != nil {
			errors = append(errors, err.Error())
		}
		c.masterConn = nil
	}

	c.connected = false

	if len(errors) > 0 {
		return NewErrorf(ErrCodeConnectionFailed, "errors closing connections: %v", errors)
	}

	if c.logger != nil {
		c.logger.Info("读写分离连接已关闭")
	}

	return nil
}

// Ping 测试连接
func (c *ReadWriteSplitConnection) Ping() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.masterConn == nil {
		return NewError(ErrCodeConnectionClosed, "connection not established")
	}

	return c.masterConn.Ping()
}

// IsConnected 检查是否已连接
func (c *ReadWriteSplitConnection) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// getMasterConnection 获取主库连接（写操作）
// Requirements: 1.9
func (c *ReadWriteSplitConnection) getMasterConnection() Connection {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.masterConn
}

// getSlaveConnection 获取从库连接（读操作）
// Requirements: 1.8
func (c *ReadWriteSplitConnection) getSlaveConnection() Connection {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.slaveConns) == 0 {
		return c.masterConn
	}

	return c.loadBalancer.Select(c.slaveConns)
}

// Query 执行查询（路由到从库）
// Requirements: 1.8
func (c *ReadWriteSplitConnection) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return c.QueryContext(context.Background(), query, args...)
}

// QueryContext 执行带上下文的查询（路由到从库）
// Requirements: 1.8
func (c *ReadWriteSplitConnection) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	slave := c.getSlaveConnection()
	if slave == nil {
		return nil, NewError(ErrCodeConnectionClosed, "no slave connection available")
	}

	if c.logger != nil {
		c.logger.Debug("读操作路由到从库",
			"host", slave.GetConfig().Host)
	}

	return slave.QueryContext(ctx, query, args...)
}

// QueryRow 执行单行查询（路由到从库）
// Requirements: 1.8
func (c *ReadWriteSplitConnection) QueryRow(query string, args ...interface{}) *sql.Row {
	return c.QueryRowContext(context.Background(), query, args...)
}

// QueryRowContext 执行带上下文的单行查询（路由到从库）
// Requirements: 1.8
func (c *ReadWriteSplitConnection) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	slave := c.getSlaveConnection()
	if slave == nil {
		return nil
	}

	if c.logger != nil {
		c.logger.Debug("读操作路由到从库",
			"host", slave.GetConfig().Host)
	}

	return slave.QueryRowContext(ctx, query, args...)
}

// Exec 执行SQL语句（路由到主库）
// Requirements: 1.9
func (c *ReadWriteSplitConnection) Exec(query string, args ...interface{}) (sql.Result, error) {
	return c.ExecContext(context.Background(), query, args...)
}

// ExecContext 执行带上下文的SQL语句（路由到主库）
// Requirements: 1.9
func (c *ReadWriteSplitConnection) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	master := c.getMasterConnection()
	if master == nil {
		return nil, NewError(ErrCodeConnectionClosed, "no master connection available")
	}

	if c.logger != nil {
		c.logger.Debug("写操作路由到主库",
			"host", master.GetConfig().Host)
	}

	return master.ExecContext(ctx, query, args...)
}

// Begin 开始事务（在主库上）
// Requirements: 1.9
func (c *ReadWriteSplitConnection) Begin() (TxInterface, error) {
	return c.BeginTx(context.Background(), nil)
}

// BeginTx 开始事务（在主库上，带选项）
// Requirements: 1.9
func (c *ReadWriteSplitConnection) BeginTx(ctx context.Context, opts *sql.TxOptions) (TxInterface, error) {
	master := c.getMasterConnection()
	if master == nil {
		return nil, NewError(ErrCodeConnectionClosed, "no master connection available")
	}

	if c.logger != nil {
		c.logger.Debug("事务在主库上开始",
			"host", master.GetConfig().Host)
	}

	return master.BeginTx(ctx, opts)
}

// RunTransaction 执行事务（在主库上）
// Requirements: 1.9
func (c *ReadWriteSplitConnection) RunTransaction(fn func(tx TxInterface) error) error {
	master := c.getMasterConnection()
	if master == nil {
		return NewError(ErrCodeConnectionClosed, "no master connection available")
	}

	return master.RunTransaction(fn)
}

// GetConfig 获取配置
func (c *ReadWriteSplitConnection) GetConfig() *Config {
	return c.config
}

// GetDriver 获取驱动名称
func (c *ReadWriteSplitConnection) GetDriver() string {
	return c.config.GetDriverName()
}

// GetDB 获取底层数据库连接（返回主库）
func (c *ReadWriteSplitConnection) GetDB() *sql.DB {
	master := c.getMasterConnection()
	if master == nil {
		return nil
	}
	return master.GetDB()
}

// GetBuilder 获取SQL构建器
func (c *ReadWriteSplitConnection) GetBuilder() Builder {
	master := c.getMasterConnection()
	if master == nil {
		return nil
	}
	return master.GetBuilder()
}

// GetTableFields 获取表字段信息（从主库获取）
func (c *ReadWriteSplitConnection) GetTableFields(table string) ([]ConnFieldInfo, error) {
	master := c.getMasterConnection()
	if master == nil {
		return nil, NewError(ErrCodeConnectionClosed, "no master connection available")
	}
	return master.GetTableFields(table)
}

// GetTables 获取所有表名（从主库获取）
func (c *ReadWriteSplitConnection) GetTables() ([]string, error) {
	master := c.getMasterConnection()
	if master == nil {
		return nil, NewError(ErrCodeConnectionClosed, "no master connection available")
	}
	return master.GetTables()
}

// GetStats 获取连接统计
func (c *ReadWriteSplitConnection) GetStats() sql.DBStats {
	master := c.getMasterConnection()
	if master == nil {
		return sql.DBStats{}
	}
	return master.GetStats()
}

// SetLogger 设置日志记录器
func (c *ReadWriteSplitConnection) SetLogger(logger LoggerInterface) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger = logger
}

// GetLogger 获取日志记录器
func (c *ReadWriteSplitConnection) GetLogger() LoggerInterface {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.logger
}

// GetMasterConnection 获取主库连接（公开方法）
func (c *ReadWriteSplitConnection) GetMasterConnection() Connection {
	return c.getMasterConnection()
}

// GetSlaveConnections 获取所有从库连接
func (c *ReadWriteSplitConnection) GetSlaveConnections() []Connection {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]Connection, len(c.slaveConns))
	copy(result, c.slaveConns)
	return result
}

// GetSlaveCount 获取从库数量
func (c *ReadWriteSplitConnection) GetSlaveCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.slaveConns)
}

// UseMaster 强制使用主库执行查询
// 用于需要读取最新数据的场景
func (c *ReadWriteSplitConnection) UseMaster() *MasterOnlyConnection {
	return &MasterOnlyConnection{
		parent: c,
	}
}

// MasterOnlyConnection 仅使用主库的连接包装器
type MasterOnlyConnection struct {
	parent *ReadWriteSplitConnection
}

// Query 在主库上执行查询
func (m *MasterOnlyConnection) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return m.QueryContext(context.Background(), query, args...)
}

// QueryContext 在主库上执行带上下文的查询
func (m *MasterOnlyConnection) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	master := m.parent.getMasterConnection()
	if master == nil {
		return nil, NewError(ErrCodeConnectionClosed, "no master connection available")
	}
	return master.QueryContext(ctx, query, args...)
}

// QueryRow 在主库上执行单行查询
func (m *MasterOnlyConnection) QueryRow(query string, args ...interface{}) *sql.Row {
	return m.QueryRowContext(context.Background(), query, args...)
}

// QueryRowContext 在主库上执行带上下文的单行查询
func (m *MasterOnlyConnection) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	master := m.parent.getMasterConnection()
	if master == nil {
		return nil
	}
	return master.QueryRowContext(ctx, query, args...)
}
