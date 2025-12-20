# 最佳实践

本文档汇总了使用 TORM 的最佳实践，包括代码组织、性能优化和安全性建议。

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
│   │   └── user_service.go
│   └── repositories/      # 数据访问层
│       └── user_repository.go
├── configs/               # 配置文件
│   └── database.yaml
└── migrations/            # 数据库迁移
    └── 001_create_users.go
```

### 代码组织示例

```go
// internal/models/user.go
package models

import "github.com/zhoudm1743/torm/model"

type User struct {
    model.BaseModel
}

func NewUser() *User {
    user := &User{BaseModel: *model.NewBaseModel()}
    user.SetTable("users")
    user.EnableTimestamps()
    return user
}

// internal/repositories/user_repository.go
package repositories

type UserRepository struct{}

func (r *UserRepository) FindByEmail(email string) (map[string]interface{}, error) {
    return torm.Table("users").
        Where("email", "=", email).
        First()
}

func (r *UserRepository) FindActive() ([]map[string]interface{}, error) {
    return torm.Table("users").
        Where("status", "=", "active").
        OrderBy("created_at", "desc").
        Get()
}

// internal/services/user_service.go
package services

type UserService struct {
    userRepo *repositories.UserRepository
}

func (s *UserService) CreateUser(data map[string]interface{}) (int64, error) {
    // 业务逻辑验证
    if err := s.validateUserData(data); err != nil {
        return 0, err
    }
    
    return torm.Table("users").Insert(data)
}
```

## 🔍 查询最佳实践

### 选择性字段查询

```go
// ✅ 好的做法：只查询需要的字段
users, _ := torm.Table("users").
    Select("id", "name", "email").
    Where("status", "=", "active").
    Get()

// ❌ 避免：查询所有字段
// users, _ := torm.Table("users").Get()
```

### 使用参数化查询

```go
// ✅ 好的做法：使用参数绑定
users, _ := torm.Table("users").
    Where("email", "=", userEmail).
    Where("status", "=", "active").
    Get()

// ❌ 危险做法：字符串拼接
// sql := "SELECT * FROM users WHERE email = '" + userEmail + "'"
```

### 分页查询优化

```go
// ✅ 小数据量：传统分页
result, _ := torm.Table("users").
    Where("status", "=", "active").
    Paginate(page, 20)

// ✅ 大数据量：游标分页
users, _ := torm.Table("users").
    Where("id", ">", lastID).
    Where("status", "=", "active").
    OrderBy("id", "asc").
    Limit(20).
    Get()
```

## 🔒 安全性

### SQL 注入防护

```go
// ✅ 好的做法：使用参数绑定
users, _ := torm.Table("users").
    Where("email", "=", userInput).
    Get()

// ✅ 好的做法：使用 WhereIn
users, _ := torm.Table("users").
    WhereIn("id", userIDs).
    Get()

// ❌ 危险做法：直接拼接用户输入
// query.WhereRaw("email = '" + userInput + "'")
```

### 敏感数据处理

```go
type User struct {
    model.BaseModel
    Name     string `json:"name"`
    Email    string `json:"email"`
    Password string `json:"-"` // 不序列化到 JSON
}

// 密码加密
func (u *User) SetPassword(password string) error {
    hashedPassword, err := bcrypt.GenerateFromPassword(
        []byte(password), 
        bcrypt.DefaultCost,
    )
    if err != nil {
        return err
    }
    u.SetAttribute("password", string(hashedPassword))
    return nil
}

// 密码验证
func (u *User) CheckPassword(password string) bool {
    hashedPassword := u.GetAttribute("password").(string)
    err := bcrypt.CompareHashAndPassword(
        []byte(hashedPassword), 
        []byte(password),
    )
    return err == nil
}
```

## ⚠️ 错误处理

### 统一错误处理

```go
// 定义错误类型
var (
    ErrUserNotFound = errors.New("用户不存在")
    ErrEmailExists  = errors.New("邮箱已存在")
)

// 业务层错误处理
func (s *UserService) CreateUser(data map[string]interface{}) (int64, error) {
    // 检查邮箱是否存在
    existing, err := s.userRepo.FindByEmail(data["email"].(string))
    if err == nil && existing != nil {
        return 0, ErrEmailExists
    }
    
    id, err := torm.Table("users").Insert(data)
    if err != nil {
        log.Printf("创建用户失败: %v", err)
        return 0, fmt.Errorf("创建用户失败: %w", err)
    }
    
    return id, nil
}
```

### 事务错误处理

```go
func (s *UserService) CreateUserWithProfile(userData, profileData map[string]interface{}) error {
    return torm.Transaction(func(tx torm.TransactionInterface) error {
        // 创建用户
        result, err := tx.Exec(
            "INSERT INTO users (name, email) VALUES (?, ?)",
            userData["name"], userData["email"],
        )
        if err != nil {
            return fmt.Errorf("创建用户失败: %w", err)
        }
        
        userID, _ := result.LastInsertId()
        
        // 创建档案
        _, err = tx.Exec(
            "INSERT INTO profiles (user_id, bio) VALUES (?, ?)",
            userID, profileData["bio"],
        )
        if err != nil {
            return fmt.Errorf("创建档案失败: %w", err)
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
    torm.AddConnection("test", &torm.Config{
        Driver:   "sqlite",
        Database: ":memory:",
    })
    
    // 创建测试表
    torm.Table("users", "test").Exec(`
        CREATE TABLE users (
            id INTEGER PRIMARY KEY,
            name TEXT,
            email TEXT UNIQUE
        )
    `)
    
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
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            id, err := torm.Table("users", "test").Insert(tt.userData)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Greater(t, id, int64(0))
            }
        })
    }
}
```

## 💡 其他建议

### 1. 使用连接池

```go
config := &torm.Config{
    MaxOpenConns:    100,
    MaxIdleConns:    10,
    ConnMaxLifetime: time.Hour,
}
```

### 2. 合理使用缓存

```go
// 缓存不常变化的数据
config, _ := torm.Table("config").
    Cache(time.Hour).
    First()

// 更新后清除缓存
torm.Table("config").Where("id", "=", 1).Update(data)
torm.ClearCacheByTags("config")
```

### 3. 使用预加载避免 N+1

```go
// ✅ 使用预加载
users, _ := user.With("Posts", "Profile").Get()

// ❌ 避免循环查询
// for _, user := range users {
//     posts, _ := torm.Table("posts").Where("user_id", "=", user["id"]).Get()
// }
```

### 4. 日志记录

```go
// 开发环境启用 SQL 日志
if isDevelopment {
    torm.EnableSQLLogging()
}

// 生产环境只记录错误
if isProduction {
    torm.SetSQLLogging(logger.ERROR, true)
}
```

## 🔗 相关文档

- [查询构建器](Query-Builder) - 查询最佳实践
- [模型系统](Model-System) - 模型设计指南
- [性能优化](Performance) - 性能优化技巧
- [安全指南](Troubleshooting) - 安全性最佳实践
