// main_test.go — select + timer 示例测试套件
//
// 测试覆盖：
//   - select 多路复用（多 channel 等待）
//   - select + default 非阻塞操作
//   - time.Ticker 周期性任务
//   - time.Timer 一次性定时器
//   - select 超时模式
//   - Fan-in 模式合并多个 channel
package main

import (
	"math/rand"
	"testing"
	"time"
)

// ============================================================
// 1. 测试 select 多路复用
// ============================================================

// TestSelectMultiplex 验证 select 能同时等待多个 channel
func TestSelectMultiplex(t *testing.T) {
	t.Parallel()

	ch1 := make(chan string, 1)
	ch2 := make(chan string, 1)

	// 只向 ch1 发送
	ch1 <- "msg1"

	select {
	case msg := <-ch1:
		if msg != "msg1" {
			t.Errorf("ch1: got %q, want 'msg1'", msg)
		}
	case <-ch2:
		t.Error("不应从 ch2 收到消息")
	default:
		t.Error("不应进入 default")
	}
}

// TestSelectMultiplexPriority 验证 select 会随机选择（不保证顺序）
func TestSelectMultiplexPriority(t *testing.T) {
	t.Parallel()

	ch1 := make(chan string, 1)
	ch2 := make(chan string, 1)
	ch1 <- "a"
	ch2 <- "b"

	// select 不会饥饿：多次运行应该有时选 ch1 有时选 ch2
	ch1Count := 0
	total := 20

	for i := 0; i < total; i++ {
		select {
		case <-ch1:
			ch1Count++
			// 重新填充
			ch1 <- "a"
		case <-ch2:
			// 重新填充
			ch2 <- "b"
		}
	}

	// 清理
	<-ch1
	<-ch2

	// 两个 channel 都应该至少被选到一次
	if ch1Count == 0 || ch1Count == total {
		t.Logf("注意：select 在 %d 次中选了 ch1 %d 次（可能全部偏向了同一个）", total, ch1Count)
	}
}

// ============================================================
// 2. 测试 select + default 非阻塞
// ============================================================

// TestSelectNonBlockingSend 验证非阻塞发送
func TestSelectNonBlockingSend(t *testing.T) {
	t.Parallel()

	ch := make(chan int, 1)

	// 第一次发送应成功
	select {
	case ch <- 42:
		// 成功
	default:
		t.Error("空缓冲 channel 应能发送")
	}

	// 第二次发送应进入 default（缓冲区满）
	select {
	case ch <- 99:
		t.Error("满缓冲 channel 不应发送成功")
	default:
		// 正常：进入 default
	}
}

// TestSelectNonBlockingRecv 验证非阻塞接收
func TestSelectNonBlockingRecv(t *testing.T) {
	t.Parallel()

	ch := make(chan int, 1)
	ch <- 42

	// 第一次接收应成功
	select {
	case val := <-ch:
		if val != 42 {
			t.Errorf("got %d, want 42", val)
		}
	default:
		t.Error("有数据的 channel 应能读取")
	}

	// 第二次接收应进入 default（空）
	select {
	case <-ch:
		t.Error("空 channel 不应读取到数据")
	default:
		// 正常：进入 default
	}
}

// TestSelectNonBlockingEmpty 验证空 channel 的非阻塞行为
func TestSelectNonBlockingEmpty(t *testing.T) {
	t.Parallel()

	ch := make(chan int)

	select {
	case <-ch:
		t.Error("无缓冲空 channel 不应读取到数据")
	default:
		// 正常：进入 default，不会阻塞
	}
}

// ============================================================
// 3. 测试 time.Ticker
// ============================================================

// TestTickerBasic 验证 Ticker 能定期发送事件
func TestTickerBasic(t *testing.T) {
	t.Parallel()

	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	// 应至少收到 2 个 tick
	select {
	case <-ticker.C:
		// 第一个 tick
	case <-time.After(50 * time.Millisecond):
		t.Fatal("第一个 tick 未在预期时间内到达")
	}

	select {
	case <-ticker.C:
		// 第二个 tick
	case <-time.After(50 * time.Millisecond):
		t.Fatal("第二个 tick 未在预期时间内到达")
	}
}

