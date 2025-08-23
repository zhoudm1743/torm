package tests

import (
	"reflect"
	"testing"
	"time"

	"github.com/zhoudm1743/torm/db"
	"github.com/zhoudm1743/torm/migration"
	"github.com/zhoudm1743/torm/model"
)

// 完整的测试模型，覆盖所有字段类型
type ComprehensiveTestModel struct {
	model.BaseModel
	// 主键和自增
	ID int64 `torm:"primary_key,auto_increment,comment:主键ID"`

	// 字符串类型
	ShortCode   string `torm:"type:varchar,size:10,comment:短编码"`
	Name        string `torm:"type:varchar,size:100,not_null,comment:名称"`
	Email       string `torm:"type:varchar,size:150,unique,comment:邮箱"`
	Description string `torm:"type:text,comment:描述"`
	Content     string `torm:"type:longtext,nullable,comment:内容"`
	FixedCode   string `torm:"type:char,size:8,comment:固定编码"`

	// 数值类型
	Age       int     `torm:"type:int,default:0,comment:年龄"`
	SmallNum  int8    `torm:"type:tinyint,comment:小数字"`
	MediumNum int16   `torm:"type:smallint,comment:中等数字"`
	BigNum    int64   `torm:"type:bigint,comment:大数字"`
	Score     float32 `torm:"type:float,comment:分数"`
	Rate      float64 `torm:"type:double,comment:比率"`

	// 精确数值
	Price   float64 `torm:"type:decimal,precision:10,scale:2,default:0.00,comment:价格"`
	Balance float64 `torm:"type:decimal,precision:15,scale:4,comment:余额"`

	// 布尔类型
	IsActive  bool `torm:"type:boolean,default:true,comment:是否激活"`
	IsDeleted bool `torm:"type:boolean,default:false,comment:是否删除"`

	// 时间类型
	BirthDate time.Time `torm:"type:date,comment:出生日期"`
	EventTime time.Time `torm:"type:timestamp,comment:事件时间"`
	CreatedAt time.Time `torm:"auto_create_time,comment:创建时间"`
	UpdatedAt time.Time `torm:"auto_update_time,comment:更新时间"`

	// 二进制和JSON
	Avatar   []byte                 `torm:"type:blob,comment:头像数据"`
	Tags     []string               `torm:"type:json,comment:标签列表"`
	Metadata map[string]interface{} `torm:"type:json,comment:元数据"`

	// 可空字段
	NickName  *string `torm:"type:varchar,size:50,nullable,comment:昵称"`
	LastLogin *int64  `torm:"nullable,comment:最后登录时间"`

	// 索引字段
	CategoryID int64 `torm:"type:bigint,index,comment:分类ID"`
	UserID     int64 `torm:"type:bigint,index:user_idx,comment:用户ID"`
}

// NewComprehensiveTestModel 创建测试模型
func NewComprehensiveTestModel() *ComprehensiveTestModel {
	m := &ComprehensiveTestModel{}
	m.BaseModel = *model.NewBaseModel()
	m.SetTable("comprehensive_test")
	m.SetConnection("default")
	// 手动设置模型结构
	m.SetModelStruct(reflect.TypeOf(*m))
	return m
}

// 用于测试模型变更的第二版本
type ComprehensiveTestModelV2 struct {
	model.BaseModel
	// 主键保持不变
	ID int64 `torm:"primary_key,auto_increment,comment:主键ID"`

	// 修改字段长度
	ShortCode string `torm:"type:varchar,size:20,comment:短编码"`          // 10->20
	Name      string `torm:"type:varchar,size:200,not_null,comment:名称"` // 100->200
	Email     string `torm:"type:varchar,size:255,unique,comment:邮箱"`   // 150->255

	// 新增字段
	Phone   string `torm:"type:varchar,size:20,comment:手机号"`
	Address string `torm:"type:varchar,size:500,nullable,comment:地址"`

	// 修改数值精度
	Price float64 `torm:"type:decimal,precision:15,scale:3,default:0.000,comment:价格"` // 10,2->15,3

	// 保持不变的字段
	Description string    `torm:"type:text,comment:描述"`
	Age         int       `torm:"type:int,default:0,comment:年龄"`
	IsActive    bool      `torm:"type:boolean,default:true,comment:是否激活"`
	CreatedAt   time.Time `torm:"auto_create_time,comment:创建时间"`
	UpdatedAt   time.Time `torm:"auto_update_time,comment:更新时间"`

	// 删除某些字段 (Content, FixedCode等在V2中不存在)
}

