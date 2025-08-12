package query

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zhoudm1743/torm/db"
)

// AdvancedQueryBuilder 高级查询构建器
type AdvancedQueryBuilder struct {
	base db.QueryInterface
}

// NewAdvancedQueryBuilder 创建高级查询构建器
func NewAdvancedQueryBuilder(base db.QueryInterface) *AdvancedQueryBuilder {
	return &AdvancedQueryBuilder{
		base: base,
	}
}

// JSON查询支持

// WhereJSON JSON字段查询
func (aq *AdvancedQueryBuilder) WhereJSON(column, path, operator string, value interface{}) *AdvancedQueryBuilder {
	// 构建JSON路径查询
	jsonPath := aq.buildJSONPath(column, path)
	aq.base = aq.base.WhereRaw(fmt.Sprintf("%s %s ?", jsonPath, operator), value)
	return aq
}

// WhereJSONContains JSON字段包含查询
func (aq *AdvancedQueryBuilder) WhereJSONContains(column, path string, value interface{}) *AdvancedQueryBuilder {
	// 不同数据库的JSON包含查询语法不同
	driver := aq.getDriver()

	switch driver {
	case "mysql":
		// MySQL: JSON_CONTAINS(column, JSON_QUOTE(value), path)
		jsonValue, _ := json.Marshal(value)
		aq.base = aq.base.WhereRaw(
			fmt.Sprintf("JSON_CONTAINS(%s, ?, %s)", column, aq.quoteJSONPath(path)),
			string(jsonValue),
		)
	case "postgres", "postgresql":
		// PostgreSQL: column @> '{"path": "value"}'::jsonb
		if path == "$" {
			jsonValue, _ := json.Marshal(value)
			aq.base = aq.base.WhereRaw(fmt.Sprintf("%s @> ?::jsonb", column), string(jsonValue))
		} else {
			// 构建嵌套JSON对象
			jsonPath := aq.buildPostgreSQLJSONPath(path, value)
			aq.base = aq.base.WhereRaw(fmt.Sprintf("%s @> ?::jsonb", column), jsonPath)
		}
	case "sqlite":
		// SQLite: JSON查询支持有限，使用LIKE替代
		jsonValue, _ := json.Marshal(value)
		aq.base = aq.base.Where(column, "LIKE", fmt.Sprintf("%%%s%%", string(jsonValue)))
	default:
		// 默认使用原生字符串查询
		aq.base = aq.base.Where(column, "LIKE", fmt.Sprintf("%%%v%%", value))
	}

	return aq
}

// WhereJSONLength JSON数组长度查询
func (aq *AdvancedQueryBuilder) WhereJSONLength(column, path, operator string, length int) *AdvancedQueryBuilder {
	driver := aq.getDriver()

	switch driver {
	case "mysql":
		// MySQL: JSON_LENGTH(column, path) = length
		jsonPath := aq.quoteJSONPath(path)
		aq.base = aq.base.WhereRaw(
			fmt.Sprintf("JSON_LENGTH(%s, %s) %s ?", column, jsonPath, operator),
			length,
		)
	case "postgres", "postgresql":
		// PostgreSQL: jsonb_array_length(column->path) = length
		pgPath := aq.buildPostgreSQLPath(path)
		aq.base = aq.base.WhereRaw(
			fmt.Sprintf("jsonb_array_length(%s%s) %s ?", column, pgPath, operator),
			length,
		)
	default:
		// 简单实现：将JSON解析后检查长度
		aq.base = aq.base.WhereRaw(fmt.Sprintf("LENGTH(%s) %s ?", column, operator), length)
	}

	return aq
}

// WhereJSONExtract 提取JSON值进行查询
func (aq *AdvancedQueryBuilder) WhereJSONExtract(column, path, operator string, value interface{}) *AdvancedQueryBuilder {
	jsonPath := aq.buildJSONPath(column, path)
	aq.base = aq.base.WhereRaw(fmt.Sprintf("%s %s ?", jsonPath, operator), value)
	return aq
}

// 子查询支持

// WhereExists 存在性子查询
func (aq *AdvancedQueryBuilder) WhereExists(callback func(db.QueryInterface) db.QueryInterface) *AdvancedQueryBuilder {
	subQuery := aq.buildSubQuery(callback)
	aq.base = aq.base.WhereRaw(fmt.Sprintf("EXISTS (%s)", subQuery.SQL), subQuery.Bindings...)
	return aq
}

// WhereNotExists 不存在性子查询
func (aq *AdvancedQueryBuilder) WhereNotExists(callback func(db.QueryInterface) db.QueryInterface) *AdvancedQueryBuilder {
	subQuery := aq.buildSubQuery(callback)
	aq.base = aq.base.WhereRaw(fmt.Sprintf("NOT EXISTS (%s)", subQuery.SQL), subQuery.Bindings...)
	return aq
}

