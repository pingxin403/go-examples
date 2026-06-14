// memory/main_test.go — Go 内存管理核心概念测试
package main

import (
	"testing"
	"unsafe"
)

// TestCreateInt — 测试 createInt 返回 *int 且值正确
func TestCreateInt(t *testing.T) {
	p := createInt()
	if p == nil {
		t.Fatal("createInt() 返回 nil")
	}
	if *p != 42 {
		t.Errorf("createInt() = %d, 期望 42", *p)
	}

	// 独立调用应返回不同地址（每次逃逸分配新的堆对象）
	p2 := createInt()
	if p == p2 {
		t.Log("注意: 两次 createInt 返回相同地址 (编译器优化)")
	}
}

// TestCreateLargeArray — 测试 createLargeArray 返回正确长度和值的数组
func TestCreateLargeArray(t *testing.T) {
	arr := createLargeArray()
	if len(arr) != 1024 {
		t.Fatalf("createLargeArray() 长度 = %d, 期望 1024", len(arr))
	}
	// 所有元素应为零值
	for i, v := range arr {
		if v != 0 {
			t.Errorf("arr[%d] = %d, 期望 0", i, v)
		}
	}
}

// TestCreateClosure — 测试闭包捕获变量的行为
func TestCreateClosure(t *testing.T) {
	fn := createClosure()
	if fn == nil {
		t.Fatal("createClosure() 返回 nil")
	}

	// 第一次调用: 100++ → 101
	if got := fn(); got != 101 {
		t.Errorf("第一次调用: %d, 期望 101", got)
	}

	// 第二次调用: 101++ → 102 (验证闭包变量可被修改)
	if got := fn(); got != 102 {
		t.Errorf("第二次调用: %d, 期望 102", got)
	}
}

// TestCreateClosure独立性 — 多次调用 createClosure 应返回独立闭包
func TestCreateClosureIndependence(t *testing.T) {
	fn1 := createClosure()
	fn2 := createClosure()

	// fn1 执行两次
	fn1() // 101
	fn1() // 102

	// fn2 应从 101 开始（独立捕获 x）
	if got := fn2(); got != 101 {
		t.Errorf("独立闭包 fn2 首次调用: %d, 期望 101", got)
	}
}

// TestMakeVsNew — 验证 make 和 new 的行为差异（核心教学点）
func TestMakeVsNew(t *testing.T) {
	// new(int) 应返回 *int 且值为 0
	np := new(int)
	if np == nil {
		t.Fatal("new(int) 返回 nil")
	}
	if *np != 0 {
		t.Errorf("*new(int) = %d, 期望 0", *np)
	}

	// new([]int) 返回 nil slice
	npsl := new([]int)
	if npsl == nil {
		t.Fatal("new([]int) 返回 nil")
	}
	if *npsl != nil {
		t.Errorf("new([]int) = %v, 期望 nil slice", *npsl)
	}

	// make([]int, 5, 10) 返回有底层数组的 slice
	sl := make([]int, 5, 10)
	if len(sl) != 5 {
		t.Errorf("make([]int,5,10) len = %d, 期望 5", len(sl))
	}
	if cap(sl) != 10 {
		t.Errorf("make([]int,5,10) cap = %d, 期望 10", cap(sl))
	}
	// 所有元素应为零值
	for i, v := range sl {
		if v != 0 {
			t.Errorf("sl[%d] = %d, 期望 0", i, v)
		}
	}
}

// TestUnsafeSizeof — 验证 unsafe.Sizeof 对基本类型的预期值
func TestUnsafeSizeof(t *testing.T) {
	tests := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"int 大小", unsafe.Sizeof(int(0)), 8},       // 64 位系统
		{"int32 大小", unsafe.Sizeof(int32(0)), 4},
		{"int64 大小", unsafe.Sizeof(int64(0)), 8},
		{"float64 大小", unsafe.Sizeof(float64(0)), 8},
		{"bool 大小", unsafe.Sizeof(true), 1},
		{"byte 大小", unsafe.Sizeof(byte(0)), 1},
		{"string 大小", unsafe.Sizeof(""), 16}, // 指针(8) + 长度(8)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("unsafe.Sizeof = %d, 期望 %d", tt.got, tt.want)
			}
		})
	}
}

// TestStructAlign — 验证结构体对齐行为和大小
func TestStructAlign(t *testing.T) {
	// Small{A bool; B int32}: 1 + 3 pad + 4 = 8
	if got := unsafe.Sizeof(Small{}); got != 8 {
		t.Errorf("unsafe.Sizeof(Small) = %d, 期望 8 (对齐优化后)", got)
	}

	// BadAligned 含 int64，对齐后应较大
	// 1 + 7pad + 8 + 1 + 7pad = 24
	if got := unsafe.Sizeof(BadAligned{}); got != 24 {
		t.Errorf("unsafe.Sizeof(BadAligned) = %d, 期望 24 (有填充)", got)
	}

	// GoodAligned 优化后更紧凑: 8 + 1 + 1 + 6pad = 16
	if got := unsafe.Sizeof(GoodAligned{}); got != 16 {
		t.Errorf("unsafe.Sizeof(GoodAligned) = %d, 期望 16 (更紧凑)", got)
	}
}

// 类型定义 — 从 main.go 中复制以在测试中引用
type Small struct {
	A bool
	B int32
}

type BadAligned struct {
	A bool
	B int64
	C bool
}

type GoodAligned struct {
	B int64
	A bool
	C bool
}