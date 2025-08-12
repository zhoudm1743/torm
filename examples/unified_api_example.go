package main

import (
	"fmt"
	"log"

	"github.com/zhoudm1743/torm"
)

type SimpleUser struct {
	torm.BaseModel
	Name  string `json:"name"`
	Email string `json:"email"`
}

func unifiedAPIExample() {
	fmt.Printf("TORM Version: %s\n", torm.Version())

	// 使用统一入口配置数据库
	config := &torm.Config{
		Driver:   "sqlite",
		Database: ":memory:",
	}

	// 添加连接
	err := torm.AddConnection("default", config)
	if err != nil {
		log.Fatal("Failed to add connection:", err)
	}

	// 使用统一入口进行查询
	query, err := torm.Query()
	if err != nil {
		log.Fatal("Failed to create query:", err)
	}

	// 创建表查询
	userQuery, err := torm.Table("users")
	if err != nil {
		log.Fatal("Failed to create table query:", err)
	}

	fmt.Println("Query and Table methods are working!")
	fmt.Printf("Query type: %T\n", query)
	fmt.Printf("User query type: %T\n", userQuery)

	// 显示可用的数据库连接
	db, err := torm.DB()
	if err != nil {
		log.Fatal("Failed to get database connection:", err)
	}

	fmt.Printf("Database connection type: %T\n", db)
	fmt.Println("Successfully connected to database via unified API!")
}
