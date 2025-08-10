package model

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"torm/pkg/db"
)

// RelationType 关联关系类型
type RelationType string

const (
	HasOneType     RelationType = "has_one"
	HasManyType    RelationType = "has_many"
	BelongsToType  RelationType = "belongs_to"
	ManyToManyType RelationType = "many_to_many"
)

// RelationInterface 关联关系接口
type RelationInterface interface {
	// 获取关联数据
	Get(ctx context.Context) (interface{}, error)
	// 获取关联查询构造器
	GetQuery() (db.QueryInterface, error)
	// 获取关联类型
	GetType() RelationType
	// 获取关联模型
	GetRelated() interface{}
	// 设置约束条件
	Where(field string, operator string, value interface{}) RelationInterface
	// 添加关联
	Associate(ctx context.Context, model interface{}) error
	// 移除关联
	Dissociate(ctx context.Context, model interface{}) error
}

// BaseRelation 基础关联关系
type BaseRelation struct {
	// 父模型实例
	parent interface{}
	// 关联模型类型
	related interface{}
	// 外键字段名
	foreignKey string
	// 本地键字段名
	localKey string
	// 关联关系类型
	relationType RelationType
	// 查询条件
	wheres []db.WhereClause
	// 数据库连接名
	connection string
}

// NewBaseRelation 创建基础关联关系
func NewBaseRelation(parent interface{}, related interface{}, foreignKey, localKey string, relationType RelationType) *BaseRelation {
	// 尝试从父模型获取连接名
	connectionName := "default"
	if baseModel, ok := parent.(*BaseModel); ok {
		if baseModel.connection != "" {
			connectionName = baseModel.connection
		}
	}

	return &BaseRelation{
		parent:       parent,
		related:      related,
		foreignKey:   foreignKey,
		localKey:     localKey,
		relationType: relationType,
		wheres:       make([]db.WhereClause, 0),
		connection:   connectionName,
	}
}

// GetType 获取关联类型
func (r *BaseRelation) GetType() RelationType {
	return r.relationType
}

// GetRelated 获取关联模型
func (r *BaseRelation) GetRelated() interface{} {
	return r.related
}

// Where 添加查询条件
func (r *BaseRelation) Where(field string, operator string, value interface{}) *BaseRelation {
	r.wheres = append(r.wheres, db.WhereClause{
		Type:     "and",
		Field:    field,
		Operator: operator,
		Value:    value,
	})
	return r
}

// getParentKey 获取父模型的键值
func (r *BaseRelation) getParentKey() interface{} {
	if baseModel, ok := r.parent.(*BaseModel); ok {
		return baseModel.GetAttribute(r.localKey)
	}

	// 通过反射获取值
	parentValue := reflect.ValueOf(r.parent)
	if parentValue.Kind() == reflect.Ptr {
		parentValue = parentValue.Elem()
	}

	// 尝试调用GetAttribute方法
	method := parentValue.MethodByName("GetAttribute")
	if method.IsValid() {
		results := method.Call([]reflect.Value{reflect.ValueOf(r.localKey)})
		if len(results) > 0 {
			return results[0].Interface()
		}
	}

	return nil
}

// getRelatedTableName 获取关联模型表名
func (r *BaseRelation) getRelatedTableName() string {
	if r.related == nil {
		return ""
	}

	// 直接根据类型名推断表名，避免反射调用方法
	relatedValue := reflect.ValueOf(r.related)
	if !relatedValue.IsValid() {
		return ""
	}

	var typeName string
	if relatedValue.Kind() == reflect.Ptr {
		if relatedValue.IsNil() {
			typeName = relatedValue.Type().Elem().Name()
		} else {
			typeName = relatedValue.Elem().Type().Name()
		}
	} else {
		typeName = relatedValue.Type().Name()
	}

	// 根据类型名推断表名
	switch typeName {
	case "User":
		return "users"
	case "Profile":
		return "profiles"
	case "Post":
		return "posts"
	case "Tag":
		return "tags"
	default:
		// 默认使用类型名的复数形式
		return strings.ToLower(typeName) + "s"
	}
}

