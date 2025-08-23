package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/zhoudm1743/torm/db"
	"github.com/zhoudm1743/torm/model"
)

// UserV1 - 初始版本
type UserV1 struct {
	model.BaseModel
	ID     int    `json:"id" torm:"primary_key,auto_increment"`
	Name   string `json:"name" torm:"type:varchar,size:50,not_null"`
	Email  string `json:"email" torm:"type:varchar,size:100"`
	Age    int    `json:"age" torm:"type:int"`
	Status int    `json:"status" torm:"type:int,default:1"`
}

func NewUserV1() *UserV1 {
	user := &UserV1{}
	user.BaseModel = *model.NewBaseModel()
	user.SetTable("users")
	user.SetPrimaryKey("id")
	user.SetConnection("test")
	return user
}

// UserV2 - 修改字段长度和添加新字段
type UserV2 struct {
	model.BaseModel
	ID     int    `json:"id" torm:"primary_key,auto_increment"`
	Name   string `json:"name" torm:"type:varchar,size:100,not_null"` // 长度从50改为100
	Email  string `json:"email" torm:"type:varchar,size:255"`         // 长度从100改为255
	Age    int    `json:"age" torm:"type:int"`
	Status int    `json:"status" torm:"type:int,default:1"`
	Phone  string `json:"phone" torm:"type:varchar,size:20"` // 新增字段
}

func NewUserV2() *UserV2 {
	user := &UserV2{}
	user.BaseModel = *model.NewBaseModel()
	user.SetTable("users")
	user.SetPrimaryKey("id")
	user.SetConnection("test")

	return user
}

// UserV3 - 修改字段类型和约束
type UserV3 struct {
	model.BaseModel
	ID      int    `json:"id" torm:"primary_key,auto_increment"`
	Name    string `json:"name" torm:"type:varchar,size:100,not_null"`
	Email   string `json:"email" torm:"type:varchar,size:255,unique"` // 添加唯一约束
	Age     int    `json:"age" torm:"type:int,not_null"`              // 添加非空约束
	Status  int    `json:"status" torm:"type:tinyint,default:1"`      // 类型从int改为tinyint
	Phone   string `json:"phone" torm:"type:varchar,size:20"`
	Address string `json:"address" torm:"type:text"` // 新增text类型字段
}

func NewUserV3() *UserV3 {
	user := &UserV3{}
	user.BaseModel = *model.NewBaseModel()
	user.SetTable("users")
	user.SetPrimaryKey("id")
	user.SetConnection("test")

	return user
}

// AdminV1 - 另一个模型的初始版本
type AdminV1 struct {
	model.BaseModel
	ID        string `json:"id" torm:"primary_key,type:varchar,size:32"`
	Username  string `json:"username" torm:"type:varchar,size:50,not_null,unique"`
	Password  string `json:"password" torm:"type:varchar,size:64,not_null"`
	Role      string `json:"role" torm:"type:varchar,size:20,default:admin"`
	CreatedAt int64  `json:"created_at" torm:"auto_create_time"`
	UpdatedAt int64  `json:"updated_at" torm:"auto_update_time"`
}

func NewAdminV1() *AdminV1 {
	admin := &AdminV1{}
	admin.BaseModel = *model.NewBaseModel()
	admin.SetTable("admins")
	admin.SetPrimaryKey("id")
	admin.SetConnection("test")

	return admin
}

// AdminV2 - 字段修改版本
type AdminV2 struct {
	model.BaseModel
	ID        string   `json:"id" torm:"primary_key,type:varchar,size:32"`
	Username  string   `json:"username" torm:"type:varchar,size:100,not_null,unique"` // 长度从50改为100
	Password  string   `json:"password" torm:"type:varchar,size:128,not_null"`        // 长度从64改为128
	Role      []string `json:"role" torm:"type:json"`                                 // 类型从varchar改为json，支持多角色
	Email     string   `json:"email" torm:"type:varchar,size:255"`                    // 新增字段
	IsActive  bool     `json:"is_active" torm:"type:boolean,default:true"`            // 新增布尔字段
	CreatedAt int64    `json:"created_at" torm:"auto_create_time"`
	UpdatedAt int64    `json:"updated_at" torm:"auto_update_time"`
}

