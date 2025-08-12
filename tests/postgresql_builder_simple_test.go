package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zhoudm1743/torm/db"
)

func TestPostgreSQLBuilderBasic(t *testing.T) {
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

		// 验证包含所有字段
		assert.Contains(t, sql, `"name"`)
		assert.Contains(t, sql, `"email"`)
		assert.Contains(t, sql, `"age"`)
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

	t.Run("标识符和值引用", func(t *testing.T) {
		// 标识符引用测试
		assert.Equal(t, `"users"`, builder.QuoteIdentifier("users"))
		assert.Equal(t, `"user_name"`, builder.QuoteIdentifier("user_name"))

		// 已经包含引号的不再添加
		assert.Equal(t, `"users"."name"`, builder.QuoteIdentifier(`"users"."name"`))

		// 函数调用不加引号
		assert.Equal(t, "COUNT(*)", builder.QuoteIdentifier("COUNT(*)"))
		assert.Equal(t, "MAX(age)", builder.QuoteIdentifier("MAX(age)"))

		// 值引用测试
		assert.Equal(t, "'test'", builder.QuoteValue("test"))
		assert.Equal(t, "NULL", builder.QuoteValue(nil))
		assert.Equal(t, "123", builder.QuoteValue(123))
		assert.Equal(t, "'don''t'", builder.QuoteValue("don't")) // SQL转义
	})

	t.Run("空数据处理", func(t *testing.T) {
		// 空INSERT数据
		_, _, err := builder.BuildInsert("users", map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no data to insert")
	})

	t.Run("SQL注入防护", func(t *testing.T) {
		data := map[string]interface{}{
			"name": "'; DROP TABLE users; --",
		}

		sql, bindings, err := builder.BuildInsert("users", data)
		assert.NoError(t, err)
		assert.Contains(t, sql, "$1")
		assert.Equal(t, "'; DROP TABLE users; --", bindings[0])
		// 确保恶意代码在绑定参数中，而不是直接在SQL中
		assert.NotContains(t, sql, "DROP TABLE")
	})
}

func TestPostgreSQLBuilderPlaceholders(t *testing.T) {
	builder := db.NewPostgreSQLBuilder()

	t.Run("占位符编号正确", func(t *testing.T) {
		// 测试多个字段的INSERT
		data := map[string]interface{}{
			"field1": "value1",
			"field2": "value2",
			"field3": "value3",
			"field4": "value4",
			"field5": "value5",
		}

		sql, bindings, err := builder.BuildInsert("test_table", data)
		assert.NoError(t, err)
		assert.Len(t, bindings, 5)

		// 验证占位符从$1到$5都存在
		assert.Contains(t, sql, "$1")
		assert.Contains(t, sql, "$2")
		assert.Contains(t, sql, "$3")
		assert.Contains(t, sql, "$4")
		assert.Contains(t, sql, "$5")
	})

	t.Run("批量插入占位符", func(t *testing.T) {
		data := []map[string]interface{}{
			{"a": 1, "b": 2},
			{"a": 3, "b": 4},
			{"a": 5, "b": 6},
		}

		sql, bindings, err := builder.BuildInsertBatch("test_table", data)
		assert.NoError(t, err)
		assert.Len(t, bindings, 6) // 3行 × 2列 = 6个值

		// 验证占位符从$1到$6都存在
		for i := 1; i <= 6; i++ {
			assert.Contains(t, sql, fmt.Sprintf("$%d", i))
		}
	})
}
