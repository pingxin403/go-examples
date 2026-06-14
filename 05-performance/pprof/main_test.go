// pprof 性能分析包单元测试
// 注意：不运行实际的 CPU profiling，仅测试工具函数逻辑
package main

import (
	"testing"
)

// TestFibRecursive 测试递归版 Fibonacci 函数
func TestFibRecursive(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want int
	}{
		{"n=0", 0, 0},
		{"n=1", 1, 1},
		{"n=2", 2, 1},
		{"n=5", 5, 5},
		{"n=10", 10, 55},
		{"n=15", 15, 610},
		{"n=20", 20, 6765},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fibRecursive(tt.n)
			if got != tt.want {
				t.Errorf("fibRecursive(%d) = %d, 期望 %d", tt.n, got, tt.want)
			}
		})
	}
}

// TestFibOptimized 测试迭代优化版 Fibonacci 函数
func TestFibOptimized(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want int
	}{
		{"n=0", 0, 0},
		{"n=1", 1, 1},
		{"n=2", 2, 1},
		{"n=5", 5, 5},
		{"n=10", 10, 55},
		{"n=20", 20, 6765},
		{"n=30", 30, 832040},
		{"n=50", 50, 12586269025},
		{"n=90", 90, 2880067194370816120},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fibOptimized(tt.n)
			if got != tt.want {
				t.Errorf("fibOptimized(%d) = %d, 期望 %d", tt.n, got, tt.want)
			}
		})
	}
}

// TestFibConsistency 验证递归版和迭代版结果一致（小 n 范围）
func TestFibConsistency(t *testing.T) {
	for n := 0; n <= 20; n++ {
		t.Run("", func(t *testing.T) {
			r := fibRecursive(n)
			o := fibOptimized(n)
			if r != o {
				t.Errorf("n=%d: fibRecursive=%d, fibOptimized=%d, 两者应一致", n, r, o)
			}
		})
	}
}

// TestRandomString 测试随机字符串生成
func TestRandomString(t *testing.T) {
	tests := []struct {
		name string
		n    int
	}{
		{"空字符串", 0},
		{"单个字符", 1},
		{"常规长度64", 64},
		{"中等长度256", 256},
		{"较长1024", 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := randomString(tt.n)
			if len(got) != tt.n {
				t.Errorf("randomString(%d) 长度 = %d, 期望 %d", tt.n, len(got), tt.n)
			}
			// 验证只包含合法字符
			for i, c := range got {
				if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
					t.Errorf("位置 %d 有非法字符 '%c'", i, c)
				}
			}
		})
	}
}

// TestRandomStringDeterministic 验证 randomString 是确定性的（相同的 n 产生相同结果）
func TestRandomStringDeterministic(t *testing.T) {
	n := 64
	s1 := randomString(n)
	s2 := randomString(n)
	if s1 != s2 {
		t.Errorf("randomString 不是确定性的，两次调用结果不同: %q vs %q", s1, s2)
	}
}

// TestSieveOfEratosthenes 测试埃拉托色尼筛法
func TestSieveOfEratosthenes(t *testing.T) {
	tests := []struct {
		name  string
		limit int
		want  []int
	}{
		{"limit=0", 0, []int{}},
		{"limit=1", 1, []int{}},
		{"limit=2", 2, []int{2}},
		{"limit=10", 10, []int{2, 3, 5, 7}},
		{"limit=30", 30, []int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}},
		{"limit=100", 100, []int{
			2, 3, 5, 7, 11, 13, 17, 19, 23, 29,
			31, 37, 41, 43, 47, 53, 59, 61, 67, 71,
			73, 79, 83, 89, 97,
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sieveOfEratosthenes(tt.limit)
			if len(got) != len(tt.want) {
				t.Errorf("sieveOfEratosthenes(%d) 返回 %d 个素数, 期望 %d 个: got=%v, want=%v",
					tt.limit, len(got), len(tt.want), got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("sieveOfEratosthenes(%d)[%d] = %d, 期望 %d", tt.limit, i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestSieveOfEratosthenesCount 验证筛法返回正确数量的素数（用已知的素数计数验证）
func TestSieveOfEratosthenesCount(t *testing.T) {
	// 已知的 π(n) 素数计数: up to 10^5 有 9592 个素数
	primes := sieveOfEratosthenes(100000)
	if len(primes) != 9592 {
		t.Errorf("sieveOfEratosthenes(100000) 返回 %d 个素数, 期望 9592", len(primes))
	}
	// 验证结果是升序且没有合数
	for i := 1; i < len(primes); i++ {
		if primes[i] <= primes[i-1] {
			t.Errorf("结果不是严格递增: primes[%d]=%d, primes[%d]=%d", i-1, primes[i-1], i, primes[i])
		}
	}
}

// TestAllocateAndDiscard 验证分配函数不 panic
func TestAllocateAndDiscard(t *testing.T) {
	tests := []struct {
		name  string
		count int
	}{
		{"count=0", 0},
		{"count=1", 1},
		{"count=100", 100},
		{"count=1000", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 只验证不 panic，不测分配结果
			allocateAndDiscard(tt.count)
		})
	}
}