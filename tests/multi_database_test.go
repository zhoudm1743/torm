package tests

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"torm/pkg/db"
)

func setupMultiDatabaseTest(t *testing.T) context.Context {
	ctx := context.Background()

	// 配置SQLite
	sqliteConfig := &db.Config{
		Driver:          "sqlite",
		Database:        "test_multi.db",
		MaxOpenConns:    50,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
		LogQueries:      false,
	}
	err := db.AddConnection("test_sqlite", sqliteConfig)
	require.NoError(t, err)

	// 配置MySQL（可能不可用）
	mysqlConfig := &db.Config{
		Driver:          "mysql",
		Host:            "127.0.0.1",
		Port:            3306,
		Database:        "orm",
		Username:        "root",
		Password:        "123456",
		Charset:         "utf8mb4",
		MaxOpenConns:    100,
		MaxIdleConns:    20,
		ConnMaxLifetime: time.Hour,
		LogQueries:      false,
	}
	db.AddConnection("test_mysql", mysqlConfig) // 可能失败，但不要求必须成功

	// 配置MongoDB（可能不可用）
	mongoConfig := &db.Config{
		Driver:          "mongodb",
		Host:            "127.0.0.1",
		Port:            27017,
		Database:        "orm_test",
		MaxOpenConns:    50,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
		LogQueries:      false,
	}
	db.AddConnection("test_mongodb", mongoConfig) // 可能失败，但不要求必须成功

	return ctx
}

func cleanupMultiDatabaseTest(t *testing.T, ctx context.Context) {
	// 清理SQLite
	if conn, err := db.DB("test_sqlite"); err == nil {
		conn.Exec(ctx, "DROP TABLE IF EXISTS multi_test_users")
	}

	// 清理MySQL
	if conn, err := db.DB("test_mysql"); err == nil {
		conn.Exec(ctx, "DROP TABLE IF EXISTS multi_test_users")
	}

	// 清理MongoDB
	if conn, err := db.DB("test_mongodb"); err == nil {
		mongoConn := db.GetMongoConnection(conn)
		if mongoConn != nil {
			mongoConn.GetCollection("multi_test_users").Drop(ctx)
		}
	}

	// 清理SQLite文件
	os.Remove("test_multi.db")
}

func TestMultiDatabaseConnections(t *testing.T) {
	ctx := setupMultiDatabaseTest(t)
	defer cleanupMultiDatabaseTest(t, ctx)

	databases := []string{"test_sqlite", "test_mysql", "test_mongodb"}
	connectedDatabases := make(map[string]bool)

	for _, dbName := range databases {
		t.Run(fmt.Sprintf("测试%s连接", dbName), func(t *testing.T) {
			conn, err := db.DB(dbName)
			if err != nil {
				t.Skipf("%s 连接获取失败: %v", dbName, err)
				return
			}

			err = conn.Ping(ctx)
			if err != nil {
				t.Skipf("%s 连接测试失败: %v", dbName, err)
				return
			}

			assert.NotEmpty(t, conn.GetDriver())
			connectedDatabases[dbName] = true
			t.Logf("✅ %s 连接成功 (驱动: %s)", dbName, conn.GetDriver())
		})
	}

	// 确保至少有一个数据库连接成功
	assert.Greater(t, len(connectedDatabases), 0, "至少应该有一个数据库连接成功")
}

