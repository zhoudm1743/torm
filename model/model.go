package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/zhoudm1743/torm/db"
)

// Model 模型接口 - 定义所有模型必须实现的方法
type Model interface {
	// 表信息
	TableName() string
	PrimaryKey() string
	Connection() string

	// 数据操作
	Save() error
	Delete() error

	// 属性操作
	GetAttribute(key string) interface{}
	SetAttribute(key string, value interface{}) Model
	GetAttributes() map[string]interface{}
	SetAttributes(attributes map[string]interface{}) Model

	// 状态检查
	IsNew() bool
	IsDirty() bool
	GetDirty() map[string]interface{}

	// 事件钩子
	BeforeSave() error
	AfterSave() error
	BeforeCreate() error
	AfterCreate() error
	BeforeUpdate() error
	AfterUpdate() error
	BeforeDelete() error
	AfterDelete() error
}

// SoftDeleteModel 软删除模型接口
type SoftDeleteModel interface {
	Model
	SoftDelete() error
	ForceDelete() error
	Restore() error
	IsSoftDeleted() bool
	GetDeletedAtField() string
}

// TimestampModel 时间戳模型接口
type TimestampModel interface {
	Model
	GetCreatedAtField() string
	GetUpdatedAtField() string
	EnableTimestamps() Model
	DisableTimestamps() Model
}

// Tabler 自定义表名接口 - 模型可以实现此接口来自定义表名
type Tabler interface {
	TableName() string
}

// FieldScanner 字段扫描器接口 - 用于自定义类型
type FieldScanner interface {
	Scan(value interface{}) error
	Value() (driver.Value, error)
}

// JSONField JSON字段类型 - 自动序列化/反序列化
type JSONField map[string]interface{}

// Scan 实现 sql.Scanner 接口
func (j *JSONField) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan type %T into JSONField", value)
	}

	return json.Unmarshal(bytes, j)
}

// Value 实现 driver.Valuer 接口
func (j JSONField) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// JSONArray JSON数组类型
type JSONArray []interface{}

// Scan 实现 sql.Scanner 接口
func (j *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan type %T into JSONArray", value)
	}

	return json.Unmarshal(bytes, j)
}

// Value 实现 driver.Valuer 接口
func (j JSONArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// DeletedAt 软删除时间类型
type DeletedAt struct {
	Time  time.Time
	Valid bool
}

// Scan 实现 sql.Scanner 接口
func (d *DeletedAt) Scan(value interface{}) error {
	if value == nil {
		d.Time, d.Valid = time.Time{}, false
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		d.Time, d.Valid = v, true
	case string:
		if v == "" {
			d.Time, d.Valid = time.Time{}, false
		} else {
			t, err := time.Parse("2006-01-02 15:04:05", v)
			if err != nil {
				return err
			}
			d.Time, d.Valid = t, true
		}
	case []byte:
		if len(v) == 0 {
			d.Time, d.Valid = time.Time{}, false
		} else {
			t, err := time.Parse("2006-01-02 15:04:05", string(v))
			if err != nil {
				return err
			}
			d.Time, d.Valid = t, true
		}
	default:
		return fmt.Errorf("cannot scan type %T into DeletedAt", value)
	}
	return nil
}

// Value 实现 driver.Valuer 接口
func (d DeletedAt) Value() (driver.Value, error) {
	if !d.Valid {
		return nil, nil
	}
	return d.Time, nil
}

// modelRegistry 模型注册表 - 用于缓存模型元数据
var modelRegistry = struct {
	sync.RWMutex
	models map[reflect.Type]*ModelMeta
}{
	models: make(map[reflect.Type]*ModelMeta),
}

// ModelMeta 模型元数据
type ModelMeta struct {
	TableName    string
	PrimaryKey   string
	Connection   string
	Fields       []FieldMeta
	FieldMap     map[string]*FieldMeta
	Timestamps   bool
	SoftDeletes  bool
	CreatedAtCol string
	UpdatedAtCol string
	DeletedAtCol string
}

// FieldMeta 字段元数据
type FieldMeta struct {
	Name       string       // Go字段名
	Column     string       // 数据库列名
	Type       reflect.Type // 字段类型
	PrimaryKey bool         // 是否主键
	AutoInc    bool         // 是否自增
	Nullable   bool         // 是否可空
	Default    interface{}  // 默认值
	JSON       bool         // 是否JSON字段
	CreateTime bool         // 是否创建时间
	UpdateTime bool         // 是否更新时间
	SoftDelete bool         // 是否软删除字段
}

// GetModelMeta 获取模型元数据
func GetModelMeta(model interface{}) *ModelMeta {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 尝试从缓存获取
	modelRegistry.RLock()
	if meta, exists := modelRegistry.models[modelType]; exists {
		modelRegistry.RUnlock()
		return meta
	}
	modelRegistry.RUnlock()

	// 解析模型元数据
	meta := parseModelMeta(model)

	// 缓存元数据
	modelRegistry.Lock()
	modelRegistry.models[modelType] = meta
	modelRegistry.Unlock()

	return meta
}

// parseModelMeta 解析模型元数据
func parseModelMeta(model interface{}) *ModelMeta {
	meta := &ModelMeta{
		PrimaryKey:   "id",
		Connection:   "default",
		Timestamps:   true,
		SoftDeletes:  false,
		CreatedAtCol: "created_at",
		UpdatedAtCol: "updated_at",
		DeletedAtCol: "deleted_at",
		Fields:       make([]FieldMeta, 0),
		FieldMap:     make(map[string]*FieldMeta),
	}

	// 获取表名
	if tabler, ok := model.(Tabler); ok {
		meta.TableName = tabler.TableName()
	} else if getter, ok := model.(interface{ GetTableName() string }); ok {
		meta.TableName = getter.GetTableName()
	} else {
		// 从类型名推断表名
		modelType := reflect.TypeOf(model)
		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}
		meta.TableName = toSnakeCase(modelType.Name())
	}

	// 解析字段
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)

		// 跳过嵌入的BaseModel字段
		if field.Anonymous || field.Type.Name() == "BaseModel" {
			continue
		}

		fieldMeta := parseFieldMeta(field)
		meta.Fields = append(meta.Fields, fieldMeta)
		meta.FieldMap[fieldMeta.Column] = &meta.Fields[len(meta.Fields)-1]

		// 更新模型级别的配置
		if fieldMeta.PrimaryKey {
			meta.PrimaryKey = fieldMeta.Column
		}
		if fieldMeta.CreateTime {
			meta.CreatedAtCol = fieldMeta.Column
		}
		if fieldMeta.UpdateTime {
			meta.UpdatedAtCol = fieldMeta.Column
		}
		if fieldMeta.SoftDelete {
			meta.SoftDeletes = true
			meta.DeletedAtCol = fieldMeta.Column
		}
	}

	return meta
}

