package model

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/zhoudm1743/torm/db"
)

// RelationType 关联类型
type RelationType int

const (
	RelationHasOne RelationType = iota
	RelationHasMany
	RelationBelongsTo
	RelationBelongsToMany
)

// RelationInterface 关联接口
type RelationInterface interface {
	// GetQuery 获取查询构建器
	GetQuery() *db.QueryBuilder
	// GetResults 获取关联结果
	GetResults() (interface{}, error)
	// First 获取第一个关联结果
	First() (map[string]interface{}, error)
	// Get 获取所有关联结果
	Get() ([]map[string]interface{}, error)
	// Where 添加查询条件
	Where(column, operator string, value interface{}) RelationInterface
	// GetType 获取关联类型
	GetType() RelationType
	// GetForeignKey 获取外键
	GetForeignKey() string
	// GetLocalKey 获取本地键
	GetLocalKey() string
	// GetRelatedTable 获取关联表名
	GetRelatedTable() string
	// EagerLoad 预加载关联数据
	EagerLoad(models []map[string]interface{}, relation string) ([]map[string]interface{}, error)
}

// RelationDefinition 关联定义 - 用于存储关联配置
type RelationDefinition struct {
	Name        string                                          // 关联名称
	Type        RelationType                                    // 关联类型
	Related     reflect.Type                                    // 关联模型类型
	ForeignKey  string                                          // 外键
	LocalKey    string                                          // 本地键
	PivotTable  string                                          // 中间表（多对多）
	Builder     func(parent *BaseModel) RelationInterface       // 关联构建器
	Constraints []func(query *db.QueryBuilder) *db.QueryBuilder // 约束条件
}

// RelationLoader 关联加载器 - 用于预加载
type RelationLoader struct {
	relations   map[string]*RelationDefinition
	constraints map[string][]func(query *db.QueryBuilder) *db.QueryBuilder
	mu          sync.RWMutex
}

// BaseRelation 基础关联结构 - 重构版本
type BaseRelation struct {
	// 父模型
	parent *BaseModel
	// 关联模型类型
	related reflect.Type
	// 查询构建器
	query *db.QueryBuilder
	// 外键
	foreignKey string
	// 本地键
	localKey string
	// 关联表名
	relatedTable string
	// 关联类型
	relationType RelationType
	// 是否已加载
	loaded bool
	// 缓存的结果
	cachedResults interface{}
}

// NewBaseRelation 创建基础关联
func NewBaseRelation(parent *BaseModel, related reflect.Type, foreignKey, localKey string) *BaseRelation {
	relatedTable := getTableNameFromType(related)
	return NewBaseRelationWithTable(parent, related, relatedTable, foreignKey, localKey)
}

// NewBaseRelationWithTable 创建带自定义表名的基础关联
func NewBaseRelationWithTable(parent *BaseModel, related reflect.Type, tableName, foreignKey, localKey string) *BaseRelation {
	query, err := db.NewQueryBuilder(parent.GetConnection())
	if err != nil {
		// 如果创建失败，使用默认连接
		query, _ = db.NewQueryBuilder("default")
	}

	if query != nil {
		query = query.From(tableName)
	}

	return &BaseRelation{
		parent:       parent,
		related:      related,
		query:        query,
		foreignKey:   foreignKey,
		localKey:     localKey,
		relatedTable: tableName,
	}
}

// GetQuery 获取查询构建器
func (r *BaseRelation) GetQuery() *db.QueryBuilder {
	return r.query
}

// GetType 获取关联类型
func (r *BaseRelation) GetType() RelationType {
	return r.relationType
}

// GetForeignKey 获取外键
func (r *BaseRelation) GetForeignKey() string {
	return r.foreignKey
}

// GetLocalKey 获取本地键
func (r *BaseRelation) GetLocalKey() string {
	return r.localKey
}

// GetRelatedTable 获取关联表名
func (r *BaseRelation) GetRelatedTable() string {
	return r.relatedTable
}

