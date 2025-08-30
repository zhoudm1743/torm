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

// 性能优化：全局BaseModel对象池
var (
	baseModelPool = sync.Pool{
		New: func() interface{} {
			return &BaseModel{
				attributes:  make(map[string]interface{}),
				timeFields:  make([]db.TimeFieldInfo, 0, 4), // 预分配4个元素
				timeManager: db.NewTimeFieldManager(),
			}
		},
	}

	// 时间字段缓存，避免重复分析
	timeFieldsCache      = make(map[reflect.Type][]db.TimeFieldInfo)
	timeFieldsCacheMutex sync.RWMutex
)

// BaseModel 基础模型 - 性能优化版本
type BaseModel struct {
	*db.QueryBuilder // 直接嵌入查询构建器，自动继承所有查询方法

	// 基础配置
	tableName  string
	primaryKey string
	connection string
	attributes map[string]interface{}

	// 状态
	exists bool

	// 时间戳
	timestamps  bool
	createdAt   string
	updatedAt   string
	timeManager *db.TimeFieldManager // 时间字段管理器
	timeFields  []db.TimeFieldInfo   // 时间字段信息缓存

	// 软删除
	softDeletes bool
	deletedAt   string

	// 性能优化：缓存访问器处理器
	accessorProcessor *db.AccessorProcessor
	accessorMutex     sync.RWMutex
}

// NewModel 创建模型 - 支持多种调用方式
// NewModel() - 创建空模型，需要手动设置表名
// NewModel(tableName) - 创建指定表名的模型
// NewModel(tableName, connection) - 创建指定表名和连接的模型
// NewModel(structInstance) - 从结构体创建模型并解析标签
func NewModel(args ...interface{}) *BaseModel {
	model := &BaseModel{
		connection:  "default",
		primaryKey:  "id",
		attributes:  make(map[string]interface{}),
		exists:      false,
		timestamps:  true,
		createdAt:   "created_at",
		updatedAt:   "updated_at",
		softDeletes: false,
		deletedAt:   "deleted_at",
		timeManager: db.NewTimeFieldManager(),
		timeFields:  make([]db.TimeFieldInfo, 0),
	}

	var tableName string
	var connection string = "default"
	var structInstance interface{}

	// 解析参数
	switch len(args) {
	case 0:
		// 无参数，创建空模型
	case 1:
		// 一个参数，可能是表名或结构体实例
		if str, ok := args[0].(string); ok {
			tableName = str
		} else {
			// 是结构体实例，解析标签
			structInstance = args[0]
		}
	case 2:
		// 两个参数，表名和连接
		if str1, ok := args[0].(string); ok {
			tableName = str1
		}
		if str2, ok := args[1].(string); ok {
			connection = str2
		}
	}

	model.connection = connection

	// 如果有结构体实例，解析标签
	if structInstance != nil {
		// 首先尝试从实例获取已设置的表名
		if modeler, ok := structInstance.(interface{ GetTableName() string }); ok {
			existingTableName := modeler.GetTableName()
			if existingTableName != "" {
				model.tableName = existingTableName
				tableName = existingTableName
			}
		}

		// 解析torm标签
		model.ParseModelTags(structInstance)

		// 分析时间字段
		model.timeFields = model.timeManager.AnalyzeModelTimeFields(structInstance)

		// 如果还是没有表名，从结构体名称推断（作为最后手段）
		if model.tableName == "" {
			modelType := reflect.TypeOf(structInstance)
			if modelType.Kind() == reflect.Ptr {
				modelType = modelType.Elem()
			}
			model.tableName = toSnakeCase(modelType.Name())
		}
		tableName = model.tableName
	} else if tableName != "" {
		model.tableName = tableName
	}

	// 初始化查询构建器
	query, err := db.NewQueryBuilder(connection)
	if err != nil {
		emptyQuery, _ := db.NewQueryBuilder("default")
		model.QueryBuilder = emptyQuery
		return model
	}

	if tableName != "" {
		model.QueryBuilder = query.From(tableName)
	} else {
		model.QueryBuilder = query
	}

	return model
}

