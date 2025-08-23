package model

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/zhoudm1743/torm/db"
	"github.com/zhoudm1743/torm/migration"
)

// DeletedTime 软删除时间类型
type DeletedTime struct {
	*time.Time
}

// FieldTags 字段标签配置
type FieldTags struct {
	PrimaryKey     bool   // pk标签
	AutoCreateTime bool   // autoCreateTime标签
	AutoUpdateTime bool   // autoUpdateTime标签
	SoftDelete     bool   // 软删除字段
	FieldName      string // db字段名
}

// ModelMetadata 模型元数据
type ModelMetadata struct {
	TableName      string
	PrimaryKeys    []string
	CreatedAtField string
	UpdatedAtField string
	DeletedAtField string
	HasTimestamps  bool
	HasSoftDeletes bool
	FieldTags      map[string]*FieldTags
}

// ParseModelTags 解析模型标签
func ParseModelTags(model interface{}) *ModelMetadata {
	metadata := &ModelMetadata{
		FieldTags: make(map[string]*FieldTags),
	}

	// 处理nil输入
	if model == nil {
		// 返回默认配置
		metadata.PrimaryKeys = []string{"id"}
		return metadata
	}

	modelType := reflect.TypeOf(model)
	if modelType == nil {
		// 返回默认配置
		metadata.PrimaryKeys = []string{"id"}
		return metadata
	}

	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 如果解引用后仍然为nil，返回默认配置
	if modelType == nil {
		metadata.PrimaryKeys = []string{"id"}
		return metadata
	}

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)

		// 跳过BaseModel字段
		if field.Name == "BaseModel" {
			continue
		}

		tags := &FieldTags{}

		// 解析db标签
		dbTag := field.Tag.Get("db")
		if dbTag != "" && dbTag != "-" {
			// 解析db标签中的选项，如 "created_at;autoCreateTime"
			parts := strings.Split(dbTag, ";")
			tags.FieldName = parts[0]

			for _, part := range parts[1:] {
				switch part {
				case "autoCreateTime":
					tags.AutoCreateTime = true
					metadata.HasTimestamps = true
					metadata.CreatedAtField = tags.FieldName
				case "autoUpdateTime":
					tags.AutoUpdateTime = true
					metadata.HasTimestamps = true
					metadata.UpdatedAtField = tags.FieldName
				}
			}
		} else {
			// 如果没有db标签，使用字段名的小写形式
			tags.FieldName = strings.ToLower(field.Name)
		}

		// 解析pk标签
		// 使用Lookup检查是否存在pk标签，支持 pk="true", pk="", pk
		_, hasPKTag := field.Tag.Lookup("pk")
		if hasPKTag {
			// 如果有pk标签（不管值是什么），则认为是主键
			tags.PrimaryKey = true
			metadata.PrimaryKeys = append(metadata.PrimaryKeys, tags.FieldName)
		}

		// 检查软删除字段
		if field.Type == reflect.TypeOf(DeletedTime{}) {
			tags.SoftDelete = true
			metadata.HasSoftDeletes = true
			metadata.DeletedAtField = tags.FieldName
		}

		metadata.FieldTags[field.Name] = tags
	}

	// 推断表名
	if metadata.TableName == "" {
		modelName := modelType.Name()
		metadata.TableName = strings.ToLower(modelName) + "s" // 简单复数形式
	}

	// 如果没有主键，默认使用id
	if len(metadata.PrimaryKeys) == 0 {
		metadata.PrimaryKeys = []string{"id"}
	}

	return metadata
}

// BaseModel 基础模型
type BaseModel struct {
	// 数据库连接名
	connection string
	// 表名
	tableName string
	// 主键字段（支持复合主键）
	primaryKeys []string
	// 模型属性
	attributes map[string]interface{}
	// 原始属性（用于检测变更）
	original map[string]interface{}
	// 关联数据
	relations map[string]interface{}
	// 是否为新记录
	isNew bool
	// 是否存在于数据库中
	exists bool
	// 时间戳字段
	timestamps bool
	createdAt  string
	updatedAt  string
	// 软删除
	softDeletes bool
	deletedAt   string
	// 内置查询构建器
	queryBuilder db.QueryInterface
	// 模型结构体类型（用于自动迁移）
	modelType reflect.Type
}

// NewBaseModel 创建基础模型实例
func NewBaseModel() *BaseModel {
	return &BaseModel{
		connection:   "default",
		primaryKeys:  []string{"id"},
		attributes:   make(map[string]interface{}),
		original:     make(map[string]interface{}),
		relations:    make(map[string]interface{}),
		isNew:        true,
		exists:       false,
		timestamps:   true,
		createdAt:    "created_at",
		updatedAt:    "updated_at",
		softDeletes:  false,
		deletedAt:    "deleted_at",
		queryBuilder: nil, // 延迟初始化，当第一次使用时创建
	}
}

// NewBaseModelWithAutoDetect 创建基础模型实例并自动检测配置
// 推荐在模型构造函数中使用此方法
func NewBaseModelWithAutoDetect(modelInstance interface{}) *BaseModel {
	baseModel := NewBaseModel()

	// 自动检测配置
	baseModel.DetectConfigFromStruct(modelInstance)

	return baseModel
}

// TableName 获取表名
func (m *BaseModel) TableName() string {
	if m.tableName != "" {
		return m.tableName
	}
	// 如果没有设置表名，返回空字符串，让外部推断
	// 注意：这里不做推断是因为BaseModel没有上下文知道自己被嵌入到哪个结构体中
	return ""
}

// SetTable 设置表名
func (m *BaseModel) SetTable(table string) *BaseModel {
	m.tableName = table
	return m
}

// PrimaryKey 获取主键字段名
func (m *BaseModel) PrimaryKey() string {
	if len(m.primaryKeys) > 0 {
		return m.primaryKeys[0]
	}
	return ""
}

// SetPrimaryKey 设置主键字段名
func (m *BaseModel) SetPrimaryKey(key string) *BaseModel {
	m.primaryKeys = []string{key}
	return m
}

// PrimaryKeys 获取所有主键字段名
func (m *BaseModel) PrimaryKeys() []string {
	return m.primaryKeys
}

// SetPrimaryKeys 设置多个主键字段名（复合主键）
func (m *BaseModel) SetPrimaryKeys(keys []string) *BaseModel {
	m.primaryKeys = keys
	return m
}

// HasCompositePrimaryKey 检查是否有复合主键
func (m *BaseModel) HasCompositePrimaryKey() bool {
	return len(m.primaryKeys) > 1
}

// GetPrimaryKeyValues 获取所有主键的值（用于复合主键）
func (m *BaseModel) GetPrimaryKeyValues() map[string]interface{} {
	values := make(map[string]interface{})
	for _, key := range m.primaryKeys {
		values[key] = m.GetAttribute(key)
	}
	return values
}

// SetPrimaryKeyValues 设置所有主键的值（用于复合主键）
func (m *BaseModel) SetPrimaryKeyValues(values map[string]interface{}) *BaseModel {
	for key, value := range values {
		if m.containsKey(key) {
			m.SetAttribute(key, value)
		}
	}
	return m
}

// containsKey 检查键是否在主键列表中
func (m *BaseModel) containsKey(key string) bool {
	for _, pk := range m.primaryKeys {
		if pk == key {
			return true
		}
	}
	return false
}

// DetectPrimaryKeysFromStruct 从结构体标签中检测主键字段
func (m *BaseModel) DetectPrimaryKeysFromStruct(structValue interface{}) *BaseModel {
	val := reflect.ValueOf(structValue)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()
	var primaryKeys []string

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// 跳过BaseModel字段
		if field.Name == "BaseModel" {
			continue
		}

		// 检查primary标签
		if primaryTag := field.Tag.Get("primary"); primaryTag == "true" {
			// 确定字段名
			fieldName := ""
			if dbTag := field.Tag.Get("db"); dbTag != "" && dbTag != "-" {
				fieldName = dbTag
			} else if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
				fieldName = jsonTag
			} else {
				fieldName = strings.ToLower(field.Name)
			}

			primaryKeys = append(primaryKeys, fieldName)
		}
	}

	// 如果找到了主键标签，使用它们
	if len(primaryKeys) > 0 {
		m.primaryKeys = primaryKeys
	}
	// 否则保持默认的["id"]

	return m
}

