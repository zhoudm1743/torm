package migration

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/zhoudm1743/torm/db"
)

// AutoMigrator 自动迁移器
type AutoMigrator struct {
	connection     db.ConnectionInterface
	analyzer       *ModelAnalyzer
	tableCache     map[string]bool // 表存在性缓存
	cacheEnabled   bool            // 是否启用缓存
	skipIfExists   bool            // 如果表存在则跳过检查（快速模式）
	structureCache map[string]bool // 表结构检查缓存
}

// NewAutoMigrator 创建自动迁移器
func NewAutoMigrator(conn db.ConnectionInterface) *AutoMigrator {
	return &AutoMigrator{
		connection:     conn,
		analyzer:       NewModelAnalyzer(),
		tableCache:     make(map[string]bool),
		cacheEnabled:   true,  // 默认启用缓存
		skipIfExists:   false, // 默认不跳过结构检查
		structureCache: make(map[string]bool),
	}
}

// SetCacheEnabled 设置是否启用表存在性缓存
func (am *AutoMigrator) SetCacheEnabled(enabled bool) {
	am.cacheEnabled = enabled
	if !enabled {
		am.tableCache = make(map[string]bool) // 清空缓存
	}
}

// ClearCache 清空表存在性缓存
func (am *AutoMigrator) ClearCache() {
	am.tableCache = make(map[string]bool)
	am.structureCache = make(map[string]bool)
}

// SetSkipIfExists 设置快速模式（如果表存在则跳过结构检查）
func (am *AutoMigrator) SetSkipIfExists(skip bool) {
	am.skipIfExists = skip
}

// MigrateModel 迁移模型到数据库
func (am *AutoMigrator) MigrateModel(modelInstance interface{}, tableName string) error {
	// 获取模型类型
	modelType := reflect.TypeOf(modelInstance)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 分析模型结构
	columns, err := am.analyzer.AnalyzeModel(modelType)
	if err != nil {
		return fmt.Errorf("分析模型失败: %w", err)
	}

	if len(columns) == 0 {
		return fmt.Errorf("模型没有可迁移的字段")
	}

	// 检查表是否存在
	exists, err := am.tableExists(tableName)
	if err != nil {
		return fmt.Errorf("检查表存在性失败: %w", err)
	}

	if !exists {
		// 创建新表
		return am.createTable(tableName, columns)
	}

	// 表已存在
	if am.skipIfExists {
		// 快速模式：如果表存在则跳过结构检查

		return nil
	}

	// 检查结构缓存
	if am.cacheEnabled {
		if checked, found := am.structureCache[tableName]; found && checked {

			return nil
		}
	}

	// 检查和更新表结构
	err = am.updateTableStructure(tableName, columns)

	// 缓存结构检查结果
	if am.cacheEnabled && err == nil {
		am.structureCache[tableName] = true
	}

	return err
}

// tableExists 检查表是否存在
func (am *AutoMigrator) tableExists(tableName string) (bool, error) {
	// 检查缓存
	if am.cacheEnabled {
		if exists, found := am.tableCache[tableName]; found {
			return exists, nil
		}
	}

	// 获取数据库驱动类型
	driver := am.getDriverType()

	var query string
	var args []interface{}

	switch driver {
	case "mysql":
		query = "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?"
		args = []interface{}{tableName}
	case "sqlite", "sqlite3":
		query = "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?"
		args = []interface{}{tableName}
	case "postgres", "postgresql":
		query = "SELECT COUNT(*) FROM information_schema.tables WHERE table_name = $1"
		args = []interface{}{tableName}
	default:
		return false, fmt.Errorf("不支持的数据库驱动: %s", driver)
	}

	row := am.connection.QueryRow(query, args...)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}

	exists := count > 0

	// 缓存结果
	if am.cacheEnabled {
		am.tableCache[tableName] = exists
	}

	return exists, nil
}

// createTable 创建表
func (am *AutoMigrator) createTable(tableName string, columns []ModelColumn) error {
	driver := am.getDriverType()

	sql := am.buildCreateTableSQL(tableName, columns, driver)

	_, err := am.connection.Exec(sql)
	if err != nil {
		return fmt.Errorf("创建表失败: %w", err)
	}

	// 更新缓存：表已存在
	if am.cacheEnabled {
		am.tableCache[tableName] = true
	}

	// 创建索引
	err = am.createIndexes(tableName, columns, driver)
	if err != nil {
		return fmt.Errorf("创建索引失败: %w", err)
	}

	// 创建外键约束
	err = am.createForeignKeys(tableName, columns, driver)
	if err != nil {
		return fmt.Errorf("创建外键约束失败: %w", err)
	}

	return nil
}

