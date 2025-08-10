package tests

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/zhoudm1743/torm/pkg/db"
	"github.com/zhoudm1743/torm/pkg/model"
)

// User 用户模型
type User struct {
	*model.BaseModel
	ID        interface{} `json:"id" db:"id"`
	Name      string      `json:"name" db:"name"`
	Email     string      `json:"email" db:"email"`
	Age       int         `json:"age" db:"age"`
	Status    string      `json:"status" db:"status"`
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt time.Time   `json:"updated_at" db:"updated_at"`
}

// NewUser 创建新用户实例
func NewUser() *User {
	user := &User{
		BaseModel: model.NewBaseModel(),
		Status:    "active",
	}
	user.SetTable("users")
	user.SetPrimaryKey("id")
	user.SetConnection("default")
	return user
}

// TableName 返回表名
func (u *User) TableName() string {
	return "users"
}

// Validate 验证用户数据
func (u *User) Validate() error {
	if strings.TrimSpace(u.Name) == "" {
		return fmt.Errorf("姓名不能为空")
	}

	// 验证邮箱格式
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(u.Email) {
		return fmt.Errorf("邮箱格式不正确")
	}

	if u.Age < 0 {
		return fmt.Errorf("年龄必须大于0")
	}

	return nil
}

// Save 保存用户
func (u *User) Save() error {
	if err := u.Validate(); err != nil {
		return err
	}

	// 同步字段到attributes
	u.SetAttribute("name", u.Name)
	u.SetAttribute("email", u.Email)
	u.SetAttribute("age", u.Age)
	u.SetAttribute("status", u.Status)

	err := u.BaseModel.Save()
	if err != nil {
		return err
	}

	// 保存后同步回结构体字段
	u.syncFromAttributes()

	return nil
}

// syncFromAttributes 从attributes同步数据到结构体字段
func (u *User) syncFromAttributes() {
	if id := u.GetAttribute("id"); id != nil {
		u.ID = id
	}
	if name := u.GetAttribute("name"); name != nil {
		u.Name = name.(string)
	}
	if email := u.GetAttribute("email"); email != nil {
		u.Email = email.(string)
	}
	if age := u.GetAttribute("age"); age != nil {
		if ageInt, ok := age.(int64); ok {
			u.Age = int(ageInt)
		} else if ageInt, ok := age.(int); ok {
			u.Age = ageInt
		}
	}
	if status := u.GetAttribute("status"); status != nil {
		u.Status = status.(string)
	}
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
}

// Fill 填充用户数据
func (u *User) Fill(data map[string]interface{}) *User {
	u.BaseModel.Fill(data)
	u.syncFromAttributes()
	return u
}

// Find 根据主键查找用户
func (u *User) Find(id interface{}) error {
	_, err := u.BaseModel.Find(id)
	if err != nil {
		return err
	}
	u.syncFromAttributes()
	return nil
}

// Profile 获取用户资料 (HasOne关系)
func (u *User) Profile() *model.HasOne {
	return model.NewHasOne(u, reflect.TypeOf(&Profile{}), "user_id", "id")
}

// Posts 获取用户文章 (HasMany关系)
func (u *User) Posts() *model.HasMany {
	return model.NewHasMany(u, reflect.TypeOf(&Post{}), "user_id", "id")
}

// FindByEmail 根据邮箱查找用户
func FindByEmail(email string) (*User, error) {
	query, err := db.Table("users")
	if err != nil {
		return nil, err
	}

	data, err := query.Where("email", "=", email).First()
	if err != nil {
		return nil, err
	}

	user := NewUser()
	user.Fill(data)
	return user, nil
}

// FindActiveUsers 查找活跃用户
func FindActiveUsers(limit int) ([]*User, error) {
	query, err := db.Table("users")
	if err != nil {
		return nil, err
	}

	results, err := query.Where("status", "=", "active").Limit(limit).Get()
	if err != nil {
		return nil, err
	}

	var users []*User
	for _, data := range results {
		user := NewUser()
		user.Fill(data)
		users = append(users, user)
	}

	return users, nil
}

// CountByStatus 按状态统计用户数量
func CountByStatus(status string) (int64, error) {
	query, err := db.Table("users")
	if err != nil {
		return 0, err
	}

	return query.Where("status", "=", status).Count()
}

// Profile 用户资料模型
type Profile struct {
	*model.BaseModel
	ID        interface{} `json:"id" db:"id"`
	UserID    interface{} `json:"user_id" db:"user_id"`
	Avatar    string      `json:"avatar" db:"avatar"`
	Bio       string      `json:"bio" db:"bio"`
	Website   string      `json:"website" db:"website"`
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt time.Time   `json:"updated_at" db:"updated_at"`
}

