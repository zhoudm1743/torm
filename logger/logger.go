package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	DEBUG  LogLevel = iota // 0 - 调试级别（最详细）
	INFO                   // 1 - 信息级别
	WARN                   // 2 - 警告级别
	ERROR                  // 3 - 错误级别
	FATAL                  // 4 - 致命错误级别（最严重）
	SILENT                 // 5 - 静默级别（不输出任何日志）
)

// String 返回日志级别字符串
func (level LogLevel) String() string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	case SILENT:
		return "SILENT"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel 解析日志级别字符串
func ParseLogLevel(s string) LogLevel {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	case "SILENT":
		return SILENT
	default:
		return INFO
	}
}

// LoggerInterface 日志接口
// 所有日志实现必须实现此接口
type LoggerInterface interface {
	// 基础日志方法
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})

	// 级别控制
	SetLevel(level LogLevel)
	GetLevel() LogLevel

	// SQL日志
	LogSQL(sql string, args []interface{}, duration time.Duration, err error)
	LogSQLWithRows(sql string, args []interface{}, duration time.Duration, rowsAffected int64, err error)

	// 事务日志
	LogTransaction(action string, duration time.Duration)

	// 连接日志
	LogConnection(action string, driver string, database string, duration time.Duration)

	// 查询日志控制
	EnableQueryLogging()
	DisableQueryLogging()
	IsQueryLoggingEnabled() bool
}

// Logger 基础日志记录器
type Logger struct {
	level      LogLevel
	output     io.Writer
	logger     *log.Logger
	showCaller bool
	mu         sync.RWMutex
}

// NewLogger 创建日志记录器
func NewLogger(level LogLevel) *Logger {
	return &Logger{
		level:      level,
		output:     os.Stdout,
		logger:     log.New(os.Stdout, "", 0),
		showCaller: false,
	}
}

// NewLoggerWithOutput 创建带自定义输出的日志记录器
func NewLoggerWithOutput(level LogLevel, output io.Writer) *Logger {
	return &Logger{
		level:      level,
		output:     output,
		logger:     log.New(output, "", 0),
		showCaller: false,
	}
}

// NewFileLogger 创建文件日志记录器
func NewFileLogger(level LogLevel, filename string) (*Logger, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Logger{
		level:      level,
		output:     file,
		logger:     log.New(file, "", 0),
		showCaller: false,
	}, nil
}

// SetShowCaller 设置是否显示调用者信息
func (l *Logger) SetShowCaller(show bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.showCaller = show
}

// SetOutput 设置输出目标
func (l *Logger) SetOutput(output io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = output
	l.logger = log.New(output, "", 0)
}

// shouldLog 检查是否应该输出日志
func (l *Logger) shouldLog(level LogLevel) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return level >= l.level && l.level != SILENT
}

// formatMessage 格式化消息
func (l *Logger) formatMessage(level LogLevel, msg string, fields ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	var message string
	if len(fields) > 0 {
		// 将fields转换为key=value格式
		var parts []string
		for i := 0; i < len(fields); i += 2 {
			if i+1 < len(fields) {
				parts = append(parts, fmt.Sprintf("%v=%v", fields[i], fields[i+1]))
			} else {
				parts = append(parts, fmt.Sprintf("%v", fields[i]))
			}
		}
		if len(parts) > 0 {
			message = fmt.Sprintf("%s [%s] %s [%s]", timestamp, level.String(), msg, strings.Join(parts, " "))
		} else {
			message = fmt.Sprintf("%s [%s] %s", timestamp, level.String(), msg)
		}
	} else {
		message = fmt.Sprintf("%s [%s] %s", timestamp, level.String(), msg)
	}

	return message
}

// log 内部日志方法
func (l *Logger) log(level LogLevel, msg string, fields ...interface{}) {
	if !l.shouldLog(level) {
		return
	}

	message := l.formatMessage(level, msg, fields...)
	l.logger.Println(message)

	// 如果是FATAL级别，退出程序
	if level == FATAL {
		os.Exit(1)
	}
}

// Debug 调试日志
func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.log(DEBUG, msg, fields...)
}

// Info 信息日志
func (l *Logger) Info(msg string, fields ...interface{}) {
	l.log(INFO, msg, fields...)
}

// Warn 警告日志
func (l *Logger) Warn(msg string, fields ...interface{}) {
	l.log(WARN, msg, fields...)
}

// Error 错误日志
func (l *Logger) Error(msg string, fields ...interface{}) {
	l.log(ERROR, msg, fields...)
}

