package model

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/zhoudm1743/torm/db"
	"github.com/zhoudm1743/torm/migration"
)

// 确保 BaseModel 实现了 Model 接口
var _ Model = (*BaseModel)(nil)
var _ SoftDeleteModel = (*BaseModel)(nil)
var _ TimestampModel = (*BaseModel)(nil)

// ModelConfig 模型配置
type ModelConfig struct {
	TableName    string
	PrimaryKey   string
	Connection   string
	Timestamps   bool
	CreatedAtCol string
	UpdatedAtCol string
	SoftDeletes  bool
	DeletedAtCol string
}

// DefaultModelConfig 默认模型配置
func DefaultModelConfig() ModelConfig {
	return ModelConfig{
		PrimaryKey:   "id",
		Connection:   "default",
		Timestamps:   true,
		CreatedAtCol: "created_at",
		UpdatedAtCol: "updated_at",
		SoftDeletes:  false,
		DeletedAtCol: "deleted_at",
	}
}

// BaseModel 基础模型 - 重构版本
// 职责分离：模型专注于数据操作，查询通过组合方式提供
type BaseModel struct {
	// 模型配置
	config ModelConfig

	// 模型数据
	attributes map[string]interface{}

	// 原始数据（用于脏数据追踪）
	original map[string]interface{}

	// 模型状态
	exists bool

	// 时间管理
	timeManager *db.TimeFieldManager
	timeFields  []db.TimeFieldInfo

	// 互斥锁（保护并发访问）
	mu sync.RWMutex
}

// NewModel 创建模型 - 简化和优化版本
// NewModel(tableName) - 创建指定表名的模型
// NewModel(config) - 使用配置创建模型
// NewModel(structInstance) - 从结构体创建模型并解析标签
// NewModel(structInstance, config) - 从结构体创建模型并合并配置（torm标签优先级最高）
func NewModel(args ...interface{}) *BaseModel {
	config := DefaultModelConfig()
	var structInstance interface{}

	// 解析参数
	switch len(args) {
	case 0:
		// 无参数，使用默认配置
	case 1:
		switch arg := args[0].(type) {
		case string:
			// 表名
			config.TableName = arg
		case ModelConfig:
			// 配置对象
			config = arg
		default:
			// 结构体实例，解析标签
			structInstance = arg
			config = parseModelFromStruct(arg)
		}
	case 2:
		// 可能是：tableName + connection 或 structInstance + config
		if str, ok := args[0].(string); ok {
			// 表名和连接
			config.TableName = str
			if connection, ok := args[1].(string); ok {
				config.Connection = connection
			}
		} else {
			// 结构体实例 + 配置对象
			structInstance = args[0]
			if userConfig, ok := args[1].(ModelConfig); ok {
				// 使用新的函数来处理结构体和用户配置的合并
				config = parseModelFromStructWithConfig(structInstance, userConfig)
			} else {
				// 只是结构体实例
				config = parseModelFromStruct(structInstance)
			}
		}
	}

	model := &BaseModel{
		config:      config,
		attributes:  make(map[string]interface{}),
		original:    make(map[string]interface{}),
		exists:      false,
		timeManager: db.NewTimeFieldManager(),
		timeFields:  make([]db.TimeFieldInfo, 0),
	}

	// 如果有结构体实例，分析时间字段
	if structInstance != nil && model.timeManager != nil {
		model.timeFields = model.timeManager.AnalyzeModelTimeFields(structInstance)
	}

	return model
}

// parseModelFromStruct 从结构体解析模型配置
func parseModelFromStruct(structInstance interface{}) ModelConfig {
	config := DefaultModelConfig()

	// 尝试从实例获取表名
	if modeler, ok := structInstance.(interface{ GetTableName() string }); ok {
		if tableName := modeler.GetTableName(); tableName != "" {
			config.TableName = tableName
		}
	}

	// 如果还是没有表名，从类型推断（作为备用）
	if config.TableName == "" {
		modelType := reflect.TypeOf(structInstance)
		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}
		config.TableName = toSnakeCase(modelType.Name())
	}

	// 解析torm标签
	if structInstance != nil {
		parseTagsIntoConfig(structInstance, &config)
	}

	return config
}

// parseModelFromStructWithConfig 从结构体解析模型配置并应用用户配置
func parseModelFromStructWithConfig(structInstance interface{}, userConfig ModelConfig) ModelConfig {
	// 先获取结构体的基础配置
	config := DefaultModelConfig()

	// 应用用户配置作为基础
	config = userConfig

	// 尝试从实例获取表名（优先级高于用户配置）
	if modeler, ok := structInstance.(interface{ GetTableName() string }); ok {
		if tableName := modeler.GetTableName(); tableName != "" {
			config.TableName = tableName
		}
	}

	// 如果还是没有表名，保持用户配置的表名（如果有的话）
	// 如果用户也没有配置表名，则从类型推断
	if config.TableName == "" {
		modelType := reflect.TypeOf(structInstance)
		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}
		config.TableName = toSnakeCase(modelType.Name())
	}

	// 解析torm标签，这将覆盖用户配置中的对应字段
	if structInstance != nil {
		parseTagsIntoConfig(structInstance, &config)
	}

	return config
}

// ============================================================================
// 模型配置方法
// ============================================================================

// SetTable 设置表名
func (m *BaseModel) SetTable(tableName string) *BaseModel {
	m.config.TableName = tableName
	return m
}

// Table 设置表名（GORM兼容方法）
// 这是SetTable的别名，提供与GORM一致的API
func (m *BaseModel) Table(tableName string) *BaseModel {
	return m.SetTable(tableName)
}

// GetTableName 获取表名
func (m *BaseModel) GetTableName() string {
	return m.config.TableName
}

// TableName 静态函数，用于查询构建器直接获取表名
// 这个函数可以被子模型重写来提供自定义表名
func (m *BaseModel) TableName() string {
	return m.config.TableName
}

// WithTableName 设置表名并返回模型（链式调用）
func (m *BaseModel) WithTableName(tableName string) *BaseModel {
	m.config.TableName = tableName
	return m
}

// SetTableName 实现 DynamicTableNamer 接口
func (m *BaseModel) SetTableName(name string) {
	m.config.TableName = name
}

