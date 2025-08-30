package db

import (
	"reflect"
	"unsafe"
)

// ZeroCopyProcessor 零拷贝处理器
type ZeroCopyProcessor struct {
	stringPool   *StringPool
	bufferPool   *BufferPool
	reflectCache map[reflect.Type]*TypeInfo
}

// TypeInfo 类型信息缓存
type TypeInfo struct {
	Type        reflect.Type
	Fields      []FieldInfo
	Size        uintptr
	Align       uintptr
	IsStruct    bool
	HasPointers bool
}

// FieldInfo 字段信息
type FieldInfo struct {
	Name      string
	Type      reflect.Type
	Offset    uintptr
	Size      uintptr
	IsPointer bool
	IsString  bool
}

// StringPool 字符串池（零拷贝）
type StringPool struct {
	pool []string
	size int
	cap  int
}

// BufferPool 缓冲区池
type BufferPool struct {
	pool [][]byte
	size int
	cap  int
}

// NewZeroCopyProcessor 创建零拷贝处理器
func NewZeroCopyProcessor() *ZeroCopyProcessor {
	return &ZeroCopyProcessor{
		stringPool:   NewStringPool(1024),
		bufferPool:   NewBufferPool(1024, 4096),
		reflectCache: make(map[reflect.Type]*TypeInfo),
	}
}

// NewStringPool 创建字符串池
func NewStringPool(capacity int) *StringPool {
	return &StringPool{
		pool: make([]string, 0, capacity),
		cap:  capacity,
	}
}

// NewBufferPool 创建缓冲区池
func NewBufferPool(poolSize, bufferSize int) *BufferPool {
	pool := make([][]byte, 0, poolSize)
	for i := 0; i < poolSize; i++ {
		pool = append(pool, make([]byte, 0, bufferSize))
	}
	return &BufferPool{
		pool: pool,
		cap:  poolSize,
	}
}

