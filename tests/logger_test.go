package tests

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zhoudm1743/torm/pkg/db"
	"github.com/zhoudm1743/torm/pkg/logger"
)

func TestLogger_LogLevels(t *testing.T) {
	// 创建测试logger
	testLogger := logger.NewLogger(logger.DEBUG)

	// 测试不同日志级别
	assert.Equal(t, logger.DEBUG, testLogger.GetLevel())

	// 测试设置日志级别
	testLogger.SetLevel(logger.INFO)
	assert.Equal(t, logger.INFO, testLogger.GetLevel())

	testLogger.SetLevel(logger.WARN)
	assert.Equal(t, logger.WARN, testLogger.GetLevel())

	testLogger.SetLevel(logger.ERROR)
	assert.Equal(t, logger.ERROR, testLogger.GetLevel())
}

func TestLogger_LogLevelString(t *testing.T) {
	assert.Equal(t, "DEBUG", logger.DEBUG.String())
	assert.Equal(t, "INFO", logger.INFO.String())
	assert.Equal(t, "WARN", logger.WARN.String())
	assert.Equal(t, "ERROR", logger.ERROR.String())
	assert.Equal(t, "FATAL", logger.FATAL.String())

	// 测试未知级别
	unknownLevel := logger.LogLevel(999)
	assert.Equal(t, "UNKNOWN", unknownLevel.String())
}

func TestLogger_Creation(t *testing.T) {
	// 测试创建标准logger
	testLogger := logger.NewLogger(logger.INFO)
	assert.NotNil(t, testLogger)
	assert.Equal(t, logger.INFO, testLogger.GetLevel())

	// 测试创建文件logger
	tempFile := "test_log.log"
	defer os.Remove(tempFile) // 清理测试文件

	fileLogger, err := logger.NewFileLogger(logger.DEBUG, tempFile)
	require.NoError(t, err)
	assert.NotNil(t, fileLogger)
	assert.Equal(t, logger.DEBUG, fileLogger.GetLevel())
}

func TestLogger_FileLogging(t *testing.T) {
	tempFile := "test_file_log.log"
	defer os.Remove(tempFile)

	// 创建文件logger
	fileLogger, err := logger.NewFileLogger(logger.INFO, tempFile)
	require.NoError(t, err)

	// 写入一些日志
	fileLogger.Info("Test info message")
	fileLogger.Warn("Test warning message")
	fileLogger.Error("Test error message")

	// 读取文件内容验证
	content, err := os.ReadFile(tempFile)
	require.NoError(t, err)

	logContent := string(content)
	assert.Contains(t, logContent, "[INFO] Test info message")
	assert.Contains(t, logContent, "[WARN] Test warning message")
	assert.Contains(t, logContent, "[ERROR] Test error message")
}

func TestSQLLogger_Creation(t *testing.T) {
	sqlLogger := logger.NewSQLLogger(logger.DEBUG, true)
	assert.NotNil(t, sqlLogger)
	assert.Equal(t, logger.DEBUG, sqlLogger.GetLevel())
	assert.True(t, sqlLogger.IsQueryLoggingEnabled())

	// 测试禁用查询日志
	sqlLogger.DisableQueryLogging()
	assert.False(t, sqlLogger.IsQueryLoggingEnabled())

	// 测试启用查询日志
	sqlLogger.EnableQueryLogging()
	assert.True(t, sqlLogger.IsQueryLoggingEnabled())
}

func TestSQLLogger_QueryLogging(t *testing.T) {
	sqlLogger := logger.NewSQLLogger(logger.DEBUG, true)

	// 测试方法调用不会panic

	// 测试记录查询
	sqlLogger.LogQuery("SELECT * FROM users WHERE id = ?", []interface{}{1}, 10*time.Millisecond)

	// 测试记录查询错误
	sqlLogger.LogQueryError("SELECT * FROM users", []interface{}{},
		assert.AnError, 5*time.Millisecond)

	// 测试记录事务
	sqlLogger.LogTransaction("BEGIN", 1*time.Millisecond)
	sqlLogger.LogTransaction("COMMIT", 5*time.Millisecond)
	sqlLogger.LogTransaction("ROLLBACK", 2*time.Millisecond)

	// 测试记录连接
	config := &db.Config{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "test",
	}
	sqlLogger.LogConnection("connect", config)

	// 这些调用不应该panic
	assert.NotPanics(t, func() {
		sqlLogger.LogQuery("SELECT 1", []interface{}{}, time.Microsecond)
	})
}

func TestSQLLogger_QueryLoggingDisabled(t *testing.T) {
	// 创建禁用查询日志的SQLLogger
	sqlLogger := logger.NewSQLLogger(logger.DEBUG, false)

	// 这些调用应该不会产生日志输出，但也不应该panic
	assert.NotPanics(t, func() {
		sqlLogger.LogQuery("SELECT * FROM users", []interface{}{}, 10*time.Millisecond)
		sqlLogger.LogTransaction("BEGIN", 1*time.Millisecond)
	})

	assert.False(t, sqlLogger.IsQueryLoggingEnabled())
}

