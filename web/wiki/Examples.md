# 实例代码

本页面提供基于TORM实际测试代码的完整示例。

## 🚀 快速开始示例

```go
package main

import (
    "fmt"
    "time"
    "github.com/zhoudm1743/torm"
)

type User struct {
    torm.BaseModel
    ID        int       `json:"id" torm:"primary_key,auto_increment"`
    Username  string    `json:"username" torm:"type:varchar,size:50,unique,index"`
    Email     string    `json:"email" torm:"type:varchar,size:100,unique"`
    Age       int       `json:"age" torm:"type:int,unsigned,default:0"`
    Status    string    `json:"status" torm:"type:varchar,size:20,default:active,index"`
    CreatedAt time.Time `json:"created_at" torm:"auto_create_time"`
    UpdatedAt time.Time `json:"updated_at" torm:"auto_update_time"`
}

func main() {
    // 配置数据库
    err := torm.AddConnection("default", &torm.Config{
        Driver:   "sqlite",
        Database: "quickstart.db",
    })
    if err != nil {
        panic(err)
    }
    
    // 自动创建表结构
    user := &User{}
    if err := user.AutoMigrate(); err != nil {
        panic(err)
    }
    
    // 创建用户
    newUser := &User{
        Username: "zhangsan",
        Email:    "zhangsan@example.com",
        Age:      25,
        Status:   "active",
    }
    if err := newUser.Save(); err != nil {
        panic(err)
    }
    
    // 查询用户
    users, err := torm.Table("users").
        Where("status", "=", "active").
        Where("age", ">=", 18).
        Get()
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("✅ 查询到 %d 个活跃用户\n", len(users))
}
```

## 🏷️ TORM标签完整示例

```go
type CompleteTagDemo struct {
    torm.BaseModel
    
    // 主键和自增
    ID int64 `torm:"primary_key,auto_increment,comment:主键ID"`
    
    // 字符串类型
    Username    string `torm:"type:varchar,size:50,unique,index,comment:用户名"`
    Email       string `torm:"type:varchar,size:100,unique,comment:邮箱"`
    Address     string `torm:"type:text,comment:地址"`
    Biography   string `torm:"type:longtext,comment:简介"`
    
    // 数值类型
    Age         int     `torm:"type:int,unsigned,default:0,comment:年龄"`
    Balance     float64 `torm:"type:decimal,precision:10,scale:2,default:0.00,comment:余额"`
    
    // 布尔类型
    IsActive    bool `torm:"type:boolean,default:1,comment:是否激活"`
    IsVip       bool `torm:"type:boolean,default:0,comment:是否VIP"`
    
    // 状态字段
    Status      string `torm:"type:varchar,size:20,default:active,index,comment:状态"`
    
    // JSON类型
    Preferences map[string]interface{} `torm:"type:json,comment:偏好"`
    Tags        []string               `torm:"type:json,comment:标签"`
    
    // 自动时间戳
    CreatedAt   time.Time `torm:"auto_create_time,comment:创建时间"`
    UpdatedAt   time.Time `torm:"auto_update_time,comment:更新时间"`
}

// 外键关系演示
type Department struct {
    torm.BaseModel
    ID       int     `torm:"primary_key,auto_increment"`
    Name     string  `torm:"type:varchar,size:100,unique"`
    Budget   float64 `torm:"type:decimal,precision:12,scale:2,default:0.00"`
    IsActive bool    `torm:"type:boolean,default:1"`
}

type Employee struct {
    torm.BaseModel
    ID       int     `torm:"primary_key,auto_increment"`
    Name     string  `torm:"type:varchar,size:100,not_null"`
    Email    string  `torm:"type:varchar,size:100,unique"`
    Salary   float64 `torm:"type:decimal,precision:10,scale:2"`
    
    // 外键
    DeptID   int `torm:"type:int,references:departments.id,on_delete:set_null"`
}
```

## 🔄 自动迁移示例

```go
// 第一版：基础模型
type UserV1 struct {
    torm.BaseModel
    ID   int    `torm:"primary_key,auto_increment"`
    Name string `torm:"type:varchar,size:50"`
}

// 第二版：添加字段
type UserV2 struct {
    torm.BaseModel
    ID    int    `torm:"primary_key,auto_increment"`
    Name  string `torm:"type:varchar,size:50"`
    Email string `torm:"type:varchar,size:100,unique"`  // 新增
    Age   int    `torm:"type:int,default:0"`            // 新增
}

// 第三版：修改字段
type UserV3 struct {
    torm.BaseModel
    ID     int    `torm:"primary_key,auto_increment"`
    Name   string `torm:"type:varchar,size:100"`         // 长度从50改为100
    Email  string `torm:"type:varchar,size:100,unique"`
    Age    int    `torm:"type:int,default:0"`
    Status string `torm:"type:varchar,size:20,default:active,index"` // 新增
}

// 迁移演示
func demonstrateMigration() {
    // v1.0.0 部署
    userV1 := &UserV1{}
    userV1.AutoMigrate() // 创建基础表
    
    // v1.1.0 部署
    userV2 := &UserV2{}
    userV2.AutoMigrate() // 智能添加新字段
    
    // v1.2.0 部署
    userV3 := &UserV3{}
    userV3.AutoMigrate() // 智能修改字段和添加新字段
}
```