// NewProfile 创建新资料实例
func NewProfile() *Profile {
	profile := &Profile{
		BaseModel: model.NewBaseModel(),
	}
	profile.SetTable("profiles")
	profile.SetPrimaryKey("id")
	profile.SetConnection("default")
	return profile
}

// Save 保存资料
func (p *Profile) Save() error {
	// 同步字段到attributes
	p.SetAttribute("user_id", p.UserID)
	p.SetAttribute("avatar", p.Avatar)
	p.SetAttribute("bio", p.Bio)
	p.SetAttribute("website", p.Website)

	err := p.BaseModel.Save()
	if err != nil {
		return err
	}

	// 保存后同步回结构体字段
	p.syncFromAttributes()
	return nil
}

// syncFromAttributes 从attributes同步数据到结构体字段
func (p *Profile) syncFromAttributes() {
	if id := p.GetAttribute("id"); id != nil {
		p.ID = id
	}
	if userID := p.GetAttribute("user_id"); userID != nil {
		p.UserID = userID
	}
	if avatar := p.GetAttribute("avatar"); avatar != nil {
		p.Avatar = avatar.(string)
	}
	if bio := p.GetAttribute("bio"); bio != nil {
		p.Bio = bio.(string)
	}
	if website := p.GetAttribute("website"); website != nil {
		p.Website = website.(string)
	}
	if createdAt := p.GetAttribute("created_at"); createdAt != nil {
		if t, ok := createdAt.(time.Time); ok {
			p.CreatedAt = t
		}
	}
	if updatedAt := p.GetAttribute("updated_at"); updatedAt != nil {
		if t, ok := updatedAt.(time.Time); ok {
			p.UpdatedAt = t
		}
	}
}

// Fill 填充资料数据
func (p *Profile) Fill(data map[string]interface{}) *Profile {
	p.BaseModel.Fill(data)
	p.syncFromAttributes()
	return p
}

// Find 根据主键查找资料
func (p *Profile) Find(id interface{}) error {
	_, err := p.BaseModel.Find(id)
	if err != nil {
		return err
	}
	p.syncFromAttributes()
	return nil
}

// User 获取资料对应的用户 (BelongsTo关系)
func (p *Profile) User() *model.BelongsTo {
	return model.NewBelongsTo(p, reflect.TypeOf(&User{}), "user_id", "id")
}

// FindProfileByUserID 根据用户ID查找资料
func FindProfileByUserID(userID interface{}) (*Profile, error) {
	query, err := db.Table("profiles")
	if err != nil {
		return nil, err
	}

	data, err := query.Where("user_id", "=", userID).First()
	if err != nil {
		return nil, err
	}

	profile := NewProfile()
	profile.Fill(data)
	return profile, nil
}

