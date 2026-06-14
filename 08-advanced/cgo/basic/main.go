//go:build cgo

package main

/*
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// 自定义 C 函数：向标准输出打印问候语
void greet(const char* name) {
    printf("Hello from C, %s!\n", name);
}

// 自定义 C 函数：两个整数相加
int add(int a, int b) {
    return a + b;
}

// 自定义 C 函数：字符串长度（模拟 strlen）
int str_len(const char* s) {
    return (int)strlen(s);
}

// 自定义 C 函数：拼接两个字符串并打印
void concat_and_print(const char* a, const char* b) {
    char buf[256];
    snprintf(buf, sizeof(buf), "%s %s", a, b);
    printf("C concat: %s\n", buf);
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

/*
 * ===========================================
 *  CGO 基础示例
 * ===========================================
 *
 * CGO（C Go）是 Go 调用 C 代码的桥梁。通过 import "C" 和注释中的 C 代码，
 * Go 程序可以直接调用 C 函数、使用 C 类型、操作 C 内存。
 *
 * 核心概念：
 *   - import "C" 前面的注释块是 C 代码（preamble）
 *   - C.TYPE 映射：C.int → int32, C.long → int64（64位）, C.double → float64
 *   - C.CString / C.GoString 用于 Go string ↔ C string 转换
 *   - C 分配的内存必须由 C.free 释放
 *   - C.sizeof_TYPE 获取 C 类型大小
 *
 * 构建条件：
 *   CGO_ENABLED=1（默认启用），需要系统安装 C 编译器（gcc/clang）
 *   本文件使用 //go:build cgo 构建标签确保只在 CGO 开启时编译
 */

func main() {
	fmt.Println("================================================")
	fmt.Println("  CGO 基础示例")
	fmt.Println("================================================")

	// ====================================================
	// 1. 调用 C 标准库函数（puts）
	// ====================================================
	fmt.Println("\n--- 1. 调用 C 标准库函数 ---")
	cs := C.CString("Hello from CGO!")
	defer C.free(unsafe.Pointer(cs))
	C.puts(cs)

	// ====================================================
	// 2. 调用自定义 C 函数
	// ====================================================
	fmt.Println("\n--- 2. 调用自定义 C 函数 ---")

	name := C.CString("Gopher")
	defer C.free(unsafe.Pointer(name))
	C.greet(name)

	// 带返回值的自定义函数
	sum := C.add(3, 4)
	fmt.Printf("C.add(3, 4) = %d (Go type: %T)\n", int(sum), sum)

	// 字符串长度
	text := C.CString("CGO")
	defer C.free(unsafe.Pointer(text))
	length := C.str_len(text)
	fmt.Printf("C.str_len(\"CGO\") = %d\n", int(length))

	// ====================================================
	// 3. 类型映射
	// ====================================================
	fmt.Println("\n--- 3. 类型映射 ---")

	var a C.int = 42
	var b C.double = 3.14159
	var c C.long = 9223372036854775807
	var d C.char = 'A'
	var e C.uint = 100
	var f C.float = 2.71828

	fmt.Printf("C.int(%d)    → Go int32  (%T)\n", int32(a), a)
	fmt.Printf("C.long(%d)   → Go int64  (%T) (64 位系统)\n", int64(c), c)
	fmt.Printf("C.uint(%d)   → Go uint32 (%T)\n", uint32(e), e)
	fmt.Printf("C.float(%f)  → Go float32(%T)\n", float32(f), f)
	fmt.Printf("C.double(%f) → Go float64(%T)\n", float64(b), b)
	fmt.Printf("C.char('%c') → Go byte   (%T)\n", byte(d), d)

	// ====================================================
	// 4. C.sizeof 获取类型大小
	// ====================================================
	fmt.Println("\n--- 4. C.sizeof ---")

	fmt.Printf("C.sizeof_int    = %d 字节\n", C.sizeof_int)
	fmt.Printf("C.sizeof_long   = %d 字节\n", C.sizeof_long)
	fmt.Printf("C.sizeof_double = %d 字节\n", C.sizeof_double)
	fmt.Printf("C.sizeof_float  = %d 字节\n", C.sizeof_float)
	fmt.Printf("C.sizeof_char   = %d 字节\n", C.sizeof_char)
	fmt.Printf("C.sizeof_void   = %d 字节\n", C.sizeof_void)
	fmt.Printf("C.sizeof_uint   = %d 字节\n", C.sizeof_uint)

	// ====================================================
	// 5. 字符串转换：C.CString / C.GoString
	// ====================================================
	fmt.Println("\n--- 5. 字符串转换 ---")

	// Go string → C string
	goStr := "Hello, 世界"
	cs2 := C.CString(goStr)
	defer C.free(unsafe.Pointer(cs2))

	// C string → Go string（完整转换）
	backToGo := C.GoString(cs2)
	fmt.Printf("C.GoString: %q\n", backToGo)

	// C string → Go string（指定长度）
	cs3 := C.CString("Hello CGO World")
	defer C.free(unsafe.Pointer(cs3))
	part := C.GoStringN(cs3, 5)
	fmt.Printf("C.GoStringN(first 5): %q\n", part)

	// 拼接演示
	C.concat_and_print(
		C.CString("Hello"),
		C.CString("CGO"),
	)

	// ====================================================
	// 6. 数值运算：混合使用 Go 和 C 的数值类型
	// ====================================================
	fmt.Println("\n--- 6. 数值运算 ---")

	x := C.int(10)
	y := C.int(20)
	z := C.add(x, y)
	fmt.Printf("C.add(C.int(%d), C.int(%d)) = %d\n", int(x), int(y), int(z))

	// 类型转换后参与 Go 运算
	goX := int(x)
	goY := int(y)
	fmt.Printf("After Go conversion: %d + %d = %d\n", goX, goY, goX+goY)

	// ====================================================
	// 7. 指针和地址
	// ====================================================
	fmt.Println("\n--- 7. 指针和地址 ---")

	val := C.int(99)
	ptr := &val
	fmt.Printf("C.int value: %d, address: %v\n", *ptr, ptr)

	// unsafe.Pointer 是 CGO 中 Go 指针 ↔ C 指针的桥梁
	voidPtr := unsafe.Pointer(ptr)
	fmt.Printf("unsafe.Pointer: %v\n", voidPtr)

	fmt.Println("\n================================================")
	fmt.Println("  总结")
	fmt.Println("================================================")
	fmt.Println(`
CGO 核心 API：

  类型转换：
    C.int, C.long, C.double, C.char, C.uint, C.float
    → 对应 Go 的 int32, int64, float64, byte, uint32, float32

  字符串转换：
    C.CString(goStr)   → *C.char（分配 C 内存，需要 defer C.free）
    C.GoString(cStr)    → string（复制到 Go 堆）
    C.GoStringN(cStr,n) → string（指定字节数）

  内存管理：
    C.malloc(size)      → 分配 C 内存
    C.free(ptr)         → 释放 C 内存
    unsafe.Pointer      → 两种指针类型的桥梁

  注意事项：
    • C.CString 分配的 C 内存必须由 C.free 释放（最好 defer）
    • CGO 调用有少量开销，高频调用应考虑批处理
    • C 代码中的 static 变量在每个 CGO 调用间保持状态
    • Go 指针不能长期被 C 持有（只能短暂传递给 C 函数）
    • CGO 开启后无法交叉编译（除非设置 CC 环境变量）`)
}