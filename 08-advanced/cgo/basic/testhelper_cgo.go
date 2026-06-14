//go:build cgo

package main

/*
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// 静态函数避免与 main.go 中的 C 代码产生链接冲突

static int test_add(int a, int b) {
    return a + b;
}

static int test_str_len(const char* s) {
    return (int)strlen(s);
}

static void test_greet(const char* name) {
    printf("Hello from C test, %s!\n", name);
}
*/
import "C"
import "unsafe"

// cgoAdd 包装 C.add 供测试使用
func cgoAdd(a, b int) int {
	return int(C.test_add(C.int(a), C.int(b)))
}

// cgoStrLen 包装 C.str_len 供测试使用
func cgoStrLen(s string) int {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	return int(C.test_str_len(cs))
}

// cgoGreet 调用 C.greet，仅测试不 panic
func cgoGreet(name string) {
	cs := C.CString(name)
	defer C.free(unsafe.Pointer(cs))
	C.test_greet(cs)
}