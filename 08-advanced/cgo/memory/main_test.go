package main

import (
	"testing"
)

// ============ C.fill_ints 测试 ============

// TestCgoFillInts 测试 C 函数填充 Go 切片
func TestCgoFillInts(t *testing.T) {
	slice := []int32{1, 2, 3, 4, 5}
	cgoFillInts(slice, 99)
	expected := []int32{99, 99, 99, 99, 99}
	for i, v := range slice {
		if v != expected[i] {
			t.Errorf("slice[%d] = %d，期望 %d", i, v, expected[i])
		}
	}
}

// TestCgoFillInts_Zero 测试填充 0
func TestCgoFillInts_Zero(t *testing.T) {
	slice := []int32{1, 2, 3}
	cgoFillInts(slice, 0)
	for i, v := range slice {
		if v != 0 {
			t.Errorf("slice[%d] = %d，期望 0", i, v)
		}
	}
}

// TestCgoFillInts_Empty 测试空切片
func TestCgoFillInts_Empty(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("空切片 panic: %v", r)
		}
	}()
	cgoFillInts(nil, 99)
	var empty []int32
	cgoFillInts(empty, 99)
}

// TestCgoFillInts_Single 测试单个元素
func TestCgoFillInts_Single(t *testing.T) {
	slice := []int32{42}
	cgoFillInts(slice, 7)
	if slice[0] != 7 {
		t.Errorf("slice[0] = %d，期望 7", slice[0])
	}
}

// ============ C.make_array 测试 ============

// TestCgoMakeArray 测试在 C 堆上创建数组
func TestCgoMakeArray(t *testing.T) {
	result := cgoMakeArray(10, 42)
	if len(result) != 10 {
		t.Fatalf("结果长度应为 10，实际为 %d", len(result))
	}
	for i, v := range result {
		if v != 42 {
			t.Errorf("result[%d] = %d，期望 42", i, v)
		}
	}
}

// TestCgoMakeArray_Zero 测试创建全零数组
func TestCgoMakeArray_Zero(t *testing.T) {
	result := cgoMakeArray(5, 0)
	for i, v := range result {
		if v != 0 {
			t.Errorf("result[%d] = %d，期望 0", i, v)
		}
	}
}

// TestCgoMakeArray_Single 测试单元素数组
func TestCgoMakeArray_Single(t *testing.T) {
	result := cgoMakeArray(1, 999)
	if len(result) != 1 || result[0] != 999 {
		t.Errorf("结果应为 [999]，实际为 %v", result)
	}
}

// ============ C.memcpy 数据拷贝测试 ============

// TestCgoMemcpy 测试 Go → C 内存拷贝并读回
func TestCgoMemcpy(t *testing.T) {
	src := []int32{0, 10, 20, 30, 40, 50, 60, 70, 80, 90}
	result := cgoMemcpyFromGo(src)
	if len(result) != len(src) {
		t.Fatalf("结果长度 %d 与源 %d 不一致", len(result), len(src))
	}
	for i, v := range result {
		if v != src[i] {
			t.Errorf("result[%d] = %d，期望 %d", i, v, src[i])
		}
	}
}

// TestCgoMemcpy_Empty 测试空切片拷贝
func TestCgoMemcpy_Empty(t *testing.T) {
	result := cgoMemcpyFromGo(nil)
	if result == nil {
		t.Fatal("空切片拷贝结果应为非 nil 的空切片")
	}
	if len(result) != 0 {
		t.Errorf("空切片拷贝结果长度应为 0，实际为 %d", len(result))
	}
}

// ============ C 数组到 Go slice 读测试 ============

// TestCgoMatrix 测试 C 矩阵分配和读取
func TestCgoMatrix(t *testing.T) {
	result := cgoMatrixAlloc(3, 4, 7)
	if len(result) != 3 {
		t.Fatalf("行数应为 3，实际为 %d", len(result))
	}
	for i, row := range result {
		if len(row) != 4 {
			t.Fatalf("result[%d] 列数应为 4，实际为 %d", i, len(row))
		}
		for j, v := range row {
			if v != 7 {
				t.Errorf("result[%d][%d] = %d，期望 7", i, j, v)
			}
		}
	}
}

// TestCgoMatrix_Small 测试小矩阵
func TestCgoMatrix_Small(t *testing.T) {
	result := cgoMatrixAlloc(1, 1, 5)
	if result[0][0] != 5 {
		t.Errorf("单元素矩阵值应为 5，实际为 %d", result[0][0])
	}
}

// ============ 调用不 panic 验证 ============

// TestCPrintIntsNoPanic 测试 C.print_ints 不 panic
func TestCPrintIntsNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("C.print_ints panic: %v", r)
		}
	}()
	cgoFillInts([]int32{10, 20, 30}, 1)
}

// ============ 数据完整性测试 ============

// TestCgoDataIntegrity 测试 C 操作后 Go 数据的完整性
func TestCgoDataIntegrity(t *testing.T) {
	original := []int32{1, 2, 3, 4, 5}
	backup := make([]int32, len(original))
	copy(backup, original)

	// C 填充后检查
	cgoFillInts(original, 99)
	for i, v := range original {
		if v != 99 {
			t.Errorf("填充后 original[%d] = %d，期望 99", i, v)
		}
	}
	// 备份未被修改
	for i, v := range backup {
		if v != int32(i+1) {
			t.Errorf("备份被修改: backup[%d] = %d，期望 %d", i, v, i+1)
		}
	}
}

// ============ 多次 CGO 调用 ============

// TestRepeatedCGOCalls 测试连续 CGO 调用的稳定性
func TestRepeatedCGOCalls(t *testing.T) {
	for i := 0; i < 20; i++ {
		result := cgoMakeArray(5, i)
		for j, v := range result {
			if v != int32(i) {
				t.Errorf("第 %d 次: result[%d] = %d，期望 %d", i, j, v, i)
			}
		}
	}
}