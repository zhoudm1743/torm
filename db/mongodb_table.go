package db

import (
	"reflect"
	"strings"
	
	"go.mongodb.org/mongo-driver/bson"
)

// MongoTable 创建MongoDB查询构建器
func MongoTable(collectionName string, connectionName ...string) (*MongoQueryBuilder, error) {
	// 确定连接名称
	connName := "default"
	if len(connectionName) > 0 && connectionName[0] != "" {
		connName = connectionName[0]
	}
	
	// 获取连接
	conn, err := DefaultManager().Connection(connName)
	if err != nil {
		return nil, WrapError(err, ErrCodeConnectionFailed, "获取MongoDB连接失败").
			WithContext("connection", connName).
			WithContext("collection", collectionName)
	}
	
	// 检查是否是MongoDB连接
	mongoConn, ok := conn.(*MongoConnection)
	if !ok {
		return nil, NewError(ErrCodeInvalidParameter, "连接不是MongoDB类型").
			WithContext("connection", connName).
			WithContext("driver", conn.GetDriver())
	}
	
	// 创建MongoDB查询构建器
	builder := NewMongoQueryBuilder(mongoConn)
	builder.Collection(collectionName)
	
	return builder, nil
}

// MongoModel 为模型创建MongoDB查询构建器
func MongoModel(model interface{}, connectionName ...string) (*MongoQueryBuilder, error) {
	// 从模型获取集合名称
	collectionName := getCollectionNameFromModel(model)
	if collectionName == "" {
		return nil, NewError(ErrCodeInvalidParameter, "无法从模型获取集合名称")
	}
	
	// 创建查询构建器
	builder, err := MongoTable(collectionName, connectionName...)
	if err != nil {
		return nil, err
	}
	
	// 设置模型
	builder.model = model
	
	return builder, nil
}

// getCollectionNameFromModel 从模型获取集合名称
func getCollectionNameFromModel(model interface{}) string {
	// 使用反射获取类型名称
	modelType := getModelType(model)
	if modelType == nil {
		return ""
	}
	
	// 获取模型值
	modelValue := reflect.ValueOf(model)
	
	// 检查是否有CollectionName方法
	if method := modelValue.MethodByName("CollectionName"); method.IsValid() {
		if results := method.Call([]reflect.Value{}); len(results) > 0 {
			if name, ok := results[0].Interface().(string); ok && name != "" {
				return name
			}
		}
	}
	
	// 检查是否有TableName方法（兼容SQL模型）
	if method := modelValue.MethodByName("TableName"); method.IsValid() {
		if results := method.Call([]reflect.Value{}); len(results) > 0 {
			if name, ok := results[0].Interface().(string); ok && name != "" {
				return name
			}
		}
	}
	
	// 默认使用类型名称的复数形式（小写+s）
	typeName := strings.ToLower(modelType.Name())
	if !strings.HasSuffix(typeName, "s") {
		typeName += "s"
	}
	
	return typeName
}

// getModelType 获取模型的反射类型
func getModelType(model interface{}) reflect.Type {
	if model == nil {
		return nil
	}
	
	modelType := reflect.TypeOf(model)
	
	// 如果是指针，获取元素类型
	for modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	
	// 只处理结构体类型
	if modelType.Kind() != reflect.Struct {
		return nil
	}
	
	return modelType
}

// MongoAggregate 创建MongoDB聚合管道
type MongoAggregate struct {
	builder  *MongoQueryBuilder
	pipeline []bson.M
}

// NewMongoAggregate 创建新的聚合管道
func NewMongoAggregate(collectionName string, connectionName ...string) (*MongoAggregate, error) {
	builder, err := MongoTable(collectionName, connectionName...)
	if err != nil {
		return nil, err
	}
	
	return &MongoAggregate{
		builder:  builder,
		pipeline: []bson.M{},
	}, nil
}

// Match 添加$match阶段
func (a *MongoAggregate) Match(filter bson.M) *MongoAggregate {
	a.pipeline = append(a.pipeline, bson.M{"$match": filter})
	return a
}

// Group 添加$group阶段
func (a *MongoAggregate) Group(group bson.M) *MongoAggregate {
	a.pipeline = append(a.pipeline, bson.M{"$group": group})
	return a
}

// Sort 添加$sort阶段
func (a *MongoAggregate) Sort(sort bson.M) *MongoAggregate {
	a.pipeline = append(a.pipeline, bson.M{"$sort": sort})
	return a
}

// Project 添加$project阶段
func (a *MongoAggregate) Project(project bson.M) *MongoAggregate {
	a.pipeline = append(a.pipeline, bson.M{"$project": project})
	return a
}

// Limit 添加$limit阶段
func (a *MongoAggregate) Limit(limit int64) *MongoAggregate {
	a.pipeline = append(a.pipeline, bson.M{"$limit": limit})
	return a
}

// Skip 添加$skip阶段
func (a *MongoAggregate) Skip(skip int64) *MongoAggregate {
	a.pipeline = append(a.pipeline, bson.M{"$skip": skip})
	return a
}

// Lookup 添加$lookup阶段（JOIN）
func (a *MongoAggregate) Lookup(from, localField, foreignField, as string) *MongoAggregate {
	lookup := bson.M{
		"$lookup": bson.M{
			"from":         from,
			"localField":   localField,
			"foreignField": foreignField,
			"as":           as,
		},
	}
	a.pipeline = append(a.pipeline, lookup)
	return a
}

// Unwind 添加$unwind阶段
func (a *MongoAggregate) Unwind(path string, preserveNullAndEmptyArrays ...bool) *MongoAggregate {
	unwind := bson.M{"$unwind": path}
	if len(preserveNullAndEmptyArrays) > 0 && preserveNullAndEmptyArrays[0] {
		unwind = bson.M{
			"$unwind": bson.M{
				"path":                       path,
				"preserveNullAndEmptyArrays": true,
			},
		}
	}
	a.pipeline = append(a.pipeline, unwind)
	return a
}

// AddStage 添加自定义聚合阶段
func (a *MongoAggregate) AddStage(stage bson.M) *MongoAggregate {
	a.pipeline = append(a.pipeline, stage)
	return a
}

// Execute 执行聚合管道
func (a *MongoAggregate) Execute() ([]map[string]interface{}, error) {
	if a.builder.collectionName == "" {
		return nil, NewError(ErrCodeInvalidParameter, "集合名称不能为空")
	}
	
	collection := a.builder.connection.Collection(a.builder.collectionName)
	if collection == nil {
		return nil, NewError(ErrCodeConnectionClosed, "MongoDB连接未建立")
	}
	
	// 执行聚合
	cursor, err := collection.Aggregate(a.builder.ctx, a.pipeline)
	if err != nil {
		return nil, WrapError(err, ErrCodeQueryFailed, "MongoDB聚合失败").
			WithContext("collection", a.builder.collectionName).
			WithContext("pipeline", a.pipeline)
	}
	defer cursor.Close(a.builder.ctx)
	
	// 解析结果
	var results []map[string]interface{}
	err = cursor.All(a.builder.ctx, &results)
	if err != nil {
		return nil, WrapError(err, ErrCodeQueryFailed, "MongoDB聚合结果解析失败").
			WithContext("collection", a.builder.collectionName)
	}
	
	return results, nil
}

// GetPipeline 获取聚合管道
func (a *MongoAggregate) GetPipeline() []bson.M {
	return a.pipeline
}