// SetPrimaryKey 设置主键
func (m *BaseModel) SetPrimaryKey(key string) *BaseModel {
	m.config.PrimaryKey = key
	return m
}

// GetPrimaryKey 获取主键
func (m *BaseModel) GetPrimaryKey() string {
	return m.config.PrimaryKey
}

// PrimaryKey 获取主键名称（实现 Model 接口）
func (m *BaseModel) PrimaryKey() string {
	return m.config.PrimaryKey
}

// SetConnection 设置连接
func (m *BaseModel) SetConnection(connection string) *BaseModel {
	m.config.Connection = connection
	return m
}

// GetConnection 获取连接
func (m *BaseModel) GetConnection() string {
	return m.config.Connection
}

// Connection 获取连接名称（实现 Model 接口）
func (m *BaseModel) Connection() string {
	return m.config.Connection
}

// EnableTimestamps 启用时间戳
func (m *BaseModel) EnableTimestamps() Model {
	m.config.Timestamps = true
	return m
}

// DisableTimestamps 禁用时间戳
func (m *BaseModel) DisableTimestamps() Model {
	m.config.Timestamps = false
	return m
}

// SetCreatedAtField 设置创建时间字段
func (m *BaseModel) SetCreatedAtField(field string) *BaseModel {
	m.config.CreatedAtCol = field
	return m
}

// SetUpdatedAtField 设置更新时间字段
func (m *BaseModel) SetUpdatedAtField(field string) *BaseModel {
	m.config.UpdatedAtCol = field
	return m
}

// GetCreatedAtField 获取创建时间字段
func (m *BaseModel) GetCreatedAtField() string {
	return m.config.CreatedAtCol
}

// GetUpdatedAtField 获取更新时间字段
func (m *BaseModel) GetUpdatedAtField() string {
	return m.config.UpdatedAtCol
}

// EnableSoftDeletes 启用软删除
func (m *BaseModel) EnableSoftDeletes() *BaseModel {
	m.config.SoftDeletes = true
	return m
}

// DisableSoftDeletes 禁用软删除
func (m *BaseModel) DisableSoftDeletes() *BaseModel {
	m.config.SoftDeletes = false
	return m
}

// SetDeletedAtField 设置软删除字段
func (m *BaseModel) SetDeletedAtField(field string) *BaseModel {
	m.config.DeletedAtCol = field
	return m
}

// ============================================================================
// 查询方法 - 通过组合方式提供，完全兼容参数式查询
// ============================================================================

// RawQuery 创建查询构建器（不自动添加软删除过滤条件）
// 用于需要绕过软删除过滤的场景，如 WithTrashed、OnlyTrashed、Restore 等
func (m *BaseModel) RawQuery() (*db.QueryBuilder, error) {
	if m.config.TableName == "" {
		return nil, fmt.Errorf("表名未设置，请使用 SetTable() 方法设置表名")
	}

	query, err := db.NewQueryBuilder(m.config.Connection)
	if err != nil {
		return nil, fmt.Errorf("创建查询构建器失败: %w", err)
	}

	// 绑定模型实例以支持访问器处理
	query = query.From(m.config.TableName).WithModel(m)

	return query, nil
}

// Query 创建查询构建器
// 如果启用了软删除，会自动添加 deleted_at IS NULL 条件
func (m *BaseModel) Query() (*db.QueryBuilder, error) {
	query, err := m.RawQuery()
	if err != nil {
		return nil, err
	}

	// 如果启用了软删除，自动添加 deleted_at IS NULL 条件
	if m.config.SoftDeletes {
		query = query.WhereNull(m.config.DeletedAtCol)
	}

	return query, nil
}

// Where 支持多种参数式查询格式，返回QueryBuilder以支持链式调用
// 支持格式：
// - Where("name", "=", "John")           // 字段, 操作符, 值
// - Where("name = ?", "John")            // SQL + 参数
// - Where("id IN (?)", []int{1,2,3})     // SQL + 数组参数
// - Where("name = 'John'")               // 纯SQL
func (m *BaseModel) Where(args ...interface{}) (*db.QueryBuilder, error) {
	query, err := m.Query()
	if err != nil {
		return nil, err
	}
	return query.Where(args...), nil
}

// OrWhere 支持OR条件查询
func (m *BaseModel) OrWhere(args ...interface{}) (*db.QueryBuilder, error) {
	query, err := m.Query()
	if err != nil {
		return nil, err
	}
	return query.OrWhere(args...), nil
}

// WhereIn 字段值在数组中
func (m *BaseModel) WhereIn(field string, values []interface{}) (*db.QueryBuilder, error) {
	query, err := m.Query()
	if err != nil {
		return nil, err
	}
	return query.WhereIn(field, values), nil
}

// WhereNotIn 字段值不在数组中
func (m *BaseModel) WhereNotIn(field string, values []interface{}) (*db.QueryBuilder, error) {
	query, err := m.Query()
	if err != nil {
		return nil, err
	}
	return query.WhereNotIn(field, values), nil
}

// WhereBetween 字段值在范围内
func (m *BaseModel) WhereBetween(field string, values []interface{}) (*db.QueryBuilder, error) {
	query, err := m.Query()
	if err != nil {
		return nil, err
	}
	return query.WhereBetween(field, values), nil
}

// WhereNotBetween 字段值不在范围内
func (m *BaseModel) WhereNotBetween(field string, values []interface{}) (*db.QueryBuilder, error) {
	query, err := m.Query()
	if err != nil {
		return nil, err
	}
	return query.WhereNotBetween(field, values), nil
}

// WhereNull 字段为NULL
func (m *BaseModel) WhereNull(field string) (*db.QueryBuilder, error) {
	query, err := m.Query()
	if err != nil {
		return nil, err
	}
	return query.WhereNull(field), nil
}

// WhereNotNull 字段不为NULL
func (m *BaseModel) WhereNotNull(field string) (*db.QueryBuilder, error) {
	query, err := m.Query()
	if err != nil {
		return nil, err
	}
	return query.WhereNotNull(field), nil
}