// IsLoaded 检查是否已加载
func (r *BaseRelation) IsLoaded() bool {
	return r.loaded
}

// SetLoaded 设置加载状态
func (r *BaseRelation) SetLoaded(loaded bool) {
	r.loaded = loaded
}

// GetCachedResults 获取缓存的结果
func (r *BaseRelation) GetCachedResults() interface{} {
	return r.cachedResults
}

// SetCachedResults 设置缓存的结果
func (r *BaseRelation) SetCachedResults(results interface{}) {
	r.cachedResults = results
	r.loaded = true
}

// ============================================================================
// HasOne 一对一关联
// ============================================================================

// HasOne 一对一关联
type HasOne struct {
	*BaseRelation
}

// NewHasOne 创建一对一关联
func NewHasOne(parent *BaseModel, related reflect.Type, foreignKey, localKey string) *HasOne {
	if foreignKey == "" {
		foreignKey = getDefaultForeignKey(parent.GetTableName())
	}
	if localKey == "" {
		localKey = parent.GetPrimaryKey()
	}

	baseRelation := NewBaseRelation(parent, related, foreignKey, localKey)
	baseRelation.relationType = RelationHasOne
	return &HasOne{BaseRelation: baseRelation}
}

// NewHasOneWithTable 创建带自定义表名的一对一关联
func NewHasOneWithTable(parent *BaseModel, related reflect.Type, tableName, foreignKey, localKey string) *HasOne {
	if foreignKey == "" {
		foreignKey = getDefaultForeignKey(parent.GetTableName())
	}
	if localKey == "" {
		localKey = parent.GetPrimaryKey()
	}

	baseRelation := NewBaseRelationWithTable(parent, related, tableName, foreignKey, localKey)
	baseRelation.relationType = RelationHasOne
	return &HasOne{BaseRelation: baseRelation}
}

// GetResults 获取关联结果
func (h *HasOne) GetResults() (interface{}, error) {
	return h.First()
}

// First 获取第一个关联结果
func (h *HasOne) First() (map[string]interface{}, error) {
	localValue := h.parent.GetAttribute(h.localKey)
	if localValue == nil {
		return nil, fmt.Errorf("本地键值为空")
	}

	return h.query.Where(h.foreignKey, "=", localValue).FirstRaw()
}

// Get 获取所有关联结果
func (h *HasOne) Get() ([]map[string]interface{}, error) {
	result, err := h.First()
	if err != nil {
		return nil, err
	}
	if result == nil {
		return []map[string]interface{}{}, nil
	}
	return []map[string]interface{}{result}, nil
}

// Where 添加查询约束
func (h *HasOne) Where(column, operator string, value interface{}) RelationInterface {
	if h.query != nil {
		h.query = h.query.Where(column, operator, value)
	}
	return h
}

// EagerLoad 预加载一对一关联数据
func (h *HasOne) EagerLoad(models []map[string]interface{}, relation string) ([]map[string]interface{}, error) {
	if len(models) == 0 {
		return models, nil
	}

	// 收集所有本地键值
	localKeys := make([]interface{}, 0, len(models))
	keyMap := make(map[interface{}]int) // 用于快速查找

	for i, model := range models {
		if localValue, ok := model[h.localKey]; ok && localValue != nil {
			localKeys = append(localKeys, localValue)
			keyMap[localValue] = i
		}
	}

	if len(localKeys) == 0 {
		return models, nil
	}

	// 查询关联数据
	query, err := db.NewQueryBuilder(h.parent.GetConnection())
	if err != nil {
		return nil, fmt.Errorf("创建查询构建器失败: %w", err)
	}

	relatedData, err := query.From(h.relatedTable).WhereIn(h.foreignKey, localKeys).GetRaw()
	if err != nil {
		return nil, fmt.Errorf("查询关联数据失败: %w", err)
	}

	// 将关联数据映射到父模型
	relatedMap := make(map[interface{}]map[string]interface{})
	for _, related := range relatedData {
		if fkValue, ok := related[h.foreignKey]; ok {
			relatedMap[fkValue] = related
		}
	}

	// 将关联数据附加到父模型
	for i, model := range models {
		if localValue, ok := model[h.localKey]; ok {
			if relatedItem, exists := relatedMap[localValue]; exists {
				models[i][relation] = relatedItem
			} else {
				models[i][relation] = nil
			}
		}
	}

	return models, nil
}

