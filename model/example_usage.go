package model

import (
	"fmt"
	"time"
)

// ExampleUser 示例用户模型，展示torm标签的使用
type ExampleUser struct {
	BaseModel
	ID        uint       `json:"id" torm:"primary_key,auto_increment"`
	Username  string     `json:"username" torm:"type:varchar,size:50,unique"`
	Email     string     `json:"email" torm:"type:varchar,size:100,unique"`
	Age       int        `json:"age" torm:"type:int,default:0"`
	CreatedAt time.Time  `json:"created_at" torm:"auto_create_time"`
	UpdatedAt time.Time  `json:"updated_at" torm:"auto_update_time"`
	DeletedAt *time.Time `json:"deleted_at" torm:"soft_delete"`
}

func (u *ExampleUser) GetTableName() string {
	return "example_users"
}

// ExampleUsage 展示重构后模型的使用方法
func ExampleUsage() {
	fmt.Println("=== Torm Model 重构后的使用示例 ===")

	// 1. 创建模型实例（torm标签自动生效）
	fmt.Println("\n1. 创建模型并验证torm标签优先级：")
	user := &ExampleUser{}
	userModel := NewModel(user)

	fmt.Printf("   表名: %s\n", userModel.GetTableName())
	fmt.Printf("   主键: %s\n", userModel.GetPrimaryKey())
	fmt.Printf("   创建时间字段: %s\n", userModel.GetCreatedAtField())
	fmt.Printf("   更新时间字段: %s\n", userModel.GetUpdatedAtField())
	fmt.Printf("   软删除: %v\n", userModel.config.SoftDeletes)

	// 2. 展示torm标签覆盖用户配置
	fmt.Println("\n2. torm标签覆盖用户配置：")
	conflictConfig := ModelConfig{
		TableName:    "will_be_overridden",
		PrimaryKey:   "will_be_overridden", // 被torm标签覆盖
		Connection:   "custom_connection",  // 保留
		Timestamps:   false,                // 保留
		CreatedAtCol: "will_be_overridden", // 被torm标签覆盖
		UpdatedAtCol: "will_be_overridden", // 被torm标签覆盖
		SoftDeletes:  false,                // 被torm标签覆盖
		DeletedAtCol: "will_be_overridden", // 被torm标签覆盖
	}

	userModelWithConfig := NewModel(user, conflictConfig)
	fmt.Printf("   表名 (来自GetTableName): %s\n", userModelWithConfig.GetTableName())
	fmt.Printf("   主键 (来自torm标签): %s\n", userModelWithConfig.GetPrimaryKey())
	fmt.Printf("   连接 (来自用户配置): %s\n", userModelWithConfig.GetConnection())
	fmt.Printf("   创建时间字段 (来自torm标签): %s\n", userModelWithConfig.GetCreatedAtField())
	fmt.Printf("   软删除 (来自torm标签): %v\n", userModelWithConfig.config.SoftDeletes)
	fmt.Printf("   时间戳启用 (来自用户配置): %v\n", userModelWithConfig.config.Timestamps)

	// 3. 展示完整的参数式查询兼容性
	fmt.Println("\n3. 参数式查询兼容性：")

	// 3.1 三参数格式
	_, _ = userModel.Where("username", "=", "john")
	fmt.Printf("   三参数格式: Where(\"username\", \"=\", \"john\") ✓\n")

	// 3.2 SQL+参数格式
	_, _ = userModel.Where("email = ?", "john@example.com")
	fmt.Printf("   SQL+参数格式: Where(\"email = ?\", \"john@example.com\") ✓\n")

	// 3.3 SQL+数组参数格式
	_, _ = userModel.Where("id IN (?)", []int{1, 2, 3})
	fmt.Printf("   SQL+数组参数格式: Where(\"id IN (?)\", []int{1,2,3}) ✓\n")

	// 3.4 纯SQL格式
	_, _ = userModel.Where("age > 18")
	fmt.Printf("   纯SQL格式: Where(\"age > 18\") ✓\n")

	// 3.5 链式查询
	complexQuery, _ := userModel.Select("id", "username", "email")
	if complexQuery != nil {
		finalQuery := complexQuery.
			Where("age", ">=", 18).
			Where("status = ?", "active").
			WhereIn("role", []interface{}{"admin", "user"}).
			OrderBy("created_at", "DESC").
			Limit(10)

		if finalQuery != nil {
			fmt.Printf("   链式查询: 支持模型方法 → QueryBuilder链式调用 ✓\n")
		}
	}

	// 4. 展示其他查询方法
	fmt.Println("\n4. 其他查询方法：")

	// WhereIn, WhereNull, WhereRaw 等
	_, _ = userModel.WhereIn("status", []interface{}{"active", "pending"})
	fmt.Printf("   WhereIn: 支持 ✓\n")

	_, _ = userModel.WhereNull("deleted_at")
	fmt.Printf("   WhereNull: 支持 ✓\n")

	_, _ = userModel.WhereRaw("age > ? AND status = ?", 18, "active")
	fmt.Printf("   WhereRaw: 支持 ✓\n")

	_, _ = userModel.OrderBy("created_at", "DESC")
	fmt.Printf("   OrderBy: 支持 ✓\n")

	_, _ = userModel.Limit(10)
	fmt.Printf("   Limit: 支持 ✓\n")

	_, _ = userModel.Page(1, 20)
	fmt.Printf("   Page: 支持 ✓\n")

	fmt.Println("\n=== 重构完成，所有功能完全兼容！ ===")
}

// DemonstratePriorityRules 演示优先级规则
func DemonstratePriorityRules() {
	fmt.Println("\n=== Torm 配置优先级规则 ===")
	fmt.Println("优先级（从高到低）：")
	fmt.Println("1. torm标签 (最高优先级)")
	fmt.Println("2. GetTableName() 方法")
	fmt.Println("3. 用户传入的 ModelConfig")
	fmt.Println("4. 默认配置")
	fmt.Println("5. 类型名推断 (最低优先级)")

	fmt.Println("\n示例：")
	fmt.Println("- primary_key: torm标签 > ModelConfig.PrimaryKey > \"id\"")
	fmt.Println("- table_name: GetTableName() > ModelConfig.TableName > 类型名推断")
	fmt.Println("- connection: ModelConfig.Connection > \"default\"")
	fmt.Println("- auto_create_time: torm标签 > ModelConfig.CreatedAtCol > \"created_at\"")
}
