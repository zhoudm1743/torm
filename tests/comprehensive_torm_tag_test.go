package tests

import (
	"reflect"
	"testing"

	"github.com/zhoudm1743/torm/migration"
	"github.com/zhoudm1743/torm/model"
)

// TestModelWithAllTormTags 包含所有torm标签的测试模型
type TestModelWithAllTormTags struct {
	model.BaseModel
	// 主键字段
	ID string `json:"id" torm:"primary_key,type:varchar,size:32,comment:主键ID"`

	// 基础字段类型测试
	Name        string  `json:"name" torm:"type:varchar,size:100,not_null,comment:姓名"`
	Email       string  `json:"email" torm:"type:varchar,size:255,unique,comment:邮箱"`
	Age         int     `json:"age" torm:"type:int,default:18,comment:年龄"`
	Salary      float64 `json:"salary" torm:"type:decimal,precision:10,scale:2,comment:薪水"`
	IsActive    bool    `json:"is_active" torm:"type:boolean,default:true,comment:是否激活"`
	Description string  `json:"description" torm:"type:text,nullable,comment:描述"`

	// 自动时间戳字段
	CreatedAt int64 `json:"created_at" torm:"auto_create_time,comment:创建时间"`
	UpdatedAt int64 `json:"updated_at" torm:"auto_update_time,comment:更新时间"`

	// 自增字段
	SerialNumber int64 `json:"serial_number" torm:"type:bigint,auto_increment,comment:序列号"`

	// JSON字段
	Metadata map[string]interface{} `json:"metadata" torm:"type:json,comment:元数据"`
	Tags     []string               `json:"tags" torm:"type:varchar,size:500,comment:标签列表"`
}

