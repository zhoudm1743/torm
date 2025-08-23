package tests

import (
	"reflect"
	"testing"
	"time"

	"github.com/zhoudm1743/torm/db"
	"github.com/zhoudm1743/torm/migration"
	"github.com/zhoudm1743/torm/model"
)

// TestSchemaModel 测试模型
type TestSchemaModel struct {
	model.BaseModel
	ID        int64     `json:"id" torm:"primary_key,auto_increment,comment:用户ID"`
	Name      string    `json:"name" torm:"type:varchar,size:100,comment:用户名"`
	Email     string    `json:"email" torm:"type:varchar,size:150,unique,comment:邮箱"`
	Age       int       `json:"age" torm:"type:int,default:0,comment:年龄"`
	Balance   float64   `json:"balance" torm:"type:decimal,precision:10,scale:2,default:0.00,comment:余额"`
	Status    string    `json:"status" torm:"type:varchar,size:20,default:active,comment:状态"`
	CreatedAt time.Time `json:"created_at" torm:"auto_create_time,comment:创建时间"`
	UpdatedAt time.Time `json:"updated_at" torm:"auto_update_time,comment:更新时间"`
}

// NewTestSchemaModel 创建测试模型
func NewTestSchemaModel() *TestSchemaModel {
	m := &TestSchemaModel{}
	m.BaseModel = *model.NewBaseModelWithAutoDetect(m)
	m.SetTable("test_schema_users")
	m.SetConnection("default")
	return m
}

func TestSchemaComparator(t *testing.T) {
	// 配置测试数据库
	config := &db.Config{
		Driver:   "sqlite",
		Database: "test_schema.db",
	}

	err := db.AddConnection("default", config)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	conn, err := db.DB("default")
	if err != nil {
		t.Fatalf("Failed to get connection: %v", err)
	}

	err = conn.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// 清理测试表
	conn.Exec("DROP TABLE IF EXISTS test_schema_users")

	t.Run("ModelAnalyzer", func(t *testing.T) {
		analyzer := migration.NewModelAnalyzer()
		modelType := reflect.TypeOf(TestSchemaModel{})

		columns, err := analyzer.AnalyzeModel(modelType)
		if err != nil {
			t.Fatalf("Failed to analyze model: %v", err)
		}

		// 验证列数量（排除BaseModel）
		expectedColumns := 8 // ID, Name, Email, Age, Balance, Status, CreatedAt, UpdatedAt
		if len(columns) != expectedColumns {
			t.Errorf("Expected %d columns, got %d", expectedColumns, len(columns))
		}

		// 验证特定列
		nameColumn := findColumn(columns, "name")
		if nameColumn == nil {
			t.Error("Name column not found")
		} else {
			if nameColumn.Type != migration.ColumnTypeVarchar {
				t.Errorf("Expected VARCHAR type for name, got %s", nameColumn.Type)
			}
			if nameColumn.Length != 100 {
				t.Errorf("Expected length 100 for name, got %d", nameColumn.Length)
			}
		}

		balanceColumn := findColumn(columns, "balance")
		if balanceColumn == nil {
			t.Error("Balance column not found")
		} else {
			if balanceColumn.Type != migration.ColumnTypeDecimal {
				t.Errorf("Expected DECIMAL type for balance, got %s", balanceColumn.Type)
			}
			if balanceColumn.Precision != 10 {
				t.Errorf("Expected precision 10 for balance, got %d", balanceColumn.Precision)
			}
			if balanceColumn.Scale != 2 {
				t.Errorf("Expected scale 2 for balance, got %d", balanceColumn.Scale)
			}
		}
	})

	t.Run("SchemaComparator", func(t *testing.T) {
		// 先创建一个简单的表
		createSQL := `
			CREATE TABLE test_schema_users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT,
				email TEXT,
				age INTEGER
			)
		`
		_, err := conn.Exec(createSQL)
		if err != nil {
			t.Fatalf("Failed to create test table: %v", err)
		}

		comparator := migration.NewSchemaComparator(conn)

		// 获取数据库列
		dbColumns, err := comparator.GetDatabaseColumns("test_schema_users")
		if err != nil {
			t.Fatalf("Failed to get database columns: %v", err)
		}

		if len(dbColumns) != 4 {
			t.Errorf("Expected 4 database columns, got %d", len(dbColumns))
		}

		// 获取模型列
		analyzer := migration.NewModelAnalyzer()
		modelType := reflect.TypeOf(TestSchemaModel{})
		modelColumns, err := analyzer.AnalyzeModel(modelType)
		if err != nil {
			t.Fatalf("Failed to analyze model: %v", err)
		}

		// 比较差异
		differences := comparator.CompareColumns(dbColumns, modelColumns)

		// 应该有差异（模型有更多列）
		if len(differences) == 0 {
			t.Error("Expected differences between database and model")
		}

		// 验证有添加列的差异
		hasAddColumn := false
		for _, diff := range differences {
			if diff.Type == "add" {
				hasAddColumn = true
				break
			}
		}
		if !hasAddColumn {
			t.Error("Expected at least one ADD column difference")
		}
	})

	t.Run("AlterGenerator", func(t *testing.T) {
		generator := migration.NewAlterGenerator(conn)

		// 创建一些测试差异
		differences := []migration.ColumnDifference{
			{
				Column: "balance",
				Type:   "add",
				NewValue: migration.ModelColumn{
					Name:      "balance",
					Type:      migration.ColumnTypeDecimal,
					Precision: 10,
					Scale:     2,
					NotNull:   false,
					Comment:   "余额",
				},
				Reason: "New column",
			},
			{
				Column: "name",
				Type:   "modify",
				OldValue: migration.DatabaseColumn{
					Name: "name",
					Type: "TEXT",
				},
				NewValue: migration.ModelColumn{
					Name:    "name",
					Type:    migration.ColumnTypeVarchar,
					Length:  100,
					NotNull: true,
					Comment: "用户名",
				},
				Reason: "Type and constraints changed",
			},
		}

		statements, err := generator.GenerateAlterSQL("test_schema_users", differences)
		if err != nil {
			t.Fatalf("Failed to generate ALTER SQL: %v", err)
		}

		if len(statements) == 0 {
			t.Error("Expected ALTER statements to be generated")
		}

		t.Logf("Generated SQL statements:")
		for i, stmt := range statements {
			t.Logf("  %d. %s", i+1, stmt)
		}
	})

	t.Run("SafeMigrator_DryRun", func(t *testing.T) {
		safeMigrator := migration.NewSafeMigrator(conn).SetDryRun(true)

		differences := []migration.ColumnDifference{
			{
				Column: "status",
				Type:   "add",
				NewValue: migration.ModelColumn{
					Name:    "status",
					Type:    migration.ColumnTypeVarchar,
					Length:  20,
					NotNull: false,
					Comment: "状态",
				},
				Reason: "New status column",
			},
		}

		result, err := safeMigrator.SafeAlterTable("test_schema_users", differences)
		if err != nil {
			t.Fatalf("Failed to execute safe migration: %v", err)
		}

		if !result.Success {
			t.Error("Dry run should succeed")
		}

		if result.BackupTable != "" {
			t.Error("Dry run should not create backup tables")
		}
	})

	// 清理
	conn.Exec("DROP TABLE IF EXISTS test_schema_users")
}