// WhereIn 子查询IN
func (aq *AdvancedQueryBuilder) WhereInSubQuery(column string, callback func(db.QueryInterface) db.QueryInterface) *AdvancedQueryBuilder {
	subQuery := aq.buildSubQuery(callback)
	aq.base = aq.base.WhereRaw(fmt.Sprintf("%s IN (%s)", column, subQuery.SQL), subQuery.Bindings...)
	return aq
}

// WhereNotIn 子查询NOT IN
func (aq *AdvancedQueryBuilder) WhereNotInSubQuery(column string, callback func(db.QueryInterface) db.QueryInterface) *AdvancedQueryBuilder {
	subQuery := aq.buildSubQuery(callback)
	aq.base = aq.base.WhereRaw(fmt.Sprintf("%s NOT IN (%s)", column, subQuery.SQL), subQuery.Bindings...)
	return aq
}

// 窗口函数支持

// WithRowNumber 添加行号
func (aq *AdvancedQueryBuilder) WithRowNumber(alias, partitionBy, orderBy string) *AdvancedQueryBuilder {
	windowSQL := aq.buildWindowFunction("ROW_NUMBER()", partitionBy, orderBy, alias)
	return aq.addSelectRaw(windowSQL)
}

// WithRank 添加排名
func (aq *AdvancedQueryBuilder) WithRank(alias, partitionBy, orderBy string) *AdvancedQueryBuilder {
	windowSQL := aq.buildWindowFunction("RANK()", partitionBy, orderBy, alias)
	return aq.addSelectRaw(windowSQL)
}

// WithDenseRank 添加密集排名
func (aq *AdvancedQueryBuilder) WithDenseRank(alias, partitionBy, orderBy string) *AdvancedQueryBuilder {
	windowSQL := aq.buildWindowFunction("DENSE_RANK()", partitionBy, orderBy, alias)
	return aq.addSelectRaw(windowSQL)
}

// WithLag 添加滞后值
func (aq *AdvancedQueryBuilder) WithLag(column, alias, partitionBy, orderBy string, offset int, defaultValue interface{}) *AdvancedQueryBuilder {
	var lagSQL string
	if defaultValue != nil {
		lagSQL = fmt.Sprintf("LAG(%s, %d, %v) OVER (%s)", column, offset, defaultValue, aq.buildOverClause(partitionBy, orderBy))
	} else {
		lagSQL = fmt.Sprintf("LAG(%s, %d) OVER (%s)", column, offset, aq.buildOverClause(partitionBy, orderBy))
	}
	return aq.addSelectRaw(fmt.Sprintf("%s AS %s", lagSQL, alias))
}

// 聚合查询增强

// WithCountWindow 窗口计数
func (aq *AdvancedQueryBuilder) WithCountWindow(alias, partitionBy string) *AdvancedQueryBuilder {
	windowSQL := aq.buildWindowFunction("COUNT(*)", partitionBy, "", alias)
	return aq.addSelectRaw(windowSQL)
}

// WithSumWindow 窗口求和
func (aq *AdvancedQueryBuilder) WithSumWindow(column, alias, partitionBy string) *AdvancedQueryBuilder {
	windowSQL := aq.buildWindowFunction(fmt.Sprintf("SUM(%s)", column), partitionBy, "", alias)
	return aq.addSelectRaw(windowSQL)
}

// WithAvgWindow 窗口平均值
func (aq *AdvancedQueryBuilder) WithAvgWindow(column, alias, partitionBy string) *AdvancedQueryBuilder {
	windowSQL := aq.buildWindowFunction(fmt.Sprintf("AVG(%s)", column), partitionBy, "", alias)
	return aq.addSelectRaw(windowSQL)
}

// 通用查询方法

// Get 执行查询
func (aq *AdvancedQueryBuilder) Get() ([]map[string]interface{}, error) {
	return aq.base.Get()
}

// First 获取第一条记录
func (aq *AdvancedQueryBuilder) First() (map[string]interface{}, error) {
	return aq.base.First()
}

// Count 计数
func (aq *AdvancedQueryBuilder) Count() (int64, error) {
	return aq.base.Count()
}

// Paginate 分页
func (aq *AdvancedQueryBuilder) Paginate(page, perPage int) (interface{}, error) {
	return aq.base.Paginate(page, perPage)
}

// 辅助方法

// SubQuery 子查询结构
type SubQuery struct {
	SQL      string
	Bindings []interface{}
}

// buildSubQuery 构建子查询
func (aq *AdvancedQueryBuilder) buildSubQuery(callback func(db.QueryInterface) db.QueryInterface) *SubQuery {
	// 这里需要根据具体的查询构建器实现来获取SQL和绑定参数
	// 简化实现，实际需要更复杂的逻辑
	_ = callback(aq.base)

	// 假设查询构建器有GetSQL方法（需要在实际实现中添加）
	// sql, bindings := subQueryBuilder.GetSQL()

	return &SubQuery{
		SQL:      "SELECT * FROM subquery", // 简化实现
		Bindings: []interface{}{},
	}
}

