package db

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// AccessorProcessor 访问器处理器
type AccessorProcessor struct {
	modelType    reflect.Type
	getAccessors map[string]*regexp.Regexp
	setAccessors map[string]*regexp.Regexp
	methodCache  map[string]reflect.Method
	initialized  bool
}

// NewAccessorProcessor 创建访问器处理器
func NewAccessorProcessor(modelInstance interface{}) *AccessorProcessor {
	processor := &AccessorProcessor{
		getAccessors: make(map[string]*regexp.Regexp),
		setAccessors: make(map[string]*regexp.Regexp),
		methodCache:  make(map[string]reflect.Method),
		initialized:  false,
	}

	if modelInstance != nil {
		processor.modelType = reflect.TypeOf(modelInstance)
		if processor.modelType.Kind() == reflect.Ptr {
			processor.modelType = processor.modelType.Elem()
		}
		processor.initializeAccessors(modelInstance)
	}

	return processor
}

// initializeAccessors 初始化访问器缓存
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

	// 编译正则表达式
	getPattern := regexp.MustCompile(`^Get([A-Z][a-zA-Z0-9_]*)Attr$`)
	setPattern := regexp.MustCompile(`^Set([A-Z][a-zA-Z0-9_]*)Attr$`)

	// 扫描所有方法（包括指针接收者方法）
	for i := 0; i < modelType.NumMethod(); i++ {
		method := modelType.Method(i)
		methodName := method.Name

		// 检查获取器 Get[Name]Attr
		if matches := getPattern.FindStringSubmatch(methodName); len(matches) > 1 {
			attrName := camelToSnake(matches[1])
			ap.getAccessors[attrName] = getPattern
			ap.methodCache[methodName] = method
		}

		// 检查设置器 Set[Name]Attr
		if matches := setPattern.FindStringSubmatch(methodName); len(matches) > 1 {
			attrName := camelToSnake(matches[1])
			ap.setAccessors[attrName] = setPattern
			ap.methodCache[methodName] = method
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

// hasGetAccessor 检查是否有获取器
func (ap *AccessorProcessor) hasGetAccessor(key string) bool {
	if ap.getAccessors == nil {
		return false
	}
	_, exists := ap.getAccessors[key]
	return exists
}

// hasSetAccessor 检查是否有设置器
func (ap *AccessorProcessor) hasSetAccessor(key string) bool {
	if ap.setAccessors == nil {
		return false
	}
	_, exists := ap.setAccessors[key]
	return exists
}

// callGetAccessor 调用获取器
func (ap *AccessorProcessor) callGetAccessor(key string, value interface{}) interface{} {
	methodName := fmt.Sprintf("Get%sAttr", snakeToCamel(key))

	if method, exists := ap.methodCache[methodName]; exists {
		return ap.callMethod(method, value)
	}

	// 如果缓存中没有，尝试动态查找
	if ap.modelType != nil {
		if method, exists := ap.modelType.MethodByName(methodName); exists {
			return ap.callMethod(method, value)
		}
	}

	return value
}

// callSetAccessor 调用设置器
func (ap *AccessorProcessor) callSetAccessor(key string, value interface{}) interface{} {
	methodName := fmt.Sprintf("Set%sAttr", snakeToCamel(key))

	if method, exists := ap.methodCache[methodName]; exists {
		return ap.callMethod(method, value)
	}

	// 如果缓存中没有，尝试动态查找
	if ap.modelType != nil {
		if method, exists := ap.modelType.MethodByName(methodName); exists {
			return ap.callMethod(method, value)
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

// callMethod 调用反射方法
func (ap *AccessorProcessor) callMethod(method reflect.Method, value interface{}) interface{} {
	methodType := method.Type
	if methodType.NumIn() != 2 || methodType.NumOut() != 1 {
		return value
	}

	// 创建零值接收者 - 这里只是为了调用方法，不需要真实的实例
	receiverType := methodType.In(0)
	var receiver reflect.Value

	if receiverType.Kind() == reflect.Ptr {
		receiver = reflect.New(receiverType.Elem())
	} else {
		receiver = reflect.Zero(receiverType)
	}

	// 准备参数
	var param reflect.Value
	if value == nil {
		param = reflect.Zero(methodType.In(1))
	} else {
		param = reflect.ValueOf(value)
		if !param.Type().AssignableTo(methodType.In(1)) {
			if param.Type().ConvertibleTo(methodType.In(1)) {
				param = param.Convert(methodType.In(1))
			} else {
				return value
			}
		}
	}

	// 调用方法
	results := method.Func.Call([]reflect.Value{receiver, param})
	if len(results) > 0 {
		return results[0].Interface()
	}

	return value
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

// camelToSnake 驼峰转蛇形命名（增强版，支持连续大写字母）
func camelToSnake(str string) string {
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
