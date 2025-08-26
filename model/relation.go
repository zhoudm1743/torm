package model

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/zhoudm1743/torm/db"
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
}

// BaseRelation 基础关联结构
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
}

// NewBaseRelation 创建基础关联
func NewBaseRelation(parent *BaseModel, related reflect.Type, foreignKey, localKey string) *BaseRelation {
	relatedTable := getTableNameFromType(related)

	query, err := db.NewQueryBuilder(parent.GetConnection())
	if err != nil {
		// 如果创建失败，使用默认连接
		query, _ = db.NewQueryBuilder("default")
	}

	if query != nil {
		query = query.From(relatedTable)
	}

	return &BaseRelation{
		parent:       parent,
		related:      related,
		query:        query,
		foreignKey:   foreignKey,
		localKey:     localKey,
		relatedTable: relatedTable,
	}
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

// ============================================================================
// 辅助函数
// ============================================================================

// getTableNameFromType 从类型获取表名
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

// getDefaultForeignKey 获取默认外键名
func getDefaultForeignKey(tableName string) string {
	// 去掉表名的复数形式并添加 _id
	singular := strings.TrimSuffix(tableName, "s")
	return singular + "_id"
}

// WithConstraints 添加查询约束（链式调用）
func (h *HasOne) Where(column string, operator string, value interface{}) *HasOne {
	if h.query != nil {
		h.query = h.query.Where(column, operator, value)
	}
	return h
}

func (h *HasMany) Where(column string, operator string, value interface{}) *HasMany {
	if h.query != nil {
		h.query = h.query.Where(column, operator, value)
	}
	return h
}

func (b *BelongsTo) Where(column string, operator string, value interface{}) *BelongsTo {
	if b.query != nil {
		b.query = b.query.Where(column, operator, value)
	}
	return b
}

func (b *BelongsToMany) Where(column string, operator string, value interface{}) *BelongsToMany {
	if b.query != nil {
		b.query = b.query.Where(column, operator, value)
	}
	return b
}

// OrderBy 添加排序（链式调用）
func (h *HasMany) OrderBy(column, direction string) *HasMany {
	if h.query != nil {
		h.query = h.query.OrderBy(column, direction)
	}
	return h
}

func (b *BelongsToMany) OrderBy(column, direction string) *BelongsToMany {
	if b.query != nil {
		b.query = b.query.OrderBy(column, direction)
	}
	return b
}

// Limit 限制结果数量
func (h *HasMany) Limit(limit int) *HasMany {
	if h.query != nil {
		h.query = h.query.Limit(limit)
	}
	return h
}

func (b *BelongsToMany) Limit(limit int) *BelongsToMany {
	if b.query != nil {
		b.query = b.query.Limit(limit)
	}
	return b
}
