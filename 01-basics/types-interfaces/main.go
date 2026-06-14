package main

import (
	"fmt"
	"strconv"
)

// ============================================================
// 1. 接口定义 — Go 的接口是隐式实现的（Duck Typing）
// ============================================================

// Speaker 接口：定义了一个 Speak 方法
// 任何实现了 Speak() string 方法的类型都自动满足此接口
type Speaker interface {
	Speak() string
}

// ============================================================
// 2. 结构体实现接口（不需要显式声明 implements）
// ============================================================

// Dog 结构体
type Dog struct {
	Name string
}

// Speak 实现了 Speaker 接口
func (d Dog) Speak() string {
	return fmt.Sprintf("汪汪！我是 %s 🐕", d.Name)
}

// Cat 结构体
type Cat struct {
	Name string
}

// Speak 也实现了 Speaker 接口
func (c Cat) Speak() string {
	return fmt.Sprintf("喵～我是 %s 🐱", c.Name)
}

// Robot 结构体
type Robot struct {
	Model string
}

// Speak 也实现了 Speaker 接口
func (r Robot) Speak() string {
	return fmt.Sprintf("哔哔 %s 正在用合成语音说话 🤖", r.Model)
}

// ============================================================
// 3. 接口的使用 — 多态
// ============================================================

// 接受任何 Speaker 类型的参数
func makeSpeak(speakers ...Speaker) {
	for _, s := range speakers {
		fmt.Printf("  %s\n", s.Speak())
	}
}

// ============================================================
// 4. 空接口 interface{} / any
// ============================================================

// Go 1.18 引入了 any 作为 interface{} 的别名
// 空接口可以持有任何类型的值

func describeAny(v any) {
	fmt.Printf("  any 值: %v, 类型: %T\n", v, v)
}

// ============================================================
// 5. 类型断言 (Type Assertion)
// ============================================================

// 安全类型断言：如果类型不匹配不会 panic
func safeTypeAssert(v any) {
	if s, ok := v.(string); ok {
		fmt.Printf("  安全断言: %q 是 string, 长度=%d\n", s, len(s))
	} else {
		fmt.Printf("  安全断言: %T 不是 string\n", v)
	}
}

// 非安全类型断言：不匹配时会 panic
func unsafeTypeAssert(v any) {
	// 注意：这里如果类型不对会 panic
	s := v.(string)
	fmt.Printf("  非安全断言: %q (警告: 类型不对会 panic!)\n", s)
}

// ============================================================
// 6. 类型开关 (Type Switch)
// ============================================================

func typeSwitch(v any) {
	switch val := v.(type) {
	case nil:
		fmt.Println("  类型: nil")
	case int:
		fmt.Printf("  类型: int, 值=%d\n", val)
	case float64:
		fmt.Printf("  类型: float64, 值=%f\n", val)
	case string:
		fmt.Printf("  类型: string, 值=%q\n", val)
	case bool:
		fmt.Printf("  类型: bool, 值=%t\n", val)
	case Speaker:
		fmt.Printf("  类型: Speaker, 说话=%s\n", val.Speak())
	default:
		fmt.Printf("  未知类型: %T, 值=%v\n", val, val)
	}
}

// ============================================================
// 7. 接口嵌入 (Interface Embedding)
// ============================================================

// 基础接口
type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer interface {
	Write(p []byte) (n int, err error)
}

// 嵌入 Reader 和 Writer 形成新的接口
// 等价于 ReadWriter 包含 Read + Write + Close
type ReadWriter interface {
	Reader
	Writer
	Close() error
}

// 模拟文件结构体，实现 ReadWriter 接口
type SimulatedFile struct {
	Name string
}

func (f SimulatedFile) Read(p []byte) (n int, err error) {
	fmt.Printf("  [%s] 读取数据...\n", f.Name)
	return len(p), nil
}

func (f SimulatedFile) Write(p []byte) (n int, err error) {
	fmt.Printf("  [%s] 写入 %d 字节\n", f.Name, len(p))
	return len(p), nil
}

