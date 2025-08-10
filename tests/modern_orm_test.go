package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zhoudm1743/torm/pkg/db"
	"github.com/zhoudm1743/torm/pkg/migration"
)

// setupModernORM 设置现代化的ORM测试环境
func setupModernORM(t *testing.T) {
	config := &db.Config{
		Driver:          "sqlite",
		Database:        ":memory:",
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: time.Hour,
		LogQueries:      false,
	}

	err := db.AddConnection("default", config)
	require.NoError(t, err)

	conn, err := db.DB("default")
	require.NoError(t, err)

	err = conn.Connect()
	require.NoError(t, err)

	// 使用迁移工具创建表结构
	migrator := migration.NewMigrator(conn, nil)

	// 创建用户表
	migrator.RegisterFunc("20240101_000001", "创建用户表", func(conn db.ConnectionInterface) error {
		_, err := conn.Exec(`
			CREATE TABLE users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				email TEXT UNIQUE NOT NULL,
				age INTEGER DEFAULT 0,
				status TEXT DEFAULT 'active',
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)
		`)
		return err
	}, func(conn db.ConnectionInterface) error {
		_, err := conn.Exec("DROP TABLE IF EXISTS users")
		return err
	})

	// 创建文章表
	migrator.RegisterFunc("20240101_000002", "创建文章表", func(conn db.ConnectionInterface) error {
		_, err := conn.Exec(`
			CREATE TABLE posts (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				title TEXT NOT NULL,
				content TEXT NOT NULL,
				status TEXT DEFAULT 'draft',
				view_count INTEGER DEFAULT 0,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			)
		`)
		return err
	}, func(conn db.ConnectionInterface) error {
		_, err := conn.Exec("DROP TABLE IF EXISTS posts")
		return err
	})

	// 执行迁移
	err = migrator.Up()
	require.NoError(t, err)
}

// TestModernORM_QueryBuilder 测试使用查询构建器而不是原始SQL
func TestModernORM_QueryBuilder(t *testing.T) {
	setupModernORM(t)

	// 使用查询构建器插入用户数据
	userQuery, err := db.Table("users")
	require.NoError(t, err)

	userID, err := userQuery.Insert(map[string]interface{}{
		"name":   "张三",
		"email":  "zhangsan@example.com",
		"age":    28,
		"status": "active",
	})
	require.NoError(t, err)
	assert.NotZero(t, userID)

	// 使用查询构建器查询用户
	userData, err := userQuery.Where("email", "=", "zhangsan@example.com").First()
	require.NoError(t, err)
	assert.Equal(t, "张三", userData["name"])
	assert.Equal(t, "zhangsan@example.com", userData["email"])
	assert.Equal(t, int64(28), userData["age"]) // SQLite返回int64

	// 使用查询构建器插入文章数据
	postQuery, err := db.Table("posts")
	require.NoError(t, err)

	postID, err := postQuery.Insert(map[string]interface{}{
		"user_id":    userID,
		"title":      "我的第一篇文章",
		"content":    "这是文章内容",
		"status":     "published",
		"view_count": 100,
	})
	require.NoError(t, err)
	assert.NotZero(t, postID)

	// 使用查询构建器进行关联查询
	results, err := userQuery.
		Select("users.name", "posts.title", "posts.view_count").
		InnerJoin("posts", "users.id", "=", "posts.user_id").
		Where("posts.status", "=", "published").
		Get()
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "张三", results[0]["name"])
	assert.Equal(t, "我的第一篇文章", results[0]["title"])
	assert.Equal(t, int64(100), results[0]["view_count"])
}

