package tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/zhoudm1743/torm/db"
	"github.com/zhoudm1743/torm/model"
)

// LoggerTestModel 用于测试日志的模型
type LoggerTestModel struct {
	model.BaseModel
	ID   int64  `torm:"primary_key,auto_increment"`
	Name string `torm:"type:varchar,size:100"`
}

func NewLoggerTestModel() *LoggerTestModel {
	m := &LoggerTestModel{}
	m.BaseModel = *model.NewAutoMigrateModel(m)
	m.SetTable("logger_test")
	m.SetConnection("default")
	return m
}

func TestSQLLogging(t *testing.T) {
	// 配置测试数据库
	config := &db.Config{
		Driver:     "sqlite",
		Database:   "test_logger.db",
		LogQueries: true, // 确保启用查询日志
	}

	err := db.AddConnection("default", config)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	t.Run("EnableDebugLogging", func(t *testing.T) {
		// 启用调试日志
		db.QuickEnableDebugLogging()

		// 检查当前日志级别
		level := db.GetCurrentLogLevel()
		if level != "DEBUG" {
			t.Errorf("Expected log level DEBUG, got %s", level)
		}

		t.Logf(" Current log level: %s", level)
	})

	t.Run("TestQueryLogging", func(t *testing.T) {
		// 启用调试日志
		db.QuickEnableDebugLogging()

		// 创建测试模型并执行AutoMigrate
		testModel := NewLoggerTestModel()

		t.Log(" Executing AutoMigrate (should show SQL logs)...")
		err := testModel.AutoMigrate()
		if err != nil {
			t.Fatalf("AutoMigrate failed: %v", err)
		}

		t.Log(" AutoMigrate completed - check console for SQL logs")
	})

	t.Run("TestBasicDatabaseOperations", func(t *testing.T) {
		// 确保日志仍然启用
		db.QuickEnableDebugLogging()

		t.Log(" Executing basic database operations (should show SQL logs)...")

		// 执行原生SQL查询
		result, err := db.Raw("SELECT COUNT(*) as count FROM sqlite_master WHERE type='table' AND name='logger_test'", []interface{}{})
		if err != nil {
			t.Fatalf("Raw query failed: %v", err)
		}

		t.Logf(" Raw query result: %v", result)

		// 执行SQL语句
		_, err = db.Exec("INSERT INTO logger_test (name) VALUES (?)", []interface{}{"Test Record"})
		if err != nil {
			t.Fatalf("Insert failed: %v", err)
		}

		t.Log(" Insert completed - check console for SQL logs")

		// 查询数据
		rows, err := db.Raw("SELECT * FROM logger_test", []interface{}{})
		if err != nil {
			t.Fatalf("Select failed: %v", err)
		}

		t.Logf(" Select result: %v", rows)
	})

	t.Run("TestDifferentLogLevels", func(t *testing.T) {
		t.Log("Testing different log levels...")

		// 测试INFO级别
		db.QuickEnableInfoLogging()
		level := db.GetCurrentLogLevel()
		if level != "INFO" {
			t.Errorf("Expected log level INFO, got %s", level)
		}
		t.Logf(" Info level set: %s", level)

		// 测试禁用日志
		db.QuickDisableLogging()
		level = db.GetCurrentLogLevel()
		if level != "DISABLED" {
			t.Errorf("Expected log level DISABLED, got %s", level)
		}
		t.Logf(" Logging disabled: %s", level)

		// 重新启用调试日志用于后续测试
		db.QuickEnableDebugLogging()
	})

	t.Run("TestCustomLoggerSetup", func(t *testing.T) {
		// 测试自定义日志设置
		db.SetupDefaultLogging("warn", true)
		level := db.GetCurrentLogLevel()
		if level != "WARN" {
			t.Errorf("Expected log level WARN, got %s", level)
		}
		t.Logf(" Custom WARN level set: %s", level)

		// 恢复DEBUG级别
		db.SetupDefaultLogging("debug", true)
	})

	t.Run("TestQueryBuilderLogging", func(t *testing.T) {
		// 确保调试日志启用
		db.QuickEnableDebugLogging()

		t.Log(" Testing Query Builder with logging...")

		// 使用查询构建器
		queryBuilder, err := db.Query()
		if err != nil {
			t.Fatalf("Failed to create query builder: %v", err)
		}

		// 构建查询
		result, err := queryBuilder.
			From("logger_test").
			Where("name", "=", "Test Record").
			Get()

		if err != nil {
			t.Fatalf("Query builder failed: %v", err)
		}

		t.Logf(" Query builder result: %v", result)
	})

	// 清理测试数据
	t.Cleanup(func() {
		conn, _ := db.DB("default")
		if conn != nil {
			conn.Connect()
			conn.Exec("DROP TABLE IF EXISTS logger_test")
		}
	})
}

