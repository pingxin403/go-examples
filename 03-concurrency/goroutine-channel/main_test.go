// main_test.go — goroutine+channel 示例完整测试套件
//
// 测试覆盖：
//   - worker 函数（工作池核心函数）
//   - 生成器模式（gen 闭包）
//   - channel 方向约束（sendOnly / recvOnly）
//   - 有缓冲 / 无缓冲 channel 通信
//   - 生产者-消费者模式
package main

import (
	"fmt"
	"testing"
	"time"
)

// ============================================================
// 1. 测试 worker 函数 — 工作池核心
// ============================================================

// TestWorker 验证 worker 能正确处理任务并返回结果
func TestWorker(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		jobs   []int
		wantFn func(results []int) bool
	}{
		{
			name: "多个任务",
			jobs: []int{1, 2, 3, 4, 5},
			wantFn: func(results []int) bool {
				if len(results) != 5 {
					return false
				}
				for i, v := range results {
					if v != (i+1)*2 {
						return false
					}
				}
				return true
			},
		},
		{
			name: "单个任务",
			jobs: []int{10},
			wantFn: func(results []int) bool {
				return len(results) == 1 && results[0] == 20
			},
		},
		{
			name: "空任务列表",
			jobs: []int{},
			wantFn: func(results []int) bool {
				return len(results) == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobs := make(chan int, len(tt.jobs))
			results := make(chan int, len(tt.jobs))

			// 发送任务
			for _, j := range tt.jobs {
				jobs <- j
			}
			close(jobs)

			// 启动一个 worker 处理
			go worker(1, jobs, results)

			// 收集结果
			var got []int
			for i := 0; i < len(tt.jobs); i++ {
				select {
				case r := <-results:
					got = append(got, r)
				case <-time.After(500 * time.Millisecond):
					t.Fatal("worker 处理超时")
				}
			}

			if !tt.wantFn(got) {
				t.Errorf("worker 结果不符合预期: got %v", got)
			}
		})
	}
}

// TestWorkerMultiple 测试多个 worker 并发处理
func TestWorkerMultiple(t *testing.T) {
	t.Parallel()

	const numJobs = 10
	const numWorkers = 3

	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)

	// 发送任务
	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}
	close(jobs)

	// 启动多个 worker
	for w := 1; w <= numWorkers; w++ {
		go worker(w, jobs, results)
	}

	// 收集结果
	resultMap := make(map[int]int)
	for i := 0; i < numJobs; i++ {
		select {
		case r := <-results:
			resultMap[r/2] = r // r = job * 2
		case <-time.After(1 * time.Second):
			t.Fatalf("等待结果超时，已收到 %d/%d", len(resultMap), numJobs)
		}
	}

	// 验证每个 job 都被处理且结果正确
	if len(resultMap) != numJobs {
		t.Errorf("结果数不匹配: got %d, want %d", len(resultMap), numJobs)
	}
	for jobID := 1; jobID <= numJobs; jobID++ {
		if resultMap[jobID] != jobID*2 {
			t.Errorf("job %d 结果错误: got %d, want %d", jobID, resultMap[jobID], jobID*2)
		}
	}
}

// ============================================================
// 2. 测试生成器模式（gen 闭包）
// ============================================================