// ============================================================================
// 模型特有方法（不与 QueryInterface 冲突的）
// ============================================================================

// 配置方法

// SetTable 设置表名
func (m *BaseModel) SetTable(tableName string) *BaseModel {
	m.tableName = tableName
	// 同步更新 QueryBuilder 的表名
	if m.QueryBuilder != nil {
		m.QueryBuilder = m.QueryBuilder.From(tableName)
	}
	return m
}

// GetTableName 获取表名
func (m *BaseModel) GetTableName() string {
	// 绝对优先返回显式设置的表名，不允许被覆盖
	if m.tableName != "" {
		return m.tableName
	}

	// 如果没有设置表名，返回空字符串强制用户设置
	// 这避免了意外的反射推断导致的表名不一致问题
	return ""
}

// SetPrimaryKey 设置主键
func (m *BaseModel) SetPrimaryKey(key string) *BaseModel {
	m.primaryKey = key
	return m
}

// GetPrimaryKey 获取主键
func (m *BaseModel) GetPrimaryKey() string {
	return m.primaryKey
}

// SetConnection 设置连接
func (m *BaseModel) SetConnection(connection string) *BaseModel {
	m.connection = connection
	// 同时更新嵌入的QueryBuilder的连接
	if m.QueryBuilder != nil {
		// 重新创建QueryBuilder以使用新连接
		query, err := db.NewQueryBuilder(connection)
		if err == nil && m.GetTableName() != "" {
			m.QueryBuilder = query.From(m.GetTableName())
		}
	}
	return m
}

// GetConnection 获取连接
func (m *BaseModel) GetConnection() string {
	return m.connection
}

// Model 创建新的查询实例（重置查询状态并使用正确的连接） - 优化版本
// 这个方法解决了嵌入QueryBuilder可能使用错误连接的问题
func (m *BaseModel) Model(model interface{}) *BaseModel {
	// 使用对象池获取BaseModel实例
	newModel := baseModelPool.Get().(*BaseModel)
	
	// 重置实例状态
	newModel.tableName = m.tableName
	newModel.primaryKey = m.primaryKey
	newModel.connection = m.connection
	newModel.exists = false
	newModel.timestamps = m.timestamps
	newModel.createdAt = m.createdAt
	newModel.updatedAt = m.updatedAt
	newModel.softDeletes = m.softDeletes
	newModel.deletedAt = m.deletedAt
	newModel.timeManager = m.timeManager
	newModel.timeFields = m.timeFields
	
	// 清空属性映射
	for k := range newModel.attributes {
		delete(newModel.attributes, k)
	}

	// 如果传入了模型实例，解析其配置
	if model != nil {
		if modeler, ok := model.(interface{ GetTableName() string }); ok {
			tableName := modeler.GetTableName()
			if tableName != "" {
				newModel.tableName = tableName
			}
		}
		if modeler, ok := model.(interface{ GetConnection() string }); ok {
			connection := modeler.GetConnection()
			if connection != "" {
				newModel.connection = connection
			}
		}
	}

	// 复用QueryBuilder或创建新的
	if m.QueryBuilder != nil && m.QueryBuilder.GetConnection() == newModel.connection {
		// 复用现有QueryBuilder，但重置查询状态
		newModel.QueryBuilder = m.QueryBuilder.Reset()
	} else {
		// 创建新的QueryBuilder实例
		query, err := db.NewQueryBuilder(newModel.connection)
		if err != nil {
			// 如果创建失败，尝试使用默认连接
			query, _ = db.NewQueryBuilder("default")
		}
		
		if newModel.GetTableName() != "" {
			newModel.QueryBuilder = query.From(newModel.GetTableName())
		} else {
			newModel.QueryBuilder = query
		}
	}

	return newModel
}

// NewQuery 创建新的查询实例（简化版本，不需要参数）
// 这是Model()方法的简化版，用于重置查询状态
func (m *BaseModel) NewQuery() *BaseModel {
	return m.Model(nil)
}

