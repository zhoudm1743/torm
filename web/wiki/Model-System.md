# 模型系统

TORM 采用 Active Record 模式的模型系统，让你可以用面向对象的方式操作数据库。每个模型对应一个数据库表，模型实例对应表中的一行记录。

**重要说明**: TORM模型系统内置使用`db`包的`QueryInterface`进行数据库操作，通过`getQueryBuilder()`方法获取查询构建器，所有模型的查询方法都是对底层查询构建器的封装。

**模型操作特性**:
- **默认表操作**: 所有模型操作都默认操作当前模型对应的表，无需手动指定表名
- **智能表名处理**: JOIN查询中如果字段名不包含表名，自动添加当前模型表名
- **关联查询**: 关联查询自动使用相关模型的表名，完全基于模型定义

### 模型 vs 查询构建器

| 功能 | 模型方式 | 查询构建器方式 |
|------|----------|----------------|
| 基础查询 | `user.Where("status", "=", "active").Get()` | `db.Table("users").Where("status", "=", "active").Get()` |
| 参数化查询 | ✅ `user.Where("name = ?", "张三").Get()` | ✅ `db.Table("users").Where("name = ?", "张三").Get()` |
| 数据填充 | ✅ 自动填充模型属性 | ❌ 返回`map[string]interface{}` |
| 生命周期钩子 | ✅ 支持BeforeSave、AfterCreate等 | ❌ 不支持 |
| 时间戳管理 | ✅ 自动管理created_at、updated_at | ❌ 需要手动处理 |
| 软删除 | ✅ 自动处理deleted_at | ❌ 需要手动添加条件 |

**建议**: 
- 简单查询使用**模型**或**查询构建器**（都支持参数化查询）
- 业务逻辑和数据管理使用**模型**（自动处理生命周期和数据填充）
- 复杂SQL查询使用**查询构建器**（更灵活的原生SQL支持）

## 📋 目录

