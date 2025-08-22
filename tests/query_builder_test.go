package tests

import (
	"testing"

	"github.com/zhoudm1743/torm/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryBuilder_BasicSelect(t *testing.T) {
	query := &db.QueryBuilder{}

	sql, bindings, err := query.
		From("users").
		Select("id", "name", "email").
		Where("age", ">", 18).
		ToSQL()

	require.NoError(t, err)
	assert.Equal(t, "SELECT id, name, email FROM users WHERE age > ?", sql)
	assert.Equal(t, []interface{}{18}, bindings)
}

func TestQueryBuilder_WhereConditions(t *testing.T) {
	query := &db.QueryBuilder{}

	sql, bindings, err := query.
		From("products").
		Select("*").
		Where("price", ">", 100).
		Where("price", "<", 1000).
		WhereIn("category_id", []interface{}{1, 2, 3}).
		WhereNotIn("status", []interface{}{"deleted", "archived"}).
		WhereBetween("created_at", []interface{}{"2024-01-01", "2024-12-31"}).
		WhereNull("deleted_at").
		WhereNotNull("updated_at").
		ToSQL()

	require.NoError(t, err)
	expectedSQL := "SELECT * FROM products WHERE price > ? AND price < ? AND category_id IN (?, ?, ?) AND status NOT IN (?, ?) AND created_at BETWEEN ? AND ? AND deleted_at IS NULL AND updated_at IS NOT NULL"
	assert.Equal(t, expectedSQL, sql)
	expectedBindings := []interface{}{100, 1000, 1, 2, 3, "deleted", "archived", "2024-01-01", "2024-12-31"}
	assert.Equal(t, expectedBindings, bindings)
}

func TestQueryBuilder_OrWhere(t *testing.T) {
	query := &db.QueryBuilder{}

	sql, bindings, err := query.
		From("users").
		Select("*").
		Where("age", ">", 18).
		OrWhere("is_vip", "=", true).
		ToSQL()

	require.NoError(t, err)
	assert.Equal(t, "SELECT * FROM users WHERE age > ? OR is_vip = ?", sql)
	assert.Equal(t, []interface{}{18, true}, bindings)
}

func TestQueryBuilder_WhereRaw(t *testing.T) {
	query := &db.QueryBuilder{}

	sql, bindings, err := query.
		From("users").
		Select("*").
		Where("age", ">", 18).
		WhereRaw("YEAR(created_at) = ?", 2024).
		ToSQL()

	require.NoError(t, err)
	assert.Equal(t, "SELECT * FROM users WHERE age > ? AND YEAR(created_at) = ?", sql)
	assert.Equal(t, []interface{}{18, 2024}, bindings)
}

func TestQueryBuilder_Joins(t *testing.T) {
	query := &db.QueryBuilder{}

	sql, bindings, err := query.
		From("orders o").
		Select("o.id", "u.name", "p.title").
		InnerJoin("users u", "o.user_id", "=", "u.id").
		LeftJoin("profiles p", "u.id", "=", "p.user_id").
		RightJoin("payments pay", "o.id", "=", "pay.order_id").
		Where("o.status", "=", "completed").
		ToSQL()

	require.NoError(t, err)
	expectedSQL := "SELECT o.id, u.name, p.title FROM orders o INNER JOIN users u ON o.user_id = u.id LEFT JOIN profiles p ON u.id = p.user_id RIGHT JOIN payments pay ON o.id = pay.order_id WHERE o.status = ?"
	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, []interface{}{"completed"}, bindings)
}

func TestQueryBuilder_GroupByHaving(t *testing.T) {
	query := &db.QueryBuilder{}

	sql, bindings, err := query.
		From("sales").
		SelectRaw("DATE(created_at) as date, SUM(amount) as total").
		Where("status", "=", "completed").
		GroupBy("DATE(created_at)").
		Having("SUM(amount)", ">", 1000).
		ToSQL()

	require.NoError(t, err)
	expectedSQL := "SELECT DATE(created_at) as date, SUM(amount) as total FROM sales WHERE status = ? GROUP BY DATE(created_at) HAVING SUM(amount) > ?"
	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, []interface{}{"completed", 1000}, bindings)
}

func TestQueryBuilder_OrderBy(t *testing.T) {
	query := &db.QueryBuilder{}

	sql, bindings, err := query.
		From("users").
		Select("*").
		OrderBy("created_at", "desc").
		OrderBy("name", "asc").
		ToSQL()

	require.NoError(t, err)
	assert.Equal(t, "SELECT * FROM users ORDER BY created_at DESC, name ASC", sql)
	assert.Empty(t, bindings)
}