func TestDefaultLogger(t *testing.T) {
	// 测试默认logger
	defaultLogger := logger.DefaultLogger()
	assert.NotNil(t, defaultLogger)
	assert.Equal(t, logger.INFO, defaultLogger.GetLevel())

	// 测试设置默认级别
	logger.SetDefaultLevel(logger.DEBUG)
	assert.Equal(t, logger.DEBUG, defaultLogger.GetLevel())

	// 重置为INFO
	logger.SetDefaultLevel(logger.INFO)
}

func TestDefaultLoggerFunctions(t *testing.T) {
	// 测试全局日志函数不会panic
	assert.NotPanics(t, func() {
		logger.Debug("Debug message", "key", "value")
		logger.Info("Info message", "user", "test")
		logger.Warn("Warning message", "count", 42)
		logger.Error("Error message", "error", "test error")
	})
}

func TestLogger_WithFields(t *testing.T) {
	// 创建临时文件来验证日志输出
	tempFile := "test_fields_log.log"
	defer os.Remove(tempFile)

	fileLogger, err := logger.NewFileLogger(logger.DEBUG, tempFile)
	require.NoError(t, err)

	// 测试带字段的日志记录
	fileLogger.Info("User login", "user_id", 123, "ip", "192.168.1.1")
	fileLogger.Error("Database error", "table", "users", "error", "connection timeout")

	// 读取并验证日志内容
	content, err := os.ReadFile(tempFile)
	require.NoError(t, err)

	logContent := string(content)
	assert.Contains(t, logContent, "[INFO] User login")
	assert.Contains(t, logContent, "user_id 123")
	assert.Contains(t, logContent, "ip 192.168.1.1")
	assert.Contains(t, logContent, "[ERROR] Database error")
	assert.Contains(t, logContent, "table users")
}

func TestLogger_LevelFiltering(t *testing.T) {
	tempFile := "test_level_filtering.log"
	defer os.Remove(tempFile)

	// 创建WARN级别的logger
	fileLogger, err := logger.NewFileLogger(logger.WARN, tempFile)
	require.NoError(t, err)

	// 记录不同级别的日志
	fileLogger.Debug("This should not appear")
	fileLogger.Info("This should not appear either")
	fileLogger.Warn("This warning should appear")
	fileLogger.Error("This error should appear")

	// 读取日志文件
	content, err := os.ReadFile(tempFile)
	require.NoError(t, err)

	logContent := string(content)

	// 只有WARN和ERROR级别的日志应该出现
	assert.NotContains(t, logContent, "This should not appear")
	assert.Contains(t, logContent, "[WARN] This warning should appear")
	assert.Contains(t, logContent, "[ERROR] This error should appear")
}

func TestLogger_TimestampFormat(t *testing.T) {
	tempFile := "test_timestamp.log"
	defer os.Remove(tempFile)

	fileLogger, err := logger.NewFileLogger(logger.INFO, tempFile)
	require.NoError(t, err)

	fileLogger.Info("Test timestamp message")

	// 读取日志内容
	content, err := os.ReadFile(tempFile)
	require.NoError(t, err)

	logContent := string(content)

	// 验证时间戳格式 (YYYY-MM-DD HH:MM:SS)
	lines := strings.Split(strings.TrimSpace(logContent), "\n")
	assert.Greater(t, len(lines), 0)

	// 检查第一行是否包含正确格式的时间戳
	firstLine := lines[0]
	assert.Regexp(t, `^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`, firstLine)
	assert.Contains(t, firstLine, "[INFO] Test timestamp message")
}

func TestLogger_EmptyMessage(t *testing.T) {
	tempFile := "test_empty_unique.log"
	defer os.Remove(tempFile)

	// 确保文件不存在
	os.Remove(tempFile)

	fileLogger, err := logger.NewFileLogger(logger.INFO, tempFile)
	require.NoError(t, err)

	// 测试空消息
	fileLogger.Info("")
	fileLogger.Info("", "key", "value")

	// 读取日志内容
	content, err := os.ReadFile(tempFile)
	require.NoError(t, err)

	logContent := string(content)
	lines := strings.Split(strings.TrimSpace(logContent), "\n")

	// 过滤掉空行
	var nonEmptyLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}

	// 应该有两行非空日志
	assert.Equal(t, 2, len(nonEmptyLines))

	// 第一行应该只有时间戳和级别
	assert.Contains(t, nonEmptyLines[0], "[INFO]")

	// 第二行应该包含字段
	assert.Contains(t, nonEmptyLines[1], "[INFO]")
	assert.Contains(t, nonEmptyLines[1], "key value")
}
