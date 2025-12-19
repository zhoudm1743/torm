package migration

import (
	"fmt"
	"strings"

	"github.com/zhoudm1743/torm/db"
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

	// PostgreSQL SERIAL类型
	ColumnTypeSerial      ColumnType = "SERIAL"
	ColumnTypeBigSerial   ColumnType = "BIGSERIAL"
	ColumnTypeSmallSerial ColumnType = "SMALLSERIAL"
)

// ModelColumn 模型列定义（用于分析器）
type ModelColumn struct {
	Name          string
	Type          ColumnType
	Length        int
	Precision     int
	Scale         int
	NotNull       bool
	PrimaryKey    bool
	AutoIncrement bool
	Default       interface{} // 默认值
	Comment       string
	GoType        string // Go语言中的类型
	Unique        bool   // 是否唯一

	// 扩展属性
	Unsigned  bool   // 无符号数字类型 (MySQL)
	Zerofill  bool   // 零填充 (MySQL)
	Binary    bool   // 二进制存储
	Encrypted bool   // 加密字段标记
	Hidden    bool   // 隐藏字段
	Readonly  bool   // 只读字段
	Generated string // 生成列类型: "virtual", "stored", ""

	// 索引相关
	Index         bool   // 普通索引
	FulltextIndex bool   // 全文索引
	SpatialIndex  bool   // 空间索引
	IndexType     string // 索引类型: "btree", "hash", "rtree"

	// 外键相关
	ForeignKey string // 外键表.字段
	OnDelete   string // 删除时动作: "cascade", "restrict", "set null", "set default"
	OnUpdate   string // 更新时动作: "cascade", "restrict", "set null", "set default"

	// 时间管理
	AutoCreateTime bool // 自动创建时间字段
	AutoUpdateTime bool // 自动更新时间字段

	// 字段权限（用于读写控制）
	// 支持的torm标签格式:
	// `torm:"<-:create"` - 允许读和创建
	// `torm:"<-:update"` - 允许读和更新
	// `torm:"<-"`        - 允许读和写（创建和更新）
	// `torm:"<-:false"`  - 允许读，禁止写
	// `torm:"->"`        - 只读
	// `torm:"->;<-:create"` - 允许读和创建
	// `torm:"->:false;<-:create"` - 仅创建（禁止从db读）
	// `torm:"-"`         - 通过struct读写会忽略该字段
	// `torm:"-:all"`     - 通过struct读写、迁移会忽略该字段
	// `torm:"-:migration"` - 通过struct迁移会忽略该字段
	CanRead         bool // 是否允许从数据库读取
	CanCreate       bool // 是否允许创建时写入
	CanUpdate       bool // 是否允许更新时写入
	IgnoreAll       bool // 完全忽略该字段（读写和迁移）
	IgnoreMigration bool // 迁移时忽略该字段
	IgnoreReadWrite bool // 读写时忽略该字段（但迁移时包含）
}

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
	conn   db.Connection
	driver string
}