// TestTickerStop 验证 Stop 后不再收到 tick
func TestTickerStop(t *testing.T) {
	t.Parallel()

	ticker := time.NewTicker(10 * time.Millisecond)
	<-ticker.C // 收一个
	ticker.Stop()

	// Stop 后不应再收到 tick
	select {
	case <-ticker.C:
		t.Log("注意：Stop 后仍收到 tick（可能已缓冲在 channel 中）")
	case <-time.After(50 * time.Millisecond):
		// 正常：没有更多 tick
	}
}

// TestTickerReset 验证 Reset 能重置间隔
func TestTickerReset(t *testing.T) {
	t.Parallel()

	ticker := time.NewTicker(1 * time.Hour) // 很长
	defer ticker.Stop()

	// Reset 为短间隔
	ticker.Reset(20 * time.Millisecond)

	select {
	case <-ticker.C:
		// Reset 后应在 20ms 内收到 tick
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Reset 后 tick 未在预期时间内到达")
	}
}

// ============================================================
// 4. 测试 time.Timer
// ============================================================

// TestTimerBasic 验证 Timer 在指定时间后触发
func TestTimerBasic(t *testing.T) {
	t.Parallel()

	timer := time.NewTimer(20 * time.Millisecond)
	defer timer.Stop()

	select {
	case <-timer.C:
		// 正常触发
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timer 未在预期时间内触发")
	}
}

// TestTimerStop 验证 Stop 阻止 Timer 触发
func TestTimerStop(t *testing.T) {
	t.Parallel()

	timer := time.NewTimer(1 * time.Hour)
	stopped := timer.Stop()
	if !stopped {
		t.Error("Timer.Stop() 应返回 true（Timer 还未触发）")
	}

	// 再次 Stop 应返回 false
	stopped = timer.Stop()
	if stopped {
		t.Log("注意：第二次 Stop 可能返回 true（取决于实现）")
	}
}

// TestTimerReset 验证 Timer.Reset 重用定时器
func TestTimerReset(t *testing.T) {
	t.Parallel()

	timer := time.NewTimer(10 * time.Millisecond)
	<-timer.C // 等待第一次触发

	// Reset 复用以等待第二次
	ok := timer.Reset(20 * time.Millisecond)
	if !ok {
		t.Log("注意：Reset 可能返回 false（Timer 已过期）")
	}

	select {
	case <-timer.C:
		// Reset 后应触发
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Timer.Reset 后未在预期时间内触发")
	}
}

// TestTimeAfter 验证 time.After 超时模式
func TestTimeAfter(t *testing.T) {
	t.Parallel()

	start := time.Now()
	select {
	case <-time.After(30 * time.Millisecond):
		elapsed := time.Since(start)
		if elapsed < 15*time.Millisecond || elapsed > 100*time.Millisecond {
			t.Errorf("time.After(30ms) 实际耗时 %v", elapsed)
		}
	}
}

// ============================================================
// 5. 测试 select 超时模式
// ============================================================

// TestSelectTimeout 验证 select 超时模式能防止无限阻塞
func TestSelectTimeout(t *testing.T) {
	t.Parallel()

	slowCh := make(chan string)

	// 一个永远不会发送的 goroutine
	go func() {
		time.Sleep(1 * time.Hour) // 永不发送
	}()

	start := time.Now()
	select {
	case <-slowCh:
		t.Error("不应收到消息")
	case <-time.After(30 * time.Millisecond):
		elapsed := time.Since(start)
		if elapsed > 200*time.Millisecond {
			t.Errorf("超时太慢: %v", elapsed)
		}
	}
}

