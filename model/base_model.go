package model

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/zhoudm1743/torm/db"
	"github.com/zhoudm1743/torm/migration"
)

// BaseModel 基础模型 - 极简化版本
type BaseModel struct {
	// 基础配置
	tableName  string
	primaryKey string
	connection string
	attributes map[string]interface{}

	// 状态
	exists bool

	// 时间戳
	timestamps bool
	createdAt  string
	updatedAt  string

	// 软删除
	softDeletes bool
	deletedAt   string
}

// NewBaseModel 创建新的基础模型
func NewBaseModel() *BaseModel {
	return &BaseModel{
		connection:  "default",
		primaryKey:  "id",
		attributes:  make(map[string]interface{}),
		exists:      false,
		timestamps:  true,
		createdAt:   "created_at",
		updatedAt:   "updated_at",
		softDeletes: false,
		deletedAt:   "deleted_at",
	}
}

// 配置方法

// SetTable 设置表名
func (m *BaseModel) SetTable(tableName string) *BaseModel {
	m.tableName = tableName
	return m
}

// TableName 获取表名
func (m *BaseModel) TableName() string {
	return m.tableName
}

// SetPrimaryKey 设置主键
func (m *BaseModel) SetPrimaryKey(key string) *BaseModel {
	m.primaryKey = key
	return m
}

// SetConnection 设置数据库连接
func (m *BaseModel) SetConnection(connection string) *BaseModel {
	m.connection = connection
	return m
}

// EnableTimestamps 启用时间戳
func (m *BaseModel) EnableTimestamps() *BaseModel {
	m.timestamps = true
	return m
}

// DisableTimestamps 禁用时间戳
func (m *BaseModel) DisableTimestamps() *BaseModel {
	m.timestamps = false
	return m
}

// EnableSoftDeletes 启用软删除
func (m *BaseModel) EnableSoftDeletes() *BaseModel {
	m.softDeletes = true
	return m
}

// 属性方法

// SetAttribute 设置属性
func (m *BaseModel) SetAttribute(key string, value interface{}) *BaseModel {
	m.attributes[key] = value
	return m
}

// GetAttribute 获取属性
func (m *BaseModel) GetAttribute(key string) interface{} {
	return m.attributes[key]
}

// GetAttributes 获取所有属性
func (m *BaseModel) GetAttributes() map[string]interface{} {
	return m.attributes
}

// Fill 批量设置属性
func (m *BaseModel) Fill(data map[string]interface{}) *BaseModel {
	for key, value := range data {
		m.attributes[key] = value
	}
	return m
}

// ToMap 转换为Map
func (m *BaseModel) ToMap() map[string]interface{} {
	return m.attributes
}

