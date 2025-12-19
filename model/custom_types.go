package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ============================================================================
// 自定义类型接口定义
// ============================================================================

// Scanner 扫描器接口 - 用于从数据库读取数据
type Scanner interface {
	Scan(value interface{}) error
}

// Valuer 值接口 - 用于向数据库写入数据
type Valuer interface {
	Value() (driver.Value, error)
}

// CustomType 自定义类型接口 - 组合Scanner和Valuer
type CustomType interface {
	Scanner
	Valuer
}

// TypeConverter 类型转换器接口
type TypeConverter interface {
	// ConvertToDatabase 转换为数据库值
	ConvertToDatabase(value interface{}) (driver.Value, error)
	// ConvertFromDatabase 从数据库值转换
	ConvertFromDatabase(value interface{}) (interface{}, error)
}

// ============================================================================
// 自定义类型注册表
// ============================================================================

// customTypeRegistry 自定义类型注册表
var customTypeRegistry = struct {
	sync.RWMutex
	types      map[reflect.Type]TypeConverter
	converters map[string]TypeConverter
}{
	types:      make(map[reflect.Type]TypeConverter),
	converters: make(map[string]TypeConverter),
}

// RegisterCustomType 注册自定义类型转换器
func RegisterCustomType(typ reflect.Type, converter TypeConverter) {
	customTypeRegistry.Lock()
	defer customTypeRegistry.Unlock()
	customTypeRegistry.types[typ] = converter
}

// RegisterCustomTypeByName 按名称注册自定义类型转换器
func RegisterCustomTypeByName(name string, converter TypeConverter) {
	customTypeRegistry.Lock()
	defer customTypeRegistry.Unlock()
	customTypeRegistry.converters[name] = converter
}

// GetCustomTypeConverter 获取自定义类型转换器
func GetCustomTypeConverter(typ reflect.Type) (TypeConverter, bool) {
	customTypeRegistry.RLock()
	defer customTypeRegistry.RUnlock()
	converter, exists := customTypeRegistry.types[typ]
	return converter, exists
}

// GetCustomTypeConverterByName 按名称获取自定义类型转换器
func GetCustomTypeConverterByName(name string) (TypeConverter, bool) {
	customTypeRegistry.RLock()
	defer customTypeRegistry.RUnlock()
	converter, exists := customTypeRegistry.converters[name]
	return converter, exists
}

// ============================================================================
// NullString - 可空字符串类型
// ============================================================================

// NullString 可空字符串类型
type NullString struct {
	String string
	Valid  bool
}

// NewNullString 创建新的NullString
func NewNullString(s string) NullString {
	return NullString{String: s, Valid: true}
}

// Scan 实现 sql.Scanner 接口
func (ns *NullString) Scan(value interface{}) error {
	if value == nil {
		ns.String, ns.Valid = "", false
		return nil
	}
	switch v := value.(type) {
	case string:
		ns.String, ns.Valid = v, true
	case []byte:
		ns.String, ns.Valid = string(v), true
	default:
		ns.String, ns.Valid = fmt.Sprintf("%v", v), true
	}
	return nil
}

// Value 实现 driver.Valuer 接口
func (ns NullString) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.String, nil
}

// MarshalJSON 实现 json.Marshaler 接口
func (ns NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (ns *NullString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		ns.String, ns.Valid = "", false
		return nil
	}
	if err := json.Unmarshal(data, &ns.String); err != nil {
		return err
	}
	ns.Valid = true
	return nil
}

// IsNull 检查是否为空
func (ns NullString) IsNull() bool {
	return !ns.Valid
}

// SetNull 设置为空
func (ns *NullString) SetNull() {
	ns.String, ns.Valid = "", false
}

// Set 设置值
func (ns *NullString) Set(s string) {
	ns.String, ns.Valid = s, true
}

// ============================================================================
// NullInt64 - 可空整数类型
// ============================================================================

// NullInt64 可空整数类型
type NullInt64 struct {
	Int64 int64
	Valid bool
}

// NewNullInt64 创建新的NullInt64
func NewNullInt64(i int64) NullInt64 {
	return NullInt64{Int64: i, Valid: true}
}

