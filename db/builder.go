package db

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"
)

// QueryBuilder 查询构建器 - TORM的核心
type QueryBuilder struct {
	connection ConnectionInterface
	tableName  string
	model      interface{} // 关联的模型实例

	// 查询组件
	selectColumns    []string
	whereConditions  []WhereCondition
	joinClauses      []JoinClause
	orderByColumns   []OrderByClause
	groupByColumns   []string
	havingConditions []WhereCondition

	// 分页和限制
	limitCount  int
	offsetCount int

	// 事务相关
	transaction TransactionInterface

	// 缓存相关
	cacheEnabled bool
	cacheTTL     time.Duration
	cacheTags    []string
	cacheKey     string

	// 上下文
	ctx context.Context
}

// WhereCondition WHERE条件
type WhereCondition struct {
	Column   string
	Operator string
	Value    interface{}
	Logic    string        // AND, OR
	Raw      string        // 原生SQL
	Values   []interface{} // 原生SQL的参数
}

// JoinClause JOIN子句
type JoinClause struct {
	Type      string // LEFT, RIGHT, INNER
	Table     string
	Condition string
}

// OrderByClause 排序子句
type OrderByClause struct {
	Column    string
	Direction string // ASC, DESC
}

// NewQueryBuilder 创建新的查询构建器
func NewQueryBuilder(connectionName string) (*QueryBuilder, error) {
	conn, err := DefaultManager().Connection(connectionName)
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	return &QueryBuilder{
		connection:       conn,
		selectColumns:    []string{},
		whereConditions:  []WhereCondition{},
		joinClauses:      []JoinClause{},
		orderByColumns:   []OrderByClause{},
		groupByColumns:   []string{},
		havingConditions: []WhereCondition{},
		ctx:              context.Background(),
	}, nil
}

// 注意：Table和Model函数已移至manager.go

// getTableNameFromModel 从模型获取表名
func getTableNameFromModel(model interface{}) string {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 将结构体名转换为蛇形命名
	name := modelType.Name()
	return toSnakeCase(name)
}

// toSnakeCase 转换为蛇形命名
func toSnakeCase(str string) string {
	var result strings.Builder
	for i, r := range str {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r - 'A' + 'a')
	}
	return result.String()
}

// Select 设置选择的字段
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.selectColumns = append(qb.selectColumns, columns...)
	return qb
}

// Where 添加WHERE条件 - 支持多种格式
func (qb *QueryBuilder) Where(args ...interface{}) *QueryBuilder {
	switch len(args) {
	case 1:
		// Where("id IN (1,2,3)") - 纯SQL
		if sql, ok := args[0].(string); ok {
			qb.whereConditions = append(qb.whereConditions, WhereCondition{
				Raw:   sql,
				Logic: "AND",
			})
		}
	case 2:
		// Where("name = ?", value) 或 Where("status IN (?)", []string{"active", "pending"})
		if sql, ok := args[0].(string); ok {
			// 检查第二个参数是否是数组/切片
			if qb.isSliceOrArray(args[1]) {
				// 处理数组参数，如 Where("status IN (?)", []string{"active", "pending"})
				values := qb.convertToInterfaceSlice(args[1])
				if len(values) > 0 {
					// 为数组生成多个占位符
					placeholders := strings.Repeat("?,", len(values))
					placeholders = placeholders[:len(placeholders)-1] // 去掉最后的逗号

					// 替换SQL中的单个?为多个占位符
					processedSQL := strings.Replace(sql, "?", placeholders, 1)

					qb.whereConditions = append(qb.whereConditions, WhereCondition{
						Raw:    processedSQL,
						Values: values,
						Logic:  "AND",
					})
				}
			} else {
				// 普通单值参数
				qb.whereConditions = append(qb.whereConditions, WhereCondition{
					Raw:    sql,
					Values: []interface{}{args[1]},
					Logic:  "AND",
				})
			}
		}
	case 3:
		// Where("name", "=", value)
		if column, ok := args[0].(string); ok {
			if operator, ok := args[1].(string); ok {
				qb.whereConditions = append(qb.whereConditions, WhereCondition{
					Column:   column,
					Operator: operator,
					Value:    args[2],
					Logic:    "AND",
				})
			}
		}
	default:
		// Where("status IN (?, ?)", "active", "pending") - 多参数
		if len(args) > 1 {
			if sql, ok := args[0].(string); ok {
				qb.whereConditions = append(qb.whereConditions, WhereCondition{
					Raw:    sql,
					Values: args[1:], // 剩余所有参数作为值
					Logic:  "AND",
				})
			}
		}
	}
	return qb
}

