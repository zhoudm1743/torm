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

func TestSimpleRelations(t *testing.T) {
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

	t.Run("基础关联关系测试", func(t *testing.T) {
		// 创建用户
		user := model.NewUser()
		user.SetConnection("test")
		user.SetName("关联测试用户").SetEmail("simple_relation_test@example.com").SetAge(25)
		err = user.Save(ctx)
		require.NoError(t, err)
		assert.NotNil(t, user.GetID())

		// 创建用户资料
		profile := model.NewProfile()
		profile.SetConnection("test")
		profile.SetUserID(user.GetID()).SetBio("测试用户简介")
		err = profile.Save(ctx)
		require.NoError(t, err)
		assert.NotNil(t, profile.GetID())

		// 测试HasOne关系
		userProfileRelation := user.Profile()
		assert.Equal(t, model.HasOneType, userProfileRelation.GetType())

		// 测试BelongsTo关系
		profileUserRelation := profile.User()
		assert.Equal(t, model.BelongsToType, profileUserRelation.GetType())

		// 测试获取查询构造器
		profileQuery, err := userProfileRelation.GetQuery()
		assert.NoError(t, err)
		assert.NotNil(t, profileQuery)

		userQuery, err := profileUserRelation.GetQuery()
		assert.NoError(t, err)
		assert.NotNil(t, userQuery)
	})
}