// parseFieldMeta 解析字段元数据
func parseFieldMeta(field reflect.StructField) FieldMeta {
	meta := FieldMeta{
		Name:   field.Name,
		Column: getColumnName(field),
		Type:   field.Type,
	}

	// 解析torm标签
	tormTag := field.Tag.Get("torm")
	if tormTag != "" {
		parts := strings.Split(tormTag, ",")
		for _, part := range parts {
			part = strings.TrimSpace(strings.ToLower(part))
			switch {
			case part == "primary_key" || part == "pk" || part == "primarykey":
				meta.PrimaryKey = true
			case part == "auto_increment" || part == "autoincrement" || part == "auto_inc":
				meta.AutoInc = true
			case part == "nullable" || part == "null":
				meta.Nullable = true
			case part == "json" || part == "is_json":
				meta.JSON = true
			case part == "auto_create_time" || part == "create_time" || part == "created_at":
				meta.CreateTime = true
			case part == "auto_update_time" || part == "update_time" || part == "updated_at":
				meta.UpdateTime = true
			case part == "soft_delete" || part == "soft_deletes" || part == "deleted_at":
				meta.SoftDelete = true
			case strings.HasPrefix(part, "default:"):
				meta.Default = strings.TrimPrefix(part, "default:")
			}
		}
	}

	return meta
}

// getColumnName 获取列名
func getColumnName(field reflect.StructField) string {
	// 检查torm标签中的列名
	if tormTag := field.Tag.Get("torm"); tormTag != "" {
		parts := strings.Split(tormTag, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "column:") {
				return strings.TrimPrefix(part, "column:")
			}
			if strings.HasPrefix(part, "db:") {
				return strings.TrimPrefix(part, "db:")
			}
		}
	}

	// 检查json标签
	if jsonTag := field.Tag.Get("json"); jsonTag != "" {
		parts := strings.Split(jsonTag, ",")
		if parts[0] != "" && parts[0] != "-" {
			return parts[0]
		}
	}

	// 使用字段名的蛇形命名
	return toSnakeCase(field.Name)
}

// InferTableName 推断表名 - 从结构体名称推断
func InferTableName(model interface{}) string {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	return toSnakeCase(modelType.Name())
}

// InferPrimaryKey 推断主键 - 默认为 "id"
func InferPrimaryKey(model interface{}) string {
	meta := GetModelMeta(model)
	return meta.PrimaryKey
}

// NewQueryForModel 为模型创建查询构建器
func NewQueryForModel(model interface{}, connectionName ...string) (*db.QueryBuilder, error) {
	meta := GetModelMeta(model)

	connName := meta.Connection
	if len(connectionName) > 0 && connectionName[0] != "" {
		connName = connectionName[0]
	}

	query, err := db.NewQueryBuilder(connName)
	if err != nil {
		return nil, fmt.Errorf("创建查询构建器失败: %w", err)
	}

	return query.From(meta.TableName).WithModel(model), nil
}
