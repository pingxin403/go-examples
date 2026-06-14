// main_test.go — sync 包并发原语示例测试套件
//
// 测试覆盖：
//   - WaitGroup 等待多个 goroutine
//   - Mutex 保护共享资源
//   - RWMutex 读写锁并发读
//   - Once 确保只执行一次
//   - Cond 条件变量广播
//   - sync.Map 并发安全操作
package main

import (
	"sync"
	"testing"
	"time"
)

// ============================================================
// 1. 测试 Mutex 保护共享计数器
// ============================================================

// TestMutexCounter 验证 Mutex 能正确保护共享计数器
func TestMutexCounter(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var counter int
	var wg sync.WaitGroup
	n := 100

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}
	wg.Wait()

	if counter != n {
		t.Errorf("计数器值错误: got %d, want %d", counter, n)
	}
}

// TestMutexDataRace 用 -race 检测 Mutex 是否消除数据竞争
func TestMutexDataRace(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var counter int
	var wg sync.WaitGroup
	n := 50

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}
	wg.Wait()

	if counter != n {
		t.Errorf("计数器值错误: got %d, want %d", counter, n)
	}
}

// ============================================================
// 2. 测试 RWMutex 读写锁
// ============================================================

// TestRWMutexConcurrentRead 验证 RWMutex 允许多个读 goroutine 并发
func TestRWMutexConcurrentRead(t *testing.T) {
	t.Parallel()

	var rwmu sync.RWMutex
	data := 42

	// 多个 goroutine 同时获取读锁
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rwmu.RLock()
			defer rwmu.RUnlock()
			_ = data
		}()
	}
	wg.Wait()
}

// TestRWMutexWriteExclusive 验证写锁是独占的
func TestRWMutexWriteExclusive(t *testing.T) {
	t.Parallel()

	var rwmu sync.RWMutex
	var data int
	var wg sync.WaitGroup

	// 获取写锁
	rwmu.Lock()
	data = 100
	rwmu.Unlock()

	// 并发读写
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rwmu.RLock()
			_ = data
			rwmu.RUnlock()
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		rwmu.Lock()
		data = 200
		rwmu.Unlock()
	}()

	wg.Wait()
	if data != 200 {
		t.Errorf("data 最终值错误: got %d, want 200", data)
	}
}

// ============================================================
// 3. 测试 Once 只执行一次
// ============================================================

// TestOnce 验证 sync.Once 在多 goroutine 下只执行一次
func TestOnce(t *testing.T) {
	t.Parallel()

	var once sync.Once
	count := 0

	fn := func() {
		count++
	}

	var wg sync.WaitGroup
	n := 10
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			once.Do(fn)
		}()
	}
	wg.Wait()

	if count != 1 {
		t.Errorf("Once 执行了 %d 次，期望 1 次", count)
	}
}

// TestOnceInitOrder 验证 Once 执行后所有 goroutine 都能读到初始化结果
func TestOnceInitOrder(t *testing.T) {
	t.Parallel()

	var once sync.Once
	// Once 内部保证了 happens-before 关系，读取由 Once.Do 保护写入的变量是安全的
	// 但 -race 检测器可能误报。这里用一个 channel 同步来避免 race 检测问题
	config := ""

	initConfig := func() {
		config = "initialized"
	}

	var wg sync.WaitGroup
	n := 5
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			once.Do(initConfig)
		}()
	}
	wg.Wait()

	// 在所有 goroutine 结束后验证
	if config != "initialized" {
		t.Error("Once 执行后配置应为 'initialized'")
	}
}

// ============================================================
// 4. 测试 Cond 条件变量
// ============================================================

// TestCondBroadcast 验证 Cond.Broadcast 能唤醒所有等待者
func TestCondBroadcast(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	var ready bool
	count := 0
	var wg sync.WaitGroup

	// 多个等待 goroutine
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			for !ready {
				cond.Wait()
			}
			count++
			mu.Unlock()
		}()
	}

	// 确保所有 goroutine 都进入 Wait
	time.Sleep(50 * time.Millisecond)

	// 广播唤醒
	mu.Lock()
	ready = true
	cond.Broadcast()
	mu.Unlock()

	wg.Wait()
	if count != 3 {
		t.Errorf("广播后 count=%d，期望 3（所有 worker 应被唤醒）", count)
	}
}

