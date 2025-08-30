package torm

import (
	"time"

	"github.com/zhoudm1743/torm/db"
	"github.com/zhoudm1743/torm/logger"
	"github.com/zhoudm1743/torm/migration"
	"github.com/zhoudm1743/torm/model"
)

// 导出核心类型
type (
	// 数据库相关
	Config               = db.Config
	ConnectionInterface  = db.ConnectionInterface
	TransactionInterface = db.TransactionInterface
	QueryBuilder         = db.QueryBuilder
	Manager              = db.Manager

	// 错误相关
	TormError = db.TormError
	ErrorCode = db.ErrorCode

	// 模型相关
	BaseModel = model.BaseModel

	// 迁移相关
	Migration = migration.Migration

	// 日志相关
	Logger = logger.Logger
)

// 导出核心函数
var (
	// 数据库连接管理
	NewManager     = db.NewManager
	DefaultManager = db.DefaultManager
	AddConnection  = db.AddConnection
	DB             = db.DB
	Table          = db.Table
	Model          = db.Model
	Transaction    = db.Transaction

	// 模型相关
	NewModel = model.NewModel

	// MongoDB相关
	MongoTable        = db.MongoTable
	MongoModel        = db.MongoModel
	NewMongoAggregate = db.NewMongoAggregate

	// 日志相关
	NewLogger            = logger.NewLogger
	NewSQLLogger         = logger.NewSQLLogger
	NewManagerWithLogger = db.NewManagerWithLogger

	// 迁移相关
	NewMigration = migration.NewMigration

	// 缓存相关
	ClearCacheByTags = db.ClearCacheByTags
	ClearAllCache    = db.ClearAllCache
	GetCacheStats    = db.GetCacheStats

	// 连接池相关
	GetConnectionStats    = db.GetConnectionStats
	GetHealthyConnections = db.GetHealthyConnections
	WarmUpConnections     = db.WarmUpConnections
	CloseAllConnections   = db.CloseAllConnections

	// 错误相关
	ErrCodeQueryFailed     = db.ErrCodeQueryFailed
	ErrCodeModelSaveFailed = db.ErrCodeModelSaveFailed
	NewError               = db.NewError
	WrapError              = db.WrapError
	IsQueryError           = db.IsQueryError
	IsModelError           = db.IsModelError
	IsNotFoundError        = db.IsNotFoundError
)

// SetLogger 设置默认管理器的日志记录器
func SetLogger(log db.LoggerInterface) {
	DefaultManager().SetLogger(log)
}

// EnableSQLLogging 启用SQL日志记录
func EnableSQLLogging() {
	sqlLogger := logger.NewSQLLogger(logger.DEBUG, true)
	SetLogger(sqlLogger)
}

// SetSQLLogging 设置SQL日志记录
func SetSQLLogging(level logger.LogLevel, enabled bool) {
	sqlLogger := logger.NewSQLLogger(level, enabled)
	SetLogger(sqlLogger)
}

// SetConnectionPoolConfig 设置连接池配置
func SetConnectionPoolConfig(maxConnections int, connectionTimeout, idleTimeout, cleanupInterval time.Duration) {
	db.SetConnectionPoolConfig(maxConnections, connectionTimeout, idleTimeout, cleanupInterval)
}

// Version 返回 TORM 版本
func Version() string {
	return "1.2.12"
}
