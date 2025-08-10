package migration

import (
	"fmt"
	"strings"

	"github.com/zhoudm1743/torm/pkg/db"
)

// ColumnType 列类型
type ColumnType string

const (
	// 数字类型
	ColumnTypeInt      ColumnType = "INT"
	ColumnTypeBigInt   ColumnType = "BIGINT"
	ColumnTypeSmallInt ColumnType = "SMALLINT"
	ColumnTypeTinyInt  ColumnType = "TINYINT"
	ColumnTypeFloat    ColumnType = "FLOAT"
	ColumnTypeDouble   ColumnType = "DOUBLE"
	ColumnTypeDecimal  ColumnType = "DECIMAL"

	// 字符串类型
	ColumnTypeVarchar  ColumnType = "VARCHAR"
	ColumnTypeChar     ColumnType = "CHAR"
	ColumnTypeText     ColumnType = "TEXT"
	ColumnTypeLongText ColumnType = "LONGTEXT"

	// 时间类型
	ColumnTypeDateTime  ColumnType = "DATETIME"
	ColumnTypeTimestamp ColumnType = "TIMESTAMP"
	ColumnTypeDate      ColumnType = "DATE"
	ColumnTypeTime      ColumnType = "TIME"

	// 其他类型
	ColumnTypeBoolean ColumnType = "BOOLEAN"
	ColumnTypeBlob    ColumnType = "BLOB"
	ColumnTypeJSON    ColumnType = "JSON"
)

// Column 列定义
type Column struct {
	Name          string
	Type          ColumnType
	Length        int
	Precision     int
	Scale         int
	NotNull       bool
	PrimaryKey    bool
	AutoIncrement bool
	Unique        bool
	Default       interface{}
	Comment       string
}

// Index 索引定义
type Index struct {
	Name    string
	Columns []string
	Unique  bool
	Type    string // 索引类型，如 BTREE, HASH
}

// ForeignKey 外键定义
type ForeignKey struct {
	Name              string
	Columns           []string
	ReferencedTable   string
	ReferencedColumns []string
	OnUpdate          string // CASCADE, RESTRICT, SET NULL, SET DEFAULT
	OnDelete          string
}

// Table 表定义
type Table struct {
	Name        string
	Columns     []*Column
	Indexes     []*Index
	ForeignKeys []*ForeignKey
	Engine      string // MySQL引擎
	Charset     string // 字符集
	Comment     string
}

// SchemaBuilder 结构构建器
type SchemaBuilder struct {
	conn   db.ConnectionInterface
	driver string
}

// NewSchemaBuilder 创建结构构建器
func NewSchemaBuilder(conn db.ConnectionInterface) *SchemaBuilder {
	return &SchemaBuilder{
		conn:   conn,
		driver: conn.GetDriver(),
	}
}

// CreateTable 创建表
func (sb *SchemaBuilder) CreateTable(table *Table) error {
	sql, err := sb.generateCreateTableSQL(table)
	if err != nil {
		return err
	}

	_, err = sb.conn.Exec(sql)
	if err != nil {
		return fmt.Errorf("failed to create table %s: %w", table.Name, err)
	}

	// 创建索引
	for _, index := range table.Indexes {
		if err := sb.CreateIndex(table.Name, index); err != nil {
			return fmt.Errorf("failed to create index %s: %w", index.Name, err)
		}
	}

	// 创建外键
	for _, fk := range table.ForeignKeys {
		if err := sb.CreateForeignKey(table.Name, fk); err != nil {
			return fmt.Errorf("failed to create foreign key %s: %w", fk.Name, err)
		}
	}

	return nil
}

// DropTable 删除表
func (sb *SchemaBuilder) DropTable(tableName string) error {
	sql := fmt.Sprintf("DROP TABLE IF EXISTS %s", sb.quoteName(tableName))
	_, err := sb.conn.Exec(sql)
	if err != nil {
		return fmt.Errorf("failed to drop table %s: %w", tableName, err)
	}
	return nil
}

// AddColumn 添加列
func (sb *SchemaBuilder) AddColumn(tableName string, column *Column) error {
	sql, err := sb.generateAddColumnSQL(tableName, column)
	if err != nil {
		return err
	}

	_, err = sb.conn.Exec(sql)
	if err != nil {
		return fmt.Errorf("failed to add column %s to table %s: %w", column.Name, tableName, err)
	}
	return nil
}

// DropColumn 删除列
func (sb *SchemaBuilder) DropColumn(tableName, columnName string) error {
	sql := sb.generateDropColumnSQL(tableName, columnName)
	_, err := sb.conn.Exec(sql)
	if err != nil {
		return fmt.Errorf("failed to drop column %s from table %s: %w", columnName, tableName, err)
	}
	return nil
}

