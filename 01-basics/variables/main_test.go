// main_test.go — 对 variables 包中变量声明、类型转换、零值等特性的完整测试套件
//
// 本文件演示 Go 测试的多种模式：
//   - 表格驱动测试（Table-Driven Tests）
//   - 子测试（t.Run）
//   - 零值断言测试
//   - 常量与 iota 验证测试
package main

import (
	"testing"
)

// ============================================================
// 1. 表格驱动测试 — 变量声明与类型推断
// ============================================================

// TestPiConstant 验证 Pi 常量值
func TestPiConstant(t *testing.T) {
	tests := []struct {
		name string
		got  float64
		want float64
	}{
		{name: "Pi 精确值", got: Pi, want: 3.1415926},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("Pi = %f; 期望 %f", tt.got, tt.want)
			}
		})
	}
}

// TestIotaConstants 验证 iota 枚举常量的连续值
func TestIotaConstants(t *testing.T) {
	tests := []struct {
		name string
		got  int
		want int
	}{
		{name: "StatusPending 应为 0", got: StatusPending, want: 0},
		{name: "StatusActive 应为 1", got: StatusActive, want: 1},
		{name: "StatusInactive 应为 2", got: StatusInactive, want: 2},
		{name: "StatusDeleted 应为 3", got: StatusDeleted, want: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("iota 常量值 = %d; 期望 %d", tt.got, tt.want)
			}
		})
	}
}

// TestPermissionFlags 验证权限位标志的位运算值
func TestPermissionFlags(t *testing.T) {
	tests := []struct {
		name string
		got  int
		want int
	}{
		{name: "Read 应为 1 (0001)", got: Read, want: 1},
		{name: "Write 应为 2 (0010)", got: Write, want: 2},
		{name: "Execute 应为 4 (0100)", got: Execute, want: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("权限标志 = %d; 期望 %d", tt.got, tt.want)
			}
		})
	}
}

// TestPermissionCombination 验证权限组合的位运算
func TestPermissionCombination(t *testing.T) {
	// Read | Write = 3 (0011)
	perm := Read | Write
	if perm != 3 {
		t.Errorf("Read | Write = %d; 期望 3 (0011)", perm)
	}

	// 验证没有 Execute 权限
	if perm&Execute != 0 {
		t.Errorf("不应有 Execute 权限，但 perm&Execute != 0")
	}

	// 验证有 Read 权限
	if perm&Read == 0 {
		t.Errorf("应有 Read 权限，但 perm&Read == 0")
	}
}

// ============================================================
// 2. 零值测试 — Go 变量不初始化也有默认值
// ============================================================

// TestZeroValues 验证各种类型的零值
func TestZeroValues(t *testing.T) {
	t.Run("int 零值应为 0", func(t *testing.T) {
		if zeroInt != 0 {
			t.Errorf("zeroInt = %d; 期望 0", zeroInt)
		}
	})

	t.Run("float64 零值应为 0.0", func(t *testing.T) {
		if zeroFloat != 0.0 {
			t.Errorf("zeroFloat = %f; 期望 0.0", zeroFloat)
		}
	})

	t.Run("bool 零值应为 false", func(t *testing.T) {
		if zeroBool != false {
			t.Errorf("zeroBool = %t; 期望 false", zeroBool)
		}
	})

	t.Run("string 零值应为空串", func(t *testing.T) {
		if zeroString != "" {
			t.Errorf("zeroString = %q; 期望空串", zeroString)
		}
	})

	t.Run("pointer 零值应为 nil", func(t *testing.T) {
		if zeroPtr != nil {
			t.Errorf("zeroPtr = %v; 期望 nil", zeroPtr)
		}
	})

	t.Run("slice 零值应为 nil", func(t *testing.T) {
		if zeroSlice != nil {
			t.Errorf("zeroSlice = %v; 期望 nil", zeroSlice)
		}
		if len(zeroSlice) != 0 {
			t.Errorf("zeroSlice len = %d; 期望 0", len(zeroSlice))
		}
	})

	t.Run("map 零值应为 nil", func(t *testing.T) {
		if zeroMap != nil {
			t.Errorf("zeroMap = %v; 期望 nil", zeroMap)
		}
	})
}

// ============================================================
// 3. 类型转换测试
// ============================================================

