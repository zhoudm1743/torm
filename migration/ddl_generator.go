package migration

import (
	"fmt"
	"strings"
)

// DDLGenerator DDL语句生成器
// 支持MySQL、PostgreSQL、SQLite的跨数据库DDL生成
type DDLGenerator struct {
	driver string
}

// NewDDLGenerator 创建DDL生成器
func NewDDLGenerator(driver string) *DDLGenerator {
	return &DDLGenerator{
		driver: driver,
	}
}

// GetDriver 获取数据库驱动类型
func (g *DDLGenerator) GetDriver() string {
	return g.driver
}

// QuoteIdentifier 引用标识符（表名、列名等）
func (g *DDLGenerator) QuoteIdentifier(name string) string {
	switch g.driver {
	case "mysql":
		return "`" + name + "`"
	case "postgres", "postgresql":
		return `"` + name + `"`
	case "sqlite", "sqlite3":
		return `"` + name + `"`
	case "sqlserver", "mssql":
		return "[" + name + "]"
	default:
		return name
	}
}

// GenerateCreateTableSQL 生成创建表的SQL
func (g *DDLGenerator) GenerateCreateTableSQL(table *Table) string {
	var sql strings.Builder

	sql.WriteString("CREATE TABLE ")
	sql.WriteString(g.QuoteIdentifier(table.Name))
	sql.WriteString(" (\n")

	var columnDefs []string
	var primaryKeys []string

	// 生成列定义
	for _, col := range table.Columns {
		colDef := g.generateColumnDefinition(col)
		columnDefs = append(columnDefs, "  "+colDef)

		if col.PrimaryKey {
			primaryKeys = append(primaryKeys, g.QuoteIdentifier(col.Name))
		}
	}

	sql.WriteString(strings.Join(columnDefs, ",\n"))

	// 添加主键约束
	if len(primaryKeys) > 0 {
		sql.WriteString(",\n  PRIMARY KEY (")
		sql.WriteString(strings.Join(primaryKeys, ", "))
		sql.WriteString(")")
	}

	// SQLite: 在表定义中添加外键
	if g.driver == "sqlite" || g.driver == "sqlite3" {
		for _, fk := range table.ForeignKeys {
			sql.WriteString(",\n  ")
			sql.WriteString(g.generateInlineForeignKey(fk))
		}
	}

	sql.WriteString("\n)")

	// 添加表选项
	sql.WriteString(g.generateTableOptions(table))

	return sql.String()
}

// generateColumnDefinition 生成列定义
func (g *DDLGenerator) generateColumnDefinition(col *Column) string {
	var def strings.Builder

	def.WriteString(g.QuoteIdentifier(col.Name))
	def.WriteString(" ")
	def.WriteString(g.generateColumnType(col))

	// NOT NULL
	if col.NotNull {
		def.WriteString(" NOT NULL")
	}

	// AUTO_INCREMENT
	if col.AutoIncrement {
		switch g.driver {
		case "mysql":
			def.WriteString(" AUTO_INCREMENT")
		case "sqlserver", "mssql":
			def.WriteString(" IDENTITY(1,1)")
			// PostgreSQL 使用 SERIAL 类型，SQLite 使用 INTEGER PRIMARY KEY
		}
	}

	// UNIQUE
	if col.Unique && !col.PrimaryKey {
		def.WriteString(" UNIQUE")
	}

	// DEFAULT
	if col.Default != nil {
		def.WriteString(" DEFAULT ")
		def.WriteString(g.formatDefaultValue(col.Default))
	}

	// COMMENT (MySQL only)
	if col.Comment != "" && g.driver == "mysql" {
		def.WriteString(fmt.Sprintf(" COMMENT '%s'", strings.ReplaceAll(col.Comment, "'", "''")))
	}

	return def.String()
}

// generateColumnType 生成列类型
func (g *DDLGenerator) generateColumnType(col *Column) string {
	switch g.driver {
	case "mysql":
		return g.generateMySQLColumnType(col)
	case "postgres", "postgresql":
		return g.generatePostgreSQLColumnType(col)
	case "sqlite", "sqlite3":
		return g.generateSQLiteColumnType(col)
	case "sqlserver", "mssql":
		return g.generateSQLServerColumnType(col)
	default:
		return string(col.Type)
	}
}

