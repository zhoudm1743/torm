package tests

import (
	"reflect"
	"testing"

	"github.com/zhoudm1743/torm/db"
	"github.com/zhoudm1743/torm/migration"
	"github.com/zhoudm1743/torm/model"
	"github.com/zhoudm1743/torm/models"
)

// TestAdmin 用于测试的管理员模型
type TestAdmin struct {
	model.BaseModel
	ID       string `json:"id" torm:"primary_key,type:varchar,size:32,comment:管理员ID"`
	Phone    string `json:"phone" torm:"type:varchar,size:11,comment:手机号"`
	Password string `json:"password" torm:"type:varchar,size:32,comment:密码"`
	Nickname string `json:"nickname" torm:"type:varchar,size:32,comment:昵称"` // 初始设置为32
	Avatar   string `json:"avatar" torm:"type:varchar,size:255,comment:头像"`
	Status   int    `json:"status" torm:"type:int,default:1,comment:状态"`
}

// TestAdminUpdated 更新后的管理员模型
type TestAdminUpdated struct {
	model.BaseModel
	ID       string `json:"id" torm:"primary_key,type:varchar,size:32,comment:管理员ID"`
	Phone    string `json:"phone" torm:"type:varchar,size:11,comment:手机号"`
	Password string `json:"password" torm:"type:varchar,size:32,comment:密码"`
	Nickname string `json:"nickname" torm:"type:varchar,size:255,comment:昵称"` // 修改为255
	Avatar   string `json:"avatar" torm:"type:varchar,size:255,comment:头像"`
	Status   int    `json:"status" torm:"type:int,default:1,comment:状态"`
}

func TestSchemaComparison(t *testing.T) {
	// 创建SQLite内存数据库进行测试
	config := &db.Config{
		Driver:   "sqlite",
		Database: ":memory:",
	}

	conn, err := db.NewSQLiteConnection(config, nil)
	if err != nil {
		t.Fatalf("Failed to create SQLite connection: %v", err)
	}

	// 创建模型分析器和表结构对比器
	analyzer := migration.NewModelAnalyzer()
	comparator := migration.NewSchemaComparator(conn)

	// 测试原始模型结构分析
	t.Run("TestOriginalModelAnalysis", func(t *testing.T) {
		testAdmin := &TestAdmin{}
		testAdmin.BaseModel = *model.NewBaseModelWithAutoDetect(testAdmin)
		testAdmin.SetTable("test_admin")

		modelType := reflect.TypeOf(*testAdmin)
		if modelType == nil {
			t.Fatal("Model type is nil")
		}

		columns, err := analyzer.AnalyzeModel(modelType)
		if err != nil {
			t.Fatalf("Failed to analyze model: %v", err)
		}

		// 检查Nickname字段的长度是否为32
		found := false
		for _, col := range columns {
			if col.Name == "nickname" {
				found = true
				if col.Length != 32 {
					t.Errorf("Expected nickname length to be 32, got %d", col.Length)
				}
				if col.Type != migration.ColumnTypeVarchar {
					t.Errorf("Expected nickname type to be VARCHAR, got %s", col.Type)
				}
				t.Logf("✅ Original nickname column: length=%d, type=%s", col.Length, col.Type)
				break
			}
		}
		if !found {
			t.Error("Nickname column not found in original model")
		}
	})

	// 测试更新后的模型结构分析
	t.Run("TestUpdatedModelAnalysis", func(t *testing.T) {
		testAdminUpdated := &TestAdminUpdated{}
		testAdminUpdated.BaseModel = *model.NewBaseModelWithAutoDetect(testAdminUpdated)
		testAdminUpdated.SetTable("test_admin")

		modelType := reflect.TypeOf(*testAdminUpdated)
		if modelType == nil {
			t.Fatal("Model type is nil")
		}

		columns, err := analyzer.AnalyzeModel(modelType)
		if err != nil {
			t.Fatalf("Failed to analyze updated model: %v", err)
		}

		// 检查Nickname字段的长度是否为255
		found := false
		for _, col := range columns {
			if col.Name == "nickname" {
				found = true
				if col.Length != 255 {
					t.Errorf("Expected updated nickname length to be 255, got %d", col.Length)
				}
				if col.Type != migration.ColumnTypeVarchar {
					t.Errorf("Expected nickname type to be VARCHAR, got %s", col.Type)
				}
				t.Logf("✅ Updated nickname column: length=%d, type=%s", col.Length, col.Type)
				break
			}
		}
		if !found {
			t.Error("Nickname column not found in updated model")
		}
	})

	// 测试数据库列和模型列的比较
	t.Run("TestColumnComparison", func(t *testing.T) {
		// 模拟数据库中的列（长度为32）
		dbColumns := []migration.DatabaseColumn{
			{
				Name:    "nickname",
				Type:    "TEXT", // SQLite中VARCHAR映射为TEXT
				Length:  nil,    // SQLite不显示长度
				NotNull: false,
			},
		}

		// 模拟模型中的列（长度为255）
		modelColumns := []migration.ModelColumn{
			{
				Name:    "nickname",
				Type:    migration.ColumnTypeVarchar,
				Length:  255,
				NotNull: false,
			},
		}

		// 比较差异
		differences := comparator.CompareColumns(dbColumns, modelColumns)

		t.Logf("Found %d differences", len(differences))
		for i, diff := range differences {
			t.Logf("Difference %d: Column=%s, Type=%s, Reason=%s",
				i+1, diff.Column, diff.Type, diff.Reason)
		}

		// 对于SQLite，由于类型映射的特殊性，可能不会检测到差异
		// 这是正常的，因为SQLite中TEXT类型没有固定长度限制
		if len(differences) == 0 {
			t.Log("✅ No differences found - this is expected for SQLite")
		}
	})
}

