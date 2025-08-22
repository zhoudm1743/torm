package tests

import (
	"testing"

	"github.com/zhoudm1743/torm/db"
	"github.com/zhoudm1743/torm/model"
)

// TestEnhancedWhereQueries 测试增强的WHERE查询方法
func TestEnhancedWhereQueries(t *testing.T) {
	// 配置SQLite数据库进行测试
	config := &db.Config{
		Driver:   "sqlite",
		Database: ":memory:",
	}

	err := db.AddConnection("test", config)
	if err != nil {
		t.Fatalf("添加连接失败: %v", err)
	}

	t.Run("WhereNull查询", func(t *testing.T) {
		query, err := db.Table("users", "test")
		if err != nil {
			t.Fatalf("创建查询失败: %v", err)
		}

		sql, bindings, err := query.WhereNull("email").ToSQL()
		if err != nil {
			t.Fatalf("构建SQL失败: %v", err)
		}

		expectedSQL := "SELECT * FROM users WHERE email IS NULL"
		if sql != expectedSQL {
			t.Errorf("期望SQL: %s, 实际SQL: %s", expectedSQL, sql)
		}

		if len(bindings) != 0 {
			t.Errorf("期望绑定参数数量: 0, 实际数量: %d", len(bindings))
		}
	})

	t.Run("WhereNotNull查询", func(t *testing.T) {
		query, err := db.Table("users", "test")
		if err != nil {
			t.Fatalf("创建查询失败: %v", err)
		}

		sql, bindings, err := query.WhereNotNull("email").ToSQL()
		if err != nil {
			t.Fatalf("构建SQL失败: %v", err)
		}

		expectedSQL := "SELECT * FROM users WHERE email IS NOT NULL"
		if sql != expectedSQL {
			t.Errorf("期望SQL: %s, 实际SQL: %s", expectedSQL, sql)
		}

		if len(bindings) != 0 {
			t.Errorf("期望绑定参数数量: 0, 实际数量: %d", len(bindings))
		}
	})

	t.Run("WhereBetween查询", func(t *testing.T) {
		query, err := db.Table("users", "test")
		if err != nil {
			t.Fatalf("创建查询失败: %v", err)
		}

		sql, bindings, err := query.WhereBetween("age", []interface{}{18, 65}).ToSQL()
		if err != nil {
			t.Fatalf("构建SQL失败: %v", err)
		}

		expectedSQL := "SELECT * FROM users WHERE age BETWEEN ? AND ?"
		if sql != expectedSQL {
			t.Errorf("期望SQL: %s, 实际SQL: %s", expectedSQL, sql)
		}

		if len(bindings) != 2 {
			t.Errorf("期望绑定参数数量: 2, 实际数量: %d", len(bindings))
		}

		if bindings[0] != 18 || bindings[1] != 65 {
			t.Errorf("期望绑定参数: [18, 65], 实际参数: %v", bindings)
		}
	})

	t.Run("WhereNotBetween查询", func(t *testing.T) {
		query, err := db.Table("users", "test")
		if err != nil {
			t.Fatalf("创建查询失败: %v", err)
		}

		sql, bindings, err := query.WhereNotBetween("age", []interface{}{18, 65}).ToSQL()
		if err != nil {
			t.Fatalf("构建SQL失败: %v", err)
		}

		expectedSQL := "SELECT * FROM users WHERE age NOT BETWEEN ? AND ?"
		if sql != expectedSQL {
			t.Errorf("期望SQL: %s, 实际SQL: %s", expectedSQL, sql)
		}

		if len(bindings) != 2 {
			t.Errorf("期望绑定参数数量: 2, 实际数量: %d", len(bindings))
		}
	})

	t.Run("WhereExists查询", func(t *testing.T) {
		query, err := db.Table("users", "test")
		if err != nil {
			t.Fatalf("创建查询失败: %v", err)
		}

		subQuery := "SELECT 1 FROM posts WHERE posts.user_id = users.id"
		sql, bindings, err := query.WhereExists(subQuery).ToSQL()
		if err != nil {
			t.Fatalf("构建SQL失败: %v", err)
		}

		expectedSQL := "SELECT * FROM users WHERE EXISTS (SELECT 1 FROM posts WHERE posts.user_id = users.id)"
		if sql != expectedSQL {
			t.Errorf("期望SQL: %s, 实际SQL: %s", expectedSQL, sql)
		}

		if len(bindings) != 0 {
			t.Errorf("期望绑定参数数量: 0, 实际数量: %d", len(bindings))
		}
	})

	t.Run("WhereNotExists查询", func(t *testing.T) {
		query, err := db.Table("users", "test")
		if err != nil {
			t.Fatalf("创建查询失败: %v", err)
		}

		subQuery := "SELECT 1 FROM posts WHERE posts.user_id = users.id"
		sql, bindings, err := query.WhereNotExists(subQuery).ToSQL()
		if err != nil {
			t.Fatalf("构建SQL失败: %v", err)
		}

		expectedSQL := "SELECT * FROM users WHERE NOT EXISTS (SELECT 1 FROM posts WHERE posts.user_id = users.id)"
		if sql != expectedSQL {
			t.Errorf("期望SQL: %s, 实际SQL: %s", expectedSQL, sql)
		}

		if len(bindings) != 0 {
			t.Errorf("期望绑定参数数量: 0, 实际数量: %d", len(bindings))
		}
	})
}