// generateMySQLColumnType 生成MySQL列类型
func (g *DDLGenerator) generateMySQLColumnType(col *Column) string {
	switch col.Type {
	case ColumnTypeInt:
		return "INT"
	case ColumnTypeBigInt:
		return "BIGINT"
	case ColumnTypeSmallInt:
		return "SMALLINT"
	case ColumnTypeTinyInt:
		return "TINYINT"
	case ColumnTypeFloat:
		return "FLOAT"
	case ColumnTypeDouble:
		return "DOUBLE"
	case ColumnTypeDecimal:
		if col.Precision > 0 && col.Scale > 0 {
			return fmt.Sprintf("DECIMAL(%d,%d)", col.Precision, col.Scale)
		}
		return "DECIMAL(10,2)"
	case ColumnTypeVarchar:
		length := col.Length
		if length <= 0 {
			length = 255
		}
		return fmt.Sprintf("VARCHAR(%d)", length)
	case ColumnTypeChar:
		length := col.Length
		if length <= 0 {
			length = 1
		}
		return fmt.Sprintf("CHAR(%d)", length)
	case ColumnTypeText:
		return "TEXT"
	case ColumnTypeLongText:
		return "LONGTEXT"
	case ColumnTypeDateTime:
		return "DATETIME"
	case ColumnTypeTimestamp:
		return "TIMESTAMP"
	case ColumnTypeDate:
		return "DATE"
	case ColumnTypeTime:
		return "TIME"
	case ColumnTypeBoolean:
		return "TINYINT(1)"
	case ColumnTypeBlob:
		return "BLOB"
	case ColumnTypeJSON:
		return "JSON"
	default:
		return string(col.Type)
	}
}

// generatePostgreSQLColumnType 生成PostgreSQL列类型
func (g *DDLGenerator) generatePostgreSQLColumnType(col *Column) string {
	switch col.Type {
	case ColumnTypeInt:
		if col.AutoIncrement {
			return "SERIAL"
		}
		return "INTEGER"
	case ColumnTypeBigInt:
		if col.AutoIncrement {
			return "BIGSERIAL"
		}
		return "BIGINT"
	case ColumnTypeSmallInt:
		if col.AutoIncrement {
			return "SMALLSERIAL"
		}
		return "SMALLINT"
	case ColumnTypeTinyInt:
		return "SMALLINT"
	case ColumnTypeFloat:
		return "REAL"
	case ColumnTypeDouble:
		return "DOUBLE PRECISION"
	case ColumnTypeDecimal:
		if col.Precision > 0 && col.Scale > 0 {
			return fmt.Sprintf("DECIMAL(%d,%d)", col.Precision, col.Scale)
		}
		return "DECIMAL(10,2)"
	case ColumnTypeVarchar:
		length := col.Length
		if length <= 0 {
			length = 255
		}
		return fmt.Sprintf("VARCHAR(%d)", length)
	case ColumnTypeChar:
		length := col.Length
		if length <= 0 {
			length = 1
		}
		return fmt.Sprintf("CHAR(%d)", length)
	case ColumnTypeText, ColumnTypeLongText:
		return "TEXT"
	case ColumnTypeDateTime, ColumnTypeTimestamp:
		return "TIMESTAMP"
	case ColumnTypeDate:
		return "DATE"
	case ColumnTypeTime:
		return "TIME"
	case ColumnTypeBoolean:
		return "BOOLEAN"
	case ColumnTypeBlob:
		return "BYTEA"
	case ColumnTypeJSON:
		return "JSONB"
	default:
		return string(col.Type)
	}
}

// generateSQLiteColumnType 生成SQLite列类型
func (g *DDLGenerator) generateSQLiteColumnType(col *Column) string {
	switch col.Type {
	case ColumnTypeInt, ColumnTypeBigInt, ColumnTypeSmallInt, ColumnTypeTinyInt:
		return "INTEGER"
	case ColumnTypeFloat, ColumnTypeDouble, ColumnTypeDecimal:
		return "REAL"
	case ColumnTypeVarchar, ColumnTypeChar, ColumnTypeText, ColumnTypeLongText:
		return "TEXT"
	case ColumnTypeDateTime, ColumnTypeTimestamp, ColumnTypeDate, ColumnTypeTime:
		return "DATETIME"
	case ColumnTypeBoolean:
		return "INTEGER"
	case ColumnTypeBlob:
		return "BLOB"
	case ColumnTypeJSON:
		return "TEXT"
	default:
		return "TEXT"
	}
}