// WhereRaw 原始SQL条件
func (m *BaseModel) WhereRaw(raw string, bindings ...interface{}) (*db.QueryBuilder, error) {
	query, err := m.Query()
	if err != nil {
		return nil, err
	}
	return query.WhereRaw(raw, bindings...), nil
}

// Select 选择字段
func (m *BaseModel) Select(fields ...interface{}) (*db.QueryBuilder, error) {
	query, err := m.Query()
	if err != nil {
		return nil, err
	}
	return query.Select(fields...), nil
}

// OrderBy 排序
func (m *BaseModel) OrderBy(column, direction string) (*db.QueryBuilder, error) {
	query, err := m.Query()
	if err != nil {
		return nil, err
	}
	return query.OrderBy(column, direction), nil
}

// GroupBy 分组
func (m *BaseModel) GroupBy(columns ...string) (*db.QueryBuilder, error) {
	query, err := m.Query()
	if err != nil {
		return nil, err
	}
	return query.GroupBy(columns...), nil
}

// Having Having条件
func (m *BaseModel) Having(args ...interface{}) (*db.QueryBuilder, error) {
	query, err := m.Query()
	if err != nil {
		return nil, err
	}
	return query.Having(args...), nil
}

// Limit 限制记录数
func (m *BaseModel) Limit(limit int) (*db.QueryBuilder, error) {
	query, err := m.Query()
	if err != nil {
		return nil, err
	}
	return query.Limit(limit), nil
}

// Offset 偏移量
func (m *BaseModel) Offset(offset int) (*db.QueryBuilder, error) {
	query, err := m.Query()
	if err != nil {
		return nil, err
	}
	return query.Offset(offset), nil
}

// Page 分页
func (m *BaseModel) Page(page, size int) (*db.QueryBuilder, error) {
	query, err := m.Query()
	if err != nil {
		return nil, err
	}
	return query.Page(page, size), nil
}

// Join 内连接
func (m *BaseModel) Join(table, first, operator, second string) (*db.QueryBuilder, error) {
	query, err := m.Query()
	if err != nil {
		return nil, err
	}
	return query.Join(table, first, operator, second), nil
}

// LeftJoin 左连接
func (m *BaseModel) LeftJoin(table, first, operator, second string) (*db.QueryBuilder, error) {
	query, err := m.Query()
	if err != nil {
		return nil, err
	}
	return query.LeftJoin(table, first, operator, second), nil
}

// RightJoin 右连接
func (m *BaseModel) RightJoin(table, first, operator, second string) (*db.QueryBuilder, error) {
	query, err := m.Query()
	if err != nil {
		return nil, err
	}
	return query.RightJoin(table, first, operator, second), nil
}

// ============================================================================
// 数据操作方法
// ============================================================================

// SetAttribute 设置属性
func (m *BaseModel) SetAttribute(key string, value interface{}) Model {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.attributes[key] = value
	return m
}

// GetAttribute 获取属性
func (m *BaseModel) GetAttribute(key string) interface{} {
	return m.attributes[key]
}

// SetAttributes 批量设置属性
func (m *BaseModel) SetAttributes(attributes map[string]interface{}) Model {
	m.mu.Lock()
	defer m.mu.Unlock()
	for key, value := range attributes {
		m.attributes[key] = value
	}
	return m
}

// GetAttributes 获取所有属性
func (m *BaseModel) GetAttributes() map[string]interface{} {
	return m.attributes
}

// ClearAttributes 清空属性
func (m *BaseModel) ClearAttributes() *BaseModel {
	m.attributes = make(map[string]interface{})
	return m
}

// Fill 填充模型属性
func (m *BaseModel) Fill(data map[string]interface{}) *BaseModel {
	m.mu.Lock()
	defer m.mu.Unlock()
	for key, value := range data {
		m.attributes[key] = value
	}
	return m
}

// FillAndSync 填充模型属性并同步原始数据（用于从数据库加载）
func (m *BaseModel) FillAndSync(data map[string]interface{}) *BaseModel {
	m.mu.Lock()
	defer m.mu.Unlock()
	for key, value := range data {
		m.attributes[key] = value
		m.original[key] = value
	}
	return m
}

// ============================================================================
// 状态管理方法
// ============================================================================

// IsNew 检查是否是新记录
func (m *BaseModel) IsNew() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return !m.exists
}

// IsDirty 检查模型是否有脏数据（已修改但未保存）
func (m *BaseModel) IsDirty() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.getDirtyInternal()) > 0
}

// GetDirty 获取脏数据（已修改的字段）
func (m *BaseModel) GetDirty() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.getDirtyInternal()
}

// getDirtyInternal 内部方法获取脏数据（不加锁）
func (m *BaseModel) getDirtyInternal() map[string]interface{} {
	dirty := make(map[string]interface{})
	for key, value := range m.attributes {
		if key == m.config.PrimaryKey {
			continue
		}
		originalValue, exists := m.original[key]
		if !exists || !reflect.DeepEqual(value, originalValue) {
			dirty[key] = value
		}
	}
	return dirty
}

// IsDirtyField 检查指定字段是否有脏数据
func (m *BaseModel) IsDirtyField(field string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exists := m.attributes[field]
	if !exists {
		return false
	}
	originalValue, originalExists := m.original[field]
	if !originalExists {
		return true
	}
	return !reflect.DeepEqual(value, originalValue)
}

// SyncOriginal 同步原始数据（保存后调用）
func (m *BaseModel) SyncOriginal() *BaseModel {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.original = make(map[string]interface{})
	for key, value := range m.attributes {
		m.original[key] = value
	}
	return m
}

// GetOriginal 获取原始数据
func (m *BaseModel) GetOriginal(key ...string) interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(key) > 0 {
		return m.original[key[0]]
	}
	result := make(map[string]interface{})
	for k, v := range m.original {
		result[k] = v
	}
	return result
}

// IsExists 检查记录是否存在
func (m *BaseModel) IsExists() bool {
	return m.exists
}

// MarkAsExists 标记为已存在
func (m *BaseModel) MarkAsExists() *BaseModel {
	m.exists = true
	return m
}

// MarkAsNew 标记为新记录
func (m *BaseModel) MarkAsNew() *BaseModel {
	m.exists = false
	return m
}

