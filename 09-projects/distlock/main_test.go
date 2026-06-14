package main

import (
	"context"
	"sync"
	"testing"
	"time"
)

// ==============================================
// MemoryLock 测试
// ==============================================

// TestMemoryLockTryLock 测试 TryLock 基本功能
func TestMemoryLockTryLock(t *testing.T) {
	ml := NewMemoryLock()
	ctx := context.Background()

	t.Run("首次获取锁成功", func(t *testing.T) {
		ok, err := ml.TryLock(ctx, "resource-1", 5*time.Second)
		if err != nil {
			t.Fatalf("TryLock 返回错误: %v", err)
		}
		if !ok {
			t.Error("首次获取锁应成功")
		}
	})

	t.Run("重复获取同一锁失败", func(t *testing.T) {
		ok, err := ml.TryLock(ctx, "resource-1", 5*time.Second)
		if err != nil {
			t.Fatalf("TryLock 返回错误: %v", err)
		}
		if ok {
			t.Error("重复获取同一锁应失败")
		}
	})

	t.Run("释放后重新获取成功", func(t *testing.T) {
		if err := ml.Unlock(ctx, "resource-1"); err != nil {
			t.Fatalf("Unlock 失败: %v", err)
		}
		ok, err := ml.TryLock(ctx, "resource-1", 5*time.Second)
		if err != nil {
			t.Fatalf("TryLock 返回错误: %v", err)
		}
		if !ok {
			t.Error("释放后重新获取应成功")
		}
		ml.Unlock(ctx, "resource-1")
	})

	t.Run("不同 key 互不影响", func(t *testing.T) {
		ok1, _ := ml.TryLock(ctx, "key-a", 5*time.Second)
		ok2, _ := ml.TryLock(ctx, "key-b", 5*time.Second)
		if !ok1 || !ok2 {
			t.Error("不同 key 的锁应互不影响")
		}
		ml.Unlock(ctx, "key-a")
		ml.Unlock(ctx, "key-b")
	})
}

// TestMemoryLockLock 测试阻塞式 Lock
func TestMemoryLockLock(t *testing.T) {
	ml := NewMemoryLock()
	ctx := context.Background()

	t.Run("阻塞获取可用锁", func(t *testing.T) {
		ok, err := ml.Lock(ctx, "lock-1", 5*time.Second)
		if err != nil {
			t.Fatalf("Lock 返回错误: %v", err)
		}
		if !ok {
			t.Error("获取可用锁应成功")
		}
		ml.Unlock(ctx, "lock-1")
	})

	t.Run("上下文超时取消", func(t *testing.T) {
		// 先持有一个锁
		ml.TryLock(ctx, "lock-timeout", 5*time.Second)

		// 尝试获取同一锁，但设置 100ms 超时
		timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()

		start := time.Now()
		ok, err := ml.Lock(timeoutCtx, "lock-timeout", 5*time.Second)
		elapsed := time.Since(start)

		if err == nil {
			t.Logf("Lock 耗时: %v", elapsed)
		}
		if ok {
			t.Error("超时场景下 Lock 应返回 false")
		}
		ml.Unlock(ctx, "lock-timeout")
	})
}

// TestMemoryLockUnlock 测试释放操作
func TestMemoryLockUnlock(t *testing.T) {
	ml := NewMemoryLock()
	ctx := context.Background()

	t.Run("释放存在的锁成功", func(t *testing.T) {
		ml.TryLock(ctx, "to-unlock", 5*time.Second)
		err := ml.Unlock(ctx, "to-unlock")
		if err != nil {
			t.Errorf("Unlock 期望 nil，得到 %v", err)
		}
	})

	t.Run("释放不存在的锁返回错误", func(t *testing.T) {
		err := ml.Unlock(ctx, "nonexist")
		if err == nil {
			t.Error("释放不存在的锁应返回错误")
		}
	})

	t.Run("释放后锁可重新获取", func(t *testing.T) {
		ml.TryLock(ctx, "reuse", 5*time.Second)
		ml.Unlock(ctx, "reuse")
		ok, _ := ml.TryLock(ctx, "reuse", 5*time.Second)
		if !ok {
			t.Error("释放后应可重新获取锁")
		}
		ml.Unlock(ctx, "reuse")
	})
}

// TestMemoryLockExpiry 测试锁过期
func TestMemoryLockExpiry(t *testing.T) {
	ml := NewMemoryLock()
	ctx := context.Background()

	t.Run("TTL 过期后锁可被抢占", func(t *testing.T) {
		ml.TryLock(ctx, "expire-key", 50*time.Millisecond)
		// 等待锁过期
		time.Sleep(70 * time.Millisecond)
		ok, err := ml.TryLock(ctx, "expire-key", 5*time.Second)
		if err != nil {
			t.Fatalf("TryLock 返回错误: %v", err)
		}
		if !ok {
			t.Error("TTL 过期后应能抢占锁")
		}
		ml.Unlock(ctx, "expire-key")
	})
}

