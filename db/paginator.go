package db

import (
	"math"
)

// PaginationResult 分页结果
type PaginationResult struct {
	Data        []map[string]interface{} `json:"data"`         // 数据
	Total       int64                    `json:"total"`        // 总记录数
	PerPage     int                      `json:"per_page"`     // 每页数量
	CurrentPage int                      `json:"current_page"` // 当前页码
	LastPage    int                      `json:"last_page"`    // 最后一页
	HasMore     bool                     `json:"has_more"`     // 是否有更多
}

// Paginate 分页查询
func (qb *QueryBuilder) Paginate(page, perPage int) (*PaginationResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 15
	}

	// 获取总数（创建一个新的查询构建器副本用于计数）
	countBuilder := *qb
	countBuilder.selectColumns = []string{}
	countBuilder.orderByColumns = []OrderByClause{}
	countBuilder.limitCount = 0
	countBuilder.offsetCount = 0

	total, err := countBuilder.Count()
	if err != nil {
		return nil, err
	}

	// 计算分页信息
	lastPage := int(math.Ceil(float64(total) / float64(perPage)))
	offset := (page - 1) * perPage

	// 获取分页数据
	qb.Limit(perPage).Offset(offset)
	data, err := qb.GetRaw()
	if err != nil {
		return nil, err
	}

	return &PaginationResult{
		Data:        data,
		Total:       total,
		PerPage:     perPage,
		CurrentPage: page,
		LastPage:    lastPage,
		HasMore:     page < lastPage,
	}, nil
}

// SimplePaginate 简单分页（不计算总数，适用于大数据集）
func (qb *QueryBuilder) SimplePaginate(page, perPage int) ([]map[string]interface{}, bool, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 15
	}

	offset := (page - 1) * perPage

	// 多查询一条记录来判断是否还有更多数据
	qb.Limit(perPage + 1).Offset(offset)
	data, err := qb.GetRaw()
	if err != nil {
		return nil, false, err
	}

	hasMore := len(data) > perPage
	if hasMore {
		data = data[:perPage] // 移除多查询的那条记录
	}

	return data, hasMore, nil
}