// GetKey 获取主键值
func (m *BaseModel) GetKey() interface{} {
	return m.GetAttribute(m.config.PrimaryKey)
}

// SetKey 设置主键值
func (m *BaseModel) SetKey(key interface{}) *BaseModel {
	m.SetAttribute(m.config.PrimaryKey, key)
	return m
}

// ============================================================================
// 查询执行方法 - 直接在BaseModel上执行查询
// ============================================================================

// Get 执行查询并返回数据（应用访问器处理）
func (m *BaseModel) Get() ([]map[string]interface{}, error) {
	query, err := m.Query()
	if err != nil {
		return nil, err
	}
	return query.Get()
}

// GetRaw 执行查询并返回原始数据（不应用访问器处理）
func (m *BaseModel) GetRaw() ([]map[string]interface{}, error) {
	query, err := m.Query()
	if err != nil {
		return nil, err
	}
	return query.GetRaw()
}

// First 获取第一条记录（应用访问器处理）
func (m *BaseModel) First() (map[string]interface{}, error) {
	query, err := m.Query()
	if err != nil {
		return nil, err
	}
	return query.First()
}

// FirstRaw 获取第一条记录的原始数据（不应用访问器处理）
func (m *BaseModel) FirstRaw() (map[string]interface{}, error) {
	query, err := m.Query()
	if err != nil {
		return nil, err
	}
	return query.FirstRaw()
}

// Count 计算记录数量
func (m *BaseModel) Count() (int64, error) {
	query, err := m.Query()
	if err != nil {
		return 0, err
	}
	return query.Count()
}

// ============================================================================
// 访问器支持方法
// ============================================================================

// SetAttributeWithAccessor 设置属性值并应用设置器
func (m *BaseModel) SetAttributeWithAccessor(model interface{}, key string, value interface{}) *BaseModel {
	processor := db.NewAccessorProcessor(model)
	processedData := processor.ProcessSetData(map[string]interface{}{key: value})
	if processedValue, exists := processedData[key]; exists {
		m.SetAttribute(key, processedValue)
	} else {
		m.SetAttribute(key, value)
	}
	return m
}

// SetAttributesWithAccessor 批量设置属性值并应用设置器
func (m *BaseModel) SetAttributesWithAccessor(model interface{}, data map[string]interface{}) *BaseModel {
	processor := db.NewAccessorProcessor(model)
	processedData := processor.ProcessSetData(data)
	m.SetAttributes(processedData)
	return m
}

// ============================================================================
// 持久化操作方法
// ============================================================================

// Save 保存模型
func (m *BaseModel) Save() error {
	// 调用保存前钩子
	if err := m.BeforeSave(); err != nil {
		return err
	}

	query, err := m.Query()
	if err != nil {
		return err
	}

	isNew := m.IsNew()

	if isNew {
		// 调用创建前钩子
		if err := m.BeforeCreate(); err != nil {
			return err
		}

		// 插入新记录
		data := m.prepareForInsert()
		if len(data) == 0 {
			return fmt.Errorf("没有要插入的数据")
		}

		id, err := query.Insert(data)
		if err != nil {
			return fmt.Errorf("模型插入失败: %w", err)
		}

		// 设置主键值（如果是自增的）
		if id > 0 {
			m.SetAttribute(m.config.PrimaryKey, id)
		}

		m.MarkAsExists()
		m.SyncOriginal()

		// 调用创建后钩子
		if err := m.AfterCreate(); err != nil {
			return err
		}
	} else {
		// 调用更新前钩子
		if err := m.BeforeUpdate(); err != nil {
			return err
		}

		// 只更新脏数据
		data := m.prepareForUpdate()
		if len(data) == 0 {
			return nil // 没有需要更新的数据
		}

		pk := m.GetKey()
		if pk == nil {
			return fmt.Errorf("主键值不能为空")
		}

		affected, err := query.Where(m.config.PrimaryKey, "=", pk).Update(data)
		if err != nil {
			return fmt.Errorf("模型更新失败: %w", err)
		}

		if affected == 0 {
			return fmt.Errorf("没有找到要更新的记录")
		}

		m.SyncOriginal()

		// 调用更新后钩子
		if err := m.AfterUpdate(); err != nil {
			return err
		}
	}

	// 调用保存后钩子
	if err := m.AfterSave(); err != nil {
		return err
	}

	return nil
}

// FindByPK 根据主键查找
func (m *BaseModel) FindByPK(key interface{}) error {
	if key == nil {
		return fmt.Errorf("主键值不能为空")
	}

	query, err := m.Query()
	if err != nil {
		return err
	}

	result, err := query.Where(m.config.PrimaryKey, "=", key).FirstRaw()
	if err != nil {
		return fmt.Errorf("查找模型失败: %w", err)
	}

	m.FillAndSync(result)
	m.MarkAsExists()
	return nil
}

// Reload 重新加载模型数据
func (m *BaseModel) Reload() error {
	pk := m.GetKey()
	if pk == nil {
		return fmt.Errorf("主键值不能为空")
	}
	return m.FindByPK(pk)
}

// Find 根据主键查找记录（别名方法）
func (m *BaseModel) Find(pk interface{}) error {
	return m.FindByPK(pk)
}

// Delete 删除记录
func (m *BaseModel) Delete() error {
	// 调用删除前钩子
	if err := m.BeforeDelete(); err != nil {
		return err
	}

	var err error
	if m.config.SoftDeletes {
		err = m.SoftDelete()
	} else {
		err = m.ForceDelete()
	}

	if err != nil {
		return err
	}

	// 调用删除后钩子
	return m.AfterDelete()
}

