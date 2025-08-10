package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"torm/pkg/db"
)

func setupMongoDBTest(t *testing.T) (context.Context, *db.MongoDBConnection) {
	// 配置MongoDB测试连接
	config := &db.Config{
		Driver:          "mongodb",
		Host:            "127.0.0.1",
		Port:            27017,
		Database:        "torm_test",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		LogQueries:      false,
	}

	err := db.AddConnection("mongodb_test", config)
	require.NoError(t, err)

	ctx := context.Background()
	conn, err := db.DB("mongodb_test")
	require.NoError(t, err)

	// 测试连接
	err = conn.Ping(ctx)
	if err != nil {
		t.Skipf("MongoDB not available: %v", err)
	}

	mongoConn := db.GetMongoConnection(conn)
	require.NotNil(t, mongoConn)

	// 清理测试集合
	collection := mongoConn.GetCollection("test_users")
	collection.Drop(ctx)

	return ctx, mongoConn
}

func TestMongoDBConnection(t *testing.T) {
	ctx, mongoConn := setupMongoDBTest(t)

	t.Run("连接测试", func(t *testing.T) {
		err := mongoConn.Ping(ctx)
		assert.NoError(t, err)
		assert.True(t, mongoConn.IsConnected())
		assert.Equal(t, "mongodb", mongoConn.GetDriver())
	})

	t.Run("配置测试", func(t *testing.T) {
		config := mongoConn.GetConfig()
		assert.NotNil(t, config)
		assert.Equal(t, "mongodb", config.Driver)
		assert.Equal(t, "127.0.0.1", config.Host)
		assert.Equal(t, 27017, config.Port)
	})

	t.Run("数据库和集合操作", func(t *testing.T) {
		database := mongoConn.GetDatabase()
		assert.NotNil(t, database)
		assert.Equal(t, "torm_test", database.Name())

		collection := mongoConn.GetCollection("test_collection")
		assert.NotNil(t, collection)
		assert.Equal(t, "test_collection", collection.Name())
	})
}

func TestMongoDBCRUD(t *testing.T) {
	ctx, mongoConn := setupMongoDBTest(t)

	collection := mongoConn.GetCollection("test_users")
	query := db.NewMongoQuery(collection, nil)

	t.Run("插入数据", func(t *testing.T) {
		user := bson.M{
			"name":       "MongoDB测试用户",
			"email":      "mongodb_test@example.com",
			"age":        25,
			"status":     "active",
			"created_at": time.Now(),
		}

		result, err := query.InsertOne(ctx, user)
		assert.NoError(t, err)
		assert.NotNil(t, result.InsertedID)
	})

	t.Run("查询数据", func(t *testing.T) {
		// 查询所有数据
		cursor, err := query.Find(ctx)
		assert.NoError(t, err)
		defer cursor.Close(ctx)

		var users []bson.M
		err = db.ToArray(ctx, cursor, &users)
		assert.NoError(t, err)
		assert.Greater(t, len(users), 0)

		// 条件查询
		activeCursor, err := query.Where("status", "active").Find(ctx)
		assert.NoError(t, err)
		defer activeCursor.Close(ctx)

		var activeUsers []bson.M
		err = db.ToArray(ctx, activeCursor, &activeUsers)
		assert.NoError(t, err)
		assert.Greater(t, len(activeUsers), 0)

		// 查询单条记录
		var foundUser bson.M
		err = query.Where("email", "mongodb_test@example.com").FindOne(ctx).Decode(&foundUser)
		assert.NoError(t, err)
		assert.Equal(t, "MongoDB测试用户", foundUser["name"])
	})

	t.Run("更新数据", func(t *testing.T) {
		update := bson.M{
			"$set": bson.M{
				"age":        26,
				"updated_at": time.Now(),
			},
		}

		result, err := query.Where("email", "mongodb_test@example.com").UpdateOne(ctx, update)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), result.ModifiedCount)

		// 验证更新
		var updatedUser bson.M
		err = query.Where("email", "mongodb_test@example.com").FindOne(ctx).Decode(&updatedUser)
		assert.NoError(t, err)
		assert.Equal(t, int32(26), updatedUser["age"])
	})

	t.Run("批量插入", func(t *testing.T) {
		batchUsers := []interface{}{
			bson.M{"name": "批量用户1", "email": "batch1@example.com", "age": 20, "status": "active", "created_at": time.Now()},
			bson.M{"name": "批量用户2", "email": "batch2@example.com", "age": 21, "status": "inactive", "created_at": time.Now()},
			bson.M{"name": "批量用户3", "email": "batch3@example.com", "age": 22, "status": "active", "created_at": time.Now()},
		}

		result, err := query.InsertMany(ctx, batchUsers)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(result.InsertedIDs))
	})

	t.Run("聚合查询", func(t *testing.T) {
		// 计数
		count, err := query.Count(ctx)
		assert.NoError(t, err)
		assert.Greater(t, count, int64(0))

		// 条件计数
		activeCount, err := query.Where("status", "active").Count(ctx)
		assert.NoError(t, err)
		assert.Greater(t, activeCount, int64(0))

		// Distinct查询
		distinctValues, err := query.Distinct(ctx, "status")
		assert.NoError(t, err)
		assert.Greater(t, len(distinctValues), 0)
	})

	t.Run("删除数据", func(t *testing.T) {
		deleteQuery := db.NewMongoQuery(collection, nil)
		result, err := deleteQuery.Where("email", "batch2@example.com").DeleteOne(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), result.DeletedCount)

		// 验证删除
		verifyQuery := db.NewMongoQuery(collection, nil)
		var deletedUser bson.M
		err = verifyQuery.Where("email", "batch2@example.com").FindOne(ctx).Decode(&deletedUser)
		assert.Error(t, err) // 应该找不到记录
	})
}