// ToJSON 转换为JSON
func (m *BaseModel) ToJSON() (string, error) {
	data, err := json.Marshal(m.attributes)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetKey 获取主键值
func (m *BaseModel) GetKey() interface{} {
	return m.attributes[m.primaryKey]
}

// 查询方法 - 基于查询构建器

// newQuery 创建查询构建器
func (m *BaseModel) newQuery() (*db.QueryBuilder, error) {
	if m.tableName == "" {
		return nil, fmt.Errorf("表名未设置")
	}

	builder, err := db.Table(m.tableName, m.connection)
	if err != nil {
		return nil, err
	}

	// 如果启用了软删除，自动添加过滤条件
	if m.softDeletes {
		builder = builder.Where(m.deletedAt + " IS NULL")
	}

	return builder, nil
}

// Where 添加查询条件
func (m *BaseModel) Where(args ...interface{}) *ModelQueryBuilder {
	query, err := m.newQuery()
	if err != nil {
		return &ModelQueryBuilder{err: err, model: m}
	}

	return &ModelQueryBuilder{
		query: query.Where(args...),
		model: m,
	}
}

// WhereIn WHERE IN条件
func (m *BaseModel) WhereIn(field string, values []interface{}) *ModelQueryBuilder {
	query, err := m.newQuery()
	if err != nil {
		return &ModelQueryBuilder{err: err, model: m}
	}

	return &ModelQueryBuilder{
		query: query.WhereIn(field, values),
		model: m,
	}
}

// WhereNotIn WHERE NOT IN条件
func (m *BaseModel) WhereNotIn(field string, values []interface{}) *ModelQueryBuilder {
	query, err := m.newQuery()
	if err != nil {
		return &ModelQueryBuilder{err: err, model: m}
	}

	return &ModelQueryBuilder{
		query: query.WhereNotIn(field, values),
		model: m,
	}
}

// WhereBetween WHERE BETWEEN条件
func (m *BaseModel) WhereBetween(field string, values []interface{}) *ModelQueryBuilder {
	query, err := m.newQuery()
	if err != nil {
		return &ModelQueryBuilder{err: err, model: m}
	}

	return &ModelQueryBuilder{
		query: query.WhereBetween(field, values),
		model: m,
	}
}

// WhereNotBetween WHERE NOT BETWEEN条件
func (m *BaseModel) WhereNotBetween(field string, values []interface{}) *ModelQueryBuilder {
	query, err := m.newQuery()
	if err != nil {
		return &ModelQueryBuilder{err: err, model: m}
	}

	return &ModelQueryBuilder{
		query: query.WhereNotBetween(field, values),
		model: m,
	}
}

// WhereNull WHERE IS NULL条件
func (m *BaseModel) WhereNull(field string) *ModelQueryBuilder {
	query, err := m.newQuery()
	if err != nil {
		return &ModelQueryBuilder{err: err, model: m}
	}

	return &ModelQueryBuilder{
		query: query.WhereNull(field),
		model: m,
	}
}

// WhereNotNull WHERE IS NOT NULL条件
func (m *BaseModel) WhereNotNull(field string) *ModelQueryBuilder {
	query, err := m.newQuery()
	if err != nil {
		return &ModelQueryBuilder{err: err, model: m}
	}

	return &ModelQueryBuilder{
		query: query.WhereNotNull(field),
		model: m,
	}
}

// WhereExists WHERE EXISTS条件
func (m *BaseModel) WhereExists(subQuery interface{}) *ModelQueryBuilder {
	query, err := m.newQuery()
	if err != nil {
		return &ModelQueryBuilder{err: err, model: m}
	}

	return &ModelQueryBuilder{
		query: query.WhereExists(subQuery),
		model: m,
	}
}

// WhereNotExists WHERE NOT EXISTS条件
func (m *BaseModel) WhereNotExists(subQuery interface{}) *ModelQueryBuilder {
	query, err := m.newQuery()
	if err != nil {
		return &ModelQueryBuilder{err: err, model: m}
	}

	return &ModelQueryBuilder{
		query: query.WhereNotExists(subQuery),
		model: m,
	}
}

// WhereRaw 原生WHERE条件
func (m *BaseModel) WhereRaw(raw string, bindings ...interface{}) *ModelQueryBuilder {
	query, err := m.newQuery()
	if err != nil {
		return &ModelQueryBuilder{err: err, model: m}
	}

	return &ModelQueryBuilder{
		query: query.WhereRaw(raw, bindings...),
		model: m,
	}
}

// OrWhere 添加OR WHERE条件
func (m *BaseModel) OrWhere(args ...interface{}) *ModelQueryBuilder {
	query, err := m.newQuery()
	if err != nil {
		return &ModelQueryBuilder{err: err, model: m}
	}

	return &ModelQueryBuilder{
		query: query.OrWhere(args...),
		model: m,
	}
}

// Find 根据主键查找
func (m *BaseModel) Find(id interface{}) error {
	query, err := m.newQuery()
	if err != nil {
		return err
	}

	result, err := query.Where(m.primaryKey, "=", id).First()
	if err != nil {
		return err
	}

	m.fill(result)
	m.exists = true
	return nil
}

// First 获取第一条记录
func (m *BaseModel) First() error {
	query, err := m.newQuery()
	if err != nil {
		return err
	}

	result, err := query.First()
	if err != nil {
		return err
	}

	m.fill(result)
	m.exists = true
	return nil
}

// Get 获取多条记录
func (m *BaseModel) Get() ([]map[string]interface{}, error) {
	query, err := m.newQuery()
	if err != nil {
		return nil, err
	}

	return query.Get()
}

// Count 计算记录数量
func (m *BaseModel) Count() (int64, error) {
	query, err := m.newQuery()
	if err != nil {
		return 0, err
	}

	return query.Count()
}

// Exists 检查记录是否存在
func (m *BaseModel) Exists() (bool, error) {
	query, err := m.newQuery()
	if err != nil {
		return false, err
	}

	return query.Exists()
}

// Insert 插入数据
func (m *BaseModel) Insert(data map[string]interface{}) (int64, error) {
	query, err := m.newQuery()
	if err != nil {
		return 0, err
	}

	// 处理时间戳
	if m.timestamps {
		now := time.Now()
		data[m.createdAt] = now
		data[m.updatedAt] = now
	}

	return query.Insert(data)
}

// InsertBatch 批量插入数据
func (m *BaseModel) InsertBatch(data []map[string]interface{}) (int64, error) {
	query, err := m.newQuery()
	if err != nil {
		return 0, err
	}

	// 处理时间戳
	if m.timestamps {
		now := time.Now()
		for i := range data {
			data[i][m.createdAt] = now
			data[i][m.updatedAt] = now
		}
	}

	return query.InsertBatch(data)
}

// Select 选择字段
func (m *BaseModel) Select(fields ...string) *ModelQueryBuilder {
	query, err := m.newQuery()
	if err != nil {
		return &ModelQueryBuilder{err: err, model: m}
	}

	return &ModelQueryBuilder{
		query: query.Select(fields...),
		model: m,
	}
}

// OrderBy 排序
func (m *BaseModel) OrderBy(field string, direction string) *ModelQueryBuilder {
	query, err := m.newQuery()
	if err != nil {
		return &ModelQueryBuilder{err: err, model: m}
	}

	return &ModelQueryBuilder{
		query: query.OrderBy(field, direction),
		model: m,
	}
}

// Limit 限制数量
func (m *BaseModel) Limit(limit int) *ModelQueryBuilder {
	query, err := m.newQuery()
	if err != nil {
		return &ModelQueryBuilder{err: err, model: m}
	}

	return &ModelQueryBuilder{
		query: query.Limit(limit),
		model: m,
	}
}

// Offset 偏移量
func (m *BaseModel) Offset(offset int) *ModelQueryBuilder {
	query, err := m.newQuery()
	if err != nil {
		return &ModelQueryBuilder{err: err, model: m}
	}

	return &ModelQueryBuilder{
		query: query.Offset(offset),
		model: m,
	}
}

// GroupBy 分组查询
func (m *BaseModel) GroupBy(fields ...string) *ModelQueryBuilder {
	query, err := m.newQuery()
	if err != nil {
		return &ModelQueryBuilder{err: err, model: m}
	}

	return &ModelQueryBuilder{
		query: query.GroupBy(fields...),
		model: m,
	}
}

// Join 内连接
func (m *BaseModel) Join(table string, first string, operator string, second string) *ModelQueryBuilder {
	query, err := m.newQuery()
	if err != nil {
		return &ModelQueryBuilder{err: err, model: m}
	}

	return &ModelQueryBuilder{
		query: query.Join(table, first, operator, second),
		model: m,
	}
}

// LeftJoin 左连接
func (m *BaseModel) LeftJoin(table string, first string, operator string, second string) *ModelQueryBuilder {
	query, err := m.newQuery()
	if err != nil {
		return &ModelQueryBuilder{err: err, model: m}
	}

	return &ModelQueryBuilder{
		query: query.LeftJoin(table, first, operator, second),
		model: m,
	}
}

// RightJoin 右连接
func (m *BaseModel) RightJoin(table string, first string, operator string, second string) *ModelQueryBuilder {
	query, err := m.newQuery()
	if err != nil {
		return &ModelQueryBuilder{err: err, model: m}
	}

	return &ModelQueryBuilder{
		query: query.RightJoin(table, first, operator, second),
		model: m,
	}
}

// InnerJoin 内连接
func (m *BaseModel) InnerJoin(table string, first string, operator string, second string) *ModelQueryBuilder {
	query, err := m.newQuery()
	if err != nil {
		return &ModelQueryBuilder{err: err, model: m}
	}

	return &ModelQueryBuilder{
		query: query.InnerJoin(table, first, operator, second),
		model: m,
	}
}

// Save 保存模型
func (m *BaseModel) Save() error {
	query, err := m.newQuery()
	if err != nil {
		return err
	}

	// 准备数据
	data := make(map[string]interface{})
	for key, value := range m.attributes {
		data[key] = value
	}

	// 处理时间戳
	if m.timestamps {
		now := time.Now()
		if !m.exists {
			data[m.createdAt] = now
		}
		data[m.updatedAt] = now
	}

	if m.exists {
		// 更新
		primaryKeyValue := m.GetKey()
		if primaryKeyValue == nil {
			return fmt.Errorf("主键值不能为空")
		}

		delete(data, m.primaryKey) // 移除主键，避免更新主键
		_, err = query.Where(m.primaryKey, "=", primaryKeyValue).Update(data)
		if err != nil {
			return err
		}

		// 更新属性
		for key, value := range data {
			m.attributes[key] = value
		}
	} else {
		// 插入
		id, err := query.Insert(data)
		if err != nil {
			return err
		}

		// 设置主键 - 只有在ID大于0时才设置，PostgreSQL可能返回0
		if id > 0 {
			m.attributes[m.primaryKey] = id
		}
		m.exists = true

		// 更新其他属性
		for key, value := range data {
			m.attributes[key] = value
		}
	}

	return nil
}

// Delete 删除模型
func (m *BaseModel) Delete() error {
	if !m.exists {
		return fmt.Errorf("记录不存在")
	}

	primaryKeyValue := m.GetKey()
	if primaryKeyValue == nil {
		return fmt.Errorf("主键值不能为空")
	}

	query, err := m.newQuery()
	if err != nil {
		return err
	}

	if m.softDeletes {
		// 软删除
		data := map[string]interface{}{
			m.deletedAt: time.Now(),
		}
		_, err = query.Where(m.primaryKey, "=", primaryKeyValue).Update(data)
	} else {
		// 硬删除
		_, err = query.Where(m.primaryKey, "=", primaryKeyValue).Delete()
	}

	return err
}

// fill 填充属性
func (m *BaseModel) fill(data map[string]interface{}) {
	for key, value := range data {
		m.attributes[key] = value
	}
}

// AutoMigrate 自动迁移 - 支持传入多个模型实例
func (m *BaseModel) AutoMigrate(models ...interface{}) error {
	if m.tableName == "" {
		return fmt.Errorf("表名未设置")
	}

	// 获取数据库连接
	conn, err := db.DefaultManager().Connection(m.connection)
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 创建自动迁移器
	migrator := migration.NewAutoMigrator(conn)

	// 如果没有传入模型，尝试使用当前模型
	if len(models) == 0 {
		// 这种情况下无法确定具体模型类型，返回提示信息
		return fmt.Errorf("请传入模型实例进行迁移，例如: admin.AutoMigrate(admin)")
	}

	// 迁移所有传入的模型
	for i, model := range models {
		// 对于第一个模型，使用当前BaseModel的表名
		tableName := m.tableName

		// 对于后续模型，尝试从模型本身获取表名
		if i > 0 {
			if tableNamer, ok := model.(interface{ TableName() string }); ok {
				if customTableName := tableNamer.TableName(); customTableName != "" {
					tableName = customTableName
				}
			} else {
				// 如果模型没有TableName方法，从类型名推断
				tableName = getTableNameFromModelType(model)
			}
		}

		// 执行迁移
		err = migrator.MigrateModel(model, tableName)
		if err != nil {
			return fmt.Errorf("模型 %d 自动迁移失败: %w", i+1, err)
		}
	}

	return nil
}

// getTableNameFromModelType 从模型类型推断表名
func getTableNameFromModelType(model interface{}) string {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	typeName := modelType.Name()

	// 简单的命名转换：User -> users, Admin -> admins
	// 你可以根据需要添加更复杂的命名规则
	tableName := strings.ToLower(typeName)

	// 如果不是以s结尾，添加s（简单的复数化）
	if !strings.HasSuffix(tableName, "s") {
		tableName += "s"
	}

	return tableName
}

// DetectConfigFromStruct 从结构体检测配置（保持兼容性）
func (m *BaseModel) DetectConfigFromStruct(model interface{}) {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 如果没有设置表名，从结构体名推断
	if m.tableName == "" {
		tableName := toSnakeCase(modelType.Name())
		m.SetTable(tableName)
	}

	// TODO: 这里可以解析TORM标签，配置字段信息
	// 新版本会大大简化这个过程
}

// toSnakeCase 转换为蛇形命名
func toSnakeCase(str string) string {
	var result strings.Builder
	for i, r := range str {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r - 'A' + 'a')
	}
	return result.String()
}