// SoftDelete 软删除（如果启用）
func (m *BaseModel) SoftDelete() error {
	if !m.config.SoftDeletes {
		return fmt.Errorf("该模型未启用软删除")
	}

	query, err := m.Query()
	if err != nil {
		return err
	}

	pk := m.GetKey()
	if pk == nil {
		return fmt.Errorf("主键值不能为空")
	}

	now := time.Now()
	data := map[string]interface{}{
		m.config.DeletedAtCol: now,
	}

	affected, err := query.Where(m.config.PrimaryKey, "=", pk).Update(data)
	if err != nil {
		return fmt.Errorf("软删除失败: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("没有找到要删除的记录")
	}

	// 更新模型属性
	m.SetAttribute(m.config.DeletedAtCol, now)

	return nil
}

// IsSoftDeleted 检查是否已软删除
func (m *BaseModel) IsSoftDeleted() bool {
	if !m.config.SoftDeletes {
		return false
	}
	deletedAt := m.GetAttribute(m.config.DeletedAtCol)
	if deletedAt == nil {
		return false
	}
	// 检查是否为有效的时间值
	switch v := deletedAt.(type) {
	case time.Time:
		return !v.IsZero()
	case *time.Time:
		return v != nil && !v.IsZero()
	case DeletedAt:
		return v.Valid
	case *DeletedAt:
		return v != nil && v.Valid
	default:
		return deletedAt != nil
	}
}

// GetDeletedAtField 获取软删除字段名
func (m *BaseModel) GetDeletedAtField() string {
	return m.config.DeletedAtCol
}

// Restore 恢复软删除的记录
func (m *BaseModel) Restore() error {
	if !m.config.SoftDeletes {
		return fmt.Errorf("该模型未启用软删除")
	}

	// 使用 RawQuery 绕过软删除过滤，否则无法找到已软删除的记录
	query, err := m.RawQuery()
	if err != nil {
		return err
	}

	pk := m.GetKey()
	if pk == nil {
		return fmt.Errorf("主键值不能为空")
	}

	data := map[string]interface{}{
		m.config.DeletedAtCol: nil,
	}

	_, err = query.Where(m.config.PrimaryKey, "=", pk).Update(data)
	return err
}

// ForceDelete 强制删除（真实删除）
func (m *BaseModel) ForceDelete() error {
	// 使用 RawQuery 绕过软删除过滤，确保可以删除已软删除的记录
	query, err := m.RawQuery()
	if err != nil {
		return err
	}

	pk := m.GetKey()
	if pk == nil {
		return fmt.Errorf("主键值不能为空")
	}

	affected, err := query.Where(m.config.PrimaryKey, "=", pk).Delete()
	if err != nil {
		return fmt.Errorf("强制删除失败: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("没有找到要删除的记录")
	}

	m.MarkAsNew()
	return nil
}

// ============================================================================
// 事件钩子方法
// ============================================================================

// BeforeSave 保存前钩子
func (m *BaseModel) BeforeSave() error {
	return nil
}

// AfterSave 保存后钩子
func (m *BaseModel) AfterSave() error {
	return nil
}

// BeforeCreate 创建前钩子
func (m *BaseModel) BeforeCreate() error {
	return nil
}

// AfterCreate 创建后钩子
func (m *BaseModel) AfterCreate() error {
	return nil
}

// BeforeUpdate 更新前钩子
func (m *BaseModel) BeforeUpdate() error {
	return nil
}

// AfterUpdate 更新后钩子
func (m *BaseModel) AfterUpdate() error {
	return nil
}

// BeforeDelete 删除前钩子
func (m *BaseModel) BeforeDelete() error {
	return nil
}

// AfterDelete 删除后钩子
func (m *BaseModel) AfterDelete() error {
	return nil
}

// ============================================================================
// 序列化方法
// ============================================================================

// ToMap 转换为map
func (m *BaseModel) ToMap() map[string]interface{} {
	return m.GetAttributes()
}

// ToJSON 转换为JSON字符串
func (m *BaseModel) ToJSON() (string, error) {
	data := m.GetAttributes()
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// FromJSON 从JSON字符串填充属性
func (m *BaseModel) FromJSON(jsonStr string) error {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return err
	}

	m.SetAttributes(data)
	return nil
}

// ============================================================================
// 关联方法
// ============================================================================

// HasOne 一对一关联
func (m *BaseModel) HasOne(modelType interface{}, foreignKey, localKey string) *HasOne {
	relatedType := getReflectType(modelType)
	relatedTable := getTableNameFromModel(modelType)
	return NewHasOneWithTable(m, relatedType, relatedTable, foreignKey, localKey)
}

// HasMany 一对多关联
func (m *BaseModel) HasMany(modelType interface{}, foreignKey, localKey string) *HasMany {
	relatedType := getReflectType(modelType)
	relatedTable := getTableNameFromModel(modelType)
	return NewHasManyWithTable(m, relatedType, relatedTable, foreignKey, localKey)
}

// BelongsTo 反向关联（多对一/一对一）
func (m *BaseModel) BelongsTo(modelType interface{}, foreignKey, localKey string) *BelongsTo {
	relatedType := getReflectType(modelType)
	relatedTable := getTableNameFromModel(modelType)
	return NewBelongsToWithTable(m, relatedType, relatedTable, foreignKey, localKey)
}

// BelongsToMany 多对多关联
func (m *BaseModel) BelongsToMany(modelType interface{}, pivotTable, foreignKey, localKey string) *BelongsToMany {
	relatedType := getReflectType(modelType)
	relatedTable := getTableNameFromModel(modelType)
	return NewBelongsToManyWithTable(m, relatedType, relatedTable, pivotTable, foreignKey, localKey)
}

// ============================================================================
// 预加载方法
// ============================================================================

// With 预加载关联数据
// 支持多个关联名称，也支持嵌套关联（如 "posts.comments"）
func (m *BaseModel) With(relations ...string) *ModelQueryBuilder {
	mqb := NewModelQueryBuilder(m)
	return mqb.With(relations...)
}

// WithConstraint 带约束的预加载
func (m *BaseModel) WithConstraint(relation string, constraint func(*db.QueryBuilder) *db.QueryBuilder) *ModelQueryBuilder {
	mqb := NewModelQueryBuilder(m)
	return mqb.WithConstraint(relation, constraint)
}

// Load 延迟加载关联数据到当前模型
func (m *BaseModel) Load(relations ...string) error {
	if !m.IsExists() {
		return fmt.Errorf("模型不存在，无法加载关联")
	}

	// 获取当前模型数据
	data := m.GetAttributes()
	results := []map[string]interface{}{data}

	// 加载关联
	loadedResults, err := LoadRelations(m, results, relations)
	if err != nil {
		return err
	}

	// 更新模型属性
	if len(loadedResults) > 0 {
		m.SetAttributes(loadedResults[0])
	}

	return nil
}

// GetRelation 获取已加载的关联数据
func (m *BaseModel) GetRelation(name string) interface{} {
	return m.GetAttribute(name)
}

// SetRelation 设置关联数据
func (m *BaseModel) SetRelation(name string, data interface{}) *BaseModel {
	m.SetAttribute(name, data)
	return m
}

// GetRelationOrLoad 获取关联数据，如果未加载则自动加载（延迟加载）
func (m *BaseModel) GetRelationOrLoad(name string) (interface{}, error) {
	// 检查是否已加载
	if data := m.GetAttribute(name); data != nil {
		return data, nil
	}

	// 尝试延迟加载
	if err := m.Load(name); err != nil {
		return nil, err
	}

	return m.GetAttribute(name), nil
}

// HasRelationLoaded 检查关联是否已加载
func (m *BaseModel) HasRelationLoaded(name string) bool {
	return m.GetAttribute(name) != nil
}

// ============================================================================
// 迁移方法
// ============================================================================

// AutoMigrate 自动迁移表结构
func (m *BaseModel) AutoMigrate(models ...interface{}) error {
	// 获取数据库管理器并创建连接
	manager := db.DefaultManager()
	conn, err := manager.Connection(m.config.Connection)
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 创建迁移器
	migrator := migration.NewAutoMigrator(conn)

	// 启用性能优化：表存在性缓存
	migrator.SetCacheEnabled(true)

	// 如果没有传入模型，使用当前模型
	if len(models) == 0 {
		models = []interface{}{m}
	}

	// 迁移每个模型
	for _, model := range models {
		var tableName string

		// 获取表名
		if baseModel, ok := model.(*BaseModel); ok {
			tableName = baseModel.GetTableName()
		} else {
			// 检查是否有嵌入的BaseModel
			modelValue := reflect.ValueOf(model)
			if modelValue.Kind() == reflect.Ptr {
				modelValue = modelValue.Elem()
			}

			if modelValue.Kind() == reflect.Struct {
				baseModelField := modelValue.FieldByName("BaseModel")
				if baseModelField.IsValid() && baseModelField.Type() == reflect.TypeOf(BaseModel{}) {
					if baseModelField.CanAddr() {
						baseModelPtr := baseModelField.Addr().Interface().(*BaseModel)
						tableName = baseModelPtr.GetTableName()
					}
				}
			}
		}

		// 如果还是没有表名，则使用类型推断
		if tableName == "" {
			modelType := reflect.TypeOf(model)
			if modelType.Kind() == reflect.Ptr {
				modelType = modelType.Elem()
			}
			tableName = toSnakeCase(modelType.Name())
		}

		err := migrator.MigrateModel(model, tableName)
		if err != nil {
			return fmt.Errorf("迁移模型 %s 失败: %w", tableName, err)
		}
	}

	return nil
}

// ============================================================================
// 内部辅助方法
// ============================================================================

// prepareForInsert 准备插入数据
func (m *BaseModel) prepareForInsert() map[string]interface{} {
	data := make(map[string]interface{})

	// 获取所有属性
	for key, value := range m.attributes {
		data[key] = value
	}

	// 处理时间戳字段
	if m.config.Timestamps {
		now := time.Now()
		data[m.config.CreatedAtCol] = now
		data[m.config.UpdatedAtCol] = now
	}

	// 处理时间字段管理
	if m.timeManager != nil && len(m.timeFields) > 0 {
		data = m.timeManager.ProcessInsertData(data, m.timeFields)
	}

	return data
}

// prepareForUpdate 准备更新数据（只更新脏数据）
func (m *BaseModel) prepareForUpdate() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 只获取脏数据
	data := m.getDirtyInternal()

	// 如果没有脏数据，返回空
	if len(data) == 0 {
		// 但如果启用了时间戳，仍然需要更新 updated_at
		if m.config.Timestamps {
			data = make(map[string]interface{})
			data[m.config.UpdatedAtCol] = time.Now()
		} else {
			return data
		}
	}

	// 处理时间戳字段
	if m.config.Timestamps {
		data[m.config.UpdatedAtCol] = time.Now()
	}

	// 处理时间字段管理
	if m.timeManager != nil && len(m.timeFields) > 0 {
		data = m.timeManager.ProcessUpdateData(data, m.timeFields)
	}

	return data
}

// parseTagsIntoConfig 解析标签到配置
func parseTagsIntoConfig(structInstance interface{}, config *ModelConfig) {
	modelType := reflect.TypeOf(structInstance)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 解析模型的标签
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)

		// 跳过嵌入的BaseModel字段
		if field.Type.Name() == "BaseModel" || field.Anonymous {
			continue
		}

		tormTag := field.Tag.Get("torm")
		if tormTag == "" {
			continue
		}

		// 解析torm标签
		parseFieldTag(field, tormTag, config)
	}
}

