package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zhoudm1743/torm/pkg/db"
	"github.com/zhoudm1743/torm/pkg/migration"
)

// setupTestDB 设置测试数据库（使用迁移系统）
func setupTestDB(t *testing.T) {
	config := &db.Config{
		Driver:          "sqlite",
		Database:        ":memory:",
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: time.Hour,
		LogQueries:      false,
	}

	err := db.AddConnection("default", config)
	require.NoError(t, err)

	// 获取连接
	conn, err := db.DB("default")
	require.NoError(t, err)

	err = conn.Connect()
	require.NoError(t, err)

	// 使用迁移工具创建表结构
	migrator := migration.NewMigrator(conn, nil)

	// 注册用户表迁移
	migrator.RegisterFunc("20240101_000001", "创建用户表", func(conn db.ConnectionInterface) error {
		_, err := conn.Exec(`
			CREATE TABLE users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				email TEXT UNIQUE NOT NULL,
				age INTEGER DEFAULT 0,
				status TEXT DEFAULT 'active',
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)
		`)
		return err
	}, func(conn db.ConnectionInterface) error {
		_, err := conn.Exec("DROP TABLE IF EXISTS users")
		return err
	})

	// 注册资料表迁移
	migrator.RegisterFunc("20240101_000002", "创建资料表", func(conn db.ConnectionInterface) error {
		_, err := conn.Exec(`
			CREATE TABLE profiles (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				avatar TEXT,
				bio TEXT,
				website TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			)
		`)
		return err
	}, func(conn db.ConnectionInterface) error {
		_, err := conn.Exec("DROP TABLE IF EXISTS profiles")
		return err
	})

	// 注册文章表迁移
	migrator.RegisterFunc("20240101_000003", "创建文章表", func(conn db.ConnectionInterface) error {
		_, err := conn.Exec(`
			CREATE TABLE posts (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				title TEXT NOT NULL,
				content TEXT NOT NULL,
				status TEXT DEFAULT 'draft',
				published_at DATETIME,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			)
		`)
		return err
	}, func(conn db.ConnectionInterface) error {
		_, err := conn.Exec("DROP TABLE IF EXISTS posts")
		return err
	})

	// 注册标签表迁移
	migrator.RegisterFunc("20240101_000004", "创建标签表", func(conn db.ConnectionInterface) error {
		_, err := conn.Exec(`
			CREATE TABLE tags (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL UNIQUE,
				slug TEXT NOT NULL UNIQUE,
				description TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)
		`)
		return err
	}, func(conn db.ConnectionInterface) error {
		_, err := conn.Exec("DROP TABLE IF EXISTS tags")
		return err
	})

	// 注册文章标签关联表迁移
	migrator.RegisterFunc("20240101_000005", "创建文章标签关联表", func(conn db.ConnectionInterface) error {
		_, err := conn.Exec(`
			CREATE TABLE post_tags (
				post_id INTEGER NOT NULL,
				tag_id INTEGER NOT NULL,
				PRIMARY KEY (post_id, tag_id),
				FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
				FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
			)
		`)
		return err
	}, func(conn db.ConnectionInterface) error {
		_, err := conn.Exec("DROP TABLE IF EXISTS post_tags")
		return err
	})

	// 执行迁移
	err = migrator.Up()
	require.NoError(t, err)
}

func TestUser_Create(t *testing.T) {
	setupTestDB(t)

	user := NewUser()
	user.Name = "测试用户"
	user.Email = "test@example.com"
	user.Age = 25

	err := user.Save()
	require.NoError(t, err)

	// 验证ID被设置
	assert.NotZero(t, user.ID)

	// 验证时间戳被设置
	assert.NotZero(t, user.CreatedAt)
	assert.NotZero(t, user.UpdatedAt)
}

func TestUser_Read(t *testing.T) {
	setupTestDB(t)

	// 先创建用户
	originalUser := NewUser()
	originalUser.Name = "读取测试用户"
	originalUser.Email = "read@example.com"
	originalUser.Age = 30

	err := originalUser.Save()
	require.NoError(t, err)

	// 通过ID读取用户
	readUser := NewUser()
	err = readUser.Find(originalUser.ID)
	require.NoError(t, err)

	// 验证数据正确性
	assert.Equal(t, originalUser.ID, readUser.ID)
	assert.Equal(t, originalUser.Name, readUser.Name)
	assert.Equal(t, originalUser.Email, readUser.Email)
	assert.Equal(t, originalUser.Age, readUser.Age)
}

