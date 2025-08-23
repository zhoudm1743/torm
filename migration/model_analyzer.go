package migration

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// ModelAnalyzer 模型分析器
type ModelAnalyzer struct{}

// NewModelAnalyzer 创建模型分析器
func NewModelAnalyzer() *ModelAnalyzer {
	return &ModelAnalyzer{}
}

// AnalyzeModel 分析模型结构体，提取列信息
func (ma *ModelAnalyzer) AnalyzeModel(modelType reflect.Type) ([]ModelColumn, error) {
	var columns []ModelColumn

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)

		// 跳过嵌入的BaseModel字段
		if field.Type.Name() == "BaseModel" {
			continue
		}

		// 跳过没有数据库标签的字段
		if !ma.HasDBTag(field) {
			continue
		}

		column, err := ma.analyzeField(field)
		if err != nil {
			return nil, err
		}

		if column != nil {
			columns = append(columns, *column)
		}
	}

	return columns, nil
}

// HasDBTag 检查字段是否有数据库相关的标签
func (ma *ModelAnalyzer) HasDBTag(field reflect.StructField) bool {
	// 检查是否有torm标签或传统的数据库标签
	if field.Tag.Get("torm") != "" {
		return true
	}

	// 检查传统标签
	tags := []string{"db", "primaryKey", "autoIncrement", "size", "unique", "default", "comment"}
	for _, tag := range tags {
		if field.Tag.Get(tag) != "" {
			return true
		}
	}

	return false
}

// analyzeField 分析单个字段
func (ma *ModelAnalyzer) analyzeField(field reflect.StructField) (*ModelColumn, error) {
	column := &ModelColumn{
		Name: ma.GetColumnName(field),
	}

	// 分析Go类型并映射到数据库类型
	column.Type = ma.mapGoTypeToColumnType(field.Type)

	// 解析标签
	if err := ma.parseFieldTags(field, column); err != nil {
		return nil, err
	}

	return column, nil
}

// GetColumnName 获取列名
func (ma *ModelAnalyzer) GetColumnName(field reflect.StructField) string {
	// 优先使用torm标签中的列名
	if tormTag := field.Tag.Get("torm"); tormTag != "" {
		if name := ma.ExtractColumnNameFromTorm(tormTag); name != "" {
			return name
		}
	}

	// 使用db标签
	if dbTag := field.Tag.Get("db"); dbTag != "" {
		return dbTag
	}

	// 使用字段名的蛇形命名
	return ma.toSnakeCase(field.Name)
}

// ExtractColumnNameFromTorm 从torm标签中提取列名
func (ma *ModelAnalyzer) ExtractColumnNameFromTorm(tormTag string) string {
	parts := strings.Split(tormTag, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "column:") {
			return strings.TrimPrefix(part, "column:")
		}
	}
	return ""
}

// parseFieldTags 解析字段标签
func (ma *ModelAnalyzer) parseFieldTags(field reflect.StructField, column *ModelColumn) error {
	// 优先解析torm标签
	if tormTag := field.Tag.Get("torm"); tormTag != "" {
		return ma.parseTormTag(tormTag, column)
	}

	// 解析传统标签
	return ma.parseLegacyTags(field, column)
}

// parseTormTag 解析torm标签
func (ma *ModelAnalyzer) parseTormTag(tormTag string, column *ModelColumn) error {
	parts := strings.Split(tormTag, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, ":") {
			// 键值对形式: key:value
			if err := ma.ParseTormKeyValue(part, column); err != nil {
				return err
			}
		} else {
			// 标志形式: primary_key, unique等
			ma.ParseTormFlag(part, column)
		}
	}

	return nil
}

