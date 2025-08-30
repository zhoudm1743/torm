package main

import (
	"fmt"
	"log"
	"time"

	"github.com/zhoudm1743/torm/db"
	"github.com/zhoudm1743/torm/model"
)

// SimpleUser 简单用户模型 - 测试时间管理
type SimpleUser struct {
	model.BaseModel

	ID        int       `torm:"primary_key,auto_increment"`
	Username  string    `torm:"type:varchar,size:50"`
	CreatedAt time.Time `torm:"auto_create_time"` // 自动创建时间
	UpdatedAt time.Time `torm:"auto_update_time"` // 自动更新时间
}

func (u *SimpleUser) GetTableName() string {
	return "simple_users"
}

func main() {
	fmt.Println("=== 时间管理功能简单测试 ===")

	// 1. 配置SQLite数据库
	err := db.AddConnection("default", &db.Config{
		Driver:   "sqlite",
		Database: "test_simple.db",
	})
	if err != nil {
		log.Fatal("数据库配置失败:", err)
	}

	// 2. 创建用户模型
	user := &SimpleUser{}

	// 3. 自动迁移
	fmt.Println("正在迁移表结构...")
	err = user.AutoMigrate(user)
	if err != nil {
		log.Fatal("表迁移失败:", err)
	}
	fmt.Println("表迁移成功")

	// 4. 分析时间字段
	fmt.Println("\n--- 时间字段分析 ---")
	timeFields := user.GetTimeFields()
	fmt.Printf("检测到 %d 个时间字段:\n", len(timeFields))
	for _, field := range timeFields {
		fmt.Printf("  - %s (%s) -> %s [创建:%t, 更新:%t]\n",
			field.FieldName, field.FieldType, field.ColumnName,
			field.IsCreateTime, field.IsUpdateTime)
	}

	// 5. 测试时间管理器
	fmt.Println("\n--- 测试时间管理器 ---")
	timeManager := db.NewTimeFieldManager()

	// 显示时间管理器创建成功
	fmt.Println("时间管理器创建成功")

	// 6. 测试数据插入
	fmt.Println("\n--- 测试插入数据 ---")
	data := map[string]interface{}{
		"username": "testuser",
	}

	// 手动处理时间字段
	if len(timeFields) > 0 {
		data = timeManager.ProcessInsertData(data, timeFields)
		fmt.Println("处理后的插入数据:")
		for k, v := range data {
			fmt.Printf("  %s: %v (%T)\n", k, v, v)
		}
	}

	// 7. 测试时间转换
	fmt.Println("\n--- 测试时间转换 ---")
	now := time.Now()

	// 演示不同类型的时间格式
	stringTime := now.Format("2006-01-02 15:04:05")
	int64Time := now.Unix()
	timeTime := now

	fmt.Printf("当前时间转换演示:\n")
	fmt.Printf("  string 格式: %v (%T)\n", stringTime, stringTime)
	fmt.Printf("  int64 时间戳: %v (%T)\n", int64Time, int64Time)
	fmt.Printf("  time.Time: %v (%T)\n", timeTime, timeTime)

	fmt.Println("\n=== 测试完成 ===")
}
