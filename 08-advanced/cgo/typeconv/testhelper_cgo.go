//go:build cgo

package main

/*
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <math.h>

enum TestColor {
    TEST_RED    = 0,
    TEST_GREEN  = 1,
    TEST_BLUE   = 2,
    TEST_YELLOW = 3,
};

typedef struct {
    int x;
    int y;
} TestPoint;

typedef struct {
    TestPoint topLeft;
    TestPoint bottomRight;
    int   color;
} TestRect;

static int test_compare_ints(const void* a, const void* b) {
    int ia = *(const int*)a;
    int ib = *(const int*)b;
    if (ia < ib) return -1;
    if (ia > ib) return 1;
    return 0;
}

static void test_sort_ints(int* arr, int len) {
    qsort(arr, len, sizeof(int), test_compare_ints);
}

static double test_degrees_to_radians(double deg) {
    return deg * M_PI / 180.0;
}
*/
import "C"
import "unsafe"

// cgoSortInts 使用 C 的 qsort 排序 int 切片
func cgoSortInts(slice []int) []int {
	if len(slice) == 0 {
		return slice
	}
	cArr := make([]C.int, len(slice))
	for i, v := range slice {
		cArr[i] = C.int(v)
	}
	C.test_sort_ints(
		(*C.int)(unsafe.Pointer(&cArr[0])),
		C.int(len(cArr)),
	)
	result := make([]int, len(slice))
	for i, v := range cArr {
		result[i] = int(v)
	}
	return result
}

// cgoDegreesToRadians 使用 C 函数将角度转弧度
func cgoDegreesToRadians(deg float64) float64 {
	return float64(C.test_degrees_to_radians(C.double(deg)))
}

// cgoPointX 返回 TestPoint 的 x 值
func cgoPointX(x, y int) int {
	p := C.TestPoint{x: C.int(x), y: C.int(y)}
	return int(p.x)
}

// cgoPointY 返回 TestPoint 的 y 值
func cgoPointY(x, y int) int {
	p := C.TestPoint{x: C.int(x), y: C.int(y)}
	return int(p.y)
}

// cgoRectInfo 返回矩形信息
func cgoRectInfo(x1, y1, x2, y2, color int) (int, int, int, int, int) {
	rect := C.TestRect{
		topLeft:     C.TestPoint{x: C.int(x1), y: C.int(y1)},
		bottomRight: C.TestPoint{x: C.int(x2), y: C.int(y2)},
		color:       C.int(color),
	}
	return int(rect.topLeft.x), int(rect.topLeft.y),
		int(rect.bottomRight.x), int(rect.bottomRight.y),
		int(rect.color)
}

// cgoEnumName 返回枚举的名称
func cgoEnumName(color int) string {
	names := map[C.enum_TestColor]string{
		C.TEST_RED: "RED", C.TEST_GREEN: "GREEN",
		C.TEST_BLUE: "BLUE", C.TEST_YELLOW: "YELLOW",
	}
	return names[C.enum_TestColor(color)]
}

// cgoVerifyMathLib 验证数学库链接正确
func cgoVerifyMathLib() bool {
	rad := C.test_degrees_to_radians(C.double(180.0))
	return float64(rad) > 3.14 && float64(rad) < 3.15
}