func TestSQLiteOperations(t *testing.T) {
	ctx := setupMultiDatabaseTest(t)
	defer cleanupMultiDatabaseTest(t, ctx)

	conn, err := db.DB("test_sqlite")
	require.NoError(t, err)

	// 创建表
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS multi_test_users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		age INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`

	_, err = conn.Exec(ctx, createTableSQL)
	require.NoError(t, err)

	t.Run("SQLite CRUD操作", func(t *testing.T) {
		// 插入数据
		insertSQL := `INSERT INTO multi_test_users (name, email, age) VALUES (?, ?, ?)`
		result, err := conn.Exec(ctx, insertSQL, "SQLite测试用户", "sqlite_test@example.com", 25)
		assert.NoError(t, err)

		id, err := result.LastInsertId()
		assert.NoError(t, err)
		assert.Greater(t, id, int64(0))

		// 查询数据
		selectSQL := `SELECT id, name, email, age FROM multi_test_users WHERE email = ?`
		row := conn.QueryRow(ctx, selectSQL, "sqlite_test@example.com")

		var foundID int64
		var foundName, foundEmail string
		var foundAge int

		err = row.Scan(&foundID, &foundName, &foundEmail, &foundAge)
		assert.NoError(t, err)
		assert.Equal(t, "SQLite测试用户", foundName)
		assert.Equal(t, "sqlite_test@example.com", foundEmail)
		assert.Equal(t, 25, foundAge)

		// 更新数据
		updateSQL := `UPDATE multi_test_users SET age = ? WHERE id = ?`
		updateResult, err := conn.Exec(ctx, updateSQL, 26, foundID)
		assert.NoError(t, err)

		rowsAffected, err := updateResult.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), rowsAffected)

		// 删除数据
		deleteSQL := `DELETE FROM multi_test_users WHERE id = ?`
		deleteResult, err := conn.Exec(ctx, deleteSQL, foundID)
		assert.NoError(t, err)

		rowsAffected, err = deleteResult.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), rowsAffected)
	})
}

func TestMySQLOperations(t *testing.T) {
	ctx := setupMultiDatabaseTest(t)
	defer cleanupMultiDatabaseTest(t, ctx)

	conn, err := db.DB("test_mysql")
	if err != nil {
		t.Skipf("MySQL 连接失败: %v", err)
		return
	}

	err = conn.Ping(ctx)
	if err != nil {
		t.Skipf("MySQL 连接测试失败: %v", err)
		return
	}

	// 创建表
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS multi_test_users (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		age INT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`

	_, err = conn.Exec(ctx, createTableSQL)
	require.NoError(t, err)

	t.Run("MySQL CRUD操作", func(t *testing.T) {
		// 插入数据
		insertSQL := `INSERT INTO multi_test_users (name, email, age) VALUES (?, ?, ?)`
		result, err := conn.Exec(ctx, insertSQL, "MySQL测试用户", "mysql_test@example.com", 30)
		assert.NoError(t, err)

		id, err := result.LastInsertId()
		assert.NoError(t, err)
		assert.Greater(t, id, int64(0))

		// 查询数据
		selectSQL := `SELECT id, name, email, age FROM multi_test_users WHERE email = ?`
		row := conn.QueryRow(ctx, selectSQL, "mysql_test@example.com")

		var foundID int64
		var foundName, foundEmail string
		var foundAge int

		err = row.Scan(&foundID, &foundName, &foundEmail, &foundAge)
		assert.NoError(t, err)
		assert.Equal(t, "MySQL测试用户", foundName)
		assert.Equal(t, "mysql_test@example.com", foundEmail)
		assert.Equal(t, 30, foundAge)

		// 更新数据
		updateSQL := `UPDATE multi_test_users SET age = ? WHERE id = ?`
		updateResult, err := conn.Exec(ctx, updateSQL, 31, foundID)
		assert.NoError(t, err)

		rowsAffected, err := updateResult.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), rowsAffected)

		// 删除数据
		deleteSQL := `DELETE FROM multi_test_users WHERE id = ?`
		deleteResult, err := conn.Exec(ctx, deleteSQL, foundID)
		assert.NoError(t, err)

		rowsAffected, err = deleteResult.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), rowsAffected)
	})
}

func TestMongoDBOperations(t *testing.T) {
	ctx := setupMultiDatabaseTest(t)
	defer cleanupMultiDatabaseTest(t, ctx)

	conn, err := db.DB("test_mongodb")
	if err != nil {
		t.Skipf("MongoDB 连接失败: %v", err)
		return
	}

	err = conn.Ping(ctx)
	if err != nil {
		t.Skipf("MongoDB 连接测试失败: %v", err)
		return
	}

	mongoConn := db.GetMongoConnection(conn)
	require.NotNil(t, mongoConn)

	collection := mongoConn.GetCollection("multi_test_users")
	query := db.NewMongoQuery(collection, nil)

	t.Run("MongoDB CRUD操作", func(t *testing.T) {
		// 插入数据
		mongoUser := bson.M{
			"name":       "MongoDB测试用户",
			"email":      "mongo_test@example.com",
			"age":        28,
			"created_at": time.Now(),
		}

		result, err := query.InsertOne(ctx, mongoUser)
		assert.NoError(t, err)
		assert.NotNil(t, result.InsertedID)

		// 查询数据
		var foundUser bson.M
		err = query.Where("email", "mongo_test@example.com").FindOne(ctx).Decode(&foundUser)
		assert.NoError(t, err)
		assert.Equal(t, "MongoDB测试用户", foundUser["name"])
		assert.Equal(t, "mongo_test@example.com", foundUser["email"])
		assert.Equal(t, int32(28), foundUser["age"])

		// 更新数据
		update := bson.M{
			"$set": bson.M{
				"age":        int32(29),
				"updated_at": time.Now(),
			},
		}

		updateQuery := db.NewMongoQuery(collection, nil)
		updateResult, err := updateQuery.Where("email", "mongo_test@example.com").UpdateOne(ctx, update)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), updateResult.ModifiedCount)

		// 删除数据
		deleteQuery := db.NewMongoQuery(collection, nil)
		deleteResult, err := deleteQuery.Where("email", "mongo_test@example.com").DeleteOne(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), deleteResult.DeletedCount)
	})
}

