package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"torm/pkg/cache"
)

func TestMemoryCache_BasicOperations(t *testing.T) {
	ctx := context.Background()
	cache := cache.NewMemoryCache()

	// 测试设置和获取
	err := cache.Set(ctx, "test_key", "test_value", 5*time.Minute)
	require.NoError(t, err)

	value, err := cache.Get(ctx, "test_key")
	require.NoError(t, err)
	assert.Equal(t, "test_value", value)

	// 测试检查键是否存在
	exists, err := cache.Has(ctx, "test_key")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = cache.Has(ctx, "non_existent_key")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestMemoryCache_ComplexData(t *testing.T) {
	ctx := context.Background()
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

	err := cache.Set(ctx, "complex_data", complexData, 10*time.Minute)
	require.NoError(t, err)

	retrievedData, err := cache.Get(ctx, "complex_data")
	require.NoError(t, err)

	retrievedMap, ok := retrievedData.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, 123, retrievedMap["id"])
	assert.Equal(t, "测试用户", retrievedMap["name"])
	assert.Equal(t, 95.5, retrievedMap["score"])
	assert.Equal(t, true, retrievedMap["active"])
}

func TestMemoryCache_Expiration(t *testing.T) {
	ctx := context.Background()
	cache := cache.NewMemoryCache()

	// 设置短期缓存
	err := cache.Set(ctx, "short_lived", "will_expire", 100*time.Millisecond)
	require.NoError(t, err)

	// 立即获取应该成功
	value, err := cache.Get(ctx, "short_lived")
	require.NoError(t, err)
	assert.Equal(t, "will_expire", value)

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	// 过期后获取应该失败
	_, err = cache.Get(ctx, "short_lived")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")

	// Has方法也应该返回false
	exists, err := cache.Has(ctx, "short_lived")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestMemoryCache_NoExpiration(t *testing.T) {
	ctx := context.Background()
	cache := cache.NewMemoryCache()

	// 设置永久缓存（TTL为0）
	err := cache.Set(ctx, "permanent", "never_expires", 0)
	require.NoError(t, err)

	// 等待一段时间
	time.Sleep(100 * time.Millisecond)

	// 应该仍然能获取到
	value, err := cache.Get(ctx, "permanent")
	require.NoError(t, err)
	assert.Equal(t, "never_expires", value)
}

func TestMemoryCache_Delete(t *testing.T) {
	ctx := context.Background()
	cache := cache.NewMemoryCache()

	// 设置缓存
	err := cache.Set(ctx, "to_delete", "delete_me", 5*time.Minute)
	require.NoError(t, err)

	// 确认存在
	exists, err := cache.Has(ctx, "to_delete")
	require.NoError(t, err)
	assert.True(t, exists)

	// 删除
	err = cache.Delete(ctx, "to_delete")
	require.NoError(t, err)

	// 确认已删除
	exists, err = cache.Has(ctx, "to_delete")
	require.NoError(t, err)
	assert.False(t, exists)

	// 获取应该失败
	_, err = cache.Get(ctx, "to_delete")
	assert.Error(t, err)
}

func TestMemoryCache_Clear(t *testing.T) {
	ctx := context.Background()
	cache := cache.NewMemoryCache()

	// 设置多个缓存项
	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for key, value := range testData {
		err := cache.Set(ctx, key, value, 5*time.Minute)
		require.NoError(t, err)
	}

	// 确认都存在
	assert.Equal(t, 3, cache.Size())

	// 清空所有缓存
	err := cache.Clear(ctx)
	require.NoError(t, err)

	// 确认都被清空
	assert.Equal(t, 0, cache.Size())

	// 确认都不存在
	for key := range testData {
		exists, err := cache.Has(ctx, key)
		require.NoError(t, err)
		assert.False(t, exists)
	}
}

func TestMemoryCache_Size(t *testing.T) {
	ctx := context.Background()
	cache := cache.NewMemoryCache()

	// 初始大小应该为0
	assert.Equal(t, 0, cache.Size())

	// 添加缓存项
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := fmt.Sprintf("value_%d", i)
		err := cache.Set(ctx, key, value, 5*time.Minute)
		require.NoError(t, err)
	}

	// 大小应该为5
	assert.Equal(t, 5, cache.Size())

	// 删除一个
	err := cache.Delete(ctx, "key_0")
	require.NoError(t, err)

	// 大小应该为4
	assert.Equal(t, 4, cache.Size())
}