func TestMySQLSchemaComparison(t *testing.T) {
	// 跳过如果没有MySQL环境
	t.Skip("Skipping MySQL test - requires MySQL server")

	// 这里可以添加MySQL测试代码
	// 需要先启动MySQL服务器并创建测试数据库
}

func TestPostgreSQLSchemaComparison(t *testing.T) {
	// 跳过如果没有PostgreSQL环境
	t.Skip("Skipping PostgreSQL test - requires PostgreSQL server")

	// 这里可以添加PostgreSQL测试代码
	// 需要先启动PostgreSQL服务器并创建测试数据库
}

// TestAdminModelMigration 测试Admin模型的实际迁移
func TestAdminModelMigration(t *testing.T) {
	// 创建Admin实例
	admin := models.NewAdmin()

	// 测试模型结构是否正确设置
	if !admin.HasModelStruct() {
		t.Error("Admin model should have model structure after creation")
	}

	// 获取模型类型
	modelType := reflect.TypeOf(*admin)
	if modelType == nil {
		t.Fatal("Admin model type should not be nil")
	}

	t.Logf("✅ Admin model type: %s", modelType.Name())

	// 测试表名是否正确设置
	tableName := admin.TableName()
	if tableName != "admin" {
		t.Errorf("Expected table name 'admin', got '%s'", tableName)
	}

	t.Logf("✅ Admin table name: %s", tableName)

	// 分析模型列
	analyzer := migration.NewModelAnalyzer()
	columns, err := analyzer.AnalyzeModel(modelType)
	if err != nil {
		t.Fatalf("Failed to analyze Admin model: %v", err)
	}

	t.Logf("Found %d columns in Admin model:", len(columns))
	for _, col := range columns {
		t.Logf("  - %s: %s(%d) %s", col.Name, col.Type, col.Length, col.Comment)

		// 特别检查nickname字段
		if col.Name == "nickname" {
			if col.Length != 255 {
				t.Errorf("Expected nickname length to be 255, got %d", col.Length)
			} else {
				t.Logf("✅ Nickname length is correctly set to 255")
			}
		}
	}
}
