package main

import (
	"fmt"
	"sync"
	"time"
)

// Item 缓存条目，包含值和过期时间
type Item[T any] struct {
	Value     T
	ExpiresAt int64 // Unix 纳秒时间戳，0 表示永不过期
}

// Cache 通用缓存结构
type Cache[K comparable, V any] struct {
	items map[K]Item[V] // 存储条目
	mu    sync.RWMutex  // 读写锁
	stop  chan struct{} // 停止后台清理的信号
}

// NewCache 创建新的缓存实例，可选择启动后台清理协程
func NewCache[K comparable, V any](cleanupInterval time.Duration) *Cache[K, V] {
	c := &Cache[K, V]{
		items: make(map[K]Item[V]),
		stop:  make(chan struct{}),
	}
	if cleanupInterval > 0 {
		go c.startCleanup(cleanupInterval)
	}
	return c
}

// Set 设置键值对，可选过期时间（传 0 表示永不过期）
func (c *Cache[K, V]) Set(key K, value V, ttl time.Duration) {
	var expiresAt int64
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl).UnixNano()
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = Item[V]{
		Value:     value,
		ExpiresAt: expiresAt,
	}
}

// Get 获取键对应的值，如果键不存在或已过期，返回零值和 false
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()
	if !ok {
		var zero V
		return zero, false
	}
	// 检查是否过期
	if item.ExpiresAt > 0 && time.Now().UnixNano() > item.ExpiresAt {
		// 惰性删除：在获取时发现过期则删除
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		var zero V
		return zero, false
	}
	return item.Value, true
}

// Delete 删除指定键
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Len 返回当前缓存中的条目数（包括已过期但尚未清理的）
func (c *Cache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Clear 清空缓存
func (c *Cache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[K]Item[V])
}

// Close 停止后台清理协程
func (c *Cache[K, V]) Close() {
	close(c.stop)
}

// startCleanup 定期清理过期键
func (c *Cache[K, V]) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stop:
			return
		}
	}
}

// cleanup 删除所有过期键
func (c *Cache[K, V]) cleanup() {
	now := time.Now().UnixNano()
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range c.items {
		if v.ExpiresAt > 0 && now > v.ExpiresAt {
			delete(c.items, k)
		}
	}
}

func main() {
	// 创建一个缓存，每隔 5 秒清理一次过期键
	cache := NewCache[string, string](5 * time.Second)

	// 设置键值，有效期 2 秒
	cache.Set("name", "Alice", 2*time.Second)

	// 立即获取
	if val, ok := cache.Get("name"); ok {
		fmt.Println("Get:", val) // 输出: Get: Alice
	}

	// 等待 3 秒后获取（过期）
	time.Sleep(3 * time.Second)
	if val, ok := cache.Get("name"); ok {
		fmt.Println("Get:", val)
	} else {
		fmt.Println("Get: key not found or expired")
	}

	// 设置永不过期的值
	cache.Set("city", "Beijing", 0)
	time.Sleep(6 * time.Second)
	if val, ok := cache.Get("city"); ok {
		fmt.Println("Get city:", val)
	}

	// 关闭缓存（停止后台清理）
	defer cache.Close()

}
