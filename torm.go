// Package torm 是一个现代化的 Go ORM 库，支持多种数据库
package torm

import (
	"github.com/zhoudm1743/torm/cache"
	"github.com/zhoudm1743/torm/db"
	"github.com/zhoudm1743/torm/logger"
	"github.com/zhoudm1743/torm/migration"
	"github.com/zhoudm1743/torm/model"
	"github.com/zhoudm1743/torm/query"
)

// 导出主要类型和接口
type (
	// 数据库相关
	Config               = db.Config
	ConnectionInterface  = db.ConnectionInterface
	TransactionInterface = db.TransactionInterface
	QueryInterface       = db.QueryInterface
	LoggerInterface      = db.LoggerInterface
	Manager              = db.Manager

	// 模型相关
	BaseModel = model.BaseModel

	// 查询相关
	AdvancedQueryBuilder = query.AdvancedQueryBuilder

	// 缓存相关
	MemoryCache = cache.MemoryCache

	// 迁移相关
	Migration = migration.Migration

	// 日志相关
	Logger = logger.Logger
)

// 导出主要函数
var (
	// 数据库连接管理
	NewManager     = db.NewManager
	DefaultManager = db.DefaultManager
	AddConnection  = db.AddConnection
	DB             = db.DB
	Query          = db.Query
	Table          = db.Table
	Raw            = db.Raw
	Exec           = db.Exec
	Transaction    = db.Transaction

	// 日志相关
	NewLogger = logger.NewLogger

	// 缓存相关
	NewMemoryCache = cache.NewMemoryCache

	// 迁移相关
	NewMigration = migration.NewMigration
)

// Version 返回 TORM 版本
func Version() string {
	return "1.1.3"
}
