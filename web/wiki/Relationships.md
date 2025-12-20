# 关联关系

TORM 提供了完整的模型关联支持，包括 HasOne、HasMany、BelongsTo、BelongsToMany 四种关系类型，支持预加载和延迟加载。

## 📋 目录

- [快速开始](#快速开始)
- [关系类型](#关系类型)
- [定义关联](#定义关联)
- [查询关联](#查询关联)
- [预加载](#预加载)
- [延迟加载](#延迟加载)
- [关联操作](#关联操作)
- [最佳实践](#最佳实践)

## 🚀 快速开始

### 模型定义

```go
package models

import (
    "time"
    "github.com/zhoudm1743/torm/model"
)

// Department 部门模型
type Department struct {
    model.BaseModel
    ID        int       `json:"id" torm:"primary_key,auto_increment"`
    Name      string    `json:"name" torm:"type:varchar,size:100,unique"`
    Budget    float64   `json:"budget" torm:"type:decimal,precision:12,scale:2"`
    CreatedAt time.Time `json:"created_at" torm:"auto_create_time"`
}

func (d *Department) GetTableName() string {
    return "departments"
}

// User 用户模型
type User struct {
    model.BaseModel
    ID        int       `json:"id" torm:"primary_key,auto_increment"`
    Username  string    `json:"username" torm:"type:varchar,size:50,unique"`
    Email     string    `json:"email" torm:"type:varchar,size:100"`
    DeptID    int       `json:"dept_id" torm:"type:int,references:departments.id"`
    CreatedAt time.Time `json:"created_at" torm:"auto_create_time"`
}

func (u *User) GetTableName() string {
    return "users"
}

// Role 角色模型
type Role struct {
    model.BaseModel
    ID   int    `json:"id" torm:"primary_key,auto_increment"`
    Name string `json:"name" torm:"type:varchar,size:50,unique"`
}

func (r *Role) GetTableName() string {
    return "roles"
}

// Profile 用户档案模型
type Profile struct {
    model.BaseModel
    ID     int    `json:"id" torm:"primary_key,auto_increment"`
    UserID int    `json:"user_id" torm:"type:int,references:users.id"`
    Avatar string `json:"avatar" torm:"type:varchar,size:255"`
    Bio    string `json:"bio" torm:"type:text"`
}

func (p *Profile) GetTableName() string {
    return "profiles"
}
```

## 🔗 关系类型

### 1. HasOne (一对一)

用户有一个档案：

```go
// 在 User 模型中定义
func (u *User) Profile() *model.HasOne {
    return u.HasOne(&Profile{}, "user_id", "id")
}

// 使用
user := model.NewModel(&User{})
user.SetConnection("default")
user.Find(1)

profile, err := user.Profile().First()
if err == nil && profile != nil {
    fmt.Printf("头像: %v\n", profile["avatar"])
}
```

### 2. HasMany (一对多)

部门有多个用户：

```go
// 在 Department 模型中定义
func (d *Department) Users() *model.HasMany {
    return d.HasMany(&User{}, "dept_id", "id")
}

// 使用
dept := model.NewModel(&Department{})
dept.SetConnection("default")
dept.Find(1)

users, err := dept.Users().Get()
if err == nil {
    fmt.Printf("部门有 %d 个用户\n", len(users))
}

// 带条件查询
activeUsers, err := dept.Users().
    Where("status", "=", "active").
    OrderBy("created_at", "DESC").
    Get()
```

### 3. BelongsTo (反向关联)

用户属于部门：

```go
// 在 User 模型中定义
func (u *User) Department() *model.BelongsTo {
    return u.BelongsTo(&Department{}, "dept_id", "id")
}

// 使用
user := model.NewModel(&User{})
user.SetConnection("default")
user.Find(1)

dept, err := user.Department().First()
if err == nil && dept != nil {
    fmt.Printf("所属部门: %v\n", dept["name"])
}
```

### 4. BelongsToMany (多对多)

用户和角色的多对多关系：

```go
// 在 User 模型中定义
func (u *User) Roles() *model.BelongsToMany {
    return u.BelongsToMany(&Role{}, "user_roles", "user_id", "role_id")
}

// 在 Role 模型中定义
func (r *Role) Users() *model.BelongsToMany {
    return r.BelongsToMany(&User{}, "user_roles", "role_id", "user_id")
}

// 使用
user := model.NewModel(&User{})
user.SetConnection("default")
user.Find(1)

roles, err := user.Roles().Get()
if err == nil {
    for _, role := range roles {
        fmt.Printf("角色: %v\n", role["name"])
    }
}
```

## 📝 定义关联

### HasOne 关联

```go
// 基础定义
// HasOne(relatedModel, foreignKey, localKey)
func (u *User) Profile() *model.HasOne {
    return u.HasOne(&Profile{}, "user_id", "id")
}

// 参数说明：
// - relatedModel: 关联模型实例
// - foreignKey: 关联表中的外键字段（profile.user_id）
// - localKey: 本表的主键字段（user.id）
```

### HasMany 关联

```go
// 基础定义
// HasMany(relatedModel, foreignKey, localKey)
func (d *Department) Users() *model.HasMany {
    return d.HasMany(&User{}, "dept_id", "id")
}

// 带默认条件
func (d *Department) ActiveUsers() *model.HasMany {
    return d.HasMany(&User{}, "dept_id", "id").
        Where("status", "=", "active")
}

// 带默认排序
func (d *Department) RecentUsers() *model.HasMany {
    return d.HasMany(&User{}, "dept_id", "id").
        OrderBy("created_at", "DESC")
}
```

### BelongsTo 关联

```go
// 基础定义
// BelongsTo(relatedModel, foreignKey, ownerKey)
func (u *User) Department() *model.BelongsTo {
    return u.BelongsTo(&Department{}, "dept_id", "id")
}

// 参数说明：
// - relatedModel: 关联模型实例
// - foreignKey: 本表中的外键字段（user.dept_id）
// - ownerKey: 关联表的主键字段（department.id）
```

### BelongsToMany 关联

```go
// 基础定义
// BelongsToMany(relatedModel, pivotTable, foreignKey, relatedKey)
func (u *User) Roles() *model.BelongsToMany {
    return u.BelongsToMany(&Role{}, "user_roles", "user_id", "role_id")
}

// 参数说明：
// - relatedModel: 关联模型实例
// - pivotTable: 中间表名
// - foreignKey: 中间表中本模型的外键（user_roles.user_id）
// - relatedKey: 中间表中关联模型的外键（user_roles.role_id）

// 中间表结构示例：
// CREATE TABLE user_roles (
//     user_id INT NOT NULL,
//     role_id INT NOT NULL,
//     PRIMARY KEY (user_id, role_id)
// );
```

## 🔍 查询关联

### 基础查询

```go
user := model.NewModel(&User{})
user.SetConnection("default")
user.Find(1)

// 获取所有关联记录
roles, err := user.Roles().Get()

// 获取第一条关联记录
dept, err := user.Department().First()

// 计数
roleCount, err := user.Roles().Count()

// 检查是否存在
hasRoles, err := user.Roles().Exists()
```

### 条件查询

```go
dept := model.NewModel(&Department{})
dept.SetConnection("default")
dept.Find(1)

// 带条件的关联查询
activeUsers, err := dept.Users().
    Where("status", "=", "active").
    Where("age", ">", 18).
    Get()

// 排序和限制
recentUsers, err := dept.Users().
    OrderBy("created_at", "DESC").
    Limit(10).
    Get()

// 选择特定字段
users, err := dept.Users().
    Select("id", "username", "email").
    Get()
```

### 聚合查询

```go
dept := model.NewModel(&Department{})
dept.SetConnection("default")
dept.Find(1)

// 计数
userCount, err := dept.Users().Count()

// 求和
totalSalary, err := dept.Users().Sum("salary")

// 平均值
avgAge, err := dept.Users().Avg("age")

// 最大/最小值
maxSalary, err := dept.Users().Max("salary")
minAge, err := dept.Users().Min("age")
```

## ⚡ 预加载

预加载可以解决 N+1 查询问题，一次性加载所有关联数据。

### 基础预加载

```go
// 使用 With 方法预加载
user := model.NewModel(&User{})
user.SetConnection("default")

// 预加载单个关联
results, err := user.With("Department").Get()

// 预加载多个关联
results, err := user.With("Department", "Profile", "Roles").Get()

// 访问预加载的数据
for _, row := range results {
    fmt.Printf("用户: %v\n", row["username"])
    if dept, ok := row["Department"].(map[string]interface{}); ok {
        fmt.Printf("部门: %v\n", dept["name"])
    }
}
```

### 带约束的预加载

```go
user := model.NewModel(&User{})
user.SetConnection("default")

// 预加载时添加约束条件
results, err := user.WithConstraint("Roles", func(q *db.QueryBuilder) *db.QueryBuilder {
    return q.Where("status", "=", "active").
        OrderBy("name", "ASC")
}).Get()
```

### 嵌套预加载

```go
// 预加载嵌套关联（如：用户 -> 部门 -> 公司）
results, err := user.With("Department.Company").Get()

// 多层嵌套
results, err := user.With("Posts.Comments.User").Get()
```

## 🔄 延迟加载

延迟加载在需要时才加载关联数据。

### Load 方法

```go
user := model.NewModel(&User{})
user.SetConnection("default")
user.Find(1)

// 延迟加载关联
err := user.Load("Department", "Roles")
if err != nil {
    log.Printf("加载关联失败: %v", err)
}

// 访问已加载的关联
dept := user.GetRelation("Department")
roles := user.GetRelation("Roles")
```

### GetRelationOrLoad 方法

```go
user := model.NewModel(&User{})
user.SetConnection("default")
user.Find(1)

// 获取关联，如果未加载则自动加载
dept, err := user.GetRelationOrLoad("Department")
if err != nil {
    log.Printf("加载部门失败: %v", err)
}
```

### 检查关联是否已加载

```go
user := model.NewModel(&User{})
user.SetConnection("default")
user.Find(1)

// 检查是否已加载
if user.HasRelationLoaded("Department") {
    dept := user.GetRelation("Department")
    // 使用已加载的数据
} else {
    // 需要加载
    user.Load("Department")
}
```

## 🔧 关联操作

### 设置关联数据

```go
user := model.NewModel(&User{})
user.SetConnection("default")
user.Find(1)

// 手动设置关联数据
user.SetRelation("Department", map[string]interface{}{
    "id":   1,
    "name": "技术部",
})

// 获取关联数据
dept := user.GetRelation("Department")
```

### 通过关联创建记录

```go
dept := model.NewModel(&Department{})
dept.SetConnection("default")
dept.Find(1)

// 通过关联创建用户
query := dept.Users()
newUserID, err := query.Insert(map[string]interface{}{
    "username": "newuser",
    "email":    "newuser@example.com",
    "dept_id":  dept.GetKey(), // 自动设置外键
})
```

### 通过关联更新记录

```go
dept := model.NewModel(&Department{})
dept.SetConnection("default")
dept.Find(1)

// 更新部门下所有用户的状态
affected, err := dept.Users().Update(map[string]interface{}{
    "status": "inactive",
})
```

### 通过关联删除记录

```go
dept := model.NewModel(&Department{})
dept.SetConnection("default")
dept.Find(1)

// 删除部门下所有用户
affected, err := dept.Users().Delete()
```

## 📊 完整示例

### 用户-部门-角色系统

```go
package main

import (
    "fmt"
    "github.com/zhoudm1743/torm/db"
    "github.com/zhoudm1743/torm/model"
)

// 模型定义
type Department struct {
    model.BaseModel
    ID   int    `torm:"primary_key,auto_increment"`
    Name string `torm:"type:varchar,size:100"`
}

func (d *Department) GetTableName() string { return "departments" }

func (d *Department) Users() *model.HasMany {
    return d.HasMany(&User{}, "dept_id", "id")
}

type User struct {
    model.BaseModel
    ID       int    `torm:"primary_key,auto_increment"`
    Username string `torm:"type:varchar,size:50"`
    DeptID   int    `torm:"type:int"`
}

func (u *User) GetTableName() string { return "users" }

func (u *User) Department() *model.BelongsTo {
    return u.BelongsTo(&Department{}, "dept_id", "id")
}

func (u *User) Roles() *model.BelongsToMany {
    return u.BelongsToMany(&Role{}, "user_roles", "user_id", "role_id")
}

type Role struct {
    model.BaseModel
    ID   int    `torm:"primary_key,auto_increment"`
    Name string `torm:"type:varchar,size:50"`
}

func (r *Role) GetTableName() string { return "roles" }

// 使用示例
func main() {
    // 配置数据库
    db.AddConnection("default", &db.Config{
        Driver:   "mysql",
        Host:     "localhost",
        Database: "myapp",
        Username: "root",
        Password: "password",
    })
    
    // 查询用户及其部门
    user := model.NewModel(&User{})
    user.SetConnection("default")
    user.Find(1)
    
    // 获取用户所属部门
    dept, err := user.Department().First()
    if err == nil && dept != nil {
        fmt.Printf("用户 %v 属于部门: %v\n", 
            user.GetAttribute("username"), 
            dept["name"])
    }
    
    // 获取用户的所有角色
    roles, err := user.Roles().Get()
    if err == nil {
        fmt.Printf("用户有 %d 个角色:\n", len(roles))
        for _, role := range roles {
            fmt.Printf("  - %v\n", role["name"])
        }
    }
    
    // 查询部门及其用户（预加载）
    deptModel := model.NewModel(&Department{})
    deptModel.SetConnection("default")
    
    depts, err := deptModel.With("Users").Get()
    if err == nil {
        for _, d := range depts {
            fmt.Printf("部门: %v\n", d["name"])
            if users, ok := d["Users"].([]map[string]interface{}); ok {
                for _, u := range users {
                    fmt.Printf("  - 用户: %v\n", u["username"])
                }
            }
        }
    }
}
```

## 💡 最佳实践

### 1. 避免 N+1 查询

```go
// ❌ 错误：N+1 查询问题
users, _ := userModel.Get()
for _, user := range users {
    // 每次循环都会执行一次查询
    dept, _ := user.Department().First()
}

// ✅ 正确：使用预加载
users, _ := userModel.With("Department").Get()
for _, user := range users {
    // 直接访问已加载的数据，无额外查询
    dept := user["Department"]
}
```

### 2. 选择性加载

```go
// ✅ 只加载需要的字段
users, _ := userModel.With("Department").
    Select("id", "username", "dept_id").
    Get()

// ✅ 带约束的预加载
users, _ := userModel.WithConstraint("Roles", func(q *db.QueryBuilder) *db.QueryBuilder {
    return q.Select("id", "name").Where("status", "=", "active")
}).Get()
```

### 3. 合理使用延迟加载

```go
// ✅ 条件性加载
user := model.NewModel(&User{})
user.SetConnection("default")
user.Find(1)

// 只在需要时加载
if needDepartmentInfo {
    user.Load("Department")
}

if needRoles {
    user.Load("Roles")
}
```

### 4. 关联方法命名规范

```go
// ✅ 清晰的方法命名
func (u *User) Department() *model.BelongsTo { ... }      // 单数：一对一/多对一
func (u *User) Posts() *model.HasMany { ... }             // 复数：一对多
func (u *User) Roles() *model.BelongsToMany { ... }       // 复数：多对多

// ✅ 带条件的关联使用描述性名称
func (u *User) ActivePosts() *model.HasMany {
    return u.HasMany(&Post{}, "user_id", "id").
        Where("status", "=", "active")
}

func (u *User) PublishedPosts() *model.HasMany {
    return u.HasMany(&Post{}, "user_id", "id").
        Where("status", "=", "published")
}
```

### 5. 错误处理

```go
// ✅ 完整的错误处理
user := model.NewModel(&User{})
user.SetConnection("default")

if err := user.Find(1); err != nil {
    return fmt.Errorf("用户不存在: %w", err)
}

dept, err := user.Department().First()
if err != nil {
    return fmt.Errorf("获取部门失败: %w", err)
}

if dept == nil {
    log.Println("用户未分配部门")
}
```

## 🔗 相关文档

- [模型系统](Model-System) - 模型基础功能
- [查询构建器](Query-Builder) - 高级查询
- [事务处理](Transactions) - 关联操作的事务支持
- [缓存系统](Caching) - 关联数据缓存
