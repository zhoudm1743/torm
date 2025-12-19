package model

import (
	"fmt"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/zhoudm1743/torm/db"
)

// **Property 4: Model Data Round Trip**
// **Validates: Requirements 5.1-5.6, 5.9-5.12**
//
// *For any* Model instance with valid field values, calling Save() followed by Find()
// SHALL return data equivalent to the original (accounting for accessor/mutator
// transformations and automatic timestamps).

// setupDefaultManagerForTest configures the default manager for testing
func setupDefaultManagerForTest(t *testing.T) func() {
	config := &db.Config{
		Driver:   "sqlite",
		Database: ":memory:",
	}

	// Use the default manager (singleton) that models use
	err := db.AddConnection("default", config)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	// Return cleanup function
	return func() {
		db.CloseAllConnections()
	}
}

// setupModelTestTable creates a test table for model operations
func setupModelTestTable(tableName string) error {
	conn, err := db.DB("default")
	if err != nil {
		return err
	}

	createSQL := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT,
		age INTEGER DEFAULT 0,
		balance REAL DEFAULT 0.0,
		is_active INTEGER DEFAULT 1,
		metadata TEXT,
		created_at DATETIME,
		updated_at DATETIME
	)`, tableName)

	_, err = conn.Exec(createSQL)
	return err
}

// cleanupModelTestTable removes all data from test table
func cleanupModelTestTable(tableName string) error {
	conn, err := db.DB("default")
	if err != nil {
		return err
	}
	_, err = conn.Exec(fmt.Sprintf("DELETE FROM %s", tableName))
	return err
}

// dropModelTestTable drops the test table
func dropModelTestTable(tableName string) error {
	conn, err := db.DB("default")
	if err != nil {
		return err
	}
	_, err = conn.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName))
	return err
}

// TestProperty4_ModelDataRoundTrip tests model data round trip consistency
func TestProperty4_ModelDataRoundTrip(t *testing.T) {
	// Setup default manager for testing
	cleanup := setupDefaultManagerForTest(t)
	defer cleanup()

	tableName := "model_round_trip_test"
	err := setupModelTestTable(tableName)
	if err != nil {
		t.Fatalf("Failed to setup test table: %v", err)
	}
	defer dropModelTestTable(tableName)

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 4.1: Save new model then Find returns equivalent data
	// Validates: Requirements 5.1, 5.2, 5.3, 5.4
	properties.Property("Save new model then Find returns equivalent data", prop.ForAll(
		func(name string, age int) bool {
			// Clean up before test
			cleanupModelTestTable(tableName)

			// Create new model
			model := NewModel(tableName)
			model.SetAttribute("name", name)
			model.SetAttribute("age", age)

			// Save model
			err := model.Save()
			if err != nil {
				t.Logf("Save failed: %v", err)
				return false
			}

			// Get the inserted ID
			insertedID := model.GetKey()
			if insertedID == nil {
				t.Log("Expected inserted ID to be set")
				return false
			}

			// Create new model and find by ID
			findModel := NewModel(tableName)
			err = findModel.Find(insertedID)
			if err != nil {
				t.Logf("Find failed: %v", err)
				return false
			}

			// Verify data matches
			if findModel.GetAttribute("name") != name {
				t.Logf("Name mismatch: expected %s, got %v", name, findModel.GetAttribute("name"))
				return false
			}

			// SQLite returns int64 for INTEGER columns
			foundAge, ok := findModel.GetAttribute("age").(int64)
			if !ok {
				t.Logf("Age type mismatch: %T", findModel.GetAttribute("age"))
				return false
			}
			if foundAge != int64(age) {
				t.Logf("Age mismatch: expected %d, got %d", age, foundAge)
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
		gen.IntRange(0, 150),
	))

	// Property 4.2: Update existing model then Find returns updated data
	// Validates: Requirements 5.5, 5.6
	properties.Property("Update existing model then Find returns updated data", prop.ForAll(
		func(name string, age int, newAge int) bool {
			cleanupModelTestTable(tableName)

			// Create and save initial model
			model := NewModel(tableName)
			model.SetAttribute("name", name)
			model.SetAttribute("age", age)

			err := model.Save()
			if err != nil {
				t.Logf("Initial save failed: %v", err)
				return false
			}

			insertedID := model.GetKey()

			// Update the model with different values
			newName := name + "_updated"
			model.SetAttribute("name", newName)
			model.SetAttribute("age", newAge)

			err = model.Save()
			if err != nil {
				t.Logf("Update save failed: %v", err)
				return false
			}

			// Find and verify updated data
			findModel := NewModel(tableName)
			err = findModel.Find(insertedID)
			if err != nil {
				t.Logf("Find after update failed: %v", err)
				return false
			}

			if findModel.GetAttribute("name") != newName {
				t.Logf("Updated name mismatch: expected %s, got %v", newName, findModel.GetAttribute("name"))
				return false
			}

			foundAge, ok := findModel.GetAttribute("age").(int64)
			if !ok {
				return false
			}
			if foundAge != int64(newAge) {
				t.Logf("Updated age mismatch: expected %d, got %d", newAge, foundAge)
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
		gen.IntRange(0, 150),
		gen.IntRange(0, 150),
	))

	// Property 4.3: Dirty tracking correctly identifies changed fields
	// Validates: Requirements 5.5, 5.6
	properties.Property("Dirty tracking correctly identifies changed fields", prop.ForAll(
		func(name string) bool {
			cleanupModelTestTable(tableName)

			// Create and save model
			model := NewModel(tableName)
			model.SetAttribute("name", name)
			model.SetAttribute("age", 25)

			err := model.Save()
			if err != nil {
				return false
			}

			// After save, model should not be dirty
			if model.IsDirty() {
				t.Log("Model should not be dirty after save")
				return false
			}

			// Modify attribute with a different value
			newName := name + "_modified"
			model.SetAttribute("name", newName)

			// Model should be dirty after modification
			if !model.IsDirty() {
				t.Log("Model should be dirty after modification")
				return false
			}

			// Check dirty fields
			dirty := model.GetDirty()
			if _, exists := dirty["name"]; !exists {
				t.Log("Expected 'name' to be in dirty fields")
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
	))

	// Property 4.4: Model state transitions correctly (IsNew, IsExists)
	// Validates: Requirements 5.4, 5.5
	properties.Property("Model state transitions correctly", prop.ForAll(
		func(name string) bool {
			if name == "" {
				return true
			}

			cleanupModelTestTable(tableName)

			// New model should be marked as new
			model := NewModel(tableName)
			if !model.IsNew() {
				t.Log("New model should be marked as new")
				return false
			}
			if model.IsExists() {
				t.Log("New model should not exist")
				return false
			}

			// After save, model should exist
			model.SetAttribute("name", name)
			err := model.Save()
			if err != nil {
				return false
			}

			if model.IsNew() {
				t.Log("Saved model should not be new")
				return false
			}
			if !model.IsExists() {
				t.Log("Saved model should exist")
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
	))

	properties.TestingRun(t)
}

// TestProperty4_TimestampManagement tests automatic timestamp management
// Validates: Requirements 5.9
func TestProperty4_TimestampManagement(t *testing.T) {
	cleanup := setupDefaultManagerForTest(t)
	defer cleanup()

	tableName := "timestamp_test"
	err := setupModelTestTable(tableName)
	if err != nil {
		t.Fatalf("Failed to setup test table: %v", err)
	}
	defer dropModelTestTable(tableName)

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: Timestamps are automatically set on create
	properties.Property("Timestamps are automatically set on create", prop.ForAll(
		func(name string) bool {
			if name == "" {
				return true
			}

			cleanupModelTestTable(tableName)

			beforeSave := time.Now().Add(-time.Second)

			model := NewModel(tableName)
			model.SetAttribute("name", name)

			err := model.Save()
			if err != nil {
				t.Logf("Save failed: %v", err)
				return false
			}

			afterSave := time.Now().Add(time.Second)

			// Find the saved model
			findModel := NewModel(tableName)
			err = findModel.Find(model.GetKey())
			if err != nil {
				t.Logf("Find failed: %v", err)
				return false
			}

			// Check created_at was set
			createdAt := findModel.GetAttribute("created_at")
			if createdAt == nil {
				t.Log("created_at should be set")
				return false
			}

			// Parse and verify timestamp is within expected range
			var createdTime time.Time
			switch v := createdAt.(type) {
			case time.Time:
				createdTime = v
			case string:
				createdTime, _ = time.Parse("2006-01-02 15:04:05", v)
			}

			if !createdTime.IsZero() {
				if createdTime.Before(beforeSave) || createdTime.After(afterSave) {
					t.Logf("created_at %v not in expected range [%v, %v]", createdTime, beforeSave, afterSave)
					return false
				}
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
	))

	properties.TestingRun(t)
}

// TestProperty4_JSONFieldRoundTrip tests JSON field serialization/deserialization
// Validates: Requirements 5.12
func TestProperty4_JSONFieldRoundTrip(t *testing.T) {
	cleanup := setupDefaultManagerForTest(t)
	defer cleanup()

	tableName := "json_field_test"
	err := setupModelTestTable(tableName)
	if err != nil {
		t.Fatalf("Failed to setup test table: %v", err)
	}
	defer dropModelTestTable(tableName)

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: JSON serialization round trip preserves data
	properties.Property("JSON serialization round trip preserves data", prop.ForAll(
		func(name string, idx int) bool {
			cleanupModelTestTable(tableName)

			// Generate key and value based on index to avoid SuchThat filtering
			key := fmt.Sprintf("key_%d", idx)
			value := fmt.Sprintf("value_%d", idx)

			// Create model with JSON metadata
			model := NewModel(tableName)
			model.SetAttribute("name", name)

			// Create JSON data
			jsonData := JSONMap{key: value}
			jsonBytes, err := jsonData.Value()
			if err != nil {
				t.Logf("JSON marshal failed: %v", err)
				return false
			}

			model.SetAttribute("metadata", string(jsonBytes.([]byte)))

			err = model.Save()
			if err != nil {
				t.Logf("Save failed: %v", err)
				return false
			}

			// Find and verify JSON data
			findModel := NewModel(tableName)
			err = findModel.Find(model.GetKey())
			if err != nil {
				t.Logf("Find failed: %v", err)
				return false
			}

			// Parse JSON from found model
			metadataStr, ok := findModel.GetAttribute("metadata").(string)
			if !ok {
				t.Logf("metadata type mismatch: %T", findModel.GetAttribute("metadata"))
				return false
			}

			var foundJSON JSONMap
			err = foundJSON.Scan(metadataStr)
			if err != nil {
				t.Logf("JSON unmarshal failed: %v", err)
				return false
			}

			// Verify data matches
			if foundJSON[key] != value {
				t.Logf("JSON value mismatch: expected %s, got %v", value, foundJSON[key])
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
		gen.IntRange(1, 1000),
	))

	properties.TestingRun(t)
}

// TestProperty4_ModelToJSONRoundTrip tests model JSON serialization
// Validates: Requirements 5.10, 5.11, 5.12
func TestProperty4_ModelToJSONRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: ToJSON then FromJSON preserves model attributes
	properties.Property("ToJSON then FromJSON preserves model attributes", prop.ForAll(
		func(name string, email string, age int) bool {
			if name == "" {
				return true
			}

			// Create model with attributes
			model := NewModel("test_table")
			model.SetAttribute("name", name)
			model.SetAttribute("email", email)
			model.SetAttribute("age", age)

			// Convert to JSON
			jsonStr, err := model.ToJSON()
			if err != nil {
				t.Logf("ToJSON failed: %v", err)
				return false
			}

			// Create new model and parse from JSON
			newModel := NewModel("test_table")
			err = newModel.FromJSON(jsonStr)
			if err != nil {
				t.Logf("FromJSON failed: %v", err)
				return false
			}

			// Verify attributes match
			if newModel.GetAttribute("name") != name {
				t.Logf("Name mismatch after JSON round trip")
				return false
			}

			if newModel.GetAttribute("email") != email {
				t.Logf("Email mismatch after JSON round trip")
				return false
			}

			// JSON unmarshals numbers as float64
			foundAge, ok := newModel.GetAttribute("age").(float64)
			if !ok {
				t.Logf("Age type mismatch: %T", newModel.GetAttribute("age"))
				return false
			}
			if int(foundAge) != age {
				t.Logf("Age mismatch after JSON round trip: expected %d, got %v", age, foundAge)
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) < 100 }),
		gen.IntRange(0, 1000),
	))

	properties.TestingRun(t)
}

// **Property 5: Soft Delete Exclusion**
// **Validates: Requirements 5.7, 5.8**
//
// *For any* Model with soft delete enabled, after calling Delete(), the record
// SHALL still exist in the database but SHALL NOT be returned by normal queries,
// and SHALL be returned when explicitly including soft-deleted records.

// setupSoftDeleteTestTable creates a test table with soft delete support
func setupSoftDeleteTestTable(tableName string) error {
	conn, err := db.DB("default")
	if err != nil {
		return err
	}

	createSQL := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		status TEXT DEFAULT 'active',
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`, tableName)

	_, err = conn.Exec(createSQL)
	return err
}

