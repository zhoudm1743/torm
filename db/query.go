package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Query 查询构建器 - 符合设计文档的新实现
// 提供流畅的链式API用于构建SQL查询
type Query struct {
	connection Connection
	builder    Builder

	// 查询组件
	table      string
	alias      map[string]string
	fields     []string
	selectRaws []QuerySelectRaw // 原生SELECT表达式
	distinct   bool
	joins      []QueryJoinClause
	wheres     []QueryWhereClause
	groups     []string
	havings    []QueryWhereClause
	orders     []QueryOrderClause
	limit      int
	offset     int

	// 绑定参数
	bindings []interface{}

	// 缓存
	cacheKey  string
	cacheTTL  time.Duration
	cacheTags []string

	// 上下文
	ctx        context.Context
	baseCtx    context.Context    // 基础上下文，用于创建超时/截止时间上下文
	cancelFunc context.CancelFunc // 用于取消上下文

	// 模型
	model interface{}

	// 事务
	tx TxInterface
}

// QueryWhereClause WHERE子句
type QueryWhereClause struct {
	Type     string        // AND, OR
	Field    string        // 字段名
	Operator string        // 操作符
	Value    interface{}   // 值
	Raw      string        // 原生SQL
	Bindings []interface{} // 原生SQL的参数
	Nested   []QueryWhereClause
}

// QueryJoinClause JOIN子句
type QueryJoinClause struct {
	Type      string        // INNER, LEFT, RIGHT, CROSS
	Table     string        // 表名
	Condition string        // 条件
	Bindings  []interface{} // 绑定参数
}

// QueryOrderClause ORDER子句
type QueryOrderClause struct {
	Field     string // 字段名
	Direction string // ASC, DESC
	Raw       string // 原生排序表达式
}

// QuerySelectRaw 原生SELECT表达式
type QuerySelectRaw struct {
	Expression string        // 原生表达式
	Bindings   []interface{} // 绑定参数
}

// NewQuery 创建新的查询构建器
func NewQuery(conn Connection) *Query {
	baseCtx := context.Background()
	q := &Query{
		connection: conn,
		alias:      make(map[string]string),
		fields:     make([]string, 0),
		selectRaws: make([]QuerySelectRaw, 0),
		joins:      make([]QueryJoinClause, 0),
		wheres:     make([]QueryWhereClause, 0),
		groups:     make([]string, 0),
		havings:    make([]QueryWhereClause, 0),
		orders:     make([]QueryOrderClause, 0),
		bindings:   make([]interface{}, 0),
		cacheTags:  make([]string, 0),
		ctx:        baseCtx,
		baseCtx:    baseCtx,
	}

	if conn != nil {
		q.builder = conn.GetBuilder()
	}

	return q
}

// Table 设置查询表名
func (q *Query) Table(name string) *Query {
	q.table = name
	return q
}

// Select 设置选择的字段
func (q *Query) Select(fields ...string) *Query {
	q.fields = append(q.fields, fields...)
	return q
}

// SelectRaw 添加原生SELECT表达式
// 支持参数绑定，例如: SelectRaw("COUNT(*) as cnt")
// 或: SelectRaw("COALESCE(?, ?) as value", defaultVal, fallbackVal)
// Requirements: 11.4
func (q *Query) SelectRaw(raw string, bindings ...interface{}) *Query {
	q.selectRaws = append(q.selectRaws, QuerySelectRaw{
		Expression: raw,
		Bindings:   bindings,
	})
	return q
}

// FieldRaw 添加原生字段表达式（SelectRaw的别名）
// Requirements: 11.4
func (q *Query) FieldRaw(raw string, bindings ...interface{}) *Query {
	return q.SelectRaw(raw, bindings...)
}

// Distinct 设置DISTINCT查询
func (q *Query) Distinct() *Query {
	q.distinct = true
	return q
}

// GetTable 获取表名
func (q *Query) GetTable() string {
	return q.table
}

// GetFields 获取选择的字段
func (q *Query) GetFields() []string {
	return q.fields
}

// GetSelectRaws 获取原生SELECT表达式
func (q *Query) GetSelectRaws() []QuerySelectRaw {
	return q.selectRaws
}

// IsDistinct 是否DISTINCT查询
func (q *Query) IsDistinct() bool {
	return q.distinct
}

// GetConnection 获取数据库连接
func (q *Query) GetConnection() Connection {
	return q.connection
}

// SetConnection 设置数据库连接
func (q *Query) SetConnection(conn Connection) *Query {
	q.connection = conn
	if conn != nil {
		q.builder = conn.GetBuilder()
	}
	return q
}

// Clone 克隆查询构建器
func (q *Query) Clone() *Query {
	newQuery := &Query{
		connection: q.connection,
		builder:    q.builder,
		table:      q.table,
		alias:      make(map[string]string),
		fields:     make([]string, len(q.fields)),
		selectRaws: make([]QuerySelectRaw, len(q.selectRaws)),
		distinct:   q.distinct,
		joins:      make([]QueryJoinClause, len(q.joins)),
		wheres:     make([]QueryWhereClause, len(q.wheres)),
		groups:     make([]string, len(q.groups)),
		havings:    make([]QueryWhereClause, len(q.havings)),
		orders:     make([]QueryOrderClause, len(q.orders)),
		limit:      q.limit,
		offset:     q.offset,
		bindings:   make([]interface{}, len(q.bindings)),
		cacheKey:   q.cacheKey,
		cacheTTL:   q.cacheTTL,
		cacheTags:  make([]string, len(q.cacheTags)),
		ctx:        q.ctx,
		baseCtx:    q.baseCtx,
		cancelFunc: nil, // 克隆不复制cancel函数，避免多次取消
		model:      q.model,
		tx:         q.tx,
	}

	// 复制map
	for k, v := range q.alias {
		newQuery.alias[k] = v
	}

	// 复制切片
	copy(newQuery.fields, q.fields)
	copy(newQuery.selectRaws, q.selectRaws)
	copy(newQuery.joins, q.joins)
	copy(newQuery.wheres, q.wheres)
	copy(newQuery.groups, q.groups)
	copy(newQuery.havings, q.havings)
	copy(newQuery.orders, q.orders)
	copy(newQuery.bindings, q.bindings)
	copy(newQuery.cacheTags, q.cacheTags)

	return newQuery
}

