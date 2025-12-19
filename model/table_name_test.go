package model

import (
	"testing"
)

// TestTableNameConfig 测试表名配置
func TestTableNameConfig(t *testing.T) {
	// 保存原始配置
	originalConfig := GetTableNameConfig()
	defer SetTableNameConfig(originalConfig)

	// 测试设置前缀
	SetTablePrefix("torm_")
	config := GetTableNameConfig()
	if config.Prefix != "torm_" {
		t.Errorf("Expected prefix 'torm_', got '%s'", config.Prefix)
	}

	// 测试设置后缀
	SetTableSuffix("_table")
	config = GetTableNameConfig()
	if config.Suffix != "_table" {
		t.Errorf("Expected suffix '_table', got '%s'", config.Suffix)
	}

	// 测试设置单数表名
	SetSingularTable(false)
	config = GetTableNameConfig()
	if config.SingularTable != false {
		t.Errorf("Expected SingularTable false, got %v", config.SingularTable)
	}

	// 测试完整配置
	newConfig := TableNameConfig{
		Prefix:        "app_",
		Suffix:        "_v2",
		SingularTable: true,
	}
	SetTableNameConfig(newConfig)
	config = GetTableNameConfig()
	if config.Prefix != "app_" || config.Suffix != "_v2" || config.SingularTable != true {
		t.Errorf("Config not set correctly: %+v", config)
	}
}

// TestTableNameBuilder 测试表名构建器
func TestTableNameBuilder(t *testing.T) {
	// 测试基本构建
	builder := NewTableNameBuilder("users")
	if builder.Build() != "users" {
		t.Errorf("Expected 'users', got '%s'", builder.Build())
	}

	// 测试带前缀
	builder = NewTableNameBuilder("users").WithPrefix("app_")
	if builder.Build() != "app_users" {
		t.Errorf("Expected 'app_users', got '%s'", builder.Build())
	}

	// 测试带后缀
	builder = NewTableNameBuilder("users").WithSuffix("_v2")
	if builder.Build() != "users_v2" {
		t.Errorf("Expected 'users_v2', got '%s'", builder.Build())
	}

	// 测试带schema
	builder = NewTableNameBuilder("users").WithSchema("public")
	if builder.Build() != "public.users" {
		t.Errorf("Expected 'public.users', got '%s'", builder.Build())
	}

	// 测试链式调用
	builder = NewTableNameBuilder("users").
		WithPrefix("app_").
		WithSuffix("_v2").
		WithSchema("public")
	expected := "public.app_users_v2"
	if builder.Build() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, builder.Build())
	}

	// 测试String方法
	if builder.String() != expected {
		t.Errorf("Expected String() to return '%s', got '%s'", expected, builder.String())
	}
}

// TestResolveTableName 测试表名解析
func TestResolveTableName(t *testing.T) {
	// 保存原始配置
	originalConfig := GetTableNameConfig()
	defer SetTableNameConfig(originalConfig)

	// 重置配置
	SetTableNameConfig(TableNameConfig{
		Prefix:        "",
		Suffix:        "",
		SingularTable: true,
	})

	// 测试从实现Tabler接口的模型获取表名
	user := &TestUser{}
	tableName := ResolveTableName(user)
	if tableName != "test_users" {
		t.Errorf("Expected 'test_users', got '%s'", tableName)
	}

	// 测试从没有实现Tabler接口的模型推断表名
	type SimpleModel struct {
		ID   int
		Name string
	}
	simple := &SimpleModel{}
	tableName = ResolveTableName(simple)
	if tableName != "simple_model" {
		t.Errorf("Expected 'simple_model', got '%s'", tableName)
	}
}

// TestTableNameWithPrefix 测试带前缀的表名
func TestTableNameWithPrefix(t *testing.T) {
	user := &TestUser{}
	tableName := TableNameWithPrefix(user, "app_")
	if tableName != "app_test_users" {
		t.Errorf("Expected 'app_test_users', got '%s'", tableName)
	}

	// 测试已有前缀的情况
	tableName = TableNameWithPrefix(user, "test_")
	if tableName != "test_users" {
		t.Errorf("Expected 'test_users' (no duplicate prefix), got '%s'", tableName)
	}
}

// TestTableNameWithSuffix 测试带后缀的表名
func TestTableNameWithSuffix(t *testing.T) {
	user := &TestUser{}
	tableName := TableNameWithSuffix(user, "_v2")
	if tableName != "test_users_v2" {
		t.Errorf("Expected 'test_users_v2', got '%s'", tableName)
	}

	// 测试已有后缀的情况
	tableName = TableNameWithSuffix(user, "_users")
	if tableName != "test_users" {
		t.Errorf("Expected 'test_users' (no duplicate suffix), got '%s'", tableName)
	}
}

