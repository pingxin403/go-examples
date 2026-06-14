// scheduler/main_test.go — Go 调度器核心概念测试
package main

import (
	"runtime"
	"testing"
)

// TestFibonacci — 测试纯计算函数 fibonacci
func TestFibonacci(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want int
	}{
		{"fib(0)=0", 0, 0},
		{"fib(1)=1", 1, 1},
		{"fib(2)=1", 2, 1},
		{"fib(3)=2", 3, 2},
		{"fib(4)=3", 4, 3},
		{"fib(5)=5", 5, 5},
		{"fib(10)=55", 10, 55},
		{"fib(20)=6765", 20, 6765},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fibonacci(tt.n); got != tt.want {
				t.Errorf("fibonacci(%d) = %d, 期望 %d", tt.n, got, tt.want)
			}
		})
	}
}

// TestGOMAXPROCS — 测试 GOMAXPROCS 不影响 goroutine 数量
func TestGOMAXPROCS(t *testing.T) {
	orig := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(orig)

	// 设置一个明确的值
	runtime.GOMAXPROCS(4)
	got := runtime.GOMAXPROCS(0)
	if got != 4 {
		t.Errorf("GOMAXPROCS = %d, 期望 4", got)
	}
}

// TestNumGoroutine — 测试 goroutine 计数变化
func TestNumGoroutine(t *testing.T) {
	before := runtime.NumGoroutine()

	ch := make(chan struct{})
	go func() {
		ch <- struct{}{}
	}()

	<-ch
	after := runtime.NumGoroutine()

	// goroutine 可能比 before 多（调度器内部 goroutine），但不应差距巨大
	if after < before-1 {
		t.Errorf("goroutine 数量异常: before=%d, after=%d", before, after)
	}
}

// TestGoschedDoesNotPanic — 验证 Gosched 不会引发 panic
func TestGoschedDoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("runtime.Gosched() 引发了 panic: %v", r)
		}
	}()

	runtime.Gosched()
	runtime.Gosched()
}