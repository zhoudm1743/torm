package tests

import (
	"testing"

	"github.com/zhoudm1743/torm/db"
	"github.com/zhoudm1743/torm/model"
)

// EmptyModel 空模型（只有BaseModel）
type EmptyModel struct {
	model.BaseModel
}

// NoTagsModel 没有任何标签的模型
type NoTagsModel struct {
	model.BaseModel
	ID   uint
	Name string
}

// InvalidTagsModel 有无效标签的模型
type InvalidTagsModel struct {
	model.BaseModel
	ID   uint   `db:""`  // 空db标签
	Name string `db:"-"` // 忽略的字段
}

// CustomPKModel 自定义主键的模型
type CustomPKModel struct {
	model.BaseModel
	UUID string `json:"uuid" db:"uuid" pk:""`
	Name string `json:"name" db:"name"`
}

// MultiPKModel 复合主键模型
type MultiPKModel struct {
	model.BaseModel
	TenantID uint   `json:"tenant_id" db:"tenant_id" pk:""`
	UserID   uint   `json:"user_id" db:"user_id" pk:""`
	Name     string `json:"name" db:"name"`
}

func TestEdgeCases(t *testing.T) {
	t.Run("空模型", func(t *testing.T) {
		empty := &EmptyModel{}
		metadata := model.ParseModelTags(empty)

		if metadata == nil {
			t.Error("元数据不应该为nil")
			return
		}

		// 应该有默认的主键
		if len(metadata.PrimaryKeys) == 0 {
			t.Error("应该有默认的主键字段")
		}

		// 不应该有时间戳
		if metadata.HasTimestamps {
			t.Error("空模型不应该有时间戳")
		}

		// 不应该有软删除
		if metadata.HasSoftDeletes {
			t.Error("空模型不应该有软删除")
		}

		t.Logf("空模型元数据: %+v", metadata)
	})

	t.Run("无标签模型", func(t *testing.T) {
		noTags := &NoTagsModel{}
		metadata := model.ParseModelTags(noTags)

		if metadata == nil {
			t.Error("元数据不应该为nil")
			return
		}

		// 检查字段标签是否正确解析
		if metadata.FieldTags["ID"] == nil {
			t.Error("ID字段应该被解析")
		} else {
			// 没有pk标签，不应该是主键
			if metadata.FieldTags["ID"].PrimaryKey {
				t.Error("没有pk标签的ID字段不应该是主键")
			}
		}

		t.Logf("无标签模型元数据: %+v", metadata)
	})

	t.Run("无效标签模型", func(t *testing.T) {
		invalid := &InvalidTagsModel{}
		metadata := model.ParseModelTags(invalid)

		if metadata == nil {
			t.Error("元数据不应该为nil")
			return
		}

		// 空db标签的字段应该被跳过
		// 忽略的字段（db:"-"）应该被跳过
		expectedFields := 0
		for _, fieldTag := range metadata.FieldTags {
			if fieldTag.FieldName != "" && fieldTag.FieldName != "-" {
				expectedFields++
			}
		}

		t.Logf("无效标签模型有效字段数: %d", expectedFields)
		t.Logf("无效标签模型元数据: %+v", metadata)
	})

	t.Run("自定义主键模型", func(t *testing.T) {
		customPK := &CustomPKModel{}
		metadata := model.ParseModelTags(customPK)

		if metadata == nil {
			t.Error("元数据不应该为nil")
			return
		}

		// 应该识别UUID为主键
		if len(metadata.PrimaryKeys) != 1 {
			t.Errorf("应该有1个主键，实际有: %d", len(metadata.PrimaryKeys))
		} else if metadata.PrimaryKeys[0] != "uuid" {
			t.Errorf("主键应该是 'uuid'，实际是: %s", metadata.PrimaryKeys[0])
		}

		t.Logf("自定义主键模型元数据: %+v", metadata)
	})

	t.Run("复合主键模型", func(t *testing.T) {
		multiPK := &MultiPKModel{}
		metadata := model.ParseModelTags(multiPK)

		if metadata == nil {
			t.Error("元数据不应该为nil")
			return
		}

		// 应该识别两个主键
		if len(metadata.PrimaryKeys) != 2 {
			t.Errorf("应该有2个主键，实际有: %d", len(metadata.PrimaryKeys))
		}

		// 验证主键字段
		expectedPKs := map[string]bool{"tenant_id": true, "user_id": true}
		for _, pk := range metadata.PrimaryKeys {
			if !expectedPKs[pk] {
				t.Errorf("意外的主键字段: %s", pk)
			}
		}

		t.Logf("复合主键模型元数据: %+v", metadata)
	})
}

// UserModel 用于表名推断测试
type UserModel struct{ model.BaseModel }
type ProductModel struct{ model.BaseModel }
type CategoryModel struct{ model.BaseModel }

func TestTableNameInferenceEdgeCases(t *testing.T) {
	testCases := []struct {
		name          string
		model         interface{}
		expectedTable string
	}{
		{"UserModel", &UserModel{}, "usermodels"},
		{"ProductModel", &ProductModel{}, "productmodels"},
		{"CategoryModel", &CategoryModel{}, "categorymodels"},
		{"EmptyModel", &EmptyModel{}, "emptymodels"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 这里我们测试表名推断逻辑
			// 注意：由于我们的getTableNameFromModel在db包中不能直接调用，
			// 我们使用db.Model来间接测试
			_, err := db.Model(tc.model)

			// 由于连接可能不存在，我们主要验证错误信息
			if err != nil {
				t.Logf("模型 %s 的表名推断测试: %v", tc.name, err)
				// 如果是"cannot determine table name"错误，说明表名推断有问题
				if err.Error() == "cannot determine table name from model" {
					t.Errorf("模型 %s 的表名推断失败", tc.name)
				}
			}
		})
	}
}

func TestNilModelHandling(t *testing.T) {
	t.Run("Model方法的nil检查", func(t *testing.T) {
		_, err := db.Model(nil)
		if err == nil {
			t.Error("传入nil模型应该返回错误")
		}

		if err.Error() != "model cannot be nil" {
			t.Errorf("错误信息不正确，期望: 'model cannot be nil'，实际: '%s'", err.Error())
		}
	})

	t.Run("ParseModelTags的nil检查", func(t *testing.T) {
		metadata := model.ParseModelTags(nil)
		if metadata == nil {
			t.Error("即使传入nil，也应该返回有效的元数据结构")
		}
	})
}