// Reset 重置查询状态
func (q *Query) Reset() *Query {
	q.table = ""
	q.alias = make(map[string]string)
	q.fields = make([]string, 0)
	q.selectRaws = make([]QuerySelectRaw, 0)
	q.distinct = false
	q.joins = make([]QueryJoinClause, 0)
	q.wheres = make([]QueryWhereClause, 0)
	q.groups = make([]string, 0)
	q.havings = make([]QueryWhereClause, 0)
	q.orders = make([]QueryOrderClause, 0)
	q.limit = 0
	q.offset = 0
	q.bindings = make([]interface{}, 0)
	q.cacheKey = ""
	q.cacheTTL = 0
	q.cacheTags = make([]string, 0)
	q.model = nil
	return q
}

// WithContext 设置上下文
// 支持传递context.Context用于查询取消和超时控制
// Requirements: 15.1
func (q *Query) WithContext(ctx context.Context) *Query {
	// 如果之前有cancel函数，先调用它以释放资源
	if q.cancelFunc != nil {
		q.cancelFunc()
		q.cancelFunc = nil
	}
	q.ctx = ctx
	q.baseCtx = ctx // 更新基础上下文
	return q
}

// WithTimeout 设置超时
// 创建一个带超时的上下文，超时后查询将被取消
// Requirements: 15.3
func (q *Query) WithTimeout(timeout time.Duration) *Query {
	// 如果之前有cancel函数，先调用它以释放资源
	if q.cancelFunc != nil {
		q.cancelFunc()
	}
	// 从基础上下文创建超时上下文，避免上下文链问题
	ctx, cancel := context.WithTimeout(q.baseCtx, timeout)
	q.ctx = ctx
	q.cancelFunc = cancel
	return q
}

// WithDeadline 设置截止时间
// 创建一个带截止时间的上下文，到达截止时间后查询将被取消
func (q *Query) WithDeadline(deadline time.Time) *Query {
	// 如果之前有cancel函数，先调用它以释放资源
	if q.cancelFunc != nil {
		q.cancelFunc()
	}
	// 从基础上下文创建截止时间上下文，避免上下文链问题
	ctx, cancel := context.WithDeadline(q.baseCtx, deadline)
	q.ctx = ctx
	q.cancelFunc = cancel
	return q
}

// Cancel 取消当前查询的上下文
// 如果设置了超时或截止时间，调用此方法可以提前取消
func (q *Query) Cancel() {
	if q.cancelFunc != nil {
		q.cancelFunc()
		q.cancelFunc = nil
	}
}

// GetContext 获取当前上下文
func (q *Query) GetContext() context.Context {
	return q.ctx
}

// Model 绑定模型
func (q *Query) Model(model interface{}) *Query {
	q.model = model
	// 自动从模型获取表名
	if q.table == "" {
		q.table = getTableNameFromModel(model)
	}
	return q
}

// InTransaction 在事务中执行
func (q *Query) InTransaction(tx TxInterface) *Query {
	q.tx = tx
	return q
}

// Cache 启用查询缓存
func (q *Query) Cache(ttl time.Duration, key ...string) *Query {
	q.cacheTTL = ttl
	if len(key) > 0 {
		q.cacheKey = key[0]
	}
	return q
}

// CacheWithTags 启用带标签的查询缓存
func (q *Query) CacheWithTags(ttl time.Duration, tags ...string) *Query {
	q.cacheTTL = ttl
	q.cacheTags = tags
	return q
}

// getDriverName 获取数据库驱动名称
func (q *Query) getDriverName() string {
	if q.connection != nil {
		return q.connection.GetDriver()
	}
	return ""
}

// buildPlaceholder 根据数据库类型构建占位符
func (q *Query) buildPlaceholder(index int) string {
	driverName := q.getDriverName()
	switch driverName {
	case "postgres", "postgresql", "pq":
		return fmt.Sprintf("$%d", index+1)
	case "sqlserver", "mssql":
		return fmt.Sprintf("@p%d", index+1)
	default:
		return "?"
	}
}

// processPlaceholders 处理原始SQL中的占位符
func (q *Query) processPlaceholders(sqlStr string, startIndex int) string {
	driverName := q.getDriverName()

	switch driverName {
	case "postgres", "postgresql", "pq":
		result := sqlStr
		placeholderCount := strings.Count(sqlStr, "?")
		for i := 0; i < placeholderCount; i++ {
			placeholder := fmt.Sprintf("$%d", startIndex+i+1)
			result = strings.Replace(result, "?", placeholder, 1)
		}
		return result
	case "sqlserver", "mssql":
		result := sqlStr
		placeholderCount := strings.Count(sqlStr, "?")
		for i := 0; i < placeholderCount; i++ {
			placeholder := fmt.Sprintf("@p%d", startIndex+i+1)
			result = strings.Replace(result, "?", placeholder, 1)
		}
		return result
	default:
		return sqlStr
	}
}

// executeQuery 执行查询
// 支持上下文取消和超时处理
// Requirements: 15.2, 15.4
func (q *Query) executeQuery(sqlStr string, args []interface{}) (*sql.Rows, error) {
	// 检查上下文是否已取消
	if err := q.ctx.Err(); err != nil {
		return nil, q.wrapContextError(err, sqlStr, args)
	}

	var rows *sql.Rows
	var err error

	if q.tx != nil {
		rows, err = q.tx.QueryContext(q.ctx, sqlStr, args...)
	} else if q.connection != nil {
		rows, err = q.connection.QueryContext(q.ctx, sqlStr, args...)
	} else {
		return nil, NewError(ErrCodeConnectionClosed, "no database connection")
	}

	// 处理上下文相关错误
	if err != nil {
		return nil, q.wrapContextError(err, sqlStr, args)
	}

	return rows, nil
}

