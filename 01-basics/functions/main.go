package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ============================================================
// 包级别变量 — init 函数可以修改它
// ============================================================
var appVersion string

// init 函数在 main 之前自动执行，每个包可以有多个 init
// 适合初始化配置、注册信息等
func init() {
	appVersion = "1.0.0"
	fmt.Println("[init 1] appVersion 已初始化为", appVersion)
}

// 第二个 init，按文件内的声明顺序执行（跨文件按导入顺序）
func init() {
	fmt.Println("[init 2] 其他初始化工作完成")
}

// ============================================================
// 1. 多返回值
// ============================================================
// 计算数字的统计信息，返回多个值
func stats(nums []int) (min int, max int, sum int, avg float64) {
	if len(nums) == 0 {
		return 0, 0, 0, 0
	}
	// 这里使用了命名返回值，直接 return 即可
	min = nums[0]
	max = nums[0]
	for _, v := range nums {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
	}
	avg = float64(sum) / float64(len(nums))
	return // 裸返回，返回命名返回值的当前值
}

// ============================================================
// 2. 命名返回值（Named Return Values）
// ============================================================
// 命名返回值在函数签名中就定义了变量名，函数体可以直接使用
// 注意：裸返回 (naked return) 在短函数中可读性好，长函数中应避免
func split(sum int) (x, y int) {
	x = sum * 4 / 9
	y = sum - x
	return // 裸返回：返回 x, y 的当前值
}

// ============================================================
// 3. 变参函数（Variadic Function）
// ============================================================
// ...int 表示可以传入任意数量的 int 参数
// 在函数内部，nums 的类型是 []int
func sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// 变参函数也可以接受 0 个参数
func greet(prefix string, names ...string) string {
	result := prefix
	for i, name := range names {
		if i > 0 {
			result += ", "
		}
		result += name
	}
	return result
}

// ============================================================
// 4. defer — 延迟执行（LIFO 后进先出）
// ============================================================
func demonstrateDefer() {
	fmt.Println("  → 开始 demonstrateDefer")

	// defer 语句在函数返回时执行，后进先出（栈序）
	for i := 1; i <= 3; i++ {
		defer fmt.Printf("  → defer #%d (LIFO: 后进先出)\n", i)
	}

	fmt.Println("  → 函数体执行完毕，准备返回...")
	// 返回时依次输出 defer #3, #2, #1
}

// 实用 defer: 文件操作中的资源清理
func readConfig() error {
	// 获取当前目录下的 go.mod
	path := filepath.Join(".", "go.mod")
	fmt.Printf("尝试打开 %s ...\n", path)

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	// defer 确保文件一定会被关闭，无论函数走哪个 return 路径
	defer func() {
		fmt.Println("  [defer] 关闭文件句柄")
		file.Close()
	}()

	// 读取文件信息 — 模拟一些操作
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}
	fmt.Printf("  文件大小: %d bytes\n", info.Size())

	return nil // 这里 file.Close() 会通过 defer 自动调用
}

// ============================================================
// 5. 函数作为值（First-Class Functions）
// ============================================================
// 函数类型可以像其他类型一样被赋值给变量
func double(x int) int {
	return x * 2
}

// ============================================================
// 6. 匿名函数
// ============================================================
func useAnonymous() {
	// 匿名函数：没有名字，直接定义并调用
	result := func(a, b int) int {
		return a + b
	}(3, 4) // 末尾的 () 表示立即调用
	fmt.Printf("匿名函数立即调用: 3 + 4 = %d\n", result)

	// 匿名函数赋值给变量
	multiply := func(a, b int) int {
		return a * b
	}
	fmt.Printf("匿名函数赋值给变量: 3 × 4 = %d\n", multiply(3, 4))
}

// ============================================================
// 7. 闭包（Closure）
// ============================================================
// 闭包 = 函数 + 它引用的外部变量
// 外部变量的生命周期被延长到闭包的生命周期

// counter 返回一个闭包：每次调用递增并返回当前值
func counter() func() int {
	i := 0 // 这个变量被闭包捕获
	return func() int {
		i++
		return i
	}
}

// 更实用的闭包：带起始值和步长
func newCounter(start, step int) func() int {
	i := start - step // 内部状态，外部不可访问
	return func() int {
		i += step
		return i
	}
}

// 闭包实现 Fibonacci 数列生成器
func fibonacci() func() int {
	a, b := 0, 1
	return func() int {
		result := a
		a, b = b, a+b
		return result
	}
}

// ============================================================
// 辅助：演示错误处理的函数
// ============================================================
func safeDiv(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("除数不能为 0")
	}
	return a / b, nil
}

