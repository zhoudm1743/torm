package db

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

// BenchmarkHighConcurrencyCache_Set 基准测试并发写入
func BenchmarkHighConcurrencyCache_Set(b *testing.B) {
	cache := NewMemoryCache()
	defer cache.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key_%d", i)
			cache.Set(key, fmt.Sprintf("value_%d", i), time.Minute)
			i++
		}
	})
}

// BenchmarkHighConcurrencyCache_Get 基准测试并发读取
func BenchmarkHighConcurrencyCache_Get(b *testing.B) {
	cache := NewMemoryCache()
	defer cache.Close()

	// 预填充数据
	for i := 0; i < 10000; i++ {
		cache.Set(fmt.Sprintf("key_%d", i), fmt.Sprintf("value_%d", i), time.Minute)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key_%d", i%10000)
			cache.Get(key)
			i++
		}
	})
}

// BenchmarkHighConcurrencyCache_Mixed 基准测试混合读写
func BenchmarkHighConcurrencyCache_Mixed(b *testing.B) {
	cache := NewMemoryCache()
	defer cache.Close()

	// 预填充数据
	for i := 0; i < 1000; i++ {
		cache.Set(fmt.Sprintf("key_%d", i), fmt.Sprintf("value_%d", i), time.Minute)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key_%d", i%1000)
			if i%10 < 7 { // 70% 读操作
				cache.Get(key)
			} else { // 30% 写操作
				cache.Set(key, fmt.Sprintf("new_value_%d", i), time.Minute)
			}
			i++
		}
	})
}

// BenchmarkHighConcurrencyCache_GetMulti 基准测试批量获取
func BenchmarkHighConcurrencyCache_GetMulti(b *testing.B) {
	cache := NewMemoryCache()
	defer cache.Close()

	// 预填充数据
	for i := 0; i < 10000; i++ {
		cache.Set(fmt.Sprintf("key_%d", i), fmt.Sprintf("value_%d", i), time.Minute)
	}

	keys := make([]string, 100)
	for i := 0; i < 100; i++ {
		keys[i] = fmt.Sprintf("key_%d", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.GetMulti(keys)
	}
}

// BenchmarkOldCache_Set 原始缓存实现基准测试（用于对比）
func BenchmarkOldCache_Set(b *testing.B) {
	// 创建原始实现的缓存（单分片，模拟原始实现）
	config := &CacheConfig{
		ShardCount:      1, // 单分片模拟原始实现
		MaxSize:         100000,
		DefaultTTL:      time.Hour,
		CleanupInterval: time.Minute,
		EvictionPolicy:  EvictionPolicyLRU,
	}
	cache := NewMemoryCacheWithConfig(config)
	defer cache.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key_%d", i)
			cache.Set(key, fmt.Sprintf("value_%d", i), time.Minute)
			i++
		}
	})
}

// BenchmarkOldCache_Get 原始缓存实现基准测试（用于对比）
func BenchmarkOldCache_Get(b *testing.B) {
	config := &CacheConfig{
		ShardCount:      1,
		MaxSize:         100000,
		DefaultTTL:      time.Hour,
		CleanupInterval: time.Minute,
		EvictionPolicy:  EvictionPolicyLRU,
	}
	cache := NewMemoryCacheWithConfig(config)
	defer cache.Close()

	// 预填充数据
	for i := 0; i < 10000; i++ {
		cache.Set(fmt.Sprintf("key_%d", i), fmt.Sprintf("value_%d", i), time.Minute)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key_%d", i%10000)
			cache.Get(key)
			i++
		}
	})
}

// TestConcurrentOperations 并发操作测试
func TestConcurrentOperations(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	const numGoroutines = 100
	const numOperations = 1000

	var wg sync.WaitGroup
	
	// 并发写入
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				value := fmt.Sprintf("value_%d_%d", id, j)
				cache.Set(key, value, time.Minute)
			}
		}(i)
	}

	// 并发读取
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				cache.Get(key)
			}
		}(i)
	}

	wg.Wait()

	// 验证数据完整性
	expectedSize := numGoroutines * numOperations
	actualSize := cache.Size()
	
	if actualSize != expectedSize {
		t.Errorf("Expected cache size %d, got %d", expectedSize, actualSize)
	}

	// 验证统计信息
	stats := cache.Stats()
	t.Logf("Cache stats: %+v", stats)
	
	if stats["total_items"].(int64) != int64(expectedSize) {
		t.Errorf("Stats total_items mismatch: expected %d, got %d", expectedSize, stats["total_items"])
	}
}

// TestEvictionPolicies 淘汰策略测试
func TestEvictionPolicies(t *testing.T) {
	testCases := []struct {
		name   string
		policy EvictionPolicy
	}{
		{"LRU", EvictionPolicyLRU},
		{"LFU", EvictionPolicyLFU},
		{"TTL", EvictionPolicyTTL},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &CacheConfig{
				ShardCount:      4,
				MaxSize:         100, // 小容量以触发淘汰
				DefaultTTL:      time.Hour,
				CleanupInterval: time.Minute,
				EvictionPolicy:  tc.policy,
			}
			cache := NewMemoryCacheWithConfig(config)
			defer cache.Close()

			// 写入超过容量的数据
			for i := 0; i < 150; i++ {
				cache.Set(fmt.Sprintf("key_%d", i), fmt.Sprintf("value_%d", i), time.Minute)
				// 模拟不同的访问模式
				if tc.policy == EvictionPolicyLFU && i%10 == 0 {
					// 某些key访问更频繁
					for j := 0; j < 5; j++ {
						cache.Get(fmt.Sprintf("key_%d", i))
					}
				}
			}

			stats := cache.Stats()
			t.Logf("Eviction policy %s stats: %+v", tc.name, stats)
			
			if stats["total_evicted"].(int64) == 0 {
				t.Errorf("Expected some evictions with policy %s", tc.name)
			}
		})
	}
}