// TestModernORM_AdvancedQueries 测试高级查询功能
func TestModernORM_AdvancedQueries(t *testing.T) {
	setupModernORM(t)

	// 批量插入用户数据
	userQuery, err := db.Table("users")
	require.NoError(t, err)

	users := []map[string]interface{}{
		{"name": "用户1", "email": "user1@example.com", "age": 25, "status": "active"},
		{"name": "用户2", "email": "user2@example.com", "age": 30, "status": "inactive"},
		{"name": "用户3", "email": "user3@example.com", "age": 35, "status": "active"},
		{"name": "用户4", "email": "user4@example.com", "age": 28, "status": "pending"},
	}

	_, err = userQuery.InsertBatch(users)
	require.NoError(t, err)

	// 测试条件查询
	activeUsers, err := userQuery.Where("status", "=", "active").Get()
	require.NoError(t, err)
	assert.Len(t, activeUsers, 2)

	// 测试范围查询
	adultUsers, err := userQuery.Where("age", ">=", 30).Get()
	require.NoError(t, err)
	assert.Len(t, adultUsers, 2)

	// 测试IN查询
	specificUsers, err := userQuery.WhereIn("status", []interface{}{"active", "pending"}).Get()
	require.NoError(t, err)
	assert.Len(t, specificUsers, 3)

	// 测试计数查询
	totalCount, err := userQuery.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(4), totalCount)

	// 测试状态计数
	activeCount, err := userQuery.Where("status", "=", "active").Count()
	require.NoError(t, err)
	assert.Equal(t, int64(2), activeCount)

	// 测试排序和限制
	sortedUsers, err := userQuery.OrderBy("age", "desc").Limit(2).Get()
	require.NoError(t, err)
	assert.Len(t, sortedUsers, 2)
	assert.Equal(t, int64(35), sortedUsers[0]["age"]) // 年龄最大的用户

	// 测试更新操作
	affected, err := userQuery.Where("status", "=", "pending").Update(map[string]interface{}{
		"status": "active",
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), affected)

	// 验证更新结果
	newActiveCount, err := userQuery.Where("status", "=", "active").Count()
	require.NoError(t, err)
	assert.Equal(t, int64(3), newActiveCount)
}

// TestModernORM_Transactions 测试事务功能
func TestModernORM_Transactions(t *testing.T) {
	setupModernORM(t)

	// 事务成功案例
	err := db.Transaction(func(tx db.TransactionInterface) error {
		// 在事务中插入用户
		result, err := tx.Exec(`
			INSERT INTO users (name, email, age, status) 
			VALUES (?, ?, ?, ?)
		`, "事务用户", "transaction@example.com", 25, "active")
		if err != nil {
			return err
		}

		userID, err := result.LastInsertId()
		if err != nil {
			return err
		}

		// 在事务中插入文章
		_, err = tx.Exec(`
			INSERT INTO posts (user_id, title, content, status) 
			VALUES (?, ?, ?, ?)
		`, userID, "事务文章", "这是在事务中创建的文章", "published")
		return err
	})
	require.NoError(t, err)

	// 验证事务数据已提交
	userQuery, err := db.Table("users")
	require.NoError(t, err)

	user, err := userQuery.Where("email", "=", "transaction@example.com").First()
	require.NoError(t, err)
	assert.Equal(t, "事务用户", user["name"])

	postQuery, err := db.Table("posts")
	require.NoError(t, err)

	post, err := postQuery.Where("title", "=", "事务文章").First()
	require.NoError(t, err)
	assert.Equal(t, "这是在事务中创建的文章", post["content"])

	// 事务回滚案例
	initialUserCount, err := userQuery.Count()
	require.NoError(t, err)

	err = db.Transaction(func(tx db.TransactionInterface) error {
		// 插入用户
		_, err := tx.Exec(`
			INSERT INTO users (name, email, age, status) 
			VALUES (?, ?, ?, ?)
		`, "回滚用户", "rollback@example.com", 30, "active")
		if err != nil {
			return err
		}

		// 模拟错误，触发回滚
		return assert.AnError
	})
	assert.Error(t, err)

	// 验证事务已回滚
	finalUserCount, err := userQuery.Count()
	require.NoError(t, err)
	assert.Equal(t, initialUserCount, finalUserCount)

	// 验证回滚用户不存在
	_, err = userQuery.Where("email", "=", "rollback@example.com").First()
	assert.Error(t, err) // 应该找不到记录
}