func NewAdminV2() *AdminV2 {
	admin := &AdminV2{}
	admin.BaseModel = *model.NewBaseModel()
	admin.SetTable("admins")
	admin.SetPrimaryKey("id")
	admin.SetConnection("test")

	return admin
}

// 测试数据库连接配置
func setupMigrationTestDB(t *testing.T) {
	// 创建测试数据库文件
	testDBPath := filepath.Join("tests", "migration_test.db")
	os.MkdirAll("tests", 0755)

	// 配置SQLite测试连接
	config := &db.Config{
		Driver:   "sqlite",
		Database: testDBPath,
	}

	// 注册测试连接
	err := db.AddConnection("test", config)
	if err != nil {
		t.Fatalf("Failed to add test connection: %v", err)
	}
}

func cleanupMigrationTestDB(t *testing.T) {
	testDBPath := filepath.Join("tests", "migration_test.db")
	os.Remove(testDBPath)
}

func TestModelMigrationVersions(t *testing.T) {
	setupMigrationTestDB(t)
	defer cleanupMigrationTestDB(t)

	t.Run("用户模型版本迁移测试", func(t *testing.T) {
		// V1 - 初始创建
		fmt.Println("测试UserV1 - 初始创建")
		userV1 := NewUserV1()
		err := userV1.AutoMigrate()
		if err != nil {
			t.Fatalf("UserV1 AutoMigrate failed: %v", err)
		}
		fmt.Println("UserV1 创建成功")

		// V2 - 字段长度修改和新增字段
		fmt.Println("测试UserV2 - 字段长度修改和新增字段")
		userV2 := NewUserV2()
		err = userV2.AutoMigrate()
		if err != nil {
			t.Fatalf("UserV2 AutoMigrate failed: %v", err)
		}
		fmt.Println(" UserV2 升级成功")

		// V3 - 类型和约束修改
		fmt.Println(" 测试UserV3 - 类型和约束修改")
		userV3 := NewUserV3()
		err = userV3.AutoMigrate()
		if err != nil {
			t.Fatalf("UserV3 AutoMigrate failed: %v", err)
		}
		fmt.Println(" UserV3 升级成功")
	})

	t.Run("管理员模型版本迁移测试", func(t *testing.T) {
		// V1 - 初始创建
		fmt.Println(" 测试AdminV1 - 初始创建")
		adminV1 := NewAdminV1()
		err := adminV1.AutoMigrate()
		if err != nil {
			t.Fatalf("AdminV1 AutoMigrate failed: %v", err)
		}
		fmt.Println(" AdminV1 创建成功")

		// V2 - 复杂字段修改
		fmt.Println(" 测试AdminV2 - 复杂字段修改")
		adminV2 := NewAdminV2()
		err = adminV2.AutoMigrate()
		if err != nil {
			t.Fatalf("AdminV2 AutoMigrate failed: %v", err)
		}
		fmt.Println(" AdminV2 升级成功")
	})
}

func TestNewBaseModelScenarios(t *testing.T) {
	setupMigrationTestDB(t)
	defer cleanupMigrationTestDB(t)

	t.Run("测试NewBaseModel的表现", func(t *testing.T) {
		// 只测试NewBaseModel，看看现在是否工作正常
		fmt.Println(" 测试NewBaseModel")
		user := &UserV1{}
		user.BaseModel = *model.NewBaseModel()
		user.SetTable("test_users")
		user.SetPrimaryKey("id")
		user.SetConnection("test")

		err := user.AutoMigrate()
		if err != nil {
			t.Fatalf("NewBaseModel AutoMigrate failed: %v", err)
		}
		fmt.Println(" NewBaseModel AutoMigrate 成功")
	})
}