// Post 文章模型
type Post struct {
	*model.BaseModel
	ID          interface{} `json:"id" db:"id"`
	UserID      interface{} `json:"user_id" db:"user_id"`
	Title       string      `json:"title" db:"title"`
	Content     string      `json:"content" db:"content"`
	Status      string      `json:"status" db:"status"`
	PublishedAt *time.Time  `json:"published_at" db:"published_at"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

// NewPost 创建新文章实例
func NewPost() *Post {
	post := &Post{
		BaseModel: model.NewBaseModel(),
		Status:    "draft",
	}
	post.SetTable("posts")
	post.SetPrimaryKey("id")
	post.SetConnection("default")
	return post
}

// Save 保存文章
func (p *Post) Save() error {
	p.SetAttribute("user_id", p.UserID)
	p.SetAttribute("title", p.Title)
	p.SetAttribute("content", p.Content)
	p.SetAttribute("status", p.Status)
	if p.PublishedAt != nil {
		p.SetAttribute("published_at", *p.PublishedAt)
	}

	err := p.BaseModel.Save()
	if err != nil {
		return err
	}

	if id := p.GetAttribute("id"); id != nil {
		p.ID = id
	}
	return nil
}

// Fill 填充文章数据
func (p *Post) Fill(data map[string]interface{}) *Post {
	p.BaseModel.Fill(data)
	if id, ok := data["id"]; ok {
		p.ID = id
	}
	if userID, ok := data["user_id"]; ok {
		p.UserID = userID
	}
	if title, ok := data["title"]; ok {
		p.Title = title.(string)
	}
	if content, ok := data["content"]; ok {
		p.Content = content.(string)
	}
	if status, ok := data["status"]; ok {
		p.Status = status.(string)
	}
	return p
}

// Author 获取文章作者 (BelongsTo关系)
func (p *Post) Author() *model.BelongsTo {
	return model.NewBelongsTo(p, reflect.TypeOf(&User{}), "user_id", "id")
}

// Tags 获取文章标签 (ManyToMany关系)
func (p *Post) Tags() *model.ManyToMany {
	return model.NewManyToMany(p, reflect.TypeOf(&Tag{}), "post_tags", "post_id", "tag_id", "id", "id")
}

// FindPostsByUserID 根据用户ID查找文章
func FindPostsByUserID(userID interface{}) ([]*Post, error) {
	query, err := db.Table("posts")
	if err != nil {
		return nil, err
	}

	results, err := query.
		Where("user_id", "=", userID).
		Where("status", "=", "published").
		OrderBy("created_at", "desc").
		Get()
	if err != nil {
		return nil, err
	}

	var posts []*Post
	for _, data := range results {
		post := NewPost()
		post.Fill(data)
		posts = append(posts, post)
	}

	return posts, nil
}

// FindPublishedPosts 查找已发布的文章
func FindPublishedPosts(limit int) ([]*Post, error) {
	query, err := db.Table("posts")
	if err != nil {
		return nil, err
	}

	results, err := query.
		Where("status", "=", "published").
		OrderBy("published_at", "desc").
		Limit(limit).
		Get()
	if err != nil {
		return nil, err
	}

	var posts []*Post
	for _, data := range results {
		post := NewPost()
		post.Fill(data)
		posts = append(posts, post)
	}

	return posts, nil
}

// Tag 标签模型
type Tag struct {
	*model.BaseModel
	ID          interface{} `json:"id" db:"id"`
	Name        string      `json:"name" db:"name"`
	Slug        string      `json:"slug" db:"slug"`
	Description string      `json:"description" db:"description"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

// NewTag 创建新标签实例
func NewTag() *Tag {
	tag := &Tag{
		BaseModel: model.NewBaseModel(),
	}
	tag.SetTable("tags")
	tag.SetPrimaryKey("id")
	tag.SetConnection("default")
	return tag
}

// Save 保存标签
func (t *Tag) Save() error {
	t.SetAttribute("name", t.Name)
	t.SetAttribute("slug", t.Slug)
	t.SetAttribute("description", t.Description)

	err := t.BaseModel.Save()
	if err != nil {
		return err
	}

	if id := t.GetAttribute("id"); id != nil {
		t.ID = id
	}
	return nil
}

// Fill 填充标签数据
func (t *Tag) Fill(data map[string]interface{}) *Tag {
	t.BaseModel.Fill(data)
	if id, ok := data["id"]; ok {
		t.ID = id
	}
	if name, ok := data["name"]; ok {
		t.Name = name.(string)
	}
	if slug, ok := data["slug"]; ok {
		t.Slug = slug.(string)
	}
	if description, ok := data["description"]; ok {
		t.Description = description.(string)
	}
	return t
}

// Posts 获取标签下的文章 (ManyToMany关系)
func (t *Tag) Posts() *model.ManyToMany {
	return model.NewManyToMany(t, reflect.TypeOf(&Post{}), "post_tags", "tag_id", "post_id", "id", "id")
}

// FindTagByName 根据名称查找标签
func FindTagByName(name string) (*Tag, error) {
	query, err := db.Table("tags")
	if err != nil {
		return nil, err
	}

	data, err := query.Where("name", "=", name).First()
	if err != nil {
		return nil, err
	}

	tag := NewTag()
	tag.Fill(data)
	return tag, nil
}

// FindTagBySlug 根据slug查找标签
func FindTagBySlug(slug string) (*Tag, error) {
	query, err := db.Table("tags")
	if err != nil {
		return nil, err
	}

	data, err := query.Where("slug", "=", slug).First()
	if err != nil {
		return nil, err
	}

	tag := NewTag()
	tag.Fill(data)
	return tag, nil
}

// FindPopularTags 查找热门标签
func FindPopularTags(limit int) ([]*Tag, error) {
	query, err := db.Table("tags")
	if err != nil {
		return nil, err
	}

	// 简化版本，实际应该按文章数量排序
	results, err := query.OrderBy("created_at", "desc").Limit(limit).Get()
	if err != nil {
		return nil, err
	}

	var tags []*Tag
	for _, data := range results {
		tag := NewTag()
		tag.Fill(data)
		tags = append(tags, tag)
	}

	return tags, nil
}
