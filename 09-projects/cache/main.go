package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Item 缓存条目，存储值和过期时间
type Item struct {
	Value      any
	ExpiresAt  time.Time
	Expiration time.Duration // 原始的 TTL 值，用于日志
}

// IsExpired 判断条目是否已过期
func (i Item) IsExpired() bool {
	return !i.ExpiresAt.IsZero() && time.Now().After(i.ExpiresAt)
}

// Cache 泛型化的内存缓存（内部用 any，外部用类型断言）
type Cache struct {
	items map[string]*Item
	mu    sync.RWMutex      // 读写锁保证并发安全
	stats CacheStats        // 缓存统计
	stop  chan struct{}     // 停止清理协程的信号
}

// CacheStats 缓存命中/未命中统计
type CacheStats struct {
	Hits   int64
	Misses int64
}

// NewCache 创建新缓存，并启动定期清理协程
func NewCache(cleanupInterval time.Duration) *Cache {
	c := &Cache{
		items: make(map[string]*Item),
		stop:  make(chan struct{}),
	}
	// 启动后台协程定期清理过期条目
	if cleanupInterval > 0 {
		go c.cleanupLoop(cleanupInterval)
	}
	return c
}

// cleanupLoop 定期清理过期条目
func (c *Cache) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.deleteExpired()
		case <-c.stop:
			return
		}
	}
}

// deleteExpired 删除所有过期条目
func (c *Cache) deleteExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for k, v := range c.items {
		if !v.ExpiresAt.IsZero() && now.After(v.ExpiresAt) {
			delete(c.items, k)
		}
	}
}

// Set 设置缓存条目，支持 TTL（0 表示永不过期）
func (c *Cache) Set(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	c.items[key] = &Item{
		Value:      value,
		ExpiresAt:  expiresAt,
		Expiration: ttl,
	}
}

// Get 获取缓存条目。ok=false 表示 key 不存在或已过期
func (c *Cache) Get(key string) (value any, ok bool) {
	c.mu.RLock()
	item, found := c.items[key]
	c.mu.RUnlock()

	if !found {
		c.mu.Lock()
		c.stats.Misses++
		c.mu.Unlock()
		return nil, false
	}

	if item.IsExpired() {
		// 惰性删除：发现过期时立即删除
		c.mu.Lock()
		delete(c.items, key)
		c.stats.Misses++
		c.mu.Unlock()
		return nil, false
	}

	c.mu.Lock()
	c.stats.Hits++
	c.mu.Unlock()
	return item.Value, true
}

// GetString 便捷方法：获取 string 类型值
func (c *Cache) GetString(key string) (string, bool) {
	v, ok := c.Get(key)
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// GetInt 便捷方法：获取 int 类型值
func (c *Cache) GetInt(key string) (int, bool) {
	v, ok := c.Get(key)
	if !ok {
		return 0, false
	}
	i, ok := v.(int)
	return i, ok
}

// Delete 删除缓存条目
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Clear 清空所有缓存
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*Item)
	c.stats = CacheStats{}
}

// Len 返回缓存条目数（含过期但尚未清理的）
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Stats 返回缓存统计信息
func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats
}

// Keys 返回所有未过期的键（用于调试）
func (c *Cache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var keys []string
	now := time.Now()
	for k, v := range c.items {
		if v.ExpiresAt.IsZero() || !now.After(v.ExpiresAt) {
			keys = append(keys, k)
		}
	}
	return keys
}

// Stop 停止清理协程
func (c *Cache) Stop() {
	close(c.stop)
}

func main() {
	fmt.Println("=== Go 内存缓存示例 ===")
	fmt.Println()

	// 创建缓存，每 5 秒清理一次过期条目
	cache := NewCache(5 * time.Second)
	defer cache.Stop()

	// 1. 基本 Set/Get
	fmt.Println("--- 1. 基本 Set/Get ---")
	cache.Set("name", "张三", 0) // 永不过期
	cache.Set("age", 30, 0)

	if v, ok := cache.Get("name"); ok {
		fmt.Printf("name = %v (类型: %T)\n", v, v)
	}
	if v, ok := cache.GetInt("age"); ok {
		fmt.Printf("age = %d\n", v)
	}
	fmt.Printf("当前缓存数: %d\n", cache.Len())

	// 2. 类型断言
	fmt.Println("\n--- 2. 类型断言 ---")
	cache.Set("score", 95.5, 0)
	if v, ok := cache.Get("score"); ok {
		if score, ok := v.(float64); ok {
			fmt.Printf("score = %.1f\n", score)
		}
	}

	// 3. TTL 过期
	fmt.Println("\n--- 3. TTL 过期 ---")
	cache.Set("temp", "临时数据", 100*time.Millisecond)
	fmt.Printf("立即读取: %v\n", mustGet(cache, "temp"))
	time.Sleep(150 * time.Millisecond)
	if _, ok := cache.Get("temp"); !ok {
		fmt.Println("过期后读取: key 已过期（未找到）")
	}
	fmt.Printf("TTL 测试后缓存数: %d\n", cache.Len())

	// 4. 删除和清空
	fmt.Println("\n--- 4. Delete / Clear ---")
	cache.Set("a", 1, 0)
	cache.Set("b", 2, 0)
	fmt.Printf("删除前: %d\n", cache.Len())
	cache.Delete("a")
	fmt.Printf("删除 'a' 后: %d\n", cache.Len())
	cache.Clear()
	fmt.Printf("清空后: %d\n", cache.Len())

	// 5. 并发安全测试
	fmt.Println("\n--- 5. 并发安全 ---")
	var wg sync.WaitGroup
	concurrentCache := NewCache(0)

	// 启动 10 个 goroutine 同时写
	for i := range 10 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("key_%d", n)
			concurrentCache.Set(key, n*100, 0)
		}(i)
	}
	wg.Wait()
	fmt.Printf("并发写入后缓存数: %d\n", concurrentCache.Len())

	// 启动 10 个 goroutine 同时读
	for i := range 10 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("key_%d", n)
			if v, ok := concurrentCache.Get(key); ok {
				log.Printf("读取 %s = %v", key, v)
			}
		}(i)
	}
	wg.Wait()

	// 6. 统计信息
	fmt.Println("\n--- 6. 统计信息 ---")
	stats := cache.Stats()
	fmt.Printf("命中: %d, 未命中: %d\n", stats.Hits, stats.Misses)

	// 测试命中/未命中
	cache.Set("hit_test", "test", 0)
	cache.Get("hit_test")  // 命中
	cache.Get("nonexist")  // 未命中
	stats = cache.Stats()
	fmt.Printf("两次操作后 - 命中: %d, 未命中: %d\n", stats.Hits, stats.Misses)

	fmt.Println("\n✅ 缓存示例运行完成")
}

// mustGet 辅助函数：获取值或返回 "<nil>"
func mustGet(c *Cache, key string) string {
	v, ok := c.Get(key)
	if !ok {
		return "<nil>"
	}
	return fmt.Sprintf("%v", v)
}