package model

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/zhoudm1743/torm/db"
)

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

// TableName 获取表名
func (m *BaseModel) TableName() string {
	if m.tableName != "" {
		return m.tableName
	}
	// 如果没有设置表名，使用结构体名的复数形式
	return m.tableName
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

// Delete 删除记录
func (m *BaseModel) Delete() error {
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

// Find 根据主键查找记录
// 如果传入指针，也会填充到指针指向的对象
// 返回原始的map数据
func (m *BaseModel) Find(id interface{}, dest ...interface{}) (map[string]interface{}, error) {
	query, err := db.Table(m.TableName(), m.connection)
	if err != nil {
		return nil, err
	}

	data, err := query.Where(m.PrimaryKey(), "=", id).First()
	if err != nil {
		return nil, err
	}

	// 填充当前模型属性
	m.attributes = data
	m.syncOriginal()
	m.isNew = false
	m.exists = true

	// 如果传入了指针，也填充到指针指向的对象
	if len(dest) > 0 && dest[0] != nil {
		err = m.LoadModel(dest[0], data)
		if err != nil {
			return data, fmt.Errorf("failed to load model: %w", err)
		}
	}

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
func (m *BaseModel) Where(field string, operator string, value interface{}) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.Where(field, operator, value)
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

// WhereNotIn 添加WHERE NOT IN条件
func (m *BaseModel) WhereNotIn(field string, values []interface{}) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.WhereNotIn(field, values)
	}
	return m
}

// WhereBetween 添加BETWEEN条件
func (m *BaseModel) WhereBetween(field string, start, end interface{}) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.WhereBetween(field, start, end)
	}
	return m
}

// WhereNull 添加IS NULL条件
func (m *BaseModel) WhereNull(field string) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.WhereNull(field)
	}
	return m
}

// WhereNotNull 添加IS NOT NULL条件
func (m *BaseModel) WhereNotNull(field string) *BaseModel {
	query := m.getQueryBuilder()
	if query != nil {
		m.queryBuilder = query.WhereNotNull(field)
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

func (m *BaseModel) Update() error {
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