// OrWhere 添加OR WHERE条件
func (qb *QueryBuilder) OrWhere(args ...interface{}) *QueryBuilder {
	switch len(args) {
	case 1:
		if sql, ok := args[0].(string); ok {
			qb.whereConditions = append(qb.whereConditions, WhereCondition{
				Raw:   sql,
				Logic: "OR",
			})
		}
	case 2:
		if sql, ok := args[0].(string); ok {
			// 检查第二个参数是否是数组/切片
			if qb.isSliceOrArray(args[1]) {
				// 处理数组参数
				values := qb.convertToInterfaceSlice(args[1])
				if len(values) > 0 {
					// 为数组生成多个占位符
					placeholders := strings.Repeat("?,", len(values))
					placeholders = placeholders[:len(placeholders)-1] // 去掉最后的逗号

					// 替换SQL中的单个?为多个占位符
					processedSQL := strings.Replace(sql, "?", placeholders, 1)

					qb.whereConditions = append(qb.whereConditions, WhereCondition{
						Raw:    processedSQL,
						Values: values,
						Logic:  "OR",
					})
				}
			} else {
				// 普通单值参数
				qb.whereConditions = append(qb.whereConditions, WhereCondition{
					Raw:    sql,
					Values: []interface{}{args[1]},
					Logic:  "OR",
				})
			}
		}
	case 3:
		if column, ok := args[0].(string); ok {
			if operator, ok := args[1].(string); ok {
				qb.whereConditions = append(qb.whereConditions, WhereCondition{
					Column:   column,
					Operator: operator,
					Value:    args[2],
					Logic:    "OR",
				})
			}
		}
	default:
		// OrWhere("status IN (?, ?)", "active", "pending") - 多参数
		if len(args) > 1 {
			if sql, ok := args[0].(string); ok {
				qb.whereConditions = append(qb.whereConditions, WhereCondition{
					Raw:    sql,
					Values: args[1:], // 剩余所有参数作为值
					Logic:  "OR",
				})
			}
		}
	}
	return qb
}

// Join 内连接（通用方法，等同于InnerJoin）
func (qb *QueryBuilder) Join(table, localKey, operator, foreignKey string) *QueryBuilder {
	return qb.InnerJoin(table, localKey, operator, foreignKey)
}

// LeftJoin 左连接
func (qb *QueryBuilder) LeftJoin(table, localKey, operator, foreignKey string) *QueryBuilder {
	condition := fmt.Sprintf("%s.%s %s %s.%s", qb.tableName, localKey, operator, table, foreignKey)
	qb.joinClauses = append(qb.joinClauses, JoinClause{
		Type:      "LEFT",
		Table:     table,
		Condition: condition,
	})
	return qb
}

// RightJoin 右连接
func (qb *QueryBuilder) RightJoin(table, localKey, operator, foreignKey string) *QueryBuilder {
	condition := fmt.Sprintf("%s.%s %s %s.%s", qb.tableName, localKey, operator, table, foreignKey)
	qb.joinClauses = append(qb.joinClauses, JoinClause{
		Type:      "RIGHT",
		Table:     table,
		Condition: condition,
	})
	return qb
}

// InnerJoin 内连接
func (qb *QueryBuilder) InnerJoin(table, localKey, operator, foreignKey string) *QueryBuilder {
	condition := fmt.Sprintf("%s.%s %s %s.%s", qb.tableName, localKey, operator, table, foreignKey)
	qb.joinClauses = append(qb.joinClauses, JoinClause{
		Type:      "INNER",
		Table:     table,
		Condition: condition,
	})
	return qb
}

// OrderBy 排序
func (qb *QueryBuilder) OrderBy(column, direction string) *QueryBuilder {
	qb.orderByColumns = append(qb.orderByColumns, OrderByClause{
		Column:    column,
		Direction: strings.ToUpper(direction),
	})
	return qb
}

// GroupBy 分组
func (qb *QueryBuilder) GroupBy(columns ...string) *QueryBuilder {
	qb.groupByColumns = append(qb.groupByColumns, columns...)
	return qb
}

// Having HAVING条件
func (qb *QueryBuilder) Having(args ...interface{}) *QueryBuilder {
	switch len(args) {
	case 1:
		// Having("COUNT(*) > 5") - 纯SQL
		if sql, ok := args[0].(string); ok {
			qb.havingConditions = append(qb.havingConditions, WhereCondition{
				Raw:   sql,
				Logic: "AND",
			})
		}
	case 2:
		// Having("COUNT(*) > ?", 5)
		if sql, ok := args[0].(string); ok {
			qb.havingConditions = append(qb.havingConditions, WhereCondition{
				Raw:    sql,
				Values: []interface{}{args[1]},
				Logic:  "AND",
			})
		}
	case 3:
		// Having("column", ">", value)
		if column, ok := args[0].(string); ok {
			if operator, ok := args[1].(string); ok {
				qb.havingConditions = append(qb.havingConditions, WhereCondition{
					Column:   column,
					Operator: operator,
					Value:    args[2],
					Logic:    "AND",
				})
			}
		}
	default:
		// Having("column IN (?, ?)", value1, value2) - 多参数
		if len(args) > 1 {
			if sql, ok := args[0].(string); ok {
				qb.havingConditions = append(qb.havingConditions, WhereCondition{
					Raw:    sql,
					Values: args[1:],
					Logic:  "AND",
				})
			}
		}
	}
	return qb
}

