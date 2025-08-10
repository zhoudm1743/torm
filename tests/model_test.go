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

// setupTestDB 设置测试数据库
func setupTestDB(t *testing.T) {
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

	err := db.AddConnection("default", config)
	require.NoError(t, err)

	// 创建测试表
	conn, err := db.DB("default")
	require.NoError(t, err)

	ctx := context.Background()

	// 删除并重新创建表
	_, err = conn.Exec(ctx, "DROP TABLE IF EXISTS users")
	require.NoError(t, err)

	createTableSQL := `
	CREATE TABLE users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		age INT DEFAULT 0,
		status ENUM('active', 'inactive', 'pending') DEFAULT 'active',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP NULL DEFAULT NULL
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`
	_, err = conn.Exec(ctx, createTableSQL)
	require.NoError(t, err)
}

func TestNewUser(t *testing.T) {
	user := model.NewUser()

	assert.True(t, user.IsNew())
	assert.False(t, user.Exists())
	assert.Equal(t, "users", user.TableName())
	assert.Equal(t, "id", user.PrimaryKey())
	assert.Equal(t, "default", user.GetConnection())
}

func TestUserAttributes(t *testing.T) {
	user := model.NewUser()

	// 测试设置和获取属性
	user.SetName("张三").SetEmail("zhangsan@example.com").SetAge(25)

	assert.Equal(t, "张三", user.GetName())
	assert.Equal(t, "zhangsan@example.com", user.GetEmail())
	assert.Equal(t, 25, user.GetAge())

	// 测试默认值
	assert.Equal(t, "", user.GetStatus())
	assert.Equal(t, nil, user.GetID())
}

func TestUserBusinessMethods(t *testing.T) {
	user := model.NewUser()
	user.SetAge(25).SetStatus("active")

	assert.True(t, user.IsActive())
	assert.False(t, user.IsPending())
	assert.False(t, user.IsInactive())
	assert.True(t, user.IsAdult())

	// 测试状态切换
	user.Deactivate()
	assert.True(t, user.IsInactive())

	user.Activate()
	assert.True(t, user.IsActive())
}

func TestUserValidation(t *testing.T) {
	user := model.NewUser()

	// 测试验证失败
	user.SetName("").SetEmail("invalid-email").SetAge(-5)
	err := user.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "姓名不能为空")
	assert.Contains(t, err.Error(), "邮箱格式不正确")
	assert.Contains(t, err.Error(), "年龄必须在0-150之间")

	// 测试验证成功
	user.SetName("张三").SetEmail("zhangsan@example.com").SetAge(25).SetStatus("active")
	err = user.Validate()
	assert.NoError(t, err)
}

func TestUserCRUD(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	// 测试创建用户
	user := model.NewUser()
	user.SetName("测试用户").SetEmail("test@example.com").SetAge(30)

	assert.True(t, user.IsNew())
	assert.True(t, user.IsDirty())

	err := user.Save(ctx)
	require.NoError(t, err)

	assert.False(t, user.IsNew())
	assert.True(t, user.Exists())
	assert.NotNil(t, user.GetID())
	assert.False(t, user.IsDirty()) // 保存后应该没有未保存的更改

	userID := user.GetID()

	// 测试更新用户
	user.SetAge(31)
	assert.True(t, user.IsDirty())

	dirty := user.GetDirty()
	assert.Contains(t, dirty, "age")
	assert.Equal(t, 31, dirty["age"])

	err = user.Save(ctx)
	require.NoError(t, err)
	assert.False(t, user.IsDirty())

	// 测试查找用户
	foundUser := model.NewUser()
	err = foundUser.Find(ctx, userID)
	require.NoError(t, err)

	assert.Equal(t, user.GetName(), foundUser.GetName())
	assert.Equal(t, user.GetEmail(), foundUser.GetEmail())
	assert.Equal(t, user.GetAge(), foundUser.GetAge())
	assert.False(t, foundUser.IsNew())
	assert.True(t, foundUser.Exists())

	// 测试重载
	// 直接修改数据库
	conn, _ := db.DB("default")
	_, err = conn.Exec(ctx, "UPDATE users SET age = 99 WHERE id = ?", userID)
	require.NoError(t, err)

	assert.Equal(t, 31, user.GetAge()) // 模型中的值还是旧的

	err = user.Reload(ctx)
	require.NoError(t, err)
	assert.Equal(t, 99, user.GetAge()) // 重载后应该是新值

	// 测试删除用户
	err = user.Delete(ctx)
	require.NoError(t, err)
	assert.False(t, user.Exists())
}

