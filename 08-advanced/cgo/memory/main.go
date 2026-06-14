//go:build cgo

package main

/*
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// 将 C 数组的所有元素设为指定值
void fill_ints(int* arr, int len, int value) {
    for (int i = 0; i < len; i++) {
        arr[i] = value;
    }
}

// 打印 C 数组内容
void print_ints(int* arr, int len) {
    printf("[");
    for (int i = 0; i < len; i++) {
        printf("%d", arr[i]);
        if (i < len - 1) printf(", ");
    }
    printf("]\n");
}

// 在 C 堆上分配并初始化一个数组
int* make_array(int len, int init_val) {
    int* arr = (int*)malloc(sizeof(int) * len);
    for (int i = 0; i < len; i++) {
        arr[i] = init_val;
    }
    return arr;
}

// 在 C 堆上分配一个矩阵
int** make_matrix(int rows, int cols, int init_val) {
    int** mat = (int**)malloc(sizeof(int*) * rows);
    for (int i = 0; i < rows; i++) {
        mat[i] = (int*)malloc(sizeof(int) * cols);
        for (int j = 0; j < cols; j++) {
            mat[i][j] = init_val;
        }
    }
    return mat;
}

// 释放 C 堆上分配的矩阵
void free_matrix(int** mat, int rows) {
    for (int i = 0; i < rows; i++) {
        free(mat[i]);
    }
    free(mat);
}

// 从 C 堆复制数据到 Go 提供的缓冲区
void copy_to_buffer(int* src, int* dst, int len) {
    memcpy(dst, src, sizeof(int) * len);
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

/*
 * ===========================================
 *  CGO 内存管理示例
 * ===========================================
 *
 * CGO 的内存管理是 Go 与 C 交互中最容易出错的部分。
 * 核心规则：
 *
 *   1. C 分配的内存必须由 C 释放
 *      - C.CString, C.malloc → 必须 C.free
 *      - 推荐使用 defer 确保释放
 *
 *   2. Go 指针传递给 C 的规则（CGO 规范）：
 *      - Go 指针可以传递给 C 函数，但 C 不能长期持有
 *      - 在 CGO 调用返回后，C 必须停止使用 Go 指针
 *      - 违反规则会导致程序崩溃（Go 的 GC 会移动内存）
 *
 *   3. 大块内存建议在 C 堆上分配
 *      - 避免 Go GC 频繁扫描大对象
 *      - 适合长期持有的数据结构
 *
 *   4. 使用 unsafe.Pointer 作为 Go/C 指针的桥梁
 */

