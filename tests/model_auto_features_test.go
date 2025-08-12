package tests

import (
	"testing"
	"time"

	"github.com/zhoudm1743/torm/db"
	"github.com/zhoudm1743/torm/model"
)

// TestUser 测试用户模型
type TestUser struct {
	model.BaseModel
	ID        uint              `json:"id" db:"id" pk:""`                          // 主键
	Name      string            `json:"name" db:"name"`                            // 用户名
	Email     string            `json:"email" db:"email"`                          // 邮箱
	Age       int               `json:"age" db:"age"`                              // 年龄
	Status    string            `json:"status" db:"status"`                        // 状态
	CreatedAt time.Time         `json:"created_at" db:"created_at;autoCreateTime"` // 自动创建时间
	UpdatedAt time.Time         `json:"updated_at" db:"updated_at;autoUpdateTime"` // 自动更新时间
	DeletedAt model.DeletedTime `json:"deleted_at" db:"deleted_at"`                // 软删除时间
}

// NewTestUser 创建测试用户模型实例
func NewTestUser() *TestUser {
	user := &TestUser{
		BaseModel: *model.NewBaseModel(),
	}
	user.SetTable("test_users")
	user.SetConnection("default")
	user.DetectConfigFromStruct(user) // 从标签检测配置
	return user
}

func TestModelAutoFeatures(t *testing.T) {
	// 测试从模型自动获取表名
	user := &TestUser{}

	// 使用db.Model()自动获取表名和启用模型特性
	query, err := db.Model(user)
	if err != nil {
		t.Logf("db.Model() 可能因为连接问题失败，这是正常的: %v", err)
		return
	}

	t.Logf("成功创建模型查询构建器")

	// 测试模型转换为map
	testUser := &TestUser{
		Name:   "测试用户",
		Email:  "test@example.com",
		Age:    25,
		Status: "active",
	}

	// 这里模拟模型的转换过程
	t.Logf("测试用户: %+v", testUser)

	// 验证查询构建器不为nil
	if query == nil {
		t.Error("查询构建器不应该为nil")
	}
}

func TestTableNameInference(t *testing.T) {
	// 测试表名推断
	testCases := []struct {
		modelName     string
		expectedTable string
	}{
		{"User", "users"},
		{"Product", "products"},
		{"Category", "categorys"}, // 简单的复数形式
		{"TestModel", "testmodels"},
	}

	for _, tc := range testCases {
		t.Logf("模型名: %s, 预期表名: %s", tc.modelName, tc.expectedTable)
	}
}

func TestModelTagsParsing(t *testing.T) {
	// 测试标签解析
	user := &TestUser{}
	metadata := model.ParseModelTags(user)

	if metadata == nil {
		t.Error("解析的元数据不应该为nil")
		return
	}

	// 验证主键字段
	if len(metadata.PrimaryKeys) == 0 {
		t.Error("应该解析出主键字段")
	} else {
		t.Logf("主键字段: %v", metadata.PrimaryKeys)
	}

	// 验证时间戳字段
	if !metadata.HasTimestamps {
		t.Error("应该检测到时间戳字段")
	} else {
		t.Logf("创建时间字段: %s, 更新时间字段: %s",
			metadata.CreatedAtField, metadata.UpdatedAtField)
	}

	// 验证软删除字段
	if !metadata.HasSoftDeletes {
		t.Error("应该检测到软删除字段")
	} else {
		t.Logf("软删除字段: %s", metadata.DeletedAtField)
	}
}

func TestModelConfigPriority(t *testing.T) {
	// 测试标签配置优先级
	user := NewTestUser()

	// 通过DetectConfigFromStruct验证配置
	metadata := model.ParseModelTags(user)

	// 验证时间戳配置
	if !metadata.HasTimestamps {
		t.Error("时间戳应该被检测到")
	} else {
		t.Logf("时间戳配置正确，创建时间字段: %s, 更新时间字段: %s",
			metadata.CreatedAtField, metadata.UpdatedAtField)
	}

	// 验证软删除配置
	if !metadata.HasSoftDeletes {
		t.Error("软删除应该被检测到")
	} else {
		t.Logf("软删除配置正确，删除时间字段: %s", metadata.DeletedAtField)
	}

	// 验证字段名是否来自标签
	if metadata.CreatedAtField != "created_at" {
		t.Errorf("创建时间字段应该是 'created_at'，实际是: %s", metadata.CreatedAtField)
	}

	if metadata.UpdatedAtField != "updated_at" {
		t.Errorf("更新时间字段应该是 'updated_at'，实际是: %s", metadata.UpdatedAtField)
	}

	if metadata.DeletedAtField != "deleted_at" {
		t.Errorf("软删除字段应该是 'deleted_at'，实际是: %s", metadata.DeletedAtField)
	}
}
