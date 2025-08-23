package tests

import (
	"database/sql"
	"fmt"
	"reflect"
	"testing"

	"github.com/zhoudm1743/torm/db"
	"github.com/zhoudm1743/torm/migration"
)

// 单元测试，专注于各个组件的独立功能

func TestModelAnalyzer_Unit(t *testing.T) {
	analyzer := migration.NewModelAnalyzer()

	t.Run("hasDBTag", func(t *testing.T) {
		type TestStruct struct {
			WithTorm    string `torm:"type:varchar"`
			WithDB      string `db:"field_name"`
			WithPK      string `primaryKey:"true"`
			WithoutTags string
		}

		structType := reflect.TypeOf(TestStruct{})

		// 测试有torm标签的字段
		field1, _ := structType.FieldByName("WithTorm")
		if !analyzer.HasDBTag(field1) {
			t.Error("Should detect torm tag")
		}

		// 测试有db标签的字段
		field2, _ := structType.FieldByName("WithDB")
		if !analyzer.HasDBTag(field2) {
			t.Error("Should detect db tag")
		}

		// 测试有primaryKey标签的字段
		field3, _ := structType.FieldByName("WithPK")
		if !analyzer.HasDBTag(field3) {
			t.Error("Should detect primaryKey tag")
		}

		// 测试没有标签的字段
		field4, _ := structType.FieldByName("WithoutTags")
		if analyzer.HasDBTag(field4) {
			t.Error("Should not detect tags on field without tags")
		}
	})

	t.Run("getColumnName", func(t *testing.T) {
		type TestStruct struct {
			WithTormColumn string `torm:"column:custom_name,type:varchar"`
			WithDB         string `db:"db_field_name"`
			WithoutTags    string
			CamelCase      string
		}

		structType := reflect.TypeOf(TestStruct{})

		// 测试torm标签中的column名称
		field1, _ := structType.FieldByName("WithTormColumn")
		name1 := analyzer.GetColumnName(field1)
		if name1 != "custom_name" {
			t.Errorf("Expected 'custom_name', got '%s'", name1)
		}

		// 测试db标签
		field2, _ := structType.FieldByName("WithDB")
		name2 := analyzer.GetColumnName(field2)
		if name2 != "db_field_name" {
			t.Errorf("Expected 'db_field_name', got '%s'", name2)
		}

		// 测试驼峰转蛇形
		field3, _ := structType.FieldByName("CamelCase")
		name3 := analyzer.GetColumnName(field3)
		if name3 != "camel_case" {
			t.Errorf("Expected 'camel_case', got '%s'", name3)
		}
	})

	t.Run("extractColumnNameFromTorm", func(t *testing.T) {
		tests := []struct {
			tag      string
			expected string
		}{
			{"column:test_name,type:varchar", "test_name"},
			{"type:varchar,column:another_name", "another_name"},
			{"type:varchar,size:100", ""},
			{"", ""},
		}

		for _, test := range tests {
			result := analyzer.ExtractColumnNameFromTorm(test.tag)
			if result != test.expected {
				t.Errorf("Tag '%s': expected '%s', got '%s'", test.tag, test.expected, result)
			}
		}
	})

	t.Run("parseTormKeyValue", func(t *testing.T) {
		column := &migration.ModelColumn{}

		tests := []struct {
			input     string
			checkFunc func(*migration.ModelColumn) bool
			desc      string
		}{
			{"type:varchar", func(c *migration.ModelColumn) bool { return c.Type == migration.ColumnTypeVarchar }, "type parsing"},
			{"size:100", func(c *migration.ModelColumn) bool { return c.Length == 100 }, "size parsing"},
			{"precision:10", func(c *migration.ModelColumn) bool { return c.Precision == 10 }, "precision parsing"},
			{"scale:2", func(c *migration.ModelColumn) bool { return c.Scale == 2 }, "scale parsing"},
			{"comment:test comment", func(c *migration.ModelColumn) bool { return c.Comment == "test comment" }, "comment parsing"},
			{"default:test_value", func(c *migration.ModelColumn) bool { return c.Default != nil && *c.Default == "'test_value'" }, "default parsing"},
		}

		for _, test := range tests {
			column = &migration.ModelColumn{} // 重置
			err := analyzer.ParseTormKeyValue(test.input, column)
			if err != nil {
				t.Errorf("Failed to parse '%s': %v", test.input, err)
				continue
			}

			if !test.checkFunc(column) {
				t.Errorf("Failed %s for input '%s'", test.desc, test.input)
			}
		}
	})

	t.Run("parseTormFlag", func(t *testing.T) {
		tests := []struct {
			flag      string
			checkFunc func(*migration.ModelColumn) bool
			desc      string
		}{
			{"primary_key", func(c *migration.ModelColumn) bool { return c.PrimaryKey }, "primary key flag"},
			{"pk", func(c *migration.ModelColumn) bool { return c.PrimaryKey }, "pk flag"},
			{"auto_increment", func(c *migration.ModelColumn) bool { return c.AutoIncrement }, "auto increment flag"},
			{"unique", func(c *migration.ModelColumn) bool { return c.Unique }, "unique flag"},
			{"not_null", func(c *migration.ModelColumn) bool { return c.NotNull }, "not null flag"},
			{"nullable", func(c *migration.ModelColumn) bool { return !c.NotNull }, "nullable flag"},
			{"auto_create_time", func(c *migration.ModelColumn) bool { return c.NotNull && c.Default != nil }, "auto create time flag"},
			{"auto_update_time", func(c *migration.ModelColumn) bool { return c.NotNull && c.Default != nil }, "auto update time flag"},
		}

		for _, test := range tests {
			column := &migration.ModelColumn{}
			analyzer.ParseTormFlag(test.flag, column)

			if !test.checkFunc(column) {
				t.Errorf("Failed %s for flag '%s'", test.desc, test.flag)
			}
		}
	})

	t.Run("setColumnType", func(t *testing.T) {
		tests := []struct {
			typeStr       string
			expectedType  migration.ColumnType
			checkLength   bool
			expectedLen   int
			checkPrec     bool
			expectedPrec  int
			expectedScale int
		}{
			{"varchar", migration.ColumnTypeVarchar, true, 255, false, 0, 0},
			{"char", migration.ColumnTypeChar, true, 255, false, 0, 0},
			{"text", migration.ColumnTypeText, false, 0, false, 0, 0},
			{"int", migration.ColumnTypeInt, false, 0, false, 0, 0},
			{"bigint", migration.ColumnTypeBigInt, false, 0, false, 0, 0},
			{"decimal", migration.ColumnTypeDecimal, false, 0, true, 10, 2},
			{"boolean", migration.ColumnTypeBoolean, false, 0, false, 0, 0},
			{"json", migration.ColumnTypeJSON, false, 0, false, 0, 0},
		}

		for _, test := range tests {
			column := &migration.ModelColumn{}
			analyzer.SetColumnType(test.typeStr, column)

			if column.Type != test.expectedType {
				t.Errorf("Type '%s': expected %s, got %s", test.typeStr, test.expectedType, column.Type)
			}

			if test.checkLength && column.Length != test.expectedLen {
				t.Errorf("Type '%s': expected length %d, got %d", test.typeStr, test.expectedLen, column.Length)
			}

			if test.checkPrec && column.Precision != test.expectedPrec {
				t.Errorf("Type '%s': expected precision %d, got %d", test.typeStr, test.expectedPrec, column.Precision)
			}

			if test.checkPrec && column.Scale != test.expectedScale {
				t.Errorf("Type '%s': expected scale %d, got %d", test.typeStr, test.expectedScale, column.Scale)
			}
		}
	})

	t.Run("parseDefaultValue", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"null", "NULL"},
			{"NULL", "NULL"},
			{"current_timestamp", "CURRENT_TIMESTAMP"},
			{"CURRENT_TIMESTAMP", "CURRENT_TIMESTAMP"},
			{"now()", "CURRENT_TIMESTAMP"},
			{"true", "1"},
			{"TRUE", "1"},
			{"false", "0"},
			{"FALSE", "0"},
			{"123", "123"},
			{"123.45", "123.45"},
			{"test_string", "'test_string'"},
			{"string with spaces", "'string with spaces'"},
			{"string'with'quotes", "'string''with''quotes'"},
		}

		for _, test := range tests {
			result := analyzer.ParseDefaultValue(test.input)
			if result != test.expected {
				t.Errorf("Input '%s': expected '%s', got '%s'", test.input, test.expected, result)
			}
		}
	})

	// 注意：mapGoTypeToColumnType 和 toSnakeCase 是私有方法，我们通过公开的 AnalyzeModel 方法间接测试
	t.Run("AnalyzeModel_TypeMapping", func(t *testing.T) {
		type TypeMappingModel struct {
			StringField string
			IntField    int
			BoolField   bool
			FloatField  float64
		}

		modelType := reflect.TypeOf(TypeMappingModel{})
		columns, err := analyzer.AnalyzeModel(modelType)
		if err != nil {
			t.Fatalf("Failed to analyze type mapping model: %v", err)
		}

		// 由于没有标签，这些字段会被跳过
		if len(columns) != 0 {
			t.Errorf("Expected 0 columns without tags, got %d", len(columns))
		}
	})
}