func TestAutoMigrateWithSchemaUpdate(t *testing.T) {
	// 配置测试数据库
	config := &db.Config{
		Driver:   "sqlite",
		Database: "test_auto_migrate_schema.db",
	}

	err := db.AddConnection("default", config)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	// 清理测试表
	conn, _ := db.DB("default")
	conn.Connect()
	defer conn.Close()
	conn.Exec("DROP TABLE IF EXISTS test_schema_users")

	t.Run("FirstTimeAutoMigrate", func(t *testing.T) {
		// 第一次运行AutoMigrate，应该创建表
		model := NewTestSchemaModel()

		err := model.AutoMigrate()
		if err != nil {
			t.Fatalf("First AutoMigrate failed: %v", err)
		}

		// 验证表是否创建
		var count int
		row := conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='test_schema_users'")
		err = row.Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check table existence: %v", err)
		}

		if count != 1 {
			t.Error("Table should be created")
		}
	})

	t.Run("SecondTimeAutoMigrate_NoChanges", func(t *testing.T) {
		// 第二次运行AutoMigrate，没有变化，应该不报错
		model := NewTestSchemaModel()

		err := model.AutoMigrate()
		if err != nil {
			t.Fatalf("Second AutoMigrate failed: %v", err)
		}
	})

	// 清理
	conn.Exec("DROP TABLE IF EXISTS test_schema_users")
}

// findColumn 查找指定名称的列
func findColumn(columns []migration.ModelColumn, name string) *migration.ModelColumn {
	for _, col := range columns {
		if col.Name == name {
			return &col
		}
	}
	return nil
}
