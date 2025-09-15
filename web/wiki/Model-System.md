# 模型系统

TORM的模型系统基于TORM标签和AutoMigrate功能，让数据库表设计变得简单而精确。

## 🚀 快速开始

### 基础模型定义

```go
package main

import (
    "time"
    "github.com/zhoudm1743/torm/model"
)

// 用户模型 
type User struct {
    model.BaseModel
    
    // 主键和自增
    ID int `json:"id" torm:"primary_key,auto_increment"`
    
    // 字符串类型和约束
    Username string `json:"username" torm:"type:varchar,size:50,unique,index"`
    Email    string `json:"email" torm:"type:varchar,size:100,unique,index:btree"`
    Password string `json:"password" torm:"type:varchar,size:255"`
    
    // 数值类型
    Age    int     `json:"age" torm:"type:int,unsigned,default:0"`
    Salary float64 `json:"salary" torm:"type:decimal,precision:10,scale:2,default:0.00"`
    
    // 状态和布尔
    Status   string `json:"status" torm:"type:varchar,size:20,default:active,index"`
    IsActive bool   `json:"is_active" torm:"type:boolean,default:1"`
    
    // 文本类型
    Bio string `json:"bio" torm:"type:text"`
    
    // 外键关联
    DeptID int `json:"dept_id" torm:"type:int,references:departments.id,on_delete:set_null"`
    
    // 自动时间戳
    CreatedAt time.Time `json:"created_at" torm:"auto_create_time"`
    UpdatedAt time.Time `json:"updated_at" torm:"auto_update_time"`
}

func main() {
    // 配置数据库
    db.AddConnection("default", &db.Config{
        Driver:   "mysql",
        Host:     "localhost",
        Database: "myapp",
        Username: "root",
        Password: "password",
    })
    
    // 自动创建表结构 - 使用 NewModel
    userModel := model.NewModel(&User{})
    userModel.SetConnection("default")
    userModel.AutoMigrate(&User{})
    
    // 开始使用模型
    newUser := model.NewModel("users")
    newUser.SetPrimaryKey("id").SetConnection("default")
    newUser.Fill(map[string]interface{}{
        "username":  "zhangsan",
        "email":     "zhangsan@example.com",
        "age":       25,
        "status":    "active",
        "is_active": true,
        "bio":       "这是用户简介",
    })
    
    // 保存到数据库
    newUser.Save()
}
```

## 🏗️ BaseModel基础

### BaseModel 功能

```go
// BaseModel 提供的核心功能
type User struct {
    model.BaseModel  // 继承基础功能
    // ... 你的字段
}

// BaseModel 提供的方法：
// - Save() error                               // 保存模型
// - Delete() error                             // 删除模型  
// - AutoMigrate(models ...interface{}) error  // 自动迁移
// - SetTable(name string) *BaseModel          // 设置表名
// - SetConnection(name string) *BaseModel     // 设置连接
// - SetPrimaryKey(key string) *BaseModel      // 设置主键
// - SetAttribute(key string, value interface{}) *BaseModel // 设置属性
// - GetAttribute(key string) interface{}      // 获取属性
// - SetAttributes(attrs map[string]interface{}) *BaseModel // 批量设置属性
// - GetAttributes() map[string]interface{}    // 获取所有属性
// - Fill(data map[string]interface{}) *BaseModel // 填充数据
// - Where(conditions...) *db.QueryBuilder    // 条件查询
// - OrderBy(column, direction string) *db.QueryBuilder // 排序
// - Find(id interface{}) error               // 根据主键查找
// - FindByPK(pk interface{}) error           // 根据主键查找
// - IsNew() bool                             // 是否新记录
// - IsExists() bool                          // 是否已存在
// - MarkAsNew() *BaseModel                   // 标记为新记录
// - MarkAsExists() *BaseModel                // 标记为已存在
// - GetKey() interface{}                     // 获取主键值
// - SetKey(key interface{}) *BaseModel       // 设置主键值
// - ToJSON() (string, error)                 // 转为JSON
// - FromJSON(jsonStr string) error           // 从JSON加载
// - ToMap() map[string]interface{}           // 转为Map
// - ClearAttributes() *BaseModel             // 清空属性
```

### 模型初始化

```go
// 创建新的用户实例 - 多种方式
// 方式1: 直接指定表名
user := model.NewModel("users")
user.SetPrimaryKey("id").SetConnection("default")

// 方式2: 从结构体自动解析（推荐）
userModel := model.NewModel(&User{})
userModel.SetConnection("default")

// 方式3: 指定表名和连接
user := model.NewModel("users")
user.SetConnection("mysql_connection")

// 使用 Fill 方法填充数据
user.Fill(map[string]interface{}{
    "username": "test",
    "email":    "test@example.com",
    "age":      25,
    "status":   "active",
})

// 或者使用 SetAttribute 逐个设置
user.SetAttribute("username", "test")
user.SetAttribute("email", "test@example.com")

// 保存到数据库
err := user.Save()

// 查询示例
foundUser := model.NewModel("users")
foundUser.SetConnection("default")
err = foundUser.Find(1) // 根据主键查找

// 获取属性值
username := foundUser.GetAttribute("username")
email := foundUser.GetAttribute("email")
```

## 🏷️ TORM标签系统

### 标签语法结构