// Limit 限制返回数量
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limitCount = limit
	return qb
}

// Offset 设置偏移量
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offsetCount = offset
	return qb
}

// Cache 启用查询缓存
func (qb *QueryBuilder) Cache(ttl time.Duration) *QueryBuilder {
	qb.cacheEnabled = true
	qb.cacheTTL = ttl
	return qb
}

// CacheWithTags 启用带标签的查询缓存
func (qb *QueryBuilder) CacheWithTags(ttl time.Duration, tags ...string) *QueryBuilder {
	qb.cacheEnabled = true
	qb.cacheTTL = ttl
	qb.cacheTags = tags
	return qb
}

// CacheKey 设置自定义缓存键
func (qb *QueryBuilder) CacheKey(key string) *QueryBuilder {
	qb.cacheKey = key
	return qb
}

// Get 执行查询并返回结果
func (qb *QueryBuilder) Get() ([]map[string]interface{}, error) {
	// 如果启用了缓存并且不在事务中，尝试从缓存获取
	if qb.cacheEnabled && qb.transaction == nil {
		cacheKey := qb.generateCacheKey()
		if cached, err := GetDefaultCache().Get(cacheKey); err == nil {
			if result, ok := cached.([]map[string]interface{}); ok {
				return result, nil
			}
		}
	}

	sqlStr, args := qb.buildSelectSQL()

	var rows *sql.Rows
	var err error

	if qb.transaction != nil {
		rows, err = qb.transaction.Query(sqlStr, args...)
	} else {
		rows, err = qb.connection.Query(sqlStr, args...)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result, err := qb.scanRows(rows)
	if err != nil {
		return nil, err
	}

	// 如果启用了缓存，将结果存入缓存
	if qb.cacheEnabled && qb.transaction == nil {
		cacheKey := qb.generateCacheKey()
		if len(qb.cacheTags) > 0 {
			if memCache, ok := GetDefaultCache().(*MemoryCache); ok {
				memCache.SetWithTags(cacheKey, result, qb.cacheTTL, qb.cacheTags)
			}
		} else {
			GetDefaultCache().Set(cacheKey, result, qb.cacheTTL)
		}
	}

	return result, nil
}

// First 获取第一条记录
func (qb *QueryBuilder) First() (map[string]interface{}, error) {
	qb.Limit(1)
	results, err := qb.Get()
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("record not found")
	}

	return results[0], nil
}

// Count 计算记录数量
func (qb *QueryBuilder) Count() (int64, error) {
	originalSelect := qb.selectColumns
	qb.selectColumns = []string{"COUNT(*) as count"}

	result, err := qb.First()
	if err != nil {
		return 0, err
	}

	qb.selectColumns = originalSelect

	if count, ok := result["count"].(int64); ok {
		return count, nil
	}

	return 0, fmt.Errorf("count查询结果类型错误")
}

// Insert 插入数据
func (qb *QueryBuilder) Insert(data map[string]interface{}) (int64, error) {
	sqlStr, args := qb.buildInsertSQL(data)
	driverName := qb.getDriverName()

	if driverName == "postgres" {
		// PostgreSQL使用RETURNING获取ID
		if !strings.Contains(sqlStr, "RETURNING") {
			sqlStr += " RETURNING id"
		}

		var lastID int64
		var err error

		if qb.transaction != nil {
			err = qb.transaction.QueryRow(sqlStr, args...).Scan(&lastID)
		} else {
			db := qb.connection.GetDB()
			err = db.QueryRow(sqlStr, args...).Scan(&lastID)
		}

		if err != nil {
			// 如果没有id字段或其他错误，仍然执行插入但返回受影响行数
			originalSQL := strings.Replace(sqlStr, " RETURNING id", "", 1)
			var result interface{}

			if qb.transaction != nil {
				result, err = qb.transaction.Exec(originalSQL, args...)
			} else {
				result, err = qb.connection.Exec(originalSQL, args...)
			}

			if err != nil {
				return 0, fmt.Errorf("failed to execute PostgreSQL statement: %w", err)
			}

			if sqlResult, ok := result.(interface{ RowsAffected() (int64, error) }); ok {
				return sqlResult.RowsAffected()
			}
			return 0, nil
		}
		return lastID, nil
	} else {
		// MySQL, SQLite使用LastInsertId
		var result interface{}
		var err error

		if qb.transaction != nil {
			result, err = qb.transaction.Exec(sqlStr, args...)
		} else {
			result, err = qb.connection.Exec(sqlStr, args...)
		}

		if err != nil {
			return 0, fmt.Errorf("failed to execute %s statement: %w", driverName, err)
		}

		// 类型断言
		if sqlResult, ok := result.(interface{ LastInsertId() (int64, error) }); ok {
			return sqlResult.LastInsertId()
		}
		return 0, fmt.Errorf("无法获取插入ID")
	}
}