// executeExec 执行SQL语句
// 支持上下文取消和超时处理
// Requirements: 15.2, 15.4
func (q *Query) executeExec(sqlStr string, args []interface{}) (sql.Result, error) {
	// 检查上下文是否已取消
	if err := q.ctx.Err(); err != nil {
		return nil, q.wrapContextError(err, sqlStr, args)
	}

	var result sql.Result
	var err error

	if q.tx != nil {
		result, err = q.tx.ExecContext(q.ctx, sqlStr, args...)
	} else if q.connection != nil {
		result, err = q.connection.ExecContext(q.ctx, sqlStr, args...)
	} else {
		return nil, NewError(ErrCodeConnectionClosed, "no database connection")
	}

	// 处理上下文相关错误
	if err != nil {
		return nil, q.wrapContextError(err, sqlStr, args)
	}

	return result, nil
}

// wrapContextError 包装上下文相关错误
// 将context.Canceled和context.DeadlineExceeded转换为TORM错误
func (q *Query) wrapContextError(err error, sqlStr string, args []interface{}) error {
	if err == nil {
		return nil
	}

	// 检查是否是上下文取消错误
	if err == context.Canceled {
		return NewError(ErrCodeCanceled, "query canceled").
			WithSQL(sqlStr, args...).
			WithContext("table", q.table).
			WithCause(err)
	}

	// 检查是否是超时错误
	if err == context.DeadlineExceeded {
		return NewError(ErrCodeTimeout, "query timeout").
			WithSQL(sqlStr, args...).
			WithContext("table", q.table).
			WithCause(err)
	}

	// 检查底层错误是否包含上下文错误
	if q.ctx.Err() == context.Canceled {
		return NewError(ErrCodeCanceled, "query canceled").
			WithSQL(sqlStr, args...).
			WithContext("table", q.table).
			WithCause(err)
	}

	if q.ctx.Err() == context.DeadlineExceeded {
		return NewError(ErrCodeTimeout, "query timeout").
			WithSQL(sqlStr, args...).
			WithContext("table", q.table).
			WithCause(err)
	}

	// 返回原始错误
	return err
}

// scanRows 扫描行数据
func (q *Query) scanRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, column := range columns {
			row[column] = q.convertValue(values[i])
		}

		results = append(results, row)
	}

	return results, rows.Err()
}

// convertValue 转换数据库值
func (q *Query) convertValue(value interface{}) interface{} {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return string(v)
	default:
		return v
	}
}

// ============================================================================
// WHERE条件方法
// ============================================================================

// Where 添加WHERE条件 - 支持多种格式
// Where("name", "John") - 等于条件
// Where("age", ">", 18) - 带操作符
// Where("status = ?", "active") - 原生SQL
func (q *Query) Where(args ...interface{}) *Query {
	switch len(args) {
	case 1:
		// Where("id IN (1,2,3)") - 纯SQL
		if sql, ok := args[0].(string); ok {
			q.wheres = append(q.wheres, QueryWhereClause{
				Type: "AND",
				Raw:  sql,
			})
		}
	case 2:
		// Where("name", "John") 或 Where("status = ?", "active")
		if field, ok := args[0].(string); ok {
			if strings.Contains(field, "?") || strings.Contains(field, " ") {
				// 原生SQL带参数
				q.wheres = append(q.wheres, QueryWhereClause{
					Type:     "AND",
					Raw:      field,
					Bindings: []interface{}{args[1]},
				})
			} else {
				// 简单等于条件
				q.wheres = append(q.wheres, QueryWhereClause{
					Type:     "AND",
					Field:    field,
					Operator: "=",
					Value:    args[1],
				})
			}
		}
	case 3:
		// Where("age", ">", 18)
		if field, ok := args[0].(string); ok {
			if operator, ok := args[1].(string); ok {
				q.wheres = append(q.wheres, QueryWhereClause{
					Type:     "AND",
					Field:    field,
					Operator: operator,
					Value:    args[2],
				})
			}
		}
	default:
		// Where("status IN (?, ?)", "active", "pending") - 多参数
		if len(args) > 1 {
			if sql, ok := args[0].(string); ok {
				q.wheres = append(q.wheres, QueryWhereClause{
					Type:     "AND",
					Raw:      sql,
					Bindings: args[1:],
				})
			}
		}
	}
	return q
}

// OrWhere 添加OR WHERE条件
func (q *Query) OrWhere(args ...interface{}) *Query {
	switch len(args) {
	case 1:
		if sql, ok := args[0].(string); ok {
			q.wheres = append(q.wheres, QueryWhereClause{
				Type: "OR",
				Raw:  sql,
			})
		}
	case 2:
		if field, ok := args[0].(string); ok {
			if strings.Contains(field, "?") || strings.Contains(field, " ") {
				q.wheres = append(q.wheres, QueryWhereClause{
					Type:     "OR",
					Raw:      field,
					Bindings: []interface{}{args[1]},
				})
			} else {
				q.wheres = append(q.wheres, QueryWhereClause{
					Type:     "OR",
					Field:    field,
					Operator: "=",
					Value:    args[1],
				})
			}
		}
	case 3:
		if field, ok := args[0].(string); ok {
			if operator, ok := args[1].(string); ok {
				q.wheres = append(q.wheres, QueryWhereClause{
					Type:     "OR",
					Field:    field,
					Operator: operator,
					Value:    args[2],
				})
			}
		}
	default:
		if len(args) > 1 {
			if sql, ok := args[0].(string); ok {
				q.wheres = append(q.wheres, QueryWhereClause{
					Type:     "OR",
					Raw:      sql,
					Bindings: args[1:],
				})
			}
		}
	}
	return q
}

// WhereIn WHERE IN条件
func (q *Query) WhereIn(field string, values interface{}) *Query {
	vals := q.toInterfaceSlice(values)
	if len(vals) == 0 {
		return q
	}

	placeholders := make([]string, len(vals))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	q.wheres = append(q.wheres, QueryWhereClause{
		Type:     "AND",
		Raw:      fmt.Sprintf("%s IN (%s)", field, strings.Join(placeholders, ", ")),
		Bindings: vals,
	})
	return q
}

