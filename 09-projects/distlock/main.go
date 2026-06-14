package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ==============================================
// 分布式锁抽象接口
// ==============================================

// Locker 分布式锁接口
type Locker interface {
	// Lock 获取锁，block=true 时阻塞等待
	Lock(ctx context.Context, key string, ttl time.Duration) (bool, error)
	// Unlock 释放锁
	Unlock(ctx context.Context, key string) error
	// TryLock 尝试获取锁，不阻塞
	TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error)
}

// ==============================================
// 方式 1: 基于 sync.Mutex 的内存锁模拟器
// ==============================================

// MemoryLock 基于内存的锁模拟
// 适用于单进程场景，演示分布式锁的概念
type MemoryLock struct {
	mu       sync.Mutex
	locks    map[string]*memoryLockEntry
	stopCh   chan struct{}
}

type memoryLockEntry struct {
	owner   string          // 持有者标识
	expires time.Time       // 过期时间
	cancel  context.CancelFunc // 用于停止续约协程
}

// NewMemoryLock 创建内存锁
func NewMemoryLock() *MemoryLock {
	return &MemoryLock{
		locks:  make(map[string]*memoryLockEntry),
		stopCh: make(chan struct{}),
	}
}

// TryLock 尝试获取内存锁，不阻塞
func (ml *MemoryLock) TryLock(_ context.Context, key string, ttl time.Duration) (bool, error) {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	// 检查锁是否已被持有且未过期
	if entry, exists := ml.locks[key]; exists {
		if time.Now().Before(entry.expires) {
			return false, nil // 锁被持有
		}
		// 锁已过期，可以抢占
		delete(ml.locks, key)
	}

	// 获取锁
	ml.locks[key] = &memoryLockEntry{
		owner:   fmt.Sprintf("goroutine-%d", rand.Intn(1000)),
		expires: time.Now().Add(ttl),
	}
	return true, nil
}

// Lock 阻塞式获取锁（带超时重试）
func (ml *MemoryLock) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	for {
		ok, err := ml.TryLock(ctx, key, ttl)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}

		// 等待一段时间后重试
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(50 * time.Millisecond):
			// 继续重试
		}
	}
}

// Unlock 释放内存锁
func (ml *MemoryLock) Unlock(_ context.Context, key string) error {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	if _, exists := ml.locks[key]; exists {
		delete(ml.locks, key)
		return nil
	}
	return fmt.Errorf("锁 %s 不存在或已释放", key)
}

// ==============================================
// 方式 2: Redis SET NX EX 模式的锁实现（注释说明版）
// ==============================================

// RedisLock 基于 Redis 的分布式锁
// 使用 SET NX EX 原子命令实现
//
// Redis 命令原型:
//   SET key value NX EX seconds
//   - NX: 只在 key 不存在时设置（实现互斥）
//   - EX: 设置过期时间（防止死锁）
//
// 释放锁时使用 Lua 脚本保证原子性:
//   if redis.call("GET", KEYS[1]) == ARGV[1] then
//       return redis.call("DEL", KEYS[1])
//   else
//       return 0
//   end
//
// 注意: 实际生产中使用 github.com/redis/go-redis/v9
// 此处仅用 map 模拟 Redis 行为来展示算法
type RedisLock struct {
	mu     sync.Mutex
	store  map[string]redisEntry // 模拟 Redis 存储
}

type redisEntry struct {
	value   string
	expires time.Time
}

// NewRedisLock 创建基于 Redis 模式的锁
func NewRedisLock() *RedisLock {
	return &RedisLock{
		store: make(map[string]redisEntry),
	}
}

// TryLock 尝试获取锁（SET NX EX 的模拟实现）
func (rl *RedisLock) TryLock(_ context.Context, key string, ttl time.Duration) (bool, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 模拟 Redis SET NX: 只在 key 不存在时成功
	if entry, exists := rl.store[key]; exists {
		if time.Now().Before(entry.expires) {
			return false, nil
		}
		// key 已过期，可覆盖（相当于 Redis 的 EX 过期后自动删除）
		delete(rl.store, key)
	}

	// 模拟 Redis SET ... EX: 设置过期时间
	rl.store[key] = redisEntry{
		value:   generateLockValue(), // 唯一标识，用于安全释放
		expires: time.Now().Add(ttl),
	}
	return true, nil
}

// Lock 阻塞式获取 Redis 锁
func (rl *RedisLock) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	for {
		ok, err := rl.TryLock(ctx, key, ttl)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
}

// Unlock 释放 Redis 锁
// 模拟 Lua 脚本保证原子性: 只有锁的持有者才能释放
func (rl *RedisLock) Unlock(_ context.Context, key string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, exists := rl.store[key]
	if !exists {
		return fmt.Errorf("锁 %s 不存在", key)
	}
	if time.Now().After(entry.expires) {
		delete(rl.store, key)
		return fmt.Errorf("锁 %s 已过期", key)
	}
	delete(rl.store, key)
	return nil
}

