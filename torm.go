// Package torm 是一个高性能的 Go ORM 框架
// 参考 ThinkORM 4.0 架构设计，对 database/sql 进行深度封装
//
// 主要特性:
//   - 支持 MySQL、PostgreSQL、SQLite 多种数据库
//   - 链式查询构建器
//   - 模型系统（支持时间戳、软删除、访问器/修改器）
//   - 关联关系（HasOne、HasMany、BelongsTo、BelongsToMany）
//   - 事务支持（包括嵌套事务）
//   - 查询缓存
//   - 数据库迁移
//   - 读写分离
//   - 连接池管理
package torm

import (
	"time"

	"github.com/zhoudm1743/torm/db"
	"github.com/zhoudm1743/torm/logger"
	"github.com/zhoudm1743/torm/migration"
	"github.com/zhoudm1743/torm/model"
)

// Version TORM 版本号
const Version = "1.2.12"

// GetVersion 返回 TORM 版本号
func GetVersion() string {
	return Version
}

// =============================================================================
// 数据库配置类型
// =============================================================================

type (
	// Config 数据库配置
	Config = db.Config
)

// =============================================================================
// 连接和事务类型
// =============================================================================

type (
	// Connection 数据库连接
	Connection = db.Connection
	// ConnectionInterface 数据库连接接口
	ConnectionInterface = db.ConnectionInterface
	// TransactionInterface 事务接口
	TransactionInterface = db.TransactionInterface
	// ConnectionStats 连接统计信息
	ConnectionStats = db.ConnectionStats
)

// =============================================================================
// 查询构建器类型
// =============================================================================

type (
	// QueryBuilder 查询构建器
	QueryBuilder = db.QueryBuilder
	// QueryInterface 查询接口
	QueryInterface = db.QueryInterface
	// BuilderInterface SQL构建器接口
	BuilderInterface = db.BuilderInterface
)

// =============================================================================
// 数据库管理器类型
// =============================================================================

type (
	// DbManager 数据库管理器
	DbManager = db.DbManager
)

// =============================================================================
// 错误处理类型
// =============================================================================

type (
	// TormError TORM错误类型
	TormError = db.TormError
	// ErrorCode 错误码类型
	ErrorCode = db.ErrorCode
)

// =============================================================================
// 分页类型
// =============================================================================

type (
	// SimplePagination 简单分页结果
	SimplePagination = db.SimplePagination
	// Paginator 分页器
	Paginator = db.Paginator
	// SimplePaginator 简单分页器
	SimplePaginator = db.SimplePaginator
)

// =============================================================================
// 缓存类型
// =============================================================================

type (
	// CacheInterface 缓存接口
	CacheInterface = db.CacheInterface
	// CacheWithTagsInterface 带标签的缓存接口
	CacheWithTagsInterface = db.CacheWithTagsInterface
	// CacheWithBatchInterface 支持批量操作的缓存接口
	CacheWithBatchInterface = db.CacheWithBatchInterface
	// CacheWithStatsInterface 支持统计的缓存接口
	CacheWithStatsInterface = db.CacheWithStatsInterface
	// FullCacheInterface 完整缓存接口
	FullCacheInterface = db.FullCacheInterface
)

// =============================================================================
// 日志类型
// =============================================================================

type (
	// LoggerInterface 日志接口
	LoggerInterface = db.LoggerInterface
	// Logger 日志记录器
	Logger = logger.Logger
	// SQLLogger SQL日志记录器
	SQLLogger = logger.SQLLogger
	// LogLevel 日志级别
	LogLevel = logger.LogLevel
)

// =============================================================================
// 原生SQL查询结果类型
// =============================================================================

type (
	// RawQueryResult 原生查询结果
	RawQueryResult = db.RawQueryResult
	// RawExecResult 原生执行结果
	RawExecResult = db.RawExecResult
)

// =============================================================================
// 读写分离类型
// =============================================================================

type (
	// ReadWriteSplitConnection 读写分离连接
	ReadWriteSplitConnection = db.ReadWriteSplitConnection
	// MasterOnlyConnection 仅主库连接
	MasterOnlyConnection = db.MasterOnlyConnection
	// LoadBalancer 负载均衡器接口
	LoadBalancer = db.LoadBalancer
	// LoadBalancerType 负载均衡类型
	LoadBalancerType = db.LoadBalancerType
)