// ModifyColumn 修改列
func (sb *SchemaBuilder) ModifyColumn(tableName string, column *Column) error {
	sql, err := sb.generateModifyColumnSQL(tableName, column)
	if err != nil {
		return err
	}

	_, err = sb.conn.Exec(sql)
	if err != nil {
		return fmt.Errorf("failed to modify column %s in table %s: %w", column.Name, tableName, err)
	}
	return nil
}

// CreateIndex 创建索引
func (sb *SchemaBuilder) CreateIndex(tableName string, index *Index) error {
	sql := sb.generateCreateIndexSQL(tableName, index)
	_, err := sb.conn.Exec(sql)
	if err != nil {
		return fmt.Errorf("failed to create index %s: %w", index.Name, err)
	}
	return nil
}

// DropIndex 删除索引
func (sb *SchemaBuilder) DropIndex(tableName, indexName string) error {
	sql := sb.generateDropIndexSQL(tableName, indexName)
	_, err := sb.conn.Exec(sql)
	if err != nil {
		return fmt.Errorf("failed to drop index %s: %w", indexName, err)
	}
	return nil
}

// CreateForeignKey 创建外键
func (sb *SchemaBuilder) CreateForeignKey(tableName string, fk *ForeignKey) error {
	if sb.driver == "sqlite" {
		// SQLite 在表创建后不支持添加外键
		return nil
	}

	sql := sb.generateCreateForeignKeySQL(tableName, fk)
	_, err := sb.conn.Exec(sql)
	if err != nil {
		return fmt.Errorf("failed to create foreign key %s: %w", fk.Name, err)
	}
	return nil
}

// DropForeignKey 删除外键
func (sb *SchemaBuilder) DropForeignKey(tableName, fkName string) error {
	if sb.driver == "sqlite" {
		// SQLite 不支持删除外键
		return nil
	}

	sql := sb.generateDropForeignKeySQL(tableName, fkName)
	_, err := sb.conn.Exec(sql)
	if err != nil {
		return fmt.Errorf("failed to drop foreign key %s: %w", fkName, err)
	}
	return nil
}

// generateCreateTableSQL 生成创建表的SQL
func (sb *SchemaBuilder) generateCreateTableSQL(table *Table) (string, error) {
	var parts []string
	var primaryKeys []string

	// 列定义
	for _, column := range table.Columns {
		columnSQL, err := sb.generateColumnSQL(column)
		if err != nil {
			return "", err
		}
		parts = append(parts, columnSQL)

		if column.PrimaryKey {
			primaryKeys = append(primaryKeys, sb.quoteName(column.Name))
		}
	}

	// 主键定义
	if len(primaryKeys) > 0 {
		parts = append(parts, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(primaryKeys, ", ")))
	}

	// 外键定义（对于SQLite在创建表时添加）
	if sb.driver == "sqlite" {
		for _, fk := range table.ForeignKeys {
			fkSQL := sb.generateInlineForeignKeySQL(fk)
			if fkSQL != "" {
				parts = append(parts, fkSQL)
			}
		}
	}

	// 构建完整的CREATE TABLE语句
	sql := fmt.Sprintf("CREATE TABLE %s (\n  %s\n)",
		sb.quoteName(table.Name),
		strings.Join(parts, ",\n  "))

	// 添加表选项
	switch sb.driver {
	case "mysql":
		options := []string{}
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
			sql += " " + strings.Join(options, " ")
		}
	}

	return sql, nil
}

// generateColumnSQL 生成列的SQL
func (sb *SchemaBuilder) generateColumnSQL(column *Column) (string, error) {
	parts := []string{sb.quoteName(column.Name)}

	// 类型定义
	typeSQL, err := sb.generateColumnTypeSQL(column)
	if err != nil {
		return "", err
	}
	parts = append(parts, typeSQL)

	// NOT NULL
	if column.NotNull {
		parts = append(parts, "NOT NULL")
	}

	// AUTO_INCREMENT
	if column.AutoIncrement {
		switch sb.driver {
		case "mysql":
			parts = append(parts, "AUTO_INCREMENT")
		case "postgres", "postgresql":
			// PostgreSQL 使用 SERIAL 或 BIGSERIAL
		case "sqlite", "sqlite3":
			// SQLite 的 INTEGER PRIMARY KEY 自动递增
		case "sqlserver", "mssql":
			parts = append(parts, "IDENTITY(1,1)")
		}
	}

	// UNIQUE
	if column.Unique {
		parts = append(parts, "UNIQUE")
	}

	// DEFAULT
	if column.Default != nil {
		defaultSQL := sb.generateDefaultSQL(column.Default)
		if defaultSQL != "" {
			parts = append(parts, "DEFAULT", defaultSQL)
		}
	}

	// COMMENT
	if column.Comment != "" && (sb.driver == "mysql") {
		parts = append(parts, fmt.Sprintf("COMMENT '%s'", strings.ReplaceAll(column.Comment, "'", "''")))
	}

	return strings.Join(parts, " "), nil
}

