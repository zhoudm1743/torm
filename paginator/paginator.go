package paginator

import (
	"context"
	"math"
	"strconv"

	"github.com/zhoudm1743/torm/db"
)

// PaginatorInterface 分页器接口
type PaginatorInterface interface {
	// 获取当前页数据
	Items() []interface{}
	// 获取总记录数
	Total() int64
	// 获取每页记录数
	PerPage() int
	// 获取当前页码
	CurrentPage() int
	// 获取最后一页页码
	LastPage() int
	// 是否有下一页
	HasMore() bool
	// 是否有上一页
	HasPrev() bool
	// 获取下一页URL
	NextPageUrl() string
	// 获取上一页URL
	PrevPageUrl() string
	// 获取分页数据的JSON表示
	ToMap() map[string]interface{}
}

// SimplePaginator 简单分页器
type SimplePaginator struct {
	items       []interface{}     `json:"data"`
	perPage     int               `json:"per_page"`
	currentPage int               `json:"current_page"`
	total       int64             `json:"total"`
	lastPage    int               `json:"last_page"`
	hasMore     bool              `json:"has_more"`
	path        string            `json:"path"`
	query       map[string]string `json:"-"`
}

// NewSimplePaginator 创建简单分页器
func NewSimplePaginator(items []interface{}, total int64, perPage, currentPage int) *SimplePaginator {
	lastPage := int(math.Ceil(float64(total) / float64(perPage)))
	hasMore := currentPage < lastPage

	return &SimplePaginator{
		items:       items,
		perPage:     perPage,
		currentPage: currentPage,
		total:       total,
		lastPage:    lastPage,
		hasMore:     hasMore,
		path:        "/",
		query:       make(map[string]string),
	}
}

// Items 获取当前页数据
func (p *SimplePaginator) Items() []interface{} {
	return p.items
}

// Total 获取总记录数
func (p *SimplePaginator) Total() int64 {
	return p.total
}

// PerPage 获取每页记录数
func (p *SimplePaginator) PerPage() int {
	return p.perPage
}

// CurrentPage 获取当前页码
func (p *SimplePaginator) CurrentPage() int {
	return p.currentPage
}

// LastPage 获取最后一页页码
func (p *SimplePaginator) LastPage() int {
	return p.lastPage
}

// HasMore 是否有下一页
func (p *SimplePaginator) HasMore() bool {
	return p.hasMore
}

// HasPrev 是否有上一页
func (p *SimplePaginator) HasPrev() bool {
	return p.currentPage > 1
}

// NextPageUrl 获取下一页URL
func (p *SimplePaginator) NextPageUrl() string {
	if !p.HasMore() {
		return ""
	}
	return p.buildUrl(p.currentPage + 1)
}

// PrevPageUrl 获取上一页URL
func (p *SimplePaginator) PrevPageUrl() string {
	if !p.HasPrev() {
		return ""
	}
	return p.buildUrl(p.currentPage - 1)
}

// buildUrl 构建URL
func (p *SimplePaginator) buildUrl(page int) string {
	if p.path == "" {
		p.path = "/"
	}

	url := p.path + "?page=" + strconv.Itoa(page)
	for key, value := range p.query {
		url += "&" + key + "=" + value
	}

	return url
}

// SetPath 设置路径
func (p *SimplePaginator) SetPath(path string) *SimplePaginator {
	p.path = path
	return p
}

// AppendQuery 添加查询参数
func (p *SimplePaginator) AppendQuery(key, value string) *SimplePaginator {
	p.query[key] = value
	return p
}

// ToMap 转换为map
func (p *SimplePaginator) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"data":          p.items,
		"total":         p.total,
		"per_page":      p.perPage,
		"current_page":  p.currentPage,
		"last_page":     p.lastPage,
		"has_more":      p.hasMore,
		"has_prev":      p.HasPrev(),
		"next_page_url": p.NextPageUrl(),
		"prev_page_url": p.PrevPageUrl(),
	}
}

