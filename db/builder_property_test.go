package db

import (
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Property 11: Cross-Database SQL Compatibility**
// **Validates: Requirements 14.1, 14.2, 14.3, 14.4**

// TestProperty11_CrossDatabaseSQLCompatibility tests that SQL builders generate
// syntactically valid SQL for their respective databases with correct placeholder styles.
func TestProperty11_CrossDatabaseSQLCompatibility(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 11.1: MySQL builder uses ? placeholders
	properties.Property("MySQL builder uses ? placeholders", prop.ForAll(
		func(table string, col1 string, col2 string) bool {
			if table == "" || col1 == "" || col2 == "" {
				return true // Skip empty inputs
			}
			builder := NewMySQLBuilderImpl()

			// Test INSERT
			data := map[string]interface{}{col1: "val1", col2: "val2"}
			insertSQL, _ := builder.BuildInsert(table, data)
			if insertSQL != "" {
				// MySQL should use ? placeholders
				if strings.Contains(insertSQL, "$") {
					return false
				}
				// Should contain ? placeholders
				if !strings.Contains(insertSQL, "?") {
					return false
				}
			}

			return true
		},
		genValidTableName(),
		genValidColumnName(),
		genValidColumnName(),
	))

	// Property 11.2: PostgreSQL builder uses $N placeholders
	properties.Property("PostgreSQL builder uses $N placeholders", prop.ForAll(
		func(table string, col1 string, col2 string) bool {
			if table == "" || col1 == "" || col2 == "" {
				return true // Skip empty inputs
			}
			builder := NewPostgreSQLBuilderImpl()

			// Test INSERT
			data := map[string]interface{}{col1: "val1", col2: "val2"}
			insertSQL, _ := builder.BuildInsert(table, data)
			if insertSQL != "" {
				// PostgreSQL should use $N placeholders
				if strings.Contains(insertSQL, "?") {
					return false
				}
				// Should contain $1, $2, etc.
				if !strings.Contains(insertSQL, "$1") {
					return false
				}
			}

			return true
		},
		genValidTableName(),
		genValidColumnName(),
		genValidColumnName(),
	))

	// Property 11.3: SQLite builder uses ? placeholders
	properties.Property("SQLite builder uses ? placeholders", prop.ForAll(
		func(table string, col1 string, col2 string) bool {
			if table == "" || col1 == "" || col2 == "" {
				return true // Skip empty inputs
			}
			builder := NewSQLiteBuilderImpl()

			// Test INSERT
			data := map[string]interface{}{col1: "val1", col2: "val2"}
			insertSQL, _ := builder.BuildInsert(table, data)
			if insertSQL != "" {
				// SQLite should use ? placeholders
				if strings.Contains(insertSQL, "$") {
					return false
				}
				// Should contain ? placeholders
				if !strings.Contains(insertSQL, "?") {
					return false
				}
			}

			return true
		},
		genValidTableName(),
		genValidColumnName(),
		genValidColumnName(),
	))

	// Property 11.4: All builders generate valid INSERT SQL structure
	properties.Property("All builders generate valid INSERT SQL structure", prop.ForAll(
		func(table string, col1 string, col2 string, val1 string, val2 string) bool {
			if table == "" || col1 == "" || col2 == "" {
				return true // Skip empty inputs
			}
			data := map[string]interface{}{col1: val1, col2: val2}

			// Test MySQL
			mysqlBuilder := NewMySQLBuilderImpl()
			mysqlSQL, mysqlArgs := mysqlBuilder.BuildInsert(table, data)
			if !validateInsertSQL(mysqlSQL, table, len(data), len(mysqlArgs)) {
				return false
			}

			// Test PostgreSQL
			pgsqlBuilder := NewPostgreSQLBuilderImpl()
			pgsqlSQL, pgsqlArgs := pgsqlBuilder.BuildInsert(table, data)
			if !validateInsertSQL(pgsqlSQL, table, len(data), len(pgsqlArgs)) {
				return false
			}

			// Test SQLite
			sqliteBuilder := NewSQLiteBuilderImpl()
			sqliteSQL, sqliteArgs := sqliteBuilder.BuildInsert(table, data)
			if !validateInsertSQL(sqliteSQL, table, len(data), len(sqliteArgs)) {
				return false
			}

			return true
		},
		genValidTableName(),
		genValidColumnName(),
		genValidColumnName(),
		gen.AnyString(),
		gen.AnyString(),
	))

	// Property 11.5: All builders generate valid UPDATE SQL structure
	properties.Property("All builders generate valid UPDATE SQL structure", prop.ForAll(
		func(table string, col1 string, val1 string) bool {
			if table == "" || col1 == "" {
				return true // Skip empty inputs
			}
			data := map[string]interface{}{col1: val1}
			conditions := []WhereClause{
				{Column: "id", Operator: "=", Value: 1, Logic: "AND"},
			}

			// Test MySQL
			mysqlBuilder := NewMySQLBuilderImpl()
			mysqlSQL, mysqlArgs := mysqlBuilder.BaseBuilder.BuildUpdateSQL(table, data, conditions)
			if !validateUpdateSQL(mysqlSQL, table, len(mysqlArgs)) {
				return false
			}

			// Test PostgreSQL
			pgsqlBuilder := NewPostgreSQLBuilderImpl()
			pgsqlSQL, pgsqlArgs := pgsqlBuilder.BaseBuilder.BuildUpdateSQL(table, data, conditions)
			if !validateUpdateSQL(pgsqlSQL, table, len(pgsqlArgs)) {
				return false
			}

			// Test SQLite
			sqliteBuilder := NewSQLiteBuilderImpl()
			sqliteSQL, sqliteArgs := sqliteBuilder.BaseBuilder.BuildUpdateSQL(table, data, conditions)
			if !validateUpdateSQL(sqliteSQL, table, len(sqliteArgs)) {
				return false
			}

			return true
		},
		genValidTableName(),
		genValidColumnName(),
		gen.AnyString(),
	))

	// Property 11.6: All builders generate valid DELETE SQL structure
	properties.Property("All builders generate valid DELETE SQL structure", prop.ForAll(
		func(table string) bool {
			if table == "" {
				return true // Skip empty inputs
			}
			conditions := []WhereClause{
				{Column: "id", Operator: "=", Value: 1, Logic: "AND"},
			}

			// Test MySQL
			mysqlBuilder := NewMySQLBuilderImpl()
			mysqlSQL, _ := mysqlBuilder.BaseBuilder.BuildDeleteSQL(table, conditions)
			if !validateDeleteSQL(mysqlSQL, table) {
				return false
			}

			// Test PostgreSQL
			pgsqlBuilder := NewPostgreSQLBuilderImpl()
			pgsqlSQL, _ := pgsqlBuilder.BaseBuilder.BuildDeleteSQL(table, conditions)
			if !validateDeleteSQL(pgsqlSQL, table) {
				return false
			}

			// Test SQLite
			sqliteBuilder := NewSQLiteBuilderImpl()
			sqliteSQL, _ := sqliteBuilder.BaseBuilder.BuildDeleteSQL(table, conditions)
			if !validateDeleteSQL(sqliteSQL, table) {
				return false
			}

			return true
		},
		genValidTableName(),
	))

	// Property 11.7: Placeholder count matches argument count
	properties.Property("Placeholder count matches argument count", prop.ForAll(
		func(numColumns int) bool {
			if numColumns <= 0 || numColumns > 10 {
				return true // Skip invalid cases
			}

			data := make(map[string]interface{})
			for i := 0; i < numColumns; i++ {
				data[genColumnName(i)] = "value"
			}

			// Test MySQL
			mysqlBuilder := NewMySQLBuilderImpl()
			mysqlSQL, mysqlArgs := mysqlBuilder.BuildInsert("test_table", data)
			mysqlPlaceholders := strings.Count(mysqlSQL, "?")
			if mysqlPlaceholders != len(mysqlArgs) {
				return false
			}

			// Test PostgreSQL
			pgsqlBuilder := NewPostgreSQLBuilderImpl()
			pgsqlSQL, pgsqlArgs := pgsqlBuilder.BuildInsert("test_table", data)
			pgsqlPlaceholders := countDollarPlaceholders(pgsqlSQL)
			if pgsqlPlaceholders != len(pgsqlArgs) {
				return false
			}

			// Test SQLite
			sqliteBuilder := NewSQLiteBuilderImpl()
			sqliteSQL, sqliteArgs := sqliteBuilder.BuildInsert("test_table", data)
			sqlitePlaceholders := strings.Count(sqliteSQL, "?")
			if sqlitePlaceholders != len(sqliteArgs) {
				return false
			}

			return true
		},
		gen.IntRange(1, 10),
	))

	// Property 11.8: Batch INSERT generates correct number of value groups
	properties.Property("Batch INSERT generates correct number of value groups", prop.ForAll(
		func(numRows int) bool {
			if numRows <= 0 || numRows > 10 {
				return true // Skip invalid cases
			}

			data := make([]map[string]interface{}, numRows)
			for i := 0; i < numRows; i++ {
				data[i] = map[string]interface{}{"col1": "val1", "col2": "val2"}
			}

			// Test MySQL
			mysqlBuilder := NewMySQLBuilderImpl()
			mysqlSQL, mysqlArgs := mysqlBuilder.BuildInsertBatch("test_table", data)
			// Each row should have 2 columns, so total args = numRows * 2
			if len(mysqlArgs) != numRows*2 {
				return false
			}
			// Should have numRows value groups (count opening parens in VALUES section)
			valuesIdx := strings.Index(strings.ToUpper(mysqlSQL), "VALUES")
			if valuesIdx == -1 {
				return false
			}
			valuesPart := mysqlSQL[valuesIdx:]
			valueGroups := strings.Count(valuesPart, "(")
			if valueGroups != numRows {
				return false
			}

			// Test PostgreSQL
			pgsqlBuilder := NewPostgreSQLBuilderImpl()
			_, pgsqlArgs := pgsqlBuilder.BuildInsertBatch("test_table", data)
			if len(pgsqlArgs) != numRows*2 {
				return false
			}

			// Test SQLite
			sqliteBuilder := NewSQLiteBuilderImpl()
			_, sqliteArgs := sqliteBuilder.BuildInsertBatch("test_table", data)
			if len(sqliteArgs) != numRows*2 {
				return false
			}

			return true
		},
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t)
}

// Helper generators

func genValidTableName() gopter.Gen {
	// Generate table names with a prefix to ensure they're always valid
	return gen.Identifier().Map(func(s string) string {
		if len(s) == 0 {
			return "test_table"
		}
		// Ensure it starts with a letter and is lowercase
		result := "t_" + strings.ToLower(s)
		if len(result) > 64 {
			result = result[:64]
		}
		return result
	})
}

func genValidColumnName() gopter.Gen {
	// Generate column names with a prefix to ensure they're always valid
	return gen.Identifier().Map(func(s string) string {
		if len(s) == 0 {
			return "test_column"
		}
		// Ensure it starts with a letter and is lowercase
		result := "c_" + strings.ToLower(s)
		if len(result) > 64 {
			result = result[:64]
		}
		return result
	})
}

func genColumnName(index int) string {
	return strings.ToLower(string(rune('a'+index))) + "_col"
}

// Helper validation functions

func validateInsertSQL(sql string, table string, numColumns int, numArgs int) bool {
	if sql == "" {
		return false
	}
	// Should start with INSERT INTO
	if !strings.HasPrefix(strings.ToUpper(sql), "INSERT INTO") {
		return false
	}
	// Should contain table name
	if !strings.Contains(sql, table) {
		return false
	}
	// Should contain VALUES
	if !strings.Contains(strings.ToUpper(sql), "VALUES") {
		return false
	}
	// Number of args should match number of columns
	if numArgs != numColumns {
		return false
	}
	return true
}

func validateUpdateSQL(sql string, table string, numArgs int) bool {
	if sql == "" {
		return false
	}
	// Should start with UPDATE
	if !strings.HasPrefix(strings.ToUpper(sql), "UPDATE") {
		return false
	}
	// Should contain table name
	if !strings.Contains(sql, table) {
		return false
	}
	// Should contain SET
	if !strings.Contains(strings.ToUpper(sql), "SET") {
		return false
	}
	// Should have at least one argument
	if numArgs < 1 {
		return false
	}
	return true
}

func validateDeleteSQL(sql string, table string) bool {
	if sql == "" {
		return false
	}
	// Should start with DELETE FROM
	if !strings.HasPrefix(strings.ToUpper(sql), "DELETE FROM") {
		return false
	}
	// Should contain table name
	if !strings.Contains(sql, table) {
		return false
	}
	return true
}

func countDollarPlaceholders(sql string) int {
	count := 0
	for i := 1; i <= 100; i++ {
		placeholder := "$" + string(rune('0'+i/10)) + string(rune('0'+i%10))
		if i < 10 {
			placeholder = "$" + string(rune('0'+i))
		}
		if strings.Contains(sql, placeholder) {
			count++
		}
	}
	return count
}
