// 基准测试与性能对比示例
// 运行方式:
//
//	cd 05-performance/benchmark
//	go test -bench=. -benchmem -count=5
//
// -benchmem 会显示每次操作的字节分配和分配次数
// -count=5  运行 5 次以消除冷启动偏差
package main

import (
	"strings"
	"testing"
)

// ============================================================================
// BenchmarkSliceAppend: append 动态扩容 vs 预分配容量
// ============================================================================

// BenchmarkSliceAppendNoPrealloc 不预分配，让 Go 运行时自动扩容
func BenchmarkSliceAppendNoPrealloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var s []int
		for j := 0; j < 10000; j++ {
			s = append(s, j)
		}
	}
}

// BenchmarkSliceAppendPrealloc 预先分配好容量，避免扩容时的拷贝
func BenchmarkSliceAppendPrealloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := make([]int, 0, 10000)
		for j := 0; j < 10000; j++ {
			s = append(s, j)
		}
	}
}

// ============================================================================
// BenchmarkStringConcat: 字符串拼接方式对比
// ============================================================================

const concatCount = 1000

// BenchmarkStringConcatOperator 使用 + 运算符拼接（每次产生新字符串）
func BenchmarkStringConcatOperator(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var s string
		for j := 0; j < concatCount; j++ {
			s += "hello"
		}
		_ = s
	}
}

// BenchmarkStringConcatJoin 使用 strings.Join 拼接
func BenchmarkStringConcatJoin(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parts := make([]string, concatCount)
		for j := 0; j < concatCount; j++ {
			parts[j] = "hello"
		}
		_ = strings.Join(parts, "")
	}
}

// BenchmarkStringConcatBuilder 使用 strings.Builder 拼接（最佳实践）
func BenchmarkStringConcatBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var sb strings.Builder
		sb.Grow(concatCount * len("hello")) // 预分配避免扩容
		for j := 0; j < concatCount; j++ {
			sb.WriteString("hello")
		}
		_ = sb.String()
	}
}

// ============================================================================
// BenchmarkMapLookup: 不同 map 大小下的查找性能
// ============================================================================

func buildMap(n int) map[int]int {
	m := make(map[int]int, n)
	for i := 0; i < n; i++ {
		m[i] = i * 2
	}
	return m
}

// BenchmarkMapLookupSmall 小 map（100 元素）查找
func BenchmarkMapLookupSmall(b *testing.B) {
	m := buildMap(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m[i%100]
	}
}

// BenchmarkMapLookupMedium 中等 map（10000 元素）查找
func BenchmarkMapLookupMedium(b *testing.B) {
	m := buildMap(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m[i%10000]
	}
}

// BenchmarkMapLookupLarge 大 map（100000 元素）查找
func BenchmarkMapLookupLarge(b *testing.B) {
	m := buildMap(100000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m[i%100000]
	}
}

// BenchmarkMapLookupMissing 查找不存在的 key（最坏情况）
func BenchmarkMapLookupMissing(b *testing.B) {
	m := buildMap(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, ok := m[999999]
		_ = ok
	}
}

// ============================================================================
// Test 函数：验证基准测试所用的基础操作是正确的
// ============================================================================

// TestSliceAppend 验证 append 与预分配结果一致
func TestSliceAppend(t *testing.T) {
	// append 方式
	var s1 []int
	for j := 0; j < 100; j++ {
		s1 = append(s1, j)
	}
	// 预分配方式
	s2 := make([]int, 0, 100)
	for j := 0; j < 100; j++ {
		s2 = append(s2, j)
	}
	if len(s1) != len(s2) {
		t.Fatalf("长度不一致: %d vs %d", len(s1), len(s2))
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			t.Fatalf("元素 %d 不一致: %d vs %d", i, s1[i], s2[i])
		}
	}
}

// TestStringConcat 验证三种拼接方式结果一致
func TestStringConcat(t *testing.T) {
	n := 100
	want := strings.Repeat("hello", n)

	// 运算符拼接
	var s1 string
	for j := 0; j < n; j++ {
		s1 += "hello"
	}
	if s1 != want {
		t.Fatalf("运算符拼接结果错误: got %q", s1)
	}

	// strings.Join
	parts := make([]string, n)
	for j := 0; j < n; j++ {
		parts[j] = "hello"
	}
	s2 := strings.Join(parts, "")
	if s2 != want {
		t.Fatalf("strings.Join 结果错误: got %q", s2)
	}

	// strings.Builder
	var sb strings.Builder
	sb.Grow(n * len("hello"))
	for j := 0; j < n; j++ {
		sb.WriteString("hello")
	}
	s3 := sb.String()
	if s3 != want {
		t.Fatalf("strings.Builder 结果错误: got %q", s3)
	}
}

// TestMapLookup 验证 map 查找函数正确
func TestMapLookup(t *testing.T) {
	m := buildMap(100)
	for i := 0; i < 100; i++ {
		v, ok := m[i]
		if !ok {
			t.Fatalf("key %d 不存在", i)
		}
		if v != i*2 {
			t.Fatalf("key %d 的值应为 %d, 实际为 %d", i, i*2, v)
		}
	}
	// 不存在的 key
	_, ok := m[999]
	if ok {
		t.Fatal("不存在的 key 不应返回 ok=true")
	}
}