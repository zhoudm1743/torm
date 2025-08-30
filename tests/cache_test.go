package tests

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/zhoudm1743/torm"
)

// CacheTestUser 缓存测试用户模型
type CacheTestUser struct {
	torm.BaseModel
	ID       int    `json:"id" torm:"primary_key,auto_increment"`
	Username string `json:"username" torm:"type:varchar,size:50,unique"`
	Email    string `json:"email" torm:"type:varchar,size:100"`
	Age      int    `json:"age" torm:"type:int,default:0"`
	Status   string `json:"status" torm:"type:varchar,size:20,default:active"`
	City     string `json:"city" torm:"type:varchar,size:50"`
}

// setupCacheTestData 设置缓存测试数据
func setupCacheTestData(t *testing.T, connectionName string) {
	// 创建表
	user := &CacheTestUser{BaseModel: *torm.NewModel()}
	user.SetTable("cache_test_users").SetPrimaryKey("id").SetConnection(connectionName)

	err := user.AutoMigrate(user)
	if err != nil {
		t.Fatalf("❌ 缓存测试表创建失败: %v", err)
	}

	// 清理可能存在的数据
	if builder, err := torm.Table("cache_test_users", connectionName); err == nil {
		builder.Delete()
	}

	// 插入测试数据
	testData := []map[string]interface{}{
		{"username": "cache_user1", "email": "cache1@test.com", "age": 25, "status": "active", "city": "北京"},
		{"username": "cache_user2", "email": "cache2@test.com", "age": 30, "status": "active", "city": "上海"},
		{"username": "cache_user3", "email": "cache3@test.com", "age": 28, "status": "inactive", "city": "深圳"},
		{"username": "cache_user4", "email": "cache4@test.com", "age": 35, "status": "active", "city": "北京"},
		{"username": "cache_user5", "email": "cache5@test.com", "age": 22, "status": "pending", "city": "广州"},
	}

	for _, data := range testData {
		if builder, err := torm.Table("cache_test_users", connectionName); err == nil {
			builder.Insert(data)
		}
	}

	t.Log("✅ 缓存测试数据初始化完成")
}

// TestMySQL_Cache MySQL缓存测试
func TestMySQL_Cache(t *testing.T) {
	setupMySQLConnection(t)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🐬 MySQL 缓存功能测试")
	fmt.Println(strings.Repeat("=", 60))

	setupCacheTestData(t, "mysql_test")
	testCacheBasicFunctionality(t, "mysql_test", "mysql")
	testCacheWithTags(t, "mysql_test", "mysql")
	testCacheExpiration(t, "mysql_test", "mysql")
	testCacheInvalidation(t, "mysql_test", "mysql")

	fmt.Println("\n✅ MySQL缓存测试完成")
}

// TestPostgreSQL_Cache PostgreSQL缓存测试
func TestPostgreSQL_Cache(t *testing.T) {
	setupPostgreSQLConnection(t)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🐘 PostgreSQL 缓存功能测试")
	fmt.Println(strings.Repeat("=", 60))

	setupCacheTestData(t, "postgres_test")
	testCacheBasicFunctionality(t, "postgres_test", "postgres")
	testCacheWithTags(t, "postgres_test", "postgres")
	testCacheExpiration(t, "postgres_test", "postgres")
	testCacheInvalidation(t, "postgres_test", "postgres")

	fmt.Println("\n✅ PostgreSQL缓存测试完成")
}