func TestComprehensiveTormTagParsing(t *testing.T) {
	testModel := &TestModelWithAllTormTags{}
	testModel.BaseModel = *model.NewBaseModelWithAutoDetect(testModel)

	analyzer := migration.NewModelAnalyzer()
	modelType := reflect.TypeOf(*testModel)

	columns, err := analyzer.AnalyzeModel(modelType)
	if err != nil {
		t.Fatalf("Failed to analyze model: %v", err)
	}

	t.Logf("Found %d columns", len(columns))

	// 创建字段映射便于测试
	columnMap := make(map[string]migration.ModelColumn)
	for _, col := range columns {
		columnMap[col.Name] = col
		t.Logf("Column: %s, Type: %s, Length: %d, NotNull: %v, Default: %v, Comment: %s",
			col.Name, col.Type, col.Length, col.NotNull, col.Default, col.Comment)
	}

	// 测试各种字段属性
	testCases := []struct {
		fieldName       string
		expectedType    migration.ColumnType
		expectedLen     int
		expectedNotNull bool
		expectedPK      bool
		expectedUnique  bool
		expectedAutoInc bool
		expectedComment string
		hasDefault      bool
	}{
		{
			fieldName:       "i_d", // 转换为snake_case
			expectedType:    migration.ColumnTypeVarchar,
			expectedLen:     32,
			expectedNotNull: false,
			expectedPK:      true,
			expectedUnique:  false,
			expectedAutoInc: false,
			expectedComment: "主键ID",
			hasDefault:      false,
		},
		{
			fieldName:       "name",
			expectedType:    migration.ColumnTypeVarchar,
			expectedLen:     100,
			expectedNotNull: true,
			expectedPK:      false,
			expectedUnique:  false,
			expectedAutoInc: false,
			expectedComment: "姓名",
			hasDefault:      false,
		},
		{
			fieldName:       "email",
			expectedType:    migration.ColumnTypeVarchar,
			expectedLen:     255,
			expectedNotNull: false,
			expectedPK:      false,
			expectedUnique:  true,
			expectedAutoInc: false,
			expectedComment: "邮箱",
			hasDefault:      false,
		},
		{
			fieldName:       "age",
			expectedType:    migration.ColumnTypeInt,
			expectedLen:     0,
			expectedNotNull: false,
			expectedPK:      false,
			expectedUnique:  false,
			expectedAutoInc: false,
			expectedComment: "年龄",
			hasDefault:      true,
		},
		{
			fieldName:       "salary",
			expectedType:    migration.ColumnTypeDecimal,
			expectedLen:     0, // decimal uses precision/scale
			expectedNotNull: false,
			expectedPK:      false,
			expectedUnique:  false,
			expectedAutoInc: false,
			expectedComment: "薪水",
			hasDefault:      false,
		},
		{
			fieldName:       "is_active",
			expectedType:    migration.ColumnTypeBoolean,
			expectedLen:     0,
			expectedNotNull: false,
			expectedPK:      false,
			expectedUnique:  false,
			expectedAutoInc: false,
			expectedComment: "是否激活",
			hasDefault:      true,
		},
		{
			fieldName:       "description",
			expectedType:    migration.ColumnTypeText,
			expectedLen:     0,
			expectedNotNull: false, // nullable explicitly set
			expectedPK:      false,
			expectedUnique:  false,
			expectedAutoInc: false,
			expectedComment: "描述",
			hasDefault:      false,
		},
		{
			fieldName:       "serial_number",
			expectedType:    migration.ColumnTypeBigInt,
			expectedLen:     0,
			expectedNotNull: false,
			expectedPK:      false,
			expectedUnique:  false,
			expectedAutoInc: true,
			expectedComment: "序列号",
			hasDefault:      false,
		},
	}

	for _, tc := range testCases {
		t.Run("Test_"+tc.fieldName, func(t *testing.T) {
			col, exists := columnMap[tc.fieldName]
			if !exists {
				t.Errorf("Column %s not found", tc.fieldName)
				return
			}

			// 测试类型
			if col.Type != tc.expectedType {
				t.Errorf("Column %s: expected type %s, got %s", tc.fieldName, tc.expectedType, col.Type)
			}

			// 测试长度
			if tc.expectedLen > 0 && col.Length != tc.expectedLen {
				t.Errorf("Column %s: expected length %d, got %d", tc.fieldName, tc.expectedLen, col.Length)
			}

			// 测试NOT NULL
			if col.NotNull != tc.expectedNotNull {
				t.Errorf("Column %s: expected NotNull %v, got %v", tc.fieldName, tc.expectedNotNull, col.NotNull)
			}

			// 测试主键
			if col.PrimaryKey != tc.expectedPK {
				t.Errorf("Column %s: expected PrimaryKey %v, got %v", tc.fieldName, tc.expectedPK, col.PrimaryKey)
			}

			// 测试唯一性
			if col.Unique != tc.expectedUnique {
				t.Errorf("Column %s: expected Unique %v, got %v", tc.fieldName, tc.expectedUnique, col.Unique)
			}

			// 测试自增
			if col.AutoIncrement != tc.expectedAutoInc {
				t.Errorf("Column %s: expected AutoIncrement %v, got %v", tc.fieldName, tc.expectedAutoInc, col.AutoIncrement)
			}

			// 测试注释
			if col.Comment != tc.expectedComment {
				t.Errorf("Column %s: expected comment '%s', got '%s'", tc.fieldName, tc.expectedComment, col.Comment)
			}

			// 测试默认值
			hasDefault := col.Default != nil
			if hasDefault != tc.hasDefault {
				t.Errorf("Column %s: expected hasDefault %v, got %v", tc.fieldName, tc.hasDefault, hasDefault)
			}

			if hasDefault {
				t.Logf("Column %s has default value: %v", tc.fieldName, *col.Default)
			}
		})
	}

	// 特别测试decimal字段的precision和scale
	t.Run("Test_decimal_precision_scale", func(t *testing.T) {
		col, exists := columnMap["salary"]
		if !exists {
			t.Error("Salary column not found")
			return
		}

		if col.Precision != 10 {
			t.Errorf("Salary: expected precision 10, got %d", col.Precision)
		}

		if col.Scale != 2 {
			t.Errorf("Salary: expected scale 2, got %d", col.Scale)
		}
	})

	// 测试auto_create_time和auto_update_time字段
	t.Run("Test_auto_timestamp_fields", func(t *testing.T) {
		createdAtCol, exists := columnMap["created_at"]
		if !exists {
			t.Error("created_at column not found")
			return
		}

		updatedAtCol, exists := columnMap["updated_at"]
		if !exists {
			t.Error("updated_at column not found")
			return
		}

		// 检查created_at字段
		if !createdAtCol.NotNull {
			t.Error("created_at should be NOT NULL")
		}
		if createdAtCol.Default == nil {
			t.Error("created_at should have default value")
		} else {
			t.Logf("created_at default: %s", *createdAtCol.Default)
		}

		// 检查updated_at字段
		if !updatedAtCol.NotNull {
			t.Error("updated_at should be NOT NULL")
		}
		if updatedAtCol.Default == nil {
			t.Error("updated_at should have default value")
		} else {
			t.Logf("updated_at default: %s", *updatedAtCol.Default)
		}
	})
}

// TestMissingTormTagSupport 测试缺失的torm标签支持
func TestMissingTormTagSupport(t *testing.T) {
	t.Log("检查当前torm标签解析器支持的功能...")

	// 检查ParseTormKeyValue中支持的键
	supportedKeys := []string{"type", "size", "precision", "scale", "default", "comment"}
	t.Logf("支持的键值对标签: %v", supportedKeys)

	// 检查ParseTormFlag中支持的标志
	supportedFlags := []string{"primary_key", "pk", "auto_increment", "unique", "not_null", "nullable", "auto_create_time", "auto_update_time"}
	t.Logf("支持的标志标签: %v", supportedFlags)

	// 可能缺失的标签支持
	potentialMissingTags := []string{
		"index",                 // 普通索引
		"foreign_key",           // 外键
		"column",                // 自定义列名
		"charset",               // 字符集
		"collation",             // 排序规则
		"on_update",             // 更新时动作
		"on_delete",             // 删除时动作
		"check",                 // 检查约束
		"virtual",               // 虚拟列
		"stored",                // 存储列
		"unsigned",              // 无符号
		"zerofill",              // 零填充
		"binary",                // 二进制
		"auto_create_time_nano", // 纳秒级创建时间
		"auto_update_time_nano", // 纳秒级更新时间
	}

	t.Logf("可能需要支持的标签: %v", potentialMissingTags)
}
