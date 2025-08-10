package model

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"torm/pkg/db"
)

// WithConfig 预载入配置
type WithConfig struct {
	Relations []string
	Closure   func(db.QueryInterface) db.QueryInterface
	Field     []string
	Cache     *CacheConfig
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Enable bool
	Key    string
	TTL    int // 秒
}

// EagerLoadManager 预加载管理器
type EagerLoadManager struct {
	model      *BaseModel
	withConfig map[string]*WithConfig
}

// NewEagerLoadManager 创建预加载管理器
func NewEagerLoadManager(model *BaseModel) *EagerLoadManager {
	return &EagerLoadManager{
		model:      model,
		withConfig: make(map[string]*WithConfig),
	}
}

// With 设置预载入关联
func (elm *EagerLoadManager) With(relations ...string) *EagerLoadManager {
	for _, relation := range relations {
		elm.withConfig[relation] = &WithConfig{
			Relations: []string{relation},
		}
	}
	return elm
}

// WithClosure 设置预载入关联和查询条件
func (elm *EagerLoadManager) WithClosure(relation string, closure func(db.QueryInterface) db.QueryInterface) *EagerLoadManager {
	elm.withConfig[relation] = &WithConfig{
		Relations: []string{relation},
		Closure:   closure,
	}
	return elm
}

// WithField 设置预载入关联字段
func (elm *EagerLoadManager) WithField(relation string, fields ...string) *EagerLoadManager {
	if config, exists := elm.withConfig[relation]; exists {
		config.Field = fields
	} else {
		elm.withConfig[relation] = &WithConfig{
			Relations: []string{relation},
			Field:     fields,
		}
	}
	return elm
}

// LoadRelations 批量加载关联数据
func (elm *EagerLoadManager) LoadRelations(ctx context.Context, models []interface{}) error {
	if len(models) == 0 {
		return nil
	}

	for relationName, config := range elm.withConfig {
		if err := elm.loadRelation(ctx, models, relationName, config); err != nil {
			return err
		}
	}

	return nil
}

// loadRelation 加载单个关联
func (elm *EagerLoadManager) loadRelation(ctx context.Context, models []interface{}, relationName string, config *WithConfig) error {
	// 获取关联定义方法
	relationMethod := elm.getRelationMethod(relationName)
	if relationMethod == nil {
		return fmt.Errorf("relation method %s not found", relationName)
	}

	// 根据关联类型执行不同的预载入策略
	switch relation := relationMethod.(type) {
	case *HasMany:
		return elm.eagerlyLoadHasMany(ctx, models, relationName, relation, config)
	case *HasOne:
		return elm.eagerlyLoadHasOne(ctx, models, relationName, relation, config)
	case *BelongsTo:
		return elm.eagerlyLoadBelongsTo(ctx, models, relationName, relation, config)
	case *ManyToMany:
		return elm.eagerlyLoadManyToMany(ctx, models, relationName, relation, config)
	default:
		return fmt.Errorf("unsupported relation type for %s", relationName)
	}
}

// getRelationMethod 获取关联方法
func (elm *EagerLoadManager) getRelationMethod(relationName string) interface{} {
	// 通过反射调用模型的关联方法
	modelValue := reflect.ValueOf(elm.model)
	methodName := strings.Title(relationName)

	method := modelValue.MethodByName(methodName)
	if !method.IsValid() {
		return nil
	}

	// 调用关联方法获取关联对象
	results := method.Call([]reflect.Value{})
	if len(results) > 0 {
		return results[0].Interface()
	}

	return nil
}

