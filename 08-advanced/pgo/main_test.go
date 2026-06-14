package main

import (
	"math/rand"
	"testing"
)

// ============ generateDataset 测试 ============

// TestGenerateDataset 测试数据集生成
func TestGenerateDataset(t *testing.T) {
	size := 100
	data := generateDataset(size)
	if len(data) != size {
		t.Errorf("数据集长度应为 %d，实际为 %d", size, len(data))
	}
	// 验证数据都是非负整数（rand.Intn 返回 [0,n)）
	for i, v := range data {
		if v < 0 || v >= 1000000 {
			t.Errorf("data[%d] = %d 超出预期范围 [0, 1000000)", i, v)
		}
	}
}

// TestGenerateDataset_Empty 测试生成空数据集
func TestGenerateDataset_Empty(t *testing.T) {
	data := generateDataset(0)
	if data == nil {
		t.Fatal("空数据集不应返回 nil")
	}
	if len(data) != 0 {
		t.Errorf("空数据集长度应为 0，实际为 %d", len(data))
	}
}

// TestGenerateDataset_DifferentSeeds 验证不同种子产生不同数据（测试随机性）
func TestGenerateDataset_DifferentSeeds(t *testing.T) {
	rand.Seed(42)
	data1 := generateDataset(50)

	rand.Seed(99)
	data2 := generateDataset(50)

	if len(data1) != len(data2) {
		t.Fatalf("长度不一致: %d vs %d", len(data1), len(data2))
	}
	// 不同种子不太可能产生完全相同的数据
	same := true
	for i := range data1 {
		if data1[i] != data2[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("不同种子产生了相同的数据，预期随机性")
	}
}

// ============ processDatasetHotPath 测试 ============

// TestProcessDatasetHotPath 测试排序和统计功能
func TestProcessDatasetHotPath(t *testing.T) {
	data := []int{5, 3, 1, 4, 2}
	sorted, sum, avg, median := processDatasetHotPath(data)

	// 验证排序结果
	expectedSorted := []int{1, 2, 3, 4, 5}
	if len(sorted) != len(expectedSorted) {
		t.Fatalf("排序后长度应为 %d，实际为 %d", len(expectedSorted), len(sorted))
	}
	for i, v := range sorted {
		if v != expectedSorted[i] {
			t.Errorf("sorted[%d] = %d，期望 %d", i, v, expectedSorted[i])
		}
	}

	// 验证总和
	if sum != 15 {
		t.Errorf("总和应为 15，实际为 %d", sum)
	}

	// 验证平均值
	if avg != 3.0 {
		t.Errorf("平均值应为 3.0，实际为 %f", avg)
	}

	// 验证中位数（单数序列）
	if median != 3 {
		t.Errorf("中位数应为 3，实际为 %d", median)
	}

	// 验证原始数据未被修改（processDatasetHotPath 应复制一份）
	original := []int{5, 3, 1, 4, 2}
	for i, v := range data {
		if v != original[i] {
			t.Errorf("原始数据被修改: data[%d] = %d，期望 %d", i, v, original[i])
		}
	}
}

// TestProcessDatasetHotPath_EvenLength 测试偶数长度的中位数
func TestProcessDatasetHotPath_EvenLength(t *testing.T) {
	data := []int{4, 1, 3, 2} // 排序后 [1,2,3,4]，中位数 (2+3)/2=2
	_, _, _, median := processDatasetHotPath(data)
	if median != 2 {
		t.Errorf("偶数序列中位数应为 2，实际为 %d", median)
	}
}

// TestProcessDatasetHotPath_SingleElement 测试单个元素
func TestProcessDatasetHotPath_SingleElement(t *testing.T) {
	data := []int{42}
	sorted, sum, avg, median := processDatasetHotPath(data)

	if len(sorted) != 1 || sorted[0] != 42 {
		t.Errorf("排序结果错误: %v", sorted)
	}
	if sum != 42 {
		t.Errorf("总和应为 42，实际为 %d", sum)
	}
	if avg != 42.0 {
		t.Errorf("平均值应为 42.0，实际为 %f", avg)
	}
	if median != 42 {
		t.Errorf("中位数应为 42，实际为 %d", median)
	}
}

// TestProcessDatasetHotPath_Duplicates 测试重复元素
func TestProcessDatasetHotPath_Duplicates(t *testing.T) {
	data := []int{5, 5, 5, 1, 1}
	sorted, sum, avg, median := processDatasetHotPath(data)

	if len(sorted) != 5 {
		t.Fatalf("排序后长度应为 5，实际为 %d", len(sorted))
	}
	// 排序后: [1,1,5,5,5]
	if sorted[0] != 1 || sorted[4] != 5 {
		t.Errorf("排序结果不正确: %v", sorted)
	}
	if sum != 17 {
		t.Errorf("总和应为 17，实际为 %d", sum)
	}
	if avg != 3.4 {
		t.Errorf("平均值应为 3.4，实际为 %f", avg)
	}
	if median != 5 {
		t.Errorf("中位数应为 5，实际为 %d", median)
	}
}

// TestProcessDatasetHotPath_NegativeNotApplicable 生成数据为非负整数，测试正常处理
func TestProcessDatasetHotPath_LargeDataset(t *testing.T) {
	data := generateDataset(1000)
	sorted, sum, avg, median := processDatasetHotPath(data)

	if len(sorted) != 1000 {
		t.Errorf("排序后长度应为 1000，实际为 %d", len(sorted))
	}
	// 检查排序正确性
	for i := 1; i < len(sorted); i++ {
		if sorted[i-1] > sorted[i] {
			t.Errorf("排序不正确: sorted[%d]=%d > sorted[%d]=%d", i-1, sorted[i-1], i, sorted[i])
			break
		}
	}
	// 检查总和为正
	if sum <= 0 {
		t.Errorf("总和应 > 0，实际为 %d", sum)
	}
	// 平均值应在合理范围
	if avg < 0 || avg >= 1000000 {
		t.Errorf("平均值 %f 超出合理范围", avg)
	}
	// 中位数应在合理范围
	if median < 0 || median >= 1000000 {
		t.Errorf("中位数 %d 超出合理范围", median)
	}
}

// ============ textProcessor 测试 ============

// TestNewTextProcessor 测试文本处理器的创建
func TestNewTextProcessor(t *testing.T) {
	tp := newTextProcessor()
	if tp == nil {
		t.Fatal("newTextProcessor() 返回 nil")
	}
	if tp.buffer == nil {
		t.Fatal("处理器 buffer 为 nil")
	}
	if len(tp.buffer) != 0 {
		t.Errorf("新处理器 buffer 长度应为 0，实际为 %d", len(tp.buffer))
	}
}

// TestTextProcessor_Ingest 测试文本摄入
func TestTextProcessor_Ingest(t *testing.T) {
	tp := newTextProcessor()
	lines := []string{"hello world", "foo bar", "baz qux"}
	tp.ingest(lines)
	if len(tp.buffer) != 3 {
		t.Errorf("摄入后 buffer 长度应为 3，实际为 %d", len(tp.buffer))
	}
}

// TestTextProcessor_Analyze 测试文本分析
func TestTextProcessor_Analyze(t *testing.T) {
	tp := newTextProcessor()
	tp.ingest([]string{
		"apple banana",
		"banana cherry",
		"cherry date",
	})

	result := tp.analyze()

	// 统计首字母：'a' -> 1, 'b' -> 1, 'c' -> 1
	if result["a"] != 1 {
		t.Errorf(`结果中 'a' 应为 1，实际为 %d`, result["a"])
	}
	if result["b"] != 1 {
		t.Errorf(`结果中 'b' 应为 1，实际为 %d`, result["b"])
	}
	if result["c"] != 1 {
		t.Errorf(`结果中 'c' 应为 1，实际为 %d`, result["c"])
	}
}

// TestTextProcessor_Analyze_EmptyBuffer 测试空 buffer 的分析
func TestTextProcessor_Analyze_EmptyBuffer(t *testing.T) {
	tp := newTextProcessor()
	result := tp.analyze()
	if len(result) != 0 {
		t.Errorf("空 buffer 分析结果应为空，实际为 %v", result)
	}
}

// TestTextProcessor_Analyze_LowercaseOnly 测试只统计小写字母开头的行
// analyze 函数在行内寻找第一个 'a'-'z' 字符并统计其出现次数
func TestTextProcessor_Analyze_LowercaseOnly(t *testing.T) {
	tp := newTextProcessor()
	tp.ingest([]string{
		"uppercase",
	})
	result := tp.analyze()
	// 'u' 是小写字母，应被统计
	if result["u"] != 1 {
		t.Errorf("'u' 应为 1，实际: %v", result)
	}
}

// TestTextProcessor_Analyze_SkipNoLowercase 测试没有小写字母的行应被忽略
func TestTextProcessor_Analyze_SkipNoLowercase(t *testing.T) {
	tp := newTextProcessor()
	tp.ingest([]string{
		"12345!@#$%",
		"UPPERCASE", // 大写开头，循环查找直到遇到大写字母，不会有 'a'-'z' 匹配
	})
	result := tp.analyze()
	if len(result) != 0 {
		t.Errorf("无小写字母的行不应被统计，实际: %v", result)
	}
}

// ============ DataWorkload 测试 ============

// TestNewDataWorkload 测试工作负载创建
func TestNewDataWorkload(t *testing.T) {
	dw := NewDataWorkload(5, 100)
	if dw == nil {
		t.Fatal("NewDataWorkload() 返回 nil")
	}
	if len(dw.datasets) != 5 {
		t.Errorf("数据集合数应为 5，实际为 %d", len(dw.datasets))
	}
	for i, data := range dw.datasets {
		if len(data) != 100 {
			t.Errorf("datasets[%d] 长度应为 100，实际为 %d", i, len(data))
		}
	}
}

// TestRunWorkload 测试工作负载运行
func TestRunWorkload(t *testing.T) {
	dw := NewDataWorkload(3, 50)
	totalSum := dw.RunWorkload()
	if totalSum <= 0 {
		t.Errorf("总和尚应 > 0，实际为 %d", totalSum)
	}
}

// ============ Benchmark 测试（用于 PGO profiling）============

// BenchmarkProcessDatasetHotPath 用于 PGO 性能分析
func BenchmarkProcessDatasetHotPath(b *testing.B) {
	data := generateDataset(100000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processDatasetHotPath(data)
	}
}

// BenchmarkTextProcessorAnalyze 用于 PGO 性能分析
func BenchmarkTextProcessorAnalyze(b *testing.B) {
	tp := newTextProcessor()
	lines := []string{
		"apple banana cherry",
		"date elderberry fig",
		"grape honeydew iilama",
		"jackfruit kiwi lemon",
	}
	tp.ingest(lines)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tp.analyze()
	}
}

// BenchmarkRunWorkload 用于 PGO 性能分析
func BenchmarkRunWorkload(b *testing.B) {
	dw := NewDataWorkload(5, 50000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dw.RunWorkload()
	}
}

// BenchmarkGenerateDataset 用于 PGO 性能分析
func BenchmarkGenerateDataset(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generateDataset(100000)
	}
}