// Update 更新数据
func (qb *QueryBuilder) Update(data map[string]interface{}) (int64, error) {
	sqlStr, args := qb.buildUpdateSQL(data)

	if qb.transaction != nil {
		result, err := qb.transaction.Exec(sqlStr, args...)
		if err != nil {
			return 0, err
		}
		return result.RowsAffected()
	} else {
		result, err := qb.connection.Exec(sqlStr, args...)
		if err != nil {
			return 0, err
		}
		return result.RowsAffected()
	}
}

// Delete 删除数据
func (qb *QueryBuilder) Delete() (int64, error) {
	sqlStr, args := qb.buildDeleteSQL()

	if qb.transaction != nil {
		result, err := qb.transaction.Exec(sqlStr, args...)
		if err != nil {
			return 0, err
		}
		return result.RowsAffected()
	} else {
		result, err := qb.connection.Exec(sqlStr, args...)
		if err != nil {
			return 0, err
		}
		return result.RowsAffected()
	}
}

// buildSelectSQL 构建SELECT SQL
func (qb *QueryBuilder) buildSelectSQL() (string, []interface{}) {
	var sql strings.Builder
	var args []interface{}
	argIndex := 0

	// SELECT子句
	sql.WriteString("SELECT ")
	if len(qb.selectColumns) > 0 {
		sql.WriteString(strings.Join(qb.selectColumns, ", "))
	} else {
		sql.WriteString("*")
	}

	// FROM子句
	sql.WriteString(" FROM ")
	sql.WriteString(qb.tableName)

	// JOIN子句
	for _, join := range qb.joinClauses {
		sql.WriteString(fmt.Sprintf(" %s JOIN %s ON %s", join.Type, join.Table, join.Condition))
	}

	// WHERE子句
	if len(qb.whereConditions) > 0 {
		sql.WriteString(" WHERE ")
		for i, condition := range qb.whereConditions {
			if i > 0 {
				sql.WriteString(" " + condition.Logic + " ")
			}

			if condition.Raw != "" {
				processedSQL := qb.processPlaceholders(condition.Raw, argIndex)
				sql.WriteString(processedSQL)
				if len(condition.Values) > 0 {
					args = append(args, condition.Values...)
					argIndex += len(condition.Values)
				}
			} else {
				placeholder := qb.buildPlaceholder(argIndex)
				sql.WriteString(fmt.Sprintf("%s %s %s", condition.Column, condition.Operator, placeholder))
				args = append(args, condition.Value)
				argIndex++
			}
		}
	}

	// GROUP BY子句
	if len(qb.groupByColumns) > 0 {
		sql.WriteString(" GROUP BY ")
		sql.WriteString(strings.Join(qb.groupByColumns, ", "))
	}

	// HAVING子句
	if len(qb.havingConditions) > 0 {
		sql.WriteString(" HAVING ")
		for i, condition := range qb.havingConditions {
			if i > 0 {
				sql.WriteString(" " + condition.Logic + " ")
			}

			if condition.Raw != "" {
				processedSQL := qb.processPlaceholders(condition.Raw, argIndex)
				sql.WriteString(processedSQL)
				if len(condition.Values) > 0 {
					args = append(args, condition.Values...)
					argIndex += len(condition.Values)
				}
			} else {
				placeholder := qb.buildPlaceholder(argIndex)
				sql.WriteString(fmt.Sprintf("%s %s %s", condition.Column, condition.Operator, placeholder))
				args = append(args, condition.Value)
				argIndex++
			}
		}
	}

	// ORDER BY子句
	if len(qb.orderByColumns) > 0 {
		sql.WriteString(" ORDER BY ")
		orderParts := make([]string, len(qb.orderByColumns))
		for i, order := range qb.orderByColumns {
			orderParts[i] = fmt.Sprintf("%s %s", order.Column, order.Direction)
		}
		sql.WriteString(strings.Join(orderParts, ", "))
	}

	// LIMIT和OFFSET子句
	if qb.limitCount > 0 {
		sql.WriteString(fmt.Sprintf(" LIMIT %d", qb.limitCount))
		if qb.offsetCount > 0 {
			sql.WriteString(fmt.Sprintf(" OFFSET %d", qb.offsetCount))
		}
	}

	return sql.String(), args
}

