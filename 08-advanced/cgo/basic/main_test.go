package main

import (
	"testing"
)

// ============ cgoAdd 函数测试 ============

// TestCgoAdd 测试整数相加
func TestCgoAdd(t *testing.T) {
	result := cgoAdd(3, 4)
	if result != 7 {
		t.Errorf("cgoAdd(3, 4) = %d，期望 7", result)
	}
}

// TestCgoAdd_Negative 测试负数相加
func TestCgoAdd_Negative(t *testing.T) {
	result := cgoAdd(-5, 10)
	if result != 5 {
		t.Errorf("cgoAdd(-5, 10) = %d，期望 5", result)
	}
}

// TestCgoAdd_Zero 测试零值相加
func TestCgoAdd_Zero(t *testing.T) {
	result := cgoAdd(0, 0)
	if result != 0 {
		t.Errorf("cgoAdd(0, 0) = %d，期望 0", result)
	}
}

// TestCgoAdd_Multiple 测试多次调用
func TestCgoAdd_Multiple(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{1, 2, 3},
		{10, 20, 30},
		{100, 200, 300},
		{-1, -2, -3},
		{0, 5, 5},
	}
	for _, tt := range tests {
		result := cgoAdd(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("cgoAdd(%d, %d) = %d，期望 %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

// ============ cgoStrLen 函数测试 ============

// TestCgoStrLen 测试字符串长度计算
func TestCgoStrLen(t *testing.T) {
	result := cgoStrLen("CGO")
	if result != 3 {
		t.Errorf("cgoStrLen(\"CGO\") = %d，期望 3", result)
	}
}

// TestCgoStrLen_Empty 测试空字符串长度
func TestCgoStrLen_Empty(t *testing.T) {
	result := cgoStrLen("")
	if result != 0 {
		t.Errorf("cgoStrLen(\"\") = %d，期望 0", result)
	}
}

// TestCgoStrLen_Chinese 测试中文字符的字节长度
func TestCgoStrLen_Chinese(t *testing.T) {
	result := cgoStrLen("你好世界")
	// "你好世界" 在 UTF-8 中每个汉字 3 字节，共 12 字节
	if result != 12 {
		t.Errorf("cgoStrLen(\"你好世界\") = %d，期望 12（UTF-8 字节数）", result)
	}
}

// TestCgoStrLen_Various 测试多种字符串
func TestCgoStrLen_Various(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"a", 1},
		{"hello", 5},
		{"hello world", 11},
		{"\n\t", 2},
		{"  spaces  ", 10},
	}
	for _, tt := range tests {
		result := cgoStrLen(tt.input)
		if result != tt.expected {
			t.Errorf("cgoStrLen(%q) = %d，期望 %d", tt.input, result, tt.expected)
		}
	}
}

// ============ cgoGreet 调用测试 ============

// TestCgoGreet 测试 C.greet 调用（仅验证不 panic）
func TestCgoGreet(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("cgoGreet 调用 panic: %v", r)
		}
	}()
	cgoGreet("Gopher")
}

// ============ 类型信息验证 ============

// TestCIntMapping 通过 cgo 验证 C.int 与 Go 类型的映射
func TestCIntMapping(t *testing.T) {
	// 验证 C 的 add 返回结果匹配 Go 类型预期
	a, b := 100, 200
	result := cgoAdd(a, b)
	if result != 300 {
		t.Errorf("C 整数运算结果不正确: %d", result)
	}
}

// ============ 多次调用测试 ============

// TestRepeatedCalls 测试连续多次 CGO 调用
func TestRepeatedCalls(t *testing.T) {
	for i := 0; i < 50; i++ {
		result := cgoAdd(i, i)
		if result != i*2 {
			t.Errorf("第 %d 次调用: cgoAdd(%d, %d) = %d，期望 %d", i, i, i, result, i*2)
		}
	}
}