// buildJSONPath 构建JSON路径表达式
func (aq *AdvancedQueryBuilder) buildJSONPath(column, path string) string {
	driver := aq.getDriver()

	switch driver {
	case "mysql":
		if path == "" || path == "$" {
			return column
		}
		return fmt.Sprintf("JSON_EXTRACT(%s, %s)", column, aq.quoteJSONPath(path))
	case "postgres", "postgresql":
		if path == "" || path == "$" {
			return column
		}
		return fmt.Sprintf("%s%s", column, aq.buildPostgreSQLPath(path))
	case "sqlite":
		if path == "" || path == "$" {
			return column
		}
		return fmt.Sprintf("JSON_EXTRACT(%s, %s)", column, aq.quoteJSONPath(path))
	default:
		return column
	}
}

// buildPostgreSQLPath 构建PostgreSQL JSON路径
func (aq *AdvancedQueryBuilder) buildPostgreSQLPath(path string) string {
	if strings.HasPrefix(path, "$.") {
		path = path[2:] // 移除 $.
	}

	parts := strings.Split(path, ".")
	result := ""

	for _, part := range parts {
		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			// 数组索引: field[0] -> ->'field'->0
			arrayPart := strings.Split(part, "[")
			field := arrayPart[0]
			index := strings.TrimSuffix(arrayPart[1], "]")
			result += fmt.Sprintf("->'%s'->%s", field, index)
		} else {
			// 普通字段: field -> ->'field'
			result += fmt.Sprintf("->'%s'", part)
		}
	}

	return result
}

// buildPostgreSQLJSONPath 构建PostgreSQL JSON包含查询路径
func (aq *AdvancedQueryBuilder) buildPostgreSQLJSONPath(path string, value interface{}) string {
	if strings.HasPrefix(path, "$.") {
		path = path[2:]
	}

	parts := strings.Split(path, ".")
	result := make(map[string]interface{})
	current := result

	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
		} else {
			current[part] = make(map[string]interface{})
			current = current[part].(map[string]interface{})
		}
	}

	jsonBytes, _ := json.Marshal(result)
	return string(jsonBytes)
}

// quoteJSONPath 引用JSON路径
func (aq *AdvancedQueryBuilder) quoteJSONPath(path string) string {
	if !strings.HasPrefix(path, "'") && !strings.HasPrefix(path, "\"") {
		if !strings.HasPrefix(path, "$") {
			path = "$." + path
		}
		return fmt.Sprintf("'%s'", path)
	}
	return path
}

// buildWindowFunction 构建窗口函数
func (aq *AdvancedQueryBuilder) buildWindowFunction(function, partitionBy, orderBy, alias string) string {
	overClause := aq.buildOverClause(partitionBy, orderBy)
	return fmt.Sprintf("%s OVER (%s) AS %s", function, overClause, alias)
}

// buildOverClause 构建OVER子句
func (aq *AdvancedQueryBuilder) buildOverClause(partitionBy, orderBy string) string {
	var parts []string

	if partitionBy != "" {
		parts = append(parts, fmt.Sprintf("PARTITION BY %s", partitionBy))
	}

	if orderBy != "" {
		parts = append(parts, fmt.Sprintf("ORDER BY %s", orderBy))
	}

	return strings.Join(parts, " ")
}

// addSelectRaw 添加原生SELECT字段
func (aq *AdvancedQueryBuilder) addSelectRaw(expression string) *AdvancedQueryBuilder {
	// 这里需要根据实际的查询构建器实现
	// aq.base = aq.base.SelectRaw(expression)
	return aq
}

// getDriver 获取数据库驱动类型
func (aq *AdvancedQueryBuilder) getDriver() string {
	// 这里需要从连接中获取驱动类型
	// 简化实现，返回默认值
	return "mysql"
}

// 链式调用支持

// Where 基础WHERE条件
func (aq *AdvancedQueryBuilder) Where(column, operator string, value interface{}) *AdvancedQueryBuilder {
	aq.base = aq.base.Where(column, operator, value)
	return aq
}

// WhereIn IN条件
func (aq *AdvancedQueryBuilder) WhereIn(column string, values []interface{}) *AdvancedQueryBuilder {
	aq.base = aq.base.WhereIn(column, values)
	return aq
}

// OrderBy 排序
func (aq *AdvancedQueryBuilder) OrderBy(column, direction string) *AdvancedQueryBuilder {
	aq.base = aq.base.OrderBy(column, direction)
	return aq
}

// Limit 限制
func (aq *AdvancedQueryBuilder) Limit(limit int) *AdvancedQueryBuilder {
	aq.base = aq.base.Limit(limit)
	return aq
}

// Offset 偏移
func (aq *AdvancedQueryBuilder) Offset(offset int) *AdvancedQueryBuilder {
	aq.base = aq.base.Offset(offset)
	return aq
}
