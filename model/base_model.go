package model

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/zhoudm1743/torm/db"
	"github.com/zhoudm1743/torm/migration"
)

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

	// 模型状态
	exists bool

	// 时间管理
	timeManager *db.TimeFieldManager
	timeFields  []db.TimeFieldInfo
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

// GetTableName 获取表名
func (m *BaseModel) GetTableName() string {
	return m.config.TableName
}

// TableName 静态函数，用于查询构建器直接获取表名
// 这个函数可以被子模型重写来提供自定义表名
func (m *BaseModel) TableName() string {
	return m.config.TableName
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

// SetConnection 设置连接
func (m *BaseModel) SetConnection(connection string) *BaseModel {
	m.config.Connection = connection
	return m
}

// GetConnection 获取连接
func (m *BaseModel) GetConnection() string {
	return m.config.Connection
}

// EnableTimestamps 启用时间戳
func (m *BaseModel) EnableTimestamps() *BaseModel {
	m.config.Timestamps = true
	return m
}

// DisableTimestamps 禁用时间戳
func (m *BaseModel) DisableTimestamps() *BaseModel {
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

// Query 创建查询构建器
func (m *BaseModel) Query() (*db.QueryBuilder, error) {
	if m.config.TableName == "" {
		return nil, fmt.Errorf("表名未设置，请使用 SetTable() 方法设置表名")
	}

	query, err := db.NewQueryBuilder(m.config.Connection)
	if err != nil {
		return nil, fmt.Errorf("创建查询构建器失败: %w", err)
	}

	// 绑定模型实例以支持访问器处理
	return query.From(m.config.TableName).WithModel(m), nil
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
func (m *BaseModel) SetAttribute(key string, value interface{}) *BaseModel {
	m.attributes[key] = value
	return m
}

// GetAttribute 获取属性
func (m *BaseModel) GetAttribute(key string) interface{} {
	return m.attributes[key]
}

// SetAttributes 批量设置属性
func (m *BaseModel) SetAttributes(attributes map[string]interface{}) *BaseModel {
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
	for key, value := range data {
		m.attributes[key] = value
	}
	return m
}

// ============================================================================
// 状态管理方法
// ============================================================================

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
	query, err := m.Query()
	if err != nil {
		return err
	}

	if m.IsNew() {
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
		return nil
	} else {
		// 更新现有记录
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

		return nil
	}
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

	m.Fill(result)
	m.MarkAsExists()
	return nil
}

// Find 根据主键查找记录（别名方法）
func (m *BaseModel) Find(pk interface{}) error {
	return m.FindByPK(pk)
}

// Delete 删除记录
func (m *BaseModel) Delete() error {
	if m.config.SoftDeletes {
		return m.SoftDelete()
	}
	return m.ForceDelete()
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

	data := map[string]interface{}{
		m.config.DeletedAtCol: time.Now(),
	}

	affected, err := query.Where(m.config.PrimaryKey, "=", pk).Update(data)
	if err != nil {
		return fmt.Errorf("软删除失败: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("没有找到要删除的记录")
	}

	return nil
}

// Restore 恢复软删除的记录
func (m *BaseModel) Restore() error {
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

	data := map[string]interface{}{
		m.config.DeletedAtCol: nil,
	}

	_, err = query.Where(m.config.PrimaryKey, "=", pk).Update(data)
	return err
}

// ForceDelete 强制删除（真实删除）
func (m *BaseModel) ForceDelete() error {
	query, err := m.Query()
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

// prepareForUpdate 准备更新数据
func (m *BaseModel) prepareForUpdate() map[string]interface{} {
	data := make(map[string]interface{})

	// 获取所有属性，除了主键
	for key, value := range m.attributes {
		if key != m.config.PrimaryKey {
			data[key] = value
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
func parseFieldTag(field reflect.StructField, tormTag string, config *ModelConfig) {
	parts := strings.Split(tormTag, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
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
func parseTormKeyValue(part string, field reflect.StructField, config *ModelConfig) {
	kv := strings.SplitN(part, ":", 2)
	if len(kv) != 2 {
		return
	}

	key := strings.ToLower(strings.TrimSpace(kv[0]))
	_ = strings.TrimSpace(kv[1]) // value currently not used in model config, handled by migration

	switch key {
	case "type":
		// 数据库类型：type:varchar, type:decimal等
		// 模型配置层面不处理具体的数据库类型，这由migration包处理

	case "size":
		// 字段大小：size:50, size:100等
		// 模型配置层面不处理，由migration包处理

	case "precision":
		// 精度：precision:10 (用于decimal类型)
		// 模型配置层面不处理，由migration包处理

	case "scale":
		// 标度：scale:2 (用于decimal类型)
		// 模型配置层面不处理，由migration包处理

	case "default":
		// 默认值：default:0, default:'active'等
		// 模型配置层面不处理，由migration包处理

	case "comment":
		// 字段注释：comment:'用户名'
		// 模型配置层面不处理，由migration包处理

	case "column", "db":
		// 自定义列名：column:user_name, db:user_id
		// 这主要用于字段映射，在模型配置层面不需要特殊处理

	case "foreign_key", "fk", "references", "ref":
		// 外键：foreign_key:users.id, references:roles(id)
		// 模型配置层面不处理，由migration包处理

	case "on_delete":
		// 删除时动作：on_delete:cascade
		// 模型配置层面不处理，由migration包处理

	case "on_update":
		// 更新时动作：on_update:cascade
		// 模型配置层面不处理，由migration包处理

	case "generated":
		// 生成列：generated:virtual, generated:stored
		// 模型配置层面不处理，由migration包处理

	case "index":
		// 带类型的索引：index:btree, index:hash
		// 模型配置层面不处理，由migration包处理
	}
}

// parseTormFlagAdvanced 解析torm标签的标志 - 扩展版本
func parseTormFlagAdvanced(flag string, field reflect.StructField, config *ModelConfig) {
	flag = strings.ToLower(flag)
	columnName := getColumnNameFromField(field)

	switch flag {
	case "primary_key", "pk", "primary", "primarykey":
		// 设置主键
		config.PrimaryKey = columnName

	case "auto_increment", "autoincrement", "auto_inc", "autoinc":
		// 自增字段 - 这通常与primary_key一起使用
		// 在模型配置中，我们主要关心主键设置
		// auto_increment的具体处理在迁移时由migration包负责
		// PostgreSQL会自动使用SERIAL/BIGSERIAL等序列类型

	case "auto_create_time", "create_time", "created_at", "auto_created_at",
		"autocreate_time", "autocreatetime", "auto_create", "autocreate":
		// 自动创建时间字段 - 支持多种标记格式
		config.CreatedAtCol = columnName

	case "auto_update_time", "update_time", "updated_at", "auto_updated_at",
		"autoupdate_time", "autoupdatetime", "auto_update", "autoupdate":
		// 自动更新时间字段 - 支持多种标记格式
		config.UpdatedAtCol = columnName

	case "soft_delete", "soft_deletes", "deleted_at":
		// 软删除字段
		config.DeletedAtCol = columnName
		config.SoftDeletes = true

	// 以下标志主要用于数据库迁移，模型配置不直接处理
	// 但我们仍然识别它们以确保标签解析的完整性
	case "unique", "uniq":
		// 唯一索引 - 由migration包处理

	case "not_null", "not null", "notnull", "not_nil", "notnil":
		// 非空约束 - 由migration包处理

	case "nullable", "null":
		// 可空约束 - 由migration包处理

	case "timestamp", "current_timestamp":
		// 时间戳字段 - 由migration包处理

	case "unsigned":
		// 无符号数字类型 (主要用于MySQL) - 由migration包处理

	case "zerofill":
		// 零填充 (主要用于MySQL) - 由migration包处理

	case "binary":
		// 二进制存储 - 由migration包处理

	case "index", "idx":
		// 普通索引 - 由migration包处理

	case "fulltext", "fulltext_index":
		// 全文索引 - 由migration包处理

	case "spatial", "spatial_index":
		// 空间索引 - 由migration包处理

	case "json", "is_json":
		// JSON字段标记 - 由migration包处理

	case "encrypted", "encrypt":
		// 加密字段标记 - 可能需要在模型层处理，但目前不实现

	case "hidden", "invisible":
		// 隐藏字段标记 - 可能需要在模型层处理，但目前不实现

	case "readonly", "immutable":
		// 只读字段标记 - 可能需要在模型层处理，但目前不实现

	// PostgreSQL序列相关 - 这些都由migration包自动处理
	case "serial":
		// PostgreSQL SERIAL类型 - 等同于auto_increment

	case "bigserial":
		// PostgreSQL BIGSERIAL类型

	case "smallserial":
		// PostgreSQL SMALLSERIAL类型

	// 外键级联操作标志
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

	// 生成列相关
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
