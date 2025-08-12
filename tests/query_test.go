package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zhoudm1743/torm/db"
)

// TestTORMStyleWhereDemo 测试TORM风格的Where查询
func TestTORMStyleWhereDemo(t *testing.T) {
	query := &db.QueryBuilder{}

	t.Run("基础参数化查询", func(t *testing.T) {
		sql, bindings, err := query.
			From("users").
			Where("name = ?", "张三").
			ToSQL()

		require.NoError(t, err)
		assert.Equal(t, "SELECT * FROM users WHERE name = ?", sql)
		assert.Equal(t, []interface{}{"张三"}, bindings)
	})

	t.Run("多参数查询", func(t *testing.T) {
		sql, bindings, err := query.
			From("users").
			Where("name = ? AND age >= ?", "张三", 22).
			ToSQL()

		require.NoError(t, err)
		assert.Equal(t, "SELECT * FROM users WHERE name = ? AND age >= ?", sql)
		assert.Equal(t, []interface{}{"张三", 22}, bindings)
	})

	t.Run("不等于查询", func(t *testing.T) {
		sql, bindings, err := query.
			From("users").
			Where("name <> ?", "张三").
			ToSQL()

		require.NoError(t, err)
		assert.Equal(t, "SELECT * FROM users WHERE name <> ?", sql)
		assert.Equal(t, []interface{}{"张三"}, bindings)
	})

	t.Run("LIKE查询", func(t *testing.T) {
		sql, bindings, err := query.
			From("users").
			Where("name LIKE ?", "%jin%").
			ToSQL()

		require.NoError(t, err)
		assert.Equal(t, "SELECT * FROM users WHERE name LIKE ?", sql)
		assert.Equal(t, []interface{}{"%jin%"}, bindings)
	})

	t.Run("BETWEEN查询", func(t *testing.T) {
		sql, bindings, err := query.
			From("users").
			Where("created_at BETWEEN ? AND ?", "2000-01-01", "2000-01-08").
			ToSQL()

		require.NoError(t, err)
		assert.Equal(t, "SELECT * FROM users WHERE created_at BETWEEN ? AND ?", sql)
		assert.Equal(t, []interface{}{"2000-01-01", "2000-01-08"}, bindings)
	})

	t.Run("IN查询", func(t *testing.T) {
		sql, bindings, err := query.
			From("users").
			Where("name IN (?, ?)", "张三", "李四").
			ToSQL()

		require.NoError(t, err)
		assert.Equal(t, "SELECT * FROM users WHERE name IN (?, ?)", sql)
		assert.Equal(t, []interface{}{"张三", "李四"}, bindings)
	})

	t.Run("兼容传统三参数方式", func(t *testing.T) {
		sql, bindings, err := query.
			From("users").
			Where("age", ">", 18).
			ToSQL()

		require.NoError(t, err)
		assert.Equal(t, "SELECT * FROM users WHERE age > ?", sql)
		assert.Equal(t, []interface{}{18}, bindings)
	})

	t.Run("混合使用新旧方式", func(t *testing.T) {
		sql, bindings, err := query.
			From("users").
			Where("name = ?", "张三").
			Where("age", ">", 18).
			Where("status IN (?, ?)", "active", "pending").
			ToSQL()

		require.NoError(t, err)
		expectedSQL := "SELECT * FROM users WHERE name = ? AND age > ? AND status IN (?, ?)"
		assert.Equal(t, expectedSQL, sql)
		assert.Equal(t, []interface{}{"张三", 18, "active", "pending"}, bindings)
	})
}

// TestTORMStyleComplexQueriesDemo 测试复杂的TORM风格查询
func TestTORMStyleComplexQueriesDemo(t *testing.T) {
	query := &db.QueryBuilder{}

	t.Run("复杂WHERE条件", func(t *testing.T) {
		sql, bindings, err := query.
			From("users").
			Select("id", "name", "email").
			Where("(name = ? OR email = ?) AND age >= ? AND status = ?", "john", "john@example.com", 18, "active").
			OrderBy("created_at", "desc").
			Limit(10).
			ToSQL()

		require.NoError(t, err)
		expectedSQL := "SELECT id, name, email FROM users WHERE (name = ? OR email = ?) AND age >= ? AND status = ? ORDER BY created_at DESC LIMIT ?"
		assert.Equal(t, expectedSQL, sql)
		assert.Equal(t, []interface{}{"john", "john@example.com", 18, "active", 10}, bindings)
	})

	t.Run("多个WHERE条件组合", func(t *testing.T) {
		sql, bindings, err := query.
			From("orders").
			Where("user_id = ?", 1).
			Where("total >= ? AND total <= ?", 100, 1000).
			Where("status IN (?, ?, ?)", "pending", "processing", "completed").
			ToSQL()

		require.NoError(t, err)
		expectedSQL := "SELECT * FROM orders WHERE user_id = ? AND total >= ? AND total <= ? AND status IN (?, ?, ?)"
		assert.Equal(t, expectedSQL, sql)
		assert.Equal(t, []interface{}{1, 100, 1000, "pending", "processing", "completed"}, bindings)
	})
}

// TestTORMStyleOrWhereDemo 测试TORM风格的OrWhere查询
func TestTORMStyleOrWhereDemo(t *testing.T) {
	query := &db.QueryBuilder{}

	t.Run("基础OR查询", func(t *testing.T) {
		sql, bindings, err := query.
			From("users").
			Where("name = ?", "张三").
			OrWhere("email = ?", "user@example.com").
			ToSQL()

		require.NoError(t, err)
		assert.Equal(t, "SELECT * FROM users WHERE name = ? OR email = ?", sql)
		assert.Equal(t, []interface{}{"张三", "user@example.com"}, bindings)
	})

	t.Run("兼容传统OR查询", func(t *testing.T) {
		sql, bindings, err := query.
			From("users").
			Where("age", ">", 18).
			OrWhere("is_vip", "=", true).
			ToSQL()

		require.NoError(t, err)
		assert.Equal(t, "SELECT * FROM users WHERE age > ? OR is_vip = ?", sql)
		assert.Equal(t, []interface{}{18, true}, bindings)
	})
}

// TestWhereMethodFlexibilityDemo 测试Where方法的灵活性
func TestWhereMethodFlexibilityDemo(t *testing.T) {
	query := &db.QueryBuilder{}

	t.Run("空参数", func(t *testing.T) {
		newQuery := query.From("users").Where()
		sql, bindings, err := newQuery.ToSQL()

		require.NoError(t, err)
		assert.Equal(t, "SELECT * FROM users", sql)
		assert.Empty(t, bindings)
	})

	t.Run("单参数字符串", func(t *testing.T) {
		sql, bindings, err := query.
			From("users").
			Where("active = 1").
			ToSQL()

		require.NoError(t, err)
		assert.Equal(t, "SELECT * FROM users WHERE active = 1", sql)
		assert.Empty(t, bindings)
	})

	t.Run("时间类型参数", func(t *testing.T) {
		sql, bindings, err := query.
			From("users").
			Where("updated_at > ?", "2000-01-01 00:00:00").
			ToSQL()

		require.NoError(t, err)
		assert.Equal(t, "SELECT * FROM users WHERE updated_at > ?", sql)
		assert.Equal(t, []interface{}{"2000-01-01 00:00:00"}, bindings)
	})
}
