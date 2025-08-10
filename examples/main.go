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

	// 演示First和Find新功能
	log.Println("===== First和Find新功能演示 =====")

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

	// 演示db包的First和Find方法
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

	// 演示自定义主键功能
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

	log.Println("\n===== 演示完成 =====")
}