// buildCreateTableSQL 构建创建表的SQL
func (am *AutoMigrator) buildCreateTableSQL(tableName string, columns []ModelColumn, driver string) string {
	var sql strings.Builder

	sql.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", am.quoteIdentifier(tableName, driver)))

	var columnDefs []string
	var primaryKeys []string

	for _, col := range columns {
		colDef := am.buildColumnDefinition(col, driver)
		columnDefs = append(columnDefs, "  "+colDef)

		if col.PrimaryKey {
			primaryKeys = append(primaryKeys, am.quoteIdentifier(col.Name, driver))
		}
	}

	sql.WriteString(strings.Join(columnDefs, ",\n"))

	// 添加主键约束
	if len(primaryKeys) > 0 {
		sql.WriteString(",\n  PRIMARY KEY (")
		sql.WriteString(strings.Join(primaryKeys, ", "))
		sql.WriteString(")")
	}

	sql.WriteString("\n)")

	// MySQL特定的表选项
	if driver == "mysql" {
		sql.WriteString(" ENGINE=InnoDB DEFAULT CHARSET=utf8mb4")
	}

	return sql.String()
}

// buildColumnDefinition 构建列定义
func (am *AutoMigrator) buildColumnDefinition(col ModelColumn, driver string) string {
	var def strings.Builder

	def.WriteString(am.quoteIdentifier(col.Name, driver))
	def.WriteString(" ")
	def.WriteString(am.getColumnTypeSQL(col, driver))

	// NOT NULL
	if col.NotNull {
		def.WriteString(" NOT NULL")
	}

	// AUTO_INCREMENT (MySQL) / AUTOINCREMENT (SQLite)
	if col.AutoIncrement {
		switch driver {
		case "mysql":
			def.WriteString(" AUTO_INCREMENT")
		case "sqlite", "sqlite3":
			// SQLite的AUTOINCREMENT通常不需要显式指定
			// INTEGER PRIMARY KEY会自动递增
		case "postgres", "postgresql":
			// PostgreSQL使用SERIAL类型，不需要额外的AUTO_INCREMENT
			// SERIAL已经包含了自增功能
		}
	}

	// UNSIGNED (MySQL) - 必须在DEFAULT之前
	if col.Unsigned && driver == "mysql" {
		def.WriteString(" UNSIGNED")
	}

	// DEFAULT
	if col.Default != nil {
		def.WriteString(" DEFAULT ")
		def.WriteString(am.formatDefaultValue(col.Default, driver))
	}

	// UNIQUE
	if col.Unique {
		def.WriteString(" UNIQUE")
	}

	// ZEROFILL (MySQL)
	if col.Zerofill && driver == "mysql" {
		def.WriteString(" ZEROFILL")
	}

	// BINARY
	if col.Binary {
		switch driver {
		case "mysql":
			def.WriteString(" BINARY")
		case "sqlite", "sqlite3":
			// SQLite没有专门的BINARY修饰符
		case "postgres", "postgresql":
			// PostgreSQL使用不同的语法
		}
	}

	// 生成列 (MySQL 5.7+, PostgreSQL)
	if col.Generated != "" {
		switch driver {
		case "mysql":
			def.WriteString(fmt.Sprintf(" GENERATED ALWAYS AS (%s)", col.Generated))
			if col.Generated == "stored" {
				def.WriteString(" STORED")
			} else {
				def.WriteString(" VIRTUAL")
			}
		case "postgres", "postgresql":
			def.WriteString(fmt.Sprintf(" GENERATED ALWAYS AS (%s) STORED", col.Generated))
		}
	}

	// 外键约束（SQLite在列定义中处理，其他数据库在表级别处理）
	if col.ForeignKey != "" && (driver == "sqlite" || driver == "sqlite3") {
		refTable, refColumn := am.parseForeignKeyReference(col.ForeignKey)
		if refTable != "" && refColumn != "" {
			def.WriteString(fmt.Sprintf(" REFERENCES %s(%s)",
				am.quoteIdentifier(refTable, driver),
				am.quoteIdentifier(refColumn, driver)))

			// 添加删除和更新动作
			if col.OnDelete != "" {
				action := strings.ToUpper(strings.ReplaceAll(col.OnDelete, "_", " "))
				def.WriteString(fmt.Sprintf(" ON DELETE %s", action))
			}
			if col.OnUpdate != "" {
				action := strings.ToUpper(strings.ReplaceAll(col.OnUpdate, "_", " "))
				def.WriteString(fmt.Sprintf(" ON UPDATE %s", action))
			}
		}
	}

	// COMMENT (MySQL支持)
	if col.Comment != "" && driver == "mysql" {
		def.WriteString(fmt.Sprintf(" COMMENT '%s'", strings.ReplaceAll(col.Comment, "'", "''")))
	}

	return def.String()
}