func (f SimulatedFile) Close() error {
	fmt.Printf("  [%s] 关闭文件\n", f.Name)
	return nil
}

// 使用 ReadWriter 接口作为参数类型
func processFile(rw ReadWriter) {
	data := make([]byte, 10)
	rw.Read(data)
	rw.Write([]byte("hello"))
	rw.Close()
}

// ============================================================
// 8. 实用示例：Stringer 接口
// ============================================================

// fmt.Stringer 是标准库接口，要求实现 String() string
// 任何实现了 String() 方法的类型都会被 fmt 包识别

// Person 结构体
type Person struct {
	FirstName string
	LastName  string
	Age       int
}

// String 实现 fmt.Stringer 接口
func (p Person) String() string {
	return fmt.Sprintf("%s %s (%d 岁)", p.LastName, p.FirstName, p.Age)
}

// IPAddr 自定义 IP 地址类型
type IPAddr [4]byte

// String 实现 fmt.Stringer
func (ip IPAddr) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}

// Money 自定义货币类型
type Money struct {
	Amount   float64
	Currency string // 货币符号，如 ¥, $, €
}

// String 实现 fmt.Stringer
func (m Money) String() string {
	// 保留两位小数的格式化
	return fmt.Sprintf("%s%.2f", m.Currency, m.Amount)
}

// ============================================================
// 9. 接口值的底层表示 (iface)
// ============================================================

func interfaceInternals() {
	fmt.Println("\n--- 接口值的底层表示 ---")
	fmt.Println("Go 的接口值在内部由两部分组成:")
	fmt.Println("  1. 具体类型的指针 (type)")
	fmt.Println("  2. 具体值的指针 (value)")
	fmt.Println("接口值为 nil 当且仅当 type 和 value 都是 nil")
	fmt.Println()

	// 接口值为 nil 的情况
	var s Speaker
	fmt.Printf("  未初始化的接口: s=%v, s==nil=%t\n", s, s == nil)

	// 赋值后
	s = Dog{Name: "旺财"}
	fmt.Printf("  赋值后: s=%v, s==nil=%t\n", s, s == nil)

	// 注意：包含 nil 指针的接口 != nil
	var p *Dog // nil 指针
	s = p      // 接口 s 的 type 是 *Dog, value 是 nil
	fmt.Printf("  包含 nil 指针的接口: s=%v, s==nil=%t (⚠️ 陷阱!)\n", s, s == nil)
}

