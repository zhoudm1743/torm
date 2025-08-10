package db

import (
	"fmt"
	"strconv"
	"strings"
)

// PostgreSQLBuilder PostgreSQL查询构建器
type PostgreSQLBuilder struct{}

// NewPostgreSQLBuilder 创建PostgreSQL查询构建器
func NewPostgreSQLBuilder() BuilderInterface {
	return &PostgreSQLBuilder{}
}

// BuildSelect 构建SELECT语句
func (b *PostgreSQLBuilder) BuildSelect(query QueryInterface) (string, []interface{}, error) {
	// 使用通用的查询构建器，但需要处理PostgreSQL特定的语法
	qb, ok := query.(*QueryBuilder)
	if !ok {
		return "", nil, fmt.Errorf("invalid query builder type")
	}

	if qb.table == "" {
		return "", nil, fmt.Errorf("table name is required")
	}

	var sql strings.Builder
	var bindings []interface{}

	// SELECT
	sql.WriteString("SELECT ")
	if qb.distinct {
		sql.WriteString("DISTINCT ")
	}
	sql.WriteString(strings.Join(qb.fields, ", "))

	// FROM
	sql.WriteString(" FROM ")
	sql.WriteString(b.QuoteIdentifier(qb.table))

	// JOIN
	for _, join := range qb.joins {
		sql.WriteString(" ")
		sql.WriteString(join.Type)
		sql.WriteString(" JOIN ")
		sql.WriteString(b.QuoteIdentifier(join.Table))
		sql.WriteString(" ON ")
		sql.WriteString(b.QuoteIdentifier(join.First))
		sql.WriteString(" ")
		sql.WriteString(join.Op)
		sql.WriteString(" ")
		sql.WriteString(b.QuoteIdentifier(join.Second))
	}

	// WHERE
	if len(qb.wheres) > 0 {
		whereSQL, whereBindings := b.buildWhereClause(qb.wheres)
		sql.WriteString(" WHERE ")
		sql.WriteString(whereSQL)
		bindings = append(bindings, whereBindings...)
	}

	// GROUP BY
	if len(qb.groups) > 0 {
		sql.WriteString(" GROUP BY ")
		quotedGroups := make([]string, len(qb.groups))
		for i, group := range qb.groups {
			quotedGroups[i] = b.QuoteIdentifier(group)
		}
		sql.WriteString(strings.Join(quotedGroups, ", "))
	}

	// HAVING
	if len(qb.havings) > 0 {
		havingSQL, havingBindings := b.buildWhereClause(qb.havings)
		sql.WriteString(" HAVING ")
		sql.WriteString(havingSQL)
		bindings = append(bindings, havingBindings...)
	}

	// ORDER BY
	if len(qb.orders) > 0 {
		sql.WriteString(" ORDER BY ")
		orderParts := make([]string, len(qb.orders))
		for i, order := range qb.orders {
			if order.Raw != "" {
				orderParts[i] = order.Raw
				bindings = append(bindings, order.RawBindings...)
			} else {
				orderParts[i] = b.QuoteIdentifier(order.Field) + " " + order.Direction
			}
		}
		sql.WriteString(strings.Join(orderParts, ", "))
	}

	// LIMIT
	if qb.limitNum > 0 {
		sql.WriteString(" LIMIT ")
		sql.WriteString(strconv.Itoa(qb.limitNum))
	}

	// OFFSET
	if qb.offsetNum > 0 {
		sql.WriteString(" OFFSET ")
		sql.WriteString(strconv.Itoa(qb.offsetNum))
	}

	return sql.String(), bindings, nil
}

