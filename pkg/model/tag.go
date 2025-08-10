package model

import (
	"context"
	"fmt"
	"strings"

	"torm/pkg/db"
)

// Tag 标签模型
type Tag struct {
	*BaseModel
}

// NewTag 创建标签模型实例
func NewTag() *Tag {
	tag := &Tag{
		BaseModel: NewBaseModel(),
	}
	tag.SetTable("tags")
	return tag
}

// NewTagWithData 使用数据创建标签模型
func NewTagWithData(data map[string]interface{}) *Tag {
	tag := NewTag()
	tag.Fill(data)
	if tag.GetAttribute("id") != nil {
		tag.isNew = false
		tag.exists = true
		tag.syncOriginal()
	}
	return tag
}

// 属性访问器

// GetID 获取标签ID
func (t *Tag) GetID() interface{} {
	return t.GetAttribute("id")
}

// SetID 设置标签ID
func (t *Tag) SetID(id interface{}) *Tag {
	t.SetAttribute("id", id)
	return t
}

// GetName 获取标签名称
func (t *Tag) GetName() string {
	if name := t.GetAttribute("name"); name != nil {
		return name.(string)
	}
	return ""
}

// SetName 设置标签名称
func (t *Tag) SetName(name string) *Tag {
	t.SetAttribute("name", name)
	return t
}

// GetSlug 获取标签别名
func (t *Tag) GetSlug() string {
	if slug := t.GetAttribute("slug"); slug != nil {
		return slug.(string)
	}
	return ""
}

// SetSlug 设置标签别名
func (t *Tag) SetSlug(slug string) *Tag {
	t.SetAttribute("slug", slug)
	return t
}

// GetDescription 获取标签描述
func (t *Tag) GetDescription() string {
	if description := t.GetAttribute("description"); description != nil {
		return description.(string)
	}
	return ""
}

// SetDescription 设置标签描述
func (t *Tag) SetDescription(description string) *Tag {
	t.SetAttribute("description", description)
	return t
}

// 关联关系

// Posts 获取标签下的文章 (ManyToMany)
func (t *Tag) Posts() *ManyToMany {
	return t.BelongsToMany(&Post{}, "post_tags", "tag_id", "post_id", "id", "id")
}

// 验证方法

// Validate 验证标签数据
func (t *Tag) Validate() error {
	if t.GetName() == "" {
		return fmt.Errorf("标签名称不能为空")
	}

	if len(t.GetName()) > 50 {
		return fmt.Errorf("标签名称长度不能超过50个字符")
	}

	return nil
}

// BeforeSave 保存前验证
func (t *Tag) BeforeSave() error {
	return t.Validate()
}

// 事件钩子

// BeforeCreate 创建前钩子
func (t *Tag) BeforeCreate() error {
	// 自动生成slug
	if t.GetSlug() == "" {
		slug := strings.ToLower(strings.ReplaceAll(t.GetName(), " ", "-"))
		t.SetSlug(slug)
	}

	fmt.Printf("正在创建标签: 名称=%s, 别名=%s\n", t.GetName(), t.GetSlug())
	return nil
}

// AfterCreate 创建后钩子
func (t *Tag) AfterCreate() error {
	fmt.Printf("标签创建成功: ID=%v, 名称=%s\n", t.GetID(), t.GetName())
	return nil
}

// 静态查询方法

// FindByName 根据名称查找标签
func FindTagByName(ctx context.Context, name string) (*Tag, error) {
	query, err := db.Table("tags")
	if err != nil {
		return nil, err
	}

	data, err := query.Where("name", "=", name).First(ctx)
	if err != nil {
		return nil, err
	}

	return NewTagWithData(data), nil
}

// FindBySlug 根据别名查找标签
func FindTagBySlug(ctx context.Context, slug string) (*Tag, error) {
	query, err := db.Table("tags")
	if err != nil {
		return nil, err
	}

	data, err := query.Where("slug", "=", slug).First(ctx)
	if err != nil {
		return nil, err
	}

	return NewTagWithData(data), nil
}

// FindPopularTags 查找热门标签
func FindPopularTags(ctx context.Context, limit int) ([]*Tag, error) {
	query, err := db.Table("tags")
	if err != nil {
		return nil, err
	}

	// 通过关联的文章数量排序
	results, err := query.
		Select("tags.*", "COUNT(post_tags.post_id) as post_count").
		LeftJoin("post_tags", "tags.id", "=", "post_tags.tag_id").
		GroupBy("tags.id").
		OrderBy("post_count", "desc").
		Limit(limit).
		Get(ctx)

	if err != nil {
		return nil, err
	}

	var tags []*Tag
	for _, data := range results {
		tags = append(tags, NewTagWithData(data))
	}

	return tags, nil
}
