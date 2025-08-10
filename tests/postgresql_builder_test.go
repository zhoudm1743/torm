package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"torm/pkg/db"
)

func TestPostgreSQLBuilderQueries(t *testing.T) {
	builder := db.NewPostgreSQLBuilder()

	t.Run("INSERT语句构建", func(t *testing.T) {
		data := map[string]interface{}{
			"name":  "测试用户",
			"email": "test@example.com",
			"age":   25,
		}

		sql, bindings, err := builder.BuildInsert("users", data)
		assert.NoError(t, err)
		assert.Contains(t, sql, `INSERT INTO "users"`)
		assert.Contains(t, sql, "RETURNING id")
		assert.Contains(t, sql, "$1")
		assert.Contains(t, sql, "$2")
		assert.Contains(t, sql, "$3")
		assert.Len(t, bindings, 3)
	})

	t.Run("批量INSERT语句构建", func(t *testing.T) {
		data := []map[string]interface{}{
			{"name": "用户1", "email": "user1@example.com", "age": 20},
			{"name": "用户2", "email": "user2@example.com", "age": 25},
		}

		sql, bindings, err := builder.BuildInsertBatch("users", data)
		assert.NoError(t, err)
		assert.Contains(t, sql, `INSERT INTO "users"`)
		assert.Contains(t, sql, "VALUES")
		assert.Contains(t, sql, "$1")
		assert.Contains(t, sql, "$6") // 应该有6个占位符（2行×3列）
		assert.Len(t, bindings, 6)
	})

	t.Run("UPDATE语句构建", func(t *testing.T) {
		// 通过Table方法创建正确的查询构建器
		qb, err := db.Table("users", "")
		require.NoError(t, err)
		qb.Where("id", "=", 1)

		updateData := map[string]interface{}{
			"name": "更新的用户",
			"age":  30,
		}

		sql, bindings, err := builder.BuildUpdate(qb, updateData)
		assert.NoError(t, err)
		assert.Contains(t, sql, `UPDATE "users" SET`)
		assert.Contains(t, sql, `"name" = $1`)
		assert.Contains(t, sql, `"age" = $2`)
		assert.Contains(t, sql, "WHERE")
		assert.Len(t, bindings, 3) // 2个更新值 + 1个WHERE条件
	})

	t.Run("DELETE语句构建", func(t *testing.T) {
		qb, err := db.Table("users", "")
		require.NoError(t, err)
		qb.Where("status", "=", "inactive")

		sql, bindings, err := builder.BuildDelete(qb)
		assert.NoError(t, err)
		assert.Contains(t, sql, `DELETE FROM "users"`)
		assert.Contains(t, sql, "WHERE")
		assert.Contains(t, sql, `"status" = $1`)
		assert.Len(t, bindings, 1)
	})

	t.Run("SELECT语句构建", func(t *testing.T) {
		qb, err := db.Table("users", "")
		require.NoError(t, err)
		qb.Select("id", "name", "email")
		qb.Where("age", ">", 18)
		qb.OrderBy("created_at", "DESC")
		qb.Limit(10)
		qb.Offset(5)

		sql, bindings, err := builder.BuildSelect(qb)
		assert.NoError(t, err)
		assert.Contains(t, sql, "SELECT id, name, email")
		assert.Contains(t, sql, `FROM "users"`)
		assert.Contains(t, sql, "WHERE")
		assert.Contains(t, sql, `"age" > $1`)
		assert.Contains(t, sql, `ORDER BY "created_at" DESC`)
		assert.Contains(t, sql, "LIMIT 10")
		assert.Contains(t, sql, "OFFSET 5")
		assert.Len(t, bindings, 1)
	})

	t.Run("复杂SELECT语句构建", func(t *testing.T) {
		qb, err := db.Table("users", "")
		require.NoError(t, err)
		qb.Select("users.name", "profiles.bio")
		qb.LeftJoin("profiles", "users.id", "=", "profiles.user_id")
		qb.Where("users.status", "=", "active")
		qb.Where("users.age", ">=", 21)
		qb.GroupBy("users.id")
		qb.Having("COUNT(posts.id)", ">", 5)
		qb.OrderBy("users.created_at", "DESC")

		sql, bindings, err := builder.BuildSelect(qb)
		assert.NoError(t, err)
		assert.Contains(t, sql, "SELECT users.name, profiles.bio")
		assert.Contains(t, sql, `FROM "users"`)
		assert.Contains(t, sql, `LEFT JOIN "profiles"`)
		assert.Contains(t, sql, "WHERE")
		assert.Contains(t, sql, "GROUP BY")
		assert.Contains(t, sql, "HAVING")
		assert.Contains(t, sql, "ORDER BY")
		assert.Len(t, bindings, 3) // 2个WHERE条件 + 1个HAVING条件
	})

	t.Run("COUNT语句构建", func(t *testing.T) {
		qb, err := db.Table("users", "")
		require.NoError(t, err)
		qb.Where("status", "=", "active")

		sql, bindings, err := builder.BuildCount(qb)
		assert.NoError(t, err)
		assert.Contains(t, sql, "SELECT COUNT(*)")
		assert.Contains(t, sql, `FROM "users"`)
		assert.Contains(t, sql, "WHERE")
		assert.Contains(t, sql, `"status" = $1`)
		assert.Len(t, bindings, 1)
	})
}