// Release 释放模型实例到对象池（优化内存使用）
func (m *BaseModel) Release() {
	if m == nil {
		return
	}
	
	// 释放QueryBuilder资源
	if m.QueryBuilder != nil {
		m.QueryBuilder.Release()
		m.QueryBuilder = nil
	}
	
	// 清空属性映射
	for k := range m.attributes {
		delete(m.attributes, k)
	}
	
	// 重置字段
	m.tableName = ""
	m.primaryKey = "id"
	m.connection = "default"
	m.exists = false
	m.timestamps = true
	m.createdAt = "created_at"
	m.updatedAt = "updated_at"
	m.softDeletes = false
	m.deletedAt = "deleted_at"
	
	// 重置时间字段缓存
	m.timeFields = m.timeFields[:0]
	
	// 释放访问器处理器
	m.accessorProcessor = nil
	
	// 放回对象池
	baseModelPool.Put(m)
}

// ============================================================================
// 重写查询方法，自动使用正确的连接（用户透明）
// ============================================================================

// Where 条件查询（重写以确保使用正确连接）
func (m *BaseModel) Where(args ...interface{}) *BaseModel {
	// 如果已经有QueryBuilder，直接使用它继续链式调用
	if m.QueryBuilder != nil {
		return m.wrapQueryBuilder(m.QueryBuilder.Where(args...))
	}

	// 否则创建新查询
	newQuery := m.newQuerySafe()
	if newQuery == nil {
		return m // fallback 到原始查询
	}
	return m.wrapQueryBuilder(newQuery.Where(args...))
}

// Count 计数查询（重写以确保使用正确连接）
func (m *BaseModel) Count() (int64, error) {
	// 如果已经有QueryBuilder且有条件，直接使用
	if m.QueryBuilder != nil {
		return m.QueryBuilder.Count()
	}

	// 否则创建新查询
	newQuery := m.newQuerySafe()
	if newQuery == nil {
		return 0, fmt.Errorf("无法创建查询构建器")
	}
	return newQuery.Count()
}

// Get 获取所有记录（重写以确保使用正确连接）
func (m *BaseModel) Get() ([]map[string]interface{}, error) {
	// 如果已经有QueryBuilder且有条件，直接使用
	if m.QueryBuilder != nil {
		return m.QueryBuilder.Get()
	}

	// 否则创建新查询
	newQuery := m.newQuerySafe()
	if newQuery == nil {
		return nil, fmt.Errorf("无法创建查询构建器")
	}
	return newQuery.Get()
}

// First 获取第一条记录（重写以确保使用正确连接）
func (m *BaseModel) First() (map[string]interface{}, error) {
	// 如果已经有QueryBuilder且有条件，直接使用
	if m.QueryBuilder != nil {
		return m.QueryBuilder.First()
	}

	// 否则创建新查询
	newQuery := m.newQuerySafe()
	if newQuery == nil {
		return nil, fmt.Errorf("无法创建查询构建器")
	}
	return newQuery.First()
}

// Update 更新记录（重写以确保使用正确连接）
func (m *BaseModel) Update(data map[string]interface{}) (int64, error) {
	// 如果已经有QueryBuilder且有条件，直接使用
	if m.QueryBuilder != nil {
		return m.QueryBuilder.Update(data)
	}

	// 否则创建新查询
	newQuery := m.newQuerySafe()
	if newQuery == nil {
		return 0, fmt.Errorf("无法创建查询构建器")
	}
	return newQuery.Update(data)
}

// Delete 删除记录（重写以确保使用正确连接）
func (m *BaseModel) Delete() (int64, error) {
	// 如果已经有QueryBuilder且有条件，直接使用
	if m.QueryBuilder != nil {
		return m.QueryBuilder.Delete()
	}

	// 否则创建新查询
	newQuery := m.newQuerySafe()
	if newQuery == nil {
		return 0, fmt.Errorf("无法创建查询构建器")
	}
	return newQuery.Delete()
}

// Select 选择字段（重写以确保使用正确连接）
func (m *BaseModel) Select(args ...interface{}) *BaseModel {
	// 如果已经有QueryBuilder，直接使用它继续链式调用
	if m.QueryBuilder != nil {
		return m.wrapQueryBuilder(m.QueryBuilder.Select(args...))
	}

	// 否则创建新查询
	newQuery := m.newQuerySafe()
	if newQuery == nil {
		return m // fallback
	}
	return m.wrapQueryBuilder(newQuery.Select(args...))
}

