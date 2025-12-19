package db

import (
	"fmt"
	"strings"
)

// SQLiteBuilderImpl SQLite SQL构建器完整实现
type SQLiteBuilderImpl struct {
	*BaseBuilder
}

// NewSQLiteBuilderImpl 创建SQLite构建器
func NewSQLiteBuilderImpl() *SQLiteBuilderImpl {
	return &SQLiteBuilderImpl{
		BaseBuilder: NewBaseBuilder(PlaceholderQuestion, "\""),
	}
}

// BuildSelect 构建SELECT语句
func (b *SQLiteBuilderImpl) BuildSelect(q interface{}) (string, []interface{}) {
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
func (b *SQLiteBuilderImpl) BuildInsert(table string, data map[string]interface{}) (string, []interface{}) {
	return b.BaseBuilder.BuildInsertSQL(table, data)
}

// BuildInsertBatch 构建批量INSERT语句
func (b *SQLiteBuilderImpl) BuildInsertBatch(table string, data []map[string]interface{}) (string, []interface{}) {
	return b.BaseBuilder.BuildInsertBatchSQL(table, data)
}

// BuildUpdate 构建UPDATE语句
func (b *SQLiteBuilderImpl) BuildUpdate(q interface{}, data map[string]interface{}) (string, []interface{}) {
	queryData := ExtractQueryData(q)
	if queryData == nil {
		return "", nil
	}
	return b.BaseBuilder.BuildUpdateSQL(queryData.Table, data, queryData.WhereConditions)
}

// BuildDelete 构建DELETE语句
func (b *SQLiteBuilderImpl) BuildDelete(q interface{}) (string, []interface{}) {
	queryData := ExtractQueryData(q)
	if queryData == nil {
		return "", nil
	}
	return b.BaseBuilder.BuildDeleteSQL(queryData.Table, queryData.WhereConditions)
}

// BuildCount 构建COUNT语句
func (b *SQLiteBuilderImpl) BuildCount(q interface{}) (string, []interface{}) {
	queryData := ExtractQueryData(q)
	if queryData == nil {
		return "", nil
	}
	return b.BaseBuilder.BuildCountSQL(queryData.Table, queryData.WhereConditions, queryData.JoinClauses)
}

// BuildAggregate 构建聚合语句
func (b *SQLiteBuilderImpl) BuildAggregate(q interface{}, fn string, field string) (string, []interface{}) {
	queryData := ExtractQueryData(q)
	if queryData == nil {
		return "", nil
	}
	return b.BaseBuilder.BuildAggregateSQL(queryData.Table, fn, field, queryData.WhereConditions, queryData.JoinClauses)
}

// QuoteIdentifier 引用标识符（SQLite使用双引号）
func (b *SQLiteBuilderImpl) QuoteIdentifier(name string) string {
	return b.BaseBuilder.QuoteIdentifier(name)
}

// QuotePlaceholder 返回占位符（SQLite使用?）
func (b *SQLiteBuilderImpl) QuotePlaceholder(index int) string {
	return "?"
}

// GetPlaceholderStyle 获取占位符风格
func (b *SQLiteBuilderImpl) GetPlaceholderStyle() PlaceholderStyle {
	return PlaceholderQuestion
}

// SQLite特定功能

// BuildInsertOrReplace 构建INSERT OR REPLACE语句（SQLite特有）
func (b *SQLiteBuilderImpl) BuildInsertOrReplace(table string, data map[string]interface{}) (string, []interface{}) {
	sql, args := b.BuildInsert(table, data)
	if sql == "" {
		return "", nil
	}
	return strings.Replace(sql, "INSERT INTO", "INSERT OR REPLACE INTO", 1), args
}

// BuildInsertOrIgnore 构建INSERT OR IGNORE语句（SQLite特有）
func (b *SQLiteBuilderImpl) BuildInsertOrIgnore(table string, data map[string]interface{}) (string, []interface{}) {
	sql, args := b.BuildInsert(table, data)
	if sql == "" {
		return "", nil
	}
	return strings.Replace(sql, "INSERT INTO", "INSERT OR IGNORE INTO", 1), args
}

// BuildInsertOrAbort 构建INSERT OR ABORT语句（SQLite特有）
func (b *SQLiteBuilderImpl) BuildInsertOrAbort(table string, data map[string]interface{}) (string, []interface{}) {
	sql, args := b.BuildInsert(table, data)
	if sql == "" {
		return "", nil
	}
	return strings.Replace(sql, "INSERT INTO", "INSERT OR ABORT INTO", 1), args
}

// BuildInsertOrRollback 构建INSERT OR ROLLBACK语句（SQLite特有）
func (b *SQLiteBuilderImpl) BuildInsertOrRollback(table string, data map[string]interface{}) (string, []interface{}) {
	sql, args := b.BuildInsert(table, data)
	if sql == "" {
		return "", nil
	}
	return strings.Replace(sql, "INSERT INTO", "INSERT OR ROLLBACK INTO", 1), args
}

// BuildInsertOrFail 构建INSERT OR FAIL语句（SQLite特有）
func (b *SQLiteBuilderImpl) BuildInsertOrFail(table string, data map[string]interface{}) (string, []interface{}) {
	sql, args := b.BuildInsert(table, data)
	if sql == "" {
		return "", nil
	}
	return strings.Replace(sql, "INSERT INTO", "INSERT OR FAIL INTO", 1), args
}

// BuildUpsert 构建UPSERT语句（SQLite 3.24+支持）
func (b *SQLiteBuilderImpl) BuildUpsert(table string, data map[string]interface{}, conflictColumns []string, updateColumns []string) (string, []interface{}) {
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
			updateParts = append(updateParts, fmt.Sprintf("%s = excluded.%s", col, col))
		}
		sql += " DO UPDATE SET " + strings.Join(updateParts, ", ")
	}

	return sql, args
}

