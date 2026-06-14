// scheduler/main.go — Go 调度器核心概念演示
//
// 展示 GOMAXPROCS、goroutine 与 OS 线程的关系、主动让出、
// goroutine 数量监控、线程锁定等调度器相关特性。
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main() {
	// ============================================================
	// 1. GOMAXPROCS 对并行度的影响
	// ============================================================
	fmt.Println("=== 1. GOMAXPROCS 与并行度 ===")

	// 显示当前 GOMAXPROCS 值
	fmt.Printf("默认 GOMAXPROCS = %d (CPU 核心数)\n", runtime.GOMAXPROCS(0))

	// 分别用 GOMAXPROCS=1 和 GOMAXPROCS=N 演示
	for _, procs := range []int{1, runtime.NumCPU()} {
		runtime.GOMAXPROCS(procs)
		startProcs := time.Now()

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				// 纯计算任务，无 IO 阻塞
				_ = fibonacci(30)
			}(i)
		}
		wg.Wait()

		fmt.Printf("GOMAXPROCS=%d: 100 个 goroutine 计算 fibonacci(30) 耗时 %v\n",
			procs, time.Since(startProcs))
	}
	// 恢复为 NumCPU
	runtime.GOMAXPROCS(runtime.NumCPU())

	// ============================================================
	// 2. Goroutine 不是 OS 线程 — 大量 goroutine 也能调度
	// ============================================================
	fmt.Println("\n=== 2. Goroutine 不是 OS 线程 ===")

	var wg2 sync.WaitGroup
	startGoroutines := 100000 // 启动 10 万个 goroutine
	fmt.Printf("启动 %d 个 goroutine...\n", startGoroutines)

	startGoroutine := time.Now()
	for i := 0; i < startGoroutines; i++ {
		wg2.Add(1)
		go func(id int) {
			defer wg2.Done()
			_ = id * id
		}(i)
	}
	wg2.Wait()
	elapsed := time.Since(startGoroutine)
	fmt.Printf("%d 个 goroutine 全部执行完毕，耗时 %v (如果是 OS 线程不可能这么轻量)\n",
		startGoroutines, elapsed)

	// ============================================================
	// 3. runtime.Gosched() 主动让出 CPU
	// ============================================================
	fmt.Println("\n=== 3. runtime.Gosched() 主动让出 CPU ===")

	ch := make(chan int)
	go func() {
		for i := 0; i < 5; i++ {
			// 模拟在某个循环中主动让出，给其他 goroutine 执行机会
			runtime.Gosched()
			ch <- i
		}
		close(ch)
	}()

	for v := range ch {
		fmt.Printf("收到: %d\n", v)
	}

	// 演示无 Gosched 时的饥饿
	fmt.Println("\n--- 无 Gosched 可能导致忙等 goroutine 饿死 ---")
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				// 忙等循环，不主动让出
			}
		}
	}()

	// 给忙等 goroutine 一点时间
	time.Sleep(10 * time.Millisecond)
	close(done)
	fmt.Println("忙等 goroutine 已关闭（在单核下它可能饿死其他 goroutine）")

	// ============================================================
	// 4. runtime.NumGoroutine() 监控 goroutine 数量
	// ============================================================
	fmt.Println("\n=== 4. runtime.NumGoroutine() 监控 goroutine 数量 ===")

	fmt.Printf("当前 goroutine 数量: %d\n", runtime.NumGoroutine())

	var wg3 sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg3.Add(1)
		go func(id int) {
			defer wg3.Done()
			time.Sleep(50 * time.Millisecond)
		}(i)
	}
	fmt.Printf("启动 10 个 goroutine 后: %d\n", runtime.NumGoroutine())
	wg3.Wait()
	fmt.Printf("全部完成后: %d\n", runtime.NumGoroutine())

	// ============================================================
	// 5. runtime.LockOSThread — 线程锁定
	// ============================================================
	fmt.Println("\n=== 5. runtime.LockOSThread 线程锁定 ===")

	// 启动一个 goroutine 并锁定到特定 OS 线程
	locked := make(chan bool)
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		fmt.Printf("已锁定到 OS 线程 (goroutine ID 不可直接获取，仅演示)\n")

		// 在这个 goroutine 中执行 CGO 等需要固定线程的操作
		_ = fibonacci(10)

		close(locked)
	}()
	<-locked
	fmt.Println("LockOSThread 示例完成")

	// ============================================================
	// 6. runtime.GOMAXPROCS(0) 与 runtime.NumCPU 信息汇总
	// ============================================================
	fmt.Println("\n=== 6. 运行时信息汇总 ===")
	fmt.Printf("NumCPU:       %d\n", runtime.NumCPU())
	fmt.Printf("GOMAXPROCS:   %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("NumGoroutine: %d\n", runtime.NumGoroutine())
	fmt.Printf("NumCgoCall:   %d\n", runtime.NumCgoCall())
	fmt.Printf("GOARCH:       %s\n", runtime.GOARCH)
	fmt.Printf("GOOS:         %s\n", runtime.GOOS)
	fmt.Printf("Version:      %s\n", runtime.Version())

	fmt.Println("\n调度器概念演示完毕 ✓")
}

// fibonacci — 纯计算函数，用于模拟 CPU 密集任务
func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}