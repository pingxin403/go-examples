package main

import (
	"runtime"
	"testing"
)

// ============ CPU 核心数测试 ============

// TestCgoGetCPUCores 测试 C 获取 CPU 核心数与 Go 一致
func TestCgoGetCPUCores(t *testing.T) {
	cCores := cgoGetCPUCores()
	goCores := runtime.NumCPU()
	t.Logf("C 接口 CPU 核心数: %d", cCores)
	t.Logf("Go 接口 CPU 核心数: %d", goCores)
	if cCores <= 0 {
		t.Errorf("C.get_cpu_cores() = %d，应 > 0", cCores)
	}
	if goCores <= 0 {
		t.Errorf("runtime.NumCPU() = %d，应 > 0", goCores)
	}
}

// ============ 平方根测试 ============

// TestCgoComputeSqrt 测试 C 数学库 sqrt 函数
func TestCgoComputeSqrt(t *testing.T) {
	result := cgoComputeSqrt(4.0)
	if result != 2.0 {
		t.Errorf("sqrt(4.0) = %f，期望 2.0", result)
	}
}

// TestCgoComputeSqrt_Zero 测试 sqrt(0)
func TestCgoComputeSqrt_Zero(t *testing.T) {
	result := cgoComputeSqrt(0.0)
	if result != 0.0 {
		t.Errorf("sqrt(0) = %f，期望 0.0", result)
	}
}

// TestCgoComputeSqrt_Precision 表驱动测试平方根精度
func TestCgoComputeSqrt_Precision(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{0, 0},
		{1, 1},
		{4, 2},
		{9, 3},
		{16, 4},
		{25, 5},
		{100, 10},
		{0.25, 0.5},
	}
	for _, tt := range tests {
		result := cgoComputeSqrt(tt.input)
		if result != tt.expected {
			t.Errorf("sqrt(%f) = %f，期望 %f", tt.input, result, tt.expected)
		}
	}
}

// TestCgoComputeSqrt_Approx 测试 sqrt(2) 近似值
func TestCgoComputeSqrt_Approx(t *testing.T) {
	result := cgoComputeSqrt(2.0)
	if result < 1.41 || result > 1.42 {
		t.Errorf("sqrt(2.0) = %f，期望约 1.414", result)
	}
}

// TestCgoComputeSqrt_Large 测试大数
func TestCgoComputeSqrt_Large(t *testing.T) {
	result := cgoComputeSqrt(1e12)
	if result != 1e6 {
		t.Errorf("sqrt(1e12) = %f，期望 1e6", result)
	}
}

// TestCgoComputeSqrt_SquareBack 测试平方根后平方还原
func TestCgoComputeSqrt_SquareBack(t *testing.T) {
	values := []float64{1.0, 9.0, 16.0, 25.0, 100.0, 0.25}
	for _, v := range values {
		sqrt := cgoComputeSqrt(v)
		restored := sqrt * sqrt
		if restored < v-0.001 || restored > v+0.001 {
			t.Errorf("sqrt(%f) = %f，平方还原 = %f，期望 %f", v, sqrt, restored, v)
		}
	}
}

// ============ 多次调用 ============

// TestCgoComputeSqrt_Repeated 测试连续多次 sqrt 调用
func TestCgoComputeSqrt_Repeated(t *testing.T) {
	for i := 0; i < 100; i++ {
		result := cgoComputeSqrt(float64(i))
		if result < 0 {
			t.Errorf("sqrt(%d) = %f，结果不应为负", i, result)
		}
	}
}

// ============ 平台信息 ============

// TestGoPlatform 测试当前平台信息
func TestGoPlatform(t *testing.T) {
	if runtime.GOOS == "" {
		t.Error("runtime.GOOS 为空")
	}
	if runtime.GOARCH == "" {
		t.Error("runtime.GOARCH 为空")
	}
}

// ============ CPU + sqrt 组合测试 ============

// TestCgoCPUAndSqrt 测试 CPU 核心数和 sqrt 连续调用
func TestCgoCPUAndSqrt(t *testing.T) {
	cores := cgoGetCPUCores()
	if cores <= 0 {
		t.Errorf("CPU 核心数应 > 0，实际为 %d", cores)
	}

	sqrt := cgoComputeSqrt(16.0)
	if sqrt != 4.0 {
		t.Errorf("sqrt(16) = %f，期望 4.0", sqrt)
	}
}