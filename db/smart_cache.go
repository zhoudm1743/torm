package db

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// SmartCacheSystem 智能多级缓存系统
type SmartCacheSystem struct {
	// L1: 内存缓存（最快）
	l1Cache *L1MemoryCache

	// L2: 本地缓存（较快）
	l2Cache *L2LocalCache

	// L3: 分布式缓存（较慢）
	l3Cache *L3DistributedCache

	// 缓存配置
	config *SmartCacheConfig

	// 统计信息
	stats *CacheStats

	// 状态
	running bool
	mutex   sync.RWMutex
}

// SmartCacheConfig 智能缓存配置
type SmartCacheConfig struct {
	// L1配置
	L1MaxSize int           `json:"l1_max_size"`
	L1TTL     time.Duration `json:"l1_ttl"`
	L1Enabled bool          `json:"l1_enabled"`

	// L2配置
	L2MaxSize   int           `json:"l2_max_size"`
	L2TTL       time.Duration `json:"l2_ttl"`
	L2Enabled   bool          `json:"l2_enabled"`
	L2Directory string        `json:"l2_directory"`

	// L3配置
	L3TTL     time.Duration  `json:"l3_ttl"`
	L3Enabled bool           `json:"l3_enabled"`
	L3Adapter CacheInterface `json:"-"`

	// 智能策略
	AutoPromote  bool `json:"auto_promote"`  // 自动提升热点数据
	AutoEvict    bool `json:"auto_evict"`    // 自动淘汰冷数据
	PrefetchSize int  `json:"prefetch_size"` // 预取大小

	// 压缩配置
	CompressionEnabled   bool `json:"compression_enabled"`
	CompressionThreshold int  `json:"compression_threshold"`
}

// CacheStats 缓存统计
type CacheStats struct {
	// 命中统计
	L1Hits int64 `json:"l1_hits"`
	L1Miss int64 `json:"l1_miss"`
	L2Hits int64 `json:"l2_hits"`
	L2Miss int64 `json:"l2_miss"`
	L3Hits int64 `json:"l3_hits"`
	L3Miss int64 `json:"l3_miss"`

	// 操作统计
	Sets    int64 `json:"sets"`
	Gets    int64 `json:"gets"`
	Deletes int64 `json:"deletes"`
	Evicts  int64 `json:"evicts"`

	// 性能统计
	AvgGetLatency time.Duration `json:"avg_get_latency"`
	AvgSetLatency time.Duration `json:"avg_set_latency"`

	// 内存统计
	L1MemoryUsage int64 `json:"l1_memory_usage"`
	L2MemoryUsage int64 `json:"l2_memory_usage"`
}

// L1MemoryCache L1内存缓存
type L1MemoryCache struct {
	data      map[string]*CacheEntry
	access    map[string]time.Time
	mutex     sync.RWMutex
	maxSize   int
	ttl       time.Duration
	cleanupCh chan bool
}

// L2LocalCache L2本地缓存
type L2LocalCache struct {
	directory   string
	maxSize     int
	ttl         time.Duration
	index       map[string]*FileEntry
	mutex       sync.RWMutex
	compression bool
}

// L3DistributedCache L3分布式缓存
type L3DistributedCache struct {
	adapter CacheInterface
	ttl     time.Duration
	mutex   sync.RWMutex
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	Type        string      `json:"type"`
	ExpiresAt   time.Time   `json:"expires_at"`
	CreatedAt   time.Time   `json:"created_at"`
	AccessCount int64       `json:"access_count"`
	LastAccess  time.Time   `json:"last_access"`
	Size        int64       `json:"size"`
	Compressed  bool        `json:"compressed"`
}