// DetectConfigFromStruct 从结构体标签中检测完整配置（时间戳、软删除等）
// 结构体标签优先级高于BaseModel基础配置
func (m *BaseModel) DetectConfigFromStruct(structValue interface{}) *BaseModel {
	metadata := ParseModelTags(structValue)

	// 保存模型结构体类型信息（用于 AutoMigrate）
	modelType := reflect.TypeOf(structValue)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	m.modelType = modelType

	// 更新主键配置
	if len(metadata.PrimaryKeys) > 0 {
		m.primaryKeys = metadata.PrimaryKeys
	}

	// 更新时间戳配置 - 结构体标签优先级更高
	if metadata.HasTimestamps {
		m.timestamps = true
		if metadata.CreatedAtField != "" {
			m.createdAt = metadata.CreatedAtField
		}
		if metadata.UpdatedAtField != "" {
			m.updatedAt = metadata.UpdatedAtField
		}
	}

	// 更新软删除配置 - 结构体标签优先级更高
	if metadata.HasSoftDeletes {
		m.softDeletes = true
		if metadata.DeletedAtField != "" {
			m.deletedAt = metadata.DeletedAtField
		}
	}

	// 设置表名（如果模型没有设置的话）
	if m.tableName == "" {
		m.tableName = metadata.TableName
	}

	return m
}

// GetConnection 获取连接名
func (m *BaseModel) GetConnection() string {
	return m.connection
}

// SetConnection 设置连接名
func (m *BaseModel) SetConnection(connection string) *BaseModel {
	m.connection = connection
	return m
}

// GetAttribute 获取属性值
func (m *BaseModel) GetAttribute(key string) interface{} {
	return m.attributes[key]
}

// SetAttribute 设置属性值
func (m *BaseModel) SetAttribute(key string, value interface{}) {
	// 如果值是 []byte 类型，转换为字符串
	if bytes, ok := value.([]byte); ok {
		m.attributes[key] = string(bytes)
	} else {
		m.attributes[key] = value
	}
}

// GetAttributes 获取所有属性
func (m *BaseModel) GetAttributes() map[string]interface{} {
	return m.attributes
}

// SetAttributes 设置多个属性
func (m *BaseModel) SetAttributes(attributes map[string]interface{}) {
	for key, value := range attributes {
		// 如果值是 []byte 类型，转换为字符串
		if bytes, ok := value.([]byte); ok {
			m.attributes[key] = string(bytes)
		} else {
			m.attributes[key] = value
		}
	}
}

// IsNew 检查是否为新记录
func (m *BaseModel) IsNew() bool {
	return m.isNew
}

// Exists 检查是否存在于数据库中
func (m *BaseModel) Exists() bool {
	return m.exists
}

// IsDirty 检查是否有未保存的更改
func (m *BaseModel) IsDirty() bool {
	return len(m.GetDirty()) > 0
}

// GetDirty 获取已更改的属性
func (m *BaseModel) GetDirty() map[string]interface{} {
	dirty := make(map[string]interface{})

	for key, value := range m.attributes {
		if original, exists := m.original[key]; !exists || !reflect.DeepEqual(value, original) {
			dirty[key] = value
		}
	}

	return dirty
}

// Fill 批量赋值
func (m *BaseModel) Fill(attributes map[string]interface{}) *BaseModel {
	m.SetAttributes(attributes)
	return m
}

// Save 保存模型到数据库
func (m *BaseModel) Save() error {
	if m.isNew {
		return m.create()
	}
	return m.update()
}

// create 创建新记录
func (m *BaseModel) create() error {
	// 添加时间戳
	if m.timestamps {
		now := time.Now()
		if m.createdAt != "" && m.GetAttribute(m.createdAt) == nil {
			m.SetAttribute(m.createdAt, now)
		}
		if m.updatedAt != "" && m.GetAttribute(m.updatedAt) == nil {
			m.SetAttribute(m.updatedAt, now)
		}
	}

	// 执行 before_create 钩子
	if err := m.BeforeCreate(); err != nil {
		return err
	}

	// 执行 before_save 钩子
	if err := m.BeforeSave(); err != nil {
		return err
	}

	// 获取查询构造器
	query, err := db.Table(m.TableName(), m.connection)
	if err != nil {
		return err
	}

	// 插入数据
	id, err := query.Insert(m.attributes)
	if err != nil {
		return err
	}

	// 设置主键值
	m.SetAttribute(m.PrimaryKey(), id)

	// 更新状态
	m.isNew = false
	m.exists = true
	m.syncOriginal()

	// 执行 after_create 钩子
	if err := m.AfterCreate(); err != nil {
		return err
	}

	// 执行 after_save 钩子
	return m.AfterSave()
}

// update 更新记录
func (m *BaseModel) update() error {
	dirty := m.GetDirty()
	if len(dirty) == 0 {
		return nil // 没有更改，无需更新
	}

	// 添加更新时间戳
	if m.timestamps && m.updatedAt != "" {
		dirty[m.updatedAt] = time.Now()
		m.SetAttribute(m.updatedAt, dirty[m.updatedAt])
	}

	// 执行 before_update 钩子
	if err := m.BeforeUpdate(); err != nil {
		return err
	}

	// 执行 before_save 钩子
	if err := m.BeforeSave(); err != nil {
		return err
	}

	// 获取查询构造器
	query, err := db.Table(m.TableName(), m.connection)
	if err != nil {
		return err
	}

	// 更新数据
	pkValue := m.GetAttribute(m.PrimaryKey())
	if pkValue == nil {
		return fmt.Errorf("primary key value is required for update")
	}

	_, err = query.Where(m.PrimaryKey(), "=", pkValue).Update(dirty)
	if err != nil {
		return err
	}

	// 同步原始数据
	m.syncOriginal()

	// 执行 after_update 钩子
	if err := m.AfterUpdate(); err != nil {
		return err
	}

	// 执行 after_save 钩子
	return m.AfterSave()
}

// Delete 删除记录 - 支持两种调用方式
// 1. Delete() - 删除当前模型实例
// 2. 链式调用如 Where(...).Delete() - 批量删除
func (m *BaseModel) Delete() (interface{}, error) {
	// 如果有查询条件，执行批量删除
	if m.queryBuilder != nil {
		return m.deleteBatch()
	}

	// 否则删除当前模型实例
	return nil, m.deleteCurrentModel()
}

// deleteCurrentModel 删除当前模型实例
func (m *BaseModel) deleteCurrentModel() error {
	if m.isNew {
		return fmt.Errorf("cannot delete unsaved model")
	}

	// 执行 before_delete 钩子
	if err := m.BeforeDelete(); err != nil {
		return err
	}

	// 获取查询构造器
	query, err := db.Table(m.TableName(), m.connection)
	if err != nil {
		return err
	}

	// 构建主键条件
	query = m.buildPrimaryKeyConditions(query)
	if query == nil {
		return fmt.Errorf("primary key values are required for delete")
	}

	if m.softDeletes {
		// 软删除
		deleteData := map[string]interface{}{
			m.deletedAt: time.Now(),
		}
		_, err = query.Update(deleteData)
		if err != nil {
			return err
		}
		m.SetAttribute(m.deletedAt, time.Now())
	} else {
		// 硬删除
		_, err = query.Delete()
		if err != nil {
			return err
		}
		m.exists = false
	}

	// 执行 after_delete 钩子
	return m.AfterDelete()
}

// deleteBatch 批量删除记录 - 适配db.Delete
func (m *BaseModel) deleteBatch() (int64, error) {
	query := m.getQueryBuilder()
	if query == nil {
		return 0, fmt.Errorf("failed to create query builder")
	}

	var result int64
	var err error

	if m.softDeletes {
		// 软删除：更新deleted_at字段
		deleteData := map[string]interface{}{
			m.deletedAt: time.Now(),
		}
		result, err = query.Update(deleteData)
	} else {
		// 硬删除
		result, err = query.Delete()
	}

	m.resetQueryBuilder() // 执行后重置查询构建器
	return result, err
}