// TestGenChannel 验证生成器能生成正确数量的值并关闭 channel
func TestGenChannel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		count int
		want  []int
	}{
		{name: "生成 3 个值", count: 3, want: []int{10, 20, 30}},
		{name: "生成 0 个值", count: 0, want: []int{}},
		{name: "生成 1 个值", count: 1, want: []int{10}},
		{name: "生成 5 个值", count: 5, want: []int{10, 20, 30, 40, 50}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := genChannel(tt.count)

			var got []int
			for v := range ch {
				got = append(got, v)
			}

			if len(got) != len(tt.want) {
				t.Fatalf("长度不匹配: got %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("索引 %d: got %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// genChannel — 从 main.go 提取的生成器闭包（独立函数方便测试）
func genChannel(count int) chan int {
	ch := make(chan int)
	go func() {
		for i := 1; i <= count; i++ {
			ch <- i * 10
		}
		close(ch)
	}()
	return ch
}

// ============================================================
// 3. 测试 channel 方向约束
// ============================================================

// TestSendOnlyChannel 验证只发送 channel 能正确发送数据
func TestSendOnlyChannel(t *testing.T) {
	t.Parallel()

	ch := make(chan string, 3)

	sendTask := func(out chan<- string, id int) {
		out <- fmt.Sprintf("任务 %d 完成", id)
	}

	sendTask(ch, 1)
	sendTask(ch, 2)
	sendTask(ch, 3)

	close(ch)

	var got []string
	for v := range ch {
		got = append(got, v)
	}

	if len(got) != 3 {
		t.Fatalf("期望 3 条消息，got %d", len(got))
	}
	for i, msg := range got {
		want := fmt.Sprintf("任务 %d 完成", i+1)
		if msg != want {
			t.Errorf("消息 %d: got %q, want %q", i, msg, want)
		}
	}
}

// TestReceiveOnlyChannel 验证只接收 channel 能正确读取数据
func TestReceiveOnlyChannel(t *testing.T) {
	t.Parallel()

	ch := make(chan string, 3)
	ch <- "msg1"
	ch <- "msg2"
	ch <- "msg3"
	close(ch)

	recvTask := func(in <-chan string) int {
		msg := <-in
		return len(msg)
	}

	// 重新创建 channel 以避免方向问题
	ch2 := make(chan string, 3)
	ch2 <- "hello"
	ch2 <- "world!"
	ch2 <- "hi"
	close(ch2)

	lengths := []int{recvTask(ch2), recvTask(ch2), recvTask(ch2)}
	expected := []int{5, 6, 2}
	for i, l := range lengths {
		if l != expected[i] {
			t.Errorf("消息 %d 长度: got %d, want %d", i, l, expected[i])
		}
	}
}

// ============================================================
// 4. 测试无缓冲 channel 同步通信
// ============================================================

// TestUnbufferedChannel 验证无缓冲 channel 的同步特性
func TestUnbufferedChannel(t *testing.T) {
	t.Parallel()

	ch := make(chan string)

	// 在一个 goroutine 中发送，主 goroutine 接收
	go func() {
		ch <- "sync message"
	}()

	select {
	case msg := <-ch:
		if msg != "sync message" {
			t.Errorf("got %q, want 'sync message'", msg)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("无缓冲 channel 通信超时")
	}
}

// ============================================================
// 5. 测试有缓冲 channel
// ============================================================

// TestBufferedChannel 验证有缓冲 channel 的异步特性
func TestBufferedChannel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		buffer int
		values []int
	}{
		{name: "缓冲区=3 发送 3 个值", buffer: 3, values: []int{10, 20, 30}},
		{name: "缓冲区=5 发送 2 个值", buffer: 5, values: []int{1, 2}},
		{name: "缓冲区=1 发送 1 个值", buffer: 1, values: []int{42}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan int, tt.buffer)

			// 发送所有值（不应阻塞，因为缓冲区足够）
			for _, v := range tt.values {
				select {
				case ch <- v:
				case <-time.After(100 * time.Millisecond):
					t.Fatalf("发送 %d 阻塞（不应该）", v)
				}
			}

			// 读取验证
			for i, expected := range tt.values {
				select {
				case got := <-ch:
					if got != expected {
						t.Errorf("索引 %d: got %d, want %d", i, got, expected)
					}
				case <-time.After(100 * time.Millisecond):
					t.Fatalf("读取索引 %d 超时", i)
				}
			}
		})
	}
}

// ============================================================
// 6. 测试 range over channel
// ============================================================

// TestRangeOverChannel 验证 for range 能正确遍历已关闭的 channel
func TestRangeOverChannel(t *testing.T) {
	t.Parallel()

	ch := make(chan int, 3)
	ch <- 1
	ch <- 2
	ch <- 3
	close(ch)

	var got []int
	for v := range ch {
		got = append(got, v)
	}

	want := []int{1, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("长度不匹配: got %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("索引 %d: got %d, want %d", i, got[i], want[i])
		}
	}
}

// TestRangeOnClosedChannel 验证关闭后的 channel 能直接 range 完毕
func TestRangeOnClosedChannel(t *testing.T) {
	t.Parallel()

	ch := make(chan int)
	close(ch)

	count := 0
	for range ch {
		count++
	}
	if count != 0 {
		t.Errorf("空 channel range 应不产生值: got %d", count)
	}
}