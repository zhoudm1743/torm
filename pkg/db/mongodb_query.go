package db

import (
	"context"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoQuery MongoDB查询构建器
type MongoQuery struct {
	collection *mongo.Collection
	session    mongo.Session
	filter     bson.M
	projection bson.M
	sort       bson.M
	skip       *int64
	limit      *int64
	logger     LoggerInterface
}

// NewMongoQuery 创建MongoDB查询构建器
func NewMongoQuery(collection *mongo.Collection, logger LoggerInterface) *MongoQuery {
	return &MongoQuery{
		collection: collection,
		filter:     bson.M{},
		projection: bson.M{},
		sort:       bson.M{},
		logger:     logger,
	}
}

// WithSession 设置会话（用于事务）
func (q *MongoQuery) WithSession(session mongo.Session) *MongoQuery {
	q.session = session
	return q
}

// Where 添加查询条件
func (q *MongoQuery) Where(field string, value interface{}) *MongoQuery {
	newQuery := *q
	newQuery.filter = make(bson.M)
	for k, v := range q.filter {
		newQuery.filter[k] = v
	}
	newQuery.filter[field] = value
	return &newQuery
}

// WhereIn 添加IN查询条件
func (q *MongoQuery) WhereIn(field string, values []interface{}) *MongoQuery {
	q.filter[field] = bson.M{"$in": values}
	return q
}

// WhereNot 添加NOT查询条件
func (q *MongoQuery) WhereNot(field string, value interface{}) *MongoQuery {
	q.filter[field] = bson.M{"$ne": value}
	return q
}

// WhereGt 添加大于查询条件
func (q *MongoQuery) WhereGt(field string, value interface{}) *MongoQuery {
	q.filter[field] = bson.M{"$gt": value}
	return q
}

// WhereGte 添加大于等于查询条件
func (q *MongoQuery) WhereGte(field string, value interface{}) *MongoQuery {
	q.filter[field] = bson.M{"$gte": value}
	return q
}

// WhereLt 添加小于查询条件
func (q *MongoQuery) WhereLt(field string, value interface{}) *MongoQuery {
	q.filter[field] = bson.M{"$lt": value}
	return q
}

// WhereLte 添加小于等于查询条件
func (q *MongoQuery) WhereLte(field string, value interface{}) *MongoQuery {
	q.filter[field] = bson.M{"$lte": value}
	return q
}

// WhereRegex 添加正则表达式查询条件
func (q *MongoQuery) WhereRegex(field string, pattern string, options string) *MongoQuery {
	regex := primitive.Regex{Pattern: pattern, Options: options}
	q.filter[field] = regex
	return q
}

// WhereExists 添加字段存在查询条件
func (q *MongoQuery) WhereExists(field string, exists bool) *MongoQuery {
	q.filter[field] = bson.M{"$exists": exists}
	return q
}

// Select 设置返回字段
func (q *MongoQuery) Select(fields ...string) *MongoQuery {
	for _, field := range fields {
		q.projection[field] = 1
	}
	return q
}

// Exclude 排除字段
func (q *MongoQuery) Exclude(fields ...string) *MongoQuery {
	for _, field := range fields {
		q.projection[field] = 0
	}
	return q
}

// OrderBy 设置排序
func (q *MongoQuery) OrderBy(field string, order int) *MongoQuery {
	q.sort[field] = order // 1 for ASC, -1 for DESC
	return q
}

// OrderByAsc 升序排序
func (q *MongoQuery) OrderByAsc(field string) *MongoQuery {
	return q.OrderBy(field, 1)
}

// OrderByDesc 降序排序
func (q *MongoQuery) OrderByDesc(field string) *MongoQuery {
	return q.OrderBy(field, -1)
}

// Skip 设置跳过记录数
func (q *MongoQuery) Skip(skip int64) *MongoQuery {
	q.skip = &skip
	return q
}

// Limit 设置限制记录数
func (q *MongoQuery) Limit(limit int64) *MongoQuery {
	q.limit = &limit
	return q
}

// Page 分页查询
func (q *MongoQuery) Page(page, pageSize int64) *MongoQuery {
	skip := (page - 1) * pageSize
	q.Skip(skip)
	q.Limit(pageSize)
	return q
}

// buildFindOptions 构建查询选项
func (q *MongoQuery) buildFindOptions() *options.FindOptions {
	opts := options.Find()

	if len(q.projection) > 0 {
		opts.SetProjection(q.projection)
	}

	if len(q.sort) > 0 {
		opts.SetSort(q.sort)
	}

	if q.skip != nil {
		opts.SetSkip(*q.skip)
	}

	if q.limit != nil {
		opts.SetLimit(*q.limit)
	}

	return opts
}

// Find 查询多条记录
func (q *MongoQuery) Find(ctx context.Context) (*mongo.Cursor, error) {
	opts := q.buildFindOptions()

	if q.session != nil {
		return q.collection.Find(mongo.NewSessionContext(ctx, q.session), q.filter, opts)
	}

	return q.collection.Find(ctx, q.filter, opts)
}

// FindOne 查询单条记录
func (q *MongoQuery) FindOne(ctx context.Context) *mongo.SingleResult {
	opts := options.FindOne()

	if len(q.projection) > 0 {
		opts.SetProjection(q.projection)
	}

	if len(q.sort) > 0 {
		opts.SetSort(q.sort)
	}

	if q.skip != nil {
		opts.SetSkip(*q.skip)
	}

	if q.session != nil {
		return q.collection.FindOne(mongo.NewSessionContext(ctx, q.session), q.filter, opts)
	}

	return q.collection.FindOne(ctx, q.filter, opts)
}

// Count 统计记录数
func (q *MongoQuery) Count(ctx context.Context) (int64, error) {
	if q.session != nil {
		return q.collection.CountDocuments(mongo.NewSessionContext(ctx, q.session), q.filter)
	}

	return q.collection.CountDocuments(ctx, q.filter)
}

// InsertOne 插入单条记录
func (q *MongoQuery) InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
	if q.session != nil {
		return q.collection.InsertOne(mongo.NewSessionContext(ctx, q.session), document)
	}

	return q.collection.InsertOne(ctx, document)
}

