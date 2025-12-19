package model

import (
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/zhoudm1743/torm/db"
)

// ============================================================================
// 访问器和修改器支持 - 与GORM兼容的字段获取器/设置器
// ============================================================================

// AccessorMutator 访问器/修改器接口
// 模型可以实现此接口来自定义字段的获取和设置行为
type AccessorMutator interface {
	// HasAccessor 检查是否有指定字段的访问器
	HasAccessor(field string) bool
	// HasMutator 检查是否有指定字段的修改器
	HasMutator(field string) bool
	// GetAccessor 获取字段值（通过访问器）
	GetAccessor(field string, value interface{}) interface{}
	// SetMutator 设置字段值（通过修改器）
	SetMutator(field string, value interface{}) interface{}
}

// ModelAccessorProcessor 模型访问器处理器
// 封装了db.AccessorProcessor，提供模型层的访问器支持
type ModelAccessorProcessor struct {
	processor *db.AccessorProcessor
	model     interface{}
}

// NewModelAccessorProcessor 创建模型访问器处理器
func NewModelAccessorProcessor(model interface{}) *ModelAccessorProcessor {
	return &ModelAccessorProcessor{
		processor: db.NewAccessorProcessor(model),
		model:     model,
	}
}

// ProcessGetData 处理获取数据（应用访问器）
func (p *ModelAccessorProcessor) ProcessGetData(data map[string]interface{}) map[string]interface{} {
	if p.processor == nil {
		return data
	}
	return p.processor.ProcessData(data)
}

// ProcessSetData 处理设置数据（应用修改器）
func (p *ModelAccessorProcessor) ProcessSetData(data map[string]interface{}) map[string]interface{} {
	if p.processor == nil {
		return data
	}
	return p.processor.ProcessSetData(data)
}

// ============================================================================
// 字段权限控制 - 与GORM兼容
// ============================================================================

// FieldPermission 字段权限
type FieldPermission int

const (
	// PermissionReadWrite 可读可写（默认）
	PermissionReadWrite FieldPermission = iota
	// PermissionReadOnly 只读
	PermissionReadOnly
	// PermissionWriteOnly 只写
	PermissionWriteOnly
	// PermissionCreateOnly 只在创建时可写
	PermissionCreateOnly
	// PermissionUpdateOnly 只在更新时可写
	PermissionUpdateOnly
	// PermissionIgnore 忽略字段
	PermissionIgnore
)

// FieldPermissionInfo 字段权限信息
type FieldPermissionInfo struct {
	Field      string          // 字段名
	Permission FieldPermission // 权限
	Readable   bool            // 是否可读
	Writable   bool            // 是否可写
	CreateOnly bool            // 只在创建时可写
	UpdateOnly bool            // 只在更新时可写
}

// 全局字段权限缓存
var (
	fieldPermissionCache = make(map[reflect.Type]map[string]*FieldPermissionInfo)
	permissionCacheMutex sync.RWMutex
)

// GetFieldPermissions 获取模型的字段权限信息
func GetFieldPermissions(model interface{}) map[string]*FieldPermissionInfo {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 尝试从缓存获取
	permissionCacheMutex.RLock()
	if perms, exists := fieldPermissionCache[modelType]; exists {
		permissionCacheMutex.RUnlock()
		return perms
	}
	permissionCacheMutex.RUnlock()

	// 解析字段权限
	perms := parseFieldPermissions(model)

	// 缓存结果
	permissionCacheMutex.Lock()
	fieldPermissionCache[modelType] = perms
	permissionCacheMutex.Unlock()

	return perms
}

// parseFieldPermissions 解析字段权限
func parseFieldPermissions(model interface{}) map[string]*FieldPermissionInfo {
	perms := make(map[string]*FieldPermissionInfo)

	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)

		// 跳过嵌入字段
		if field.Anonymous {
			continue
		}

		// 获取列名
		columnName := getColumnNameFromField(field)

		// 默认权限：可读可写
		info := &FieldPermissionInfo{
			Field:      columnName,
			Permission: PermissionReadWrite,
			Readable:   true,
			Writable:   true,
			CreateOnly: false,
			UpdateOnly: false,
		}

		// 解析torm标签中的权限设置
		tormTag := field.Tag.Get("torm")
		if tormTag != "" {
			parsePermissionFromTag(tormTag, info)
		}

		// 解析gorm标签中的权限设置（兼容GORM）
		gormTag := field.Tag.Get("gorm")
		if gormTag != "" {
			parsePermissionFromTag(gormTag, info)
		}

		perms[columnName] = info
	}

	return perms
}

