package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
)

// ============================================================================
// JSON字段支持 - 自动序列化/反序列化
// ============================================================================

// JSONSerializer JSON序列化器接口
type JSONSerializer interface {
	// MarshalJSON 序列化为JSON
	MarshalJSON() ([]byte, error)
	// UnmarshalJSON 从JSON反序列化
	UnmarshalJSON(data []byte) error
}

// JSONValue 通用JSON值类型 - 可以存储任意JSON数据
type JSONValue struct {
	Data  interface{}
	Valid bool
}

// NewJSONValue 创建新的JSON值
func NewJSONValue(data interface{}) JSONValue {
	return JSONValue{
		Data:  data,
		Valid: true,
	}
}

// Scan 实现 sql.Scanner 接口
func (j *JSONValue) Scan(value interface{}) error {
	if value == nil {
		j.Data, j.Valid = nil, false
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan type %T into JSONValue", value)
	}

	if len(bytes) == 0 {
		j.Data, j.Valid = nil, false
		return nil
	}

	var data interface{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return err
	}

	j.Data, j.Valid = data, true
	return nil
}

// Value 实现 driver.Valuer 接口
func (j JSONValue) Value() (driver.Value, error) {
	if !j.Valid || j.Data == nil {
		return nil, nil
	}
	return json.Marshal(j.Data)
}

// MarshalJSON 实现 json.Marshaler 接口
func (j JSONValue) MarshalJSON() ([]byte, error) {
	if !j.Valid || j.Data == nil {
		return []byte("null"), nil
	}
	return json.Marshal(j.Data)
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (j *JSONValue) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		j.Data, j.Valid = nil, false
		return nil
	}

	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	j.Data, j.Valid = v, true
	return nil
}

// String 返回JSON字符串表示
func (j JSONValue) String() string {
	if !j.Valid || j.Data == nil {
		return "null"
	}
	bytes, err := json.Marshal(j.Data)
	if err != nil {
		return "null"
	}
	return string(bytes)
}

// IsNull 检查是否为空
func (j JSONValue) IsNull() bool {
	return !j.Valid || j.Data == nil
}

// AsMap 转换为map
func (j JSONValue) AsMap() (map[string]interface{}, bool) {
	if !j.Valid || j.Data == nil {
		return nil, false
	}
	if m, ok := j.Data.(map[string]interface{}); ok {
		return m, true
	}
	return nil, false
}

// AsSlice 转换为切片
func (j JSONValue) AsSlice() ([]interface{}, bool) {
	if !j.Valid || j.Data == nil {
		return nil, false
	}
	if s, ok := j.Data.([]interface{}); ok {
		return s, true
	}
	return nil, false
}

// Get 获取嵌套值（支持点号路径）
func (j JSONValue) Get(path string) interface{} {
	if !j.Valid || j.Data == nil {
		return nil
	}
	return getNestedValue(j.Data, path)
}

// Set 设置嵌套值（支持点号路径）
func (j *JSONValue) Set(path string, value interface{}) {
	if !j.Valid {
		j.Data = make(map[string]interface{})
		j.Valid = true
	}
	setNestedValue(&j.Data, path, value)
}

// ============================================================================
// JSONMap 类型 - 专门用于存储JSON对象
// ============================================================================

// JSONMap JSON对象类型 - 存储键值对
type JSONMap map[string]interface{}

// Scan 实现 sql.Scanner 接口
func (j *JSONMap) Scan(value interface{}) error {
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
		return fmt.Errorf("cannot scan type %T into JSONMap", value)
	}

	if len(bytes) == 0 {
		*j = nil
		return nil
	}

	return json.Unmarshal(bytes, j)
}

// Value 实现 driver.Valuer 接口
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// MarshalJSON 实现 json.Marshaler 接口
func (j JSONMap) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return json.Marshal(map[string]interface{}(j))
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (j *JSONMap) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*j = nil
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	*j = m
	return nil
}

// Get 获取值
func (j JSONMap) Get(key string) interface{} {
	if j == nil {
		return nil
	}
	return j[key]
}

// Set 设置值
func (j JSONMap) Set(key string, value interface{}) {
	if j != nil {
		j[key] = value
	}
}

// Has 检查键是否存在
func (j JSONMap) Has(key string) bool {
	if j == nil {
		return false
	}
	_, exists := j[key]
	return exists
}

// Delete 删除键
func (j JSONMap) Delete(key string) {
	if j != nil {
		delete(j, key)
	}
}

// Keys 获取所有键
func (j JSONMap) Keys() []string {
	if j == nil {
		return nil
	}
	keys := make([]string, 0, len(j))
	for k := range j {
		keys = append(keys, k)
	}
	return keys
}

// ============================================================================
// JSONSlice 类型 - 专门用于存储JSON数组
// ============================================================================

// JSONSlice JSON数组类型
type JSONSlice []interface{}

// Scan 实现 sql.Scanner 接口
func (j *JSONSlice) Scan(value interface{}) error {
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
		return fmt.Errorf("cannot scan type %T into JSONSlice", value)
	}

	if len(bytes) == 0 {
		*j = nil
		return nil
	}

	return json.Unmarshal(bytes, j)
}

// Value 实现 driver.Valuer 接口
func (j JSONSlice) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// MarshalJSON 实现 json.Marshaler 接口
func (j JSONSlice) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return json.Marshal([]interface{}(j))
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (j *JSONSlice) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*j = nil
		return nil
	}
	var s []interface{}
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*j = s
	return nil
}

// Len 获取长度
func (j JSONSlice) Len() int {
	return len(j)
}

// Get 获取指定索引的值
func (j JSONSlice) Get(index int) interface{} {
	if index < 0 || index >= len(j) {
		return nil
	}
	return j[index]
}