// ParseTormKeyValue 解析torm标签的键值对
func (ma *ModelAnalyzer) ParseTormKeyValue(part string, column *ModelColumn) error {
	kv := strings.SplitN(part, ":", 2)
	if len(kv) != 2 {
		return nil
	}

	key := strings.ToLower(strings.TrimSpace(kv[0]))
	value := strings.TrimSpace(kv[1])

	switch key {
	case "type":
		ma.SetColumnType(value, column)
	case "size":
		if size, err := strconv.Atoi(value); err == nil {
			column.Length = size
		}
	case "precision":
		if precision, err := strconv.Atoi(value); err == nil {
			column.Precision = precision
		}
	case "scale":
		if scale, err := strconv.Atoi(value); err == nil {
			column.Scale = scale
		}
	case "default":
		defaultVal := ma.ParseDefaultValue(value)
		column.Default = &defaultVal
	case "comment":
		column.Comment = value
	case "column":
		// 自定义列名支持
		column.Name = value
	case "length", "len":
		// 长度的别名支持
		if size, err := strconv.Atoi(value); err == nil {
			column.Length = size
		}
	case "width":
		// 宽度的别名支持(与长度相同)
		if size, err := strconv.Atoi(value); err == nil {
			column.Length = size
		}
	case "prec":
		// 精度的别名支持
		if precision, err := strconv.Atoi(value); err == nil {
			column.Precision = precision
		}
	case "digits":
		// 小数位的别名支持
		if scale, err := strconv.Atoi(value); err == nil {
			column.Scale = scale
		}
	case "charset":
		// 字符集支持 (仅MySQL相关，可以存储在扩展属性中)
		// 这里先忽略，因为ModelColumn结构体没有字符集字段
	case "collation", "collate":
		// 排序规则支持 (仅MySQL相关)
		// 这里先忽略，因为ModelColumn结构体没有排序规则字段
	case "engine":
		// 存储引擎 (仅MySQL相关)
		// 这里先忽略，因为这是表级别的属性
	case "auto_update", "on_update":
		// ON UPDATE 子句
		if strings.ToLower(value) == "current_timestamp" {
			// 这是 auto_update_time 的另一种写法
			column.NotNull = true
			if column.Default == nil {
				defaultVal := "CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"
				column.Default = &defaultVal
			}
		}
	case "auto_create", "on_create":
		// 创建时自动设置
		if strings.ToLower(value) == "current_timestamp" {
			column.NotNull = true
			if column.Default == nil {
				defaultVal := "CURRENT_TIMESTAMP"
				column.Default = &defaultVal
			}
		}
	}

	return nil
}

// ParseTormFlag 解析torm标签的标志
func (ma *ModelAnalyzer) ParseTormFlag(flag string, column *ModelColumn) {
	flag = strings.ToLower(flag)

	switch flag {
	case "primary_key", "pk", "primary", "primarykey":
		column.PrimaryKey = true
	case "auto_increment", "autoincrement", "auto_inc", "autoinc":
		column.AutoIncrement = true
	case "unique", "uniq":
		column.Unique = true
	case "not_null", "not null", "notnull", "not_nil", "notnil":
		column.NotNull = true
	case "nullable", "null":
		column.NotNull = false
	case "auto_create_time", "create_time", "created_at", "auto_created_at":
		// 自动创建时间字段
		column.NotNull = true
		if column.Default == nil {
			defaultVal := "CURRENT_TIMESTAMP"
			column.Default = &defaultVal
		}
	case "auto_update_time", "update_time", "updated_at", "auto_updated_at":
		// 自动更新时间字段
		column.NotNull = true
		if column.Default == nil {
			defaultVal := "CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"
			column.Default = &defaultVal
		}
	case "timestamp", "current_timestamp":
		// 时间戳字段
		column.NotNull = true
		if column.Default == nil {
			defaultVal := "CURRENT_TIMESTAMP"
			column.Default = &defaultVal
		}
	case "unsigned":
		// 无符号数字类型 (主要用于MySQL)
		// 这里可以通过修改类型名称来处理，但ModelColumn结构体没有专门的unsigned字段
		// 可以考虑在类型名称后添加UNSIGNED
	case "zerofill":
		// 零填充 (主要用于MySQL)
		// 类似unsigned，这里先忽略
	case "binary":
		// 二进制存储
		// 类似上面，先忽略
	case "index", "idx":
		// 普通索引标记
		// 由于ModelColumn结构体没有索引相关字段，这里先忽略
		// 实际实现中需要在Table级别处理索引
	case "fulltext", "fulltext_index":
		// 全文索引标记
		// 这里先忽略，需要在Table级别处理
	case "spatial", "spatial_index":
		// 空间索引标记
		// 这里先忽略，需要在Table级别处理
	case "btree", "hash", "rtree":
		// 索引类型标记
		// 这里先忽略，需要在Table级别处理
	case "foreign_key", "fk":
		// 外键标记
		// 这里先忽略，需要在Table级别处理外键关系
	case "references", "ref":
		// 引用标记
		// 这里先忽略，需要在Table级别处理
	case "on_delete_cascade", "cascade_delete":
		// 级联删除
		// 这里先忽略，外键相关
	case "on_update_cascade", "cascade_update":
		// 级联更新
		// 这里先忽略，外键相关
	case "generated", "virtual", "stored":
		// 生成列标记
		// 这里先忽略，需要扩展ModelColumn结构体
	case "json", "is_json":
		// JSON字段标记
		if column.Type == "" {
			column.Type = ColumnTypeJSON
		}
	case "encrypted", "encrypt":
		// 加密字段标记
		// 这里先忽略，需要在应用层处理
	case "hidden", "invisible":
		// 隐藏字段标记
		// 这里先忽略，需要扩展ModelColumn结构体
	case "readonly", "immutable":
		// 只读字段标记
		// 这里先忽略，需要在应用层处理
	}
}