// Reload 重新加载模型数据
func (m *BaseModel) Reload() error {
	if m.isNew {
		return fmt.Errorf("cannot reload unsaved model")
	}

	pkValue := m.GetAttribute(m.PrimaryKey())
	if pkValue == nil {
		return fmt.Errorf("primary key value is required for reload")
	}

	// 获取查询构造器
	query, err := db.Table(m.TableName(), m.connection)
	if err != nil {
		return err
	}

	// 查询数据
	data, err := query.Where(m.PrimaryKey(), "=", pkValue).First()
	if err != nil {
		return err
	}

	// 更新属性
	m.attributes = data
	m.syncOriginal()

	return nil
}

// syncOriginal 同步原始数据
func (m *BaseModel) syncOriginal() {
	m.original = make(map[string]interface{})
	for key, value := range m.attributes {
		m.original[key] = value
	}
}

// Find 查找记录 - 支持多种调用方式
// 1. Find(id, dest...) - 根据主键查找
// 2. Find(dest) - 根据之前的Where条件查找
// 返回原始的map数据
func (m *BaseModel) Find(args ...interface{}) (map[string]interface{}, error) {
	var data map[string]interface{}
	var err error

	if len(args) == 0 {
		return nil, fmt.Errorf("Find() requires at least one argument")
	}

	// 判断调用方式
	firstArg := args[0]

	// 如果第一个参数是指针类型，说明是Find(dest)方式
	if reflect.TypeOf(firstArg).Kind() == reflect.Ptr {
		// 使用现有的查询条件查找
		query := m.getQueryBuilder()
		if query == nil {
			return nil, fmt.Errorf("failed to create query builder")
		}

		data, err = query.First()
		m.resetQueryBuilder() // 执行后重置查询构建器

		if err != nil {
			return nil, err
		}

		// 填充到指针指向的对象
		err = m.LoadModel(firstArg, data)
		if err != nil {
			return data, fmt.Errorf("failed to load model: %w", err)
		}
	} else {
		// 否则是Find(id, dest...)方式
		id := firstArg
		query, err := db.Table(m.TableName(), m.connection)
		if err != nil {
			return nil, err
		}

		data, err = query.Where(m.PrimaryKey(), "=", id).First()
		if err != nil {
			return nil, err
		}

		// 如果有第二个参数且是指针，填充到指针指向的对象
		if len(args) > 1 && args[1] != nil {
			if reflect.TypeOf(args[1]).Kind() == reflect.Ptr {
				err = m.LoadModel(args[1], data)
				if err != nil {
					return data, fmt.Errorf("failed to load model: %w", err)
				}
			}
		}
	}

	// 填充当前模型属性
	m.attributes = data
	m.syncOriginal()
	m.isNew = false
	m.exists = true

	// 执行 after_read 钩子
	err = m.AfterRead()
	return data, err
}

// NewQuery 创建查询构造器
func (m *BaseModel) NewQuery() (db.QueryInterface, error) {
	query, err := db.Table(m.TableName(), m.connection)
	if err != nil {
		return nil, err
	}

	// 如果启用软删除，自动添加条件
	if m.softDeletes {
		query = query.WhereNull(m.deletedAt)
	}

	return query, nil
}

// 事件钩子方法（可被子类重写）

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

// AfterRead 读取后钩子
func (m *BaseModel) AfterRead() error {
	return nil
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

// ToMap 转换为map
func (m *BaseModel) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range m.attributes {
		result[key] = value
	}
	return result
}

// SetRelation 设置关联数据
func (m *BaseModel) SetRelation(name string, value interface{}) {
	m.relations[name] = value
}

// GetRelation 获取关联数据
func (m *BaseModel) GetRelation(name string) interface{} {
	return m.relations[name]
}

// GetRelations 获取所有关联数据
func (m *BaseModel) GetRelations() map[string]interface{} {
	return m.relations
}

// HasRelation 检查是否存在关联
func (m *BaseModel) HasRelation(name string) bool {
	_, exists := m.relations[name]
	return exists
}

// String 字符串表示
func (m *BaseModel) String() string {
	var parts []string
	for key, value := range m.attributes {
		parts = append(parts, fmt.Sprintf("%s: %v", key, value))
	}
	return fmt.Sprintf("%s{%s}", m.TableName(), strings.Join(parts, ", "))
}

// ===== 查询构建器便捷方法 =====

// GetQueryBuilder 获取查询构建器（公共方法）
func (m *BaseModel) GetQueryBuilder() db.QueryInterface {
	return m.getQueryBuilder()
}

// getQueryBuilder 获取或创建内置查询构建器
func (m *BaseModel) getQueryBuilder() db.QueryInterface {
	if m.queryBuilder == nil {
		query, err := db.Table(m.TableName(), m.connection)
		if err != nil {
			// 如果创建失败，返回 nil，调用方需要处理错误
			return nil
		}

		// 如果启用软删除，自动添加条件
		if m.softDeletes {
			query = query.WhereNull(m.deletedAt)
		}

		m.queryBuilder = query
	}
	return m.queryBuilder
}

// resetQueryBuilder 重置查询构建器（用于新查询）
func (m *BaseModel) resetQueryBuilder() {
	m.queryBuilder = nil
}

// Where 添加WHERE条件 - 返回自身便于链式调用
// 支持两种调用方式:
// 1. Where(field, operator, value) - 传统三参数方式
// 2. Where(condition, args...) - 参数化查询方式
func (m *BaseModel) Where(args ...interface{}) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.Where(args...)
	}
	return m
}

// OrWhere 添加OR WHERE条件 - 返回自身便于链式调用
// 支持两种调用方式:
// 1. OrWhere(field, operator, value) - 传统三参数方式
// 2. OrWhere(condition, args...) - 参数化查询方式
func (m *BaseModel) OrWhere(args ...interface{}) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.OrWhere(args...)
	}
	return m
}

// WhereIn 添加WHERE IN条件
func (m *BaseModel) WhereIn(field string, values []interface{}) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.WhereIn(field, values)
	}
	return m
}

// WhereNull 添加WHERE NULL条件
func (m *BaseModel) WhereNull(field string) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.WhereNull(field)
	}
	return m
}

// WhereNotNull 添加WHERE NOT NULL条件
func (m *BaseModel) WhereNotNull(field string) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.WhereNotNull(field)
	}
	return m
}

// WhereBetween 添加WHERE BETWEEN条件
func (m *BaseModel) WhereBetween(field string, values []interface{}) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.WhereBetween(field, values)
	}
	return m
}

// WhereNotBetween 添加WHERE NOT BETWEEN条件
func (m *BaseModel) WhereNotBetween(field string, values []interface{}) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.WhereNotBetween(field, values)
	}
	return m
}

// WhereExists 添加WHERE EXISTS条件
func (m *BaseModel) WhereExists(subQuery interface{}) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.WhereExists(subQuery)
	}
	return m
}

// WhereNotExists 添加WHERE NOT EXISTS条件
func (m *BaseModel) WhereNotExists(subQuery interface{}) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.WhereNotExists(subQuery)
	}
	return m
}

// OrderRand 随机排序
func (m *BaseModel) OrderRand() *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.OrderRand()
	}
	return m
}

// OrderField 按字段值排序
func (m *BaseModel) OrderField(field string, values []interface{}, direction string) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.OrderField(field, values, direction)
	}
	return m
}

// FieldRaw 添加原生字段表达式
func (m *BaseModel) FieldRaw(raw string, bindings ...interface{}) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.FieldRaw(raw, bindings...)
	}
	return m
}

// WhereRaw 添加原生WHERE条件
func (m *BaseModel) WhereRaw(raw string, bindings ...interface{}) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.WhereRaw(raw, bindings...)
	}
	return m
}

// OrderBy 添加排序
func (m *BaseModel) OrderBy(field string, direction string) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.OrderBy(field, direction)
	}
	return m
}

