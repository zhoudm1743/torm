package db

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// Config 数据库配置
type Config struct {
	// 基础配置
	Driver   string `json:"driver" yaml:"driver"`     // 数据库驱动
	Host     string `json:"host" yaml:"host"`         // 主机地址
	Port     int    `json:"port" yaml:"port"`         // 端口
	Database string `json:"database" yaml:"database"` // 数据库名
	Username string `json:"username" yaml:"username"` // 用户名
	Password string `json:"password" yaml:"password"` // 密码
	Charset  string `json:"charset" yaml:"charset"`   // 字符集
	Timezone string `json:"timezone" yaml:"timezone"` // 时区

	// 连接池配置
	MaxOpenConns    int           `json:"max_open_conns" yaml:"max_open_conns"`         // 最大打开连接数
	MaxIdleConns    int           `json:"max_idle_conns" yaml:"max_idle_conns"`         // 最大空闲连接数
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" yaml:"conn_max_lifetime"`   // 连接最大生存时间
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time" yaml:"conn_max_idle_time"` // 连接最大空闲时间

	// 读写分离配置
	ReadHosts  []string `json:"read_hosts" yaml:"read_hosts"`   // 读库主机列表
	WriteHosts []string `json:"write_hosts" yaml:"write_hosts"` // 写库主机列表

	// SSL配置
	SSLMode     string `json:"ssl_mode" yaml:"ssl_mode"`           // SSL模式
	SSLCert     string `json:"ssl_cert" yaml:"ssl_cert"`           // SSL证书文件
	SSLKey      string `json:"ssl_key" yaml:"ssl_key"`             // SSL私钥文件
	SSLRootCert string `json:"ssl_root_cert" yaml:"ssl_root_cert"` // SSL根证书文件

	// 其他配置
	Prefix     string            `json:"prefix" yaml:"prefix"`           // 表前缀
	Options    map[string]string `json:"options" yaml:"options"`         // 其他选项
	Debug      bool              `json:"debug" yaml:"debug"`             // 是否开启调试
	LogQueries bool              `json:"log_queries" yaml:"log_queries"` // 是否记录查询日志
}

// DSN 构建数据源名称
func (c *Config) DSN() string {
	switch c.Driver {
	case "mysql":
		return c.buildMySQLDSN()
	case "postgres", "postgresql":
		return c.buildPostgresDSN()
	case "sqlite", "sqlite3":
		return c.buildSQLiteDSN()
	case "sqlserver", "mssql":
		return c.buildSQLServerDSN()
	default:
		return ""
	}
}

// buildMySQLDSN 构建MySQL DSN
func (c *Config) buildMySQLDSN() string {
	dsn := c.Username + ":" + c.Password + "@tcp(" + c.Host
	if c.Port > 0 {
		dsn += fmt.Sprintf(":%d", c.Port)
	}
	dsn += ")/" + c.Database

	params := make([]string, 0)
	if c.Charset != "" {
		params = append(params, "charset="+c.Charset)
	}
	if c.Timezone != "" {
		params = append(params, "loc="+url.QueryEscape(c.Timezone))
	}

	// 添加其他参数
	for k, v := range c.Options {
		params = append(params, k+"="+v)
	}

	if len(params) > 0 {
		dsn += "?" + strings.Join(params, "&")
	}

	return dsn
}

// buildPostgresDSN 构建PostgreSQL DSN
func (c *Config) buildPostgresDSN() string {
	params := make([]string, 0)
	params = append(params, "host="+c.Host)
	if c.Port > 0 {
		params = append(params, fmt.Sprintf("port=%d", c.Port))
	}
	params = append(params, "user="+c.Username)
	params = append(params, "password="+c.Password)
	params = append(params, "dbname="+c.Database)

	if c.SSLMode != "" {
		params = append(params, "sslmode="+c.SSLMode)
	}
	if c.Timezone != "" {
		params = append(params, "timezone="+c.Timezone)
	}

	// 添加其他参数
	for k, v := range c.Options {
		params = append(params, k+"="+v)
	}

	return strings.Join(params, " ")
}

// buildSQLiteDSN 构建SQLite DSN
func (c *Config) buildSQLiteDSN() string {
	return c.Database
}

// buildSQLServerDSN 构建SQL Server DSN
func (c *Config) buildSQLServerDSN() string {
	dsn := fmt.Sprintf("server=%s", c.Host)
	if c.Port > 0 {
		dsn += fmt.Sprintf(",%d", c.Port)
	}
	dsn += fmt.Sprintf(";user id=%s;password=%s;database=%s", c.Username, c.Password, c.Database)

	// 添加其他参数
	for k, v := range c.Options {
		dsn += fmt.Sprintf(";%s=%s", k, v)
	}

	return dsn
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Driver == "" {
		return fmt.Errorf("driver is required")
	}

	switch c.Driver {
	case "mysql", "postgres", "postgresql", "sqlserver", "mssql":
		if c.Host == "" {
			return fmt.Errorf("host is required for %s driver", c.Driver)
		}
		if c.Username == "" {
			return fmt.Errorf("username is required for %s driver", c.Driver)
		}
		if c.Database == "" {
			return fmt.Errorf("database is required for %s driver", c.Driver)
		}
	case "sqlite", "sqlite3":
		if c.Database == "" {
			return fmt.Errorf("database file path is required for sqlite driver")
		}
	default:
		return fmt.Errorf("unsupported driver: %s", c.Driver)
	}

	return nil
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Driver:          "mysql",
		Host:            "localhost",
		Port:            3306,
		Charset:         "utf8mb4",
		Timezone:        "UTC",
		MaxOpenConns:    100,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: time.Minute * 30,
		SSLMode:         "disable",
		Options:         make(map[string]string),
		Debug:           false,
		LogQueries:      true,
	}
}