// ============================================================================
// HasMany 一对多关联
// ============================================================================

// HasMany 一对多关联
type HasMany struct {
	*BaseRelation
}

// NewHasMany 创建一对多关联
func NewHasMany(parent *BaseModel, related reflect.Type, foreignKey, localKey string) *HasMany {
	if foreignKey == "" {
		foreignKey = getDefaultForeignKey(parent.GetTableName())
	}
	if localKey == "" {
		localKey = parent.GetPrimaryKey()
	}

	baseRelation := NewBaseRelation(parent, related, foreignKey, localKey)
	baseRelation.relationType = RelationHasMany
	return &HasMany{BaseRelation: baseRelation}
}

// NewHasManyWithTable 创建带自定义表名的一对多关联
func NewHasManyWithTable(parent *BaseModel, related reflect.Type, tableName, foreignKey, localKey string) *HasMany {
	if foreignKey == "" {
		foreignKey = getDefaultForeignKey(parent.GetTableName())
	}
	if localKey == "" {
		localKey = parent.GetPrimaryKey()
	}

	baseRelation := NewBaseRelationWithTable(parent, related, tableName, foreignKey, localKey)
	baseRelation.relationType = RelationHasMany
	return &HasMany{BaseRelation: baseRelation}
}

// GetResults 获取关联结果
func (h *HasMany) GetResults() (interface{}, error) {
	return h.Get()
}

// First 获取第一个关联结果
func (h *HasMany) First() (map[string]interface{}, error) {
	localValue := h.parent.GetAttribute(h.localKey)
	if localValue == nil {
		return nil, fmt.Errorf("本地键值为空")
	}

	return h.query.Where(h.foreignKey, "=", localValue).FirstRaw()
}

// Get 获取所有关联结果
func (h *HasMany) Get() ([]map[string]interface{}, error) {
	localValue := h.parent.GetAttribute(h.localKey)
	if localValue == nil {
		return []map[string]interface{}{}, nil
	}

	return h.query.Where(h.foreignKey, "=", localValue).GetRaw()
}

// Where 添加查询约束
func (h *HasMany) Where(column, operator string, value interface{}) RelationInterface {
	if h.query != nil {
		h.query = h.query.Where(column, operator, value)
	}
	return h
}

// OrderBy 添加排序
func (h *HasMany) OrderBy(column, direction string) *HasMany {
	if h.query != nil {
		h.query = h.query.OrderBy(column, direction)
	}
	return h
}

// Limit 限制结果数量
func (h *HasMany) Limit(limit int) *HasMany {
	if h.query != nil {
		h.query = h.query.Limit(limit)
	}
	return h
}