// buildInsertSQL 构建INSERT SQL
func (qb *QueryBuilder) buildInsertSQL(data map[string]interface{}) (string, []interface{}) {
	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	args := make([]interface{}, 0, len(data))

	for column, value := range data {
		columns = append(columns, column)
		args = append(args, value)
	}

	// 根据数据库类型生成占位符
	driverName := qb.getDriverName()
	for i := range columns {
		if driverName == "postgres" {
			placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		} else {
			placeholders = append(placeholders, "?")
		}
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		qb.tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	return sql, args
}

// buildUpdateSQL 构建UPDATE SQL
func (qb *QueryBuilder) buildUpdateSQL(data map[string]interface{}) (string, []interface{}) {
	var sql strings.Builder
	var args []interface{}

	sql.WriteString("UPDATE ")
	sql.WriteString(qb.tableName)
	sql.WriteString(" SET ")

	setParts := make([]string, 0, len(data))
	argIndex := 0
	for column, value := range data {
		placeholder := qb.buildPlaceholder(argIndex)
		setParts = append(setParts, column+" = "+placeholder)
		args = append(args, value)
		argIndex++
	}
	sql.WriteString(strings.Join(setParts, ", "))

	// WHERE子句
	if len(qb.whereConditions) > 0 {
		sql.WriteString(" WHERE ")
		for i, condition := range qb.whereConditions {
			if i > 0 {
				sql.WriteString(" " + condition.Logic + " ")
			}

			if condition.Raw != "" {
				// 处理原始SQL中的占位符
				processedSQL := qb.processPlaceholders(condition.Raw, argIndex)
				sql.WriteString(processedSQL)
				if len(condition.Values) > 0 {
					args = append(args, condition.Values...)
					argIndex += len(condition.Values)
				}
			} else {
				placeholder := qb.buildPlaceholder(argIndex)
				sql.WriteString(fmt.Sprintf("%s %s %s", condition.Column, condition.Operator, placeholder))
				args = append(args, condition.Value)
				argIndex++
			}
		}
	}

	return sql.String(), args
}

// buildDeleteSQL 构建DELETE SQL
func (qb *QueryBuilder) buildDeleteSQL() (string, []interface{}) {
	var sql strings.Builder
	var args []interface{}
	argIndex := 0

	sql.WriteString("DELETE FROM ")
	sql.WriteString(qb.tableName)

	// WHERE子句
	if len(qb.whereConditions) > 0 {
		sql.WriteString(" WHERE ")
		for i, condition := range qb.whereConditions {
			if i > 0 {
				sql.WriteString(" " + condition.Logic + " ")
			}

			if condition.Raw != "" {
				processedSQL := qb.processPlaceholders(condition.Raw, argIndex)
				sql.WriteString(processedSQL)
				if len(condition.Values) > 0 {
					args = append(args, condition.Values...)
					argIndex += len(condition.Values)
				}
			} else {
				placeholder := qb.buildPlaceholder(argIndex)
				sql.WriteString(fmt.Sprintf("%s %s %s", condition.Column, condition.Operator, placeholder))
				args = append(args, condition.Value)
				argIndex++
			}
		}
	}

	return sql.String(), args
}

// scanRows 扫描行数据
func (qb *QueryBuilder) scanRows(rows *sql.Rows) ([]map[string]interface{}, error) {
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
			row[column] = qb.convertDatabaseValue(values[i])
		}

		results = append(results, row)
	}

	return results, nil
}

// convertDatabaseValue 转换数据库返回值为合适的Go类型
func (qb *QueryBuilder) convertDatabaseValue(value interface{}) interface{} {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []byte:
		// 智能处理字节数组 - 需要判断其实际内容类型
		return qb.convertByteArraySmart(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		// 注意：uint64可能超出int64范围，但为了一致性转换为int64
		if v <= 9223372036854775807 { // math.MaxInt64
			return int64(v)
		}
		return v // 保持原类型，避免溢出
	case float32:
		return float64(v)
	case float64:
		return v
	case bool:
		return v
	case string:
		// 字符串可能是Base64编码的，但要谨慎处理
		return qb.tryBase64DecodeIfText(v)
	case time.Time:
		return v
	default:
		// 对于其他复杂类型，尝试转换为字符串
		if stringer, ok := v.(fmt.Stringer); ok {
			return stringer.String()
		}
		return v
	}
}