// getColumnTypeSQL 获取列类型的SQL表示
func (am *AutoMigrator) getColumnTypeSQL(col ModelColumn, driver string) string {
	baseType := string(col.Type)

	switch col.Type {
	case ColumnTypeVarchar:
		length := col.Length
		if length <= 0 {
			length = 255 // 默认长度
		}
		return fmt.Sprintf("VARCHAR(%d)", length)

	case ColumnTypeChar:
		length := col.Length
		if length <= 0 {
			length = 1
		}
		return fmt.Sprintf("CHAR(%d)", length)

	case ColumnTypeDecimal:
		if col.Precision > 0 && col.Scale > 0 {
			return fmt.Sprintf("DECIMAL(%d,%d)", col.Precision, col.Scale)
		}
		return "DECIMAL(10,2)" // 默认精度

	case ColumnTypeInt:
		if driver == "sqlite" {
			return "INTEGER"
		} else if driver == "postgres" || driver == "postgresql" {
			// PostgreSQL中，如果是自增主键，使用SERIAL
			if col.AutoIncrement {
				return "SERIAL"
			}
			return "INTEGER"
		}
		return "INT"

	case ColumnTypeBigInt:
		if driver == "sqlite" {
			return "INTEGER"
		} else if driver == "postgres" || driver == "postgresql" {
			// PostgreSQL中，如果是自增主键，使用BIGSERIAL
			if col.AutoIncrement {
				return "BIGSERIAL"
			}
			return "BIGINT"
		}
		return "BIGINT"

	case ColumnTypeSmallInt:
		if driver == "sqlite" {
			return "INTEGER"
		} else if driver == "postgres" || driver == "postgresql" {
			// PostgreSQL中，如果是自增主键，使用SMALLSERIAL
			if col.AutoIncrement {
				return "SMALLSERIAL"
			}
			return "SMALLINT"
		}
		return "SMALLINT"

	// PostgreSQL SERIAL类型的直接处理
	case ColumnTypeSerial:
		if driver == "postgres" || driver == "postgresql" {
			return "SERIAL"
		}
		return "INT AUTO_INCREMENT" // 其他数据库退化为自增整型

	case ColumnTypeBigSerial:
		if driver == "postgres" || driver == "postgresql" {
			return "BIGSERIAL"
		}
		return "BIGINT AUTO_INCREMENT"

	case ColumnTypeSmallSerial:
		if driver == "postgres" || driver == "postgresql" {
			return "SMALLSERIAL"
		}
		return "SMALLINT AUTO_INCREMENT"

	case ColumnTypeText:
		return "TEXT"

	case ColumnTypeDateTime:
		if driver == "mysql" {
			return "DATETIME"
		} else if driver == "sqlite" {
			return "DATETIME"
		}
		return "TIMESTAMP"

	case ColumnTypeTimestamp:
		return "TIMESTAMP"

	case ColumnTypeBoolean:
		if driver == "mysql" {
			return "TINYINT(1)"
		} else if driver == "sqlite" {
			return "INTEGER"
		}
		return "BOOLEAN"

	default:
		return baseType
	}
}

// formatDefaultValue 格式化默认值
func (am *AutoMigrator) formatDefaultValue(value interface{}, driver string) string {
	if value == nil {
		return "NULL"
	}

	switch v := value.(type) {
	case string:
		// 检查是否是SQL关键字或函数
		lowerValue := strings.ToLower(v)
		if lowerValue == "current_timestamp" ||
			lowerValue == "now()" ||
			lowerValue == "null" ||
			strings.Contains(lowerValue, "current_timestamp") {
			// PostgreSQL和SQLite不支持 ON UPDATE CURRENT_TIMESTAMP
			if driver == "postgres" || driver == "postgresql" || driver == "sqlite" || driver == "sqlite3" {
				if strings.Contains(lowerValue, "on update") {
					return "CURRENT_TIMESTAMP" // 只返回CURRENT_TIMESTAMP，去掉ON UPDATE部分
				}
			}
			return v // 不加引号
		}
		return fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
	case int, int32, int64, float32, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		if v {
			return "1"
		}
		return "0"
	default:
		return fmt.Sprintf("'%v'", v)
	}
}

