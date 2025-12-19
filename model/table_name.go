package model

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// ============================================================================
// 自定义表名支持
// ============================================================================

// TableNamer 表名接口 - 模型可以实现此接口来自定义表名
// 这是Tabler接口的别名，提供更好的语义
type TableNamer interface {
	TableName() string
}

// DynamicTableNamer 动态表名接口 - 支持运行时动态设置表名
// 实现此接口的模型可以在运行时更改表名
type DynamicTableNamer interface {
	TableNamer
	SetTableName(name string)
}

// TableNameResolver 表名解析器接口 - 用于自定义表名解析逻辑
type TableNameResolver interface {
	ResolveTableName(model interface{}) string
}

// TableNameConfig 表名配置
type TableNameConfig struct {
	Prefix        string // 表名前缀
	Suffix        string // 表名后缀
	SingularTable bool   // 是否使用单数表名（默认复数）
}

// 全局表名配置
var (
	globalTableNameConfig = TableNameConfig{
		Prefix:        "",
		Suffix:        "",
		SingularTable: true, // 默认使用单数表名
	}
	tableNameConfigMutex sync.RWMutex
)

// SetTableNameConfig 设置全局表名配置
func SetTableNameConfig(config TableNameConfig) {
	tableNameConfigMutex.Lock()
	defer tableNameConfigMutex.Unlock()
	globalTableNameConfig = config
}

// GetTableNameConfig 获取全局表名配置
func GetTableNameConfig() TableNameConfig {
	tableNameConfigMutex.RLock()
	defer tableNameConfigMutex.RUnlock()
	return globalTableNameConfig
}

// SetTablePrefix 设置全局表名前缀
func SetTablePrefix(prefix string) {
	tableNameConfigMutex.Lock()
	defer tableNameConfigMutex.Unlock()
	globalTableNameConfig.Prefix = prefix
}

// SetTableSuffix 设置全局表名后缀
func SetTableSuffix(suffix string) {
	tableNameConfigMutex.Lock()
	defer tableNameConfigMutex.Unlock()
	globalTableNameConfig.Suffix = suffix
}

// SetSingularTable 设置是否使用单数表名
func SetSingularTable(singular bool) {
	tableNameConfigMutex.Lock()
	defer tableNameConfigMutex.Unlock()
	globalTableNameConfig.SingularTable = singular
}

// ResolveTableName 解析表名
// 优先级：
// 1. 模型实现的GetTableName()方法（优先，因为这是自定义方法）
// 2. 模型实现的TableName()方法（Tabler接口）
// 3. 从结构体名称推断
func ResolveTableName(model interface{}) string {
	// 1. 检查是否有GetTableName方法（优先级最高，因为这是自定义方法）
	if getter, ok := model.(interface{ GetTableName() string }); ok {
		if name := getter.GetTableName(); name != "" {
			return applyTableNameConfig(name)
		}
	}

	// 2. 检查是否实现了Tabler接口（TableName方法）
	// 注意：需要检查返回值是否为空，因为嵌入的BaseModel可能返回空字符串
	if tabler, ok := model.(Tabler); ok {
		if name := tabler.TableName(); name != "" {
			return applyTableNameConfig(name)
		}
	}

	// 3. 检查是否实现了TableNamer接口
	if namer, ok := model.(TableNamer); ok {
		if name := namer.TableName(); name != "" {
			return applyTableNameConfig(name)
		}
	}

	// 4. 从结构体名称推断
	return inferTableName(model)
}

// inferTableName 从结构体名称推断表名
func inferTableName(model interface{}) string {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 获取结构体名称
	name := modelType.Name()

	// 转换为蛇形命名
	tableName := toSnakeCase(name)

	// 应用全局配置
	return applyTableNameConfig(tableName)
}

// applyTableNameConfig 应用表名配置
func applyTableNameConfig(tableName string) string {
	config := GetTableNameConfig()

	// 如果表名已经包含前缀，不再添加
	if config.Prefix != "" && !strings.HasPrefix(tableName, config.Prefix) {
		tableName = config.Prefix + tableName
	}

	// 如果表名已经包含后缀，不再添加
	if config.Suffix != "" && !strings.HasSuffix(tableName, config.Suffix) {
		tableName = tableName + config.Suffix
	}

	return tableName
}

// ============================================================================
// BaseModel 表名方法扩展
// ============================================================================

// SetTablePrefix 设置模型的表名前缀
func (m *BaseModel) SetTablePrefix(prefix string) *BaseModel {
	if m.config.TableName != "" && !strings.HasPrefix(m.config.TableName, prefix) {
		m.config.TableName = prefix + m.config.TableName
	}
	return m
}

