package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReadWriteSplitConfig 测试读写分离配置
// Requirements: 1.7
func TestReadWriteSplitConfig(t *testing.T) {
	t.Run("WithReadWriteSplit", func(t *testing.T) {
		config := DefaultMySQLConfig().
			WithReadWriteSplit(
				[]string{"master.example.com"},
				[]string{"slave1.example.com", "slave2.example.com"},
			)

		assert.Equal(t, []string{"master.example.com"}, config.WriteHosts)
		assert.Equal(t, []string{"slave1.example.com", "slave2.example.com"}, config.ReadHosts)
		assert.True(t, config.IsReadWriteSplitEnabled())
	})

	t.Run("WithWriteHosts", func(t *testing.T) {
		config := DefaultMySQLConfig().
			WithWriteHosts("master1.example.com", "master2.example.com")

		assert.Equal(t, []string{"master1.example.com", "master2.example.com"}, config.WriteHosts)
		assert.True(t, config.IsReadWriteSplitEnabled())
	})

	t.Run("WithReadHosts", func(t *testing.T) {
		config := DefaultMySQLConfig().
			WithReadHosts("slave1.example.com", "slave2.example.com")

		assert.Equal(t, []string{"slave1.example.com", "slave2.example.com"}, config.ReadHosts)
		assert.True(t, config.IsReadWriteSplitEnabled())
	})

	t.Run("GetWriteHost", func(t *testing.T) {
		// 有WriteHosts时返回第一个
		config := DefaultMySQLConfig().
			WithWriteHosts("master.example.com")
		assert.Equal(t, "master.example.com", config.GetWriteHost())

		// 没有WriteHosts时返回Host
		config2 := DefaultMySQLConfig().
			WithHost("default.example.com")
		assert.Equal(t, "default.example.com", config2.GetWriteHost())
	})

	t.Run("GetReadHosts", func(t *testing.T) {
		// 有ReadHosts时返回配置的列表
		config := DefaultMySQLConfig().
			WithReadHosts("slave1.example.com", "slave2.example.com")
		assert.Equal(t, []string{"slave1.example.com", "slave2.example.com"}, config.GetReadHosts())

		// 没有ReadHosts时返回Host
		config2 := DefaultMySQLConfig().
			WithHost("default.example.com")
		assert.Equal(t, []string{"default.example.com"}, config2.GetReadHosts())
	})

	t.Run("IsReadWriteSplitEnabled", func(t *testing.T) {
		// 没有配置读写分离
		config := DefaultMySQLConfig()
		assert.False(t, config.IsReadWriteSplitEnabled())

		// 只配置了ReadHosts
		config2 := DefaultMySQLConfig().WithReadHosts("slave.example.com")
		assert.True(t, config2.IsReadWriteSplitEnabled())

		// 只配置了WriteHosts
		config3 := DefaultMySQLConfig().WithWriteHosts("master.example.com")
		assert.True(t, config3.IsReadWriteSplitEnabled())
	})
}

// TestRoundRobinBalancer 测试轮询负载均衡器
// Requirements: 1.8
func TestRoundRobinBalancer(t *testing.T) {
	balancer := &RoundRobinBalancer{}

	// 创建模拟连接
	configs := []*Config{
		{Host: "slave1"},
		{Host: "slave2"},
		{Host: "slave3"},
	}

	slaves := make([]Connection, len(configs))
	for i, cfg := range configs {
		conn := &mockConnection{config: cfg}
		slaves[i] = conn
	}

	// 测试轮询
	selected := make(map[string]int)
	for i := 0; i < 9; i++ {
		conn := balancer.Select(slaves)
		require.NotNil(t, conn)
		host := conn.GetConfig().Host
		selected[host]++
	}

	// 每个从库应该被选中3次
	assert.Equal(t, 3, selected["slave1"])
	assert.Equal(t, 3, selected["slave2"])
	assert.Equal(t, 3, selected["slave3"])
}

// TestRandomBalancer 测试随机负载均衡器
// Requirements: 1.8
func TestRandomBalancer(t *testing.T) {
	balancer := NewRandomBalancer()

	// 创建模拟连接
	configs := []*Config{
		{Host: "slave1"},
		{Host: "slave2"},
	}

	slaves := make([]Connection, len(configs))
	for i, cfg := range configs {
		conn := &mockConnection{config: cfg}
		slaves[i] = conn
	}

	// 测试随机选择（多次选择应该有不同结果）
	selected := make(map[string]int)
	for i := 0; i < 100; i++ {
		conn := balancer.Select(slaves)
		require.NotNil(t, conn)
		host := conn.GetConfig().Host
		selected[host]++
	}

	// 两个从库都应该被选中过
	assert.Greater(t, selected["slave1"], 0)
	assert.Greater(t, selected["slave2"], 0)
}