// OrderByRaw 原生ORDER BY - 适配db.OrderByRaw
func (m *BaseModel) OrderByRaw(raw string, bindings ...interface{}) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.OrderByRaw(raw, bindings...)
	}
	return m
}

// Limit 限制结果数量
func (m *BaseModel) Limit(limit int) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.Limit(limit)
	}
	return m
}

// Offset 设置偏移量
func (m *BaseModel) Offset(offset int) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.Offset(offset)
	}
	return m
}

// Select 指定查询字段
func (m *BaseModel) Select(fields ...string) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.Select(fields...)
	}
	return m
}

// SelectRaw 原生SELECT字段 - 适配db.SelectRaw
func (m *BaseModel) SelectRaw(raw string, bindings ...interface{}) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.SelectRaw(raw, bindings...)
	}
	return m
}

// Distinct 去重查询 - 适配db.Distinct
func (m *BaseModel) Distinct() *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.Distinct()
	}
	return m
}

// GroupBy 添加分组
func (m *BaseModel) GroupBy(fields ...string) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.GroupBy(fields...)
	}
	return m
}

// Having 添加HAVING条件
func (m *BaseModel) Having(field string, operator string, value interface{}) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.Having(field, operator, value)
	}
	return m
}

// ===== JOIN查询方法 - 适配db包的Join功能 =====
// 注意：JOIN操作自动基于当前模型表，无需手动指定主表

// Join 内连接 - 适配db.Join
// first/second参数中如果不包含表名，会自动使用当前模型表名
func (m *BaseModel) Join(table string, first string, operator string, second string) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		// 智能处理字段名：如果字段名不包含表名，自动添加当前模型表名
		first = m.qualifyColumn(first)
		second = m.qualifyColumn(second)
		m.queryBuilder = query.Join(table, first, operator, second)
	}
	return m
}

// LeftJoin 左连接 - 适配db.LeftJoin
func (m *BaseModel) LeftJoin(table string, first string, operator string, second string) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		first = m.qualifyColumn(first)
		second = m.qualifyColumn(second)
		m.queryBuilder = query.LeftJoin(table, first, operator, second)
	}
	return m
}

// RightJoin 右连接 - 适配db.RightJoin
func (m *BaseModel) RightJoin(table string, first string, operator string, second string) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		first = m.qualifyColumn(first)
		second = m.qualifyColumn(second)
		m.queryBuilder = query.RightJoin(table, first, operator, second)
	}
	return m
}

// InnerJoin 内连接 - 适配db.InnerJoin
func (m *BaseModel) InnerJoin(table string, first string, operator string, second string) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		first = m.qualifyColumn(first)
		second = m.qualifyColumn(second)
		m.queryBuilder = query.InnerJoin(table, first, operator, second)
	}
	return m
}

// qualifyColumn 智能处理列名：如果不包含表名则添加当前模型表名
func (m *BaseModel) qualifyColumn(column string) string {
	// 如果列名已经包含表名（包含.），则直接返回
	if strings.Contains(column, ".") {
		return column
	}
	// 否则添加当前模型表名
	return m.TableName() + "." + column
}

// All 获取所有记录
func (m *BaseModel) All() ([]map[string]interface{}, error) {
	query := m.getQueryBuilder()
	if query == nil {
		return nil, fmt.Errorf("failed to create query builder")
	}

	results, err := query.Get()
	m.resetQueryBuilder() // 执行完成后重置查询构建器
	return results, err
}

// Get 获取所有记录 - All的别名
func (m *BaseModel) Get() ([]map[string]interface{}, error) {
	return m.All()
}

// First 获取第一条记录并填充到当前模型
// 如果传入指针，也会填充到指针指向的对象
// 返回原始的map数据
func (m *BaseModel) First(dest ...interface{}) (map[string]interface{}, error) {
	query := m.getQueryBuilder()
	if query == nil {
		return nil, fmt.Errorf("failed to create query builder")
	}

	result, err := query.First()
	m.resetQueryBuilder() // 执行完成后重置查询构建器

	if err != nil {
		return nil, err
	}

	// 填充当前模型属性
	m.Fill(result)
	m.exists = true
	m.isNew = false

	// 如果传入了指针，也填充到指针指向的对象
	if len(dest) > 0 && dest[0] != nil {
		err = m.LoadModel(dest[0], result)
		if err != nil {
			return result, fmt.Errorf("failed to load model: %w", err)
		}
	}

	return result, nil
}

// TakeFirst 链式查询后获取第一条记录并填充到当前模型
func (m *BaseModel) TakeFirst(dest ...interface{}) (map[string]interface{}, error) {
	return m.First(dest...)
}

// FirstOrCreate 查找第一条记录，如果不存在则创建
func (m *BaseModel) FirstOrCreate(attributes map[string]interface{}) error {
	// 先尝试查找
	query := m.getQueryBuilder()
	if query == nil {
		return fmt.Errorf("failed to create query builder")
	}

	result, err := query.First()
	m.resetQueryBuilder()

	if err == nil {
		// 找到了，填充模型
		m.Fill(result)
		m.exists = true
		m.isNew = false
		return nil
	}

	// 没找到，创建新记录
	id, err := m.Create(attributes)
	if err != nil {
		return err
	}

	// 填充模型
	m.Fill(attributes)
	m.SetAttribute(m.PrimaryKey(), id)
	m.exists = true
	m.isNew = false

	return nil
}

// FirstOrNew 查找第一条记录，如果不存在则创建新模型实例（不保存到数据库）
func (m *BaseModel) FirstOrNew(attributes map[string]interface{}) error {
	// 先尝试查找
	query := m.getQueryBuilder()
	if query == nil {
		return fmt.Errorf("failed to create query builder")
	}

	result, err := query.First()
	m.resetQueryBuilder()

	if err == nil {
		// 找到了，填充模型
		m.Fill(result)
		m.exists = true
		m.isNew = false
		return nil
	}

	// 没找到，填充新属性但不保存
	m.Fill(attributes)
	m.exists = false
	m.isNew = true

	return nil
}

// Count 统计记录数
func (m *BaseModel) Count() (int64, error) {
	query := m.getQueryBuilder()
	if query == nil {
		return 0, fmt.Errorf("failed to create query builder")
	}

	count, err := query.Count()
	m.resetQueryBuilder() // 执行完成后重置查询构建器
	return count, err
}

// HasRecords 检查是否存在记录
func (m *BaseModel) HasRecords() (bool, error) {
	query := m.getQueryBuilder()
	if query == nil {
		return false, fmt.Errorf("failed to create query builder")
	}

	exists, err := query.Exists()
	m.resetQueryBuilder() // 执行完成后重置查询构建器
	return exists, err
}

// CheckExists 检查查询条件是否有匹配记录 - 适配db.Exists
// 这个方法与Exists()不同，Exists()检查模型实例是否存在于数据库中
func (m *BaseModel) CheckExists() (bool, error) {
	query := m.getQueryBuilder()
	if query == nil {
		return false, fmt.Errorf("failed to create query builder")
	}

	exists, err := query.Exists()
	m.resetQueryBuilder() // 执行完成后重置查询构建器
	return exists, err
}

// Paginate 分页查询
func (m *BaseModel) Paginate(page, perPage int) (interface{}, error) {
	query := m.getQueryBuilder()
	if query == nil {
		return nil, fmt.Errorf("failed to create query builder")
	}

	result, err := query.Paginate(page, perPage)
	m.resetQueryBuilder() // 执行完成后重置查询构建器
	return result, err
}

// ToSQL 获取SQL语句（不执行）
func (m *BaseModel) ToSQL() (string, []interface{}, error) {
	query := m.getQueryBuilder()
	if query == nil {
		return "", nil, fmt.Errorf("failed to create query builder")
	}

	sql, bindings, err := query.ToSQL()
	// 注意：ToSQL 不重置查询构建器，因为它不执行查询
	return sql, bindings, err
}

