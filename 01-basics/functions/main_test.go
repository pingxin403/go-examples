// main_test.go — 对 functions 包中多返回值、变参函数、defer、闭包等特性的完整测试套件
//
// 本文件演示 Go 测试的多种模式：
//   - 表格驱动测试（Table-Driven Tests）
//   - 子测试（t.Run）
//   - 闭包行为验证测试
//   - 错误处理与边界值测试
//   - 命名返回值验证
package main

import (
	"testing"
)

// ============================================================
// 1. 多返回值测试
// ============================================================

// TestStats 表格驱动测试多返回值 stats 函数
func TestStats(t *testing.T) {
	tests := []struct {
		name       string
		nums       []int
		wantMin    int
		wantMax    int
		wantSum    int
		wantAvg    float64
	}{
		{name: "正常数据", nums: []int{34, 12, 89, 5, 67, 23}, wantMin: 5, wantMax: 89, wantSum: 230, wantAvg: 38.333333333333336},
		{name: "单个元素", nums: []int{42}, wantMin: 42, wantMax: 42, wantSum: 42, wantAvg: 42.0},
		{name: "相同元素", nums: []int{5, 5, 5}, wantMin: 5, wantMax: 5, wantSum: 15, wantAvg: 5.0},
		{name: "负数", nums: []int{-10, -5, 0, 5}, wantMin: -10, wantMax: 5, wantSum: -10, wantAvg: -2.5},
		{name: "空 slice", nums: []int{}, wantMin: 0, wantMax: 0, wantSum: 0, wantAvg: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			min, max, sum, avg := stats(tt.nums)
			if min != tt.wantMin {
				t.Errorf("stats(%v) min = %d; 期望 %d", tt.nums, min, tt.wantMin)
			}
			if max != tt.wantMax {
				t.Errorf("stats(%v) max = %d; 期望 %d", tt.nums, max, tt.wantMax)
			}
			if sum != tt.wantSum {
				t.Errorf("stats(%v) sum = %d; 期望 %d", tt.nums, sum, tt.wantSum)
			}
			if avg != tt.wantAvg {
				t.Errorf("stats(%v) avg = %f; 期望 %f", tt.nums, avg, tt.wantAvg)
			}
		})
	}
}

// TestStatsIgnore 验证使用 _ 忽略不需要的返回值
func TestStatsIgnore(t *testing.T) {
	_, _, onlySum, _ := stats([]int{1, 2, 3})
	if onlySum != 6 {
		t.Errorf("忽略其他返回值后 sum = %d; 期望 6", onlySum)
	}
}

// ============================================================
// 2. 命名返回值测试
// ============================================================

// TestSplit 表格驱动测试命名返回值
func TestSplit(t *testing.T) {
	tests := []struct {
		name string
		sum  int
		wantX, wantY int
	}{
		{name: "27 split", sum: 27, wantX: 12, wantY: 15},
		{name: "0 split", sum: 0, wantX: 0, wantY: 0},
		{name: "9 split", sum: 9, wantX: 4, wantY: 5},
		{name: "负数", sum: -9, wantX: -4, wantY: -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y := split(tt.sum)
			if x != tt.wantX || y != tt.wantY {
				t.Errorf("split(%d) = (%d, %d); 期望 (%d, %d)",
					tt.sum, x, y, tt.wantX, tt.wantY)
			}
			// 验证 x + y == sum
			if x+y != tt.sum {
				t.Errorf("split(%d): %d + %d != %d", tt.sum, x, y, tt.sum)
			}
		})
	}
}

// ============================================================
// 3. 变参函数测试
// ============================================================

// TestSum 表格驱动测试变参函数 sum
func TestSum(t *testing.T) {
	tests := []struct {
		name string
		nums []int
		want int
	}{
		{name: "0 个参数", nums: []int{}, want: 0},
		{name: "1 个参数", nums: []int{42}, want: 42},
		{name: "2 个参数", nums: []int{1, 2}, want: 3},
		{name: "5 个参数", nums: []int{1, 2, 3, 4, 5}, want: 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sum(tt.nums...)
			if got != tt.want {
				t.Errorf("sum(%v) = %d; 期望 %d", tt.nums, got, tt.want)
			}
		})
	}
}

// TestSumSpreadSlice 验证展开 slice 传入变参函数
func TestSumSpreadSlice(t *testing.T) {
	nums := []int{10, 20, 30}
	got := sum(nums...)
	if got != 60 {
		t.Errorf("sum(nums...) = %d; 期望 60", got)
	}
}

// TestGreet 表格驱动测试字符串变参函数
func TestGreet(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		names  []string
		want   string
	}{
		{name: "单个名字", prefix: "你好", names: []string{"张三"}, want: "你好张三"},
		{name: "多个名字", prefix: "Hello", names: []string{"Alice", "Bob", "Charlie"}, want: "HelloAlice, Bob, Charlie"},
		{name: "无名字", prefix: "Hi", names: []string{}, want: "Hi"},
		{name: "两个名字", prefix: "欢迎", names: []string{"小明", "小红"}, want: "欢迎小明, 小红"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := greet(tt.prefix, tt.names...)
			if got != tt.want {
				t.Errorf("greet(%q, %v) = %q; 期望 %q", tt.prefix, tt.names, got, tt.want)
			}
		})
	}
}

// ============================================================
// 4. defer 测试
// ============================================================

