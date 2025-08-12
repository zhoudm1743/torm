package tests

import (
	"testing"

	"github.com/zhoudm1743/torm/db"
	"github.com/zhoudm1743/torm/model"
)

// PriorityTestModel 优先级测试模型
type PriorityTestModel struct {
	model.BaseModel
	ID   uint   `json:"id" db:"id" pk:""`
	Name string `json:"name" db:"name"`
}

func TestTableNamePriority(t *testing.T) {
	t.Run("手动设置表名优先", func(t *testing.T) {
		// 创建模型实例
		testModel := &PriorityTestModel{
			BaseModel: *model.NewBaseModel(),
		}

		// 手动设置表名
		testModel.SetTable("custom_table_name")

		// 验证TableName()方法返回设置的表名
		tableName := testModel.TableName()
		if tableName != "custom_table_name" {
			t.Errorf("期望表名: 'custom_table_name'，实际: '%s'", tableName)
		}

		// 测试db.Model()是否使用手动设置的表名
		_, err := db.Model(testModel)
		if err != nil {
			// 预期的连接错误，检查错误信息是否包含正确的表名信息
			t.Logf("db.Model()结果（预期连接错误）: %v", err)
			// 只要不是"cannot determine table name"错误就说明表名获取成功
			if err.Error() == "cannot determine table name from model" {
				t.Error("应该能够获取到手动设置的表名")
			}
		}

		t.Logf("手动设置的表名: %s", tableName)
	})

	t.Run("没有设置时使用推断", func(t *testing.T) {
		// 创建模型实例，但不设置表名
		testModel := &PriorityTestModel{
			BaseModel: *model.NewBaseModel(),
		}

		// BaseModel的TableName()方法应该返回空（因为没有设置）
		tableName := testModel.TableName()
		if tableName != "" {
			t.Errorf("没有设置表名时，TableName()应该返回空字符串，实际: '%s'", tableName)
		}

		// 但是db.Model()应该能够推断出表名
		_, err := db.Model(testModel)
		if err != nil {
			// 预期的连接错误，检查错误信息
			t.Logf("db.Model()结果（预期连接错误）: %v", err)
			if err.Error() == "cannot determine table name from model" {
				t.Error("db.Model()应该能够推断出表名")
			}
		}

		t.Logf("BaseModel.TableName(): '%s'（应该为空）", tableName)
		t.Log("db.Model()能够成功推断表名（从错误信息可以看出不是表名推断失败）")
	})

	t.Run("设置空表名时使用推断", func(t *testing.T) {
		// 创建模型实例
		testModel := &PriorityTestModel{
			BaseModel: *model.NewBaseModel(),
		}

		// 设置空表名
		testModel.SetTable("")

		// BaseModel的TableName()应该返回设置的空字符串
		tableName := testModel.TableName()
		if tableName != "" {
			t.Errorf("设置空表名后，TableName()应该返回空字符串，实际: '%s'", tableName)
		}

		// 但是db.Model()应该能够推断出表名
		_, err := db.Model(testModel)
		if err != nil {
			t.Logf("db.Model()结果（预期连接错误）: %v", err)
			if err.Error() == "cannot determine table name from model" {
				t.Error("设置空表名时，db.Model()应该回退到推断")
			}
		}

		t.Logf("设置空表名后，BaseModel.TableName(): '%s'", tableName)
		t.Log("db.Model()能够推断出表名")
	})

	t.Run("重新设置表名", func(t *testing.T) {
		// 创建模型实例
		testModel := &PriorityTestModel{
			BaseModel: *model.NewBaseModel(),
		}

		// 先设置一个表名
		testModel.SetTable("first_table")
		firstTable := testModel.TableName()

		// 再设置另一个表名
		testModel.SetTable("second_table")
		secondTable := testModel.TableName()

		if firstTable != "first_table" {
			t.Errorf("第一次设置失败，期望: 'first_table'，实际: '%s'", firstTable)
		}

		if secondTable != "second_table" {
			t.Errorf("第二次设置失败，期望: 'second_table'，实际: '%s'", secondTable)
		}

		t.Logf("第一次设置: %s, 第二次设置: %s", firstTable, secondTable)
	})
}