## 🔍 查询构建器示例

```go
func queryExamples() {
    // 基础查询
    users, _ := torm.Table("users").
        Where("status", "=", "active").
        Where("age", ">=", 18).
        OrderBy("created_at", "desc").
        Limit(10).
        Get()
    
    // 参数化查询（防SQL注入）
    paramUsers, _ := torm.Table("users").
        Where("username = ? AND age >= ?", "zhangsan", 25).
        Get()
    
    // 数组参数查询
    arrayUsers, _ := torm.Table("users").
        Where("status IN (?)", []string{"active", "premium", "vip"}).
        Where("id IN (?)", []int{1, 2, 3, 4, 5}).
        Get()
    
    // 复杂条件
    complexUsers, _ := torm.Table("users").
        Where("(status = ? OR vip_level > ?) AND age BETWEEN ? AND ?", 
              "premium", 3, 18, 65).
        Get()
    
    // 聚合查询
    count, _ := torm.Table("users").
        Where("status", "=", "active").
        Count()
    
    // 分页查询
    pagination, _ := torm.Table("users").
        Where("status", "=", "active").
        Paginate(1, 20)
    
    // JOIN查询
    joinResults, _ := torm.Table("users").
        LeftJoin("profiles", "profiles.user_id", "=", "users.id").
        Select("users.name", "profiles.avatar").
        Get()
}
```

## 💼 事务处理示例

```go
// 银行转账事务
func transferMoney(fromUserID, toUserID int64, amount float64) error {
    return torm.Transaction(func(tx torm.TransactionInterface) error {
        // 检查发送方余额
        var fromBalance float64
        row := tx.QueryRow("SELECT balance FROM accounts WHERE user_id = ?", fromUserID)
        if err := row.Scan(&fromBalance); err != nil {
            return fmt.Errorf("查询余额失败: %v", err)
        }
        
        if fromBalance < amount {
            return fmt.Errorf("余额不足")
        }
        
        // 扣除发送方余额
        _, err := tx.Exec("UPDATE accounts SET balance = balance - ? WHERE user_id = ?", 
                         amount, fromUserID)
        if err != nil {
            return err
        }
        
        // 增加接收方余额
        _, err = tx.Exec("UPDATE accounts SET balance = balance + ? WHERE user_id = ?", 
                        amount, toUserID)
        if err != nil {
            return err
        }
        
        // 记录日志
        _, err = tx.Exec("INSERT INTO transfer_logs (from_user_id, to_user_id, amount) VALUES (?, ?, ?)",
                        fromUserID, toUserID, amount)
        if err != nil {
            return err
        }
        
        return nil // 自动提交
    })
}

// 订单创建事务
func createOrderWithItems(userID int64, items []OrderItem) error {
    return torm.Transaction(func(tx torm.TransactionInterface) error {
        // 创建订单
        result, err := tx.Exec("INSERT INTO orders (user_id, total_amount, status) VALUES (?, ?, ?)",
                              userID, calculateTotal(items), "pending")
        if err != nil {
            return err
        }
        
        orderID, _ := result.LastInsertId()
        
        // 创建订单项并扣减库存
        for _, item := range items {
            // 检查库存
            var stock int
            row := tx.QueryRow("SELECT stock FROM products WHERE id = ?", item.ProductID)
            if err := row.Scan(&stock); err != nil {
                return err
            }
            
            if stock < item.Quantity {
                return fmt.Errorf("库存不足")
            }
            
            // 扣减库存
            _, err = tx.Exec("UPDATE products SET stock = stock - ? WHERE id = ?", 
                            item.Quantity, item.ProductID)
            if err != nil {
                return err
            }
            
            // 创建订单项
            _, err = tx.Exec("INSERT INTO order_items (order_id, product_id, quantity, price) VALUES (?, ?, ?, ?)",
                            orderID, item.ProductID, item.Quantity, item.Price)
            if err != nil {
                return err
            }
        }
        
        return nil
    })
}
```

## 🌐 跨数据库示例

