package db

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Property 7: Cache Consistency**
// **Validates: Requirements 7.1, 7.2, 7.3, 7.4**

// TestProperty7_CacheConsistency tests that cache operations maintain consistency:
// - Cached queries return cached results within TTL
// - ClearCacheByTags invalidates tagged caches
// - Fresh results are returned after cache invalidation

func TestProperty7_CacheConsistency(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 7.1: Set then Get returns same value
	properties.Property("Set then Get returns same value", prop.ForAll(
		func(key string, value string) bool {
			if key == "" {
				return true // Skip empty keys
			}

			cache := NewMemoryCache()
			defer cache.Close()

			// Set value
			err := cache.Set(key, value, time.Minute)
			if err != nil {
				return false
			}

			// Get value
			retrieved, err := cache.Get(key)
			if err != nil {
				return false
			}

			// Should match
			return retrieved == value
		},
		genCacheKey(),
		gen.AnyString(),
	))

	// Property 7.2: Cached value expires after TTL
	properties.Property("Cached value expires after TTL", prop.ForAll(
		func(key string, value string) bool {
			if key == "" {
				return true
			}

			cache := NewMemoryCacheWithConfig(&CacheConfig{
				ShardCount:      4,
				MaxSize:         1000,
				DefaultTTL:      time.Millisecond * 50,
				CleanupInterval: time.Millisecond * 10,
				EvictionPolicy:  EvictionPolicyLRU,
			})
			defer cache.Close()

			// Set with very short TTL
			err := cache.Set(key, value, time.Millisecond*50)
			if err != nil {
				return false
			}

			// Should exist immediately
			exists, _ := cache.Has(key)
			if !exists {
				return false
			}

			// Wait for expiration
			time.Sleep(time.Millisecond * 100)

			// Should not exist after TTL
			exists, _ = cache.Has(key)
			return !exists
		},
		genCacheKey(),
		gen.AnyString(),
	))

	// Property 7.3: SetWithTags then DeleteByTags removes value
	properties.Property("SetWithTags then DeleteByTags removes value", prop.ForAll(
		func(key string, value string, tag string) bool {
			if key == "" || tag == "" {
				return true
			}

			cache := NewMemoryCache()
			defer cache.Close()

			// Set with tag
			err := cache.SetWithTags(key, value, time.Minute, []string{tag})
			if err != nil {
				return false
			}

			// Should exist
			exists, _ := cache.Has(key)
			if !exists {
				return false
			}

			// Delete by tag
			err = cache.DeleteByTags([]string{tag})
			if err != nil {
				return false
			}

			// Should not exist
			exists, _ = cache.Has(key)
			return !exists
		},
		genCacheKey(),
		gen.AnyString(),
		genCacheTag(),
	))

	// Property 7.4: Multiple keys with same tag are all deleted
	properties.Property("Multiple keys with same tag are all deleted", prop.ForAll(
		func(numKeys int, tag string) bool {
			if numKeys <= 0 || numKeys > 10 || tag == "" {
				return true
			}

			cache := NewMemoryCache()
			defer cache.Close()

			// Set multiple keys with same tag
			keys := make([]string, numKeys)
			for i := 0; i < numKeys; i++ {
				keys[i] = fmt.Sprintf("key_%d_%s", i, tag)
				err := cache.SetWithTags(keys[i], fmt.Sprintf("value_%d", i), time.Minute, []string{tag})
				if err != nil {
					return false
				}
			}

			// All should exist
			for _, key := range keys {
				exists, _ := cache.Has(key)
				if !exists {
					return false
				}
			}

			// Delete by tag
			err := cache.DeleteByTags([]string{tag})
			if err != nil {
				return false
			}

			// None should exist
			for _, key := range keys {
				exists, _ := cache.Has(key)
				if exists {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 10),
		genCacheTag(),
	))

	// Property 7.5: Clear removes all values
	properties.Property("Clear removes all values", prop.ForAll(
		func(numKeys int) bool {
			if numKeys <= 0 || numKeys > 20 {
				return true
			}

			cache := NewMemoryCache()
			defer cache.Close()

			// Set multiple keys
			for i := 0; i < numKeys; i++ {
				key := fmt.Sprintf("clear_test_key_%d", i)
				err := cache.Set(key, fmt.Sprintf("value_%d", i), time.Minute)
				if err != nil {
					return false
				}
			}

			// Should have items
			if cache.Size() == 0 {
				return false
			}

			// Clear all
			err := cache.Clear()
			if err != nil {
				return false
			}

			// Should be empty
			return cache.Size() == 0
		},
		gen.IntRange(1, 20),
	))

	// Property 7.6: Delete removes only specified key
	properties.Property("Delete removes only specified key", prop.ForAll(
		func(key1 string, key2 string, value1 string, value2 string) bool {
			if key1 == "" || key2 == "" || key1 == key2 {
				return true
			}

			cache := NewMemoryCache()
			defer cache.Close()

			// Set two keys
			cache.Set(key1, value1, time.Minute)
			cache.Set(key2, value2, time.Minute)

			// Delete first key
			err := cache.Delete(key1)
			if err != nil {
				return false
			}

			// First should not exist
			exists1, _ := cache.Has(key1)
			if exists1 {
				return false
			}

			// Second should still exist
			exists2, _ := cache.Has(key2)
			return exists2
		},
		genCacheKey(),
		genCacheKey(),
		gen.AnyString(),
		gen.AnyString(),
	))

	// Property 7.7: GetOrSet returns cached value on second call
	properties.Property("GetOrSet returns cached value on second call", prop.ForAll(
		func(key string) bool {
			if key == "" {
				return true
			}

			cache := NewMemoryCache()
			defer cache.Close()

			callCount := 0
			valueFunc := func() (interface{}, error) {
				callCount++
				return fmt.Sprintf("generated_%d", callCount), nil
			}

			// First call should generate value
			val1, err := cache.GetOrSet(key, valueFunc, time.Minute)
			if err != nil {
				return false
			}

			// Second call should return cached value
			val2, err := cache.GetOrSet(key, valueFunc, time.Minute)
			if err != nil {
				return false
			}

			// Values should be the same (from cache)
			if val1 != val2 {
				return false
			}

			// valueFunc should only be called once
			return callCount == 1
		},
		genCacheKey(),
	))

	// Property 7.8: Cache key generation is deterministic
	properties.Property("Cache key generation is deterministic", prop.ForAll(
		func(prefix string, data string) bool {
			if prefix == "" {
				return true
			}

			key1 := GenerateCacheKey(prefix, data)
			key2 := GenerateCacheKey(prefix, data)

			// Same input should produce same key
			return key1 == key2
		},
		genCacheKey(),
		gen.AnyString(),
	))

	// Property 7.9: Different data produces different cache keys
	properties.Property("Different data produces different cache keys", prop.ForAll(
		func(prefix string, data1 string, data2 string) bool {
			if prefix == "" || data1 == data2 {
				return true
			}

			key1 := GenerateCacheKey(prefix, data1)
			key2 := GenerateCacheKey(prefix, data2)

			// Different input should produce different keys
			return key1 != key2
		},
		genCacheKey(),
		gen.AnyString(),
		gen.AnyString(),
	))

	// Property 7.10: Stats reflect cache operations
	properties.Property("Stats reflect cache operations", prop.ForAll(
		func(numSets int, numGets int) bool {
			if numSets <= 0 || numSets > 10 || numGets <= 0 || numGets > 10 {
				return true
			}

			cache := NewMemoryCache()
			defer cache.Close()

			// Reset stats
			cache.ResetStats()

			// Perform sets
			for i := 0; i < numSets; i++ {
				cache.Set(fmt.Sprintf("stats_key_%d", i), fmt.Sprintf("value_%d", i), time.Minute)
			}

			// Perform gets (some hits, some misses)
			for i := 0; i < numGets; i++ {
				cache.Get(fmt.Sprintf("stats_key_%d", i%numSets))
			}

			// Get stats
			stats := cache.Stats()

			// Should have recorded items
			totalItems, ok := stats["total_items"].(int64)
			if !ok {
				return false
			}

			return totalItems == int64(numSets)
		},
		gen.IntRange(1, 10),
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t)
}

// Helper generators for cache tests

func genCacheKey() gopter.Gen {
	return gen.Identifier().Map(func(s string) string {
		if len(s) == 0 {
			return "test_cache_key"
		}
		result := "cache_" + strings.ToLower(s)
		if len(result) > 64 {
			result = result[:64]
		}
		return result
	})
}

func genCacheTag() gopter.Gen {
	return gen.Identifier().Map(func(s string) string {
		if len(s) == 0 {
			return "test_tag"
		}
		result := "tag_" + strings.ToLower(s)
		if len(result) > 32 {
			result = result[:32]
		}
		return result
	})
}