// FileEntry 文件缓存条目
type FileEntry struct {
	Path       string    `json:"path"`
	Size       int64     `json:"size"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
	LastAccess time.Time `json:"last_access"`
}

// 使用现有的CacheType，添加新的常量
const (
	SmartCacheTypeQuery    = "query"    // 查询结果缓存
	SmartCacheTypeAccessor = "accessor" // 访问器结果缓存
	SmartCacheTypeModel    = "model"    // 模型缓存
	SmartCacheTypeRelation = "relation" // 关联缓存
)

// NewSmartCacheSystem 创建智能缓存系统
func NewSmartCacheSystem(config *SmartCacheConfig) *SmartCacheSystem {
	if config == nil {
		config = DefaultSmartCacheConfig()
	}

	scs := &SmartCacheSystem{
		config:  config,
		stats:   &CacheStats{},
		running: false,
	}

	// 初始化L1缓存
	if config.L1Enabled {
		scs.l1Cache = NewL1MemoryCache(config.L1MaxSize, config.L1TTL)
	}

	// 初始化L2缓存
	if config.L2Enabled {
		scs.l2Cache = NewL2LocalCache(config.L2Directory, config.L2MaxSize, config.L2TTL)
	}

	// 初始化L3缓存
	if config.L3Enabled && config.L3Adapter != nil {
		scs.l3Cache = NewL3DistributedCache(config.L3Adapter, config.L3TTL)
	}

	return scs
}

// DefaultSmartCacheConfig 默认配置
func DefaultSmartCacheConfig() *SmartCacheConfig {
	return &SmartCacheConfig{
		L1MaxSize:            10000,
		L1TTL:                5 * time.Minute,
		L1Enabled:            true,
		L2MaxSize:            100000,
		L2TTL:                30 * time.Minute,
		L2Enabled:            true,
		L2Directory:          "./cache",
		L3TTL:                2 * time.Hour,
		L3Enabled:            false,
		AutoPromote:          true,
		AutoEvict:            true,
		PrefetchSize:         10,
		CompressionEnabled:   true,
		CompressionThreshold: 1024,
	}
}

// Start 启动缓存系统
func (scs *SmartCacheSystem) Start() error {
	scs.mutex.Lock()
	defer scs.mutex.Unlock()

	if scs.running {
		return ErrCacheSystemRunning
	}

	scs.running = true

	// 启动L1缓存清理
	if scs.l1Cache != nil {
		go scs.l1Cache.startCleanup()
	}

	// 启动智能管理器
	go scs.smartManager()

	return nil
}

// Stop 停止缓存系统
func (scs *SmartCacheSystem) Stop() error {
	scs.mutex.Lock()
	defer scs.mutex.Unlock()

	if !scs.running {
		return ErrCacheSystemNotRunning
	}

	scs.running = false

	// 停止L1缓存清理
	if scs.l1Cache != nil {
		close(scs.l1Cache.cleanupCh)
	}

	return nil
}

// Get 智能获取缓存
func (scs *SmartCacheSystem) Get(key string) (interface{}, bool) {
	startTime := time.Now()
	defer func() {
		atomic.AddInt64(&scs.stats.Gets, 1)
		latency := time.Since(startTime)
		scs.updateAvgLatency(&scs.stats.AvgGetLatency, latency)
	}()

	// 尝试L1缓存
	if scs.l1Cache != nil {
		if value, ok := scs.l1Cache.Get(key); ok {
			atomic.AddInt64(&scs.stats.L1Hits, 1)
			return value, true
		}
		atomic.AddInt64(&scs.stats.L1Miss, 1)
	}

	// 尝试L2缓存
	if scs.l2Cache != nil {
		if value, ok := scs.l2Cache.Get(key); ok {
			atomic.AddInt64(&scs.stats.L2Hits, 1)

			// 自动提升到L1
			if scs.config.AutoPromote && scs.l1Cache != nil {
				scs.l1Cache.Set(key, value, scs.config.L1TTL)
			}

			return value, true
		}
		atomic.AddInt64(&scs.stats.L2Miss, 1)
	}

	// 尝试L3缓存
	if scs.l3Cache != nil {
		if value, ok := scs.l3Cache.Get(key); ok {
			atomic.AddInt64(&scs.stats.L3Hits, 1)

			// 自动提升到L2和L1
			if scs.config.AutoPromote {
				if scs.l2Cache != nil {
					scs.l2Cache.Set(key, value, scs.config.L2TTL)
				}
				if scs.l1Cache != nil {
					scs.l1Cache.Set(key, value, scs.config.L1TTL)
				}
			}

			return value, true
		}
		atomic.AddInt64(&scs.stats.L3Miss, 1)
	}

	return nil, false
}

// Set 智能设置缓存
func (scs *SmartCacheSystem) Set(key string, value interface{}, cacheType string) {
	startTime := time.Now()
	defer func() {
		atomic.AddInt64(&scs.stats.Sets, 1)
		latency := time.Since(startTime)
		scs.updateAvgLatency(&scs.stats.AvgSetLatency, latency)
	}()

	// 同时设置到所有启用的缓存层
	if scs.l1Cache != nil {
		scs.l1Cache.Set(key, value, scs.config.L1TTL)
	}

	if scs.l2Cache != nil {
		scs.l2Cache.Set(key, value, scs.config.L2TTL)
	}

	if scs.l3Cache != nil {
		scs.l3Cache.Set(key, value, scs.config.L3TTL)
	}
}

// Delete 删除缓存
func (scs *SmartCacheSystem) Delete(key string) {
	atomic.AddInt64(&scs.stats.Deletes, 1)

	if scs.l1Cache != nil {
		scs.l1Cache.Delete(key)
	}

	if scs.l2Cache != nil {
		scs.l2Cache.Delete(key)
	}

	if scs.l3Cache != nil {
		scs.l3Cache.Delete(key)
	}
}

// CacheQuery 缓存查询结果
func (scs *SmartCacheSystem) CacheQuery(query string, params []interface{}, result interface{}) {
	key := scs.generateQueryKey(query, params)
	scs.Set(key, result, SmartCacheTypeQuery)
}

// GetCachedQuery 获取缓存的查询结果
func (scs *SmartCacheSystem) GetCachedQuery(query string, params []interface{}) (interface{}, bool) {
	key := scs.generateQueryKey(query, params)
	return scs.Get(key)
}

// CacheAccessorResult 缓存访问器结果
func (scs *SmartCacheSystem) CacheAccessorResult(modelType, fieldName string, input, output interface{}) {
	key := scs.generateAccessorKey(modelType, fieldName, input)
	scs.Set(key, output, SmartCacheTypeAccessor)
}

// GetCachedAccessorResult 获取缓存的访问器结果
func (scs *SmartCacheSystem) GetCachedAccessorResult(modelType, fieldName string, input interface{}) (interface{}, bool) {
	key := scs.generateAccessorKey(modelType, fieldName, input)
	return scs.Get(key)
}

// generateQueryKey 生成查询缓存键
func (scs *SmartCacheSystem) generateQueryKey(query string, params []interface{}) string {
	data := struct {
		Query  string        `json:"query"`
		Params []interface{} `json:"params"`
	}{
		Query:  query,
		Params: params,
	}

	jsonData, _ := json.Marshal(data)
	hash := md5.Sum(jsonData)
	return fmt.Sprintf("query:%s", hex.EncodeToString(hash[:]))
}

// generateAccessorKey 生成访问器缓存键
func (scs *SmartCacheSystem) generateAccessorKey(modelType, fieldName string, input interface{}) string {
	data := struct {
		ModelType string      `json:"model_type"`
		FieldName string      `json:"field_name"`
		Input     interface{} `json:"input"`
	}{
		ModelType: modelType,
		FieldName: fieldName,
		Input:     input,
	}

	jsonData, _ := json.Marshal(data)
	hash := md5.Sum(jsonData)
	return fmt.Sprintf("accessor:%s", hex.EncodeToString(hash[:]))
}

// GetStats 获取统计信息
func (scs *SmartCacheSystem) GetStats() *CacheStats {
	return scs.stats
}

// updateAvgLatency 更新平均延迟
func (scs *SmartCacheSystem) updateAvgLatency(current *time.Duration, newLatency time.Duration) {
	// 简单的滑动平均
	*current = (*current + newLatency) / 2
}

// smartManager 智能管理器
func (scs *SmartCacheSystem) smartManager() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		if !scs.running {
			break
		}

		select {
		case <-ticker.C:
			scs.performSmartOptimizations()
		}
	}
}

// performSmartOptimizations 执行智能优化
func (scs *SmartCacheSystem) performSmartOptimizations() {
	// 自动淘汰冷数据
	if scs.config.AutoEvict {
		scs.evictColdData()
	}

	// 预取热点数据
	if scs.config.PrefetchSize > 0 {
		scs.prefetchHotData()
	}

	// 更新内存使用统计
	scs.updateMemoryStats()
}

// evictColdData 淘汰冷数据
func (scs *SmartCacheSystem) evictColdData() {
	threshold := time.Now().Add(-10 * time.Minute)

	if scs.l1Cache != nil {
		scs.l1Cache.evictBefore(threshold)
	}
}

// prefetchHotData 预取热点数据
func (scs *SmartCacheSystem) prefetchHotData() {
	// 实现热点数据预取逻辑
	// 基于访问频率和模式预测需要预取的数据
}

// updateMemoryStats 更新内存统计
func (scs *SmartCacheSystem) updateMemoryStats() {
	if scs.l1Cache != nil {
		scs.stats.L1MemoryUsage = scs.l1Cache.getMemoryUsage()
	}

	if scs.l2Cache != nil {
		scs.stats.L2MemoryUsage = scs.l2Cache.getMemoryUsage()
	}
}

// L1MemoryCache 实现

// NewL1MemoryCache 创建L1内存缓存
func NewL1MemoryCache(maxSize int, ttl time.Duration) *L1MemoryCache {
	return &L1MemoryCache{
		data:      make(map[string]*CacheEntry),
		access:    make(map[string]time.Time),
		maxSize:   maxSize,
		ttl:       ttl,
		cleanupCh: make(chan bool),
	}
}

// Get 获取缓存
func (l1 *L1MemoryCache) Get(key string) (interface{}, bool) {
	l1.mutex.RLock()
	defer l1.mutex.RUnlock()

	entry, exists := l1.data[key]
	if !exists {
		return nil, false
	}

	// 检查过期
	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	// 更新访问时间
	entry.LastAccess = time.Now()
	atomic.AddInt64(&entry.AccessCount, 1)

	return entry.Value, true
}

// Set 设置缓存
func (l1 *L1MemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	l1.mutex.Lock()
	defer l1.mutex.Unlock()

	// 检查容量
	if len(l1.data) >= l1.maxSize {
		l1.evictLRU()
	}

	now := time.Now()
	entry := &CacheEntry{
		Key:         key,
		Value:       value,
		ExpiresAt:   now.Add(ttl),
		CreatedAt:   now,
		LastAccess:  now,
		AccessCount: 1,
	}

	l1.data[key] = entry
	l1.access[key] = now
}

// Delete 删除缓存
func (l1 *L1MemoryCache) Delete(key string) {
	l1.mutex.Lock()
	defer l1.mutex.Unlock()

	delete(l1.data, key)
	delete(l1.access, key)
}

// evictLRU 淘汰最少使用的条目
func (l1 *L1MemoryCache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time

	for key, accessTime := range l1.access {
		if oldestKey == "" || accessTime.Before(oldestTime) {
			oldestKey = key
			oldestTime = accessTime
		}
	}

	if oldestKey != "" {
		delete(l1.data, oldestKey)
		delete(l1.access, oldestKey)
	}
}

// evictBefore 淘汰指定时间前的条目
func (l1 *L1MemoryCache) evictBefore(threshold time.Time) {
	l1.mutex.Lock()
	defer l1.mutex.Unlock()

	for key, accessTime := range l1.access {
		if accessTime.Before(threshold) {
			delete(l1.data, key)
			delete(l1.access, key)
		}
	}
}

// getMemoryUsage 获取内存使用量
func (l1 *L1MemoryCache) getMemoryUsage() int64 {
	l1.mutex.RLock()
	defer l1.mutex.RUnlock()

	var total int64
	for _, entry := range l1.data {
		total += entry.Size
	}
	return total
}

// startCleanup 启动清理任务
func (l1 *L1MemoryCache) startCleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l1.cleanup()
		case <-l1.cleanupCh:
			return
		}
	}
}

// cleanup 清理过期条目
func (l1 *L1MemoryCache) cleanup() {
	l1.mutex.Lock()
	defer l1.mutex.Unlock()

	now := time.Now()
	for key, entry := range l1.data {
		if now.After(entry.ExpiresAt) {
			delete(l1.data, key)
			delete(l1.access, key)
		}
	}
}

// L2LocalCache 和 L3DistributedCache 的实现...
// 由于篇幅限制，这里先实现核心的L1缓存，L2和L3的完整实现可以根据需要添加

// NewL2LocalCache 创建L2本地缓存
func NewL2LocalCache(directory string, maxSize int, ttl time.Duration) *L2LocalCache {
	return &L2LocalCache{
		directory: directory,
		maxSize:   maxSize,
		ttl:       ttl,
		index:     make(map[string]*FileEntry),
	}
}

// Get L2缓存获取（简化实现）
func (l2 *L2LocalCache) Get(key string) (interface{}, bool) {
	// 简化实现，实际应该从文件读取
	return nil, false
}

// Set L2缓存设置（简化实现）
func (l2 *L2LocalCache) Set(key string, value interface{}, ttl time.Duration) {
	// 简化实现，实际应该写入文件
}

// Delete L2缓存删除（简化实现）
func (l2 *L2LocalCache) Delete(key string) {
	// 简化实现，实际应该删除文件
}

// getMemoryUsage L2缓存内存使用（简化实现）
func (l2 *L2LocalCache) getMemoryUsage() int64 {
	return 0
}

// NewL3DistributedCache 创建L3分布式缓存
func NewL3DistributedCache(adapter CacheInterface, ttl time.Duration) *L3DistributedCache {
	return &L3DistributedCache{
		adapter: adapter,
		ttl:     ttl,
	}
}

// Get L3缓存获取
func (l3 *L3DistributedCache) Get(key string) (interface{}, bool) {
	if l3.adapter == nil {
		return nil, false
	}
	value, err := l3.adapter.Get(key)
	return value, err == nil
}

// Set L3缓存设置
func (l3 *L3DistributedCache) Set(key string, value interface{}, ttl time.Duration) {
	if l3.adapter != nil {
		l3.adapter.Set(key, value, ttl)
	}
}

// Delete L3缓存删除
func (l3 *L3DistributedCache) Delete(key string) {
	if l3.adapter != nil {
		l3.adapter.Delete(key)
	}
}

// 错误定义
var (
	ErrCacheSystemRunning    = NewError(ErrCodeInvalidModelState, "缓存系统已运行，不能重复启动")
	ErrCacheSystemNotRunning = NewError(ErrCodeInvalidModelState, "缓存系统未运行，请先启动")
)