// WhereNotIn WHERE NOT IN条件
func (q *Query) WhereNotIn(field string, values interface{}) *Query {
	vals := q.toInterfaceSlice(values)
	if len(vals) == 0 {
		return q
	}

	placeholders := make([]string, len(vals))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	q.wheres = append(q.wheres, QueryWhereClause{
		Type:     "AND",
		Raw:      fmt.Sprintf("%s NOT IN (%s)", field, strings.Join(placeholders, ", ")),
		Bindings: vals,
	})
	return q
}

// WhereBetween WHERE BETWEEN条件
func (q *Query) WhereBetween(field string, min, max interface{}) *Query {
	q.wheres = append(q.wheres, QueryWhereClause{
		Type:     "AND",
		Raw:      fmt.Sprintf("%s BETWEEN ? AND ?", field),
		Bindings: []interface{}{min, max},
	})
	return q
}

// WhereNotBetween WHERE NOT BETWEEN条件
func (q *Query) WhereNotBetween(field string, min, max interface{}) *Query {
	q.wheres = append(q.wheres, QueryWhereClause{
		Type:     "AND",
		Raw:      fmt.Sprintf("%s NOT BETWEEN ? AND ?", field),
		Bindings: []interface{}{min, max},
	})
	return q
}

// WhereNull WHERE IS NULL条件
func (q *Query) WhereNull(field string) *Query {
	q.wheres = append(q.wheres, QueryWhereClause{
		Type: "AND",
		Raw:  fmt.Sprintf("%s IS NULL", field),
	})
	return q
}

// WhereNotNull WHERE IS NOT NULL条件
func (q *Query) WhereNotNull(field string) *Query {
	q.wheres = append(q.wheres, QueryWhereClause{
		Type: "AND",
		Raw:  fmt.Sprintf("%s IS NOT NULL", field),
	})
	return q
}

// WhereRaw 原生WHERE条件
func (q *Query) WhereRaw(raw string, bindings ...interface{}) *Query {
	q.wheres = append(q.wheres, QueryWhereClause{
		Type:     "AND",
		Raw:      raw,
		Bindings: bindings,
	})
	return q
}

// toInterfaceSlice 将各种类型的切片转换为[]interface{}
func (q *Query) toInterfaceSlice(value interface{}) []interface{} {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []interface{}:
		return v
	case []string:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	case []int:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	case []int64:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	default:
		return []interface{}{value}
	}
}

// ============================================================================
// JOIN操作方法
// ============================================================================

// Join 内连接
// Join("users", "users.id = posts.user_id")
// Join("users", "users.id", "=", "posts.user_id")
func (q *Query) Join(table string, condition string, args ...interface{}) *Query {
	return q.addJoin("INNER", table, condition, args...)
}

// LeftJoin 左连接
func (q *Query) LeftJoin(table string, condition string, args ...interface{}) *Query {
	return q.addJoin("LEFT", table, condition, args...)
}

// RightJoin 右连接
func (q *Query) RightJoin(table string, condition string, args ...interface{}) *Query {
	return q.addJoin("RIGHT", table, condition, args...)
}

// InnerJoin 内连接（与Join相同）
func (q *Query) InnerJoin(table string, condition string, args ...interface{}) *Query {
	return q.addJoin("INNER", table, condition, args...)
}

// CrossJoin 交叉连接
func (q *Query) CrossJoin(table string) *Query {
	q.joins = append(q.joins, QueryJoinClause{
		Type:  "CROSS",
		Table: table,
	})
	return q
}

// addJoin 内部方法 - 添加JOIN子句
func (q *Query) addJoin(joinType, table, condition string, args ...interface{}) *Query {
	q.joins = append(q.joins, QueryJoinClause{
		Type:      joinType,
		Table:     table,
		Condition: condition,
		Bindings:  args,
	})
	return q
}

// GetJoins 获取JOIN子句
func (q *Query) GetJoins() []QueryJoinClause {
	return q.joins
}

// ============================================================================
// 排序和分组方法
// ============================================================================

// OrderBy 排序
func (q *Query) OrderBy(field string, direction ...string) *Query {
	dir := "ASC"
	if len(direction) > 0 {
		dir = strings.ToUpper(direction[0])
		if dir != "ASC" && dir != "DESC" {
			dir = "ASC"
		}
	}

	q.orders = append(q.orders, QueryOrderClause{
		Field:     field,
		Direction: dir,
	})
	return q
}

// OrderByRaw 原生排序
func (q *Query) OrderByRaw(raw string, bindings ...interface{}) *Query {
	q.orders = append(q.orders, QueryOrderClause{
		Raw: raw,
	})
	q.bindings = append(q.bindings, bindings...)
	return q
}

// GroupBy 分组
func (q *Query) GroupBy(fields ...string) *Query {
	q.groups = append(q.groups, fields...)
	return q
}

// Having HAVING条件
func (q *Query) Having(args ...interface{}) *Query {
	switch len(args) {
	case 1:
		if sql, ok := args[0].(string); ok {
			q.havings = append(q.havings, QueryWhereClause{
				Type: "AND",
				Raw:  sql,
			})
		}
	case 2:
		if field, ok := args[0].(string); ok {
			if strings.Contains(field, "?") || strings.Contains(field, " ") {
				q.havings = append(q.havings, QueryWhereClause{
					Type:     "AND",
					Raw:      field,
					Bindings: []interface{}{args[1]},
				})
			} else {
				q.havings = append(q.havings, QueryWhereClause{
					Type:     "AND",
					Field:    field,
					Operator: "=",
					Value:    args[1],
				})
			}
		}
	case 3:
		if field, ok := args[0].(string); ok {
			if operator, ok := args[1].(string); ok {
				q.havings = append(q.havings, QueryWhereClause{
					Type:     "AND",
					Field:    field,
					Operator: operator,
					Value:    args[2],
				})
			}
		}
	default:
		if len(args) > 1 {
			if sql, ok := args[0].(string); ok {
				q.havings = append(q.havings, QueryWhereClause{
					Type:     "AND",
					Raw:      sql,
					Bindings: args[1:],
				})
			}
		}
	}
	return q
}