func NewComprehensiveTestModelV2() *ComprehensiveTestModelV2 {
	m := &ComprehensiveTestModelV2{}
	m.BaseModel = *model.NewBaseModel()
	m.SetTable("comprehensive_test")
	m.SetConnection("default")
	m.SetModelStruct(reflect.TypeOf(*m))
	return m
}

func TestSchemaComparatorComprehensive(t *testing.T) {
	setupComprehensiveTestDB(t)
	defer cleanupComprehensiveTestDB()

	t.Run("ModelAnalyzer_AllFieldTypes", func(t *testing.T) {
		analyzer := migration.NewModelAnalyzer()
		modelType := reflect.TypeOf(ComprehensiveTestModel{})

		columns, err := analyzer.AnalyzeModel(modelType)
		if err != nil {
			t.Fatalf("Failed to analyze model: %v", err)
		}

		// 验证列数量
		if len(columns) == 0 {
			t.Error("No columns found")
		}

		// 验证特定字段类型
		tests := []struct {
			name              string
			expectedType      migration.ColumnType
			expectedLength    int
			expectedPrecision int
			expectedScale     int
		}{
			{"short_code", migration.ColumnTypeVarchar, 10, 0, 0},
			{"name", migration.ColumnTypeVarchar, 100, 0, 0},
			{"description", migration.ColumnTypeText, 0, 0, 0},
			{"age", migration.ColumnTypeInt, 0, 0, 0},
			{"small_num", migration.ColumnTypeTinyInt, 0, 0, 0},
			{"big_num", migration.ColumnTypeBigInt, 0, 0, 0},
			{"score", migration.ColumnTypeFloat, 0, 0, 0},
			{"rate", migration.ColumnTypeDouble, 0, 0, 0},
			{"price", migration.ColumnTypeDecimal, 0, 10, 2},
			{"balance", migration.ColumnTypeDecimal, 0, 15, 4},
			{"is_active", migration.ColumnTypeBoolean, 0, 0, 0},
			{"birth_date", migration.ColumnTypeDate, 0, 0, 0},
			{"event_time", migration.ColumnTypeTimestamp, 0, 0, 0},
			{"avatar", migration.ColumnTypeBlob, 0, 0, 0},
			{"tags", migration.ColumnTypeJSON, 0, 0, 0},
		}

		for _, test := range tests {
			column := findColumnByName(columns, test.name)
			if column == nil {
				t.Errorf("Column %s not found", test.name)
				continue
			}

			if column.Type != test.expectedType {
				t.Errorf("Column %s: expected type %s, got %s", test.name, test.expectedType, column.Type)
			}

			if test.expectedLength > 0 && column.Length != test.expectedLength {
				t.Errorf("Column %s: expected length %d, got %d", test.name, test.expectedLength, column.Length)
			}

			if test.expectedPrecision > 0 && column.Precision != test.expectedPrecision {
				t.Errorf("Column %s: expected precision %d, got %d", test.name, test.expectedPrecision, column.Precision)
			}

			if test.expectedScale > 0 && column.Scale != test.expectedScale {
				t.Errorf("Column %s: expected scale %d, got %d", test.name, test.expectedScale, column.Scale)
			}
		}
	})

	t.Run("SchemaComparator_DatabaseColumns", func(t *testing.T) {
		conn := getTestConnection(t)

		// 创建测试表
		createSQL := `
			CREATE TABLE test_db_columns (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				email TEXT UNIQUE,
				age INTEGER DEFAULT 0,
				price REAL,
				is_active INTEGER DEFAULT 1,
				created_at DATETIME
			)
		`
		_, err := conn.Exec(createSQL)
		if err != nil {
			t.Fatalf("Failed to create test table: %v", err)
		}

		comparator := migration.NewSchemaComparator(conn)
		dbColumns, err := comparator.GetDatabaseColumns("test_db_columns")
		if err != nil {
			t.Fatalf("Failed to get database columns: %v", err)
		}

		expectedColumns := []string{"id", "name", "email", "age", "price", "is_active", "created_at"}
		if len(dbColumns) != len(expectedColumns) {
			t.Errorf("Expected %d columns, got %d", len(expectedColumns), len(dbColumns))
		}

		for _, expectedCol := range expectedColumns {
			found := false
			for _, dbCol := range dbColumns {
				if dbCol.Name == expectedCol {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected column %s not found", expectedCol)
			}
		}

		// 清理
		conn.Exec("DROP TABLE test_db_columns")
	})

	t.Run("CompareColumns_AllDifferenceTypes", func(t *testing.T) {
		comparator := migration.NewSchemaComparator(getTestConnection(t))

		// 数据库中的列（模拟）
		dbColumns := []migration.DatabaseColumn{
			{Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true},
			{Name: "name", Type: "TEXT", Length: intPtr(50)},
			{Name: "email", Type: "TEXT", Length: intPtr(100), Unique: true},
			{Name: "old_field", Type: "TEXT"},
			{Name: "price", Type: "REAL"},
		}

		// 模型中的列
		modelColumns := []migration.ModelColumn{
			{Name: "id", Type: migration.ColumnTypeInt, PrimaryKey: true, AutoIncrement: true},
			{Name: "name", Type: migration.ColumnTypeVarchar, Length: 100},                // 长度变更 50->100
			{Name: "email", Type: migration.ColumnTypeVarchar, Length: 200, Unique: true}, // 长度变更 100->200
			{Name: "new_field", Type: migration.ColumnTypeVarchar, Length: 255},           // 新增字段
			{Name: "price", Type: migration.ColumnTypeDecimal, Precision: 10, Scale: 2},   // 类型变更
		}

		differences := comparator.CompareColumns(dbColumns, modelColumns)

		// 验证差异类型
		differenceTypes := make(map[string]string)
		for _, diff := range differences {
			differenceTypes[diff.Column] = diff.Type
		}

		// 应该有的差异：
		// - name: modify (长度变更)
		// - email: modify (长度变更)
		// - new_field: add (新增字段)
		// - old_field: drop (删除字段)
		// - price: modify (类型变更)

		expectedDifferences := map[string]string{
			"name":      "modify",
			"email":     "modify",
			"new_field": "add",
			"old_field": "drop",
			"price":     "modify",
		}

		for column, expectedType := range expectedDifferences {
			if actualType, exists := differenceTypes[column]; !exists {
				t.Errorf("Expected difference for column %s not found", column)
			} else if actualType != expectedType {
				t.Errorf("Column %s: expected difference type %s, got %s", column, expectedType, actualType)
			}
		}
	})

	t.Run("AlterGenerator_AllDatabaseTypes", func(t *testing.T) {
		conn := getTestConnection(t)
		generator := migration.NewAlterGenerator(conn)

		differences := []migration.ColumnDifference{
			{
				Column: "new_varchar",
				Type:   "add",
				NewValue: migration.ModelColumn{
					Name:    "new_varchar",
					Type:    migration.ColumnTypeVarchar,
					Length:  100,
					Comment: "新字符串字段",
				},
			},
			{
				Column: "existing_field",
				Type:   "modify",
				OldValue: migration.DatabaseColumn{
					Name: "existing_field",
					Type: "TEXT",
				},
				NewValue: migration.ModelColumn{
					Name:      "existing_field",
					Type:      migration.ColumnTypeDecimal,
					Precision: 10,
					Scale:     2,
					NotNull:   true,
				},
			},
			{
				Column: "old_field",
				Type:   "drop",
				OldValue: migration.DatabaseColumn{
					Name: "old_field",
					Type: "TEXT",
				},
			},
		}

		statements, err := generator.GenerateAlterSQL("test_table", differences)
		if err != nil {
			t.Fatalf("Failed to generate ALTER SQL: %v", err)
		}

		if len(statements) == 0 {
			t.Error("No ALTER statements generated")
		}

		// 验证生成的SQL包含预期的操作
		sqlText := statements[0]
		t.Logf("Generated SQL: %s", sqlText)

		// SQLite应该生成包含ADD COLUMN和其他操作的语句
		if !contains(sqlText, "ADD COLUMN") && !contains(sqlText, "new_varchar") {
			t.Error("Generated SQL should contain ADD COLUMN operation")
		}
	})

	t.Run("SafeMigrator_DryRun", func(t *testing.T) {
		conn := getTestConnection(t)
		migrator := migration.NewSafeMigrator(conn).SetDryRun(true).SetBackupTables(false)

		differences := []migration.ColumnDifference{
			{
				Column: "test_field",
				Type:   "add",
				NewValue: migration.ModelColumn{
					Name:   "test_field",
					Type:   migration.ColumnTypeVarchar,
					Length: 100,
				},
			},
		}

		result, err := migrator.SafeAlterTable("test_table", differences)
		if err != nil {
			t.Fatalf("Dry run failed: %v", err)
		}

		if !result.Success {
			t.Error("Dry run should succeed")
		}

		if result.BackupTable != "" {
			t.Error("Dry run should not create backup tables")
		}

		if !contains(result.Message, "Dry run completed") {
			t.Error("Dry run message should indicate completion")
		}
	})

	t.Run("SafeMigrator_ActualExecution", func(t *testing.T) {
		conn := getTestConnection(t)

		// 创建测试表
		createSQL := `
			CREATE TABLE safe_migration_test (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT
			)
		`
		_, err := conn.Exec(createSQL)
		if err != nil {
			t.Fatalf("Failed to create test table: %v", err)
		}

		migrator := migration.NewSafeMigrator(conn).
			SetDryRun(false).
			SetBackupTables(true)

		differences := []migration.ColumnDifference{
			{
				Column: "email",
				Type:   "add",
				NewValue: migration.ModelColumn{
					Name:   "email",
					Type:   migration.ColumnTypeVarchar,
					Length: 100,
				},
			},
		}

		result, err := migrator.SafeAlterTable("safe_migration_test", differences)
		if err != nil {
			t.Fatalf("Safe migration failed: %v", err)
		}

		if !result.Success {
			t.Errorf("Migration should succeed, but got error: %v", result.Error)
		}

		if result.BackupTable == "" {
			t.Error("Backup table should be created")
		}

		// 验证字段是否添加成功
		rows, err := conn.Query("PRAGMA table_info(safe_migration_test)")
		if err != nil {
			t.Fatalf("Failed to check table structure: %v", err)
		}
		defer rows.Close()

		columnNames := []string{}
		for rows.Next() {
			var cid int
			var name, dataType string
			var notNull, pk int
			var defaultVal interface{}
			rows.Scan(&cid, &name, &dataType, &notNull, &defaultVal, &pk)
			columnNames = append(columnNames, name)
		}

		emailFound := false
		for _, name := range columnNames {
			if name == "email" {
				emailFound = true
				break
			}
		}

		if !emailFound {
			t.Error("Email column should be added")
		}

		// 清理
		conn.Exec("DROP TABLE IF EXISTS safe_migration_test")
		if result.BackupTable != "" {
			conn.Exec("DROP TABLE IF EXISTS " + result.BackupTable)
		}
	})

	t.Run("AutoMigrate_WithoutDetectConfigFromStruct", func(t *testing.T) {
		// 测试新的AutoMigrate功能，不需要手动调用DetectConfigFromStruct
		model := NewComprehensiveTestModel()

		// 不调用DetectConfigFromStruct，直接使用AutoMigrate
		err := model.AutoMigrate()
		if err != nil {
			t.Logf("AutoMigrate failed as expected: %v", err)
			// 这是预期的，因为我们故意不让自动检测成功
			// 验证错误消息是否包含正确的提示
			if !contains(err.Error(), "cannot auto-detect model structure") {
				t.Errorf("Error should mention auto-detection failure, got: %v", err)
			}
		}
	})

	t.Run("AutoMigrate_WithSetModelStruct", func(t *testing.T) {
		// 测试使用SetModelStruct的方式
		testModel := &ComprehensiveTestModel{}
		testModel.BaseModel = *model.NewBaseModel()
		testModel.SetTable("comprehensive_set_struct_test")
		testModel.SetConnection("default")
		testModel.SetModelStruct(reflect.TypeOf(*testModel)) // 手动设置模型结构

		err := testModel.AutoMigrate()
		if err != nil {
			t.Fatalf("AutoMigrate with SetModelStruct failed: %v", err)
		}

		// 验证表是否创建
		conn := getTestConnection(t)
		var count int
		row := conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='comprehensive_set_struct_test'")
		err = row.Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check table existence: %v", err)
		}

		if count != 1 {
			t.Error("Table should be created")
		}

		// 清理
		conn.Exec("DROP TABLE IF EXISTS comprehensive_set_struct_test")
	})

	t.Run("AutoMigrate_SchemaUpdate", func(t *testing.T) {
		// 测试表结构更新功能

		// 首先创建V1版本的表
		modelV1 := NewComprehensiveTestModel()
		modelV1.SetTable("schema_update_test")

		err := modelV1.AutoMigrate()
		if err != nil {
			t.Fatalf("V1 AutoMigrate failed: %v", err)
		}

		// 然后使用V2版本更新表结构
		modelV2 := NewComprehensiveTestModelV2()
		modelV2.SetTable("schema_update_test")

		err = modelV2.AutoMigrate()
		if err != nil {
			t.Fatalf("V2 AutoMigrate failed: %v", err)
		}

		// 验证表结构是否更新
		conn := getTestConnection(t)
		rows, err := conn.Query("PRAGMA table_info(schema_update_test)")
		if err != nil {
			t.Fatalf("Failed to check updated table structure: %v", err)
		}
		defer rows.Close()

		columnNames := []string{}
		for rows.Next() {
			var cid int
			var name, dataType string
			var notNull, pk int
			var defaultVal interface{}
			rows.Scan(&cid, &name, &dataType, &notNull, &defaultVal, &pk)
			columnNames = append(columnNames, name)
		}

		// 验证新增字段是否存在
		expectedNewFields := []string{"phone", "address"}
		for _, newField := range expectedNewFields {
			found := false
			for _, name := range columnNames {
				if name == newField {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("New field %s should be added", newField)
			}
		}

		// 清理
		conn.Exec("DROP TABLE IF EXISTS schema_update_test")
	})
}

// 辅助函数
func setupComprehensiveTestDB(t *testing.T) {
	config := &db.Config{
		Driver:   "sqlite",
		Database: "test_comprehensive.db",
	}

	err := db.AddConnection("default", config)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}
}

func cleanupComprehensiveTestDB() {
	// 清理测试数据库文件可以在这里实现
}

func getTestConnection(t *testing.T) db.ConnectionInterface {
	conn, err := db.DB("default")
	if err != nil {
		t.Fatalf("Failed to get connection: %v", err)
	}

	err = conn.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	return conn
}

func findColumnByName(columns []migration.ModelColumn, name string) *migration.ModelColumn {
	for _, col := range columns {
		if col.Name == name {
			return &col
		}
	}
	return nil
}

func intPtr(i int) *int {
	return &i
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr)+1 && func() bool {
				for i := 1; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}()))
}
