package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// QueryBuilder 查询构造器
type QueryBuilder struct {
	connection ConnectionInterface
	table      string
	fields     []string
	wheres     []WhereClause
	joins      []JoinClause
	orders     []OrderClause
	groups     []string
	havings    []WhereClause
	limitNum   int
	offsetNum  int
	distinct   bool
	// 移除bindings字段，改为在构建SQL时动态收集
	ctx context.Context // 添加context字段
}

// WhereClause WHERE子句
type WhereClause struct {
	Type        string // and, or
	Field       string
	Operator    string
	Value       interface{}
	Raw         string
	RawBindings []interface{} // 用于存储Raw语句的绑定参数
}

// JoinClause JOIN子句
type JoinClause struct {
	Type   string // inner, left, right
	Table  string
	First  string
	Op     string
	Second string
}

// OrderClause ORDER BY子句
type OrderClause struct {
	Field       string
	Direction   string
	Raw         string
	RawBindings []interface{} // 用于存储Raw语句的绑定参数
}

// NewQuery 创建查询构造器
func NewQuery(conn ConnectionInterface) QueryInterface {
	return &QueryBuilder{
		connection: conn,
		fields:     []string{"*"},
		wheres:     make([]WhereClause, 0),
		joins:      make([]JoinClause, 0),
		orders:     make([]OrderClause, 0),
		groups:     make([]string, 0),
		havings:    make([]WhereClause, 0),
	}
}

// From 设置表名
func (q *QueryBuilder) From(table string) QueryInterface {
	newQuery := q.clone()
	newQuery.table = table
	return newQuery
}

// Select 选择字段
func (q *QueryBuilder) Select(fields ...string) QueryInterface {
	newQuery := q.clone()
	newQuery.fields = fields
	return newQuery
}

// SelectRaw 原生字段选择
func (q *QueryBuilder) SelectRaw(raw string, bindings ...interface{}) QueryInterface {
	newQuery := q.clone()
	newQuery.fields = []string{raw}
	// 对于SelectRaw，我们暂时不处理绑定参数，因为它们通常在字段名中不需要
	return newQuery
}

// Distinct 去重
func (q *QueryBuilder) Distinct() QueryInterface {
	newQuery := q.clone()
	newQuery.distinct = true
	return newQuery
}

// Where 添加WHERE条件
func (q *QueryBuilder) Where(field string, operator string, value interface{}) QueryInterface {
	newQuery := q.clone()
	newQuery.wheres = append(newQuery.wheres, WhereClause{
		Type:     "and",
		Field:    field,
		Operator: operator,
		Value:    value,
	})
	return newQuery
}

// WhereIn 添加WHERE IN条件
func (q *QueryBuilder) WhereIn(field string, values []interface{}) QueryInterface {
	newQuery := q.clone()
	newQuery.wheres = append(newQuery.wheres, WhereClause{
		Type:     "and",
		Field:    field,
		Operator: "IN",
		Value:    values,
	})
	return newQuery
}

// WhereNotIn 添加WHERE NOT IN条件
func (q *QueryBuilder) WhereNotIn(field string, values []interface{}) QueryInterface {
	newQuery := q.clone()
	newQuery.wheres = append(newQuery.wheres, WhereClause{
		Type:     "and",
		Field:    field,
		Operator: "NOT IN",
		Value:    values,
	})
	return newQuery
}

// WhereBetween 添加WHERE BETWEEN条件
func (q *QueryBuilder) WhereBetween(field string, start, end interface{}) QueryInterface {
	newQuery := q.clone()
	newQuery.wheres = append(newQuery.wheres, WhereClause{
		Type:     "and",
		Field:    field,
		Operator: "BETWEEN",
		Value:    []interface{}{start, end},
	})
	return newQuery
}

// WhereNull 添加WHERE IS NULL条件
func (q *QueryBuilder) WhereNull(field string) QueryInterface {
	newQuery := q.clone()
	newQuery.wheres = append(newQuery.wheres, WhereClause{
		Type:     "and",
		Field:    field,
		Operator: "IS NULL",
	})
	return newQuery
}

// WhereNotNull 添加WHERE IS NOT NULL条件
func (q *QueryBuilder) WhereNotNull(field string) QueryInterface {
	newQuery := q.clone()
	newQuery.wheres = append(newQuery.wheres, WhereClause{
		Type:     "and",
		Field:    field,
		Operator: "IS NOT NULL",
	})
	return newQuery
}

