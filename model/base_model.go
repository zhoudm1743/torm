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

// BaseModel 基础模型 - 直接嵌入 QueryBuilder，自动获得所有查询方法
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
	timestamps bool
	createdAt  string
	updatedAt  string

	// 软删除
	softDeletes bool
	deletedAt   string
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
	return m
}

// GetConnection 获取连接
func (m *BaseModel) GetConnection() string {
	return m.connection
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

// SetAttribute 设置属性
func (m *BaseModel) SetAttribute(key string, value interface{}) *BaseModel {
	return m.SetAttributeWithAccessor(key, value)
}

// GetAttribute 获取属性
func (m *BaseModel) GetAttribute(key string) interface{} {
	return m.GetAttributeWithAccessor(key)
}

// GetAttributes 获取所有属性
func (m *BaseModel) GetAttributes() map[string]interface{} {
	return m.attributes
}

// SetAttributes 批量设置属性
func (m *BaseModel) SetAttributes(attributes map[string]interface{}) *BaseModel {
	return m.SetAttributesWithAccessors(attributes)
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
		return err
	}

	if m.IsNew() {
		// 插入新记录
		data := m.prepareForInsert()
		id, err := query.Insert(data)
		if err != nil {
			return err
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
			return fmt.Errorf("主键值不能为空")
		}

		_, err := query.Where(m.primaryKey, "=", pk).Update(data)
		return err
	}
}

// FindByPK 根据主键查找
func (m *BaseModel) FindByPK(key interface{}) error {
	query, err := m.newQuery()
	if err != nil {
		return err
	}

	result, err := query.Where(m.primaryKey, "=", key).FirstRaw()
	if err != nil {
		return err
	}

	m.fill(result)
	m.MarkAsExists()
	return nil
}

// SoftDelete 软删除（如果启用）
func (m *BaseModel) SoftDelete() error {
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
		m.deletedAt: time.Now(),
	}

	_, err = query.Where(m.primaryKey, "=", pk).Update(data)
	return err
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
		return err
	}

	pk := m.GetAttribute(m.primaryKey)
	if pk == nil {
		return fmt.Errorf("主键值不能为空")
	}

	_, err = query.Where(m.primaryKey, "=", pk).Delete()
	if err == nil {
		m.MarkAsNew()
	}
	return err
}

// ToJSON 转换为JSON字符串
func (m *BaseModel) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(m.attributes)
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
		return nil, err
	}

	tableName := m.GetTableName()
	if tableName == "" {
		return nil, fmt.Errorf("表名未设置，请使用 SetTable() 方法设置表名")
	}

	return query.From(tableName), nil
}

// fill 填充模型属性
func (m *BaseModel) fill(data map[string]interface{}) {
	// 创建 Result 包装器来处理数据
	result := db.NewResult(data, m)

	// 使用修改器处理每个属性
	for key, value := range data {
		processedValue := result.CallSetAccessor(key, value)
		m.attributes[key] = processedValue
	}
}

// prepareForInsert 准备插入数据
func (m *BaseModel) prepareForInsert() map[string]interface{} {
	data := make(map[string]interface{})

	// 获取所有属性（支持访问器）
	attrs := m.GetAttributesWithAccessors()
	for key, value := range attrs {
		data[key] = value
	}

	// 处理时间戳
	if m.timestamps {
		now := time.Now()
		data[m.createdAt] = now
		data[m.updatedAt] = now
	}

	return data
}

// prepareForUpdate 准备更新数据
func (m *BaseModel) prepareForUpdate() map[string]interface{} {
	data := make(map[string]interface{})

	// 获取所有属性（支持访问器），除了主键
	attrs := m.GetAttributesWithAccessors()
	for key, value := range attrs {
		if key != m.primaryKey {
			data[key] = value
		}
	}

	// 处理时间戳
	if m.timestamps {
		data[m.updatedAt] = time.Now()
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
						// 如果无法获取地址，尝试值访问
						baseModelInstance := baseModelField.Interface().(BaseModel)
						tableName = baseModelInstance.GetTableName()
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

// FindAsResult 根据主键查找记录，返回 Result 类型
func (m *BaseModel) FindAsResult(pk interface{}) (*db.Result, error) {
	query, err := m.newQuery()
	if err != nil {
		return nil, err
	}

	return query.Model(m).Where(m.primaryKey, "=", pk).First()
}

// GetAsResults 查询多条记录，返回 ResultCollection
func (m *BaseModel) GetAsResults() (*db.ResultCollection, error) {
	query, err := m.newQuery()
	if err != nil {
		return nil, err
	}

	return query.Model(m).Get()
}

// FirstAsResult 查询第一条记录，返回 Result 类型
func (m *BaseModel) FirstAsResult() (*db.Result, error) {
	query, err := m.newQuery()
	if err != nil {
		return nil, err
	}

	return query.Model(m).First()
}

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
