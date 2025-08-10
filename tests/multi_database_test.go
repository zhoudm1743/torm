package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zhoudm1743/torm/pkg/db"
)

func setupMultiDatabase(t *testing.T) {
	// SQLite配置 - 使用内存数据库避免文件锁定问题
	sqliteConfig := &db.Config{
		Driver:          "sqlite",
		Database:        ":memory:",
		MaxOpenConns:    1, // 内存数据库使用单连接
		MaxIdleConns:    1,
		ConnMaxLifetime: time.Hour,
		LogQueries:      false,
	}

	// MySQL配置
	mysqlConfig := &db.Config{
		Driver:          "mysql",
		Host:            "127.0.0.1",
		Port:            3306,
		Database:        "test_multi",
		Username:        "root",
		Password:        "123456",
		Charset:         "utf8mb4",
		Timezone:        "UTC",
		MaxOpenConns:    20,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
		LogQueries:      false,
	}

	// PostgreSQL配置
	postgresConfig := &db.Config{
		Driver:          "postgres",
		Host:            "127.0.0.1",
		Port:            5432,
		Database:        "test_multi",
		Username:        "postgres",
		Password:        "123456",
		SSLMode:         "disable",
		MaxOpenConns:    15,
		MaxIdleConns:    8,
		ConnMaxLifetime: time.Hour,
		LogQueries:      false,
	}

	// 添加连接
	err := db.AddConnection("sqlite", sqliteConfig)
	require.NoError(t, err)

	err = db.AddConnection("mysql", mysqlConfig)
	if err != nil {
		t.Logf("MySQL连接失败，跳过MySQL测试: %v", err)
	}

	err = db.AddConnection("postgres", postgresConfig)
	if err != nil {
		t.Logf("PostgreSQL连接失败，跳过PostgreSQL测试: %v", err)
	}
}

func TestMultiDatabaseConnections(t *testing.T) {
	setupMultiDatabase(t)

	tests := []struct {
		name     string
		connName string
		skipMsg  string
	}{
		{"SQLite", "sqlite", ""},
		{"MySQL", "mysql", "MySQL连接不可用"},
		{"PostgreSQL", "postgres", "PostgreSQL连接不可用"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := db.DB(tt.connName)
			if err != nil {
				if tt.skipMsg != "" {
					t.Skip(tt.skipMsg)
				}
				require.NoError(t, err)
			}

			// 测试连接
			err = conn.Connect()
			if err != nil {
				if tt.skipMsg != "" {
					t.Skip(tt.skipMsg)
				}
				require.NoError(t, err)
			}

			// 测试Ping
			err = conn.Ping()
			if err != nil {
				if tt.skipMsg != "" {
					t.Skip(tt.skipMsg)
				}
				assert.NoError(t, err)
			}

			// 测试简单查询
			var testQuery string
			switch tt.connName {
			case "sqlite":
				testQuery = "SELECT 1"
			case "mysql":
				testQuery = "SELECT 1"
			case "postgres":
				testQuery = "SELECT 1"
			}

			rows, err := conn.Query(testQuery)
			if err != nil {
				if tt.skipMsg != "" {
					t.Skip(tt.skipMsg)
				}
				require.NoError(t, err)
			}
			defer rows.Close()

			assert.True(t, rows.Next())
			var result int
			err = rows.Scan(&result)
			assert.NoError(t, err)
			assert.Equal(t, 1, result)

			conn.Close()
		})
	}
}

func TestDatabaseCRUD(t *testing.T) {
	setupMultiDatabase(t)

	databases := []string{"sqlite"}

	for _, dbName := range databases {
		t.Run(dbName, func(t *testing.T) {
			conn, err := db.DB(dbName)
			if err != nil {
				t.Skipf("%s连接不可用", dbName)
			}

			err = conn.Connect()
			if err != nil {
				t.Skipf("%s连接失败", dbName)
			}
			defer conn.Close()

			// 创建测试表
			createTableSQL := `
			CREATE TABLE IF NOT EXISTS test_users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`

			_, err = conn.Exec(createTableSQL)
			require.NoError(t, err)

			// 插入测试数据
			insertSQL := "INSERT INTO test_users (name, email) VALUES (?, ?)"
			result, err := conn.Exec(insertSQL, "测试用户", "test@example.com")
			require.NoError(t, err)

			lastID, err := result.LastInsertId()
			require.NoError(t, err)
			assert.Greater(t, lastID, int64(0))

			// 查询数据
			selectSQL := "SELECT id, name, email FROM test_users WHERE id = ?"
			row := conn.QueryRow(selectSQL, lastID)

			var id int64
			var name, email string
			err = row.Scan(&id, &name, &email)
			require.NoError(t, err)

			assert.Equal(t, lastID, id)
			assert.Equal(t, "测试用户", name)
			assert.Equal(t, "test@example.com", email)

			// 更新数据
			updateSQL := "UPDATE test_users SET name = ? WHERE id = ?"
			_, err = conn.Exec(updateSQL, "更新用户", lastID)
			require.NoError(t, err)

			// 删除数据
			deleteSQL := "DELETE FROM test_users WHERE id = ?"
			_, err = conn.Exec(deleteSQL, lastID)
			require.NoError(t, err)

			// 清理表
			_, err = conn.Exec("DROP TABLE test_users")
			require.NoError(t, err)
		})
	}
}

