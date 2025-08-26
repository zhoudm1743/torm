package model

import (
	"fmt"
	"reflect"
	"strings"
)

// AttributeAccessorInterface 属性访问器接口
type AttributeAccessorInterface interface {
	// CallGetAccessor 调用属性访问器（获取器）
	CallGetAccessor(key string, value interface{}) interface{}
	// CallSetAccessor 调用属性修改器（设置器）
	CallSetAccessor(key string, value interface{}) interface{}
	// HasGetAccessor 检查是否有获取器
	HasGetAccessor(key string) bool
	// HasSetAccessor 检查是否有设置器
	HasSetAccessor(key string) bool
}

// AttributeCache 属性访问器缓存
type AttributeCache struct {
	getAccessors map[string]reflect.Method
	setAccessors map[string]reflect.Method
	initialized  bool
}

// newAttributeCache 创建属性缓存
func newAttributeCache() *AttributeCache {
	return &AttributeCache{
		getAccessors: make(map[string]reflect.Method),
		setAccessors: make(map[string]reflect.Method),
		initialized:  false,
	}
}

// initializeAccessors 初始化访问器缓存
func (ac *AttributeCache) initializeAccessors(modelInstance interface{}) {
	if ac.initialized {
		return
	}

	modelType := reflect.TypeOf(modelInstance)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 扫描所有方法，查找访问器
	for i := 0; i < modelType.NumMethod(); i++ {
		method := modelType.Method(i)
		methodName := method.Name

		// 检查是否是获取器 (GetXxxAttr)
		if strings.HasPrefix(methodName, "Get") && strings.HasSuffix(methodName, "Attr") {
			// 提取属性名：GetStatusAttr -> status
			attrName := extractAttributeName(methodName, "Get", "Attr")
			if attrName != "" {
				ac.getAccessors[attrName] = method
			}
		}

		// 检查是否是设置器 (SetXxxAttr)
		if strings.HasPrefix(methodName, "Set") && strings.HasSuffix(methodName, "Attr") {
			// 提取属性名：SetStatusAttr -> status
			attrName := extractAttributeName(methodName, "Set", "Attr")
			if attrName != "" {
				ac.setAccessors[attrName] = method
			}
		}
	}

	ac.initialized = true
}

// extractAttributeName 从方法名提取属性名
func extractAttributeName(methodName, prefix, suffix string) string {
	if !strings.HasPrefix(methodName, prefix) || !strings.HasSuffix(methodName, suffix) {
		return ""
	}

	// 移除前缀和后缀
	attrPart := methodName[len(prefix) : len(methodName)-len(suffix)]
	if attrPart == "" {
		return ""
	}

	// 转换为蛇形命名：StatusName -> status_name
	return attrToSnakeCase(attrPart)
}