// Clone 克隆查询构建器 - 适配db.Clone
func (m *BaseModel) Clone() *BaseModel {
	query := m.getQueryBuilder()
	if query == nil {
		return m
	}

	// 创建一个新的模型实例
	newModel := &BaseModel{
		connection:   m.connection,
		tableName:    m.tableName,
		primaryKeys:  m.primaryKeys,
		attributes:   make(map[string]interface{}),
		original:     make(map[string]interface{}),
		relations:    make(map[string]interface{}),
		isNew:        m.isNew,
		exists:       m.exists,
		timestamps:   m.timestamps,
		createdAt:    m.createdAt,
		updatedAt:    m.updatedAt,
		softDeletes:  m.softDeletes,
		deletedAt:    m.deletedAt,
		queryBuilder: query.Clone(), // 克隆查询构建器
	}

	// 复制属性
	for k, v := range m.attributes {
		newModel.attributes[k] = v
	}
	for k, v := range m.original {
		newModel.original[k] = v
	}
	for k, v := range m.relations {
		newModel.relations[k] = v
	}

	return newModel
}

// Create 创建记录
func (m *BaseModel) Create(data map[string]interface{}) (int64, error) {
	query, err := db.Table(m.TableName(), m.connection)
	if err != nil {
		return 0, err
	}

	// 添加时间戳
	if m.timestamps {
		now := time.Now()
		data[m.createdAt] = now
		data[m.updatedAt] = now
	}

	return query.Insert(data)
}

// Insert 插入单条记录 - 对db.Insert的直接封装
func (m *BaseModel) Insert(data map[string]interface{}) (int64, error) {
	query, err := db.Table(m.TableName(), m.connection)
	if err != nil {
		return 0, err
	}

	// 添加时间戳
	if m.timestamps {
		now := time.Now()
		data[m.createdAt] = now
		data[m.updatedAt] = now
	}

	return query.Insert(data)
}

// InsertBatch 批量插入记录 - 适配db.InsertBatch
func (m *BaseModel) InsertBatch(data []map[string]interface{}) (int64, error) {
	query, err := db.Table(m.TableName(), m.connection)
	if err != nil {
		return 0, err
	}

	// 为每条记录添加时间戳
	if m.timestamps {
		now := time.Now()
		for i := range data {
			data[i][m.createdAt] = now
			data[i][m.updatedAt] = now
		}
	}

	return query.InsertBatch(data)
}

// GetKey 获取主键值（单主键返回值，复合主键返回map）
func (m *BaseModel) GetKey() interface{} {
	if m.HasCompositePrimaryKey() {
		return m.GetPrimaryKeyValues()
	}
	return m.GetAttribute(m.PrimaryKey())
}

// ===== 静态方法（需要通过具体模型实例调用） =====

// FindOrFail 根据主键查找记录，找不到则返回错误
func (m *BaseModel) FindOrFail(id interface{}) error {
	_, err := m.Find(id)
	if err != nil {
		return fmt.Errorf("model not found with id: %v", id)
	}
	return nil
}

// FirstOrFail 获取第一条记录，找不到则返回错误
func (m *BaseModel) FirstOrFail() error {
	_, err := m.First()
	if err != nil {
		return fmt.Errorf("no records found")
	}
	return nil
}

// UpdateOrCreate 更新或创建记录
func (m *BaseModel) UpdateOrCreate(conditions, values map[string]interface{}) error {
	query, err := m.NewQuery()
	if err != nil {
		return err
	}

	// 添加查询条件
	for field, value := range conditions {
		query = query.Where(field, "=", value)
	}

	// 检查是否存在
	exists, err := query.Exists()
	if err != nil {
		return err
	}

	if exists {
		// 更新记录
		if m.timestamps {
			values[m.updatedAt] = time.Now()
		}
		_, err = query.Update(values)
		return err
	} else {
		// 创建记录
		mergedData := make(map[string]interface{})
		for k, v := range conditions {
			mergedData[k] = v
		}
		for k, v := range values {
			mergedData[k] = v
		}

		if m.timestamps {
			now := time.Now()
			mergedData[m.createdAt] = now
			mergedData[m.updatedAt] = now
		}

		insertQuery, err := db.Table(m.TableName(), m.connection)
		if err != nil {
			return err
		}

		id, err := insertQuery.Insert(mergedData)
		if err != nil {
			return err
		}

		// 设置主键值并填充模型
		m.Fill(mergedData)
		m.SetAttribute(m.PrimaryKey(), id)
		m.isNew = false
		m.exists = true

		return nil
	}
}

// Chunk 分块处理大量数据
func (m *BaseModel) Chunk(size int, callback func([]map[string]interface{}) error) error {
	offset := 0

	for {
		query, err := m.NewQuery()
		if err != nil {
			return err
		}

		results, err := query.Limit(size).Offset(offset).Get()
		if err != nil {
			return err
		}

		if len(results) == 0 {
			break
		}

		if err := callback(results); err != nil {
			return err
		}

		offset += size

		// 如果结果数量小于分块大小，说明已经是最后一批
		if len(results) < size {
			break
		}
	}

	return nil
}

// LoadModel 将map数据填充到指针指向的结构体
func (m *BaseModel) LoadModel(dest interface{}, result map[string]interface{}) error {
	if result == nil {
		return fmt.Errorf("no data to load")
	}

	// 使用反射填充目标模型
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("destination must be a pointer")
	}

	destValue = destValue.Elem()
	destType := destValue.Type()

	// 如果目标是BaseModel或包含BaseModel的结构体，填充BaseModel字段
	if destType.Name() == "BaseModel" {
		baseModel := dest.(*BaseModel)
		// 确保BaseModel的map已初始化
		if baseModel.attributes == nil {
			baseModel.attributes = make(map[string]interface{})
		}
		if baseModel.original == nil {
			baseModel.original = make(map[string]interface{})
		}
		if baseModel.relations == nil {
			baseModel.relations = make(map[string]interface{})
		}
		baseModel.Fill(result)
		baseModel.exists = true
		baseModel.isNew = false
	} else if baseModelField := destValue.FieldByName("BaseModel"); baseModelField.IsValid() {
		// 获取BaseModel字段
		baseModel := baseModelField.Addr().Interface().(*BaseModel)
		// 确保BaseModel的map已初始化
		if baseModel.attributes == nil {
			baseModel.attributes = make(map[string]interface{})
		}
		if baseModel.original == nil {
			baseModel.original = make(map[string]interface{})
		}
		if baseModel.relations == nil {
			baseModel.relations = make(map[string]interface{})
		}
		baseModel.Fill(result)
		baseModel.exists = true
		baseModel.isNew = false
	}

	// 填充结构体字段
	for i := 0; i < destType.NumField(); i++ {
		field := destType.Field(i)

		// 跳过BaseModel字段，已经在上面处理了
		if field.Name == "BaseModel" {
			continue
		}

		dbTag := field.Tag.Get("db")
		jsonTag := field.Tag.Get("json")

		var fieldName string
		if dbTag != "" && dbTag != "-" {
			fieldName = dbTag
		} else if jsonTag != "" && jsonTag != "-" {
			fieldName = jsonTag
		} else {
			fieldName = strings.ToLower(field.Name)
		}

		if value, exists := result[fieldName]; exists && destValue.Field(i).CanSet() {
			fieldValue := destValue.Field(i)
			if fieldValue.Kind() == reflect.Ptr {
				if value != nil {
					// 为指针字段分配内存
					newValue := reflect.New(fieldValue.Type().Elem())
					if newValue.Elem().Type() == reflect.TypeOf(value) {
						newValue.Elem().Set(reflect.ValueOf(value))
						fieldValue.Set(newValue)
					}
				}
			} else {
				if value != nil && reflect.TypeOf(value).AssignableTo(fieldValue.Type()) {
					fieldValue.Set(reflect.ValueOf(value))
				}
			}
		}
	}

	return nil
}

// Update 更新记录 - 支持两种调用方式
// 1. Update() - 更新当前模型的变更字段
// 2. Update(data) - 根据之前的Where条件批量更新
func (m *BaseModel) Update(data ...map[string]interface{}) (interface{}, error) {
	// 如果没有传入data参数，更新当前模型
	if len(data) == 0 {
		return nil, m.updateCurrentModel()
	}

	// 否则执行批量更新
	return m.updateBatch(data[0])
}