// Fatal 致命错误日志
func (l *Logger) Fatal(msg string, fields ...interface{}) {
	l.log(FATAL, msg, fields...)
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel 获取日志级别
func (l *Logger) GetLevel() LogLevel {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

// SQLLogger SQL查询日志记录器
type SQLLogger struct {
	*Logger
	logQueries bool
	mu         sync.RWMutex
}

// NewSQLLogger 创建SQL日志记录器
func NewSQLLogger(level LogLevel, logQueries bool) *SQLLogger {
	return &SQLLogger{
		Logger:     NewLogger(level),
		logQueries: logQueries,
	}
}

// NewSQLLoggerWithOutput 创建带自定义输出的SQL日志记录器
func NewSQLLoggerWithOutput(level LogLevel, logQueries bool, output io.Writer) *SQLLogger {
	return &SQLLogger{
		Logger:     NewLoggerWithOutput(level, output),
		logQueries: logQueries,
	}
}

// NewSQLFileLogger 创建SQL文件日志记录器
func NewSQLFileLogger(level LogLevel, logQueries bool, filename string) (*SQLLogger, error) {
	baseLogger, err := NewFileLogger(level, filename)
	if err != nil {
		return nil, err
	}

	return &SQLLogger{
		Logger:     baseLogger,
		logQueries: logQueries,
	}, nil
}

// LogSQL 统一记录SQL执行日志
func (l *SQLLogger) LogSQL(sql string, args []interface{}, duration time.Duration, err error) {
	if err != nil {
		// 错误格式：[ERROR] [耗时] SQL语句 | 错误信息
		finalSQL := FormatSQLWithArgs(sql, args)
		errorMsg := fmt.Sprintf("[%v] %s | Error: %v", duration, finalSQL, err)
		l.Error(errorMsg)
	} else if l.IsQueryLoggingEnabled() {
		// 成功格式：[INFO] [耗时] SQL语句 (参数已替换)
		finalSQL := FormatSQLWithArgs(sql, args)
		successMsg := fmt.Sprintf("[%v] %s", duration, finalSQL)
		l.Info(successMsg)
	}
}

// LogSQLWithRows 记录SQL执行日志（包含行数信息）
func (l *SQLLogger) LogSQLWithRows(sql string, args []interface{}, duration time.Duration, rowsAffected int64, err error) {
	if err != nil {
		// 错误格式：[ERROR] [耗时] SQL语句 | 错误信息
		finalSQL := FormatSQLWithArgs(sql, args)
		errorMsg := fmt.Sprintf("[%v] %s | Error: %v", duration, finalSQL, err)
		l.Error(errorMsg)
	} else if l.IsQueryLoggingEnabled() {
		// 成功格式：[INFO] [耗时] [rows:affected] SQL语句 (参数已替换)
		finalSQL := FormatSQLWithArgs(sql, args)
		successMsg := fmt.Sprintf("[%v] [rows:%d] %s", duration, rowsAffected, finalSQL)
		l.Info(successMsg)
	}
}

// LogTransaction 记录事务操作
func (l *SQLLogger) LogTransaction(action string, duration time.Duration) {
	if !l.IsQueryLoggingEnabled() {
		return
	}

	l.Debug(fmt.Sprintf("Transaction %s [%v]", action, duration))
}

// LogConnection 记录连接操作
func (l *SQLLogger) LogConnection(action string, driver string, database string, duration time.Duration) {
	l.Info(fmt.Sprintf("Connection %s", action),
		"driver", driver,
		"database", database,
		"duration", duration.String())
}

// EnableQueryLogging 启用查询日志
func (l *SQLLogger) EnableQueryLogging() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logQueries = true
}

// DisableQueryLogging 禁用查询日志
func (l *SQLLogger) DisableQueryLogging() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logQueries = false
}

// IsQueryLoggingEnabled 检查查询日志是否启用
func (l *SQLLogger) IsQueryLoggingEnabled() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.logQueries
}

// FormatSQLWithArgs 将参数替换到SQL中（用于日志显示）
func FormatSQLWithArgs(sql string, args []interface{}) string {
	if len(args) == 0 {
		return sql
	}

	result := sql
	for i, arg := range args {
		var replacement string
		switch v := arg.(type) {
		case string:
			replacement = fmt.Sprintf("'%s'", escapeString(v))
		case nil:
			replacement = "NULL"
		case time.Time:
			replacement = fmt.Sprintf("'%s'", v.Format("2006-01-02 15:04:05"))
		case []byte:
			replacement = fmt.Sprintf("'%s'", string(v))
		case bool:
			if v {
				replacement = "TRUE"
			} else {
				replacement = "FALSE"
			}
		default:
			replacement = fmt.Sprintf("%v", v)
		}

		// 替换 ? 占位符
		result = strings.Replace(result, "?", replacement, 1)

		// 替换 $N 占位符 (PostgreSQL)
		placeholder := fmt.Sprintf("$%d", i+1)
		result = strings.Replace(result, placeholder, replacement, 1)
	}

	return result
}

