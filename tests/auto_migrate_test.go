package tests

import (
	"os"
	"testing"

	"github.com/zhoudm1743/torm/db"
	"github.com/zhoudm1743/torm/model"
)

// TestModel 测试模型
type TestModel struct {
	model.BaseModel
	ID        int64                  `json:"id" db:"id" torm:"primary_key,auto_increment,comment:用户ID"`
	Name      string                 `json:"name" db:"name" torm:"size:100,unique,comment:姓名"`
	Email     *string                `json:"email" db:"email" torm:"size:150,nullable,comment:邮箱（可空）"`
	Age       int                    `json:"age" db:"age" torm:"default:0,comment:年龄"`
	Score     float64                `json:"score" db:"score" torm:"type:decimal,precision:10,scale:2,comment:分数"`
	IsActive  bool                   `json:"is_active" db:"is_active" torm:"default:true,comment:是否激活"`
	Data      []byte                 `json:"data" db:"data" torm:"comment:二进制数据"`
	Tags      []string               `json:"tags" db:"tags" torm:"comment:标签列表"`
	Profile   map[string]interface{} `json:"profile" db:"profile" torm:"comment:用户资料"`
	UserID    *int64                 `json:"user_id" db:"user_id" torm:"index,comment:关联用户ID"`
	CreatedAt int64                  `json:"created_at" db:"created_at" torm:"auto_create_time,comment:创建时间"`
	UpdatedAt int64                  `json:"updated_at" db:"updated_at" torm:"auto_update_time,comment:更新时间"`
}

// NewTestModel 创建测试模型实例
func NewTestModel() *TestModel {
	tm := &TestModel{BaseModel: *model.NewBaseModel()}
	tm.SetTable("test_models")
	tm.SetPrimaryKey("id")
	tm.SetConnection("test")
	tm.DetectConfigFromStruct(tm)
	return tm
}

func TestAutoMigrate(t *testing.T) {
	// 设置测试数据库
	testDB := "./test_auto_migrate.db"
	defer os.Remove(testDB) // 测试后清理

	config := &db.Config{
		Driver:   "sqlite",
		Database: testDB,
	}

	if err := db.AddConnection("test", config); err != nil {
		t.Fatalf("添加测试数据库配置失败: %v", err)
	}

	// 创建测试模型
	testModel := NewTestModel()

	// 测试 AutoMigrate 方法
	t.Run("AutoMigrate基础功能", func(t *testing.T) {
		err := testModel.AutoMigrate()
		if err != nil {
			t.Errorf("AutoMigrate 失败: %v", err)
		} else {
			t.Log("AutoMigrate 执行成功")

			// 验证表是否确实被创建
			conn, err := db.DB("test")
			if err != nil {
				t.Fatalf("获取数据库连接失败: %v", err)
			}

			// 检查表是否存在 - 通过查询表结构
			query := "SELECT name FROM sqlite_master WHERE type='table' AND name='test_models'"
			rows, err := conn.Query(query)
			if err != nil {
				t.Errorf("查询表信息失败: %v", err)
			} else {
				defer rows.Close()
				if rows.Next() {
					t.Log("表创建验证成功")
				} else {
					t.Error("表未被创建")
				}
			}
		}
	})

	// 测试模型配置检测
	t.Run("模型配置检测", func(t *testing.T) {
		if testModel.TableName() != "test_models" {
			t.Errorf("期望表名 'test_models'，得到 '%s'", testModel.TableName())
		}

		if testModel.PrimaryKey() != "id" {
			t.Errorf("期望主键 'id'，得到 '%s'", testModel.PrimaryKey())
		}

		if testModel.GetConnection() != "test" {
			t.Errorf("期望连接名 'test'，得到 '%s'", testModel.GetConnection())
		}
	})

	// 测试结构体信息检测
	t.Run("结构体信息保存", func(t *testing.T) {
		// 检查是否已设置模型结构体信息
		if !testModel.HasModelStruct() {
			t.Error("模型结构体信息未设置")
		} else {
			structName := testModel.GetModelStructName()
			if structName == "" {
				t.Error("模型结构体名称为空")
			} else {
				t.Logf("模型结构体信息保存成功: %s", structName)

				if structName != "TestModel" {
					t.Errorf("期望结构体名称 'TestModel'，得到 '%s'", structName)
				}
			}
		}
	})
}
