package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"torm/pkg/db"
	"torm/pkg/migration"
)

// setupTestDatabase 设置测试数据库连接
func setupTestDatabase(t *testing.T) db.ConnectionInterface {
	config := &db.Config{
		Driver:          "sqlite",
		Database:        ":memory:", // 使用内存数据库避免文件锁定
		MaxOpenConns:    1,          // SQLite使用单连接避免并发问题
		MaxIdleConns:    1,
		ConnMaxLifetime: time.Hour,
		LogQueries:      false,
	}

	// 添加连接配置
	err := db.AddConnection("test", config)
	if t != nil {
		require.NoError(t, err)
	}

	// 获取连接
	conn, err := db.DB("test")
	if t != nil {
		require.NoError(t, err)
	}

	// 连接数据库
	err = conn.Connect()
	if t != nil {
		require.NoError(t, err)
	}

	return conn
}

func TestMigrationBasic(t *testing.T) {
	conn := setupTestDatabase(t)
	defer conn.Close()

	// 创建迁移器
	migrator := migration.NewMigrator(conn, nil)

	// 注册测试迁移
	migrator.RegisterFunc("20240101_000001", "创建用户表",
		func(conn db.ConnectionInterface) error {
			// 创建用户表
			schema := migration.NewSchemaBuilder(conn)

			table := &migration.Table{
				Name: "test_users",
				Columns: []*migration.Column{
					{Name: "id", Type: migration.ColumnTypeInt, PrimaryKey: true, AutoIncrement: true},
					{Name: "name", Type: migration.ColumnTypeVarchar, Length: 100, NotNull: true},
					{Name: "email", Type: migration.ColumnTypeVarchar, Length: 100, NotNull: true, Unique: true},
					{Name: "created_at", Type: migration.ColumnTypeTimestamp, Default: "CURRENT_TIMESTAMP"},
				},
			}

			return schema.CreateTable(table)
		},
		func(conn db.ConnectionInterface) error {
			// 删除用户表
			schema := migration.NewSchemaBuilder(conn)
			return schema.DropTable("test_users")
		},
	)

	// 检查初始状态
	status, err := migrator.Status()
	require.NoError(t, err)
	assert.Len(t, status, 1)
	assert.False(t, status[0].Applied)

	// 执行迁移
	err = migrator.Up()
	require.NoError(t, err)

	// 验证迁移状态
	status, err = migrator.Status()
	require.NoError(t, err)
	assert.Len(t, status, 1)
	assert.True(t, status[0].Applied)

	// 验证表已创建
	rows, err := conn.Query("SELECT name FROM sqlite_master WHERE type='table' AND name='test_users'")
	require.NoError(t, err)
	defer rows.Close()

	assert.True(t, rows.Next())

	var tableName string
	err = rows.Scan(&tableName)
	require.NoError(t, err)
	assert.Equal(t, "test_users", tableName)
}

func TestMigrationMultiple(t *testing.T) {
	conn := setupTestDatabase(t)
	defer conn.Close()

	migrator := migration.NewMigrator(conn, nil)

	// 注册多个迁移
	migrator.RegisterFunc("20240101_000001", "创建用户表",
		func(conn db.ConnectionInterface) error {
			schema := migration.NewSchemaBuilder(conn)
			table := &migration.Table{
				Name: "users",
				Columns: []*migration.Column{
					{Name: "id", Type: migration.ColumnTypeInt, PrimaryKey: true, AutoIncrement: true},
					{Name: "name", Type: migration.ColumnTypeVarchar, Length: 100, NotNull: true},
				},
			}
			return schema.CreateTable(table)
		},
		func(conn db.ConnectionInterface) error {
			schema := migration.NewSchemaBuilder(conn)
			return schema.DropTable("users")
		},
	)

	migrator.RegisterFunc("20240101_000002", "添加邮箱字段",
		func(conn db.ConnectionInterface) error {
			schema := migration.NewSchemaBuilder(conn)
			column := &migration.Column{
				Name:    "email",
				Type:    migration.ColumnTypeVarchar,
				Length:  100,
				NotNull: true,
			}
			return schema.AddColumn("users", column)
		},
		func(conn db.ConnectionInterface) error {
			schema := migration.NewSchemaBuilder(conn)
			return schema.DropColumn("users", "email")
		},
	)

	// 执行所有迁移
	err := migrator.Up()
	require.NoError(t, err)

	// 验证所有迁移都已应用
	status, err := migrator.Status()
	require.NoError(t, err)
	assert.Len(t, status, 2)

	for _, s := range status {
		assert.True(t, s.Applied)
	}
}

func TestSchemaBuilder(t *testing.T) {
	conn := setupTestDatabase(t)
	defer conn.Close()

	schema := migration.NewSchemaBuilder(conn)

	// 测试创建表
	table := &migration.Table{
		Name: "schema_test",
		Columns: []*migration.Column{
			{Name: "id", Type: migration.ColumnTypeInt, PrimaryKey: true, AutoIncrement: true},
			{Name: "name", Type: migration.ColumnTypeVarchar, Length: 50, NotNull: true},
			{Name: "email", Type: migration.ColumnTypeVarchar, Length: 100, Unique: true},
			{Name: "age", Type: migration.ColumnTypeInt, Default: 0},
			{Name: "created_at", Type: migration.ColumnTypeTimestamp, Default: "CURRENT_TIMESTAMP"},
		},
		Indexes: []*migration.Index{
			{Name: "idx_name", Columns: []string{"name"}},
			{Name: "idx_age", Columns: []string{"age"}},
		},
	}

	err := schema.CreateTable(table)
	require.NoError(t, err)

	// 验证表存在
	rows, err := conn.Query("SELECT name FROM sqlite_master WHERE type='table' AND name='schema_test'")
	require.NoError(t, err)
	defer rows.Close()
	assert.True(t, rows.Next())
}

func TestMigrationError(t *testing.T) {
	conn := setupTestDatabase(t)
	defer conn.Close()

	migrator := migration.NewMigrator(conn, nil)

	// 注册一个会失败的迁移
	migrator.RegisterFunc("20240101_000001", "失败迁移",
		func(conn db.ConnectionInterface) error {
			// 故意返回错误
			return assert.AnError
		},
		func(conn db.ConnectionInterface) error {
			return nil
		},
	)

	// 执行迁移应该失败
	err := migrator.Up()
	assert.Error(t, err)

	// 验证迁移没有被记录为已应用
	status, err := migrator.Status()
	require.NoError(t, err)
	assert.Len(t, status, 1)
	assert.False(t, status[0].Applied)
}

func TestMigrationPrintStatus(t *testing.T) {
	conn := setupTestDatabase(t)
	defer conn.Close()

	migrator := migration.NewMigrator(conn, nil)

	// 注册测试迁移
	migrator.RegisterFunc("20240101_000001", "打印状态测试",
		func(conn db.ConnectionInterface) error {
			return nil
		},
		func(conn db.ConnectionInterface) error {
			return nil
		},
	)

	// 打印状态应该不出错
	err := migrator.PrintStatus()
	assert.NoError(t, err)

	// 执行迁移后再打印
	err = migrator.Up()
	require.NoError(t, err)

	err = migrator.PrintStatus()
	assert.NoError(t, err)
}
