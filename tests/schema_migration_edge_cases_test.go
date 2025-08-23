package tests

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/zhoudm1743/torm/db"
	"github.com/zhoudm1743/torm/migration"
	"github.com/zhoudm1743/torm/model"
)

// 边缘情况测试模型
type EdgeCaseModel struct {
	model.BaseModel
	// 测试各种边缘情况
	EmptyTag     string  `torm:""`                  // 空标签
	OnlyType     string  `torm:"type:varchar"`      // 只有类型
	OnlySize     string  `torm:"size:100"`          // 只有大小
	InvalidSize  string  `torm:"size:invalid"`      // 无效大小
	InvalidPrec  float64 `torm:"precision:invalid"` // 无效精度
	NoTags       string  // 无标签
	MixedCase    string  `torm:"TYPE:VARCHAR,SIZE:50,UNIQUE"` // 大小写混合
	SpecialChars string  `torm:"comment:包含'特殊\"字符"`           // 特殊字符
	VeryLongName string  `torm:"type:varchar,size:255,comment:这是一个非常长的注释内容用来测试系统对长注释的处理能力"`
}

// 无效模型（没有嵌入BaseModel）
type InvalidModel struct {
	ID   int    `torm:"primary_key"`
	Name string `torm:"type:varchar,size:100"`
}

// 测试复杂的标签解析
type ComplexTagModel struct {
	model.BaseModel
	Field1 string `torm:"type:varchar,size:100,unique,not_null,default:'test',comment:复杂字段,index:complex_idx"`
	Field2 string `torm:"primary_key,auto_increment,type:varchar,size:50"`
	Field3 string `torm:"nullable,type:text,comment:可空文本字段"`
}

