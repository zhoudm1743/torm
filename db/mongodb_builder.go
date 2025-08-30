package db

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoQueryBuilder MongoDB查询构建器
type MongoQueryBuilder struct {
	connection     *MongoConnection
	collectionName string
	model          interface{}

	// MongoDB特有的查询条件
	filter     bson.M
	projection bson.M
	sort       bson.M
	skip       int64
	limit      int64

	// 聚合管道
	pipeline []bson.M

	// 事务
	session mongo.Session

	// 时间管理
	timeManager *TimeFieldManager
	timeFields  []TimeFieldInfo

	// 上下文
	ctx context.Context
}

// NewMongoQueryBuilder 创建MongoDB查询构建器
func NewMongoQueryBuilder(conn *MongoConnection) *MongoQueryBuilder {
	return &MongoQueryBuilder{
		connection:  conn,
		filter:      bson.M{},
		projection:  bson.M{},
		sort:        bson.M{},
		timeManager: NewTimeFieldManager(),
		timeFields:  make([]TimeFieldInfo, 0),
		ctx:         context.Background(),
	}
}

// Collection 设置集合名称
func (m *MongoQueryBuilder) Collection(name string) *MongoQueryBuilder {
	m.collectionName = name
	return m
}

// SetModel 设置关联的模型实例并分析时间字段
func (m *MongoQueryBuilder) SetModel(model interface{}) *MongoQueryBuilder {
	m.model = model
	if m.timeManager != nil {
		m.timeFields = m.timeManager.AnalyzeModelTimeFields(model)
	}
	return m
}

// GetModel 获取关联的模型实例
func (m *MongoQueryBuilder) GetModel() interface{} {
	return m.model
}

// Where 添加查询条件
func (m *MongoQueryBuilder) Where(field string, value interface{}) *MongoQueryBuilder {
	m.filter[field] = m.parseValue(field, value)
	return m
}

// WhereOp 添加操作符查询条件
func (m *MongoQueryBuilder) WhereOp(field string, operator string, value interface{}) *MongoQueryBuilder {
	mongoOp := m.parseOperator(operator)
	if mongoOp == "" {
		// 如果操作符无效，使用等于
		m.filter[field] = m.parseValue(field, value)
		return m
	}

	m.filter[field] = bson.M{mongoOp: m.parseValue(field, value)}
	return m
}

// WhereIn 添加IN查询条件
func (m *MongoQueryBuilder) WhereIn(field string, values []interface{}) *MongoQueryBuilder {
	parsedValues := make([]interface{}, len(values))
	for i, v := range values {
		parsedValues[i] = m.parseValue(field, v)
	}
	m.filter[field] = bson.M{"$in": parsedValues}
	return m
}

// WhereNotIn 添加NOT IN查询条件
func (m *MongoQueryBuilder) WhereNotIn(field string, values []interface{}) *MongoQueryBuilder {
	parsedValues := make([]interface{}, len(values))
	for i, v := range values {
		parsedValues[i] = m.parseValue(field, v)
	}
	m.filter[field] = bson.M{"$nin": parsedValues}
	return m
}

// WhereBetween 添加BETWEEN查询条件
func (m *MongoQueryBuilder) WhereBetween(field string, min, max interface{}) *MongoQueryBuilder {
	m.filter[field] = bson.M{
		"$gte": m.parseValue(field, min),
		"$lte": m.parseValue(field, max),
	}
	return m
}

// WhereRegex 添加正则表达式查询条件
func (m *MongoQueryBuilder) WhereRegex(field string, pattern string, options string) *MongoQueryBuilder {
	regex := primitive.Regex{Pattern: pattern, Options: options}
	m.filter[field] = regex
	return m
}

// WhereExists 添加字段存在查询条件
func (m *MongoQueryBuilder) WhereExists(field string) *MongoQueryBuilder {
	m.filter[field] = bson.M{"$exists": true}
	return m
}

// WhereNotExists 添加字段不存在查询条件
func (m *MongoQueryBuilder) WhereNotExists(field string) *MongoQueryBuilder {
	m.filter[field] = bson.M{"$exists": false}
	return m
}

// WhereNull 添加空值查询条件
func (m *MongoQueryBuilder) WhereNull(field string) *MongoQueryBuilder {
	m.filter[field] = nil
	return m
}

// WhereNotNull 添加非空值查询条件
func (m *MongoQueryBuilder) WhereNotNull(field string) *MongoQueryBuilder {
	m.filter[field] = bson.M{"$ne": nil}
	return m
}

// OrWhere 添加OR查询条件
func (m *MongoQueryBuilder) OrWhere(conditions ...bson.M) *MongoQueryBuilder {
	if len(conditions) == 0 {
		return m
	}

	if existing, ok := m.filter["$or"]; ok {
		if orArray, ok := existing.([]bson.M); ok {
			m.filter["$or"] = append(orArray, conditions...)
		}
	} else {
		m.filter["$or"] = conditions
	}

	return m
}

