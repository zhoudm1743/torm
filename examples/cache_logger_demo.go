package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"torm/pkg/cache"
	"torm/pkg/db"
	"torm/pkg/logger"
	"torm/pkg/model"
)

func main() {
	fmt.Println("=== TORM 缓存和日志系统演示 ===")

	// 创建日志记录器
	appLogger := logger.NewLogger(logger.INFO)
	sqlLogger := logger.NewSQLLogger(logger.DEBUG, true)

	fmt.Println("✅ 日志系统初始化完成")

	// 创建缓存系统
	memoryCache := cache.NewMemoryCache()

	fmt.Println("✅ 缓存系统初始化完成")

	// 配置数据库连接
	config := &db.Config{
		Driver:          "mysql",
		Host:            "127.0.0.1",
		Port:            3306,
		Database:        "orm",
		Username:        "root",
		Password:        "123456",
		Charset:         "utf8mb4",
		Timezone:        "UTC",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		LogQueries:      true,
	}

	err := db.AddConnection("default", config)
	if err != nil {
		log.Fatal("Failed to add connection:", err)
	}

	// 获取管理器并设置缓存和日志
	manager := db.DefaultManager()
	manager.SetCache(memoryCache)
	manager.SetLogger(appLogger)

	sqlLogger.LogConnection("connect", config)

	ctx := context.Background()

	// 准备测试表
	conn, _ := db.DB("default")
	conn.Exec(ctx, "DROP TABLE IF EXISTS users")
	createTableSQL := `
	CREATE TABLE users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		age INT DEFAULT 0,
		status ENUM('active', 'inactive', 'pending') DEFAULT 'active',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`
	conn.Exec(ctx, createTableSQL)

	fmt.Println("✅ 数据库准备完成")

	// 演示1：日志系统
	fmt.Println("\n=== 1. 日志系统演示 ===")

	appLogger.Info("应用启动")
	appLogger.Debug("这是调试信息", "user_id", 123)
	appLogger.Warn("这是警告信息", "memory_usage", "85%")
	appLogger.Error("这是错误信息", "error", "连接超时")

	// 演示2：基础缓存操作
	fmt.Println("\n=== 2. 基础缓存操作演示 ===")

	// 设置缓存
	err = memoryCache.Set(ctx, "user:1", map[string]interface{}{
		"id":    1,
		"name":  "张三",
		"email": "zhangsan@example.com",
	}, 5*time.Minute)
	if err != nil {
		appLogger.Error("设置缓存失败", "error", err)
	} else {
		appLogger.Info("缓存设置成功", "key", "user:1")
	}

	// 获取缓存
	cachedUser, err := memoryCache.Get(ctx, "user:1")
	if err != nil {
		appLogger.Error("获取缓存失败", "error", err)
	} else {
		appLogger.Info("缓存获取成功", "data", cachedUser)
		fmt.Printf("缓存的用户数据: %v\n", cachedUser)
	}

	// 检查缓存是否存在
	exists, _ := memoryCache.Has(ctx, "user:1")
	fmt.Printf("缓存键 'user:1' 是否存在: %v\n", exists)

	exists, _ = memoryCache.Has(ctx, "user:999")
	fmt.Printf("缓存键 'user:999' 是否存在: %v\n", exists)

	// 演示3：缓存过期
	fmt.Println("\n=== 3. 缓存过期演示 ===")

	// 设置短期缓存
	err = memoryCache.Set(ctx, "temp:data", "这是临时数据", 2*time.Second)
	if err == nil {
		appLogger.Info("短期缓存设置成功", "ttl", "2秒")
	}

	// 立即获取
	tempData, err := memoryCache.Get(ctx, "temp:data")
	if err == nil {
		fmt.Printf("立即获取临时数据: %v\n", tempData)
	}

	// 等待过期
	fmt.Println("等待缓存过期...")
	time.Sleep(3 * time.Second)

	// 尝试再次获取
	tempData, err = memoryCache.Get(ctx, "temp:data")
	if err != nil {
		fmt.Printf("缓存已过期: %v\n", err)
		appLogger.Info("缓存过期验证成功")
	}

	// 演示4：结合数据库和缓存的用户查询
	fmt.Println("\n=== 4. 数据库+缓存演示 ===")

	// 先创建一些用户
	users := []*model.User{
		model.NewUser().SetName("缓存用户1").SetEmail("cache1@example.com").SetAge(25),
		model.NewUser().SetName("缓存用户2").SetEmail("cache2@example.com").SetAge(30),
		model.NewUser().SetName("缓存用户3").SetEmail("cache3@example.com").SetAge(28),
	}

	for i, user := range users {
		start := time.Now()
		err = user.Save(ctx)
		duration := time.Since(start)

		if err != nil {
			sqlLogger.LogQueryError("INSERT", []interface{}{}, err, duration)
		} else {
			sqlLogger.LogQuery("INSERT INTO users", []interface{}{user.GetName(), user.GetEmail(), user.GetAge()}, duration)
			appLogger.Info("用户创建成功", "id", user.GetID(), "name", user.GetName())

			// 将用户数据缓存起来
			cacheKey := fmt.Sprintf("user:%v", user.GetID())
			memoryCache.Set(ctx, cacheKey, user.ToMap(), 10*time.Minute)
			appLogger.Info("用户数据已缓存", "cache_key", cacheKey)
		}

		fmt.Printf("创建用户 %d: ID=%v, 姓名=%s\n", i+1, user.GetID(), user.GetName())
	}

	// 演示缓存查询
	fmt.Println("\n=== 5. 缓存查询演示 ===")

	userID := users[0].GetID()
	cacheKey := fmt.Sprintf("user:%v", userID)

	// 第一次查询 - 从缓存获取
	start := time.Now()
	cachedUserData, err := memoryCache.Get(ctx, cacheKey)
	cacheDuration := time.Since(start)

	if err == nil {
		appLogger.Info("从缓存获取用户成功", "duration", cacheDuration.String())
		fmt.Printf("从缓存获取用户 (耗时 %v): %v\n", cacheDuration, cachedUserData)
	}

	// 删除缓存，模拟缓存未命中
	memoryCache.Delete(ctx, cacheKey)
	appLogger.Info("缓存已清除", "cache_key", cacheKey)

	// 第二次查询 - 从数据库获取
	start = time.Now()
	dbUser := model.NewUser()
	err = dbUser.Find(ctx, userID)
	dbDuration := time.Since(start)

	if err == nil {
		sqlLogger.LogQuery("SELECT FROM users WHERE id = ?", []interface{}{userID}, dbDuration)
		appLogger.Info("从数据库获取用户成功", "duration", dbDuration.String())
		fmt.Printf("从数据库获取用户 (耗时 %v): 姓名=%s\n", dbDuration, dbUser.GetName())

		// 重新缓存
		memoryCache.Set(ctx, cacheKey, dbUser.ToMap(), 10*time.Minute)
		appLogger.Info("用户数据重新缓存", "cache_key", cacheKey)
	}

	fmt.Printf("性能对比: 缓存查询 %v vs 数据库查询 %v\n", cacheDuration, dbDuration)

	// 演示6：批量缓存操作
	fmt.Println("\n=== 6. 批量缓存操作演示 ===")

	// 批量设置缓存
	batchData := map[string]interface{}{
		"config:max_users":    1000,
		"config:cache_ttl":    300,
		"config:debug_mode":   true,
		"config:app_version":  "1.0.0",
		"stats:total_queries": 156,
		"stats:cache_hits":    89,
		"stats:cache_misses":  67,
	}

	for key, value := range batchData {
		err = memoryCache.Set(ctx, key, value, 30*time.Minute)
		if err == nil {
			appLogger.Debug("批量缓存设置", "key", key, "value", value)
		}
	}

	fmt.Printf("批量设置了 %d 个缓存项\n", len(batchData))

	// 获取缓存统计
	fmt.Printf("当前缓存项数量: %d\n", memoryCache.Size())
	fmt.Printf("缓存键列表: %v\n", memoryCache.Keys())

	// 演示7：缓存清理
	fmt.Println("\n=== 7. 缓存清理演示 ===")

	fmt.Printf("清理前缓存项数量: %d\n", memoryCache.Size())

	// 清空所有缓存
	err = memoryCache.Clear(ctx)
	if err == nil {
		appLogger.Info("缓存清空成功")
		fmt.Printf("清理后缓存项数量: %d\n", memoryCache.Size())
	}

	// 演示8：事务日志
	fmt.Println("\n=== 8. 事务日志演示 ===")

	start = time.Now()
	err = db.Transaction(ctx, func(tx db.TransactionInterface) error {
		sqlLogger.LogTransaction("BEGIN", 0)

		// 在事务中执行一些操作
		_, err := tx.Exec(ctx, "INSERT INTO users (name, email, age) VALUES (?, ?, ?)",
			"事务用户", "tx@example.com", 35)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, "UPDATE users SET age = age + 1 WHERE status = ?", "active")
		if err != nil {
			return err
		}

		return nil
	})

	txDuration := time.Since(start)

	if err != nil {
		sqlLogger.LogTransaction("ROLLBACK", txDuration)
		appLogger.Error("事务执行失败", "error", err, "duration", txDuration.String())
	} else {
		sqlLogger.LogTransaction("COMMIT", txDuration)
		appLogger.Info("事务执行成功", "duration", txDuration.String())
	}

	// 最终统计
	fmt.Println("\n=== 最终统计 ===")

	finalUserCount, _ := model.CountByStatus(ctx, "active")
	fmt.Printf("活跃用户总数: %d\n", finalUserCount)
	fmt.Printf("当前缓存项数量: %d\n", memoryCache.Size())

	appLogger.Info("演示完成", "final_user_count", finalUserCount)

	fmt.Println("\n🎉 TORM 缓存和日志系统演示完成！")
}