// HavingRaw 原生HAVING条件
func (q *Query) HavingRaw(raw string, bindings ...interface{}) *Query {
	q.havings = append(q.havings, QueryWhereClause{
		Type:     "AND",
		Raw:      raw,
		Bindings: bindings,
	})
	return q
}

// GetOrders 获取排序子句
func (q *Query) GetOrders() []QueryOrderClause {
	return q.orders
}

// GetGroups 获取分组字段
func (q *Query) GetGroups() []string {
	return q.groups
}

// GetHavings 获取HAVING子句
func (q *Query) GetHavings() []QueryWhereClause {
	return q.havings
}

// ============================================================================
// 限制和分页方法
// ============================================================================

// Limit 限制返回数量
func (q *Query) Limit(n int) *Query {
	q.limit = n
	return q
}

// Offset 设置偏移量
func (q *Query) Offset(n int) *Query {
	q.offset = n
	return q
}

// Page 分页设置
func (q *Query) Page(page, perPage int) *Query {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	q.limit = perPage
	q.offset = (page - 1) * perPage
	return q
}

// GetLimit 获取限制数量
func (q *Query) GetLimit() int {
	return q.limit
}

// GetOffset 获取偏移量
func (q *Query) GetOffset() int {
	return q.offset
}

// ============================================================================
// SQL构建方法
// ============================================================================

// ToSQL 返回生成的SQL字符串和绑定参数
func (q *Query) ToSQL() (string, []interface{}) {
	return q.buildSelectSQL()
}

// buildSelectSQL 构建SELECT SQL
func (q *Query) buildSelectSQL() (string, []interface{}) {
	var sql strings.Builder
	var args []interface{}
	argIndex := 0

	// SELECT子句
	sql.WriteString("SELECT ")
	if q.distinct {
		sql.WriteString("DISTINCT ")
	}

	// 构建字段列表（包括普通字段和原生表达式）
	var selectParts []string

	// 添加普通字段
	selectParts = append(selectParts, q.fields...)

	// 添加原生SELECT表达式（带参数绑定）
	for _, raw := range q.selectRaws {
		processedExpr := q.processPlaceholders(raw.Expression, argIndex)
		selectParts = append(selectParts, processedExpr)
		if len(raw.Bindings) > 0 {
			args = append(args, raw.Bindings...)
			argIndex += len(raw.Bindings)
		}
	}

	if len(selectParts) > 0 {
		sql.WriteString(strings.Join(selectParts, ", "))
	} else {
		sql.WriteString("*")
	}

	// FROM子句
	sql.WriteString(" FROM ")
	sql.WriteString(q.table)

	// JOIN子句
	for _, join := range q.joins {
		if join.Type == "CROSS" {
			sql.WriteString(fmt.Sprintf(" CROSS JOIN %s", join.Table))
		} else {
			processedCondition := q.processPlaceholders(join.Condition, argIndex)
			sql.WriteString(fmt.Sprintf(" %s JOIN %s ON %s", join.Type, join.Table, processedCondition))
			if len(join.Bindings) > 0 {
				args = append(args, join.Bindings...)
				argIndex += len(join.Bindings)
			}
		}
	}

	// WHERE子句
	if len(q.wheres) > 0 {
		sql.WriteString(" WHERE ")
		for i, where := range q.wheres {
			if i > 0 {
				sql.WriteString(" " + where.Type + " ")
			}

			if where.Raw != "" {
				processedSQL := q.processPlaceholders(where.Raw, argIndex)
				sql.WriteString(processedSQL)
				if len(where.Bindings) > 0 {
					args = append(args, where.Bindings...)
					argIndex += len(where.Bindings)
				}
			} else {
				placeholder := q.buildPlaceholder(argIndex)
				sql.WriteString(fmt.Sprintf("%s %s %s", where.Field, where.Operator, placeholder))
				args = append(args, where.Value)
				argIndex++
			}
		}
	}

	// GROUP BY子句
	if len(q.groups) > 0 {
		sql.WriteString(" GROUP BY ")
		sql.WriteString(strings.Join(q.groups, ", "))
	}

	// HAVING子句
	if len(q.havings) > 0 {
		sql.WriteString(" HAVING ")
		for i, having := range q.havings {
			if i > 0 {
				sql.WriteString(" " + having.Type + " ")
			}

			if having.Raw != "" {
				processedSQL := q.processPlaceholders(having.Raw, argIndex)
				sql.WriteString(processedSQL)
				if len(having.Bindings) > 0 {
					args = append(args, having.Bindings...)
					argIndex += len(having.Bindings)
				}
			} else {
				placeholder := q.buildPlaceholder(argIndex)
				sql.WriteString(fmt.Sprintf("%s %s %s", having.Field, having.Operator, placeholder))
				args = append(args, having.Value)
				argIndex++
			}
		}
	}

	// ORDER BY子句
	if len(q.orders) > 0 {
		sql.WriteString(" ORDER BY ")
		orderParts := make([]string, len(q.orders))
		for i, order := range q.orders {
			if order.Raw != "" {
				orderParts[i] = order.Raw
			} else {
				orderParts[i] = fmt.Sprintf("%s %s", order.Field, order.Direction)
			}
		}
		sql.WriteString(strings.Join(orderParts, ", "))
	}

	// LIMIT和OFFSET子句
	driverName := q.getDriverName()
	if q.limit > 0 {
		switch driverName {
		case "sqlserver", "mssql":
			if q.offset > 0 {
				sql.WriteString(fmt.Sprintf(" OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", q.offset, q.limit))
			} else {
				sql.WriteString(fmt.Sprintf(" OFFSET 0 ROWS FETCH NEXT %d ROWS ONLY", q.limit))
			}
		default:
			sql.WriteString(fmt.Sprintf(" LIMIT %d", q.limit))
			if q.offset > 0 {
				sql.WriteString(fmt.Sprintf(" OFFSET %d", q.offset))
			}
		}
	} else if q.offset > 0 {
		switch driverName {
		case "sqlserver", "mssql":
			sql.WriteString(fmt.Sprintf(" OFFSET %d ROWS", q.offset))
		default:
			sql.WriteString(fmt.Sprintf(" OFFSET %d", q.offset))
		}
	}

	return sql.String(), args
}