// Append 追加值
func (j *JSONSlice) Append(value interface{}) {
	*j = append(*j, value)
}

// ============================================================================
// JSONString 类型 - 存储为JSON字符串的结构体
// ============================================================================

// JSONString 将结构体序列化为JSON字符串存储
type JSONString[T any] struct {
	Data  T
	Valid bool
}

// NewJSONString 创建新的JSONString
func NewJSONString[T any](data T) JSONString[T] {
	return JSONString[T]{
		Data:  data,
		Valid: true,
	}
}

// Scan 实现 sql.Scanner 接口
func (j *JSONString[T]) Scan(value interface{}) error {
	if value == nil {
		j.Valid = false
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan type %T into JSONString", value)
	}

	if len(bytes) == 0 {
		j.Valid = false
		return nil
	}

	if err := json.Unmarshal(bytes, &j.Data); err != nil {
		return err
	}

	j.Valid = true
	return nil
}

// Value 实现 driver.Valuer 接口
func (j JSONString[T]) Value() (driver.Value, error) {
	if !j.Valid {
		return nil, nil
	}
	return json.Marshal(j.Data)
}

// MarshalJSON 实现 json.Marshaler 接口
func (j JSONString[T]) MarshalJSON() ([]byte, error) {
	if !j.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(j.Data)
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (j *JSONString[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		j.Valid = false
		return nil
	}
	if err := json.Unmarshal(data, &j.Data); err != nil {
		return err
	}
	j.Valid = true
	return nil
}

// ============================================================================
// 辅助函数
// ============================================================================

// getNestedValue 获取嵌套值
func getNestedValue(data interface{}, path string) interface{} {
	if path == "" {
		return data
	}

	parts := splitPath(path)
	current := data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			var exists bool
			current, exists = v[part]
			if !exists {
				return nil
			}
		case []interface{}:
			// 尝试解析为索引
			var index int
			if _, err := fmt.Sscanf(part, "%d", &index); err == nil {
				if index >= 0 && index < len(v) {
					current = v[index]
				} else {
					return nil
				}
			} else {
				return nil
			}
		default:
			return nil
		}
	}

	return current
}

// setNestedValue 设置嵌套值
func setNestedValue(data *interface{}, path string, value interface{}) {
	if path == "" {
		*data = value
		return
	}

	parts := splitPath(path)
	if len(parts) == 0 {
		return
	}

	// 确保根是map
	if *data == nil {
		*data = make(map[string]interface{})
	}

	current := *data
	for i, part := range parts[:len(parts)-1] {
		switch v := current.(type) {
		case map[string]interface{}:
			if _, exists := v[part]; !exists {
				v[part] = make(map[string]interface{})
			}
			current = v[part]
		default:
			// 创建新的map
			newMap := make(map[string]interface{})
			if i == 0 {
				*data = newMap
			}
			current = newMap
		}
	}

	// 设置最终值
	if m, ok := current.(map[string]interface{}); ok {
		m[parts[len(parts)-1]] = value
	}
}

// splitPath 分割路径
func splitPath(path string) []string {
	var parts []string
	var current string

	for _, c := range path {
		if c == '.' || c == '[' || c == ']' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// ============================================================================
// JSON字段检测和处理
// ============================================================================

// IsJSONField 检查字段是否为JSON字段
func IsJSONField(field reflect.StructField) bool {
	// 检查torm标签
	tormTag := field.Tag.Get("torm")
	if containsJSONTag(tormTag) {
		return true
	}

	// 检查gorm标签
	gormTag := field.Tag.Get("gorm")
	if containsJSONTag(gormTag) {
		return true
	}

	// 检查字段类型
	fieldType := field.Type
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}

	// 检查是否为JSON类型
	switch fieldType.Name() {
	case "JSONField", "JSONArray", "JSONValue", "JSONMap", "JSONSlice":
		return true
	}

	// 检查是否为map或slice类型（可能需要JSON序列化）
	switch fieldType.Kind() {
	case reflect.Map, reflect.Slice:
		// 检查是否有serializer:json标签
		if containsSerializerJSON(tormTag) || containsSerializerJSON(gormTag) {
			return true
		}
	}

	return false
}

// containsJSONTag 检查标签是否包含JSON标记
func containsJSONTag(tag string) bool {
	if tag == "" {
		return false
	}

	// 分割标签
	var parts []string
	if containsChar(tag, ';') {
		parts = splitByChar(tag, ';')
	} else {
		parts = splitByChar(tag, ',')
	}

	for _, part := range parts {
		part = trimSpaceStr(part)
		partLower := toLowerStr(part)
		if partLower == "json" || partLower == "is_json" || partLower == "isjson" {
			return true
		}
	}

	return false
}

// containsSerializerJSON 检查是否有serializer:json标签
func containsSerializerJSON(tag string) bool {
	if tag == "" {
		return false
	}

	// 分割标签
	var parts []string
	if containsChar(tag, ';') {
		parts = splitByChar(tag, ';')
	} else {
		parts = splitByChar(tag, ',')
	}

	for _, part := range parts {
		part = trimSpaceStr(part)
		partLower := toLowerStr(part)
		if partLower == "serializer:json" {
			return true
		}
	}

	return false
}

// 辅助函数
func containsChar(s string, c rune) bool {
	for _, r := range s {
		if r == c {
			return true
		}
	}
	return false
}

func splitByChar(s string, c rune) []string {
	var result []string
	var current string
	for _, r := range s {
		if r == c {
			result = append(result, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	result = append(result, current)
	return result
}

func trimSpaceStr(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func toLowerStr(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}