func TestPostgreSQLBuilderSpecialCases(t *testing.T) {
	builder := db.NewPostgreSQLBuilder()

	t.Run("空数据处理", func(t *testing.T) {
		// 空INSERT数据
		_, _, err := builder.BuildInsert("users", map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no data to insert")

		// 空UPDATE数据
		qb, err := db.Table("users", "")
		require.NoError(t, err)
		_, _, err = builder.BuildUpdate(qb, map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no data to update")
	})

	t.Run("特殊字符处理", func(t *testing.T) {
		// 包含单引号的字符串
		value := builder.QuoteValue("don't worry")
		assert.Equal(t, "'don''t worry'", value)

		// SQL注入防护测试
		data := map[string]interface{}{
			"name": "'; DROP TABLE users; --",
		}

		sql, bindings, err := builder.BuildInsert("users", data)
		assert.NoError(t, err)
		assert.Contains(t, sql, "$1")
		assert.Equal(t, "'; DROP TABLE users; --", bindings[0])
	})

	t.Run("标识符引用测试", func(t *testing.T) {
		// 普通标识符
		assert.Equal(t, `"users"`, builder.QuoteIdentifier("users"))

		// 已经有引号的标识符
		assert.Equal(t, `"users"."name"`, builder.QuoteIdentifier(`"users"."name"`))

		// 函数调用不加引号
		assert.Equal(t, "COUNT(*)", builder.QuoteIdentifier("COUNT(*)"))
		assert.Equal(t, "MAX(age)", builder.QuoteIdentifier("MAX(age)"))

		// 表达式不加引号
		assert.Equal(t, "users.id + 1", builder.QuoteIdentifier("users.id + 1"))
	})

	t.Run("复杂WHERE条件", func(t *testing.T) {
		qb, err := db.Table("users", "")
		require.NoError(t, err)
		qb.Where("age", "IN", []interface{}{18, 25, 30})
		qb.Where("created_at", "BETWEEN", []interface{}{"2023-01-01", "2023-12-31"})
		qb.Where("deleted_at", "IS NULL", nil)
		qb.OrWhere("status", "=", "vip")

		sql, bindings, err := builder.BuildSelect(qb)
		assert.NoError(t, err)
		assert.Contains(t, sql, "IN ($1, $2, $3)")
		assert.Contains(t, sql, "BETWEEN $4 AND $5")
		assert.Contains(t, sql, "IS NULL")
		assert.Contains(t, sql, "OR")
		assert.Len(t, bindings, 6) // IN的3个值 + BETWEEN的2个值 + OR条件的1个值
	})
}
