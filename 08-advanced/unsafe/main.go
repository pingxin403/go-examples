package main

import (
	"fmt"
	"reflect"
	"unsafe"
)

/*
 * ===============================
 * Go unsafe 包高级操作示例
 * ===============================
 *
 * ⚠️  安全警告 ⚠️
 * unsafe 包突破了 Go 的类型安全边界，可能导致：
 *   1. 内存损坏（memory corruption）
 *   2. 数据竞争（data race）
 *   3. 不可移植（依赖特定架构的对齐和内存布局）
 *   4. GC 无法正确处理引用（导致 dangling pointer）
 *   5. 编译器优化下的未定义行为
 *
 * 使用原则：不到万不得已不用，用了就要有充分的理由和注释。
 *
 * 适用场景：
 *   - 与 C 代码交互（cgo）
 *   - 极致性能优化的 hot path
 *   - 系统编程（操作系统、内存管理）
 *   - 序列化/反序列化库
 */

// ============ 1. 用于演示的结构体 ============

// Point 二维坐标点，演示结构体内存布局
type Point struct {
	X float64
	Y float64
	Z string
}

// person 小写结构体（未导出），演示访问未导出字段
type person struct {
	name   string
	age    int
	secret string // 未导出且私有
}

// ============ 2. 用于演示字符串/字节转换的类型定义 ============

// stringHeader 模拟 runtime.stringHeader 结构
// 注意：不同 Go 版本可能不同（Go 1.20+ 使用 unsafe.String）
type stringHeader struct {
	Data unsafe.Pointer
	Len  int
}

// sliceHeader 模拟 runtime.sliceHeader 结构
type sliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

