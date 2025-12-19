package db

import (
	"math"
)

// Paginator 完整分页器 - 包含总数和完整分页元数据
// 适用于需要显示总页数、总记录数的场景
type Paginator struct {
	Data        []map[string]interface{} `json:"data"`         // 当前页数据
	Total       int64                    `json:"total"`        // 总记录数
	PerPage     int                      `json:"per_page"`     // 每页记录数
	CurrentPage int                      `json:"current_page"` // 当前页码
	LastPage    int                      `json:"last_page"`    // 最后一页页码
	From        int                      `json:"from"`         // 当前页第一条记录的序号
	To          int                      `json:"to"`           // 当前页最后一条记录的序号
}

// SimplePaginator 简单分页器 - 不计算总数，适用于大数据集
// 只提供是否有下一页的信息，性能更好
type SimplePaginator struct {
	Data        []map[string]interface{} `json:"data"`         // 当前页数据
	PerPage     int                      `json:"per_page"`     // 每页记录数
	CurrentPage int                      `json:"current_page"` // 当前页码
	HasMore     bool                     `json:"has_more"`     // 是否有更多数据
}

// PaginationResult 分页结果（向下兼容旧代码）
type PaginationResult = SimplePagination

// NewPaginator 创建新的分页器
func NewPaginator(data []map[string]interface{}, total int64, page, perPage int) *Paginator {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 15
	}

	lastPage := int(math.Ceil(float64(total) / float64(perPage)))
	if lastPage < 1 {
		lastPage = 1
	}

	offset := (page - 1) * perPage
	from := offset + 1
	to := offset + len(data)

	// 如果没有数据，from 和 to 都应该是 0
	if len(data) == 0 {
		from = 0
		to = 0
	}

	return &Paginator{
		Data:        data,
		Total:       total,
		PerPage:     perPage,
		CurrentPage: page,
		LastPage:    lastPage,
		From:        from,
		To:          to,
	}
}

// NewSimplePaginator 创建新的简单分页器
func NewSimplePaginator(data []map[string]interface{}, page, perPage int, hasMore bool) *SimplePaginator {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 15
	}

	return &SimplePaginator{
		Data:        data,
		PerPage:     perPage,
		CurrentPage: page,
		HasMore:     hasMore,
	}
}

// ============================================================================
// Paginator 方法
// ============================================================================

// HasPages 是否有多页
func (p *Paginator) HasPages() bool {
	return p.LastPage > 1
}

// HasMorePages 是否有更多页
func (p *Paginator) HasMorePages() bool {
	return p.CurrentPage < p.LastPage
}

// OnFirstPage 是否在第一页
func (p *Paginator) OnFirstPage() bool {
	return p.CurrentPage <= 1
}

// OnLastPage 是否在最后一页
func (p *Paginator) OnLastPage() bool {
	return p.CurrentPage >= p.LastPage
}

// PreviousPage 上一页页码
func (p *Paginator) PreviousPage() int {
	if p.CurrentPage <= 1 {
		return 1
	}
	return p.CurrentPage - 1
}

// NextPage 下一页页码
func (p *Paginator) NextPage() int {
	if p.CurrentPage >= p.LastPage {
		return p.LastPage
	}
	return p.CurrentPage + 1
}

// Items 获取当前页数据
func (p *Paginator) Items() []map[string]interface{} {
	return p.Data
}

// Count 获取当前页记录数
func (p *Paginator) Count() int {
	return len(p.Data)
}

// IsEmpty 当前页是否为空
func (p *Paginator) IsEmpty() bool {
	return len(p.Data) == 0
}

// IsNotEmpty 当前页是否不为空
func (p *Paginator) IsNotEmpty() bool {
	return len(p.Data) > 0
}

// GetPageRange 获取页码范围（用于分页导航）
// 返回以当前页为中心的页码列表
func (p *Paginator) GetPageRange(onEachSide int) []int {
	if onEachSide < 1 {
		onEachSide = 3
	}

	start := p.CurrentPage - onEachSide
	if start < 1 {
		start = 1
	}

	end := p.CurrentPage + onEachSide
	if end > p.LastPage {
		end = p.LastPage
	}

	pages := make([]int, 0, end-start+1)
	for i := start; i <= end; i++ {
		pages = append(pages, i)
	}

	return pages
}

// ToMap 转换为map（用于JSON序列化）
func (p *Paginator) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"data":         p.Data,
		"total":        p.Total,
		"per_page":     p.PerPage,
		"current_page": p.CurrentPage,
		"last_page":    p.LastPage,
		"from":         p.From,
		"to":           p.To,
	}
}

// ============================================================================
// SimplePaginator 方法
// ============================================================================

// HasMorePages 是否有更多页
func (p *SimplePaginator) HasMorePages() bool {
	return p.HasMore
}

// OnFirstPage 是否在第一页
func (p *SimplePaginator) OnFirstPage() bool {
	return p.CurrentPage <= 1
}