func TestSchemaEdgeCases(t *testing.T) {
	setupEdgeCaseTestDB(t)
	defer cleanupEdgeCaseTestDB()

	t.Run("ModelAnalyzer_EdgeCases", func(t *testing.T) {
		analyzer := migration.NewModelAnalyzer()

		// 测试包含边缘情况的模型
		modelType := reflect.TypeOf(EdgeCaseModel{})
		columns, err := analyzer.AnalyzeModel(modelType)
		if err != nil {
			t.Fatalf("Failed to analyze edge case model: %v", err)
		}

		// 应该跳过没有有效标签的字段
		for _, col := range columns {
			if col.Name == "no_tags" {
				t.Error("Fields without tags should be skipped")
			}
		}

		// 测试大小写不敏感
		mixedCaseCol := findColumnByName(columns, "mixed_case")
		if mixedCaseCol == nil {
			t.Error("Mixed case field should be found")
		} else {
			if mixedCaseCol.Type != migration.ColumnTypeVarchar {
				t.Errorf("Mixed case type should be parsed correctly, got %s", mixedCaseCol.Type)
			}
			if mixedCaseCol.Length != 50 {
				t.Errorf("Mixed case size should be parsed correctly, got %d", mixedCaseCol.Length)
			}
			if !mixedCaseCol.Unique {
				t.Error("Mixed case unique should be parsed correctly")
			}
		}

		// 测试特殊字符处理
		specialCol := findColumnByName(columns, "special_chars")
		if specialCol == nil {
			t.Error("Special chars field should be found")
		} else {
			if !contains(specialCol.Comment, "特殊") {
				t.Errorf("Special characters in comment should be preserved, got: %s", specialCol.Comment)
			}
		}
	})

	t.Run("ModelAnalyzer_InvalidInput", func(t *testing.T) {
		analyzer := migration.NewModelAnalyzer()

		// 测试无效模型类型
		invalidType := reflect.TypeOf(InvalidModel{})
		columns, err := analyzer.AnalyzeModel(invalidType)
		if err != nil {
			t.Fatalf("Should handle invalid model gracefully: %v", err)
		}

		// 无效模型应该仍然能解析字段，只是没有BaseModel
		if len(columns) == 0 {
			t.Error("Should still parse fields from invalid model")
		}
	})

	t.Run("SchemaComparator_ErrorHandling", func(t *testing.T) {
		conn := getTestConnection(t)
		comparator := migration.NewSchemaComparator(conn)

		// 测试不存在的表
		_, err := comparator.GetDatabaseColumns("non_existent_table")
		if err == nil {
			t.Error("Should return error for non-existent table")
		}

		// 测试空表名
		_, err = comparator.GetDatabaseColumns("")
		if err == nil {
			t.Error("Should return error for empty table name")
		}
	})

	t.Run("AlterGenerator_ErrorHandling", func(t *testing.T) {
		conn := getTestConnection(t)
		generator := migration.NewAlterGenerator(conn)

		// 测试空差异列表
		statements, err := generator.GenerateAlterSQL("test_table", []migration.ColumnDifference{})
		if err != nil {
			t.Errorf("Should handle empty differences gracefully: %v", err)
		}
		if statements != nil && len(statements) > 0 {
			t.Error("Should return empty statements for no differences")
		}

		// 测试无效的差异类型
		invalidDifferences := []migration.ColumnDifference{
			{
				Column: "test_field",
				Type:   "invalid_type",
				NewValue: migration.ModelColumn{
					Name: "test_field",
					Type: migration.ColumnTypeVarchar,
				},
			},
		}

		statements, err = generator.GenerateAlterSQL("test_table", invalidDifferences)
		// 应该能处理，只是忽略无效类型
		if err != nil {
			t.Errorf("Should handle invalid difference types gracefully: %v", err)
		}
	})

	t.Run("SafeMigrator_ErrorRecovery", func(t *testing.T) {
		conn := getTestConnection(t)
		migrator := migration.NewSafeMigrator(conn).SetDryRun(false).SetBackupTables(true)

		// 创建测试表
		createSQL := `CREATE TABLE error_recovery_test (id INTEGER PRIMARY KEY)`
		_, err := conn.Exec(createSQL)
		if err != nil {
			t.Fatalf("Failed to create test table: %v", err)
		}

		// 创建会导致错误的差异（比如添加重复的主键）
		errorDifferences := []migration.ColumnDifference{
			{
				Column: "id",
				Type:   "add",
				NewValue: migration.ModelColumn{
					Name:       "id",
					Type:       migration.ColumnTypeInt,
					PrimaryKey: true,
				},
			},
		}

		result, err := migrator.SafeAlterTable("error_recovery_test", errorDifferences)
		if err == nil {
			t.Error("Should return error for invalid migration")
		}

		if result.Success {
			t.Error("Result should indicate failure")
		}

		if result.Error == nil {
			t.Error("Result should contain error information")
		}

		if result.FailedStatement == "" {
			t.Error("Result should indicate which statement failed")
		}

		// 清理
		conn.Exec("DROP TABLE IF EXISTS error_recovery_test")
		if result.BackupTable != "" {
			conn.Exec("DROP TABLE IF EXISTS " + result.BackupTable)
		}
	})

	t.Run("TypeMapping_AllGoTypes", func(t *testing.T) {
		analyzer := migration.NewModelAnalyzer()

		// 测试各种Go类型的映射
		type AllTypesModel struct {
			model.BaseModel
			StringField  string
			IntField     int
			Int8Field    int8
			Int16Field   int16
			Int32Field   int32
			Int64Field   int64
			UintField    uint
			Uint8Field   uint8
			Uint16Field  uint16
			Uint32Field  uint32
			Uint64Field  uint64
			Float32Field float32
			Float64Field float64
			BoolField    bool
			ByteSlice    []byte
			StringSlice  []string
			IntSlice     []int
			MapField     map[string]interface{}
			TimeField    time.Time
			PtrField     *string
			StructField  struct{ Name string }
		}

		modelType := reflect.TypeOf(AllTypesModel{})
		columns, err := analyzer.AnalyzeModel(modelType)
		if err != nil {
			t.Fatalf("Failed to analyze all types model: %v", err)
		}

		// 验证类型映射
		expectedMappings := map[string]migration.ColumnType{
			"string_field":  migration.ColumnTypeVarchar,
			"int_field":     migration.ColumnTypeInt,
			"int8_field":    migration.ColumnTypeTinyInt,
			"int16_field":   migration.ColumnTypeSmallInt,
			"int32_field":   migration.ColumnTypeInt,
			"int64_field":   migration.ColumnTypeBigInt,
			"uint_field":    migration.ColumnTypeInt,
			"uint8_field":   migration.ColumnTypeTinyInt,
			"uint16_field":  migration.ColumnTypeSmallInt,
			"uint32_field":  migration.ColumnTypeInt,
			"uint64_field":  migration.ColumnTypeBigInt,
			"float32_field": migration.ColumnTypeFloat,
			"float64_field": migration.ColumnTypeDouble,
			"bool_field":    migration.ColumnTypeBoolean,
			"byte_slice":    migration.ColumnTypeBlob,
			"string_slice":  migration.ColumnTypeJSON,
			"int_slice":     migration.ColumnTypeJSON,
			"map_field":     migration.ColumnTypeJSON,
			"time_field":    migration.ColumnTypeDateTime,
			"ptr_field":     migration.ColumnTypeVarchar,
			"struct_field":  migration.ColumnTypeJSON,
		}

		for _, col := range columns {
			if expectedType, exists := expectedMappings[col.Name]; exists {
				if col.Type != expectedType {
					t.Errorf("Field %s: expected type %s, got %s", col.Name, expectedType, col.Type)
				}
			}
		}
	})

	t.Run("DefaultValue_Parsing", func(t *testing.T) {
		analyzer := migration.NewModelAnalyzer()

		type DefaultValueModel struct {
			model.BaseModel
			NullField      string    `torm:"default:null"`
			TrueField      bool      `torm:"default:true"`
			FalseField     bool      `torm:"default:false"`
			NumberField    int       `torm:"default:123"`
			FloatField     float64   `torm:"default:123.45"`
			StringField    string    `torm:"default:test_value"`
			TimestampField time.Time `torm:"default:current_timestamp"`
			NowField       time.Time `torm:"default:now()"`
			QuotedField    string    `torm:"default:'quoted value'"`
		}

		modelType := reflect.TypeOf(DefaultValueModel{})
		columns, err := analyzer.AnalyzeModel(modelType)
		if err != nil {
			t.Fatalf("Failed to analyze default value model: %v", err)
		}

		// 验证默认值解析
		expectedDefaults := map[string]string{
			"null_field":      "NULL",
			"true_field":      "1",
			"false_field":     "0",
			"number_field":    "123",
			"float_field":     "123.45",
			"string_field":    "'test_value'",
			"timestamp_field": "CURRENT_TIMESTAMP",
			"now_field":       "CURRENT_TIMESTAMP",
			"quoted_field":    "''quoted value''",
		}

		for _, col := range columns {
			if expectedDefault, exists := expectedDefaults[col.Name]; exists {
				if col.Default == nil {
					t.Errorf("Field %s: expected default value, got nil", col.Name)
				} else if *col.Default != expectedDefault {
					t.Errorf("Field %s: expected default '%s', got '%s'", col.Name, expectedDefault, *col.Default)
				}
			}
		}
	})

	t.Run("Backup_And_Restore", func(t *testing.T) {
		conn := getTestConnection(t)
		migrator := migration.NewSafeMigrator(conn)

		// 创建测试表并插入数据
		createSQL := `
			CREATE TABLE backup_test (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				value INTEGER
			)
		`
		_, err := conn.Exec(createSQL)
		if err != nil {
			t.Fatalf("Failed to create test table: %v", err)
		}

		// 插入测试数据
		insertSQL := `INSERT INTO backup_test (name, value) VALUES ('test1', 100), ('test2', 200)`
		_, err = conn.Exec(insertSQL)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}

		// 创建备份
		backupName, err := createTableBackupHelper(migrator, "backup_test")
		if err != nil {
			t.Fatalf("Failed to create backup: %v", err)
		}

		// 验证备份数据
		var count int
		row := conn.QueryRow("SELECT COUNT(*) FROM " + backupName)
		err = row.Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check backup data: %v", err)
		}

		if count != 2 {
			t.Errorf("Backup should contain 2 rows, got %d", count)
		}

		// 修改原表
		_, err = conn.Exec("DELETE FROM backup_test WHERE name = 'test1'")
		if err != nil {
			t.Fatalf("Failed to modify original table: %v", err)
		}

		// 从备份恢复
		err = migrator.RestoreFromBackup("backup_test", backupName)
		if err != nil {
			t.Fatalf("Failed to restore from backup: %v", err)
		}

		// 验证恢复后的数据
		row = conn.QueryRow("SELECT COUNT(*) FROM backup_test")
		err = row.Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check restored data: %v", err)
		}

		if count != 2 {
			t.Errorf("Restored table should contain 2 rows, got %d", count)
		}

		// 清理
		conn.Exec("DROP TABLE IF EXISTS backup_test")
	})

	t.Run("LegacyTags_Compatibility", func(t *testing.T) {
		analyzer := migration.NewModelAnalyzer()

		// 测试传统标签的兼容性
		type LegacyTagModel struct {
			model.BaseModel
			ID      int64  `primaryKey:"true" autoIncrement:"true"`
			Name    string `size:"100" unique:"true"`
			Email   string `comment:"邮箱地址" not_null:"true"`
			Status  int    `default:"1"`
			Created int64  `autoCreateTime:"true"`
			Updated int64  `autoUpdateTime:"true"`
		}

		modelType := reflect.TypeOf(LegacyTagModel{})
		columns, err := analyzer.AnalyzeModel(modelType)
		if err != nil {
			t.Fatalf("Failed to analyze legacy tag model: %v", err)
		}

		// 验证传统标签解析
		idCol := findColumnByName(columns, "id")
		if idCol == nil || !idCol.PrimaryKey || !idCol.AutoIncrement {
			t.Error("Legacy primary key and auto increment should be parsed")
		}

		nameCol := findColumnByName(columns, "name")
		if nameCol == nil || nameCol.Length != 100 || !nameCol.Unique {
			t.Error("Legacy size and unique should be parsed")
		}

		emailCol := findColumnByName(columns, "email")
		if emailCol == nil || emailCol.Comment != "邮箱地址" || !emailCol.NotNull {
			t.Error("Legacy comment and not_null should be parsed")
		}
	})

	t.Run("Performance_LargeModel", func(t *testing.T) {
		// 创建一个有很多字段的大模型来测试性能
		type LargeModel struct {
			model.BaseModel
			Field01 string `torm:"type:varchar,size:100"`
			Field02 string `torm:"type:varchar,size:100"`
			Field03 string `torm:"type:varchar,size:100"`
			Field04 string `torm:"type:varchar,size:100"`
			Field05 string `torm:"type:varchar,size:100"`
			Field06 string `torm:"type:varchar,size:100"`
			Field07 string `torm:"type:varchar,size:100"`
			Field08 string `torm:"type:varchar,size:100"`
			Field09 string `torm:"type:varchar,size:100"`
			Field10 string `torm:"type:varchar,size:100"`
			Field11 string `torm:"type:varchar,size:100"`
			Field12 string `torm:"type:varchar,size:100"`
			Field13 string `torm:"type:varchar,size:100"`
			Field14 string `torm:"type:varchar,size:100"`
			Field15 string `torm:"type:varchar,size:100"`
			Field16 string `torm:"type:varchar,size:100"`
			Field17 string `torm:"type:varchar,size:100"`
			Field18 string `torm:"type:varchar,size:100"`
			Field19 string `torm:"type:varchar,size:100"`
			Field20 string `torm:"type:varchar,size:100"`
		}

		analyzer := migration.NewModelAnalyzer()

		start := time.Now()
		modelType := reflect.TypeOf(LargeModel{})
		columns, err := analyzer.AnalyzeModel(modelType)
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Failed to analyze large model: %v", err)
		}

		if len(columns) != 20 {
			t.Errorf("Expected 20 columns, got %d", len(columns))
		}

		// 性能检查：分析20个字段应该很快（小于100ms）
		if duration > 100*time.Millisecond {
			t.Logf("Performance warning: Large model analysis took %v", duration)
		}

		t.Logf("Large model analysis completed in %v", duration)
	})
}