func TestUser_Update(t *testing.T) {
	setupTestDB(t)

	// 创建用户
	user := NewUser()
	user.Name = "更新前用户"
	user.Email = "update@example.com"
	user.Age = 25

	err := user.Save()
	require.NoError(t, err)

	originalUpdatedAt := user.UpdatedAt

	// 等待一秒确保时间戳会变化
	time.Sleep(1 * time.Second)

	// 更新用户
	user.Name = "更新后用户"
	user.Age = 26

	err = user.Save()
	require.NoError(t, err)

	// 验证更新时间戳变化
	assert.True(t, user.UpdatedAt.After(originalUpdatedAt))

	// 重新读取验证
	updatedUser := NewUser()
	err = updatedUser.Find(user.ID)
	require.NoError(t, err)

	assert.Equal(t, "更新后用户", updatedUser.Name)
	assert.Equal(t, 26, updatedUser.Age)
}

func TestUser_Delete(t *testing.T) {
	setupTestDB(t)

	// 创建用户
	user := NewUser()
	user.Name = "待删除用户"
	user.Email = "delete@example.com"
	user.Age = 25

	err := user.Save()
	require.NoError(t, err)

	userID := user.ID

	// 删除用户
	err = user.Delete()
	require.NoError(t, err)

	// 验证用户已被删除
	deletedUser := NewUser()
	err = deletedUser.Find(userID)
	assert.Error(t, err)
}

func TestUser_Validation(t *testing.T) {
	setupTestDB(t)

	// 测试邮箱验证
	user := NewUser()
	user.Name = "验证测试用户"
	user.Email = "invalid-email"
	user.Age = 25

	err := user.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "邮箱格式不正确")

	// 测试年龄验证
	user.Email = "valid@example.com"
	user.Age = -5

	err = user.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "年龄必须大于0")

	// 测试正确的数据
	user.Age = 25
	err = user.Validate()
	assert.NoError(t, err)
}

func TestUser_FindByEmail(t *testing.T) {
	setupTestDB(t)

	// 创建用户
	originalUser := NewUser()
	originalUser.Name = "邮箱查找测试"
	originalUser.Email = "findemail@example.com"
	originalUser.Age = 28

	err := originalUser.Save()
	require.NoError(t, err)

	// 通过邮箱查找
	foundUser, err := FindByEmail("findemail@example.com")
	require.NoError(t, err)

	assert.Equal(t, originalUser.ID, foundUser.ID)
	assert.Equal(t, originalUser.Name, foundUser.Name)
	assert.Equal(t, originalUser.Email, foundUser.Email)

	// 测试查找不存在的邮箱
	_, err = FindByEmail("nonexistent@example.com")
	assert.Error(t, err)
}

func TestUser_FindActiveUsers(t *testing.T) {
	setupTestDB(t)

	// 创建多个用户
	for i := 0; i < 5; i++ {
		user := NewUser()
		user.Name = fmt.Sprintf("活跃用户%d", i+1)
		user.Email = fmt.Sprintf("active%d@example.com", i+1)
		user.Age = 20 + i
		user.Status = "active"

		err := user.Save()
		require.NoError(t, err)
	}

	// 创建一个非活跃用户
	inactiveUser := NewUser()
	inactiveUser.Name = "非活跃用户"
	inactiveUser.Email = "inactive@example.com"
	inactiveUser.Age = 30
	inactiveUser.Status = "inactive"

	err := inactiveUser.Save()
	require.NoError(t, err)

	// 查找活跃用户
	activeUsers, err := FindActiveUsers(10)
	require.NoError(t, err)

	// 验证只返回活跃用户
	assert.Equal(t, 5, len(activeUsers))
	for _, user := range activeUsers {
		assert.Equal(t, "active", user.Status)
	}
}

