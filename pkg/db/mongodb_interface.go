package db

import (
	"context"
	"database/sql"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoConnectionInterface MongoDB连接接口
type MongoConnectionInterface interface {
	Connect(ctx context.Context) error
	Close() error
	Ping(ctx context.Context) error
	IsConnected() bool
	GetClient() *mongo.Client
	GetDatabase() *mongo.Database
	GetCollection(name string) *mongo.Collection
	GetConfig() *Config
	GetDriver() string

	// 事务支持
	BeginMongo(ctx context.Context) (MongoTransactionInterface, error)
	BeginMongoTx(ctx context.Context, opts *options.TransactionOptions) (MongoTransactionInterface, error)
}

// MongoTransactionInterface MongoDB事务接口
type MongoTransactionInterface interface {
	GetSession() mongo.Session
	Commit() error
	Rollback() error
}

// 让MongoDBConnection也实现通用的ConnectionInterface（部分方法）
type mongoConnectionAdapter struct {
	*MongoDBConnection
}

// Query 适配器方法
func (m *mongoConnectionAdapter) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, fmt.Errorf("Query method not supported for MongoDB, use GetCollection instead")
}

// QueryRow 适配器方法
func (m *mongoConnectionAdapter) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return nil
}

// Exec 适配器方法
func (m *mongoConnectionAdapter) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, fmt.Errorf("Exec method not supported for MongoDB, use GetCollection instead")
}

// Begin 适配器方法
func (m *mongoConnectionAdapter) Begin(ctx context.Context) (TransactionInterface, error) {
	return nil, fmt.Errorf("Begin method not supported for MongoDB, use BeginMongo instead")
}

// BeginTx 适配器方法
func (m *mongoConnectionAdapter) BeginTx(ctx context.Context, opts *sql.TxOptions) (TransactionInterface, error) {
	return nil, fmt.Errorf("BeginTx method not supported for MongoDB, use BeginMongoTx instead")
}

// GetStats 适配器方法
func (m *mongoConnectionAdapter) GetStats() sql.DBStats {
	return sql.DBStats{}
}

// NewMongoDBConnectionAdapter 创建MongoDB连接适配器
func NewMongoDBConnectionAdapter(mongoConn *MongoDBConnection) ConnectionInterface {
	return &mongoConnectionAdapter{MongoDBConnection: mongoConn}
}

// GetMongoConnection 从通用连接接口获取MongoDB连接
func GetMongoConnection(conn ConnectionInterface) *MongoDBConnection {
	if adapter, ok := conn.(*mongoConnectionAdapter); ok {
		return adapter.MongoDBConnection
	}
	return nil
}