// AndWhere 添加AND查询条件
func (m *MongoQueryBuilder) AndWhere(conditions ...bson.M) *MongoQueryBuilder {
	if len(conditions) == 0 {
		return m
	}

	if existing, ok := m.filter["$and"]; ok {
		if andArray, ok := existing.([]bson.M); ok {
			m.filter["$and"] = append(andArray, conditions...)
		}
	} else {
		m.filter["$and"] = conditions
	}

	return m
}

// Select 设置投影字段
func (m *MongoQueryBuilder) Select(fields ...string) *MongoQueryBuilder {
	m.projection = bson.M{}
	for _, field := range fields {
		m.projection[field] = 1
	}
	return m
}

// Exclude 排除字段
func (m *MongoQueryBuilder) Exclude(fields ...string) *MongoQueryBuilder {
	if len(m.projection) == 0 {
		m.projection = bson.M{}
	}
	for _, field := range fields {
		m.projection[field] = 0
	}
	return m
}

// OrderBy 添加排序
func (m *MongoQueryBuilder) OrderBy(field string, direction string) *MongoQueryBuilder {
	order := 1
	if strings.ToLower(direction) == "desc" {
		order = -1
	}
	m.sort[field] = order
	return m
}

// OrderByDesc 降序排序
func (m *MongoQueryBuilder) OrderByDesc(field string) *MongoQueryBuilder {
	return m.OrderBy(field, "desc")
}

// OrderByAsc 升序排序
func (m *MongoQueryBuilder) OrderByAsc(field string) *MongoQueryBuilder {
	return m.OrderBy(field, "asc")
}

// Skip 设置跳过数量
func (m *MongoQueryBuilder) Skip(n int64) *MongoQueryBuilder {
	m.skip = n
	return m
}

// Limit 设置限制数量
func (m *MongoQueryBuilder) Limit(n int64) *MongoQueryBuilder {
	m.limit = n
	return m
}

// Page 分页设置
func (m *MongoQueryBuilder) Page(page, pageSize int64) *MongoQueryBuilder {
	m.skip = (page - 1) * pageSize
	m.limit = pageSize
	return m
}

// Context 设置上下文
func (m *MongoQueryBuilder) Context(ctx context.Context) *MongoQueryBuilder {
	m.ctx = ctx
	return m
}

// WithSession 设置会话（用于事务）
func (m *MongoQueryBuilder) WithSession(session mongo.Session) *MongoQueryBuilder {
	m.session = session
	return m
}

// Find 查找多个文档
func (m *MongoQueryBuilder) Find() ([]map[string]interface{}, error) {
	if m.collectionName == "" {
		return nil, NewError(ErrCodeInvalidParameter, "集合名称不能为空")
	}

	collection := m.connection.Collection(m.collectionName)
	if collection == nil {
		return nil, NewError(ErrCodeConnectionClosed, "MongoDB连接未建立")
	}

	// 构建查询选项
	opts := options.Find()

	if len(m.projection) > 0 {
		opts.SetProjection(m.projection)
	}

	if len(m.sort) > 0 {
		opts.SetSort(m.sort)
	}

	if m.skip > 0 {
		opts.SetSkip(m.skip)
	}

	if m.limit > 0 {
		opts.SetLimit(m.limit)
	}

	// 执行查询
	var cursor *mongo.Cursor
	var err error

	if m.session != nil {
		cursor, err = collection.Find(mongo.NewSessionContext(m.ctx, m.session), m.filter, opts)
	} else {
		cursor, err = collection.Find(m.ctx, m.filter, opts)
	}

	if err != nil {
		return nil, WrapError(err, ErrCodeQueryFailed, "MongoDB查询失败").
			WithContext("collection", m.collectionName).
			WithContext("filter", m.filter)
	}
	defer cursor.Close(m.ctx)

	// 解析结果
	var results []map[string]interface{}
	err = cursor.All(m.ctx, &results)
	if err != nil {
		return nil, WrapError(err, ErrCodeQueryFailed, "MongoDB结果解析失败").
			WithContext("collection", m.collectionName)
	}

	// 应用访问器处理
	return m.applyAccessors(results), nil
}