// WhereRaw 添加原生WHERE条件
func (q *QueryBuilder) WhereRaw(raw string, bindings ...interface{}) QueryInterface {
	newQuery := q.clone()
	newQuery.wheres = append(newQuery.wheres, WhereClause{
		Type:        "and",
		Raw:         raw,
		RawBindings: bindings,
	})
	return newQuery
}

// OrWhere 添加OR WHERE条件
func (q *QueryBuilder) OrWhere(field string, operator string, value interface{}) QueryInterface {
	newQuery := q.clone()
	newQuery.wheres = append(newQuery.wheres, WhereClause{
		Type:     "or",
		Field:    field,
		Operator: operator,
		Value:    value,
	})
	return newQuery
}

// Join 添加JOIN
func (q *QueryBuilder) Join(table string, first string, operator string, second string) QueryInterface {
	newQuery := q.clone()
	newQuery.joins = append(newQuery.joins, JoinClause{
		Type:   "INNER",
		Table:  table,
		First:  first,
		Op:     operator,
		Second: second,
	})
	return newQuery
}

// LeftJoin 添加LEFT JOIN
func (q *QueryBuilder) LeftJoin(table string, first string, operator string, second string) QueryInterface {
	newQuery := q.clone()
	newQuery.joins = append(newQuery.joins, JoinClause{
		Type:   "LEFT",
		Table:  table,
		First:  first,
		Op:     operator,
		Second: second,
	})
	return newQuery
}

// RightJoin 添加RIGHT JOIN
func (q *QueryBuilder) RightJoin(table string, first string, operator string, second string) QueryInterface {
	newQuery := q.clone()
	newQuery.joins = append(newQuery.joins, JoinClause{
		Type:   "RIGHT",
		Table:  table,
		First:  first,
		Op:     operator,
		Second: second,
	})
	return newQuery
}

// InnerJoin 添加INNER JOIN
func (q *QueryBuilder) InnerJoin(table string, first string, operator string, second string) QueryInterface {
	return q.Join(table, first, operator, second)
}

// GroupBy 添加GROUP BY
func (q *QueryBuilder) GroupBy(fields ...string) QueryInterface {
	newQuery := q.clone()
	newQuery.groups = append(newQuery.groups, fields...)
	return newQuery
}

// Having 添加HAVING条件
func (q *QueryBuilder) Having(field string, operator string, value interface{}) QueryInterface {
	newQuery := q.clone()
	newQuery.havings = append(newQuery.havings, WhereClause{
		Type:     "and",
		Field:    field,
		Operator: operator,
		Value:    value,
	})
	return newQuery
}

// OrderBy 添加ORDER BY
func (q *QueryBuilder) OrderBy(field string, direction string) QueryInterface {
	newQuery := q.clone()
	if direction == "" {
		direction = "ASC"
	}
	newQuery.orders = append(newQuery.orders, OrderClause{
		Field:     field,
		Direction: strings.ToUpper(direction),
	})
	return newQuery
}

// OrderByRaw 添加原生ORDER BY
func (q *QueryBuilder) OrderByRaw(raw string, bindings ...interface{}) QueryInterface {
	newQuery := q.clone()
	newQuery.orders = append(newQuery.orders, OrderClause{
		Raw:         raw,
		RawBindings: bindings,
	})
	return newQuery
}

// Limit 设置LIMIT
func (q *QueryBuilder) Limit(limit int) QueryInterface {
	newQuery := q.clone()
	newQuery.limitNum = limit
	return newQuery
}

// Offset 设置OFFSET
func (q *QueryBuilder) Offset(offset int) QueryInterface {
	newQuery := q.clone()
	newQuery.offsetNum = offset
	return newQuery
}

// Page 设置分页
func (q *QueryBuilder) Page(page, pageSize int) QueryInterface {
	offset := (page - 1) * pageSize
	return q.Limit(pageSize).Offset(offset)
}

// WithContext 设置查询的上下文
func (q *QueryBuilder) WithContext(ctx context.Context) QueryInterface {
	clone := q.Clone().(*QueryBuilder)
	clone.ctx = ctx
	return clone
}

// WithTimeout 设置查询超时
func (q *QueryBuilder) WithTimeout(timeout time.Duration) QueryInterface {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	return q.WithContext(ctx)
}

// getContext 获取查询上下文，如果没有设置则使用默认值
func (q *QueryBuilder) getContext() context.Context {
	if q.ctx != nil {
		return q.ctx
	}
	return context.Background()
}

