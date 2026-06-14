package main

import (
	"fmt"
	"strings"
)

// ============================================================
// 1. 泛型函数 — 类型参数在函数名后的方括号中声明
// ============================================================

// Min 返回两个值中的较小者
// [T constraints.Ordered] 表示 T 必须支持 < <= > >= 等比较操作
// 注意: Go 1.21+ 中的 cmp.Ordered 在 cmp 包中
// 这里我们手动使用 ~ 约束自定义类型

// Comparable 自定义约束：允许所有底层类型为有序类型的值
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64 | ~string
}

// Min 泛型函数 — 返回两个值中的较小者
func Min[T Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Max 泛型函数 — 返回两个值中的较大者
func Max[T Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// ============================================================
// 2. 泛型函数 — 类型参数可用于多个参数
// ============================================================

// MapKeys 返回 map 中所有 key 的 slice
// K 是键的类型（必须可比较），V 是值的类型（任意类型）
func MapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// MapValues 返回 map 中所有 value 的 slice
func MapValues[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// Filter 过滤 slice 中满足条件的元素
func Filter[T any](s []T, predicate func(T) bool) []T {
	var result []T
	for _, v := range s {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

// Transform 将 []T 转换为 []U
func Transform[T, U any](s []T, mapper func(T) U) []U {
	result := make([]U, len(s))
	for i, v := range s {
		result[i] = mapper(v)
	}
	return result
}

// ============================================================
// 3. 泛型结构体
// ============================================================

// Pair 泛型结构体 — 持有两个不同类型的值
type Pair[T, U any] struct {
	First  T
	Second U
}

// String 实现了 fmt.Stringer
func (p Pair[T, U]) String() string {
	return fmt.Sprintf("(%v, %v)", p.First, p.Second)
}

// Swap 交换两个字段并返回新 Pair
func (p Pair[T, U]) Swap() Pair[U, T] {
	return Pair[U, T]{First: p.Second, Second: p.First}
}

// ============================================================
// 4. 泛型 Stack 实现
// ============================================================

// Stack 泛型栈 — 先进后出 (LIFO)
type Stack[T any] struct {
	elements []T
}

// Push 压入元素
func (s *Stack[T]) Push(v T) {
	s.elements = append(s.elements, v)
}

// Pop 弹出栈顶元素
func (s *Stack[T]) Pop() (T, error) {
	if s.IsEmpty() {
		var zero T
		return zero, fmt.Errorf("栈为空")
	}
	index := len(s.elements) - 1
	v := s.elements[index]
	s.elements = s.elements[:index]
	return v, nil
}

// Peek 查看栈顶元素但不弹出
func (s *Stack[T]) Peek() (T, error) {
	if s.IsEmpty() {
		var zero T
		return zero, fmt.Errorf("栈为空")
	}
	return s.elements[len(s.elements)-1], nil
}

// IsEmpty 检查栈是否为空
func (s *Stack[T]) IsEmpty() bool {
	return len(s.elements) == 0
}

// Size 返回栈中元素数量
func (s *Stack[T]) Size() int {
	return len(s.elements)
}

// ============================================================
// 5. 类型约束示例
// ============================================================

// Numeric 数值类型约束 — 包含所有整数和浮点数
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// Sum 对 slice 中的数值求和
func Sum[T Numeric](s []T) T {
	var sum T
	for _, v := range s {
		sum += v
	}
	return sum
}

// Average 计算 slice 中数值的平均值
// 注意：返回值是 float64，因为整数除法会截断
func Average[T Numeric](s []T) float64 {
	if len(s) == 0 {
		return 0
	}
	var sum T
	for _, v := range s {
		sum += v
	}
	return float64(sum) / float64(len(s))
}

// ============================================================
// 6. 使用 ~ 的类型约束（支持自定义类型）
// ============================================================

// Celsius 自定义温度类型 — 底层类型是 float64
type Celsius float64

// Fahrenheit 自定义温度类型 — 底层类型是 float64
type Fahrenheit float64

// String 让 Celsius 可打印
func (c Celsius) String() string {
	return fmt.Sprintf("%.1f°C", c)
}

// String 让 Fahrenheit 可打印
func (f Fahrenheit) String() string {
	return fmt.Sprintf("%.1f°F", f)
}

// 如果没有 ~，Sum 不能用于 Celsius 类型
// 使用 ~float64 后，Celsius 和 Fahrenheit 也满足约束

// ============================================================
// 7. comparable 约束
// ============================================================

// Contains 检查 slice 中是否包含指定元素
// comparable 约束要求类型支持 == 和 !=
func Contains[T comparable](slice []T, target T) bool {
	for _, v := range slice {
		if v == target {
			return true
		}
	}
	return false
}

// IndexOf 返回元素在 slice 中的索引，未找到返回 -1
func IndexOf[T comparable](slice []T, target T) int {
	for i, v := range slice {
		if v == target {
			return i
		}
	}
	return -1
}

// Unique 去重 — 返回 slice 中不重复的元素
func Unique[T comparable](slice []T) []T {
	seen := make(map[T]struct{})
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// ============================================================
// 8. main 函数 — 演示所有泛型特性
// ============================================================

func main() {
	fmt.Println("=== Go 泛型 (Go 1.18+) 综合演示 ===")

	// ============================================================
	// 8a. 泛型函数
	// ============================================================
	fmt.Println("--- 泛型函数 ---")

	fmt.Println("Min 函数（类型自动推断）:")
	fmt.Printf("  Min(3, 7) = %d\n", Min(3, 7))
	fmt.Printf("  Min(3.14, 2.71) = %.2f\n", Min(3.14, 2.71))
	fmt.Printf("  Min(\"apple\", \"banana\") = %q\n", Min("apple", "banana"))

	fmt.Println("\nMax 函数:")
	fmt.Printf("  Max(3, 7) = %d\n", Max(3, 7))
	fmt.Printf("  Max(3.14, 2.71) = %.2f\n", Max(3.14, 2.71))
	fmt.Printf("  Max(\"apple\", \"banana\") = %q\n", Max("apple", "banana"))

	// 显式指定类型参数
	fmt.Println("\n显式类型参数:")
	fmt.Printf("  Min[int](100, 200) = %d\n", Min[int](100, 200))

	// ============================================================
	// 8b. MapKeys / MapValues
	// ============================================================
	fmt.Println("\n--- MapKeys / MapValues ---")

	scores := map[string]int{
		"张三": 95,
		"李四": 82,
		"王五": 88,
		"赵六": 73,
	}

	keys := MapKeys(scores)
	values := MapValues(scores)
	fmt.Printf("  map: %v\n", scores)
	fmt.Printf("  keys:   %v (类型: %T)\n", keys, keys)
	fmt.Printf("  values: %v (类型: %T)\n", values, values)

	// ============================================================
	// 8c. Filter / Transform
	// ============================================================
	fmt.Println("\n--- Filter / Transform ---")

	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// 只保留偶数
	evens := Filter(numbers, func(n int) bool { return n%2 == 0 })
	fmt.Printf("  原始: %v\n", numbers)
	fmt.Printf("  偶数: %v\n", evens)

	// 转换成字符串
	strs := Transform(numbers, func(n int) string {
		return fmt.Sprintf("num-%d", n)
	})
	fmt.Printf("  转字符串: %v\n", strs)

	// ============================================================
	// 8d. 泛型结构体 Pair
	// ============================================================
	fmt.Println("\n--- 泛型结构体 Pair ---")

	p1 := Pair[string, int]{First: "age", Second: 30}
	p2 := Pair[string, string]{First: "name", Second: "张三"}
	p3 := Pair[float64, bool]{First: 3.14, Second: true}

	fmt.Printf("  p1: %v\n", p1)
	fmt.Printf("  p2: %v\n", p2)
	fmt.Printf("  p3: %v\n", p3)

	// 交换 Pair
	swapped := p1.Swap()
	fmt.Printf("  交换后: %v (类型: %T)\n", swapped, swapped)

	// ============================================================
	// 8e. 泛型 Stack
	// ============================================================
	fmt.Println("\n--- 泛型 Stack ---")

	// int 栈
	intStack := Stack[int]{}
	intStack.Push(10)
	intStack.Push(20)
	intStack.Push(30)

	fmt.Printf("  intStack size: %d\n", intStack.Size())
	for !intStack.IsEmpty() {
		v, err := intStack.Pop()
		if err != nil {
			fmt.Printf("  错误: %v\n", err)
		} else {
			fmt.Printf("  弹出: %d\n", v)
		}
	}

	// 空栈弹出
	_, err := intStack.Pop()
	if err != nil {
		fmt.Printf("  空栈弹出测试: %v\n", err)
	}

	// string 栈
	stringStack := Stack[string]{}
	stringStack.Push("Hello")
	stringStack.Push("World")

	top, _ := stringStack.Peek()
	fmt.Printf("  stringStack 栈顶: %q\n", top)

	// 支持任意类型
	anyStack := Stack[any]{}
	anyStack.Push(42)
	anyStack.Push("mixed")
	anyStack.Push(3.14)
	anyStack.Push(true)

	fmt.Println("  anyStack (任意类型):")
	for !anyStack.IsEmpty() {
		v, _ := anyStack.Pop()
		fmt.Printf("    弹出: %v (类型: %T)\n", v, v)
	}

	// ============================================================
	// 8f. Sum / Average 数值约束
	// ============================================================
	fmt.Println("\n--- Sum / Average (数值约束) ---")

	ints := []int{1, 2, 3, 4, 5}
	floats := []float64{1.5, 2.5, 3.5}

	fmt.Printf("  ints=%v\n", ints)
	fmt.Printf("    Sum=%d, Average=%.2f\n", Sum(ints), Average(ints))

	fmt.Printf("  floats=%v\n", floats)
	fmt.Printf("    Sum=%.1f, Average=%.2f\n", Sum(floats), Average(floats))

	// ============================================================
	// 8g. 自定义类型（~ 约束）
	// ============================================================
	fmt.Println("\n--- 自定义类型（~ 约束）---")

	// Celsius 和 Fahrenheit 的底层类型是 float64
	// 由于约束使用 ~float64 而非 float64，这些自定义类型也满足约束
	temps := []Celsius{22.5, 18.0, 30.5, 15.0}
	fmt.Printf("  Celsius 温度: ")
	for _, t := range temps {
		fmt.Printf("%v ", t)
	}
	fmt.Println()
	fmt.Printf("  平均温度: %.2f°C\n", Average(temps))

	// ============================================================
	// 8h. comparable 约束
	// ============================================================
	fmt.Println("\n--- comparable 约束 ---")

	names := []string{"张三", "李四", "王五", "张三"}
	fmt.Printf("  names: %v\n", names)
	fmt.Printf("  包含 \"王五\"? %t\n", Contains(names, "王五"))
	fmt.Printf("  包含 \"赵六\"? %t\n", Contains(names, "赵六"))
	fmt.Printf("  \"张三\" 索引: %d\n", IndexOf(names, "张三"))
	fmt.Printf("  去重后: %v\n", Unique(names))

	ints2 := []int{1, 2, 3, 2, 1, 4, 5, 3}
	fmt.Printf("\n  ints: %v\n", ints2)
	fmt.Printf("  包含 4? %t\n", Contains(ints2, 4))
	fmt.Printf("  去重后: %v\n", Unique(ints2))

	// ============================================================
	// 8i. 类型推断 vs 显式指定
	// ============================================================
	fmt.Println("\n--- 类型推断 vs 显式指定 ---")

	// Go 编译器通常能推断类型参数
	fmt.Println("  类型推断 (自动):")
	fmt.Printf("    Min(5, 10) = %d\n", Min(5, 10))

	// 无法推断时（如返回值不确定），需要显式指定
	fmt.Println("  显式指定类型参数:")
	fmt.Printf("    Transform[int, string](...) = %v\n",
		Transform(ints, func(n int) string {
			return strings.Repeat("*", n)
		}))

	// ============================================================
	// 8j. 泛型方法
	// ============================================================
	fmt.Println("\n--- 泛型方法 ---")
	fmt.Println("注意: Go 泛型方法不能有额外的类型参数（不同于泛型函数）")
	fmt.Println("结构体上的方法可以引用结构体的类型参数但不能添加新参数")
	fmt.Println("泛型函数可以添加任意数量的类型参数")
}