```go
type Example struct {
    model.BaseModel
    
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
    model.BaseModel
    ID   int    `torm:"primary_key,auto_increment"`
    Name string `torm:"type:varchar,size:100"`
}

type Post struct {
    model.BaseModel
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
    model.BaseModel
    
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
    model.BaseModel
    
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

### 实际测试案例

```go
// 基于实际模型定义
import (
    "time"
    "github.com/zhoudm1743/torm/model"
)

// 部门模型
type Department struct {
    model.BaseModel
    ID        int       `json:"id" torm:"primary_key,auto_increment"`
    Name      string    `json:"name" torm:"type:varchar,size:100,unique"`
    Budget    float64   `json:"budget" torm:"type:decimal,precision:12,scale:2,default:0.00"`
    Location  string    `json:"location" torm:"type:varchar,size:255"`
    IsActive  bool      `json:"is_active" torm:"type:boolean,default:1"`
    CreatedAt time.Time `json:"created_at" torm:"auto_create_time"`
    UpdatedAt time.Time `json:"updated_at" torm:"auto_update_time"`
}

// 用户模型（包含外键关联）
type User struct {
    model.BaseModel
    ID        int       `json:"id" torm:"primary_key,auto_increment"`
    Username  string    `json:"username" torm:"type:varchar,size:50,unique,index"`
    Email     string    `json:"email" torm:"type:varchar,size:100,unique,index:btree"`
    Password  string    `json:"password" torm:"type:varchar,size:255"`
    Age       int       `json:"age" torm:"type:int,unsigned,default:0"`
    Salary    float64   `json:"salary" torm:"type:decimal,precision:10,scale:2,default:0.00"`
    Status    string    `json:"status" torm:"type:varchar,size:20,default:active,index"`
    Bio       string    `json:"bio" torm:"type:text"`
    IsActive  bool      `json:"is_active" torm:"type:boolean,default:1"`
    DeptID    int       `json:"dept_id" torm:"type:int,references:departments.id,on_delete:set_null"`
    CreatedAt time.Time `json:"created_at" torm:"auto_create_time"`
    UpdatedAt time.Time `json:"updated_at" torm:"auto_update_time"`
}

// 角色模型（用于多对多关联）
type Role struct {
    model.BaseModel
    ID          int       `json:"id" torm:"primary_key,auto_increment"`
    Name        string    `json:"name" torm:"type:varchar,size:50,unique"`
    Description string    `json:"description" torm:"type:text"`
    IsActive    bool      `json:"is_active" torm:"type:boolean,default:1"`
    CreatedAt   time.Time `json:"created_at" torm:"auto_create_time"`
    UpdatedAt   time.Time `json:"updated_at" torm:"auto_update_time"`
}

// 项目模型（多外键关联）
type Project struct {
    model.BaseModel
    ID          int       `json:"id" torm:"primary_key,auto_increment"`
    Name        string    `json:"name" torm:"type:varchar,size:100"`
    Description string    `json:"description" torm:"type:text"`
    UserID      int       `json:"user_id" torm:"type:int,references:users.id"`
    DeptID      int       `json:"dept_id" torm:"type:int,references:departments.id"`
    Status      string    `json:"status" torm:"type:varchar,size:20,default:active"`
    CreatedAt   time.Time `json:"created_at" torm:"auto_create_time"`
    UpdatedAt   time.Time `json:"updated_at" torm:"auto_update_time"`
}
```

### 模型关联定义

```go
// User 关联方法定义
func (u *User) Department() *model.BelongsTo {
    dept := &Department{}
    return u.BelongsTo(dept, "dept_id", "id")
}

func (u *User) Projects() *model.HasMany {
    project := &Project{}
    return u.HasMany(project, "user_id", "id")
}

func (u *User) Roles() *model.BelongsToMany {
    role := &Role{}
    return u.BelongsToMany(role, "user_roles", "role_id", "user_id")
}

// Department 关联方法定义
func (d *Department) Users() *model.HasMany {
    user := &User{}
    return d.HasMany(user, "dept_id", "id")
}

func (d *Department) Projects() *model.HasMany {
    project := &Project{}
    return d.HasMany(project, "dept_id", "id")
}

// Project 关联方法定义
func (p *Project) User() *model.BelongsTo {
    user := &User{}
    return p.BelongsTo(user, "user_id", "id")
}

func (p *Project) Department() *model.BelongsTo {
    dept := &Department{}
    return p.BelongsTo(dept, "dept_id", "id")
}

// Role 关联方法定义
func (r *Role) Users() *model.BelongsToMany {
    user := &User{}
    return r.BelongsToMany(user, "user_roles", "user_id", "role_id")
}
```

## 🔄 自动迁移

### AutoMigrate 核心功能

```go
// 基础自动迁移 - 基于实际测试案例
func basicAutoMigrate() {
    // 单模型迁移 - 使用 NewModel
    deptModel := model.NewModel(&Department{})
    deptModel.SetConnection("default")
    err := deptModel.AutoMigrate(&Department{})
    if err != nil {
        log.Fatal(err)
    }
    
    // 多模型迁移（注意顺序：先创建被引用的表）
    userModel := model.NewModel(&User{})
    userModel.SetConnection("default")
    err = userModel.AutoMigrate(&User{})  // User 模型有外键引用 Department
    if err != nil {
        log.Fatal(err)
    }
    
    // 多表一次性迁移
    err = userModel.AutoMigrate(&User{}, &Department{})
    if err != nil {
        log.Fatal(err)
    }
    
    log.Println("✅ 自动迁移完成")
}
```

### 智能增量更新

```go
// 第一版模型
type UserV1 struct {
    model.BaseModel
    ID   int    `torm:"primary_key,auto_increment"`
    Name string `torm:"type:varchar,size:50"`
}