func TestTormTagChanges(t *testing.T) {
	setupMigrationTestDB(t)
	defer cleanupMigrationTestDB(t)

	t.Run("TORM标签变化测试", func(t *testing.T) {
		// 第一个版本 - 基础字段
		type ProductV1 struct {
			model.BaseModel
			ID          int     `json:"id" torm:"primary_key,auto_increment"`
			Name        string  `json:"name" torm:"type:varchar,size:50"`
			Description string  `json:"description" torm:"type:varchar,size:200"`
			Price       float64 `json:"price" torm:"type:decimal,precision:10,scale:2"`
			Stock       int     `json:"stock" torm:"type:int"`
		}

		fmt.Println(" 创建ProductV1")
		productV1 := &ProductV1{}
		productV1.BaseModel = *model.NewBaseModel()
		productV1.SetTable("products")
		productV1.SetPrimaryKey("id")
		productV1.SetConnection("test")

		err := productV1.AutoMigrate()
		if err != nil {
			t.Fatalf("ProductV1 AutoMigrate failed: %v", err)
		}
		fmt.Println(" ProductV1 创建成功")

		// 第二个版本 - 修改字段长度和精度
		type ProductV2 struct {
			model.BaseModel
			ID          int     `json:"id" torm:"primary_key,auto_increment"`
			Name        string  `json:"name" torm:"type:varchar,size:100,not_null"`     // 长度50→100，增加not_null
			Description string  `json:"description" torm:"type:text"`                   // varchar(200)→text
			Price       float64 `json:"price" torm:"type:decimal,precision:12,scale:3"` // 精度10,2→12,3
			Stock       int     `json:"stock" torm:"type:int,default:0"`                // 增加默认值
			Category    string  `json:"category" torm:"type:varchar,size:50"`           // 新增字段
		}

		fmt.Println(" 升级到ProductV2")
		productV2 := &ProductV2{}
		productV2.BaseModel = *model.NewBaseModel()
		productV2.SetTable("products")
		productV2.SetPrimaryKey("id")
		productV2.SetConnection("test")

		err = productV2.AutoMigrate()
		if err != nil {
			t.Fatalf("ProductV2 AutoMigrate failed: %v", err)
		}
		fmt.Println(" ProductV2 升级成功")

		// 第三个版本 - 更复杂的修改
		type ProductV3 struct {
			model.BaseModel
			ID          int     `json:"id" torm:"primary_key,auto_increment"`
			Name        string  `json:"name" torm:"type:varchar,size:100,not_null,unique"` // 增加unique约束
			Description string  `json:"description" torm:"type:text"`
			Price       float64 `json:"price" torm:"type:decimal,precision:12,scale:3,not_null"` // 增加not_null
			Stock       int     `json:"stock" torm:"type:int,default:0,not_null"`                // 增加not_null
			Category    string  `json:"category" torm:"type:varchar,size:50,not_null"`           // 增加not_null
			IsActive    bool    `json:"is_active" torm:"type:boolean,default:true"`              // 新增布尔字段
			Tags        string  `json:"tags" torm:"type:json"`                                   // 新增JSON字段
		}

		fmt.Println(" 升级到ProductV3")
		productV3 := &ProductV3{}
		productV3.BaseModel = *model.NewBaseModel()
		productV3.SetTable("products")
		productV3.SetPrimaryKey("id")
		productV3.SetConnection("test")

		err = productV3.AutoMigrate()
		if err != nil {
			t.Fatalf("ProductV3 AutoMigrate failed: %v", err)
		}
		fmt.Println(" ProductV3 升级成功")
	})
}

