package db

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

// ErrorCode 错误代码类型
type ErrorCode int

const (
	// 通用错误 1000-1999
	ErrCodeUnknown ErrorCode = 1000 + iota
	ErrCodeInvalidParameter
	ErrCodeConfigurationError
	ErrCodeValidationFailed
	ErrCodeTimeout
	ErrCodeNotImplemented

	// 连接错误 2000-2999
	ErrCodeConnectionFailed ErrorCode = 2000 + iota
	ErrCodeConnectionClosed
	ErrCodeConnectionTimeout
	ErrCodeConnectionPoolExhausted
	ErrCodeDriverNotSupported

	// 查询错误 3000-3999
	ErrCodeQueryFailed ErrorCode = 3000 + iota
	ErrCodeQuerySyntaxError
	ErrCodeQueryTimeout
	ErrCodeRecordNotFound
	ErrCodeMultipleRecordsFound
	ErrCodeDuplicateKey

	// 事务错误 4000-4999
	ErrCodeTransactionFailed ErrorCode = 4000 + iota
	ErrCodeTransactionCommitFailed
	ErrCodeTransactionRollbackFailed
	ErrCodeTransactionAlreadyStarted
	ErrCodeTransactionNotStarted
	ErrCodeDeadlockDetected

	// 模型错误 5000-5999
	ErrCodeModelValidationFailed ErrorCode = 5000 + iota
	ErrCodeModelNotFound
	ErrCodeModelSaveFailed
	ErrCodeModelDeleteFailed
	ErrCodeInvalidModelState
	ErrCodeRelationshipError

	// 迁移错误 6000-6999
	ErrCodeMigrationFailed ErrorCode = 6000 + iota
	ErrCodeMigrationVersionConflict
	ErrCodeMigrationRollbackFailed
	ErrCodeSchemaError

	// 缓存错误 7000-7999
	ErrCodeCacheFailed ErrorCode = 7000 + iota
	ErrCodeCacheKeyNotFound
	ErrCodeCacheExpired
	ErrCodeCacheConnectionFailed
)

// String 返回错误代码字符串
func (code ErrorCode) String() string {
	switch {
	case code >= 1000 && code < 2000:
		return "GENERAL_ERROR"
	case code >= 2000 && code < 3000:
		return "CONNECTION_ERROR"
	case code >= 3000 && code < 4000:
		return "QUERY_ERROR"
	case code >= 4000 && code < 5000:
		return "TRANSACTION_ERROR"
	case code >= 5000 && code < 6000:
		return "MODEL_ERROR"
	case code >= 6000 && code < 7000:
		return "MIGRATION_ERROR"
	case code >= 7000 && code < 8000:
		return "CACHE_ERROR"
	default:
		return "UNKNOWN_ERROR"
	}
}

