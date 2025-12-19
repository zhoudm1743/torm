package db

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// Driver 数据库驱动类型
type Driver string

const (
	DriverMySQL     Driver = "mysql"
	DriverPostgres  Driver = "postgres"
	DriverSQLite    Driver = "sqlite"
	DriverSQLServer Driver = "sqlserver"
)

// Config 数据库配置
type Config struct {
	// 基础配置
	Driver   string `json:"driver" yaml:"driver"`     // 数据库驱动: mysql, postgres, sqlite, sqlserver
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
	SSLMode     string `json:"ssl_mode" yaml:"ssl_mode"`           // SSL模式: disable, require, verify-ca, verify-full
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
	switch c.normalizeDriver() {
	case string(DriverMySQL):
		return c.buildMySQLDSN()
	case string(DriverPostgres):
		return c.buildPostgresDSN()
	case string(DriverSQLite):
		return c.buildSQLiteDSN()
	case string(DriverSQLServer):
		return c.buildSQLServerDSN()
	default:
		return ""
	}
}

// normalizeDriver 标准化驱动名称
func (c *Config) normalizeDriver() string {
	driver := strings.ToLower(c.Driver)
	switch driver {
	case "mysql":
		return string(DriverMySQL)
	case "postgres", "postgresql", "pgsql":
		return string(DriverPostgres)
	case "sqlite", "sqlite3":
		return string(DriverSQLite)
	case "sqlserver", "mssql":
		return string(DriverSQLServer)
	default:
		return driver
	}
}

// buildMySQLDSN 构建MySQL DSN
// 格式: username:password@tcp(host:port)/database?params
func (c *Config) buildMySQLDSN() string {
	var dsn strings.Builder

	// 用户名和密码
	dsn.WriteString(c.Username)
	dsn.WriteString(":")
	dsn.WriteString(c.Password)
	dsn.WriteString("@tcp(")
	dsn.WriteString(c.Host)

	// 端口
	if c.Port > 0 {
		dsn.WriteString(fmt.Sprintf(":%d", c.Port))
	}
	dsn.WriteString(")/")
	dsn.WriteString(c.Database)

	// 构建参数
	params := make([]string, 0)

	// 字符集
	if c.Charset != "" {
		params = append(params, "charset="+c.Charset)
	}

	// 时区
	if c.Timezone != "" {
		params = append(params, "loc="+url.QueryEscape(c.Timezone))
	}

	// 解析时间
	params = append(params, "parseTime=true")

	// 添加其他参数
	for k, v := range c.Options {
		params = append(params, k+"="+v)
	}

	if len(params) > 0 {
		dsn.WriteString("?")
		dsn.WriteString(strings.Join(params, "&"))
	}

	return dsn.String()
}

