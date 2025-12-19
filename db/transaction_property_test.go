package db

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"
	"github.com/zhoudm1743/torm/logger"
)

// **Property 3: Transaction Atomicity**
// **Validates: Requirements 4.1, 4.2, 4.3, 4.4, 4.5, 4.6**
//
// *For any* sequence of database operations executed within a Transaction callback,
// either all operations SHALL be committed (if callback returns nil) or all operations
// SHALL be rolled back (if callback returns error), with no partial state visible.

// Helper functions for transaction tests

func setupTxTestConnection(t *testing.T) Connection {
	config := &Config{
		Driver:   "sqlite",
		Database: ":memory:",
	}

	log := logger.NewSQLLogger(logger.INFO, true)
	conn, err := NewSQLiteConnection(config, log)
	if err != nil {
		t.Fatalf("Failed to create connection: %v", err)
	}

	err = conn.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	return conn
}

func createTxTestTable(t *testing.T, conn Connection, tableName string) {
	sql := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		value INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`, tableName)

	_, err := conn.Exec(sql)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}
}

func dropTxTestTable(conn Connection, tableName string) {
	sql := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
	conn.Exec(sql)
}

func cleanTxTable(conn Connection, tableName string) {
	sql := fmt.Sprintf("DELETE FROM %s", tableName)
	conn.Exec(sql)
}

// TestTransactionAtomicityProperty tests that transactions are atomic
func TestTransactionAtomicityProperty(t *testing.T) {
	// Setup SQLite connection for testing
	conn := setupTxTestConnection(t)
	defer conn.Close()

	// Create test table
	createTxTestTable(t, conn, "tx_test")
	defer dropTxTestTable(conn, "tx_test")

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 50

	properties := gopter.NewProperties(parameters)

	// Property: Committed transactions persist all data
	properties.Property("Committed transactions persist all data", prop.ForAll(
		func(values []int) bool {
			// Clean table before test
			cleanTxTable(conn, "tx_test")

			// Execute transaction that should commit
			err := conn.RunTransaction(func(tx TxInterface) error {
				for _, v := range values {
					_, err := tx.Exec("INSERT INTO tx_test (value) VALUES (?)", v)
					if err != nil {
						return err
					}
				}
				return nil // Return nil to commit
			})

			if err != nil {
				t.Logf("Transaction failed: %v", err)
				return false
			}

			// Verify all values were inserted
			query := NewQuery(conn)
			count, err := query.Table("tx_test").Count()
			if err != nil {
				t.Logf("Count failed: %v", err)
				return false
			}

			return count == int64(len(values))
		},
		gen.SliceOf(gen.IntRange(1, 1000)),
	))

	// Property: Rolled back transactions persist no data
	properties.Property("Rolled back transactions persist no data", prop.ForAll(
		func(values []int) bool {
			if len(values) == 0 {
				return true
			}

			// Clean table before test
			cleanTxTable(conn, "tx_test")

			// Get initial count
			query := NewQuery(conn)
			initialCount, _ := query.Table("tx_test").Count()

			// Execute transaction that should rollback
			testErr := errors.New("intentional rollback")
			err := conn.RunTransaction(func(tx TxInterface) error {
				for _, v := range values {
					_, err := tx.Exec("INSERT INTO tx_test (value) VALUES (?)", v)
					if err != nil {
						return err
					}
				}
				return testErr // Return error to rollback
			})

			// Should return the error
			if err != testErr {
				t.Logf("Expected testErr, got: %v", err)
				return false
			}

			// Verify no values were inserted (rollback worked)
			query2 := NewQuery(conn)
			finalCount, err := query2.Table("tx_test").Count()
			if err != nil {
				t.Logf("Count failed: %v", err)
				return false
			}

			return finalCount == initialCount
		},
		gen.SliceOf(gen.IntRange(1, 1000)),
	))

	properties.TestingRun(t)
}

// TestTransactionSavepointProperty tests nested transactions via savepoints
func TestTransactionSavepointProperty(t *testing.T) {
	// Setup SQLite connection for testing
	conn := setupTxTestConnection(t)
	defer conn.Close()

	// Create test table
	createTxTestTable(t, conn, "savepoint_test")
	defer dropTxTestTable(conn, "savepoint_test")

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 20

	properties := gopter.NewProperties(parameters)

	// Property: Savepoint rollback only affects operations after savepoint
	properties.Property("Savepoint rollback preserves operations before savepoint", prop.ForAll(
		func(beforeValues []int, afterValues []int) bool {
			// Clean table before test
			cleanTxTable(conn, "savepoint_test")

			tx, err := conn.Begin()
			if err != nil {
				t.Logf("Begin failed: %v", err)
				return false
			}

			// Insert values before savepoint
			for _, v := range beforeValues {
				_, err := tx.Exec("INSERT INTO savepoint_test (value) VALUES (?)", v)
				if err != nil {
					tx.Rollback()
					t.Logf("Insert before savepoint failed: %v", err)
					return false
				}
			}

			// Create savepoint
			err = tx.Savepoint("sp1")
			if err != nil {
				tx.Rollback()
				t.Logf("Savepoint failed: %v", err)
				return false
			}

			// Insert values after savepoint
			for _, v := range afterValues {
				_, err := tx.Exec("INSERT INTO savepoint_test (value) VALUES (?)", v)
				if err != nil {
					tx.Rollback()
					t.Logf("Insert after savepoint failed: %v", err)
					return false
				}
			}

			// Rollback to savepoint
			err = tx.RollbackTo("sp1")
			if err != nil {
				tx.Rollback()
				t.Logf("RollbackTo failed: %v", err)
				return false
			}

			// Commit transaction
			err = tx.Commit()
			if err != nil {
				t.Logf("Commit failed: %v", err)
				return false
			}

			// Verify only beforeValues were inserted
			query := NewQuery(conn)
			count, err := query.Table("savepoint_test").Count()
			if err != nil {
				t.Logf("Count failed: %v", err)
				return false
			}

			return count == int64(len(beforeValues))
		},
		gen.SliceOf(gen.IntRange(1, 100)),
		gen.SliceOf(gen.IntRange(101, 200)),
	))

	properties.TestingRun(t)
}

// TestTransactionQueryBuilderIntegration tests Query builder with transactions
func TestTransactionQueryBuilderIntegration(t *testing.T) {
	// Setup SQLite connection for testing
	conn := setupTxTestConnection(t)
	defer conn.Close()

	// Create test table
	createTxTestTable(t, conn, "query_tx_test")
	defer dropTxTestTable(conn, "query_tx_test")

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 30

	properties := gopter.NewProperties(parameters)

	// Property: Query builder operations in transaction are atomic
	properties.Property("Query builder operations in transaction are atomic", prop.ForAll(
		func(values []int) bool {
			// Clean table before test
			cleanTxTable(conn, "query_tx_test")

			tx, err := conn.Begin()
			if err != nil {
				t.Logf("Begin failed: %v", err)
				return false
			}

			// Use Query builder with transaction
			for _, v := range values {
				query := NewQuery(conn).Table("query_tx_test").InTransaction(tx)
				_, err := query.Insert(map[string]interface{}{"value": v})
				if err != nil {
					tx.Rollback()
					t.Logf("Insert failed: %v", err)
					return false
				}
			}

			// Commit
			err = tx.Commit()
			if err != nil {
				t.Logf("Commit failed: %v", err)
				return false
			}

			// Verify all values were inserted
			query := NewQuery(conn)
			count, err := query.Table("query_tx_test").Count()
			if err != nil {
				t.Logf("Count failed: %v", err)
				return false
			}

			return count == int64(len(values))
		},
		gen.SliceOf(gen.IntRange(1, 1000)),
	))

	// Property: Query builder rollback works correctly
	properties.Property("Query builder rollback works correctly", prop.ForAll(
		func(values []int) bool {
			if len(values) == 0 {
				return true
			}

			// Clean table before test
			cleanTxTable(conn, "query_tx_test")

			tx, err := conn.Begin()
			if err != nil {
				t.Logf("Begin failed: %v", err)
				return false
			}

			// Use Query builder with transaction
			for _, v := range values {
				query := NewQuery(conn).Table("query_tx_test").InTransaction(tx)
				_, err := query.Insert(map[string]interface{}{"value": v})
				if err != nil {
					tx.Rollback()
					t.Logf("Insert failed: %v", err)
					return false
				}
			}

			// Rollback instead of commit
			err = tx.Rollback()
			if err != nil {
				t.Logf("Rollback failed: %v", err)
				return false
			}

			// Verify no values were inserted
			query := NewQuery(conn)
			count, err := query.Table("query_tx_test").Count()
			if err != nil {
				t.Logf("Count failed: %v", err)
				return false
			}

			return count == 0
		},
		gen.SliceOf(gen.IntRange(1, 1000)),
	))

	properties.TestingRun(t)
}

// TestManualTransactionControl tests manual Begin/Commit/Rollback
func TestManualTransactionControl(t *testing.T) {
	// Setup SQLite connection for testing
	conn := setupTxTestConnection(t)
	defer conn.Close()

	// Create test table
	createTxTestTable(t, conn, "manual_tx_test")
	defer dropTxTestTable(conn, "manual_tx_test")

	t.Run("Manual commit persists data", func(t *testing.T) {
		cleanTxTable(conn, "manual_tx_test")

		tx, err := conn.Begin()
		assert.NoError(t, err)

		_, err = tx.Exec("INSERT INTO manual_tx_test (value) VALUES (?)", 100)
		assert.NoError(t, err)

		_, err = tx.Exec("INSERT INTO manual_tx_test (value) VALUES (?)", 200)
		assert.NoError(t, err)

		err = tx.Commit()
		assert.NoError(t, err)

		// Verify data persisted
		query := NewQuery(conn)
		count, err := query.Table("manual_tx_test").Count()
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("Manual rollback discards data", func(t *testing.T) {
		cleanTxTable(conn, "manual_tx_test")

		tx, err := conn.Begin()
		assert.NoError(t, err)

		_, err = tx.Exec("INSERT INTO manual_tx_test (value) VALUES (?)", 100)
		assert.NoError(t, err)

		_, err = tx.Exec("INSERT INTO manual_tx_test (value) VALUES (?)", 200)
		assert.NoError(t, err)

		err = tx.Rollback()
		assert.NoError(t, err)

		// Verify data was not persisted
		query := NewQuery(conn)
		count, err := query.Table("manual_tx_test").Count()
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("Savepoint and RollbackTo work correctly", func(t *testing.T) {
		cleanTxTable(conn, "manual_tx_test")

		tx, err := conn.Begin()
		assert.NoError(t, err)

		// Insert first value
		_, err = tx.Exec("INSERT INTO manual_tx_test (value) VALUES (?)", 100)
		assert.NoError(t, err)

		// Create savepoint
		err = tx.Savepoint("sp1")
		assert.NoError(t, err)

		// Insert second value
		_, err = tx.Exec("INSERT INTO manual_tx_test (value) VALUES (?)", 200)
		assert.NoError(t, err)

		// Rollback to savepoint
		err = tx.RollbackTo("sp1")
		assert.NoError(t, err)

		// Commit
		err = tx.Commit()
		assert.NoError(t, err)

		// Verify only first value persisted
		query := NewQuery(conn)
		count, err := query.Table("manual_tx_test").Count()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)

		// Verify the value is 100
		result, err := query.Table("manual_tx_test").First()
		assert.NoError(t, err)
		assert.Equal(t, int64(100), result["value"])
	})
}

// TestTransactionPanicRecovery tests that panics in transactions cause rollback
func TestTransactionPanicRecovery(t *testing.T) {
	// Setup SQLite connection for testing
	conn := setupTxTestConnection(t)
	defer conn.Close()

	// Create test table
	createTxTestTable(t, conn, "panic_test")
	defer dropTxTestTable(conn, "panic_test")

	cleanTxTable(conn, "panic_test")

	// Use a flag to track if panic was recovered
	panicRecovered := false

	func() {
		defer func() {
			if r := recover(); r != nil {
				panicRecovered = true
			}
		}()

		conn.RunTransaction(func(tx TxInterface) error {
			_, err := tx.Exec("INSERT INTO panic_test (value) VALUES (?)", 100)
			if err != nil {
				return err
			}
			panic("intentional panic")
		})
	}()

	assert.True(t, panicRecovered, "Expected panic to be propagated")

	// After panic recovery, verify no data was inserted
	// Need to wait a bit for SQLite to release locks
	time.Sleep(10 * time.Millisecond)

	query := NewQuery(conn)
	count, err := query.Table("panic_test").Count()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// TestTransactionIsolation tests that uncommitted changes are not visible after rollback
// Note: SQLite with single connection cannot test true isolation (concurrent reads during transaction)
// because SQLite uses connection-level locking. This test verifies rollback behavior instead.
func TestTransactionIsolation(t *testing.T) {
	// Setup SQLite connection for testing
	conn := setupTxTestConnection(t)
	defer conn.Close()

	// Create test table
	createTxTestTable(t, conn, "isolation_test")
	defer dropTxTestTable(conn, "isolation_test")

	t.Run("Rollback ensures no data persists", func(t *testing.T) {
		cleanTxTable(conn, "isolation_test")

		// Verify table is empty
		query := NewQuery(conn)
		initialCount, err := query.Table("isolation_test").Count()
		assert.NoError(t, err)
		assert.Equal(t, int64(0), initialCount)

		// Start transaction and insert data
		tx, err := conn.Begin()
		assert.NoError(t, err)

		_, err = tx.Exec("INSERT INTO isolation_test (value) VALUES (?)", 100)
		assert.NoError(t, err)

		_, err = tx.Exec("INSERT INTO isolation_test (value) VALUES (?)", 200)
		assert.NoError(t, err)

		// Rollback the transaction
		err = tx.Rollback()
		assert.NoError(t, err)

		// After rollback, count should still be 0
		query2 := NewQuery(conn)
		finalCount, err := query2.Table("isolation_test").Count()
		assert.NoError(t, err)
		assert.Equal(t, int64(0), finalCount)
	})

	t.Run("Commit persists data", func(t *testing.T) {
		cleanTxTable(conn, "isolation_test")

		// Start transaction and insert data
		tx, err := conn.Begin()
		assert.NoError(t, err)

		_, err = tx.Exec("INSERT INTO isolation_test (value) VALUES (?)", 100)
		assert.NoError(t, err)

		// Commit the transaction
		err = tx.Commit()
		assert.NoError(t, err)

		// After commit, count should be 1
		query := NewQuery(conn)
		count, err := query.Table("isolation_test").Count()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})
}

// Benchmark tests
func BenchmarkTransactionCommit(b *testing.B) {
	config := &Config{
		Driver:   "sqlite",
		Database: ":memory:",
	}

	conn, _ := NewSQLiteConnection(config, nil)
	conn.Connect()
	defer conn.Close()

	conn.Exec(`CREATE TABLE bench_tx (id INTEGER PRIMARY KEY, value INTEGER)`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn.RunTransaction(func(tx TxInterface) error {
			_, err := tx.Exec("INSERT INTO bench_tx (value) VALUES (?)", i)
			return err
		})
	}
}

func BenchmarkTransactionRollback(b *testing.B) {
	config := &Config{
		Driver:   "sqlite",
		Database: ":memory:",
	}

	conn, _ := NewSQLiteConnection(config, nil)
	conn.Connect()
	defer conn.Close()

	conn.Exec(`CREATE TABLE bench_tx_rb (id INTEGER PRIMARY KEY, value INTEGER)`)

	testErr := errors.New("rollback")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn.RunTransaction(func(tx TxInterface) error {
			tx.Exec("INSERT INTO bench_tx_rb (value) VALUES (?)", i)
			return testErr
		})
	}
}

func BenchmarkSavepoint(b *testing.B) {
	config := &Config{
		Driver:   "sqlite",
		Database: ":memory:",
	}

	conn, _ := NewSQLiteConnection(config, nil)
	conn.Connect()
	defer conn.Close()

	conn.Exec(`CREATE TABLE bench_sp (id INTEGER PRIMARY KEY, value INTEGER)`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx, _ := conn.Begin()
		tx.Exec("INSERT INTO bench_sp (value) VALUES (?)", i)
		tx.Savepoint("sp1")
		tx.Exec("INSERT INTO bench_sp (value) VALUES (?)", i+1)
		tx.RollbackTo("sp1")
		tx.Commit()
	}
}
