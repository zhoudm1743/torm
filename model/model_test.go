package model

import (
	"strings"
	"testing"
	"time"
)

// TestUser 测试用户模型
type TestUser struct {
	BaseModel
	ID       int    `json:"id" torm:"primary_key,auto_increment"`
	Username string `json:"username" torm:"type:varchar,size:50"`
	Email    string `json:"email" torm:"type:varchar,size:100"`
	Age      int    `json:"age" torm:"type:int,default:0"`
}

func (u *TestUser) GetTableName() string {
	return "test_users"
}

// TestProfile 测试用户资料模型
type TestProfile struct {
	BaseModel
	ID     int    `json:"id" torm:"primary_key,auto_increment"`
	UserID int    `json:"user_id" torm:"type:int"`
	Bio    string `json:"bio" torm:"type:text"`
}

func (p *TestProfile) GetTableName() string {
	return "test_profiles"
}

// TestModelWithTags 带有torm标签的测试模型
type TestModelWithTags struct {
	BaseModel
	ID        int        `json:"id" torm:"primary_key,auto_increment"`
	Name      string     `json:"name" torm:"type:varchar,size:100"`
	CreatedAt time.Time  `json:"created_at" torm:"auto_create_time"`
	UpdatedAt time.Time  `json:"updated_at" torm:"auto_update_time"`
	DeletedAt *time.Time `json:"deleted_at" torm:"soft_delete"`
}

func (t *TestModelWithTags) GetTableName() string {
	return "test_models_with_tags"
}

func TestNewModel(t *testing.T) {
	// 测试用表名创建模型
	model1 := NewModel("users")
	if model1.GetTableName() != "users" {
		t.Errorf("Expected table name 'users', got '%s'", model1.GetTableName())
	}
	if model1.GetPrimaryKey() != "id" {
		t.Errorf("Expected primary key 'id', got '%s'", model1.GetPrimaryKey())
	}
	if model1.GetConnection() != "default" {
		t.Errorf("Expected connection 'default', got '%s'", model1.GetConnection())
	}

	// 测试用配置创建模型
	config := ModelConfig{
		TableName:  "custom_table",
		PrimaryKey: "custom_id",
		Connection: "custom_conn",
	}
	model2 := NewModel(config)
	if model2.GetTableName() != "custom_table" {
		t.Errorf("Expected table name 'custom_table', got '%s'", model2.GetTableName())
	}
	if model2.GetPrimaryKey() != "custom_id" {
		t.Errorf("Expected primary key 'custom_id', got '%s'", model2.GetPrimaryKey())
	}
	if model2.GetConnection() != "custom_conn" {
		t.Errorf("Expected connection 'custom_conn', got '%s'", model2.GetConnection())
	}

	// 测试用结构体创建模型
	user := &TestUser{}
	model3 := NewModel(user)
	if model3.GetTableName() != "test_users" {
		t.Errorf("Expected table name 'test_users', got '%s'", model3.GetTableName())
	}
}

func TestModelConfiguration(t *testing.T) {
	model := NewModel()

	// 测试链式配置
	model.SetTable("users").
		SetPrimaryKey("user_id").
		SetConnection("test_conn").
		EnableTimestamps().
		SetCreatedAtField("create_time").
		SetUpdatedAtField("update_time").
		EnableSoftDeletes().
		SetDeletedAtField("delete_time")

	if model.GetTableName() != "users" {
		t.Errorf("Expected table name 'users', got '%s'", model.GetTableName())
	}
	if model.GetPrimaryKey() != "user_id" {
		t.Errorf("Expected primary key 'user_id', got '%s'", model.GetPrimaryKey())
	}
	if model.GetConnection() != "test_conn" {
		t.Errorf("Expected connection 'test_conn', got '%s'", model.GetConnection())
	}
	if model.GetCreatedAtField() != "create_time" {
		t.Errorf("Expected created_at field 'create_time', got '%s'", model.GetCreatedAtField())
	}
	if model.GetUpdatedAtField() != "update_time" {
		t.Errorf("Expected updated_at field 'update_time', got '%s'", model.GetUpdatedAtField())
	}
}

