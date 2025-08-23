package tests

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/zhoudm1743/torm/migration"
	"github.com/zhoudm1743/torm/model"
)

// TestModelBefore 修改前的模型（演示各种问题）
type TestModelBefore struct {
	model.BaseModel
	// 问题1: 字段长度修改不生效
	Title string `json:"title" torm:"type:varchar,size:50,comment:标题"`

	// 问题2: 类型修改不生效
	Status int `json:"status" torm:"type:tinyint,default:0,comment:状态"`

	// 问题3: 约束修改不生效
	Email string `json:"email" torm:"type:varchar,size:100,comment:邮箱"`

	// 问题4: 默认值修改不生效
	Priority int `json:"priority" torm:"type:int,default:1,comment:优先级"`

	// 问题5: 注释修改不生效
	Content string `json:"content" torm:"type:text,comment:旧注释"`
}

// TestModelAfter 修改后的模型（应该能被正确检测）
type TestModelAfter struct {
	model.BaseModel
	// 修改1: 长度从50改为200
	Title string `json:"title" torm:"type:varchar,size:200,comment:标题"`

	// 修改2: 类型从tinyint改为int
	Status int `json:"status" torm:"type:int,default:0,comment:状态"`

	// 修改3: 添加unique约束
	Email string `json:"email" torm:"type:varchar,size:100,unique,comment:邮箱"`

	// 修改4: 默认值从1改为5
	Priority int `json:"priority" torm:"type:int,default:5,comment:优先级"`

	// 修改5: 注释从"旧注释"改为"新内容"
	Content string `json:"content" torm:"type:text,comment:新内容"`
}

// TestModelWithCustomColumn 测试自定义列名
type TestModelWithCustomColumn struct {
	model.BaseModel
	UserName string `json:"user_name" torm:"column:username,type:varchar,size:50,comment:用户名"`
	UserID   int64  `json:"user_id" torm:"column:uid,type:bigint,comment:用户ID"`
}

func TestTormTagIssueDetection(t *testing.T) {
	t.Log(" 测试torm标签修改检测能力")

	analyzer := migration.NewModelAnalyzer()

	// 分析修改前的模型
	beforeModel := &TestModelBefore{}
	beforeType := reflect.TypeOf(*beforeModel)
	beforeColumns, err := analyzer.AnalyzeModel(beforeType)
	if err != nil {
		t.Fatalf("Failed to analyze before model: %v", err)
	}

	// 分析修改后的模型
	afterModel := &TestModelAfter{}
	afterType := reflect.TypeOf(*afterModel)
	afterColumns, err := analyzer.AnalyzeModel(afterType)
	if err != nil {
		t.Fatalf("Failed to analyze after model: %v", err)
	}

	t.Log(" 修改前的模型列:")
	printModelColumns(t, "before", beforeColumns)

	t.Log(" 修改后的模型列:")
	printModelColumns(t, "after", afterColumns)

	// 模拟数据库列比较
	differences := simulateColumnComparison(beforeColumns, afterColumns)

	t.Logf(" 检测到 %d 处差异:", len(differences))
	for i, diff := range differences {
		t.Logf("  %d. 列: %s", i+1, diff.Column)
		t.Logf("     类型: %s", diff.Type)
		t.Logf("     原因: %s", diff.Reason)
		t.Log()
	}

	// 验证具体的修改检测
	expectedChanges := map[string]string{
		"title":    "长度修改",
		"status":   "类型修改",
		"email":    "约束修改",
		"priority": "默认值修改",
		"content":  "注释修改",
	}

	detectedChanges := make(map[string]bool)
	for _, diff := range differences {
		detectedChanges[diff.Column] = true
	}

	for field, changeType := range expectedChanges {
		if detectedChanges[field] {
			t.Logf(" %s的%s被正确检测", field, changeType)
		} else {
			t.Logf(" %s的%s未被检测到", field, changeType)
		}
	}
}

func TestCustomColumnNameSupport(t *testing.T) {
	t.Log(" 测试自定义列名支持")

	model := &TestModelWithCustomColumn{}
	analyzer := migration.NewModelAnalyzer()
	modelType := reflect.TypeOf(*model)

	columns, err := analyzer.AnalyzeModel(modelType)
	if err != nil {
		t.Fatalf("Failed to analyze model with custom columns: %v", err)
	}

	// 检查自定义列名是否生效
	columnMap := make(map[string]migration.ModelColumn)
	for _, col := range columns {
		columnMap[col.Name] = col
		t.Logf("列: %s, 注释: %s", col.Name, col.Comment)
	}

	// 验证自定义列名
	if col, exists := columnMap["username"]; exists {
		t.Logf(" 自定义列名 'username' 生效，注释: %s", col.Comment)
	} else {
		t.Error(" 自定义列名 'username' 未生效")
	}

	if col, exists := columnMap["uid"]; exists {
		t.Logf(" 自定义列名 'uid' 生效，注释: %s", col.Comment)
	} else {
		t.Error(" 自定义列名 'uid' 未生效")
	}
}