- [模型定义](#模型定义)
- [自动迁移](#自动迁移)
- [基础操作](#基础操作)
- [查询方法](#查询方法)
- [属性管理](#属性管理)
- [事件钩子](#事件钩子)
- [时间戳](#时间戳)
- [软删除](#软删除)
- [自定义主键](#自定义主键)
- [作用域](#作用域)
- [序列化](#序列化)

## 🚀 快速开始

### 基础模型定义

```go
package models

import (
    "time"
    "github.com/zhoudm1743/torm/model"
)

// User 用户模型
type User struct {
    model.BaseModel                                    // 嵌入基础模型
    ID        interface{} `json:"id" db:"id"`         // 主键
    Name      string      `json:"name" db:"name"`     // 用户名
    Email     string      `json:"email" db:"email"`   // 邮箱
    Age       int         `json:"age" db:"age"`       // 年龄
    Status    string      `json:"status" db:"status"` // 状态
    CreatedAt time.Time   `json:"created_at" db:"created_at"`
    UpdatedAt time.Time   `json:"updated_at" db:"updated_at"`
}

// NewUser 创建用户模型实例
func NewUser() *User {
    user := &User{}
    user.BaseModel = *model.NewBaseModelWithAutoDetect(user) // 自动检测配置
    user.SetTable("users")        // 设置表名
    user.SetConnection("default") // 设置数据库连接
    return user
}
```

## 🗄️ 自动迁移

### AutoMigrate 功能

TORM v1.1.6 引入了强大的 `AutoMigrate()` 功能，可以根据模型结构体自动创建和更新数据库表结构。

#### 基本使用

```go
type Product struct {
    model.BaseModel
    ID          int64     `json:"id" db:"id" torm:"primary_key,auto_increment,comment:产品ID"`
    Name        string    `json:"name" db:"name" torm:"size:200,comment:产品名称"`
    Description string    `json:"description" db:"description" torm:"type:text,comment:产品描述"`
    Price       float64   `json:"price" db:"price" torm:"type:decimal,precision:10,scale:2,comment:价格"`
    SKU         string    `json:"sku" db:"sku" torm:"size:50,unique,comment:产品编码"`
    CategoryID  int64     `json:"category_id" db:"category_id" torm:"index,comment:分类ID"`
    IsActive    bool      `json:"is_active" db:"is_active" torm:"default:true,comment:是否启用"`
    Tags        []string  `json:"tags" db:"tags" torm:"comment:标签列表（JSON）"`
    Images      []byte    `json:"images" db:"images" torm:"comment:图片数据"`
    Metadata    map[string]interface{} `json:"metadata" db:"metadata" torm:"comment:元数据"`
    CreatedAt   int64     `json:"created_at" db:"created_at" torm:"auto_create_time,comment:创建时间"`
    UpdatedAt   int64     `json:"updated_at" db:"updated_at" torm:"auto_update_time,comment:更新时间"`
}

// NewProduct 创建产品模型
func NewProduct() *Product {
    product := &Product{}
    product.BaseModel = *model.NewBaseModelWithAutoDetect(product)
    product.SetTable("products")
    product.SetConnection("default")
    return product
}

func main() {
    // 配置数据库
    config := &db.Config{
        Driver:   "mysql",
        Host:     "localhost",
        Database: "myapp",
        Username: "root",
        Password: "password",
    }
    db.AddConnection("default", config)
    
    // 创建模型并执行自动迁移
    product := NewProduct()
    if err := product.AutoMigrate(); err != nil {
        log.Fatalf("自动迁移失败: %v", err)
    }
    
    fmt.Println("产品表创建成功！")
}
```

#### TORM 统一标签语法

TORM v1.1.6 引入了统一的 `torm` 标签，大大简化了模型定义。所有标签都支持**大小写不敏感**：

| 类型 | 语法格式 | 示例 | 说明 |
|------|----------|------|------|
| **主键约束** | `primary_key`, `pk` | `torm:"primary_key"` | 标记为主键字段 |
| **自增长** | `auto_increment` | `torm:"primary_key,auto_increment"` | 自动递增（通常与主键一起） |
| **唯一约束** | `unique` | `torm:"unique"` | 唯一性约束 |
| **索引** | `index`, `index:名称` | `torm:"index"`, `torm:"index:user_idx"` | 创建索引，可指定名称 |
| **数据类型** | `type:类型名` | `torm:"type:varchar,size:100"` | 明确指定数据库列类型 |
| **字段长度** | `size:数字` | `torm:"size:100"` | 字符串类型的长度 |
| **数值精度** | `precision:数字` | `torm:"type:decimal,precision:10"` | DECIMAL类型精度 |
| **小数位** | `scale:数字` | `torm:"precision:10,scale:2"` | DECIMAL类型小数位数 |
| **默认值** | `default:值` | `torm:"default:active"` | 设置默认值 |
| **允许NULL** | `nullable` | `torm:"nullable"` | 明确允许NULL值 |
| **不允许NULL** | `not_null` | `torm:"not_null"` | 明确不允许NULL值 |
| **自动时间** | `auto_create_time` | `torm:"auto_create_time"` | 创建时自动设置当前时间 |
| **自动更新** | `auto_update_time` | `torm:"auto_update_time"` | 更新时自动设置当前时间 |
| **字段注释** | `comment:描述` | `torm:"comment:用户名"` | 添加列注释 |

#### 大小写不敏感支持

TORM 支持完全的大小写不敏感，以下所有写法都完全等效：

```go
// 所有这些定义都会产生相同的结果
type FlexibleModel struct {
    model.BaseModel
    // 全小写（推荐）
    Field1 string `torm:"type:varchar,size:50,unique,comment:字段1"`
    
    // 全大写
    Field2 string `torm:"TYPE:VARCHAR,SIZE:50,UNIQUE,COMMENT:字段2"`
    
    // 首字母大写
    Field3 string `torm:"Type:VarChar,Size:50,Unique,Comment:字段3"`
    
    // 混合大小写
    Field4 string `torm:"TYPE:varchar,SIZE:50,unique,COMMENT:字段4"`
    
    // 随意大小写（不推荐，但支持）
    Field5 string `torm:"tYpE:VaRcHaR,sIzE:50,UnIqUe,CoMmEnT:字段5"`
}
```

**大小写不敏感范围：**
- ✅ 标志位：`primary_key` = `PRIMARY_KEY` = `Primary_Key`
- ✅ 类型名：`varchar` = `VARCHAR` = `VarChar`
- ✅ 属性名：`size` = `SIZE` = `Size`
- ✅ 默认值：`true` = `TRUE` = `True`, `null` = `NULL` = `Null`
- ✅ 注释：`comment` = `COMMENT` = `Comment`

#### 组合使用示例

```go
type User struct {
    model.BaseModel
    // 主键：自增、主键、注释
    ID        int64  `db:"id" torm:"primary_key,auto_increment,comment:用户ID"`
    
    // 字符串：长度、唯一、注释  
    Email     string `db:"email" torm:"size:100,unique,comment:邮箱地址"`
    
    // 数值：类型、精度、默认值
    Balance   float64 `db:"balance" torm:"type:decimal,precision:10,scale:2,default:0.00"`
    
    // 索引：自定义索引名、注释
    UserID    int64  `db:"user_id" torm:"index:user_relation_idx,comment:关联用户"`
    
    // 时间戳：自动创建时间
    CreatedAt int64  `db:"created_at" torm:"auto_create_time,comment:创建时间"`
    
    // 可空字段：允许NULL、注释
    Avatar    *string `db:"avatar" torm:"nullable,comment:头像URL"`
}
```

#### 详细类型长度和精度控制

**字符串类型长度：**
```go
type StringExamples struct {
    model.BaseModel
    ShortCode   string `db:"short_code" torm:"type:varchar,size:10,comment:短编码"`     // VARCHAR(10)
    Name        string `db:"name" torm:"type:varchar,size:50,comment:名称"`           // VARCHAR(50)  
    Description string `db:"description" torm:"type:varchar,size:200,comment:描述"`  // VARCHAR(200)
    FixedCode   string `db:"fixed_code" torm:"type:char,size:8,comment:固定编码"`     // CHAR(8)
    CountryCode string `db:"country_code" torm:"type:char,size:2,comment:国家代码"`   // CHAR(2)
    LongText    string `db:"long_text" torm:"type:text,comment:长文本"`              // TEXT
}
```

**数值类型精度和小数位：**
```go
type NumericExamples struct {
    model.BaseModel
    // DECIMAL(precision, scale) - precision总位数，scale小数位数
    Price       float64 `db:"price" torm:"type:decimal,precision:10,scale:2,comment:价格"`        // DECIMAL(10,2) - 最大8位整数,2位小数
    Rate        float64 `db:"rate" torm:"type:decimal,precision:5,scale:4,comment:利率"`          // DECIMAL(5,4)  - 最大1位整数,4位小数  
    Amount      float64 `db:"amount" torm:"type:decimal,precision:15,scale:2,comment:金额"`       // DECIMAL(15,2) - 最大13位整数,2位小数
    Percentage  float64 `db:"percentage" torm:"type:decimal,precision:6,scale:3,comment:百分比"`  // DECIMAL(6,3)  - 最大3位整数,3位小数
    Weight      float64 `db:"weight" torm:"type:decimal,precision:8,scale:3,comment:重量"`        // DECIMAL(8,3)  - 最大5位整数,3位小数
}
```

**实际数据示例：**
| 业务场景 | 数据例子 | 推荐类型 | TORM标签 |
|----------|----------|----------|----------|
| 商品价格 | 123.45 | DECIMAL(10,2) | `torm:"type:decimal,precision:10,scale:2"` |
| 利率 | 0.0325 (3.25%) | DECIMAL(5,4) | `torm:"type:decimal,precision:5,scale:4"` |
| 银行金额 | 1234567.89 | DECIMAL(15,2) | `torm:"type:decimal,precision:15,scale:2"` |
| 百分比得分 | 98.456% | DECIMAL(6,3) | `torm:"type:decimal,precision:6,scale:3"` |
| 商品重量 | 12.345kg | DECIMAL(8,3) | `torm:"type:decimal,precision:8,scale:3"` |
| 产品编码 | "P12345" | VARCHAR(10) | `torm:"type:varchar,size:10"` |
| 国家代码 | "CN" | CHAR(2) | `torm:"type:char,size:2"` |

#### 自定义类型映射

```go
type AdvancedModel struct {
    model.BaseModel
    // 字符串类型
    Title       string  `db:"title" torm:"type:varchar,size:200"`
    Content     string  `db:"content" torm:"type:text"`
    Summary     string  `db:"summary" torm:"type:longtext"`
    Code        string  `db:"code" torm:"type:char,size:10"`
    
    // 数值类型
    SmallNum    int8    `db:"small_num" torm:"type:tinyint"`
    MediumNum   int16   `db:"medium_num" torm:"type:smallint"`
    BigNum      int64   `db:"big_num" torm:"type:bigint"`
    Price       float64 `db:"price" torm:"type:decimal,precision:10,scale:2"`
    
    // 时间类型
    CreatedDate time.Time `db:"created_date" torm:"type:date"`
    UpdatedTime time.Time `db:"updated_time" torm:"type:timestamp"`
    
    // 二进制和JSON
    BinaryData  []byte              `db:"binary_data" torm:"type:blob"`
    JsonData    map[string]interface{} `db:"json_data" torm:"type:json"`
    
    // 布尔类型
    IsEnabled   bool    `db:"is_enabled" torm:"type:boolean,default:true"`
}
```

#### 跨数据库兼容

AutoMigrate 自动适配不同数据库的类型映射：

| Go类型 | MySQL | PostgreSQL | SQLite |
|--------|-------|------------|--------|
| `string` | `VARCHAR(n)` | `VARCHAR(n)` | `TEXT` |
| `int64` | `BIGINT` | `BIGINT` | `INTEGER` |
| `float64` | `DOUBLE` | `DOUBLE PRECISION` | `REAL` |
| `bool` | `BOOLEAN` | `BOOLEAN` | `INTEGER` |
| `[]byte` | `BLOB` | `BYTEA` | `BLOB` |
| `[]string` | `JSON` | `JSONB` | `TEXT` |
| `map[string]interface{}` | `JSON` | `JSONB` | `TEXT` |
| `time.Time` | `DATETIME` | `TIMESTAMP` | `DATETIME` |

#### 自动索引创建

AutoMigrate 会自动为以下情况创建索引：

1. **唯一字段**: `torm:"unique"` 自动创建唯一索引
2. **明确索引**: `torm:"index"` 创建普通索引
3. **外键字段**: 以 `_id` 结尾的字段自动创建索引
4. **自定义索引名**: `torm:"index:custom_name"` 使用指定名称

```go
type UserProfile struct {
    model.BaseModel
    UserID      int64  `db:"user_id" torm:"index"`                    // 自动索引: idx_user_profiles_user_id
    Email       string `db:"email" torm:"unique"`                     // 唯一索引: idx_user_profiles_email_unique  
    Phone       string `db:"phone" torm:"index:phone_idx"`            // 自定义索引: phone_idx
    CompanyID   int64  `db:"company_id"`                              // 自动索引: idx_user_profiles_company_id（_id后缀）
}
```

#### 最佳实践

```go
// ✅ 推荐：使用 NewBaseModelWithAutoDetect
func NewUser() *User {
    user := &User{}
    user.BaseModel = *model.NewBaseModelWithAutoDetect(user)
    user.SetTable("users")
    user.SetConnection("default")
    return user
}

// ✅ 推荐：在应用启动时执行 AutoMigrate
func initDatabase() {
    models := []interface{}{
        NewUser(),
        NewProduct(),
        NewOrder(),
    }
    
    for _, model := range models {
        if migrator, ok := model.(interface{ AutoMigrate() error }); ok {
            if err := migrator.AutoMigrate(); err != nil {
                log.Printf("AutoMigrate failed for %T: %v", model, err)
            }
        }
    }
}

// ✅ 推荐：结合传统迁移使用
func setupDatabase() {
    // 1. 使用 AutoMigrate 快速创建基础表结构
    user := NewUser()
    user.AutoMigrate()
    
    // 2. 使用传统迁移处理复杂变更
    migrator := migration.NewMigrator(conn, logger)
    migrator.RegisterFunc("20240101_001", "添加用户表索引", addUserIndexes, dropUserIndexes)
    migrator.Up()
}
```

## 📊 模型定义

### 字段标签

```go
type User struct {
    model.BaseModel
    ID        uint       `json:"id" db:"id" pk:""`                           // 主键标签
    Name      string     `json:"name" db:"name" validate:"required"`         // 验证标签
    Email     string     `json:"email" db:"email" unique:"true"`             // 唯一索引
    Password  string     `json:"-" db:"password"`                            // 隐藏字段
    Profile   string     `json:"profile" db:"profile" type:"json"`           // JSON字段
    Avatar    *string    `json:"avatar" db:"avatar"`                         // 可空字段
    CreatedAt time.Time  `json:"created_at" db:"created_at;autoCreateTime"`  // 自动创建时间
    UpdatedAt time.Time  `json:"updated_at" db:"updated_at;autoUpdateTime"`  // 自动更新时间
    DeletedAt model.DeletedTime `json:"deleted_at" db:"deleted_at"`          // 软删除字段
}
```

#### 支持的标签

- **`pk`**: 主键标签，标记为主键字段
- **`autoCreateTime`**: 自动创建时间，插入时自动设置当前时间
- **`autoUpdateTime`**: 自动更新时间，插入和更新时自动设置当前时间
- **`model.DeletedTime`**: 软删除字段类型，自动启用软删除功能

#### 标签优先级

结构体字段标签的优先级**高于**BaseModel的基础配置：

```go
func NewUser() *User {
    user := &User{BaseModel: *model.NewBaseModel()}
    user.SetTable("users")
    user.SetConnection("default")
    user.DetectConfigFromStruct(user) // 从标签检测配置，优先级更高
    return user
}
```

### 表名约定

```go
// 自动推断表名（结构体名的复数形式）
type User struct { /* ... */ }        // 对应表名: users
type BlogPost struct { /* ... */ }    // 对应表名: blog_posts

// 自定义表名
func (u *User) TableName() string {
    return "custom_users"
}

// 在模型初始化时设置
func NewUser() *User {
    user := &User{BaseModel: *model.NewBaseModel()}
    user.SetTable("users")
    return user
}
```

### 连接配置

```go
// 设置数据库连接
user.SetConnection("mysql")    // 使用指定连接
user.SetConnection("default")  // 使用默认连接

// 不同模型使用不同数据库
type User struct { /* ... */ }      // 使用主数据库
type Log struct { /* ... */ }       // 使用日志数据库

func NewLog() *Log {
    log := &Log{BaseModel: *model.NewBaseModel()}
    log.SetConnection("log_db")
    return log
}
```

## 🎯 基础操作

### 创建记录

```go
// 方法1：直接创建
user := NewUser()
user.Name = "张三"
user.Email = "zhangsan@example.com"
user.Age = 25
err := user.Save()

// 方法2：批量设置属性
user := NewUser()
err := user.Fill(map[string]interface{}{
    "name":  "李四",
    "email": "lisi@example.com",
    "age":   30,
}).Save()

// 方法3：使用Create方法
user := NewUser()
err := user.Create(map[string]interface{}{
    "name":  "王五",
    "email": "wangwu@example.com",
    "age":   28,
})
```

### 查找记录

```go
// 根据主键查找
user := NewUser()
err := user.Find(1)  // 查找ID为1的用户

// 查找第一条记录
user := NewUser()
err := user.First()

// 带条件查找 - 传统方式
user := NewUser()
err := user.Where("email", "=", "user@example.com").First()

// 带条件查找 - 参数化方式
user2 := NewUser()
err = user2.Where("email = ?", "user@example.com").First()

// 查找或失败（找不到会返回错误）
user := NewUser()
err := user.FindOrFail(1)
```

### 更新记录

```go
// 查找并更新
user := NewUser()
err := user.Find(1)
if err == nil {
    user.Name = "新名字"
    user.Age = 26
    err = user.Save()
}

// 直接更新
user := NewUser()
err := user.Where("id", "=", 1).Update(map[string]interface{}{
    "name": "更新的名字",
    "age":  27,
})

// 批量更新 - 适配db.Update
user := NewUser()
affected, err := user.Where("status = ?", "inactive").
    Update(map[string]interface{}{
        "status": "archived",
    })

// 批量插入 - 适配db.InsertBatch
insertedCount, err := user.InsertBatch([]map[string]interface{}{
    {"name": "用户1", "email": "user1@example.com", "age": 25},
    {"name": "用户2", "email": "user2@example.com", "age": 30},
    {"name": "用户3", "email": "user3@example.com", "age": 28},
})
```

### 删除记录

```go
// 删除单条记录
user := NewUser()
err := user.Find(1)
if err == nil {
    err = user.Delete()
}

// 条件删除
user := NewUser()
affected, err := user.Where("status", "=", "inactive").Delete()

// 批量删除
user := NewUser()
affected, err := user.WhereIn("id", []interface{}{1, 2, 3}).Delete()
```

## 🔍 查询方法

### 基础查询

```go
user := NewUser()

// 获取所有记录
users, err := user.All()

// 条件查询 - 传统三参数方式
users, err := user.Where("age", ">", 18).
    Where("status", "=", "active").
    Get()

// 条件查询 - 参数化查询方式
users, err = user.Where("age > ? AND status = ?", 18, "active").Get()

// 混合使用
users, err = user.Where("age", ">", 18).           // 传统方式
    Where("name LIKE ?", "%admin%").              // 参数化方式
    Where("status", "=", "active").               // 传统方式
    Get()

// 排序
users, err := user.OrderBy("created_at", "desc").Get()

// 限制数量
users, err := user.Limit(10).Get()

// 分页
result, err := user.Paginate(1, 10) // 第1页，每页10条
```

### 聚合查询

```go
user := NewUser()

// 计数
count, err := user.Where("status", "=", "active").Count()

// 检查存在（使用HasRecords方法）
exists, err := user.Where("email", "=", "test@example.com").HasRecords()

// 检查记录是否存在
exists, err := user.Where("email", "=", "test@example.com").HasRecords()

// 注意：当前版本暂不支持Max、Min、Sum、Avg等聚合函数
// 可以使用原生SQL查询实现复杂聚合操作
```

### 高级查询

```go
user := NewUser()

// 原生SQL条件
users, err := user.WhereRaw("YEAR(created_at) = ?", 2023).Get()

// 复杂参数化查询
users, err = user.Where("(age BETWEEN ? AND ?) OR status IN (?, ?)", 
    18, 65, "active", "premium").Get()

// OR条件
users, err = user.Where("name = ?", "admin").
    OrWhere("email = ?", "admin@example.com").Get()

// JOIN查询 - 自动处理当前模型表名
users, err = user.
    LeftJoin("profiles", "user_id", "=", "id").  // 自动添加表名：profiles.user_id = users.id
    Select("users.*", "profiles.avatar").
    Where("status = ?", "active").Get()          // 自动使用users.status

// 也可以显式指定表名
users, err = user.
    LeftJoin("profiles", "profiles.user_id", "=", "users.id").
    Select("users.*", "profiles.avatar").
    Where("users.status = ?", "active").Get()

// 分组和聚合
users, err = user.
    SelectRaw("status, COUNT(*) as count").
    GroupBy("status").
    Having("count", ">", 10).Get()

// 去重查询
users, err = user.Select("city").Distinct().Get()

// 子查询
users, err := user.WhereExists(func(q db.QueryInterface) db.QueryInterface {
    return q.Table("orders").
        Where("orders.user_id", "=", "users.id").
        Where("orders.status", "=", "completed")
}).Get()

// JOIN查询
users, err := user.
    LeftJoin("profiles", "profiles.user_id", "=", "users.id").
    Select("users.*", "profiles.avatar").
    Get()
```

## 💼 属性管理

### 属性访问

```go
user := NewUser()
err := user.Find(1)

// 获取属性
name := user.GetAttribute("name")
email := user.GetAttribute("email")

// 设置属性
user.SetAttribute("name", "新名字")
user.SetAttribute("age", 30)

// 批量设置
user.SetAttributes(map[string]interface{}{
    "name": "批量设置的名字",
    "age":  35,
})

// 获取所有属性
attributes := user.GetAttributes()
```

### 脏数据检测

```go
user := NewUser()
err := user.Find(1)

// 修改属性
user.Name = "新名字"
user.Age = 30

// 检查是否有变更
isDirty := user.IsDirty()           // true
isDirtyName := user.IsDirty("name") // true
isDirtyEmail := user.IsDirty("email") // false

// 获取变更的字段
dirty := user.GetDirty() // map[string]interface{}{"name": "新名字", "age": 30}

// 获取原始值
original := user.GetOriginal("name") // 原始名字
```

### 属性转换

```go
// 自定义getter和setter
type User struct {
    model.BaseModel
    // ... 其他字段
}

// 自定义getter
func (u *User) GetNameAttribute() string {
    name := u.GetAttribute("name")
    if name == nil {
        return ""
    }
    return strings.ToUpper(name.(string)) // 总是返回大写
}

// 自定义setter
func (u *User) SetPasswordAttribute(password string) {
    // 密码加密后存储
    hashedPassword := hashPassword(password)
    u.SetAttribute("password", hashedPassword)
}
```

## 🎣 事件钩子

### 生命周期钩子

```go
type User struct {
    model.BaseModel
    // ... 字段定义
}

// 保存前
func (u *User) BeforeSave() error {
    // 数据验证
    if u.GetAttribute("email") == "" {
        return errors.New("邮箱不能为空")
    }
    return nil
}

// 保存后
func (u *User) AfterSave() error {
    // 发送通知、清除缓存等
    log.Printf("用户 %s 已保存", u.GetAttribute("name"))
    return nil
}

// 创建前
func (u *User) BeforeCreate() error {
    // 设置默认值
    u.SetAttribute("status", "active")
    return nil
}

// 创建后
func (u *User) AfterCreate() error {
    // 创建用户档案、发送欢迎邮件等
    return u.createUserProfile()
}

// 更新前
func (u *User) BeforeUpdate() error {
    // 更新时间戳
    u.SetAttribute("updated_at", time.Now())
    return nil
}

// 更新后
func (u *User) AfterUpdate() error {
    // 清除相关缓存
    return clearUserCache(u.GetKey())
}

// 删除前
func (u *User) BeforeDelete() error {
    // 检查是否可以删除
    if u.GetAttribute("status") == "admin" {
        return errors.New("管理员用户不能删除")
    }
    return nil
}

// 删除后
func (u *User) AfterDelete() error {
    // 清理相关数据
    return u.cleanupUserData()
}
```

### 查找钩子

```go
// 查找后
func (u *User) AfterFind() error {
    // 解密敏感数据、格式化显示等
    return nil
}
```

## ⏰ 时间戳

### 自动时间戳

```go
type User struct {
    model.BaseModel
    // ... 其他字段
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func NewUser() *User {
    user := &User{BaseModel: *model.NewBaseModel()}
    user.SetTable("users")
    
    // 启用自动时间戳
    user.EnableTimestamps()
    
    // 自定义时间戳字段名
    user.SetCreatedAtColumn("created_at")
    user.SetUpdatedAtColumn("updated_at")
    
    return user
}
```

### 禁用时间戳

```go
user := NewUser()
user.DisableTimestamps() // 禁用自动时间戳

// 或者在特定操作中禁用
user.WithoutTimestamps(func() error {
    return user.Save() // 这次保存不会更新时间戳
})
```

## 🗑️ 软删除

### 启用软删除

```go
type User struct {
    model.BaseModel
    // ... 其他字段
    DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
}

func NewUser() *User {
    user := &User{BaseModel: *model.NewBaseModel()}
    user.SetTable("users")
    
    // 启用软删除
    user.EnableSoftDeletes()
    user.SetDeletedAtColumn("deleted_at")
    
    return user
}
```

### 软删除操作

```go
user := NewUser()

// 软删除（设置deleted_at字段）
err := user.Find(1)
err = user.Delete() // 软删除

// 查询时自动排除软删除记录
users, err := user.Where("status", "=", "active").Get() // 不包含软删除记录

// 包含软删除记录
users, err := user.WithTrashed().Get()

// 只查询软删除记录
users, err := user.OnlyTrashed().Get()

// 恢复软删除记录
err = user.Find(1) // 这会失败，因为记录被软删除
err = user.WithTrashed().Find(1)
err = user.Restore()

// 硬删除（彻底删除）
err = user.WithTrashed().Find(1)
err = user.ForceDelete()
```

## 🔑 自定义主键

### UUID主键

```go
type Product struct {
    model.BaseModel
    UUID      string    `json:"uuid" db:"uuid" primary:"true"`
    Name      string    `json:"name" db:"name"`
    Price     float64   `json:"price" db:"price"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func NewProduct() *Product {
    product := &Product{BaseModel: *model.NewBaseModel()}
    product.SetTable("products")
    // 自动检测主键标签
    product.DetectPrimaryKeysFromStruct(product)
    return product
}
```

### 复合主键

```go
type UserRole struct {
    model.BaseModel
    TenantID  string    `json:"tenant_id" db:"tenant_id" primary:"true"`
    UserID    string    `json:"user_id" db:"user_id" primary:"true"`
    Role      string    `json:"role" db:"role"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func NewUserRole() *UserRole {
    userRole := &UserRole{BaseModel: *model.NewBaseModel()}
    userRole.SetTable("user_roles")
    userRole.DetectPrimaryKeysFromStruct(userRole)
    return userRole
}

// 使用复合主键
userRole := NewUserRole()
userRole.SetAttribute("tenant_id", "tenant-001")
userRole.SetAttribute("user_id", "user-001")
userRole.SetAttribute("role", "admin")
err := userRole.Save()

// 根据复合主键查找
userRole2 := NewUserRole()
userRole2.SetAttribute("tenant_id", "tenant-001")
userRole2.SetAttribute("user_id", "user-001")
err = userRole2.Find() // Find方法会使用所有主键字段
```

## 🎯 作用域

### 自定义查询方法（替代作用域）

```go
type User struct {
    model.BaseModel
    // ... 字段定义
}

// 定义自定义查询方法 - 默认操作users表
func (u *User) GetActiveUsers() ([]map[string]interface{}, error) {
    return u.Where("status = ?", "active").Get()  // 自动查询users表
}

func (u *User) GetAdultUsers() ([]map[string]interface{}, error) {
    return u.Where("age >= ?", 18).Get()  // 自动查询users表
}

func (u *User) GetUsersByCity(city string) ([]map[string]interface{}, error) {
    return u.Where("city = ?", city).Get()  // 自动查询users表
}

// 复合条件查询方法
func (u *User) GetActiveAdultUsers() ([]map[string]interface{}, error) {
    return u.Where("status = ? AND age >= ?", "active", 18).Get()  // 默认users表，无需指定
}

// 带JOIN的自定义查询 - 智能处理表名
func (u *User) GetUsersWithProfiles() ([]map[string]interface{}, error) {
    return u.LeftJoin("profiles", "user_id", "=", "id").  // 自动：profiles.user_id = users.id
        Select("users.*", "profiles.avatar").
        Where("status = ?", "active").Get()  // 自动：users.status
}
```

### 使用自定义查询方法

```go
user := NewUser()

// 使用自定义查询方法
activeUsers, err := user.GetActiveUsers()

// 复合条件查询
activeAdults, err := user.GetActiveAdultUsers()

// 带参数的查询
beijingUsers, err := user.GetUsersByCity("北京")

// 与链式查询结合
users, err := user.Where("vip_level", ">", 3).
    Where("status", "=", "active").
    OrderBy("created_at", "desc").
    Get()
```

### 查询方法说明

```go
// 注意：当前版本暂不支持作用域（Scope）功能
// 推荐使用自定义查询方法或直接链式调用Where方法

// 示例：实现复杂查询逻辑
func (u *User) GetPremiumUsers(minVipLevel int) ([]map[string]interface{}, error) {
    return u.Where("status", "=", "active").
        Where("vip_level", ">=", minVipLevel).
        Where("deleted_at", "IS", nil).
        OrderBy("created_at", "desc").
        Get()
}
```

## 📤 序列化

### Map序列化

```go
user := NewUser()
err := user.Find(1)

// 转换为Map
userData := user.ToMap()

// 获取所有属性
attributes := user.GetAttributes()

// 获取主键值
keyValue := user.GetKey()

// 注意：当前版本暂不支持ToJSON()方法
// 可以使用encoding/json包手动序列化ToMap()的结果

// 隐藏敏感字段（在结构体定义时）
type User struct {
    model.BaseModel
    Name     string `json:"name"`
    Email    string `json:"email"`
    Password string `json:"-"`        // 不会被序列化
    Secret   string `json:"secret,omitempty"` // 空值时不序列化
}
```

### 自定义序列化

```go
type User struct {
    model.BaseModel
    // ... 字段定义
}

// 自定义序列化格式
func (u *User) Serialize() map[string]interface{} {
    return map[string]interface{}{
        "id":         u.GetAttribute("id"),
        "name":       u.GetAttribute("name"),
        "email":      u.GetAttribute("email"),
        "avatar_url": u.getAvatarURL(),
        "is_admin":   u.isAdmin(),
    }
}

// 转换为JSON字符串（手动实现）
func (u *User) ToJSONString() (string, error) {
    import "encoding/json"
    
    data := u.ToMap()
    jsonBytes, err := json.Marshal(data)
    if err != nil {
        return "", err
    }
    return string(jsonBytes), nil
}

func (u *User) getAvatarURL() string {
    avatar := u.GetAttribute("avatar")
    if avatar == nil {
        return "/default-avatar.png"
    }
    return "/avatars/" + avatar.(string)
}

func (u *User) isAdmin() bool {
    role := u.GetAttribute("role")
    return role == "admin"
}
```

## 🔧 高级功能

### 模型工厂

```go
// 定义工厂方法
func UserFactory() *User {
    user := NewUser()
    user.Fill(map[string]interface{}{
        "name":   "测试用户",
        "email":  fmt.Sprintf("test%d@example.com", rand.Int()),
        "age":    rand.Intn(50) + 18,
        "status": "active",
    })
    return user
}

// 批量创建测试数据
func CreateTestUsers(count int) error {
    for i := 0; i < count; i++ {
        user := UserFactory()
        if err := user.Save(); err != nil {
            return err
        }
    }
    return nil
}
```

### 模型观察者

```go
// 注册模型观察者
type UserObserver struct{}

func (o *UserObserver) Creating(user *User) error {
    // 创建前的处理
    return nil
}

func (o *UserObserver) Created(user *User) error {
    // 创建后的处理
    log.Printf("新用户创建: %s", user.GetAttribute("name"))
    return nil
}

// 注册观察者
func init() {
    model.RegisterObserver(&User{}, &UserObserver{})
}
```

## 📚 最佳实践

### 1. 模型结构设计

```go
// 好的做法：清晰的模型结构
type User struct {
    model.BaseModel
    
    // 基础字段
    ID    int64  `json:"id" db:"id" primary:"true"`
    Name  string `json:"name" db:"name" validate:"required"`
    Email string `json:"email" db:"email" validate:"required,email" unique:"true"`
    
    // 状态字段
    Status    string `json:"status" db:"status" default:"active"`
    IsActive  bool   `json:"is_active" db:"is_active" default:"true"`
    
    // 时间戳
    CreatedAt time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
    DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
}
```

### 2. 错误处理

```go
// 好的做法：完整的错误处理
func CreateUser(userData map[string]interface{}) (*User, error) {
    user := NewUser()
    
    // 数据验证
    if err := validateUserData(userData); err != nil {
        return nil, fmt.Errorf("数据验证失败: %w", err)
    }
    
    // 填充数据
    user.Fill(userData)
    
    // 保存
    if err := user.Save(); err != nil {
        return nil, fmt.Errorf("保存用户失败: %w", err)
    }
    
    return user, nil
}
```

### 3. 性能优化

```go
// 好的做法：只查询需要的字段
users, err := user.Select("id", "name", "email").
    Where("status", "=", "active").
    Limit(100).
    Get()

// 使用分页避免大量数据
result, err := user.Where("status", "=", "active").
    Paginate(page, 20)

// 使用索引优化查询
users, err := user.Where("email", "=", email). // email应该有索引
    Where("status", "=", "active").              // 复合索引
    Get()
```

## 🔗 查询构建器模型支持

TORM的查询构建器现在也支持模型特性！通过`WithModel()`方法绑定模型后，查询构建器能够：

### 自动时间戳管理

```go
type User struct {
    model.BaseModel
    ID        uint      `db:"id" pk:""`
    Name      string    `db:"name"`
    CreatedAt time.Time `db:"created_at;autoCreateTime"`
    UpdatedAt time.Time `db:"updated_at;autoUpdateTime"`
}

// 直接从模型创建查询构建器 - 自动表名 + 模型特性
userModel := &User{}
query, err := db.Model(userModel)  // 自动获取表名，自动启用模型特性

// 插入时自动设置创建时间和更新时间
newUser := &User{Name: "张三"}
id, err := query.InsertModel(newUser)

// 更新时自动设置更新时间
user.Name = "李四"
affected, err := query.Where("id = ?", id).UpdateModel(user)
```

### 自动软删除

```go
type User struct {
    model.BaseModel
    ID        uint      `db:"id" pk:""`
    Name      string    `db:"name"`
    DeletedAt model.DeletedTime `db:"deleted_at"`  // 启用软删除
}

// 软删除功能自动启用
query, err := db.Model(&User{})  // 自动获取表名和软删除配置

// 查询时自动排除软删除记录
users, err := query.Where("status = ?", "active").Get()

// 删除时自动设置deleted_at而不是物理删除
affected, err := query.Where("id = ?", 1).Delete()
```

### 智能主键识别

```go
type Product struct {
    model.BaseModel
    UUID string `db:"uuid" pk:""`  // 自定义主键
    Name string `db:"name"`
}

// 自动识别主键字段
query, err := db.Model(&Product{})  // 自动获取表名和主键配置
var product Product
err := query.FindModel("some-uuid", &product)  // 自动使用uuid字段查询
```

### 方法对比

| 功能 | 传统查询构建器 | 模型查询构建器 |
|------|----------------|----------------|
| 创建查询 | `db.Table("users")` | `db.Model(&User{})` |
| 表名 | 手动指定 | 自动从模型获取 |
| 插入数据 | `Insert(map[string]interface{}{...})` | `InsertModel(&User{...})` |
| 更新数据 | `Update(map[string]interface{}{...})` | `UpdateModel(&User{...})` |
| 查找数据 | `First()` 返回map | `FindModel(id, &user)` 直接填充结构体 |
| 时间戳 | 手动设置 | 自动根据标签设置 |
| 软删除 | 手动添加WHERE条件 | 自动过滤软删除记录 |
| 主键 | 硬编码字段名 | 自动从标签识别 |

### API演进对比

```go
// 旧方式：需要手动指定表名和绑定模型
db.Table("users").WithModel(&User{}).Where("age > ?", 18).Get()

// 新方式：一步到位，自动获取所有配置
db.Model(&User{}).Where("age > ?", 18).Get()
```

### 表名获取优先级

TORM的表名获取遵循以下优先级：

```go
user := &User{}

// 优先级1：手动设置的表名（最高优先级）
user.SetTable("custom_users")
db.Model(user) // 使用 "custom_users"

// 优先级2：结构体名称推断（没有手动设置时）
// 没有调用SetTable()
db.Model(&User{}) // 自动推断为 "users"

// 优先级3：空表名回退到推断
user.SetTable("")
db.Model(user) // 回退推断为 "users"
```

#### 表名推断规则

1. **手动设置优先**：`user.SetTable("table_name")` 
2. **结构体名推断**：`User` → `users`（小写+复数）
3. **复数形式简单**：直接添加"s"后缀

## ⚠️ 重要说明

### 架构设计
TORM模型系统的核心设计原则：
1. **内置db包**: 模型通过`getQueryBuilder()`方法使用`db.Table()`获取查询构建器
2. **封装而非重复**: 所有模型查询方法都是对`db.QueryInterface`的封装
3. **一致性**: 模型查询语法与查询构建器保持一致，都支持传统三参数`Where(field, operator, value)`语法

### 当前版本限制
- ❌ 不支持作用域（Scope）功能 - 推荐使用自定义查询方法
- ❌ 不支持`ToJSON()`方法 - 使用`ToMap()`配合`encoding/json`
- ❌ 不支持`Avg`、`Sum`、`Max`、`Min`等聚合函数 - 使用原生SQL查询
- ❌ 不支持全局作用域 - 在查询时手动添加条件

### 支持的查询方式
- ✅ **传统三参数**: `Where(field, operator, value)` 
- ✅ **参数化查询**: `Where(condition, args...)` 
- ✅ **原生SQL条件**: `WhereRaw(sql, bindings...)`
- ✅ **OR条件**: `OrWhere(...)` 支持参数化和传统方式

### 全面适配db包功能
- ✅ **字段选择**: `Select()`, `SelectRaw()`, `Distinct()`
- ✅ **连接查询**: `Join()`, `LeftJoin()`, `RightJoin()`, `InnerJoin()`
- ✅ **分组排序**: `GroupBy()`, `Having()`, `OrderBy()`, `OrderByRaw()`
- ✅ **数据操作**: `Insert()`, `InsertBatch()`, `Update()`, `Delete()`
- ✅ **查询执行**: `Find()`, `First()`, `Get()`, `Count()`, `CheckExists()`
- ✅ **工具方法**: `ToSQL()`, `Clone()`, `Paginate()`

### 推荐使用方式
```go
// ✅ 推荐：参数化查询（更安全、更简洁）
users, err := user.Where("status = ? AND age >= ?", "active", 18).
    OrderBy("created_at", "desc").
    Get()

// ✅ 推荐：混合使用
users, err := user.Where("status", "=", "active").     // 传统方式
    Where("name LIKE ?", "%admin%").                   // 参数化方式
    WhereRaw("created_at > DATE_SUB(NOW(), INTERVAL ? DAY)", 30). // 原生SQL
    Get()

// ✅ 推荐：自定义查询方法
func (u *User) GetActiveAdults() ([]map[string]interface{}, error) {
    return u.Where("status = ? AND age >= ?", "active", 18).Get()
}
```

## 🆕 v1.1.6 增强功能

### AutoMigrate 自动迁移

v1.1.6 的核心新功能，支持根据模型结构体自动创建数据库表：

```go
// 创建模型
type User struct {
    model.BaseModel
    ID        int64  `json:"id" db:"id" primaryKey:"true" autoIncrement:"true"`
    Email     string `json:"email" db:"email" size:"100" unique:"true"`
    Name      string `json:"name" db:"name" size:"50"`
    CreatedAt int64  `json:"created_at" db:"created_at" autoCreateTime:"true"`
}

// 一键创建表结构
func NewUser() *User {
    user := &User{}
    user.BaseModel = *model.NewBaseModelWithAutoDetect(user)
    user.SetTable("users")
    user.AutoMigrate() // 自动创建表
    return user
}
```

### 新增WHERE查询方法

所有新增的查询方法都支持模型链式调用：

```go
// NULL值查询
activeUsers := user.WhereNotNull("email").WhereNull("deleted_at")

// 范围查询
adultUsers := user.WhereBetween("age", []interface{}{18, 65}).
    WhereNotBetween("score", []interface{}{0, 60})

// 子查询存在性检查
usersWithOrders := user.WhereExists("SELECT 1 FROM orders WHERE orders.user_id = users.id")

// 高级排序
randomUsers := user.OrderRand().Limit(10)
priorityUsers := user.OrderField("status", []interface{}{"premium", "active"}, "asc")

// 原生字段表达式
userStats := user.FieldRaw("COUNT(*) as total").GroupBy("city")
```

### 增强的模型创建

新的 `NewBaseModelWithAutoDetect` 简化了模型创建：

```go
// v1.1.6 新方式（推荐）
func NewProduct() *Product {
    product := &Product{}
    product.BaseModel = *model.NewBaseModelWithAutoDetect(product)
    product.SetTable("products")
    return product
}

// 旧方式（仍然支持）
func NewProductOld() *Product {
    product := &Product{BaseModel: *model.NewBaseModel()}
    product.SetTable("products")
    product.DetectConfigFromStruct(product)
    return product
}
```

### 完整的字段类型支持

支持所有主流数据库类型和精确控制：

```go
type CompleteModel struct {
    model.BaseModel
    // 精确数值类型
    Price     float64 `type:"decimal" precision:"10" scale:"2"`
    Count     int8    `type:"tinyint"`
    BigNumber int64   `type:"bigint"`
    
    // 文本类型精确控制
    Title     string `type:"varchar" size:"200"`
    Content   string `type:"text"`
    LongText  string `type:"longtext"`
    FixedCode string `type:"char" size:"10"`
    
    // 二进制和JSON
    Data      []byte                 `type:"blob"`
    Config    map[string]interface{} `type:"json"`
    
    // 时间类型
    BirthDate time.Time `type:"date"`
    EventTime time.Time `type:"timestamp"`
}
```

### 完整链式调用示例

```go
// 复杂查询组合
result := user.WhereNotNull("email").
    WhereBetween("age", []interface{}{25, 45}).
    WhereExists("SELECT 1 FROM profiles WHERE profiles.user_id = users.id").
    OrderField("status", []interface{}{"premium", "active", "trial"}, "asc").
    OrderRand().
    FieldRaw("TIMESTAMPDIFF(YEAR, birth_date, CURDATE()) as calculated_age").
    Limit(50).
    Get()
```

## 🔗 相关文档

- [查询构建器](Query-Builder) - 了解底层查询构建器
- [关联关系](Relationships) - 模型间的关联关系  
- [数据迁移](Migrations) - 数据库结构管理
- [API参考](API-Reference) - 完整API文档 