package db

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 全局访问器缓存 - 优化重复创建处理器
var (
	globalAccessorCache = make(map[reflect.Type]*AccessorProcessor)
	accessorCacheMutex  sync.RWMutex

	// 预编译的正则表达式 - 避免重复编译
	getPatternCache = regexp.MustCompile(`^Get([A-Z][a-zA-Z0-9_]*)Attr$`)
	setPatternCache = regexp.MustCompile(`^Set([A-Z][a-zA-Z0-9_]*)Attr$`)
)

// AccessorProcessor 访问器处理器 - 性能优化版本
type AccessorProcessor struct {
	modelType    reflect.Type
	getAccessors map[string]string // 简化存储，只存储属性名->方法名映射
	setAccessors map[string]string // 简化存储，只存储属性名->方法名映射
	methodCache  map[string]reflect.Method
	initialized  bool

	// 性能优化：缓存常用反射结果
	modelValue reflect.Value
	isPointer  bool
}

// NewAccessorProcessor 创建访问器处理器 - 性能优化版本
func NewAccessorProcessor(modelInstance interface{}) *AccessorProcessor {
	if modelInstance == nil {
		return &AccessorProcessor{
			getAccessors: make(map[string]string),
			setAccessors: make(map[string]string),
			methodCache:  make(map[string]reflect.Method),
			initialized:  false,
		}
	}

	modelType := reflect.TypeOf(modelInstance)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 尝试从全局缓存获取
	accessorCacheMutex.RLock()
	if cached, exists := globalAccessorCache[modelType]; exists {
		accessorCacheMutex.RUnlock()
		// 创建一个新实例，共享访问器信息但独立状态
		return &AccessorProcessor{
			modelType:    cached.modelType,
			getAccessors: cached.getAccessors,
			setAccessors: cached.setAccessors,
			methodCache:  cached.methodCache,
			initialized:  true,
			isPointer:    reflect.TypeOf(modelInstance).Kind() == reflect.Ptr,
		}
	}
	accessorCacheMutex.RUnlock()

	// 不在缓存中，创建新的处理器
	processor := &AccessorProcessor{
		modelType:    modelType,
		getAccessors: make(map[string]string),
		setAccessors: make(map[string]string),
		methodCache:  make(map[string]reflect.Method),
		initialized:  false,
		isPointer:    reflect.TypeOf(modelInstance).Kind() == reflect.Ptr,
	}

	processor.initializeAccessors(modelInstance)

	// 加入全局缓存
	accessorCacheMutex.Lock()
	globalAccessorCache[modelType] = processor
	accessorCacheMutex.Unlock()

	return processor
}

// initializeAccessors 初始化访问器缓存 - 性能优化版本
func (ap *AccessorProcessor) initializeAccessors(modelInstance interface{}) {
	if ap.initialized {
		return
	}

	// 获取类型，需要检查指针类型的方法
	modelType := reflect.TypeOf(modelInstance)
	if modelType == nil {
		ap.initialized = true
		return
	}

	// 使用预编译的正则表达式
	getPattern := getPatternCache
	setPattern := setPatternCache

	// 扫描所有方法（包括指针接收者方法）
	methodCount := modelType.NumMethod()
	for i := 0; i < methodCount; i++ {
		method := modelType.Method(i)
		methodName := method.Name

		// 性能优化：先检查方法名模式，避免不必要的正则匹配
		if len(methodName) > 4 && strings.HasSuffix(methodName, "Attr") {
			// 检查获取器 Get[Name]Attr
			if strings.HasPrefix(methodName, "Get") && len(methodName) > 7 {
				if matches := getPattern.FindStringSubmatch(methodName); len(matches) > 1 {
					attrName := camelToSnakeOptimized(matches[1])
					ap.getAccessors[attrName] = methodName // 直接存储方法名
					ap.methodCache[methodName] = method
				}
			}

			// 检查设置器 Set[Name]Attr
			if strings.HasPrefix(methodName, "Set") && len(methodName) > 7 {
				if matches := setPattern.FindStringSubmatch(methodName); len(matches) > 1 {
					attrName := camelToSnakeOptimized(matches[1])
					ap.setAccessors[attrName] = methodName // 直接存储方法名
					ap.methodCache[methodName] = method
				}
			}
		}
	}

	ap.initialized = true
}

