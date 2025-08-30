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
	Type      string // LEFT, RIGHT, INNER, CROSS
	Table     string
	Condition string        // 条件字符串
	Raw       string        // 原生 SQL 条件
	Values    []interface{} // 绑定参数
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

	// 将结构体名转换为蛇形命名并复数化
	name := modelType.Name()
	snakeName := toSnakeCase(name)
	return pluralize(snakeName)
}

// pluralize 将单数名词转换为复数形式（简化版英文复数规则）
func pluralize(word string) string {
	if word == "" {
		return ""
	}

	// 特殊情况
	specialCases := map[string]string{
		"person":   "people",
		"child":    "children",
		"tooth":    "teeth",
		"foot":     "feet",
		"mouse":    "mice",
		"goose":    "geese",
		"user":     "users", // 常见情况
		"product":  "products",
		"category": "categories",
		"company":  "companies",
		"city":     "cities",
		"country":  "countries",
	}

	if plural, exists := specialCases[word]; exists {
		return plural
	}

	// 一般规则
	if strings.HasSuffix(word, "s") || strings.HasSuffix(word, "sh") ||
		strings.HasSuffix(word, "ch") || strings.HasSuffix(word, "x") ||
		strings.HasSuffix(word, "z") {
		return word + "es"
	}

	if strings.HasSuffix(word, "y") && len(word) > 1 {
		prevChar := word[len(word)-2]
		// 如果y前面是辅音字母，变y为ies
		if prevChar != 'a' && prevChar != 'e' && prevChar != 'i' && prevChar != 'o' && prevChar != 'u' {
			return word[:len(word)-1] + "ies"
		}
	}

	if strings.HasSuffix(word, "f") {
		return word[:len(word)-1] + "ves"
	}

	if strings.HasSuffix(word, "fe") {
		return word[:len(word)-2] + "ves"
	}

	// 默认加s
	return word + "s"
}

