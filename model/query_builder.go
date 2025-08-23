package model

import (
	"context"
	"time"

	"github.com/zhoudm1743/torm/db"
)

// ModelQueryBuilder 模型查询构建器
type ModelQueryBuilder struct {
	query *db.QueryBuilder
	model *BaseModel
	err   error
}

// Where 添加查询条件
func (mqb *ModelQueryBuilder) Where(args ...interface{}) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.Where(args...)
	return mqb
}

// OrWhere 添加OR查询条件
func (mqb *ModelQueryBuilder) OrWhere(args ...interface{}) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.OrWhere(args...)
	return mqb
}

// WhereIn WHERE IN条件
func (mqb *ModelQueryBuilder) WhereIn(field string, values []interface{}) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.WhereIn(field, values)
	return mqb
}

// WhereNotIn WHERE NOT IN条件
func (mqb *ModelQueryBuilder) WhereNotIn(field string, values []interface{}) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.WhereNotIn(field, values)
	return mqb
}

// WhereBetween WHERE BETWEEN条件
func (mqb *ModelQueryBuilder) WhereBetween(field string, values []interface{}) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.WhereBetween(field, values)
	return mqb
}

// WhereNotBetween WHERE NOT BETWEEN条件
func (mqb *ModelQueryBuilder) WhereNotBetween(field string, values []interface{}) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.WhereNotBetween(field, values)
	return mqb
}

// WhereNull WHERE IS NULL条件
func (mqb *ModelQueryBuilder) WhereNull(field string) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.WhereNull(field)
	return mqb
}

// WhereNotNull WHERE IS NOT NULL条件
func (mqb *ModelQueryBuilder) WhereNotNull(field string) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.WhereNotNull(field)
	return mqb
}

// WhereExists WHERE EXISTS条件
func (mqb *ModelQueryBuilder) WhereExists(subQuery interface{}) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.WhereExists(subQuery)
	return mqb
}

// WhereNotExists WHERE NOT EXISTS条件
func (mqb *ModelQueryBuilder) WhereNotExists(subQuery interface{}) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.WhereNotExists(subQuery)
	return mqb
}

// WhereRaw 原生WHERE条件
func (mqb *ModelQueryBuilder) WhereRaw(raw string, bindings ...interface{}) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.WhereRaw(raw, bindings...)
	return mqb
}

// OrderBy 排序
func (mqb *ModelQueryBuilder) OrderBy(column, direction string) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.OrderBy(column, direction)
	return mqb
}

// Limit 限制数量
func (mqb *ModelQueryBuilder) Limit(limit int) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.Limit(limit)
	return mqb
}

// Offset 偏移量
func (mqb *ModelQueryBuilder) Offset(offset int) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.Offset(offset)
	return mqb
}

// GroupBy 分组查询
func (mqb *ModelQueryBuilder) GroupBy(columns ...string) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.GroupBy(columns...)
	return mqb
}

// Having HAVING条件
func (mqb *ModelQueryBuilder) Having(field string, operator string, value interface{}) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.Having(field, operator, value)
	return mqb
}

// HavingRaw 原生HAVING条件（支持多种参数格式）
func (mqb *ModelQueryBuilder) HavingRaw(args ...interface{}) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.HavingRaw(args...)
	return mqb
}

// Select 选择字段
func (mqb *ModelQueryBuilder) Select(columns ...string) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.Select(columns...)
	return mqb
}

// Join 内连接
func (mqb *ModelQueryBuilder) Join(table, localKey, operator, foreignKey string) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.Join(table, localKey, operator, foreignKey)
	return mqb
}

// LeftJoin 左连接
func (mqb *ModelQueryBuilder) LeftJoin(table, localKey, operator, foreignKey string) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.LeftJoin(table, localKey, operator, foreignKey)
	return mqb
}

// RightJoin 右连接
func (mqb *ModelQueryBuilder) RightJoin(table, localKey, operator, foreignKey string) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.RightJoin(table, localKey, operator, foreignKey)
	return mqb
}

// InnerJoin 内连接（别名）
func (mqb *ModelQueryBuilder) InnerJoin(table, localKey, operator, foreignKey string) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.InnerJoin(table, localKey, operator, foreignKey)
	return mqb
}

// First 获取第一条记录
func (mqb *ModelQueryBuilder) First() error {
	if mqb.err != nil {
		return mqb.err
	}

	result, err := mqb.query.First()
	if err != nil {
		return err
	}

	mqb.model.fill(result)
	mqb.model.exists = true
	return nil
}

// Get 获取多条记录
func (mqb *ModelQueryBuilder) Get() ([]map[string]interface{}, error) {
	if mqb.err != nil {
		return nil, mqb.err
	}

	return mqb.query.Get()
}

// Count 计算数量
func (mqb *ModelQueryBuilder) Count() (int64, error) {
	if mqb.err != nil {
		return 0, mqb.err
	}

	return mqb.query.Count()
}