// 部署第一版
userV1 := &UserV1{}
userV1.AutoMigrate()
// SQL: CREATE TABLE users (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(50))

// 第二版模型 - 添加字段
type UserV2 struct {
    model.BaseModel
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
    model.BaseModel
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
    model.BaseModel
    ID       int       `torm:"primary_key,auto_increment"`
    Username string    `torm:"type:varchar,size:50,unique"`
    Balance  float64   `torm:"type:decimal,precision:10,scale:2"`
    IsActive bool      `torm:"type:boolean,default:1"`
    Data     map[string]interface{} `torm:"type:json"`
    CreatedAt time.Time `torm:"auto_create_time"`
}

// 同一模型，不同数据库自动适配
func crossDatabaseMigration() {
    // MySQL 环境
    mysqlModel := model.NewModel(&User{})
    mysqlModel.SetConnection("mysql")
    mysqlModel.AutoMigrate(&User{})
    // 生成: CREATE TABLE users (
    //   id INT AUTO_INCREMENT PRIMARY KEY,
    //   username VARCHAR(50) UNIQUE,
    //   balance DECIMAL(10,2),
    //   is_active BOOLEAN DEFAULT 1,
    //   data JSON,
    //   created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    // )
    
    // PostgreSQL 环境
    pgModel := model.NewModel(&User{})
    pgModel.SetConnection("postgres")
    pgModel.AutoMigrate(&User{})
    // 生成: CREATE TABLE users (
    //   id SERIAL PRIMARY KEY,
    //   username VARCHAR(50) UNIQUE,
    //   balance DECIMAL(10,2),
    //   is_active BOOLEAN DEFAULT true,
    //   data JSONB,
    //   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    // )
    
    // SQLite 环境
    sqliteModel := model.NewModel(&User{})
    sqliteModel.SetConnection("sqlite")
    sqliteModel.AutoMigrate(&User{})
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
// 创建记录 - 基于实际测试案例
user := model.NewModel("users")
user.SetPrimaryKey("id").SetConnection("default")

// 使用 Fill 方法批量填充数据
user.Fill(map[string]interface{}{
    "username":  "zhangsan",
    "email":     "zhangsan@example.com",
    "password":  "password123",
    "age":       25,
    "salary":    15000.50,
    "status":    "active",
    "bio":       "这是用户简介",
    "is_active": true,
    "dept_id":   1,
})

// 保存到数据库
err := user.Save()
if err != nil {
    log.Printf("保存失败: %v", err)
}

// 查询记录
foundUser := model.NewModel("users")
foundUser.SetPrimaryKey("id").SetConnection("default")

// 根据主键查找
err = foundUser.Find(user.GetKey())
if err != nil {
    log.Printf("查询失败: %v", err)
}

// 获取属性值
username := foundUser.GetAttribute("username")
email := foundUser.GetAttribute("email")
log.Printf("用户查询成功: %s (email: %s)", username, email)

// 更新记录 - 使用 SetAttribute
foundUser.SetAttribute("salary", 18000.00)
foundUser.SetAttribute("status", "promoted")
err = foundUser.Save()
if err != nil {
    log.Printf("更新失败: %v", err)
}

// 验证更新结果
salary := foundUser.GetAttribute("salary")
status := foundUser.GetAttribute("status")
log.Printf("用户更新成功: salary=%v, status=%s", salary, status)

// 删除记录
err = foundUser.Delete()
```

### 高级查询操作

```go
// 创建查询模型
userModel := model.NewModel("users")
userModel.SetConnection("default")

// 参数化查询 - 通过 Query() 获取查询构建器
query, err := userModel.Query()
if err != nil {
    log.Fatal(err)
}

activeUsers, err := query.Where("status = ? AND age >= ?", "active", 18).GetRaw()

// 数组参数查询
query2, _ := userModel.Query()
premiumUsers, err := query2.Where("status IN (?)", []string{"premium", "vip"}).GetRaw()

// 复杂条件
query3, _ := userModel.Query()
complexResults, err := query3.
    Where("(status = ? OR vip_level > ?) AND age BETWEEN ? AND ?", 
          "premium", 3, 18, 65).
    GetRaw()

// 聚合查询
query4, _ := userModel.Query()
count, err := query4.Where("status", "=", "active").Count()

// 分页查询
query5, _ := userModel.Query()
pagination, err := query5.Where("status", "=", "active").
    OrderBy("created_at", "desc").
    Paginate(1, 20)
```

### 模型属性操作（基于实际测试案例）

```go
// 模型属性功能测试
func testModelAttributes() {
    // 1. 创建模型实例
    user := model.NewModel("users")
    user.SetPrimaryKey("id").SetConnection("default")
    
    // 2. 测试 SetAttribute 和 GetAttribute
    // 单个属性设置
    user.SetAttribute("username", "test_attr_user")
    user.SetAttribute("age", 30)
    user.SetAttribute("salary", 25000.50)
    user.SetAttribute("is_active", true)
    
    // 验证获取属性
    if username := user.GetAttribute("username"); username == "test_attr_user" {
        log.Println("✅ SetAttribute/GetAttribute 测试成功")
    }
    
    // 3. 测试 SetAttributes 批量设置
    batchData := map[string]interface{}{
        "email":    "batch_test@example.com",
        "password": "batch_password",
        "status":   "batch_active",
        "bio":      "这是批量设置的测试用户",
    }
    user.SetAttributes(batchData)
    
    // 验证批量设置的属性
    for key, expectedValue := range batchData {
        if actualValue := user.GetAttribute(key); actualValue == expectedValue {
            log.Printf("✅ SetAttributes 成功: %s = %v", key, actualValue)
        }
    }
    
    // 4. 测试 GetAttributes 获取所有属性
    allAttributes := user.GetAttributes()
    log.Printf("✅ GetAttributes 成功: 共 %d 个属性", len(allAttributes))
    
    // 5. 测试 ClearAttributes 清空属性
    user.ClearAttributes()
    clearedAttributes := user.GetAttributes()
    if len(clearedAttributes) == 0 {
        log.Println("✅ ClearAttributes 成功")
    }
    
    // 6. 测试 Fill 方法
    fillData := map[string]interface{}{
        "username":  "fill_user",
        "email":     "fill@example.com",
        "age":       35,
        "salary":    30000.00,
        "status":    "active",
        "is_active": true,
    }
    user.Fill(fillData)
    log.Println("✅ Fill 方法测试成功")
    
    // 7. 测试 GetKey 和 SetKey
    user.SetKey(12345)
    if key := user.GetKey(); key == 12345 {
        log.Println("✅ SetKey/GetKey 测试成功")
    }
    
    // 8. 测试 ToJSON 和 FromJSON
    jsonStr, err := user.ToJSON()
    if err == nil {
        newUser := model.NewModel("users")
        newUser.SetPrimaryKey("id").SetConnection("default")
        
        err = newUser.FromJSON(jsonStr)
        if err == nil && newUser.GetAttribute("username") == user.GetAttribute("username") {
            log.Println("✅ ToJSON/FromJSON 测试成功")
        }
    }
    
    // 9. 测试状态方法
    if user.IsNew() {
        log.Println("✅ IsNew: 新模型是新记录")
    }
    
    user.MarkAsExists()
    if !user.IsNew() && user.IsExists() {
        log.Println("✅ MarkAsExists: 标记为已存在")
    }
    
    user.MarkAsNew()
    if user.IsNew() {
        log.Println("✅ MarkAsNew: 标记为新记录")
    }
}
```

### 模型关联操作（基于实际测试案例）

```go
// 模型关联功能测试
func testModelRelationships() {
    // 1. 创建测试数据
    dept := model.NewModel("departments")
    dept.SetConnection("default")
    dept.Fill(map[string]interface{}{
        "name":      "技术部",
        "budget":    1000000.50,
        "location":  "北京",
        "is_active": true,
    })
    dept.Save()
    
    user := model.NewModel("users")
    user.SetPrimaryKey("id").SetConnection("default")
    user.Fill(map[string]interface{}{
        "username":  "test_user",
        "email":     "test@example.com",
        "dept_id":   dept.GetKey(),
    })
    user.Save()
    
    // 2. 测试 BelongsTo 关联（用户所属部门）
    testUser := model.NewModel("users")
    testUser.SetPrimaryKey("id").SetConnection("default")
    testUser.SetAttribute("id", user.GetKey())
    testUser.SetAttribute("dept_id", dept.GetKey())
    testUser.MarkAsExists()
    
    // 查询用户所属部门
    deptResult, err := testUser.Department().First()
    if err == nil && deptResult != nil {
        log.Printf("✅ BelongsTo 关联成功: 用户所属部门 %v", deptResult["name"])
    }
    
    // 3. 测试 HasMany 关联（部门下的用户）
    testDept := model.NewModel("departments")
    testDept.SetPrimaryKey("id").SetConnection("default")
    testDept.SetAttribute("id", dept.GetKey())
    testDept.MarkAsExists()
    
    // 查询部门下的所有用户
    deptUsers, err := testDept.Users().Get()
    if err == nil {
        log.Printf("✅ HasMany 关联成功: 部门有 %d 个用户", len(deptUsers))
    }
    
    // 4. 测试关联查询的链式调用
    activeUsers, err := testDept.Users().Where("status", "=", "active").Get()
    if err == nil {
        log.Printf("✅ 关联链式调用成功: 活跃用户 %d 个", len(activeUsers))
    }
    
    // 5. 测试关联查询的排序和限制
    limitedUsers, err := testDept.Users().OrderBy("created_at", "DESC").Limit(2).Get()
    if err == nil {
        log.Printf("✅ 关联排序限制成功: 获取 %d 个用户（限制2个）", len(limitedUsers))
    }
}
```

## 💡 最佳实践

### 1. 模型设计原则

```go
// ✅ 好的模型设计（基于实际测试案例）
type User struct {
    model.BaseModel
    
    // 明确的主键
    ID int `json:"id" torm:"primary_key,auto_increment"`
    
    // 有意义的约束和索引
    Username string `json:"username" torm:"type:varchar,size:50,unique,index"`
    Email    string `json:"email" torm:"type:varchar,size:100,unique,index:btree"`
    Password string `json:"password" torm:"type:varchar,size:255"`
    
    // 合适的数据类型
    Age      int     `json:"age" torm:"type:int,unsigned,default:0"`
    Salary   float64 `json:"salary" torm:"type:decimal,precision:10,scale:2,default:0.00"`
    
    // 状态管理
    Status   string `json:"status" torm:"type:varchar,size:20,default:active,index"`
    IsActive bool   `json:"is_active" torm:"type:boolean,default:1"`
    
    // 文本字段
    Bio      string `json:"bio" torm:"type:text"`
    
    // 外键关联
    DeptID   int    `json:"dept_id" torm:"type:int,references:departments.id,on_delete:set_null"`
    
    // 自动时间戳
    CreatedAt time.Time `json:"created_at" torm:"auto_create_time"`
    UpdatedAt time.Time `json:"updated_at" torm:"auto_update_time"`
}

// ✅ 推荐的模型初始化方式
func createUser() {
    // 使用 NewModel 进行迁移
    userModel := model.NewModel(&User{})
    userModel.SetConnection("default")
    userModel.AutoMigrate(&User{})
    
    // 使用 NewModel 进行操作
    newUser := model.NewModel("users")
    newUser.SetPrimaryKey("id").SetConnection("default")
    
    // 使用 Fill 填充数据
    newUser.Fill(map[string]interface{}{
        "username": "test",
        "email":    "test@example.com",
        "age":      25,
    })
    newUser.Save()
}

// ❌ 避免的设计
type BadUser struct {
    model.BaseModel
    ID       string  `torm:"primary_key"`                    // 没有auto_increment
    Name     string  // 没有type和size，数据库兼容性差
    Money    float64 // 金额用float64精度不够
    Flag     int     // 布尔值用int，语义不清
    Created  string  // 时间用string，失去数据库功能
}
```

### 2. 迁移策略

```go
// ✅ 推荐的迁移策略（基于实际测试案例）
func deploymentMigration() {
    // 1. 按依赖顺序迁移 - 先创建被引用的表
    
    // 创建部门表（被用户表引用）
    deptModel := model.NewModel(&Department{})
    deptModel.SetConnection("default")
    if err := deptModel.AutoMigrate(&Department{}); err != nil {
        log.Fatalf("部门表迁移失败: %v", err)
    }
    log.Println("✅ 部门表迁移成功")
    
    // 创建用户表（引用部门表）
    userModel := model.NewModel(&User{})
    userModel.SetConnection("default")
    if err := userModel.AutoMigrate(&User{}); err != nil {
        log.Fatalf("用户表迁移失败: %v", err)
    }
    log.Println("✅ 用户表迁移成功")
    
    // 创建其他表
    roleModel := model.NewModel(&Role{})
    roleModel.SetConnection("default")
    if err := roleModel.AutoMigrate(&Role{}); err != nil {
        log.Fatalf("角色表迁移失败: %v", err)
    }
    
    projectModel := model.NewModel(&Project{})
    projectModel.SetConnection("default")
    if err := projectModel.AutoMigrate(&Project{}); err != nil {
        log.Fatalf("项目表迁移失败: %v", err)
    }
    
    // 2. 多表一次性迁移（推荐）
    if err := userModel.AutoMigrate(&User{}, &Department{}, &Role{}, &Project{}); err != nil {
        log.Fatalf("多表迁移失败: %v", err)
    }
    log.Println("✅ 多表迁移成功")
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
    
    userModel := model.NewModel(&User{})
    userModel.SetConnection(connectionName)
    userModel.AutoMigrate(&User{})
}
```

### 3. 性能优化

```go
type OptimizedUser struct {
    model.BaseModel
    
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
// ✅ 安全的模型操作（基于实际测试案例）
func safeModelOperations() {
    // 1. 初始化模型
    user := model.NewModel("users")
    user.SetPrimaryKey("id").SetConnection("default")
    
    // 2. 自动迁移错误处理
    migrationModel := model.NewModel(&User{})
    migrationModel.SetConnection("default")
    if err := migrationModel.AutoMigrate(&User{}); err != nil {
        log.Printf("AutoMigrate失败: %v", err)
        return
    }
    log.Println("✅ 自动迁移成功")
    
    // 3. 数据填充
    user.Fill(map[string]interface{}{
        "username": "test",
        "email":    "test@example.com",
        "age":      25,
        "status":   "active",
    })
    
    // 4. 保存错误处理
    if err := user.Save(); err != nil {
        if strings.Contains(err.Error(), "Duplicate entry") {
            log.Printf("用户已存在: %v", err)
        } else {
            log.Printf("保存失败: %v", err)
        }
        return
    }
    log.Printf("✅ 用户创建成功, ID: %v", user.GetKey())
    
    // 5. 查询错误处理
    queryUser := model.NewModel("users")
    queryUser.SetConnection("default")
    
    results, err := queryUser.Where("status", "=", "active").Get()
    if err != nil {
        log.Printf("查询失败: %v", err)
        return
    }
    
    if len(results) == 0 {
        log.Printf("未找到匹配记录")
        return
    }
    
    log.Printf("查询成功，找到 %d 条记录", len(results))
    
    // 6. 属性获取和验证
    foundUser := model.NewModel("users")
    foundUser.SetConnection("default")
    
    if err := foundUser.Find(user.GetKey()); err != nil {
        log.Printf("根据主键查找失败: %v", err)
        return
    }
    
    username := foundUser.GetAttribute("username")
    email := foundUser.GetAttribute("email")
    log.Printf("✅ 用户查询成功: %s (email: %s)", username, email)
    
    // 7. 更新操作
    foundUser.SetAttribute("status", "premium")
    if err := foundUser.Save(); err != nil {
        log.Printf("更新失败: %v", err)
        return
    }
    log.Printf("✅ 用户更新成功: status=%s", foundUser.GetAttribute("status"))
}
```

### 5. 开发工作流

```go
// ✅ 推荐的开发工作流
func developmentWorkflow() {
	// 1. 开发阶段：使用AutoMigrate
	if os.Getenv("APP_ENV") == "development" {
		userModel := model.NewModel(&User{})
		userModel.SetConnection("default")
		userModel.AutoMigrate(&User{})
		
		deptModel := model.NewModel(&Department{})
		deptModel.SetConnection("default")
		deptModel.AutoMigrate(&Department{})
	}
	
	// 2. 测试阶段：确保模型一致性
	if os.Getenv("APP_ENV") == "testing" {
		// 按顺序迁移测试表
		deptModel := model.NewModel(&Department{})
		deptModel.SetConnection("test")
		deptModel.AutoMigrate(&Department{})
		
		userModel := model.NewModel(&User{})
		userModel.SetConnection("test")
		userModel.AutoMigrate(&User{})
		
		roleModel := model.NewModel(&Role{})
		roleModel.SetConnection("test")
		roleModel.AutoMigrate(&Role{})
	}
	
	// 3. 生产阶段：谨慎使用AutoMigrate
	if os.Getenv("APP_ENV") == "production" {
		// 可以使用AutoMigrate，但要有完整的备份和回滚计划
		log.Printf("生产环境，执行AutoMigrate...")
		
		userModel := model.NewModel(&User{})
		userModel.SetConnection("production")
		if err := userModel.AutoMigrate(&User{}); err != nil {
			log.Fatalf("生产环境迁移失败: %v", err)
		}
		log.Println("✅ 生产环境迁移成功")
	}
}
```

## 🎨 访问器系统 (Accessor System)

TORM 提供了强大的属性访问器（Accessor）和修改器（Mutator）系统，类似于 ThinkPHP 的模型访问器，但更加强大和灵活。

### 🚀 访问器基础

#### 基本概念

```go
// 访问器 (Accessor): 在获取属性时自动调用，用于格式化显示数据
// 命名规则：Get[AttributeName]Attr
func (u *User) GetStatusAttr(value interface{}) interface{} {
    // value 是数据库中的原始值
    // 返回值是格式化后的显示值
}

// 修改器 (Mutator): 在设置属性时自动调用，用于格式化存储数据  
// 命名规则：Set[AttributeName]Attr
func (u *User) SetStatusAttr(value interface{}) interface{} {
    // value 是输入的值
    // 返回值是要存储到数据库的值
}
```

#### 智能命名转换

TORM 支持复杂的命名转换，完美处理各种缩写和连续大写字母：

```go
// 蛇形命名 -> 访问器方法名的转换规则：
// user_id        -> GetUserIDAttr      (ID 特殊处理)
// icbc_card_no   -> GetICBCCardNoAttr  (ICBC 银行代码)
// db_link_url    -> GetDBLinkURLAttr   (连续缩写处理)
// html_parser    -> GetHTMLParserAttr  (HTML 技术缩写)
// api_version    -> GetAPIVersionAttr  (API 接口缩写)
// json_config    -> GetJSONConfigAttr  (JSON 格式缩写)
// xml_data       -> GetXMLDataAttr     (XML 格式缩写)
// sql_query      -> GetSQLQueryAttr    (SQL 查询缩写)
// ip_address     -> GetIPAddressAttr   (IP 网络缩写)
// uuid_token     -> GetUUIDTokenAttr   (UUID 标识符)
// md5_hash       -> GetMD5HashAttr     (MD5 哈希算法)
```

### 🎯 实际应用案例

#### 用户状态管理

```go
type User struct {
    model.BaseModel
    ID     int    `json:"id" torm:"primary_key,auto_increment"`
    Status int    `json:"status" torm:"type:int,default:1"`        // 0=禁用, 1=正常, 2=待审核
    Gender int    `json:"gender" torm:"type:int,default:1"`        // 0=女, 1=男, 2=其他
}

// 状态访问器 - 将数字转换为可读状态
func (u *User) GetStatusAttr(value interface{}) interface{} {
    status := convertToInt(value)
    statusMap := map[int]map[string]interface{}{
        0: {"code": 0, "name": "已禁用", "color": "red", "can_login": false},
        1: {"code": 1, "name": "正常", "color": "green", "can_login": true},
        2: {"code": 2, "name": "待审核", "color": "orange", "can_login": false},
    }
    return statusMap[status]
}

// 状态修改器 - 支持多种输入格式
func (u *User) SetStatusAttr(value interface{}) interface{} {
    if str, ok := value.(string); ok {
        switch str {
        case "禁用", "disabled": return 0
        case "正常", "active":   return 1
        case "待审核", "pending": return 2
        }
    }
    return convertToInt(value)
}

// 性别访问器 - 丰富的性别信息
func (u *User) GetGenderAttr(value interface{}) interface{} {
    gender := convertToInt(value)
    return map[string]interface{}{
        "code":   gender,
        "name":   []string{"女士", "先生", "其他"}[min(gender, 2)],
        "symbol": []string{"♀", "♂", "⚥"}[min(gender, 2)],
        "color":  []string{"#ff69b4", "#4169e1", "#9370db"}[min(gender, 2)],
    }
}
```

#### 银行卡信息处理

```go
type BankUser struct {
    model.BaseModel
    ICBCCardNo string `json:"icbc_card_no" torm:"type:varchar,size:20"`
    Balance    int    `json:"balance" torm:"type:int,default:0"`        // 以分为单位
}

// ICBC银行卡访问器 - 自动脱敏处理
func (u *BankUser) GetICBCCardNoAttr(value interface{}) interface{} {
    cardNo := fmt.Sprintf("%v", value)
    if len(cardNo) >= 8 {
        return map[string]interface{}{
            "number":     cardNo,
            "masked":     cardNo[:4] + "****" + cardNo[len(cardNo)-4:],
            "bank":       "中国工商银行",
            "is_valid":   len(cardNo) >= 16,
            "card_type":  getCardType(cardNo),
        }
    }
    return cardNo
}

// ICBC银行卡修改器 - 自动清理格式
func (u *BankUser) SetICBCCardNoAttr(value interface{}) interface{} {
    cardNo := fmt.Sprintf("%v", value)
    // 移除所有非数字字符
    var result strings.Builder
    for _, r := range cardNo {
        if r >= '0' && r <= '9' {
            result.WriteRune(r)
        }
    }
    return result.String()
}

// 余额访问器 - 智能金额格式化
func (u *BankUser) GetBalanceAttr(value interface{}) interface{} {
    cents := convertToInt(value)
    yuan := float64(cents) / 100.0
    
    return map[string]interface{}{
        "cents":       cents,
        "yuan":        yuan,
        "formatted":   fmt.Sprintf("¥%.2f", yuan),
        "level":       getBalanceLevel(yuan),
        "is_positive": cents > 0,
    }
}
```

#### 技术字段处理

```go
type TechUser struct {
    model.BaseModel
    APIVersion string `json:"api_version" torm:"type:varchar,size:20"`
    JSONConfig string `json:"json_config" torm:"type:text"`
    XMLData    string `json:"xml_data" torm:"type:text"`
    SQLQuery   string `json:"sql_query" torm:"type:text"`
    IPAddress  string `json:"ip_address" torm:"type:varchar,size:45"`
}

// API版本访问器
func (u *TechUser) GetAPIVersionAttr(value interface{}) interface{} {
    version := fmt.Sprintf("%v", value)
    return map[string]interface{}{
        "version":    version,
        "is_latest":  version == "v2.0",
        "changelog":  fmt.Sprintf("API %s 变更日志", version),
        "docs_url":   fmt.Sprintf("https://api.docs.com/%s", version),
    }
}

// JSON配置访问器 - 自动解析验证
func (u *TechUser) GetJSONConfigAttr(value interface{}) interface{} {
    configStr := fmt.Sprintf("%v", value)
    var config map[string]interface{}
    
    if err := json.Unmarshal([]byte(configStr), &config); err == nil {
        return map[string]interface{}{
            "config":      config,
            "is_valid":    true,
            "format":      "JSON",
            "size_bytes":  len(configStr),
        }
    }
    
    return map[string]interface{}{
        "raw_value": configStr,
        "is_valid":  false,
        "error":     "JSON格式错误",
    }
}

// IP地址访问器 - 地理位置和安全检查
func (u *TechUser) GetIPAddressAttr(value interface{}) interface{} {
    ip := fmt.Sprintf("%v", value)
    
    ipType := "公网IP"
    location := "未知地区"
    
    if strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") {
        ipType = "内网IP"
        location = "局域网"
    }
    
    return map[string]interface{}{
        "ip":       ip,
        "type":     ipType,
        "location": location,
        "is_safe":  !strings.Contains(ip, "malicious"),
        "country":  "中国",
    }
}
```

### 📊 访问器系统

TORM 提供了强大的访问器系统，支持在查询时自动应用 Get/Set 访问器：

#### 访问器查询

```go
// 查询时自动应用访问器（返回原生 []map[string]interface{}）
users, err := torm.Table("users").Model(&User{}).Where("status", "=", 1).Get()

// 数据已经过访问器处理
for _, user := range users {
    status := user["status"]          // 返回: {"code": 1, "name": "正常", ...}
    gender := user["gender"]          // 返回: {"code": 1, "name": "先生", ...}
}

// 查询第一条记录
user, err := torm.Table("users").Model(&User{}).First()
if err == nil && user != nil {
    status := user["status"]          // 自动应用访问器
    gender := user["gender"]          // 自动应用访问器
}

// 原始数据查询（不应用访问器）
rawUsers, err := torm.Table("users").Where("status", "=", 1).GetRaw()
for _, user := range rawUsers {
    status := user["status"]          // 返回: 1 (原始值)
    gender := user["gender"]          // 返回: 1 (原始值)
}
```

#### 设置器使用

```go
// 通过模型设置数据（自动应用设置器）
user := &User{}
user.SetAttributeWithAccessor(user, "status", "正常")      // 自动转换为 1 存储
user.SetAttributeWithAccessor(user, "icbc_card_no", "6222-0212-3456-7890") // 自动清理

// 批量设置
data := map[string]interface{}{
    "status": "正常",
    "icbc_card_no": "6222-0212-3456-7890",
}
user.SetAttributesWithAccessor(user, data)

// 查看设置后的值
storedStatus := user.GetAttribute("status")          // 1
storedCardNo := user.GetAttribute("icbc_card_no")    // "6222021234567890"
```

#### 数据操作

```go
// 查询多条记录 - 直接使用 Model() 方法，自动获取表名
users, err := torm.Model(&User{}).Where("status", "=", 1).Get()

// 基本操作
count := len(users)                   // 记录总数
isEmpty := len(users) == 0            // 是否为空

// 遍历记录
for i, user := range users {
    username := user["username"]
    status := user["status"]
    fmt.Printf("[%d] %s: %v\n", i, username, status)
}

// JSON 输出
accessorJSON, _ := json.Marshal(users)    // 包含访问器处理的完整JSON
```

### 🔧 高级特性

#### 自动[]byte处理

```go
// TORM 自动处理数据库返回的 []byte 数据
testData := map[string]interface{}{
    "user_id":      []byte("12345"),           // 自动转换为 int: 12345
    "icbc_card_no": []byte("6222021234567890"), // 自动转换为 int64
    "api_version":  []byte("v2.0"),            // 自动转换为 string: "v2.0"
    "balance":      []byte("123456"),          // 自动转换为 int: 123456
    "is_active":    []byte("true"),            // 自动转换为 bool: true
    "created_at":   []byte("2024-01-01 10:00:00"), // 自动转换为时间
    "settings":     []byte(`{"theme":"dark"}`), // 自动解析为 JSON
}

// 访问器处理器会自动处理这些数据类型
processor := db.NewAccessorProcessor(&User{})
processedData := processor.ProcessData(testData)
// 所有访问器都会收到正确类型的处理后数据
```

#### 性能优化

```go
// 访问器缓存机制
// TORM 使用反射缓存和正则匹配优化性能

// 1. 方法发现只在首次调用时进行
// 2. 正则匹配结果会被缓存
// 3. 反射方法调用会被优化

// 性能对比（1000次调用）:
// 原始map访问:     100μs
// 访问器处理:      280μs (2.8x)

// 实际使用建议：
// - 显示数据使用 Model().Get()（自动应用访问器）
// - 计算逻辑使用 GetRaw()（原始数据）
// - 批量处理使用 collection 操作
```

#### 调试和错误处理

```go
// 访问器调试
func (u *User) GetStatusAttr(value interface{}) interface{} {
    // 可以添加日志来调试访问器调用
    log.Printf("GetStatusAttr called with: %v (%T)", value, value)
    
    // 类型安全处理
    status, ok := value.(int)
    if !ok {
        log.Printf("Warning: status value is not int: %v", value)
        return map[string]interface{}{
            "error": "invalid status type",
            "value": value,
        }
    }
    
    // 返回处理结果
    return processStatus(status)
}

// 错误恢复
func (u *User) GetBalanceAttr(value interface{}) interface{} {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Balance accessor panic: %v", r)
        }
    }()
    
    // 安全的访问器逻辑
    return processBalance(value)
}
```

### 💡 最佳实践

#### 1. 访问器设计原则

```go
// ✅ 好的访问器设计
func (u *User) GetStatusAttr(value interface{}) interface{} {
    // 1. 类型安全
    status := convertToInt(value)
    
    // 2. 返回结构化数据
    return map[string]interface{}{
        "code":        status,
        "name":        getStatusName(status),
        "color":       getStatusColor(status),
        "permissions": getStatusPermissions(status),
    }
}