// Scan 实现 sql.Scanner 接口
func (ni *NullInt64) Scan(value interface{}) error {
	if value == nil {
		ni.Int64, ni.Valid = 0, false
		return nil
	}
	switch v := value.(type) {
	case int64:
		ni.Int64, ni.Valid = v, true
	case int:
		ni.Int64, ni.Valid = int64(v), true
	case int32:
		ni.Int64, ni.Valid = int64(v), true
	case float64:
		ni.Int64, ni.Valid = int64(v), true
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
		ni.Int64, ni.Valid = i, true
	case []byte:
		i, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return err
		}
		ni.Int64, ni.Valid = i, true
	default:
		return fmt.Errorf("cannot scan type %T into NullInt64", value)
	}
	return nil
}

// Value 实现 driver.Valuer 接口
func (ni NullInt64) Value() (driver.Value, error) {
	if !ni.Valid {
		return nil, nil
	}
	return ni.Int64, nil
}

// MarshalJSON 实现 json.Marshaler 接口
func (ni NullInt64) MarshalJSON() ([]byte, error) {
	if !ni.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ni.Int64)
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (ni *NullInt64) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		ni.Int64, ni.Valid = 0, false
		return nil
	}
	if err := json.Unmarshal(data, &ni.Int64); err != nil {
		return err
	}
	ni.Valid = true
	return nil
}

// IsNull 检查是否为空
func (ni NullInt64) IsNull() bool {
	return !ni.Valid
}

// Set 设置值
func (ni *NullInt64) Set(i int64) {
	ni.Int64, ni.Valid = i, true
}

// ============================================================================
// NullFloat64 - 可空浮点数类型
// ============================================================================

// NullFloat64 可空浮点数类型
type NullFloat64 struct {
	Float64 float64
	Valid   bool
}

// NewNullFloat64 创建新的NullFloat64
func NewNullFloat64(f float64) NullFloat64 {
	return NullFloat64{Float64: f, Valid: true}
}

// Scan 实现 sql.Scanner 接口
func (nf *NullFloat64) Scan(value interface{}) error {
	if value == nil {
		nf.Float64, nf.Valid = 0, false
		return nil
	}
	switch v := value.(type) {
	case float64:
		nf.Float64, nf.Valid = v, true
	case float32:
		nf.Float64, nf.Valid = float64(v), true
	case int64:
		nf.Float64, nf.Valid = float64(v), true
	case int:
		nf.Float64, nf.Valid = float64(v), true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return err
		}
		nf.Float64, nf.Valid = f, true
	case []byte:
		f, err := strconv.ParseFloat(string(v), 64)
		if err != nil {
			return err
		}
		nf.Float64, nf.Valid = f, true
	default:
		return fmt.Errorf("cannot scan type %T into NullFloat64", value)
	}
	return nil
}

// Value 实现 driver.Valuer 接口
func (nf NullFloat64) Value() (driver.Value, error) {
	if !nf.Valid {
		return nil, nil
	}
	return nf.Float64, nil
}

// MarshalJSON 实现 json.Marshaler 接口
func (nf NullFloat64) MarshalJSON() ([]byte, error) {
	if !nf.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nf.Float64)
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (nf *NullFloat64) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		nf.Float64, nf.Valid = 0, false
		return nil
	}
	if err := json.Unmarshal(data, &nf.Float64); err != nil {
		return err
	}
	nf.Valid = true
	return nil
}

// IsNull 检查是否为空
func (nf NullFloat64) IsNull() bool {
	return !nf.Valid
}

// Set 设置值
func (nf *NullFloat64) Set(f float64) {
	nf.Float64, nf.Valid = f, true
}

// ============================================================================
// NullBool - 可空布尔类型
// ============================================================================

// NullBool 可空布尔类型
type NullBool struct {
	Bool  bool
	Valid bool
}

// NewNullBool 创建新的NullBool
func NewNullBool(b bool) NullBool {
	return NullBool{Bool: b, Valid: true}
}

