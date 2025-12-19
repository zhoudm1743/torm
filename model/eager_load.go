package model

import (
	"fmt"
	"strings"
	"sync"

	"github.com/zhoudm1743/torm/db"
)

// EagerLoader 预加载器 - 处理关联数据的预加载
type EagerLoader struct {
	// 要预加载的关联
	relations []string
	// 关联约束
	constraints map[string][]func(*db.QueryBuilder) *db.QueryBuilder
	// 互斥锁
	mu sync.RWMutex
}

// NewEagerLoader 创建预加载器
func NewEagerLoader() *EagerLoader {
	return &EagerLoader{
		relations:   make([]string, 0),
		constraints: make(map[string][]func(*db.QueryBuilder) *db.QueryBuilder),
	}
}

// With 添加要预加载的关联
func (e *EagerLoader) With(relations ...string) *EagerLoader {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.relations = append(e.relations, relations...)
	return e
}

// WithConstraint 添加带约束的预加载
func (e *EagerLoader) WithConstraint(relation string, constraint func(*db.QueryBuilder) *db.QueryBuilder) *EagerLoader {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 添加关联（如果不存在）
	found := false
	for _, r := range e.relations {
		if r == relation || strings.HasPrefix(r, relation+".") {
			found = true
			break
		}
	}
	if !found {
		e.relations = append(e.relations, relation)
	}

	// 添加约束
	e.constraints[relation] = append(e.constraints[relation], constraint)
	return e
}

// GetRelations 获取要预加载的关联列表
func (e *EagerLoader) GetRelations() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.relations
}

// GetConstraints 获取关联的约束
func (e *EagerLoader) GetConstraints(relation string) []func(*db.QueryBuilder) *db.QueryBuilder {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.constraints[relation]
}

// HasRelations 检查是否有要预加载的关联
func (e *EagerLoader) HasRelations() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.relations) > 0
}

// Clear 清空预加载配置
func (e *EagerLoader) Clear() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.relations = make([]string, 0)
	e.constraints = make(map[string][]func(*db.QueryBuilder) *db.QueryBuilder)
}

// RelationRegistry 关联注册表 - 存储模型的关联定义
type RelationRegistry struct {
	definitions map[string]map[string]*RelationDefinition // modelType -> relationName -> definition
	mu          sync.RWMutex
}

// 全局关联注册表
var globalRelationRegistry = &RelationRegistry{
	definitions: make(map[string]map[string]*RelationDefinition),
}

// GetRelationRegistry 获取全局关联注册表
func GetRelationRegistry() *RelationRegistry {
	return globalRelationRegistry
}

// Register 注册关联定义
func (r *RelationRegistry) Register(modelType string, name string, def *RelationDefinition) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.definitions[modelType] == nil {
		r.definitions[modelType] = make(map[string]*RelationDefinition)
	}
	r.definitions[modelType][name] = def
}

// Get 获取关联定义
func (r *RelationRegistry) Get(modelType string, name string) *RelationDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if defs, ok := r.definitions[modelType]; ok {
		return defs[name]
	}
	return nil
}

// GetAll 获取模型的所有关联定义
func (r *RelationRegistry) GetAll(modelType string) map[string]*RelationDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if defs, ok := r.definitions[modelType]; ok {
		// 返回副本
		result := make(map[string]*RelationDefinition)
		for k, v := range defs {
			result[k] = v
		}
		return result
	}
	return nil
}

// ModelWithRelations 带关联的模型查询结果
type ModelWithRelations struct {
	Data      map[string]interface{}
	Relations map[string]interface{}
}

// LoadRelations 加载关联数据到模型结果
func LoadRelations(parent *BaseModel, results []map[string]interface{}, relations []string) ([]map[string]interface{}, error) {
	if len(results) == 0 || len(relations) == 0 {
		return results, nil
	}

	for _, relationPath := range relations {
		// 处理嵌套关联（如 "posts.comments"）
		parts := strings.SplitN(relationPath, ".", 2)
		relationName := parts[0]

		// 获取关联定义
		relation := getRelationFromModel(parent, relationName)
		if relation == nil {
			continue // 跳过未定义的关联
		}

		// 执行预加载
		var err error
		results, err = relation.EagerLoad(results, relationName)
		if err != nil {
			return nil, fmt.Errorf("预加载关联 %s 失败: %w", relationName, err)
		}

		// 处理嵌套关联
		if len(parts) > 1 {
			nestedRelation := parts[1]
			for i, result := range results {
				if relatedData, ok := result[relationName]; ok {
					switch data := relatedData.(type) {
					case []map[string]interface{}:
						// 一对多或多对多关联
						if len(data) > 0 {
							// 创建临时父模型用于嵌套加载
							tempParent := NewModel(parent.GetTableName())
							nestedResults, err := LoadRelations(tempParent, data, []string{nestedRelation})
							if err != nil {
								return nil, err
							}
							results[i][relationName] = nestedResults
						}
					case map[string]interface{}:
						// 一对一或反向关联
						if data != nil {
							tempParent := NewModel(parent.GetTableName())
							nestedResults, err := LoadRelations(tempParent, []map[string]interface{}{data}, []string{nestedRelation})
							if err != nil {
								return nil, err
							}
							if len(nestedResults) > 0 {
								results[i][relationName] = nestedResults[0]
							}
						}
					}
				}
			}
		}
	}

	return results, nil
}

// getRelationFromModel 从模型获取关联实例
func getRelationFromModel(parent *BaseModel, relationName string) RelationInterface {
	// 这里需要根据关联名称获取关联定义
	// 由于Go不支持动态方法调用，我们需要通过注册表或反射来获取

	// 首先尝试从注册表获取
	modelType := parent.GetTableName()
	def := GetRelationRegistry().Get(modelType, relationName)
	if def != nil && def.Builder != nil {
		return def.Builder(parent)
	}

	return nil
}

// RegisterRelation 注册模型关联
func RegisterRelation(modelType string, name string, relationType RelationType, builder func(parent *BaseModel) RelationInterface) {
	def := &RelationDefinition{
		Name:    name,
		Type:    relationType,
		Builder: builder,
	}
	GetRelationRegistry().Register(modelType, name, def)
}