// generateColumnTypeSQL 生成列类型SQL
func (sb *SchemaBuilder) generateColumnTypeSQL(column *Column) (string, error) {
	switch sb.driver {
	case "mysql":
		return sb.generateMySQLColumnType(column), nil
	case "postgres", "postgresql":
		return sb.generatePostgreSQLColumnType(column), nil
	case "sqlite", "sqlite3":
		return sb.generateSQLiteColumnType(column), nil
	case "sqlserver", "mssql":
		return sb.generateSQLServerColumnType(column), nil
	default:
		return "", fmt.Errorf("unsupported driver: %s", sb.driver)
	}
}

// generateMySQLColumnType 生成MySQL列类型
func (sb *SchemaBuilder) generateMySQLColumnType(column *Column) string {
	switch column.Type {
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
		if column.Precision > 0 && column.Scale > 0 {
			return fmt.Sprintf("DECIMAL(%d,%d)", column.Precision, column.Scale)
		}
		return "DECIMAL"
	case ColumnTypeVarchar:
		length := column.Length
		if length <= 0 {
			length = 255
		}
		return fmt.Sprintf("VARCHAR(%d)", length)
	case ColumnTypeChar:
		length := column.Length
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
		return "BOOLEAN"
	case ColumnTypeBlob:
		return "BLOB"
	case ColumnTypeJSON:
		return "JSON"
	default:
		return string(column.Type)
	}
}

// generatePostgreSQLColumnType 生成PostgreSQL列类型
func (sb *SchemaBuilder) generatePostgreSQLColumnType(column *Column) string {
	switch column.Type {
	case ColumnTypeInt:
		return "INTEGER"
	case ColumnTypeBigInt:
		if column.AutoIncrement {
			return "BIGSERIAL"
		}
		return "BIGINT"
	case ColumnTypeSmallInt:
		return "SMALLINT"
	case ColumnTypeTinyInt:
		return "SMALLINT"
	case ColumnTypeFloat:
		return "REAL"
	case ColumnTypeDouble:
		return "DOUBLE PRECISION"
	case ColumnTypeDecimal:
		if column.Precision > 0 && column.Scale > 0 {
			return fmt.Sprintf("DECIMAL(%d,%d)", column.Precision, column.Scale)
		}
		return "DECIMAL"
	case ColumnTypeVarchar:
		length := column.Length
		if length <= 0 {
			length = 255
		}
		return fmt.Sprintf("VARCHAR(%d)", length)
	case ColumnTypeChar:
		length := column.Length
		if length <= 0 {
			length = 1
		}
		return fmt.Sprintf("CHAR(%d)", length)
	case ColumnTypeText:
		return "TEXT"
	case ColumnTypeLongText:
		return "TEXT"
	case ColumnTypeDateTime:
		return "TIMESTAMP"
	case ColumnTypeTimestamp:
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
		return string(column.Type)
	}
}