func TestModelAttributes(t *testing.T) {
	model := NewModel("users")

	// 测试设置单个属性
	model.SetAttribute("name", "John")
	if model.GetAttribute("name") != "John" {
		t.Errorf("Expected attribute 'name' to be 'John', got '%v'", model.GetAttribute("name"))
	}

	// 测试批量设置属性
	attrs := map[string]interface{}{
		"email": "john@example.com",
		"age":   30,
	}
	model.SetAttributes(attrs)

	if model.GetAttribute("email") != "john@example.com" {
		t.Errorf("Expected attribute 'email' to be 'john@example.com', got '%v'", model.GetAttribute("email"))
	}
	if model.GetAttribute("age") != 30 {
		t.Errorf("Expected attribute 'age' to be 30, got '%v'", model.GetAttribute("age"))
	}

	// 测试获取所有属性
	allAttrs := model.GetAttributes()
	if len(allAttrs) != 3 {
		t.Errorf("Expected 3 attributes, got %d", len(allAttrs))
	}

	// 测试填充
	newData := map[string]interface{}{
		"status": "active",
		"score":  95.5,
	}
	model.Fill(newData)

	if model.GetAttribute("status") != "active" {
		t.Errorf("Expected attribute 'status' to be 'active', got '%v'", model.GetAttribute("status"))
	}

	// 测试清空属性
	model.ClearAttributes()
	if len(model.GetAttributes()) != 0 {
		t.Errorf("Expected 0 attributes after clear, got %d", len(model.GetAttributes()))
	}
}

func TestModelState(t *testing.T) {
	model := NewModel("users")

	// 测试初始状态
	if !model.IsNew() {
		t.Error("Expected model to be new initially")
	}
	if model.IsExists() {
		t.Error("Expected model to not exist initially")
	}

	// 测试标记为存在
	model.MarkAsExists()
	if model.IsNew() {
		t.Error("Expected model to not be new after marking as exists")
	}
	if !model.IsExists() {
		t.Error("Expected model to exist after marking as exists")
	}

	// 测试标记为新记录
	model.MarkAsNew()
	if !model.IsNew() {
		t.Error("Expected model to be new after marking as new")
	}
	if model.IsExists() {
		t.Error("Expected model to not exist after marking as new")
	}
}

func TestModelKey(t *testing.T) {
	model := NewModel("users")

	// 测试设置和获取主键值
	model.SetKey(123)
	if model.GetKey() != 123 {
		t.Errorf("Expected key to be 123, got '%v'", model.GetKey())
	}

	// 测试使用自定义主键
	model.SetPrimaryKey("user_id")
	model.SetKey("abc123")
	if model.GetKey() != "abc123" {
		t.Errorf("Expected key to be 'abc123', got '%v'", model.GetKey())
	}
}