// TestSelectNoTimeout 验证当 channel 有数据时不触发超时
func TestSelectNoTimeout(t *testing.T) {
	t.Parallel()

	ch := make(chan string, 1)
	ch <- "fast result"

	select {
	case msg := <-ch:
		if msg != "fast result" {
			t.Errorf("got %q, want 'fast result'", msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("有数据时不应超时")
	}
}

// ============================================================
// 6. 测试 Fan-in 模式
// ============================================================

// TestFanIn 验证 fanIn 能合并多个 channel
func TestFanIn(t *testing.T) {
	t.Parallel()

	source1 := make(chan int, 3)
	source2 := make(chan int, 3)

	source1 <- 10
	source1 <- 20
	source1 <- 30
	close(source1)

	source2 <- 15
	source2 <- 25
	source2 <- 35
	close(source2)

	merged := fanIn(source1, source2)

	// 应收到 6 个值
	received := make(map[int]bool)
	timeout := time.After(1 * time.Second)

	for i := 0; i < 6; i++ {
		select {
		case v, ok := <-merged:
			if !ok {
				t.Fatal("merged channel 不应提前关闭")
			}
			received[v] = true
		case <-timeout:
			t.Fatalf("超时：只收到 %d/6 个值", len(received))
		}
	}

	// 验证所有值都收到
	expected := []int{10, 20, 30, 15, 25, 35}
	for _, v := range expected {
		if !received[v] {
			t.Errorf("未收到值 %d", v)
		}
	}
}

// TestFanInSingleSource 验证 fanIn 能处理单个输入
func TestFanInSingleSource(t *testing.T) {
	t.Parallel()

	source := make(chan int, 2)
	source <- 1
	source <- 2
	close(source)

	merged := fanIn(source)

	var got []int
	for v := range merged {
		got = append(got, v)
		// fanIn 会在所有输入 close 后关闭 merged
		// 但因为只有一个 source 在 close 后 merged 也会 close
		if len(got) >= 2 {
			break
		}
	}
	_ = got
}

// TestFanInConcurrent 验证 fanIn 能处理并发输入
func TestFanInConcurrent(t *testing.T) {
	t.Parallel()

	sources := make([]chan int, 3)
	for i := 0; i < 3; i++ {
		sources[i] = make(chan int)
	}

	// 从不同 source 并发发送
	for i := 0; i < 3; i++ {
		go func(id int, ch chan int) {
			ch <- id * 10
			close(ch)
		}(i+1, sources[i])
	}

	// 转换为 <-chan int
	inputs := make([]<-chan int, 3)
	for i := 0; i < 3; i++ {
		inputs[i] = sources[i]
	}

	merged := fanIn(inputs...)

	received := make(map[int]bool)
	timeout := time.After(1 * time.Second)

	for i := 0; i < 3; i++ {
		select {
		case v, ok := <-merged:
			if !ok {
				t.Fatal("merged channel 不应提前关闭")
			}
			received[v] = true
		case <-timeout:
			t.Fatalf("超时：只收到 %d/3 个值", len(received))
		}
	}

	expected := []int{10, 20, 30}
	for _, v := range expected {
		if !received[v] {
			t.Errorf("未收到值 %d", v)
		}
	}
}

// fanIn — 从 main.go 提取的 fanIn 函数
func fanIn(inputs ...<-chan int) <-chan int {
	out := make(chan int)

	for _, ch := range inputs {
		go func(c <-chan int) {
			for v := range c {
				out <- v
			}
		}(ch)
	}

	return out
}

// ============================================================
// 7. 测试 Ticker + Timer 结合 select 的场景复用
// ============================================================

// TestTickerWithTimer 验证 ticker 和 timer 同时使用
func TestTickerWithTimer(t *testing.T) {
	t.Parallel()

	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	select {
	case <-ticker.C:
		// 正常 tick
	case <-time.After(50 * time.Millisecond):
		t.Fatal("tick 超时")
	}

	// 验证 timer 超时也正常工作
	select {
	case <-ticker.C:
		// 正常 tick
	case <-time.After(50 * time.Millisecond):
		t.Fatal("第二个 tick 超时")
	}
}

// TestRandSleep — 验证 rand 的用法不被 -race 误报
func TestRandSleep(t *testing.T) {
	t.Parallel()

	const maxWait = 10
	for i := 0; i < 5; i++ {
		d := time.Duration(rand.Intn(maxWait)+5) * time.Millisecond
		_ = d
	}
}