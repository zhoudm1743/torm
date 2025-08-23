package migration

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/zhoudm1743/torm/db"
)

// DatabaseColumn 数据库中的列信息
type DatabaseColumn struct {
	Name          string
	Type          string
	Length        *int
	Precision     *int
	Scale         *int
	NotNull       bool
	Default       *string
	Comment       string
	PrimaryKey    bool
	Unique        bool
	AutoIncrement bool
}

// ModelColumn 模型中的列信息
type ModelColumn struct {
	Name          string
	Type          ColumnType
	Length        int
	Precision     int
	Scale         int
	NotNull       bool
	Default       *string
	Comment       string
	PrimaryKey    bool
	Unique        bool
	AutoIncrement bool
}

// ColumnDifference 列差异信息
type ColumnDifference struct {
	Column   string
	Type     string // "add", "modify", "drop"
	OldValue interface{}
	NewValue interface{}
	Reason   string
}

// SchemaDifference 表结构差异信息
type SchemaDifference struct {
	TableName string
	Columns   []ColumnDifference
	Indexes   []IndexDifference
}

// IndexDifference 索引差异信息
type IndexDifference struct {
	IndexName string
	Type      string // "add", "drop", "modify"
	OldIndex  *Index
	NewIndex  *Index
}

// SchemaComparator 表结构对比器
type SchemaComparator struct {
	conn   db.ConnectionInterface
	driver string
}

// NewSchemaComparator 创建表结构对比器
func NewSchemaComparator(conn db.ConnectionInterface) *SchemaComparator {
	return &SchemaComparator{
		conn:   conn,
		driver: conn.GetDriver(),
	}
}

// GetDatabaseColumns 获取数据库中表的列信息
func (sc *SchemaComparator) GetDatabaseColumns(tableName string) ([]DatabaseColumn, error) {
	switch sc.driver {
	case "mysql":
		return sc.getMySQLColumns(tableName)
	case "postgres", "postgresql":
		return sc.getPostgreSQLColumns(tableName)
	case "sqlite", "sqlite3":
		return sc.getSQLiteColumns(tableName)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", sc.driver)
	}
}

