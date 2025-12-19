package db

import (
	"fmt"
	"sort"
	"strings"
)

// QueryData 查询数据结构，用于Builder接口
// 这是一个通用的查询数据结构，用于在Builder和QueryBuilder之间传递数据
type QueryData struct {
	Table            string
	SelectColumns    []string
	Distinct         bool
	WhereConditions  []WhereClause
	JoinClauses      []JoinData
	GroupByColumns   []string
	HavingConditions []WhereClause
	OrderByColumns   []OrderData
	LimitCount       int
	OffsetCount      int
}

// WhereClause WHERE子句数据
type WhereClause struct {
	Column   string
	Operator string
	Value    interface{}
	Logic    string        // AND, OR
	Raw      string        // 原生SQL
	Values   []interface{} // 原生SQL的参数
	Nested   []WhereClause // 嵌套条件
}

// JoinData JOIN子句数据
type JoinData struct {
	Type      string // INNER, LEFT, RIGHT, CROSS
	Table     string
	Condition string
	Raw       string        // 原生SQL条件
	Values    []interface{} // 绑定参数
}

// OrderData 排序子句数据
type OrderData struct {
	Column    string
	Direction string // ASC, DESC
	Raw       string // 原生排序表达式
}

// BaseBuilder 基础SQL构建器
// 提供所有数据库构建器的通用功能
type BaseBuilder struct {
	placeholderStyle PlaceholderStyle
	identifierQuote  string
}

// NewBaseBuilder 创建基础构建器
func NewBaseBuilder(style PlaceholderStyle, quote string) *BaseBuilder {
	return &BaseBuilder{
		placeholderStyle: style,
		identifierQuote:  quote,
	}
}

// GetPlaceholderStyle 获取占位符风格
func (b *BaseBuilder) GetPlaceholderStyle() PlaceholderStyle {
	return b.placeholderStyle
}

// QuoteIdentifier 引用标识符
func (b *BaseBuilder) QuoteIdentifier(name string) string {
	if name == "*" {
		return name
	}
	// 处理带表前缀的字段名 (table.column)
	if strings.Contains(name, ".") {
		parts := strings.Split(name, ".")
		quoted := make([]string, len(parts))
		for i, part := range parts {
			if part == "*" {
				quoted[i] = part
			} else {
				quoted[i] = b.identifierQuote + part + b.identifierQuote
			}
		}
		return strings.Join(quoted, ".")
	}
	return b.identifierQuote + name + b.identifierQuote
}

// QuotePlaceholder 返回占位符
func (b *BaseBuilder) QuotePlaceholder(index int) string {
	switch b.placeholderStyle {
	case PlaceholderDollar:
		return fmt.Sprintf("$%d", index+1)
	default:
		return "?"
	}
}

// BuildPlaceholders 构建多个占位符
func (b *BaseBuilder) BuildPlaceholders(count int, startIndex int) string {
	if count <= 0 {
		return ""
	}
	placeholders := make([]string, count)
	for i := 0; i < count; i++ {
		placeholders[i] = b.QuotePlaceholder(startIndex + i)
	}
	return strings.Join(placeholders, ", ")
}

// ProcessPlaceholders 处理SQL中的占位符，将?转换为对应数据库的占位符格式
func (b *BaseBuilder) ProcessPlaceholders(sql string, startIndex int) string {
	if b.placeholderStyle == PlaceholderQuestion {
		return sql
	}

	// 对于PostgreSQL，将?替换为$N
	result := sql
	index := startIndex
	for strings.Contains(result, "?") {
		result = strings.Replace(result, "?", b.QuotePlaceholder(index), 1)
		index++
	}
	return result
}

// BuildSelectClause 构建SELECT子句
func (b *BaseBuilder) BuildSelectClause(columns []string, distinct bool) string {
	var sql strings.Builder
	sql.WriteString("SELECT ")
	if distinct {
		sql.WriteString("DISTINCT ")
	}
	if len(columns) == 0 {
		sql.WriteString("*")
	} else {
		sql.WriteString(strings.Join(columns, ", "))
	}
	return sql.String()
}

// BuildFromClause 构建FROM子句
func (b *BaseBuilder) BuildFromClause(table string) string {
	return " FROM " + table
}

// BuildJoinClause 构建JOIN子句
func (b *BaseBuilder) BuildJoinClause(joins []JoinData, argIndex *int, args *[]interface{}) string {
	if len(joins) == 0 {
		return ""
	}

	var sql strings.Builder
	for _, join := range joins {
		joinType := strings.ToUpper(join.Type)
		if joinType == "" {
			joinType = "INNER"
		}

		if joinType == "CROSS" {
			sql.WriteString(fmt.Sprintf(" CROSS JOIN %s", join.Table))
		} else if join.Raw != "" {
			processedSQL := b.ProcessPlaceholders(join.Raw, *argIndex)
			sql.WriteString(fmt.Sprintf(" %s JOIN %s ON %s", joinType, join.Table, processedSQL))
			if len(join.Values) > 0 {
				*args = append(*args, join.Values...)
				*argIndex += len(join.Values)
			}
		} else if join.Condition != "" {
			sql.WriteString(fmt.Sprintf(" %s JOIN %s ON %s", joinType, join.Table, join.Condition))
		}
	}
	return sql.String()
}