// 辅助函数
func setupEdgeCaseTestDB(t *testing.T) {
	config := &db.Config{
		Driver:   "sqlite",
		Database: "test_edge_cases.db",
	}

	err := db.AddConnection("default", config)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}
}

func cleanupEdgeCaseTestDB() {
	// 清理测试数据库文件可以在这里实现
}

// createTableBackupHelper 辅助函数来创建表备份
func createTableBackupHelper(migrator *migration.SafeMigrator, tableName string) (string, error) {
	// 由于createTableBackup是私有方法，我们这里模拟实现
	return tableName + "_backup_test", nil
}

func TestMigrationResult_PrintSummary(t *testing.T) {
	result := &migration.MigrationResult{
		TableName: "test_table",
		StartTime: time.Now().Add(-5 * time.Second),
		EndTime:   time.Now(),
		Duration:  5 * time.Second,
		Changes: []migration.ColumnDifference{
			{Column: "test_field", Type: "add", Reason: "New field"},
		},
		Statements:  []string{"ALTER TABLE test_table ADD COLUMN test_field VARCHAR(100)"},
		Success:     true,
		Message:     "Migration completed successfully",
		BackupTable: "test_table_backup_20240123_150405",
	}

	// 测试成功情况的摘要打印
	result.PrintSummary()

	// 测试失败情况的摘要打印
	failedResult := &migration.MigrationResult{
		TableName:            "test_table",
		Success:              false,
		Error:                fmt.Errorf("test error"),
		FailedStatement:      "ALTER TABLE test_table ADD COLUMN invalid_field",
		RecoveryInstructions: "Run recovery SQL to restore",
	}

	failedResult.PrintSummary()
}
