package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoConnection MongoDB连接实现
type MongoConnection struct {
	client     *mongo.Client
	database   *mongo.Database
	config     *Config
	logger     LoggerInterface
	connected  bool
}

// NewMongoConnection 创建MongoDB连接
func NewMongoConnection(config *Config) (*MongoConnection, error) {
	conn := &MongoConnection{
		config: config,
	}
	
	return conn, nil
}

// Connect 连接到MongoDB
func (m *MongoConnection) Connect() error {
	if m.connected {
		return nil
	}

	// 构建连接URI
	uri := m.buildConnectionURI()
	
	// 设置客户端选项
	clientOptions := options.Client().ApplyURI(uri)
	
	// 设置连接超时
	connectTimeout := 30 * time.Second
	if m.config.ConnMaxLifetime > 0 {
		connectTimeout = m.config.ConnMaxLifetime
	}
	clientOptions.SetConnectTimeout(connectTimeout)
	
	// 设置最大连接池大小
	if m.config.MaxOpenConns > 0 {
		clientOptions.SetMaxPoolSize(uint64(m.config.MaxOpenConns))
	}
	
	// 创建客户端
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return WrapError(err, ErrCodeConnectionFailed, "MongoDB连接失败").
			WithContext("uri", uri).
			WithDetails(fmt.Sprintf("连接错误: %v", err))
	}
	
	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	err = client.Ping(ctx, nil)
	if err != nil {
		return WrapError(err, ErrCodeConnectionFailed, "MongoDB连接测试失败").
			WithContext("uri", uri).
			WithDetails(fmt.Sprintf("Ping错误: %v", err))
	}
	
	m.client = client
	m.database = client.Database(m.config.Database)
	m.connected = true
	
	if m.logger != nil {
		// MongoDB连接成功，记录日志
		// 注意：LoggerInterface可能没有LogQuery方法，这里简化处理
	}
	
	return nil
}

// Close 关闭连接
func (m *MongoConnection) Close() error {
	if !m.connected || m.client == nil {
		return nil
	}
	
	err := m.client.Disconnect(context.TODO())
	if err != nil {
		return WrapError(err, ErrCodeConnectionFailed, "MongoDB断开连接失败")
	}
	
	m.connected = false
	m.client = nil
	m.database = nil
	
	return nil
}

// Ping 测试连接
func (m *MongoConnection) Ping() error {
	if !m.connected || m.client == nil {
		return NewError(ErrCodeConnectionClosed, "MongoDB连接未建立")
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err := m.client.Ping(ctx, nil)
	if err != nil {
		return WrapError(err, ErrCodeConnectionTimeout, "MongoDB连接测试失败")
	}
	
	return nil
}

// IsConnected 检查连接状态
func (m *MongoConnection) IsConnected() bool {
	return m.connected && m.client != nil
}

// Query MongoDB查询（适配SQL接口）
func (m *MongoConnection) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return nil, NewError(ErrCodeNotImplemented, "MongoDB不支持SQL查询，请使用MongoDB专用方法")
}

// QueryRow MongoDB单行查询（适配SQL接口）
func (m *MongoConnection) QueryRow(query string, args ...interface{}) *sql.Row {
	// MongoDB不支持SQL查询，返回空的Row
	return &sql.Row{}
}

// Exec MongoDB执行（适配SQL接口）
func (m *MongoConnection) Exec(query string, args ...interface{}) (sql.Result, error) {
	return nil, NewError(ErrCodeNotImplemented, "MongoDB不支持SQL执行，请使用MongoDB专用方法")
}

// Begin 开始事务（适配SQL接口）
func (m *MongoConnection) Begin() (TransactionInterface, error) {
	return m.BeginTx(nil)
}

// BeginTx 开始事务
func (m *MongoConnection) BeginTx(opts *sql.TxOptions) (TransactionInterface, error) {
	if !m.connected {
		return nil, NewError(ErrCodeConnectionClosed, "MongoDB连接未建立")
	}
	
	// MongoDB事务需要副本集支持
	session, err := m.client.StartSession()
	if err != nil {
		return nil, WrapError(err, ErrCodeTransactionFailed, "MongoDB事务启动失败")
	}
	
	err = session.StartTransaction()
	if err != nil {
		session.EndSession(context.TODO())
		return nil, WrapError(err, ErrCodeTransactionFailed, "MongoDB事务开始失败")
	}
	
	return &MongoTransaction{
		session:    session,
		connection: m,
	}, nil
}