// convertByteArraySmart 智能转换字节数组
func (qb *QueryBuilder) convertByteArraySmart(data []byte) interface{} {
	if len(data) == 0 {
		return ""
	}

	// 检查是否是有效的UTF-8字符串
	if !utf8.Valid(data) {
		// 如果不是有效UTF-8，可能是二进制数据，返回原字节数组
		return data
	}

	// 转换为字符串用于分析
	str := string(data)

	// 1. NULL值处理
	if qb.isNullValue(str) {
		return nil
	}

	// 2. 检查是否是整数
	if intVal, ok := qb.tryParseInteger(str); ok {
		return intVal
	}

	// 3. 检查是否是浮点数
	if floatVal, ok := qb.tryParseFloat(str); ok {
		return floatVal
	}

	// 4. 检查是否是布尔值
	if boolVal, ok := qb.tryParseBool(str); ok {
		return boolVal
	}

	// 5. 检查是否是时间/日期
	if timeVal, ok := qb.tryParseTime(str); ok {
		return timeVal
	}

	// 6. 检查是否是JSON
	if jsonVal, ok := qb.tryParseJSON(str); ok {
		return jsonVal
	}

	// 7. 检查是否是UUID
	if qb.isUUID(str) {
		return str // UUID保持为字符串
	}

	// 8. 检查是否应该跳过Base64解码
	if qb.shouldSkipBase64Decode(str) {
		return str
	}

	// 9. 最后尝试Base64解码
	return qb.tryBase64DecodeIfText(str)
}

// isNullValue 检查是否是NULL值表示
func (qb *QueryBuilder) isNullValue(str string) bool {
	switch strings.ToLower(strings.TrimSpace(str)) {
	case "null", "nil", "<null>", "\\n":
		return true
	}
	return false
}

// tryParseInteger 尝试解析整数
func (qb *QueryBuilder) tryParseInteger(str string) (int64, bool) {
	str = strings.TrimSpace(str)
	if str == "" {
		return 0, false
	}

	// 使用正则表达式检查是否是纯整数格式
	intRegex := regexp.MustCompile(`^[+-]?\d+$`)
	if !intRegex.MatchString(str) {
		return 0, false
	}

	if val, err := strconv.ParseInt(str, 10, 64); err == nil {
		return val, true
	}
	return 0, false
}

// tryParseFloat 尝试解析浮点数
func (qb *QueryBuilder) tryParseFloat(str string) (float64, bool) {
	str = strings.TrimSpace(str)
	if str == "" {
		return 0, false
	}

	// 使用正则表达式检查是否是浮点数格式
	floatRegex := regexp.MustCompile(`^[+-]?(\d+\.?\d*|\d*\.\d+)([eE][+-]?\d+)?$`)
	if !floatRegex.MatchString(str) {
		return 0, false
	}

	// 必须包含小数点或科学记数法才认为是浮点数
	if !strings.Contains(str, ".") && !strings.ContainsAny(strings.ToLower(str), "e") {
		return 0, false
	}

	if val, err := strconv.ParseFloat(str, 64); err == nil {
		return val, true
	}
	return 0, false
}

// tryParseBool 尝试解析布尔值
func (qb *QueryBuilder) tryParseBool(str string) (bool, bool) {
	str = strings.ToLower(strings.TrimSpace(str))
	switch str {
	case "true", "yes", "on", "1", "y", "t":
		return true, true
	case "false", "no", "off", "0", "n", "f":
		return false, true
	}
	return false, false
}

// tryParseTime 尝试解析时间
func (qb *QueryBuilder) tryParseTime(str string) (time.Time, bool) {
	str = strings.TrimSpace(str)
	if str == "" {
		return time.Time{}, false
	}

	// 常见的时间格式
	timeFormats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02T15:04:05.999999999",
		"2006-01-02",
		"15:04:05",
		"2006/01/02 15:04:05",
		"2006/01/02",
		"01/02/2006",
		"01-02-2006",
	}

	for _, format := range timeFormats {
		if t, err := time.Parse(format, str); err == nil {
			return t, true
		}
	}

	// 尝试解析Unix时间戳
	if qb.looksLikeTimestamp(str) {
		if timestamp, err := strconv.ParseInt(str, 10, 64); err == nil {
			// 区分秒级和毫秒级时间戳
			if timestamp > 1e10 { // 毫秒级时间戳
				return time.Unix(timestamp/1000, (timestamp%1000)*1000000), true
			} else { // 秒级时间戳
				return time.Unix(timestamp, 0), true
			}
		}
	}

	return time.Time{}, false
}