// eagerlyLoadHasMany 预载入一对多关联
func (elm *EagerLoadManager) eagerlyLoadHasMany(ctx context.Context, models []interface{}, relationName string, relation *HasMany, config *WithConfig) error {
	// 收集所有父模型的主键值
	localKeys := make([]interface{}, 0, len(models))
	modelMap := make(map[interface{}][]interface{})

	for _, model := range models {
		if baseModel, ok := model.(*BaseModel); ok {
			localKey := baseModel.GetAttribute(relation.localKey)
			if localKey != nil {
				localKeys = append(localKeys, localKey)
				if _, exists := modelMap[localKey]; !exists {
					modelMap[localKey] = make([]interface{}, 0)
				}
			}
		}
	}

	if len(localKeys) == 0 {
		return nil
	}

	// 构建关联查询
	query, err := relation.GetQuery()
	if err != nil {
		return err
	}

	// 添加外键条件
	query = query.WhereIn(relation.foreignKey, localKeys)

	// 应用额外的查询条件
	if config.Closure != nil {
		query = config.Closure(query)
	}

	// 字段限制
	if len(config.Field) > 0 {
		query = query.Select(config.Field...)
	}

	// 执行查询
	relationData, err := query.Get()
	if err != nil {
		return err
	}

	// 组织关联数据
	relationMap := make(map[interface{}][]map[string]interface{})
	for _, data := range relationData {
		foreignKeyValue := data[relation.foreignKey]
		if _, exists := relationMap[foreignKeyValue]; !exists {
			relationMap[foreignKeyValue] = make([]map[string]interface{}, 0)
		}
		relationMap[foreignKeyValue] = append(relationMap[foreignKeyValue], data)
	}

	// 设置关联数据到模型
	for _, model := range models {
		if baseModel, ok := model.(*BaseModel); ok {
			localKey := baseModel.GetAttribute(relation.localKey)
			if relationData, exists := relationMap[localKey]; exists {
				baseModel.SetRelation(relationName, relationData)
			} else {
				baseModel.SetRelation(relationName, []map[string]interface{}{})
			}
		}
	}

	return nil
}

// eagerlyLoadHasOne 预载入一对一关联
func (elm *EagerLoadManager) eagerlyLoadHasOne(ctx context.Context, models []interface{}, relationName string, relation *HasOne, config *WithConfig) error {
	// 收集所有父模型的主键值
	localKeys := make([]interface{}, 0, len(models))

	for _, model := range models {
		if baseModel, ok := model.(*BaseModel); ok {
			localKey := baseModel.GetAttribute(relation.localKey)
			if localKey != nil {
				localKeys = append(localKeys, localKey)
			}
		}
	}

	if len(localKeys) == 0 {
		return nil
	}

	// 构建关联查询
	query, err := relation.GetQuery()
	if err != nil {
		return err
	}

	// 添加外键条件
	query = query.WhereIn(relation.foreignKey, localKeys)

	// 应用额外的查询条件
	if config.Closure != nil {
		query = config.Closure(query)
	}

	// 字段限制
	if len(config.Field) > 0 {
		query = query.Select(config.Field...)
	}

	// 执行查询
	relationData, err := query.Get()
	if err != nil {
		return err
	}

	// 组织关联数据
	relationMap := make(map[interface{}]map[string]interface{})
	for _, data := range relationData {
		foreignKeyValue := data[relation.foreignKey]
		relationMap[foreignKeyValue] = data
	}

	// 设置关联数据到模型
	for _, model := range models {
		if baseModel, ok := model.(*BaseModel); ok {
			localKey := baseModel.GetAttribute(relation.localKey)
			if relationData, exists := relationMap[localKey]; exists {
				baseModel.SetRelation(relationName, relationData)
			} else {
				baseModel.SetRelation(relationName, nil)
			}
		}
	}

	return nil
}

// eagerlyLoadBelongsTo 预载入多对一关联
func (elm *EagerLoadManager) eagerlyLoadBelongsTo(ctx context.Context, models []interface{}, relationName string, relation *BelongsTo, config *WithConfig) error {
	// 收集所有外键值
	foreignKeys := make([]interface{}, 0, len(models))

	for _, model := range models {
		if baseModel, ok := model.(*BaseModel); ok {
			foreignKey := baseModel.GetAttribute(relation.foreignKey)
			if foreignKey != nil {
				foreignKeys = append(foreignKeys, foreignKey)
			}
		}
	}

	if len(foreignKeys) == 0 {
		return nil
	}

	// 构建关联查询
	query, err := relation.GetQuery()
	if err != nil {
		return err
	}

	// 添加主键条件
	query = query.WhereIn(relation.localKey, foreignKeys)

	// 应用额外的查询条件
	if config.Closure != nil {
		query = config.Closure(query)
	}

	// 字段限制
	if len(config.Field) > 0 {
		query = query.Select(config.Field...)
	}

	// 执行查询
	relationData, err := query.Get()
	if err != nil {
		return err
	}

	// 组织关联数据
	relationMap := make(map[interface{}]map[string]interface{})
	for _, data := range relationData {
		localKeyValue := data[relation.localKey]
		relationMap[localKeyValue] = data
	}

	// 设置关联数据到模型
	for _, model := range models {
		if baseModel, ok := model.(*BaseModel); ok {
			foreignKey := baseModel.GetAttribute(relation.foreignKey)
			if relationData, exists := relationMap[foreignKey]; exists {
				baseModel.SetRelation(relationName, relationData)
			} else {
				baseModel.SetRelation(relationName, nil)
			}
		}
	}

	return nil
}