// toSnakeCase 转换为蛇形命名（增强版，支持连续大写字母）
func toSnakeCase(str string) string {
	if str == "" {
		return ""
	}

	var result strings.Builder
	runes := []rune(str)

	for i, r := range runes {
		// 当前字符是大写字母
		if r >= 'A' && r <= 'Z' {
			// 需要添加下划线的条件：
			// 1. 不是第一个字符
			// 2. 前一个字符是小写字母，或者
			// 3. 当前字符后面跟着小写字母（处理连续大写的情况，如 HTMLParser -> html_parser）
			if i > 0 && ((runes[i-1] >= 'a' && runes[i-1] <= 'z') || // 前一个是小写
				(i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z')) { // 后一个是小写
				result.WriteRune('_')
			}
			result.WriteRune(r - 'A' + 'a') // 转为小写
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// Select 设置选择的字段 - 支持字符串参数和数组
func (qb *QueryBuilder) Select(args ...interface{}) *QueryBuilder {
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			// 单个字符串字段
			qb.selectColumns = append(qb.selectColumns, v)
		case []string:
			// 字符串数组
			qb.selectColumns = append(qb.selectColumns, v...)
		case []interface{}:
			// interface{} 数组，需要转换为字符串
			for _, item := range v {
				if str, ok := item.(string); ok {
					qb.selectColumns = append(qb.selectColumns, str)
				}
			}
		default:
			// 尝试使用反射处理其他类型的切片
			if qb.isSliceOrArray(arg) {
				slice := qb.convertToStringSlice(arg)
				qb.selectColumns = append(qb.selectColumns, slice...)
			} else if str, ok := arg.(string); ok {
				qb.selectColumns = append(qb.selectColumns, str)
			}
		}
	}
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

// Join 内连接 - 支持多种调用方式
func (qb *QueryBuilder) Join(args ...interface{}) *QueryBuilder {
	return qb.addJoin("INNER", args...)
}

// LeftJoin 左连接 - 支持多种调用方式
func (qb *QueryBuilder) LeftJoin(args ...interface{}) *QueryBuilder {
	return qb.addJoin("LEFT", args...)
}

// RightJoin 右连接 - 支持多种调用方式
func (qb *QueryBuilder) RightJoin(args ...interface{}) *QueryBuilder {
	return qb.addJoin("RIGHT", args...)
}

// InnerJoin 内连接 - 支持多种调用方式
func (qb *QueryBuilder) InnerJoin(args ...interface{}) *QueryBuilder {
	return qb.addJoin("INNER", args...)
}

// CrossJoin 交叉连接
func (qb *QueryBuilder) CrossJoin(table string) *QueryBuilder {
	qb.joinClauses = append(qb.joinClauses, JoinClause{
		Type:  "CROSS",
		Table: table,
	})
	return qb
}

// JoinRaw 原生 JOIN 语句
func (qb *QueryBuilder) JoinRaw(joinType, table, condition string, bindings ...interface{}) *QueryBuilder {
	qb.joinClauses = append(qb.joinClauses, JoinClause{
		Type:   strings.ToUpper(joinType),
		Table:  table,
		Raw:    condition,
		Values: bindings,
	})
	return qb
}

// addJoin 内部方法 - 处理各种 JOIN 参数格式
func (qb *QueryBuilder) addJoin(joinType string, args ...interface{}) *QueryBuilder {
	switch len(args) {
	case 2:
		// Join("users", "users.id = posts.user_id") - 表名和原生条件
		if table, ok := args[0].(string); ok {
			if condition, ok := args[1].(string); ok {
				qb.joinClauses = append(qb.joinClauses, JoinClause{
					Type:      joinType,
					Table:     table,
					Condition: condition,
				})
			}
		}
	case 3:
		// Join("users", "users.id = posts.user_id", bindings) - 带参数的原生条件
		if table, ok := args[0].(string); ok {
			if condition, ok := args[1].(string); ok {
				var values []interface{}
				if bindings, ok := args[2].([]interface{}); ok {
					values = bindings
				} else {
					values = []interface{}{args[2]}
				}
				qb.joinClauses = append(qb.joinClauses, JoinClause{
					Type:   joinType,
					Table:  table,
					Raw:    condition,
					Values: values,
				})
			}
		}
	case 4:
		// Join("users", "id", "=", "posts.user_id") - 传统四参数方式
		if table, ok := args[0].(string); ok {
			if localKey, ok := args[1].(string); ok {
				if operator, ok := args[2].(string); ok {
					if foreignKey, ok := args[3].(string); ok {
						// 智能判断是否需要表前缀
						leftField := qb.addTablePrefix(localKey)
						rightField := qb.addTablePrefix(foreignKey, table)
						condition := fmt.Sprintf("%s %s %s", leftField, operator, rightField)

						qb.joinClauses = append(qb.joinClauses, JoinClause{
							Type:      joinType,
							Table:     table,
							Condition: condition,
						})
					}
				}
			}
		}
	case 5:
		// Join("users u", "u.id", "=", "posts.user_id", bindings) - 带别名和参数
		if tableAlias, ok := args[0].(string); ok {
			if localKey, ok := args[1].(string); ok {
				if operator, ok := args[2].(string); ok {
					if foreignKey, ok := args[3].(string); ok {
						// 解析表名和别名
						tableParts := strings.Fields(tableAlias)
						table := tableParts[0]

						leftField := qb.addTablePrefix(localKey)
						rightField := qb.addTablePrefix(foreignKey, table)
						condition := fmt.Sprintf("%s %s %s", leftField, operator, rightField)

						var values []interface{}
						if bindings, ok := args[4].([]interface{}); ok {
							values = bindings
						} else {
							values = []interface{}{args[4]}
						}

						qb.joinClauses = append(qb.joinClauses, JoinClause{
							Type:   joinType,
							Table:  tableAlias,
							Raw:    condition,
							Values: values,
						})
					}
				}
			}
		}
	default:
		// 不支持的参数格式，忽略
		break
	}
	return qb
}

// addTablePrefix 智能添加表前缀
func (qb *QueryBuilder) addTablePrefix(field string, defaultTable ...string) string {
	// 如果字段已经包含表前缀，直接返回
	if strings.Contains(field, ".") {
		return field
	}

	// 如果指定了默认表，使用默认表
	if len(defaultTable) > 0 && defaultTable[0] != "" {
		// 处理表别名情况
		tableParts := strings.Fields(defaultTable[0])
		if len(tableParts) > 1 {
			// 有别名，使用别名
			return tableParts[1] + "." + field
		}
		return defaultTable[0] + "." + field
	}

	// 否则使用主表名
	if qb.tableName != "" {
		return qb.tableName + "." + field
	}

	return field
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

// Having HAVING条件 - 支持多种格式
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
		// Having("COUNT(*) > ?", 5) 或 Having("status IN (?)", []string{"active", "pending"})
		if sql, ok := args[0].(string); ok {
			// 检查第二个参数是否是数组/切片
			if qb.isSliceOrArray(args[1]) {
				// 处理数组参数，如 Having("status IN (?)", []string{"active", "pending"})
				values := qb.convertToInterfaceSlice(args[1])
				if len(values) > 0 {
					// 为数组生成多个占位符
					placeholders := strings.Repeat("?,", len(values))
					placeholders = placeholders[:len(placeholders)-1] // 去掉最后的逗号

					// 替换SQL中的单个?为多个占位符
					processedSQL := strings.Replace(sql, "?", placeholders, 1)

					qb.havingConditions = append(qb.havingConditions, WhereCondition{
						Raw:    processedSQL,
						Values: values,
						Logic:  "AND",
					})
				}
			} else {
				// 普通单值参数
				qb.havingConditions = append(qb.havingConditions, WhereCondition{
					Raw:    sql,
					Values: []interface{}{args[1]},
					Logic:  "AND",
				})
			}
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
					Values: args[1:], // 剩余所有参数作为值
					Logic:  "AND",
				})
			}
		}
	}
	return qb
}