// generateSQLServerColumnType 生成SQL Server列类型
func (g *DDLGenerator) generateSQLServerColumnType(col *Column) string {
	switch col.Type {
	case ColumnTypeInt:
		return "INT"
	case ColumnTypeBigInt:
		return "BIGINT"
	case ColumnTypeSmallInt:
		return "SMALLINT"
	case ColumnTypeTinyInt:
		return "TINYINT"
	case ColumnTypeFloat, ColumnTypeDouble:
		return "FLOAT"
	case ColumnTypeDecimal:
		if col.Precision > 0 && col.Scale > 0 {
			return fmt.Sprintf("DECIMAL(%d,%d)", col.Precision, col.Scale)
		}
		return "DECIMAL(10,2)"
	case ColumnTypeVarchar:
		length := col.Length
		if length <= 0 {
			length = 255
		}
		return fmt.Sprintf("NVARCHAR(%d)", length)
	case ColumnTypeChar:
		length := col.Length
		if length <= 0 {
			length = 1
		}
		return fmt.Sprintf("NCHAR(%d)", length)
	case ColumnTypeText, ColumnTypeLongText:
		return "NTEXT"
	case ColumnTypeDateTime, ColumnTypeTimestamp:
		return "DATETIME2"
	case ColumnTypeDate:
		return "DATE"
	case ColumnTypeTime:
		return "TIME"
	case ColumnTypeBoolean:
		return "BIT"
	case ColumnTypeBlob:
		return "VARBINARY(MAX)"
	case ColumnTypeJSON:
		return "NVARCHAR(MAX)"
	default:
		return string(col.Type)
	}
}

// formatDefaultValue 格式化默认值
func (g *DDLGenerator) formatDefaultValue(value interface{}) string {
	if value == nil {
		return "NULL"
	}

	switch v := value.(type) {
	case string:
		lowerValue := strings.ToLower(v)
		// SQL关键字和函数不加引号
		if lowerValue == "current_timestamp" || lowerValue == "now()" || lowerValue == "null" {
			// PostgreSQL和SQLite不支持 ON UPDATE CURRENT_TIMESTAMP
			if (g.driver == "postgres" || g.driver == "postgresql" || g.driver == "sqlite" || g.driver == "sqlite3") &&
				strings.Contains(lowerValue, "on update") {
				return "CURRENT_TIMESTAMP"
			}
			return v
		}
		return fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%v", v)
	case float32, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		if g.driver == "postgres" || g.driver == "postgresql" {
			if v {
				return "true"
			}
			return "false"
		}
		if v {
			return "1"
		}
		return "0"
	default:
		return fmt.Sprintf("'%v'", v)
	}
}

// generateTableOptions 生成表选项
func (g *DDLGenerator) generateTableOptions(table *Table) string {
	switch g.driver {
	case "mysql":
		var options []string
		if table.Engine != "" {
			options = append(options, fmt.Sprintf("ENGINE=%s", table.Engine))
		} else {
			options = append(options, "ENGINE=InnoDB")
		}
		if table.Charset != "" {
			options = append(options, fmt.Sprintf("DEFAULT CHARSET=%s", table.Charset))
		} else {
			options = append(options, "DEFAULT CHARSET=utf8mb4")
		}
		if table.Comment != "" {
			options = append(options, fmt.Sprintf("COMMENT='%s'", strings.ReplaceAll(table.Comment, "'", "''")))
		}
		if len(options) > 0 {
			return " " + strings.Join(options, " ")
		}
	}
	return ""
}

// generateInlineForeignKey 生成内联外键定义（用于SQLite）
func (g *DDLGenerator) generateInlineForeignKey(fk *ForeignKey) string {
	columns := make([]string, len(fk.Columns))
	for i, col := range fk.Columns {
		columns[i] = g.QuoteIdentifier(col)
	}

	refColumns := make([]string, len(fk.ReferencedColumns))
	for i, col := range fk.ReferencedColumns {
		refColumns[i] = g.QuoteIdentifier(col)
	}

	sql := fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s (%s)",
		strings.Join(columns, ", "),
		g.QuoteIdentifier(fk.ReferencedTable),
		strings.Join(refColumns, ", "))

	if fk.OnUpdate != "" {
		sql += " ON UPDATE " + fk.OnUpdate
	}
	if fk.OnDelete != "" {
		sql += " ON DELETE " + fk.OnDelete
	}

	return sql
}

// GenerateDropTableSQL 生成删除表的SQL
func (g *DDLGenerator) GenerateDropTableSQL(tableName string) string {
	return fmt.Sprintf("DROP TABLE IF EXISTS %s", g.QuoteIdentifier(tableName))
}

// GenerateAddColumnSQL 生成添加列的SQL
func (g *DDLGenerator) GenerateAddColumnSQL(tableName string, col *Column) string {
	colDef := g.generateColumnDefinition(col)
	return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", g.QuoteIdentifier(tableName), colDef)
}

