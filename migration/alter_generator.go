package migration

import (
	"fmt"
	"strings"

	"github.com/zhoudm1743/torm/db"
)

// AlterGenerator ALTER TABLE SQL生成器
type AlterGenerator struct {
	driver string
}

// NewAlterGenerator 创建ALTER TABLE SQL生成器
func NewAlterGenerator(conn db.ConnectionInterface) *AlterGenerator {
	return &AlterGenerator{
		driver: conn.GetDriver(),
	}
}

// GenerateAlterSQL 生成ALTER TABLE SQL语句
func (ag *AlterGenerator) GenerateAlterSQL(tableName string, differences []ColumnDifference) ([]string, error) {
	var alterStatements []string

	switch ag.driver {
	case "mysql":
		return ag.generateMySQLAlterSQL(tableName, differences)
	case "postgres", "postgresql":
		return ag.generatePostgreSQLAlterSQL(tableName, differences)
	case "sqlite", "sqlite3":
		return ag.generateSQLiteAlterSQL(tableName, differences)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", ag.driver)
	}

	return alterStatements, nil
}

// generateMySQLAlterSQL 生成MySQL的ALTER TABLE语句
func (ag *AlterGenerator) generateMySQLAlterSQL(tableName string, differences []ColumnDifference) ([]string, error) {
	if len(differences) == 0 {
		return nil, nil
	}

	var modifications []string

	for _, diff := range differences {
		switch diff.Type {
		case "add":
			modelCol := diff.NewValue.(ModelColumn)
			colDef := ag.buildMySQLColumnDefinition(modelCol)
			modifications = append(modifications, fmt.Sprintf("ADD COLUMN %s %s", modelCol.Name, colDef))

		case "modify":
			modelCol := diff.NewValue.(ModelColumn)
			colDef := ag.buildMySQLColumnDefinition(modelCol)
			modifications = append(modifications, fmt.Sprintf("MODIFY COLUMN %s %s", modelCol.Name, colDef))

		case "drop":
			modifications = append(modifications, fmt.Sprintf("DROP COLUMN %s", diff.Column))
		}
	}

	if len(modifications) == 0 {
		return nil, nil
	}

	// MySQL支持在一个ALTER TABLE语句中进行多个修改
	alterSQL := fmt.Sprintf("ALTER TABLE %s %s", tableName, strings.Join(modifications, ", "))
	return []string{alterSQL}, nil
}

// generatePostgreSQLAlterSQL 生成PostgreSQL的ALTER TABLE语句
func (ag *AlterGenerator) generatePostgreSQLAlterSQL(tableName string, differences []ColumnDifference) ([]string, error) {
	var statements []string

	for _, diff := range differences {
		switch diff.Type {
		case "add":
			modelCol := diff.NewValue.(ModelColumn)
			colDef := ag.buildPostgreSQLColumnDefinition(modelCol)
			statement := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, modelCol.Name, colDef)
			statements = append(statements, statement)

		case "modify":
			modelCol := diff.NewValue.(ModelColumn)
			dbCol := diff.OldValue.(DatabaseColumn)

			// PostgreSQL需要分别处理类型、默认值、NOT NULL等
			modifyStatements := ag.generatePostgreSQLModifyStatements(tableName, modelCol, dbCol)
			statements = append(statements, modifyStatements...)

		case "drop":
			statement := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", tableName, diff.Column)
			statements = append(statements, statement)
		}
	}

	return statements, nil
}

