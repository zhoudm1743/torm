# 最佳实践

本文档汇总了使用 TORM 的最佳实践，包括代码组织、性能优化、安全性和可维护性等方面的建议。

## 📋 目录

- [项目结构](#项目结构)
- [模型设计](#模型设计)
- [查询优化](#查询优化)
- [安全性](#安全性)
- [错误处理](#错误处理)
- [测试策略](#测试策略)

## 🏗️ 项目结构

### 推荐的目录结构

```
project/
├── cmd/                    # 应用入口
│   └── server/
│       └── main.go
├── internal/               # 私有代码
│   ├── models/            # 数据模型
│   │   ├── user.go
│   │   └── post.go
│   ├── services/          # 业务逻辑
│   │   ├── user_service.go
│   │   └── post_service.go
│   ├── repositories/      # 数据访问层
│   │   ├── user_repository.go
│   │   └── post_repository.go
│   └── handlers/          # HTTP处理器
│       ├── user_handler.go
│       └── post_handler.go
├── configs/               # 配置文件
│   ├── config.yaml
│   └── database.yaml
├── migrations/            # 数据库迁移
│   ├── 001_create_users.go
│   └── 002_create_posts.go
└── tests/                 # 测试文件
    ├── models/
    ├── services/
    └── integration/
```

### 代码组织

```go
// internal/models/user.go
package models

import (
    "time"
    "github.com/zhoudm1743/torm/model"
)

type User struct {
    model.BaseModel
    ID        int64     `json:"id" db:"id"`
    Name      string    `json:"name" db:"name"`
    Email     string    `json:"email" db:"email"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func NewUser() *User {
    user := &User{BaseModel: *model.NewBaseModel()}
    user.SetTable("users")
    return user
}

// internal/repositories/user_repository.go
package repositories

type UserRepository struct {
    db db.DatabaseInterface
}

func NewUserRepository(database db.DatabaseInterface) *UserRepository {
    return &UserRepository{db: database}
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
    user := models.NewUser()
    err := user.Where("email", "=", email).First()
    return user, err
}

// internal/services/user_service.go
package services

type UserService struct {
    userRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
    return &UserService{userRepo: userRepo}
}

func (s *UserService) CreateUser(userData map[string]interface{}) (*models.User, error) {
    // 业务逻辑验证
    if err := s.validateUserData(userData); err != nil {
        return nil, err
    }
    
    user := models.NewUser()
    user.Fill(userData)
    
    if err := user.Save(); err != nil {
        return nil, fmt.Errorf("创建用户失败: %w", err)
    }
    
    return user, nil
}
```

## 📊 模型设计

### 模型定义最佳实践

```go
type User struct {
    model.BaseModel
    
    // 基础字段 - 明确的数据类型
    ID    int64  `json:"id" db:"id" primary:"true"`
    Name  string `json:"name" db:"name" validate:"required,min=2,max=50"`
    Email string `json:"email" db:"email" validate:"required,email" unique:"true"`
    
    // 状态字段
    Status   string `json:"status" db:"status" default:"active"`
    IsActive bool   `json:"is_active" db:"is_active" default:"true"`
    
    // 时间戳
    CreatedAt time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
    DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
    
    // 敏感字段
    Password string `json:"-" db:"password"` // 不序列化
}

// 工厂方法
func NewUser() *User {
    user := &User{BaseModel: *model.NewBaseModel()}
    user.SetTable("users")
    user.EnableTimestamps()
    user.EnableSoftDeletes()
    return user
}

// 业务方法
func (u *User) IsAdmin() bool {
    return u.GetAttribute("role") == "admin"
}

func (u *User) CanEdit(resource interface{}) bool {
    // 权限检查逻辑
    return true
}

// 钩子方法
func (u *User) BeforeSave() error {
    // 数据验证和清理
    if email := u.GetAttribute("email"); email != nil {
        u.SetAttribute("email", strings.ToLower(email.(string)))
    }
    return nil
}
```

### 关联关系设计

```go
// 用户关联方法
func (u *User) Posts() *model.HasMany {
    return u.HasMany(&Post{}, "user_id", "id")
}

func (u *User) Profile() *model.HasOne {
    return u.HasOne(&Profile{}, "user_id", "id")
}

func (u *User) Roles() *model.ManyToMany {
    return u.ManyToMany(&Role{}, "user_roles", "user_id", "role_id")
}

// 预加载优化
func (u *User) LoadWithRelations() error {
    return u.With("Profile", "Posts", "Roles").First()
}
```

## 🔍 查询优化

### 查询构建器最佳实践

```go
// ✅ 好的做法：链式查询，选择性字段
users, err := db.Table("users").
    Select("id", "name", "email").
    Where("status", "=", "active").
    Where("created_at", ">", startDate).
    OrderBy("created_at", "desc").
    Limit(20).
    Get()

// ✅ 好的做法：使用预编译查询防止SQL注入
query := db.Table("users").Where("email", "=", userEmail)

// ✅ 好的做法：使用索引字段
query.Where("email", "=", email).      // email有唯一索引
      Where("status", "=", "active")   // status是复合索引的一部分

// ❌ 避免：查询所有字段
// users, err := db.Table("users").Get()

// ❌ 避免：字符串拼接
// sql := "SELECT * FROM users WHERE name = '" + userName + "'"
```

### 分页查询优化

```go
// ✅ 传统分页 - 适合小数据量
result, err := db.Table("users").
    Where("status", "=", "active").
    Paginate(page, 20)

// ✅ 游标分页 - 适合大数据量
users, err := db.Table("users").
    Where("id", ">", lastID).
    Where("status", "=", "active").
    OrderBy("id", "asc").
    Limit(20).
    Get()

// ✅ 计数优化 - 缓存总数
cachedCount, err := cache.Get("active_users_count")
if err != nil {
    count, err := db.Table("users").Where("status", "=", "active").Count()
    cache.Set("active_users_count", count, 5*time.Minute)
}
```

## 🔒 安全性

### SQL注入防护

```go
// ✅ 好的做法：使用参数绑定
users, err := db.Table("users").
    Where("email", "=", userInput).
    Where("status", "=", "active").
    Get()

// ✅ 好的做法：验证输入
func ValidateEmail(email string) error {
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    if !emailRegex.MatchString(email) {
        return errors.New("无效的邮箱格式")
    }
    return nil
}

// ❌ 危险做法：字符串拼接
// sql := "SELECT * FROM users WHERE email = '" + userInput + "'"
```

### 敏感数据处理

```go
type User struct {
    model.BaseModel
    Name     string `json:"name" db:"name"`
    Email    string `json:"email" db:"email"`
    Password string `json:"-" db:"password"`        // 不序列化
    Phone    string `json:"phone" db:"phone" mask:"true"` // 脱敏显示
}

// 密码处理
func (u *User) SetPassword(password string) error {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    u.SetAttribute("password", string(hashedPassword))
    return nil
}

func (u *User) CheckPassword(password string) bool {
    hashedPassword := u.GetAttribute("password").(string)
    return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

// 序列化时脱敏
func (u *User) Serialize() map[string]interface{} {
    data := u.ToMap()
    if phone, ok := data["phone"]; ok && phone != nil {
        data["phone"] = maskPhone(phone.(string))
    }
    delete(data, "password") // 移除敏感字段
    return data
}
```

### 权限控制

```go
// 基于角色的访问控制
type Permission struct {
    model.BaseModel
    ID     int64  `json:"id" db:"id"`
    Name   string `json:"name" db:"name"`
    Action string `json:"action" db:"action"`
    Resource string `json:"resource" db:"resource"`
}

func (u *User) HasPermission(action, resource string) bool {
    permissions, err := u.Permissions().
        Where("action", "=", action).
        Where("resource", "=", resource).
        Get()
    return err == nil && len(permissions) > 0
}

// 数据访问控制
func (u *User) CanAccess(model interface{}) bool {
    switch v := model.(type) {
    case *Post:
        return v.UserID == u.ID || u.IsAdmin()
    default:
        return u.IsAdmin()
    }
}
```

## ⚠️ 错误处理

### 统一错误处理

```go
// 定义错误类型
type AppError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

func (e *AppError) Error() string {
    return e.Message
}

// 错误常量
var (
    ErrUserNotFound     = &AppError{Code: 1001, Message: "用户不存在"}
    ErrInvalidPassword  = &AppError{Code: 1002, Message: "密码错误"}
    ErrEmailExists      = &AppError{Code: 1003, Message: "邮箱已存在"}
)

// 业务层错误处理
func (s *UserService) CreateUser(userData map[string]interface{}) (*models.User, error) {
    // 验证邮箱唯一性
    existing, err := s.userRepo.FindByEmail(userData["email"].(string))
    if err == nil && existing != nil {
        return nil, ErrEmailExists
    }
    
    user := models.NewUser()
    user.Fill(userData)
    
    if err := user.Save(); err != nil {
        log.Printf("创建用户失败: %v", err)
        return nil, &AppError{
            Code:    500,
            Message: "系统错误",
            Details: err.Error(),
        }
    }
    
    return user, nil
}

// HTTP层错误处理
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    user, err := h.userService.CreateUser(userData)
    if err != nil {
        if appErr, ok := err.(*AppError); ok {
            http.Error(w, appErr.Message, appErr.Code)
        } else {
            http.Error(w, "内部服务器错误", 500)
        }
        return
    }
    
    json.NewEncoder(w).Encode(user)
}
```

### 事务错误处理

```go
func (s *UserService) CreateUserWithProfile(userData, profileData map[string]interface{}) error {
    return db.Transaction(func(tx db.TransactionInterface) error {
        // 创建用户
        user := models.NewUser()
        user.Fill(userData)
        if err := user.WithTransaction(tx).Save(); err != nil {
            return fmt.Errorf("创建用户失败: %w", err)
        }
        
        // 创建档案
        profile := models.NewProfile()
        profileData["user_id"] = user.ID
        profile.Fill(profileData)
        if err := profile.WithTransaction(tx).Save(); err != nil {
            return fmt.Errorf("创建用户档案失败: %w", err)
        }
        
        return nil
    })
}
```

## 🧪 测试策略

### 单元测试

```go
func TestUserService_CreateUser(t *testing.T) {
    // 设置测试数据库
    testDB := setupTestDB()
    defer teardownTestDB(testDB)
    
    userRepo := repositories.NewUserRepository(testDB)
    userService := services.NewUserService(userRepo)
    
    tests := []struct {
        name     string
        userData map[string]interface{}
        wantErr  bool
    }{
        {
            name: "创建用户成功",
            userData: map[string]interface{}{
                "name":  "测试用户",
                "email": "test@example.com",
            },
            wantErr: false,
        },
        {
            name: "邮箱重复",
            userData: map[string]interface{}{
                "name":  "重复用户",
                "email": "duplicate@example.com",
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            user, err := userService.CreateUser(tt.userData)
            if tt.wantErr {
                assert.Error(t, err)
                assert.Nil(t, user)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, user)
                assert.Equal(t, tt.userData["name"], user.Name)
            }
        })
    }
}
```

### 集成测试

```go
func TestUserAPI_Integration(t *testing.T) {
    // 设置测试服务器
    server := setupTestServer()
    defer server.Close()
    
    client := &http.Client{}
    
    t.Run("创建用户", func(t *testing.T) {
        userData := map[string]interface{}{
            "name":  "集成测试用户",
            "email": "integration@example.com",
        }
        
        body, _ := json.Marshal(userData)
        req, _ := http.NewRequest("POST", server.URL+"/users", bytes.NewReader(body))
        req.Header.Set("Content-Type", "application/json")
        
        resp, err := client.Do(req)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusCreated, resp.StatusCode)
        
        var user models.User
        json.NewDecoder(resp.Body).Decode(&user)
        assert.Equal(t, userData["name"], user.Name)
    })
}
```

## 📚 代码质量

### 代码注释

```go
// UserService 提供用户相关的业务逻辑
type UserService struct {
    userRepo *repositories.UserRepository
    logger   logger.Interface
}

// CreateUser 创建新用户
// 参数:
//   userData: 包含用户信息的map，必须包含name和email
// 返回:
//   *models.User: 创建的用户对象
//   error: 如果创建失败则返回错误
func (s *UserService) CreateUser(userData map[string]interface{}) (*models.User, error) {
    // 验证输入数据
    if err := s.validateUserData(userData); err != nil {
        return nil, fmt.Errorf("用户数据验证失败: %w", err)
    }
    
    // 检查邮箱唯一性
    if exists, err := s.emailExists(userData["email"].(string)); err != nil {
        return nil, fmt.Errorf("检查邮箱失败: %w", err)
    } else if exists {
        return nil, ErrEmailExists
    }
    
    // 创建用户对象
    user := models.NewUser()
    user.Fill(userData)
    
    // 保存到数据库
    if err := user.Save(); err != nil {
        s.logger.Error("创建用户失败", "error", err, "userData", userData)
        return nil, fmt.Errorf("保存用户失败: %w", err)
    }
    
    s.logger.Info("用户创建成功", "userID", user.ID, "email", user.Email)
    return user, nil
}
```

## 🔗 相关文档

- [查询构建器](Query-Builder) - 查询最佳实践
- [模型系统](Model-System) - 模型设计指南
- [性能优化](Performance) - 性能优化技巧
- [安全指南](Security) - 安全性最佳实践 