// getMySQLColumns 获取MySQL表列信息
func (sc *SchemaComparator) getMySQLColumns(tableName string) ([]DatabaseColumn, error) {
	query := `
		SELECT 
			COLUMN_NAME,
			DATA_TYPE,
			CHARACTER_MAXIMUM_LENGTH,
			NUMERIC_PRECISION,
			NUMERIC_SCALE,
			IS_NULLABLE,
			COLUMN_DEFAULT,
			COLUMN_COMMENT,
			COLUMN_KEY,
			EXTRA
		FROM information_schema.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	rows, err := sc.conn.Query(query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query MySQL columns: %w", err)
	}
	defer rows.Close()

	var columns []DatabaseColumn
	for rows.Next() {
		var col DatabaseColumn
		var dataType string
		var maxLength, precision, scale interface{}
		var nullable, defaultVal, comment, columnKey, extra interface{}

		err := rows.Scan(
			&col.Name, &dataType, &maxLength, &precision, &scale,
			&nullable, &defaultVal, &comment, &columnKey, &extra,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan MySQL column: %w", err)
		}

		col.Type = strings.ToUpper(dataType)

		// 处理长度
		if maxLength != nil {
			if length, ok := maxLength.(int64); ok {
				lengthInt := int(length)
				col.Length = &lengthInt
			}
		}

		// 处理精度
		if precision != nil {
			if prec, ok := precision.(int64); ok {
				precInt := int(prec)
				col.Precision = &precInt
			}
		}

		// 处理小数位
		if scale != nil {
			if sc, ok := scale.(int64); ok {
				scaleInt := int(sc)
				col.Scale = &scaleInt
			}
		}

		// 处理可空性
		if nullable != nil {
			col.NotNull = strings.ToUpper(nullable.(string)) == "NO"
		}

		// 处理默认值
		if defaultVal != nil {
			defaultStr := defaultVal.(string)
			col.Default = &defaultStr
		}

		// 处理注释
		if comment != nil {
			col.Comment = comment.(string)
		}

		// 处理主键
		if columnKey != nil {
			col.PrimaryKey = strings.ToUpper(columnKey.(string)) == "PRI"
			col.Unique = strings.ToUpper(columnKey.(string)) == "UNI"
		}

		// 处理自增
		if extra != nil {
			col.AutoIncrement = strings.Contains(strings.ToLower(extra.(string)), "auto_increment")
		}

		columns = append(columns, col)
	}

	return columns, nil
}

// getPostgreSQLColumns 获取PostgreSQL表列信息
func (sc *SchemaComparator) getPostgreSQLColumns(tableName string) ([]DatabaseColumn, error) {
	query := `
		SELECT 
			c.column_name,
			c.data_type,
			c.character_maximum_length,
			c.numeric_precision,
			c.numeric_scale,
			c.is_nullable,
			c.column_default,
			COALESCE(pgd.description, ''),
			CASE WHEN pk.column_name IS NOT NULL THEN 'PRI' ELSE '' END as column_key,
			CASE WHEN c.column_default LIKE 'nextval%' THEN 'auto_increment' ELSE '' END as extra
		FROM information_schema.columns c
		LEFT JOIN pg_catalog.pg_statio_all_tables st ON st.relname = c.table_name
		LEFT JOIN pg_catalog.pg_description pgd ON pgd.objoid = st.relid AND pgd.objsubid = c.ordinal_position
		LEFT JOIN (
			SELECT ku.column_name
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage ku ON tc.constraint_name = ku.constraint_name
			WHERE tc.constraint_type = 'PRIMARY KEY' AND tc.table_name = ?
		) pk ON pk.column_name = c.column_name
		WHERE c.table_schema = 'public' AND c.table_name = ?
		ORDER BY c.ordinal_position
	`

	rows, err := sc.conn.Query(query, tableName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query PostgreSQL columns: %w", err)
	}
	defer rows.Close()

	var columns []DatabaseColumn
	for rows.Next() {
		var col DatabaseColumn
		var dataType string
		var maxLength, precision, scale interface{}
		var nullable, defaultVal, comment, columnKey, extra interface{}

		err := rows.Scan(
			&col.Name, &dataType, &maxLength, &precision, &scale,
			&nullable, &defaultVal, &comment, &columnKey, &extra,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan PostgreSQL column: %w", err)
		}

		col.Type = strings.ToUpper(dataType)

		// 处理长度
		if maxLength != nil {
			if length, ok := maxLength.(int64); ok {
				lengthInt := int(length)
				col.Length = &lengthInt
			}
		}

		// 处理精度和小数位（类似MySQL处理）
		if precision != nil {
			if prec, ok := precision.(int64); ok {
				precInt := int(prec)
				col.Precision = &precInt
			}
		}

		if scale != nil {
			if sc, ok := scale.(int64); ok {
				scaleInt := int(sc)
				col.Scale = &scaleInt
			}
		}

		// 处理其他属性（类似MySQL处理）
		if nullable != nil {
			col.NotNull = strings.ToUpper(nullable.(string)) == "NO"
		}

		if defaultVal != nil {
			defaultStr := defaultVal.(string)
			col.Default = &defaultStr
		}

		if comment != nil {
			col.Comment = comment.(string)
		}

		if columnKey != nil {
			col.PrimaryKey = strings.ToUpper(columnKey.(string)) == "PRI"
		}

		if extra != nil {
			col.AutoIncrement = strings.Contains(strings.ToLower(extra.(string)), "auto_increment")
		}

		columns = append(columns, col)
	}

	return columns, nil
}

// getSQLiteColumns 获取SQLite表列信息
func (sc *SchemaComparator) getSQLiteColumns(tableName string) ([]DatabaseColumn, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)

	rows, err := sc.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query SQLite columns: %w", err)
	}
	defer rows.Close()

	var columns []DatabaseColumn
	for rows.Next() {
		var cid int
		var col DatabaseColumn
		var dataType string
		var notNull int
		var defaultVal interface{}
		var pk int

		err := rows.Scan(&cid, &col.Name, &dataType, &notNull, &defaultVal, &pk)
		if err != nil {
			return nil, fmt.Errorf("failed to scan SQLite column: %w", err)
		}

		// 解析类型和长度
		col.Type, col.Length = sc.parseSQLiteType(dataType)
		col.NotNull = notNull == 1
		col.PrimaryKey = pk == 1

		if defaultVal != nil {
			defaultStr := fmt.Sprintf("%v", defaultVal)
			col.Default = &defaultStr
		}

		// SQLite中检测自增
		if col.PrimaryKey && strings.ToUpper(col.Type) == "INTEGER" {
			// 检查是否有AUTOINCREMENT关键字
			createSQL, err := sc.getSQLiteCreateSQL(tableName)
			if err == nil && strings.Contains(strings.ToUpper(createSQL), "AUTOINCREMENT") {
				col.AutoIncrement = true
			}
		}

		columns = append(columns, col)
	}

	return columns, nil
}

// parseSQLiteType 解析SQLite类型定义
func (sc *SchemaComparator) parseSQLiteType(dataType string) (string, *int) {
	dataType = strings.ToUpper(dataType)

	// 查找括号中的长度信息
	if strings.Contains(dataType, "(") {
		parts := strings.Split(dataType, "(")
		baseType := strings.TrimSpace(parts[0])

		if len(parts) > 1 {
			lengthPart := strings.TrimSuffix(parts[1], ")")
			if length, err := strconv.Atoi(strings.TrimSpace(lengthPart)); err == nil {
				return baseType, &length
			}
		}
		return baseType, nil
	}

	return dataType, nil
}

// getSQLiteCreateSQL 获取SQLite表的CREATE语句
func (sc *SchemaComparator) getSQLiteCreateSQL(tableName string) (string, error) {
	query := "SELECT sql FROM sqlite_master WHERE type='table' AND name = ?"

	row := sc.conn.QueryRow(query, tableName)
	var createSQL string
	err := row.Scan(&createSQL)
	if err != nil {
		return "", fmt.Errorf("failed to get SQLite create SQL: %w", err)
	}

	return createSQL, nil
}

// CompareColumns 比较数据库列和模型列的差异
func (sc *SchemaComparator) CompareColumns(dbColumns []DatabaseColumn, modelColumns []ModelColumn) []ColumnDifference {
	var differences []ColumnDifference

	// 创建映射以便快速查找
	dbColMap := make(map[string]DatabaseColumn)
	for _, col := range dbColumns {
		dbColMap[col.Name] = col
	}

	modelColMap := make(map[string]ModelColumn)
	for _, col := range modelColumns {
		modelColMap[col.Name] = col
	}

	// 检查需要添加的列（在模型中但不在数据库中）
	for _, modelCol := range modelColumns {
		if _, exists := dbColMap[modelCol.Name]; !exists {
			differences = append(differences, ColumnDifference{
				Column:   modelCol.Name,
				Type:     "add",
				NewValue: modelCol,
				Reason:   "Column exists in model but not in database",
			})
		}
	}

	// 检查需要删除的列（在数据库中但不在模型中）
	for _, dbCol := range dbColumns {
		if _, exists := modelColMap[dbCol.Name]; !exists {
			differences = append(differences, ColumnDifference{
				Column:   dbCol.Name,
				Type:     "drop",
				OldValue: dbCol,
				Reason:   "Column exists in database but not in model",
			})
		}
	}

	// 检查需要修改的列
	for _, modelCol := range modelColumns {
		if dbCol, exists := dbColMap[modelCol.Name]; exists {
			if sc.needsModification(dbCol, modelCol) {
				differences = append(differences, ColumnDifference{
					Column:   modelCol.Name,
					Type:     "modify",
					OldValue: dbCol,
					NewValue: modelCol,
					Reason:   sc.getModificationReason(dbCol, modelCol),
				})
			}
		}
	}

	return differences
}

// needsModification 检查列是否需要修改
func (sc *SchemaComparator) needsModification(dbCol DatabaseColumn, modelCol ModelColumn) bool {
	// 比较类型
	if !sc.typesEqual(dbCol.Type, modelCol.Type) {
		return true
	}

	// 比较长度
	if !sc.lengthsEqual(dbCol.Length, modelCol.Length) {
		return true
	}

	// 比较精度和小数位
	if !sc.precisionsEqual(dbCol.Precision, modelCol.Precision) {
		return true
	}

	if !sc.scalesEqual(dbCol.Scale, modelCol.Scale) {
		return true
	}

	// 比较NOT NULL约束
	if dbCol.NotNull != modelCol.NotNull {
		return true
	}

	// 比较默认值
	if !sc.defaultsEqual(dbCol.Default, modelCol.Default) {
		return true
	}

	// 比较注释
	if dbCol.Comment != modelCol.Comment {
		return true
	}

	return false
}

// typesEqual 比较类型是否相等
func (sc *SchemaComparator) typesEqual(dbType string, modelType ColumnType) bool {
	// 将模型类型转换为数据库类型
	expectedDBType := sc.modelTypeToDBType(modelType)
	return strings.ToUpper(dbType) == strings.ToUpper(expectedDBType)
}

// modelTypeToDBType 将模型类型转换为数据库类型
func (sc *SchemaComparator) modelTypeToDBType(modelType ColumnType) string {
	switch sc.driver {
	case "mysql":
		return sc.modelTypeToMySQL(modelType)
	case "postgres", "postgresql":
		return sc.modelTypeToPostgreSQL(modelType)
	case "sqlite", "sqlite3":
		return sc.modelTypeToSQLite(modelType)
	default:
		return string(modelType)
	}
}

// modelTypeToMySQL 转换为MySQL类型
func (sc *SchemaComparator) modelTypeToMySQL(modelType ColumnType) string {
	switch modelType {
	case ColumnTypeVarchar:
		return "VARCHAR"
	case ColumnTypeChar:
		return "CHAR"
	case ColumnTypeText:
		return "TEXT"
	case ColumnTypeLongText:
		return "LONGTEXT"
	case ColumnTypeInt:
		return "INT"
	case ColumnTypeTinyInt:
		return "TINYINT"
	case ColumnTypeSmallInt:
		return "SMALLINT"
	case ColumnTypeBigInt:
		return "BIGINT"
	case ColumnTypeFloat:
		return "FLOAT"
	case ColumnTypeDouble:
		return "DOUBLE"
	case ColumnTypeDecimal:
		return "DECIMAL"
	case ColumnTypeBoolean:
		return "BOOLEAN"
	case ColumnTypeDate:
		return "DATE"
	case ColumnTypeDateTime:
		return "DATETIME"
	case ColumnTypeTimestamp:
		return "TIMESTAMP"
	case ColumnTypeTime:
		return "TIME"
	case ColumnTypeBlob:
		return "BLOB"
	case ColumnTypeJSON:
		return "JSON"
	default:
		return string(modelType)
	}
}

// modelTypeToPostgreSQL 转换为PostgreSQL类型
func (sc *SchemaComparator) modelTypeToPostgreSQL(modelType ColumnType) string {
	switch modelType {
	case ColumnTypeVarchar:
		return "CHARACTER VARYING"
	case ColumnTypeChar:
		return "CHARACTER"
	case ColumnTypeText:
		return "TEXT"
	case ColumnTypeLongText:
		return "TEXT"
	case ColumnTypeInt:
		return "INTEGER"
	case ColumnTypeTinyInt:
		return "SMALLINT"
	case ColumnTypeSmallInt:
		return "SMALLINT"
	case ColumnTypeBigInt:
		return "BIGINT"
	case ColumnTypeFloat:
		return "REAL"
	case ColumnTypeDouble:
		return "DOUBLE PRECISION"
	case ColumnTypeDecimal:
		return "NUMERIC"
	case ColumnTypeBoolean:
		return "BOOLEAN"
	case ColumnTypeDate:
		return "DATE"
	case ColumnTypeDateTime:
		return "TIMESTAMP"
	case ColumnTypeTimestamp:
		return "TIMESTAMP"
	case ColumnTypeTime:
		return "TIME"
	case ColumnTypeBlob:
		return "BYTEA"
	case ColumnTypeJSON:
		return "JSONB"
	default:
		return string(modelType)
	}
}

// modelTypeToSQLite 转换为SQLite类型
func (sc *SchemaComparator) modelTypeToSQLite(modelType ColumnType) string {
	switch modelType {
	case ColumnTypeVarchar, ColumnTypeChar, ColumnTypeText, ColumnTypeLongText:
		return "TEXT"
	case ColumnTypeInt, ColumnTypeTinyInt, ColumnTypeSmallInt, ColumnTypeBigInt:
		return "INTEGER"
	case ColumnTypeFloat, ColumnTypeDouble, ColumnTypeDecimal:
		return "REAL"
	case ColumnTypeBoolean:
		return "INTEGER"
	case ColumnTypeDate, ColumnTypeDateTime, ColumnTypeTimestamp, ColumnTypeTime:
		return "DATETIME"
	case ColumnTypeBlob:
		return "BLOB"
	case ColumnTypeJSON:
		return "TEXT"
	default:
		return "TEXT"
	}
}

// lengthsEqual 比较长度是否相等
func (sc *SchemaComparator) lengthsEqual(dbLength *int, modelLength int) bool {
	if dbLength == nil && modelLength == 0 {
		return true
	}
	if dbLength == nil || modelLength == 0 {
		return false
	}
	return *dbLength == modelLength
}

// precisionsEqual 比较精度是否相等
func (sc *SchemaComparator) precisionsEqual(dbPrecision *int, modelPrecision int) bool {
	if dbPrecision == nil && modelPrecision == 0 {
		return true
	}
	if dbPrecision == nil || modelPrecision == 0 {
		return false
	}
	return *dbPrecision == modelPrecision
}

// scalesEqual 比较小数位是否相等
func (sc *SchemaComparator) scalesEqual(dbScale *int, modelScale int) bool {
	if dbScale == nil && modelScale == 0 {
		return true
	}
	if dbScale == nil || modelScale == 0 {
		return false
	}
	return *dbScale == modelScale
}

// defaultsEqual 比较默认值是否相等
func (sc *SchemaComparator) defaultsEqual(dbDefault *string, modelDefault *string) bool {
	if dbDefault == nil && modelDefault == nil {
		return true
	}
	if dbDefault == nil || modelDefault == nil {
		return false
	}
	return *dbDefault == *modelDefault
}

// getModificationReason 获取修改原因
func (sc *SchemaComparator) getModificationReason(dbCol DatabaseColumn, modelCol ModelColumn) string {
	var reasons []string

	if !sc.typesEqual(dbCol.Type, modelCol.Type) {
		reasons = append(reasons, fmt.Sprintf("type changed from %s to %s", dbCol.Type, sc.modelTypeToDBType(modelCol.Type)))
	}

	if !sc.lengthsEqual(dbCol.Length, modelCol.Length) {
		oldLen := "none"
		if dbCol.Length != nil {
			oldLen = fmt.Sprintf("%d", *dbCol.Length)
		}
		newLen := "none"
		if modelCol.Length > 0 {
			newLen = fmt.Sprintf("%d", modelCol.Length)
		}
		reasons = append(reasons, fmt.Sprintf("length changed from %s to %s", oldLen, newLen))
	}

	if !sc.precisionsEqual(dbCol.Precision, modelCol.Precision) {
		oldPrec := "none"
		if dbCol.Precision != nil {
			oldPrec = fmt.Sprintf("%d", *dbCol.Precision)
		}
		newPrec := "none"
		if modelCol.Precision > 0 {
			newPrec = fmt.Sprintf("%d", modelCol.Precision)
		}
		reasons = append(reasons, fmt.Sprintf("precision changed from %s to %s", oldPrec, newPrec))
	}

	if !sc.scalesEqual(dbCol.Scale, modelCol.Scale) {
		oldScale := "none"
		if dbCol.Scale != nil {
			oldScale = fmt.Sprintf("%d", *dbCol.Scale)
		}
		newScale := "none"
		if modelCol.Scale > 0 {
			newScale = fmt.Sprintf("%d", modelCol.Scale)
		}
		reasons = append(reasons, fmt.Sprintf("scale changed from %s to %s", oldScale, newScale))
	}

	if dbCol.NotNull != modelCol.NotNull {
		oldNull := "NULL"
		if dbCol.NotNull {
			oldNull = "NOT NULL"
		}
		newNull := "NULL"
		if modelCol.NotNull {
			newNull = "NOT NULL"
		}
		reasons = append(reasons, fmt.Sprintf("nullability changed from %s to %s", oldNull, newNull))
	}

	if !sc.defaultsEqual(dbCol.Default, modelCol.Default) {
		oldDefault := "none"
		if dbCol.Default != nil {
			oldDefault = *dbCol.Default
		}
		newDefault := "none"
		if modelCol.Default != nil {
			newDefault = *modelCol.Default
		}
		reasons = append(reasons, fmt.Sprintf("default changed from '%s' to '%s'", oldDefault, newDefault))
	}

	if dbCol.Comment != modelCol.Comment {
		reasons = append(reasons, fmt.Sprintf("comment changed from '%s' to '%s'", dbCol.Comment, modelCol.Comment))
	}

	return strings.Join(reasons, "; ")
}