func TestSchemaComparator_Unit(t *testing.T) {
	// 创建一个模拟的数据库连接用于测试
	mockConn := &MockConnection{driver: "sqlite"}
	comparator := migration.NewSchemaComparator(mockConn)

	t.Run("NewSchemaComparator", func(t *testing.T) {
		if comparator == nil {
			t.Error("NewSchemaComparator should return a valid instance")
		}
	})

	t.Run("CompareColumns_Basic", func(t *testing.T) {
		// 测试基本的列对比功能
		dbColumns := []migration.DatabaseColumn{
			{Name: "id", Type: "INTEGER", PrimaryKey: true},
			{Name: "name", Type: "TEXT"},
		}

		modelColumns := []migration.ModelColumn{
			{Name: "id", Type: migration.ColumnTypeInt, PrimaryKey: true},
			{Name: "name", Type: migration.ColumnTypeVarchar, Length: 100},
			{Name: "email", Type: migration.ColumnTypeVarchar, Length: 150}, // 新增字段
		}

		differences := comparator.CompareColumns(dbColumns, modelColumns)

		// 应该至少有一个差异（新增email字段）
		if len(differences) == 0 {
			t.Error("Expected differences but got none")
		}

		// 查找新增字段的差异
		var emailDiff *migration.ColumnDifference
		for _, diff := range differences {
			if diff.Column == "email" && diff.Type == "add" {
				emailDiff = &diff
				break
			}
		}

		if emailDiff == nil {
			t.Error("Expected to find email field addition difference")
		}
	})
}