// TestTableNameWithSchema 测试带schema的表名
func TestTableNameWithSchema(t *testing.T) {
	user := &TestUser{}
	tableName := TableNameWithSchema(user, "public")
	if tableName != "public.test_users" {
		t.Errorf("Expected 'public.test_users', got '%s'", tableName)
	}

	// 测试空schema
	tableName = TableNameWithSchema(user, "")
	if tableName != "test_users" {
		t.Errorf("Expected 'test_users', got '%s'", tableName)
	}
}

// TestTableNameCache 测试表名缓存
func TestTableNameCache(t *testing.T) {
	// 清除缓存
	ClearTableNameCache()

	user := &TestUser{}

	// 第一次获取（会缓存）
	tableName1 := GetCachedTableName(user)
	if tableName1 != "test_users" {
		t.Errorf("Expected 'test_users', got '%s'", tableName1)
	}

	// 第二次获取（从缓存）
	tableName2 := GetCachedTableName(user)
	if tableName2 != tableName1 {
		t.Errorf("Expected cached table name '%s', got '%s'", tableName1, tableName2)
	}

	// 清除缓存后重新获取
	ClearTableNameCache()
	tableName3 := GetCachedTableName(user)
	if tableName3 != "test_users" {
		t.Errorf("Expected 'test_users' after cache clear, got '%s'", tableName3)
	}
}

// TestCustomTableNameResolver 测试自定义表名解析器
func TestCustomTableNameResolver(t *testing.T) {
	// 保存原始解析器
	originalResolver := GetTableNameResolver()
	defer SetTableNameResolver(originalResolver)

	// 设置自定义解析器
	customResolver := &testTableNameResolver{prefix: "custom_"}
	SetTableNameResolver(customResolver)

	user := &TestUser{}
	tableName := ResolveTableNameWithResolver(user)
	if tableName != "custom_test_users" {
		t.Errorf("Expected 'custom_test_users', got '%s'", tableName)
	}

	// 清除解析器
	ClearTableNameResolver()
	tableName = ResolveTableNameWithResolver(user)
	if tableName != "test_users" {
		t.Errorf("Expected 'test_users' after clearing resolver, got '%s'", tableName)
	}
}

// testTableNameResolver 测试用的表名解析器
type testTableNameResolver struct {
	prefix string
}

func (r *testTableNameResolver) ResolveTableName(model interface{}) string {
	return r.prefix + ResolveTableName(model)
}

// TestBaseModelTableMethods 测试BaseModel的表名方法
func TestBaseModelTableMethods(t *testing.T) {
	model := NewModel("users")

	// 测试SetTable
	model.SetTable("custom_users")
	if model.GetTableName() != "custom_users" {
		t.Errorf("Expected 'custom_users', got '%s'", model.GetTableName())
	}

	// 测试Table（GORM兼容方法）
	model.Table("another_users")
	if model.GetTableName() != "another_users" {
		t.Errorf("Expected 'another_users', got '%s'", model.GetTableName())
	}

	// 测试WithTableName
	model.WithTableName("final_users")
	if model.GetTableName() != "final_users" {
		t.Errorf("Expected 'final_users', got '%s'", model.GetTableName())
	}

	// 测试SetTableName（DynamicTableNamer接口）
	model.SetTableName("dynamic_users")
	if model.GetTableName() != "dynamic_users" {
		t.Errorf("Expected 'dynamic_users', got '%s'", model.GetTableName())
	}

	// 测试TableName方法
	if model.TableName() != "dynamic_users" {
		t.Errorf("Expected TableName() to return 'dynamic_users', got '%s'", model.TableName())
	}
}

// TestBaseModelTablePrefix 测试BaseModel的表名前缀方法
func TestBaseModelTablePrefix(t *testing.T) {
	model := NewModel("users")

	// 测试SetTablePrefix
	model.SetTablePrefix("app_")
	if model.GetTableName() != "app_users" {
		t.Errorf("Expected 'app_users', got '%s'", model.GetTableName())
	}

	// 测试不重复添加前缀
	model.SetTablePrefix("app_")
	if model.GetTableName() != "app_users" {
		t.Errorf("Expected 'app_users' (no duplicate), got '%s'", model.GetTableName())
	}
}