// Get 执行查询并返回所有记录
func (q *QueryBuilder) Get() ([]map[string]interface{}, error) {
	sql, bindings, err := q.ToSQL()
	if err != nil {
		return nil, err
	}

	// 对于支持context的连接，使用内部方法
	ctx := q.getContext()
	rows, err := q.queryWithContext(ctx, sql, bindings...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	return results, rows.Err()
}

// First 执行查询并返回第一条记录
func (q *QueryBuilder) First() (map[string]interface{}, error) {
	results, err := q.Limit(1).Get()
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no records found")
	}
	return results[0], nil
}

// Find 根据ID查找记录
func (q *QueryBuilder) Find(id interface{}) (map[string]interface{}, error) {
	return q.Where("id", "=", id).First()
}

// Count 获取记录数量
func (q *QueryBuilder) Count() (int64, error) {
	sql, bindings, err := q.buildCountSQL()
	if err != nil {
		return 0, err
	}

	ctx := q.getContext()
	row := q.queryRowWithContext(ctx, sql, bindings...)
	var count int64
	err = row.Scan(&count)
	return count, err
}

// Exists 检查是否存在记录
func (q *QueryBuilder) Exists() (bool, error) {
	count, err := q.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Insert 插入记录
func (q *QueryBuilder) Insert(data map[string]interface{}) (int64, error) {
	sql, bindings, err := q.buildInsertSQL(data)
	if err != nil {
		return 0, err
	}

	ctx := q.getContext()
	result, err := q.execWithContext(ctx, sql, bindings...)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// InsertBatch 批量插入记录
func (q *QueryBuilder) InsertBatch(data []map[string]interface{}) (int64, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("no data to insert")
	}

	sql, bindings, err := q.buildInsertBatchSQL(data)
	if err != nil {
		return 0, err
	}

	ctx := q.getContext()
	result, err := q.execWithContext(ctx, sql, bindings...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Update 更新记录
func (q *QueryBuilder) Update(data map[string]interface{}) (int64, error) {
	sql, bindings, err := q.buildUpdateSQL(data)
	if err != nil {
		return 0, err
	}

	ctx := q.getContext()
	result, err := q.execWithContext(ctx, sql, bindings...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Delete 删除记录
func (q *QueryBuilder) Delete() (int64, error) {
	sql, bindings, err := q.buildDeleteSQL()
	if err != nil {
		return 0, err
	}

	ctx := q.getContext()
	result, err := q.execWithContext(ctx, sql, bindings...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// ToSQL 构建SQL语句
func (q *QueryBuilder) ToSQL() (string, []interface{}, error) {
	return q.buildSelectSQL()
}

// Clone 克隆查询构造器
func (q *QueryBuilder) Clone() QueryInterface {
	return q.clone()
}

// clone 内部克隆方法
func (q *QueryBuilder) clone() *QueryBuilder {
	newQuery := &QueryBuilder{
		connection: q.connection,
		table:      q.table,
		fields:     make([]string, len(q.fields)),
		wheres:     make([]WhereClause, len(q.wheres)),
		joins:      make([]JoinClause, len(q.joins)),
		orders:     make([]OrderClause, len(q.orders)),
		groups:     make([]string, len(q.groups)),
		havings:    make([]WhereClause, len(q.havings)),
		limitNum:   q.limitNum,
		offsetNum:  q.offsetNum,
		distinct:   q.distinct,
		ctx:        q.ctx, // 克隆context
	}

	copy(newQuery.fields, q.fields)
	copy(newQuery.wheres, q.wheres)
	copy(newQuery.joins, q.joins)
	copy(newQuery.orders, q.orders)
	copy(newQuery.groups, q.groups)
	copy(newQuery.havings, q.havings)

	// 如果fields为空，设置默认值
	if len(newQuery.fields) == 0 {
		newQuery.fields = []string{"*"}
	}

	return newQuery
}

// buildSelectSQL 构建SELECT语句
func (q *QueryBuilder) buildSelectSQL() (string, []interface{}, error) {
	if q.table == "" {
		return "", nil, fmt.Errorf("table name is required")
	}

	var sql strings.Builder
	var bindings []interface{}

	// SELECT
	sql.WriteString("SELECT ")
	if q.distinct {
		sql.WriteString("DISTINCT ")
	}
	sql.WriteString(strings.Join(q.fields, ", "))

	// FROM
	sql.WriteString(" FROM ")
	sql.WriteString(q.table)

	// JOIN
	for _, join := range q.joins {
		sql.WriteString(" ")
		sql.WriteString(join.Type)
		sql.WriteString(" JOIN ")
		sql.WriteString(join.Table)
		sql.WriteString(" ON ")
		sql.WriteString(join.First)
		sql.WriteString(" ")
		sql.WriteString(join.Op)
		sql.WriteString(" ")
		sql.WriteString(join.Second)
	}

	// WHERE
	if len(q.wheres) > 0 {
		sql.WriteString(" WHERE ")
		for i, where := range q.wheres {
			if i > 0 {
				sql.WriteString(" ")
				sql.WriteString(strings.ToUpper(where.Type))
				sql.WriteString(" ")
			}

			if where.Raw != "" {
				sql.WriteString(where.Raw)
				bindings = append(bindings, where.RawBindings...)
			} else {
				sql.WriteString(where.Field)
				sql.WriteString(" ")
				sql.WriteString(where.Operator)

				switch where.Operator {
				case "IN", "NOT IN":
					values := where.Value.([]interface{})
					placeholders := make([]string, len(values))
					for j := range values {
						placeholders[j] = "?"
					}
					sql.WriteString(" (")
					sql.WriteString(strings.Join(placeholders, ", "))
					sql.WriteString(")")
					bindings = append(bindings, values...)
				case "BETWEEN":
					values := where.Value.([]interface{})
					sql.WriteString(" ? AND ?")
					bindings = append(bindings, values...)
				case "IS NULL", "IS NOT NULL":
					// 不需要添加占位符
				default:
					sql.WriteString(" ?")
					bindings = append(bindings, where.Value)
				}
			}
		}
	}

	// GROUP BY
	if len(q.groups) > 0 {
		sql.WriteString(" GROUP BY ")
		sql.WriteString(strings.Join(q.groups, ", "))
	}

	// HAVING
	if len(q.havings) > 0 {
		sql.WriteString(" HAVING ")
		for i, having := range q.havings {
			if i > 0 {
				sql.WriteString(" ")
				sql.WriteString(strings.ToUpper(having.Type))
				sql.WriteString(" ")
			}
			sql.WriteString(having.Field)
			sql.WriteString(" ")
			sql.WriteString(having.Operator)
			sql.WriteString(" ?")
			bindings = append(bindings, having.Value)
		}
	}

	// ORDER BY
	if len(q.orders) > 0 {
		sql.WriteString(" ORDER BY ")
		orderParts := make([]string, len(q.orders))
		for i, order := range q.orders {
			if order.Raw != "" {
				orderParts[i] = order.Raw
				bindings = append(bindings, order.RawBindings...)
			} else {
				orderParts[i] = order.Field + " " + order.Direction
			}
		}
		sql.WriteString(strings.Join(orderParts, ", "))
	}

	// LIMIT
	if q.limitNum > 0 {
		sql.WriteString(" LIMIT ?")
		bindings = append(bindings, q.limitNum)
	}

	// OFFSET
	if q.offsetNum > 0 {
		sql.WriteString(" OFFSET ?")
		bindings = append(bindings, q.offsetNum)
	}

	return sql.String(), bindings, nil
}

// buildCountSQL 构建COUNT语句
func (q *QueryBuilder) buildCountSQL() (string, []interface{}, error) {
	countQuery := q.clone()
	countQuery.fields = []string{"COUNT(*) as count"}
	countQuery.orders = nil
	countQuery.limitNum = 0
	countQuery.offsetNum = 0
	return countQuery.buildSelectSQL()
}

// buildInsertSQL 构建INSERT语句
func (q *QueryBuilder) buildInsertSQL(data map[string]interface{}) (string, []interface{}, error) {
	if q.table == "" {
		return "", nil, fmt.Errorf("table name is required")
	}
	if len(data) == 0 {
		return "", nil, fmt.Errorf("insert data is required")
	}

	fields := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	bindings := make([]interface{}, 0, len(data))

	for field, value := range data {
		fields = append(fields, field)
		placeholders = append(placeholders, "?")
		bindings = append(bindings, value)
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		q.table,
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "))

	return sql, bindings, nil
}

// buildInsertBatchSQL 构建批量INSERT语句
func (q *QueryBuilder) buildInsertBatchSQL(data []map[string]interface{}) (string, []interface{}, error) {
	if q.table == "" {
		return "", nil, fmt.Errorf("table name is required")
	}
	if len(data) == 0 {
		return "", nil, fmt.Errorf("insert data is required")
	}

	// 获取字段列表（使用第一条记录的字段）
	var fields []string
	for field := range data[0] {
		fields = append(fields, field)
	}

	// 构建值占位符
	valueParts := make([]string, len(data))
	bindings := make([]interface{}, 0, len(data)*len(fields))

	for i, row := range data {
		placeholders := make([]string, len(fields))
		for j, field := range fields {
			placeholders[j] = "?"
			bindings = append(bindings, row[field])
		}
		valueParts[i] = "(" + strings.Join(placeholders, ", ") + ")"
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		q.table,
		strings.Join(fields, ", "),
		strings.Join(valueParts, ", "))

	return sql, bindings, nil
}

// buildUpdateSQL 构建UPDATE语句
func (q *QueryBuilder) buildUpdateSQL(data map[string]interface{}) (string, []interface{}, error) {
	if q.table == "" {
		return "", nil, fmt.Errorf("table name is required")
	}
	if len(data) == 0 {
		return "", nil, fmt.Errorf("update data is required")
	}

	var sql strings.Builder
	var bindings []interface{}

	sql.WriteString("UPDATE ")
	sql.WriteString(q.table)
	sql.WriteString(" SET ")

	setParts := make([]string, 0, len(data))
	for field, value := range data {
		setParts = append(setParts, field+" = ?")
		bindings = append(bindings, value)
	}
	sql.WriteString(strings.Join(setParts, ", "))

	// WHERE
	if len(q.wheres) > 0 {
		sql.WriteString(" WHERE ")
		for i, where := range q.wheres {
			if i > 0 {
				sql.WriteString(" ")
				sql.WriteString(strings.ToUpper(where.Type))
				sql.WriteString(" ")
			}
			sql.WriteString(where.Field)
			sql.WriteString(" ")
			sql.WriteString(where.Operator)
			sql.WriteString(" ?")
			bindings = append(bindings, where.Value)
		}
	}

	return sql.String(), bindings, nil
}

// buildDeleteSQL 构建DELETE语句
func (q *QueryBuilder) buildDeleteSQL() (string, []interface{}, error) {
	if q.table == "" {
		return "", nil, fmt.Errorf("table name is required")
	}

	var sql strings.Builder
	var bindings []interface{}

	sql.WriteString("DELETE FROM ")
	sql.WriteString(q.table)

	// WHERE
	if len(q.wheres) > 0 {
		sql.WriteString(" WHERE ")
		for i, where := range q.wheres {
			if i > 0 {
				sql.WriteString(" ")
				sql.WriteString(strings.ToUpper(where.Type))
				sql.WriteString(" ")
			}
			sql.WriteString(where.Field)
			sql.WriteString(" ")
			sql.WriteString(where.Operator)
			sql.WriteString(" ?")
			bindings = append(bindings, where.Value)
		}
	}

	return sql.String(), bindings, nil
}

// scanRows 扫描行数据
func (q *QueryBuilder) scanRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	columnTypes, err := rows.ColumnTypes()
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
		for i, col := range columns {
			val := values[i]

			// 处理NULL值
			if val == nil {
				row[col] = nil
				continue
			}

			// 根据数据库类型转换值
			switch columnTypes[i].DatabaseTypeName() {
			case "VARCHAR", "TEXT", "CHAR":
				if b, ok := val.([]byte); ok {
					row[col] = string(b)
				} else {
					row[col] = val
				}
			case "INT", "INTEGER", "BIGINT":
				row[col] = val
			case "FLOAT", "DOUBLE", "DECIMAL":
				row[col] = val
			case "BOOLEAN", "BOOL":
				row[col] = val
			case "TIMESTAMP", "DATETIME", "DATE", "TIME":
				row[col] = val
			default:
				row[col] = val
			}
		}

		results = append(results, row)
	}

	return results, rows.Err()
}

// queryWithContext 内部方法：执行带context的查询
func (q *QueryBuilder) queryWithContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	// 尝试使用带context的方法，如果不支持则回退到普通方法
	if ctxConn, ok := q.connection.(interface {
		QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	}); ok {
		return ctxConn.QueryContext(ctx, query, args...)
	}
	return q.connection.Query(query, args...)
}

// queryRowWithContext 内部方法：执行带context的单行查询
func (q *QueryBuilder) queryRowWithContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	// 尝试使用带context的方法，如果不支持则回退到普通方法
	if ctxConn, ok := q.connection.(interface {
		QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	}); ok {
		return ctxConn.QueryRowContext(ctx, query, args...)
	}
	return q.connection.QueryRow(query, args...)
}

// execWithContext 内部方法：执行带context的SQL语句
func (q *QueryBuilder) execWithContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	// 尝试使用带context的方法，如果不支持则回退到普通方法
	if ctxConn, ok := q.connection.(interface {
		ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	}); ok {
		return ctxConn.ExecContext(ctx, query, args...)
	}
	return q.connection.Exec(query, args...)
}
