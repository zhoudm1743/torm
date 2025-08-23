package tests

import (
	"os"
	"testing"

	"github.com/zhoudm1743/torm/db"
	"github.com/zhoudm1743/torm/model"
)

// TypeTestModel 类型测试模型
type TypeTestModel struct {
	model.BaseModel
	ID          int64   `json:"id" db:"id" torm:"primary_key,auto_increment"`
	ShortCode   string  `json:"short_code" db:"short_code" torm:"type:varchar,size:10,comment:短编码"`
	LongName    string  `json:"long_name" db:"long_name" torm:"type:varchar,size:200,comment:长名称"`
	FixedCode   string  `json:"fixed_code" db:"fixed_code" torm:"type:char,size:5,comment:固定编码"`
	Price       float64 `json:"price" db:"price" torm:"type:decimal,precision:10,scale:2,comment:价格"`
	Percentage  float64 `json:"percentage" db:"percentage" torm:"type:decimal,precision:5,scale:4,comment:百分比"`
	Description string  `json:"description" db:"description" torm:"type:text,comment:描述"`
}

func NewTypeTestModel() *TypeTestModel {
	tm := &TypeTestModel{BaseModel: *model.NewBaseModel()}
	tm.SetTable("type_test_models")
	tm.SetConnection("test")
	tm.DetectConfigFromStruct(tm)
	return tm
}

func TestTormTagTypes(t *testing.T) {
	// 设置测试数据库
	testDB := "./test_torm_tag.db"
	defer os.Remove(testDB) // 测试后清理

	config := &db.Config{
		Driver:   "sqlite",
		Database: testDB,
	}

	err := db.AddConnection("test", config)
	if err != nil {
		t.Fatalf("数据库配置失败: %v", err)
	}

	// 创建测试模型
	testModel := NewTypeTestModel()

	// 测试 AutoMigrate
	t.Run("类型长度和精度测试", func(t *testing.T) {
		err := testModel.AutoMigrate()
		if err != nil {
			t.Errorf("AutoMigrate 失败: %v", err)
		} else {
			t.Log("类型测试模型创建成功")

			// 验证表是否确实被创建
			conn, err := db.DB("test")
			if err != nil {
				t.Fatalf("获取数据库连接失败: %v", err)
			}

			// 检查表是否存在
			query := "SELECT name FROM sqlite_master WHERE type='table' AND name='type_test_models'"
			rows, err := conn.Query(query)
			if err != nil {
				t.Errorf("查询表信息失败: %v", err)
			} else {
				defer rows.Close()
				if rows.Next() {
					t.Log("类型测试表创建验证成功")
				} else {
					t.Error("类型测试表未被创建")
				}
			}
		}
	})

	t.Run("检查模型配置", func(t *testing.T) {
		if testModel.TableName() != "type_test_models" {
			t.Errorf("期望表名 'type_test_models'，得到 '%s'", testModel.TableName())
		}

		if testModel.HasModelStruct() {
			modelName := testModel.GetModelStructName()
			if modelName != "TypeTestModel" {
				t.Errorf("期望模型名 'TypeTestModel'，得到 '%s'", modelName)
			} else {
				t.Logf("模型结构体信息保存成功: %s", modelName)
			}
		} else {
			t.Error("模型结构体信息未保存")
		}
	})
}