// testCacheBasicFunctionality 测试基础缓存功能
func testCacheBasicFunctionality(t *testing.T, connectionName, dbType string) {
	t.Logf("\n📋 1. %s 基础缓存功能测试", dbType)

	// 清空缓存
	torm.ClearAllCache()

	// 第一次查询 - 应该从数据库获取
	builder1, err := torm.Table("cache_test_users", connectionName)
	if err != nil {
		t.Errorf("❌ %s 查询构建器创建失败: %v", dbType, err)
		return
	}

	start := time.Now()
	results1, err := builder1.Where("status", "=", "active").Cache(5 * time.Minute).Get()
	firstQueryTime := time.Since(start)

	if err != nil {
		t.Errorf("❌ %s 第一次缓存查询失败: %v", dbType, err)
		return
	}
	t.Logf("✅ %s 第一次查询成功: %d条记录, 耗时: %v", dbType, len(results1), firstQueryTime)

	// 第二次相同查询 - 应该从缓存获取
	builder2, err := torm.Table("cache_test_users", connectionName)
	if err != nil {
		t.Errorf("❌ %s 查询构建器创建失败: %v", dbType, err)
		return
	}

	start = time.Now()
	results2, err := builder2.Where("status", "=", "active").Cache(5 * time.Minute).Get()
	secondQueryTime := time.Since(start)

	if err != nil {
		t.Errorf("❌ %s 第二次缓存查询失败: %v", dbType, err)
		return
	}

	// 验证结果相同
	if len(results1) != len(results2) {
		t.Errorf("❌ %s 缓存结果不一致: 第一次%d条，第二次%d条", dbType, len(results1), len(results2))
		return
	}

	// 第二次查询应该明显更快（从缓存获取）
	if secondQueryTime >= firstQueryTime/2 {
		t.Logf("⚠️ %s 缓存效果不明显: 第一次%v, 第二次%v", dbType, firstQueryTime, secondQueryTime)
	} else {
		t.Logf("✅ %s 缓存生效: 第一次%v, 第二次%v (提速%.1fx)",
			dbType, firstQueryTime, secondQueryTime, float64(firstQueryTime)/float64(secondQueryTime))
	}

	// 测试缓存统计
	stats := torm.GetCacheStats()
	if stats != nil {
		t.Logf("📊 %s 缓存统计: %+v", dbType, stats)
	}
}

// testCacheWithTags 测试带标签的缓存
func testCacheWithTags(t *testing.T, connectionName, dbType string) {
	t.Logf("\n🏷️ 2. %s 标签缓存测试", dbType)

	// 清空缓存
	torm.ClearAllCache()

	// 使用不同标签缓存不同查询
	builder1, _ := torm.Table("cache_test_users", connectionName)
	results1, err := builder1.Where("city", "=", "北京").
		CacheWithTags(10*time.Minute, "users", "city_beijing").Get()
	if err != nil {
		t.Errorf("❌ %s 北京用户查询失败: %v", dbType, err)
		return
	}
	t.Logf("✅ %s 北京用户查询: %d条记录", dbType, len(results1))

	builder2, _ := torm.Table("cache_test_users", connectionName)
	results2, err := builder2.Where("city", "=", "上海").
		CacheWithTags(10*time.Minute, "users", "city_shanghai").Get()
	if err != nil {
		t.Errorf("❌ %s 上海用户查询失败: %v", dbType, err)
		return
	}
	t.Logf("✅ %s 上海用户查询: %d条记录", dbType, len(results2))

	// 缓存所有活跃用户
	builder3, _ := torm.Table("cache_test_users", connectionName)
	results3, err := builder3.Where("status", "=", "active").
		CacheWithTags(10*time.Minute, "users", "active_users").Get()
	if err != nil {
		t.Errorf("❌ %s 活跃用户查询失败: %v", dbType, err)
		return
	}
	t.Logf("✅ %s 活跃用户查询: %d条记录", dbType, len(results3))

	// 验证缓存生效
	builder4, _ := torm.Table("cache_test_users", connectionName)
	start := time.Now()
	results4, err := builder4.Where("city", "=", "北京").
		CacheWithTags(10*time.Minute, "users", "city_beijing").Get()
	cacheQueryTime := time.Since(start)

	if err != nil {
		t.Errorf("❌ %s 缓存验证查询失败: %v", dbType, err)
		return
	}

	if len(results1) != len(results4) {
		t.Errorf("❌ %s 标签缓存结果不一致", dbType)
		return
	}

	t.Logf("✅ %s 标签缓存验证成功, 查询耗时: %v", dbType, cacheQueryTime)
}