// SetColumnType 设置列类型
func (ma *ModelAnalyzer) SetColumnType(typeStr string, column *ModelColumn) {
	typeStr = strings.ToLower(typeStr)

	switch typeStr {
	// 字符串类型
	case "varchar", "string":
		column.Type = ColumnTypeVarchar
	case "char", "character":
		column.Type = ColumnTypeChar
	case "text", "longtext", "mediumtext":
		column.Type = ColumnTypeText
	case "tinytext":
		column.Type = ColumnTypeText

	// 整数类型
	case "int", "integer", "int32":
		column.Type = ColumnTypeInt
	case "tinyint", "int8", "byte":
		column.Type = ColumnTypeTinyInt
	case "smallint", "int16", "short":
		column.Type = ColumnTypeSmallInt
	case "bigint", "int64", "long":
		column.Type = ColumnTypeBigInt
	case "mediumint":
		column.Type = ColumnTypeInt

	// 浮点类型
	case "float", "float32", "real":
		column.Type = ColumnTypeFloat
	case "double", "float64", "double_precision":
		column.Type = ColumnTypeDouble
	case "decimal", "numeric", "money":
		column.Type = ColumnTypeDecimal
		if column.Precision == 0 {
			column.Precision = 10 // 默认精度
		}
		if column.Scale == 0 {
			column.Scale = 2 // 默认小数位
		}

	// 布尔类型
	case "boolean", "bool", "bit":
		column.Type = ColumnTypeBoolean

	// 日期时间类型
	case "date":
		column.Type = ColumnTypeDate
	case "datetime", "datetime2":
		column.Type = ColumnTypeDateTime
	case "timestamp", "timestamptz":
		column.Type = ColumnTypeTimestamp
	case "time", "timetz":
		column.Type = ColumnTypeTime
	case "year":
		column.Type = ColumnTypeInt // 年份作为整数处理

	// 二进制类型
	case "blob", "binary", "varbinary":
		column.Type = ColumnTypeBlob
	case "tinyblob":
		column.Type = ColumnTypeBlob
	case "mediumblob":
		column.Type = ColumnTypeBlob
	case "longblob":
		column.Type = ColumnTypeBlob

	// JSON类型
	case "json", "jsonb":
		column.Type = ColumnTypeJSON

	// UUID类型 (作为VARCHAR处理)
	case "uuid", "guid":
		column.Type = ColumnTypeVarchar
		if column.Length == 0 {
			column.Length = 36 // UUID标准长度
		}

	// 枚举类型 (作为VARCHAR处理)
	case "enum", "set":
		column.Type = ColumnTypeVarchar
		if column.Length == 0 {
			column.Length = 255
		}

	// 几何类型 (作为TEXT处理)
	case "geometry", "point", "linestring", "polygon":
		column.Type = ColumnTypeText

	// 其他类型
	case "xml":
		column.Type = ColumnTypeText
	case "inet", "cidr", "macaddr":
		column.Type = ColumnTypeVarchar
	case "array":
		column.Type = ColumnTypeJSON // 数组作为JSON处理

	default:
		// 如果没有明确指定类型，保持从Go类型推断的类型
	}
}