// Scan 实现 sql.Scanner 接口
func (nb *NullBool) Scan(value interface{}) error {
	if value == nil {
		nb.Bool, nb.Valid = false, false
		return nil
	}
	switch v := value.(type) {
	case bool:
		nb.Bool, nb.Valid = v, true
	case int64:
		nb.Bool, nb.Valid = v != 0, true
	case int:
		nb.Bool, nb.Valid = v != 0, true
	case string:
		b, err := strconv.ParseBool(v)
		if err != nil {
			// 尝试解析数字
			if v == "1" || strings.ToLower(v) == "true" || strings.ToLower(v) == "yes" {
				nb.Bool, nb.Valid = true, true
				return nil
			}
			if v == "0" || strings.ToLower(v) == "false" || strings.ToLower(v) == "no" {
				nb.Bool, nb.Valid = false, true
				return nil
			}
			return err
		}
		nb.Bool, nb.Valid = b, true
	case []byte:
		return nb.Scan(string(v))
	default:
		return fmt.Errorf("cannot scan type %T into NullBool", value)
	}
	return nil
}

// Value 实现 driver.Valuer 接口
func (nb NullBool) Value() (driver.Value, error) {
	if !nb.Valid {
		return nil, nil
	}
	return nb.Bool, nil
}

// MarshalJSON 实现 json.Marshaler 接口
func (nb NullBool) MarshalJSON() ([]byte, error) {
	if !nb.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nb.Bool)
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (nb *NullBool) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		nb.Bool, nb.Valid = false, false
		return nil
	}
	if err := json.Unmarshal(data, &nb.Bool); err != nil {
		return err
	}
	nb.Valid = true
	return nil
}

// IsNull 检查是否为空
func (nb NullBool) IsNull() bool {
	return !nb.Valid
}

// Set 设置值
func (nb *NullBool) Set(b bool) {
	nb.Bool, nb.Valid = b, true
}

// ============================================================================
// NullTime - 可空时间类型
// ============================================================================

// NullTime 可空时间类型
type NullTime struct {
	Time  time.Time
	Valid bool
}

// NewNullTime 创建新的NullTime
func NewNullTime(t time.Time) NullTime {
	return NullTime{Time: t, Valid: true}
}

// NullTimeNow 创建当前时间的NullTime
func NullTimeNow() NullTime {
	return NewNullTime(time.Now())
}

// Scan 实现 sql.Scanner 接口
func (nt *NullTime) Scan(value interface{}) error {
	if value == nil {
		nt.Time, nt.Valid = time.Time{}, false
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		nt.Time, nt.Valid = v, true
	case int64:
		nt.Time, nt.Valid = time.Unix(v, 0), true
	case string:
		if v == "" {
			nt.Time, nt.Valid = time.Time{}, false
			return nil
		}
		t, err := parseTimeFormats(v)
		if err != nil {
			return err
		}
		nt.Time, nt.Valid = t, true
	case []byte:
		return nt.Scan(string(v))
	default:
		return fmt.Errorf("cannot scan type %T into NullTime", value)
	}
	return nil
}

// Value 实现 driver.Valuer 接口
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

// MarshalJSON 实现 json.Marshaler 接口
func (nt NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nt.Time.Format(time.RFC3339))
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (nt *NullTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		nt.Time, nt.Valid = time.Time{}, false
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	t, err := parseTimeFormats(s)
	if err != nil {
		return err
	}
	nt.Time, nt.Valid = t, true
	return nil
}

// IsNull 检查是否为空
func (nt NullTime) IsNull() bool {
	return !nt.Valid
}

// Set 设置值
func (nt *NullTime) Set(t time.Time) {
	nt.Time, nt.Valid = t, true
}

// String 返回格式化的时间字符串
func (nt NullTime) String() string {
	if !nt.Valid {
		return ""
	}
	return nt.Time.Format("2006-01-02 15:04:05")
}

// parseTimeFormats 尝试多种格式解析时间字符串
func parseTimeFormats(s string) (time.Time, error) {
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
	return time.Time{}, fmt.Errorf("cannot parse time: %s", s)
}

// ============================================================================
// Decimal - 高精度十进制类型
// ============================================================================

// Decimal 高精度十进制类型（用于金额等需要精确计算的场景）
type Decimal struct {
	Str   string // 使用字符串存储以保持精度
	Valid bool
}

