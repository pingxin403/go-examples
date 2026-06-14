// main_test.go — atomic 包原子操作示例测试套件
//
// 测试覆盖：
//   - atomic.AddInt64 原子计数器
//   - atomic.Load / atomic.Store 安全读写
//   - atomic.CompareAndSwap CAS 操作
//   - atomic.Value 类型安全原子值
package main

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ============================================================
// 1. 测试 atomic.AddInt64 原子计数器
// ============================================================

// TestAtomicAddInt64 验证原子计数器在多 goroutine 下结果正确
func TestAtomicAddInt64(t *testing.T) {
	t.Parallel()

	var counter int64
	var wg sync.WaitGroup
	n := 1000

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt64(&counter, 1)
		}()
	}
	wg.Wait()

	if counter != int64(n) {
		t.Errorf("计数器值错误: got %d, want %d", counter, n)
	}
}

// TestAtomicAddInt64Negative 验证原子减法
func TestAtomicAddInt64Negative(t *testing.T) {
	t.Parallel()

	var counter int64 = 100
	var wg sync.WaitGroup
	n := 50

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt64(&counter, -1)
		}()
	}
	wg.Wait()

	if counter != 50 {
		t.Errorf("计数器值错误: got %d, want 50", counter)
	}
}

// ============================================================
// 2. 测试 atomic.Load / atomic.Store
// ============================================================

// TestAtomicLoadStore 验证原子读写一致性
func TestAtomicLoadStore(t *testing.T) {
	t.Parallel()

	var status int64

	// Store 写入
	atomic.StoreInt64(&status, 42)

	// Load 读取
	got := atomic.LoadInt64(&status)
	if got != 42 {
		t.Errorf("LoadInt64: got %d, want 42", got)
	}
}

// TestAtomicStoreThenLoadSequential 验证顺序一致性
func TestAtomicStoreThenLoadSequential(t *testing.T) {
	t.Parallel()

	var val int64

	values := []int64{0, 1, 100, -5, 999}
	for _, v := range values {
		atomic.StoreInt64(&val, v)
		got := atomic.LoadInt64(&val)
		if got != v {
			t.Errorf("Store(%d) 后 Load = %d", v, got)
		}
	}
}

// ============================================================
// 3. 测试 atomic.CompareAndSwap — CAS 操作
// ============================================================

// TestCompareAndSwapSuccess 验证 CAS 成功场景
func TestCompareAndSwapSuccess(t *testing.T) {
	t.Parallel()

	var val int64 = 10

	// CAS: 期望值是 10，交换为 20
	swapped := atomic.CompareAndSwapInt64(&val, 10, 20)
	if !swapped {
		t.Error("CAS 应成功（当前值=10，期望=10）")
	}
	if val != 20 {
		t.Errorf("CAS 后 val=%d, want 20", val)
	}
}

// TestCompareAndSwapFailure 验证 CAS 失败场景
func TestCompareAndSwapFailure(t *testing.T) {
	t.Parallel()

	var val int64 = 10

	// CAS: 期望值是 99，但当前值是 10，应失败
	swapped := atomic.CompareAndSwapInt64(&val, 99, 20)
	if swapped {
		t.Error("CAS 应失败（当前值=10，期望=99）")
	}
	if val != 10 {
		t.Errorf("CAS 失败后 val 应保持不变: got %d, want 10", val)
	}
}

// TestSpinLockViaCAS 验证用 CAS 实现自旋锁的正确性
func TestSpinLockViaCAS(t *testing.T) {
	t.Parallel()

	var spinLock int64 // 0=未锁定
	var shared int64
	var wg sync.WaitGroup
	n := 3

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 自旋等待锁
			for !atomic.CompareAndSwapInt64(&spinLock, 0, 1) {
				time.Sleep(time.Microsecond)
			}
			shared++
			atomic.StoreInt64(&spinLock, 0)
		}()
	}
	wg.Wait()

	if shared != int64(n) {
		t.Errorf("自旋锁保护后 shared=%d, want %d", shared, n)
	}
	if atomic.LoadInt64(&spinLock) != 0 {
		t.Error("自旋锁应在所有 goroutine 完成后释放")
	}
}

