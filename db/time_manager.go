package db

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// TimeFieldManager 时间字段管理器
type TimeFieldManager struct {
	createTimeFields []string // 自动创建时间字段列表
	updateTimeFields []string // 自动更新时间字段列表
}

// NewTimeFieldManager 创建时间字段管理器
func NewTimeFieldManager() *TimeFieldManager {
	return &TimeFieldManager{
		createTimeFields: make([]string, 0),
		updateTimeFields: make([]string, 0),
	}
}

// TimeFieldInfo 时间字段信息
type TimeFieldInfo struct {
	FieldName    string      // 字段名
	ColumnName   string      // 数据库列名
	FieldType    reflect.Type // 字段类型
	IsCreateTime bool        // 是否为创建时间字段
	IsUpdateTime bool        // 是否为更新时间字段
}

// AnalyzeModelTimeFields 分析模型的时间字段
func (tfm *TimeFieldManager) AnalyzeModelTimeFields(modelInstance interface{}) []TimeFieldInfo {
	var timeFields []TimeFieldInfo
	
	if modelInstance == nil {
		return timeFields
	}

	modelType := reflect.TypeOf(modelInstance)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		
		// 跳过嵌入的BaseModel字段
		if field.Type.Name() == "BaseModel" || field.Anonymous {
			continue
		}

		timeFieldInfo := tfm.analyzeField(field)
		if timeFieldInfo != nil {
			timeFields = append(timeFields, *timeFieldInfo)
		}
	}

	return timeFields
}

// analyzeField 分析单个字段
func (tfm *TimeFieldManager) analyzeField(field reflect.StructField) *TimeFieldInfo {
	// 检查是否为时间字段
	isCreateTime, isUpdateTime := tfm.checkTimeFieldTags(field)
	
	if !isCreateTime && !isUpdateTime {
		return nil
	}

	// 检查字段类型是否支持时间处理
	if !tfm.isTimeCompatibleType(field.Type) {
		return nil
	}

	columnName := tfm.getColumnNameFromField(field)
	
	return &TimeFieldInfo{
		FieldName:    field.Name,
		ColumnName:   columnName,
		FieldType:    field.Type,
		IsCreateTime: isCreateTime,
		IsUpdateTime: isUpdateTime,
	}
}

// checkTimeFieldTags 检查时间字段标签
func (tfm *TimeFieldManager) checkTimeFieldTags(field reflect.StructField) (isCreateTime, isUpdateTime bool) {
	// 检查torm标签
	if tormTag := field.Tag.Get("torm"); tormTag != "" {
		parts := strings.Split(tormTag, ",")
		for _, part := range parts {
			part = strings.TrimSpace(strings.ToLower(part))
			
			// 检查创建时间标记
			if tfm.isCreateTimeTag(part) {
				isCreateTime = true
			}
			
			// 检查更新时间标记  
			if tfm.isUpdateTimeTag(part) {
				isUpdateTime = true
			}
		}
	}

	// 检查传统标签格式
	if !isCreateTime {
		createTimeTags := []string{"autoCreateTime", "AutoCreateTime", "auto_create_time"}
		for _, tag := range createTimeTags {
			if field.Tag.Get(tag) != "" {
				isCreateTime = true
				break
			}
		}
	}

	if !isUpdateTime {
		updateTimeTags := []string{"autoUpdateTime", "AutoUpdateTime", "auto_update_time"}
		for _, tag := range updateTimeTags {
			if field.Tag.Get(tag) != "" {
				isUpdateTime = true
				break
			}
		}
	}

	return
}

// isCreateTimeTag 检查是否为创建时间标记
func (tfm *TimeFieldManager) isCreateTimeTag(tag string) bool {
	createTimeTags := []string{
		"auto_create_time", "create_time", "created_at", "auto_created_at",
		"autocreate_time", "autocreatetime", "auto_create", "autocreate",
	}
	
	for _, createTag := range createTimeTags {
		if tag == createTag {
			return true
		}
	}
	return false
}