func TestQueryBuilder_OrderByRaw(t *testing.T) {
	query := &db.QueryBuilder{}

	sql, bindings, err := query.
		From("users").
		Select("*").
		OrderByRaw("FIELD(status, ?, ?, ?)", "active", "pending", "inactive").
		ToSQL()

	require.NoError(t, err)
	assert.Equal(t, "SELECT * FROM users ORDER BY FIELD(status, ?, ?, ?)", sql)
	assert.Equal(t, []interface{}{"active", "pending", "inactive"}, bindings)
}

func TestQueryBuilder_LimitOffset(t *testing.T) {
	query := &db.QueryBuilder{}

	sql, bindings, err := query.
		From("users").
		Select("*").
		Limit(10).
		Offset(20).
		ToSQL()

	require.NoError(t, err)
	assert.Equal(t, "SELECT * FROM users LIMIT ? OFFSET ?", sql)
	assert.Equal(t, []interface{}{10, 20}, bindings)
}

func TestQueryBuilder_Page(t *testing.T) {
	query := &db.QueryBuilder{}

	sql, bindings, err := query.
		From("users").
		Select("*").
		Page(3, 15). // 第3页，每页15条记录
		ToSQL()

	require.NoError(t, err)
	assert.Equal(t, "SELECT * FROM users LIMIT ? OFFSET ?", sql)
	assert.Equal(t, []interface{}{15, 30}, bindings) // offset = (3-1) * 15 = 30
}

func TestQueryBuilder_Distinct(t *testing.T) {
	query := &db.QueryBuilder{}

	sql, bindings, err := query.
		From("user_activities").
		Select("user_id", "activity_type").
		Distinct().
		Where("created_at", ">=", "2024-01-01").
		ToSQL()

	require.NoError(t, err)
	assert.Equal(t, "SELECT DISTINCT user_id, activity_type FROM user_activities WHERE created_at >= ?", sql)
	assert.Equal(t, []interface{}{"2024-01-01"}, bindings)
}

func TestQueryBuilder_ComplexQuery(t *testing.T) {
	query := &db.QueryBuilder{}

	sql, bindings, err := query.
		From("users u").
		Select("u.id", "u.name", "p.title").
		LeftJoin("profiles p", "u.id", "=", "p.user_id").
		Where("u.age", ">=", 18).
		Where("u.age", "<=", 65).
		WhereNotNull("u.email").
		OrWhere("u.is_vip", "=", true).
		GroupBy("u.id").
		Having("COUNT(p.id)", ">", 0).
		OrderBy("u.created_at", "desc").
		Limit(20).
		Offset(40).
		ToSQL()

	require.NoError(t, err)
	expectedSQL := "SELECT u.id, u.name, p.title FROM users u LEFT JOIN profiles p ON u.id = p.user_id WHERE u.age >= ? AND u.age <= ? AND u.email IS NOT NULL OR u.is_vip = ? GROUP BY u.id HAVING COUNT(p.id) > ? ORDER BY u.created_at DESC LIMIT ? OFFSET ?"
	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, []interface{}{18, 65, true, 0, 20, 40}, bindings)
}

func TestQueryBuilder_Clone(t *testing.T) {
	original := &db.QueryBuilder{}
	original = original.From("users").(*db.QueryBuilder)
	original = original.Where("age", ">", 18).(*db.QueryBuilder)

	cloned := original.Clone().(*db.QueryBuilder)
	cloned = cloned.Where("status", "=", "active").(*db.QueryBuilder)

	// 原始查询不应该受到克隆查询的影响
	originalSQL, originalBindings, err := original.ToSQL()
	require.NoError(t, err)
	assert.Equal(t, "SELECT * FROM users WHERE age > ?", originalSQL)
	assert.Equal(t, []interface{}{18}, originalBindings)

	// 克隆查询应该包含新的条件
	clonedSQL, clonedBindings, err := cloned.ToSQL()
	require.NoError(t, err)
	assert.Equal(t, "SELECT * FROM users WHERE age > ? AND status = ?", clonedSQL)
	assert.Equal(t, []interface{}{18, "active"}, clonedBindings)
}

func TestQueryBuilder_EmptyTable(t *testing.T) {
	query := &db.QueryBuilder{}

	sql, bindings, err := query.
		Select("*").
		Where("id", "=", 1).
		ToSQL()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "table name is required")
	assert.Empty(t, sql)
	assert.Empty(t, bindings)
}

func TestQueryBuilder_MultipleGroupBy(t *testing.T) {
	query := &db.QueryBuilder{}

	sql, bindings, err := query.
		From("sales").
		Select("year", "month", "SUM(amount)").
		GroupBy("year", "month").
		ToSQL()

	require.NoError(t, err)
	assert.Equal(t, "SELECT year, month, SUM(amount) FROM sales GROUP BY year, month", sql)
	assert.Empty(t, bindings)
}
