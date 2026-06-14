//go:build cgo

package main

/*
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// 定义一个 C 结构体，演示 Go 函数如何操作 C 数据
typedef struct {
    const char* name;
    int age;
    double score;
} Person;

// C 辅助函数：分配 Person 数组
// 使用 static 避免与 //export 生成的符号冲突
static Person* allocPeople(int count) {
    return (Person*)malloc(sizeof(Person) * count);
}

// C 辅助函数：创建字符串副本
static char* copyStr(const char* s) {
    char* copy = (char*)malloc(strlen(s) + 1);
    strcpy(copy, s);
    return copy;
}

// C 函数：遍历 Person 数组并打印
static void printPeople(Person* people, int count) {
    for (int i = 0; i < count; i++) {
        printf("  C.printPeople: %s (age=%d, score=%.1f)\n",
               people[i].name, people[i].age, people[i].score);
    }
}

// C 函数：整数翻倍
static void doubleInts(int* arr, int len) {
    for (int i = 0; i < len; i++) {
        arr[i] *= 2;
    }
}

// C 函数：计算数组和
static long sumInts(int* arr, int len) {
    long sum = 0;
    for (int i = 0; i < len; i++) {
        sum += arr[i];
    }
    return sum;
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

/*
 * ===========================================
 *  CGO 导出函数示例
 * ===========================================
 *
 * CGO 的 //export 注解可以将 Go 函数导出为 C 符号，
 * 供外部 C/C++ 程序调用。典型用途：
 *
 * 1. 构建共享库 (-buildmode=c-shared)
 *    go build -buildmode=c-shared -o libexport.so .
 *    生成 libexport.so（动态库）和 libexport.h（头文件）
 *    纯 C 程序 #include "libexport.h" 即可使用
 *
 * 2. 构建 C 归档 (-buildmode=c-archive)
 *    go build -buildmode=c-archive -o libexport.a .
 *    生成静态库 + 头文件
 *
 * 3. Go 插件 (-buildmode=c-shared) 嵌入大型 C/C++ 项目
 *
 * 重要：//export 函数在生成 .h 文件后供 C 程序使用。
 * 在同一 Go 包中，//export 函数可以直接作为 Go 函数调用，
 * 但不能通过 C.funcName 引用（CGO 不暴露它们到 Go 的 C 包）。
 *
 * 函数指针回调的正确模式：
 *   Go → C 注册函数指针（用 unsafe 转为函数指针）
 *   C 保存指针 → 后续通过指针调用 Go 函数
 */

//export goGreetPerson
func goGreetPerson(p *C.Person) {
	name := C.GoString(p.name)
	age := int(p.age)
	score := float64(p.score)
	fmt.Printf("  goGreetPerson: %s (age=%d, score=%.1f)\n", name, age, score)
}

//export goDouble
func goDouble(x C.int) C.int {
	return x * 2
}

//export goSumSlice
func goSumSlice(arr *C.int, length C.int) C.long {
	slice := (*[1 << 28]C.int)(unsafe.Pointer(arr))[:length:length]
	var sum C.long
	for _, v := range slice {
		sum += C.long(v)
	}
	return sum
}

