package model

import (
	"fmt"
	"testing"

	"github.com/zhoudm1743/torm/db"
)

func TestDebugSoftDelete(t *testing.T) {
	// Setup - exactly like property test
	config := &db.Config{
		Driver:   "sqlite",
		Database: ":memory:",
	}
	err := db.AddConnection("default", config)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}
	defer db.CloseAllConnections()

	tableName := "soft_delete_test"

	// Create table - exactly like property test
	conn, err := db.DB("default")
	if err != nil {
		t.Fatalf("Failed to get connection: %v", err)
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
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Run multiple iterations like property test
	for i := 0; i < 5; i++ {
		t.Logf("=== Iteration %d ===", i)

		// Cleanup like property test
		_, err = conn.Exec(fmt.Sprintf("DELETE FROM %s", tableName))
		if err != nil {
			t.Fatalf("Cleanup failed: %v", err)
		}

		// Create model with soft delete - exactly like property test
		model := NewModel(ModelConfig{
			TableName:    tableName,
			PrimaryKey:   "id",
			Connection:   "default",
			SoftDeletes:  true,
			DeletedAtCol: "deleted_at",
		})

		name := fmt.Sprintf("test_%d", i)
		model.SetAttribute("name", name)
		model.SetAttribute("status", "active")

		t.Logf("Before Save - Model config: %+v", model.config)

		err = model.Save()
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}
		t.Logf("Inserted ID: %v", model.GetKey())

		// Soft delete
		err = model.Delete()
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
		t.Logf("Soft deleted")

		// Check model state after delete
		t.Logf("After Delete - Model config: %+v", model.config)
		t.Logf("After Delete - Model table name: %s", model.GetTableName())

		// Test WithTrashed - exactly like property test
		sdq := model.WithTrashed()
		t.Logf("SoftDeleteQuery scope: %d", sdq.scope)
		t.Logf("SoftDeleteQuery model pointer: %p", sdq.model)
		t.Logf("Original model pointer: %p", model)

		query, err := sdq.Query()
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		t.Logf("Query builder: %+v", query)

		results, err := query.Get()
		if err != nil {
			t.Fatalf("WithTrashed Get failed: %v", err)
		}
		t.Logf("WithTrashed Results count: %d", len(results))

		if len(results) != 1 {
			t.Errorf("Expected 1 result from WithTrashed query, got %d", len(results))
		}
	}
}