// parseFieldTag 解析单个字段的标签
// 支持两种格式：
// - torm格式：使用逗号分隔，如 `torm:"primary_key,auto_increment"`
// - GORM格式：使用分号分隔，如 `torm:"primaryKey;autoIncrement"`
func parseFieldTag(field reflect.StructField, tormTag string, config *ModelConfig) {
	// 检测分隔符：如果包含分号，使用GORM格式；否则使用torm格式
	var parts []string
	if strings.Contains(tormTag, ";") {
		parts = strings.Split(tormTag, ";")
	} else {
		parts = strings.Split(tormTag, ",")
	}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// 检查忽略标记
		if part == "-" {
			return // 忽略此字段
		}

		if strings.Contains(part, ":") {
			// 键值对形式: key:value
			parseTormKeyValue(part, field, config)
		} else {
			// 标志形式: primary_key, unique等
			parseTormFlagAdvanced(part, field, config)
		}
	}
}

// getColumnNameFromField 从字段获取列名
func getColumnNameFromField(field reflect.StructField) string {
	// 检查json标签作为列名
	if jsonTag := field.Tag.Get("json"); jsonTag != "" {
		parts := strings.Split(jsonTag, ",")
		if parts[0] != "" && parts[0] != "-" {
			return parts[0]
		}
	}

	// 使用字段名的蛇形命名
	return toSnakeCase(field.Name)
}

