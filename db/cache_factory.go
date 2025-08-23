package db

import (
	"fmt"
	"sync"
	"time"
)

// CacheType 缓存类型
type CacheType string

const (
	CacheTypeMemory CacheType = "memory"
	CacheTypeRedis  CacheType = "redis"
	CacheTypeCustom CacheType = "custom"
)

// CacheFactory 缓存工厂接口
type CacheFactory interface {
	CreateCache(cacheType CacheType, config interface{}) (FullCacheInterface, error)
	RegisterCacheProvider(cacheType CacheType, provider CacheProvider) error
	GetAvailableTypes() []CacheType
}

// CacheProvider 缓存提供者接口
type CacheProvider interface {
	CreateCache(config interface{}) (FullCacheInterface, error)
	ValidateConfig(config interface{}) error
	GetConfigExample() interface{}
}

// DefaultCacheFactory 默认缓存工厂实现
type DefaultCacheFactory struct {
	providers map[CacheType]CacheProvider
	mutex     sync.RWMutex
}

// NewCacheFactory 创建新的缓存工厂
func NewCacheFactory() CacheFactory {
	factory := &DefaultCacheFactory{
		providers: make(map[CacheType]CacheProvider),
	}

	// 注册默认的内存缓存提供者
	factory.RegisterCacheProvider(CacheTypeMemory, &MemoryCacheProvider{})

	return factory
}

// CreateCache 创建缓存实例
func (f *DefaultCacheFactory) CreateCache(cacheType CacheType, config interface{}) (FullCacheInterface, error) {
	f.mutex.RLock()
	provider, exists := f.providers[cacheType]
	f.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unsupported cache type: %s", cacheType)
	}

	if err := provider.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config for cache type %s: %w", cacheType, err)
	}

	return provider.CreateCache(config)
}

// RegisterCacheProvider 注册缓存提供者
func (f *DefaultCacheFactory) RegisterCacheProvider(cacheType CacheType, provider CacheProvider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.providers[cacheType] = provider
	return nil
}

// GetAvailableTypes 获取所有可用的缓存类型
func (f *DefaultCacheFactory) GetAvailableTypes() []CacheType {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	types := make([]CacheType, 0, len(f.providers))
	for cacheType := range f.providers {
		types = append(types, cacheType)
	}

	return types
}

// MemoryCacheProvider 内存缓存提供者
type MemoryCacheProvider struct{}

// CreateCache 创建内存缓存实例
func (p *MemoryCacheProvider) CreateCache(config interface{}) (FullCacheInterface, error) {
	var cacheConfig *CacheConfig

	switch c := config.(type) {
	case *CacheConfig:
		cacheConfig = c
	case CacheConfig:
		cacheConfig = &c
	case nil:
		cacheConfig = DefaultCacheConfig()
	default:
		return nil, fmt.Errorf("invalid config type for memory cache, expected *CacheConfig")
	}

	return NewHighConcurrencyMemoryCache(cacheConfig), nil
}

// ValidateConfig 验证配置
func (p *MemoryCacheProvider) ValidateConfig(config interface{}) error {
	switch c := config.(type) {
	case *CacheConfig:
		if c.ShardCount <= 0 {
			return fmt.Errorf("shard count must be positive")
		}
		if c.MaxSize <= 0 {
			return fmt.Errorf("max size must be positive")
		}
		if c.CleanupInterval <= 0 {
			return fmt.Errorf("cleanup interval must be positive")
		}
	case CacheConfig:
		if c.ShardCount <= 0 {
			return fmt.Errorf("shard count must be positive")
		}
		if c.MaxSize <= 0 {
			return fmt.Errorf("max size must be positive")
		}
		if c.CleanupInterval <= 0 {
			return fmt.Errorf("cleanup interval must be positive")
		}
	case nil:
		// nil config is acceptable, will use defaults
	default:
		return fmt.Errorf("invalid config type for memory cache, expected *CacheConfig")
	}

	return nil
}

// GetConfigExample 获取配置示例
func (p *MemoryCacheProvider) GetConfigExample() interface{} {
	return &CacheConfig{
		ShardCount:      16,
		MaxSize:         10000,
		DefaultTTL:      time.Hour,
		CleanupInterval: time.Minute,
		EvictionPolicy:  EvictionPolicyLRU,
	}
}

// CacheManager 缓存管理器
type CacheManager struct {
	caches  map[string]FullCacheInterface
	factory CacheFactory
	mutex   sync.RWMutex
}

