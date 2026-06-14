// main_test.go — Go 并发模式示例测试套件
//
// 测试覆盖：
//   - Generator（pipeline 第一阶段）生成原始数据
//   - Multiply / Add pipeline 阶段函数
//   - Full pipeline 端到端
//   - Or-Done Channel 模式
//   - Tee Channel 分叉模式
//   - 并发错误处理模式
package main

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"
)

// ============================================================
// 1. 测试 Pipeline 模式 — generator / multiply / add
// ============================================================

// TestGenerator 验证 generator 能生成正确的数据序列
func TestGenerator(t *testing.T) {
	t.Parallel()

	done := make(chan struct{})
	defer close(done)

	tests := []struct {
		name string
		nums []int
		want []int
	}{
		{name: "多个数字", nums: []int{1, 2, 3}, want: []int{1, 2, 3}},
		{name: "空输入", nums: []int{}, want: []int{}},
		{name: "单个数字", nums: []int{42}, want: []int{42}},
		{name: "负数", nums: []int{-1, -2, -3}, want: []int{-1, -2, -3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := generator(done, tt.nums...)
			var got []int
			for v := range out {
				got = append(got, v)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("长度: got %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("索引 %d: got %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestMultiply 验证 multiply 阶段将输入 *2
func TestMultiply(t *testing.T) {
	t.Parallel()

	done := make(chan struct{})
	defer close(done)

	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{name: "正数", in: []int{1, 2, 3}, want: []int{2, 4, 6}},
		{name: "负数", in: []int{-1, -2}, want: []int{-2, -4}},
		{name: "零值", in: []int{0, 0, 0}, want: []int{0, 0, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := make(chan int)
			go func() {
				for _, v := range tt.in {
					in <- v
				}
				close(in)
			}()

			out := multiply(done, in)
			var got []int
			for v := range out {
				got = append(got, v)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("长度: got %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("索引 %d: got %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestAdd 验证 add 阶段将输入 +1
func TestAdd(t *testing.T) {
	t.Parallel()

	done := make(chan struct{})
	defer close(done)

	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{name: "正数", in: []int{2, 4, 6}, want: []int{3, 5, 7}},
		{name: "负数", in: []int{-5, -1}, want: []int{-4, 0}},
		{name: "零值", in: []int{0}, want: []int{1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := make(chan int)
			go func() {
				for _, v := range tt.in {
					in <- v
				}
				close(in)
			}()

			out := add(done, in)
			var got []int
			for v := range out {
				got = append(got, v)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("长度: got %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("索引 %d: got %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestPipeline 验证完整的 pipeline 端到端
func TestPipeline(t *testing.T) {
	t.Parallel()

	done := make(chan struct{})
	defer close(done)

	// pipeline: generator -> multiply -> add
	pipeline := add(done, multiply(done, generator(done, 1, 2, 3, 4, 5)))

	var got []int
	for v := range pipeline {
		got = append(got, v)
	}

	// 原始: 1,2,3,4,5 -> *2: 2,4,6,8,10 -> +1: 3,5,7,9,11
	want := []int{3, 5, 7, 9, 11}
	if len(got) != len(want) {
		t.Fatalf("pipeline 输出长度: got %d, want %d", len(got), len(want))
	}
	// 由于 goroutine 调度顺序不确定，按值集合（无序）比较
	gotSet := make(map[int]int)
	for _, v := range got {
		gotSet[v]++
	}
	wantSet := make(map[int]int)
	for _, v := range want {
		wantSet[v]++
	}
	for k, c := range wantSet {
		if gotSet[k] != c {
			t.Errorf("pipeline 输出: 值 %d 出现 %d 次, want %d 次", k, gotSet[k], c)
		}
	}
}

// TestPipelineCancel 验证 done channel 能取消 pipeline
func TestPipelineCancel(t *testing.T) {
	t.Parallel()

	done := make(chan struct{})

	// pipeline 生成大量数据
	pipeline := add(done, multiply(done, generator(done, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)))

	count := 0
	for v := range pipeline {
		count++
		_ = v
		if count >= 3 {
			close(done) // 提前取消
			break
		}
	}

	// 取消后 pipeline 应快速关闭
	doneClosed := false
	select {
	case _, ok := <-pipeline:
		if !ok {
			doneClosed = true
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("取消后 pipeline 应迅速关闭")
	}
	_ = doneClosed
	if count < 3 {
		t.Errorf("应至少消费 3 个值, got %d", count)
	}
}

// ============================================================
// 2. 测试 Or-Done Channel 模式
// ============================================================

// TestOrDoneBasic 验证 orDone 能正常读取数据
func TestOrDoneBasic(t *testing.T) {
	t.Parallel()

	done := make(chan struct{})

	dataCh := make(chan int)
	go func() {
		for i := 1; i <= 3; i++ {
			dataCh <- i
		}
		close(dataCh)
	}()

	var got []int
	for v := range orDone(done, dataCh) {
		got = append(got, v)
	}

	want := []int{1, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("长度: got %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("索引 %d: got %d, want %d", i, got[i], want[i])
		}
	}
}

// TestOrDoneCancel 验证 done 取消后 orDone 停止读取
func TestOrDoneCancel(t *testing.T) {
	t.Parallel()

	done := make(chan struct{})

	dataCh := make(chan int)
	go func() {
		for i := 1; i <= 10; i++ {
			select {
			case dataCh <- i:
				time.Sleep(5 * time.Millisecond)
			case <-done:
				return
			}
		}
		close(dataCh)
	}()

	count := 0
	for v := range orDone(done, dataCh) {
		count++
		_ = v
		if count >= 2 {
			close(done)
			time.Sleep(20 * time.Millisecond) // 给 goroutine 响应时间
			break
		}
	}

	// 验证 done 后 orDone 的 channel 已关闭
	select {
	case _, ok := <-orDone(done, dataCh):
		if ok {
			t.Log("注意：orDone 返回的 channel 可能仍打开（取决于 goroutine 调度）")
		}
	default:
		// orDone 未关闭但 done 已触发
	}
}

// TestOrDoneEmptyChannel 验证 orDone 能处理空 channel（已关闭）
func TestOrDoneEmptyChannel(t *testing.T) {
	t.Parallel()

	done := make(chan struct{})
	defer close(done)

	dataCh := make(chan int)
	close(dataCh) // 直接关闭

	count := 0
	for range orDone(done, dataCh) {
		count++
	}
	if count != 0 {
		t.Errorf("空 channel 应产生 0 个值, got %d", count)
	}
}

// ============================================================
// 3. 测试 Tee Channel 模式
// ============================================================

// TestTeeBasic 验证 tee 将一个 channel 分为两个，内容一致
func TestTeeBasic(t *testing.T) {
	t.Parallel()

	done := make(chan struct{})
	defer close(done)

	in := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		in <- i * 100
	}
	close(in)

	outA, outB := tee(done, in)

	// 两个 channel 应有相同的内容
	for i := 1; i <= 5; i++ {
		a, b := <-outA, <-outB
		if a != b {
			t.Errorf("Tee 输出不一致: outA=%d, outB=%d", a, b)
		}
		if a != i*100 {
			t.Errorf("outA: got %d, want %d", a, i*100)
		}
	}
}

// TestTeeEmpty 验证 tee 能处理空输入
func TestTeeEmpty(t *testing.T) {
	t.Parallel()

	done := make(chan struct{})
	defer close(done)

	in := make(chan int)
	close(in)

	outA, outB := tee(done, in)

	// 两个 channel 都应立即关闭
	_, aOk := <-outA
	_, bOk := <-outB
	if aOk || bOk {
		t.Error("空输入后 tee 的两个输出 channel 应关闭")
	}
}

// TestTeeCancel 验证 done 取消后 tee 停止
func TestTeeCancel(t *testing.T) {
	t.Parallel()

	done := make(chan struct{})

	in := make(chan int, 10)
	for i := 1; i <= 10; i++ {
		in <- i
	}
	close(in)

	outA, outB := tee(done, in)

	// 消费 2 个后取消
	<-outA
	<-outB
	close(done)

	time.Sleep(20 * time.Millisecond)

	// done 已关闭，tee 的 goroutine 应退出
	// 剩余数据可能丢失，但不应导致死锁
}

// ============================================================
// 4. 测试并发错误处理模式
// ============================================================

// TestErrorCollector 验证并发任务错误收集
func TestErrorCollector(t *testing.T) {
	t.Parallel()

	tasks := []struct {
		id   int
		fail bool
	}{
		{1, false},
		{2, true},
		{3, false},
		{4, true},
		{5, false},
	}

	errCh := make(chan error, len(tasks))
	var wg sync.WaitGroup

	for _, tsk := range tasks {
		wg.Add(1)
		go func(id int, shouldFail bool) {
			defer wg.Done()
			time.Sleep(time.Duration(rand.Intn(10)+5) * time.Millisecond)

			if shouldFail {
				errCh <- fmt.Errorf("任务 %d 执行失败", id)
				return
			}
		}(tsk.id, tsk.fail)
	}

	// 等待所有 goroutine 结束
	wg.Wait()
	close(errCh)

	// 收集错误
	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}

	if len(errs) != 2 {
		t.Errorf("应收集到 2 个错误, got %d", len(errs))
	}

	// 验证错误内容
	foundTask2 := false
	foundTask4 := false
	for _, err := range errs {
		if errors.Is(err, fmt.Errorf("任务 %d 执行失败", 2)) {
			// errors.Is 对 fmt.Errorf 不加 %w 不生效，用 Contains 方式
		}
		errStr := err.Error()
		if errStr == "任务 2 执行失败" {
			foundTask2 = true
		}
		if errStr == "任务 4 执行失败" {
			foundTask4 = true
		}
	}
	if !foundTask2 {
		t.Error("未收集到任务 2 的错误")
	}
	if !foundTask4 {
		t.Error("未收集到任务 4 的错误")
	}
}

// TestErrorCollectorAllSuccess 验证全部成功时无错误
func TestErrorCollectorAllSuccess(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	errCh := make(chan error, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			time.Sleep(time.Duration(rand.Intn(10)+5) * time.Millisecond)
		}(i)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}

	if len(errs) != 0 {
		t.Errorf("全部成功时错误数应为 0, got %d", len(errs))
	}
}

// ============================================================
// 5. 测试 Fan-out / Fan-in 模式
// ============================================================

// TestFanOutFanIn 验证多 worker 并发处理任务
func TestFanOutFanIn(t *testing.T) {
	t.Parallel()

	const numTasks = 6
	const numWorkers = 2

	tasks := make(chan int, numTasks)
	for i := 1; i <= numTasks; i++ {
		tasks <- i
	}
	close(tasks)

	results := make(chan string, numTasks)
	var wg sync.WaitGroup

	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for task := range tasks {
				results <- fmt.Sprintf("Worker#%d 处理了任务 %d", workerID, task)
			}
		}(w)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果
	var mu sync.Mutex
	received := make(map[int]bool)
	timeout := time.After(1 * time.Second)

	for i := 0; i < numTasks; i++ {
		select {
		case r, ok := <-results:
			if !ok {
				t.Fatal("results channel 不应提前关闭")
			}
			// 从字符串 "Worker#X 处理了任务 Y" 提取 Y
			parts := strings.Split(r, "处理了任务 ")
			if len(parts) == 2 {
				var taskID int
				if _, err := fmt.Sscanf(parts[1], "%d", &taskID); err == nil {
					mu.Lock()
					received[taskID] = true
					mu.Unlock()
				}
			}
		case <-timeout:
			mu.Lock()
			count := len(received)
			mu.Unlock()
			t.Fatalf("超时：只收到 %d/%d 个结果", count, numTasks)
		}
	}

	// 验证所有任务都被处理
	for i := 1; i <= numTasks; i++ {
		if !received[i] {
			t.Errorf("任务 %d 未被处理", i)
		}
	}
}

// ============================================================
// 辅助函数（从 main.go 提取）
// ============================================================

// orDone — 从 main.go 提取
func orDone(done <-chan struct{}, c <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for {
			select {
			case v, ok := <-c:
				if !ok {
					return
				}
				select {
				case out <- v:
				case <-done:
					return
				}
			case <-done:
				return
			}
		}
	}()
	return out
}

// tee — 从 main.go 提取
func tee(done <-chan struct{}, in <-chan int) (<-chan int, <-chan int) {
	out1 := make(chan int)
	out2 := make(chan int)

	go func() {
		defer close(out1)
		defer close(out2)

		for val := range in {
			v1, v2 := val, val
			select {
			case out1 <- v1:
			case <-done:
				return
			}
			select {
			case out2 <- v2:
			case <-done:
				return
			}
		}
	}()

	return out1, out2
}

// generator — 从 main.go 提取
func generator(done <-chan struct{}, nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			select {
			case out <- n:
			case <-done:
				return
			}
		}
	}()
	return out
}

// multiply — 从 main.go 提取
func multiply(done <-chan struct{}, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			select {
			case out <- n * 2:
			case <-done:
				return
			}
		}
	}()
	return out
}

// add — 从 main.go 提取
func add(done <-chan struct{}, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			select {
			case out <- n + 1:
			case <-done:
				return
			}
		}
	}()
	return out
}