// parseTormKeyValue 解析torm标签的键值对 - 与migration包保持一致
// 支持GORM兼容的驼峰命名和torm的蛇形命名
func parseTormKeyValue(part string, field reflect.StructField, config *ModelConfig) {
	kv := strings.SplitN(part, ":", 2)
	if len(kv) != 2 {
		return
	}

	key := strings.ToLower(strings.TrimSpace(kv[0]))
	value := strings.TrimSpace(kv[1])

	switch key {
	// ============================================================================
	// 列名相关 - GORM: column, torm: column/db
	// ============================================================================
	case "column", "db":
		// 自定义列名：column:user_name, db:user_id
		// 这主要用于字段映射，在模型配置层面不需要特殊处理
		// 但我们可以在这里记录以供后续使用

	// ============================================================================
	// 数据类型相关 - GORM: type, torm: type
	// ============================================================================
	case "type":
		// 数据库类型：type:varchar, type:decimal等
		// 模型配置层面不处理具体的数据库类型，这由migration包处理

	// ============================================================================
	// 序列化器 - GORM: serializer
	// ============================================================================
	case "serializer":
		// 序列化器：serializer:json, serializer:gob等
		// 模型配置层面不处理，由migration包处理

	// ============================================================================
	// 大小相关 - GORM: size, torm: size
	// ============================================================================
	case "size":
		// 字段大小：size:50, size:100等
		// 模型配置层面不处理，由migration包处理

	// ============================================================================
	// 精度相关 - GORM: precision/scale, torm: precision/scale
	// ============================================================================
	case "precision":
		// 精度：precision:10 (用于decimal类型)
		// 模型配置层面不处理，由migration包处理

	case "scale":
		// 标度：scale:2 (用于decimal类型)
		// 模型配置层面不处理，由migration包处理

	// ============================================================================
	// 默认值 - GORM: default, torm: default
	// ============================================================================
	case "default":
		// 默认值：default:0, default:'active'等
		// 模型配置层面不处理，由migration包处理

	// ============================================================================
	// 注释 - GORM: comment, torm: comment
	// ============================================================================
	case "comment":
		// 字段注释：comment:'用户名'
		// 模型配置层面不处理，由migration包处理

	// ============================================================================
	// 自增步长 - GORM: autoIncrementIncrement
	// ============================================================================
	case "autoincrementincrement", "auto_increment_increment":
		// 自增步长：autoIncrementIncrement:5
		// 模型配置层面不处理，由migration包处理

	// ============================================================================
	// 嵌入前缀 - GORM: embeddedPrefix
	// ============================================================================
	case "embeddedprefix", "embedded_prefix":
		// 嵌入前缀：embeddedPrefix:author_
		// 模型配置层面不处理，由migration包处理

	// ============================================================================
	// 时间戳精度 - GORM: autoCreateTime:nano/milli, autoUpdateTime:nano/milli
	// ============================================================================
	case "autocreatetime", "auto_create_time":
		// 自动创建时间：autoCreateTime:nano, autoCreateTime:milli
		columnName := getColumnNameFromField(field)
		config.CreatedAtCol = columnName

	case "autoupdatetime", "auto_update_time":
		// 自动更新时间：autoUpdateTime:nano, autoUpdateTime:milli
		columnName := getColumnNameFromField(field)
		config.UpdatedAtCol = columnName

	// ============================================================================
	// 索引相关 - GORM: index/uniqueIndex
	// ============================================================================
	case "index":
		// 带类型的索引：index:btree, index:hash
		// 或带名称的索引：index:idx_name
		// 模型配置层面不处理，由migration包处理

	case "uniqueindex", "unique_index":
		// 唯一索引：uniqueIndex:idx_name
		// 模型配置层面不处理，由migration包处理

	// ============================================================================
	// 检查约束 - GORM: check
	// ============================================================================
	case "check":
		// 检查约束：check:age > 0
		// 模型配置层面不处理，由migration包处理

	// ============================================================================
	// 外键相关 - torm扩展
	// ============================================================================
	case "foreign_key", "fk", "foreignkey":
		// 外键：foreign_key:users.id
		// 模型配置层面不处理，由migration包处理

	case "references", "ref":
		// 引用：references:roles(id)
		// 模型配置层面不处理，由migration包处理

	case "on_delete", "ondelete":
		// 删除时动作：on_delete:cascade
		// 模型配置层面不处理，由migration包处理

	case "on_update", "onupdate":
		// 更新时动作：on_update:cascade
		// 模型配置层面不处理，由migration包处理

	// ============================================================================
	// 生成列 - torm扩展
	// ============================================================================
	case "generated":
		// 生成列：generated:virtual, generated:stored
		// 模型配置层面不处理，由migration包处理

	// ============================================================================
	// 字段权限 - GORM: <-:create/<-:update/<-:false/->:false
	// ============================================================================
	case "<-":
		// 写入权限：<-:create, <-:update, <-:false
		// 由accessor.go处理
		_ = value

	case "->":
		// 读取权限：->:false
		// 由accessor.go处理
		_ = value
	}
}

