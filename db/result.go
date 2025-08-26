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

// Result 查询结果封装类型
type Result struct {
	data      map[string]interface{}
	modelType reflect.Type
	accessors *AccessorCache
}

// AccessorCache 访问器缓存
type AccessorCache struct {
	getAccessors map[string]*regexp.Regexp
	setAccessors map[string]*regexp.Regexp
	methodCache  map[string]reflect.Method
	initialized  bool
}

// NewResult 创建新的结果实例
func NewResult(data map[string]interface{}, modelInstance interface{}) *Result {
	result := &Result{
		data:      data,
		accessors: newAccessorCache(),
	}

	if modelInstance != nil {
		result.modelType = reflect.TypeOf(modelInstance)
		if result.modelType.Kind() == reflect.Ptr {
			result.modelType = result.modelType.Elem()
		}
		result.initializeAccessors(modelInstance)
	}

	return result
}

// newAccessorCache 创建访问器缓存
func newAccessorCache() *AccessorCache {
	return &AccessorCache{
		getAccessors: make(map[string]*regexp.Regexp),
		setAccessors: make(map[string]*regexp.Regexp),
		methodCache:  make(map[string]reflect.Method),
		initialized:  false,
	}
}

// initializeAccessors 初始化访问器缓存（使用正则匹配）
func (r *Result) initializeAccessors(modelInstance interface{}) {
	if r.accessors.initialized {
		return
	}

	// 获取类型，需要检查指针类型的方法
	modelType := reflect.TypeOf(modelInstance)
	if modelType == nil {
		r.accessors.initialized = true
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
			r.accessors.getAccessors[attrName] = getPattern
			r.accessors.methodCache[methodName] = method
		}

		// 检查设置器 Set[Name]Attr
		if matches := setPattern.FindStringSubmatch(methodName); len(matches) > 1 {
			attrName := camelToSnake(matches[1])
			r.accessors.setAccessors[attrName] = setPattern
			r.accessors.methodCache[methodName] = method
		}
	}

	r.accessors.initialized = true
}

// Get 获取属性值（支持访问器）
func (r *Result) Get(key string) interface{} {
	if r.data == nil {
		return nil
	}

	// 获取原始值
	rawValue := r.data[key]

	// 如果没有访问器，直接返回处理后的值
	if !r.hasGetAccessor(key) {
		return r.processValue(rawValue)
	}

	// 调用访问器
	return r.callGetAccessor(key, rawValue)
}

// Set 设置属性值（支持修改器）
func (r *Result) Set(key string, value interface{}) *Result {
	if r.data == nil {
		r.data = make(map[string]interface{})
	}

	// 如果没有设置器，直接设置
	if !r.hasSetAccessor(key) {
		r.data[key] = value
		return r
	}

	// 调用设置器
	processedValue := r.callSetAccessor(key, value)
	r.data[key] = processedValue
	return r
}

// GetAll 获取所有属性（支持访问器）
func (r *Result) GetAll() map[string]interface{} {
	if r.data == nil {
		return make(map[string]interface{})
	}

	result := make(map[string]interface{})
	for key := range r.data {
		result[key] = r.Get(key)
	}
	return result
}

// SetAll 批量设置属性（支持修改器）
func (r *Result) SetAll(data map[string]interface{}) *Result {
	for key, value := range data {
		r.Set(key, value)
	}
	return r
}

// GetRaw 获取原始值（不经过访问器）
func (r *Result) GetRaw(key string) interface{} {
	if r.data == nil {
		return nil
	}
	return r.processValue(r.data[key])
}

// GetRawAll 获取所有原始值
func (r *Result) GetRawAll() map[string]interface{} {
	if r.data == nil {
		return make(map[string]interface{})
	}

	result := make(map[string]interface{})
	for key, value := range r.data {
		result[key] = r.processValue(value)
	}
	return result
}

// processValue 处理 []byte 和其他类型的值
func (r *Result) processValue(value interface{}) interface{} {
	if value == nil {
		return nil
	}

	// 处理 []byte 类型
	if bytes, ok := value.([]byte); ok {
		return r.processBytesValue(bytes)
	}

	return value
}