// TestModernORM_WithTimeout 测试WithTimeout功能
func TestModernORM_WithTimeout(t *testing.T) {
	setupModernORM(t)

	// 插入测试数据
	userQuery, err := db.Table("users")
	require.NoError(t, err)

	_, err = userQuery.Insert(map[string]interface{}{
		"name":   "超时测试用户",
		"email":  "timeout@example.com",
		"age":    25,
		"status": "active",
	})
	require.NoError(t, err)

	// 测试带超时的查询
	results, err := userQuery.
		WithTimeout(5*time.Second).
		Where("email", "=", "timeout@example.com").
		Get()
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "超时测试用户", results[0]["name"])

	// 测试WithContext功能 (使用默认context)
	user, err := userQuery.
		WithContext(nil). // 传入nil会使用context.Background()
		Where("status", "=", "active").
		First()
	require.NoError(t, err)
	assert.Equal(t, "超时测试用户", user["name"])
}

// TestModernORM_ComplexJoins 测试复杂关联查询
func TestModernORM_ComplexJoins(t *testing.T) {
	setupModernORM(t)

	// 创建测试数据
	userQuery, err := db.Table("users")
	require.NoError(t, err)

	// 插入多个用户
	userIDs := make([]interface{}, 0)
	for i := 1; i <= 3; i++ {
		userID, err := userQuery.Insert(map[string]interface{}{
			"name":   fmt.Sprintf("作者%d", i),
			"email":  fmt.Sprintf("author%d@example.com", i),
			"age":    20 + i*5,
			"status": "active",
		})
		require.NoError(t, err)
		userIDs = append(userIDs, userID)
	}

	// 为每个用户创建多篇文章
	postQuery, err := db.Table("posts")
	require.NoError(t, err)

	for i, userID := range userIDs {
		for j := 1; j <= 2; j++ {
			_, err := postQuery.Insert(map[string]interface{}{
				"user_id":    userID,
				"title":      fmt.Sprintf("作者%d的第%d篇文章", i+1, j),
				"content":    fmt.Sprintf("这是作者%d的第%d篇文章内容", i+1, j),
				"status":     "published",
				"view_count": (i + 1) * j * 50, // 不同的浏览量
			})
			require.NoError(t, err)
		}
	}

	// 测试内连接查询 - 获取所有作者及其文章
	results, err := userQuery.
		Select("users.name as author_name", "posts.title", "posts.view_count").
		InnerJoin("posts", "users.id", "=", "posts.user_id").
		OrderBy("posts.view_count", "desc").
		Get()
	require.NoError(t, err)
	assert.Len(t, results, 6) // 3个用户 × 2篇文章 = 6条记录

	// 验证最高浏览量的文章
	assert.Equal(t, "作者3", results[0]["author_name"])
	assert.Equal(t, int64(300), results[0]["view_count"])

	// 测试左连接 - 即使用户没有文章也显示
	// 先插入一个没有文章的用户
	_, err = userQuery.Insert(map[string]interface{}{
		"name":   "无文章作者",
		"email":  "no-posts@example.com",
		"age":    40,
		"status": "active",
	})
	require.NoError(t, err)

	leftJoinResults, err := userQuery.
		Select("users.name", "posts.title").
		LeftJoin("posts", "users.id", "=", "posts.user_id").
		Where("users.status", "=", "active").
		OrderBy("users.name", "asc").
		Get()
	require.NoError(t, err)
	assert.Greater(t, len(leftJoinResults), 6) // 应该包含没有文章的用户

	// 测试聚合查询 - 每个作者的文章统计
	statsResults, err := userQuery.
		Select("users.name", "COUNT(posts.id) as post_count", "AVG(posts.view_count) as avg_views").
		LeftJoin("posts", "users.id", "=", "posts.user_id").
		GroupBy("users.id", "users.name").
		Having("post_count", ">", 0).
		OrderBy("avg_views", "desc").
		Get()
	require.NoError(t, err)
	assert.Len(t, statsResults, 3) // 只有有文章的3个作者

	// 验证统计数据
	for _, stat := range statsResults {
		assert.Equal(t, int64(2), stat["post_count"])    // 每个作者2篇文章
		assert.Greater(t, stat["avg_views"], float64(0)) // 平均浏览量大于0
	}
}