func TestModelSerialization(t *testing.T) {
	model := NewModel("users")
	model.SetAttributes(map[string]interface{}{
		"name":  "John",
		"email": "john@example.com",
		"age":   30,
	})

	// 测试转换为 Map
	data := model.ToMap()
	if len(data) != 3 {
		t.Errorf("Expected 3 items in map, got %d", len(data))
	}

	// 测试转换为 JSON
	jsonStr, err := model.ToJSON()
	if err != nil {
		t.Errorf("Failed to convert to JSON: %v", err)
	}
	if jsonStr == "" {
		t.Error("Expected non-empty JSON string")
	}

	// 测试从 JSON 创建
	newModel := NewModel("users")
	err = newModel.FromJSON(jsonStr)
	if err != nil {
		t.Errorf("Failed to parse from JSON: %v", err)
	}
	if newModel.GetAttribute("name") != "John" {
		t.Errorf("Expected name 'John' after JSON parse, got '%v'", newModel.GetAttribute("name"))
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"UserModel", "user_model"},
		{"HTTPClient", "http_client"},
		{"XMLParser", "xml_parser"},
		{"ID", "id"},
		{"HTML", "html"},
		{"simple", "simple"},
		{"", ""},
	}

	for _, test := range tests {
		result := toSnakeCase(test.input)
		if result != test.expected {
			t.Errorf("toSnakeCase(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestRelationCreation(t *testing.T) {
	user := NewModel("users")
	profile := NewModel("profiles")

	// 测试创建 HasOne 关联
	hasOne := user.HasOne(&TestProfile{}, "user_id", "id")
	if hasOne == nil {
		t.Error("Expected HasOne relation to be created")
	}

	// 测试创建 HasMany 关联
	hasMany := user.HasMany(&TestProfile{}, "user_id", "id")
	if hasMany == nil {
		t.Error("Expected HasMany relation to be created")
	}

	// 测试创建 BelongsTo 关联
	belongsTo := profile.BelongsTo(&TestUser{}, "user_id", "id")
	if belongsTo == nil {
		t.Error("Expected BelongsTo relation to be created")
	}

	// 测试创建 BelongsToMany 关联
	belongsToMany := user.BelongsToMany(&TestProfile{}, "user_profiles", "user_id", "profile_id")
	if belongsToMany == nil {
		t.Error("Expected BelongsToMany relation to be created")
	}
}

func TestDefaultModelConfig(t *testing.T) {
	config := DefaultModelConfig()

	if config.PrimaryKey != "id" {
		t.Errorf("Expected default primary key 'id', got '%s'", config.PrimaryKey)
	}
	if config.Connection != "default" {
		t.Errorf("Expected default connection 'default', got '%s'", config.Connection)
	}
	if !config.Timestamps {
		t.Error("Expected timestamps to be enabled by default")
	}
	if config.CreatedAtCol != "created_at" {
		t.Errorf("Expected default created_at column 'created_at', got '%s'", config.CreatedAtCol)
	}
	if config.UpdatedAtCol != "updated_at" {
		t.Errorf("Expected default updated_at column 'updated_at', got '%s'", config.UpdatedAtCol)
	}
	if config.SoftDeletes {
		t.Error("Expected soft deletes to be disabled by default")
	}
	if config.DeletedAtCol != "deleted_at" {
		t.Errorf("Expected default deleted_at column 'deleted_at', got '%s'", config.DeletedAtCol)
	}
}

// 基准测试
func BenchmarkNewModel(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewModel("users")
	}
}

func BenchmarkSetAttribute(b *testing.B) {
	model := NewModel("users")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.SetAttribute("name", "John")
	}
}

func BenchmarkGetAttribute(b *testing.B) {
	model := NewModel("users")
	model.SetAttribute("name", "John")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.GetAttribute("name")
	}
}

// TestParameterizedQueries 测试参数式查询兼容性
func TestParameterizedQueries(t *testing.T) {
	model := NewModel("users")

	// 测试三参数格式: Where("name", "=", "John")
	query1, err := model.Where("name", "=", "John")
	if err != nil {
		t.Errorf("Failed to create query with 3 parameters: %v", err)
	}
	if query1 == nil {
		t.Error("Expected query to be created, got nil")
	}

	// 测试SQL+参数格式: Where("name = ?", "John")
	query2, err := model.Where("name = ?", "John")
	if err != nil {
		t.Errorf("Failed to create query with SQL+parameter: %v", err)
	}
	if query2 == nil {
		t.Error("Expected query to be created, got nil")
	}

	// 测试SQL+数组参数格式: Where("id IN (?)", []int{1,2,3})
	query3, err := model.Where("id IN (?)", []int{1, 2, 3})
	if err != nil {
		t.Errorf("Failed to create query with SQL+array parameter: %v", err)
	}
	if query3 == nil {
		t.Error("Expected query to be created, got nil")
	}

	// 测试纯SQL格式: Where("name = 'John'")
	query4, err := model.Where("name = 'John'")
	if err != nil {
		t.Errorf("Failed to create query with pure SQL: %v", err)
	}
	if query4 == nil {
		t.Error("Expected query to be created, got nil")
	}

	// 测试WhereIn
	query5, err := model.WhereIn("status", []interface{}{"active", "pending"})
	if err != nil {
		t.Errorf("Failed to create WhereIn query: %v", err)
	}
	if query5 == nil {
		t.Error("Expected WhereIn query to be created, got nil")
	}

	// 测试WhereNull
	query6, err := model.WhereNull("deleted_at")
	if err != nil {
		t.Errorf("Failed to create WhereNull query: %v", err)
	}
	if query6 == nil {
		t.Error("Expected WhereNull query to be created, got nil")
	}

	// 测试WhereRaw
	query7, err := model.WhereRaw("age > ? AND status = ?", 18, "active")
	if err != nil {
		t.Errorf("Failed to create WhereRaw query: %v", err)
	}
	if query7 == nil {
		t.Error("Expected WhereRaw query to be created, got nil")
	}
}

