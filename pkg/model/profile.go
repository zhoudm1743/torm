package model

import (
	"context"
	"fmt"
	"torm/pkg/db"
)

// Profile 用户资料模型
type Profile struct {
	*BaseModel
}

// NewProfile 创建用户资料模型实例
func NewProfile() *Profile {
	profile := &Profile{
		BaseModel: NewBaseModel(),
	}
	profile.SetTable("profiles")
	return profile
}

// NewProfileWithData 使用数据创建用户资料模型
func NewProfileWithData(data map[string]interface{}) *Profile {
	profile := NewProfile()
	profile.Fill(data)
	if profile.GetAttribute("id") != nil {
		profile.isNew = false
		profile.exists = true
		profile.syncOriginal()
	}
	return profile
}

// 属性访问器

// GetID 获取资料ID
func (p *Profile) GetID() interface{} {
	return p.GetAttribute("id")
}

// SetID 设置资料ID
func (p *Profile) SetID(id interface{}) *Profile {
	p.SetAttribute("id", id)
	return p
}

// GetUserID 获取用户ID
func (p *Profile) GetUserID() interface{} {
	return p.GetAttribute("user_id")
}

// SetUserID 设置用户ID
func (p *Profile) SetUserID(userID interface{}) *Profile {
	p.SetAttribute("user_id", userID)
	return p
}

// GetAvatar 获取头像
func (p *Profile) GetAvatar() string {
	if avatar := p.GetAttribute("avatar"); avatar != nil {
		return avatar.(string)
	}
	return ""
}

// SetAvatar 设置头像
func (p *Profile) SetAvatar(avatar string) *Profile {
	p.SetAttribute("avatar", avatar)
	return p
}

// GetBio 获取个人简介
func (p *Profile) GetBio() string {
	if bio := p.GetAttribute("bio"); bio != nil {
		return bio.(string)
	}
	return ""
}

// SetBio 设置个人简介
func (p *Profile) SetBio(bio string) *Profile {
	p.SetAttribute("bio", bio)
	return p
}

// GetWebsite 获取网站
func (p *Profile) GetWebsite() string {
	if website := p.GetAttribute("website"); website != nil {
		return website.(string)
	}
	return ""
}

// SetWebsite 设置网站
func (p *Profile) SetWebsite(website string) *Profile {
	p.SetAttribute("website", website)
	return p
}

// 关联关系

// User 获取所属用户 (BelongsTo)
func (p *Profile) User() *BelongsTo {
	return p.BelongsTo(&User{}, "user_id", "id")
}

// 事件钩子

// BeforeCreate 创建前钩子
func (p *Profile) BeforeCreate() error {
	fmt.Printf("正在创建用户资料: 用户ID=%v\n", p.GetUserID())
	return nil
}

// AfterCreate 创建后钩子
func (p *Profile) AfterCreate() error {
	fmt.Printf("用户资料创建成功: ID=%v\n", p.GetID())
	return nil
}

// 验证方法

// Validate 验证资料数据
func (p *Profile) Validate() error {
	if p.GetUserID() == nil {
		return fmt.Errorf("用户ID不能为空")
	}

	return nil
}

// BeforeSave 保存前验证
func (p *Profile) BeforeSave() error {
	return p.Validate()
}

// 静态查询方法

// FindByUserID 根据用户ID查找资料
func FindProfileByUserID(ctx context.Context, userID interface{}) (*Profile, error) {
	query, err := db.Table("profiles")
	if err != nil {
		return nil, err
	}

	data, err := query.Where("user_id", "=", userID).First(ctx)
	if err != nil {
		return nil, err
	}

	return NewProfileWithData(data), nil
}
