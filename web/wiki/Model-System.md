# 模型系统

TORM的模型系统基于强大的TORM标签和零配置的AutoMigrate功能，让数据库表设计变得简单而精确。

## 📋 目录

- [快速开始](#快速开始)
- [BaseModel基础](#BaseModel基础)
- [TORM标签系统](#TORM标签系统)
- [自动迁移](#自动迁移)
- [模型操作](#模型操作)
- [最佳实践](#最佳实践)

## 🚀 快速开始

### 基础模型定义

```go
package main

import (
    "time"
    "github.com/zhoudm1743/torm"
)

// 用户模型 - 展示基础TORM标签
type User struct {
    torm.BaseModel
    
    // 主键和自增
    ID int `json:"id" torm:"primary_key,auto_increment"`
    
    // 字符串类型和约束
    Username string `json:"username" torm:"type:varchar,size:50,unique,index"`
    Email    string `json:"email" torm:"type:varchar,size:100,unique"`
    Password string `json:"password" torm:"type:varchar,size:255"`
    
    // 数值类型
    Age    int     `json:"age" torm:"type:int,unsigned,default:0"`
    Salary float64 `json:"salary" torm:"type:decimal,precision:10,scale:2,default:0.00"`
    
    // 状态和布尔
    Status   string `json:"status" torm:"type:varchar,size:20,default:active,index"`
    IsActive bool   `json:"is_active" torm:"type:boolean,default:1"`
    
    // 文本类型
    Bio string `json:"bio" torm:"type:text"`
    
    // 自动时间戳
    CreatedAt time.Time `json:"created_at" torm:"auto_create_time"`
    UpdatedAt time.Time `json:"updated_at" torm:"auto_update_time"`
}

func main() {
    // 配置数据库
    torm.AddConnection("default", &torm.Config{
        Driver:   "mysql",
        Host:     "localhost",
        Database: "myapp",
        Username: "root",
        Password: "password",
    })
    
    // 自动创建表结构
    user := &User{}
    user.AutoMigrate()
    
    // 开始使用模型
    newUser := &User{
        Username: "zhangsan",
        Email:    "zhangsan@example.com",
        Age:      25,
        Status:   "active",
        IsActive: true,
        Bio:      "这是用户简介",
    }
    
    // 保存到数据库
    newUser.Save()
}
```

## 🏗️ BaseModel基础

### BaseModel 功能

```go
// BaseModel 提供的核心功能
type User struct {
    torm.BaseModel  // 继承基础功能
    // ... 你的字段
}

// BaseModel 提供的方法：
// - Save() error                           // 保存模型
// - Delete() error                         // 删除模型
// - AutoMigrate() error                    // 自动迁移
// - SetTable(name string)                  // 设置表名
// - SetConnection(name string)             // 设置连接
// - Where(conditions...) QueryBuilder     // 条件查询
// - OrderBy(column, direction string)      // 排序
// - Get() ([]map[string]interface{}, error) // 获取记录
// - First() (map[string]interface{}, error) // 获取单条
```

### 模型初始化

```go
// 创建新的用户实例
user := &User{
    Username: "test",
    Email:    "test@example.com",
}

// 可以设置特定表名（可选）
user.SetTable("custom_users")

// 可以设置特定连接（可选）
user.SetConnection("mysql_connection")

// 保存到数据库
err := user.Save()
```

## 🏷️ TORM标签系统

### 标签语法结构

```go
type Example struct {
    torm.BaseModel
    
    // 基础语法：`torm:"tag1,tag2:value,tag3"`
    Field string `torm:"type:varchar,size:100,unique,index"`
}
```

### 完整标签参考

#### 主键和自增

```go
type Model struct {
    // 主键设置
    ID     int    `torm:"primary_key"`                    // 设为主键
    UserID string `torm:"primary_key,type:varchar,size:32"` // 字符串主键
    
    // 自增设置
    ID   int   `torm:"primary_key,auto_increment"`      // 自增主键
    Code int64 `torm:"auto_increment"`                  // 单独自增
}

// 跨数据库自增适配：
// MySQL:      AUTO_INCREMENT
// PostgreSQL: SERIAL / BIGSERIAL
// SQLite:     AUTOINCREMENT
```

#### 数据类型控制

```go
type TypeExamples struct {
    // 字符串类型
    Name      string `torm:"type:varchar,size:100"`      // VARCHAR(100)
    Code      string `torm:"type:char,size:10"`          // CHAR(10)
    Bio       string `torm:"type:text"`                  // TEXT
    Content   string `torm:"type:longtext"`              // LONGTEXT
    
    // 数值类型
    Age       int     `torm:"type:int"`                  // INT
    BigNum    int64   `torm:"type:bigint"`               // BIGINT
    SmallNum  int16   `torm:"type:smallint"`             // SMALLINT
    TinyNum   int8    `torm:"type:tinyint"`              // TINYINT
    
    // 精度控制
    Price     float64 `torm:"type:decimal,precision:10,scale:2"`  // DECIMAL(10,2)
    Rate      float64 `torm:"type:decimal,precision:5,scale:4"`   // DECIMAL(5,4)
    
    // 布尔类型
    IsActive  bool    `torm:"type:boolean"`              // BOOLEAN
    
    // 日期时间
    BirthDate time.Time `torm:"type:date"`               // DATE
    LoginTime time.Time `torm:"type:datetime"`           // DATETIME
    EventTime time.Time `torm:"type:timestamp"`          // TIMESTAMP
    
    // 二进制和JSON
    Avatar    []byte                 `torm:"type:blob"`  // BLOB
    Settings  map[string]interface{} `torm:"type:json"`  // JSON
    Tags      []string               `torm:"type:json"`  // JSON数组
}
```

#### 约束和默认值

```go
type ConstraintExamples struct {
    // 唯一约束
    Email    string `torm:"type:varchar,size:100,unique"`
    Username string `torm:"type:varchar,size:50,unique"`
    
    // 非空约束
    Name     string `torm:"type:varchar,size:100,not_null"`
    
    // 允许空值（默认行为）
    Phone    string `torm:"type:varchar,size:20,nullable"`
    
    // 默认值
    Status   string  `torm:"type:varchar,size:20,default:active"`
    Age      int     `torm:"type:int,default:0"`
    Balance  float64 `torm:"type:decimal,precision:10,scale:2,default:0.00"`
    IsActive bool    `torm:"type:boolean,default:1"`
    
    // 无符号数值
    Count    int     `torm:"type:int,unsigned"`
    Amount   float64 `torm:"type:decimal,precision:10,scale:2,unsigned"`
}
```

#### 索引系统

```go
type IndexExamples struct {
    // 普通索引
    Category  string `torm:"type:varchar,size:50,index"`
    Status    string `torm:"type:varchar,size:20,index"`
    
    // 自定义索引名
    SearchKey string `torm:"type:varchar,size:100,index:search_idx"`
    
    // 唯一索引
    Email     string `torm:"type:varchar,size:100,unique"`
    Username  string `torm:"type:varchar,size:50,unique"`
    
    // 全文索引
    Title     string `torm:"type:varchar,size:200,fulltext"`
    Content   string `torm:"type:text,fulltext"`
    
    // 空间索引
    Location  string `torm:"type:varchar,size:100,spatial"`
}
```

#### 外键关系

```go
type User struct {
    torm.BaseModel
    ID   int    `torm:"primary_key,auto_increment"`
    Name string `torm:"type:varchar,size:100"`
}

type Post struct {
    torm.BaseModel
    ID     int    `torm:"primary_key,auto_increment"`
    Title  string `torm:"type:varchar,size:200"`
    
    // 外键定义
    UserID int `torm:"type:int,references:users.id,on_delete:cascade,on_update:cascade"`
    
    // 可选的外键（允许NULL）
    CategoryID int `torm:"type:int,references:categories.id,on_delete:set_null"`
}

// 支持的外键操作：
// on_delete: cascade, set_null, restrict, no_action
// on_update: cascade, set_null, restrict, no_action
```

#### 自动时间戳

```go
type TimestampExamples struct {
    torm.BaseModel
    
    // 自动创建时间（INSERT时自动设置）
    CreatedAt time.Time `torm:"auto_create_time"`
    
    // 自动更新时间（INSERT和UPDATE时自动设置）
    UpdatedAt time.Time `torm:"auto_update_time"`
    
    // 自定义时间戳字段
    PublishedAt time.Time `torm:"type:datetime,default:current_timestamp"`
    
    // MySQL特有的ON UPDATE
    ModifiedAt time.Time `torm:"type:timestamp,default:current_timestamp,on_update:current_timestamp"`
}
```

#### 字段注释

```go
type CommentExamples struct {
    torm.BaseModel
    
    ID       int     `torm:"primary_key,auto_increment,comment:主键ID"`
    Username string  `torm:"type:varchar,size:50,unique,comment:用户名"`
    Email    string  `torm:"type:varchar,size:100,unique,comment:邮箱地址"`
    Age      int     `torm:"type:int,unsigned,default:0,comment:年龄"`
    Salary   float64 `torm:"type:decimal,precision:10,scale:2,comment:薪资"`
    Bio      string  `torm:"type:text,comment:个人简介"`
    
    CreatedAt time.Time `torm:"auto_create_time,comment:创建时间"`
    UpdatedAt time.Time `torm:"auto_update_time,comment:更新时间"`
}
```

### 复杂模型示例

```go
// 完整的电商产品模型
type Product struct {
    torm.BaseModel
    
    // 主键
    ID int64 `json:"id" torm:"primary_key,auto_increment,comment:产品ID"`
    
    // 基础信息
    Name        string  `json:"name" torm:"type:varchar,size:200,not_null,comment:产品名称"`
    SKU         string  `json:"sku" torm:"type:varchar,size:50,unique,index,comment:产品编码"`
    Barcode     string  `json:"barcode" torm:"type:varchar,size:50,unique,comment:条形码"`
    
    // 分类和品牌
    CategoryID  int     `json:"category_id" torm:"type:int,references:categories.id,on_delete:cascade,index,comment:分类ID"`
    BrandID     int     `json:"brand_id" torm:"type:int,references:brands.id,on_delete:set_null,index,comment:品牌ID"`
    
    // 价格信息
    Price       float64 `json:"price" torm:"type:decimal,precision:10,scale:2,unsigned,default:0.00,comment:售价"`
    CostPrice   float64 `json:"cost_price" torm:"type:decimal,precision:10,scale:2,unsigned,default:0.00,comment:成本价"`
    
    // 库存信息
    Stock       int     `json:"stock" torm:"type:int,unsigned,default:0,comment:库存数量"`
    MinStock    int     `json:"min_stock" torm:"type:int,unsigned,default:0,comment:最小库存"`
    MaxStock    int     `json:"max_stock" torm:"type:int,unsigned,default:999999,comment:最大库存"`
    
    // 物理属性
    Weight      float64 `json:"weight" torm:"type:decimal,precision:8,scale:3,unsigned,default:0.000,comment:重量(kg)"`
    Dimensions  string  `json:"dimensions" torm:"type:varchar,size:100,comment:尺寸(长x宽x高)"`
    
    // 文本信息
    Description string  `json:"description" torm:"type:text,comment:产品描述"`
    Features    string  `json:"features" torm:"type:longtext,comment:产品特性"`
    
    // 搜索优化
    SearchKeywords string `json:"search_keywords" torm:"type:varchar,size:500,fulltext,comment:搜索关键词"`
    
    // 状态管理
    Status      string  `json:"status" torm:"type:varchar,size:20,default:draft,index,comment:状态"`
    IsActive    bool    `json:"is_active" torm:"type:boolean,default:1,comment:是否启用"`
    IsFeatured  bool    `json:"is_featured" torm:"type:boolean,default:0,index,comment:是否推荐"`
    
    // JSON数据
    Images      []string               `json:"images" torm:"type:json,comment:产品图片"`
    Attributes  map[string]interface{} `json:"attributes" torm:"type:json,comment:产品属性"`
    SEOData     map[string]interface{} `json:"seo_data" torm:"type:json,comment:SEO数据"`
    
    // 时间戳
    CreatedAt   time.Time `json:"created_at" torm:"auto_create_time,comment:创建时间"`
    UpdatedAt   time.Time `json:"updated_at" torm:"auto_update_time,comment:更新时间"`
    PublishedAt time.Time `json:"published_at" torm:"type:datetime,comment:发布时间"`
}
```

## 🔄 自动迁移

### AutoMigrate 核心功能

```go
// 基础自动迁移
func basicAutoMigrate() {
    // 单模型迁移
    user := &User{}
    err := user.AutoMigrate()
    if err != nil {
        log.Fatal(err)
    }
    
    // 多模型迁移（注意顺序：先创建被引用的表）
    dept := &Department{}
    dept.AutoMigrate()
    
    user := &User{}  // User 模型有外键引用 Department
    user.AutoMigrate()
}
```

### 智能增量更新

```go
// 第一版模型
type UserV1 struct {
    torm.BaseModel
    ID   int    `torm:"primary_key,auto_increment"`
    Name string `torm:"type:varchar,size:50"`
}

// 部署第一版
userV1 := &UserV1{}
userV1.AutoMigrate()
// SQL: CREATE TABLE users (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(50))

// 第二版模型 - 添加字段
type UserV2 struct {
    torm.BaseModel
    ID    int    `torm:"primary_key,auto_increment"`
    Name  string `torm:"type:varchar,size:50"`
    Email string `torm:"type:varchar,size:100,unique"`  // 新增
    Age   int    `torm:"type:int,default:0"`            // 新增
}

// 部署第二版
userV2 := &UserV2{}
userV2.AutoMigrate()
// 智能检测差异，只执行必要变更：
// SQL: ALTER TABLE users ADD COLUMN email VARCHAR(100) UNIQUE
// SQL: ALTER TABLE users ADD COLUMN age INT DEFAULT 0

// 第三版模型 - 修改字段
type UserV3 struct {
    torm.BaseModel
    ID    int    `torm:"primary_key,auto_increment"`
    Name  string `torm:"type:varchar,size:100"`        // 长度从50改为100
    Email string `torm:"type:varchar,size:100,unique"`
    Age   int    `torm:"type:int,default:0"`
}

// 部署第三版
userV3 := &UserV3{}
userV3.AutoMigrate()
// 智能检测字段变更：
// SQL: ALTER TABLE users MODIFY COLUMN name VARCHAR(100)
```

### 跨数据库迁移

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

// 同一模型，不同数据库自动适配
func crossDatabaseMigration() {
    user := &User{}
    
    // MySQL 环境
    user.SetConnection("mysql")
    user.AutoMigrate()
    // 生成: CREATE TABLE users (
    //   id INT AUTO_INCREMENT PRIMARY KEY,
    //   username VARCHAR(50) UNIQUE,
    //   balance DECIMAL(10,2),
    //   is_active BOOLEAN DEFAULT 1,
    //   data JSON,
    //   created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    // )
    
    // PostgreSQL 环境
    user.SetConnection("postgres")
    user.AutoMigrate()
    // 生成: CREATE TABLE users (
    //   id SERIAL PRIMARY KEY,
    //   username VARCHAR(50) UNIQUE,
    //   balance DECIMAL(10,2),
    //   is_active BOOLEAN DEFAULT true,
    //   data JSONB,
    //   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    // )
    
    // SQLite 环境
    user.SetConnection("sqlite")
    user.AutoMigrate()
    // 生成: CREATE TABLE users (
    //   id INTEGER PRIMARY KEY AUTOINCREMENT,
    //   username TEXT UNIQUE,
    //   balance REAL,
    //   is_active INTEGER DEFAULT 1,
    //   data TEXT,
    //   created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    // )
}
```

### 批量模型迁移

```go
func batchAutoMigrate() {
    // 定义迁移顺序（先创建被引用的表）
    models := []interface{}{
        &Category{},    // 被Product引用
        &Brand{},       // 被Product引用
        &User{},        // 被Order引用
        &Product{},     // 引用Category和Brand
        &Order{},       // 引用User
        &OrderItem{},   // 引用Order和Product
    }
    
    // 按顺序迁移
    for _, model := range models {
        if migrator, ok := model.(interface{ AutoMigrate() error }); ok {
            if err := migrator.AutoMigrate(); err != nil {
                log.Printf("AutoMigrate失败 %T: %v", model, err)
            } else {
                log.Printf("AutoMigrate成功 %T", model)
            }
        }
    }
}
```

## 📊 模型操作

### 基础CRUD操作

```go
// 创建记录
user := &User{
    Username: "zhangsan",
    Email:    "zhangsan@example.com",
    Age:      25,
    Status:   "active",
    IsActive: true,
}

// 保存到数据库
err := user.Save()
if err != nil {
    log.Printf("保存失败: %v", err)
}

// 查询记录
foundUser := &User{}
foundUser.SetConnection("default") // 可选：设置连接

// 根据条件查询
results, err := foundUser.Where("status", "=", "active").
    Where("age", ">=", 18).
    OrderBy("created_at", "desc").
    Get()

// 查询单条记录
result, err := foundUser.Where("username", "=", "zhangsan").First()

// 更新记录
user.Age = 26
user.Status = "premium"
err = user.Save()

// 删除记录
err = user.Delete()
```

### 高级查询操作

```go
user := &User{}

// 参数化查询
activeUsers, err := user.Where("status = ? AND age >= ?", "active", 18).Get()

// 数组参数查询
premiumUsers, err := user.Where("status IN (?)", []string{"premium", "vip"}).Get()

// 复杂条件
complexResults, err := user.
    Where("(status = ? OR vip_level > ?) AND age BETWEEN ? AND ?", 
          "premium", 3, 18, 65).
    Get()

// 聚合查询
count, err := user.Where("status", "=", "active").Count()

// 分页查询
pagination, err := user.Where("status", "=", "active").
    OrderBy("created_at", "desc").
    Paginate(1, 20)
```

### 模型关联

```go
// 定义关联模型
type User struct {
    torm.BaseModel
    ID   int    `torm:"primary_key,auto_increment"`
    Name string `torm:"type:varchar,size:100"`
}

type Profile struct {
    torm.BaseModel
    ID     int    `torm:"primary_key,auto_increment"`
    UserID int    `torm:"type:int,references:users.id,on_delete:cascade"`
    Avatar string `torm:"type:varchar,size:255"`
    Bio    string `torm:"type:text"`
}

type Post struct {
    torm.BaseModel
    ID     int    `torm:"primary_key,auto_increment"`
    UserID int    `torm:"type:int,references:users.id,on_delete:cascade"`
    Title  string `torm:"type:varchar,size:200"`
    Content string `torm:"type:text"`
}

// 手动关联查询（当前版本）
func getUserWithProfileAndPosts(userID int) {
    // 查询用户
    user := &User{}
    userResult, err := user.Where("id", "=", userID).First()
    
    // 查询用户资料
    profile := &Profile{}
    profileResult, err := profile.Where("user_id", "=", userID).First()
    
    // 查询用户文章
    post := &Post{}
    posts, err := post.Where("user_id", "=", userID).
    OrderBy("created_at", "desc").
    Get()
    
    // 组合结果
    result := map[string]interface{}{
        "user":    userResult,
        "profile": profileResult,
        "posts":   posts,
    }
}
```

## 💡 最佳实践

### 1. 模型设计原则

```go
// ✅ 好的模型设计
type User struct {
    torm.BaseModel
    
    // 明确的主键
    ID int64 `json:"id" torm:"primary_key,auto_increment,comment:用户ID"`
    
    // 有意义的约束
    Username string `json:"username" torm:"type:varchar,size:50,unique,index,comment:用户名"`
    Email    string `json:"email" torm:"type:varchar,size:100,unique,comment:邮箱"`
    
    // 合适的数据类型
    Age      int     `json:"age" torm:"type:int,unsigned,default:0,comment:年龄"`
    Balance  float64 `json:"balance" torm:"type:decimal,precision:10,scale:2,default:0.00,comment:余额"`
    
    // 状态管理
    Status   string `json:"status" torm:"type:varchar,size:20,default:active,index,comment:状态"`
    IsActive bool   `json:"is_active" torm:"type:boolean,default:1,comment:是否启用"`
    
    // 自动时间戳
    CreatedAt time.Time `json:"created_at" torm:"auto_create_time,comment:创建时间"`
    UpdatedAt time.Time `json:"updated_at" torm:"auto_update_time,comment:更新时间"`
}

// ❌ 避免的设计
type BadUser struct {
    torm.BaseModel
    ID       string  `torm:"primary_key"`                    // 没有auto_increment
    Name     string  // 没有type和size，数据库兼容性差
    Money    float64 // 金额用float64精度不够
    Flag     int     // 布尔值用int，语义不清
    Created  string  // 时间用string，失去数据库功能
}
```

### 2. 迁移策略

```go
// ✅ 推荐的迁移策略
func deploymentMigration() {
    // 1. 按依赖顺序迁移
    models := []interface{}{
        &Category{},   // 基础数据
        &User{},       // 用户数据
        &Product{},    // 业务数据（依赖Category）
        &Order{},      // 订单数据（依赖User和Product）
    }
    
    // 2. 错误处理
    for _, model := range models {
        if migrator, ok := model.(interface{ AutoMigrate() error }); ok {
            if err := migrator.AutoMigrate(); err != nil {
                log.Fatalf("迁移失败 %T: %v", model, err)
            }
            log.Printf("✅ 迁移成功: %T", model)
        }
    }
}

// ✅ 环境隔离
func environmentMigration(env string) {
    var connectionName string
    switch env {
    case "development":
        connectionName = "dev"
    case "testing":
        connectionName = "test"
    case "production":
        connectionName = "prod"
    }
    
    user := &User{}
    user.SetConnection(connectionName)
    user.AutoMigrate()
}
```

### 3. 性能优化

```go
type OptimizedUser struct {
    torm.BaseModel
    
    ID int64 `torm:"primary_key,auto_increment"`
    
    // ✅ 为经常查询的字段添加索引
    Username string `torm:"type:varchar,size:50,unique,index"`
    Email    string `torm:"type:varchar,size:100,unique"`
    Status   string `torm:"type:varchar,size:20,index"`
    
    // ✅ 选择合适的数据类型
    Age      int8    `torm:"type:tinyint,unsigned"`      // 年龄用tinyint足够
    Level    int16   `torm:"type:smallint,unsigned"`     // 等级用smallint
    
    // ✅ 合理的字符串长度
    Phone    string  `torm:"type:varchar,size:20"`       // 手机号
    Name     string  `torm:"type:varchar,size:100"`      // 姓名
    
    // ✅ 金额字段使用DECIMAL
    Balance  float64 `torm:"type:decimal,precision:10,scale:2"`
    
    CreatedAt time.Time `torm:"auto_create_time"`
    UpdatedAt time.Time `torm:"auto_update_time"`
}
```

### 4. 错误处理

```go
func safeModelOperations() {
    user := &User{
        Username: "test",
        Email:    "test@example.com",
    }
    
    // ✅ 自动迁移错误处理
    if err := user.AutoMigrate(); err != nil {
        log.Printf("AutoMigrate失败: %v", err)
        return
    }
    
    // ✅ 保存错误处理
    if err := user.Save(); err != nil {
        if strings.Contains(err.Error(), "Duplicate entry") {
            log.Printf("用户已存在: %v", err)
        } else {
            log.Printf("保存失败: %v", err)
        }
        return
    }
    
    // ✅ 查询错误处理
    results, err := user.Where("status", "=", "active").Get()
    if err != nil {
        log.Printf("查询失败: %v", err)
        return
    }
    
    if len(results) == 0 {
        log.Printf("未找到匹配记录")
        return
    }
    
    log.Printf("查询成功，找到 %d 条记录", len(results))
}
```

### 5. 开发工作流

```go
// ✅ 推荐的开发工作流
func developmentWorkflow() {
    // 1. 开发阶段：使用AutoMigrate
    if os.Getenv("APP_ENV") == "development" {
user := &User{}
        user.AutoMigrate()
        
        product := &Product{}
        product.AutoMigrate()
    }
    
    // 2. 测试阶段：确保模型一致性
    if os.Getenv("APP_ENV") == "testing" {
        models := []interface{}{&User{}, &Product{}, &Order{}}
        for _, model := range models {
            if migrator, ok := model.(interface{ AutoMigrate() error }); ok {
                migrator.AutoMigrate()
            }
        }
    }
    
    // 3. 生产阶段：谨慎使用AutoMigrate
    if os.Getenv("APP_ENV") == "production" {
        // 可以使用AutoMigrate，但要有完整的备份和回滚计划
        log.Printf("生产环境，执行AutoMigrate...")
    user := &User{}
        if err := user.AutoMigrate(); err != nil {
            log.Fatalf("生产环境迁移失败: %v", err)
        }
    }
}
```