// eagerlyLoadManyToMany 预载入多对多关联
func (elm *EagerLoadManager) eagerlyLoadManyToMany(ctx context.Context, models []interface{}, relationName string, relation *ManyToMany, config *WithConfig) error {
	// 收集所有父模型的主键值
	localKeys := make([]interface{}, 0, len(models))

	for _, model := range models {
		if baseModel, ok := model.(*BaseModel); ok {
			localKey := baseModel.GetAttribute(relation.localKey)
			if localKey != nil {
				localKeys = append(localKeys, localKey)
			}
		}
	}

	if len(localKeys) == 0 {
		return nil
	}

	// 获取关联表名
	relatedTableName := relation.getRelatedTableName()

	// 构建多对多查询 - 需要联接中间表
	query, err := db.Table(relatedTableName, relation.connection)
	if err != nil {
		return err
	}

	// JOIN中间表
	query = query.InnerJoin(
		relation.pivotTable,
		relatedTableName+"."+relation.localKey,
		"=",
		relation.pivotTable+"."+relation.pivotRelatedKey,
	).WhereIn(relation.pivotTable+"."+relation.pivotForeignKey, localKeys)

	// 应用额外的查询条件
	if config.Closure != nil {
		query = config.Closure(query)
	}

	// 字段限制 - 需要包含关联字段
	selectFields := []string{relatedTableName + ".*", relation.pivotTable + "." + relation.pivotForeignKey + " as __pivot_foreign_key"}
	if len(config.Field) > 0 {
		selectFields = make([]string, len(config.Field))
		for i, field := range config.Field {
			selectFields[i] = relatedTableName + "." + field
		}
		selectFields = append(selectFields, relation.pivotTable+"."+relation.pivotForeignKey+" as __pivot_foreign_key")
	}
	query = query.Select(selectFields...)

	// 执行查询
	relationData, err := query.Get()
	if err != nil {
		return err
	}

	// 组织关联数据
	relationMap := make(map[interface{}][]map[string]interface{})
	for _, data := range relationData {
		pivotForeignKey := data["__pivot_foreign_key"]
		if _, exists := relationMap[pivotForeignKey]; !exists {
			relationMap[pivotForeignKey] = make([]map[string]interface{}, 0)
		}
		// 移除pivot字段
		delete(data, "__pivot_foreign_key")
		relationMap[pivotForeignKey] = append(relationMap[pivotForeignKey], data)
	}

	// 设置关联数据到模型
	for _, model := range models {
		if baseModel, ok := model.(*BaseModel); ok {
			localKey := baseModel.GetAttribute(relation.localKey)
			if relationData, exists := relationMap[localKey]; exists {
				baseModel.SetRelation(relationName, relationData)
			} else {
				baseModel.SetRelation(relationName, []map[string]interface{}{})
			}
		}
	}

	return nil
}

// ModelCollection 模型集合 - 支持预载入
type ModelCollection struct {
	models []interface{}
	elm    *EagerLoadManager
}

// NewModelCollection 创建模型集合
func NewModelCollection(models []interface{}) *ModelCollection {
	var elm *EagerLoadManager
	if len(models) > 0 {
		if baseModel, ok := models[0].(*BaseModel); ok {
			elm = NewEagerLoadManager(baseModel)
		}
	}

	return &ModelCollection{
		models: models,
		elm:    elm,
	}
}

// With 设置预载入关联
func (mc *ModelCollection) With(relations ...string) *ModelCollection {
	if mc.elm != nil {
		mc.elm.With(relations...)
	}
	return mc
}

// WithClosure 设置预载入关联和查询条件
func (mc *ModelCollection) WithClosure(relation string, closure func(db.QueryInterface) db.QueryInterface) *ModelCollection {
	if mc.elm != nil {
		mc.elm.WithClosure(relation, closure)
	}
	return mc
}

// Load 执行预载入
func (mc *ModelCollection) Load(ctx context.Context) error {
	if mc.elm != nil {
		return mc.elm.LoadRelations(ctx, mc.models)
	}
	return nil
}

// Models 获取模型数组
func (mc *ModelCollection) Models() []interface{} {
	return mc.models
}

// Count 获取模型数量
func (mc *ModelCollection) Count() int {
	return len(mc.models)
}

// ToSlice 转换为切片
func (mc *ModelCollection) ToSlice() []map[string]interface{} {
	result := make([]map[string]interface{}, len(mc.models))
	for i, model := range mc.models {
		if baseModel, ok := model.(*BaseModel); ok {
			result[i] = baseModel.ToMap()
		}
	}
	return result
}