// TestShardingDistribution 分片分布测试
func TestShardingDistribution(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	// 写入大量数据
	const numKeys = 10000
	for i := 0; i < numKeys; i++ {
		cache.Set(fmt.Sprintf("key_%d", i), fmt.Sprintf("value_%d", i), time.Minute)
	}

	stats := cache.Stats()
	shardStats := stats["shard_stats"].([]map[string]interface{})
	
	// 检查分片分布是否均匀
	minItems := numKeys
	maxItems := 0
	
	for i, shardStat := range shardStats {
		items := shardStat["items"].(int)
		t.Logf("Shard %d: %d items", i, items)
		
		if items < minItems {
			minItems = items
		}
		if items > maxItems {
			maxItems = items
		}
	}
	
	// 分布应该相对均匀，最大和最小的差异不应超过平均值的50%
	avg := numKeys / len(shardStats)
	tolerance := avg / 2
	
	if maxItems-minItems > tolerance {
		t.Errorf("Shard distribution is uneven: min=%d, max=%d, avg=%d", minItems, maxItems, avg)
	}
}

// TestCacheExpiration TTL过期测试
func TestCacheExpiration(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	// 设置短TTL的项目
	cache.Set("short_ttl", "value1", 100*time.Millisecond)
	cache.Set("long_ttl", "value2", time.Minute)

	// 立即读取应该成功
	if _, err := cache.Get("short_ttl"); err != nil {
		t.Errorf("Expected to get short_ttl immediately")
	}
	
	if _, err := cache.Get("long_ttl"); err != nil {
		t.Errorf("Expected to get long_ttl immediately")
	}

	// 等待短TTL过期
	time.Sleep(150 * time.Millisecond)

	// 短TTL应该过期
	if _, err := cache.Get("short_ttl"); err == nil {
		t.Errorf("Expected short_ttl to be expired")
	}

	// 长TTL应该仍然有效
	if _, err := cache.Get("long_ttl"); err != nil {
		t.Errorf("Expected long_ttl to still be valid")
	}
}

// TestTaggedCache 标签缓存测试
func TestTaggedCache(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	// 设置带标签的缓存项
	cache.SetWithTags("user:1", "user1_data", time.Minute, []string{"user", "active"})
	cache.SetWithTags("user:2", "user2_data", time.Minute, []string{"user", "inactive"})
	cache.SetWithTags("post:1", "post1_data", time.Minute, []string{"post", "public"})
	cache.SetWithTags("post:2", "post2_data", time.Minute, []string{"post", "private"})

	// 验证缓存项存在
	if _, err := cache.Get("user:1"); err != nil {
		t.Errorf("Expected to get user:1")
	}

	// 根据标签删除
	cache.DeleteByTags([]string{"user"})

	// 用户相关的应该被删除
	if _, err := cache.Get("user:1"); err == nil {
		t.Errorf("Expected user:1 to be deleted")
	}
	if _, err := cache.Get("user:2"); err == nil {
		t.Errorf("Expected user:2 to be deleted")
	}

	// 帖子相关的应该仍然存在
	if _, err := cache.Get("post:1"); err != nil {
		t.Errorf("Expected post:1 to still exist")
	}
	if _, err := cache.Get("post:2"); err != nil {
		t.Errorf("Expected post:2 to still exist")
	}
}

// BenchmarkCacheWithDifferentShardCounts 不同分片数量的性能对比
func BenchmarkCacheWithDifferentShardCounts(b *testing.B) {
	shardCounts := []int{1, 2, 4, 8, 16, 32}
	
	for _, shardCount := range shardCounts {
		b.Run(fmt.Sprintf("Shards_%d", shardCount), func(b *testing.B) {
			config := &CacheConfig{
				ShardCount:      shardCount,
				MaxSize:         100000,
				DefaultTTL:      time.Hour,
				CleanupInterval: time.Minute,
				EvictionPolicy:  EvictionPolicyLRU,
			}
			cache := NewMemoryCacheWithConfig(config)
			defer cache.Close()

			// 预填充数据
			for i := 0; i < 1000; i++ {
				cache.Set(fmt.Sprintf("key_%d", i), fmt.Sprintf("value_%d", i), time.Minute)
			}

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					key := fmt.Sprintf("key_%d", i%1000)
					if i%4 == 0 { // 25% 写操作
						cache.Set(key, fmt.Sprintf("new_value_%d", i), time.Minute)
					} else { // 75% 读操作
						cache.Get(key)
					}
					i++
				}
			})
		})
	}
}

// MemoryUsageTest 内存使用测试
func TestMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory usage test in short mode")
	}

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	cache := NewMemoryCache()
	defer cache.Close()

	// 写入大量数据
	const numItems = 100000
	for i := 0; i < numItems; i++ {
		cache.Set(fmt.Sprintf("key_%d", i), fmt.Sprintf("value_%d_long_string_to_test_memory_usage", i), time.Hour)
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	memUsed := m2.Alloc - m1.Alloc
	memPerItem := memUsed / numItems

	t.Logf("Memory used: %d bytes for %d items (%.2f bytes per item)", memUsed, numItems, float64(memPerItem))
	
	// 验证缓存功能正常
	if cache.Size() != numItems {
		t.Errorf("Expected cache size %d, got %d", numItems, cache.Size())
	}

	// 清理缓存
	cache.Clear()
	if cache.Size() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", cache.Size())
	}
}