// generateSQLiteColumnType 生成SQLite列类型
func (sb *SchemaBuilder) generateSQLiteColumnType(column *Column) string {
	switch column.Type {
	case ColumnTypeInt, ColumnTypeBigInt, ColumnTypeSmallInt, ColumnTypeTinyInt:
		return "INTEGER"
	case ColumnTypeFloat, ColumnTypeDouble:
		return "REAL"
	case ColumnTypeDecimal:
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
func (sb *SchemaBuilder) generateSQLServerColumnType(column *Column) string {
	switch column.Type {
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
		return "FLOAT"
	case ColumnTypeDecimal:
		if column.Precision > 0 && column.Scale > 0 {
			return fmt.Sprintf("DECIMAL(%d,%d)", column.Precision, column.Scale)
		}
		return "DECIMAL"
	case ColumnTypeVarchar:
		length := column.Length
		if length <= 0 {
			length = 255
		}
		return fmt.Sprintf("NVARCHAR(%d)", length)
	case ColumnTypeChar:
		length := column.Length
		if length <= 0 {
			length = 1
		}
		return fmt.Sprintf("NCHAR(%d)", length)
	case ColumnTypeText:
		return "NTEXT"
	case ColumnTypeLongText:
		return "NTEXT"
	case ColumnTypeDateTime:
		return "DATETIME2"
	case ColumnTypeTimestamp:
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
		return string(column.Type)
	}
}

// generateDefaultSQL 生成默认值SQL
func (sb *SchemaBuilder) generateDefaultSQL(value interface{}) string {
	switch v := value.(type) {
	case string:
		if v == "CURRENT_TIMESTAMP" || v == "NOW()" {
			return v
		}
		return fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%v", v)
	case float32, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		if sb.driver == "postgres" || sb.driver == "postgresql" {
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

// 其他辅助方法继续...
func (sb *SchemaBuilder) generateAddColumnSQL(tableName string, column *Column) (string, error) {
	columnSQL, err := sb.generateColumnSQL(column)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", sb.quoteName(tableName), columnSQL), nil
}

func (sb *SchemaBuilder) generateDropColumnSQL(tableName, columnName string) string {
	return fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", sb.quoteName(tableName), sb.quoteName(columnName))
}

func (sb *SchemaBuilder) generateModifyColumnSQL(tableName string, column *Column) (string, error) {
	columnSQL, err := sb.generateColumnSQL(column)
	if err != nil {
		return "", err
	}

	switch sb.driver {
	case "mysql":
		return fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s", sb.quoteName(tableName), columnSQL), nil
	case "postgres", "postgresql":
		// PostgreSQL 需要分步修改
		return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s",
			sb.quoteName(tableName), sb.quoteName(column.Name), columnSQL), nil
	case "sqlite", "sqlite3":
		// SQLite 不支持修改列，需要重建表
		return "", fmt.Errorf("SQLite does not support modifying columns")
	default:
		return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s", sb.quoteName(tableName), columnSQL), nil
	}
}

func (sb *SchemaBuilder) generateCreateIndexSQL(tableName string, index *Index) string {
	indexType := ""
	if index.Unique {
		indexType = "UNIQUE "
	}

	columns := make([]string, len(index.Columns))
	for i, col := range index.Columns {
		columns[i] = sb.quoteName(col)
	}

	return fmt.Sprintf("CREATE %sINDEX %s ON %s (%s)",
		indexType, sb.quoteName(index.Name), sb.quoteName(tableName), strings.Join(columns, ", "))
}

func (sb *SchemaBuilder) generateDropIndexSQL(tableName, indexName string) string {
	switch sb.driver {
	case "mysql":
		return fmt.Sprintf("DROP INDEX %s ON %s", sb.quoteName(indexName), sb.quoteName(tableName))
	default:
		return fmt.Sprintf("DROP INDEX %s", sb.quoteName(indexName))
	}
}

func (sb *SchemaBuilder) generateCreateForeignKeySQL(tableName string, fk *ForeignKey) string {
	columns := make([]string, len(fk.Columns))
	for i, col := range fk.Columns {
		columns[i] = sb.quoteName(col)
	}

	refColumns := make([]string, len(fk.ReferencedColumns))
	for i, col := range fk.ReferencedColumns {
		refColumns[i] = sb.quoteName(col)
	}

	sql := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)",
		sb.quoteName(tableName), sb.quoteName(fk.Name),
		strings.Join(columns, ", "), sb.quoteName(fk.ReferencedTable), strings.Join(refColumns, ", "))

	if fk.OnUpdate != "" {
		sql += " ON UPDATE " + fk.OnUpdate
	}
	if fk.OnDelete != "" {
		sql += " ON DELETE " + fk.OnDelete
	}

	return sql
}

func (sb *SchemaBuilder) generateDropForeignKeySQL(tableName, fkName string) string {
	return fmt.Sprintf("ALTER TABLE %s DROP FOREIGN KEY %s", sb.quoteName(tableName), sb.quoteName(fkName))
}

func (sb *SchemaBuilder) generateInlineForeignKeySQL(fk *ForeignKey) string {
	columns := make([]string, len(fk.Columns))
	for i, col := range fk.Columns {
		columns[i] = sb.quoteName(col)
	}

	refColumns := make([]string, len(fk.ReferencedColumns))
	for i, col := range fk.ReferencedColumns {
		refColumns[i] = sb.quoteName(col)
	}

	sql := fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s (%s)",
		strings.Join(columns, ", "), sb.quoteName(fk.ReferencedTable), strings.Join(refColumns, ", "))

	if fk.OnUpdate != "" {
		sql += " ON UPDATE " + fk.OnUpdate
	}
	if fk.OnDelete != "" {
		sql += " ON DELETE " + fk.OnDelete
	}

	return sql
}

// quoteName 引用名称
func (sb *SchemaBuilder) quoteName(name string) string {
	switch sb.driver {
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