// EagerLoad 预加载一对多关联数据
func (h *HasMany) EagerLoad(models []map[string]interface{}, relation string) ([]map[string]interface{}, error) {
	if len(models) == 0 {
		return models, nil
	}

	// 收集所有本地键值
	localKeys := make([]interface{}, 0, len(models))
	keyIndexMap := make(map[interface{}][]int) // 一个键可能对应多个模型

	for i, model := range models {
		if localValue, ok := model[h.localKey]; ok && localValue != nil {
			if _, exists := keyIndexMap[localValue]; !exists {
				localKeys = append(localKeys, localValue)
			}
			keyIndexMap[localValue] = append(keyIndexMap[localValue], i)
		}
	}

	if len(localKeys) == 0 {
		return models, nil
	}

	// 查询关联数据
	query, err := db.NewQueryBuilder(h.parent.GetConnection())
	if err != nil {
		return nil, fmt.Errorf("创建查询构建器失败: %w", err)
	}

	relatedData, err := query.From(h.relatedTable).WhereIn(h.foreignKey, localKeys).GetRaw()
	if err != nil {
		return nil, fmt.Errorf("查询关联数据失败: %w", err)
	}

	// 按外键分组关联数据
	relatedMap := make(map[interface{}][]map[string]interface{})
	for _, related := range relatedData {
		if fkValue, ok := related[h.foreignKey]; ok {
			relatedMap[fkValue] = append(relatedMap[fkValue], related)
		}
	}

	// 将关联数据附加到父模型
	for i, model := range models {
		if localValue, ok := model[h.localKey]; ok {
			if relatedItems, exists := relatedMap[localValue]; exists {
				models[i][relation] = relatedItems
			} else {
				models[i][relation] = []map[string]interface{}{}
			}
		}
	}

	return models, nil
}

// ============================================================================
// BelongsTo 反向关联
// ============================================================================

// BelongsTo 反向一对一/一对多关联
type BelongsTo struct {
	*BaseRelation
}

// NewBelongsTo 创建反向关联
func NewBelongsTo(parent *BaseModel, related reflect.Type, foreignKey, localKey string) *BelongsTo {
	if foreignKey == "" {
		foreignKey = getDefaultForeignKey(getTableNameFromType(related))
	}
	if localKey == "" {
		// BelongsTo 关联中，localKey 通常是关联模型的主键
		localKey = "id"
	}

	baseRelation := NewBaseRelation(parent, related, foreignKey, localKey)
	baseRelation.relationType = RelationBelongsTo
	return &BelongsTo{BaseRelation: baseRelation}
}

// NewBelongsToWithTable 创建带自定义表名的反向关联
func NewBelongsToWithTable(parent *BaseModel, related reflect.Type, tableName, foreignKey, localKey string) *BelongsTo {
	if foreignKey == "" {
		foreignKey = getDefaultForeignKey(tableName)
	}
	if localKey == "" {
		// BelongsTo 关联中，localKey 通常是关联模型的主键
		localKey = "id"
	}

	baseRelation := NewBaseRelationWithTable(parent, related, tableName, foreignKey, localKey)
	baseRelation.relationType = RelationBelongsTo
	return &BelongsTo{BaseRelation: baseRelation}
}

// GetResults 获取关联结果
func (b *BelongsTo) GetResults() (interface{}, error) {
	return b.First()
}

// First 获取第一个关联结果
func (b *BelongsTo) First() (map[string]interface{}, error) {
	foreignValue := b.parent.GetAttribute(b.foreignKey)
	if foreignValue == nil {
		return nil, fmt.Errorf("外键值为空")
	}

	return b.query.Where(b.localKey, "=", foreignValue).FirstRaw()
}

// Get 获取所有关联结果
func (b *BelongsTo) Get() ([]map[string]interface{}, error) {
	result, err := b.First()
	if err != nil {
		return nil, err
	}
	if result == nil {
		return []map[string]interface{}{}, nil
	}
	return []map[string]interface{}{result}, nil
}

// Where 添加查询约束
func (b *BelongsTo) Where(column, operator string, value interface{}) RelationInterface {
	if b.query != nil {
		b.query = b.query.Where(column, operator, value)
	}
	return b
}