// createIndexes 创建索引
func (am *AutoMigrator) createIndexes(tableName string, columns []ModelColumn, driver string) error {
	for _, col := range columns {
		// 普通索引
		if col.Index {
			indexName := fmt.Sprintf("idx_%s_%s", tableName, col.Name)
			indexType := col.IndexType
			if indexType == "" {
				indexType = "btree" // 默认B-tree索引
			}

			var sql string
			switch driver {
			case "mysql":
				sql = fmt.Sprintf("CREATE INDEX %s ON %s (%s) USING %s",
					indexName, am.quoteIdentifier(tableName, driver),
					am.quoteIdentifier(col.Name, driver), strings.ToUpper(indexType))
			case "postgres", "postgresql":
				sql = fmt.Sprintf("CREATE INDEX %s ON %s USING %s (%s)",
					indexName, am.quoteIdentifier(tableName, driver),
					indexType, am.quoteIdentifier(col.Name, driver))
			case "sqlite", "sqlite3":
				sql = fmt.Sprintf("CREATE INDEX %s ON %s (%s)",
					indexName, am.quoteIdentifier(tableName, driver),
					am.quoteIdentifier(col.Name, driver))
			}

			_, err := am.connection.Exec(sql)
			if err != nil {
				return fmt.Errorf("创建索引 %s 失败: %w", indexName, err)
			}
		}

		// 全文索引
		if col.FulltextIndex {
			indexName := fmt.Sprintf("ft_%s_%s", tableName, col.Name)

			var sql string
			switch driver {
			case "mysql":
				sql = fmt.Sprintf("CREATE FULLTEXT INDEX %s ON %s (%s)",
					indexName, am.quoteIdentifier(tableName, driver),
					am.quoteIdentifier(col.Name, driver))
			case "postgres", "postgresql":
				// PostgreSQL使用GIN索引进行全文搜索
				sql = fmt.Sprintf("CREATE INDEX %s ON %s USING gin(to_tsvector('english', %s))",
					indexName, am.quoteIdentifier(tableName, driver),
					am.quoteIdentifier(col.Name, driver))
			case "sqlite", "sqlite3":
				// SQLite需要FTS扩展
				continue // 暂时跳过
			}

			_, err := am.connection.Exec(sql)
			if err != nil {
				return fmt.Errorf("创建全文索引 %s 失败: %w", indexName, err)
			}
		}

		// 空间索引
		if col.SpatialIndex {
			indexName := fmt.Sprintf("sp_%s_%s", tableName, col.Name)

			var sql string
			switch driver {
			case "mysql":
				sql = fmt.Sprintf("CREATE SPATIAL INDEX %s ON %s (%s)",
					indexName, am.quoteIdentifier(tableName, driver),
					am.quoteIdentifier(col.Name, driver))
			case "postgres", "postgresql":
				sql = fmt.Sprintf("CREATE INDEX %s ON %s USING gist (%s)",
					indexName, am.quoteIdentifier(tableName, driver),
					am.quoteIdentifier(col.Name, driver))
			case "sqlite", "sqlite3":
				// SQLite需要扩展
				continue // 暂时跳过
			}

			_, err := am.connection.Exec(sql)
			if err != nil {
				return fmt.Errorf("创建空间索引 %s 失败: %w", indexName, err)
			}
		}
	}

	return nil
}