// parseTormFlagAdvanced 解析torm标签的标志 - 扩展版本
// 支持GORM兼容的驼峰命名和torm的蛇形命名
func parseTormFlagAdvanced(flag string, field reflect.StructField, config *ModelConfig) {
	flagLower := strings.ToLower(flag)
	columnName := getColumnNameFromField(field)

	switch flagLower {
	// ============================================================================
	// 主键相关 - GORM: primaryKey, torm: primary_key
	// ============================================================================
	case "primary_key", "pk", "primary", "primarykey":
		config.PrimaryKey = columnName

	// ============================================================================
	// 自增相关 - GORM: autoIncrement, torm: auto_increment
	// ============================================================================
	case "auto_increment", "autoincrement", "auto_inc", "autoinc":
		// 自增字段 - 这通常与primary_key一起使用
		// 在模型配置中，我们主要关心主键设置
		// auto_increment的具体处理在迁移时由migration包负责

	// ============================================================================
	// 时间戳相关 - GORM: autoCreateTime/autoUpdateTime, torm: auto_create_time/auto_update_time
	// ============================================================================
	case "auto_create_time", "create_time", "created_at", "auto_created_at",
		"autocreate_time", "autocreatetime", "auto_create", "autocreate",
		"autocreatetime:nano", "autocreatetime:milli": // GORM支持纳秒/毫秒
		config.CreatedAtCol = columnName

	case "auto_update_time", "update_time", "updated_at", "auto_updated_at",
		"autoupdate_time", "autoupdatetime", "auto_update", "autoupdate",
		"autoupdatetime:nano", "autoupdatetime:milli": // GORM支持纳秒/毫秒
		config.UpdatedAtCol = columnName

	// ============================================================================
	// 软删除相关 - GORM: softDelete, torm: soft_delete
	// ============================================================================
	case "soft_delete", "soft_deletes", "deleted_at", "softdelete", "softdeletes":
		config.DeletedAtCol = columnName
		config.SoftDeletes = true

	// ============================================================================
	// 唯一约束 - GORM: unique, torm: unique
	// ============================================================================
	case "unique", "uniq", "uniqueindex":
		// 唯一索引 - 由migration包处理

	// ============================================================================
	// 非空约束 - GORM: not null, torm: not_null
	// ============================================================================
	case "not_null", "not null", "notnull", "not_nil", "notnil":
		// 非空约束 - 由migration包处理

	case "nullable", "null":
		// 可空约束 - 由migration包处理

	// ============================================================================
	// 嵌入结构体 - GORM: embedded/embeddedPrefix
	// ============================================================================
	case "embedded":
		// 嵌入结构体 - 由migration包处理

	// ============================================================================
	// 索引相关 - GORM: index/uniqueIndex
	// ============================================================================
	case "index", "idx":
		// 普通索引 - 由migration包处理

	case "fulltext", "fulltext_index":
		// 全文索引 - 由migration包处理

	case "spatial", "spatial_index":
		// 空间索引 - 由migration包处理

	// ============================================================================
	// 数据类型相关
	// ============================================================================
	case "timestamp", "current_timestamp":
		// 时间戳字段 - 由migration包处理

	case "unsigned":
		// 无符号数字类型 (主要用于MySQL) - 由migration包处理

	case "zerofill":
		// 零填充 (主要用于MySQL) - 由migration包处理

	case "binary":
		// 二进制存储 - 由migration包处理

	case "json", "is_json":
		// JSON字段标记 - 由migration包处理

	// ============================================================================
	// 字段权限相关 - GORM: <-:create/<-:update/<-:false/->:false/-
	// ============================================================================
	case "readonly", "read_only", "immutable":
		// 只读字段标记 - 由accessor.go处理

	case "writeonly", "write_only":
		// 只写字段标记 - 由accessor.go处理

	case "createonly", "create_only":
		// 只在创建时可写 - 由accessor.go处理

	case "updateonly", "update_only":
		// 只在更新时可写 - 由accessor.go处理

	case "hidden", "invisible":
		// 隐藏字段标记 - 由accessor.go处理

	case "encrypted", "encrypt":
		// 加密字段标记 - 可能需要在模型层处理，但目前不实现

	// ============================================================================
	// PostgreSQL序列相关
	// ============================================================================
	case "serial":
		// PostgreSQL SERIAL类型 - 等同于auto_increment

	case "bigserial":
		// PostgreSQL BIGSERIAL类型

	case "smallserial":
		// PostgreSQL SMALLSERIAL类型

	// ============================================================================
	// 外键级联操作标志
	// ============================================================================
	case "cascade_delete", "on_delete_cascade":
		// 级联删除 - 由migration包处理

	case "cascade_update", "on_update_cascade":
		// 级联更新 - 由migration包处理

	case "restrict_delete", "on_delete_restrict":
		// 限制删除 - 由migration包处理

	case "restrict_update", "on_update_restrict":
		// 限制更新 - 由migration包处理

	case "set_null_delete", "on_delete_set_null":
		// 删除时设为NULL - 由migration包处理

	case "set_null_update", "on_update_set_null":
		// 更新时设为NULL - 由migration包处理

	case "set_default_delete", "on_delete_set_default":
		// 删除时设为默认值 - 由migration包处理

	case "set_default_update", "on_update_set_default":
		// 更新时设为默认值 - 由migration包处理

	// ============================================================================
	// 生成列相关
	// ============================================================================
	case "generated":
		// 生成列（无值版本，默认虚拟列） - 由migration包处理

	case "virtual":
		// 虚拟生成列 - 由migration包处理

	case "stored":
		// 存储生成列 - 由migration包处理
	}
}

// toSnakeCase 转换为蛇形命名
func toSnakeCase(str string) string {
	if str == "" {
		return ""
	}

	var result strings.Builder
	runes := []rune(str)

	for i, r := range runes {
		// 当前字符是大写字母
		if r >= 'A' && r <= 'Z' {
			// 需要添加下划线的条件：
			// 1. 不是第一个字符
			// 2. 前一个字符是小写字母，或者
			// 3. 当前字符后面跟着小写字母（处理连续大写的情况，如 HTMLParser -> html_parser）
			if i > 0 && ((runes[i-1] >= 'a' && runes[i-1] <= 'z') || // 前一个是小写
				(i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z')) { // 后一个是小写
				result.WriteRune('_')
			}
			result.WriteRune(r - 'A' + 'a') // 转为小写
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// getReflectType 获取反射类型
func getReflectType(modelType interface{}) reflect.Type {
	if t, ok := modelType.(reflect.Type); ok {
		return t
	}
	return reflect.TypeOf(modelType)
}

// getTableNameFromModel 从模型实例获取表名
func getTableNameFromModel(modelType interface{}) string {
	// 如果是模型实例，直接获取表名
	if modeler, ok := modelType.(interface{ GetTableName() string }); ok {
		return modeler.GetTableName()
	}

	// 如果是类型，回退到类型名推断
	reflectType := getReflectType(modelType)
	if reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}
	return toSnakeCase(reflectType.Name())
}
