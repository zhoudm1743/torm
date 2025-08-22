# 关联关系

TORM 提供了完整的关联关系支持，包括 HasOne、HasMany、BelongsTo、ManyToMany 等关系类型，并支持关联预加载来解决 N+1 查询问题。

## 📋 目录

- [关系类型](#关系类型)
- [定义关联](#定义关联)
- [查询关联](#查询关联)
- [预加载](#预加载)
- [关联操作](#关联操作)
- [多态关联](#多态关联)
- [性能优化](#性能优化)

## 🚀 快速开始

### 基础关联定义

```go
// User 用户模型
type User struct {
    model.BaseModel
    ID        int64     `json:"id" db:"id"`
    Name      string    `json:"name" db:"name"`
    Email     string    `json:"email" db:"email"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Post 文章模型
type Post struct {
    model.BaseModel
    ID        int64     `json:"id" db:"id"`
    UserID    int64     `json:"user_id" db:"user_id"`
    Title     string    `json:"title" db:"title"`
    Content   string    `json:"content" db:"content"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Profile 用户档案模型
type Profile struct {
    model.BaseModel
    ID       int64  `json:"id" db:"id"`
    UserID   int64  `json:"user_id" db:"user_id"`
    Avatar   string `json:"avatar" db:"avatar"`
    Bio      string `json:"bio" db:"bio"`
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

// 使用关联
user := NewUser()
err := user.Find(1)
profile, err := user.Profile().First()
```

### 2. HasMany (一对多)

用户有多篇文章：

```go
// 在 User 模型中定义
func (u *User) Posts() *model.HasMany {
    return u.HasMany(&Post{}, "user_id", "id")
}

// 使用关联
user := NewUser()
err := user.Find(1)
posts, err := user.Posts().Get()

// 带条件的关联查询
publishedPosts, err := user.Posts().
    Where("status", "=", "published").
    OrderBy("created_at", "desc").
    Get()
```

### 3. BelongsTo (反向一对一/一对多)

文章属于用户：

```go
// 在 Post 模型中定义
func (p *Post) User() *model.BelongsTo {
    return p.BelongsTo(&User{}, "user_id", "id")
}

// 使用关联
post := NewPost()
err := post.Find(1)
user, err := post.User().First()
```

### 4. ManyToMany (多对多)

用户和角色的多对多关系：

```go
// Role 角色模型
type Role struct {
    model.BaseModel
    ID   int64  `json:"id" db:"id"`
    Name string `json:"name" db:"name"`
}

// 在 User 模型中定义
func (u *User) Roles() *model.ManyToMany {
    return u.ManyToMany(&Role{}, "user_roles", "user_id", "role_id")
}

// 在 Role 模型中定义
func (r *Role) Users() *model.ManyToMany {
    return r.ManyToMany(&User{}, "user_roles", "role_id", "user_id")
}

// 使用关联
user := NewUser()
err := user.Find(1)
roles, err := user.Roles().Get()
```

## 📝 定义关联

### HasOne 关联

```go
// 基础定义
func (u *User) Profile() *model.HasOne {
    return u.HasOne(&Profile{}, "user_id", "id")
}

// 自定义外键和本地键
func (u *User) Profile() *model.HasOne {
    return u.HasOne(&Profile{}, "owner_id", "user_id")
}

// 带默认条件
func (u *User) ActiveProfile() *model.HasOne {
    return u.HasOne(&Profile{}, "user_id", "id").
        Where("status", "=", "active")
}
```

### HasMany 关联

```go
// 基础定义
func (u *User) Posts() *model.HasMany {
    return u.HasMany(&Post{}, "user_id", "id")
}

// 带默认排序
func (u *User) Posts() *model.HasMany {
    return u.HasMany(&Post{}, "user_id", "id").
        OrderBy("created_at", "desc")
}

// 特定类型的文章
func (u *User) PublishedPosts() *model.HasMany {
    return u.HasMany(&Post{}, "user_id", "id").
        Where("status", "=", "published")
}
```

### BelongsTo 关联

```go
// 基础定义
func (p *Post) User() *model.BelongsTo {
    return p.BelongsTo(&User{}, "user_id", "id")
}

// 自定义外键
func (p *Post) Author() *model.BelongsTo {
    return p.BelongsTo(&User{}, "author_id", "id")
}
```

### ManyToMany 关联

```go
// 基础定义
func (u *User) Roles() *model.ManyToMany {
    return u.ManyToMany(&Role{}, "user_roles", "user_id", "role_id")
}

// 带中间表额外字段
func (u *User) Roles() *model.ManyToMany {
    return u.ManyToMany(&Role{}, "user_roles", "user_id", "role_id").
        WithPivot("assigned_at", "assigned_by")
}

// 带中间表条件
func (u *User) ActiveRoles() *model.ManyToMany {
    return u.ManyToMany(&Role{}, "user_roles", "user_id", "role_id").
        WherePivot("status", "=", "active")
}
```

## 🔍 查询关联

### 基础查询

```go
user := NewUser()
err := user.Find(1)

// 获取所有文章
posts, err := user.Posts().Get()

// 获取第一篇文章
firstPost, err := user.Posts().First()

// 计数
postCount, err := user.Posts().Count()

// 检查是否存在
hasPosts, err := user.Posts().Exists()
```

### 条件查询

```go
user := NewUser()
err := user.Find(1)

// 带条件的关联查询
recentPosts, err := user.Posts().
    Where("created_at", ">", "2023-01-01").
    Where("status", "=", "published").
    OrderBy("created_at", "desc").
    Limit(10).
    Get()

// 聚合查询
totalViews, err := user.Posts().Sum("views")
avgScore, err := user.Posts().Avg("score")
```

### 嵌套关联

```go
// 查询用户的文章的评论
user := NewUser()
err := user.Find(1)

posts, err := user.Posts().With("Comments").Get()

// 或者使用点号语法
posts, err := user.Posts().With("Comments.User").Get()
```

## ⚡ 预加载 (Eager Loading)

### 基础预加载

```go
// 预加载用户的档案
users, err := user.With("Profile").Get()

// 预加载多个关联
users, err := user.With("Profile", "Posts").Get()

// 嵌套预加载
users, err := user.With("Posts.Comments").Get()
```

### 高级预加载

```go
// 带条件的预加载
users, err := user.With("Posts", func(q *model.HasMany) *model.HasMany {
    return q.Where("status", "=", "published").
        OrderBy("created_at", "desc").
        Limit(5)
}).Get()

// 多层嵌套预加载
users, err := user.With("Posts.Comments.User").Get()

// 计数预加载
users, err := user.WithCount("Posts", "Comments").Get()
```

### 批量预加载

```go
// 获取用户数据
users, err := user.Where("status", "=", "active").Get()

// 批量预加载关联数据
collection := model.NewModelCollection(users)
err = collection.With("Profile", "Posts").Load()

// 现在可以访问预加载的数据而不会产生额外查询
for _, userInterface := range collection.Models() {
    if u, ok := userInterface.(*User); ok {
        profile := u.GetRelation("Profile")
        posts := u.GetRelation("Posts")
    }
}
```

## 🔧 关联操作

### 创建关联记录

```go
user := NewUser()
err := user.Find(1)

// 创建关联的文章
post := &Post{
    Title:   "新文章",
    Content: "文章内容",
}
createdPost, err := user.Posts().Create(post)

// 批量创建
posts := []*Post{
    {Title: "文章1", Content: "内容1"},
    {Title: "文章2", Content: "内容2"},
}
err = user.Posts().CreateMany(posts)
```

### 关联现有记录

```go
user := NewUser()
err := user.Find(1)

// 关联现有角色
roleIDs := []int64{1, 2, 3}
err = user.Roles().Attach(roleIDs)

// 带中间表数据的关联
err = user.Roles().AttachWithPivot(map[int64]map[string]interface{}{
    1: {"assigned_at": time.Now(), "assigned_by": "admin"},
    2: {"assigned_at": time.Now(), "assigned_by": "admin"},
})
```

### 分离关联

```go
user := NewUser()
err := user.Find(1)

// 分离特定角色
err = user.Roles().Detach([]int64{1, 2})

// 分离所有角色
err = user.Roles().DetachAll()
```

### 同步关联

```go
user := NewUser()
err := user.Find(1)

// 同步角色（删除不在列表中的，添加新的）
newRoleIDs := []int64{2, 3, 4}
err = user.Roles().Sync(newRoleIDs)
```

### 更新关联

```go
user := NewUser()
err := user.Find(1)

// 更新用户的所有文章
err = user.Posts().Update(map[string]interface{}{
    "updated_at": time.Now(),
    "status":     "reviewed",
})

// 更新中间表数据
err = user.Roles().UpdatePivot(1, map[string]interface{}{
    "updated_at": time.Now(),
})
```

## 🎭 多态关联

### 定义多态关联

```go
// Comment 评论模型（可以评论文章或视频）
type Comment struct {
    model.BaseModel
    ID              int64  `json:"id" db:"id"`
    Content         string `json:"content" db:"content"`
    CommentableID   int64  `json:"commentable_id" db:"commentable_id"`
    CommentableType string `json:"commentable_type" db:"commentable_type"`
}

// 在 Post 模型中定义多态关联
func (p *Post) Comments() *model.MorphMany {
    return p.MorphMany(&Comment{}, "commentable")
}

// 在 Video 模型中定义多态关联
func (v *Video) Comments() *model.MorphMany {
    return v.MorphMany(&Comment{}, "commentable")
}

// 在 Comment 模型中定义反向多态关联
func (c *Comment) Commentable() *model.MorphTo {
    return c.MorphTo("commentable")
}
```

### 使用多态关联

```go
// 为文章创建评论
post := NewPost()
err := post.Find(1)

comment := &Comment{Content: "很好的文章！"}
err = post.Comments().Create(comment)

// 获取评论的可评论对象
comment := NewComment()
err := comment.Find(1)
commentable, err := comment.Commentable().First()

// 类型判断
switch v := commentable.(type) {
case *Post:
    fmt.Printf("评论的是文章: %s", v.Title)
case *Video:
    fmt.Printf("评论的是视频: %s", v.Title)
}
```

## 📈 性能优化

### 避免 N+1 查询

```go
// 错误的做法 - 会产生 N+1 查询
users, err := user.Get() // 1 个查询
for _, u := range users {
    posts, _ := u.Posts().Get() // N 个查询
}

// 正确的做法 - 使用预加载
users, err := user.With("Posts").Get() // 2 个查询
for _, u := range users {
    posts := u.GetRelation("Posts") // 无额外查询
}
```

### 选择性加载

```go
// 只加载需要的字段
users, err := user.
    Select("id", "name", "email").
    With("Posts:id,title,user_id").
    Get()

// 条件预加载
users, err := user.With("Posts", func(q *model.HasMany) *model.HasMany {
    return q.Where("status", "=", "published").
        Select("id", "title", "user_id").
        Limit(3)
}).Get()
```

### 关联计数

```go
// 加载关联数量而不是关联数据
users, err := user.WithCount("Posts", "Comments").Get()

// 访问计数
for _, u := range users {
    postCount := u.GetAttribute("posts_count")
    commentCount := u.GetAttribute("comments_count")
}

// 带条件的计数
users, err := user.WithCount("Posts", func(q *model.HasMany) *model.HasMany {
    return q.Where("status", "=", "published")
}, "PublishedPostsCount").Get()
```

### 批量预加载

```go
// 对于大量数据，使用分批预加载
users := make([]*User, 0)
err := user.Chunk(1000, func(chunk []map[string]interface{}) bool {
    // 处理每批数据
    chunkUsers := make([]*User, len(chunk))
    for i, data := range chunk {
        u := NewUser()
        u.Fill(data)
        chunkUsers[i] = u
    }
    
    // 预加载关联数据
    collection := model.NewModelCollection(chunkUsers)
    collection.With("Profile", "Posts").Load()
    
    users = append(users, chunkUsers...)
    return true
})
```

## 🔄 关联缓存

### 关联结果缓存

```go
// 缓存关联查询结果
posts, err := user.Posts().Cache(5 * time.Minute).Get()

// 带标签的缓存
posts, err := user.Posts().
    CacheWithTags(5*time.Minute, "user_posts", fmt.Sprintf("user_%d", user.ID)).
    Get()

// 清除相关缓存
db.FlushCacheByTags("user_posts")
```

### 关联数据同步

```go
// 当用户数据更新时，清除相关缓存
func (u *User) AfterUpdate() error {
    // 清除用户相关的所有缓存
    cacheKey := fmt.Sprintf("user_%d", u.ID)
    db.FlushCacheByTags(cacheKey)
    return nil
}
```

## 📚 最佳实践

### 1. 关联定义

```go
// 好的做法：使用明确的方法名
func (u *User) Posts() *model.HasMany {
    return u.HasMany(&Post{}, "user_id", "id")
}

func (u *User) PublishedPosts() *model.HasMany {
    return u.HasMany(&Post{}, "user_id", "id").
        Where("status", "=", "published")
}

// 避免：在关联中进行复杂的业务逻辑
```

### 2. 性能优化

```go
// 使用预加载避免 N+1 查询
users, err := user.With("Profile", "Posts").Get()

// 只加载需要的字段
users, err := user.
    Select("id", "name").
    With("Posts:id,title,user_id").
    Get()

// 使用关联计数代替加载完整数据
users, err := user.WithCount("Posts").Get()
```

### 3. 错误处理

```go
// 完整的错误处理
user := NewUser()
err := user.Find(1)
if err != nil {
    return fmt.Errorf("查找用户失败: %w", err)
}

posts, err := user.Posts().Get()
if err != nil {
    return fmt.Errorf("查找用户文章失败: %w", err)
}
```

## 🔗 相关文档

- [模型系统](Model-System) - 了解模型的基础功能
- [查询构建器](Query-Builder) - 底层查询构建
- [性能优化](Performance) - 关联查询优化
- [缓存系统](Caching) - 关联数据缓存 