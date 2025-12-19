package model

import (
	"database/sql/driver"
	"reflect"
	"time"
)

// TimestampConfig 时间戳配置
type TimestampConfig struct {
	Enabled      bool   // 是否启用时间戳
	CreatedField string // 创建时间字段名
	UpdatedField string // 更新时间字段名
	Format       string // 时间格式（用于字符串类型）
	UseUnix      bool   // 是否使用Unix时间戳
}

// DefaultTimestampConfig 默认时间戳配置
func DefaultTimestampConfig() TimestampConfig {
	return TimestampConfig{
		Enabled:      true,
		CreatedField: "created_at",
		UpdatedField: "updated_at",
		Format:       "2006-01-02 15:04:05",
		UseUnix:      false,
	}
}

// Timestamp 时间戳类型 - 支持多种格式
type Timestamp struct {
	Time   time.Time
	Valid  bool
	Format string
}

// NewTimestamp 创建新的时间戳
func NewTimestamp(t time.Time) Timestamp {
	return Timestamp{
		Time:   t,
		Valid:  true,
		Format: "2006-01-02 15:04:05",
	}
}

// Now 创建当前时间的时间戳
func Now() Timestamp {
	return NewTimestamp(time.Now())
}

// Scan 实现 sql.Scanner 接口
func (ts *Timestamp) Scan(value interface{}) error {
	if value == nil {
		ts.Time, ts.Valid = time.Time{}, false
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		ts.Time, ts.Valid = v, true
	case int64:
		ts.Time, ts.Valid = time.Unix(v, 0), true
	case string:
		if v == "" {
			ts.Time, ts.Valid = time.Time{}, false
		} else {
			t, err := parseTimeString(v)
			if err != nil {
				return err
			}
			ts.Time, ts.Valid = t, true
		}
	case []byte:
		if len(v) == 0 {
			ts.Time, ts.Valid = time.Time{}, false
		} else {
			t, err := parseTimeString(string(v))
			if err != nil {
				return err
			}
			ts.Time, ts.Valid = t, true
		}
	}
	return nil
}

// Value 实现 driver.Valuer 接口
func (ts Timestamp) Value() (driver.Value, error) {
	if !ts.Valid {
		return nil, nil
	}
	if ts.Format != "" {
		return ts.Time.Format(ts.Format), nil
	}
	return ts.Time, nil
}

// String 返回格式化的时间字符串
func (ts Timestamp) String() string {
	if !ts.Valid {
		return ""
	}
	format := ts.Format
	if format == "" {
		format = "2006-01-02 15:04:05"
	}
	return ts.Time.Format(format)
}

// Unix 返回Unix时间戳
func (ts Timestamp) Unix() int64 {
	if !ts.Valid {
		return 0
	}
	return ts.Time.Unix()
}

// UnixMilli 返回毫秒级Unix时间戳
func (ts Timestamp) UnixMilli() int64 {
	if !ts.Valid {
		return 0
	}
	return ts.Time.UnixMilli()
}

// IsZero 检查是否为零值
func (ts Timestamp) IsZero() bool {
	return !ts.Valid || ts.Time.IsZero()
}

// parseTimeString 解析时间字符串
func parseTimeString(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02T15:04:05.999999999",
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02",
		"2006/01/02 15:04:05",
		"2006/01/02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, nil
}

// TimestampManager 时间戳管理器
type TimestampManager struct {
	config TimestampConfig
}

// NewTimestampManager 创建时间戳管理器
func NewTimestampManager(config ...TimestampConfig) *TimestampManager {
	cfg := DefaultTimestampConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	return &TimestampManager{config: cfg}
}

// ProcessInsertData 处理插入数据的时间戳
func (tm *TimestampManager) ProcessInsertData(data map[string]interface{}) map[string]interface{} {
	if !tm.config.Enabled {
		return data
	}

	now := time.Now()
	value := tm.formatTime(now)

	// 设置创建时间
	if tm.config.CreatedField != "" {
		if _, exists := data[tm.config.CreatedField]; !exists {
			data[tm.config.CreatedField] = value
		}
	}

	// 设置更新时间
	if tm.config.UpdatedField != "" {
		data[tm.config.UpdatedField] = value
	}

	return data
}

// ProcessUpdateData 处理更新数据的时间戳
func (tm *TimestampManager) ProcessUpdateData(data map[string]interface{}) map[string]interface{} {
	if !tm.config.Enabled {
		return data
	}

	now := time.Now()
	value := tm.formatTime(now)

	// 只设置更新时间
	if tm.config.UpdatedField != "" {
		data[tm.config.UpdatedField] = value
	}

	return data
}

// formatTime 格式化时间
func (tm *TimestampManager) formatTime(t time.Time) interface{} {
	if tm.config.UseUnix {
		return t.Unix()
	}
	if tm.config.Format != "" {
		return t.Format(tm.config.Format)
	}
	return t
}

// InferTimestampFields 从结构体推断时间戳字段
func InferTimestampFields(model interface{}) (createdAt, updatedAt string) {
	createdAt = "created_at"
	updatedAt = "updated_at"

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

		tormTag := field.Tag.Get("torm")
		if tormTag == "" {
			continue
		}

		// 检查是否为创建时间字段
		if containsTag(tormTag, "auto_create_time", "create_time", "created_at", "autocreate") {
			createdAt = getColumnNameFromTag(field)
		}

		// 检查是否为更新时间字段
		if containsTag(tormTag, "auto_update_time", "update_time", "updated_at", "autoupdate") {
			updatedAt = getColumnNameFromTag(field)
		}
	}

	return
}

// containsTag 检查标签是否包含指定值
func containsTag(tag string, values ...string) bool {
	for _, v := range values {
		if containsTagValue(tag, v) {
			return true
		}
	}
	return false
}

// containsTagValue 检查标签是否包含指定值
func containsTagValue(tag, value string) bool {
	parts := splitTag(tag)
	for _, part := range parts {
		if part == value {
			return true
		}
	}
	return false
}

// splitTag 分割标签
func splitTag(tag string) []string {
	var parts []string
	for _, part := range splitString(tag, ",") {
		part = trimSpace(part)
		if part != "" {
			parts = append(parts, toLower(part))
		}
	}
	return parts
}

// getColumnNameFromTag 从字段标签获取列名
func getColumnNameFromTag(field reflect.StructField) string {
	// 检查json标签
	if jsonTag := field.Tag.Get("json"); jsonTag != "" {
		parts := splitString(jsonTag, ",")
		if parts[0] != "" && parts[0] != "-" {
			return parts[0]
		}
	}

	// 使用字段名的蛇形命名
	return toSnakeCase(field.Name)
}

// 辅助函数
func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if string(s[i]) == sep {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
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

func toLower(s string) string {
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