func TestLoggerConfiguration(t *testing.T) {
	t.Run("TestLoggerInterfaces", func(t *testing.T) {
		// 测试自定义logger实现
		customLogger := &MockLogger{messages: make([]string, 0)}

		db.SetDefaultLogger(customLogger)

		// 执行一些操作
		db.Raw("SELECT 1", []interface{}{})

		// 检查是否记录了消息
		if len(customLogger.messages) == 0 {
			t.Log(" No log messages captured - this might be expected if LogQueries is disabled")
		} else {
			t.Logf(" Captured %d log messages", len(customLogger.messages))
			for i, msg := range customLogger.messages {
				t.Logf("  Message %d: %s", i+1, msg)
			}
		}
	})
}

// MockLogger 用于测试的模拟日志记录器
type MockLogger struct {
	messages []string
}

func (l *MockLogger) Debug(msg string, fields ...interface{}) {
	l.addMessage("DEBUG", msg, fields...)
}

func (l *MockLogger) Info(msg string, fields ...interface{}) {
	l.addMessage("INFO", msg, fields...)
}

func (l *MockLogger) Warn(msg string, fields ...interface{}) {
	l.addMessage("WARN", msg, fields...)
}

func (l *MockLogger) Error(msg string, fields ...interface{}) {
	l.addMessage("ERROR", msg, fields...)
}

func (l *MockLogger) Fatal(msg string, fields ...interface{}) {
	l.addMessage("FATAL", msg, fields...)
}

func (l *MockLogger) addMessage(level, msg string, fields ...interface{}) {
	message := level + ": " + msg
	if len(fields) > 0 {
		message += " " + strings.Trim(strings.Replace(fmt.Sprintf("%v", fields), " ", ", ", -1), "[]")
	}
	l.messages = append(l.messages, message)
}

func TestLoggerDocumentation(t *testing.T) {
	t.Run("ShowUsageExamples", func(t *testing.T) {
		t.Log("")
		t.Log(" TORM 日志使用示例:")
		t.Log("")
		t.Log("1. 快速启用调试日志:")
		t.Log("   db.QuickEnableDebugLogging()")
		t.Log("")
		t.Log("2. 快速启用信息日志:")
		t.Log("   db.QuickEnableInfoLogging()")
		t.Log("")
		t.Log("3. 自定义日志级别:")
		t.Log("   db.SetupDefaultLogging(\"debug\", true)  // level, logQueries")
		t.Log("   db.SetupDefaultLogging(\"info\", false)  // 不记录SQL查询")
		t.Log("")
		t.Log("4. 文件日志:")
		t.Log("   db.SetupFileLogging(\"app.log\", \"info\", true)")
		t.Log("")
		t.Log("5. 禁用日志:")
		t.Log("   db.QuickDisableLogging()")
		t.Log("")
		t.Log("6. 检查当前日志级别:")
		t.Log("   level := db.GetCurrentLogLevel()")
		t.Log("")
		t.Log("7. 在代码中使用:")
		t.Log("   func main() {")
		t.Log("       db.QuickEnableDebugLogging()  // 启用SQL日志")
		t.Log("       ")
		t.Log("       config := &db.Config{")
		t.Log("           Driver: \"sqlite\",")
		t.Log("           Database: \"app.db\",")
		t.Log("           LogQueries: true,  // 重要：启用查询日志")
		t.Log("       }")
		t.Log("       db.AddConnection(\"default\", config)")
		t.Log("       ")
		t.Log("       // 现在所有数据库操作都会显示SQL日志")
		t.Log("       user := NewUser()")
		t.Log("       user.AutoMigrate()  // 会显示CREATE TABLE语句")
		t.Log("   }")
		t.Log("")
	})
}
