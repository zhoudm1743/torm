package db

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// LoadModel 将map数据填充到指针指向的结构体
func LoadModel(dest interface{}, result map[string]interface{}) error {
	if result == nil {
		return fmt.Errorf("no data to load")
	}

	// 使用反射填充目标模型
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("destination must be a pointer")
	}

	destValue = destValue.Elem()
	destType := destValue.Type()

	// 如果目标有BaseModel字段，通过SetAttribute方法设置属性
	if baseModelField := destValue.FieldByName("BaseModel"); baseModelField.IsValid() {
		// 尝试调用SetAttribute方法设置每个属性
		setAttrMethod := baseModelField.Addr().MethodByName("SetAttribute")
		if setAttrMethod.IsValid() {
			for key, value := range result {
				// 使用defer + recover避免因为未初始化而崩溃
				func() {
					defer func() {
						if r := recover(); r != nil {
							// 忽略错误，继续处理其他字段
						}
					}()
					setAttrMethod.Call([]reflect.Value{
						reflect.ValueOf(key),
						reflect.ValueOf(value),
					})
				}()
			}
		}
	}

	// 填充结构体字段
	for i := 0; i < destType.NumField(); i++ {
		field := destType.Field(i)

		// 跳过BaseModel字段
		if field.Name == "BaseModel" {
			continue
		}

		dbTag := field.Tag.Get("db")
		jsonTag := field.Tag.Get("json")

		var fieldName string
		if dbTag != "" && dbTag != "-" {
			fieldName = dbTag
		} else if jsonTag != "" && jsonTag != "-" {
			fieldName = jsonTag
		} else {
			fieldName = strings.ToLower(field.Name)
		}

		if value, exists := result[fieldName]; exists && destValue.Field(i).CanSet() {
			fieldValue := destValue.Field(i)
			if fieldValue.Kind() == reflect.Ptr {
				if value != nil {
					// 为指针字段分配内存
					newValue := reflect.New(fieldValue.Type().Elem())
					if newValue.Elem().Type() == reflect.TypeOf(value) {
						newValue.Elem().Set(reflect.ValueOf(value))
						fieldValue.Set(newValue)
					}
				}
			} else {
				if value != nil && reflect.TypeOf(value).AssignableTo(fieldValue.Type()) {
					fieldValue.Set(reflect.ValueOf(value))
				}
			}
		}
	}

	return nil
}

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

	// 模型支持
	boundModel    interface{} // 绑定的模型实例
	modelMetadata interface{} // 模型元数据（从model包导入会造成循环依赖，使用interface{}）
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

