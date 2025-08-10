package model

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"torm/pkg/db"
)

// User 用户模型
type User struct {
	*BaseModel
}

// NewUser 创建用户模型实例
func NewUser() *User {
	user := &User{
		BaseModel: NewBaseModel(),
	}
	user.SetTable("users")
	return user
}

// NewUserWithData 使用数据创建用户模型
func NewUserWithData(data map[string]interface{}) *User {
	user := NewUser()
	user.Fill(data)
	if user.GetAttribute("id") != nil {
		user.isNew = false
		user.exists = true
		user.syncOriginal()
	}
	return user
}

// 属性访问器

// GetID 获取用户ID
func (u *User) GetID() interface{} {
	return u.GetAttribute("id")
}

// SetID 设置用户ID
func (u *User) SetID(id interface{}) *User {
	u.SetAttribute("id", id)
	return u
}

// GetName 获取用户名
func (u *User) GetName() string {
	if name := u.GetAttribute("name"); name != nil {
		return name.(string)
	}
	return ""
}

// SetName 设置用户名
func (u *User) SetName(name string) *User {
	u.SetAttribute("name", name)
	return u
}

// GetEmail 获取邮箱
func (u *User) GetEmail() string {
	if email := u.GetAttribute("email"); email != nil {
		return email.(string)
	}
	return ""
}

// SetEmail 设置邮箱
func (u *User) SetEmail(email string) *User {
	u.SetAttribute("email", email)
	return u
}