// TestMultipleCASAttempts 验证连续多次 CAS 操作
func TestMultipleCASAttempts(t *testing.T) {
	t.Parallel()

	var val int64 = 0

	// 模拟自增（非原子的 CAS 循环）
	for i := int64(1); i <= 10; i++ {
		for {
			current := atomic.LoadInt64(&val)
			if atomic.CompareAndSwapInt64(&val, current, current+1) {
				break
			}
		}
	}

	if val != 10 {
		t.Errorf("CAS 循环后 val=%d, want 10", val)
	}
}

// ============================================================
// 4. 测试 atomic.Value
// ============================================================

// TestAtomicValueStoreLoad 验证 atomic.Value 的 Store 和 Load
func TestAtomicValueStoreLoad(t *testing.T) {
	t.Parallel()

	// atomic.Value 只能存储单一类型，每个实例只能存同一种类型
	var vStr atomic.Value
	vStr.Store("hello")
	got := vStr.Load().(string)
	if got != "hello" {
		t.Errorf("atomic.Value Load: got %q, want 'hello'", got)
	}

	// 使用独立的 Value 存储整数
	var vInt atomic.Value
	vInt.Store(42)
	gotInt := vInt.Load().(int)
	if gotInt != 42 {
		t.Errorf("atomic.Value Load: got %d, want 42", gotInt)
	}
}

// TestAtomicValueMap 验证 atomic.Value 存储 map
func TestAtomicValueMap(t *testing.T) {
	t.Parallel()

	var config atomic.Value

	initial := map[string]string{
		"host": "localhost",
		"port": "3306",
	}
	config.Store(initial)

	// 读取
	cfg := config.Load().(map[string]string)
	if cfg["host"] != "localhost" {
		t.Errorf("host: got %q, want 'localhost'", cfg["host"])
	}
	if cfg["port"] != "3306" {
		t.Errorf("port: got %q, want '3306'", cfg["port"])
	}

	// 更新配置
	updated := map[string]string{
		"host": "prod-server",
		"port": "5432",
	}
	config.Store(updated)

	cfg2 := config.Load().(map[string]string)
	if cfg2["host"] != "prod-server" {
		t.Errorf("更新后 host: got %q, want 'prod-server'", cfg2["host"])
	}
	if cfg2["port"] != "5432" {
		t.Errorf("更新后 port: got %q, want '5432'", cfg2["port"])
	}
}

// TestAtomicValueConcurrent 验证 atomic.Value 并发读写安全
func TestAtomicValueConcurrent(t *testing.T) {
	t.Parallel()

	var config atomic.Value
	config.Store(map[string]string{"version": "v1"})

	var wg sync.WaitGroup

	// 多个读 goroutine
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				cfg := config.Load().(map[string]string)
				_ = cfg["version"]
			}
		}()
	}

	// 写 goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		config.Store(map[string]string{"version": "v2"})
		config.Store(map[string]string{"version": "v3"})
	}()

	wg.Wait()

	// 最终值应为 v2 或 v3（取决于执行顺序）
	cfg := config.Load().(map[string]string)
	v := cfg["version"]
	if v != "v2" && v != "v3" {
		t.Errorf("最终 version: got %q, want 'v2' or 'v3'", v)
	}
}

// TestAtomicValueNil 验证 atomic.Value 不能 Store nil
func TestAtomicValueNil(t *testing.T) {
	t.Parallel()

	var v atomic.Value

	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		v.Store(nil)
	}()

	if !panicked {
		t.Error("Store(nil) 应 panic")
	}
}

// TestAtomicValueTypeMismatch 验证 atomic.Value 不能存储不同类型
func TestAtomicValueTypeMismatch(t *testing.T) {
	t.Parallel()

	var v atomic.Value
	v.Store("hello")

	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		v.Store(42) // 不同类型
	}()

	if !panicked {
		t.Error("Store 不同类型应 panic")
	}
}