// OrderBy 排序（重写以确保使用正确连接）
func (m *BaseModel) OrderBy(column string, direction string) *BaseModel {
	// 如果已经有QueryBuilder，直接使用它继续链式调用
	if m.QueryBuilder != nil {
		return m.wrapQueryBuilder(m.QueryBuilder.OrderBy(column, direction))
	}

	// 否则创建新查询
	newQuery := m.newQuerySafe()
	if newQuery == nil {
		return m // fallback
	}
	return m.wrapQueryBuilder(newQuery.OrderBy(column, direction))
}

// Page 分页（重写以确保使用正确连接）
func (m *BaseModel) Page(page, size int) *BaseModel {
	// 如果已经有QueryBuilder，直接使用它继续链式调用
	if m.QueryBuilder != nil {
		return m.wrapQueryBuilder(m.QueryBuilder.Page(page, size))
	}

	// 否则创建新查询
	newQuery := m.newQuerySafe()
	if newQuery == nil {
		return m // fallback
	}
	return m.wrapQueryBuilder(newQuery.Page(page, size))
}

// newQuerySafe 安全地创建新查询（不会 panic）
func (m *BaseModel) newQuerySafe() *db.QueryBuilder {
	query, err := db.NewQueryBuilder(m.connection)
	if err != nil {
		return nil
	}

	tableName := m.GetTableName()
	if tableName == "" {
		return nil
	}

	return query.From(tableName)
}

// wrapQueryBuilder 包装QueryBuilder为BaseModel实例
func (m *BaseModel) wrapQueryBuilder(qb *db.QueryBuilder) *BaseModel {
	return &BaseModel{
		QueryBuilder: qb,
		tableName:    m.tableName,
		primaryKey:   m.primaryKey,
		connection:   m.connection,
		attributes:   make(map[string]interface{}),
		exists:       false,
		timestamps:   m.timestamps,
		createdAt:    m.createdAt,
		updatedAt:    m.updatedAt,
		softDeletes:  m.softDeletes,
		deletedAt:    m.deletedAt,
		timeManager:  m.timeManager,
		timeFields:   m.timeFields,
	}
}

// EnableTimestamps 启用时间戳
func (m *BaseModel) EnableTimestamps() *BaseModel {
	m.timestamps = true
	return m
}

// DisableTimestamps 禁用时间戳
func (m *BaseModel) DisableTimestamps() *BaseModel {
	m.timestamps = false
	return m
}

// SetCreatedAtField 设置创建时间字段
func (m *BaseModel) SetCreatedAtField(field string) *BaseModel {
	m.createdAt = field
	return m
}

// SetUpdatedAtField 设置更新时间字段
func (m *BaseModel) SetUpdatedAtField(field string) *BaseModel {
	m.updatedAt = field
	return m
}

// GetCreatedAtField 获取创建时间字段
func (m *BaseModel) GetCreatedAtField() string {
	return m.createdAt
}

// GetUpdatedAtField 获取更新时间字段
func (m *BaseModel) GetUpdatedAtField() string {
	return m.updatedAt
}

// EnableSoftDeletes 启用软删除
func (m *BaseModel) EnableSoftDeletes() *BaseModel {
	m.softDeletes = true
	return m
}

// DisableSoftDeletes 禁用软删除
func (m *BaseModel) DisableSoftDeletes() *BaseModel {
	m.softDeletes = false
	return m
}

// SetDeletedAtField 设置软删除字段
func (m *BaseModel) SetDeletedAtField(field string) *BaseModel {
	m.deletedAt = field
	return m
}

// 属性方法

// SetAttribute 设置属性（支持设置器处理）
func (m *BaseModel) SetAttribute(key string, value interface{}) *BaseModel {
	// 简化实现：直接设置值，不在这里处理访问器
	// 访问器处理应该在具体的模型方法中调用
	m.attributes[key] = value
	return m
}

// GetAttribute 获取属性
func (m *BaseModel) GetAttribute(key string) interface{} {
	return m.attributes[key]
}

