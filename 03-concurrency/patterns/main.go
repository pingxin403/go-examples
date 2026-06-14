package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ============================================================
// Go 并发模式示例
// 涵盖：Fan-out/Fan-in、Pipeline、Or-Done Channel、
// Tee Channel、errgroup 替代（WaitGroup+error channel）
// ============================================================

func main() {
	// ============================================================
	// 1. Fan-out / Fan-in 模式
	//    Fan-out: 将一个输入 channel 分发到多个 worker
	//    Fan-in:  将多个结果 channel 合并为一个
	// ============================================================
	fmt.Println("=== 1. Fan-out / Fan-in ===")

	const numTasks = 10
	const numWorkers = 3

	// 输入任务
	tasks := make(chan int, numTasks)
	for i := 1; i <= numTasks; i++ {
		tasks <- i
	}
	close(tasks)

	// Fan-out: 启动多个 worker，每个 worker 从 tasks channel 消费
	results := make(chan string, numTasks)

	var wgFan sync.WaitGroup
	for w := 1; w <= numWorkers; w++ {
		wgFan.Add(1)
		go func(workerID int) {
			defer wgFan.Done()
			for task := range tasks {
				// 模拟工作
				time.Sleep(time.Duration(rand.Intn(20)+10) * time.Millisecond)
				results <- fmt.Sprintf("Worker#%d 处理了任务 %d", workerID, task)
			}
		}(w)
	}

	// 等待所有 worker 完成，然后关闭 results
	go func() {
		wgFan.Wait()
		close(results)
	}()

	// Fan-in: 从 results 收集结果
	resultCount := 0
	for r := range results {
		fmt.Printf("  %s\n", r)
		resultCount++
	}
	fmt.Printf("  ✅ 共完成 %d 个任务\n", resultCount)

	// ============================================================
	// 2. Pipeline 模式 — 将处理拆分为多个阶段
	// ============================================================
	fmt.Println("\n=== 2. Pipeline 模式 ===")

	// stage1: 生成原始数据
	generator := func(done <-chan struct{}, nums ...int) <-chan int {
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

	// stage2: 数据 *2
	multiply := func(done <-chan struct{}, in <-chan int) <-chan int {
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

	// stage3: 数据 +1
	add := func(done <-chan struct{}, in <-chan int) <-chan int {
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

	done := make(chan struct{})
	defer close(done) // 确保 pipeline 结束时清理

	// pipeline: generator -> multiply -> add -> sink
	pipeline := add(done, multiply(done, generator(done, 1, 2, 3, 4, 5)))

	for result := range pipeline {
		fmt.Printf("  Pipeline 输出: %d\n", result)
		// 原始: 1,2,3,4,5
		// *2:   2,4,6,8,10
		// +1:   3,5,7,9,11
	}

	// ============================================================
	// 3. Or-Done Channel 模式
	//    将 done channel 和 数据 channel 合并为一个 channel
	// ============================================================
	fmt.Println("\n=== 3. Or-Done Channel 模式 ===")

	// orDone 函数封装了 select 模式，避免重复编写
	orDone := func(done <-chan struct{}, c <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for {
				select {
				case v, ok := <-c:
					if !ok {
						return // 数据 channel 关闭
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

	dataCh := make(chan int)
	doneOr := make(chan struct{})

	// 启动发送
	go func() {
		for i := 1; i <= 3; i++ {
			dataCh <- i
			time.Sleep(10 * time.Millisecond)
		}
		close(dataCh)
	}()

	// 用 orDone 安全读取
	for v := range orDone(doneOr, dataCh) {
		fmt.Printf("  orDone 读取: %d\n", v)
	}
	fmt.Println("  orDone: 数据已消费完")

	// 测试 done 取消
	dataCh2 := make(chan int)
	doneOr2 := make(chan struct{})

	go func() {
		for i := 1; i <= 5; i++ {
			select {
			case dataCh2 <- i:
				time.Sleep(10 * time.Millisecond)
			case <-doneOr2:
				return
			}
		}
		close(dataCh2)
	}()

	// 消费两个就取消
	count := 0
	for v := range orDone(doneOr2, dataCh2) {
		fmt.Printf("  提前取消演示 - 读取: %d\n", v)
		count++
		if count >= 2 {
			close(doneOr2) // 取消
			fmt.Println("  orDone: 已取消，不再消费剩余数据")
			time.Sleep(20 * time.Millisecond) // 给 goroutine 时间响应取消
			break
		}
	}

	// ============================================================
	// 4. Tee Channel 模式 — 将一个 channel 拆分为两个
	// ============================================================
	fmt.Println("\n=== 4. Tee Channel（分叉）模式 ===")

	// tee 函数将一个输入 channel 分叉为两个输出 channel
	tee := func(done <-chan struct{}, in <-chan int) (<-chan int, <-chan int) {
		out1 := make(chan int)
		out2 := make(chan int)

		go func() {
			defer close(out1)
			defer close(out2)

			for val := range in {
				// 对每个输入值，分别发送到两个输出
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

	// 创建输入
	teeIn := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		teeIn <- i * 100
	}
	close(teeIn)

	doneTee := make(chan struct{})
	defer close(doneTee)

	outA, outB := tee(doneTee, teeIn)

	// 分别从两个 channel 读取
	for i := 1; i <= 5; i++ {
		a, b := <-outA, <-outB
		fmt.Printf("  Tee: outA=%d, outB=%d（两者相同 ✅）\n", a, b)
	}

	// ============================================================
	// 5. Error 处理模式 — 用 WaitGroup + 错误收集
	// ============================================================
	fmt.Println("\n=== 5. 并发错误处理（WaitGroup + error channel）===")

	// 模拟一批任务，某些可能失败
	tasksErr := []struct {
		id   int
		fail bool
	}{
		{1, false},
		{2, true},  // 这个会失败
		{3, false},
		{4, true},  // 这个也会失败
		{5, false},
	}

	errCh := make(chan error, len(tasksErr))
	var wgErr sync.WaitGroup

	for _, t := range tasksErr {
		wgErr.Add(1)
		go func(id int, shouldFail bool) {
			defer wgErr.Done()
			time.Sleep(time.Duration(rand.Intn(20)+10) * time.Millisecond)

			if shouldFail {
				errCh <- fmt.Errorf("任务 %d 执行失败", id)
				return
			}
			fmt.Printf("  任务 %d 成功完成\n", id)
		}(t.id, t.fail)
	}

	// 等所有任务结束，关闭 error channel
	go func() {
		wgErr.Wait()
		close(errCh)
	}()

	// 收集所有错误
	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		fmt.Printf("  ❌ 共 %d 个错误:\n", len(errors))
		for _, err := range errors {
			fmt.Printf("    - %v\n", err)
		}
	} else {
		fmt.Println("  ✅ 所有任务成功")
	}

	fmt.Println("\n✅ 所有并发模式示例完成")
}