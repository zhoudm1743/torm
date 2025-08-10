package model

import (
	"context"
	"fmt"
	"time"

	"torm/pkg/db"
)

// Post 文章模型
type Post struct {
	*BaseModel
}

// NewPost 创建文章模型实例
func NewPost() *Post {
	post := &Post{
		BaseModel: NewBaseModel(),
	}
	post.SetTable("posts")
	return post
}

// NewPostWithData 使用数据创建文章模型
func NewPostWithData(data map[string]interface{}) *Post {
	post := NewPost()
	post.Fill(data)
	if post.GetAttribute("id") != nil {
		post.isNew = false
		post.exists = true
		post.syncOriginal()
	}
	return post
}

// 属性访问器

// GetID 获取文章ID
func (p *Post) GetID() interface{} {
	return p.GetAttribute("id")
}

// SetID 设置文章ID
func (p *Post) SetID(id interface{}) *Post {
	p.SetAttribute("id", id)
	return p
}

// GetUserID 获取作者ID
func (p *Post) GetUserID() interface{} {
	return p.GetAttribute("user_id")
}

// SetUserID 设置作者ID
func (p *Post) SetUserID(userID interface{}) *Post {
	p.SetAttribute("user_id", userID)
	return p
}

// GetTitle 获取文章标题
func (p *Post) GetTitle() string {
	if title := p.GetAttribute("title"); title != nil {
		return title.(string)
	}
	return ""
}

// SetTitle 设置文章标题
func (p *Post) SetTitle(title string) *Post {
	p.SetAttribute("title", title)
	return p
}

// GetContent 获取文章内容
func (p *Post) GetContent() string {
	if content := p.GetAttribute("content"); content != nil {
		return content.(string)
	}
	return ""
}

// SetContent 设置文章内容
func (p *Post) SetContent(content string) *Post {
	p.SetAttribute("content", content)
	return p
}

// GetStatus 获取文章状态
func (p *Post) GetStatus() string {
	if status := p.GetAttribute("status"); status != nil {
		return status.(string)
	}
	return ""
}

// SetStatus 设置文章状态
func (p *Post) SetStatus(status string) *Post {
	p.SetAttribute("status", status)
	return p
}

// GetPublishedAt 获取发布时间
func (p *Post) GetPublishedAt() time.Time {
	if publishedAt := p.GetAttribute("published_at"); publishedAt != nil {
		switch v := publishedAt.(type) {
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

// SetPublishedAt 设置发布时间
func (p *Post) SetPublishedAt(publishedAt time.Time) *Post {
	p.SetAttribute("published_at", publishedAt)
	return p
}

// 业务方法

// IsPublished 检查是否已发布
func (p *Post) IsPublished() bool {
	return p.GetStatus() == "published"
}

// IsDraft 检查是否为草稿
func (p *Post) IsDraft() bool {
	return p.GetStatus() == "draft"
}

// Publish 发布文章
func (p *Post) Publish() *Post {
	p.SetStatus("published")
	if p.GetPublishedAt().IsZero() {
		p.SetPublishedAt(time.Now())
	}
	return p
}

// UnPublish 取消发布
func (p *Post) UnPublish() *Post {
	p.SetStatus("draft")
	return p
}

// 关联关系

// Author 获取文章作者 (BelongsTo)
func (p *Post) Author() *BelongsTo {
	return p.BelongsTo(&User{}, "user_id", "id")
}

// Tags 获取文章标签 (ManyToMany)
func (p *Post) Tags() *ManyToMany {
	return p.BelongsToMany(&Tag{}, "post_tags", "post_id", "tag_id", "id", "id")
}

// 验证方法

// Validate 验证文章数据
func (p *Post) Validate() error {
	if p.GetTitle() == "" {
		return fmt.Errorf("文章标题不能为空")
	}

	if p.GetContent() == "" {
		return fmt.Errorf("文章内容不能为空")
	}

	if p.GetUserID() == nil {
		return fmt.Errorf("文章作者不能为空")
	}

	// 验证状态
	status := p.GetStatus()
	if status != "" && status != "draft" && status != "published" && status != "archived" {
		return fmt.Errorf("文章状态不正确")
	}

	return nil
}

// BeforeSave 保存前验证
func (p *Post) BeforeSave() error {
	return p.Validate()
}

// 事件钩子

// BeforeCreate 创建前钩子
func (p *Post) BeforeCreate() error {
	// 设置默认状态
	if p.GetStatus() == "" {
		p.SetStatus("draft")
	}

	fmt.Printf("正在创建文章: 标题=%s, 作者ID=%v\n", p.GetTitle(), p.GetUserID())
	return nil
}

// AfterCreate 创建后钩子
func (p *Post) AfterCreate() error {
	fmt.Printf("文章创建成功: ID=%v, 标题=%s\n", p.GetID(), p.GetTitle())
	return nil
}

// 静态查询方法

// FindByUserID 根据用户ID查找文章
func FindPostsByUserID(ctx context.Context, userID interface{}) ([]*Post, error) {
	query, err := db.Table("posts")
	if err != nil {
		return nil, err
	}

	results, err := query.Where("user_id", "=", userID).OrderBy("created_at", "desc").Get(ctx)
	if err != nil {
		return nil, err
	}

	var posts []*Post
	for _, data := range results {
		posts = append(posts, NewPostWithData(data))
	}

	return posts, nil
}

// FindPublishedPosts 查找已发布的文章
func FindPublishedPosts(ctx context.Context, limit int) ([]*Post, error) {
	query, err := db.Table("posts")
	if err != nil {
		return nil, err
	}

	results, err := query.
		Where("status", "=", "published").
		OrderBy("published_at", "desc").
		Limit(limit).
		Get(ctx)

	if err != nil {
		return nil, err
	}

	var posts []*Post
	for _, data := range results {
		posts = append(posts, NewPostWithData(data))
	}

	return posts, nil
}
