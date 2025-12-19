package db

import (
	"fmt"
	"strings"
)

// MySQLBuilderImpl MySQL SQL构建器完整实现
type MySQLBuilderImpl struct {
	*BaseBuilder
}

// NewMySQLBuilderImpl 创建MySQL构建器
func NewMySQLBuilderImpl() *MySQLBuilderImpl {
	return &MySQLBuilderImpl{
		BaseBuilder: NewBaseBuilder(PlaceholderQuestion, "`"),
	}
}

// BuildSelect 构建SELECT语句
func (b *MySQLBuilderImpl) BuildSelect(q interface{}) (string, []interface{}) {
	data := ExtractQueryData(q)
	if data == nil {
		return "", nil
	}

	var sql strings.Builder
	var args []interface{}
	argIndex := 0

	// SELECT子句
	sql.WriteString(b.BuildSelectClause(data.SelectColumns, data.Distinct))

	// FROM子句
	sql.WriteString(b.BuildFromClause(data.Table))

	// JOIN子句
	joinSQL := b.BuildJoinClause(data.JoinClauses, &argIndex, &args)
	sql.WriteString(joinSQL)

	// WHERE子句
	whereSQL := b.BuildWhereClause(data.WhereConditions, &argIndex, &args)
	sql.WriteString(whereSQL)

	// GROUP BY子句
	sql.WriteString(b.BuildGroupByClause(data.GroupByColumns))

	// HAVING子句
	havingSQL := b.BuildHavingClause(data.HavingConditions, &argIndex, &args)
	sql.WriteString(havingSQL)

	// ORDER BY子句
	sql.WriteString(b.BuildOrderByClause(data.OrderByColumns))

	// LIMIT和OFFSET子句
	sql.WriteString(b.BuildLimitOffsetClause(data.LimitCount, data.OffsetCount))

	return sql.String(), args
}

// BuildInsert 构建INSERT语句
func (b *MySQLBuilderImpl) BuildInsert(table string, data map[string]interface{}) (string, []interface{}) {
	return b.BaseBuilder.BuildInsertSQL(table, data)
}

// BuildInsertBatch 构建批量INSERT语句
func (b *MySQLBuilderImpl) BuildInsertBatch(table string, data []map[string]interface{}) (string, []interface{}) {
	return b.BaseBuilder.BuildInsertBatchSQL(table, data)
}

// BuildUpdate 构建UPDATE语句
func (b *MySQLBuilderImpl) BuildUpdate(q interface{}, data map[string]interface{}) (string, []interface{}) {
	queryData := ExtractQueryData(q)
	if queryData == nil {
		return "", nil
	}
	return b.BaseBuilder.BuildUpdateSQL(queryData.Table, data, queryData.WhereConditions)
}

// BuildDelete 构建DELETE语句
func (b *MySQLBuilderImpl) BuildDelete(q interface{}) (string, []interface{}) {
	queryData := ExtractQueryData(q)
	if queryData == nil {
		return "", nil
	}
	return b.BaseBuilder.BuildDeleteSQL(queryData.Table, queryData.WhereConditions)
}

// BuildCount 构建COUNT语句
func (b *MySQLBuilderImpl) BuildCount(q interface{}) (string, []interface{}) {
	queryData := ExtractQueryData(q)
	if queryData == nil {
		return "", nil
	}
	return b.BaseBuilder.BuildCountSQL(queryData.Table, queryData.WhereConditions, queryData.JoinClauses)
}

// BuildAggregate 构建聚合语句
func (b *MySQLBuilderImpl) BuildAggregate(q interface{}, fn string, field string) (string, []interface{}) {
	queryData := ExtractQueryData(q)
	if queryData == nil {
		return "", nil
	}
	return b.BaseBuilder.BuildAggregateSQL(queryData.Table, fn, field, queryData.WhereConditions, queryData.JoinClauses)
}

// QuoteIdentifier 引用标识符（MySQL使用反引号）
func (b *MySQLBuilderImpl) QuoteIdentifier(name string) string {
	return b.BaseBuilder.QuoteIdentifier(name)
}

// QuotePlaceholder 返回占位符（MySQL使用?）
func (b *MySQLBuilderImpl) QuotePlaceholder(index int) string {
	return "?"
}

// GetPlaceholderStyle 获取占位符风格
func (b *MySQLBuilderImpl) GetPlaceholderStyle() PlaceholderStyle {
	return PlaceholderQuestion
}

// MySQL特定功能

// BuildInsertIgnore 构建INSERT IGNORE语句（MySQL特有）
func (b *MySQLBuilderImpl) BuildInsertIgnore(table string, data map[string]interface{}) (string, []interface{}) {
	sql, args := b.BuildInsert(table, data)
	if sql == "" {
		return "", nil
	}
	return strings.Replace(sql, "INSERT INTO", "INSERT IGNORE INTO", 1), args
}

// BuildReplace 构建REPLACE语句（MySQL特有）
func (b *MySQLBuilderImpl) BuildReplace(table string, data map[string]interface{}) (string, []interface{}) {
	sql, args := b.BuildInsert(table, data)
	if sql == "" {
		return "", nil
	}
	return strings.Replace(sql, "INSERT INTO", "REPLACE INTO", 1), args
}

// BuildInsertOnDuplicateKeyUpdate 构建INSERT ... ON DUPLICATE KEY UPDATE语句（MySQL特有）
func (b *MySQLBuilderImpl) BuildInsertOnDuplicateKeyUpdate(table string, data map[string]interface{}, updateColumns []string) (string, []interface{}) {
	sql, args := b.BuildInsert(table, data)
	if sql == "" {
		return "", nil
	}

	if len(updateColumns) == 0 {
		return sql, args
	}

	var updateParts []string
	for _, col := range updateColumns {
		updateParts = append(updateParts, fmt.Sprintf("%s = VALUES(%s)", b.QuoteIdentifier(col), b.QuoteIdentifier(col)))
	}

	sql += " ON DUPLICATE KEY UPDATE " + strings.Join(updateParts, ", ")
	return sql, args
}

// BuildLimitOffset 构建MySQL风格的LIMIT OFFSET子句
func (b *MySQLBuilderImpl) BuildLimitOffset(limit, offset int) string {
	return b.BuildLimitOffsetClause(limit, offset)
}

// BuildForUpdate 构建FOR UPDATE子句（行锁）
func (b *MySQLBuilderImpl) BuildForUpdate() string {
	return " FOR UPDATE"
}

// BuildForShare 构建FOR SHARE子句（共享锁）
func (b *MySQLBuilderImpl) BuildForShare() string {
	return " FOR SHARE"
}

// BuildUseIndex 构建USE INDEX提示
func (b *MySQLBuilderImpl) BuildUseIndex(indexes ...string) string {
	if len(indexes) == 0 {
		return ""
	}
	return fmt.Sprintf(" USE INDEX (%s)", strings.Join(indexes, ", "))
}

// BuildForceIndex 构建FORCE INDEX提示
func (b *MySQLBuilderImpl) BuildForceIndex(indexes ...string) string {
	if len(indexes) == 0 {
		return ""
	}
	return fmt.Sprintf(" FORCE INDEX (%s)", strings.Join(indexes, ", "))
}

// BuildIgnoreIndex 构建IGNORE INDEX提示
func (b *MySQLBuilderImpl) BuildIgnoreIndex(indexes ...string) string {
	if len(indexes) == 0 {
		return ""
	}
	return fmt.Sprintf(" IGNORE INDEX (%s)", strings.Join(indexes, ", "))
}