// TestLoadBalancerEmptySlaves 测试空从库列表
func TestLoadBalancerEmptySlaves(t *testing.T) {
	t.Run("RoundRobin", func(t *testing.T) {
		balancer := &RoundRobinBalancer{}
		conn := balancer.Select(nil)
		assert.Nil(t, conn)

		conn = balancer.Select([]Connection{})
		assert.Nil(t, conn)
	})

	t.Run("Random", func(t *testing.T) {
		balancer := NewRandomBalancer()
		conn := balancer.Select(nil)
		assert.Nil(t, conn)

		conn = balancer.Select([]Connection{})
		assert.Nil(t, conn)
	})
}

// TestReadWriteSplitConnectionCreation 测试读写分离连接创建
// Requirements: 1.7
func TestReadWriteSplitConnectionCreation(t *testing.T) {
	t.Run("NilConfig", func(t *testing.T) {
		_, err := NewReadWriteSplitConnection(nil, nil)
		assert.Error(t, err)
	})

	t.Run("NoWriteHost", func(t *testing.T) {
		config := &Config{
			Driver: "mysql",
			// 没有Host和WriteHosts
		}
		_, err := NewReadWriteSplitConnection(config, nil)
		assert.Error(t, err)
	})

	t.Run("ValidConfig", func(t *testing.T) {
		config := DefaultMySQLConfig().
			WithHost("master.example.com").
			WithReadHosts("slave1.example.com", "slave2.example.com")

		conn, err := NewReadWriteSplitConnection(config, nil)
		require.NoError(t, err)
		assert.NotNil(t, conn)
		assert.False(t, conn.IsConnected())
	})
}

// TestReadWriteSplitSetLoadBalancer 测试设置负载均衡器
// Requirements: 1.8
func TestReadWriteSplitSetLoadBalancer(t *testing.T) {
	config := DefaultMySQLConfig().
		WithHost("master.example.com")

	conn, err := NewReadWriteSplitConnection(config, nil)
	require.NoError(t, err)

	// 测试设置自定义负载均衡器
	customBalancer := NewRandomBalancer()
	conn.SetLoadBalancer(customBalancer)

	// 测试设置负载均衡类型
	conn.SetLoadBalancerType(LoadBalancerRoundRobin)
	conn.SetLoadBalancerType(LoadBalancerRandom)
}

// mockConnection 模拟连接（用于测试）
type mockConnection struct {
	config *Config
}

func (m *mockConnection) Connect() error    { return nil }
func (m *mockConnection) Close() error      { return nil }
func (m *mockConnection) Ping() error       { return nil }
func (m *mockConnection) IsConnected() bool { return true }
func (m *mockConnection) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}
func (m *mockConnection) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}
func (m *mockConnection) QueryRow(query string, args ...interface{}) *sql.Row { return nil }
func (m *mockConnection) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return nil
}
func (m *mockConnection) Exec(query string, args ...interface{}) (sql.Result, error) { return nil, nil }
func (m *mockConnection) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}
func (m *mockConnection) Begin() (TxInterface, error) { return nil, nil }
func (m *mockConnection) BeginTx(ctx context.Context, opts *sql.TxOptions) (TxInterface, error) {
	return nil, nil
}
func (m *mockConnection) RunTransaction(fn func(tx TxInterface) error) error { return nil }
func (m *mockConnection) GetConfig() *Config                                 { return m.config }
func (m *mockConnection) GetDriver() string                                  { return "mock" }
func (m *mockConnection) GetDB() *sql.DB                                     { return nil }
func (m *mockConnection) GetBuilder() Builder                                { return nil }
func (m *mockConnection) GetTableFields(table string) ([]ConnFieldInfo, error) {
	return nil, nil
}
func (m *mockConnection) GetTables() ([]string, error)     { return nil, nil }
func (m *mockConnection) GetStats() sql.DBStats            { return sql.DBStats{} }
func (m *mockConnection) SetLogger(logger LoggerInterface) {}
func (m *mockConnection) GetLogger() LoggerInterface       { return nil }