// BuildInsert 构建INSERT语句
func (b *PostgreSQLBuilder) BuildInsert(table string, data map[string]interface{}) (string, []interface{}, error) {
	if len(data) == 0 {
		return "", nil, fmt.Errorf("no data to insert")
	}

	var fields []string
	var placeholders []string
	var bindings []interface{}

	i := 1
	for field, value := range data {
		fields = append(fields, b.QuoteIdentifier(field))
		placeholders = append(placeholders, "$"+strconv.Itoa(i))
		bindings = append(bindings, value)
		i++
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id",
		b.QuoteIdentifier(table),
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "))

	return sql, bindings, nil
}

// BuildInsertBatch 构建批量INSERT语句
func (b *PostgreSQLBuilder) BuildInsertBatch(table string, data []map[string]interface{}) (string, []interface{}, error) {
	if len(data) == 0 {
		return "", nil, fmt.Errorf("no data to insert")
	}

	// 获取字段列表（从第一条记录）
	var fields []string
	for field := range data[0] {
		fields = append(fields, b.QuoteIdentifier(field))
	}

	var valueParts []string
	var bindings []interface{}
	placeholder := 1

	for _, row := range data {
		var rowPlaceholders []string
		for _, field := range fields {
			// 移除引号来获取原始字段名
			originalField := strings.Trim(field, `"`)
			value, exists := row[originalField]
			if !exists {
				value = nil
			}
			rowPlaceholders = append(rowPlaceholders, "$"+strconv.Itoa(placeholder))
			bindings = append(bindings, value)
			placeholder++
		}
		valueParts = append(valueParts, "("+strings.Join(rowPlaceholders, ", ")+")")
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		b.QuoteIdentifier(table),
		strings.Join(fields, ", "),
		strings.Join(valueParts, ", "))

	return sql, bindings, nil
}

// BuildUpdate 构建UPDATE语句
func (b *PostgreSQLBuilder) BuildUpdate(query QueryInterface, data map[string]interface{}) (string, []interface{}, error) {
	qb, ok := query.(*QueryBuilder)
	if !ok {
		return "", nil, fmt.Errorf("invalid query builder type")
	}

	if qb.table == "" {
		return "", nil, fmt.Errorf("table name is required")
	}

	if len(data) == 0 {
		return "", nil, fmt.Errorf("no data to update")
	}

	var setParts []string
	var bindings []interface{}
	placeholder := 1

	for field, value := range data {
		setParts = append(setParts, b.QuoteIdentifier(field)+" = $"+strconv.Itoa(placeholder))
		bindings = append(bindings, value)
		placeholder++
	}

	sql := fmt.Sprintf("UPDATE %s SET %s",
		b.QuoteIdentifier(qb.table),
		strings.Join(setParts, ", "))

	// WHERE
	if len(qb.wheres) > 0 {
		whereSQL, whereBindings := b.buildWhereClauseWithPlaceholder(qb.wheres, placeholder)
		sql += " WHERE " + whereSQL
		bindings = append(bindings, whereBindings...)
	}

	return sql, bindings, nil
}

// BuildDelete 构建DELETE语句
func (b *PostgreSQLBuilder) BuildDelete(query QueryInterface) (string, []interface{}, error) {
	qb, ok := query.(*QueryBuilder)
	if !ok {
		return "", nil, fmt.Errorf("invalid query builder type")
	}

	if qb.table == "" {
		return "", nil, fmt.Errorf("table name is required")
	}

	sql := "DELETE FROM " + b.QuoteIdentifier(qb.table)
	var bindings []interface{}

	// WHERE
	if len(qb.wheres) > 0 {
		whereSQL, whereBindings := b.buildWhereClause(qb.wheres)
		sql += " WHERE " + whereSQL
		bindings = append(bindings, whereBindings...)
	}

	return sql, bindings, nil
}

// BuildCount 构建COUNT语句
func (b *PostgreSQLBuilder) BuildCount(query QueryInterface) (string, []interface{}, error) {
	qb, ok := query.(*QueryBuilder)
	if !ok {
		return "", nil, fmt.Errorf("invalid query builder type")
	}

	if qb.table == "" {
		return "", nil, fmt.Errorf("table name is required")
	}

	sql := "SELECT COUNT(*) FROM " + b.QuoteIdentifier(qb.table)
	var bindings []interface{}

	// JOIN
	for _, join := range qb.joins {
		sql += " " + join.Type + " JOIN " + b.QuoteIdentifier(join.Table) +
			" ON " + b.QuoteIdentifier(join.First) + " " + join.Op + " " + b.QuoteIdentifier(join.Second)
	}

	// WHERE
	if len(qb.wheres) > 0 {
		whereSQL, whereBindings := b.buildWhereClause(qb.wheres)
		sql += " WHERE " + whereSQL
		bindings = append(bindings, whereBindings...)
	}

	return sql, bindings, nil
}

// QuoteIdentifier 引用标识符（PostgreSQL使用双引号）
func (b *PostgreSQLBuilder) QuoteIdentifier(identifier string) string {
	// 如果已经包含引号或者是表达式，直接返回
	if strings.Contains(identifier, `"`) || strings.Contains(identifier, ".") || strings.Contains(identifier, "(") {
		return identifier
	}
	return `"` + identifier + `"`
}

// QuoteValue 引用值
func (b *PostgreSQLBuilder) QuoteValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return "'" + strings.ReplaceAll(v, "'", "''") + "'"
	case nil:
		return "NULL"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// buildWhereClause 构建WHERE子句
func (b *PostgreSQLBuilder) buildWhereClause(wheres []WhereClause) (string, []interface{}) {
	return b.buildWhereClauseWithPlaceholder(wheres, 1)
}

// buildWhereClauseWithPlaceholder 构建WHERE子句（指定起始占位符）
func (b *PostgreSQLBuilder) buildWhereClauseWithPlaceholder(wheres []WhereClause, startPlaceholder int) (string, []interface{}) {
	if len(wheres) == 0 {
		return "", nil
	}

	var parts []string
	var bindings []interface{}
	placeholder := startPlaceholder

	for i, where := range wheres {
		if i > 0 {
			parts = append(parts, strings.ToUpper(where.Type))
		}

		if where.Raw != "" {
			parts = append(parts, where.Raw)
			bindings = append(bindings, where.RawBindings...)
		} else {
			switch where.Operator {
			case "IN", "NOT IN":
				values := where.Value.([]interface{})
				placeholders := make([]string, len(values))
				for j, value := range values {
					placeholders[j] = "$" + strconv.Itoa(placeholder)
					bindings = append(bindings, value)
					placeholder++
				}
				parts = append(parts, fmt.Sprintf("%s %s (%s)",
					b.QuoteIdentifier(where.Field),
					where.Operator,
					strings.Join(placeholders, ", ")))

			case "BETWEEN":
				values := where.Value.([]interface{})
				if len(values) >= 2 {
					parts = append(parts, fmt.Sprintf("%s BETWEEN $%d AND $%d",
						b.QuoteIdentifier(where.Field),
						placeholder, placeholder+1))
					bindings = append(bindings, values[0], values[1])
					placeholder += 2
				}

			case "IS NULL", "IS NOT NULL":
				parts = append(parts, fmt.Sprintf("%s %s",
					b.QuoteIdentifier(where.Field),
					where.Operator))

			default:
				parts = append(parts, fmt.Sprintf("%s %s $%d",
					b.QuoteIdentifier(where.Field),
					where.Operator,
					placeholder))
				bindings = append(bindings, where.Value)
				placeholder++
			}
		}
	}

	return strings.Join(parts, " "), bindings
}
