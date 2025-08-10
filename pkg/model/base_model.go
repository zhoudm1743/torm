package model

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"torm/pkg/db"
)

// BaseModel 基础模型
type BaseModel struct {
	// 数据库连接名
	connection string
	// 表名
	table string
	// 主键字段名
	primaryKey string
	// 模型属性
	attributes map[string]interface{}
	// 原始属性（用于检测变更）
	original map[string]interface{}
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
}

// NewBaseModel 创建基础模型实例
func NewBaseModel() *BaseModel {
	return &BaseModel{
		connection:  "default",
		primaryKey:  "id",
		attributes:  make(map[string]interface{}),
		original:    make(map[string]interface{}),
		isNew:       true,
		exists:      false,
		timestamps:  true,
		createdAt:   "created_at",
		updatedAt:   "updated_at",
		softDeletes: false,
		deletedAt:   "deleted_at",
	}
}

// TableName 获取表名
func (m *BaseModel) TableName() string {
	if m.table != "" {
		return m.table
	}
	// 如果没有设置表名，使用结构体名的复数形式
	return m.table
}

// SetTable 设置表名
func (m *BaseModel) SetTable(table string) *BaseModel {
	m.table = table
	return m
}

// PrimaryKey 获取主键字段名
func (m *BaseModel) PrimaryKey() string {
	return m.primaryKey
}

// SetPrimaryKey 设置主键字段名
func (m *BaseModel) SetPrimaryKey(key string) *BaseModel {
	m.primaryKey = key
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
	m.attributes[key] = value
}

// GetAttributes 获取所有属性
func (m *BaseModel) GetAttributes() map[string]interface{} {
	return m.attributes
}

// SetAttributes 设置多个属性
func (m *BaseModel) SetAttributes(attributes map[string]interface{}) {
	for key, value := range attributes {
		m.attributes[key] = value
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
func (m *BaseModel) Save(ctx context.Context) error {
	if m.isNew {
		return m.create(ctx)
	}
	return m.update(ctx)
}

// create 创建新记录
func (m *BaseModel) create(ctx context.Context) error {
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
	id, err := query.Insert(ctx, m.attributes)
	if err != nil {
		return err
	}

	// 设置主键值
	m.SetAttribute(m.primaryKey, id)

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
func (m *BaseModel) update(ctx context.Context) error {
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
	pkValue := m.GetAttribute(m.primaryKey)
	if pkValue == nil {
		return fmt.Errorf("primary key value is required for update")
	}

	_, err = query.Where(m.primaryKey, "=", pkValue).Update(ctx, dirty)
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
func (m *BaseModel) Delete(ctx context.Context) error {
	if m.isNew {
		return fmt.Errorf("cannot delete unsaved model")
	}

	// 执行 before_delete 钩子
	if err := m.BeforeDelete(); err != nil {
		return err
	}

	pkValue := m.GetAttribute(m.primaryKey)
	if pkValue == nil {
		return fmt.Errorf("primary key value is required for delete")
	}

	// 获取查询构造器
	query, err := db.Table(m.TableName(), m.connection)
	if err != nil {
		return err
	}

	if m.softDeletes {
		// 软删除
		deleteData := map[string]interface{}{
			m.deletedAt: time.Now(),
		}
		_, err = query.Where(m.primaryKey, "=", pkValue).Update(ctx, deleteData)
		if err != nil {
			return err
		}
		m.SetAttribute(m.deletedAt, deleteData[m.deletedAt])
	} else {
		// 硬删除
		_, err = query.Where(m.primaryKey, "=", pkValue).Delete(ctx)
		if err != nil {
			return err
		}
		m.exists = false
	}

	// 执行 after_delete 钩子
	return m.AfterDelete()
}

// Reload 重新加载模型数据
func (m *BaseModel) Reload(ctx context.Context) error {
	if m.isNew {
		return fmt.Errorf("cannot reload unsaved model")
	}

	pkValue := m.GetAttribute(m.primaryKey)
	if pkValue == nil {
		return fmt.Errorf("primary key value is required for reload")
	}

	// 获取查询构造器
	query, err := db.Table(m.TableName(), m.connection)
	if err != nil {
		return err
	}

	// 查询数据
	data, err := query.Where(m.primaryKey, "=", pkValue).First(ctx)
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
func (m *BaseModel) Find(ctx context.Context, id interface{}) error {
	query, err := db.Table(m.TableName(), m.connection)
	if err != nil {
		return err
	}

	data, err := query.Where(m.primaryKey, "=", id).First(ctx)
	if err != nil {
		return err
	}

	// 填充属性
	m.attributes = data
	m.syncOriginal()
	m.isNew = false
	m.exists = true

	// 执行 after_read 钩子
	return m.AfterRead()
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

// String 字符串表示
func (m *BaseModel) String() string {
	var parts []string
	for key, value := range m.attributes {
		parts = append(parts, fmt.Sprintf("%s: %v", key, value))
	}
	return fmt.Sprintf("%s{%s}", m.TableName(), strings.Join(parts, ", "))
}
