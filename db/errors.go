package db

import (
	"errors"
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
	ErrCodeCanceled
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

// Category 返回错误类别
func (code ErrorCode) Category() string {
	return code.String()
}

// TormError TORM统一错误类型
type TormError struct {
	Code      ErrorCode              `json:"code"`              // 错误代码
	Message   string                 `json:"message"`           // 错误消息
	SQL       string                 `json:"sql,omitempty"`     // SQL语句
	Args      []interface{}          `json:"args,omitempty"`    // SQL参数
	Details   string                 `json:"details,omitempty"` // 详细信息
	Cause     error                  `json:"-"`                 // 原因错误
	Context   map[string]interface{} `json:"context,omitempty"` // 上下文信息
	Timestamp time.Time              `json:"timestamp"`         // 时间戳
	Stack     string                 `json:"stack,omitempty"`   // 调用栈
	mutex     sync.RWMutex           `json:"-"`                 // 并发安全保护
}

// Error 实现error接口
func (e *TormError) Error() string {
	var parts []string

	// 基础错误信息
	parts = append(parts, fmt.Sprintf("[%d] %s", e.Code, e.Message))

	// 添加SQL信息
	if e.SQL != "" {
		parts = append(parts, fmt.Sprintf("SQL: %s", e.SQL))
	}

	// 添加参数信息
	if len(e.Args) > 0 {
		parts = append(parts, fmt.Sprintf("Args: %v", e.Args))
	}

	// 添加详细信息
	if e.Details != "" {
		parts = append(parts, e.Details)
	}

	// 添加原因错误
	if e.Cause != nil {
		parts = append(parts, fmt.Sprintf("Cause: %v", e.Cause))
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

// As 支持errors.As
func (e *TormError) As(target interface{}) bool {
	if te, ok := target.(**TormError); ok {
		*te = e
		return true
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

// WithSQL 添加SQL语句和参数
func (e *TormError) WithSQL(sql string, args ...interface{}) *TormError {
	e.SQL = sql
	e.Args = args
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

// GetContext 获取上下文值
func (e *TormError) GetContext(key string) (interface{}, bool) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	if e.Context == nil {
		return nil, false
	}
	val, ok := e.Context[key]
	return val, ok
}

// GetSQL 获取SQL语句
func (e *TormError) GetSQL() string {
	return e.SQL
}

// GetArgs 获取SQL参数
func (e *TormError) GetArgs() []interface{} {
	return e.Args
}

// GetCode 获取错误代码
func (e *TormError) GetCode() ErrorCode {
	return e.Code
}

// GetMessage 获取错误消息
func (e *TormError) GetMessage() string {
	return e.Message
}

// GetCause 获取原因错误
func (e *TormError) GetCause() error {
	return e.Cause
}

// GetStack 获取调用栈
func (e *TormError) GetStack() string {
	return e.Stack
}

// Clone 克隆错误
func (e *TormError) Clone() *TormError {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	clone := &TormError{
		Code:      e.Code,
		Message:   e.Message,
		SQL:       e.SQL,
		Details:   e.Details,
		Cause:     e.Cause,
		Timestamp: e.Timestamp,
		Stack:     e.Stack,
	}

	if e.Args != nil {
		clone.Args = make([]interface{}, len(e.Args))
		copy(clone.Args, e.Args)
	}

	if e.Context != nil {
		clone.Context = make(map[string]interface{}, len(e.Context))
		for k, v := range e.Context {
			clone.Context[k] = v
		}
	}

	return clone
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

	// 如果已经是TormError，保留原始信息
	if te, ok := err.(*TormError); ok {
		clone := te.Clone()
		clone.Message = message + ": " + clone.Message
		return clone
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

// NewQueryError 创建查询错误（包含SQL和参数）
func NewQueryError(err error, sql string, args []interface{}) *TormError {
	return &TormError{
		Code:      ErrCodeQueryFailed,
		Message:   "query execution failed",
		SQL:       sql,
		Args:      args,
		Cause:     err,
		Timestamp: time.Now(),
		Stack:     getStack(),
	}
}

// NewConnectionError 创建连接错误
func NewConnectionError(err error, driver, host, database string) *TormError {
	return &TormError{
		Code:      ErrCodeConnectionFailed,
		Message:   "database connection failed",
		Cause:     err,
		Timestamp: time.Now(),
		Stack:     getStack(),
		Context: map[string]interface{}{
			"driver":   driver,
			"host":     host,
			"database": database,
		},
	}
}

// NewTransactionError 创建事务错误
func NewTransactionError(err error, operation string) *TormError {
	return &TormError{
		Code:      ErrCodeTransactionFailed,
		Message:   fmt.Sprintf("transaction %s failed", operation),
		Cause:     err,
		Timestamp: time.Now(),
		Stack:     getStack(),
		Context: map[string]interface{}{
			"operation": operation,
		},
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
		// 过滤掉标准库和测试框架的调用
		if strings.Contains(frame.File, "runtime/") ||
			strings.Contains(frame.File, "testing/") {
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
	ErrConnectionFailed   = NewError(ErrCodeConnectionFailed, "database connection failed")
	ErrConnectionClosed   = NewError(ErrCodeConnectionClosed, "database connection closed")
	ErrConnectionTimeout  = NewError(ErrCodeConnectionTimeout, "database connection timeout")
	ErrDriverNotSupported = NewError(ErrCodeDriverNotSupported, "unsupported database driver")

	// 查询错误
	ErrQueryFailed          = NewError(ErrCodeQueryFailed, "query execution failed")
	ErrQuerySyntaxError     = NewError(ErrCodeQuerySyntaxError, "SQL syntax error")
	ErrRecordNotFound       = NewError(ErrCodeRecordNotFound, "record not found")
	ErrMultipleRecordsFound = NewError(ErrCodeMultipleRecordsFound, "multiple records found, expected one")
	ErrDuplicateKey         = NewError(ErrCodeDuplicateKey, "duplicate key violation")

	// 事务错误
	ErrTransactionFailed         = NewError(ErrCodeTransactionFailed, "transaction failed")
	ErrTransactionCommitFailed   = NewError(ErrCodeTransactionCommitFailed, "transaction commit failed")
	ErrTransactionRollbackFailed = NewError(ErrCodeTransactionRollbackFailed, "transaction rollback failed")
	ErrTransactionNotStarted     = NewError(ErrCodeTransactionNotStarted, "transaction not started")
	ErrDeadlockDetected          = NewError(ErrCodeDeadlockDetected, "deadlock detected")

	// 模型错误
	ErrModelValidationFailed = NewError(ErrCodeModelValidationFailed, "model validation failed")
	ErrModelNotFound         = NewError(ErrCodeModelNotFound, "model not found")
	ErrModelSaveFailed       = NewError(ErrCodeModelSaveFailed, "model save failed")
	ErrInvalidModelState     = NewError(ErrCodeInvalidModelState, "invalid model state")

	// 参数错误
	ErrInvalidParameter = NewError(ErrCodeInvalidParameter, "invalid parameter")
	ErrValidationFailed = NewError(ErrCodeValidationFailed, "validation failed")

	// 超时和取消
	ErrTimeout  = NewError(ErrCodeTimeout, "operation timeout")
	ErrCanceled = NewError(ErrCodeCanceled, "operation canceled")

	// 缓存错误
	ErrCacheFailed      = NewError(ErrCodeCacheFailed, "cache operation failed")
	ErrCacheKeyNotFound = NewError(ErrCodeCacheKeyNotFound, "cache key not found")
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

// IsTimeoutError 检查是否为超时错误
func IsTimeoutError(err error) bool {
	if te, ok := err.(*TormError); ok {
		return te.Code == ErrCodeTimeout || te.Code == ErrCodeQueryTimeout || te.Code == ErrCodeConnectionTimeout
	}
	return false
}

// IsCanceledError 检查是否为取消错误
func IsCanceledError(err error) bool {
	if te, ok := err.(*TormError); ok {
		return te.Code == ErrCodeCanceled
	}
	return false
}

// ExtractTormError 从错误链中提取TormError
func ExtractTormError(err error) (*TormError, bool) {
	var te *TormError
	if errors.As(err, &te) {
		return te, true
	}
	return nil, false
}

// GetErrorCode 获取错误代码，如果不是TormError则返回ErrCodeUnknown
func GetErrorCode(err error) ErrorCode {
	if te, ok := err.(*TormError); ok {
		return te.Code
	}
	return ErrCodeUnknown
}

// ErrorLogger 错误日志记录器接口
type ErrorLogger interface {
	LogError(err *TormError)
}

// DefaultErrorLogger 默认错误日志记录器
type DefaultErrorLogger struct {
	debug bool
}

// NewDefaultErrorLogger 创建默认错误日志记录器
func NewDefaultErrorLogger(debug bool) *DefaultErrorLogger {
	return &DefaultErrorLogger{debug: debug}
}

// LogError 记录错误日志
func (l *DefaultErrorLogger) LogError(err *TormError) {
	fmt.Printf("[TORM ERROR] %s [%d] %s\n",
		err.Timestamp.Format("2006-01-02 15:04:05"),
		err.Code,
		err.Message)

	// 打印SQL信息
	if err.SQL != "" {
		fmt.Printf("  SQL: %s\n", err.SQL)
	}
	if len(err.Args) > 0 {
		fmt.Printf("  Args: %v\n", err.Args)
	}

	// 打印重要的上下文信息
	if err.Context != nil {
		if table, ok := err.Context["table"].(string); ok && table != "" {
			fmt.Printf("  Table: %s\n", table)
		}
		if operation, ok := err.Context["operation"].(string); ok && operation != "" {
			fmt.Printf("  Operation: %s\n", operation)
		}
	}

	// 打印原因错误
	if err.Cause != nil {
		fmt.Printf("  Cause: %v\n", err.Cause)
	}

	// 仅在调试模式下打印完整上下文和堆栈
	if l.debug {
		if len(err.Context) > 0 {
			fmt.Printf("  Context: %+v\n", err.Context)
		}
		if err.Stack != "" {
			fmt.Printf("  Stack:\n%s\n", err.Stack)
		}
	}
}

// 全局错误日志记录器
var globalErrorLogger ErrorLogger = &DefaultErrorLogger{debug: false}

// SetErrorLogger 设置全局错误日志记录器
func SetErrorLogger(logger ErrorLogger) {
	if logger != nil {
		globalErrorLogger = logger
	}
}

// LogError 记录错误到全局日志记录器
func LogError(err error) {
	if te, ok := err.(*TormError); ok && globalErrorLogger != nil {
		globalErrorLogger.LogError(te)
	}
}

// EnableDebugErrors 启用调试模式的错误日志
func EnableDebugErrors() {
	globalErrorLogger = &DefaultErrorLogger{debug: true}
}

// DisableDebugErrors 禁用调试模式的错误日志
func DisableDebugErrors() {
	globalErrorLogger = &DefaultErrorLogger{debug: false}
}