// EagerLoad 预加载反向关联数据
func (b *BelongsTo) EagerLoad(models []map[string]interface{}, relation string) ([]map[string]interface{}, error) {
	if len(models) == 0 {
		return models, nil
	}

	// 收集所有外键值
	foreignKeys := make([]interface{}, 0, len(models))
	keySet := make(map[interface{}]bool)

	for _, model := range models {
		if fkValue, ok := model[b.foreignKey]; ok && fkValue != nil {
			if !keySet[fkValue] {
				foreignKeys = append(foreignKeys, fkValue)
				keySet[fkValue] = true
			}
		}
	}

	if len(foreignKeys) == 0 {
		return models, nil
	}

	// 查询关联数据
	query, err := db.NewQueryBuilder(b.parent.GetConnection())
	if err != nil {
		return nil, fmt.Errorf("创建查询构建器失败: %w", err)
	}

	relatedData, err := query.From(b.relatedTable).WhereIn(b.localKey, foreignKeys).GetRaw()
	if err != nil {
		return nil, fmt.Errorf("查询关联数据失败: %w", err)
	}

	// 将关联数据映射
	relatedMap := make(map[interface{}]map[string]interface{})
	for _, related := range relatedData {
		if pkValue, ok := related[b.localKey]; ok {
			relatedMap[pkValue] = related
		}
	}

	// 将关联数据附加到父模型
	for i, model := range models {
		if fkValue, ok := model[b.foreignKey]; ok {
			if relatedItem, exists := relatedMap[fkValue]; exists {
				models[i][relation] = relatedItem
			} else {
				models[i][relation] = nil
			}
		}
	}

	return models, nil
}

// ============================================================================
// BelongsToMany 多对多关联
// ============================================================================

// BelongsToMany 多对多关联
type BelongsToMany struct {
	*BaseRelation
	// 中间表
	pivotTable string
	// 中间表外键
	pivotForeignKey string
	// 中间表本地键
	pivotLocalKey string
}

// NewBelongsToMany 创建多对多关联
func NewBelongsToMany(parent *BaseModel, related reflect.Type, pivotTable, foreignKey, localKey string) *BelongsToMany {
	if foreignKey == "" {
		foreignKey = getDefaultForeignKey(getTableNameFromType(related))
	}
	if localKey == "" {
		localKey = getDefaultForeignKey(parent.GetTableName())
	}
	if pivotTable == "" {
		// 默认中间表名：两个表名按字母顺序组合
		table1 := parent.GetTableName()
		table2 := getTableNameFromType(related)
		if table1 > table2 {
			table1, table2 = table2, table1
		}
		pivotTable = table1 + "_" + table2
	}

	baseRelation := NewBaseRelation(parent, related, foreignKey, localKey)
	baseRelation.relationType = RelationBelongsToMany

	return &BelongsToMany{
		BaseRelation:    baseRelation,
		pivotTable:      pivotTable,
		pivotForeignKey: foreignKey,
		pivotLocalKey:   localKey,
	}
}

// NewBelongsToManyWithTable 创建带自定义表名的多对多关联
func NewBelongsToManyWithTable(parent *BaseModel, related reflect.Type, tableName, pivotTable, foreignKey, localKey string) *BelongsToMany {
	if foreignKey == "" {
		foreignKey = getDefaultForeignKey(tableName)
	}
	if localKey == "" {
		localKey = getDefaultForeignKey(parent.GetTableName())
	}
	if pivotTable == "" {
		// 默认中间表名：两个表名按字母顺序组合
		table1 := parent.GetTableName()
		table2 := tableName
		if table1 > table2 {
			table1, table2 = table2, table1
		}
		pivotTable = table1 + "_" + table2
	}

	baseRelation := NewBaseRelationWithTable(parent, related, tableName, foreignKey, localKey)
	baseRelation.relationType = RelationBelongsToMany

	return &BelongsToMany{
		BaseRelation:    baseRelation,
		pivotTable:      pivotTable,
		pivotForeignKey: foreignKey,
		pivotLocalKey:   localKey,
	}
}

// GetResults 获取关联结果
func (b *BelongsToMany) GetResults() (interface{}, error) {
	return b.Get()
}

