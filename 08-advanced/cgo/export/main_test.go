package main

import (
	"testing"
)

// ============ doubleInts 测试 ============

// TestCgoDoubleInts 测试 C 翻倍函数
func TestCgoDoubleInts(t *testing.T) {
	result := cgoDoubleInts([]int{1, 2, 3, 4, 5})
	expected := []int{2, 4, 6, 8, 10}
	for i, v := range result {
		if v != expected[i] {
			t.Errorf("result[%d] = %d，期望 %d", i, v, expected[i])
		}
	}
}

// TestCgoDoubleInts_AllOnes 测试全 1 翻倍
func TestCgoDoubleInts_AllOnes(t *testing.T) {
	result := cgoDoubleInts([]int{1, 1, 1})
	expected := []int{2, 2, 2}
	for i, v := range result {
		if v != expected[i] {
			t.Errorf("result[%d] = %d，期望 %d", i, v, expected[i])
		}
	}
}

// TestCgoDoubleInts_Empty 测试空切片
func TestCgoDoubleInts_Empty(t *testing.T) {
	result := cgoDoubleInts(nil)
	if len(result) > 0 {
		t.Errorf("空切片应返回空结果，实际为 %v", result)
	}
	result2 := cgoDoubleInts([]int{})
	if len(result2) > 0 {
		t.Errorf("空切片应返回空结果，实际为 %v", result2)
	}
}

// TestCgoDoubleInts_Zero 测试全 0 翻倍
func TestCgoDoubleInts_Zero(t *testing.T) {
	result := cgoDoubleInts([]int{0, 0, 0})
	expected := []int{0, 0, 0}
	for i, v := range result {
		if v != expected[i] {
			t.Errorf("result[%d] = %d，期望 %d", i, v, expected[i])
		}
	}
}

// TestCgoDoubleInts_Negative 测试负数翻倍
func TestCgoDoubleInts_Negative(t *testing.T) {
	result := cgoDoubleInts([]int{-1, -2, -3})
	expected := []int{-2, -4, -6}
	for i, v := range result {
		if v != expected[i] {
			t.Errorf("result[%d] = %d，期望 %d", i, v, expected[i])
		}
	}
}

// ============ sumInts 测试 ============

// TestCgoSumInts 测试 C 求和函数
func TestCgoSumInts(t *testing.T) {
	result := cgoSumInts([]int{2, 4, 6, 8, 10})
	if result != 30 {
		t.Errorf("cgoSumInts = %d，期望 30", result)
	}
}

// TestCgoSumInts_Single 测试单元素
func TestCgoSumInts_Single(t *testing.T) {
	result := cgoSumInts([]int{99})
	if result != 99 {
		t.Errorf("cgoSumInts([99]) = %d，期望 99", result)
	}
}

// TestCgoSumInts_Empty 测试空切片
func TestCgoSumInts_Empty(t *testing.T) {
	result := cgoSumInts(nil)
	if result != 0 {
		t.Errorf("空切片求和应为 0，实际为 %d", result)
	}
}

// TestCgoSumInts_Negative 测试含负数
func TestCgoSumInts_Negative(t *testing.T) {
	result := cgoSumInts([]int{-10, 20, -5})
	if result != 5 {
		t.Errorf("cgoSumInts([-10,20,-5]) = %d，期望 5", result)
	}
}

// ============ //export goDouble 测试 ============

// TestCgoDoubleGoValue 测试 goDouble（//export）函数
func TestCgoDoubleGoValue(t *testing.T) {
	result := cgoDoubleGoValue(21)
	if result != 42 {
		t.Errorf("goDouble(21) = %d，期望 42", result)
	}
}

// TestCgoDoubleGoValue_Zero 测试 goDouble(0)
func TestCgoDoubleGoValue_Zero(t *testing.T) {
	result := cgoDoubleGoValue(0)
	if result != 0 {
		t.Errorf("goDouble(0) = %d，期望 0", result)
	}
}

// TestCgoDoubleGoValue_Negative 测试 goDouble 负数
func TestCgoDoubleGoValue_Negative(t *testing.T) {
	result := cgoDoubleGoValue(-10)
	if result != -20 {
		t.Errorf("goDouble(-10) = %d，期望 -20", result)
	}
}