// generatePostgreSQLModifyStatements 生成PostgreSQL的列修改语句
func (ag *AlterGenerator) generatePostgreSQLModifyStatements(tableName string, modelCol ModelColumn, dbCol DatabaseColumn) []string {
	var statements []string

	// 修改类型
	newType := ag.modelTypeToPostgreSQL(modelCol.Type)
	if !strings.EqualFold(dbCol.Type, newType) {
		typeClause := newType
		if modelCol.Length > 0 && ag.needsLength(modelCol.Type) {
			typeClause = fmt.Sprintf("%s(%d)", newType, modelCol.Length)
		} else if modelCol.Precision > 0 && modelCol.Type == ColumnTypeDecimal {
			if modelCol.Scale > 0 {
				typeClause = fmt.Sprintf("%s(%d,%d)", newType, modelCol.Precision, modelCol.Scale)
			} else {
				typeClause = fmt.Sprintf("%s(%d)", newType, modelCol.Precision)
			}
		}
		statement := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s", tableName, modelCol.Name, typeClause)
		statements = append(statements, statement)
	}

	// 修改默认值
	if !ag.defaultsEqual(dbCol.Default, modelCol.Default) {
		if modelCol.Default != nil {
			statement := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s", tableName, modelCol.Name, *modelCol.Default)
			statements = append(statements, statement)
		} else {
			statement := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT", tableName, modelCol.Name)
			statements = append(statements, statement)
		}
	}

	// 修改NOT NULL约束
	if dbCol.NotNull != modelCol.NotNull {
		if modelCol.NotNull {
			statement := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET NOT NULL", tableName, modelCol.Name)
			statements = append(statements, statement)
		} else {
			statement := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL", tableName, modelCol.Name)
			statements = append(statements, statement)
		}
	}

	return statements
}

// generateSQLiteAlterSQL 生成SQLite的ALTER TABLE语句
func (ag *AlterGenerator) generateSQLiteAlterSQL(tableName string, differences []ColumnDifference) ([]string, error) {
	// SQLite对ALTER TABLE的支持有限，复杂的修改需要重建表
	var statements []string
	needsRecreate := false

	var addColumns []ModelColumn
	var dropColumns []string

	for _, diff := range differences {
		switch diff.Type {
		case "add":
			modelCol := diff.NewValue.(ModelColumn)
			addColumns = append(addColumns, modelCol)

		case "modify":
			// SQLite不支持直接修改列，需要重建表
			needsRecreate = true

		case "drop":
			// SQLite不支持直接删除列，需要重建表
			needsRecreate = true
			dropColumns = append(dropColumns, diff.Column)
		}
	}

	// 处理简单的添加列操作
	for _, col := range addColumns {
		if !needsRecreate {
			colDef := ag.buildSQLiteColumnDefinition(col)
			statement := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, col.Name, colDef)
			statements = append(statements, statement)
		}
	}

	// 如果需要重建表，生成重建表的语句序列
	if needsRecreate {
		recreateStatements := ag.generateSQLiteRecreateStatements(tableName, differences)
		statements = append(statements, recreateStatements...)
	}

	return statements, nil
}

// generateSQLiteRecreateStatements 生成SQLite重建表的语句序列
func (ag *AlterGenerator) generateSQLiteRecreateStatements(tableName string, differences []ColumnDifference) []string {
	var statements []string

	// 这是一个复杂的过程，需要：
	// 1. 创建新表
	// 2. 复制数据
	// 3. 删除旧表
	// 4. 重命名新表

	// 注意：这里简化处理，实际使用时需要更复杂的逻辑
	statements = append(statements,
		fmt.Sprintf("-- WARNING: SQLite table %s requires recreation for complex changes", tableName),
		fmt.Sprintf("-- Please use migration system for complex SQLite schema changes"),
	)

	return statements
}