// updateBatch 批量更新记录 - 适配db.Update
func (m *BaseModel) updateBatch(data map[string]interface{}) (int64, error) {
	query := m.getQueryBuilder()
	if query == nil {
		return 0, fmt.Errorf("failed to create query builder")
	}

	// 添加更新时间戳
	if m.timestamps && m.updatedAt != "" {
		data[m.updatedAt] = time.Now()
	}

	result, err := query.Update(data)
	m.resetQueryBuilder() // 执行后重置查询构建器
	return result, err
}

// updateCurrentModel 更新当前模型实例
func (m *BaseModel) updateCurrentModel() error {
	// 检查是否有变更
	dirty := m.GetDirty()
	if len(dirty) == 0 {
		return nil // 没有变更，不需要更新
	}

	// 创建查询构造器
	query, err := db.Table(m.TableName(), m.connection)
	if err != nil {
		return err
	}

	// 构建主键条件
	query = m.buildPrimaryKeyConditions(query)
	if query == nil {
		return fmt.Errorf("primary key values are required for update")
	}

	// 执行更新
	_, err = query.Update(dirty)
	if err != nil {
		return err
	}

	// 同步原始数据
	m.syncOriginal()

	return m.AfterUpdate()
}

// buildPrimaryKeyConditions 构建主键查询条件（支持复合主键）
func (m *BaseModel) buildPrimaryKeyConditions(query db.QueryInterface) db.QueryInterface {
	hasValidPrimaryKey := false

	for _, pkField := range m.primaryKeys {
		pkValue := m.GetAttribute(pkField)
		if pkValue == nil {
			return nil // 主键值不能为空
		}
		query = query.Where(pkField, "=", pkValue)
		hasValidPrimaryKey = true
	}

	if !hasValidPrimaryKey {
		return nil
	}

	return query
}

// AutoMigrate 自动迁移模型到数据库
func (m *BaseModel) AutoMigrate() error {
	// 获取数据库连接
	manager := db.DefaultManager()
	conn, err := manager.Connection(m.connection)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// 检测模型结构
	if err := m.detectModelStructure(); err != nil {
		return fmt.Errorf("failed to detect model structure: %w", err)
	}

	// 获取表名
	tableName := m.TableName()
	if tableName == "" {
		return fmt.Errorf("table name is required for auto migration")
	}

	// 检查表是否存在
	exists, err := m.tableExists(conn, tableName)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if !exists {
		// 创建新表
		return m.createTable(conn, tableName)
	}

	// 表已存在，检查是否需要更新结构
	return m.updateTableStructure(conn, tableName)
}

// detectModelStructure 检测模型结构
func (m *BaseModel) detectModelStructure() error {
	// 这个方法将在调用 DetectConfigFromStruct 时由外部实现
	// 这里只是确保必要的配置已经设置
	if m.tableName == "" {
		return fmt.Errorf("table name must be set before auto migration")
	}
	return nil
}

// tableExists 检查表是否存在
func (m *BaseModel) tableExists(conn db.ConnectionInterface, tableName string) (bool, error) {
	driver := conn.GetDriver()
	var query string
	var args []interface{}

	switch driver {
	case "mysql":
		query = "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?"
		args = []interface{}{tableName}
	case "postgres", "postgresql":
		query = "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = ?"
		args = []interface{}{tableName}
	case "sqlite", "sqlite3":
		query = "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name = ?"
		args = []interface{}{tableName}
	default:
		return false, fmt.Errorf("unsupported database driver: %s", driver)
	}

	row := conn.QueryRow(query, args...)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("failed to check table existence: %w", err)
	}

	return count > 0, nil
}

// createTable 创建新表
func (m *BaseModel) createTable(conn db.ConnectionInterface, tableName string) error {
	// 获取模型的结构体信息
	modelStruct, err := m.getModelStruct()
	if err != nil {
		return fmt.Errorf("failed to get model structure: %w", err)
	}

	// 创建表定义
	table := &migration.Table{
		Name:    tableName,
		Columns: make([]*migration.Column, 0),
		Indexes: make([]*migration.Index, 0),
	}

	// 设置数据库引擎和字符集（MySQL）
	driver := conn.GetDriver()
	if driver == "mysql" {
		table.Engine = "InnoDB"
		table.Charset = "utf8mb4"
	}

	// 解析字段
	if err := m.parseFieldsForMigration(modelStruct, table); err != nil {
		return fmt.Errorf("failed to parse model fields: %w", err)
	}

	// 添加自动索引
	m.addAutoIndexes(table)

	// 使用 SchemaBuilder 创建表
	schemaBuilder := migration.NewSchemaBuilder(conn)
	return schemaBuilder.CreateTable(table)
}

// addAutoIndexes 为表添加自动索引
func (m *BaseModel) addAutoIndexes(table *migration.Table) {
	// 为有 unique 标签的字段添加唯一索引
	for _, column := range table.Columns {
		if column.Unique && !column.PrimaryKey {
			index := &migration.Index{
				Name:    fmt.Sprintf("idx_%s_%s_unique", table.Name, column.Name),
				Columns: []string{column.Name},
				Unique:  true,
			}
			table.Indexes = append(table.Indexes, index)
		}

		// 解析 comment 中的索引信息
		if strings.Contains(column.Comment, "INDEX:") {
			parts := strings.Split(column.Comment, "INDEX:")
			if len(parts) > 1 {
				indexName := strings.TrimSpace(parts[1])
				// 移除索引标记，保留原始注释
				column.Comment = strings.TrimSpace(parts[0])

				// 如果没有指定索引名，生成默认名称
				if indexName == "" || indexName == "true" {
					indexName = fmt.Sprintf("idx_%s_%s", table.Name, column.Name)
				}

				index := &migration.Index{
					Name:    indexName,
					Columns: []string{column.Name},
					Unique:  false,
				}
				table.Indexes = append(table.Indexes, index)
			}
		}
	}

	// 为外键字段添加索引（约定：以 _id 结尾的字段）
	for _, column := range table.Columns {
		if strings.HasSuffix(column.Name, "_id") && !column.PrimaryKey {
			// 检查是否已有索引
			hasIndex := false
			for _, index := range table.Indexes {
				for _, indexCol := range index.Columns {
					if indexCol == column.Name {
						hasIndex = true
						break
					}
				}
				if hasIndex {
					break
				}
			}

			if !hasIndex {
				index := &migration.Index{
					Name:    fmt.Sprintf("idx_%s_%s", table.Name, column.Name),
					Columns: []string{column.Name},
				}
				table.Indexes = append(table.Indexes, index)
			}
		}
	}
}

// getModelStruct 获取模型的结构体信息
func (m *BaseModel) getModelStruct() (reflect.Type, error) {
	if m.modelType == nil {
		return nil, fmt.Errorf("model structure not available - call DetectConfigFromStruct first")
	}
	return m.modelType, nil
}

// HasModelStruct 检查是否已设置模型结构体信息
func (m *BaseModel) HasModelStruct() bool {
	return m.modelType != nil
}

// GetModelStructName 获取模型结构体名称（用于调试和测试）
func (m *BaseModel) GetModelStructName() string {
	if m.modelType == nil {
		return ""
	}
	return m.modelType.Name()
}

// parseFieldsForMigration 解析字段用于迁移
func (m *BaseModel) parseFieldsForMigration(modelType reflect.Type, table *migration.Table) error {
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)

		// 跳过 BaseModel 字段
		if field.Name == "BaseModel" {
			continue
		}

		// 创建列定义
		column, err := m.fieldToColumn(field)
		if err != nil {
			return fmt.Errorf("failed to convert field %s: %w", field.Name, err)
		}

		if column != nil {
			table.Columns = append(table.Columns, column)
		}
	}

	return nil
}

