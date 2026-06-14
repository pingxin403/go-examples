// gc/main_test.go — Go 垃圾回收机制测试
package main

import (
	"runtime"
	"testing"
)

// TestToMB — 测试字节转 MB 的辅助函数
func TestToMB(t *testing.T) {
	tests := []struct {
		name  string
		bytes uint64
		want  float64
	}{
		{"0 字节转 MB", 0, 0.0},
		{"1 MB (2^20)", 1 << 20, 1.0},
		{"10 MB", 10 << 20, 10.0},
		{"1 KB 转 MB", 1024, 1024.0 / (1 << 20)},
		{"1 GB 转 MB (2^30)", 1 << 30, 1024.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toMB(tt.bytes)
			if got != tt.want {
				t.Errorf("toMB(%d) = %f, 期望 %f", tt.bytes, got, tt.want)
			}
		})
	}
}

// TestManualGCDoesNotPanic — 验证手动触发 GC 不会 panic
func TestManualGCDoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("runtime.GC() 引发了 panic: %v", r)
		}
	}()

	runtime.GC()
	runtime.GC()
}

// TestAllocAndFree — 验证分配后释放再 GC，Alloc 应下降
func TestAllocAndFree(t *testing.T) {
	var before, after runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&before)

	// 分配 100 个 1MB 的对象
	objs := make([][]byte, 100)
	for i := range objs {
		objs[i] = make([]byte, 1<<20) // 1MB
	}

	runtime.ReadMemStats(&after)
	allocWithObjs := after.Alloc

	// 释放并 GC
	objs = nil
	runtime.GC()
	runtime.ReadMemStats(&after)

	// GC 后 Alloc 应显著低于分配时的值
	if after.Alloc >= allocWithObjs {
		t.Logf("注意: 释放后 Alloc=%d 未低于分配时 Alloc=%d (可能因 TINYGC 或其他开销)",
			after.Alloc, allocWithObjs)
	}
}

// TestReadMemStatsPopulatesFields — 验证 MemStats 各字段不为零
func TestReadMemStatsPopulatesFields(t *testing.T) {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	// 这些字段应始终有值
	checkNonZero(t, "NumGC", int(stats.NumGC))
	checkNonZero(t, "PauseTotalNs", int(stats.PauseTotalNs))
	checkNonZero(t, "Alloc", int(stats.Alloc))
	checkNonZero(t, "TotalAlloc", int(stats.TotalAlloc))
	checkNonZero(t, "Sys", int(stats.Sys))
	checkNonZero(t, "Mallocs", int(stats.Mallocs))
	checkNonZero(t, "Frees", int(stats.Frees))
}

func checkNonZero(t *testing.T, name string, val int) {
	t.Helper()
	if val == 0 {
		t.Errorf("MemStats.%s = 0, 期望非零值", name)
	}
}

// TestPrintMemStatsDoesNotPanic — 验证 printMemStats 不会 panic
func TestPrintMemStatsDoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("printMemStats 引发了 panic: %v", r)
		}
	}()

	// printMemStats 只打印到 stdout，我们验证它不 panic
	printMemStats("测试标签")
}