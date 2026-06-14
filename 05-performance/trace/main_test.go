// trace 执行轨迹包单元测试
// 注意：不生成 trace 文件，仅测试工具函数逻辑
//
// busyWork 内部使用 rand.Intn()，结果具有随机性，
// 因此仅验证其不 panic 和结果非负，不验证具体返回值。
package main

import (
	"testing"
)

// TestBusyWorkDoesNotPanic 验证 busyWork 不会 panic
func TestBusyWorkDoesNotPanic(t *testing.T) {
	tests := []struct {
		name string
		n    int
	}{
		{"n=0", 0},
		{"n=1", 1},
		{"n=10", 10},
		{"n=100", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 只验证不 panic
			_ = busyWork(tt.n)
		})
	}
}

// TestBusyWorkNonNegative 验证繁忙计算结果非负
func TestBusyWorkNonNegative(t *testing.T) {
	for n := 0; n <= 200; n++ {
		result := busyWork(n)
		if result < 0 {
			t.Errorf("busyWork(%d) = %d, 结果不应为负", n, result)
		}
	}
}

// TestAllocateHeap 测试堆分配函数
func TestAllocateHeap(t *testing.T) {
	tests := []struct {
		name  string
		count int
	}{
		{"空切片", 0},
		{"单个元素", 1},
		{"少量元素", 10},
		{"中等数量", 100},
		{"较大数量", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := allocateHeap(tt.count)
			if len(result) != tt.count {
				t.Errorf("allocateHeap(%d) 返回长度 %d, 期望 %d", tt.count, len(result), tt.count)
			}
			// 验证每个元素非空
			for i, s := range result {
				if s == "" {
					t.Errorf("result[%d] 是空字符串", i)
				}
			}
		})
	}
}

// TestAllocateHeapContent 验证 allocateHeap 结果的格式
func TestAllocateHeapContent(t *testing.T) {
	result := allocateHeap(5)
	for i, s := range result {
		if len(s) == 0 {
			t.Errorf("result[%d] 内容为空", i)
		}
		// 应该以 "data-" 开头
		if len(s) < 5 || s[:5] != "data-" {
			t.Errorf("result[%d] = %q, 期望以 'data-' 开头", i, s)
		}
	}
}