// =============================================================================
// 模型类型
// =============================================================================

type (
	// BaseModel 基础模型
	BaseModel = model.BaseModel
	// Model 模型接口
	Model = model.Model
	// SoftDeleteModel 软删除模型接口
	SoftDeleteModel = model.SoftDeleteModel
	// TimestampModel 时间戳模型接口
	TimestampModel = model.TimestampModel
	// ModelConfig 模型配置
	ModelConfig = model.ModelConfig
	// ModelMeta 模型元数据
	ModelMeta = model.ModelMeta
	// FieldMeta 字段元数据
	FieldMeta = model.FieldMeta
)

// =============================================================================
// 模型字段类型
// =============================================================================

type (
	// JSONField JSON字段类型
	JSONField = model.JSONField
	// JSONArray JSON数组类型
	JSONArray = model.JSONArray
	// DeletedAt 软删除时间类型
	DeletedAt = model.DeletedAt
)

// =============================================================================
// 关联关系类型
// =============================================================================

type (
	// RelationInterface 关联接口
	RelationInterface = model.RelationInterface
	// RelationType 关联类型
	RelationType = model.RelationType
	// HasOne 一对一关联
	HasOne = model.HasOne
	// HasMany 一对多关联
	HasMany = model.HasMany
	// BelongsTo 反向关联
	BelongsTo = model.BelongsTo
	// BelongsToMany 多对多关联
	BelongsToMany = model.BelongsToMany
)

// =============================================================================
// 迁移类型
// =============================================================================

type (
	// Migration 迁移
	Migration = migration.Migration
	// MigrationInterface 迁移接口
	MigrationInterface = migration.MigrationInterface
	// Migrator 迁移器
	Migrator = migration.Migrator
	// MigrationStatus 迁移状态
	MigrationStatus = migration.MigrationStatus
	// MigrationRecord 迁移记录
	MigrationRecord = migration.MigrationRecord
	// AutoMigrator 自动迁移器
	AutoMigrator = migration.AutoMigrator
)

// =============================================================================
// 日志级别常量
// =============================================================================

const (
	DEBUG  = logger.DEBUG
	INFO   = logger.INFO
	WARN   = logger.WARN
	ERROR  = logger.ERROR
	FATAL  = logger.FATAL
	SILENT = logger.SILENT
)

// =============================================================================
// 关联类型常量
// =============================================================================

const (
	RelationHasOne        = model.RelationHasOne
	RelationHasMany       = model.RelationHasMany
	RelationBelongsTo     = model.RelationBelongsTo
	RelationBelongsToMany = model.RelationBelongsToMany
)

// =============================================================================
// 负载均衡类型常量
// =============================================================================

const (
	LoadBalancerRoundRobin = db.LoadBalancerRoundRobin
	LoadBalancerRandom     = db.LoadBalancerRandom
	LoadBalancerWeighted   = db.LoadBalancerWeighted
)

// =============================================================================
// 错误码常量
// =============================================================================

var (
	ErrCodeQueryFailed        = db.ErrCodeQueryFailed
	ErrCodeModelSaveFailed    = db.ErrCodeModelSaveFailed
	ErrCodeConnectionFailed   = db.ErrCodeConnectionFailed
	ErrCodeTransactionFailed  = db.ErrCodeTransactionFailed
	ErrCodeConfigurationError = db.ErrCodeConfigurationError
	ErrCodeDriverNotSupported = db.ErrCodeDriverNotSupported
	ErrCodeRecordNotFound     = db.ErrCodeRecordNotFound
	ErrCodeMigrationFailed    = db.ErrCodeMigrationFailed
)

// =============================================================================
// 预定义错误
// =============================================================================

var (
	ErrRecordNotFound   = db.ErrRecordNotFound
	ErrInvalidParameter = db.ErrInvalidParameter
)

// =============================================================================
// 数据库连接管理函数
// =============================================================================