func main() {
	fmt.Println("=== Go 接口与类型系统综合演示 ===")

	// ============================================================
	// 1. 接口定义与隐式实现
	// ============================================================
	fmt.Println("--- 接口定义与隐式实现 (Duck Typing) ---")
	fmt.Print("Go 的接口是隐式满足的。类型不需要声明 implements，")
	fmt.Println("只要实现了接口要求的所有方法，就自动满足该接口。")

	dog := Dog{Name: "旺财"}
	cat := Cat{Name: "咪咪"}
	robot := Robot{Model: "R2-D2"}

	makeSpeak(dog, cat, robot)

	// ============================================================
	// 2. 多态 — 接口作为参数
	// ============================================================
	fmt.Println("\n--- 多态 ---")
	fmt.Println("可以将不同的具体类型传递给同一个接口参数:")
	makeSpeak(dog, cat, robot)

	// 也可以放在 slice 里
	fmt.Println("\nSpeaker slice 演示:")
	speakers := []Speaker{dog, cat, robot}
	for _, s := range speakers {
		fmt.Printf("  %s\n", s.Speak())
	}

	// ============================================================
	// 3. 空接口 / any
	// ============================================================
	fmt.Println("\n--- 空接口 / any ---")
	fmt.Println("any 可以持有任何类型的值:")

	describeAny(42)
	describeAny("hello")
	describeAny(3.14)
	describeAny(true)
	describeAny(dog)
	describeAny([]int{1, 2, 3})

	// ============================================================
	// 4. 类型断言
	// ============================================================
	fmt.Println("\n--- 类型断言 ---")

	fmt.Println("安全类型断言 (ok pattern):")
	safeTypeAssert("安全字符串")
	safeTypeAssert(42)
	safeTypeAssert(3.14)

	fmt.Println("\n非安全类型断言 (谨慎使用):")
	unsafeTypeAssert("hello")
	// 下面这行如果取消注释会导致 panic:
	// unsafeTypeAssert(42) // panic: interface conversion: interface {} is int, not string

	// ============================================================
	// 5. 类型开关 (Type Switch)
	// ============================================================
	fmt.Println("\n--- Type Switch ---")
	fmt.Println("类型开关可以区分接口值的具体类型:")

	values := []any{42, 3.14, "hello", true, nil, dog}
	for _, v := range values {
		typeSwitch(v)
	}

	// ============================================================
	// 6. 接口嵌入
	// ============================================================
	fmt.Println("\n--- 接口嵌入 ---")
	fmt.Println("接口可以嵌入其他接口，组合多个接口的需求:")

	file := SimulatedFile{Name: "data.txt"}
	processFile(file)

	// ============================================================
	// 7. Stringer 接口实用示例
	// ============================================================
	fmt.Println("\n--- Stringer 接口实用示例 ---")
	fmt.Println("实现 fmt.Stringer 接口可以自定义类型的打印格式:")

	// Person 实现了 Stringer
	p := Person{FirstName: "小明", LastName: "张", Age: 28}
	fmt.Printf("  Person: %v\n", p) // fmt 自动调用 String()

	// IPAddr 实现了 Stringer
	ip := IPAddr{127, 0, 0, 1}
	fmt.Printf("  IP: %v\n", ip)

	// Money 实现了 Stringer
	m1 := Money{Amount: 99.99, Currency: "¥"}
	m2 := Money{Amount: 19.99, Currency: "$"}
	fmt.Printf("  Money1: %v\n", m1)
	fmt.Printf("  Money2: %v\n", m2)

	// 在 Sprintf 中也生效
	description := fmt.Sprintf("购买 %s 的商品花费 %v", "电子产品", m1)
	fmt.Printf("  Sprintf 输出: %s\n", description)

	// ============================================================
	// 8. 接口值的底层表示 + nil 陷阱
	// ============================================================
	interfaceInternals()

	// ============================================================
	// 9. 更复杂的接口组合例子
	// ============================================================
	fmt.Println("\n--- 实战：排序接口演示 ---")

	// Go 标准库的 sort.Interface 需要三个方法：
	// Len() int, Less(i, j int) bool, Swap(i, j int)

	type SortablePeople []Person

	people := SortablePeople{
		{FirstName: "明", LastName: "张", Age: 28},
		{FirstName: "红", LastName: "李", Age: 32},
		{FirstName: "强", LastName: "王", Age: 25},
		{FirstName: "芳", LastName: "赵", Age: 30},
	}

	fmt.Println("人员列表 (实现 fmt.Stringer):")
	for _, person := range people {
		fmt.Printf("  %v\n", person)
	}

	// ============================================================
	// 10. 类型转换与接口的关系
	// ============================================================
	fmt.Println("\n--- 类型转换与接口 ---")
	fmt.Println("具体类型 → 接口: 自动（隐式转换）")
	var s Speaker = Dog{Name: "小白"} // 隐式转换
	fmt.Printf("  Speaker 接口持有: %s\n", s.Speak())

	fmt.Println("接口 → 具体类型: 需要类型断言")
	if dog, ok := s.(Dog); ok {
		fmt.Printf("  类型断言成功: %+v\n", dog)
	}

	// 类型转换 int → string（不是接口相关，但常见混淆）
	// 这会将整数解释为 Unicode 码点
	_ = strconv.Itoa(65) // 正确: "65"
	s2 := string(rune(65)) // 正确语法: Unicode 码点 'A'
	fmt.Printf("  注意: string(rune(65)) = %q (是 Unicode 码点，不是数字!)\n", s2)
}