// SetAttributeWithAccessor 设置属性并应用设置器（需要传入具体模型实例）
func (m *BaseModel) SetAttributeWithAccessor(modelInstance interface{}, key string, value interface{}) *BaseModel {
	// 创建访问器处理器
	processor := db.NewAccessorProcessor(modelInstance)

	// 应用设置器处理
	processedData := processor.ProcessSetData(map[string]interface{}{key: value})

	// 设置处理后的值
	if processedValue, exists := processedData[key]; exists {
		m.attributes[key] = processedValue
	} else {
		m.attributes[key] = value
	}

	return m
}

// SetAttributesWithAccessor 批量设置属性并应用设置器（需要传入具体模型实例）- 性能优化版本
func (m *BaseModel) SetAttributesWithAccessor(modelInstance interface{}, attributes map[string]interface{}) *BaseModel {
	// 使用缓存的访问器处理器
	processor := m.getAccessorProcessor(modelInstance)

	// 应用设置器处理
	processedData := processor.ProcessSetData(attributes)

	// 设置处理后的值 - 优化批量设置
	if len(processedData) > 0 {
		for key, value := range processedData {
			m.attributes[key] = value
		}
	}

	return m
}

// getAccessorProcessor 获取缓存的访问器处理器 - 性能优化
func (m *BaseModel) getAccessorProcessor(modelInstance interface{}) *db.AccessorProcessor {
	m.accessorMutex.RLock()
	if m.accessorProcessor != nil {
		processor := m.accessorProcessor
		m.accessorMutex.RUnlock()
		return processor
	}
	m.accessorMutex.RUnlock()

	// 创建新的处理器
	m.accessorMutex.Lock()
	defer m.accessorMutex.Unlock()

	// 双重检查
	if m.accessorProcessor == nil {
		m.accessorProcessor = db.NewAccessorProcessor(modelInstance)
	}

	return m.accessorProcessor
}

// GetAttributes 获取所有属性
func (m *BaseModel) GetAttributes() map[string]interface{} {
	return m.attributes
}

// SetAttributes 批量设置属性（支持设置器处理）
func (m *BaseModel) SetAttributes(attributes map[string]interface{}) *BaseModel {
	// 简化实现：直接设置值，不在这里处理访问器
	// 访问器处理应该在具体的模型方法中调用
	for key, value := range attributes {
		m.attributes[key] = value
	}
	return m
}

// ClearAttributes 清空属性
func (m *BaseModel) ClearAttributes() *BaseModel {
	m.attributes = make(map[string]interface{})
	return m
}

// 状态方法