func main() {
	fmt.Println("========================================")
	fmt.Println("  unsafe 包操作示例（⚠️  高风险操作）")
	fmt.Println("========================================")

	// ====================================================
	// 1. unsafe.Sizeof / Alignof / Offsetof
	// ====================================================
	fmt.Println("\n--- 1. Sizeof / Alignof / Offsetof ---")

	var i int = 42
	var f float64 = 3.14
	var b bool = true
	var s string = "hello"

	fmt.Printf("int     Sizeof: %d  Alignof: %d\n", unsafe.Sizeof(i), unsafe.Alignof(i))
	fmt.Printf("float64 Sizeof: %d  Alignof: %d\n", unsafe.Sizeof(f), unsafe.Alignof(f))
	fmt.Printf("bool    Sizeof: %d  Alignof: %d\n", unsafe.Sizeof(b), unsafe.Alignof(b))
	fmt.Printf("string  Sizeof: %d  Alignof: %d\n", unsafe.Sizeof(s), unsafe.Alignof(s))

	// 结构体内存布局分析
	p := Point{X: 1.0, Y: 2.0, Z: "point"}
	fmt.Printf("\nPoint 结构体总大小: %d 字节\n", unsafe.Sizeof(p))
	fmt.Printf("  X (float64) 偏移: %d, 大小: %d\n",
		unsafe.Offsetof(p.X), unsafe.Sizeof(p.X))
	fmt.Printf("  Y (float64) 偏移: %d, 大小: %d\n",
		unsafe.Offsetof(p.Y), unsafe.Sizeof(p.Y))
	fmt.Printf("  Z (string)  偏移: %d, 大小: %d\n",
		unsafe.Offsetof(p.Z), unsafe.Sizeof(p.Z))

	// 观察内存对齐填充（padding）
	fmt.Println("\n  ⚠️  注意：结构体可能有内存对齐填充，不同架构结果不同")

	// ====================================================
	// 2. float64 → uint64 位转换（不改变二进制位模式）
	// ====================================================
	fmt.Println("\n--- 2. float64 ↔ uint64 位转换 ---")

	var pi float64 = 3.141592653589793
	bits := *(*uint64)(unsafe.Pointer(&pi))
	fmt.Printf("float64 %v 的位表示为: 0x%016x\n", pi, bits)

	// 还原
	restored := *(*float64)(unsafe.Pointer(&bits))
	fmt.Printf("从 uint64 还原为 float64: %v\n", restored)
	fmt.Println("  应用：浮点数位级操作、NaN 检测、hash 计算")

	// ====================================================
	// 3. 零拷贝字符串 ↔ []byte 转换
	// ====================================================
	fmt.Println("\n--- 3. 零拷贝 string ↔ []byte 转换 ---")

	original := "Hello, 不安全的世界！"
	fmt.Printf("原始字符串: %q (长度 %d)\n", original, len(original))

	// 字符串 → []byte（零拷贝！）
	// 原理：string 和 []byte 在内存中都是 [ptr, len] 的布局
	// 通过 unsafe 直接复用底层的字节数组，不分配新内存
	strHdr := (*stringHeader)(unsafe.Pointer(&original))
	byteSlice := *(*[]byte)(unsafe.Pointer(&sliceHeader{
		Data: strHdr.Data,
		Len:  strHdr.Len,
		Cap:  strHdr.Len,
	}))

	fmt.Printf("零拷贝转 []byte: %v\n", byteSlice)
	fmt.Println("  ⚠️  修改 byteSlice 会修改原字符串！")

	// 演示：修改字节切片会影响原字符串（⚠️ 危险操作）
	// 注意：Go 中字符串字面量可能存储在只读段，修改会导致 crash！
	// 这里我们用一个变量字符串而不是字面量
	mutableStr := string([]byte("mutable string example"))
	mutHdr := (*stringHeader)(unsafe.Pointer(&mutableStr))
	mutSlice := *(*[]byte)(unsafe.Pointer(&sliceHeader{
		Data: mutHdr.Data,
		Len:  mutHdr.Len,
		Cap:  mutHdr.Len,
	}))
	mutSlice[0] = 'M' // ⚠️  修改字符串内容
	fmt.Printf("修改后 mutableStr: %q\n", mutableStr)

	fmt.Println("\n  🏁 应用：高性能字符串处理，避免大量 []byte(str) 分配")
	fmt.Println("  ⚠️  风险：违反字符串不可变性语义，GC 可能无法正确处理")

	// ====================================================
	// 4. 访问未导出（私有）结构体字段
	// ====================================================
	fmt.Println("\n--- 4. 访问未导出结构体字段 ---")

	alice := person{
		name:   "Alice",
		age:    30,
		secret: "这是秘密信息",
	}
	fmt.Printf("person 结构体: %+v\n", alice)

	// 正常情况下无法访问未导出字段：
	// fmt.Println(alice.secret) // 如果在不同包会编译错误

	// 但通过 unsafe 可以绕过访问限制：
	// 方法1：通过 unsafe.Pointer + 偏移量
	// 需要知道字段的偏移（依赖 struct 布局，不可靠！）
	fmt.Println("\n  通过偏移量访问（方法1）:")
	ptr := unsafe.Pointer(&alice)

	// name 字段在偏移 0
	namePtr := (*string)(ptr)
	fmt.Printf("    name = %q\n", *namePtr)

	// age 字段：在 string 之后（string = 16 bytes on 64-bit）
	agePtr := (*int)(unsafe.Pointer(uintptr(ptr) + unsafe.Sizeof(string(""))))
	fmt.Printf("    age = %d\n", *agePtr)

	// secret 字段：在 age 之后（int 大小取决于架构，64位一般是8字节）
	stringSize := unsafe.Sizeof(string(""))
	intSize := unsafe.Sizeof(int(0))
	secretPtr := (*string)(unsafe.Pointer(uintptr(ptr) + stringSize + intSize))
	fmt.Printf("    secret = %q\n", *secretPtr)

	// 方法2：通过 reflect + unsafe（更可控但更绕）
	fmt.Println("\n  通过 reflect + unsafe 访问（方法2）:")
	pv := reflect.ValueOf(&alice).Elem()
	// 虽然字段未导出，但 Type 信息仍然包含它
	pt := pv.Type()
	for i := 0; i < pt.NumField(); i++ {
		field := pt.Field(i)
		// 通过 unsafe 绕过 reflect 的 CanInterface 检查
		fv := pv.Field(i)
		// 对于未导出字段，reflect 会 panic，但我们可以读取底层数据
		// 使用 reflect.NewAt + Interface 的技巧
		fieldAddr := unsafe.Pointer(fv.UnsafeAddr())
		var val interface{}
		switch field.Type.Kind() {
		case reflect.String:
			val = *(*string)(fieldAddr)
		case reflect.Int:
			val = *(*int)(fieldAddr)
		}
		fmt.Printf("    %s = %v\n", field.Name, val)
	}

	fmt.Println("\n  ⚠️  风险：")
	fmt.Println("    1. 依赖结构体内存布局，Go 版本升级可能改变布局")
	fmt.Println("    2. GC 可能移动对象，uintptr 会失效")
	fmt.Println("    3. 违反封装性，破坏模块设计")
	fmt.Println("    4. 编译器优化可能导致未定义行为")

	// ====================================================
	// 5. 指针运算（类似 C 语言的指针算术）
	// ====================================================
	fmt.Println("\n--- 5. 指针运算（⚠️  高风险） ---")

	arr := [5]int{10, 20, 30, 40, 50}
	fmt.Printf("数组: %v\n", arr)

	// 获取第一个元素的指针
	first := unsafe.Pointer(&arr[0])
	fmt.Printf("arr[0] 地址: %v\n", first)

	// 通过指针运算访问第三个元素
	// int 大小：64位系统通常是 8 字节
	elemSize := unsafe.Sizeof(arr[0])
	third := (*int)(unsafe.Pointer(uintptr(first) + 2*elemSize))
	fmt.Printf("arr[2] (通过指针运算): %d\n", *third)

	// 修改第五个元素
	fifth := (*int)(unsafe.Pointer(uintptr(first) + 4*elemSize))
	*fifth = 99
	fmt.Printf("修改后数组: %v\n", arr)

	fmt.Println("\n  ⚠️  风险：")
	fmt.Println("    1. 越界访问没有保护（不像 slice 有 bounds check）")
	fmt.Println("    2. uintptr 是整数，GC 不会跟踪它指向的对象")
	fmt.Println("    3. 如果对象在 uintptr 计算期间被 GC 移动，结果是野指针")

	// ====================================================
	// 6. 与 cgo 配合：C 内存操作模拟
	// ====================================================
	fmt.Println("\n--- 6. 内存读写操作 ---")

	// 分配一块内存（类似 C 的 malloc）
	// 使用 make 创建一个 slice 作为底层存储
	data := make([]byte, 64)
	dataPtr := unsafe.Pointer(&data[0])

	// 在这块内存中写入不同类型的数据
	*(*int32)(dataPtr) = 0x12345678                     // 写入 int32
	*(*float64)(unsafe.Pointer(uintptr(dataPtr) + 8)) = 3.14 // 偏移 8 写入 float64
	*(*[4]byte)(unsafe.Pointer(uintptr(dataPtr) + 16)) = [4]byte{'G', 'o', '!', 0} // 偏移 16 写入字符串

	fmt.Printf("字节数组前 24 字节: % x\n", data[:24])
	fmt.Println("  应用：序列化/反序列化、网络协议解析、共享内存")

	fmt.Println("\n========================================")
	fmt.Println("  安全总结")
	fmt.Println("========================================")
	fmt.Println("1. 优先使用标准库的 encoding/binary、reflect 等安全手段")
	fmt.Println("2. 如果必须用 unsafe：")
	fmt.Println("   a. 充分注释为什么非用不可")
	fmt.Println("   b. 用 unsafe.Pointer 而非 uintptr")
	fmt.Println("   c. 避免 uintptr 跨越 GC 安全点")
	fmt.Println("   d. 加上 go:nocheckptr 要非常谨慎")
	fmt.Println("   e. 编写充分的测试，使用 -race 检测")
	fmt.Println("3. Go 1.x 兼容性承诺不包括 unsafe 的行为不变性")
}