// First 查找第一个文档
func (m *MongoQueryBuilder) First() (map[string]interface{}, error) {
	if m.collectionName == "" {
		return nil, NewError(ErrCodeInvalidParameter, "集合名称不能为空")
	}

	collection := m.connection.Collection(m.collectionName)
	if collection == nil {
		return nil, NewError(ErrCodeConnectionClosed, "MongoDB连接未建立")
	}

	// 构建查询选项
	opts := options.FindOne()

	if len(m.projection) > 0 {
		opts.SetProjection(m.projection)
	}

	if len(m.sort) > 0 {
		opts.SetSort(m.sort)
	}

	if m.skip > 0 {
		opts.SetSkip(m.skip)
	}

	// 执行查询
	var result map[string]interface{}
	var err error

	if m.session != nil {
		err = collection.FindOne(mongo.NewSessionContext(m.ctx, m.session), m.filter, opts).Decode(&result)
	} else {
		err = collection.FindOne(m.ctx, m.filter, opts).Decode(&result)
	}

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrRecordNotFound.WithContext("collection", m.collectionName)
		}
		return nil, WrapError(err, ErrCodeQueryFailed, "MongoDB查询失败").
			WithContext("collection", m.collectionName).
			WithContext("filter", m.filter)
	}

	// 应用访问器处理
	if m.model != nil {
		processor := NewAccessorProcessor(m.model)
		result = processor.ProcessData(result)
	}

	return result, nil
}

// Count 计算文档数量
func (m *MongoQueryBuilder) Count() (int64, error) {
	if m.collectionName == "" {
		return 0, NewError(ErrCodeInvalidParameter, "集合名称不能为空")
	}

	collection := m.connection.Collection(m.collectionName)
	if collection == nil {
		return 0, NewError(ErrCodeConnectionClosed, "MongoDB连接未建立")
	}

	// 执行计数
	var count int64
	var err error

	if m.session != nil {
		count, err = collection.CountDocuments(mongo.NewSessionContext(m.ctx, m.session), m.filter)
	} else {
		count, err = collection.CountDocuments(m.ctx, m.filter)
	}

	if err != nil {
		return 0, WrapError(err, ErrCodeQueryFailed, "MongoDB计数失败").
			WithContext("collection", m.collectionName).
			WithContext("filter", m.filter)
	}

	return count, nil
}

// Insert 插入文档
func (m *MongoQueryBuilder) Insert(data map[string]interface{}) (interface{}, error) {
	if m.collectionName == "" {
		return nil, NewError(ErrCodeInvalidParameter, "集合名称不能为空")
	}

	if len(data) == 0 {
		return nil, NewError(ErrCodeInvalidParameter, "插入数据不能为空")
	}

	// 处理时间字段
	if m.timeManager != nil && len(m.timeFields) > 0 {
		data = m.timeManager.ProcessInsertData(data, m.timeFields)
	}

	collection := m.connection.Collection(m.collectionName)
	if collection == nil {
		return nil, NewError(ErrCodeConnectionClosed, "MongoDB连接未建立")
	}

	// 执行插入
	var result *mongo.InsertOneResult
	var err error

	if m.session != nil {
		result, err = collection.InsertOne(mongo.NewSessionContext(m.ctx, m.session), data)
	} else {
		result, err = collection.InsertOne(m.ctx, data)
	}

	if err != nil {
		return nil, WrapError(err, ErrCodeQueryFailed, "MongoDB插入失败").
			WithContext("collection", m.collectionName).
			WithContext("data", data)
	}

	return result.InsertedID, nil
}

// InsertMany 插入多个文档
func (m *MongoQueryBuilder) InsertMany(documents []interface{}) ([]interface{}, error) {
	if m.collectionName == "" {
		return nil, NewError(ErrCodeInvalidParameter, "集合名称不能为空")
	}

	if len(documents) == 0 {
		return nil, NewError(ErrCodeInvalidParameter, "插入数据不能为空")
	}

	// 处理时间字段
	if m.timeManager != nil && len(m.timeFields) > 0 {
		for i, doc := range documents {
			if docMap, ok := doc.(map[string]interface{}); ok {
				documents[i] = m.timeManager.ProcessInsertData(docMap, m.timeFields)
			}
		}
	}

	collection := m.connection.Collection(m.collectionName)
	if collection == nil {
		return nil, NewError(ErrCodeConnectionClosed, "MongoDB连接未建立")
	}

	// 执行批量插入
	var result *mongo.InsertManyResult
	var err error

	if m.session != nil {
		result, err = collection.InsertMany(mongo.NewSessionContext(m.ctx, m.session), documents)
	} else {
		result, err = collection.InsertMany(m.ctx, documents)
	}

	if err != nil {
		return nil, WrapError(err, ErrCodeQueryFailed, "MongoDB批量插入失败").
			WithContext("collection", m.collectionName).
			WithContext("count", len(documents))
	}

	return result.InsertedIDs, nil
}