// ZeroCopyStringToBytes 零拷贝字符串转字节数组
// 警告：返回的字节数组不能修改，否则会影响原字符串
func (zcp *ZeroCopyProcessor) ZeroCopyStringToBytes(s string) []byte {
	if s == "" {
		return nil
	}

	// 使用unsafe进行零拷贝转换
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

// ZeroCopyBytesToString 零拷贝字节数组转字符串
// 警告：原字节数组不能再修改，否则会影响返回的字符串
func (zcp *ZeroCopyProcessor) ZeroCopyBytesToString(b []byte) string {
	if len(b) == 0 {
		return ""
	}

	// 使用unsafe进行零拷贝转换
	return *(*string)(unsafe.Pointer(&b))
}

// ZeroCopySliceToSlice 零拷贝切片类型转换
func (zcp *ZeroCopyProcessor) ZeroCopySliceToSlice(src interface{}, dstType reflect.Type) interface{} {
	srcValue := reflect.ValueOf(src)
	if srcValue.Kind() != reflect.Slice {
		return src
	}

	srcHeader := (*reflect.SliceHeader)(unsafe.Pointer(srcValue.UnsafeAddr()))

	// 创建目标类型的切片
	dstSlice := reflect.New(dstType).Elem()
	dstHeader := (*reflect.SliceHeader)(unsafe.Pointer(dstSlice.UnsafeAddr()))

	dstHeader.Data = srcHeader.Data
	dstHeader.Len = srcHeader.Len
	dstHeader.Cap = srcHeader.Cap

	return dstSlice.Interface()
}

// ZeroCopyStructCopy 零拷贝结构体复制（相同类型）
func (zcp *ZeroCopyProcessor) ZeroCopyStructCopy(src, dst interface{}) error {
	srcValue := reflect.ValueOf(src)
	dstValue := reflect.ValueOf(dst)

	if srcValue.Type() != dstValue.Type() {
		return ErrTypeMismatch
	}

	if srcValue.Kind() == reflect.Ptr {
		srcValue = srcValue.Elem()
	}
	if dstValue.Kind() == reflect.Ptr {
		dstValue = dstValue.Elem()
	}

	if srcValue.Kind() != reflect.Struct || dstValue.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	// 获取类型信息
	typeInfo := zcp.getTypeInfo(srcValue.Type())

	// 使用内存拷贝进行零拷贝复制
	srcPtr := unsafe.Pointer(srcValue.UnsafeAddr())
	dstPtr := unsafe.Pointer(dstValue.UnsafeAddr())

	// 复制整个结构体内存
	copyMemory(dstPtr, srcPtr, typeInfo.Size)

	return nil
}

// ZeroCopyMapToStruct 零拷贝map到结构体转换
func (zcp *ZeroCopyProcessor) ZeroCopyMapToStruct(data map[string]interface{}, dst interface{}) error {
	dstValue := reflect.ValueOf(dst)
	if dstValue.Kind() != reflect.Ptr {
		return ErrNotPointer
	}

	dstElem := dstValue.Elem()
	if dstElem.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	typeInfo := zcp.getTypeInfo(dstElem.Type())

	// 直接写入结构体内存
	structPtr := unsafe.Pointer(dstElem.UnsafeAddr())

	for _, field := range typeInfo.Fields {
		if value, exists := data[field.Name]; exists {
			fieldPtr := unsafe.Pointer(uintptr(structPtr) + field.Offset)

			if field.IsString {
				zcp.setStringField(fieldPtr, value)
			} else {
				zcp.setField(fieldPtr, field.Type, value)
			}
		}
	}

	return nil
}

// ZeroCopyStructToMap 零拷贝结构体到map转换
func (zcp *ZeroCopyProcessor) ZeroCopyStructToMap(src interface{}) (map[string]interface{}, error) {
	srcValue := reflect.ValueOf(src)
	if srcValue.Kind() == reflect.Ptr {
		srcValue = srcValue.Elem()
	}

	if srcValue.Kind() != reflect.Struct {
		return nil, ErrNotStruct
	}

	typeInfo := zcp.getTypeInfo(srcValue.Type())
	result := make(map[string]interface{}, len(typeInfo.Fields))

	structPtr := unsafe.Pointer(srcValue.UnsafeAddr())

	for _, field := range typeInfo.Fields {
		fieldPtr := unsafe.Pointer(uintptr(structPtr) + field.Offset)

		if field.IsString {
			value := zcp.getStringField(fieldPtr)
			result[field.Name] = value
		} else {
			value := zcp.getField(fieldPtr, field.Type)
			result[field.Name] = value
		}
	}

	return result, nil
}

// getTypeInfo 获取类型信息（缓存）
func (zcp *ZeroCopyProcessor) getTypeInfo(t reflect.Type) *TypeInfo {
	if info, exists := zcp.reflectCache[t]; exists {
		return info
	}

	info := &TypeInfo{
		Type:        t,
		Size:        t.Size(),
		Align:       uintptr(t.Align()),
		IsStruct:    t.Kind() == reflect.Struct,
		HasPointers: hasPointers(t),
		Fields:      make([]FieldInfo, 0, t.NumField()),
	}

	if info.IsStruct {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fieldInfo := FieldInfo{
				Name:      field.Name,
				Type:      field.Type,
				Offset:    field.Offset,
				Size:      field.Type.Size(),
				IsPointer: field.Type.Kind() == reflect.Ptr,
				IsString:  field.Type.Kind() == reflect.String,
			}
			info.Fields = append(info.Fields, fieldInfo)
		}
	}

	zcp.reflectCache[t] = info
	return info
}

// setStringField 设置字符串字段（零拷贝）
func (zcp *ZeroCopyProcessor) setStringField(fieldPtr unsafe.Pointer, value interface{}) {
	var str string

	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = zcp.ZeroCopyBytesToString(v)
	default:
		str = ""
	}

	// 直接写入字符串结构
	strHeader := (*reflect.StringHeader)(fieldPtr)
	srcHeader := (*reflect.StringHeader)(unsafe.Pointer(&str))

	strHeader.Data = srcHeader.Data
	strHeader.Len = srcHeader.Len
}