// NewDecimal 创建新的Decimal
func NewDecimal(value string) Decimal {
	return Decimal{Str: value, Valid: true}
}

// NewDecimalFromFloat 从浮点数创建Decimal
func NewDecimalFromFloat(f float64) Decimal {
	return Decimal{Str: strconv.FormatFloat(f, 'f', -1, 64), Valid: true}
}

// NewDecimalFromInt 从整数创建Decimal
func NewDecimalFromInt(i int64) Decimal {
	return Decimal{Str: strconv.FormatInt(i, 10), Valid: true}
}

// Scan 实现 sql.Scanner 接口
func (d *Decimal) Scan(value interface{}) error {
	if value == nil {
		d.Str, d.Valid = "", false
		return nil
	}
	switch v := value.(type) {
	case string:
		d.Str, d.Valid = v, true
	case []byte:
		d.Str, d.Valid = string(v), true
	case float64:
		d.Str, d.Valid = strconv.FormatFloat(v, 'f', -1, 64), true
	case int64:
		d.Str, d.Valid = strconv.FormatInt(v, 10), true
	default:
		d.Str, d.Valid = fmt.Sprintf("%v", v), true
	}
	return nil
}

// Value 实现 driver.Valuer 接口
func (d Decimal) Value() (driver.Value, error) {
	if !d.Valid {
		return nil, nil
	}
	return d.Str, nil
}

// MarshalJSON 实现 json.Marshaler 接口
func (d Decimal) MarshalJSON() ([]byte, error) {
	if !d.Valid {
		return []byte("null"), nil
	}
	// 返回数字而不是字符串
	return []byte(d.Str), nil
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (d *Decimal) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		d.Str, d.Valid = "", false
		return nil
	}
	// 尝试解析为数字
	s := strings.Trim(string(data), "\"")
	d.Str, d.Valid = s, true
	return nil
}

// Float64 转换为float64
func (d Decimal) Float64() (float64, error) {
	if !d.Valid {
		return 0, nil
	}
	return strconv.ParseFloat(d.Str, 64)
}

// Int64 转换为int64
func (d Decimal) Int64() (int64, error) {
	if !d.Valid {
		return 0, nil
	}
	f, err := strconv.ParseFloat(d.Str, 64)
	if err != nil {
		return 0, err
	}
	return int64(f), nil
}

// String 返回字符串表示
func (d Decimal) String() string {
	if !d.Valid {
		return ""
	}
	return d.Str
}

// IsNull 检查是否为空
func (d Decimal) IsNull() bool {
	return !d.Valid
}

// ============================================================================
// StringSlice - 字符串切片类型（存储为逗号分隔的字符串）
// ============================================================================

// StringSlice 字符串切片类型
type StringSlice []string

// Scan 实现 sql.Scanner 接口
func (ss *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*ss = nil
		return nil
	}
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("cannot scan type %T into StringSlice", value)
	}
	if s == "" {
		*ss = nil
		return nil
	}
	*ss = strings.Split(s, ",")
	return nil
}

// Value 实现 driver.Valuer 接口
func (ss StringSlice) Value() (driver.Value, error) {
	if ss == nil {
		return nil, nil
	}
	return strings.Join(ss, ","), nil
}

// MarshalJSON 实现 json.Marshaler 接口
func (ss StringSlice) MarshalJSON() ([]byte, error) {
	if ss == nil {
		return []byte("null"), nil
	}
	return json.Marshal([]string(ss))
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (ss *StringSlice) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*ss = nil
		return nil
	}
	var s []string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*ss = s
	return nil
}

// Contains 检查是否包含指定字符串
func (ss StringSlice) Contains(s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

// ============================================================================
// IntSlice - 整数切片类型（存储为逗号分隔的字符串）
// ============================================================================

// IntSlice 整数切片类型
type IntSlice []int64

// Scan 实现 sql.Scanner 接口
func (is *IntSlice) Scan(value interface{}) error {
	if value == nil {
		*is = nil
		return nil
	}
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("cannot scan type %T into IntSlice", value)
	}
	if s == "" {
		*is = nil
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]int64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		i, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return err
		}
		result = append(result, i)
	}
	*is = result
	return nil
}