func TestCrossDatabaseSync(t *testing.T) {
	ctx := setupMultiDatabaseTest(t)
	defer cleanupMultiDatabaseTest(t, ctx)

	// 检查哪些数据库可用
	availableDatabases := make(map[string]db.ConnectionInterface)

	for _, dbName := range []string{"test_sqlite", "test_mysql", "test_mongodb"} {
		if conn, err := db.DB(dbName); err == nil {
			if conn.Ping(ctx) == nil {
				availableDatabases[dbName] = conn
			}
		}
	}

	if len(availableDatabases) < 2 {
		t.Skip("需要至少2个数据库连接才能测试跨数据库同步")
		return
	}

	t.Logf("可用数据库: %v", getKeys(availableDatabases))

	// 准备测试数据
	testData := map[string]interface{}{
		"name":  "跨数据库同步用户",
		"email": "sync_test@example.com",
		"age":   25,
	}

	// 在第一个可用数据库中插入数据
	var sourceDB string
	for dbName := range availableDatabases {
		sourceDB = dbName
		break
	}

	err := insertTestData(ctx, availableDatabases[sourceDB], sourceDB, testData)
	require.NoError(t, err, "源数据库插入失败")

	// 同步到其他数据库
	for dbName, conn := range availableDatabases {
		if dbName != sourceDB {
			err := insertTestData(ctx, conn, dbName, testData)
			assert.NoError(t, err, fmt.Sprintf("同步到%s失败", dbName))
		}
	}

	// 验证所有数据库都有相同的数据
	for dbName, conn := range availableDatabases {
		found, err := verifyTestData(ctx, conn, dbName, testData)
		assert.NoError(t, err, fmt.Sprintf("验证%s数据失败", dbName))
		assert.True(t, found, fmt.Sprintf("%s中未找到同步的数据", dbName))
	}
}

func TestPerformanceComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	ctx := setupMultiDatabaseTest(t)
	defer cleanupMultiDatabaseTest(t, ctx)

	// 检查可用数据库
	availableDatabases := make(map[string]db.ConnectionInterface)

	for _, dbName := range []string{"test_sqlite", "test_mysql", "test_mongodb"} {
		if conn, err := db.DB(dbName); err == nil {
			if conn.Ping(ctx) == nil {
				availableDatabases[dbName] = conn
			}
		}
	}

	if len(availableDatabases) == 0 {
		t.Skip("没有可用的数据库连接")
		return
	}

	batchSize := 50 // 减小批量大小以加快测试
	results := make(map[string]time.Duration)

	for dbName, conn := range availableDatabases {
		t.Run(fmt.Sprintf("性能测试_%s", dbName), func(t *testing.T) {
			start := time.Now()
			err := performBatchInsert(ctx, conn, dbName, batchSize)
			duration := time.Since(start)

			assert.NoError(t, err, fmt.Sprintf("%s批量插入失败", dbName))
			results[dbName] = duration
			t.Logf("%s 批量插入%d条记录耗时: %v", dbName, batchSize, duration)
		})
	}

	// 输出性能排名
	if len(results) > 1 {
		t.Logf("性能对比结果 (批量插入%d条记录):", batchSize)
		for dbName, duration := range results {
			t.Logf("  %s: %v", dbName, duration)
		}
	}
}

