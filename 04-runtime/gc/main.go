// gc/main.go — Go 垃圾回收机制演示
//
// 展示 runtime.GC() 手动触发 GC、ReadMemStats 内存统计、
// GC 频率控制、debug.SetGCPercent 调优等核心概念。
package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"time"
)

func main() {
	// ============================================================
	// 1. runtime.GC() 手动触发 GC + ReadMemStats 观察效果
	// ============================================================
	fmt.Println("=== 1. 手动触发 GC 与 MemStats ===")

	// 先分配一些内存
	_ = make([]byte, 10<<20) // 10MB

	printMemStats("分配 10MB 后")

	// 手动触发 GC
	runtime.GC()
	printMemStats("手动 GC 后")

	// ============================================================
	// 2. 分配大量对象，观察 GC 前后的内存变化
	// ============================================================
	fmt.Println("\n=== 2. 大量分配后 GC 效果 ===")

	// 分配大量小对象
	const numObjs = 100000
	objs := make([][]byte, numObjs)
	for i := 0; i < numObjs; i++ {
		objs[i] = make([]byte, 1024) // 每个 1KB
	}
	fmt.Printf("分配 %d 个 1KB 对象 (共约 %.1f MB)\n", numObjs, float64(numObjs*1024)/(1<<20))

	printMemStats("大量分配后")

	// 释放引用
	objs = nil

	// 触发 GC
	runtime.GC()
	printMemStats("释放引用 + GC 后")

	// ============================================================
	// 3. NumGC 和 PauseTotalNs — GC 频率与暂停时间
	// ============================================================
	fmt.Println("\n=== 3. GC 频率与暂停时间 ===")

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	fmt.Printf("当前已发生 GC 次数  : %d\n", stats.NumGC)
	fmt.Printf("GC 总暂停时间       : %v\n", time.Duration(stats.PauseTotalNs))
	fmt.Printf("最近一次 GC 暂停    : %v\n", time.Duration(stats.PauseNs[(stats.NumGC-1)%256]))
	if stats.NumGC > 0 {
		total := time.Duration(stats.PauseTotalNs)
		avg := total / time.Duration(stats.NumGC)
		fmt.Printf("平均每次 GC 暂停时间 : %v\n", avg)
	}

	// 频繁分配大量对象，观察 GC 被触发的频率
	fmt.Println("\n--- 压力分配：观察 GC 频次 ---")
	for i := 0; i < 5; i++ {
		start := time.Now()

		// 分配和释放大量短生命周期对象
		for j := 0; j < 100000; j++ {
			_ = make([]byte, 4096)
		}

		runtime.ReadMemStats(&stats)
		fmt.Printf("第 %d 轮: 分配完成 (%v), 总 GC 次数: %d\n",
			i+1, time.Since(start), stats.NumGC)
	}

	// ============================================================
	// 4. debug.SetGCPercent — 调整 GC 触发频率
	// ============================================================
	fmt.Println("\n=== 4. debug.SetGCPercent 调整 GC 频率 ===")

	// 保存原始值
	oldGCPercent := debug.SetGCPercent(100) // 默认值
	fmt.Printf("默认 GCPercent = %d%%\n", oldGCPercent)

	// GCPercent = 100: 堆增长到上次 GC 后堆大小的 100% 时触发 GC
	// 值越小 GC 越频繁，内存使用越低，CPU 开销越大
	// 值越大 GC 越少，内存使用越高，CPU 开销越小

	// 演示在高 GCPercent 下（减少 GC）
	debug.SetGCPercent(1000) // 堆增长到 1000% 才触发 GC
	fmt.Println("设置 GCPercent = 1000%（减少 GC 频率）")

	var stats2 runtime.MemStats
	runtime.ReadMemStats(&stats2)
	startGC := stats2.NumGC

	// 大量分配
	for i := 0; i < 500000; i++ {
		_ = make([]byte, 512)
	}

	runtime.ReadMemStats(&stats2)
	fmt.Printf("GCPercent=1000: 分配 50 万次后 GC 增量: %d\n", stats2.NumGC-startGC)

	// 恢复默认
	debug.SetGCPercent(oldGCPercent)

	// 用低 GCPercent 再做一轮（GC 更频繁）
	debug.SetGCPercent(10) // 堆增长到 10% 就触发 GC
	fmt.Println("设置 GCPercent = 10%（增加 GC 频率）")

	runtime.ReadMemStats(&stats2)
	startGC = stats2.NumGC

	for i := 0; i < 500000; i++ {
		_ = make([]byte, 512)
	}

	runtime.ReadMemStats(&stats2)
	fmt.Printf("GCPercent=10:  分配 50 万次后 GC 增量: %d\n", stats2.NumGC-startGC)

	// 恢复默认值
	debug.SetGCPercent(oldGCPercent)

	// 显示 GOGC 环境变量说明
	fmt.Printf("\nGOGC 环境变量可控制全局 GC 频率，设置 SetGCPercent(%d) 等价于 GOGC=%d\n",
		oldGCPercent, oldGCPercent)

	// ============================================================
	// 5. 详细 MemStats 字段说明
	// ============================================================
	fmt.Println("\n=== 5. MemStats 字段详解 ===")

	runtime.GC()
	runtime.ReadMemStats(&stats)
	fmt.Printf("Alloc         = %v MB (当前堆内存使用量)\n", toMB(stats.Alloc))
	fmt.Printf("TotalAlloc    = %v MB (累计分配总量)\n", toMB(stats.TotalAlloc))
	fmt.Printf("Sys           = %v MB (从 OS 获取的总内存)\n", toMB(stats.Sys))
	fmt.Printf("Lookups       = %d (指针查找次数)\n", stats.Lookups)
	fmt.Printf("Mallocs       = %d (累计分配对象数)\n", stats.Mallocs)
	fmt.Printf("Frees         = %d (累计释放对象数)\n", stats.Frees)
	fmt.Printf("HeapAlloc     = %v MB (堆分配量)\n", toMB(stats.HeapAlloc))
	fmt.Printf("HeapSys       = %v MB (堆从 OS 申请的总量)\n", toMB(stats.HeapSys))
	fmt.Printf("HeapIdle      = %v MB (堆空闲量，可归还 OS)\n", toMB(stats.HeapIdle))
	fmt.Printf("HeapInuse     = %v MB (堆正在使用量)\n", toMB(stats.HeapInuse))
	fmt.Printf("HeapReleased  = %v MB (已归还 OS 的堆内存)\n", toMB(stats.HeapReleased))
	fmt.Printf("HeapObjects   = %d (堆中存活对象数)\n", stats.HeapObjects)
	fmt.Printf("NumGC         = %d (GC 总次数)\n", stats.NumGC)
	fmt.Printf("NumForcedGC   = %d (强制 GC 次数)\n", stats.NumForcedGC)
	fmt.Printf("GCCPUFraction = %.4f (GC 占 CPU 时间比)\n", stats.GCCPUFraction)
	fmt.Printf("PauseTotalNs  = %v (GC 总暂停时间)\n", time.Duration(stats.PauseTotalNs))
	fmt.Printf("LastGC        = %v (上次 GC 时间)\n", time.Unix(0, int64(stats.LastGC)))

	fmt.Println("\nGC 概念演示完毕 ✓")
}

// printMemStats 打印关键内存指标
func printMemStats(label string) {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	fmt.Printf("[%s] Alloc=%v MB  TotalAlloc=%v MB  Sys=%v MB  NumGC=%d\n",
		label,
		toMB(stats.Alloc),
		toMB(stats.TotalAlloc),
		toMB(stats.Sys),
		stats.NumGC,
	)
}

// toMB 将字节转换为 MB（浮点数）
func toMB(bytes uint64) float64 {
	return float64(bytes) / (1 << 20)
}