func TestUserSoftDelete(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	// 创建启用软删除的用户
	user := model.NewUser()
	user.EnableSoftDeletes()
	user.SetName("软删除用户").SetEmail("soft@example.com").SetAge(25)

	err := user.Save(ctx)
	require.NoError(t, err)

	userID := user.GetID()

	// 执行软删除
	err = user.Delete(ctx)
	require.NoError(t, err)

	assert.True(t, user.Exists()) // 软删除后仍然存在
	assert.NotNil(t, user.GetAttribute("deleted_at"))

	// 验证数据库中的记录
	conn, _ := db.DB("default")
	row := conn.QueryRow(ctx, "SELECT deleted_at FROM users WHERE id = ?", userID)
	var deletedAt interface{}
	err = row.Scan(&deletedAt)
	require.NoError(t, err)
	assert.NotNil(t, deletedAt)
}

func TestStaticQueryMethods(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	// 创建测试用户
	users := []*model.User{
		model.NewUser().SetName("用户1").SetEmail("user1@example.com").SetAge(20).SetStatus("active"),
		model.NewUser().SetName("用户2").SetEmail("user2@example.com").SetAge(25).SetStatus("active"),
		model.NewUser().SetName("用户3").SetEmail("user3@example.com").SetAge(16).SetStatus("pending"),
		model.NewUser().SetName("用户4").SetEmail("user4@example.com").SetAge(30).SetStatus("inactive"),
	}

	for _, user := range users {
		err := user.Save(ctx)
		require.NoError(t, err)
	}

	// 测试根据邮箱查找
	foundUser, err := model.FindByEmail(ctx, "user1@example.com")
	require.NoError(t, err)
	assert.Equal(t, "用户1", foundUser.GetName())

	// 测试查找活跃用户
	activeUsers, err := model.FindActiveUsers(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, activeUsers, 2)

	// 测试查找成年用户
	adultUsers, err := model.FindAdultUsers(ctx)
	require.NoError(t, err)
	assert.Len(t, adultUsers, 2) // 只有20岁和25岁的用户

	// 测试按状态统计
	activeCount, err := model.CountByStatus(ctx, "active")
	require.NoError(t, err)
	assert.Equal(t, int64(2), activeCount)

	pendingCount, err := model.CountByStatus(ctx, "pending")
	require.NoError(t, err)
	assert.Equal(t, int64(1), pendingCount)

	inactiveCount, err := model.CountByStatus(ctx, "inactive")
	require.NoError(t, err)
	assert.Equal(t, int64(1), inactiveCount)
}

func TestUserFill(t *testing.T) {
	user := model.NewUser()

	data := map[string]interface{}{
		"name":   "填充用户",
		"email":  "fill@example.com",
		"age":    28,
		"status": "active",
	}

	user.Fill(data)

	assert.Equal(t, "填充用户", user.GetName())
	assert.Equal(t, "fill@example.com", user.GetEmail())
	assert.Equal(t, 28, user.GetAge())
	assert.Equal(t, "active", user.GetStatus())
}

func TestUserWithData(t *testing.T) {
	data := map[string]interface{}{
		"id":     123,
		"name":   "已存在用户",
		"email":  "existing@example.com",
		"age":    35,
		"status": "active",
	}

	user := model.NewUserWithData(data)

	assert.False(t, user.IsNew()) // 有ID的用户不是新记录
	assert.True(t, user.Exists())
	assert.Equal(t, 123, user.GetID())
	assert.Equal(t, "已存在用户", user.GetName())
	assert.False(t, user.IsDirty()) // 刚创建时没有更改
}

func TestUserToMap(t *testing.T) {
	user := model.NewUser()
	user.SetName("测试用户").SetEmail("test@example.com").SetAge(25)

	userMap := user.ToMap()

	assert.Equal(t, "测试用户", userMap["name"])
	assert.Equal(t, "test@example.com", userMap["email"])
	assert.Equal(t, 25, userMap["age"])
}

func TestUserString(t *testing.T) {
	user := model.NewUser()
	user.SetName("测试用户").SetEmail("test@example.com")

	str := user.String()
	assert.Contains(t, str, "users{")
	assert.Contains(t, str, "name: 测试用户")
	assert.Contains(t, str, "email: test@example.com")
}