// QueryPaginator 查询分页器 - 集成到查询构造器
type QueryPaginator struct {
	query   db.QueryInterface
	ctx     context.Context
	perPage int
	page    int
}

// NewQueryPaginator 创建查询分页器
func NewQueryPaginator(query db.QueryInterface, ctx context.Context) *QueryPaginator {
	return &QueryPaginator{
		query:   query,
		ctx:     ctx,
		perPage: 15, // 默认每页15条
		page:    1,  // 默认第1页
	}
}

// SetPerPage 设置每页记录数
func (qp *QueryPaginator) SetPerPage(perPage int) *QueryPaginator {
	qp.perPage = perPage
	return qp
}

// SetPage 设置当前页
func (qp *QueryPaginator) SetPage(page int) *QueryPaginator {
	if page < 1 {
		page = 1
	}
	qp.page = page
	return qp
}

// Paginate 执行分页查询
func (qp *QueryPaginator) Paginate() (PaginatorInterface, error) {
	// 计算总数
	total, err := qp.query.Count()
	if err != nil {
		return nil, err
	}

	// 计算偏移量
	offset := (qp.page - 1) * qp.perPage

	// 执行查询
	data, err := qp.query.Limit(qp.perPage).Offset(offset).Get()
	if err != nil {
		return nil, err
	}

	// 转换数据格式
	items := make([]interface{}, len(data))
	for i, item := range data {
		items[i] = item
	}

	return NewSimplePaginator(items, total, qp.perPage, qp.page), nil
}

// CursorPaginator 游标分页器 - 适用于大数据量
type CursorPaginator struct {
	items      []interface{} `json:"data"`
	perPage    int           `json:"per_page"`
	nextCursor string        `json:"next_cursor,omitempty"`
	prevCursor string        `json:"prev_cursor,omitempty"`
	hasMore    bool          `json:"has_more"`
}

// NewCursorPaginator 创建游标分页器
func NewCursorPaginator(items []interface{}, perPage int, nextCursor, prevCursor string) *CursorPaginator {
	return &CursorPaginator{
		items:      items,
		perPage:    perPage,
		nextCursor: nextCursor,
		prevCursor: prevCursor,
		hasMore:    len(items) == perPage, // 如果返回的数据等于每页数量，说明可能还有更多数据
	}
}

// Items 获取当前页数据
func (cp *CursorPaginator) Items() []interface{} {
	return cp.items
}

// PerPage 获取每页记录数
func (cp *CursorPaginator) PerPage() int {
	return cp.perPage
}

// NextCursor 获取下一页游标
func (cp *CursorPaginator) NextCursor() string {
	return cp.nextCursor
}

// PrevCursor 获取上一页游标
func (cp *CursorPaginator) PrevCursor() string {
	return cp.prevCursor
}

// HasMore 是否有更多数据
func (cp *CursorPaginator) HasMore() bool {
	return cp.hasMore
}

// ToMap 转换为map
func (cp *CursorPaginator) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"data":     cp.items,
		"per_page": cp.perPage,
		"has_more": cp.hasMore,
	}

	if cp.nextCursor != "" {
		result["next_cursor"] = cp.nextCursor
	}
	if cp.prevCursor != "" {
		result["prev_cursor"] = cp.prevCursor
	}

	return result
}

// PaginateConfig 分页配置
type PaginateConfig struct {
	PerPage         int               `json:"per_page"`
	PageName        string            `json:"page_name"`
	Path            string            `json:"path"`
	Query           map[string]string `json:"query"`
	Fragment        string            `json:"fragment"`
	OnEachSide      int               `json:"on_each_side"`      // 当前页两侧显示的页码数
	ShowQuickJumper bool              `json:"show_quick_jumper"` // 是否显示快速跳转
}

// DefaultPaginateConfig 默认分页配置
func DefaultPaginateConfig() *PaginateConfig {
	return &PaginateConfig{
		PerPage:         15,
		PageName:        "page",
		Path:            "/",
		Query:           make(map[string]string),
		Fragment:        "",
		OnEachSide:      3,
		ShowQuickJumper: true,
	}
}