// BuildWhereClause 构建WHERE子句
func (b *BaseBuilder) BuildWhereClause(conditions []WhereClause, argIndex *int, args *[]interface{}) string {
	if len(conditions) == 0 {
		return ""
	}

	var sql strings.Builder
	sql.WriteString(" WHERE ")

	for i, condition := range conditions {
		if i > 0 {
			sql.WriteString(" " + condition.Logic + " ")
		}

		if condition.Raw != "" {
			processedSQL := b.ProcessPlaceholders(condition.Raw, *argIndex)
			sql.WriteString(processedSQL)
			if len(condition.Values) > 0 {
				*args = append(*args, condition.Values...)
				*argIndex += len(condition.Values)
			}
		} else {
			placeholder := b.QuotePlaceholder(*argIndex)
			sql.WriteString(fmt.Sprintf("%s %s %s", condition.Column, condition.Operator, placeholder))
			*args = append(*args, condition.Value)
			*argIndex++
		}
	}
	return sql.String()
}

// BuildGroupByClause 构建GROUP BY子句
func (b *BaseBuilder) BuildGroupByClause(columns []string) string {
	if len(columns) == 0 {
		return ""
	}
	return " GROUP BY " + strings.Join(columns, ", ")
}

// BuildHavingClause 构建HAVING子句
func (b *BaseBuilder) BuildHavingClause(conditions []WhereClause, argIndex *int, args *[]interface{}) string {
	if len(conditions) == 0 {
		return ""
	}

	var sql strings.Builder
	sql.WriteString(" HAVING ")

	for i, condition := range conditions {
		if i > 0 {
			sql.WriteString(" " + condition.Logic + " ")
		}

		if condition.Raw != "" {
			processedSQL := b.ProcessPlaceholders(condition.Raw, *argIndex)
			sql.WriteString(processedSQL)
			if len(condition.Values) > 0 {
				*args = append(*args, condition.Values...)
				*argIndex += len(condition.Values)
			}
		} else {
			placeholder := b.QuotePlaceholder(*argIndex)
			sql.WriteString(fmt.Sprintf("%s %s %s", condition.Column, condition.Operator, placeholder))
			*args = append(*args, condition.Value)
			*argIndex++
		}
	}
	return sql.String()
}

// BuildOrderByClause 构建ORDER BY子句
func (b *BaseBuilder) BuildOrderByClause(orders []OrderData) string {
	if len(orders) == 0 {
		return ""
	}

	var parts []string
	for _, order := range orders {
		if order.Raw != "" {
			parts = append(parts, order.Raw)
		} else {
			direction := strings.ToUpper(order.Direction)
			if direction != "ASC" && direction != "DESC" {
				direction = "ASC"
			}
			parts = append(parts, fmt.Sprintf("%s %s", order.Column, direction))
		}
	}
	return " ORDER BY " + strings.Join(parts, ", ")
}

// BuildLimitOffsetClause 构建LIMIT和OFFSET子句（MySQL/SQLite风格）
func (b *BaseBuilder) BuildLimitOffsetClause(limit, offset int) string {
	var sql strings.Builder
	if limit > 0 {
		sql.WriteString(fmt.Sprintf(" LIMIT %d", limit))
	}
	if offset > 0 {
		sql.WriteString(fmt.Sprintf(" OFFSET %d", offset))
	}
	return sql.String()
}