// attrToSnakeCase 转换为蛇形命名（避免重复定义）
func attrToSnakeCase(str string) string {
	if str == "" {
		return ""
	}

	var result strings.Builder
	runes := []rune(str)

	for i, r := range runes {
		// 当前字符是大写字母
		if r >= 'A' && r <= 'Z' {
			// 需要添加下划线的条件：
			// 1. 不是第一个字符
			// 2. 前一个字符是小写字母，或者
			// 3. 当前字符后面跟着小写字母（处理连续大写的情况，如 HTMLParser -> html_parser）
			if i > 0 && ((runes[i-1] >= 'a' && runes[i-1] <= 'z') || // 前一个是小写
				(i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z')) { // 后一个是小写
				result.WriteRune('_')
			}
			result.WriteRune(r - 'A' + 'a') // 转为小写
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// toPascalCase 转换为帕斯卡命名
func toPascalCase(str string) string {
	if str == "" {
		return ""
	}

	parts := strings.Split(str, "_")
	var result strings.Builder

	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(string(part[0])))
			if len(part) > 1 {
				result.WriteString(strings.ToLower(part[1:]))
			}
		}
	}

	return result.String()
}

// GetAttributeWithAccessor 获取属性（支持访问器）
func (m *BaseModel) GetAttributeWithAccessor(key string) interface{} {
	// 获取原始值
	rawValue := m.attributes[key]

	// 检查是否是支持访问器的模型
	if accessorModel, ok := interface{}(m).(AttributeAccessorInterface); ok {
		if accessorModel.HasGetAccessor(key) {
			return accessorModel.CallGetAccessor(key, rawValue)
		}
	}

	// 如果没有访问器，尝试反射调用
	return m.callGetAccessorByReflection(key, rawValue)
}

// SetAttributeWithAccessor 设置属性（支持修改器）
func (m *BaseModel) SetAttributeWithAccessor(key string, value interface{}) *BaseModel {
	// 检查是否是支持访问器的模型
	if accessorModel, ok := interface{}(m).(AttributeAccessorInterface); ok {
		if accessorModel.HasSetAccessor(key) {
			processedValue := accessorModel.CallSetAccessor(key, value)
			m.attributes[key] = processedValue
			return m
		}
	}

	// 如果没有修改器，尝试反射调用
	processedValue := m.callSetAccessorByReflection(key, value)
	m.attributes[key] = processedValue
	return m
}

// callGetAccessorByReflection 通过反射调用获取器
func (m *BaseModel) callGetAccessorByReflection(key string, value interface{}) interface{} {
	// 构造方法名：status -> GetStatusAttr
	methodName := fmt.Sprintf("Get%sAttr", toPascalCase(key))

	modelValue := reflect.ValueOf(m)
	method := modelValue.MethodByName(methodName)

	if !method.IsValid() {
		// 没有找到访问器，返回原始值
		return value
	}

	// 检查方法签名
	methodType := method.Type()
	if methodType.NumIn() != 1 || methodType.NumOut() != 1 {
		// 方法签名不正确，返回原始值
		return value
	}

	// 调用访问器方法
	var param reflect.Value
	if value == nil {
		param = reflect.Zero(methodType.In(0))
	} else {
		param = reflect.ValueOf(value)
		// 如果类型不匹配，尝试转换
		if !param.Type().AssignableTo(methodType.In(0)) {
			if param.Type().ConvertibleTo(methodType.In(0)) {
				param = param.Convert(methodType.In(0))
			} else {
				// 无法转换，返回原始值
				return value
			}
		}
	}

	result := method.Call([]reflect.Value{param})
	if len(result) > 0 {
		return result[0].Interface()
	}

	return value
}

// callSetAccessorByReflection 通过反射调用设置器
func (m *BaseModel) callSetAccessorByReflection(key string, value interface{}) interface{} {
	// 构造方法名：status -> SetStatusAttr
	methodName := fmt.Sprintf("Set%sAttr", toPascalCase(key))

	modelValue := reflect.ValueOf(m)
	method := modelValue.MethodByName(methodName)

	if !method.IsValid() {
		// 没有找到修改器，返回原始值
		return value
	}

	// 检查方法签名
	methodType := method.Type()
	if methodType.NumIn() != 1 || methodType.NumOut() != 1 {
		// 方法签名不正确，返回原始值
		return value
	}

	// 调用修改器方法
	var param reflect.Value
	if value == nil {
		param = reflect.Zero(methodType.In(0))
	} else {
		param = reflect.ValueOf(value)
		// 如果类型不匹配，尝试转换
		if !param.Type().AssignableTo(methodType.In(0)) {
			if param.Type().ConvertibleTo(methodType.In(0)) {
				param = param.Convert(methodType.In(0))
			} else {
				// 无法转换，返回原始值
				return value
			}
		}
	}

	result := method.Call([]reflect.Value{param})
	if len(result) > 0 {
		return result[0].Interface()
	}

	return value
}

// GetAttributesWithAccessors 获取所有属性（支持访问器）
func (m *BaseModel) GetAttributesWithAccessors() map[string]interface{} {
	result := make(map[string]interface{})

	// 检查是否重写了 GetAttributes 方法
	if hasCustomGetAttributes(m) {
		// 调用自定义的 GetAttributes 方法
		customAttrs := m.GetAttributes()
		// 对每个属性应用访问器
		for key := range customAttrs {
			result[key] = m.GetAttributeWithAccessor(key)
		}
		return result
	}

	// 使用默认实现，对每个属性应用访问器
	for key := range m.attributes {
		result[key] = m.GetAttributeWithAccessor(key)
	}

	return result
}

// SetAttributesWithAccessors 批量设置属性（支持修改器）
func (m *BaseModel) SetAttributesWithAccessors(attributes map[string]interface{}) *BaseModel {
	// 检查是否重写了 SetAttributes 方法
	if hasCustomSetAttributes(m) {
		// 先应用修改器，然后调用自定义的 SetAttributes 方法
		processedAttrs := make(map[string]interface{})
		for key, value := range attributes {
			processedValue := m.callSetAccessorByReflection(key, value)
			processedAttrs[key] = processedValue
		}
		m.SetAttributes(processedAttrs)
		return m
	}

	// 使用默认实现，对每个属性应用修改器
	for key, value := range attributes {
		m.SetAttributeWithAccessor(key, value)
	}

	return m
}

// hasCustomGetAttributes 检查是否有自定义的 GetAttributes 方法
func hasCustomGetAttributes(m *BaseModel) bool {
	modelType := reflect.TypeOf(m)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 检查是否有同名方法（非 BaseModel 的方法）
	_, exists := modelType.MethodByName("GetAttributes")
	if !exists {
		return false
	}

	// 检查方法是否是在当前类型中定义的（而不是继承的）
	return true // 简化检查，假设存在就是自定义的
}

// hasCustomSetAttributes 检查是否有自定义的 SetAttributes 方法
func hasCustomSetAttributes(m *BaseModel) bool {
	modelType := reflect.TypeOf(m)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 检查是否有同名方法（非 BaseModel 的方法）
	_, exists := modelType.MethodByName("SetAttributes")
	if !exists {
		return false
	}

	// 检查方法是否是在当前类型中定义的（而不是继承的）
	return true // 简化检查，假设存在就是自定义的
}

// GetAttr 简写方法 - 获取属性
func (m *BaseModel) GetAttr(key string) interface{} {
	return m.GetAttributeWithAccessor(key)
}

// SetAttr 简写方法 - 设置属性
func (m *BaseModel) SetAttr(key string, value interface{}) *BaseModel {
	return m.SetAttributeWithAccessor(key, value)
}

// GetAttrs 简写方法 - 获取所有属性
func (m *BaseModel) GetAttrs() map[string]interface{} {
	return m.GetAttributesWithAccessors()
}

// SetAttrs 简写方法 - 批量设置属性
func (m *BaseModel) SetAttrs(attributes map[string]interface{}) *BaseModel {
	return m.SetAttributesWithAccessors(attributes)
}