// getStringField 获取字符串字段（零拷贝）
func (zcp *ZeroCopyProcessor) getStringField(fieldPtr unsafe.Pointer) string {
	strHeader := (*reflect.StringHeader)(fieldPtr)
	return *(*string)(unsafe.Pointer(strHeader))
}

// setField 设置字段值
func (zcp *ZeroCopyProcessor) setField(fieldPtr unsafe.Pointer, fieldType reflect.Type, value interface{}) {
	switch fieldType.Kind() {
	case reflect.Int:
		*(*int)(fieldPtr) = value.(int)
	case reflect.Int8:
		*(*int8)(fieldPtr) = value.(int8)
	case reflect.Int16:
		*(*int16)(fieldPtr) = value.(int16)
	case reflect.Int32:
		*(*int32)(fieldPtr) = value.(int32)
	case reflect.Int64:
		*(*int64)(fieldPtr) = value.(int64)
	case reflect.Uint:
		*(*uint)(fieldPtr) = value.(uint)
	case reflect.Uint8:
		*(*uint8)(fieldPtr) = value.(uint8)
	case reflect.Uint16:
		*(*uint16)(fieldPtr) = value.(uint16)
	case reflect.Uint32:
		*(*uint32)(fieldPtr) = value.(uint32)
	case reflect.Uint64:
		*(*uint64)(fieldPtr) = value.(uint64)
	case reflect.Float32:
		*(*float32)(fieldPtr) = value.(float32)
	case reflect.Float64:
		*(*float64)(fieldPtr) = value.(float64)
	case reflect.Bool:
		*(*bool)(fieldPtr) = value.(bool)
	}
}

// getField 获取字段值
func (zcp *ZeroCopyProcessor) getField(fieldPtr unsafe.Pointer, fieldType reflect.Type) interface{} {
	switch fieldType.Kind() {
	case reflect.Int:
		return *(*int)(fieldPtr)
	case reflect.Int8:
		return *(*int8)(fieldPtr)
	case reflect.Int16:
		return *(*int16)(fieldPtr)
	case reflect.Int32:
		return *(*int32)(fieldPtr)
	case reflect.Int64:
		return *(*int64)(fieldPtr)
	case reflect.Uint:
		return *(*uint)(fieldPtr)
	case reflect.Uint8:
		return *(*uint8)(fieldPtr)
	case reflect.Uint16:
		return *(*uint16)(fieldPtr)
	case reflect.Uint32:
		return *(*uint32)(fieldPtr)
	case reflect.Uint64:
		return *(*uint64)(fieldPtr)
	case reflect.Float32:
		return *(*float32)(fieldPtr)
	case reflect.Float64:
		return *(*float64)(fieldPtr)
	case reflect.Bool:
		return *(*bool)(fieldPtr)
	default:
		return nil
	}
}

// 内存拷贝
func copyMemory(dst, src unsafe.Pointer, size uintptr) {
	// 使用汇编或runtime.memmove进行高效内存拷贝
	for i := uintptr(0); i < size; i++ {
		*(*byte)(unsafe.Pointer(uintptr(dst) + i)) = *(*byte)(unsafe.Pointer(uintptr(src) + i))
	}
}

// hasPointers 检查类型是否包含指针
func hasPointers(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Chan, reflect.Func, reflect.Interface, reflect.Slice:
		return true
	case reflect.Array:
		return hasPointers(t.Elem())
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if hasPointers(t.Field(i).Type) {
				return true
			}
		}
	}
	return false
}

// 错误定义
var (
	ErrTypeMismatch = NewError(ErrCodeValidationFailed, "类型不匹配，零拷贝操作要求相同类型")
	ErrNotStruct    = NewError(ErrCodeValidationFailed, "不是结构体，零拷贝操作要求结构体类型")
	ErrNotPointer   = NewError(ErrCodeValidationFailed, "不是指针，零拷贝操作要求指针类型")
)