// TestQueryBuilderMethods 测试查询构建器方法兼容性
func TestQueryBuilderMethods(t *testing.T) {
	model := NewModel("users")

	// 测试Select
	query1, err := model.Select("id", "name", "email")
	if err != nil {
		t.Errorf("Failed to create Select query: %v", err)
	}
	if query1 == nil {
		t.Error("Expected Select query to be created, got nil")
	}

	// 测试OrderBy
	query2, err := model.OrderBy("created_at", "DESC")
	if err != nil {
		t.Errorf("Failed to create OrderBy query: %v", err)
	}
	if query2 == nil {
		t.Error("Expected OrderBy query to be created, got nil")
	}

	// 测试Limit
	query3, err := model.Limit(10)
	if err != nil {
		t.Errorf("Failed to create Limit query: %v", err)
	}
	if query3 == nil {
		t.Error("Expected Limit query to be created, got nil")
	}

	// 测试Page
	query4, err := model.Page(1, 10)
	if err != nil {
		t.Errorf("Failed to create Page query: %v", err)
	}
	if query4 == nil {
		t.Error("Expected Page query to be created, got nil")
	}

	// 测试Join
	query5, err := model.Join("profiles", "users.id", "=", "profiles.user_id")
	if err != nil {
		t.Errorf("Failed to create Join query: %v", err)
	}
	if query5 == nil {
		t.Error("Expected Join query to be created, got nil")
	}
}

// TestChainedQueries 测试链式查询兼容性（通过QueryBuilder继续链式调用）
func TestChainedQueries(t *testing.T) {
	model := NewModel("users")

	// 测试从模型开始的链式查询
	query, err := model.Where("status", "=", "active")
	if err != nil {
		t.Errorf("Failed to create initial query: %v", err)
		return
	}

	// 继续使用QueryBuilder进行链式调用
	finalQuery := query.Where("age", ">", 18).OrderBy("created_at", "DESC").Limit(10)
	if finalQuery == nil {
		t.Error("Expected chained query to work, got nil")
	}

	// 测试复杂的链式查询
	complexQuery, err := model.Select("id", "name", "email")
	if err != nil {
		t.Errorf("Failed to create select query: %v", err)
		return
	}

	result := complexQuery.
		Where("status", "=", "active").
		Where("age", ">=", 18).
		WhereIn("role", []interface{}{"admin", "user"}).
		OrderBy("name", "ASC").
		Limit(50)

	if result == nil {
		t.Error("Expected complex chained query to work, got nil")
	}
}

// TestTormTagPriority 测试torm标签优先级
func TestTormTagPriority(t *testing.T) {
	// 创建带有torm标签的模型
	testModel := &TestModelWithTags{}
	model := NewModel(testModel)

	// 验证torm标签中的primary_key设置
	if model.GetPrimaryKey() != "id" {
		t.Errorf("Expected primary key 'id' from torm tag, got '%s'", model.GetPrimaryKey())
	}

	// 验证torm标签中的auto_create_time设置
	if model.GetCreatedAtField() != "created_at" {
		t.Errorf("Expected created_at field 'created_at' from torm tag, got '%s'", model.GetCreatedAtField())
	}

	// 验证torm标签中的auto_update_time设置
	if model.GetUpdatedAtField() != "updated_at" {
		t.Errorf("Expected updated_at field 'updated_at' from torm tag, got '%s'", model.GetUpdatedAtField())
	}

	// 验证torm标签中的soft_delete设置
	if !model.config.SoftDeletes {
		t.Error("Expected soft deletes to be enabled from torm tag")
	}
	if model.config.DeletedAtCol != "deleted_at" {
		t.Errorf("Expected deleted_at field 'deleted_at' from torm tag, got '%s'", model.config.DeletedAtCol)
	}
}