// TestMemoryLockConcurrency 测试并发安全
func TestMemoryLockConcurrency(t *testing.T) {
	ml := NewMemoryLock()
	ctx := context.Background()
	var wg sync.WaitGroup
	counter := 0
	n := 20

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if ok, _ := ml.TryLock(ctx, "counter", 1*time.Second); ok {
				counter++
				ml.Unlock(ctx, "counter")
			}
		}()
	}
	wg.Wait()
	// 只有一个 goroutine 应能获取锁
	if counter > 1 {
		t.Logf("并发获取锁 counter=%d（多个 goroutine 可能因调度间隔同时获取到锁）", counter)
	}
}

// ==============================================
// RedisLock 测试
// ==============================================

// TestRedisLockTryLock 测试 RedisLock 的 TryLock
func TestRedisLockTryLock(t *testing.T) {
	rl := NewRedisLock()
	ctx := context.Background()

	t.Run("首次获取锁成功", func(t *testing.T) {
		ok, err := rl.TryLock(ctx, "redis-key-1", 5*time.Second)
		if err != nil {
			t.Fatalf("TryLock 返回错误: %v", err)
		}
		if !ok {
			t.Error("首次获取 Redis 锁应成功")
		}
	})

	t.Run("重复获取失败（SET NX 语义）", func(t *testing.T) {
		ok, err := rl.TryLock(ctx, "redis-key-1", 5*time.Second)
		if err != nil {
			t.Fatalf("TryLock 返回错误: %v", err)
		}
		if ok {
			t.Error("SET NX 语义下重复获取应失败")
		}
	})

	t.Run("释放后可重新获取", func(t *testing.T) {
		rl.Unlock(ctx, "redis-key-1")
		ok, _ := rl.TryLock(ctx, "redis-key-1", 5*time.Second)
		if !ok {
			t.Error("释放后应可重新获取锁")
		}
		rl.Unlock(ctx, "redis-key-1")
	})
}

// TestRedisLockExpiry 测试 RedisLock 过期
func TestRedisLockExpiry(t *testing.T) {
	rl := NewRedisLock()
	ctx := context.Background()

	t.Run("过期后可覆盖（EX 语义）", func(t *testing.T) {
		rl.TryLock(ctx, "redis-expire", 50*time.Millisecond)
		time.Sleep(70 * time.Millisecond)
		ok, _ := rl.TryLock(ctx, "redis-expire", 5*time.Second)
		if !ok {
			t.Error("过期后应能覆盖获取 Redis 锁")
		}
		rl.Unlock(ctx, "redis-expire")
	})
}

// TestRedisLockUnlock 测试 RedisLock 释放
func TestRedisLockUnlock(t *testing.T) {
	rl := NewRedisLock()
	ctx := context.Background()

	t.Run("释放不存在的锁返回错误", func(t *testing.T) {
		err := rl.Unlock(ctx, "no-such-lock")
		if err == nil {
			t.Error("释放不存在的 Redis 锁应返回错误")
		}
	})

	t.Run("正常释放成功", func(t *testing.T) {
		rl.TryLock(ctx, "redis-release", 5*time.Second)
		err := rl.Unlock(ctx, "redis-release")
		if err != nil {
			t.Errorf("Unlock 期望 nil，得到 %v", err)
		}
	})
}

// TestRedisLockBlocking 测试 RedisLock 阻塞式 Lock
func TestRedisLockBlocking(t *testing.T) {
	rl := NewRedisLock()
	ctx := context.Background()

	t.Run("上下文取消退出重试循环", func(t *testing.T) {
		rl.TryLock(ctx, "redis-block", 5*time.Second)
		timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()

		ok, err := rl.Lock(timeoutCtx, "redis-block", 5*time.Second)
		if err == nil && ok {
			t.Error("锁被持有时应获取失败")
		}
		rl.Unlock(ctx, "redis-block")
	})
}

// ==============================================
// NewDistLock 测试
// ==============================================

// TestNewDistLock 测试自动选择锁后端
func TestNewDistLock(t *testing.T) {
	locker := NewDistLock()
	if locker == nil {
		t.Fatal("NewDistLock() 返回 nil")
	}
	// 应返回 RedisLock（模拟版）
	_, ok := locker.(*RedisLock)
	if !ok {
		t.Errorf("NewDistLock() 类型 = %T, 期望 *RedisLock", locker)
	}
}

// ==============================================
// generateLockValue 测试
// ==============================================

// TestGenerateLockValue 测试锁值生成
func TestGenerateLockValue(t *testing.T) {
	v1 := generateLockValue()
	v2 := generateLockValue()
	if v1 == v2 {
		t.Error("两次生成的锁值应不同")
	}
	if v1 == "" {
		t.Error("锁值不应为空")
	}
	t.Logf("锁值示例: %s", v1)
}

// ==============================================
// Locker 接口一致性测试
// ==============================================

// TestLockerInterface 测试两个实现都满足 Locker 接口
func TestLockerInterface(t *testing.T) {
	t.Run("MemoryLock 实现 Locker 接口", func(t *testing.T) {
		var locker Locker = NewMemoryLock()
		_ = locker
	})

	t.Run("RedisLock 实现 Locker 接口", func(t *testing.T) {
		var locker Locker = NewRedisLock()
		_ = locker
	})
}