// SetTableSuffix 设置模型的表名后缀
func (m *BaseModel) SetTableSuffix(suffix string) *BaseModel {
	if m.config.TableName != "" && !strings.HasSuffix(m.config.TableName, suffix) {
		m.config.TableName = m.config.TableName + suffix
	}
	return m
}

// ============================================================================
// 表名工具函数
// ============================================================================

// TableNameFromStruct 从结构体获取表名
func TableNameFromStruct(structType interface{}) string {
	return ResolveTableName(structType)
}

// TableNameFromType 从反射类型获取表名
func TableNameFromType(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// 创建零值实例
	instance := reflect.New(t).Interface()

	// 检查是否实现了Tabler接口
	if tabler, ok := instance.(Tabler); ok {
		return applyTableNameConfig(tabler.TableName())
	}

	// 从类型名推断
	return applyTableNameConfig(toSnakeCase(t.Name()))
}

// ============================================================================
// 表名缓存
// ============================================================================

var (
	tableNameCache      = make(map[reflect.Type]string)
	tableNameCacheMutex sync.RWMutex
)

// GetCachedTableName 获取缓存的表名
func GetCachedTableName(model interface{}) string {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 尝试从缓存获取
	tableNameCacheMutex.RLock()
	if name, exists := tableNameCache[modelType]; exists {
		tableNameCacheMutex.RUnlock()
		return name
	}
	tableNameCacheMutex.RUnlock()

	// 解析表名
	name := ResolveTableName(model)

	// 缓存结果
	tableNameCacheMutex.Lock()
	tableNameCache[modelType] = name
	tableNameCacheMutex.Unlock()

	return name
}

// ClearTableNameCache 清除表名缓存
func ClearTableNameCache() {
	tableNameCacheMutex.Lock()
	defer tableNameCacheMutex.Unlock()
	tableNameCache = make(map[reflect.Type]string)
}

// ============================================================================
// 自定义表名解析器
// ============================================================================

var (
	customTableNameResolver TableNameResolver
	resolverMutex           sync.RWMutex
)

// SetTableNameResolver 设置自定义表名解析器
// 允许用户完全自定义表名解析逻辑
func SetTableNameResolver(resolver TableNameResolver) {
	resolverMutex.Lock()
	defer resolverMutex.Unlock()
	customTableNameResolver = resolver
}

// GetTableNameResolver 获取当前的表名解析器
func GetTableNameResolver() TableNameResolver {
	resolverMutex.RLock()
	defer resolverMutex.RUnlock()
	return customTableNameResolver
}

// ClearTableNameResolver 清除自定义表名解析器
func ClearTableNameResolver() {
	resolverMutex.Lock()
	defer resolverMutex.Unlock()
	customTableNameResolver = nil
}

// ============================================================================
// 表名构建器 - 支持链式调用
// ============================================================================

// TableNameBuilder 表名构建器
type TableNameBuilder struct {
	baseName   string
	prefix     string
	suffix     string
	schema     string
	connection string
}

// NewTableNameBuilder 创建表名构建器
func NewTableNameBuilder(baseName string) *TableNameBuilder {
	return &TableNameBuilder{
		baseName: baseName,
	}
}

// WithPrefix 设置前缀
func (b *TableNameBuilder) WithPrefix(prefix string) *TableNameBuilder {
	b.prefix = prefix
	return b
}

// WithSuffix 设置后缀
func (b *TableNameBuilder) WithSuffix(suffix string) *TableNameBuilder {
	b.suffix = suffix
	return b
}

// WithSchema 设置数据库schema（用于PostgreSQL等）
func (b *TableNameBuilder) WithSchema(schema string) *TableNameBuilder {
	b.schema = schema
	return b
}

// WithConnection 设置连接名（用于多数据库场景）
func (b *TableNameBuilder) WithConnection(connection string) *TableNameBuilder {
	b.connection = connection
	return b
}

// Build 构建最终表名
func (b *TableNameBuilder) Build() string {
	name := b.baseName

	// 应用前缀
	if b.prefix != "" && !strings.HasPrefix(name, b.prefix) {
		name = b.prefix + name
	}

	// 应用后缀
	if b.suffix != "" && !strings.HasSuffix(name, b.suffix) {
		name = name + b.suffix
	}

	// 应用schema
	if b.schema != "" {
		name = b.schema + "." + name
	}

	return name
}

// String 返回构建的表名
func (b *TableNameBuilder) String() string {
	return b.Build()
}

// ============================================================================
// 表名工具函数扩展
// ============================================================================