func TestUser_CountByStatus(t *testing.T) {
	setupTestDB(t)

	// 创建不同状态的用户
	statuses := []string{"active", "active", "inactive", "pending", "active"}
	for i, status := range statuses {
		user := NewUser()
		user.Name = fmt.Sprintf("状态测试用户%d", i+1)
		user.Email = fmt.Sprintf("status%d@example.com", i+1)
		user.Age = 20 + i
		user.Status = status

		err := user.Save()
		require.NoError(t, err)
	}

	// 统计各状态用户数量
	activeCount, err := CountByStatus("active")
	require.NoError(t, err)
	assert.Equal(t, int64(3), activeCount)

	inactiveCount, err := CountByStatus("inactive")
	require.NoError(t, err)
	assert.Equal(t, int64(1), inactiveCount)

	pendingCount, err := CountByStatus("pending")
	require.NoError(t, err)
	assert.Equal(t, int64(1), pendingCount)

	nonExistentCount, err := CountByStatus("deleted")
	require.NoError(t, err)
	assert.Equal(t, int64(0), nonExistentCount)
}

func TestUser_BatchOperations(t *testing.T) {
	setupTestDB(t)

	// 批量创建用户
	users := make([]*User, 10)
	for i := 0; i < 10; i++ {
		user := NewUser()
		user.Name = fmt.Sprintf("批量用户%d", i+1)
		user.Email = fmt.Sprintf("batch%d@example.com", i+1)
		user.Age = 20 + i

		err := user.Save()
		require.NoError(t, err)

		users[i] = user
	}

	// 验证所有用户都被创建
	for _, user := range users {
		assert.NotZero(t, user.ID)
	}

	// 批量查询验证
	activeUsers, err := FindActiveUsers(20)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(activeUsers), 10)
}

func TestProfile_Relations(t *testing.T) {
	setupTestDB(t)

	// 创建用户
	user := NewUser()
	user.Name = "关系测试用户"
	user.Email = "relation@example.com"
	user.Age = 30

	err := user.Save()
	require.NoError(t, err)

	// 创建用户资料
	profile := NewProfile()
	profile.UserID = user.ID
	profile.Bio = "这是个人简介"
	profile.Website = "https://example.com"

	err = profile.Save()
	require.NoError(t, err)

	// 测试通过函数查找资料
	foundProfile, err := FindProfileByUserID(user.ID)
	require.NoError(t, err)
	assert.Equal(t, profile.Bio, foundProfile.Bio)
}

func TestPost_Operations(t *testing.T) {
	setupTestDB(t)

	// 创建用户
	user := NewUser()
	user.Name = "文章作者"
	user.Email = "author@example.com"
	user.Age = 35

	err := user.Save()
	require.NoError(t, err)

	// 创建文章
	post := NewPost()
	post.UserID = user.ID
	post.Title = "测试文章"
	post.Content = "这是测试文章的内容"
	post.Status = "published"
	now := time.Now()
	post.PublishedAt = &now

	err = post.Save()
	require.NoError(t, err)

	// 测试查找用户的文章
	posts, err := FindPostsByUserID(user.ID)
	require.NoError(t, err)
	assert.Len(t, posts, 1)
	assert.Equal(t, "测试文章", posts[0].Title)

	// 测试查找已发布的文章
	publishedPosts, err := FindPublishedPosts(10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(publishedPosts), 1)
}

func TestTag_Operations(t *testing.T) {
	setupTestDB(t)

	// 创建标签
	tag := NewTag()
	tag.Name = "Go语言"
	tag.Slug = "golang"
	tag.Description = "Go编程语言相关内容"

	err := tag.Save()
	require.NoError(t, err)

	// 测试根据名称查找标签
	foundTag, err := FindTagByName("Go语言")
	require.NoError(t, err)
	assert.Equal(t, tag.ID, foundTag.ID)
	assert.Equal(t, tag.Name, foundTag.Name)

	// 测试根据slug查找标签
	foundBySlug, err := FindTagBySlug("golang")
	require.NoError(t, err)
	assert.Equal(t, tag.ID, foundBySlug.ID)

	// 测试查找热门标签
	popularTags, err := FindPopularTags(10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(popularTags), 1)
}

// Benchmark tests
func BenchmarkUser_Create(b *testing.B) {
	setupTestDB(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := NewUser()
		user.Name = "基准测试用户"
		user.Email = fmt.Sprintf("bench%d@example.com", i)
		user.Age = 25

		user.Save()
	}
}

func BenchmarkUser_Read(b *testing.B) {
	setupTestDB(nil)

	// 创建一个用户用于读取测试
	user := NewUser()
	user.Name = "读取基准测试"
	user.Email = "readbench@example.com"
	user.Age = 25
	user.Save()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		readUser := NewUser()
		readUser.Find(user.ID)
	}
}
