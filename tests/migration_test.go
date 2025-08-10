package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"torm/pkg/db"
	"torm/pkg/migration"
)

func setupMigrationTest(t *testing.T) (context.Context, *migration.Migrator) {
	ctx := context.Background()

	// 配置SQLite测试数据库
	config := &db.Config{
		Driver:          "sqlite",
		Database:        "test_migration.db",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		LogQueries:      false,
	}

	err := db.AddConnection("test_migration", config)
	require.NoError(t, err)

	conn, err := db.DB("test_migration")
	require.NoError(t, err)

	migrator := migration.NewMigrator(conn, nil)
	return ctx, migrator
}

func cleanupMigrationTest(t *testing.T) {
	// 删除测试数据库文件
	os.Remove("test_migration.db")
}

func TestMigrationBasicOperations(t *testing.T) {
	ctx, migrator := setupMigrationTest(t)
	defer cleanupMigrationTest(t)

	// 注册简单的测试迁移
	migrator.RegisterFunc(
		"20240101_001",
		"创建测试表",
		func(ctx context.Context, conn db.ConnectionInterface) error {
			schema := migration.NewSchemaBuilder(conn)
			table := &migration.Table{
				Name: "test_table",
				Columns: []*migration.Column{
					{
						Name:          "id",
						Type:          migration.ColumnTypeBigInt,
						PrimaryKey:    true,
						AutoIncrement: true,
						NotNull:       true,
					},
					{
						Name:    "name",
						Type:    migration.ColumnTypeVarchar,
						Length:  100,
						NotNull: true,
					},
				},
			}
			return schema.CreateTable(ctx, table)
		},
		func(ctx context.Context, conn db.ConnectionInterface) error {
			schema := migration.NewSchemaBuilder(conn)
			return schema.DropTable(ctx, "test_table")
		},
	)

	t.Run("初始状态检查", func(t *testing.T) {
		status, err := migrator.Status(ctx)
		assert.NoError(t, err)
		assert.Len(t, status, 1)
		assert.Equal(t, "20240101_001", status[0].Version)
		assert.False(t, status[0].Applied)
	})

	t.Run("执行迁移", func(t *testing.T) {
		err := migrator.Up(ctx)
		assert.NoError(t, err)

		// 检查状态
		status, err := migrator.Status(ctx)
		assert.NoError(t, err)
		assert.Len(t, status, 1)
		assert.True(t, status[0].Applied)
		assert.Equal(t, 1, status[0].Batch)
	})

	t.Run("验证表创建", func(t *testing.T) {
		conn, err := db.DB("test_migration")
		require.NoError(t, err)

		// 尝试插入数据验证表存在
		_, err = conn.Exec(ctx, "INSERT INTO test_table (name) VALUES (?)", "测试数据")
		assert.NoError(t, err)

		// 查询数据
		row := conn.QueryRow(ctx, "SELECT name FROM test_table WHERE id = 1")
		var name string
		err = row.Scan(&name)
		assert.NoError(t, err)
		assert.Equal(t, "测试数据", name)
	})

	t.Run("回滚迁移", func(t *testing.T) {
		err := migrator.Down(ctx, 1)
		assert.NoError(t, err)

		// 检查状态
		status, err := migrator.Status(ctx)
		assert.NoError(t, err)
		assert.Len(t, status, 1)
		assert.False(t, status[0].Applied)
	})

	t.Run("验证表删除", func(t *testing.T) {
		conn, err := db.DB("test_migration")
		require.NoError(t, err)

		// 尝试查询表，应该失败
		_, err = conn.Exec(ctx, "SELECT * FROM test_table")
		assert.Error(t, err)
	})
}

