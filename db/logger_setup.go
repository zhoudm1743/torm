package db

import (
	"fmt"
	"os"
)

// SetupDefaultLogging 设置默认日志记录
// level: "debug", "info", "warn", "error"
// logQueries: 是否记录SQL查询
func SetupDefaultLogging(level string, logQueries bool) {
	logger := &ConsoleLogger{
		level:      parseLogLevel(level),
		logQueries: logQueries,
	}
	SetDefaultLogger(logger)
}

// SetupFileLogging 设置文件日志记录
func SetupFileLogging(filename, level string, logQueries bool) error {
	logger, err := NewFileLogger(parseLogLevel(level), filename, logQueries)
	if err != nil {
		return fmt.Errorf("failed to create file logger: %w", err)
	}
	SetDefaultLogger(logger)
	return nil
}

// LogLevel 日志级别
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// parseLogLevel 解析日志级别字符串
func parseLogLevel(level string) LogLevel {
	switch level {
	case "debug", "DEBUG":
		return LogLevelDebug
	case "info", "INFO":
		return LogLevelInfo
	case "warn", "WARN", "warning", "WARNING":
		return LogLevelWarn
	case "error", "ERROR":
		return LogLevelError
	case "fatal", "FATAL":
		return LogLevelFatal
	default:
		return LogLevelInfo
	}
}

// ConsoleLogger 控制台日志记录器
type ConsoleLogger struct {
	level      LogLevel
	logQueries bool
}

// Debug 调试日志
func (l *ConsoleLogger) Debug(msg string, fields ...interface{}) {
	if l.level <= LogLevelDebug {
		l.print("DEBUG", msg, fields...)
	}
}

// Info 信息日志
func (l *ConsoleLogger) Info(msg string, fields ...interface{}) {
	if l.level <= LogLevelInfo {
		l.print("INFO", msg, fields...)
	}
}

// Warn 警告日志
func (l *ConsoleLogger) Warn(msg string, fields ...interface{}) {
	if l.level <= LogLevelWarn {
		l.print("WARN", msg, fields...)
	}
}

// Error 错误日志
func (l *ConsoleLogger) Error(msg string, fields ...interface{}) {
	if l.level <= LogLevelError {
		l.print("ERROR", msg, fields...)
	}
}

// Fatal 致命错误日志
func (l *ConsoleLogger) Fatal(msg string, fields ...interface{}) {
	l.print("FATAL", msg, fields...)
	os.Exit(1)
}

// print 打印日志
func (l *ConsoleLogger) print(level, msg string, fields ...interface{}) {
	fmt.Printf("[%s] %s", level, msg)
	if len(fields) > 0 {
		fmt.Printf(" %v", fields)
	}
	fmt.Println()
}

// FileLogger 文件日志记录器
type FileLogger struct {
	*ConsoleLogger
	file *os.File
}

// NewFileLogger 创建文件日志记录器
func NewFileLogger(level LogLevel, filename string, logQueries bool) (*FileLogger, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &FileLogger{
		ConsoleLogger: &ConsoleLogger{
			level:      level,
			logQueries: logQueries,
		},
		file: file,
	}, nil
}

// print 重写打印方法，输出到文件
func (l *FileLogger) print(level, msg string, fields ...interface{}) {
	output := fmt.Sprintf("[%s] %s", level, msg)
	if len(fields) > 0 {
		output += fmt.Sprintf(" %v", fields)
	}
	output += "\n"

	l.file.WriteString(output)
	l.file.Sync() // 确保立即写入磁盘
}

// Close 关闭文件
func (l *FileLogger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// QuickEnableDebugLogging 快速启用调试日志（包括SQL查询）
func QuickEnableDebugLogging() {
	SetupDefaultLogging("debug", true)
	fmt.Println("🔍 Debug logging enabled with SQL query logging")
}

// QuickEnableInfoLogging 快速启用信息日志（包括SQL查询）
func QuickEnableInfoLogging() {
	SetupDefaultLogging("info", true)
	fmt.Println("ℹ️ Info logging enabled with SQL query logging")
}

// QuickDisableLogging 快速禁用日志
func QuickDisableLogging() {
	SetDefaultLogger(nil)
	fmt.Println("🔇 Logging disabled")
}

// GetCurrentLogLevel 获取当前日志级别（用于调试）
func GetCurrentLogLevel() string {
	logger := defaultManager.GetLogger()
	if logger == nil {
		return "DISABLED"
	}

	if consoleLogger, ok := logger.(*ConsoleLogger); ok {
		switch consoleLogger.level {
		case LogLevelDebug:
			return "DEBUG"
		case LogLevelInfo:
			return "INFO"
		case LogLevelWarn:
			return "WARN"
		case LogLevelError:
			return "ERROR"
		case LogLevelFatal:
			return "FATAL"
		}
	}

	return "UNKNOWN"
}
