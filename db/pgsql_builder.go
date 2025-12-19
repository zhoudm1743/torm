package db

import (
	"fmt"
	"strings"
)

// PostgreSQLBuilderImpl PostgreSQL SQL构建器完整实现
type PostgreSQLBuilderImpl struct {
	*BaseBuilder
}

// NewPostgreSQLBuilderImpl 创建PostgreSQL构建器
func NewPostgreSQLBuilderImpl() *PostgreSQLBuilderImpl {
	return &PostgreSQLBuilderImpl{
		BaseBuilder: NewBaseBuilder(PlaceholderDollar, "\""),
	}
}

// BuildSelect 构建SELECT语句
func (b *PostgreSQLBuilderImpl) BuildSelect(q interface{}) (string, []interface{}) {
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
func (b *PostgreSQLBuilderImpl) BuildInsert(table string, data map[string]interface{}) (string, []interface{}) {
	return b.BaseBuilder.BuildInsertSQL(table, data)
}

// BuildInsertBatch 构建批量INSERT语句
func (b *PostgreSQLBuilderImpl) BuildInsertBatch(table string, data []map[string]interface{}) (string, []interface{}) {
	return b.BaseBuilder.BuildInsertBatchSQL(table, data)
}

// BuildUpdate 构建UPDATE语句
func (b *PostgreSQLBuilderImpl) BuildUpdate(q interface{}, data map[string]interface{}) (string, []interface{}) {
	queryData := ExtractQueryData(q)
	if queryData == nil {
		return "", nil
	}
	return b.BaseBuilder.BuildUpdateSQL(queryData.Table, data, queryData.WhereConditions)
}

// BuildDelete 构建DELETE语句
func (b *PostgreSQLBuilderImpl) BuildDelete(q interface{}) (string, []interface{}) {
	queryData := ExtractQueryData(q)
	if queryData == nil {
		return "", nil
	}
	return b.BaseBuilder.BuildDeleteSQL(queryData.Table, queryData.WhereConditions)
}

// BuildCount 构建COUNT语句
func (b *PostgreSQLBuilderImpl) BuildCount(q interface{}) (string, []interface{}) {
	queryData := ExtractQueryData(q)
	if queryData == nil {
		return "", nil
	}
	return b.BaseBuilder.BuildCountSQL(queryData.Table, queryData.WhereConditions, queryData.JoinClauses)
}

// BuildAggregate 构建聚合语句
func (b *PostgreSQLBuilderImpl) BuildAggregate(q interface{}, fn string, field string) (string, []interface{}) {
	queryData := ExtractQueryData(q)
	if queryData == nil {
		return "", nil
	}
	return b.BaseBuilder.BuildAggregateSQL(queryData.Table, fn, field, queryData.WhereConditions, queryData.JoinClauses)
}

// QuoteIdentifier 引用标识符（PostgreSQL使用双引号）
func (b *PostgreSQLBuilderImpl) QuoteIdentifier(name string) string {
	return b.BaseBuilder.QuoteIdentifier(name)
}

// QuotePlaceholder 返回占位符（PostgreSQL使用$N）
func (b *PostgreSQLBuilderImpl) QuotePlaceholder(index int) string {
	return fmt.Sprintf("$%d", index+1)
}

// GetPlaceholderStyle 获取占位符风格
func (b *PostgreSQLBuilderImpl) GetPlaceholderStyle() PlaceholderStyle {
	return PlaceholderDollar
}

// PostgreSQL特定功能

// BuildInsertReturning 构建INSERT ... RETURNING语句（PostgreSQL特有）
func (b *PostgreSQLBuilderImpl) BuildInsertReturning(table string, data map[string]interface{}, returning ...string) (string, []interface{}) {
	sql, args := b.BuildInsert(table, data)
	if sql == "" {
		return "", nil
	}

	if len(returning) == 0 {
		returning = []string{"id"}
	}

	sql += " RETURNING " + strings.Join(returning, ", ")
	return sql, args
}

// BuildInsertOnConflict 构建INSERT ... ON CONFLICT语句（PostgreSQL特有，UPSERT）
func (b *PostgreSQLBuilderImpl) BuildInsertOnConflict(table string, data map[string]interface{}, conflictColumns []string, updateColumns []string) (string, []interface{}) {
	sql, args := b.BuildInsert(table, data)
	if sql == "" {
		return "", nil
	}

	if len(conflictColumns) == 0 {
		return sql, args
	}

	sql += fmt.Sprintf(" ON CONFLICT (%s)", strings.Join(conflictColumns, ", "))

	if len(updateColumns) == 0 {
		sql += " DO NOTHING"
	} else {
		var updateParts []string
		for _, col := range updateColumns {
			updateParts = append(updateParts, fmt.Sprintf("%s = EXCLUDED.%s", col, col))
		}
		sql += " DO UPDATE SET " + strings.Join(updateParts, ", ")
	}

	return sql, args
}

// BuildUpdateReturning 构建UPDATE ... RETURNING语句（PostgreSQL特有）
func (b *PostgreSQLBuilderImpl) BuildUpdateReturning(q interface{}, data map[string]interface{}, returning ...string) (string, []interface{}) {
	sql, args := b.BuildUpdate(q, data)
	if sql == "" {
		return "", nil
	}

	if len(returning) == 0 {
		returning = []string{"*"}
	}

	sql += " RETURNING " + strings.Join(returning, ", ")
	return sql, args
}

// BuildDeleteReturning 构建DELETE ... RETURNING语句（PostgreSQL特有）
func (b *PostgreSQLBuilderImpl) BuildDeleteReturning(q interface{}, returning ...string) (string, []interface{}) {
	sql, args := b.BuildDelete(q)
	if sql == "" {
		return "", nil
	}

	if len(returning) == 0 {
		returning = []string{"*"}
	}

	sql += " RETURNING " + strings.Join(returning, ", ")
	return sql, args
}

// BuildForUpdate 构建FOR UPDATE子句（行锁）
func (b *PostgreSQLBuilderImpl) BuildForUpdate() string {
	return " FOR UPDATE"
}

// BuildForShare 构建FOR SHARE子句（共享锁）
func (b *PostgreSQLBuilderImpl) BuildForShare() string {
	return " FOR SHARE"
}

// BuildForNoKeyUpdate 构建FOR NO KEY UPDATE子句（PostgreSQL特有）
func (b *PostgreSQLBuilderImpl) BuildForNoKeyUpdate() string {
	return " FOR NO KEY UPDATE"
}

// BuildForKeyShare 构建FOR KEY SHARE子句（PostgreSQL特有）
func (b *PostgreSQLBuilderImpl) BuildForKeyShare() string {
	return " FOR KEY SHARE"
}

// BuildSkipLocked 构建SKIP LOCKED子句（PostgreSQL特有）
func (b *PostgreSQLBuilderImpl) BuildSkipLocked() string {
	return " SKIP LOCKED"
}

// BuildNowait 构建NOWAIT子句（PostgreSQL特有）
func (b *PostgreSQLBuilderImpl) BuildNowait() string {
	return " NOWAIT"
}

// BuildDistinctOn 构建DISTINCT ON子句（PostgreSQL特有）
func (b *PostgreSQLBuilderImpl) BuildDistinctOn(columns ...string) string {
	if len(columns) == 0 {
		return ""
	}
	return fmt.Sprintf("DISTINCT ON (%s) ", strings.Join(columns, ", "))
}

// BuildArrayAgg 构建ARRAY_AGG聚合函数（PostgreSQL特有）
func (b *PostgreSQLBuilderImpl) BuildArrayAgg(column string, orderBy ...string) string {
	sql := fmt.Sprintf("ARRAY_AGG(%s", column)
	if len(orderBy) > 0 {
		sql += " ORDER BY " + strings.Join(orderBy, ", ")
	}
	sql += ")"
	return sql
}

// BuildJsonAgg 构建JSON_AGG聚合函数（PostgreSQL特有）
func (b *PostgreSQLBuilderImpl) BuildJsonAgg(column string) string {
	return fmt.Sprintf("JSON_AGG(%s)", column)
}

// BuildJsonbAgg 构建JSONB_AGG聚合函数（PostgreSQL特有）
func (b *PostgreSQLBuilderImpl) BuildJsonbAgg(column string) string {
	return fmt.Sprintf("JSONB_AGG(%s)", column)
}

// BuildCTE 构建WITH (CTE)子句（PostgreSQL支持更丰富的CTE）
func (b *PostgreSQLBuilderImpl) BuildCTE(name string, query string, recursive bool) string {
	if recursive {
		return fmt.Sprintf("WITH RECURSIVE %s AS (%s)", name, query)
	}
	return fmt.Sprintf("WITH %s AS (%s)", name, query)
}