// fieldToColumn 将结构体字段转换为数据库列定义
func (m *BaseModel) fieldToColumn(field reflect.StructField) (*migration.Column, error) {
	// 解析 db 标签
	dbTag := field.Tag.Get("db")
	if dbTag == "-" {
		return nil, nil // 跳过不需要持久化的字段
	}

	columnName := dbTag
	if columnName == "" {
		// 如果没有 db 标签，使用字段名的小写形式
		columnName = strings.ToLower(field.Name)
	}

	// 创建列定义
	column := &migration.Column{
		Name: columnName,
	}

	// 映射 Go 类型到数据库类型
	if err := m.mapGoTypeToColumnType(field, column); err != nil {
		return nil, err
	}

	// 解析标签中的属性
	m.parseFieldTags(field, column)

	return column, nil
}

// mapGoTypeToColumnType 映射 Go 类型到数据库列类型
func (m *BaseModel) mapGoTypeToColumnType(field reflect.StructField, column *migration.Column) error {
	fieldType := field.Type

	// 处理指针类型
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
		// 指针类型默认可为空
		column.NotNull = false
	} else {
		// 非指针类型默认不为空
		column.NotNull = true
	}

	// 优先检查 type 标签的自定义类型
	if typeTag := field.Tag.Get("type"); typeTag != "" {
		return m.mapCustomType(typeTag, field, column)
	}

	// 特殊类型检查
	if fieldType == reflect.TypeOf(time.Time{}) {
		return m.mapTimeType(field, column)
	}
	if fieldType == reflect.TypeOf(DeletedTime{}) {
		column.Type = migration.ColumnTypeTimestamp
		column.NotNull = false // 软删除字段可为空
		return nil
	}

	// 基本类型映射
	switch fieldType.Kind() {
	case reflect.String:
		return m.mapStringType(field, column)

	case reflect.Int:
		column.Type = migration.ColumnTypeInt

	case reflect.Int8:
		column.Type = migration.ColumnTypeTinyInt

	case reflect.Int16:
		column.Type = migration.ColumnTypeSmallInt

	case reflect.Int32:
		column.Type = migration.ColumnTypeInt

	case reflect.Int64:
		// 检查是否为时间戳
		if field.Tag.Get("autoCreateTime") != "" || field.Tag.Get("autoUpdateTime") != "" {
			column.Type = migration.ColumnTypeBigInt
		} else {
			column.Type = migration.ColumnTypeBigInt
		}

	case reflect.Uint:
		column.Type = migration.ColumnTypeInt // 注意：无符号类型映射

	case reflect.Uint8:
		column.Type = migration.ColumnTypeTinyInt

	case reflect.Uint16:
		column.Type = migration.ColumnTypeSmallInt

	case reflect.Uint32:
		column.Type = migration.ColumnTypeInt

	case reflect.Uint64:
		column.Type = migration.ColumnTypeBigInt

	case reflect.Float32:
		return m.mapFloatType(field, column, true)

	case reflect.Float64:
		return m.mapFloatType(field, column, false)

	case reflect.Bool:
		column.Type = migration.ColumnTypeBoolean

	case reflect.Slice, reflect.Array:
		return m.mapSliceType(fieldType, column)

	case reflect.Map, reflect.Struct:
		// Map 和 复杂结构体使用 JSON
		column.Type = migration.ColumnTypeJSON

	case reflect.Interface:
		// interface{} 类型使用 JSON
		column.Type = migration.ColumnTypeJSON

	default:
		// 默认为文本类型
		column.Type = migration.ColumnTypeText
	}

	return nil
}

// mapCustomType 处理自定义类型标签
func (m *BaseModel) mapCustomType(typeTag string, field reflect.StructField, column *migration.Column) error {
	switch strings.ToLower(typeTag) {
	// 字符串类型
	case "varchar":
		column.Type = migration.ColumnTypeVarchar
		if sizeTag := field.Tag.Get("size"); sizeTag != "" {
			column.Length = parseInt(sizeTag)
		}
		if column.Length == 0 {
			column.Length = 255
		}
	case "char":
		column.Type = migration.ColumnTypeChar
		if sizeTag := field.Tag.Get("size"); sizeTag != "" {
			column.Length = parseInt(sizeTag)
		}
		if column.Length == 0 {
			column.Length = 1
		}
	case "text":
		column.Type = migration.ColumnTypeText
	case "longtext":
		column.Type = migration.ColumnTypeLongText

	// 数值类型
	case "tinyint":
		column.Type = migration.ColumnTypeTinyInt
	case "smallint":
		column.Type = migration.ColumnTypeSmallInt
	case "int", "integer":
		column.Type = migration.ColumnTypeInt
	case "bigint":
		column.Type = migration.ColumnTypeBigInt
	case "float":
		column.Type = migration.ColumnTypeFloat
	case "double":
		column.Type = migration.ColumnTypeDouble
	case "decimal", "numeric":
		column.Type = migration.ColumnTypeDecimal
		if precisionTag := field.Tag.Get("precision"); precisionTag != "" {
			column.Precision = parseInt(precisionTag)
		}
		if scaleTag := field.Tag.Get("scale"); scaleTag != "" {
			column.Scale = parseInt(scaleTag)
		}

	// 时间类型
	case "datetime":
		column.Type = migration.ColumnTypeDateTime
	case "timestamp":
		column.Type = migration.ColumnTypeTimestamp
	case "date":
		column.Type = migration.ColumnTypeDate
	case "time":
		column.Type = migration.ColumnTypeTime

	// 其他类型
	case "boolean", "bool":
		column.Type = migration.ColumnTypeBoolean
	case "blob":
		column.Type = migration.ColumnTypeBlob
	case "json":
		column.Type = migration.ColumnTypeJSON

	default:
		return fmt.Errorf("unsupported custom type: %s", typeTag)
	}

	return nil
}

// mapStringType 处理字符串类型映射
func (m *BaseModel) mapStringType(field reflect.StructField, column *migration.Column) error {
	// 检查长度标签
	sizeTag := field.Tag.Get("size")
	if sizeTag != "" {
		size := parseInt(sizeTag)
		if size > 0 {
			if size <= 255 {
				column.Type = migration.ColumnTypeVarchar
				column.Length = size
			} else if size <= 65535 {
				column.Type = migration.ColumnTypeText
			} else {
				column.Type = migration.ColumnTypeLongText
			}
			return nil
		}
	}

	// 检查是否为固定长度
	if field.Tag.Get("fixed") == "true" {
		column.Type = migration.ColumnTypeChar
		if column.Length == 0 {
			column.Length = 255
		}
		return nil
	}

	// 默认 VARCHAR(255)
	column.Type = migration.ColumnTypeVarchar
	column.Length = 255
	return nil
}

// mapTimeType 处理时间类型映射
func (m *BaseModel) mapTimeType(field reflect.StructField, column *migration.Column) error {
	// 检查自动时间戳标签
	if field.Tag.Get("autoCreateTime") != "" || field.Tag.Get("autoUpdateTime") != "" {
		column.Type = migration.ColumnTypeTimestamp
		return nil
	}

	// 检查类型偏好
	if field.Tag.Get("timestamp") == "true" {
		column.Type = migration.ColumnTypeTimestamp
	} else {
		column.Type = migration.ColumnTypeDateTime
	}

	return nil
}

// mapFloatType 处理浮点类型映射
func (m *BaseModel) mapFloatType(field reflect.StructField, column *migration.Column, isFloat32 bool) error {
	// 检查是否指定为 DECIMAL
	if field.Tag.Get("decimal") == "true" {
		column.Type = migration.ColumnTypeDecimal

		if precisionTag := field.Tag.Get("precision"); precisionTag != "" {
			column.Precision = parseInt(precisionTag)
		} else {
			column.Precision = 10 // 默认精度
		}

		if scaleTag := field.Tag.Get("scale"); scaleTag != "" {
			column.Scale = parseInt(scaleTag)
		} else {
			column.Scale = 2 // 默认小数位
		}

		return nil
	}

	// 默认浮点类型
	if isFloat32 {
		column.Type = migration.ColumnTypeFloat
	} else {
		column.Type = migration.ColumnTypeDouble
	}

	return nil
}