// NewCacheManager 创建缓存管理器
func NewCacheManager() *CacheManager {
	return &CacheManager{
		caches:  make(map[string]FullCacheInterface),
		factory: NewCacheFactory(),
	}
}

// NewCacheManagerWithFactory 使用自定义工厂创建缓存管理器
func NewCacheManagerWithFactory(factory CacheFactory) *CacheManager {
	return &CacheManager{
		caches:  make(map[string]FullCacheInterface),
		factory: factory,
	}
}

// AddCache 添加缓存实例
func (m *CacheManager) AddCache(name string, cacheType CacheType, config interface{}) error {
	cache, err := m.factory.CreateCache(cacheType, config)
	if err != nil {
		return fmt.Errorf("failed to create cache %s: %w", name, err)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 关闭已存在的缓存
	if existingCache, exists := m.caches[name]; exists {
		existingCache.Close()
	}

	m.caches[name] = cache
	return nil
}

// GetCache 获取缓存实例
func (m *CacheManager) GetCache(name string) (FullCacheInterface, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	cache, exists := m.caches[name]
	if !exists {
		return nil, fmt.Errorf("cache %s not found", name)
	}

	return cache, nil
}

// GetOrCreateCache 获取或创建缓存实例
func (m *CacheManager) GetOrCreateCache(name string, cacheType CacheType, config interface{}) (FullCacheInterface, error) {
	// 首先尝试获取已存在的缓存
	m.mutex.RLock()
	cache, exists := m.caches[name]
	m.mutex.RUnlock()

	if exists {
		return cache, nil
	}

	// 不存在则创建新的缓存
	return m.createCacheWithLock(name, cacheType, config)
}

// createCacheWithLock 在锁保护下创建缓存
func (m *CacheManager) createCacheWithLock(name string, cacheType CacheType, config interface{}) (FullCacheInterface, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 双重检查，防止并发创建
	if cache, exists := m.caches[name]; exists {
		return cache, nil
	}

	cache, err := m.factory.CreateCache(cacheType, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache %s: %w", name, err)
	}

	m.caches[name] = cache
	return cache, nil
}

// RemoveCache 移除缓存实例
func (m *CacheManager) RemoveCache(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	cache, exists := m.caches[name]
	if !exists {
		return fmt.Errorf("cache %s not found", name)
	}

	cache.Close()
	delete(m.caches, name)
	return nil
}

// ListCaches 列出所有缓存实例名称
func (m *CacheManager) ListCaches() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	names := make([]string, 0, len(m.caches))
	for name := range m.caches {
		names = append(names, name)
	}

	return names
}

// RegisterCacheProvider 注册缓存提供者
func (m *CacheManager) RegisterCacheProvider(cacheType CacheType, provider CacheProvider) error {
	return m.factory.RegisterCacheProvider(cacheType, provider)
}

// GetAvailableTypes 获取所有可用的缓存类型
func (m *CacheManager) GetAvailableTypes() []CacheType {
	return m.factory.GetAvailableTypes()
}

// CloseAll 关闭所有缓存实例
func (m *CacheManager) CloseAll() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var lastError error
	for name, cache := range m.caches {
		if err := cache.Close(); err != nil {
			lastError = fmt.Errorf("failed to close cache %s: %w", name, err)
		}
	}

	m.caches = make(map[string]FullCacheInterface)
	return lastError
}

// GetStats 获取所有缓存的统计信息
func (m *CacheManager) GetStats() map[string]map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := make(map[string]map[string]interface{})
	for name, cache := range m.caches {
		stats[name] = cache.Stats()
	}

	return stats
}

// 全局缓存管理器实例
var defaultCacheManager *CacheManager

func init() {
	defaultCacheManager = NewCacheManager()

	// 添加默认的内存缓存实例
	defaultCacheManager.AddCache("default", CacheTypeMemory, DefaultCacheConfig())
}

// GetDefaultCacheManager 获取默认缓存管理器
func GetDefaultCacheManager() *CacheManager {
	return defaultCacheManager
}

// GetCache 从默认管理器获取缓存实例
func GetCache(name string) (FullCacheInterface, error) {
	return defaultCacheManager.GetCache(name)
}

// AddCache 向默认管理器添加缓存实例
func AddCache(name string, cacheType CacheType, config interface{}) error {
	return defaultCacheManager.AddCache(name, cacheType, config)
}
