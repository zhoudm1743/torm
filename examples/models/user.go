package models

import (
	"fmt"
	"strconv"
	"time"

	"github.com/zhoudm1743/torm/pkg/model"
)

type User struct {
	model.BaseModel             // 改为值嵌入
	ID              interface{} `json:"id" db:"id"`
	Name            string      `json:"name" db:"name"`
	Email           string      `json:"email" db:"email"`
	Age             int         `json:"age" db:"age"`
	Status          string      `json:"status" db:"status"`
	CreatedAt       time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at" db:"updated_at"`
}

// UserWithUUID 使用UUID作为主键的用户模型
type UserWithUUID struct {
	model.BaseModel
	UUID      string    `json:"uuid" db:"uuid" primary:"true"` // 自定义UUID主键
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserWithCompositePK 使用复合主键的用户模型
type UserWithCompositePK struct {
	model.BaseModel
	TenantID  string    `json:"tenant_id" db:"tenant_id" primary:"true"` // 复合主键1
	UserID    string    `json:"user_id" db:"user_id" primary:"true"`     // 复合主键2
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// NewUser 创建用户模型
func NewUser() *User {
	user := &User{
		BaseModel: *model.NewBaseModel(),
	}
	user.SetTable("users")
	user.SetConnection("default")
	return user
}

// NewUserWithUUID 创建UUID主键的用户模型
func NewUserWithUUID() *UserWithUUID {
	user := &UserWithUUID{
		BaseModel: *model.NewBaseModel(),
	}
	user.SetTable("users_uuid")
	user.SetConnection("default")
	// 自动检测主键标签
	user.DetectPrimaryKeysFromStruct(user)
	return user
}

// NewUserWithCompositePK 创建复合主键的用户模型
func NewUserWithCompositePK() *UserWithCompositePK {
	user := &UserWithCompositePK{
		BaseModel: *model.NewBaseModel(),
	}
	user.SetTable("users_composite")
	user.SetConnection("default")
	// 自动检测主键标签
	user.DetectPrimaryKeysFromStruct(user)
	return user
}

// ===== 方法转发：确保BaseModel方法在User上可用，并返回*User便于链式调用 =====

// Where 转发到BaseModel
func (u *User) Where(field string, operator string, value interface{}) *User {
	u.BaseModel.Where(field, operator, value)
	return u
}

// WhereIn 转发到BaseModel
func (u *User) WhereIn(field string, values []interface{}) *User {
	u.BaseModel.WhereIn(field, values)
	return u
}

// WhereNotIn 转发到BaseModel
func (u *User) WhereNotIn(field string, values []interface{}) *User {
	u.BaseModel.WhereNotIn(field, values)
	return u
}

// WhereBetween 转发到BaseModel
func (u *User) WhereBetween(field string, start, end interface{}) *User {
	u.BaseModel.WhereBetween(field, start, end)
	return u
}

// WhereNull 转发到BaseModel
func (u *User) WhereNull(field string) *User {
	u.BaseModel.WhereNull(field)
	return u
}

// WhereNotNull 转发到BaseModel
func (u *User) WhereNotNull(field string) *User {
	u.BaseModel.WhereNotNull(field)
	return u
}

// OrderBy 转发到BaseModel
func (u *User) OrderBy(field string, direction string) *User {
	u.BaseModel.OrderBy(field, direction)
	return u
}

// Limit 转发到BaseModel
func (u *User) Limit(limit int) *User {
	u.BaseModel.Limit(limit)
	return u
}

// Offset 转发到BaseModel
func (u *User) Offset(offset int) *User {
	u.BaseModel.Offset(offset)
	return u
}

// Select 转发到BaseModel
func (u *User) Select(fields ...string) *User {
	u.BaseModel.Select(fields...)
	return u
}

// GroupBy 转发到BaseModel
func (u *User) GroupBy(fields ...string) *User {
	u.BaseModel.GroupBy(fields...)
	return u
}

// Having 转发到BaseModel
func (u *User) Having(field string, operator string, value interface{}) *User {
	u.BaseModel.Having(field, operator, value)
	return u
}

// 执行方法转发
func (u *User) Get() ([]map[string]interface{}, error) {
	return u.BaseModel.Get()
}

func (u *User) All() ([]map[string]interface{}, error) {
	return u.BaseModel.All()
}

func (u *User) First(dest ...interface{}) (map[string]interface{}, error) {
	result, err := u.BaseModel.First(dest...)
	if err != nil {
		return result, err
	}

	// 填充User特有字段
	u.ID = u.GetAttribute("id")
	u.Name = u.getStringAttribute("name")
	u.Email = u.getStringAttribute("email")
	u.Age = u.getIntAttribute("age")
	u.Status = u.getStringAttribute("status")

	if createdAt := u.GetAttribute("created_at"); createdAt != nil {
		if t, ok := createdAt.(time.Time); ok {
			u.CreatedAt = t
		}
	}

	if updatedAt := u.GetAttribute("updated_at"); updatedAt != nil {
		if t, ok := updatedAt.(time.Time); ok {
			u.UpdatedAt = t
		}
	}

	return result, nil
}

// TakeFirst 链式查询后获取第一条记录并填充User模型
func (u *User) TakeFirst(dest ...interface{}) (map[string]interface{}, error) {
	return u.First(dest...)
}

// FirstOrCreate 查找第一条记录，如果不存在则创建
func (u *User) FirstOrCreate(attributes map[string]interface{}) error {
	err := u.BaseModel.FirstOrCreate(attributes)
	if err != nil {
		return err
	}

	// 填充User特有字段
	u.ID = u.GetAttribute("id")
	u.Name = u.getStringAttribute("name")
	u.Email = u.getStringAttribute("email")
	u.Age = u.getIntAttribute("age")
	u.Status = u.getStringAttribute("status")

	if createdAt := u.GetAttribute("created_at"); createdAt != nil {
		if t, ok := createdAt.(time.Time); ok {
			u.CreatedAt = t
		}
	}

	if updatedAt := u.GetAttribute("updated_at"); updatedAt != nil {
		if t, ok := updatedAt.(time.Time); ok {
			u.UpdatedAt = t
		}
	}

	return nil
}

// FirstOrNew 查找第一条记录，如果不存在则创建新模型实例（不保存）
func (u *User) FirstOrNew(attributes map[string]interface{}) error {
	err := u.BaseModel.FirstOrNew(attributes)
	if err != nil {
		return err
	}

	// 填充User特有字段
	u.ID = u.GetAttribute("id")
	u.Name = u.getStringAttribute("name")
	u.Email = u.getStringAttribute("email")
	u.Age = u.getIntAttribute("age")
	u.Status = u.getStringAttribute("status")

	if createdAt := u.GetAttribute("created_at"); createdAt != nil {
		if t, ok := createdAt.(time.Time); ok {
			u.CreatedAt = t
		}
	}

	if updatedAt := u.GetAttribute("updated_at"); updatedAt != nil {
		if t, ok := updatedAt.(time.Time); ok {
			u.UpdatedAt = t
		}
	}

	return nil
}

// Count 转发到BaseModel
func (u *User) Count() (int64, error) {
	return u.BaseModel.Count()
}

// Paginate 转发到BaseModel
func (u *User) Paginate(page, perPage int) (interface{}, error) {
	return u.BaseModel.Paginate(page, perPage)
}

// ToSQL 转发到BaseModel
func (u *User) ToSQL() (string, []interface{}, error) {
	return u.BaseModel.ToSQL()
}

// GetKey 转发到BaseModel
func (u *User) GetKey() interface{} {
	return u.BaseModel.GetKey()
}

// UpdateOrCreate 转发到BaseModel
func (u *User) UpdateOrCreate(conditions, values map[string]interface{}) error {
	return u.BaseModel.UpdateOrCreate(conditions, values)
}

// Chunk 转发到BaseModel
func (u *User) Chunk(size int, callback func([]map[string]interface{}) error) error {
	return u.BaseModel.Chunk(size, callback)
}

// Find 转发到BaseModel并填充User字段
func (u *User) Find(id interface{}, dest ...interface{}) (map[string]interface{}, error) {
	result, err := u.BaseModel.Find(id, dest...)
	if err != nil {
		return result, err
	}

	// 填充User特有字段
	u.ID = u.GetAttribute("id")
	u.Name = u.getStringAttribute("name")
	u.Email = u.getStringAttribute("email")
	u.Age = u.getIntAttribute("age")
	u.Status = u.getStringAttribute("status")

	if createdAt := u.GetAttribute("created_at"); createdAt != nil {
		if t, ok := createdAt.(time.Time); ok {
			u.CreatedAt = t
		}
	}

	if updatedAt := u.GetAttribute("updated_at"); updatedAt != nil {
		if t, ok := updatedAt.(time.Time); ok {
			u.UpdatedAt = t
		}
	}

	return result, nil
}

// 辅助方法
func (u *User) getStringAttribute(key string) string {
	if val := u.GetAttribute(key); val != nil {
		switch v := val.(type) {
		case string:
			return v
		case []byte:
			return string(v)
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

func (u *User) getIntAttribute(key string) int {
	if val := u.GetAttribute(key); val != nil {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case int32:
			return int(v)
		case float64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		case []byte:
			if i, err := strconv.Atoi(string(v)); err == nil {
				return i
			}
		}
	}
	return 0
}
