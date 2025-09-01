package cache

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/subconverter/subconverter-go/internal/infra/config"
)

// Cache defines the interface for cache operations
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Health(ctx context.Context) error
}

// MemoryCache implements in-memory cache
type MemoryCache struct {
	data map[string]cacheItem
	mutex sync.RWMutex
}

type cacheItem struct {
	value []byte
	expiry time.Time
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		data: make(map[string]cacheItem),
	}
	
	// Start cleanup goroutine
	go cache.cleanup()
	
	return cache
}

func (c *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	item, exists := c.data[key]
	if !exists || time.Now().After(item.expiry) {
		return nil, nil
	}
	
	return item.value, nil
}

func (c *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.data[key] = cacheItem{
		value: value,
		expiry: time.Now().Add(ttl),
	}
	
	return nil
}

func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	delete(c.data, key)
	return nil
}

func (c *MemoryCache) Health(ctx context.Context) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	return nil
}

func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		c.mutex.Lock()
		now := time.Now()
		for key, item := range c.data {
			if now.After(item.expiry) {
				delete(c.data, key)
			}
		}
		c.mutex.Unlock()
	}
}

// RedisCache implements Redis-based cache
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(cfg *config.RedisConfig) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + cfg.Port,
		Password: cfg.Password,
		DB:       cfg.Database,
	})
	
	return &RedisCache{client: client}
}

func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *RedisCache) Health(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}