func main() {
	// ============================================================
	// 1. 多返回值
	// ============================================================
	fmt.Println("=== 多返回值 ===")
	numbers := []int{34, 12, 89, 5, 67, 23}

	// 调用多返回值函数
	minVal, maxVal, totalVal, avgVal := stats(numbers)
	fmt.Printf("numbers=%v\n", numbers)
	fmt.Printf("min=%d, max=%d, sum=%d, avg=%.2f\n", minVal, maxVal, totalVal, avgVal)

	// 使用 _ 忽略不需要的返回值
	_, _, onlySum, _ := stats(numbers)
	fmt.Printf("只取 sum: %d\n", onlySum)

	// ============================================================
	// 2. 命名返回值
	// ============================================================
	fmt.Println("\n=== 命名返回值 ===")
	x, y := split(27)
	fmt.Printf("split(27) = (%d, %d), 验证: %d + %d = %d\n", x, y, x, y, x+y)

	// ============================================================
	// 3. 变参函数
	// ============================================================
	fmt.Println("\n=== 变参函数 ===")
	fmt.Printf("sum() = %d\n", sum())                     // 0 个参数
	fmt.Printf("sum(1, 2) = %d\n", sum(1, 2))             // 2 个参数
	fmt.Printf("sum(1, 2, 3, 4, 5) = %d\n", sum(1, 2, 3, 4, 5)) // 5 个参数

	// 展开 slice 传入变参函数
	nums := []int{10, 20, 30}
	fmt.Printf("sum(nums...) = %d (展开 slice)\n", sum(nums...))

	// 字符串变参
	fmt.Println(greet("你好", "张三"))
	fmt.Println(greet("Hello", "Alice", "Bob", "Charlie"))

	// ============================================================
	// 4. defer — LIFO 演示
	// ============================================================
	fmt.Println("\n=== defer (LIFO 后进先出) ===")
	demonstrateDefer()
	fmt.Println("  [main] demonstrateDefer 已返回")

	// ============================================================
	// 5. defer — 实用资源清理
	// ============================================================
	fmt.Println("\n=== defer 实用: 文件操作 ===")
	err := readConfig()
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	}

	// ============================================================
	// 6. 函数作为值
	// ============================================================
	fmt.Println("\n=== 函数作为值 ===")
	// 函数可以赋值给变量
	fn := double
	fmt.Printf("double(7) = %d (通过变量调用)\n", fn(7))

	// 函数可以作为参数传递
	apply := func(f func(int) int, v int) int {
		return f(v)
	}
	fmt.Printf("apply(double, 5) = %d\n", apply(double, 5))

	// 函数可以作为返回值（见下面的闭包部分）

	// ============================================================
	// 7. 匿名函数
	// ============================================================
	fmt.Println("\n=== 匿名函数 ===")
	useAnonymous()

	// 匿名函数在 go 中的常见用法: goroutine
	go func(msg string) {
		fmt.Println("goroutine 中执行:", msg)
	}("并发执行!")
	// 注意: 这里不等待 goroutine 完成，为了输出完整，下面用闭包演示时等待一下
	// 这里只是演示语法，实际需要 sync.WaitGroup

	// ============================================================
	// 8. 闭包（Closure）
	// ============================================================
	fmt.Println("\n=== 闭包 ===")

	// 闭包示例: 计数器工厂
	c1 := counter()
	c2 := counter() // 独立的计数器，各自维护自己的 i

	fmt.Println("计数器 c1:")
	fmt.Printf("  c1() = %d\n", c1())
	fmt.Printf("  c1() = %d\n", c1())
	fmt.Printf("  c1() = %d\n", c1())

	fmt.Println("计数器 c2 (独立):")
	fmt.Printf("  c2() = %d\n", c2())
	fmt.Printf("  c2() = %d\n", c2())

	fmt.Println("c1 继续 (不受 c2 影响):")
	fmt.Printf("  c1() = %d\n", c1())

	// 带参数的闭包工厂
	evenCounter := newCounter(0, 2) // 偶数序列: 0, 2, 4, ...
	fmt.Println("偶数计数器:")
	for i := 0; i < 3; i++ {
		fmt.Printf("  %d\n", evenCounter())
	}

	// Fibonacci 生成器
	fib := fibonacci()
	fmt.Println("Fibonacci 数列:")
	for i := 0; i < 10; i++ {
		fmt.Printf("  fib(%d) = %d\n", i+1, fib())
	}

	// ============================================================
	// 闭包陷阱: 循环变量捕获（经典问题）
	// ============================================================
	fmt.Println("\n=== 闭包陷阱: 循环变量捕获 ===")
	// 错误示例：所有闭包共享同一个 loop 变量
	var funcs []func()
	for i := 1; i <= 3; i++ {
		funcs = append(funcs, func() {
			fmt.Printf("  错误闭包: i=%d\n", i) // 全部输出 3
		})
	}
	fmt.Println("错误示例（所有闭包都引用同一个 i）:")
	for _, f := range funcs {
		f()
	}

	// 正确修复：每次循环创建新变量
	var funcsFixed []func()
	for i := 1; i <= 3; i++ {
		i := i // 创建新变量，捕获当前循环的值
		funcsFixed = append(funcsFixed, func() {
			fmt.Printf("  正确闭包: i=%d\n", i)
		})
	}
	fmt.Println("正确修复（每个闭包捕获不同的 i）:")
	for _, f := range funcsFixed {
		f()
	}

	// Go 1.22+ 已经修复了这个陷阱，但为了兼容旧版本代码，了解此模式仍然重要

	// ============================================================
	// 总结
	// ============================================================
	fmt.Println("\n=== 函数特性总结 ===")
	fmt.Println("✅ 多返回值 — 错误处理、解包赋值")
	fmt.Println("✅ 命名返回值 — 裸返回，短函数适用")
	fmt.Println("✅ 变参函数 — 灵活参数传递")
	fmt.Println("✅ defer — LIFO 延迟执行，资源清理")
	fmt.Println("✅ 一等公民 — 赋值、传参、返回")
	fmt.Println("✅ 匿名函数 — 立即执行或赋值")
	fmt.Println("✅ 闭包 — 捕获外部变量，状态保持")
	fmt.Println("✅ init — 包初始化，main 前执行")
}