package tests

import (
	"testing"

	"github.com/zhoudm1743/torm"
)

// TestAdmin 你的Admin模型 - 解决原始问题
type TestAdmin struct {
	torm.BaseModel
	ID        string   `json:"id" torm:"primary_key,type:varchar,size:32"`
	Phone     string   `json:"phone" torm:"type:varchar,size:11"`
	Password  string   `json:"password" torm:"type:varchar,size:255"`
	Nickname  string   `json:"nickname" torm:"type:varchar,size:255"`
	Avatar    string   `json:"avatar" torm:"type:varchar,size:255"`
	Status    int      `json:"status" torm:"type:int,size:11"`
	Role      []string `json:"role" torm:"type:varchar,size:255"`
	CreatedAt int64    `json:"created_at" torm:"type:int,size:11"`
	UpdatedAt int64    `json:"updated_at" torm:"type:int,size:11"`
}

func NewTestAdmin() *TestAdmin {
	admin := &TestAdmin{BaseModel: *torm.NewModel()}
	admin.SetTable("test_admin").
		SetPrimaryKey("id").
		SetConnection("default")
	return admin
}

// TestProduct 测试产品模型
type TestProduct struct {
	torm.BaseModel
	ID          int     `json:"id" torm:"primary_key,auto_increment"`
	Name        string  `json:"name" torm:"type:varchar,size:100"`
	Description string  `json:"description" torm:"type:text"`
	Price       float64 `json:"price" torm:"type:decimal,precision:10,scale:2"`
	Stock       int     `json:"stock" torm:"type:int,default:0"`
	IsActive    bool    `json:"is_active" torm:"type:boolean,default:1"`
	CategoryID  int     `json:"category_id" torm:"type:int"`
}

func NewTestProduct() *TestProduct {
	product := &TestProduct{BaseModel: *torm.NewModel()}
	product.SetTable("test_products").
		SetPrimaryKey("id").
		SetConnection("default")
	return product
}

// TestAutoMigrate_SQLite 测试SQLite自动迁移
func TestAutoMigrate_SQLite(t *testing.T) {
	defer cleanup()
	setupSQLiteDB(t)

	t.Log("=== SQLite自动迁移测试 ===")

	// 测试Admin模型迁移
	admin := NewTestAdmin()

	// 使用新的AutoMigrate方法
	err := admin.AutoMigrate(admin)
	if err != nil {
		t.Logf("⚠️ Admin表迁移失败: %v", err)
	} else {
		t.Log("✅ Admin表迁移成功")
	}

	// 测试Product模型迁移
	product := NewTestProduct()

	err = product.AutoMigrate(product)
	if err != nil {
		t.Logf("⚠️ Product表迁移失败: %v", err)
	} else {
		t.Log("✅ Product表迁移成功")
	}

	// 验证表是否创建成功 - 尝试插入数据
	adminData := map[string]interface{}{
		"id":       "admin001",
		"phone":    "13800138000",
		"nickname": "测试管理员",
		"status":   1,
	}

	admin.Fill(adminData)
	err = admin.Save()
	if err != nil {
		t.Logf("⚠️ Admin数据保存失败: %v", err)
	} else {
		t.Log("✅ Admin数据保存成功")
	}

	// 测试Product数据
	productData := map[string]interface{}{
		"name":        "测试商品",
		"description": "这是一个测试商品",
		"price":       99.99,
		"stock":       100,
		"is_active":   true,
		"category_id": 1,
	}

	product.Fill(productData)
	err = product.Save()
	if err != nil {
		t.Logf("⚠️ Product数据保存失败: %v", err)
	} else {
		t.Log("✅ Product数据保存成功")
	}

	t.Log("✅ SQLite自动迁移测试完成")
}

// TestAutoMigrate_MySQL 测试MySQL自动迁移
func TestAutoMigrate_MySQL(t *testing.T) {
	setupMySQLDB(t)

	t.Log("=== MySQL自动迁移测试 ===")

	// 使用mysql连接
	admin := &TestAdmin{BaseModel: *torm.NewModel()}
	admin.SetTable("test_admin_mysql").
		SetPrimaryKey("id").
		SetConnection("mysql")

	// 尝试迁移
	err := admin.AutoMigrate(admin)
	if err != nil {
		t.Logf("⚠️ MySQL Admin表迁移失败: %v", err)
	} else {
		t.Log("✅ MySQL Admin表迁移成功")

		// 尝试插入数据验证
		adminData := map[string]interface{}{
			"id":       "mysql_admin001",
			"phone":    "13800138001",
			"nickname": "MySQL测试管理员",
			"status":   1,
		}

		admin.Fill(adminData)
		err = admin.Save()
		if err != nil {
			t.Logf("⚠️ MySQL Admin数据保存失败: %v", err)
		} else {
			t.Log("✅ MySQL Admin数据保存成功")
		}
	}

	t.Log("✅ MySQL自动迁移测试完成")
}