// OrHaving 添加OR HAVING条件
func (qb *QueryBuilder) OrHaving(args ...interface{}) *QueryBuilder {
	switch len(args) {
	case 1:
		// OrHaving("COUNT(*) > 5") - 纯SQL
		if sql, ok := args[0].(string); ok {
			qb.havingConditions = append(qb.havingConditions, WhereCondition{
				Raw:   sql,
				Logic: "OR",
			})
		}
	case 2:
		// OrHaving("COUNT(*) > ?", 5) 或 OrHaving("status IN (?)", []string{"active", "pending"})
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

					qb.havingConditions = append(qb.havingConditions, WhereCondition{
						Raw:    processedSQL,
						Values: values,
						Logic:  "OR",
					})
				}
			} else {
				// 普通单值参数
				qb.havingConditions = append(qb.havingConditions, WhereCondition{
					Raw:    sql,
					Values: []interface{}{args[1]},
					Logic:  "OR",
				})
			}
		}
	case 3:
		// OrHaving("column", ">", value)
		if column, ok := args[0].(string); ok {
			if operator, ok := args[1].(string); ok {
				qb.havingConditions = append(qb.havingConditions, WhereCondition{
					Column:   column,
					Operator: operator,
					Value:    args[2],
					Logic:    "OR",
				})
			}
		}
	default:
		// OrHaving("column IN (?, ?)", value1, value2) - 多参数
		if len(args) > 1 {
			if sql, ok := args[0].(string); ok {
				qb.havingConditions = append(qb.havingConditions, WhereCondition{
					Raw:    sql,
					Values: args[1:], // 剩余所有参数作为值
					Logic:  "OR",
				})
			}
		}
	}
	return qb
}