func TestQueryBuilder(t *testing.T) {
	// 为此测试使用独立的连接
	testConfig := &db.Config{
		Driver:   "sqlite",
		Database: ":memory:",
	}

	connName := "query_test_" + t.Name()
	err := db.AddConnection(connName, testConfig)
	require.NoError(t, err)

	conn, err := db.DB(connName)
	require.NoError(t, err)

	err = conn.Connect()
	require.NoError(t, err)
	defer conn.Close()

	// 创建测试表
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS query_test (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		age INTEGER,
		status TEXT DEFAULT 'active'
	)`

	_, err = conn.Exec(createTableSQL)
	require.NoError(t, err)

	// 使用查询构建器，指定连接名
	query, err := db.Table("query_test", connName)
	require.NoError(t, err)

	// 插入测试数据
	testData := []map[string]interface{}{
		{"name": "用户1", "age": 25, "status": "active"},
		{"name": "用户2", "age": 30, "status": "inactive"},
		{"name": "用户3", "age": 22, "status": "active"},
	}

	for _, data := range testData {
		_, err = query.Insert(data)
		require.NoError(t, err)
	}

	// 查询所有数据
	results, err := query.Get()
	require.NoError(t, err)
	assert.Len(t, results, 3)

	// 条件查询
	activeQuery, err := db.Table("query_test", connName)
	require.NoError(t, err)

	activeResults, err := activeQuery.Where("status", "=", "active").Get()
	require.NoError(t, err)
	assert.Len(t, activeResults, 2)

	// 清理
	_, err = conn.Exec("DROP TABLE query_test")
	require.NoError(t, err)
}

func TestTransactions(t *testing.T) {
	// 为此测试使用时间戳的独立数据库文件
	dbName := fmt.Sprintf("test_tx_%d.db", time.Now().UnixNano())
	testConfig := &db.Config{
		Driver:   "sqlite",
		Database: dbName,
	}

	connName := fmt.Sprintf("tx_test_%d", time.Now().UnixNano())
	err := db.AddConnection(connName, testConfig)
	require.NoError(t, err)

	conn, err := db.DB(connName)
	require.NoError(t, err)

	err = conn.Connect()
	require.NoError(t, err)
	defer func() {
		conn.Close()
		// 清理测试文件
		_, _ = conn.Exec("DROP TABLE IF EXISTS tx_test")
	}()

	// 创建测试表
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS tx_test (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		amount INTEGER
	)`

	_, err = conn.Exec(createTableSQL)
	require.NoError(t, err)

	// 测试事务提交
	err = db.Transaction(func(tx db.TransactionInterface) error {
		_, err := tx.Exec("INSERT INTO tx_test (name, amount) VALUES (?, ?)", "账户A", 1000)
		if err != nil {
			return err
		}

		_, err = tx.Exec("INSERT INTO tx_test (name, amount) VALUES (?, ?)", "账户B", 500)
		return err
	}, connName)
	require.NoError(t, err)

	// 验证数据已提交
	rows, err := conn.Query("SELECT COUNT(*) FROM tx_test")
	require.NoError(t, err)
	defer rows.Close()

	assert.True(t, rows.Next())
	var count int
	err = rows.Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	// 测试事务回滚
	err = db.Transaction(func(tx db.TransactionInterface) error {
		_, err := tx.Exec("INSERT INTO tx_test (name, amount) VALUES (?, ?)", "账户C", 300)
		if err != nil {
			return err
		}
		// 故意返回错误来触发回滚
		return fmt.Errorf("故意回滚")
	}, connName)
	assert.Error(t, err) // 应该有错误

	// 验证数据已回滚，仍然是2条记录
	rows, err = conn.Query("SELECT COUNT(*) FROM tx_test")
	require.NoError(t, err)
	defer rows.Close()

	assert.True(t, rows.Next())
	err = rows.Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count) // 仍然是2条记录

	// 清理
	_, err = conn.Exec("DROP TABLE tx_test")
	require.NoError(t, err)
}