// buildInsertSQL 构建INSERT SQL
func (q *Query) buildInsertSQL(data map[string]interface{}) (string, []interface{}) {
	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	args := make([]interface{}, 0, len(data))

	i := 0
	for column, value := range data {
		columns = append(columns, column)
		placeholders = append(placeholders, q.buildPlaceholder(i))
		args = append(args, value)
		i++
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		q.table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	return sql, args
}

// buildUpdateSQL 构建UPDATE SQL
func (q *Query) buildUpdateSQL(data map[string]interface{}) (string, []interface{}) {
	var sql strings.Builder
	var args []interface{}
	argIndex := 0

	sql.WriteString("UPDATE ")
	sql.WriteString(q.table)
	sql.WriteString(" SET ")

	setParts := make([]string, 0, len(data))
	for column, value := range data {
		placeholder := q.buildPlaceholder(argIndex)
		setParts = append(setParts, column+" = "+placeholder)
		args = append(args, value)
		argIndex++
	}
	sql.WriteString(strings.Join(setParts, ", "))

	// WHERE子句
	if len(q.wheres) > 0 {
		sql.WriteString(" WHERE ")
		for i, where := range q.wheres {
			if i > 0 {
				sql.WriteString(" " + where.Type + " ")
			}

			if where.Raw != "" {
				processedSQL := q.processPlaceholders(where.Raw, argIndex)
				sql.WriteString(processedSQL)
				if len(where.Bindings) > 0 {
					args = append(args, where.Bindings...)
					argIndex += len(where.Bindings)
				}
			} else {
				placeholder := q.buildPlaceholder(argIndex)
				sql.WriteString(fmt.Sprintf("%s %s %s", where.Field, where.Operator, placeholder))
				args = append(args, where.Value)
				argIndex++
			}
		}
	}

	return sql.String(), args
}

// buildDeleteSQL 构建DELETE SQL
func (q *Query) buildDeleteSQL() (string, []interface{}) {
	var sql strings.Builder
	var args []interface{}
	argIndex := 0

	sql.WriteString("DELETE FROM ")
	sql.WriteString(q.table)

	// WHERE子句
	if len(q.wheres) > 0 {
		sql.WriteString(" WHERE ")
		for i, where := range q.wheres {
			if i > 0 {
				sql.WriteString(" " + where.Type + " ")
			}

			if where.Raw != "" {
				processedSQL := q.processPlaceholders(where.Raw, argIndex)
				sql.WriteString(processedSQL)
				if len(where.Bindings) > 0 {
					args = append(args, where.Bindings...)
					argIndex += len(where.Bindings)
				}
			} else {
				placeholder := q.buildPlaceholder(argIndex)
				sql.WriteString(fmt.Sprintf("%s %s %s", where.Field, where.Operator, placeholder))
				args = append(args, where.Value)
				argIndex++
			}
		}
	}

	return sql.String(), args
}

// ============================================================================
// 查询执行方法
// ============================================================================

// Get 执行查询并返回多条记录
func (q *Query) Get() ([]map[string]interface{}, error) {
	// 检查是否启用缓存
	if q.cacheTTL > 0 {
		return q.getWithCache()
	}

	return q.executeGet()
}

// executeGet 执行实际的查询操作
func (q *Query) executeGet() ([]map[string]interface{}, error) {
	sqlStr, args := q.buildSelectSQL()

	rows, err := q.executeQuery(sqlStr, args)
	if err != nil {
		return nil, WrapError(err, ErrCodeQueryFailed, "query execution failed").
			WithContext("sql", sqlStr).
			WithContext("args", args)
	}
	defer rows.Close()

	return q.scanRows(rows)
}

// getWithCache 带缓存的查询
func (q *Query) getWithCache() ([]map[string]interface{}, error) {
	cache := DefaultManager().GetCache()
	if cache == nil {
		// 没有缓存实例，直接执行查询
		return q.executeGet()
	}

	// 生成缓存键
	cacheKey := q.generateCacheKey()

	// 尝试从缓存获取
	if cached, err := cache.Get(cacheKey); err == nil {
		if results, ok := cached.([]map[string]interface{}); ok {
			return results, nil
		}
	}

	// 缓存未命中，执行查询
	results, err := q.executeGet()
	if err != nil {
		return nil, err
	}

	// 存储到缓存
	if len(q.cacheTags) > 0 {
		// 检查是否支持标签
		if tagCache, ok := cache.(CacheWithTagsInterface); ok {
			tagCache.SetWithTags(cacheKey, results, q.cacheTTL, q.cacheTags)
		} else {
			cache.Set(cacheKey, results, q.cacheTTL)
		}
	} else {
		cache.Set(cacheKey, results, q.cacheTTL)
	}

	return results, nil
}

// generateCacheKey 生成缓存键
func (q *Query) generateCacheKey() string {
	// 如果用户指定了缓存键，直接使用
	if q.cacheKey != "" {
		return q.cacheKey
	}

	// 根据查询内容生成缓存键
	sqlStr, args := q.buildSelectSQL()
	return GenerateCacheKey("query", map[string]interface{}{
		"sql":  sqlStr,
		"args": args,
	})
}

// CacheKey 设置自定义缓存键
func (q *Query) CacheKey(key string) *Query {
	q.cacheKey = key
	return q
}

// First 获取第一条记录
func (q *Query) First() (map[string]interface{}, error) {
	// 检查是否启用缓存
	if q.cacheTTL > 0 {
		return q.firstWithCache()
	}

	return q.executeFirst()
}

// executeFirst 执行实际的First查询
func (q *Query) executeFirst() (map[string]interface{}, error) {
	q.Limit(1)
	results, err := q.executeGet()
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, ErrRecordNotFound.WithContext("table", q.table)
	}

	return results[0], nil
}