// buildQuery 构建基础查询
func (r *BaseRelation) buildQuery() (db.QueryInterface, error) {
	tableName := r.getRelatedTableName()
	query, err := db.Table(tableName, r.connection)
	if err != nil {
		return nil, err
	}

	// 添加额外的WHERE条件
	for _, where := range r.wheres {
		query = query.Where(where.Field, where.Operator, where.Value)
	}

	return query, nil
}

// HasOne 一对一关联关系
type HasOne struct {
	*BaseRelation
}

// NewHasOne 创建一对一关联
func NewHasOne(parent interface{}, related interface{}, foreignKey, localKey string) *HasOne {
	if foreignKey == "" {
		foreignKey = "id" // 默认外键
	}
	if localKey == "" {
		localKey = "id" // 默认本地键
	}

	return &HasOne{
		BaseRelation: NewBaseRelation(parent, related, foreignKey, localKey, HasOneType),
	}
}

// Get 获取关联数据
func (h *HasOne) Get(ctx context.Context) (interface{}, error) {
	query, err := h.GetQuery()
	if err != nil {
		return nil, err
	}

	parentKey := h.getParentKey()
	if parentKey == nil {
		return nil, nil
	}

	data, err := query.Where(h.foreignKey, "=", parentKey).First(ctx)
	if err != nil {
		return nil, err
	}

	// 创建关联模型实例并填充数据
	return h.createRelatedInstance(data), nil
}

// GetQuery 获取查询构造器
func (h *HasOne) GetQuery() (db.QueryInterface, error) {
	return h.buildQuery()
}

// Where 添加查询条件
func (h *HasOne) Where(field string, operator string, value interface{}) RelationInterface {
	h.BaseRelation.Where(field, operator, value)
	return h
}

// Associate 关联模型
func (h *HasOne) Associate(ctx context.Context, model interface{}) error {
	parentKey := h.getParentKey()
	if parentKey == nil {
		return fmt.Errorf("parent key is nil")
	}

	// 设置外键值
	// 尝试通过反射设置属性
	modelValue := reflect.ValueOf(model)
	if modelValue.Kind() == reflect.Ptr {
		// 检查是否有SetAttribute方法
		setAttrMethod := modelValue.MethodByName("SetAttribute")
		saveMethod := modelValue.MethodByName("Save")

		if setAttrMethod.IsValid() && saveMethod.IsValid() {
			// 调用SetAttribute方法
			setAttrMethod.Call([]reflect.Value{
				reflect.ValueOf(h.foreignKey),
				reflect.ValueOf(parentKey),
			})

			// 调用Save方法
			results := saveMethod.Call([]reflect.Value{reflect.ValueOf(ctx)})
			if len(results) > 0 {
				if err, ok := results[0].Interface().(error); ok && err != nil {
					return err
				}
			}
			return nil
		}
	}

	return fmt.Errorf("model does not implement required interface")
}

// Dissociate 取消关联
func (h *HasOne) Dissociate(ctx context.Context, model interface{}) error {
	if baseModel, ok := model.(*BaseModel); ok {
		baseModel.SetAttribute(h.foreignKey, nil)
		return baseModel.Save(ctx)
	}

	return fmt.Errorf("model does not implement required interface")
}

// createRelatedInstance 创建关联模型实例
func (h *HasOne) createRelatedInstance(data map[string]interface{}) interface{} {
	// 这里需要根据具体的模型类型创建实例
	// 简化实现，返回map数据
	return data
}

// HasMany 一对多关联关系
type HasMany struct {
	*BaseRelation
}

// NewHasMany 创建一对多关联
func NewHasMany(parent interface{}, related interface{}, foreignKey, localKey string) *HasMany {
	if foreignKey == "" {
		foreignKey = "id"
	}
	if localKey == "" {
		localKey = "id"
	}

	return &HasMany{
		BaseRelation: NewBaseRelation(parent, related, foreignKey, localKey, HasManyType),
	}
}

// Get 获取关联数据
func (h *HasMany) Get(ctx context.Context) (interface{}, error) {
	query, err := h.GetQuery()
	if err != nil {
		return nil, err
	}

	parentKey := h.getParentKey()
	if parentKey == nil {
		return []interface{}{}, nil
	}

	data, err := query.Where(h.foreignKey, "=", parentKey).Get(ctx)
	if err != nil {
		return nil, err
	}

	// 转换为关联模型实例数组
	var results []interface{}
	for _, item := range data {
		results = append(results, h.createRelatedInstance(item))
	}

	return results, nil
}

