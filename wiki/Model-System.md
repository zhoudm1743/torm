# 模型系统

TORM 采用 Active Record 模式的模型系统，让你可以用面向对象的方式操作数据库。每个模型对应一个数据库表，模型实例对应表中的一行记录。

## 📋 目录

- [模型定义](#模型定义)
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
    "github.com/zhoudm1743/torm/pkg/model"
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
    user := &User{
        BaseModel: *model.NewBaseModel(),
    }
    user.SetTable("users")      // 设置表名
    user.SetConnection("default") // 设置数据库连接
    return user
}
```

## 📊 模型定义

### 字段标签

```go
type User struct {
    model.BaseModel
    ID       int64     `json:"id" db:"id" primary:"true"`           // 主键标签
    Name     string    `json:"name" db:"name" validate:"required"`  // 验证标签
    Email    string    `json:"email" db:"email" unique:"true"`      // 唯一索引
    Password string    `json:"-" db:"password"`                     // 隐藏字段
    Profile  string    `json:"profile" db:"profile" type:"json"`    // JSON字段
    Avatar   *string   `json:"avatar" db:"avatar"`                  // 可空字段
    CreatedAt time.Time `json:"created_at" db:"created_at" auto:"true"` // 自动时间戳
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

// 带条件查找
user := NewUser()
err := user.Where("email", "=", "user@example.com").First()

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

// 批量更新
user := NewUser()
affected, err := user.Where("status", "=", "inactive").
    Update(map[string]interface{}{
        "status": "archived",
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

// 条件查询
users, err := user.Where("age", ">", 18).
    Where("status", "=", "active").
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

// 检查存在
exists, err := user.Where("email", "=", "test@example.com").Exists()

// 最大值、最小值
maxAge, err := user.Max("age")
minAge, err := user.Min("age")

// 求和、平均值
totalAge, err := user.Sum("age")
avgAge, err := user.Avg("age")
```

### 高级查询

```go
user := NewUser()

// 原生SQL
users, err := user.WhereRaw("YEAR(created_at) = ?", 2023).Get()

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

### 定义作用域

```go
type User struct {
    model.BaseModel
    // ... 字段定义
}

// 定义作用域方法
func (u *User) ScopeActive(query db.QueryInterface) db.QueryInterface {
    return query.Where("status", "=", "active")
}

func (u *User) ScopeAdult(query db.QueryInterface) db.QueryInterface {
    return query.Where("age", ">=", 18)
}

func (u *User) ScopeByCity(query db.QueryInterface, city string) db.QueryInterface {
    return query.Where("city", "=", city)
}
```

### 使用作用域

```go
user := NewUser()

// 使用单个作用域
users, err := user.Active().Get()

// 链式使用多个作用域
users, err := user.Active().Adult().Get()

// 带参数的作用域
users, err := user.Active().ByCity("北京").Get()

// 与其他查询条件结合
users, err := user.Active().
    Where("vip_level", ">", 3).
    OrderBy("created_at", "desc").
    Get()
```

### 全局作用域

```go
type User struct {
    model.BaseModel
    // ... 字段定义
}

func NewUser() *User {
    user := &User{BaseModel: *model.NewBaseModel()}
    user.SetTable("users")
    
    // 添加全局作用域（自动应用到所有查询）
    user.AddGlobalScope("active", func(query db.QueryInterface) db.QueryInterface {
        return query.Where("status", "!=", "deleted")
    })
    
    return user
}

// 移除全局作用域
user := NewUser()
users, err := user.WithoutGlobalScope("active").Get() // 包含已删除用户
```

## 📤 序列化

### JSON序列化

```go
user := NewUser()
err := user.Find(1)

// 转换为JSON
jsonData, err := user.ToJSON()

// 转换为Map
userData := user.ToMap()

// 隐藏敏感字段
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

## 🔗 相关文档

- [查询构建器](Query-Builder) - 了解底层查询构建器
- [关联关系](Relationships) - 模型间的关联关系
- [数据迁移](Migrations) - 数据库结构管理
- [验证系统](Validation) - 数据验证功能 