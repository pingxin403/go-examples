// go:build cgo

package main

/*
// 平台相关的 C 代码
// 在 Windows 上使用不同的数学库
#cgo darwin CFLAGS: -I/usr/local/include
#cgo linux CFLAGS: -I/usr/include
#cgo windows CFLAGS: -I../lib/win

// 链接数学库
#cgo linux LDFLAGS: -lm
#cgo darwin LDFLAGS: -framework Accelerate
#cgo windows LDFLAGS: -lm

// 条件编译：不同平台调用不同的 C 函数
#include <stdlib.h>
#include <stdio.h>
#include <math.h>

// 跨平台：获取处理器核心数（简化版）
#ifdef _WIN32
    #include <windows.h>
    int get_cpu_cores() {
        SYSTEM_INFO sysinfo;
        GetSystemInfo(&sysinfo);
        return (int)sysinfo.dwNumberOfProcessors;
    }
#else
    #include <unistd.h>
    int get_cpu_cores() {
        return (int)sysconf(_SC_NPROCESSORS_ONLN);
    }
#endif

// 使用数学库的函数
double compute_sqrt(double x) {
    return sqrt(x);
}
*/
import "C"
import (
	"fmt"
	"runtime"
)

func main() {
	fmt.Println("=== CGO 构建配置与跨平台编译示例 ===")

	// 获取 CPU 核心数（通过 C 调用不同平台的 API）
	cores := C.get_cpu_cores()
	fmt.Printf("CPU 核心数（C 接口）: %d\n", cores)
	fmt.Printf("CPU 核心数（Go 接口）: %d\n", runtime.NumCPU())

	// 使用数学库
	x := C.double(2.0)
	result := C.compute_sqrt(x)
	fmt.Printf("sqrt(2.0) = %.6f\n", result)

	fmt.Println("\n=== 构建命令参考 ===")
	fmt.Println("当前平台:")
	fmt.Printf("  GOOS=%s GOARCH=%s\n", runtime.GOOS, runtime.GOARCH)

	fmt.Println("\n交叉编译到 Linux amd64:")
	fmt.Println(`  CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=x86_64-linux-gnu-gcc go build`)

	fmt.Println("\n交叉编译到 ARM64:")
	fmt.Println(`  CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc go build`)

	fmt.Println("\n静态链接构建:")
	fmt.Println(`  CGO_ENABLED=1 go build -ldflags '-extldflags "-static"'`)

	fmt.Println("\n调试 CGO 编译过程:")
	fmt.Println("  go build -x -v")
}