// GetQuery 获取查询构造器
func (h *HasMany) GetQuery() (db.QueryInterface, error) {
	return h.buildQuery()
}

// Where 添加查询条件
func (h *HasMany) Where(field string, operator string, value interface{}) RelationInterface {
	h.BaseRelation.Where(field, operator, value)
	return h
}

// Associate 关联模型
func (h *HasMany) Associate(ctx context.Context, model interface{}) error {
	parentKey := h.getParentKey()
	if parentKey == nil {
		return fmt.Errorf("parent key is nil")
	}

	if baseModel, ok := model.(*BaseModel); ok {
		baseModel.SetAttribute(h.foreignKey, parentKey)
		return baseModel.Save(ctx)
	}

	return fmt.Errorf("model does not implement required interface")
}

// Dissociate 取消关联
func (h *HasMany) Dissociate(ctx context.Context, model interface{}) error {
	if baseModel, ok := model.(*BaseModel); ok {
		baseModel.SetAttribute(h.foreignKey, nil)
		return baseModel.Save(ctx)
	}

	return fmt.Errorf("model does not implement required interface")
}

// createRelatedInstance 创建关联模型实例
func (h *HasMany) createRelatedInstance(data map[string]interface{}) interface{} {
	return data
}

// BelongsTo 多对一关联关系
type BelongsTo struct {
	*BaseRelation
}

// NewBelongsTo 创建多对一关联
func NewBelongsTo(parent interface{}, related interface{}, foreignKey, ownerKey string) *BelongsTo {
	if foreignKey == "" {
		foreignKey = "parent_id" // 默认外键
	}
	if ownerKey == "" {
		ownerKey = "id" // 默认关联键
	}

	return &BelongsTo{
		BaseRelation: NewBaseRelation(parent, related, foreignKey, ownerKey, BelongsToType),
	}
}

// Get 获取关联数据
func (b *BelongsTo) Get(ctx context.Context) (interface{}, error) {
	query, err := b.GetQuery()
	if err != nil {
		return nil, err
	}

	// 对于BelongsTo，外键在父模型中
	foreignKeyValue := b.getForeignKeyValue()
	if foreignKeyValue == nil {
		return nil, nil
	}

	data, err := query.Where(b.localKey, "=", foreignKeyValue).First(ctx)
	if err != nil {
		return nil, err
	}

	return b.createRelatedInstance(data), nil
}

// GetQuery 获取查询构造器
func (b *BelongsTo) GetQuery() (db.QueryInterface, error) {
	return b.buildQuery()
}

// getForeignKeyValue 获取外键值
func (b *BelongsTo) getForeignKeyValue() interface{} {
	if baseModel, ok := b.parent.(*BaseModel); ok {
		return baseModel.GetAttribute(b.foreignKey)
	}
	return nil
}

// Where 添加查询条件
func (b *BelongsTo) Where(field string, operator string, value interface{}) RelationInterface {
	b.BaseRelation.Where(field, operator, value)
	return b
}

// Associate 关联模型
func (b *BelongsTo) Associate(ctx context.Context, model interface{}) error {
	// 获取关联模型的主键值
	var relatedKey interface{}
	if baseModel, ok := model.(*BaseModel); ok {
		relatedKey = baseModel.GetAttribute(b.localKey)
	}

	if relatedKey == nil {
		return fmt.Errorf("related model key is nil")
	}

	// 设置父模型的外键值
	if parentModel, ok := b.parent.(*BaseModel); ok {
		parentModel.SetAttribute(b.foreignKey, relatedKey)
		return parentModel.Save(ctx)
	}

	return fmt.Errorf("parent model does not implement required interface")
}

// Dissociate 取消关联
func (b *BelongsTo) Dissociate(ctx context.Context, model interface{}) error {
	if parentModel, ok := b.parent.(*BaseModel); ok {
		parentModel.SetAttribute(b.foreignKey, nil)
		return parentModel.Save(ctx)
	}

	return fmt.Errorf("parent model does not implement required interface")
}

// createRelatedInstance 创建关联模型实例
func (b *BelongsTo) createRelatedInstance(data map[string]interface{}) interface{} {
	return data
}