func TestMultipleMigrations(t *testing.T) {
	ctx, migrator := setupMigrationTest(t)
	defer cleanupMigrationTest(t)

	// 注册多个迁移
	migrator.RegisterFunc(
		"20240101_001",
		"创建用户表",
		func(ctx context.Context, conn db.ConnectionInterface) error {
			schema := migration.NewSchemaBuilder(conn)
			table := &migration.Table{
				Name: "users",
				Columns: []*migration.Column{
					{
						Name:          "id",
						Type:          migration.ColumnTypeBigInt,
						PrimaryKey:    true,
						AutoIncrement: true,
						NotNull:       true,
					},
					{
						Name:    "name",
						Type:    migration.ColumnTypeVarchar,
						Length:  100,
						NotNull: true,
					},
				},
			}
			return schema.CreateTable(ctx, table)
		},
		func(ctx context.Context, conn db.ConnectionInterface) error {
			schema := migration.NewSchemaBuilder(conn)
			return schema.DropTable(ctx, "users")
		},
	)

	migrator.RegisterFunc(
		"20240101_002",
		"添加邮箱字段",
		func(ctx context.Context, conn db.ConnectionInterface) error {
			schema := migration.NewSchemaBuilder(conn)
			column := &migration.Column{
				Name:   "email",
				Type:   migration.ColumnTypeVarchar,
				Length: 100,
			}
			return schema.AddColumn(ctx, "users", column)
		},
		func(ctx context.Context, conn db.ConnectionInterface) error {
			schema := migration.NewSchemaBuilder(conn)
			return schema.DropColumn(ctx, "users", "email")
		},
	)

	migrator.RegisterFunc(
		"20240101_003",
		"创建文章表",
		func(ctx context.Context, conn db.ConnectionInterface) error {
			schema := migration.NewSchemaBuilder(conn)
			table := &migration.Table{
				Name: "posts",
				Columns: []*migration.Column{
					{
						Name:          "id",
						Type:          migration.ColumnTypeBigInt,
						PrimaryKey:    true,
						AutoIncrement: true,
						NotNull:       true,
					},
					{
						Name:    "title",
						Type:    migration.ColumnTypeVarchar,
						Length:  200,
						NotNull: true,
					},
					{
						Name:    "user_id",
						Type:    migration.ColumnTypeBigInt,
						NotNull: true,
					},
				},
			}
			return schema.CreateTable(ctx, table)
		},
		func(ctx context.Context, conn db.ConnectionInterface) error {
			schema := migration.NewSchemaBuilder(conn)
			return schema.DropTable(ctx, "posts")
		},
	)

	t.Run("执行所有迁移", func(t *testing.T) {
		err := migrator.Up(ctx)
		assert.NoError(t, err)

		// 检查所有迁移都已应用
		status, err := migrator.Status(ctx)
		assert.NoError(t, err)
		assert.Len(t, status, 3)

		for _, s := range status {
			assert.True(t, s.Applied)
			assert.Equal(t, 1, s.Batch) // 所有迁移应该在同一个批次
		}
	})

	t.Run("验证表结构", func(t *testing.T) {
		conn, err := db.DB("test_migration")
		require.NoError(t, err)

		// 验证users表
		_, err = conn.Exec(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "测试用户", "test@example.com")
		assert.NoError(t, err)

		// 验证posts表
		_, err = conn.Exec(ctx, "INSERT INTO posts (title, user_id) VALUES (?, ?)", "测试文章", 1)
		assert.NoError(t, err)

		// 验证数据关联
		row := conn.QueryRow(ctx, `
			SELECT u.name, p.title 
			FROM users u 
			JOIN posts p ON u.id = p.user_id 
			WHERE u.id = 1
		`)
		var userName, postTitle string
		err = row.Scan(&userName, &postTitle)
		assert.NoError(t, err)
		assert.Equal(t, "测试用户", userName)
		assert.Equal(t, "测试文章", postTitle)
	})

	t.Run("部分回滚", func(t *testing.T) {
		// 回滚最后一个迁移
		err := migrator.Down(ctx, 1)
		assert.NoError(t, err)

		status, err := migrator.Status(ctx)
		assert.NoError(t, err)

		// 检查前两个迁移仍然应用，最后一个被回滚
		assert.True(t, status[0].Applied)  // users表
		assert.True(t, status[1].Applied)  // email字段
		assert.False(t, status[2].Applied) // posts表

		// 验证posts表不存在
		conn, err := db.DB("test_migration")
		require.NoError(t, err)
		_, err = conn.Exec(ctx, "SELECT * FROM posts")
		assert.Error(t, err)
	})

	t.Run("重新应用迁移", func(t *testing.T) {
		err := migrator.Up(ctx)
		assert.NoError(t, err)

		status, err := migrator.Status(ctx)
		assert.NoError(t, err)

		// 检查posts表迁移在新的批次中
		assert.Equal(t, 1, status[0].Batch) // users表
		assert.Equal(t, 1, status[1].Batch) // email字段
		assert.Equal(t, 2, status[2].Batch) // posts表（新批次）
	})

	t.Run("完全重置", func(t *testing.T) {
		err := migrator.Reset(ctx)
		assert.NoError(t, err)

		status, err := migrator.Status(ctx)
		assert.NoError(t, err)

		// 所有迁移都应该被回滚
		for _, s := range status {
			assert.False(t, s.Applied)
		}
	})
}

