package model

import (
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