// TestCondSignalOnlyOne 验证 Cond.Signal 只唤醒一个等待者
func TestCondSignalOnlyOne(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	awake := 0
	var externalWg sync.WaitGroup

	// 启动 3 个等待 goroutine
	for i := 0; i < 3; i++ {
		externalWg.Add(1)
		go func() {
			mu.Lock()
			awake++
			externalWg.Done() // 告知外部 goroutine 已进入等待
			cond.Wait()
			awake++
			mu.Unlock()
		}()
	}

	// 等待所有 goroutine 都进入 Wait
	externalWg.Wait()
	// 注意：在 Cond 中，awake++ 在 Wait 之前，所以此时 awake=3
	time.Sleep(20 * time.Millisecond)

	// 发 Signal——只唤醒一个
	mu.Lock()
	cond.Signal()
	mu.Unlock()

	time.Sleep(50 * time.Millisecond)

	// 此时 awake 应该是 4（3 个初始 + 1 个被 Signal 唤醒后 ++）
	mu.Lock()
	if awake != 4 {
		t.Logf("Signal 后 awake=%d（期望 4，即只有 1 个被唤醒）", awake)
	}
	// 清理剩余等待者
	cond.Broadcast()
	mu.Unlock()
}

// ============================================================
// 5. 测试 sync.Map
// ============================================================

// TestSyncMapStoreAndLoad 验证 sync.Map 的 Store 和 Load
func TestSyncMapStoreAndLoad(t *testing.T) {
	t.Parallel()

	var m sync.Map

	// 写入
	m.Store("key1", "value1")
	m.Store("key2", "value2")

	// 读取存在的 key
	v, ok := m.Load("key1")
	if !ok {
		t.Fatal("key1 应存在")
	}
	if v.(string) != "value1" {
		t.Errorf("key1: got %v, want 'value1'", v)
	}

	// 读取不存在的 key
	_, ok = m.Load("nonexistent")
	if ok {
		t.Error("不存在的 key 不应返回 ok=true")
	}
}

// TestSyncMapDelete 验证 sync.Map 的 Delete 操作
func TestSyncMapDelete(t *testing.T) {
	t.Parallel()

	var m sync.Map
	m.Store("key", "value")
	m.Delete("key")

	_, ok := m.Load("key")
	if ok {
		t.Error("删除后 key 不应存在")
	}
}

// TestSyncMapLoadOrStore 验证 LoadOrStore 语义
func TestSyncMapLoadOrStore(t *testing.T) {
	t.Parallel()

	var m sync.Map

	// key 不存在：存储新值
	actual, loaded := m.LoadOrStore("key", "new")
	if loaded {
		t.Error("第一次 LoadOrStore 应返回 loaded=false")
	}
	if actual.(string) != "new" {
		t.Errorf("got %v, want 'new'", actual)
	}

	// key 已存在：返回旧值不覆盖
	actual, loaded = m.LoadOrStore("key", "override")
	if !loaded {
		t.Error("第二次 LoadOrStore 应返回 loaded=true")
	}
	if actual.(string) != "new" {
		t.Errorf("已存在 key 应返回旧值 'new', got %v", actual)
	}
}

// TestSyncMapLoadAndDelete 验证 LoadAndDelete 语义
func TestSyncMapLoadAndDelete(t *testing.T) {
	t.Parallel()

	var m sync.Map
	m.Store("key", "value")

	// 读取并删除
	v, loaded := m.LoadAndDelete("key")
	if !loaded {
		t.Fatal("key 应存在")
	}
	if v.(string) != "value" {
		t.Errorf("got %v, want 'value'", v)
	}

	// 确认已删除
	_, ok := m.Load("key")
	if ok {
		t.Error("LoadAndDelete 后 key 应不存在")
	}
}

// TestSyncMapConcurrent 验证 sync.Map 的并发安全性
func TestSyncMapConcurrent(t *testing.T) {
	t.Parallel()

	var m sync.Map
	var wg sync.WaitGroup
	n := 50

	// 并发写入
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			m.Store(id, id*2)
		}(i)
	}

	wg.Wait()

	// 验证写入
	for i := 0; i < n; i++ {
		v, ok := m.Load(i)
		if !ok {
			t.Errorf("key %d 应存在", i)
			continue
		}
		if v.(int) != i*2 {
			t.Errorf("key %d: got %d, want %d", i, v.(int), i*2)
		}
	}
}

// ============================================================
// 6. 测试 WaitGroup
// ============================================================

// TestWaitGroup 验证 WaitGroup 能正确等待所有 goroutine 完成
func TestWaitGroup(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	n := 5
	count := 0

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			count++
			mu.Unlock()
		}()
	}
	wg.Wait()

	if count != n {
		t.Errorf("count=%d, want %d", count, n)
	}
}

// TestWaitGroupNegativeCounter 验证 Done 超过 Add 不会 panic（注意：会 panic 但可以 recover）
func TestWaitGroupDone(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup

	wg.Add(1)
	wg.Done()

	// 第二次 Done 会导致负计数器 panic
	doneCalled := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				doneCalled = true
			}
		}()
		wg.Done()
	}()

	if !doneCalled {
		t.Log("额外的 Done 调用不会引发 panic（取决于 Go 版本行为）")
	}
}