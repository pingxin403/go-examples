// 逃逸分析包单元测试
// 测试各逃逸场景下函数的逻辑正确性，不验证编译器逃逸决策
package main

import (
	"testing"
)

// TestValueOnStack 测试返回结构体值的函数
func TestValueOnStack(t *testing.T) {
	p := ValueOnStack()
	if p.X != 1 || p.Y != 2 {
		t.Errorf("ValueOnStack() = (%d, %d), 期望 (1, 2)", p.X, p.Y)
	}
}

// TestPointerEscape 测试返回结构体指针的函数
func TestPointerEscape(t *testing.T) {
	p := PointerEscape()
	if p == nil {
		t.Fatal("PointerEscape() 返回了 nil")
	}
	if p.X != 1 || p.Y != 2 {
		t.Errorf("PointerEscape() = (%d, %d), 期望 (1, 2)", p.X, p.Y)
	}
}

// TestValueAndPointerConsistency 验证值返回和指针返回的结果一致
func TestValueAndPointerConsistency(t *testing.T) {
	gotValue := ValueOnStack()
	gotPtr := PointerEscape()
	if gotValue != *gotPtr {
		t.Errorf("ValueOnStack()=%v 与 PointerEscape()=%v 不一致", gotValue, *gotPtr)
	}
}

// TestNoFmt 测试不使用 fmt 的函数
func TestNoFmt(t *testing.T) {
	tests := []struct {
		name string
		x    int
		want int
	}{
		{"x=0", 0, 0},
		{"x=1", 1, 1},
		{"x=5", 5, 25},
		{"x=10", 10, 100},
		{"x=42", 42, 1764},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NoFmt(tt.x)
			if got != tt.want {
				t.Errorf("NoFmt(%d) = %d, 期望 %d", tt.x, got, tt.want)
			}
		})
	}
}

// TestFmtSprintf 测试使用 fmt.Sprintf 的函数
func TestFmtSprintf(t *testing.T) {
	tests := []struct {
		name string
		x    int
		want string
	}{
		{"x=0", 0, "value: 0"},
		{"x=1", 1, "value: 1"},
		{"x=42", 42, "value: 42"},
		{"x=100", 100, "value: 100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FmtSprintf(tt.x)
			if got != tt.want {
				t.Errorf("FmtSprintf(%d) = %q, 期望 %q", tt.x, got, tt.want)
			}
		})
	}
}

// TestClosureNoEscape 测试不逃逸的闭包
func TestClosureNoEscape(t *testing.T) {
	result := ClosureNoEscape()
	if result != 6 {
		t.Errorf("ClosureNoEscape() = %d, 期望 6（1+2+3）", result)
	}
}

// TestClosureEscape 测试逃逸的闭包
func TestClosureEscape(t *testing.T) {
	adder := ClosureEscape()
	if adder == nil {
		t.Fatal("ClosureEscape() 返回了 nil")
	}

	tests := []struct {
		name string
		n    int
		want int
	}{
		{"第1次调用", 1, 1},
		{"第2次调用", 2, 3},
		{"第3次调用", 3, 6},
		{"第4次调用", 4, 10},
		{"第5次调用", 5, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := adder(tt.n)
			if got != tt.want {
				t.Errorf("adder(%d) = %d, 期望 %d（累加和应为 %d）", tt.n, got, tt.want, tt.want)
			}
		})
	}
}

// TestClosureEscapeMultiple 测试逃逸闭包的状态隔离
func TestClosureEscapeMultiple(t *testing.T) {
	adder1 := ClosureEscape()
	adder2 := ClosureEscape()

	adder1(10)
	adder2(5)

	if got := adder1(1); got != 11 {
		t.Errorf("adder1(1) = %d, 期望 11（adder1 应保持独立状态）", got)
	}
	if got := adder2(1); got != 6 {
		t.Errorf("adder2(1) = %d, 期望 6（adder2 应保持独立状态）", got)
	}
}

// TestDogSpeak 测试 Dog 的 Speak 方法
func TestDogSpeak(t *testing.T) {
	d := Dog{}
	if d.Speak() != 1 {
		t.Errorf("Dog.Speak() = %d, 期望 1", d.Speak())
	}
}

// TestInterfaceNoEscape 测试具体类型调用方法
func TestInterfaceNoEscape(t *testing.T) {
	result := InterfaceNoEscape()
	if result != 1 {
		t.Errorf("InterfaceNoEscape() = %d, 期望 1", result)
	}
}

// TestInterfaceEscape 测试接口转换调用方法
func TestInterfaceEscape(t *testing.T) {
	result := InterfaceEscape()
	if result != 1 {
		t.Errorf("InterfaceEscape() = %d, 期望 1", result)
	}
}

// TestInterfaceConsistency 验证直接调用和接口调用的结果一致
func TestInterfaceConsistency(t *testing.T) {
	gotDirect := InterfaceNoEscape()
	gotInterface := InterfaceEscape()
	if gotDirect != gotInterface {
		t.Errorf("InterfaceNoEscape()=%d 与 InterfaceEscape()=%d 不一致", gotDirect, gotInterface)
	}
}

// TestSmallArrayOnStack 测试小数组分配（仅验证不 panic）
func TestSmallArrayOnStack(t *testing.T) {
	// 只验证不 panic
	SmallArrayOnStack()
}

// TestLargeArrayHeap 测试大数组分配（仅验证不 panic）
func TestLargeArrayHeap(t *testing.T) {
	// 只验证不 panic
	LargeArrayHeap()
}

// TestSliceOnStack 测试小切片分配
func TestSliceOnStack(t *testing.T) {
	// 只验证不 panic
	SliceOnStack()
}

// TestSliceEscape 测试动态大小切片分配
func TestSliceEscape(t *testing.T) {
	tests := []struct {
		name string
		n    int
	}{
		{"n=0", 0},
		{"n=1", 1},
		{"n=10", 10},
		{"n=100", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SliceEscape(tt.n)
			if len(result) != tt.n {
				t.Errorf("SliceEscape(%d) 返回长度 %d, 期望 %d", tt.n, len(result), tt.n)
			}
			// 验证内容（每个元素应等于其下标）
			for i, v := range result {
				if v != i {
					t.Errorf("SliceEscape(%d)[%d] = %d, 期望 %d", tt.n, i, v, i)
				}
			}
		})
	}
}

// TestSetGlobal 测试全局变量赋值
func TestSetGlobal(t *testing.T) {
	Global = nil // 重置全局变量
	SetGlobal()
	if Global == nil {
		t.Fatal("SetGlobal() 后 Global 仍为 nil")
	}
	if Global.X != 10 || Global.Y != 20 {
		t.Errorf("Global = (%d, %d), 期望 (10, 20)", Global.X, Global.Y)
	}
}

// TestPoint 测试 Point 结构体
func TestPoint(t *testing.T) {
	tests := []struct {
		name string
		p    Point
		x, y int
	}{
		{"zero", Point{}, 0, 0},
		{"value", Point{X: 3, Y: 4}, 3, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.p.X != tt.x || tt.p.Y != tt.y {
				t.Errorf("Point = (%d, %d), 期望 (%d, %d)", tt.p.X, tt.p.Y, tt.x, tt.y)
			}
		})
	}
}