// TestTypeConversions 验证类型转换的正确行为
func TestTypeConversions(t *testing.T) {
	t.Run("int 转 float64", func(t *testing.T) {
		var intVal int = 42
		floatVal := float64(intVal)
		if floatVal != 42.0 {
			t.Errorf("int(42) → float64 = %f; 期望 42.0", floatVal)
		}
	})

	t.Run("float64 转 int（截断小数）", func(t *testing.T) {
		var piFloat float64 = 3.1415926
		piInt := int(piFloat)
		if piInt != 3 {
			t.Errorf("float64(3.1415926) → int = %d; 期望 3", piInt)
		}
	})

	t.Run("int 转 string（Unicode 码点）", func(t *testing.T) {
		char := string(rune(65))
		if char != "A" {
			t.Errorf("string(rune(65)) = %q; 期望 'A'", char)
		}
	})

	t.Run("string 转 rune slice", func(t *testing.T) {
		hello := "你好"
		runes := []rune(hello)
		if len(runes) != 2 {
			t.Errorf("[]rune('你好') 长度 = %d; 期望 2", len(runes))
		}
	})
}

// ============================================================
// 4. 多重赋值与交换测试
// ============================================================

// TestSwapValues 验证多重赋值交换变量
func TestSwapValues(t *testing.T) {
	tests := []struct {
		name   string
		a, b   int
		wantA, wantB int
	}{
		{name: "正数交换", a: 1, b: 2, wantA: 2, wantB: 1},
		{name: "负数交换", a: -5, b: 10, wantA: 10, wantB: -5},
		{name: "相等值交换", a: 7, b: 7, wantA: 7, wantB: 7},
		{name: "零值交换", a: 0, b: 100, wantA: 100, wantB: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotA, gotB := SwapValues(tt.a, tt.b)
			if gotA != tt.wantA || gotB != tt.wantB {
				t.Errorf("SwapValues(%d, %d) = (%d, %d); 期望 (%d, %d)",
					tt.a, tt.b, gotA, gotB, tt.wantA, tt.wantB)
			}
		})
	}
}

// TestDivide 验证多返回值的除法函数
func TestDivide(t *testing.T) {
	tests := []struct {
		name       string
		a, b       int
		wantQuo    int
		wantRem    int
	}{
		{name: "10 / 3", a: 10, b: 3, wantQuo: 3, wantRem: 1},
		{name: "9 / 3", a: 9, b: 3, wantQuo: 3, wantRem: 0},
		{name: "0 / 5", a: 0, b: 5, wantQuo: 0, wantRem: 0},
		{name: "负数除法", a: -10, b: 3, wantQuo: -3, wantRem: -1},
		{name: "大数除法", a: 1000000, b: 7, wantQuo: 142857, wantRem: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQuo, gotRem := Divide(tt.a, tt.b)
			if gotQuo != tt.wantQuo || gotRem != tt.wantRem {
				t.Errorf("Divide(%d, %d) = (%d, %d); 期望 (%d, %d)",
					tt.a, tt.b, gotQuo, gotRem, tt.wantQuo, tt.wantRem)
			}
		})
	}
}

// ============================================================
// 5. 命名规范测试
// ============================================================

// TestIsCamelCase 验证驼峰命名检测
func TestIsCamelCase(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{name: "小写开头是驼峰", s: "userName", want: true},
		{name: "大写开头不是驼峰", s: "UserName", want: false},
		{name: "空字符串不是驼峰", s: "", want: false},
		{name: "单字母小写是驼峰", s: "a", want: true},
		{name: "单字母大写不是驼峰", s: "A", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCamelCase(tt.s)
			if got != tt.want {
				t.Errorf("IsCamelCase(%q) = %t; 期望 %t", tt.s, got, tt.want)
			}
		})
	}
}

// ============================================================
// 6. 常量组合验证
// ============================================================

// TestGetIotaConstants 验证 iota 常量映射
func TestGetIotaConstants(t *testing.T) {
	constants := GetIotaConstants()
	tests := []struct {
		name string
		key  string
		want int
	}{
		{name: "StatusPending", key: "StatusPending", want: 0},
		{name: "StatusActive", key: "StatusActive", want: 1},
		{name: "StatusInactive", key: "StatusInactive", want: 2},
		{name: "StatusDeleted", key: "StatusDeleted", want: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := constants[tt.key]
			if got != tt.want {
				t.Errorf("iota %s = %d; 期望 %d", tt.key, got, tt.want)
			}
		})
	}
}