package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ============================================================
// sync 包并发原语示例
// 涵盖：WaitGroup / Mutex / RWMutex / Once / Cond / Map
// ============================================================

func main() {
	// ============================================================
	// 1. WaitGroup — 协调多个 goroutine 完成
	// ============================================================
	fmt.Println("=== 1. WaitGroup — 等待多个 goroutine ===")

	var wg sync.WaitGroup

	for i := 1; i <= 3; i++ {
		wg.Add(1) // 计数器 +1
		go func(id int) {
			defer wg.Done() // 计数器 -1
			workTime := time.Duration(rand.Intn(50)+20) * time.Millisecond
			time.Sleep(workTime)
			fmt.Printf("  任务 %d 完成（耗时 %v）\n", id, workTime)
		}(i)
	}

	wg.Wait() // 阻塞直到计数器归零
	fmt.Println("  ✅ 所有任务完成")

	// ============================================================
	// 2. Mutex — 保护共享资源（互斥锁）
	// ============================================================
	fmt.Println("\n=== 2. Mutex — 互斥锁保护共享计数器 ===")

	var (
		mu         sync.Mutex
		counter int
	)
	var wg2 sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			mu.Lock()
			counter++ // 如果不加锁，会出现数据竞争
			mu.Unlock()
		}()
	}

	wg2.Wait()
	fmt.Printf("  无锁会导致数据竞争，使用 Mutex 后 counter=%d（期望 1000）\n", counter)

	// ============================================================
	// 3. RWMutex — 读写锁（读共享、写独占）
	// ============================================================
	fmt.Println("\n=== 3. RWMutex — 读写锁 ===")

	var (
		rwmu      sync.RWMutex
		sharedData int
	)
	var wg3 sync.WaitGroup

	// 多个读 goroutine — 可以同时读
	for i := 1; i <= 3; i++ {
		wg3.Add(1)
		go func(id int) {
			defer wg3.Done()
			rwmu.RLock() // 读锁：多个 goroutine 可同时持有
			fmt.Printf("  读 Goroutine #%d: data=%d（RLock 可并发）\n", id, sharedData)
			time.Sleep(10 * time.Millisecond)
			rwmu.RUnlock()
		}(i)
	}

	// 写 goroutine — 必须独占
	wg3.Add(1)
	go func() {
		defer wg3.Done()
		time.Sleep(5 * time.Millisecond) // 等读 goroutine 先获得读锁
		rwmu.Lock()                      // 写锁：需要等待所有读锁释放
		fmt.Printf("  写 Goroutine: data 从 %d 改为 %d（独占写锁）\n", sharedData, 42)
		sharedData = 42
		time.Sleep(10 * time.Millisecond)
		rwmu.Unlock()
	}()

	wg3.Wait()

	// ============================================================
	// 4. Once — 只执行一次（用于单例初始化）
	// ============================================================
	fmt.Println("\n=== 4. Once — 只执行一次 ===")

	var once sync.Once
	var config string

	initConfig := func() {
		config = "database:localhost:3306"
		fmt.Println("  配置已初始化（仅一次）")
	}

	// 模拟多个 goroutine 同时尝试初始化
	var wg4 sync.WaitGroup
	for i := 1; i <= 5; i++ {
		wg4.Add(1)
		go func(id int) {
			defer wg4.Done()
			once.Do(initConfig) // 即使 5 个 goroutine 调用，initConfig 只执行一次
			fmt.Printf("  Goroutine #%d 读取 config=%q\n", id, config)
		}(i)
	}
	wg4.Wait()

	// ============================================================
	// 5. Cond — 条件变量（广播通知）
	// ============================================================
	fmt.Println("\n=== 5. Cond — 条件变量广播 ===")

	var (
		muCond   sync.Mutex
		cond     = sync.NewCond(&muCond)
		ready    bool
	)
	var wg5 sync.WaitGroup

	// 多个等待者
	for i := 1; i <= 3; i++ {
		wg5.Add(1)
		go func(id int) {
			defer wg5.Done()
			muCond.Lock()
			for !ready { // 必须用 for 循环，不能用 if（防止虚假唤醒）
				cond.Wait() // 释放锁并等待 Broadcast/Signal
			}
			muCond.Unlock()
			fmt.Printf("  Worker #%d 收到信号，开始工作\n", id)
		}(i)
	}

	time.Sleep(50 * time.Millisecond) // 确保所有 worker 已进入 Wait

	// 发送广播
	muCond.Lock()
	ready = true
	cond.Broadcast() // 唤醒所有等待者（Signal 只唤醒一个）
	muCond.Unlock()

	wg5.Wait()
	fmt.Println("  所有 worker 已响应广播")

	// ============================================================
	// 6. sync.Map — 并发安全 Map（适合读多写少的场景）
	// ============================================================
	fmt.Println("\n=== 6. sync.Map — 并发安全字典 ===")

	var safeMap sync.Map
	var wg6 sync.WaitGroup

	// 并发写入
	for i := 1; i <= 5; i++ {
		wg6.Add(1)
		go func(id int) {
			defer wg6.Done()
			safeMap.Store(fmt.Sprintf("key-%d", id), id*100)
		}(i)
	}

	wg6.Wait()

	// 读取 + 遍历
	safeMap.Range(func(key, value interface{}) bool {
		fmt.Printf("  sync.Map 条目: %v -> %v\n", key, value)
		return true // 继续遍历；返回 false 停止
	})

	// LoadOrStore — 存在就读，不存在就写
	actual, loaded := safeMap.LoadOrStore("key-1", 999)
	fmt.Printf("  LoadOrStore key-1: 已存在=%t, value=%v\n", loaded, actual)

	// LoadAndDelete — 读取并删除
	val, deleted := safeMap.LoadAndDelete("key-2")
	fmt.Printf("  LoadAndDelete key-2: value=%v, 存在=%t\n", val, deleted)

	// 再次 Load — 确认已删除
	_, found := safeMap.Load("key-2")
	fmt.Printf("  Load key-2 删除后: 存在=%t\n", found)

	fmt.Println("\n✅ 所有 sync 包示例完成")
}