// firstWithCache 带缓存的First查询
func (q *Query) firstWithCache() (map[string]interface{}, error) {
	cache := DefaultManager().GetCache()
	if cache == nil {
		return q.executeFirst()
	}

	// 生成缓存键（添加first标识）
	q.Limit(1)
	cacheKey := q.generateCacheKey() + ":first"
	if q.cacheKey != "" {
		cacheKey = q.cacheKey
	}

	// 尝试从缓存获取
	if cached, err := cache.Get(cacheKey); err == nil {
		if result, ok := cached.(map[string]interface{}); ok {
			return result, nil
		}
	}

	// 缓存未命中，执行查询
	results, err := q.executeGet()
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, ErrRecordNotFound.WithContext("table", q.table)
	}

	result := results[0]

	// 存储到缓存
	if len(q.cacheTags) > 0 {
		if tagCache, ok := cache.(CacheWithTagsInterface); ok {
			tagCache.SetWithTags(cacheKey, result, q.cacheTTL, q.cacheTags)
		} else {
			cache.Set(cacheKey, result, q.cacheTTL)
		}
	} else {
		cache.Set(cacheKey, result, q.cacheTTL)
	}

	return result, nil
}

// Find 按主键查找
func (q *Query) Find(id interface{}) (map[string]interface{}, error) {
	q.Where("id", "=", id)
	return q.First()
}