// BuildInsertSQL 构建INSERT语句
func (b *BaseBuilder) BuildInsertSQL(table string, data map[string]interface{}) (string, []interface{}) {
	if len(data) == 0 {
		return "", nil
	}

	// 获取排序后的列名以保证顺序一致
	columns := make([]string, 0, len(data))
	for col := range data {
		columns = append(columns, col)
	}
	sort.Strings(columns)

	placeholders := make([]string, len(columns))
	args := make([]interface{}, len(columns))

	for i, col := range columns {
		placeholders[i] = b.QuotePlaceholder(i)
		args[i] = data[col]
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	return sql, args
}

// BuildInsertBatchSQL 构建批量INSERT语句
func (b *BaseBuilder) BuildInsertBatchSQL(table string, data []map[string]interface{}) (string, []interface{}) {
	if len(data) == 0 {
		return "", nil
	}

	// 从第一条数据获取列名
	columns := make([]string, 0, len(data[0]))
	for col := range data[0] {
		columns = append(columns, col)
	}
	sort.Strings(columns)

	var args []interface{}
	var valueGroups []string
	argIndex := 0

	for _, row := range data {
		placeholders := make([]string, len(columns))
		for i, col := range columns {
			placeholders[i] = b.QuotePlaceholder(argIndex)
			args = append(args, row[col])
			argIndex++
		}
		valueGroups = append(valueGroups, "("+strings.Join(placeholders, ", ")+")")
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		table,
		strings.Join(columns, ", "),
		strings.Join(valueGroups, ", "))

	return sql, args
}

// BuildUpdateSQL 构建UPDATE语句
func (b *BaseBuilder) BuildUpdateSQL(table string, data map[string]interface{}, conditions []WhereClause) (string, []interface{}) {
	if len(data) == 0 {
		return "", nil
	}

	// 获取排序后的列名
	columns := make([]string, 0, len(data))
	for col := range data {
		columns = append(columns, col)
	}
	sort.Strings(columns)

	var sql strings.Builder
	var args []interface{}
	argIndex := 0

	sql.WriteString("UPDATE ")
	sql.WriteString(table)
	sql.WriteString(" SET ")

	setParts := make([]string, len(columns))
	for i, col := range columns {
		setParts[i] = fmt.Sprintf("%s = %s", col, b.QuotePlaceholder(argIndex))
		args = append(args, data[col])
		argIndex++
	}
	sql.WriteString(strings.Join(setParts, ", "))

	// WHERE子句
	whereSQL := b.BuildWhereClause(conditions, &argIndex, &args)
	sql.WriteString(whereSQL)

	return sql.String(), args
}

// BuildDeleteSQL 构建DELETE语句
func (b *BaseBuilder) BuildDeleteSQL(table string, conditions []WhereClause) (string, []interface{}) {
	var sql strings.Builder
	var args []interface{}
	argIndex := 0

	sql.WriteString("DELETE FROM ")
	sql.WriteString(table)

	// WHERE子句
	whereSQL := b.BuildWhereClause(conditions, &argIndex, &args)
	sql.WriteString(whereSQL)

	return sql.String(), args
}

// BuildCountSQL 构建COUNT语句
func (b *BaseBuilder) BuildCountSQL(table string, conditions []WhereClause, joins []JoinData) (string, []interface{}) {
	var sql strings.Builder
	var args []interface{}
	argIndex := 0

	sql.WriteString("SELECT COUNT(*) as count FROM ")
	sql.WriteString(table)

	// JOIN子句
	joinSQL := b.BuildJoinClause(joins, &argIndex, &args)
	sql.WriteString(joinSQL)

	// WHERE子句
	whereSQL := b.BuildWhereClause(conditions, &argIndex, &args)
	sql.WriteString(whereSQL)

	return sql.String(), args
}

// BuildAggregateSQL 构建聚合语句
func (b *BaseBuilder) BuildAggregateSQL(table string, fn string, field string, conditions []WhereClause, joins []JoinData) (string, []interface{}) {
	var sql strings.Builder
	var args []interface{}
	argIndex := 0

	sql.WriteString(fmt.Sprintf("SELECT %s(%s) as result FROM ", strings.ToUpper(fn), field))
	sql.WriteString(table)

	// JOIN子句
	joinSQL := b.BuildJoinClause(joins, &argIndex, &args)
	sql.WriteString(joinSQL)

	// WHERE子句
	whereSQL := b.BuildWhereClause(conditions, &argIndex, &args)
	sql.WriteString(whereSQL)

	return sql.String(), args
}

// ExtractQueryData 从QueryBuilder提取查询数据
// 这个函数用于将QueryBuilder的内部状态转换为QueryData结构
func ExtractQueryData(q interface{}) *QueryData {
	qb, ok := q.(*QueryBuilder)
	if !ok {
		return nil
	}

	data := &QueryData{
		Table:          qb.tableName,
		SelectColumns:  qb.selectColumns,
		GroupByColumns: qb.groupByColumns,
		LimitCount:     qb.limitCount,
		OffsetCount:    qb.offsetCount,
	}

	// 转换WHERE条件
	for _, cond := range qb.whereConditions {
		data.WhereConditions = append(data.WhereConditions, WhereClause{
			Column:   cond.Column,
			Operator: cond.Operator,
			Value:    cond.Value,
			Logic:    cond.Logic,
			Raw:      cond.Raw,
			Values:   cond.Values,
		})
	}

	// 转换JOIN子句
	for _, join := range qb.joinClauses {
		data.JoinClauses = append(data.JoinClauses, JoinData{
			Type:      join.Type,
			Table:     join.Table,
			Condition: join.Condition,
			Raw:       join.Raw,
			Values:    join.Values,
		})
	}

	// 转换HAVING条件
	for _, cond := range qb.havingConditions {
		data.HavingConditions = append(data.HavingConditions, WhereClause{
			Column:   cond.Column,
			Operator: cond.Operator,
			Value:    cond.Value,
			Logic:    cond.Logic,
			Raw:      cond.Raw,
			Values:   cond.Values,
		})
	}

	// 转换ORDER BY子句
	for _, order := range qb.orderByColumns {
		data.OrderByColumns = append(data.OrderByColumns, OrderData{
			Column:    order.Column,
			Direction: order.Direction,
		})
	}

	return data
}