// Value 实现 driver.Valuer 接口
func (is IntSlice) Value() (driver.Value, error) {
	if is == nil {
		return nil, nil
	}
	parts := make([]string, len(is))
	for i, v := range is {
		parts[i] = strconv.FormatInt(v, 10)
	}
	return strings.Join(parts, ","), nil
}

// MarshalJSON 实现 json.Marshaler 接口
func (is IntSlice) MarshalJSON() ([]byte, error) {
	if is == nil {
		return []byte("null"), nil
	}
	return json.Marshal([]int64(is))
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (is *IntSlice) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*is = nil
		return nil
	}
	var s []int64
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*is = s
	return nil
}

// Contains 检查是否包含指定整数
func (is IntSlice) Contains(i int64) bool {
	for _, v := range is {
		if v == i {
			return true
		}
	}
	return false
}

// ============================================================================
// Enum - 枚举类型
// ============================================================================

// Enum 枚举类型
type Enum struct {
	Str     string
	Valid   bool
	Options []string // 允许的选项
}

// NewEnum 创建新的Enum
func NewEnum(value string, options ...string) Enum {
	return Enum{Str: value, Valid: true, Options: options}
}

// Scan 实现 sql.Scanner 接口
func (e *Enum) Scan(value interface{}) error {
	if value == nil {
		e.Str, e.Valid = "", false
		return nil
	}
	switch v := value.(type) {
	case string:
		e.Str, e.Valid = v, true
	case []byte:
		e.Str, e.Valid = string(v), true
	default:
		return fmt.Errorf("cannot scan type %T into Enum", value)
	}
	return nil
}

// Value 实现 driver.Valuer 接口
func (e Enum) Value() (driver.Value, error) {
	if !e.Valid {
		return nil, nil
	}
	return e.Str, nil
}

// MarshalJSON 实现 json.Marshaler 接口
func (e Enum) MarshalJSON() ([]byte, error) {
	if !e.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(e.Str)
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (e *Enum) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		e.Str, e.Valid = "", false
		return nil
	}
	if err := json.Unmarshal(data, &e.Str); err != nil {
		return err
	}
	e.Valid = true
	return nil
}

// IsValid 检查值是否在允许的选项中
func (e Enum) IsValid() bool {
	if !e.Valid || len(e.Options) == 0 {
		return e.Valid
	}
	for _, opt := range e.Options {
		if opt == e.Str {
			return true
		}
	}
	return false
}