// BuildVacuum 构建VACUUM语句（SQLite特有，用于优化数据库）
func (b *SQLiteBuilderImpl) BuildVacuum() string {
	return "VACUUM"
}

// BuildAnalyze 构建ANALYZE语句（SQLite特有，用于更新统计信息）
func (b *SQLiteBuilderImpl) BuildAnalyze(table ...string) string {
	if len(table) == 0 {
		return "ANALYZE"
	}
	return fmt.Sprintf("ANALYZE %s", table[0])
}

// BuildReindex 构建REINDEX语句（SQLite特有）
func (b *SQLiteBuilderImpl) BuildReindex(indexOrTable ...string) string {
	if len(indexOrTable) == 0 {
		return "REINDEX"
	}
	return fmt.Sprintf("REINDEX %s", indexOrTable[0])
}

// BuildPragma 构建PRAGMA语句（SQLite特有）
func (b *SQLiteBuilderImpl) BuildPragma(name string, value ...interface{}) string {
	if len(value) == 0 {
		return fmt.Sprintf("PRAGMA %s", name)
	}
	return fmt.Sprintf("PRAGMA %s = %v", name, value[0])
}

// BuildAttachDatabase 构建ATTACH DATABASE语句（SQLite特有）
func (b *SQLiteBuilderImpl) BuildAttachDatabase(filename, alias string) string {
	return fmt.Sprintf("ATTACH DATABASE '%s' AS %s", filename, alias)
}

// BuildDetachDatabase 构建DETACH DATABASE语句（SQLite特有）
func (b *SQLiteBuilderImpl) BuildDetachDatabase(alias string) string {
	return fmt.Sprintf("DETACH DATABASE %s", alias)
}

// BuildWithoutRowid 返回WITHOUT ROWID表选项（SQLite特有）
func (b *SQLiteBuilderImpl) BuildWithoutRowid() string {
	return " WITHOUT ROWID"
}

// BuildStrict 返回STRICT表选项（SQLite 3.37+特有）
func (b *SQLiteBuilderImpl) BuildStrict() string {
	return " STRICT"
}

// BuildIfNotExists 返回IF NOT EXISTS子句
func (b *SQLiteBuilderImpl) BuildIfNotExists() string {
	return "IF NOT EXISTS "
}

// BuildIfExists 返回IF EXISTS子句
func (b *SQLiteBuilderImpl) BuildIfExists() string {
	return "IF EXISTS "
}

// BuildGlob 构建GLOB表达式（SQLite特有，区分大小写的模式匹配）
func (b *SQLiteBuilderImpl) BuildGlob(column, pattern string) string {
	return fmt.Sprintf("%s GLOB '%s'", column, pattern)
}

// BuildCoalesce 构建COALESCE函数
func (b *SQLiteBuilderImpl) BuildCoalesce(columns ...string) string {
	return fmt.Sprintf("COALESCE(%s)", strings.Join(columns, ", "))
}

// BuildIfNull 构建IFNULL函数（SQLite特有）
func (b *SQLiteBuilderImpl) BuildIfNull(column, defaultValue string) string {
	return fmt.Sprintf("IFNULL(%s, %s)", column, defaultValue)
}

// BuildNullIf 构建NULLIF函数
func (b *SQLiteBuilderImpl) BuildNullIf(column, value string) string {
	return fmt.Sprintf("NULLIF(%s, %s)", column, value)
}

// BuildGroupConcat 构建GROUP_CONCAT聚合函数（SQLite特有）
func (b *SQLiteBuilderImpl) BuildGroupConcat(column string, separator ...string) string {
	if len(separator) == 0 {
		return fmt.Sprintf("GROUP_CONCAT(%s)", column)
	}
	return fmt.Sprintf("GROUP_CONCAT(%s, '%s')", column, separator[0])
}

// BuildTotalChanges 返回total_changes()函数（SQLite特有）
func (b *SQLiteBuilderImpl) BuildTotalChanges() string {
	return "total_changes()"
}

// BuildChanges 返回changes()函数（SQLite特有）
func (b *SQLiteBuilderImpl) BuildChanges() string {
	return "changes()"
}

// BuildLastInsertRowid 返回last_insert_rowid()函数（SQLite特有）
func (b *SQLiteBuilderImpl) BuildLastInsertRowid() string {
	return "last_insert_rowid()"
}