func TestAlterGenerator_Unit(t *testing.T) {
	mockConn := &MockConnection{driver: "mysql"}
	generator := migration.NewAlterGenerator(mockConn)

	t.Run("NewAlterGenerator", func(t *testing.T) {
		if generator == nil {
			t.Error("NewAlterGenerator should return a valid instance")
		}
	})

	t.Run("GenerateAlterSQL_Basic", func(t *testing.T) {
		differences := []migration.ColumnDifference{
			{
				Column: "new_field",
				Type:   "add",
				NewValue: migration.ModelColumn{
					Name:   "new_field",
					Type:   migration.ColumnTypeVarchar,
					Length: 100,
				},
			},
		}

		statements, err := generator.GenerateAlterSQL("test_table", differences)
		if err != nil {
			t.Fatalf("Failed to generate ALTER SQL: %v", err)
		}

		if len(statements) == 0 {
			t.Error("Expected ALTER statements to be generated")
		}

		// MySQL 应该生成包含 ADD COLUMN 的语句
		if len(statements) > 0 {
			sql := statements[0]
			if !contains(sql, "ADD COLUMN") || !contains(sql, "new_field") {
				t.Errorf("Expected ADD COLUMN statement, got: %s", sql)
			}
		}
	})
}

// MockConnection 模拟数据库连接，实现 db.ConnectionInterface
type MockConnection struct {
	driver string
}

func (m *MockConnection) GetDriver() string {
	return m.driver
}

func (m *MockConnection) Connect() error {
	return nil
}

func (m *MockConnection) Close() error {
	return nil
}

func (m *MockConnection) Ping() error {
	return nil
}

func (m *MockConnection) IsConnected() bool {
	return true
}

func (m *MockConnection) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return nil, fmt.Errorf("mock query not implemented")
}

func (m *MockConnection) QueryRow(query string, args ...interface{}) *sql.Row {
	return nil
}

func (m *MockConnection) Exec(query string, args ...interface{}) (sql.Result, error) {
	return nil, fmt.Errorf("mock exec not implemented")
}

func (m *MockConnection) Begin() (db.TransactionInterface, error) {
	return nil, fmt.Errorf("mock begin not implemented")
}

func (m *MockConnection) BeginTx(opts *sql.TxOptions) (db.TransactionInterface, error) {
	return nil, fmt.Errorf("mock begin not implemented")
}

func (m *MockConnection) GetConfig() *db.Config {
	return &db.Config{Driver: m.driver}
}

func (m *MockConnection) GetStats() sql.DBStats {
	return sql.DBStats{}
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}

func TestSafeMigrator_Unit(t *testing.T) {
	t.Run("NewSafeMigrator", func(t *testing.T) {
		mockConn := &MockConnection{driver: "sqlite"}
		migrator := migration.NewSafeMigrator(mockConn)

		if migrator == nil {
			t.Error("NewSafeMigrator should return a valid instance")
		}

		// 测试链式配置
		migrator = migrator.SetDryRun(true).SetBackupTables(false)
		if migrator == nil {
			t.Error("Chained configuration should return valid instance")
		}
	})
}

func TestMigrationResult_Unit(t *testing.T) {
	t.Run("PrintSummary_Success", func(t *testing.T) {
		result := &migration.MigrationResult{
			TableName: "test_table",
			Success:   true,
			Message:   "Test success message",
			Changes: []migration.ColumnDifference{
				{Column: "field1", Type: "add"},
				{Column: "field2", Type: "modify"},
			},
		}

		// 主要是测试不会panic，实际输出在控制台
		result.PrintSummary()
	})

	t.Run("PrintSummary_Failure", func(t *testing.T) {
		result := &migration.MigrationResult{
			TableName:            "test_table",
			Success:              false,
			Error:                fmt.Errorf("test error"),
			FailedStatement:      "ALTER TABLE test ADD COLUMN invalid",
			RecoveryInstructions: "Run recovery command",
		}

		// 主要是测试不会panic
		result.PrintSummary()
	})
}
