package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zhoudm1743/torm/pkg/cache"
)

func TestMemoryCache_BasicOperations(t *testing.T) {
	cache := cache.NewMemoryCache()

	// 测试设置和获取
	err := cache.Set("test_key", "test_value", 5*time.Minute)
	require.NoError(t, err)

	value, err := cache.Get("test_key")
	require.NoError(t, err)
	assert.Equal(t, "test_value", value)

	// 测试检查键是否存在
	exists, err := cache.Has("test_key")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = cache.Has("non_existent_key")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestMemoryCache_ComplexData(t *testing.T) {
	cache := cache.NewMemoryCache()

	// 测试存储复杂数据结构
	complexData := map[string]interface{}{
		"id":     123,
		"name":   "测试用户",
		"tags":   []string{"tag1", "tag2", "tag3"},
		"meta":   map[string]string{"role": "admin", "department": "IT"},
		"score":  95.5,
		"active": true,
	}

	err := cache.Set("complex_key", complexData, 10*time.Minute)
	require.NoError(t, err)

	value, err := cache.Get("complex_key")
	require.NoError(t, err)
	assert.Equal(t, complexData, value)
}

func TestMemoryCache_Expiration(t *testing.T) {
	cache := cache.NewMemoryCache()

	// 测试过期
	err := cache.Set("expiring_key", "expiring_value", 100*time.Millisecond)
	require.NoError(t, err)

	// 立即获取应该成功
	value, err := cache.Get("expiring_key")
	require.NoError(t, err)
	assert.Equal(t, "expiring_value", value)

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	// 应该已经过期
	_, err = cache.Get("expiring_key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")

	// 检查是否存在也应该返回false
	exists, err := cache.Has("expiring_key")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestMemoryCache_Delete(t *testing.T) {
	cache := cache.NewMemoryCache()

	// 设置一些数据
	err := cache.Set("delete_key", "delete_value", 5*time.Minute)
	require.NoError(t, err)

	// 确认存在
	exists, err := cache.Has("delete_key")
	require.NoError(t, err)
	assert.True(t, exists)

	// 删除
	err = cache.Delete("delete_key")
	require.NoError(t, err)

	// 确认已删除
	exists, err = cache.Has("delete_key")
	require.NoError(t, err)
	assert.False(t, exists)

	// 获取应该失败
	_, err = cache.Get("delete_key")
	assert.Error(t, err)
}

func TestMemoryCache_Clear(t *testing.T) {
	cache := cache.NewMemoryCache()

	// 设置多个键
	keys := []string{"key1", "key2", "key3", "key4", "key5"}
	for i, key := range keys {
		err := cache.Set(key, fmt.Sprintf("value%d", i+1), 5*time.Minute)
		require.NoError(t, err)
	}

	// 确认所有键都存在
	for _, key := range keys {
		exists, err := cache.Has(key)
		require.NoError(t, err)
		assert.True(t, exists)
	}

	// 清空缓存
	err := cache.Clear()
	require.NoError(t, err)

	// 确认所有键都不存在
	for _, key := range keys {
		exists, err := cache.Has(key)
		require.NoError(t, err)
		assert.False(t, exists)
	}
}

func TestMemoryCache_NoExpiration(t *testing.T) {
	cache := cache.NewMemoryCache()

	// 设置永不过期的缓存（TTL为0）
	err := cache.Set("permanent_key", "permanent_value", 0)
	require.NoError(t, err)

	// 等待一段时间
	time.Sleep(100 * time.Millisecond)

	// 应该仍然存在
	value, err := cache.Get("permanent_key")
	require.NoError(t, err)
	assert.Equal(t, "permanent_value", value)

	exists, err := cache.Has("permanent_key")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestMemoryCache_OverwriteValue(t *testing.T) {
	cache := cache.NewMemoryCache()

	// 设置初始值
	err := cache.Set("overwrite_key", "initial_value", 5*time.Minute)
	require.NoError(t, err)

	value, err := cache.Get("overwrite_key")
	require.NoError(t, err)
	assert.Equal(t, "initial_value", value)

	// 覆写值
	err = cache.Set("overwrite_key", "new_value", 5*time.Minute)
	require.NoError(t, err)

	value, err = cache.Get("overwrite_key")
	require.NoError(t, err)
	assert.Equal(t, "new_value", value)
}

func TestMemoryCache_ConcurrentAccess(t *testing.T) {
	cache := cache.NewMemoryCache()

	// 测试并发访问
	done := make(chan bool, 10)

	// 启动多个goroutine进行并发读写
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			key := fmt.Sprintf("concurrent_key_%d", id)
			value := fmt.Sprintf("concurrent_value_%d", id)

			// 设置值
			err := cache.Set(key, value, 5*time.Minute)
			assert.NoError(t, err)

			// 读取值
			retrievedValue, err := cache.Get(key)
			assert.NoError(t, err)
			assert.Equal(t, value, retrievedValue)

			// 检查存在性
			exists, err := cache.Has(key)
			assert.NoError(t, err)
			assert.True(t, exists)
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestMemoryCache_EdgeCases(t *testing.T) {
	cache := cache.NewMemoryCache()

	// 测试空键
	err := cache.Set("", "empty_key_value", 5*time.Minute)
	require.NoError(t, err)

	value, err := cache.Get("")
	require.NoError(t, err)
	assert.Equal(t, "empty_key_value", value)

	// 测试nil值
	err = cache.Set("nil_key", nil, 5*time.Minute)
	require.NoError(t, err)

	value, err = cache.Get("nil_key")
	require.NoError(t, err)
	assert.Nil(t, value)

	// 测试非常长的键
	longKey := string(make([]byte, 1000))
	for i := range longKey {
		longKey = string(append([]byte(longKey[:i]), 'a'))
	}
	longKey = longKey[:1000] // 确保长度为1000

	err = cache.Set(longKey, "long_key_value", 5*time.Minute)
	require.NoError(t, err)

	value, err = cache.Get(longKey)
	require.NoError(t, err)
	assert.Equal(t, "long_key_value", value)
}

func TestMemoryCache_TypeSafety(t *testing.T) {
	cache := cache.NewMemoryCache()

	// 测试不同类型的值
	testCases := []struct {
		key   string
		value interface{}
	}{
		{"string_key", "string_value"},
		{"int_key", 42},
		{"float_key", 3.14159},
		{"bool_key", true},
		{"slice_key", []int{1, 2, 3, 4, 5}},
		{"map_key", map[string]int{"a": 1, "b": 2}},
		{"struct_key", struct{ Name string }{"test"}},
	}

	// 设置所有值
	for _, tc := range testCases {
		err := cache.Set(tc.key, tc.value, 5*time.Minute)
		require.NoError(t, err)
	}

	// 验证所有值
	for _, tc := range testCases {
		value, err := cache.Get(tc.key)
		require.NoError(t, err)
		assert.Equal(t, tc.value, value)
	}
}

func TestMemoryCache_ErrorCases(t *testing.T) {
	cache := cache.NewMemoryCache()

	// 测试获取不存在的键
	_, err := cache.Get("non_existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// 测试删除不存在的键（应该不报错）
	err = cache.Delete("non_existent")
	assert.NoError(t, err)
}

func BenchmarkMemoryCache_Set(b *testing.B) {
	cache := cache.NewMemoryCache()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		cache.Set(key, "benchmark_value", 5*time.Minute)
	}
}

func BenchmarkMemoryCache_Get(b *testing.B) {
	cache := cache.NewMemoryCache()

	// 预设一些数据
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		cache.Set(key, "benchmark_value", 5*time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_key_%d", i%1000)
		cache.Get(key)
	}
}

func BenchmarkMemoryCache_Has(b *testing.B) {
	cache := cache.NewMemoryCache()

	// 预设一些数据
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		cache.Set(key, "benchmark_value", 5*time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_key_%d", i%1000)
		cache.Has(key)
	}
}