// createForeignKeys 创建外键约束
func (am *AutoMigrator) createForeignKeys(tableName string, columns []ModelColumn, driver string) error {
	// SQLite外键约束必须在创建表时指定，不能后续添加
	if driver == "sqlite" || driver == "sqlite3" {
		fmt.Printf("ℹ️ SQLite外键约束已在表创建时定义\n")
		return nil
	}

	for _, col := range columns {
		if col.ForeignKey != "" {
			fkName := fmt.Sprintf("fk_%s_%s", tableName, col.Name)

			// 解析外键引用: "table.column" 或 "table(column)"
			refTable, refColumn := am.parseForeignKeyReference(col.ForeignKey)
			if refTable == "" || refColumn == "" {
				continue // 跳过格式错误的外键
			}

			var sql strings.Builder
			sql.WriteString(fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)",
				am.quoteIdentifier(tableName, driver), fkName,
				am.quoteIdentifier(col.Name, driver),
				am.quoteIdentifier(refTable, driver),
				am.quoteIdentifier(refColumn, driver)))

			// 添加删除和更新动作
			if col.OnDelete != "" {
				action := strings.ToUpper(strings.ReplaceAll(col.OnDelete, "_", " "))
				sql.WriteString(fmt.Sprintf(" ON DELETE %s", action))
			}
			if col.OnUpdate != "" {
				action := strings.ToUpper(strings.ReplaceAll(col.OnUpdate, "_", " "))
				sql.WriteString(fmt.Sprintf(" ON UPDATE %s", action))
			}

			_, err := am.connection.Exec(sql.String())
			if err != nil {
				return fmt.Errorf("创建外键约束 %s 失败: %w", fkName, err)
			}
		}
	}

	return nil
}

// parseForeignKeyReference 解析外键引用
func (am *AutoMigrator) parseForeignKeyReference(reference string) (table, column string) {
	// 支持 "table.column" 和 "table(column)" 格式
	if strings.Contains(reference, ".") {
		parts := strings.Split(reference, ".")
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		}
	} else if strings.Contains(reference, "(") && strings.Contains(reference, ")") {
		table = strings.TrimSpace(reference[:strings.Index(reference, "(")])
		column = strings.TrimSpace(reference[strings.Index(reference, "(")+1 : strings.Index(reference, ")")])
		return table, column
	}

	return "", ""
}

// quoteIdentifier 引用标识符
func (am *AutoMigrator) quoteIdentifier(identifier, driver string) string {
	switch driver {
	case "mysql":
		return "`" + identifier + "`"
	case "postgres", "postgresql":
		return `"` + identifier + `"`
	case "sqlite", "sqlite3":
		return `"` + identifier + `"`
	default:
		return identifier
	}
}

// updateTableStructure 更新表结构
func (am *AutoMigrator) updateTableStructure(tableName string, modelColumns []ModelColumn) error {

	// 获取当前表的列信息
	existingColumns, err := am.getTableColumns(tableName)
	if err != nil {
		return fmt.Errorf("获取表列信息失败: %w", err)
	}

	// 比较列差异
	toAdd, toModify, err := am.compareColumns(existingColumns, modelColumns)
	if err != nil {
		return fmt.Errorf("比较列差异失败: %w", err)
	}

	// 执行表结构更新
	var alterCount int

	// 添加新列
	for _, column := range toAdd {
		if err := am.addColumn(tableName, column); err != nil {
			return fmt.Errorf("添加列 %s 失败: %w", column.Name, err)
		}
		alterCount++
	}

	// 修改现有列
	for _, column := range toModify {
		if err := am.modifyColumn(tableName, column); err != nil {
			return fmt.Errorf("修改列 %s 失败: %w", column.Name, err)
		}
		alterCount++
	}

	// 添加索引和约束
	if err := am.addIndexesAndConstraints(tableName, modelColumns); err != nil {
		return fmt.Errorf("添加索引和约束失败: %w", err)
	}

	return nil
}

// getTableColumns 获取表的现有列信息
func (am *AutoMigrator) getTableColumns(tableName string) (map[string]ModelColumn, error) {
	driver := am.getDriverType()
	columns := make(map[string]ModelColumn)

	var query string
	var args []interface{}

	switch driver {
	case "mysql":
		query = "DESCRIBE " + tableName
	case "postgres":
		query = `SELECT 
			column_name,
			data_type,
			is_nullable,
			column_default,
			character_maximum_length,
			numeric_precision,
			numeric_scale
		FROM information_schema.columns 
		WHERE table_name = $1 
		ORDER BY ordinal_position`
		args = []interface{}{tableName}
	case "sqlite":
		query = "PRAGMA table_info(" + tableName + ")"
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %s", driver)
	}

	rows, err := am.connection.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var column ModelColumn
		switch driver {
		case "mysql":
			var field, fieldType, null, key, extra string
			var defaultVal sql.NullString
			if err := rows.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra); err != nil {
				return nil, err
			}
			defaultValue := ""
			if defaultVal.Valid {
				defaultValue = defaultVal.String
			}
			column = am.parseMySQLColumn(field, fieldType, null, key, defaultValue, extra)
		case "postgres":
			var columnName, dataType, isNullable string
			var columnDefault sql.NullString
			var charMaxLength, numericPrecision, numericScale sql.NullInt64
			if err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault,
				&charMaxLength, &numericPrecision, &numericScale); err != nil {
				return nil, err
			}
			column = am.parsePostgreSQLColumn(columnName, dataType, isNullable, columnDefault,
				charMaxLength, numericPrecision, numericScale)
		case "sqlite":
			var cid int
			var name, colType string
			var notNull, pk int
			var defaultVal sql.NullString
			if err := rows.Scan(&cid, &name, &colType, &notNull, &defaultVal, &pk); err != nil {
				return nil, err
			}
			column = am.parseSQLiteColumn(name, colType, notNull, pk, defaultVal)
		}
		columns[column.Name] = column
	}

	return columns, nil
}