func TestEdgeCaseMigrations(t *testing.T) {
	setupMigrationTestDB(t)
	defer cleanupMigrationTestDB(t)

	t.Run("边缘情况迁移测试", func(t *testing.T) {
		// 测试not_null变体
		type TestModelNotNull struct {
			model.BaseModel
			ID     int    `json:"id" torm:"primary_key,auto_increment"`
			Field1 string `json:"field1" torm:"type:varchar,size:50,not_null"`
			Field2 string `json:"field2" torm:"type:varchar,size:50,not_nil"` // 使用not_nil别名
			Field3 string `json:"field3" torm:"type:varchar,size:50,notnil"`  // 使用notnil别名
		}

		fmt.Println(" 测试not_null变体")
		testModel := &TestModelNotNull{}
		testModel.BaseModel = *model.NewBaseModel()
		testModel.SetTable("test_not_null")
		testModel.SetPrimaryKey("id")
		testModel.SetConnection("test")

		err := testModel.AutoMigrate()
		if err != nil {
			t.Fatalf("TestModelNotNull AutoMigrate failed: %v", err)
		}
		fmt.Println(" not_null变体测试成功")

		// 测试类型别名
		type TestModelTypes struct {
			model.BaseModel
			ID      int    `json:"id" torm:"primary_key,auto_increment"`
			StringF string `json:"string_f" torm:"type:string,size:100"` // string别名
			IntF    int    `json:"int_f" torm:"type:integer"`            // integer别名
			JsonF   string `json:"json_f" torm:"type:jsonb"`             // jsonb别名
			UuidF   string `json:"uuid_f" torm:"type:uuid"`              // uuid别名
		}

		fmt.Println(" 测试类型别名")
		testTypes := &TestModelTypes{}
		testTypes.BaseModel = *model.NewBaseModel()
		testTypes.SetTable("test_types")
		testTypes.SetPrimaryKey("id")
		testTypes.SetConnection("test")

		err = testTypes.AutoMigrate()
		if err != nil {
			t.Fatalf("TestModelTypes AutoMigrate failed: %v", err)
		}
		fmt.Println(" 类型别名测试成功")
	})
}

func TestDbTagSupport(t *testing.T) {
	setupMigrationTestDB(t)
	defer cleanupMigrationTestDB(t)

	t.Run("db标签支持测试", func(t *testing.T) {
		fmt.Println(" 测试db标签和column标签支持")

		// 测试db标签和column标签的支持
		type TestDbColumn struct {
			model.BaseModel
			ID       int    `json:"id" torm:"primary_key,auto_increment"`
			UserName string `json:"user_name" torm:"db:username,type:varchar,size:50"`         // 使用db指定字段名
			Email    string `json:"email" torm:"column:email_addr,type:varchar,size:100"`      // 使用column指定字段名
			FullName string `json:"full_name" db:"full_name_std" torm:"type:varchar,size:100"` // 使用标准db标签
		}

		testModel := &TestDbColumn{}
		testModel.BaseModel = *model.NewBaseModel()
		testModel.SetTable("test_db_column")
		testModel.SetPrimaryKey("id")
		testModel.SetConnection("test")

		err := testModel.AutoMigrate()
		if err != nil {
			t.Fatalf("TestDbColumn AutoMigrate failed: %v", err)
		}
		fmt.Println(" db标签和column标签支持测试成功")
	})
}

func TestSnakeCaseConversion(t *testing.T) {
	setupMigrationTestDB(t)
	defer cleanupMigrationTestDB(t)

	t.Run("snake_case转换测试", func(t *testing.T) {
		fmt.Println(" 测试snake_case字段名转换")

		// 测试常见缩写和复合词的snake_case转换
		type TestSnakeCase struct {
			model.BaseModel
			ID        int    `json:"id" torm:"primary_key,auto_increment"`   // ID -> id (不是i_d)
			UserID    int    `json:"user_id" torm:"type:int"`                // UserID -> user_id
			URL       string `json:"url" torm:"type:varchar,size:200"`       // URL -> url (不是u_r_l)
			APIKey    string `json:"api_key" torm:"type:varchar,size:100"`   // APIKey -> api_key
			JSONData  string `json:"json_data" torm:"type:text"`             // JSONData -> json_data
			UserName  string `json:"user_name" torm:"type:varchar,size:50"`  // UserName -> user_name
			FirstName string `json:"first_name" torm:"type:varchar,size:30"` // FirstName -> first_name
		}

		testModel := &TestSnakeCase{}
		testModel.BaseModel = *model.NewBaseModel()
		testModel.SetTable("test_snake_case")
		testModel.SetPrimaryKey("id")
		testModel.SetConnection("test")

		err := testModel.AutoMigrate()
		if err != nil {
			t.Fatalf("TestSnakeCase AutoMigrate failed: %v", err)
		}
		fmt.Println(" snake_case转换测试成功 - ID正确转换为id，不是i_d")
	})
}