```go
type User struct {
    torm.BaseModel
    ID       int       `torm:"primary_key,auto_increment"`
    Username string    `torm:"type:varchar,size:50,unique"`
    Balance  float64   `torm:"type:decimal,precision:10,scale:2"`
    IsActive bool      `torm:"type:boolean,default:1"`
    Data     map[string]interface{} `torm:"type:json"`
    CreatedAt time.Time `torm:"auto_create_time"`
}

func setupMultipleDatabases() {
    // SQLite - 开发环境
    torm.AddConnection("dev", &torm.Config{
        Driver:   "sqlite",
        Database: "dev.db",
    })
    
    // MySQL - 测试环境
    torm.AddConnection("test", &torm.Config{
        Driver:   "mysql",
        Host:     "localhost",
        Database: "test_db",
        Username: "user",
        Password: "pass",
    })
    
    // PostgreSQL - 生产环境
    torm.AddConnection("prod", &torm.Config{
        Driver:   "postgres",
        Host:     "prod.server",
        Database: "prod_db",
        Username: "user",
        Password: "pass",
        SSLMode:  "require",
    })
}

func deployToAllEnvironments() {
    environments := []string{"dev", "test", "prod"}
    
    for _, env := range environments {
        user := &User{}
        user.SetConnection(env)
        
        // 同一模型，自动适配不同数据库
        if err := user.AutoMigrate(); err != nil {
            fmt.Printf("❌ %s 环境部署失败: %v\n", env, err)
        } else {
            fmt.Printf("✅ %s 环境部署成功\n", env)
        }
    }
}
```

## 🛒 电商系统示例

```go
// 电商模型定义
type Category struct {
    torm.BaseModel
    ID   int    `torm:"primary_key,auto_increment"`
    Name string `torm:"type:varchar,size:100,unique"`
}

type Product struct {
    torm.BaseModel
    ID         int     `torm:"primary_key,auto_increment"`
    Name       string  `torm:"type:varchar,size:200"`
    SKU        string  `torm:"type:varchar,size:50,unique"`
    Price      float64 `torm:"type:decimal,precision:10,scale:2"`
    Stock      int     `torm:"type:int,unsigned,default:0"`
    CategoryID int     `torm:"type:int,references:categories.id,on_delete:cascade"`
    IsActive   bool    `torm:"type:boolean,default:1"`
}

type Customer struct {
    torm.BaseModel
    ID       int     `torm:"primary_key,auto_increment"`
    Username string  `torm:"type:varchar,size:50,unique"`
    Email    string  `torm:"type:varchar,size:100,unique"`
    Balance  float64 `torm:"type:decimal,precision:10,scale:2,default:0.00"`
}

type Order struct {
    torm.BaseModel
    ID          int     `torm:"primary_key,auto_increment"`
    OrderNo     string  `torm:"type:varchar,size:32,unique"`
    CustomerID  int     `torm:"type:int,references:customers.id,on_delete:cascade"`
    TotalAmount float64 `torm:"type:decimal,precision:10,scale:2"`
    Status      string  `torm:"type:varchar,size:20,default:pending,index"`
}

type OrderItem struct {
    torm.BaseModel
    ID        int     `torm:"primary_key,auto_increment"`
    OrderID   int     `torm:"type:int,references:orders.id,on_delete:cascade"`
    ProductID int     `torm:"type:int,references:products.id,on_delete:cascade"`
    Quantity  int     `torm:"type:int,unsigned"`
    Price     float64 `torm:"type:decimal,precision:10,scale:2"`
}

// 初始化电商系统
func setupECommerceSystem() {
    torm.AddConnection("default", &torm.Config{
        Driver:   "mysql",
        Host:     "localhost",
        Database: "ecommerce",
        Username: "root",
        Password: "password",
    })
    
    // 按依赖顺序创建表
    (&Category{}).AutoMigrate()
    (&Customer{}).AutoMigrate()
    (&Product{}).AutoMigrate()
    (&Order{}).AutoMigrate()
    (&OrderItem{}).AutoMigrate()
}

// 商品搜索
func searchProducts(keyword string, categoryID int) ([]map[string]interface{}, error) {
    return torm.Table("products").
        LeftJoin("categories", "categories.id", "=", "products.category_id").
        Select("products.*", "categories.name as category_name").
        Where("products.is_active", "=", true).
        Where("products.stock", ">", 0).
        Where("products.name LIKE ?", "%"+keyword+"%").
        Where("products.category_id", "=", categoryID).
        OrderBy("products.created_at", "desc").
        Get()
}

// 订单统计
func getOrderStats() {
    // 今日订单数
    todayCount, _ := torm.Table("orders").
        Where("DATE(created_at) = ?", time.Now().Format("2006-01-02")).
        Count()
    
    // 客户订单统计
    customerStats, _ := torm.Table("orders").
        LeftJoin("customers", "customers.id", "=", "orders.customer_id").
        Select("customers.username", "COUNT(orders.id) as order_count", "SUM(orders.total_amount) as total_spent").
        GroupBy("customers.id").
        Having("order_count", ">", 0).
        Get()
    
    fmt.Printf("今日订单: %d, 活跃客户: %d\n", todayCount, len(customerStats))
}
```