// GetConfig 获取配置
func (m *MongoConnection) GetConfig() *Config {
	return m.config
}

// GetDriver 获取驱动名称
func (m *MongoConnection) GetDriver() string {
	return "mongodb"
}

// GetStats 获取连接统计（MongoDB不支持）
func (m *MongoConnection) GetStats() sql.DBStats {
	return sql.DBStats{}
}

// GetDB 获取原始数据库对象（适配SQL接口）
func (m *MongoConnection) GetDB() *sql.DB {
	return nil // MongoDB不使用sql.DB
}

// GetClient 获取MongoDB客户端
func (m *MongoConnection) GetClient() *mongo.Client {
	return m.client
}

// GetDatabase 获取MongoDB数据库
func (m *MongoConnection) GetDatabase() *mongo.Database {
	return m.database
}

// Collection 获取集合
func (m *MongoConnection) Collection(name string) *mongo.Collection {
	if m.database == nil {
		return nil
	}
	return m.database.Collection(name)
}

// SetLogger 设置日志记录器
func (m *MongoConnection) SetLogger(logger LoggerInterface) {
	m.logger = logger
}

// GetLogger 获取日志记录器
func (m *MongoConnection) GetLogger() LoggerInterface {
	return m.logger
}

// buildConnectionURI 构建连接URI
func (m *MongoConnection) buildConnectionURI() string {
	config := m.config
	
	if config.DSN() != "" {
		return config.DSN()
	}
	
	// 构建MongoDB URI
	uri := "mongodb://"
	
	// 添加认证信息
	if config.Username != "" {
		uri += config.Username
		if config.Password != "" {
			uri += ":" + config.Password
		}
		uri += "@"
	}
	
	// 添加主机和端口
	host := config.Host
	if host == "" {
		host = "localhost"
	}
	
	port := config.Port
	if port == 0 {
		port = 27017
	}
	
	uri += fmt.Sprintf("%s:%d", host, port)
	
	// 添加数据库名
	if config.Database != "" {
		uri += "/" + config.Database
	}
	
	// 添加参数
	if len(config.Options) > 0 {
		uri += "?"
		first := true
		for key, value := range config.Options {
			if !first {
				uri += "&"
			}
			uri += fmt.Sprintf("%s=%s", key, value)
			first = false
		}
	}
	
	return uri
}

// MongoTransaction MongoDB事务实现
type MongoTransaction struct {
	session    mongo.Session
	connection *MongoConnection
}

// Query 事务中的查询（适配SQL接口）
func (t *MongoTransaction) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return nil, NewError(ErrCodeNotImplemented, "MongoDB事务不支持SQL查询")
}

// QueryRow 事务中的单行查询（适配SQL接口）
func (t *MongoTransaction) QueryRow(query string, args ...interface{}) *sql.Row {
	return &sql.Row{}
}

// Exec 事务中的执行（适配SQL接口）
func (t *MongoTransaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	return nil, NewError(ErrCodeNotImplemented, "MongoDB事务不支持SQL执行")
}

// Commit 提交事务
func (t *MongoTransaction) Commit() error {
	if t.session == nil {
		return NewError(ErrCodeTransactionNotStarted, "MongoDB事务未启动")
	}
	
	err := t.session.CommitTransaction(context.TODO())
	if err != nil {
		return WrapError(err, ErrCodeTransactionCommitFailed, "MongoDB事务提交失败")
	}
	
	t.session.EndSession(context.TODO())
	t.session = nil
	
	return nil
}

// Rollback 回滚事务
func (t *MongoTransaction) Rollback() error {
	if t.session == nil {
		return NewError(ErrCodeTransactionNotStarted, "MongoDB事务未启动")
	}
	
	err := t.session.AbortTransaction(context.TODO())
	if err != nil {
		return WrapError(err, ErrCodeTransactionRollbackFailed, "MongoDB事务回滚失败")
	}
	
	t.session.EndSession(context.TODO())
	t.session = nil
	
	return nil
}

// GetSession 获取MongoDB会话
func (t *MongoTransaction) GetSession() mongo.Session {
	return t.session
}