var (
	NewDbManager           = db.NewDbManager
	NewDbManagerWithLogger = db.NewDbManagerWithLogger
	DefaultManager         = db.DefaultManager
	AddConnection          = db.AddConnection
	DB                     = db.DB
	Table                  = db.Table
	CloseAllConnections    = db.CloseAllConnections
)

// Transaction 执行事务
func Transaction(fn func(tx TransactionInterface) error, connectionName ...string) error {
	connName := "default"
	if len(connectionName) > 0 && connectionName[0] != "" {
		connName = connectionName[0]
	}

	conn, err := db.DB(connName)
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

// =============================================================================
// 模型函数
// =============================================================================

var (
	NewModel         = model.NewModel
	GetModelMeta     = model.GetModelMeta
	NewQueryForModel = model.NewQueryForModel
)

// =============================================================================
// 日志函数
// =============================================================================

var (
	NewLogger            = logger.NewLogger
	NewSQLLogger         = logger.NewSQLLogger
	NewNullLogger        = logger.NewNullLogger
	NewDebugLogger       = logger.NewDebugLogger
	NewInfoLogger        = logger.NewInfoLogger
	NewQuietLogger       = logger.NewQuietLogger
	NewSilentLogger      = logger.NewSilentLogger
	NewProductionLogger  = logger.NewProductionLogger
	NewDevelopmentLogger = logger.NewDevelopmentLogger
	ParseLogLevel        = logger.ParseLogLevel
)

// =============================================================================
// 迁移函数
// =============================================================================

var (
	NewMigration    = migration.NewMigration
	NewMigrator     = migration.NewMigrator
	NewAutoMigrator = migration.NewAutoMigrator
)

// =============================================================================
// 缓存函数
// =============================================================================

var (
	GetDefaultCache  = db.GetDefaultCache
	ClearCacheByTags = db.ClearCacheByTags
	ClearAllCache    = db.ClearAllCache
	GetCacheStats    = db.GetCacheStats
)

// =============================================================================
// 连接池函数
// =============================================================================

var (
	GetConnectionStats    = db.GetConnectionStats
	GetHealthyConnections = db.GetHealthyConnections
	WarmUpConnections     = db.WarmUpConnections
)

// =============================================================================
// 错误处理函数
// =============================================================================

var (
	NewError        = db.NewError
	NewErrorf       = db.NewErrorf
	WrapError       = db.WrapError
	IsQueryError    = db.IsQueryError
	IsModelError    = db.IsModelError
	IsNotFoundError = db.IsNotFoundError
)

// =============================================================================
// 原生SQL查询函数
// =============================================================================

var (
	RawQuery   = db.RawQuery
	RawQueryOn = db.RawQueryOn
	RawExec    = db.RawExec
	RawExecOn  = db.RawExecOn
)

// =============================================================================
// 读写分离函数
// =============================================================================

var (
	UseMasterDB           = db.UseMasterDB
	GetSlaveCountDB       = db.GetSlaveCountDB
	SetLoadBalancerDB     = db.SetLoadBalancerDB
	SetLoadBalancerTypeDB = db.SetLoadBalancerTypeDB
)

// =============================================================================
// 日志配置便捷函数
// =============================================================================

// SetLogger 设置默认管理器的日志记录器
func SetLogger(log LoggerInterface) {
	DefaultManager().SetLogger(log)
}

// EnableSQLLogging 启用SQL日志记录
func EnableSQLLogging() {
	sqlLogger := logger.NewSQLLogger(logger.DEBUG, true)
	SetLogger(sqlLogger)
}

// SetSQLLogging 设置SQL日志记录
func SetSQLLogging(level LogLevel, enabled bool) {
	sqlLogger := logger.NewSQLLogger(level, enabled)
	SetLogger(sqlLogger)
}

// =============================================================================
// 连接池配置便捷函数
// =============================================================================

// SetConnectionPoolConfig 设置连接池配置
func SetConnectionPoolConfig(maxConnections int, connectionTimeout, idleTimeout, cleanupInterval time.Duration) {
	db.SetConnectionPoolConfig(maxConnections, connectionTimeout, idleTimeout, cleanupInterval)
}
