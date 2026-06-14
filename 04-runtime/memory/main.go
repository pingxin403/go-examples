// memory/main.go — Go 内存管理核心概念演示
//
// 展示栈 vs 堆分配、make vs new、对象大小与对齐、
// 内存分配追踪、uintptr 与 unsafe.Pointer 基础。
package main

import (
	"fmt"
	"runtime"
	"unsafe"
)

func main() {
	// ============================================================
	// 1. 栈 vs 堆分配（通过逃逸分析注释）
	// ============================================================
	fmt.Println("=== 1. 栈 vs 堆分配（逃逸分析） ===")

	// 在 main 中定义，值不逃逸 → 栈分配
	a := 42
	fmt.Printf("a = %d, 地址: %p (大概率在栈上)\n", a, &a)

	// 返回值指针 → 逃逸到堆
	p := createInt()
	fmt.Printf("createInt() 返回 *int = %d, 地址: %p (在堆上，因为逃逸了)\n", *p, p)

	// 大数组 → 栈上放不下也会逃逸
	large := createLargeArray()
	fmt.Printf("large array 长度: %d, 地址: %p (大对象逃逸到堆)\n", len(large), &large[0])

	// 接口类型存储 — 常导致逃逸
	var iface interface{} = 42
	fmt.Printf("interface{} 装箱的 int = %v (接口装箱常导致堆分配)\n", iface)

	// 闭包捕获变量 → 逃逸到堆
	fn := createClosure()
	fmt.Printf("闭包返回: %d (闭包捕获的变量逃逸到堆)\n", fn())

	// 查看逃逸分析结果：go build -gcflags="-m" 04-runtime/memory/main.go
	fmt.Println("\n提示: 用以下命令查看逃逸分析:")
	fmt.Println("  go build -gcflags=\"-m\" ./04-runtime/memory/")

	// ============================================================
	// 2. make vs new
	// ============================================================
	fmt.Println("\n=== 2. make vs new ===")

	// new(T) — 返回 *T，零值初始化
	np := new(int)
	fmt.Printf("new(int): *np = %d, 类型: %T\n", *np, np)

	nps := new(string)
	fmt.Printf("new(string): *nps = %q, 类型: %T\n", *nps, nps)

	npsl := new([]int)    // 注意：new 返回 nil slice
	fmt.Printf("new([]int): %v, len=%d, cap=%d (nil slice!)\n", *npsl, len(*npsl), cap(*npsl))

	// make(T, ...) — 只用于 slice/map/channel，返回 T（非指针）
	sl := make([]int, 5, 10)
	fmt.Printf("make([]int, 5, 10): %v, len=%d, cap=%d\n", sl, len(sl), cap(sl))

	m := make(map[string]int, 10)
	m["hello"] = 42
	fmt.Printf("make map: m[\"hello\"] = %d\n", m["hello"])

	ch := make(chan int, 3)
	ch <- 1
	fmt.Printf("make chan: cap=%d, len=%d\n", cap(ch), len(ch))

	// 等价关系：make(T, n) 相当于 new(T) + 初始化
	fmt.Println("\n--- new vs make 对比 ---")
	fmt.Println("new(T)     → 返回 *T，内存零值初始化，适用于任何类型")
	fmt.Println("make(T,n)  → 返回 T（初始化后的值），只适用于 slice/map/channel")
	fmt.Println("new([]int) 是 nil slice，make([]int,5) 是有底层数组的 slice")

	// ============================================================
	// 3. 对象大小与对齐
	// ============================================================
	fmt.Println("\n=== 3. 对象大小与对齐 ===")

	// 基本类型大小
	fmt.Printf("int       大小: %d 字节\n", unsafe.Sizeof(int(0)))
	fmt.Printf("int32     大小: %d 字节\n", unsafe.Sizeof(int32(0)))
	fmt.Printf("int64     大小: %d 字节\n", unsafe.Sizeof(int64(0)))
	fmt.Printf("float64   大小: %d 字节\n", unsafe.Sizeof(float64(0)))
	fmt.Printf("string    大小: %d 字节 (16 字节 = 指针 + 长度)\n", unsafe.Sizeof(""))
	fmt.Printf("bool      大小: %d 字节\n", unsafe.Sizeof(true))
	fmt.Printf("byte      大小: %d 字节\n", unsafe.Sizeof(byte(0)))

	// 结构体大小与对齐
	type Small struct {
		A bool  // 1 字节
		B int32 // 4 字节（需要 4 字节对齐，偏移 4）
	}
	// bool(1) + 3 padding + int32(4) = 8
	fmt.Printf("Small{A bool; B int32}   大小: %d 字节 (有填充)\n", unsafe.Sizeof(Small{}))

	type BadAligned struct {
		A bool   // 1 字节
		B int64  // 8 字节 -> 需要偏移 8，前面填充 7 字节
		C bool   // 1 字节
		// 尾部填充 7 字节
	}
	// 1 + 7pad + 8 + 1 + 7pad = 24
	fmt.Printf("BadAligned{A bool; B int64; C bool}  大小: %d 字节 (有大量填充)\n", unsafe.Sizeof(BadAligned{}))

	type GoodAligned struct {
		B int64 // 8 字节（最早放最大字段）
		A bool  // 1 字节
		C bool  // 1 字节
		// 尾部填充 6 字节
	}
	// 8 + 1 + 1 + 6pad = 16
	fmt.Printf("GoodAligned{B int64; A bool; C bool}  大小: %d 字节 (更紧凑)\n", unsafe.Sizeof(GoodAligned{}))

	fmt.Println("\n优化提示: 结构体字段按大小降序排列可减少填充字节")

	// ============================================================
	// 4. runtime.MemStats 追踪内存分配
	// ============================================================
	fmt.Println("\n=== 4. MemStats 追踪内存分配 ===")

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	startAlloc := stats.TotalAlloc

	// 分配一系列对象，跟踪分配量
	allocs := make([][]byte, 100)
	for i := range allocs {
		allocs[i] = make([]byte, 1024) // 每个 1KB
	}
	runtime.ReadMemStats(&stats)
	fmt.Printf("分配 100 个 1KB slice 后 TotalAlloc 增量: ~%d KB\n",
		(stats.TotalAlloc-startAlloc)/1024)

	// 观察 Alloc（当前在用）vs TotalAlloc（累计分配）
	fmt.Printf("当前堆使用 (Alloc):     %d KB\n", stats.Alloc/1024)
	fmt.Printf("累计分配 (TotalAlloc):  %d KB\n", stats.TotalAlloc/1024)
	fmt.Printf("系统占用 (Sys):         %d KB\n", stats.Sys/1024)

	// ============================================================
	// 5. uintptr vs unsafe.Pointer 基础
	// ============================================================
	fmt.Println("\n=== 5. uintptr vs unsafe.Pointer ===")

	type Point struct {
		X int64
		Y int64
	}

	p2 := Point{X: 10, Y: 20}

	// unsafe.Pointer: 可以转换为任意指针类型，GC 能感知
	ptr := unsafe.Pointer(&p2)
	intPtr := (*int64)(ptr)
	fmt.Printf("unsafe.Pointer 转换为 *int64: *intPtr = %d (对应 Point.X)\n", *intPtr)

	// 指针运算：通过偏移量访问结构体字段
	// unsafe.Offsetof 获取字段偏移
	yOffset := unsafe.Offsetof(p2.Y)
	yPtr := (*int64)(unsafe.Add(ptr, yOffset)) // Go 1.17+ unsafe.Add
	fmt.Printf("通过偏移量访问 Point.Y: %d (unsafe.Add(base, %d))\n", *yPtr, yOffset)

	// uintptr: 整型，GC 不追踪，不能直接解引用
	addr := uintptr(ptr)
	fmt.Printf("uintptr 只是整数: 0x%x (GC 不追踪，指向的对象可能被回收)\n", addr)

	fmt.Println("\n--- uintptr 危险示例 ---")
	fmt.Println("uintptr 不持有 GC 引用，如果对象移动或回收，uintptr 变成野指针!")
	fmt.Println("正确做法: 只在表达式内使用 uintptr，不长期保存")

	// unsafe.Sizeof/Offsetof/Alignof 是编译器常量，安全使用
	fmt.Printf("\n编译时常量 (安全):\n")
	fmt.Printf("  Point.X 偏移: %d\n", unsafe.Offsetof(Point{}.X))
	fmt.Printf("  Point.Y 偏移: %d\n", unsafe.Offsetof(Point{}.Y))
	fmt.Printf("  Point 对齐:   %d\n", unsafe.Alignof(Point{}))
	fmt.Printf("  int64 对齐:   %d\n", unsafe.Alignof(int64(0)))

	fmt.Println("\n内存管理概念演示完毕 ✓")
}

// createInt — 返回指针，变量逃逸到堆
func createInt() *int {
	x := 42 // x 本应在栈上，但被返回引用，逃逸到堆
	return &x
}

func createLargeArray() [1024]int {
	var arr [1024]int // 大数组逃逸到堆
	return arr
}

// createClosure — 闭包捕获变量，变量逃逸到堆
func createClosure() func() int {
	x := 100 // 被闭包捕获，逃逸到堆
	return func() int {
		x++ // 闭包引用 x
		return x
	}
}