// TestDemonstrateDefer 验证 defer 执行不 panic
func TestDemonstrateDefer(t *testing.T) {
	// 验证 demonstrateDefer 不会 panic 且能正常返回
	demonstrateDefer()
}

// ============================================================
// 5. 函数作为值测试
// ============================================================

// TestDouble 验证函数可赋值给变量
func TestDouble(t *testing.T) {
	tests := []struct {
		name string
		input int
		want int
	}{
		{name: "double(7)", input: 7, want: 14},
		{name: "double(0)", input: 0, want: 0},
		{name: "double(-3)", input: -3, want: -6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := double
			got := fn(tt.input)
			if got != tt.want {
				t.Errorf("double(%d) = %d; 期望 %d", tt.input, got, tt.want)
			}
		})
	}
}

// ============================================================
// 6. 闭包测试
// ============================================================

// TestCounterClosure 验证 counter 闭包的独立状态
func TestCounterClosure(t *testing.T) {
	t.Run("计数器递增行为", func(t *testing.T) {
		c := counter()
		if got := c(); got != 1 {
			t.Errorf("c() 第 1 次 = %d; 期望 1", got)
		}
		if got := c(); got != 2 {
			t.Errorf("c() 第 2 次 = %d; 期望 2", got)
		}
		if got := c(); got != 3 {
			t.Errorf("c() 第 3 次 = %d; 期望 3", got)
		}
	})

	t.Run("独立计数器互不影响", func(t *testing.T) {
		c1 := counter()
		c2 := counter()

		c1()                   // c1: 1
		c1()                   // c1: 2
		_ = c2()               // c2: 1

		if got := c1(); got != 3 {
			t.Errorf("c1() 应为 3, 得到 %d", got)
		}
		if got := c2(); got != 2 {
			t.Errorf("c2() 应为 2, 得到 %d", got)
		}
	})
}

// TestNewCounter 表格驱动测试带参数的计数器闭包
func TestNewCounter(t *testing.T) {
	t.Run("偶数序列", func(t *testing.T) {
		evenCounter := newCounter(0, 2)
		expected := []int{0, 2, 4}
		for i, want := range expected {
			if got := evenCounter(); got != want {
				t.Errorf("evenCounter() 第 %d 次 = %d; 期望 %d", i+1, got, want)
			}
		}
	})

	t.Run("奇数序列", func(t *testing.T) {
		oddCounter := newCounter(1, 2)
		expected := []int{1, 3, 5}
		for i, want := range expected {
			if got := oddCounter(); got != want {
				t.Errorf("oddCounter() 第 %d 次 = %d; 期望 %d", i+1, got, want)
			}
		}
	})
}

// TestFibonacci 验证 Fibonacci 闭包生成器
func TestFibonacci(t *testing.T) {
	fib := fibonacci()
	expected := []int{0, 1, 1, 2, 3, 5, 8, 13, 21, 34}
	for i, want := range expected {
		if got := fib(); got != want {
			t.Errorf("fib(%d) = %d; 期望 %d", i+1, got, want)
		}
	}
}

// ============================================================
// 7. init 函数测试
// ============================================================

// TestAppVersion 验证 init 函数正确初始化 appVersion
func TestAppVersion(t *testing.T) {
	if appVersion != "1.0.0" {
		t.Errorf("appVersion = %q; 期望 %q (由 init 初始化)", appVersion, "1.0.0")
	}
}

// ============================================================
// 8. safeDiv 测试
// ============================================================

// TestSafeDiv 表格驱动测试带错误处理的除法
func TestSafeDiv(t *testing.T) {
	t.Run("正常除法", func(t *testing.T) {
		got, err := safeDiv(10, 3)
		if err != nil {
			t.Fatalf("期望无错误，得到 %v", err)
		}
		if got != 3 {
			t.Errorf("safeDiv(10, 3) = %d; 期望 3", got)
		}
	})

	t.Run("整除", func(t *testing.T) {
		got, err := safeDiv(9, 3)
		if err != nil {
			t.Fatalf("期望无错误，得到 %v", err)
		}
		if got != 3 {
			t.Errorf("safeDiv(9, 3) = %d; 期望 3", got)
		}
	})

	t.Run("除零错误", func(t *testing.T) {
		_, err := safeDiv(5, 0)
		if err == nil {
			t.Fatal("期望除零错误，但没有得到")
		}
		if err.Error() != "除数不能为 0" {
			t.Errorf("期望错误 '除数不能为 0'，得到 %v", err)
		}
	})
}

// ============================================================
// 9. 闭包陷阱测试
// ============================================================

// TestClosureTrap 验证循环变量捕获的经典问题
func TestClosureTrap(t *testing.T) {
	t.Run("错误示例：所有闭包共享 i", func(t *testing.T) {
		// Go 1.22+ 已修复循环变量捕获，此测试验证正确行为
		got := CheckClosureTrap()
		want := []int{1, 2, 3}
		for i, g := range got {
			if g != want[i] {
				t.Errorf("got %d, want %d", g, want[i])
			}
		}
	})

	t.Run("正确示例：每个闭包捕获不同的 i", func(t *testing.T) {
		var funcs []func() int
		for i := 1; i <= 3; i++ {
			i := i // 创建新变量
			funcs = append(funcs, func() int {
				return i
			})
		}
		for i, f := range funcs {
			if got := f(); got != i+1 {
				t.Errorf("func[%d]() = %d; 期望 %d", i, got, i+1)
			}
		}
	})
}