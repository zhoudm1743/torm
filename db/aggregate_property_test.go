package db

import (
	"fmt"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Property 9: Aggregate Correctness**
// **Validates: Requirements 9.1, 9.2, 9.3, 9.4, 9.5**

// TestProperty9_AggregateCorrectness tests that aggregate functions return
// values that match manual calculation over the same dataset.
func TestProperty9_AggregateCorrectness(t *testing.T) {
	// Setup SQLite in-memory database for testing
	manager := NewDbManager()
	defer manager.CloseAll()

	config := &Config{
		Driver:   "sqlite",
		Database: ":memory:",
	}

	err := manager.SetConfig("aggregate_test", config)
	if err != nil {
		t.Fatalf("Failed to set config: %v", err)
	}

	conn, err := manager.Connection("aggregate_test")
	if err != nil {
		t.Fatalf("Failed to get connection: %v", err)
	}

	tableName := "aggregate_test_table"
	err = setupAggregateTestTable(conn, tableName)
	if err != nil {
		t.Fatalf("Failed to setup test table: %v", err)
	}
	defer dropTestTable(conn, tableName)

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 9.1: Count returns correct number of records
	// Validates: Requirements 9.1
	properties.Property("Count returns correct number of records", prop.ForAll(
		func(count int) bool {
			if count <= 0 || count > 20 {
				return true
			}

			cleanupTestTable(conn, tableName)

			// Insert records
			for i := 0; i < count; i++ {
				q := NewQuery(conn).Table(tableName)
				_, err := q.Insert(map[string]interface{}{
					"name":  fmt.Sprintf("item_%d", i),
					"value": i * 10,
				})
				if err != nil {
					t.Logf("Insert failed: %v", err)
					return false
				}
			}

			// Count
			q := NewQuery(conn).Table(tableName)
			result, err := q.Count()
			if err != nil {
				t.Logf("Count failed: %v", err)
				return false
			}

			if result != int64(count) {
				t.Logf("Count mismatch: expected %d, got %d", count, result)
				return false
			}

			return true
		},
		gen.IntRange(1, 20),
	))

	// Property 9.2: Sum returns correct sum of values
	// Validates: Requirements 9.2
	properties.Property("Sum returns correct sum of values", prop.ForAll(
		func(values []int) bool {
			if len(values) == 0 || len(values) > 20 {
				return true
			}

			cleanupTestTable(conn, tableName)

			// Calculate expected sum
			expectedSum := 0
			for i, v := range values {
				expectedSum += v
				q := NewQuery(conn).Table(tableName)
				_, err := q.Insert(map[string]interface{}{
					"name":  fmt.Sprintf("item_%d", i),
					"value": v,
				})
				if err != nil {
					return false
				}
			}

			// Sum
			q := NewQuery(conn).Table(tableName)
			result, err := q.Sum("value")
			if err != nil {
				t.Logf("Sum failed: %v", err)
				return false
			}

			if int(result) != expectedSum {
				t.Logf("Sum mismatch: expected %d, got %f", expectedSum, result)
				return false
			}

			return true
		},
		gen.SliceOfN(10, gen.IntRange(-100, 100)),
	))

	// Property 9.3: Avg returns correct average of values
	// Validates: Requirements 9.3
	properties.Property("Avg returns correct average of values", prop.ForAll(
		func(values []int) bool {
			if len(values) == 0 || len(values) > 20 {
				return true
			}

			cleanupTestTable(conn, tableName)

			// Calculate expected average
			sum := 0
			for i, v := range values {
				sum += v
				q := NewQuery(conn).Table(tableName)
				_, err := q.Insert(map[string]interface{}{
					"name":  fmt.Sprintf("item_%d", i),
					"value": v,
				})
				if err != nil {
					return false
				}
			}
			expectedAvg := float64(sum) / float64(len(values))

			// Avg
			q := NewQuery(conn).Table(tableName)
			result, err := q.Avg("value")
			if err != nil {
				t.Logf("Avg failed: %v", err)
				return false
			}

			// Allow small floating point difference
			diff := result - expectedAvg
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.001 {
				t.Logf("Avg mismatch: expected %f, got %f", expectedAvg, result)
				return false
			}

			return true
		},
		gen.SliceOfN(10, gen.IntRange(-100, 100)),
	))

	// Property 9.4: Max returns correct maximum value
	// Validates: Requirements 9.4
	properties.Property("Max returns correct maximum value", prop.ForAll(
		func(values []int) bool {
			if len(values) == 0 || len(values) > 20 {
				return true
			}

			cleanupTestTable(conn, tableName)

			// Calculate expected max
			expectedMax := values[0]
			for i, v := range values {
				if v > expectedMax {
					expectedMax = v
				}
				q := NewQuery(conn).Table(tableName)
				_, err := q.Insert(map[string]interface{}{
					"name":  fmt.Sprintf("item_%d", i),
					"value": v,
				})
				if err != nil {
					return false
				}
			}

			// Max
			q := NewQuery(conn).Table(tableName)
			result, err := q.Max("value")
			if err != nil {
				t.Logf("Max failed: %v", err)
				return false
			}

			// Convert result to int64 for comparison
			var resultInt int64
			switch v := result.(type) {
			case int64:
				resultInt = v
			case float64:
				resultInt = int64(v)
			default:
				t.Logf("Unexpected Max result type: %T", result)
				return false
			}

			if resultInt != int64(expectedMax) {
				t.Logf("Max mismatch: expected %d, got %d", expectedMax, resultInt)
				return false
			}

			return true
		},
		gen.SliceOfN(10, gen.IntRange(-100, 100)),
	))

	// Property 9.5: Min returns correct minimum value
	// Validates: Requirements 9.5
	properties.Property("Min returns correct minimum value", prop.ForAll(
		func(values []int) bool {
			if len(values) == 0 || len(values) > 20 {
				return true
			}

			cleanupTestTable(conn, tableName)

			// Calculate expected min
			expectedMin := values[0]
			for i, v := range values {
				if v < expectedMin {
					expectedMin = v
				}
				q := NewQuery(conn).Table(tableName)
				_, err := q.Insert(map[string]interface{}{
					"name":  fmt.Sprintf("item_%d", i),
					"value": v,
				})
				if err != nil {
					return false
				}
			}

			// Min
			q := NewQuery(conn).Table(tableName)
			result, err := q.Min("value")
			if err != nil {
				t.Logf("Min failed: %v", err)
				return false
			}

			// Convert result to int64 for comparison
			var resultInt int64
			switch v := result.(type) {
			case int64:
				resultInt = v
			case float64:
				resultInt = int64(v)
			default:
				t.Logf("Unexpected Min result type: %T", result)
				return false
			}

			if resultInt != int64(expectedMin) {
				t.Logf("Min mismatch: expected %d, got %d", expectedMin, resultInt)
				return false
			}

			return true
		},
		gen.SliceOfN(10, gen.IntRange(-100, 100)),
	))

	// Property 9.6: Count with WHERE returns correct filtered count
	// Validates: Requirements 9.1, 3.4
	properties.Property("Count with WHERE returns correct filtered count", prop.ForAll(
		func(threshold int) bool {
			cleanupTestTable(conn, tableName)

			// Insert fixed set of values
			values := []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
			expectedCount := 0
			for i, v := range values {
				if v > threshold {
					expectedCount++
				}
				q := NewQuery(conn).Table(tableName)
				_, err := q.Insert(map[string]interface{}{
					"name":  fmt.Sprintf("item_%d", i),
					"value": v,
				})
				if err != nil {
					return false
				}
			}

			// Count with WHERE
			q := NewQuery(conn).Table(tableName).Where("value", ">", threshold)
			result, err := q.Count()
			if err != nil {
				t.Logf("Count with WHERE failed: %v", err)
				return false
			}

			if result != int64(expectedCount) {
				t.Logf("Filtered count mismatch: expected %d, got %d (threshold: %d)", expectedCount, result, threshold)
				return false
			}

			return true
		},
		gen.IntRange(0, 100),
	))

	// Property 9.7: Sum with WHERE returns correct filtered sum
	// Validates: Requirements 9.2
	properties.Property("Sum with WHERE returns correct filtered sum", prop.ForAll(
		func(threshold int) bool {
			cleanupTestTable(conn, tableName)

			// Insert fixed set of values
			values := []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
			expectedSum := 0
			for i, v := range values {
				if v > threshold {
					expectedSum += v
				}
				q := NewQuery(conn).Table(tableName)
				_, err := q.Insert(map[string]interface{}{
					"name":  fmt.Sprintf("item_%d", i),
					"value": v,
				})
				if err != nil {
					return false
				}
			}

			// Sum with WHERE
			q := NewQuery(conn).Table(tableName).Where("value", ">", threshold)
			result, err := q.Sum("value")
			if err != nil {
				t.Logf("Sum with WHERE failed: %v", err)
				return false
			}

			if int(result) != expectedSum {
				t.Logf("Filtered sum mismatch: expected %d, got %f (threshold: %d)", expectedSum, result, threshold)
				return false
			}

			return true
		},
		gen.IntRange(0, 100),
	))

	properties.TestingRun(t)
}

// setupAggregateTestTable creates a test table for aggregate operations
func setupAggregateTestTable(conn Connection, tableName string) error {
	driver := conn.GetDriver()
	var createSQL string

	switch driver {
	case "postgres":
		createSQL = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			value INTEGER DEFAULT 0
		)`, tableName)
	case "mysql":
		createSQL = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			value INT DEFAULT 0
		)`, tableName)
	default: // sqlite
		createSQL = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			value INTEGER DEFAULT 0
		)`, tableName)
	}

	_, err := conn.Exec(createSQL)
	return err
}