// buildPostgresDSN 构建PostgreSQL DSN
// 格式: host=xxx port=xxx user=xxx password=xxx dbname=xxx sslmode=xxx
func (c *Config) buildPostgresDSN() string {
	params := make([]string, 0, 10)

	// 基础参数
	params = append(params, "host="+c.Host)
	if c.Port > 0 {
		params = append(params, fmt.Sprintf("port=%d", c.Port))
	}
	params = append(params, "user="+c.Username)
	params = append(params, "password="+c.Password)
	params = append(params, "dbname="+c.Database)

	// SSL模式
	if c.SSLMode != "" {
		params = append(params, "sslmode="+c.SSLMode)
	} else {
		params = append(params, "sslmode=disable")
	}

	// SSL证书
	if c.SSLCert != "" {
		params = append(params, "sslcert="+c.SSLCert)
	}
	if c.SSLKey != "" {
		params = append(params, "sslkey="+c.SSLKey)
	}
	if c.SSLRootCert != "" {
		params = append(params, "sslrootcert="+c.SSLRootCert)
	}

	// 时区
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
// 格式: 文件路径或:memory:
func (c *Config) buildSQLiteDSN() string {
	dsn := c.Database

	// 添加参数
	if len(c.Options) > 0 {
		params := make([]string, 0, len(c.Options))
		for k, v := range c.Options {
			params = append(params, k+"="+v)
		}
		dsn += "?" + strings.Join(params, "&")
	}

	return dsn
}

// buildSQLServerDSN 构建SQL Server DSN
// 格式: server=xxx,port;user id=xxx;password=xxx;database=xxx
func (c *Config) buildSQLServerDSN() string {
	var dsn strings.Builder

	dsn.WriteString(fmt.Sprintf("server=%s", c.Host))
	if c.Port > 0 {
		dsn.WriteString(fmt.Sprintf(",%d", c.Port))
	}
	dsn.WriteString(fmt.Sprintf(";user id=%s;password=%s;database=%s",
		c.Username, c.Password, c.Database))

	// 添加其他参数
	for k, v := range c.Options {
		dsn.WriteString(fmt.Sprintf(";%s=%s", k, v))
	}

	return dsn.String()
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Driver == "" {
		return NewError(ErrCodeConfigurationError, "driver is required")
	}

	normalizedDriver := c.normalizeDriver()

	switch normalizedDriver {
	case string(DriverMySQL), string(DriverPostgres), string(DriverSQLServer):
		if c.Host == "" {
			return NewErrorf(ErrCodeConfigurationError, "host is required for %s driver", c.Driver)
		}
		if c.Username == "" {
			return NewErrorf(ErrCodeConfigurationError, "username is required for %s driver", c.Driver)
		}
		if c.Database == "" {
			return NewErrorf(ErrCodeConfigurationError, "database is required for %s driver", c.Driver)
		}
	case string(DriverSQLite):
		if c.Database == "" {
			return NewError(ErrCodeConfigurationError, "database file path is required for sqlite driver")
		}
	default:
		return NewErrorf(ErrCodeConfigurationError, "unsupported driver: %s", c.Driver)
	}

	// 验证连接池配置
	if c.MaxOpenConns < 0 {
		return NewError(ErrCodeConfigurationError, "max_open_conns cannot be negative")
	}
	if c.MaxIdleConns < 0 {
		return NewError(ErrCodeConfigurationError, "max_idle_conns cannot be negative")
	}
	if c.MaxIdleConns > c.MaxOpenConns && c.MaxOpenConns > 0 {
		return NewError(ErrCodeConfigurationError, "max_idle_conns cannot be greater than max_open_conns")
	}
	if c.ConnMaxLifetime < 0 {
		return NewError(ErrCodeConfigurationError, "conn_max_lifetime cannot be negative")
	}
	if c.ConnMaxIdleTime < 0 {
		return NewError(ErrCodeConfigurationError, "conn_max_idle_time cannot be negative")
	}

	// 验证SSL模式
	if c.SSLMode != "" {
		validSSLModes := map[string]bool{
			"disable": true, "require": true, "verify-ca": true, "verify-full": true,
			// PostgreSQL specific
			"allow": true, "prefer": true,
		}
		if !validSSLModes[c.SSLMode] {
			return NewErrorf(ErrCodeConfigurationError, "invalid ssl_mode: %s", c.SSLMode)
		}
	}

	return nil
}

// ValidateWithContext 验证配置并返回带上下文的错误
func (c *Config) ValidateWithContext() error {
	err := c.Validate()
	if err != nil {
		if tormErr, ok := err.(*TormError); ok {
			return tormErr.
				WithContext("driver", c.Driver).
				WithContext("host", c.Host).
				WithContext("database", c.Database)
		}
	}
	return err
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Driver:          string(DriverMySQL),
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

// DefaultMySQLConfig 默认MySQL配置
func DefaultMySQLConfig() *Config {
	return &Config{
		Driver:          string(DriverMySQL),
		Host:            "localhost",
		Port:            3306,
		Charset:         "utf8mb4",
		Timezone:        "Local",
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

// DefaultPostgresConfig 默认PostgreSQL配置
func DefaultPostgresConfig() *Config {
	return &Config{
		Driver:          string(DriverPostgres),
		Host:            "localhost",
		Port:            5432,
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

// DefaultSQLiteConfig 默认SQLite配置
func DefaultSQLiteConfig() *Config {
	return &Config{
		Driver:          string(DriverSQLite),
		Database:        ":memory:",
		MaxOpenConns:    1, // SQLite建议单连接
		MaxIdleConns:    1,
		ConnMaxLifetime: 0, // 不限制
		ConnMaxIdleTime: 0,
		Options:         make(map[string]string),
		Debug:           false,
		LogQueries:      true,
	}
}

// Clone 克隆配置
func (c *Config) Clone() *Config {
	clone := *c

	// 深拷贝切片
	if c.ReadHosts != nil {
		clone.ReadHosts = make([]string, len(c.ReadHosts))
		copy(clone.ReadHosts, c.ReadHosts)
	}
	if c.WriteHosts != nil {
		clone.WriteHosts = make([]string, len(c.WriteHosts))
		copy(clone.WriteHosts, c.WriteHosts)
	}

	// 深拷贝map
	if c.Options != nil {
		clone.Options = make(map[string]string, len(c.Options))
		for k, v := range c.Options {
			clone.Options[k] = v
		}
	}

	return &clone
}

// GetDriverName 获取标准化的驱动名称
func (c *Config) GetDriverName() string {
	return c.normalizeDriver()
}

// GetDefaultPort 获取驱动的默认端口
func (c *Config) GetDefaultPort() int {
	switch c.normalizeDriver() {
	case string(DriverMySQL):
		return 3306
	case string(DriverPostgres):
		return 5432
	case string(DriverSQLServer):
		return 1433
	default:
		return 0
	}
}

// WithHost 设置主机
func (c *Config) WithHost(host string) *Config {
	c.Host = host
	return c
}

// WithPort 设置端口
func (c *Config) WithPort(port int) *Config {
	c.Port = port
	return c
}

// WithDatabase 设置数据库名
func (c *Config) WithDatabase(database string) *Config {
	c.Database = database
	return c
}

// WithCredentials 设置用户名和密码
func (c *Config) WithCredentials(username, password string) *Config {
	c.Username = username
	c.Password = password
	return c
}

// WithPoolConfig 设置连接池配置
func (c *Config) WithPoolConfig(maxOpen, maxIdle int, maxLifetime, maxIdleTime time.Duration) *Config {
	c.MaxOpenConns = maxOpen
	c.MaxIdleConns = maxIdle
	c.ConnMaxLifetime = maxLifetime
	c.ConnMaxIdleTime = maxIdleTime
	return c
}

// WithSSL 设置SSL配置
func (c *Config) WithSSL(mode, cert, key, rootCert string) *Config {
	c.SSLMode = mode
	c.SSLCert = cert
	c.SSLKey = key
	c.SSLRootCert = rootCert
	return c
}

// WithDebug 设置调试模式
func (c *Config) WithDebug(debug bool) *Config {
	c.Debug = debug
	return c
}

// WithPrefix 设置表前缀
func (c *Config) WithPrefix(prefix string) *Config {
	c.Prefix = prefix
	return c
}

// WithOption 添加选项
func (c *Config) WithOption(key, value string) *Config {
	if c.Options == nil {
		c.Options = make(map[string]string)
	}
	c.Options[key] = value
	return c
}

// WithReadWriteSplit 配置读写分离
// Requirements: 1.7
func (c *Config) WithReadWriteSplit(writeHosts []string, readHosts []string) *Config {
	c.WriteHosts = writeHosts
	c.ReadHosts = readHosts
	return c
}

// WithWriteHosts 设置写库主机列表
// Requirements: 1.7
func (c *Config) WithWriteHosts(hosts ...string) *Config {
	c.WriteHosts = hosts
	return c
}

// WithReadHosts 设置读库主机列表
// Requirements: 1.7
func (c *Config) WithReadHosts(hosts ...string) *Config {
	c.ReadHosts = hosts
	return c
}

// IsReadWriteSplitEnabled 检查是否启用了读写分离
func (c *Config) IsReadWriteSplitEnabled() bool {
	return len(c.ReadHosts) > 0 || len(c.WriteHosts) > 0
}

// GetWriteHost 获取写库主机（第一个写库或默认主机）
func (c *Config) GetWriteHost() string {
	if len(c.WriteHosts) > 0 {
		return c.WriteHosts[0]
	}
	return c.Host
}

// GetReadHosts 获取读库主机列表
func (c *Config) GetReadHosts() []string {
	if len(c.ReadHosts) > 0 {
		return c.ReadHosts
	}
	// 如果没有配置读库，返回主机作为默认读库
	if c.Host != "" {
		return []string{c.Host}
	}
	return nil
}

// MaskPassword 返回密码脱敏后的DSN（用于日志）
func (c *Config) MaskPassword() string {
	masked := c.Clone()
	if masked.Password != "" {
		masked.Password = "****"
	}
	return masked.DSN()
}

// String 返回配置的字符串表示（密码脱敏）
func (c *Config) String() string {
	return fmt.Sprintf("Config{Driver: %s, Host: %s, Port: %d, Database: %s, User: %s}",
		c.Driver, c.Host, c.Port, c.Database, c.Username)
}