// processBytesValue 智能处理 []byte 数据
func (r *Result) processBytesValue(bytes []byte) interface{} {
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

// CallGetAccessor 调用获取器（公开方法）
func (r *Result) CallGetAccessor(key string, value interface{}) interface{} {
	return r.callGetAccessor(key, value)
}

// CallSetAccessor 调用设置器（公开方法）
func (r *Result) CallSetAccessor(key string, value interface{}) interface{} {
	return r.callSetAccessor(key, value)
}

// callGetAccessor 调用获取器
func (r *Result) callGetAccessor(key string, value interface{}) interface{} {
	methodName := fmt.Sprintf("Get%sAttr", snakeToCamel(key))

	if method, exists := r.accessors.methodCache[methodName]; exists {
		// 先处理 []byte 再传给访问器
		processedValue := r.processValue(value)
		return r.callMethod(method, processedValue)
	}

	// 如果缓存中没有，尝试动态查找
	if r.modelType != nil {
		if method, exists := r.modelType.MethodByName(methodName); exists {
			// 先处理 []byte 再传给访问器
			processedValue := r.processValue(value)
			return r.callMethod(method, processedValue)
		}
	}

	return r.processValue(value)
}

// callSetAccessor 调用设置器
func (r *Result) callSetAccessor(key string, value interface{}) interface{} {
	methodName := fmt.Sprintf("Set%sAttr", snakeToCamel(key))

	if method, exists := r.accessors.methodCache[methodName]; exists {
		// 设置器不需要预处理 []byte，让用户代码自己处理
		return r.callMethod(method, value)
	}

	// 如果缓存中没有，尝试动态查找
	if r.modelType != nil {
		if method, exists := r.modelType.MethodByName(methodName); exists {
			// 设置器不需要预处理 []byte，让用户代码自己处理
			return r.callMethod(method, value)
		}
	}

	return value
}

// callMethod 调用反射方法
func (r *Result) callMethod(method reflect.Method, value interface{}) interface{} {
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

// hasGetAccessor 检查是否有获取器
func (r *Result) hasGetAccessor(key string) bool {
	if r.accessors == nil {
		return false
	}
	_, exists := r.accessors.getAccessors[key]
	return exists
}

// hasSetAccessor 检查是否有设置器
func (r *Result) hasSetAccessor(key string) bool {
	if r.accessors == nil {
		return false
	}
	_, exists := r.accessors.setAccessors[key]
	return exists
}

// ToMap 转换为 map
func (r *Result) ToMap() map[string]interface{} {
	return r.GetAll()
}

// ToRawMap 转换为原始 map
func (r *Result) ToRawMap() map[string]interface{} {
	return r.GetRawAll()
}

// ToJSON 转换为JSON字符串
func (r *Result) ToJSON() (string, error) {
	data := r.GetAll()
	bytes, err := json.Marshal(data)
	return string(bytes), err
}

// ToRawJSON 转换为原始JSON字符串
func (r *Result) ToRawJSON() (string, error) {
	data := r.GetRawAll()
	bytes, err := json.Marshal(data)
	return string(bytes), err
}

// IsEmpty 检查是否为空
func (r *Result) IsEmpty() bool {
	return r.data == nil || len(r.data) == 0
}

// Keys 获取所有键
func (r *Result) Keys() []string {
	keys := make([]string, 0, len(r.data))
	for key := range r.data {
		keys = append(keys, key)
	}
	return keys
}

// Has 检查是否存在某个键
func (r *Result) Has(key string) bool {
	if r.data == nil {
		return false
	}
	_, exists := r.data[key]
	return exists
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

// ResultCollection 结果集合类型
type ResultCollection struct {
	results   []*Result
	modelType reflect.Type
}

// NewResultCollection 创建结果集合
func NewResultCollection(data []map[string]interface{}, modelInstance interface{}) *ResultCollection {
	collection := &ResultCollection{
		results: make([]*Result, len(data)),
	}

	if modelInstance != nil {
		collection.modelType = reflect.TypeOf(modelInstance)
		if collection.modelType.Kind() == reflect.Ptr {
			collection.modelType = collection.modelType.Elem()
		}
	}

	for i, item := range data {
		collection.results[i] = NewResult(item, modelInstance)
	}

	return collection
}

// Get 获取指定索引的结果
func (rc *ResultCollection) Get(index int) *Result {
	if index < 0 || index >= len(rc.results) {
		return NewResult(nil, nil)
	}
	return rc.results[index]
}

// First 获取第一个结果
func (rc *ResultCollection) First() *Result {
	return rc.Get(0)
}

// Last 获取最后一个结果
func (rc *ResultCollection) Last() *Result {
	return rc.Get(len(rc.results) - 1)
}

// Count 获取结果数量
func (rc *ResultCollection) Count() int {
	return len(rc.results)
}

// IsEmpty 检查是否为空
func (rc *ResultCollection) IsEmpty() bool {
	return len(rc.results) == 0
}

// ToSlice 转换为切片
func (rc *ResultCollection) ToSlice() []*Result {
	return rc.results
}

// ToMapSlice 转换为 map 切片（支持访问器）
func (rc *ResultCollection) ToMapSlice() []map[string]interface{} {
	result := make([]map[string]interface{}, len(rc.results))
	for i, item := range rc.results {
		result[i] = item.GetAll()
	}
	return result
}

// ToRawMapSlice 转换为原始 map 切片
func (rc *ResultCollection) ToRawMapSlice() []map[string]interface{} {
	result := make([]map[string]interface{}, len(rc.results))
	for i, item := range rc.results {
		result[i] = item.GetRawAll()
	}
	return result
}

// ToJSON 转换为JSON字符串
func (rc *ResultCollection) ToJSON() (string, error) {
	data := rc.ToMapSlice()
	bytes, err := json.Marshal(data)
	return string(bytes), err
}

// ToRawJSON 转换为原始JSON字符串
func (rc *ResultCollection) ToRawJSON() (string, error) {
	data := rc.ToRawMapSlice()
	bytes, err := json.Marshal(data)
	return string(bytes), err
}

// Each 遍历每个结果
func (rc *ResultCollection) Each(fn func(index int, result *Result) bool) {
	for i, result := range rc.results {
		if !fn(i, result) {
			break
		}
	}
}

// Filter 过滤结果
func (rc *ResultCollection) Filter(fn func(*Result) bool) *ResultCollection {
	var filtered []*Result
	for _, result := range rc.results {
		if fn(result) {
			filtered = append(filtered, result)
		}
	}

	return &ResultCollection{
		results:   filtered,
		modelType: rc.modelType,
	}
}

// Map 映射结果
func (rc *ResultCollection) Map(fn func(*Result) interface{}) []interface{} {
	result := make([]interface{}, len(rc.results))
	for i, item := range rc.results {
		result[i] = fn(item)
	}
	return result
}