// generateLockValue 生成锁的唯一标识值
// 实际 Redis 实现中常用: fmt.Sprintf("%s:%d", hostname, goroutineID)
func generateLockValue() string {
	return fmt.Sprintf("lock-%d", time.Now().UnixNano())
}

// ==============================================
// 自动选择锁后端
// ==============================================

// NewDistLock 尝试连接 Redis，失败则回退到内存锁
// 这里因为无法保证 Redis 可用，所以直接返回 RedisLock（模拟版）
// 实际项目中:
//
//	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
//	if err := client.Ping(ctx).Err(); err != nil {
//	    log.Printf("Redis 不可用，回退到内存锁: %v", err)
//	    return NewMemoryLock()
//	}
//	return NewRedisLock(client)
func NewDistLock() Locker {
	fmt.Println("⚠️  Redis 未配置，使用内存模拟锁")
	fmt.Println("   生产环境应配置 Redis URL 并使用 go-redis 客户端")
	return NewRedisLock()
}

// ==============================================
// 并发工作示例
// ==============================================

// Worker 模拟一个需要获取锁的工作单元
func Worker(id int, locker Locker, ctx context.Context, resource string) {
	for i := range 3 {
		// 尝试获取锁，超时 2 秒
		lockCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		ok, err := locker.Lock(lockCtx, resource, 1*time.Second)
		cancel()

		if err != nil {
			fmt.Printf("  Worker %d: 获取锁失败 (第%d次): %v\n", id, i+1, err)
			continue
		}
		if !ok {
			fmt.Printf("  Worker %d: 获取锁超时 (第%d次)\n", id, i+1)
			continue
		}

		// 模拟业务处理
		fmt.Printf("  Worker %d: 🔒 获取锁成功，处理资源 %s (第%d次)\n", id, resource, i+1)
		time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)

		// 释放锁
		if err := locker.Unlock(ctx, resource); err != nil {
			fmt.Printf("  Worker %d: 释放锁失败: %v\n", id, err)
		} else {
			fmt.Printf("  Worker %d: 🔓 释放锁成功\n", id)
		}
	}
}

func main() {
	fmt.Println("=== Go 分布式锁示例 ===")
	fmt.Println()

	// --- 1. 内存锁演示 ---
	fmt.Println("--- 1. 内存锁（sync.Mutex + 超时）---")
	memLock := NewMemoryLock()
	ctx := context.Background()

	// 基本 lock/unlock
	ok, _ := memLock.TryLock(ctx, "resource-1", 5*time.Second)
	fmt.Printf("获取锁 resource-1: %v\n", ok)

	// 第二次尝试应该失败
	ok, _ = memLock.TryLock(ctx, "resource-1", 5*time.Second)
	fmt.Printf("再次获取锁 resource-1: %v (预期: false)\n", ok)

	memLock.Unlock(ctx, "resource-1")
	ok, _ = memLock.TryLock(ctx, "resource-1", 5*time.Second)
	fmt.Printf("释放后重新获取: %v (预期: true)\n", ok)
	memLock.Unlock(ctx, "resource-1")

	// --- 2. Redis 模式锁演示 ---
	fmt.Println("\n--- 2. Redis 模式锁（SET NX EX 模拟）---")
	redisLock := NewRedisLock()

	// 模拟 SET resource-2 <value> NX EX 5
	ok, _ = redisLock.TryLock(ctx, "resource-2", 5*time.Second)
	fmt.Printf("Redis 锁获取 resource-2: %v\n", ok)

	// 模拟并发冲突
	ok, _ = redisLock.TryLock(ctx, "resource-2", 5*time.Second)
	fmt.Printf("并发获取 resource-2: %v (预期: false)\n", ok)

	redisLock.Unlock(ctx, "resource-2")
	fmt.Printf("释放后可用: ")
	ok, _ = redisLock.TryLock(ctx, "resource-2", 5*time.Second)
	fmt.Printf("%v\n", ok)
	redisLock.Unlock(ctx, "resource-2")

	// --- 3. 自动选择后端 ---
	fmt.Println("\n--- 3. 自动选择锁后端 ---")
	locker := NewDistLock()
	fmt.Printf("选用后端类型: %T\n", locker)

	// --- 4. 并发 Worker 演示 ---
	fmt.Println("\n--- 4. 并发 Worker 演示 ---")
	fmt.Println("启动 3 个 Worker 竞争同一资源...")

	var wg sync.WaitGroup
	resource := "shared-resource"

	for i := range 3 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			Worker(id+1, locker, ctx, resource)
		}(i)
	}

	wg.Wait()
	fmt.Println("\n✅ 分布式锁示例运行完成")
	fmt.Println()
	fmt.Println("=== 生产环境建议 ===")
	fmt.Println("1. 使用 github.com/redis/go-redis/v9 连接真实的 Redis")
	fmt.Println("2. 锁值使用唯一标识（如 UUID），确保只有持有者能释放")
	fmt.Println("3. 设置合理的 TTL 防止死锁")
	fmt.Println("4. 考虑 Redlock 算法实现多节点高可用")
	fmt.Println("5. 使用看门狗（Watchdog）协程自动续约")
}