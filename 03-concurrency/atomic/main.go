package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================
// atomic 包 — 原子操作示例
// 涵盖：AddInt64 / Load+Store / CompareAndSwap / atomic.Value
// 并与 Mutex 做性能对比
// ============================================================

func main() {
	// ============================================================
	// 1. atomic.AddInt64 — 原子计数器
	// ============================================================
	fmt.Println("=== 1. atomic.AddInt64 原子计数器 ===")

	var atomicCounter int64
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt64(&atomicCounter, 1) // 原子递增，无需锁
		}()
	}
	wg.Wait()
	fmt.Printf("  原子计数器: %d（期望 1000）\n", atomicCounter)

	// ============================================================
	// 2. atomic.Load / atomic.Store — 安全读写
	// ============================================================
	fmt.Println("\n=== 2. atomic.Load / atomic.Store 安全传播 ===")

	var status int64 // 0=空闲, 1=忙碌

	// 一个 goroutine 更新状态
	go func() {
		time.Sleep(50 * time.Millisecond)
		atomic.StoreInt64(&status, 1) // 安全写入
		fmt.Println("  状态已更新为忙碌（Store）")
	}()

	// 主 goroutine 轮询读取
	for atomic.LoadInt64(&status) == 0 {
		fmt.Println("  等待状态变更...（Load）")
		time.Sleep(10 * time.Millisecond)
	}
	fmt.Println("  检测到状态已变为忙碌 ✅")

	// ============================================================
	// 3. atomic.CompareAndSwap — CAS 实现自旋锁
	// ============================================================
	fmt.Println("\n=== 3. CompareAndSwap — CAS 自旋锁 ===")

	var spinLock int64 // 0=未锁定, 1=已锁定
	var shared int64
	var wgSpin sync.WaitGroup

	// CAS 自旋锁函数：不断尝试获得锁，直到成功
	spinLockFn := func(id int) {
		defer wgSpin.Done()

		// 自旋等待获取锁
		for !atomic.CompareAndSwapInt64(&spinLock, 0, 1) {
			// CAS 失败，自旋等待
			time.Sleep(time.Microsecond) // 避免忙等过热
		}
		// 获得锁
		shared++
		fmt.Printf("  Goroutine #%d 获得自旋锁, shared=%d\n", id, shared)
		time.Sleep(10 * time.Millisecond)

		// 释放锁
		atomic.StoreInt64(&spinLock, 0)
	}

	for i := 1; i <= 3; i++ {
		wgSpin.Add(1)
		go spinLockFn(i)
	}
	wgSpin.Wait()
	fmt.Printf("  ✅ 自旋锁保护后 shared=%d\n", shared)

	// ============================================================
	// 4. atomic.Value — 类型安全的原子值
	// ============================================================
	fmt.Println("\n=== 4. atomic.Value — 类型安全原子值 ===")

	var config atomic.Value // 存储任意类型（interface{}）

	// 初始配置
	config.Store(map[string]string{
		"host": "localhost",
		"port": "3306",
	})

	var wgCfg sync.WaitGroup

	// 读 goroutine — 反复读取配置
	for i := 1; i <= 3; i++ {
		wgCfg.Add(1)
		go func(id int) {
			defer wgCfg.Done()
			for j := 0; j < 2; j++ {
				cfg := config.Load().(map[string]string) // 类型断言
				fmt.Printf("  Reader #%d: host=%s, port=%s\n", id, cfg["host"], cfg["port"])
				time.Sleep(20 * time.Millisecond)
			}
		}(i)
	}

	// 写 goroutine — 更新配置
	time.Sleep(30 * time.Millisecond)
	config.Store(map[string]string{
		"host": "prod-server",
		"port": "5432",
	})
	fmt.Println("  配置已更新（Store）")

	wgCfg.Wait()

	// ============================================================
	// 5. Mutex vs Atomic 对比
	// ============================================================
	fmt.Println("\n=== 5. Mutex vs Atomic 性能对比 ===")

	const iterations = 1_000_000

	// Atomic 版本
	var atomicResult int64
	startAtomic := time.Now()

	var wgA sync.WaitGroup
	for i := 0; i < 4; i++ {
		wgA.Add(1)
		go func() {
			defer wgA.Done()
			for j := 0; j < iterations/4; j++ {
				atomic.AddInt64(&atomicResult, 1)
			}
		}()
	}
	wgA.Wait()
	atomicTime := time.Since(startAtomic)

	// Mutex 版本
	var mu sync.Mutex
	var mutexResult int64
	startMutex := time.Now()

	var wgM sync.WaitGroup
	for i := 0; i < 4; i++ {
		wgM.Add(1)
		go func() {
			defer wgM.Done()
			for j := 0; j < iterations/4; j++ {
				mu.Lock()
				mutexResult++
				mu.Unlock()
			}
		}()
	}
	wgM.Wait()
	mutexTime := time.Since(startMutex)

	fmt.Printf("  Atomic.AddInt64:  %v (result=%d)\n", atomicTime, atomicResult)
	fmt.Printf("  sync.Mutex:      %v (result=%d)\n", mutexTime, mutexResult)
	fmt.Printf("  速度比: Atomic 是 Mutex 的 %.2f 倍\n",
		float64(mutexTime)/float64(atomicTime))

	fmt.Println("\n✅ 所有 atomic 示例完成")
}