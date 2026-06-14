package main

import (
	"testing"
)

// ============ C qsort 排序测试 ============

// TestCgoSortInts 测试 C qsort 排序
func TestCgoSortInts(t *testing.T) {
	input := []int{9, 3, 7, 1, 5, 8, 2, 6, 4, 0}
	result := cgoSortInts(input)
	for i := 0; i < len(result); i++ {
		if result[i] != i {
			t.Errorf("result[%d] = %d，期望 %d", i, result[i], i)
		}
	}
}

// TestCgoSortInts_Sorted 测试已排序数组
func TestCgoSortInts_Sorted(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	result := cgoSortInts(input)
	for i, v := range result {
		if v != i+1 {
			t.Errorf("result[%d] = %d，期望 %d", i, v, i+1)
		}
	}
}

// TestCgoSortInts_Empty 测试空数组
func TestCgoSortInts_Empty(t *testing.T) {
	result := cgoSortInts(nil)
	if len(result) != 0 {
		t.Errorf("空数组结果长度应为 0，实际为 %d", len(result))
	}
}

// TestCgoSortInts_Single 测试单元素
func TestCgoSortInts_Single(t *testing.T) {
	result := cgoSortInts([]int{42})
	if len(result) != 1 || result[0] != 42 {
		t.Errorf("单元素排序结果应为 [42]，实际为 %v", result)
	}
}

// TestCgoSortInts_Duplicates 测试重复元素
func TestCgoSortInts_Duplicates(t *testing.T) {
	result := cgoSortInts([]int{5, 1, 3, 1, 3})
	expected := []int{1, 1, 3, 3, 5}
	for i, v := range result {
		if v != expected[i] {
			t.Errorf("result[%d] = %d，期望 %d", i, v, expected[i])
		}
	}
}

// TestCgoSortInts_Reverse 测试逆序数组
func TestCgoSortInts_Reverse(t *testing.T) {
	result := cgoSortInts([]int{5, 4, 3, 2, 1})
	for i := 0; i < 5; i++ {
		if result[i] != i+1 {
			t.Errorf("result[%d] = %d，期望 %d", i, result[i], i+1)
		}
	}
}

// TestCgoSortInts_Negative 测试含负数
func TestCgoSortInts_Negative(t *testing.T) {
	result := cgoSortInts([]int{-3, 5, 0, -1, 2})
	expected := []int{-3, -1, 0, 2, 5}
	for i, v := range result {
		if v != expected[i] {
			t.Errorf("result[%d] = %d，期望 %d", i, v, expected[i])
		}
	}
}

// ============ C.degrees_to_radians 测试 ============

// TestCgoDegreesToRadians 测试角度转弧度
func TestCgoDegreesToRadians(t *testing.T) {
	rad := cgoDegreesToRadians(180.0)
	if rad < 3.14 || rad > 3.15 {
		t.Errorf("degrees_to_radians(180) = %f，期望约 3.14159", rad)
	}
}

// TestCgoDegreesToRadians_Zero 测试 0 度
func TestCgoDegreesToRadians_Zero(t *testing.T) {
	rad := cgoDegreesToRadians(0.0)
	if rad != 0.0 {
		t.Errorf("degrees_to_radians(0) = %f，期望 0", rad)
	}
}

// TestCgoDegreesToRadians_90 测试 90 度
func TestCgoDegreesToRadians_90(t *testing.T) {
	rad := cgoDegreesToRadians(90.0)
	if rad < 1.57 || rad > 1.58 {
		t.Errorf("degrees_to_radians(90) = %f，期望约 1.5708", rad)
	}
}

// TestCgoDegreesToRadians_Table 表驱动测试角度转换
func TestCgoDegreesToRadians_Table(t *testing.T) {
	tests := []struct {
		deg      float64
		min, max float64
	}{
		{0, 0, 0},
		{30, 0.523, 0.524},
		{45, 0.785, 0.786},
		{60, 1.047, 1.048},
		{90, 1.570, 1.571},
		{180, 3.141, 3.142},
		{360, 6.283, 6.284},
	}
	for _, tt := range tests {
		rad := cgoDegreesToRadians(tt.deg)
		if rad < tt.min || rad > tt.max {
			t.Errorf("degrees_to_radians(%f) = %f，不在 [%f, %f] 范围内",
				tt.deg, rad, tt.min, tt.max)
		}
	}
}

// ============ C.Point 结构体测试 ============

// TestCgoPoint 测试 C.Point 结构体的创建和访问
func TestCgoPoint(t *testing.T) {
	x := cgoPointX(10, 20)
	y := cgoPointY(10, 20)
	if x != 10 {
		t.Errorf("Point.x = %d，期望 10", x)
	}
	if y != 20 {
		t.Errorf("Point.y = %d，期望 20", y)
	}
}

// TestCgoPoint_Zero 测试零值点
func TestCgoPoint_Zero(t *testing.T) {
	x := cgoPointX(0, 0)
	y := cgoPointY(0, 0)
	if x != 0 || y != 0 {
		t.Errorf("零值点应为 (0,0)，实际为 (%d,%d)", x, y)
	}
}

// ============ 嵌套结构体 Rect 测试 ============

// TestCgoRect 测试嵌套结构体
func TestCgoRect(t *testing.T) {
	x1, y1, x2, y2, color := cgoRectInfo(0, 0, 100, 200, 0)
	if x1 != 0 || y1 != 0 {
		t.Errorf("topLeft 应为 (0,0)，实际为 (%d,%d)", x1, y1)
	}
	if x2 != 100 || y2 != 200 {
		t.Errorf("bottomRight 应为 (100,200)，实际为 (%d,%d)", x2, y2)
	}
	if color != 0 { // RED = 0
		t.Errorf("color 应为 0 (RED)，实际为 %d", color)
	}
}

// ============ 枚举测试 ============

// TestCgoEnum 测试枚举值
func TestCgoEnum(t *testing.T) {
	if cgoEnumName(0) != "RED" {
		t.Errorf("枚举 0 应为 RED，实际为 %s", cgoEnumName(0))
	}
	if cgoEnumName(1) != "GREEN" {
		t.Errorf("枚举 1 应为 GREEN，实际为 %s", cgoEnumName(1))
	}
	if cgoEnumName(2) != "BLUE" {
		t.Errorf("枚举 2 应为 BLUE，实际为 %s", cgoEnumName(2))
	}
	if cgoEnumName(3) != "YELLOW" {
		t.Errorf("枚举 3 应为 YELLOW，实际为 %s", cgoEnumName(3))
	}
}

// ============ 数学库可用性测试 ============

// TestCgoMathLib 验证数学库链接正确
func TestCgoMathLib(t *testing.T) {
	if !cgoVerifyMathLib() {
		t.Error("数学库验证失败：degrees_to_radians(180) 结果不在 π 附近")
	}
}

// ============ 排序 + 数学组合测试 ============

// TestCgoSortAndMath 测试排序和数学函数组合调用
func TestCgoSortAndMath(t *testing.T) {
	// 先排序
	sorted := cgoSortInts([]int{3, 1, 2})
	if sorted[0] != 1 || sorted[1] != 2 || sorted[2] != 3 {
		t.Fatalf("排序结果不正确: %v", sorted)
	}
	// 再调用数学函数
	rad := cgoDegreesToRadians(180.0)
	if rad < 3.14 || rad > 3.15 {
		t.Errorf("数学函数结果不正确: %f", rad)
	}
}