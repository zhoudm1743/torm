package model

import (
	"fmt"

	"github.com/zhoudm1743/torm/db"
)

// ModelQueryBuilder 模型查询构建器 - 支持预加载
type ModelQueryBuilder struct {
	model       *BaseModel
	query       *db.QueryBuilder
	eagerLoader *EagerLoader
}

// NewModelQueryBuilder 创建模型查询构建器
func NewModelQueryBuilder(model *BaseModel) *ModelQueryBuilder {
	query, _ := model.Query()
	return &ModelQueryBuilder{
		model:       model,
		query:       query,
		eagerLoader: NewEagerLoader(),
	}
}

// With 添加要预加载的关联
func (mqb *ModelQueryBuilder) With(relations ...string) *ModelQueryBuilder {
	mqb.eagerLoader.With(relations...)
	return mqb
}

// WithConstraint 添加带约束的预加载
func (mqb *ModelQueryBuilder) WithConstraint(relation string, constraint func(*db.QueryBuilder) *db.QueryBuilder) *ModelQueryBuilder {
	mqb.eagerLoader.WithConstraint(relation, constraint)
	return mqb
}

// Where 添加WHERE条件
func (mqb *ModelQueryBuilder) Where(args ...interface{}) *ModelQueryBuilder {
	if mqb.query != nil {
		mqb.query = mqb.query.Where(args...)
	}
	return mqb
}

// OrWhere 添加OR WHERE条件
func (mqb *ModelQueryBuilder) OrWhere(args ...interface{}) *ModelQueryBuilder {
	if mqb.query != nil {
		mqb.query = mqb.query.OrWhere(args...)
	}
	return mqb
}

// WhereIn WHERE IN条件
func (mqb *ModelQueryBuilder) WhereIn(field string, values []interface{}) *ModelQueryBuilder {
	if mqb.query != nil {
		mqb.query = mqb.query.WhereIn(field, values)
	}
	return mqb
}

// WhereNotIn WHERE NOT IN条件
func (mqb *ModelQueryBuilder) WhereNotIn(field string, values []interface{}) *ModelQueryBuilder {
	if mqb.query != nil {
		mqb.query = mqb.query.WhereNotIn(field, values)
	}
	return mqb
}

// WhereBetween WHERE BETWEEN条件
func (mqb *ModelQueryBuilder) WhereBetween(field string, values []interface{}) *ModelQueryBuilder {
	if mqb.query != nil {
		mqb.query = mqb.query.WhereBetween(field, values)
	}
	return mqb
}

// WhereNull WHERE IS NULL条件
func (mqb *ModelQueryBuilder) WhereNull(field string) *ModelQueryBuilder {
	if mqb.query != nil {
		mqb.query = mqb.query.WhereNull(field)
	}
	return mqb
}

// WhereNotNull WHERE IS NOT NULL条件
func (mqb *ModelQueryBuilder) WhereNotNull(field string) *ModelQueryBuilder {
	if mqb.query != nil {
		mqb.query = mqb.query.WhereNotNull(field)
	}
	return mqb
}

// OrderBy 排序
func (mqb *ModelQueryBuilder) OrderBy(column, direction string) *ModelQueryBuilder {
	if mqb.query != nil {
		mqb.query = mqb.query.OrderBy(column, direction)
	}
	return mqb
}

// GroupBy 分组
func (mqb *ModelQueryBuilder) GroupBy(columns ...string) *ModelQueryBuilder {
	if mqb.query != nil {
		mqb.query = mqb.query.GroupBy(columns...)
	}
	return mqb
}

// Having HAVING条件
func (mqb *ModelQueryBuilder) Having(args ...interface{}) *ModelQueryBuilder {
	if mqb.query != nil {
		mqb.query = mqb.query.Having(args...)
	}
	return mqb
}

// Limit 限制记录数
func (mqb *ModelQueryBuilder) Limit(limit int) *ModelQueryBuilder {
	if mqb.query != nil {
		mqb.query = mqb.query.Limit(limit)
	}
	return mqb
}

// Offset 偏移量
func (mqb *ModelQueryBuilder) Offset(offset int) *ModelQueryBuilder {
	if mqb.query != nil {
		mqb.query = mqb.query.Offset(offset)
	}
	return mqb
}

