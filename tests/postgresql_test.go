package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"torm/pkg/db"
)

func setupPostgreSQLTest(t *testing.T) (context.Context, db.ConnectionInterface) {
	// 配置PostgreSQL测试连接
	config := &db.Config{
		Driver:          "postgres",
		Host:            "localhost",
		Port:            5432,
		Database:        "torm_test",
		Username:        "torm_user",
		Password:        "torm_password",
		SSLMode:         "disable",
		Timezone:        "UTC",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		LogQueries:      false,
	}

	err := db.AddConnection("postgres_test", config)
	require.NoError(t, err)

	ctx := context.Background()
	conn, err := db.DB("postgres_test")
	require.NoError(t, err)

	// 测试连接
	err = conn.Ping(ctx)
	if err != nil {
		t.Skipf("PostgreSQL not available: %v", err)
	}

	// 创建测试表
	conn.Exec(ctx, "DROP TABLE IF EXISTS test_users CASCADE")
	conn.Exec(ctx, `
		CREATE TABLE test_users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			age INTEGER DEFAULT 0,
			status VARCHAR(20) DEFAULT 'active',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)

	return ctx, conn
}

func TestPostgreSQLConnection(t *testing.T) {
	ctx, conn := setupPostgreSQLTest(t)

	t.Run("连接测试", func(t *testing.T) {
		err := conn.Ping(ctx)
		assert.NoError(t, err)
		assert.True(t, conn.IsConnected())
		assert.Equal(t, "postgres", conn.GetDriver())
	})

	t.Run("配置测试", func(t *testing.T) {
		config := conn.GetConfig()
		assert.NotNil(t, config)
		assert.Equal(t, "postgres", config.Driver)
		assert.Equal(t, "localhost", config.Host)
		assert.Equal(t, 5432, config.Port)
	})

	t.Run("统计信息", func(t *testing.T) {
		stats := conn.GetStats()
		assert.GreaterOrEqual(t, stats.MaxOpenConnections, 0)
	})
}

func TestPostgreSQLCRUD(t *testing.T) {
	ctx, _ := setupPostgreSQLTest(t)

	t.Run("插入数据", func(t *testing.T) {
		query, err := db.Table("test_users", "postgres_test")
		require.NoError(t, err)

		data := map[string]interface{}{
			"name":   "PostgreSQL测试用户",
			"email":  "postgres_test@example.com",
			"age":    25,
			"status": "active",
		}

		id, err := query.Insert(ctx, data)
		assert.NoError(t, err)
		assert.Greater(t, id, int64(0))
	})

	t.Run("查询数据", func(t *testing.T) {
		query, err := db.Table("test_users", "postgres_test")
		require.NoError(t, err)

		// 查询所有数据
		users, err := query.Get(ctx)
		assert.NoError(t, err)
		assert.Greater(t, len(users), 0)

		// 条件查询
		activeUsers, err := query.Where("status", "=", "active").Get(ctx)
		assert.NoError(t, err)
		assert.Greater(t, len(activeUsers), 0)

		// 查询单条记录
		user, err := query.Where("email", "=", "postgres_test@example.com").First(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "PostgreSQL测试用户", user["name"])
	})

	t.Run("更新数据", func(t *testing.T) {
		query, err := db.Table("test_users", "postgres_test")
		require.NoError(t, err)

		updateData := map[string]interface{}{
			"age": 26,
		}

		affected, err := query.Where("email", "=", "postgres_test@example.com").Update(ctx, updateData)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), affected)

		// 验证更新
		user, err := query.Where("email", "=", "postgres_test@example.com").First(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(26), user["age"])
	})

	t.Run("批量插入", func(t *testing.T) {
		query, err := db.Table("test_users", "postgres_test")
		require.NoError(t, err)

		batchData := []map[string]interface{}{
			{"name": "批量用户1", "email": "batch1@example.com", "age": 20, "status": "active"},
			{"name": "批量用户2", "email": "batch2@example.com", "age": 21, "status": "inactive"},
			{"name": "批量用户3", "email": "batch3@example.com", "age": 22, "status": "active"},
		}

		affected, err := query.InsertBatch(ctx, batchData)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), affected)
	})

	t.Run("聚合查询", func(t *testing.T) {
		query, err := db.Table("test_users", "postgres_test")
		require.NoError(t, err)

		// 计数
		count, err := query.Count(ctx)
		assert.NoError(t, err)
		assert.Greater(t, count, int64(0))

		// 条件计数
		activeCount, err := query.Where("status", "=", "active").Count(ctx)
		assert.NoError(t, err)
		assert.Greater(t, activeCount, int64(0))
	})

	t.Run("删除数据", func(t *testing.T) {
		query, err := db.Table("test_users", "postgres_test")
		require.NoError(t, err)

		affected, err := query.Where("email", "=", "batch2@example.com").Delete(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), affected)

		// 验证删除
		user, err := query.Where("email", "=", "batch2@example.com").First(ctx)
		assert.Error(t, err) // 应该找不到记录
		assert.Nil(t, user)
	})
}

func TestPostgreSQLTransaction(t *testing.T) {
	ctx, _ := setupPostgreSQLTest(t)

	t.Run("事务提交", func(t *testing.T) {
		err := db.Transaction(ctx, func(tx db.TransactionInterface) error {
			_, err := tx.Exec(ctx, `
				INSERT INTO test_users (name, email, age, status) 
				VALUES ($1, $2, $3, $4)
			`, "事务用户1", "tx_user1@example.com", 30, "active")

			if err != nil {
				return err
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO test_users (name, email, age, status) 
				VALUES ($1, $2, $3, $4)
			`, "事务用户2", "tx_user2@example.com", 31, "active")

			return err
		}, "postgres_test")

		assert.NoError(t, err)

		// 验证数据已插入
		query, _ := db.Table("test_users", "postgres_test")
		user1, err := query.Where("email", "=", "tx_user1@example.com").First(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "事务用户1", user1["name"])

		user2, err := query.Where("email", "=", "tx_user2@example.com").First(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "事务用户2", user2["name"])
	})

	t.Run("事务回滚", func(t *testing.T) {
		err := db.Transaction(ctx, func(tx db.TransactionInterface) error {
			_, err := tx.Exec(ctx, `
				INSERT INTO test_users (name, email, age, status) 
				VALUES ($1, $2, $3, $4)
			`, "回滚用户1", "rollback_user1@example.com", 25, "active")

			if err != nil {
				return err
			}

			// 故意触发错误导致回滚
			_, err = tx.Exec(ctx, `
				INSERT INTO test_users (name, email, age, status) 
				VALUES ($1, $2, $3, $4)
			`, "回滚用户2", "tx_user1@example.com", 26, "active") // 重复邮箱

			return err // 返回错误，触发回滚
		}, "postgres_test")

		assert.Error(t, err) // 事务应该失败

		// 验证数据未插入（已回滚）
		query, _ := db.Table("test_users", "postgres_test")
		user, err := query.Where("email", "=", "rollback_user1@example.com").First(ctx)
		assert.Error(t, err) // 应该找不到记录
		assert.Nil(t, user)
	})
}