// Set 设置值
func (e *Enum) Set(value string) error {
	if len(e.Options) > 0 {
		valid := false
		for _, opt := range e.Options {
			if opt == value {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid enum value: %s, allowed: %v", value, e.Options)
		}
	}
	e.Str, e.Valid = value, true
	return nil
}

// GetValue 获取枚举值
func (e Enum) GetValue() string {
	return e.Str
}

// ============================================================================
// UUID - UUID类型
// ============================================================================

// UUID UUID类型
type UUID string

// Scan 实现 sql.Scanner 接口
func (u *UUID) Scan(value interface{}) error {
	if value == nil {
		*u = ""
		return nil
	}
	switch v := value.(type) {
	case string:
		*u = UUID(v)
	case []byte:
		*u = UUID(string(v))
	default:
		return fmt.Errorf("cannot scan type %T into UUID", value)
	}
	return nil
}

// Value 实现 driver.Valuer 接口
func (u UUID) Value() (driver.Value, error) {
	if u == "" {
		return nil, nil
	}
	return string(u), nil
}

// String 返回字符串表示
func (u UUID) String() string {
	return string(u)
}

// IsEmpty 检查是否为空
func (u UUID) IsEmpty() bool {
	return u == ""
}

// ============================================================================
// IP - IP地址类型
// ============================================================================

// IP IP地址类型
type IP string

// Scan 实现 sql.Scanner 接口
func (ip *IP) Scan(value interface{}) error {
	if value == nil {
		*ip = ""
		return nil
	}
	switch v := value.(type) {
	case string:
		*ip = IP(v)
	case []byte:
		*ip = IP(string(v))
	default:
		return fmt.Errorf("cannot scan type %T into IP", value)
	}
	return nil
}

// Value 实现 driver.Valuer 接口
func (ip IP) Value() (driver.Value, error) {
	if ip == "" {
		return nil, nil
	}
	return string(ip), nil
}

// String 返回字符串表示
func (ip IP) String() string {
	return string(ip)
}

// ============================================================================
// 类型处理工具函数
// ============================================================================

// IsCustomType 检查类型是否为自定义类型
func IsCustomType(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// 检查是否实现了Scanner和Valuer接口
	scannerType := reflect.TypeOf((*sql.Scanner)(nil)).Elem()
	valuerType := reflect.TypeOf((*driver.Valuer)(nil)).Elem()
	ptrType := reflect.PtrTo(t)
	return ptrType.Implements(scannerType) && (t.Implements(valuerType) || ptrType.Implements(valuerType))
}

// ProcessCustomTypeValue 处理自定义类型的值（用于写入数据库）
func ProcessCustomTypeValue(value interface{}) (driver.Value, error) {
	if value == nil {
		return nil, nil
	}
	// 检查是否实现了Valuer接口
	if valuer, ok := value.(driver.Valuer); ok {
		return valuer.Value()
	}
	// 检查是否有注册的转换器
	t := reflect.TypeOf(value)
	if converter, exists := GetCustomTypeConverter(t); exists {
		return converter.ConvertToDatabase(value)
	}
	return value, nil
}

// ScanCustomTypeValue 扫描自定义类型的值（用于从数据库读取）
func ScanCustomTypeValue(dest interface{}, value interface{}) error {
	if value == nil {
		return nil
	}
	// 检查是否实现了Scanner接口
	if scanner, ok := dest.(sql.Scanner); ok {
		return scanner.Scan(value)
	}
	// 检查是否有注册的转换器
	t := reflect.TypeOf(dest)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if converter, exists := GetCustomTypeConverter(t); exists {
		converted, err := converter.ConvertFromDatabase(value)
		if err != nil {
			return err
		}
		reflect.ValueOf(dest).Elem().Set(reflect.ValueOf(converted))
		return nil
	}
	return fmt.Errorf("cannot scan into type %T", dest)
}

// CreateCustomTypeInstance 创建自定义类型实例
func CreateCustomTypeInstance(t reflect.Type) interface{} {
	if t.Kind() == reflect.Ptr {
		return reflect.New(t.Elem()).Interface()
	}
	return reflect.New(t).Interface()
}

// GetCustomTypeName 获取自定义类型名称
func GetCustomTypeName(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

// ============================================================================
// 内置类型转换器
// ============================================================================

// jsonConverter JSON类型转换器
type jsonConverter struct{}

func (c *jsonConverter) ConvertToDatabase(value interface{}) (driver.Value, error) {
	return json.Marshal(value)
}

func (c *jsonConverter) ConvertFromDatabase(value interface{}) (interface{}, error) {
	var result interface{}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return nil, fmt.Errorf("cannot convert %T to JSON", value)
	}
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// timeConverter 时间类型转换器
type timeConverter struct {
	format string
}

func (c *timeConverter) ConvertToDatabase(value interface{}) (driver.Value, error) {
	if t, ok := value.(time.Time); ok {
		if c.format != "" {
			return t.Format(c.format), nil
		}
		return t, nil
	}
	return nil, fmt.Errorf("cannot convert %T to time", value)
}

func (c *timeConverter) ConvertFromDatabase(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case string:
		return parseTimeFormats(v)
	case []byte:
		return parseTimeFormats(string(v))
	case int64:
		return time.Unix(v, 0), nil
	}
	return nil, fmt.Errorf("cannot convert %T to time", value)
}

// RegisterBuiltinConverters 注册内置转换器
func RegisterBuiltinConverters() {
	RegisterCustomTypeByName("json", &jsonConverter{})
	RegisterCustomTypeByName("time", &timeConverter{format: "2006-01-02 15:04:05"})
	RegisterCustomTypeByName("date", &timeConverter{format: "2006-01-02"})
	RegisterCustomTypeByName("datetime", &timeConverter{format: "2006-01-02 15:04:05"})
}

// init 初始化时注册内置转换器
func init() {
	RegisterBuiltinConverters()
}
