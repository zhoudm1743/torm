package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"torm/pkg/db"
	"torm/pkg/model"
)

func setupRelationTest(t *testing.T) context.Context {
	// 配置数据库连接
	config := &db.Config{
		Driver:          "mysql",
		Host:            "127.0.0.1",
		Port:            3306,
		Database:        "orm",
		Username:        "root",
		Password:        "123456",
		Charset:         "utf8mb4",
		Timezone:        "UTC",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		LogQueries:      false,
	}

	err := db.AddConnection("test", config)
	require.NoError(t, err)

	ctx := context.Background()
	conn, err := db.DB("test")
	require.NoError(t, err)

	// 创建测试表
	conn.Exec(ctx, "DROP TABLE IF EXISTS post_tags")
	conn.Exec(ctx, "DROP TABLE IF EXISTS posts")
	conn.Exec(ctx, "DROP TABLE IF EXISTS tags")
	conn.Exec(ctx, "DROP TABLE IF EXISTS profiles")
	conn.Exec(ctx, "DROP TABLE IF EXISTS users")

	// 创建用户表
	conn.Exec(ctx, `
		CREATE TABLE users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			age INT DEFAULT 0,
			status ENUM('active', 'inactive', 'pending') DEFAULT 'active',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`)

	// 创建用户资料表
	conn.Exec(ctx, `
		CREATE TABLE profiles (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			avatar VARCHAR(255),
			bio TEXT,
			website VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`)

	// 创建文章表
	conn.Exec(ctx, `
		CREATE TABLE posts (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			title VARCHAR(255) NOT NULL,
			content TEXT NOT NULL,
			status ENUM('draft', 'published', 'archived') DEFAULT 'draft',
			published_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`)

	// 创建标签表
	conn.Exec(ctx, `
		CREATE TABLE tags (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(50) NOT NULL UNIQUE,
			slug VARCHAR(50) NOT NULL UNIQUE,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`)

	// 创建文章标签关联表
	conn.Exec(ctx, `
		CREATE TABLE post_tags (
			post_id INT NOT NULL,
			tag_id INT NOT NULL,
			PRIMARY KEY (post_id, tag_id),
			FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
			FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`)

	return ctx
}

func TestHasOneRelation(t *testing.T) {
	ctx := setupRelationTest(t)

	// 创建用户
	user := model.NewUser()
	user.SetConnection("test")
	user.SetName("测试用户").SetEmail("hasone_test@example.com").SetAge(30)
	err := user.Save(ctx)
	require.NoError(t, err)
	assert.NotNil(t, user.GetID())

	// 创建用户资料
	profile := model.NewProfile()
	profile.SetConnection("test")
	profile.SetUserID(user.GetID()).
		SetAvatar("https://example.com/avatar.jpg").
		SetBio("测试用户的个人简介").
		SetWebsite("https://test.com")

	err = profile.Save(ctx)
	require.NoError(t, err)
	assert.NotNil(t, profile.GetID())

	// 测试HasOne关系
	t.Run("HasOne关系基本功能", func(t *testing.T) {
		relation := user.Profile()
		assert.Equal(t, model.HasOneType, relation.GetType())

		// 测试获取查询构造器
		query, err := relation.GetQuery()
		assert.NoError(t, err)
		assert.NotNil(t, query)

		// 测试手动查询
		profileData, err := query.Where("user_id", "=", user.GetID()).First(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, profileData)
		assert.Equal(t, user.GetID(), profileData["user_id"])
	})

	t.Run("HasOne关系查询条件", func(t *testing.T) {
		relation := user.Profile().Where("bio", "!=", "")
		query, err := relation.GetQuery()
		assert.NoError(t, err)

		profileData, err := query.Where("user_id", "=", user.GetID()).First(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, profileData)
	})
}

func TestBelongsToRelation(t *testing.T) {
	ctx := setupRelationTest(t)

	// 创建用户
	user := model.NewUser()
	user.SetName("测试用户").SetEmail("belongsto_test@example.com").SetAge(25)
	err := user.Save(ctx)
	require.NoError(t, err)

	// 创建用户资料
	profile := model.NewProfile()
	profile.SetUserID(user.GetID()).
		SetBio("测试资料")

	err = profile.Save(ctx)
	require.NoError(t, err)

	// 测试BelongsTo关系
	t.Run("BelongsTo关系基本功能", func(t *testing.T) {
		relation := profile.User()
		assert.Equal(t, model.BelongsToType, relation.GetType())

		// 测试获取查询构造器
		query, err := relation.GetQuery()
		assert.NoError(t, err)
		assert.NotNil(t, query)

		// 测试手动查询
		userData, err := query.Where("id", "=", profile.GetUserID()).First(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, userData)
		assert.Equal(t, profile.GetUserID(), userData["id"])
	})
}