// Update 更新记录
func (mqb *ModelQueryBuilder) Update(data map[string]interface{}) (int64, error) {
	if mqb.err != nil {
		return 0, mqb.err
	}

	// 处理时间戳
	if mqb.model.timestamps {
		data[mqb.model.updatedAt] = time.Now()
	}

	return mqb.query.Update(data)
}

// Delete 删除记录
func (mqb *ModelQueryBuilder) Delete() (int64, error) {
	if mqb.err != nil {
		return 0, mqb.err
	}

	if mqb.model.softDeletes {
		// 软删除
		data := map[string]interface{}{
			mqb.model.deletedAt: time.Now(),
		}
		return mqb.query.Update(data)
	} else {
		// 硬删除
		return mqb.query.Delete()
	}
}

// Paginate 分页查询
func (mqb *ModelQueryBuilder) Paginate(page, perPage int) (*db.PaginationResult, error) {
	if mqb.err != nil {
		return nil, mqb.err
	}

	return mqb.query.Paginate(page, perPage)
}

// SelectRaw 原生SELECT语句
func (mqb *ModelQueryBuilder) SelectRaw(raw string, bindings ...interface{}) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.SelectRaw(raw, bindings...)
	return mqb
}

// FieldRaw 原生字段表达式
func (mqb *ModelQueryBuilder) FieldRaw(raw string, bindings ...interface{}) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.FieldRaw(raw, bindings...)
	return mqb
}

// Distinct 去重查询
func (mqb *ModelQueryBuilder) Distinct() *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.Distinct()
	return mqb
}

// OrderByRaw 原生排序
func (mqb *ModelQueryBuilder) OrderByRaw(raw string, bindings ...interface{}) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.OrderByRaw(raw, bindings...)
	return mqb
}

// OrderRand 随机排序
func (mqb *ModelQueryBuilder) OrderRand() *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.OrderRand()
	return mqb
}

// OrderField 字段排序
func (mqb *ModelQueryBuilder) OrderField(field string, values []interface{}, direction string) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.OrderField(field, values, direction)
	return mqb
}

// Page 分页设置
func (mqb *ModelQueryBuilder) Page(page, pageSize int) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.Page(page, pageSize)
	return mqb
}

// WithContext 设置上下文
func (mqb *ModelQueryBuilder) WithContext(ctx context.Context) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.WithContext(ctx)
	return mqb
}

// WithTimeout 设置超时
func (mqb *ModelQueryBuilder) WithTimeout(timeout time.Duration) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.WithTimeout(timeout)
	return mqb
}

// Cache 启用查询缓存
func (mqb *ModelQueryBuilder) Cache(ttl time.Duration) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.Cache(ttl)
	return mqb
}

// CacheWithTags 启用带标签的查询缓存
func (mqb *ModelQueryBuilder) CacheWithTags(ttl time.Duration, tags ...string) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.CacheWithTags(ttl, tags...)
	return mqb
}

// CacheKey 设置自定义缓存键
func (mqb *ModelQueryBuilder) CacheKey(key string) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.CacheKey(key)
	return mqb
}

// Find 根据ID查找
func (mqb *ModelQueryBuilder) Find(args ...interface{}) (map[string]interface{}, error) {
	if mqb.err != nil {
		return nil, mqb.err
	}

	return mqb.query.Find(args...)
}

// Exists 检查记录是否存在
func (mqb *ModelQueryBuilder) Exists() (bool, error) {
	if mqb.err != nil {
		return false, mqb.err
	}

	return mqb.query.Exists()
}

// Insert 插入数据
func (mqb *ModelQueryBuilder) Insert(data map[string]interface{}) (int64, error) {
	if mqb.err != nil {
		return 0, mqb.err
	}

	// 处理时间戳
	if mqb.model.timestamps {
		now := time.Now()
		data[mqb.model.createdAt] = now
		data[mqb.model.updatedAt] = now
	}

	return mqb.query.Insert(data)
}

// InsertBatch 批量插入数据
func (mqb *ModelQueryBuilder) InsertBatch(data []map[string]interface{}) (int64, error) {
	if mqb.err != nil {
		return 0, mqb.err
	}

	// 处理时间戳
	if mqb.model.timestamps {
		now := time.Now()
		for i := range data {
			data[i][mqb.model.createdAt] = now
			data[i][mqb.model.updatedAt] = now
		}
	}

	return mqb.query.InsertBatch(data)
}

// Exp 高级表达式
func (mqb *ModelQueryBuilder) Exp(field string, expression string, bindings ...interface{}) *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}
	mqb.query = mqb.query.Exp(field, expression, bindings...)
	return mqb
}

// ToSQL 构建SQL语句
func (mqb *ModelQueryBuilder) ToSQL() (string, []interface{}, error) {
	if mqb.err != nil {
		return "", nil, mqb.err
	}

	return mqb.query.ToSQL()
}

// Clone 克隆查询构建器
func (mqb *ModelQueryBuilder) Clone() *ModelQueryBuilder {
	if mqb.err != nil {
		return mqb
	}

	return &ModelQueryBuilder{
		query: mqb.query.Clone(),
		model: mqb.model,
		err:   nil,
	}
}