func main() {
	fmt.Println("================================================")
	fmt.Println("  CGO 导出函数示例")
	fmt.Println("================================================")

	// ====================================================
	// 1. Go 直接调用 C 函数（不含 //export 回调）
	// ====================================================
	fmt.Println("\n--- 1. Go 调用 C 函数 ---")

	// 在 C 堆上分配 Person 数组
	count := 3
	people := C.allocPeople(C.int(count))
	defer C.free(unsafe.Pointer(people))

	names := []string{"Alice", "Bob", "Charlie"}
	ages := []int{30, 25, 35}
	scores := []float64{95.5, 88.0, 92.3}

	for i := range count {
		p := (*C.Person)(unsafe.Pointer(
			uintptr(unsafe.Pointer(people)) + uintptr(i)*unsafe.Sizeof(C.Person{}),
		))
		p.name = C.copyStr(C.CString(names[i]))
		p.age = C.int(ages[i])
		p.score = C.double(scores[i])
		defer C.free(unsafe.Pointer(p.name))
	}

	// C 函数打印
	fmt.Println("  C.printPeople call:")
	C.printPeople(people, C.int(count))

	// C 函数处理数组
	nums := []C.int{1, 2, 3, 4, 5}
	C.doubleInts((*C.int)(unsafe.Pointer(&nums[0])), C.int(len(nums)))
	fmt.Printf("  C.doubleInts result: ")
	for _, v := range nums {
		fmt.Printf("%d ", int(v))
	}
	fmt.Println()

	// C 函数求和
	s := C.sumInts((*C.int)(unsafe.Pointer(&nums[0])), C.int(len(nums)))
	fmt.Printf("  C.sumInts result: %d\n", int64(s))

	// ====================================================
	// 2. //export 函数：编译为共享库供 C 程序使用
	// ====================================================
	fmt.Println("\n--- 2. //export 函数说明 ---")

	println(`
  下面的 Go 函数通过 //export 导出为 C 符号，
  编译为共享库后，纯 C 程序即可直接调用：

  Go 代码：
    //export goGreetPerson
    func goGreetPerson(p *C.Person) { ... }

    //export goDouble
    func goDouble(x C.int) C.int { return x * 2 }

    //export goSumSlice
    func goSumSlice(arr *C.int, length C.int) C.long { ... }

  C 侧生成的声明 (libexport.h)：
    extern void goGreetPerson(Person* p);
    extern int goDouble(int x);
    extern long goSumSlice(int* arr, int length);

  C 程序调用示例：
    #include "libexport.h"
    int main() {
        Person p = {"Bob", 25, 88.5};
        goGreetPerson(&p);

        int result = goDouble(21);
        printf("goDouble(21) = %d\n", result);

        int arr[] = {10, 20, 30};
        long sum = goSumSlice(arr, 3);
        printf("goSumSlice = %ld\n", sum);
        return 0;
    }
`)

	// ====================================================
	// 3. 构建和链接说明
	// ====================================================
	fmt.Println("--- 3. 构建和链接说明 ---")

	fmt.Print(`
  构建 C 可调用的共享库：

  # 编译
  go build -buildmode=c-shared -o libexport.so .

  # 产物
  #   libexport.so  - 动态链接库
  #   libexport.h   - C 头文件（包含所有 //export 函数声明）

  # 编译 C 消费者程序
  gcc -o myapp myapp.c -L. -lexport

  # 运行
  LD_LIBRARY_PATH=. ./myapp

  # macOS 上可能需要
  DYLD_LIBRARY_PATH=. ./myapp
`)

	// ====================================================
	// 4. 导出函数注意事项
	// ====================================================
	fmt.Println("\n================================================")
	fmt.Println("  导出函数注意事项")
	fmt.Println("================================================")
	fmt.Print(`
导出函数的核心规则和常见错误：

  1. //export 位置
     //export 注解必须紧挨着 import "C"，放在函数定义之前
     错误：将 //export 放在 import "C" 之后

  2. 不要在 preamble 中重复 extern 声明
     CGO 在 _cgo_export.h 中自动生成所有 //export 函数声明
     在 preamble 中写 extern 会导致 "conflicting types" 错误

  3. 类型限制
     导出的函数只能接受和返回 C 兼容类型
     不支持 Go 特有的类型（slice, map, interface, channel 等）

  4. 不能是 variadic 函数
     导出的函数不能使用 ... 变参

  5. 不能从 Go 侧通过 C.funcName 引用
     //export 生成的 C 符号不在 Go 的 C 包作用域内
     Go 代码中直接调用函数名即可（因为函数也是 Go 函数）

  6. 性能开销
     每次跨越 CGO 边界约有 30-100ns 额外开销
     C → CGO bridge → Go runtime 的调用链有固定成本

  7. 多线程安全
     C 侧可能从多个线程调用导出函数
     需要 Go 侧自行保证线程安全（sync.Mutex 等）

  8. 内存管理
     导出函数中分配的资源（C.CString 等）由 Go 侧管理
     C 调用者不应释放 Go 函数返回的 C 内存`)
}