func TestMemoryCache_Keys(t *testing.T) {
	ctx := context.Background()
	cache := cache.NewMemoryCache()

	// 添加一些缓存项
	expectedKeys := []string{"user:1", "user:2", "config:timeout", "stats:count"}
	for _, key := range expectedKeys {
		err := cache.Set(ctx, key, fmt.Sprintf("value_%s", key), 5*time.Minute)
		require.NoError(t, err)
	}

	// 获取所有键
	keys := cache.Keys()
	assert.Equal(t, len(expectedKeys), len(keys))

	// 检查所有预期的键都存在
	for _, expectedKey := range expectedKeys {
		assert.Contains(t, keys, expectedKey)
	}
}

func TestMemoryCache_KeysWithExpiredItems(t *testing.T) {
	ctx := context.Background()
	cache := cache.NewMemoryCache()

	// 添加一个永久缓存项
	err := cache.Set(ctx, "permanent", "permanent_value", 0)
	require.NoError(t, err)

	// 添加一个短期缓存项
	err = cache.Set(ctx, "temporary", "temp_value", 50*time.Millisecond)
	require.NoError(t, err)

	// 立即获取键列表
	keys := cache.Keys()
	assert.Equal(t, 2, len(keys))
	assert.Contains(t, keys, "permanent")
	assert.Contains(t, keys, "temporary")

	// 等待短期缓存过期
	time.Sleep(100 * time.Millisecond)

	// 再次获取键列表，过期的键不应该包含在内
	keys = cache.Keys()
	assert.Equal(t, 1, len(keys))
	assert.Contains(t, keys, "permanent")
	assert.NotContains(t, keys, "temporary")
}

func TestMemoryCache_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	cache := cache.NewMemoryCache()

	// 并发写入
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(index int) {
			key := fmt.Sprintf("concurrent_%d", index)
			value := fmt.Sprintf("value_%d", index)
			err := cache.Set(ctx, key, value, 5*time.Minute)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// 等待所有写入完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证所有项都被写入
	assert.Equal(t, 10, cache.Size())

	// 并发读取
	for i := 0; i < 10; i++ {
		go func(index int) {
			key := fmt.Sprintf("concurrent_%d", index)
			expectedValue := fmt.Sprintf("value_%d", index)

			value, err := cache.Get(ctx, key)
			assert.NoError(t, err)
			assert.Equal(t, expectedValue, value)
			done <- true
		}(i)
	}

	// 等待所有读取完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestMemoryCache_UpdateValue(t *testing.T) {
	ctx := context.Background()
	cache := cache.NewMemoryCache()

	// 设置初始值
	err := cache.Set(ctx, "update_test", "initial_value", 5*time.Minute)
	require.NoError(t, err)

	// 获取初始值
	value, err := cache.Get(ctx, "update_test")
	require.NoError(t, err)
	assert.Equal(t, "initial_value", value)

	// 更新值
	err = cache.Set(ctx, "update_test", "updated_value", 5*time.Minute)
	require.NoError(t, err)

	// 获取更新后的值
	value, err = cache.Get(ctx, "update_test")
	require.NoError(t, err)
	assert.Equal(t, "updated_value", value)

	// 缓存大小应该还是1
	assert.Equal(t, 1, cache.Size())
}

func TestMemoryCache_GetNonExistentKey(t *testing.T) {
	ctx := context.Background()
	cache := cache.NewMemoryCache()

	// 获取不存在的键应该返回错误
	_, err := cache.Get(ctx, "non_existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