// isUpdateTimeTag 检查是否为更新时间标记
func (tfm *TimeFieldManager) isUpdateTimeTag(tag string) bool {
	updateTimeTags := []string{
		"auto_update_time", "update_time", "updated_at", "auto_updated_at",
		"autoupdate_time", "autoupdatetime", "auto_update", "autoupdate",
	}
	
	for _, updateTag := range updateTimeTags {
		if tag == updateTag {
			return true
		}
	}
	return false
}

// isTimeCompatibleType 检查字段类型是否支持时间处理
func (tfm *TimeFieldManager) isTimeCompatibleType(fieldType reflect.Type) bool {
	// 处理指针类型
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}

	switch fieldType.Kind() {
	case reflect.String:
		return true // string 类型支持
	case reflect.Int, reflect.Int64:
		return true // int64/int 类型支持（Unix时间戳）
	case reflect.Struct:
		// 检查是否是time.Time
		return fieldType.PkgPath() == "time" && fieldType.Name() == "Time"
	default:
		return false
	}
}

// getColumnNameFromField 从字段获取列名
func (tfm *TimeFieldManager) getColumnNameFromField(field reflect.StructField) string {
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

	// 检查json标签作为列名
	if jsonTag := field.Tag.Get("json"); jsonTag != "" {
		parts := strings.Split(jsonTag, ",")
		if parts[0] != "" && parts[0] != "-" {
			return parts[0]
		}
	}

	// 检查db标签
	if dbTag := field.Tag.Get("db"); dbTag != "" {
		return dbTag
	}

	// 使用字段名的蛇形命名
	return tfm.toSnakeCase(field.Name)
}

// ProcessInsertData 处理插入数据的时间字段
func (tfm *TimeFieldManager) ProcessInsertData(data map[string]interface{}, timeFields []TimeFieldInfo) map[string]interface{} {
	if len(timeFields) == 0 {
		return data
	}

	result := make(map[string]interface{})
	// 复制原有数据
	for k, v := range data {
		result[k] = v
	}

	now := time.Now()

	for _, fieldInfo := range timeFields {
		if fieldInfo.IsCreateTime {
			// 对于创建时间字段，总是设置当前时间（如果用户没有设置的话）
			if _, exists := result[fieldInfo.ColumnName]; !exists {
				result[fieldInfo.ColumnName] = tfm.convertToFieldType(now, fieldInfo.FieldType)
			}
		}
		
		if fieldInfo.IsUpdateTime {
			// 对于更新时间字段，在插入时也设置当前时间
			result[fieldInfo.ColumnName] = tfm.convertToFieldType(now, fieldInfo.FieldType)
		}
	}

	return result
}

// ProcessUpdateData 处理更新数据的时间字段
func (tfm *TimeFieldManager) ProcessUpdateData(data map[string]interface{}, timeFields []TimeFieldInfo) map[string]interface{} {
	if len(timeFields) == 0 {
		return data
	}

	result := make(map[string]interface{})
	// 复制原有数据
	for k, v := range data {
		result[k] = v
	}

	now := time.Now()

	for _, fieldInfo := range timeFields {
		if fieldInfo.IsUpdateTime {
			// 对于更新时间字段，总是设置当前时间
			result[fieldInfo.ColumnName] = tfm.convertToFieldType(now, fieldInfo.FieldType)
		}
		// 创建时间字段在更新时不处理
	}

	return result
}