// ProcessData 处理单条记录数据，应用访问器
func (ap *AccessorProcessor) ProcessData(data map[string]interface{}) map[string]interface{} {
	if data == nil || !ap.initialized {
		return data
	}

	result := make(map[string]interface{})
	for key, value := range data {
		// 先处理基本类型转换
		processedValue := ap.processValue(value)

		// 检查是否有访问器，如果有则调用
		if ap.hasGetAccessor(key) {
			result[key] = ap.callGetAccessor(key, processedValue)
		} else {
			result[key] = processedValue
		}
	}
	return result
}

// ProcessDataSlice 处理多条记录数据，应用访问器
func (ap *AccessorProcessor) ProcessDataSlice(dataSlice []map[string]interface{}) []map[string]interface{} {
	if dataSlice == nil || !ap.initialized {
		return dataSlice
	}

	result := make([]map[string]interface{}, len(dataSlice))
	for i, data := range dataSlice {
		result[i] = ap.ProcessData(data)
	}
	return result
}

// hasGetAccessor 检查是否有获取器 - 性能优化版本
func (ap *AccessorProcessor) hasGetAccessor(key string) bool {
	if ap.getAccessors == nil {
		return false
	}
	_, exists := ap.getAccessors[key]
	return exists
}

// hasSetAccessor 检查是否有设置器 - 性能优化版本
func (ap *AccessorProcessor) hasSetAccessor(key string) bool {
	if ap.setAccessors == nil {
		return false
	}
	_, exists := ap.setAccessors[key]
	return exists
}

// callGetAccessor 调用获取器 - 性能优化版本
func (ap *AccessorProcessor) callGetAccessor(key string, value interface{}) interface{} {
	// 直接从缓存获取方法名
	if methodName, exists := ap.getAccessors[key]; exists {
		if method, methodExists := ap.methodCache[methodName]; methodExists {
			return ap.callMethodOptimized(method, value)
		}
	}

	return value
}

// callSetAccessor 调用设置器 - 性能优化版本
func (ap *AccessorProcessor) callSetAccessor(key string, value interface{}) interface{} {
	// 直接从缓存获取方法名
	if methodName, exists := ap.setAccessors[key]; exists {
		if method, methodExists := ap.methodCache[methodName]; methodExists {
			return ap.callMethodOptimized(method, value)
		}
	}

	return value
}

// ProcessSetData 处理要设置的数据，应用设置器
func (ap *AccessorProcessor) ProcessSetData(data map[string]interface{}) map[string]interface{} {
	if data == nil || !ap.initialized {
		return data
	}

	result := make(map[string]interface{})
	for key, value := range data {
		// 检查是否有设置器，如果有则调用
		if ap.hasSetAccessor(key) {
			result[key] = ap.callSetAccessor(key, value)
		} else {
			result[key] = value
		}
	}
	return result
}

// callMethod 调用反射方法 - 保持向后兼容
func (ap *AccessorProcessor) callMethod(method reflect.Method, value interface{}) interface{} {
	return ap.callMethodOptimized(method, value)
}

// processValue 处理 []byte 和其他类型的值
func (ap *AccessorProcessor) processValue(value interface{}) interface{} {
	if value == nil {
		return nil
	}

	// 处理 []byte 类型
	if bytes, ok := value.([]byte); ok {
		return ap.processBytesValue(bytes)
	}

	return value
}

// processBytesValue 智能处理 []byte 数据
func (ap *AccessorProcessor) processBytesValue(bytes []byte) interface{} {
	if len(bytes) == 0 {
		return ""
	}

	str := string(bytes)

	// 1. 尝试解析为数字
	if i, err := strconv.ParseInt(str, 10, 64); err == nil {
		// 检查是否为合理的整数范围
		if i >= -2147483648 && i <= 2147483647 {
			return int(i)
		}
		return i
	}

	// 2. 尝试解析为浮点数
	if f, err := strconv.ParseFloat(str, 64); err == nil {
		return f
	}

	// 3. 尝试解析为布尔值
	if b, err := strconv.ParseBool(str); err == nil {
		return b
	}

	// 4. 尝试解析为时间
	if t, err := time.Parse("2006-01-02 15:04:05", str); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02", str); err == nil {
		return t
	}
	if t, err := time.Parse(time.RFC3339, str); err == nil {
		return t
	}

	// 5. 尝试解析为JSON
	var jsonData interface{}
	if err := json.Unmarshal(bytes, &jsonData); err == nil {
		return jsonData
	}

	// 6. 返回字符串
	return str
}