// 辅助函数
func getKeys(m map[string]db.ConnectionInterface) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func insertTestData(ctx context.Context, conn db.ConnectionInterface, dbName string, data map[string]interface{}) error {
	switch conn.GetDriver() {
	case "sqlite", "mysql":
		// 创建表
		var createTableSQL string
		if conn.GetDriver() == "sqlite" {
			createTableSQL = `
			CREATE TABLE IF NOT EXISTS multi_test_users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				email TEXT UNIQUE NOT NULL,
				age INTEGER,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`
		} else {
			createTableSQL = `
			CREATE TABLE IF NOT EXISTS multi_test_users (
				id BIGINT AUTO_INCREMENT PRIMARY KEY,
				name VARCHAR(100) NOT NULL,
				email VARCHAR(100) UNIQUE NOT NULL,
				age INT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`
		}

		_, err := conn.Exec(ctx, createTableSQL)
		if err != nil {
			return err
		}

		// 插入数据
		insertSQL := `INSERT INTO multi_test_users (name, email, age) VALUES (?, ?, ?)`
		_, err = conn.Exec(ctx, insertSQL, data["name"], data["email"], data["age"])
		return err

	case "mongodb":
		mongoConn := db.GetMongoConnection(conn)
		if mongoConn == nil {
			return fmt.Errorf("无法获取MongoDB连接")
		}

		collection := mongoConn.GetCollection("multi_test_users")
		query := db.NewMongoQuery(collection, nil)

		mongoDoc := bson.M{
			"name":       data["name"],
			"email":      data["email"],
			"age":        data["age"],
			"created_at": time.Now(),
		}

		_, err := query.InsertOne(ctx, mongoDoc)
		return err

	default:
		return fmt.Errorf("不支持的数据库驱动: %s", conn.GetDriver())
	}
}

func verifyTestData(ctx context.Context, conn db.ConnectionInterface, dbName string, data map[string]interface{}) (bool, error) {
	switch conn.GetDriver() {
	case "sqlite", "mysql":
		selectSQL := `SELECT name, email, age FROM multi_test_users WHERE email = ?`
		row := conn.QueryRow(ctx, selectSQL, data["email"])

		var foundName, foundEmail string
		var foundAge int

		err := row.Scan(&foundName, &foundEmail, &foundAge)
		if err != nil {
			return false, err
		}

		return foundName == data["name"].(string) &&
			foundEmail == data["email"].(string) &&
			foundAge == data["age"].(int), nil

	case "mongodb":
		mongoConn := db.GetMongoConnection(conn)
		if mongoConn == nil {
			return false, fmt.Errorf("无法获取MongoDB连接")
		}

		collection := mongoConn.GetCollection("multi_test_users")
		query := db.NewMongoQuery(collection, nil)

		var foundUser bson.M
		err := query.Where("email", data["email"]).FindOne(ctx).Decode(&foundUser)
		if err != nil {
			return false, err
		}

		return foundUser["name"] == data["name"] &&
			foundUser["email"] == data["email"] &&
			foundUser["age"] == int32(data["age"].(int)), nil

	default:
		return false, fmt.Errorf("不支持的数据库驱动: %s", conn.GetDriver())
	}
}

func performBatchInsert(ctx context.Context, conn db.ConnectionInterface, dbName string, count int) error {
	switch conn.GetDriver() {
	case "sqlite", "mysql":
		// 创建表
		var createTableSQL string
		if conn.GetDriver() == "sqlite" {
			createTableSQL = `
			CREATE TABLE IF NOT EXISTS multi_test_users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				email TEXT UNIQUE NOT NULL,
				age INTEGER,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`
		} else {
			createTableSQL = `
			CREATE TABLE IF NOT EXISTS multi_test_users (
				id BIGINT AUTO_INCREMENT PRIMARY KEY,
				name VARCHAR(100) NOT NULL,
				email VARCHAR(100) UNIQUE NOT NULL,
				age INT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`
		}

		_, err := conn.Exec(ctx, createTableSQL)
		if err != nil {
			return err
		}

		// 批量插入
		insertSQL := `INSERT INTO multi_test_users (name, email, age) VALUES (?, ?, ?)`
		for i := 0; i < count; i++ {
			email := fmt.Sprintf("perf_%s_%d@test.com", dbName, i)
			_, err := conn.Exec(ctx, insertSQL, "性能测试用户", email, 25+i%20)
			if err != nil {
				return err
			}
		}
		return nil

	case "mongodb":
		mongoConn := db.GetMongoConnection(conn)
		if mongoConn == nil {
			return fmt.Errorf("无法获取MongoDB连接")
		}

		collection := mongoConn.GetCollection("multi_test_users")
		query := db.NewMongoQuery(collection, nil)

		for i := 0; i < count; i++ {
			mongoDoc := bson.M{
				"name":       "性能测试用户",
				"email":      fmt.Sprintf("perf_%s_%d@test.com", dbName, i),
				"age":        25 + i%20,
				"created_at": time.Now(),
			}

			_, err := query.InsertOne(ctx, mongoDoc)
			if err != nil {
				return err
			}
		}
		return nil

	default:
		return fmt.Errorf("不支持的数据库驱动: %s", conn.GetDriver())
	}
}