// compareColumns 比较现有列和模型列，返回需要添加和修改的列
func (am *AutoMigrator) compareColumns(existing map[string]ModelColumn, model []ModelColumn) ([]ModelColumn, []ModelColumn, error) {
	var toAdd []ModelColumn
	var toModify []ModelColumn

	// 检查模型中的每一列
	for _, modelCol := range model {
		existingCol, exists := existing[modelCol.Name]

		if !exists {
			// 列不存在，需要添加
			toAdd = append(toAdd, modelCol)
		} else {
			// 列存在，检查是否需要修改
			if am.columnNeedsUpdate(existingCol, modelCol) {
				toModify = append(toModify, modelCol)
			}
		}
	}

	return toAdd, toModify, nil
}

// columnNeedsUpdate 检查列是否需要更新
func (am *AutoMigrator) columnNeedsUpdate(existing, model ModelColumn) bool {
	driver := am.getDriverType()

	// PostgreSQL的SERIAL类型特殊处理
	if driver == "postgres" {
		// 如果现有列是SERIAL类型，且模型列是带AUTO_INCREMENT的INT类型，认为不需要更新
		if (existing.Type == ColumnTypeSerial || strings.ToUpper(string(existing.Type)) == "SERIAL") &&
			model.Type == ColumnTypeInt && model.AutoIncrement {
			return false
		}

		// 如果现有列是BIGSERIAL类型，且模型列是带AUTO_INCREMENT的BIGINT类型，认为不需要更新
		if (existing.Type == ColumnTypeBigSerial || strings.ToUpper(string(existing.Type)) == "BIGSERIAL") &&
			model.Type == ColumnTypeBigInt && model.AutoIncrement {
			return false
		}

		// 如果现有列是SMALLSERIAL类型，且模型列是带AUTO_INCREMENT的SMALLINT类型，认为不需要更新
		if (existing.Type == ColumnTypeSmallSerial || strings.ToUpper(string(existing.Type)) == "SMALLSERIAL") &&
			model.Type == ColumnTypeSmallInt && model.AutoIncrement {
			return false
		}

		// 如果现有列是INT类型但模型要求SERIAL，也认为不需要更新（因为SERIAL本质上是INT + SEQUENCE）
		if existing.Type == ColumnTypeInt && existing.AutoIncrement &&
			model.Type == ColumnTypeInt && model.AutoIncrement {
			return false
		}
	}

	// 简单比较，后续可以扩展
	if existing.Type != model.Type {
		return true
	}
	if existing.Length != model.Length {
		return true
	}
	if existing.NotNull != model.NotNull {
		return true
	}
	// 检查默认值
	if fmt.Sprintf("%v", existing.Default) != fmt.Sprintf("%v", model.Default) {
		return true
	}
	return false
}

// addColumn 添加新列
func (am *AutoMigrator) addColumn(tableName string, column ModelColumn) error {
	sql := am.buildAddColumnSQL(tableName, column)
	return am.execSQL(sql)
}

// modifyColumn 修改现有列
func (am *AutoMigrator) modifyColumn(tableName string, column ModelColumn) error {
	sql := am.buildModifyColumnSQL(tableName, column)
	return am.execSQL(sql)
}

// buildAddColumnSQL 构建添加列的SQL
func (am *AutoMigrator) buildAddColumnSQL(tableName string, column ModelColumn) string {
	driver := am.getDriverType()
	columnSQL := am.buildColumnDefinition(column, driver)

	switch driver {
	case "mysql":
		return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tableName, columnSQL)
	case "postgres":
		return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tableName, columnSQL)
	case "sqlite":
		return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tableName, columnSQL)
	default:
		return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tableName, columnSQL)
	}
}