// callMethodOptimized 优化版本的方法调用
func (ap *AccessorProcessor) callMethodOptimized(method reflect.Method, value interface{}) interface{} {
	methodType := method.Type
	if methodType.NumIn() != 2 || methodType.NumOut() != 1 {
		return value
	}

	// 创建零值接收者 - 优化版本
	receiverType := methodType.In(0)
	var receiver reflect.Value

	if receiverType.Kind() == reflect.Ptr {
		receiver = reflect.New(receiverType.Elem())
	} else {
		receiver = reflect.Zero(receiverType)
	}

	// 准备参数 - 优化版本，减少反射调用
	var param reflect.Value
	if value == nil {
		param = reflect.Zero(methodType.In(1))
	} else {
		valueType := reflect.TypeOf(value)
		paramType := methodType.In(1)

		// 性能优化：直接类型匹配，避免转换
		if valueType == paramType {
			param = reflect.ValueOf(value)
		} else if valueType.AssignableTo(paramType) {
			param = reflect.ValueOf(value)
		} else if valueType.ConvertibleTo(paramType) {
			param = reflect.ValueOf(value).Convert(paramType)
		} else {
			return value
		}
	}

	// 调用方法
	results := method.Func.Call([]reflect.Value{receiver, param})
	if len(results) > 0 {
		return results[0].Interface()
	}

	return value
}

// camelToSnakeOptimized 优化版本的驼峰转蛇形命名
func camelToSnakeOptimized(str string) string {
	if str == "" {
		return ""
	}

	// 预分配适当大小的builder
	var result strings.Builder
	result.Grow(len(str) + len(str)/3) // 预估大小

	runes := []rune(str)
	runeCount := len(runes)

	for i, r := range runes {
		// 当前字符是大写字母
		if r >= 'A' && r <= 'Z' {
			// 需要添加下划线的条件（优化判断逻辑）
			if i > 0 {
				prevIsLower := runes[i-1] >= 'a' && runes[i-1] <= 'z'
				nextIsLower := i+1 < runeCount && runes[i+1] >= 'a' && runes[i+1] <= 'z'
				if prevIsLower || nextIsLower {
					result.WriteRune('_')
				}
			}
			result.WriteRune(r - 'A' + 'a') // 转为小写
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// camelToSnake 驼峰转蛇形命名（保持向后兼容）
func camelToSnake(str string) string {
	return camelToSnakeOptimized(str)
}

// snakeToCamel 蛇形转驼峰命名（增强版，智能处理缩写）
func snakeToCamel(str string) string {
	if str == "" {
		return ""
	}

	parts := strings.Split(str, "_")
	var result strings.Builder

	for _, part := range parts {
		if len(part) == 0 {
			continue
		}

		// 特殊处理常见缩写
		switch strings.ToLower(part) {
		case "id":
			result.WriteString("ID")
		case "url":
			result.WriteString("URL")
		case "http":
			result.WriteString("HTTP")
		case "https":
			result.WriteString("HTTPS")
		case "api":
			result.WriteString("API")
		case "json":
			result.WriteString("JSON")
		case "xml":
			result.WriteString("XML")
		case "html":
			result.WriteString("HTML")
		case "css":
			result.WriteString("CSS")
		case "js":
			result.WriteString("JS")
		case "sql":
			result.WriteString("SQL")
		case "db":
			result.WriteString("DB")
		case "tcp":
			result.WriteString("TCP")
		case "udp":
			result.WriteString("UDP")
		case "ip":
			result.WriteString("IP")
		case "uuid":
			result.WriteString("UUID")
		case "md5":
			result.WriteString("MD5")
		case "sha":
			result.WriteString("SHA")
		case "rsa":
			result.WriteString("RSA")
		case "aes":
			result.WriteString("AES")
		case "icbc":
			result.WriteString("ICBC")
		case "cmb":
			result.WriteString("CMB")
		case "abc":
			result.WriteString("ABC")
		case "ccb":
			result.WriteString("CCB")
		case "boc":
			result.WriteString("BOC")
		default:
			// 普通单词首字母大写
			result.WriteString(strings.ToUpper(string(part[0])))
			if len(part) > 1 {
				result.WriteString(strings.ToLower(part[1:]))
			}
		}
	}

	return result.String()
}