// Exists 检查记录是否存在
func (q *Query) Exists() (bool, error) {
	count, err := q.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Count 计算记录数量
func (q *Query) Count() (int64, error) {
	// 保存原始配置
	originalFields := q.fields
	originalLimit := q.limit
	originalOffset := q.offset

	// 设置COUNT查询
	q.fields = []string{"COUNT(*) as count"}
	q.limit = 0
	q.offset = 0

	sqlStr, args := q.buildSelectSQL()

	// 恢复原始配置
	q.fields = originalFields
	q.limit = originalLimit
	q.offset = originalOffset

	rows, err := q.executeQuery(sqlStr, args)
	if err != nil {
		return 0, WrapError(err, ErrCodeQueryFailed, "count query failed").
			WithContext("sql", sqlStr).
			WithContext("args", args)
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, nil
	}

	var count int64
	if err := rows.Scan(&count); err != nil {
		return 0, WrapError(err, ErrCodeQueryFailed, "scan count result failed")
	}

	return count, nil
}

// Sum 求和
func (q *Query) Sum(field string) (float64, error) {
	return q.aggregate("SUM", field)
}

// Avg 平均值
func (q *Query) Avg(field string) (float64, error) {
	return q.aggregate("AVG", field)
}

// Max 最大值
func (q *Query) Max(field string) (interface{}, error) {
	originalFields := q.fields
	q.fields = []string{fmt.Sprintf("MAX(%s) as result", field)}

	sqlStr, args := q.buildSelectSQL()
	q.fields = originalFields

	rows, err := q.executeQuery(sqlStr, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	var result interface{}
	if err := rows.Scan(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// Min 最小值
func (q *Query) Min(field string) (interface{}, error) {
	originalFields := q.fields
	q.fields = []string{fmt.Sprintf("MIN(%s) as result", field)}

	sqlStr, args := q.buildSelectSQL()
	q.fields = originalFields

	rows, err := q.executeQuery(sqlStr, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	var result interface{}
	if err := rows.Scan(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// aggregate 聚合查询辅助方法
func (q *Query) aggregate(fn, field string) (float64, error) {
	originalFields := q.fields
	q.fields = []string{fmt.Sprintf("%s(%s) as result", fn, field)}

	sqlStr, args := q.buildSelectSQL()
	q.fields = originalFields

	rows, err := q.executeQuery(sqlStr, args)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, nil
	}

	// 使用 sql.NullFloat64 处理 NULL 值（当没有匹配记录时 SUM/AVG 返回 NULL）
	var result sql.NullFloat64
	if err := rows.Scan(&result); err != nil {
		return 0, err
	}

	if result.Valid {
		return result.Float64, nil
	}
	return 0, nil
}

// ============================================================================
// 数据操作方法
// ============================================================================

// Insert 插入数据
func (q *Query) Insert(data map[string]interface{}) (int64, error) {
	if len(data) == 0 {
		return 0, ErrInvalidParameter.WithDetails("insert data cannot be empty")
	}

	sqlStr, args := q.buildInsertSQL(data)
	driverName := q.getDriverName()

	if driverName == "postgres" {
		// PostgreSQL使用RETURNING获取ID
		sqlStr += " RETURNING id"
		rows, err := q.executeQuery(sqlStr, args)
		if err != nil {
			return 0, WrapError(err, ErrCodeQueryFailed, "insert failed").
				WithContext("sql", sqlStr).
				WithContext("args", args)
		}
		defer rows.Close()

		if rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				return 0, err
			}
			return id, nil
		}
		return 0, nil
	}

	// MySQL, SQLite使用LastInsertId
	result, err := q.executeExec(sqlStr, args)
	if err != nil {
		return 0, WrapError(err, ErrCodeQueryFailed, "insert failed").
			WithContext("sql", sqlStr).
			WithContext("args", args)
	}

	return result.LastInsertId()
}

// InsertGetId 插入数据并返回ID
func (q *Query) InsertGetId(data map[string]interface{}) (int64, error) {
	return q.Insert(data)
}

// InsertBatch 批量插入数据
func (q *Query) InsertBatch(data []map[string]interface{}) (int64, error) {
	if len(data) == 0 {
		return 0, nil
	}

	// 获取所有列名
	columnSet := make(map[string]bool)
	for _, row := range data {
		for column := range row {
			columnSet[column] = true
		}
	}

	columns := make([]string, 0, len(columnSet))
	for column := range columnSet {
		columns = append(columns, column)
	}

	// 构建SQL
	var sql strings.Builder
	sql.WriteString(fmt.Sprintf("INSERT INTO %s (%s) VALUES ",
		q.table, strings.Join(columns, ", ")))

	var args []interface{}
	valueParts := make([]string, len(data))
	argIndex := 0

	for i, row := range data {
		placeholders := make([]string, len(columns))
		for j, column := range columns {
			if value, exists := row[column]; exists {
				args = append(args, value)
			} else {
				args = append(args, nil)
			}
			placeholders[j] = q.buildPlaceholder(argIndex)
			argIndex++
		}
		valueParts[i] = fmt.Sprintf("(%s)", strings.Join(placeholders, ", "))
	}

	sql.WriteString(strings.Join(valueParts, ", "))

	result, err := q.executeExec(sql.String(), args)
	if err != nil {
		return 0, WrapError(err, ErrCodeQueryFailed, "batch insert failed")
	}

	return result.RowsAffected()
}

// Update 更新数据
func (q *Query) Update(data map[string]interface{}) (int64, error) {
	if len(data) == 0 {
		return 0, ErrInvalidParameter.WithDetails("update data cannot be empty")
	}

	sqlStr, args := q.buildUpdateSQL(data)

	result, err := q.executeExec(sqlStr, args)
	if err != nil {
		return 0, WrapError(err, ErrCodeQueryFailed, "update failed").
			WithContext("sql", sqlStr).
			WithContext("args", args)
	}

	return result.RowsAffected()
}

// Delete 删除数据
func (q *Query) Delete() (int64, error) {
	sqlStr, args := q.buildDeleteSQL()

	result, err := q.executeExec(sqlStr, args)
	if err != nil {
		return 0, WrapError(err, ErrCodeQueryFailed, "delete failed").
			WithContext("sql", sqlStr).
			WithContext("args", args)
	}

	return result.RowsAffected()
}

// Increment 字段自增
func (q *Query) Increment(field string, amount ...int) (int64, error) {
	inc := 1
	if len(amount) > 0 {
		inc = amount[0]
	}

	data := map[string]interface{}{
		field: fmt.Sprintf("%s + %d", field, inc),
	}

	// 使用原生SQL更新
	var sql strings.Builder
	var args []interface{}
	argIndex := 0

	sql.WriteString(fmt.Sprintf("UPDATE %s SET %s = %s + %d", q.table, field, field, inc))

	// WHERE子句
	if len(q.wheres) > 0 {
		sql.WriteString(" WHERE ")
		for i, where := range q.wheres {
			if i > 0 {
				sql.WriteString(" " + where.Type + " ")
			}

			if where.Raw != "" {
				processedSQL := q.processPlaceholders(where.Raw, argIndex)
				sql.WriteString(processedSQL)
				if len(where.Bindings) > 0 {
					args = append(args, where.Bindings...)
					argIndex += len(where.Bindings)
				}
			} else {
				placeholder := q.buildPlaceholder(argIndex)
				sql.WriteString(fmt.Sprintf("%s %s %s", where.Field, where.Operator, placeholder))
				args = append(args, where.Value)
				argIndex++
			}
		}
	}

	result, err := q.executeExec(sql.String(), args)
	if err != nil {
		return 0, WrapError(err, ErrCodeQueryFailed, "increment failed")
	}

	_ = data // 避免未使用警告
	return result.RowsAffected()
}

// Decrement 字段自减
func (q *Query) Decrement(field string, amount ...int) (int64, error) {
	dec := 1
	if len(amount) > 0 {
		dec = amount[0]
	}

	var sql strings.Builder
	var args []interface{}
	argIndex := 0

	sql.WriteString(fmt.Sprintf("UPDATE %s SET %s = %s - %d", q.table, field, field, dec))

	// WHERE子句
	if len(q.wheres) > 0 {
		sql.WriteString(" WHERE ")
		for i, where := range q.wheres {
			if i > 0 {
				sql.WriteString(" " + where.Type + " ")
			}

			if where.Raw != "" {
				processedSQL := q.processPlaceholders(where.Raw, argIndex)
				sql.WriteString(processedSQL)
				if len(where.Bindings) > 0 {
					args = append(args, where.Bindings...)
					argIndex += len(where.Bindings)
				}
			} else {
				placeholder := q.buildPlaceholder(argIndex)
				sql.WriteString(fmt.Sprintf("%s %s %s", where.Field, where.Operator, placeholder))
				args = append(args, where.Value)
				argIndex++
			}
		}
	}

	result, err := q.executeExec(sql.String(), args)
	if err != nil {
		return 0, WrapError(err, ErrCodeQueryFailed, "decrement failed")
	}

	return result.RowsAffected()
}

// Pluck 获取单列值
func (q *Query) Pluck(field string) ([]interface{}, error) {
	q.fields = []string{field}
	results, err := q.Get()
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(results))
	for i, row := range results {
		values[i] = row[field]
	}

	return values, nil
}

// Value 获取单个值
func (q *Query) Value(field string) (interface{}, error) {
	q.fields = []string{field}
	q.Limit(1)

	results, err := q.Get()
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	return results[0][field], nil
}

// GetWheres 获取WHERE子句（用于测试）
func (q *Query) GetWheres() []QueryWhereClause {
	return q.wheres
}

// ============================================================================
// 读写分离相关方法
// ============================================================================

// UseMaster 强制使用主库执行查询
// 用于需要读取最新数据的场景（如刚写入后立即读取）
// Requirements: 1.9
func (q *Query) UseMaster() *Query {
	if q.connection == nil {
		return q
	}

	// 检查是否是读写分离连接
	if rwConn, ok := q.connection.(*ReadWriteSplitConnection); ok {
		// 获取主库连接
		masterConn := rwConn.GetMasterConnection()
		if masterConn != nil {
			q.connection = masterConn
			if masterConn.GetBuilder() != nil {
				q.builder = masterConn.GetBuilder()
			}
		}
	}

	return q
}

// ForceRead 强制使用从库执行查询（默认行为，用于明确意图）
// Requirements: 1.8
func (q *Query) ForceRead() *Query {
	// 默认行为就是使用从库，这个方法主要用于代码可读性
	return q
}