// IsNew 检查是否是新记录
func (m *BaseModel) IsNew() bool {
	return !m.exists
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

// 模型特有的查询方法（重写以处理时间戳等）

// Save 保存模型
func (m *BaseModel) Save() error {
	query, err := m.newQuery()
	if err != nil {
		return db.WrapError(err, db.ErrCodeModelSaveFailed, "创建查询失败")
	}

	if m.IsNew() {
		// 插入新记录
		data := m.prepareForInsert()
		if len(data) == 0 {
			return db.ErrModelValidationFailed.WithDetails("没有要插入的数据")
		}

		id, err := query.Insert(data)
		if err != nil {
			if db.IsDuplicateError(err) {
				wrappedErr := db.WrapError(err, db.ErrCodeDuplicateKey, "模型保存失败：违反唯一性约束").
					WithContext("table", m.GetTableName()).
					WithContext("data", data)
				db.LogError(wrappedErr)
				return wrappedErr
			}
			wrappedErr := db.WrapError(err, db.ErrCodeModelSaveFailed, "模型插入失败").
				WithContext("table", m.GetTableName()).
				WithContext("data", data)
			db.LogError(wrappedErr)
			return wrappedErr
		}

		// 设置主键值（如果是自增的）
		if id > 0 {
			m.SetAttribute(m.primaryKey, id)
		}

		m.MarkAsExists()
		return nil
	} else {
		// 更新现有记录
		data := m.prepareForUpdate()
		if len(data) == 0 {
			return nil // 没有需要更新的数据
		}

		pk := m.GetAttribute(m.primaryKey)
		if pk == nil {
			return db.ErrInvalidModelState.
				WithDetails("主键值不能为空").
				WithContext("table", m.GetTableName()).
				WithContext("primary_key", m.primaryKey)
		}

		affected, err := query.Where(m.primaryKey, "=", pk).Update(data)
		if err != nil {
			if db.IsDuplicateError(err) {
				return db.WrapError(err, db.ErrCodeDuplicateKey, "模型更新失败：违反唯一性约束").
					WithContext("table", m.GetTableName()).
					WithContext("primary_key", pk).
					WithContext("data", data)
			}
			return db.WrapError(err, db.ErrCodeModelSaveFailed, "模型更新失败").
				WithContext("table", m.GetTableName()).
				WithContext("primary_key", pk).
				WithContext("data", data)
		}

		if affected == 0 {
			return db.ErrModelNotFound.
				WithDetails("没有找到要更新的记录").
				WithContext("table", m.GetTableName()).
				WithContext("primary_key", pk)
		}

		return nil
	}
}

// FindByPK 根据主键查找
func (m *BaseModel) FindByPK(key interface{}) error {
	if key == nil {
		return db.ErrInvalidParameter.WithDetails("主键值不能为空")
	}

	query, err := m.newQuery()
	if err != nil {
		return db.WrapError(err, db.ErrCodeQueryFailed, "创建查询失败")
	}

	result, err := query.Where(m.primaryKey, "=", key).FirstRaw()
	if err != nil {
		if db.IsNotFoundError(err) {
			return db.ErrModelNotFound.
				WithContext("table", m.GetTableName()).
				WithContext("primary_key", m.primaryKey).
				WithContext("key_value", key)
		}
		return db.WrapError(err, db.ErrCodeModelNotFound, "查找模型失败").
			WithContext("table", m.GetTableName()).
			WithContext("primary_key", m.primaryKey).
			WithContext("key_value", key)
	}

	m.fill(result)
	m.MarkAsExists()
	return nil
}

// SoftDelete 软删除（如果启用）
func (m *BaseModel) SoftDelete() error {
	if !m.softDeletes {
		return db.ErrInvalidModelState.WithDetails("该模型未启用软删除")
	}

	query, err := m.newQuery()
	if err != nil {
		return db.WrapError(err, db.ErrCodeModelDeleteFailed, "创建查询失败")
	}

	pk := m.GetAttribute(m.primaryKey)
	if pk == nil {
		return db.ErrInvalidModelState.
			WithDetails("主键值不能为空").
			WithContext("table", m.GetTableName()).
			WithContext("primary_key", m.primaryKey)
	}

	data := map[string]interface{}{
		m.deletedAt: time.Now(),
	}

	affected, err := query.Where(m.primaryKey, "=", pk).Update(data)
	if err != nil {
		return db.WrapError(err, db.ErrCodeModelDeleteFailed, "软删除失败").
			WithContext("table", m.GetTableName()).
			WithContext("primary_key", pk)
	}

	if affected == 0 {
		return db.ErrModelNotFound.
			WithDetails("没有找到要删除的记录").
			WithContext("table", m.GetTableName()).
			WithContext("primary_key", pk)
	}

	return nil
}

// Restore 恢复软删除的记录
func (m *BaseModel) Restore() error {
	if !m.softDeletes {
		return fmt.Errorf("该模型未启用软删除")
	}

	query, err := m.newQuery()
	if err != nil {
		return err
	}

	pk := m.GetAttribute(m.primaryKey)
	if pk == nil {
		return fmt.Errorf("主键值不能为空")
	}

	data := map[string]interface{}{
		m.deletedAt: nil,
	}

	_, err = query.Where(m.primaryKey, "=", pk).Update(data)
	return err
}

// ForceDelete 强制删除（真实删除）
func (m *BaseModel) ForceDelete() error {
	query, err := m.newQuery()
	if err != nil {
		return db.WrapError(err, db.ErrCodeModelDeleteFailed, "创建查询失败")
	}

	pk := m.GetAttribute(m.primaryKey)
	if pk == nil {
		return db.ErrInvalidModelState.
			WithDetails("主键值不能为空").
			WithContext("table", m.GetTableName()).
			WithContext("primary_key", m.primaryKey)
	}

	affected, err := query.Where(m.primaryKey, "=", pk).Delete()
	if err != nil {
		return db.WrapError(err, db.ErrCodeModelDeleteFailed, "强制删除失败").
			WithContext("table", m.GetTableName()).
			WithContext("primary_key", pk)
	}

	if affected == 0 {
		return db.ErrModelNotFound.
			WithDetails("没有找到要删除的记录").
			WithContext("table", m.GetTableName()).
			WithContext("primary_key", pk)
	}

	m.MarkAsNew()
	return nil
}

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
// 内部辅助方法
// ============================================================================

// newQuery 创建新的查询构建器
func (m *BaseModel) newQuery() (*db.QueryBuilder, error) {
	query, err := db.NewQueryBuilder(m.connection)
	if err != nil {
		return nil, db.WrapError(err, db.ErrCodeConnectionFailed, "创建查询构建器失败").
			WithContext("connection", m.connection)
	}

	tableName := m.GetTableName()
	if tableName == "" {
		return nil, db.ErrInvalidModelState.
			WithDetails("表名未设置，请使用 SetTable() 方法设置表名").
			WithContext("model_type", fmt.Sprintf("%T", m))
	}

	return query.From(tableName), nil
}

// fill 填充模型属性
func (m *BaseModel) fill(data map[string]interface{}) {
	// 直接填充数据，不再使用Result系统
	for key, value := range data {
		m.attributes[key] = value
	}
}

// prepareForInsert 准备插入数据
func (m *BaseModel) prepareForInsert() map[string]interface{} {
	data := make(map[string]interface{})

	// 获取所有属性
	attrs := m.GetAttributes()
	for key, value := range attrs {
		data[key] = value
	}

	// 处理传统时间戳字段（向后兼容）
	if m.timestamps {
		now := time.Now()
		data[m.createdAt] = now
		data[m.updatedAt] = now
	}

	// 处理新的时间字段管理
	if m.timeManager != nil && len(m.timeFields) > 0 {
		data = m.timeManager.ProcessInsertData(data, m.timeFields)
	}

	return data
}

// prepareForUpdate 准备更新数据
func (m *BaseModel) prepareForUpdate() map[string]interface{} {
	data := make(map[string]interface{})

	// 获取所有属性，除了主键
	attrs := m.GetAttributes()
	for key, value := range attrs {
		if key != m.primaryKey {
			data[key] = value
		}
	}

	// 处理传统时间戳字段（向后兼容）
	if m.timestamps {
		data[m.updatedAt] = time.Now()
	}

	// 处理新的时间字段管理
	if m.timeManager != nil && len(m.timeFields) > 0 {
		data = m.timeManager.ProcessUpdateData(data, m.timeFields)
	}

	return data
}

// toSnakeCase 转换为蛇形命名（增强版，支持连续大写字母）
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

// ============================================================================
// 测试和迁移相关方法
// ============================================================================

// AutoMigrate 自动迁移表结构
func (m *BaseModel) AutoMigrate(models ...interface{}) error {
	// 获取数据库管理器并创建连接
	manager := db.DefaultManager()
	conn, err := manager.Connection(m.connection)
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 创建迁移器
	migrator := migration.NewAutoMigrator(conn)

	// 启用性能优化：表存在性缓存
	migrator.SetCacheEnabled(true)

	// 在测试环境中启用快速模式（跳过重复的结构检查）
	if len(models) > 1 {
		migrator.SetSkipIfExists(true)
	}

	// 如果没有传入模型，使用当前模型
	if len(models) == 0 {
		models = []interface{}{m}
	}

	// 迁移每个模型
	for _, model := range models {
		var tableName string

		// 获取表名 - 绝对优先使用用户设置的表名
		if baseModel, ok := model.(*BaseModel); ok {
			// 直接是BaseModel实例
			tableName = baseModel.GetTableName()
		} else {
			// 检查是否有嵌入的BaseModel
			modelValue := reflect.ValueOf(model)
			if modelValue.Kind() == reflect.Ptr {
				modelValue = modelValue.Elem()
			}

			// 尝试获取BaseModel字段
			if modelValue.Kind() == reflect.Struct {
				baseModelField := modelValue.FieldByName("BaseModel")
				if baseModelField.IsValid() && baseModelField.Type() == reflect.TypeOf(BaseModel{}) {
					// 获取BaseModel实例 - 使用指针以确保获取正确的表名
					if baseModelField.CanAddr() {
						baseModelPtr := baseModelField.Addr().Interface().(*BaseModel)
						tableName = baseModelPtr.GetTableName()
					} else {
						// 如果无法获取地址，跳过以避免锁复制
						// 这种情况下使用默认的类型推断
					}
				}
			}
		}

		// 如果还是没有表名，则使用类型推断作为最后手段
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

// Fill 填充模型属性（公开方法）
func (m *BaseModel) Fill(data map[string]interface{}) *BaseModel {
	for key, value := range data {
		m.attributes[key] = value
	}
	return m
}

// GetKey 获取主键值
func (m *BaseModel) GetKey() interface{} {
	return m.GetAttribute(m.primaryKey)
}

// SetKey 设置主键值
func (m *BaseModel) SetKey(key interface{}) *BaseModel {
	m.SetAttribute(m.primaryKey, key)
	return m
}

// Find 根据主键查找记录（重载版本，返回两个值以兼容测试）
func (m *BaseModel) Find(pk interface{}) error {
	return m.FindByPK(pk)
}

// 注释：已移除Result系统相关方法，现在统一使用原始数据API
// 如需要访问器功能，请在业务层手动处理

// ============================================================================
// 模型关联方法
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

// ============================================================================
// 标签处理方法
// ============================================================================

// ParseModelTags 解析模型的torm标签，自动配置模型属性
func (m *BaseModel) ParseModelTags(modelInstance interface{}) *BaseModel {
	modelType := reflect.TypeOf(modelInstance)
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
		m.parseFieldTag(field, tormTag)
	}

	return m
}

// RefreshTimeFields 重新分析模型的时间字段
func (m *BaseModel) RefreshTimeFields(modelInstance interface{}) *BaseModel {
	if m.timeManager != nil {
		m.timeFields = m.timeManager.AnalyzeModelTimeFields(modelInstance)
	}
	return m
}

// GetTimeFields 获取时间字段信息
func (m *BaseModel) GetTimeFields() []db.TimeFieldInfo {
	return m.timeFields
}

// HasTimeFields 检查是否有时间字段
func (m *BaseModel) HasTimeFields() bool {
	return len(m.timeFields) > 0
}

// parseFieldTag 解析单个字段的标签
func (m *BaseModel) parseFieldTag(field reflect.StructField, tormTag string) {
	parts := strings.Split(tormTag, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)

		switch {
		case part == "primary_key":
			// 设置主键
			columnName := getColumnNameFromField(field)
			m.SetPrimaryKey(columnName)

		case part == "auto_increment":
			// 自增字段处理（通常与primary_key一起使用）
			continue

		case strings.HasPrefix(part, "type:"):
			// 字段类型，用于数据库迁移
			continue

		case strings.HasPrefix(part, "size:"):
			// 字段大小，用于数据库迁移
			continue

		case strings.HasPrefix(part, "default:"):
			// 默认值，用于数据库迁移
			continue

		case part == "unique":
			// 唯一约束，用于数据库迁移
			continue

		case part == "index":
			// 索引，用于数据库迁移
			continue

		case strings.HasPrefix(part, "references:"):
			// 外键引用，用于数据库迁移
			continue

		case part == "auto_create_time":
			// 自动创建时间字段
			columnName := getColumnNameFromField(field)
			m.SetCreatedAtField(columnName)

		case part == "auto_update_time":
			// 自动更新时间字段
			columnName := getColumnNameFromField(field)
			m.SetUpdatedAtField(columnName)

		case part == "soft_delete":
			// 软删除字段
			columnName := getColumnNameFromField(field)
			m.SetDeletedAtField(columnName)
			m.EnableSoftDeletes()
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
