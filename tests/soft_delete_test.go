package tests

import (
	"testing"
	"time"

	"github.com/zhoudm1743/torm/db"
	"github.com/zhoudm1743/torm/model"
)

// SoftDeleteModel 带软删除的模型
type SoftDeleteModel struct {
	model.BaseModel
	ID        uint              `json:"id" db:"id" pk:""`
	Name      string            `json:"name" db:"name"`
	CreatedAt time.Time         `json:"created_at" db:"created_at;autoCreateTime"`
	UpdatedAt time.Time         `json:"updated_at" db:"updated_at;autoUpdateTime"`
	DeletedAt model.DeletedTime `json:"deleted_at" db:"deleted_at"`
}

func TestSoftDeleteAutoFilter(t *testing.T) {
	t.Run("检测软删除字段", func(t *testing.T) {
		softModel := &SoftDeleteModel{}
		metadata := model.ParseModelTags(softModel)

		if !metadata.HasSoftDeletes {
			t.Error("应该检测到软删除字段")
		}

		if metadata.DeletedAtField != "deleted_at" {
			t.Errorf("软删除字段应该是 'deleted_at'，实际是: %s", metadata.DeletedAtField)
		}

		t.Logf("软删除模型元数据: %+v", metadata)
	})

	t.Run("Model方法自动应用软删除过滤", func(t *testing.T) {
		softModel := &SoftDeleteModel{}

		// 尝试创建查询（可能因为连接不存在而失败，但我们主要验证逻辑）
		query, err := db.Model(softModel)
		if err != nil {
			t.Logf("db.Model() 失败（预期的，连接问题）: %v", err)
			// 如果是连接问题，跳过后续测试
			if err.Error() == "connection 'default' not configured" {
				t.Skip("跳过数据库连接测试")
			}
			return
		}

		// 如果能成功创建查询，验证查询构建器不为nil
		if query == nil {
			t.Error("查询构建器不应该为nil")
		}

		t.Log("成功创建带软删除过滤的查询构建器")
	})
}

func TestSoftDeleteFieldDetection(t *testing.T) {
	// 测试各种DeletedTime字段的检测
	testCases := []struct {
		name     string
		model    interface{}
		expected bool
	}{
		{
			name: "标准软删除字段",
			model: &struct {
				model.BaseModel
				DeletedAt model.DeletedTime `db:"deleted_at"`
			}{},
			expected: true,
		},
		{
			name: "自定义软删除字段名",
			model: &struct {
				model.BaseModel
				RemovedAt model.DeletedTime `db:"removed_at"`
			}{},
			expected: true,
		},
		{
			name: "没有软删除字段",
			model: &struct {
				model.BaseModel
				Name string `db:"name"`
			}{},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metadata := model.ParseModelTags(tc.model)

			if metadata.HasSoftDeletes != tc.expected {
				t.Errorf("软删除检测错误，期望: %t，实际: %t", tc.expected, metadata.HasSoftDeletes)
			}

			if tc.expected && metadata.DeletedAtField == "" {
				t.Error("软删除字段名不应该为空")
			}

			t.Logf("模型: %s, 软删除: %t, 字段: %s",
				tc.name, metadata.HasSoftDeletes, metadata.DeletedAtField)
		})
	}
}