// HavingRaw 原生HAVING条件（支持多种参数格式）
func (qb *QueryBuilder) HavingRaw(args ...interface{}) *QueryBuilder {
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

// Get 执行查询并返回数据（支持访问器处理）
func (qb *QueryBuilder) Get() ([]map[string]interface{}, error) {
	// 如果启用了缓存并且不在事务中，尝试从缓存获取
	if qb.cacheEnabled && qb.transaction == nil {
		cacheKey := qb.generateCacheKey()
		if cached, err := GetDefaultCache().Get(cacheKey); err == nil {
			if result, ok := cached.([]map[string]interface{}); ok {
				return qb.applyAccessors(result), nil
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
		wrappedErr := WrapError(err, ErrCodeQueryFailed, "查询执行失败").
			WithContext("sql", sqlStr).
			WithContext("args", args).
			WithContext("table", qb.tableName)
		LogError(wrappedErr)
		return nil, wrappedErr
	}
	defer rows.Close()

	result, err := qb.scanRows(rows)
	if err != nil {
		wrappedErr := WrapError(err, ErrCodeQueryFailed, "扫描查询结果失败").
			WithContext("table", qb.tableName)
		LogError(wrappedErr)
		return nil, wrappedErr
	}

	// 如果启用了缓存，将原始结果存入缓存
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

	// 应用访问器处理
	return qb.applyAccessors(result), nil
}

// First 获取第一条记录（支持访问器处理）
func (qb *QueryBuilder) First(dest ...interface{}) (map[string]interface{}, error) {
	qb.Limit(1)
	results, err := qb.Get()
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, ErrRecordNotFound.WithContext("table", qb.tableName)
	}

	return results[0], nil
}

// Count 计算记录数量
func (qb *QueryBuilder) Count() (int64, error) {
	originalSelect := qb.selectColumns
	qb.selectColumns = []string{"COUNT(*) as count"}

	result, err := qb.First()
	if err != nil {
		// 如果是记录不存在，返回0而不是错误
		if IsNotFoundError(err) {
			qb.selectColumns = originalSelect
			return 0, nil
		}
		qb.selectColumns = originalSelect
		return 0, WrapError(err, ErrCodeQueryFailed, "执行count查询失败")
	}

	qb.selectColumns = originalSelect

	countValue := result["count"]
	if count, ok := countValue.(int64); ok {
		return count, nil
	}
	if count, ok := countValue.(int); ok {
		return int64(count), nil
	}

	return 0, NewError(ErrCodeQueryFailed, "count查询结果类型错误").
		WithContext("result_type", fmt.Sprintf("%T", countValue)).
		WithContext("result_value", countValue)
}

// Insert 插入数据
func (qb *QueryBuilder) Insert(data map[string]interface{}) (int64, error) {
	if len(data) == 0 {
		return 0, ErrInvalidParameter.WithDetails("插入数据不能为空")
	}

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
				// 检查是否为重复键错误
				if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
					return 0, WrapError(err, ErrCodeDuplicateKey, "违反唯一性约束").
						WithContext("sql", originalSQL).
						WithContext("args", args).
						WithContext("table", qb.tableName)
				}
				return 0, WrapError(err, ErrCodeQueryFailed, "PostgreSQL插入失败").
					WithContext("sql", originalSQL).
					WithContext("args", args).
					WithContext("table", qb.tableName)
			}

			if sqlResult, ok := result.(interface{ RowsAffected() (int64, error) }); ok {
				affected, err := sqlResult.RowsAffected()
				if err != nil {
					return 0, WrapError(err, ErrCodeQueryFailed, "获取影响行数失败")
				}
				return affected, nil
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
			// 检查是否为重复键错误
			if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "UNIQUE") {
				return 0, WrapError(err, ErrCodeDuplicateKey, "违反唯一性约束").
					WithContext("sql", sqlStr).
					WithContext("args", args).
					WithContext("table", qb.tableName)
			}
			return 0, WrapErrorf(err, ErrCodeQueryFailed, "%s插入失败", driverName).
				WithContext("sql", sqlStr).
				WithContext("args", args).
				WithContext("table", qb.tableName)
		}

		// 类型断言
		if sqlResult, ok := result.(interface{ LastInsertId() (int64, error) }); ok {
			id, err := sqlResult.LastInsertId()
			if err != nil {
				return 0, WrapError(err, ErrCodeQueryFailed, "获取插入ID失败")
			}
			return id, nil
		}
		return 0, NewError(ErrCodeQueryFailed, "无法获取插入ID").
			WithContext("driver", driverName).
			WithContext("table", qb.tableName)
	}
}

// Update 更新数据
func (qb *QueryBuilder) Update(data map[string]interface{}) (int64, error) {
	if len(data) == 0 {
		return 0, ErrInvalidParameter.WithDetails("更新数据不能为空")
	}

	sqlStr, args := qb.buildUpdateSQL(data)

	var result interface{}
	var err error

	if qb.transaction != nil {
		result, err = qb.transaction.Exec(sqlStr, args...)
	} else {
		result, err = qb.connection.Exec(sqlStr, args...)
	}

	if err != nil {
		// 检查是否为重复键错误
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "UNIQUE") {
			return 0, WrapError(err, ErrCodeDuplicateKey, "违反唯一性约束").
				WithContext("sql", sqlStr).
				WithContext("args", args).
				WithContext("table", qb.tableName)
		}
		return 0, WrapError(err, ErrCodeQueryFailed, "更新数据失败").
			WithContext("sql", sqlStr).
			WithContext("args", args).
			WithContext("table", qb.tableName)
	}

	if sqlResult, ok := result.(interface{ RowsAffected() (int64, error) }); ok {
		affected, err := sqlResult.RowsAffected()
		if err != nil {
			return 0, WrapError(err, ErrCodeQueryFailed, "获取影响行数失败")
		}
		return affected, nil
	}

	return 0, NewError(ErrCodeQueryFailed, "无法获取影响行数").
		WithContext("table", qb.tableName)
}