// TestEnhancedOrderQueries 测试增强的ORDER查询方法
func TestEnhancedOrderQueries(t *testing.T) {
	config := &db.Config{
		Driver:   "sqlite",
		Database: ":memory:",
	}

	err := db.AddConnection("test_order", config)
	if err != nil {
		t.Fatalf("添加连接失败: %v", err)
	}

	t.Run("OrderRand随机排序", func(t *testing.T) {
		query, err := db.Table("users", "test_order")
		if err != nil {
			t.Fatalf("创建查询失败: %v", err)
		}

		sql, bindings, err := query.OrderRand().ToSQL()
		if err != nil {
			t.Fatalf("构建SQL失败: %v", err)
		}

		expectedSQL := "SELECT * FROM users ORDER BY RANDOM()"
		if sql != expectedSQL {
			t.Errorf("期望SQL: %s, 实际SQL: %s", expectedSQL, sql)
		}

		if len(bindings) != 0 {
			t.Errorf("期望绑定参数数量: 0, 实际数量: %d", len(bindings))
		}
	})

	t.Run("OrderField按字段值排序", func(t *testing.T) {
		query, err := db.Table("users", "test_order")
		if err != nil {
			t.Fatalf("创建查询失败: %v", err)
		}

		values := []interface{}{"active", "inactive", "pending"}
		sql, bindings, err := query.OrderField("status", values, "asc").ToSQL()
		if err != nil {
			t.Fatalf("构建SQL失败: %v", err)
		}

		expectedSQL := "SELECT * FROM users ORDER BY CASE WHEN status = ? THEN 0 WHEN status = ? THEN 1 WHEN status = ? THEN 2 ELSE 999 END ASC"
		if sql != expectedSQL {
			t.Errorf("期望SQL: %s, 实际SQL: %s", expectedSQL, sql)
		}

		if len(bindings) != 3 {
			t.Errorf("期望绑定参数数量: 3, 实际数量: %d", len(bindings))
		}
	})

	t.Run("FieldRaw原生字段", func(t *testing.T) {
		query, err := db.Table("users", "test_order")
		if err != nil {
			t.Fatalf("创建查询失败: %v", err)
		}

		sql, bindings, err := query.FieldRaw("COUNT(*) as total").ToSQL()
		if err != nil {
			t.Fatalf("构建SQL失败: %v", err)
		}

		// FieldRaw应该添加到字段列表中
		expectedSQL := "SELECT *, COUNT(*) as total FROM users"
		if sql != expectedSQL {
			t.Errorf("期望SQL: %s, 实际SQL: %s", expectedSQL, sql)
		}

		if len(bindings) != 0 {
			t.Errorf("期望绑定参数数量: 0, 实际数量: %d", len(bindings))
		}
	})
}