// ResolveTableNameWithResolver 使用自定义解析器解析表名
func ResolveTableNameWithResolver(model interface{}) string {
	// 检查是否有自定义解析器
	resolver := GetTableNameResolver()
	if resolver != nil {
		return resolver.ResolveTableName(model)
	}

	// 使用默认解析逻辑
	return ResolveTableName(model)
}

// TableNameWithSchema 获取带schema的表名
func TableNameWithSchema(model interface{}, schema string) string {
	tableName := ResolveTableName(model)
	if schema != "" {
		return schema + "." + tableName
	}
	return tableName
}

// TableNameWithPrefix 获取带前缀的表名
func TableNameWithPrefix(model interface{}, prefix string) string {
	tableName := ResolveTableName(model)
	if prefix != "" && !strings.HasPrefix(tableName, prefix) {
		return prefix + tableName
	}
	return tableName
}

// TableNameWithSuffix 获取带后缀的表名
func TableNameWithSuffix(model interface{}, suffix string) string {
	tableName := ResolveTableName(model)
	if suffix != "" && !strings.HasSuffix(tableName, suffix) {
		return tableName + suffix
	}
	return tableName
}

// ============================================================================
// 分表支持
// ============================================================================

// ShardingTableNamer 分表接口 - 支持按规则分表
type ShardingTableNamer interface {
	TableNamer
	// ShardingKey 返回分表键
	ShardingKey() string
	// ShardingTableName 根据分表键值返回实际表名
	ShardingTableName(keyValue interface{}) string
}

// ShardingStrategy 分表策略
type ShardingStrategy int

const (
	// ShardingByHash 按哈希分表
	ShardingByHash ShardingStrategy = iota
	// ShardingByRange 按范围分表
	ShardingByRange
	// ShardingByDate 按日期分表
	ShardingByDate
	// ShardingByMod 按取模分表
	ShardingByMod
)

// ShardingConfig 分表配置
type ShardingConfig struct {
	Strategy    ShardingStrategy
	ShardCount  int    // 分表数量（用于Hash和Mod策略）
	DateFormat  string // 日期格式（用于Date策略）
	RangeSize   int64  // 范围大小（用于Range策略）
	TablePrefix string // 表名前缀
	TableSuffix string // 表名后缀
}

// DefaultShardingConfig 默认分表配置
func DefaultShardingConfig() ShardingConfig {
	return ShardingConfig{
		Strategy:   ShardingByMod,
		ShardCount: 10,
		DateFormat: "200601", // YYYYMM
		RangeSize:  1000000,
	}
}

// GetShardingTableName 根据分表配置获取实际表名
func GetShardingTableName(baseTable string, keyValue interface{}, config ShardingConfig) string {
	var suffix string

	switch config.Strategy {
	case ShardingByMod:
		// 按取模分表
		if intVal, ok := toInt64(keyValue); ok {
			suffix = fmt.Sprintf("_%d", intVal%int64(config.ShardCount))
		}
	case ShardingByHash:
		// 按哈希分表
		hash := hashValue(keyValue)
		suffix = fmt.Sprintf("_%d", hash%uint64(config.ShardCount))
	case ShardingByDate:
		// 按日期分表
		if t, ok := keyValue.(interface{ Format(string) string }); ok {
			suffix = "_" + t.Format(config.DateFormat)
		}
	case ShardingByRange:
		// 按范围分表
		if intVal, ok := toInt64(keyValue); ok {
			rangeIndex := intVal / config.RangeSize
			suffix = fmt.Sprintf("_%d", rangeIndex)
		}
	}

	result := baseTable + suffix
	if config.TablePrefix != "" {
		result = config.TablePrefix + result
	}
	if config.TableSuffix != "" {
		result = result + config.TableSuffix
	}

	return result
}

// toInt64 尝试将值转换为int64
func toInt64(v interface{}) (int64, bool) {
	switch val := v.(type) {
	case int:
		return int64(val), true
	case int8:
		return int64(val), true
	case int16:
		return int64(val), true
	case int32:
		return int64(val), true
	case int64:
		return val, true
	case uint:
		return int64(val), true
	case uint8:
		return int64(val), true
	case uint16:
		return int64(val), true
	case uint32:
		return int64(val), true
	case uint64:
		return int64(val), true
	default:
		return 0, false
	}
}

// hashValue 计算值的哈希
func hashValue(v interface{}) uint64 {
	// 简单的FNV-1a哈希实现
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)

	hash := uint64(offset64)
	str := fmt.Sprintf("%v", v)
	for i := 0; i < len(str); i++ {
		hash ^= uint64(str[i])
		hash *= prime64
	}
	return hash
}