func TestMongoDBAdvancedQueries(t *testing.T) {
	ctx, mongoConn := setupMongoDBTest(t)

	collection := mongoConn.GetCollection("advanced_test")
	query := db.NewMongoQuery(collection, nil)

	// 准备测试数据
	testData := []interface{}{
		bson.M{"name": "Alice", "age": 25, "city": "北京", "score": 85},
		bson.M{"name": "Bob", "age": 30, "city": "上海", "score": 92},
		bson.M{"name": "Charlie", "age": 35, "city": "广州", "score": 78},
		bson.M{"name": "David", "age": 28, "city": "深圳", "score": 88},
	}

	_, err := query.InsertMany(ctx, testData)
	require.NoError(t, err)

	t.Run("范围查询", func(t *testing.T) {
		rangeQuery := db.NewMongoQuery(collection, nil)
		cursor, err := rangeQuery.WhereGte("age", 28).WhereLte("age", 32).Find(ctx)
		assert.NoError(t, err)
		defer cursor.Close(ctx)

		var users []bson.M
		err = db.ToArray(ctx, cursor, &users)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(users)) // Bob and David
	})

	t.Run("IN查询", func(t *testing.T) {
		inQuery := db.NewMongoQuery(collection, nil)
		cities := []interface{}{"北京", "上海"}
		cursor, err := inQuery.WhereIn("city", cities).Find(ctx)
		assert.NoError(t, err)
		defer cursor.Close(ctx)

		var users []bson.M
		err = db.ToArray(ctx, cursor, &users)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(users)) // Alice and Bob
	})

	t.Run("正则表达式查询", func(t *testing.T) {
		regexQuery := db.NewMongoQuery(collection, nil)
		cursor, err := regexQuery.WhereRegex("name", "^A", "i").Find(ctx)
		assert.NoError(t, err)
		defer cursor.Close(ctx)

		var users []bson.M
		err = db.ToArray(ctx, cursor, &users)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(users)) // Alice
		assert.Equal(t, "Alice", users[0]["name"])
	})

	t.Run("排序和分页", func(t *testing.T) {
		// 按年龄升序排序，获取前2条
		sortQuery := db.NewMongoQuery(collection, nil)
		cursor, err := sortQuery.OrderByAsc("age").Limit(2).Find(ctx)
		assert.NoError(t, err)
		defer cursor.Close(ctx)

		var users []bson.M
		err = db.ToArray(ctx, cursor, &users)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(users), 1) // 至少有1条记录
		if len(users) >= 1 {
			assert.Equal(t, "Alice", users[0]["name"]) // 25岁，最小
		}
		if len(users) >= 2 {
			assert.Equal(t, "David", users[1]["name"]) // 28岁，第二小
		}
	})

	t.Run("字段选择", func(t *testing.T) {
		cursor, err := query.Select("name", "age").Find(ctx)
		assert.NoError(t, err)
		defer cursor.Close(ctx)

		var users []bson.M
		err = db.ToArray(ctx, cursor, &users)
		assert.NoError(t, err)
		assert.Greater(t, len(users), 0)

		// 验证只返回了指定字段
		for _, user := range users {
			assert.Contains(t, user, "_id") // MongoDB总是返回_id
			assert.Contains(t, user, "name")
			assert.Contains(t, user, "age")
			assert.NotContains(t, user, "city")  // 应该被排除
			assert.NotContains(t, user, "score") // 应该被排除
		}
	})

	t.Run("聚合管道", func(t *testing.T) {
		pipeline := []bson.M{
			{
				"$group": bson.M{
					"_id":       "$city",
					"avg_age":   bson.M{"$avg": "$age"},
					"avg_score": bson.M{"$avg": "$score"},
					"count":     bson.M{"$sum": 1},
				},
			},
			{
				"$sort": bson.M{"avg_score": -1},
			},
		}

		cursor, err := query.Aggregate(ctx, pipeline)
		assert.NoError(t, err)
		defer cursor.Close(ctx)

		var results []bson.M
		err = db.ToArray(ctx, cursor, &results)
		assert.NoError(t, err)
		assert.Greater(t, len(results), 0)

		// 验证聚合结果包含期望的字段
		for _, result := range results {
			assert.Contains(t, result, "_id")       // 城市
			assert.Contains(t, result, "avg_age")   // 平均年龄
			assert.Contains(t, result, "avg_score") // 平均分数
			assert.Contains(t, result, "count")     // 计数
		}
	})
}