// TormError TORM统一错误类型
type TormError struct {
	Code      ErrorCode              `json:"code"`
	Message   string                 `json:"message"`
	Details   string                 `json:"details,omitempty"`
	Cause     error                  `json:"cause,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Stack     string                 `json:"stack,omitempty"`
	mutex     sync.RWMutex           `json:"-"` // 并发安全保护
}

// Error 实现error接口
func (e *TormError) Error() string {
	var parts []string

	// 基础错误信息
	parts = append(parts, fmt.Sprintf("[%d] %s", e.Code, e.Message))

	// 添加详细信息
	if e.Details != "" {
		parts = append(parts, e.Details)
	}

	// 添加原因错误
	if e.Cause != nil {
		parts = append(parts, fmt.Sprintf("原因: %v", e.Cause))
	}

	// 添加重要的上下文信息
	e.mutex.RLock()
	contextCopy := make(map[string]interface{})
	if e.Context != nil {
		for k, v := range e.Context {
			contextCopy[k] = v
		}
	}
	e.mutex.RUnlock()

	if len(contextCopy) > 0 {
		var contextParts []string

		// 优先显示SQL和参数
		if sql, ok := contextCopy["sql"].(string); ok && sql != "" {
			contextParts = append(contextParts, fmt.Sprintf("SQL: %s", sql))
		}
		if args, ok := contextCopy["args"]; ok && args != nil {
			contextParts = append(contextParts, fmt.Sprintf("参数: %v", args))
		}
		if table, ok := contextCopy["table"].(string); ok && table != "" {
			contextParts = append(contextParts, fmt.Sprintf("表: %s", table))
		}

		// 添加其他重要上下文
		for key, value := range contextCopy {
			if key != "sql" && key != "args" && key != "table" {
				if str := fmt.Sprintf("%v", value); len(str) < 100 { // 限制长度
					contextParts = append(contextParts, fmt.Sprintf("%s: %v", key, value))
				}
			}
		}

		if len(contextParts) > 0 {
			parts = append(parts, fmt.Sprintf("上下文: {%s}", strings.Join(contextParts, ", ")))
		}
	}

	return strings.Join(parts, " | ")
}

// Unwrap 支持errors.Unwrap
func (e *TormError) Unwrap() error {
	return e.Cause
}

// Is 支持errors.Is
func (e *TormError) Is(target error) bool {
	if te, ok := target.(*TormError); ok {
		return e.Code == te.Code
	}
	return false
}

// WithContext 添加上下文信息
func (e *TormError) WithContext(key string, value interface{}) *TormError {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithDetails 添加详细信息
func (e *TormError) WithDetails(details string) *TormError {
	e.Details = details
	return e
}

// WithCause 添加原因错误
func (e *TormError) WithCause(cause error) *TormError {
	e.Cause = cause
	return e
}

// NewError 创建新的TORM错误
func NewError(code ErrorCode, message string) *TormError {
	return &TormError{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
		Stack:     getStack(),
	}
}

// NewErrorf 创建格式化的TORM错误
func NewErrorf(code ErrorCode, format string, args ...interface{}) *TormError {
	return &TormError{
		Code:      code,
		Message:   fmt.Sprintf(format, args...),
		Timestamp: time.Now(),
		Stack:     getStack(),
	}
}

// WrapError 包装现有错误
func WrapError(err error, code ErrorCode, message string) *TormError {
	if err == nil {
		return nil
	}

	return &TormError{
		Code:      code,
		Message:   message,
		Cause:     err,
		Timestamp: time.Now(),
		Stack:     getStack(),
	}
}

// WrapErrorf 包装现有错误（格式化消息）
func WrapErrorf(err error, code ErrorCode, format string, args ...interface{}) *TormError {
	if err == nil {
		return nil
	}

	return &TormError{
		Code:      code,
		Message:   fmt.Sprintf(format, args...),
		Cause:     err,
		Timestamp: time.Now(),
		Stack:     getStack(),
	}
}

// getStack 获取调用栈
func getStack() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var stack strings.Builder
	for {
		frame, more := frames.Next()
		if !strings.Contains(frame.File, "torm") {
			if !more {
				break
			}
			continue
		}
		stack.WriteString(fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function))
		if !more {
			break
		}
	}
	return stack.String()
}

// 预定义的常用错误
var (
	// 连接错误
	ErrConnectionFailed   = NewError(ErrCodeConnectionFailed, "数据库连接失败")
	ErrConnectionClosed   = NewError(ErrCodeConnectionClosed, "数据库连接已关闭")
	ErrConnectionTimeout  = NewError(ErrCodeConnectionTimeout, "数据库连接超时")
	ErrDriverNotSupported = NewError(ErrCodeDriverNotSupported, "不支持的数据库驱动")

	// 查询错误
	ErrQueryFailed          = NewError(ErrCodeQueryFailed, "查询执行失败")
	ErrQuerySyntaxError     = NewError(ErrCodeQuerySyntaxError, "SQL语法错误")
	ErrRecordNotFound       = NewError(ErrCodeRecordNotFound, "记录不存在")
	ErrMultipleRecordsFound = NewError(ErrCodeMultipleRecordsFound, "找到多条记录，期望只有一条")
	ErrDuplicateKey         = NewError(ErrCodeDuplicateKey, "违反唯一性约束")

	// 事务错误
	ErrTransactionFailed         = NewError(ErrCodeTransactionFailed, "事务执行失败")
	ErrTransactionCommitFailed   = NewError(ErrCodeTransactionCommitFailed, "事务提交失败")
	ErrTransactionRollbackFailed = NewError(ErrCodeTransactionRollbackFailed, "事务回滚失败")
	ErrDeadlockDetected          = NewError(ErrCodeDeadlockDetected, "检测到死锁")

	// 模型错误
	ErrModelValidationFailed = NewError(ErrCodeModelValidationFailed, "模型验证失败")
	ErrModelNotFound         = NewError(ErrCodeModelNotFound, "模型不存在")
	ErrModelSaveFailed       = NewError(ErrCodeModelSaveFailed, "模型保存失败")
	ErrInvalidModelState     = NewError(ErrCodeInvalidModelState, "无效的模型状态")

	// 参数错误
	ErrInvalidParameter = NewError(ErrCodeInvalidParameter, "无效的参数")
	ErrValidationFailed = NewError(ErrCodeValidationFailed, "数据验证失败")

	// 缓存错误
	ErrCacheFailed      = NewError(ErrCodeCacheFailed, "缓存操作失败")
	ErrCacheKeyNotFound = NewError(ErrCodeCacheKeyNotFound, "缓存键不存在")
)

// IsConnectionError 检查是否为连接错误
func IsConnectionError(err error) bool {
	if te, ok := err.(*TormError); ok {
		return te.Code >= 2000 && te.Code < 3000
	}
	return false
}

// IsQueryError 检查是否为查询错误
func IsQueryError(err error) bool {
	if te, ok := err.(*TormError); ok {
		return te.Code >= 3000 && te.Code < 4000
	}
	return false
}

// IsTransactionError 检查是否为事务错误
func IsTransactionError(err error) bool {
	if te, ok := err.(*TormError); ok {
		return te.Code >= 4000 && te.Code < 5000
	}
	return false
}

// IsModelError 检查是否为模型错误
func IsModelError(err error) bool {
	if te, ok := err.(*TormError); ok {
		return te.Code >= 5000 && te.Code < 6000
	}
	return false
}

// IsNotFoundError 检查是否为记录不存在错误
func IsNotFoundError(err error) bool {
	if te, ok := err.(*TormError); ok {
		return te.Code == ErrCodeRecordNotFound
	}
	return false
}

// IsDuplicateError 检查是否为重复键错误
func IsDuplicateError(err error) bool {
	if te, ok := err.(*TormError); ok {
		return te.Code == ErrCodeDuplicateKey
	}
	return false
}

// ErrorLogger 错误日志记录器
type ErrorLogger interface {
	LogError(err *TormError)
}

// DefaultErrorLogger 默认错误日志记录器
type DefaultErrorLogger struct{}

// LogError 记录错误日志
func (l *DefaultErrorLogger) LogError(err *TormError) {
	fmt.Printf("[TORM ERROR] %s %s\n", time.Now().Format("2006-01-02 15:04:05"), err.Error())

	// 打印重要的上下文信息
	if err.Context != nil {
		if sql, ok := err.Context["sql"].(string); ok && sql != "" {
			fmt.Printf("  SQL: %s\n", sql)
		}
		if args, ok := err.Context["args"]; ok && args != nil {
			fmt.Printf("  参数: %v\n", args)
		}
		if table, ok := err.Context["table"].(string); ok && table != "" {
			fmt.Printf("  表: %s\n", table)
		}
		if operation, ok := err.Context["operation"].(string); ok && operation != "" {
			fmt.Printf("  操作: %s\n", operation)
		}
	}

	// 打印原因错误
	if err.Cause != nil {
		fmt.Printf("  根本原因: %v\n", err.Cause)
	}

	// 仅在调试模式下打印完整上下文和堆栈
	if isDebugMode() {
		if err.Context != nil {
			fmt.Printf("  完整上下文: %+v\n", err.Context)
		}
		if err.Stack != "" {
			fmt.Printf("  调用栈:\n%s\n", err.Stack)
		}
	}
}

// isDebugMode 检查是否为调试模式
func isDebugMode() bool {
	// 可以通过环境变量或其他方式控制
	// 这里简化处理，可以根据需要扩展
	return false
}

// 全局错误日志记录器
var globalErrorLogger ErrorLogger = &DefaultErrorLogger{}

// SetErrorLogger 设置全局错误日志记录器
func SetErrorLogger(logger ErrorLogger) {
	globalErrorLogger = logger
}

// LogError 记录错误到全局日志记录器
func LogError(err error) {
	if te, ok := err.(*TormError); ok && globalErrorLogger != nil {
		globalErrorLogger.LogError(te)
	}
}