// ParseDefaultValue 解析默认值
func (ma *ModelAnalyzer) ParseDefaultValue(value string) string {
	value = strings.TrimSpace(value)
	lowerValue := strings.ToLower(value)

	switch lowerValue {
	case "null":
		return "NULL"
	case "current_timestamp", "now()":
		return "CURRENT_TIMESTAMP"
	case "true":
		return "1"
	case "false":
		return "0"
	default:
		// 如果是数字，直接返回
		if _, err := strconv.ParseFloat(value, 64); err == nil {
			return value
		}
		// 其他情况添加引号
		return fmt.Sprintf("'%s'", strings.ReplaceAll(value, "'", "''"))
	}
}

// parseLegacyTags 解析传统标签
func (ma *ModelAnalyzer) parseLegacyTags(field reflect.StructField, column *ModelColumn) error {
	// 主键
	if field.Tag.Get("primaryKey") == "true" || field.Tag.Get("pk") != "" {
		column.PrimaryKey = true
	}

	// 自增
	if field.Tag.Get("autoIncrement") == "true" {
		column.AutoIncrement = true
	}

	// 大小
	if sizeStr := field.Tag.Get("size"); sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil {
			column.Length = size
		}
	}

	// 唯一
	if field.Tag.Get("unique") == "true" {
		column.Unique = true
	}

	// 默认值
	if defaultVal := field.Tag.Get("default"); defaultVal != "" {
		parsedDefault := ma.ParseDefaultValue(defaultVal)
		column.Default = &parsedDefault
	}

	// 注释
	if comment := field.Tag.Get("comment"); comment != "" {
		column.Comment = comment
	}

	// NOT NULL
	if field.Tag.Get("not_null") == "true" {
		column.NotNull = true
	}

	if field.Tag.Get("nullable") == "true" {
		column.NotNull = false
	}

	// 自动创建时间
	if field.Tag.Get("autoCreateTime") != "" {
		column.NotNull = true
		if column.Default == nil {
			defaultVal := "CURRENT_TIMESTAMP"
			column.Default = &defaultVal
		}
	}

	// 自动更新时间
	if field.Tag.Get("autoUpdateTime") != "" {
		column.NotNull = true
		if column.Default == nil {
			defaultVal := "CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"
			column.Default = &defaultVal
		}
	}

	return nil
}

// mapGoTypeToColumnType 将Go类型映射到数据库列类型
func (ma *ModelAnalyzer) mapGoTypeToColumnType(goType reflect.Type) ColumnType {
	// 处理指针类型
	if goType.Kind() == reflect.Ptr {
		goType = goType.Elem()
	}

	switch goType.Kind() {
	case reflect.String:
		return ColumnTypeVarchar
	case reflect.Int, reflect.Int32:
		return ColumnTypeInt
	case reflect.Int8:
		return ColumnTypeTinyInt
	case reflect.Int16:
		return ColumnTypeSmallInt
	case reflect.Int64:
		return ColumnTypeBigInt
	case reflect.Uint, reflect.Uint32:
		return ColumnTypeInt
	case reflect.Uint8:
		return ColumnTypeTinyInt
	case reflect.Uint16:
		return ColumnTypeSmallInt
	case reflect.Uint64:
		return ColumnTypeBigInt
	case reflect.Float32:
		return ColumnTypeFloat
	case reflect.Float64:
		return ColumnTypeDouble
	case reflect.Bool:
		return ColumnTypeBoolean
	case reflect.Slice:
		if goType.Elem().Kind() == reflect.Uint8 {
			// []byte
			return ColumnTypeBlob
		}
		// []string, []int等 - 存储为JSON
		return ColumnTypeJSON
	case reflect.Map:
		// map[string]interface{}等 - 存储为JSON
		return ColumnTypeJSON
	case reflect.Struct:
		// 检查是否是time.Time
		if goType.PkgPath() == "time" && goType.Name() == "Time" {
			return ColumnTypeDateTime
		}
		// 其他结构体存储为JSON
		return ColumnTypeJSON
	default:
		return ColumnTypeVarchar
	}
}

// toSnakeCase 将驼峰命名转换为蛇形命名
func (ma *ModelAnalyzer) toSnakeCase(str string) string {
	var result strings.Builder
	for i, r := range str {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