// PreviousPage 上一页页码
func (p *SimplePaginator) PreviousPage() int {
	if p.CurrentPage <= 1 {
		return 1
	}
	return p.CurrentPage - 1
}

// NextPage 下一页页码（如果没有更多页则返回当前页）
func (p *SimplePaginator) NextPage() int {
	if !p.HasMore {
		return p.CurrentPage
	}
	return p.CurrentPage + 1
}

// Items 获取当前页数据
func (p *SimplePaginator) Items() []map[string]interface{} {
	return p.Data
}

// Count 获取当前页记录数
func (p *SimplePaginator) Count() int {
	return len(p.Data)
}

// IsEmpty 当前页是否为空
func (p *SimplePaginator) IsEmpty() bool {
	return len(p.Data) == 0
}

// IsNotEmpty 当前页是否不为空
func (p *SimplePaginator) IsNotEmpty() bool {
	return len(p.Data) > 0
}

// ToMap 转换为map（用于JSON序列化）
func (p *SimplePaginator) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"data":         p.Data,
		"per_page":     p.PerPage,
		"current_page": p.CurrentPage,
		"has_more":     p.HasMore,
	}
}

// ============================================================================
// Query 分页方法
// ============================================================================

// Paginate 分页查询 - 返回完整分页信息（包含总数）
// 适用于需要显示总页数的场景
func (q *Query) Paginate(page, perPage int) (*Paginator, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 15
	}

	// 克隆查询用于计数（避免修改原查询）
	countQuery := q.Clone()

	// 获取总数
	total, err := countQuery.Count()
	if err != nil {
		return nil, WrapError(err, ErrCodeQueryFailed, "pagination count failed")
	}

	// 计算分页信息
	offset := (page - 1) * perPage

	// 设置分页参数并获取数据
	q.Limit(perPage).Offset(offset)
	data, err := q.Get()
	if err != nil {
		return nil, WrapError(err, ErrCodeQueryFailed, "pagination query failed")
	}

	return NewPaginator(data, total, page, perPage), nil
}

// SimplePaginate 简单分页查询 - 不计算总数，适用于大数据集
// 只判断是否有下一页，性能更好
func (q *Query) SimplePaginate(page, perPage int) (*SimplePaginator, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 15
	}

	offset := (page - 1) * perPage

	// 多查询一条记录来判断是否还有更多数据
	q.Limit(perPage + 1).Offset(offset)
	data, err := q.Get()
	if err != nil {
		return nil, WrapError(err, ErrCodeQueryFailed, "simple pagination query failed")
	}

	hasMore := len(data) > perPage
	if hasMore {
		data = data[:perPage] // 移除多查询的那条记录
	}

	return NewSimplePaginator(data, page, perPage, hasMore), nil
}

// ============================================================================
// QueryBuilder 分页方法（向下兼容）
// ============================================================================

// Paginate 分页查询 - 返回完整分页信息（包含总数）
// 适用于需要显示总页数的场景
func (qb *QueryBuilder) Paginate(page, perPage int) (*SimplePagination, error) {
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
		return nil, WrapError(err, ErrCodeQueryFailed, "pagination count failed")
	}

	// 计算分页信息
	lastPage := int(math.Ceil(float64(total) / float64(perPage)))
	if lastPage < 1 {
		lastPage = 1
	}
	offset := (page - 1) * perPage
	from := offset + 1
	to := offset + perPage
	if to > int(total) {
		to = int(total)
	}

	// 获取分页数据
	qb.Limit(perPage).Offset(offset)
	data, err := qb.Get()
	if err != nil {
		return nil, WrapError(err, ErrCodeQueryFailed, "pagination query failed")
	}

	// 如果没有数据，from 和 to 都应该是 0
	if len(data) == 0 {
		from = 0
		to = 0
	}

	return &SimplePagination{
		Data:        data,
		Total:       total,
		PerPage:     perPage,
		CurrentPage: page,
		LastPage:    lastPage,
		From:        from,
		To:          to,
	}, nil
}

// SimplePaginate 简单分页（不计算总数，适用于大数据集）
// 返回数据、是否有更多数据、错误
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
		return nil, false, WrapError(err, ErrCodeQueryFailed, "simple pagination query failed")
	}

	hasMore := len(data) > perPage
	if hasMore {
		data = data[:perPage] // 移除多查询的那条记录
	}

	return data, hasMore, nil
}

// SimplePaginateWithPaginator 简单分页查询 - 返回SimplePaginator结构
// 适用于需要分页器对象的场景
func (qb *QueryBuilder) SimplePaginateWithPaginator(page, perPage int) (*SimplePaginator, error) {
	data, hasMore, err := qb.SimplePaginate(page, perPage)
	if err != nil {
		return nil, err
	}

	return NewSimplePaginator(data, page, perPage, hasMore), nil
}