// Delete 删除数据
func (qb *QueryBuilder) Delete() (int64, error) {
	sqlStr, args := qb.buildDeleteSQL()

	var result interface{}
	var err error

	if qb.transaction != nil {
		result, err = qb.transaction.Exec(sqlStr, args...)
	} else {
		result, err = qb.connection.Exec(sqlStr, args...)
	}

	if err != nil {
		return 0, WrapError(err, ErrCodeQueryFailed, "删除数据失败").
			WithContext("sql", sqlStr).
			WithContext("args", args).
			WithContext("table", qb.tableName)
	}

	if sqlResult, ok := result.(interface{ RowsAffected() (int64, error) }); ok {
		affected, err := sqlResult.RowsAffected()
		if err != nil {
			return 0, WrapError(err, ErrCodeQueryFailed, "获取影响行数失败")
		}
		return affected, nil
	}

	return 0, NewError(ErrCodeQueryFailed, "无法获取影响行数").
		WithContext("table", qb.tableName)
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
		if join.Type == "CROSS" {
			// CROSS JOIN 不需要 ON 条件
			sql.WriteString(fmt.Sprintf(" CROSS JOIN %s", join.Table))
		} else if join.Raw != "" {
			// 使用原生 SQL 条件
			processedSQL := qb.processPlaceholders(join.Raw, argIndex)
			sql.WriteString(fmt.Sprintf(" %s JOIN %s ON %s", join.Type, join.Table, processedSQL))
			if len(join.Values) > 0 {
				args = append(args, join.Values...)
				argIndex += len(join.Values)
			}
		} else if join.Condition != "" {
			// 使用普通条件
			sql.WriteString(fmt.Sprintf(" %s JOIN %s ON %s", join.Type, join.Table, join.Condition))
		}
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

// convertToStringSlice 将各种类型的切片转换为[]string
func (qb *QueryBuilder) convertToStringSlice(value interface{}) []string {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []string:
		return v
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	default:
		// 使用反射处理其他类型的切片
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			result := make([]string, 0, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				if str, ok := rv.Index(i).Interface().(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
		return nil
	}
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

// 实现QueryInterface中缺失的方法

// From 设置查询表名
func (qb *QueryBuilder) From(table string) *QueryBuilder {
	qb.tableName = table
	return qb
}

// Model 设置关联的模型实例并自动获取表名
func (qb *QueryBuilder) Model(model interface{}) *QueryBuilder {
	qb.model = model

	// 自动从模型获取表名（如果尚未设置表名）
	if qb.tableName == "" {
		qb.tableName = getTableNameFromModel(model)
	}

	return qb
}

// GetRaw 执行查询并返回原始 map 数据（向下兼容，现在直接调用Get）
func (qb *QueryBuilder) GetRaw() ([]map[string]interface{}, error) {
	return qb.Get()
}

// FirstRaw 获取第一条记录的原始 map 数据（向下兼容，现在直接调用First）
func (qb *QueryBuilder) FirstRaw() (map[string]interface{}, error) {
	return qb.First()
}

// WhereIn WHERE IN条件
func (qb *QueryBuilder) WhereIn(field string, values []interface{}) *QueryBuilder {
	if len(values) == 0 {
		return qb
	}

	placeholders := make([]string, len(values))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	sql := fmt.Sprintf("%s IN (%s)", field, strings.Join(placeholders, ", "))
	qb.whereConditions = append(qb.whereConditions, WhereCondition{
		Raw:    sql,
		Values: values,
		Logic:  "AND",
	})
	return qb
}

// WhereNotIn WHERE NOT IN条件
func (qb *QueryBuilder) WhereNotIn(field string, values []interface{}) *QueryBuilder {
	if len(values) == 0 {
		return qb
	}

	placeholders := make([]string, len(values))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	sql := fmt.Sprintf("%s NOT IN (%s)", field, strings.Join(placeholders, ", "))
	qb.whereConditions = append(qb.whereConditions, WhereCondition{
		Raw:    sql,
		Values: values,
		Logic:  "AND",
	})
	return qb
}

// WhereBetween WHERE BETWEEN条件
func (qb *QueryBuilder) WhereBetween(field string, values []interface{}) *QueryBuilder {
	if len(values) != 2 {
		return qb
	}

	sql := fmt.Sprintf("%s BETWEEN ? AND ?", field)
	qb.whereConditions = append(qb.whereConditions, WhereCondition{
		Raw:    sql,
		Values: values,
		Logic:  "AND",
	})
	return qb
}

// WhereNotBetween WHERE NOT BETWEEN条件
func (qb *QueryBuilder) WhereNotBetween(field string, values []interface{}) *QueryBuilder {
	if len(values) != 2 {
		return qb
	}

	sql := fmt.Sprintf("%s NOT BETWEEN ? AND ?", field)
	qb.whereConditions = append(qb.whereConditions, WhereCondition{
		Raw:    sql,
		Values: values,
		Logic:  "AND",
	})
	return qb
}

// WhereNull WHERE IS NULL条件
func (qb *QueryBuilder) WhereNull(field string) *QueryBuilder {
	qb.whereConditions = append(qb.whereConditions, WhereCondition{
		Raw:   fmt.Sprintf("%s IS NULL", field),
		Logic: "AND",
	})
	return qb
}

// WhereNotNull WHERE IS NOT NULL条件
func (qb *QueryBuilder) WhereNotNull(field string) *QueryBuilder {
	qb.whereConditions = append(qb.whereConditions, WhereCondition{
		Raw:   fmt.Sprintf("%s IS NOT NULL", field),
		Logic: "AND",
	})
	return qb
}

// WhereExists WHERE EXISTS条件
func (qb *QueryBuilder) WhereExists(subQuery interface{}) *QueryBuilder {
	var sql string
	var values []interface{}

	switch sq := subQuery.(type) {
	case string:
		sql = fmt.Sprintf("EXISTS (%s)", sq)
	case *QueryBuilder:
		subSQL, subArgs := sq.buildSelectSQL()
		sql = fmt.Sprintf("EXISTS (%s)", subSQL)
		values = subArgs
	default:
		sql = fmt.Sprintf("EXISTS (%v)", subQuery)
	}

	qb.whereConditions = append(qb.whereConditions, WhereCondition{
		Raw:    sql,
		Values: values,
		Logic:  "AND",
	})
	return qb
}

// WhereNotExists WHERE NOT EXISTS条件
func (qb *QueryBuilder) WhereNotExists(subQuery interface{}) *QueryBuilder {
	var sql string
	var values []interface{}

	switch sq := subQuery.(type) {
	case string:
		sql = fmt.Sprintf("NOT EXISTS (%s)", sq)
	case *QueryBuilder:
		subSQL, subArgs := sq.buildSelectSQL()
		sql = fmt.Sprintf("NOT EXISTS (%s)", subSQL)
		values = subArgs
	default:
		sql = fmt.Sprintf("NOT EXISTS (%v)", subQuery)
	}

	qb.whereConditions = append(qb.whereConditions, WhereCondition{
		Raw:    sql,
		Values: values,
		Logic:  "AND",
	})
	return qb
}

// WhereRaw 原生WHERE条件
func (qb *QueryBuilder) WhereRaw(raw string, bindings ...interface{}) *QueryBuilder {
	qb.whereConditions = append(qb.whereConditions, WhereCondition{
		Raw:    raw,
		Values: bindings,
		Logic:  "AND",
	})
	return qb
}

// SelectRaw 原生SELECT语句
func (qb *QueryBuilder) SelectRaw(raw string, bindings ...interface{}) *QueryBuilder {
	// 这里需要处理原生SQL和绑定参数，暂时简单处理
	qb.selectColumns = append(qb.selectColumns, raw)
	return qb
}

// FieldRaw 原生字段表达式
func (qb *QueryBuilder) FieldRaw(raw string, bindings ...interface{}) *QueryBuilder {
	// FieldRaw用于添加复杂字段表达式
	// 注意：当前实现不支持参数绑定，bindings参数保留以便将来扩展
	// 建议直接在raw中包含完整的表达式，或使用SelectRaw

	// 添加原生字段表达式到选择列
	qb.selectColumns = append(qb.selectColumns, raw)
	return qb
}

// Distinct 去重查询
func (qb *QueryBuilder) Distinct() *QueryBuilder {
	// 修改第一个选择列为DISTINCT
	if len(qb.selectColumns) > 0 {
		qb.selectColumns[0] = "DISTINCT " + qb.selectColumns[0]
	} else {
		qb.selectColumns = []string{"DISTINCT *"}
	}
	return qb
}

// OrderByRaw 原生排序
func (qb *QueryBuilder) OrderByRaw(raw string, bindings ...interface{}) *QueryBuilder {
	qb.orderByColumns = append(qb.orderByColumns, OrderByClause{
		Column:    raw,
		Direction: "", // 原生SQL不需要方向
	})
	return qb
}

// OrderRand 随机排序
func (qb *QueryBuilder) OrderRand() *QueryBuilder {
	driverName := qb.getDriverName()
	var randFunc string

	switch driverName {
	case "mysql":
		randFunc = "RAND()"
	case "postgres":
		randFunc = "RANDOM()"
	case "sqlite":
		randFunc = "RANDOM()"
	default:
		randFunc = "RANDOM()"
	}

	qb.orderByColumns = append(qb.orderByColumns, OrderByClause{
		Column:    randFunc,
		Direction: "",
	})
	return qb
}

// OrderField 字段排序
func (qb *QueryBuilder) OrderField(field string, values []interface{}, direction string) *QueryBuilder {
	// 生成FIELD()或CASE WHEN排序
	driverName := qb.getDriverName()

	if driverName == "mysql" {
		// MySQL使用FIELD()函数，直接生成完整SQL
		valueParts := make([]string, len(values)+1)
		valueParts[0] = field
		for i, value := range values {
			// 将值直接嵌入SQL（安全性：这里的值通常是预定义的枚举值）
			valueParts[i+1] = fmt.Sprintf("'%v'", value)
		}
		orderExpr := fmt.Sprintf("FIELD(%s)", strings.Join(valueParts, ", "))

		qb.orderByColumns = append(qb.orderByColumns, OrderByClause{
			Column:    orderExpr,
			Direction: direction,
		})
	} else {
		// 其他数据库使用CASE WHEN，直接生成完整SQL
		var caseSQL strings.Builder
		caseSQL.WriteString("CASE ")
		for i, value := range values {
			caseSQL.WriteString(fmt.Sprintf("WHEN %s = '%v' THEN %d ", field, value, i))
		}
		caseSQL.WriteString("ELSE 999 END")

		qb.orderByColumns = append(qb.orderByColumns, OrderByClause{
			Column:    caseSQL.String(),
			Direction: direction,
		})
	}

	return qb
}

// Page 分页设置
func (qb *QueryBuilder) Page(page, pageSize int) *QueryBuilder {
	qb.limitCount = pageSize
	qb.offsetCount = (page - 1) * pageSize
	return qb
}

// WithContext 设置上下文
func (qb *QueryBuilder) WithContext(ctx context.Context) *QueryBuilder {
	qb.ctx = ctx
	return qb
}

// WithTimeout 设置超时
func (qb *QueryBuilder) WithTimeout(timeout time.Duration) *QueryBuilder {
	ctx, cancel := context.WithTimeout(qb.ctx, timeout)
	// 注意：这里我们保存cancel函数，但实际使用时需要在适当时机调用
	// 对于查询构建器，通常在查询执行完成后调用cancel
	_ = cancel // 暂时忽略warning，实际项目中需要合理管理
	qb.ctx = ctx
	return qb
}

// Find 根据条件查找（支持访问器处理）
func (qb *QueryBuilder) Find(args ...interface{}) (map[string]interface{}, error) {
	// 支持 Find(id) 或 Find(dest) 模式
	if len(args) == 1 {
		// 假设是根据主键查找
		qb = qb.Where("id", "=", args[0])
	}

	return qb.First()
}

// Last 获取最后一条记录（支持访问器处理）
func (qb *QueryBuilder) Last() (map[string]interface{}, error) {
	// 反转排序以获取最后一条记录
	if len(qb.orderByColumns) == 0 {
		// 如果没有排序，默认按id降序
		qb.OrderBy("id", "DESC")
	} else {
		// 反转现有排序
		for i := range qb.orderByColumns {
			if qb.orderByColumns[i].Direction == "ASC" {
				qb.orderByColumns[i].Direction = "DESC"
			} else {
				qb.orderByColumns[i].Direction = "ASC"
			}
		}
	}

	return qb.First()
}

// Exists 检查记录是否存在
func (qb *QueryBuilder) Exists() (bool, error) {
	count, err := qb.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// InsertBatch 批量插入数据
func (qb *QueryBuilder) InsertBatch(data []map[string]interface{}) (int64, error) {
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
		qb.tableName, strings.Join(columns, ", ")))

	// 构建VALUES部分
	var args []interface{}
	valueParts := make([]string, len(data))

	for i, row := range data {
		placeholders := make([]string, len(columns))
		for j, column := range columns {
			if value, exists := row[column]; exists {
				args = append(args, value)
			} else {
				args = append(args, nil)
			}

			// 根据数据库类型生成占位符
			driverName := qb.getDriverName()
			if driverName == "postgres" {
				placeholders[j] = fmt.Sprintf("$%d", len(args))
			} else {
				placeholders[j] = "?"
			}
		}
		valueParts[i] = fmt.Sprintf("(%s)", strings.Join(placeholders, ", "))
	}

	sql.WriteString(strings.Join(valueParts, ", "))

	// 执行插入
	var result interface{}
	var err error

	if qb.transaction != nil {
		result, err = qb.transaction.Exec(sql.String(), args...)
	} else {
		result, err = qb.connection.Exec(sql.String(), args...)
	}

	if err != nil {
		return 0, err
	}

	// 返回影响的行数
	if sqlResult, ok := result.(interface{ RowsAffected() (int64, error) }); ok {
		return sqlResult.RowsAffected()
	}

	return int64(len(data)), nil
}

// Exp 高级表达式
func (qb *QueryBuilder) Exp(field string, expression string, bindings ...interface{}) *QueryBuilder {
	// 将表达式作为原生SQL添加到WHERE条件中
	sql := fmt.Sprintf("%s %s", field, expression)
	qb.whereConditions = append(qb.whereConditions, WhereCondition{
		Raw:    sql,
		Values: bindings,
		Logic:  "AND",
	})
	return qb
}

// ToSQL 构建SQL语句
func (qb *QueryBuilder) ToSQL() (string, []interface{}, error) {
	sql, args := qb.buildSelectSQL()
	return sql, args, nil
}

// Clone 克隆查询构建器
func (qb *QueryBuilder) Clone() *QueryBuilder {
	newBuilder := &QueryBuilder{
		connection:       qb.connection,
		tableName:        qb.tableName,
		model:            qb.model,
		selectColumns:    make([]string, len(qb.selectColumns)),
		whereConditions:  make([]WhereCondition, len(qb.whereConditions)),
		joinClauses:      make([]JoinClause, len(qb.joinClauses)),
		orderByColumns:   make([]OrderByClause, len(qb.orderByColumns)),
		groupByColumns:   make([]string, len(qb.groupByColumns)),
		havingConditions: make([]WhereCondition, len(qb.havingConditions)),
		limitCount:       qb.limitCount,
		offsetCount:      qb.offsetCount,
		transaction:      qb.transaction,
		cacheEnabled:     qb.cacheEnabled,
		cacheTTL:         qb.cacheTTL,
		cacheTags:        make([]string, len(qb.cacheTags)),
		cacheKey:         qb.cacheKey,
		ctx:              qb.ctx,
	}

	// 复制切片内容
	copy(newBuilder.selectColumns, qb.selectColumns)
	copy(newBuilder.whereConditions, qb.whereConditions)
	copy(newBuilder.joinClauses, qb.joinClauses)
	copy(newBuilder.orderByColumns, qb.orderByColumns)
	copy(newBuilder.groupByColumns, qb.groupByColumns)
	copy(newBuilder.havingConditions, qb.havingConditions)
	copy(newBuilder.cacheTags, qb.cacheTags)

	return newBuilder
}

// applyAccessors 应用访问器处理数据
func (qb *QueryBuilder) applyAccessors(data []map[string]interface{}) []map[string]interface{} {
	// 如果没有绑定模型，直接返回原始数据
	if qb.model == nil {
		return data
	}

	// 创建访问器处理器
	processor := NewAccessorProcessor(qb.model)

	// 应用访问器处理
	return processor.ProcessDataSlice(data)
}

// WithModel 绑定模型（用于模型支持）
func (qb *QueryBuilder) WithModel(model interface{}) *QueryBuilder {
	qb.model = model
	return qb
}

// InsertModel 插入模型实例
func (qb *QueryBuilder) InsertModel(model interface{}) (int64, error) {
	// 这里需要将模型转换为map[string]interface{}
	// 暂时返回错误，需要反射处理
	return 0, fmt.Errorf("InsertModel not implemented yet")
}

// UpdateModel 更新模型实例
func (qb *QueryBuilder) UpdateModel(model interface{}) (int64, error) {
	// 这里需要将模型转换为map[string]interface{}
	// 暂时返回错误，需要反射处理
	return 0, fmt.Errorf("UpdateModel not implemented yet")
}

// FindModel 查找并填充模型
func (qb *QueryBuilder) FindModel(id interface{}, model interface{}) error {
	// 这里需要将查询结果填充到模型中
	// 暂时返回错误，需要反射处理
	return fmt.Errorf("FindModel not implemented yet")
}