// parsePermissionFromTag 从标签解析权限
func parsePermissionFromTag(tag string, info *FieldPermissionInfo) {
	// 使用分号分隔（GORM格式）或逗号分隔（torm格式）
	var parts []string
	if strings.Contains(tag, ";") {
		parts = strings.Split(tag, ";")
	} else {
		parts = strings.Split(tag, ",")
	}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// 检查忽略标记
		if part == "-" {
			info.Permission = PermissionIgnore
			info.Readable = false
			info.Writable = false
			return
		}

		// 检查读取权限 ->
		if strings.HasPrefix(part, "->") {
			if part == "->:false" || part == "->false" {
				info.Readable = false
			}
			continue
		}

		// 检查写入权限 <-
		if strings.HasPrefix(part, "<-") {
			switch part {
			case "<-:false", "<-false":
				info.Writable = false
			case "<-:create", "<-create":
				info.CreateOnly = true
				info.Permission = PermissionCreateOnly
			case "<-:update", "<-update":
				info.UpdateOnly = true
				info.Permission = PermissionUpdateOnly
			}
			continue
		}

		// 检查其他权限标记
		partLower := strings.ToLower(part)
		switch partLower {
		case "readonly", "read_only":
			info.Permission = PermissionReadOnly
			info.Writable = false
		case "writeonly", "write_only":
			info.Permission = PermissionWriteOnly
			info.Readable = false
		case "createonly", "create_only":
			info.Permission = PermissionCreateOnly
			info.CreateOnly = true
		case "updateonly", "update_only":
			info.Permission = PermissionUpdateOnly
			info.UpdateOnly = true
		case "ignore", "ignored":
			info.Permission = PermissionIgnore
			info.Readable = false
			info.Writable = false
		}
	}
}

// ============================================================================
// 访问器方法命名约定
// ============================================================================

// 访问器方法命名约定：
// - 获取器: Get{FieldName}Attr(value interface{}) interface{}
// - 设置器: Set{FieldName}Attr(value interface{}) interface{}
//
// 示例:
// type User struct {
//     BaseModel
//     Name     string `json:"name"`
//     Password string `json:"password"`
// }
//
// // 获取器 - 获取name时自动调用
// func (u *User) GetNameAttr(value interface{}) interface{} {
//     if s, ok := value.(string); ok {
//         return strings.ToUpper(s)
//     }
//     return value
// }
//
// // 设置器 - 设置password时自动调用
// func (u *User) SetPasswordAttr(value interface{}) interface{} {
//     if s, ok := value.(string); ok {
//         return hashPassword(s)
//     }
//     return value
// }

// AccessorMethodPattern 访问器方法模式
var (
	getterPattern = regexp.MustCompile(`^Get([A-Z][a-zA-Z0-9_]*)Attr$`)
	setterPattern = regexp.MustCompile(`^Set([A-Z][a-zA-Z0-9_]*)Attr$`)
)

// FindAccessorMethods 查找模型的访问器方法
func FindAccessorMethods(model interface{}) (getters, setters map[string]string) {
	getters = make(map[string]string)
	setters = make(map[string]string)

	modelType := reflect.TypeOf(model)
	if modelType == nil {
		return
	}

	// 扫描所有方法
	for i := 0; i < modelType.NumMethod(); i++ {
		method := modelType.Method(i)
		methodName := method.Name

		// 检查获取器
		if matches := getterPattern.FindStringSubmatch(methodName); len(matches) > 1 {
			fieldName := toSnakeCase(matches[1])
			getters[fieldName] = methodName
		}

		// 检查设置器
		if matches := setterPattern.FindStringSubmatch(methodName); len(matches) > 1 {
			fieldName := toSnakeCase(matches[1])
			setters[fieldName] = methodName
		}
	}

	return
}

// ============================================================================
// BaseModel 访问器支持方法
// ============================================================================

// GetAttributeWithAccessor 获取属性值（应用访问器）
func (m *BaseModel) GetAttributeWithAccessor(model interface{}, key string) interface{} {
	value := m.GetAttribute(key)
	processor := db.NewAccessorProcessor(model)
	result := processor.ProcessData(map[string]interface{}{key: value})
	if v, exists := result[key]; exists {
		return v
	}
	return value
}

// GetAttributesWithAccessor 获取所有属性（应用访问器）
func (m *BaseModel) GetAttributesWithAccessor(model interface{}) map[string]interface{} {
	processor := db.NewAccessorProcessor(model)
	return processor.ProcessData(m.GetAttributes())
}

// FilterWritableFields 过滤可写字段（用于插入/更新）
func (m *BaseModel) FilterWritableFields(model interface{}, data map[string]interface{}, isCreate bool) map[string]interface{} {
	perms := GetFieldPermissions(model)
	result := make(map[string]interface{})

	for key, value := range data {
		info, exists := perms[key]
		if !exists {
			// 没有权限信息，默认可写
			result[key] = value
			continue
		}

		// 检查是否可写
		if !info.Writable {
			continue
		}

		// 检查创建/更新限制
		if isCreate && info.UpdateOnly {
			continue
		}
		if !isCreate && info.CreateOnly {
			continue
		}

		result[key] = value
	}

	return result
}

// FilterReadableFields 过滤可读字段（用于查询结果）
func (m *BaseModel) FilterReadableFields(model interface{}, data map[string]interface{}) map[string]interface{} {
	perms := GetFieldPermissions(model)
	result := make(map[string]interface{})

	for key, value := range data {
		info, exists := perms[key]
		if !exists {
			// 没有权限信息，默认可读
			result[key] = value
			continue
		}

		// 检查是否可读
		if info.Readable {
			result[key] = value
		}
	}

	return result
}