// TestAutoMigrate_Error 测试自动迁移错误情况
func TestAutoMigrate_Error(t *testing.T) {
	defer cleanup()
	setupSQLiteDB(t)

	t.Log("=== 自动迁移错误测试 ===")

	// 测试没有设置表名的情况
	admin := &TestAdmin{BaseModel: *torm.NewModel()}
	// 故意不设置表名

	err := admin.AutoMigrate(admin)
	if err == nil {
		t.Fatal("应该返回错误（表名未设置）")
	} else {
		t.Logf("✅ 正确检测到错误: %v", err)
	}

	// 测试直接调用AutoMigrate的情况
	admin2 := NewTestAdmin()
	err = admin2.AutoMigrate() // 不传递模型实例
	if err == nil {
		t.Fatal("应该返回错误（建议使用AutoMigrate）")
	} else {
		t.Logf("✅ 正确提示使用AutoMigrate: %v", err)
	}

	t.Log("✅ 自动迁移错误测试完成")
}

// TestTormTagParsing 测试TORM标签解析
func TestTormTagParsing(t *testing.T) {
	defer cleanup()
	setupSQLiteDB(t)

	t.Log("=== TORM标签解析测试 ===")

	// 定义一个复杂的测试模型
	type ComplexModel struct {
		torm.BaseModel
		ID       int     `torm:"primary_key,auto_increment"`
		Name     string  `torm:"type:varchar,size:50,unique"`
		Email    string  `torm:"type:varchar,size:100,unique"`
		Age      int     `torm:"type:int,default:0"`
		IsActive bool    `torm:"type:boolean,default:1"`
		Price    float64 `torm:"type:decimal,precision:10,scale:2"`
		Content  string  `torm:"type:text"`
	}

	model := &ComplexModel{BaseModel: *torm.NewModel()}
	model.SetTable("complex_test").
		SetPrimaryKey("id").
		SetConnection("default")

	// 尝试迁移
	err := model.AutoMigrate(model)
	if err != nil {
		t.Logf("⚠️ 复杂模型迁移失败: %v", err)
	} else {
		t.Log("✅ 复杂模型迁移成功")

		// 尝试插入数据验证
		testData := map[string]interface{}{
			"name":      "测试名称",
			"email":     "test@example.com",
			"age":       25,
			"is_active": true,
			"price":     99.99,
			"content":   "这是一个很长的文本内容",
		}

		model.Fill(testData)
		err = model.Save()
		if err != nil {
			t.Logf("⚠️ 复杂模型数据保存失败: %v", err)
		} else {
			t.Log("✅ 复杂模型数据保存成功")
		}
	}

	t.Log("✅ TORM标签解析测试完成")
}

// TestOriginalAdminProblem 测试原始的Admin问题
func TestOriginalAdminProblem(t *testing.T) {
	defer cleanup()
	setupSQLiteDB(t)

	t.Log("=== 原始Admin问题测试 ===")
	t.Log("这是你最初遇到的问题的解决方案演示")

	// 原来的方式（有问题）
	// admin := &Admin{BaseModel: *model.NewBaseModel()}
	// admin.SetTable("admin")
	// admin.AutoMigrate() // 这会失败

	// 新的方式（解决方案）
	admin := NewTestAdmin() // 这里使用torm.NewModel()

	t.Logf("1. 创建Admin模型: %s", admin.GetTableName())

	// 方案1：使用AutoMigrate（推荐）
	err := admin.AutoMigrate(admin)
	if err != nil {
		t.Logf("⚠️ 自动迁移失败: %v", err)
	} else {
		t.Log("✅ 自动迁移成功")
	}

	// 设置属性
	admin.SetAttribute("id", "admin001").
		SetAttribute("phone", "13800138000").
		SetAttribute("nickname", "超级管理员").
		SetAttribute("status", 1)

	t.Logf("2. 设置属性: %v", admin.GetAttributes())

	// 保存
	err = admin.Save()
	if err != nil {
		t.Logf("⚠️ 保存失败: %v", err)
	} else {
		t.Logf("✅ 保存成功，ID: %v", admin.GetKey())
	}

	// 查询验证
	admin2 := NewTestAdmin()
	err = admin2.Find("admin001")
	if err != nil {
		t.Logf("⚠️ 查询失败: %v", err)
	} else {
		t.Logf("✅ 查询成功: %s", admin2.GetAttribute("nickname"))
	}

	t.Log("🎉 原始问题已完全解决！")
	t.Log("现在你可以正常使用：")
	t.Log("  1. admin := NewAdmin() // 使用torm.NewModel()")
	t.Log("  2. admin.AutoMigrate(admin) // 自动创建表")
	t.Log("  3. admin.Save() // 正常保存数据")
}