func TestConnectionPooling(t *testing.T) {
	// 简化连接池测试，使用时间戳的独立数据库文件
	dbName := fmt.Sprintf("test_pool_%d.db", time.Now().UnixNano())
	testConfig := &db.Config{
		Driver:       "sqlite",
		Database:     dbName,
		MaxOpenConns: 5,
		MaxIdleConns: 2,
	}

	connName := fmt.Sprintf("pool_test_%d", time.Now().UnixNano())
	err := db.AddConnection(connName, testConfig)
	require.NoError(t, err)

	conn, err := db.DB(connName)
	require.NoError(t, err)

	err = conn.Connect()
	require.NoError(t, err)
	defer func() {
		conn.Close()
		// 清理测试文件
		_, _ = conn.Exec("DROP TABLE IF EXISTS pool_test")
	}()

	// 创建测试表
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS pool_test (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		worker_id INTEGER,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	)`

	_, err = conn.Exec(createTableSQL)
	require.NoError(t, err)

	// 顺序测试多个操作
	for i := 0; i < 5; i++ {
		_, err := conn.Exec("INSERT INTO pool_test (worker_id) VALUES (?)", i)
		require.NoError(t, err)
	}

	// 验证总数
	rows, err := conn.Query("SELECT COUNT(*) FROM pool_test")
	require.NoError(t, err)
	defer rows.Close()

	assert.True(t, rows.Next())
	var totalCount int
	err = rows.Scan(&totalCount)
	require.NoError(t, err)
	assert.Equal(t, 5, totalCount)

	// 验证每个worker
	for i := 0; i < 5; i++ {
		rows, err := conn.Query("SELECT COUNT(*) FROM pool_test WHERE worker_id = ?", i)
		require.NoError(t, err)
		if rows != nil {
			assert.True(t, rows.Next())
			var count int
			err = rows.Scan(&count)
			require.NoError(t, err)
			assert.Equal(t, 1, count)
			rows.Close()
		}
	}
}

func TestDifferentDatabaseFeatures(t *testing.T) {
	setupMultiDatabase(t)

	t.Run("SQLite特性测试", func(t *testing.T) {
		conn, err := db.DB("sqlite")
		if err != nil {
			t.Skip("SQLite连接不可用")
		}

		err = conn.Connect()
		if err != nil {
			t.Skip("SQLite连接失败")
		}
		defer conn.Close()

		// 测试SQLite特有的PRAGMA命令
		rows, err := conn.Query("PRAGMA table_info(sqlite_master)")
		require.NoError(t, err)
		defer rows.Close()

		// 应该能获取到表信息
		columnCount := 0
		for rows.Next() {
			columnCount++
		}
		assert.Greater(t, columnCount, 0)
	})

	// 可以添加更多数据库特性测试
}

func TestErrorHandling(t *testing.T) {
	setupMultiDatabase(t)

	conn, err := db.DB("sqlite")
	if err != nil {
		t.Skip("SQLite连接不可用")
	}

	err = conn.Connect()
	if err != nil {
		t.Skip("SQLite连接失败")
	}
	defer conn.Close()

	// 测试无效SQL
	_, err = conn.Exec("INVALID SQL STATEMENT")
	assert.Error(t, err)

	// 测试查询不存在的表
	_, err = conn.Query("SELECT * FROM non_existent_table")
	assert.Error(t, err)

	// 测试事务中的错误
	tx, err := conn.Begin()
	require.NoError(t, err)

	_, err = tx.Exec("INVALID SQL IN TRANSACTION")
	assert.Error(t, err)

	// 即使有错误，也应该能够回滚
	err = tx.Rollback()
	assert.NoError(t, err)
}

// Benchmark tests
func BenchmarkMultiDatabaseInsert(b *testing.B) {
	setupMultiDatabase(nil)

	conn, err := db.DB("sqlite")
	if err != nil {
		b.Skip("SQLite连接不可用")
	}

	err = conn.Connect()
	if err != nil {
		b.Skip("SQLite连接失败")
	}
	defer conn.Close()

	// 创建测试表
	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS bench_test (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			value INTEGER
		)
	`)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn.Exec("INSERT INTO bench_test (name, value) VALUES (?, ?)", "测试", i)
	}

	// 清理
	conn.Exec("DROP TABLE bench_test")
}

func BenchmarkMultiDatabaseQuery(b *testing.B) {
	setupMultiDatabase(nil)

	conn, err := db.DB("sqlite")
	if err != nil {
		b.Skip("SQLite连接不可用")
	}

	err = conn.Connect()
	if err != nil {
		b.Skip("SQLite连接失败")
	}
	defer conn.Close()

	// 创建测试表并插入数据
	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS bench_query_test (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			value INTEGER
		)
	`)
	if err != nil {
		b.Fatal(err)
	}

	// 插入测试数据
	for i := 0; i < 1000; i++ {
		conn.Exec("INSERT INTO bench_query_test (name, value) VALUES (?, ?)", "测试", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows, err := conn.Query("SELECT * FROM bench_query_test LIMIT 10")
		if err == nil {
			rows.Close()
		}
	}

	// 清理
	conn.Exec("DROP TABLE bench_query_test")
}