// 更多BaseModel的查询构建器接口

// Cache 启用查询缓存
func (m *BaseModel) Cache(ttl time.Duration) *ModelQueryBuilder {
	query, err := m.newQuery()
	if err != nil {
		return &ModelQueryBuilder{err: err, model: m}
	}

	return &ModelQueryBuilder{
		query: query.Cache(ttl),
		model: m,
	}
}

// CacheWithTags 启用带标签的查询缓存
func (m *BaseModel) CacheWithTags(ttl time.Duration, tags ...string) *ModelQueryBuilder {
	query, err := m.newQuery()
	if err != nil {
		return &ModelQueryBuilder{err: err, model: m}
	}

	return &ModelQueryBuilder{
		query: query.CacheWithTags(ttl, tags...),
		model: m,
	}
}

// CacheKey 设置自定义缓存键
func (m *BaseModel) CacheKey(key string) *ModelQueryBuilder {
	query, err := m.newQuery()
	if err != nil {
		return &ModelQueryBuilder{err: err, model: m}
	}

	return &ModelQueryBuilder{
		query: query.CacheKey(key),
		model: m,
	}
}

// Page 分页查询
func (m *BaseModel) Page(page, pageSize int) *ModelQueryBuilder {
	query, err := m.newQuery()
	if err != nil {
		return &ModelQueryBuilder{err: err, model: m}
	}

	return &ModelQueryBuilder{
		query: query.Page(page, pageSize),
		model: m,
	}
}

// Paginate 分页查询
func (m *BaseModel) Paginate(page, perPage int) (interface{}, error) {
	query, err := m.newQuery()
	if err != nil {
		return nil, err
	}

	return query.Paginate(page, perPage)
}