// GenerateDropColumnSQL 生成删除列的SQL
func (g *DDLGenerator) GenerateDropColumnSQL(tableName, columnName string) string {
	if g.driver == "sqlite" || g.driver == "sqlite3" {
		return fmt.Sprintf("-- SQLite不支持直接删除列: %s.%s", tableName, columnName)
	}
	return fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s",
		g.QuoteIdentifier(tableName), g.QuoteIdentifier(columnName))
}

// GenerateModifyColumnSQL 生成修改列的SQL
func (g *DDLGenerator) GenerateModifyColumnSQL(tableName string, col *Column) string {
	colDef := g.generateColumnDefinition(col)

	switch g.driver {
	case "mysql":
		return fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s", g.QuoteIdentifier(tableName), colDef)
	case "postgres", "postgresql":
		// PostgreSQL需要分步修改
		colType := g.generateColumnType(col)
		return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s",
			g.QuoteIdentifier(tableName), g.QuoteIdentifier(col.Name), colType)
	case "sqlite", "sqlite3":
		return fmt.Sprintf("-- SQLite不支持直接修改列: %s.%s", tableName, col.Name)
	default:
		return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s", g.QuoteIdentifier(tableName), colDef)
	}
}

// GenerateRenameColumnSQL 生成重命名列的SQL
func (g *DDLGenerator) GenerateRenameColumnSQL(tableName, oldName, newName string) string {
	switch g.driver {
	case "mysql":
		// MySQL 8.0+
		return fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s",
			g.QuoteIdentifier(tableName), g.QuoteIdentifier(oldName), g.QuoteIdentifier(newName))
	case "postgres", "postgresql":
		return fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s",
			g.QuoteIdentifier(tableName), g.QuoteIdentifier(oldName), g.QuoteIdentifier(newName))
	case "sqlite", "sqlite3":
		// SQLite 3.25.0+
		return fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s",
			g.QuoteIdentifier(tableName), g.QuoteIdentifier(oldName), g.QuoteIdentifier(newName))
	case "sqlserver", "mssql":
		return fmt.Sprintf("EXEC sp_rename '%s.%s', '%s', 'COLUMN'", tableName, oldName, newName)
	default:
		return fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s",
			g.QuoteIdentifier(tableName), g.QuoteIdentifier(oldName), g.QuoteIdentifier(newName))
	}
}

// GenerateCreateIndexSQL 生成创建索引的SQL
func (g *DDLGenerator) GenerateCreateIndexSQL(tableName string, index *Index) string {
	indexType := ""
	if index.Unique {
		indexType = "UNIQUE "
	}

	columns := make([]string, len(index.Columns))
	for i, col := range index.Columns {
		columns[i] = g.QuoteIdentifier(col)
	}

	var sql string
	switch g.driver {
	case "mysql":
		if index.Type != "" && index.Type != "btree" {
			sql = fmt.Sprintf("CREATE %sINDEX %s ON %s (%s) USING %s",
				indexType, g.QuoteIdentifier(index.Name), g.QuoteIdentifier(tableName),
				strings.Join(columns, ", "), strings.ToUpper(index.Type))
		} else {
			sql = fmt.Sprintf("CREATE %sINDEX %s ON %s (%s)",
				indexType, g.QuoteIdentifier(index.Name), g.QuoteIdentifier(tableName),
				strings.Join(columns, ", "))
		}
	case "postgres", "postgresql":
		if index.Type != "" && index.Type != "btree" {
			sql = fmt.Sprintf("CREATE %sINDEX %s ON %s USING %s (%s)",
				indexType, g.QuoteIdentifier(index.Name), g.QuoteIdentifier(tableName),
				index.Type, strings.Join(columns, ", "))
		} else {
			sql = fmt.Sprintf("CREATE %sINDEX %s ON %s (%s)",
				indexType, g.QuoteIdentifier(index.Name), g.QuoteIdentifier(tableName),
				strings.Join(columns, ", "))
		}
	default:
		sql = fmt.Sprintf("CREATE %sINDEX %s ON %s (%s)",
			indexType, g.QuoteIdentifier(index.Name), g.QuoteIdentifier(tableName),
			strings.Join(columns, ", "))
	}

	return sql
}

// GenerateDropIndexSQL 生成删除索引的SQL
func (g *DDLGenerator) GenerateDropIndexSQL(tableName, indexName string) string {
	switch g.driver {
	case "mysql":
		return fmt.Sprintf("DROP INDEX %s ON %s", g.QuoteIdentifier(indexName), g.QuoteIdentifier(tableName))
	default:
		return fmt.Sprintf("DROP INDEX IF EXISTS %s", g.QuoteIdentifier(indexName))
	}
}