// NewSchemaBuilder 创建结构构建器
func NewSchemaBuilder(conn db.Connection) *SchemaBuilder {
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

// GenerateIndexesFromModelColumns 从模型列生成索引定义
func (sb *SchemaBuilder) GenerateIndexesFromModelColumns(tableName string, columns []ModelColumn) []*Index {
	var indexes []*Index

	for _, col := range columns {
		// 普通索引
		if col.Index && !col.Unique && !col.PrimaryKey {
			indexName := fmt.Sprintf("idx_%s_%s", tableName, col.Name)
			indexes = append(indexes, &Index{
				Name:    indexName,
				Columns: []string{col.Name},
				Unique:  false,
				Type:    col.IndexType,
			})
		}

		// 唯一索引（非主键）
		if col.Unique && !col.PrimaryKey {
			indexName := fmt.Sprintf("idx_%s_%s_unique", tableName, col.Name)
			indexes = append(indexes, &Index{
				Name:    indexName,
				Columns: []string{col.Name},
				Unique:  true,
			})
		}
	}

	return indexes
}

// GenerateForeignKeysFromModelColumns 从模型列生成外键定义
func (sb *SchemaBuilder) GenerateForeignKeysFromModelColumns(tableName string, columns []ModelColumn) []*ForeignKey {
	var foreignKeys []*ForeignKey

	for _, col := range columns {
		if col.ForeignKey != "" {
			// 解析外键引用: "table.column" 或 "table(column)"
			refTable, refColumn := sb.parseForeignKeyReference(col.ForeignKey)
			if refTable != "" && refColumn != "" {
				fkName := fmt.Sprintf("fk_%s_%s", tableName, col.Name)
				fk := &ForeignKey{
					Name:              fkName,
					Columns:           []string{col.Name},
					ReferencedTable:   refTable,
					ReferencedColumns: []string{refColumn},
				}

				if col.OnDelete != "" {
					fk.OnDelete = strings.ToUpper(strings.ReplaceAll(col.OnDelete, "_", " "))
				}
				if col.OnUpdate != "" {
					fk.OnUpdate = strings.ToUpper(strings.ReplaceAll(col.OnUpdate, "_", " "))
				}

				foreignKeys = append(foreignKeys, fk)
			}
		}
	}

	return foreignKeys
}

// parseForeignKeyReference 解析外键引用
func (sb *SchemaBuilder) parseForeignKeyReference(reference string) (table, column string) {
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

// ConvertModelColumnsToColumns 将ModelColumn转换为Column
func (sb *SchemaBuilder) ConvertModelColumnsToColumns(modelColumns []ModelColumn) []*Column {
	var columns []*Column

	for _, mc := range modelColumns {
		// 跳过迁移忽略的字段
		if mc.IgnoreAll || mc.IgnoreMigration {
			continue
		}

		col := &Column{
			Name:          mc.Name,
			Type:          mc.Type,
			Length:        mc.Length,
			Precision:     mc.Precision,
			Scale:         mc.Scale,
			NotNull:       mc.NotNull,
			PrimaryKey:    mc.PrimaryKey,
			AutoIncrement: mc.AutoIncrement,
			Unique:        mc.Unique,
			Default:       mc.Default,
			Comment:       mc.Comment,
		}
		columns = append(columns, col)
	}

	return columns
}

// CreateTableFromModelColumns 从模型列创建表
func (sb *SchemaBuilder) CreateTableFromModelColumns(tableName string, modelColumns []ModelColumn) error {
	// 转换列定义
	columns := sb.ConvertModelColumnsToColumns(modelColumns)

	// 生成索引
	indexes := sb.GenerateIndexesFromModelColumns(tableName, modelColumns)

	// 生成外键
	foreignKeys := sb.GenerateForeignKeysFromModelColumns(tableName, modelColumns)

	// 创建表定义
	table := &Table{
		Name:        tableName,
		Columns:     columns,
		Indexes:     indexes,
		ForeignKeys: foreignKeys,
	}

	return sb.CreateTable(table)
}

// RenameColumn 重命名列
func (sb *SchemaBuilder) RenameColumn(tableName, oldName, newName string) error {
	var sql string

	switch sb.driver {
	case "mysql":
		// MySQL 8.0+ 支持 RENAME COLUMN
		sql = fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s",
			sb.quoteName(tableName), sb.quoteName(oldName), sb.quoteName(newName))
	case "postgres", "postgresql":
		sql = fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s",
			sb.quoteName(tableName), sb.quoteName(oldName), sb.quoteName(newName))
	case "sqlite", "sqlite3":
		// SQLite 3.25.0+ 支持 RENAME COLUMN
		sql = fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s",
			sb.quoteName(tableName), sb.quoteName(oldName), sb.quoteName(newName))
	case "sqlserver", "mssql":
		sql = fmt.Sprintf("EXEC sp_rename '%s.%s', '%s', 'COLUMN'",
			tableName, oldName, newName)
	default:
		return fmt.Errorf("unsupported driver for rename column: %s", sb.driver)
	}

	_, err := sb.conn.Exec(sql)
	if err != nil {
		return fmt.Errorf("failed to rename column %s to %s in table %s: %w",
			oldName, newName, tableName, err)
	}

	return nil
}

// HasTable 检查表是否存在
func (sb *SchemaBuilder) HasTable(tableName string) (bool, error) {
	var query string
	var args []interface{}

	switch sb.driver {
	case "mysql":
		query = "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?"
		args = []interface{}{tableName}
	case "postgres", "postgresql":
		query = "SELECT COUNT(*) FROM information_schema.tables WHERE table_name = $1"
		args = []interface{}{tableName}
	case "sqlite", "sqlite3":
		query = "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?"
		args = []interface{}{tableName}
	case "sqlserver", "mssql":
		query = "SELECT COUNT(*) FROM sys.tables WHERE name = ?"
		args = []interface{}{tableName}
	default:
		return false, fmt.Errorf("unsupported driver: %s", sb.driver)
	}

	row := sb.conn.QueryRow(query, args...)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}

// HasColumn 检查列是否存在
func (sb *SchemaBuilder) HasColumn(tableName, columnName string) (bool, error) {
	var query string
	var args []interface{}

	switch sb.driver {
	case "mysql":
		query = "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?"
		args = []interface{}{tableName, columnName}
	case "postgres", "postgresql":
		query = "SELECT COUNT(*) FROM information_schema.columns WHERE table_name = $1 AND column_name = $2"
		args = []interface{}{tableName, columnName}
	case "sqlite", "sqlite3":
		// SQLite 需要使用 PRAGMA
		query = fmt.Sprintf("SELECT COUNT(*) FROM pragma_table_info('%s') WHERE name = ?", tableName)
		args = []interface{}{columnName}
	case "sqlserver", "mssql":
		query = "SELECT COUNT(*) FROM sys.columns WHERE object_id = OBJECT_ID(?) AND name = ?"
		args = []interface{}{tableName, columnName}
	default:
		return false, fmt.Errorf("unsupported driver: %s", sb.driver)
	}

	row := sb.conn.QueryRow(query, args...)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}

// HasIndex 检查索引是否存在
func (sb *SchemaBuilder) HasIndex(tableName, indexName string) (bool, error) {
	var query string
	var args []interface{}

	switch sb.driver {
	case "mysql":
		query = "SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?"
		args = []interface{}{tableName, indexName}
	case "postgres", "postgresql":
		query = "SELECT COUNT(*) FROM pg_indexes WHERE tablename = $1 AND indexname = $2"
		args = []interface{}{tableName, indexName}
	case "sqlite", "sqlite3":
		query = "SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND tbl_name = ? AND name = ?"
		args = []interface{}{tableName, indexName}
	case "sqlserver", "mssql":
		query = "SELECT COUNT(*) FROM sys.indexes WHERE object_id = OBJECT_ID(?) AND name = ?"
		args = []interface{}{tableName, indexName}
	default:
		return false, fmt.Errorf("unsupported driver: %s", sb.driver)
	}

	row := sb.conn.QueryRow(query, args...)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}