// TestModelEnhancedMethods 测试模型的增强方法
func TestModelEnhancedMethods(t *testing.T) {
	config := &db.Config{
		Driver:   "sqlite",
		Database: ":memory:",
	}

	err := db.AddConnection("test_model", config)
	if err != nil {
		t.Fatalf("添加连接失败: %v", err)
	}

	// 创建测试模型
	type TestUser struct {
		model.BaseModel
		ID     int    `json:"id" db:"id"`
		Name   string `json:"name" db:"name"`
		Email  string `json:"email" db:"email"`
		Status string `json:"status" db:"status"`
		Age    int    `json:"age" db:"age"`
	}

	newTestUser := func() *TestUser {
		user := &TestUser{BaseModel: *model.NewBaseModel()}
		user.SetTable("test_users")
		user.SetConnection("test_model")
		return user
	}

	t.Run("模型WhereNull方法", func(t *testing.T) {
		user := newTestUser()
		user.WhereNull("email")

		query := user.GetQueryBuilder()
		if query == nil {
			t.Fatal("获取查询构建器失败")
		}

		sql, bindings, err := query.ToSQL()
		if err != nil {
			t.Fatalf("构建SQL失败: %v", err)
		}

		expectedSQL := "SELECT * FROM test_users WHERE email IS NULL"
		if sql != expectedSQL {
			t.Errorf("期望SQL: %s, 实际SQL: %s", expectedSQL, sql)
		}

		if len(bindings) != 0 {
			t.Errorf("期望绑定参数数量: 0, 实际数量: %d", len(bindings))
		}
	})

	t.Run("模型WhereBetween方法", func(t *testing.T) {
		user := newTestUser()
		user.WhereBetween("age", []interface{}{18, 65})

		query := user.GetQueryBuilder()
		if query == nil {
			t.Fatal("获取查询构建器失败")
		}

		sql, bindings, err := query.ToSQL()
		if err != nil {
			t.Fatalf("构建SQL失败: %v", err)
		}

		expectedSQL := "SELECT * FROM test_users WHERE age BETWEEN ? AND ?"
		if sql != expectedSQL {
			t.Errorf("期望SQL: %s, 实际SQL: %s", expectedSQL, sql)
		}

		if len(bindings) != 2 {
			t.Errorf("期望绑定参数数量: 2, 实际数量: %d", len(bindings))
		}
	})

	t.Run("模型OrderRand方法", func(t *testing.T) {
		user := newTestUser()
		user.OrderRand()

		query := user.GetQueryBuilder()
		if query == nil {
			t.Fatal("获取查询构建器失败")
		}

		sql, bindings, err := query.ToSQL()
		if err != nil {
			t.Fatalf("构建SQL失败: %v", err)
		}

		expectedSQL := "SELECT * FROM test_users ORDER BY RANDOM()"
		if sql != expectedSQL {
			t.Errorf("期望SQL: %s, 实际SQL: %s", expectedSQL, sql)
		}

		if len(bindings) != 0 {
			t.Errorf("期望绑定参数数量: 0, 实际数量: %d", len(bindings))
		}
	})

	t.Run("模型链式调用", func(t *testing.T) {
		user := newTestUser()
		user.WhereNotNull("email").
			WhereBetween("age", []interface{}{18, 65}).
			OrderRand()

		query := user.GetQueryBuilder()
		if query == nil {
			t.Fatal("获取查询构建器失败")
		}

		sql, bindings, err := query.ToSQL()
		if err != nil {
			t.Fatalf("构建SQL失败: %v", err)
		}

		expectedSQL := "SELECT * FROM test_users WHERE email IS NOT NULL AND age BETWEEN ? AND ? ORDER BY RANDOM()"
		if sql != expectedSQL {
			t.Errorf("期望SQL: %s, 实际SQL: %s", expectedSQL, sql)
		}

		if len(bindings) != 2 {
			t.Errorf("期望绑定参数数量: 2, 实际数量: %d", len(bindings))
		}
	})
}