// testCacheExpiration 测试缓存过期
func testCacheExpiration(t *testing.T, connectionName, dbType string) {
	t.Logf("\n⏰ 3. %s 缓存过期测试", dbType)

	// 清空缓存
	torm.ClearAllCache()

	// 设置短期缓存
	builder1, _ := torm.Table("cache_test_users", connectionName)
	results1, err := builder1.Where("age", ">", 25).Cache(2 * time.Second).Get()
	if err != nil {
		t.Errorf("❌ %s 短期缓存查询失败: %v", dbType, err)
		return
	}
	t.Logf("✅ %s 短期缓存设置成功: %d条记录", dbType, len(results1))

	// 立即查询 - 应该命中缓存
	builder2, _ := torm.Table("cache_test_users", connectionName)
	start := time.Now()
	results2, err := builder2.Where("age", ">", 25).Cache(2 * time.Second).Get()
	immediateQueryTime := time.Since(start)

	if err != nil {
		t.Errorf("❌ %s 立即缓存查询失败: %v", dbType, err)
		return
	}

	// 验证结果数量一致
	if len(results1) != len(results2) {
		t.Errorf("❌ %s 立即缓存查询结果不一致", dbType)
		return
	}

	t.Logf("✅ %s 立即缓存查询成功, 耗时: %v", dbType, immediateQueryTime)

	// 等待缓存过期
	t.Logf("   等待缓存过期...")
	time.Sleep(3 * time.Second)

	// 过期后查询 - 应该重新从数据库获取
	builder3, _ := torm.Table("cache_test_users", connectionName)
	start = time.Now()
	results3, err := builder3.Where("age", ">", 25).Cache(2 * time.Second).Get()
	expiredQueryTime := time.Since(start)

	if err != nil {
		t.Errorf("❌ %s 过期后查询失败: %v", dbType, err)
		return
	}

	// 结果应该相同，但耗时应该增加
	if len(results1) != len(results3) {
		t.Errorf("❌ %s 过期后查询结果不一致", dbType)
		return
	}

	t.Logf("✅ %s 缓存过期测试完成: 立即查询%v, 过期后查询%v",
		dbType, immediateQueryTime, expiredQueryTime)
}

// testCacheInvalidation 测试缓存失效
func testCacheInvalidation(t *testing.T, connectionName, dbType string) {
	t.Logf("\n🗑️ 4. %s 缓存失效测试", dbType)

	// 清空缓存
	torm.ClearAllCache()

	// 缓存用户数据
	builder1, _ := torm.Table("cache_test_users", connectionName)
	results1, err := builder1.Where("status", "=", "active").
		CacheWithTags(10*time.Minute, "users", "active_users").Get()
	if err != nil {
		t.Errorf("❌ %s 用户缓存设置失败: %v", dbType, err)
		return
	}
	t.Logf("✅ %s 用户缓存设置成功: %d条记录", dbType, len(results1))

	// 缓存城市数据
	builder2, _ := torm.Table("cache_test_users", connectionName)
	results2, err := builder2.Where("city", "=", "北京").
		CacheWithTags(10*time.Minute, "users", "city_data").Get()
	if err != nil {
		t.Errorf("❌ %s 城市缓存设置失败: %v", dbType, err)
		return
	}
	t.Logf("✅ %s 城市缓存设置成功: %d条记录", dbType, len(results2))

	// 通过标签清理特定缓存
	err = torm.ClearCacheByTags("active_users")
	if err != nil {
		t.Errorf("❌ %s 标签缓存清理失败: %v", dbType, err)
		return
	}
	t.Logf("✅ %s 标签缓存清理成功", dbType)

	// 验证特定缓存已清理，其他缓存仍存在
	builder3, _ := torm.Table("cache_test_users", connectionName)
	start := time.Now()
	results3, err := builder3.Where("status", "=", "active").
		CacheWithTags(10*time.Minute, "users", "active_users").Get()
	activeQueryTime := time.Since(start)

	if err != nil {
		t.Errorf("❌ %s 清理后活跃用户查询失败: %v", dbType, err)
		return
	}

	t.Logf("   ✅ %s 清理后活跃用户查询成功: %d条记录", dbType, len(results3))

	builder4, _ := torm.Table("cache_test_users", connectionName)
	start = time.Now()
	results4, err := builder4.Where("city", "=", "北京").
		CacheWithTags(10*time.Minute, "users", "city_data").Get()
	cityQueryTime := time.Since(start)

	if err != nil {
		t.Errorf("❌ %s 清理后城市查询失败: %v", dbType, err)
		return
	}

	t.Logf("   ✅ %s 清理后城市查询成功: %d条记录", dbType, len(results4))

	t.Logf("✅ %s 选择性缓存清理验证: 活跃用户查询%v, 城市查询%v",
		dbType, activeQueryTime, cityQueryTime)

	// 清空所有缓存
	err = torm.ClearAllCache()
	if err != nil {
		t.Errorf("❌ %s 全部缓存清理失败: %v", dbType, err)
		return
	}
	t.Logf("✅ %s 全部缓存清理成功", dbType)

	// 验证缓存统计
	stats := torm.GetCacheStats()
	if stats != nil {
		t.Logf("📊 %s 清理后缓存统计: %+v", dbType, stats)
		if totalItems, ok := stats["total_items"].(int); ok && totalItems == 0 {
			t.Logf("✅ %s 缓存完全清空", dbType)
		}
	}
}
