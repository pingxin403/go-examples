//go:build cgo

package main

/*
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

typedef struct {
    const char* name;
    int age;
    double score;
} TestPerson;

static TestPerson* test_allocPeople(int count) {
    return (TestPerson*)malloc(sizeof(TestPerson) * count);
}

static char* test_copyStr(const char* s) {
    char* copy = (char*)malloc(strlen(s) + 1);
    strcpy(copy, s);
    return copy;
}

static void test_doubleInts(int* arr, int len) {
    for (int i = 0; i < len; i++) {
        arr[i] *= 2;
    }
}

static long test_sumInts(int* arr, int len) {
    long sum = 0;
    for (int i = 0; i < len; i++) {
        sum += arr[i];
    }
    return sum;
}
*/
import "C"
import (
	"unsafe"
)

// cgoDoubleInts 包装 C.doubleInts，对 []int 执行翻倍
func cgoDoubleInts(slice []int) []int {
	if len(slice) == 0 {
		return slice
	}
	cSlice := make([]C.int, len(slice))
	for i, v := range slice {
		cSlice[i] = C.int(v)
	}
	C.test_doubleInts((*C.int)(unsafe.Pointer(&cSlice[0])), C.int(len(cSlice)))
	result := make([]int, len(slice))
	for i, v := range cSlice {
		result[i] = int(v)
	}
	return result
}

// cgoSumInts 包装 C.sumInts
func cgoSumInts(slice []int) int64 {
	if len(slice) == 0 {
		return 0
	}
	cSlice := make([]C.int, len(slice))
	for i, v := range slice {
		cSlice[i] = C.int(v)
	}
	sum := C.test_sumInts((*C.int)(unsafe.Pointer(&cSlice[0])), C.int(len(cSlice)))
	return int64(sum)
}

// cgoAllocPeople 在 C 堆上分配 TestPerson 数组并验证数据
func cgoAllocPeople(names []string, ages []int, scores []float64) bool {
	count := len(names)
	people := C.test_allocPeople(C.int(count))
	defer C.free(unsafe.Pointer(people))

	for i := 0; i < count; i++ {
		p := (*C.TestPerson)(unsafe.Pointer(
			uintptr(unsafe.Pointer(people)) + uintptr(i)*unsafe.Sizeof(C.TestPerson{}),
		))
		p.name = C.test_copyStr(C.CString(names[i]))
		p.age = C.int(ages[i])
		p.score = C.double(scores[i])
		defer C.free(unsafe.Pointer(p.name))
	}

	for i := 0; i < count; i++ {
		p := (*C.TestPerson)(unsafe.Pointer(
			uintptr(unsafe.Pointer(people)) + uintptr(i)*unsafe.Sizeof(C.TestPerson{}),
		))
		if C.GoString(p.name) != names[i] {
			return false
		}
		if int(p.age) != ages[i] {
			return false
		}
	}
	return true
}

// cgoSumGoSlice 通过 C 函数求和 []int
func cgoSumGoSlice(slice []int) int64 {
	if len(slice) == 0 {
		return 0
	}
	cSlice := make([]C.int, len(slice))
	for i, v := range slice {
		cSlice[i] = C.int(v)
	}
	// 直接调用 C.test_sumInts 而非 goSumSlice（goSumSlice 是 //export 但可直接调用）
	sum := C.test_sumInts((*C.int)(unsafe.Pointer(&cSlice[0])), C.int(len(cSlice)))
	return int64(sum)
}

// cgoDoubleGoValue 测试 goDouble（//export）函数
func cgoDoubleGoValue(x int) int {
	return int(goDouble(C.int(x)))
}

// cgoChainedOperation 测试 doubleInts + sumInts 链式调用
func cgoChainedOperation(slice []int) int64 {
	doubled := cgoDoubleInts(slice)
	return cgoSumInts(doubled)
}