// tryParseJSON 尝试解析JSON
func (qb *QueryBuilder) tryParseJSON(str string) (interface{}, bool) {
	str = strings.TrimSpace(str)
	if str == "" {
		return nil, false
	}

	// 快速检查是否看起来像JSON
	if !qb.looksLikeJSON(str) {
		return nil, false
	}

	var result interface{}
	if err := json.Unmarshal([]byte(str), &result); err == nil {
		return result, true
	}

	return nil, false
}

// isUUID 检查是否是UUID格式
func (qb *QueryBuilder) isUUID(str string) bool {
	// UUID正则表达式
	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
	return uuidRegex.MatchString(str)
}

// shouldSkipBase64Decode 判断是否应该跳过Base64解码
func (qb *QueryBuilder) shouldSkipBase64Decode(str string) bool {
	// 1. 长度太短
	if len(str) < 4 {
		return true
	}

	// 2. 包含明显的非Base64特征
	if strings.ContainsAny(str, "@./\\:") {
		return true
	}

	// 3. 看起来像URL、邮箱等
	if qb.looksLikeURL(str) || qb.looksLikeEmail(str) {
		return true
	}

	// 4. 包含常见的中文词汇（避免误解码）
	commonChineseWords := []string{"用户", "管理", "系统", "数据", "测试", "高级", "初级", "活跃", "状态"}
	for _, word := range commonChineseWords {
		if strings.Contains(str, word) {
			return true
		}
	}

	// 5. 如果包含中文字符，通常不是Base64
	for _, r := range str {
		if r > 127 { // 非ASCII字符
			return true
		}
	}

	return false
}

// looksLikeJSON 检查是否看起来像JSON
func (qb *QueryBuilder) looksLikeJSON(str string) bool {
	str = strings.TrimSpace(str)
	return (strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}")) ||
		(strings.HasPrefix(str, "[") && strings.HasSuffix(str, "]"))
}

// looksLikeURL 检查是否看起来像URL
func (qb *QueryBuilder) looksLikeURL(str string) bool {
	return strings.HasPrefix(str, "http://") ||
		strings.HasPrefix(str, "https://") ||
		strings.HasPrefix(str, "ftp://") ||
		strings.Contains(str, "://")
}

// looksLikeEmail 检查是否看起来像邮箱
func (qb *QueryBuilder) looksLikeEmail(str string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(str)
}

// tryBase64DecodeIfText 只对真正的文本字段尝试Base64解码
func (qb *QueryBuilder) tryBase64DecodeIfText(str string) string {
	// 空字符串直接返回
	if str == "" {
		return str
	}

	// 检查是否符合Base64格式
	if !qb.isValidBase64Format(str) {
		return str
	}

	// 尝试解码
	decoded, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		// 尝试URL安全的Base64解码
		if decoded, err = base64.URLEncoding.DecodeString(str); err != nil {
			return str
		}
	}

	// 检查解码后的内容是否是有效的UTF-8
	if !utf8.Valid(decoded) {
		return str
	}

	decodedStr := string(decoded)

	// 验证解码后的内容是否看起来像有意义的文本
	if !qb.isDecodedTextMeaningful(decodedStr) {
		return str
	}

	// 解码成功且内容合理，返回解码后的字符串
	return decodedStr
}

// isValidBase64Format 检查是否是有效的Base64格式
func (qb *QueryBuilder) isValidBase64Format(str string) bool {
	// 长度必须是4的倍数（除非使用了填充）
	if len(str) < 4 {
		return false
	}

	// Base64字符集检查
	base64Regex := regexp.MustCompile(`^[A-Za-z0-9+/]*={0,2}$`)
	if !base64Regex.MatchString(str) {
		// 尝试URL安全的Base64
		base64URLRegex := regexp.MustCompile(`^[A-Za-z0-9_-]*={0,2}$`)
		if !base64URLRegex.MatchString(str) {
			return false
		}
	}

	// 填充字符只能出现在末尾
	if strings.Contains(str[:len(str)-2], "=") {
		return false
	}

	return true
}

