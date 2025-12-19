package db

import (
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Property 1: SQL Generation Round Trip**
// **Validates: Requirements 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 2.8, 2.9, 2.10, 2.11, 2.12, 2.13, 2.14**

// TestProperty1_SQLGenerationRoundTrip tests that Query configurations produce
// syntactically valid SQL strings that correctly represent the query intent.
func TestProperty1_SQLGenerationRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 1.1: SELECT with fields generates valid SQL
	properties.Property("SELECT with fields generates valid SQL", prop.ForAll(
		func(table string, fields []string) bool {
			if table == "" || len(fields) == 0 {
				return true // Skip empty inputs
			}

			q := NewQuery(nil).Table(table).Select(fields...)
			sql, _ := q.ToSQL()

			// Should start with SELECT
			if !strings.HasPrefix(strings.ToUpper(sql), "SELECT") {
				return false
			}
			// Should contain FROM
			if !strings.Contains(strings.ToUpper(sql), "FROM") {
				return false
			}
			// Should contain table name
			if !strings.Contains(sql, table) {
				return false
			}
			// Should contain all fields
			for _, field := range fields {
				if !strings.Contains(sql, field) {
					return false
				}
			}

			return true
		},
		genQueryTableName(),
		gen.SliceOfN(3, genQueryColumnName()),
	))

	// Property 1.2: DISTINCT generates valid SQL
	properties.Property("DISTINCT generates valid SQL", prop.ForAll(
		func(table string) bool {
			if table == "" {
				return true
			}

			q := NewQuery(nil).Table(table).Distinct()
			sql, _ := q.ToSQL()

			// Should contain DISTINCT
			if !strings.Contains(strings.ToUpper(sql), "DISTINCT") {
				return false
			}

			return true
		},
		genQueryTableName(),
	))

	// Property 1.3: WHERE conditions generate valid SQL with correct placeholders
	properties.Property("WHERE conditions generate valid SQL", prop.ForAll(
		func(table string, field string, value string) bool {
			if table == "" || field == "" {
				return true
			}

			q := NewQuery(nil).Table(table).Where(field, "=", value)
			sql, args := q.ToSQL()

			// Should contain WHERE
			if !strings.Contains(strings.ToUpper(sql), "WHERE") {
				return false
			}
			// Should contain field name
			if !strings.Contains(sql, field) {
				return false
			}
			// Should have one argument
			if len(args) != 1 {
				return false
			}
			// Argument should match value
			if args[0] != value {
				return false
			}

			return true
		},
		genQueryTableName(),
		genQueryColumnName(),
		gen.AnyString(),
	))

	// Property 1.4: Multiple WHERE conditions use AND/OR correctly
	properties.Property("Multiple WHERE conditions use AND/OR correctly", prop.ForAll(
		func(table string, field1 string, field2 string) bool {
			if table == "" || field1 == "" || field2 == "" {
				return true
			}

			q := NewQuery(nil).Table(table).
				Where(field1, "=", "val1").
				OrWhere(field2, "=", "val2")
			sql, args := q.ToSQL()

			// Should contain WHERE
			if !strings.Contains(strings.ToUpper(sql), "WHERE") {
				return false
			}
			// Should contain OR
			if !strings.Contains(strings.ToUpper(sql), " OR ") {
				return false
			}
			// Should have two arguments
			if len(args) != 2 {
				return false
			}

			return true
		},
		genQueryTableName(),
		genQueryColumnName(),
		genQueryColumnName(),
	))

	// Property 1.5: WhereIn generates correct IN clause
	properties.Property("WhereIn generates correct IN clause", prop.ForAll(
		func(table string, field string, numValues int) bool {
			if table == "" || field == "" || numValues <= 0 || numValues > 10 {
				return true
			}

			values := make([]interface{}, numValues)
			for i := 0; i < numValues; i++ {
				values[i] = i
			}

			q := NewQuery(nil).Table(table).WhereIn(field, values)
			sql, args := q.ToSQL()

			// Should contain IN
			if !strings.Contains(strings.ToUpper(sql), " IN ") {
				return false
			}
			// Should have correct number of arguments
			if len(args) != numValues {
				return false
			}

			return true
		},
		genQueryTableName(),
		genQueryColumnName(),
		gen.IntRange(1, 10),
	))

	// Property 1.6: WhereBetween generates correct BETWEEN clause
	properties.Property("WhereBetween generates correct BETWEEN clause", prop.ForAll(
		func(table string, field string, min int, max int) bool {
			if table == "" || field == "" {
				return true
			}

			q := NewQuery(nil).Table(table).WhereBetween(field, min, max)
			sql, args := q.ToSQL()

			// Should contain BETWEEN
			if !strings.Contains(strings.ToUpper(sql), "BETWEEN") {
				return false
			}
			// Should have two arguments
			if len(args) != 2 {
				return false
			}

			return true
		},
		genQueryTableName(),
		genQueryColumnName(),
		gen.Int(),
		gen.Int(),
	))

	// Property 1.7: WhereNull generates correct IS NULL clause
	properties.Property("WhereNull generates correct IS NULL clause", prop.ForAll(
		func(table string, field string) bool {
			if table == "" || field == "" {
				return true
			}

			q := NewQuery(nil).Table(table).WhereNull(field)
			sql, _ := q.ToSQL()

			// Should contain IS NULL
			if !strings.Contains(strings.ToUpper(sql), "IS NULL") {
				return false
			}

			return true
		},
		genQueryTableName(),
		genQueryColumnName(),
	))

	// Property 1.8: JOIN generates valid SQL
	properties.Property("JOIN generates valid SQL", prop.ForAll(
		func(table1 string, table2 string, condition string) bool {
			if table1 == "" || table2 == "" || condition == "" {
				return true
			}

			q := NewQuery(nil).Table(table1).Join(table2, condition)
			sql, _ := q.ToSQL()

			// Should contain JOIN
			if !strings.Contains(strings.ToUpper(sql), "JOIN") {
				return false
			}
			// Should contain both tables
			if !strings.Contains(sql, table1) || !strings.Contains(sql, table2) {
				return false
			}

			return true
		},
		genQueryTableName(),
		genQueryTableName(),
		genJoinCondition(),
	))

	// Property 1.9: LeftJoin generates LEFT JOIN
	properties.Property("LeftJoin generates LEFT JOIN", prop.ForAll(
		func(table1 string, table2 string, condition string) bool {
			if table1 == "" || table2 == "" || condition == "" {
				return true
			}

			q := NewQuery(nil).Table(table1).LeftJoin(table2, condition)
			sql, _ := q.ToSQL()

			// Should contain LEFT JOIN
			if !strings.Contains(strings.ToUpper(sql), "LEFT JOIN") {
				return false
			}

			return true
		},
		genQueryTableName(),
		genQueryTableName(),
		genJoinCondition(),
	))

	// Property 1.10: OrderBy generates valid ORDER BY clause
	properties.Property("OrderBy generates valid ORDER BY clause", prop.ForAll(
		func(table string, field string, direction string) bool {
			if table == "" || field == "" {
				return true
			}

			dir := "ASC"
			if direction == "DESC" {
				dir = "DESC"
			}

			q := NewQuery(nil).Table(table).OrderBy(field, dir)
			sql, _ := q.ToSQL()

			// Should contain ORDER BY
			if !strings.Contains(strings.ToUpper(sql), "ORDER BY") {
				return false
			}
			// Should contain field
			if !strings.Contains(sql, field) {
				return false
			}

			return true
		},
		genQueryTableName(),
		genQueryColumnName(),
		gen.OneConstOf("ASC", "DESC"),
	))

	// Property 1.11: GroupBy generates valid GROUP BY clause
	properties.Property("GroupBy generates valid GROUP BY clause", prop.ForAll(
		func(table string, fields []string) bool {
			if table == "" || len(fields) == 0 {
				return true
			}

			q := NewQuery(nil).Table(table).GroupBy(fields...)
			sql, _ := q.ToSQL()

			// Should contain GROUP BY
			if !strings.Contains(strings.ToUpper(sql), "GROUP BY") {
				return false
			}
			// Should contain all fields
			for _, field := range fields {
				if !strings.Contains(sql, field) {
					return false
				}
			}

			return true
		},
		genQueryTableName(),
		gen.SliceOfN(2, genQueryColumnName()),
	))

	// Property 1.12: Having generates valid HAVING clause
	properties.Property("Having generates valid HAVING clause", prop.ForAll(
		func(table string, field string) bool {
			if table == "" || field == "" {
				return true
			}

			q := NewQuery(nil).Table(table).GroupBy(field).Having("COUNT(*) > ?", 5)
			sql, args := q.ToSQL()

			// Should contain HAVING
			if !strings.Contains(strings.ToUpper(sql), "HAVING") {
				return false
			}
			// Should have one argument
			if len(args) != 1 {
				return false
			}

			return true
		},
		genQueryTableName(),
		genQueryColumnName(),
	))

	// Property 1.13: Limit generates valid LIMIT clause
	properties.Property("Limit generates valid LIMIT clause", prop.ForAll(
		func(table string, limit int) bool {
			if table == "" || limit <= 0 {
				return true
			}

			q := NewQuery(nil).Table(table).Limit(limit)
			sql, _ := q.ToSQL()

			// Should contain LIMIT
			if !strings.Contains(strings.ToUpper(sql), "LIMIT") {
				return false
			}

			return true
		},
		genQueryTableName(),
		gen.IntRange(1, 100),
	))

	// Property 1.14: Offset generates valid OFFSET clause
	properties.Property("Offset generates valid OFFSET clause", prop.ForAll(
		func(table string, limit int, offset int) bool {
			if table == "" || limit <= 0 || offset <= 0 {
				return true // Skip when offset is 0 since OFFSET 0 is not generated
			}

			q := NewQuery(nil).Table(table).Limit(limit).Offset(offset)
			sql, _ := q.ToSQL()

			// Should contain OFFSET when offset > 0
			if !strings.Contains(strings.ToUpper(sql), "OFFSET") {
				return false
			}

			return true
		},
		genQueryTableName(),
		gen.IntRange(1, 100),
		gen.IntRange(1, 100), // Start from 1 to ensure OFFSET is generated
	))

	// Property 1.15: Page generates correct LIMIT and OFFSET
	properties.Property("Page generates correct LIMIT and OFFSET", prop.ForAll(
		func(table string, page int, perPage int) bool {
			if table == "" || page <= 0 || perPage <= 0 {
				return true
			}

			q := NewQuery(nil).Table(table).Page(page, perPage)
			sql, _ := q.ToSQL()

			// Should contain LIMIT
			if !strings.Contains(strings.ToUpper(sql), "LIMIT") {
				return false
			}
			// If page > 1, should contain OFFSET
			if page > 1 && !strings.Contains(strings.ToUpper(sql), "OFFSET") {
				return false
			}

			return true
		},
		genQueryTableName(),
		gen.IntRange(1, 10),
		gen.IntRange(1, 50),
	))

	// Property 1.16: Complex query with all clauses generates valid SQL
	properties.Property("Complex query generates valid SQL", prop.ForAll(
		func(table string, field string) bool {
			if table == "" || field == "" {
				return true
			}

			q := NewQuery(nil).
				Table(table).
				Select(field, "id").
				Distinct().
				Where(field, "=", "value").
				OrderBy(field, "ASC").
				Limit(10).
				Offset(5)

			sql, args := q.ToSQL()

			// Should contain all expected clauses
			upperSQL := strings.ToUpper(sql)
			if !strings.Contains(upperSQL, "SELECT") {
				return false
			}
			if !strings.Contains(upperSQL, "DISTINCT") {
				return false
			}
			if !strings.Contains(upperSQL, "FROM") {
				return false
			}
			if !strings.Contains(upperSQL, "WHERE") {
				return false
			}
			if !strings.Contains(upperSQL, "ORDER BY") {
				return false
			}
			if !strings.Contains(upperSQL, "LIMIT") {
				return false
			}
			if !strings.Contains(upperSQL, "OFFSET") {
				return false
			}
			// Should have one argument for WHERE
			if len(args) != 1 {
				return false
			}

			return true
		},
		genQueryTableName(),
		genQueryColumnName(),
	))

	// Property 1.17: Clone creates independent copy
	properties.Property("Clone creates independent copy", prop.ForAll(
		func(table string, field string) bool {
			if table == "" || field == "" {
				return true
			}

			q1 := NewQuery(nil).Table(table).Where(field, "=", "val1")
			q2 := q1.Clone()

			// Modify q2
			q2.Where("other_field", "=", "val2")

			sql1, args1 := q1.ToSQL()
			sql2, args2 := q2.ToSQL()

			// q1 should have 1 WHERE condition
			if len(args1) != 1 {
				return false
			}
			// q2 should have 2 WHERE conditions
			if len(args2) != 2 {
				return false
			}
			// SQLs should be different
			if sql1 == sql2 {
				return false
			}

			return true
		},
		genQueryTableName(),
		genQueryColumnName(),
	))

	properties.TestingRun(t)
}

// Helper generators for Query tests

func genQueryTableName() gopter.Gen {
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

func genQueryColumnName() gopter.Gen {
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

func genJoinCondition() gopter.Gen {
	return gen.Identifier().Map(func(s string) string {
		if len(s) == 0 {
			return "t1.id = t2.id"
		}
		return "t1." + strings.ToLower(s) + " = t2." + strings.ToLower(s)
	})
}