// buildMySQLColumnDefinition 构建MySQL列定义
func (ag *AlterGenerator) buildMySQLColumnDefinition(col ModelColumn) string {
	var parts []string

	// 类型
	typeDef := ag.modelTypeToMySQL(col.Type)

	// 长度
	if col.Length > 0 && ag.needsLength(col.Type) {
		typeDef = fmt.Sprintf("%s(%d)", typeDef, col.Length)
	} else if col.Precision > 0 && col.Type == ColumnTypeDecimal {
		if col.Scale > 0 {
			typeDef = fmt.Sprintf("%s(%d,%d)", typeDef, col.Precision, col.Scale)
		} else {
			typeDef = fmt.Sprintf("%s(%d)", typeDef, col.Precision)
		}
	}

	parts = append(parts, typeDef)

	// NOT NULL
	if col.NotNull {
		parts = append(parts, "NOT NULL")
	}

	// 默认值
	if col.Default != nil {
		parts = append(parts, fmt.Sprintf("DEFAULT %s", *col.Default))
	}

	// AUTO_INCREMENT
	if col.AutoIncrement {
		parts = append(parts, "AUTO_INCREMENT")
	}

	// 注释
	if col.Comment != "" {
		parts = append(parts, fmt.Sprintf("COMMENT '%s'", strings.ReplaceAll(col.Comment, "'", "''")))
	}

	return strings.Join(parts, " ")
}

// buildPostgreSQLColumnDefinition 构建PostgreSQL列定义
func (ag *AlterGenerator) buildPostgreSQLColumnDefinition(col ModelColumn) string {
	var parts []string

	// 类型
	typeDef := ag.modelTypeToPostgreSQL(col.Type)

	// 长度
	if col.Length > 0 && ag.needsLength(col.Type) {
		typeDef = fmt.Sprintf("%s(%d)", typeDef, col.Length)
	} else if col.Precision > 0 && col.Type == ColumnTypeDecimal {
		if col.Scale > 0 {
			typeDef = fmt.Sprintf("%s(%d,%d)", typeDef, col.Precision, col.Scale)
		} else {
			typeDef = fmt.Sprintf("%s(%d)", typeDef, col.Precision)
		}
	}

	parts = append(parts, typeDef)

	// NOT NULL
	if col.NotNull {
		parts = append(parts, "NOT NULL")
	}

	// 默认值
	if col.Default != nil {
		parts = append(parts, fmt.Sprintf("DEFAULT %s", *col.Default))
	}

	return strings.Join(parts, " ")
}

// buildSQLiteColumnDefinition 构建SQLite列定义
func (ag *AlterGenerator) buildSQLiteColumnDefinition(col ModelColumn) string {
	var parts []string

	// SQLite类型
	typeDef := ag.modelTypeToSQLite(col.Type)

	// SQLite中某些类型可以有长度（虽然不强制）
	if col.Length > 0 && (col.Type == ColumnTypeVarchar || col.Type == ColumnTypeChar) {
		typeDef = fmt.Sprintf("%s(%d)", typeDef, col.Length)
	}

	parts = append(parts, typeDef)

	// NOT NULL
	if col.NotNull {
		parts = append(parts, "NOT NULL")
	}

	// 默认值
	if col.Default != nil {
		parts = append(parts, fmt.Sprintf("DEFAULT %s", *col.Default))
	}

	return strings.Join(parts, " ")
}

// needsLength 检查类型是否需要长度
func (ag *AlterGenerator) needsLength(colType ColumnType) bool {
	return colType == ColumnTypeVarchar || colType == ColumnTypeChar
}

// 类型转换方法（复用SchemaComparator中的方法）
func (ag *AlterGenerator) modelTypeToMySQL(modelType ColumnType) string {
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

func (ag *AlterGenerator) modelTypeToPostgreSQL(modelType ColumnType) string {
	switch modelType {
	case ColumnTypeVarchar:
		return "VARCHAR"
	case ColumnTypeChar:
		return "CHAR"
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

func (ag *AlterGenerator) modelTypeToSQLite(modelType ColumnType) string {
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

// defaultsEqual 比较默认值是否相等（复用）
func (ag *AlterGenerator) defaultsEqual(dbDefault *string, modelDefault *string) bool {
	if dbDefault == nil && modelDefault == nil {
		return true
	}
	if dbDefault == nil || modelDefault == nil {
		return false
	}
	return *dbDefault == *modelDefault
}