// TestBaseModelTableSuffix 测试BaseModel的表名后缀方法
func TestBaseModelTableSuffix(t *testing.T) {
	model := NewModel("users")

	// 测试SetTableSuffix
	model.SetTableSuffix("_v2")
	if model.GetTableName() != "users_v2" {
		t.Errorf("Expected 'users_v2', got '%s'", model.GetTableName())
	}

	// 测试不重复添加后缀
	model.SetTableSuffix("_v2")
	if model.GetTableName() != "users_v2" {
		t.Errorf("Expected 'users_v2' (no duplicate), got '%s'", model.GetTableName())
	}
}

// TestShardingConfig 测试分表配置
func TestShardingConfig(t *testing.T) {
	config := DefaultShardingConfig()

	if config.Strategy != ShardingByMod {
		t.Errorf("Expected default strategy ShardingByMod, got %v", config.Strategy)
	}
	if config.ShardCount != 10 {
		t.Errorf("Expected default shard count 10, got %d", config.ShardCount)
	}
	if config.DateFormat != "200601" {
		t.Errorf("Expected default date format '200601', got '%s'", config.DateFormat)
	}
	if config.RangeSize != 1000000 {
		t.Errorf("Expected default range size 1000000, got %d", config.RangeSize)
	}
}

// TestGetShardingTableName 测试分表表名获取
func TestGetShardingTableName(t *testing.T) {
	// 测试取模分表
	config := ShardingConfig{
		Strategy:   ShardingByMod,
		ShardCount: 10,
	}
	tableName := GetShardingTableName("users", 123, config)
	if tableName != "users_3" {
		t.Errorf("Expected 'users_3' (123 %% 10 = 3), got '%s'", tableName)
	}

	// 测试哈希分表
	config.Strategy = ShardingByHash
	tableName = GetShardingTableName("users", "test_key", config)
	// 哈希值是确定的，但我们只验证格式
	if tableName == "" || tableName == "users" {
		t.Errorf("Expected sharded table name, got '%s'", tableName)
	}

	// 测试范围分表
	config.Strategy = ShardingByRange
	config.RangeSize = 1000
	tableName = GetShardingTableName("users", int64(2500), config)
	if tableName != "users_2" {
		t.Errorf("Expected 'users_2' (2500 / 1000 = 2), got '%s'", tableName)
	}

	// 测试带前缀和后缀
	config.Strategy = ShardingByMod
	config.TablePrefix = "app_"
	config.TableSuffix = "_v2"
	tableName = GetShardingTableName("users", 5, config)
	if tableName != "app_users_5_v2" {
		t.Errorf("Expected 'app_users_5_v2', got '%s'", tableName)
	}
}

// TestDynamicTableNamerInterface 测试DynamicTableNamer接口
func TestDynamicTableNamerInterface(t *testing.T) {
	model := NewModel("users")

	// 验证BaseModel实现了DynamicTableNamer接口
	var namer DynamicTableNamer = model

	// 测试TableName
	if namer.TableName() != "users" {
		t.Errorf("Expected 'users', got '%s'", namer.TableName())
	}

	// 测试SetTableName
	namer.SetTableName("dynamic_users")
	if namer.TableName() != "dynamic_users" {
		t.Errorf("Expected 'dynamic_users', got '%s'", namer.TableName())
	}
}

// TestTableNameFromStruct 测试从结构体获取表名
func TestTableNameFromStruct(t *testing.T) {
	// 保存原始配置
	originalConfig := GetTableNameConfig()
	defer SetTableNameConfig(originalConfig)

	// 重置配置
	SetTableNameConfig(TableNameConfig{
		Prefix:        "",
		Suffix:        "",
		SingularTable: true,
	})

	user := &TestUser{}
	tableName := TableNameFromStruct(user)
	if tableName != "test_users" {
		t.Errorf("Expected 'test_users', got '%s'", tableName)
	}
}

// BenchmarkResolveTableName 基准测试表名解析
func BenchmarkResolveTableName(b *testing.B) {
	user := &TestUser{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ResolveTableName(user)
	}
}

// BenchmarkGetCachedTableName 基准测试缓存表名获取
func BenchmarkGetCachedTableName(b *testing.B) {
	user := &TestUser{}
	// 预热缓存
	GetCachedTableName(user)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetCachedTableName(user)
	}
}

// BenchmarkTableNameBuilder 基准测试表名构建器
func BenchmarkTableNameBuilder(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewTableNameBuilder("users").
			WithPrefix("app_").
			WithSuffix("_v2").
			WithSchema("public").
			Build()
	}
}