func main() {
	fmt.Println("================================================")
	fmt.Println("  CGO 内存管理示例")
	fmt.Println("================================================")

	// ====================================================
	// 1. C.CString + defer C.free 模式
	// ====================================================
	fmt.Println("\n--- 1. C.CString + defer C.free ---")

	s := C.CString("temporary C string")
	defer C.free(unsafe.Pointer(s))
	C.puts(s)
	fmt.Println("  ✓ C string allocated and freed")

	// 多个字符串的内存管理
	strs := []string{"apple", "banana", "cherry"}
	for _, str := range strs {
		cs := C.CString(str)
		C.puts(cs)
		C.free(unsafe.Pointer(cs)) // 不在 defer 中，立即释放
	}
	fmt.Println("  ✓ Multiple C strings allocated and freed")

	// ====================================================
	// 2. C.malloc / C.free 分配原始内存
	// ====================================================
	fmt.Println("\n--- 2. C.malloc / C.free 原始内存分配 ---")

	// 分配 10 个 int 的空间
	size := C.size_t(10 * C.sizeof_int)
	ptr := C.malloc(size)
	defer C.free(ptr)
	fmt.Printf("  Allocated %d bytes at %v\n", int(size), ptr)

	// 写内存：将整个区域设为 0xFF
	C.memset(ptr, 0xFF, size)

	// 读内存：取回前几个字节
	data := (*[10]C.int)(ptr)
	fmt.Printf("  After memset(0xFF): ")
	for i := 0; i < 10; i++ {
		fmt.Printf("%d ", int(data[i]))
	}
	fmt.Println()

	// ====================================================
	// 3. Go slice → C 数组（传递 Go 内存给 C 函数）
	// ====================================================
	fmt.Println("\n--- 3. Go Slice 传递给 C 函数 ---")

	goSlice := []int32{1, 2, 3, 4, 5}
	fmt.Printf("  Before C.fill_ints: %v\n", goSlice)

	// Go 切片的底层数组指针转为 *C.int 后传给 C 函数
	// 注意：Go 指针可以短时间传递给 C，但 C 不能在 Go 返回后继续使用
	C.fill_ints(
		(*C.int)(unsafe.Pointer(&goSlice[0])),
		C.int(len(goSlice)),
		C.int(99),
	)
	fmt.Printf("  After  C.fill_ints: %v\n", goSlice)

	// C 函数读取并打印 Go 数组
	fmt.Print("  C.print_ints: ")
	C.print_ints(
		(*C.int)(unsafe.Pointer(&goSlice[0])),
		C.int(len(goSlice)),
	)

	// ====================================================
	// 4. C 数组 → Go slice（读取 C 堆上的数据）
	// ====================================================
	fmt.Println("\n--- 4. C 数组转为 Go Slice 读取 ---")

	// 在 C 堆上创建数组
	cArr := C.make_array(10, 42)
	defer C.free(unsafe.Pointer(cArr))

	// 方式一：通过 unsafe 转为固定长度数组
	goArr := (*[10]C.int)(unsafe.Pointer(cArr))
	fmt.Print("  C array via (*[10]C.int): ")
	for _, v := range goArr {
		fmt.Printf("%d ", int(v))
	}
	fmt.Println()

	// 方式二：转为动态 slice（通用模式，适合任意长度）
	length := 10
	goSlice2 := (*[1 << 28]C.int)(unsafe.Pointer(cArr))[:length:length]
	fmt.Print("  C array via dynamic slice: ")
	for _, v := range goSlice2 {
		fmt.Printf("%d ", int(v))
	}
	fmt.Println()

	// C 函数打印
	fmt.Print("  C.print_ints: ")
	C.print_ints(cArr, C.int(length))

	// ====================================================
	// 5. 在 C 堆上分配与释放复杂结构（矩阵）
	// ====================================================
	fmt.Println("\n--- 5. C 堆上分配和释放矩阵 ---")

	rows, cols := 3, 4
	mat := C.make_matrix(C.int(rows), C.int(cols), C.int(7))
	defer C.free_matrix(mat, C.int(rows))

	// 通过 unsafe 读取矩阵数据
	fmt.Println("  C matrix (3x4):")
	for i := 0; i < rows; i++ {
		// 获取第 i 行指针
		rowPtr := *(*unsafe.Pointer)(unsafe.Pointer(
			uintptr(unsafe.Pointer(mat)) + uintptr(i)*unsafe.Sizeof(mat),
		))
		rowSlice := (*[1 << 28]C.int)(rowPtr)[:cols:cols]
		fmt.Print("    ")
		for _, v := range rowSlice {
			fmt.Printf("%4d ", int(v))
		}
		fmt.Println()
	}

	// ====================================================
	// 6. 数据拷贝：C <-> Go 安全传递
	// ====================================================
	fmt.Println("\n--- 6. C <-> Go 数据拷贝 ---")

	// 在 Go 中创建源数据
	src := make([]int32, 10)
	for i := range src {
		src[i] = int32(i * 10)
	}

	// 在 C 堆上分配目标缓冲区
	dst := C.malloc(C.size_t(10 * C.sizeof_int))
	defer C.free(dst)

	// 将 Go slice 数据拷贝到 C 缓冲区
	C.memcpy(dst, unsafe.Pointer(&src[0]), C.size_t(10*C.sizeof_int))

	// 读回验证
	result := (*[10]C.int)(dst)
	fmt.Print("  Copied data (Go→C→Go): ")
	for _, v := range result {
		fmt.Printf("%d ", int(v))
	}
	fmt.Println()

	// ====================================================
	// 7. 内存管理注意事项
	// ====================================================
	fmt.Println("\n================================================")
	fmt.Println("  内存管理注意事项")
	fmt.Println("================================================")
	fmt.Println(`
CGO 内存管理三原则：

原则 1：谁分配谁释放
  • C.CString 分配的 → 用 C.free 释放
  • C.malloc 分配的  → 用 C.free 释放
  • Go 分配的       → Go GC 自动回收

原则 2：Go 指针不可被 C 长期持有
  • 传给 C 函数的 Go 指针只能在函数调用期间使用
  • C 不能保存 Go 指针并在后续调用中继续使用
  • 违反规则通常不会立即报错，而是随机崩溃（GC 移动了内存）

原则 3：大块持久数据建议用 C 堆
  • 避免 Go GC 反复扫描大对象
  • 适用于长期存在的缓存、配置、查找表
  • 必须在程序中正确地释放，否则内存泄漏

常见陷阱：
  • 忘记 C.free → 内存泄漏（valgrind 可以检测）
  • 重复 C.free → double-free 崩溃
  • 传递 Go 指针给 C 后继续使用 → 数据不一致或崩溃
  • C.CString 传入中文等宽字符时，C.sizeof 计算的是字节数而非字符数
  • 多线程环境下 C 释放和 Go 使用之间的竞态条件`)
}