package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota // 0 - 调试级别（最详细）
	INFO                  // 1 - 信息级别
	WARN                  // 2 - 警告级别
	ERROR                 // 3 - 错误级别
	FATAL                 // 4 - 致命错误级别（最严重）
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
	default:
		return "UNKNOWN"
	}
}

// Logger 日志记录器
type Logger struct {
	level      LogLevel
	logger     *log.Logger
	showCaller bool
}

// NewLogger 创建日志记录器
func NewLogger(level LogLevel) *Logger {
	return &Logger{
		level:      level,
		logger:     log.New(os.Stdout, "", 0),
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
		logger:     log.New(file, "", 0),
		showCaller: false,
	}, nil
}

// SetShowCaller 设置是否显示调用者信息
func (l *Logger) SetShowCaller(show bool) {
	l.showCaller = show
}

// shouldLog 检查是否应该输出日志
func (l *Logger) shouldLog(level LogLevel) bool {
	return level >= l.level
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
	l.level = level
}

// GetLevel 获取日志级别
func (l *Logger) GetLevel() LogLevel {
	return l.level
}

// SQLLogger SQL查询日志记录器
type SQLLogger struct {
	*Logger
	logQueries bool
}

// NewSQLLogger 创建SQL日志记录器
func NewSQLLogger(level LogLevel, logQueries bool) *SQLLogger {
	return &SQLLogger{
		Logger:     NewLogger(level),
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

// LogSQL 统一记录SQL执行日志（推荐使用）
func (l *SQLLogger) LogSQL(sql string, args []interface{}, duration time.Duration, err error) {
	if err != nil {
		// 错误格式：[ERROR] [耗时] SQL语句 | 错误信息
		finalSQL := formatSQLWithArgs(sql, args)
		errorMsg := fmt.Sprintf("[%v] %s | %v", duration, finalSQL, err)
		l.Error(errorMsg)
	} else if l.logQueries {
		// 成功格式：[INFO] [耗时] [rows:affected] SQL语句 (参数已替换)
		finalSQL := formatSQLWithArgs(sql, args)
		// 注意：这里我们无法获取实际的行数，所以先显示为未知
		successMsg := fmt.Sprintf("[%v] %s", duration, finalSQL)
		l.Info(successMsg)
	}
}

// LogSQLWithRows 记录SQL执行日志（包含行数信息）
func (l *SQLLogger) LogSQLWithRows(sql string, args []interface{}, duration time.Duration, rowsAffected int64, err error) {
	if err != nil {
		// 错误格式：[ERROR] [耗时] SQL语句 | 错误信息
		finalSQL := formatSQLWithArgs(sql, args)
		errorMsg := fmt.Sprintf("[%v] %s | %v", duration, finalSQL, err)
		l.Error(errorMsg)
	} else if l.logQueries {
		// 成功格式：[INFO] [耗时] [rows:affected] SQL语句 (参数已替换)
		finalSQL := formatSQLWithArgs(sql, args)
		successMsg := fmt.Sprintf("[%v] [rows:%d] %s", duration, rowsAffected, finalSQL)
		l.Info(successMsg)
	}
}

// formatSQLWithArgs 将参数替换到SQL中
func formatSQLWithArgs(sql string, args []interface{}) string {
	if len(args) == 0 {
		return sql
	}

	// 简单的参数替换：将 ? 替换为实际参数值
	result := sql
	for _, arg := range args {
		// 找到第一个 ? 并替换
		if strings.Contains(result, "?") {
			var replacement string
			switch v := arg.(type) {
			case string:
				replacement = fmt.Sprintf("'%s'", v)
			case nil:
				replacement = "NULL"
			default:
				replacement = fmt.Sprintf("%v", v)
			}
			result = strings.Replace(result, "?", replacement, 1)
		}
	}
	return result
}

// LogQuery 记录SQL查询（兼容性保留）
func (l *SQLLogger) LogQuery(sql string, args []interface{}, duration time.Duration) {
	l.LogSQL(sql, args, duration, nil)
}

// LogQueryError 记录SQL查询错误（兼容性保留）
func (l *SQLLogger) LogQueryError(sql string, args []interface{}, err error, duration time.Duration) {
	l.LogSQL(sql, args, duration, err)
}

// LogTransaction 记录事务操作
func (l *SQLLogger) LogTransaction(action string, duration time.Duration) {
	if !l.logQueries {
		return
	}

	l.Debug("事务操作", "action", action, "duration", duration.String())
}

// LogConnection 记录连接操作
func (l *SQLLogger) LogConnection(action string, driver string, database string, duration time.Duration) {
	l.Info("数据库连接", "action", action, "driver", driver, "database", database, "duration", duration.String())
}

// EnableQueryLogging 启用查询日志
func (l *SQLLogger) EnableQueryLogging() {
	l.logQueries = true
}

// DisableQueryLogging 禁用查询日志
func (l *SQLLogger) DisableQueryLogging() {
	l.logQueries = false
}

// IsQueryLoggingEnabled 检查查询日志是否启用
func (l *SQLLogger) IsQueryLoggingEnabled() bool {
	return l.logQueries
}

// 默认日志记录器实例
var defaultLogger = NewLogger(INFO)

// DefaultLogger 获取默认日志记录器
func DefaultLogger() *Logger {
	return defaultLogger
}

// SetDefaultLevel 设置默认日志级别
func SetDefaultLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

// 便捷函数，使用默认日志记录器

// Debug 调试日志
func Debug(msg string, fields ...interface{}) {
	defaultLogger.Debug(msg, fields...)
}

// Info 信息日志
func Info(msg string, fields ...interface{}) {
	defaultLogger.Info(msg, fields...)
}

// Warn 警告日志
func Warn(msg string, fields ...interface{}) {
	defaultLogger.Warn(msg, fields...)
}

// Error 错误日志
func Error(msg string, fields ...interface{}) {
	defaultLogger.Error(msg, fields...)
}

// Fatal 致命错误日志
func Fatal(msg string, fields ...interface{}) {
	defaultLogger.Fatal(msg, fields...)
}

// 预设的日志记录器

// NewDebugLogger 创建DEBUG级别的日志记录器（显示所有日志）
func NewDebugLogger() *SQLLogger {
	return NewSQLLogger(DEBUG, true)
}

// NewInfoLogger 创建INFO级别的日志记录器
func NewInfoLogger() *SQLLogger {
	return NewSQLLogger(INFO, true)
}

// NewQuietLogger 创建WARN级别的日志记录器（较少输出）
func NewQuietLogger() *SQLLogger {
	return NewSQLLogger(WARN, false)
}

// NewSilentLogger 创建ERROR级别的日志记录器（最少输出）
func NewSilentLogger() *SQLLogger {
	return NewSQLLogger(ERROR, false)
}
