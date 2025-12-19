package db

import (
	"fmt"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Property 2: CRUD Operations Consistency**
// **Validates: Requirements 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.7, 3.8, 3.9, 3.10**

// setupTestTable creates a test table for CRUD operations
func setupTestTable(conn Connection, tableName string) error {
	driver := conn.GetDriver()
	var createSQL string

	switch driver {
	case "postgres":
		createSQL = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			value INTEGER DEFAULT 0,
			status VARCHAR(50) DEFAULT 'active'
		)`, tableName)
	case "mysql":
		createSQL = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			value INT DEFAULT 0,
			status VARCHAR(50) DEFAULT 'active'
		)`, tableName)
	default: // sqlite
		createSQL = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			value INTEGER DEFAULT 0,
			status TEXT DEFAULT 'active'
		)`, tableName)
	}

	_, err := conn.Exec(createSQL)
	return err
}

// cleanupTestTable removes all data from test table
func cleanupTestTable(conn Connection, tableName string) error {
	_, err := conn.Exec(fmt.Sprintf("DELETE FROM %s", tableName))
	return err
}

// dropTestTable drops the test table
func dropTestTable(conn Connection, tableName string) error {
	_, err := conn.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName))
	return err
}

// TestProperty2_CRUDOperationsConsistency tests CRUD operations consistency
// For any valid data map inserted into a table, the same data SHALL be retrievable
// via Get/First/Find, updatable via Update, and deletable via Delete.
func TestProperty2_CRUDOperationsConsistency(t *testing.T) {
	// Setup SQLite in-memory database for testing
	manager := NewDbManager()
	defer manager.CloseAll()

	config := &Config{
		Driver:   "sqlite",
		Database: ":memory:",
	}

	err := manager.SetConfig("crud_test", config)
	if err != nil {
		t.Fatalf("Failed to set config: %v", err)
	}

	conn, err := manager.Connection("crud_test")
	if err != nil {
		t.Fatalf("Failed to get connection: %v", err)
	}

	tableName := "crud_test_table"
	err = setupTestTable(conn, tableName)
	if err != nil {
		t.Fatalf("Failed to setup test table: %v", err)
	}
	defer dropTestTable(conn, tableName)

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 2.1: Insert then Get returns inserted data
	// Validates: Requirements 3.1, 3.5
	properties.Property("Insert then Get returns inserted data", prop.ForAll(
		func(name string, value int) bool {
			if name == "" {
				return true // Skip empty names
			}

			// Clean up before test
			cleanupTestTable(conn, tableName)

			// Insert data
			q := NewQuery(conn).Table(tableName)
			insertedID, err := q.Insert(map[string]interface{}{
				"name":  name,
				"value": value,
			})
			if err != nil {
				t.Logf("Insert failed: %v", err)
				return false
			}
			if insertedID <= 0 {
				t.Logf("Expected positive insert ID, got %d", insertedID)
				return false
			}

			// Get data
			q2 := NewQuery(conn).Table(tableName).Where("id", "=", insertedID)
			results, err := q2.Get()
			if err != nil {
				t.Logf("Get failed: %v", err)
				return false
			}

			if len(results) != 1 {
				t.Logf("Expected 1 result, got %d", len(results))
				return false
			}

			// Verify data matches
			result := results[0]
			if result["name"] != name {
				t.Logf("Name mismatch: expected %s, got %v", name, result["name"])
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
		gen.IntRange(-1000, 1000),
	))

	// Property 2.2: First returns single record
	// Validates: Requirements 3.2
	properties.Property("First returns single record", prop.ForAll(
		func(name string) bool {
			if name == "" {
				return true
			}

			cleanupTestTable(conn, tableName)

			// Insert multiple records
			q := NewQuery(conn).Table(tableName)
			q.Insert(map[string]interface{}{"name": name + "_1", "value": 1})
			q2 := NewQuery(conn).Table(tableName)
			q2.Insert(map[string]interface{}{"name": name + "_2", "value": 2})

			// First should return only one record
			q3 := NewQuery(conn).Table(tableName).OrderBy("id", "ASC")
			result, err := q3.First()
			if err != nil {
				t.Logf("First failed: %v", err)
				return false
			}

			if result == nil {
				t.Log("First returned nil")
				return false
			}

			// Should be the first inserted record
			if result["name"] != name+"_1" {
				t.Logf("Expected first record, got %v", result["name"])
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
	))

	// Property 2.3: Find by ID returns correct record
	// Validates: Requirements 3.3
	properties.Property("Find by ID returns correct record", prop.ForAll(
		func(name string, value int) bool {
			if name == "" {
				return true
			}

			cleanupTestTable(conn, tableName)

			// Insert data
			q := NewQuery(conn).Table(tableName)
			insertedID, err := q.Insert(map[string]interface{}{
				"name":  name,
				"value": value,
			})
			if err != nil {
				return false
			}

			// Find by ID
			q2 := NewQuery(conn).Table(tableName)
			result, err := q2.Find(insertedID)
			if err != nil {
				t.Logf("Find failed: %v", err)
				return false
			}

			if result["name"] != name {
				t.Logf("Name mismatch in Find: expected %s, got %v", name, result["name"])
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
		gen.IntRange(-1000, 1000),
	))

	// Property 2.4: Exists returns true for existing records
	// Validates: Requirements 3.9
	properties.Property("Exists returns true for existing records", prop.ForAll(
		func(name string) bool {
			if name == "" {
				return true
			}

			cleanupTestTable(conn, tableName)

			// Insert data
			q := NewQuery(conn).Table(tableName)
			_, err := q.Insert(map[string]interface{}{"name": name, "value": 1})
			if err != nil {
				return false
			}

			// Check exists
			q2 := NewQuery(conn).Table(tableName).Where("name", "=", name)
			exists, err := q2.Exists()
			if err != nil {
				t.Logf("Exists failed: %v", err)
				return false
			}

			if !exists {
				t.Log("Exists returned false for existing record")
				return false
			}

			// Check non-existing
			q3 := NewQuery(conn).Table(tableName).Where("name", "=", name+"_nonexistent")
			notExists, err := q3.Exists()
			if err != nil {
				return false
			}

			if notExists {
				t.Log("Exists returned true for non-existing record")
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
	))

	// Property 2.5: Update modifies existing records
	// Validates: Requirements 3.7
	properties.Property("Update modifies existing records", prop.ForAll(
		func(name string, oldValue int, newValue int) bool {
			if name == "" || oldValue == newValue {
				return true
			}

			cleanupTestTable(conn, tableName)

			// Insert data
			q := NewQuery(conn).Table(tableName)
			insertedID, err := q.Insert(map[string]interface{}{
				"name":  name,
				"value": oldValue,
			})
			if err != nil {
				return false
			}

			// Update data
			q2 := NewQuery(conn).Table(tableName).Where("id", "=", insertedID)
			affected, err := q2.Update(map[string]interface{}{"value": newValue})
			if err != nil {
				t.Logf("Update failed: %v", err)
				return false
			}

			if affected != 1 {
				t.Logf("Expected 1 affected row, got %d", affected)
				return false
			}

			// Verify update
			q3 := NewQuery(conn).Table(tableName)
			result, err := q3.Find(insertedID)
			if err != nil {
				return false
			}

			// SQLite returns int64 for INTEGER columns
			resultValue, ok := result["value"].(int64)
			if !ok {
				t.Logf("Value type mismatch: %T", result["value"])
				return false
			}

			if resultValue != int64(newValue) {
				t.Logf("Value not updated: expected %d, got %d", newValue, resultValue)
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
		gen.IntRange(-1000, 1000),
		gen.IntRange(-1000, 1000),
	))

	// Property 2.6: Delete removes records
	// Validates: Requirements 3.8
	properties.Property("Delete removes records", prop.ForAll(
		func(name string) bool {
			if name == "" {
				return true
			}

			cleanupTestTable(conn, tableName)

			// Insert data
			q := NewQuery(conn).Table(tableName)
			insertedID, err := q.Insert(map[string]interface{}{"name": name, "value": 1})
			if err != nil {
				return false
			}

			// Delete data
			q2 := NewQuery(conn).Table(tableName).Where("id", "=", insertedID)
			affected, err := q2.Delete()
			if err != nil {
				t.Logf("Delete failed: %v", err)
				return false
			}

			if affected != 1 {
				t.Logf("Expected 1 affected row, got %d", affected)
				return false
			}

			// Verify deletion
			q3 := NewQuery(conn).Table(tableName).Where("id", "=", insertedID)
			exists, err := q3.Exists()
			if err != nil {
				return false
			}

			if exists {
				t.Log("Record still exists after delete")
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
	))

	// Property 2.7: InsertBatch inserts multiple records
	// Validates: Requirements 3.6
	properties.Property("InsertBatch inserts multiple records", prop.ForAll(
		func(count int) bool {
			if count <= 0 || count > 10 {
				return true
			}

			cleanupTestTable(conn, tableName)

			// Prepare batch data
			data := make([]map[string]interface{}, count)
			for i := 0; i < count; i++ {
				data[i] = map[string]interface{}{
					"name":  fmt.Sprintf("batch_%d", i),
					"value": i,
				}
			}

			// Insert batch
			q := NewQuery(conn).Table(tableName)
			affected, err := q.InsertBatch(data)
			if err != nil {
				t.Logf("InsertBatch failed: %v", err)
				return false
			}

			if affected != int64(count) {
				t.Logf("Expected %d affected rows, got %d", count, affected)
				return false
			}

			// Verify count
			q2 := NewQuery(conn).Table(tableName)
			totalCount, err := q2.Count()
			if err != nil {
				return false
			}

			if totalCount != int64(count) {
				t.Logf("Expected %d records, got %d", count, totalCount)
				return false
			}

			return true
		},
		gen.IntRange(1, 10),
	))

	// Property 2.8: Increment increases field value
	// Validates: Requirements 3.10
	properties.Property("Increment increases field value", prop.ForAll(
		func(initialValue int, incrementAmount int) bool {
			if incrementAmount <= 0 || incrementAmount > 100 {
				return true
			}

			cleanupTestTable(conn, tableName)

			// Insert data
			q := NewQuery(conn).Table(tableName)
			insertedID, err := q.Insert(map[string]interface{}{
				"name":  "increment_test",
				"value": initialValue,
			})
			if err != nil {
				return false
			}

			// Increment
			q2 := NewQuery(conn).Table(tableName).Where("id", "=", insertedID)
			_, err = q2.Increment("value", incrementAmount)
			if err != nil {
				t.Logf("Increment failed: %v", err)
				return false
			}

			// Verify
			q3 := NewQuery(conn).Table(tableName)
			result, err := q3.Find(insertedID)
			if err != nil {
				return false
			}

			resultValue, ok := result["value"].(int64)
			if !ok {
				return false
			}

			expectedValue := int64(initialValue + incrementAmount)
			if resultValue != expectedValue {
				t.Logf("Increment failed: expected %d, got %d", expectedValue, resultValue)
				return false
			}

			return true
		},
		gen.IntRange(0, 1000),
		gen.IntRange(1, 100),
	))

	// Property 2.9: Decrement decreases field value
	// Validates: Requirements 3.10
	properties.Property("Decrement decreases field value", prop.ForAll(
		func(initialValue int, decrementAmount int) bool {
			if decrementAmount <= 0 || decrementAmount > 100 {
				return true
			}

			cleanupTestTable(conn, tableName)

			// Insert data
			q := NewQuery(conn).Table(tableName)
			insertedID, err := q.Insert(map[string]interface{}{
				"name":  "decrement_test",
				"value": initialValue,
			})
			if err != nil {
				return false
			}

			// Decrement
			q2 := NewQuery(conn).Table(tableName).Where("id", "=", insertedID)
			_, err = q2.Decrement("value", decrementAmount)
			if err != nil {
				t.Logf("Decrement failed: %v", err)
				return false
			}

			// Verify
			q3 := NewQuery(conn).Table(tableName)
			result, err := q3.Find(insertedID)
			if err != nil {
				return false
			}

			resultValue, ok := result["value"].(int64)
			if !ok {
				return false
			}

			expectedValue := int64(initialValue - decrementAmount)
			if resultValue != expectedValue {
				t.Logf("Decrement failed: expected %d, got %d", expectedValue, resultValue)
				return false
			}

			return true
		},
		gen.IntRange(100, 1000),
		gen.IntRange(1, 100),
	))

	properties.TestingRun(t)
}