// TestTormTagOverridesConfig 测试torm标签覆盖ModelConfig
func TestTormTagOverridesConfig(t *testing.T) {
	// 创建一个与torm标签冲突的配置
	userConfig := ModelConfig{
		TableName:    "custom_table",
		PrimaryKey:   "custom_id", // 这个会被torm标签覆盖
		Connection:   "custom_connection",
		Timestamps:   false,         // 这个会被保留
		CreatedAtCol: "create_time", // 这个会被torm标签覆盖
		UpdatedAtCol: "update_time", // 这个会被torm标签覆盖
		SoftDeletes:  false,         // 这个会被torm标签覆盖
		DeletedAtCol: "delete_time", // 这个会被torm标签覆盖
	}

	testModel := &TestModelWithTags{}
	model := NewModel(testModel, userConfig)

	// 验证torm标签覆盖了配置
	if model.GetPrimaryKey() != "id" {
		t.Errorf("Expected torm tag to override primary key, got '%s'", model.GetPrimaryKey())
	}

	if model.GetCreatedAtField() != "created_at" {
		t.Errorf("Expected torm tag to override created_at field, got '%s'", model.GetCreatedAtField())
	}

	if model.GetUpdatedAtField() != "updated_at" {
		t.Errorf("Expected torm tag to override updated_at field, got '%s'", model.GetUpdatedAtField())
	}

	if !model.config.SoftDeletes {
		t.Error("Expected torm tag to override soft deletes setting")
	}

	if model.config.DeletedAtCol != "deleted_at" {
		t.Errorf("Expected torm tag to override deleted_at field, got '%s'", model.config.DeletedAtCol)
	}

	// 验证非冲突的配置被保留
	if model.GetTableName() != "test_models_with_tags" {
		t.Errorf("Expected table name from struct method, got '%s'", model.GetTableName())
	}

	if model.GetConnection() != "custom_connection" {
		t.Errorf("Expected custom connection to be preserved, got '%s'", model.GetConnection())
	}

	// Timestamps配置应该被保留，因为没有torm标签覆盖
	if model.config.Timestamps != false {
		t.Errorf("Expected timestamps config to be preserved, got '%v'", model.config.Timestamps)
	}
}