// mapSliceType 处理切片类型映射
func (m *BaseModel) mapSliceType(fieldType reflect.Type, column *migration.Column) error {
	elemType := fieldType.Elem()

	// 检查元素类型
	switch elemType.Kind() {
	case reflect.Uint8:
		// []byte 映射为 BLOB
		column.Type = migration.ColumnTypeBlob
	case reflect.String:
		// []string 映射为 JSON
		column.Type = migration.ColumnTypeJSON
	default:
		// 其他切片类型都映射为 JSON
		column.Type = migration.ColumnTypeJSON
	}

	return nil
}

// parseFieldTags 解析字段标签
func (m *BaseModel) parseFieldTags(field reflect.StructField, column *migration.Column) {
	// 优先解析 torm 标签
	if tormTag := field.Tag.Get("torm"); tormTag != "" {
		m.parseTormTag(tormTag, column)
		return
	}

	// 向后兼容：检查传统标签
	m.parseLegacyTags(field, column)
}

// parseTormTag 解析 torm 标签
func (m *BaseModel) parseTormTag(tormTag string, column *migration.Column) {
	// 分割标签内容
	parts := strings.Split(tormTag, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// 检查是否是 key:value 格式
		if strings.Contains(part, ":") {
			keyValue := strings.SplitN(part, ":", 2)
			key := strings.TrimSpace(keyValue[0])
			value := strings.TrimSpace(keyValue[1])

			m.parseTormKeyValue(key, value, column)
		} else {
			// 简单标志位
			m.parseTormFlag(part, column)
		}
	}
}

// parseTormFlag 解析 torm 标签中的标志位
func (m *BaseModel) parseTormFlag(flag string, column *migration.Column) {
	switch strings.ToLower(flag) {
	case "primary_key", "pk":
		column.PrimaryKey = true
		column.NotNull = true
	case "auto_increment", "autoincrement":
		column.AutoIncrement = true
		column.PrimaryKey = true
		column.NotNull = true
	case "unique":
		column.Unique = true
	case "nullable", "null":
		column.NotNull = false
	case "not_null", "notnull":
		column.NotNull = true
	case "auto_create_time", "autocreate":
		column.Default = "CURRENT_TIMESTAMP"
	case "auto_update_time", "autoupdate":
		column.Default = "CURRENT_TIMESTAMP"
		// TODO: 添加 ON UPDATE CURRENT_TIMESTAMP 支持
	case "index":
		m.markColumnForIndex(column, "")
	}
}

// parseTormKeyValue 解析 torm 标签中的 key:value 对
func (m *BaseModel) parseTormKeyValue(key, value string, column *migration.Column) {
	switch strings.ToLower(key) {
	case "type":
		// 直接设置类型，不使用 mapCustomType（避免查找单独标签的问题）
		m.setColumnType(value, column)
	case "size", "length":
		if size := parseInt(value); size > 0 {
			column.Length = size
		}
	case "precision":
		if precision := parseInt(value); precision > 0 {
			column.Precision = precision
		}
	case "scale":
		if scale := parseInt(value); scale > 0 {
			column.Scale = scale
		}
	case "default":
		m.parseDefaultValue(value, column)
	case "comment":
		column.Comment = value
	case "index":
		m.markColumnForIndex(column, value)
	}
}

// setColumnType 直接设置列类型（不依赖单独标签）
func (m *BaseModel) setColumnType(typeStr string, column *migration.Column) {
	switch strings.ToLower(typeStr) {
	// 字符串类型
	case "varchar":
		column.Type = migration.ColumnTypeVarchar
		if column.Length == 0 {
			column.Length = 255 // 默认长度
		}
	case "char":
		column.Type = migration.ColumnTypeChar
		if column.Length == 0 {
			column.Length = 1 // 默认长度
		}
	case "text":
		column.Type = migration.ColumnTypeText
	case "longtext":
		column.Type = migration.ColumnTypeLongText

	// 数值类型
	case "tinyint":
		column.Type = migration.ColumnTypeTinyInt
	case "smallint":
		column.Type = migration.ColumnTypeSmallInt
	case "int", "integer":
		column.Type = migration.ColumnTypeInt
	case "bigint":
		column.Type = migration.ColumnTypeBigInt
	case "float":
		column.Type = migration.ColumnTypeFloat
	case "double":
		column.Type = migration.ColumnTypeDouble
	case "decimal", "numeric":
		column.Type = migration.ColumnTypeDecimal
		// 默认精度和小数位
		if column.Precision == 0 {
			column.Precision = 10
		}
		if column.Scale == 0 {
			column.Scale = 2
		}

	// 时间类型
	case "datetime":
		column.Type = migration.ColumnTypeDateTime
	case "timestamp":
		column.Type = migration.ColumnTypeTimestamp
	case "date":
		column.Type = migration.ColumnTypeDate
	case "time":
		column.Type = migration.ColumnTypeTime

	// 其他类型
	case "boolean", "bool":
		column.Type = migration.ColumnTypeBoolean
	case "blob":
		column.Type = migration.ColumnTypeBlob
	case "json":
		column.Type = migration.ColumnTypeJSON
	}
}

// parseDefaultValue 解析默认值
func (m *BaseModel) parseDefaultValue(value string, column *migration.Column) {
	switch strings.ToLower(value) {
	case "null":
		column.Default = nil
	case "current_timestamp", "now()":
		column.Default = "CURRENT_TIMESTAMP"
	case "true":
		column.Default = true
	case "false":
		column.Default = false
	default:
		// 尝试解析为数字
		if intVal := parseInt(value); intVal != 0 || value == "0" {
			column.Default = intVal
		} else {
			column.Default = value
		}
	}
}

// markColumnForIndex 标记列需要创建索引
func (m *BaseModel) markColumnForIndex(column *migration.Column, indexName string) {
	if indexName == "" || indexName == "true" {
		indexName = "auto"
	}

	if column.Comment == "" {
		column.Comment = "INDEX:" + indexName
	} else {
		column.Comment += " INDEX:" + indexName
	}
}

// parseLegacyTags 解析传统标签（向后兼容）
func (m *BaseModel) parseLegacyTags(field reflect.StructField, column *migration.Column) {
	// 检查主键标签
	if field.Tag.Get("primaryKey") == "true" || field.Tag.Get("pk") != "" {
		column.PrimaryKey = true
		column.NotNull = true
	}

	// 检查唯一性约束
	if field.Tag.Get("unique") == "true" {
		column.Unique = true
	}

	// 检查自增标签
	if field.Tag.Get("autoIncrement") == "true" || field.Tag.Get("auto_increment") == "true" {
		column.AutoIncrement = true
		column.PrimaryKey = true
		column.NotNull = true
	}

	// 检查默认值
	if defaultValue := field.Tag.Get("default"); defaultValue != "" {
		m.parseDefaultValue(defaultValue, column)
	}

	// 检查注释
	if comment := field.Tag.Get("comment"); comment != "" {
		column.Comment = comment
	}

	// 检查非空约束
	if field.Tag.Get("not_null") == "true" {
		column.NotNull = true
	} else if field.Tag.Get("nullable") == "true" {
		column.NotNull = false
	}

	// 检查时间戳字段
	if field.Tag.Get("autoCreateTime") != "" {
		column.Default = "CURRENT_TIMESTAMP"
	}
	if field.Tag.Get("autoUpdateTime") != "" {
		column.Default = "CURRENT_TIMESTAMP"
	}

	// 检查索引标签
	if field.Tag.Get("index") != "" {
		m.markColumnForIndex(column, field.Tag.Get("index"))
	}
}

// parseInt 解析整数字符串
func parseInt(s string) int {
	result := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			result = result*10 + int(r-'0')
		} else {
			return 0
		}
	}
	return result
}

// updateTableStructure 更新表结构
func (m *BaseModel) updateTableStructure(conn db.ConnectionInterface, tableName string) error {
	// 检查现有表结构与模型的差异，并执行必要的ALTER TABLE操作
	// 暂时返回nil，表示不进行结构更新
	return nil
}
