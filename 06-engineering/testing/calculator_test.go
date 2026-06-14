// calculator_test.go — 对 calculator 包的完整测试套件
//
// 本文件演示 Go 测试的多种模式：
//   - 表格驱动测试（Table-Driven Tests）
//   - 错误断言测试
//   - 子测试（t.Run）
//   - 性能基准测试（Benchmark）
//   - 示例函数（Example）
package calculator

import (
	"errors"
	"fmt"
	"testing"
)

// ============================================================
// 1. 表格驱动测试 — Go 最推荐的测试模式
// ============================================================

// TestAdd 使用表格驱动测试验证 Add 函数
func TestAdd(t *testing.T) {
	// 定义测试用例表格：输入 + 期望输出
	tests := []struct {
		name string // 测试用例名称，用于 t.Run 子测试
		a, b int
		want int
	}{
		{name: "正数相加", a: 1, b: 2, want: 3},
		{name: "负数相加", a: -1, b: -2, want: -3},
		{name: "正负相加", a: 5, b: -3, want: 2},
		{name: "零值相加", a: 0, b: 0, want: 0},
		{name: "大数相加", a: 1000000, b: 2000000, want: 3000000},
	}

	// 遍历表格，使用子测试隔离每个用例
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Add(tt.a, tt.b)
			if got != tt.want {
				// 使用 t.Errorf（不终止）而非 t.Fatalf，让一个用例的失败不影响其他用例
				t.Errorf("Add(%d, %d) = %d; 期望 %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestSubtract 表格驱动测试减法
func TestSubtract(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{name: "正数相减", a: 5, b: 3, want: 2},
		{name: "结果为负", a: 3, b: 5, want: -2},
		{name: "零减正数", a: 0, b: 5, want: -5},
		{name: "负数相减", a: -5, b: -3, want: -2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Subtract(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("Subtract(%d, %d) = %d; 期望 %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestMultiply 表格驱动测试乘法
func TestMultiply(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{name: "正数相乘", a: 3, b: 4, want: 12},
		{name: "乘以零", a: 5, b: 0, want: 0},
		{name: "负数相乘", a: -2, b: 3, want: -6},
		{name: "负负得正", a: -2, b: -3, want: 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Multiply(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("Multiply(%d, %d) = %d; 期望 %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// ============================================================
// 2. 测试错误返回 — 包含正常情况和边界情况
// ============================================================

// TestDivide 测试除法，包含成功和错误两种情况
func TestDivide(t *testing.T) {
	// 成功用例
	t.Run("正常除法", func(t *testing.T) {
		got, err := Divide(10, 3)
		if err != nil {
			t.Fatalf("期望无错误，得到 %v", err)
		}
		if got != 3 {
			t.Errorf("Divide(10, 3) = %d; 期望 3", got)
		}
	})

	// 整除场景
	t.Run("整除", func(t *testing.T) {
		got, err := Divide(9, 3)
		if err != nil {
			t.Fatalf("期望无错误，得到 %v", err)
		}
		if got != 3 {
			t.Errorf("Divide(9, 3) = %d; 期望 3", got)
		}
	})

	// ============================================================
	// 错误预期测试：使用 errors.Is 检查哨兵错误
	// ============================================================
	t.Run("除零错误", func(t *testing.T) {
		_, err := Divide(5, 0)
		if err == nil {
			t.Fatal("期望除零错误，但没有得到")
		}
		// 使用 errors.Is 检查是否是期望的哨兵错误
		if !IsDivideByZero(err) {
			t.Errorf("期望 ErrDivideByZero，得到 %v", err)
		}
		// 验证错误消息包含上下文信息（Wrap 后的额外内容）
		t.Logf("错误消息: %v", err)
	})

	// 零除以非零
	t.Run("零除以正数", func(t *testing.T) {
		got, err := Divide(0, 5)
		if err != nil {
			t.Fatalf("期望无错误，得到 %v", err)
		}
		if got != 0 {
			t.Errorf("Divide(0, 5) = %d; 期望 0", got)
		}
	})

	// 大数测试
	t.Run("大数除法", func(t *testing.T) {
		got, err := Divide(1000000, 3)
		if err != nil {
			t.Fatalf("期望无错误，得到 %v", err)
		}
		// 整数除法截断：1000000 / 3 = 333333
		if got != 333333 {
			t.Errorf("Divide(1000000, 3) = %d; 期望 333333", got)
		}
	})
}

// ============================================================
// 3. 测试辅助函数
// ============================================================

// IsDivideByZero 是一个测试辅助函数，用于检查错误是否由除零引起
// 使用 errors.Is 处理被 fmt.Errorf("%w: ...") 包裹的哨兵错误
func IsDivideByZero(err error) bool {
	return errors.Is(err, ErrDivideByZero)
}

// ============================================================
// 4. 基准测试（Benchmark）
// ============================================================

// BenchmarkAdd 基准测试 Add 函数的性能
//
// 运行方式：
//
//	go test -bench=BenchmarkAdd -benchmem .
func BenchmarkAdd(b *testing.B) {
	// b.N 由框架自动调整，直到获得稳定的计时结果
	for i := 0; i < b.N; i++ {
		Add(100, 200)
	}
}

// BenchmarkMultiply 基准测试 Multiply 函数的性能
func BenchmarkMultiply(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Multiply(100, 200)
	}
}

// BenchmarkDivide 基准测试 Divide 函数（正常情况）
func BenchmarkDivide(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Divide(100, 3)
	}
}

// BenchmarkAll 在一个基准测试中顺序执行所有运算
// 用 b.RunParallel 演示并行基准测试
func BenchmarkAll(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Add(1, 2)
			Subtract(5, 3)
			Multiply(3, 4)
			Divide(10, 2)
		}
	})
}

// ============================================================
// 5. 示例函数（Example Functions）— 可测试的文档
// ============================================================

// ExampleAdd 是 Add 函数的示例，会出现在 godoc 中。
// 如果输出包含 // Output: 注释，go test 会自动验证输出是否正确。
func ExampleAdd() {
	sum := Add(3, 4)
	fmt.Println(sum)
	// Output: 7
}

// ExampleDivide 展示了带错误处理的除法用法
func ExampleDivide() {
	result, err := Divide(10, 3)
	if err != nil {
		fmt.Println("出错了:", err)
		return
	}
	fmt.Println(result)

	result, err = Divide(5, 0)
	if err != nil {
		fmt.Println("出错了:", err)
	} else {
		fmt.Println(result)
	}
	// Output:
	// 3
	// 出错了: 除数不能为零: a=5, b=0
}

// ExampleCalculator_表格驱动风格 演示多个示例输入
// 注意：函数名使用 _ 连接多个单词也是合法的示例命名
func ExampleAdd_multiple() {
	fmt.Println(Add(0, 0))
	fmt.Println(Add(-5, 5))
	fmt.Println(Add(999, 1))
	// Output:
	// 0
	// 0
	// 1000
}