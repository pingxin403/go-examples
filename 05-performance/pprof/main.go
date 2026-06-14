// pprof 性能分析示例
// 运行方式：
//
//	终端1: go run main.go
//	终端2: go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30
//
// 然后在浏览器中打开 http://localhost:8080 查看 CPU 火焰图、top、graph 等
// 其他有用的 profiling endpoint:
//   - /debug/pprof/heap          — 堆内存分配
//   - /debug/pprof/goroutine     — goroutine 堆栈
//   - /debug/pprof/allocs        — 所有分配历史
//   - /debug/pprof/block         — 同步原语阻塞
//   - /debug/pprof/mutex         — 互斥锁竞争
package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof" // 注册 pprof handler 到默认 ServeMux
	"runtime"
	"time"
)

func main() {
	// 启动 pprof HTTP server（goroutine）
	go func() {
		fmt.Println("pprof server listening on :6060")
		// pprof handler 通过 _ "net/http/pprof" 自动注册到默认 ServeMux
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// 让程序运行足够久以采集 profile 数据
	fmt.Println("程序正在运行，执行 CPU 密集和内存分配操作...")
	fmt.Println("在另一个终端运行: go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30")
	fmt.Println()

	// 混合执行任务，持续 60 秒
	done := make(chan struct{})
	go cpuWork(done)
	go memWork(done)
	go mixedWork(done)

	time.Sleep(60 * time.Second)
	close(done)
	time.Sleep(100 * time.Millisecond)
	fmt.Println("profile 采集完成")
}

// cpuWork 执行 CPU 密集型计算
func cpuWork(done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		default:
			// 递归版 Fibonacci — 大量重复计算，CPU 密集
			_ = fibRecursive(35)
			// 优化版 Fibonacci — 迭代，仍然消耗 CPU
			_ = fibOptimized(100000)
		}
	}
}

// fibRecursive 递归版 Fibonacci（指数级复杂度）
func fibRecursive(n int) int {
	if n <= 1 {
		return n
	}
	return fibRecursive(n-1) + fibRecursive(n-2)
}

// fibOptimized 迭代版 Fibonacci（线性复杂度）
func fibOptimized(n int) int {
	if n <= 1 {
		return n
	}
	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

// memWork 反复分配和丢弃大量对象，产生 GC 压力
func memWork(done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		default:
			allocateAndDiscard(10000)
			runtime.Gosched()
		}
	}
}

// allocateAndDiscard 分配大量短生命周期对象
func allocateAndDiscard(count int) {
	// 分配大量小对象
	data := make([]string, count)
	for i := 0; i < count; i++ {
		data[i] = fmt.Sprintf("object-%d-的数据内容: %s", i, randomString(64))
	}
	// 构造 map 并立即丢弃
	m := make(map[int][]byte)
	for i := 0; i < 500; i++ {
		m[i] = make([]byte, 1024)
		_ = m[i]
	}
	_ = data // 阻止编译器优化掉分配
}

// randomString 生成指定长度的随机字符串（模拟真实负载）
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[(i*17+31)%len(letters)]
	}
	return string(b)
}

// mixedWork 混合 CPU 计算和内存分配
func mixedWork(done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		default:
			// 计算素数
			_ = sieveOfEratosthenes(50000)
			// 分配并释放
			allocateAndDiscard(5000)
		}
	}
}

// sieveOfEratosthenes 埃拉托色尼筛法（CPU + 内存混合）
func sieveOfEratosthenes(limit int) []int {
	isPrime := make([]bool, limit+1)
	for i := 2; i <= limit; i++ {
		isPrime[i] = true
	}
	for i := 2; i*i <= limit; i++ {
		if isPrime[i] {
			for j := i * i; j <= limit; j += i {
				isPrime[j] = false
			}
		}
	}
	var primes []int
	for i := 2; i <= limit; i++ {
		if isPrime[i] {
			primes = append(primes, i)
		}
	}
	return primes
}