// GenerateAddForeignKeySQL 生成添加外键的SQL
func (g *DDLGenerator) GenerateAddForeignKeySQL(tableName string, fk *ForeignKey) string {
	if g.driver == "sqlite" || g.driver == "sqlite3" {
		return fmt.Sprintf("-- SQLite不支持在表创建后添加外键: %s", fk.Name)
	}

	columns := make([]string, len(fk.Columns))
	for i, col := range fk.Columns {
		columns[i] = g.QuoteIdentifier(col)
	}

	refColumns := make([]string, len(fk.ReferencedColumns))
	for i, col := range fk.ReferencedColumns {
		refColumns[i] = g.QuoteIdentifier(col)
	}

	sql := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)",
		g.QuoteIdentifier(tableName), g.QuoteIdentifier(fk.Name),
		strings.Join(columns, ", "), g.QuoteIdentifier(fk.ReferencedTable),
		strings.Join(refColumns, ", "))

	if fk.OnUpdate != "" {
		sql += " ON UPDATE " + fk.OnUpdate
	}
	if fk.OnDelete != "" {
		sql += " ON DELETE " + fk.OnDelete
	}

	return sql
}

// GenerateDropForeignKeySQL 生成删除外键的SQL
func (g *DDLGenerator) GenerateDropForeignKeySQL(tableName, fkName string) string {
	if g.driver == "sqlite" || g.driver == "sqlite3" {
		return fmt.Sprintf("-- SQLite不支持删除外键: %s", fkName)
	}

	switch g.driver {
	case "mysql":
		return fmt.Sprintf("ALTER TABLE %s DROP FOREIGN KEY %s",
			g.QuoteIdentifier(tableName), g.QuoteIdentifier(fkName))
	case "postgres", "postgresql":
		return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s",
			g.QuoteIdentifier(tableName), g.QuoteIdentifier(fkName))
	default:
		return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s",
			g.QuoteIdentifier(tableName), g.QuoteIdentifier(fkName))
	}
}

// GenerateRenameTableSQL 生成重命名表的SQL
func (g *DDLGenerator) GenerateRenameTableSQL(oldName, newName string) string {
	switch g.driver {
	case "mysql":
		return fmt.Sprintf("RENAME TABLE %s TO %s", g.QuoteIdentifier(oldName), g.QuoteIdentifier(newName))
	case "postgres", "postgresql", "sqlite", "sqlite3":
		return fmt.Sprintf("ALTER TABLE %s RENAME TO %s", g.QuoteIdentifier(oldName), g.QuoteIdentifier(newName))
	case "sqlserver", "mssql":
		return fmt.Sprintf("EXEC sp_rename '%s', '%s'", oldName, newName)
	default:
		return fmt.Sprintf("ALTER TABLE %s RENAME TO %s", g.QuoteIdentifier(oldName), g.QuoteIdentifier(newName))
	}
}

// GenerateTruncateTableSQL 生成清空表的SQL
func (g *DDLGenerator) GenerateTruncateTableSQL(tableName string) string {
	switch g.driver {
	case "sqlite", "sqlite3":
		// SQLite 不支持 TRUNCATE，使用 DELETE
		return fmt.Sprintf("DELETE FROM %s", g.QuoteIdentifier(tableName))
	default:
		return fmt.Sprintf("TRUNCATE TABLE %s", g.QuoteIdentifier(tableName))
	}
}

// GenerateTableExistsSQL 生成检查表是否存在的SQL
func (g *DDLGenerator) GenerateTableExistsSQL(tableName string) string {
	switch g.driver {
	case "mysql":
		return "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?"
	case "postgres", "postgresql":
		return "SELECT COUNT(*) FROM information_schema.tables WHERE table_name = $1"
	case "sqlite", "sqlite3":
		return "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?"
	case "sqlserver", "mssql":
		return "SELECT COUNT(*) FROM sys.tables WHERE name = ?"
	default:
		return "SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ?"
	}
}

// GenerateColumnExistsSQL 生成检查列是否存在的SQL
func (g *DDLGenerator) GenerateColumnExistsSQL(tableName string) string {
	switch g.driver {
	case "mysql":
		return "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?"
	case "postgres", "postgresql":
		return "SELECT COUNT(*) FROM information_schema.columns WHERE table_name = $1 AND column_name = $2"
	case "sqlite", "sqlite3":
		return fmt.Sprintf("SELECT COUNT(*) FROM pragma_table_info('%s') WHERE name = ?", tableName)
	case "sqlserver", "mssql":
		return "SELECT COUNT(*) FROM sys.columns WHERE object_id = OBJECT_ID(?) AND name = ?"
	default:
		return "SELECT COUNT(*) FROM information_schema.columns WHERE table_name = ? AND column_name = ?"
	}
}