// InsertMany 插入多条记录
func (q *MongoQuery) InsertMany(ctx context.Context, documents []interface{}) (*mongo.InsertManyResult, error) {
	if q.session != nil {
		return q.collection.InsertMany(mongo.NewSessionContext(ctx, q.session), documents)
	}

	return q.collection.InsertMany(ctx, documents)
}

// UpdateOne 更新单条记录
func (q *MongoQuery) UpdateOne(ctx context.Context, update interface{}) (*mongo.UpdateResult, error) {
	if q.session != nil {
		return q.collection.UpdateOne(mongo.NewSessionContext(ctx, q.session), q.filter, update)
	}

	return q.collection.UpdateOne(ctx, q.filter, update)
}

// UpdateMany 更新多条记录
func (q *MongoQuery) UpdateMany(ctx context.Context, update interface{}) (*mongo.UpdateResult, error) {
	if q.session != nil {
		return q.collection.UpdateMany(mongo.NewSessionContext(ctx, q.session), q.filter, update)
	}

	return q.collection.UpdateMany(ctx, q.filter, update)
}

// ReplaceOne 替换单条记录
func (q *MongoQuery) ReplaceOne(ctx context.Context, replacement interface{}) (*mongo.UpdateResult, error) {
	if q.session != nil {
		return q.collection.ReplaceOne(mongo.NewSessionContext(ctx, q.session), q.filter, replacement)
	}

	return q.collection.ReplaceOne(ctx, q.filter, replacement)
}

// DeleteOne 删除单条记录
func (q *MongoQuery) DeleteOne(ctx context.Context) (*mongo.DeleteResult, error) {
	if q.session != nil {
		return q.collection.DeleteOne(mongo.NewSessionContext(ctx, q.session), q.filter)
	}

	return q.collection.DeleteOne(ctx, q.filter)
}

// DeleteMany 删除多条记录
func (q *MongoQuery) DeleteMany(ctx context.Context) (*mongo.DeleteResult, error) {
	if q.session != nil {
		return q.collection.DeleteMany(mongo.NewSessionContext(ctx, q.session), q.filter)
	}

	return q.collection.DeleteMany(ctx, q.filter)
}

// Distinct 获取不重复的字段值
func (q *MongoQuery) Distinct(ctx context.Context, field string) ([]interface{}, error) {
	if q.session != nil {
		return q.collection.Distinct(mongo.NewSessionContext(ctx, q.session), field, q.filter)
	}

	return q.collection.Distinct(ctx, field, q.filter)
}

// Aggregate 聚合查询
func (q *MongoQuery) Aggregate(ctx context.Context, pipeline interface{}) (*mongo.Cursor, error) {
	if q.session != nil {
		return q.collection.Aggregate(mongo.NewSessionContext(ctx, q.session), pipeline)
	}

	return q.collection.Aggregate(ctx, pipeline)
}

// Inc 字段自增
func (q *MongoQuery) Inc(field string, value interface{}) *MongoQuery {
	return q
}

// Set 设置字段值
func (q *MongoQuery) Set(field string, value interface{}) *MongoQuery {
	return q
}

// Unset 删除字段
func (q *MongoQuery) Unset(fields ...string) *MongoQuery {
	return q
}

// ToArray 将游标转换为数组
func ToArray(ctx context.Context, cursor *mongo.Cursor, result interface{}) error {
	return cursor.All(ctx, result)
}

// ToArrayWithType 将游标转换为指定类型的数组
func ToArrayWithType(ctx context.Context, cursor *mongo.Cursor, elementType reflect.Type) (interface{}, error) {
	// 创建slice类型
	sliceType := reflect.SliceOf(elementType)
	sliceValue := reflect.New(sliceType).Elem()

	// 解析所有文档
	for cursor.Next(ctx) {
		// 创建新元素
		elemValue := reflect.New(elementType).Elem()

		// 解码到元素中
		if err := cursor.Decode(elemValue.Addr().Interface()); err != nil {
			return nil, err
		}

		// 添加到slice中
		sliceValue = reflect.Append(sliceValue, elemValue)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return sliceValue.Interface(), nil
}
