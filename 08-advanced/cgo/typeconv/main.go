//go:build cgo

package main

/*
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <math.h>

// ---- 枚举 ----
enum Color {
    RED    = 0,
    GREEN  = 1,
    BLUE   = 2,
    YELLOW = 3,
};

// ---- 结构体 ----
typedef struct {
    int x;
    int y;
} Point;

typedef struct {
    char  name[64];
    int   id;
    float salary;
    int   active;
} Employee;

// ---- 联合体 ----
typedef union {
    int   i;
    float f;
    char  s[8];
} Data;

// ---- 嵌套结构体 ----
typedef struct {
    Point topLeft;
    Point bottomRight;
    int   color;
} Rect;

// ---- C 实现的 qsort 比较器 ----
// 注意：Go 导出的比较器函数无法直接作为 C 函数指针使用。
// 因此我们在 C 侧实现比较逻辑，Go 调用 C 封装的 qsort。
int compare_ints(const void* a, const void* b) {
    int ia = *(const int*)a;
    int ib = *(const int*)b;
    if (ia < ib) return -1;
    if (ia > ib) return 1;
    return 0;
}

void sort_ints(int* arr, int len) {
    qsort(arr, len, sizeof(int), compare_ints);
}

// 调用数学库函数
double degrees_to_radians(double deg) {
    return deg * M_PI / 180.0;
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

/*
 * ===========================================
 *  CGO 类型转换示例
 * ===========================================
 *
 * 本示例涵盖 CGO 中各种类型的转换技巧：
 *   1. 基本数值类型映射
 *   2. 结构体（struct）访问
 *   3. 枚举（enum）使用
 *   4. 联合体（union）处理
 *   5. 字符串数组（char**）转换
 *   6. C 的 qsort 调用
 *   7. void* 通用指针
 *   8. const 类型处理
 */