// Page 分页
func (mqb *ModelQueryBuilder) Page(page, size int) *ModelQueryBuilder {
	if mqb.query != nil {
		mqb.query = mqb.query.Page(page, size)
	}
	return mqb
}

// Select 选择字段
func (mqb *ModelQueryBuilder) Select(fields ...interface{}) *ModelQueryBuilder {
	if mqb.query != nil {
		mqb.query = mqb.query.Select(fields...)
	}
	return mqb
}

// Join 内连接
func (mqb *ModelQueryBuilder) Join(table, first, operator, second string) *ModelQueryBuilder {
	if mqb.query != nil {
		mqb.query = mqb.query.Join(table, first, operator, second)
	}
	return mqb
}

// LeftJoin 左连接
func (mqb *ModelQueryBuilder) LeftJoin(table, first, operator, second string) *ModelQueryBuilder {
	if mqb.query != nil {
		mqb.query = mqb.query.LeftJoin(table, first, operator, second)
	}
	return mqb
}

// RightJoin 右连接
func (mqb *ModelQueryBuilder) RightJoin(table, first, operator, second string) *ModelQueryBuilder {
	if mqb.query != nil {
		mqb.query = mqb.query.RightJoin(table, first, operator, second)
	}
	return mqb
}

// Get 执行查询并返回数据（带预加载）
func (mqb *ModelQueryBuilder) Get() ([]map[string]interface{}, error) {
	if mqb.query == nil {
		return nil, fmt.Errorf("查询构建器未初始化")
	}

	// 执行主查询
	results, err := mqb.query.GetRaw()
	if err != nil {
		return nil, err
	}

	// 如果有预加载关联，执行预加载
	if mqb.eagerLoader.HasRelations() {
		results, err = LoadRelations(mqb.model, results, mqb.eagerLoader.GetRelations())
		if err != nil {
			return nil, fmt.Errorf("预加载关联失败: %w", err)
		}
	}

	return results, nil
}

// First 获取第一条记录（带预加载）
func (mqb *ModelQueryBuilder) First() (map[string]interface{}, error) {
	mqb.Limit(1)
	results, err := mqb.Get()
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, db.ErrRecordNotFound.WithContext("table", mqb.model.GetTableName())
	}

	return results[0], nil
}

// Find 按主键查找（带预加载）
func (mqb *ModelQueryBuilder) Find(id interface{}) (map[string]interface{}, error) {
	mqb.Where(mqb.model.GetPrimaryKey(), "=", id)
	return mqb.First()
}

// Count 计算记录数量
func (mqb *ModelQueryBuilder) Count() (int64, error) {
	if mqb.query == nil {
		return 0, fmt.Errorf("查询构建器未初始化")
	}
	return mqb.query.Count()
}

// Exists 检查记录是否存在
func (mqb *ModelQueryBuilder) Exists() (bool, error) {
	count, err := mqb.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Paginate 分页查询（带预加载）
func (mqb *ModelQueryBuilder) Paginate(page, perPage int) (*db.Paginator, error) {
	if mqb.query == nil {
		return nil, fmt.Errorf("查询构建器未初始化")
	}

	// 获取总数
	total, err := mqb.query.Count()
	if err != nil {
		return nil, err
	}

	// 计算分页
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	lastPage := int((total + int64(perPage) - 1) / int64(perPage))
	if lastPage < 1 {
		lastPage = 1
	}

	// 执行分页查询
	mqb.Page(page, perPage)
	results, err := mqb.Get()
	if err != nil {
		return nil, err
	}

	// 计算from和to
	from := (page-1)*perPage + 1
	to := from + len(results) - 1
	if len(results) == 0 {
		from = 0
		to = 0
	}

	return &db.Paginator{
		Data:        results,
		Total:       total,
		PerPage:     perPage,
		CurrentPage: page,
		LastPage:    lastPage,
		From:        from,
		To:          to,
	}, nil
}

// GetQuery 获取底层查询构建器
func (mqb *ModelQueryBuilder) GetQuery() *db.QueryBuilder {
	return mqb.query
}

// GetModel 获取关联的模型
func (mqb *ModelQueryBuilder) GetModel() *BaseModel {
	return mqb.model
}