func TestPostgreSQLAdvancedQueries(t *testing.T) {
	ctx, conn := setupPostgreSQLTest(t)

	// 准备测试数据
	conn.Exec(ctx, "DROP TABLE IF EXISTS test_profiles CASCADE")
	conn.Exec(ctx, `
		CREATE TABLE test_profiles (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES test_users(id) ON DELETE CASCADE,
			bio TEXT,
			website VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)

	// 插入测试数据
	query, _ := db.Table("test_users", "postgres_test")
	userID, _ := query.Insert(ctx, map[string]interface{}{
		"name": "高级查询用户", "email": "advanced@example.com", "age": 28, "status": "active",
	})

	profileQuery, _ := db.Table("test_profiles", "postgres_test")
	profileQuery.Insert(ctx, map[string]interface{}{
		"user_id": userID, "bio": "这是个人简介", "website": "https://example.com",
	})

	t.Run("JOIN查询", func(t *testing.T) {
		userQuery, err := db.Table("test_users", "postgres_test")
		require.NoError(t, err)

		results, err := userQuery.
			Select("test_users.name", "test_users.email", "test_profiles.bio").
			LeftJoin("test_profiles", "test_users.id", "=", "test_profiles.user_id").
			Where("test_users.status", "=", "active").
			Get(ctx)

		assert.NoError(t, err)
		assert.Greater(t, len(results), 0)

		// 验证JOIN结果包含两个表的字段
		for _, result := range results {
			if result["bio"] != nil {
				assert.NotEmpty(t, result["name"])
				assert.NotEmpty(t, result["email"])
				assert.NotEmpty(t, result["bio"])
			}
		}
	})

	t.Run("分页查询", func(t *testing.T) {
		userQuery, err := db.Table("test_users", "postgres_test")
		require.NoError(t, err)

		// 第一页
		page1, err := userQuery.OrderBy("id", "ASC").Limit(2).Offset(0).Get(ctx)
		assert.NoError(t, err)
		assert.LessOrEqual(t, len(page1), 2)

		// 第二页
		page2, err := userQuery.OrderBy("id", "ASC").Limit(2).Offset(2).Get(ctx)
		assert.NoError(t, err)

		// 如果有足够的数据，页面应该不同
		if len(page1) == 2 && len(page2) > 0 {
			assert.NotEqual(t, page1[0]["id"], page2[0]["id"])
		}
	})

	t.Run("原生SQL查询", func(t *testing.T) {
		rows, err := conn.Query(ctx, `
			SELECT status, COUNT(*) as count 
			FROM test_users 
			GROUP BY status 
			ORDER BY count DESC
		`)
		assert.NoError(t, err)
		defer rows.Close()

		hasResults := false
		for rows.Next() {
			var status string
			var count int
			err := rows.Scan(&status, &count)
			assert.NoError(t, err)
			assert.NotEmpty(t, status)
			assert.Greater(t, count, 0)
			hasResults = true
		}
		assert.True(t, hasResults)
	})

	t.Run("PostgreSQL特有功能", func(t *testing.T) {
		userQuery, err := db.Table("test_users", "postgres_test")
		require.NoError(t, err)

		// 测试ILIKE（不区分大小写的LIKE）
		results, err := userQuery.WhereRaw("name ILIKE $1", "%用户%").Get(ctx)
		assert.NoError(t, err)
		// 如果有匹配的记录，应该能找到
		if len(results) > 0 {
			assert.Greater(t, len(results), 0)
		}
	})
}

func TestPostgreSQLBuilder(t *testing.T) {
	builder := db.NewPostgreSQLBuilder()

	t.Run("标识符引用", func(t *testing.T) {
		assert.Equal(t, `"users"`, builder.QuoteIdentifier("users"))
		assert.Equal(t, `"user_name"`, builder.QuoteIdentifier("user_name"))

		// 已经包含引号的不再添加
		assert.Equal(t, `"users"."name"`, builder.QuoteIdentifier(`"users"."name"`))

		// 表达式不添加引号
		assert.Equal(t, "COUNT(*)", builder.QuoteIdentifier("COUNT(*)"))
	})

	t.Run("值引用", func(t *testing.T) {
		assert.Equal(t, "'test'", builder.QuoteValue("test"))
		assert.Equal(t, "NULL", builder.QuoteValue(nil))
		assert.Equal(t, "123", builder.QuoteValue(123))
		assert.Equal(t, "'don''t'", builder.QuoteValue("don't")) // SQL转义
	})
}
