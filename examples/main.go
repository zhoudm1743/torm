package main

import (
	"examples/models"
	"log"

	"github.com/zhoudm1743/torm/pkg/db"
)

func main() {
	conf := &db.Config{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "123456",
		Database: "orm",
	}
	err := db.AddConnection("default", conf)
	if err != nil {
		log.Fatal(err)
	}

	// ===== 查询构建器演示 =====
	log.Println("===== 查询构建器演示 =====")

	// 基础查询
	query, err := db.Table("users", "default")
	if err == nil {
		// 查询所有用户
		users, err := query.Select("id", "name", "email", "age").
			Where("status", "=", "active").
			OrderBy("created_at", "desc").
			Limit(5).
			Get()
		if err == nil {
			log.Printf("查询到 %d 个活跃用户", len(users))
		}

		// 条件查询
		adults, err := query.Where("age", ">=", 18).
			Where("status", "=", "active").
			Count()
		if err == nil {
			log.Printf("成年活跃用户数量: %d", adults)
		}
	}

	// ===== First和Find新功能演示 =====
	log.Println("\n===== First和Find新功能演示 =====")

	// First方法 - 只填充当前模型
	user1 := models.NewUser()
	_, err = user1.Where("id", "=", 1).First()
	if err != nil {
		log.Printf("查询失败: %v", err)
	} else {
		log.Printf("First结果: Name=%s, Age=%d", user1.Name, user1.Age)
	}

	// First方法 - 同时填充传入的指针
	user2 := models.NewUser()
	var anotherUser models.User
	_, err = user2.Where("id", "=", 2).First(&anotherUser)
	if err != nil {
		log.Printf("查询失败: %v", err)
	} else {
		log.Printf("First + 指针填充: 当前=%s, 指针=%s", user2.Name, anotherUser.Name)
	}

	// Find方法 - 同时填充传入的指针
	user3 := models.NewUser()
	var targetUser models.User
	_, err = user3.Find(1, &targetUser)
	if err != nil {
		log.Printf("Find失败: %v", err)
	} else {
		log.Printf("Find + 指针填充: 当前=%s, 指针=%s", user3.Name, targetUser.Name)
	}

	// ===== db包First和Find方法演示 =====
	log.Println("\n===== db包First和Find方法演示 =====")

	// db.Table().First()
	query1, err := db.Table("users", "default")
	if err == nil {
		dbResult1, err := query1.Where("id", "=", 1).First()
		if err == nil {
			log.Printf("db.First() 结果: %s", dbResult1["name"])
		}
	}

	// db.Table().First(&model)
	query2, err := db.Table("users", "default")
	if err == nil {
		var userStruct models.User
		_, err := query2.Where("id", "=", 1).First(&userStruct)
		if err == nil {
			log.Printf("db.First(&model) 结果: Name=%s", userStruct.Name)
		}
	}

	// ===== 自定义主键功能演示 =====
	log.Println("\n===== 自定义主键功能演示 =====")

	// 默认主键
	user4 := models.NewUser()
	log.Printf("默认主键: %v", user4.PrimaryKeys())

	// UUID主键
	userUUID := models.NewUserWithUUID()
	userUUID.UUID = "550e8400-e29b-41d4-a716-446655440000"
	userUUID.SetAttribute("uuid", userUUID.UUID)
	log.Printf("UUID主键: %v, 值: %v", userUUID.PrimaryKeys(), userUUID.GetKey())

	// 复合主键
	userComposite := models.NewUserWithCompositePK()
	userComposite.SetAttribute("tenant_id", "tenant-001")
	userComposite.SetAttribute("user_id", "user-001")
	log.Printf("复合主键: %v, 值: %v", userComposite.PrimaryKeys(), userComposite.GetKey())

	// 手动设置主键
	user5 := models.NewUser()
	user5.SetPrimaryKeys([]string{"tenant_id", "user_code"})
	log.Printf("手动设置复合主键: %v", user5.PrimaryKeys())

	// ===== 高级查询功能演示 =====
	log.Println("\n===== 高级查询功能演示 =====")

	// 复杂条件查询
	complexQuery, err := db.Table("users", "default")
	if err == nil {
		result, err := complexQuery.
			Select("id", "name", "email").
			Where("age", "BETWEEN", []interface{}{20, 40}).
			WhereIn("status", []interface{}{"active", "pending"}).
			OrderBy("age", "ASC").
			OrderBy("name", "DESC").
			Limit(10).
			Get()
		if err == nil {
			log.Printf("复杂查询结果数量: %d", len(result))
		}
	}

	// 聚合查询
	aggregateQuery, err := db.Table("users", "default")
	if err == nil {
		count, err := aggregateQuery.Where("status", "=", "active").Count()
		if err == nil {
			log.Printf("活跃用户总数: %d", count)
		}
	}

	log.Println("\n===== 演示完成 =====")
}