// escapeString 转义字符串中的特殊字符
func escapeString(s string) string {
	s = strings.ReplaceAll(s, "'", "''")
	s = strings.ReplaceAll(s, "\\", "\\\\")
	return s
}

// NullLogger 空日志记录器（不输出任何内容）
type NullLogger struct{}

func (l *NullLogger) Debug(msg string, fields ...interface{})                                  {}
func (l *NullLogger) Info(msg string, fields ...interface{})                                   {}
func (l *NullLogger) Warn(msg string, fields ...interface{})                                   {}
func (l *NullLogger) Error(msg string, fields ...interface{})                                  {}
func (l *NullLogger) Fatal(msg string, fields ...interface{})                                  { os.Exit(1) }
func (l *NullLogger) SetLevel(level LogLevel)                                                  {}
func (l *NullLogger) GetLevel() LogLevel                                                       { return SILENT }
func (l *NullLogger) LogSQL(sql string, args []interface{}, duration time.Duration, err error) {}
func (l *NullLogger) LogSQLWithRows(sql string, args []interface{}, d time.Duration, r int64, e error) {
}
func (l *NullLogger) LogTransaction(action string, duration time.Duration)                         {}
func (l *NullLogger) LogConnection(action string, driver string, database string, d time.Duration) {}
func (l *NullLogger) EnableQueryLogging()                                                          {}
func (l *NullLogger) DisableQueryLogging()                                                         {}
func (l *NullLogger) IsQueryLoggingEnabled() bool                                                  { return false }

// NewNullLogger 创建空日志记录器
func NewNullLogger() LoggerInterface {
	return &NullLogger{}
}

// 默认日志记录器实例
var (
	defaultLogger     LoggerInterface = NewSQLLogger(INFO, true)
	defaultLoggerOnce sync.Once
	defaultLoggerMu   sync.RWMutex
)

// DefaultLogger 获取默认日志记录器
func DefaultLogger() LoggerInterface {
	defaultLoggerMu.RLock()
	defer defaultLoggerMu.RUnlock()
	return defaultLogger
}

// SetDefaultLogger 设置默认日志记录器
func SetDefaultLogger(logger LoggerInterface) {
	defaultLoggerMu.Lock()
	defer defaultLoggerMu.Unlock()
	if logger != nil {
		defaultLogger = logger
	}
}

// SetDefaultLevel 设置默认日志级别
func SetDefaultLevel(level LogLevel) {
	defaultLoggerMu.Lock()
	defer defaultLoggerMu.Unlock()
	defaultLogger.SetLevel(level)
}

// 便捷函数，使用默认日志记录器

// Debug 调试日志
func Debug(msg string, fields ...interface{}) {
	DefaultLogger().Debug(msg, fields...)
}

// Info 信息日志
func Info(msg string, fields ...interface{}) {
	DefaultLogger().Info(msg, fields...)
}

// Warn 警告日志
func Warn(msg string, fields ...interface{}) {
	DefaultLogger().Warn(msg, fields...)
}

// Error 错误日志
func Error(msg string, fields ...interface{}) {
	DefaultLogger().Error(msg, fields...)
}

// Fatal 致命错误日志
func Fatal(msg string, fields ...interface{}) {
	DefaultLogger().Fatal(msg, fields...)
}

// 预设的日志记录器工厂函数

// NewDebugLogger 创建DEBUG级别的日志记录器（显示所有日志）
func NewDebugLogger() LoggerInterface {
	return NewSQLLogger(DEBUG, true)
}

// NewInfoLogger 创建INFO级别的日志记录器
func NewInfoLogger() LoggerInterface {
	return NewSQLLogger(INFO, true)
}

// NewQuietLogger 创建WARN级别的日志记录器（较少输出）
func NewQuietLogger() LoggerInterface {
	return NewSQLLogger(WARN, false)
}

// NewSilentLogger 创建静默日志记录器（不输出任何内容）
func NewSilentLogger() LoggerInterface {
	return NewNullLogger()
}

// NewProductionLogger 创建生产环境日志记录器
func NewProductionLogger() LoggerInterface {
	return NewSQLLogger(WARN, false)
}

// NewDevelopmentLogger 创建开发环境日志记录器
func NewDevelopmentLogger() LoggerInterface {
	logger := NewSQLLogger(DEBUG, true)
	logger.SetShowCaller(true)
	return logger
}
