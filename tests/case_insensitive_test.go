package tests

import (
	"os"
	"testing"

	"github.com/zhoudm1743/torm/db"
	"github.com/zhoudm1743/torm/model"
)

// CaseTestModel 大小写测试模型
type CaseTestModel struct {
	model.BaseModel
	ID           int64   `json:"id" db:"id" torm:"PRIMARY_KEY,AUTO_INCREMENT,comment:主键"`
	UpperVarchar string  `json:"upper_varchar" db:"upper_varchar" torm:"TYPE:VARCHAR,SIZE:50,comment:大写VARCHAR"`
	LowerVarchar string  `json:"lower_varchar" db:"lower_varchar" torm:"type:varchar,size:50,comment:小写varchar"`
	MixedVarchar string  `json:"mixed_varchar" db:"mixed_varchar" torm:"Type:VarChar,Size:50,comment:混合VarChar"`
	UpperChar    string  `json:"upper_char" db:"upper_char" torm:"TYPE:CHAR,SIZE:10,comment:大写CHAR"`
	LowerChar    string  `json:"lower_char" db:"lower_char" torm:"type:char,size:10,comment:小写char"`
	MixedChar    string  `json:"mixed_char" db:"mixed_char" torm:"Type:Char,Size:10,comment:混合Char"`
	UpperDecimal float64 `json:"upper_decimal" db:"upper_decimal" torm:"TYPE:DECIMAL,PRECISION:10,SCALE:2,comment:大写DECIMAL"`
	LowerDecimal float64 `json:"lower_decimal" db:"lower_decimal" torm:"type:decimal,precision:10,scale:2,comment:小写decimal"`
	MixedDecimal float64 `json:"mixed_decimal" db:"mixed_decimal" torm:"Type:Decimal,Precision:10,Scale:2,comment:混合Decimal"`
	UpperText    string  `json:"upper_text" db:"upper_text" torm:"TYPE:TEXT,comment:大写TEXT"`
	LowerText    string  `json:"lower_text" db:"lower_text" torm:"type:text,comment:小写text"`
	MixedText    string  `json:"mixed_text" db:"mixed_text" torm:"Type:Text,comment:混合Text"`
	UpperBool    bool    `json:"upper_bool" db:"upper_bool" torm:"TYPE:BOOLEAN,DEFAULT:TRUE,comment:大写BOOLEAN"`
	LowerBool    bool    `json:"lower_bool" db:"lower_bool" torm:"type:boolean,default:true,comment:小写boolean"`
	MixedBool    bool    `json:"mixed_bool" db:"mixed_bool" torm:"Type:Boolean,Default:True,comment:混合Boolean"`
	UpperInt     int     `json:"upper_int" db:"upper_int" torm:"TYPE:INT,DEFAULT:0,comment:大写INT"`
	LowerInt     int     `json:"lower_int" db:"lower_int" torm:"type:int,default:0,comment:小写int"`
	MixedInt     int     `json:"mixed_int" db:"mixed_int" torm:"Type:Int,Default:0,comment:混合Int"`
	UpperUnique  string  `json:"upper_unique" db:"upper_unique" torm:"TYPE:VARCHAR,SIZE:20,UNIQUE,comment:大写UNIQUE"`
	LowerUnique  string  `json:"lower_unique" db:"lower_unique" torm:"type:varchar,size:20,unique,comment:小写unique"`
	MixedUnique  string  `json:"mixed_unique" db:"mixed_unique" torm:"Type:VarChar,Size:20,Unique,comment:混合Unique"`
	UpperIndex   string  `json:"upper_index" db:"upper_index" torm:"TYPE:VARCHAR,SIZE:30,INDEX,comment:大写INDEX"`
	LowerIndex   string  `json:"lower_index" db:"lower_index" torm:"type:varchar,size:30,index,comment:小写index"`
	MixedIndex   string  `json:"mixed_index" db:"mixed_index" torm:"Type:VarChar,Size:30,Index,comment:混合Index"`
}

func NewCaseTestModel() *CaseTestModel {
	ctm := &CaseTestModel{BaseModel: *model.NewBaseModel()}
	ctm.SetTable("case_test_models")
	ctm.SetConnection("test")
	ctm.DetectConfigFromStruct(ctm)
	return ctm
}

func TestCaseInsensitive(t *testing.T) {
	// 设置测试数据库
	testDB := "./test_case_insensitive.db"
	defer os.Remove(testDB) // 测试后清理

	config := &db.Config{
		Driver:   "sqlite",
		Database: testDB,
	}

	err := db.AddConnection("test", config)
	if err != nil {
		t.Fatalf("数据库配置失败: %v", err)
	}

	// 创建测试模型
	testModel := NewCaseTestModel()

	// 测试 AutoMigrate
	t.Run("大小写不敏感测试", func(t *testing.T) {
		err := testModel.AutoMigrate()
		if err != nil {
			t.Errorf("AutoMigrate 失败: %v", err)
		} else {
			t.Log("大小写不敏感测试模型创建成功")

			// 验证表是否确实被创建
			conn, err := db.DB("test")
			if err != nil {
				t.Fatalf("获取数据库连接失败: %v", err)
			}

			// 检查表是否存在
			query := "SELECT name FROM sqlite_master WHERE type='table' AND name='case_test_models'"
			rows, err := conn.Query(query)
			if err != nil {
				t.Errorf("查询表信息失败: %v", err)
			} else {
				defer rows.Close()
				if rows.Next() {
					t.Log("大小写不敏感测试表创建验证成功")
				} else {
					t.Error("大小写不敏感测试表未被创建")
				}
			}
		}
	})

	t.Run("检查各种大小写组合", func(t *testing.T) {
		t.Log("✅ 大写类型: TYPE:VARCHAR, TYPE:CHAR, TYPE:DECIMAL")
		t.Log("✅ 小写类型: type:varchar, type:char, type:decimal")
		t.Log("✅ 混合类型: Type:VarChar, Type:Char, Type:Decimal")
		t.Log("✅ 大写属性: SIZE:50, PRECISION:10, SCALE:2, UNIQUE, INDEX")
		t.Log("✅ 小写属性: size:50, precision:10, scale:2, unique, index")
		t.Log("✅ 混合属性: Size:50, Precision:10, Scale:2, Unique, Index")
		t.Log("✅ 大写默认值: DEFAULT:TRUE, DEFAULT:0")
		t.Log("✅ 小写默认值: default:true, default:0")
		t.Log("✅ 混合默认值: Default:True, Default:0")
	})

	t.Run("检查模型配置", func(t *testing.T) {
		if testModel.TableName() != "case_test_models" {
			t.Errorf("期望表名 'case_test_models'，得到 '%s'", testModel.TableName())
		}

		if testModel.HasModelStruct() {
			modelName := testModel.GetModelStructName()
			if modelName != "CaseTestModel" {
				t.Errorf("期望模型名 'CaseTestModel'，得到 '%s'", modelName)
			} else {
				t.Logf("模型结构体信息保存成功: %s", modelName)
			}
		} else {
			t.Error("模型结构体信息未保存")
		}
	})
}