// cleanupSoftDeleteTestTable removes all data from test table
func cleanupSoftDeleteTestTable(tableName string) error {
	conn, err := db.DB("default")
	if err != nil {
		return err
	}
	_, err = conn.Exec(fmt.Sprintf("DELETE FROM %s", tableName))
	return err
}

// dropSoftDeleteTestTable drops the test table
func dropSoftDeleteTestTable(tableName string) error {
	conn, err := db.DB("default")
	if err != nil {
		return err
	}
	_, err = conn.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName))
	return err
}

// TestProperty5_SoftDeleteExclusion tests soft delete behavior
func TestProperty5_SoftDeleteExclusion(t *testing.T) {
	// Setup default manager for testing
	cleanup := setupDefaultManagerForTest(t)
	defer cleanup()

	tableName := "soft_delete_test"
	err := setupSoftDeleteTestTable(tableName)
	if err != nil {
		t.Fatalf("Failed to setup test table: %v", err)
	}
	defer dropSoftDeleteTestTable(tableName)

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 5.1: Soft deleted records still exist in database
	// Validates: Requirements 5.7
	properties.Property("Soft deleted records still exist in database", prop.ForAll(
		func(name string) bool {
			cleanupSoftDeleteTestTable(tableName)

			// Create model with soft delete enabled
			model := NewModel(ModelConfig{
				TableName:    tableName,
				PrimaryKey:   "id",
				Connection:   "default",
				SoftDeletes:  true,
				DeletedAtCol: "deleted_at",
			})
			model.SetAttribute("name", name)
			model.SetAttribute("status", "active")

			err := model.Save()
			if err != nil {
				t.Logf("Save failed: %v", err)
				return false
			}

			insertedID := model.GetKey()

			// Soft delete the record
			err = model.Delete()
			if err != nil {
				t.Logf("Delete failed: %v", err)
				return false
			}

			// Verify record still exists in database using raw query
			conn, err := db.DB("default")
			if err != nil {
				t.Logf("Failed to get connection: %v", err)
				return false
			}

			var count int64
			row := conn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE id = ?", tableName), insertedID)
			err = row.Scan(&count)
			if err != nil {
				t.Logf("Count query failed: %v", err)
				return false
			}

			if count != 1 {
				t.Logf("Expected record to still exist in database, got count: %d", count)
				return false
			}

			// Verify deleted_at is set
			var deletedAt interface{}
			row = conn.QueryRow(fmt.Sprintf("SELECT deleted_at FROM %s WHERE id = ?", tableName), insertedID)
			err = row.Scan(&deletedAt)
			if err != nil {
				t.Logf("deleted_at query failed: %v", err)
				return false
			}

			if deletedAt == nil {
				t.Log("Expected deleted_at to be set")
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
	))

	// Property 5.2: Normal queries exclude soft deleted records
	// Validates: Requirements 5.7
	properties.Property("Normal queries exclude soft deleted records", prop.ForAll(
		func(name string) bool {
			cleanupSoftDeleteTestTable(tableName)

			// Create model with soft delete enabled
			model := NewModel(ModelConfig{
				TableName:    tableName,
				PrimaryKey:   "id",
				Connection:   "default",
				SoftDeletes:  true,
				DeletedAtCol: "deleted_at",
			})
			model.SetAttribute("name", name)
			model.SetAttribute("status", "active")

			err := model.Save()
			if err != nil {
				t.Logf("Save failed: %v", err)
				return false
			}

			// Soft delete the record
			err = model.Delete()
			if err != nil {
				t.Logf("Delete failed: %v", err)
				return false
			}

			// Query using normal model query (should exclude soft deleted)
			query, err := model.Query()
			if err != nil {
				t.Logf("Query failed: %v", err)
				return false
			}

			results, err := query.Get()
			if err != nil {
				t.Logf("Get failed: %v", err)
				return false
			}

			// Should return 0 results (soft deleted record excluded)
			if len(results) != 0 {
				t.Logf("Expected 0 results from normal query, got %d", len(results))
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
	))

	// Property 5.3: WithTrashed includes soft deleted records
	// Validates: Requirements 5.8
	properties.Property("WithTrashed includes soft deleted records", prop.ForAll(
		func(name string) bool {
			cleanupSoftDeleteTestTable(tableName)

			// Create model with soft delete enabled
			model := NewModel(ModelConfig{
				TableName:    tableName,
				PrimaryKey:   "id",
				Connection:   "default",
				SoftDeletes:  true,
				DeletedAtCol: "deleted_at",
			})
			model.SetAttribute("name", name)
			model.SetAttribute("status", "active")

			err := model.Save()
			if err != nil {
				t.Logf("Save failed: %v", err)
				return false
			}

			// Soft delete the record
			err = model.Delete()
			if err != nil {
				t.Logf("Delete failed: %v", err)
				return false
			}

			// Query using WithTrashed (should include soft deleted)
			results, err := model.WithTrashed().Get()
			if err != nil {
				t.Logf("WithTrashed Get failed: %v", err)
				return false
			}

			// Should return 1 result (soft deleted record included)
			if len(results) != 1 {
				t.Logf("Expected 1 result from WithTrashed query, got %d", len(results))
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
	))

	// Property 5.4: OnlyTrashed returns only soft deleted records
	// Validates: Requirements 5.8
	properties.Property("OnlyTrashed returns only soft deleted records", prop.ForAll(
		func(name1 string, idx int) bool {
			cleanupSoftDeleteTestTable(tableName)

			// Generate unique name2 based on index to ensure it's different from name1
			name2 := fmt.Sprintf("active_%d", idx)

			// Create first model (will be soft deleted)
			model1 := NewModel(ModelConfig{
				TableName:    tableName,
				PrimaryKey:   "id",
				Connection:   "default",
				SoftDeletes:  true,
				DeletedAtCol: "deleted_at",
			})
			model1.SetAttribute("name", name1)
			model1.SetAttribute("status", "deleted")

			err := model1.Save()
			if err != nil {
				t.Logf("Save model1 failed: %v", err)
				return false
			}

			deletedID := model1.GetKey()

			// Soft delete the first record
			err = model1.Delete()
			if err != nil {
				t.Logf("Delete model1 failed: %v", err)
				return false
			}

			// Create second model (will remain active)
			model2 := NewModel(ModelConfig{
				TableName:    tableName,
				PrimaryKey:   "id",
				Connection:   "default",
				SoftDeletes:  true,
				DeletedAtCol: "deleted_at",
			})
			model2.SetAttribute("name", name2)
			model2.SetAttribute("status", "active")

			err = model2.Save()
			if err != nil {
				t.Logf("Save model2 failed: %v", err)
				return false
			}

			// Query using OnlyTrashed (should return only soft deleted)
			results, err := model1.OnlyTrashed().Get()
			if err != nil {
				t.Logf("OnlyTrashed Get failed: %v", err)
				return false
			}

			// Should return 1 result (only the soft deleted record)
			if len(results) != 1 {
				t.Logf("Expected 1 result from OnlyTrashed query, got %d", len(results))
				return false
			}

			// Verify it's the deleted record by checking the ID
			resultID := results[0]["id"]
			if resultID != deletedID {
				t.Logf("Expected deleted record ID %v, got %v", deletedID, resultID)
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
		gen.IntRange(1, 10000),
	))

	// Property 5.5: ForceDelete permanently removes records
	// Validates: Requirements 5.8
	properties.Property("ForceDelete permanently removes records", prop.ForAll(
		func(name string) bool {
			cleanupSoftDeleteTestTable(tableName)

			// Create model with soft delete enabled
			model := NewModel(ModelConfig{
				TableName:    tableName,
				PrimaryKey:   "id",
				Connection:   "default",
				SoftDeletes:  true,
				DeletedAtCol: "deleted_at",
			})
			model.SetAttribute("name", name)
			model.SetAttribute("status", "active")

			err := model.Save()
			if err != nil {
				t.Logf("Save failed: %v", err)
				return false
			}

			insertedID := model.GetKey()

			// Force delete the record
			err = model.ForceDelete()
			if err != nil {
				t.Logf("ForceDelete failed: %v", err)
				return false
			}

			// Verify record no longer exists in database
			conn, err := db.DB("default")
			if err != nil {
				t.Logf("Failed to get connection: %v", err)
				return false
			}

			var count int64
			row := conn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE id = ?", tableName), insertedID)
			err = row.Scan(&count)
			if err != nil {
				t.Logf("Count query failed: %v", err)
				return false
			}

			if count != 0 {
				t.Logf("Expected record to be permanently deleted, got count: %d", count)
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
	))

	// Property 5.6: Restore brings back soft deleted records
	// Validates: Requirements 5.8
	properties.Property("Restore brings back soft deleted records", prop.ForAll(
		func(name string) bool {
			cleanupSoftDeleteTestTable(tableName)

			// Create model with soft delete enabled
			model := NewModel(ModelConfig{
				TableName:    tableName,
				PrimaryKey:   "id",
				Connection:   "default",
				SoftDeletes:  true,
				DeletedAtCol: "deleted_at",
			})
			model.SetAttribute("name", name)
			model.SetAttribute("status", "active")

			err := model.Save()
			if err != nil {
				t.Logf("Save failed: %v", err)
				return false
			}

			// Soft delete the record
			err = model.Delete()
			if err != nil {
				t.Logf("Delete failed: %v", err)
				return false
			}

			// Verify record is soft deleted (not in normal query)
			query, err := model.Query()
			if err != nil {
				t.Logf("Query failed: %v", err)
				return false
			}

			results, err := query.Get()
			if err != nil {
				t.Logf("Get failed: %v", err)
				return false
			}

			if len(results) != 0 {
				t.Logf("Expected 0 results after soft delete, got %d", len(results))
				return false
			}

			// Restore the record
			err = model.Restore()
			if err != nil {
				t.Logf("Restore failed: %v", err)
				return false
			}

			// Verify record is back in normal query
			query, err = model.Query()
			if err != nil {
				t.Logf("Query after restore failed: %v", err)
				return false
			}

			results, err = query.Get()
			if err != nil {
				t.Logf("Get after restore failed: %v", err)
				return false
			}

			if len(results) != 1 {
				t.Logf("Expected 1 result after restore, got %d", len(results))
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
	))

	properties.TestingRun(t)
}
