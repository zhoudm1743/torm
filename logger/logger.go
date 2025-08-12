package logger

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/zhoudm1743/torm/db"
)

// LogLevel 日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
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
	level  LogLevel
	logger *log.Logger
}

// NewLogger 创建日志记录器
func NewLogger(level LogLevel) *Logger {
	return &Logger{
		level:  level,
		logger: log.New(os.Stdout, "", 0),
	}
}

// NewFileLogger 创建文件日志记录器
func NewFileLogger(level LogLevel, filename string) (*Logger, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Logger{
		level:  level,
		logger: log.New(file, "", 0),
	}, nil
}

// log 内部日志方法
func (l *Logger) log(level LogLevel, msg string, fields ...interface{}) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	var message string
	if len(fields) > 0 {
		message = fmt.Sprintf("%s [%s] %s %v", timestamp, level.String(), msg, fields)
	} else {
		message = fmt.Sprintf("%s [%s] %s", timestamp, level.String(), msg)
	}

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

// LogQuery 记录SQL查询
func (l *SQLLogger) LogQuery(sql string, args []interface{}, duration time.Duration) {
	if !l.logQueries {
		return
	}

	l.Info("SQL Query", map[string]interface{}{
		"sql":      sql,
		"args":     args,
		"duration": duration.String(),
	})
}

// LogQueryError 记录SQL查询错误
func (l *SQLLogger) LogQueryError(sql string, args []interface{}, err error, duration time.Duration) {
	l.Error("SQL Query Error", map[string]interface{}{
		"sql":      sql,
		"args":     args,
		"error":    err.Error(),
		"duration": duration.String(),
	})
}

// LogTransaction 记录事务操作
func (l *SQLLogger) LogTransaction(action string, duration time.Duration) {
	if !l.logQueries {
		return
	}

	l.Info("Transaction", map[string]interface{}{
		"action":   action,
		"duration": duration.String(),
	})
}

// LogConnection 记录连接操作
func (l *SQLLogger) LogConnection(action string, config *db.Config) {
	l.Info("Database Connection", map[string]interface{}{
		"action":   action,
		"driver":   config.Driver,
		"host":     config.Host,
		"port":     config.Port,
		"database": config.Database,
	})
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

// 确保 Logger 实现了 LoggerInterface 接口
var _ db.LoggerInterface = (*Logger)(nil)

// 默认日志记录器
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