// buildModifyColumnSQL 构建修改列的SQL
func (am *AutoMigrator) buildModifyColumnSQL(tableName string, column ModelColumn) string {
	driver := am.getDriverType()

	switch driver {
	case "mysql":
		// MySQL使用MODIFY COLUMN，需要完整的列定义包括AUTO_INCREMENT
		columnSQL := am.buildColumnDefinition(column, driver)
		return fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s", tableName, columnSQL)
	case "postgres":
		// PostgreSQL需要分步修改，只修改类型
		columnSQL := am.getColumnTypeSQL(column, driver)
		return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s", tableName, column.Name, columnSQL)
	case "sqlite":
		// SQLite不支持修改列，需要重建表
		return fmt.Sprintf("-- SQLite不支持直接修改列 %s", column.Name)
	default:
		columnSQL := am.getColumnTypeSQL(column, driver)
		return fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s %s", tableName, column.Name, columnSQL)
	}
}

// execSQL 执行SQL语句
func (am *AutoMigrator) execSQL(sql string) error {
	// 跳过注释语句
	if strings.HasPrefix(strings.TrimSpace(sql), "--") {
		return nil
	}

	_, err := am.connection.Exec(sql)
	return err
}

// getDriverType 获取数据库驱动类型
func (am *AutoMigrator) getDriverType() string {
	// 尝试从连接接口获取驱动信息
	if drivenConn, ok := am.connection.(interface{ GetDriver() string }); ok {
		return drivenConn.GetDriver()
	}

	// 默认返回mysql（后续可以改进）
	return "mysql"
}

// parseMySQLColumn 解析MySQL列信息
func (am *AutoMigrator) parseMySQLColumn(field, fieldType, null, key, defaultVal, extra string) ModelColumn {
	column := ModelColumn{
		Name:    field,
		NotNull: null == "NO",
	}

	// 解析类型和长度
	if strings.Contains(fieldType, "(") {
		parts := strings.Split(fieldType, "(")
		column.Type = ColumnType(strings.ToUpper(parts[0]))
		if len(parts) > 1 {
			sizeStr := strings.TrimSuffix(parts[1], ")")
			if sizeStr != "" {
				// 简单处理长度，转换为int
				var length int
				fmt.Sscanf(sizeStr, "%d", &length)
				column.Length = length
			}
		}
	} else {
		column.Type = ColumnType(strings.ToUpper(fieldType))
	}

	// 处理默认值
	if defaultVal != "" && defaultVal != "NULL" {
		column.Default = defaultVal
	}

	// 处理主键
	if key == "PRI" {
		column.PrimaryKey = true
	}

	// 处理自增
	if strings.Contains(extra, "auto_increment") {
		column.AutoIncrement = true
	}

	return column
}

// parsePostgreSQLColumn 解析PostgreSQL列信息
func (am *AutoMigrator) parsePostgreSQLColumn(columnName, dataType, isNullable string,
	columnDefault sql.NullString, charMaxLength, numericPrecision, numericScale sql.NullInt64) ModelColumn {

	column := ModelColumn{
		Name:    columnName,
		NotNull: isNullable == "NO",
	}

	// 解析类型
	switch strings.ToLower(dataType) {
	case "integer", "int4":
		column.Type = ColumnTypeInt
		// 检查是否是SERIAL类型（通过默认值判断）
		if columnDefault.Valid && strings.Contains(columnDefault.String, "nextval") {
			column.Type = ColumnType("SERIAL")
			column.AutoIncrement = true
		}
	case "bigint", "int8":
		column.Type = ColumnTypeBigInt
		// 检查是否是BIGSERIAL类型
		if columnDefault.Valid && strings.Contains(columnDefault.String, "nextval") {
			column.Type = ColumnType("BIGSERIAL")
			column.AutoIncrement = true
		}
	case "smallint", "int2":
		column.Type = ColumnTypeSmallInt
		// 检查是否是SMALLSERIAL类型
		if columnDefault.Valid && strings.Contains(columnDefault.String, "nextval") {
			column.Type = ColumnType("SMALLSERIAL")
			column.AutoIncrement = true
		}
	case "character varying", "varchar":
		column.Type = ColumnTypeVarchar
		if charMaxLength.Valid {
			column.Length = int(charMaxLength.Int64)
		}
	case "text":
		column.Type = ColumnTypeText
	case "timestamp without time zone", "timestamp":
		column.Type = ColumnTypeTimestamp
	case "boolean":
		column.Type = ColumnTypeBoolean
	case "numeric", "decimal":
		column.Type = ColumnTypeDecimal
		if numericPrecision.Valid {
			column.Precision = int(numericPrecision.Int64)
		}
		if numericScale.Valid {
			column.Scale = int(numericScale.Int64)
		}
	default:
		column.Type = ColumnType(strings.ToUpper(dataType))
	}

	// 处理默认值
	if columnDefault.Valid && columnDefault.String != "" {
		column.Default = columnDefault.String
	}

	return column
}