// GetAge 获取年龄
func (u *User) GetAge() int {
	if age := u.GetAttribute("age"); age != nil {
		switch v := age.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return 0
}

// SetAge 设置年龄
func (u *User) SetAge(age int) *User {
	u.SetAttribute("age", age)
	return u
}

// GetStatus 获取状态
func (u *User) GetStatus() string {
	if status := u.GetAttribute("status"); status != nil {
		return status.(string)
	}
	return ""
}

// SetStatus 设置状态
func (u *User) SetStatus(status string) *User {
	u.SetAttribute("status", status)
	return u
}

// GetCreatedAt 获取创建时间
func (u *User) GetCreatedAt() time.Time {
	if createdAt := u.GetAttribute("created_at"); createdAt != nil {
		switch v := createdAt.(type) {
		case time.Time:
			return v
		case string:
			if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}

// GetUpdatedAt 获取更新时间
func (u *User) GetUpdatedAt() time.Time {
	if updatedAt := u.GetAttribute("updated_at"); updatedAt != nil {
		switch v := updatedAt.(type) {
		case time.Time:
			return v
		case string:
			if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}

// 业务方法

// IsActive 检查用户是否激活
func (u *User) IsActive() bool {
	return u.GetStatus() == "active"
}

// IsPending 检查用户是否待审核
func (u *User) IsPending() bool {
	return u.GetStatus() == "pending"
}

// IsInactive 检查用户是否非激活
func (u *User) IsInactive() bool {
	return u.GetStatus() == "inactive"
}

// Activate 激活用户
func (u *User) Activate() *User {
	u.SetStatus("active")
	return u
}

// Deactivate 停用用户
func (u *User) Deactivate() *User {
	u.SetStatus("inactive")
	return u
}

// IsAdult 检查是否成年
func (u *User) IsAdult() bool {
	return u.GetAge() >= 18
}

// GetDisplayName 获取显示名称
func (u *User) GetDisplayName() string {
	name := u.GetName()
	if name == "" {
		return u.GetEmail()
	}
	return name
}

// 验证方法

// Validate 验证用户数据
func (u *User) Validate() error {
	var errors []string

	// 验证姓名
	name := u.GetName()
	if name == "" {
		errors = append(errors, "姓名不能为空")
	} else if len(name) > 100 {
		errors = append(errors, "姓名长度不能超过100个字符")
	}

	// 验证邮箱
	email := u.GetEmail()
	if email == "" {
		errors = append(errors, "邮箱不能为空")
	} else if !u.isValidEmail(email) {
		errors = append(errors, "邮箱格式不正确")
	}

	// 验证年龄
	age := u.GetAge()
	if age < 0 || age > 150 {
		errors = append(errors, "年龄必须在0-150之间")
	}

	// 验证状态
	status := u.GetStatus()
	if status != "" && !u.isValidStatus(status) {
		errors = append(errors, "状态值不正确")
	}

	if len(errors) > 0 {
		return fmt.Errorf("验证失败: %s", strings.Join(errors, ", "))
	}

	return nil
}

// isValidEmail 验证邮箱格式
func (u *User) isValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	regex := regexp.MustCompile(pattern)
	return regex.MatchString(email)
}

// isValidStatus 验证状态值
func (u *User) isValidStatus(status string) bool {
	validStatuses := []string{"active", "inactive", "pending"}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// 事件钩子重写

// BeforeSave 保存前验证
func (u *User) BeforeSave() error {
	return u.Validate()
}

// BeforeCreate 创建前钩子
func (u *User) BeforeCreate() error {
	// 设置默认状态
	if u.GetStatus() == "" {
		u.SetStatus("active")
	}

	// 设置默认年龄
	if u.GetAge() == 0 {
		u.SetAge(18)
	}

	fmt.Printf("正在创建用户: %s (%s)\n", u.GetName(), u.GetEmail())
	return nil
}

// AfterCreate 创建后钩子
func (u *User) AfterCreate() error {
	fmt.Printf("用户创建成功: ID=%v, 姓名=%s\n", u.GetID(), u.GetName())
	return nil
}

// BeforeUpdate 更新前钩子
func (u *User) BeforeUpdate() error {
	fmt.Printf("正在更新用户: ID=%v\n", u.GetID())
	return nil
}

// AfterUpdate 更新后钩子
func (u *User) AfterUpdate() error {
	fmt.Printf("用户更新成功: ID=%v, 姓名=%s\n", u.GetID(), u.GetName())
	return nil
}

// BeforeDelete 删除前钩子
func (u *User) BeforeDelete() error {
	fmt.Printf("正在删除用户: ID=%v, 姓名=%s\n", u.GetID(), u.GetName())
	return nil
}

// AfterDelete 删除后钩子
func (u *User) AfterDelete() error {
	fmt.Printf("用户删除成功: ID=%v\n", u.GetID())
	return nil
}

// 静态查询方法

// FindByEmail 根据邮箱查找用户
func FindByEmail(ctx context.Context, email string) (*User, error) {
	query, err := db.Table("users")
	if err != nil {
		return nil, err
	}

	data, err := query.Where("email", "=", email).First(ctx)
	if err != nil {
		return nil, err
	}

	return NewUserWithData(data), nil
}

// FindActiveUsers 查找活跃用户
func FindActiveUsers(ctx context.Context, limit int) ([]*User, error) {
	query, err := db.Table("users")
	if err != nil {
		return nil, err
	}

	results, err := query.
		Where("status", "=", "active").
		OrderBy("created_at", "desc").
		Limit(limit).
		Get(ctx)

	if err != nil {
		return nil, err
	}

	var users []*User
	for _, data := range results {
		users = append(users, NewUserWithData(data))
	}

	return users, nil
}

// FindAdultUsers 查找成年用户
func FindAdultUsers(ctx context.Context) ([]*User, error) {
	query, err := db.Table("users")
	if err != nil {
		return nil, err
	}

	results, err := query.
		Where("age", ">=", 18).
		Where("status", "=", "active").
		OrderBy("age", "asc").
		Get(ctx)

	if err != nil {
		return nil, err
	}

	var users []*User
	for _, data := range results {
		users = append(users, NewUserWithData(data))
	}

	return users, nil
}

// CountByStatus 按状态统计用户数量
func CountByStatus(ctx context.Context, status string) (int64, error) {
	query, err := db.Table("users")
	if err != nil {
		return 0, err
	}

	return query.Where("status", "=", status).Count(ctx)
}

// 关联关系

// Profile 获取用户资料 (HasOne)
func (u *User) Profile() *HasOne {
	return u.HasOne(&Profile{}, "user_id", "id")
}

// Posts 获取用户文章 (HasMany)
func (u *User) Posts() *HasMany {
	return u.HasMany(&Post{}, "user_id", "id")
}