// First 获取第一个关联结果
func (b *BelongsToMany) First() (map[string]interface{}, error) {
	results, err := b.Get()
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

// Get 获取所有关联结果
func (b *BelongsToMany) Get() ([]map[string]interface{}, error) {
	localValue := b.parent.GetAttribute(b.parent.GetPrimaryKey())
	if localValue == nil {
		return []map[string]interface{}{}, nil
	}

	// 构建多对多查询
	return b.query.
		Join(b.pivotTable, fmt.Sprintf("%s.%s", b.relatedTable, "id"), "=", fmt.Sprintf("%s.%s", b.pivotTable, b.pivotForeignKey)).
		Where(fmt.Sprintf("%s.%s", b.pivotTable, b.pivotLocalKey), "=", localValue).
		Select(fmt.Sprintf("%s.*", b.relatedTable)).
		GetRaw()
}

// Where 添加查询约束
func (b *BelongsToMany) Where(column, operator string, value interface{}) RelationInterface {
	if b.query != nil {
		b.query = b.query.Where(column, operator, value)
	}
	return b
}

// OrderBy 添加排序
func (b *BelongsToMany) OrderBy(column, direction string) *BelongsToMany {
	if b.query != nil {
		b.query = b.query.OrderBy(column, direction)
	}
	return b
}

// Limit 限制结果数量
func (b *BelongsToMany) Limit(limit int) *BelongsToMany {
	if b.query != nil {
		b.query = b.query.Limit(limit)
	}
	return b
}

// Attach 添加关联关系
func (b *BelongsToMany) Attach(relatedID interface{}) error {
	localValue := b.parent.GetAttribute(b.parent.GetPrimaryKey())
	if localValue == nil {
		return fmt.Errorf("主键值为空")
	}

	query, err := db.NewQueryBuilder(b.parent.GetConnection())
	if err != nil {
		return fmt.Errorf("创建查询构建器失败: %w", err)
	}

	_, err = query.From(b.pivotTable).Insert(map[string]interface{}{
		b.pivotLocalKey:   localValue,
		b.pivotForeignKey: relatedID,
	})

	return err
}

// Detach 移除关联关系
func (b *BelongsToMany) Detach(relatedID interface{}) error {
	localValue := b.parent.GetAttribute(b.parent.GetPrimaryKey())
	if localValue == nil {
		return fmt.Errorf("主键值为空")
	}

	query, err := db.NewQueryBuilder(b.parent.GetConnection())
	if err != nil {
		return fmt.Errorf("创建查询构建器失败: %w", err)
	}

	_, err = query.From(b.pivotTable).
		Where(b.pivotLocalKey, "=", localValue).
		Where(b.pivotForeignKey, "=", relatedID).
		Delete()

	return err
}

// Sync 同步关联关系（替换所有关联）
func (b *BelongsToMany) Sync(relatedIDs []interface{}) error {
	localValue := b.parent.GetAttribute(b.parent.GetPrimaryKey())
	if localValue == nil {
		return fmt.Errorf("主键值为空")
	}

	query, err := db.NewQueryBuilder(b.parent.GetConnection())
	if err != nil {
		return fmt.Errorf("创建查询构建器失败: %w", err)
	}

	// 删除现有关联
	_, err = query.From(b.pivotTable).
		Where(b.pivotLocalKey, "=", localValue).
		Delete()
	if err != nil {
		return fmt.Errorf("删除现有关联失败: %w", err)
	}

	// 添加新关联
	for _, relatedID := range relatedIDs {
		newQuery, _ := db.NewQueryBuilder(b.parent.GetConnection())
		_, err = newQuery.From(b.pivotTable).Insert(map[string]interface{}{
			b.pivotLocalKey:   localValue,
			b.pivotForeignKey: relatedID,
		})
		if err != nil {
			return fmt.Errorf("添加关联失败: %w", err)
		}
	}

	return nil
}

// EagerLoad 预加载多对多关联数据
func (b *BelongsToMany) EagerLoad(models []map[string]interface{}, relation string) ([]map[string]interface{}, error) {
	if len(models) == 0 {
		return models, nil
	}

	// 收集所有本地键值（父模型的主键）
	localKeys := make([]interface{}, 0, len(models))
	keySet := make(map[interface{}]bool)

	parentPK := b.parent.GetPrimaryKey()
	for _, model := range models {
		if localValue, ok := model[parentPK]; ok && localValue != nil {
			if !keySet[localValue] {
				localKeys = append(localKeys, localValue)
				keySet[localValue] = true
			}
		}
	}

	if len(localKeys) == 0 {
		return models, nil
	}

	// 查询中间表数据
	pivotQuery, err := db.NewQueryBuilder(b.parent.GetConnection())
	if err != nil {
		return nil, fmt.Errorf("创建查询构建器失败: %w", err)
	}

	pivotData, err := pivotQuery.From(b.pivotTable).WhereIn(b.pivotLocalKey, localKeys).GetRaw()
	if err != nil {
		return nil, fmt.Errorf("查询中间表数据失败: %w", err)
	}

	// 收集所有关联模型的ID
	relatedIDs := make([]interface{}, 0)
	relatedIDSet := make(map[interface{}]bool)
	pivotMap := make(map[interface{}][]interface{}) // localKey -> []foreignKey

	for _, pivot := range pivotData {
		localVal := pivot[b.pivotLocalKey]
		foreignVal := pivot[b.pivotForeignKey]
		if localVal != nil && foreignVal != nil {
			pivotMap[localVal] = append(pivotMap[localVal], foreignVal)
			if !relatedIDSet[foreignVal] {
				relatedIDs = append(relatedIDs, foreignVal)
				relatedIDSet[foreignVal] = true
			}
		}
	}

	if len(relatedIDs) == 0 {
		// 没有关联数据，设置空数组
		for i := range models {
			models[i][relation] = []map[string]interface{}{}
		}
		return models, nil
	}

	// 查询关联模型数据
	relatedQuery, err := db.NewQueryBuilder(b.parent.GetConnection())
	if err != nil {
		return nil, fmt.Errorf("创建查询构建器失败: %w", err)
	}

	relatedData, err := relatedQuery.From(b.relatedTable).WhereIn("id", relatedIDs).GetRaw()
	if err != nil {
		return nil, fmt.Errorf("查询关联数据失败: %w", err)
	}

	// 将关联数据映射
	relatedMap := make(map[interface{}]map[string]interface{})
	for _, related := range relatedData {
		if id, ok := related["id"]; ok {
			relatedMap[id] = related
		}
	}

	// 将关联数据附加到父模型
	for i, model := range models {
		localValue := model[parentPK]
		if localValue == nil {
			models[i][relation] = []map[string]interface{}{}
			continue
		}

		foreignKeys := pivotMap[localValue]
		relatedItems := make([]map[string]interface{}, 0, len(foreignKeys))
		for _, fk := range foreignKeys {
			if relatedItem, exists := relatedMap[fk]; exists {
				relatedItems = append(relatedItems, relatedItem)
			}
		}
		models[i][relation] = relatedItems
	}

	return models, nil
}

// GetPivotTable 获取中间表名
func (b *BelongsToMany) GetPivotTable() string {
	return b.pivotTable
}

// ============================================================================
// 辅助函数
// ============================================================================

// getTableNameFromType 从类型获取表名 - 改进版本
func getTableNameFromType(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// 尝试创建实例并获取表名
	if t.Kind() == reflect.Struct {
		// 创建一个临时实例来获取正确的表名
		instance := reflect.New(t).Interface()
		if baseModeler, ok := instance.(interface{ GetTableName() string }); ok {
			tableName := baseModeler.GetTableName()
			if tableName != "" {
				return tableName
			}
		}
	}

	// 回退到类型名推断
	return toSnakeCase(t.Name())
}

// getDefaultForeignKey 获取默认外键名 - 改进版本
func getDefaultForeignKey(tableName string) string {
	if tableName == "" {
		return "id"
	}

	// 去掉表名的复数形式并添加 _id
	singular := strings.TrimSuffix(tableName, "s")
	return singular + "_id"
}