// Update 更新文档
func (m *MongoQueryBuilder) Update(data map[string]interface{}) (int64, error) {
	if m.collectionName == "" {
		return 0, NewError(ErrCodeInvalidParameter, "集合名称不能为空")
	}

	if len(data) == 0 {
		return 0, NewError(ErrCodeInvalidParameter, "更新数据不能为空")
	}

	// 处理时间字段
	if m.timeManager != nil && len(m.timeFields) > 0 {
		data = m.timeManager.ProcessUpdateData(data, m.timeFields)
	}

	collection := m.connection.Collection(m.collectionName)
	if collection == nil {
		return 0, NewError(ErrCodeConnectionClosed, "MongoDB连接未建立")
	}

	// 构建更新文档
	update := bson.M{"$set": data}

	// 执行更新
	var result *mongo.UpdateResult
	var err error

	if m.session != nil {
		result, err = collection.UpdateMany(mongo.NewSessionContext(m.ctx, m.session), m.filter, update)
	} else {
		result, err = collection.UpdateMany(m.ctx, m.filter, update)
	}

	if err != nil {
		return 0, WrapError(err, ErrCodeQueryFailed, "MongoDB更新失败").
			WithContext("collection", m.collectionName).
			WithContext("filter", m.filter).
			WithContext("data", data)
	}

	return result.ModifiedCount, nil
}

// Delete 删除文档
func (m *MongoQueryBuilder) Delete() (int64, error) {
	if m.collectionName == "" {
		return 0, NewError(ErrCodeInvalidParameter, "集合名称不能为空")
	}

	collection := m.connection.Collection(m.collectionName)
	if collection == nil {
		return 0, NewError(ErrCodeConnectionClosed, "MongoDB连接未建立")
	}

	// 执行删除
	var result *mongo.DeleteResult
	var err error

	if m.session != nil {
		result, err = collection.DeleteMany(mongo.NewSessionContext(m.ctx, m.session), m.filter)
	} else {
		result, err = collection.DeleteMany(m.ctx, m.filter)
	}

	if err != nil {
		return 0, WrapError(err, ErrCodeQueryFailed, "MongoDB删除失败").
			WithContext("collection", m.collectionName).
			WithContext("filter", m.filter)
	}

	return result.DeletedCount, nil
}

// parseOperator 解析操作符
func (m *MongoQueryBuilder) parseOperator(operator string) string {
	operatorMap := map[string]string{
		"=":         "$eq",
		"!=":        "$ne",
		"<>":        "$ne",
		">":         "$gt",
		">=":        "$gte",
		"<":         "$lt",
		"<=":        "$lte",
		"in":        "$in",
		"not in":    "$nin",
		"nin":       "$nin",
		"exists":    "$exists",
		"regex":     "$regex",
		"like":      "$regex",
		"size":      "$size",
		"type":      "$type",
		"all":       "$all",
		"mod":       "$mod",
		"elemMatch": "$elemMatch",
	}

	return operatorMap[strings.ToLower(operator)]
}

// parseValue 解析值
func (m *MongoQueryBuilder) parseValue(field string, value interface{}) interface{} {
	// 处理ObjectID
	if field == "_id" || strings.HasSuffix(field, "_id") {
		if str, ok := value.(string); ok {
			if objectID, err := primitive.ObjectIDFromHex(str); err == nil {
				return objectID
			}
		}
	}

	// 处理时间
	if t, ok := value.(time.Time); ok {
		return primitive.NewDateTimeFromTime(t)
	}

	// 处理字符串时间
	if str, ok := value.(string); ok {
		if t, err := time.Parse(time.RFC3339, str); err == nil {
			return primitive.NewDateTimeFromTime(t)
		}
		if t, err := time.Parse("2006-01-02 15:04:05", str); err == nil {
			return primitive.NewDateTimeFromTime(t)
		}
	}

	return value
}

// applyAccessors 应用访问器
func (m *MongoQueryBuilder) applyAccessors(data []map[string]interface{}) []map[string]interface{} {
	if m.model == nil {
		return data
	}

	processor := NewAccessorProcessor(m.model)
	return processor.ProcessDataSlice(data)
}

// ToFilter 获取过滤条件（用于调试）
func (m *MongoQueryBuilder) ToFilter() bson.M {
	return m.filter
}

// Clone 克隆查询构建器
func (m *MongoQueryBuilder) Clone() *MongoQueryBuilder {
	clone := &MongoQueryBuilder{
		connection:     m.connection,
		collectionName: m.collectionName,
		model:          m.model,
		filter:         bson.M{},
		projection:     bson.M{},
		sort:           bson.M{},
		skip:           m.skip,
		limit:          m.limit,
		session:        m.session,
		ctx:            m.ctx,
	}

	// 深拷贝过滤条件
	for k, v := range m.filter {
		clone.filter[k] = v
	}

	// 深拷贝投影
	for k, v := range m.projection {
		clone.projection[k] = v
	}

	// 深拷贝排序
	for k, v := range m.sort {
		clone.sort[k] = v
	}

	return clone
}