func TestSchemaBuilder(t *testing.T) {
	ctx, _ := setupMigrationTest(t)
	defer cleanupMigrationTest(t)

	conn, err := db.DB("test_migration")
	require.NoError(t, err)

	schema := migration.NewSchemaBuilder(conn)

	t.Run("创建复杂表结构", func(t *testing.T) {
		table := &migration.Table{
			Name: "complex_table",
			Columns: []*migration.Column{
				{
					Name:          "id",
					Type:          migration.ColumnTypeBigInt,
					PrimaryKey:    true,
					AutoIncrement: true,
					NotNull:       true,
				},
				{
					Name:    "title",
					Type:    migration.ColumnTypeVarchar,
					Length:  200,
					NotNull: true,
				},
				{
					Name: "content",
					Type: migration.ColumnTypeText,
				},
				{
					Name:      "price",
					Type:      migration.ColumnTypeDecimal,
					Precision: 10,
					Scale:     2,
					Default:   0.00,
				},
				{
					Name:    "is_active",
					Type:    migration.ColumnTypeBoolean,
					Default: true,
				},
				{
					Name:    "created_at",
					Type:    migration.ColumnTypeDateTime,
					Default: "CURRENT_TIMESTAMP",
				},
			},
			Indexes: []*migration.Index{
				{
					Name:    "idx_title",
					Columns: []string{"title"},
				},
				{
					Name:    "idx_price_active",
					Columns: []string{"price", "is_active"},
				},
			},
		}

		err := schema.CreateTable(ctx, table)
		assert.NoError(t, err)

		// 验证表创建成功
		_, err = conn.Exec(ctx, `
			INSERT INTO complex_table (title, content, price, is_active) 
			VALUES (?, ?, ?, ?)
		`, "测试标题", "测试内容", 99.99, true)
		assert.NoError(t, err)

		// 验证数据
		row := conn.QueryRow(ctx, "SELECT title, price FROM complex_table WHERE id = 1")
		var title string
		var price float64
		err = row.Scan(&title, &price)
		assert.NoError(t, err)
		assert.Equal(t, "测试标题", title)
		assert.Equal(t, 99.99, price)
	})

	t.Run("修改表结构", func(t *testing.T) {
		// 添加新列
		newColumn := &migration.Column{
			Name: "description",
			Type: migration.ColumnTypeText,
		}

		err := schema.AddColumn(ctx, "complex_table", newColumn)
		assert.NoError(t, err)

		// 验证新列
		_, err = conn.Exec(ctx, "UPDATE complex_table SET description = ? WHERE id = 1", "测试描述")
		assert.NoError(t, err)

		// 创建新索引
		newIndex := &migration.Index{
			Name:    "idx_description",
			Columns: []string{"description"},
		}

		err = schema.CreateIndex(ctx, "complex_table", newIndex)
		assert.NoError(t, err)
	})

	t.Run("清理表结构", func(t *testing.T) {
		// 删除索引
		err := schema.DropIndex(ctx, "complex_table", "idx_description")
		assert.NoError(t, err)

		// 删除列
		err = schema.DropColumn(ctx, "complex_table", "description")
		assert.NoError(t, err)

		// 删除表
		err = schema.DropTable(ctx, "complex_table")
		assert.NoError(t, err)

		// 验证表已删除
		_, err = conn.Exec(ctx, "SELECT * FROM complex_table")
		assert.Error(t, err)
	})
}

func TestMigrationErrorHandling(t *testing.T) {
	ctx, migrator := setupMigrationTest(t)
	defer cleanupMigrationTest(t)

	// 注册会失败的迁移
	migrator.RegisterFunc(
		"20240101_001",
		"失败的迁移",
		func(ctx context.Context, conn db.ConnectionInterface) error {
			_, err := conn.Exec(ctx, "INVALID SQL STATEMENT")
			return err
		},
		func(ctx context.Context, conn db.ConnectionInterface) error {
			return nil
		},
	)

	t.Run("迁移失败处理", func(t *testing.T) {
		err := migrator.Up(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "migration 20240101_001 failed")

		// 验证没有记录失败的迁移
		status, err := migrator.Status(ctx)
		assert.NoError(t, err)
		assert.Len(t, status, 1)
		assert.False(t, status[0].Applied)
	})

	t.Run("回滚不存在的迁移", func(t *testing.T) {
		err := migrator.Down(ctx, 1)
		assert.NoError(t, err) // 应该不报错，因为没有迁移可回滚
	})
}

func TestMigrationTableCustomization(t *testing.T) {
	ctx, migrator := setupMigrationTest(t)
	defer cleanupMigrationTest(t)

	// 设置自定义迁移表名
	migrator.SetTableName("custom_migrations")

	// 注册简单迁移
	migrator.RegisterFunc(
		"20240101_001",
		"测试自定义表名",
		func(ctx context.Context, conn db.ConnectionInterface) error {
			_, err := conn.Exec(ctx, "CREATE TABLE test_custom (id INTEGER PRIMARY KEY)")
			return err
		},
		func(ctx context.Context, conn db.ConnectionInterface) error {
			_, err := conn.Exec(ctx, "DROP TABLE test_custom")
			return err
		},
	)

	t.Run("使用自定义迁移表", func(t *testing.T) {
		err := migrator.Up(ctx)
		assert.NoError(t, err)

		// 验证使用了自定义表名
		conn, err := db.DB("test_migration")
		require.NoError(t, err)

		row := conn.QueryRow(ctx, "SELECT COUNT(*) FROM custom_migrations")
		var count int
		err = row.Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)

		// 验证默认表名不存在（但不一定会失败，因为可能之前的测试创建过）
		// 这里我们只验证自定义表名被使用，不验证默认表名是否存在
	})
}