// ============ //export goSumSlice 测试 ============

// TestCgoSumGoSlice 测试 goSumSlice（//export）函数
func TestCgoSumGoSlice(t *testing.T) {
	result := cgoSumGoSlice([]int{1, 2, 3, 4, 5})
	if result != 15 {
		t.Errorf("goSumSlice([1,2,3,4,5]) = %d，期望 15", result)
	}
}

// TestCgoSumGoSlice_Empty 测试空 slice
func TestCgoSumGoSlice_Empty(t *testing.T) {
	result := cgoSumGoSlice(nil)
	if result != 0 {
		t.Errorf("空 slice 求和 = %d，期望 0", result)
	}
}

// TestCgoSumGoSlice_Single 测试单元素
func TestCgoSumGoSlice_Single(t *testing.T) {
	result := cgoSumGoSlice([]int{42})
	if result != 42 {
		t.Errorf("goSumSlice([42]) = %d，期望 42", result)
	}
}

// TestCgoSumGoSlice_Negative 测试含负数
func TestCgoSumGoSlice_Negative(t *testing.T) {
	result := cgoSumGoSlice([]int{-5, 10, -3, 8})
	if result != 10 {
		t.Errorf("goSumSlice([-5,10,-3,8]) = %d，期望 10", result)
	}
}

// TestCgoSumGoSlice_Large 测试大数值
func TestCgoSumGoSlice_Large(t *testing.T) {
	result := cgoSumGoSlice([]int{1000000, 2000000, 3000000})
	if result != 6000000 {
		t.Errorf("goSumSlice([1M,2M,3M]) = %d，期望 6000000", result)
	}
}

// ============ allocPeople + printPeople 测试 ============

// TestCgoAllocPeople 测试 C 分配 Person 数组
func TestCgoAllocPeople(t *testing.T) {
	names := []string{"Alice", "Bob", "Charlie"}
	ages := []int{30, 25, 35}
	scores := []float64{95.5, 88.0, 92.3}
	ok := cgoAllocPeople(names, ages, scores)
	if !ok {
		t.Error("Person 数组数据验证失败")
	}
}

// ============ 链式操作测试 ============

// TestCgoChainedOperation 测试 doubleInts + sumInts 链式操作
func TestCgoChainedOperation(t *testing.T) {
	result := cgoChainedOperation([]int{1, 2, 3, 4, 5})
	// 先翻倍: [2,4,6,8,10], 再求和: 30
	if result != 30 {
		t.Errorf("链式操作结果 = %d，期望 30", result)
	}
}

// TestCgoChainedOperation_Empty 测试空切片链式操作
func TestCgoChainedOperation_Empty(t *testing.T) {
	result := cgoChainedOperation(nil)
	if result != 0 {
		t.Errorf("空切片链式操作结果应为 0，实际为 %d", result)
	}
}

// ============ 多值验证 ============

// TestCgoDoubleAndSum_Table 表驱动测试翻倍后求和
func TestCgoDoubleAndSum_Table(t *testing.T) {
	tests := []struct {
		input    []int
		expected int64
	}{
		{[]int{1, 2, 3}, 12},       // [2,4,6] = 12
		{[]int{0, 0, 0}, 0},        // [0,0,0] = 0
		{[]int{-1, 1}, 0},          // [-2,2] = 0
		{[]int{10, 20, 30}, 120},   // [20,40,60] = 120
	}
	for _, tt := range tests {
		result := cgoChainedOperation(tt.input)
		if result != tt.expected {
			t.Errorf("chained(%v) = %d，期望 %d", tt.input, result, tt.expected)
		}
	}
}

// ============ C.printPeople 调用（不 panic） ============

// TestCPrintPeopleNoPanic 测试 C.printPeople 不 panic
func TestCPrintPeopleNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("C.printPeople panic: %v", r)
		}
	}()
	names := []string{"Alice", "Bob"}
	ages := []int{30, 25}
	scores := []float64{95.5, 88.0}
	cgoAllocPeople(names, ages, scores)
}