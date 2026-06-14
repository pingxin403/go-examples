// main_test.go — Go 泛型示例测试套件
//
// 测试覆盖：
//   - 泛型函数 Min/Max（多类型）
//   - 泛型函数 MapKeys / MapValues
//   - 类型约束和类型推导
package main

import (
	"testing"
)

// ============================================================
// 1. 泛型函数 Min / Max 测试
// ============================================================

// TestMin 使用表格驱动测试验证泛型 Min 函数
func TestMin(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		if got := Min(3, 7); got != 3 {
			t.Errorf("Min(3, 7) = %d; want 3", got)
		}
		if got := Min(5, 5); got != 5 {
			t.Errorf("Min(5, 5) = %d; want 5", got)
		}
		if got := Min(-5, 3); got != -5 {
			t.Errorf("Min(-5, 3) = %d; want -5", got)
		}
	})

	t.Run("float64", func(t *testing.T) {
		if got := Min(3.14, 2.71); got != 2.71 {
			t.Errorf("Min(3.14, 2.71) = %v; want 2.71", got)
		}
	})

	t.Run("string", func(t *testing.T) {
		if got := Min("apple", "banana"); got != "apple" {
			t.Errorf(`Min("apple", "banana") = %q; want "apple"`, got)
		}
	})
}

// TestMax 测试泛型 Max 函数
func TestMax(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		if got := Max(3, 7); got != 7 {
			t.Errorf("Max(3, 7) = %d; want 7", got)
		}
	})

	t.Run("float64", func(t *testing.T) {
		if got := Max(3.14, 2.71); got != 3.14 {
			t.Errorf("Max(3.14, 2.71) = %v; want 3.14", got)
		}
	})

	t.Run("string", func(t *testing.T) {
		if got := Max("apple", "banana"); got != "banana" {
			t.Errorf(`Max("apple", "banana") = %q; want "banana"`, got)
		}
	})
}

// ============================================================
// 2. 泛型函数 MapKeys / MapValues 测试
// ============================================================

// TestMapKeys 测试泛型 MapKeys 函数
func TestMapKeys(t *testing.T) {
	t.Run("string->int", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		keys := MapKeys(m)
		if len(keys) != 3 {
			t.Fatalf("MapKeys 长度 = %d; want 3", len(keys))
		}
		got := make(map[string]bool)
		for _, k := range keys {
			got[k] = true
		}
		for _, k := range []string{"a", "b", "c"} {
			if !got[k] {
				t.Errorf("缺少 key %q", k)
			}
		}
	})

	t.Run("int->string", func(t *testing.T) {
		m := map[int]string{1: "one", 2: "two"}
		keys := MapKeys(m)
		if len(keys) != 2 {
			t.Fatalf("MapKeys 长度 = %d; want 2", len(keys))
		}
	})
}

// TestMapValues 测试泛型 MapValues 函数
func TestMapValues(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	values := MapValues(m)
	if len(values) != 3 {
		t.Fatalf("MapValues 长度 = %d; want 3", len(values))
	}
	// 验证所有值都在结果中
	got := make(map[int]bool)
	for _, v := range values {
		got[v] = true
	}
	for _, v := range []int{1, 2, 3} {
		if !got[v] {
			t.Errorf("缺少值 %d", v)
		}
	}
}

// ============================================================
// 3. 类型推导测试
// ============================================================

// TestTypeInference 验证泛型类型推导
func TestTypeInference(t *testing.T) {
	t.Run("Min 类型推导", func(t *testing.T) {
		if got := Min(100, 200); got != 100 {
			t.Errorf("Min(100, 200) = %d; want 100", got)
		}
		if got := Min(3.14, 2.71); got != 2.71 {
			t.Errorf("Min(3.14, 2.71) = %v; want 2.71", got)
		}
	})

	t.Run("Max 类型推导", func(t *testing.T) {
		if got := Max(100, 200); got != 200 {
			t.Errorf("Max(100, 200) = %d; want 200", got)
		}
	})
}