//go:build cgo

package main

/*
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static void test_fill_ints(int* arr, int len, int value) {
    for (int i = 0; i < len; i++) {
        arr[i] = value;
    }
}

static int* test_make_array(int len, int init_val) {
    int* arr = (int*)malloc(sizeof(int) * len);
    for (int i = 0; i < len; i++) {
        arr[i] = init_val;
    }
    return arr;
}

static int** test_make_matrix(int rows, int cols, int init_val) {
    int** mat = (int**)malloc(sizeof(int*) * rows);
    for (int i = 0; i < rows; i++) {
        mat[i] = (int*)malloc(sizeof(int) * cols);
        for (int j = 0; j < cols; j++) {
            mat[i][j] = init_val;
        }
    }
    return mat;
}

static void test_free_matrix(int** mat, int rows) {
    for (int i = 0; i < rows; i++) {
        free(mat[i]);
    }
    free(mat);
}
*/
import "C"
import (
	"unsafe"
)

// cgoFillInts 使用 C 函数填充 Go 切片
func cgoFillInts(slice []int32, value int32) {
	if len(slice) == 0 {
		return
	}
	C.test_fill_ints(
		(*C.int)(unsafe.Pointer(&slice[0])),
		C.int(len(slice)),
		C.int(value),
	)
}

// cgoMakeArray 在 C 堆上创建数组并返回 Go 切片
func cgoMakeArray(n int, initVal int) []int32 {
	cArr := C.test_make_array(C.int(n), C.int(initVal))
	defer C.free(unsafe.Pointer(cArr))

	result := make([]int32, n)
	for i := 0; i < n; i++ {
		val := *(*C.int)(unsafe.Pointer(uintptr(unsafe.Pointer(cArr)) + uintptr(i)*C.sizeof_int))
		result[i] = int32(val)
	}
	return result
}

// cgoMemcpyFromGo 将 Go 数据拷贝到 C 堆再读回
func cgoMemcpyFromGo(src []int32) []int32 {
	if len(src) == 0 {
		return []int32{}
	}
	dst := C.malloc(C.size_t(len(src) * int(C.sizeof_int)))
	defer C.free(dst)

	C.memcpy(dst, unsafe.Pointer(&src[0]), C.size_t(len(src)*int(C.sizeof_int)))

	result := make([]int32, len(src))
	for i := 0; i < len(src); i++ {
		val := *(*C.int)(unsafe.Pointer(uintptr(dst) + uintptr(i)*C.sizeof_int))
		result[i] = int32(val)
	}
	return result
}

// cgoMatrixAlloc 分配矩阵并用 C 函数验证
func cgoMatrixAlloc(rows, cols, initVal int) [][]int32 {
	mat := C.test_make_matrix(C.int(rows), C.int(cols), C.int(initVal))
	defer C.test_free_matrix(mat, C.int(rows))

	result := make([][]int32, rows)
	for i := 0; i < rows; i++ {
		rowPtr := *(*unsafe.Pointer)(unsafe.Pointer(
			uintptr(unsafe.Pointer(mat)) + uintptr(i)*unsafe.Sizeof(mat),
		))
		result[i] = make([]int32, cols)
		for j := 0; j < cols; j++ {
			val := *(*C.int)(unsafe.Pointer(uintptr(rowPtr) + uintptr(j)*C.sizeof_int))
			result[i][j] = int32(val)
		}
	}
	return result
}