// ❌ 避免的设计
func (u *User) GetStatusAttr(value interface{}) interface{} {
    // 不要直接返回字符串，丢失了结构化信息
    return "正常"
}
```

#### 2. 命名规范

```go
// ✅ 推荐的字段命名（会被正确转换）
type User struct {
    UserID      int    `json:"user_id"`       // -> GetUserIDAttr
    ICBCCardNo  string `json:"icbc_card_no"`  // -> GetICBCCardNoAttr  
    HTMLContent string `json:"html_content"`  // -> GetHTMLContentAttr
    APIKey      string `json:"api_key"`       // -> GetAPIKeyAttr
    JSONData    string `json:"json_data"`     // -> GetJSONDataAttr
}

// ❌ 避免的命名
type User struct {
    userid      int    // 全小写，访问器匹配困难
    HTML_data   string // 混合命名风格
    api_Key     string // 不一致的大小写
}
```

#### 3. 数据类型选择

```go
// ✅ 合适的数据库字段类型
type User struct {
    Balance     int     `torm:"type:int"`                    // 金额用分存储
    Status      int     `torm:"type:tinyint"`               // 状态用小整数
    Config      string  `torm:"type:json"`                  // JSON配置
    Avatar      string  `torm:"type:varchar,size:255"`      // URL字段
    Description string  `torm:"type:text"`                  // 长文本
}

// 对应的访问器处理
func (u *User) GetBalanceAttr(value interface{}) interface{} {
    cents := convertToInt(value)
    return map[string]interface{}{
        "cents":     cents,
        "yuan":      float64(cents) / 100.0,
        "formatted": fmt.Sprintf("¥%.2f", float64(cents)/100.0),
    }
}
```

访问器系统让 TORM 的数据处理更加灵活和强大，支持复杂的业务逻辑和数据转换需求。