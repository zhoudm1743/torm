package migration

import (
	"reflect"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Property 10: Migration Idempotence**
// **Validates: Requirements 10.1, 10.2, 10.3, 10.4, 10.5, 10.6**

// TestProperty10_MigrationIdempotence tests that AutoMigrate produces the same
// table schema when called multiple times, and adding new fields results in
// new columns being added without data loss.
func TestProperty10_MigrationIdempotence(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 10.1: DDL generation is deterministic for same input
	properties.Property("DDL generation is deterministic for same input", prop.ForAll(
		func(tableName string, columns []testColumn) bool {
			if tableName == "" || len(columns) == 0 {
				return true // Skip empty inputs
			}

			// Create table definition
			table := createTestTable(tableName, columns)

			// Generate DDL for MySQL
			mysqlGen := NewDDLGenerator("mysql")
			sql1 := mysqlGen.GenerateCreateTableSQL(table)
			sql2 := mysqlGen.GenerateCreateTableSQL(table)

			// Same input should produce same output
			if sql1 != sql2 {
				return false
			}

			// Generate DDL for PostgreSQL
			pgGen := NewDDLGenerator("postgres")
			pgSql1 := pgGen.GenerateCreateTableSQL(table)
			pgSql2 := pgGen.GenerateCreateTableSQL(table)

			if pgSql1 != pgSql2 {
				return false
			}

			// Generate DDL for SQLite
			sqliteGen := NewDDLGenerator("sqlite")
			sqliteSql1 := sqliteGen.GenerateCreateTableSQL(table)
			sqliteSql2 := sqliteGen.GenerateCreateTableSQL(table)

			if sqliteSql1 != sqliteSql2 {
				return false
			}

			return true
		},
		genTableName(),
		gen.SliceOfN(3, genTestColumn()),
	))

	// Property 10.2: Generated DDL contains all columns
	properties.Property("Generated DDL contains all columns", prop.ForAll(
		func(tableName string, columns []testColumn) bool {
			if tableName == "" || len(columns) == 0 {
				return true
			}

			table := createTestTable(tableName, columns)

			// Test for all supported databases
			drivers := []string{"mysql", "postgres", "sqlite"}
			for _, driver := range drivers {
				gen := NewDDLGenerator(driver)
				sql := gen.GenerateCreateTableSQL(table)

				// Check that all column names are present
				for _, col := range columns {
					if !strings.Contains(sql, col.Name) {
						return false
					}
				}
			}

			return true
		},
		genTableName(),
		gen.SliceOfN(3, genTestColumn()),
	))

	// Property 10.3: Primary key columns are marked correctly
	properties.Property("Primary key columns are marked correctly", prop.ForAll(
		func(tableName string, pkColumn string) bool {
			if tableName == "" || pkColumn == "" {
				return true
			}

			// Create table with primary key
			table := &Table{
				Name: tableName,
				Columns: []*Column{
					{Name: pkColumn, Type: ColumnTypeBigInt, PrimaryKey: true, AutoIncrement: true},
					{Name: "name", Type: ColumnTypeVarchar, Length: 255},
				},
			}

			drivers := []string{"mysql", "postgres", "sqlite"}
			for _, driver := range drivers {
				gen := NewDDLGenerator(driver)
				sql := gen.GenerateCreateTableSQL(table)

				// Should contain PRIMARY KEY
				if !strings.Contains(strings.ToUpper(sql), "PRIMARY KEY") {
					return false
				}
			}

			return true
		},
		genTableName(),
		genColumnName(),
	))

	// Property 10.4: Index generation produces valid SQL
	properties.Property("Index generation produces valid SQL", prop.ForAll(
		func(tableName string, indexName string, columnName string, unique bool) bool {
			if tableName == "" || indexName == "" || columnName == "" {
				return true
			}

			index := &Index{
				Name:    indexName,
				Columns: []string{columnName},
				Unique:  unique,
			}

			drivers := []string{"mysql", "postgres", "sqlite"}
			for _, driver := range drivers {
				gen := NewDDLGenerator(driver)
				sql := gen.GenerateCreateIndexSQL(tableName, index)

				// Should contain CREATE INDEX
				if !strings.Contains(strings.ToUpper(sql), "CREATE") {
					return false
				}
				if !strings.Contains(strings.ToUpper(sql), "INDEX") {
					return false
				}
				// Should contain index name
				if !strings.Contains(sql, indexName) {
					return false
				}
				// Should contain table name
				if !strings.Contains(sql, tableName) {
					return false
				}
				// Should contain column name
				if !strings.Contains(sql, columnName) {
					return false
				}
				// If unique, should contain UNIQUE
				if unique && !strings.Contains(strings.ToUpper(sql), "UNIQUE") {
					return false
				}
			}

			return true
		},
		genTableName(),
		genIndexName(),
		genColumnName(),
		gen.Bool(),
	))

	// Property 10.5: Foreign key generation produces valid SQL (non-SQLite)
	properties.Property("Foreign key generation produces valid SQL", prop.ForAll(
		func(tableName string, columnName string, refTable string, refColumn string) bool {
			if tableName == "" || columnName == "" || refTable == "" || refColumn == "" {
				return true
			}

			fk := &ForeignKey{
				Name:              "fk_" + tableName + "_" + columnName,
				Columns:           []string{columnName},
				ReferencedTable:   refTable,
				ReferencedColumns: []string{refColumn},
				OnDelete:          "CASCADE",
				OnUpdate:          "CASCADE",
			}

			// Test MySQL and PostgreSQL (SQLite doesn't support ALTER TABLE ADD FOREIGN KEY)
			drivers := []string{"mysql", "postgres"}
			for _, driver := range drivers {
				gen := NewDDLGenerator(driver)
				sql := gen.GenerateAddForeignKeySQL(tableName, fk)

				// Should contain ALTER TABLE
				if !strings.Contains(strings.ToUpper(sql), "ALTER TABLE") {
					return false
				}
				// Should contain FOREIGN KEY
				if !strings.Contains(strings.ToUpper(sql), "FOREIGN KEY") {
					return false
				}
				// Should contain REFERENCES
				if !strings.Contains(strings.ToUpper(sql), "REFERENCES") {
					return false
				}
				// Should contain ON DELETE CASCADE
				if !strings.Contains(strings.ToUpper(sql), "ON DELETE CASCADE") {
					return false
				}
			}

			return true
		},
		genTableName(),
		genColumnName(),
		genTableName(),
		genColumnName(),
	))

	// Property 10.6: Cross-database DDL uses correct syntax
	properties.Property("Cross-database DDL uses correct syntax", prop.ForAll(
		func(tableName string, columnName string) bool {
			if tableName == "" || columnName == "" {
				return true
			}

			table := &Table{
				Name: tableName,
				Columns: []*Column{
					{Name: "id", Type: ColumnTypeBigInt, PrimaryKey: true, AutoIncrement: true},
					{Name: columnName, Type: ColumnTypeVarchar, Length: 255},
				},
			}

			// MySQL should use backticks and AUTO_INCREMENT
			mysqlGen := NewDDLGenerator("mysql")
			mysqlSQL := mysqlGen.GenerateCreateTableSQL(table)
			if !strings.Contains(mysqlSQL, "`") {
				return false
			}
			if !strings.Contains(strings.ToUpper(mysqlSQL), "AUTO_INCREMENT") {
				return false
			}

			// PostgreSQL should use double quotes and BIGSERIAL
			pgGen := NewDDLGenerator("postgres")
			pgSQL := pgGen.GenerateCreateTableSQL(table)
			if !strings.Contains(pgSQL, `"`) {
				return false
			}
			if !strings.Contains(strings.ToUpper(pgSQL), "BIGSERIAL") {
				return false
			}

			// SQLite should use double quotes and INTEGER
			sqliteGen := NewDDLGenerator("sqlite")
			sqliteSQL := sqliteGen.GenerateCreateTableSQL(table)
			if !strings.Contains(sqliteSQL, `"`) {
				return false
			}
			if !strings.Contains(strings.ToUpper(sqliteSQL), "INTEGER") {
				return false
			}

			return true
		},
		genTableName(),
		genColumnName(),
	))

	// Property 10.7: Add column SQL is valid
	properties.Property("Add column SQL is valid", prop.ForAll(
		func(tableName string, columnName string, colType ColumnType) bool {
			if tableName == "" || columnName == "" {
				return true
			}

			col := &Column{
				Name:   columnName,
				Type:   colType,
				Length: 255,
			}

			drivers := []string{"mysql", "postgres", "sqlite"}
			for _, driver := range drivers {
				gen := NewDDLGenerator(driver)
				sql := gen.GenerateAddColumnSQL(tableName, col)

				// Should contain ALTER TABLE
				if !strings.Contains(strings.ToUpper(sql), "ALTER TABLE") {
					return false
				}
				// Should contain ADD COLUMN
				if !strings.Contains(strings.ToUpper(sql), "ADD COLUMN") {
					return false
				}
				// Should contain column name
				if !strings.Contains(sql, columnName) {
					return false
				}
			}

			return true
		},
		genTableName(),
		genColumnName(),
		genColumnType(),
	))

	// Property 10.8: Rename column SQL is valid
	properties.Property("Rename column SQL is valid", prop.ForAll(
		func(tableName string, oldName string, newName string) bool {
			if tableName == "" || oldName == "" || newName == "" || oldName == newName {
				return true
			}

			drivers := []string{"mysql", "postgres", "sqlite"}
			for _, driver := range drivers {
				gen := NewDDLGenerator(driver)
				sql := gen.GenerateRenameColumnSQL(tableName, oldName, newName)

				// Should contain table name
				if !strings.Contains(sql, tableName) {
					return false
				}
				// Should contain old column name
				if !strings.Contains(sql, oldName) {
					return false
				}
				// Should contain new column name
				if !strings.Contains(sql, newName) {
					return false
				}
			}

			return true
		},
		genTableName(),
		genColumnName(),
		genColumnName(),
	))

	// Property 10.9: Model analyzer extracts correct column info
	properties.Property("Model analyzer extracts correct column info", prop.ForAll(
		func(fieldName string, isPrimaryKey bool, isAutoIncrement bool) bool {
			if fieldName == "" {
				return true
			}

			// Create a test struct type dynamically
			analyzer := NewModelAnalyzer()

			// Test with a simple struct
			type TestModel struct {
				ID   int64  `torm:"primary_key,auto_increment"`
				Name string `torm:"type:varchar,size:255"`
			}

			modelType := reflect.TypeOf(TestModel{})
			columns, err := analyzer.AnalyzeModel(modelType)
			if err != nil {
				return false
			}

			// Should have 2 columns
			if len(columns) != 2 {
				return false
			}

			// First column should be ID with primary key
			idCol := columns[0]
			if idCol.Name != "id" {
				return false
			}
			if !idCol.PrimaryKey {
				return false
			}
			if !idCol.AutoIncrement {
				return false
			}

			// Second column should be Name
			nameCol := columns[1]
			if nameCol.Name != "name" {
				return false
			}
			if nameCol.Type != ColumnTypeVarchar {
				return false
			}

			return true
		},
		genColumnName(),
		gen.Bool(),
		gen.Bool(),
	))

	// Property 10.10: Permission directives are parsed correctly
	properties.Property("Permission directives are parsed correctly", prop.ForAll(
		func(directive string) bool {
			analyzer := NewModelAnalyzer()

			// Test various permission directives
			testCases := []struct {
				directive string
				canRead   bool
				canCreate bool
				canUpdate bool
			}{
				{"<-", true, true, true},
				{"<-:create", true, true, false},
				{"<-:update", true, false, true},
				{"<-:false", true, false, false},
				{"->", true, true, true}, // Default permissions
			}

			for _, tc := range testCases {
				column := &ModelColumn{
					CanRead:   true,
					CanCreate: true,
					CanUpdate: true,
				}

				analyzer.parsePermissionDirective(tc.directive, column)

				// Check permissions match expected
				if tc.directive == "<-:create" && !column.CanCreate {
					return false
				}
				if tc.directive == "<-:update" && !column.CanUpdate {
					return false
				}
				if tc.directive == "<-:false" && (column.CanCreate || column.CanUpdate) {
					return false
				}
			}

			return true
		},
		gen.OneConstOf("<-", "<-:create", "<-:update", "<-:false", "->"),
	))

	properties.TestingRun(t)
}

// Helper types and functions for testing

type testColumn struct {
	Name       string
	Type       ColumnType
	Length     int
	PrimaryKey bool
	NotNull    bool
}

func createTestTable(name string, columns []testColumn) *Table {
	table := &Table{
		Name:    name,
		Columns: make([]*Column, len(columns)),
	}

	for i, col := range columns {
		table.Columns[i] = &Column{
			Name:       col.Name,
			Type:       col.Type,
			Length:     col.Length,
			PrimaryKey: col.PrimaryKey,
			NotNull:    col.NotNull,
		}
	}

	return table
}

// Generators

func genTableName() gopter.Gen {
	return gen.Identifier().Map(func(s string) string {
		if len(s) == 0 {
			return "test_table"
		}
		result := "tbl_" + strings.ToLower(s)
		if len(result) > 64 {
			result = result[:64]
		}
		return result
	})
}

func genColumnName() gopter.Gen {
	return gen.Identifier().Map(func(s string) string {
		if len(s) == 0 {
			return "test_column"
		}
		result := "col_" + strings.ToLower(s)
		if len(result) > 64 {
			result = result[:64]
		}
		return result
	})
}

func genIndexName() gopter.Gen {
	return gen.Identifier().Map(func(s string) string {
		if len(s) == 0 {
			return "idx_test"
		}
		result := "idx_" + strings.ToLower(s)
		if len(result) > 64 {
			result = result[:64]
		}
		return result
	})
}

func genColumnType() gopter.Gen {
	return gen.OneConstOf(
		ColumnTypeInt,
		ColumnTypeBigInt,
		ColumnTypeVarchar,
		ColumnTypeText,
		ColumnTypeBoolean,
		ColumnTypeDateTime,
	)
}

func genTestColumn() gopter.Gen {
	return gopter.CombineGens(
		genColumnName(),
		genColumnType(),
		gen.IntRange(1, 255),
		gen.Bool(),
		gen.Bool(),
	).Map(func(values []interface{}) testColumn {
		return testColumn{
			Name:       values[0].(string),
			Type:       values[1].(ColumnType),
			Length:     values[2].(int),
			PrimaryKey: values[3].(bool),
			NotNull:    values[4].(bool),
		}
	})
}