func main() {
	fmt.Println("================================================")
	fmt.Println("  CGO 类型转换示例")
	fmt.Println("================================================")

	// ====================================================
	// 1. 基本数值类型映射
	// ====================================================
	fmt.Println("\n--- 1. 基本数值类型映射 ---")

	var (
		ci  C.int     = 42
		cl  C.long    = 9223372036854775807
		cll C.longlong = 123456789012345
		cf  C.float   = 3.14
		cd  C.double  = 2.718281828459045
		cu  C.uint    = 100
		cub C.uchar   = 'A'
		cus C.ushort  = 65535
	)

	fmt.Printf("C.int(%d)       → int32(%d)\n", int32(ci), int32(ci))
	fmt.Printf("C.long(%d)      → int64(%d)\n", int64(cl), int64(cl))
	fmt.Printf("C.longlong(%d)   → int64(%d)\n", int64(cll), int64(cll))
	fmt.Printf("C.float(%f)     → float32(%f)\n", float32(cf), float32(cf))
	fmt.Printf("C.double(%f)    → float64(%f)\n", float64(cd), float64(cd))
	fmt.Printf("C.uint(%d)      → uint32(%d)\n", uint32(cu), uint32(cu))
	fmt.Printf("C.uchar('%c')   → uint8(%d)\n", byte(cub), byte(cub))
	fmt.Printf("C.ushort(%d)    → uint16(%d)\n", uint16(cus), uint16(cus))

	// 类型映射表
	fmt.Println("\n类型映射表（64 位系统）：")
	fmt.Printf("  %-20s → %-10s (字节数: %d)\n", "C.int", "int32", C.sizeof_int)
	fmt.Printf("  %-20s → %-10s (字节数: %d)\n", "C.uint", "uint32", C.sizeof_uint)
	fmt.Printf("  %-20s → %-10s (字节数: %d)\n", "C.long", "int64", C.sizeof_long)
	fmt.Printf("  %-20s → %-10s (字节数: %d)\n", "C.ulong", "uint64", C.sizeof_ulong)
	fmt.Printf("  %-20s → %-10s (字节数: %d)\n", "C.longlong", "int64", C.sizeof_longlong)
	fmt.Printf("  %-20s → %-10s (字节数: %d)\n", "C.float", "float32", C.sizeof_float)
	fmt.Printf("  %-20s → %-10s (字节数: %d)\n", "C.double", "float64", C.sizeof_double)
	fmt.Printf("  %-20s → %-10s (字节数: %d)\n", "C.char", "byte/int8", C.sizeof_char)
	fmt.Printf("  %-20s → %-10s (字节数: %d)\n", "C.uchar", "uint8", C.sizeof_uchar)
	fmt.Printf("  %-20s → %-10s (字节数: %d)\n", "C.short", "int16", C.sizeof_short)
	fmt.Printf("  %-20s → %-10s (字节数: %d)\n", "C.ushort", "uint16", C.sizeof_ushort)

	// ====================================================
	// 2. C 结构体访问
	// ====================================================
	fmt.Println("\n--- 2. C 结构体访问 ---")

	// 在 Go 中创建和初始化 C 结构体
	p := C.Point{
		x: 10,
		y: 20,
	}
	fmt.Printf("C.Point{x=%d, y=%d}\n", int(p.x), int(p.y))
	fmt.Printf("  C.sizeof_Point = %d 字节\n", C.sizeof_Point)

	// 结构体指针
	pp := &p
	pp.x = 100
	pp.y = 200
	fmt.Printf("After pointer: C.Point{x=%d, y=%d}\n", int(pp.x), int(pp.y))

	// 嵌套结构体
	rect := C.Rect{
		topLeft:     C.Point{x: 0, y: 0},
		bottomRight: C.Point{x: 100, y: 200},
		color:       C.RED,
	}
	fmt.Printf("Rect{topLeft=(%d,%d), bottomRight=(%d,%d), color=%d}\n",
		int(rect.topLeft.x), int(rect.topLeft.y),
		int(rect.bottomRight.x), int(rect.bottomRight.y),
		int(rect.color))
	fmt.Printf("  C.sizeof_Rect = %d 字节\n", C.sizeof_Rect)

	// 复杂结构体（含字符串数组字段）
	var emp C.Employee
	empName := C.CString("John Doe")
	defer C.free(unsafe.Pointer(empName))
	C.strcpy(&emp.name[0], empName)
	emp.id = C.int(1001)
	emp.salary = C.float(75000.50)
	emp.active = C.int(1)

	empNameGo := C.GoString(&emp.name[0])
	fmt.Printf("Employee{name=%q, id=%d, salary=%.2f, active=%d}\n",
		empNameGo, int(emp.id), float32(emp.salary), int(emp.active))
	fmt.Printf("  C.sizeof_Employee = %d 字节\n", C.sizeof_Employee)

	// ====================================================
	// 3. C 枚举使用
	// ====================================================
	fmt.Println("\n--- 3. C 枚举使用 ---")

	colorNames := map[C.enum_Color]string{
		C.RED:    "RED",
		C.GREEN:  "GREEN",
		C.BLUE:   "BLUE",
		C.YELLOW: "YELLOW",
	}

	for c, name := range colorNames {
		fmt.Printf("  C.%s = %d\n", name, int(c))
	}
	fmt.Printf("  C.sizeof_int = %d （枚举底层类型）\n", C.sizeof_int)

	// 枚举在 switch 中的使用
	selectedColor := C.GREEN
	switch selectedColor {
	case C.RED:
		fmt.Println("  选择红色")
	case C.GREEN:
		fmt.Println("  选择绿色")
	case C.BLUE:
		fmt.Println("  选择蓝色")
	case C.YELLOW:
		fmt.Println("  选择黄色")
	}

	// ====================================================
	// 4. C 联合体处理
	// ====================================================
	fmt.Println("\n--- 4. C 联合体处理 ---")

	fmt.Printf("  C.sizeof_Data = %d 字节（联合体大小 = 最大成员大小）\n", C.sizeof_Data)

	// CGO 不直接暴露联合体的字段，必须通过 unsafe 指针访问
	// 联合体所有成员共享同一块内存地址
	var rawData [8]byte // 联合体最大成员对应的大小

	// 写入 int（4 字节）
	*(*C.int)(unsafe.Pointer(&rawData)) = 42
	intVal := *(*C.int)(unsafe.Pointer(&rawData))
	fmt.Printf("  Data as int: %d\n", int(intVal))

	// 写入 float（覆盖同一片内存）
	*(*C.float)(unsafe.Pointer(&rawData)) = 3.14159
	floatVal := *(*C.float)(unsafe.Pointer(&rawData))
	intVal2 := *(*C.int)(unsafe.Pointer(&rawData))
	fmt.Printf("  Data as float: %f — int 已不可读 (now: %d)\n", float32(floatVal), int(intVal2))

	// 直接操作原始字节
	*(*C.int)(unsafe.Pointer(&rawData)) = 0x41424344 // "ABCD"
	// 读出每个字节
	fmt.Printf("  Raw bytes: ")
	for i := 0; i < len(rawData); i++ {
		b := *(*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(&rawData[0])) + uintptr(i)))
		if b >= 32 && b < 127 {
			fmt.Printf("%c", byte(b))
		} else {
			fmt.Printf("\\x%02x", byte(b))
		}
	}
	fmt.Println()

	fmt.Println(`
  联合体注意事项：
    • CGO 不生成联合体字段访问器，需用 unsafe.Pointer
    • 每个字段在同一偏移地址读写（联合体特性）
    • 切换活跃成员会覆盖之前写入的数据
    • 建议用 unsafe 包直接操作内存`)

	// ====================================================
	// 5. 字符串数组（[]string ↔ char**）
	// ====================================================
	fmt.Println("\n--- 5. 字符串数组转换 ---")

	// Go []string → C char**
	names2 := []string{"apple", "banana", "cherry", "date"}
	cNames := make([]*C.char, len(names2))
	for i, s := range names2 {
		cNames[i] = C.CString(s)
		defer C.free(unsafe.Pointer(cNames[i]))
	}

	fmt.Println("  Go → C char** 字符串数组:")
	for i, cs := range cNames {
		fmt.Printf("    [%d] %q\n", i, C.GoString(cs))
	}

	// 用 C 函数读回字符串
	fmt.Print("  C puts each string: ")
	for _, cs := range cNames {
		C.puts(cs)
	}

	// C char** → Go []string
	fmt.Println("  C char** → Go []string:")
	back := make([]string, len(cNames))
	for i, cs := range cNames {
		back[i] = C.GoString(cs)
		fmt.Printf("    [%d] %q\n", i, back[i])
	}

	// ====================================================
	// 6. C qsort 排序
	// ====================================================
	fmt.Println("\n--- 6. C qsort 排序 ---")

	// 使用 C 的 qsort（比较器也在 C 侧实现）
	arr := []C.int{9, 3, 7, 1, 5, 8, 2, 6, 4, 0}
	fmt.Print("  Before C.sort_ints: ")
	for _, v := range arr {
		fmt.Printf("%d ", int(v))
	}
	fmt.Println()

	C.sort_ints(
		(*C.int)(unsafe.Pointer(&arr[0])),
		C.int(len(arr)),
	)

	fmt.Print("  After  C.sort_ints: ")
	for _, v := range arr {
		fmt.Printf("%d ", int(v))
	}
	fmt.Println()

	// ====================================================
	// 7. void* 通用指针
	// ====================================================
	fmt.Println("\n--- 7. void* 通用指针 ---")

	// void* 可以持有任意类型的指针
	var voidPtr unsafe.Pointer

	iv := C.int(999)
	voidPtr = unsafe.Pointer(&iv)
	back2 := *(*C.int)(voidPtr)
	fmt.Printf("  void* → *C.int: %d\n", int(back2))

	doubleVal := C.double(3.14159)
	voidPtr = unsafe.Pointer(&doubleVal)
	backD := *(*C.double)(voidPtr)
	fmt.Printf("  void* → *C.double: %f\n", float64(backD))

	// ====================================================
	// 8. const 类型处理
	// ====================================================
	fmt.Println("\n--- 8. const char* 处理 ---")

	// C 函数的 const char* 参数可以直接用 C.CString 传入
	constMsg := C.CString("this is a const string")
	defer C.free(unsafe.Pointer(constMsg))

	// 调用 C 函数角度转换
	deg := C.double(180.0)
	rad := C.degrees_to_radians(deg)
	fmt.Printf("  C.degrees_to_radians(%f) = %f\n", float64(deg), float64(rad))

	// ====================================================
	// 9. 类型转换注意事项
	// ====================================================
	fmt.Println("\n================================================")
	fmt.Println("  类型转换注意事项")
	fmt.Println("================================================")
	fmt.Println(`
类型转换常见陷阱：

  1. 指针类型不能直接转换
     *C.int 不能直接赋值给 *C.long
     必须通过 unsafe.Pointer 做中转
     正确: (*C.long)(unsafe.Pointer(intPtr))

  2. 结构体布局对齐
     C 和 Go 可能有不同的内存对齐规则
     复杂 C 结构体建议用 C.malloc + unsafe 访问

  3. 枚举的大小
     枚举通常是 C.int 大小，但具体取决于实现
     用 C.sizeof_int 确认

  4. 联合体是 unsafe 操作
     联合体的各成员共享内存
     Go 侧需要自己管理当前活跃的成员

  5. 字符串数组的内存管理
     每个 C.CString 都需要 C.free
     统一管理释放时机，避免泄漏

  6. Go 导出函数不能作为 C 函数指针
     //export 的函数无法被取地址转为 C 函数指针
     C 函数指针需要 C 侧实现（如本例的 compare_ints）

  7. const 在 CGO 中被忽略
     C 的 const char* 在 Go 侧就是 *C.char
     Go 侧没有 const 概念，需自行保证不修改只读数据`)
}