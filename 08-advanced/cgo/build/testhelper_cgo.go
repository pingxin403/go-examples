//go:build cgo

package main

/*
#include <stdio.h>
#include <stdlib.h>
#include <math.h>
#include <unistd.h>

static int test_get_cpu_cores() {
    return (int)sysconf(_SC_NPROCESSORS_ONLN);
}

static double test_compute_sqrt(double x) {
    return sqrt(x);
}
*/
import "C"

// cgoGetCPUCores 通过 C 函数获取 CPU 核心数
func cgoGetCPUCores() int {
	return int(C.test_get_cpu_cores())
}

// cgoComputeSqrt 通过 C 数学库计算平方根
func cgoComputeSqrt(x float64) float64 {
	return float64(C.test_compute_sqrt(C.double(x)))
}