func TestMongoDBTransaction(t *testing.T) {
	ctx, mongoConn := setupMongoDBTest(t)

	t.Run("事务提交", func(t *testing.T) {
		session, err := mongoConn.BeginMongo(ctx)
		require.NoError(t, err)

		_, err = session.GetSession().WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
			// 获取集合
			usersCollection := mongoConn.GetCollection("tx_users")
			logsCollection := mongoConn.GetCollection("tx_logs")

			// 在事务中插入用户
			userQuery := db.NewMongoQuery(usersCollection, nil).WithSession(session.GetSession())
			user := bson.M{
				"name":       "事务用户",
				"email":      "tx_user@example.com",
				"created_at": time.Now(),
			}

			insertResult, err := userQuery.InsertOne(ctx, user)
			if err != nil {
				return nil, err
			}

			// 在事务中插入日志
			logQuery := db.NewMongoQuery(logsCollection, nil).WithSession(session.GetSession())
			logEntry := bson.M{
				"user_id":   insertResult.InsertedID,
				"operation": "user_created",
				"timestamp": time.Now(),
			}

			_, err = logQuery.InsertOne(ctx, logEntry)
			return nil, err
		})

		assert.NoError(t, err)

		// 验证数据已插入
		usersCollection := mongoConn.GetCollection("tx_users")
		userQuery := db.NewMongoQuery(usersCollection, nil)
		var user bson.M
		err = userQuery.Where("email", "tx_user@example.com").FindOne(ctx).Decode(&user)
		assert.NoError(t, err)
		assert.Equal(t, "事务用户", user["name"])

		logsCollection := mongoConn.GetCollection("tx_logs")
		logQuery := db.NewMongoQuery(logsCollection, nil)
		count, err := logQuery.Where("operation", "user_created").Count(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)

		session.Commit()
	})
}

func TestMongoDBQueryBuilder(t *testing.T) {
	ctx, mongoConn := setupMongoDBTest(t)

	collection := mongoConn.GetCollection("query_test")
	query := db.NewMongoQuery(collection, nil)

	t.Run("查询构建器链式调用", func(t *testing.T) {
		// 测试链式调用不会panic
		assert.NotPanics(t, func() {
			query.Where("status", "active").
				WhereGte("age", 18).
				WhereLte("age", 65).
				OrderByDesc("created_at").
				Skip(0).
				Limit(10).
				Select("name", "email", "age")
		})
	})

	t.Run("查询条件组合", func(t *testing.T) {
		// 插入测试数据
		testUsers := []interface{}{
			bson.M{"name": "用户A", "age": 20, "status": "active"},
			bson.M{"name": "用户B", "age": 30, "status": "inactive"},
			bson.M{"name": "用户C", "age": 40, "status": "active"},
		}

		_, err := query.InsertMany(ctx, testUsers)
		require.NoError(t, err)

		// 复合条件查询
		cursor, err := query.
			Where("status", "active").
			WhereGte("age", 25).
			OrderByAsc("age").
			Find(ctx)

		assert.NoError(t, err)
		defer cursor.Close(ctx)

		var users []bson.M
		err = db.ToArray(ctx, cursor, &users)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(users)) // 只有用户C满足条件
		assert.Equal(t, "用户C", users[0]["name"])
	})
}

func TestMongoDBErrorHandling(t *testing.T) {
	ctx, mongoConn := setupMongoDBTest(t)

	t.Run("无效集合操作", func(t *testing.T) {
		collection := mongoConn.GetCollection("nonexistent")
		query := db.NewMongoQuery(collection, nil)

		// 查询不存在的文档
		var result bson.M
		err := query.Where("_id", "nonexistent").FindOne(ctx).Decode(&result)
		assert.Error(t, err) // 应该返回错误
	})

	t.Run("连接状态检查", func(t *testing.T) {
		assert.True(t, mongoConn.IsConnected())

		// 关闭连接
		err := mongoConn.Close()
		assert.NoError(t, err)

		// 连接应该已关闭
		assert.False(t, mongoConn.IsConnected())
	})
}