// isDecodedTextMeaningful 验证解码后的内容是否有意义
func (qb *QueryBuilder) isDecodedTextMeaningful(decoded string) bool {
	// 空字符串认为有效
	if decoded == "" {
		return true
	}

	// 长度检查：太短或太长都可能不是有意义的文本
	if len(decoded) < 1 || len(decoded) > 10000 {
		return false
	}

	// 检查控制字符比例
	controlCharCount := 0
	printableCharCount := 0

	for _, r := range decoded {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			controlCharCount++
		} else {
			printableCharCount++
		}
	}

	// 如果控制字符过多，可能不是文本
	if controlCharCount > 0 && float64(controlCharCount)/float64(len(decoded)) > 0.1 {
		return false
	}

	// 必须有可打印字符
	if printableCharCount == 0 {
		return false
	}

	// 检查是否包含常见的无意义字符序列
	meaninglessPatterns := []string{
		"\x00\x00\x00",                         // 连续的空字节
		"\xff\xff\xff",                         // 连续的0xFF
		string([]byte{0x89, 0x50, 0x4E, 0x47}), // PNG文件头
		string([]byte{0xFF, 0xD8, 0xFF}),       // JPEG文件头
		"BM",                                   // BMP文件头
		"GIF",                                  // GIF文件头
	}

	for _, pattern := range meaninglessPatterns {
		if strings.Contains(decoded, pattern) {
			return false
		}
	}

	return true
}

// looksLikeTimestamp 检查字符串是否看起来像时间戳（保留用于时间解析）
func (qb *QueryBuilder) looksLikeTimestamp(str string) bool {
	// Unix时间戳通常是10位或13位数字
	if len(str) == 10 || len(str) == 13 {
		for _, c := range str {
			if c < '0' || c > '9' {
				return false
			}
		}
		return true
	}
	return false
}

// InTransaction 在事务中执行
func (qb *QueryBuilder) InTransaction(tx TransactionInterface) *QueryBuilder {
	newBuilder := *qb
	newBuilder.transaction = tx
	return &newBuilder
}

// Connection 设置连接
func (qb *QueryBuilder) Connection(connectionName string) *QueryBuilder {
	conn, err := DefaultManager().Connection(connectionName)
	if err == nil {
		qb.connection = conn
	}
	return qb
}

// Context 设置上下文
func (qb *QueryBuilder) Context(ctx context.Context) *QueryBuilder {
	qb.ctx = ctx
	return qb
}

// isSliceOrArray 检查值是否是切片或数组
func (qb *QueryBuilder) isSliceOrArray(value interface{}) bool {
	if value == nil {
		return false
	}

	switch value.(type) {
	case []string, []int, []int8, []int16, []int32, []int64,
		[]uint, []uint8, []uint16, []uint32, []uint64,
		[]float32, []float64, []bool, []interface{}:
		return true
	default:
		// 使用反射检查其他类型的切片
		v := reflect.ValueOf(value)
		return v.Kind() == reflect.Slice || v.Kind() == reflect.Array
	}
}

// getDriverName 获取数据库驱动名称
func (qb *QueryBuilder) getDriverName() string {
	if qb.connection != nil {
		return qb.connection.GetDriver()
	}
	return ""
}

// buildPlaceholder 根据数据库类型构建占位符
func (qb *QueryBuilder) buildPlaceholder(index int) string {
	driverName := qb.getDriverName()
	if driverName == "postgres" {
		return fmt.Sprintf("$%d", index+1)
	}
	return "?"
}

// processPlaceholders 处理原始SQL中的占位符
func (qb *QueryBuilder) processPlaceholders(sql string, startIndex int) string {
	driverName := qb.getDriverName()
	if driverName != "postgres" {
		return sql // MySQL和SQLite使用?占位符，无需转换
	}

	// PostgreSQL需要将?转换为$1, $2...
	result := sql
	placeholderCount := strings.Count(sql, "?")
	for i := 0; i < placeholderCount; i++ {
		placeholder := fmt.Sprintf("$%d", startIndex+i+1)
		result = strings.Replace(result, "?", placeholder, 1)
	}
	return result
}

// convertToInterfaceSlice 将各种类型的切片转换为[]interface{}
func (qb *QueryBuilder) convertToInterfaceSlice(value interface{}) []interface{} {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
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
	case []int8:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	case []int16:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	case []int32:
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
	case []uint:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	case []uint8:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	case []uint16:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	case []uint32:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	case []uint64:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	case []float32:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	case []float64:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	case []bool:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	case []interface{}:
		return v
	default:
		// 使用反射处理其他类型的切片
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			result := make([]interface{}, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				result[i] = rv.Index(i).Interface()
			}
			return result
		}
		return nil
	}
}

// generateCacheKey 生成缓存键
func (qb *QueryBuilder) generateCacheKey() string {
	if qb.cacheKey != "" {
		return qb.cacheKey
	}

	cacheData := map[string]interface{}{
		"table":  qb.tableName,
		"select": qb.selectColumns,
		"where":  qb.whereConditions,
		"join":   qb.joinClauses,
		"group":  qb.groupByColumns,
		"having": qb.havingConditions,
		"order":  qb.orderByColumns,
		"limit":  qb.limitCount,
		"offset": qb.offsetCount,
	}

	return GenerateCacheKey("query", cacheData)
}