// parseSQLiteColumn 解析SQLite列信息
func (am *AutoMigrator) parseSQLiteColumn(name, colType string, notNull, pk int, defaultVal sql.NullString) ModelColumn {
	column := ModelColumn{
		Name:    name,
		NotNull: notNull == 1,
	}

	// 解析类型
	upperType := strings.ToUpper(colType)
	if strings.Contains(upperType, "(") {
		parts := strings.Split(upperType, "(")
		column.Type = ColumnType(parts[0])
		if len(parts) > 1 {
			sizeStr := strings.TrimSuffix(parts[1], ")")
			if sizeStr != "" {
				var length int
				fmt.Sscanf(sizeStr, "%d", &length)
				column.Length = length
			}
		}
	} else {
		column.Type = ColumnType(upperType)
	}

	// 处理默认值
	if defaultVal.Valid && defaultVal.String != "" {
		column.Default = defaultVal.String
	}

	// 处理主键
	if pk == 1 {
		column.PrimaryKey = true
	}

	return column
}

// addIndexesAndConstraints 添加索引和约束
func (am *AutoMigrator) addIndexesAndConstraints(tableName string, columns []ModelColumn) error {
	for _, column := range columns {
		// 创建普通索引
		if column.Index && !column.Unique && !column.PrimaryKey {
			indexName := fmt.Sprintf("idx_%s_%s", tableName, column.Name)
			if err := am.createIndex(tableName, indexName, column.Name, false); err != nil {
				fmt.Printf("  ⚠️ 创建索引 %s 失败: %v\n", indexName, err)
			}
		}

		// 创建唯一约束/唯一索引
		if column.Unique && !column.PrimaryKey {
			indexName := fmt.Sprintf("idx_%s_%s_unique", tableName, column.Name)
			if err := am.createIndex(tableName, indexName, column.Name, true); err != nil {
				fmt.Printf("  ⚠️ 创建唯一约束 %s 失败: %v\n", indexName, err)
			}
		}
	}
	return nil
}

// createIndex 创建索引
func (am *AutoMigrator) createIndex(tableName, indexName, columnName string, unique bool) error {
	var sql string

	// 检查索引是否已存在
	if am.indexExists(tableName, indexName) {
		return nil // 索引已存在，跳过
	}

	driver := am.connection.GetDriver()
	switch driver {
	case "mysql":
		if unique {
			sql = fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s (%s)", indexName, tableName, columnName)
		} else {
			sql = fmt.Sprintf("CREATE INDEX %s ON %s (%s)", indexName, tableName, columnName)
		}
	case "postgres":
		if unique {
			sql = fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s (%s)", indexName, tableName, columnName)
		} else {
			sql = fmt.Sprintf("CREATE INDEX %s ON %s (%s)", indexName, tableName, columnName)
		}
	case "sqlite":
		if unique {
			sql = fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s (%s)", indexName, tableName, columnName)
		} else {
			sql = fmt.Sprintf("CREATE INDEX %s ON %s (%s)", indexName, tableName, columnName)
		}
	default:
		return fmt.Errorf("不支持的数据库类型: %s", driver)
	}

	_, err := am.connection.Exec(sql)
	return err
}

// indexExists 检查索引是否存在
func (am *AutoMigrator) indexExists(tableName, indexName string) bool {
	var count int
	var sql string

	driver := am.connection.GetDriver()
	switch driver {
	case "mysql":
		sql = "SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?"
		am.connection.QueryRow(sql, tableName, indexName).Scan(&count)
	case "postgres":
		sql = "SELECT COUNT(*) FROM pg_indexes WHERE tablename = $1 AND indexname = $2"
		am.connection.QueryRow(sql, tableName, indexName).Scan(&count)
	case "sqlite":
		sql = "SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name = ? AND tbl_name = ?"
		am.connection.QueryRow(sql, indexName, tableName).Scan(&count)
	}

	return count > 0
}
