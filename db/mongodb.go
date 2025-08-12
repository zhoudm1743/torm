package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoDBConnection MongoDB数据库连接
type MongoDBConnection struct {
	config *Config
	client *mongo.Client
	db     *mongo.Database
	logger LoggerInterface
}

// NewMongoDBConnection 创建MongoDB连接
func NewMongoDBConnection(config *Config, logger LoggerInterface) (ConnectionInterface, error) {
	conn := &MongoDBConnection{
		config: config,
		logger: logger,
	}
	return NewMongoDBConnectionAdapter(conn), nil
}

// Connect 连接到MongoDB数据库
func (c *MongoDBConnection) Connect() error {
	// 创建内部context
	ctx := context.Background()

	uri := c.config.DSN()
	if c.logger != nil {
		c.logger.Debug("Connecting to MongoDB", "uri", uri)
	}

	// 设置客户端选项
	clientOptions := options.Client().ApplyURI(uri)

	// 配置连接池
	if c.config.MaxOpenConns > 0 {
		maxPoolSize := uint64(c.config.MaxOpenConns)
		clientOptions.SetMaxPoolSize(maxPoolSize)
	}
	if c.config.MaxIdleConns > 0 {
		minPoolSize := uint64(c.config.MaxIdleConns)
		clientOptions.SetMinPoolSize(minPoolSize)
	}
	if c.config.ConnMaxLifetime > 0 {
		clientOptions.SetMaxConnIdleTime(c.config.ConnMaxLifetime)
	}

	// 创建客户端
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Failed to connect to MongoDB", "error", err)
		}
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// 测试连接
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		client.Disconnect(ctx)
		if c.logger != nil {
			c.logger.Error("Failed to ping MongoDB database", "error", err)
		}
		return fmt.Errorf("failed to ping MongoDB database: %w", err)
	}

	c.client = client
	c.db = client.Database(c.config.Database)

	if c.logger != nil {
		c.logger.Info("MongoDB connection established successfully")
	}
	return nil
}

// Close 关闭连接
func (c *MongoDBConnection) Close() error {
	if c.client != nil {
		if c.logger != nil {
			c.logger.Debug("Closing MongoDB connection")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := c.client.Disconnect(ctx)
		c.client = nil
		c.db = nil
		return err
	}
	return nil
}

// Ping 测试连接
func (c *MongoDBConnection) Ping() error {
	if c.client == nil {
		return fmt.Errorf("MongoDB connection is not established")
	}
	// 创建内部context
	ctx := context.Background()
	return c.client.Ping(ctx, readpref.Primary())
}

// IsConnected 检查连接状态
func (c *MongoDBConnection) IsConnected() bool {
	if c.client == nil {
		return false
	}
	// 创建内部context
	ctx := context.Background()
	return c.client.Ping(ctx, readpref.Primary()) == nil
}

// GetClient 获取MongoDB客户端
func (c *MongoDBConnection) GetClient() *mongo.Client {
	return c.client
}

// GetDatabase 获取数据库对象
func (c *MongoDBConnection) GetDatabase() *mongo.Database {
	return c.db
}

// GetCollection 获取集合对象
func (c *MongoDBConnection) GetCollection(name string) *mongo.Collection {
	if c.db == nil {
		return nil
	}
	return c.db.Collection(name)
}

// GetConfig 获取配置
func (c *MongoDBConnection) GetConfig() *Config {
	return c.config
}

// GetDriver 获取驱动名称
func (c *MongoDBConnection) GetDriver() string {
	return "mongodb"
}

// BeginMongo 开始MongoDB事务
func (c *MongoDBConnection) BeginMongo(ctx context.Context) (MongoTransactionInterface, error) {
	if c.client == nil {
		return nil, fmt.Errorf("MongoDB connection is not established")
	}

	session, err := c.client.StartSession()
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Failed to start MongoDB session", "error", err)
		}
		return nil, fmt.Errorf("failed to start MongoDB session: %w", err)
	}

	if c.logger != nil {
		c.logger.Debug("MongoDB session started")
	}

	return &MongoDBTransaction{
		session: session,
		logger:  c.logger,
		config:  c.config,
	}, nil
}

// BeginMongoTx 开始MongoDB事务（带选项）
func (c *MongoDBConnection) BeginMongoTx(ctx context.Context, opts *options.TransactionOptions) (MongoTransactionInterface, error) {
	if c.client == nil {
		return nil, fmt.Errorf("MongoDB connection is not established")
	}

	session, err := c.client.StartSession()
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Failed to start MongoDB session", "error", err)
		}
		return nil, fmt.Errorf("failed to start MongoDB session: %w", err)
	}

	if err := session.StartTransaction(opts); err != nil {
		session.EndSession(ctx)
		if c.logger != nil {
			c.logger.Error("Failed to start MongoDB transaction", "error", err)
		}
		return nil, fmt.Errorf("failed to start MongoDB transaction: %w", err)
	}

	if c.logger != nil {
		c.logger.Debug("MongoDB transaction started")
	}

	return &MongoDBTransaction{
		session: session,
		logger:  c.logger,
		config:  c.config,
	}, nil
}

// MongoDBTransaction MongoDB事务
type MongoDBTransaction struct {
	session mongo.Session
	logger  LoggerInterface
	config  *Config
}

// GetSession 获取会话
func (t *MongoDBTransaction) GetSession() mongo.Session {
	return t.session
}

// Commit 提交事务
func (t *MongoDBTransaction) Commit() error {
	if t.session == nil {
		return fmt.Errorf("MongoDB session is not active")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := t.session.CommitTransaction(ctx)
	if err != nil {
		if t.logger != nil {
			t.logger.Error("Failed to commit MongoDB transaction", "error", err)
		}
		t.session.EndSession(ctx)
		t.session = nil
		return fmt.Errorf("failed to commit MongoDB transaction: %w", err)
	}

	if t.logger != nil {
		t.logger.Debug("MongoDB transaction committed")
	}

	t.session.EndSession(ctx)
	t.session = nil
	return nil
}

// Rollback 回滚事务
func (t *MongoDBTransaction) Rollback() error {
	if t.session == nil {
		return fmt.Errorf("MongoDB session is not active")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := t.session.AbortTransaction(ctx)
	if err != nil {
		if t.logger != nil {
			t.logger.Error("Failed to rollback MongoDB transaction", "error", err)
		}
	} else if t.logger != nil {
		t.logger.Debug("MongoDB transaction rolled back")
	}

	t.session.EndSession(ctx)
	t.session = nil
	return err
}
