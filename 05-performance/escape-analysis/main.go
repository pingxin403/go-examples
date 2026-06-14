// 逃逸分析（Escape Analysis）示例
// 运行方式:
//
//	go build -gcflags='-m' main.go        # 查看编译器逃逸分析决策
//	go build -gcflags='-m -m' main.go      # 更详细的逃逸分析信息
//	go build -gcflags='-l -m' main.go      # 禁用内联 + 逃逸分析
//
// 逃逸分析的目的是决定变量应该分配在栈上还是堆上：
//   - 栈分配：函数返回后自动回收，零 GC 开销 -> 高性能
//   - 堆分配：需要 GC 回收，有额外开销 -> 尽可能避免
//
// 变量逃逸到堆的常见原因：
//   1. 返回指向局部变量的指针
//   2. 变量被闭包捕获
//   3. 变量被存储在接口中
//   4. 变量大小超过栈帧限制（通常 > 64KB）
//   5. 变量被分配给全局变量或其它堆对象
package main

import "fmt"

// ============================================================================
// 场景 1: 返回值（不逃逸） vs 返回指针（逃逸）
// ============================================================================

// ValueOnStack 返回结构体值 — 在栈上分配，不逃逸
// go build -gcflags='-m' 输出: moved to the stack 或未提及
func ValueOnStack() Point {
	// p 在栈上分配，通过值拷贝返回
	p := Point{X: 1, Y: 2}
	return p
}

// PointerEscape 返回结构体指针 — 逃逸到堆
// go build -gcflags='-m' 输出: moved to heap: p
func PointerEscape() *Point {
	// p 在堆上分配，因为函数返回后栈帧被销毁，但指针仍被使用
	p := &Point{X: 1, Y: 2}
	return p
}

// ============================================================================
// 场景 2: fmt 包的参数（interface{} 逃逸）
// ============================================================================

// NoFmt 不使用 fmt，全部在栈上
// 逃逸分析输出: 无 heap 分配
func NoFmt(x int) int {
	return x * x
}

// FmtSprintf 使用 fmt.Sprintf — 参数逃逸到堆
// 逃逸分析输出: x escapes to heap
// 原因: fmt.Sprintf 接受 interface{} 参数，具体类型到接口的转换导致逃逸
func FmtSprintf(x int) string {
	return fmt.Sprintf("value: %d", x)
}

// ============================================================================
// 场景 3: 闭包（Closure）捕获变量 — 逃逸
// ============================================================================

// ClosureNoEscape 闭包在函数内部调用，没有逃逸
// 注意：闭包没有返回或赋值到堆，所以不会逃逸
func ClosureNoEscape() int {
	sum := 0
	// adder 闭包捕获 sum 的引用
	adder := func(n int) {
		sum += n
	}
	adder(1)
	adder(2)
	adder(3)
	return sum
}

// ClosureEscape 返回闭包 — 被捕获的变量逃逸到堆
// 逃逸分析输出: sum escapes to heap
// 原因: 返回的函数（闭包）引用 sum，sum 必须存活在堆上
func ClosureEscape() func(int) int {
	sum := 0
	return func(n int) int {
		sum += n
		return sum
	}
}

// ============================================================================
// 场景 4: 接口（interface）赋值 — 逃逸
// ============================================================================

// InterfaceNoEscape 具体类型调用方法，没有接口转换
func InterfaceNoEscape() int {
	var d Dog
	return d.Speak()
}

// InterfaceEscape 将具体类型赋值给接口 — 逃逸到堆
// 逃逸分析输出: d escapes to heap
// 原因: 具体类型到 interface 的装箱操作需要在堆上分配
func InterfaceEscape() int {
	var a Animal = Dog{} // Dog 逃逸到堆
	return a.Speak()
}

// ============================================================================
// 场景 5: 大对象超出栈帧限制 — 逃逸
// ============================================================================

const smallSize = 1024       // 1KB — 通常不逃逸
const largeSize = 1024 * 100 // 100KB — 可能逃逸（栈帧限制通常 64KB）

// SmallArrayOnStack 小数组在栈上分配
func SmallArrayOnStack() {
	var arr [smallSize]byte
	for i := 0; i < smallSize; i++ {
		arr[i] = byte(i)
	}
	_ = arr
}

// LargeArrayHeap 大数组可能逃逸到堆（取决于 Go 版本）
func LargeArrayHeap() {
	var arr [largeSize]byte // 100KB 超过 goroutine 栈初始大小
	for i := 0; i < largeSize; i++ {
		arr[i] = byte(i)
	}
	_ = arr
}

// ============================================================================
// 场景 6: 全局变量 — 必然在堆上
// ============================================================================

// Global 全局变量分配在堆上（BSS/Data 段）
var Global *Point

func SetGlobal() {
	// p 逃逸到堆，因为被全局变量引用
	p := &Point{X: 10, Y: 20}
	Global = p
}

// ============================================================================
// 场景 7: 切片扩容导致逃逸
// ============================================================================

// SliceOnStack make 小切片已知大小，可能不逃逸
func SliceOnStack() {
	s := make([]int, 10)
	for i := 0; i < 10; i++ {
		s[i] = i
	}
	_ = s
}

// SliceEscape 切片大小由变量决定，编译器不确定大小时可能逃逸
func SliceEscape(n int) []int {
	s := make([]int, n) // n 未知，在堆上分配
	for i := 0; i < n; i++ {
		s[i] = i
	}
	return s
}

// ============================================================================
// 类型定义
// ============================================================================

// Point 二维点
type Point struct {
	X, Y int
}

// Animal 动物接口
type Animal interface {
	Speak() int
}

// Dog 狗实现 Animal 接口
type Dog struct{}

func (d Dog) Speak() int {
	return 1
}

// ============================================================================
// main 函数：逐一执行并打印逃逸决策
// ============================================================================

func main() {
	fmt.Println("=== 逃逸分析演示 ===")
	fmt.Println("运行: go build -gcflags='-m' main.go 查看各函数的逃逸决策")
	fmt.Println()

	fmt.Println("ValueOnStack:", ValueOnStack())
	fmt.Println("PointerEscape:", PointerEscape())
	fmt.Println("NoFmt:", NoFmt(42))
	fmt.Println("FmtSprintf:", FmtSprintf(42))
	fmt.Println("ClosureNoEscape:", ClosureNoEscape())

	adder := ClosureEscape()
	fmt.Println("ClosureEscape:", adder(1), adder(2))

	fmt.Println("InterfaceNoEscape:", InterfaceNoEscape())
	fmt.Println("InterfaceEscape:", InterfaceEscape())

	SmallArrayOnStack()
	LargeArrayHeap()
	SliceOnStack()
	_ = SliceEscape(100)

	SetGlobal()
	fmt.Println("Global:", Global)

	fmt.Println()
	fmt.Println("=== 小结: 减少堆分配的技巧 ===")
	fmt.Println("1. 优先返回值而不是指针")
	fmt.Println("2. 避免不必要的 fmt 调用（尤其是热路径）")
	fmt.Println("3. 避免将闭包赋值给外部变量或返回闭包")
	fmt.Println("4. 优先具体类型而不是接口类型")
	fmt.Println("5. 对 make 使用常量大小（而非变量）")
}