// Where 添加WHERE条件 - 支持多种调用方式
// 1. Where(field, operator, value) - 传统方式
// 2. Where(condition, args...) - TORM风格参数化查询
func (q *QueryBuilder) Where(args ...interface{}) QueryInterface {
	newQuery := q.clone()

	if len(args) == 0 {
		return newQuery
	}

	// 支持传统的三参数方式: Where(field, operator, value)
	if len(args) == 3 {
		field, ok1 := args[0].(string)
		operator, ok2 := args[1].(string)
		if ok1 && ok2 && isValidSQLOperator(operator) {
			newQuery.wheres = append(newQuery.wheres, WhereClause{
				Type:     "and",
				Field:    field,
				Operator: operator,
				Value:    args[2],
			})
			return newQuery
		}
	}

	// 支持TORM风格的参数化查询: Where("name = ?", "张三") 或 Where("name = ? AND age >= ?", "张三", 22)
	if len(args) >= 1 {
		condition, ok := args[0].(string)
		if ok {
			bindings := args[1:]
			newQuery.wheres = append(newQuery.wheres, WhereClause{
				Type:        "and",
				Raw:         condition,
				RawBindings: bindings,
			})
			return newQuery
		}
	}

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
func (q *QueryBuilder) WhereBetween(field string, values []interface{}) QueryInterface {
	if len(values) != 2 {
		return q
	}
	newQuery := q.clone()
	newQuery.wheres = append(newQuery.wheres, WhereClause{
		Type:     "and",
		Field:    field,
		Operator: "BETWEEN",
		Value:    values,
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

// OrWhere 添加OR WHERE条件 - 支持多种调用方式
// 1. OrWhere(field, operator, value) - 传统方式
// 2. OrWhere(condition, args...) - TORM风格参数化查询
func (q *QueryBuilder) OrWhere(args ...interface{}) QueryInterface {
	newQuery := q.clone()

	if len(args) == 0 {
		return newQuery
	}

	// 支持传统的三参数方式: OrWhere(field, operator, value)
	if len(args) == 3 {
		field, ok1 := args[0].(string)
		operator, ok2 := args[1].(string)
		if ok1 && ok2 {
			newQuery.wheres = append(newQuery.wheres, WhereClause{
				Type:     "or",
				Field:    field,
				Operator: operator,
				Value:    args[2],
			})
			return newQuery
		}
	}

	// 支持TORM风格的参数化查询
	if len(args) >= 1 {
		condition, ok := args[0].(string)
		if ok {
			bindings := args[1:]
			newQuery.wheres = append(newQuery.wheres, WhereClause{
				Type:        "or",
				Raw:         condition,
				RawBindings: bindings,
			})
			return newQuery
		}
	}

	return newQuery
}

// WhereNotBetween 添加NOT BETWEEN查询条件
func (q *QueryBuilder) WhereNotBetween(field string, values []interface{}) QueryInterface {
	if len(values) != 2 {
		return q
	}
	newQuery := q.clone()
	newQuery.wheres = append(newQuery.wheres, WhereClause{
		Type:     "AND",
		Field:    field,
		Operator: "NOT BETWEEN",
		Value:    values,
	})
	return newQuery
}

// WhereExists 添加EXISTS子查询条件
func (q *QueryBuilder) WhereExists(subQuery interface{}) QueryInterface {
	newQuery := q.clone()

	var sql string
	var bindings []interface{}

	// 处理不同类型的子查询
	switch sq := subQuery.(type) {
	case string:
		sql = sq
	case QueryInterface:
		var err error
		sql, bindings, err = sq.ToSQL()
		if err != nil {
			// 如果构建SQL失败，返回原查询
			return q
		}
	default:
		return q
	}

	newQuery.wheres = append(newQuery.wheres, WhereClause{
		Type:        "AND",
		Field:       "",
		Operator:    "EXISTS",
		Value:       nil,
		Raw:         fmt.Sprintf("EXISTS (%s)", sql),
		RawBindings: bindings,
	})
	return newQuery
}

// WhereNotExists 添加NOT EXISTS子查询条件
func (q *QueryBuilder) WhereNotExists(subQuery interface{}) QueryInterface {
	newQuery := q.clone()

	var sql string
	var bindings []interface{}

	switch sq := subQuery.(type) {
	case string:
		sql = sq
	case QueryInterface:
		var err error
		sql, bindings, err = sq.ToSQL()
		if err != nil {
			return q
		}
	default:
		return q
	}

	newQuery.wheres = append(newQuery.wheres, WhereClause{
		Type:        "AND",
		Field:       "",
		Operator:    "NOT EXISTS",
		Value:       nil,
		Raw:         fmt.Sprintf("NOT EXISTS (%s)", sql),
		RawBindings: bindings,
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

// OrderRand 随机排序
func (q *QueryBuilder) OrderRand() QueryInterface {
	newQuery := q.clone()
	// 根据不同数据库使用不同的随机函数
	driver := q.getDriverName()
	var randomFunc string
	switch driver {
	case "mysql":
		randomFunc = "RAND()"
	case "postgres", "postgresql":
		randomFunc = "RANDOM()"
	case "sqlite":
		randomFunc = "RANDOM()"
	default:
		randomFunc = "RAND()"
	}

	newQuery.orders = append(newQuery.orders, OrderClause{
		Raw: randomFunc,
	})
	return newQuery
}

// OrderField 按字段值排序
// 例如: OrderField("status", []interface{}{"active", "inactive", "pending"}, "asc")
func (q *QueryBuilder) OrderField(field string, values []interface{}, direction string) QueryInterface {
	if len(values) == 0 {
		return q
	}

	newQuery := q.clone()

	// 构建CASE WHEN语句
	var caseWhen strings.Builder
	caseWhen.WriteString("CASE ")
	bindings := make([]interface{}, 0, len(values))

	for i, value := range values {
		caseWhen.WriteString(fmt.Sprintf("WHEN %s = ? THEN %d ", field, i))
		bindings = append(bindings, value)
	}
	caseWhen.WriteString("ELSE 999 END")

	if direction != "" {
		caseWhen.WriteString(" " + strings.ToUpper(direction))
	}

	newQuery.orders = append(newQuery.orders, OrderClause{
		Raw:         caseWhen.String(),
		RawBindings: bindings,
	})
	return newQuery
}

// FieldRaw 添加原生字段表达式
func (q *QueryBuilder) FieldRaw(raw string, bindings ...interface{}) QueryInterface {
	newQuery := q.clone()
	newQuery.fields = append(newQuery.fields, raw)
	// 如果有绑定参数，需要存储起来用于最终SQL构建
	// 这里简化处理，实际可能需要更复杂的参数管理
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

	return q.scanRows(rows)
}

// First 执行查询并返回第一条记录
// 如果传入指针，也会填充到指针指向的对象
func (q *QueryBuilder) First(dest ...interface{}) (map[string]interface{}, error) {
	results, err := q.Limit(1).Get()
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no records found")
	}

	result := results[0]

	// 如果传入了指针，填充到指针指向的对象
	if len(dest) > 0 && dest[0] != nil {
		err = LoadModel(dest[0], result)
		if err != nil {
			return result, fmt.Errorf("failed to load model: %w", err)
		}
	}

	return result, nil
}

// Find 根据ID查找记录，或执行Where条件查找多条记录
// 用法1: Find(1, &user) - 根据ID查找
// 用法2: Find(&users) - 查找所有符合条件的记录（需要之前调用Where）
func (q *QueryBuilder) Find(args ...interface{}) (map[string]interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("find requires at least one argument")
	}

	// 如果第一个参数不是指针，说明是ID查找模式
	firstArgValue := reflect.ValueOf(args[0])
	if firstArgValue.Kind() != reflect.Ptr {
		// ID查找模式: Find(id, dest...)
		id := args[0]
		dest := args[1:]
		return q.Where("id", "=", id).First(dest...)
	}

	// 如果第一个参数是指针，说明是查找多条记录模式
	dest := args[0]
	results, err := q.Get()
	if err != nil {
		return nil, err
	}

	// 对于切片类型，填充多条记录
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() == reflect.Ptr && destValue.Elem().Kind() == reflect.Slice {
		sliceValue := destValue.Elem()
		sliceType := sliceValue.Type()
		elemType := sliceType.Elem()

		// 创建新的切片
		newSlice := reflect.MakeSlice(sliceType, 0, len(results))

		for _, result := range results {
			// 为切片元素类型创建新实例
			var elem reflect.Value
			if elemType.Kind() == reflect.Ptr {
				elem = reflect.New(elemType.Elem())
			} else {
				elem = reflect.New(elemType)
			}

			// 填充数据
			err = LoadModel(elem.Interface(), result)
			if err != nil {
				continue // 忽略单个记录的错误
			}

			// 添加到切片
			if elemType.Kind() == reflect.Ptr {
				newSlice = reflect.Append(newSlice, elem)
			} else {
				newSlice = reflect.Append(newSlice, elem.Elem())
			}
		}

		// 设置切片
		sliceValue.Set(newSlice)

		// 返回第一条记录的map（如果存在）
		if len(results) > 0 {
			return results[0], nil
		}
		return nil, nil
	}

	// 对于单个记录，使用First方法
	return q.First(dest)
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
		connection:    q.connection,
		table:         q.table,
		fields:        make([]string, len(q.fields)),
		wheres:        make([]WhereClause, len(q.wheres)),
		joins:         make([]JoinClause, len(q.joins)),
		orders:        make([]OrderClause, len(q.orders)),
		groups:        make([]string, len(q.groups)),
		havings:       make([]WhereClause, len(q.havings)),
		limitNum:      q.limitNum,
		offsetNum:     q.offsetNum,
		distinct:      q.distinct,
		ctx:           q.ctx, // 克隆context
		boundModel:    q.boundModel,
		modelMetadata: q.modelMetadata,
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
				case "BETWEEN", "NOT BETWEEN":
					values := where.Value.([]interface{})
					sql.WriteString(" ? AND ?")
					bindings = append(bindings, values...)
				case "IS NULL", "IS NOT NULL":
					// 不需要添加占位符
				case "EXISTS", "NOT EXISTS":
					// EXISTS和NOT EXISTS不需要占位符，Raw已经包含完整条件
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

			// 获取数据库类型名称（用于调试）
			dbTypeName := columnTypes[i].DatabaseTypeName()

			// 根据数据库类型转换值
			switch dbTypeName {
			case "VARCHAR", "TEXT", "CHAR", "MEDIUMTEXT", "LONGTEXT", "TINYTEXT",
				"ENUM", "SET", "JSON", "BINARY", "VARBINARY":
				// 文本类型：强制转换[]byte为string
				if b, ok := val.([]byte); ok {
					row[col] = string(b)
				} else {
					row[col] = val
				}
			case "INT", "INTEGER", "BIGINT", "TINYINT", "SMALLINT", "MEDIUMINT",
				"UNSIGNED INT", "UNSIGNED BIGINT", "UNSIGNED TINYINT", "UNSIGNED SMALLINT", "UNSIGNED MEDIUMINT":
				// 整数类型：保持原样
				row[col] = val
			case "FLOAT", "DOUBLE", "DECIMAL", "NUMERIC":
				// 浮点类型：保持原样
				row[col] = val
			case "BOOLEAN", "BOOL", "BIT":
				// 布尔类型：保持原样
				row[col] = val
			case "TIMESTAMP", "DATETIME", "DATE", "TIME", "YEAR":
				// 日期时间类型：可能需要转换[]byte
				if b, ok := val.([]byte); ok {
					row[col] = string(b)
				} else {
					row[col] = val
				}
			default:
				// 默认情况下，如果是[]byte类型，尝试转换为字符串
				// 这是一个安全的后备方案，适用于未明确处理的类型
				if b, ok := val.([]byte); ok {
					row[col] = string(b)
				} else {
					row[col] = val
				}
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

// Paginate 实现分页查询
func (q *QueryBuilder) Paginate(page, perPage int) (interface{}, error) {
	// 导入分页器包并创建分页查询
	// 由于循环导入问题，这里使用简化实现
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 15
	}

	// 计算总数
	total, err := q.Count()
	if err != nil {
		return nil, err
	}

	// 计算偏移量
	offset := (page - 1) * perPage

	// 执行查询
	data, err := q.Limit(perPage).Offset(offset).Get()
	if err != nil {
		return nil, err
	}

	// 转换数据格式
	items := make([]interface{}, len(data))
	for i, item := range data {
		items[i] = item
	}

	// 创建简化的分页结果
	lastPage := int((total + int64(perPage) - 1) / int64(perPage))
	hasMore := page < lastPage

	result := map[string]interface{}{
		"data":         items,
		"total":        total,
		"per_page":     perPage,
		"current_page": page,
		"last_page":    lastPage,
		"has_more":     hasMore,
		"has_prev":     page > 1,
	}

	return result, nil
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

// getDriverName 获取数据库驱动名称
func (q *QueryBuilder) getDriverName() string {
	if q.connection != nil {
		if conn, ok := q.connection.(interface{ GetDriver() string }); ok {
			return conn.GetDriver()
		}
	}
	return "mysql" // 默认值
}

// Exp 设置字段表达式 (用于UPDATE等操作)
func (q *QueryBuilder) Exp(field string, expression string, bindings ...interface{}) QueryInterface {
	// 这个方法主要用于UPDATE操作中的字段表达式设置
	// 由于需要扩展数据结构，这里先提供基础实现
	newQuery := q.clone()
	// 实际实现需要在UPDATE方法中处理表达式
	return newQuery
}

// isValidSQLOperator 检查是否为有效的SQL操作符
func isValidSQLOperator(operator string) bool {
	validOperators := map[string]bool{
		"=":           true,
		"!=":          true,
		"<>":          true,
		"<":           true,
		"<=":          true,
		">":           true,
		">=":          true,
		"LIKE":        true,
		"NOT LIKE":    true,
		"IN":          true,
		"NOT IN":      true,
		"BETWEEN":     true,
		"NOT BETWEEN": true,
		"IS":          true,
		"IS NOT":      true,
		"IS NULL":     true,
		"IS NOT NULL": true,
		"REGEXP":      true,
		"NOT REGEXP":  true,
		"RLIKE":       true,
	}

	return validOperators[strings.ToUpper(operator)]
}

// ===== 模型支持方法 =====

// WithModel 绑定模型，启用模型特性
func (q *QueryBuilder) WithModel(model interface{}) QueryInterface {
	newQuery := q.clone()
	newQuery.boundModel = model
	newQuery.modelMetadata = model

	// 自动应用软删除过滤
	if model != nil {
		deletedAtField := q.getModelDeletedAtField(model)
		if deletedAtField != "" {
			// 自动添加软删除过滤条件
			newQuery = newQuery.WhereNull(deletedAtField).(*QueryBuilder)
		}
	}

	return newQuery
}

// InsertModel 插入模型实例
func (q *QueryBuilder) InsertModel(model interface{}) (int64, error) {
	if model == nil {
		return 0, fmt.Errorf("model cannot be nil")
	}

	// 将模型转换为map数据
	data, err := q.modelToMap(model)
	if err != nil {
		return 0, err
	}

	// 应用模型的时间戳规则
	data = q.applyTimestamps(data, model, "insert")

	return q.Insert(data)
}

// UpdateModel 更新模型实例
func (q *QueryBuilder) UpdateModel(model interface{}) (int64, error) {
	if model == nil {
		return 0, fmt.Errorf("model cannot be nil")
	}

	// 将模型转换为map数据
	data, err := q.modelToMap(model)
	if err != nil {
		return 0, err
	}

	// 应用模型的时间戳规则
	data = q.applyTimestamps(data, model, "update")

	return q.Update(data)
}

// FindModel 查找并填充模型
func (q *QueryBuilder) FindModel(id interface{}, model interface{}) error {
	if model == nil {
		return fmt.Errorf("model cannot be nil")
	}

	// 获取主键字段名
	pkField := q.getModelPrimaryKey(model)
	if pkField == "" {
		pkField = "id" // 默认使用id
	}

	// 查询数据
	result, err := q.Where(pkField, "=", id).First()
	if err != nil {
		return err
	}

	// 填充模型
	return LoadModel(model, result)
}

// ===== 辅助方法 =====

// modelToMap 将模型实例转换为map数据
func (q *QueryBuilder) modelToMap(model interface{}) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	modelValue := reflect.ValueOf(model)
	if modelValue.Kind() == reflect.Ptr {
		modelValue = modelValue.Elem()
	}

	if modelValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("model must be a struct or pointer to struct")
	}

	modelType := modelValue.Type()

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		fieldValue := modelValue.Field(i)

		// 跳过BaseModel字段
		if field.Name == "BaseModel" {
			continue
		}

		// 获取db标签作为字段名
		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		// 解析db标签，提取字段名
		fieldName := strings.Split(dbTag, ";")[0]

		// 获取字段值，跳过零值
		if fieldValue.CanInterface() && !fieldValue.IsZero() {
			data[fieldName] = fieldValue.Interface()
		}
	}

	return data, nil
}

// applyTimestamps 应用时间戳规则
func (q *QueryBuilder) applyTimestamps(data map[string]interface{}, model interface{}, operation string) map[string]interface{} {
	if model == nil {
		return data
	}

	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	now := time.Now()

	// 检查字段标签，查找autoCreateTime和autoUpdateTime
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		dbTag := field.Tag.Get("db")

		if dbTag == "" || dbTag == "-" {
			continue
		}

		parts := strings.Split(dbTag, ";")
		fieldName := parts[0]

		// 检查标签选项
		for _, part := range parts[1:] {
			switch part {
			case "autoCreateTime":
				if operation == "insert" {
					data[fieldName] = now
				}
			case "autoUpdateTime":
				if operation == "insert" || operation == "update" {
					data[fieldName] = now
				}
			}
		}
	}

	return data
}

// getModelPrimaryKey 获取模型的主键字段名
func (q *QueryBuilder) getModelPrimaryKey(model interface{}) string {
	if model == nil {
		return ""
	}

	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 查找带有pk标签的字段
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		if field.Tag.Get("pk") != "" {
			dbTag := field.Tag.Get("db")
			if dbTag != "" && dbTag != "-" {
				return strings.Split(dbTag, ";")[0]
			}
		}
	}

	return "id" // 默认返回id
}

// getModelDeletedAtField 获取模型的软删除字段名
func (q *QueryBuilder) getModelDeletedAtField(model interface{}) string {
	if model == nil {
		return ""
	}

	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 查找DeletedTime类型的字段
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)

		// 检查字段类型名是否包含DeletedTime
		fieldTypeName := field.Type.String()
		if strings.Contains(fieldTypeName, "DeletedTime") {
			dbTag := field.Tag.Get("db")
			if dbTag != "" && dbTag != "-" {
				return strings.Split(dbTag, ";")[0]
			}
			// 如果没有db标签，使用字段名的小写形式
			return strings.ToLower(field.Name)
		}
	}

	return "" // 没有软删除字段
}