func TestHasManyRelation(t *testing.T) {
	ctx := setupRelationTest(t)

	// 创建用户
	user := model.NewUser()
	user.SetName("文章作者").SetEmail("hasmany_test@example.com").SetAge(35)
	err := user.Save(ctx)
	require.NoError(t, err)

	// 创建多篇文章
	posts := []*model.Post{
		model.NewPost().SetUserID(user.GetID()).SetTitle("第一篇文章").SetContent("内容1").SetStatus("published"),
		model.NewPost().SetUserID(user.GetID()).SetTitle("第二篇文章").SetContent("内容2").SetStatus("draft"),
		model.NewPost().SetUserID(user.GetID()).SetTitle("第三篇文章").SetContent("内容3").SetStatus("published"),
	}

	for _, post := range posts {
		err = post.Save(ctx)
		require.NoError(t, err)
		assert.NotNil(t, post.GetID())
	}

	// 测试HasMany关系
	t.Run("HasMany关系基本功能", func(t *testing.T) {
		relation := user.Posts()
		assert.Equal(t, model.HasManyType, relation.GetType())

		// 测试获取查询构造器
		query, err := relation.GetQuery()
		assert.NoError(t, err)
		assert.NotNil(t, query)

		// 测试手动查询所有文章
		postsData, err := query.Where("user_id", "=", user.GetID()).Get(ctx)
		assert.NoError(t, err)
		assert.Len(t, postsData, 3)
	})

	t.Run("HasMany关系查询条件", func(t *testing.T) {
		relation := user.Posts().Where("status", "=", "published")
		query, err := relation.GetQuery()
		assert.NoError(t, err)

		// 测试查询已发布文章
		publishedPosts, err := query.Where("user_id", "=", user.GetID()).Get(ctx)
		assert.NoError(t, err)
		assert.Len(t, publishedPosts, 2) // 应该有2篇已发布的文章
	})
}

func TestManyToManyRelation(t *testing.T) {
	ctx := setupRelationTest(t)

	// 创建用户和文章
	user := model.NewUser()
	user.SetName("作者").SetEmail("manytomany_test@example.com").SetAge(28)
	err := user.Save(ctx)
	require.NoError(t, err)

	post := model.NewPost()
	post.SetUserID(user.GetID()).SetTitle("测试文章").SetContent("测试内容").SetStatus("published")
	err = post.Save(ctx)
	require.NoError(t, err)

	// 创建标签
	tags := []*model.Tag{
		model.NewTag().SetName("Go语言"),
		model.NewTag().SetName("数据库"),
		model.NewTag().SetName("后端开发"),
	}

	for _, tag := range tags {
		err = tag.Save(ctx)
		require.NoError(t, err)
		assert.NotNil(t, tag.GetID())
	}

	// 测试ManyToMany关系
	t.Run("ManyToMany关系基本功能", func(t *testing.T) {
		postTagsRelation := post.Tags()
		assert.Equal(t, model.ManyToManyType, postTagsRelation.GetType())

		// 测试关联
		err = postTagsRelation.Associate(ctx, tags[0])
		assert.NoError(t, err)

		err = postTagsRelation.Associate(ctx, tags[1])
		assert.NoError(t, err)

		// 测试重复关联（应该不会出错）
		err = postTagsRelation.Associate(ctx, tags[0])
		assert.NoError(t, err)

		// 验证关联数据
		_, err = postTagsRelation.GetQuery()
		assert.NoError(t, err)

		// 手动查询验证
		conn, _ := db.DB("test")
		pivotData, err := conn.Query(ctx, "SELECT * FROM post_tags WHERE post_id = ?", post.GetID())
		assert.NoError(t, err)
		defer pivotData.Close()

		count := 0
		for pivotData.Next() {
			count++
		}
		assert.Equal(t, 2, count) // 应该有2个关联记录
	})

	t.Run("ManyToMany取消关联", func(t *testing.T) {
		postTagsRelation := post.Tags()

		// 取消关联
		err = postTagsRelation.Dissociate(ctx, tags[0])
		assert.NoError(t, err)

		// 验证关联已取消
		conn, _ := db.DB("test")
		pivotData, err := conn.Query(ctx, "SELECT * FROM post_tags WHERE post_id = ? AND tag_id = ?", post.GetID(), tags[0].GetID())
		assert.NoError(t, err)
		defer pivotData.Close()

		hasRows := pivotData.Next()
		assert.False(t, hasRows) // 应该没有关联记录
	})

	t.Run("反向ManyToMany关系", func(t *testing.T) {
		tagPostsRelation := tags[1].Posts()
		assert.Equal(t, model.ManyToManyType, tagPostsRelation.GetType())

		// 验证反向关联
		query, err := tagPostsRelation.GetQuery()
		assert.NoError(t, err)
		assert.NotNil(t, query)
	})
}

func TestRelationAssociation(t *testing.T) {
	ctx := setupRelationTest(t)

	// 创建用户
	user := model.NewUser()
	user.SetName("关联测试用户").SetEmail("association_test@example.com").SetAge(30)
	err := user.Save(ctx)
	require.NoError(t, err)

	t.Run("HasOne关联操作", func(t *testing.T) {
		profile := model.NewProfile()
		profile.SetBio("新建资料")

		relation := user.Profile()

		// 测试关联
		err = relation.Associate(ctx, profile)
		assert.NoError(t, err)
		assert.Equal(t, user.GetID(), profile.GetUserID())
		assert.NotNil(t, profile.GetID()) // 应该已保存

		// 测试取消关联
		err = relation.Dissociate(ctx, profile)
		assert.NoError(t, err)
		assert.Nil(t, profile.GetUserID()) // 外键应该被清空
	})

	t.Run("BelongsTo关联操作", func(t *testing.T) {
		profile := model.NewProfile()
		profile.SetBio("BelongsTo测试")
		err = profile.Save(ctx)
		require.NoError(t, err)

		relation := profile.User()

		// 测试关联
		err = relation.Associate(ctx, user)
		assert.NoError(t, err)
		assert.Equal(t, user.GetID(), profile.GetUserID())

		// 测试取消关联
		err = relation.Dissociate(ctx, user)
		assert.NoError(t, err)
		assert.Nil(t, profile.GetUserID())
	})
}