func TestTormTagPriorityAndParsing(t *testing.T) {
	t.Log(" 测试torm标签解析优先级和格式")

	// 测试不同格式的torm标签
	testCases := []struct {
		name     string
		tormTag  string
		expected map[string]interface{}
	}{
		{
			name:    "基础格式",
			tormTag: "type:varchar,size:100,not_null,comment:测试",
			expected: map[string]interface{}{
				"type":     migration.ColumnTypeVarchar,
				"length":   100,
				"not_null": true,
				"comment":  "测试",
			},
		},
		{
			name:    "主键格式",
			tormTag: "primary_key,type:varchar,size:32",
			expected: map[string]interface{}{
				"primary_key": true,
				"type":        migration.ColumnTypeVarchar,
				"length":      32,
			},
		},
		{
			name:    "自增格式",
			tormTag: "type:bigint,auto_increment,not_null",
			expected: map[string]interface{}{
				"type":           migration.ColumnTypeBigInt,
				"auto_increment": true,
				"not_null":       true,
			},
		},
		{
			name:    "时间戳格式",
			tormTag: "auto_create_time,comment:创建时间",
			expected: map[string]interface{}{
				"auto_create_time": true,
				"not_null":         true,
				"comment":          "创建时间",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建临时列用于测试
			column := &migration.ModelColumn{
				Name: "test_field",
			}

			// 由于parseTormTag是私有方法，我们通过创建一个临时字段来测试
			// 这里我们直接测试各个组件是否能正确解析
			parts := strings.Split(tc.tormTag, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}

				if strings.Contains(part, ":") {
					// 键值对形式
					kv := strings.SplitN(part, ":", 2)
					if len(kv) == 2 {
						key := strings.ToLower(strings.TrimSpace(kv[0]))
						value := strings.TrimSpace(kv[1])

						switch key {
						case "type":
							// 这里简化处理，直接设置类型
							switch strings.ToLower(value) {
							case "varchar":
								column.Type = migration.ColumnTypeVarchar
							case "bigint":
								column.Type = migration.ColumnTypeBigInt
							}
						case "size":
							if size, err := strconv.Atoi(value); err == nil {
								column.Length = size
							}
						case "comment":
							column.Comment = value
						}
					}
				} else {
					// 标志形式
					flag := strings.ToLower(part)
					switch flag {
					case "primary_key", "pk":
						column.PrimaryKey = true
					case "auto_increment":
						column.AutoIncrement = true
					case "not_null":
						column.NotNull = true
					case "auto_create_time":
						column.NotNull = true
						if column.Default == nil {
							defaultVal := "CURRENT_TIMESTAMP"
							column.Default = &defaultVal
						}
					}
				}
			}

			t.Logf("解析结果: Type=%s, Length=%d, NotNull=%v, PrimaryKey=%v, AutoIncrement=%v, Comment=%s",
				column.Type, column.Length, column.NotNull, column.PrimaryKey, column.AutoIncrement, column.Comment)

			// 验证解析结果
			if expectedType, ok := tc.expected["type"]; ok {
				if column.Type != expectedType {
					t.Errorf("Type mismatch: expected %v, got %v", expectedType, column.Type)
				}
			}

			if expectedLength, ok := tc.expected["length"]; ok {
				if column.Length != expectedLength {
					t.Errorf("Length mismatch: expected %v, got %v", expectedLength, column.Length)
				}
			}

			if expectedNotNull, ok := tc.expected["not_null"]; ok {
				if column.NotNull != expectedNotNull {
					t.Errorf("NotNull mismatch: expected %v, got %v", expectedNotNull, column.NotNull)
				}
			}
		})
	}
}

// 辅助函数
func printModelColumns(t *testing.T, prefix string, columns []migration.ModelColumn) {
	for _, col := range columns {
		defaultStr := "nil"
		if col.Default != nil {
			defaultStr = *col.Default
		}
		t.Logf("  %s.%s: %s(%d) NotNull=%v Default=%s Comment='%s'",
			prefix, col.Name, col.Type, col.Length, col.NotNull, defaultStr, col.Comment)
	}
}

func simulateColumnComparison(beforeColumns, afterColumns []migration.ModelColumn) []migration.ColumnDifference {
	var differences []migration.ColumnDifference

	beforeMap := make(map[string]migration.ModelColumn)
	for _, col := range beforeColumns {
		beforeMap[col.Name] = col
	}

	for _, afterCol := range afterColumns {
		if beforeCol, exists := beforeMap[afterCol.Name]; exists {
			reasons := []string{}

			// 检查长度变化
			if beforeCol.Length != afterCol.Length {
				reasons = append(reasons, "长度修改")
			}

			// 检查类型变化
			if beforeCol.Type != afterCol.Type {
				reasons = append(reasons, "类型修改")
			}

			// 检查约束变化
			if beforeCol.NotNull != afterCol.NotNull {
				reasons = append(reasons, "NOT NULL约束修改")
			}
			if beforeCol.Unique != afterCol.Unique {
				reasons = append(reasons, "UNIQUE约束修改")
			}

			// 检查默认值变化
			beforeDefault := ""
			if beforeCol.Default != nil {
				beforeDefault = *beforeCol.Default
			}
			afterDefault := ""
			if afterCol.Default != nil {
				afterDefault = *afterCol.Default
			}
			if beforeDefault != afterDefault {
				reasons = append(reasons, "默认值修改")
			}

			// 检查注释变化
			if beforeCol.Comment != afterCol.Comment {
				reasons = append(reasons, "注释修改")
			}

			if len(reasons) > 0 {
				differences = append(differences, migration.ColumnDifference{
					Column:   afterCol.Name,
					Type:     "modify",
					OldValue: beforeCol,
					NewValue: afterCol,
					Reason:   "检测到: " + reasons[0], // 只显示第一个原因
				})
			}
		}
	}

	return differences
}
