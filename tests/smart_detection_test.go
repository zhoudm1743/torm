package tests

import (
	"reflect"
	"strings"
	"testing"

	"github.com/zhoudm1743/torm/db"
	"github.com/zhoudm1743/torm/model"
)

// SmartTestModel 测试智能检测的模型
type SmartTestModel struct {
	model.BaseModel
	ID    int64  `torm:"primary_key,auto_increment,comment:主键ID"`
	Name  string `torm:"type:varchar,size:100,not_null,comment:名称"`
	Email string `torm:"type:varchar,size:150,unique,comment:邮箱"`
}

// NewSmartTestModelOldWay 旧的方式创建模型
func NewSmartTestModelOldWay() *SmartTestModel {
	m := &SmartTestModel{}
	m.BaseModel = *model.NewBaseModel() // 这里会触发智能检测的上下文提示
	m.SetTable("smart_test_old")
	m.SetConnection("default")
	return m
}

// NewSmartTestModelNewWay 新推荐的方式创建模型
func NewSmartTestModelNewWay() *SmartTestModel {
	user := &SmartTestModel{}
	user.BaseModel = *model.NewAutoMigrateModel(user) // 推荐方式
	user.SetTable("smart_test_new")
	user.SetConnection("default")
	return user
}

// NewSmartTestModelManualWay 手动设置方式
func NewSmartTestModelManualWay() *SmartTestModel {
	m := &SmartTestModel{}
	m.BaseModel = *model.NewBaseModel()
	m.SetTable("smart_test_manual")
	m.SetConnection("default")
	m.SetModelStruct(reflect.TypeOf(*m)) // 手动设置模型结构
	return m
}

func TestSmartDetection(t *testing.T) {
	// 配置测试数据库
	config := &db.Config{
		Driver:   "sqlite",
		Database: "test_smart_detection.db",
	}

	err := db.AddConnection("default", config)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	t.Run("OldWay_WithFriendlyError", func(t *testing.T) {
		// 测试旧方式会产生友好的错误提示
		model := NewSmartTestModelOldWay()

		err := model.AutoMigrate()
		if err != nil {
			// 应该包含友好的错误提示
			if !strings.Contains(err.Error(), "💡 Quick fix") {
				t.Errorf("Expected friendly error message, got: %v", err)
			}
			if !strings.Contains(err.Error(), "NewAutoMigrateModel") {
				t.Errorf("Expected suggestion for NewAutoMigrateModel, got: %v", err)
			}
			t.Logf("✅ Friendly error message provided: %v", err)
		} else {
			t.Error("Expected error due to missing model structure detection")
		}
	})

	t.Run("NewWay_Success", func(t *testing.T) {
		// 测试新推荐方式可以成功
		model := NewSmartTestModelNewWay()

		err := model.AutoMigrate()
		if err != nil {
			t.Errorf("NewAutoMigrateModel should work seamlessly: %v", err)
		} else {
			t.Log("✅ NewAutoMigrateModel works perfectly")
		}

		// 验证表是否创建
		conn, _ := db.DB("default")
		conn.Connect()

		var count int
		row := conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='smart_test_new'")
		err = row.Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check table existence: %v", err)
		}

		if count != 1 {
			t.Error("Table should be created successfully")
		} else {
			t.Log("✅ Table created successfully")
		}

		// 清理
		conn.Exec("DROP TABLE IF EXISTS smart_test_new")
	})

	t.Run("ManualWay_Success", func(t *testing.T) {
		// 测试手动设置方式可以成功
		model := NewSmartTestModelManualWay()

		err := model.AutoMigrate()
		if err != nil {
			t.Errorf("Manual SetModelStruct should work: %v", err)
		} else {
			t.Log("✅ Manual SetModelStruct works correctly")
		}

		// 验证表是否创建
		conn, _ := db.DB("default")
		conn.Connect()

		var count int
		row := conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='smart_test_manual'")
		err = row.Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check table existence: %v", err)
		}

		if count != 1 {
			t.Error("Table should be created successfully")
		} else {
			t.Log("✅ Manual setup table created successfully")
		}

		// 清理
		conn.Exec("DROP TABLE IF EXISTS smart_test_manual")
	})

	t.Run("ContextHints_Detection", func(t *testing.T) {
		// 测试上下文提示功能
		baseModel := model.NewBaseModel()

		// 检查是否设置了构造函数调用的上下文提示
		// 注意：这个测试可能不会触发，因为它不是在模型构造函数中调用的
		hint := baseModel.GetContextHint("constructor_call")
		t.Logf("📝 Constructor call hint: %v", hint)

		// 测试手动设置上下文提示
		baseModel.SetContextHint("test_key", "test_value")
		value := baseModel.GetContextHint("test_key")
		if value != "test_value" {
			t.Errorf("Expected 'test_value', got %v", value)
		} else {
			t.Log("✅ Context hints work correctly")
		}
	})

	t.Run("DirectAutoMigrate_WithoutSetup", func(t *testing.T) {
		// 测试直接调用AutoMigrate而没有任何设置
		baseModel := model.NewBaseModel()
		baseModel.SetTable("direct_test")
		baseModel.SetConnection("default")

		err := baseModel.AutoMigrate()
		if err != nil {
			// 应该包含详细的使用指导
			if !strings.Contains(err.Error(), "NewAutoMigrateModel") {
				t.Errorf("Expected guidance for NewAutoMigrateModel, got: %v", err)
			}
			t.Logf("✅ Helpful guidance provided: %v", err)
		} else {
			t.Error("Expected error due to missing model structure")
		}
	})
}

func TestSmartDetection_APIComparison(t *testing.T) {
	t.Run("API_Comparison", func(t *testing.T) {
		t.Log("📊 API Comparison Demo:")
		t.Log("")

		// 演示不同的API方式
		t.Log("🔴 Old Way (will show friendly error):")
		t.Log("   user.BaseModel = *model.NewBaseModel()")
		t.Log("   user.AutoMigrate() // ❌ Requires manual setup")
		t.Log("")

		t.Log("🟢 New Recommended Way:")
		t.Log("   user.BaseModel = *model.NewAutoMigrateModel(user)")
		t.Log("   user.AutoMigrate() // ✅ Works seamlessly")
		t.Log("")

		t.Log("🟡 Manual Way:")
		t.Log("   user.BaseModel = *model.NewBaseModel()")
		t.Log("   user.SetModelStruct(reflect.TypeOf(*user))")
		t.Log("   user.AutoMigrate() // ✅ Works with manual setup")
		t.Log("")

		t.Log("🔵 Traditional Way (still supported):")
		t.Log("   user.BaseModel = *model.NewBaseModelWithAutoDetect(user)")
		t.Log("   user.AutoMigrate() // ✅ Works with explicit detection")
	})
}

// GetContextHint 为测试暴露内部方法
func (m *SmartTestModel) GetContextHint(key string) interface{} {
	return m.BaseModel.GetContextHint(key)
}

// SetContextHint 为测试暴露内部方法
func (m *SmartTestModel) SetContextHint(key string, value interface{}) {
	m.BaseModel.SetContextHint(key, value)
}