// TestConfigWithoutTags 测试没有torm标签时使用ModelConfig
func TestConfigWithoutTags(t *testing.T) {
	// 创建没有torm标签的简单结构体
	type SimpleModel struct {
		BaseModel
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	userConfig := ModelConfig{
		TableName:    "simple_models",
		PrimaryKey:   "custom_id",
		Connection:   "test_connection",
		Timestamps:   true,
		CreatedAtCol: "create_time",
		UpdatedAtCol: "update_time",
		SoftDeletes:  true,
		DeletedAtCol: "delete_time",
	}

	simpleModel := &SimpleModel{}
	model := NewModel(simpleModel, userConfig)

	// 验证ModelConfig的所有设置都被应用
	if model.GetTableName() != "simple_models" {
		t.Errorf("Expected table name 'simple_models', got '%s'", model.GetTableName())
	}

	if model.GetPrimaryKey() != "custom_id" {
		t.Errorf("Expected primary key 'custom_id', got '%s'", model.GetPrimaryKey())
	}

	if model.GetConnection() != "test_connection" {
		t.Errorf("Expected connection 'test_connection', got '%s'", model.GetConnection())
	}

	if model.GetCreatedAtField() != "create_time" {
		t.Errorf("Expected created_at field 'create_time', got '%s'", model.GetCreatedAtField())
	}

	if model.GetUpdatedAtField() != "update_time" {
		t.Errorf("Expected updated_at field 'update_time', got '%s'", model.GetUpdatedAtField())
	}

	if !model.config.SoftDeletes {
		t.Error("Expected soft deletes to be enabled")
	}

	if model.config.DeletedAtCol != "delete_time" {
		t.Errorf("Expected deleted_at field 'delete_time', got '%s'", model.config.DeletedAtCol)
	}
}

// TestAdvancedTormTags 测试高级torm标签支持
func TestAdvancedTormTags(t *testing.T) {
	// 创建带有各种torm标签的复杂模型
	type AdvancedModel struct {
		BaseModel
		ID        uint64     `json:"id" torm:"primary_key,auto_increment,type:bigint"`
		Username  string     `json:"username" torm:"type:varchar,size:50,unique,not_null"`
		Email     string     `json:"email" torm:"type:varchar,size:100,unique,index"`
		Age       int        `json:"age" torm:"type:int,default:0,unsigned"`
		Balance   float64    `json:"balance" torm:"type:decimal,precision:10,scale:2,default:0.00"`
		Status    string     `json:"status" torm:"type:varchar,size:20,default:'active',comment:'用户状态'"`
		Avatar    string     `json:"avatar" torm:"type:text,nullable"`
		Metadata  string     `json:"metadata" torm:"type:json"`
		CreatedAt time.Time  `json:"created_at" torm:"auto_create_time,timestamp"`
		UpdatedAt time.Time  `json:"updated_at" torm:"auto_update_time,timestamp"`
		DeletedAt *time.Time `json:"deleted_at" torm:"soft_delete,nullable"`
		ProfileID uint       `json:"profile_id" torm:"foreign_key:profiles.id,on_delete:cascade"`
		CompanyID uint       `json:"company_id" torm:"references:companies(id),on_update:restrict"`
	}

	advancedModel := &AdvancedModel{}
	model := NewModel(advancedModel)

	// 验证基础配置被正确解析
	if model.GetPrimaryKey() != "id" {
		t.Errorf("Expected primary key 'id', got '%s'", model.GetPrimaryKey())
	}

	if model.GetCreatedAtField() != "created_at" {
		t.Errorf("Expected created_at field 'created_at', got '%s'", model.GetCreatedAtField())
	}

	if model.GetUpdatedAtField() != "updated_at" {
		t.Errorf("Expected updated_at field 'updated_at', got '%s'", model.GetUpdatedAtField())
	}

	if !model.config.SoftDeletes {
		t.Error("Expected soft deletes to be enabled")
	}

	if model.config.DeletedAtCol != "deleted_at" {
		t.Errorf("Expected deleted_at field 'deleted_at', got '%s'", model.config.DeletedAtCol)
	}

	// 验证AutoMigrate可以正常工作（这会测试与migration包的兼容性）
	err := model.AutoMigrate(advancedModel)
	if err != nil {
		// 在没有数据库连接的情况下，这是预期的
		// 但至少验证了没有语法错误或panic
		if !strings.Contains(err.Error(), "连接") && !strings.Contains(err.Error(), "connection") {
			t.Errorf("Unexpected error in AutoMigrate: %v", err)
		}
	}
}

// TestPostgreSQLSequenceSupport 测试PostgreSQL序列支持
func TestPostgreSQLSequenceSupport(t *testing.T) {
	// 创建带有PostgreSQL特定标签的模型
	type PostgreSQLModel struct {
		BaseModel
		ID        uint64 `json:"id" torm:"primary_key,bigserial"`               // PostgreSQL BIGSERIAL
		SmallID   uint16 `json:"small_id" torm:"smallserial,unique"`            // PostgreSQL SMALLSERIAL
		NormalID  uint32 `json:"normal_id" torm:"serial,index"`                 // PostgreSQL SERIAL
		AutoIncID uint64 `json:"auto_inc_id" torm:"auto_increment,type:bigint"` // auto_increment without primary_key
		Name      string `json:"name" torm:"type:varchar,size:100"`
	}

	pgsqlModel := &PostgreSQLModel{}
	model := NewModel(pgsqlModel)

	// 验证模型配置
	if model.GetPrimaryKey() != "id" {
		t.Errorf("Expected primary key 'id', got '%s'", model.GetPrimaryKey())
	}

	// 验证可以创建查询
	query, err := model.Query()
	if err != nil {
		// 在没有数据库连接的情况下这是预期的
		if !strings.Contains(err.Error(), "连接") && !strings.Contains(err.Error(), "connection") {
			t.Errorf("Unexpected error creating query: %v", err)
		}
	} else if query == nil {
		t.Error("Expected query to be created")
	}
}

// TestMigrationPackageCompatibility 测试与migration包的兼容性
func TestMigrationPackageCompatibility(t *testing.T) {
	// 创建包含所有支持的torm标签的模型
	type CompleteModel struct {
		BaseModel
		// 主键和自增
		ID uint64 `json:"id" torm:"primary_key,auto_increment"`

		// 字符串类型with各种约束
		Username string `json:"username" torm:"type:varchar,size:50,unique,not_null"`
		Email    string `json:"email" torm:"type:varchar,size:100,unique,index:btree"`

		// 数字类型with精度
		Age     int     `json:"age" torm:"type:int,default:0,unsigned"`
		Balance float64 `json:"balance" torm:"type:decimal,precision:10,scale:2,default:0.00"`

		// 时间字段
		CreatedAt time.Time  `json:"created_at" torm:"auto_create_time"`
		UpdatedAt time.Time  `json:"updated_at" torm:"auto_update_time"`
		DeletedAt *time.Time `json:"deleted_at" torm:"soft_delete"`

		// 外键
		ProfileID uint `json:"profile_id" torm:"foreign_key:profiles.id,on_delete:cascade"`

		// JSON字段
		Metadata string `json:"metadata" torm:"type:json"`

		// 全文索引
		Description string `json:"description" torm:"type:text,fulltext"`

		// 生成列
		FullName string `json:"full_name" torm:"generated:virtual"`
	}

	completeModel := &CompleteModel{}
	model := NewModel(completeModel)

	// 验证所有相关配置都被正确设置
	if model.GetPrimaryKey() != "id" {
		t.Errorf("Expected primary key 'id', got '%s'", model.GetPrimaryKey())
	}

	if model.GetCreatedAtField() != "created_at" {
		t.Errorf("Expected created_at field 'created_at', got '%s'", model.GetCreatedAtField())
	}

	if model.GetUpdatedAtField() != "updated_at" {
		t.Errorf("Expected updated_at field 'updated_at', got '%s'", model.GetUpdatedAtField())
	}

	if !model.config.SoftDeletes {
		t.Error("Expected soft deletes to be enabled")
	}

	if model.config.DeletedAtCol != "deleted_at" {
		t.Errorf("Expected deleted_at field 'deleted_at', got '%s'", model.config.DeletedAtCol)
	}

	// 验证模型可以正常使用查询功能
	_, err := model.Where("username", "=", "test")
	if err != nil {
		if !strings.Contains(err.Error(), "连接") && !strings.Contains(err.Error(), "connection") {
			t.Errorf("Unexpected error in Where query: %v", err)
		}
	}

	// 验证参数式查询仍然工作
	_, err = model.Where("balance > ? AND status = ?", 100.0, "active")
	if err != nil {
		if !strings.Contains(err.Error(), "连接") && !strings.Contains(err.Error(), "connection") {
			t.Errorf("Unexpected error in parameterized query: %v", err)
		}
	}
}

// TestTableNameFunction 测试TableName函数功能
func TestTableNameFunction(t *testing.T) {
	// 测试默认TableName方法
	user := NewModel(&TestUser{})
	user.SetTable("custom_users")

	if tableName := user.TableName(); tableName != "custom_users" {
		t.Errorf("Expected table name 'custom_users', got '%s'", tableName)
	}

	// 测试GetTableName方法
	if tableName := user.GetTableName(); tableName != "custom_users" {
		t.Errorf("Expected table name 'custom_users', got '%s'", tableName)
	}
}

// TestCustomTableNameModel 测试自定义TableName方法的模型
type TestCustomTableNameModel struct {
	BaseModel
	ID   int    `json:"id" torm:"primary_key,auto_increment"`
	Name string `json:"name" torm:"type:varchar,size:100"`
}

// TableName 重写TableName方法
func (m TestCustomTableNameModel) TableName() string {
	return "my_custom_table"
}

func TestCustomTableName(t *testing.T) {
	model := &TestCustomTableNameModel{}

	// 测试静态TableName方法
	if tableName := model.TableName(); tableName != "my_custom_table" {
		t.Errorf("Expected table name 'my_custom_table', got '%s'", tableName)
	}

	// 测试空结构体的TableName方法
	emptyModel := TestCustomTableNameModel{}
	if tableName := emptyModel.TableName(); tableName != "my_custom_table" {
		t.Errorf("Expected table name 'my_custom_table' from empty struct, got '%s'", tableName)
	}
}