// convertToFieldType 将时间转换为字段对应的类型
func (tfm *TimeFieldManager) convertToFieldType(t time.Time, fieldType reflect.Type) interface{} {
	// 处理指针类型
	isPtr := fieldType.Kind() == reflect.Ptr
	if isPtr {
		fieldType = fieldType.Elem()
	}

	var result interface{}

	switch fieldType.Kind() {
	case reflect.String:
		// 转换为字符串格式
		result = t.Format("2006-01-02 15:04:05")
	case reflect.Int64:
		// 转换为Unix时间戳（秒）
		result = t.Unix()
	case reflect.Int:
		// 转换为Unix时间戳（秒）
		result = int(t.Unix())
	case reflect.Struct:
		// 检查是否是time.Time
		if fieldType.PkgPath() == "time" && fieldType.Name() == "Time" {
			result = t
		} else {
			result = t.Format("2006-01-02 15:04:05")
		}
	default:
		result = t
	}

	// 如果是指针类型，返回指针
	if isPtr {
		resultValue := reflect.ValueOf(result)
		ptr := reflect.New(resultValue.Type())
		ptr.Elem().Set(resultValue)
		return ptr.Interface()
	}

	return result
}

// ParseTimeValue 解析时间值（从数据库读取时使用）
func (tfm *TimeFieldManager) ParseTimeValue(value interface{}, fieldType reflect.Type) interface{} {
	if value == nil {
		return nil
	}

	// 处理指针类型
	isPtr := fieldType.Kind() == reflect.Ptr
	if isPtr {
		fieldType = fieldType.Elem()
	}

	var result interface{}

	switch v := value.(type) {
	case time.Time:
		result = tfm.convertFromTime(v, fieldType)
	case string:
		if t, err := tfm.parseTimeString(v); err == nil {
			result = tfm.convertFromTime(t, fieldType)
		} else {
			result = v
		}
	case int64:
		t := time.Unix(v, 0)
		result = tfm.convertFromTime(t, fieldType)
	case int:
		t := time.Unix(int64(v), 0)
		result = tfm.convertFromTime(t, fieldType)
	default:
		result = value
	}

	// 如果是指针类型，返回指针
	if isPtr && result != nil {
		resultValue := reflect.ValueOf(result)
		ptr := reflect.New(resultValue.Type())
		ptr.Elem().Set(resultValue)
		return ptr.Interface()
	}

	return result
}

// convertFromTime 从time.Time转换为目标类型
func (tfm *TimeFieldManager) convertFromTime(t time.Time, fieldType reflect.Type) interface{} {
	switch fieldType.Kind() {
	case reflect.String:
		return t.Format("2006-01-02 15:04:05")
	case reflect.Int64:
		return t.Unix()
	case reflect.Int:
		return int(t.Unix())
	case reflect.Struct:
		if fieldType.PkgPath() == "time" && fieldType.Name() == "Time" {
			return t
		}
		return t.Format("2006-01-02 15:04:05")
	default:
		return t
	}
}

// parseTimeString 解析时间字符串
func (tfm *TimeFieldManager) parseTimeString(timeStr string) (time.Time, error) {
	timeStr = strings.TrimSpace(timeStr)
	if timeStr == "" {
		return time.Time{}, fmt.Errorf("empty time string")
	}

	// 尝试解析各种时间格式
	timeFormats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02T15:04:05.999999999",
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02",
		"15:04:05",
		"2006/01/02 15:04:05",
		"2006/01/02",
		"01/02/2006",
		"01-02-2006",
	}

	for _, format := range timeFormats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	// 尝试解析Unix时间戳
	if timestamp, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
		// 区分秒级和毫秒级时间戳
		if timestamp > 1e10 { // 毫秒级时间戳
			return time.Unix(timestamp/1000, (timestamp%1000)*1000000), nil
		} else { // 秒级时间戳
			return time.Unix(timestamp, 0), nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time string: %s", timeStr)
}

// toSnakeCase 将驼峰命名转换为蛇形命名
func (tfm *TimeFieldManager) toSnakeCase(str string) string {
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
			// 3. 当前字符后面跟着小写字母（处理连续大写的情况）
			if i > 0 && ((runes[i-1] >= 'a' && runes[i-1] <= 'z') || 
				(i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z')) {
				result.WriteRune('_')
			}
			result.WriteRune(r - 'A' + 'a') // 转为小写
		} else {
			result.WriteRune(r)
		}
	}
	
	return result.String()
}
