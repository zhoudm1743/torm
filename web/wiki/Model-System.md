# 模型系统

TORM 的模型系统提供了强大的 ORM 功能，基于 `BaseModel` 实现数据操作、关联关系、软删除、时间戳管理等特性。

## � 快目录

- [快速开始](#快速开始)
- [模型定义](#模型定义)
- [模型配置](#模型配置)
- [CRUD 操作](#crud-操作)
- [属性管理](#属性管理)
- [查询方法](#查询方法)
- [软删除](#软删除)
- [时间戳](#时间戳)
- [事件钩子](#事件钩子)
- [序列化](#序列化)
- [最佳实践](#最佳实践)

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
    model.BaseModel
    
    ID        int       `json:"id" torm:"primary_key,auto_increment"`
    Username  string    `json:"username" torm:"type:varchar,size:50,unique,index"`
    Email     string    `json:"email" torm:"type:varchar,size:100,unique"`
    Password  string    `json:"password" torm:"type:varchar,size:255"`
    Age       int       `json:"age" torm:"type:int,unsigned,default:0"`
    Status    string    `json:"status" torm:"type:varchar,size:20,default:active"`
    CreatedAt time.Time `json:"created_at" torm:"auto_create_time"`
    UpdatedAt time.Time `json:"updated_at" torm:"auto_update_time"`
}

// GetTableName 返回表名
func (u *User) GetTableName() string {
    return "users"
}
```

### 基础使用

```go
import "github.com/zhoudm1743/torm/model"

func main() {
    // 创建模型实例
    user := model.NewModel(&User{})
    user.SetConnection("default")
    
    // 自动迁移
    user.AutoMigrate(&User{})
    
    // 创建记录
    newUser := model.NewModel("users")
    newUser.SetPrimaryKey("id").SetConnection("default")
    newUser.Fill(map[string]interface{}{
        "username": "zhangsan",
        "email":    "zhangsan@example.com",
        "age":      25,
    })
    newUser.Save()
    
    // 查询记录
    foundUser := model.NewModel("users")
    foundUser.SetConnection("default")
    foundUser.Find(1)
    
    fmt.Println(foundUser.GetAttribute("username"))
}
```

## 🏗️ 模型定义

### NewModel 创建方式

```go
// 方式1: 指定表名
user := model.NewModel("users")
user.SetPrimaryKey("id").SetConnection("default")

// 方式2: 从结构体创建（推荐）
user := model.NewModel(&User{})
user.SetConnection("default")

// 方式3: 表名 + 连接名
user := model.NewModel("users", "mysql_connection")

// 方式4: 结构体 + 配置
config := model.ModelConfig{
    TableName:  "users",
    PrimaryKey: "id",
    Connection: "default",
}
user := model.NewModel(&User{}, config)
```

### BaseModel 核心方法

```go
// 配置方法
SetTable(name string) *BaseModel          // 设置表名
SetPrimaryKey(key string) *BaseModel      // 设置主键
SetConnection(name string) *BaseModel     // 设置连接
GetTableName() string                     // 获取表名
GetPrimaryKey() string                    // 获取主键
GetConnection() string                    // 获取连接

// 属性方法
SetAttribute(key string, value interface{}) Model
GetAttribute(key string) interface{}
SetAttributes(attrs map[string]interface{}) Model
GetAttributes() map[string]interface{}
Fill(data map[string]interface{}) *BaseModel
ClearAttributes() *BaseModel

// 状态方法
IsNew() bool                              // 是否新记录
IsExists() bool                           // 是否已存在
IsDirty() bool                            // 是否有脏数据
MarkAsNew() *BaseModel                    // 标记为新记录
MarkAsExists() *BaseModel                 // 标记为已存在

// 主键方法
GetKey() interface{}                      // 获取主键值
SetKey(key interface{}) *BaseModel        // 设置主键值

// 持久化方法
Save() error                              // 保存（插入或更新）
Delete() error                            // 删除
Find(pk interface{}) error                // 根据主键查找
FindByPK(pk interface{}) error            // 根据主键查找
Reload() error                            // 重新加载

// 查询方法
Query() (*db.QueryBuilder, error)         // 获取查询构建器
Where(args ...interface{}) (*db.QueryBuilder, error)
Get() ([]map[string]interface{}, error)
First() (map[string]interface{}, error)
Count() (int64, error)

// 序列化方法
ToMap() map[string]interface{}
ToJSON() (string, error)
FromJSON(jsonStr string) error
```

## ⚙️ 模型配置

### ModelConfig 配置结构

```go
type ModelConfig struct {
    TableName    string  // 表名
    PrimaryKey   string  // 主键字段（默认: id）
    Connection   string  // 连接名（默认: default）
    Timestamps   bool    // 是否启用时间戳（默认: true）
    CreatedAtCol string  // 创建时间字段（默认: created_at）
    UpdatedAtCol string  // 更新时间字段（默认: updated_at）
    SoftDeletes  bool    // 是否启用软删除（默认: false）
    DeletedAtCol string  // 软删除字段（默认: deleted_at）
}
```

### 配置示例

```go
// 使用默认配置
config := model.DefaultModelConfig()

// 自定义配置
config := model.ModelConfig{
    TableName:    "users",
    PrimaryKey:   "user_id",
    Connection:   "mysql",
    Timestamps:   true,
    CreatedAtCol: "create_time",
    UpdatedAtCol: "update_time",
    SoftDeletes:  true,
    DeletedAtCol: "delete_time",
}

user := model.NewModel(&User{}, config)
```

### 链式配置

```go
user := model.NewModel("users").
    SetPrimaryKey("id").
    SetConnection("default").
    EnableTimestamps().
    EnableSoftDeletes().
    SetCreatedAtField("created_at").
    SetUpdatedAtField("updated_at").
    SetDeletedAtField("deleted_at")
```

## 📝 CRUD 操作

### 创建记录

```go
// 方式1: Fill + Save
user := model.NewModel("users")
user.SetPrimaryKey("id").SetConnection("default")
user.Fill(map[string]interface{}{
    "username": "zhangsan",
    "email":    "zhangsan@example.com",
    "age":      25,
    "status":   "active",
})
err := user.Save()
if err != nil {
    log.Printf("创建失败: %v", err)
}
fmt.Printf("新用户ID: %v\n", user.GetKey())

// 方式2: SetAttribute + Save
user := model.NewModel("users")
user.SetPrimaryKey("id").SetConnection("default")
user.SetAttribute("username", "lisi")
user.SetAttribute("email", "lisi@example.com")
user.SetAttribute("age", 30)
err := user.Save()
```

### 查询记录

```go
// 根据主键查找
user := model.NewModel("users")
user.SetPrimaryKey("id").SetConnection("default")
err := user.Find(1)
if err != nil {
    log.Printf("查找失败: %v", err)
}
fmt.Printf("用户名: %v\n", user.GetAttribute("username"))

// 条件查询
user := model.NewModel("users")
user.SetConnection("default")
query, _ := user.Where("status", "=", "active")
results, err := query.Get()
for _, row := range results {
    fmt.Printf("用户: %v\n", row["username"])
}

// 获取第一条
query, _ := user.Where("age", ">", 18)
first, err := query.First()
```

### 更新记录

```go
// 方式1: 查找后更新
user := model.NewModel("users")
user.SetPrimaryKey("id").SetConnection("default")
user.Find(1)

user.SetAttribute("age", 26)
user.SetAttribute("status", "vip")
err := user.Save()

// 方式2: 直接更新（通过查询构建器）
user := model.NewModel("users")
user.SetConnection("default")
query, _ := user.Where("id", "=", 1)
affected, err := query.Update(map[string]interface{}{
    "age":    26,
    "status": "vip",
})
```

### 删除记录

```go
// 方式1: 查找后删除
user := model.NewModel("users")
user.SetPrimaryKey("id").SetConnection("default")
user.Find(1)
err := user.Delete()

// 方式2: 直接删除（通过查询构建器）
user := model.NewModel("users")
user.SetConnection("default")
query, _ := user.Where("id", "=", 1)
affected, err := query.Delete()
```

## 📊 属性管理

### 设置和获取属性

```go
user := model.NewModel("users")
user.SetPrimaryKey("id").SetConnection("default")

// 单个属性
user.SetAttribute("username", "test")
username := user.GetAttribute("username")

// 批量设置
user.SetAttributes(map[string]interface{}{
    "email": "test@example.com",
    "age":   25,
})

// 获取所有属性
attrs := user.GetAttributes()
for key, value := range attrs {
    fmt.Printf("%s: %v\n", key, value)
}

// 清空属性
user.ClearAttributes()
```

### 脏数据追踪

```go
user := model.NewModel("users")
user.SetPrimaryKey("id").SetConnection("default")
user.Find(1)

// 检查是否有脏数据
fmt.Println(user.IsDirty()) // false

// 修改属性
user.SetAttribute("age", 30)
fmt.Println(user.IsDirty()) // true

// 获取脏数据
dirty := user.GetDirty()
fmt.Printf("修改的字段: %v\n", dirty)

// 检查特定字段
fmt.Println(user.IsDirtyField("age"))   // true
fmt.Println(user.IsDirtyField("email")) // false

// 保存后同步原始数据
user.Save()
fmt.Println(user.IsDirty()) // false
```

### 原始数据

```go
user := model.NewModel("users")
user.SetPrimaryKey("id").SetConnection("default")
user.Find(1)

// 获取原始值
originalAge := user.GetOriginal("age")

// 修改后对比
user.SetAttribute("age", 30)
fmt.Printf("原始年龄: %v, 当前年龄: %v\n", 
    user.GetOriginal("age"), 
    user.GetAttribute("age"))
```

## 🔍 查询方法

### 基础查询

```go
user := model.NewModel("users")
user.SetConnection("default")

// Where 条件
query, _ := user.Where("status", "=", "active")
query, _ = user.Where("age", ">", 18)
query, _ = user.Where("name LIKE ?", "%zhang%")

// OrWhere
query, _ := user.Where("status", "=", "active")
query = query.OrWhere("vip_level", ">", 5)

// WhereIn / WhereNotIn
query, _ := user.WhereIn("status", []interface{}{"active", "vip"})
query, _ = user.WhereNotIn("id", []interface{}{1, 2, 3})

// WhereBetween
query, _ := user.WhereBetween("age", []interface{}{18, 60})

// WhereNull / WhereNotNull
query, _ := user.WhereNull("deleted_at")
query, _ = user.WhereNotNull("email")

// WhereRaw
query, _ := user.WhereRaw("YEAR(created_at) = ?", 2024)
```

### 排序和分页

```go
user := model.NewModel("users")
user.SetConnection("default")

// 排序
query, _ := user.OrderBy("created_at", "DESC")

// 限制和偏移
query, _ := user.Limit(10)
query, _ = user.Offset(20)

// 分页
query, _ := user.Page(2, 10) // 第2页，每页10条

// 组合使用
query, _ := user.Where("status", "=", "active")
query = query.OrderBy("created_at", "DESC").Limit(10)
results, _ := query.Get()
```

### 连接查询

```go
user := model.NewModel("users")
user.SetConnection("default")

// 内连接
query, _ := user.Join("orders", "users.id", "=", "orders.user_id")

// 左连接
query, _ := user.LeftJoin("profiles", "users.id", "=", "profiles.user_id")

// 右连接
query, _ := user.RightJoin("departments", "users.dept_id", "=", "departments.id")
```

### 分组和聚合

```go
user := model.NewModel("users")
user.SetConnection("default")

// 分组
query, _ := user.Select("status", "COUNT(*) as count")
query = query.GroupBy("status")
results, _ := query.Get()

// Having
query, _ := user.Select("status", "COUNT(*) as count")
query = query.GroupBy("status").Having("count", ">", 10)

// 聚合函数
count, _ := user.Count()
```

## 🗑️ 软删除

### 启用软删除

```go
// 方式1: 配置启用
config := model.ModelConfig{
    TableName:    "users",
    SoftDeletes:  true,
    DeletedAtCol: "deleted_at",
}
user := model.NewModel(&User{}, config)

// 方式2: 链式启用
user := model.NewModel("users").
    SetConnection("default").
    EnableSoftDeletes().
    SetDeletedAtField("deleted_at")
```

### 软删除操作

```go
user := model.NewModel("users")
user.SetPrimaryKey("id").SetConnection("default")
user.EnableSoftDeletes()
user.Find(1)

// 软删除
err := user.Delete() // 设置 deleted_at = NOW()

// 检查是否已软删除
if user.IsSoftDeleted() {
    fmt.Println("用户已被软删除")
}

// 恢复软删除
err = user.Restore() // 设置 deleted_at = NULL

// 强制删除（真实删除）
err = user.ForceDelete()
```

### 查询软删除记录

```go
user := model.NewModel("users")
user.SetConnection("default")
user.EnableSoftDeletes()

// 默认查询（排除软删除）
query, _ := user.Query()
results, _ := query.Get() // 只返回未删除的记录

// 包含软删除记录
query, _ := user.RawQuery() // 不自动添加软删除过滤
results, _ := query.Get()

// 只查询软删除记录
query, _ := user.RawQuery()
query = query.WhereNotNull("deleted_at")
results, _ := query.Get()
```

## ⏰ 时间戳

### 自动时间戳

```go
type User struct {
    model.BaseModel
    ID        int       `torm:"primary_key,auto_increment"`
    Username  string    `torm:"type:varchar,size:50"`
    CreatedAt time.Time `torm:"auto_create_time"` // 插入时自动设置
    UpdatedAt time.Time `torm:"auto_update_time"` // 插入和更新时自动设置
}
```

### 配置时间戳

```go
// 启用/禁用时间戳
user := model.NewModel("users")
user.EnableTimestamps()  // 启用
user.DisableTimestamps() // 禁用

// 自定义字段名
user.SetCreatedAtField("create_time")
user.SetUpdatedAtField("update_time")
```

### 时间戳行为

```go
user := model.NewModel("users")
user.SetPrimaryKey("id").SetConnection("default")

// 创建时自动设置 created_at 和 updated_at
user.Fill(map[string]interface{}{
    "username": "test",
})
user.Save()
// created_at = NOW(), updated_at = NOW()

// 更新时自动更新 updated_at
user.Find(1)
user.SetAttribute("username", "test2")
user.Save()
// updated_at = NOW()
```

## 🎣 事件钩子

### 可用钩子

```go
type User struct {
    model.BaseModel
    // ...
}

// 保存前（创建和更新都会触发）
func (u *User) BeforeSave() error {
    // 数据验证、格式化等
    return nil
}

// 保存后
func (u *User) AfterSave() error {
    // 清理缓存、发送通知等
    return nil
}

// 创建前
func (u *User) BeforeCreate() error {
    // 设置默认值、生成UUID等
    return nil
}

// 创建后
func (u *User) AfterCreate() error {
    // 发送欢迎邮件等
    return nil
}

// 更新前
func (u *User) BeforeUpdate() error {
    // 记录变更日志等
    return nil
}

// 更新后
func (u *User) AfterUpdate() error {
    // 同步数据等
    return nil
}

// 删除前
func (u *User) BeforeDelete() error {
    // 检查关联数据等
    return nil
}

// 删除后
func (u *User) AfterDelete() error {
    // 清理关联数据等
    return nil
}
```

### 钩子示例

```go
type User struct {
    model.BaseModel
    ID       int    `torm:"primary_key,auto_increment"`
    Username string `torm:"type:varchar,size:50"`
    Password string `torm:"type:varchar,size:255"`
    Status   string `torm:"type:varchar,size:20"`
}

// 创建前：加密密码
func (u *User) BeforeCreate() error {
    if password := u.GetAttribute("password"); password != nil {
        hashed := hashPassword(password.(string))
        u.SetAttribute("password", hashed)
    }
    return nil
}

// 创建后：发送欢迎邮件
func (u *User) AfterCreate() error {
    email := u.GetAttribute("email")
    if email != nil {
        go sendWelcomeEmail(email.(string))
    }
    return nil
}

// 删除前：检查是否可删除
func (u *User) BeforeDelete() error {
    status := u.GetAttribute("status")
    if status == "admin" {
        return fmt.Errorf("不能删除管理员账户")
    }
    return nil
}
```

## 📤 序列化

### 转换为 Map

```go
user := model.NewModel("users")
user.SetPrimaryKey("id").SetConnection("default")
user.Find(1)

// 转为 map
data := user.ToMap()
fmt.Printf("用户数据: %v\n", data)
```

### JSON 序列化

```go
user := model.NewModel("users")
user.SetPrimaryKey("id").SetConnection("default")
user.Find(1)

// 转为 JSON
jsonStr, err := user.ToJSON()
if err != nil {
    log.Printf("序列化失败: %v", err)
}
fmt.Println(jsonStr)

// 从 JSON 加载
newUser := model.NewModel("users")
newUser.SetPrimaryKey("id").SetConnection("default")
err = newUser.FromJSON(jsonStr)
```

## 💡 最佳实践

### 1. 模型定义规范

```go
// ✅ 推荐：完整的模型定义
type User struct {
    model.BaseModel
    
    // 主键
    ID int64 `json:"id" torm:"primary_key,auto_increment"`
    
    // 业务字段（带约束）
    Username string `json:"username" torm:"type:varchar,size:50,unique,index"`
    Email    string `json:"email" torm:"type:varchar,size:100,unique"`
    Password string `json:"-" torm:"type:varchar,size:255"` // json:"-" 隐藏敏感字段
    
    // 状态字段（带默认值）
    Status   string `json:"status" torm:"type:varchar,size:20,default:active,index"`
    
    // 外键
    DeptID int64 `json:"dept_id" torm:"type:bigint,references:departments.id"`
    
    // 时间戳
    CreatedAt time.Time `json:"created_at" torm:"auto_create_time"`
    UpdatedAt time.Time `json:"updated_at" torm:"auto_update_time"`
}

// 实现 GetTableName 方法
func (u *User) GetTableName() string {
    return "users"
}
```

### 2. 错误处理

```go
// ✅ 完整的错误处理
func createUser(data map[string]interface{}) error {
    user := model.NewModel("users")
    user.SetPrimaryKey("id").SetConnection("default")
    user.Fill(data)
    
    if err := user.Save(); err != nil {
        if strings.Contains(err.Error(), "Duplicate entry") {
            return fmt.Errorf("用户已存在")
        }
        return fmt.Errorf("创建用户失败: %w", err)
    }
    
    return nil
}

func findUser(id int64) (*model.BaseModel, error) {
    user := model.NewModel("users")
    user.SetPrimaryKey("id").SetConnection("default")
    
    if err := user.Find(id); err != nil {
        return nil, fmt.Errorf("用户不存在: %w", err)
    }
    
    return user, nil
}
```

### 3. 查询优化

```go
// ✅ 只查询需要的字段
query, _ := user.Select("id", "username", "email")
results, _ := query.Where("status", "=", "active").Get()

// ✅ 使用索引字段查询
query, _ := user.Where("email", "=", "test@example.com") // email 有索引

// ✅ 分页查询大数据集
query, _ := user.Where("status", "=", "active")
results, _ := query.OrderBy("id", "ASC").Page(1, 100).Get()
```

### 4. 事务使用

```go
import "github.com/zhoudm1743/torm/db"

// ✅ 使用事务保证数据一致性
err := db.Transaction(func(tx db.TransactionInterface) error {
    // 创建用户
    user := model.NewModel("users")
    user.SetPrimaryKey("id").SetConnection("default")
    user.Fill(userData)
    if err := user.Save(); err != nil {
        return err
    }
    
    // 创建用户档案
    profile := model.NewModel("profiles")
    profile.SetPrimaryKey("id").SetConnection("default")
    profile.Fill(map[string]interface{}{
        "user_id": user.GetKey(),
        "avatar":  "default.png",
    })
    if err := profile.Save(); err != nil {
        return err
    }
    
    return nil
}, "default")
```

## 🔗 相关文档

- [TORM 标签系统](Migrations#torm标签完整指南) - 完整的标签参考
- [关联关系](Relationships) - 模型关联
- [查询构建器](Query-Builder) - 高级查询
- [事务处理](Transactions) - 事务管理
- [自动迁移](Migrations) - 数据库迁移