// ManyToMany 多对多关联关系
type ManyToMany struct {
	*BaseRelation
	// 中间表名
	pivotTable string
	// 中间表中的父模型外键
	pivotForeignKey string
	// 中间表中的关联模型外键
	pivotRelatedKey string
}

// NewManyToMany 创建多对多关联
func NewManyToMany(parent interface{}, related interface{}, pivotTable, pivotForeignKey, pivotRelatedKey, localKey, relatedKey string) *ManyToMany {
	if localKey == "" {
		localKey = "id"
	}
	if relatedKey == "" {
		relatedKey = "id"
	}

	return &ManyToMany{
		BaseRelation:    NewBaseRelation(parent, related, "", localKey, ManyToManyType),
		pivotTable:      pivotTable,
		pivotForeignKey: pivotForeignKey,
		pivotRelatedKey: pivotRelatedKey,
	}
}

// Get 获取关联数据
func (m *ManyToMany) Get(ctx context.Context) (interface{}, error) {
	parentKey := m.getParentKey()
	if parentKey == nil {
		return []interface{}{}, nil
	}

	relatedTableName := m.getRelatedTableName()

	// 构建多对多查询
	query, err := db.Table(relatedTableName, m.connection)
	if err != nil {
		return nil, err
	}

	// JOIN中间表
	query = query.InnerJoin(m.pivotTable, relatedTableName+"."+m.localKey, "=", m.pivotTable+"."+m.pivotRelatedKey).
		Where(m.pivotTable+"."+m.pivotForeignKey, "=", parentKey)

	// 添加额外的WHERE条件
	for _, where := range m.wheres {
		query = query.Where(where.Field, where.Operator, where.Value)
	}

	data, err := query.Get(ctx)
	if err != nil {
		return nil, err
	}

	var results []interface{}
	for _, item := range data {
		results = append(results, m.createRelatedInstance(item))
	}

	return results, nil
}

// GetQuery 获取查询构造器
func (m *ManyToMany) GetQuery() (db.QueryInterface, error) {
	return db.Table(m.getRelatedTableName(), m.connection)
}

// Where 添加查询条件
func (m *ManyToMany) Where(field string, operator string, value interface{}) RelationInterface {
	m.BaseRelation.Where(field, operator, value)
	return m
}

// Associate 关联模型
func (m *ManyToMany) Associate(ctx context.Context, model interface{}) error {
	parentKey := m.getParentKey()
	if parentKey == nil {
		return fmt.Errorf("parent key is nil")
	}

	var relatedKey interface{}
	if baseModel, ok := model.(*BaseModel); ok {
		relatedKey = baseModel.GetAttribute(m.localKey)
	}

	if relatedKey == nil {
		return fmt.Errorf("related model key is nil")
	}

	// 检查关联是否已存在
	query, err := db.Table(m.pivotTable, m.connection)
	if err != nil {
		return err
	}

	exists, err := query.
		Where(m.pivotForeignKey, "=", parentKey).
		Where(m.pivotRelatedKey, "=", relatedKey).
		Exists(ctx)

	if err != nil {
		return err
	}

	if exists {
		return nil // 关联已存在
	}

	// 插入中间表记录
	_, err = query.Insert(ctx, map[string]interface{}{
		m.pivotForeignKey: parentKey,
		m.pivotRelatedKey: relatedKey,
	})

	return err
}

// Dissociate 取消关联
func (m *ManyToMany) Dissociate(ctx context.Context, model interface{}) error {
	parentKey := m.getParentKey()
	if parentKey == nil {
		return fmt.Errorf("parent key is nil")
	}

	var relatedKey interface{}
	if baseModel, ok := model.(*BaseModel); ok {
		relatedKey = baseModel.GetAttribute(m.localKey)
	}

	if relatedKey == nil {
		return fmt.Errorf("related model key is nil")
	}

	// 删除中间表记录
	query, err := db.Table(m.pivotTable, m.connection)
	if err != nil {
		return err
	}

	_, err = query.
		Where(m.pivotForeignKey, "=", parentKey).
		Where(m.pivotRelatedKey, "=", relatedKey).
		Delete(ctx)

	return err
}

// createRelatedInstance 创建关联模型实例
func (m *ManyToMany) createRelatedInstance(data map[string]interface{}) interface{} {
	return data
}
